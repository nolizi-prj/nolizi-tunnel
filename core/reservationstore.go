package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"syscall"
	"time"
)

// The durable half of a reservation (spec/0004 §7 slice 2, specified in §14).
//
// Reservations answers "whose name is it?". This file is what makes that
// answer outlive the process, which is the question a restart destroys the
// answer to and the only thing a reconnect cannot recover. It does NOT make a
// restart shorter: a live TCP connection cannot outlive the process at either
// end, and no line here changes that (§4). What it removes is the LOSS — an
// in-memory set is empty after a restart, so every claimed name is
// trust-on-first-use again for the length of the reclaim window.
//
// Three decisions were taken in the spec before this file existed, because
// each of them is a product and not an implementation detail (§14):
//
//	fsync    both — the temp file before the rename, the directory after it.
//	         The rename buys ATOMICITY; only the fsyncs buy durability, and
//	         §4's host-restart column is a claim about durability.
//	corrupt  start EMPTY, log ERROR, move the damaged bytes aside. Refusing to
//	         start turns a bookkeeping problem into a total outage of every
//	         path, including the anonymous one that never touches this file.
//	sharing  refused, by flock on a sibling lock file. Two relays over one
//	         path do not interleave — each writes the WHOLE document, so they
//	         take turns silently destroying each other's claims.

// Store errors.
var (
	// ErrStoreLocked is a second store over a path a live store holds.
	ErrStoreLocked = errors.New("core: reservation store is locked by another process")
	// ErrReservationsFull is a new claim against a set at MaxReservations.
	// An owner returning to a name already in the set never sees it.
	ErrReservationsFull = errors.New("core: reservation set is full")
)

const (
	// ReservationTTL is how long a claim survives without its owner. Swept at
	// load and nowhere else: a background sweeper would race the handshake
	// path for the same document, for an expiry whose whole purpose is to
	// bound growth over months.
	ReservationTTL = 30 * 24 * time.Hour

	// MaxReservations caps the SET, not one token's share of it. At the cap a
	// NEW name is refused and an existing owner is not (§14.5). The number is
	// where a full-document rewrite is still trivial — 10,000 records at
	// roughly 150 bytes is about 1.5 MB — and far above any plausible
	// legitimate use of one relay.
	//
	// It is not a defence against namespace exhaustion and this comment will
	// not pretend otherwise: a party who can complete 10,000 handshakes can
	// fill the set. Before the cap that party filled the disk instead.
	MaxReservations = 10000

	// storeVersion makes a future format change DETECTABLE rather than
	// guessed at. A document at any other version is not parsed further and
	// takes the corrupt path.
	storeVersion = 1
)

// storeDoc is the on-disk document. Written whole, every time.
type storeDoc struct {
	Version      int           `json:"version"`
	Reservations []storeRecord `json:"reservations"`
}

type storeRecord struct {
	Subdomain string    `json:"subdomain"`
	TokenHash string    `json:"token_hash"`
	TCPPort   int       `json:"tcp_port,omitempty"`
	LastSeen  time.Time `json:"last_seen"`
}

// store is one path, its lock, and the write path onto it.
type store struct {
	path string
	lock *os.File // <path>.lock, held for the life of the store
	log  *slog.Logger

	// failWriteAfterTemp is D-2's hook and exists for no other reason. A
	// crash between the temp file and the rename is the one moment the
	// temp-and-rename design is FOR, and it cannot be reached from outside
	// this process without killing it mid-syscall. Nil in every path a
	// binary takes.
	failWriteAfterTemp func(tmp string) error
}

// OpenReservations returns a reservation set backed by a file at path.
//
// It takes an exclusive lock on a sibling <path>.lock, loads what is there,
// sweeps anything idle longer than ReservationTTL, and returns a set whose
// every mutation is written through. Close releases the lock.
//
// A path that does not exist yet is the ordinary first boot: empty set, no
// error, and nothing logged. A path that exists and cannot be loaded is
// §14.2 — an empty set, one ERROR, and the damaged bytes moved aside.
func OpenReservations(path string, log *slog.Logger) (*Reservations, error) {
	if log == nil {
		log = slog.New(slog.NewTextHandler(nopWriter{}, nil))
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("core: reservation store directory: %w", err)
	}

	// The lock is on a SIBLING and not on path itself, because path is
	// replaced by rename on every write and a lock on the old inode guards
	// nothing at all.
	lockPath := path + ".lock"
	lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("core: reservation store lock: %w", err)
	}
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		lock.Close()
		return nil, fmt.Errorf("%w: %s", ErrStoreLocked, path)
	}

	s := &store{path: path, lock: lock, log: log}
	r := NewReservations()
	r.store = s

	if err := s.load(r); err != nil {
		// Never fatal. §14.2: starting empty degrades to the behaviour every
		// release before this one shipped, and refusing to start would take
		// the anonymous path down with the reservation set.
		log.Error("reservation store could not be loaded; starting with an empty set",
			"path", path, "error", err)
		s.moveAside()
	}
	return r, nil
}

// Close releases the lock. A store that is not closed is released by the
// kernel when the process ends, so there is no stale lock to clean up by hand
// and no lock file to delete.
func (r *Reservations) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.store == nil || r.store.lock == nil {
		return nil
	}
	f := r.store.lock
	r.store.lock = nil
	// Unlocked implicitly by the close; done explicitly so a caller reading
	// this does not have to know that.
	syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	return f.Close()
}

// load reads the document into r, sweeping as it goes. r is assumed fresh.
func (s *store) load(r *Reservations) error {
	raw, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			// First boot. Not an error, and deliberately not logged: a store
			// that shouted here would train an operator to ignore it.
			return nil
		}
		// An unreadable existing file takes the same path as an unparseable
		// one. Different cause, same available responses, and one rule beats
		// two a reader has to hold apart (§14.2).
		return fmt.Errorf("reading %s: %w", s.path, err)
	}

	var doc storeDoc
	if err := json.Unmarshal(raw, &doc); err != nil {
		return fmt.Errorf("parsing %s: %w", s.path, err)
	}
	if doc.Version != storeVersion {
		return fmt.Errorf("%s is format version %d, this relay writes %d", s.path, doc.Version, storeVersion)
	}

	now := time.Now()
	cutoff := now.Add(-ReservationTTL)
	swept := 0
	for _, rec := range doc.Reservations {
		name := normaliseName(rec.Subdomain)
		if name == "" || rec.TokenHash == "" {
			swept++
			continue
		}
		seen := rec.LastSeen
		if seen.IsZero() {
			// Treated as seen NOW, not as the epoch. Reading zero as the
			// epoch would drop every such record at once, and losing a whole
			// namespace to a missing field is a far worse failure than
			// holding one name thirty days too long (§14.5).
			seen = now
		}
		if seen.Before(cutoff) {
			swept++
			continue
		}
		if len(r.byName) >= MaxReservations {
			swept++
			continue
		}
		// A port is dropped with the name that held it, never left behind:
		// a hold with nothing owning it is a number no one can ever be given.
		if rec.TCPPort != 0 {
			if owner, taken := r.byPort[rec.TCPPort]; taken && owner != name {
				rec.TCPPort = 0
			}
		}
		r.byName[name] = Reservation{
			Subdomain: name,
			TokenHash: rec.TokenHash,
			TCPPort:   rec.TCPPort,
			LastSeen:  seen,
		}
		if rec.TCPPort != 0 {
			r.byPort[rec.TCPPort] = name
		}
	}
	if swept > 0 {
		s.log.Info("reservation store loaded", "path", s.path, "kept", len(r.byName), "dropped", swept)
	}
	return nil
}

// moveAside renames a damaged document out of the way, so the first write
// after the restart does not overwrite the only evidence of what went wrong.
// A failure here is logged and no more: preserving evidence must not be the
// thing that stops the relay either (§14.2).
func (s *store) moveAside() {
	kept := fmt.Sprintf("%s.corrupt-%s", s.path, time.Now().UTC().Format(time.RFC3339))
	if err := os.Rename(s.path, kept); err != nil {
		if !os.IsNotExist(err) {
			s.log.Error("could not preserve the damaged reservation store", "path", s.path, "error", err)
		}
		return
	}
	s.log.Error("the damaged reservation store was kept", "path", kept)
}

// write persists the whole set. The caller holds r.mu.
//
// Write-to-temp-and-rename, with both fsyncs, in the order §14.1 specifies:
// a crash mid-write leaves the PREVIOUS set rather than half of one, and the
// two Syncs are what make that true of a host losing power rather than only
// of a process dying.
func (s *store) write(byName map[string]Reservation) error {
	doc := storeDoc{Version: storeVersion, Reservations: make([]storeRecord, 0, len(byName))}
	for _, held := range byName {
		doc.Reservations = append(doc.Reservations, storeRecord{
			Subdomain: held.Subdomain,
			TokenHash: held.TokenHash,
			TCPPort:   held.TCPPort,
			LastSeen:  held.LastSeen.UTC(),
		})
	}
	// Sorted, so two identical sets produce identical bytes and a diff of two
	// store files is a diff of their contents rather than of map order.
	sort.Slice(doc.Reservations, func(i, j int) bool {
		return doc.Reservations[i].Subdomain < doc.Reservations[j].Subdomain
	})
	body, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("core: encoding the reservation store: %w", err)
	}
	body = append(body, '\n')

	// The temp file is in the SAME directory as the target: rename(2) across
	// filesystems fails, and is not atomic where it does not.
	dir := filepath.Dir(s.path)
	tmp := s.path + ".tmp"
	// 0600 — the document holds token digests. A digest is not a live
	// credential (§3.1) and is still the input to an offline attack.
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("core: reservation store temp file: %w", err)
	}
	if _, err := f.Write(body); err != nil {
		f.Close()
		os.Remove(tmp)
		return fmt.Errorf("core: writing the reservation store: %w", err)
	}
	// Before the rename, so the rename cannot become durable ahead of the
	// bytes it points at.
	if err := f.Sync(); err != nil {
		f.Close()
		os.Remove(tmp)
		return fmt.Errorf("core: syncing the reservation store: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("core: closing the reservation store: %w", err)
	}

	if s.failWriteAfterTemp != nil {
		if err := s.failWriteAfterTemp(tmp); err != nil {
			return err
		}
	}

	if err := os.Rename(tmp, s.path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("core: renaming the reservation store: %w", err)
	}

	// And the directory entry the rename created, or the rename itself can be
	// lost to a power failure that the file's own Sync survived.
	d, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("core: opening the reservation store directory: %w", err)
	}
	defer d.Close()
	if err := d.Sync(); err != nil {
		return fmt.Errorf("core: syncing the reservation store directory: %w", err)
	}
	return nil
}

// nopWriter is the discard sink for a nil logger, so this file never has to
// ask whether s.log is present.
type nopWriter struct{}

func (nopWriter) Write(p []byte) (int, error) { return len(p), nil }

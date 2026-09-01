package core

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// A registration and a reservation answer two different questions, and this
// package keeps them apart on purpose (spec/0004 §2).
//
//	Registry     — is this name in use *right now*?  Dies with a connection.
//	Reservations — whose name is it?                 Outlives every connection.
//
// A relay that can only answer the first cannot keep a promise about restarts,
// because the second is the question a restart destroys the answer to. Before
// this type existed, core.Tunnel had a Reserved field computed from the shape
// of a request and read by nothing at all.
//
// The policy here is pure and the I/O is not. A set built by NewReservations
// does no I/O at all and behaves exactly as it did before slice 2 existed; a
// set built by OpenReservations (core/reservationstore.go) writes every
// mutation through to a file, which is what makes ownership outlive the
// process rather than only the connection.

// Reservation errors.
var (
	ErrNameReserved  = errors.New("core: subdomain is reserved")
	ErrPortReserved  = errors.New("core: public tcp port is reserved")
	ErrTokenTooShort = errors.New("core: token is too short")
)

// MinTokenLen is the shortest token that may claim a name.
//
// It is what lets a reservation store a plain SHA-256 digest rather than a
// password-hashing function: a KDF exists to make a *low-entropy* secret
// expensive to guess, and this constant forbids low-entropy secrets instead of
// compensating for them. Against a secret this long, an offline attack on the
// digest is not the cheapest way in — the plaintext control connection is
// (spec/0004 §3).
//
// Lowering this makes that argument wrong, and the hash has to change with it.
const MinTokenLen = 16

// Reservation is a claim on a name that outlives the connection using it.
//
// Every field here is read. LastSeen arrived with slice 2, in the same change
// as the sweep that reads it — slice 1 declined to ship it early, because
// shipping a field ahead of its reader is the defect this type exists to
// repair. There is still no CreatedAt, for the same reason: nothing reads one.
type Reservation struct {
	// Subdomain is the claimed name, lowercased.
	Subdomain string
	// TokenHash is the hex SHA-256 of the bearer secret. The secret itself is
	// never stored, so a leaked set is not a set of live credentials.
	TokenHash string
	// TCPPort is the public port claimed alongside the name; 0 for none.
	TCPPort int
	// LastSeen is when the owner last proved it held this name, which is to
	// say the last successful Claim. Check does not touch it: a stranger
	// being refused a name must not refresh its owner's clock. It is what the
	// load-time sweep measures ReservationTTL against (spec/0004 §14.5).
	LastSeen time.Time
}

// Reservations is the set of claimed names. Safe for concurrent use: the relay
// claims from a handshake while serving requests on many goroutines.
type Reservations struct {
	mu     sync.RWMutex
	byName map[string]Reservation
	byPort map[int]string // public TCP port -> owning subdomain

	// store is nil for an in-memory set, which is what NewReservations
	// returns and what a relay with no -reservations path uses. When it is
	// present every mutation is written through before it is returned.
	store *store
}

// NewReservations returns an empty set. An empty set is not a disabled one:
// every name in it is unclaimed, which is exactly today's behaviour, and the
// rules below apply from the first claim onward.
func NewReservations() *Reservations {
	return &Reservations{
		byName: make(map[string]Reservation),
		byPort: make(map[int]string),
	}
}

// HashToken is the one place a token becomes a digest, so the store format and
// the comparison cannot drift apart.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// normaliseName matches what Registry does to a name arriving from the
// network, so a claim and a lookup cannot disagree about case.
func normaliseName(subdomain string) string {
	return strings.ToLower(strings.TrimSpace(subdomain))
}

// checkToken rejects a token that is present but too weak to claim with.
// Absent is allowed here and refused later by whoever needs one: an anonymous
// agent has no token and that is not an error.
func checkToken(token string) error {
	if token == "" {
		return nil
	}
	if len(token) < MinTokenLen {
		return fmt.Errorf("%w: %d characters, minimum is %d", ErrTokenTooShort, len(token), MinTokenLen)
	}
	return nil
}

// sameToken compares a presented token against a stored digest in constant
// time, so a wrong token does not leak how much of it was right.
func sameToken(token, storedHash string) bool {
	return subtle.ConstantTimeCompare([]byte(HashToken(token)), []byte(storedHash)) == 1
}

// Check reports whether a caller holding token may use subdomain. It records
// nothing.
//
// An unclaimed name is available to anyone, which is what keeps the anonymous
// case working: this type narrows which name an accepted agent may have, and
// never decides whether the agent is accepted.
func (r *Reservations) Check(subdomain, token string) error {
	if err := checkToken(token); err != nil {
		return err
	}
	name := normaliseName(subdomain)

	r.mu.RLock()
	defer r.mu.RUnlock()
	held, ok := r.byName[name]
	if !ok {
		return nil
	}
	if token != "" && sameToken(token, held.TokenHash) {
		return nil
	}
	return fmt.Errorf("%w: %q", ErrNameReserved, name)
}

// Claim is trust on first use: an unclaimed name becomes the caller's, and a
// name already theirs stays theirs. created reports whether this call recorded
// a new reservation, so a handshake that fails afterwards can undo exactly what
// it did and no more.
//
// tcpPort is the public port the caller wants claimed with the name. Zero
// asserts nothing and leaves any port already claimed in place — an owner that
// reconnects as a plain HTTP tunnel should not lose an address to an argument
// it merely omitted.
//
// A reservation is one address and not a growing set: asking for a second,
// different port under a name that already holds one is refused rather than
// accumulated, which is what stops one token draining the pool a reconnect at
// a time.
func (r *Reservations) Claim(subdomain, token string, tcpPort int) (created bool, err error) {
	if err := checkToken(token); err != nil {
		return false, err
	}
	if token == "" {
		return false, fmt.Errorf("%w: a claim needs a token", ErrTokenTooShort)
	}
	name := normaliseName(subdomain)

	r.mu.Lock()
	defer r.mu.Unlock()

	held, exists := r.byName[name]
	if exists && !sameToken(token, held.TokenHash) {
		return false, fmt.Errorf("%w: %q", ErrNameReserved, name)
	}
	// A port may belong to one name only. Checked before anything is written,
	// so a refusal leaves the set exactly as it was.
	if tcpPort != 0 {
		if owner, taken := r.byPort[tcpPort]; taken && owner != name {
			return false, fmt.Errorf("%w: port %d belongs to %q", ErrPortReserved, tcpPort, owner)
		}
		if exists && held.TCPPort != 0 && held.TCPPort != tcpPort {
			return false, fmt.Errorf("%w: %q already holds port %d", ErrPortReserved, name, held.TCPPort)
		}
	}

	// The cap is on the SET and it refuses only a NEW name. An owner coming
	// back to a name already in the set is never refused by it — a cap that
	// could lock out an existing owner would destroy the property this whole
	// type exists to establish (spec/0004 §14.5).
	if !exists && len(r.byName) >= MaxReservations {
		return false, fmt.Errorf("%w: %d names", ErrReservationsFull, len(r.byName))
	}

	if !exists {
		held = Reservation{Subdomain: name, TokenHash: HashToken(token)}
	}
	// Enough to undo this call exactly, and no more, if the write fails.
	before, portAdded := held, false
	if tcpPort != 0 && held.TCPPort == 0 {
		held.TCPPort = tcpPort
		r.byPort[tcpPort] = name
		portAdded = true
	}
	// Set here and nowhere else. This is the one call in which the owner
	// proves it still holds the name, which is what "last seen" means.
	held.LastSeen = time.Now()
	r.byName[name] = held

	// Written through before the claim is returned. A claim that has come
	// back to the handshake has been persisted, which is the only way its
	// return value can mean what spec/0004 §14 says it means.
	if err := r.persist(); err != nil {
		// Rolled back to exactly what was there. A claim that cannot be
		// persisted is not a claim: the only thing a claim promises over a
		// plain registration is that it survives the process, so refusing is
		// the honest answer and it costs the anonymous path nothing, which
		// never reaches this code (§14.6).
		if portAdded {
			delete(r.byPort, tcpPort)
		}
		if exists {
			r.byName[name] = before
		} else {
			delete(r.byName, name)
		}
		return false, err
	}
	return !exists, nil
}

// persist writes the set through to its store, if it has one. The caller
// holds r.mu. An in-memory set returns nil and touches no disk, which is what
// makes a relay with no -reservations path exactly the relay of every release
// before this one.
func (r *Reservations) persist() error {
	if r.store == nil {
		return nil
	}
	return r.store.write(r.byName)
}

// Discard removes a claim, and reports whether the store agrees.
//
// It is not a user-facing release and nothing outside the relay calls it: it
// exists so a handshake that claimed a name and then failed — the registry
// refused it as live, the public port would not bind, the announce could not be
// written — leaves nothing behind. A name is not consumed by a connection that
// never opened.
//
// UNLIKE Claim, a failed write does NOT roll this back. Discard exists to keep
// that invariant now, and making it fail would leave in memory the very claim
// it was called to remove. The caller gets the error and logs it. What that
// asymmetry leaves behind is written down in spec/0004 §14.6 rather than left
// to be discovered: a handshake whose Claim write succeeded and whose Discard
// write then failed leaves a record on disk that memory does not have, and it
// comes back at the next restart, bounded by the sweep.
//
// The error return was added by slice 2. A call used as a statement is
// unaffected by it, so no frozen case moved.
func (r *Reservations) Discard(subdomain string) error {
	name := normaliseName(subdomain)
	r.mu.Lock()
	defer r.mu.Unlock()
	held, ok := r.byName[name]
	if !ok {
		return nil
	}
	if held.TCPPort != 0 {
		delete(r.byPort, held.TCPPort)
	}
	delete(r.byName, name)
	return r.persist()
}

// PortHolder reports which name holds a public TCP port, or "" if none does.
func (r *Reservations) PortHolder(port int) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.byPort[port]
}

// Get returns the reservation on a name, if there is one. It is what tells a
// live tunnel whether its name belongs to anybody.
func (r *Reservations) Get(subdomain string) (Reservation, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	held, ok := r.byName[normaliseName(subdomain)]
	return held, ok
}

// Snapshot returns every reservation in the set, in no particular order.
//
// It exists for one caller: a relay rebuilding its port pool's holds at
// startup. A durable reservation set is the record of who owns a public port,
// and the pool is DERIVED from it — a pool built empty beside a set that
// survived a restart would hand a stranger the number that set says is
// spoken for, which is half of spec/0004 §4's middle column silently missing.
func (r *Reservations) Snapshot() []Reservation {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Reservation, 0, len(r.byName))
	for _, held := range r.byName {
		out = append(out, held)
	}
	return out
}

// Len reports how many names are claimed.
func (r *Reservations) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.byName)
}

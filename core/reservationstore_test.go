package core

// Acceptance cases D-2 .. D-6 for spec/0004-names-with-owners, added to
// acceptance/CASES.md by the SPEC.md §14 amendment before any of this code
// existed. See CASES.md for what execution makes each one fail.
//
// This file is `package core` and not `package core_test`, unlike every other
// test in this package, and D-2 is the whole reason. A crash between the temp
// file and the rename is the one moment write-to-temp-and-rename exists FOR,
// and it cannot be reached from outside this process without killing it
// mid-syscall. The hook it uses (store.failWriteAfterTemp) is nil on every
// path a binary takes and is unreachable from outside this package.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	storeOwnerToken    = "owner-token-0123456789"
	storeSecondToken   = "second-token-9876543210"
	storeStrangerToken = "stranger-token-abcdefghij"
)

func storePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "reservations.json")
}

func mustOpen(t *testing.T, path string) *Reservations {
	t.Helper()
	r, err := OpenReservations(path, nil)
	if err != nil {
		t.Fatalf("OpenReservations(%s): %v", path, err)
	}
	return r
}

// D-2 · A crash mid-write leaves the previous set.
//
// Fails when the document is written in place rather than temp-and-renamed, so
// the interrupted write leaves a truncated file and the WHOLE set — not just
// the new claim — is gone at the next load.
func TestACrashMidWriteLeavesThePreviousSet(t *testing.T) {
	path := storePath(t)

	first := mustOpen(t, path)
	if _, err := first.Claim("first", storeOwnerToken, 20000); err != nil {
		t.Fatalf("claiming first: %v", err)
	}
	// LastSeen is set by Claim and by nothing else (SPEC.md §14.5). Asserted
	// here because this is the first claim any case in this file makes.
	held, ok := first.Get("first")
	if !ok {
		t.Fatal("the claim on first was not recorded")
	}
	if time.Since(held.LastSeen) > time.Minute || held.LastSeen.IsZero() {
		t.Fatalf("Claim left LastSeen at %v, want about now", held.LastSeen)
	}

	// The crash: the temp file is written, synced and closed, and the process
	// dies before the rename. Asserted rather than assumed — if the temp file
	// is not there at this instant, this case is not testing the moment it
	// says it is.
	sawTemp := false
	first.store.failWriteAfterTemp = func(tmp string) error {
		if _, err := os.Stat(tmp); err == nil {
			sawTemp = true
		}
		return errors.New("the process died here")
	}
	if _, err := first.Claim("second", storeSecondToken, 20001); err == nil {
		t.Fatal("a claim whose write failed was reported as successful")
	}
	if !sawTemp {
		t.Fatal("the write failed before the temp file existed, which is not the moment this case is about")
	}
	// And the failed claim rolled back in memory too (§14.6): a claim that
	// cannot be persisted is not a claim, in either half of the set.
	if _, ok := first.Get("second"); ok {
		t.Error("a claim whose write failed is still in memory")
	}
	if holder := first.PortHolder(20001); holder != "" {
		t.Errorf("a claim whose write failed left port 20001 held by %q", holder)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("closing the first store: %v", err)
	}

	// What the next process finds is the PREVIOUS set, record for record.
	second := mustOpen(t, path)
	t.Cleanup(func() { second.Close() })
	if second.Len() != 1 {
		t.Fatalf("after a crash mid-write the set has %d names, want 1", second.Len())
	}
	back, ok := second.Get("first")
	if !ok {
		t.Fatal("a crash mid-write lost the set that was already written")
	}
	if back.TokenHash != HashToken(storeOwnerToken) {
		t.Error("the reloaded record does not carry the owner's digest")
	}
	if back.TCPPort != 20000 {
		t.Errorf("the reloaded record holds port %d, want 20000", back.TCPPort)
	}
	if holder := second.PortHolder(20000); holder != "first" {
		t.Errorf("port 20000 is held by %q after the reload, want \"first\"", holder)
	}
	if _, ok := second.Get("second"); ok {
		t.Error("the claim whose write was interrupted came back anyway")
	}
}

// D-3 · An idle reservation is swept at load, and a live one is not.
//
// Fails when nothing sweeps, so a name is held forever for an owner who has
// vanished; or the sweep drops by name and leaves the port held, so a number
// is unallocatable with nothing owning it.
func TestAnIdleReservationIsSweptAtLoad(t *testing.T) {
	path := storePath(t)
	now := time.Now().UTC()

	// Written by hand, in the real on-disk format, so this case is evidence
	// about the format as well as about the sweep.
	doc := fmt.Sprintf(`{
  "version": 1,
  "reservations": [
    {"subdomain": "stale", "token_hash": %q, "tcp_port": 20000, "last_seen": %q},
    {"subdomain": "fresh", "token_hash": %q, "tcp_port": 20001, "last_seen": %q}
  ]
}
`,
		HashToken(storeOwnerToken), now.Add(-ReservationTTL-24*time.Hour).Format(time.RFC3339),
		HashToken(storeSecondToken), now.Add(-ReservationTTL+24*time.Hour).Format(time.RFC3339))
	if err := os.WriteFile(path, []byte(doc), 0o600); err != nil {
		t.Fatalf("writing the store by hand: %v", err)
	}

	r := mustOpen(t, path)
	t.Cleanup(func() { r.Close() })

	if _, ok := r.Get("stale"); ok {
		t.Error("a reservation idle longer than ReservationTTL survived the load")
	}
	if holder := r.PortHolder(20000); holder != "" {
		t.Errorf("the swept name left port 20000 held by %q — a number nobody can be given", holder)
	}
	fresh, ok := r.Get("fresh")
	if !ok {
		t.Fatal("a reservation inside ReservationTTL was swept")
	}
	if fresh.TCPPort != 20001 || r.PortHolder(20001) != "fresh" {
		t.Errorf("the surviving name lost its port: record %d, holder %q", fresh.TCPPort, r.PortHolder(20001))
	}
	// And the swept name is genuinely free: a stranger may now take it.
	if _, err := r.Claim("stale", storeStrangerToken, 0); err != nil {
		t.Errorf("a swept name was not claimable afterwards: %v", err)
	}
}

// D-4 · The cap refuses a new name and never an existing owner.
//
// Fails when the cap is applied to every Claim, so a full set locks its own
// owners out — the property this whole spec exists to establish, destroyed by
// the guard added to bound it.
func TestTheCapRefusesANewNameAndNotAnOwner(t *testing.T) {
	// In memory on purpose: the cap lives in Claim, not in the store, and
	// MaxReservations full-document writes would measure fsync rather than
	// the rule.
	r := NewReservations()
	for i := 0; i < MaxReservations; i++ {
		if _, err := r.Claim(fmt.Sprintf("name%d", i), storeOwnerToken, 0); err != nil {
			t.Fatalf("filling the set stopped at %d: %v", i, err)
		}
	}
	if r.Len() != MaxReservations {
		t.Fatalf("the set holds %d names, want %d", r.Len(), MaxReservations)
	}

	if _, err := r.Claim("onemore", storeSecondToken, 0); !errors.Is(err, ErrReservationsFull) {
		t.Errorf("a new name against a full set got %v, want ErrReservationsFull", err)
	}
	// The half that matters: an owner coming back is never refused by the cap.
	if _, err := r.Claim("name0", storeOwnerToken, 0); err != nil {
		t.Errorf("an existing owner was locked out by the cap: %v", err)
	}
	// And a wrong token is still refused for its own reason, not the cap's.
	if _, err := r.Claim("name0", storeStrangerToken, 0); !errors.Is(err, ErrNameReserved) {
		t.Errorf("a stranger on a full set got %v, want ErrNameReserved", err)
	}
}

// D-5 · A corrupt file yields an empty set, and the damaged bytes survive.
//
// Fails when OpenReservations returns an error instead of a set, so a damaged
// bookkeeping file becomes a total outage of every path including the
// anonymous one; or it starts empty and the first write destroys the only
// evidence of what went wrong.
func TestACorruptStoreStartsEmptyAndKeepsTheEvidence(t *testing.T) {
	path := storePath(t)
	damaged := []byte("{\"version\": 1, \"reservations\": [ this is not json")
	if err := os.WriteFile(path, damaged, 0o600); err != nil {
		t.Fatalf("writing the damaged store: %v", err)
	}

	r, err := OpenReservations(path, nil)
	if err != nil {
		t.Fatalf("a corrupt store refused to open, which would take the anonymous path down with it: %v", err)
	}
	t.Cleanup(func() { r.Close() })
	if r.Len() != 0 {
		t.Errorf("a corrupt store loaded %d names", r.Len())
	}

	// The relay is usable: the degradation is to the behaviour every release
	// before this one shipped, not to a hole.
	if _, err := r.Claim("myapi", storeOwnerToken, 0); err != nil {
		t.Fatalf("a name was not claimable after a corrupt load: %v", err)
	}

	// And the evidence is still on disk, under a name the next write cannot
	// reach. Checked by CONTENT, so a file that merely has the right name
	// does not pass.
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatalf("reading the store directory: %v", err)
	}
	found := ""
	for _, e := range entries {
		if strings.Contains(e.Name(), ".corrupt-") {
			found = filepath.Join(filepath.Dir(path), e.Name())
		}
	}
	if found == "" {
		t.Fatalf("the damaged bytes were not preserved; the directory holds %v", entries)
	}
	kept, err := os.ReadFile(found)
	if err != nil {
		t.Fatalf("reading %s: %v", found, err)
	}
	if string(kept) != string(damaged) {
		t.Errorf("%s does not hold the damaged bytes", found)
	}

	// A version from the future takes the same path, because guessing at a
	// format is worse than either available answer (§14.4).
	future := storePath(t)
	if err := os.WriteFile(future, []byte(`{"version":99,"reservations":[]}`), 0o600); err != nil {
		t.Fatalf("writing the future-version store: %v", err)
	}
	fr, err := OpenReservations(future, nil)
	if err != nil {
		t.Fatalf("a store from a future format version refused to open: %v", err)
	}
	t.Cleanup(func() { fr.Close() })
	if fr.Len() != 0 {
		t.Errorf("a store from a future format version loaded %d names", fr.Len())
	}
}

// D-6 · A second store over a held path is refused.
//
// Fails when nothing locks, so two relays each write the whole document and
// take turns silently destroying each other's claims.
func TestTwoStoresMayNotShareAPath(t *testing.T) {
	path := storePath(t)

	first := mustOpen(t, path)
	if _, err := first.Claim("myapi", storeOwnerToken, 0); err != nil {
		t.Fatalf("claiming: %v", err)
	}

	if second, err := OpenReservations(path, nil); !errors.Is(err, ErrStoreLocked) {
		if second != nil {
			second.Close()
		}
		t.Fatalf("a second store over a held path got %v, want ErrStoreLocked", err)
	}

	// And the lock goes back when the first store does — otherwise D-1's
	// second relay.New could never open the path the first one wrote.
	if err := first.Close(); err != nil {
		t.Fatalf("closing the first store: %v", err)
	}
	third, err := OpenReservations(path, nil)
	if err != nil {
		t.Fatalf("the lock was not released by Close: %v", err)
	}
	t.Cleanup(func() { third.Close() })
	if _, ok := third.Get("myapi"); !ok {
		t.Error("the third store did not load what the first one wrote")
	}
	// Closing twice is not an error: a relay's Close and a deferred one in
	// main are both allowed to run.
	if err := first.Close(); err != nil {
		t.Errorf("closing an already-closed store: %v", err)
	}
}

// Not an acceptance case: the document is the product's on-disk contract, so
// its shape is asserted directly rather than only through behaviour.
func TestTheStoreDocumentIsSortedAndVersioned(t *testing.T) {
	path := storePath(t)
	r := mustOpen(t, path)
	t.Cleanup(func() { r.Close() })
	for _, name := range []string{"zeta", "alpha", "mid"} {
		if _, err := r.Claim(name, storeOwnerToken, 0); err != nil {
			t.Fatalf("claiming %s: %v", name, err)
		}
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading the store: %v", err)
	}
	var doc struct {
		Version      int `json:"version"`
		Reservations []struct {
			Subdomain string `json:"subdomain"`
			TokenHash string `json:"token_hash"`
			Token     string `json:"token"`
		} `json:"reservations"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("the store is not the document it claims to be: %v", err)
	}
	if doc.Version != storeVersion {
		t.Errorf("the store says version %d, want %d", doc.Version, storeVersion)
	}
	var names []string
	for _, rec := range doc.Reservations {
		names = append(names, rec.Subdomain)
	}
	want := []string{"alpha", "mid", "zeta"}
	if strings.Join(names, ",") != strings.Join(want, ",") {
		t.Errorf("the store holds %v, want %v sorted", names, want)
	}
	// R-5's property, on disk this time: a leaked file is not a set of live
	// credentials.
	if strings.Contains(string(raw), storeOwnerToken) {
		t.Error("the store holds the token in the clear")
	}

	// And the mode, because the file holds digests.
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("the store is mode %o, want 600", perm)
	}
}

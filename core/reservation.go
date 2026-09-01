package core

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
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
// Nothing here does I/O. Making the set outlive the process is spec/0004
// slice 2 and is not built yet; what is built is ownership, which is what
// holds a name across the seconds an agent spends reconnecting.

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
// Every field here is read. There is deliberately no CreatedAt and no
// LastSeen: expiry belongs to the durable set (slice 2) and arrives with the
// code that reads it. Shipping a field ahead of its reader is the defect this
// type exists to repair, and doing it again one struct over would be a poor
// joke.
type Reservation struct {
	// Subdomain is the claimed name, lowercased.
	Subdomain string
	// TokenHash is the hex SHA-256 of the bearer secret. The secret itself is
	// never stored, so a leaked set is not a set of live credentials.
	TokenHash string
	// TCPPort is the public port claimed alongside the name; 0 for none.
	TCPPort int
}

// Reservations is the set of claimed names. Safe for concurrent use: the relay
// claims from a handshake while serving requests on many goroutines.
type Reservations struct {
	mu     sync.RWMutex
	byName map[string]Reservation
	byPort map[int]string // public TCP port -> owning subdomain
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

	if !exists {
		held = Reservation{Subdomain: name, TokenHash: HashToken(token)}
	}
	if tcpPort != 0 && held.TCPPort == 0 {
		held.TCPPort = tcpPort
		r.byPort[tcpPort] = name
	}
	r.byName[name] = held
	return !exists, nil
}

// Discard removes a claim.
//
// It is not a user-facing release and nothing outside the relay calls it: it
// exists so a handshake that claimed a name and then failed — the registry
// refused it as live, the public port would not bind, the announce could not be
// written — leaves nothing behind. A name is not consumed by a connection that
// never opened.
func (r *Reservations) Discard(subdomain string) {
	name := normaliseName(subdomain)
	r.mu.Lock()
	defer r.mu.Unlock()
	held, ok := r.byName[name]
	if !ok {
		return
	}
	if held.TCPPort != 0 {
		delete(r.byPort, held.TCPPort)
	}
	delete(r.byName, name)
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

// Len reports how many names are claimed.
func (r *Reservations) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.byName)
}

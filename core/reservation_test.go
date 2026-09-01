package core_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/pumasi-ai/pumasi-tunnel/core"
)

// Acceptance cases R-1 .. R-8 for spec/0004-names-with-owners.
// Frozen at spec review. See spec/0004-names-with-owners/acceptance/CASES.md.

// Long enough to claim with, per core.MinTokenLen.
const (
	ownerToken     = "owner-token-0123456789"
	strangerToken  = "stranger-token-9876543210"
	shortlishToken = "too-short"
)

// R-1 · A name claimed with one token is refused to a different token.
//
// Fails when Claim records nothing, or compares nothing, so the second token
// is accepted.
func TestAClaimedNameIsRefusedToAnotherToken(t *testing.T) {
	res := core.NewReservations()

	created, err := res.Claim("myapi", ownerToken, 0)
	if err != nil || !created {
		t.Fatalf("first claim: created=%v err=%v, want true, nil", created, err)
	}

	_, err = res.Claim("myapi", strangerToken, 0)
	if !errors.Is(err, core.ErrNameReserved) {
		t.Errorf("a stranger claiming myapi got %v, want ErrNameReserved", err)
	}
	if err := res.Check("myapi", strangerToken); !errors.Is(err, core.ErrNameReserved) {
		t.Errorf("a stranger checking myapi got %v, want ErrNameReserved", err)
	}

	// The owner still has it, and asking again does not re-create it.
	created, err = res.Claim("myapi", ownerToken, 0)
	if err != nil {
		t.Errorf("the owner reclaiming its own name: %v", err)
	}
	if created {
		t.Error("a second claim by the owner reported created=true; only the first call creates")
	}
}

// R-2 · A name claimed with a token is refused to a caller with no token.
//
// Fails when Check answers only about liveness, so an anonymous caller passes.
func TestAClaimedNameIsRefusedToAnAnonymousCaller(t *testing.T) {
	res := core.NewReservations()
	if _, err := res.Claim("myapi", ownerToken, 0); err != nil {
		t.Fatalf("claim: %v", err)
	}

	if err := res.Check("myapi", ""); !errors.Is(err, core.ErrNameReserved) {
		t.Errorf("an anonymous caller asking for myapi got %v, want ErrNameReserved", err)
	}
}

// R-3 · A name nobody has claimed is given to a caller with no token.
//
// Fails when a change makes a token mandatory for a named request — the
// withdrawal SPEC.md §6 forbids.
func TestAnUnclaimedNameIsStillAnonymous(t *testing.T) {
	res := core.NewReservations()
	if _, err := res.Claim("myapi", ownerToken, 0); err != nil {
		t.Fatalf("claim: %v", err)
	}

	// A different, unclaimed name is nobody's business.
	if err := res.Check("other", ""); err != nil {
		t.Errorf("an anonymous caller asking for the unclaimed \"other\" got %v, want nil", err)
	}
	// And on an empty set, every name is available.
	if err := core.NewReservations().Check("myapi", ""); err != nil {
		t.Errorf("an empty set refused %q: %v — an empty set is not a disabled one", "myapi", err)
	}
}

// R-4 · A token shorter than MinTokenLen cannot claim a name.
//
// Fails when short tokens are hashed and accepted, which is the assumption
// SPEC.md §3.1 rests SHA-256 on.
func TestAShortTokenCannotClaim(t *testing.T) {
	res := core.NewReservations()

	if len(shortlishToken) >= core.MinTokenLen {
		t.Fatalf("this case needs a token shorter than MinTokenLen=%d", core.MinTokenLen)
	}
	if _, err := res.Claim("myapi", shortlishToken, 0); !errors.Is(err, core.ErrTokenTooShort) {
		t.Errorf("claiming with a short token got %v, want ErrTokenTooShort", err)
	}
	if res.Len() != 0 {
		t.Errorf("a refused claim recorded %d reservations, want 0", res.Len())
	}
	// Check refuses it too, rather than quietly treating it as no token: being
	// handed a name you believe you protected is the worse failure.
	if err := res.Check("myapi", shortlishToken); !errors.Is(err, core.ErrTokenTooShort) {
		t.Errorf("checking with a short token got %v, want ErrTokenTooShort", err)
	}
}

// R-5 · The record holds a digest, never the secret.
//
// Fails when the token is stored in the clear, so a leaked set is a set of
// live credentials.
func TestAReservationDoesNotStoreTheToken(t *testing.T) {
	res := core.NewReservations()
	if _, err := res.Claim("myapi", ownerToken, 20000); err != nil {
		t.Fatalf("claim: %v", err)
	}

	held, ok := res.Get("myapi")
	if !ok {
		t.Fatal("Get found no reservation for a name just claimed")
	}
	if strings.Contains(held.TokenHash, ownerToken) || held.TokenHash == ownerToken {
		t.Errorf("the stored value contains the token itself: %q", held.TokenHash)
	}
	if held.TokenHash != core.HashToken(ownerToken) {
		t.Errorf("stored %q, want the hex sha256 of the token", held.TokenHash)
	}
	if len(held.TokenHash) != 64 {
		t.Errorf("stored a %d-character value, want a 64-character hex sha256", len(held.TokenHash))
	}
	if held.Subdomain != "myapi" || held.TCPPort != 20000 {
		t.Errorf("reservation is %+v, want myapi holding port 20000", held)
	}
}

// R-6 · A name claimed with port P refuses a claim for the same name at a
// different port, and a reconnect naming no port keeps P.
//
// Fails when a reservation accumulates ports, so one token can drain the pool
// one reconnect at a time; or when tcpPort == 0 is read as "give it up" and an
// address is lost to an omitted argument.
func TestAReservationIsOneAddress(t *testing.T) {
	res := core.NewReservations()
	if _, err := res.Claim("sshlike", ownerToken, 20000); err != nil {
		t.Fatalf("claim: %v", err)
	}

	if _, err := res.Claim("sshlike", ownerToken, 20001); !errors.Is(err, core.ErrPortReserved) {
		t.Errorf("claiming a second port under one name got %v, want ErrPortReserved", err)
	}
	if held, _ := res.Get("sshlike"); held.TCPPort != 20000 {
		t.Errorf("after the refusal the reservation holds port %d, want 20000 untouched", held.TCPPort)
	}
	if owner := res.PortHolder(20001); owner != "" {
		t.Errorf("the refused port 20001 is recorded as held by %q, want nobody", owner)
	}

	// Reconnecting as a plain HTTP tunnel must not give the port up: nothing
	// was said about it, so nothing changes.
	if _, err := res.Claim("sshlike", ownerToken, 0); err != nil {
		t.Fatalf("reclaiming without naming a port: %v", err)
	}
	if held, _ := res.Get("sshlike"); held.TCPPort != 20000 {
		t.Errorf("an omitted port argument dropped the claim to %d, want 20000 kept", held.TCPPort)
	}
	if owner := res.PortHolder(20000); owner != "sshlike" {
		t.Errorf("port 20000 is held by %q, want \"sshlike\"", owner)
	}
}

// R-7 · A port claimed by one name cannot be claimed by a second name.
//
// Fails when Claim checks only the name, so two reservations hold one number
// and whichever agent connects second is bound to a port it does not own.
func TestAPortBelongsToOneName(t *testing.T) {
	res := core.NewReservations()
	if _, err := res.Claim("first", ownerToken, 20000); err != nil {
		t.Fatalf("claim: %v", err)
	}

	// Even the same token may not point a second name at the same port.
	if _, err := res.Claim("second", ownerToken, 20000); !errors.Is(err, core.ErrPortReserved) {
		t.Errorf("a second name claiming port 20000 got %v, want ErrPortReserved", err)
	}
	if _, err := res.Claim("third", strangerToken, 20000); !errors.Is(err, core.ErrPortReserved) {
		t.Errorf("a stranger claiming port 20000 got %v, want ErrPortReserved", err)
	}
	if owner := res.PortHolder(20000); owner != "first" {
		t.Errorf("port 20000 is held by %q, want \"first\"", owner)
	}
	if _, ok := res.Get("second"); ok {
		t.Error("a refused claim left a reservation for \"second\" behind")
	}
}

// R-8 · A claim recorded by a handshake that then fails is discarded.
//
// Fails when a failed handshake consumes the name permanently, so a bind
// failure on a first connection locks a name to a token that never opened a
// tunnel.
func TestADiscardedClaimLeavesNothing(t *testing.T) {
	res := core.NewReservations()
	if _, err := res.Claim("myapi", ownerToken, 20000); err != nil {
		t.Fatalf("claim: %v", err)
	}

	res.Discard("myapi")

	if _, ok := res.Get("myapi"); ok {
		t.Error("the name is still claimed after Discard")
	}
	if owner := res.PortHolder(20000); owner != "" {
		t.Errorf("the port is still held by %q after Discard; a discarded claim must take its port with it", owner)
	}
	if err := res.Check("myapi", ""); err != nil {
		t.Errorf("an anonymous caller is still refused a discarded name: %v", err)
	}
	// Discarding what nobody claimed is a no-op, not a panic: the caller's
	// intent is already true.
	res.Discard("never-claimed")
}

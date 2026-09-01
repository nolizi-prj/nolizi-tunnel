package relay

// An in-package test, deliberately. What it asserts is a contract of an
// unexported method, and the hazard it guards cannot be reached through real
// connections — spec/0004 §11 records the measurement that established that,
// and why a race that never fires is not evidence (L-006).

import (
	"testing"

	"github.com/pumasi-ai/pumasi-tunnel/core"
)

const discardTestToken = "owner-token-0123456789"

// discardClaim must not pull a reservation out from under a tunnel that is
// registered on it, and must still discard one that nothing is using.
//
// The hazard: two connections presenting the SAME token for the same name.
// The first to Claim creates the reservation and carries newClaim; the second
// finds it already its own and carries "". Whichever loses the registry's
// ErrNameTaken calls discardClaim — and if that is the one that created the
// claim, an unguarded discard destroys the reservation while the winner is
// registered on it. The winner's tunnel would then report Reserved with
// nothing behind it, and its name would be free the instant it dropped, which
// is the whole defect spec/0004 closes.
func TestDiscardClaimWillNotUnderrunALiveTunnel(t *testing.T) {
	r, err := New(Config{BaseDomain: "pumasi.link"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if _, err := r.reservations.Claim("myapi", discardTestToken, 0); err != nil {
		t.Fatalf("Claim: %v", err)
	}
	// The winner of the race, registered on the name.
	if err := r.registry.Register(core.Tunnel{
		Subdomain: "myapi", AgentID: "the-winner", Reserved: true,
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	// The loser undoes what it created. It must find the name in use and stop.
	r.discardClaim("myapi")

	if _, ok := r.reservations.Get("myapi"); !ok {
		t.Fatal("the loser discarded a reservation the winner is registered on; the winner's name is now free the moment it drops")
	}
	if err := r.reservations.Check("myapi", ""); err == nil {
		t.Error("an anonymous caller is no longer refused myapi")
	}

	// With nothing registered, the same call does discard — the guard is a
	// guard and not a disablement.
	r.registry.Unregister("myapi")
	r.discardClaim("myapi")

	if _, ok := r.reservations.Get("myapi"); ok {
		t.Error("discardClaim left a claim behind for a name nothing is using")
	}
}

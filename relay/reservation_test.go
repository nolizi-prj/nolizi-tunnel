package relay_test

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/pumasi-ai/pumasi-tunnel/core"
	"github.com/pumasi-ai/pumasi-tunnel/relay"
)

// Acceptance cases C-1 .. C-6 for spec/0004-names-with-owners.
// Frozen at spec review. See spec/0004-names-with-owners/acceptance/CASES.md.
//
// No case here contains a race. Each one sequences a disconnect and a second
// connection, and reaches the state between them by waiting on an observable
// fact — Registry().Has(name) going false — rather than on a clock. A case
// that raced that would sometimes be refused by the REGISTRY (the name is
// still live) instead of by the RESERVATION (the name is owned), and would
// then pass against the unfixed tree for a reason that is not its own. Every
// refusal below is asserted to name the sentinel it should — see refusalOf,
// which says why identity is not available at this boundary.

const (
	// Long enough to claim with; core.MinTokenLen is what makes that true.
	resOwnerToken    = "owner-token-0123456789"
	resStrangerToken = "stranger-token-9876543210"
)

// resPortBase is where this file's public TCP ports start, each case taking a
// block of ten. Below the kernel's ephemeral floor and clear of
// bindOrderBase (20500-20559) and tcpHarnessBase (21000-21039), for the
// reason bindorder_test.go states once (L-007).
const resPortBase = 21200

var resPortCursor int32 = resPortBase

func resPorts() (low, high int) {
	low = int(atomic.AddInt32(&resPortCursor, 10)) - 10
	return low, low + 9
}

// resRelay starts a relay and an agent listener in front of it, and returns
// the address an agent dials.
func resRelay(t *testing.T, cfg relay.Config) (*relay.Relay, string) {
	t.Helper()
	if cfg.BaseDomain == "" {
		cfg.BaseDomain = "pumasi.link"
	}
	// Generous, because several cases hold a connection open across another
	// connection's whole handshake.
	if cfg.HandshakeTimeout == 0 {
		cfg.HandshakeTimeout = 60 * time.Second
	}
	r, err := relay.New(cfg)
	if err != nil {
		t.Fatalf("relay.New: %v", err)
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("agent listener: %v", err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go r.ServeAgent(conn)
		}
	}()
	return r, ln.Addr().String()
}

// handshake performs one agent handshake by hand and leaves the connection
// open. Doing it here rather than through agent.Agent is what makes these
// cases deterministic: agent.Run reconnects on its own schedule, and a case
// about what happens WHILE an agent is away cannot be written on top of
// something that decides for itself when to come back.
//
// The returned conn is the live tunnel; closing it is the disconnect.
func handshake(t *testing.T, addr string, req core.AuthRequest) (net.Conn, core.AuthResponse, error) {
	t.Helper()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial relay: %v", err)
	}
	frame, err := core.EncodeAuthRequest(req)
	if err != nil {
		t.Fatalf("EncodeAuthRequest: %v", err)
	}
	wire, err := frame.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if _, err := conn.Write(wire); err != nil {
		t.Fatalf("write handshake: %v", err)
	}
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	reply, err := core.DecodeFrame(conn)
	if err != nil {
		conn.Close()
		t.Fatalf("read handshake reply: %v", err)
	}
	conn.SetReadDeadline(time.Time{})
	resp, err := core.DecodeAuthResponse(reply)
	if err != nil {
		conn.Close()
		return nil, core.AuthResponse{}, err
	}
	return conn, resp, nil
}

// refusalOf reports whether a refusal that crossed the handshake is the one a
// sentinel names.
//
// errors.Is cannot be used here and the reason is the protocol, not the test:
// the relay sends its reason as text (core.ErrorFrame) and the agent side
// rebuilds an error from that text (core.DecodeAuthResponse), so no wrapped
// sentinel survives the wire. Matching the sentinel's own message is as close
// to identity as this boundary allows, and it moves if the sentinel's text
// moves rather than drifting away from it.
func refusalOf(err error, sentinel error) bool {
	return err != nil && strings.Contains(err.Error(), sentinel.Error())
}

// waitGone blocks until the relay has finished tearing a name down. This is
// the whole of how these cases avoid a race: ServeAgent's cleanup runs in a
// deferred function after the session closes, so a second agent that dialled
// immediately could arrive either side of it.
func waitGone(t *testing.T, r *relay.Relay, name string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if !r.Registry().Has(name) {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatalf("the relay still has %q registered 10s after its connection closed", name)
}

// C-1 · A name survives its owner's disconnect.
//
// Fails when the deferred registry.UnregisterAgent is the only thing that
// happens on disconnect, so the name is free and the anonymous agent IS given
// it. That is the unfixed tree, and it is skk6g7tyrs beside sshsteward on the
// live relay.
func TestAClaimedNameIsHeldAcrossADisconnect(t *testing.T) {
	r, addr := resRelay(t, relay.Config{})

	conn, resp, err := handshake(t, addr, core.AuthRequest{Subdomain: "myapi", Token: resOwnerToken})
	if err != nil {
		t.Fatalf("the owner's first connection was refused: %v", err)
	}
	if resp.Subdomain != "myapi" {
		t.Fatalf("the owner got %q, want \"myapi\"", resp.Subdomain)
	}

	// The disconnect, and the relay finishing with it.
	conn.Close()
	waitGone(t, r, "myapi")

	// While the owner is away, a stranger asks for its name. Two shapes, and
	// both must be refused for the same reason.
	if _, _, err := handshake(t, addr, core.AuthRequest{Subdomain: "myapi"}); !refusalOf(err, core.ErrNameReserved) {
		t.Errorf("an anonymous agent asking for myapi got %v, want ErrNameReserved", err)
	}
	if _, _, err := handshake(t, addr, core.AuthRequest{Subdomain: "myapi", Token: resStrangerToken}); !refusalOf(err, core.ErrNameReserved) {
		t.Errorf("a stranger with a token asking for myapi got %v, want ErrNameReserved", err)
	}
	if r.Registry().Has("myapi") {
		t.Fatal("a refused agent was nonetheless registered on myapi")
	}

	// And the owner gets it back.
	conn2, resp2, err := handshake(t, addr, core.AuthRequest{Subdomain: "myapi", Token: resOwnerToken})
	if err != nil {
		t.Fatalf("the owner was refused its own name on reconnect: %v", err)
	}
	defer conn2.Close()
	if resp2.Subdomain != "myapi" {
		t.Errorf("the owner reconnected onto %q, want \"myapi\"", resp2.Subdomain)
	}
}

// C-2 · A public TCP port survives its owner's disconnect.
//
// Fails when releaseTCP -> ReleaseOwner returns P to the free pool, so the
// next generic Allocate walks straight onto it and answers the anonymous agent
// with the steward's address.
func TestAClaimedPortIsHeldAcrossADisconnect(t *testing.T) {
	low, _ := resPorts()
	// A range of exactly one, so "the pool refused" and "the pool happened to
	// pick something else" cannot be confused.
	r, addr := resRelay(t, relay.Config{
		TCPPortLow: low, TCPPortHigh: low, TCPBindHost: "127.0.0.1",
	})

	conn, resp, err := handshake(t, addr, core.AuthRequest{
		Subdomain: "sshlike", Token: resOwnerToken, TCP: true, TCPPort: low,
	})
	if err != nil {
		t.Fatalf("the owner's first connection was refused: %v", err)
	}
	if want := fmt.Sprintf("pumasi.link:%d", low); resp.TCPAddr != want {
		t.Fatalf("the owner got %q, want %q", resp.TCPAddr, want)
	}

	conn.Close()
	waitGone(t, r, "sshlike")

	// A stranger asking for ANY public port must not be handed this one. On a
	// one-port range the honest answer is that the pool is empty.
	if _, _, err := handshake(t, addr, core.AuthRequest{TCP: true}); err == nil {
		t.Error("a stranger asking for any TCP port was handed the one held by sshlike")
	} else if !strings.Contains(err.Error(), "no free public TCP port") {
		t.Errorf("a stranger asking for any TCP port got %v, want the pool reported empty", err)
	}
	// And a stranger naming the number is refused too.
	if _, _, err := handshake(t, addr, core.AuthRequest{TCP: true, TCPPort: low}); err == nil {
		t.Error("a stranger naming the held port was given it")
	}

	// The owner gets the same address back.
	conn2, resp2, err := handshake(t, addr, core.AuthRequest{
		Subdomain: "sshlike", Token: resOwnerToken, TCP: true, TCPPort: low,
	})
	if err != nil {
		t.Fatalf("the owner was refused its own port on reconnect: %v", err)
	}
	defer conn2.Close()
	if want := fmt.Sprintf("pumasi.link:%d", low); resp2.TCPAddr != want {
		t.Errorf("the owner reconnected onto %q, want %q", resp2.TCPAddr, want)
	}
}

// C-3 · Nothing is withdrawn from the anonymous case.
//
// Fails when a change makes tokens mandatory for named requests, or makes
// generated names pass through the reservation set. SPEC.md §6's first column
// is the specification of this case.
func TestTheAnonymousCaseStillWorks(t *testing.T) {
	r, addr := resRelay(t, relay.Config{})

	owned, _, err := handshake(t, addr, core.AuthRequest{Subdomain: "myapi", Token: resOwnerToken})
	if err != nil {
		t.Fatalf("claiming myapi: %v", err)
	}
	defer owned.Close()

	// No name at all: still assigned one, as on any relay before this change.
	anon, respAnon, err := handshake(t, addr, core.AuthRequest{LocalPort: 8080})
	if err != nil {
		t.Fatalf("an anonymous agent asking for nothing was refused: %v", err)
	}
	defer anon.Close()
	if len(respAnon.Subdomain) != core.GeneratedSubdomainLen {
		t.Errorf("assigned name %q, want a generated one", respAnon.Subdomain)
	}

	// An unclaimed name, asked for by name, with no token: still granted.
	named, respNamed, err := handshake(t, addr, core.AuthRequest{Subdomain: "other"})
	if err != nil {
		t.Fatalf("an anonymous agent asking for the unclaimed \"other\" was refused: %v", err)
	}
	defer named.Close()
	if respNamed.Subdomain != "other" {
		t.Errorf("got %q, want \"other\"", respNamed.Subdomain)
	}

	if r.Registry().Len() != 3 {
		t.Errorf("%d tunnels are registered, want 3 — the owner and both anonymous agents", r.Registry().Len())
	}
}

// C-4 · The routing table carries who owns a name, while the public status
// surface does not disclose that ownership.
//
// Fails when Tunnel.Reserved is computed from the shape of the request —
// req.Subdomain != "" && req.Token != "" — which cannot tell a claimed name
// from one someone merely asked for, and when the status view has no such
// field at all. Both are the state before spec/0004.
//
// Narrowed from its first version on a cited spec-review objection; SPEC.md §9
// says why the other half is a review check and not a case.
func TestOwnershipStaysInTheRoutingTable(t *testing.T) {
	r, addr := resRelay(t, relay.Config{})

	owned, _, err := handshake(t, addr, core.AuthRequest{Subdomain: "myapi", Token: resOwnerToken})
	if err != nil {
		t.Fatalf("claiming myapi: %v", err)
	}
	defer owned.Close()

	// Asked for by name, by nobody in particular. This is the tunnel the old
	// expression could not tell apart from the one above.
	asked, _, err := handshake(t, addr, core.AuthRequest{Subdomain: "other"})
	if err != nil {
		t.Fatalf("asking for other: %v", err)
	}
	defer asked.Close()

	// The routing table's answer.
	claimed, err := r.Registry().Lookup("myapi.pumasi.link")
	if err != nil {
		t.Fatalf("Lookup myapi: %v", err)
	}
	if !claimed.Reserved {
		t.Error("the routing table says the claimed name myapi belongs to nobody")
	}
	anon, err := r.Registry().Lookup("other.pumasi.link")
	if err != nil {
		t.Fatalf("Lookup other: %v", err)
	}
	if anon.Reserved {
		t.Error("the routing table says an unclaimed name someone merely asked for is reserved")
	}
	if !anon.Requested {
		t.Error("a name asked for by name is not marked Requested; Reserved and Requested are not synonyms")
	}

	// Ownership is routing state, not a public directory.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/_pumasi/status", nil)
	req.Host = "pumasi.link"
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status returned %d", rec.Code)
	}
	for _, private := range []string{"myapi", "other", "reserved", "tunnels"} {
		if strings.Contains(rec.Body.String(), private) {
			t.Errorf("public status exposes %q: %s", private, rec.Body.String())
		}
	}
}

// C-5 · Both ingress paths obey one reservation.
//
// Fails when the check is added to ServeAgent rather than to the authorize
// both paths share — a claim true of one path (L-009).
func TestSSHIsRefusedAClaimedName(t *testing.T) {
	r, addr := resRelay(t, relay.Config{})

	owned, _, err := handshake(t, addr, core.AuthRequest{Subdomain: "myapi", Token: resOwnerToken})
	if err != nil {
		t.Fatalf("claiming myapi: %v", err)
	}
	defer owned.Close()

	sshLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("ssh listener: %v", err)
	}
	defer sshLn.Close()

	pem, err := relay.GenerateHostKeyPEM()
	if err != nil {
		t.Fatalf("GenerateHostKeyPEM: %v", err)
	}
	signer, err := ssh.ParsePrivateKey(pem)
	if err != nil {
		t.Fatalf("ParsePrivateKey: %v", err)
	}
	go func() {
		for {
			conn, err := sshLn.Accept()
			if err != nil {
				return
			}
			go r.ServeSSH(conn, signer)
		}
	}()

	// An ssh client asks for the claimed name in its username. It cannot send
	// a token — the username grammar has nowhere to put one — so it must be
	// refused, and told so on its terminal rather than by an opaque hangup.
	client, err := ssh.Dial("tcp", sshLn.Addr().String(), &ssh.ClientConfig{
		User:            "myapi",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	})
	if err != nil {
		t.Fatalf("ssh dial: %v", err)
	}
	defer client.Close()

	ch, reqs, err := client.OpenChannel("session", nil)
	if err != nil {
		t.Fatalf("opening session channel: %v", err)
	}
	go ssh.DiscardRequests(reqs)
	defer ch.Close()

	// An ssh.Channel has no deadline, so the read is bounded by a goroutine
	// and a select, as bindorder_test.go's greeting reader is.
	type readResult struct {
		n   int
		err error
	}
	buf := make([]byte, 512)
	done := make(chan readResult, 1)
	go func() {
		n, err := ch.Read(buf)
		done <- readResult{n, err}
	}()
	var told string
	select {
	case res := <-done:
		told = string(buf[:res.n])
	case <-time.After(30 * time.Second):
		t.Fatal("the ssh client was told nothing at all; a refusal must be readable on the terminal")
	}
	if !strings.Contains(told, core.ErrNameReserved.Error()) {
		t.Errorf("the ssh client was told %q, want a refusal naming the name as reserved", told)
	}

	// The decisive assertion: the claimed name still belongs to its owner and
	// the ssh client did not take it.
	claimed, err := r.Registry().Lookup("myapi.pumasi.link")
	if err != nil {
		t.Fatalf("Lookup myapi after the ssh attempt: %v", err)
	}
	if !claimed.Reserved {
		t.Error("after the ssh attempt the name is no longer reserved")
	}
	if r.Registry().Len() != 1 {
		t.Errorf("%d tunnels registered, want 1 — the ssh client opened one on a claimed name", r.Registry().Len())
	}
}

// C-6 · A short token is refused, not quietly ignored.
//
// Fails when the relay degrades a short token to no token, so a user who
// believes they are claiming a name is handed it unprotected and loses it at
// the next disconnect. SPEC.md §3.1 rests SHA-256 on this refusal existing.
func TestAShortTokenIsRefusedNotDowngraded(t *testing.T) {
	r, addr := resRelay(t, relay.Config{})

	short := "short"
	if len(short) >= core.MinTokenLen {
		t.Fatalf("this case needs a token shorter than MinTokenLen=%d", core.MinTokenLen)
	}

	_, _, err := handshake(t, addr, core.AuthRequest{Subdomain: "myapi", Token: short})
	if !refusalOf(err, core.ErrTokenTooShort) {
		t.Fatalf("an agent with a short token got %v, want ErrTokenTooShort", err)
	}
	if r.Registry().Has("myapi") {
		t.Error("the agent was refused and registered anyway")
	}

	// The refusal is the whole point: it must not have been handed the name as
	// an anonymous caller would have been.
	if _, ok := coreGet(r, "myapi"); ok {
		t.Error("a short token created a reservation")
	}
}

// coreGet asks the relay's own status surface whether a name is registered at
// all, which is the only reservation state this package can observe from
// outside. A name that was refused is neither registered nor reserved.
func coreGet(r *relay.Relay, name string) (core.Tunnel, bool) {
	t, err := r.Registry().Lookup(name + ".pumasi.link")
	return t, err == nil
}

// C-7 · A handshake that claims a name and then fails leaves nothing behind.
//
// Fails when the relay never calls Discard, so a name is consumed permanently
// by a connection that never opened. Added after the freeze on a cited spec
// review: R-8 supplies the Discard call itself, so discardClaim could be
// deleted from every call site and every other case would still pass
// (SPEC.md §12).
func TestAFailedHandshakeLeavesNoClaim(t *testing.T) {
	low, _ := resPorts()

	// Take the public port before the relay can, so the bind fails and the
	// handshake fails after the claim has been created. The pool does no I/O
	// and cannot know — core/portpool.go says so in its own header.
	squatter, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", low))
	if err != nil {
		t.Fatalf("could not occupy the public port for this case: %v", err)
	}
	defer squatter.Close()

	r, addr := resRelay(t, relay.Config{
		TCPPortLow: low, TCPPortHigh: low, TCPBindHost: "127.0.0.1",
	})

	_, _, err = handshake(t, addr, core.AuthRequest{
		Subdomain: "bindfail", Token: resOwnerToken, TCP: true, TCPPort: low,
	})
	if err == nil {
		t.Fatal("the handshake succeeded, but the public port was already bound; this case needs it to fail")
	}
	if !strings.Contains(err.Error(), "binding public port") {
		t.Fatalf("the handshake failed with %v, want a bind failure — this case must fail AFTER the claim", err)
	}

	// The decisive assertion: nobody owns the name. An anonymous agent gets it.
	conn, resp, err := handshake(t, addr, core.AuthRequest{Subdomain: "bindfail"})
	if err != nil {
		t.Fatalf("an anonymous agent was refused a name whose only claimant never opened a tunnel: %v", err)
	}
	defer conn.Close()
	if resp.Subdomain != "bindfail" {
		t.Errorf("the anonymous agent got %q, want \"bindfail\"", resp.Subdomain)
	}
	if tun, err := r.Registry().Lookup("bindfail.pumasi.link"); err != nil {
		t.Errorf("Lookup: %v", err)
	} else if tun.Reserved {
		t.Error("the anonymous agent's tunnel reports Reserved; the failed handshake's claim is still there")
	}
}

// C-8 · A claim is not consumed by losing to a squatter.
//
// Fails when the rollback guard tests "is anything registered on this name"
// instead of "is the live tunnel this claim's own": an anonymous agent holding
// an unclaimed name suppresses the discard, and the token-holder's failed
// handshake leaves a permanent claim on a name it never opened a tunnel on.
// Added after the freeze on a cited spec review (SPEC.md §13).
func TestAClaimIsNotConsumedByLosingToASquatter(t *testing.T) {
	r, addr := resRelay(t, relay.Config{})

	// An anonymous agent is sitting on an unclaimed name. This is the normal
	// state of every anonymous tunnel, not an exotic one.
	squatter, respSquat, err := handshake(t, addr, core.AuthRequest{Subdomain: "myapi"})
	if err != nil {
		t.Fatalf("the anonymous agent was refused: %v", err)
	}
	if respSquat.Subdomain != "myapi" {
		t.Fatalf("the anonymous agent got %q, want \"myapi\"", respSquat.Subdomain)
	}

	// A token-holder arrives for that name. Claim succeeds — nobody had
	// claimed it — and then Register refuses it as live.
	if _, _, err := handshake(t, addr, core.AuthRequest{Subdomain: "myapi", Token: resOwnerToken}); err == nil {
		t.Fatal("the token-holder was registered on a name already live")
	} else if !strings.Contains(err.Error(), core.ErrNameTaken.Error()) {
		t.Fatalf("the token-holder got %v, want the registry's ErrNameTaken", err)
	}

	// The claim that handshake created must be gone: it never opened a tunnel.
	squatter.Close()
	waitGone(t, r, "myapi")

	conn, resp, err := handshake(t, addr, core.AuthRequest{Subdomain: "myapi"})
	if err != nil {
		t.Fatalf("an anonymous agent is refused myapi by a claim whose handshake never opened a tunnel: %v", err)
	}
	defer conn.Close()
	if resp.Subdomain != "myapi" {
		t.Errorf("got %q, want \"myapi\"", resp.Subdomain)
	}
}

// C-9 · A discarded claim gives its public port back to the pool.
//
// Fails when discardClaim drops the reservation but leaves the pool's hold, so
// a number is unallocatable for the life of the relay with nothing owning it —
// the held state made permanent by the path that exists to undo it.
func TestADiscardedClaimReturnsItsPort(t *testing.T) {
	low, _ := resPorts()

	squatter, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", low))
	if err != nil {
		t.Fatalf("could not occupy the public port for this case: %v", err)
	}

	_, addr := resRelay(t, relay.Config{
		TCPPortLow: low, TCPPortHigh: low, TCPBindHost: "127.0.0.1",
	})

	// The bind fails, so the handshake fails after the claim was created.
	if _, _, err := handshake(t, addr, core.AuthRequest{
		Subdomain: "bindfail", Token: resOwnerToken, TCP: true, TCPPort: low,
	}); err == nil {
		t.Fatal("the handshake succeeded; this case needs the bind to fail")
	}

	// Free the port and let an unrelated anonymous agent ask for one. If the
	// hold outlived the discarded claim, the pool reports itself empty.
	squatter.Close()

	conn, resp, err := handshake(t, addr, core.AuthRequest{TCP: true})
	if err != nil {
		t.Fatalf("the port is still held by a claim that was discarded: %v", err)
	}
	defer conn.Close()
	if want := fmt.Sprintf("pumasi.link:%d", low); resp.TCPAddr != want {
		t.Errorf("got %q, want %q", resp.TCPAddr, want)
	}
}

// ─── Slice 2 · durability ────────────────────────────────────────────────────
//
// D-1 and D-7 for spec/0004-names-with-owners, acceptance/CASES.md. D-1 was
// frozen with the slice-1 cases and left unbuilt on purpose; it had no Go test
// in this tree at all until this packet, which is why a suite that passed 425
// of 425 said nothing whatever about the relay-restart half. An absent case
// cannot fail.
//
// WHAT MAKES D-1 D-1 AND NOT C-1 AGAIN: the second relay is a second
// relay.New over the same PATH, not a second handshake against the same
// process. A reconnect proves slice 1 and is evidence about the LEFT column of
// SPEC.md §4; only a second construction is evidence about the middle one. If
// this test is ever rewritten to reuse `r`, it has stopped being this case.

// D-1 · A reservation outlives the process.
//
// Fails when the reservation set is in memory, so the second relay starts
// empty and answers the anonymous agent yes.
func TestAReservationOutlivesTheRelay(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reservations.json")
	low, _ := resPorts()
	// A range of exactly one, for C-2's reason: "the pool refused" and "the
	// pool happened to pick something else" cannot then be confused.
	cfg := relay.Config{
		ReservationsPath: path,
		TCPPortLow:       low, TCPPortHigh: low, TCPBindHost: "127.0.0.1",
	}

	// ── The first process's worth of state.
	r1, addr1 := resRelay(t, cfg)
	conn, resp, err := handshake(t, addr1, core.AuthRequest{
		Subdomain: "myapi", Token: resOwnerToken, TCP: true, TCPPort: low,
	})
	if err != nil {
		t.Fatalf("the owner's first connection was refused: %v", err)
	}
	if want := fmt.Sprintf("pumasi.link:%d", low); resp.TCPAddr != want {
		t.Fatalf("the owner got %q, want %q", resp.TCPAddr, want)
	}
	conn.Close()
	waitGone(t, r1, "myapi")

	// The shutdown. Everything the first relay held — the lock on the store
	// included — goes back here, and the next line is a different process's
	// worth of state in every respect this test can reach.
	if err := r1.Close(); err != nil {
		t.Fatalf("closing the first relay: %v", err)
	}

	// ── A second relay.New over the same path. NOT a reconnect.
	r2, addr2 := resRelay(t, cfg)
	t.Cleanup(func() { r2.Close() })

	// The name is still owned, to both shapes of stranger.
	if _, _, err := handshake(t, addr2, core.AuthRequest{Subdomain: "myapi"}); !refusalOf(err, core.ErrNameReserved) {
		t.Errorf("after a restart, an anonymous agent asking for myapi got %v, want ErrNameReserved", err)
	}
	if _, _, err := handshake(t, addr2, core.AuthRequest{Subdomain: "myapi", Token: resStrangerToken}); !refusalOf(err, core.ErrNameReserved) {
		t.Errorf("after a restart, a stranger with a token asking for myapi got %v, want ErrNameReserved", err)
	}
	// And so is the port. On a one-port range the honest answer to a stranger
	// asking for any public port is that the pool is empty.
	if _, _, err := handshake(t, addr2, core.AuthRequest{TCP: true}); err == nil {
		t.Error("after a restart, a stranger asking for any TCP port was handed the one held by myapi")
	} else if !strings.Contains(err.Error(), "no free public TCP port") {
		t.Errorf("after a restart, a stranger asking for any TCP port got %v, want the pool reported empty", err)
	}

	// The owner gets both back from a relay that never saw it connect.
	conn2, resp2, err := handshake(t, addr2, core.AuthRequest{
		Subdomain: "myapi", Token: resOwnerToken, TCP: true, TCPPort: low,
	})
	if err != nil {
		t.Fatalf("after a restart, the owner was refused its own name: %v", err)
	}
	defer conn2.Close()
	if resp2.Subdomain != "myapi" {
		t.Errorf("the owner came back onto %q, want \"myapi\"", resp2.Subdomain)
	}
	if want := fmt.Sprintf("pumasi.link:%d", low); resp2.TCPAddr != want {
		t.Errorf("the owner came back onto %q, want %q", resp2.TCPAddr, want)
	}
}

// D-7 · An empty ReservationsPath is exactly today's relay.
//
// Fails when the store is opened unconditionally, so a relay run without the
// flag writes a file, takes a lock, or refuses to start — a change to the
// behaviour of every existing deployment, made by a flag nobody set.
func TestNoReservationsPathIsTodaysRelay(t *testing.T) {
	dir := t.TempDir()
	r, addr := resRelay(t, relay.Config{}) // no ReservationsPath
	t.Cleanup(func() { r.Close() })

	// C-1's sequence, unchanged: ownership still works without a store.
	conn, _, err := handshake(t, addr, core.AuthRequest{Subdomain: "myapi", Token: resOwnerToken})
	if err != nil {
		t.Fatalf("the owner's connection was refused: %v", err)
	}
	conn.Close()
	waitGone(t, r, "myapi")
	if _, _, err := handshake(t, addr, core.AuthRequest{Subdomain: "myapi"}); !refusalOf(err, core.ErrNameReserved) {
		t.Errorf("an anonymous agent asking for myapi got %v, want ErrNameReserved", err)
	}

	// And nothing was written anywhere. The directory is the one a store
	// would have had to use if a path had been given; there was none, so it
	// stays as empty as it started.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading the temp dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("a relay with no ReservationsPath left %d file(s) behind: %v", len(entries), entries)
	}
}

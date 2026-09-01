package relay_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pumasi-ai/pumasi-tunnel/agent"
	"github.com/pumasi-ai/pumasi-tunnel/core"
	"github.com/pumasi-ai/pumasi-tunnel/relay"
)

// Frozen acceptance cases for spec/0003-session-before-announce.
//
// The defect is an ordering, and spec/0002's cases showed how to observe one
// without timing: make the two orderings differ in what the client ends up
// holding rather than in when it holds it. There is no second step to make
// fail here, so these cases use the other half of the same idea — they hold
// the relay still at the exact instant under test.
//
// gatedConn stops the relay inside the write that carries the auth response.
// While it is stopped the relay has, by construction, registered the hostname
// and not yet returned from the announce, so a visitor sent in at that moment
// is testing precisely the window spec/0003 §1 is about. Nothing sleeps to
// find the window and nothing dials hoping to lose a race (L-006): the relay
// cannot proceed until the case lets it.

// announceHold is how long a case waits to see whether a visitor is answered
// while the announce is held. It is not a window the defect has to be hit
// inside — at 8540b89 the 404 is produced by a relay that is provably still
// blocked, so the case goes red immediately and this bound only lengthens a
// passing run. It must stay well under the relay's handshake deadline, which
// is still armed across the announce (spec/0003 §3.2) and which these cases
// set to sessionOrderHandshakeTimeout for headroom.
const (
	announceHold                 = 750 * time.Millisecond
	sessionOrderHandshakeTimeout = 60 * time.Second
	sessionOrderDomain           = "pumasi.link"
)

// gatedConn is the relay's end of an agent connection, with a brake on the
// first write. Every byte the relay sends — the auth response and every mux
// frame after it — passes through here, because ServeAgent hands this same
// value to mux.Server.
type gatedConn struct {
	net.Conn
	first      atomic.Bool
	announcing chan struct{} // closed as the relay enters the announce write
	release    chan struct{} // closed by the case to let the announce out
}

func newGatedConn(c net.Conn) *gatedConn {
	return &gatedConn{
		Conn:       c,
		announcing: make(chan struct{}),
		release:    make(chan struct{}),
	}
}

func (g *gatedConn) Write(b []byte) (int, error) {
	if g.first.CompareAndSwap(false, true) {
		close(g.announcing)
		<-g.release
	}
	return g.Conn.Write(b)
}

// sessionOrderFixture is one relay, one public edge, one local service and one
// real agent whose connection is gated. The agent is the product's own, dialled
// through Config.Dial onto a pipe, so nothing about the handshake is simulated.
type sessionOrderFixture struct {
	relay *relay.Relay
	edge  *httptest.Server
	gate  *gatedConn
	// connected receives the agent's auth response when its handshake
	// completes. Buffered: OnConnect runs on the agent's own read goroutine,
	// which must not block or the stream the relay is opening never lands.
	connected chan core.AuthResponse
	// entered receives one value as each visitor request reaches the relay's
	// handler. A case that has to know a visitor is *inside* ServeHTTP — and
	// not merely dispatched — waits on this rather than on a clock.
	entered   chan struct{}
	subdomain string
	host      string
}

func newSessionOrderFixture(t *testing.T, subdomain string) *sessionOrderFixture {
	t.Helper()

	local := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "local service served %s", r.URL.Path)
	}))
	t.Cleanup(local.Close)

	r, err := relay.New(relay.Config{
		BaseDomain:       sessionOrderDomain,
		HandshakeTimeout: sessionOrderHandshakeTimeout,
	})
	if err != nil {
		t.Fatalf("relay.New: %v", err)
	}

	entered := make(chan struct{}, 4)
	edge := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		entered <- struct{}{}
		r.ServeHTTP(w, req)
	}))
	t.Cleanup(edge.Close)

	agentSide, relaySide := net.Pipe()
	t.Cleanup(func() { agentSide.Close(); relaySide.Close() })
	gate := newGatedConn(relaySide)

	f := &sessionOrderFixture{
		relay:     r,
		edge:      edge,
		gate:      gate,
		connected: make(chan core.AuthResponse, 1),
		entered:   entered,
		subdomain: subdomain,
		host:      subdomain + "." + sessionOrderDomain,
	}

	ag, err := agent.New(agent.Config{
		RelayAddr: "pipe",
		LocalAddr: strings.TrimPrefix(local.URL, "http://"),
		Subdomain: subdomain,
		Dial: func(context.Context, string) (net.Conn, error) {
			return agentSide, nil
		},
		OnConnect: func(resp core.AuthResponse) { f.connected <- resp },
	})
	if err != nil {
		t.Fatalf("agent.New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go ag.Run(ctx)
	go r.ServeAgent(gate)

	return f
}

// visit issues one visitor request to the public edge with the tunnel's
// hostname, exactly as a browser would once DNS resolved. It reports on a
// channel so a case can ask whether it has been answered *yet*.
func (f *sessionOrderFixture) visit(path string) <-chan visitResult {
	out := make(chan visitResult, 1)
	go func() {
		req, err := http.NewRequest(http.MethodGet, f.edge.URL+path, nil)
		if err != nil {
			out <- visitResult{err: err}
			return
		}
		req.Host = f.host
		resp, err := f.edge.Client().Do(req)
		if err != nil {
			out <- visitResult{err: err}
			return
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		out <- visitResult{status: resp.StatusCode, body: string(body), err: err}
	}()
	return out
}

type visitResult struct {
	status int
	body   string
	err    error
}

// C-1 · A visitor that acts on the URL the instant it is announced is served,
// or waits — never told the tunnel is not there.
//
// Fails when the session is installed after the auth response is written: the
// visitor gets past registry.Lookup, finds no session, and is answered
// "404 No tunnel is open for ..." while the relay is still inside the announce.
func TestVisitorIsNotAnsweredBeforeTheSessionExists(t *testing.T) {
	f := newSessionOrderFixture(t, "sessionorder")

	// The relay is now inside the write that carries the URL, and cannot
	// leave it until this case says so. The hostname already resolves.
	<-f.gate.announcing

	visited := f.visit("/the-instant-it-arrived")

	// The request is inside the relay's handler, so what follows is a
	// statement about the relay and not about how fast a loopback socket is.
	select {
	case <-f.entered:
	case <-time.After(10 * time.Second):
		t.Fatal("the visitor request never reached the relay")
	}

	select {
	case got := <-visited:
		t.Fatalf("a visitor was answered %d %q while the relay was still inside the announce — "+
			"the URL is reachable before the session that serves it exists (spec/0003 §1)",
			got.status, strings.TrimSpace(got.body))
	case <-time.After(announceHold):
		// Correct: nothing can answer for this hostname yet, so nothing did.
	}

	close(f.gate.release)

	select {
	case resp := <-f.connected:
		if resp.URL != "http://"+f.host {
			t.Fatalf("announced URL = %q, want %q", resp.URL, "http://"+f.host)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the agent never completed its handshake after the announce was released")
	}

	select {
	case got := <-visited:
		if got.err != nil {
			t.Fatalf("visitor request failed: %v", got.err)
		}
		if got.status != http.StatusOK {
			t.Fatalf("visitor status = %d %q, want 200 — the request that waited for the "+
				"announce must then be served", got.status, strings.TrimSpace(got.body))
		}
		if !strings.Contains(got.body, "local service served /the-instant-it-arrived") {
			t.Fatalf("visitor body = %q, want the local service's answer", got.body)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the visitor was never answered after the announce completed")
	}
}

// C-2 · The announce is the first thing its connection carries, even when a
// visitor is forwarded during it.
//
// Fails when the session is installed before the announce with nothing keeping
// stream frames off the wire: the visitor's FrameOpen reaches the agent ahead
// of the auth response, DecodeAuthResponse rejects it, and the agent never
// reports a URL at all (spec/0003 §2).
func TestTheAnnounceReachesTheAgentBeforeAnyStream(t *testing.T) {
	f := newSessionOrderFixture(t, "announcefirst")

	<-f.gate.announcing

	// Forwarded — or blocked — while the announce is in flight. Which of the
	// two it is, is C-1's business, and this case asserts nothing about it:
	// its only claim is about what the agent ends up holding, which must be
	// its auth response and not a stream frame.
	visited := f.visit("/during-the-announce")
	<-f.entered
	time.Sleep(announceHold)
	close(f.gate.release)

	select {
	case resp := <-f.connected:
		if resp.Subdomain != f.subdomain {
			t.Fatalf("agent was told subdomain %q, want %q", resp.Subdomain, f.subdomain)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the agent never received an auth response it could decode — " +
			"something reached it ahead of the announce (spec/0003 §2)")
	}

	// Drained, not asserted on, so this case cannot go red for C-1's reason.
	// Reported because a visitor that never returns at all is a third outcome
	// and neither case would otherwise name it.
	select {
	case got := <-visited:
		t.Logf("the visitor sent during the announce was answered %d %q (err=%v)",
			got.status, strings.TrimSpace(got.body), got.err)
	case <-time.After(10 * time.Second):
		t.Fatal("the visitor sent during the announce was never answered at all")
	}
}

// gatedFailingConn refuses the write that carries the announce, but not
// before the case says so — so the relay can be held inside a *failing*
// announce, with the session installed and r.mu held, at the instant C-3 is
// about to ask a question about.
type gatedFailingConn struct {
	net.Conn
	first      atomic.Bool
	announcing chan struct{}
	release    chan struct{}
}

var errAnnounceRefused = errors.New("test: this connection cannot be written to")

func (c *gatedFailingConn) Write(b []byte) (int, error) {
	if c.first.CompareAndSwap(false, true) {
		close(c.announcing)
		<-c.release
		return 0, errAnnounceRefused
	}
	return c.Conn.Write(b)
}

// C-3 · An announce that fails leaves nothing behind, and leaves it behind
// atomically.
//
// A visitor is put inside ServeHTTP while the relay is held in the doomed
// announce, so it is waiting on r.mu — not racing for it — when the write
// fails. sync.RWMutex hands a released write lock to the readers already
// queued on it before it grants the next writer, so an implementation that
// deletes the session in a *second* critical section is guaranteed to show
// this visitor the session it is about to remove, and to forward it onto a
// connection already known to be dead. Under §3.3's single critical section
// the visitor cannot observe that state at all and is answered 404.
//
// Fails when the failure path does not undo the session install that now
// precedes it, or undoes it in a second critical section; or when the route
// or the subdomain outlives a handshake that never completed.
func TestAFailedAnnounceLeavesNothingBehind(t *testing.T) {
	local := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "local service served %s", r.URL.Path)
	}))
	t.Cleanup(local.Close)

	r, err := relay.New(relay.Config{
		BaseDomain:       sessionOrderDomain,
		HandshakeTimeout: sessionOrderHandshakeTimeout,
	})
	if err != nil {
		t.Fatalf("relay.New: %v", err)
	}
	entered := make(chan struct{}, 4)
	edge := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		entered <- struct{}{}
		r.ServeHTTP(w, req)
	}))
	t.Cleanup(edge.Close)

	const sub = "doomed"
	host := sub + "." + sessionOrderDomain

	visit := func(path string) <-chan visitResult {
		out := make(chan visitResult, 1)
		go func() {
			req, err := http.NewRequest(http.MethodGet, edge.URL+path, nil)
			if err != nil {
				out <- visitResult{err: err}
				return
			}
			req.Host = host
			resp, err := edge.Client().Do(req)
			if err != nil {
				out <- visitResult{err: err}
				return
			}
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			out <- visitResult{status: resp.StatusCode, body: string(body), err: err}
		}()
		return out
	}

	// A first agent whose announce can never be delivered.
	deadAgent, deadRelay := net.Pipe()
	t.Cleanup(func() { deadAgent.Close(); deadRelay.Close() })
	go func() {
		req, _ := core.EncodeAuthRequest(core.AuthRequest{Subdomain: sub, LocalPort: 8080})
		wire, _ := req.Encode()
		deadAgent.Write(wire)
	}()
	doomed := &gatedFailingConn{
		Conn:       deadRelay,
		announcing: make(chan struct{}),
		release:    make(chan struct{}),
	}
	done := make(chan struct{})
	go func() { r.ServeAgent(doomed); close(done) }()

	// The relay is inside the announce that is about to fail.
	<-doomed.announcing
	during := visit("/while-the-announce-fails")
	select {
	case <-entered:
	case <-time.After(10 * time.Second):
		t.Fatal("the visitor request never reached the relay")
	}
	// It is queued on r.mu, and cannot be answered while the relay holds it.
	select {
	case got := <-during:
		t.Fatalf("a visitor was answered %d %q while the relay held r.mu across the announce",
			got.status, strings.TrimSpace(got.body))
	case <-time.After(announceHold):
	}
	close(doomed.release)

	select {
	case got := <-during:
		if got.err != nil {
			t.Fatalf("visitor request failed: %v", got.err)
		}
		if got.status != http.StatusNotFound {
			t.Fatalf("the visitor waiting on r.mu was answered %d %q, want 404 — it was shown a "+
				"session on a connection whose announce had already failed (spec/0003 §3.3)",
				got.status, strings.TrimSpace(got.body))
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the visitor waiting on r.mu was never answered after the announce failed")
	}

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("ServeAgent did not return after its announce failed")
	}

	// Nothing answers for that hostname afterwards either.
	after := <-visit("/after")
	<-entered
	if after.err != nil {
		t.Fatalf("visitor request: %v", after.err)
	}
	if after.status != http.StatusNotFound {
		t.Fatalf("status = %d %q, want 404 after a handshake that never completed",
			after.status, strings.TrimSpace(after.body))
	}

	// And the name is free: a second agent takes it and is served.
	agentSide, relaySide := net.Pipe()
	t.Cleanup(func() { agentSide.Close(); relaySide.Close() })
	connected := make(chan core.AuthResponse, 1)
	ag, err := agent.New(agent.Config{
		RelayAddr: "pipe",
		LocalAddr: strings.TrimPrefix(local.URL, "http://"),
		Subdomain: sub,
		Dial:      func(context.Context, string) (net.Conn, error) { return agentSide, nil },
		OnConnect: func(resp core.AuthResponse) { connected <- resp },
	})
	if err != nil {
		t.Fatalf("agent.New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go ag.Run(ctx)
	go r.ServeAgent(relaySide)

	select {
	case got := <-connected:
		if got.Subdomain != sub {
			t.Fatalf("second agent got subdomain %q, want %q — the failed handshake "+
				"kept the name", got.Subdomain, sub)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("a second agent could not take the subdomain the failed announce left behind")
	}

	served := <-visit("/second")
	<-entered
	if served.err != nil {
		t.Fatalf("second visitor request: %v", served.err)
	}
	if served.status != http.StatusOK {
		t.Fatalf("second visitor status = %d %q, want 200", served.status, strings.TrimSpace(served.body))
	}
}

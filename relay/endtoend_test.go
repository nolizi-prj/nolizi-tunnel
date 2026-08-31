package relay_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pumasi-ai/pumasi-tunnel/agent"
	"github.com/pumasi-ai/pumasi-tunnel/core"
	"github.com/pumasi-ai/pumasi-tunnel/relay"
)

// harness wires the whole product together in one process: a local service,
// an agent forwarding to it, a relay accepting the agent, and a public edge
// server in front of the relay. Every hop is real — real TCP listeners, real
// HTTP — so a passing test means a request genuinely crossed the tunnel.
type harness struct {
	edge      *httptest.Server
	relay     *relay.Relay
	localAddr string
	subdomain string
	cancel    context.CancelFunc
}

func newHarness(t *testing.T, handler http.Handler) *harness {
	t.Helper()

	local := httptest.NewServer(handler)
	t.Cleanup(local.Close)
	localAddr := strings.TrimPrefix(local.URL, "http://")

	r, err := relay.New(relay.Config{BaseDomain: "pumasi.link"})
	if err != nil {
		t.Fatalf("relay.New: %v", err)
	}

	agentLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("agent listener: %v", err)
	}
	t.Cleanup(func() { agentLn.Close() })
	go func() {
		for {
			conn, err := agentLn.Accept()
			if err != nil {
				return
			}
			go r.ServeAgent(conn)
		}
	}()

	edge := httptest.NewServer(r)
	t.Cleanup(edge.Close)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	connected := make(chan core.AuthResponse, 1)
	ag, err := agent.New(agent.Config{
		RelayAddr: agentLn.Addr().String(),
		LocalAddr: localAddr,
		Subdomain: "myapp",
		OnConnect: func(resp core.AuthResponse) { connected <- resp },
	})
	if err != nil {
		t.Fatalf("agent.New: %v", err)
	}
	go ag.Run(ctx)

	select {
	case resp := <-connected:
		return &harness{edge: edge, relay: r, localAddr: localAddr, subdomain: resp.Subdomain, cancel: cancel}
	case <-time.After(10 * time.Second):
		t.Fatal("agent did not connect to the relay")
		return nil
	}
}

// get issues a request to the public edge with the tunnel's hostname, the way
// a visitor's browser would after DNS resolved.
func (h *harness) get(t *testing.T, path string) (*http.Response, string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, h.edge.URL+path, nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Host = h.subdomain + ".pumasi.link"

	resp, err := h.edge.Client().Do(req)
	if err != nil {
		t.Fatalf("request through the tunnel: %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("reading body: %v", err)
	}
	return resp, string(body)
}

// The headline test: a request to the public edge reaches a service that is
// only listening on loopback, with no inbound port open to it.
func TestRequestCrossesTheTunnel(t *testing.T) {
	h := newHarness(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "served %s to host %s", r.URL.Path, r.Host)
	}))

	resp, body := h.get(t, "/hello")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if !strings.Contains(body, "served /hello") {
		t.Errorf("body = %q", body)
	}
	// The visitor's Host must survive the trip: local apps route on it.
	if !strings.Contains(body, h.subdomain+".pumasi.link") {
		t.Errorf("the local service saw host %q, want the tunnel hostname", body)
	}
}

func TestRequestBodyAndStatusAndHeaders(t *testing.T) {
	h := newHarness(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("X-Local-Header", "from-local-service")
		w.WriteHeader(http.StatusTeapot)
		fmt.Fprintf(w, "method=%s received=%q", r.Method, body)
	}))

	req, err := http.NewRequest(http.MethodPost, h.edge.URL+"/submit", strings.NewReader("payload from visitor"))
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Host = h.subdomain + ".pumasi.link"

	resp, err := h.edge.Client().Do(req)
	if err != nil {
		t.Fatalf("POST through the tunnel: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusTeapot {
		t.Errorf("status = %d, want 418 — the local service's status must survive", resp.StatusCode)
	}
	if got := resp.Header.Get("X-Local-Header"); got != "from-local-service" {
		t.Errorf("X-Local-Header = %q, response headers did not survive", got)
	}
	if !strings.Contains(string(body), `received="payload from visitor"`) {
		t.Errorf("request body did not survive: %q", body)
	}
	if !strings.Contains(string(body), "method=POST") {
		t.Errorf("method did not survive: %q", body)
	}
}

// A response larger than one protocol frame must arrive whole — this is the
// path a file download takes.
func TestLargeResponseCrossesIntact(t *testing.T) {
	const size = core.MaxPayloadSize*2 + 4321
	payload := strings.Repeat("abcdefgh", size/8)

	h := newHarness(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, payload)
	}))

	resp, body := h.get(t, "/big")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if len(body) != len(payload) {
		t.Fatalf("got %d bytes, want %d", len(body), len(payload))
	}
	if body != payload {
		t.Error("large response was corrupted in transit")
	}
}

func TestUnknownHostIsNotFound(t *testing.T) {
	h := newHarness(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req, _ := http.NewRequest(http.MethodGet, h.edge.URL+"/", nil)
	req.Host = "nobody-is-here.pumasi.link"
	resp, err := h.edge.Client().Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	// The response a stranger sees must not carry an operator's identity.
	if strings.Contains(strings.ToLower(string(body)), "pumasi tunnel") {
		t.Errorf("404 body leaks branding: %q", body)
	}
}

// When the local service is down, the visitor must get a gateway error rather
// than a hang — the agent is connected, but nothing answers behind it.
func TestDeadLocalServiceGivesBadGateway(t *testing.T) {
	// Point the agent at a port nothing listens on.
	dead, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	deadAddr := dead.Addr().String()
	dead.Close() // now nothing is there

	r, err := relay.New(relay.Config{BaseDomain: "pumasi.link"})
	if err != nil {
		t.Fatalf("relay.New: %v", err)
	}
	agentLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer agentLn.Close()
	go func() {
		for {
			conn, err := agentLn.Accept()
			if err != nil {
				return
			}
			go r.ServeAgent(conn)
		}
	}()

	edge := httptest.NewServer(r)
	defer edge.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	connected := make(chan core.AuthResponse, 1)
	ag, err := agent.New(agent.Config{
		RelayAddr: agentLn.Addr().String(),
		LocalAddr: deadAddr,
		Subdomain: "broken",
		OnConnect: func(resp core.AuthResponse) { connected <- resp },
	})
	if err != nil {
		t.Fatalf("agent.New: %v", err)
	}
	go ag.Run(ctx)

	select {
	case <-connected:
	case <-time.After(10 * time.Second):
		t.Fatal("agent did not connect")
	}

	req, _ := http.NewRequest(http.MethodGet, edge.URL+"/", nil)
	req.Host = "broken.pumasi.link"
	resp, err := edge.Client().Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("status = %d, want 502 when nothing listens locally", resp.StatusCode)
	}
}

// Many visitors at once must not interfere with each other; each request is
// its own stream on the one agent connection.
func TestConcurrentVisitors(t *testing.T) {
	h := newHarness(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "path=%s", r.URL.Path)
	}))

	const visitors = 20
	errs := make(chan error, visitors)
	for i := 0; i < visitors; i++ {
		go func(i int) {
			req, _ := http.NewRequest(http.MethodGet, h.edge.URL+fmt.Sprintf("/visitor/%d", i), nil)
			req.Host = h.subdomain + ".pumasi.link"
			resp, err := h.edge.Client().Do(req)
			if err != nil {
				errs <- err
				return
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			want := fmt.Sprintf("path=/visitor/%d", i)
			if string(body) != want {
				errs <- fmt.Errorf("got %q, want %q — responses crossed between visitors", body, want)
				return
			}
			errs <- nil
		}(i)
	}
	for i := 0; i < visitors; i++ {
		if err := <-errs; err != nil {
			t.Error(err)
		}
	}
}

// A relay must refuse a name that would impersonate its own operator.
func TestReservedSubdomainRefused(t *testing.T) {
	r, err := relay.New(relay.Config{BaseDomain: "pumasi.link"})
	if err != nil {
		t.Fatalf("relay.New: %v", err)
	}
	agentLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer agentLn.Close()
	go func() {
		for {
			conn, err := agentLn.Accept()
			if err != nil {
				return
			}
			go r.ServeAgent(conn)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ag, err := agent.New(agent.Config{
		RelayAddr: agentLn.Addr().String(),
		LocalAddr: "127.0.0.1:1",
		Subdomain: "admin",
	})
	if err != nil {
		t.Fatalf("agent.New: %v", err)
	}

	done := make(chan error, 1)
	go func() { done <- ag.Run(ctx) }()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("Run returned nil; a reserved name must be refused")
		}
		if !strings.Contains(err.Error(), "reserved") {
			t.Errorf("error = %v, want it to state the name is reserved", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("agent neither connected nor was refused; a refusal must not be retried forever")
	}
}

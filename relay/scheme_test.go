package relay_test

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pumasi-ai/pumasi-tunnel/agent"
	"github.com/pumasi-ai/pumasi-tunnel/core"
	"github.com/pumasi-ai/pumasi-tunnel/relay"
	"golang.org/x/crypto/ssh"
)

// Acceptance cases for spec/0001-public-scheme, A-5 to A-10. Frozen at the
// spec review; spec/0001-public-scheme/acceptance/CASES.md says what
// execution makes each fail.
//
// The three surfaces a person actually reads an address on are exercised
// against a real relay here, not asserted about: the auth response an agent
// receives (which the CLI prints as its first line), the console's
// /_pumasi/status feed, and the greeting the ssh ingress writes into the
// terminal of someone who installed nothing.

// schemeHarness is one relay serving all three surfaces at a chosen scheme.
type schemeHarness struct {
	relay     *relay.Relay
	edge      *httptest.Server
	sshAddr   string
	subdomain string
	authURL   string // surface 1, as the agent received it
}

func newSchemeHarness(t *testing.T, scheme string) *schemeHarness {
	t.Helper()

	local := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	t.Cleanup(local.Close)

	r, err := relay.New(relay.Config{BaseDomain: "pumasi.link", PublicScheme: scheme})
	if err != nil {
		t.Fatalf("relay.New(%q): %v", scheme, err)
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

	// The ssh ingress, on its own listener, with a throwaway host key.
	pem, err := relay.GenerateHostKeyPEM()
	if err != nil {
		t.Fatalf("host key: %v", err)
	}
	signer, err := ssh.ParsePrivateKey(pem)
	if err != nil {
		t.Fatalf("parsing host key: %v", err)
	}
	sshLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("ssh listener: %v", err)
	}
	t.Cleanup(func() { sshLn.Close() })
	go func() {
		for {
			conn, err := sshLn.Accept()
			if err != nil {
				return
			}
			go r.ServeSSH(conn, signer)
		}
	}()

	edge := httptest.NewServer(r)
	t.Cleanup(edge.Close)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	connected := make(chan core.AuthResponse, 1)
	ag, err := agent.New(agent.Config{
		RelayAddr: agentLn.Addr().String(),
		LocalAddr: strings.TrimPrefix(local.URL, "http://"),
		Subdomain: "myapp",
		OnConnect: func(resp core.AuthResponse) { connected <- resp },
	})
	if err != nil {
		t.Fatalf("agent.New: %v", err)
	}
	go ag.Run(ctx)

	select {
	case resp := <-connected:
		return &schemeHarness{
			relay: r, edge: edge, sshAddr: sshLn.Addr().String(),
			subdomain: resp.Subdomain, authURL: resp.URL,
		}
	case <-time.After(10 * time.Second):
		t.Fatal("agent did not connect to the relay")
		return nil
	}
}

// sshTunnelName is the ssh client's username, which the ingress reads as the
// requested subdomain — so the ssh session's tunnel is a different one from
// the agent's, registered in the same registry.
const sshTunnelName = "sshuser"

// sshBanner connects with a real ssh client, the way a person with no client
// of ours installed does, and returns what the relay printed to the terminal.
func (h *schemeHarness) sshBanner(t *testing.T) string {
	t.Helper()
	client, err := ssh.Dial("tcp", h.sshAddr, &ssh.ClientConfig{
		User:            sshTunnelName,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	})
	if err != nil {
		t.Fatalf("ssh dial: %v", err)
	}
	t.Cleanup(func() { client.Close() })

	// A stock ssh client opens a session channel unless -N was passed; that
	// is the channel the relay writes its greeting to.
	ch, reqs, err := client.OpenChannel("session", nil)
	if err != nil {
		t.Fatalf("opening session channel: %v", err)
	}
	go ssh.DiscardRequests(reqs)
	defer ch.Close()

	// The greeting is one write; read what is there rather than to EOF,
	// because the relay deliberately holds the channel open afterwards.
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
	select {
	case res := <-done:
		if res.err != nil && res.n == 0 {
			t.Fatalf("reading the ssh greeting: %v", res.err)
		}
		return string(buf[:res.n])
	case <-time.After(10 * time.Second):
		t.Fatal("the ssh ingress printed no greeting")
		return ""
	}
}

// A-5 · Surface 1: the address the agent is handed at connect, which the CLI
// prints as its first line of output.
func TestAuthResponseCarriesTheRelayScheme(t *testing.T) {
	h := newSchemeHarness(t, "")
	if want := "http://" + h.subdomain + ".pumasi.link"; h.authURL != want {
		t.Errorf("default auth response URL = %q, want %q", h.authURL, want)
	}

	hs := newSchemeHarness(t, core.SchemeHTTPS)
	if want := "https://" + hs.subdomain + ".pumasi.link"; hs.authURL != want {
		t.Errorf("https auth response URL = %q, want %q", hs.authURL, want)
	}
}

// A-7 · Surface 3: the zero-install ssh banner, read off a real ssh client.
func TestSSHBannerCarriesTheRelayScheme(t *testing.T) {
	h := newSchemeHarness(t, "")
	banner := h.sshBanner(t)
	if !strings.Contains(banner, "http://") {
		t.Errorf("default ssh banner = %q, want an http:// address", banner)
	}
	if strings.Contains(banner, "https://") {
		t.Errorf("default ssh banner announced https, which nothing here serves: %q", banner)
	}

	hs := newSchemeHarness(t, core.SchemeHTTPS)
	bannerTLS := hs.sshBanner(t)
	if !strings.Contains(bannerTLS, "https://") {
		t.Errorf("https ssh banner = %q, want an https:// address", bannerTLS)
	}
}

// A-8 · The case a future copy-paste breaks: for one relay, both address
// surfaces agree with each other and with the registry that decided it. A
// surface that grows its own scheme fails here however it is spelled.
func TestAllThreeSurfacesAgreeOnTheScheme(t *testing.T) {
	for _, scheme := range []string{"", core.SchemeHTTP, core.SchemeHTTPS} {
		h := newSchemeHarness(t, scheme)
		decided := h.relay.Registry().PublicURL(h.subdomain)

		if h.authURL != decided {
			t.Errorf("scheme %q: agent was told %q, registry decided %q", scheme, h.authURL, decided)
		}
		// The ssh session asks for its own name, so the address it is greeted
		// with is the registry's decision for that name — same registry, same
		// scheme, decided in the same place.
		sshDecided := h.relay.Registry().PublicURL(sshTunnelName)
		if banner := h.sshBanner(t); !strings.Contains(banner, sshDecided) {
			t.Errorf("scheme %q: ssh banner %q does not carry %q", scheme, banner, sshDecided)
		}
		if !strings.HasPrefix(sshDecided, strings.SplitN(decided, "://", 2)[0]+"://") {
			t.Errorf("scheme %q: ssh surface decided %q while the others decided %q", scheme, sshDecided, decided)
		}
	}
}

// A-9 · A relay that could not be told which scheme it serves must not pick
// one. Announcing a scheme nothing honours is the defect; coercing an unknown
// value would restore it under a new spelling.
func TestRelayRefusesAnUnknownScheme(t *testing.T) {
	for _, bad := range []string{"ftp", "htp", "ws", "https://x", "javascript"} {
		r, err := relay.New(relay.Config{BaseDomain: "pumasi.link", PublicScheme: bad})
		if !errors.Is(err, core.ErrUnknownScheme) {
			t.Errorf("relay.New(scheme %q) error = %v, want core.ErrUnknownScheme", bad, err)
		}
		if r != nil {
			t.Errorf("relay.New(scheme %q) returned a relay alongside its error", bad)
		}
	}
}

// A-10 · Configuring the scheme changes the scheme and nothing else. The host
// keeps its shape, the label is still lowercased, and a raw TCP address —
// which is a host:port and has no scheme — is untouched.
func TestSchemeChangesNothingButTheScheme(t *testing.T) {
	for _, scheme := range []string{core.SchemeHTTP, core.SchemeHTTPS} {
		h := newSchemeHarness(t, scheme)
		got := h.authURL
		host := strings.TrimPrefix(strings.TrimPrefix(got, "https://"), "http://")
		if host != h.subdomain+".pumasi.link" {
			t.Errorf("scheme %q leaked into the host: %q", scheme, got)
		}
		if strings.ToLower(host) != host {
			t.Errorf("scheme %q: host is not lowercased: %q", scheme, host)
		}
	}

	// A raw TCP tunnel's address is a host:port. No scheme belongs on it, at
	// either setting.
	for _, scheme := range []string{core.SchemeHTTP, core.SchemeHTTPS} {
		r, err := relay.New(relay.Config{
			BaseDomain:   "pumasi.link",
			PublicScheme: scheme,
			TCPPortLow:   34500,
			TCPPortHigh:  34599,
			TCPBindHost:  "127.0.0.1",
			PublicHost:   "127.0.0.1",
		})
		if err != nil {
			t.Fatalf("relay.New: %v", err)
		}
		if r.Registry().PublicScheme() != scheme {
			t.Errorf("registry scheme = %q, want %q", r.Registry().PublicScheme(), scheme)
		}
	}
}

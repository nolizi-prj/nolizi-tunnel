package relay_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pumasi-ai/pumasi-tunnel/relay"
)

// The apex serves the console; a tunnel hostname must never be shadowed by it.
func TestApexServesConsoleAndTunnelsStillRoute(t *testing.T) {
	h := newHarness(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("from the tunnelled service"))
	}))

	req, _ := http.NewRequest(http.MethodGet, h.edge.URL+"/", nil)
	req.Host = "pumasi.link" // the apex
	resp, err := h.edge.Client().Do(req)
	if err != nil {
		t.Fatalf("apex request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("apex status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("apex content-type = %q, want html", ct)
	}

	// The tunnel must still be reachable — the console must not have taken
	// over the wildcard.
	_, body := h.get(t, "/")
	if !strings.Contains(body, "from the tunnelled service") {
		t.Errorf("tunnel routing broke after adding the console: %q", body)
	}
}

func TestPublicStatusDoesNotListOpenTunnels(t *testing.T) {
	h := newHarness(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req, _ := http.NewRequest(http.MethodGet, h.edge.URL+"/_pumasi/status", nil)
	req.Host = "pumasi.link"
	resp, err := h.edge.Client().Do(req)
	if err != nil {
		t.Fatalf("status request: %v", err)
	}
	defer resp.Body.Close()

	var got map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decoding status: %v", err)
	}
	if got["base_domain"] != "pumasi.link" {
		t.Fatalf("status = %+v, want relay configuration", got)
	}
	for _, private := range []string{"count", "tunnels", "subdomain", "local_port", "tcp_addr", "opened_at"} {
		if _, exists := got[private]; exists {
			t.Errorf("public status exposes %q: %+v", private, got)
		}
	}
}

// The status feed is public; it must not leak the agent session id, which
// identifies a connection rather than a published resource.
func TestStatusDoesNotLeakAgentID(t *testing.T) {
	h := newHarness(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req, _ := http.NewRequest(http.MethodGet, h.edge.URL+"/_pumasi/status", nil)
	req.Host = "pumasi.link"
	resp, err := h.edge.Client().Do(req)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	for _, leak := range []string{"agent_id", "agentID", "AgentID"} {
		if strings.Contains(string(body), leak) {
			t.Errorf("status body exposes %q: %s", leak, body)
		}
	}
}

func TestApexUnknownPathIsNotFound(t *testing.T) {
	r, err := relay.New(relay.Config{BaseDomain: "pumasi.link"})
	if err != nil {
		t.Fatalf("relay.New: %v", err)
	}
	edge := httptest.NewServer(r)
	defer edge.Close()

	req, _ := http.NewRequest(http.MethodGet, edge.URL+"/nope", nil)
	req.Host = "pumasi.link"
	resp, err := edge.Client().Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestOperationalEndpointsReportOneVersion(t *testing.T) {
	r, err := relay.New(relay.Config{BaseDomain: "pumasi.link"})
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	for _, path := range []string{"/version", "/healthz", "/readyz"} {
		req := httptest.NewRequest(http.MethodGet, "http://pumasi.link"+path, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d", path, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `"version":"`+relay.Version+`"`) {
			t.Errorf("%s does not report %s: %s", path, relay.Version, rec.Body.String())
		}
	}
}

func TestConsoleOffersZeroInstallSSHAndFeedback(t *testing.T) {
	r, err := relay.New(relay.Config{
		BaseDomain:      "pumasi.link",
		AgentPublicPort: ":7001",
		SSHPublicPort:   ":2222",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	req := httptest.NewRequest(http.MethodGet, "https://pumasi.link/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	body := rec.Body.String()
	for _, want := range []string{"Stock SSH", "Pumasi client", ">Feedback</button>", "unique name assigned", "name=\"method\"", "Follow three small steps", "Install the client if this is your first time", "Already installed? Skip this step", "copy-install-step", "press Ctrl+C to close the tunnel", "applyServicePreset", "serviceEl.addEventListener(\"input\"", "-T", "LogLevel=ERROR", "ExitOnForwardFailure=yes", "-R 0:127.0.0.1:", "local HTTP; visitors receive HTTPS", "different HTTPS address", "Common service", "Windows Remote Desktop · 3389", "Custom port…", "Auto-assigned", "Ctrl/⌘ ↵", "feedback-image-download", "curl -fsSL https://pumasi.link/install.sh | sh", "modern-screenshot.js"} {
		if !strings.Contains(body, want) {
			t.Errorf("console missing %q", want)
		}
	}
	if strings.Contains(body, "Install the Pumasi client <span") {
		t.Error("console still contains the redundant standalone install card")
	}
	if strings.Contains(body, "Open tunnels") {
		t.Error("console exposes the public tunnel directory")
	}

	statusReq := httptest.NewRequest(http.MethodGet, "https://pumasi.link/_pumasi/status", nil)
	statusRec := httptest.NewRecorder()
	r.ServeHTTP(statusRec, statusReq)
	if !strings.Contains(statusRec.Body.String(), `"ssh_port":":2222"`) {
		t.Errorf("status missing SSH port: %s", statusRec.Body.String())
	}
}

func TestInstallerAndScreenshotLibraryAreServedFromApex(t *testing.T) {
	r, err := relay.New(relay.Config{BaseDomain: "pumasi.link"})
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	checks := []struct {
		path, contentType, body string
	}{
		{"/install.sh", "text/x-shellscript", "checksums.txt"},
		{"/_pumasi/modern-screenshot.js", "text/javascript", "modernScreenshot"},
	}
	for _, check := range checks {
		req := httptest.NewRequest(http.MethodGet, "http://pumasi.link"+check.path, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK || !strings.HasPrefix(rec.Header().Get("Content-Type"), check.contentType) || !strings.Contains(rec.Body.String(), check.body) {
			t.Errorf("%s status=%d type=%q", check.path, rec.Code, rec.Header().Get("Content-Type"))
		}
	}
}

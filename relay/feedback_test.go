package relay_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pumasi-ai/pumasi-tunnel/relay"
)

func TestFeedbackCreatesServerSideIssueWithoutExposingCredential(t *testing.T) {
	var gotAuth string
	var got map[string]any
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		gotAuth = req.Header.Get("Authorization")
		if req.URL.Path != "/repos/pumasi-ai/pumasi-tunnel/issues" {
			t.Errorf("path = %s", req.URL.Path)
		}
		if err := json.NewDecoder(req.Body).Decode(&got); err != nil {
			t.Error(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"html_url":"https://github.com/pumasi-ai/pumasi-tunnel/issues/99"}`))
	}))
	defer api.Close()

	r, err := relay.New(relay.Config{
		BaseDomain: "pumasi.link", PublicScheme: "https",
		FeedbackGitHubToken: "server-secret", FeedbackAPIURL: api.URL,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	req := httptest.NewRequest(http.MethodPost, "https://pumasi.link/_pumasi/feedback",
		strings.NewReader(`{"type":"bug","description":"Copy button does not work","contact":"person@example.com","page":"https://pumasi.link/?secret=gone"}`))
	req.Header.Set("Origin", "https://pumasi.link")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if gotAuth != "Bearer server-secret" {
		t.Errorf("authorization = %q", gotAuth)
	}
	if strings.Contains(rec.Body.String(), "server-secret") {
		t.Fatal("credential leaked to browser")
	}
	body, _ := got["body"].(string)
	if strings.Contains(body, "?secret") || !strings.Contains(body, "person@example.com") {
		t.Errorf("issue body was not sanitized: %s", body)
	}
}

func TestFeedbackRequiresHTTPSForContactAndRateLimits(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusCreated) }))
	defer api.Close()
	r, err := relay.New(relay.Config{BaseDomain: "pumasi.link", FeedbackGitHubToken: "token", FeedbackAPIURL: api.URL})
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	req := httptest.NewRequest(http.MethodPost, "http://pumasi.link/_pumasi/feedback",
		strings.NewReader(`{"description":"hello tunnel","contact":"person@example.com"}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("contact over HTTP status=%d", rec.Code)
	}

	for i := 0; i < 5; i++ {
		req = httptest.NewRequest(http.MethodPost, "http://pumasi.link/_pumasi/feedback", strings.NewReader(`{"description":"hello tunnel"}`))
		req.RemoteAddr = "192.0.2.5:1234"
		rec = httptest.NewRecorder()
		r.ServeHTTP(rec, req)
	}
	req = httptest.NewRequest(http.MethodPost, "http://pumasi.link/_pumasi/feedback", strings.NewReader(`{"description":"hello tunnel"}`))
	req.RemoteAddr = "192.0.2.5:1234"
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("rate limit status=%d", rec.Code)
	}
}

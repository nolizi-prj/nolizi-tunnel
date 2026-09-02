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

func TestFeedbackUploadsValidatedImagesAndIncludesTransparentDiagnostics(t *testing.T) {
	var issueBody string
	var uploads int
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPut && strings.Contains(req.URL.Path, "/contents/.github/feedback-attachments/") {
			uploads++
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"content":{"download_url":"https://example.test/image.png"}}`))
			return
		}
		var payload struct {
			Body string `json:"body"`
		}
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		issueBody = payload.Body
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"html_url":"https://github.com/pumasi-ai/pumasi-tunnel/issues/100"}`))
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
	// Valid 1x1 PNG.
	png := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII="
	body := `{"type":"question","description":"How do assigned names work?","page":"https://pumasi.link/?token=secret","viewport":"390x844","timezone":"America/Chicago","user_agent":"Test Browser","online":true,"errors":[{"message":"Example error","source":"https://pumasi.link/app.js?token=secret","line":4,"column":2,"timestamp":"2026-09-02T00:00:00Z"}],"images":[{"name":"screen.png","data":"` + png + `"}]}`
	req := httptest.NewRequest(http.MethodPost, "https://pumasi.link/_pumasi/feedback", strings.NewReader(body))
	req.Header.Set("Origin", "https://pumasi.link")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if uploads != 1 {
		t.Fatalf("uploads=%d", uploads)
	}
	for _, want := range []string{"Included diagnostics", "390x844", "America/Chicago", "Example error", "Attached images (1)"} {
		if !strings.Contains(issueBody, want) {
			t.Errorf("issue body missing %q: %s", want, issueBody)
		}
	}
	if strings.Contains(issueBody, "token=secret") {
		t.Errorf("diagnostics leaked URL query: %s", issueBody)
	}
}

func TestFeedbackRejectsTooManyImages(t *testing.T) {
	r, err := relay.New(relay.Config{BaseDomain: "pumasi.link", PublicScheme: "https", FeedbackGitHubToken: "token"})
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	png := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII="
	images := strings.Repeat(`{"name":"x.png","data":"`+png+`"},`, 6)
	body := `{"description":"Too many images","images":[` + strings.TrimSuffix(images, ",") + `]}`
	req := httptest.NewRequest(http.MethodPost, "https://pumasi.link/_pumasi/feedback", strings.NewReader(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "up to 5") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

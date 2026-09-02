package relay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxFeedbackBody = 32 << 10

type feedbackRequest struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Contact     string `json:"contact"`
	Page        string `json:"page"`
}

func (r *Relay) serveFeedback(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "POST required"})
		return
	}
	if r.cfg.FeedbackGitHubToken == "" {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "Feedback is temporarily unavailable"})
		return
	}
	if origin := req.Header.Get("Origin"); origin != "" {
		want := r.cfg.PublicScheme + "://" + r.cfg.BaseDomain
		if origin != want {
			writeJSON(w, http.StatusForbidden, map[string]any{"error": "Origin not allowed"})
			return
		}
	}
	if !r.allowFeedback(req.RemoteAddr, time.Now()) {
		w.Header().Set("Retry-After", "3600")
		writeJSON(w, http.StatusTooManyRequests, map[string]any{"error": "Too many feedback reports; try again later"})
		return
	}

	var in feedbackRequest
	dec := json.NewDecoder(io.LimitReader(req.Body, maxFeedbackBody+1))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Invalid feedback"})
		return
	}
	in.Description = strings.TrimSpace(in.Description)
	if len(in.Description) < 3 || len(in.Description) > 4000 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Description must be 3–4000 characters"})
		return
	}
	if len(in.Contact) > 254 || len(in.Page) > 2048 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Feedback fields are too long"})
		return
	}
	if in.Contact != "" && r.cfg.PublicScheme != "https" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Contact details require HTTPS"})
		return
	}
	if in.Page != "" {
		u, err := url.Parse(in.Page)
		if err != nil || u.Hostname() != r.cfg.BaseDomain || (u.Scheme != "https" && u.Scheme != "http") {
			in.Page = ""
		} else {
			u.RawQuery = ""
			u.Fragment = ""
			in.Page = u.String()
		}
	}

	kind := "Feedback"
	label := "feedback"
	switch in.Type {
	case "bug":
		kind, label = "Bug", "bug"
	case "idea":
		kind, label = "Idea", "enhancement"
	}
	titleText := strings.ReplaceAll(in.Description, "\n", " ")
	if len(titleText) > 80 {
		titleText = titleText[:80] + "…"
	}
	bodyText := fmt.Sprintf("### Feedback\n\n%s\n\n| Context | Value |\n|---|---|\n| Version | `%s` |\n| Page | `%s` |\n| Contact | `%s` |\n",
		escapeMarkdown(in.Description), Version, escapeMarkdown(in.Page), escapeMarkdown(in.Contact))
	payload, _ := json.Marshal(map[string]any{
		"title":  fmt.Sprintf("[Feedback] %s: %s", kind, titleText),
		"body":   bodyText,
		"labels": []string{"feedback", label},
	})
	apiReq, err := http.NewRequestWithContext(req.Context(), http.MethodPost,
		strings.TrimRight(r.cfg.FeedbackAPIURL, "/")+"/repos/"+r.cfg.FeedbackGitHubRepo+"/issues", bytes.NewReader(payload))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "Could not prepare feedback"})
		return
	}
	apiReq.Header.Set("Authorization", "Bearer "+r.cfg.FeedbackGitHubToken)
	apiReq.Header.Set("Accept", "application/vnd.github+json")
	apiReq.Header.Set("Content-Type", "application/json")
	apiReq.Header.Set("User-Agent", "pumasi-tunnel/"+Version)
	res, err := r.cfg.FeedbackHTTPClient.Do(apiReq)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": "Feedback delivery failed"})
		return
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		io.Copy(io.Discard, io.LimitReader(res.Body, 4096))
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": "Feedback delivery failed"})
		return
	}
	var out struct {
		HTMLURL string `json:"html_url"`
	}
	_ = json.NewDecoder(io.LimitReader(res.Body, 64<<10)).Decode(&out)
	writeJSON(w, http.StatusCreated, map[string]any{"ok": true, "issue_url": out.HTMLURL})
}

func (r *Relay) allowFeedback(remote string, now time.Time) bool {
	host, _, err := net.SplitHostPort(remote)
	if err != nil {
		host = remote
	}
	cutoff := now.Add(-time.Hour)
	r.feedbackMu.Lock()
	defer r.feedbackMu.Unlock()
	recent := r.feedbackAttempts[host][:0]
	for _, at := range r.feedbackAttempts[host] {
		if at.After(cutoff) {
			recent = append(recent, at)
		}
	}
	if len(recent) >= 5 {
		r.feedbackAttempts[host] = recent
		return false
	}
	r.feedbackAttempts[host] = append(recent, now)
	return true
}

func escapeMarkdown(value string) string {
	value = strings.ReplaceAll(value, "`", "\\`")
	value = strings.ReplaceAll(value, "|", "\\|")
	return value
}

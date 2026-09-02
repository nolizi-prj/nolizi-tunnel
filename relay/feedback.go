package relay

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

const (
	maxFeedbackBody       = 28 << 20
	maxFeedbackImages     = 5
	maxFeedbackImageBytes = 4 << 20
)

type feedbackError struct {
	Message   string `json:"message"`
	Source    string `json:"source"`
	Line      int    `json:"line"`
	Column    int    `json:"column"`
	Timestamp string `json:"timestamp"`
}

type feedbackImage struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type feedbackRequest struct {
	Type        string          `json:"type"`
	Description string          `json:"description"`
	Contact     string          `json:"contact"`
	Page        string          `json:"page"`
	Viewport    string          `json:"viewport"`
	Timezone    string          `json:"timezone"`
	UserAgent   string          `json:"user_agent"`
	Online      bool            `json:"online"`
	Errors      []feedbackError `json:"errors"`
	Images      []feedbackImage `json:"images"`
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
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Invalid feedback or attachments are too large"})
		return
	}
	if err := validateFeedback(&in, r.cfg); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	attachmentURLs := make([]string, 0, len(in.Images))
	for i, image := range in.Images {
		if uploaded, err := r.uploadFeedbackImage(req, image, i); err == nil {
			attachmentURLs = append(attachmentURLs, uploaded)
		}
	}

	kind, label := feedbackKind(in.Type)
	titleText := strings.ReplaceAll(in.Description, "\n", " ")
	if len(titleText) > 80 {
		titleText = titleText[:80] + "…"
	}
	bodyText := feedbackMarkdown(in, attachmentURLs)
	labels := []string{"feedback"}
	if label != "" {
		labels = append(labels, label)
	}
	payload, _ := json.Marshal(map[string]any{
		"title":  fmt.Sprintf("[Feedback] %s: %s", kind, titleText),
		"body":   bodyText,
		"labels": labels,
	})
	apiReq, err := r.githubRequest(req, http.MethodPost, "/repos/"+r.cfg.FeedbackGitHubRepo+"/issues", payload)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "Could not prepare feedback"})
		return
	}
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
	writeJSON(w, http.StatusCreated, map[string]any{
		"ok": true, "issue_url": out.HTMLURL,
		"attachments_uploaded": len(attachmentURLs),
	})
}

func validateFeedback(in *feedbackRequest, cfg Config) error {
	in.Description = strings.TrimSpace(in.Description)
	in.Contact = strings.TrimSpace(in.Contact)
	if len(in.Description) < 3 || len(in.Description) > 5000 {
		return fmt.Errorf("description must be 3–5000 characters")
	}
	if len(in.Contact) > 254 || len(in.Page) > 2048 || len(in.Viewport) > 32 || len(in.Timezone) > 80 || len(in.UserAgent) > 512 {
		return fmt.Errorf("feedback fields are too long")
	}
	if in.Contact != "" && cfg.PublicScheme != "https" {
		return fmt.Errorf("contact details require HTTPS")
	}
	if len(in.Errors) > 5 {
		in.Errors = in.Errors[len(in.Errors)-5:]
	}
	for i := range in.Errors {
		in.Errors[i].Message = truncate(in.Errors[i].Message, 500)
		in.Errors[i].Source = truncate(sanitizePage(in.Errors[i].Source, cfg.BaseDomain), 256)
		in.Errors[i].Timestamp = truncate(in.Errors[i].Timestamp, 40)
	}
	in.Page = sanitizePage(in.Page, cfg.BaseDomain)
	if len(in.Images) > maxFeedbackImages {
		return fmt.Errorf("attach up to %d images", maxFeedbackImages)
	}
	for _, image := range in.Images {
		if _, _, err := decodeFeedbackImage(image.Data); err != nil {
			return err
		}
	}
	return nil
}

func decodeFeedbackImage(dataURL string) (string, string, error) {
	prefixes := map[string]string{
		"data:image/png;base64,":  "png",
		"data:image/jpeg;base64,": "jpg",
		"data:image/webp;base64,": "webp",
	}
	for prefix, ext := range prefixes {
		if !strings.HasPrefix(dataURL, prefix) {
			continue
		}
		encoded := strings.TrimPrefix(dataURL, prefix)
		if base64.StdEncoding.DecodedLen(len(encoded)) > maxFeedbackImageBytes {
			return "", "", fmt.Errorf("each image must be 4 MB or smaller")
		}
		raw, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil || len(raw) == 0 || len(raw) > maxFeedbackImageBytes || !validImageMagic(ext, raw) {
			return "", "", fmt.Errorf("attachment is not a valid PNG, JPEG, or WebP image")
		}
		return encoded, ext, nil
	}
	return "", "", fmt.Errorf("attachments must be PNG, JPEG, or WebP images")
}

func validImageMagic(ext string, raw []byte) bool {
	switch ext {
	case "png":
		return len(raw) >= 8 && bytes.Equal(raw[:8], []byte("\x89PNG\r\n\x1a\n"))
	case "jpg":
		return len(raw) >= 3 && raw[0] == 0xff && raw[1] == 0xd8 && raw[2] == 0xff
	case "webp":
		return len(raw) >= 12 && string(raw[:4]) == "RIFF" && string(raw[8:12]) == "WEBP"
	}
	return false
}

func (r *Relay) uploadFeedbackImage(req *http.Request, image feedbackImage, index int) (string, error) {
	content, ext, err := decodeFeedbackImage(image.Data)
	if err != nil {
		return "", err
	}
	random := make([]byte, 4)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	name := strings.TrimSuffix(filepath.Base(image.Name), filepath.Ext(image.Name))
	name = safeFilename(name)
	if name == "" {
		name = "image"
	}
	filename := fmt.Sprintf("%s-%d-%s-%s.%s", time.Now().UTC().Format("20060102"), index+1, name, hex.EncodeToString(random), ext)
	path := ".github/feedback-attachments/" + filename
	payload, _ := json.Marshal(map[string]string{
		"message": "feedback: attach " + path,
		"content": content,
		"branch":  "main",
	})
	apiReq, err := r.githubRequest(req, http.MethodPut, "/repos/"+r.cfg.FeedbackGitHubRepo+"/contents/"+path, payload)
	if err != nil {
		return "", err
	}
	res, err := r.cfg.FeedbackHTTPClient.Do(apiReq)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		io.Copy(io.Discard, io.LimitReader(res.Body, 4096))
		return "", fmt.Errorf("GitHub attachment upload failed")
	}
	var out struct {
		Content struct {
			DownloadURL string `json:"download_url"`
		} `json:"content"`
	}
	_ = json.NewDecoder(io.LimitReader(res.Body, 64<<10)).Decode(&out)
	if out.Content.DownloadURL != "" {
		return out.Content.DownloadURL, nil
	}
	return "https://raw.githubusercontent.com/" + r.cfg.FeedbackGitHubRepo + "/main/" + path, nil
}

func (r *Relay) githubRequest(req *http.Request, method, path string, payload []byte) (*http.Request, error) {
	apiReq, err := http.NewRequestWithContext(req.Context(), method, strings.TrimRight(r.cfg.FeedbackAPIURL, "/")+path, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	apiReq.Header.Set("Authorization", "Bearer "+r.cfg.FeedbackGitHubToken)
	apiReq.Header.Set("Accept", "application/vnd.github+json")
	apiReq.Header.Set("Content-Type", "application/json")
	apiReq.Header.Set("User-Agent", "pumasi-tunnel/"+Version)
	return apiReq, nil
}

func feedbackMarkdown(in feedbackRequest, attachmentURLs []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "### Feedback\n\n%s\n\n### Submitter\n\n- **Type:** %s\n- **Contact:** %s\n- **Submitted:** `%s`\n\n", escapeMarkdown(in.Description), feedbackKindName(in.Type), markdownValue(in.Contact, "Anonymous"), time.Now().UTC().Format(time.RFC3339))
	b.WriteString("### Included diagnostics\n\n| Context | Value |\n|---|---|\n")
	fmt.Fprintf(&b, "| Product version | `%s` |\n| Reported from | `%s` |\n| Viewport | `%s` |\n| Timezone | `%s` |\n| Browser | `%s` |\n| Network online | `%t` |\n", Version, escapeMarkdown(in.Page), escapeMarkdown(in.Viewport), escapeMarkdown(in.Timezone), escapeMarkdown(in.UserAgent), in.Online)
	if len(in.Errors) > 0 {
		fmt.Fprintf(&b, "\n<details><summary>Recent browser errors (%d)</summary>\n\n", len(in.Errors))
		for _, item := range in.Errors {
			fmt.Fprintf(&b, "- `%s` %s (`%s:%d:%d`)\n", escapeMarkdown(item.Timestamp), escapeMarkdown(item.Message), escapeMarkdown(item.Source), item.Line, item.Column)
		}
		b.WriteString("\n</details>\n")
	}
	if len(attachmentURLs) > 0 {
		fmt.Fprintf(&b, "\n### Attached images (%d)\n\n", len(attachmentURLs))
		for i, imageURL := range attachmentURLs {
			fmt.Fprintf(&b, "<details%s><summary>Image %d</summary>\n\n![Feedback image %d](%s)\n\n</details>\n\n", map[bool]string{true: " open"}[i == 0], i+1, i+1, imageURL)
		}
	}
	return b.String()
}

func feedbackKind(kind string) (string, string) {
	switch kind {
	case "bug":
		return "Bug", "bug"
	case "idea":
		return "Idea", "enhancement"
	case "question":
		return "Question", ""
	default:
		return "Feedback", ""
	}
}

func feedbackKindName(kind string) string {
	name, _ := feedbackKind(kind)
	return name
}

func sanitizePage(raw, baseDomain string) string {
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") {
		return ""
	}
	host := strings.ToLower(u.Hostname())
	if host != baseDomain && !strings.HasSuffix(host, "."+baseDomain) {
		return ""
	}
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

func safeFilename(value string) string {
	value = strings.ToLower(value)
	var b strings.Builder
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	return truncate(b.String(), 40)
}

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max]
}

func markdownValue(value, empty string) string {
	if value == "" {
		return "_" + empty + "_"
	}
	return "`" + escapeMarkdown(value) + "`"
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

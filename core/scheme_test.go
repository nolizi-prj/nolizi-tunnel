package core

import (
	"errors"
	"strings"
	"testing"
)

// Acceptance cases for spec/0001-public-scheme, A-1 to A-4. Frozen at the
// spec review; see spec/0001-public-scheme/acceptance/CASES.md for what
// execution makes each of them fail.

// A-1 · A relay told nothing about TLS announces http://, because http is
// what it serves with nothing in front of it. This is the case the live
// relay failed on 2026-08-31: it announced https://sshsteward.pumasi.link
// while port 443 refused every connection.
func TestPublicURLDefaultsToHTTP(t *testing.T) {
	// Both spellings of "not configured": the explicit default, and the
	// empty string an operator gets by not passing the flag at all.
	for _, scheme := range []string{SchemeHTTP, ""} {
		r := NewRegistry("pumasi.link", scheme)
		got := r.PublicURL("myapi")
		if got != "http://myapi.pumasi.link" {
			t.Errorf("PublicURL with scheme %q = %q, want http://myapi.pumasi.link", scheme, got)
		}
		if strings.HasPrefix(got, "https://") {
			t.Errorf("scheme %q announced https, which nothing in this process can honour", scheme)
		}
		if r.PublicScheme() != SchemeHTTP {
			t.Errorf("PublicScheme() = %q, want %q", r.PublicScheme(), SchemeHTTP)
		}
	}
}

// A-2 · An operator who did put a terminator in front says so once and gets
// https:// back.
func TestPublicURLHonoursConfiguredScheme(t *testing.T) {
	r := NewRegistry("pumasi.link", SchemeHTTPS)
	if got := r.PublicURL("myapi"); got != "https://myapi.pumasi.link" {
		t.Errorf("PublicURL = %q, want https://myapi.pumasi.link", got)
	}
	if r.PublicScheme() != SchemeHTTPS {
		t.Errorf("PublicScheme() = %q, want %q", r.PublicScheme(), SchemeHTTPS)
	}
	// The label is still lowercased and the base domain still normalised;
	// configuring a scheme must not have moved anything else.
	if got := r.PublicURL("MyApi"); got != "https://myapi.pumasi.link" {
		t.Errorf("PublicURL(MyApi) = %q, want the lowercased host", got)
	}
}

// A-3 · The legal set is written down once, spelling is forgiven, and
// anything else is refused rather than coerced into a scheme the relay
// cannot honour.
func TestParsePublicScheme(t *testing.T) {
	accepted := map[string]string{
		"":           SchemeHTTP,
		"http":       SchemeHTTP,
		"HTTP":       SchemeHTTP,
		"  http  ":   SchemeHTTP,
		"http://":    SchemeHTTP,
		"https":      SchemeHTTPS,
		"HTTPS":      SchemeHTTPS,
		" https ":    SchemeHTTPS,
		"https://":   SchemeHTTPS,
		" HTTPS:// ": SchemeHTTPS,
	}
	for in, want := range accepted {
		got, err := ParsePublicScheme(in)
		if err != nil {
			t.Errorf("ParsePublicScheme(%q) errored: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("ParsePublicScheme(%q) = %q, want %q", in, got, want)
		}
	}

	refused := []string{"ftp", "htp", "ws", "wss", "gopher", "http s", "https://x", "//https", "HTTPs://x"}
	for _, in := range refused {
		got, err := ParsePublicScheme(in)
		if !errors.Is(err, ErrUnknownScheme) {
			t.Errorf("ParsePublicScheme(%q) = %q, %v; want ErrUnknownScheme", in, got, err)
		}
		if got != "" {
			t.Errorf("ParsePublicScheme(%q) returned %q alongside its error", in, got)
		}
	}
}

// A-4 · A registry handed an unvalidated, illegal scheme falls closed to
// http. It never interpolates the string it was given — an address built from
// a scheme no client speaks is worse than the defect this replaces.
func TestNewRegistryFailsClosedToHTTP(t *testing.T) {
	for _, bad := range []string{"ftp", "javascript", "gopher", "https://x"} {
		r := NewRegistry("pumasi.link", bad)
		if got := r.PublicURL("myapi"); got != "http://myapi.pumasi.link" {
			t.Errorf("NewRegistry(%q).PublicURL = %q, want the http fallback", bad, got)
		}
		if r.PublicScheme() != SchemeHTTP {
			t.Errorf("NewRegistry(%q).PublicScheme() = %q, want %q", bad, r.PublicScheme(), SchemeHTTP)
		}
	}
}

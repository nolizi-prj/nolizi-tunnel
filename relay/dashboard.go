package relay

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
)

// The console the relay serves at its own apex.
//
// Shape follows the incumbent study (docs/ux/incumbent-ux-spec.md): for a
// tunnel service the dashboard's centre of gravity is a **command builder**,
// not a resource manager — a form whose only output is the one line to paste
// — with a live table beside it and no create action, because a tunnel comes
// into being by running the command, never by filling in a web form (§3.2,
// §5.1). Patterns are taken from that spec; nothing of the incumbents'
// expression is.
//
// It is one embedded file with no build step and no external fetches, so the
// relay stays a single static binary that can be dropped on a host and run.

//go:embed dashboard.html
var dashboardHTML []byte

//go:embed modern-screenshot.js
var modernScreenshotJS []byte

type statusView struct {
	Version    string `json:"version"`
	BaseDomain string `json:"base_domain"`
	AgentPort  string `json:"agent_port"`
	SSHPort    string `json:"ssh_port,omitempty"`
	TCPRange   string `json:"tcp_range,omitempty"`
}

// serveDashboard renders the console page.
func (r *Relay) serveDashboard(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Write(dashboardHTML)
}

func serveModernScreenshot(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(modernScreenshotJS)
}

// serveStatus gives the public console only relay configuration. Tunnel
// addresses, local ports, ownership, and connection times are private even
// when an individual address is intentionally shared with a recipient.
func (r *Relay) serveStatus(w http.ResponseWriter, _ *http.Request) {
	status := statusView{
		Version:    Version,
		BaseDomain: r.cfg.BaseDomain,
		AgentPort:  r.cfg.AgentPublicPort,
		SSHPort:    r.cfg.SSHPublicPort,
	}
	if r.pool != nil {
		status.TCPRange = fmt.Sprintf("%d–%d", r.cfg.TCPPortLow, r.cfg.TCPPortHigh)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(status)
}

package relay

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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

// tunnelView is one row of the live table, shaped for the page rather than
// for the routing table.
type tunnelView struct {
	Subdomain string `json:"subdomain"`
	URL       string `json:"url"`
	TCPAddr   string `json:"tcp_addr,omitempty"`
	LocalPort int    `json:"local_port"`
	// Fixed means the address was asked for by name or by number rather than
	// handed out. It says what the agent requested and nothing about who owns
	// it — an anonymous agent can ask for a free name and get it.
	Fixed bool `json:"fixed"`
	// Reserved means the name belongs to a token, so it is held for its owner
	// through a disconnect instead of being free the instant the agent drops.
	// This is the field roadmap/BACKLOG.md item 2 found written and read
	// nowhere; it is read here (spec/0004 §5.3).
	Reserved bool   `json:"reserved"`
	OpenedAt string `json:"opened_at"`
	AgeSecs  int64  `json:"age_secs"`
}

type statusView struct {
	BaseDomain string       `json:"base_domain"`
	AgentPort  string       `json:"agent_port"`
	TCPRange   string       `json:"tcp_range,omitempty"`
	Count      int          `json:"count"`
	Tunnels    []tunnelView `json:"tunnels"`
	ServerTime string       `json:"server_time"`
}

// serveDashboard renders the console page.
func (r *Relay) serveDashboard(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// The page is a live view of who is connected; a cached copy would show a
	// tunnel that has since closed.
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Write(dashboardHTML)
}

// serveStatus is the JSON the page polls. It deliberately exposes only what
// is already public — a tunnel's hostname is served to any visitor anyway —
// and never the agent id, which identifies a session rather than a resource.
func (r *Relay) serveStatus(w http.ResponseWriter, _ *http.Request) {
	now := time.Now().UTC()
	tunnels := r.registry.List()

	views := make([]tunnelView, 0, len(tunnels))
	for _, t := range tunnels {
		v := tunnelView{
			Subdomain: t.Subdomain,
			URL:       r.registry.PublicURL(t.Subdomain),
			LocalPort: t.LocalPort,
			Fixed:     t.Requested,
			Reserved:  t.Reserved,
			OpenedAt:  t.OpenedAt.Format(time.RFC3339),
			AgeSecs:   int64(now.Sub(t.OpenedAt).Seconds()),
		}
		if t.TCPPort != 0 {
			v.TCPAddr = fmt.Sprintf("%s:%d", r.cfg.PublicHost, t.TCPPort)
		}
		views = append(views, v)
	}

	status := statusView{
		BaseDomain: r.cfg.BaseDomain,
		AgentPort:  r.cfg.AgentPublicPort,
		Count:      len(views),
		Tunnels:    views,
		ServerTime: now.Format(time.RFC3339),
	}
	if r.pool != nil {
		status.TCPRange = fmt.Sprintf("%d–%d", r.cfg.TCPPortLow, r.cfg.TCPPortHigh)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(status)
}

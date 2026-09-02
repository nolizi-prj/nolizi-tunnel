package relay

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// tunnelAbuseGuard bounds anonymous tunnel creation by source address. It is
// deliberately in-memory: these are circuit breakers for one relay, not an
// identity or billing system.
type tunnelAbuseGuard struct {
	mu        sync.Mutex
	active    map[string]int
	attempts  map[string][]time.Time
	maxActive int
	maxStarts int
	window    time.Duration
}

func newTunnelAbuseGuard(maxActive, maxStarts int, window time.Duration) *tunnelAbuseGuard {
	return &tunnelAbuseGuard{
		active: make(map[string]int), attempts: make(map[string][]time.Time),
		maxActive: maxActive, maxStarts: maxStarts, window: window,
	}
}

func sourceHost(remote string) string {
	host, _, err := net.SplitHostPort(remote)
	if err == nil && host != "" {
		return host
	}
	return remote
}

func (g *tunnelAbuseGuard) acquire(remote string, now time.Time) (func(), error) {
	key := sourceHost(remote)
	g.mu.Lock()
	defer g.mu.Unlock()

	cutoff := now.Add(-g.window)
	recent := g.attempts[key][:0]
	for _, at := range g.attempts[key] {
		if at.After(cutoff) {
			recent = append(recent, at)
		}
	}
	if len(recent) >= g.maxStarts {
		g.attempts[key] = recent
		return nil, fmt.Errorf("tunnel creation rate limit reached for %s", key)
	}
	recent = append(recent, now)
	g.attempts[key] = recent
	if g.active[key] >= g.maxActive {
		return nil, fmt.Errorf("active tunnel limit reached for %s", key)
	}
	g.active[key]++

	var once sync.Once
	return func() {
		once.Do(func() {
			g.mu.Lock()
			defer g.mu.Unlock()
			g.active[key]--
			if g.active[key] == 0 {
				delete(g.active, key)
			}
		})
	}, nil
}

func (r *Relay) beginTunnel(conn net.Conn) (func(), bool) {
	release, err := r.abuse.acquire(conn.RemoteAddr().String(), time.Now())
	if err != nil {
		r.log.Warn("tunnel connection rate-limited", "peer", conn.RemoteAddr(), "error", err)
		conn.Close()
		return nil, false
	}
	return release, true
}

func (r *Relay) acquireVisitor(agentID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.visitorCounts[agentID] >= r.cfg.MaxConnectionsPerTunnel {
		return false
	}
	r.visitorCounts[agentID]++
	return true
}

func (r *Relay) releaseVisitor(agentID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.visitorCounts[agentID]--
	if r.visitorCounts[agentID] <= 0 {
		delete(r.visitorCounts, agentID)
	}
}

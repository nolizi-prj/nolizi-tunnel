package core

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Routing errors.
var (
	ErrNoTunnel      = errors.New("core: no tunnel registered for host")
	ErrForeignHost   = errors.New("core: host is not under the tunnel base domain")
	ErrNameTaken     = errors.New("core: subdomain already registered")
	ErrEmptyHost     = errors.New("core: empty host")
	ErrNestedSubname = errors.New("core: nested subdomains are not routable")
)

// SplitHost reduces a request's Host header to the tunnel label beneath
// baseDomain. It lowercases, drops any port, tolerates a trailing root dot,
// and refuses hosts that do not sit exactly one label under the base — so
// "a.b.pumasi.link" is rejected rather than silently routed to "a.b", which
// would let one tenant's wildcard certificate cover another's name.
//
// SplitHost("MyApi.pumasi.link:8080", "pumasi.link") == "myapi".
func SplitHost(host, baseDomain string) (string, error) {
	h := strings.ToLower(strings.TrimSpace(host))
	if h == "" {
		return "", ErrEmptyHost
	}
	// Strip the port. IPv6 literals are bracketed and can never be a tunnel
	// host, so a plain last-colon split is safe here.
	if i := strings.LastIndex(h, ":"); i != -1 && !strings.Contains(h[i:], "]") {
		h = h[:i]
	}
	h = strings.TrimSuffix(h, ".")

	base := strings.ToLower(strings.TrimSpace(baseDomain))
	base = strings.TrimSuffix(base, ".")
	if h == base {
		// The apex is the marketing site, never a tunnel.
		return "", fmt.Errorf("%w: %q is the base domain itself", ErrForeignHost, host)
	}

	suffix := "." + base
	if !strings.HasSuffix(h, suffix) {
		return "", fmt.Errorf("%w: %q is not under %q", ErrForeignHost, host, baseDomain)
	}

	label := strings.TrimSuffix(h, suffix)
	if label == "" {
		return "", fmt.Errorf("%w: %q", ErrEmptyHost, host)
	}
	if strings.Contains(label, ".") {
		return "", fmt.Errorf("%w: %q", ErrNestedSubname, host)
	}
	return label, nil
}

// IsApexHost reports whether a Host header addresses the base domain itself
// rather than a tunnel beneath it. The relay serves its own console there.
func IsApexHost(host, baseDomain string) bool {
	h := strings.ToLower(strings.TrimSpace(host))
	if i := strings.LastIndex(h, ":"); i != -1 && !strings.Contains(h[i:], "]") {
		h = h[:i]
	}
	h = strings.TrimSuffix(h, ".")
	base := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(baseDomain)), ".")
	return h != "" && h == base
}

// Tunnel is one registered forwarding target. It is a value describing what
// the relay should do with matching traffic; it holds no connection itself,
// which is what keeps this package free of I/O.
type Tunnel struct {
	// Subdomain is the label under the base domain, e.g. "myapi".
	Subdomain string
	// AgentID identifies the connected client session that owns this tunnel.
	AgentID string
	// LocalPort is the port on the agent's machine that traffic is forwarded
	// to. Recorded for display in the dashboard and the CLI.
	LocalPort int
	// TCPPort, when non-zero, is the public TCP port allocated to this tunnel
	// for raw (non-HTTP) forwarding.
	TCPPort int
	// Reserved marks a tunnel whose subdomain belongs to an account and
	// survives disconnects, as opposed to an ephemeral one.
	Reserved bool
	// Requested marks an address the agent asked for by name or port rather
	// than accepted from the relay — the one that will still be there after a
	// reconnect, and so the one worth writing down.
	Requested bool
	// OpenedAt is when the tunnel registered. The registry stamps it, so a
	// caller cannot report a tunnel as older than it is.
	OpenedAt time.Time
}

// Registry maps hostnames and public TCP ports to tunnels. It is safe for
// concurrent use: the relay registers from its control loop while serving
// requests on many goroutines.
type Registry struct {
	baseDomain string

	mu      sync.RWMutex
	byName  map[string]Tunnel
	byTCP   map[int]string // public TCP port -> subdomain
	byAgent map[string][]string
}

// NewRegistry returns an empty registry rooted at baseDomain (e.g.
// "pumasi.link").
func NewRegistry(baseDomain string) *Registry {
	return &Registry{
		baseDomain: strings.ToLower(strings.TrimSuffix(strings.TrimSpace(baseDomain), ".")),
		byName:     make(map[string]Tunnel),
		byTCP:      make(map[int]string),
		byAgent:    make(map[string][]string),
	}
}

// BaseDomain reports the domain tunnels are published under.
func (r *Registry) BaseDomain() string { return r.baseDomain }

// Register adds a tunnel. The subdomain must already be valid — which means
// lowercase, since ValidateSubdomain refuses any other case — and free. A
// caller wanting a generated name calls AllocateSubdomain first; a caller
// accepting human input normalises it before this point, so that a request
// for "MyApi" is answered at the edge rather than silently rewritten here.
//
// Lookup and Unregister are case-insensitive by contrast: their input arrives
// from the network, where a Host header's case is not the caller's choice.
func (r *Registry) Register(t Tunnel) error {
	if err := ValidateSubdomain(t.Subdomain); err != nil {
		return err
	}
	name := t.Subdomain
	t.OpenedAt = time.Now().UTC()

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byName[name]; exists {
		return fmt.Errorf("%w: %q", ErrNameTaken, name)
	}
	if t.TCPPort != 0 {
		if owner, exists := r.byTCP[t.TCPPort]; exists {
			return fmt.Errorf("core: tcp port %d already serves %q", t.TCPPort, owner)
		}
		r.byTCP[t.TCPPort] = name
	}
	r.byName[name] = t
	if t.AgentID != "" {
		r.byAgent[t.AgentID] = append(r.byAgent[t.AgentID], name)
	}
	return nil
}

// Lookup resolves a Host header to its tunnel.
func (r *Registry) Lookup(host string) (Tunnel, error) {
	name, err := SplitHost(host, r.baseDomain)
	if err != nil {
		return Tunnel{}, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.byName[name]
	if !ok {
		return Tunnel{}, fmt.Errorf("%w: %q", ErrNoTunnel, host)
	}
	return t, nil
}

// LookupTCP resolves a public TCP port to its tunnel.
func (r *Registry) LookupTCP(port int) (Tunnel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	name, ok := r.byTCP[port]
	if !ok {
		return Tunnel{}, fmt.Errorf("%w: tcp port %d", ErrNoTunnel, port)
	}
	return r.byName[name], nil
}

// Unregister removes a tunnel by subdomain, releasing its TCP port. It
// reports whether anything was removed, so a double disconnect is not an
// error.
func (r *Registry) Unregister(subdomain string) bool {
	name := strings.ToLower(strings.TrimSpace(subdomain))
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.byName[name]
	if !ok {
		return false
	}
	delete(r.byName, name)
	if t.TCPPort != 0 {
		delete(r.byTCP, t.TCPPort)
	}
	if t.AgentID != "" {
		r.byAgent[t.AgentID] = removeString(r.byAgent[t.AgentID], name)
		if len(r.byAgent[t.AgentID]) == 0 {
			delete(r.byAgent, t.AgentID)
		}
	}
	return true
}

// UnregisterAgent drops every tunnel owned by an agent, which is what the
// relay does when a client connection dies. It returns the subdomains freed.
func (r *Registry) UnregisterAgent(agentID string) []string {
	r.mu.Lock()
	names := append([]string(nil), r.byAgent[agentID]...)
	r.mu.Unlock()

	freed := make([]string, 0, len(names))
	for _, name := range names {
		if r.Unregister(name) {
			freed = append(freed, name)
		}
	}
	return freed
}

// List returns every registered tunnel, sorted by subdomain so a dashboard
// renders in a stable order rather than map-iteration order.
func (r *Registry) List() []Tunnel {
	r.mu.RLock()
	out := make([]Tunnel, 0, len(r.byName))
	for _, t := range r.byName {
		out = append(out, t)
	}
	r.mu.RUnlock()

	sort.Slice(out, func(i, j int) bool { return out[i].Subdomain < out[j].Subdomain })
	return out
}

// Len reports how many tunnels are registered.
func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.byName)
}

// Has reports whether a subdomain is taken.
func (r *Registry) Has(subdomain string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.byName[strings.ToLower(strings.TrimSpace(subdomain))]
	return ok
}

// PublicURL is the address a caller reaches a tunnel on.
func (r *Registry) PublicURL(subdomain string) string {
	return "https://" + strings.ToLower(subdomain) + "." + r.baseDomain
}

func removeString(haystack []string, needle string) []string {
	out := haystack[:0]
	for _, s := range haystack {
		if s != needle {
			out = append(out, s)
		}
	}
	return out
}

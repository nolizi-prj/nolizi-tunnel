// Package relay is the public side of a tunnel: it accepts long-lived agent
// connections, and turns each visitor request into a stream travelling back
// down the agent's connection.
//
// The relay never dials the agent. Every byte moves over the connection the
// agent opened outbound, which is the whole point — it is what makes this
// work from behind a NAT with no port forwarding.
package relay

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"

	"github.com/pumasi-ai/pumasi-tunnel/core"
	"github.com/pumasi-ai/pumasi-tunnel/mux"
)

// Authenticator decides whether a connecting agent may have what it asked
// for. Returning an error refuses the tunnel with that reason. A relay
// running an open commons instance can accept everything; one serving
// reserved names checks the token here.
type Authenticator interface {
	Authorize(req core.AuthRequest) error
}

// AllowAll accepts every agent, including anonymous ones. It is the default
// so a relay works out of the box, matching the product's promise of a
// working tunnel from one command with no account.
type AllowAll struct{}

func (AllowAll) Authorize(core.AuthRequest) error { return nil }

// Config configures a relay.
type Config struct {
	// BaseDomain is what tunnels are published under, e.g. "pumasi.link".
	BaseDomain string
	// Auth decides who may connect; AllowAll when nil.
	Auth Authenticator
	// Logger receives connection and routing events; discarded when nil.
	Logger *slog.Logger
	// HandshakeTimeout bounds how long a new connection may take to
	// authenticate, so an idle or hostile dialler cannot hold a slot open
	// forever. Defaults to 10s.
	HandshakeTimeout time.Duration
}

// Relay accepts agents and serves visitor traffic for them.
type Relay struct {
	cfg      Config
	registry *core.Registry
	log      *slog.Logger

	mu       sync.RWMutex
	sessions map[string]*mux.Session // agent id -> session
}

// New builds a relay.
func New(cfg Config) (*Relay, error) {
	if cfg.BaseDomain == "" {
		return nil, fmt.Errorf("relay: BaseDomain is required")
	}
	if cfg.Auth == nil {
		cfg.Auth = AllowAll{}
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	if cfg.HandshakeTimeout == 0 {
		cfg.HandshakeTimeout = 10 * time.Second
	}
	return &Relay{
		cfg:      cfg,
		registry: core.NewRegistry(cfg.BaseDomain),
		log:      cfg.Logger,
		sessions: make(map[string]*mux.Session),
	}, nil
}

// Registry exposes the routing table, for a status endpoint or a dashboard.
func (r *Relay) Registry() *core.Registry { return r.registry }

// ServeAgent runs one agent connection: handshake, registration, then serve
// until the connection ends. It blocks, so callers run it per accepted
// connection.
func (r *Relay) ServeAgent(conn net.Conn) {
	defer conn.Close()

	// The handshake happens on the raw connection, before the multiplexer, so
	// an unauthenticated peer never gets a stream. The deadline is cleared
	// once authenticated: a healthy tunnel is idle for long stretches and
	// must not be dropped for it.
	if err := conn.SetDeadline(time.Now().Add(r.cfg.HandshakeTimeout)); err != nil {
		r.log.Warn("could not set handshake deadline", "error", err)
	}

	frame, err := core.DecodeFrame(conn)
	if err != nil {
		r.log.Info("agent handshake failed", "error", err, "peer", conn.RemoteAddr())
		return
	}
	req, err := core.DecodeAuthRequest(frame)
	if err != nil {
		r.writeFrame(conn, core.ErrorFrame(err.Error()))
		r.log.Info("agent sent a bad auth request", "error", err, "peer", conn.RemoteAddr())
		return
	}

	resp, err := r.authorize(req)
	if err != nil {
		r.writeFrame(conn, core.ErrorFrame(err.Error()))
		r.log.Info("agent refused", "error", err, "peer", conn.RemoteAddr(), "requested", req.Subdomain)
		return
	}

	okFrame, err := core.EncodeAuthResponse(resp)
	if err != nil {
		r.log.Error("could not encode auth response", "error", err)
		r.registry.Unregister(resp.Subdomain)
		return
	}
	if err := r.writeFrame(conn, okFrame); err != nil {
		r.registry.Unregister(resp.Subdomain)
		return
	}
	if err := conn.SetDeadline(time.Time{}); err != nil {
		r.log.Warn("could not clear handshake deadline", "error", err)
	}

	session := mux.Server(conn)
	r.mu.Lock()
	r.sessions[resp.AgentID] = session
	r.mu.Unlock()

	r.log.Info("tunnel open",
		"agent", resp.AgentID, "subdomain", resp.Subdomain, "url", resp.URL,
		"local_port", req.LocalPort, "peer", conn.RemoteAddr())

	// Everything this agent owns is released when its connection ends —
	// whether it closed cleanly or the laptop lid shut.
	defer func() {
		r.mu.Lock()
		delete(r.sessions, resp.AgentID)
		r.mu.Unlock()
		freed := r.registry.UnregisterAgent(resp.AgentID)
		session.Close()
		r.log.Info("tunnel closed", "agent", resp.AgentID, "released", freed)
	}()

	<-session.CloseChan()
}

// authorize validates the request, picks a name, and registers the tunnel.
func (r *Relay) authorize(req core.AuthRequest) (core.AuthResponse, error) {
	if err := r.cfg.Auth.Authorize(req); err != nil {
		return core.AuthResponse{}, err
	}

	agentID, err := newAgentID()
	if err != nil {
		return core.AuthResponse{}, fmt.Errorf("relay: could not mint an agent id: %w", err)
	}

	name := req.Subdomain
	if name == "" {
		name, err = core.AllocateSubdomain(r.registry, nil)
		if err != nil {
			return core.AuthResponse{}, err
		}
	} else if err := core.ValidateSubdomain(name); err != nil {
		return core.AuthResponse{}, err
	}

	tunnel := core.Tunnel{
		Subdomain: name,
		AgentID:   agentID,
		LocalPort: req.LocalPort,
		Reserved:  req.Subdomain != "" && req.Token != "",
	}
	if err := r.registry.Register(tunnel); err != nil {
		return core.AuthResponse{}, err
	}

	return core.AuthResponse{
		Subdomain: name,
		URL:       r.registry.PublicURL(name),
		AgentID:   agentID,
	}, nil
}

func (r *Relay) writeFrame(w io.Writer, f core.Frame) error {
	wire, err := f.Encode()
	if err != nil {
		return err
	}
	_, err = w.Write(wire)
	return err
}

// ServeHTTP routes a visitor request to the agent that owns its hostname.
// It satisfies http.Handler, so the edge listener is an ordinary net/http
// server and TLS is whatever the operator wraps around it.
func (r *Relay) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	tunnel, err := r.registry.Lookup(req.Host)
	if err != nil {
		r.notFound(w, req, err)
		return
	}

	r.mu.RLock()
	session := r.sessions[tunnel.AgentID]
	r.mu.RUnlock()
	if session == nil {
		// The registry and the session map disagreed: the agent went away
		// between the lookup and here.
		r.notFound(w, req, fmt.Errorf("tunnel %q has no live session", tunnel.Subdomain))
		return
	}

	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.Out.URL.Scheme = "http"
			// The agent forwards to a fixed local address, so the authority
			// here is a placeholder; the agent never resolves it. The
			// visitor's original Host is preserved for the local app, which
			// commonly routes on it.
			pr.Out.URL.Host = tunnel.Subdomain
			pr.Out.Host = pr.In.Host
			pr.SetXForwarded()
		},
		Transport: &http.Transport{
			// Every outbound "dial" is a new stream on the agent's existing
			// connection — this is where the tunnel actually happens.
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				stream, err := session.Open(0)
				if err != nil {
					return nil, err
				}
				return stream.NetConn(), nil
			},
			// Streams are cheap and per-request; pooling them across requests
			// would keep an agent's stream table populated for no gain.
			DisableKeepAlives:     true,
			ResponseHeaderTimeout: 60 * time.Second,
		},
		ErrorHandler: func(w http.ResponseWriter, _ *http.Request, err error) {
			r.log.Info("forwarding failed", "subdomain", tunnel.Subdomain, "error", err)
			http.Error(w, "the tunnelled service did not respond", http.StatusBadGateway)
		},
	}
	proxy.ServeHTTP(w, req)
}

// notFound answers a request for a hostname nothing serves. It is plain text
// and carries no branding: this response is seen by strangers, and an
// operator's identity does not belong on it.
func (r *Relay) notFound(w http.ResponseWriter, req *http.Request, err error) {
	r.log.Debug("no route", "host", req.Host, "error", err)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "No tunnel is open for %s\n", req.Host)
}

func newAgentID() (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

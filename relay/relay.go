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

	// TCPPortLow and TCPPortHigh bound the public ports handed to raw TCP
	// tunnels (SSH, RDP, databases). Leave both zero to refuse TCP tunnels
	// and serve HTTP only.
	TCPPortLow, TCPPortHigh int
	// TCPBindHost is the interface public TCP ports bind on. Empty means all
	// interfaces, which is what a relay on a public host wants; tests set
	// 127.0.0.1.
	TCPBindHost string
	// PublicHost is the address a visitor dials for raw TCP, reported to the
	// agent so the CLI can print it. Defaults to BaseDomain.
	PublicHost string
	// PublicScheme is the scheme HTTP tunnel addresses are announced under —
	// core.SchemeHTTP, or core.SchemeHTTPS when a TLS terminator sits in
	// front of this relay. Empty means http, which is what the relay serves
	// on its own. New refuses anything else rather than announcing a scheme
	// the relay cannot honour.
	PublicScheme string
	// AgentPublicPort is the port agents dial, shown by the console's command
	// builder so a person can copy a command that actually works. Display
	// only — the listener's address is the caller's business.
	AgentPublicPort string
}

// Relay accepts agents and serves visitor traffic for them.
type Relay struct {
	cfg      Config
	registry *core.Registry
	log      *slog.Logger

	pool *core.PortPool // nil when this relay serves HTTP only

	mu           sync.RWMutex
	sessions     map[string]tunnelSession // agent id -> live client connection
	tcpListeners map[string][]*tcpListener
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
	if cfg.PublicHost == "" {
		cfg.PublicHost = cfg.BaseDomain
	}
	// Validated here, at startup, so an operator's typo is a refusal to start
	// rather than an address handed to every user of this relay.
	scheme, err := core.ParsePublicScheme(cfg.PublicScheme)
	if err != nil {
		return nil, err
	}
	cfg.PublicScheme = scheme

	r := &Relay{
		cfg:          cfg,
		registry:     core.NewRegistry(cfg.BaseDomain, cfg.PublicScheme),
		log:          cfg.Logger,
		sessions:     make(map[string]tunnelSession),
		tcpListeners: make(map[string][]*tcpListener),
	}
	if cfg.TCPPortLow != 0 || cfg.TCPPortHigh != 0 {
		pool, err := core.NewPortPool(cfg.TCPPortLow, cfg.TCPPortHigh)
		if err != nil {
			return nil, err
		}
		r.pool = pool
	}
	return r, nil
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

	resp, tcpPort, err := r.authorize(req)
	if err != nil {
		r.writeFrame(conn, core.ErrorFrame(err.Error()))
		r.log.Info("agent refused", "error", err, "peer", conn.RemoteAddr(), "requested", req.Subdomain)
		return
	}

	// Bind before the address leaves. The auth response carries TCPAddr, and
	// until the listener exists that string names nothing — so a bind failure
	// has to be the answer to the handshake rather than a correction sent
	// after the agent, and its user, already had a public address (spec/0002).
	var listener *tcpListener
	if resp.TCPAddr != "" {
		listener, err = r.listenTCP(resp.AgentID, tcpPort)
		if err != nil {
			r.log.Error("could not bind the public tcp port", "error", err)
			// Give the name and the port up before saying so, for the same
			// reason the bind comes before the announce: when the client learns
			// the outcome, the state behind it is already true. Telling it
			// first leaves a window in which the pool has released the port
			// (listenTCP does that) while the registry still records it, and
			// the next agent to ask is refused a port that is free.
			r.registry.UnregisterAgent(resp.AgentID)
			r.writeFrame(conn, core.ErrorFrame(err.Error()))
			return
		}
	}

	// From here a failure has to undo the listener too, or the port is held by
	// a tunnel that never opened.
	abandon := func() {
		if listener != nil {
			r.releaseTCP(resp.AgentID)
		}
		r.registry.UnregisterAgent(resp.AgentID)
	}

	okFrame, err := core.EncodeAuthResponse(resp)
	if err != nil {
		r.log.Error("could not encode auth response", "error", err)
		abandon()
		return
	}
	if err := r.writeFrame(conn, okFrame); err != nil {
		abandon()
		return
	}
	if err := conn.SetDeadline(time.Time{}); err != nil {
		r.log.Warn("could not clear handshake deadline", "error", err)
	}

	session := muxSession{s: mux.Server(conn)}
	r.mu.Lock()
	r.sessions[resp.AgentID] = session
	r.mu.Unlock()

	// The socket has been answering since before the announce; start draining
	// it now that there is a session to forward into. A visitor that arrived
	// in between waited in the accept queue and is served here.
	if listener != nil {
		go r.serveTCP(session, listener, resp.Subdomain)
	}

	r.log.Info("tunnel open",
		"agent", resp.AgentID, "subdomain", resp.Subdomain, "url", resp.URL,
		"tcp", resp.TCPAddr, "local_port", req.LocalPort, "peer", conn.RemoteAddr())

	// Everything this agent owns is released when its connection ends —
	// whether it closed cleanly or the laptop lid shut.
	defer func() {
		r.mu.Lock()
		delete(r.sessions, resp.AgentID)
		r.mu.Unlock()
		freedPorts := r.releaseTCP(resp.AgentID)
		freed := r.registry.UnregisterAgent(resp.AgentID)
		session.Close()
		r.log.Info("tunnel closed", "agent", resp.AgentID, "released", freed, "ports", freedPorts)
	}()

	<-session.CloseChan()
}

// authorize validates the request, picks a name, and registers the tunnel. It
// returns the public TCP port it allocated, or 0 for a request that asked for
// none. That port is returned rather than looked up again from the registry
// because the caller has to bind it before it announces the address, and a
// second route to a number the caller already holds is a place for the two to
// disagree — which is exactly what it used to do (spec/0002, L-007).
func (r *Relay) authorize(req core.AuthRequest) (core.AuthResponse, int, error) {
	if err := r.cfg.Auth.Authorize(req); err != nil {
		return core.AuthResponse{}, 0, err
	}

	agentID, err := newAgentID()
	if err != nil {
		return core.AuthResponse{}, 0, fmt.Errorf("relay: could not mint an agent id: %w", err)
	}

	name := req.Subdomain
	if name == "" {
		name, err = core.AllocateSubdomain(r.registry, nil)
		if err != nil {
			return core.AuthResponse{}, 0, err
		}
	} else if err := core.ValidateSubdomain(name); err != nil {
		return core.AuthResponse{}, 0, err
	}

	tunnel := core.Tunnel{
		Subdomain: name,
		AgentID:   agentID,
		LocalPort: req.LocalPort,
		Reserved:  req.Subdomain != "" && req.Token != "",
		Requested: req.Subdomain != "" || req.TCPPort != 0,
	}

	resp := core.AuthResponse{
		Subdomain: name,
		URL:       r.registry.PublicURL(name),
		AgentID:   agentID,
	}

	// A raw TCP tunnel needs its public port decided here, because the port
	// number is the address and it has to travel back in this response.
	if req.TCP {
		port, err := r.allocateTCPPort(agentID, req.TCPPort)
		if err != nil {
			return core.AuthResponse{}, 0, err
		}
		tunnel.TCPPort = port
		resp.TCPAddr = fmt.Sprintf("%s:%d", r.cfg.PublicHost, port)
	}

	if err := r.registry.Register(tunnel); err != nil {
		if tunnel.TCPPort != 0 {
			r.pool.Release(tunnel.TCPPort)
		}
		return core.AuthResponse{}, 0, err
	}
	return resp, tunnel.TCPPort, nil
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
	// The apex is the relay's own console, never a tunnel — SplitHost already
	// refuses to route it, so serving it here costs a tunnel nothing.
	if core.IsApexHost(req.Host, r.cfg.BaseDomain) {
		switch req.URL.Path {
		case "/_pumasi/status":
			r.serveStatus(w, req)
		case "/":
			r.serveDashboard(w, req)
		default:
			http.NotFound(w, req)
		}
		return
	}

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
				return session.OpenStream(false)
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

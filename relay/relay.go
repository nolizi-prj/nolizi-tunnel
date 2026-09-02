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
	"encoding/json"
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
	// Reservations records which names, and which public ports, belong to
	// which token. Nil means an empty set — not a disabled one: every name in
	// an empty set is unclaimed, which is exactly the behaviour of a relay
	// that had no such set at all, and the rules apply from the first claim
	// onward (spec/0004 §5.3).
	Reservations *core.Reservations
	// ReservationsPath is the file the reservation set is kept in, so that a
	// claim outlives the process (spec/0004 §14). Empty is today's relay in
	// every respect: no file, no lock, no sweep and no write. Setting this
	// AND Reservations is refused by New — two sources of truth for one set
	// is an operator's typo, and it belongs at startup rather than in a
	// silent choice between them.
	ReservationsPath string
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
	// SSHPublicPort is the port used by the zero-install stock SSH path.
	// Empty hides that path from the console.
	SSHPublicPort string
	// FeedbackGitHubToken is server-only and creates issues from the console.
	// Empty disables submission without putting a credential in the browser.
	FeedbackGitHubToken string
	FeedbackGitHubRepo  string
	FeedbackAPIURL      string
	FeedbackHTTPClient  *http.Client
}

// Relay accepts agents and serves visitor traffic for them.
type Relay struct {
	cfg      Config
	registry *core.Registry
	log      *slog.Logger

	pool         *core.PortPool // nil when this relay serves HTTP only
	reservations *core.Reservations
	// ownReservations is the set THIS relay opened from ReservationsPath, and
	// is nil when the caller supplied one or when there is no path. Close
	// closes this and never cfg.Reservations.
	ownReservations *core.Reservations

	mu               sync.RWMutex
	sessions         map[string]tunnelSession // agent id -> live client connection
	tcpListeners     map[string][]*tcpListener
	feedbackMu       sync.Mutex
	feedbackAttempts map[string][]time.Time
}

// New builds a relay.
//
// Anything New opens, it closes again on its own way out: the reservation
// store takes a lock for the life of the process (spec/0004 §14.3), and a New
// that returned an error while still holding one would make the next attempt
// on the same path fail for a reason that is not its own.
func New(cfg Config) (*Relay, error) {
	if cfg.BaseDomain == "" {
		return nil, fmt.Errorf("relay: BaseDomain is required")
	}
	if cfg.Auth == nil {
		cfg.Auth = AllowAll{}
	}
	if cfg.Reservations != nil && cfg.ReservationsPath != "" {
		return nil, fmt.Errorf("relay: Reservations and ReservationsPath are two sources of truth for one set; give one")
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
	if cfg.FeedbackGitHubRepo == "" {
		cfg.FeedbackGitHubRepo = "pumasi-ai/pumasi-tunnel"
	}
	if cfg.FeedbackAPIURL == "" {
		cfg.FeedbackAPIURL = "https://api.github.com"
	}
	if cfg.FeedbackHTTPClient == nil {
		cfg.FeedbackHTTPClient = &http.Client{Timeout: 15 * time.Second}
	}
	// Validated here, at startup, so an operator's typo is a refusal to start
	// rather than an address handed to every user of this relay. Before the
	// store is opened, so a bad scheme does not take and drop a lock.
	scheme, err := core.ParsePublicScheme(cfg.PublicScheme)
	if err != nil {
		return nil, err
	}
	cfg.PublicScheme = scheme

	var pool *core.PortPool
	if cfg.TCPPortLow != 0 || cfg.TCPPortHigh != 0 {
		pool, err = core.NewPortPool(cfg.TCPPortLow, cfg.TCPPortHigh)
		if err != nil {
			return nil, err
		}
	}

	// Last, because it is the only step that acquires something the process
	// has to give back. A path an operator cannot use is a refusal to start
	// rather than a promise of durability the relay cannot keep.
	var own *core.Reservations
	if cfg.ReservationsPath != "" {
		opened, err := core.OpenReservations(cfg.ReservationsPath, cfg.Logger)
		if err != nil {
			return nil, err
		}
		own = opened
		cfg.Reservations = opened
	}
	if cfg.Reservations == nil {
		cfg.Reservations = core.NewReservations()
	}

	// The pool is DERIVED from the reservation set, not built beside it. A set
	// that survived a restart says which public ports are spoken for, and a
	// pool that started empty would hand the next stranger asking for "any
	// port" a number this relay has just been told belongs to somebody —
	// which is half of spec/0004 §4's middle column quietly missing while the
	// name half looked delivered.
	if pool != nil {
		for _, held := range cfg.Reservations.Snapshot() {
			if held.TCPPort == 0 {
				continue
			}
			if err := pool.Hold(held.TCPPort, held.Subdomain); err != nil {
				// Out of this pool's range, or held twice. Neither stops the
				// relay: the name is still owned, and a port outside the
				// operator's current range is the operator having narrowed it
				// since the claim, not a fault in the claim.
				cfg.Logger.Warn("a reserved public port is not in this relay's pool",
					"subdomain", held.Subdomain, "port", held.TCPPort, "error", err)
			}
		}
	}

	return &Relay{
		cfg:              cfg,
		registry:         core.NewRegistry(cfg.BaseDomain, cfg.PublicScheme),
		log:              cfg.Logger,
		pool:             pool,
		reservations:     cfg.Reservations,
		ownReservations:  own,
		sessions:         make(map[string]tunnelSession),
		tcpListeners:     make(map[string][]*tcpListener),
		feedbackAttempts: make(map[string][]time.Time),
	}, nil
}

// Registry exposes the routing table, for a status endpoint or a dashboard.
func (r *Relay) Registry() *core.Registry { return r.registry }

// Close releases what New opened, and nothing else. A relay handed a
// Reservations it did not open does not close it: the caller who built the set
// owns its lifetime, and a Close that reached past what it opened would take a
// lock out from under whoever else was holding the set.
func (r *Relay) Close() error {
	if r.ownReservations == nil {
		return nil
	}
	return r.ownReservations.Close()
}

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

	resp, tcpPort, newClaim, err := r.authorize(req)
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
			r.discardClaim(newClaim)
			r.writeFrame(conn, core.ErrorFrame(err.Error()))
			return
		}
	}

	// From here a failure has to undo the listener too, or the port is held by
	// a tunnel that never opened.
	abandon := func() {
		if listener != nil {
			r.releaseTCP(resp.AgentID, resp.Subdomain)
		}
		r.registry.UnregisterAgent(resp.AgentID)
		// A name is not consumed by a connection that never opened, so a
		// claim this handshake created is undone with everything else it did
		// (spec/0004 §5.1). A claim it merely *verified* is left alone.
		r.discardClaim(newClaim)
	}

	okFrame, err := core.EncodeAuthResponse(resp)
	if err != nil {
		r.log.Error("could not encode auth response", "error", err)
		abandon()
		return
	}
	// Install the session before the announce, and announce while holding the
	// lock that guards it. The URL in this frame is live the moment the agent
	// reads it, so the thing that serves the URL has to be in place first
	// (spec/0003) — the same rule as the bind above, for the thing that
	// answers an HTTP hostname.
	//
	// r.mu is doing two jobs here, and the second is the reason the write is
	// inside it rather than after it. ServeHTTP takes this lock to read
	// r.sessions, so a visitor arriving now waits for the lock instead of
	// being told there is no tunnel — and, because it cannot reach OpenStream
	// without first holding the session, no stream frame can reach the wire
	// ahead of this frame. The agent is in DecodeFrame waiting for exactly
	// one frame and would decode a stream open as its auth response.
	//
	// The write is bounded: the handshake deadline set above is still armed
	// and is cleared only after this, so an agent that never reads costs at
	// most HandshakeTimeout rather than the routing table.
	session := muxSession{s: mux.Server(conn)}
	r.mu.Lock()
	r.sessions[resp.AgentID] = session
	err = r.writeFrame(conn, okFrame)
	if err != nil {
		// Undone in the same critical section: between two of them a visitor
		// would be handed a stream on a connection already known to be dead.
		delete(r.sessions, resp.AgentID)
	}
	r.mu.Unlock()
	if err != nil {
		session.Close()
		abandon()
		return
	}
	if err := conn.SetDeadline(time.Time{}); err != nil {
		r.log.Warn("could not clear handshake deadline", "error", err)
	}

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
		freedPorts := r.releaseTCP(resp.AgentID, resp.Subdomain)
		freed := r.registry.UnregisterAgent(resp.AgentID)
		session.Close()
		r.log.Info("tunnel closed", "agent", resp.AgentID, "released", freed, "ports", freedPorts)
	}()

	<-session.CloseChan()
}

// authorize validates the request, decides whether this caller may have the
// name it asked for, picks one if it asked for none, and registers the tunnel.
//
// It returns three things beyond the response. The public TCP port it
// allocated, or 0 — returned rather than looked up again from the registry,
// because the caller has to bind it before it announces the address and a
// second route to a number the caller already holds is a place for the two to
// disagree (spec/0002, L-007). And the subdomain whose reservation *this
// handshake created*, or "" — so a caller whose handshake fails further on can
// undo exactly what it did and no more (spec/0004 §5.1). A name that was
// already claimed, or was never claimed at all, comes back as "".
//
// Both ingress paths run through here — ServeAgent and ServeSSH — which is why
// the reservation rule is written once. A claim enforced on one path and not
// the other is a claim about a two-path system that is true of one path
// (lessons/L-009).
func (r *Relay) authorize(req core.AuthRequest) (core.AuthResponse, int, string, error) {
	if err := r.cfg.Auth.Authorize(req); err != nil {
		return core.AuthResponse{}, 0, "", err
	}

	agentID, err := newAgentID()
	if err != nil {
		return core.AuthResponse{}, 0, "", fmt.Errorf("relay: could not mint an agent id: %w", err)
	}

	name := req.Subdomain
	if name == "" {
		name, err = core.AllocateSubdomain(r.registry, nil)
		if err != nil {
			return core.AuthResponse{}, 0, "", err
		}
	} else if err := core.ValidateSubdomain(name); err != nil {
		return core.AuthResponse{}, 0, "", err
	}

	// Whose name is it? Only a name the caller asked for by name can be
	// claimed or refused: a generated one is fresh, is nobody's, and asking
	// the reservation set about it would be asking a question with one answer.
	//
	// A token long enough to claim with does; anything shorter takes the
	// read-only path, where Check refuses it as too short rather than
	// quietly treating it as no token at all. Being handed a name you believe
	// you protected is the worse of those two failures (spec/0004 §3.1).
	var newClaim string
	if req.Subdomain != "" {
		wantPort := 0
		if req.TCP {
			wantPort = req.TCPPort
		}
		if len(req.Token) >= core.MinTokenLen {
			created, err := r.reservations.Claim(name, req.Token, wantPort)
			if err != nil {
				return core.AuthResponse{}, 0, "", err
			}
			if created {
				newClaim = name
			}
		} else if err := r.reservations.Check(name, req.Token); err != nil {
			return core.AuthResponse{}, 0, "", err
		}
	}

	// Whatever the reservation says about a port has to reach the pool before
	// the pool is asked for one, or a generic request could be walking onto
	// this very number while this handshake is deciding it is spoken for.
	held, isReserved := r.reservations.Get(name)
	if isReserved && held.TCPPort != 0 && r.pool != nil {
		if err := r.pool.Hold(held.TCPPort, name); err != nil {
			r.discardClaim(newClaim)
			return core.AuthResponse{}, 0, "", err
		}
	}

	tunnel := core.Tunnel{
		Subdomain: name,
		AgentID:   agentID,
		LocalPort: req.LocalPort,
		// Read from the record rather than guessed from the shape of the
		// request. The old expression — a name was sent and a token was sent —
		// could not tell a name someone owns from one they merely asked for,
		// and nothing read it either way.
		Reserved:  isReserved,
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
		port, err := r.allocateTCPPort(name, req.TCPPort)
		if err != nil {
			r.discardClaim(newClaim)
			return core.AuthResponse{}, 0, "", err
		}
		tunnel.TCPPort = port
		resp.TCPAddr = fmt.Sprintf("%s:%d", r.cfg.PublicHost, port)
	}

	if err := r.registry.Register(tunnel); err != nil {
		if tunnel.TCPPort != 0 {
			r.pool.Release(tunnel.TCPPort)
		}
		r.discardClaim(newClaim)
		return core.AuthResponse{}, 0, "", err
	}
	return resp, tunnel.TCPPort, newClaim, nil
}

// discardClaim undoes a reservation that a handshake created and then failed
// to make use of, and does nothing at all for a name this handshake merely
// proved it already owned. Passing "" is the ordinary case and is a no-op.
//
// A name is not consumed by a connection that never opened. The cost of that
// choice is small and real and spec/0004 §5.1 states it: a transient bind
// failure on a first connection leaves the name claimable by a stranger for as
// long as the owner takes to retry.
//
// It will not pull a reservation out from under the tunnel that reservation
// belongs to, and the shape of that guard is narrower than it first looks.
//
// Two connections presenting the SAME token for the same name race: the first
// to Claim creates the reservation, the second finds it already its own and
// creates nothing. Whichever loses Register gets ErrNameTaken — and if that is
// the one that created the claim, discarding here would destroy the
// reservation while the winner is registered on it, leaving a tunnel the
// registry calls Reserved whose name is free the moment it drops.
//
// But "something is registered on this name" is the WRONG test for that, and
// getting it wrong is a defect of its own. The common case is an anonymous
// agent squatting an unclaimed name when a token-holder arrives: Claim
// succeeds, Register refuses ErrNameTaken, and suppressing the discard there
// would consume the name by a connection that never opened — precisely what
// spec/0004 §5.1 says cannot happen. The two are told apart by whether the
// LIVE tunnel is the claim's own: only a co-claimant registers Reserved on a
// name this handshake has just claimed, because before that claim there was
// nothing to be reserved by.
func (r *Relay) discardClaim(subdomain string) {
	if subdomain == "" {
		return
	}
	if live, err := r.registry.Lookup(subdomain + "." + r.cfg.BaseDomain); err == nil && live.Reserved {
		return
	}
	if held, ok := r.reservations.Get(subdomain); ok && held.TCPPort != 0 && r.pool != nil {
		r.pool.Unhold(held.TCPPort)
	}
	// The name is out of memory either way; the error says only whether the
	// store agrees, and spec/0004 §14.6 names the residual when it does not.
	// Logged at ERROR because nothing else in this process can act on it: the
	// record returns at the next restart and is then held for up to
	// ReservationTTL by a token whose handshake never opened a tunnel.
	if err := r.reservations.Discard(subdomain); err != nil {
		r.log.Error("a discarded claim could not be removed from the reservation store",
			"subdomain", subdomain, "error", err)
	}
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
		case "/version":
			writeJSON(w, http.StatusOK, map[string]any{"version": Version})
		case "/healthz":
			writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "version": Version})
		case "/readyz":
			writeJSON(w, http.StatusOK, map[string]any{"status": "ready", "version": Version})
		case "/_pumasi/feedback":
			r.serveFeedback(w, req)
		case "/install.sh":
			serveInstallScript(w, req)
		case "/_pumasi/modern-screenshot.js":
			serveModernScreenshot(w, req)
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
		// Teardown, and only teardown: the agent's connection ended between
		// the lookup above and here, and ServeAgent has removed the session
		// but not yet unregistered the route. The other direction — a URL
		// announced before its session existed — is closed by ServeAgent
		// installing the session and writing the auth response under this
		// same lock (spec/0003 §1), so a visitor that beats the announce
		// waits for it here rather than arriving in this branch.
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

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
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

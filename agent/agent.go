// Package agent is the side that runs on the developer's machine: it holds
// one outbound connection to a relay and forwards every stream arriving on it
// to a local address.
//
// Only outbound connections are made, which is what lets this work from
// behind NAT, a corporate firewall, or a mobile hotspot with no port
// forwarding and no inbound rule.
package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/pumasi-ai/pumasi-tunnel/core"
	"github.com/pumasi-ai/pumasi-tunnel/mux"
)

// Config configures an agent.
type Config struct {
	// RelayAddr is the relay's control address, "host:port".
	RelayAddr string
	// LocalAddr is what tunnelled traffic is forwarded to, e.g.
	// "127.0.0.1:8080".
	LocalAddr string
	// Subdomain requests a specific name; empty lets the relay choose.
	Subdomain string
	// Token authenticates a reserved subdomain.
	Token string
	// TCP asks the relay for a public TCP port instead of an HTTP hostname,
	// which is what SSH, RDP and database clients need.
	TCP bool
	// TCPPort requests one specific public port so the address survives
	// reconnects; zero lets the relay choose.
	TCPPort int
	// Dial opens the control connection; net.Dial when nil. Tests and a
	// future TLS transport substitute here.
	Dial func(ctx context.Context, addr string) (net.Conn, error)
	// Logger receives events; discarded when nil.
	Logger *slog.Logger
	// OnConnect is called with the relay's answer each time a tunnel opens,
	// so a CLI can print the public URL — including after a reconnect.
	OnConnect func(core.AuthResponse)
}

// Agent maintains one tunnel.
type Agent struct {
	cfg Config
	log *slog.Logger

	mu       sync.Mutex
	requests uint64
}

// New builds an agent.
func New(cfg Config) (*Agent, error) {
	if cfg.RelayAddr == "" {
		return nil, errors.New("agent: RelayAddr is required")
	}
	if cfg.LocalAddr == "" {
		return nil, errors.New("agent: LocalAddr is required")
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	if cfg.Dial == nil {
		cfg.Dial = func(ctx context.Context, addr string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "tcp", addr)
		}
	}
	return &Agent{cfg: cfg, log: cfg.Logger}, nil
}

// Requests reports how many streams this agent has served, for the status
// line the CLI prints.
func (a *Agent) Requests() uint64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.requests
}

// Run holds the tunnel open until ctx is cancelled or the relay refuses.
// It returns nil on a clean shutdown.
//
// A dropped connection is not an error the caller has to handle: Run
// reconnects with backoff, because a laptop that sleeps or a network that
// flaps is the normal case for this product, not an exception.
func (a *Agent) Run(ctx context.Context) error {
	backoff := time.Second
	const maxBackoff = 30 * time.Second

	for {
		err := a.runOnce(ctx)
		switch {
		case ctx.Err() != nil:
			return nil
		case errors.Is(err, errRefused):
			// The relay stated a reason — a taken name, a bad token. Retrying
			// would fail identically, so surface it instead of looping.
			return err
		case err != nil:
			a.log.Info("tunnel dropped, reconnecting", "error", err, "in", backoff)
		default:
			a.log.Info("tunnel closed by the relay, reconnecting", "in", backoff)
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(backoff):
		}
		if backoff < maxBackoff {
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
}

// errRefused marks a refusal that retrying cannot fix.
var errRefused = errors.New("agent: the relay refused the tunnel")

func (a *Agent) runOnce(ctx context.Context) error {
	conn, err := a.cfg.Dial(ctx, a.cfg.RelayAddr)
	if err != nil {
		return fmt.Errorf("agent: dialling relay: %w", err)
	}
	defer conn.Close()

	req := core.AuthRequest{
		Token:     a.cfg.Token,
		Subdomain: a.cfg.Subdomain,
		TCP:       a.cfg.TCP,
		TCPPort:   a.cfg.TCPPort,
	}
	if _, portStr, splitErr := net.SplitHostPort(a.cfg.LocalAddr); splitErr == nil {
		fmt.Sscanf(portStr, "%d", &req.LocalPort)
	}

	frame, err := core.EncodeAuthRequest(req)
	if err != nil {
		return err
	}
	wire, err := frame.Encode()
	if err != nil {
		return err
	}
	if _, err := conn.Write(wire); err != nil {
		return fmt.Errorf("agent: sending handshake: %w", err)
	}

	replyFrame, err := core.DecodeFrame(conn)
	if err != nil {
		return fmt.Errorf("agent: reading handshake reply: %w", err)
	}
	resp, err := core.DecodeAuthResponse(replyFrame)
	if err != nil {
		// A stated refusal is final; a malformed reply is a transport fault
		// worth retrying.
		if replyFrame.Type == core.FrameError {
			return fmt.Errorf("%w: %v", errRefused, err)
		}
		return err
	}

	a.log.Info("tunnel open", "url", resp.URL, "forwarding_to", a.cfg.LocalAddr)
	if a.cfg.OnConnect != nil {
		a.cfg.OnConnect(resp)
	}

	session := mux.Client(conn)
	defer session.Close()

	// Cancelling the context must tear the session down, or Accept below
	// would block until the relay noticed.
	go func() {
		select {
		case <-ctx.Done():
			session.Close()
		case <-session.CloseChan():
		}
	}()

	for {
		stream, err := session.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		a.mu.Lock()
		a.requests++
		a.mu.Unlock()
		go a.forward(stream)
	}
}

// forward pipes one stream to the local service and back. It is a byte copy
// in both directions: the agent does not parse HTTP, so WebSocket upgrades,
// server-sent events, and raw TCP all pass through unaltered.
func (a *Agent) forward(stream *mux.Stream) {
	defer stream.Close()

	local, err := net.DialTimeout("tcp", a.cfg.LocalAddr, 10*time.Second)
	if err != nil {
		// Nothing is listening locally. Closing the stream lets the relay
		// answer the visitor with a gateway error rather than hanging.
		a.log.Info("local service unreachable", "addr", a.cfg.LocalAddr, "error", err)
		return
	}
	defer local.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		// Visitor -> local service. Half-closing tells a local server the
		// request body is complete, which is what lets it answer.
		io.Copy(local, stream)
		if tcp, ok := local.(*net.TCPConn); ok {
			tcp.CloseWrite()
		}
	}()

	go func() {
		defer wg.Done()
		// Local service -> visitor.
		io.Copy(stream, local)
		stream.CloseWrite()
	}()

	wg.Wait()
}

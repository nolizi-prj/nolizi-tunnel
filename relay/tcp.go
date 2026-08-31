package relay

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/pumasi-ai/pumasi-tunnel/core"
	"github.com/pumasi-ai/pumasi-tunnel/mux"
)

// Raw TCP tunnels are what carry SSH, RDP, and database clients — the traffic
// that is not HTTP and so cannot be routed by a Host header. Each gets its own
// public port, because a port is the only addressing a plain TCP client has.
//
// This is the capability the incumbents charge for or refuse: ngrok reserves
// TCP addresses for paid accounts and additionally demands a card on file,
// and Cloudflare Tunnel requires client-side software to reach a raw port at
// all. Here it is ordinary.

// tcpListener is one public port bound on behalf of one agent.
type tcpListener struct {
	port     int
	ln       net.Listener
	agentID  string
	stopOnce sync.Once
}

func (t *tcpListener) stop() {
	t.stopOnce.Do(func() { t.ln.Close() })
}

// allocateTCPPort reserves a public port for an agent. It runs during the
// handshake, before any session exists, because the port number has to travel
// back in the auth response — the agent cannot be told its address afterwards.
//
// A requested port is honoured when free, so a reconnecting agent keeps the
// address its users already know. Refusing rather than silently substituting
// matters here: an agent asked for a specific port because something is
// configured to dial it, and quietly handing back a different one would look
// like success while nothing could reach it.
func (r *Relay) allocateTCPPort(agentID string, requested int) (int, error) {
	if r.pool == nil {
		return 0, errors.New("relay: this relay has no TCP port range configured")
	}
	if requested != 0 {
		if err := r.pool.AllocateSpecific(requested, agentID); err != nil {
			return 0, err
		}
		return requested, nil
	}
	return r.pool.Allocate(agentID)
}

// bindTCP binds an already-allocated port and serves it until the agent's
// session ends.
func (r *Relay) bindTCP(session *mux.Session, agentID, subdomain string, port int) error {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", r.cfg.TCPBindHost, port))
	if err != nil {
		// Hand the port back rather than leaking it: a port the pool believes
		// is in use but nothing is listening on is unrecoverable without a
		// restart.
		r.pool.Release(port)
		return fmt.Errorf("relay: binding public port %d: %w", port, err)
	}

	listener := &tcpListener{port: port, ln: ln, agentID: agentID}
	r.mu.Lock()
	r.tcpListeners[agentID] = append(r.tcpListeners[agentID], listener)
	r.mu.Unlock()

	go r.serveTCP(session, listener, subdomain)
	return nil
}

// serveTCP accepts visitors on a public port and gives each one a stream.
func (r *Relay) serveTCP(session *mux.Session, listener *tcpListener, subdomain string) {
	defer listener.stop()

	// A dead session must free the port immediately; otherwise the listener
	// would keep accepting connections that can never be forwarded.
	go func() {
		<-session.CloseChan()
		listener.stop()
	}()

	for {
		conn, err := listener.ln.Accept()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				r.log.Warn("tcp accept failed", "port", listener.port, "error", err)
			}
			return
		}

		stream, err := session.Open(core.FlagTCP)
		if err != nil {
			// The agent is gone; refuse this visitor rather than hold it open.
			conn.Close()
			r.log.Info("tcp visitor refused, agent is gone",
				"port", listener.port, "subdomain", subdomain, "error", err)
			return
		}

		r.log.Debug("tcp connection", "port", listener.port, "from", conn.RemoteAddr(), "stream", stream.ID())
		go pipe(conn, stream)
	}
}

// pipe copies bytes both ways between a visitor's connection and a stream,
// and does not return until both directions are finished. Nothing is parsed:
// SSH, RDP and Postgres all pass through as the opaque byte streams they are.
func pipe(conn net.Conn, stream *mux.Stream) {
	defer conn.Close()
	defer stream.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(stream, conn)
		// Tell the far side this direction is done, so a protocol that waits
		// for end-of-input can proceed instead of stalling.
		stream.CloseWrite()
	}()

	go func() {
		defer wg.Done()
		io.Copy(conn, stream)
		if tcp, ok := conn.(*net.TCPConn); ok {
			tcp.CloseWrite()
		}
	}()

	wg.Wait()
}

// releaseTCP closes every public port an agent held and returns them to the
// pool. Called when the agent's session ends, however it ended.
func (r *Relay) releaseTCP(agentID string) []int {
	r.mu.Lock()
	listeners := r.tcpListeners[agentID]
	delete(r.tcpListeners, agentID)
	r.mu.Unlock()

	freed := make([]int, 0, len(listeners))
	for _, l := range listeners {
		l.stop()
		freed = append(freed, l.port)
	}
	if r.pool != nil {
		r.pool.ReleaseOwner(agentID)
	}
	return freed
}

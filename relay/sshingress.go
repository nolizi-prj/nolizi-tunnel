package relay

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/pumasi-ai/pumasi-tunnel/core"
	"golang.org/x/crypto/ssh"
)

// Zero-install tunnelling over a stock ssh client:
//
//	ssh -p 2222 -R 0:localhost:8080 pumasi.link
//	ssh -p 2222 -R 0:localhost:22   tcp@pumasi.link
//	ssh -p 2222 -R 0:localhost:8080 myapi@pumasi.link
//
// Nothing is downloaded and no account is needed. Options ride in the SSH
// username, which is the only field a plain ssh client lets a person set
// without configuration — the same trick the incumbent study found (§1,
// checklist item 4), and the reason that product needs no binary at all.
//
// The relay speaks just enough of the SSH connection protocol for reverse
// forwarding: it accepts the client's "tcpip-forward" request, then opens a
// "forwarded-tcpip" channel per visitor (RFC 4254 §7).

// sshSession adapts an SSH connection to the relay's session interface.
type sshSession struct {
	conn   ssh.Conn
	closed chan struct{}
	once   sync.Once
}

func (s *sshSession) CloseChan() <-chan struct{} { return s.closed }
func (s *sshSession) Kind() string               { return "ssh" }

func (s *sshSession) Close() error {
	s.once.Do(func() { close(s.closed) })
	return s.conn.Close()
}

// forwardedTCPIP is the channel payload from RFC 4254 §7.2. The address and
// port are what the client asked to forward; a client that dialled
// `-R 0:localhost:8080` gets its assigned port echoed back here.
type forwardedTCPIP struct {
	Addr       string
	Port       uint32
	OriginAddr string
	OriginPort uint32
}

// OpenStream gives one visitor a channel back to the ssh client.
func (s *sshSession) OpenStream(bool) (net.Conn, error) {
	payload := ssh.Marshal(&forwardedTCPIP{
		Addr:       "localhost",
		Port:       uint32(sshForwardPort),
		OriginAddr: "pumasi-relay",
		OriginPort: 0,
	})
	ch, reqs, err := s.conn.OpenChannel("forwarded-tcpip", payload)
	if err != nil {
		return nil, fmt.Errorf("relay: opening ssh channel: %w", err)
	}
	go ssh.DiscardRequests(reqs)
	return &sshChannelConn{Channel: ch}, nil
}

// sshForwardPort is the port reported in the forwarded-tcpip payload. The
// client already knows which local service it bound, so the value only has to
// be consistent; ssh clients do not route on it for a `-R 0:` forward.
const sshForwardPort = 80

// sshChannelConn presents an SSH channel as a net.Conn, so the same proxy and
// pipe code serves both client kinds.
type sshChannelConn struct{ ssh.Channel }

type sshAddr struct{ label string }

func (a sshAddr) Network() string { return "ssh-channel" }
func (a sshAddr) String() string  { return a.label }

func (c *sshChannelConn) LocalAddr() net.Addr  { return sshAddr{"ssh/local"} }
func (c *sshChannelConn) RemoteAddr() net.Addr { return sshAddr{"ssh/remote"} }

// An SSH channel has no timer of its own, so the deadline setters report that
// plainly instead of pretending to arm one that could never fire. Cancellation
// is by closing the channel.
var errSSHDeadline = errors.New("relay: ssh channels have no deadlines; close the channel to cancel")

func (c *sshChannelConn) SetDeadline(time.Time) error      { return errSSHDeadline }
func (c *sshChannelConn) SetReadDeadline(time.Time) error  { return errSSHDeadline }
func (c *sshChannelConn) SetWriteDeadline(time.Time) error { return errSSHDeadline }

// Compile-time proof the adapter really satisfies net.Conn.
var _ net.Conn = (*sshChannelConn)(nil)

// sshOptions is what a client encoded in its username.
type sshOptions struct {
	Subdomain string
	TCP       bool
}

// parseSSHUser reads options out of the SSH username. The grammar is
// deliberately tiny, because a person types it by hand into an ssh command:
//
//	(empty) | anything unrecognised  -> HTTP tunnel, relay picks the name
//	tcp                              -> raw TCP tunnel, relay picks the port
//	<name>                           -> HTTP tunnel published at <name>
//	tcp+<name> / <name>+tcp          -> raw TCP, name recorded for the console
//
// An unrecognised value is treated as a name rather than refused: the default
// username is the local account name of whoever ran ssh, and rejecting that
// would break the zero-configuration case this whole path exists for.
func parseSSHUser(user string) sshOptions {
	var opts sshOptions
	for _, part := range strings.Split(strings.ToLower(strings.TrimSpace(user)), "+") {
		part = strings.TrimSpace(part)
		switch {
		case part == "" || part == "ssh" || part == "http" || part == "https":
			// Nothing to say; keep the defaults.
		case part == "tcp":
			opts.TCP = true
		default:
			// Only accept it as a name if it could actually be one; otherwise
			// ignore it and let the relay assign.
			if core.ValidateSubdomain(part) == nil {
				opts.Subdomain = part
			}
		}
	}
	return opts
}

// ServeSSH runs one incoming ssh connection as a tunnel. Callers run it per
// accepted connection, as with ServeAgent.
func (r *Relay) ServeSSH(nConn net.Conn, hostKey ssh.Signer) {
	cfg := &ssh.ServerConfig{
		// Anonymous by design: the product's promise is a working tunnel from
		// one command with no account. A relay that wants identity supplies an
		// Authenticator, which still runs below.
		NoClientAuth:  true,
		ServerVersion: "SSH-2.0-pumasi-tunnel",
	}
	cfg.AddHostKey(hostKey)

	sshConn, chans, globalReqs, err := ssh.NewServerConn(nConn, cfg)
	if err != nil {
		r.log.Info("ssh handshake failed", "error", err, "peer", nConn.RemoteAddr())
		nConn.Close()
		return
	}
	defer sshConn.Close()

	opts := parseSSHUser(sshConn.User())
	session := &sshSession{conn: sshConn, closed: make(chan struct{})}

	resp, err := r.authorize(core.AuthRequest{
		Subdomain:     opts.Subdomain,
		TCP:           opts.TCP,
		ClientVersion: string(sshConn.ClientVersion()),
	})
	if err != nil {
		// The client is a terminal, so the refusal has to be readable there —
		// a protocol-level error would show as an opaque disconnect.
		r.sshTell(chans, "pumasi: "+err.Error())
		r.log.Info("ssh tunnel refused", "error", err, "peer", nConn.RemoteAddr())
		return
	}

	r.mu.Lock()
	r.sessions[resp.AgentID] = session
	r.mu.Unlock()

	if resp.TCPAddr != "" {
		tunnel, lookupErr := r.registry.Lookup(resp.Subdomain + "." + r.cfg.BaseDomain)
		if lookupErr == nil {
			if err := r.bindTCP(session, resp.AgentID, resp.Subdomain, tunnel.TCPPort); err != nil {
				r.sshTell(chans, "pumasi: "+err.Error())
				r.registry.UnregisterAgent(resp.AgentID)
				return
			}
		}
	}

	address := resp.URL
	if resp.TCPAddr != "" {
		address = resp.TCPAddr
	}
	r.log.Info("tunnel open", "agent", resp.AgentID, "via", "ssh",
		"subdomain", resp.Subdomain, "url", address, "peer", nConn.RemoteAddr())

	defer func() {
		r.mu.Lock()
		delete(r.sessions, resp.AgentID)
		r.mu.Unlock()
		ports := r.releaseTCP(resp.AgentID)
		freed := r.registry.UnregisterAgent(resp.AgentID)
		session.Close()
		r.log.Info("tunnel closed", "agent", resp.AgentID, "via", "ssh", "released", freed, "ports", ports)
	}()

	// A person running ssh by hand needs to be told the address; there is no
	// CLI of ours to print it.
	go r.sshGreet(chans, address)

	for req := range globalReqs {
		switch req.Type {
		case "tcpip-forward":
			// This is the `-R` the client asked for. The reply carries the
			// bound port when the client requested port 0.
			if req.WantReply {
				req.Reply(true, binary.BigEndian.AppendUint32(nil, uint32(sshForwardPort)))
			}
		case "cancel-tcpip-forward":
			if req.WantReply {
				req.Reply(true, nil)
			}
			return
		default:
			if req.WantReply {
				req.Reply(false, nil)
			}
		}
	}
}

// sshGreet accepts the client's session channel (ssh opens one unless -N was
// passed) and prints the tunnel's address into the user's terminal.
func (r *Relay) sshGreet(chans <-chan ssh.NewChannel, address string) {
	for newCh := range chans {
		if newCh.ChannelType() != "session" {
			newCh.Reject(ssh.UnknownChannelType, "only session channels are accepted")
			continue
		}
		ch, reqs, err := newCh.Accept()
		if err != nil {
			return
		}
		go func() {
			for req := range reqs {
				// Refuse shell/exec/pty: this is a tunnel endpoint, not a
				// shell host. Saying no explicitly is what makes that safe.
				switch req.Type {
				case "shell", "exec", "pty-req", "env", "subsystem":
					if req.WantReply {
						req.Reply(false, nil)
					}
				default:
					if req.WantReply {
						req.Reply(false, nil)
					}
				}
			}
		}()

		fmt.Fprintf(ch, "\r\n  %s\r\n\r\n  forwarding to your local port over this ssh session\r\n  press ctrl-c to close the tunnel\r\n\r\n", address)
		// Hold the channel open so the message stays on screen for as long as
		// the tunnel lives; closing it here would clear some clients.
		go func() {
			io.Copy(io.Discard, ch)
			ch.Close()
		}()
	}
}

// sshTell delivers one line to the client's terminal and hangs up. Used for
// refusals, which otherwise reach the user as an unexplained disconnect.
func (r *Relay) sshTell(chans <-chan ssh.NewChannel, msg string) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		for newCh := range chans {
			if newCh.ChannelType() != "session" {
				newCh.Reject(ssh.UnknownChannelType, msg)
				continue
			}
			ch, reqs, err := newCh.Accept()
			if err != nil {
				return
			}
			go ssh.DiscardRequests(reqs)
			fmt.Fprintf(ch, "\r\n  %s\r\n\r\n", msg)
			ch.Close()
			return
		}
	}()
	// Do not wait forever for a client that opened no session channel: a
	// refusal must not hold the goroutine open for the life of the process.
	select {
	case <-done:
	case <-time.After(3 * time.Second):
	}
}

// LoadOrCreateHostKey returns the relay's SSH host key, generating and saving
// one if the file is absent. A stable key matters: clients pin it in
// known_hosts, and a key that changed on every restart would greet every
// returning user with a host-key warning.
func LoadOrCreateHostKey(path string, generate func() ([]byte, error), read func(string) ([]byte, error), write func(string, []byte) error) (ssh.Signer, error) {
	pem, err := read(path)
	if err != nil {
		pem, err = generate()
		if err != nil {
			return nil, fmt.Errorf("relay: generating ssh host key: %w", err)
		}
		if err := write(path, pem); err != nil {
			return nil, fmt.Errorf("relay: saving ssh host key: %w", err)
		}
	}
	signer, err := ssh.ParsePrivateKey(pem)
	if err != nil {
		return nil, fmt.Errorf("relay: parsing ssh host key: %w", err)
	}
	return signer, nil
}

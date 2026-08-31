package relay_test

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/pumasi-ai/pumasi-tunnel/agent"
	"github.com/pumasi-ai/pumasi-tunnel/core"
	"github.com/pumasi-ai/pumasi-tunnel/relay"
)

// Frozen acceptance cases for spec/0002-bind-before-announce.
//
// The defect is an ordering, and an ordering can be observed without timing by
// making the second step fail: occupy the relay's whole public port range, and
// a relay that binds first can only refuse, while a relay that announces first
// hands out an address and corrects itself afterwards. The two differ in what
// the client ends up holding, not in when it holds it — so nothing here sleeps,
// polls, or hopes to lose a race (L-006).
//
// A second family of cases uses a TCP-only relay (BaseDomain: ""), a legitimate
// `-domain ""` invocation. There the registry lookup these paths used to make
// fails for every session, which turns the ssh path's silent skip from a race
// into a certainty.

// These cases draw from a range of their own, so they cannot collide with the
// 34000-series harness in tcp_test.go, and each case takes a block of it so
// they cannot collide with each other either.
const (
	// Deliberately below /proc/sys/net/ipv4/ip_local_port_range's floor
	// (32768 on this machine). A fixed test port inside the ephemeral range
	// can be taken transiently by any outgoing connection on the host, which
	// makes a bind failure look like the defect under test.
	bindOrderBase = 20500
	sshTCPUser    = "tcp+remotebox"

	// sshGreetingMarker is the sentence sshGreet prints and sshTell does not,
	// so it tells an announcement apart from a refusal.
	sshGreetingMarker = "forwarding to your local port over this ssh session"

	// lookupBreakingDomain is an ordinary domain with a leading space — what a
	// stray character in a systemd unit produces. relay.New accepts it (it
	// tests only for the empty string) while core.NewRegistry trims it, so the
	// lookup key the serve paths used to build never matched and the bind was
	// skipped for every session. relay.New rejects "" outright, which is why
	// that is not the trigger used here (SPEC 6.1).
	lookupBreakingDomain = " pumasi.link"
)

// bindOrderPorts hands each case a block of ten ports of its own.
//
// These cases originally shared one fixed port. Nothing in the relay's contract
// promises the public listener is closed by the time the case that opened it
// returns: releaseTCP runs when the relay notices the agent session ended, and
// that is concurrent with the case's own cleanup. So the next case's occupy()
// could race a listener that was still closing and fail EADDRINUSE — on the
// harness, not on the defect under test, 1 run in 40 (SPEC 6.5). Giving each
// case a block of its own removes the shared resource rather than waiting on
// it, so this stays a suite that never sleeps and never retries.
var bindOrderCursor int32 = bindOrderBase

func bindOrderPorts(t *testing.T) int {
	t.Helper()
	return int(atomic.AddInt32(&bindOrderCursor, 10)) - 10
}

// occupy holds a port for the duration of a test, so the relay's bind of the
// same port is guaranteed to fail.
func occupy(t *testing.T, port int) net.Listener {
	t.Helper()
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("could not occupy port %d for the test: %v", port, err)
	}
	t.Cleanup(func() { ln.Close() })
	return ln
}

// bindOrderRelay starts a relay with a one-port TCP range, plus an agent
// listener and an ssh listener, and returns their addresses.
type bindOrderRelay struct {
	relay     *relay.Relay
	agentAddr string
	sshAddr   string
	localAddr string
}

func newBindOrderRelay(t *testing.T, baseDomain string, low, high int) *bindOrderRelay {
	t.Helper()

	// A line-echo server standing in for whatever the tunnel publishes.
	localLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("local listener: %v", err)
	}
	t.Cleanup(func() { localLn.Close() })
	go func() {
		for {
			conn, err := localLn.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				scanner := bufio.NewScanner(c)
				for scanner.Scan() {
					fmt.Fprintf(c, "echo: %s\n", scanner.Text())
				}
			}(conn)
		}
	}()

	r, err := relay.New(relay.Config{
		BaseDomain:  baseDomain,
		TCPPortLow:  low,
		TCPPortHigh: high,
		TCPBindHost: "127.0.0.1",
		PublicHost:  "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("relay.New: %v", err)
	}

	agentLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("agent listener: %v", err)
	}
	t.Cleanup(func() { agentLn.Close() })
	go func() {
		for {
			conn, err := agentLn.Accept()
			if err != nil {
				return
			}
			go r.ServeAgent(conn)
		}
	}()

	pem, err := relay.GenerateHostKeyPEM()
	if err != nil {
		t.Fatalf("host key: %v", err)
	}
	signer, err := ssh.ParsePrivateKey(pem)
	if err != nil {
		t.Fatalf("parsing host key: %v", err)
	}
	sshLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("ssh listener: %v", err)
	}
	t.Cleanup(func() { sshLn.Close() })
	go func() {
		for {
			conn, err := sshLn.Accept()
			if err != nil {
				return
			}
			go r.ServeSSH(conn, signer)
		}
	}()

	return &bindOrderRelay{
		relay:     r,
		agentAddr: agentLn.Addr().String(),
		sshAddr:   sshLn.Addr().String(),
		localAddr: localLn.Addr().String(),
	}
}

// handshake speaks the agent protocol by hand rather than using the agent
// package, so that the case can assert on exactly which frames the relay sent
// and in what order — which is the whole subject here.
func (b *bindOrderRelay) handshake(t *testing.T, req core.AuthRequest) []core.Frame {
	t.Helper()
	conn, err := net.DialTimeout("tcp", b.agentAddr, 5*time.Second)
	if err != nil {
		t.Fatalf("dialling the relay's agent port: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	frame, err := core.EncodeAuthRequest(req)
	if err != nil {
		t.Fatalf("encoding the auth request: %v", err)
	}
	wire, err := frame.Encode()
	if err != nil {
		t.Fatalf("encoding the frame: %v", err)
	}
	if _, err := conn.Write(wire); err != nil {
		t.Fatalf("sending the auth request: %v", err)
	}

	// Read every frame the relay sends before it stops talking. A relay that
	// announces first and fails second sends two, and closes the connection
	// after a refusal — so a refused handshake ends at EOF rather than at the
	// deadline.
	var frames []core.Frame
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	for {
		f, err := core.DecodeFrame(conn)
		if err != nil {
			return frames
		}
		frames = append(frames, f)
		if len(frames) == 4 {
			return frames
		}
		// An accepted TCP tunnel is held open by the relay, so waiting for a
		// second frame here would wait out the whole deadline on every run.
		// The answer has arrived; stop reading.
		if resp, err := core.DecodeAuthResponse(f); err == nil && resp.TCPAddr != "" {
			return frames
		}
	}
}

// sshGreeting connects with a real ssh client, the way a person with nothing
// of ours installed does, and returns what the relay printed to the terminal.
func (b *bindOrderRelay) sshGreeting(t *testing.T, user string) string {
	t.Helper()
	greeting, _ := b.sshGreetingWithClient(t, user)
	return greeting
}

// sshGreetingWithClient is sshGreeting, returning the client too so a caller
// can serve the reverse forward before reading the greeting.
func (b *bindOrderRelay) sshGreetingWithClient(t *testing.T, user string) (string, *ssh.Client) {
	t.Helper()
	client, err := ssh.Dial("tcp", b.sshAddr, &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	})
	if err != nil {
		t.Fatalf("ssh dial: %v", err)
	}
	t.Cleanup(func() { client.Close() })
	b.serveReverseForward(t, client)

	ch, reqs, err := client.OpenChannel("session", nil)
	if err != nil {
		t.Fatalf("opening session channel: %v", err)
	}
	go ssh.DiscardRequests(reqs)
	defer ch.Close()

	type readResult struct {
		n   int
		err error
	}
	buf := make([]byte, 512)
	done := make(chan readResult, 1)
	go func() {
		n, err := ch.Read(buf)
		done <- readResult{n, err}
	}()
	select {
	case res := <-done:
		if res.err != nil && res.n == 0 {
			t.Fatalf("reading the ssh greeting: %v", res.err)
		}
		return string(buf[:res.n]), client
	case <-time.After(10 * time.Second):
		t.Fatal("the ssh ingress said nothing at all")
		return "", nil
	}
}

// serveReverseForward makes the test's ssh client behave like `ssh -R`: it
// accepts the "forwarded-tcpip" channels the relay opens for each visitor and
// serves each one from the local echo server. Without this the client is not
// the thing the product is for, and the announced address would be bound but
// unable to carry anything — which is a different property from the one under
// test.
func (b *bindOrderRelay) serveReverseForward(t *testing.T, client *ssh.Client) {
	t.Helper()
	chans := client.HandleChannelOpen("forwarded-tcpip")
	go func() {
		for newCh := range chans {
			ch, reqs, err := newCh.Accept()
			if err != nil {
				return
			}
			go ssh.DiscardRequests(reqs)
			go func(c ssh.Channel) {
				defer c.Close()
				local, err := net.DialTimeout("tcp", b.localAddr, 5*time.Second)
				if err != nil {
					return
				}
				defer local.Close()
				go io.Copy(local, c)
				io.Copy(c, local)
			}(ch)
		}
	}()
}

// echoThrough proves an address is not merely bound but wired to the tunnel:
// it dials, sends a line, and requires the echo back.
func echoThrough(addr, msg string) error {
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("dialling the announced address %s: %w", addr, err)
	}
	defer conn.Close()
	if _, err := fmt.Fprintln(conn, msg); err != nil {
		return fmt.Errorf("writing to %s: %w", addr, err)
	}
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return fmt.Errorf("reading from %s: %w", addr, err)
	}
	if got, want := strings.TrimSpace(line), "echo: "+msg; got != want {
		return fmt.Errorf("got %q through %s, want %q", got, addr, want)
	}
	return nil
}

// B-1 · Agent path, ordering. With the public port already held by something
// else, the relay cannot bind it. A relay that binds before it announces has
// only one thing it can say: no. A relay that announces first says yes, hands
// over a public address, and takes it back — and that first answer is the
// defect, because the agent and its user already have the address.
func TestAgentIsNotToldAnAddressItCannotBeGiven(t *testing.T) {
	port := bindOrderPorts(t)
	occupy(t, port)
	b := newBindOrderRelay(t, "pumasi.link", port, port)

	frames := b.handshake(t, core.AuthRequest{Subdomain: "remotebox", TCP: true})
	if len(frames) == 0 {
		t.Fatal("the relay answered the handshake with nothing at all")
	}

	for i, f := range frames {
		resp, err := core.DecodeAuthResponse(f)
		if err != nil {
			continue // a refusal frame: that is the allowed answer
		}
		if resp.TCPAddr != "" {
			t.Errorf("frame %d of %d announced the public address %q, "+
				"but nothing was listening on it and the relay knew that only afterwards",
				i+1, len(frames), resp.TCPAddr)
		}
	}

	// The refusal must actually arrive; silence would be a different defect.
	var refused bool
	for _, f := range frames {
		if f.Type == core.FrameError {
			refused = true
		}
	}
	if !refused {
		t.Errorf("the relay never told the agent the port could not be bound; frames = %d", len(frames))
	}
}

// B-2 · SSH path, ordering. The same construction over a stock ssh client. The
// greeting is what a person sees, so an address in it is an address they will
// use.
func TestSSHIsNotToldAnAddressItCannotBeGiven(t *testing.T) {
	port := bindOrderPorts(t)
	occupy(t, port)
	b := newBindOrderRelay(t, "pumasi.link", port, port)

	greeting := b.sshGreeting(t, sshTCPUser)

	// The relay has two things it can say, and they are distinguishable by
	// which one it said rather than by whether the address appears: a refusal
	// is sshTell's "pumasi: <error>", an announcement is sshGreet's address
	// line followed by the forwarding sentence. The bind error names the
	// address it could not bind, so the bare string proves nothing (SPEC 6.2).
	if strings.Contains(greeting, sshGreetingMarker) {
		t.Errorf("the ssh terminal was given the tunnel announcement, but the "+
			"relay could not bind port %d; greeting = %q", port, greeting)
	}
	if !strings.Contains(greeting, "pumasi: ") || !strings.Contains(greeting, "binding public port") {
		t.Errorf("the ssh terminal was not told the bind failed; greeting = %q", greeting)
	}
}

// B-3 · Agent path, no lookup. On a TCP-only relay the subdomain lookup these
// paths used to make fails for every session, because core.SplitHost("x.", "")
// is ErrForeignHost. The address announced must still be one that answers.
func TestTCPOnlyRelayAnnouncesAnAddressThatAnswers(t *testing.T) {
	port := bindOrderPorts(t)
	b := newBindOrderRelay(t, lookupBreakingDomain, port, port+9)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	connected := make(chan core.AuthResponse, 1)
	ag, err := agent.New(agent.Config{
		RelayAddr: b.agentAddr,
		LocalAddr: b.localAddr,
		Subdomain: "remotebox",
		TCP:       true,
		OnConnect: func(resp core.AuthResponse) { connected <- resp },
	})
	if err != nil {
		t.Fatalf("agent.New: %v", err)
	}
	go ag.Run(ctx)

	select {
	case resp := <-connected:
		if resp.TCPAddr == "" {
			t.Fatal("no public TCP address was announced for a TCP tunnel")
		}
		if err := echoThrough(resp.TCPAddr, "b3"); err != nil {
			t.Errorf("the announced address does not answer: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the agent never connected")
	}
}

// B-4 · SSH path, the silent skip. Same relay, over ssh. The old code skipped
// the bind when the lookup failed, said nothing about it in the log, and
// greeted the terminal with the address anyway.
func TestSSHTCPOnlyRelayNeverGreetsADeadAddress(t *testing.T) {
	port := bindOrderPorts(t)
	b := newBindOrderRelay(t, lookupBreakingDomain, port, port+9)

	greeting := b.sshGreeting(t, sshTCPUser)

	// Whatever address it printed, it has to be one that answers. Pull it back
	// out of the greeting rather than assuming which port was allocated.
	var announced string
	for _, field := range strings.Fields(greeting) {
		if strings.HasPrefix(field, "127.0.0.1:") {
			announced = strings.Trim(field, ",")
			break
		}
	}
	if announced == "" {
		// Refusing is an acceptable answer; greeting with a dead address is not.
		if !strings.Contains(strings.ToLower(greeting), "pumasi:") {
			t.Errorf("the ssh greeting neither announced an address nor refused: %q", greeting)
		}
		return
	}
	if err := echoThrough(announced, "b4"); err != nil {
		t.Errorf("the ssh greeting announced %q but it does not answer: %v", announced, err)
	}
}

// B-5 · A refused bind must not strand the port. The pool believing a port is
// in use while nothing listens on it cannot be recovered without a restart.
func TestPortReturnsToThePoolWhenTheBindFails(t *testing.T) {
	port := bindOrderPorts(t)
	blocker := occupy(t, port)
	b := newBindOrderRelay(t, "pumasi.link", port, port)

	// First agent: the port is held, so it is refused.
	b.handshake(t, core.AuthRequest{Subdomain: "first", TCP: true})

	// The obstruction goes away.
	blocker.Close()

	// Second agent: the one port in the range must be available again.
	frames := b.handshake(t, core.AuthRequest{Subdomain: "second", TCP: true})
	var addr string
	for _, f := range frames {
		if resp, err := core.DecodeAuthResponse(f); err == nil && resp.TCPAddr != "" {
			addr = resp.TCPAddr
		}
	}
	if addr == "" {
		t.Fatalf("after a failed bind the port never returned to the pool; "+
			"the second agent got %d frame(s) and no address", len(frames))
	}
}

// B-6 · The happy path is unchanged: splitting the bind from the accept loop
// must not drop the accept loop, the session wiring, or the byte pipe. The
// address announced is the address that works.
func TestAnnouncedAddressIsTheWorkingAddress(t *testing.T) {
	port := bindOrderPorts(t)
	b := newBindOrderRelay(t, "pumasi.link", port, port+9)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	connected := make(chan core.AuthResponse, 1)
	ag, err := agent.New(agent.Config{
		RelayAddr: b.agentAddr,
		LocalAddr: b.localAddr,
		Subdomain: "remotebox",
		TCP:       true,
		OnConnect: func(resp core.AuthResponse) { connected <- resp },
	})
	if err != nil {
		t.Fatalf("agent.New: %v", err)
	}
	go ag.Run(ctx)

	select {
	case resp := <-connected:
		if err := echoThrough(resp.TCPAddr, "b6-there"); err != nil {
			t.Errorf("first exchange: %v", err)
		}
		// Twice, because a listener that serves one connection and stops is a
		// different bug that one exchange would not see.
		if err := echoThrough(resp.TCPAddr, "b6-and-back"); err != nil {
			t.Errorf("second exchange: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the agent never connected")
	}
}

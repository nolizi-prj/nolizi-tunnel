package relay_test

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/pumasi-ai/pumasi-tunnel/agent"
	"github.com/pumasi-ai/pumasi-tunnel/core"
	"github.com/pumasi-ai/pumasi-tunnel/relay"
)

// tcpHarness runs a relay with a TCP port range, an agent asking for a raw
// TCP tunnel, and a local line-oriented server standing in for sshd. Every
// hop is a real socket.
type tcpHarness struct {
	publicAddr string
	cancel     context.CancelFunc
}

func newTCPHarness(t *testing.T, serve func(net.Conn)) *tcpHarness {
	t.Helper()

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
			go serve(conn)
		}
	}()

	r, err := relay.New(relay.Config{
		BaseDomain: "pumasi.link",
		// A high, unprivileged range; the test only needs one free port.
		TCPPortLow:  34000,
		TCPPortHigh: 34099,
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

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	connected := make(chan core.AuthResponse, 1)
	ag, err := agent.New(agent.Config{
		RelayAddr: agentLn.Addr().String(),
		LocalAddr: localLn.Addr().String(),
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
			t.Fatal("relay did not report a public TCP address for a TCP tunnel")
		}
		return &tcpHarness{publicAddr: resp.TCPAddr, cancel: cancel}
	case <-time.After(10 * time.Second):
		t.Fatal("agent did not connect")
		return nil
	}
}

// The capability the incumbents gate: a plain TCP client, with no helper
// software, reaching a service that listens only on loopback.
func TestRawTCPCrossesTheTunnel(t *testing.T) {
	// A line-echo server standing in for sshd: read a line, answer it.
	h := newTCPHarness(t, func(conn net.Conn) {
		defer conn.Close()
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Fprintf(conn, "echo: %s\n", scanner.Text())
		}
	})

	conn, err := net.DialTimeout("tcp", h.publicAddr, 10*time.Second)
	if err != nil {
		t.Fatalf("dialling the public TCP address: %v", err)
	}
	defer conn.Close()

	if _, err := fmt.Fprintln(conn, "SSH-2.0-pumasi-test"); err != nil {
		t.Fatalf("write: %v", err)
	}
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if strings.TrimSpace(line) != "echo: SSH-2.0-pumasi-test" {
		t.Errorf("got %q through the raw TCP tunnel", line)
	}
}

// A protocol where the server speaks first — like SSH's version banner —
// must work, which requires the relay to open the stream on connect rather
// than waiting for the client to send something.
func TestServerSpeaksFirstOverTCP(t *testing.T) {
	h := newTCPHarness(t, func(conn net.Conn) {
		defer conn.Close()
		fmt.Fprintln(conn, "SSH-2.0-OpenSSH_9.6")
		io.Copy(io.Discard, conn)
	})

	conn, err := net.DialTimeout("tcp", h.publicAddr, 10*time.Second)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	banner, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		t.Fatalf("reading the server's banner: %v", err)
	}
	if !strings.HasPrefix(banner, "SSH-2.0-") {
		t.Errorf("banner = %q, want an SSH version string", banner)
	}
}

// Several clients on one public port must not see each other's bytes.
func TestConcurrentTCPClients(t *testing.T) {
	h := newTCPHarness(t, func(conn net.Conn) {
		defer conn.Close()
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Fprintf(conn, "echo: %s\n", scanner.Text())
		}
	})

	const clients = 10
	errs := make(chan error, clients)
	for i := 0; i < clients; i++ {
		go func(i int) {
			conn, err := net.DialTimeout("tcp", h.publicAddr, 10*time.Second)
			if err != nil {
				errs <- err
				return
			}
			defer conn.Close()

			msg := fmt.Sprintf("client-%d", i)
			fmt.Fprintln(conn, msg)
			conn.SetReadDeadline(time.Now().Add(10 * time.Second))
			line, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				errs <- err
				return
			}
			want := "echo: " + msg
			if strings.TrimSpace(line) != want {
				errs <- fmt.Errorf("got %q, want %q — clients crossed streams", strings.TrimSpace(line), want)
				return
			}
			errs <- nil
		}(i)
	}
	for i := 0; i < clients; i++ {
		if err := <-errs; err != nil {
			t.Error(err)
		}
	}
}

// When the agent goes away the public port must stop accepting and return to
// the pool, or the relay slowly runs out of ports.
func TestTCPPortReleasedWhenAgentDisconnects(t *testing.T) {
	h := newTCPHarness(t, func(conn net.Conn) { conn.Close() })

	// Prove the port is live before tearing the agent down.
	probe, err := net.DialTimeout("tcp", h.publicAddr, 5*time.Second)
	if err != nil {
		t.Fatalf("port should be live before disconnect: %v", err)
	}
	probe.Close()

	h.cancel() // the agent stops

	// The listener closes asynchronously; poll briefly rather than sleeping a
	// fixed amount and hoping.
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", h.publicAddr, time.Second)
		if err != nil {
			return // refused: the port was released
		}
		conn.Close()
		time.Sleep(200 * time.Millisecond)
	}
	t.Error("the public TCP port still accepted connections after the agent disconnected")
}

// A relay with no TCP range configured must refuse a TCP tunnel clearly,
// rather than accept it and silently never forward anything.
func TestTCPRefusedWhenNoRangeConfigured(t *testing.T) {
	r, err := relay.New(relay.Config{BaseDomain: "pumasi.link"}) // HTTP only
	if err != nil {
		t.Fatalf("relay.New: %v", err)
	}
	agentLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer agentLn.Close()
	go func() {
		for {
			conn, err := agentLn.Accept()
			if err != nil {
				return
			}
			go r.ServeAgent(conn)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ag, err := agent.New(agent.Config{
		RelayAddr: agentLn.Addr().String(),
		LocalAddr: "127.0.0.1:22",
		TCP:       true,
	})
	if err != nil {
		t.Fatalf("agent.New: %v", err)
	}

	done := make(chan error, 1)
	go func() { done <- ag.Run(ctx) }()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected a refusal from an HTTP-only relay")
		}
		if !strings.Contains(err.Error(), "TCP port range") {
			t.Errorf("error = %v, want it to name the missing TCP range", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("agent was neither connected nor refused")
	}
}

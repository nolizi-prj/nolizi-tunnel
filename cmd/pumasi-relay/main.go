// Command pumasi-relay is the public side of a tunnel. It listens for agents
// on one port and for visitors on another, and forwards between them.
//
// TLS is deliberately not terminated here: run this behind a TLS listener or
// a reverse proxy that holds the wildcard certificate. Keeping certificate
// handling out of the relay means an operator can choose ACME, a purchased
// certificate, or none at all on a private network.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/pumasi-ai/pumasi-tunnel/relay"
)

func main() {
	var (
		baseDomain = flag.String("domain", "pumasi.link", "base domain tunnels are published under")
		agentAddr  = flag.String("agent-addr", ":7000", "address to accept agent connections on")
		httpAddr   = flag.String("http-addr", ":8000", "address to accept visitor HTTP on")
		tcpLow     = flag.Int("tcp-low", 0, "lowest public port for raw TCP tunnels (0 disables TCP)")
		tcpHigh    = flag.Int("tcp-high", 0, "highest public port for raw TCP tunnels")
		tcpBind    = flag.String("tcp-bind", "", "interface to bind public TCP ports on (empty = all)")
		sshAddr    = flag.String("ssh-addr", "", "address for zero-install ssh tunnelling, e.g. :2222 (empty disables it)")
		hostKey    = flag.String("ssh-hostkey", "/var/lib/pumasi-relay/ssh_host_ed25519_key", "path to the relay's ssh host key; generated if absent")
		publicHost = flag.String("public-host", "", "hostname visitors dial for raw TCP (defaults to -domain)")
		verbose    = flag.Bool("v", false, "log every routing decision")
	)
	flag.Parse()

	level := slog.LevelInfo
	if *verbose {
		level = slog.LevelDebug
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	r, err := relay.New(relay.Config{
		BaseDomain:      *baseDomain,
		Logger:          log,
		TCPPortLow:      *tcpLow,
		TCPPortHigh:     *tcpHigh,
		TCPBindHost:     *tcpBind,
		PublicHost:      *publicHost,
		AgentPublicPort: *agentAddr,
	})
	if err != nil {
		log.Error("could not start", "error", err)
		os.Exit(1)
	}

	agentLn, err := net.Listen("tcp", *agentAddr)
	if err != nil {
		log.Error("could not listen for agents", "addr", *agentAddr, "error", err)
		os.Exit(1)
	}
	defer agentLn.Close()

	go func() {
		for {
			conn, err := agentLn.Accept()
			if err != nil {
				// A closed listener is the shutdown path, not a fault.
				if errors.Is(err, net.ErrClosed) {
					return
				}
				log.Warn("accept failed", "error", err)
				continue
			}
			go r.ServeAgent(conn)
		}
	}()

	// Zero-install path: a stock ssh client can open a tunnel with no binary
	// downloaded and no account, which the incumbent study found to be the
	// highest-leverage onboarding behaviour in this category.
	if *sshAddr != "" {
		signer, err := relay.LoadOrCreateHostKey(*hostKey, relay.GenerateHostKeyPEM, os.ReadFile, func(path string, pem []byte) error {
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				return err
			}
			// 0600: a leaked host key lets anyone impersonate this relay to
			// every client that has it in known_hosts.
			return os.WriteFile(path, pem, 0o600)
		})
		if err != nil {
			log.Error("ssh ingress disabled", "error", err)
			os.Exit(1)
		}
		sshLn, err := net.Listen("tcp", *sshAddr)
		if err != nil {
			log.Error("could not listen for ssh", "addr", *sshAddr, "error", err)
			os.Exit(1)
		}
		defer sshLn.Close()
		go func() {
			log.Info("ssh ingress listening", "addr", *sshAddr)
			for {
				conn, err := sshLn.Accept()
				if err != nil {
					if errors.Is(err, net.ErrClosed) {
						return
					}
					log.Warn("ssh accept failed", "error", err)
					continue
				}
				go r.ServeSSH(conn, signer)
			}
		}()
	}

	srv := &http.Server{
		Addr:              *httpAddr,
		Handler:           r,
		ReadHeaderTimeout: 15 * time.Second,
	}

	go func() {
		log.Info("relay listening",
			"domain", *baseDomain, "agents", *agentAddr, "visitors", *httpAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("visitor listener failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	fmt.Fprintln(os.Stderr)
	log.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

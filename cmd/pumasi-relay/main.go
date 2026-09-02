// Command pumasi-relay is the public side of a tunnel. It listens for agents
// on one port and for visitors on another, and forwards between them.
//
// TLS is deliberately not terminated here: run this behind a TLS listener or
// a reverse proxy that holds the wildcard certificate. Keeping certificate
// handling out of the relay means an operator can choose ACME, a purchased
// certificate, or none at all on a private network.
//
// Because that choice is made outside this process, the relay cannot observe
// it — so -public-scheme is how it is told. It defaults to http, which is what
// this binary serves on its own; an operator who put a terminator in front
// passes -public-scheme=https, and every address the relay announces says so
// at once. Announcing https with nothing listening on 443 is the one thing
// this flag exists to stop.
package main

import (
	"context"
	"crypto/tls"
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

	"github.com/pumasi-ai/pumasi-tunnel/internal/buildinfo"
	"github.com/pumasi-ai/pumasi-tunnel/relay"
)

func main() {
	var (
		baseDomain     = flag.String("domain", "pumasi.link", "base domain tunnels are published under")
		agentAddr      = flag.String("agent-addr", ":7000", "address to accept agent connections on")
		agentTLSAddr   = flag.String("agent-tls-addr", "", "address to accept TLS agent connections on; empty disables it")
		httpAddr       = flag.String("http-addr", ":8000", "address to accept visitor HTTP on")
		httpsAddr      = flag.String("https-addr", "", "address to accept visitor HTTPS on; empty disables it")
		tlsCert        = flag.String("tls-cert", "", "PEM certificate for visitor and agent TLS")
		tlsKey         = flag.String("tls-key", "", "PEM private key for visitor and agent TLS")
		tcpLow         = flag.Int("tcp-low", 0, "lowest public port for raw TCP tunnels (0 disables TCP)")
		tcpHigh        = flag.Int("tcp-high", 0, "highest public port for raw TCP tunnels")
		tcpBind        = flag.String("tcp-bind", "", "interface to bind public TCP ports on (empty = all)")
		sshAddr        = flag.String("ssh-addr", "", "address for zero-install ssh tunnelling, e.g. :2222 (empty disables it)")
		hostKey        = flag.String("ssh-hostkey", "/var/lib/pumasi-relay/ssh_host_ed25519_key", "path to the relay's ssh host key; generated if absent")
		publicHost     = flag.String("public-host", "", "hostname visitors dial for raw TCP (defaults to -domain)")
		pubScheme      = flag.String("public-scheme", "http", "scheme tunnel addresses are announced under: http, or https when a TLS terminator sits in front of this relay")
		resvPath       = flag.String("reservations", "", "file to keep name and port reservations in, so a claim outlives a restart; empty keeps them in memory as before")
		maxTunnelsIP   = flag.Int("max-tunnels-per-ip", 20, "maximum simultaneous tunnels from one source IP")
		startsMinute   = flag.Int("tunnel-starts-per-minute", 60, "maximum tunnel connection attempts per source IP per minute")
		maxTunnelConns = flag.Int("max-connections-per-tunnel", 64, "maximum simultaneous visitor connections per tunnel")
		verbose        = flag.Bool("v", false, "log every routing decision")
		showVersion    = flag.Bool("version", false, "print the Pumasi Tunnel version and exit")
	)
	flag.Parse()
	if *showVersion {
		fmt.Println("pumasi-relay " + buildinfo.Version)
		return
	}

	level := slog.LevelInfo
	if *verbose {
		level = slog.LevelDebug
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	r, err := relay.New(relay.Config{
		BaseDomain:   *baseDomain,
		Logger:       log,
		TCPPortLow:   *tcpLow,
		TCPPortHigh:  *tcpHigh,
		TCPBindHost:  *tcpBind,
		PublicHost:   *publicHost,
		PublicScheme: *pubScheme,
		AgentPublicPort: func() string {
			if *agentTLSAddr != "" {
				return *agentTLSAddr
			}
			return *agentAddr
		}(),
		SSHPublicPort:           *sshAddr,
		MaxTunnelsPerIP:         *maxTunnelsIP,
		TunnelStartsPerMinute:   *startsMinute,
		MaxConnectionsPerTunnel: *maxTunnelConns,
		// Empty is this relay exactly as every release before it: no file, no
		// lock, no sweep and no write. With a path, a claimed name and its
		// public port survive a restart of this process — the connection
		// never does, and no surface here says otherwise (spec/0004 §4).
		ReservationsPath:    *resvPath,
		FeedbackGitHubToken: os.Getenv("PUMASI_GITHUB_TOKEN"),
		FeedbackGitHubRepo:  os.Getenv("PUMASI_GITHUB_REPO"),
	})
	if err != nil {
		log.Error("could not start", "error", err)
		os.Exit(1)
	}
	defer r.Close()

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

	if (*tlsCert == "") != (*tlsKey == "") {
		log.Error("TLS needs both -tls-cert and -tls-key")
		os.Exit(1)
	}
	var tlsConfig *tls.Config
	if *tlsCert != "" {
		cert, err := tls.LoadX509KeyPair(*tlsCert, *tlsKey)
		if err != nil {
			log.Error("could not load TLS certificate", "error", err)
			os.Exit(1)
		}
		tlsConfig = &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12}
	}
	if *agentTLSAddr != "" {
		if tlsConfig == nil {
			log.Error("-agent-tls-addr requires -tls-cert and -tls-key")
			os.Exit(1)
		}
		ln, err := tls.Listen("tcp", *agentTLSAddr, tlsConfig)
		if err != nil {
			log.Error("could not listen for TLS agents", "addr", *agentTLSAddr, "error", err)
			os.Exit(1)
		}
		defer ln.Close()
		go serveAgents(ln, r, log)
		log.Info("TLS agent ingress listening", "addr", *agentTLSAddr)
	}

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

	var visitorHandler http.Handler = r
	if *httpsAddr != "" {
		visitorHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			target := "https://" + req.Host + req.URL.RequestURI()
			http.Redirect(w, req, target, http.StatusPermanentRedirect)
		})
	}
	srv := &http.Server{
		Addr:              *httpAddr,
		Handler:           visitorHandler,
		ReadHeaderTimeout: 15 * time.Second,
	}
	var httpsSrv *http.Server
	if *httpsAddr != "" {
		if tlsConfig == nil {
			log.Error("-https-addr requires -tls-cert and -tls-key")
			os.Exit(1)
		}
		httpsSrv = &http.Server{Addr: *httpsAddr, Handler: r, ReadHeaderTimeout: 15 * time.Second, TLSConfig: tlsConfig}
		go func() {
			log.Info("HTTPS relay listening", "addr", *httpsAddr)
			if err := httpsSrv.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Error("HTTPS visitor listener failed", "error", err)
				os.Exit(1)
			}
		}()
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
	if httpsSrv != nil {
		httpsSrv.Shutdown(ctx)
	}
}

func serveAgents(ln net.Listener, r *relay.Relay, log *slog.Logger) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Warn("agent accept failed", "error", err)
			continue
		}
		go r.ServeAgent(conn)
	}
}

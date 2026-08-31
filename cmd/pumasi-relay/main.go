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
		BaseDomain:  *baseDomain,
		Logger:      log,
		TCPPortLow:  *tcpLow,
		TCPPortHigh: *tcpHigh,
		TCPBindHost: *tcpBind,
		PublicHost:  *publicHost,
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

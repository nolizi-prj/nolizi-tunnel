// Command pumasi-tunnel publishes a port on this machine at a public address,
// using only an outbound connection — no port forwarding, no inbound firewall
// rule.
//
//	pumasi-tunnel --relay relay.example:7000 8080
//	pumasi-tunnel --relay relay.example:7000 --subdomain myapi 8080
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/pumasi-ai/pumasi-tunnel/agent"
	"github.com/pumasi-ai/pumasi-tunnel/core"
)

func main() {
	var (
		relayAddr = flag.String("relay", "", "relay control address, host:port (required)")
		subdomain = flag.String("subdomain", "", "request a specific name; omit to be assigned one")
		token     = flag.String("token", "", "token for a reserved subdomain")
		host      = flag.String("host", "127.0.0.1", "local host to forward to")
		tcp       = flag.Bool("tcp", false, "raw TCP tunnel (SSH, RDP, databases) instead of HTTP")
		verbose   = flag.Bool("v", false, "log each forwarded stream")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: pumasi-tunnel --relay host:port [flags] <local-port>\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *relayAddr == "" || flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	port, err := strconv.Atoi(flag.Arg(0))
	if err != nil || port < 1 || port > 65535 {
		fmt.Fprintf(os.Stderr, "not a port number: %q\n", flag.Arg(0))
		os.Exit(2)
	}

	// A requested name is normalised here rather than in the relay, so a
	// person who types "MyApi" is told what happened instead of silently
	// getting something else.
	requested := strings.ToLower(strings.TrimSpace(*subdomain))
	if requested != *subdomain && *subdomain != "" {
		fmt.Fprintf(os.Stderr, "using subdomain %q (names are lowercase)\n", requested)
	}
	if requested != "" {
		if err := core.ValidateSubdomain(requested); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	}

	level := slog.LevelWarn
	if *verbose {
		level = slog.LevelInfo
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	local := fmt.Sprintf("%s:%d", *host, port)
	ag, err := agent.New(agent.Config{
		RelayAddr: *relayAddr,
		LocalAddr: local,
		Subdomain: requested,
		Token:     *token,
		TCP:       *tcp,
		Logger:    log,
		OnConnect: func(resp core.AuthResponse) {
			// Printed on every connect, reconnects included, because after a
			// network flap the first thing a person wants to know is whether
			// the address still holds.
			addr := resp.URL
			if resp.TCPAddr != "" {
				addr = resp.TCPAddr
			}
			fmt.Printf("\n  %s  ->  %s\n\n", addr, local)
		},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		fmt.Fprintf(os.Stderr, "\nclosing the tunnel (%d requests served)\n", ag.Requests())
		cancel()
	}()

	if err := ag.Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

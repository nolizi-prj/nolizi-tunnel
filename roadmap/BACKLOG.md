# BACKLOG — Pumasi Tunnel

**Owned by the Product Manager role** ([`pumasi-ops/roles/product-manager.md`](https://github.com/pumasi-ai/pumasi-ops/blob/main/roles/product-manager.md)).
Seeded 2026-08-30 from candidate `0011-developer-tunnels.md`.

---

## The Prioritized Backlog

1. **Pure Core Stream Multiplexer (`core/mux.go`)**:
   - Deterministic framing and byte proxying between TCP connections and multiplexed virtual streams (`yamux` / `quic-go`).
   - Unit test suite covering packet loss, reconnects, half-close handling, and throughput benchmarks.

2. **SSH Ingress Gateway (`service/ssh_server.go`)**:
   - Embedded SSH server on ports 22/443 (`gliderlabs/ssh`).
   - Handles `ssh -R <remote_port>:localhost:<local_port> tunnel.pumasi.ai` with zero client installation.
   - Generates and returns ephemeral public subdomain URLs in the SSH banner.

3. **HTTP/HTTPS Wildcard Host Router (`service/http_router.go`)**:
   - Dynamic subdomain matching (`*.pumasi.link`).
   - Auto-provisioned Let's Encrypt / ACME TLS certificates for wildcard domains.
   - Forwards raw HTTP headers and request bodies directly without interstitial HTML pages.

4. **Raw TCP Port Pool Manager (`service/tcp_allocator.go`)**:
   - Allocates dedicated public TCP ports for non-HTTP services (Windows Remote Desktop RDP 3389, databases 5432).
   - Direct byte-for-byte forwarding to client multiplexed connections.

5. **Client CLI (`cmd/pumasi-tunnel/main.go`)**:
   - Single static binary for Linux, macOS, and Windows.
   - Interactive terminal UI showing live requests, response times, and status codes.
   - Flags for `--subdomain`, `--http`, `--tcp`, and `--auth`.

6. **Local Request Inspector & Webhook Replay (`web/inspector`)**:
   - Embedded web UI on `http://127.0.0.1:4040`.
   - Real-time SSE / WebSocket streaming of incoming requests.
   - JSON formatting, header inspection, and "Replay Request" button.

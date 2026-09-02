# Pumasi Tunnel

Publish localhost through one outbound connection. Pumasi Tunnel supports
HTTPS applications and raw TCP, using either stock `ssh` or one static Go
client. No account or session timer is required. Apache-2.0.

Live relay: [pumasi.link](https://pumasi.link) · current version: `0.1.5`

## Quickstart

### Stock SSH — no install

```bash
# Publish localhost:8080 at an assigned HTTPS address
ssh -p 2222 -R 0:localhost:8080 pumasi.link

# Request myapi.pumasi.link
ssh -p 2222 -R 0:localhost:8080 myapi@pumasi.link

# Publish localhost:22 at an assigned raw TCP port
ssh -p 2222 -R 0:localhost:22 tcp@pumasi.link
```

Keep the SSH session open. The relay prints the public address.

### Native client

```bash
curl -fsSL https://pumasi.link/install.sh | sh

# Secure TLS control connection; assigned HTTPS address
pumasi-tunnel --relay pumasi.link:7001 8080

# Requested hostname
pumasi-tunnel --relay pumasi.link:7001 --subdomain myapi 8080

# Raw TCP and a fixed public port
pumasi-tunnel --relay pumasi.link:7001 --tcp --tcp-port 20050 22
```

The installer supports Linux and macOS on amd64 and arm64 and verifies the
published SHA-256 checksum. Windows binaries and source archives are available
from [GitHub Releases](https://github.com/pumasi-ai/pumasi-tunnel/releases).

The native client verifies TLS by default. `--insecure` exists only for a
plaintext self-hosted development relay. A token of at least 16 characters can
claim a requested hostname or TCP port and reclaim it after disconnects and
relay restarts:

```bash
./pumasi-tunnel --relay pumasi.link:7001 \
  --subdomain myapi --token 'replace-with-a-long-secret' 8080
```

Stock SSH cannot carry a reservation token; use the native client for owned
addresses.

## What Phase 1 includes

- HTTP host tunnels with automatic wildcard HTTPS
- raw TCP tunnels from public ports `20000–20099`
- encrypted native-agent control on `7001`
- zero-install SSH ingress on `2222`
- assigned, requested, and token-reserved addresses
- persistent reservations
- live console, guided command builder, health/readiness/version endpoints
- feedback that creates a GitHub issue without exposing its credential

The concise Phase 2 and Phase 3 plan is in
[`roadmap/THREE_PHASE_DEVELOPMENT_PLAN.md`](roadmap/THREE_PHASE_DEVELOPMENT_PLAN.md).

## Self-hosting

For plaintext local development:

```bash
go build -o pumasi-relay ./cmd/pumasi-relay
./pumasi-relay -domain localhost -http-addr :8000
./pumasi-tunnel --insecure --relay localhost:7000 8080
```

For an internet relay, configure a certificate containing the apex and
wildcard names, then enable HTTPS and TLS agent ingress:

```bash
./pumasi-relay \
  -domain example.com -public-host example.com -public-scheme https \
  -agent-addr 127.0.0.1:7000 -agent-tls-addr :7001 \
  -http-addr :80 -https-addr :443 \
  -tls-cert /secure/example.com.crt -tls-key /secure/example.com.key \
  -ssh-addr :2222 -tcp-low 20000 -tcp-high 20099 \
  -reservations /var/lib/pumasi-relay/reservations.json
```

The repository’s hardened systemd unit is
[`deploy/pumasi-relay.service`](deploy/pumasi-relay.service).

## Operations

```bash
curl -fsS https://pumasi.link/version
curl -fsS https://pumasi.link/healthz
curl -fsS https://pumasi.link/readyz

go test ./...
go vet ./...
go test -race ./...
```

The production relay is a Vultr host. Its HTTPS certificate is a Let’s Encrypt
apex + wildcard certificate issued with DNS-01. Do not place Cloudflare or
GitHub credentials in this repository; deployment reads them from the Pumasi
encrypted secret store or a root-only service environment file.

## Architecture

```text
browser / webhook ─HTTPS─┐
TCP client ──────────────┼─> relay ─one multiplexed TLS connection─> agent ─> localhost
stock ssh -R ────────────┘
```

- `core/`: wire frames, routing, allocation, reservations
- `mux/`: multiple logical streams over one connection
- `relay/`: HTTP/TCP routing, SSH ingress, console, operational endpoints
- `agent/`: outbound client and local forwarding
- `cmd/`: relay and client binaries

## Known limits

- One relay in one region; no high-availability claim yet.
- No request inspector/replay yet (Phase 2).
- No accounts, teams, access policies, custom domains, SDKs, or Kubernetes
  integration yet (Phases 2–3).
- One stalled stream can currently delay sibling streams; per-stream credit
  windows are later work.

## Research basis

The product and phased plan were informed by public documentation for
[ngrok](https://ngrok.com/docs), [frp](https://gofrp.org/en/docs/),
[zrok](https://docs.zrok.io/), and the locally reviewed open-source tunnel
projects. The UI is an original implementation; research was used to understand
workflows, failure cases, and technical tradeoffs.

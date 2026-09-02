# Pumasi Tunnel — three-phase plan

**Production:** `pumasi.link` on the Vultr relay.
**Rule:** a feature is available only after it is tested and verified there.

## Evidence

- Leading product: ngrok screenshots in `pumasi-site-screenshot/ngrok_pages/`
  and the official ngrok CLI, API, endpoint, traffic inspection, and security
  documentation.
- Open-source references: official docs and local source for frp, zrok,
  rathole, Pangolin, and cloudflared in `tunnel_sites/`.
- Adopt: guided quickstart, copyable commands, honest endpoint state, secure
  defaults, explicit errors, responsive console, bounded data capture.
- Retain: no-account stock-SSH path, static Go client, free raw TCP, and
  self-hostable relay.
- Defer: accounts, teams, billing, Kubernetes, policy engines, and hosted
  traffic retention.

## Phase 1 — trustworthy first tunnel — complete 2026-09-02 (`0.1.0`)

- HTTPS for the apex and wildcard tunnel hosts.
- TLS for the native client-to-relay connection; SSH stays encrypted by SSH.
- Deploy persistent reservations and honest URL/bind ordering already on main.
- Guided HTTP/TCP quickstart, copy controls, endpoint state, useful empty and
  failure states, responsive keyboard-accessible UI.
- Version in CLI, console, `/version`, `/healthz`, and `/readyz`.
- Feedback button with server-side GitHub delivery, rate limits, no browser
  credential, and optional contact accepted only over HTTPS.
- Protocol, persistence, TLS, HTTP, and real-browser tests.
- Production smoke test: open a local HTTP tunnel, reach it over HTTPS, and
  verify raw TCP still works.

**Exit:** tests pass; production reports the release; HTTPS and native TLS are
verified; real HTTP and TCP tunnels work; no release-blocking issue is open.

**Evidence:** exit criteria met on `pumasi.link`; details are recorded in
[`STAGE.md`](STAGE.md).

## Phase 2 — developer debugging and ownership

- Local traffic inspector on `127.0.0.1:4040`, metadata by default, opt-in
  bounded body capture, redaction, filters, and safe replay.
- Local agent status UI, config file, environment token, release artifacts,
  diagnostics, and update checks.
- Account-backed named endpoints, token rotation, endpoint management, audit
  events, and private per-user views.
- HTTP/OIDC access protection, IP policy, header rewriting, and limits.

## Phase 3 — teams and resilient infrastructure

- Organizations, roles, service accounts, API, and automation.
- Multiple relays/regions, draining, metrics, alerts, backups, and tested
  disaster recovery.
- Custom domains, automated certificates, private networking, UDP, SDKs,
  Kubernetes, Terraform, and enterprise identity based on measured demand.

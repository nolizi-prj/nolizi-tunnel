# STAGE — Pumasi Tunnel

**Current Stage:** `Alpha`
**Stage 1 exit gate:** **MET 2026-08-31** (evidence below).
**Stage 2 (Beta) work:** in progress. The `beta` *label* is not claimed yet, and
what holds it back is listed under "Why not `beta`".
**Selected Date:** 2026-08-30
**Steward Directive:** selected as 3rd ecosystem product for immediate remote
access dogfooding, zero-cost developer top-of-funnel distribution, and
multi-agent pipeline calibration.

Owned by the product-manager role
([`pumasi-ops/roles/product-manager.md`](https://github.com/pumasi-ai/pumasi-ops/blob/main/roles/product-manager.md),
duty 6); the stage ladder and its meanings are that file's table, and the
stage-by-stage gates are
[`pumasi-ops/STAGE_PLAYBOOK.md`](https://github.com/pumasi-ai/pumasi-ops/blob/main/STAGE_PLAYBOOK.md).
Neither is restated here (L-007).

---

## Maturity Gates

| Stage | Criteria | Status |
| :--- | :--- | :--- |
| **0. Candidate** | Scored by 3 model families (40/60) and selected by steward | **COMPLETE** |
| **1. Alpha** | Pure-core suite passes 100%; both public landing surfaces live | **COMPLETE 2026-08-31** |
| **2. Beta** | Real end-to-end users complete workflows without engineer intervention | **IN PROGRESS** |
| **3. Launched** | Production hardening, cross-model regression, 7-day veto window | PENDING |

---

## The Stage 1 gate, measured

Measured 2026-08-31 at `3652e15` by the product-manager evaluation, by running
the commands rather than reading the claim.

**Pure core passes.** `go test -count=1 ./...` exits 0: `ok core`, `ok mux`,
`ok relay`. Coverage of the three: 79.5%, 83.5%, 55.9%.

**Both surfaces are live.**
- *Surface A, the commons catalog*: `pumasi-web`
  `content/products/pumasi-tunnel.md` (`c2084a8`, 2026-08-30);
  `https://pumasi.ai/products/pumasi-tunnel/` → 200, and the entry appears in
  `https://pumasi.ai/llms.txt`.
- *Surface B, the product's own domain*: `http://pumasi.link/` → 200, serving
  the console (`relay/dashboard.html`, `b3585f6`).

**And the product carries real traffic.** `http://pumasi.link/_pumasi/status`
at 14:08 UTC reported one tunnel open: `sshsteward`, `pumasi.link:20000` →
local port 22, opened 06:18:13Z, **28 240 s** (7 h 50 m) and unbroken. A banner
grab on `s.pumasi.link:20000` returns `SSH-2.0-OpenSSH_10.2p1 Ubuntu-2ubuntu3.5`
— this machine's own sshd, reached across the tunnel. That is
`RESOURCES.md` §4's remote access path, working, and it is the strongest
evidence this product has.

### Two qualifications recorded with the pass, not against it

Both are `BACKLOG.md` entries, not gate failures — but the gate's number is
only as good as what it covers (L-006), so neither is left implicit.

1. **The suite is not deterministic under `-cover`.** `go test -count=1 -cover
   ./...` failed **2 of 12** runs, always the same way:
   `TestRawTCPCrossesTheTunnel … dial tcp 127.0.0.1:34000: connect: connection
   refused`. It is not a test artifact. `relay/relay.go` writes the auth
   response carrying `TCPAddr` (line ~162) *before* `bindTCP` (line ~181), so
   an agent is handed its public address before anything listens on it, and a
   bind failure is reported only after the address has been announced.
   Instrumentation widens a window that exists in the product.
2. **Three packages have no test files, `agent/` among them.** `agent`,
   `cmd/pumasi-relay`, `cmd/pumasi-tunnel` report *no test files*. `core`,
   `mux` and `relay` are the pure core the gate names and they are covered;
   `agent/` is not core, but it is half of every tunnel and today nothing
   exercises it except the relay's end-to-end tests using it as a fixture.

---

## Why not `beta`

`beta` means strangers can rely on it and their data survives. Three verified
facts say not yet. Each is a backlog entry; the order below is the backlog's.

1. **Every tunnel is handed an address that does not answer.**
   `Registry.PublicURL` (`core/route.go:255`) returns `https://<name>.pumasi.link`
   unconditionally, the console prints it, and **nothing listens on 443**:
   `curl https://sshsteward.pumasi.link` → *Could not connect to server*,
   port 443 refused. TLS termination is deliberately outside the relay
   (`cmd/pumasi-relay/main.go` header) and outside it there is currently
   nothing.
2. **A name belongs to nobody.** `Tunnel.Reserved` is computed at
   `relay/relay.go:236` and **never read anywhere in the tree**; the relay
   binary exposes no auth flag, so `AllowAll` (`relay/relay.go:40`) is the only
   authenticator it can run. Any anonymous agent may take any free name.
3. **Nothing survives a restart.** The registry and the port pool are in
   memory. A relay restart drops every subdomain, every reserved TCP port, and
   every live tunnel — including the one carrying this machine's remote access.
   `--tcp-port` keeps an address across an *agent* reconnect (`a5b77fc`), not
   across a relay one.

Also gating, from `PRODUCT-RULES.md` (v1.0, read fresh 2026-08-31; still only on
the unmerged `worktree-product-rules` branch, `0115758` — its absence from
`main` is not compliance): **PR-1** (a version number — this product has none
anywhere: no version file, no `/version`, nothing on the console) binds always;
**PR-2** (in-app feedback) binds at the `beta` promotion and is unbuilt. Both
are backlog entries.

## What `launched` additionally requires

Stage 2's exit gate — real end-to-end users completing workflows without an
engineer — plus production hardening. Not enumerated further here while the
`beta` list above is open.

## Known gaps a user should know about today

- No TLS. Every tunnel is plain HTTP, whatever URL the relay printed.
- No accounts, no tokens in force, no name ownership.
- One relay, one host (`RESOURCES.md` §3: Vultr, Chicago, ~$5–6/month). A
  restart or a host failure ends every tunnel. Tailscale is kept as the
  independent fallback for reaching `m-gtr`, deliberately.
- The local request inspector on `127.0.0.1:4040` **does not exist** — `web/`
  is an empty directory. `VALUE.md` no longer claims it, and the commons
  catalog page already disclaims it correctly (`pumasi-web` `843bdef`).
- No client TUI: `cmd/pumasi-tunnel` is flags and logs.

## For the marketing manager, from this evaluation

The commons catalog page quotes the gate table above *verbatim* and was
correct against the version it read. Two things it now needs, and neither is
this seat's to write: the table it quotes has changed, and its lead sentence
**"There is no hosted relay"** is false — `pumasi.link` resolves to
`64.177.118.159`, serves the console on 80 and a zero-install ssh ingress on
2222, and has been carrying a tunnel for nearly eight hours. `tunnel.pumasi.ai`
indeed does not resolve; the hosted relay is on the other domain.

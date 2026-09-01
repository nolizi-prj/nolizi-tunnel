# Pumasi Tunnel

**Localhost tunnels over a single outbound connection — from a stock `ssh`
client or one static Go binary. Raw TCP is an ordinary capability here, not a
paid one. No account, no session timer, Apache-2.0.**

[![Stage: Alpha](https://img.shields.io/badge/stage-ALPHA%20%E2%80%94%20active%20development-orange)](roadmap/STAGE.md)

**Stage: `Alpha`** — see [`roadmap/STAGE.md`](roadmap/STAGE.md), which owns the
stage and the evidence behind it. Read the limitations in
[What is not true yet](#what-is-not-true-yet) before you rely on this. The
shortest version: **nothing here serves HTTPS**, and the relay process running
at `pumasi.link` is an **older build than `main`** — several merged fixes have
not been deployed.

Part of [Pumasi](https://github.com/pumasi-ai/pumasi), a commons of software
built by agents and governed by people. Apache-2.0, inbound equals outbound.

---

## Quickstart

### Method 1 — nothing installed, stock `ssh`

The relay speaks enough of the SSH connection protocol to accept a reverse
forward ([`relay/sshingress.go`](relay/sshingress.go)). The ingress is on
**port 2222** — port 22 on that host is its own sshd — so `-p 2222` is part of
every command.

```bash
# Expose local port 8080 over HTTP; the relay assigns a name
ssh -p 2222 -R 0:localhost:8080 pumasi.link

# Expose local port 22 as a raw TCP port (SSH, RDP, databases)
ssh -p 2222 -R 0:localhost:22 tcp@pumasi.link

# Ask for a specific name
ssh -p 2222 -R 0:localhost:8080 myapi@pumasi.link
```

Options ride in the SSH username because it is the only field a plain `ssh`
client lets you set without configuration. They combine with `+`, e.g.
`myapi+tcp@pumasi.link` ([`relay/sshingress.go:119`](relay/sshingress.go)).
Your address is printed into your terminal; no account, no key registration,
no download.

> **Read the address you are given carefully.** The relay currently running at
> `pumasi.link` predates the scheme fix on `main` and still prints an
> `https://` URL. Nothing listens on 443. Use `http://` on the same hostname.
> This is [`roadmap/BACKLOG.md`](roadmap/BACKLOG.md) item 1 and
> [`roadmap/STAGE.md`](roadmap/STAGE.md) §1.

### Method 2 — the CLI

```bash
go build ./cmd/pumasi-tunnel

# HTTP tunnel for a local web app on 8080
./pumasi-tunnel --relay pumasi.link:7000 8080

# Raw TCP instead of HTTP
./pumasi-tunnel --relay pumasi.link:7000 --tcp 3389

# Ask for a name, and for a public port that survives reconnects
./pumasi-tunnel --relay pumasi.link:7000 --subdomain myapi 8080
./pumasi-tunnel --relay pumasi.link:7000 --tcp --tcp-port 20050 22
```

`--relay` is **required** and the local port is a positional argument; without
both, the program prints its usage and exits 2
([`cmd/pumasi-tunnel/main.go`](cmd/pumasi-tunnel/main.go)).

| Flag | Default | What it does |
| :--- | :--- | :--- |
| `--relay` | *(none — required)* | relay control address, `host:port` |
| `--subdomain` | *(assigned)* | request a specific name; lowercase, letters/digits/inner hyphens, and not on the reserved list ([`core/subdomain.go`](core/subdomain.go)) |
| `--token` | *(empty)* | claims `--subdomain` for whoever holds this token, so nobody else may use that name — **16 characters minimum**; the name is held while you are disconnected but **not across a relay restart**, see below |
| `--host` | `127.0.0.1` | local host to forward to |
| `--tcp` | `false` | raw TCP tunnel instead of HTTP |
| `--tcp-port` | `0` | request this exact public port, so the address survives reconnects |
| `-v` | `false` | log each forwarded stream |

---

## Architecture

No QUIC, no yamux, no smux, no third-party SSH server. The entire module
depends on the Go standard library plus `golang.org/x/crypto`
([`go.mod`](go.mod)) — that is the whole dependency list.

```
  visitor: browser · webhook sender · mstsc · psql · ssh
        │                                    │
        │ http://<name>.pumasi.link:80       │ pumasi.link:<20000-20099>
        ▼                                    ▼
┌───────────────────────────────────────────────────────────────┐
│  RELAY  —  relay/ , one process, TLS deliberately outside it  │
│                                                               │
│  · HTTP host router          relay/relay.go   ServeHTTP       │
│  · raw TCP port pool         relay/tcp.go   + core/portpool   │
│  · ssh ingress, port 2222    relay/sshingress.go              │
│      golang.org/x/crypto/ssh, RFC 4254 §7 reverse forwarding  │
│  · console + status JSON     relay/dashboard.go               │
└───────────────┬───────────────────────────┬───────────────────┘
                │ agent control conn         │ ssh transport
                │ plain TCP, port 7000       │ encrypted by ssh
                │ (NOT TLS — see below)      │
                ▼                            ▼
┌───────────────────────────────────────────────────────────────┐
│  mux/  —  many logical streams over that one connection       │
│  · frame protocol from core/: OPEN DATA CLOSE PING PONG       │
│                               AUTH AUTH_OK ERROR              │
│  · stream ids partitioned odd/even so both ends may open      │
│  · backpressure is per-stream buffering, not a credit window  │
└───────────────────────────────┬───────────────────────────────┘
                                ▼
┌───────────────────────────────────────────────────────────────┐
│  AGENT  —  agent/ , cmd/pumasi-tunnel                         │
│  · dials outbound only; no inbound rule, no port forwarding   │
│  · forwards each stream to localhost:<port>                   │
└───────────────────────────────────────────────────────────────┘
```

**`core/` is pure.** Nothing in it opens a socket, reads a clock, or touches a
database — the wire protocol, the host-to-tunnel routing table, subdomain
allocation and the public TCP port pool are all deterministic in their inputs
([`core/frame.go`](core/frame.go)). `mux/` is the I/O shell that moves those
bytes. The relay and the CLI are thin shells around both.

**Backpressure is honest rather than clever.** A reader that stops reading
stalls its own stream and then the connection behind it. That bounds memory,
and it does let one stalled stream hold up its siblings; a credit window is the
fix and is deferred ([`mux/session.go`](mux/session.go)).

---

## Self-hosting the relay

The relay you would run is the relay running at `pumasi.link`, from this
repository ([`cmd/pumasi-relay/main.go`](cmd/pumasi-relay/main.go)).

```bash
go build ./cmd/pumasi-relay
./pumasi-relay --domain example.com --tcp-low 20000 --tcp-high 20099 --ssh-addr :2222
```

Three flags decide most of what the relay is:

- **`--ssh-addr`** defaults to **empty, which disables the ssh ingress
  entirely.** Whether zero-install `ssh -R` works is a property of how a given
  relay was started, not of this code.
- **`--tcp-low` / `--tcp-high`** default to `0`, which **disables raw TCP.**
  Choose a range *outside* your kernel's ephemeral range
  (`/proc/sys/net/ipv4/ip_local_port_range`, commonly `32768 60999`), or some
  unrelated process will already be holding a port you try to hand out. The
  deployed relay uses `20000–20099`, below that floor.
- **`--public-scheme`** defaults to `http`, which is what this binary serves on
  its own. Pass `https` **only** when you have actually put a TLS terminator in
  front of it. Announcing `https` with nothing on 443 is the one thing this
  flag exists to stop — and it is exactly what the deployed relay is doing
  today, because it predates the flag.

TLS is deliberately not terminated in the relay, so an operator can choose
ACME, a purchased certificate, or none at all on a private network.

---

## How it compares

**Every pricing and plan figure below is quoted from
[`roadmap/MARKET.md`](roadmap/MARKET.md)**, which read each vendor's own page on
**2026-08-31** and records the URL and the fetch date for every number. Prices
move; the date is part of the claim. The one row that is not from there — the
inspector — is from [`docs/ux/incumbent-ux-spec.md`](docs/ux/incumbent-ux-spec.md)
§3, a clean-room tour of all three products dated **2026-08-30**, and is marked.
A cell that cannot be sourced from one of those two is not in this table.

Comparators are the three this product was actually built against:
[ngrok](https://ngrok.com/pricing), [Pinggy](https://pinggy.io/#pricing),
[LocalXpose](https://localxpose.io/pricing).

| | ngrok | Pinggy | LocalXpose | **Pumasi Tunnel** |
| :--- | :--- | :--- | :--- | :--- |
| **Entry paid plan, as printed** | Hobbyist $10/mo; Pay-as-you-go $20/mo + usage | Pro — price not printed on the pricing page | PRO $8/mo, or $96/yr billed annually | **none — Apache-2.0** |
| **Client to start** | own agent binary | the OS `ssh` client | own binary (CLI + local GUI) | **stock `ssh`, or one static binary** |
| **Account to start** | yes | **no** (free tier) | yes | **no** |
| **Free-tier session ceiling** | none printed | 60 minutes | "Time limits" | **none** |
| **Raw TCP on the free path** | credit-card verification | included | not included | **included** |
| **Stable hostname** | ngrok-branded free; custom at $0.01/active hour on Pay-as-you-go | random free, custom on Pro | excluded from free Starter | **free, and unclaimed rather than owned — see below** |
| **Interstitial warning page** | none printed | none printed | printed on free Starter | **none** |
| **Self-hostable relay** | not offered publicly | on-premise on Enterprise | not offered publicly | **yes, same repository** |
| **HTTPS on your tunnel** | yes | yes | yes | **no — see below** |
| **Request inspector** *(source: UX tour §3)* | hosted, plus a local one | local web debugger on a fixed port, plus a terminal TUI | local GUI pane, per-tunnel toggle | **not built** |
| **Edge footprint** | vendor edge | vendor edge | vendor edge | **one host** |

**What that adds up to, stated only as wide as the citations allow.** An
untimed raw TCP tunnel with no account and no card on file is not offered by
any of the three on a free tier: Pinggy includes free TCP but times the session
out at 60 minutes, LocalXpose excludes TCP from free Starter, and ngrok's free
TCP wants card verification. That is the wedge, and it is about **price**, not
about reliability — see the next section for why that distinction is load-
bearing.

Nothing above is a claim about reliability, performance, support or security
posture. None of that was measured.

---

## What is not true yet

Kept in the README rather than in a footnote, because
[`roadmap/MARKET.md`](roadmap/MARKET.md) §4 records where this comparison goes
*against* us and a table that omits those is copy, not evidence.

- **There is no TLS.** `pumasi.link` has never listened on 443 and every HTTP
  tunnel here is plaintext. A webhook sender that requires an `https://`
  destination can be pointed at all three incumbents and at none of ours. The
  fix is an operator action nobody has taken ([`roadmap/BACKLOG.md`](roadmap/BACKLOG.md)
  item 1).
- **The CLI's control connection is plain TCP.** The agent dials the relay with
  `net.Dial` and nothing wraps it — there is no `crypto/tls` in the transport at
  all. The `ssh -R` path *is* encrypted, by SSH itself. The two methods on this
  page do not have the same security properties, and this page will not pretend
  they do.
- **A name can be owned, but only until the relay restarts.** `--token` now
  means something: it claims a `--subdomain`, and its public `--tcp-port` with
  it, on first use, and after that nobody without that token may have either —
  including in the gap between your reconnects, which is where a name used to
  be free for anyone to take
  ([`spec/0004-names-with-owners`](spec/0004-names-with-owners/SPEC.md)). Three
  limits, and none of them is small. **A relay restart still drops every claim
  at once**, because the reservation set is in memory — that is slice 2 of the
  same spec and it is not built. **A token is a bearer secret on a plaintext
  connection**, readable by anyone on the path, until the TLS gap above is
  closed. And **whoever claims a name first owns it**, so a stranger can still
  take a name you have never used — trust on first use is a strict improvement
  over *anyone, at any moment, forever*, and it is not a solution. `AllowAll` is
  still the only authenticator the relay can run: a token narrows *which name*
  you may have, never *whether you may connect*. The word *permanent* stays
  withdrawn until that restart column is filled
  ([`roadmap/BACKLOG.md`](roadmap/BACKLOG.md) item 2, *"A subdomain belongs to
  nobody, and nothing survives a relay restart"*).
- **The `ssh -R` path cannot hold a name.** Its username grammar has nowhere to
  put a token, so a zero-install tunnel can be *refused* a claimed name but can
  never claim or reclaim one. Every `ssh -R` tunnel is unreservable.
- **The running relay is behind `main`.** Merged fixes — the announced scheme,
  and the ordering of bind-before-announce — are on `main` and have not been
  deployed. A tunnel opened against `pumasi.link` today still gets the old
  behaviour. [`roadmap/STAGE.md`](roadmap/STAGE.md) §1 is entirely about keeping
  those two things apart.
- **One relay, one host.** Every vendor above runs an edge network. This runs a
  single small machine. Nothing on this page is an availability claim.
- **There is no request inspector.** No code in this tree binds port 4040, and
  `web/` is an empty directory. All three comparators ship one
  ([`docs/ux/incumbent-ux-spec.md`](docs/ux/incumbent-ux-spec.md) §3), so this
  is a gap and not a difference of opinion
  ([`roadmap/VALUE.md`](roadmap/VALUE.md) claim 5).

Two things did move recently, and both are narrower than they sound. The
announce-before-bind race on the TCP path is fixed and measured, and a flaky
test suite became deterministic on a machine where it previously failed every
run. **Neither is a claim that the product got more reliable**, neither is a
stage promotion, and — like everything else on `main` — neither is live. The
run counts, the arms they were measured on and the defect that is still open
are in [`roadmap/STAGE.md`](roadmap/STAGE.md) §2 and §7, which is the only
place they are stated.

---

## Repository map

| Path | What is in it |
| :--- | :--- |
| [`core/`](core) | pure protocol: frames, routing registry, subdomain rules, TCP port pool |
| [`mux/`](mux) | stream multiplexer over one connection |
| [`relay/`](relay) | the public side: HTTP router, TCP pool, ssh ingress, console |
| [`agent/`](agent) | the client side of the tunnel |
| [`cmd/`](cmd) | `pumasi-tunnel` (client) and `pumasi-relay` (server) |
| [`roadmap/`](roadmap) | [STAGE](roadmap/STAGE.md) · [VALUE](roadmap/VALUE.md) · [MARKET](roadmap/MARKET.md) · [BACKLOG](roadmap/BACKLOG.md) |
| [`spec/`](spec) | frozen specifications and their acceptance cases |
| [`reviews/`](reviews) | cross-model review transcripts, committed |

If this page and `roadmap/` ever disagree, `roadmap/` is right and this page is
the defect.

---

## License

Apache-2.0 — [`LICENSE`](LICENSE), [`NOTICE`](NOTICE). Inbound equals outbound.

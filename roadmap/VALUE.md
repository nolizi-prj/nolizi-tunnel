# VALUE — Pumasi Tunnel

**Owned by the product-manager role**
([`pumasi-ops/roles/product-manager.md`](https://github.com/pumasi-ai/pumasi-ops/blob/main/roles/product-manager.md),
duty 2). Seeded 2026-08-30; honest pass **2026-08-31** at `3652e15`, each claim
re-checked against the tree and the live relay. Every claim below carries what
would falsify it, and the two that are false today say so.

**Who it is for.** A developer who needs something on their own machine
reachable from the internet for a while — a webhook endpoint, a demo, an RDP or
SSH session into a machine behind NAT — and who does not want an account, a
binary, or a subscription in the way.

**The pain.** Reaching a machine you already own requires either router
configuration you may not control, or a hosted tunnel whose free tier makes the
useful parts — a stable name, a raw TCP port — the paid parts.

**Core proposition.** Multi-protocol localhost tunnels over one outbound
connection, from a stock `ssh` client or a single static binary, with raw TCP
as an ordinary capability rather than a paid one. Apache-2.0, and the relay is
the same repository — anyone may run their own.

*Competitor pricing and feature comparisons are deliberately absent from this
file: they were uncited, and the role permits no uncited claim about a
competitor. They belong in `MARKET.md`, with sources — `BACKLOG.md` item 10.*

---

## The claims, and what would falsify each

### 1 · Zero-client access from a stock `ssh` — **holds**

A tunnel opens with `ssh -R` from any machine that has an ssh client, with no
download and no account (`3652e15`, `relay/sshingress.go`).
*Evidence 2026-08-31:* `pumasi.link:2222` answers `SSH-2.0-pumasi-tunnel`.
*Falsified by:* an ordinary OpenSSH client that cannot open a tunnel and read
its public address from the banner; or an account, key registration or download
becoming necessary first.
*Qualification:* the ingress is on **2222**, not 22 — port 22 on that host is
its own sshd. A `-p 2222` is part of every command.

### 2 · A name you asked for, given back on reconnect — **holds, narrowly, and less than "permanent"**

`--subdomain myapi` is honoured when free, and `--tcp-port` asks for an exact
public port so the address survives an agent reconnect (`a5b77fc`).
*Falsified by:* a client that reconnects within seconds and is handed a
different name or port while the old one is free.
**What this is not:** ownership, and not persistence. `Tunnel.Reserved` is
computed at `relay/relay.go:236` and never read; the relay runs `AllowAll`;
the registry and pool are in memory. So another anonymous agent may take your
name in the gap between your reconnects, and a **relay restart drops every name
and reservation at once**. The seeded word "permanent" is withdrawn until
`BACKLOG.md` item 3 lands.

### 3 · Raw TCP, natively, for SSH and RDP and databases — **holds, and is the best-evidenced claim here**

A public TCP port forwards bytes with nothing parsed and no client-side helper
(`a13e586`, `relay/tcp.go`), including protocols where the server speaks first.
*Evidence 2026-08-31:* `pumasi.link:20000` has carried this machine's own sshd
for 7 h 50 m unbroken — a banner grab returns
`SSH-2.0-OpenSSH_10.2p1 Ubuntu-2ubuntu3.5` — and it is how `m-gtr` is reached
(`pumasi-ops/RESOURCES.md` §4).
*Falsified by:* a protocol that survives a direct connection but not this one;
or a session dropped by a timer rather than by one of its ends.

### 4 · No interstitial page, and no session timer — **holds for what it says; read the TLS gap below**

Nothing in the tree inserts an HTML warning page in front of a tunnel: a
visitor's request is forwarded and the response comes back. Nothing imposes a
session or bandwidth limit; the 7 h 50 m tunnel above is the evidence, not the
absence of a feature.
*Falsified by:* any code path returning a page the tunnelled service did not
produce, other than an error when the tunnel is genuinely unreachable; or a
disconnect at a fixed age.
**The gap that matters for webhooks:** there is **no TLS**. The relay hands out
`https://<name>.pumasi.link` (`core/route.go:255`) and nothing listens on 443 —
verified: `curl https://sshsteward.pumasi.link` cannot connect; only port 80
answers. Any sender that requires an `https://` destination therefore cannot be
pointed at a tunnel today, and every HTTP tunnel is plaintext.
`BACKLOG.md` item 1.

### 5 · A local request inspector on `127.0.0.1:4040` — **not built**

Claimed in the 2026-08-30 seed. It does not exist: `web/` is an empty
directory, no code binds 4040, and there is no capture, replay or SSE anywhere
in the tree. Recorded here as unbuilt rather than removed, because it is a real
intent — `BACKLOG.md` item 9 — and because a claim that quietly disappears is
how a value proposition starts lagging the product.
*Would hold when:* a request that crossed a tunnel can be read and replayed
from a local page, without an account.

### 6 · Self-hostable, Apache-2.0, no lock-in — **holds**

Two binaries, one Go module, no dependency outside the standard library and
`golang.org/x/crypto` (`go.mod`). The relay you would run is the relay running
at `pumasi.link`, from this repository.
*Falsified by:* a hosted-only capability — anything the public relay does that
a self-run one cannot, or a build step needing a credential this repository
does not contain.

---

## What a reader should take from the shape of this file

Three claims hold and are load-bearing today; one holds and is bounded by a
missing certificate; one is narrower than the word originally used; one is not
built. That distribution is the product's honest position at `Alpha`, and
`STAGE.md` says the same thing in the same words — if these two files ever
disagree, that disagreement is the defect.

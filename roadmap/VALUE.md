# VALUE — Pumasi Tunnel

**Owned by the product-manager role**
([`pumasi-ops/roles/product-manager.md`](https://github.com/pumasi-ai/pumasi-ops/blob/main/roles/product-manager.md),
duty 2). Seeded 2026-08-30; honest pass 2026-08-31 at `3652e15`;
post-release pass 2026-08-31 at `83fd9f7`; **post-`0041` pass 2026-08-31 at
`1d9505c`**, which lifted claim 3's bound (below) and re-read the live relay at
20:06 UTC. Each claim is re-checked against the tree and the live relay. Every
claim carries what would falsify it, and the ones that are false or bounded say
so.

**Who it is for.** A developer who needs something on their own machine
reachable from the internet for a while — a webhook endpoint, a demo, an RDP or
SSH session into a machine behind NAT — and who does not want an account, a
binary, or a subscription in the way.

**The pain, and it is now cited rather than asserted.** Reaching a machine you
already own requires either router configuration you may not control, or a
hosted tunnel that meters the useful parts. That second half used to be an
uncited claim about other people's pricing; it is now
[`MARKET.md`](MARKET.md) §1, read from three vendors' own pages on 2026-08-31:
a stable hostname is a paid capability at all three, raw TCP is behind a
payment or a card-on-file at two of them, and two of the three print a
free-tier session ceiling.

**Core proposition.** Multi-protocol localhost tunnels over one outbound
connection, from a stock `ssh` client or a single static binary, with raw TCP
as an ordinary capability rather than a paid one. Apache-2.0, and the relay is
the same repository — anyone may run their own.

**The comparison is bounded, on purpose.** [`MARKET.md`](MARKET.md) §4 records
the two places this product loses: it has no TLS where all three incumbents sell
HTTPS as the ordinary case, and its free stable name is *unclaimed* rather than
*owned*. Nothing in this file should be read past those.

---

## The claims, and what would falsify each

### 1 · Zero-client access from a stock `ssh` — **holds**

A tunnel opens with `ssh -R` from any machine that has an ssh client, with no
download and no account (`3652e15`, `relay/sshingress.go`).
*Evidence, re-read 2026-08-31 20:06 UTC:* `pumasi.link:2222` answers
`SSH-2.0-pumasi-tunnel`.
*Falsified by:* an ordinary OpenSSH client that cannot open a tunnel and read
its public address from the banner; or an account, key registration or download
becoming necessary first.
*Qualification:* the ingress is on **2222**, not 22 — port 22 on that host is
its own sshd. A `-p 2222` is part of every command.
*Fair comparison:* Pinggy also requires no account to start and is also
SSH-first ([`MARKET.md`](MARKET.md) §1–2). This claim is that we match it, not
that we are alone.
*Not yet true of the one page a visitor sees:* the console still offers only
`git clone && go build` and the binary's flags, never the `ssh -R` command —
[`BACKLOG.md`](BACKLOG.md) item 4.

### 2 · A name you asked for, given back on reconnect — **holds, narrowly, and less than "permanent"**

`--subdomain myapi` is honoured when free, and `--tcp-port` asks for an exact
public port so the address survives an agent reconnect (`a5b77fc`).
*Falsified by:* a client that reconnects within seconds and is handed a
different name or port while the old one is free.
**What this is not:** ownership, and not persistence. Re-verified at `83fd9f7`:
`Tunnel.Reserved` is computed at `relay/relay.go:249` and never read anywhere;
the relay binary has eleven flags and none of them is an auth flag, so it can
only run `AllowAll` (`relay/relay.go:40`); the registry
(`core/route.go:145–147`) and the port pool (`core/portpool.go:27–29`) are
in-memory maps with no persistence path. So another anonymous agent may take
your name in the gap between your reconnects, and a **relay restart drops every
name and reservation at once**. The seeded word "permanent" stays withdrawn
until [`BACKLOG.md`](BACKLOG.md) item 3 lands.

### 3 · Raw TCP, natively, for SSH and RDP and databases — **holds, and is the best-evidenced claim here**

A public TCP port forwards bytes with nothing parsed and no client-side helper
(`a13e586`, `relay/tcp.go`), including protocols where the server speaks first.
*Evidence, re-read 2026-08-31 20:06 UTC:* `pumasi.link:20000` has carried this
machine's own sshd for **49 671 s — 13 h 48 m — unbroken**
(`"opened_at":"2026-08-31T06:18:13Z"`), and it is how `m-gtr` is reached
(`pumasi-ops/RESOURCES.md` §4).
*What makes it a differentiator, cited:* an untimed raw TCP tunnel with no
account and no card is not offered by any of the three comparators on a free
tier — Pinggy includes free TCP but times the session out at 60 minutes,
LocalXpose excludes TCP from its free Starter tier entirely, and ngrok's free
TCP requires credit-card verification ([`MARKET.md`](MARKET.md) §1, §3).
*Falsified by:* a protocol that survives a direct connection but not this one;
or a session dropped by a timer rather than by one of its ends.
**Was bounded by the announce-before-bind race; that bound is lifted at
`1d9505c`.** The address used to be handed out before the listener existed, so
this claim was intermittently false in the first instant of a tunnel's life.
Re-measured by this seat at the post-`0041` pass, with the host's own port churn
excluded so the ordering is what is being measured: **28 dial refusals in 2000
tunnel openings at `83fd9f7`, 0 in 2000 at `1d9505c`**
([`STAGE.md`](STAGE.md) §2). The relay now binds the public port before the
auth response leaves and refuses the handshake if it cannot.
*Still bounded by:* **the deploy**, not the code. `pumasi.link` runs a
pre-`83fd9f7` binary (`STAGE.md` §1), so nothing above is true of the running
relay yet — that is `BACKLOG.md` item 1(i), blocked on **Q-014**. A tunnel
opened against the live relay today still gets the old ordering.

### 4 · No interstitial page, and no session timer — **holds; read the TLS gap below**

Nothing in the tree inserts an HTML warning page in front of a tunnel: a
visitor's request is forwarded and the response comes back. Nothing imposes a
session or bandwidth limit; the 13-hour tunnel above is the evidence, not the
absence of a feature.
*What makes it a differentiator, cited:* LocalXpose's free Starter tier prints
an "Interstitial warning page" and "Time limits"; Pinggy's free tier prints a
"60 minutes tunnel timeout" ([`MARKET.md`](MARKET.md) §1). ngrok's pricing page
prints no session limit, so no claim is made about it in either direction.
*Falsified by:* any code path returning a page the tunnelled service did not
produce, other than an error when the tunnel is genuinely unreachable; or a
disconnect at a fixed age.

**The gap that matters for webhooks — half-corrected, and only on `main`.**
There is **no TLS**. What changed at `83fd9f7` is that the relay no longer
*lies* about it: `-public-scheme` defaults to `http`, is validated once at
startup, and is read by all three surfaces that show a person an address
(`core/route.go:311`, `relay/dashboard.go:71`, `relay/sshingress.go:190`).
**What did not change is the internet.** `pumasi.link` runs a pre-`83fd9f7`
binary and, re-verified 20:06 UTC, still announces
`"url":"https://sshsteward.pumasi.link"` while `curl https://pumasi.link/`
fails to connect on a refused 443. Either way, every HTTP tunnel here is
plaintext and any sender that requires an `https://` destination cannot be
pointed at one. [`BACKLOG.md`](BACKLOG.md) item 1;
[`STAGE.md`](STAGE.md) §1 for the merged-versus-served split, and **Q-014** for
why the deploy has not happened.

### 5 · A local request inspector on `127.0.0.1:4040` — **not built**

Claimed in the 2026-08-30 seed. It does not exist: `web/` is an empty
directory, no code binds 4040, and there is no capture, replay or SSE anywhere
in the tree. Recorded here as unbuilt rather than removed, because it is a real
intent — [`BACKLOG.md`](BACKLOG.md) item 9 — and because a claim that quietly
disappears is how a value proposition starts lagging the product. All three
comparators ship one ([`MARKET.md`](MARKET.md) §2), so this is a gap and not a
difference of opinion.
*Would hold when:* a request that crossed a tunnel can be read and replayed
from a local page, without an account.

### 6 · Self-hostable, Apache-2.0, no lock-in — **holds**

Two binaries, one Go module, no dependency outside the standard library and
`golang.org/x/crypto` (`go.mod`), and `LICENSE` is present in this repository.
The relay you would run is the relay running at `pumasi.link`, from this
repository.
*What makes it a differentiator, and how far:* of the three comparators, the
only self-hosting option printed on a vendor page consulted on 2026-08-31 is
Pinggy's Enterprise plan ("Dedicated Servers / On Premise", price "Custom") —
[`MARKET.md`](MARKET.md) §1, §3. This claim is about what is offered below
Enterprise; it is **not** a claim that any of them is closed-source generally.
*Falsified by:* a hosted-only capability — anything the public relay does that
a self-run one cannot, or a build step needing a credential this repository
does not contain.

---

## What a reader should take from the shape of this file

Three claims hold and are load-bearing today; one holds but has a measured hole
in it that a coder packet is about to close; one is bounded by a certificate
nobody has installed; one is not built. The distribution barely moved this week
and that is the honest report: the `-public-scheme` release **removed a false
statement rather than adding a capability**, which is worth doing and is not
progress up the ladder.

`STAGE.md` says the same things in the same words, and every competitor number
this file leans on lives in `MARKET.md` with a URL and a date. If any two of the
three disagree, that disagreement is the defect.

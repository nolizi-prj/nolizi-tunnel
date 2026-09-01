# VALUE — Pumasi Tunnel

**Owned by the product-manager role**
([`pumasi-ops/roles/product-manager.md`](https://github.com/pumasi-ai/pumasi-ops/blob/main/roles/product-manager.md),
duty 2). Seeded 2026-08-30; honest pass 2026-08-31 at `3652e15`; post-release
pass 2026-08-31 at `83fd9f7`; post-`0041` pass 2026-08-31 at `1d9505c`, which
lifted claim 3's bound (below); **post-`0060`/`0066` pass 2026-09-01 at
`87244af`**, which re-read the live relay at 02:48 UTC and audited every
cross-reference in this file. Each claim is re-checked against the tree and the
live relay. Every claim carries what would falsify it, and the ones that are
false or bounded say so.

**What this pass changed, and what it did not.** No claim here moved. `fd523e8`
— the merge this evaluation was queued for — fixed the HTTP path's
announce-before-serve ordering, and **this file never made a claim it delivers**:
the equivalent promise for the TCP path is claim 3, whose bound was lifted at
`1d9505c`, and the HTTP path had no matching sentence to lift. What changed is
the file's *citations*: **four were wrong and all four are repaired below**, with
how each broke recorded beside it, because two of them broke in a way that will
recur.

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
**three** places this product loses — this paragraph said *two* until this pass,
and it was wrong on the day it was written: it has no TLS where all three
incumbents sell HTTPS as the ordinary case; its free stable name is *unclaimed*
rather than *owned*; and it runs **one $5–6/month machine in Chicago** where
every vendor named above runs an edge, so nothing in this file is an
availability claim. Nothing here should be read past those three.

---

## The claims, and what would falsify each

### 1 · Zero-client access from a stock `ssh` — **holds**

A tunnel opens with `ssh -R` from any machine that has an ssh client, with no
download and no account (`3652e15`, `relay/sshingress.go`).
*Evidence, re-read 2026-09-01 02:48 UTC:* `pumasi.link:2222` answers
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
[`BACKLOG.md`](BACKLOG.md) item 4, *the console never offers the zero-install
`ssh -R` command* (re-verified at `87244af`: `relay/dashboard.html` contains 0
occurrences of `ssh -R` and 1 of `git clone`).
**That number was stale until this pass, and it is correct again by
coincidence.** The entry was item 4 when this line was written at `dda04c7`,
became item 5 at `076f747`, and is item 4 again only because the entry above it
was ticked as delivered this pass — so for two evaluations this line pointed at
the port-pool bindability item instead, and job `0066` inherited the wrong
number into its own hand-off. A citation that is right by coincidence is the
same defect as one that is wrong, which is why the entry's title is now here
beside its number.

### 2 · A name you asked for, given back on reconnect — **holds, narrowly, and less than "permanent"**

`--subdomain myapi` is honoured when free, and `--tcp-port` asks for an exact
public port so the address survives an agent reconnect (`a5b77fc`).
*Falsified by:* a client that reconnects within seconds and is handed a
different name or port while the old one is free.
**What this is not:** ownership, and not persistence. Re-verified at `87244af`:
`Tunnel.Reserved` is written once, at `relay/relay.go:297`, and read nowhere in
the tree; the relay binary defines eleven flags and none of them is an auth
flag, so it can only run `AllowAll` (`relay/relay.go:40`); the registry
(`core/route.go:145`–`:147`) and the port pool (`core/portpool.go:27`–`:29`) are
in-memory maps with no persistence path. So another anonymous agent may take
your name in the gap between your reconnects, and a **relay restart drops every
name and reservation at once**. The seeded word "permanent" stays withdrawn
until [`BACKLOG.md`](BACKLOG.md) item 2 lands — *a subdomain belongs to nobody,
and nothing survives a relay restart*, the top build entry as of this pass.

*Observed rather than argued, this pass:* the live relay is now carrying **two**
tunnels, and the second — subdomain `skk6g7tyrs`, opened 01:48:23 UTC by an
agent this seat cannot identify — is exactly the anonymous claim on a free name
that this section says the relay cannot refuse.

### 3 · Raw TCP, natively, for SSH and RDP and databases — **holds, and is the best-evidenced claim here**

A public TCP port forwards bytes with nothing parsed and no client-side helper
(`a13e586`, `relay/tcp.go`), including protocols where the server speaks first.
*Evidence, re-read 2026-09-01 02:48 UTC:* `pumasi.link:20000` has carried this
machine's own sshd for **73 789 s — 20 h 30 m — unbroken**, the same connection
as at the last two passes (`"opened_at":"2026-08-31T06:18:13Z"`), and it is how
`m-gtr` is reached (`pumasi-ops/RESOURCES.md` §4). A second raw TCP tunnel —
`pumasi.link:20002` → a `"local_port":3389` — has been open 59 minutes on the
same relay, opened by someone this seat cannot identify; it is not evidence this
claim is asking for, and it is recorded in
[`BACKLOG.md`](BACKLOG.md) *Not on this list* rather than counted here.
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
in the tree — re-verified at `87244af`. Recorded here as unbuilt rather than
removed, because it is a real intent — [`BACKLOG.md`](BACKLOG.md) **item 11**,
*local request inspector on `127.0.0.1:4040`* — and because a claim that quietly
disappears is how a value proposition starts lagging the product. All three
comparators ship one
([`docs/ux/incumbent-ux-spec.md`](../docs/ux/incumbent-ux-spec.md), the
clean-room tour of 2026-08-30: the comparison table at **line 78** gives an
Inspector row for each of the three, and **§6.1, §6.4 and §6.5** give one
section per vendor), so this is a gap and not a difference of opinion.
*Would hold when:* a request that crossed a tunnel can be read and replayed
from a local page, without an account.

**Both of this section's citations were wrong, and they broke in two different
ways — which is why the repair is two repairs and not one.**

- ***"`MARKET.md` §2"* was never right.** It was added in `dda04c7`, the same
  commit that created `MARKET.md`, and that file has never contained the string
  `inspector` at any revision it has ever had (`grep -ci inspector
  roadmap/MARKET.md` → **0**, checked at every commit that touched it). What
  happened is visible in `MARKET.md` §2's own header: it says it is drawn *"from
  `docs/ux/incumbent-ux-spec.md` §1"* and it carries five of that source table's
  rows. **The Inspector row is one of the ones it did not carry.** The citation
  named the summary; the claim only ever lived in the source.
- ***"`BACKLOG.md` item 9"* was right when it was written**, at `e29dc0e`, where
  the inspector genuinely was item 9. It became item 11 at `076f747` and item 12
  at the `b3d251d` pass, and neither renumbering came back here. It is item 11
  again today, for a third and unrelated reason. **Nothing was wrong with the
  citation except that the thing it pointed into is deliberately reordered every
  pass** — which is the argument for naming the title beside the number, now
  written into `BACKLOG.md`'s own preamble.

Job `0066` reported the first of these and not the second; this pass found the
second while checking the first, and a third of the same kind at claim 1. One
correction to `0066`'s note, since it will be read again: it placed the
comparison row in the UX spec's **§3**. Line 78 sits in **§1**'s comparison
table; §3 is the CLI/agent experience, and the per-vendor inspector detail is
**§6**.

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

**The cross-reference audit this pass ran, stated so the next seat knows what
was and was not checked.** All **eight** `MARKET.md` references in this file —
the seven section citations plus the general one in the closing paragraph —
were read against `MARKET.md` at `87244af`, not carried:

| Where | Cites | Verdict |
| :--- | :--- | :--- |
| *The pain* | §1 | **holds** — §1's *"What this establishes"* states all three: paid stable hostname at three of three, TCP behind payment or a card at two, a printed session ceiling at two |
| *The comparison is bounded* | §4 | **was wrong, repaired above** — §4 records three bullets, not two; the third is the single-host availability caveat |
| claim 1, *fair comparison* | §1–2 | **holds** — §1 prints *"No account is required to start on the free tier"* for Pinggy; §2's table gives it the OS `ssh` client |
| claim 3, *differentiator* | §1, §3 | **holds** — the 60-minute timeout, the Starter-tier TCP exclusion and the card-verification line are all in §1, and §3 claim 1 restates them |
| claim 4, *differentiator* | §1 | **holds** — LocalXpose's *"Interstitial warning page"* and *"Time limits"*, Pinggy's *"60 minutes tunnel timeout"*, and ngrok printing no session limit |
| claim 5 | §2 | **was wrong, repaired above** — `MARKET.md` has never mentioned an inspector |
| claim 6, *differentiator* | §1, §3 | **holds** — Pinggy Enterprise, *"dedicated servers / on-premise"*, price *"Custom"* |
| closing paragraph | general | **holds** — not a section citation |

Two of the eight were wrong and both are repaired. The five `BACKLOG.md` item
citations were checked the same way: **two were wrong** (claims 1 and 5, above);
one — claim 2's *"until item 3 lands"* — was correct and is renumbered to item 2
by this pass's own re-rank rather than by any defect; and claim 3's `item 1(i)`
and claim 4's `item 1` are correct and unaffected.
**Nothing in `MARKET.md` itself was edited this pass**, and no competitor claim
in this file was widened — every repair points an existing claim at the file
that already carried its evidence.

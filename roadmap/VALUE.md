# VALUE — Pumasi Tunnel

**Owned by the product-manager role**
([`pumasi-ops/roles/product-manager.md`](https://github.com/pumasi-ai/pumasi-ops/blob/main/roles/product-manager.md),
duty 2). Seeded 2026-08-30; honest pass 2026-08-31 at `3652e15`; post-release
pass 2026-08-31 at `83fd9f7`; post-`0041` pass 2026-08-31 at `1d9505c`, which
lifted claim 3's bound (below); post-`0060`/`0066` pass 2026-09-01 at `87244af`,
which audited every cross-reference in this file; post-`0081` pass 2026-09-01
at `9e2de66`, which re-read the live relay at 05:51:04 UTC; **post-`9cc9e65`
pass 2026-09-01 at `9cc9e65`** (job `0093`), which re-read it at 18:51:06 UTC.
Each claim is re-checked against the tree and the live relay. Every claim
carries what would falsify it, and the ones that are false or bounded say so.
falsify it, and the ones that are false or bounded say so.

**What this pass changed.** `spec/0004-names-with-owners` **slice 2** merged
at `9cc9e65` (spec `9dd067a`), by a hand run outside the queue, **and claim 2
is again the claim it was built to make true** — so claim 2 is extended rather
than re-labelled. Its two halves are now both true on `main` and both fall
along the line this file keeps drawing: **what is true on `main` is not what
`pumasi.link` serves.**

- **On `main`:** a name and its public TCP port are held for whoever first
  proved they held them, across a disconnect, refused to everyone else — and,
  **when the relay is started with `-reservations <path>`, across the relay
  restarting**, for thirty idle days.
- **On `pumasi.link`:** none of it. The relay running there predates both
  slices, which this seat established without touching the host — see claim 2.
- **On neither, and not on the record either:** a second family's reading of
  the merge that delivered the restart half. It is
  [`BACKLOG.md`](BACKLOG.md) item 2 — *`9cc9e65` has no cross-family code
  review* — and it is why the seeded word *"permanent"* stays withdrawn even on
  `main`.

**No other claim moved, and this pass will not manufacture movement in one.**
Claims 1, 3, 4 and 6 hold as written; claim 5 is still not built. Claim 1's
bound from last pass — slice 1 narrowed the zero-install path — is unchanged
by slice 2 in either direction; see claim 1 and `BACKLOG.md` item 5.

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
*Evidence, re-read 2026-09-01 19:01 UTC:* `pumasi.link:2222` answers
`SSH-2.0-pumasi-tunnel`.
*Falsified by:* an ordinary OpenSSH client that cannot open a tunnel and read
its public address from the banner; or an account, key registration or download
becoming necessary first.
*Qualification:* the ingress is on **2222**, not 22 — port 22 on that host is
its own sshd. A `-p 2222` is part of every command.
***A second qualification, new at `9e2de66`, and it is a narrowing rather than a
caveat.*** Since slice 1 (claim 2) the CLI path can hold a name with a token and
**the ssh path cannot** — `relay/sshingress.go:162`–`:165` says so above the
`core.AuthRequest` it builds with no `Token` field (re-read at `9cc9e65`; slice
2 moved the lines and not the fact), and the username's `+` grammar cannot
carry one, because a token is itself a valid subdomain and `name+token` would
parse as a name. So a zero-install tunnel can now be **refused** a name somebody
else has claimed (frozen case **C-5**, and the refusal is correct) and can never
**claim or reclaim** one. That is strictly more refusals and no new capability,
on the path this file leads with. It is bounded rather than fatal — an ssh
tunnel is session-scoped by nature, with a person at the terminal holding it
open — and it is [`BACKLOG.md`](BACKLOG.md) item 5, *the zero-install `ssh -R`
path can be refused a name and can never hold one*, which is a **new** entry
this pass and needs an intent statement before a packet.
*Fair comparison:* Pinggy also requires no account to start and is also
SSH-first ([`MARKET.md`](MARKET.md) §1–2). This claim is that we match it, not
that we are alone.
*Not yet true of the one page a visitor sees:* the console still offers only
`git clone && go build` and the binary's flags, never the `ssh -R` command —
[`BACKLOG.md`](BACKLOG.md) item 4, *the console never offers the zero-install
`ssh -R` command* (re-verified at `9cc9e65`: `relay/dashboard.html` contains 0
occurrences of `ssh -R` and 1 of `git clone`). **Read that entry with item 5
beside it**: putting the command on the console sends more people down the one
path that cannot hold an address.
**That number was stale two passes ago and it is correct again — this time
because it was checked, not by coincidence.** The entry was item 4 when this line
was written at `dda04c7`, became item 5 at `076f747`, was repaired to item 4 at
`1853218`, and is **still item 4** after this pass's re-rank. A citation that is
right by coincidence is the same defect as one that is wrong, which is why the
entry's title is here beside its number and why every number in this file was
re-read against `BACKLOG.md` before the re-rank and re-pointed in the same
commit as it.

### 2 · A name you asked for, and — with a token — held for you while you are away, and across the relay restarting — **holds on `main` and nowhere a user can reach; still less than "permanent"**

**Rewritten at `9e2de66` when slice 1 made half of it true, and extended at
`9cc9e65` when slice 2 made the other half true — on `main`, with a flag, and
without a code review.** This is the claim
[`spec/0004-names-with-owners`](../spec/0004-names-with-owners/SPEC.md) exists
to make true.

**What holds on `main` since `9e2de66`, re-read at `9cc9e65`.** `--subdomain myapi --token <16+ chars>`
**claims** the name; with `--tcp --tcp-port P` it claims that public port too.
Both are then held for that token **across a disconnect** and refused to every
other caller, including one presenting no token at all. The owner reconnecting
with the same token is given both back. Delivered `4489fbe` · `c12d11a` ·
`20e9d57`; frozen cases **C-1** (the name) and **C-2** (the port).
*Re-verified in the tree by this seat at `9e2de66`, not carried:* `.Reserved`
returned **nothing at all** from `grep -rn "\.Reserved" --include=*.go` at
`1853218` and now returns **two non-test reads** — `relay/dashboard.go:80` and
`relay/relay.go:497` (was `:429`; slice 2 moved it); `core/reservation.go` is
pure and, since slice 2, carries `LastSeen`; `MinTokenLen` is **16**
(`core/reservation.go:48`, was `:45`) and a shorter token is **refused**, never
quietly downgraded to anonymous (case **C-6**).

**None of it is live, and this seat established that without touching the
host.** Slice 1 adds a `"reserved"` key to every tunnel in the status view
(`relay/dashboard.go:42` — a plain `bool` with no `omitempty`, so it is always
present). `GET http://pumasi.link/_pumasi/status` at **18:51:06 UTC 2026-09-01**
returns one tunnel and **no `reserved` key on it**. So the relay serving
`pumasi.link` predates `4489fbe`, and **for anyone using this product today this
claim reads exactly as it did two passes ago.** What blocks the deploy is
[`BACKLOG.md`](BACKLOG.md) item 1(i) and **Q-014**, and nothing here asks for it.

**What became true on `main` at `9cc9e65`, and what it costs to say so.** A
relay started with **`-reservations <path>`** keeps its reservation set on disk
— one JSON document of names, `sha256` token digests, ports and a `last_seen`
clock, written atomically and fsynced — and a relay restarted over that path
finds every name **and its port** still held for its token and refused to
everyone else. *Verified at `9cc9e65`:* `core/reservationstore.go` is new;
`LastSeen` is `core/reservation.go:68`; `cmd/pumasi-relay` has twelve flags and
the twelfth is `-reservations` (`main.go:46`); and the acceptance case,
`relay.TestAReservationOutlivesTheRelay`, **exists at
`relay/reservation_test.go:640`** and is a second `relay.New` over the same
path, not a reconnect. Delivered `9dd067a` · `9cc9e65`;
[`BACKLOG.md`](BACKLOG.md) *Delivered*, first entry. **Three bounds travel
with it.** It is **opt-in** — an empty flag is the relay exactly as before, and
a deploy that omits it delivers nothing here. It **expires**: a reservation
idle for thirty days is swept at the next load, so *held while you are away*
has a horizon. And **it has no code review** — the merge landed by a hand run
outside the queue with two spec-round transcripts and no code-round one
([`BACKLOG.md`](BACKLOG.md) item 2, *`9cc9e65` has no cross-family code
review*), so this file states the capability as measured and does not lean on
it. **The seeded word "permanent" stays withdrawn**: a thirty-day idle horizon,
trust on first use, a bearer secret on a plaintext wire and a flag the operator
has to set are four reasons, and the fifth is that nothing above is live.

**What a token proves, stated because "ownership" invites three things it does
not mean** (`SPEC.md` §3):

- **Not identity.** There is no account and no signup — **Q-002** — and no human
  is named by a reservation. *Owner* here means *the holder of a secret*.
- **Not permission.** `AllowAll` still answers *may this agent use this relay* —
  yes, to everyone. A token narrows *which name* an accepted agent may have; it
  does not decide whether the agent is accepted.
- **Not confidentiality.** It is a **bearer secret on a plaintext connection**:
  anyone on the path between agent and relay can read it and replay it. What
  bounds that is [`BACKLOG.md`](BACKLOG.md) item 1(ii), a certificate nobody has
  installed, and not this change. The relay stores `sha256(token)` and never the
  token (case **R-5**), so a leaked *reservation set* is not a set of live
  credentials — but the wire is not protected by that.

**And nothing was withdrawn from the anonymous case, on purpose.** An agent with
no token still gets a generated name, and still gets an *unclaimed* name it asks
for by name (`SPEC.md` §6, third column; frozen case **C-3** exists to go red if
anyone ever changes that). *Observed rather than argued:* at the last two
passes the live relay carried a second tunnel, `skk6g7tyrs`, opened by an agent
nobody here could identify; at this pass's 18:51:06 UTC read it is **gone**.
While it was there it was the anonymous case working as specified — taking a
**free** name anonymously is the behaviour slice 1 keeps; what slice 1 stops is
taking a **claimed** one — on a relay that in any event does not have the
change.

**One user of this product gets none of it: the zero-install path.** An `ssh -R`
client cannot present a token — see claim 1's new bound and
[`BACKLOG.md`](BACKLOG.md) item 5.

*Falsified by:* a client reconnecting with the same token and being handed a
different name or port while its own is held; a name claimed by one token being
given to another; a short token being accepted rather than refused; or an
anonymous agent being refused an **unclaimed** name, which would mean the
withdrawal `SPEC.md` §6 forbids had happened; or a relay restarted over a
`-reservations` path handing a held name or port to a stranger. *And falsified
in this file's favour by:* a reservation surviving a restart on a relay started
**without** the flag, or surviving more than thirty idle days — either would
mean this section is understating the product, and neither is true of the tree.

### 3 · Raw TCP, natively, for SSH and RDP and databases — **holds, and is the best-evidenced claim here**

A public TCP port forwards bytes with nothing parsed and no client-side helper
(`a13e586`, `relay/tcp.go`), including protocols where the server speaks first.
*Evidence, re-read 2026-09-01 18:51:06 UTC:* `pumasi.link:20000` has carried
this machine's own sshd for **131 572 s — 36 h 32 m — unbroken**, the same
connection as at every pass since the first (`"opened_at":"2026-08-31T06:18:13Z"`),
and it is how `m-gtr` is reached (`pumasi-ops/RESOURCES.md` §4). The second raw
TCP tunnel that was open beside it at the last two passes — `pumasi.link:20002`
→ a `"local_port":3389`, opened by someone this seat could not identify — is
**gone** at this read; it was never evidence this claim was asking for, and its
record is in [`BACKLOG.md`](BACKLOG.md) *Not on this list*.
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
session or bandwidth limit; the **36-hour** tunnel above is the evidence, not
the absence of a feature. (This line said *13-hour*, then *23-hour*, at earlier
passes — the same connection, older readings of it.)
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
through the one function that composes it — `core.Registry.PublicURL`
(`core/route.go:311`), called from `relay/dashboard.go:77` and
`relay/relay.go:371`. *Re-read at `9e2de66`; the previous citation named
`relay/sshingress.go:190` as a third call site and there are two, the ssh path
reaching the same address through `authorize`'s response rather than composing
its own.*
**What did not change is the internet.** `pumasi.link` runs a pre-`83fd9f7`
binary and, re-verified **18:51:06 UTC 2026-09-01**, still announces
`"url":"https://sshsteward.pumasi.link"` while `curl https://pumasi.link/`
fails to connect on a refused 443 (exit 7). Either way, every HTTP tunnel here is
plaintext and any sender that requires an `https://` destination cannot be
pointed at one. [`BACKLOG.md`](BACKLOG.md) item 1;
[`STAGE.md`](STAGE.md) §1 for the merged-versus-served split, and **Q-014** for
why the deploy has not happened.

### 5 · A local request inspector on `127.0.0.1:4040` — **not built**

Claimed in the 2026-08-30 seed. It does not exist: `web/` is an empty
directory, no code binds 4040, and there is no capture, replay or SSE anywhere
in the tree — re-verified at `9cc9e65`. Recorded here as unbuilt rather than
removed, because it is a real intent — [`BACKLOG.md`](BACKLOG.md) **item 12**,
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
  at the `b3d251d` pass, and neither renumbering came back here. It was repaired
  to item 11 at `1853218` and is **item 12** after this pass's re-rank, moved by
  the new entry inserted at 5 — the fourth number this one line has had.
  **Nothing was ever wrong with the citation except that the thing it points
  into is deliberately reordered every pass** — which is the argument for naming
  the title beside the number, now written into `BACKLOG.md`'s own preamble, and
  for re-pointing every number in the same commit as the re-rank, which is what
  this pass did.

Job `0066` reported the first of these and not the second; this pass found the
second while checking the first, and a third of the same kind at claim 1. One
correction to `0066`'s note, since it will be read again: it placed the
comparison row in the UX spec's **§3**. Line 78 sits in **§1**'s comparison
table; §3 is the CLI/agent experience, and the per-vendor inspector detail is
**§6**.

### 6 · Self-hostable, Apache-2.0, no lock-in — **holds**

Two binaries, one Go module, no dependency outside the standard library and
`golang.org/x/crypto` (`go.mod`, re-read at `9cc9e65`: `golang.org/x/crypto
v0.55.0` and `golang.org/x/sys v0.47.0`, both indirect), and `LICENSE` is present
in this repository. The relay you would run is **this repository's relay**.
*Sharpened this pass, because the old wording said something else:* it read
*"the relay you would run is the relay running at `pumasi.link`"*, and that is
no longer true in the only sense it could be checked — `pumasi.link` runs a
binary that is **nine Go-tree merges behind `main`**, seven of them changing
behaviour (`STAGE.md` §1). A self-hoster who builds from `main` today gets
**more** than the public relay serves, not less, so the claim survives; but this
file will not describe two different builds as one thing.
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

Three claims hold and are load-bearing today; **one holds on `main` and nowhere
a user can reach**; one is bounded by a certificate nobody has installed; one is
not built. **The distribution moved again this pass, and again on `main`
only.** Two coder passes — one under the queue, one by hand — closed the hole in
claim 2 on `main`: a name and its port belong to somebody, survive a disconnect,
and, when the operator sets a flag, survive the relay restarting. Every user of
`pumasi.link` is still waiting on a deploy that **Q-014** gates, and the merge
that closed the second half is waiting on a code review nobody has run
([`BACKLOG.md`](BACKLOG.md) item 2 — *`9cc9e65` has no cross-family code
review*).

**The honest summary of the week, which is not the same as the summary of the
merge.** Two ordering defects, a flaky suite and both halves of the ownership
gap have been fixed on `main`; **the product a stranger meets at `pumasi.link`
is unchanged since before 2026-08-31 10:47.** Nine merges touching the Go tree
now sit between the two, seven of them changing behaviour. Any sentence in this
file that reads as progress should be read against that.

`STAGE.md` says the same things in the same words, and every competitor number
this file leans on lives in `MARKET.md` with a URL and a date. If any two of the
three disagree, that disagreement is the defect.

**The cross-reference audit, run at the `87244af` pass and re-checked at
`9e2de66`.** All **eight** `MARKET.md` references in this file — the seven
section citations plus the general one in the closing paragraph — were read
against `MARKET.md`, which **no pass since has edited**; the verdicts below
stand unchanged and are **carried, not re-derived**, with that reason:

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

Two of the eight were wrong and both were repaired at the `87244af` pass.

**Re-audited at `9cc9e65`.** No entry was renumbered this pass, so every number
below still points where it did; **item 2 changed substance under the same
number** — *the relay-restart half* is delivered and the entry is now the
missing code review — so claim 2's two citations of it by title are re-pointed
in this commit, to *Delivered* and to the new item 2 respectively. The
`MARKET.md` verdicts above are still carried, with the same reason.

**The `BACKLOG.md` citations were re-audited at `9e2de66`, and this time none
was wrong.** Every one was read against `BACKLOG.md` **as it stood before this
pass's re-rank** — claim 1's item 4, claim 2's item 2, claim 3's item 1(i),
claim 4's item 1 and claim 5's item 11 — and all five pointed at the entry they
named. They are then **re-pointed for the new numbering in the same commit as
the re-rank**, so no revision of this repository exists in which they are wrong.
Claim 5's moves from 11 to 12 and claim 2's target is split; the other three are
unaffected. Claims 1 and 2 also gain **new** citations to item 5, an entry this
pass created.
**Nothing in `MARKET.md` itself was edited this pass**, and no competitor claim
in this file was widened or added — the only claim that moved is claim 2, and it
moved on this repository's own tree with the measurements beside it.

# STAGE — Pumasi Tunnel

**Current Stage:** `Alpha` — unchanged by this evaluation, and §4 says what it
would cost to move.
**Stage 1 exit gate:** **MET 2026-08-31**, carried forward, re-measured at
`87244af` by the seat writing this: **0 failures in 300** runs of the gate's own
command, **0 in 100** with `-cover`, **0 in 25** with `-race`. §2 gives every arm
and its count.
**The qualification this file has carried since the gate was first read is
spent, and what replaces it is not another defect.** The previous pass qualified
`MET` on *"one known unfixed ordering defect on the HTTP path"*; that defect was
fixed and merged at `fd523e8` and this seat verified the fix in the tree. **No
identified source of suite non-determinism is open.** What the gate is qualified
by now is the limit of the evidence rather than a named bug — one machine, one
seat, nothing re-running it, and one latent fixture defect still sitting in a
frozen file (§3.1b). §2 says it in full.
**This pass changes the evidence under `Q-024` and does not retire it.** That
entry turns on whether a flaky suite can support a `MET` reading, its named
retirement condition is the fix that has now landed plus the clean runs that now
exist, and **closing it, dating it or reading a default off it is the steward's
act and not this seat's.** `STAGE_PLAYBOOK.md` **Event 3 stays held** (§7).
**Stage 2 (Beta) work:** in progress. The `beta` *label* is not claimed, and
what holds it back is §4.
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

**Re-evaluated 2026-09-01 at `87244af`** (post-`0060`, the session-before-
announce merge, and post-`0066`, the README repair), `STAGE_PLAYBOOK.md`
Event 2. **Every number in §2 was taken in this pass, on this machine, at this
SHA, and the number of runs behind each is stated beside it** — job `0060`
published its own figures and none of them is carried here. Where a fact was
taken from an earlier pass without being re-taken, it says so in those words
rather than being presented as fresh.

**Re-evaluated 2026-08-31 at `b3d251d`** (post-`0047`, the test-port move),
`STAGE_PLAYBOOK.md` Event 2. Its text is kept below where it still holds.

**Re-evaluated 2026-08-31 at `1d9505c`** (post-`0041`, bind-before-announce),
`STAGE_PLAYBOOK.md` Event 2. The previous pass was at `83fd9f7` and its text is
kept below where it still holds.

**Evaluated 2026-08-31 at `83fd9f7`**, triggered by the release note
`pumasi/releases/2026-08-31-pumasi-tunnel-public-scheme.md` (`pumasi`
`33a96b0`; **Q-020**) — `STAGE_PLAYBOOK.md` Event 2. Every measurement below
was taken in this pass; where it disagrees with the number this file carried
before, the new one is here and the change is called out.

---

## 1 · What `main` does, and what `pumasi.link` serves — they are not the same

**Read this before anything else in this file.** A stage file that reports a
merge as the user-visible truth is the L-009 failure this fleet has already met
twice, so the two are separated here and stay separated until a deploy closes
the gap.

| | on `main` @ `87244af` | on `pumasi.link`, re-measured **02:48 UTC 2026-09-01** |
| :--- | :--- | :--- |
| Scheme announced | `http://` by default, `-public-scheme=https` when a terminator is in front | **`https://`** — `"url":"https://sshsteward.pumasi.link"` |
| Port 443 | not the relay's job, by design | **refused** — `curl https://pumasi.link/` exits 7, could not connect |
| Port 80 | serves the console and visitor traffic | **200** |
| ssh ingress 2222 | serves | **`SSH-2.0-pumasi-tunnel`** |
| TCP range in use | `-tcp-low`/`-tcp-high`, no default | **20000–20099** — `"tcp_range"`, unchanged, below the ephemeral floor (§3.1a) |
| A URL is serviceable when it is announced | **yes, on both paths** — TCP since `1d9505c`, HTTP since `fd523e8` | **no, on either** — the running binary predates both |
| Suite determinism | **0 failures in 425 runs** across three arms (§2) | not applicable — nothing runs a suite there |
| Live tunnels | — | **2** — `"count":2`, and the second is not this project's (below) |
| Which build is answering | — | **unknowable from any surface** — see §5 |

So: **a person who opens a tunnel today is still handed an `https://` address
that nothing answers.** Every cell in the right-hand column was re-taken this
pass, not carried: `GET http://pumasi.link/_pumasi/status` at 02:48 UTC returns
`"url":"https://sshsteward.pumasi.link"`, `https://pumasi.link/` does not
connect, `http://pumasi.link/` answers `200`, and `pumasi.link:2222` still
greets with `SSH-2.0-pumasi-tunnel`. The fix is merged and undeployed. The relay
process on the Vultr host predates `83fd9f7` and will keep announcing `https://`
until it is restarted.

**Five commits now sit on `main` that the host does not have, and three of them
change what a user would meet** — `83fd9f7` (the scheme), `3480990` (bind before
announce) and `fd523e8` (session before announce). The other two, `e40a224` and
`b3d251d`, touch tests and specs only; they are counted here because they are
merges waiting on the same restart, and they are named as harmless because a
count that does not distinguish them would overstate what a deploy would deliver.

**A second live tunnel appeared on that relay during this evaluation, and it is
not this project's.** The status read above reports `"count":2`. Alongside
`sshsteward` (`pumasi.link:20000` → this machine's port 22, `"fixed":true`,
`"opened_at":"2026-08-31T06:18:13Z"`, **`"age_secs":73789` — 20 h 30 m
unbroken**, the same connection as at the last two passes) there is
**`skk6g7tyrs`**, `pumasi.link:20002` → a `"local_port":3389`, `"fixed":false`,
opened `2026-09-01T01:48:23Z` and 59 minutes old at the read. **This seat did
not establish who opened it and will not guess.** It is recorded because it
changes two things this file has been asserting: the relay is no longer carrying
exactly one tunnel, and **a restart now costs two people their address rather
than one** — which is the fact `Q-014` is built on, and that entry's text still
describes the live set as *"exactly one"*. It is also the first thing on this
relay that is not the steward's own route, and it is `AllowAll` doing what §4's
fact 2 says it must.

**Why it has not been restarted, stated rather than left unsaid.**
`pumasi/DECISIONS.md` **Q-014** is open: it asks who may restart a relay that
keeps no durable state and whose live tunnel was, when the entry was written,
`sshsteward` alone — `pumasi.link:20000` → this machine's port 22,
`pumasi-ops/RESOURCES.md` §4's remote-access route, **open 73 789 s (20 h 30 m)
and unbroken** at the measurement above, the same connection rather than a new
one. As of this pass it is not alone; see above. Q-014 is **explicitly outside CHARTER Part 0's
proceed-on-default rule**. This evaluation therefore did not deploy, did not
treat the deploy as a judgement call, and **does not ask a coder packet to take
it either**. It is named as a blocker in `BACKLOG.md` item 1(i) and it waits for
the steward.

---

## 2 · Maturity gates

| Stage | Criteria | Status |
| :--- | :--- | :--- |
| **0. Candidate** | Scored by 3 model families (40/60) and selected by steward | **COMPLETE** |
| **1. Alpha** | Pure-core suite passes 100%; both public landing surfaces live | **COMPLETE 2026-08-31** — **0 failures in 425 runs** at `87244af` on this machine across three arms, and for the first time **no identified defect is behind the qualification**; what is left of it is stated below |
| **2. Beta** | Real end-to-end users complete workflows without engineer intervention | **IN PROGRESS** — §4 |
| **3. Launched** | Production hardening, cross-model regression, 7-day veto window | PENDING |

### The Stage 1 gate, re-read at `87244af` — the qualification changes shape

**The short answer to the question this evaluation was asked.** The previous
order's `BACKLOG.md` item 2 — the HTTP path announcing a URL before the session
that serves it exists — was the reason this file still qualified a `MET` gate. It is merged at
`fd523e8`, verified in the tree here, and the qualification it carried is gone.
Nothing replaced it: **there is no longer a named, open source of
non-determinism in this suite.** What remains is a statement about the evidence
rather than about the code, and it is written out below rather than dropped.

**What was measured, by this seat, on this machine, at `87244af`, this pass.**
Run counts are given because a gate whose number is inherited rather than
measured is what produced the wrong 12-run reading that raised **Q-024**
(rider (b)).

| Arm | Runs | Failures |
| :--- | ---: | ---: |
| `go test -count=1 ./...` — the command the gate names | **300** | **0** |
| `go test -count=1 -cover ./...` | **100** | **0** |
| `go test -race -count=1 ./...` | **25** | **0** |

**425 full-suite executions at this SHA, 0 failures**, 02:44:39–03:05:24 UTC.
*One honest caveat about the environment rather than the result:* an unowned
uncommitted change to `cmd/pumasi-tunnel/main.go` was in the working tree
throughout, and this seat left it there deliberately
([`BACKLOG.md`](BACKLOG.md), *Not on this list*). It is compiled by
`go test ./...` and exercised by nothing — that package reports *no test files*
and is imported by no test — so it cannot have moved these figures; it is named
because a run taken on a tree that is not exactly `87244af` should say so.
Job `0060` published 300 / 100 / 25 for the same three arms; these are this
seat's own runs and not that report. The host did not get quieter to produce
them: `/proc/sys/net/ipv4/ip_local_port_range` is still **32768 60999**, four
`workerd` processes were running throughout, and `ss -tan` found **9** sockets
inside 34500–34599 — the block §3.1b is about — several of them `ESTAB`. It
found **no listening socket from any other process** inside 21000–21039 or
20500–20559, the two blocks the suite actually binds.

**The whole gate is not `PASS` at this SHA, and the reason is not the tests.**
`pumasi/tools/gate.sh` at `87244af` prints `── 1/4 tests … tests: PASS` and then
`GATE: FAIL` on `missing trailer: Spec:` — because `87244af` is a README commit
and a documents commit has no spec to name. **Step 1 is the thing Stage 1's exit
gate names, and step 1 passes.** The previous pass's *"40 × `GATE: PASS`"* was
taken at a code SHA that carried a `Spec:` trailer and it is not reproducible
here for a reason that has nothing to do with this product's suite; it is not
carried forward, and no whole-gate figure is claimed at this SHA.

**Why the gate is still qualified, and what the qualification now is.** It is no
longer *"the number cannot be measured"*, and it is no longer *"one known
unfixed ordering defect on the HTTP path"*. Both of those are closed. What is
left has three parts and not one of them is a bug in this repository:

1. **Every figure this gate has ever had was taken on one machine, by the seat
   that wrote the file.** Nothing re-runs it — no CI, no second host, no other
   party. That is `Q-025`'s question arriving through this file's front door, and
   `Q-024`'s rider (c) already says in as many words that this is weaker than CI
   and is to be written down as weaker rather than as absent.
2. **The defect that was closed this pass was never reproduced here.** Job
   `0047` saw it twice in 240 runs; the previous pass saw it zero times in 640
   executions plus 4,900 targeted visitor requests. What establishes the fix is
   not an incidence figure but a **frozen case that fails for the right reason**:
   C-1 `TestVisitorIsNotAnsweredBeforeTheSessionExists` holds the relay inside
   the announce with a `gatedConn` and went red at `8540b89` with the exact
   `404 No tunnel is open` string, deterministically, in 0.00 s. That is better
   evidence than a rate, and it is a different kind of evidence, which is worth
   saying out loud rather than folding into a run count.
3. **One latent fixture defect is still in a frozen file** — §3.1b, `BACKLOG.md`
   item 8 — and it is blocked on a governance reading rather than on work (§8).
   It binds nothing today and this pass re-measured that; the day it binds, it
   stops being latent silently.

**What this does and does not do to `Q-024`.**

- It **does** complete both halves of that entry's own stated retirement
  condition. The named fix — bind before the response leaves, on both the agent
  and the ssh paths — landed at `1d9505c`. The *"40 clean runs of each
  invocation recorded in `roadmap/STAGE.md`"* exist here at 300 and 100, plus 25
  with `-race`. And the successor defect that the previous pass gave as its
  reason for not retiring the entry — the non-determinism having *moved* to the
  HTTP path rather than gone — is itself now merged.
- It **does not** retire it, and this pass does not attempt to. **Closing a
  window is the steward's act and never the seat that records the evidence**,
  and this pass is precisely the seat that recorded the evidence. The window is
  not closed, not dated, not extended, not softened, and no default is read off
  it here. The evidence is written into `pumasi/DECISIONS.md` under that entry in
  its own idiom and stops there.
- **The gate therefore stays `MET` and stays qualified**, under Q-024's own
  named default, and the qualification is now: **the pure-core suite passed 425
  of 425 on one machine at this SHA, with no identified defect open behind the
  number and nothing but this seat re-taking it.** That is a stronger claim than
  the last one and a narrower one than it sounds.
- **Q-024's rider (a) still binds and its force is undiminished by the better
  number.** No stage-promotion announcement may be published off this gate —
  `STAGE_PLAYBOOK.md` **Event 3 stays held**, §7 — and nothing public may quote
  the gate's figure. Rider (c) is why: evidence getting stronger is not a
  promotion ground, and evidence strength moving is nearly the whole of this
  pass again.

**Coverage, re-measured at `87244af` over the 100 `-cover` runs.**

| Package | At `b3d251d` (last pass) | At `87244af` (this pass) |
| :--- | :--- | :--- |
| `core` | 80.3% | **80.3%** |
| `mux` | 83.5% | **84.0%** |
| `relay` | 82.0% | **83.3%** |
| `agent`, `cmd/pumasi-relay`, `cmd/pumasi-tunnel` | 0.0% | **0.0%** |

`relay`'s 1.3 points come from `fd523e8`'s own 468 lines of new acceptance test
in `relay/sessionorder_test.go` — a change improving the evidence for the code
it touched, which is the ordinary and unremarkable case. `mux` is back to the
84.0% it read at `83fd9f7`, having read 83.5% last pass; nothing in either pass
touched `mux`, and this file reports the movement rather than inventing a cause
for half a point. **`agent` is still 0.0% and the merge that just landed did not
change that** — every one of its new test lines is in `relay` (§3.2).

**Both surfaces are live.**
- *Surface A, the commons catalog*: `pumasi-web`
  `content/products/pumasi-tunnel.md` (`c2084a8`, 2026-08-30);
  `https://pumasi.ai/products/pumasi-tunnel/` → 200, and the entry appears in
  `https://pumasi.ai/llms.txt`. **But not `pumasi/catalog.json` — see §6.**
- *Surface B, the product's own domain*: `http://pumasi.link/` → 200, serving
  the console (`relay/dashboard.html`, `b3585f6`).

**And the product carries real traffic.** The `sshsteward` tunnel in §1 —
**20 h 30 m unbroken** at 02:48 UTC, carrying this machine's own sshd across
`pumasi.link:20000`. That is `RESOURCES.md` §4's remote access path, working,
and it remains the strongest evidence this product has. **A second tunnel was
open beside it at the same read** (§1), forwarding somebody's `3389`; it is
recorded, it is not attributed, and it is not counted as evidence of anything
until somebody can say whose it is.

---

## 3 · What the gate's number does not cover — three flags, and the ordering ones are closed

The Stage 1 gate is `go test` exiting 0. The gate's number is only as good as
what it covers (L-006), so what it does not cover is recorded here rather than
left implicit. Every entry below is a `BACKLOG.md` item. **The flag this section
has led with since it was written — an announce-ordering defect — is now closed
on both paths**, and §3.1 is the record of that rather than a live warning.

### 3.1 · Both announce-ordering defects are closed, on both paths

**Retired by merge, and verified in the tree by this seat at `87244af` rather
than read off a commit subject.** `relay.ServeAgent` now registers the tunnel,
**binds** the public TCP port if there is one (`relay/relay.go:175` — `0041`'s
fix, `spec/0002`), then builds the mux session, **takes `r.mu`, installs
`r.sessions[resp.AgentID]` and writes the auth response inside that one critical
section** (`:222`–`:231`), deleting the session again in the same section if the
write fails (`:232`–`:236`). `ServeHTTP` takes the same lock to read
`r.sessions` (`:255`–`:257`), so a visitor arriving in what used to be the window
waits for the lock instead of being answered `404 No tunnel is open`. That is
`fd523e8`, `spec/0003-session-before-announce`, and it closes the flag this
section carried last pass.

**The fix is not the one this project predicted, and the difference is the part
worth keeping.** The previous order's `BACKLOG.md` item 2 said the insert
*"can simply move above"*
the announce. It cannot: the announce is written raw (`relay/relay.go:327`) while
`mux.Session.Open` writes a `FrameOpen` on the same connection under the
session's own `writeMu` (`mux/session.go:88`, `:102`, `:181`), and the agent is
in `core.DecodeFrame` waiting for exactly one frame. A bare reorder puts a stream
frame ahead of the auth response and **drops the tunnel** — C-1 answering `502`
where it used to answer `404` honestly, C-2 waiting 10.76 s for an `OnConnect`
that never comes (`spec/0003/SPEC.md` §2, §6). **It would have passed the 500
clean runs this file published at `b3d251d`.** The full record is in
[`BACKLOG.md`](BACKLOG.md) under *Delivered*, because that is where a seat
looking up the old item 2 will land.

**What this does not entitle anyone to say.** It does not make the product more
reliable for any user: `fd523e8` is the third of three merged behaviour changes
that `pumasi.link` does not have (§1). It does not retire **Q-024** (§2). And it
does not turn a code reading into a reproduction — the defect's incidence was
never measured on this host by any pass, and what shows the fix works is a frozen
case that fails deterministically without it, not a rate.

### 3.1a · The product-side half, which is not the tests' fault

`core.PortPool` "does no I/O — it decides which number to use; binding the
listener is the relay's job" (`core/portpool.go:21–22`). When the number it
picks turns out not to be bindable, `relay.listenTCP` returns the port and an
error (`relay/tcp.go:66–70`) and the relay refuses the tunnel outright — it does
not ask for the next of the 99 other free ports. One busy port defeats a
100-port pool. `-tcp-low`/`-tcp-high` also accept a range wholly inside the
ephemeral range with no warning. The **running** relay's range is 20000–20099
(measured, §1), below the ephemeral floor, so this is bounded in production
today and unbounded for anyone who configures otherwise. `BACKLOG.md` item 3 —
*a public port the pool believes is free may not be bindable*.

### 3.1b · A frozen acceptance case still draws from the ephemeral range

`relay/scheme_test.go:314`–`:315` still configures 34500–34599, inside this
host's ephemeral range. It is inside **A-10**
(`TestSchemeChangesNothingButTheScheme`, `:295`), frozen under `spec/0001`. Job
`0047` moved it using CHARTER §3 requirement 2's own remedy — amend the spec in
the open, take a fresh cross-family spec review — which gemini approved; the
code review then objected under the same clause because the *builder* wrote the
amendment, and a cited objection governs, so the whole half was reverted.

**Its measured cost is nothing, and this pass re-measured that rather than
carrying it.** A-10 calls `relay.New`, which hands the range to
`core.NewPortPool` (`relay/relay.go:124`) — a type that *"does no I/O"*
(`core/portpool.go:21`–`:22`). It binds nothing. `ss -tan` found **9** sockets
inside 34500–34599 during this pass, several of them `ESTAB` rather than merely
lingering, and the suite still passed **425 of 425**. The two constants were
re-read at `87244af` and are unchanged. `BACKLOG.md` **item 8**, ranked below
everything buildable because what blocks it is a governance reading rather than
work — see §8.

### 3.2 · Three packages have no test files, `agent/` among them

`agent`, `cmd/pumasi-relay` and `cmd/pumasi-tunnel` report *no test files* and
0.0% coverage. `core`, `mux` and `relay` are the pure core the gate names and
they are covered. `agent/` is not core, but it is half of every tunnel, and
today nothing exercises it except the relay's end-to-end tests using it as a
fixture. Re-measured at `87244af` over 100 `-cover` runs: still **0.0%**, while
`relay` rose to 83.3%. **The merge this evaluation was queued for is the sharpest
illustration this flag has had**: `fd523e8` added 468 lines of acceptance test
and every one of them is in `relay/sessionorder_test.go`, so the package that is
half of every tunnel gained nothing — and one of the three frozen cases in that
file drives the *agent* through a failed handshake to prove its point.
[`BACKLOG.md`](BACKLOG.md) item 5 — *`agent/` has no tests*. **Fixer: the
coder.**

---

## 4 · Why not `beta`

`beta` means strangers can rely on it and their data survives. **Each fact
below was re-verified against the tree at `87244af` in this pass**, by grep and
by reading, not carried. The order is the backlog's. **None of the three
changed, and that is the honest report on a pass that merged a real fix**:
`fd523e8` closed an ordering defect on the HTTP path, which is none of the three
things below.

1. **There is no TLS, and what is running still says there is.** *Changed by
   the release, and only halfway.* On `main` the relay now announces the scheme
   it serves — `-public-scheme`, default `http`, validated once at startup,
   applied in one place, read by all three surfaces. **On `pumasi.link`, nothing
   listens on 443 and the relay still prints `https://`** (§1). Even once
   deployed, every HTTP tunnel here is plaintext and no `https://`-only webhook
   sender can be pointed at one. [`BACKLOG.md`](BACKLOG.md) item 1 — **operator
   action plus a deploy blocked on Q-014, not a build.**
2. **A name belongs to nobody.** Still true. `Tunnel.Reserved` is written at
   `relay/relay.go:297` (it has been line 236, then 249, then 274, now 297; the
   line keeps moving and the fact does not) and is **never read anywhere in the
   tree** — the field is defined at `core/route.go:127`, written once, and read
   nowhere. The relay binary defines **eleven flags and no auth flag** — `grep -n
   auth cmd/pumasi-relay/main.go` finds nothing — so `AllowAll`
   (`relay/relay.go:40`, installed at `:97` when `cfg.Auth` is nil) is the only
   authenticator it can run. Any anonymous agent may take any free name,
   including one another person is using between reconnects.
   **Observed this pass rather than argued**: a second tunnel appeared on the
   live relay at 01:48 UTC under an assigned name, opened by an agent nobody
   here can identify (§1). That is this fact, happening.
   [`BACKLOG.md`](BACKLOG.md) item 2.
3. **Nothing survives a restart.** Still true. The registry
   (`core/route.go:145–147`) and the port pool (`core/portpool.go:27–29`) are
   plain in-memory maps, and there is no persistence path anywhere in `core/` or
   `relay/`. A relay restart drops every subdomain, every reserved TCP port and
   every live tunnel — including the one carrying this machine's remote access,
   which is precisely why Q-014 exists — and as of 02:48 UTC it would drop
   **two** tunnels rather than one (§1). `--tcp-port` keeps an address across an
   *agent* reconnect (`a5b77fc`), not across a relay one.
   [`BACKLOG.md`](BACKLOG.md) item 2.

**In the order that shortens the distance fastest**, which is the backlog's
order and its reasoning in one line each — and **this is §4's cost-to-move,
changed by this pass**: **item 2** first — *a subdomain belongs to nobody, and
nothing survives a relay restart* — because it is the whole of facts 2 and 3
above and because it is what **retires Q-014**: once a restart costs nobody their
address, deploying stops being a steward question and becomes an ordinary one.
Then **item 3**, the relay's own bindability gap, bounded in production today
only because the deployed range happens to sit below the ephemeral floor.
**Item 1** runs in parallel and needs no coder, only a decision and a
certificate.

**What the previous pass put first here is done, and it bought nothing this
section can spend.** The mux session installed after the URL is announced was
this section's first cost-to-move item and it is delivered at `fd523e8`.
**The distance to `beta` did not shorten**, and this file will not pretend
otherwise: none of the three numbered facts above moved, the change is one of
three that `pumasi.link` still does not have, and what it bought was a gate whose
qualification is no longer a named defect — evidence strength, which Q-024
rider (c) says in as many words is not a promotion ground. **Item 2 is on top by
subtraction rather than by having become more urgent.** It is also, by a wide
margin, the largest item on the list, so the distance to `beta` is now honestly
described as *one large piece of work* rather than *a reordering and then a large
piece of work*.

**Also gating, from `PRODUCT-RULES.md`** (v1.0, read fresh 2026-09-01; still
only on the unmerged `worktree-product-rules` branch, `0115758` — **Q-017** —
and its absence from `main` is not compliance): **PR-1** (a user-visible version
number) binds **always** and this product has none anywhere — which §5 shows is
no longer theoretical; **PR-2** (in-app feedback) binds at the `beta` promotion
and is unbuilt. `BACKLOG.md` items 6 and 7. **Q-017 re-checked this pass, and it
is the ninth consecutive evaluation to report the same thing:**
`PRODUCT-RULES.md` is **still not reachable on `pumasi` `main`** — `git ls-tree`
at `pumasi` `2ab3a4f` finds it on neither `main` nor `origin/main`, only on
`worktree-product-rules` (`0115758`). Read fresh from that branch for this
evaluation, as duty 1 of the role file requires. Its absence from `main` is not
compliance, and this seat neither merged it nor proposed merging it — that is
Q-017's own question and it is the steward's.

## What `launched` additionally requires

Stage 2's exit gate — real end-to-end users completing workflows without an
engineer — plus production hardening. Not enumerated further while §4 is open.

---

## 5 · The version gap stopped being theoretical this week

`PRODUCT-RULES.md` PR-1 asks for a version that moves and is user-visible.
There is now a build on `main` that behaves differently from the build on the
host, and **no surface of this product will tell you which one is answering** —
not the console, not `/_pumasi/status`, not the logs. This evaluation could only
distinguish them by reading the `url` field and inferring from its scheme. That
is a diagnostic accident, not a capability. **It got worse this pass, not
better:** five merges on `main` are now absent from the host, three of them
behaviour changes (`83fd9f7`, `3480990`, `fd523e8`) and two test-and-spec only
(`e40a224`, `b3d251d`), and **the same single inference still has to carry all
five** — it can distinguish "pre-`83fd9f7`" from "post-", and nothing else.
This is the fourth consecutive evaluation to make that inference.
[`BACKLOG.md`](BACKLOG.md) item 6 — *PR-1 compliance: a version that moves and
is user-visible*. **Fixer: the coder** — `core.AuthRequest.ClientVersion` is
already in the wire protocol (`core/handshake.go:33`) and no binary of this
product sets it; the only writer is `relay/sshingress.go:165`, filling it with
the *ssh client's* version string.

**A second reader arrived this pass and made the gap concrete.** The status
endpoint showed a tunnel this seat could not attribute (§1). Had a version
string been on the wire, the question *"which build is that agent running"*
would have been answerable; as it is, neither half of that connection can be
identified from any surface this product exposes.

---

## 6 · Known gaps a user should know about today

- **No TLS.** Every tunnel is plain HTTP, whatever URL the relay printed — and
  the running relay still prints `https://` (§1).
- **No accounts, no tokens in force, no name ownership.**
- **One relay, one host** (`RESOURCES.md` §3: Vultr, Chicago, ~$5–6/month). A
  restart or a host failure ends every tunnel — **two of them, as of 02:48 UTC
  2026-09-01** (§1). Tailscale is kept as the independent fallback for reaching
  `m-gtr`, deliberately.
- **The commons index does not know this product exists.** `pumasi/catalog.json`
  contains **zero occurrences of the string `tunnel`** — no `products[]` entry,
  no `items[]` entry — re-verified this pass at `pumasi` `2ab3a4f`, `grep -c
  tunnel catalog.json` = **0**. `README.md` tells every
  arriving agent to start with that file and treats it as the charter's
  duplication check, so an agent running that check today is told Pumasi Tunnel
  does not exist. **Recorded here and deliberately not fixed:** it is not this
  repository's file and **no role file owns it** — `pumasi/DECISIONS.md`
  **Q-019**, open, whose named default would give first registration to the
  marketing manager and ongoing `status`/`maturity` upkeep to this seat. Until
  that resolves, this is the honest place for it.
- **The local request inspector on `127.0.0.1:4040` does not exist** — `web/` is
  an empty directory, re-checked at `87244af`. `VALUE.md` claim 5 says so, and
  the commons catalog page already disclaims it correctly (`pumasi-web`
  `843bdef`). *(`VALUE.md` claim 5 cited the wrong two files for that gap until
  this pass; both citations are repaired there and the repair is explained in
  place, because the way each one broke is different and one of them will
  recur.)*
- **No client TUI:** `cmd/pumasi-tunnel` is flags and logs.
- **No version anywhere** (§5).
- **`go test ./...` is no longer at the mercy of your other processes, and the
  two ordering defects behind its flakiness are both closed.** The suite's TCP
  harness draws from 21000–21039 and the bind-order cases from 20500–20559, both
  below the 32768 ephemeral floor; the TCP announce-before-bind was fixed at
  `1d9505c` and the HTTP announce-before-serve at `fd523e8`. On the machine this
  pass ran on — the same machine where the suite failed **every** run two passes
  ago — it passed **425 of 425** across three arms (§2). What remains: one
  acceptance case still names 34500–34599 but binds nothing (§3.1b,
  `BACKLOG.md` item 8), and the fact that nothing but this seat, on this machine,
  has ever taken any of these numbers (§2).

---

## 7 · For the marketing manager, from this evaluation

None of this is this seat's to write, and all of it is a page that contradicts a
file. Job `0066` took the product's own `README.md` on 2026-09-01 and this
section is re-stated against what is left.

1. **`pumasi-web`'s lead sentence "There is no hosted relay" is false**, and has
   been false at three consecutive evaluations. Re-measured 02:48 UTC
   2026-09-01: `pumasi.link` resolves to `64.177.118.159`, answers `200` on 80,
   greets `SSH-2.0-pumasi-tunnel` on 2222, and is carrying **two** tunnels, one
   of them unbroken for 20 h 30 m. `tunnel.pumasi.ai` indeed does not resolve
   (`getent hosts` exits 2); the hosted relay is on the other domain. **`0066`
   filed this as a `priority: high` hand-off for the commons marketing seat and
   it has not been taken.**
2. **The gate table in §2 has changed again** — the coverage figures moved
   (`relay` 82.0% → 83.3%, `mux` 83.5% → 84.0%) and, more importantly, **the
   Stage 1 qualification is no longer a named defect**. A page that quotes the
   old qualification is quoting something this file has withdrawn. See item 5
   before quoting the new one: the answer is still *do not*.
3. **[`MARKET.md`](MARKET.md) is the only place a public page may take a
   competitor claim from.** Every figure in it carries a vendor URL and a fetch
   date, and its §4 records the **three** comparisons that go *against* this
   product — TLS, unowned names, and one host against three vendor edges. (This
   section said "two" until this pass, as did `VALUE.md`; §4 has had three
   bullets since it was written.) A page that states a competitor price without
   that citation is the failure `pumasi-booking` `0d1674d` already had to undo.
   **And `MARKET.md` says nothing about request inspectors** — that comparison
   lives in `docs/ux/incumbent-ux-spec.md`, line 78 and §6, which is where
   `VALUE.md` claim 5 now points.
4. **Nothing public may say the `https://` problem is fixed for users, and the
   same now goes for the ordering fixes.** Three merged behaviour changes sit on
   `main` that `pumasi.link` does not have (§1). The distinction between merged
   and served is the whole of this file's §1 and it has not narrowed.
5. **`STAGE_PLAYBOOK.md` Event 3 is still held — and the case for firing it has
   never looked better, which is exactly when the hold matters.** That trigger
   chains a stage-promotion announcement and a public badge update off a product
   manager confirming an exit gate `MET`. This evaluation kept Stage 1 at `MET`
   on runs it took itself — **0 failures in 300 / 100 / 25** across three arms
   (§2) — and, for the first time, **with no identified defect standing behind
   the qualification.** A previous pass would have read that as licence. **It is
   not, and nothing public may be published off this gate.** Three reasons, all
   binding: **Q-024 is open and unretired**, and its rider (a) holds regardless
   of the rate; rider (c) says evidence getting stronger is not a promotion
   ground, and evidence strength moving is nearly the whole of this pass for the
   third time running; and §4's three reasons not to be `beta` are all
   untouched — the product did not get better for any user this pass, because
   nothing deployed.

   Any page still quoting *"passes, 37 times in 40"*, the 0-in-40, the 0-in-500,
   or *"one known unfixed ordering defect on the HTTP path"* is quoting something
   this file has withdrawn. The current figures are the three in §2 and they
   carry their run counts. Nothing is lost by waiting: this product is `alpha`,
   is not asking to move, and §4 lists three separate reasons it should not.

6. **What a page may now say that it could not before, stated precisely because
   the temptation to overstate it is real.** Two things. *(a)* **Both
   announce-ordering defects are fixed on `main`** — the TCP one measured at 28
   dial refusals in 2000 before and 0 after (`1d9505c`), the HTTP one closed at
   `fd523e8` and pinned by a frozen case that fails deterministically without it.
   *(b)* The suite is deterministic on a machine where it once failed every run.

   **Neither may be published as a Stage 1 pass rate, and both carry the same
   caveat: none of it is live.** All three behaviour changes are behind the
   undeployed restart in §1. A page may say the project fixed two ordering
   defects and a flaky test suite. It may **not** say the product got more
   reliable for anyone using `pumasi.link`, quote the gate, imply a promotion, or
   attach a run count to any of it.

7. **One thing this file cannot help with, offered so nobody publishes it by
   accident.** A second, unattributed tunnel is live on the relay (§1). It is the
   first traffic here that is not the steward's own route, and it is **not**
   evidence of adoption, a user, or a customer — nobody has established what it
   is. It is in this file because a restart now costs two parties rather than
   one, and for no other reason.

---

## 8 · An open governance reading — `pumasi/DECISIONS.md` **Q-030**, still open

Not marketing's and not this seat's to decide, recorded here because §3.1b
depends on it and a reader of that section should be able to find out why a
one-constant fix is not simply queued. It was escalated by the `b3d251d` pass
and is filed as **Q-030**, with a named default and no deadline; nothing here
closes, dates or softens it.

**The question.** May a builder use CHARTER §3 requirement 2's own remedy — *"If
the tests are wrong, amend the spec in the open and take a fresh cross-family
spec review"* — to amend a **frozen** acceptance case, given that Part 3
requirement 1 has the builder authoring every spec in the first place?

**Why it is live rather than academic.** Job `0047` did exactly that for A-10:
amendment written in the open as SPEC 0001 §7, fresh cross-family spec review
obtained and approved. The code review then objected under the same clause
because the builder wrote the amendment. Cited objections govern, so the work was
reverted. **The reviewers contradicted each other and each switched sides on the
same clause** across the two ranges reviewed — one approved the smaller change
and objected to the larger, the other did the reverse; all four transcripts are
committed under `reviews/`. And `spec/0002` §6.5 is precedent the other way: the
same builder amended a frozen *fixture* in the open (Amendment 3, each case
given its own port block) and it stood, reviewed and merged. **Both readings
cannot hold.** If the objection is right, requirement 2's remedy is unavailable
to anyone, because requirement 1 makes the builder the author; if §6.5 is right,
A-10 could have been moved the same way.

**A third data point arrived with `fd523e8`, and it is the avoidance case.**
Job `0060` found that `spec/0002` §2 states something its own change
contradicts — that the mux session cannot be created before the handshake
response. It did **not** amend that frozen spec. It wrote a new one,
`spec/0003-session-before-announce`, and said why in as many words: folding a
contradiction into a frozen document as an *"amendment"* is how a rule forks
(L-007). That is a builder routing around this question at the cost of a second
spec covering one rule, and it is worth weighing beside the two instances Q-030
already records — because the cost of leaving the reading open is not only the
reverted fix, it is also the specs written to avoid needing it.

**This seat proposes and does not decide**, and it has not set a window or a
deadline. Q-030's own text points at *"`pumasi-tunnel` `BACKLOG.md` item 9"*;
after this pass's re-rank that entry is **item 8**, unchanged in substance.

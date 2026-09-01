# STAGE — Pumasi Tunnel

**Current Stage:** `Alpha` — unchanged by this evaluation, and §4 says what it
would cost to move.
**Stage 1 exit gate:** **MET 2026-08-31**, carried forward, **re-measured at
`9e2de66` by the seat writing this** — the figures are in §2 with the run count
beside each arm. Job `0081` published *"three consecutive runs plus `-race`"*;
none of that is carried here.
**The qualification is the same one the last pass arrived at, and it is still not
a defect.** No identified source of suite non-determinism is open. What qualifies
the gate is the limit of the evidence — one machine, one seat, nothing re-running
it — and one latent fixture defect still sitting in a frozen file (§3.1b). §2
says it in full.
**One of §4's three reasons not to be `beta` moved this pass, and it is the
first time any of them has.** `spec/0004-names-with-owners` **slice 1** merged:
fact 2, *a name belongs to nobody*, is **no longer true on `main`** — and is
still entirely true of what `pumasi.link` serves. Fact 3, *nothing survives a
restart*, is untouched. **The stage does not move**, and §4 says why in the same
words it used before: nothing deployed, so no user got anything.
**This pass adds evidence under `Q-024` and under `Q-014`, and retires
neither.** Closing a window, dating it, or reading a default off it is the
steward's act and never the seat that records the evidence — and this pass is
precisely that seat. `STAGE_PLAYBOOK.md` **Event 3 stays held** (§7).
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

**Re-evaluated 2026-09-01 at `9e2de66`** (post-`0081`, the names-with-owners
slice-1 merge — eight commits, `1853218` → `9e2de66`), `STAGE_PLAYBOOK.md`
Event 2. **Every number in §2 was taken in this pass, on this machine, at this
SHA, and the number of runs behind each is stated beside it.** Where a fact was
taken from an earlier pass without being re-taken, it says so **in those words**
— *carried, not confirmed* — rather than being presented as fresh.

**Re-evaluated 2026-09-01 at `87244af`** (post-`0060`, the session-before-
announce merge, and post-`0066`, the README repair), `STAGE_PLAYBOOK.md`
Event 2. Its text is kept below where it still holds.

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

| | on `main` @ `9e2de66` | on `pumasi.link`, re-measured **05:51:04 UTC 2026-09-01** |
| :--- | :--- | :--- |
| Scheme announced | `http://` by default, `-public-scheme=https` when a terminator is in front | **`https://`** — `"url":"https://sshsteward.pumasi.link"` |
| Port 443 | not the relay's job, by design | **refused** — `curl https://pumasi.link/` exits 7, could not connect |
| Port 80 | serves the console and visitor traffic | **200** |
| ssh ingress 2222 | serves | **`SSH-2.0-pumasi-tunnel`** |
| TCP range in use | `-tcp-low`/`-tcp-high`, no default | **20000–20099** — `"tcp_range"`, unchanged, below the ephemeral floor (§3.1a) |
| A URL is serviceable when it is announced | **yes, on both paths** — TCP since `1d9505c`, HTTP since `fd523e8` | **no, on either** — the running binary predates both |
| **Does a name belong to anybody** | **yes, to a token** — `--subdomain N --token T` claims N and its port, held across a disconnect, refused to everyone else (`4489fbe`) | **no** — any anonymous agent may take any free name, including one somebody is using between reconnects |
| **Does the status view report who owns a name** | **yes** — `relay/dashboard.go:42`, a `"reserved"` bool on every tunnel | **no such key on either tunnel**, which is how this pass dated the running binary (below) |
| Does a reservation survive a relay restart | **no** — in memory only; `spec/0004` slice 2, `BACKLOG.md` item 2 | **no** |
| Suite determinism | **0 failures in 425 runs** across three arms (§2) | not applicable — nothing runs a suite there |
| Live tunnels | — | **2** — `"count":2`, and the second is not this project's (below) |
| Which build is answering | — | **still not stated by any surface** — two accidental discriminators, no version; §5 |

So: **a person who opens a tunnel today is still handed an `https://` address
that nothing answers, and a name that belongs to nobody.** Every cell in the
right-hand column was re-taken this pass, not carried: `GET
http://pumasi.link/_pumasi/status` at 05:51:04 UTC returns
`"url":"https://sshsteward.pumasi.link"`, `https://pumasi.link/` does not
connect, `http://pumasi.link/` answers `200`, and `pumasi.link:2222` still
greets with `SSH-2.0-pumasi-tunnel`. The relay process on the Vultr host
predates `83fd9f7` and will keep announcing `https://` until it is restarted.

**Eight commits touching the Go tree now sit on `main` that the host does not
have, and six of them change non-test code** — up from five and three at the last
pass, counted on the same basis and re-derived here by `git rev-list
83fd9f7~1..9e2de66` filtered on `*.go`, not carried:

| commit | what it changes | user-visible once deployed |
| :--- | :--- | :--- |
| `83fd9f7` | the announced scheme | **yes** |
| `3480990` | bind the public port before announcing it | **yes** |
| `e40a224` | an acceptance fixture | no |
| `b3d251d` | the TCP harness's port block | no |
| `fd523e8` | the mux session exists before the URL is announced | **yes** |
| `4489fbe` · `c12d11a` · `20e9d57` | a name and its port belong to a token and survive a disconnect | **yes** — one capability, three commits |

**Four distinct behaviour changes, up from three.** The two test-only merges are
counted because they wait on the same restart and named as harmless because a
count that does not distinguish them would overstate what a deploy would deliver.
And **three of the six non-test commits are mutually invisible from outside** —
nothing a visitor or an operator can read distinguishes `4489fbe` from `20e9d57`,
which is §5's point arriving with a fourth capability behind it.

**How this pass dated the running binary without touching the host, which is
worth recording because it is the second time a diagnostic accident has done the
job a version number should do.** Slice 1 adds a `"reserved"` key to every tunnel
in the status view — `relay/dashboard.go:42`, a plain `bool` with no
`omitempty`, so it is present on every tunnel of every relay built after
`4489fbe`. The 05:51:04 UTC read carries **no `reserved` key on either tunnel**.
So the host predates `4489fbe`, and the scheme in the `url` field puts it before
`83fd9f7`. **Two accidents, from two unrelated changes, neither of which will
recur for the next merge.** §5.

**Both tunnels are still there, and the second is still unattributed.** The
status read above reports `"count":2`. Alongside `sshsteward`
(`pumasi.link:20000` → this machine's port 22, `"fixed":true`,
`"opened_at":"2026-08-31T06:18:13Z"`, **`"age_secs":84771` — 23 h 32 m
unbroken**, the same connection as at the last three passes) there is
**`skk6g7tyrs`**, `pumasi.link:20002` → a `"local_port":3389`, `"fixed":false`,
opened `2026-09-01T01:48:23Z` and **4 h 02 m** old at this read — the same
`opened_at`, so one connection persisting rather than a series of new ones.
**This seat did not establish who opened it and will not guess.** It is recorded
because **a restart still costs two people their address rather than one** —
which is the fact `Q-014` is built on, and that entry's text still describes the
live set as *"exactly one"*.

**One reading of it has to be corrected now that slice 1 exists, and this file
would otherwise carry it forward.** The previous pass called this tunnel
*"`AllowAll` doing what §4's fact 2 says it must"*. **Slice 1 does not stop it
and was never meant to.** `spec/0004` §6's third column keeps a tokenless
request working on an *unclaimed* name, and frozen case **C-3** exists to go red
if anyone ever removes that. What slice 1 refuses is a **claimed** name to
someone who cannot prove they claimed it. So `skk6g7tyrs` is the anonymous case
working as specified — on a relay that in any event does not have the change —
and not a gap the merge left open. It remains the first thing on this relay that
is not the steward's own route, and it is still not evidence of adoption (§7).

**Why it has not been restarted, stated rather than left unsaid.**
`pumasi/DECISIONS.md` **Q-014** is open: it asks who may restart a relay that
keeps no durable state and whose live tunnel was, when the entry was written,
`sshsteward` alone — `pumasi.link:20000` → this machine's port 22,
`pumasi-ops/RESOURCES.md` §4's remote-access route, **open 84 771 s (23 h 32 m)
and unbroken** at the measurement above, the same connection rather than a new
one. As of this pass it is still not alone; see above. Q-014 is **explicitly outside CHARTER Part 0's
proceed-on-default rule**. This evaluation therefore did not deploy, did not
treat the deploy as a judgement call, and **does not ask a coder packet to take
it either**. It is named as a blocker in `BACKLOG.md` item 1(i) and it waits for
the steward.

**Two things this pass learned bear on that entry's premise, and both were
appended to it as evidence rather than acted on.** *(a)* Q-014's own *What
retires this entry* row names **durable registry and port reservations** — which
is `spec/0004` **slice 2**, `BACKLOG.md` item 2, and explicitly **not** the slice
that merged; `spec/0004` §4 says in as many words that *"saying slice 1 retires
Q-014 would be false."* *(b)* Slice 1 does narrow the premise — a restart costs
the route only if somebody takes the name or the port during the reclaim window,
and ownership is what prevents that — **but only once deployed**, and deploying
is the act Q-014 gates. Until then the steward's route is exactly as exposed as
it was before the merge. **And on the day of a deploy it would still gain
nothing**: `pumasi-ops/tools/pumasi-tunnel-keepalive.sh` passes no `--token`, so
`sshsteward` would be unclaimed. That is `pumasi-ops`'s file, it is named in
`BACKLOG.md` item 1(i) as a precondition, and it is handed up rather than reached
for.

---

## 2 · Maturity gates

| Stage | Criteria | Status |
| :--- | :--- | :--- |
| **0. Candidate** | Scored by 3 model families (40/60) and selected by steward | **COMPLETE** |
| **1. Alpha** | Pure-core suite passes 100%; both public landing surfaces live | **COMPLETE 2026-08-31** — **0 failures in 425 runs** at `9e2de66` on this machine across three arms, plus **3 of 3 `GATE: PASS`**, and no identified defect behind the qualification; what is left of it, including a property with no test at all (§3.3), is stated below |
| **2. Beta** | Real end-to-end users complete workflows without engineer intervention | **IN PROGRESS** — §4 |
| **3. Launched** | Production hardening, cross-model regression, 7-day veto window | PENDING |

### The Stage 1 gate, re-read at `9e2de66` — same qualification, a larger suite behind it

**The short answer to the question this evaluation was asked.** Job `0081`
reported *"the suite green on three consecutive runs plus `-race`"* at
`9e2de66`. That is a builder's self-report; this is the seat that measures, and
it re-ran the arms rather than carrying them. **The gate stays `MET` and stays
qualified, and the qualification has not changed shape since the last pass** —
it is the limit of the evidence, not a named defect.

**What was measured, by this seat, on this machine, at `9e2de66`, this pass.**
Run counts are given because a gate whose number is inherited rather than
measured is what produced the wrong 12-run reading that raised **Q-024**
(rider (b)).

| Arm | Runs | Failures |
| :--- | ---: | ---: |
| `go test -count=1 ./...` — the command the gate names | **300** | **0** |
| `go test -count=1 -cover ./...` | **100** | **0** |
| `go test -race -count=1 ./...` | **25** | **0** |
| `pumasi/tools/gate.sh`, whole gate, `SKIP_FAMILY_PROBE=1` | **3** | **3 × `GATE: PASS`** |

**425 full-suite executions at this SHA, 0 failures**, 05:49:49–06:10:19 UTC,
plus three whole-gate runs at 06:10. **The working tree was clean throughout** —
`git status --porcelain` empty at the start of the run, which is a change from
the last pass: the unowned `cmd/pumasi-tunnel/main.go` diff that sat in the tree
under every previous measurement was **stashed** by job `0081`
([`BACKLOG.md`](BACKLOG.md), *Not on this list*), so these are the first figures
this file has published from a tree that is exactly the SHA it names.

**And the whole gate passes at this SHA, which the last pass could not claim.**
`9e2de66` is a reviews commit carrying a `Spec:` trailer, so `gate.sh` reaches
step 4 and prints `GATE: PASS`; the previous pass's `GATE: FAIL` was
`missing trailer: Spec:` on a README commit and had nothing to do with the suite.
Three runs, three passes. Breadth is `SKIPPED` and therefore **UNVERIFIED** on
each, which the gate prints itself and this file repeats rather than rounding up.

**The host did not get quieter to produce these figures — it got busier.**
`/proc/sys/net/ipv4/ip_local_port_range` is still **32768 60999**, four
`workerd` processes were running throughout, and `ss -tan` found **19** sockets
inside 34500–34599 — the block §3.1b is about, and more than double the 9 the
last pass found. It found **no listening socket from any other process** inside
21000–21039 or 20500–20559, the two blocks the suite actually binds.

**Why the gate is still qualified, and what the qualification now is.** It is no
longer *"the number cannot be measured"*, and it is no longer *"one known
unfixed ordering defect on the HTTP path"*. Both of those are closed. What is
left has three parts and not one of them is a bug in this repository:

1. **Every figure this gate has ever had was taken on one machine, by the seat
   that wrote the file.** Nothing re-runs it — no CI, no second host, no other
   party. That is `Q-025`'s question arriving through this file's front door, and
   `Q-024`'s rider (c) already says in as many words that this is weaker than CI
   and is to be written down as weaker rather than as absent.
2. **A green suite is not evidence that the code is right, and this pass has the
   sharpest demonstration of that the project has produced.** Slice 1 was merged
   with the suite green and a code review approving the diff — and **three
   defects were found afterwards**, the third introduced by the fix for the
   second (`spec/0004` §11, §12, §13). What established each of them was a
   **mutation**: reinstate the defect, run only the case that names it, and
   confirm it goes red. `spec/0004` §10's table is now twelve mutations covering
   every built case, and under the over-broad rollback guard exactly **one** case
   in the whole suite reddens. **The run counts above say the change broke
   nothing. They do not say it is correct, and this file will not let them be
   read that way.**
3. **What the number covers has a hole in it that is new this pass** — §3.3.
   The property `BACKLOG.md` item 2 exists to deliver has **no test in this
   repository**: `relay.TestAReservationOutlivesTheRelay` is specified in
   `CASES.md` and absent from the tree. So 425 green runs are not evidence about
   the restart half in either direction.
4. **One latent fixture defect is still in a frozen file** — §3.1b, `BACKLOG.md`
   item 9 — and it is blocked on a governance reading rather than on work (§8).
   It binds nothing today and this pass re-measured that against a **busier**
   host than last time (19 contended sockets in its block, against 9); the day it
   binds, it stops being latent silently.

**What this does and does not do to `Q-024`.**

- It **does** complete both halves of that entry's own stated retirement
  condition, and now does so at a second SHA. The named fix — bind before the
  response leaves, on both the agent and the ssh paths — landed at `1d9505c`. The
  *"40 clean runs of each invocation recorded in `roadmap/STAGE.md`"* exist here
  at 300 and 100, plus 25 with `-race`, on a tree eight Go-tree merges further on
  and with 989 more lines of test in it. And the successor defect that an earlier
  pass gave as its reason for not retiring the entry — the non-determinism having
  *moved* to the HTTP path rather than gone — is itself merged.
- It **does not** retire it, and this pass does not attempt to. **Closing a
  window is the steward's act and never the seat that records the evidence**,
  and this pass is precisely the seat that recorded the evidence. The window is
  not closed, not dated, not extended, not softened, and no default is read off
  it here. The evidence is written into `pumasi/DECISIONS.md` under that entry in
  its own idiom and stops there.
- **The gate therefore stays `MET` and stays qualified**, under Q-024's own
  named default, and the qualification is: **the pure-core suite passed 425 of
  425 on one machine at this SHA, plus 3 of 3 whole-gate runs, with no identified
  defect open behind the number and nothing but this seat re-taking it.** That is
  the same claim as last pass with a different SHA under it and one more arm — and
  §3.3 now bounds it further: **it is a claim about the cases that exist.**
- **Q-024's rider (a) still binds and its force is undiminished by the better
  number.** No stage-promotion announcement may be published off this gate —
  `STAGE_PLAYBOOK.md` **Event 3 stays held**, §7 — and nothing public may quote
  the gate's figure. Rider (c) is why: evidence getting stronger is not a
  promotion ground, and evidence strength moving is nearly the whole of this
  pass again.

**Coverage, re-measured at `9e2de66` over the 100 `-cover` runs.**

| Package | At `87244af` (last pass) | At `9e2de66` (this pass) |
| :--- | :--- | :--- |
| `core` | 80.3% | **84.1%** |
| `mux` | 84.0% | **84.0%** |
| `relay` | 83.3% | **85.7%** |
| `agent`, `cmd/pumasi-relay`, `cmd/pumasi-tunnel` | 0.0% | **0.0%** |

`core` gains **3.8** points and `relay` **2.4**, from slice 1's 989 lines of new
test in four files — a change improving the evidence for the code it touched,
which is the ordinary case. `mux` is unmoved and nothing this pass touched it.
**`agent` is still 0.0% and the merge that just landed did not change that** —
`git diff --stat 1853218..9e2de66 -- '*_test.go'` reports four files and none of
them is in `agent/`, which is the **second consecutive merge** to make that point
and the first to make it against `agent/`'s own stated deadline (§3.2).

**Both surfaces are live, re-measured 06:11 UTC 2026-09-01.**
- *Surface A, the commons catalog*: `pumasi-web`
  `content/products/pumasi-tunnel.md` (`c2084a8`, 2026-08-30);
  `https://pumasi.ai/products/pumasi-tunnel/` → **200**, and the entry appears
  **once** in `https://pumasi.ai/llms.txt`. **But not `pumasi/catalog.json` — see
  §6.**
- *Surface B, the product's own domain*: `http://pumasi.link/` → 200, serving
  the console (`relay/dashboard.html`, `b3585f6`).

**And the product carries real traffic.** The `sshsteward` tunnel in §1 —
**23 h 32 m unbroken** at 05:51:04 UTC, carrying this machine's own sshd across
`pumasi.link:20000`. That is `RESOURCES.md` §4's remote access path, working,
and it remains the strongest evidence this product has. **A second tunnel was
open beside it at the same read** (§1), forwarding somebody's `3389`; it is
recorded, it is not attributed, and it is not counted as evidence of anything
until somebody can say whose it is.

---

## 3 · What the gate's number does not cover — and this pass adds a flag rather than closing one

The Stage 1 gate is `go test` exiting 0. The gate's number is only as good as
what it covers (L-006), so what it does not cover is recorded here rather than
left implicit. Every entry below is a `BACKLOG.md` item. The flag this section
led with until the last pass — an announce-ordering defect — is closed on both
paths, and §3.1 is the record of that rather than a live warning. **§3.3 is new
this pass, and it is the sharpest instance of this section's own premise that
the file has carried: the property `BACKLOG.md` item 2 exists to deliver has no
test in this repository at all, so the suite cannot go red on it and a green run
is not evidence about it in either direction.**

### 3.1 · Both announce-ordering defects are closed, on both paths

**Retired by merge, and verified in the tree by this seat — at `87244af` when
this section was written and re-read at `9e2de66`, where slice 1 moved the line
numbers below without changing the ordering they describe.** `relay.ServeAgent` now registers the tunnel,
**binds** the public TCP port if there is one (`relay/relay.go:186` — `0041`'s
fix, `spec/0002`), then builds the mux session, **takes `r.mu`, installs
`r.sessions[resp.AgentID]` and writes the auth response inside that one critical
section** (`:239`–`:241`), deleting the session again in the same section if the
write fails (`:245`). `ServeHTTP` (`:450`) takes the same lock to read
`r.sessions`, so a visitor arriving in what used to be the window waits for the
lock instead of being answered `404 No tunnel is open`. That is `fd523e8`,
`spec/0003-session-before-announce`, and it closes the flag this section carried
two passes ago. **Every line number in this paragraph moved with slice 1 and was
re-read at `9e2de66`; the ordering they describe did not move.**

**The fix is not the one this project predicted, and the difference is the part
worth keeping.** The previous order's `BACKLOG.md` item 2 said the insert
*"can simply move above"*
the announce. It cannot: the announce is written raw (`relay/relay.go:438`) while
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
listener is the relay's job" (`core/portpool.go:21`–`:22`, re-read at
`9e2de66`). When the number it picks turns out not to be bindable,
`relay.listenTCP` returns the port and an error (`relay/tcp.go:74`) and the relay
refuses the tunnel outright — it does not ask for the next of the 99 other free
ports. **Slice 1 changed both files and left this untouched**, which this seat
established by reading the changed code rather than assuming
([`BACKLOG.md`](BACKLOG.md) item 3 records the correction, because that entry
had carried the sentence *"nothing since `0047` changed"* these files). One busy port defeats a
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
`core.NewPortPool` (`relay/relay.go:135`, moved from `:124` by slice 1) — a type
that *"does no I/O"* (`core/portpool.go:21`–`:22`). It binds nothing. `ss -tan` found **19** sockets
inside 34500–34599 during this pass — more than double the 9 the last pass
found, several of them `ESTAB` rather than merely lingering — and the suite still
passed every run in §2. The two constants were re-read at `9e2de66`, at the same
two line numbers, and are unchanged. `BACKLOG.md` **item 9** (was item 8; the
substance is unchanged and the number moved with this pass's re-rank), ranked
below everything buildable because what blocks it is a governance reading rather
than work — see §8.

### 3.2 · Three packages have no test files, `agent/` among them

`agent`, `cmd/pumasi-relay` and `cmd/pumasi-tunnel` report *no test files* and
0.0% coverage. `core`, `mux` and `relay` are the pure core the gate names and
they are covered. `agent/` is not core, but it is half of every tunnel, and
today nothing exercises it except the relay's end-to-end tests using it as a
fixture. Re-measured at `9e2de66` over 100 `-cover` runs: still **0.0%**, while
`core` rose to 84.1% and `relay` to 85.7%.

**Two consecutive merges have now made this flag's point, and the second one
made it against the flag's own deadline.** `fd523e8` put all **468** of its new
test lines in `relay/sessionorder_test.go`. Slice 1 put all **989** of its new
test lines in `core/reservation_test.go`, `core/portpool_test.go`,
`relay/reservation_test.go` and `relay/discardclaim_test.go` — `git diff --stat
1853218..9e2de66 -- '*_test.go'` reports four files and **zero** in `agent/`. And
slice 1 is precisely a change to what happens **across a reconnect**, which is
the behaviour this entry said should be tested *before* it was rewritten. It was
rewritten first. [`BACKLOG.md`](BACKLOG.md) **item 7** — *`agent/` has no tests*
(was item 5; it moved down one this pass because the deadline in its why-here has
passed, not because the gap shrank). **Fixer: the coder.**

### 3.3 · The property item 2 exists to deliver has no test in this repository

**New this pass, and it is the reason a green suite must not be read as progress
on the restart half.** `spec/0004/acceptance/CASES.md` specifies **D-1** — *a
reservation outlives the process*: a relay with a store takes a claim, is shut
down, and **a second `relay.New` is built over the same path**, not a reconnect.
It is written down deliberately, at the freeze, so slice 2 would be measured
against a case that existed before its implementation did.

**It is not in the tree.** `grep -rn "func TestAReservationOutlivesTheRelay("
--include=*_test.go .` returns **0** at `9e2de66`, against **1** for each of the
other nineteen cases the same file names (R-1..R-8, P-1..P-2, C-1..C-9 — checked
one function name at a time by this seat).

**So the wording matters, and one sentence written elsewhere needs correcting.**
`pumasi-ops/DIGEST.md` describes D-1 as *"written and left red so no reader of a
green suite concludes the restart half shipped."* It is written **in
`CASES.md`** and **absent from the Go tree**; nothing in this suite is red, and
nothing was red in any of this pass's runs. `CASES.md`'s own text is correct —
*"specified and not built by this packet"*. The distinction is this section's
whole subject: **a case that does not exist cannot fail, which is why the gate's
number says nothing about the property, and why the evidence that the restart
half is outstanding is the missing function rather than a failing one.**
[`BACKLOG.md`](BACKLOG.md) **item 2**.

---

## 4 · Why not `beta`

`beta` means strangers can rely on it and their data survives. **Each fact
below was re-verified against the tree at `9e2de66` in this pass**, by grep and
by reading, not carried. The order is the backlog's.

**One of the three changed this pass, for the first time since this section was
written — and it changed on `main` only.** Fact 2 was *a name belongs to
nobody*. On `main` that is now false: a name and its public TCP port belong to
whoever first proved they held them. On `pumasi.link` it is still true in every
word. Fact 1 and fact 3 are untouched. **The distance to `beta` therefore
shortened by exactly nothing for any user**, and this section says so below
rather than at the end, because a stage file that reports a merge as the
user-visible truth is the L-009 failure this fleet has already met twice (§1).

1. **There is no TLS, and what is running still says there is.** *Changed by
   the release, and only halfway.* On `main` the relay now announces the scheme
   it serves — `-public-scheme`, default `http`, validated once at startup,
   applied in one place, read by all three surfaces. **On `pumasi.link`, nothing
   listens on 443 and the relay still prints `https://`** (§1). Even once
   deployed, every HTTP tunnel here is plaintext and no `https://`-only webhook
   sender can be pointed at one. [`BACKLOG.md`](BACKLOG.md) item 1 — **operator
   action plus a deploy blocked on Q-014, not a build.**
2. **A name belongs to nobody — *false on `main` since `4489fbe`, true in every
   word of what `pumasi.link` serves*.** This is the fact that moved, and it is
   split rather than ticked.
   **On `main`, re-verified at `9e2de66`:** `core.Reservations` records a name,
   a `sha256` of its bearer token and its public TCP port; `Relay.authorize`
   consults it on the one path both ingresses share (`relay/relay.go:333`,
   `:341`); a claimed name is refused to any other token and to a tokenless
   caller; the port pool has a third state, `held`, so a tenant's number is not
   handed out between its connections. `Tunnel.Reserved` — written at what was
   `relay/relay.go:297` and is now `:365`, and at every previous pass **read
   nowhere at all** — is now read at `relay/dashboard.go:80` and
   `relay/relay.go:429`, and is set from `Reservations.Get(name)` rather than
   from the shape of the request. The line that kept moving (236 → 249 → 274 →
   297 → 365) has stopped being the interesting fact about that field.
   **On `pumasi.link`:** none of it. The status view carries no `"reserved"` key
   (§1), so the running binary predates the change. Any anonymous agent may still
   take any free name there, including one another person is using between
   reconnects.
   **What has not changed on either.** The relay binary still defines **eleven
   flags and no auth flag** — `grep -n auth cmd/pumasi-relay/main.go` finds
   nothing — so `AllowAll` (`relay/relay.go:40`, installed at `:104` when
   `cfg.Auth` is nil) is still the only authenticator it can run. A token proves **continuity, not permission** (`spec/0004` §3):
   it narrows *which name* an accepted agent may have and does not decide whether
   the agent is accepted. And an anonymous agent taking a **free** name is
   specified behaviour that slice 1 deliberately preserves (case **C-3**) — so
   the unattributed `skk6g7tyrs` tunnel is not this fact any more, and §1
   corrects the previous pass's reading of it.
   [`BACKLOG.md`](BACKLOG.md) *Delivered*, and item 5 for the zero-install path
   this does not reach.
3. **Nothing survives a restart.** Still true, and **unchanged by this pass** —
   which is the half of item 2 that did not merge. The registry
   (`core/route.go:145`–`:147`), the port pool (`core/portpool.go:38`–`:41`,
   which slice 1 grew from two maps to three) and now the reservation set
   (`core/reservation.go`) are all plain in-memory maps.
   *Verified at `9e2de66`:* no store — `grep -rn
   "os.WriteFile\|os.Rename\|os.Create\|encoding/json"` over `core/` and
   `relay/` excluding tests returns two hits, both wire encoding; no `LastSeen`;
   no `-reservations` flag; and **the acceptance case for the property does not
   exist in the tree** (§3.3). A relay restart drops every subdomain, every
   reservation, every reserved TCP port and every live tunnel — including the one
   carrying this machine's remote access, which is precisely why Q-014 exists —
   and as of 05:51:04 UTC it would drop **two** tunnels rather than one (§1).
   `--tcp-port` keeps an address across an *agent* reconnect (`a5b77fc`) and,
   since `4489fbe`, keeps it **against other callers** too; it still does not
   keep it across a relay one.
   [`BACKLOG.md`](BACKLOG.md) item 2 — *the relay-restart half*.

**In the order that shortens the distance fastest**, which is the backlog's
order and its reasoning in one line each — and **this is §4's cost-to-move,
changed by this pass**: **item 2** first — *the relay-restart half — a
reservation that outlives the process* — because it is what is left of fact 3
and because it is **Q-014's own written retirement condition**, that entry's
*What retires this entry* row naming *"durable registry and port reservations"*
and not the ownership slice that merged. Then **item 3**, the relay's own
bindability gap, bounded in production today only because the deployed range
happens to sit below the ephemeral floor. **Item 1** runs in parallel and needs
no coder, only a decision and a certificate.

**What this pass's merge bought, stated exactly, because it is the first thing
this section has had to size that was not a disappointment.** It closed fact 2
**on `main`**. It did not close fact 3, does not retire Q-014 (`spec/0004` §4
says so in as many words), and reaches no user, because it is one of four
behaviour changes `pumasi.link` does not have. **The distance to `beta` did not
shorten**, and this file will not pretend otherwise: `beta` is what a *stranger*
can rely on, and the stranger's relay is unchanged since before 2026-08-31
10:47.

**And the residual is smaller than the entry it came from — this file will not
carry the old size forward either.** The previous pass described the distance as
*one large piece of work*. Slice 1 was the larger half of that work and it is
merged. What is left is a store behind an existing type, one flag, and three
questions the spec says need a reviewer before code. **So the honest description
of the distance to `beta` is now: one medium piece of work on `main`, plus a
deploy that only the steward can authorise, plus a certificate nobody has
installed** — and the second and third of those are not a coder's to take.

**Also gating, from `PRODUCT-RULES.md`** (v1.0, read fresh 2026-09-01; still
only on the unmerged `worktree-product-rules` branch, `0115758` — **Q-017** —
and its absence from `main` is not compliance): **PR-1** (a user-visible version
number) binds **always** and this product has none anywhere — which §5 shows is
no longer theoretical and got worse again this pass; **PR-2** (in-app feedback)
binds at the `beta` promotion and is unbuilt. `BACKLOG.md` items 6 and 8.
**Q-017 re-checked this pass, and it is the tenth consecutive evaluation to
report the same thing:** `PRODUCT-RULES.md` is **still not reachable on `pumasi`
`main`** — `git ls-tree` at `pumasi` `196b749` finds it on neither `main` nor
`origin/main`, only on `worktree-product-rules` (`0115758`). Read fresh from that
branch for this evaluation, as duty 1 of the role file requires. Its absence from
`main` is not compliance, and this seat neither merged it nor proposed merging it
— that is Q-017's own question and it is the steward's.

## What `launched` additionally requires

Stage 2's exit gate — real end-to-end users completing workflows without an
engineer — plus production hardening. Not enumerated further while §4 is open.

---

## 5 · The version gap stopped being theoretical this week

`PRODUCT-RULES.md` PR-1 asks for a version that moves and is user-visible.
There is a build on `main` that behaves differently from the build on the host,
and **no surface of this product will tell you which one is answering** — not the
console, not `/_pumasi/status`, not the logs. **This is the fifth consecutive
evaluation to date the running binary by inference rather than by reading it.**
[`BACKLOG.md`](BACKLOG.md) item 6 — *PR-1 compliance: a version that moves and
is user-visible*, moved **up one** this pass. **Fixer: the coder** —
`core.AuthRequest.ClientVersion` is already in the wire protocol
(`core/handshake.go:33`) and no binary of this product sets it; the only writer
is `relay/sshingress.go:169`, filling it with the *ssh client's* version string.

**It got worse this pass in the way that matters, and slightly better in a way
that must not be mistaken for progress.**

*Worse:* **eight** merges touching the Go tree are now absent from the host, of
which **six** change non-test code and deliver a **fourth** distinct capability
(§1). The gap between what `main` does and what a stranger meets is the widest it
has been.

*Apparently better, and it is an accident:* this pass gained a **second**
discriminator. Slice 1 puts a `"reserved"` key on every tunnel in
`/_pumasi/status` (`relay/dashboard.go:42`), and the live read has none, which
dates the host as pre-`4489fbe` — where the scheme in the `url` field only dated
it as pre-`83fd9f7`. **Two accidents are not a version, and this file will not
count them as coverage.** Neither came from anyone deciding a build should be
identifiable; neither will recur for the next merge; and **three of this pass's
six non-test commits are mutually invisible** — nothing distinguishes `4489fbe`
from `c12d11a` from `20e9d57` on any surface, which is exactly the range in which
this pass's three post-review defects were fixed. A deployer confirming *"the
guard fix is live"* has no way to check.

**And the gap has a second reader now.** The status endpoint shows a tunnel this
seat cannot attribute (§1). Had a version string been on the wire, *"which build
is that agent running"* would be answerable; as it is, neither half of that
connection can be identified from any surface this product exposes.

---

## 6 · Known gaps a user should know about today

- **No TLS.** Every tunnel is plain HTTP, whatever URL the relay printed — and
  the running relay still prints `https://` (§1).
- **No accounts.** There is no signup and no identity — **Q-002** — and none of
  what follows changes that.
- **On `pumasi.link`: no name ownership at all.** Any anonymous agent may take
  any free name, including one you are using between reconnects.
- **On `main`: a name can be owned, and here is exactly what that does and does
  not buy** — because *"ownership"* invites three readings it does not support
  (`spec/0004` §3, §8). A `--token` of at least 16 characters claims a name and
  its public TCP port and holds them across a disconnect. It **does not** prove
  who you are; it **does not** decide whether the relay accepts you at all
  (`AllowAll` still says yes to everyone); and it is a **bearer secret on a
  plaintext connection**, readable and replayable by anyone on the path between
  agent and relay until there is TLS. Trust is **on first use**, so a stranger
  who claims your name before you do has it — a strict improvement on today,
  where anyone may take it repeatedly and forever, and not a solution. **A lost
  token is a lost name**: there is no recovery path, because recovery needs an
  identity to recover to and there is none.
- **The zero-install `ssh -R` path can hold nothing.** An ssh client cannot
  present a token, so it can be refused a name somebody else has claimed and can
  never claim or reclaim one (`BACKLOG.md` item 5).
- **Nothing survives a relay restart**, on `main` or on the host — not the name,
  not the reserved port, not the tunnel (§4, fact 3).
- **One relay, one host** (`RESOURCES.md` §3: Vultr, Chicago, ~$5–6/month). A
  restart or a host failure ends every tunnel — **two of them, as of 05:51:04 UTC
  2026-09-01** (§1). Tailscale is kept as the independent fallback for reaching
  `m-gtr`, deliberately.
- **The commons index does not know this product exists.** `pumasi/catalog.json`
  contains **zero occurrences of the string `tunnel`** — no `products[]` entry,
  no `items[]` entry — re-verified this pass at `pumasi` `196b749`, `grep -c
  tunnel catalog.json` = **0**. `README.md` tells every
  arriving agent to start with that file and treats it as the charter's
  duplication check, so an agent running that check today is told Pumasi Tunnel
  does not exist. **Recorded here and deliberately not fixed:** it is not this
  repository's file and **no role file owns it** — `pumasi/DECISIONS.md`
  **Q-019**, open, whose named default would give first registration to the
  marketing manager and ongoing `status`/`maturity` upkeep to this seat. Until
  that resolves, this is the honest place for it.
- **One merged change may owe a release note and does not have one.**
  `pumasi/releases/` at `196b749` carries a single tunnel note, for `83fd9f7`.
  Under `pumasi/DECISIONS.md` **Q-034**'s default an *ordinary* merge owes
  nothing, so `1d9505c` and `fd523e8` are not gaps; what is live is whether
  **slice 1** is *can-hurt* under CHARTER §2.1, which it has a real argument for
  (a bearer secret on a plaintext wire, and requests that were accepted before
  now refused). **Neither the classification nor the note is this seat's** —
  `BACKLOG.md` *Not on this list* states the reasoning and the routing.
- **The local request inspector on `127.0.0.1:4040` does not exist** — `web/` is
  an empty directory, re-checked at `9e2de66`. `VALUE.md` claim 5 says so, and
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
  pass ran on — the same machine where the suite failed **every** run three passes
  ago — it passed **425 of 425** across three arms plus **3 of 3** whole-gate runs
  (§2). What remains: one acceptance case still names 34500–34599 but binds
  nothing (§3.1b, `BACKLOG.md` item 9); one specified case has no Go test at all
  (§3.3, item 2); and nothing but this seat, on this machine, has ever taken any
  of these numbers (§2).

---

## 7 · For the marketing manager, from this evaluation

None of this is this seat's to write, and all of it is a page that contradicts a
file. Job `0066` took the product's own `README.md` on 2026-09-01 and this
section is re-stated against what is left.

1. **`pumasi-web`'s lead sentence "There is no hosted relay" is false**, and has
   been false at four consecutive evaluations. Re-measured 05:51:04 UTC
   2026-09-01: `pumasi.link` answers `200` on 80, greets
   `SSH-2.0-pumasi-tunnel` on 2222, and is carrying **two** tunnels, one of them
   unbroken for 23 h 32 m. The hosted relay is on `pumasi.link`, not
   `tunnel.pumasi.ai`. **`0066` filed this as a `priority: high` hand-off for the
   commons marketing seat and it has not been taken; this is the third pass to
   hand it up.**
2. **The gate table in §2 has changed again** — the coverage figures moved
   (`core` 80.3% → **84.1%**, `relay` 83.3% → **85.7%**, `mux` unmoved) and the
   whole gate now passes at this SHA where it could not at the last one. A page
   that quotes an older figure is quoting something this file has withdrawn. See
   item 5 before quoting the new ones: the answer is still *do not*.
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
   same goes for the ordering fixes and — new this pass — for name ownership.**
   **Four** merged behaviour changes now sit on `main` that `pumasi.link` does
   not have (§1), across eight Go-tree commits. The distinction between merged
   and served is the whole of this file's §1 and it has **widened**, not narrowed.
   **In particular: no page may say this product gives you a name that is yours.**
   On `pumasi.link` it does not, and `VALUE.md` claim 2 now states the split
   explicitly — that is the wording to take, and it is the only wording.
5. **`STAGE_PLAYBOOK.md` Event 3 is still held — and this is the first pass at
   which one of §4's three reasons actually moved, which is exactly when the hold
   matters most.** That trigger chains a stage-promotion announcement and a
   public badge update off a product manager confirming an exit gate `MET`. This
   evaluation kept Stage 1 at `MET` on runs it took itself — **0 failures in
   300 / 100 / 25**, plus 3 of 3 whole-gate runs (§2) — and closed §4's fact 2
   **on `main`**. A previous pass would have read that as licence. **It is not,
   and nothing public may be published off this gate.** Three reasons, all
   binding: **Q-024 is open and unretired**, and its rider (a) holds regardless of
   the rate; rider (c) says evidence getting stronger is not a promotion ground;
   and **§4's fact 2 moved on `main` and not for any user** — the stage ladder
   asks what a stranger can rely on, and the stranger's relay is unchanged since
   before 2026-08-31 10:47.

   Any page still quoting *"passes, 37 times in 40"*, the 0-in-40, the 0-in-500,
   or *"one known unfixed ordering defect on the HTTP path"* is quoting something
   this file has withdrawn. The current figures are the three in §2 and they
   carry their run counts. Nothing is lost by waiting: this product is `alpha`,
   is not asking to move, and §4 lists three separate reasons it should not.

6. **What a page may now say that it could not before, stated precisely because
   the temptation to overstate it is real.** Three things. *(a)* **Both
   announce-ordering defects are fixed on `main`** — the TCP one measured at 28
   dial refusals in 2000 before and 0 after (`1d9505c`), the HTTP one closed at
   `fd523e8` and pinned by a frozen case that fails deterministically without it.
   *(b)* The suite is deterministic on a machine where it once failed every run.
   *(c)* **On `main`, a name and its public TCP port can belong to whoever first
   proved they held them, and are refused to everyone else across a disconnect.**

   **None of it may be published as a Stage 1 pass rate, and all of it carries
   the same caveat: none of it is live.** All four behaviour changes are behind
   the undeployed restart in §1. A page may say the project fixed two ordering
   defects, a flaky test suite, and half of the name-ownership gap **on `main`**.
   It may **not** say the product got more reliable or more permanent for anyone
   using `pumasi.link`, quote the gate, imply a promotion, or attach a run count
   to any of it.

   **And three things about (c) specifically must travel with it or it becomes a
   false claim** (`spec/0004` §3, §8): a token proves **continuity, not
   identity** — there are still no accounts; it is a **bearer secret on a
   plaintext wire** until there is TLS; and the **zero-install `ssh -R` path
   cannot use it at all**. A page that says "your subdomain, permanently" is
   wrong on three counts, and one of them is `VALUE.md`'s withdrawn word.

7. **One thing this file cannot help with, offered so nobody publishes it by
   accident.** A second, unattributed tunnel is live on the relay (§1), still the
   same connection, now 4 h 02 m old. It is the first traffic here that is not the
   steward's own route, and it is **not** evidence of adoption, a user, or a
   customer — nobody has established what it is. It is in this file because a
   restart still costs two parties rather than one, and for no other reason.

8. **New this pass, and it is the only *new* hand-off here.** The zero-install
   `ssh -R` path — the differentiator `MARKET.md` §2 cites and `VALUE.md` claim 1
   leads with — **got narrower on `main`**: it can now be refused a name someone
   else has claimed and can never claim one itself. Any page that leads with the
   ssh command should not also promise a stable name in the same breath. It is
   `BACKLOG.md` item 5, it needs an intent statement before a packet, and it is
   handed up here rather than fixed on a page.

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

**A fourth instance was recorded on 2026-09-01 (`pumasi` `196b749`), and it is
not this repository's.** It came from `pumasi-sign` job `0080`, where a rename
tripped a frozen case and the contested part turned out to be not *who may amend*
but *whether the amendment's own account of itself is true*. **This repository
added no instance this pass, and that was checked rather than assumed**:
`spec/0004` §12 records that no frozen case from `spec/0001`–`0003` was touched,
and the range `1853218..9e2de66` changes four test files, all of them
`spec/0004`'s own. So the count of instances rose and this product's exposure did
not.

**This seat proposes and does not decide**, and it has not set a window or a
deadline. Q-030's own text points at *"`pumasi-tunnel` `BACKLOG.md` item 9"*;
that entry was item 8 after the `1853218` re-rank and is **item 9** again after
this one — unchanged in substance, and a third demonstration of why a bare number
is a poor citation into a file that reorders on purpose.

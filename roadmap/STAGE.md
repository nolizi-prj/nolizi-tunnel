# STAGE — Pumasi Tunnel

**Current Stage:** `Alpha` — unchanged by this evaluation, and §4 says what it
would cost to move.
**Stage 1 exit gate:** **MET 2026-08-31**, carried forward, **re-measured at
`9cc9e65` by the seat writing this** — the figures are in §2 with the run count
beside each arm. The merge under evaluation, `9cc9e65`, published **no test
count and no `GATE:` line at all**; §2's figures are the first this tree has
had.
**The qualification changed shape this pass, and for the worse in one
respect.** No identified source of suite non-determinism is open, and the hole
§3.3 named last pass — a specified property with no test — is closed: D-1
exists. What qualifies the gate now is the limit of the evidence — one machine,
one seat, nothing re-running it — **plus a merge of 1,089 lines to `core/` and
`relay/` that no family other than its builder's has read** (§3.3, rewritten),
and one latent fixture defect still in a frozen file (§3.1b). §2 says it in
full.
**The second of §4's three reasons not to be `beta` moved this pass, and again
on `main` only.** `spec/0004-names-with-owners` **slice 2** merged at `9cc9e65`
— by a hand run outside the queue, after the dispatcher had marked its packet
`FAILED` — so fact 3, *nothing survives a restart*, is **no longer true on
`main`** for a relay started with `-reservations`, and is still entirely true
of what `pumasi.link` serves. Fact 1 is untouched. **The stage does not move**,
and §4 says why in the same words it used before: nothing deployed, so no user
got anything.
**This pass adds evidence under `Q-024` and under `Q-014`, and retires
neither.** Q-014's written retirement condition is **met on `main`** and unmet
on the host (§1). Closing that window, dating it, or reading a default off it
is the steward's act and never the seat that records the evidence — and this
pass is precisely that seat. `STAGE_PLAYBOOK.md` **Event 3 stays held** (§7).
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

**Re-evaluated 2026-09-01 at `9cc9e65`** (job `0093`; post-`9cc9e65`, the
slice-2 merge — two commits, `41f373c` → `9cc9e65`, landed by a hand run after
the dispatcher had marked its packet failed), `STAGE_PLAYBOOK.md` Event 2.
**Every number in §2 was taken in this pass, on this machine, at this SHA, and
the number of runs behind each is stated beside it.** Where a fact was taken
from an earlier pass without being re-taken, it says so **in those words** —
*carried, not confirmed* — rather than being presented as fresh.

**Re-evaluated 2026-09-01 at `9e2de66`** (post-`0081`, the names-with-owners
slice-1 merge — eight commits, `1853218` → `9e2de66`), `STAGE_PLAYBOOK.md`
Event 2. Its text is kept below where it still holds.

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

| | on `main` @ `9cc9e65` | on `pumasi.link`, re-measured **18:51:06 UTC 2026-09-01** |
| :--- | :--- | :--- |
| Scheme announced | `http://` by default, `-public-scheme=https` when a terminator is in front | **`https://`** — `"url":"https://sshsteward.pumasi.link"` |
| Port 443 | not the relay's job, by design | **refused** — `curl https://pumasi.link/` could not connect |
| Port 80 | serves the console and visitor traffic | **200** |
| ssh ingress 2222 | serves | **`SSH-2.0-pumasi-tunnel`** |
| TCP range in use | `-tcp-low`/`-tcp-high`, no default | **20000–20099** — `"tcp_range"`, unchanged, below the ephemeral floor (§3.1a) |
| A URL is serviceable when it is announced | **yes, on both paths** — TCP since `1d9505c`, HTTP since `fd523e8` | **no, on either** — the running binary predates both |
| Does a name belong to anybody | **yes, to a token** — `--subdomain N --token T` claims N and its port, held across a disconnect, refused to everyone else (`4489fbe`) | **no** — any anonymous agent may take any free name, including one somebody is using between reconnects |
| Does the status view report who owns a name | **yes** — `relay/dashboard.go:42`, a `"reserved"` bool on every tunnel | **no such key on the one tunnel it reports**, which is how this pass dated the running binary (below) |
| **Does a reservation survive a relay restart** | **yes, when the relay is started with `-reservations <path>`** — a second `relay.New` over the same path finds the name *and* its port held (`9cc9e65`; D-1, `relay/reservation_test.go:640`). With the flag empty, **no**, exactly as before (D-7) | **no** — the running binary has no store and no flag |
| Suite determinism | **0 failures in 425 runs** across three arms (§2) | not applicable — nothing runs a suite there |
| Live tunnels | — | **1** — `"count":1`; the unattributed second tunnel of the last two passes is gone (below) |
| Which build is answering | — | **still not stated by any surface** — two accidental discriminators, no version, and the newest merge has none; §5 |

So: **a person who opens a tunnel today is still handed an `https://` address
that nothing answers, and a name that belongs to nobody.** Every cell in the
right-hand column was re-taken this pass, not carried: `GET
http://pumasi.link/_pumasi/status` at 18:51:06 UTC returns
`"url":"https://sshsteward.pumasi.link"` and `"count":1`, `https://pumasi.link/`
does not connect, `http://pumasi.link/` answers `200`, and `pumasi.link:2222`
still greets with `SSH-2.0-pumasi-tunnel`. The relay process on the Vultr host
predates `83fd9f7` and will keep announcing `https://` until it is restarted.

**Nine commits touching the Go tree now sit on `main` that the host does not
have, and seven of them change non-test code** — up from eight and six at the
last pass, counted on the same basis and re-derived here by `git rev-list
83fd9f7~1..9cc9e65` filtered on `*.go`, not carried:

| commit | what it changes | user-visible once deployed |
| :--- | :--- | :--- |
| `83fd9f7` | the announced scheme | **yes** |
| `3480990` | bind the public port before announcing it | **yes** |
| `e40a224` | an acceptance fixture | no |
| `b3d251d` | the TCP harness's port block | no |
| `fd523e8` | the mux session exists before the URL is announced | **yes** |
| `4489fbe` · `c12d11a` · `20e9d57` | a name and its port belong to a token and survive a disconnect | **yes** — one capability, three commits |
| `9cc9e65` | a reservation and its port outlive the relay process, when `-reservations` is set | **yes, only if the deploy sets the flag** — an empty flag is today's relay (D-7) |

**Five distinct behaviour changes, up from four.** The two test-only merges are
counted because they wait on the same restart and named as harmless because a
count that does not distinguish them would overstate what a deploy would
deliver. **Four of the seven non-test commits are mutually invisible from
outside** — nothing a visitor or an operator can read distinguishes `4489fbe`
from `20e9d57`, and `9cc9e65` changes no external surface at all: a relay
running with `-reservations` looks exactly like one without it from anything
this product exposes. That is §5's point arriving with a fifth capability
behind it.

**How this pass dated the running binary without touching the host — the
third time a diagnostic accident has done the job a version number should do.**
Slice 1 adds a `"reserved"` key to every tunnel in the status view
(`relay/dashboard.go:42`, a plain `bool` with no `omitempty`, so it is present
on every tunnel of every relay built after `4489fbe`). The 18:51:06 UTC read
carries **no `reserved` key on the one tunnel it reports**. So the host
predates `4489fbe`, and the scheme in the `url` field puts it before `83fd9f7`.
**Slice 2 supplies no third accident**: `9cc9e65` touches neither
`relay/dashboard.go` nor `dashboard.html`, so on the day it is deployed nothing
readable will say whether it was deployed with the flag. §5.

**The second tunnel is gone, and it was never attributed.** The status read
above reports `"count":1`: `sshsteward` alone (`pumasi.link:20000` → this
machine's port 22, `"fixed":true`, `"opened_at":"2026-08-31T06:18:13Z"`,
**`"age_secs":131572` — 36 h 32 m unbroken**, the same connection as at every
pass since the first). `skk6g7tyrs` — `pumasi.link:20002` → a
`"local_port":3389`, `"fixed":false`, opened `2026-09-01T01:48:23Z` — was
present at 02:48 and 05:51 UTC and absent at 18:51, so it lived at most
seventeen hours. **Nobody here established who opened it, and this file still
does not guess.** It was, as the last pass corrected, the anonymous case working
as specified — `spec/0004` §6's third column keeps a tokenless request working
on an *unclaimed* name, and frozen case **C-3** exists to go red if anyone ever
removes that — on a relay that in any event does not have the change. What its
departure changes is one premise: **a restart now costs one party their
address again, not two**, which is the *"exactly one"* Q-014 was written on.
What it does not change: the relay ran `AllowAll` throughout and would accept
the next one. It was never evidence of adoption (§7).

**Why it has not been restarted, stated rather than left unsaid.**
`pumasi/DECISIONS.md` **Q-014** is open: it asks who may restart a relay that
keeps no durable state and whose live tunnel is `sshsteward` —
`pumasi.link:20000` → this machine's port 22, `pumasi-ops/RESOURCES.md` §4's
remote-access route, **open 131 572 s (36 h 32 m) and unbroken** at the
measurement above, the same connection rather than a new one. Q-014 is
**explicitly outside CHARTER Part 0's proceed-on-default rule**. This
evaluation therefore did not deploy, did not treat the deploy as a judgement
call, and **does not ask a coder packet to take it either**. It is named as a
blocker in `BACKLOG.md` item 1(i) and it waits for the steward.

**Three things this pass learned bear on that entry, and all three were
appended to it as evidence rather than acted on.** *(a)* **Q-014's own *What
retires this entry* row — *durable registry and port reservations* — is met on
`main` at `9cc9e65`.** A relay started with `-reservations <path>` rebuilds its
reservation set and the port pool's holds from the file at boot, and D-1 is a
second `relay.New` over the same path proving it, not a reconnect. It is met
**only when the operator sets the flag**: an empty `-reservations` is today's
relay in every respect (D-7), so a deploy that omits it delivers nothing on
this point. *(b)* **It is met nowhere a user can reach**, and deploying is the
act Q-014 gates — the same circularity as last pass, one slice further along.
Until a deploy the steward's route is exactly as exposed as it was before
either slice. *(c)* **The keepalive precondition got sharper.**
`pumasi-ops/tools/pumasi-tunnel-keepalive.sh` passes no `--token` (`grep -c
token` → 0, re-read this pass), so `sshsteward` would be *unclaimed* after a
deploy; before slice 2 a stranger who took it held it until the next restart,
and **with `-reservations` set they hold it for thirty idle days across every
restart**. That is `pumasi-ops`'s file, it is named in `BACKLOG.md` item 1(i)
as a precondition, and it is handed up rather than reached for. Whether the
condition being met on `main` retires the entry is the steward's reading; this
pass wrote the evidence and set nothing.

---

## 2 · Maturity gates

| Stage | Criteria | Status |
| :--- | :--- | :--- |
| **0. Candidate** | Scored by 3 model families (40/60) and selected by steward | **COMPLETE** |
| **1. Alpha** | Pure-core suite passes 100%; both public landing surfaces live | **COMPLETE 2026-08-31** — **0 failures in 425 runs** at `9cc9e65` on this machine across three arms, plus **3 of 3 `GATE: PASS`**, and no identified defect behind the qualification; what is left of it — including a merge no family but its builder's has read (§3.3) — is stated below |
| **2. Beta** | Real end-to-end users complete workflows without engineer intervention | **IN PROGRESS** — §4 |
| **3. Launched** | Production hardening, cross-model regression, 7-day veto window | PENDING |

### The Stage 1 gate, re-read at `9cc9e65` — same qualification, a larger suite behind it, and a merge nobody else has read

**The short answer to the question this evaluation was asked.** The merge under
evaluation published **nothing**: `9cc9e65`'s message carries no test count and
no `GATE:` line, and job `0089`'s log is one sentence about a usage limit. This
is the seat that measures, and it ran the arms rather than looking for figures
to carry. **The gate stays `MET` and stays qualified**, and the qualification
gained one part and lost one (below).

**What was measured, by this seat, on this machine, at `9cc9e65`, this pass.**
Run counts are given because a gate whose number is inherited rather than
measured is what produced the wrong 12-run reading that raised **Q-024**
(rider (b)).

| Arm | Runs | Failures |
| :--- | ---: | ---: |
| `go test -count=1 ./...` — the command the gate names | **300** | **0** |
| `go test -count=1 -cover ./...` | **100** | **0** |
| `go test -race -count=1 ./...` | **25** | **0** |
| `pumasi/tools/gate.sh`, whole gate | **3** | **3 × `GATE: PASS`** |

**425 full-suite executions at this SHA, 0 failures**, 18:50:33–19:12:16
UTC, plus three whole-gate runs afterwards — against job `0084`'s **300/0,
100/0, 25/0** at `9e2de66`. **The working tree was clean throughout** —
`git status --porcelain` empty at the start of the run, and the stash job
`0081` left is still one entry and still not in the tree.

**And the whole gate passes at this SHA, with two lines that must travel with
the word "passes".** `9cc9e65` carries a `Spec:` trailer, so `gate.sh` reaches
step 4 and prints `GATE: PASS`. Step 3 prints *"no review — ADVISORY
pre-launched (Part 0); mandatory at launched"* — which is the gate saying in
its own words that the merge has no code review and that this is legal below
`launched`. Step 4 prints *"tools/families.sh missing — breadth UNVERIFIED"*.
Three runs, three passes, both lines on each; this file repeats them rather
than rounding up.

**The host did not get quieter to produce these figures.**
`/proc/sys/net/ipv4/ip_local_port_range` is still **32768 60999**, four
`workerd` processes were running throughout, and `ss -tan` found **20**
sockets inside 34500–34599 — the block §3.1b is about. It found
**2 (measured within a minute of the campaign ending — the suite's own harness, lingering)** inside 21000–21039 and **0** inside 20500–20559, the
blocks the suite binds, after the campaign had finished. The reservation cases
draw from their own block, `resPortBase` in `relay/reservation_test.go`, also
below the floor.

**Why the gate is still qualified, and what the qualification now is.** Both of
the old reasons stay closed — the number can be measured, and no ordering
defect is open. What is left has four parts, and the third is new:

1. **Every figure this gate has ever had was taken on one machine, by the seat
   that wrote the file.** Nothing re-runs it — no CI, no second host, no other
   party. That is `Q-025`'s question arriving through this file's front door,
   and `Q-024`'s rider (c) already says in as many words that this is weaker
   than CI and is to be written down as weaker rather than as absent.
2. **A green suite is not evidence that the code is right, and this
   repository's sharpest demonstration of that is now one merge old.** Slice 1
   was merged with the suite green and a code review approving the diff — and
   **three defects were found afterwards**, the third introduced by the fix for
   the second (`spec/0004` §11, §12, §13). What established each was a
   **mutation**: reinstate the defect, run only the case that names it, confirm
   it goes red. **The run counts above say slice 2 broke nothing. They do not
   say it is correct.**
3. **The merge these figures are about has been read by nobody but its
   builder** — §3.3. `9cc9e65` is 1,089 lines to `core/` and `relay/` with two
   spec-round transcripts (one of them 304 bytes and verdict-less) and no
   code-round transcript. Its message reports two mutations of its own — D-1 red
   before the store, red again with the port pool built empty beside the durable
   set — and this seat has no way to check either without editing code it may
   not edit. Under Part 0 the merge is legal. Under this section's premise, a
   suite that passes 425 times over a change nobody else has read is a
   claim about the cases, not about the change. `BACKLOG.md` item 2.
4. **One latent fixture defect is still in a frozen file** — §3.1b,
   `BACKLOG.md` item 9 — and it is blocked on a governance reading rather than
   on work (§8). It binds nothing today; the day it binds, it stops being
   latent silently.

**And one part is gone.** Last pass's third reason — *the property item 2
exists to deliver has no test at all* — is closed. `TestAReservationOutlivesTheRelay`
returns **1** from the grep that returned 0, and it is the case `CASES.md`
froze: a second `relay.New` over the same path.

**What this does and does not do to `Q-024`.**

- It **does** complete both halves of that entry's own stated retirement
  condition, and now does so at a third SHA. The named fix — bind before the
  response leaves, on both the agent and the ssh paths — landed at `1d9505c`.
  The *"40 clean runs of each invocation recorded in `roadmap/STAGE.md`"* exist
  here at 300 and 100, plus 25 with `-race`, on a tree nine Go-tree merges past
  `83fd9f7` and with **476** more lines of test in it than at the last pass.
  And the successor defect that an earlier pass gave as its reason for not
  retiring the entry — the non-determinism having *moved* to the HTTP path — is
  itself merged.
- It **does not** retire it, and this pass does not attempt to. **Closing a
  window is the steward's act and never the seat that records the evidence**,
  and this pass is precisely the seat that recorded the evidence. The window is
  not closed, not dated, not extended, not softened, and no default is read off
  it here. The evidence is written into `pumasi/DECISIONS.md` under that entry
  in its own idiom and stops there.
- **The gate therefore stays `MET` and stays qualified**, under Q-024's own
  named default, and the qualification is: **the pure-core suite passed
  425 of 425 on one machine at this SHA, plus 3 of 3 whole-gate runs,
  with no identified defect open behind the number, nothing but this seat
  re-taking it — and nobody but the builder having read the change the number
  is about.**
- **Q-024's rider (a) still binds and its force is undiminished by the
  number.** No stage-promotion announcement may be published off this gate —
  `STAGE_PLAYBOOK.md` **Event 3 stays held**, §7 — and nothing public may quote
  the gate's figure.

**Coverage, re-measured at `9cc9e65` over the 100 `-cover` runs.**

| Package | At `9e2de66` (last pass) | At `9cc9e65` (this pass) |
| :--- | :--- | :--- |
| `core` | 84.1% | **82.5%** |
| `mux` | 84.0% | **83.5%** |
| `relay` | 85.7% | **85.1%** |
| `agent`, `cmd/pumasi-relay`, `cmd/pumasi-tunnel` | 0.0% | **0.0%** |

**Coverage fell in both packages slice 2 touched, and this file says so rather
than rounding it.** `core` 84.1 → 82.5% and `relay` 85.7 → 85.1%:
the merge inserted **542** non-test lines and removed 26 (`core/reservationstore.go`
328 new, `core/reservation.go` +108, `relay/relay.go` +106) against **476** of
test, and covered a smaller share of what it added than the packages had
before — which is where a code reviewer should look first, since a store's
uncovered lines are its error paths (`BACKLOG.md` item 2). `mux`, which nothing
touched, reads 83.5% where it read 84.0 — every coverage figure this file
has ever carried is the last run's single reading, and five further `-cover` runs taken after the campaign put `mux` at 83.5–85.1 and `relay` at 84.3–85.1 with `core` fixed at 82.5 on every one — so the `relay` drop is inside the suite's own spread and the `core` drop is not. Slice 2's
**476** lines of new test are in two files — `core/reservationstore_test.go`
(357, new) and `relay/reservation_test.go` (119) — and none of them in
`agent/`. **`agent` is still 0.0% and the merge that just landed did not change
that**: `git diff --stat 9e2de66..9cc9e65 -- '*_test.go'` reports two files,
neither in `agent/`, which is the **third consecutive merge** to make that
point (§3.2).

**Both surfaces are live, re-measured 19:01 UTC 2026-09-01.**
- *Surface A, the commons catalog*: `pumasi-web`
  `content/products/pumasi-tunnel.md`;
  `https://pumasi.ai/products/pumasi-tunnel/` → **200**, and the entry appears
  **once** in `https://pumasi.ai/llms.txt`. **But not `pumasi/catalog.json` —
  see §6.**
- *Surface B, the product's own domain*: `http://pumasi.link/` → 200, serving
  the console (`relay/dashboard.html`, `b3585f6`).

**And the product carries real traffic.** The `sshsteward` tunnel in §1 —
**36 h 32 m unbroken** at 18:51:06 UTC, carrying this machine's own sshd across
`pumasi.link:20000`. That is `RESOURCES.md` §4's remote access path, working,
and it remains the strongest evidence this product has. The second tunnel that
was beside it at the last two passes is gone (§1); it was never counted as
evidence of anything and it is not counted now.

---

## 3 · What the gate's number does not cover — and this pass closes one flag and opens a different one in its place

The Stage 1 gate is `go test` exiting 0. The gate's number is only as good as
what it covers (L-006), so what it does not cover is recorded here rather than
left implicit. Every entry below is a `BACKLOG.md` item. The flag this section
led with two passes ago — an announce-ordering defect — is closed on both
paths, and §3.1 is the record of that. **§3.3 was, last pass, the sharpest
instance of this section's premise the file had carried: a specified property
with no test at all. That flag is closed — D-1 exists — and §3.3 is rewritten
around what replaced it: the merge that created D-1 is the one merge on this
tree that no family other than its builder's has read, so the green run over it
is a claim about the cases and not about the change.**

### 3.1 · Both announce-ordering defects are closed, on both paths

**Retired by merge, and verified in the tree by this seat — at `87244af` when
this section was written, re-read at `9e2de66`, and re-read again at `9cc9e65`,
where slice 2 moved the line numbers below a second time without changing the
ordering they describe.** `relay.ServeAgent` now registers the tunnel,
**binds** the public TCP port if there is one (`relay/relay.go:254` — `0041`'s
fix, `spec/0002`), then builds the mux session, **takes `r.mu`, installs
`r.sessions[resp.AgentID]` and writes the auth response inside that one critical
section** (`:308`–`:309`), deleting the session again in the same section if the
write fails (`:313`). `ServeHTTP` (`:526`) takes the same lock to read
`r.sessions`, so a visitor arriving in what used to be the window waits for the
lock instead of being answered `404 No tunnel is open`. That is `fd523e8`,
`spec/0003-session-before-announce`, and it closes the flag this section carried
three passes ago. **Every line number in this paragraph moved with slice 1 and
again with slice 2, and was re-read at `9cc9e65`; the ordering they describe did
not move.**

**The fix is not the one this project predicted, and the difference is the part
worth keeping.** The previous order's `BACKLOG.md` item 2 said the insert
*"can simply move above"*
the announce. It cannot: the announce is written raw (`relay/relay.go:309`) while
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
`9cc9e65`). When the number it picks turns out not to be bindable,
`relay.listenTCP` returns the port and an error (`relay/tcp.go:74`) and the relay
refuses the tunnel outright — it does not ask for the next of the 99 other free
ports. **Slice 1 changed both files and left this untouched, and slice 2 changed
`relay/relay.go` again and left it untouched again** — established by reading the changed code rather than assuming
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
`core.NewPortPool` (`relay/relay.go:145`, moved from `:135` by slice 2) — a type
that *"does no I/O"* (`core/portpool.go:21`–`:22`). It binds nothing. `ss -tan`
found **20** sockets inside 34500–34599 during this pass, and the suite
still passed every run in §2. The two constants were re-read at `9cc9e65`, at
the same two line numbers (`:314`–`:315`), and are unchanged. `BACKLOG.md`
**item 9**, ranked
below everything buildable because what blocks it is a governance reading rather
than work — see §8.

### 3.2 · Three packages have no test files, `agent/` among them

`agent`, `cmd/pumasi-relay` and `cmd/pumasi-tunnel` report *no test files* and
0.0% coverage. `core`, `mux` and `relay` are the pure core the gate names and
they are covered. `agent/` is not core, but it is half of every tunnel, and
today nothing exercises it except the relay's end-to-end tests using it as a
fixture. Re-measured at `9cc9e65` over 100 `-cover` runs: still **0.0%**, while
`core` is 82.5% and `relay` 85.1%.

**Three consecutive merges have now made this flag's point.** `fd523e8` put
all **468** of its new test lines in `relay/sessionorder_test.go`. Slice 1 put
all **989** of its new test lines in `core/reservation_test.go`,
`core/portpool_test.go`, `relay/reservation_test.go` and
`relay/discardclaim_test.go` — and slice 1 is precisely a change to what
happens **across a reconnect**, which is the behaviour this entry said should
be tested *before* it was rewritten. Slice 2 put all **476** of its new test
lines in `core/reservationstore_test.go` and `relay/reservation_test.go` —
`git diff --stat 9e2de66..9cc9e65 -- '*_test.go'` reports two files and
**zero** in `agent/`. [`BACKLOG.md`](BACKLOG.md) **item 7** — *`agent/` has no
tests*. **Fixer: the coder.**

### 3.3 · The case exists now — and the merge that made it exist is the one nothing but its builder has read

**Rewritten this pass.** At `9e2de66` this section recorded that
`spec/0004/acceptance/CASES.md` specified **D-1** — *a reservation outlives
the process*: a second `relay.New` over the same path, not a reconnect —
written down deliberately at the freeze so slice 2 would be measured against a
case that predated its implementation, and that **it was not in the tree**:
`grep -rn "func TestAReservationOutlivesTheRelay(" --include=*_test.go .`
returned **0** against **1** for each of the other nineteen cases. That was
true at `9e2de66` and it is kept here as the dated record it was.

**It is in the tree at `9cc9e65`, and it is the case that was frozen.** The
same grep returns **1**, at `relay/reservation_test.go:640`, and the function
does what `CASES.md` said: claims on relay 1, closes the relay, builds a
**second `relay.New` over the same path**, and asserts the name and the port
are still held against two kinds of stranger and returned to the owner. The
comment above it says in so many words that reusing `r` would make it stop
being this case. D-2..D-7, added by `9dd067a` without editing D-1's row, exist
one function each. **So the suite can now go red on the restart half**, and
the green runs in §2 are evidence about it for the first time.

**What replaces the flag is not a missing test but a missing reader.**
`9cc9e65` — 1,089 lines to `core/` and `relay/`, a new file that writes token
digests to disk, a rewritten boot path that rebuilds the port pool from that
file — was merged by a hand run outside the queue with **no cross-family code
review**: `git show --stat` lists two spec-round transcripts and no code-round
one, one of the two is **304 bytes with no verdict**, and `reviews/` holds
nothing after 06:48 UTC. Its message carries no test count and no `GATE:`
line, and it reports two mutations of its own — D-1 red before the store, and
red again with the name half alone because the port pool was built empty
beside the durable set — that nobody but the builder has seen. Under CHARTER
Part 0 that merge is legal below `launched`. Under this section's premise, a
number over a change nobody else has read covers the cases and not the change,
and this repository's own record (§2, point 2) is that a *reviewed* slice of
this same spec had three defects found after a green suite. **The flag this
section carries is therefore: the newest merge on `main` is the
weakest-evidenced one, and it is the one that writes to disk.**
[`BACKLOG.md`](BACKLOG.md) **item 2** — *`9cc9e65` has no cross-family code
review — take one, act on the objections, and classify both slices for
release*.

---

## 4 · Why not `beta`

`beta` means strangers can rely on it and their data survives. **Each fact
below was re-verified against the tree at `9cc9e65` in this pass**, by grep and
by reading, not carried. The order is the backlog's.

**The second of the three changed this pass, one pass after the first — and
again on `main` only.** Fact 2 moved last pass; fact 3, *nothing survives a
restart*, moves this one: on `main`, a relay started with `-reservations`
keeps its reservations and their ports across a restart. On `pumasi.link` both
facts are still true in every word. Fact 1 is untouched. **The distance to
`beta` therefore shortened by exactly nothing for any user**, and this section
says so below rather than at the end, because a stage file that reports a merge
as the user-visible truth is the L-009 failure this fleet has already met twice
(§1).

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
   **On `main`, re-verified at `9e2de66` and re-read at `9cc9e65`, where slice 2
   moved every line number below without changing what they describe:**
   `core.Reservations` records a name,
   a `sha256` of its bearer token and its public TCP port; `Relay.authorize`
   consults it on the one path both ingresses share (`relay/relay.go:402`,
   `:409`); a claimed name is refused to any other token and to a tokenless
   caller; the port pool has a third state, `held`, so a tenant's number is not
   handed out between its connections. `Tunnel.Reserved` — written at what was
   `relay/relay.go:297`, then `:365`, and is `:433` at `9cc9e65` — and at every previous pass **read
   nowhere at all** — is now read at `relay/dashboard.go:80` and
   `relay/relay.go:497` (was `:429`), and is set from `Reservations.Get(name)` rather than
   from the shape of the request. The line that kept moving (236 → 249 → 274 →
   297 → 365 → 433) has stopped being the interesting fact about that field.
   **On `pumasi.link`:** none of it. The status view carries no `"reserved"` key
   (§1), so the running binary predates the change. Any anonymous agent may still
   take any free name there, including one another person is using between
   reconnects.
   **What has not changed on either.** The relay binary defines **twelve**
   flags and still no auth flag — `grep -n auth cmd/pumasi-relay/main.go` finds
   nothing — so `AllowAll` (`relay/relay.go:38`, installed at `:120` when
   `cfg.Auth` is nil) is still the only authenticator it can run. A token proves **continuity, not permission** (`spec/0004` §3):
   it narrows *which name* an accepted agent may have and does not decide whether
   the agent is accepted. And an anonymous agent taking a **free** name is
   specified behaviour that slice 1 deliberately preserves (case **C-3**) — so
   the unattributed `skk6g7tyrs` tunnel is not this fact any more, and §1
   corrects the previous pass's reading of it.
   [`BACKLOG.md`](BACKLOG.md) *Delivered*, and item 5 for the zero-install path
   this does not reach.
3. **Nothing survives a restart — *false on `main` since `9cc9e65` for a relay
   started with `-reservations`, true in every word of what `pumasi.link`
   serves*.** This is the fact that moved this pass, and like fact 2 before it,
   it is split rather than ticked.
   **On `main`, re-verified at `9cc9e65`:** `core/reservationstore.go` is new —
   one JSON document, write-to-temp-in-the-same-directory, `f.Sync()` (`:290`),
   rename, directory `Sync()` (`:318`), `flock` on a sibling lock file
   (`:123`), a corrupt file moved to `<path>.corrupt-<ts>` (`:236`) and the
   relay starting empty. `relay.New` loads it and **replays every reservation's
   port into the pool**, so a restart no longer costs an owner their name or
   their number. `LastSeen` (`core/reservation.go:68`) is set by `Claim` only;
   `ReservationTTL` is 30 days, swept at load; `MaxReservations` is 10,000
   (`core/reservationstore.go:52`, `:63`). `cmd/pumasi-relay` has **twelve**
   flags, the twelfth `-reservations` (`main.go:46`), default empty — and empty
   is today's relay in every respect, which D-7 exists to keep true. **The
   acceptance case for the property exists in the tree** (§3.3).
   **On `pumasi.link`:** none of it. The running binary predates `4489fbe`
   (§1), so a restart there still drops every subdomain, every reservation,
   every reserved port and every live tunnel — including the one carrying this
   machine's remote access, which is precisely why Q-014 exists. As of 18:51:06
   UTC it would drop **one** tunnel again rather than two (§1).
   **What has not changed on either.** The registry (`core/route.go:145`–`:147`)
   and the live tunnel are still in memory and always will be — a TCP
   connection cannot outlive the process at either end, and `9cc9e65`'s own
   message says it *"changes nothing about how long a restart takes"*. What
   it removes is the **loss**: an in-memory set was empty after a restart, so
   every claimed name was trust-on-first-use again for the reclaim window.
   `--tcp-port` keeps an address across an *agent* reconnect (`a5b77fc`), keeps
   it **against other callers** since `4489fbe`, and keeps it **across a relay
   restart** since `9cc9e65` — on `main`, with the flag, and nowhere a user can
   reach.
   [`BACKLOG.md`](BACKLOG.md) *Delivered*, first entry; and item 2 for the
   review it did not get.

**In the order that shortens the distance fastest**, which is the backlog's
order and its reasoning in one line each — and **this is §4's cost-to-move,
changed by this pass**: **item 2** first — *`9cc9e65` has no cross-family
code review* — because every build half of this section is now on `main` and
what stands between the newest of them and a release is a reader, a
classification and a note (CHARTER §4: a can-hurt release needs one approving
non-builder family, then a published note and a 7-day window; and with no
`RISK_ZONES.yaml` here, Part 3 classes the path can-hurt by default). Then
**item 3**, the relay's own bindability gap, bounded in production today only
because the deployed range happens to sit below the ephemeral floor. **Item 1**
runs in parallel and needs no coder — a decision, a deploy that carries
`-reservations` and a path, and a certificate.

**What this pass's merge bought, stated exactly.** It closed fact 3 **on
`main`, for an operator who sets the flag**. It did not close fact 1, does not
retire Q-014 by itself (the host does not have it, and the steward closes
windows), and reaches no user, because it is one of five behaviour changes
`pumasi.link` does not have. **The distance to `beta` did not shorten**, and
this file will not pretend otherwise: `beta` is what a *stranger* can rely on,
and the stranger's relay is unchanged since before 2026-08-31 10:47.

**And the honest description of the distance is now different in kind from
last pass's.** Then it was *one medium piece of work on `main`, plus a deploy
only the steward can authorise, plus a certificate nobody has installed*. The
medium piece of work is done. What is left on `main` is **no build at all**:
a code review, a classification, a release note and its window. What is left
for a user is everything — the deploy, with the flag, and the certificate —
and none of it is a coder's to take.

**Also gating, from `PRODUCT-RULES.md`** (**v1.0, 2026-08-30, on `pumasi`
`main` since `23bbc64`**, published 2026-09-01 15:42 UTC and read fresh from
`main` this pass — the first of eleven evaluations able to; every previous one
read it off the unmerged `worktree-product-rules` branch and reported that
absence as not-compliance under **Q-017**): **PR-1** (a user-visible version
number) **binds always, from the first commit**, and this product has none
anywhere — which §5 shows is no longer theoretical and got worse again this
pass; **PR-2** (in-app feedback) **binds at the `beta` promotion** and is
unbuilt. `BACKLOG.md` items 6 and 8 cite both by version and by clause.
Whether the merge closes Q-017 is the steward's reading; this seat did not
touch the entry.

## What `launched` additionally requires

Stage 2's exit gate — real end-to-end users completing workflows without an
engineer — plus production hardening. Not enumerated further while §4 is open.

---

## 5 · The version gap stopped being theoretical this week, and the newest merge has no discriminator at all

`PRODUCT-RULES.md` PR-1 asks for a version that moves and is user-visible.
There is a build on `main` that behaves differently from the build on the host,
and **no surface of this product will tell you which one is answering** — not the
console, not `/_pumasi/status`, not the logs. **This is the sixth consecutive
evaluation to date the running binary by inference rather than by reading it.**
[`BACKLOG.md`](BACKLOG.md) item 6 — *PR-1 compliance: a version that moves and
is user-visible*. **Fixer: the coder** — `core.AuthRequest.ClientVersion` is
already in the wire protocol (`core/handshake.go:33`) and no binary of this
product sets it; the only writer is `relay/sshingress.go`, filling it with the
*ssh client's* version string.

**It got worse this pass in the way that matters, and the accident that
looked like progress last pass does not extend to the new merge.**

*Worse:* **nine** merges touching the Go tree are now absent from the host, of
which **seven** change non-test code and deliver a **fifth** distinct
capability (§1). The gap between what `main` does and what a stranger meets is
the widest it has been.

*And the accidents stop here:* the two discriminators this file has — the
scheme in the `url` field (pre/post `83fd9f7`) and the `"reserved"` key
(pre/post `4489fbe`) — still date the host as pre-`4489fbe` at the 18:51:06 UTC
read. **`9cc9e65` adds a fifth capability and no third accident**: it touches
no external surface, so a relay running with `-reservations` is
indistinguishable from one without it. On the day of a deploy, a deployer
confirming *"durability is live, with the flag"* has no way to check from
outside the host — which is the same sentence this section wrote about the
guard fix last pass, one capability later. **Four of the seven non-test
commits are mutually invisible** — slice 1's three and slice 2's one.

**The gap briefly had a second reader, and now does not.** At the last two
passes the status endpoint showed a tunnel this seat could not attribute; it is
gone (§1). Had a version string been on the wire, *"which build was that agent
running"* would have been answerable while it was there. It was not, and the
question closed with the connection.

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
- **On `main`: a reservation survives a relay restart — if the operator runs
  the relay with `-reservations <path>`, and for thirty idle days.** Without
  the flag the relay is exactly as before, and the file that makes it true
  holds `sha256` digests of every token (mode `0600`), which is not a set of
  live credentials and is still worth protecting (§14.1). **On `pumasi.link`:
  nothing survives a restart** — not the name, not the reserved port, not the
  tunnel (§4, fact 3). And a live tunnel never will, on either: a TCP
  connection cannot outlive the process at its end.
- **The zero-install `ssh -R` path can hold nothing.** An ssh client cannot
  present a token, so it can be refused a name somebody else has claimed and can
  never claim or reclaim one (`BACKLOG.md` item 5).
- **One relay, one host** (`RESOURCES.md` §3: Vultr, Chicago, ~$5–6/month). A
  restart or a host failure ends every tunnel — **one, as of 18:51:06 UTC
  2026-09-01** (§1). Tailscale is kept as the independent fallback for reaching
  `m-gtr`, deliberately.
- **The commons index does not know this product exists.** `pumasi/catalog.json`
  contains **zero occurrences of the string `tunnel`** — no `products[]` entry,
  no `items[]` entry — re-verified this pass at `pumasi` `23bbc64`, `grep -c
  tunnel catalog.json` = **0**, on a file that commit changed. `README.md`
  tells every arriving agent to start with that file and treats it as the
  charter's duplication check, so an agent running that check today is told
  Pumasi Tunnel does not exist. **Recorded here and deliberately not fixed:** it
  is not this repository's file and **no role file owns it** —
  `pumasi/DECISIONS.md` **Q-019**, open, whose named default would give first
  registration to the marketing manager and ongoing `status`/`maturity` upkeep
  to this seat. Until that resolves, this is the honest place for it.
- **Two merged changes owe a release note on this seat's reading, and neither
  has one.** `pumasi/releases/` at `23bbc64` carries a single tunnel note, for
  `83fd9f7`. Under `pumasi/DECISIONS.md` **Q-034**'s default an *ordinary*
  merge owes nothing, so `1d9505c` and `fd523e8` are not gaps; what is live is
  whether **slice 1** and **slice 2** are *can-hurt* under CHARTER §2.1. Slice
  1 has the argument it had (a bearer secret on a plaintext wire; requests that
  were accepted now refused); slice 2 writes token digests to disk and, with no
  `RISK_ZONES.yaml` in this repository, is can-hurt by Part 3's default for an
  unmapped path. **Neither the classification nor the note is this seat's** —
  `BACKLOG.md` item 2 states the reasoning and routes both to the packet that
  takes the missing review, which a can-hurt release needs first.
- **The newest merge on `main` has been read by nobody but its builder** —
  §3.3, `BACKLOG.md` item 2.
- **The local request inspector on `127.0.0.1:4040` does not exist** — `web/` is
  an empty directory, re-checked at `9cc9e65`. `VALUE.md` claim 5 says so, and
  the commons catalog page already disclaims it correctly (`pumasi-web`
  `843bdef`).
- **No client TUI:** `cmd/pumasi-tunnel` is flags and logs.
- **No version anywhere** (§5).
- **`go test ./...` is no longer at the mercy of your other processes, and the
  two ordering defects behind its flakiness are both closed.** The suite's TCP
  harness draws from 21000–21039, the bind-order cases from 20500–20559 and the
  reservation cases from 21200 upward, all below the 32768 ephemeral floor; the
  TCP announce-before-bind was fixed at `1d9505c` and the HTTP
  announce-before-serve at `fd523e8`. On the machine this pass ran on — the
  same machine where the suite failed **every** run four passes ago — it passed
  **425 of 425** across three arms plus **3 of 3** whole-gate runs (§2).
  What remains: one acceptance case still names 34500–34599 but binds nothing
  (§3.1b, `BACKLOG.md` item 9); the merge that added the newest seven cases has
  no second-family review (§3.3, item 2); and nothing but this seat, on this
  machine, has ever taken any of these numbers (§2).

---

## 7 · For the marketing manager, from this evaluation

None of this is this seat's to write, and all of it is a page that contradicts a
file. Job `0066` took the product's own `README.md` on 2026-09-01 and this
section is re-stated against what is left. **One thing is new since the last
pass and it matters to every page: `main` is on GitHub now.** Slice 1 and
slice 2 were invisible to the public at the last pass (`main` ten commits ahead
of `origin/main`); the steward's push at `9cc9e65` closed that, so a card may
cite `4489fbe` or `9cc9e65` without publishing a link that 404s.

1. **`pumasi-web`'s lead sentence "There is no hosted relay" is false**, and has
   been false at five consecutive evaluations. Re-measured 18:51:06 UTC
   2026-09-01: `pumasi.link` answers `200` on 80, greets
   `SSH-2.0-pumasi-tunnel` on 2222, and is carrying **one** tunnel, unbroken for
   36 h 32 m. The hosted relay is on `pumasi.link`, not `tunnel.pumasi.ai`.
   **`0066` filed this as a `priority: high` hand-off for the commons marketing
   seat and it has not been taken; this is the fourth pass to hand it up** —
   and `pumasi-ops` job `0096` was queued against the tunnel card this tick.
2. **The gate table in §2 has changed again** — the run counts are this pass's
   and the coverage figures moved (`core` 84.1% → **82.5%**, `relay`
   85.7% → **85.1%**, `mux` 83.5% on a package nothing touched, inside the 83.5–85.1 spread five further runs showed). A page that quotes an older
   figure is quoting something this file has withdrawn. See item 5 before
   quoting the new ones: the answer is still *do not*.
3. **[`MARKET.md`](MARKET.md) is the only place a public page may take a
   competitor claim from.** Every figure in it carries a vendor URL and a fetch
   date, and its §4 records the **three** comparisons that go *against* this
   product — TLS, unowned names, and one host against three vendor edges. A
   page that states a competitor price without that citation is the failure
   `pumasi-booking` `0d1674d` already had to undo. **And `MARKET.md` says
   nothing about request inspectors** — that comparison lives in
   `docs/ux/incumbent-ux-spec.md`, line 78 and §6, which is where `VALUE.md`
   claim 5 points.
4. **Nothing public may say the `https://` problem is fixed for users, and the
   same goes for the ordering fixes, for name ownership and — new this pass —
   for a name surviving a restart.** **Five** merged behaviour changes now sit
   on `main` that `pumasi.link` does not have (§1), across nine Go-tree
   commits. The distinction between merged and served is the whole of this
   file's §1 and it has **widened** again. **In particular: no page may say
   this product gives you a name that is yours, or one that survives the
   relay restarting.** On `pumasi.link` it does neither, and `VALUE.md` claim 2
   states the split explicitly — that is the wording to take, and it is the
   only wording.
5. **`STAGE_PLAYBOOK.md` Event 3 is still held — and this is the second pass
   in a row at which one of §4's three reasons moved, which is exactly when the
   hold matters most.** That trigger chains a stage-promotion announcement and
   a public badge update off a product manager confirming an exit gate `MET`.
   This evaluation kept Stage 1 at `MET` on runs it took itself — **0
   / 0 / 0 failures in 300 / 100 / 25**, plus 3 of 3 whole-gate
   runs (§2) — and closed §4's fact 3 **on `main`, with a flag**. **It is not
   licence, and nothing public may be published off this gate.** Four reasons,
   all binding: **Q-024 is open and unretired**, and its rider (a) holds
   regardless of the rate; rider (c) says evidence getting stronger is not a
   promotion ground; **§4's fact 3 moved on `main` and not for any user**; and
   **the merge that moved it has no code review** (§3.3) — a number over a
   change nobody else has read is not a number a page may quote.

   Any page still quoting *"passes, 37 times in 40"*, the 0-in-40, the
   0-in-500, the 0-in-425 at `9e2de66`, or *"one known unfixed ordering defect
   on the HTTP path"* is quoting something this file has withdrawn. The current
   figures are the three in §2 and they carry their run counts. Nothing is lost
   by waiting: this product is `alpha`, is not asking to move, and §4 lists
   three separate reasons it should not.

6. **What a page may now say that it could not before, stated precisely because
   the temptation to overstate it is real.** Four things. *(a)* **Both
   announce-ordering defects are fixed on `main`** — the TCP one measured at 28
   dial refusals in 2000 before and 0 after (`1d9505c`), the HTTP one closed at
   `fd523e8` and pinned by a frozen case that fails deterministically without
   it. *(b)* The suite is deterministic on a machine where it once failed every
   run. *(c)* **On `main`, a name and its public TCP port can belong to
   whoever first proved they held them, and are refused to everyone else across
   a disconnect.** *(d)* **On `main`, an operator who runs the relay with
   `-reservations <path>` keeps those reservations across a relay restart**,
   for thirty idle days, with the port held as well as the name.

   **None of it may be published as a Stage 1 pass rate, and all of it carries
   the same caveat: none of it is live.** All five behaviour changes are behind
   the undeployed restart in §1. A page may say the project fixed two ordering
   defects, a flaky test suite, and the name-ownership gap **on `main`**. It
   may **not** say the product got more reliable, more permanent or faster to
   recover for anyone using `pumasi.link`, quote the gate, imply a promotion, or
   attach a run count to any of it.

   **And four things about (c) and (d) must travel with them or they become
   false claims** (`spec/0004` §3, §8, §14): a token proves **continuity, not
   identity** — there are still no accounts; it is a **bearer secret on a
   plaintext wire** until there is TLS; the **zero-install `ssh -R` path cannot
   use it at all**; and (d) is **opt-in by flag and expires after thirty idle
   days** — a page that says "your subdomain, permanently" is wrong on four
   counts, and one of them is `VALUE.md`'s withdrawn word. **Plus one thing
   about (d) that is not a caveat but a hold:** the merge that delivers it has
   no cross-family code review yet (§3.3). Do not lead with it until it has.

7. **The second, unattributed tunnel is gone**, at the 18:51:06 UTC read (§1).
   While it was there it was recorded because a restart cost two parties rather
   than one, and for no other reason; it was never evidence of adoption, a user
   or a customer, and it is not retroactively evidence of one now.

8. **The zero-install `ssh -R` path — the differentiator `MARKET.md` §2 cites
   and `VALUE.md` claim 1 leads with — is narrower on `main` than on the
   host**: it can be refused a name someone else has claimed and can never
   claim one itself. Slice 2 did not change that in either direction. Any page
   that leads with the ssh command should not also promise a stable name in the
   same breath. It is `BACKLOG.md` item 5, it needs an intent statement before
   a packet, and it is handed up here rather than fixed on a page.

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
`spec/0004`'s own. **Slice 2 kept it that way, checked at `9cc9e65`:** the range
`9e2de66..9cc9e65` changes two test files, both `spec/0004`'s own, and `9dd067a`
edits no existing `CASES.md` row. So the count of instances rose and this
product's exposure did not.

**This seat proposes and does not decide**, and it has not set a window or a
deadline. Q-030's own text points at *"`pumasi-tunnel` `BACKLOG.md` item 9"*;
that entry was item 8 after the `1853218` re-rank and is **item 9** again after
this one — unchanged in substance, and a third demonstration of why a bare number
is a poor citation into a file that reorders on purpose.

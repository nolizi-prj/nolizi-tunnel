# STAGE — Pumasi Tunnel

**Current Stage:** `Alpha` — unchanged by this evaluation, and §4 says what it
would cost to move.
**Stage 1 exit gate:** **MET 2026-08-31**, carried forward — and for the first
time the reading rests on runs that were actually taken. At `b3d251d` the
gate's own command passed **500 runs out of 500** on this machine, against
**0 clean runs in 40** at `1d9505c`. §2 gives every arm and its count. The gate stays
`MET` and stays **qualified**, for a reason that has changed: not "the number
cannot be measured" but "one known ordering defect on the HTTP path is still
unfixed and was observed twice in 240 runs by the job that merged `b3d251d`,
and this seat could not reproduce it in ~5,400 attempts."
**Q-024 is not retired by this pass** — that is the steward's act, not this
seat's — and `STAGE_PLAYBOOK.md` **Event 3 stays held** (§7).
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

**Re-evaluated 2026-08-31 at `b3d251d`** (post-`0047`, the test-port move),
`STAGE_PLAYBOOK.md` Event 2. **Every number in §2 was taken in this pass, on
this machine, at this SHA, and the number of runs behind each is stated beside
it.** Where a fact was carried from an earlier pass without being re-taken, it
says so in those words rather than being presented as fresh.

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

| | on `main` @ `b3d251d` | on `pumasi.link`, re-measured **23:49 UTC 2026-08-31** |
| :--- | :--- | :--- |
| Scheme announced | `http://` by default, `-public-scheme=https` when a terminator is in front | **`https://`** — `"url":"https://sshsteward.pumasi.link"` |
| Port 443 | not the relay's job, by design | **refused** — `curl: (7) Failed to connect to pumasi.link port 443` |
| Port 80 | serves the console and visitor traffic | **200** |
| ssh ingress 2222 | serves | **`SSH-2.0-pumasi-tunnel`** |
| TCP range in use | `-tcp-low`/`-tcp-high`, no default | **20000–20099** — `"tcp_range"`, below the ephemeral floor (§3.1a) |
| Suite determinism | **0 failures in 500 runs** (§2) | not applicable — nothing runs a suite there |
| Which build is answering | — | **unknowable from any surface** — see §5 |

So: **a person who opens a tunnel today is still handed an `https://` address
that nothing answers.** Every cell in the right-hand column was re-taken this
pass, not carried: `GET http://pumasi.link/_pumasi/status` at 23:49 UTC returns
`"url":"https://sshsteward.pumasi.link"`, `https://pumasi.link/` does not
connect, `http://pumasi.link/` answers `200`, and `pumasi.link:2222` still
greets with `SSH-2.0-pumasi-tunnel`. The fix is merged and undeployed. The relay
process on the Vultr host predates `83fd9f7` and will keep announcing `https://`
until it is restarted — and as of `b3d251d` there are now **three** merged
changes waiting behind that same restart.

**Why it has not been restarted, stated rather than left unsaid.**
`pumasi/DECISIONS.md` **Q-014** is open: it asks who may restart a relay that
keeps no durable state and whose single live tunnel is `sshsteward` —
`pumasi.link:20000` → this machine's port 22, `pumasi-ops/RESOURCES.md` §4's
remote-access route, **open 63 073 s (17 h 31 m) and unbroken** at the
measurement above (`"opened_at":"2026-08-31T06:18:13Z"`, `"age_secs":63073`) —
nearly four hours longer than at the previous pass, and the same unbroken
connection rather than a new one. Q-014 is **explicitly outside CHARTER Part 0's
proceed-on-default rule**. This evaluation therefore did not deploy, did not
treat the deploy as a judgement call, and **does not ask a coder packet to take
it either**. It is named as a blocker in `BACKLOG.md` item 1(i) and it waits for
the steward.

---

## 2 · Maturity gates

| Stage | Criteria | Status |
| :--- | :--- | :--- |
| **0. Candidate** | Scored by 3 model families (40/60) and selected by steward | **COMPLETE** |
| **1. Alpha** | Pure-core suite passes 100%; both public landing surfaces live | **COMPLETE 2026-08-31** — and measurable at last: **0 failures in 500 runs** at `b3d251d` on this machine, against 0 clean runs in 40 at `1d9505c`. Still qualified, on a different defect; see below |
| **2. Beta** | Real end-to-end users complete workflows without engineer intervention | **IN PROGRESS** — §4 |
| **3. Launched** | Production hardening, cross-model regression, 7-day veto window | PENDING |

### The Stage 1 gate, re-read at `b3d251d` — measurable for the first time

**The short answer to the question this evaluation was asked.** The previous
pass could not take the gate's number at all: the suite drew its TCP ports from
inside this host's ephemeral range and a foreign process held one, so
`go test -count=1 ./...` failed **40 of 40** at both `83fd9f7` and `1d9505c`.
`b3d251d` moved that block. The number now exists, and it is good. It is **not**
perfect, and the imperfection has moved to a different code path, which is why
the gate stays qualified rather than becoming unqualified.

**What was measured, by this seat, on this machine, at `b3d251d`, this pass.**
Run counts are given because a gate whose number is inherited rather than
measured is what produced the wrong 12-run reading that raised **Q-024**
(rider (b)).

| Arm | Runs | Failures |
| :--- | ---: | ---: |
| `go test -count=1 ./...` — the command the gate names | **500** | **0** |
| `go test -count=1 -cover ./...` | **100** | **0** |
| `tools/gate.sh`, the whole gate, `SKIP_FAMILY_PROBE=1` | **40** | **0** — 40 × `GATE: PASS` |

640 full-suite executions at this SHA, **0 failures**. Against **40 failures in
40** at `1d9505c` on the same machine. The host did not become quieter to
produce that: `/proc/sys/net/ipv4/ip_local_port_range` is still **32768 60999**,
`workerd` still held `127.0.0.1:34000` throughout this pass, and `ss -tanp`
found **12** sockets inside 34500–34599. It found **0** inside 21000–21039 and
**0** inside 20500–20559 — the two blocks the suite now actually uses. That is
the change working, and it is a property of the code rather than of the machine,
which is what `BACKLOG.md`'s previous item 2 was for.

**Why the gate is still qualified, and what the qualification now is.** It is no
longer *"the number cannot be measured"*. It is this: **`relay/relay.go` installs
the mux session after it writes the auth response**, so between the moment the
agent is handed its URL (`:205`) and the moment a visitor can be forwarded
(`:213`–`:216`) a request that resolves is answered `404 No tunnel is open`
(`:340`–`:344`). Job `0047` observed exactly that twice in 240 runs of the
merging tree, both `TestConcurrentVisitors`, and once as
`TestLargeResponseCrossesIntact`. **This seat could not reproduce it**: 640
suite executions above, plus a targeted probe driving the unmodified `relay` and
`agent` packages from outside the repository at the earliest instant `OnConnect`
can hand over the URL — **4,900 visitor requests, 0 answered `404`** (200
iterations × 20 concurrent visitors, 500 single-visitor iterations, and 400 more
with the suite looping alongside as load, which is the condition `0047` said the
flake needs). So the defect is **established by reading and by one other job's
measurement, and not by this seat's**, and this file says so rather than
inflating either half. It is `BACKLOG.md` **item 2**, the highest build entry.

**What this does and does not do to the gate.**

- It **does** deliver, at last, the *"40 clean runs of each invocation recorded
  in `roadmap/STAGE.md`"* that **Q-024** names as half its retirement condition.
  Both invocations are above, at 500 and 100 runs, plus the whole gate at 40.
  The other half — the bind-before-announce fix on both paths — was delivered at
  `1d9505c` and verified by the previous pass at 2000 iterations.
- It **does not** entitle this seat to retire Q-024, and this pass does not.
  **Q-024 stays open.** Two reasons, and the second is the substantive one:
  closing a window is the steward's act and never the seat that records the
  evidence; and the suite's non-determinism is *not gone* — it has moved to the
  HTTP path, which is a different defect from the one Q-024 names, observed by
  another job, and not reproduced here. A window whose retirement condition is
  met on paper while its subject matter has relocated is exactly the kind of
  closure this fleet has already had to walk back.
- The gate therefore **stays `MET` and stays qualified**, under Q-024's own
  named default, and the qualification is now: **the pure-core suite passed 500
  of 500 on this machine at this SHA, with one known unfixed ordering defect on
  the HTTP path that another job saw twice in 240 runs and this one saw zero
  times in ~5,400 attempts.** That is a materially stronger claim than the one
  this file made last pass, and it is stated with its run counts so the next
  seat can tell whether it still holds.
- **Q-024's rider (a) still binds, and its force is undiminished by the better
  number.** No stage-promotion announcement may be published off this gate —
  `STAGE_PLAYBOOK.md` **Event 3 stays held**, §7 — and nothing public may quote
  the gate's figure. Rider (c) is why: evidence getting stronger is not a
  promotion ground.
- **The steward closes Q-024, not this seat.** The evidence above is recorded
  under that entry in `pumasi/DECISIONS.md` in the entry's own idiom, without
  closing, dating or softening the window.

**Coverage, re-measured at `b3d251d` over the 100 `-cover` runs.** The previous
pass could not take a figure for `relay` at all — four of its tests aborted on a
port collision — and left 74.7% standing as inherited. It is now measurable:

| Package | At `83fd9f7` (last trustworthy) | At `b3d251d` (this pass) |
| :--- | :--- | :--- |
| `core` | 80.3% | **80.3%** |
| `mux` | 84.0% | **83.5%** |
| `relay` | 74.7% *(inherited, unmeasurable at `1d9505c`)* | **82.0%** |
| `agent`, `cmd/pumasi-relay`, `cmd/pumasi-tunnel` | 0.0% | **0.0%** |

`relay`'s rise is not new test-writing: it is the four TCP tests running instead
of aborting. `mux` fell 0.5 points and nothing in this pass touched `mux`;
the figure is reported rather than explained, because this seat did not
establish a cause and will not invent one.

**Both surfaces are live.**
- *Surface A, the commons catalog*: `pumasi-web`
  `content/products/pumasi-tunnel.md` (`c2084a8`, 2026-08-30);
  `https://pumasi.ai/products/pumasi-tunnel/` → 200, and the entry appears in
  `https://pumasi.ai/llms.txt`. **But not `pumasi/catalog.json` — see §6.**
- *Surface B, the product's own domain*: `http://pumasi.link/` → 200, serving
  the console (`relay/dashboard.html`, `b3585f6`).

**And the product carries real traffic.** The `sshsteward` tunnel in §1 — over
ten hours unbroken, carrying this machine's own sshd across
`pumasi.link:20000`. That is `RESOURCES.md` §4's remote access path, working,
and it remains the strongest evidence this product has.

---

## 3 · What the gate's number does not cover — four flags, and one is retired

The Stage 1 gate is `go test` exiting 0. The gate's number is only as good as
what it covers (L-006), so what it does not cover is recorded here rather than
left implicit. Every entry below is a `BACKLOG.md` item. The flag this section
carried last pass — the suite's ephemeral-range port block — is **retired by
measurement** at `b3d251d`, and §3.1 records what took its place: a defect in
the product rather than in the fixture.

### 3.1 · The port block is fixed; an announce-before-serve on the HTTP path is what is left

**The suite's port choice is delivered.** `relay/tcp_test.go:40` sets
`tcpHarnessBase = 21000` and `tcpHarnessPorts` gives each of the four harnesses
a block of ten from an atomic cursor — 21000–21039, below the 32768 floor and
clear of `bindOrderBase`'s 20500–20559. Measured effect on this host: **40
failures in 40 at `1d9505c` → 0 failures in 500 at `b3d251d`**, with the
ephemeral range unchanged and `workerd` still holding 34000 throughout (§2).
`BACKLOG.md` records it as delivered.

**What replaces it as the gate's qualification is a product defect, not a
fixture one**, which is why it is stated plainly rather than folded into the old
entry. `relay.ServeAgent` registers the tunnel (`relay/relay.go:295`), binds the
public TCP port if there is one (`:175` — `0041`'s fix, correctly placed), writes
the auth response that gives the agent its URL (`:205`), and only **then**
installs `r.sessions[resp.AgentID]` (`:213`–`:216`). A visitor arriving between
the third step and the fourth passes `registry.Lookup` (`:331`), finds
`session == nil`, and is answered `404 No tunnel is open for <host>`
(`:340`–`:344`, `:384`).

That is the same rule `spec/0002` §1 exists to enforce — *"when the client learns
the outcome, the state behind it is already true"* — one path over from where
`0041` enforced it. The TCP listener was given an accept queue to wait in; the
mux session has none, and nothing between the two steps requires the response to
have been sent first. The comment already on that branch —
*"the agent went away between the lookup and here"* — describes only the teardown
race, and is a mis-reading of the other way the branch is reached.

**Evidence, with its counts, and its limits.** Job `0047` observed it twice in
240 runs of the merging tree (`TestConcurrentVisitors`, plus once as
`TestLargeResponseCrossesIntact`). This seat did **not** reproduce it: 640 full
suite executions (§2) and 4,900 visitor requests from a targeted out-of-tree
probe, 0 answered `404`. The window is a few instructions wide and widens under
load. `BACKLOG.md` **item 2**, the highest build entry. **Fixer: the coder** —
and a frozen acceptance case that goes red at `b3d251d` is what would turn this
from a reading into a measurement.

### 3.1a · The product-side half, which is not the tests' fault

`core.PortPool` "does no I/O — it decides which number to use; binding the
listener is the relay's job" (`core/portpool.go:21–22`). When the number it
picks turns out not to be bindable, `relay.listenTCP` returns the port and an
error (`relay/tcp.go:66–70`) and the relay refuses the tunnel outright — it does
not ask for the next of the 99 other free ports. One busy port defeats a
100-port pool. `-tcp-low`/`-tcp-high` also accept a range wholly inside the
ephemeral range with no warning. The **running** relay's range is 20000–20099
(measured, §1), below the ephemeral floor, so this is bounded in production
today and unbounded for anyone who configures otherwise. `BACKLOG.md` item 4.

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
(`core/portpool.go:21`–`:22`). It binds nothing. `ss -tanp` found **12** sockets
inside 34500–34599 during this pass and the suite still passed **500 of 500**.
`BACKLOG.md` **item 9**, ranked below everything buildable because what blocks
it is a governance reading rather than work — see §8.

### 3.2 · Three packages have no test files, `agent/` among them

`agent`, `cmd/pumasi-relay` and `cmd/pumasi-tunnel` report *no test files* and
0.0% coverage. `core`, `mux` and `relay` are the pure core the gate names and
they are covered. `agent/` is not core, but it is half of every tunnel, and
today nothing exercises it except the relay's end-to-end tests using it as a
fixture. Re-measured at `b3d251d` over 100 `-cover` runs: still **0.0%**, while
`relay` rose to 82.0% without adding a line of `agent` coverage.
[`BACKLOG.md`](BACKLOG.md) item 6. **Fixer: the coder.**

---

## 4 · Why not `beta`

`beta` means strangers can rely on it and their data survives. **Each fact
below was re-verified against the tree at `b3d251d` in this pass**, by grep and
by reading, not carried. The order is the backlog's. None of the three changed:
`0047` touched one test file and no product code.

1. **There is no TLS, and what is running still says there is.** *Changed by
   the release, and only halfway.* On `main` the relay now announces the scheme
   it serves — `-public-scheme`, default `http`, validated once at startup,
   applied in one place, read by all three surfaces. **On `pumasi.link`, nothing
   listens on 443 and the relay still prints `https://`** (§1). Even once
   deployed, every HTTP tunnel here is plaintext and no `https://`-only webhook
   sender can be pointed at one. [`BACKLOG.md`](BACKLOG.md) item 1 — **operator
   action plus a deploy blocked on Q-014, not a build.**
2. **A name belongs to nobody.** Still true. `Tunnel.Reserved` is computed at
   `relay/relay.go:274` (it has been line 236, then 249, now 274; the line keeps
   moving and the fact does not) and is **never read anywhere in the tree** — the
   field is defined at `core/route.go:127`, written once, and read nowhere.
   The relay binary exposes **eleven flags and no auth flag**, so
   `AllowAll` (`relay/relay.go:40`,
   installed at `:97` when `cfg.Auth` is nil) is the only authenticator it can
   run. Any anonymous agent may take any free name, including one another person
   is using between reconnects. [`BACKLOG.md`](BACKLOG.md) item 3.
3. **Nothing survives a restart.** Still true. The registry
   (`core/route.go:145–147`) and the port pool (`core/portpool.go:27–29`) are
   plain in-memory maps, and there is no persistence path anywhere in `core/` or
   `relay/`. A relay restart drops every subdomain, every reserved TCP port and
   every live tunnel — including the one carrying this machine's remote access,
   which is precisely why Q-014 exists. `--tcp-port` keeps an address across an
   *agent* reconnect (`a5b77fc`), not across a relay one.
   [`BACKLOG.md`](BACKLOG.md) item 3.

**In the order that shortens the distance fastest**, which is the backlog's
order and its reasoning in one line each — and **this is §4's cost-to-move,
changed by this pass**: item 2 (the mux session installed after the URL is
announced) first, because it is a three-line reordering, because it is the last
identified source of non-determinism in the suite the gate names, and because
**Q-024 turns on that determinism**; then item 3, because it is the whole of
facts 2 and 3 above and it is what **retires Q-014** — once a restart costs
nobody their address, deploying stops being a steward question; then item 4, the
relay's own bindability gap, bounded in production today only because the
deployed range happens to sit below the ephemeral floor; item 1 runs in parallel
and needs no coder, only a decision and a certificate.

**What the previous pass put first here is done.** The suite's ephemeral-range
port block was this section's first cost-to-move item and it is delivered at
`b3d251d`: **0 failures in 500 runs**, against 40 in 40 at `1d9505c` (§2). The
gate is measurable and the coverage figures are real again. **The distance to
`beta` is not shorter by that much, and this file will not pretend otherwise** —
none of the three numbered facts above moved, and what that merge bought was
evidence strength, which Q-024 rider (c) explicitly says is not a promotion
ground.

**Also gating, from `PRODUCT-RULES.md`** (v1.0, read fresh 2026-08-31; still
only on the unmerged `worktree-product-rules` branch, `0115758` — **Q-017** —
and its absence from `main` is not compliance): **PR-1** (a user-visible version
number) binds **always** and this product has none anywhere — which §5 shows is
no longer theoretical; **PR-2** (in-app feedback) binds at the `beta` promotion
and is unbuilt. `BACKLOG.md` items 7 and 8. **Q-017 re-checked this pass:**
`PRODUCT-RULES.md` is **still not reachable on `pumasi` `main`** — `git ls-tree`
finds it on neither `main` nor `origin/main`, only on
`worktree-product-rules` (`0115758`). Read fresh from that branch for this
evaluation, as duty 1 of the role file requires.

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
better:** there are now **three** merged changes on `main` that the host does not
have (`83fd9f7`, `1d9505c`, `b3d251d`), and the same one inference still has to
carry all three. [`BACKLOG.md`](BACKLOG.md) item 7.
**Fixer: the coder** — `core.AuthRequest.ClientVersion` is already in the wire
protocol and no binary sets it.

---

## 6 · Known gaps a user should know about today

- **No TLS.** Every tunnel is plain HTTP, whatever URL the relay printed — and
  the running relay still prints `https://` (§1).
- **No accounts, no tokens in force, no name ownership.**
- **One relay, one host** (`RESOURCES.md` §3: Vultr, Chicago, ~$5–6/month). A
  restart or a host failure ends every tunnel. Tailscale is kept as the
  independent fallback for reaching `m-gtr`, deliberately.
- **The commons index does not know this product exists.** `pumasi/catalog.json`
  contains **zero occurrences of the string `tunnel`** — no `products[]` entry,
  no `items[]` entry — re-verified this pass at `pumasi` `3d2c638`, `grep -c
  tunnel catalog.json` = **0**. `README.md` tells every
  arriving agent to start with that file and treats it as the charter's
  duplication check, so an agent running that check today is told Pumasi Tunnel
  does not exist. **Recorded here and deliberately not fixed:** it is not this
  repository's file and **no role file owns it** — `pumasi/DECISIONS.md`
  **Q-019**, open, whose named default would give first registration to the
  marketing manager and ongoing `status`/`maturity` upkeep to this seat. Until
  that resolves, this is the honest place for it.
- **The local request inspector on `127.0.0.1:4040` does not exist** — `web/` is
  an empty directory. `VALUE.md` says so, and the commons catalog page already
  disclaims it correctly (`pumasi-web` `843bdef`).
- **No client TUI:** `cmd/pumasi-tunnel` is flags and logs.
- **No version anywhere** (§5).
- **`go test ./...` is no longer at the mercy of your other processes, and this
  is the one gap this pass closed for users.** The suite's TCP harness now draws
  from 21000–21039 and the bind-order cases from 20500–20559, both below the
  32768 ephemeral floor. On the machine this pass ran on — the same machine where
  it failed **every** run last pass — it passed **500 of 500**. What remains:
  one acceptance case still names 34500–34599 but binds nothing (§3.1b,
  `BACKLOG.md` item 9), and a rare HTTP-path 404 at tunnel setup that another
  job saw twice in 240 runs and this pass did not see in 640 (§3.1,
  `BACKLOG.md` item 2).

---

## 7 · For the marketing manager, from this evaluation

None of this is this seat's to write, and all of it is a page that now
contradicts a file.

1. **The gate table in §2 has changed** — coverage figures moved and the Stage 1
   qualification in §3.1 is materially different from the one previously
   published. The commons catalog page quotes that table verbatim.
2. **`pumasi-web`'s lead sentence "There is no hosted relay" is false**, and was
   already false at the last evaluation. `pumasi.link` resolves to
   `64.177.118.159`, serves the console on 80 and a zero-install ssh ingress on
   2222, and has carried a tunnel for **over seventeen hours** — re-measured
   23:49 UTC, `"age_secs":63073`, the same unbroken connection as last pass.
   `tunnel.pumasi.ai` indeed does not resolve; the hosted relay is on the other
   domain.
3. **[`MARKET.md`](MARKET.md) now exists** and is the only place a public page
   may take a competitor claim from. Every figure in it carries a vendor URL and
   a fetch date, and its §4 records the two comparisons that go *against* this
   product. A page that states a competitor price without that citation is the
   failure `pumasi-booking` `0d1674d` already had to undo.
4. **Nothing public may say the `https://` problem is fixed for users.** It is
   fixed on `main` and not on the internet (§1), and the distinction is the
   whole of this file's §1.
5. **`STAGE_PLAYBOOK.md` Event 3 is still held — and this time the number is
   good, which is exactly when the hold matters.** That trigger chains a
   stage-promotion announcement and a public badge update off a product manager
   confirming an exit gate `MET`. This evaluation kept Stage 1 at `MET` and, for
   the first time, on runs that were actually taken: **0 failures in 500** of
   `go test -count=1 ./...`, **0 in 100** with `-cover`, **40 × `GATE: PASS`**
   (§2). A previous pass would have read that as licence to fire Event 3. **It
   is not, and nothing public may be published off this gate.** Three reasons,
   all binding: **Q-024 is open and unretired**, and its rider (a) holds
   regardless of the rate; its rider (c) says evidence getting stronger is not a
   promotion ground, and evidence strength moving is nearly the whole of this
   pass; and the suite's non-determinism is **not gone** — it moved to the HTTP
   path, where another job saw it twice in 240 runs and this one saw it zero
   times in 640 (§3.1, `BACKLOG.md` item 2).

   Any page still quoting *"passes, 37 times in 40"*, or the 0-in-40 this file
   published last pass, is quoting a number this file has withdrawn; the current
   figures are the three above and they carry their run counts. Nothing is lost
   by waiting: this product is `alpha`, is not asking to move, and §4 lists three
   separate reasons it should not. The reading remains escalated as
   `pumasi/DECISIONS.md` **Q-024**, where this pass recorded its evidence
   **without closing, dating or softening the window**.

6. **What a page may now say that it could not before, stated precisely
   because the temptation to overstate it is real.** Two things, and the second
   is new this pass. *(a)* The announce-before-bind race is fixed and measured —
   28 dial refusals in 2000 before, 0 after (recorded at `1d9505c`). *(b)* The
   suite is deterministic on a machine where it previously failed every run:
   **0 failures in 500** at `b3d251d`.

   **Neither may be published as a Stage 1 pass rate, and (b) is the one to be
   careful with.** A green suite is not the same claim as a green product: what
   `b3d251d` fixed was the *test harness's* choice of port numbers, and the two
   defects the suite does not catch are still open (§3.1, §3.1b). A page may say
   the project fixed a flaky test suite and cite 0-in-500. It may **not** say the
   product got more reliable, quote the gate, imply a promotion, or say any of it
   is live: like everything else on `main`, all three merges sit behind the
   undeployed restart in §1.

---

## 8 · An open governance reading, escalated by this pass

Not marketing's and not this seat's to decide, recorded here because §3.1b
depends on it and a reader of that section should be able to find out why a
one-constant fix is not simply queued.

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

The question is raised as a `DECISIONS.md` entry with a named default; see the
`pumasi-ops` `DIGEST.md` entry for this pass. **This seat proposes and does not
decide**, and it has not set a window or a deadline.

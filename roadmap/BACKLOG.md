# BACKLOG — what gets built next, in order

**Owned by the product-manager role**
([`pumasi-ops/roles/product-manager.md`](https://github.com/pumasi-ai/pumasi-ops/blob/main/roles/product-manager.md),
duty 5). Seeded 2026-08-30 from candidate `0011-developer-tunnels.md`; first
evaluation pass 2026-08-31 at `3652e15`; post-release evaluation 2026-08-31 at
`83fd9f7`; post-`0041` evaluation 2026-08-31 at `1d9505c`; post-`0047`
evaluation 2026-08-31 at `b3d251d`; **post-`0060`/`0066` evaluation 2026-09-01
at `87244af`**, after the session-before-announce merge and the README repair.

One list, features and bugs together — a priority that cannot compare them is
not a priority. Every entry points at its source and carries one line of
why-here. **The top of this file is what the project manager's next coder
packet builds** — except where an entry says in its own text that it is
operator action rather than a build; the packet then takes the highest entry
that *is* a build, and the operator item keeps its rank rather than being
demoted for being unbuildable.

> **Highest *build* entry: item 2** — a subdomain belongs to nobody and nothing
> survives a relay restart. That is the next coder packet. Item 1 outranks it
> and is **not** a build: it is operator action blocked on **Q-014**.

Reordering is a commit with the reasoning in the message; the steward vetoes by
reverting.

**This list is renumbered at every pass, so cite it by number *and* title.**
Three citations in [`VALUE.md`](VALUE.md) pointed at the wrong entry this pass,
and two of the three had been correct on the day they were written — a bare
"item 9" is a pointer into an ordering this file changes on purpose. Anything
outside this file that names an entry should name its title beside the number,
so that a renumber breaks the number and not the meaning.

**What changed in this pass, and why.** The previous item 2 — *the HTTP path
hands out a URL before the session that serves it exists* — is **delivered** at
`fd523e8` (job `0060`) and is ticked below. Everything under it moves up one.
The blockquote at the top of this file went on naming it as the next coder
packet for two commits after it was built, which is the whole reason this
evaluation was queued.

Three things follow, and the second is the one worth a future seat's attention:

- **Item 1 is unchanged and keeps its rank.** It is operator action blocked on
  **Q-014**, not a build, and this file's own convention above says an
  unbuildable item is not demoted for being unbuildable. The highest *build*
  entry is now item 2, the durability work, which was item 3.
- **The fix that entry predicted does not work, and its confidence in that
  prediction is why it was ranked as three lines.** The prediction was tested by
  building it and it fails worse than the defect it was meant to repair. See
  *Delivered* below; it is recorded there rather than deleted with the entry.
- **The figures in this file and in [`STAGE.md`](STAGE.md) §2 were re-run at
  `87244af` by this seat**, with the run count beside each. Job `0060` published
  its own; none of them is carried here.

---

## The order

**1 · Nothing serves TLS, and the relay that is actually running still says it
does** — *operator action and a blocked deploy, not a build.* Source: the
`83fd9f7` evaluation, checking the live relay rather than the code path;
unchanged by `0041`, which touched no deployed binary.

The build half is done and the lie is gone from `main`; it is not gone from the
internet. Two things are outstanding, and they are different sizes:

- **(i) Deploy the merged fix.** `pumasi.link` runs a pre-`83fd9f7` binary and
  will announce `https://` until someone restarts the relay. **This is blocked
  on `pumasi/DECISIONS.md` Q-014**, which asks who may restart a host whose one
  live tunnel is `sshsteward` → this machine's port 22 (`RESOURCES.md` §4).
  Q-014 is open and is **explicitly outside CHARTER Part 0's proceed-on-default
  rule**. Neither this seat nor a coder packet may take it, and this file does
  not ask for it. **Re-measured read-only at 02:48 UTC 2026-09-01, no host
  touched:** `http://pumasi.link/_pumasi/status` still reports
  `"url":"https://sshsteward.pumasi.link"`, `curl https://pumasi.link/` fails to
  connect on a refused 443, `http://pumasi.link/` answers `200`, and
  `pumasi.link:2222` greets `SSH-2.0-pumasi-tunnel`. **Five commits now sit on
  `main` that the host does not have, of which three change the relay's
  behaviour** — `83fd9f7` (scheme), `3480990` (bind before announce) and
  `fd523e8` (session before announce); `e40a224` and `b3d251d` touch tests and
  specs only, so they change nothing a user of the running relay would see.
  **And what a restart would cost has doubled since the last pass**: the status
  read above reports `"count":2` — see *Not on this list* for the second tunnel,
  which is not this project's and which nobody here can identify.
- **(ii) Put a wildcard certificate for `*.pumasi.link` in front of the relay
  on the Vultr host.** This is the actual TLS gap: with (i) done, every tunnel
  is *honestly* plaintext, which is still a product that no `https://`-only
  webhook sender can be pointed at. TLS termination is deliberately outside the
  relay (`cmd/pumasi-relay/main.go` header), and outside it there is still
  nothing. `RESOURCES.md` §2 warns that proxying these records through
  Cloudflare would break raw TCP, so that is not the shortcut it looks like.

Why here: it is the largest single gap between what this product is and what a
stranger could use, it is the one item on this list every visitor to
`pumasi.link` meets today, and it is not demoted for being unbuildable.

**2 · A subdomain belongs to nobody, and nothing survives a relay restart** —
**the next coder packet.** Source: `VALUE.md` claim 2, which sold "permanent"
and "stable across restarts". All three facts re-verified in the tree by this
seat at `87244af`, by grep and by reading: `Tunnel.Reserved` is written once, at
`relay/relay.go:297` — the line has been 236, 249, 274 and is now 297; the line
keeps moving and the fact does not — and is **never read anywhere**, the field
being defined at `core/route.go:127` and read by nothing; `cmd/pumasi-relay`
defines **eleven flags and none of them is an auth flag** (`grep -n auth
cmd/pumasi-relay/main.go` finds nothing), so `AllowAll` is the only
authenticator that binary can run; and the registry and port pool are plain
in-memory maps (`core/route.go:145`–`:147`, `core/portpool.go:27`–`:29`, no
persistence path anywhere in `core/` or `relay/`). A relay restart drops every
name, every reserved port and every live tunnel at once. `--tcp-port` survives
an *agent* reconnect, not a relay one.

**Why it is now the top build entry, which is a promotion by subtraction rather
than by argument.** It has always been the `beta` bar itself (`STAGE.md` §4) and
what retires **Q-014** — once a restart costs nobody their address, who may
deploy stops being a steward question and becomes an ordinary one. What changed
is only that the entry it sat under — predicted here as a three-line reordering,
and not one — was delivered at `fd523e8`. It is still the largest single piece of
work on this list, and nothing in this pass made it smaller or more urgent; it is
on top because everything cheaper above it is done or is blocked on the steward.
**Fixer: the coder** — and the packet that takes it should expect a spec, not a
patch: the *Delivered* entry for the previous item 2, below, is the record of
what happens when this file guesses at an implementation.

**One measurement that bears on it, taken this pass and not inherited.** The
live relay is now carrying **two** tunnels rather than one (§*Not on this
list*), and the second was opened by an agent this seat cannot identify — which
is `AllowAll` behaving exactly as this entry describes, observed rather than
predicted.

**3 · A public port the pool believes is free may not be bindable, and the
relay gives up instead of taking the next one** — source: this evaluation,
found while establishing why the previous order's item 2 failed. It is the
product-side half of that port-range defect — whose test-side half is now split
between *Delivered* (`b3d251d`) and item 8 — and it is a different defect from
both. Re-verified in the tree at `87244af`; nothing since `0047` changed
`core/portpool.go` or `relay/tcp.go`.

`core.PortPool` is explicit that it *"does no I/O — it decides which number to
use; binding the listener is the relay's job"* (`core/portpool.go:21–22`). So
`Allocate` can hand back a port that the OS will refuse. When it does,
`relay.listenTCP` releases the port and returns an error
(`relay/tcp.go:66–70`), and `ServeAgent` treats that error as fatal to the
handshake (`relay/relay.go:169–190`) — it unregisters the agent and refuses the
tunnel. **It never asks the pool for the next free port**, even though the other
99 in the range are available. Reproduced this pass: with one port of a 100-port
range held by an unrelated process, every single tunnel request was refused.

There is also no guard on the operator's side: `-tcp-low` / `-tcp-high`
(`cmd/pumasi-relay/main.go:39–40`) accept any range, including one wholly inside
the host's ephemeral range, and neither `core.NewPortPool` nor `relay.New` warns.

**What bounds this today, measured rather than assumed.** The running relay's
range is **20000–20099** (`http://pumasi.link/_pumasi/status`, `"tcp_range"`,
re-read 2026-09-01 02:48 UTC and unchanged), which is *below* the ephemeral
floor. So the
kernel cannot steal a port from the live relay, and the exposure there is only
another process on the Vultr host binding into that block. The unbounded case
is the operator who picks a range above 32768 with nothing to warn them, and
the suite, whose remaining half is item 8 — *a frozen acceptance case still
draws its port range from inside the kernel's ephemeral range*.

Why here: the failure is honest and self-healing since `1d9505c` — the agent is
refused rather than lied to, and it retries — so this is robustness, not a
falsehood on shipped surface, which is why it sits below item 2 rather than
above it. But it is the reason a 100-port pool can be defeated by one busy port,
and the range guard is a few lines that would have made the previous order's
item 2 impossible to write. **Fixer: the coder.**

**4 · The console never offers the zero-install `ssh -R` command** — source:
`VALUE.md` claim 1 against the live page. The ingress works: a banner grab on
`pumasi.link:2222` returns `SSH-2.0-pumasi-tunnel`. But `relay/dashboard.html`'s
command builder emits only `pumasi-tunnel --relay …` and its "First time here"
panel offers only `git clone && go build` — so the product's headline claim, the
one thing it does that needs nothing installed, is absent from the one page a
visitor sees.
Re-verified at `87244af`: `relay/dashboard.html` contains **0** occurrences of
`ssh -R` and **1** of `git clone`, and `pumasi.link:2222` still answered
`SSH-2.0-pumasi-tunnel` when this seat grabbed the banner at 02:48 UTC.
Why here: it is an afternoon's work on a page that is already live, and it
converts the strongest differentiator from a README sentence into the thing you
can paste. `MARKET.md` §2 makes that differentiator explicit and cited, which
raises what the omission costs.

**5 · `agent/` has no tests** — source: the gate reading; L-006.
`go test ./...` reports **no test files** for `agent`, `cmd/pumasi-relay` and
`cmd/pumasi-tunnel`. The two `cmd` packages are argument parsing and can wait.
`agent/` cannot: it is half of every tunnel — handshake, reconnect, local dial,
stream fan-out — and today the only thing exercising it is `relay`'s
end-to-end tests, which use it as a fixture and assert on the relay's behaviour,
not its. Reconnect and local-dial-failure have no case at all.
Re-verified at `87244af`: the three packages still report *no test files*, and
the `-cover` runs of this pass put them at **0.0%** (`STAGE.md` §2 carries the
figures and their run counts). Nothing in `fd523e8` added a line of `agent`
coverage; its 468 new test lines are all in `relay/sessionorder_test.go`.
Why here: it is the coverage the Stage 1 number does not include, and item 2 —
the durability work — will rewrite reconnect behaviour, so the tests should
exist before that and not after.

**6 · PR-1 compliance: a version that moves and is user-visible** — source:
[`PRODUCT-RULES.md` PR-1](https://github.com/pumasi-ai/pumasi/blob/worktree-product-rules/PRODUCT-RULES.md)
(v1.0, 2026-08-30; **binds always, from the first commit**; read fresh this pass
and **still only on the unmerged `worktree-product-rules` branch** — **Q-017** —
and its absence from `main` is not compliance — **re-checked this pass:
`git ls-tree` finds it on neither `main` nor `origin/main`**). This product has
no version anywhere: no version file, no `/version`, nothing in the console
footer, nothing in a release note. Re-verified at `87244af`, with one detail
that makes the fix smaller than it looks: **the repository root already has a
`package.json`** — added so `pumasi/tools/gate.sh` step 1 finds a suite — and it
has **no `version` field**, which is exactly where PR-1 says the one source of
truth belongs. `core.AuthRequest.ClientVersion` exists as a field
(`core/handshake.go:33`) and the only thing that ever sets it is
`relay/sshingress.go:165`, which fills it with the *ssh client's* version string,
not this product's.
Why here: there are now **five** merges on `main` that the host does not have —
three of which change the relay's behaviour (item 1) — and **nothing on the
console, in `/_pumasi/status` or in the logs says which build is answering**.
This pass had to infer it from the scheme in a `url` field, for the fourth
evaluation running. That is `pumasi-booking`'s Q-012 problem
arriving early, and here the answer is a few lines of Go plus a field already in
the wire protocol.

**7 · PR-2 compliance: in-app feedback** — source:
[`PRODUCT-RULES.md` PR-2](https://github.com/pumasi-ai/pumasi/blob/worktree-product-rules/PRODUCT-RULES.md)
(binds at the `beta` promotion; below `beta`, encouraged). Nothing in the
product collects feedback. The reference behaviour is `pumasi-booking`'s
(`service/src/feedback.ts`) — matched in behaviour, not copied.
Why here: it **gates the `beta` promotion**, so it must be built before the
label moves, and it is worth little before items 1–2 make the thing worth
reporting on. The natural home is the console, where a visitor already is.

**8 · A frozen acceptance case still draws its port range from inside the
kernel's ephemeral range — real, latent, and blocked on a governance reading
rather than on work** — source: the residual half of the previous order's item
2, which `b3d251d` did not land.

`relay/scheme_test.go:314`–`:315` configures `TCPPortLow: 34500`,
`TCPPortHigh: 34599`. This host's `/proc/sys/net/ipv4/ip_local_port_range` is
**32768 60999**, so that block sits inside the range the kernel hands out as
source ports for outgoing connections — the identical defect that `b3d251d`
fixed in `relay/tcp_test.go`.

**Why it did not land with the rest.** Those two lines are inside **A-10**,
`TestSchemeChangesNothingButTheScheme` (`relay/scheme_test.go:295`), a frozen
acceptance case of `spec/0001-public-scheme`. Job `0047` moved it using CHARTER
§3 requirement 2's own named remedy — *"If the tests are wrong, amend the spec in
the open and take a fresh cross-family spec review"* — wrote the amendment as
SPEC 0001 §7, and got a fresh gemini spec review that approved it
(`reviews/20260831-161500-spec-gemini.md`). The **code** review then objected,
citing that same clause, on the ground that the *builder* authored the
amendment. A cited objection governs and may not be argued past (CHARTER Part
3), so the builder reverted the whole half — the A-10 fixture, SPEC 0001 §7, the
`CASES.md` note and SPEC 0002 §6.6 — rather than proceed. The reviewers
contradicted each other and each switched sides on the same clause across the
two ranges; all four transcripts are committed. **The reading is open and is
escalated, not decided here** — see the digest entry for this pass and
[`STAGE.md`](STAGE.md) §3.1.

**Why it ranks eighth, below item 2 and below everything buildable.** Two
reasons, and the first is measured rather than argued:

- **Its cost today is nothing, and this pass re-measured that rather than
  carrying `0047`'s figure.** A-10 calls `relay.New` and reads
  `Registry().PublicScheme()` back. `relay.New` passes the range to
  `core.NewPortPool` (`relay/relay.go:124`), which is explicit that it *"does no
  I/O"* (`core/portpool.go:21`–`:22`). **A-10 binds nothing**, and this pass
  re-read the two constants at `87244af` to confirm they are still 34500 and
  34599. The block stayed contended on this host through the runs recorded in
  [`STAGE.md`](STAGE.md) §2 and the suite did not notice. A latent defect that a
  contended range cannot provoke is a latent defect.
- **Nothing a coder packet could take sits under it.** The work is one constant;
  what blocks it is whether §3 requirement 2's remedy is available to a builder
  at all, given that Part 3 requirement 1 has the builder authoring every spec in
  the first place. Ranking it above a buildable item would point the next packet
  at something that cannot be built without answering that first.

Why here rather than lower: it is a real defect in a frozen file, and the day
A-10 gains an assertion that binds — or `relay.New` gains a bind — it stops being
latent silently. It sits beside item 9 — *`relay.New` accepts a `BaseDomain`
that `core.NewRegistry` silently normalises differently* — the other entry on
this list that is real, cheap and has no live consequence. **Fixer: the coder, once the reading
resolves**; `SPEC 0002` §6.5 is the precedent on the permissive side, where the
same builder amended a frozen fixture in the open and it stood.

**9 · `relay.New` accepts a `BaseDomain` that `core.NewRegistry` silently
normalises differently** — source: `0041`, found while verifying SPEC 6.1 and
deliberately left out of its scope.

`relay.New` rejects only the empty string; `core.NewRegistry` trims and
lowercases. So `-domain " pumasi.link"` — a stray character in a systemd unit —
starts a relay whose every registry lookup is built from the untrimmed string
and never matches. `0041` recorded the boundary precisely, and it is narrower
than it first looks: **trailing space, uppercase, trailing dot and doubled dot
are all unaffected**; it is specifically the **leading** space. The condition is
now pinned by test as `lookupBreakingDomain` (`relay/bindorder_test.go:50–56`),
where it is used as a deliberate fault injection rather than fixed.
Why here, and honestly low: it has no live consequence today — the deployed
relay's domain is correct, the failure is loud and total rather than silent and
partial once you look at any tunnel, and the tests that would catch a regression
already exist. It is a validation asymmetry between two constructors, worth one
guard in `relay.New`, and it does not belong near the top of a list whose top
two items are a lie on the internet and the `beta` bar. Re-verified at
`87244af`: `lookupBreakingDomain` is still `relay/bindorder_test.go:56`, and
`relay.New` still tests `cfg.BaseDomain` only for the empty string
(`relay/relay.go:93`–`:94`).

**10 · Client CLI: the interactive terminal UI** — source: the seeded item 5;
[`docs/ux/incumbent-ux-spec.md`](../docs/ux/incumbent-ux-spec.md) §273 for what
the incumbents' status line shows. `cmd/pumasi-tunnel` prints logs; there is no
live view of requests, response times or status codes, and no released binary
for macOS or Windows.
Why here: it is polish on a path that works, and every item above it is either
something untrue or something missing that a user would rely on.

**11 · Local request inspector on `127.0.0.1:4040`** — source: the seeded item
6; [`VALUE.md`](VALUE.md) claim 5 (where it is marked *not built*). Re-verified
at `87244af`: `web/` still contains **0** entries; there is no listener, no SSE,
no replay.
Why here: last of the build items. It is a whole second product surface, and the
honest catalog page (`pumasi-web` `843bdef`) already tells visitors it does not
exist, so nobody is being misled while it waits.
**Corrected this pass:** this entry said *"`MARKET.md` §2 records that all three
incumbents ship one"*. It does not — `grep -ci inspector roadmap/MARKET.md`
returns **0**, at every revision that file has ever had, and its §2 table carries
five rows, none of them an inspector. The claim is sound and its source is
[`docs/ux/incumbent-ux-spec.md`](../docs/ux/incumbent-ux-spec.md): the
comparison table at line 78 gives an Inspector row for all three, and §6.1, §6.4
and §6.5 give one section per vendor. `VALUE.md` claim 5 carried the same wrong
citation and is repaired in the same commit; both were written into this
repository on the same day, in `dda04c7`, the commit that created `MARKET.md`.
That makes it a real gap, not a whim, but not ahead of anything above.

---

## Delivered

Verified against the tree during this evaluation, not taken from commit
subjects.

- **The HTTP path announces a URL only once the session that serves it exists**
  — delivered **`fd523e8`** (job `0060`), spec
  `spec/0003-session-before-announce`. This was **item 2 of the previous order**,
  and the entry at the top of this file went on naming it as the next coder
  packet for two commits afterwards.

  Verified in the tree by this seat, by reading `relay/relay.go` at `87244af`:
  the mux session is built, `r.mu` is taken, `r.sessions[resp.AgentID]` is
  installed and the auth response is written **inside that one critical
  section** (`:222`–`:231`), and the session is deleted again in the same section
  if the write fails (`:232`–`:236`). `ServeHTTP` takes the same lock to read
  `r.sessions` (`:255`–`:257`), so a visitor who arrives in what used to be the
  window now waits for the lock instead of being answered `404 No tunnel is
  open`. Frozen cases: `relay/sessionorder_test.go` — C-1
  `TestVisitorIsNotAnsweredBeforeTheSessionExists`, C-2
  `TestTheAnnounceReachesTheAgentBeforeAnyStream`, C-3
  `TestAFailedAnnounceLeavesNothingBehind`.

  **⚠ This entry predicted the wrong fix, and that is the half worth keeping.**
  Item 2 said, in its own words: *"there is nothing between step 3 and step 4
  that needs the response to have been sent, so the insert can simply move above
  it."* Its confidence in that reading is why it was ranked as a **three-line
  reordering** — here, in [`STAGE.md`](STAGE.md) §4's cost-to-move, and in the
  evidence rows this project wrote under **Q-024**. **The bare reorder was
  built, and it is worse than the defect it was meant to repair.**

  The mechanism, re-read by this seat rather than taken from the build report.
  The announce is written **raw** — `r.writeFrame(conn, okFrame)`, which is a
  plain `w.Write` at `relay/relay.go:327`–`:334` — while `mux.Session.Open`
  writes a `FrameOpen` on **the same connection** under the session's own
  `writeMu` (`mux/session.go:88`, `:102`, `:181`). Two writers, one socket, no
  shared lock. So a visitor forwarded in the *new* window puts a stream frame on
  the wire ahead of the auth response; the agent is sitting in
  `core.DecodeFrame` waiting for exactly one frame, `DecodeAuthResponse` rejects
  what arrives, and the tunnel drops into reconnect backoff. Measured by `0060`
  against the bare reorder and recorded in `spec/0003/SPEC.md` §2 and §6: **C-1
  answers `502` where the unfixed tree answered `404` honestly, and C-2 waits
  10.76 s for an `OnConnect` that never fires.** A wrong answer about a tunnel
  that is about to work, in place of a right answer about one that is not open
  yet.

  **Why a future seat should care rather than file this as trivia.** The
  prediction was not careless — it was read off `spec/0002` §2, which states in
  its own text that the mux session cannot be created before the handshake
  response. That half of `spec/0002` is itself wrong, which is why `0060` wrote a
  new spec instead of folding a contradiction into a frozen one (L-007). And the
  reorder would have passed everything this repository had: **500 clean suite
  runs, the whole gate, and every frozen case that existed before
  `spec/0003`** — the figures this file published at `b3d251d`. The transferable
  rule is the one this entry did not follow: **a fix named in a backlog entry is
  a claim, not evidence, and ranking an item by that claim mis-sizes it.** The
  three lines this entry promised came to ten added and five removed, of which 26
  of 36 added lines are comment explaining why a lock is now held across a socket
  write.

  **What this does not do.** It does not retire **Q-024** — that is the
  steward's act, and it is not taken here; `STAGE.md` §2 records what this pass
  measured and stops. And it is the third of three merged behaviour changes still
  waiting behind item 1's undeployed restart, so **no user of `pumasi.link` has
  it.**

- **The test suite's TCP harness moves below the kernel's ephemeral floor** —
  delivered **`b3d251d`**. This was the *first half* of item 2 of the previous
  order; the second half is item 8 above and is **not** delivered.

  Verified in the tree: `relay/tcp_test.go:40` sets `tcpHarnessBase = 21000`,
  and `tcpHarnessPorts` hands each harness a block of ten from an
  `atomic.AddInt32` cursor — four harnesses, **21000–21039**, below the 32768
  floor and clear of `bindOrderBase`'s 20500–20559. That is `bindorder_test.go`'s
  existing scheme applied rather than a second one invented, and it is SPEC 0002
  §6.5's finding (each case gets its own block) carried across to the older
  harness. No product code changed.

  **Verified by measurement, this seat's own, at `b3d251d` on this host.** The
  host is the same one whose ephemeral range made the previous pass unmeasurable
  — `/proc/sys/net/ipv4/ip_local_port_range` is still **32768 60999**, `workerd`
  still held **127.0.0.1:34000** throughout, and `ss -tanp` found **12** sockets
  inside 34500–34599. The suite no longer cares:

  | Arm | Runs | Failures |
  | :--- | ---: | ---: |
  | `go test -count=1 ./...` (the gate's step 1, verbatim) | **500** | **0** |
  | `go test -count=1 -cover ./...` | **100** | **0** |
  | `tools/gate.sh`, whole gate, `SKIP_FAMILY_PROBE=1` | **40** | **40 × `GATE: PASS`** |

  Against **40 failures in 40** at `1d9505c` on this same machine, recorded
  above. `ss -tanp` found **0** sockets in 21000–21039 and **0** in 20500–20559,
  which is the property the change was for.

  **Coverage is measurable again, and it moved.** The previous pass could not
  take a figure for `relay` at all — four of its tests aborted on a port
  collision — and carried 74.7% as inherited. Re-measured here over the 100
  `-cover` runs: `core` **80.3%**, `mux` **83.5%**, `relay` **82.0%**. `agent`,
  `cmd/pumasi-relay` and `cmd/pumasi-tunnel` remain **0.0%** — item 5.

  **What this does not do:** it does not retire **Q-024** (that is the steward's,
  and `STAGE.md` §2 records the evidence without closing the window), it does not
  fix item 8, and it is one of five merges waiting behind item 1's undeployed
  restart — though, being test-only, it is one of the two that would change
  nothing on the host even once deployed.

- **The public TCP address is bound before it is announced** — delivered
  **`1d9505c`** (`3480990` the relay change, `e40a224` the suite fixture),
  spec `spec/0002-bind-before-announce`. This was item 2 of the previous order.

  Verified in the tree: `relay/relay.go` now calls `r.listenTCP` **before**
  `core.EncodeAuthResponse` and `r.writeFrame`, and a bind failure unregisters
  the agent and answers the handshake with an error frame rather than a
  correction sent after the address left. `relay/tcp.go` separates `listenTCP`
  (bind) from `serveTCP` (forward) so the socket can be answering before a
  session exists, and a visitor who arrives in the gap waits in the accept
  queue. `relay/sshingress.go` carries the same ordering.

  **Verified by measurement, this seat's own, not the fix's report.** The
  ordering defect was isolated from the host by running the same unmodified
  `relay` and `agent` packages on a port range **outside** the ephemeral range
  (14000–14099, confirmed wholly bindable), taking the announced `TCPAddr` and
  dialling it the instant it arrived:

  | | `83fd9f7` (before) | `1d9505c` (after) |
  | :--- | :--- | :--- |
  | 200 iterations | 3 dial refusals | **0** |
  | 500 iterations | 5 dial refusals | **0** |
  | 2000 iterations | **28 dial refusals** (1.4%) | **0** |

  The defect was real, it was reproducible, and it is gone.

  **What was *not* verified, and must not be read into the above:** the
  suite-level figures the fix reported (0 in 40 of each invocation) did not
  reproduce here, and neither did the 3-in-40 baseline. `go test -count=1 ./...`
  failed **40 of 40** at `83fd9f7` and **40 of 40** at `1d9505c` on this
  machine, for the ephemeral-port reason that `b3d251d` has since fixed and
  that was never this change's doing. Both figures are superseded by this pass:
  see the entry below and [`STAGE.md`](STAGE.md) §2.
- **Item 1, half (a) · The relay announces the scheme it actually serves** —
  delivered `83fd9f7` (2026-08-31 10:47), released as
  `pumasi/releases/2026-08-31-pumasi-tunnel-public-scheme.md`, **Q-020**, 7-day
  window. `core.ParsePublicScheme` (`core/route.go:43`) is the only place the
  legal set is written; the scheme is stored once on the registry and read once
  by `Registry.PublicURL`; `relay.New` validates it at startup and **refuses to
  start** on an unknown value; `cmd/pumasi-relay` exposes `-public-scheme`
  defaulting to `http`. All three surfaces that show a person an address read
  that one function. Tests: `core/scheme_test.go`, `relay/scheme_test.go`.
  **Not delivered: the deploy.** See item 1 — the running relay still announces
  `https://`.
- **`roadmap/MARKET.md`** — written in the `83fd9f7` pass:
  [`MARKET.md`](MARKET.md), three comparators, every price read from the
  vendor's own page on 2026-08-31 with the URL and the date in the claim. Its §4
  records the two comparisons that go against this product.
  *Product-manager action, not a build.*
- **Seeded 1 · Pure-core stream multiplexer** — delivered `8ff1605` as `mux/`,
  with its own framing rather than `yamux` or `quic-go`. Tests:
  `mux/session_test.go`. *Not delivered from the seed's text: throughput
  benchmarks.*
- **Seeded 2 · SSH ingress gateway** — delivered `3652e15` as
  `relay/sshingress.go` (`golang.org/x/crypto/ssh`), serving `-ssh-addr :2222`.
  Live: `pumasi.link:2222` answers `SSH-2.0-pumasi-tunnel`. *Port 22/443 ingress
  is not delivered and is not wanted — 22 is the host's own sshd and 443 is item
  1's.*
- **Seeded 3 · HTTP wildcard host router** — delivered `a69008f` and `bf837ee`:
  `core/route.go` (case-insensitive registry, wildcard matching, reserved-name
  list in `core/subdomain.go`) plus the relay's visitor HTTP path. Tests:
  `core/route_test.go`, `relay/endtoend_test.go`. **The ACME/wildcard-TLS half
  of this seed item is *not* delivered — it is item 1(ii).**
- **Seeded 4 · Raw TCP port pool** — delivered `a13e586` and `a5b77fc`:
  `core/portpool.go` (forward-walking cursor) and `relay/tcp.go` (byte-for-byte
  `pipe`, half-close where the stream supports it). Tests:
  `core/portpool_test.go`, `relay/tcp_test.go`. Live and load-bearing:
  `pumasi.link:20000` carries this machine's sshd (last measured at the `83fd9f7` pass; not re-measured here — this seat did not touch the live host). **Its
  announce-before-bind defect is delivered above; its bindability defect is
  item 3, and its test-range defect is delivered at `b3d251d` for
  `tcp_test.go` with a residual at item 8.**
- **Seeded 5 · Client CLI, in part** — delivered `bf837ee`:
  `cmd/pumasi-tunnel` with `--relay`, `--subdomain`, `--token`, `--host`,
  `--tcp`, `--tcp-port`, `-v`, one static binary. *The interactive TUI and the
  published macOS/Windows binaries are not delivered — item 10. The seed's
  `--http` and `--auth` do not exist: HTTP is the default and the flag is
  `--token`.*
- **Not seeded, delivered anyway** — the relay console at the apex
  (`relay/dashboard.go`, `relay/dashboard.html`, `b3585f6`), shaped by
  `docs/ux/incumbent-ux-spec.md`, live at `http://pumasi.link/` and the Stage 1
  Surface B. Its gap is item 4 — *the console never offers the zero-install
  `ssh -R` command*.

---

## Not on this list, and why

- **`pumasi/catalog.json` has no entry for this product** — zero occurrences of
  the string `tunnel`, re-checked this pass at `pumasi` `2ab3a4f`
  (`grep -c tunnel catalog.json` = **0**). It is a real defect and it is recorded
  in
  [`STAGE.md`](STAGE.md) under known gaps. It is **not** a backlog item here
  because it is not this repository's file and **no role owns it** — that is
  `pumasi/DECISIONS.md` **Q-019**, open. Nothing a coder packet on this product
  could build would close it.
- **Deploying the relay** — **Q-014**, open, and outside CHARTER Part 0. Named
  in item 1(i) as a blocker; not requested of anyone here.
- **Merging `PRODUCT-RULES.md` to `pumasi` main** — **Q-017**, open. This pass
  read the file fresh from the `worktree-product-rules` branch, as items 6 and 7
  record. **Ninth consecutive evaluation to report it absent from `main`**:
  `git ls-tree` at `pumasi` `2ab3a4f` finds it on neither `main` nor
  `origin/main`. Not this repository's file.

- **A second live tunnel on `pumasi.link` that this product cannot account
  for.** `GET http://pumasi.link/_pumasi/status` at 02:48 UTC 2026-09-01 reports
  `"count":2`: the long-running `sshsteward` tunnel
  (`pumasi.link:20000` → port 22, `"opened_at":"2026-08-31T06:18:13Z"`,
  `"age_secs":73789` — 20 h 30 m unbroken, `"fixed":true`), and a second one,
  subdomain **`skk6g7tyrs`**, `pumasi.link:20002` → a `"local_port":3389`
  somewhere, `"fixed":false`, `"opened_at":"2026-09-01T01:48:23Z"` and open for
  59 minutes at the time of the read. **This seat did not establish who opened
  it and will not guess.** It is recorded because three things follow and none of
  them is a build: it is the first traffic on this relay that is not the
  steward's own ssh route; it is `AllowAll` working as item 2 describes, since
  an anonymous agent taking a free name is precisely what that entry says the
  relay cannot refuse; and it **doubles what a restart costs**, which is the
  fact **Q-014** turns on and whose text still says the live set is *"exactly
  one"*. Nothing was done to the host, the tunnel or the entry.

- **The uncommitted change to `cmd/pumasi-tunnel/main.go` that was in the
  working tree when this evaluation started, and still is.** `git diff --stat`
  reports `cmd/pumasi-tunnel/main.go | 6 +++---` — three hunks that make
  `--relay` **default to `pumasi.link:7000`** instead of being required, drop
  `--relay` from the usage string, and delete the `*relayAddr == ""` arm of the
  argument check. **No job claims it**: nothing in `pumasi-ops/jobs/done/`
  names it and `git log -3 -- cmd/pumasi-tunnel/main.go` ends at `a5b77fc`, well
  before any candidate. **This seat did not commit it, revert it, check it out
  or stash it**, and `git diff --stat` still reports it unchanged after this
  commit; the staging here was by path.

  **It contradicts the file the commit before this one repaired.**
  [`README.md`](../README.md) line 69 says `--relay` is **required**, line 75
  tabulates its default as *(none — required)*, and all four invocations at
  lines 59–66 pass it explicitly. `87244af`'s subject is *"the front door stops
  contradicting the source tree"*; this diff would make the front door wrong
  again, in the same place, one commit later.

  **It is not ranked as a build, and the reason is not that it is small.** The
  ergonomics gain is narrow: the zero-install path this product leads with is
  `ssh -p 2222 -R …`, which needs no binary and no flag at all, so a default
  only shortens a command for someone who has already cloned and built. Against
  that, omitting a required flag today prints usage and exits 2, whereas with
  this diff `./pumasi-tunnel 8080` **publishes a local port to a relay the user
  never named** — a relay that runs `AllowAll` and serves plaintext (items 1 and
  2). Turning a usage error into an unrequested publication is a change to what
  the product promises; it needs an intent statement and a window, not a
  working-tree edit, and if it is taken it moves README lines 69 and 75 in the
  same commit. **A seat that disagrees should rank it in the order above with
  that cost written down. It should not be committed as found.**

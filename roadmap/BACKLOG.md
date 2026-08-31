# BACKLOG — what gets built next, in order

**Owned by the product-manager role**
([`pumasi-ops/roles/product-manager.md`](https://github.com/pumasi-ai/pumasi-ops/blob/main/roles/product-manager.md),
duty 5). Seeded 2026-08-30 from candidate `0011-developer-tunnels.md`; first
evaluation pass 2026-08-31 at `3652e15`; post-release evaluation 2026-08-31 at
`83fd9f7`; post-`0041` evaluation 2026-08-31 at `1d9505c`; **post-`0047`
evaluation 2026-08-31 at `b3d251d`**, after the test-port merge.

One list, features and bugs together — a priority that cannot compare them is
not a priority. Every entry points at its source and carries one line of
why-here. **The top of this file is what the project manager's next coder
packet builds** — except where an entry says in its own text that it is
operator action rather than a build; the packet then takes the highest entry
that *is* a build, and the operator item keeps its rank rather than being
demoted for being unbuildable.

> **Highest *build* entry: item 2** — the HTTP path hands a visitor's hostname
> to the registry, and the agent its URL, before the session that serves that
> URL exists. That is the next coder packet. Item 1 outranks it and is **not** a
> build: it is operator action blocked on **Q-014**.

Reordering is a commit with the reasoning in the message; the steward vetoes by
reverting.

**What changed in this pass, and why.** The previous item 2 — the suite drawing
its TCP ports from inside the kernel's ephemeral range — is **half delivered**
at `b3d251d`. The half that landed is the whole of what made the gate
unmeasurable, and the effect is large: on this machine `go test -count=1 ./...`
went from **40 failures in 40** at `1d9505c` to **0 failures in 500** at
`b3d251d`. The half that did not land is a latent fixture defect inside a frozen
acceptance case; it is rewritten as item 9 rather than ticked, and the reason it
did not land is a governance reading, not work. See item 9 and
[`STAGE.md`](STAGE.md) §2.

Two things follow, and they are the substance of this pass:

- **The gate is measurable again, and it is clean.** Every figure in this file
  and in `STAGE.md` §2 was re-run by this seat at `b3d251d` on this host. The
  run counts are stated beside every number, because a gate whose number is
  inherited rather than measured is what produced the wrong 12-run reading that
  raised **Q-024**.
- **A new item enters at 2, and it is not the old flake renamed.** Job `0047`
  reported two failures in 240 runs, both `TestConcurrentVisitors`, both
  *"No tunnel is open for myapp.pumasi.link"*. That is the **HTTP** path, not the
  TCP port collision. Reading `relay/relay.go` for the cause found an ordering
  defect of exactly the class `spec/0002` was written to eliminate, still
  present: `r.sessions[resp.AgentID]` is installed **after** the auth response is
  written. This seat could not reproduce it (the runs are cited in item 2), so
  the entry claims a code reading and a symptom that matches it, and does not
  claim a reproduction.

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
  not ask for it. Note that `1d9505c` has now added a *second* merged-but-
  undeployed change behind the same blocker.
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

**2 · The HTTP path hands out a URL before the session that serves it exists** —
**the next coder packet.** Source: this evaluation, from reading `relay/relay.go`
for the cause of the two `TestConcurrentVisitors` failures job `0047` measured in
240 runs and reported as unranked.

`ServeAgent` does four things in this order (`relay/relay.go:139`–`:216`):

1. `authorize` calls `r.registry.Register(tunnel)` (`:295`) — **the hostname now
   resolves**, and `ServeHTTP`'s `r.registry.Lookup` will find it (`:331`).
2. the public TCP port is bound, if the tunnel asked for one (`:175`) — this is
   `0041`'s fix and it is correctly placed.
3. `r.writeFrame(conn, okFrame)` (`:205`) — **the agent, and its user, now have
   the URL.**
4. `r.sessions[resp.AgentID] = session` (`:213`–`:216`) — **only now can a
   visitor be forwarded anywhere.**

Between 3 and 4 a visitor who uses the URL gets past `Lookup`, finds
`session == nil`, and is answered `404 No tunnel is open for <host>` by
`notFound` (`:340`–`:344`, `:384`). That is the exact string `0047` recorded.

**This is the same defect `spec/0002` §1 exists to forbid**, one path over.
That spec's rule is *"when the client learns the outcome, the state behind it is
already true"*, and `0041` applied it to the TCP listener — deliberately moving
`serveTCP` after the announce while making the *bind* precede it, because a
visitor arriving early waits in the accept queue. The mux session has no accept
queue and no such argument: there is nothing between step 3 and step 4 that
needs the response to have been sent, so the insert can simply move above it.
The comment already on the `session == nil` branch — *"the agent went away
between the lookup and here"* — describes only the teardown race and is a
mis-reading of the other way that branch is reached.

**What this seat measured, and what it did not.** It did **not** reproduce the
failure. At `b3d251d` on this host:

| Arm | Runs | Failures |
| :--- | ---: | ---: |
| `go test -count=1 ./...` | **500** | **0** |
| `go test -count=1 -cover ./...` | **100** | **0** |
| `tools/gate.sh` (whole gate, `SKIP_FAMILY_PROBE=1`) | **40** | **0** |

and a targeted probe of the window itself, driving the unmodified `relay` and
`agent` packages from outside the repository — the method job `0042` used for
the TCP race — dialling the edge at the earliest instant `OnConnect` can hand
over the URL: **4,900 visitor requests, 0 answered `404`** (200 iterations × 20
concurrent visitors; 500 single-visitor iterations; 400 more with the full suite
looping alongside as load, which is the condition `0047` said the flake needs).

So: **the defect is established by reading, its incidence on this host is below
what ~5,400 attempts can detect, and `0047`'s two failures in 240 are the only
observation of it.** The entry says that rather than claiming a reproduction.
The window is a few instructions wide and widens under load, which is why a
loaded CI machine sees it and this one did not.

Why here, and why above the `beta` bar: the fix is to move one three-line block
above one `writeFrame` call, it is the last identified source of
non-determinism in the suite the Stage 1 gate names — which is what **Q-024**
turns on — and it is a violation of a rule this product has already adopted and
paid for once. Its user-facing cost today is small and this file will not
inflate it: a person pastes a URL seconds after the agent prints it, not
microseconds. **Fixer: the coder.** A frozen acceptance case that fails at
`b3d251d` and passes after is what would make it more than a reading.

**3 · A subdomain belongs to nobody, and nothing survives a relay restart** —
source: `VALUE.md` claim 2, which sold "permanent" and "stable across
restarts". All three facts re-verified this pass at `1d9505c`:
`Tunnel.Reserved` is computed in `relay/relay.go` and is **still never read
anywhere** (`grep -rn Reserved --include=*.go` finds the field definition at
`core/route.go:127`, that one write, and nothing else); `cmd/pumasi-relay` has
no auth flag, so `AllowAll` is the only authenticator that binary can run; and
the registry and port pool are plain in-memory maps (`core/route.go:145–147`,
`core/portpool.go:27–29`, no persistence path anywhere in `core/` or `relay/`).
A relay restart drops every name, every reserved port and every live tunnel at
once. `--tcp-port` survives an *agent* reconnect, not a relay one.
Why here: this is the `beta` bar itself (`STAGE.md` §4) and it is what retires
**Q-014** — once a restart costs nobody their address, who may deploy stops
being a steward question and becomes an ordinary one. It is the largest single
piece of work on this list, which is why it sits below one three-line
reordering rather than above it. Re-verified in the tree at `b3d251d`: all
three facts below are unchanged by `0047`, which touched one test file.

**4 · A public port the pool believes is free may not be bindable, and the
relay gives up instead of taking the next one** — source: this evaluation,
found while establishing why the previous order's item 2 failed. It is the
product-side half of that port-range defect — whose test-side half is now split
between *Delivered* (`b3d251d`) and item 9 — and it is a different defect from
both. Re-verified in the tree at `b3d251d`; `0047` changed no product code.

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
read 2026-08-31 20:06 UTC), which is *below* the ephemeral floor. So the
kernel cannot steal a port from the live relay, and the exposure there is only
another process on the Vultr host binding into that block. The unbounded case
is the operator who picks a range above 32768 with nothing to warn them, and
the suite, whose remaining half is item 9.

Why here: the failure is honest and self-healing since `1d9505c` — the agent is
refused rather than lied to, and it retries — so this is robustness, not a
falsehood on shipped surface, which is why it sits below item 3 rather than
above it. But it is the reason a 100-port pool can be defeated by one busy port,
and the range guard is a few lines that would have made the previous order's
item 2 impossible to write. **Fixer: the coder.**

**5 · The console never offers the zero-install `ssh -R` command** — source:
`VALUE.md` claim 1 against the live page. The ingress works: a banner grab on
`pumasi.link:2222` returns `SSH-2.0-pumasi-tunnel`. But `relay/dashboard.html`'s
command builder emits only `pumasi-tunnel --relay …` and its "First time here"
panel offers only `git clone && go build` — so the product's headline claim, the
one thing it does that needs nothing installed, is absent from the one page a
visitor sees.
Re-verified at `b3d251d`: `relay/dashboard.html` contains **0** occurrences of
`ssh -R` and **1** of `git clone`.
Why here: it is an afternoon's work on a page that is already live, and it
converts the strongest differentiator from a README sentence into the thing you
can paste. `MARKET.md` §2 makes that differentiator explicit and cited, which
raises what the omission costs.

**6 · `agent/` has no tests** — source: the gate reading; L-006.
`go test ./...` reports **no test files** for `agent`, `cmd/pumasi-relay` and
`cmd/pumasi-tunnel`. The two `cmd` packages are argument parsing and can wait.
`agent/` cannot: it is half of every tunnel — handshake, reconnect, local dial,
stream fan-out — and today the only thing exercising it is `relay`'s
end-to-end tests, which use it as a fixture and assert on the relay's behaviour,
not its. Reconnect and local-dial-failure have no case at all.
Re-verified at `b3d251d`: the three packages still report *no test files*, and
the 100 `-cover` runs of this pass put them at **0.0%** while `relay` rose to
82.0% — that rise came from four existing tests running instead of aborting, not
from any new coverage of `agent`.
Why here: it is the coverage the Stage 1 number does not include, and item 3
will rewrite reconnect behaviour — the tests should exist before that, not
after.

**7 · PR-1 compliance: a version that moves and is user-visible** — source:
[`PRODUCT-RULES.md` PR-1](https://github.com/pumasi-ai/pumasi/blob/worktree-product-rules/PRODUCT-RULES.md)
(v1.0, 2026-08-30; **binds always, from the first commit**; read fresh this pass
and **still only on the unmerged `worktree-product-rules` branch** — **Q-017** —
and its absence from `main` is not compliance — **re-checked this pass:
`git ls-tree` finds it on neither `main` nor `origin/main`**). This product has
no version anywhere: no version file, no `/version`, nothing in the console
footer, nothing in a release note. Re-verified at `b3d251d`, with one detail
that makes the fix smaller than it looks: **the repository root already has a
`package.json`** — added so `pumasi/tools/gate.sh` step 1 finds a suite — and it
has **no `version` field**, which is exactly where PR-1 says the one source of
truth belongs. `core.AuthRequest.ClientVersion` exists as a field
(`core/handshake.go:33`) and the only thing that ever sets it is
`relay/sshingress.go:165`, which fills it with the *ssh client's* version string,
not this product's.
Why here: there are now **three** merges on `main` that behave differently from
the build on the host, and **nothing on the console, in `/_pumasi/status` or in
the logs says which one is answering**. That is `pumasi-booking`'s Q-012 problem
arriving early, and here the answer is a few lines of Go plus a field already in
the wire protocol.

**8 · PR-2 compliance: in-app feedback** — source:
[`PRODUCT-RULES.md` PR-2](https://github.com/pumasi-ai/pumasi/blob/worktree-product-rules/PRODUCT-RULES.md)
(binds at the `beta` promotion; below `beta`, encouraged). Nothing in the
product collects feedback. The reference behaviour is `pumasi-booking`'s
(`service/src/feedback.ts`) — matched in behaviour, not copied.
Why here: it **gates the `beta` promotion**, so it must be built before the
label moves, and it is worth little before items 1–3 make the thing worth
reporting on. The natural home is the console, where a visitor already is.

**9 · A frozen acceptance case still draws its port range from inside the
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

**Why it ranks ninth, below item 3 and below everything buildable.** Two
reasons, and the first is measured rather than argued:

- **Its cost today is nothing, and this pass re-measured that rather than
  carrying `0047`'s figure.** A-10 calls `relay.New` and reads
  `Registry().PublicScheme()` back. `relay.New` passes the range to
  `core.NewPortPool` (`relay/relay.go:124`), which is explicit that it *"does no
  I/O"* (`core/portpool.go:21`–`:22`). **A-10 binds nothing.** And the block is
  genuinely contended on this host right now — `ss -tanp` at 23:44 UTC found
  **12 sockets** inside 34500–34599, including 34504, 34508, 34552, 34556 and
  34560, all `workerd` — while the suite passed **500 runs out of 500**. A latent
  defect that a contended range cannot provoke is a latent defect.
- **Nothing a coder packet could take sits under it.** The work is one constant;
  what blocks it is whether §3 requirement 2's remedy is available to a builder
  at all, given that Part 3 requirement 1 has the builder authoring every spec in
  the first place. Ranking it above a buildable item would point the next packet
  at something that cannot be built without answering that first.

Why here rather than lower: it is a real defect in a frozen file, and the day
A-10 gains an assertion that binds — or `relay.New` gains a bind — it stops being
latent silently. It sits beside item 10, the other entry on this list that is
real, cheap and has no live consequence. **Fixer: the coder, once the reading
resolves**; `SPEC 0002` §6.5 is the precedent on the permissive side, where the
same builder amended a frozen fixture in the open and it stood.

**10 · `relay.New` accepts a `BaseDomain` that `core.NewRegistry` silently
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
three items are a lie on the internet, a URL announced before it can be served,
and the `beta` bar. Re-verified at `b3d251d`: `lookupBreakingDomain` is still
`relay/bindorder_test.go:56` and `relay.New` still tests only for the empty
string.

**11 · Client CLI: the interactive terminal UI** — source: the seeded item 5;
[`docs/ux/incumbent-ux-spec.md`](../docs/ux/incumbent-ux-spec.md) §273 for what
the incumbents' status line shows. `cmd/pumasi-tunnel` prints logs; there is no
live view of requests, response times or status codes, and no released binary
for macOS or Windows.
Why here: it is polish on a path that works, and every item above it is either
something untrue or something missing that a user would rely on.

**12 · Local request inspector on `127.0.0.1:4040`** — source: the seeded item
6; `VALUE.md` (where it is marked *not built*). Re-verified at `b3d251d`: `web/`
still contains **0** entries; there is no listener, no SSE, no replay.
Why here: last of the build items. It is a whole second product surface, and the
honest catalog page (`pumasi-web` `843bdef`) already tells visitors it does not
exist, so nobody is being misled while it waits. `MARKET.md` §2 records that all
three incumbents ship one; that makes it a real gap, not a whim, but not ahead
of anything above.

---

## Delivered

Verified against the tree during this evaluation, not taken from commit
subjects.

- **The test suite's TCP harness moves below the kernel's ephemeral floor** —
  delivered **`b3d251d`**. This was the *first half* of item 2 of the previous
  order; the second half is item 9 above and is **not** delivered.

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
  `cmd/pumasi-relay` and `cmd/pumasi-tunnel` remain **0.0%** — item 6.

  **What this does not do:** it does not retire **Q-024** (that is the steward's,
  and `STAGE.md` §2 records the evidence without closing the window), it does not
  fix item 9, and it is a third merged change waiting behind item 1's undeployed
  restart.

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
  item 4, and its test-range defect is delivered at `b3d251d` for
  `tcp_test.go` with a residual at item 9.**
- **Seeded 5 · Client CLI, in part** — delivered `bf837ee`:
  `cmd/pumasi-tunnel` with `--relay`, `--subdomain`, `--token`, `--host`,
  `--tcp`, `--tcp-port`, `-v`, one static binary. *The interactive TUI and the
  published macOS/Windows binaries are not delivered — item 11. The seed's
  `--http` and `--auth` do not exist: HTTP is the default and the flag is
  `--token`.*
- **Not seeded, delivered anyway** — the relay console at the apex
  (`relay/dashboard.go`, `relay/dashboard.html`, `b3585f6`), shaped by
  `docs/ux/incumbent-ux-spec.md`, live at `http://pumasi.link/` and the Stage 1
  Surface B. Its gap is item 5.

---

## Not on this list, and why

- **`pumasi/catalog.json` has no entry for this product** — zero occurrences of
  the string `tunnel`. It is a real defect and it is recorded in
  [`STAGE.md`](STAGE.md) under known gaps. It is **not** a backlog item here
  because it is not this repository's file and **no role owns it** — that is
  `pumasi/DECISIONS.md` **Q-019**, open. Nothing a coder packet on this product
  could build would close it.
- **Deploying the relay** — **Q-014**, open, and outside CHARTER Part 0. Named
  in item 1(i) as a blocker; not requested of anyone here.
- **Merging `PRODUCT-RULES.md` to `pumasi` main** — **Q-017**, open. This pass
  read the file fresh from the `worktree-product-rules` branch, as items 7 and 8
  record. Not this repository's file.

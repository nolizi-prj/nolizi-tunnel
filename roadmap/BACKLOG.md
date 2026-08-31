# BACKLOG — what gets built next, in order

**Owned by the product-manager role**
([`pumasi-ops/roles/product-manager.md`](https://github.com/pumasi-ai/pumasi-ops/blob/main/roles/product-manager.md),
duty 5). Seeded 2026-08-30 from candidate `0011-developer-tunnels.md`; first
evaluation pass 2026-08-31 at `3652e15`; post-release evaluation 2026-08-31 at
`83fd9f7`; **post-`0041` evaluation 2026-08-31 at `1d9505c`**, after the
bind-before-announce merge.

One list, features and bugs together — a priority that cannot compare them is
not a priority. Every entry points at its source and carries one line of
why-here. **The top of this file is what the project manager's next coder
packet builds** — except where an entry says in its own text that it is
operator action rather than a build; the packet then takes the highest entry
that *is* a build, and the operator item keeps its rank rather than being
demoted for being unbuildable.

> **Highest *build* entry: item 2** — the suite draws its TCP ports from inside
> the kernel's ephemeral range. That is the next coder packet. Item 1 outranks
> it and is **not** a build: it is operator action blocked on **Q-014**.

Reordering is a commit with the reasoning in the message; the steward vetoes by
reverting.

**What changed in this pass, and why.** Item 2 of the previous order — the
public TCP address announced before anything listens on it — **is delivered at
`1d9505c`** and moves to *Delivered*, with the fix verified by measurement
rather than by its commit message. But re-measuring it did not return the clean
result the fix's own report recorded, and the difference is the substance of
this pass:

- **The fix works.** Measured against the ordering defect in isolation, on a
  port range the host cannot steal, the race goes from **28 failures in 2000**
  at `83fd9f7` to **0 in 2000** at `1d9505c`.
- **The suite is still not deterministic, for a different reason that was
  always there.** `relay/tcp_test.go` pins its TCP range to **34000–34099**,
  which lies **inside** this machine's `/proc/sys/net/ipv4/ip_local_port_range`
  (**32768–60999**). Any unrelated process holding one of those ports as an
  outbound ephemeral source port takes it from the relay. On this host, on the
  day of this pass, one did — so `go test -count=1 ./...` fails **every run**,
  at both SHAs.

That second finding is new, it is not a regression from `0041`, and it is now
item 2. It also means the Stage 1 gate's number is a property of the machine it
was measured on; [`STAGE.md`](STAGE.md) §2 says so and records the evidence
against **Q-024**.

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

**2 · The test suite draws its TCP ports from inside the kernel's ephemeral
range, so the gate's number belongs to the machine and not to the code** —
**the next coder packet.** Source: this evaluation, from re-measuring item 2 of
the previous order and failing to reproduce a clean run at either SHA.

`relay/tcp_test.go:47–48` fixes the relay's pool at **34000–34099**;
`relay/scheme_test.go:314–315` uses **34500–34599**. This machine's
`/proc/sys/net/ipv4/ip_local_port_range` is **32768 60999**, so both blocks sit
inside the range the kernel hands out as *source* ports for outgoing
connections. `core.PortPool` starts its cursor at `low`
(`core/portpool.go:43`, `cursor: low`), so a freshly built relay always
draws **34000 first** — and every historical failure this file has ever
recorded names that exact port.

Measured during this pass:

```
$ cat /proc/sys/net/ipv4/ip_local_port_range
32768	60999

$ ss -tanp | grep :34000
ESTAB 0 0 127.0.0.1:34000 127.0.0.1:43461 users:(("workerd",pid=716041,fd=455))
```

An unrelated process held 34000 for the whole of this evaluation, so the bind
failed every time and the four TCP tests failed every run at both SHAs. The
symptom differs by SHA, which is the fix working exactly as designed:

| | `83fd9f7` (before) | `1d9505c` (after) |
| :--- | :--- | :--- |
| Failure | `dial tcp 127.0.0.1:34000: connect: connection refused` | `agent did not connect` (handshake refused) |
| Where | the visitor's dial, *after* the address was announced | the handshake, *before* anything was announced |

**`0041` already knew about this hazard and fixed only its own half.**
`relay/bindorder_test.go:39–43` sets `bindOrderBase = 20500` and says why, in
the coder's own words: *"Deliberately below `/proc/sys/net/ipv4/ip_local_port_range`'s
floor (32768 on this machine). A fixed test port inside the ephemeral range can
be taken transiently by any outgoing connection on the host, which makes a bind
failure look like the defect under test."* The comment at `:35–37` then says the
new block was chosen so it *"cannot collide with the 34000-series harness in
tcp_test.go"* — routing around the older tests rather than moving them. The new
cases are safe; the three historical TCP tests and `scheme_test.go` are not.

Fix: move both blocks below 32768, as `bindorder_test.go` already does, or draw
them from a port the OS has agreed to give up. **Fixer: the coder** (this seat
may not edit tests). Why here, and why above the `beta` bar: it is two constants
and a comment, it is the cheapest item on this list, and **until it lands nobody
can measure the Stage 1 exit gate at all** — the gate's "passes 100%" is
currently a statement about whether anything else on the host happened to be
using a port. That is the question **Q-024** is open on, and this is what makes
it answerable.

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
piece of work on this list, which is why it sits below one two-constant fix
rather than above it.

**4 · A public port the pool believes is free may not be bindable, and the
relay gives up instead of taking the next one** — source: this evaluation,
found while establishing why item 2 fails. This is the product-side half of
item 2 and it is a different defect.

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
the suite, which is item 2.

Why here: the failure is honest and self-healing since `1d9505c` — the agent is
refused rather than lied to, and it retries — so this is robustness, not a
falsehood on shipped surface, which is why it sits below item 3 rather than
above it. But it is the reason a 100-port pool can be defeated by one busy port,
and the range guard is a few lines that would have made item 2 impossible to
write. **Fixer: the coder.**

**5 · The console never offers the zero-install `ssh -R` command** — source:
`VALUE.md` claim 1 against the live page. The ingress works: a banner grab on
`pumasi.link:2222` returns `SSH-2.0-pumasi-tunnel`. But `relay/dashboard.html`'s
command builder emits only `pumasi-tunnel --relay …` and its "First time here"
panel offers only `git clone && go build` — so the product's headline claim, the
one thing it does that needs nothing installed, is absent from the one page a
visitor sees.
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
Why here: it is the coverage the Stage 1 number does not include, and item 3
will rewrite reconnect behaviour — the tests should exist before that, not
after.

**7 · PR-1 compliance: a version that moves and is user-visible** — source:
[`PRODUCT-RULES.md` PR-1](https://github.com/pumasi-ai/pumasi/blob/worktree-product-rules/PRODUCT-RULES.md)
(v1.0, 2026-08-30; **binds always, from the first commit**; read fresh this pass
and **still only on the unmerged `worktree-product-rules` branch** — **Q-017** —
and its absence from `main` is not compliance). This product has no version
anywhere: no version file, no `/version`, nothing in the console footer, nothing
in a release note. `core.AuthRequest.ClientVersion` exists as a field and no
binary sets it.
Why here: there are now **two** merges on `main` that behave differently from
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
three items are a lie on the internet, an unmeasurable gate and the `beta` bar.

**10 · Client CLI: the interactive terminal UI** — source: the seeded item 5;
[`docs/ux/incumbent-ux-spec.md`](../docs/ux/incumbent-ux-spec.md) §273 for what
the incumbents' status line shows. `cmd/pumasi-tunnel` prints logs; there is no
live view of requests, response times or status codes, and no released binary
for macOS or Windows.
Why here: it is polish on a path that works, and every item above it is either
something untrue or something missing that a user would rely on.

**11 · Local request inspector on `127.0.0.1:4040`** — source: the seeded item
6; `VALUE.md` (where it is marked *not built*). `web/` is an empty directory;
there is no listener, no SSE, no replay.
Why here: last of the build items. It is a whole second product surface, and the
honest catalog page (`pumasi-web` `843bdef`) already tells visitors it does not
exist, so nobody is being misled while it waits. `MARKET.md` §2 records that all
three incumbents ship one; that makes it a real gap, not a whim, but not ahead
of anything above.

---

## Delivered

Verified against the tree during this evaluation, not taken from commit
subjects.

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
  machine, for the ephemeral-port reason that is now item 2 and is not this
  change's doing. See [`STAGE.md`](STAGE.md) §2.
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
  item 4 and its test-range defect is item 2.**
- **Seeded 5 · Client CLI, in part** — delivered `bf837ee`:
  `cmd/pumasi-tunnel` with `--relay`, `--subdomain`, `--token`, `--host`,
  `--tcp`, `--tcp-port`, `-v`, one static binary. *The interactive TUI and the
  published macOS/Windows binaries are not delivered — item 10. The seed's
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

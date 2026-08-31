# STAGE — Pumasi Tunnel

**Current Stage:** `Alpha` — unchanged by this evaluation, and §4 says what it
would cost to move.
**Stage 1 exit gate:** **MET 2026-08-31**, carried forward — but the reading is
now **weaker than it was**, not stronger, and §2 says exactly why. The gate does
**not** rest on 40 clean runs. It rests on 40 runs that could not be taken.
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

**Re-evaluated 2026-08-31 at `1d9505c`** (post-`0041`, bind-before-announce),
`STAGE_PLAYBOOK.md` Event 2. The previous pass was at `83fd9f7` and its text is
kept below where it still holds. Every number in §2 was taken in this pass.

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

| | on `main` @ `1d9505c` | on `pumasi.link`, re-measured **20:06 UTC 2026-08-31** |
| :--- | :--- | :--- |
| Scheme announced | `http://` by default, `-public-scheme=https` when a terminator is in front | **`https://`** — `"url":"https://sshsteward.pumasi.link"` |
| Port 443 | not the relay's job, by design | **refused** — `curl: (7) Failed to connect to pumasi.link port 443` |
| Port 80 | serves the console and visitor traffic | **200** |
| ssh ingress 2222 | serves | **`SSH-2.0-pumasi-tunnel`** |
| TCP range in use | `-tcp-low`/`-tcp-high`, no default | **20000–20099** — `"tcp_range"`, below the ephemeral floor (§3.1a) |
| Which build is answering | — | **unknowable from any surface** — see §5 |

So: **a person who opens a tunnel today is still handed an `https://` address
that nothing answers.** The fix for that is merged and undeployed. The relay
process on the Vultr host predates `83fd9f7` and will keep announcing `https://`
until it is restarted — and as of `1d9505c` there are now **two** merged changes
waiting behind that same restart, not one.

**Why it has not been restarted, stated rather than left unsaid.**
`pumasi/DECISIONS.md` **Q-014** is open: it asks who may restart a relay that
keeps no durable state and whose single live tunnel is `sshsteward` —
`pumasi.link:20000` → this machine's port 22, `pumasi-ops/RESOURCES.md` §4's
remote-access route, **open 49 671 s (13 h 48 m) and unbroken** at the
measurement above (`"opened_at":"2026-08-31T06:18:13Z"`). Q-014 is **explicitly outside CHARTER Part 0's
proceed-on-default rule**. This evaluation therefore did not deploy, did not
treat the deploy as a judgement call, and **does not ask a coder packet to take
it either**. It is named as a blocker in `BACKLOG.md` item 1(i) and it waits for
the steward.

---

## 2 · Maturity gates

| Stage | Criteria | Status |
| :--- | :--- | :--- |
| **0. Candidate** | Scored by 3 model families (40/60) and selected by steward | **COMPLETE** |
| **1. Alpha** | Pure-core suite passes 100%; both public landing surfaces live | **COMPLETE 2026-08-31** — but see the re-reading below: on this machine the suite passed **0 runs in 40** at `1d9505c`, for a reason that is not the product |
| **2. Beta** | Real end-to-end users complete workflows without engineer intervention | **IN PROGRESS** — §4 |
| **3. Launched** | Production hardening, cross-model regression, 7-day veto window | PENDING |

### The Stage 1 gate, re-read at `1d9505c` — and what the re-reading found

**The short answer to the question this evaluation was asked.** `0041` fixed
the defect it was sent to fix, and this seat verified that independently. It
did **not** make the suite deterministic, and the Stage 1 gate reading now
rests on **less** than it did, not more. The gate stays `MET` under Q-024's own
named default, because that default says the number must be written down beside
it; the number is below and it is not a good one.

**What was measured, by this seat, on this machine, this pass.**

*(1) The suite command the gate names.* `go test -count=1 ./...`, 40 runs at
each SHA, run sequentially:

| SHA | Failures |
| :--- | :--- |
| `83fd9f7` (before `0041`) | **40 in 40** |
| `1d9505c` (after `0041`) | **40 in 40** |

Neither figure reproduces anything previously recorded — not the 3-in-40
baseline this file carried, and not the 0-in-40 that `0041` reported. **The
cause is not the product and not `0041`.** `relay/tcp_test.go:47–48` pins the
relay's TCP pool to **34000–34099**, and this machine's
`/proc/sys/net/ipv4/ip_local_port_range` is **32768 60999** — the test block
lies inside the range the kernel hands out as source ports for outgoing
connections. Throughout this evaluation an unrelated process held 34000:

```
$ ss -tanp | grep :34000
ESTAB 0 0 127.0.0.1:34000 127.0.0.1:43461 users:(("workerd",pid=716041,fd=455))
```

`core.PortPool` starts its cursor at `low` (`core/portpool.go:43`), so every
freshly built relay draws 34000 first and every TCP test failed. That is
`BACKLOG.md` item 2, it is now the **highest build entry**, and it is two
constants.

The *symptom* differs by SHA, and the difference is `0041` working correctly:
at `83fd9f7` the address was announced and the visitor's dial was refused
(`dial tcp 127.0.0.1:34000: connect: connection refused` — the exact string
this file previously published as evidence of the race); at `1d9505c` the bind
is attempted first, fails, and the handshake is refused before any address
leaves (`agent did not connect`). The product stopped lying about the address.
The suite did not stop being a function of what else is running on the host.

*(2) The ordering defect itself, isolated from the host.* Because (1) cannot
distinguish a fixed race from a stolen port, the race was measured directly:
the same unmodified `relay` and `agent` packages, driven from outside the
repository, on a TCP range **outside** the ephemeral range (14000–14099,
confirmed 100/100 bindable), taking the announced `TCPAddr` and dialling it the
instant it arrived.

| Iterations | `83fd9f7` (before) | `1d9505c` (after) |
| :--- | :--- | :--- |
| 200 | 3 dial refusals | **0** |
| 500 | 5 dial refusals | **0** |
| 2000 | **28 dial refusals** (1.4%) | **0** |

**So the race was real, and it is gone.** That is a genuine improvement to the
product and it is recorded as delivered in `BACKLOG.md`.

**What this does and does not do to the gate.**

- It **does** retire the specific defect Q-024 named as its retirement
  condition — "bind before the response leaves, on both the agent and the ssh
  paths". That half is done and verified.
- It **does not** deliver the other half of the same condition — *"and 40 clean
  runs of each invocation recorded in `roadmap/STAGE.md`"*. There are no clean
  runs to record. This file will not write "40 clean runs"; it writes 0 in 40
  and the reason.
- The gate therefore **stays `MET` and stays qualified**, exactly as Q-024's
  named default provides for, but the qualification has changed shape. It is no
  longer "passes, 37 times in 40". It is: **the pure-core suite passes on a host
  where nothing else is using 34000–34099, and this file cannot presently say
  how often that is true of an arbitrary host.** That is a weaker claim than the
  one this file made last pass, and it is the honest one.
- **Q-024's rider (a) still binds and binds harder.** No stage-promotion
  announcement may be published off this gate — `STAGE_PLAYBOOK.md` Event 3 stays
  held, §7. The rate is not non-zero-but-small; on this machine today it is
  total.
- **The steward closes Q-024, not this seat.** This is evidence recorded under
  an open window, and the window's own retirement condition is now partly met
  and partly refuted. Both are stated in `pumasi/DECISIONS.md` under that entry.

**Coverage, at `1d9505c`.** Not re-measured this pass: `-cover` cannot produce a
meaningful figure for `relay` while four of its tests abort on a port collision.
The figures this file carried at `83fd9f7` — `core` 80.3%, `mux` 84.0%, `relay`
74.7% — are left standing as the last trustworthy reading, and marked as
inherited rather than re-measured. Re-measuring them is downstream of
`BACKLOG.md` item 2.

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

## 3 · What the gate's number does not cover — three flags now, and one is retired

The Stage 1 gate is `go test` exiting 0. The gate's number is only as good as
what it covers (L-006), so what it does not cover is recorded here rather than
left implicit. Every entry below is a `BACKLOG.md` item. The flag this section
carried first — the announce-before-bind race — is **retired by measurement**
this pass and §3.1 now records what took its place.

### 3.1 · The race is fixed; the suite's own port choice is what is left

**The race is delivered.** `relay/relay.go` now binds the public port before it
encodes and writes the auth response (`:169–190`, `listenTCP` at `:175`,
`writeFrame(conn, okFrame)` at `:205`), and `relay/sshingress.go` carries the
same ordering. A bind failure unregisters the agent and answers the handshake
with an error frame instead of correcting an address the user already has.
Measured at 2000 iterations with the host's interference removed: **28 dial
refusals before, 0 after** (§2). `BACKLOG.md` records it as delivered at
`1d9505c`.

**What replaces it as the gate's qualification** is not a product defect at all,
which is why it needs saying plainly rather than being folded into the old
entry: `relay/tcp_test.go:47–48` (34000–34099) and `relay/scheme_test.go:314–315`
(34500–34599) draw from inside the kernel's ephemeral range, so any process on
the host can take the port the relay is about to bind. `0041` identified this
hazard and fixed only the cases it wrote — `relay/bindorder_test.go:39–43` sets
`bindOrderBase = 20500`, *"Deliberately below `/proc/sys/net/ipv4/ip_local_port_range`'s
floor"* — and chose the new block so it would not collide with the 34000-series
harness rather than moving that harness. `BACKLOG.md` item 2, **the highest
build entry**. **Fixer: the coder.**

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

### 3.2 · Three packages have no test files, `agent/` among them

`agent`, `cmd/pumasi-relay` and `cmd/pumasi-tunnel` report *no test files* and
0.0% coverage. `core`, `mux` and `relay` are the pure core the gate names and
they are covered. `agent/` is not core, but it is half of every tunnel, and
today nothing exercises it except the relay's end-to-end tests using it as a
fixture. [`BACKLOG.md`](BACKLOG.md) item 5. **Fixer: the coder.**

---

## 4 · Why not `beta`

`beta` means strangers can rely on it and their data survives. Each fact below
was re-verified against the tree at `83fd9f7` in this pass. The order is the
backlog's.

1. **There is no TLS, and what is running still says there is.** *Changed by
   the release, and only halfway.* On `main` the relay now announces the scheme
   it serves — `-public-scheme`, default `http`, validated once at startup,
   applied in one place, read by all three surfaces. **On `pumasi.link`, nothing
   listens on 443 and the relay still prints `https://`** (§1). Even once
   deployed, every HTTP tunnel here is plaintext and no `https://`-only webhook
   sender can be pointed at one. [`BACKLOG.md`](BACKLOG.md) item 1 — **operator
   action plus a deploy blocked on Q-014, not a build.**
2. **A name belongs to nobody.** Still true. `Tunnel.Reserved` is computed at
   `relay/relay.go:249` (it was line 236; the line moved, the fact did not) and
   is **never read anywhere in the tree** — the field is defined at
   `core/route.go:127`, written once, and read nowhere. The relay binary exposes
   **eleven flags and no auth flag**, so `AllowAll` (`relay/relay.go:40`,
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
changed by this pass**: item 2 (the suite's ephemeral-range port block) first,
because it is two constants and because **nothing else on this list can be
measured until it lands** — including the gate in §2 and the coverage figures
§2 now has to carry as inherited rather than re-measured; then item 3, because
it is the whole of facts 2 and 3 above and it is what **retires Q-014** — once a
restart costs nobody their address, deploying stops being a steward question;
then item 4, the relay's own bindability gap, bounded in production today only
because the deployed range happens to sit below the ephemeral floor; item 1 runs
in parallel and needs no coder, only a decision and a certificate.

**What the previous pass put first here is done.** The announce-before-bind race
was this section's item 2 and it is delivered at `1d9505c`, verified at 2000
iterations (§2). The distance to `beta` is shorter by exactly that much and by
nothing else: none of the three numbered facts above changed.

**Also gating, from `PRODUCT-RULES.md`** (v1.0, read fresh 2026-08-31; still
only on the unmerged `worktree-product-rules` branch, `0115758` — **Q-017** —
and its absence from `main` is not compliance): **PR-1** (a user-visible version
number) binds **always** and this product has none anywhere — which §5 shows is
no longer theoretical; **PR-2** (in-app feedback) binds at the `beta` promotion
and is unbuilt. `BACKLOG.md` items 6 and 7.

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
is a diagnostic accident, not a capability. [`BACKLOG.md`](BACKLOG.md) item 6.
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
  no `items[]` entry — verified at `pumasi` `6489347`. `README.md` tells every
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
- **`go test ./...` may fail on your machine for reasons that are not this
  product.** `alpha` means builders, so builders should know: the suite pins its
  TCP ports to 34000–34099, inside the kernel's ephemeral range, and any other
  process holding one of those ports turns four tests red. On the machine this
  pass ran on, that was every run. `BACKLOG.md` item 2 (§2, §3.1).

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
   2222, and has carried a tunnel for over ten hours. `tunnel.pumasi.ai` indeed
   does not resolve; the hosted relay is on the other domain.
3. **[`MARKET.md`](MARKET.md) now exists** and is the only place a public page
   may take a competitor claim from. Every figure in it carries a vendor URL and
   a fetch date, and its §4 records the two comparisons that go *against* this
   product. A page that states a competitor price without that citation is the
   failure `pumasi-booking` `0d1674d` already had to undo.
4. **Nothing public may say the `https://` problem is fixed for users.** It is
   fixed on `main` and not on the internet (§1), and the distinction is the
   whole of this file's §1.
5. **`STAGE_PLAYBOOK.md` Event 3 is still held, and the reason got stronger.**
   That trigger chains a stage-promotion announcement and a public badge update
   off a product manager confirming an exit gate `MET`. This evaluation kept
   Stage 1 at `MET`, but the ordinary invocation `go test -count=1 ./...` passed
   **0 runs in 40** on the machine this pass ran on (§2) — not the 3-in-40 this
   file previously published, and not the 0-in-40 the fix reported. The cause is
   the suite's port block, not the product, and it is `BACKLOG.md` item 2.
   **Nothing public may be published off this gate**, and any page still quoting
   *"passes, 37 times in 40"* is quoting a number this file has now withdrawn.
   Nothing is lost by waiting: this product is `alpha`, is not asking to move,
   and §4 lists three separate reasons it should not. The reading remains
   escalated as `pumasi/DECISIONS.md` **Q-024**, where this pass recorded its
   evidence without closing the window.

6. **One thing a page may now say that it could not before.** The
   announce-before-bind race is fixed and measured — 28 dial refusals in 2000
   before, 0 after (§2). That is a real product improvement and it is publishable
   **as a fix**, provided the page does not imply it made the test suite green,
   does not quote a Stage 1 pass rate, and does not say it is live: like
   everything else on `main`, it sits behind the undeployed restart in §1.

# STAGE — Pumasi Tunnel

**Current Stage:** `Alpha` — unchanged by this evaluation, and §4 says what it
would cost to move.
**Stage 1 exit gate:** **MET 2026-08-31**, re-measured at `83fd9f7` (§2).
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

| | on `main` @ `83fd9f7` | on `pumasi.link`, measured 16:21 UTC 2026-08-31 |
| :--- | :--- | :--- |
| Scheme announced | `http://` by default, `-public-scheme=https` when a terminator is in front | **`https://`** — `"url":"https://sshsteward.pumasi.link"` |
| Port 443 | not the relay's job, by design | **refused** — `curl: (7) Failed to connect to pumasi.link port 443` |
| Port 80 | serves the console and visitor traffic | **200** |
| ssh ingress 2222 | serves | **`SSH-2.0-pumasi-tunnel`** |
| Which build is answering | — | **unknowable from any surface** — see §5 |

So: **a person who opens a tunnel today is still handed an `https://` address
that nothing answers.** The fix for that is merged and undeployed. The relay
process on the Vultr host predates `83fd9f7` and will keep announcing `https://`
until it is restarted.

**Why it has not been restarted, stated rather than left unsaid.**
`pumasi/DECISIONS.md` **Q-014** is open: it asks who may restart a relay that
keeps no durable state and whose single live tunnel is `sshsteward` —
`pumasi.link:20000` → this machine's port 22, `pumasi-ops/RESOURCES.md` §4's
remote-access route, **open 36 167 s (10 h 2 m) and unbroken** at the
measurement above. Q-014 is **explicitly outside CHARTER Part 0's
proceed-on-default rule**. This evaluation therefore did not deploy, did not
treat the deploy as a judgement call, and **does not ask a coder packet to take
it either**. It is named as a blocker in `BACKLOG.md` item 1(i) and it waits for
the steward.

---

## 2 · Maturity gates

| Stage | Criteria | Status |
| :--- | :--- | :--- |
| **0. Candidate** | Scored by 3 model families (40/60) and selected by steward | **COMPLETE** |
| **1. Alpha** | Pure-core suite passes 100%; both public landing surfaces live | **COMPLETE 2026-08-31**, re-measured at `83fd9f7` |
| **2. Beta** | Real end-to-end users complete workflows without engineer intervention | **IN PROGRESS** — §4 |
| **3. Launched** | Production hardening, cross-model regression, 7-day veto window | PENDING |

### The Stage 1 gate, re-measured at `83fd9f7`

**Pure core passes, and covers more than it did.** `go test -count=1 -cover ./...`
exits 0. Coverage, against the figures this file carried at `3652e15`:

| Package | at `3652e15` | **at `83fd9f7`** |
| :--- | :--- | :--- |
| `core` | 79.5% | **80.3%** |
| `mux` | 83.5% | **84.0%** |
| `relay` | 55.9% | **74.7%** |

The `relay` jump is `relay/scheme_test.go`, which the `-public-scheme` release
brought with it. That is a release improving the gate's own evidence, and it is
worth recording as such.

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

## 3 · The gate's number covers one flag, and now covers two

The Stage 1 gate is `go test` exiting 0. The gate's number is only as good as
what it covers (L-006), so what it does not cover is recorded here rather than
left implicit. Both entries below are `BACKLOG.md` items.

### 3.1 · A live race, and the previous measurement of it was too kind

This file previously recorded the suite as non-deterministic **under `-cover`**,
2 failures in 12, with 12 of 12 clean without it. **That reading does not
survive a larger sample.** Re-measured here at `83fd9f7`, 40 runs of each
invocation:

| Invocation | Failures |
| :--- | :--- |
| `go test -count=1 -cover ./...` | **3 in 40** |
| `go test -count=1 ./...` | **3 in 40** |

Coverage instrumentation is **not** the trigger. It is a **7.5% failure rate on
`main`'s ordinary test command**, and the earlier clean run of 12 was a sample
that happened to miss it. The failure now surfaces on three different tests —
`TestRawTCPCrossesTheTunnel`, `TestTCPPortReleasedWhenAgentDisconnects` and
`TestConcurrentTCPClients` — every one of them the same dial refused against
the same public TCP port, which is itself the evidence that it is not one
test's artifact:

```
--- FAIL: TestTCPPortReleasedWhenAgentDisconnects
    tcp_test.go:208: port should be live before disconnect:
                     dial tcp 127.0.0.1:34000: connect: connection refused
```

**The defect is in the product, not the tests.** `relay/relay.go` writes the
auth response carrying `TCPAddr` at line **175** and only calls `bindTCP` at
line **194**; `relay/sshingress.go:182` has the same shape. An agent — and the
person reading its output — is handed a public address before anything listens
on it, and a bind failure is reported only after the address was announced.

**Fixer: the coder.** This role may not edit product code. It is
[`BACKLOG.md`](BACKLOG.md) item 2, the highest buildable entry, and therefore
the next coder packet on this product.

**The Stage 1 gate stays MET.** It is recorded as a qualification, not a
failure: the gate asks that the pure-core suite pass, and it does — the passing
run is reproducible and the defect is a race that a passing run does not exhibit.
But the honest form of "100%" here is *"passes, 37 times in 40"*, and that is
now written down.

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
order and its reasoning in one line each: item 2 (the race) first because it is
cheap and it is wrong on shipped surface; then item 3, because it is the whole
of facts 2 and 3 above and it is what **retires Q-014** — once a restart costs
nobody their address, deploying stops being a steward question; item 1 runs in
parallel and needs no coder, only a decision and a certificate.

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
5. **`STAGE_PLAYBOOK.md` Event 3 is held, not fired.** That trigger chains a
   stage-promotion announcement and a public badge update off a product manager
   confirming an exit gate `MET`. This evaluation kept Stage 1 at `MET`, but on
   a suite that fails **3 runs in 40** in its ordinary invocation (§3.1), so no
   promotion announcement should be published off that reading while the rate is
   non-zero. Nothing is lost by waiting: this product is `alpha`, is not asking
   to move, and §4 lists three separate reasons it should not. The reading
   itself is escalated as `pumasi/DECISIONS.md` **Q-024**, with the strict
   alternative and its cost written out; the rate is re-measured at each
   evaluation rather than inherited, which is what produced the wrong 12-run
   number in the first place.

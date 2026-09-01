# Spec 0004 · A name with an owner, and what a restart may cost

**Status:** frozen at spec review · **Intent:** [`INTENT.md`](INTENT.md) ·
**Backlog:** `roadmap/BACKLOG.md` item 2, *"A subdomain belongs to nobody, and
nothing survives a relay restart"* (post-`1853218` order).

**This spec is delivered in slices and says so in §7.** The rule in §1 is
whole; slice 1 delivers its first sentence and this packet builds only that.
The second sentence is slice 2 and **is not built here**. §4's table marks every
row with the slice that makes it true, so no reader has to infer which half is
running.

## 1 · The rule

> A name, and the public TCP port that goes with it, belong to whoever first
> proved they held them, and go on belonging to that party while nobody is
> connected. That ownership outlives the relay process, so what a restart costs
> is the connection and never the address.

## 2 · Three words that are not synonyms, and the whole defect in one line

| Word | Means | Lives in | Dies when |
|---|---|---|---|
| **registered** | a tunnel is live *right now* | `core.Registry` | the agent's connection ends |
| **reserved** | a name may be used by *one party only* | `core.Reservations` (new) | slice 1: when the relay stops. slice 2: on idle expiry, and not before |
| **durable** | the reservation set outlives the *process* | a store (slice 2) | never |

The tree at `1853218` has the first and nothing else. `Tunnel.Reserved` is a
field named for the second whose value is computed from the shape of a request
(`req.Subdomain != "" && req.Token != ""`) rather than from any record, and
which nothing reads. **A relay that answers "is this name in use?" and never
answers "whose is it?" cannot keep a promise about restarts, because the second
question is the one a restart destroys the answer to.**

These stay two questions and two types. `Registry.Register` already refuses a
name that is live (`ErrNameTaken`); a reservation does **not** override that. An
owner reconnecting while its previous connection is still registered is refused
exactly as today, and that is correct: the reservation says who *may*, the
registry says who *is*. Folding them into one table is how a rule forks
([L-007](https://github.com/pumasi-ai/pumasi/blob/main/lessons/L-007-restating-a-rule-forks-it.md)).

## 3 · What a token proves

**It proves one thing: the party presenting it now is the party that presented
it before.** Continuity, not identity.

It does **not** prove:

- **who anyone is.** There is no account and no signup — **Q-002** — and no
  human is named by a reservation. "Owner" in this document means *the holder of
  a secret*, and nothing more.
- **that the bearer may use this relay at all.** That stays the
  `relay.Authenticator`'s question, and `AllowAll` still answers it yes. A
  reservation narrows *which name* an accepted agent may have; it does not
  decide whether the agent is accepted.
- **that the relay is who it claims to be.** There is no TLS on the agent
  control connection — `roadmap/BACKLOG.md` item 1 — so an agent cannot
  authenticate the relay it hands a token to.

**It is a bearer secret on a plaintext connection, today.** Anyone on the path
between agent and relay can read it and replay it. That is a real limit of this
design as deployed, it is bounded by item 1(ii) and not by this spec, and §8
records it as not covered rather than leaving it to be discovered.

### 3.1 · Stored as a digest, and why a plain hash is the right one

A reservation records **`sha256(token)` in hex**, never the token, and compares
with `crypto/subtle.ConstantTimeCompare`. So a leaked reservation set is not a
set of live credentials, and a comparison does not leak a prefix by timing.

A reviewer will reasonably ask why not bcrypt or argon2. **Because a
password-hashing function exists to make a *low-entropy* secret expensive to
guess, and this spec forbids low-entropy secrets instead.** `core.MinTokenLen`
is **16** characters and a shorter token is refused at the point of claim, not
silently accepted and weakly hashed. Against a ≥16-character secret an offline
attack on SHA-256 is not the cheapest way in — the plaintext wire is (§3) — so a
100 ms KDF on every handshake would buy nothing and cost a handshake budget.
**If `MinTokenLen` is ever lowered, this paragraph is wrong and the hash has to
change with it.**

## 4 · What survives what — three things, three different answers

The backlog entry names three things and this spec refuses to give them one
promise. Read this table as the product's actual claim.

| | agent reconnect, relay up | **relay restart** | host restart |
|---|---|---|---|
| **the name** | today: only if nobody took it in the gap. **slice 1: held for the owner** | today: no. **slice 2: yes** | as relay restart |
| **the public TCP port** | today: only if it is still free. **slice 1: held for the owner** | today: no. **slice 2: yes** | as relay restart |
| **the live tunnel** | **never, in any slice** | **never, in any slice** | **never** |

Three things follow and each one is load-bearing.

- **The live tunnel never survives anything, and no slice of this work changes
  that.** A TCP connection cannot outlive the process at either end. What
  recovers it is the agent's own reconnect loop (`agent/agent.go`: one second,
  doubling to thirty) and, for the steward's route,
  `pumasi-ops/tools/pumasi-tunnel-keepalive.sh` on a five-minute cron. **The
  promise this product can honestly make is bounded-outage-then-same-address.**
  `VALUE.md` claim 2's *"stable across restarts"* is true of the address and can
  never be true of the connection; that distinction belongs on the shipped
  surface and is handed to the product manager rather than edited here.
- **`--tcp-port` survives an *agent* reconnect and not a *relay* one, and even
  the first is luck rather than a guarantee.** Today the agent re-requests the
  same number and `AllocateSpecific` succeeds because nothing else happened to
  take it. That is the top-left cell, and slice 1 is what turns it from luck
  into a rule. **No test that reconnects an agent is evidence about the middle
  column.** The acceptance case for the middle column builds a second
  `relay.Relay` over the same store and is named as such in
  [`acceptance/CASES.md`](acceptance/CASES.md) — it is a **slice 2** case and it
  is not built by this packet.
- **The middle column is the one Q-014 turns on, and slice 1 does not fill
  it.** What slice 1 does is stop the *loss* a restart cannot undo (§[INTENT] —
  a restart's cost is already bounded by the keepalive, unless the name is gone
  when the agent returns). Slice 2 fills the column. Saying slice 1 retires
  Q-014 would be false and this spec does not say it.

## 5 · The change · slice 1

### 5.1 · `core/reservation.go` — new, pure, no I/O

```go
// Reservation is a claim on a name that outlives the connection using it.
type Reservation struct {
    Subdomain string // the claimed name, already validated and lowercased
    TokenHash string // hex sha256 of the bearer secret; never the secret
    TCPPort   int    // the public port claimed with the name; 0 for none
}

type Reservations struct { ... } // map[string]Reservation under a sync.RWMutex
```

Every field is read by §5.3. **There is deliberately no `CreatedAt` and no
`LastSeen`.** Slice 2 needs `LastSeen` for expiry and adds it in the same change
that reads it. Shipping a field ahead of its reader is precisely the defect this
spec exists to repair (§2), and it would be a poor spec that reintroduced it one
type over.

Operations, and the whole of the policy:

- `Claim(subdomain, token string, tcpPort int) (created bool, err error)` —
  trust on first use. `created` reports whether *this call* recorded a new
  reservation, so the caller can undo it if the handshake it belongs to does
  not complete.
- `Check(subdomain, token string) error` — the read-only form: does this caller
  may-have this name. `token` may be empty, which is what an anonymous agent
  presents.
- `Discard(subdomain string)` — removes a claim. **Not a user-facing release:**
  it exists only so a handshake that fails after claiming leaves nothing
  behind, and nothing outside `relay` calls it. Slice 1 has no other way for a
  reservation to end; slice 2's idle expiry is the first.
- `PortHolder(port int) string` — the subdomain holding a port, or `""`.
- `Get(subdomain string) (Reservation, bool)` — what sets `Tunnel.Reserved`.

**The policy, exhaustively.** A token that is present and shorter than
`MinTokenLen` is `ErrTokenTooShort` from both `Claim` and `Check`, and the relay
returns that refusal to the agent rather than degrading it to anonymous.
Silently treating a short token as no token would hand a name to a caller who
believes they are protecting it, which is the worse of the two failures.

| state of the name | `tcpPort` asked for | result |
|---|---|---|
| unclaimed | `0` | claim the name, no port; `created` |
| unclaimed | *P*, held by nobody | claim the name with *P*; `created` |
| unclaimed | *P*, held by another subdomain | `ErrPortReserved` |
| claimed by **this** token, holding no port | `0` | accept, unchanged |
| claimed by **this** token, holding no port | *P*, held by nobody | adopt *P* into the reservation |
| claimed by **this** token, holding no port | *P*, held by another subdomain | `ErrPortReserved` |
| claimed by **this** token, holding *P* | `0` | accept; *P* stays claimed |
| claimed by **this** token, holding *P* | *P* | accept |
| claimed by **this** token, holding *P* | *Q* ≠ *P* | `ErrPortReserved` |
| claimed by **another** token | anything | `ErrNameReserved` |

Two rows carry decisions worth naming. **A reservation is one address, not a
growing set** — the last row before the refusal is what stops one token
draining the pool a reconnect at a time. And **`tcpPort == 0` asserts nothing**,
so an owner that reconnects as a plain HTTP tunnel keeps the port it claimed
rather than silently giving it up; giving it up on an argument the user
*omitted* would lose an address by accident.

**A claim records a handshake that completed.** `authorize` calls `Claim` before
it takes a port, because the refusal and the port decision are one decision; but
if anything after that fails — the registry refuses the name as live, the bind
fails, the announce fails — the relay calls `Discard` for a claim `created` by
that handshake. So a name is never consumed by a connection that never opened.
The cost of the other choice is smaller but real, and this spec takes the
conservative one: a transient bind failure on a first connection leaves the name
claimable by a stranger for as long as the owner takes to retry.

### 5.2 · `core/portpool.go` — a held port is not a free port

The pool today has two states for a number: operator-`reserved` (never
allocatable, fixed at construction) and `inUse` (allocated to a live owner). A
tenant's port between reconnects is in neither, so `Allocate` walks onto it and
hands it to the next anonymous agent. That is window (a) for the port.

A third state, `held` — *claimed by a tenant, not currently allocated*:

- `Hold(port int, holder string) error` / `Unhold(port int)` / `Holder(port int) string`.
- `Allocate(owner)` **skips** a port held by anyone other than `owner`. A
  generic caller asked for no particular number, so skipping costs it nothing.
- `AllocateSpecific(port, owner)` succeeds when the port is unheld **or** held
  by `owner`, and returns `ErrPortOutOfRange`-class refusal otherwise.

**This makes the pool's `owner` string load-bearing, so it changes from the
agent id to the subdomain.** An agent id is minted fresh on every connection
(`newAgentID`), so a hold keyed by one could never match on reconnect — which is
the entire point of a hold. `relay/tcp.go`'s three call sites and
`releaseTCP`'s `ReleaseOwner` move together; `Owner(port)` is read only by
`core/portpool_test.go`, which passes opaque strings, and by nothing on any
shipped surface.

### 5.3 · `relay` — `authorize` consults the reservation, and `Reserved` is read

`Relay.authorize` is the **single** point where a name is decided, and both
ingress paths already run through it: `ServeAgent` (`relay/relay.go`) and
`ServeSSH` (`relay/sshingress.go:162`). One change covers both, which is the
shape [L-009](https://github.com/pumasi-ai/pumasi/blob/main/lessons/L-009-two-paths-one-claim.md)
asks for rather than two that agree today.

- `relay.Config` gains `Reservations *core.Reservations`. **Nil means a fresh
  empty set, not "disabled"** — mirroring `Auth: nil → AllowAll{}`. The rule
  holds on every relay; an empty set simply has nothing claimed in it yet.
- In `authorize`, after `ValidateSubdomain` and before any port is taken:
  a request naming a subdomain **with** a token of at least `MinTokenLen` calls
  `Claim`; one **without** calls `Check`. A refusal is returned as the
  handshake's error, which both paths already render (the CLI prints it; the ssh
  path writes it to the terminal).
- `tunnel.Reserved` is set from `Reservations.Get(name)` — *is this name
  claimed* — replacing the current guess at the shape of the request, and it
  becomes the routing table's answer to "whose is this". `relay/dashboard.go`'s
  status view reports it as a `"reserved"` field, reading it from the `Tunnel`
  the registry returns and from no second source. **That the field is read *at
  all* is a review check and not an acceptance case** — no Go test can assert a
  grep, and once both possible sources agree in slice 1 no test can tell them
  apart either. `acceptance/CASES.md` C-4 asserts the part that is assertable
  and §9 records why it was narrowed.
- `tunnel.Requested` keeps its present meaning — *the agent asked for this
  address rather than accepting one* — and is unchanged. It is not a synonym for
  `Reserved` and the two now differ observably: an agent that asks for a free
  name without a token is `Requested` and not `Reserved`.

## 6 · The decision table

The rule, exhaustively, for a request that has passed `Authenticator.Authorize`
and `ValidateSubdomain`. **Compare the first column against the third: no cell
in which today's relay hands out a name becomes a refusal, unless someone else
has claimed that name.**

| the request | name unclaimed | name claimed by *this* token | name claimed by *another* token |
|---|---|---|---|
| `--subdomain N --token T` | **claim it**, port too if asked; `Reserved` | **accept**; the port comes back with it; `Reserved` | **refuse** — `ErrNameReserved` |
| `--subdomain N`, no token | **accept**, exactly as today; not `Reserved` | **refuse** — `ErrNameReserved` | **refuse** — `ErrNameReserved` |
| no subdomain (generated) | **accept**; a generated name is never claimed | n/a | n/a |
| `--subdomain N --token T`, `len(T) < 16` | **refuse** — `ErrTokenTooShort`, never degraded to anonymous | **refuse** | **refuse** |
| ssh ingress, any username | as row 2 — **an ssh client cannot send a token** (§8) | — | **refuse** |

Trust on first use is chosen over operator provisioning **for slice 1 only**,
and the reason is the product's own promise: a working tunnel from one command
with no account. An out-of-band provisioning step would be an account by another
name. What it costs is in §8, first bullet, and slice 2 gives an operator the
seeded file that answers it.

## 7 · Slices

- **Slice 1 — ownership. Built by this packet.** §5 in full. Window (a) closes:
  a claimed name and its port are held for their owner across a disconnect, and
  refused to everyone else. In memory only. **The middle column of §4 is
  untouched and the relay-restart half of item 2 is *not* delivered.**
- **Slice 2 — durability. Specified here, not built.** A `Store` behind the same
  `Reservations` type: one JSON document, written by write-to-temp-and-rename so
  a crash mid-write leaves the previous set rather than half of one; loaded at
  boot; `-reservations <path>` on `cmd/pumasi-relay`, empty meaning in-memory as
  today. Adds `LastSeen`, a 30-day idle expiry swept at load, and a cap on the
  set's size — the three things a claim-on-first-use namespace needs before it
  is written to disk, and none of which slice 1 needs while the set dies with
  the process. Window (b) closes for the name and the port. **The acceptance
  case is a second `relay.Relay` built over the same path, never a reconnect.**
- **Slice 3 — the ssh path and the operator's hand.** A grammar by which an ssh
  client can present a token (the username's `+` grammar cannot carry one today:
  a token is a valid subdomain to `parseSSHUser`, so `name+token` would be read
  as a name), and an operator-seeded reservation for a name the operator will not
  let anyone claim first. **Not specified here beyond this sentence** — it needs
  its own intent statement, because both halves change what the zero-install
  path promises.

Slice 2 is not folded into this packet on purpose. Its open questions — fsync
policy, what a corrupt file does at boot, whether two relays may share a path —
each deserve a reviewer's objection *before* code, and the record this spec is
written under is a backlog entry that named its own fix, was ranked on the
strength of it, and was wrong in a way five hundred clean suite runs could not
catch (`roadmap/BACKLOG.md`, *Delivered*).

## 8 · Failure modes this spec does not cover, chosen rather than missed

- **A stranger can claim a name before its rightful user does.** Trust on first
  use means whoever arrives first owns it, so an attacker who claims
  `sshsteward` with their own token before the keepalive is given one has it.
  **This is a strict improvement and not a solution:** today anyone may take that
  name *at any moment, repeatedly, and forever*; after slice 1 they may take it
  *once, before it is first claimed*. The operator's answer is slice 2's seeded
  file. There is no answer available without either an account (**Q-002**) or an
  operator step, and this spec takes neither.
- **A token on the wire is readable.** §3. Bounded by item 1(ii), a wildcard
  certificate in front of the relay, which is not this repository's to install.
- **A lost token is a lost name.** There is no recovery path, because a recovery
  path needs an identity to recover *to* and there is none. The name is
  unreachable until slice 2's idle expiry retires it, and in slice 1 until the
  relay restarts.
- **An ssh client cannot hold a reservation.** `ServeSSH` builds its
  `core.AuthRequest` with no `Token` field (`relay/sshingress.go:162`) and the
  username grammar cannot carry one (§7, slice 3). So the zero-install path can
  be *refused* a claimed name but can never *claim* or *reclaim* one. Every ssh
  tunnel is therefore unreservable. That is acceptable because an ssh tunnel is
  session-scoped by nature — a person is sitting at the terminal holding it open
  — but it is a narrowing of the path this product leads with, and it is stated
  here so the steward reads it rather than infers it.
- **Nothing bounds how many names one token may claim.** One token can claim
  many names, each with its own port. A per-token cap is a policy this relay has
  nowhere to put yet; `PortPool.Capacity` still bounds the ports.
- **Concurrent relays sharing one reservation set.** Not addressed in either
  slice. There is one relay.

## 9 · Amended 2026-09-01, before the freeze, on a cited spec review

`reviews/20260831-231109-spec-qwen.md` returned `VERDICT: OBJECT` citing
`acceptance/CASES.md` **C-4**. The objection was right and is recorded here
rather than argued past.

**The objection.** C-4's clause was *"`Tunnel.Reserved` is read by something"*,
and its Go test was described only as asserting the status surface's output. A
status surface could report reservation state by consulting `Reservations.Get`
directly while `Tunnel.Reserved` stayed written and never read — so the case
could pass with its own clause false. That is
[L-006](https://github.com/pumasi-ai/pumasi/blob/main/lessons/L-006-tests-that-cannot-fail.md)
inside a spec that cites L-006.

**Why it cannot be repaired by strengthening the assertion.** The natural fix is
to make the case distinguish the two sources by finding a state in which they
disagree. There is none. Under §5's rules the reservation set and a live
tunnel's `Reserved` field agree in every reachable slice-1 state — which is the
implementation being correct, and is exactly why no test can tell the sources
apart. **A property no execution can falsify is not an acceptance case.** The
grep is a *review* check; §5.3 now says so in those words, and C-4 was narrowed
to the clause a test can fail on: the routing table carries who owns a name, the
status surface reports that, and both distinguish a claimed tunnel from an
anonymous one that merely asked for its name — which `1853218`'s
`req.Subdomain != "" && req.Token != ""` cannot do.

**Four non-blocking gaps the same review named are also closed**, because each
was a real hole rather than a wording preference:

1. **A short but present token, through the relay.** §5.1 previously had `Check`
   treat it as no token. It now refuses it — `ErrTokenTooShort` — in `Claim`,
   in `Check`, and at the relay, with the reason stated. §6 gains the row and
   `acceptance/CASES.md` gains **C-6**.
2. **Two subdomains claiming one port.** §5.1's table now has the
   `ErrPortReserved` rows, and **R-7** exercises it.
3. **Rollback when a later step fails.** `Claim` now reports `created` and the
   relay calls `Discard` when the handshake it belongs to does not complete, so
   a name is never consumed by a connection that never opened. The trade-off it
   costs is stated in §5.1 rather than hidden.
4. **An owner reconnecting without naming its port.** §5.1's table now says
   `tcpPort == 0` asserts nothing and leaves an existing port claimed, so an
   address is not lost to an omitted argument.

A fifth remark — that §2 said a reservation dies *"never, until released or
expired"* while slice 1 has neither — was a real inconsistency and §2's table
now names what ends a reservation in each slice.

## 10 · Evidence that each case fails for the right reason

A green suite proves nothing about a defect the suite was already green
against (L-006). So every case was run against the defect it names, by
reinstating that defect in the built tree and running only that case. The
mutation, the cases run, and what happened:

| | Defect reinstated | Cases run | Result |
|---|---|---|---|
| **M1** | `authorize` does not consult the reservation set — the state at `1853218`, which had no such set | C-1, C-5, C-3 | **C-1 RED, C-5 RED, C-3 GREEN** |
| **M2** | `Allocate` and `AllocateSpecific` do not know about holds — window (a) for the port | P-1, P-2, C-2 | **all three RED** |
| **M3** | the status view has no `reserved` field — the state at `1853218` | C-4 | **RED** |
| **M3b** | `Tunnel.Reserved` computed from the shape of the request, status field kept | C-4 | **GREEN — see below** |
| **M4** | a short token is downgraded to anonymous instead of refused | C-6 | **RED** |

**C-1's red run is the live defect reproduced in a test.** Its first failure
line under M1 is `an anonymous agent asking for myapi got <nil>, want
ErrNameReserved` — the anonymous agent *was handed the owner's name* while the
owner was away. That is `skk6g7tyrs` beside `sshsteward`, in a fixture.

**C-3 stays green under M1, and that is the point of it.** The case that
guards the anonymous path must not depend on the fix; if it went red here it
would be asserting the withdrawal §6 forbids rather than the absence of one.

**M3b is a finding, and it corrects this spec's own first draft of C-4.** The
old expression `req.Subdomain != "" && req.Token != ""` **agrees with the
reservation record on every tunnel that successfully registers**: shape-true
with record-false needs a claim that failed, which is a refusal and not a
registration; shape-false with record-true needs an anonymous agent on a
claimed name, which `Check` refuses. So the old field was **inert, not
wrong** — its defect was that nothing enforced it and nothing read it, exactly
as `roadmap/BACKLOG.md` item 2 says and no more. C-4's *Fails when* column
originally claimed the field "cannot tell a claimed name from a name someone
merely asked for". M3b shows that claim is false, and the row now says what
the case actually discriminates rather than what it was hoped to.

This also sharpens §2: the sentence to keep is not *"the field is wrong"* but
*"a relay that only answers **is this name in use** cannot keep a promise about
restarts."* The field was a correct answer to a question nobody asked.

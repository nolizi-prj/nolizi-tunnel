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

**When the relay enters and leaves the held state**, which the first draft of
this section left unstated and no case pinned. A hold is taken in `authorize`,
from the reservation, **before the pool is asked for anything** — otherwise a
generic request could be walking onto the number while this handshake is
deciding it is spoken for. It is taken on *every* handshake for a reservation
that carries a port, not only the first, so it is idempotent by construction.
**Nothing takes it on the disconnect side, and nothing needs to:** the hold was
never released, because `ReleaseOwner` clears `inUse` only. The single place a
hold is given up is `discardClaim`, which `Unhold`s alongside the reservation it
is undoing — otherwise a discarded claim would strand a number nothing owns for
the life of the process. C-9 is that case.

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

**Answered 2026-09-01, before slice 2's code — see §14.** All three questions
are decided there and the rest of slice 2 is specified against the answers.

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
against (L-006). So every case is run against a mutation that makes its own
clause false, by reinstating that behaviour in the built tree and running only
that case.

**This section's first version said "every case" and covered only ten of
sixteen** — no row for R-1 to R-8. `reviews/20260831-234800-spec-glm.md`
objected that the sentence overclaimed exactly the discipline it was written to
enforce, **and that the gap is how a second L-006 defect reached the freeze**
(§12). The table below is the complete one.

| | Defect reinstated | Cases run | Result |
|---|---|---|---|
| **M1** | `authorize` does not consult the reservation set — the state at `1853218`, which had no such set | C-1, C-5, C-3 | **C-1 RED, C-5 RED, C-3 GREEN** |
| **M2** | `Allocate` and `AllocateSpecific` do not know about holds — window (a) for the port | P-1, P-2, C-2 | **all three RED** |
| **M3** | the status view has no `reserved` field — the state at `1853218` | C-4 | **RED** |
| **M3b** | `Tunnel.Reserved` computed from the shape of the request, status field kept | C-4 | **GREEN — see below** |
| **M4** | a short token is downgraded to anonymous instead of refused | C-6 | **RED** |
| **M5** | `relay.discardClaim` does nothing — the relay never undoes a claim | C-7, and every other relay case | **C-7 RED, C-1..C-6 all GREEN** |
| **MC1** | `sameToken` always true | R-1 | **RED** |
| **MC2** | `checkToken` never refuses a short token | R-4 | **RED** |
| **MC3** | `Claim` stores the token in the clear | R-5 | **RED** |
| **MC4** | `Claim` ignores port conflicts | R-6, R-7 | **both RED** |
| **MC5** | `Discard` does nothing | R-8 | **RED** |
| **MC6** | `Check` never consults the reservation set | R-2, R-3 | **R-2 RED, R-3 GREEN** |
| **M6** | the §11 guard in its **over-broad** form — suppress the rollback whenever anything is registered, not only the claim's own tunnel | every relay case | **C-8 RED, and C-8 alone** |
| **M7** | `discardClaim` drops the claim and leaks the pool hold | C-9, C-2, C-7 | **C-9 RED, and C-9 alone** |

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

## 11 · A hazard found after the code review, and a test that had to be thrown away

Found by re-reading `authorize`'s concurrency after `gemini` approved the
diff — so it is recorded here rather than in a review transcript, and neither
reviewer is being blamed for it.

**The hazard.** Two connections present the **same token for the same name** at
the same time. The first to reach `Claim` creates the reservation and carries
`newClaim`; the second finds it already its own and carries `""`. Both then go
on to `Register`, and whichever loses gets `ErrNameTaken` and calls
`discardClaim`. If the loser is the one that *created* the claim, an unguarded
discard **destroys the reservation while the winner is registered on it** — the
winner's tunnel reports `Reserved` with nothing behind it, and its name is free
the instant it drops. That is the defect this whole spec closes, reintroduced
through the rollback path §5.1 asks for.

**The fix** is one guard: `discardClaim` returns early if the registry still has
the name. It is stated in §5.1's terms — a claim is undone only when nothing is
using it.

**The test written first was thrown away, and that is the part worth keeping.**
The obvious case fires two concurrent handshakes with one token and asserts the
reservation survives. It was built, and then run **with the guard removed**:
**0 failures in 5 runs at 40 attempts each, and 0 failures in a single run of
1,500 attempts.** The interleaving needs the winner of `Claim` to be preempted
before `Register` while the other connection completes both, and two separate
TCP dials through one accept loop do not deliver that on this host. **A case
that cannot fail against the defect it names is not evidence of anything**
(L-006) — it is precisely what `reviews/20260831-231109-spec-qwen.md` objected
to in C-4, and shipping it after that objection would have been the same
mistake with the ink still wet.

What replaced it is an **in-package unit test** of the guard's own clause:
register a tunnel on a claimed name, call `discardClaim`, assert the
reservation is still there — then unregister and assert the same call now
discards, so the guard is a guard and not a disablement. Measured both ways:
**green with the guard, red without it, deterministically, in 0.00s.**

**It is not a frozen acceptance case and is not numbered as one.** The frozen
cases are this spec's contract with a reviewer; this is an implementation
hazard in one method's rollback path, and inventing a C-7 after the freeze
would blur which is which. It lives in `relay/discardclaim_test.go` and points
back here.

**M5 is the objection in §12, measured.** With `discardClaim`'s body emptied,
C-7 goes red — *"an anonymous agent was refused a name whose only claimant never
opened a tunnel"* — and **C-1 through C-6 all stay green, as do all eight
R-cases**. Before C-7 existed, that mutation passed the entire frozen suite.

**MC6 and M1 are the same shape and both matter.** R-3 stays green when `Check`
stops consulting the set, exactly as C-3 stays green under M1: the two cases
that guard the anonymous path must not depend on the fix, or they would be
asserting the withdrawal §6 forbids rather than the absence of one.

## 12 · Amended again, on a second cited objection — R-8 could pass with its clause false

`reviews/20260831-234800-spec-glm.md` returned `VERDICT: OBJECT` citing
`acceptance/CASES.md` **R-8**, and it was right.

**The objection.** R-8's clause was *"a claim recorded by a handshake that then
fails is discarded"* and its *Fails when* named *"a bind failure on a first
connection"*. Both are **relay**-level: handshakes and binds live in `relay`,
and §5.1's sentence is *"the relay calls `Discard`"*. But R-8's test is
`core.TestADiscardedClaimLeavesNothing`, which **supplies the `Discard` call
itself** and can only exercise the primitive's semantics. So a relay with
`discardClaim` deleted from every call site passed R-8 green while R-8's clause
was false in exactly the way its own column described — and no other case
noticed, because C-1, C-2 and C-4 end in successful handshakes, C-3's claim
succeeds, C-5 is a `Check` refusal and C-6 is refused before anything is
created.

**This is the same pattern as the C-4 objection in §9, in a spec that had just
recorded that objection.** It is not a coincidence that it survived: §10's
opening sentence claimed every case had been run against its own defect while
its table covered ten of sixteen, so the one class of case that had never been
mutated is the one that carried the defect. The evidence gap and the test gap
are one finding seen from two sides, which is glm's phrasing and is worth
keeping.

**What changed.** **R-8 is narrowed** to `Discard`'s own semantics — remove the
name, take its port hold with it — which is what its test actually asserts.
**C-7 is new** and carries the relay clause: the public port is bound by
something else before the agent arrives, so the relay claims the name,
allocates the port and then cannot bind it; the agent is refused; and an
*anonymous* agent afterwards is **given** that name, because nobody owns it.
It is falsifiable and was falsified — §10, M5. And **§10's table is now
complete**, with a mutation for every built case. *(That sentence originally
said "sixteen"; there were seventeen, and §13 adds two more. The table was
complete; the count beside it was not — noted by the same review.)*

**What was not changed, and why.** The frozen cases C-1 to C-6, R-1 to R-7 and
P-1 to P-2 are untouched; their assertions were not in question. No frozen case
from `spec/0001`-`0003` is touched either, so **Q-030** still gets no third
instance from this work.

## 13 · Amended a third time, on a second cited objection from the same family

`reviews/20260831-235635-spec-glm.md` returned a second `VERDICT: OBJECT`, on
the **§11 guard added in answer to the first one**. It was right, and the shape
of the mistake is worth more than the fix.

### Objection 1 — the guard was correct about the case it was written for and wrong about the common one

§11 suppressed the rollback whenever *anything* was registered on the name. The
hazard it was written for is a **co-claimant** — two connections with one token.
But the ordinary state of this relay is an **anonymous agent sitting on an
unclaimed name**, and when a token-holder arrives there: `Claim` succeeds,
`Register` refuses `ErrNameTaken`, and the guard suppressed the discard because
the registry had the name. **The claim then persisted on a name whose claimant
never opened a tunnel** — exactly what §5.1 declares impossible, contradicted by
the clause added to defend §5.1. And no case could fail on it: C-7 exercises
only the registry-empty path, and §11's in-package test asserts the guarded
branch as *correct*. Its two poles sat either side of the state that was wrong.

**The fix is a narrower discriminator**, and it is the one the hazard analysis
should have named in the first place: suppress only when the **live tunnel is
the claim's own** — `Reserved` true. Only a co-claimant can be registered
`Reserved` on a name *this* handshake has just claimed, because before that
claim there was nothing to be reserved by. A squatter is `Reserved` false, and
the discard proceeds. **C-8** is the case, and under the old guard it is the
**only** case in the suite that goes red (§10, M6).

**The corollary the objection also caught.** §9 narrowed C-4 on the premise that
the reservation set and a live tunnel's `Reserved` never disagree in a reachable
slice-1 state, and §11's guard falsified it: the squatter's tunnel was
registered before the claim existed. With the narrowed guard the disagreement is
**transient and confined to a failing handshake** — it exists between the
token-holder's `Claim` and its `Discard`, and no steady state carries it. That
is weaker than §9's original "in every reachable state" and this spec says the
weaker thing rather than restating the stronger one.

### Objection 2 — a held state with no clause saying when it is entered

§5.2 introduced `Hold`/`Unhold`/`Holder` and assigned them no caller. The
implementation had one, but the spec did not, and **C-2 cannot tell the
difference** between "released to held" and other designs. §5.2 now states the
transition, and **C-9** pins the half that no case observed at all: a discarded
claim gives its port back, rather than stranding a number nothing owns for the
life of the process. Under a leaked hold, C-9 is the only case that goes red
(§10, M7).

### The pattern, which is the third instance of one thing

C-4, R-8 and now §11's guard were each **a clause whose truth no execution
could reach**. The first two were caught by reviewers; this one was *introduced
by the fix for the second*. What survives it: a guard added to protect an
invariant needs a case that fails when the guard is wrong, not only one that
passes when it is right — and §11's in-package test, which asserts the guarded
branch is correct, is exactly the test that cannot notice a guard firing too
often.

### Minor items from the same review, closed without objection

- §12 said "sixteen built cases" where there were seventeen. Corrected there.
- `CASES.md`'s preamble said "every case sequences a disconnect and a second
  connection" — false of the ten `core` cases. Scoped to the C-cases.
- §5.1's `Check` refusing short tokens is dead-path symmetry, since §5.3 never
  calls `Check` with a token. Kept deliberately: a primitive that is safe only
  because of its caller is one refactor from being unsafe.

## 14 · Amended 2026-09-01, before slice 2's code — the three open questions, answered

§7 closes by naming three questions and saying each *"deserves a reviewer's
objection **before** code"*: **fsync policy**, **what a corrupt file does at
boot**, and **whether two relays may share a path**. This section answers all
three and specifies the rest of slice 2 against those answers. It is written
before any of slice 2's implementation exists, which is the whole of why it is
a separate section rather than an edit to §7 — the record shows what was
decided and when, not a spec that looks like it always knew.

**No existing row of [`acceptance/CASES.md`](acceptance/CASES.md) is touched by
this amendment, including D-1**, which stands exactly as frozen. §14.9 adds new
rows in a new section, by the same route §13 added C-8 and C-9.

### 14.1 · Question 1 — fsync, and what the rename actually buys

**The rename buys atomicity. It does not buy durability, and this spec stops
saying "crash" as though the two crashes were one crash.**

Two different events are in scope and they have different answers:

| event | what rename alone gives | what fsync adds |
|---|---|---|
| **the relay process dies mid-write** (panic, SIGKILL, deploy) | **complete protection.** The kernel's page cache is unaffected by a process dying; the next `open` reads the previous document or the whole new one and never half of one. | nothing |
| **the host loses power or the kernel panics** | **nothing guaranteed.** The temp file's data and the directory entry are both potentially still in cache. `ext4`'s `auto_da_alloc` heuristic covers the common truncate-then-rename shape and is a heuristic, not a contract, and other filesystems do not have it. | the guarantee |

**Decision: fsync both, and the order is part of the specification.**

1. write the whole document to `<path>.tmp` in the **same directory** as
   `<path>` — never `os.TempDir()`, because `rename(2)` across filesystems
   fails and is not atomic where it does not;
2. `f.Sync()` on the temp file, **before** the rename, so the rename cannot
   become durable ahead of the bytes it points at;
3. `os.Rename(tmp, path)`;
4. **open the containing directory and `Sync()` it**, so the directory entry
   the rename created is itself durable.

Mode `0600` on the temp file. The document holds token digests; a digest is not
a live credential (§3.1) and is still the input to an offline attack, and
`0600` costs one constant.

**Why fsync is worth its cost here, stated as a cost rather than assumed away.**
Two `fsync`s happen on the path that creates or refreshes a claim — once per
handshake that presents a token, and never on the anonymous path. That is a
human-rate event on a relay whose whole live set is two tunnels. Against it:
**§4's third column says a host restart behaves as a relay restart, and that
claim is only true with step 4.** A spec that promised the middle column and
skipped the fsyncs would be making the right-hand column's promise on a
heuristic.

**What is still not guaranteed, and this is the honest boundary.** `fsync`
returns when the filesystem says the data is durable. A disk with a lying write
cache, a virtualised block device that acknowledges early, or an `fsync` that
fails and is not retried leave the guarantee where POSIX leaves it. This spec
claims what POSIX gives and no more, and it does not claim the reservation set
survives a disk that lies.

### 14.2 · Question 2 — a corrupt file at boot: **start empty, log ERROR, and move the file aside**

The two candidates are opposite products and both are defensible, which is why
§7 sent them to a reviewer. The reasons for the one taken:

**Refusing to start is the wrong failure, and its cost is not the reservation
set.** A relay that will not come up over a damaged reservations file has
turned a namespace-bookkeeping problem into a **total outage of every path**,
including the anonymous path — a tunnel from one command with no account —
which has no dependence on this file at all. On this deployment it also means
`sshsteward` does not come back, and that tunnel is `RESOURCES.md` §4's route to
the host. **A restart that can fail to complete is precisely the thing Q-014 is
open about**, and slice 2 must not add a new way for one to.

**Starting empty is not silent, and it degrades to a shipped behaviour rather
than to an unspecified one.** What a relay with an empty reservation set does is
exactly what every release before this one did: every name is unclaimed and
trust-on-first-use. That is a known product, not a hole.

**The threat model does not favour refusing, either** (L-004). The party who can
corrupt this file is the party with write access to the relay's state directory,
and that party can also delete it, rewrite it to grant themselves every name,
read every digest in it, or stop the process. *Corrupt the file to clear the
namespace* is not a new capability; it is a strictly weaker use of one that
already defeats the scheme.

**So the rule, and the three parts are not optional:**

1. **Start with an empty set** and serve.
2. **Log at ERROR**, once, naming the path and the parse error. Not `Warn`: a
   set of ownership records has been lost and the operator is the only party who
   can tell whether that matters.
3. **Rename the damaged file to `<path>.corrupt-<RFC3339 UTC>`** before the
   first write replaces it. Without this, the first claim after the restart
   overwrites the only evidence of what went wrong. If that rename fails, log
   and continue — evidence preservation must not be the thing that stops the
   relay either.

**Absent is not corrupt.** `os.IsNotExist` on the first read is the ordinary
first boot: empty set, **no error, no ERROR log, no file moved aside**. A store
that shouted on first boot would train an operator to ignore it.

**Unreadable is corrupt for this purpose.** A permissions or I/O error reading
an existing file takes the same path as a parse error. It is a different cause
with the same available responses, and one rule with one behaviour is worth more
than two rules that a reader has to hold apart. **There is no flag to choose
the other product**, because the operator who would set it is the operator who
is not there.

### 14.3 · Question 3 — two relays over one path: **refused, with a lock, not with a sentence**

§8 already records concurrent relays as *"not addressed in either slice. There
is one relay."* **Slice 2 makes leaving it unaddressed worse than it was**, and
that is the finding this question deserved: each relay holds a full in-memory
copy and each write replaces the **whole document**, so two relays over one path
do not interleave or corrupt — they take turns silently destroying each other's
claims, and the loser learns at the next restart. That is a new failure mode
created by this slice, and it produces exactly the loss the slice exists to
remove.

**Decision: the store takes an exclusive advisory lock and a second relay over
the same path refuses to start**, with an error naming the path.

- **`flock(2)`, `LOCK_EX|LOCK_NB`, on a sibling `<path>.lock`** — *not* on
  `<path>` itself, because `<path>` is replaced by rename on every write and a
  lock on the old inode guards nothing.
- Held for the life of the store and released by `Close`. **A crash needs no
  cleanup**: the kernel drops `flock` locks when the file description closes,
  so there is no stale-lock path and no lock file to delete by hand.
- **`Close` is why `relay.Relay` gains a `Close` method.** D-1 shuts the first
  relay down before constructing the second over the same path, and `flock`
  conflicts between two file descriptions *in the same process*, so D-1 is also
  what proves the lock is released rather than leaked.
- **Honest limit:** `flock` is advisory and its behaviour over NFS and some
  network filesystems is not the local one. A relay whose state directory is on
  such a filesystem gets the documented refusal on a best-effort basis. This is
  stated rather than defended: the answer to *may two relays share a path* is
  **no**, and the lock is how the common case is enforced, not a proof.

### 14.4 · The document

One JSON object, written whole:

```json
{
  "version": 1,
  "reservations": [
    {"subdomain": "myapi", "token_hash": "<hex sha256>", "tcp_port": 20000,
     "last_seen": "2026-09-01T06:24:00Z"}
  ]
}
```

`version` exists so that a future format change is **detectable rather than
guessed at**. A document whose `version` is not `1` is not parsed further and
takes §14.2's corrupt path — start empty, log, move aside. Refusing to start on
a version from the future would be the same total outage §14.2 declines, and
guessing at it would be worse than either.

The set is written **sorted by subdomain**, so two identical sets produce
identical bytes and a diff of two store files is a diff of their contents rather
than of Go's map iteration order.

### 14.5 · `LastSeen`, the 30-day sweep, and the cap

**`Reservation.LastSeen time.Time`** — the field §5.1 deliberately did not ship
ahead of its reader. Slice 2 is the reader.

- **Set by `Claim`, and by nothing else.** Both branches: a new claim and an
  owner returning to one it already holds. `Check` still records nothing and
  still says so — a stranger being refused a name must not refresh its owner's
  clock, and an owner who reconnects presents a token and therefore goes through
  `Claim` (§5.3: a token at or over `MinTokenLen` takes the claim path).
- **`ReservationTTL = 30 * 24 * time.Hour`,** swept **at load only**. No timer,
  no goroutine, no sweep on write. A relay that runs for sixty days holds a
  fifty-day-idle name until it restarts, and that is the trade taken on purpose:
  the alternative is a background writer racing the handshake path for the same
  document, for an expiry whose whole purpose is to bound *long-term* namespace
  growth.
- **A record whose `last_seen` is the zero time is treated as seen at load
  time**, not as the epoch. It is kept and given a fresh thirty days. Reading
  zero as the epoch would drop every such record instantly, and losing a whole
  namespace to a missing field is a much worse failure than holding one name
  thirty days too long. (No such document can exist today — this is format
  version 1 and there is no earlier one — so this rule is about the next format,
  and it is written now because the cheap moment to write it is now.)
- **`MaxReservations = 10000`,** a cap on the **set**, not per token. At the cap,
  `Claim` refuses a **new** name with `ErrReservationsFull`. **An owner returning
  to a name already in the set is never refused by the cap** — a cap that could
  lock out an existing owner would destroy the property the whole slice exists
  to establish. The number is chosen so that the full-document rewrite stays
  trivial (10,000 records at roughly 150 bytes is about 1.5 MB) while sitting far
  above any plausible legitimate use of one relay.

  **What the cap is and is not.** It bounds unbounded disk growth by an
  unbounded claimer. It is **not** a defence against namespace exhaustion: a
  party who can complete 10,000 handshakes can fill the set and refuse everyone
  else a new name until the sweep or a restart. Before the cap that party filled
  the disk instead. Both are bad; the cap picks the bounded one and this spec
  does not dress it up as protection. §8's *"nothing bounds how many names one
  token may claim"* is unchanged and stays true.

### 14.6 · The write path, and what a failed write does

**Every mutation writes the whole document, synchronously, on the calling
goroutine, under the same lock that guards the maps.** No queue, no batching, no
background flusher: a claim that has returned to the handshake has been written,
which is the only way `Claim`'s return value can mean what §14.1 says it means.

The two mutators answer a failed write differently, and the asymmetry is
deliberate:

- **`Claim` rolls back and returns the error.** The in-memory set is restored to
  exactly what it was — the record deleted if this call created it, the previous
  value restored if it did not, and any port hold undone with it — and the
  handshake is refused with the write error as its reason. **A claim that cannot
  be persisted is not a claim**, because the only thing a claim promises over a
  plain registration is that it survives the process. Refusing is honest, and it
  costs nothing to the anonymous path, which never reaches this code.
- **`Discard` removes the record from memory regardless, and reports the write
  error to its caller** (its signature gains an `error`; a call used as a
  statement is unaffected, so no frozen case moves). `Discard` exists to keep
  §5.1's invariant — *a name is not consumed by a connection that never opened* —
  and making it fail would leave in memory the very claim it was called to
  remove. `relay.discardClaim` logs the error at ERROR and continues.

**The residual this asymmetry leaves, named rather than smoothed over:** a
handshake whose `Claim` write **succeeded** and whose bind then failed, and
whose `Discard` write then failed, leaves a record on disk that memory does not
have. It returns at the next restart, and the name is then held for up to
`ReservationTTL` by a token whose handshake never opened a tunnel. It requires
two writes to the same file to disagree within one handshake. It is a real hole,
it is bounded by the sweep, and it is written here rather than discovered later.

### 14.7 · The change · slice 2

- **`core/reservationstore.go` — new.** `store` — open/lock/load/sweep/write for
  one path. `OpenReservations(path string) (*Reservations, error)` returns a set
  backed by it; `(*Reservations).Close() error` releases the lock.
  `NewReservations()` is unchanged and still returns an unbacked set, so every
  existing caller and every frozen case keeps exactly today's behaviour.
- **`core/reservation.go`** — `LastSeen` on `Reservation`; `Claim` stamps it and
  persists; `Discard` persists and returns `error`; `ErrReservationsFull` and
  the cap; `ErrStoreLocked` for §14.3.
- **`relay/relay.go`** — `Config.ReservationsPath string`. Empty means an
  in-memory set, **exactly as today**. **Setting both `Reservations` and
  `ReservationsPath` is refused by `New`**, at startup, for the same reason
  `PublicScheme` is validated there: two sources of truth for one set is an
  operator's typo that should be a refusal to start rather than a silent choice
  between them. `Relay.Close() error` closes what `New` opened and nothing else
  — a relay handed a `Reservations` it did not open does not close it.
- **`cmd/pumasi-relay`** — `-reservations <path>`, default `""`. A twelfth flag.
  Empty is today's relay in every respect: no file, no lock, no sweep, no write.

### 14.8 · What slice 2 still does not do

- **It does not shorten an outage, and no surface may say it does.** A live TCP
  connection cannot outlive the process at either end (§4). What already bounds
  the outage is the agent's 1 s → 30 s backoff and the operator's keepalive.
  What slice 2 removes is the **loss**: after a restart an in-memory set is
  empty, so every claimed name is trust-on-first-use again for the length of the
  reclaim window, and that window is where the unrecoverable loss lives.
- **It does not give an ssh client a reservation** (§8, slice 3).
- **It does not seed a reservation for the operator** (§7, slice 3). Trust on
  first use is still the only way a name is claimed, so §8's first bullet is
  unchanged: a stranger may still claim a name before its rightful user does,
  **once**.
- **It does not survive a disk that lies about `fsync`** (§14.1).

### 14.9 · Acceptance cases added by this amendment

Added to [`acceptance/CASES.md`](acceptance/CASES.md) in a new section. **D-1 is
not edited and no other existing row is edited.**

| # | Case | Fails when |
|---|---|---|
| **D-2** | A crash mid-write leaves the **previous** set, not half of one. | The document is written in place, so an interrupted write leaves a truncated file and the whole set is lost at the next load. |
| **D-3** | A record idle longer than `ReservationTTL` is gone after a load; one inside it is still there. | Nothing sweeps, so a claim-on-first-use namespace grows forever and a name is never recoverable from an owner who has vanished. |
| **D-4** | At `MaxReservations`, a new name is refused and an existing owner is still admitted. | The cap is applied to every `Claim`, so the set fills and locks its own owners out — the property the slice exists to establish, destroyed by the guard added to bound it. |
| **D-5** | A corrupt file starts the relay **empty** rather than not at all, and the damaged bytes are still on disk afterwards. | The store returns an error that `relay.New` propagates, and a damaged bookkeeping file becomes a total outage; or it starts empty and the first write destroys the evidence. |
| **D-6** | A second store over a path a live store holds is refused. | Two relays share a path, and each full-document write silently destroys the other's claims. |

**D-2 is demonstrated, not asserted.** The write path is driven to fail
**after** the temp file exists and **before** the rename, and the set that
loads afterwards is checked to be the previous one, record for record.

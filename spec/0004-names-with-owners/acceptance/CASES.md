# Acceptance cases · Spec 0004

Frozen at spec review, before implementation. The builder may not edit them; a
case that is wrong is fixed by amending the spec in the open and taking a fresh
cross-family spec review (CHARTER §3 requirement 2).

Every case names **what execution makes it fail**. No case here dials and hopes
to lose a race, because none of them contains one: every case sequences a
disconnect and a second connection, and the states either side of that
disconnect are reached by waiting on an observable fact rather than on a clock.
The existing suite is not an acceptance case for this spec — it is green at
`1853218` against the whole of the defect
([L-006](https://github.com/pumasi-ai/pumasi/blob/main/lessons/L-006-tests-that-cannot-fail.md)).

 > **Amended 2026-09-01, before the freeze** — see [`../SPEC.md` §9](../SPEC.md).
> The spec review objected, with a citation, that **C-4** could pass while the
> clause it named was false: a status surface can report reservation state
> without ever reading `Tunnel.Reserved`
> (`../../../reviews/20260831-231109-spec-qwen.md`). C-4 was **narrowed to the
> clause a test can fail on**, and the un-testable half was moved to `SPEC.md`
> §5.3 as a review check — no Go test can assert a grep, and under correct
> slice-1 behaviour the two possible sources never disagree, so no execution
> could distinguish them. **R-7** and **C-6** are new, from the same review.

## How a disconnect is sequenced without a race

`ServeAgent`'s cleanup — `delete(r.sessions, …)`, `releaseTCP`,
`registry.UnregisterAgent` — runs in a deferred function after the session
closes, so a second agent that dials immediately can arrive either side of it.
A case that raced that would sometimes be refused by the **registry**
(`ErrNameTaken` — the name is still live) instead of by the **reservation**
(`ErrNameReserved` — the name is owned), and would then pass at `1853218` for a
reason that is not its own.

So every case below **waits until `relay.Registry().Has(name)` is false** before
the second agent dials, and asserts the refusal it gets is `ErrNameReserved`.
At `1853218` that wait completes and the second agent is *accepted*, which is
the case going red on its own clause.

## The cases · slice 1, built by this packet

### Ownership, in `core` — `core/reservation_test.go`

| # | Case | Go test | Fails when |
|---|---|---|---|
| **R-1** | A name claimed with one token is refused to a different token. | `core.TestAClaimedNameIsRefusedToAnotherToken` | `Claim` records nothing, or compares nothing, so the second token is accepted. |
| **R-2** | A name claimed with a token is refused to a caller with no token. | `core.TestAClaimedNameIsRefusedToAnAnonymousCaller` | `Check` answers only about liveness, so an anonymous caller passes. |
| **R-3** | A name **nobody** has claimed is given to a caller with no token. | `core.TestAnUnclaimedNameIsStillAnonymous` | A change makes a token mandatory for a named request — the withdrawal `SPEC.md` §6 forbids. |
| **R-4** | A token shorter than `MinTokenLen` cannot claim a name. | `core.TestAShortTokenCannotClaim` | Short tokens are hashed and accepted, which is the assumption `SPEC.md` §3.1 rests SHA-256 on. |
| **R-5** | The record holds a digest, never the secret. | `core.TestAReservationDoesNotStoreTheToken` | The token is stored in the clear, so a leaked set is a set of live credentials. |
| **R-6** | A name claimed with port *P* refuses a claim for the same name at a different port, and a reconnect naming **no** port keeps *P*. | `core.TestAReservationIsOneAddress` | A reservation accumulates ports, so one token can drain the pool one reconnect at a time; or `tcpPort == 0` is read as "give it up" and an address is lost to an omitted argument. |
| **R-7** | A port claimed by one name cannot be claimed by a second name. | `core.TestAPortBelongsToOneName` | `Claim` checks only the name, so two reservations hold one number and whichever agent connects second is bound to a port it does not own. |
| **R-8** | `Discard` removes a claim and takes its port hold with it. | `core.TestADiscardedClaimLeavesNothing` | `Discard` leaves the name claimed, or frees the name and leaves its port held. **This is a case about `Discard`'s own semantics and no more** — that the *relay* ever calls it is **C-7**, and `SPEC.md` §12 records the objection that separated them. |

### The port pool — `core/portpool_test.go`

| # | Case | Go test | Fails when |
|---|---|---|---|
| **P-1** | A held port is not handed out by a generic `Allocate`. | `core.TestAllocateSkipsAHeldPort` | `Allocate` knows only `inUse` and the operator's `reserved`, so it walks onto a tenant's port the moment that tenant disconnects. This is window (a) for the port, and it is the state at `1853218`. |
| **P-2** | A held port is granted to its holder and refused to anyone else by `AllocateSpecific`. | `core.TestAllocateSpecificHonoursTheHolder` | Holds are advisory: either the holder cannot reclaim its own port, or a stranger naming the number gets it. |

### End to end, through the relay — `relay/reservation_test.go`

| # | Case | Go test | Fails when |
|---|---|---|---|
| **C-1** | **A name survives its owner's disconnect.** An agent claims `myapi` with a token and disconnects. While it is away — `Has("myapi")` false — an anonymous agent asking for `myapi` is refused with `ErrNameReserved`. The owner then reconnects with the same token and is given `myapi` back. | `relay.TestAClaimedNameIsHeldAcrossADisconnect` | The deferred `registry.UnregisterAgent` is the only thing that happens on disconnect, so the name is free and the anonymous agent **is given it**. That is `1853218`, and it is `skk6g7tyrs` beside `sshsteward` on the live relay. |
| **C-2** | **A public TCP port survives its owner's disconnect.** An agent claims `sshlike` with a token and `--tcp --tcp-port P` on a pool whose range is exactly `[P,P]`, then disconnects. While it is away, an anonymous agent asking for **any** TCP port is refused `ErrPortPoolExhausted` — not handed `P`. The owner reconnects and gets `P` back. | `relay.TestAClaimedPortIsHeldAcrossADisconnect` | `releaseTCP` → `ReleaseOwner` returns `P` to the free pool, so the next generic `Allocate` walks straight onto it and answers the anonymous agent with the steward's address. |
| **C-3** | **Nothing is withdrawn from the anonymous case.** On a relay with a claimed `myapi`, an agent with no token and no subdomain still gets a generated name, and an agent with no token asking for the *unclaimed* `other` still gets `other`. | `relay.TestTheAnonymousCaseStillWorks` | A change makes tokens mandatory for named requests, or makes generated names pass through the reservation set. `SPEC.md` §6's third column is the specification of this case. |
| **C-4** | **The routing table carries who owns a name, and the status surface reports it.** With `myapi` claimed by a token and `other` merely *asked for* by an anonymous agent, `Registry().Lookup` reports `Reserved` true for the first and false for the second, and the status view agrees with both. | `relay.TestTheStatusSurfaceReportsWhoOwnsAName` | **The status view has no `reserved` field at all**, which is the state at `1853218`, so nothing can report who owns a name. The registry half of this case does **not** discriminate the old expression from the new one and this row does not claim it does — see `SPEC.md` §10, M3b: `req.Subdomain != "" && req.Token != ""` agrees with the record on every tunnel that successfully registers, so the old field was **inert rather than wrong**. **Narrowed from its first version on a cited objection; `SPEC.md` §9 says why the other half is a review check and not a case.** |
| **C-6** | **A short token is refused, not quietly ignored.** An agent presenting `--subdomain myapi --token short` is refused with `ErrTokenTooShort` and is *not* given `myapi` as an anonymous caller would be. | `relay.TestAShortTokenIsRefusedNotDowngraded` | The relay degrades a short token to no token, so a user who believes they are claiming a name is handed it unprotected and loses it at the next disconnect. `SPEC.md` §3.1 rests SHA-256 on this refusal existing. |
| **C-7** | **A handshake that claims a name and then fails leaves nothing behind.** The public port is bound by something else before the agent arrives, so the relay claims `bindfail`, allocates the port and then cannot bind it. The agent is refused — and afterwards an **anonymous** agent asking for `bindfail` is given it, because nobody owns it. | `relay.TestAFailedHandshakeLeavesNoClaim` | The relay never calls `Discard`, so a name is consumed permanently by a connection that never opened: the anonymous agent is refused `ErrNameReserved` for a token that never opened a tunnel. **`discardClaim` can be deleted from every call site and every other frozen case still passes** — which is what `SPEC.md` §12's objection cited. |
| **C-5** | **Both ingress paths obey one reservation.** An ssh client whose username is a name claimed by an agent's token is refused, and told why on its terminal. | `relay.TestSSHIsRefusedAClaimedName` | The check is added to `ServeAgent` rather than to the `authorize` both paths share — one claim true of one path ([L-009](https://github.com/pumasi-ai/pumasi/blob/main/lessons/L-009-two-paths-one-claim.md)). |

## The case for slice 2, specified and **not built by this packet**

| # | Case | Go test | Fails when |
|---|---|---|---|
| **D-1** | **A reservation outlives the process.** A relay with a store at *path* takes a claim on `myapi` with port `P`. It is shut down and **a second `relay.New` is constructed over the same *path*** — a new process's worth of state, not a reconnect. An anonymous agent asking for `myapi`, or for any TCP port on a `[P,P]` pool, is refused; the owner presenting the same token is given both back. | `relay.TestAReservationOutlivesTheRelay` | The reservation set is in memory, so the second relay starts empty and answers the anonymous agent yes. That is `1853218` and it is **also the state after slice 1** — which is why this case is written here and left red rather than omitted. |

**D-1 is the middle column of `SPEC.md` §4 and it is the only case that can
speak to it.** It is written down now, at the freeze, so that slice 2 is
measured against a case that existed before its implementation did, and so that
no reader of a green slice-1 suite concludes the restart half was delivered.
**A green run of the cases above is not evidence about D-1 in either
direction.**

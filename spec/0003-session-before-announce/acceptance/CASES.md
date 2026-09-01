# Acceptance cases · Spec 0003

Frozen at spec review, before implementation. The builder may not edit them; a
case that is wrong is fixed by amending the spec in the open and taking a fresh
cross-family spec review (CHARTER §3 requirement 2).

Every case below names **what execution makes it fail**, and no case dials and
hopes to lose a race. The existing suite is not an acceptance case here: it
passes 90–95% of the time against defects of this class, and job `0059` ran
`go test -count=1 ./...` **500 times at `b3d251d` with zero failures** against
the very defect this spec fixes. A green run from it proves nothing (L-006).

Where a case cannot go red against `8540b89` — or goes red there for a reason
that is not its own — the evidence section below says so in those words rather
than presenting it as evidence it is not. Two of the three are in that position
and both are still here, for reasons stated.

> **Amended 2026-08-31, before the freeze** — see
> [`../SPEC.md` §7](../SPEC.md). The first version of C-3 asserted only what
> was true *after* a failed announce, and the spec review objected, with a
> citation, that it therefore could not fail on the clause it named
> (`../../../reviews/20260831-193215-spec-qwen.md`). C-3 was strengthened, not
> weakened: it now holds the relay inside the failing announce and asks what a
> visitor queued on `r.mu` behind it is shown.

## How the ordering is asserted without a race

`spec/0002`'s cases made the second step **fail**, so that two orderings
differed in what the client held rather than in when it held it. There is no
second step to make fail here — installing a map entry cannot be refused. These
cases use the other half of the same idea: **hold the relay still at the
instant under test.**

`gatedConn` (`relay/sessionorder_test.go`) is the relay's end of an agent
connection with a brake on its first write. It closes `announcing` as the relay
enters the write that carries the auth response, and does not return from that
write until the case closes `release`. While it is held:

- the relay has, by construction, already run `registry.Register` — the
  hostname resolves;
- the relay has, by construction, not returned from the announce — so under
  `8540b89`'s ordering it cannot possibly have installed the session.

A visitor sent in at that moment is therefore not racing anything. Under
`8540b89` it is answered `404` **immediately** — the case goes red in 0.00s,
not after a wait. Under the fix it blocks on `r.mu` and is served once the
announce completes.

The 750 ms `announceHold` in C-1 and C-2 is not a window the defect has to be
hit inside. It bounds how long a *passing* run waits to confirm that nothing
answered; the failing run does not use it. Every relay in these cases sets
`HandshakeTimeout: 60s` so the brake stays well inside the deadline the
announce is written under (`SPEC.md` §3.2).

## The cases

| # | Case | Go test | Fails when |
|---|---|---|---|
| **C-1** | **A URL that is announced is a URL that works.** With the relay held inside the announce, a visitor using the hostname is **not answered at all**; when the announce is released, that same request is served `200` by the local service. | `relay.TestVisitorIsNotAnsweredBeforeTheSessionExists` | The session is installed after the announce: the visitor passes `registry.Lookup`, reads `session == nil`, and is answered `404 No tunnel is open for <host>` while the relay is still inside `writeFrame`. |
| **C-2** | **The announce is the first thing its connection carries.** With a visitor sent in during the held announce, the agent still receives an auth response it can decode, for the subdomain it asked for. Asserts nothing about the visitor — that is C-1's — and logs its outcome so a request that is never answered at all is named rather than swallowed. | `relay.TestTheAnnounceReachesTheAgentBeforeAnyStream` | The session is installed before the announce with nothing keeping stream frames off the wire: the visitor's `FrameOpen` reaches the agent ahead of the auth response, `DecodeAuthResponse` rejects it, and `OnConnect` never fires. |
| **C-3** | **An announce that fails leaves nothing behind, and leaves it behind atomically.** The relay is held inside the announce that is about to fail, with a visitor proven to be inside `ServeHTTP` and therefore queued on `r.mu` behind it. When the write fails, that visitor is answered `404` — never forwarded onto the connection whose announce just failed. Afterwards the hostname still answers `404`, and a second agent takes the same subdomain and serves a visitor `200`. | `relay.TestAFailedAnnounceLeavesNothingBehind` | The failure path does not undo the session install that now precedes it, or undoes it in a **second** critical section — `sync.RWMutex` hands the released lock to the queued reader before the next writer, so that visitor is *guaranteed* to be shown the session on its way out and forwarded onto a dead connection (`SPEC.md` §3.3) — or the route or the name outlives a handshake that never completed. |

## Evidence these can fail for the right reason

Run before implementation against `8540b89`; against a **bare reorder** — the
fix `roadmap/BACKLOG.md` item 2 predicted, session created and installed above
`writeFrame` with no lock held across it; and against a **second critical
section** — `SPEC.md` §3.3's forbidden implementation, identical except that
the `delete` moves out of the announce's critical section into one of its own.

**At `8540b89`, unchanged:**

```
--- FAIL: TestVisitorIsNotAnsweredBeforeTheSessionExists (0.00s)
    a visitor was answered 404 "No tunnel is open for sessionorder.pumasi.link"
    while the relay was still inside the announce — the URL is reachable before
    the session that serves it exists (spec/0003 §1)
--- PASS: TestTheAnnounceReachesTheAgentBeforeAnyStream (0.75s)
    the visitor sent during the announce was answered 404 "No tunnel is open
    for announcefirst.pumasi.link" (err=<nil>)
--- FAIL: TestAFailedAnnounceLeavesNothingBehind (0.00s)
    a visitor was answered 404 "No tunnel is open for doomed.pumasi.link"
    while the relay held r.mu across the announce
```

**Against the bare reorder** — and note that C-1's answer is now the *wrong*
one rather than the honest one, and that C-2's 10.76s is `OnConnect` never
firing at all:

```
--- FAIL: TestVisitorIsNotAnsweredBeforeTheSessionExists (0.00s)
    a visitor was answered 502 "the tunnelled service did not respond" while
    the relay was still inside the announce — ...
--- FAIL: TestTheAnnounceReachesTheAgentBeforeAnyStream (10.76s)
    the agent never received an auth response it could decode — something
    reached it ahead of the announce (spec/0003 §2)
--- FAIL: TestAFailedAnnounceLeavesNothingBehind (0.75s)
    the visitor waiting on r.mu was answered 502 "the tunnelled service did not
    respond", want 404 — it was shown a session on a connection whose announce
    had already failed (spec/0003 §3.3)
```

**Against the second critical section, C-1 and C-2 pass and C-3 fails 20 runs
out of 20**, with the same text as the bare-reorder line above. That is the
only evidence C-3's own clause has, and it is the whole reason C-3 was rewritten
(`SPEC.md` §7).

### What each red run does and does not prove

- **C-1 is the case that measures this defect.** It is red at `8540b89` in
  0.00s, on the exact string job `0047` recorded, produced by a relay that is
  provably blocked rather than by a race the case won.
- **C-2 is vacuous at `8540b89` and cannot be otherwise.** Nothing is forwarded
  during the announce there — the visitor is 404'd instead — so nothing can
  precede the announce on the wire and the property holds trivially. C-2 only
  becomes load-bearing on an implementation that installs the session first,
  which is the class the fix belongs to. It is the case that tells the shipped
  fix apart from the bare reorder, and the 10.76s run above is the whole of its
  justification for existing.
- **C-3's red at `8540b89` is not evidence about C-3's clause.** It fails there
  on its first assertion, which is C-1's property applied to the failing path
  and is a precondition for asking C-3's own question at all. The second
  critical section is the tree that isolates the clause, and it is the one that
  matters.

An honest summary: **C-1 measures the defect and its repair; C-2 measures that
the repair is not the broken version of itself; C-3 measures that the repair's
own failure path is atomic.** None of the three measures the *incidence* of the
defect in production, which nothing on this host has managed to (`INTENT.md`).

## Not covered here, deliberately

- **The ssh path.** It has no handshake write to race and `spec/0002` §3.4
  already ordered it. A case asserting a property that path has always had
  would be an assertion that cannot fail (L-006).
- **The teardown race.** Still a 404, still correct in outcome, not ranked —
  `SPEC.md` §5.
- **The flake rate.** A measurement with its run count in the merge commit, not
  an acceptance criterion. A rate is evidence about a suite.
- **Contention from holding `r.mu` across the announce.** `SPEC.md` §3.2 bounds
  it by argument and names the fix if it is ever observed. Asserting a timing
  bound would be the sleeping test this suite refuses to contain.
- **Whether the Stage 1 gate reading survives** — Q-024, the steward's — and
  **deployment** — Q-014.

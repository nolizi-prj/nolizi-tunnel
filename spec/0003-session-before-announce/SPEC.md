# Spec 0003 · Install the session before the announce, on the HTTP path

**Status:** frozen at spec review · **Intent:** [`INTENT.md`](INTENT.md) ·
**Backlog:** `roadmap/BACKLOG.md` item 2 (post-`b3d251d` pass).

## 1 · The rule

> A tunnel's URL leaves this relay only after the session that serves it is
> installed and visible to `ServeHTTP`, and the announce carrying that URL is
> the first thing its connection carries. A visitor that acts on the URL
> immediately is served, or waits for the announce to finish — it is never
> told the tunnel is not there.

This is [`spec/0002`](../0002-bind-before-announce/SPEC.md) §1 restated for the
thing that answers an HTTP hostname. It is a **new** rule statement rather than
an amendment to that spec because what answers here is a `tunnelSession`, not a
listener; because the mechanism is a lock rather than a bind; and because
`spec/0002` §2 explicitly reasons that the session must be created *after* the
handshake response — a claim this spec contradicts and has to carry its own
argument for. Folding a contradiction into a frozen document as an "amendment"
is how a rule forks
([L-007](https://github.com/pumasi-ai/pumasi/blob/main/lessons/L-007-restating-a-rule-forks-it.md)).

**"Announce"** means the auth-response frame written at `relay/relay.go:205`.
The ssh path is out of scope and §5 says why.

## 2 · Why the obvious reorder is wrong, and what actually blocks it

`spec/0002` §2 argued that `mux.Server(conn)` could not be created before the
handshake response, because it *"would put the multiplexer's reader on a
connection the agent is still using for the raw handshake."* That half of the
argument is **wrong and this spec drops it**: `mux.Server` starts a `readLoop`
that blocks in `core.DecodeFrame(s.conn)`, and the agent sends nothing between
its auth request and its first mux frame, so there is nothing on the wire for
that reader to steal.

The half that is right, and that `roadmap/BACKLOG.md` item 2 does not have, is
about **writes**:

- `Session.Open` writes a `FrameOpen` on the session's connection
  (`mux/session.go`), under the session's own `writeMu`.
- The relay's announce is written raw — `r.writeFrame(conn, okFrame)` — and
  takes no such lock. It cannot: the session may not exist yet.
- The agent is inside `core.DecodeFrame(conn)` waiting for exactly one frame
  (`agent/agent.go:159`) and decodes whatever arrives first as its auth
  response.

So a bare reorder (create the session, install it, then announce) trades a
404 for a corrupted handshake: a visitor forwarded in the new window puts a
`FrameOpen` on the wire ahead of the auth response, the agent fails
`DecodeAuthResponse`, and the tunnel drops and reconnects. Measured, not
assumed — §6.

**Both halves have to hold at once:** the session must be installed *before*
the announce, and no stream frame may reach the wire *until* the announce is
out. Anything that gives only the first is not a fix.

## 3 · The change

### 3.1 · `r.mu` covers the announce, and does both jobs

`relay/relay.go`, `ServeAgent`. The session is created and installed, and the
announce written, inside one `r.mu.Lock()`:

    session := muxSession{s: mux.Server(conn)}
    r.mu.Lock()
    r.sessions[resp.AgentID] = session
    err = r.writeFrame(conn, okFrame)
    if err != nil {
        delete(r.sessions, resp.AgentID)
    }
    r.mu.Unlock()

`r.mu` is the same lock `ServeHTTP` takes (`RLock`) to read `r.sessions`. That
one fact makes it serve both halves of §2:

- **Ordering.** A visitor arriving in the window blocks in `RLock` instead of
  reading `session == nil`. When it proceeds, the session is there.
- **Exclusion.** Because that visitor cannot reach `OpenStream` until it has
  the session, and it cannot have the session until the lock is released, no
  `FrameOpen` can precede the announce on the wire. The lock is the wire's
  ordering, not just the map's.

Nothing else may take `r.mu` inside that critical section, and nothing does:
`writeFrame` encodes and writes, and neither the registry nor the port pool is
touched there.

### 3.2 · What holding a lock across a socket write costs, and what bounds it

This is the one design objection this change has to answer rather than avoid.
While an announce is in flight, **every** visitor lookup on the relay waits —
not just that tunnel's. Three things bound it:

1. **The frame is small.** An auth response is under `core.MaxHandshakeBytes`
   (4096). A fresh TCP connection's send buffer absorbs it without the peer
   having read anything, so the write returns without waiting on the agent.
2. **The handshake deadline is still armed.** `conn.SetDeadline(now +
   HandshakeTimeout)` is set at `relay/relay.go:145` and is cleared *after*
   the announce. That ordering is now load-bearing rather than incidental, and
   this spec fixes it: the announce write is deadline-bounded, so a peer that
   never reads costs at most `HandshakeTimeout` (default 10s) and then fails.
3. **It is one write, not a conversation.** No read, no dial, no allocation
   under the lock.

The alternative considered and rejected: a write gate inside `mux.Session`, so
that the exclusion is per-session and the relay's lock stays narrow. It is
strictly better on contention and strictly worse on everything else — it adds
public API to the package the whole product rests on, to express an ordering
that one existing lock already expresses, for a window measured in
microseconds. If contention here is ever observed rather than imagined, that
is the fix to reach for, and this paragraph is the note that it exists.

### 3.3 · A failed announce takes the session with it

The install now precedes the write, so the write's failure path has to undo it.
Deleting the entry happens **inside the same critical section as the write**,
not in a second `Lock()` afterwards: between two critical sections a visitor
would find a session on a connection that is already known to be broken and
would be given a stream on it, which is `spec/0002` §6.4's defect — the client
learning an outcome before the state behind it is true — in this spec's own
change.

After the lock is released, the failed path closes the session (which closes
the connection) and then runs the existing `abandon()`: release the TCP
listener if one was bound, and unregister the tunnel.

**This clause is observable, and §7 records the review that made it so.** A
visitor queued on `r.mu` while the doomed announce is in flight is handed the
released lock before the next writer can take it (`sync.RWMutex` grants waiting
readers ahead of a subsequent `Lock`), so a second critical section is not a
narrow window here — it is a guarantee that that visitor sees the session on
its way out and is forwarded onto a connection already known to be dead.

### 3.4 · The `session == nil` comment stops mis-reading its own branch

At `8540b89` that branch reads:

    // The registry and the session map disagreed: the agent went away
    // between the lookup and here.

That describes the teardown race only, and at `8540b89` it is not how the
branch is usually reached — the startup window this spec closes is the other
way, and the comment's silence about it is part of why the defect survived a
reading. After this change the teardown race is the **only** way the branch is
reached, and the comment says so, and says that the startup direction is closed
and by what.

## 4 · What does not change

- The wire protocol. `core.AuthResponse`, `core.Frame` and the frame order on
  the connection are exactly what they were: auth request, auth response, then
  mux frames.
- `mux`. No new API, no behaviour change, not one line.
- The TCP path. `listenTCP` stays above the announce and `go serveTCP(...)`
  stays below it; `spec/0002` §2's accept-queue argument is untouched and
  correct.
- Teardown. The deferred `delete(r.sessions, ...)` / `releaseTCP` /
  `UnregisterAgent` / `session.Close()` block is unchanged.
- The ssh path (§5).
- `notFound`'s text and status, and everything a visitor sees.

## 5 · Scope this spec deliberately does not take

- **The ssh path.** `ServeSSH` registers its session before it does anything
  else — it has no handshake write to race — so the defect does not exist
  there, and `spec/0002` §3.4 already put its ordering right. A case asserting
  a property that path has always had would be an assertion that cannot fail
  (L-006), and there is not one here.
- **The teardown race.** A visitor arriving between an agent's disconnect and
  the registry's `UnregisterAgent` still gets a 404. That is a different
  ordering, it is *correct* in outcome (the tunnel really is gone; only the
  message is imprecise), and it is not ranked. §3.4 makes the comment say so
  rather than making the code chase it.
- **The flake rate.** Reported as a measurement with its run count, not
  asserted by a test. `spec/0002` §5 said this and it is still true: this
  suite is green 90–95% of the time against defects of this class, so a rate is
  evidence about a suite, not an acceptance criterion for a change (L-006).
- **Whether the Stage 1 gate reading survives.** Q-024, the steward's.
- **Deployment.** Q-014.

---

## 6 · Evidence that the case fails for the right reason

Taken before implementation, on this host, and reproduced verbatim in
[`acceptance/CASES.md`](acceptance/CASES.md).

| Tree | C-1 | C-2 | C-3 |
|---|---|---|---|
| `8540b89`, unchanged | **FAIL** in 0.00s — `404 No tunnel is open for sessionorder.pumasi.link` | PASS (vacuously — nothing is forwarded there to precede the announce) | **FAIL** in 0.00s, on C-1's property, **not on its own clause** — `a visitor was answered 404 … while the relay held r.mu across the announce` |
| Bare reorder — session installed above `writeFrame`, no lock held across it | **FAIL** — `502 the tunnelled service did not respond` | **FAIL** in 10.76s — `the agent never received an auth response it could decode` | **FAIL** — `502 …, want 404` |
| Second critical section — §3.3's forbidden implementation, everything else identical | PASS | PASS | **FAIL 20 runs out of 20** — `502 …, want 404` (§7) |
| This spec's change | PASS | PASS | PASS |

Two things in that table are the substance of §2 and neither was known when the
backlog entry was written.

**The bare reorder does not merely fail to fix C-1 — it makes the answer
worse.** At `8540b89` the visitor is told, honestly, that no tunnel is open. On
the bare reorder it is handed a stream on a session whose peer is still in the
raw handshake, and gets `502 the tunnelled service did not respond`: a wrong
answer about a tunnel that is about to work, where there had been a right
answer about one that did not yet.

**And the tunnel does not open at all.** C-2's 10.75s is `OnConnect` never
firing: the `FrameOpen` arrived first, `DecodeAuthResponse` rejected it, and
the agent dropped the connection and went into its reconnect backoff. A change
that turned a 1-in-120 404 into a dropped tunnel would have passed every test
this repository had before this spec, including 500 runs of the suite.


---

## 7 · Amendment 1 — C-3 named a clause it could not exercise

**Amended 2026-08-31, after the first spec review and before the freeze.
gemini approved (`reviews/20260831-193215-spec-gemini.md`); qwen objected, with
a citation, and the objection was right** (`reviews/20260831-193215-spec-qwen.md`):

> `acceptance/CASES.md` C-3 names §3.3, but its stated behavior only checks
> post-failure outcomes … An implementation that deletes in a second critical
> section — or late enough to leave only an internal leak after cleanup — could
> still satisfy C-3's stated assertions. Thus C-3 can pass without exercising
> the exact §3.3 clause it names, which is the L-006 failure mode the
> acceptance document says it avoids.

C-3's row claimed a failure mode — *"or does it in a second critical
section"* — that nothing in the case could detect. That is L-006 inside a
document whose opening paragraph is about L-006, and no assertion was weakened
to make it go away.

**What changed is the case, not the rule.** C-3 now holds the relay inside the
announce that is about to fail, using the same brake C-1 uses on the announce
that succeeds, and puts a visitor onto `r.mu` behind it — proven to be inside
`ServeHTTP` by the edge handler rather than assumed to have got there. The
`sync.RWMutex` property in §3.3 is what makes the observation a guarantee
rather than a race.

**Measured against the implementation the clause forbids** — identical in every
respect except that the `delete` moves into a second `r.mu.Lock()`:

```
--- FAIL: TestAFailedAnnounceLeavesNothingBehind (0.75s)
    the visitor waiting on r.mu was answered 502 "the tunnelled service did not
    respond", want 404 — it was shown a session on a connection whose announce
    had already failed (spec/0003 §3.3)
```

**20 runs, 20 failures.** Before the amendment that implementation passed C-3
every time.

**A consequence worth stating rather than presenting as a gain:** the rewritten
C-3 also goes red at `8540b89` — but on its *first* assertion, which is C-1's
property applied to the failing path, not on its own clause. §6's table says so
in those words. C-3's evidence for its own clause is the 20-in-20 above and
nothing else.

---

## 8 · The review record, including the two runs that produced nothing

CHARTER §3 requirement 1 makes the builder the spec author and a different
family the reviewer. What that cost here, in full, because a partial review is
a record and a review that was abandoned has to say so in the open:

| Run | Target | gemini | qwen | kimi |
|---|---|---|---|---|
| `20260831-193215` | this spec, first version | **APPROVE** (`reviews/20260831-193215-spec-gemini.md`) | **OBJECT**, cited, and correct (`reviews/20260831-193215-spec-qwen.md`) | — |
| `20260831-194403` | after §7's amendment | **APPROVE** (`reviews/20260831-194403-spec-gemini.md`) | UNREACHABLE — `curl` timed out at its 600s ceiling | — |
| `20260831-195536` | after §7's amendment | — | UNREACHABLE — same | UNREACHABLE — same |

**The family that objected could not be reached again.** That is recorded as
what it is: not a withdrawal, not a second approval, and not breadth. The
objection was answered on its merits rather than outlasted — §7's case fails 20
runs out of 20 against the implementation qwen said C-3 could not distinguish,
which is a stronger answer than a re-read would have been.

**One mechanical observation for whoever owns `tools/review.sh`, offered
because it is a measurement and not a complaint.** The `recruit`-driven
families take their target inlined in the prompt, and this spec's bundle grew
as the spec did. qwen answered a **23,962-byte** bundle in about ninety
seconds and then timed out twice on a **28,986-byte** one; kimi timed out on
the same 28,986 bytes, both after receiving ~15.7 KB of a streamed reply. The
script's own header already records glm as "one in two drivable" at the same
600s ceiling. The threshold this crossed sits somewhere between those two
sizes, and the effect is that **a spec gets harder to review the more carefully
it is written**. Not this packet's to fix, and not filed as a decision entry by
this seat.

# Intent · A URL is not announced until the session that serves it is installed

**Published 2026-08-31 · `roadmap/BACKLOG.md` item 2 (post-`b3d251d` pass).**
CHARTER §2.1 gives an intent statement a 24-hour veto window. `roadmap/STAGE.md`
says `Alpha`, so CHARTER Part 0 applies: the window does not hold the work, and
a veto reverts rather than prevents. Recorded rather than pretended — this
statement was published at the same time as the work, not 24 hours ahead.

## The gap

`ServeAgent` (`relay/relay.go:139`) does four things in this order at
`8540b89`:

1. `authorize` → `r.registry.Register(tunnel)` (`:295`). **The hostname now
   resolves**, and `ServeHTTP`'s `r.registry.Lookup` will find it (`:331`).
2. The public TCP port is bound, if one was asked for (`:175`). This is
   `0041`'s fix (`spec/0002`) and it is correctly placed.
3. `r.writeFrame(conn, okFrame)` (`:205`). **The agent, and its user, now hold
   the URL.**
4. `session := muxSession{s: mux.Server(conn)}` and
   `r.sessions[resp.AgentID] = session` under `r.mu` (`:213`–`:216`).
   **Only now can a visitor be forwarded anywhere.**

A visitor arriving between 3 and 4 passes `Lookup`, reads `session == nil`
(`:316`ff), and is answered `404 No tunnel is open for <host>` by `notFound`
(`:380`, `:384`).

This is `spec/0002` §1 one path over. That rule — *"a public address leaves
this relay only after the listener that answers it is bound"* — was written
about a TCP port, but the sentence it is built on is general and is quoted in
`spec/0002` §6.4 as such: **when the client learns the outcome, the state
behind it is already true.** On the HTTP path the thing that answers a URL is
not a listener but a session, and the relay announces before it has one.

## The state of the evidence, which is weaker than a reproduction

This defect is established by **reading**, plus one **measured symptom that
matches the reading**, and this statement does not claim more than that.

- Job `0047` recorded the exact string twice in 240 runs of
  `TestConcurrentVisitors`: `No tunnel is open for myapp.pumasi.link`.
- Job `0059` could **not** reproduce it at `b3d251d` on this host: 500 runs of
  `go test -count=1 ./...`, 100 with `-cover`, 40 whole-gate runs, and a
  targeted out-of-tree probe of ~5,400 visitor requests fired at the earliest
  instant `OnConnect` can hand over the URL — **zero `404`s**.

So the incidence on this machine is below what ~5,400 attempts detect. The
window is a few instructions wide and widens under load, which is the
condition `0047`'s failures came from and this host did not reproduce. A green
suite is therefore not evidence about this change in either direction
([L-006](https://github.com/pumasi-ai/pumasi/blob/main/lessons/L-006-tests-that-cannot-fail.md)),
and the acceptance case in
[`acceptance/CASES.md`](acceptance/CASES.md) is designed to fail at `8540b89`
by construction rather than by losing a race.

## What the backlog entry did not know, and what changes because of it

`roadmap/BACKLOG.md` item 2 says the fix is that *"the insert can simply move
above"* the announce, because *"there is nothing between step 3 and step 4 that
needs the response to have been sent."* Reading `mux/session.go` for this work
found that there **is** such a thing, and it is not a scheduling nicety:

**The announce and every stream frame share one connection, and the announce
has to be first on it.** `mux.Server(conn)` takes ownership of `conn`
(`mux/session.go:82` starts `readLoop`), and `Session.Open` writes a
`FrameOpen` to that same `conn` (`mux/session.go:150`ff) the moment a visitor
is forwarded. The relay's auth response is written raw, outside the session's
`writeMu`. So a bare reorder — session created and installed, *then*
`writeFrame` — replaces a 404 with something worse: a visitor arriving in the
new window makes the relay put a `FrameOpen` on the wire **ahead of** the auth
response, and the agent, still in `core.DecodeFrame(conn)` waiting for its
reply (`agent/agent.go:159`), decodes a stream frame as its handshake answer
and drops the connection.

This is not a hypothetical. The acceptance case in this spec was run against a
bare reorder and it goes red there too, with a different failure text; both
runs are recorded in [`SPEC.md` §6](SPEC.md#6--evidence-that-the-case-fails-for-the-right-reason).

The fix therefore has two halves that cannot be separated: **install the
session before the announce, and make the announce exclusive of stream
frames while it happens.** `r.mu` — the lock `ServeHTTP` already takes to read
`r.sessions` — does both, and `SPEC.md` §3 says why it is allowed to.

## What should be true instead

A URL leaves this relay only after the session that serves it is installed and
reachable by `ServeHTTP`. A visitor that acts on the URL the instant it is
printed is either served or made to wait for the announce to finish — never
told that nothing is there.

## What this does not do

- **No deploy.** `pumasi.link` runs a pre-`83fd9f7` binary; restarting it is
  `pumasi/DECISIONS.md` **Q-014**, open and explicitly outside CHARTER Part 0's
  proceed-on-default rule. This change will be the fourth merged and
  undeployed one behind that blocker, and that is the correct outcome, not an
  oversight.
- **No claim on the Stage 1 gate.** Whether a gate reading survives is
  **Q-024**, the steward's, both ways. Run counts are reported; a stage is not.
- **No `roadmap/` edit.** `BACKLOG.md` and `STAGE.md` are the product
  manager's. The measured runs are handed to that seat in the digest.
- **No amendment of `spec/0001` or `spec/0002`.** **Q-030** is open on whether
  a builder may amend a frozen case, and it already has two contradictory
  instances across two products. This spec adds a new case of its own and
  touches no frozen one, which is ordinary and is not what Q-030 asks.

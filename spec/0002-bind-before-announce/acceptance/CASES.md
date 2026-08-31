# Acceptance cases · Spec 0002

Frozen at spec review, before implementation. The builder may not edit them; a
case that is wrong is fixed by amending the spec in the open and taking a fresh
cross-family spec review (CHARTER §3 requirement 2).

Every case below names **what execution makes it fail**, and every case is
**deterministic** — none of them dials and hopes to lose a race. A rerun of the
existing flaky tests is not an acceptance case here: they pass 90–95% of the
time against the defect, so a green run from them proves nothing (L-006).

## How the ordering is asserted without a race

The defect is an ordering, and an ordering is observable without timing if you
make the second step **fail**. Occupy the relay's entire public port range with
a foreign listener before the agent connects. Then:

- **bind-then-announce** (correct): the bind fails, so no address is ever
  composed — the caller gets an error and nothing else.
- **announce-then-bind** (the defect): the caller is handed a working-looking
  address *first*, and the failure arrives after it.

The two outcomes differ in what the client holds, not in when it holds it. No
sleeps, no polling, no interleaving assumptions.

The second family of cases uses a relay configured
`BaseDomain: " pumasi.link"` — an ordinary domain with a leading space, which
is what a stray character in a systemd unit produces. `relay.New` accepts it
(it tests only for the empty string) while `core.NewRegistry` trims it, so the
lookup key the two serve paths build never matches and the pre-change registry
lookup fails **every** time rather than occasionally. That turns the silent
skip from a race into a certainty.

> **Amended 2026-08-31** — see [`../SPEC.md` §6](../SPEC.md#6--amendment-1--the-reachability-claim-was-wrong).
> These cases first used `BaseDomain: ""`, which `relay.New` rejects outright,
> so they failed on `relay: BaseDomain is required` instead of on the defect.
> B-2 first asserted that a refused greeting must not contain the address
> string, which the bind error legitimately does. Both were errors in the spec,
> amended in the open and re-reviewed rather than edited by the builder.

## The cases

| # | Case | Go test | Fails when |
|---|---|---|---|
| **B-1** | **Agent path, ordering.** With the whole TCP range held by a foreign listener, an agent requesting TCP receives an error and **no** public address — never an auth response carrying `TCPAddr` followed by a failure. | `relay.TestAgentIsNotToldAnAddressItCannotBeGiven` | The announce precedes the bind, so a `TCPAddr` reaches the client before the bind is attempted. |
| **B-2** | **SSH path, ordering.** Same construction over a real `ssh` client: the terminal is told the bind failed, and is **not** given `sshGreet`'s announcement — the address line plus `forwarding to your local port over this ssh session`. The two are distinguished by which of the relay's two messages was sent, not by whether the address appears, because the bind error names the address it could not bind. | `relay.TestSSHIsNotToldAnAddressItCannotBeGiven` | The ssh greeting is composed or sent before the bind result is known. |
| **B-3** | **Agent path, no lookup.** On a relay whose every registry lookup fails (`BaseDomain: " pumasi.link"`), the address in the auth response is backed by a live listener: dialling it connects and carries bytes end to end. | `relay.TestTCPOnlyRelayAnnouncesAnAddressThatAnswers` | The bind is conditional on a registry lookup that this configuration always fails. |
| **B-4** | **SSH path, the silent skip.** On the same relay, an ssh session requesting TCP is either given an address that answers, or refused — never greeted with an address that nothing listens on. | `relay.TestSSHTCPOnlyRelayNeverGreetsADeadAddress` | `lookupErr != nil` skips the bind with no `else`, and the greeting prints `TCPAddr` regardless. |
| **B-5** | **A refused bind does not leak the port.** After B-1's refusal, the foreign listener closes and a second agent asking for the same port is served — the port went back to the pool **and** to the registry. | `relay.TestPortReturnsToThePoolWhenTheBindFails` | The failure path returns before `pool.Release`, stranding the port until restart; or it tells the client before `UnregisterAgent`, so the pool and the registry disagree and the next agent is refused a free port ([`../SPEC.md` §6.4](../SPEC.md#64--the-refusal-path-tells-the-client-before-it-has-let-go)). |
| **B-6** | **The happy path is unchanged.** An ordinary TCP tunnel still carries bytes both ways, and the address announced is the address that works. | `relay.TestAnnouncedAddressIsTheWorkingAddress` | The split of `bindTCP` into listen and serve drops the accept loop, the session wiring, or the byte pipe. |

## Evidence these can fail for the right reason

Before merging, each case is run against a worktree with the change **absent**
(`01ef62b`) and must go red there, with the failure text recorded in the
release note and the digest (L-006: *test that the test fails*). Cases that go
**green** at `01ef62b` are reported as green rather than quietly dropped —
B-2 is expected to be one of them, because the ssh path's ordering was already
correct, and reporting that is the point.

## Not covered here, deliberately

- **The flake rate itself.** Reported as a measurement, 40 runs of each
  invocation before and after, not asserted by a test. A rate is evidence about
  a suite, not an acceptance criterion for a change.
- **`https://` working, TLS, or the certificate** — `BACKLOG.md` item 1(ii).
- **Durability across a relay restart** — `BACKLOG.md` item 3. A bound listener
  still dies with the process.
- **Whether the Stage 1 gate reading survives** — Q-024, the steward's.

# Intent · A public address is not announced until something answers it

**Published 2026-08-31 · `roadmap/BACKLOG.md` item 2.**
CHARTER §2.1 gives an intent statement a 24-hour veto window. `roadmap/STAGE.md`
says `Alpha`, so CHARTER Part 0 applies: the window does not hold the work, and
a veto reverts rather than prevents. Recorded rather than pretended — this
statement was published at the same time as the work, not 24 hours ahead.

## The gap

The relay tells an agent its public TCP address, and only afterwards binds the
socket that address names. Between those two moments the address is a lie, and
the product's own test suite loses that race often enough to be measured.

Measured at `01ef62b` on this machine, before any change — not read off the
backlog, which recorded different numbers (L-007: verify against the artefact):

```
$ go test -count=1 -cover ./...   × 40   →  4 failures in 40   (10.0%)
$ go test -count=1 ./...          × 40   →  0 failures in 40
$ go test -count=1 ./...          × 120  →  6 failures in 120  (5.0%)
```

Every failure is the same refused dial against the same public port:

```
--- FAIL: TestTCPPortReleasedWhenAgentDisconnects
    tcp_test.go:208: port should be live before disconnect:
                     dial tcp 127.0.0.1:34000: connect: connection refused
```

Two facts here are worth more than the rates. First, the plain invocation
returned **0 in 40** and then **6 in the next 120** — a 40-run sample cannot
pin this number, which is the Q-024 lesson arriving again from the other side.
Second, the failure surfaces on **four** tests, not the three the backlog
records: `TestRawTCPCrossesTheTunnel`, `TestTCPPortReleasedWhenAgentDisconnects`,
`TestConcurrentTCPClients`, and `TestServerSpeaksFirstOverTCP`. That it moves
between tests run to run is the evidence that it is not any test's artifact.

## What is actually wrong, on each path

**The agent path announces before it binds.** `relay/relay.go:175` writes the
auth response carrying `TCPAddr`; `relay/relay.go:194` calls `bindTCP`. A bind
failure is therefore reported *after* the address was handed out, and the agent
holds a public address that nothing is listening on. `--tcp-port`'s documented
promise — "keeps an address across reconnects" — is intermittently false at the
one moment it matters.

**The ssh path does not have that shape, and this statement does not repeat the
claim that it does.** `roadmap/BACKLOG.md` item 2 and `pumasi/DECISIONS.md`
Q-024 both say `relay/sshingress.go:182` "has the same shape". Read at
`01ef62b`, it does not: `bindTCP` is called at **:182** and the address is
composed at **:190–:192** and printed at **:209**. That path is already
bind-then-announce. The inherited claim is wrong, and Q-024 exists because of
exactly this failure mode.

What the ssh path has instead is a **silent skip**. At `:180–:181`:

```go
tunnel, lookupErr := r.registry.Lookup(resp.Subdomain + "." + r.cfg.BaseDomain)
if lookupErr == nil {
        if err := r.bindTCP(...); err != nil { ... }
}
```

There is no `else`. When the lookup fails the bind is skipped, nothing is
logged, and the greeting still prints `resp.TCPAddr` — a public address with no
listener and no trace in the log. The agent path shares the defective lookup
and at least logs it (`relay.go:192–193`), then carries on and serves a dead
address anyway.

**How reachable this is — corrected after the first spec review** (see
[`SPEC.md` §6](SPEC.md#6--amendment-1--the-reachability-claim-was-wrong)). The
first version of this statement said a TCP-only relay started with `-domain ""`
made the lookup fail for every session. That is **wrong**, and the frozen case
built on it failed for that reason rather than for the defect:
`relay.New` refuses an empty `BaseDomain` outright
(`relay/relay.go:93–95`), so no such relay ever runs.

The reachable trigger is narrower and is a plain operator slip: **a leading
space in `-domain`**. `relay.New` checks only `cfg.BaseDomain == ""`, so
`-domain " pumasi.link"` starts normally, while `core.NewRegistry` trims the
value and the lookup key does not — leaving
`SplitHost("remotebox. pumasi.link", " pumasi.link")` returning
`ErrForeignHost` for **every** session. Verified, alongside the spellings that
are *not* affected, because a normalisation that mostly works is the kind that
gets trusted:

```
base=" pumasi.link"   relay.New OK; SplitHost -> "" err=core: host is not under the tunnel base domain
base="pumasi.link "   relay.New OK; SplitHost -> "remotebox" err=<nil>
base="PUMASI.link"    relay.New OK; SplitHost -> "remotebox" err=<nil>
base="pumasi.link."   relay.New OK; SplitHost -> "remotebox" err=<nil>
base="pumasi..link"   relay.New OK; SplitHost -> "remotebox" err=<nil>
```

So the silent skip is a live defect on a misconfiguration that a stray space in
a systemd unit produces, not on the ordinary path. That is a smaller claim than
this statement first made, and it is the one the evidence supports.

So both paths can hand out an address nothing answers. They arrive there by
different routes, and the fix has to close both.

## What should be true instead

An address is announced only after the socket that answers it is bound. If the
socket cannot be bound, the caller is told that instead of being told an
address — on both paths, and whether the failure is the port or the lookup.

## What this does not do

- **No deploy.** `pumasi.link` runs a pre-`83fd9f7` binary and restarting it is
  `pumasi/DECISIONS.md` **Q-014**, open and explicitly outside CHARTER Part 0.
  Nothing here asks for a restart.
- **No stage claim.** Whether a gate reading survives a flaky suite is
  **Q-024**, and it is the steward's, both ways. This work makes the question
  moot; it does not answer it.
- **No `roadmap/` edit.** Those files are the product manager's, written 90
  minutes before this job. The measured runs are handed over in the digest for
  that seat to record.

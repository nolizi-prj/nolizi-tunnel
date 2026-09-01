# Intent · A name belongs to whoever proved they held it before

**Published 2026-09-01 · `roadmap/BACKLOG.md` item 2, *"A subdomain belongs to
nobody, and nothing survives a relay restart"*.** CHARTER §2.1 gives an intent
statement a 24-hour veto window. `roadmap/STAGE.md` says `Alpha`, so CHARTER
Part 0 applies: the window does not hold the work, and a veto reverts rather
than prevents. Recorded rather than pretended — this statement was published at
the same time as the work, not 24 hours ahead.

## The gap

Three facts, re-read in the tree at `1853218` rather than carried:

1. **`Tunnel.Reserved` is a field that describes an intention nothing
   enforces.** It is declared at `core/route.go:127`, assigned once at
   `relay/relay.go:297` as `req.Subdomain != "" && req.Token != ""`, and
   `grep -rn "\.Reserved" --include=*.go` returns **nothing at all** — not in
   `relay/`, `core/`, `cmd/`, the tests, or the console.
2. **Nothing checks the token that sets it.** `cmd/pumasi-relay` defines twelve
   flags and none of them is an auth flag, so `AllowAll` (`relay/relay.go:40`,
   installed at `:97` when `cfg.Auth` is nil) is the only authenticator that
   binary can run, and its `Authorize` is `return nil`.
3. **The registry and the port pool are plain in-memory maps.** `Registry` is
   `byName`/`byTCP`/`byAgent` under a `sync.RWMutex` (`core/route.go:143`–`:147`);
   `PortPool` is `inUse`/`reserved` under a `sync.Mutex` and its own doc comment
   says it *"does no I/O"*. There is no persistence path anywhere in `core/` or
   `relay/`.

`roadmap/VALUE.md` claim 2 sells *"permanent"* and *"stable across restarts"*.

## The two windows, which are different sizes and have different fixes

A name can be taken from the party using it in exactly two ways, and conflating
them is how this item gets mis-sized.

- **(a) The reconnect gap, with the relay up.** `ServeAgent`'s deferred cleanup
  calls `registry.UnregisterAgent` and `releaseTCP` the instant an agent's
  connection ends. The agent's own `Run` loop redials after a backoff that
  starts at one second and doubles to thirty (`agent/agent.go`). For that whole
  interval the name is free and the port is free, and any anonymous agent may
  take either. **This needs no persistence to fix. It needs an owner.**
- **(b) The relay restart.** Everything goes at once, for everybody, and no
  reconnect recovers a name someone else has since been handed. **This needs the
  ownership record to outlive the process.**

Window (a) is open right now on `pumasi.link`, and it is not hypothetical: the
live relay carries `sshsteward` — this machine's port 22, `RESOURCES.md` §4 —
beside a second tunnel, `skk6g7tyrs`, opened by an agent nobody in this fleet
can identify. That is `AllowAll` working exactly as the backlog entry describes,
observed rather than argued.

## What this found that changes how the item should be read

**A restart cannot be made free, and persistence is not what bounds its cost.**

`pumasi-ops/tools/pumasi-tunnel-keepalive.sh` already restarts `sshsteward`
within five minutes of process death, and the agent's own backoff loop redials
within seconds of a mere connection drop. So the cost of a relay restart to the
steward's ssh route is **already** a bounded outage rather than a lost route —
*provided the name `sshsteward` and the port `20000` are still there when it
redials.* A live TCP connection cannot survive the process at either end, and no
amount of persistence changes that; what persistence buys is that the address is
still yours afterwards.

Which means the half of this item that bears on **Q-014** is the *ownership*
half, not the *durability* half. Ownership prevents a loss that a restart cannot
undo. Durability shortens nothing. This is recorded here because Q-014's premise
is that a restart costs the steward their route, and the accurate statement is
narrower: a restart costs the steward their route **only if someone takes the
name or the port during the gap** — which today nothing prevents.

## What should be true instead

A name, and the public TCP port that goes with it, belong to whoever first
proved they held them, and go on belonging to that party while nobody is
connected. The relay hands a name to a stranger only when nobody has claimed it.
What a restart costs is the connection; never the address.

## What this does not do

- **No deploy, no restart, no change to the running relay.** Five commits sit on
  `main` that `pumasi.link` does not have, three of which change relay
  behaviour; this will be the sixth. Delivering them is `pumasi/DECISIONS.md`
  **Q-014**, open and explicitly outside CHARTER Part 0's proceed-on-default
  rule. Nothing here asks for it, and the only contact this work had with the
  live host is the read-only `/_pumasi/status` endpoint.
- **No account, no signup, no identity.** A token proves continuity, not who
  anyone is. **Q-002** blocks public signup and nothing here approaches it.
- **Nothing is withdrawn from the anonymous case.** An agent that asks for no
  name still gets a generated one; an agent that asks for an unclaimed name by
  name still gets it. `SPEC.md` §6 is the whole decision table and it has no
  cell in which today's behaviour becomes a refusal for a name nobody owns.
- **No `roadmap/` edit.** `BACKLOG.md` and `VALUE.md` are the product manager's,
  and job `0076` ordered them deliberately at `1853218`.
- **No amendment of a frozen spec or a frozen case.** **Q-030** is open on
  whether a builder may amend a frozen case. This spec adds cases of its own and
  touches none.

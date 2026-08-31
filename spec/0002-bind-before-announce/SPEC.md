# Spec 0002 · Bind before announce, on both paths

**Status:** frozen at spec review · **Intent:** [`INTENT.md`](INTENT.md) ·
**Backlog:** `roadmap/BACKLOG.md` item 2.

## 1 · The rule

> A public address leaves this relay only after the listener that answers it is
> bound. If it cannot be bound, the caller is told the failure **instead of**
> an address, never after one.

"Announce" means any of: the auth response frame on the agent path
(`relay/relay.go`), and the greeting written into an ssh terminal
(`relay/sshingress.go` → `sshGreet`). Both carry `AuthResponse.TCPAddr`; both
are covered.

## 2 · Why the current order exists, and what actually blocks reordering

The bind cannot simply be moved above the announce, because `bindTCP` today
does two things at once: it creates the listener **and** starts the accept loop
that forwards into a `tunnelSession`. On the agent path the session is
`mux.Server(conn)`, and `mux.Server` starts a `readLoop` goroutine immediately
(`mux/session.go:82`). Creating it before the handshake response is written
would put the multiplexer's reader on a connection the agent is still using for
the raw handshake. That is why the announce comes first today, and it is a real
constraint rather than an oversight.

The resolution is that **those two things do not have to happen together.**
What makes an address real is the listener; what needs the session is the
forwarding. A visitor arriving in the gap is held in the kernel's accept
queue — its connection is established, and it is served as soon as the accept
loop starts. Nothing is dropped and nothing is answered wrongly.

## 3 · The change

### 3.1 · `bindTCP` splits in two

`relay/tcp.go` gains `listenTCP` and keeps `serveTCP`; `bindTCP` goes away.

    // listenTCP binds the public port and records the listener. After it
    // returns, the address answers.
    func (r *Relay) listenTCP(agentID string, port int) (*tcpListener, error)

    // serveTCP accepts visitors and gives each one a stream. Started once a
    // session exists to forward into.
    func (r *Relay) serveTCP(session tunnelSession, l *tcpListener, subdomain string)

`listenTCP` keeps `bindTCP`'s failure behaviour exactly: on `net.Listen` error
it calls `r.pool.Release(port)` and returns a wrapped error, because a port the
pool believes is in use with nothing listening on it is unrecoverable without a
restart.

### 3.2 · The registry round trip goes away

Both paths currently recover the port they just allocated by calling
`r.registry.Lookup(resp.Subdomain + "." + r.cfg.BaseDomain)`. That lookup is
the source of the ssh path's silent skip, and it is unnecessary: `authorize`
computed the port a few lines earlier.

`authorize` — unexported, two callers — returns it directly:

    func (r *Relay) authorize(req core.AuthRequest) (core.AuthResponse, int, error)

The second result is the allocated public TCP port, `0` when the request was
not a TCP one. **The lookup is deleted from both paths.** The silent skip is
then not a case that is handled correctly; it is a case that no longer exists.
Fixing it by adding an `else` would have left the round trip in place for the
next reader to reintroduce (L-007: the defect was a restatement of a fact the
caller already held).

### 3.3 · The agent path

`relay/relay.go`, `ServeAgent`, in order:

1. `authorize` → `resp`, `port`.
2. If `resp.TCPAddr != ""`: `listenTCP(resp.AgentID, port)`. On failure, write
   an error frame, `UnregisterAgent`, return. **No address has been announced.**
3. Encode and write the auth response. If either fails, close the listener,
   release the port via `releaseTCP`, unregister, return — the listener must
   not outlive the failed announce.
4. Clear the handshake deadline; create the mux session.
5. `go serveTCP(session, listener, resp.Subdomain)`.

### 3.4 · The ssh path

`relay/sshingress.go`, `ServeSSH`, in order:

1. `authorize` → `resp`, `port`.
2. Register the session (it already exists; ssh needs no handshake write).
3. If `resp.TCPAddr != ""`: `listenTCP`. On failure, `sshTell` the error,
   `UnregisterAgent`, return — the greeting never runs.
4. `go serveTCP(...)`, then compose `address` and `go sshGreet(chans, address)`.

This path's ordering was **already correct**; what changes is that its lookup —
and with it its silent skip — is gone, and that it shares one code path with the
agent path for the bind itself.

## 4 · What does not change

- The wire protocol. `core.AuthResponse` is untouched; `TCPAddr` means what it
  meant.
- Port allocation and the pool's forward-walking cursor (`core/portpool.go`).
- The accept loop, the byte pipe, and half-close handling (`relay/tcp.go`).
- Release on disconnect: `releaseTCP` still frees every listener an agent owns.
- HTTP tunnels, which never had a bind step.

## 5 · What this spec deliberately does not claim

- **That the flaky tests will stop failing.** They should, and the run counts
  are reported either way; but a green suite is not the evidence this change is
  correct, because the suite is green 90–95% of the time without it (L-006).
  The evidence is the deterministic cases in
  [`acceptance/CASES.md`](acceptance/CASES.md), which fail against `01ef62b`
  for the stated reason.
- **That the running relay is fixed.** It is not deployed — Q-014.
- **Anything about the Stage 1 gate reading.** Q-024, the steward's.

---

## 6 · Amendment 1 — the reachability claim was wrong

**Amended 2026-08-31, after the first spec review approved
(`reviews/20260831-121728-spec-gemini.md`) and before merge. Re-reviewed as
required by CHARTER §3 requirement 2; the builder did not edit a frozen case
without this.**

Implementation found two of the six frozen cases failing for reasons that were
defects in *this document*, not in the code. Recorded here rather than fixed
quietly, because a frozen case edited on the builder's own judgement is the
thing the freeze exists to prevent.

### 6.1 · `-domain ""` is not a configuration that exists

§3.2 and B-3/B-4 rested on a TCP-only relay started with `-domain ""`, on the
grounds that `core.SplitHost("remotebox.", "")` returns `ErrForeignHost`. That
function does return that error — but the relay never gets there, because
`relay.New` refuses an empty `BaseDomain` before anything else
(`relay/relay.go:93–95`). The case failed with `relay: BaseDomain is required`.

The reachable trigger is **a leading space in `-domain`**. `relay.New` tests
only for the empty string; `core.NewRegistry` trims and lowercases the value
while the lookup key built in the two serve paths does not. So
`-domain " pumasi.link"` starts a working relay whose every registry lookup
fails. Trailing space, uppercase, a trailing dot and a doubled dot are all
**unaffected** — verified, and named here because a normalisation that handles
four spellings out of five is one that gets trusted.

**Nothing about the fix changes.** Deleting the lookup removes the defect on
every trigger, reachable or not. What changes is the configuration B-3 and B-4
use to reach it, and the honesty of the claim in `INTENT.md`.

### 6.2 · B-2 asserted on a string that legitimately contains the address

B-2 required that a refused ssh session's greeting not contain
`127.0.0.1:34500`. The relay's behaviour is correct — it refuses — but the
refusal it prints is the bind error, and the bind error names the address it
could not bind:

```
pumasi: relay: binding public port 34500: listen tcp 127.0.0.1:34500:
bind: address already in use
```

So the case failed on a substring of the *right* answer. That is L-006 from the
other side: an assertion whose failing execution is not the defect.

The property B-2 exists to protect is that the terminal is told **no**, not
handed an address. The relay says those two things in distinguishable words —
a refusal is `sshTell`'s `pumasi: <error>`, an announcement is `sshGreet`'s
address line followed by `forwarding to your local port over this ssh
session`. B-2 asserts on that distinction instead of on the bare address.

### 6.3 · What did not change

B-1, B-5 and B-6 are untouched and passed as written. No acceptance case was
weakened: B-2 became more specific, and B-3/B-4 kept their assertions and
changed only the configuration that reaches the defect.

### 6.4 · The refusal path tells the client before it has let go

§3.3 and §3.4 wrote the failure branch as *write the error, unregister,
return*. Implementation showed that order has the same defect this spec exists
to remove, one step further along.

`listenTCP` releases the port to the pool before it returns its error, but the
registry still records the tunnel until `UnregisterAgent` runs. Writing the
error first opens a window in which the pool believes the port is free and the
registry believes it is taken — so the **next** agent to ask for it is refused
with `core: tcp port N already serves "<the tunnel that failed>"`, a port that
is in fact available.

Found as a 2-in-120 flake in B-5 under `go test -cover ./...`, and confirmed
deterministically rather than left as a guess: with B-5 no longer waiting out
its read deadline before the second handshake, the case fails **15 times in
15** against `01ef62b`. The wait was masking the window, not the window
proving elusive.

The order becomes **unregister, then tell** on both paths. This is §1's rule
applied to the failure branch: when the client learns the outcome, the state
behind it is already true.

**Amendment 2, 2026-08-31, before merge**, recorded here rather than folded
silently into the build for the same reason as §6.1–6.3.

# BACKLOG — what gets built next, in order

**Owned by the product-manager role**
([`pumasi-ops/roles/product-manager.md`](https://github.com/pumasi-ai/pumasi-ops/blob/main/roles/product-manager.md),
duty 5). Seeded 2026-08-30 from candidate `0011-developer-tunnels.md`; first
evaluation pass 2026-08-31 at `3652e15`; post-release evaluation 2026-08-31 at
`83fd9f7`; post-`0041` evaluation 2026-08-31 at `1d9505c`; post-`0047`
evaluation 2026-08-31 at `b3d251d`; post-`0060`/`0066` evaluation 2026-09-01 at
`87244af`; post-`0081` evaluation 2026-09-01 at `9e2de66`, after the
names-with-owners merge — which delivered *half* of what was then item 2;
**post-`9cc9e65` evaluation 2026-09-01 at `9cc9e65`** (job `0093`), after the
other half landed by a hand run outside the queue while this file still named
it as the next packet — see *What changed in this pass*.

One list, features and bugs together — a priority that cannot compare them is
not a priority. Every entry points at its source and carries one line of
why-here. **The top of this file is what the project manager's next coder
packet builds** — except where an entry says in its own text that it is
operator action rather than a build; the packet then takes the highest entry
that *is* a build, and the operator item keeps its rank rather than being
demoted for being unbuildable.

> **Highest *build* entry: item 2** — *the cross-family code review that
> `9cc9e65` never got, and whatever it objects to.* That is the next coder
> packet, and it is **not new work**: `spec/0004-names-with-owners` **slice 2**,
> which this blockquote named as the next packet at `41f373c`, is **built and on
> `origin/main`** — `9dd067a` (the spec's §14, before any code) and `9cc9e65`
> (the code) — and is ticked under *Delivered* with what this seat measured.
> What it did not get is the CHARTER Part 3 requirement 3 review: `git show
> --stat 9cc9e65` lists two spec-round transcripts and **no code-round
> transcript**, `reviews/` holds nothing dated after 06:48 UTC, and the commit's
> own message carries **no test count and no `GATE:` line**. Item 1 outranks
> item 2 and is **not** a build: it is operator action blocked on **Q-014** —
> whose written retirement condition is now **met on `main`** and unmet on the
> host, and met on `main` only when the operator sets a flag (item 1(i)).
>
> **Three things that packet must not assume**, each measured by this seat at
> `9cc9e65` rather than argued: **(a)** that the builder's evidence is already
> on the record — the two transcripts it carries are *spec* reviews of §14, and
> one of them, `reviews/20260901-014800-spec-kimi.md`, is **304 bytes with no
> verdict and no error line**, committed four minutes after its own timestamp;
> **(b)** that the review is advisory here and therefore optional — it is
> advisory pre-`launched` (Part 0), **but** this repository has **no
> `RISK_ZONES.yaml`**, CHARTER Part 3 classifies an unmapped path as *can hurt
> someone* by default, and a can-hurt release needs **one approving review from
> a family other than the builder's** before its note and its window — so the
> review is the precondition of the release, not a courtesy; **(c)** that slice
> 2 changes how long a restart takes — it does not, its own commit says so, and
> what it removes is the **loss**.

Reordering is a commit with the reasoning in the message; the steward vetoes by
reverting.

**This list is renumbered at every pass, so cite it by number *and* title.**
The convention was introduced at `1853218` after three citations in
[`VALUE.md`](VALUE.md) pointed at the wrong entry — a bare "item 9" is a pointer
into an ordering this file changes on purpose. Anything outside this file that
names an entry should name its title beside the number, so that a renumber
breaks the number and not the meaning.

**It held again, and the check was cheaper this time because nothing is
renumbered.** All `BACKLOG.md` citations in [`VALUE.md`](VALUE.md) and
[`STAGE.md`](STAGE.md) were read against this file as it stood at `9cc9e65`,
*before* the edit below, and every number pointed at the entry it named.
**Every citation of item 2 by *title* is nevertheless re-pointed in this
commit**, because item 2 keeps its number and changes its substance: what stood
there — *the relay-restart half — a reservation that outlives the process* — is
delivered and moves to *Delivered*, and what stands there now is the review it
never got. A number that survives while its title changes is the failure the
convention at `1853218` was written for, arriving from the other side.

**What changed in this pass, and why.** At `41f373c` this file named
`spec/0004` **slice 2** as the next coder packet. The project manager queued it
as job `0089`; the dispatcher recorded `FAILED: exited with code 1` at 06:39
UTC with a one-line log — *"You've hit your weekly limit"* — and **nothing ran
under the queue**. Six minutes later the work landed anyway, in two commits
trailered `Agent: claude-code · Model: claude-opus-5[1m] · Sponsor: mok`: a
hand run by the steward, outside the dispatcher, on the packet the dispatcher
had just failed. The same run pushed `main`, which had been ten commits ahead of
`origin/main`, so the whole ownership-and-durability capability is public for
the first time. **None of that is a seat's decision, and this file records it
as the steward's act.** What *is* this seat's is the register: at `9cc9e65` the
blockquote above still named a delivered item as the next packet, item 2 said
*"specified and not built"*, and [`STAGE.md`](STAGE.md) §1 said a reservation
does not survive a restart on `main`. This pass strikes each, ticks the
delivered half with both SHAs, re-measures the tree, and ranks what the hand
run left out.

| | before (`41f373c`) | after (this pass, `9cc9e65`) |
|---|---|---|
| 1 | TLS / deploy — operator action, Q-014 | **1** — unchanged rank; Q-014's retirement condition is now met on `main`; the keepalive precondition got sharper |
| 2 | the relay-restart half, `spec/0004` slice 2 | **delivered** `9dd067a` · `9cc9e65` → *Delivered*. **2** is now the **code review it never got**, plus the release classification of both slices |
| 3–5 | bindable port · console `ssh -R` · ssh path cannot hold a name | **3–5** — unchanged |
| 6, 8 | PR-1, PR-2 — cited from an unmerged branch | **6, 8** — unchanged rank; now cite `PRODUCT-RULES.md` **v1.0 on `pumasi` `main`** (`23bbc64`) |
| 7, 9–12 | `agent/` tests · frozen case · BaseDomain · TUI · inspector | unchanged |

Four things follow:

- **The residual of the residual is a review, and it ranks above every build
  for a reason the charter states rather than one this seat argues.** Part 3
  requires a cross-family code review for every merge; Part 0 makes it advisory
  below `launched`, so `9cc9e65` is a legal merge and this file does not say
  otherwise. But this repository has no `RISK_ZONES.yaml`, Part 3 says an
  unmapped path *"defaults to can hurt someone"*, and a can-hurt **release**
  needs one approving non-builder family *before* its published note and 7-day
  window. Slice 2 cannot be released under the charter until somebody other
  than its builder has read it. That is item 2.
- **The delivered half is measured, not carried.** D-1 exists as a Go test and
  is a second `relay.New` over the same path (`relay/reservation_test.go:640`,
  with the comment above it saying why a reconnect would not be this case);
  D-2..D-7 each return **1** from `grep -rn "func <name>(" --include=*_test.go
  .`; `LastSeen` is `core/reservation.go:68`; `ReservationTTL` is 30 days and
  `MaxReservations` 10,000 (`core/reservationstore.go:52`, `:63`); the twelfth
  flag is `cmd/pumasi-relay/main.go:46`; `f.Sync()` and the directory `Sync()`
  are `core/reservationstore.go:290` and `:318`; `flock` is `:123`; the
  corrupt-file rename is `:236`. **What this seat could not verify** is the
  commit message's account of D-1 going red twice — before the store, and again
  after the name half alone because the port pool was built empty beside the
  durable set. That is the builder's report of a mutation; it is quoted under
  *Delivered* as a report, and item 2 is where it gets checked.
- **Item 1's precondition got sharper, not weaker.** Before slice 2, a stranger
  who claimed `sshsteward` while the keepalive was down held it until the next
  restart. **After a deploy of slice 2 with `-reservations`, that stranger
  holds it for thirty idle days across every restart** — and
  `pumasi-ops/tools/pumasi-tunnel-keepalive.sh` still passes no `--token`
  (`grep -c token` → **0**, re-read this pass). Durability makes the cost of
  forgetting the token durable too; it has to land before the deploy.
- **The list is not renumbered.** No entry moved rank; one changed substance
  under the same number, and every cross-file citation of it by title is
  re-pointed in this commit. The suite figures in [`STAGE.md`](STAGE.md) §2
  were re-taken by this seat at `9cc9e65` with the run count beside each; the
  builder's commit published none, so there was nothing to decline to carry.

---

## The order

**1 · Nothing serves TLS, and the relay that is actually running still says it
does** — *operator action and a blocked deploy, not a build.* Source: the
`83fd9f7` evaluation, checking the live relay rather than the code path;
unchanged by `0041`, which touched no deployed binary.

The build half is done and the lie is gone from `main`; it is not gone from the
internet. Two things are outstanding, and they are different sizes:

- **(i) Deploy the merged fix.** `pumasi.link` runs a pre-`83fd9f7` binary and
  will announce `https://` until someone restarts the relay. **This is blocked
  on `pumasi/DECISIONS.md` Q-014**, which asks who may restart a host whose live
  tunnel is `sshsteward` → this machine's port 22 (`RESOURCES.md` §4). Q-014 is
  open and is **explicitly outside CHARTER Part 0's proceed-on-default rule**.
  Neither this seat nor a coder packet may take it, and this file does not ask
  for it. **Re-measured read-only at 18:51:06 UTC 2026-09-01, no host touched:**
  `http://pumasi.link/_pumasi/status` still reports
  `"url":"https://sshsteward.pumasi.link"`, `curl https://pumasi.link/` fails to
  connect on a refused 443, `http://pumasi.link/` answers `200`, and
  `"tcp_range"` is still `20000–20099`. **`"count":1`** — the unattributed
  second tunnel of the last two passes is gone (*Not on this list*), and the
  live set is back to the *exactly one* that Q-014 was written on.

  **Q-014's own retirement condition is now met on `main`, and only there.**
  That entry's *What retires this entry* row names *"durable registry and port
  reservations"*. At `9cc9e65` a relay started with `-reservations <path>`
  rebuilds its reservation set **and the port pool's holds** from the file at
  boot (D-1, `relay/reservation_test.go:640`), so a restart no longer costs an
  owner their name or their port. Two things stop that from retiring the entry,
  and this file says both rather than one: the host does not have it; and **the
  condition is met only when the operator sets the flag** — an empty
  `-reservations` is today's relay in every respect (D-7), so a deploy that
  does not carry the flag and a path on the host delivers nothing on this
  point. Retiring the entry is the steward's act; this pass added the evidence
  to it and set nothing.

  **The undeployed set is now nine commits touching the Go tree, seven of them
  changing non-test code, delivering five distinct capabilities.** Counted on
  the same basis as before — `git rev-list 83fd9f7~1..9cc9e65` filtered on
  `*.go` — the eight of the last pass plus `9cc9e65` (durability). `e40a224`
  and `b3d251d` still touch tests only. Verified this pass, not carried.

  **The precondition on this deploy that was filed at `9e2de66` is sharper at
  `9cc9e65`.** `pumasi-ops/tools/pumasi-tunnel-keepalive.sh` passes **no
  `--token`** (`grep -c token` → **0**, re-read this pass), so after a deploy
  the steward's own `sshsteward` tunnel would be *unclaimed*. Before slice 2 a
  stranger who claimed it in the reclaim window held it until the next restart;
  **with `-reservations` set, they hold it for thirty idle days across every
  restart**. Adding a token is harmless against the binary running today, which
  ignores the field. **It must land before the deploy, not after**, and it is
  `pumasi-ops`'s file — filed in this pass's return block again rather than
  reached for from here. `0084`'s implementation note still holds: the script's
  health check is `pgrep -f "pumasi-tunnel --relay $RELAY .* --tcp "`, so
  `--token` goes between `--subdomain` and `--tcp`, never after `--tcp`.
- **(ii) Put a wildcard certificate for `*.pumasi.link` in front of the relay
  on the Vultr host.** This is the actual TLS gap: with (i) done, every tunnel
  is *honestly* plaintext, which is still a product that no `https://`-only
  webhook sender can be pointed at. TLS termination is deliberately outside the
  relay (`cmd/pumasi-relay/main.go` header), and outside it there is still
  nothing. `RESOURCES.md` §2 warns that proxying these records through
  Cloudflare would break raw TCP, so that is not the shortcut it looks like.

Why here: it is the largest single gap between what this product is and what a
stranger could use, it is the one item on this list every visitor to
`pumasi.link` meets today, and it is not demoted for being unbuildable.

**2 · `9cc9e65` has no cross-family code review — take one, act on the
objections, and classify both slices of `spec/0004` for release** — **the
next coder packet.** Source: CHARTER Part 3 requirement 3, read against `git
show --stat 9cc9e65`; and Part 3's default classification for an unmapped
path, below. This entry replaces *the relay-restart half — a reservation that
outlives the process*, which is delivered (*Delivered*, first entry).

**What is missing — verified in the tree by this seat at `9cc9e65`:**

- **No code-round transcript.** `9cc9e65` commits two files under `reviews/`,
  both `spec-*`: `20260901-014558-spec-gemini.md` (80 lines, `VERDICT:
  APPROVE`, on §14) and `20260901-014800-spec-kimi.md`. Nothing under
  `reviews/` is dated after 06:48 UTC. The commit's trailers are `Agent`,
  `Model`, `Sponsor`, `Token-Cost`, `Spec` — and **no `Reviewed-By`**.
- **The kimi transcript is not a review.** It is **304 bytes**: the header
  `review.sh` writes, a rule, and nothing — no verdict, no objection, and not
  even the `curl: (28)` line that the other kimi timeouts in this directory
  carry. Its timestamp is 06:48:00 UTC and the commit containing it is
  06:52:24 UTC, so a 600 s timeout cannot have elapsed: it was committed
  before the reviewer answered. One family approved the spec amendment, and
  that is the whole review record for **1,089 lines** of change to `core/`
  and `relay/`.
- **No measurement in the record.** The commit message carries no test count
  and no `GATE:` line — grepped for both by this seat. The first figures for
  this tree are the ones in [`STAGE.md`](STAGE.md) §2, taken by this pass.
- **A mutation claim nobody but the builder has seen.** The message says D-1
  was red before the store existed and **red again after the name half
  alone**, because the port pool was built empty beside the durable set — a
  defect §14 did not name. If true it is the most important sentence in the
  commit, and it is exactly the kind of claim slice 1's review approved and
  was wrong about three times (*Delivered*, slice 1). Nobody has checked it.

**Why it is the top build entry — the charter's reason, then this
repository's.** Part 0 makes the review advisory below `launched`, so the
merge is legal and this file does not say otherwise. But **this repository
has no `RISK_ZONES.yaml`** (`find . -name RISK_ZONES.yaml` → nothing,
re-checked at `9cc9e65`), and CHARTER Part 3 says the classification
*"defaults to can hurt someone when unmapped or unclear"*. A can-hurt release
requires *"one approving review from a model family other than the
builder's"* and then a published note with a 7-day window before a staged
release. So until a non-builder family has read `9cc9e65`, slice 2 **cannot
be released under the charter** even on the day Q-014 resolves — the review
is on the deploy's critical path, and it is the only thing on that path a
coder can take. The repository's own reason is *Delivered*: slice 1 was merged
green with an approving review and had **three** defects found afterwards, so
a slice merged green with no review at all is the weakest-evidenced change on
this tree — and it is the one that writes to disk.

**The classification, stated as this seat's reading and not as a decision.**
**Q-034**'s default answers the general question: an *ordinary* merge owes no
note; a *can-hurt* release owes a published one with a window. Slice 2 is
can-hurt on two independent grounds — the default above, because nothing maps
it ordinary; and on substance, because `core/reservationstore.go` **writes
token digests to disk** (mode `0600`, and §14.1 says in its own words that a
digest *"is still the input to an offline attack"*), and because the file it
writes decides who is refused a name for the next thirty days. Slice 1 has the
argument `0084` made and nobody answered — a bearer secret on a plaintext
wire; requests that were accepted now refused. **Classification and the note
are the merging builder's under CHARTER §2.1, and `pumasi/releases/` is not on
this role's *May write* list** — Q-034's own text says so of this role — so
both are routed to the packet and not written here. The nearest precedent is
`83fd9f7`, classed can-hurt and noted under **Q-020**. One consequence worth
stating: the 7-day window can run **while Q-014 is still open**, so writing
the note is not gated on the deploy and should not wait for it.

**What the packet must not do.** It must not amend §14 or any frozen case
except by CHARTER §3 requirement 2's remedy (**Q-030** is open on exactly
that, with four instances). It must not fold in slice 3 (item 5). It must not
deploy. And it must publish **every** transcript including the blanks — the
304-byte file above is why. **Fixer: the coder**, with `tools/review.sh code
9e2de66..9cc9e65` on at least one family that is not the builder's, planned
around **D-104**'s bundle-size finding: the range is 2,710 insertions across
13 files, and job `0080` measured qwen timing out above ~150 KB. Address cited
objections; discard uncited ones in writing (DRIVER.md §4).

**3 · A public port the pool believes is free may not be bindable, and the
relay gives up instead of taking the next one** — source: this evaluation,
found while establishing why the previous order's item 2 failed. It is the
product-side half of that port-range defect — whose test-side half is now split
between *Delivered* (`b3d251d`) and item 9 — and it is a different defect from
both. **Re-verified at `9e2de66`.**

**One sentence this entry carried is now false and is corrected rather than
re-labelled.** It read *"nothing since `0047` changed `core/portpool.go` or
`relay/tcp.go`"*. Slice 1 changed **both** — `core/portpool.go` **+66** for the
third `held` state and `relay/tcp.go` **+20** — and the defect below is
unchanged by either, which this seat established by reading the changed files
rather than by assuming. `listenTCP` still hands the port back and returns an
error (`relay/tcp.go:68`–`:75`), and `ServeAgent` still treats that as fatal to
the handshake.

`core.PortPool` is explicit that it *"does no I/O — it decides which number to
use; binding the listener is the relay's job"* (`core/portpool.go:21–22`). So
`Allocate` can hand back a port that the OS will refuse. When it does,
`relay.listenTCP` releases the port and returns an error (`relay/tcp.go:68`,
`:74` — re-read at `9e2de66`, moved by slice 1), and `ServeAgent` treats that
error as fatal to the handshake (`relay/relay.go:186`–`:200`) — it unregisters
the agent, discards any claim the handshake made, and refuses the tunnel.
**It never asks the pool for the next free port**, even though the other 99 in
the range are available. Reproduced this pass: with one port of a 100-port
range held by an unrelated process, every single tunnel request was refused.

There is also no guard on the operator's side: `-tcp-low` / `-tcp-high`
(`cmd/pumasi-relay/main.go:39–40`) accept any range, including one wholly inside
the host's ephemeral range, and neither `core.NewPortPool` nor `relay.New` warns.

**Two constraints slice 1 added that the fixer must not trip over, both new
since this entry was last written.** First, the pool's `owner` string is now the
**subdomain** and no longer the agent id — an agent id is minted fresh per
connection, so a hold keyed by one could never survive a reconnect
(`spec/0004` §5.2). Second, a bind failure now also runs `discardClaim`, and
frozen cases **C-7** and **C-9** pin what it must leave behind: a name claimed by
a handshake that never opened must be released, *and* its pool hold with it.
**A retry loop that walks to the next port must not discard the claim on the way
past** — the claim is only undone when the handshake as a whole gives up.

**What bounds this today, measured rather than assumed.** The running relay's
range is **20000–20099** (`http://pumasi.link/_pumasi/status`, `"tcp_range"`,
re-read 2026-09-01 02:48 UTC and unchanged), which is *below* the ephemeral
floor. So the
kernel cannot steal a port from the live relay, and the exposure there is only
another process on the Vultr host binding into that block. The unbounded case
is the operator who picks a range above 32768 with nothing to warn them, and
the suite, whose remaining half is item 9 — *a frozen acceptance case still
draws its port range from inside the kernel's ephemeral range*.

Why here: the failure is honest and self-healing since `1d9505c` — the agent is
refused rather than lied to, and it retries — so this is robustness, not a
falsehood on shipped surface, which is why it sits below item 2 rather than
above it. But it is the reason a 100-port pool can be defeated by one busy port,
and the range guard is a few lines that would have made the previous order's
item 2 impossible to write. **Fixer: the coder.**

**4 · The console never offers the zero-install `ssh -R` command** — source:
`VALUE.md` claim 1 against the live page. The ingress works: a banner grab on
`pumasi.link:2222` returns `SSH-2.0-pumasi-tunnel`. But `relay/dashboard.html`'s
command builder emits only `pumasi-tunnel --relay …` and its "First time here"
panel offers only `git clone && go build` — so the product's headline claim, the
one thing it does that needs nothing installed, is absent from the one page a
visitor sees.
**Re-verified at `9e2de66`:** `relay/dashboard.html` contains **0** occurrences
of `ssh -R` and **1** of `git clone`, and `pumasi.link:2222` still answered
`SSH-2.0-pumasi-tunnel` when this seat grabbed the banner at 05:51 UTC.
Why here: it is an afternoon's work on a page that is already live, and it
converts the strongest differentiator from a README sentence into the thing you
can paste. `MARKET.md` §2 makes that differentiator explicit and cited, which
raises what the omission costs.

**Read item 5 before taking this one, because the two now compound.** Putting
`ssh -R` on the console sends more users down the one path that, since slice 1,
can be *refused* a name and can never *hold* one. Neither entry blocks the
other, and this one is still worth doing first — but a console that offers the
zero-install command while saying nothing about which addresses it cannot keep
would be selling the narrower half of the product without the caveat.

**5 · The zero-install `ssh -R` path can be refused a name and can never hold
one** — *new this pass, and it is a gap slice 1 opened rather than one it
found.* Source: `spec/0004-names-with-owners/SPEC.md` §8, which states it and
hands it up rather than leaving it to be inferred, and §7's **slice 3**, which
names the shape and stops. **Re-verified in the tree at `9e2de66`:**
`relay/sshingress.go:167`–`:171` builds its `core.AuthRequest` with
`Subdomain`, `TCP` and `ClientVersion` and **no `Token` field**, with a comment
saying why; and `parseSSHUser`'s `+` grammar cannot carry one, because a token is
itself a valid subdomain and `name+token` would be read as a name.

**What actually changed for a user of that path, stated as the before and the
after rather than as a principle.** Before slice 1, nobody could hold a name, so
an ssh client lost nothing by being unable to. After slice 1, the CLI path can
claim a name and its public port with `--token`; the ssh path can do neither —
and `authorize` now calls `Check` on its tokenless request, so an ssh client
asking for a name **somebody else has claimed** is refused `ErrNameReserved`
where it would previously have been given the name. That is **strictly more
refusals and no new capability**, on the path this product leads with. Frozen
case **C-5** pins the refusal and it is correct; nothing here says the behaviour
is wrong.

**Why here — fifth, above PR-1 and below the console entry.** It is the only
entry on this list that got *worse* as a direct result of what merged this pass,
and it lands on `VALUE.md` claim 1 and `MARKET.md` §2's cited differentiator
rather than on an internal surface. It is **not** ranked higher, for a reason
the spec gives and this seat agrees with: an ssh tunnel is session-scoped by
nature — a person is sitting at a terminal holding it open — so the address it
cannot keep is one it was less likely to need. And it is **not** merely a
footnote, because item 4 exists: the moment the console offers the command, this
is the experience it offers.

**Fixer: the coder, after an intent statement — not straight to a packet.**
`spec/0004` §7 is explicit that slice 3 *"needs its own intent statement, because
both halves change what the zero-install path promises"*, and the second half —
an operator-seeded reservation for a name nobody may claim first — is the
operator's answer to §8's trust-on-first-use exposure and is a policy decision
before it is code. **Slice 2 landed at `9cc9e65` and did not fold it in** —
verified: the range `9e2de66..9cc9e65` touches neither `relay/sshingress.go`
nor the `+` grammar, so this entry is exactly as it was. It is still a grammar
plus a policy, and it still needs the intent statement first.

**6 · PR-1 compliance: a version that moves and is user-visible** — source:
[`PRODUCT-RULES.md` PR-1](https://github.com/pumasi-ai/pumasi/blob/main/PRODUCT-RULES.md)
(**v1.0, 2026-08-30, on `pumasi` `main` since `23bbc64`**, published 15:42
UTC today and read fresh from `main` this pass — the first of eleven
evaluations not to read it off the unmerged `worktree-product-rules` branch;
**binds always, from the first commit**, in four clauses as v1.0 states them:
one source of truth in the repository root `package.json` *"or the ecosystem
equivalent"*; semantic versioning that still moves below 1.0.0; a version a
person can find without reading source; and the version stated in every
feedback report and every release note). This product has no version
anywhere: no version file, no `/version`, nothing in the console footer,
nothing in a release note — and the one tunnel note that exists names a
commit instead.
**Re-verified at `9cc9e65`**, with one detail that makes the fix smaller than it
looks: **the repository root already has a `package.json`** — added so
`pumasi/tools/gate.sh` step 1 finds a suite — and it still has **no `version`
field**, which is exactly where PR-1 says the one source of truth belongs.
`core.AuthRequest.ClientVersion` exists as a field (`core/handshake.go:33`) and
the only thing that ever sets it is `relay/sshingress.go:169`, which fills it
with the *ssh client's* version string, not this product's. (That line was
`:165` last pass; slice 1 moved it, the fact did not.)
**Why here, and it moved up one this pass on measured grounds rather than on a
re-reading.** There are now **nine** merges touching the Go tree that the host
does not have — **seven** of which change non-test code, delivering a **fifth**
distinct capability (item 1) — and **nothing on the console, in
`/_pumasi/status` or in the logs says which build is answering**. That is
`pumasi-booking`'s Q-012 problem arriving early, and here the answer is a few
lines of Go plus a field already in the wire protocol.

**The sixth consecutive evaluation had to infer the build from an accident,
and the merge this pass records adds a capability with no discriminator at
all.** Until the last pass the only discriminator was the scheme in the `url`
field, which separates pre-`83fd9f7` from post- and nothing else. Slice 1 added
a `"reserved"` key to every tunnel in the status view (`relay/dashboard.go:42`,
a plain `bool` with no `omitempty`, so it is always present) — and the live
read at **18:51:06 UTC** carries **no `reserved` key on the one tunnel it
reports**, which is how this seat established the host still predates
`4489fbe` without touching it. **Slice 2 changes no external surface**:
`9cc9e65` does not touch `relay/dashboard.go` or `dashboard.html`, so a relay
running with `-reservations` is indistinguishable from one without it from
anything a visitor, an operator or a deployer can read. **Two accidental
discriminators is not a version**, and four of the seven non-test commits —
slice 1's three and slice 2's one — are mutually invisible from outside. It
stays above *`agent/` has no tests* because PR-1 **binds always**, the fix is
small, and the entry below it lost the deadline that put it there.

**7 · `agent/` has no tests** — source: the gate reading; L-006.
`go test ./...` reports **no test files** for `agent`, `cmd/pumasi-relay` and
`cmd/pumasi-tunnel`. The two `cmd` packages are argument parsing and can wait.
`agent/` cannot: it is half of every tunnel — handshake, reconnect, local dial,
stream fan-out — and today the only thing exercising it is `relay`'s
end-to-end tests, which use it as a fixture and assert on the relay's behaviour,
not its. Reconnect and local-dial-failure have no case at all.
**Re-verified at `9e2de66` and again at `9cc9e65`:** the three packages still report *no test files*,
and the `-cover` runs of this pass put them at **0.0%** ([`STAGE.md`](STAGE.md)
§2 carries the figures and their run counts).

**This entry's why-here has been overtaken and moves down one rather than being
restated.** It read: *"item 2 — the durability work — will rewrite reconnect
behaviour, so the tests should exist before that and not after."* **The
reconnect behaviour was rewritten this pass and the tests did not exist.** Slice
1 is precisely a change to what happens across a reconnect, it landed at
`4489fbe`, and `agent/` gained nothing — of the **989** lines of new test in
that merge, **0** are in `agent/` (`core/reservation_test.go` 221,
`core/portpool_test.go` 89, `relay/reservation_test.go` 618,
`relay/discardclaim_test.go` 61). That is the second consecutive merge to make
the same point: `fd523e8` put all 468 of its new test lines in
`relay/sessionorder_test.go`. **And slice 2, landed at `9cc9e65`, touched
neither reconnect nor `agent/`** — all **476** of its new test lines are in
`core/reservationstore_test.go` (357) and `relay/reservation_test.go` (119),
per `git diff --stat 9e2de66..9cc9e65 -- '*_test.go'` — so the deadline this
entry set for itself has passed twice over and no new one replaces it.
Why here, now: the gap is unchanged and real — `agent/` is half of every tunnel
and reconnect and local-dial-failure still have no case — but the *argument* for
its position was a deadline, and the deadline is spent. It sits below item 6,
which binds always under a rule and is a few lines, and above everything whose
cost is smaller.

**8 · PR-2 compliance: in-app feedback** — source:
[`PRODUCT-RULES.md` PR-2](https://github.com/pumasi-ai/pumasi/blob/main/PRODUCT-RULES.md)
(**v1.0, 2026-08-30, on `pumasi` `main` since `23bbc64`**, read fresh from
`main` this pass; **binds at the `beta` promotion** — below `beta`,
encouraged). As v1.0 states it: three kinds (bug · feature request · general)
from inside the product with no tracker account; a report lands **in the
open** as a public GitHub issue in this repository labelled `feedback`;
diagnostics sanitized before they leave; contact optional; the report
inspectable before it is sent. Nothing in the product collects feedback. The
reference behaviour is `pumasi-booking`'s (`service/src/feedback.ts`) —
matched in behaviour, not copied.
product collects feedback. The reference behaviour is `pumasi-booking`'s
(`service/src/feedback.ts`) — matched in behaviour, not copied.
**Re-verified at `9cc9e65`:** nothing in `core/`, `relay/`, `cmd/` or `web/`
collects a report, and the console is still the routing table plus a command
builder.
Why here: it **gates the `beta` promotion**, so it must be built before the
label moves, and it is worth little before item 1 puts the merged product in
front of a stranger. The natural home is the console, where a visitor already is.
**Unchanged in rank by this pass** — it moves from 7 to 8 only because a new
entry was inserted above it.

**9 · A frozen acceptance case still draws its port range from inside the
kernel's ephemeral range — real, latent, and blocked on a governance reading
rather than on work** — source: the residual half of the previous order's item
2, which `b3d251d` did not land.

`relay/scheme_test.go:314`–`:315` configures `TCPPortLow: 34500`,
`TCPPortHigh: 34599` — **the same two line numbers as last pass, re-read at
`9e2de66`**. This host's `/proc/sys/net/ipv4/ip_local_port_range` is
**32768 60999**, re-read this pass and unchanged, so that block sits inside the
range the kernel hands out as source ports for outgoing connections — the
identical defect that `b3d251d` fixed in `relay/tcp_test.go`.

**Why it did not land with the rest.** Those two lines are inside **A-10**,
`TestSchemeChangesNothingButTheScheme` (`relay/scheme_test.go:295`), a frozen
acceptance case of `spec/0001-public-scheme`. Job `0047` moved it using CHARTER
§3 requirement 2's own named remedy — *"If the tests are wrong, amend the spec in
the open and take a fresh cross-family spec review"* — wrote the amendment as
SPEC 0001 §7, and got a fresh gemini spec review that approved it
(`reviews/20260831-161500-spec-gemini.md`). The **code** review then objected,
citing that same clause, on the ground that the *builder* authored the
amendment. A cited objection governs and may not be argued past (CHARTER Part
3), so the builder reverted the whole half — the A-10 fixture, SPEC 0001 §7, the
`CASES.md` note and SPEC 0002 §6.6 — rather than proceed. The reviewers
contradicted each other and each switched sides on the same clause across the
two ranges; all four transcripts are committed. **The reading is open and is
escalated, not decided here** — see the digest entry for this pass and
[`STAGE.md`](STAGE.md) §3.1.

**Why it ranks ninth, below item 2 and below everything buildable.** Two
reasons, and the first is measured rather than argued:

- **Its cost today is nothing, and this pass re-measured that rather than
  carrying `0047`'s figure.** A-10 calls `relay.New` and reads
  `Registry().PublicScheme()` back. `relay.New` passes the range to
  `core.NewPortPool` (`relay/relay.go:135`, moved from `:124` by slice 1), which
  is explicit that it *"does no I/O"* (`core/portpool.go:21`–`:22`). **A-10 binds
  nothing**, and this pass
  re-read the two constants at `9e2de66` to confirm they are still 34500 and
  34599. The block stayed contended on this host through the runs recorded in
  [`STAGE.md`](STAGE.md) §2 and the suite did not notice. A latent defect that a
  contended range cannot provoke is a latent defect.
- **Nothing a coder packet could take sits under it.** The work is one constant;
  what blocks it is whether §3 requirement 2's remedy is available to a builder
  at all, given that Part 3 requirement 1 has the builder authoring every spec in
  the first place. Ranking it above a buildable item would point the next packet
  at something that cannot be built without answering that first.

**Q-030 gained a fourth instance on 2026-09-01 (`pumasi` `196b749`) and none of
the four is this repository's newest work.** `spec/0004` amended no frozen case
from `spec/0001`–`0003` — §12 says so and this seat confirmed it: the range
`1853218..9e2de66` touches no test file outside `core/portpool_test.go`,
`core/reservation_test.go`, `relay/reservation_test.go` and
`relay/discardclaim_test.go`, all of them this spec's own. **Slice 2 kept it
that way, verified at `9cc9e65`:** `9dd067a` adds D-2..D-7 in a new §14.9 and
`git show 9dd067a -- acceptance/CASES.md` changes **no existing row**, D-1
included; `9cc9e65` touches no test file outside `core/reservationstore_test.go`
(new) and `relay/reservation_test.go`. **So the count of
instances rose and this product's exposure did not.** The reading is still open;
[`STAGE.md`](STAGE.md) §8 carries it and nothing here closes, dates or softens
it.

Why here rather than lower: it is a real defect in a frozen file, and the day
A-10 gains an assertion that binds — or `relay.New` gains a bind — it stops being
latent silently. It sits beside item 10 — *`relay.New` accepts a `BaseDomain`
that `core.NewRegistry` silently normalises differently* — the other entry on
this list that is real, cheap and has no live consequence. **Fixer: the coder, once the reading
resolves**; `SPEC 0002` §6.5 is the precedent on the permissive side, where the
same builder amended a frozen fixture in the open and it stood.

**10 · `relay.New` accepts a `BaseDomain` that `core.NewRegistry` silently
normalises differently** — source: `0041`, found while verifying SPEC 6.1 and
deliberately left out of its scope.

`relay.New` rejects only the empty string; `core.NewRegistry` trims and
lowercases. So `-domain " pumasi.link"` — a stray character in a systemd unit —
starts a relay whose every registry lookup is built from the untrimmed string
and never matches. `0041` recorded the boundary precisely, and it is narrower
than it first looks: **trailing space, uppercase, trailing dot and doubled dot
are all unaffected**; it is specifically the **leading** space. The condition is
now pinned by test as `lookupBreakingDomain` (`relay/bindorder_test.go:50–56`),
where it is used as a deliberate fault injection rather than fixed.
Why here, and honestly low: it has no live consequence today — the deployed
relay's domain is correct, the failure is loud and total rather than silent and
partial once you look at any tunnel, and the tests that would catch a regression
already exist. It is a validation asymmetry between two constructors, worth one
guard in `relay.New`, and it does not belong near the top of a list whose top
two items are a lie on the internet and the `beta` bar. **Re-verified at
`9e2de66`:** `lookupBreakingDomain` is still `relay/bindorder_test.go:56`, and
`relay.New` still tests `cfg.BaseDomain` only for the empty string — now at
`relay/relay.go:100` rather than `:93`–`:94`, slice 1 having moved it.
**One thing this pass adds:** `discardClaim` composes the lookup key the same
way (`subdomain + "." + r.cfg.BaseDomain`, `relay/relay.go:429`), so an
untrimmed `-domain` now silently defeats the rollback guard as well as every
registry lookup. That does not change the rank — the failure is still loud and
total the moment anyone opens a tunnel — but it is one more call site reading a
value nothing validated.

**11 · Client CLI: the interactive terminal UI** — source: the seeded item 5;
[`docs/ux/incumbent-ux-spec.md`](../docs/ux/incumbent-ux-spec.md) §273 for what
the incumbents' status line shows. `cmd/pumasi-tunnel` prints logs; there is no
live view of requests, response times or status codes, and no released binary
for macOS or Windows.
**Re-verified at `9e2de66`:** `cmd/pumasi-tunnel` is still flags and logs, and
`go test ./...` still reports *no test files* for it.
Why here: it is polish on a path that works, and every item above it is either
something untrue or something missing that a user would rely on. **See also
*Not on this list*** — a three-line change to this package's `--relay` flag was
found unattributed in the working tree, reasoned through, and stashed rather
than committed; if anyone wants it, it belongs here with its cost written down.

**12 · Local request inspector on `127.0.0.1:4040`** — source: the seeded item
6; [`VALUE.md`](VALUE.md) claim 5 (where it is marked *not built*). Re-verified
at `9e2de66`: `web/` still contains **0** entries; there is no listener, no SSE,
no replay.
Why here: last of the build items. It is a whole second product surface, and the
honest catalog page (`pumasi-web` `843bdef`) already tells visitors it does not
exist, so nobody is being misled while it waits.
**Corrected at `1853218`, and the correction is re-checked here rather than
carried:** this entry said *"`MARKET.md` §2 records that all three incumbents
ship one"*. It does not — `grep -ci inspector roadmap/MARKET.md`
returns **0** at `9e2de66`, as at every revision that file has ever had, and its
§2 table carries five rows, none of them an inspector. The claim is sound and its source is
[`docs/ux/incumbent-ux-spec.md`](../docs/ux/incumbent-ux-spec.md): the
comparison table at line 78 gives an Inspector row for all three, and §6.1, §6.4
and §6.5 give one section per vendor. `VALUE.md` claim 5 carried the same wrong
citation and is repaired in the same commit; both were written into this
repository on the same day, in `dda04c7`, the commit that created `MARKET.md`.
That makes it a real gap, not a whim, but not ahead of anything above.

---

## Delivered

Verified against the tree during this evaluation, not taken from commit
subjects.

- **A reservation outlives the relay process** — delivered **`9dd067a`** (the
  spec: §14 answers §7's three open questions and adds D-2..D-7, before any
  code existed) and **`9cc9e65`** (the code), spec
  [`spec/0004-names-with-owners`](../spec/0004-names-with-owners/SPEC.md)
  **slice 2 of two**. This was **item 2** of the previous order and the
  **residual** of the order before that. It landed by a hand run outside the
  queue — `pumasi-ops` job `0089` is marked `FAILED` and its work is on
  `origin/main`; the dispatcher never ran it — with trailers `Agent:
  claude-code · Model: claude-opus-5[1m] · Sponsor: mok`, **and without a
  cross-family code review**, which is item 2 above.

  **Verified in the tree by this seat at `9cc9e65`, by reading and by grep, and
  not taken from the commit message.**

  - **The case that did not exist now does, and it is the right shape.**
    `grep -rn "func TestAReservationOutlivesTheRelay(" --include=*_test.go .`
    returns **1** (`relay/reservation_test.go:640`) against **0** at `41f373c`.
    It claims `myapi` and a one-port range on relay 1, closes the agent, waits
    for the name to leave the registry, **`Close()`s the relay**, and builds a
    **second `relay.New` over the same path** — then asserts that an anonymous
    stranger and a token-bearing stranger are both refused the name, that a
    stranger asking for *any* port is told the pool is empty, and that the
    owner gets name and port back from a relay that never saw it connect. The
    comment above it says what would make it stop being D-1: reusing `r`.
    D-2..D-7 exist one function each (`core/reservationstore_test.go:50`,
    `:125`, `:172`, `:205`, `:271`; `relay/reservation_test.go:711`), and each
    returns **1** from the same grep.
  - **The store is what §14 specified.** `core/reservationstore.go` is new
    (328 lines): one JSON document, sorted by name, mode `0600`, written to a
    temp file in the same directory, `f.Sync()` before the rename (`:290`) and
    the directory `Sync()` after it (`:318`) — §14.1's four steps in §14.1's
    order; `flock(LOCK_EX|LOCK_NB)` on a sibling lock file held for the store's
    life (`:123`) — §14.3; a corrupt file renamed to `<path>.corrupt-<RFC3339>`
    (`:236`) and the relay starting empty — §14.2. `ReservationTTL` is **30
    days** (`:52`), swept at load; `MaxReservations` is **10,000** (`:63`) and
    refuses a new name, never an existing owner (`core/reservation.go:195`).
  - **`LastSeen` exists and is set by `Claim` only** (`core/reservation.go:68`,
    `:211`); `Check` still records nothing, so a stranger being refused a name
    cannot refresh its owner's clock. The comment at `:52` that used to say the
    field would arrive with slice 2 now says it did.
  - **The twelfth flag.** `cmd/pumasi-relay/main.go:46`, `-reservations`,
    default `""`. Still no auth flag; a token still proves continuity and not
    permission (§3).
  - **The port pool is replayed from the set at boot.** `relay/relay.go` (+106)
    re-holds every reservation's port after the load and logs rather than
    refuses when a held port is outside the operator's current range. This is
    the part §14 did not name and D-1 found (below).
  - **The range `9e2de66..9cc9e65` touches no test file outside the two this
    spec owns**, nothing under `agent/`, and not `relay/sshingress.go` or
    `relay/dashboard.go`; `git show 9dd067a -- acceptance/CASES.md` changes
    **no existing row** — 0 changed lines match `D-1 `. No frozen case was
    amended and **Q-030 is not reached.**

  **What it does not do. Every line is the spec's or the commit's own text.**

  - **Not a shorter outage.** The commit: *"This changes nothing about how long
    a restart takes. A live TCP connection cannot outlive the process at either
    end."* What it removes is the **loss** — an in-memory set was empty after a
    restart, so every claimed name was trust-on-first-use again for the reclaim
    window.
  - **Not on by default.** An empty `-reservations` is today's relay in every
    respect — no file, no lock, no sweep, no write — and D-7 exists to go red if
    that changes. A deploy that does not set the flag delivers none of this;
    item 1(i) says so.
  - **Not a retirement of Q-014.** The retirement condition is met **on
    `main`**; the host does not have it, and the steward closes windows. Item
    1(i).
  - **Not reviewed.** Item 2.
  - **Not a defence against a disk that lies.** §14.1's closing paragraph
    claims what POSIX gives and no more.

  **⚠ The builder's own report, quoted as a report because nobody has checked
  it — and item 2 is where it gets checked.** The commit message says D-1 went
  red on its own clause before the store existed (three failures, quoted in
  the message), and **red again after the name half alone**: *"a durable set
  said port 21200 was spoken for while a freshly built pool handed it to the
  next stranger who asked for 'any port'."* §14 did not name the port pool as
  something to rebuild from the set; only a case that asks about the port
  could have caught it, and D-1 asks. If that account holds, it is the third
  time on this spec that the frozen case and not the review found the defect —
  and it is L-010's shape again: a durable set beside an empty pool passes every
  case that asks about the *name*.

  **What it cost, on the record rather than in the message.** The commits say
  `Token-Cost: ~120k` for the spec and `~200k` for the code. What the record
  does not have is a test count, a `GATE:` line, or a second family's reading
  of the code. The first two are supplied by this pass
  ([`STAGE.md`](STAGE.md) §2); the third is item 2.
- **A name and its public TCP port belong to somebody, and survive a
  disconnect** — delivered **`4489fbe`** (the change), **`c12d11a`** and
  **`20e9d57`** (two more defects, found *after* the code review approved), spec
  [`spec/0004-names-with-owners`](../spec/0004-names-with-owners/SPEC.md),
  **slice 1 of two**. This was the **first half** of item 2 of the previous
  order. The second half — durability — is **delivered at `9cc9e65`, the entry
  above**.

  **Verified in the tree by this seat at `9e2de66`, by reading and by grep, and
  not taken from the build report.**

  - `core/reservation.go` is new (232 lines) and pure: `Claim` · `Check` ·
    `Discard` · `PortHolder` · `Get` over a `map[string]Reservation` under a
    `sync.RWMutex`, storing a hex `sha256` and never the token, compared with
    `crypto/subtle.ConstantTimeCompare`. `MinTokenLen` is **16**
    (`core/reservation.go:45`) and is enforced at `:102` rather than compensated
    for by a KDF — the reasoning is `SPEC.md` §3.1 and it is conditional on that
    constant, which the file says in as many words.
  - **The field this entry existed to name is now read.** At `1853218`,
    `grep -rn "\.Reserved" --include=*.go` returned **nothing at all**. At
    `9e2de66` it returns **two non-test reads** — `relay/dashboard.go:80`, the
    status surface, and `relay/relay.go:429`, the rollback guard — plus five in
    `relay/reservation_test.go`. The write that kept moving (line 236, then 249,
    274, 297) is gone: `Tunnel.Reserved` is now set from `Reservations.Get(name)`
    rather than from the shape of the request.
  - `core/portpool.go` gained a third state, **`held`** (`:41`), with `Hold` /
    `Unhold` / `Holder`; `Allocate` skips a port held by another owner (`:84`)
    and `AllocateSpecific` grants one held by its own (`:111`). The pool's
    `owner` string is now the **subdomain** and no longer the per-connection
    agent id — which is the only reason a hold can survive a reconnect at all.
  - **One check, in the one place both ingress paths already share.**
    `Relay.authorize` calls `Claim` for a request carrying a token of at least
    `MinTokenLen` and `Check` otherwise (`relay/relay.go:333`, `:341`), and both
    `ServeAgent` and `ServeSSH` run through it. That is
    [L-009](https://github.com/pumasi-ai/pumasi/blob/main/lessons/L-009-two-paths-one-claim.md)'s
    shape rather than two paths that happen to agree today.
  - **Nineteen of the spec's twenty acceptance cases exist as Go tests. The
    twentieth is slice 2's and does not.** Checked one function name at a time:
    R-1..R-8, P-1..P-2 and C-1..C-9 each return **1** from `grep -rn "func
    <name>(" --include=*_test.go .`; **`TestAReservationOutlivesTheRelay`**
    (D-1) returns **0** — *at `9e2de66`; at `9cc9e65` it returns 1 and so do
    D-2..D-7, see the entry above*. **989** lines of new test across four files, and — see
    item 7 — **none of them in `agent/`**.

  **What it does not do. Every line of this list is the spec's own text, not a
  qualification added here.**

  - **Not the restart half.** §7: *"The middle column of §4 is untouched and the
    relay-restart half of item 2 is **not** delivered."* *Delivered since, at
    `9cc9e65` — the entry above.*
  - **Not a retirement of Q-014.** §4: *"Saying slice 1 retires Q-014 would be
    false and this spec does not say it."* And nothing it does can reach the
    steward's route until it is deployed, which is the act Q-014 gates — item 2's
    closing paragraph names that circularity, and item 1(i) names the keepalive's
    missing token.
  - **Not a withdrawal from the anonymous case**, deliberately: `spec/0004` §6's
    third column keeps a tokenless request working on an unclaimed name, and
    frozen case **C-3** exists to go red if anyone ever removes that. So the
    unattributed `skk6g7tyrs` tunnel on the live relay is **not** a defect this
    closes — see *Not on this list*, where that sentence has been corrected.
  - **Not reachable from the zero-install path** — item 5.
  - **Not a defence of the token on the wire.** §3: it is a bearer secret on a
    plaintext connection, and what bounds that is item 1(ii), not this.

  **⚠ Three defects were found after a green suite *and* a clean code review,
  and the third was introduced by the fix for the second.** `SPEC.md` §11, §12,
  §13, and the reviews are committed. A claim could be destroyed by a race it
  lost; **R-8** could pass while its own clause was false, because its test
  supplied the `Discard` call the clause was about; and the guard added to fix
  the first suppressed the rollback whenever *anything* was registered on a name,
  so a token-holder arriving at a name an anonymous agent was already sitting on
  left a claim behind that no tunnel ever opened. Each was **a clause whose truth
  no execution could reach**. The transferable rule, §13, and the reason this
  entry carries it rather than filing it away: **a guard added to protect an
  invariant needs a case that fails when the guard is *wrong*, not only one that
  passes when it is right.** `SPEC.md` §10's mutation table is now complete —
  twelve mutations, every built case — and under the over-broad guard **C-8** is
  the only case in the whole suite that reddens. That is the measurement the
  first version of the guard did not have, and it is why a green suite is not
  evidence here.

  **What it cost, because the previous *Delivered* entry is about mis-sizing.**
  Eight commits where the shape suggested three, five spec-review rounds, one
  acceptance case narrowed and three added after the freeze, and one test built
  and thrown away — the concurrency case for §11's hazard produced **0 failures
  in 5 runs of 40 attempts and 0 in a single run of 1,500** with the guard
  removed, so it could not fail against the defect it named. This file predicted
  none of that, and it did not try to: the packet was told to write a spec rather
  than a patch, and the three defects above are the argument for having done so.

- **The HTTP path announces a URL only once the session that serves it exists**
  — delivered **`fd523e8`** (job `0060`), spec
  `spec/0003-session-before-announce`. This was **item 2 of the previous order**,
  and the entry at the top of this file went on naming it as the next coder
  packet for two commits afterwards.

  Verified in the tree at `87244af` by the seat that wrote this entry, and
  **re-read at `9e2de66` by this one** — the line numbers moved with slice 1 and
  the ordering did not. Reading `relay/relay.go` at `9e2de66`: the mux session is
  built, `r.mu` is taken, `r.sessions[resp.AgentID]` is installed and the auth
  response is written **inside that one critical section** (`:239`–`:241`), and
  the session is deleted again in the same section if the write fails (`:245`).
  `ServeHTTP` (`:450`) takes the same lock to read `r.sessions`, so a visitor who
  arrives in what used to be the window now waits for the lock instead of being
  answered `404 No tunnel is open`. Frozen cases: `relay/sessionorder_test.go` — C-1
  `TestVisitorIsNotAnsweredBeforeTheSessionExists`, C-2
  `TestTheAnnounceReachesTheAgentBeforeAnyStream`, C-3
  `TestAFailedAnnounceLeavesNothingBehind`.

  **⚠ This entry predicted the wrong fix, and that is the half worth keeping.**
  Item 2 said, in its own words: *"there is nothing between step 3 and step 4
  that needs the response to have been sent, so the insert can simply move above
  it."* Its confidence in that reading is why it was ranked as a **three-line
  reordering** — here, in [`STAGE.md`](STAGE.md) §4's cost-to-move, and in the
  evidence rows this project wrote under **Q-024**. **The bare reorder was
  built, and it is worse than the defect it was meant to repair.**

  The mechanism, re-read by this seat rather than taken from the build report.
  The announce is written **raw** — `r.writeFrame(conn, okFrame)`, which is a
  plain `w.Write` at what was then `relay/relay.go:327`–`:334` and is
  `:438`–`:445` at `9e2de66` — while `mux.Session.Open`
  writes a `FrameOpen` on **the same connection** under the session's own
  `writeMu` (`mux/session.go:88`, `:102`, `:181`). Two writers, one socket, no
  shared lock. So a visitor forwarded in the *new* window puts a stream frame on
  the wire ahead of the auth response; the agent is sitting in
  `core.DecodeFrame` waiting for exactly one frame, `DecodeAuthResponse` rejects
  what arrives, and the tunnel drops into reconnect backoff. Measured by `0060`
  against the bare reorder and recorded in `spec/0003/SPEC.md` §2 and §6: **C-1
  answers `502` where the unfixed tree answered `404` honestly, and C-2 waits
  10.76 s for an `OnConnect` that never fires.** A wrong answer about a tunnel
  that is about to work, in place of a right answer about one that is not open
  yet.

  **Why a future seat should care rather than file this as trivia.** The
  prediction was not careless — it was read off `spec/0002` §2, which states in
  its own text that the mux session cannot be created before the handshake
  response. That half of `spec/0002` is itself wrong, which is why `0060` wrote a
  new spec instead of folding a contradiction into a frozen one (L-007). And the
  reorder would have passed everything this repository had: **500 clean suite
  runs, the whole gate, and every frozen case that existed before
  `spec/0003`** — the figures this file published at `b3d251d`. The transferable
  rule is the one this entry did not follow: **a fix named in a backlog entry is
  a claim, not evidence, and ranking an item by that claim mis-sizes it.** The
  three lines this entry promised came to ten added and five removed, of which 26
  of 36 added lines are comment explaining why a lock is now held across a socket
  write.

  **What this does not do.** It does not retire **Q-024** — that is the
  steward's act, and it is not taken here; `STAGE.md` §2 records what this pass
  measured and stops. And it is the third of three merged behaviour changes still
  waiting behind item 1's undeployed restart, so **no user of `pumasi.link` has
  it.**

- **The test suite's TCP harness moves below the kernel's ephemeral floor** —
  delivered **`b3d251d`**. This was the *first half* of item 2 of the previous
  order; the second half is item 9 above and is **not** delivered.

  Verified in the tree: `relay/tcp_test.go:40` sets `tcpHarnessBase = 21000`,
  and `tcpHarnessPorts` hands each harness a block of ten from an
  `atomic.AddInt32` cursor — four harnesses, **21000–21039**, below the 32768
  floor and clear of `bindOrderBase`'s 20500–20559. That is `bindorder_test.go`'s
  existing scheme applied rather than a second one invented, and it is SPEC 0002
  §6.5's finding (each case gets its own block) carried across to the older
  harness. No product code changed.

  **Verified by measurement, this seat's own, at `b3d251d` on this host.** The
  host is the same one whose ephemeral range made the previous pass unmeasurable
  — `/proc/sys/net/ipv4/ip_local_port_range` is still **32768 60999**, `workerd`
  still held **127.0.0.1:34000** throughout, and `ss -tanp` found **12** sockets
  inside 34500–34599. The suite no longer cares:

  | Arm | Runs | Failures |
  | :--- | ---: | ---: |
  | `go test -count=1 ./...` (the gate's step 1, verbatim) | **500** | **0** |
  | `go test -count=1 -cover ./...` | **100** | **0** |
  | `tools/gate.sh`, whole gate, `SKIP_FAMILY_PROBE=1` | **40** | **40 × `GATE: PASS`** |

  Against **40 failures in 40** at `1d9505c` on this same machine, recorded
  above. `ss -tanp` found **0** sockets in 21000–21039 and **0** in 20500–20559,
  which is the property the change was for.

  **Coverage is measurable again, and it moved.** The previous pass could not
  take a figure for `relay` at all — four of its tests aborted on a port
  collision — and carried 74.7% as inherited. Re-measured here over the 100
  `-cover` runs: `core` **80.3%**, `mux` **83.5%**, `relay` **82.0%**. `agent`,
  `cmd/pumasi-relay` and `cmd/pumasi-tunnel` remain **0.0%** — item 7.

  **What this does not do:** it does not retire **Q-024** (that is the steward's,
  and `STAGE.md` §2 records the evidence without closing the window), it does not
  fix item 9, and it is one of **eight** merges waiting behind item 1's undeployed
  restart — though, being test-only, it is one of the two that would change
  nothing on the host even once deployed.

- **The public TCP address is bound before it is announced** — delivered
  **`1d9505c`** (`3480990` the relay change, `e40a224` the suite fixture),
  spec `spec/0002-bind-before-announce`. This was item 2 of the previous order.

  Verified in the tree: `relay/relay.go` now calls `r.listenTCP` **before**
  `core.EncodeAuthResponse` and `r.writeFrame`, and a bind failure unregisters
  the agent and answers the handshake with an error frame rather than a
  correction sent after the address left. `relay/tcp.go` separates `listenTCP`
  (bind) from `serveTCP` (forward) so the socket can be answering before a
  session exists, and a visitor who arrives in the gap waits in the accept
  queue. `relay/sshingress.go` carries the same ordering.

  **Verified by measurement, this seat's own, not the fix's report.** The
  ordering defect was isolated from the host by running the same unmodified
  `relay` and `agent` packages on a port range **outside** the ephemeral range
  (14000–14099, confirmed wholly bindable), taking the announced `TCPAddr` and
  dialling it the instant it arrived:

  | | `83fd9f7` (before) | `1d9505c` (after) |
  | :--- | :--- | :--- |
  | 200 iterations | 3 dial refusals | **0** |
  | 500 iterations | 5 dial refusals | **0** |
  | 2000 iterations | **28 dial refusals** (1.4%) | **0** |

  The defect was real, it was reproducible, and it is gone.

  **What was *not* verified, and must not be read into the above:** the
  suite-level figures the fix reported (0 in 40 of each invocation) did not
  reproduce here, and neither did the 3-in-40 baseline. `go test -count=1 ./...`
  failed **40 of 40** at `83fd9f7` and **40 of 40** at `1d9505c` on this
  machine, for the ephemeral-port reason that `b3d251d` has since fixed and
  that was never this change's doing. Both figures are superseded by this pass:
  see the entry below and [`STAGE.md`](STAGE.md) §2.
- **Item 1, half (a) · The relay announces the scheme it actually serves** —
  delivered `83fd9f7` (2026-08-31 10:47), released as
  `pumasi/releases/2026-08-31-pumasi-tunnel-public-scheme.md`, **Q-020**, 7-day
  window. `core.ParsePublicScheme` (`core/route.go:43`) is the only place the
  legal set is written; the scheme is stored once on the registry and read once
  by `Registry.PublicURL`; `relay.New` validates it at startup and **refuses to
  start** on an unknown value; `cmd/pumasi-relay` exposes `-public-scheme`
  defaulting to `http`. All three surfaces that show a person an address read
  that one function. Tests: `core/scheme_test.go`, `relay/scheme_test.go`.
  **Not delivered: the deploy.** See item 1 — the running relay still announces
  `https://`.
- **`roadmap/MARKET.md`** — written in the `83fd9f7` pass:
  [`MARKET.md`](MARKET.md), three comparators, every price read from the
  vendor's own page on 2026-08-31 with the URL and the date in the claim. Its §4
  records the two comparisons that go against this product.
  *Product-manager action, not a build.*
- **Seeded 1 · Pure-core stream multiplexer** — delivered `8ff1605` as `mux/`,
  with its own framing rather than `yamux` or `quic-go`. Tests:
  `mux/session_test.go`. *Not delivered from the seed's text: throughput
  benchmarks.*
- **Seeded 2 · SSH ingress gateway** — delivered `3652e15` as
  `relay/sshingress.go` (`golang.org/x/crypto/ssh`), serving `-ssh-addr :2222`.
  Live: `pumasi.link:2222` answers `SSH-2.0-pumasi-tunnel`. *Port 22/443 ingress
  is not delivered and is not wanted — 22 is the host's own sshd and 443 is item
  1's.*
- **Seeded 3 · HTTP wildcard host router** — delivered `a69008f` and `bf837ee`:
  `core/route.go` (case-insensitive registry, wildcard matching, reserved-name
  list in `core/subdomain.go`) plus the relay's visitor HTTP path. Tests:
  `core/route_test.go`, `relay/endtoend_test.go`. **The ACME/wildcard-TLS half
  of this seed item is *not* delivered — it is item 1(ii).**
- **Seeded 4 · Raw TCP port pool** — delivered `a13e586` and `a5b77fc`:
  `core/portpool.go` (forward-walking cursor) and `relay/tcp.go` (byte-for-byte
  `pipe`, half-close where the stream supports it). Tests:
  `core/portpool_test.go`, `relay/tcp_test.go`. Live and load-bearing:
  `pumasi.link:20000` carries this machine's sshd (last measured at the `83fd9f7` pass; not re-measured here — this seat did not touch the live host). **Its
  announce-before-bind defect is delivered above; its bindability defect is
  item 3, and its test-range defect is delivered at `b3d251d` for
  `tcp_test.go` with a residual at item 9. Its `held` state, and the ownership
  the pool now enforces, are delivered at `4489fbe` — see the first *Delivered*
  entry.**
- **Seeded 5 · Client CLI, in part** — delivered `bf837ee`:
  `cmd/pumasi-tunnel` with `--relay`, `--subdomain`, `--token`, `--host`,
  `--tcp`, `--tcp-port`, `-v`, one static binary. *The interactive TUI and the
  published macOS/Windows binaries are not delivered — item 11. The seed's
  `--http` and `--auth` do not exist: HTTP is the default and the flag is
  `--token`.*
- **Not seeded, delivered anyway** — the relay console at the apex
  (`relay/dashboard.go`, `relay/dashboard.html`, `b3585f6`), shaped by
  `docs/ux/incumbent-ux-spec.md`, live at `http://pumasi.link/` and the Stage 1
  Surface B. Its gap is item 4 — *the console never offers the zero-install
  `ssh -R` command*.

---

## Not on this list, and why

- **`pumasi/catalog.json` has no entry for this product** — zero occurrences of
  the string `tunnel`, re-checked this pass at `pumasi` `23bbc64`, a commit that changed the file
  (`grep -c tunnel catalog.json` = **0**). It is a real defect and it is recorded
  in
  [`STAGE.md`](STAGE.md) under known gaps. It is **not** a backlog item here
  because it is not this repository's file and **no role owns it** — that is
  `pumasi/DECISIONS.md` **Q-019**, open. Nothing a coder packet on this product
  could build would close it.
- **Deploying the relay** — **Q-014**, open, and outside CHARTER Part 0. Named
  in item 1(i) as a blocker; not requested of anyone here. **This pass added
  evidence to that entry and set nothing**: no deployer named, no date, no
  default read, no window touched.
- **Merging `PRODUCT-RULES.md` to `pumasi` main — done.** `pumasi` `23bbc64`,
  2026-09-01 15:42 UTC, **v1.0**, steward-published; `git ls-tree origin/main
  PRODUCT-RULES.md` finds it, and this is the **first** evaluation of eleven
  to read it from `main` rather than from the `worktree-product-rules` branch.
  Items 6 and 8 cite it there now, by version and by clause. Whether **Q-017**
  is closed by that is the steward's; this seat did not touch the entry.

- **Release notes for slice 1 and slice 2 — both owed on this seat's reading,
  neither written here, and now ranked rather than footnoted.**
  `pumasi/releases/` at `23bbc64` still carries exactly **one** tunnel note,
  `2026-08-31-pumasi-tunnel-public-scheme.md`, for `83fd9f7`. The reasoning
  this bullet carried at `41f373c` holds and is not repeated: **Q-034**'s
  default answers the general question (an ordinary merge owes nothing; a
  can-hurt release owes a published note with a 7-day window), so `1d9505c`
  and `fd523e8` owe nothing, and the live question was whether slice 1 is
  can-hurt. **Slice 2 joins it, on a stronger footing** — it writes token
  digests to disk and the charter classes an unmapped path as can-hurt by
  default — and both are now in **item 2**, where the review that a can-hurt
  release needs first is ranked. `pumasi/releases/` is on none of this role's
  *May write* entries and CHARTER §2.1 puts the classification with the
  merging builder; Q-034's own text says the same of this role. *(Q-034 still
  sits under `## Closed` with a Status row reading `open` at `23bbc64`; still
  reported, still not this seat's to fix.)*

- **A second live tunnel on `pumasi.link` that this product cannot account
  for.** `GET http://pumasi.link/_pumasi/status` at 02:48 UTC 2026-09-01 reports
  `"count":2`: the long-running `sshsteward` tunnel
  (`pumasi.link:20000` → port 22, `"opened_at":"2026-08-31T06:18:13Z"`,
  `"age_secs":73789` — 20 h 30 m unbroken, `"fixed":true`), and a second one,
  subdomain **`skk6g7tyrs`**, `pumasi.link:20002` → a `"local_port":3389`
  somewhere, `"fixed":false`, `"opened_at":"2026-09-01T01:48:23Z"` and open for
  59 minutes at the time of the read. **This seat did not establish who opened
  it and will not guess.** It is recorded because three things follow and none of
  them is a build: it is the first traffic on this relay that is not the
  steward's own ssh route; it is `AllowAll` doing exactly what it is specified
  to do, since **an anonymous agent taking a free name is not something slice 1
  changed and was never meant to be** (`spec/0004` §6, third column, and frozen
  case **C-3**, which exists to fail if anyone ever withdraws it); and it
  **doubles what a restart costs**, which is the fact **Q-014** turns on and
  whose text still says the live set is *"exactly one"*. Nothing was done to the
  host, the tunnel or the entry.

  **Re-read at 05:51:04 UTC 2026-09-01, no host touched, and both are still
  there**: `sshsteward` `"age_secs":84771` — **23 h 32 m** unbroken, the same
  connection as at the last three passes — and `skk6g7tyrs` `"age_secs":14561`,
  **4 h 02 m**, the same `"opened_at":"2026-09-01T01:48:23Z"` as before, so it is
  one connection persisting rather than a series of new ones. Still
  unattributed, and this seat still does not guess.

  **Gone at 18:51:06 UTC 2026-09-01, re-read read-only by this pass.**
  `"count":1`: only `sshsteward` remains, `"age_secs":131572` — **36 h 32 m**,
  the same `"opened_at":"2026-08-31T06:18:13Z"` as at every pass since the
  first. `skk6g7tyrs` was open for at most seventeen hours and nobody here ever
  learned whose it was. What that changes: Q-014's *"exactly one"* is true again
  and this pass says so in that entry's evidence. What it does not change: the
  relay ran `AllowAll` throughout, still does, and would accept the next one.

- **The unattributed change to `cmd/pumasi-tunnel/main.go` — now stashed, and
  still not ranked.** Three hunks making `--relay` **default to
  `pumasi.link:7000`** instead of being required, dropping `--relay` from the
  usage string, and deleting the `*relayAddr == ""` arm of the argument check.
  **No job ever claimed it.** Job `0081` read the reasoning below, agreed with
  it, and **`git stash`ed the diff with a message naming why** rather than
  committing or discarding it — verified by this seat at `9e2de66` and again at `9cc9e65`:
  `git stash list` carries one entry, `git stash show --stat stash@{0}` reports
  `cmd/pumasi-tunnel/main.go | 6 +++---`, and the working tree is clean. **That
  was the right call and this entry keeps its reasoning**, because a stash is
  recoverable and a decision is not:

  **It contradicts the file this repository has now repaired twice.**
  Re-read at `9e2de66`: [`README.md`](../README.md) line 69 says `--relay` is
  **required**, line 75 tabulates its default as *(none — required)*, and all
  four invocations at lines 59–66 pass it explicitly — the same six lines as
  last pass, unmoved by `3626562`'s token paragraphs. `87244af`'s subject is
  *"the front door stops contradicting the source tree"* and `3626562`'s is
  *"the front door says what a token now does, and what it still cannot do"*;
  this diff would make the front door wrong again, in the same place, for the
  third time.

  **It is not ranked as a build, and the reason is not that it is small.** The
  ergonomics gain is narrow: the zero-install path this product leads with is
  `ssh -p 2222 -R …`, which needs no binary and no flag at all, so a default
  only shortens a command for someone who has already cloned and built. Against
  that, omitting a required flag today prints usage and exits 2, whereas with
  this diff `./pumasi-tunnel 8080` **publishes a local port to a relay the user
  never named** — a relay that runs `AllowAll` and serves plaintext (item 1, and
  slice 1 does not change that: `spec/0004` §3 is explicit that a token proves
  continuity, not permission, and `AllowAll` still answers *may this agent use
  this relay* with yes). Turning a usage error into an unrequested publication is a change to what
  the product promises; it needs an intent statement and a window, not a
  working-tree edit, and if it is taken it moves README lines 69 and 75 in the
  same commit. **A seat that disagrees should rank it in the order above with
  that cost written down. It should not be committed as found.**

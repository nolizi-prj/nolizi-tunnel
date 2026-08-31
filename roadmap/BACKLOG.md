# BACKLOG — what gets built next, in order

**Owned by the product-manager role**
([`pumasi-ops/roles/product-manager.md`](https://github.com/pumasi-ai/pumasi-ops/blob/main/roles/product-manager.md),
duty 5). Seeded 2026-08-30 from candidate `0011-developer-tunnels.md`; first
evaluation pass **2026-08-31** at `3652e15`.

One list, features and bugs together — a priority that cannot compare them is
not a priority. Every entry points at its source and carries one line of
why-here. **The top of this file is what the project manager's next coder
packet builds** — except where an entry says in its own text that it is
operator action rather than a build; the packet then takes the highest entry
that *is* a build, and the operator item keeps its rank rather than being
demoted for being unbuildable. Reordering is a commit with the reasoning in the
message; the steward vetoes by reverting.

What the ordering assumes: the seeded items 1–4 are **delivered** and are
recorded under *Delivered* below, verified against the tree rather than taken
from the commit subjects. What remains is the distance between what the product
does and what it tells people it does.

---

## The order

**1 · Every tunnel is handed an `https://` address that nothing answers** —
source: this evaluation, checking the live relay rather than the code path.
`Registry.PublicURL` (`core/route.go:255`) returns
`"https://" + name + "." + baseDomain` unconditionally; the relay puts it in
the auth response, the agent prints it, and the console shows it. Nothing
listens on 443: `curl https://sshsteward.pumasi.link` → *Could not connect to
server*, port 443 refused; only 80 answers. TLS termination is deliberately
outside the relay (`cmd/pumasi-relay/main.go` header — an operator may want
ACME, a purchased certificate, or none on a private network), and outside it
there is currently nothing. Two halves: **(a)** a build — the relay must
announce the scheme it is actually serving (a `-public-scheme` or
`-tls-terminated` flag defaulting to `http`, honoured by `PublicURL`, the
console and the ssh-ingress banner alike, from one place); **(b)** *operator
action, not a build* — put a wildcard certificate for `*.pumasi.link` in front
of the relay on the Vultr host. Half (b) restarts or fronts a live host: see
`DECISIONS.md` **Q-014**, and `RESOURCES.md` §2's warning that proxying these
records through Cloudflare would break raw TCP.
Why here: it is the only defect on this list that every single user meets, on
their first line of output, and it is the product printing something untrue
about itself. Half (a) alone stops the lie in a few lines.

**2 · The public TCP address is announced before anything listens on it** —
source: this evaluation, from a flaky test traced to the product. `go test
-count=1 -cover ./...` fails **2 runs in 12**, always
`TestRawTCPCrossesTheTunnel … dial tcp 127.0.0.1:34000: connect: connection
refused`; the same suite without `-cover` passed 12 of 12. Not a test artifact:
`relay/relay.go` writes the auth response carrying `TCPAddr` (~line 162) and
only then calls `bindTCP` (~line 181), so the agent — and the user reading its
output — has the address before the listener exists, and a bind failure is
reported *after* the address was announced. Coverage instrumentation widens a
window that is in the product. `relay/sshingress.go:182` has the same shape.
Fix: bind before the response leaves, or the response waits on the bind.
Why here: it is correctness of shipped surface, it makes a documented promise
(`--tcp-port` "keeps an address across reconnects") intermittently false at the
moment it matters most, and until it is fixed the Stage 1 gate's "100%" is a
statement about one flag rather than about the code.

**3 · A subdomain belongs to nobody, and nothing survives a relay restart** —
source: `VALUE.md` pillar 2, which sells "permanent" and "stable across
restarts", checked against the tree. `Tunnel.Reserved` is computed at
`relay/relay.go:236` and **never read anywhere**; `cmd/pumasi-relay` has no
auth flag, so `AllowAll` (`relay/relay.go:40`) is the only authenticator it can
run — any anonymous agent may take any free name, including one another person
is using between reconnects. And the registry and port pool are in-memory
maps: a relay restart drops every name, every reserved port and every live
tunnel at once. `--tcp-port` (`a5b77fc`) survives an *agent* reconnect, not a
relay one.
Why here: this is the `beta` bar itself — strangers relying on it, data
surviving (`STAGE.md`, "Why not `beta`"). It is also the largest single piece
of work on this list, which is why it sits below two cheap correctness fixes
rather than above them.

**4 · The console never offers the zero-install `ssh -R` command** — source:
`VALUE.md` pillar 1 against the live page. The ingress works: a banner grab on
`pumasi.link:2222` returns `SSH-2.0-pumasi-tunnel` (`3652e15`). But
`relay/dashboard.html`'s command builder emits only
`pumasi-tunnel --relay …` (line ~223) and its "First time here" panel offers
only `git clone && go build` — so the product's headline claim, the one thing
it does that needs nothing installed, is absent from the one page a visitor
sees.
Why here: it is an afternoon's work on a page that is already live, and it
converts the strongest differentiator from a README sentence into the thing
you can paste. Cheap, so it beats the larger items below.

**5 · `agent/` has no tests** — source: this evaluation's gate reading; L-006.
`go test ./...` reports **no test files** for `agent`, `cmd/pumasi-relay` and
`cmd/pumasi-tunnel`. The two `cmd` packages are argument parsing and can wait.
`agent/` cannot: it is half of every tunnel — handshake, reconnect, local dial,
stream fan-out — and today the only thing exercising it is `relay`'s end-to-end
tests, which use it as a fixture and assert on the relay's behaviour, not its.
Reconnect and local-dial-failure have no case at all.
Why here: it is the coverage the Stage 1 number does not include, and item 3
will rewrite reconnect behaviour — the tests should exist before that, not
after.

**6 · PR-1 compliance: a version that moves and is user-visible** — source:
[`PRODUCT-RULES.md` PR-1](https://github.com/pumasi-ai/pumasi/blob/worktree-product-rules/PRODUCT-RULES.md)
(v1.0, 2026-08-30; **binds always, from the first commit** — read fresh
2026-08-31 and still only on the unmerged `worktree-product-rules` branch,
`0115758`; its absence from `main` is not compliance). This product has no
version anywhere: no version file, no `/version`, nothing in the console
footer, nothing in a release note. `core.AuthRequest.ClientVersion` exists as a
field and no binary sets it.
Why here: PR-1's own reason applies literally — the relay is a single live host
and, when the next fix lands, nothing on the console or in the logs will say
which build is serving. That is `pumasi-booking`'s Q-012 problem arriving
early, and here the answer is four lines of Go plus the field that is already
in the wire protocol.

**7 · PR-2 compliance: in-app feedback** — source:
[`PRODUCT-RULES.md` PR-2](https://github.com/pumasi-ai/pumasi/blob/worktree-product-rules/PRODUCT-RULES.md)
(binds at the `beta` promotion; below `beta`, encouraged). Nothing in the
product collects feedback. The reference behaviour is `pumasi-booking`'s
(`service/src/feedback.ts`) — matched in behaviour, not copied.
Why here: it **gates the `beta` promotion**, so it must be built before the
label moves, and it is worth little before items 1–3 make the thing worth
reporting on. The natural home is the console, where a visitor already is.

**8 · Client CLI: the interactive terminal UI** — source: the seeded item 5,
the delivered half of which is recorded below;
[`docs/ux/incumbent-ux-spec.md`](../docs/ux/incumbent-ux-spec.md) §273 for what
the incumbents' status line shows. `cmd/pumasi-tunnel` prints logs; there is no
live view of requests, response times or status codes, and no released binary
for macOS or Windows (Go cross-compiles; nothing publishes them).
Why here: it is polish on a path that works, and every item above it is either
something untrue or something missing that a user would rely on.

**9 · Local request inspector on `127.0.0.1:4040`** — source: the seeded item
6; `VALUE.md` (where it is now marked *not built*). `web/` is an empty
directory; there is no listener, no SSE, no replay. It is the seed's most
ambitious item and the one incumbent feature this product currently claims
nothing about.
Why here: last of the build items. It is a whole second product surface, and
the honest catalog page (`pumasi-web` `843bdef`) already tells visitors it does
not exist, so nobody is being misled while it waits.

**10 · `roadmap/MARKET.md` does not exist** — source: the product-manager role,
duty 3, and `VALUE.md`'s own opening, which compares this product to
alternatives "named in `MARKET.md`". Competitor pricing was removed from
`VALUE.md` in this pass because it was uncited, and uncited is not allowed to
survive anywhere; the numbers belong in `MARKET.md` with sources.
Why here: it is this seat's own work, not a coder packet, and it is listed so
that the debt is visible rather than quietly carried. *Product-manager action,
not a build.*

---

## Delivered

Verified against the tree during this evaluation, not taken from commit
subjects. Seeded items 1–4 are complete; item 5 is partly complete and its
remainder is item 8 above.

- **Seeded 1 · Pure-core stream multiplexer** — delivered `8ff1605`
  (2026-08-30 22:47) as `mux/` (`session.go`, `stream.go`, `netconn.go`), not
  the seeded `core/mux.go`, and with its own framing rather than `yamux` or
  `quic-go`. Tests: `mux/session_test.go`, 83.5% coverage. *Not delivered from
  the seed's text: throughput benchmarks.*
- **Seeded 2 · SSH ingress gateway** — delivered `3652e15` (2026-08-31 01:19)
  as `relay/sshingress.go` (`golang.org/x/crypto/ssh`, not `gliderlabs/ssh`),
  serving `-ssh-addr :2222`. Live: `pumasi.link:2222` answers
  `SSH-2.0-pumasi-tunnel`. The banner returns the public address, as the seed
  asked. *Port 22/443 ingress is not delivered and is not wanted — 22 is the
  host's own sshd and 443 is item 1's.*
- **Seeded 3 · HTTP wildcard host router** — delivered `a69008f` (22:28) and
  `bf837ee` (22:57): `core/route.go` (case-insensitive registry, wildcard
  matching, reserved-name list in `core/subdomain.go`) plus the relay's visitor
  HTTP path, forwarding bytes with no interstitial. Tests: `core/route_test.go`,
  `relay/endtoend_test.go`. **The ACME/wildcard-TLS half of this seed item is
  *not* delivered — it is item 1.**
- **Seeded 4 · Raw TCP port pool** — delivered `a13e586` (23:19) and `a5b77fc`
  (00:46): `core/portpool.go` (forward-walking cursor so a released port is not
  immediately reissued) and `relay/tcp.go` (byte-for-byte `pipe`, half-close
  where the stream supports it). Tests: `core/portpool_test.go`,
  `relay/tcp_test.go`. Live and load-bearing: `pumasi.link:20000` has carried
  this machine's sshd for nearly eight hours. **Its announce-before-bind defect
  is item 2.**
- **Seeded 5 · Client CLI, in part** — delivered `bf837ee`:
  `cmd/pumasi-tunnel` with `--relay`, `--subdomain`, `--token`, `--host`,
  `--tcp`, `--tcp-port`, `-v`, one static binary, no dependencies outside the
  standard library and `golang.org/x/crypto`. *The interactive TUI and the
  published macOS/Windows binaries are not delivered — item 8. The seed's
  `--http` and `--auth` do not exist: HTTP is the default and the flag is
  `--token`.*
- **Not seeded, delivered anyway** — the relay console at the apex
  (`relay/dashboard.go`, `relay/dashboard.html`, `b3585f6`, 01:04), shaped by
  `docs/ux/incumbent-ux-spec.md`, live at `http://pumasi.link/` and the Stage 1
  Surface B. Its gap is item 4.

# BACKLOG — what gets built next, in order

**Owned by the product-manager role**
([`pumasi-ops/roles/product-manager.md`](https://github.com/pumasi-ai/pumasi-ops/blob/main/roles/product-manager.md),
duty 5). Seeded 2026-08-30 from candidate `0011-developer-tunnels.md`; first
evaluation pass 2026-08-31 at `3652e15`; **post-release evaluation 2026-08-31
at `83fd9f7`**, after the `-public-scheme` merge and its release note
(`pumasi/releases/2026-08-31-pumasi-tunnel-public-scheme.md`, `pumasi`
`33a96b0`; **Q-020**).

One list, features and bugs together — a priority that cannot compare them is
not a priority. Every entry points at its source and carries one line of
why-here. **The top of this file is what the project manager's next coder
packet builds** — except where an entry says in its own text that it is
operator action rather than a build; the packet then takes the highest entry
that *is* a build, and the operator item keeps its rank rather than being
demoted for being unbuildable. Reordering is a commit with the reasoning in the
message; the steward vetoes by reverting.

**What changed in this pass, and why.** Item 1 was two halves. Half **(a)** —
the relay announcing the scheme it actually serves — **is merged at `83fd9f7`**
and is verified against the tree below, not against the commit subject; it
moves to *Delivered*. Half **(b)**, the certificate, is operator action and
**keeps rank 1** under this file's own rule. Item 2 is therefore the highest
*build* entry and is the next coder packet — and it earned that place twice
over, because re-measuring it found it **worse** than the number this file
inherited: the race is not an artifact of `-cover` at all. Item 10 (this file's
`MARKET.md` debt) was this seat's own work and was done in this pass; it moves
to *Delivered*.

---

## The order

**1 · Nothing serves TLS, and the relay that is actually running still says it
does** — *operator action and a blocked deploy, not a build.* Source: this
evaluation, checking the live relay rather than the code path.

The build half is done and the lie is gone from `main`; it is not gone from the
internet. Verified 2026-08-31 16:21 UTC against the running host:

```
$ curl -s http://pumasi.link/_pumasi/status
{"base_domain":"pumasi.link", ... "subdomain":"sshsteward",
 "url":"https://sshsteward.pumasi.link","tcp_addr":"pumasi.link:20000",
 "local_port":22,"fixed":true,"age_secs":36167}

$ curl https://pumasi.link/
curl: (7) Failed to connect to pumasi.link port 443 — Could not connect to server
$ curl -o /dev/null -w '%{http_code}' http://pumasi.link/   →  200
```

Two things are outstanding, and they are different sizes:

- **(i) Deploy the merged fix.** `pumasi.link` runs a pre-`83fd9f7` binary and
  will announce `https://` until someone restarts the relay. **This is blocked
  on `pumasi/DECISIONS.md` Q-014**, which asks who may restart a host whose one
  live tunnel is `sshsteward` → this machine's port 22 — `RESOURCES.md` §4's
  remote-access route, open 10 h 2 m at the measurement above. Q-014 is open and
  is **explicitly outside CHARTER Part 0's proceed-on-default rule**. Neither
  this seat nor a coder packet may take it, and this file does not ask for it.
- **(ii) Put a wildcard certificate for `*.pumasi.link` in front of the relay
  on the Vultr host.** This is the actual TLS gap: with (i) done, every tunnel
  is *honestly* plaintext, which is still a product that no `https://`-only
  webhook sender can be pointed at. TLS termination is deliberately outside the
  relay (`cmd/pumasi-relay/main.go` header — an operator may want ACME, a
  purchased certificate, or none on a private network), and outside it there is
  still nothing. `RESOURCES.md` §2 warns that proxying these records through
  Cloudflare would break raw TCP, so that is not the shortcut it looks like.

Why here: it is the largest single gap between what this product is and what a
stranger could use, it is the one item on this list every visitor to
`pumasi.link` meets today, and it is not demoted for being unbuildable. It is
also the entry a reader is most likely to mistake for done — `main` is fixed,
the internet is not.

**2 · The public TCP address is announced before anything listens on it** —
**the next coder packet.** Source: this evaluation, re-measured at `83fd9f7`.

The shape is unchanged by the merge and is still in the product. `relay/relay.go`
writes the auth response carrying `TCPAddr` at **line 175**
(`r.writeFrame(conn, okFrame)`) and only calls `bindTCP` at **line 194** — so
the agent, and the user reading its output, holds the public address before the
listener exists, and a bind failure is reported *after* the address was
announced. `relay/sshingress.go:182` has the same shape on the ssh path.

**The number this file inherited is wrong, and the truth is worse.** Measured
here, 80 runs:

| Invocation | Result |
| :--- | :--- |
| `go test -count=1 -cover ./...` | **3 failures in 40** |
| `go test -count=1 ./...` | **3 failures in 40** |

Coverage instrumentation is **not** the trigger. The earlier reading — 2 in 12
under `-cover`, 12 of 12 clean without — was a 12-run sample that happened to
miss it. Every failure is the same dial against the same public port, and it
now surfaces on **three different tests**, which is itself the evidence that it
is not one test's artifact:

```
--- FAIL: TestTCPPortReleasedWhenAgentDisconnects
    tcp_test.go:208: port should be live before disconnect:
                     dial tcp 127.0.0.1:34000: connect: connection refused
--- FAIL: TestConcurrentTCPClients
    tcp_test.go:195: dial tcp 127.0.0.1:34000: connect: connection refused
--- FAIL: TestRawTCPCrossesTheTunnel   (the previously reported one)
```

Fix: bind before the response leaves, or make the response wait on the bind —
on both paths. **Fixer: the coder** (this seat may not edit product code).
Why here: it is correctness on shipped surface, it makes a documented promise
(`--tcp-port` "keeps an address across reconnects") intermittently false at the
one moment it matters, it is a **7.5% failure rate on `main`'s own test command
in both invocations**, and until it is fixed the Stage 1 gate's "100%" is a
statement about a lucky sample rather than about the code (L-006). It is the
highest buildable entry and it is cheap.

**3 · A subdomain belongs to nobody, and nothing survives a relay restart** —
source: `VALUE.md` claim 2, which sold "permanent" and "stable across
restarts", checked against the tree at `83fd9f7`. All three facts re-verified
this pass: `Tunnel.Reserved` is computed at `relay/relay.go:236`
→ **now line 249** and is **still never read anywhere** (`grep -rn Reserved
--include=*.go` finds the field definition at `core/route.go:127`, that one
write, and nothing else); `cmd/pumasi-relay` has no auth flag among its
eleven, so `AllowAll` (`relay/relay.go:40`, installed at `relay.go:97` when
`cfg.Auth` is nil) is the only authenticator that binary can run; and the
registry and port pool are plain in-memory maps (`core/route.go:145–147`,
`core/portpool.go:27–29`, with no persistence path anywhere in `core/` or
`relay/`). A relay restart drops every name, every reserved port and every live
tunnel at once. `--tcp-port` (`a5b77fc`) survives an *agent* reconnect, not a
relay one.
Why here: this is the `beta` bar itself (`STAGE.md`, "Why not `beta`") and it
is what retires **Q-014** — once a restart costs nobody their address, who may
deploy stops being a steward question and becomes an ordinary one. It is also
the largest single piece of work on this list, which is why it sits below one
cheap correctness fix rather than above it.

**4 · The console never offers the zero-install `ssh -R` command** — source:
`VALUE.md` claim 1 against the live page. The ingress works: a banner grab on
`pumasi.link:2222` returns `SSH-2.0-pumasi-tunnel`, re-verified 2026-08-31
16:21 UTC. But `relay/dashboard.html`'s command builder emits only
`pumasi-tunnel --relay …` and its "First time here" panel offers only
`git clone && go build` (line ~166) — so the product's headline claim, the one
thing it does that needs nothing installed, is absent from the one page a
visitor sees.
Why here: it is an afternoon's work on a page that is already live, and it
converts the strongest differentiator from a README sentence into the thing you
can paste. `MARKET.md` §2 now makes that differentiator explicit and cited,
which raises what the omission costs.

**5 · `agent/` has no tests** — source: this evaluation's gate reading; L-006.
`go test ./...` reports **no test files** for `agent`, `cmd/pumasi-relay` and
`cmd/pumasi-tunnel`. The two `cmd` packages are argument parsing and can wait.
`agent/` cannot: it is half of every tunnel — handshake, reconnect, local dial,
stream fan-out — and today the only thing exercising it is `relay`'s
end-to-end tests, which use it as a fixture and assert on the relay's behaviour,
not its. Reconnect and local-dial-failure have no case at all.
Why here: it is the coverage the Stage 1 number does not include, and item 3
will rewrite reconnect behaviour — the tests should exist before that, not
after.

**6 · PR-1 compliance: a version that moves and is user-visible** — source:
[`PRODUCT-RULES.md` PR-1](https://github.com/pumasi-ai/pumasi/blob/worktree-product-rules/PRODUCT-RULES.md)
(v1.0, 2026-08-30; **binds always, from the first commit**; still only on the
unmerged `worktree-product-rules` branch — **Q-017** — and its absence from
`main` is not compliance). This product has no version anywhere: no version
file, no `/version`, nothing in the console footer, nothing in a release note.
`core.AuthRequest.ClientVersion` exists as a field and no binary sets it.
Why here, and it went up in this pass: the `-public-scheme` merge is **exactly**
the scenario PR-1 names. There is now a build on `main` that behaves differently
from the build on the host, and **nothing on the console, in `/_pumasi/status`
or in the logs says which one is answering** — the only way this evaluation
could tell them apart was to read the `url` field and infer. That is
`pumasi-booking`'s Q-012 problem arriving early, and here the answer is a few
lines of Go plus a field already in the wire protocol.

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
6; `VALUE.md` (where it is marked *not built*). `web/` is an empty directory;
there is no listener, no SSE, no replay. It is the seed's most ambitious item.
Why here: last of the build items. It is a whole second product surface, and
the honest catalog page (`pumasi-web` `843bdef`) already tells visitors it does
not exist, so nobody is being misled while it waits. `MARKET.md` §2 records
that all three incumbents ship one; that makes it a real gap, not a whim, but
not ahead of anything above.

---

## Delivered

Verified against the tree during this evaluation, not taken from commit
subjects.

- **Item 1, half (a) · The relay announces the scheme it actually serves** —
  delivered `83fd9f7` (2026-08-31 10:47), released as
  `pumasi/releases/2026-08-31-pumasi-tunnel-public-scheme.md`, **Q-020**,
  7-day window. Verified in the tree, not from the subject line:
  `core.ParsePublicScheme` (`core/route.go:43`) is the only place the legal set
  is written; the scheme is stored once on the registry
  (`core/route.go:142,159–166`) and read once by
  `Registry.PublicURL` (`core/route.go:311`); `relay.New` validates it at
  startup and **refuses to start** on an unknown value (`relay/relay.go:110–114`);
  `cmd/pumasi-relay` exposes `-public-scheme` defaulting to `http`
  (`main.go:45,63`). All three surfaces that show a person an address read that
  one function: the auth response (`relay/relay.go:255`), the console
  (`relay/dashboard.go:71`), and the zero-install ssh banner
  (`relay/sshingress.go:190` → `sshGreet`, `:261`). Tests: `core/scheme_test.go`,
  `relay/scheme_test.go`. **Not delivered: the deploy.** See item 1 above — the
  running relay still announces `https://`.
- **Item 10 · `roadmap/MARKET.md`** — written in this pass:
  [`MARKET.md`](MARKET.md), three comparators, every price read from the
  vendor's own page on 2026-08-31 with the URL and the date in the claim. It
  also records, in its §4, the two comparisons that go against this product.
  *This was product-manager action, not a build.*
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
  *not* delivered — it is item 1(ii).**
- **Seeded 4 · Raw TCP port pool** — delivered `a13e586` (23:19) and `a5b77fc`
  (00:46): `core/portpool.go` (forward-walking cursor so a released port is not
  immediately reissued) and `relay/tcp.go` (byte-for-byte `pipe`, half-close
  where the stream supports it). Tests: `core/portpool_test.go`,
  `relay/tcp_test.go`. Live and load-bearing: `pumasi.link:20000` has carried
  this machine's sshd for over ten hours. **Its announce-before-bind defect is
  item 2, and it is measurably worse than previously recorded.**
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

---

## Not on this list, and why

- **`pumasi/catalog.json` has no entry for this product** — zero occurrences of
  the string `tunnel`, verified at `pumasi` `6489347`. It is a real defect and
  it is recorded in [`STAGE.md`](STAGE.md) under known gaps. It is **not** a
  backlog item here because it is not this repository's file and **no role owns
  it** — that is `pumasi/DECISIONS.md` **Q-019**, open. Nothing a coder packet
  on this product could build would close it.
- **Deploying the relay** — **Q-014**, open, and outside CHARTER Part 0. Named
  in item 1(i) as a blocker; not requested of anyone here.

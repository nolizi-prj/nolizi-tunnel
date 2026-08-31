# Spec 0001 · The public scheme is configured once and read everywhere

**Status:** frozen at spec review · **Intent:** [`INTENT.md`](INTENT.md) ·
**Backlog:** `roadmap/BACKLOG.md` item 1, build half (a).

## 1 · The decision, and where it lives

There is exactly one place in the tree that decides what scheme a tunnel's
public address is announced under: the `publicScheme` field of
`core.Registry`, applied by `Registry.PublicURL`.

Nothing else concatenates a scheme onto a tunnel hostname. Three surfaces show
a user that address, and all three obtain it from `PublicURL` rather than
building their own:

| # | Surface | Path |
|---|---|---|
| 1 | The CLI's first line of output | `relay.authorize` → `core.AuthResponse.URL` → `agent` → `cmd/pumasi-tunnel` |
| 2 | The console at the relay's apex | `relay.serveStatus` → `/_pumasi/status` `url` → `relay/dashboard.html` |
| 3 | The zero-install ssh banner | `relay.ServeSSH` → `AuthResponse.URL` → `sshGreet` |

Surfaces 1 and 3 already share one string; the console builds its own view
struct but takes the same value. What was wrong was not that there were three
copies of the *code* — it is that there was one copy of a **constant** that
asserted something the relay cannot know. This spec keeps the single
implementation and makes the value configurable there rather than adding a
second, third and fourth way to reach it (L-007).

## 2 · The configuration

`cmd/pumasi-relay` gains one flag:

    -public-scheme string
          scheme tunnels are announced under: http, or https when a TLS
          terminator sits in front of this relay (default "http")

**Why `-public-scheme` and not `-tls-terminated`.**

1. It names the value the relay actually emits. `-tls-terminated` is a boolean
   about someone else's infrastructure, which would have to be translated into
   a scheme somewhere — and that translation is a second statement of the same
   rule, which is how a rule forks (L-007).
2. It matches the flag beside it. `-public-host` already means "what a visitor
   dials, which this process cannot observe". `-public-scheme` is the same
   kind of fact with the same kind of name, and the two are read together.
3. It cannot be true and useless. `-tls-terminated=true` says nothing about
   whether the terminator speaks HTTP/2, redirects, or listens on 443 at all;
   `https` is the whole of what the relay needs to print.

The flag's value travels `flag` → `relay.Config.PublicScheme` → `relay.New`
(validated) → `core.NewRegistry` → `Registry.publicScheme`. It is passed, never
re-derived.

## 3 · What the value may be

`core.ParsePublicScheme` is the only place the set of legal schemes is
written down.

- `http` and `https` are accepted, case-insensitively, with surrounding space
  and a trailing `://` tolerated (`HTTPS`, ` https `, `https://` all normalise
  to `https`).
- The empty string normalises to `http`. `http` is what the relay serves with
  nothing in front of it, so an unset flag under-promises rather than
  over-promises, and that is the only direction in which guessing is safe.
- Anything else is `core.ErrUnknownScheme`. `relay.New` returns that error and
  the relay does not start. A relay that could not be told which scheme it
  serves must not pick one — picking one is the defect this spec exists to
  remove.

`core.NewRegistry` takes an already-validated scheme. Given anything else it
uses `http`, and its doc comment says so: the registry fails closed to the
truthful default rather than propagating a value no surface can honour.

## 4 · Behaviour, precisely

For a registry rooted at base domain `D` with scheme `S`, and a tunnel
registered under label `N`:

    PublicURL(N) == S + "://" + lower(N) + "." + D

`D` is normalised by `NewRegistry` as it already was; `N` is lowercased as it
already was. Nothing else about routing, registration, TCP addresses or the
console changes. `AuthResponse.TCPAddr` is a `host:port` and has no scheme; it
is untouched.

## 5 · Acceptance cases

Frozen before implementation. Each names the execution that makes it fail
(L-006); the failing-first evidence is recorded in the release note.

See [`acceptance/CASES.md`](acceptance/CASES.md).

## 6 · Out of scope, and named so it is not read as covered

- **TLS in the relay.** Not added. The package header of `cmd/pumasi-relay`
  keeps its explanation and gains one sentence pointing at the flag.
- **The certificate in front of the relay** — `BACKLOG.md` item 1 half (b),
  operator action, `pumasi/DECISIONS.md` **Q-014**.
- **`BACKLOG.md` item 2**, the announce-before-bind race. `relay.authorize`
  is edited by this change; the ordering defect in `ServeAgent`/`ServeSSH` is
  not, and must not be read as fixed here.
- **`README.md`.** Its quickstart says the ssh method "outputs an instant,
  public HTTPS URL" — the same untruth, one surface over. It is not corrected
  here: that file also promises `gliderlabs/ssh` on port 22/443, a QUIC/yamux
  multiplexer and a port-4040 inspector, none of which exist, and fixing one
  line of it would leave a document that reads as checked and is not. It
  belongs to a pass over the whole file. Recorded for the roadmap owner.

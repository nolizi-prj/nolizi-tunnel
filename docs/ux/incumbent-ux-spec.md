# Incumbent UX specification — developer tunnels & localhost relay

**Product:** `pumasi-tunnel`
**Tour date:** 2026-08-30 (signed-in tours of three incumbents: ngrok, Pinggy, LocalXpose)
**Author:** tour analyst
**Status:** this document is the clean-room wall. Builders of `pumasi-tunnel` work from
this spec and do not see the tour screenshots.

## How to read this

Everything below is **behavior, flow, and information architecture**. Where a control is
described ("masked field with a reveal toggle and a copy button"), that is a pattern, not
a design. Nothing here reproduces the incumbents' wording, layout ornament, colour,
iconography, or visual identity, and nothing here should be treated as a licence to
imitate any of those. Where a screen was ambiguous or never resolved, this document says
so rather than guessing — see §10.

Vocabulary is deliberately industry-generic: *tunnel, endpoint, agent, reserved domain,
TCP address, authtoken, region, inspector, session, seat*.

Difficulty tags used in §9 follow the commons' shared taxonomy:
`trivial` (CRUD, forms, listing pages) · `moderate` (real logic, pure and specifiable) ·
`hard` (algorithmic depth, scale, or subtle correctness) · `ops-heavy` (integration,
delivery infrastructure, compliance).

---

## 1. What the product is

All three products do the same core job: **take a service listening on a private machine
and give it a publicly reachable address on the internet, without the operator touching
NAT, firewalls, DNS, or certificates.** A long-lived outbound connection from the user's
machine to the vendor's edge is reused in reverse to carry inbound public traffic. The
differences are all in emphasis.

**Incumbent A (ngrok) — breadth, and an enterprise ingress surface.** Positions itself
not as a localhost tunnel but as developer *ingress infrastructure* that routes and
secures traffic to applications, APIs and model endpoints. The tunnel is one of several
delivery mechanisms; the same control plane also fronts production services via a
Kubernetes operator and an infrastructure-as-code provider, and layers a policy engine,
secret vaults, mutual-TLS certificate authorities, IP policies, identity capture, and log
export on top. The dashboard is organised around long-lived **endpoints** (an endpoint is
the public access point; a domain or TCP address is the network reservation that backs
it) rather than around ephemeral tunnels. It carries the full enterprise checklist
(third-party security attestations, RBAC, SSO/SCIM, audit logs, data residency,
white-labelling) as a first-class marketing surface. It has by far the largest object
model and the largest paywall surface.

**Incumbent B (Pinggy) — SSH-first, zero-install simplicity.** Positions itself as public
URLs for localhost with *no binary to download*. The entire client is the operating
system's own SSH client: a single reverse-port-forward command against the vendor's edge
opens the tunnel, and every option (protocol, authentication token, debugger, QR output,
persistent domain) is encoded into the SSH username string of that one command. There is
no account requirement to start — an anonymous user gets a working HTTPS URL immediately
and is time-limited instead. The signed-in dashboard's centre of gravity is therefore not
a resource manager but a **command builder**: a form whose only output is the one-line
command to paste. Its object model is tiny (token, domain, active session) and its
paywall is expressed almost entirely as *session duration* and *hostname stability*.

**Incumbent C (LocalXpose) — a client-side GUI, regions, and a plugin-ish edge.** Ships a
single binary that is both a CLI and, when double-clicked, a **locally served web GUI**
opened in the user's browser. That local GUI — not the hosted dashboard — is where
tunnels are created, listed, started/stopped and inspected. The hosted dashboard is
comparatively thin: reservations, access token, billing, account. Region is a first-class
field on every tunnel and every reservation, with an explicit availability roadmap
(default / available / coming soon / planned). Its edge advertises built-in behaviours
that the others treat as policy actions: a request rate limiter, an HTTP→HTTPS redirect,
dynamic header editing, and a built-in static file server. It claims real-time inspection
of TCP and UDP traffic, not just HTTP.

### Where they differ, in one table

| Dimension | A (breadth) | B (SSH-first) | C (GUI/regions) |
|---|---|---|---|
| Client | Own agent binary + 5 language SDKs + K8s operator + IaC provider | None — the OS SSH client (optional first-party CLI and Docker image also offered) | One binary that is both CLI and a localhost-served web GUI |
| Account needed to start | Yes | **No** | Yes, plus email verification |
| Where tunnels are managed | Hosted dashboard + agent config file | Nowhere — the command *is* the tunnel; dashboard only builds the command | Local GUI (create/edit/start/stop), hosted dashboard for reservations only |
| Inspector | Hosted, with a local inspector too | Local web debugger on a fixed local port, plus a terminal TUI; a hosted browser debugger is being introduced | Local GUI pane, per-tunnel toggle |
| Primary object | Endpoint | Token (domains bind *to tokens*) | Tunnel (client-side), Reservation (server-side) |
| Free-tier mechanism gate | Quotas + capability removal + card-on-file for TCP | **Session timeout** + random hostname | Concurrent-tunnel cap + time limit + **interstitial warning page** + HTTP-only |

---

## 2. Onboarding & first-run

### 2.1 Signup paths

**A.** A single signup card offering two OAuth providers (a code host and a mail
provider) above a divider, then a local form of name / email / password with a
show-password toggle, a terms-acceptance checkbox that gates submission, and a link to
the log-in page. The OAuth return path lands on a small **confirmation card** that names
the email being claimed, repeats the terms checkbox, and offers cancel / create-account.
No email-verification step was observed before the dashboard; the user is dropped
straight into the product.

**B.** A single combined login/signup card offering **three** authentication mechanisms:
(1) email + password with an inline submit and a reset-password link; (2) a passwordless
**"send me a login code by email"** button; (3) two OAuth providers. Below that, a
separate create-account action, and a terms notice.

**C.** Separate login and signup pages, **email + password only** — no OAuth observed.
Signup collects email, password, and password confirmation. On submit, a transient
confirmation surface tells the user a verification message has been sent; **email
verification is a hard gate** before the product is usable, and the documented first-run
sequence lists it as step 2 of 7.

### 2.2 What happens immediately after signup

**A.** Lands on a dashboard whose default page *is* the quickstart. There is a separate
"share localhost" getting-started page with the same structure. Both are a numbered
four-step wizard:

1. **Choose your platform** — a three-column picker, each column labelled and described:
   *Command Line* ("a single executable with no runtime dependencies"): macOS, Windows,
   Linux, FreeBSD-and-others, Docker, Docker Desktop, Raspberry Pi. *SDKs* ("embed
   ingress into your application as a socket"): Go, Node.js, Rust, Python, Java.
   *Infrastructure*: Kubernetes operator, IaC provider. The chosen item is check-marked
   and rewrites steps 2–4 in place.
2. **Install** — for command-line platforms, a nested strip of package-manager tabs
   (e.g. on Windows: an OS app store, two third-party package managers, and a direct
   download), each with one copyable command. For SDK platforms, two sub-tabs:
   *clone the example repo* vs *add to an existing app*, each with copyable commands.
   For the Kubernetes path, a chart-repo add, then an install command that requires
   **both** an API key and an authtoken as values.
3. **Get a public URL** — a copyable run command (or, for the IaC path, a resource
   declaration; or for SDK paths, a code sample) with the account's assigned free
   development hostname already substituted in. Below it, a **live detection panel**: a
   spinner stating the dashboard is waiting for the app to come online and will detect
   it automatically. This is a polling loop against the control plane — a genuinely good
   pattern and cheap to build.
4. **Secure your endpoint** — adds an authentication/policy step on top of the working
   tunnel.

   Everywhere the authtoken appears inside a code block, it is masked by default with a
   **reveal control in the block's header** that swaps the placeholder for the live value
   in place. The token is never printed unmasked without an explicit action.

**B.** Lands on a quickstart that is a numbered four-step page:
1. **Choose an app/service** — two tabs: pick from a catalogue of well-known local
   services (web frameworks, CMSes, databases, media servers, game servers, IoT/SBC
   remote-access, self-hosted model runners, automation tools), *or* type a raw local
   address. Below, an **access-token selector**: a dropdown over the account's tokens,
   masked, with copy and reveal buttons.
2. **Paste this command** — OS tabs (Windows / Linux / macOS) that change only the
   surrounding instruction (which shell to open) and the command's quoting; then the
   assembled one-line command with **Copy** and **Download** (save as script) actions.
3. **More settings** — two inline toggles (enable the local web debugger; emit a QR code
   for the resulting URL), a link through to the full command builder, and a line naming
   the fixed local port the web debugger listens on.
4. **Documentation** — method tabs: SSH, first-party CLI, desktop app, Node SDK, Python
   SDK, Docker; each showing the equivalent invocation.

**C.** Lands on an empty tunnels list. Getting to a first tunnel requires the
seven-step documented path: sign up → verify email → log in → **copy access token** →
download and start the client → log in inside the client with that token → create the
tunnel. The dashboard has a dedicated **Access** page that is effectively the installer:
1. **Download** — an OS button row (Linux / macOS / Windows / FreeBSD / Docker).
   Choosing one reveals a per-OS package-manager tab strip (e.g. a system package manager
   and a JS package manager) with one copyable install command, *plus* per-distribution
   direct downloads (Debian-family, RPM-family, Arch-family, plain binary), each with an
   **architecture select** and its own download button.
2. **Access token** — a wide masked field with a copy button and a destructively styled
   regenerate button beside it.
3. **Log in** — a terminal transcript showing the client's interactive login: run the
   login subcommand, get prompted for the token, get a success line. The transcript block
   has its own copy button.
4. A pointer into the documentation.

### 2.3 What makes the first tunnel fast, and where friction is front-loaded

- **B is the fastest by an order of magnitude** and the only one where time-to-first-URL
  is bounded by typing speed: no account, no download, no token, one command using
  software the user already has. Its friction is all *back*-loaded: the tunnel dies on a
  timer and the hostname changes on every reconnect.
- **A is fast conditional on signup**: OAuth, no email verification, and the quickstart is
  the landing page with the token pre-substituted. Its friction is *breadth* — the
  platform picker presents 14 options before the user has done anything, and TCP
  endpoints additionally require a payment method on file (see §7.5).
- **C is the slowest**: account + mandatory email verification + binary download +
  architecture selection + a second in-client login before any tunnel exists. Seven
  documented steps. In exchange the user gets a persistent local GUI that manages many
  tunnels.

**Design conclusion for `pumasi-tunnel`:** the SSH-first anonymous path is the single
highest-leverage onboarding behaviour in this category and should be the default.
Everything else (agent binary, GUI, dashboard) is an optional upgrade path layered on
top of a working, no-account first URL.

---

## 3. The CLI / agent experience

### 3.1 The invocation shapes observed

Three distinct client models:

**Named-agent model (A).** A standalone executable configured once (`<agent> config
add-authtoken <token>`, writing into a YAML config file with a schema version and an
`agent.authtoken` key) and then run per-tunnel with a forwarding subcommand naming the
protocol and the local port, plus flags for a reserved hostname and an inline policy
document. Environment-variable configuration is offered as an equal alternative and shown
as a copyable `export` line. The config file also supports declaring multiple named
tunnels started together.

**Bare-SSH model (B).** No client at all:

```
ssh -p <port> -R0:<local-host>:<local-port> [-o KeepAlive/StrictHostKeyChecking opts] <options>@<edge-host>
```

The critical mechanic: **all configuration travels in the SSH username field**, with
option tokens joined by a separator. An anonymous tunnel uses a bare mode word; an
authenticated tunnel uses the access token; extra behaviours (QR output, forced-new-tunnel,
protocol selection) are appended as additional tokens in the same string. This is why the
product needs no binary and why its dashboard's job is string assembly. A first-party CLI
and a Docker image exist as alternatives but are presented as secondary.

**Dual CLI/GUI binary (C).** One executable with an account subcommand group
(`<client> account login`, interactive token prompt) and a tunnel subcommand group
(`<client> tunnel http --to <host:port>`, with `--help` on each). It reads a
`config.yaml` for declarative multi-tunnel definitions. Double-clicking the same binary
starts a local HTTP server and opens the browser to a full GUI (§3.4).

### 3.2 What a running tunnel displays in the terminal

The richest terminal surface is B's, and it is worth copying at the *structure* level:

- A **banner block**: authentication state (authenticated as `<identity>`, or an explicit
  "not authenticated" line), and for unauthenticated sessions a **countdown of remaining
  session time** with a pointer to the upgrade path.
- The **assigned public URLs**, one line per scheme (plain and TLS).
- A **counters block**, refreshed live: bytes received, bytes sent, request count,
  response count, currently-active connections, total connections.
- A **live request log**: one line per request with method, response status, and path,
  numbered in sequence, colour-differentiated by status class.
- In its full TUI mode, a second pane showing the **selected request's raw request and
  response headers**, navigable with arrow keys, with a keybinding-help affordance.

A's terminal output was not captured directly in the tour; its equivalent surface is the
hosted inspector plus a local inspector. C's equivalent is the local GUI (§3.4).

### 3.3 SDKs, containers, orchestration

- **A** offers five language SDKs framed as "ingress as a socket" — the application
  itself listens on the tunnel rather than shelling out to an agent. Each SDK path in the
  quickstart offers a runnable example repo or an add-to-existing-app snippet. It also
  offers a Kubernetes operator (creates and manages endpoints for in-cluster services,
  and can bind an endpoint so it only accepts traffic originating inside the cluster) and
  an infrastructure-as-code provider (endpoints declared as resources with an inline
  policy document). Docker and Docker Desktop are separate first-class install targets,
  as is a single-board-computer target.
- **B** offers Node and Python SDKs and a Docker image, presented as documentation tabs
  rather than as separate onboarding paths.
- **C** offers Docker and a documented `config.yaml`; no language SDKs observed.

### 3.4 The client-side GUI (C only) — the most distinctive artifact in the tour

Served by the client binary on the local machine and opened in the user's browser. A
three-pane workspace:

1. **Tunnel table** (left). One row per configured tunnel, with columns: user-assigned
   **name**; **type** badge (HTTP / TLS / TCP / UDP); **region** badge; the **public
   address stacked over the local target**; **status** badge (running / stopped); a
   per-row **inspect toggle switch**; and two row-action icon buttons (edit-or-stop, and
   delete). Independently paginated.
2. **Request log** (middle). One entry per captured request: method badge, path, status
   badge coloured by class, the host the request arrived on, and a relative timestamp.
   Independently paginated.
3. **Request/response detail** (right). Line-numbered raw headers, split into a request
   section and a response section, the response section's header bar carrying the status
   code. Horizontally scrollable.

A status bar reports **connection state to the edge** and the **client version**.

**Create-new-tunnel modal** (in the GUI):
- **Type** — select: HTTP / TLS / TCP / UDP *(required)*
- **Region** — select *(required)*
- **Enable inspect mode** — toggle
- A source/port field (partially occluded in the capture — see §10)
- **To address** — the local target, e.g. `localhost:8080`
- Checkbox: *my local server already terminates TLS* — reveals a second address field and
  explains that supplying an HTTPS local address makes the edge **pass TLS through
  without terminating it**
- **Domain** section — checkbox: *use a reserved subdomain/domain*

That modal is the cleanest, most copyable expression of the tunnel object in the whole
tour and should be the model for `pumasi-tunnel`'s create form.

---

## 4. Dashboard information architecture

### 4.1 Incumbent A — grouped left rail, plus a separate settings rail

A workspace/product selector sits above the rail; a global search item sits below it.
Groups, in order:

- **Getting started:** Quickstart · Share Localhost · Your Authtoken
- **Connectivity:** Endpoints · Agents · Kubernetes Operators
- **Traffic:** Traffic Inspector · Traffic Identities · Log Exporting
- **Network:** Domains · TCP Addresses
- **Resources:** Vaults & Secrets · IP Policies · TLS Certificates · Certificate
  Authorities · Connect URLs
- **Ungrouped, below:** Usage & Limits · Identity & Access · Settings · Help (expandable)
- **Rail footer:** signed-in identity + user menu

Entering settings **replaces the entire rail** with a settings rail (with a back
affordance) and a breadcrumb in the header:

- **Account:** General · Billing · Auth · Data Retention · IP Restrictions
- **Identity & Access:** Team Members · Service Users · Authtokens · API Keys ·
  SSH Public Keys
- **Personal:** Profile · Preferences · Security & Access · Accounts

The presence of *SSH Public Keys* under identity, and of "reverse SSH agents" in the
agents definition, confirms A also accepts a bare-SSH client path even though it is not
promoted.

### 4.2 Incumbent B — flat left rail, tabs inside settings

Flat, ungrouped: Quickstart · Configure Tunnel · Domains · Active Tunnels ·
Manage Tokens · Remote Devices · API Keys · Teams · Subscriptions · Request Feature ·
Help. The header carries an **account/workspace selector**, a light/dark toggle, and a
user menu whose only items are Settings and Log out. Settings is a single page with three
tabs: Active Sessions · Two-factor authentication · Reset password.

Paywalled items **remain visible and clickable** in the rail; clicking one does not
navigate but raises an upgrade dialog naming the specific capability (§8.3).

### 4.3 Incumbent C — short flat rail with a utility footer

Rail: Tunnels · Domains · Endpoints · Access · Billing · Settings. Footer of the rail:
Documentation · light/dark toggle · collapse-rail · account row showing the signed-in
identity, a presence dot, an **Upgrade** link, and a sign-out control. Settings is one
page with two tabs: General · Billing (Billing is reachable both from the rail and as
that tab).

### 4.4 Account-level vs tunnel-level

Across all three, the split is consistent and worth preserving:

- **Account-level:** tokens/authtokens, API keys, reserved domains, reserved TCP/UDP
  addresses, IP policies, certificates, certificate authorities, vaults, log-export
  destinations, teams/members, sessions, 2FA, billing, usage & limits.
- **Session/tunnel-level:** the live tunnel row, its captured traffic, its inspect flag,
  its bound hostname, its local target, its region.
- The join between them is the **binding**: A binds an endpoint to a domain/TCP address;
  B binds a domain **to a token** (so any tunnel started with that token inherits the
  hostname); C binds a tunnel to a reservation via a checkbox in the create modal.

B's token-scoped binding is by far the simplest of the three to implement and to explain,
and it makes hostname stability a property of the credential rather than of the process.

---

## 5. Core objects and their lifecycles

### 5.1 Endpoint / tunnel

**A — Endpoint.** Empty state: an explanatory paragraph defining an endpoint as the
access point where traffic is delivered, noting that endpoints comprise domains and TCP
addresses for services delivered online or devices connected to; a primary **New
Endpoint** action and a secondary split "documentation" dropdown. Endpoints appear
automatically when an agent connects, and can also be created as *cloud endpoints* from
the dashboard (i.e. an endpoint that exists without a live agent). Not observed
populated — see §10.

**B — Active Tunnels.** A live view, not a resource. Header carries a manual **refresh**
button. Body: a status pill ("no active sessions"), a search field over URL or token, and
a table with columns *index · tunnel URL · token · active since*. No create action — a
tunnel comes into being by running the command. Whether a live row offers a
disconnect/kill action could not be determined (§10).

**C — Tunnels.** Hosted dashboard empty state: icon, "no tunnels yet", a line saying the
button will *guide you through* starting one, and a primary **Start a tunnel** action —
i.e. the button opens an instructional flow, not a create form, because tunnels are
created in the client. The real tunnel CRUD lives in the local GUI (§3.4), where a row
offers: toggle inspect, edit/stop, delete.

### 5.2 Agent / client session

**A — Agents.** Empty state defines an agent as anything holding an active session to the
edge: standalone agents, applications built with the SDKs, Kubernetes operators, and
reverse-SSH agents. Primary action **Download an agent**; secondary documentation
dropdown. A separate **Kubernetes Operators** page lists operator installations; its empty
state offers only a documentation action, because operators are registered from the
cluster, not created in the dashboard.

**B** has no agent object — the SSH session *is* the agent, surfaced as an Active Tunnels
row.

**C** surfaces client liveness only as a connection indicator in the local GUI's status
bar.

### 5.3 Reserved domain

**A — Domains.** Explains that every account receives one free development hostname that
does not accrue endpoint-hours; that paid plans allow choosing the name of a
vendor-subdomain hostname; and that bringing your own domain requires adding a CNAME
record. Toolbar: a filter input supporting a **`key:value` filter syntax**, a filter
button, an API-reference link, and a primary **New Domain**. Table columns: status dot ·
ID (truncated, with copy-to-clipboard) · domain (the auto-provisioned development
hostname carries a distinguishing badge) · description · status · created (sortable) ·
row overflow menu. Footer: page-size select and pager.

*Create:* a **right-side drawer**, not a centred modal. One field for the domain with a
search affordance whose placeholder implies both a subdomain form and a full apex form;
helper text explaining that a vendor subdomain can be reserved without charge while a
custom domain requires an upgrade, and that usage charges apply. Footer buttons are
Cancel / **Continue** — i.e. this is a multi-step flow (subsequent steps presumably
collect DNS records and certificate provisioning; not captured, §10).

The API reference exposes the reserved-domain object's field set, which is a sound model
to adopt: identifier, resource URI, created-at, description, free-form metadata, the
domain itself, region, **CNAME target**, an attached certificate reference, a certificate
management policy (issuing authority + private key type), and a certificate management
status carrying a renewal timestamp and a provisioning job with error code, message,
started-at, and retry-at.

**B — Domains.** A genuinely different model: a **two-column drag-and-drop binding
board**. Left column lists each **token** as a card with copy / reveal / edit actions.
Middle column is the *assigned domains* drop zone for the selected token, reading "drag a
domain here to assign". Right panel lists **available domains** — and on the free tier
that panel is replaced by an inline upsell line. An instructions callout explains the
drag-and-drop model and that a custom domain or persistent subdomain is edited via a
pencil affordance. There is no separate create form for a vendor subdomain; reserving is
an upgrade action.

**C — Domains.** Header carries the page title plus a **Filter** control. Empty state:
icon, "no domains yet", a line saying nothing has been reserved, and a primary **Add new
domain**. *Create:* a small centred modal with:
- checkbox **use my own custom domain**
- **Region** select *(required)*
- **Subdomain** text input with a fixed apex suffix rendered as an input add-on
  *(required)*
- **Create** button.

No DNS instructions appear at create time; presumably they follow for the custom-domain
branch (not captured, §10).

### 5.4 TCP / UDP address reservation

**A — TCP Addresses.** On the free tier the empty state is pure upsell: it explains that
a TCP address is a hostname and port under a regional apex, that **the hostname and port
cannot be chosen and are assigned at random**, and that reserved addresses require an
upgrade. The primary action is literally **Upgrade** — the create action is *replaced*,
not disabled. This is the sharpest paywall pattern in the whole tour.

**C — Endpoints.** The equivalent surface (endpoint reservations = TCP/UDP address
reservations, per the documentation's "Reservations → Domain / Endpoint" split). Header
has a Filter control; the body never finished loading during the tour, so columns, row
actions and empty state are unknown (§10).

**B** exposes persistent TCP/UDP ports only as a plan entitlement quantity, not as a
managed list in the dashboard.

### 5.5 Authtoken / access token / API key

Three distinct designs:

**A** separates them:
- **Your Authtoken** — a single-token page laid out as three cards: *config file* (a YAML
  snippet with a reveal toggle in the card header and a copy action), *environment
  variable* (an export line with the same reveal + copy), and *reset your authtoken* — a
  destructive card warning that the action cannot be undone and that every agent, SDK
  and deployment must be updated afterwards.
- **Authtokens** (in settings) — the plural list, for multiple agent credentials.
- **API Keys** (in settings) — separate credentials for the REST API. The Kubernetes path
  needs both an API key and an authtoken.
- **SSH Public Keys** (in settings) — for the reverse-SSH client path.
- **Service Users** (in settings) — non-human principals.

**B — Manage Tokens.** A full data grid: a toolbar of icon actions (export, search,
filter, column/density control, fullscreen); columns *index · token (masked, with a
reveal eye toggle; sortable; per-column overflow menu) · token name (inline-editable via
a pencil) · plan (the tier the token is bound to) · last-updated timestamp*; per-row a
prominent **Regenerate** button and a row overflow menu; footer with rows-per-page,
result range, and pager. Note the token carries a *plan* — entitlements are attached to
the credential, not only to the account. **API Keys** is a separate nav item.

**C — Access.** A single access token, presented mid-installer as a wide masked field
with copy and a destructively styled regenerate button.

### 5.6 Teams and members

**A — Team Members.** On the free tier the page is entirely an upsell: it explains that
adding team members requires a paid plan, notes that dashboard SSO and directory
provisioning are separately purchasable add-ons, and offers **Upgrade** plus a guide
link. There is also an **Accounts** page (a distinct concept from teams): a list of the
identities' accounts, each row showing the account name, a plan badge, the subscription
name and a member count; plus a **New Account** action; selecting an account reloads the
dashboard into it.

**B — Teams.** Header has a primary **Create New Team**. Empty state is a single card
with two stacked statements distinguishing *teams you own* from *teams you are a member
of*. Team seats are metered on the paid plan; whether creating a team on the free tier is
blocked by the upgrade dialog could not be determined (§10). The header also carries a
workspace selector for switching between personal and team context.

**C** shows no teams surface at all — single-user product, with plan seats sold as a
quantity in billing.

---

## 6. Traffic inspection & debugging

### 6.1 Hosted inspector (A)

- Page description states there is a **full-capture mode** that unlocks complete request
  and response bodies and **request replay**.
- A prominent retention banner: **traffic history is retained for one day**, with an
  upgrade path extending retention up to ninety days, and an inline upgrade button.
- Toolbar: a mode selector currently set to **Live** (implying a historical/range mode),
  a **pause** control, a result count, and a **Filter** button.
- Table columns: timestamp · client IP · method · path · status · an edge-error column ·
  duration · host.
- Footer: page-size select and pager. Empty body reads as a single "no data" row.

### 6.2 Identities (A)

A derived, non-creatable collection: end-user identities captured when an endpoint is
protected by an OAuth or OIDC policy action. The empty state describes showing each
identity's activity, user agent, device and IP address, and supporting **forced
re-authentication by revoking their sessions**. Only a documentation action — nothing is
created here.

### 6.3 Log export (A)

Empty state describes streaming both traffic logs and audit logs to third-party
destinations across four families (object storage, hosted log services, streaming
platforms, APM vendors). Primary **New log export** plus documentation. Free-tier quota
observed as two event destinations (§8.1).

### 6.4 Local inspector (B)

A local web debugger on a fixed localhost port, enabled by a toggle that adds an option
to the SSH username. Plus the terminal TUI described in §3.2, whose right pane shows raw
request and response headers for the selected request. The marketing surface announces a
**browser-based debugger being folded into the hosted dashboard** with inspect, modify
and replay — so the category is converging on a hosted inspector.

### 6.5 Local inspector (C)

The GUI's middle and right panes (§3.4), with a per-tunnel **inspect toggle**. Claims
real-time inspection of **HTTP, TCP and UDP** traffic and **payload replay**, framed
around webhook debugging: explore requests, review responses, replay payloads.

### 6.6 Synthesis for `pumasi-tunnel`

The three converge on one feature set: a live request list; a request/response detail
view with raw headers and bodies; replay; and filtering. Retention is the only axis any
of them meters. The simplest correct build is: capture in the agent (which already has
the bytes), expose a local inspector immediately, and stream a bounded envelope
(metadata + optionally bodies under a size cap) to the control plane for the hosted view.

---

## 7. Security & access control surfaces

### 7.1 IP allow/deny

**A — IP Policies.** Reusable named groups of allow and deny rules over IP addresses and
CIDR ranges. Two enforcement points are described: attached to an individual endpoint via
a policy action, or applied account-wide as **IP restrictions** governing agent
connections, API callers, and dashboard sessions. Primary **Create IP policy** plus an
API-reference link. Available on the free tier.

**B** offers IP whitelisting as a per-tunnel option in the command builder.
**C** lists IP whitelisting as a core capability available on *all* plans.

### 7.2 TLS certificates

**A — TLS Certificates.** Certificates are automatically provisioned and renewed from an
ACME authority by default; **uploading your own certificate requires a paid plan**, and
the free-tier page's primary action is Upgrade.

**B** advertises built-in ACME certificates for HTTPS tunnels including on custom domains
and wildcard domains.

**C** advertises automatic ACME certificates issued on the fly for custom domains, and
lists SSL/TLS encryption even on the free tier.

### 7.3 Certificate authorities / mutual TLS (A only)

A dedicated page for uploading CAs to enforce mutual TLS on endpoints, attached via a
TLS-termination policy action. The page states plainly that the vendor **does not manage
the CA PKI** — the customer owns the CA private key and issues client certificates. A
create action is available on the free tier. This is enterprise-shaped but cheap: it is
an upload-and-reference list.

### 7.4 Secrets (A only)

**Vaults & Secrets.** Vaults hold credentials that policies reference dynamically at
request-evaluation time; the description stresses that values are never stored in
plaintext in the policy and that updating one secret propagates everywhere it is
referenced. Primary **New vault**. Free tier caps secrets-per-vault at five.

### 7.5 Identity verification as an anti-abuse gate (A) — important

Under account settings, an **identity verification** section states that using **TCP
endpoints requires a valid payment method on file**, framed explicitly as abuse
prevention, with reassurance that the card will not be charged and free use continues.
This is a distinct mechanism from the paywall: a capability is gated on *proof of
identity*, not on payment. Any commons product offering raw TCP will face the same abuse
pressure and needs an equivalent (non-payment) answer.

### 7.6 Account security

**A:** settings sections for Auth, Data Retention, IP Restrictions, Security & Access,
Service Users, SSH public keys; SSO and directory provisioning sold as add-ons on top of
a paid plan. General settings carries an account-name field with an update button and a
**danger zone** with a delete-account action warning it removes the account for everyone
on its team.

**B:** a three-tab settings page — **Active Sessions** (table of *index · device · IP
address · last accessed · created*, with the current device marked; whether other rows
offer revoke is unknown, §10); **2FA** (a vertical three-step stepper: send an email
verification code → enter that code → scan a QR and enter the authenticator code, i.e.
email confirmation gates TOTP enrolment); **Reset password**.

**C:** settings General page in a label-left/control-right layout — email address (shown
with a change action), password change (current / new / confirm, each with a reveal
toggle, plus a forgot-password link), and a **danger zone** requiring the user to type a
confirmation word into a field before the destructive delete action is accepted.

### 7.7 Which surfaces read as enterprise-only

Enterprise-shaped and safe to defer: certificate authorities / mTLS, vaults & secrets,
log export to third-party destinations, traffic identities, agent connect URLs, service
users, SSO/SCIM, data-residency controls, audit logs, RBAC. A's marketing surface carries
the full attestation checklist (third-party security attestation, healthcare BAA,
European and Californian privacy regimes, a transatlantic data-transfer framework) — all
of that is `ops-heavy` and belongs in a care register, not in v1.

---

## 8. Plans, metering & limits — the paywall evidence

This is the section the commons cares most about. Everything here is stated in
**mechanism** terms.

### 8.1 Incumbent A — quota table + capability removal + retention

A dedicated **Usage & Limits** page with: a header row offering *go to billing*, a
grouping selector ("usage by month"), and a billing-period selector badged with the
current plan; a **usage summary** card for the period with a progress bar and a
"view all usage" link; and a **limits table** with columns *usage type · limit · action*
where **every single row's action is an Upgrade link**. Free-tier values observed:

| Usage type | Free limit |
|---|---|
| Online endpoints | 3 |
| Concurrent agents | 3 |
| Tunnels per session | 5 |
| Reservable domains | **0** |
| Wildcard domains | **0** |
| TCP addresses | **0** |
| Data transfer out per month | **5 GB** |
| HTTP requests per minute | 4 000 |
| API keys | 5 |
| Secrets per vault | 5 |
| Event (log) destinations | 2 |
| Event destinations per subscription | 2 |

Below the limits table, an **endpoint reporting** card: a per-endpoint usage table with
columns *URL · active hours (with an explanatory tooltip) · a vendor-specific
transfer-unit column · requests · connections · transfer out · last activity*, an empty
message scoped to the selected period, and a link to the full endpoints view.

Beyond quotas, A removes capabilities outright on free:
- **Reserved TCP addresses:** unavailable; addresses are randomly assigned per session,
  so a TCP consumer's host:port changes on every reconnect. Create action replaced by
  Upgrade.
- **Custom / bring-your-own domains and wildcard domains:** unavailable. A single
  auto-provisioned development hostname is granted (a multi-word random label under a
  free-tier apex) and it is the only stable name.
- **Uploading your own TLS certificate:** paid only.
- **Team members:** paid only; SSO and directory provisioning are paid add-ons on top.
- **Custom agent connect URL:** explicitly named as a paid-plan feature, with the create
  action replaced by an upgrade action. Its stated purpose is a branded connect address
  *and* the ability to block the default connect address on a corporate network so agents
  cannot escape the organisation's account.
- **Traffic inspection retention: one day free, extendable to ninety days on a paid
  plan.** Full-capture mode (complete bodies) and request replay are tied to that.
- **TCP endpoints require a payment method on file** even while remaining free (§7.5).

Upgrade affordances are placed *at the point of refusal* — in the empty state of the very
page the user navigated to, and as a per-row link in the limits table.

### 8.2 Incumbent B — session timeout and hostname churn are the entire paywall

Two plan surfaces exist and they disagree slightly (§10). Common to both:

**Free (perpetual, $0):**
- single-command tunnelling
- HTTP(S), TCP, TLS and (per the marketing surface) UDP tunnels
- live header manipulation
- request/response inspection **and replay**
- **60-minute tunnel timeout** — the session is terminated on a timer, with a countdown
  printed in the terminal banner
- **random subdomains** — a fresh hostname on every connect, so no link survives a restart
- marketing surface says *unlimited data transfer*; the in-app plan table says
  *restricted bandwidth and connections* (§10)

**Paid (low single-digit dollars per seat per month, annual billing discounted ~17%, with
a monthly/yearly toggle and a per-seat stepper):**
- everything in free, plus **unlimited tunnel duration**
- **1 persistent tunnel**, **1 custom subdomain**, **1 custom domain**,
  **1 persistent TCP/UDP port**, **1 team** — quantities, presumably scaling per seat
- wildcard domain support
- remote-device management
- priority support

**Enterprise (contact sales):** dedicated or on-premises servers; unlimited persistent
tunnels, custom subdomains, custom domains, persistent TCP/UDP ports, and teams.

The free tier is remarkably generous on *capability* (all four protocols, inspection,
replay, header rewriting) and brutal on *continuity*. The upsell wording is entirely
mechanism-shaped: "60 minutes tunnel timeout", "random subdomains", "unlimited tunnel
duration", "1 persistent tunnel". This is the cleanest paywall in the category and the
one a commons product most directly negates by simply not having a timer.

A secondary conversion device sits on the marketing hero: an email capture offering a
**7-day free trial that grants a persistent URL and custom-domain support** — i.e. the
trial is sold on exactly the two mechanisms the free tier withholds.

### 8.3 Incumbent B's in-app upsell mechanics

Paywalled nav items **stay visible and enabled**. Activating one does not navigate;
instead a small centred dialog appears: a title, one sentence naming the specific gated
capability, and two actions — *maybe later* / *upgrade now*. Gated surfaces observed to
behave this way: remote devices; persistent domains (the domains board's available-domains
panel is swapped for an inline upsell line).

The billing page itself offers: *upgrade plan*, *manage subscription*, **refresh
subscription** (implying payment state is synchronised from an external processor and can
lag), and *bandwidth usage*; plus an active-subscriptions table with *plan · number of
seats · start date · end/renewal date*. Seats are added with a stepper directly on the
plan card.

There is also a **Request Feature** nav item opening a minimal modal: a title, one line of
prose, one required free-text field, cancel/submit.

### 8.4 Incumbent C — concurrency cap, time limits, and an interstitial page

Billing is a settings tab with two plan cards plus an all-plans card.

**Basic (current, $0, "no credit card required"):**
- **2 active HTTP/HTTPS tunnels** — a concurrency cap
- SSL/TLS encryption
- **unique subdomains** — assigned, not chosen
- **time limits** (unquantified on this surface)
- **an interstitial warning page** shown to visitors before the tunnelled app is
  displayed — a *visible degradation of the end-user experience*, not just a limit on the
  operator. This is a distinct and aggressive mechanism none of the others use.
- HTTP/HTTPS only: TCP, TLS and UDP are absent from the free list

**Pro (annual toggle, a seat/quantity control, price loading at capture time):**
- 10 active tunnels
- 10 reservations
- HTTP, HTTPS, TCP, TLS and UDP tunnelling
- TCP and UDP port forwarding
- custom subdomains and custom domains
- wildcard tunnels
- ACME certificate issuance
- 24/7 availability
- **unlimited bandwidth** (implying the free tier is metered)

**All plans (core, ungated):** basic authentication, key authentication, rate limiting,
IP whitelisting, request and response header editing, built-in file server.

C is the only one of the three that keeps the *security* features free and paywalls the
*protocol surface* — the opposite of A.

### 8.5 The category's paywall, distilled

Across all three, the free tier refuses, in mechanism terms:

1. **Hostname stability** — a random name that changes on every reconnect (all three).
2. **Session continuity** — a hard timer that kills the tunnel (B: 60 minutes; C: "time
   limits").
3. **Raw TCP/UDP with a stable address** — either absent entirely (C), or present with a
   randomly assigned host:port that cannot be reserved (A), or reservable only when paid
   (B).
4. **Custom and wildcard domains** — universally paid.
5. **Concurrency** — caps on simultaneous tunnels, agents, and endpoints (A: 3 endpoints
   / 3 agents / 5 tunnels per session; C: 2 tunnels).
6. **Bandwidth** — metered (A: 5 GB/month; C implied) or explicitly "restricted" (B's
   in-app copy).
7. **Request-rate ceiling** (A: 4 000 requests/minute).
8. **Inspection retention and depth** — short retention, no full bodies, no replay (A).
9. **Collaboration** — teams, roles, SSO, directory provisioning (A, B).
10. **Bring-your-own trust material** — your own TLS certificate (A).
11. **End-user experience** — an interstitial warning page in front of the app (C).
12. **Identity friction** — a payment method required on file to unlock raw TCP even for
    free use (A).

Items 1–7 and 11 are precisely what `pumasi-tunnel` exists to negate, and the README's
positioning already targets them. Items 8–10 are legitimate build scope; item 12 is a
real abuse problem the commons must answer some other way (see §10).

---

## 9. Prioritised build checklist for `pumasi-tunnel`

Ranked by how strongly the behaviour defines the category. Be honest about the shape of
the work: **the transport core is the hard part, and it is a small number of items. The
dashboard is almost entirely CRUD and is the large but easy part.** A team that builds
items 1–8 well has a product; a team that builds 9–30 without them has a control panel
for nothing.

### The core the product cannot exist without

| # | Behaviour | Difficulty |
|---|---|---|
| 1 | **Multiplexed reverse-stream transport.** One long-lived outbound control connection from the client carries many concurrent logical streams inbound; framing, flow control, backpressure, half-close, and clean teardown per stream. This is the product. | **hard** |
| 2 | **Edge HTTP router with TLS termination and SNI-based hostname→session dispatch**, including wildcard matching, WebSocket upgrade pass-through, and correct hop-by-hop header handling (`X-Forwarded-For`, `X-Forwarded-Proto`, real client IP). | **hard** |
| 3 | **Raw TCP listener allocation and lifecycle** — bind a port on an edge host, map it to a session, release it on disconnect, avoid collisions across a fleet, and survive edge restarts. Reserved (stable) ports raise this from bookkeeping to a distributed allocation problem. | **hard** |
| 4 | **Zero-client SSH ingress**: accept a standard `ssh -R` reverse forward, parse configuration out of the SSH username field, and bridge that channel into the same stream router as the native agent. The single highest-leverage onboarding behaviour in the category. | **hard** |
| 5 | **Automatic certificate issuance and renewal (ACME)** for vendor subdomains, wildcards, and customer-owned domains, with a persisted provisioning state machine (pending / issued / renewing / failed with error code and retry-at). | **ops-heavy** |
| 6 | **Anonymous first tunnel with no account** — a working HTTPS URL from one command, with the identity/quota model tolerating unauthenticated sessions. | **moderate** |
| 7 | **Session→hostname binding, stable across reconnects**, keyed on the credential rather than the process (the token-scoped model), so a restart returns the same URL. | **moderate** |
| 8 | **Live terminal status surface**: authentication state, assigned URLs per scheme, live counters (bytes in/out, requests, responses, active and total connections), and a streaming request log line per request. | **moderate** |

### The debugging surface that makes it a developer tool

| # | Behaviour | Difficulty |
|---|---|---|
| 9 | **Local request inspector** served by the client on a fixed localhost port: request list with method / path / status / host / duration / timestamp, plus a detail view of raw request and response headers and bodies. | **moderate** |
| 10 | **Request replay** — re-issue a captured request against the local target, from both the local inspector and the hosted one. | **moderate** |
| 11 | **Hosted traffic inspector** with a live/paused mode toggle, filtering, pagination, and a bounded retention window. Requires streaming a capture envelope from agent to control plane. | **moderate** |
| 12 | **Per-tunnel inspect toggle** rather than a global setting (capture costs bytes and privacy). | **trivial** |
| 13 | **Live "waiting for your app" detection** in the onboarding flow — poll for the first successful connection and advance the wizard automatically. | **trivial** |

### The object model and its CRUD

| # | Behaviour | Difficulty |
|---|---|---|
| 14 | **Tunnel create form** with type (HTTP/TLS/TCP/UDP), region, local target address, inspect toggle, a "my local server already terminates TLS" branch that switches the edge to pass-through, and an optional reservation binding. | **trivial** |
| 15 | **Reserved domain CRUD**: list with status, identifier with copy-to-clipboard, description, created-at, row overflow menu, page-size and pager; create as a subdomain-under-apex form or a bring-your-own-domain branch. | **trivial** |
| 16 | **Custom-domain verification flow**: show the CNAME target, poll DNS, surface a verification status, then trigger certificate issuance. | **moderate** |
| 17 | **Reserved TCP/UDP address CRUD** — list, create, release; region-scoped. | **trivial** |
| 18 | **Token management grid**: masked value with reveal, copy, inline-editable name, created/updated timestamps, **regenerate** as a first-class row action, delete, sorting, filtering, pagination. | **trivial** |
| 19 | **Separate API keys** from agent tokens, and a documented REST API over every object (verb + path, bearer auth, an API-version header, per-status response schemas). | **moderate** |
| 20 | **Active sessions / agents list** — what is connected right now, from which IP, since when, on which region, bound to which hostname; with a disconnect action. | **trivial** |
| 21 | **Domain↔token binding UI** — assign a reserved hostname to a credential so any tunnel started with it inherits the name. (The drag-and-drop presentation is not required; a select is fine.) | **trivial** |
| 22 | **Empty states that teach**: every list's zero state defines the object in one sentence and offers exactly one primary action plus a documentation link. | **trivial** |

### Access control and account surface

| # | Behaviour | Difficulty |
|---|---|---|
| 23 | **Per-tunnel access control**: HTTP basic auth, bearer/key auth, and IP allow-lists, configurable from the command builder and from the tunnel form. | **moderate** |
| 24 | **Reusable named IP policies** with allow and deny CIDR rules, attachable per endpoint and applicable account-wide to agent connections, API callers and dashboard sessions. | **moderate** |
| 25 | **Edge request-rate limiting** per endpoint. | **moderate** |
| 26 | **Header rewriting** — add, remove and modify request and response headers at the edge, and an HTTP→HTTPS redirect behaviour. | **moderate** |
| 27 | **Account settings**: change email, change password, active sessions list with revoke, TOTP two-factor enrolment gated behind an emailed confirmation code, and a type-to-confirm account deletion. | **trivial** |
| 28 | **Teams**: create, invite, roles, and a workspace switcher distinguishing owned teams from joined teams. | **trivial** |

### Onboarding and distribution

| # | Behaviour | Difficulty |
|---|---|---|
| 29 | **Command builder**: a form (protocol, local address, region, credential, shell/platform, plus grouped option toggles such as auto-reconnect, keep-alive, force-new-session) whose output is a single copyable command that updates on every change, with copy and download-as-script actions and named saved configurations. Extremely high perceived value for very little code — build this early. | **trivial** |
| 30 | **Platform-matrixed install page**: OS row → package-manager tabs → per-distribution artefacts with an architecture selector, plus container and orchestration paths. Plus signed release artefacts and package-manager publication. | **ops-heavy** |
| 31 | **Language SDKs** ("ingress as a socket") so an application can listen on the tunnel directly, and a container/orchestration integration. | **moderate** |
| 32 | **Usage and limits page** — per-metric current-vs-limit table and per-endpoint usage reporting (active hours, requests, connections, transfer out, last activity). Needed for honesty about capacity even in a product with no paywall. | **trivial** |
| 33 | **Abuse controls** — the one thing the incumbents solve with a credit card. Rate limits, phishing/malware interstitials on report, abuse reporting, and fast hostname revocation. | **ops-heavy** |

**Honest summary of shape:** items 1–5 are where essentially all the engineering risk
lives, and item 33 is where all the operational risk lives. Items 14–22 and 27–29 are
roughly two thirds of the visible product and are ordinary forms and tables. The commons'
strength — specifiable pure logic — maps best onto items 6–12, 16, 23–26 and 29.

---

## 10. Open questions the screenshots could not settle

**Populated states.** Every list in every product was captured empty or loading. Unknown:
what a populated endpoint/tunnel row shows, what its row overflow menu contains, and what
a live tunnel's detail page looks like in the hosted dashboards.

**Incumbent A**
- The multi-step **new-domain drawer** advanced past step one was not captured — the DNS
  record display, verification polling and certificate provisioning UI are inferred from
  the API object shape, not observed.
- The **cloud endpoint** creation flow (an endpoint with no live agent) was not captured.
- The agent's **own terminal output** was never captured; only the hosted views.
- Whether the free plan's "5 tunnels per session" and "3 online endpoints" interact (i.e.
  whether a multi-tunnel session consumes several endpoint slots) is unclear.
- The **billing page itself** was not captured — only the usage/limits page and the
  upsell surfaces. Paid-tier prices and tier names are therefore unknown from this tour.
- Rows below "data transfer out per month" in the limits table were cut off.

**Incumbent B**
- The token row's **overflow menu** was never opened — its contents beyond *regenerate*
  are unknown. (The screenshot filed as the token action menu actually captured the
  account menu.)
- Whether an **Active Tunnels row offers a disconnect/kill action** is unknown.
- Whether **Active Sessions rows other than the current device offer revoke** is unknown.
- The **Remote Devices** and **API Keys** pages were never reached (Remote Devices is
  paywalled and raised the upgrade dialog instead of navigating).
- Whether **creating a team on the free tier** succeeds or raises the upgrade dialog is
  unknown, despite the create action being visible.
- **Direct contradiction:** the marketing plan card lists *unlimited data transfer* on the
  free tier while the in-app plan table lists *restricted bandwidth and connections*. One
  of the two is stale. The in-app table is the more likely truth but this is unresolved.
- The full **options list** in the command builder was cut off after the connection group
  (auto-reconnect, keep-alive, force-new-tunnel); further groups exist but were not seen.

**Incumbent C**
- The **Endpoints** page never finished loading. Its columns, row actions and empty state
  are unknown — the only evidence that it is the TCP/UDP reservation surface is the
  documentation's Reservations → Domain / Endpoint split.
- The **Pro price and the seat control's semantics** were still loading at capture; the
  billing card's quantity control could be seats, concurrent tunnels, or reservations.
- The **create-tunnel modal's source/port field** was partially occluded by an open
  select; its exact label and whether a port is user-chosen or assigned is unknown.
- The **interstitial warning page** free-tier visitors are shown was never captured — its
  content, whether it is dismissible, and whether it appears once per visitor or per
  request are all unknown. This matters: it is the most aggressive free-tier mechanism in
  the tour.
- The free tier's **"time limits"** are unquantified anywhere in the captures.
- What the hosted **Tunnels** page shows once the local client is connected — whether
  client-side tunnels appear in the hosted dashboard at all — is unknown.

**Category-wide**
- No product's **region-selection semantics** were observable in practice: whether region
  is chosen per tunnel, inherited from the reservation, or auto-selected by latency.
  (Incumbent B offers an explicit "auto" region option; C makes region required on both
  tunnels and reservations; A attaches region to domains and TCP addresses.)
- **UDP tunnelling** is claimed by all three but no UDP flow was captured anywhere.
- No **pricing page** was captured for A or C in a resolved state, so cross-product price
  comparison is out of scope for this tour.
- **The abuse question is unanswered.** A gates raw TCP behind a payment method
  explicitly as anti-abuse. An unmetered, free, no-account tunnel service with raw TCP is
  a strictly more attractive target for phishing, C2 and content abuse than any of the
  three toured products. Nothing in this tour tells us how to solve that without a credit
  card, and it should be treated as an open design problem for `pumasi-tunnel` rather
  than an implementation detail.

**Screens not individually reviewed.** Roughly two dozen of the 121 captures are
marketing sections (testimonial walls, customer-logo strips, case studies, value-prop
grids, footers), third-party OAuth consent screens, and duplicate platform/language tabs
whose structure is identical to a sibling tab already described above. Their content is
either promotional or a repeat of a documented pattern, and nothing in them bears on the
behaviour specified here.

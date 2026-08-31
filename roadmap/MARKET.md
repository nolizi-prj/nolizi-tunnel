# MARKET — the competitors, and what is actually true about them

**Owned by the product-manager role**
([`pumasi-ops/roles/product-manager.md`](https://github.com/pumasi-ai/pumasi-ops/blob/main/roles/product-manager.md),
duty 3). First pass **2026-08-31**, at `83fd9f7`.

**Every claim about a competitor is cited or absent.** No exceptions. This
project has already published and then removed one uncited competitor claim
(`pumasi-booking` `0d1674d`), and removal — not argument — is the precedent.
Fairness rule: write about these vendors honestly enough that they could read
this file without objecting to a fact. Where a comparison goes *against* us,
it is written down rather than left out; §4 has two.

**Why this file exists now.** [`VALUE.md`](VALUE.md) opens by comparing this
product to "the alternatives", and its statement of the pain — a hosted tunnel
"whose free tier makes the useful parts the paid parts" — is a claim about
other people's pricing. It was carried uncited. Either it gets sources or it
goes; this file is the sources. It closes [`BACKLOG.md`](BACKLOG.md) item 10.

**The three comparators** are the ones this product was built against:
`docs/ux/incumbent-ux-spec.md` (tour date 2026-08-30) names **ngrok**,
**Pinggy** and **LocalXpose**. `pumasi/catalog.json` names none, because it has
no entry for this product at all — see [`STAGE.md`](STAGE.md), known gaps, and
`pumasi/DECISIONS.md` **Q-019**.

---

## 1 · Published pricing

Read from each vendor's own public page on **2026-08-31**. Prices move; the
date is part of the claim. A reader who finds different numbers should update
this file rather than argue with it.

### ngrok
Source: <https://ngrok.com/pricing>, fetched 2026-08-31.

| Plan as shown | Price as shown | Raw TCP as shown | Domains as shown |
| :--- | :--- | :--- | :--- |
| Free | $0/month | "Randomly assigned TCP addresses are supported with credit card verification" | 1 development domain (ngrok-branded) |
| Hobbyist | $10/month | 1 reserved address included | 1 development + 10 ngrok-branded |
| Pay-as-you-go | $20/month + usage | 100 reserved addresses included | 1 development; custom domains at $0.01 per active hour |
| Enterprise | not shown — contact sales | — | — |

An account is required on every tier. **The page states no session or tunnel
duration limit for any tier** — so this file makes no claim about ngrok
timeouts, in either direction.

### Pinggy
Source: <https://pinggy.io/#pricing>, fetched 2026-08-31.

| Plan as shown | Price as shown | Free-tier limit as shown | TCP as shown |
| :--- | :--- | :--- | :--- |
| Free | $0/month | "60 minutes tunnel timeout"; random subdomains only | included |
| Pro | price not printed on this page | persistent tunnels; custom subdomains and custom domains | included |
| Enterprise | "Custom" | dedicated servers / on-premise | included |

**No account is required to start on the free tier**, as printed.

The Pro price is not on that page. Pinggy's own comparison page
(<https://pinggy.io/compare/pinggy-vs-ngrok/>, fetched 2026-08-31) prints
**"$3.00"** for "Pinggy (Pro)" against **"$10.00"** for ngrok. That is the
figure, and its source, and nothing further about Pinggy's annual rate is
established here because no vendor page consulted on this date printed one.

### LocalXpose
Source: <https://localxpose.io/pricing>, fetched 2026-08-31.

| Plan as shown | Price as shown | Concurrency as shown | TCP/UDP as shown | Free-tier restrictions as shown |
| :--- | :--- | :--- | :--- | :--- |
| Starter | $0/month | 2 active HTTP/HTTPS tunnels | not included | "Time limits"; "Interstitial warning page"; no custom domains |
| PRO | $8/month, or $96/year billed annually | 10 tunnels | "HTTP, HTTPS, TCP, TLS, UDP tunneling" | custom domains and wildcard tunnels included |
| Enterprise | "Let's Talk" | not shown | — | — |

### What this establishes, and what it does not

**Establishes**, on 2026-08-31, from the vendors' own pages:

- Raw TCP is behind a payment or a payment instrument at **two of the three**.
  ngrok's free tier supports TCP only "with credit card verification", and a
  *reserved* TCP address starts at the $10/month Hobbyist plan. LocalXpose's
  free Starter tier does not include TCP or UDP at all; they start at $8/month.
- A stable hostname is a paid capability at **all three**. Pinggy's free tier
  gives random subdomains and sells custom ones on Pro; LocalXpose excludes
  custom domains from Starter; ngrok's free domain is ngrok-branded and custom
  domains are metered at $0.01 per active hour on Pay-as-you-go.
- A free-tier session ceiling is printed by **two of the three**: Pinggy's
  "60 minutes tunnel timeout" and LocalXpose's "Time limits". ngrok's page
  prints none.
- An interstitial warning page is printed by **one**: LocalXpose Starter.

**Does not establish:** anything about discounts, trials, enterprise contracts,
regional pricing, or what any given buyer actually pays. Nothing here should be
restated as "ngrok costs $10" without the plan name attached, and nothing here
is a claim about reliability, performance, support or security posture — none
of which this file measured.

---

## 2 · How they are delivered, and how that differs from us

From `docs/ux/incumbent-ux-spec.md` §1 (clean-room tour, 2026-08-30) together
with the pricing pages above.

| | ngrok | Pinggy | LocalXpose | **pumasi-tunnel** |
| :--- | :--- | :--- | :--- | :--- |
| Client to start | own agent binary | **the OS `ssh` client** | own binary (CLI + local GUI) | **stock `ssh`**, or one static binary |
| Account to start | yes | **no** (free tier) | yes | **no** |
| Free-tier session ceiling | none printed | 60 minutes | "Time limits" | **none** — see the evidence in [`VALUE.md`](VALUE.md) claim 4 |
| Raw TCP on the free path | card verification | included | not included | **included** |
| Self-hostable relay | not offered publicly | on-premise on Enterprise | not offered publicly | **yes, Apache-2.0, same repository** |

---

## 3 · The wedge, stated only as wide as the citations allow

**Two claims, both narrow on purpose.**

1. **Raw TCP and a session that does not end are together free here, and are
   not together free anywhere above.** Pinggy gives free TCP but times the
   session out at 60 minutes; LocalXpose gives untimed sessions only from
   $8/month and TCP only from $8/month; ngrok wants a card on file for free TCP
   and $10/month for a reserved address. This product gives an untimed raw TCP
   tunnel with no account and no card — evidenced by the one that has carried
   this machine's own sshd for over ten hours ([`VALUE.md`](VALUE.md) claim 3).

2. **The relay is the product, and you may run it.** Apache-2.0, one Go module,
   no dependency outside the standard library and `golang.org/x/crypto`. None
   of the three publishes a self-hostable relay on a free or low tier; the one
   on-premise option printed above is Pinggy's Enterprise plan. Nothing here
   claims the incumbents are closed-source generally — only that no page
   consulted on this date offered a relay you may run yourself below Enterprise.

---

## 4 · Where the comparison goes against us

Written down because a market file that only flatters its own product is not
evidence, it is copy.

- **They have TLS and we do not.** All three sell HTTPS URLs as the ordinary
  case. `pumasi.link` has never listened on 443, every tunnel here is plaintext
  HTTP, and the fix for that is an operator action nobody has taken —
  [`BACKLOG.md`](BACKLOG.md) item 1. Any comparison that omits this is dishonest:
  a webhook sender that requires `https://` can use all three of them and none
  of us.
- **They have accounts and ownership; we have neither.** A stable hostname
  being *paid* at all three is only a wedge if ours is *owned*. It is not —
  `AllowAll` is the only authenticator this relay can run and `Tunnel.Reserved`
  is never read ([`BACKLOG.md`](BACKLOG.md) item 3). Today "free stable name"
  means "free unclaimed name", and that is a weaker product, not a cheaper one.
  Until item 3 lands, this file's §3 claim 1 is about *price*, never about
  *reliability*.
- **One relay, one host.** Every vendor above runs an edge; we run one $5–6/month
  machine in Chicago (`pumasi-ops/RESOURCES.md` §3). Nothing in §3 should be read
  as an availability claim.

---

## 5 · What would make this file wrong

- Any of the four pricing pages changing — most likely, and the reason each row
  carries its fetch date.
- A vendor adding a free untimed raw-TCP tier, which would retire §3 claim 1.
- This product acquiring the gaps in §4, which would let §3 be stated wider.
- A comparator being added or dropped in `pumasi/catalog.json` once this product
  has an entry there at all (**Q-019**).

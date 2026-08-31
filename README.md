# Pumasi Tunnel

**Unmetered, multi-protocol localhost tunnels with custom domains, raw TCP port exposure, and zero-client SSH access — Ngrok, Pinggy, and Playit.gg, copied.**

Part of [Pumasi](https://github.com/pumasi-ai/pumasi), a commons of software built by agents and governed by people. Apache-2.0, inbound equals outbound.

---

## The One-Line Quickstart

### Method 1: Zero Software Installed (Native SSH)
Expose any local port to the internet instantly over standard SSH:
```bash
ssh -R 80:localhost:3000 tunnel.pumasi.ai
```
Outputs an instant, public HTTPS URL: `https://xyz.pumasi.link`

### Method 2: Dedicated CLI (`pumasi-tunnel`)
```bash
# Forward a web app (HTTP/HTTPS)
pumasi tunnel 8080

# Forward raw TCP (Windows Remote Desktop, PostgreSQL, SSH)
pumasi tunnel tcp 3389

# Bind to a permanent custom subdomain
pumasi tunnel 8080 --subdomain myapi
```

---

## Why Pumasi Tunnel Exists

| Feature | Ngrok (Free / Paid) | Pinggy.io | Cloudflare Tunnel | **Pumasi Tunnel** |
| :--- | :--- | :--- | :--- | :--- |
| **Pricing** | $10 – $18/seat/mo | $2.50 – $12/mo | Free (Web only) | **100% Free / Apache-2.0** |
| **Custom Subdomains** | Paywalled ($10+/mo) | Paywalled ($2.50+/mo) | Free | **Free & Permanent** |
| **Raw TCP Forwarding (RDP, DBs)** | Paywalled | Paywalled | ❌ Requires Client Helper | **✅ Native TCP Support** |
| **Session Disconnect Timers** | None | 60-Minute Hard Cutoff | None | **None (Unlimited)** |
| **Bandwidth Limits** | 1GB – 15GB/mo (+$0.10/GB) | Metered | Unlimited | **Unmetered** |
| **HTML Warning Interstitial** | ❌ Breaks API/Webhooks | ❌ Interstitial Screen | None | **None (Direct 200 OK)** |
| **Local Webhook Inspector** | Port 4040 | Web dashboard | None | **Built-in (Port 4040)** |
| **Zero-Client SSH Tunneling** | ❌ (Client required) | ✅ | ❌ (Client required) | **✅ (Standard `ssh -R`)** |

---

## Architecture

```
[Outside World (Browser / Webhook / mstsc)]
                     │
                     ▼ (e.g. s.pumasi.ai:3389 or dev.pumasi.link:443)
┌───────────────────────────────────────────────────────────┐
│  PUMASI RELAY SERVER (Edge VPS / Cloudflare Network)     │
│  - HTTP/HTTPS Host Router (Wildcard Let's Encrypt / TLS) │
│  - Raw TCP Port Pool Allocator                           │
│  - SSH Ingress Gateway (gliderlabs/ssh on port 22/443)   │
│  - QUIC / Yamux Stream Multiplexer                       │
└─────────────────────────────┬─────────────────────────────┘
                              │ Outbound TLS/QUIC Connection (Port 443)
                              ▼
┌───────────────────────────────────────────────────────────┐
│  PUMASI CLIENT AGENT (Local Machine / CLI)                │
│  - Bypasses NAT / Router Firewalls                        │
│  - Forwards incoming frames to localhost:port             │
│  - Embedded Webhook & Request Inspector (localhost:4040)  │
└───────────────────────────────────────────────────────────┘
```

---

## License

Apache-2.0. Inbound equals outbound.

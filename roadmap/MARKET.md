# Product research

Reviewed 2026-09-02:

- [ngrok CLI](https://ngrok.com/docs/agent/cli): concise HTTP/TCP start commands,
  TLS control transport, diagnostics, and policy attachment.
- [ngrok Traffic Inspector](https://ngrok.com/docs/obs/traffic-inspection):
  metadata-first inspection, opt-in full capture, retention, search, and replay.
- [ngrok API](https://ngrok.com/docs/api): authenticated, versioned HTTPS API
  with explicit rate-limit and error behavior.
- [frp documentation](https://gofrp.org/en/docs/): HTTP, HTTPS, TCP, UDP,
  dashboards, authentication, TLS, and multiplexing in a self-hosted system.
- [zrok documentation](https://docs.zrok.io/): public/private sharing,
  zero-trust concepts, and self-hosted operation.

Local source reviews also covered frp, zrok, rathole, Pangolin, and cloudflared
under `/home/m/dev/tunnel_sites/`. Screenshot research covered the ngrok public
site and dashboard under `/home/m/dev/pumasi-site-screenshot/ngrok_pages/`.

The research informs workflows and edge cases; Pumasi’s code and UI remain an
original implementation. Pricing is intentionally omitted because it changes
quickly and is not needed for the engineering roadmap.

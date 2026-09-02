# Product value

Pumasi Tunnel’s Phase 1 wedge is a useful tunnel without an account, card, or
session timer:

- stock SSH for zero-install onboarding
- a static native client for reserved addresses
- HTTPS web tunnels and raw TCP
- a self-hostable Apache-2.0 relay

This is not yet a reliability or scale advantage. Pumasi currently has one
relay, no traffic inspector, and no access-policy layer. Those gaps define
Phases 2 and 3 in [`THREE_PHASE_DEVELOPMENT_PLAN.md`](THREE_PHASE_DEVELOPMENT_PLAN.md).

The durable design rule is secure and honest defaults: announce only addresses
that work, encrypt internet control traffic, expose operational status, and do
not capture request bodies unless a user explicitly enables bounded capture.

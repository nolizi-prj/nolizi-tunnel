# Pumasi Tunnel backlog

The ordered product backlog is the unfinished work in
[`THREE_PHASE_DEVELOPMENT_PLAN.md`](THREE_PHASE_DEVELOPMENT_PLAN.md).

## Next

1. Local traffic inspector with metadata-only default, bounded opt-in body
   capture, redaction, filters, and safe replay.
2. Agent status UI, config file, release artifacts, diagnostics, and update
   checks.
3. Account-backed endpoint ownership and token rotation.
4. HTTP/OIDC access protection, IP policy, header rewriting, and limits.

## Known engineering debt

- Add per-stream flow control so one stalled stream cannot delay siblings.
- Add sustained load, reconnect, and failure-injection measurements.
- Replace the broad GitHub feedback credential with a repository-scoped token.
- Add a second relay before making an availability claim.

# Pumasi Tunnel stage

**Stage:** Alpha<br>
**Release:** `0.1.8`<br>
**Verified:** 2026-09-02 on `https://pumasi.link`

Phase 1 is complete. This is still Alpha because there are no external
customers, one relay/region, and no sustained availability evidence.

## Production evidence

- apex and wildcard HTTPS use a valid Let’s Encrypt certificate
- native client verifies TLS on port `7001`; public plaintext `7000` is closed
- stock SSH ingress works on `2222`
- real local HTTP service reached through `https://phase1e2e.pumasi.link`
- persistent raw TCP tunnel reconnects on port `20000`
- reservations use `/var/lib/pumasi-relay/reservations.json`
- `/version` and `/readyz` report the current release
- browser feedback created GitHub issue #1; the test issue was then closed
- desktop and 390px browser flows passed with no console warnings/errors
- certificate renewal timer completed a successful live check

## Verification commands

```bash
go test ./...
go vet ./...
go test -race ./...
curl -fsS https://pumasi.link/version
curl -fsS https://pumasi.link/readyz
```

The next work is Phase 2 in
[`THREE_PHASE_DEVELOPMENT_PLAN.md`](THREE_PHASE_DEVELOPMENT_PLAN.md).

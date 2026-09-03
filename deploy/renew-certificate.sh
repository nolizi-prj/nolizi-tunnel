#!/usr/bin/env bash
set -euo pipefail

ACME_ROOT=/var/lib/nolizi-acme
CERT="$ACME_ROOT/certificates/tunnel.nolizi.com.crt"
KEY="$ACME_ROOT/certificates/tunnel.nolizi.com.key"
LIVE_CERT=/etc/pumasi-relay/tls/tunnel.nolizi.com.crt
LIVE_KEY=/etc/pumasi-relay/tls/tunnel.nolizi.com.key

set -a
. /etc/pumasi-relay/cloudflare-acme.env
set +a

/usr/local/bin/lego --accept-tos --email admin@nolizi.com \
  --path "$ACME_ROOT" --dns cloudflare \
  --domains tunnel.nolizi.com --domains '*.tunnel.nolizi.com' renew --days 30

if cmp -s "$CERT" "$LIVE_CERT" && cmp -s "$KEY" "$LIVE_KEY"; then
  exit 0
fi

install -m 640 -o root -g pumasi "$CERT" "$LIVE_CERT"
install -m 640 -o root -g pumasi "$KEY" "$LIVE_KEY"
systemctl restart pumasi-relay

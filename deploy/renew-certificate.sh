#!/usr/bin/env bash
set -euo pipefail

ACME_ROOT=/var/lib/pumasi-acme
CERT="$ACME_ROOT/certificates/pumasi.link.crt"
KEY="$ACME_ROOT/certificates/pumasi.link.key"
LIVE_CERT=/etc/pumasi-relay/tls/pumasi.link.crt
LIVE_KEY=/etc/pumasi-relay/tls/pumasi.link.key

set -a
. /etc/pumasi-relay/cloudflare-acme.env
set +a

/usr/local/bin/lego --accept-tos --email atxapplellc@gmail.com \
  --path "$ACME_ROOT" --dns cloudflare \
  --domains pumasi.link --domains '*.pumasi.link' renew --days 30

if cmp -s "$CERT" "$LIVE_CERT" && cmp -s "$KEY" "$LIVE_KEY"; then
  exit 0
fi

install -m 640 -o root -g pumasi "$CERT" "$LIVE_CERT"
install -m 640 -o root -g pumasi "$KEY" "$LIVE_KEY"
systemctl restart pumasi-relay

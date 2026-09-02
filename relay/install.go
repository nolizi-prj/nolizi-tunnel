package relay

import (
	"fmt"
	"net/http"
)

func serveInstallScript(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/x-shellscript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	fmt.Fprintf(w, `#!/bin/sh
set -eu

version="v%s"
repo="https://github.com/pumasi-ai/pumasi-tunnel/releases/download/$version"
os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"

case "$os" in linux|darwin) ;; *) echo "Unsupported operating system: $os" >&2; exit 1 ;; esac
case "$arch" in x86_64|amd64) arch="amd64" ;; aarch64|arm64) arch="arm64" ;; *) echo "Unsupported architecture: $arch" >&2; exit 1 ;; esac

archive="pumasi-tunnel_${os}_${arch}.tar.gz"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT INT TERM

curl -fsSL "$repo/$archive" -o "$tmp/$archive"
curl -fsSL "$repo/checksums.txt" -o "$tmp/checksums.txt"
expected="$(awk -v file="$archive" '$2 == file { print $1 }' "$tmp/checksums.txt")"
[ -n "$expected" ] || { echo "No checksum published for $archive" >&2; exit 1; }
if command -v sha256sum >/dev/null 2>&1; then
  actual="$(sha256sum "$tmp/$archive" | awk '{print $1}')"
else
  actual="$(shasum -a 256 "$tmp/$archive" | awk '{print $1}')"
fi
[ "$actual" = "$expected" ] || { echo "Checksum verification failed" >&2; exit 1; }

install_dir="${PUMASI_INSTALL_DIR:-$HOME/.local/bin}"
mkdir -p "$install_dir"
tar -xzf "$tmp/$archive" -C "$install_dir" pumasi-tunnel
chmod 755 "$install_dir/pumasi-tunnel"
echo "Installed pumasi-tunnel %s to $install_dir/pumasi-tunnel"
case ":$PATH:" in *":$install_dir:"*) ;; *) echo "Add $install_dir to PATH to run it from anywhere." ;; esac
`, Version, Version)
}

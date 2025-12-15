#!/usr/bin/env bash
set -euo pipefail

echo "Installing cloudflared..."

BIN="$HOME/cloudflared"

curl -fsSL \
  https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 \
  -o "$BIN"
chmod +x "$BIN"


echo "Launching Cloudflare tunnel login..."
output="$("$BIN" tunnel login)"
cert_path="$(printf '%s\n' "$output" | grep -E '^/.*/cert\.pem$')"
echo "Cert path: $cert_path"

docker network create openserver_bridge

docker build -t openserver/cloudflared ./cloudflared

docker run openserver/cloudflared \
     --name cloudflared-container \
     --network openserver_bridge \
     --mount type=bind,source="$cert_path",target=/home/nonroot/.cloudflared/cert.pem
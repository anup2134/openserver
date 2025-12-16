#!/bin/sh
set -e

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
chmod +r "$cert_path"

docker network create openserver-cloudflared-control-bridge
docker network create openserver-control-user-bridge
docker volume create openserver-volume

docker build -t openserver/cloudflared ./cloudflared

docker run \
     --name cloudflared-container \
     --volume openserver-volume:/home/nonroot/persistent-shared \
     --network penserver-cloudflared-control-bridgee \
     --mount type=bind,source="$cert_path",target=/home/nonroot/.cloudflared/cert.pem \
     openserver/cloudflared


#!/bin/sh
set -e

echo "Installing cloudflared..."

BIN="$HOME/cloudflared"

curl -fsSL \
  https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 \
  -o "$BIN"
chmod +x "$BIN"


echo "Launching Cloudflare tunnel login..."
output="$("$BIN" tunnel login 2>&1)"
cert_path=$(echo "$output" | grep -oE '/[^ ]+/cert\.pem' || true)

if [ -z "$cert_path" ]; then
  echo "Error: Could not find cert.pem path in output."
  exit 1
fi

echo "Cert path: $cert_path"
chmod +r "$cert_path"

SUCCESS_CODE="0"
NETWORK_NAME="openserver-cloudflared-control-bridge"
status="0"
err_msg=$(docker network create "$NETWORK_NAME" 2>&1) || status=$?

if [ "$status" = "$SUCCESS_CODE" ]; then
    echo "Network created successfully."
elif echo "$err_msg" | grep -q "already exists"; then
    echo "Network already exists, continuing..." 
else
    echo "Error: Failed to create network"
    echo "$err_msg"
    exit "$status"
fi

VOLUME_NAME="openserver-volume"
status="0"
err_msg=$(docker volume create "$VOLUME_NAME" 2>&1) || status=$?

if [ "$status" = "$SUCCESS_CODE" ]; then
    echo "Volume created successfully."
else echo "Error: Failed to create volume"
    echo "$err_msg"
    exit "$status"
fi

echo "Building cloudflared image"
docker build -t cloudflared-hook -f ./cloudflared_hook/Dockerfile .

# echo "Building control-panel image"
# docker build -t control-panel -f ./control_panel/Dockerfile .

echo "Running cloudflared container"
docker run \
     --name cloudflared-container \
     --volume openserver-volume:/home/nonroot/.cloudflared \
     --network openserver-cloudflared-control-bridge \
     --mount type=bind,source="$cert_path",target=/home/nonroot/cert.pem \
     cloudflared-hook

# echo "Running control-panel container"
# docker run \
#      --name control-panel-container \
#      --volume openserver-volume:/home/nonroot/.cloudflared \
#      --network openserver-cloudflared-control-bridge \
#      --volume /var/run/docker.sock:/var/run/docker.sock \
#      control-panel

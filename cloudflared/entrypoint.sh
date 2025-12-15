#!/bin/sh
set -e

if ! cloudflared tunnel info openserver-tunnel >/dev/null 2>&1; then
  echo "Tunnel does not exist. Creating..."
  cloudflared tunnel create openserver-tunnel
fi

echo "Starting tunnel..."
exec cloudflared tunnel run openserver-tunnel

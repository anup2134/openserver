#!/bin/sh
set -e


if [ -f /home/nonroot/cert.pem ] && [ ! -f /home/nonroot/.cloudflared/cert.pem ]; then
    cp /home/nonroot/cert.pem /home/nonroot/.cloudflared/cert.pem
    chmod 600 /home/nonroot/.cloudflared/cert.pem
fi

if ! cloudflared tunnel info openserver-tunnel >/dev/null 2>&1; then
  echo "Tunnel does not exist. Creating..."
  rm /home/nonroot/.cloudflared/*.json
  cloudflared tunnel create openserver-tunnel

  tunnel_id=$(cloudflared tunnel list | awk '$2=="openserver-tunnel"{print $1}')
  if [ -z "$tunnel_id" ]; then
    echo "Error: tunnel 'openserver-tunnel' not found" >&2
    exit 1
  fi

  echo "Tunnel created."
fi

echo "Starting tunnel..."
exec cloudflared tunnel run openserver-tunnel

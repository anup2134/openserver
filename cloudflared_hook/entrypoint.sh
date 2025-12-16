#!/bin/sh
set -e

if ! cloudflared tunnel info openserver-tunnel >/dev/null 2>&1; then
  echo "Tunnel does not exist. Creating..."
  cloudflared tunnel create openserver-tunnel

  tunnel_id=$(cloudflared tunnel list --output json | jq -r '.[] | select(.name=="my-tunnel") | .id')
  if [ -z "$tunnel_id" ]; then
    echo "Error: tunnel 'openserver-tunnel' not found" >&2
    exit 1
  fi

  printf "tunnel: %s\ncredentials-file: /home/nonroot/.cloudflared/%s.json\n" \
         "$tunnel_id" "$tunnel_id" \
         > /home/nonroot/.cloudflared/config.yml
  echo "Tunnel created."
fi

echo "Starting tunnel..."
exec cloudflared tunnel run openserver-tunnel
#!/bin/sh
set -e

if ! cloudflared tunnel info openserver-tunnel >/dev/null 2>&1; then
  echo "Tunnel does not exist. Creating..."
  cloudflared tunnel create openserver-tunnel

  tunnel_id=$(cloudflared tunnel list | awk '$2=="openserver-tunnel"{print $1}')
  if [ -z "$tunnel_id" ]; then
    echo "Error: tunnel 'openserver-tunnel' not found" >&2
    exit 1
  fi

  printf "tunnel: %s\ncredentials-file: /home/nonroot/.cloudflared/%s.json\n\ningress:\n" \
         "$tunnel_id" "$tunnel_id" \
         > /home/nonroot/.cloudflared/config.yml
  echo "Tunnel created."
fi

echo "Starting tunnel..."
cloudflared tunnel run openserver-tunnel &
exec /home/nonroot/server-binary
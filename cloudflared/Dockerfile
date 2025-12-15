FROM alpine:3.20

RUN apk add --no-cache curl ca-certificates \
 && curl -fsSL https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 \
    -o /usr/bin/cloudflared \
 && chmod +x /usr/bin/cloudflared

RUN mkdir -p /home/nonroot/.cloudflared
RUN chown -R 65532:65532 /home/nonroot

USER 65532
ENV HOME=/home/nonroot

COPY entrypoint.sh /home/nonroot/entrypoint.sh
RUN chmod +x /home/nonroot/entrypoint.sh

ENTRYPOINT ["/home/nonroot/entrypoint.sh"]

# Get ZONE_ID
# curl -X GET "https://api.cloudflare.com/client/v4/zones?name=<DOMAIN_NAME>" \
#   -H "Authorization: Bearer <API_TOKEN>" \
#   -H "Content-Type: application/json"

# Delete DNS record
# curl -X DELETE "https://api.cloudflare.com/client/v4/zones/<ZONE_ID>/dns_records/<DNS_RECORD_ID>" \
#   -H "Authorization: Bearer <API_TOKEN>"


# RUN cloudflared tunnel create openserver-tunnel \
#  && cloduflared tunnel route dns openserver-tunnel openserver.<domain> \
#  && touch ~/.cloudflared/config.yml

# RUN TUNNEL_ID=$(cloudflared tunnel list --name openserver-tunnel --output json \
#   | jq -r '.[0].id')

# RUN "tunnel: openserver-tunnel \
# credentials-file: ~/.cloudflared/<tunnel-uuid>.json\

# ingress:\
# - hostname: <subdomain>.<domain>\
#    service: http://localhost:<port>\

# - service: http_status:404\

# cloudflared tunnel run openserver-tunnel"
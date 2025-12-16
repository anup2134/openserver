FROM alpine:3.20

RUN apk add --no-cache curl ca-certificates \
 && curl -fsSL https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 \
    -o /usr/bin/cloudflared \
 && chmod +x /usr/bin/cloudflared

RUN mkdir -p /home/nonroot/.cloudflared
RUN chown -R 65532:65532 /home/nonroot

COPY entrypoint.sh /home/nonroot/entrypoint.sh
RUN chmod +x /home/nonroot/entrypoint.sh

USER 65532
ENV HOME=/home/nonroot

ENTRYPOINT ["/home/nonroot/entrypoint.sh"]
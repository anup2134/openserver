FROM alpine:3.20

RUN apk add --no-cache curl ca-certificates \
 && curl -fsSL https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 \
    -o /usr/bin/cloudflared \
 && chmod +x /usr/bin/cloudflared

RUN mkdir -p /home/nonroot/.cloudflared
RUN chown -R 65532:65532 /home/nonroot

RUN touch /home/nonroot/.cloudflared/config.yml
RUN chmod 600 /home/nonroot/.cloudflared/config.yml
RUN chown 65532:65532 /home/nonroot/.cloudflared/config.yml

COPY entrypoint.sh /home/nonroot/entrypoint.sh
RUN chmod +x /home/nonroot/entrypoint.sh

USER 65532
ENV HOME=/home/nonroot

ENTRYPOINT ["/home/nonroot/entrypoint.sh"]
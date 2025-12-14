FROM alpine:3.20

RUN apk add --no-cache curl \
 && curl -fsSL https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 \
    -o /usr/bin/cloudflared \
 && chmod +x /usr/bin/cloudflared

# RUN mkdir -p /home/nonroot/.cloudflared
# COPY cert.pem /home/nonroot/.cloudflared/cert.pem
# RUN chown -R 65532:65532 /home/nonroot

USER 65532
ENV HOME=/home/nonroot

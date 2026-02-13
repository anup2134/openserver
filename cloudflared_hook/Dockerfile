# build context for this dockerfile is .. directory

FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY cloudflare_hook/ ./cloudflare_hook/
COPY utils/ ./utils/

WORKDIR /app/cloudflare_hook
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/cloudflared-hook-binary .

FROM alpine:3.20

RUN apk add --no-cache curl ca-certificates \
  && curl -fsSL https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 \
  -o /usr/bin/cloudflared \
  && chmod +x /usr/bin/cloudflared \
  && mkdir -p /home/nonroot/.cloudflared

COPY --from=builder /app/cloudflared-hook-binary /home/nonroot
RUN chown -R 65532:65532 /home/nonroot \
  && touch /home/nonroot/.cloudflared/config.yml \
  && chmod 600 /home/nonroot/.cloudflared/config.yml \
  && chown 65532:65532 /home/nonroot/.cloudflared/config.yml \
  && touch /home/nonroot/.cloudflared/Records.txt \
  && chmod 600 /home/nonroot/.cloudflared/Records.txt \
  && chown 65532:65532 /home/nonroot/.cloudflared/Records.txt

COPY entrypoint.sh /home/nonroot/entrypoint.sh
RUN chmod +x /home/nonroot/entrypoint.sh

USER 65532
ENV HOME=/home/nonroot

ENTRYPOINT ["/home/nonroot/entrypoint.sh"]

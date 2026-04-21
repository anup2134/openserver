FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:3.19

RUN apk add --no-cache git docker-cli ca-certificates

RUN adduser -D nonroot && \
  mkdir -p /home/nonroot/cloneTmp && \
  chown -R nonroot:nonroot /home/nonroot/cloneTmp

WORKDIR /home/nonroot
COPY --from=builder /app/main .

USER nonroot

EXPOSE 8080

ENTRYPOINT ["./main"]

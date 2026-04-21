FROM golang:1.25-alpine AS control_panel

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY utils/ ./utils/
COPY control_panel/ ./control_panel/

WORKDIR /app/control_panel
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/control_panel-binary .

FROM alpine:3.20

RUN apk add --no-cache git docker-cli ca-certificates

RUN adduser -D -u 65532 nonroot && \
  mkdir -p /home/nonroot/cloneTmp && \
  chown -R 65532:65532 /home/nonroot

WORKDIR /home/nonroot

COPY --from=control_panel /app/control_panel-binary .

RUN chown 65532:65532 control_panel-binary

USER 65532
ENV HOME=/home/nonroot

EXPOSE 8080

ENTRYPOINT ["./control_panel-binary"]

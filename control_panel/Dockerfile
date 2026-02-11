FROM golang:1.25-alpine as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app

FROM alpine:3.20

RUN apk add --no-cache docker-cli

# syntax=docker/dockerfile:1.7
ARG GO_VERSION=1.26
FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/ctf-api ./cmd/api

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata docker-cli

COPY --from=builder /out/ctf-api /app/ctf-api
COPY configs /app/configs
COPY migrations /app/migrations

RUN mkdir -p /app/storage/exports /app/storage/runtime

EXPOSE 8080

ENTRYPOINT ["/app/ctf-api"]

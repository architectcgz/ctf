# syntax=docker/dockerfile:1.7
ARG GO_VERSION=1.26
ARG MIGRATE_VERSION=v4.18.3
FROM golang:${GO_VERSION}-alpine AS builder
ARG MIGRATE_VERSION

WORKDIR /src

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOBIN=/out CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@${MIGRATE_VERSION}

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/ctf-api ./cmd/api

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/ctf-runtime-agent ./cmd/runtime-agent

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/ctf-awd-defense-ssh-gateway ./cmd/awd-defense-ssh-gateway

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/ctf-storage-gc ./cmd/storage-gc

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata docker-cli

COPY --from=builder /out/ctf-api /app/ctf-api
COPY --from=builder /out/ctf-runtime-agent /app/ctf-runtime-agent
COPY --from=builder /out/ctf-awd-defense-ssh-gateway /app/ctf-awd-defense-ssh-gateway
COPY --from=builder /out/ctf-storage-gc /app/ctf-storage-gc
COPY --from=builder /out/migrate /app/migrate
COPY configs /app/configs
COPY migrations /app/migrations
COPY scripts/docker-entrypoint.sh /app/docker-entrypoint.sh

RUN mkdir -p /app/storage/exports /app/storage/runtime
RUN chmod +x /app/docker-entrypoint.sh

EXPOSE 8080 2222 9443

ENTRYPOINT ["/app/docker-entrypoint.sh"]

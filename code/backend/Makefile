APP_ENV ?= dev
APP_NAME ?= ctf-api
GO ?= go

.PHONY: run build test fmt tidy infra-up infra-down

run:
	APP_ENV=$(APP_ENV) $(GO) run ./cmd/api

build:
	mkdir -p bin
	$(GO) build -o bin/$(APP_NAME) ./cmd/api

test:
	$(GO) test ./...

fmt:
	$(GO) fmt ./...

tidy:
	$(GO) mod tidy

infra-up:
	docker compose -f docker-compose.dev.yml up -d

infra-down:
	docker compose -f docker-compose.dev.yml down

infra-up-shared:
	$(MAKE) -C ../.. infra-up

infra-down-shared:
	$(MAKE) -C ../.. infra-down

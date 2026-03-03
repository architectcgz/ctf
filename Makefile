APP_ENV ?= dev
APP_NAME ?= ctf-api
GO ?= go
FRONTEND_DIR ?= frontend

.PHONY: run build test fmt tidy frontend-dev frontend-build frontend-typecheck infra-up infra-down

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

frontend-dev:
	npm --prefix $(FRONTEND_DIR) run dev

frontend-build:
	npm --prefix $(FRONTEND_DIR) run build

frontend-typecheck:
	npm --prefix $(FRONTEND_DIR) run typecheck

infra-up:
	docker compose -f docker-compose.dev.yml up -d

infra-down:
	docker compose -f docker-compose.dev.yml down

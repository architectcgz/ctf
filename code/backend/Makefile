APP_ENV ?= dev
APP_NAME ?= ctf-api
GO ?= go
MIGRATE_VERSION ?= v4.18.3
MIGRATE_TAGS ?= postgres
MIGRATE ?= $(GO) run -tags '$(MIGRATE_TAGS)' github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)
MIGRATIONS_DIR ?= ./migrations
MIGRATE_DATABASE_URL ?= postgres://postgres:postgres123456@127.0.0.1:5432/ctf?sslmode=disable

.PHONY: run build test fmt tidy infra-up infra-down migrate-up migrate-down migrate-create seed-demo-challenges import-challenge-packs

run:
	APP_ENV=$(APP_ENV) $(GO) run ./cmd/server

build:
	mkdir -p bin
	$(GO) build -o bin/$(APP_NAME) ./cmd/server

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

migrate-up:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(MIGRATE_DATABASE_URL)" up

migrate-down:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(MIGRATE_DATABASE_URL)" down 1

migrate-create:
	@test -n "$(name)" || (echo "usage: make migrate-create name=add_users_table" && exit 1)
	mkdir -p $(MIGRATIONS_DIR)
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

seed-demo-challenges:
	APP_ENV=$(APP_ENV) $(GO) run ./cmd/seed-demo-challenges

import-challenge-packs:
	APP_ENV=$(APP_ENV) $(GO) run ./cmd/import-challenge-packs

infra-up-shared:
	$(MAKE) -C ../.. infra-up

infra-down-shared:
	$(MAKE) -C ../.. infra-down

APP_NAME ?= sharetrip
BIN_DIR ?= bin
BIN ?= $(BIN_DIR)/$(APP_NAME)
MAIN_PKG ?= ./cmd/sharetrip
GO ?= /usr/local/go/bin/go
SERVER_PORT ?= :9090

COMPOSE_FILE ?= deploy/docker-compose.yml

DB_HOST ?= localhost
DB_PORT ?= 6543
DB_USER ?= postgres
DB_PASSWORD ?= admin
DB_NAME ?= share_trip
DB_SSLMODE ?= disable
DB_DSN ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

MIGRATIONS_DIR ?= migrations

.PHONY: deps fmt lint test build run up down migrate-up migrate-down migrate-status e2e check

deps:
	$(GO) mod download
	@echo "OK: go modules downloaded"
	$(GO) mod tidy
	@echo "OK: go modules tidied"

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')
	@echo "OK: go files formatted"

lint:
	golangci-lint run
	@echo "OK: lint passed"

test:
	$(GO) test ./...
	@echo "OK: tests passed"

build:
	mkdir -p $(BIN_DIR)
	@echo "OK: build directory ready"
	$(GO) build -o $(BIN) $(MAIN_PKG)
	@echo "OK: binary built at $(BIN)"

run:
	SERVER_PORT=$(SERVER_PORT) $(GO) run $(MAIN_PKG)
	@echo "OK: application stopped"

up:
	docker compose -f $(COMPOSE_FILE) up -d
	@echo "OK: docker compose services started"

down:
	docker compose -f $(COMPOSE_FILE) down
	@echo "OK: docker compose services stopped"

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres '$(DB_DSN)' up
	@echo "OK: migrations applied"

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres '$(DB_DSN)' down
	@echo "OK: migration rolled back"

migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres '$(DB_DSN)' status
	@echo "OK: migration status checked"

check: fmt lint test build
	@echo "OK: all checks passed"

e2e:
	@curl -s http://localhost:9090/api/ready
	@echo "OK: e2e ready check passed"

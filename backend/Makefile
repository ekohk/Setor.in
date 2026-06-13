.PHONY: help up down logs ps restart clean \
        migrate migrate-down migrate-status migrate-reset migrate-create \
	sqlc-gen build run dev test lint tidy \
	docker-build docker-run \
        keycloak-export keycloak-logs

# ─── Defaults ─────────
include .env
export

GOOSE_DRIVER ?= postgres
GOOSE_DBSTRING ?= $(DB_DSN)
GOOSE_MIGRATION_DIR ?= ./db/migrations

# ─── Help ─────────
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

# ─── Docker ─────────
up: ## Start all services (keycloak, postgres x2, mailhog)
	docker compose -f docker/docker-compose.yml up -d
	@echo ""
	@echo "  Postgres app    : localhost:5435"
	@echo "  Postgres KC     : localhost:5434"
	@echo "  Keycloak        : http://localhost:8090  (admin/admin)"
	@echo "  MailHog UI      : http://localhost:8025"

down: ## Stop all services
	docker compose -f docker/docker-compose.yml down

logs: ## Tail logs
	docker compose -f docker/docker-compose.yml logs -f

ps: ## Show running containers
	docker compose -f docker/docker-compose.yml ps

restart: down up

clean: ## Stop + remove volumes (DESTRUCTIVE)
	docker compose -f docker/docker-compose.yml down -v

# ─── Migrations (goose) ─────────
# Pakai env vars GOOSE_DRIVER & GOOSE_DBSTRING agar tidak ada masalah quoting
# di cmd.exe / PowerShell saat ada karakter '?' atau '&' di DSN.
migrate: ## Apply all pending migrations
	@set "GOOSE_DRIVER=$(GOOSE_DRIVER)" && set "GOOSE_DBSTRING=$(GOOSE_DBSTRING)" && goose -dir $(GOOSE_MIGRATION_DIR) up

migrate-down: ## Rollback 1 migration
	@set "GOOSE_DRIVER=$(GOOSE_DRIVER)" && set "GOOSE_DBSTRING=$(GOOSE_DBSTRING)" && goose -dir $(GOOSE_MIGRATION_DIR) down

migrate-status: ## Show migration status
	@set "GOOSE_DRIVER=$(GOOSE_DRIVER)" && set "GOOSE_DBSTRING=$(GOOSE_DBSTRING)" && goose -dir $(GOOSE_MIGRATION_DIR) status

migrate-reset: ## Rollback ALL migrations (DEV ONLY)
	@set "GOOSE_DRIVER=$(GOOSE_DRIVER)" && set "GOOSE_DBSTRING=$(GOOSE_DBSTRING)" && goose -dir $(GOOSE_MIGRATION_DIR) reset

migrate-create: ## Create new migration: make migrate-create name=create_orders
	@if "$(name)"=="" (echo Usage: make migrate-create name=create_xxx & exit 1) else goose -dir $(GOOSE_MIGRATION_DIR) -s create $(name) sql

# ─── sqlc ─────────
sqlc-gen: ## Generate Go code from SQL queries
	sqlc generate

# ─── Go ─────────
tidy: ## go mod tidy
	go mod tidy

build: ## Build binary to bin/api
	go build -o bin/api ./cmd/api

DOCKER_IMAGE ?= setorin-api:dev

docker-build: ## Build backend Docker image
	docker build -t $(DOCKER_IMAGE) -f Dockerfile .

docker-run: docker-build ## Run backend Docker image locally
	docker run --rm -p 8000:8000 --env-file .env $(DOCKER_IMAGE)

run: build ## Build & run
	./bin/api

dev: ## Hot-reload with air
	air

test: ## Run tests
	go test -race -cover ./...

lint: ## Run golangci-lint
	golangci-lint run

# ─── Keycloak ─────────
keycloak-logs: ## Tail Keycloak logs only
	docker compose -f docker/docker-compose.yml logs -f keycloak

keycloak-export: ## Export realm config (run after manual setup)
	-docker exec -it setorin-keycloak \
	  /opt/keycloak/bin/kc.sh export \
	  --realm setorin \
	  --file /tmp/realm-export.json
	docker cp setorin-keycloak:/tmp/realm-export.json ./docker/keycloak/realm-export.json
	@echo "Realm exported to docker/keycloak/realm-export.json"

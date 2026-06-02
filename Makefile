# Notly API

# Database configuration
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= postgres
DB_PASSWORD ?= password
DB_NAME ?= notly
DB_SSLMODE ?= disable

DB_URL = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)
MIGRATE = migrate -path internal/infrastructure/database/migrations -database "$(DB_URL)"
COMPOSE ?= docker compose

.PHONY: help migrate-up migrate-down migrate-create migrate-status migrate-force run build test test-cover lint docker-config docker-build docker-up docker-down docker-logs

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# === Database Migrations ===

migrate-up: ## Run all pending migrations
	$(MIGRATE) up

migrate-down: ## Rollback last migration
	$(MIGRATE) down 1

migrate-down-all: ## Rollback all migrations
	$(MIGRATE) down

migrate-status: ## Show migration status
	$(MIGRATE) version

migrate-force: ## Force migration version (usage: make migrate-force version=20)
	$(MIGRATE) force $(version)

migrate-create: ## Create new migration (usage: make migrate-create name=create_users)
	migrate create -ext sql -dir internal/infrastructure/database/migrations -seq $(name)

# === Development ===

run: ## Run the API server
	go run cmd/api/main.go

build: ## Build the API binary
	go build -o bin/api cmd/api/main.go

test: ## Run all tests
	go test -v ./...

test-cover: ## Run tests with coverage
	go test -cover ./...

lint: ## Run linter (requires golangci-lint)
	golangci-lint run

# === Docker / Deployment ===

docker-config: ## Validate docker compose configuration with the example deploy env
	$(COMPOSE) --env-file deploy/env.example config

docker-build: ## Build the API Docker image
	$(COMPOSE) --env-file deploy/env.example build api

docker-up: ## Build and start the full deployment stack
	$(COMPOSE) up -d --build

docker-down: ## Stop the deployment stack
	$(COMPOSE) down

docker-logs: ## Follow API and Nginx logs
	$(COMPOSE) logs -f api nginx

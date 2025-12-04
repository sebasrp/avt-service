.PHONY: help build test test-integration test-unit lint fmt clean run install-linter install-migrate install-goimports install-tools docker-up docker-down migrate migrate-down db-shell

# Default target
.DEFAULT_GOAL := help

# Path to tools
GOLANGCI_LINT := $(shell which golangci-lint 2>/dev/null || echo "$(HOME)/go/bin/golangci-lint")
GOIMPORTS := $(shell which goimports 2>/dev/null || echo "$(HOME)/go/bin/goimports")
MIGRATE := $(shell which migrate 2>/dev/null || echo "$(HOME)/go/bin/migrate")

# Database configuration
DATABASE_URL ?= postgres://telemetry_user:telemetry_pass@localhost:5432/telemetry_dev?sslmode=disable
MIGRATIONS_DIR := internal/database/migrations

## help: Display this help message
help:
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

## install-linter: Install golangci-lint
install-linter:
	@echo "Installing golangci-lint..."
	@which golangci-lint > /dev/null || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin latest
	@echo "golangci-lint installed successfully"

## install-migrate: Install golang-migrate tool
install-migrate:
	@echo "Installing migrate tool..."
	@which migrate > /dev/null || go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "migrate tool installed successfully"

## install-goimports: Install goimports tool
install-goimports:
	@echo "Installing goimports..."
	@which goimports > /dev/null || go install golang.org/x/tools/cmd/goimports@latest
	@echo "goimports installed successfully"

## install-tools: Install all development tools
install-tools: install-linter install-migrate install-goimports
	@echo "✓ All development tools installed"

## lint: Run linter on all Go files
lint:
	@echo "Running linter..."
	@$(GOLANGCI_LINT) run ./...
	@echo "✓ Linting passed"

## fmt: Format all Go files
fmt:
	@echo "Formatting Go files..."
	@gofmt -s -w .
	@$(GOIMPORTS) -w -local github.com/sebasr/avt-service .
	@echo "✓ Formatting complete"

## test: Run all tests (unit + integration, requires Docker)
test:
	@echo "Running all tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "✓ All tests passed"

## test-coverage: Run tests with coverage report
test-coverage: test
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

## build: Build the application (runs fmt, lint, unit tests before building)
build: fmt lint test-unit
	@echo "Building application..."
	@go build -o bin/server cmd/server/main.go
	@echo "✓ Build complete: bin/server"

## run: Run the application (loads .env.local if present)
run:
	@echo "Starting server..."
	@if [ -f .env.local ]; then \
		echo "Loading environment from .env.local..."; \
		set -a && . ./.env.local && set +a && go run cmd/server/main.go; \
	else \
		go run cmd/server/main.go; \
	fi

## clean: Remove build artifacts and temporary files
clean:
	@echo "Cleaning up..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "✓ Clean complete"

## check: Run all checks (fmt, lint, unit tests)
check: fmt lint test-unit
	@echo "✓ All checks passed"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✓ Dependencies downloaded"

## docker-up: Start Docker containers (TimescaleDB)
docker-up:
	@echo "Starting Docker containers..."
	@docker-compose up -d
	@echo "✓ Containers started"
	@echo "Waiting for database to be ready..."
	@for i in 1 2 3 4 5 6 7 8 9 10 11 12; do \
		if docker exec avt-timescaledb pg_isready -U telemetry_user -d telemetry_dev > /dev/null 2>&1; then \
			echo "✓ Database is ready"; \
			exit 0; \
		fi; \
		echo "  Waiting... ($$i/12)"; \
		sleep 2; \
	done; \
	echo "✗ Database failed to become ready in time"; \
	exit 1

## docker-down: Stop Docker containers
docker-down:
	@echo "Stopping Docker containers..."
	@docker-compose down
	@echo "✓ Containers stopped"

## migrate: Run database migrations
migrate:
	@echo "Running database migrations..."
	@chmod +x scripts/run-migrations.sh
	@DATABASE_URL=$(DATABASE_URL) MIGRATIONS_DIR=$(MIGRATIONS_DIR) ./scripts/run-migrations.sh
	@echo "✓ Migrations complete"

## migrate-down: Rollback last database migration
migrate-down:
	@echo "Rolling back last migration..."
	@$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1
	@echo "✓ Rollback complete"

## migrate-create: Create a new migration (usage: make migrate-create NAME=add_users_table)
migrate-create:
	@if [ -z "$(NAME)" ]; then echo "Error: NAME is required. Usage: make migrate-create NAME=add_users_table"; exit 1; fi
	@echo "Creating migration: $(NAME)"
	@$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)
	@echo "✓ Migration files created in $(MIGRATIONS_DIR)"

## db-shell: Open psql shell to database
db-shell:
	@docker-compose exec timescaledb psql -U telemetry_user -d telemetry_dev

## test-unit: Run unit tests only (fast)
test-unit:
	@echo "Running unit tests..."
	@go test -v -race -short ./...
	@echo "✓ Unit tests passed"

## test-integration: Run integration tests (requires Docker)
test-integration:
	@echo "Running integration tests..."
	@go test -v -race ./internal/repository
	@echo "✓ Integration tests passed"

## dev-setup: Set up local development environment
dev-setup: install-tools deps docker-up
	@$(MAKE) migrate
	@echo "✓ Development environment ready"
	@echo ""
	@echo "To start the server, run: make run"
	@echo "To run tests, run: make test"


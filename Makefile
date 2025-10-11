.PHONY: help build test lint fmt clean run install-linter

# Default target
.DEFAULT_GOAL := help

# Path to tools
GOLANGCI_LINT := $(shell which golangci-lint 2>/dev/null || echo "$(HOME)/go/bin/golangci-lint")
GOIMPORTS := $(shell which goimports 2>/dev/null || echo "$(HOME)/go/bin/goimports")

## help: Display this help message
help:
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

## install-linter: Install golangci-lint
install-linter:
	@echo "Installing golangci-lint..."
	@which golangci-lint > /dev/null || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin latest
	@echo "golangci-lint installed successfully"

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

## test: Run all tests
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "✓ Tests passed"

## test-coverage: Run tests with coverage report
test-coverage: test
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

## build: Build the application (runs fmt, lint, test before building)
build: fmt lint test
	@echo "Building application..."
	@go build -o bin/server cmd/server/main.go
	@echo "✓ Build complete: bin/server"

## run: Run the application
run:
	@echo "Starting server..."
	@go run cmd/server/main.go

## clean: Remove build artifacts and temporary files
clean:
	@echo "Cleaning up..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "✓ Clean complete"

## check: Run all checks (fmt, lint, test)
check: fmt lint test
	@echo "✓ All checks passed"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✓ Dependencies downloaded"


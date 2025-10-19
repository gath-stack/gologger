.PHONY: help test test-fast bench bench-long lint format doc clean check install-deps

help:
	@echo "Available targets:"
	@echo "  make test       - Run all tests"
	@echo "  make test-fast  - Run all tests (not verbose)"
	@echo "  make bench      - Run all benchmarks"
	@echo "  make lint       - Run linters"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make doc        - Start pkgsite"
	@echo "  make check      - Check linting and testing"

test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-fast:
	@echo "Running tests..."
	@go test ./...

bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

bench-long:
	@echo "Running long benchmarks (10s each)..."
	@go test -bench=. -benchtime=10s -benchmem ./...

doc:
	@echo "Starting Pkgsite..."
	@pkgsite -open .

lint:
	@echo "Running linters..."
	@golangci-lint run -v --timeout 5m ./...

format:
	@echo "Formatting code..."
	@golangci-lint run --fix -v --timeout 5m ./...
	@go fmt ./...

clean:
	@echo "Cleaning..."
	@rm -rf coverage.out coverage.html
	@echo "Clean complete"

install-deps:
	@echo "Installing development dependencies..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/pkgsite/cmd/pkgsite@latest
	@echo "Dependencies installed"

check: test lint
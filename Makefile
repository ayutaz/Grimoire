.PHONY: all build test clean run-example install deps lint fmt check-version web-build web-test

# Go version check
MIN_GO_VERSION = 1.21
GO_VERSION := $(shell go version | cut -d' ' -f3 | cut -d'.' -f2)
GO_VERSION_CHECK := $(shell echo "$(GO_VERSION) >= 21" | bc)

# Variables
BINARY_NAME=grimoire
GO_FILES=$(shell find . -name '*.go' -type f)
VERSION=$(shell git describe --tags --always --dirty)
COMMIT=$(shell git rev-parse --short HEAD)
DATE=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Default target
all: check-version build

# Check Go version
check-version:
	@if [ "$(GO_VERSION_CHECK)" != "1" ]; then \
		echo "Error: Go version $(MIN_GO_VERSION) or higher is required"; \
		echo "Current version: $(shell go version)"; \
		exit 1; \
	fi

# Build the binary
build: check-version
	go build $(LDFLAGS) -o $(BINARY_NAME) cmd/grimoire/main.go

# Build for all platforms
build-all: build-darwin build-linux build-windows

build-darwin:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 cmd/grimoire/main.go
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 cmd/grimoire/main.go

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 cmd/grimoire/main.go
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 cmd/grimoire/main.go

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe cmd/grimoire/main.go

# Build optimized binaries for release
build-release: clean-dist
	@echo "Building optimized binaries for all platforms..."
	@mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)" -o dist/$(BINARY_NAME)-darwin-amd64 cmd/grimoire/main.go
	GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)" -o dist/$(BINARY_NAME)-darwin-arm64 cmd/grimoire/main.go
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)" -o dist/$(BINARY_NAME)-linux-amd64 cmd/grimoire/main.go
	GOOS=linux GOARCH=arm64 go build -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)" -o dist/$(BINARY_NAME)-linux-arm64 cmd/grimoire/main.go
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)" -o dist/$(BINARY_NAME)-windows-amd64.exe cmd/grimoire/main.go
	@echo "Build complete. Binary sizes:"
	@ls -lh dist/

# Clean dist directory
clean-dist:
	rm -rf dist/

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...

# Run tests with coverage report
test-coverage: test
	go tool cover -html=coverage.out -o coverage.html

# Run benchmarks
bench:
	@if [ "$(OS)" = "Windows_NT" ]; then \
		echo "Running benchmarks on Windows..."; \
		cmd.exe /c "go test -bench=. -benchmem ./..."; \
	else \
		go test -bench=. -benchmem ./...; \
	fi

# Lint the code
lint:
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	golangci-lint run

# Format the code
fmt:
	go fmt ./...
	gofmt -s -w $(GO_FILES)

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/
	rm -f coverage.out coverage.html *_coverage.out
	rm -f *.test *.bench *.prof
	rm -f debug_*.png debug_*.txt
	find . -name "*.test" -o -name "*.tmp" | grep -v ".git" | xargs rm -f 2>/dev/null || true

# Install the binary
install: build
	go install $(LDFLAGS) ./cmd/grimoire

# Run example
run-example: build
	./$(BINARY_NAME) run examples/images/hello_world.png

# Debug example
debug-example: build
	./$(BINARY_NAME) debug examples/images/hello_world.png

# Quick test during development
dev: fmt lint test build

# Performance profiling
profile: build
	go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./...
	@echo "Run 'go tool pprof cpu.prof' to analyze CPU profile"
	@echo "Run 'go tool pprof mem.prof' to analyze memory profile"

# Docker build (if needed later)
docker-build:
	docker build -t grimoire:latest .

# Web WASM build
web-build:
	GOOS=js GOARCH=wasm go build -o web/static/wasm/grimoire.wasm cmd/grimoire-wasm/main.go

# Web E2E tests
web-test: web-build
	cd web/e2e && ./run-tests.sh

# Help target
help:
	@echo "Available targets:"
	@echo "  make build       - Build the binary"
	@echo "  make build-all   - Build for all platforms"
	@echo "  make test        - Run tests"
	@echo "  make lint        - Run linter"
	@echo "  make fmt         - Format code"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make install     - Install the binary"
	@echo "  make run-example - Run hello world example"
	@echo "  make dev         - Format, lint, test, and build"
	@echo "  make web-build   - Build WASM for web demo"
	@echo "  make web-test    - Run web E2E tests"
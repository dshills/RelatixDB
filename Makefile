# RelatixDB Makefile
# Provides common development tasks for the RelatixDB graph database

.PHONY: help build test test-unit test-integration test-mcp test-bench clean lint fmt vet install run dev deps check-deps profile

# Default target
.DEFAULT_GOAL := help

# Build configuration
BINARY_NAME := relatixdb
BUILD_DIR := ./build
CMD_DIR := ./cmd/relatixdb
GO_FILES := $(shell find . -name "*.go" -type f)

# Version information
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

## help: Show this help message
help:
	@echo "RelatixDB Development Makefile"
	@echo ""
	@echo "Available targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
	@echo ""

## build: Build the RelatixDB binary
build: $(BUILD_DIR)/$(BINARY_NAME)

$(BUILD_DIR)/$(BINARY_NAME): $(GO_FILES) go.mod go.sum
	@echo "Building RelatixDB..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

## install: Install RelatixDB to GOPATH/bin
install:
	@echo "Installing RelatixDB..."
	go install $(LDFLAGS) $(CMD_DIR)

## test: Run all tests
test: test-unit test-integration test-mcp

## test-unit: Run unit tests only
test-unit:
	@echo "Running unit tests..."
	go test -v ./internal/graph ./internal/storage

## test-integration: Run integration tests
test-integration:
	@echo "Running integration tests..."
	go test -v ./internal/mcp -run "TestHandler_ProcessSingleCommand|TestParseCommand|TestCommandArgsValidation"

## test-mcp: Run comprehensive MCP tool function tests
test-mcp:
	@echo "Running comprehensive MCP tool function tests..."
	go test -v ./internal/mcp -run "TestMCPToolFunctions|TestMCPCommandValidation"

## test-bench: Run benchmark tests
test-bench:
	@echo "Running benchmark tests..."
	go test -v -bench=. -benchmem ./internal/graph

## test-performance: Run performance target validation tests
test-performance:
	@echo "Running performance target validation..."
	go test -v ./internal/graph -run TestPerformanceTargets

## test-concurrent: Run concurrent operation tests
test-concurrent:
	@echo "Running concurrent operation tests..."
	go test -v ./internal/graph -run BenchmarkConcurrentOperations

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## test-race: Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	go test -race ./...

## lint: Run golangci-lint
lint:
	@echo "Running golangci-lint..."
	golangci-lint run

## fmt: Format Go code
fmt:
	@echo "Formatting Go code..."
	go fmt ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

## mod-tidy: Tidy go modules
mod-tidy:
	@echo "Tidying go modules..."
	go mod tidy

## mod-verify: Verify go modules
mod-verify:
	@echo "Verifying go modules..."
	go mod verify

## deps: Install development dependencies
deps:
	@echo "Installing development dependencies..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

## check-deps: Check if required dependencies are installed
check-deps:
	@echo "Checking dependencies..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Run 'make deps' to install." && exit 1)
	@echo "All dependencies are installed."

## clean: Clean build artifacts and test files
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	rm -f *.prof
	rm -f *.db
	rm -f test.db
	@echo "Clean complete."

## run: Build and run RelatixDB in memory mode
run: build
	@echo "Running RelatixDB in memory mode..."
	$(BUILD_DIR)/$(BINARY_NAME) -debug

## run-persistent: Build and run RelatixDB with persistent storage
run-persistent: build
	@echo "Running RelatixDB with persistent storage..."
	$(BUILD_DIR)/$(BINARY_NAME) -debug -db ./data/relatix.db

## dev: Run in development mode with debug output
dev:
	@echo "Running in development mode..."
	go run $(CMD_DIR) -debug

## profile-cpu: Run CPU profiling
profile-cpu:
	@echo "Running CPU profiling..."
	go test -cpuprofile=cpu.prof -bench=. ./internal/graph
	go tool pprof cpu.prof

## profile-mem: Run memory profiling
profile-mem:
	@echo "Running memory profiling..."
	go test -memprofile=mem.prof -bench=. ./internal/graph
	go tool pprof mem.prof

## validate: Run all validation checks (lint, vet, test)
validate: check-deps fmt vet lint test

## ci: Run continuous integration checks
ci: mod-verify validate test-race test-coverage

## release-build: Build release binaries for multiple platforms
release-build:
	@echo "Building release binaries..."
	@mkdir -p $(BUILD_DIR)/release
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)
	@echo "Release binaries built in $(BUILD_DIR)/release/"

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t relatixdb:$(VERSION) .

## benchmark-suite: Run comprehensive benchmark suite
benchmark-suite: test-bench test-performance test-concurrent
	@echo "Benchmark suite completed."

## mcp-demo: Run a demonstration of MCP functionality
mcp-demo: build
	@echo "Running MCP demonstration..."
	@echo '{"cmd": "add_node", "args": {"id": "demo:1", "type": "demo", "props": {"name": "Demo Node"}}}' | $(BUILD_DIR)/$(BINARY_NAME) -debug

## quick-test: Run essential tests quickly
quick-test:
	@echo "Running quick test suite..."
	go test ./internal/graph ./internal/mcp -run "TestMemoryGraph_AddNode|TestMCPToolFunctions"

## full-test: Run comprehensive test suite
full-test: test test-race test-coverage benchmark-suite
	@echo "Full test suite completed."

# Development workflow targets
## setup: Initial project setup for development
setup: deps mod-tidy validate
	@echo "Development environment setup complete."

## pre-commit: Run checks before committing
pre-commit: fmt vet lint test-mcp
	@echo "Pre-commit checks passed."
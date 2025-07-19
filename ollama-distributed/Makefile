# Ollama Distributed Test Suite Makefile

.PHONY: test test-unit test-integration test-e2e test-performance test-chaos test-all
.PHONY: test-ci test-coverage test-race test-bench test-security
.PHONY: setup-test-env clean-test-env
.PHONY: build build-docker run-docker
.PHONY: lint fmt vet check
.PHONY: deps deps-dev deps-test
.PHONY: help

# Variables
GO_VERSION ?= 1.21
BINARY_NAME ?= ollama-distributed
DOCKER_IMAGE ?= ollama-distributed
DOCKER_TAG ?= latest
TEST_TIMEOUT ?= 30m
TEST_NODE_COUNT ?= 3
ARTIFACTS_DIR ?= ./test-artifacts
COVERAGE_DIR ?= $(ARTIFACTS_DIR)/coverage
LOG_DIR ?= $(ARTIFACTS_DIR)/logs

# Test flags
TEST_FLAGS ?= -v
RACE_FLAGS ?= -race
BENCH_FLAGS ?= -bench=. -benchmem
COVERAGE_FLAGS ?= -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic

# Build flags
LDFLAGS=-ldflags "-X main.version=$(shell git describe --tags --always --dirty)"

# Directories
BUILD_DIR=bin
PKG_DIR=pkg
CMD_DIR=cmd
TEST_DIR=tests

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

# Default target
all: clean deps fmt vet test build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/node/main.go

# Build for Linux
build-linux:
	@echo "Building $(BINARY_NAME) for Linux..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_UNIX) $(CMD_DIR)/node/main.go

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	# Linux
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)/node/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_DIR)/node/main.go
	# macOS
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)/node/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)/node/main.go
	# Windows
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)/node/main.go

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out
	@rm -f *.prof

## Testing
test: test-unit ## Run default tests (unit tests)

test-unit: setup-test-env ## Run unit tests
	@echo "$(BLUE)Running unit tests...$(NC)"
	go test $(TEST_FLAGS) $(RACE_FLAGS) $(COVERAGE_FLAGS) ./tests/unit/...
	@echo "$(GREEN)Unit tests completed$(NC)"

test-integration: setup-test-env ## Run integration tests
	@echo "$(BLUE)Running integration tests...$(NC)"
	OLLAMA_TEST_ARTIFACTS_DIR=$(ARTIFACTS_DIR) \
	OLLAMA_TEST_NODE_COUNT=$(TEST_NODE_COUNT) \
	go test $(TEST_FLAGS) -timeout=$(TEST_TIMEOUT) -tags=integration ./tests/integration/...
	@echo "$(GREEN)Integration tests completed$(NC)"

test-e2e: setup-test-env ## Run end-to-end tests
	@echo "$(BLUE)Running E2E tests...$(NC)"
	OLLAMA_TEST_ARTIFACTS_DIR=$(ARTIFACTS_DIR) \
	OLLAMA_TEST_NODE_COUNT=$(TEST_NODE_COUNT) \
	go test $(TEST_FLAGS) -timeout=45m -tags=e2e ./tests/e2e/...
	@echo "$(GREEN)E2E tests completed$(NC)"

test-performance: setup-test-env ## Run performance tests
	@echo "$(BLUE)Running performance tests...$(NC)"
	OLLAMA_TEST_ARTIFACTS_DIR=$(ARTIFACTS_DIR) \
	OLLAMA_TEST_NODE_COUNT=5 \
	go test $(TEST_FLAGS) -timeout=60m $(BENCH_FLAGS) ./tests/performance/...
	@echo "$(GREEN)Performance tests completed$(NC)"

test-chaos: setup-test-env ## Run chaos tests
	@echo "$(BLUE)Running chaos tests...$(NC)"
	OLLAMA_TEST_ARTIFACTS_DIR=$(ARTIFACTS_DIR) \
	OLLAMA_TEST_NODE_COUNT=5 \
	go test $(TEST_FLAGS) -timeout=60m -tags=chaos ./tests/chaos/...
	@echo "$(GREEN)Chaos tests completed$(NC)"

test-all: test-unit test-integration test-e2e ## Run all tests (except performance and chaos)

test-full: test-unit test-integration test-e2e test-performance test-chaos ## Run all tests including performance and chaos

test-race: setup-test-env ## Run tests with race detection
	@echo "$(BLUE)Running tests with race detection...$(NC)"
	go test $(TEST_FLAGS) -race -short ./...

test-bench: setup-test-env ## Run benchmarks
	@echo "$(BLUE)Running benchmarks...$(NC)"
	go test $(BENCH_FLAGS) ./...

test-coverage: setup-test-env ## Generate coverage report
	@echo "$(BLUE)Generating coverage report...$(NC)"
	go test $(COVERAGE_FLAGS) ./...
	go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "$(GREEN)Coverage report generated: $(COVERAGE_DIR)/coverage.html$(NC)"

test-security: ## Run security tests
	@echo "$(BLUE)Running security tests...$(NC)"
	@which govulncheck > /dev/null || go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...
	@echo "$(GREEN)Security tests completed$(NC)"

test-ci: setup-test-env ## Run CI test suite
	@echo "$(BLUE)Running CI test suite...$(NC)"
	OLLAMA_TEST_CI=true \
	OLLAMA_TEST_ARTIFACTS_DIR=$(ARTIFACTS_DIR) \
	go test $(TEST_FLAGS) $(RACE_FLAGS) $(COVERAGE_FLAGS) -timeout=$(TEST_TIMEOUT) ./tests/unit/...
	OLLAMA_TEST_CI=true \
	OLLAMA_TEST_ARTIFACTS_DIR=$(ARTIFACTS_DIR) \
	OLLAMA_TEST_NODE_COUNT=$(TEST_NODE_COUNT) \
	go test $(TEST_FLAGS) -timeout=$(TEST_TIMEOUT) -tags=integration ./tests/integration/...
	@echo "$(GREEN)CI test suite completed$(NC)"

# Generate test coverage report
coverage:
	@echo "Generating coverage report..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

## Setup and Dependencies
deps: ## Install production dependencies
	@echo "$(BLUE)Installing production dependencies...$(NC)"
	go mod download
	go mod tidy

deps-dev: ## Install development dependencies
	@echo "$(BLUE)Installing development dependencies...$(NC)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/onsi/ginkgo/v2/ginkgo@latest
	go install github.com/axw/gocov/gocov@latest
	go install github.com/AlekSi/gocov-xml@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

deps-test: ## Install test dependencies
	@echo "$(BLUE)Installing test dependencies...$(NC)"
	go install github.com/stretchr/testify@latest

setup-test-env: ## Setup test environment
	@echo "$(BLUE)Setting up test environment...$(NC)"
	mkdir -p $(ARTIFACTS_DIR)
	mkdir -p $(COVERAGE_DIR)
	mkdir -p $(LOG_DIR)
	mkdir -p /tmp/ollama-test-data
	chmod 755 /tmp/ollama-test-data

clean-test-env: ## Clean test environment
	@echo "$(BLUE)Cleaning test environment...$(NC)"
	rm -rf $(ARTIFACTS_DIR)
	rm -rf /tmp/ollama-test-*
	docker system prune -f 2>/dev/null || true

## Code Quality
fmt: ## Format code
	@echo "$(BLUE)Formatting code...$(NC)"
	gofmt -s -w .

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	go vet ./...

lint: ## Run golangci-lint
	@echo "$(BLUE)Running golangci-lint...$(NC)"
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run --timeout=10m

staticcheck: ## Run staticcheck
	@echo "$(BLUE)Running staticcheck...$(NC)"
	@which staticcheck > /dev/null || go install honnef.co/go/tools/cmd/staticcheck@latest
	staticcheck ./...

check: fmt vet lint staticcheck ## Run all code quality checks

# Run security scanner
security:
	@echo "Running security scanner..."
	@which gosec > /dev/null || (echo "gosec not found. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest" && exit 1)
	gosec ./...

# Generate mocks
mocks:
	@echo "Generating mocks..."
	@which mockgen > /dev/null || (echo "mockgen not found. Install with: go install go.uber.org/mock/mockgen@latest" && exit 1)
	mockgen -source=pkg/p2p/node.go -destination=tests/mocks/p2p_mock.go
	mockgen -source=pkg/consensus/engine.go -destination=tests/mocks/consensus_mock.go
	mockgen -source=pkg/scheduler/engine.go -destination=tests/mocks/scheduler_mock.go

# Generate protobuf files
protobuf:
	@echo "Generating protobuf files..."
	@which protoc > /dev/null || (echo "protoc not found. Install Protocol Buffers compiler" && exit 1)
	@find . -name "*.proto" -exec protoc --go_out=. --go-grpc_out=. {} \;

# Install development tools
install-tools:
	@echo "Installing development tools..."
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOGET) -u github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	$(GOGET) -u go.uber.org/mock/mockgen@latest

# Build Docker image
docker:
	@echo "Building Docker image..."
	docker build -t ollama-distributed:latest .

# Build and tag Docker image
docker-tag:
	@echo "Building and tagging Docker image..."
	docker build -t ollama-distributed:latest -t ollama-distributed:$(shell git describe --tags --always --dirty) .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -d --name ollama-distributed -p 11434:11434 -p 8080:8080 ollama-distributed:latest

# Stop Docker container
docker-stop:
	@echo "Stopping Docker container..."
	docker stop ollama-distributed || true
	docker rm ollama-distributed || true

# Development server
dev:
	@echo "Starting development server..."
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-dev $(CMD_DIR)/node/main.go
	./$(BUILD_DIR)/$(BINARY_NAME)-dev start --config config/node.yaml

# Install binary
install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

# Uninstall binary
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@rm -f $(GOPATH)/bin/$(BINARY_NAME)

# Create release
release: clean fmt vet test build-all
	@echo "Creating release..."
	@mkdir -p release
	@cp $(BUILD_DIR)/* release/
	@cp README.md release/
	@cp LICENSE release/
	@tar -czf release/ollama-distributed-$(shell git describe --tags --always --dirty).tar.gz -C release .

# Database migration
migrate:
	@echo "Running database migrations..."
	# Add migration commands here if needed

# Start development environment
dev-env:
	@echo "Starting development environment..."
	docker-compose -f docker-compose.dev.yml up -d

# Stop development environment
dev-env-stop:
	@echo "Stopping development environment..."
	docker-compose -f docker-compose.dev.yml down

# Load test data
load-test-data:
	@echo "Loading test data..."
	# Add test data loading commands here

# Performance profiling
profile-cpu:
	@echo "Running CPU profiling..."
	$(GOTEST) -cpuprofile=cpu.prof -bench=. ./...

profile-memory:
	@echo "Running memory profiling..."
	$(GOTEST) -memprofile=mem.prof -bench=. ./...

profile-web:
	@echo "Starting web profiling server..."
	$(GOCMD) tool pprof -http=:8081 cpu.prof

## Help
help: ## Show this help message
	@echo "$(BLUE)Ollama Distributed Test Suite$(NC)"
	@echo ""
	@echo "$(YELLOW)Usage:$(NC)"
	@echo "  make <target>"
	@echo ""
	@echo "$(YELLOW)Targets:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## Test Categories
test-p2p: setup-test-env ## Run P2P related tests
	@echo "$(BLUE)Running P2P tests...$(NC)"
	go test $(TEST_FLAGS) -run=".*P2P.*|.*Discovery.*|.*Network.*" ./tests/unit/...
	go test $(TEST_FLAGS) -run=".*P2P.*|.*Discovery.*|.*Network.*" ./tests/integration/...

test-consensus: setup-test-env ## Run consensus related tests
	@echo "$(BLUE)Running consensus tests...$(NC)"
	go test $(TEST_FLAGS) -run=".*Consensus.*|.*Raft.*|.*Byzantine.*" ./tests/unit/...
	go test $(TEST_FLAGS) -run=".*Consensus.*|.*Raft.*|.*Byzantine.*" ./tests/integration/...

test-scheduler: setup-test-env ## Run scheduler related tests
	@echo "$(BLUE)Running scheduler tests...$(NC)"
	go test $(TEST_FLAGS) -run=".*Scheduler.*|.*LoadBalanc.*|.*Distribute.*" ./tests/unit/...
	go test $(TEST_FLAGS) -run=".*Scheduler.*|.*LoadBalanc.*|.*Distribute.*" ./tests/integration/...

test-models: setup-test-env ## Run model management tests
	@echo "$(BLUE)Running model management tests...$(NC)"
	go test $(TEST_FLAGS) -run=".*Model.*|.*Sync.*|.*Replication.*" ./tests/unit/...
	go test $(TEST_FLAGS) -run=".*Model.*|.*Sync.*|.*Replication.*" ./tests/integration/...

test-api: setup-test-env ## Run API related tests
	@echo "$(BLUE)Running API tests...$(NC)"
	go test $(TEST_FLAGS) -run=".*API.*|.*HTTP.*|.*REST.*" ./tests/unit/...
	go test $(TEST_FLAGS) -run=".*API.*|.*HTTP.*|.*REST.*" ./tests/integration/...

## Test Runners
test-runner: ## Run the test runner
	@echo "$(BLUE)Running test runner...$(NC)"
	go run ./tests/test_runner.go --verbose --parallel --types=unit,integration

test-runner-all: ## Run test runner with all test types
	@echo "$(BLUE)Running test runner with all test types...$(NC)"
	go run ./tests/test_runner.go --verbose --parallel --types=unit,integration,e2e,performance

test-runner-ci: ## Run test runner in CI mode
	@echo "$(BLUE)Running test runner in CI mode...$(NC)"
	go run ./tests/test_runner.go --ci --fail-fast --types=unit,integration

## Utilities
test-clean: clean-test-env ## Clean test artifacts and environment

test-logs: ## Show test logs
	@echo "$(BLUE)Recent test logs:$(NC)"
	@find $(LOG_DIR) -name "*.log" -type f -exec tail -n 50 {} \; 2>/dev/null || echo "No logs found"

test-artifacts: ## Show test artifacts
	@echo "$(BLUE)Test artifacts:$(NC)"
	@ls -la $(ARTIFACTS_DIR) 2>/dev/null || echo "No artifacts found"

test-status: ## Show test status
	@echo "$(BLUE)Test Status:$(NC)"
	@echo "Last test run: $(shell find $(ARTIFACTS_DIR) -name "*.log" -type f -exec stat -c %y {} \; 2>/dev/null | sort -r | head -1 || echo 'Never')"
	@echo "Artifacts size: $(shell du -sh $(ARTIFACTS_DIR) 2>/dev/null | cut -f1 || echo '0')"
	@echo "Coverage: $(shell grep -o '[0-9]*\.[0-9]*%' $(COVERAGE_DIR)/coverage.out 2>/dev/null | tail -1 || echo 'N/A')"

## Quick Commands
quick-test: ## Quick test run (unit tests only)
	@echo "$(BLUE)Running quick tests...$(NC)"
	go test -short ./tests/unit/...

smoke-test: ## Smoke test (basic functionality)
	@echo "$(BLUE)Running smoke tests...$(NC)"
	go test -run=".*Smoke.*|.*Basic.*" ./tests/unit/...
	go test -run=".*Smoke.*|.*Basic.*" ./tests/integration/...

.DEFAULT_GOAL := help
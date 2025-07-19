# Ollamacron Makefile
# Comprehensive build system for the Ollamacron distributed platform

# Build information
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION ?= $(shell go version | awk '{print $$3}')

# Build configuration
GO_LDFLAGS := -s -w
GO_LDFLAGS += -X main.version=$(VERSION)
GO_LDFLAGS += -X main.commit=$(COMMIT)
GO_LDFLAGS += -X main.date=$(DATE)
GO_LDFLAGS += -X main.goVersion=$(GO_VERSION)

# Directories
BUILD_DIR := build
DIST_DIR := dist
DOCKER_DIR := ../../deploy/docker
SYSTEMD_DIR := ../../deploy/systemd

# Binary name
BINARY := ollamacron

# Platform detection
OS := $(shell go env GOOS)
ARCH := $(shell go env GOARCH)

# Cross-compilation targets
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

.PHONY: all build clean test lint fmt vet deps check install uninstall docker \
        docker-build docker-run docker-push release cross-compile help

# Default target
all: clean deps check build

# Help target
help:
	@echo "$(GREEN)Ollamacron Build System$(NC)"
	@echo ""
	@echo "$(YELLOW)Available targets:$(NC)"
	@echo "  $(BLUE)build$(NC)           - Build the binary for current platform"
	@echo "  $(BLUE)clean$(NC)           - Clean build artifacts"
	@echo "  $(BLUE)test$(NC)            - Run tests"
	@echo "  $(BLUE)lint$(NC)            - Run linter"
	@echo "  $(BLUE)fmt$(NC)             - Format code"
	@echo "  $(BLUE)vet$(NC)             - Run go vet"
	@echo "  $(BLUE)deps$(NC)            - Download dependencies"
	@echo "  $(BLUE)check$(NC)           - Run all checks (fmt, vet, lint, test)"
	@echo "  $(BLUE)install$(NC)         - Install binary and systemd service"
	@echo "  $(BLUE)uninstall$(NC)       - Remove installed binary and service"
	@echo "  $(BLUE)docker$(NC)          - Build and run Docker container"
	@echo "  $(BLUE)docker-build$(NC)    - Build Docker image"
	@echo "  $(BLUE)docker-run$(NC)      - Run Docker container"
	@echo "  $(BLUE)docker-push$(NC)     - Push Docker image"
	@echo "  $(BLUE)release$(NC)         - Create release packages"
	@echo "  $(BLUE)cross-compile$(NC)   - Build for all platforms"
	@echo "  $(BLUE)help$(NC)            - Show this help message"
	@echo ""
	@echo "$(YELLOW)Environment variables:$(NC)"
	@echo "  $(BLUE)VERSION$(NC)         - Build version (default: dev)"
	@echo "  $(BLUE)COMMIT$(NC)          - Git commit hash"
	@echo "  $(BLUE)DATE$(NC)            - Build date"
	@echo ""
	@echo "$(YELLOW)Examples:$(NC)"
	@echo "  make build VERSION=1.0.0"
	@echo "  make docker VERSION=1.0.0"
	@echo "  make release VERSION=1.0.0"

# Build the binary
build:
	@echo "$(GREEN)Building $(BINARY)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build \
		-ldflags "$(GO_LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY) \
		.
	@echo "$(GREEN)Build completed: $(BUILD_DIR)/$(BINARY)$(NC)"
	@echo ""
	@echo "$(YELLOW)Build information:$(NC)"
	@$(BUILD_DIR)/$(BINARY) version

# Clean build artifacts
clean:
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DIST_DIR)
	@echo "$(GREEN)Clean completed$(NC)"

# Download dependencies
deps:
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)Dependencies updated$(NC)"

# Format code
fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)Code formatted$(NC)"

# Run go vet
vet:
	@echo "$(GREEN)Running go vet...$(NC)"
	@go vet ./...
	@echo "$(GREEN)Vet completed$(NC)"

# Run linter
lint:
	@echo "$(GREEN)Running linter...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)golangci-lint not found, skipping...$(NC)"; \
	fi
	@echo "$(GREEN)Lint completed$(NC)"

# Run tests
test:
	@echo "$(GREEN)Running tests...$(NC)"
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)Tests completed$(NC)"

# Run all checks
check: fmt vet lint test

# Cross-compile for multiple platforms
cross-compile:
	@echo "$(GREEN)Cross-compiling for multiple platforms...$(NC)"
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		binary_name=$(BINARY); \
		if [ "$$os" = "windows" ]; then \
			binary_name=$(BINARY).exe; \
		fi; \
		echo "Building for $$os/$$arch..."; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build \
			-ldflags "$(GO_LDFLAGS)" \
			-o $(DIST_DIR)/$(BINARY)-$$os-$$arch/$$binary_name \
			.; \
		tar -czf $(DIST_DIR)/$(BINARY)-$$os-$$arch.tar.gz \
			-C $(DIST_DIR)/$(BINARY)-$$os-$$arch \
			$$binary_name; \
		rm -rf $(DIST_DIR)/$(BINARY)-$$os-$$arch; \
	done
	@echo "$(GREEN)Cross-compilation completed$(NC)"

# Create release packages
release: cross-compile
	@echo "$(GREEN)Creating release packages...$(NC)"
	@cd $(DIST_DIR) && sha256sum *.tar.gz > SHA256SUMS
	@echo "$(GREEN)Release packages created in $(DIST_DIR)$(NC)"

# Install binary and systemd service
install: build
	@echo "$(GREEN)Installing $(BINARY)...$(NC)"
	@sudo cp $(BUILD_DIR)/$(BINARY) /usr/local/bin/
	@sudo chmod +x /usr/local/bin/$(BINARY)
	@if [ -f "$(SYSTEMD_DIR)/ollamacron.service" ]; then \
		sudo cp $(SYSTEMD_DIR)/ollamacron.service /etc/systemd/system/; \
		sudo systemctl daemon-reload; \
		echo "$(GREEN)Systemd service installed$(NC)"; \
	fi
	@echo "$(GREEN)Installation completed$(NC)"

# Uninstall binary and systemd service
uninstall:
	@echo "$(GREEN)Uninstalling $(BINARY)...$(NC)"
	@sudo systemctl stop ollamacron 2>/dev/null || true
	@sudo systemctl disable ollamacron 2>/dev/null || true
	@sudo rm -f /etc/systemd/system/ollamacron.service
	@sudo rm -f /usr/local/bin/$(BINARY)
	@sudo systemctl daemon-reload
	@echo "$(GREEN)Uninstallation completed$(NC)"

# Docker targets
docker: docker-build docker-run

docker-build:
	@echo "$(GREEN)Building Docker image...$(NC)"
	@docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg DATE=$(DATE) \
		-t ollamacron:$(VERSION) \
		-t ollamacron:latest \
		-f $(DOCKER_DIR)/Dockerfile \
		../..
	@echo "$(GREEN)Docker image built: ollamacron:$(VERSION)$(NC)"

docker-run:
	@echo "$(GREEN)Running Docker container...$(NC)"
	@docker run --rm -it \
		-p 11434:11434 \
		-p 4001:4001 \
		-p 8080:8080 \
		-p 9090:9090 \
		--name ollamacron \
		ollamacron:$(VERSION)

docker-push:
	@echo "$(GREEN)Pushing Docker image...$(NC)"
	@docker push ollamacron:$(VERSION)
	@docker push ollamacron:latest

# Development targets
dev-build:
	@echo "$(GREEN)Building development version...$(NC)"
	@go build -race -o $(BUILD_DIR)/$(BINARY)-dev .

dev-run: dev-build
	@echo "$(GREEN)Running development version...$(NC)"
	@$(BUILD_DIR)/$(BINARY)-dev --config ../../config/config.yaml --log-level debug

# Generate configuration
config:
	@echo "$(GREEN)Generating configuration...$(NC)"
	@mkdir -p ../../config
	@go run . config generate ../../config/config.yaml

# Show build information
info:
	@echo "$(GREEN)Build Information:$(NC)"
	@echo "  Version: $(VERSION)"
	@echo "  Commit: $(COMMIT)"
	@echo "  Date: $(DATE)"
	@echo "  Go Version: $(GO_VERSION)"
	@echo "  OS/Arch: $(OS)/$(ARCH)"
	@echo "  LDFLAGS: $(GO_LDFLAGS)"

# Benchmark tests
bench:
	@echo "$(GREEN)Running benchmarks...$(NC)"
	@go test -bench=. -benchmem ./...

# Profile build
profile:
	@echo "$(GREEN)Building with profiling...$(NC)"
	@go build -buildmode=exe -o $(BUILD_DIR)/$(BINARY)-profile .

# Security scan
security:
	@echo "$(GREEN)Running security scan...$(NC)"
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "$(YELLOW)gosec not found, skipping...$(NC)"; \
	fi

# Vulnerability check
vuln:
	@echo "$(GREEN)Checking for vulnerabilities...$(NC)"
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "$(YELLOW)govulncheck not found, skipping...$(NC)"; \
	fi
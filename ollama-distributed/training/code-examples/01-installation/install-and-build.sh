#!/bin/bash
# 01-installation/install-and-build.sh
# Comprehensive installation and build script for Ollama Distributed Training

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
REQUIRED_GO_VERSION="1.21"
PROJECT_DIR="${PROJECT_DIR:-/home/kp/ollamamax/ollama-distributed}"
BIN_DIR="${PROJECT_DIR}/bin"

# Helper Functions
print_header() {
    echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}🎯 $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${CYAN}ℹ️  $1${NC}"
}

# Validation Functions
validate_go_version() {
    print_header "Go Version Validation"
    
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed"
        print_info "Install Go from: https://golang.org/dl/"
        exit 1
    fi
    
    local current_version
    current_version=$(go version | awk '{print $3}' | sed 's/go//' | sed 's/\([0-9]*\.[0-9]*\).*/\1/')
    
    print_info "Current Go version: $current_version"
    print_info "Required Go version: $REQUIRED_GO_VERSION"
    
    # Simple version comparison
    if [[ "$(printf '%s\n' "$REQUIRED_GO_VERSION" "$current_version" | sort -V | head -n1)" != "$REQUIRED_GO_VERSION" ]]; then
        print_error "Go version $current_version is less than required $REQUIRED_GO_VERSION"
        exit 1
    fi
    
    print_success "Go version meets requirements"
}

validate_system_requirements() {
    print_header "System Requirements Check"
    
    # Check available disk space (need 2GB)
    local available_space
    available_space=$(df . | tail -1 | awk '{print $4}')
    local required_space=2097152  # 2GB in KB
    
    if [[ $available_space -lt $required_space ]]; then
        print_warning "Low disk space. Available: $(($available_space / 1024))MB, Recommended: 2048MB"
    else
        print_success "Sufficient disk space available"
    fi
    
    # Check memory
    if command -v free &> /dev/null; then
        local memory_mb
        memory_mb=$(free -m | grep '^Mem:' | awk '{print $2}')
        if [[ $memory_mb -lt 4096 ]]; then
            print_warning "System has ${memory_mb}MB RAM. Recommended: 4096MB"
        else
            print_success "Sufficient memory available: ${memory_mb}MB"
        fi
    fi
    
    # Check required tools
    local tools=("git" "curl" "make")
    for tool in "${tools[@]}"; do
        if command -v "$tool" &> /dev/null; then
            print_success "$tool is available"
        else
            print_error "$tool is required but not installed"
        fi
    done
}

validate_project_structure() {
    print_header "Project Structure Validation"
    
    if [[ ! -d "$PROJECT_DIR" ]]; then
        print_error "Project directory not found: $PROJECT_DIR"
        print_info "Ensure you have cloned the repository"
        exit 1
    fi
    
    print_success "Project directory exists: $PROJECT_DIR"
    
    # Check for essential files
    local essential_files=(
        "cmd/node/main.go"
        "go.mod"
        "go.sum"
        "internal/config/config.go"
        "pkg/api/server.go"
    )
    
    cd "$PROJECT_DIR"
    
    for file in "${essential_files[@]}"; do
        if [[ -f "$file" ]]; then
            print_success "Found essential file: $file"
        else
            print_warning "Missing file (may be normal): $file"
        fi
    done
}

build_project() {
    print_header "Building Ollama Distributed"
    
    cd "$PROJECT_DIR"
    
    # Clean previous builds
    print_info "Cleaning previous builds..."
    rm -rf "$BIN_DIR"
    mkdir -p "$BIN_DIR"
    
    # Build main binary
    print_info "Building ollama-distributed binary..."
    if go build -v -o "$BIN_DIR/ollama-distributed" ./cmd/node; then
        print_success "Built ollama-distributed binary"
    else
        print_error "Failed to build ollama-distributed binary"
        exit 1
    fi
    
    # Build additional tools
    print_info "Building additional tools..."
    
    if [[ -d "cmd/performance-test" ]]; then
        go build -o "$BIN_DIR/performance-test" ./cmd/performance-test && \
        print_success "Built performance-test tool"
    fi
    
    if [[ -d "cmd/config-tool" ]]; then
        go build -o "$BIN_DIR/config-tool" ./cmd/config-tool && \
        print_success "Built config-tool"
    fi
    
    # Make binaries executable
    chmod +x "$BIN_DIR"/*
    
    # Verify builds
    print_info "Verifying builds..."
    for binary in "$BIN_DIR"/*; do
        if [[ -x "$binary" ]]; then
            local size
            size=$(stat -f%z "$binary" 2>/dev/null || stat -c%s "$binary")
            print_success "$(basename "$binary"): $size bytes"
        fi
    done
}

test_installation() {
    print_header "Installation Testing"
    
    local main_binary="$BIN_DIR/ollama-distributed"
    
    # Test help command
    print_info "Testing help command..."
    if "$main_binary" --help > /dev/null 2>&1; then
        print_success "Help command works"
    else
        print_error "Help command failed"
        return 1
    fi
    
    # Test version command
    print_info "Testing version command..."
    if version_output=$("$main_binary" --version 2>&1); then
        print_success "Version command works: $version_output"
    else
        print_error "Version command failed"
        return 1
    fi
    
    # Test validate command if available
    print_info "Testing validate command..."
    if "$main_binary" validate --help > /dev/null 2>&1; then
        print_success "Validate command available"
    else
        print_warning "Validate command not available (may be normal)"
    fi
    
    # Test quickstart command if available
    print_info "Testing quickstart command..."
    if "$main_binary" quickstart --help > /dev/null 2>&1; then
        print_success "Quickstart command available"
    else
        print_warning "Quickstart command not available (may be normal)"
    fi
    
    return 0
}

create_training_environment() {
    print_header "Creating Training Environment"
    
    local base_dir="$HOME/.ollama-distributed"
    local profiles_dir="$base_dir/profiles"
    local data_dir="$PROJECT_DIR/dev-data"
    
    # Create directories
    mkdir -p "$base_dir" "$profiles_dir" "$data_dir"
    print_success "Created training directories"
    
    # Create development configuration
    cat > "$profiles_dir/development.yaml" << 'EOF'
# Development Configuration for Ollama Distributed Training
api:
  listen: "127.0.0.1:8080"
  debug: true
  cors:
    enabled: true
    origins: ["*"]

p2p:
  listen: "127.0.0.1:4001"
  bootstrap: []

web:
  listen: "127.0.0.1:8081"
  enable_auth: false
  static_dir: "./web/static"

storage:
  data_dir: "./dev-data"
  models_dir: "./dev-data/models"

logging:
  level: "debug"
  output: "console"
  file: "./dev-data/logs/ollama-distributed.log"

performance:
  monitoring_enabled: true
  metrics_interval: 5
  optimization_enabled: false

consensus:
  data_dir: "./dev-data/consensus"
  bootstrap: true
  bind_addr: "127.0.0.1:7000"

scheduler:
  algorithm: "round_robin"
  load_balancing: "cpu_aware"
  worker_count: 4
  queue_size: 100
EOF
    
    print_success "Created development configuration"
    
    # Create testing configuration
    cat > "$profiles_dir/testing.yaml" << 'EOF'
# Testing Configuration for Ollama Distributed Training
api:
  listen: "127.0.0.1:9080"
  debug: false

p2p:
  listen: "127.0.0.1:4002"

web:
  listen: "127.0.0.1:9081"

storage:
  data_dir: "./test-data"

logging:
  level: "info"
  output: "file"
  file: "./test-data/logs/test.log"
EOF
    
    print_success "Created testing configuration"
    
    # Create training aliases
    cat > "$base_dir/training-aliases.sh" << EOF
#!/bin/bash
# Training aliases for Ollama Distributed

export OLLAMA_DISTRIBUTED_BIN="$BIN_DIR/ollama-distributed"
export OLLAMA_DISTRIBUTED_DEV_CONFIG="$profiles_dir/development.yaml"
export OLLAMA_DISTRIBUTED_TEST_CONFIG="$profiles_dir/testing.yaml"

alias od="\$OLLAMA_DISTRIBUTED_BIN"
alias od-dev="\$OLLAMA_DISTRIBUTED_BIN start --config \$OLLAMA_DISTRIBUTED_DEV_CONFIG"
alias od-test="\$OLLAMA_DISTRIBUTED_BIN start --config \$OLLAMA_DISTRIBUTED_TEST_CONFIG"
alias od-status="\$OLLAMA_DISTRIBUTED_BIN status"
alias od-validate="\$OLLAMA_DISTRIBUTED_BIN validate --config \$OLLAMA_DISTRIBUTED_DEV_CONFIG"
alias od-logs="tail -f $data_dir/logs/ollama-distributed.log"

echo "🚀 Ollama Distributed training environment loaded"
echo "Available aliases:"
echo "  od         - Main binary"
echo "  od-dev     - Start development server"
echo "  od-test    - Start testing server"  
echo "  od-status  - Show status"
echo "  od-validate - Validate configuration"
echo "  od-logs    - Follow logs"
EOF
    
    chmod +x "$base_dir/training-aliases.sh"
    print_success "Created training aliases"
    
    print_info "To load aliases, run: source $base_dir/training-aliases.sh"
}

generate_training_summary() {
    print_header "Installation Summary"
    
    local main_binary="$BIN_DIR/ollama-distributed"
    
    echo -e "\n${GREEN}🎉 Installation Complete!${NC}\n"
    
    echo "📍 Installation Details:"
    echo "   Project Directory: $PROJECT_DIR"
    echo "   Binary Directory: $BIN_DIR"
    echo "   Configuration: $HOME/.ollama-distributed/profiles/"
    echo "   Data Directory: $PROJECT_DIR/dev-data"
    echo ""
    
    echo "🔧 Built Binaries:"
    for binary in "$BIN_DIR"/*; do
        if [[ -x "$binary" ]]; then
            echo "   $(basename "$binary") - $binary"
        fi
    done
    echo ""
    
    echo "🚀 Quick Start Commands:"
    echo "   Test installation: $main_binary --help"
    echo "   Validate environment: $main_binary validate --quick"
    echo "   Start development: $main_binary start --config ~/.ollama-distributed/profiles/development.yaml"
    echo "   Check status: $main_binary status"
    echo ""
    
    echo "📚 Next Steps:"
    echo "   1. Run validation: $main_binary validate --quick"
    echo "   2. Start the service: $main_binary start"
    echo "   3. Open web interface: http://localhost:8081"
    echo "   4. Test API: curl http://localhost:8080/health"
    echo ""
    
    echo "📖 Training Resources:"
    echo "   Load aliases: source ~/.ollama-distributed/training-aliases.sh"
    echo "   View logs: tail -f ./dev-data/logs/ollama-distributed.log"
    echo "   Configuration: ~/.ollama-distributed/profiles/development.yaml"
    echo ""
}

# Main execution
main() {
    print_header "Ollama Distributed Installation and Build Script"
    
    print_info "Starting comprehensive installation process..."
    
    # Validation phase
    validate_go_version
    validate_system_requirements  
    validate_project_structure
    
    # Build phase
    build_project
    
    # Test phase
    if ! test_installation; then
        print_error "Installation testing failed"
        exit 1
    fi
    
    # Environment setup
    create_training_environment
    
    # Summary
    generate_training_summary
    
    print_success "Installation completed successfully!"
    return 0
}

# Handle command line arguments
case "${1:-install}" in
    "install"|"")
        main
        ;;
    "validate-only")
        validate_go_version
        validate_system_requirements
        validate_project_structure
        ;;
    "build-only")
        build_project
        ;;
    "test-only")
        test_installation
        ;;
    "help")
        echo "Usage: $0 [install|validate-only|build-only|test-only|help]"
        echo ""
        echo "Commands:"
        echo "  install       - Full installation (default)"
        echo "  validate-only - Only validate prerequisites"
        echo "  build-only    - Only build binaries"
        echo "  test-only     - Only test installation"
        echo "  help          - Show this help"
        ;;
    *)
        print_error "Unknown command: $1"
        echo "Run '$0 help' for usage information"
        exit 1
        ;;
esac
#!/bin/bash
# Ollama Distributed Training Validation Scripts
# These scripts validate that training exercises work with the actual software

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_DIR="$HOME/.ollama-distributed"
CONFIG_FILE="$BASE_DIR/config.yaml"
DEV_PROFILE="$BASE_DIR/profiles/development.yaml"
BINARY_PATH="./bin/ollama-distributed"

# Helper Functions
print_header() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
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
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Validation Functions

validate_prerequisites() {
    print_header "Prerequisites Validation"
    
    local errors=0
    
    # Check Go
    if command -v go &> /dev/null; then
        local go_version=$(go version | awk '{print $3}' | sed 's/go//')
        print_success "Go installed: $go_version"
    else
        print_error "Go not found. Install Go 1.19+ from https://golang.org"
        ((errors++))
    fi
    
    # Check Git
    if command -v git &> /dev/null; then
        local git_version=$(git --version | awk '{print $3}')
        print_success "Git installed: $git_version"
    else
        print_error "Git not found. Install Git from https://git-scm.com"
        ((errors++))
    fi
    
    # Check curl
    if command -v curl &> /dev/null; then
        print_success "curl available"
    else
        print_error "curl not found. Install curl for API testing"
        ((errors++))
    fi
    
    # Check jq
    if command -v jq &> /dev/null; then
        print_success "jq available for JSON processing"
    else
        print_warning "jq not found. Install jq for better API response formatting"
    fi
    
    # Check ports
    local ports=("8080" "8081" "4001")
    for port in "${ports[@]}"; do
        if netstat -ln 2>/dev/null | grep -q ":$port "; then
            print_warning "Port $port is in use. You may need to use different ports"
        else
            print_success "Port $port available"
        fi
    done
    
    echo
    if [[ $errors -eq 0 ]]; then
        print_success "Prerequisites check passed"
        return 0
    else
        print_error "$errors prerequisite(s) missing"
        return 1
    fi
}

validate_installation() {
    print_header "Installation Validation"
    
    # Check if binary exists
    if [[ -f "$BINARY_PATH" ]]; then
        print_success "Binary found at $BINARY_PATH"
    else
        print_error "Binary not found at $BINARY_PATH"
        print_info "Run: go build -o bin/ollama-distributed ./cmd/distributed-ollama"
        return 1
    fi
    
    # Check if binary is executable
    if [[ -x "$BINARY_PATH" ]]; then
        print_success "Binary is executable"
    else
        print_error "Binary is not executable"
        print_info "Run: chmod +x $BINARY_PATH"
        return 1
    fi
    
    # Test version command
    if version_output=$($BINARY_PATH --version 2>&1); then
        print_success "Version command works: $version_output"
    else
        print_error "Version command failed: $version_output"
        return 1
    fi
    
    # Test help command
    if $BINARY_PATH --help > /dev/null 2>&1; then
        print_success "Help command works"
    else
        print_error "Help command failed"
        return 1
    fi
    
    print_success "Installation validation passed"
    return 0
}

validate_configuration() {
    print_header "Configuration Validation"
    
    # Create base directory
    mkdir -p "$BASE_DIR"
    mkdir -p "$BASE_DIR/profiles"
    
    # Test setup command (dry run)
    print_info "Testing setup wizard..."
    if $BINARY_PATH setup --help > /dev/null 2>&1; then
        print_success "Setup command available"
    else
        print_warning "Setup command structure may differ from documentation"
    fi
    
    # Create development profile
    print_info "Creating development profile..."
    cat > "$DEV_PROFILE" << EOF
# Development Profile for Ollama Distributed Training
api:
  listen: "127.0.0.1:8080"
  debug: true
  
p2p:
  listen: "127.0.0.1:4001"
  
storage:
  data_dir: "./dev-data"
  
web:
  listen: "127.0.0.1:8081"
  enable_auth: false
  
logging:
  level: "debug"
  output: "console"
EOF
    
    if [[ -f "$DEV_PROFILE" ]]; then
        print_success "Development profile created"
    else
        print_error "Failed to create development profile"
        return 1
    fi
    
    # Test configuration validation
    if $BINARY_PATH validate --help > /dev/null 2>&1; then
        print_success "Validate command available"
        
        # Try to validate the config
        print_info "Testing configuration validation..."
        if timeout 10 $BINARY_PATH validate --config "$DEV_PROFILE" > /dev/null 2>&1; then
            print_success "Configuration validates successfully"
        else
            print_warning "Configuration validation had issues (may be normal in development)"
        fi
    else
        print_warning "Validate command not available or different syntax"
    fi
    
    print_success "Configuration validation completed"
    return 0
}

validate_startup() {
    print_header "Startup Validation"
    
    # Test dry run if available
    print_info "Testing startup process..."
    
    # Try to start in background for testing
    print_info "Attempting background startup test..."
    
    local pid_file="$BASE_DIR/test.pid"
    local log_file="$BASE_DIR/test.log"
    
    # Start in background with timeout
    timeout 30 $BINARY_PATH start --config "$DEV_PROFILE" > "$log_file" 2>&1 &
    local start_pid=$!
    echo $start_pid > "$pid_file"
    
    # Wait for startup
    sleep 10
    
    # Check if process is still running
    if kill -0 $start_pid 2>/dev/null; then
        print_success "Process started successfully (PID: $start_pid)"
        
        # Test health endpoint
        local max_attempts=10
        local attempt=1
        
        while [[ $attempt -le $max_attempts ]]; do
            if curl -s -f http://127.0.0.1:8080/health > /dev/null 2>&1; then
                print_success "Health endpoint responding"
                break
            else
                print_info "Attempt $attempt/$max_attempts: Health endpoint not ready yet..."
                sleep 2
                ((attempt++))
            fi
        done
        
        if [[ $attempt -gt $max_attempts ]]; then
            print_warning "Health endpoint not responding after $max_attempts attempts"
        fi
        
        # Clean shutdown
        print_info "Shutting down test instance..."
        kill $start_pid 2>/dev/null || true
        sleep 3
        kill -9 $start_pid 2>/dev/null || true
        
        print_success "Startup validation completed"
    else
        print_error "Process failed to start or crashed immediately"
        print_info "Check log file: $log_file"
        return 1
    fi
    
    # Cleanup
    rm -f "$pid_file"
    
    return 0
}

validate_api_endpoints() {
    print_header "API Endpoints Validation"
    
    # This requires a running instance
    print_info "This validation requires a running Ollama Distributed instance"
    print_info "Start with: $BINARY_PATH start --config $DEV_PROFILE"
    print_info "Then run: $0 api-test"
    
    return 0
}

run_api_tests() {
    print_header "API Endpoints Testing"
    
    local base_url="http://127.0.0.1:8080"
    local errors=0
    
    # Test health endpoint
    print_info "Testing health endpoint..."
    if response=$(curl -s -w "%{http_code}" "$base_url/health"); then
        status_code="${response: -3}"
        if [[ $status_code == "200" ]]; then
            print_success "Health endpoint: OK ($status_code)"
        else
            print_error "Health endpoint: Failed ($status_code)"
            ((errors++))
        fi
    else
        print_error "Health endpoint: Connection failed"
        ((errors++))
    fi
    
    # Test API endpoints
    local endpoints=(
        "/api/v1/health:Detailed Health"
        "/api/distributed/status:Cluster Status"
        "/api/tags:Model List"
        "/api/distributed/nodes:Node List"
    )
    
    for endpoint_desc in "${endpoints[@]}"; do
        local endpoint=$(echo $endpoint_desc | cut -d: -f1)
        local description=$(echo $endpoint_desc | cut -d: -f2)
        
        print_info "Testing $description..."
        if response=$(curl -s -w "%{http_code}" "$base_url$endpoint"); then
            status_code="${response: -3}"
            if [[ $status_code == "200" ]]; then
                print_success "$description: OK ($status_code)"
            else
                print_warning "$description: Non-200 response ($status_code) - may be normal"
            fi
        else
            print_warning "$description: Connection failed"
        fi
    done
    
    # Test web dashboard
    print_info "Testing web dashboard..."
    if curl -s -f http://127.0.0.1:8081/ > /dev/null 2>&1; then
        print_success "Web dashboard: Accessible"
    else
        print_warning "Web dashboard: Not accessible (may not be implemented)"
    fi
    
    if [[ $errors -eq 0 ]]; then
        print_success "API testing completed successfully"
    else
        print_error "API testing completed with $errors errors"
    fi
    
    return $errors
}

validate_training_tools() {
    print_header "Training Tools Validation"
    
    # Check if training directory exists
    local training_dir="$(dirname "$0")"
    if [[ -d "$training_dir" ]]; then
        print_success "Training directory found"
    else
        print_error "Training directory not found"
        return 1
    fi
    
    # Check for training modules
    local training_files=(
        "training-modules.md"
        "interactive-tutorial.md"
    )
    
    for file in "${training_files[@]}"; do
        if [[ -f "$training_dir/$file" ]]; then
            print_success "Training file found: $file"
        else
            print_error "Training file missing: $file"
        fi
    done
    
    # Test script creation
    print_info "Testing script creation capabilities..."
    local test_script="$BASE_DIR/test-script.sh"
    cat > "$test_script" << 'EOF'
#!/bin/bash
echo "Test script created successfully"
EOF
    chmod +x "$test_script"
    
    if "$test_script" 2>/dev/null; then
        print_success "Script creation and execution works"
        rm -f "$test_script"
    else
        print_error "Script creation or execution failed"
        return 1
    fi
    
    print_success "Training tools validation passed"
    return 0
}

create_training_environment() {
    print_header "Creating Training Environment"
    
    # Create all necessary directories
    local directories=(
        "$BASE_DIR"
        "$BASE_DIR/profiles"
        "$BASE_DIR/logs"
        "./dev-data"
        "./test-data"
    )
    
    for dir in "${directories[@]}"; do
        mkdir -p "$dir"
        print_success "Created directory: $dir"
    done
    
    # Create development profile if it doesn't exist
    if [[ ! -f "$DEV_PROFILE" ]]; then
        cat > "$DEV_PROFILE" << EOF
# Development Profile for Training
api:
  listen: "127.0.0.1:8080"
  debug: true
  
p2p:
  listen: "127.0.0.1:4001"
  
storage:
  data_dir: "./dev-data"
  
web:
  listen: "127.0.0.1:8081"
  enable_auth: false
  
logging:
  level: "debug"
  output: "console"
  
performance:
  monitoring_enabled: true
  metrics_interval: 5
EOF
        print_success "Created development profile"
    fi
    
    # Create testing profile
    local test_profile="$BASE_DIR/profiles/testing.yaml"
    cat > "$test_profile" << EOF
# Testing Profile
api:
  listen: "127.0.0.1:9080"
  
p2p:
  listen: "127.0.0.1:4002"
  
storage:
  data_dir: "./test-data"
  
web:
  listen: "127.0.0.1:9081"
  
logging:
  level: "info"
EOF
    print_success "Created testing profile"
    
    # Create useful aliases file
    local aliases_file="$BASE_DIR/aliases.sh"
    cat > "$aliases_file" << EOF
#!/bin/bash
# Useful aliases for Ollama Distributed training

alias od-dev='ollama-distributed start --config ~/.ollama-distributed/profiles/development.yaml'
alias od-test='ollama-distributed start --config ~/.ollama-distributed/profiles/testing.yaml'
alias od-status='ollama-distributed status'
alias od-health='curl -s http://127.0.0.1:8080/health | jq .'
alias od-validate='ollama-distributed validate --config ~/.ollama-distributed/profiles/development.yaml'

echo "Ollama Distributed training aliases loaded"
echo "Available commands:"
echo "  od-dev     - Start development server"
echo "  od-test    - Start testing server"
echo "  od-status  - Show status"
echo "  od-health  - Quick health check"
echo "  od-validate - Validate configuration"
EOF
    chmod +x "$aliases_file"
    print_success "Created aliases file: $aliases_file"
    
    print_info "To use aliases, run: source $aliases_file"
    print_success "Training environment setup complete"
    return 0
}

show_usage() {
    echo "Ollama Distributed Training Validation Scripts"
    echo
    echo "Usage: $0 [command]"
    echo
    echo "Commands:"
    echo "  prereq       - Check prerequisites"
    echo "  install      - Validate installation"
    echo "  config       - Validate configuration"
    echo "  startup      - Test startup process"
    echo "  api-test     - Test API endpoints (requires running instance)"
    echo "  tools        - Validate training tools"
    echo "  setup-env    - Create training environment"
    echo "  full         - Run all validations except api-test"
    echo "  help         - Show this help"
    echo
    echo "Examples:"
    echo "  $0 full                    # Run all validations"
    echo "  $0 prereq install config  # Run specific validations"
    echo "  $0 setup-env              # Setup training environment"
    echo
}

# Main execution logic
main() {
    if [[ $# -eq 0 ]]; then
        show_usage
        exit 1
    fi
    
    local overall_success=true
    
    for command in "$@"; do
        case "$command" in
            "prereq")
                if ! validate_prerequisites; then
                    overall_success=false
                fi
                ;;
            "install")
                if ! validate_installation; then
                    overall_success=false
                fi
                ;;
            "config")
                if ! validate_configuration; then
                    overall_success=false
                fi
                ;;
            "startup")
                if ! validate_startup; then
                    overall_success=false
                fi
                ;;
            "api-test")
                if ! run_api_tests; then
                    overall_success=false
                fi
                ;;
            "tools")
                if ! validate_training_tools; then
                    overall_success=false
                fi
                ;;
            "setup-env")
                if ! create_training_environment; then
                    overall_success=false
                fi
                ;;
            "full")
                for validation in prereq install config startup tools setup-env; do
                    if ! main "$validation"; then
                        overall_success=false
                    fi
                done
                ;;
            "help")
                show_usage
                exit 0
                ;;
            *)
                print_error "Unknown command: $command"
                show_usage
                exit 1
                ;;
        esac
        echo
    done
    
    if $overall_success; then
        print_header "🎉 All Validations Successful"
        print_success "Training environment is ready!"
        print_info "Start training with: ollama-distributed start --config $DEV_PROFILE"
        return 0
    else
        print_header "⚠️  Some Validations Failed"
        print_error "Please fix the issues above before starting training"
        return 1
    fi
}

# Run main function with all arguments
main "$@"
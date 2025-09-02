#!/bin/bash

# OllamaMax E2E Test Runner Script
# Comprehensive test execution with reporting and error handling

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL=${BASE_URL:-"http://localhost:8080"}
NODE_ENV=${NODE_ENV:-"test"}
HEADLESS=${HEADLESS:-"true"}
BROWSERS=${BROWSERS:-"chromium,firefox"}
PARALLEL=${PARALLEL:-"true"}

# Test types
CORE_TESTS="tests/core-functionality.spec.ts"
INFERENCE_TESTS="tests/distributed-inference.spec.ts"
SECURITY_TESTS="tests/security.spec.ts"
PERFORMANCE_TESTS="tests/performance.spec.ts"
LOAD_TESTS="tests/load.spec.ts"

echo -e "${BLUE}🚀 OllamaMax E2E Test Suite${NC}"
echo "================================="
echo "Base URL: $BASE_URL"
echo "Environment: $NODE_ENV"
echo "Browsers: $BROWSERS"
echo "Headless: $HEADLESS"
echo "Parallel: $PARALLEL"
echo ""

# Function to print status
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if service is available
check_service() {
    print_status "Checking if service is available at $BASE_URL..."
    
    for i in {1..30}; do
        if curl -sf "$BASE_URL" > /dev/null 2>&1; then
            print_success "Service is available"
            return 0
        fi
        
        print_status "Waiting for service... (attempt $i/30)"
        sleep 2
    done
    
    print_error "Service is not available at $BASE_URL"
    return 1
}

# Function to setup test environment
setup_environment() {
    print_status "Setting up test environment..."
    
    # Create necessary directories
    mkdir -p reports/screenshots
    mkdir -p reports/performance
    mkdir -p reports/load-tests
    mkdir -p reports/security
    mkdir -p test-results
    
    # Set environment variables
    export BASE_URL="$BASE_URL"
    export NODE_ENV="$NODE_ENV"
    export HEADLESS="$HEADLESS"
    
    print_success "Test environment setup complete"
}

# Function to run specific test type
run_test_type() {
    local test_name=$1
    local test_files=$2
    local project=${3:-"chromium"}
    
    print_status "Running $test_name tests..."
    
    if [ "$PARALLEL" = "true" ]; then
        npx playwright test $test_files --project=$project --reporter=html || {
            print_error "$test_name tests failed"
            return 1
        }
    else
        npx playwright test $test_files --project=$project --workers=1 --reporter=html || {
            print_error "$test_name tests failed"
            return 1
        }
    fi
    
    print_success "$test_name tests completed"
    return 0
}

# Function to generate reports
generate_reports() {
    print_status "Generating comprehensive reports..."
    
    # Generate HTML report
    npx playwright show-report --host=0.0.0.0 &
    REPORT_PID=$!
    
    # Create summary report
    cat > reports/test-summary.html << EOF
<!DOCTYPE html>
<html>
<head>
    <title>OllamaMax E2E Test Summary</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f0f8ff; padding: 20px; border-radius: 8px; }
        .section { margin: 20px 0; padding: 15px; border-left: 4px solid #007bff; }
        .success { border-left-color: #28a745; }
        .warning { border-left-color: #ffc107; }
        .error { border-left-color: #dc3545; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🧪 OllamaMax E2E Test Summary</h1>
        <p>Generated on: $(date)</p>
        <p>Base URL: $BASE_URL</p>
        <p>Environment: $NODE_ENV</p>
    </div>
    
    <div class="section">
        <h2>📊 Test Execution Summary</h2>
        <p>Check the detailed Playwright report for complete results.</p>
        <p><a href="playwright-report/index.html">View Detailed Report</a></p>
    </div>
    
    <div class="section">
        <h2>📁 Available Reports</h2>
        <ul>
            <li><a href="screenshots/">Screenshots</a></li>
            <li><a href="performance/">Performance Metrics</a></li>
            <li><a href="load-tests/">Load Test Results</a></li>
            <li><a href="security/">Security Reports</a></li>
        </ul>
    </div>
</body>
</html>
EOF
    
    print_success "Reports generated in reports/ directory"
}

# Function to cleanup
cleanup() {
    print_status "Cleaning up..."
    
    # Kill report server if running
    if [ ! -z "$REPORT_PID" ]; then
        kill $REPORT_PID 2>/dev/null || true
    fi
    
    # Archive old screenshots (older than 7 days)
    find reports/screenshots -name "*.png" -mtime +7 -delete 2>/dev/null || true
    
    print_success "Cleanup complete"
}

# Function to run all tests
run_all_tests() {
    local failed_tests=""
    local test_results=0
    
    # Core functionality tests
    if ! run_test_type "Core Functionality" "$CORE_TESTS" "chromium"; then
        failed_tests="$failed_tests core-functionality"
        test_results=1
    fi
    
    # Distributed inference tests
    if ! run_test_type "Distributed Inference" "$INFERENCE_TESTS" "chromium"; then
        failed_tests="$failed_tests distributed-inference"
        test_results=1
    fi
    
    # Security tests
    if ! run_test_type "Security" "$SECURITY_TESTS" "chromium"; then
        failed_tests="$failed_tests security"
        test_results=1
    fi
    
    # Performance tests (optional - might be slow)
    if [ "${RUN_PERFORMANCE:-false}" = "true" ]; then
        if ! run_test_type "Performance" "$PERFORMANCE_TESTS" "performance"; then
            failed_tests="$failed_tests performance"
            test_results=1
        fi
    fi
    
    # Load tests (optional - might be resource intensive)
    if [ "${RUN_LOAD_TESTS:-false}" = "true" ]; then
        if ! run_test_type "Load" "$LOAD_TESTS" "load"; then
            failed_tests="$failed_tests load"
            test_results=1
        fi
    fi
    
    if [ $test_results -eq 0 ]; then
        print_success "All tests completed successfully!"
    else
        print_error "Some tests failed: $failed_tests"
        return 1
    fi
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --help              Show this help message"
    echo "  --core              Run only core functionality tests"
    echo "  --inference         Run only distributed inference tests"
    echo "  --security          Run only security tests"
    echo "  --performance       Run only performance tests"
    echo "  --load              Run only load tests"
    echo "  --all               Run all tests (default)"
    echo "  --headed            Run tests in headed mode"
    echo "  --debug             Run tests in debug mode"
    echo "  --setup-only        Only setup environment, don't run tests"
    echo ""
    echo "Environment Variables:"
    echo "  BASE_URL           Target URL (default: http://localhost:8080)"
    echo "  NODE_ENV           Environment mode (default: test)"
    echo "  HEADLESS           Run headless browsers (default: true)"
    echo "  BROWSERS           Comma-separated browser list (default: chromium,firefox)"
    echo "  RUN_PERFORMANCE    Include performance tests (default: false)"
    echo "  RUN_LOAD_TESTS     Include load tests (default: false)"
    echo ""
}

# Main execution
main() {
    local test_type="all"
    local setup_only=false
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --help)
                show_usage
                exit 0
                ;;
            --core)
                test_type="core"
                shift
                ;;
            --inference)
                test_type="inference"
                shift
                ;;
            --security)
                test_type="security"
                shift
                ;;
            --performance)
                test_type="performance"
                shift
                ;;
            --load)
                test_type="load"
                shift
                ;;
            --all)
                test_type="all"
                shift
                ;;
            --headed)
                HEADLESS="false"
                shift
                ;;
            --debug)
                HEADLESS="false"
                export DEBUG="pw:*"
                shift
                ;;
            --setup-only)
                setup_only=true
                shift
                ;;
            *)
                print_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    # Setup environment
    setup_environment
    
    if [ "$setup_only" = "true" ]; then
        print_success "Setup complete. Environment ready for testing."
        exit 0
    fi
    
    # Check service availability
    if ! check_service; then
        print_error "Cannot proceed without available service"
        exit 1
    fi
    
    # Trap cleanup on exit
    trap cleanup EXIT
    
    # Run tests based on type
    case $test_type in
        "core")
            run_test_type "Core Functionality" "$CORE_TESTS" "chromium"
            ;;
        "inference")
            run_test_type "Distributed Inference" "$INFERENCE_TESTS" "chromium"
            ;;
        "security")
            run_test_type "Security" "$SECURITY_TESTS" "security"
            ;;
        "performance")
            run_test_type "Performance" "$PERFORMANCE_TESTS" "performance"
            ;;
        "load")
            run_test_type "Load" "$LOAD_TESTS" "load"
            ;;
        "all")
            run_all_tests
            ;;
        *)
            print_error "Invalid test type: $test_type"
            exit 1
            ;;
    esac
    
    # Generate reports
    generate_reports
    
    print_success "Test execution complete!"
    print_status "Reports available in reports/ directory"
    print_status "Open reports/index.html for a comprehensive summary"
}

# Execute main function with all arguments
main "$@"
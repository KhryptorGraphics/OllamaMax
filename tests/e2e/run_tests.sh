#!/bin/bash

# E2E Test Runner for ollamamax
# Runs comprehensive end-to-end tests including Playwright and API validation

set -e

echo "🚀 Starting ollamamax E2E Test Suite"

# Configuration
BASE_URL=${BASE_URL:-"http://localhost:8080"}
TEST_TIMEOUT=${TEST_TIMEOUT:-30000}
REPORTS_DIR="./reports"
SCREENSHOTS_DIR="$REPORTS_DIR/screenshots"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Clean up function
cleanup() {
    log_info "Cleaning up test environment..."
    if [ -n "$SERVER_PID" ]; then
        log_info "Stopping test server (PID: $SERVER_PID)"
        kill $SERVER_PID 2>/dev/null || true
    fi
}

# Trap to ensure cleanup on exit
trap cleanup EXIT

# Create reports directory
log_info "Creating reports directory: $REPORTS_DIR"
mkdir -p "$REPORTS_DIR"
mkdir -p "$SCREENSHOTS_DIR"

# Check if Node.js and npm are available
if ! command -v node &> /dev/null; then
    log_error "Node.js is not installed. Please install Node.js to run E2E tests."
    exit 1
fi

if ! command -v npm &> /dev/null; then
    log_error "npm is not installed. Please install npm to run E2E tests."
    exit 1
fi

# Install dependencies if needed
if [ ! -d "node_modules" ]; then
    log_info "Installing E2E test dependencies..."
    npm install
fi

# Health check function
check_service_health() {
    local url="$1"
    local max_attempts=30
    local attempt=0
    
    log_info "Checking service health at $url"
    
    while [ $attempt -lt $max_attempts ]; do
        if curl -s -f "$url/api/v1/health" > /dev/null 2>&1; then
            log_success "Service is healthy at $url"
            return 0
        fi
        
        attempt=$((attempt + 1))
        log_info "Attempt $attempt/$max_attempts: Service not ready, waiting..."
        sleep 2
    done
    
    log_error "Service health check failed after $max_attempts attempts"
    return 1
}

# Start test server if running locally
start_test_server() {
    if [[ "$BASE_URL" == *"localhost"* ]]; then
        log_info "Starting local test server..."
        
        # Check if we have a test server script
        if [ -f "../../scripts/start_test_server.sh" ]; then
            bash ../../scripts/start_test_server.sh &
            SERVER_PID=$!
            log_info "Test server started with PID: $SERVER_PID"
            
            # Wait for server to be ready
            sleep 5
        else
            log_warning "No test server script found. Assuming service is already running."
        fi
    fi
}

# Run API validation tests
run_api_tests() {
    log_info "Running API validation tests..."
    
    # Basic API health check
    if ! curl -s -f "$BASE_URL/api/v1/health" > /dev/null; then
        log_error "API health check failed"
        return 1
    fi
    
    log_success "API health check passed"
    
    # Test API endpoints
    endpoints=(
        "/api/v1/health"
        "/api/v1/models"
        "/api/v1/metrics"
    )
    
    for endpoint in "${endpoints[@]}"; do
        log_info "Testing endpoint: $endpoint"
        
        status_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL$endpoint")
        
        if [ "$status_code" -eq 200 ]; then
            log_success "✓ $endpoint returned $status_code"
        else
            log_warning "⚠ $endpoint returned $status_code (expected 200)"
        fi
    done
}

# Run Playwright tests
run_playwright_tests() {
    log_info "Running Playwright E2E tests..."
    
    # Set environment variables for tests
    export BASE_URL="$BASE_URL"
    export TEST_TIMEOUT="$TEST_TIMEOUT"
    
    # Check if Playwright is installed
    if [ ! -d "node_modules/@playwright" ]; then
        log_warning "Playwright not found, installing..."
        npm install @playwright/test
        npx playwright install
    fi
    
    # Run the tests
    if npx jest --config=jest.config.js --verbose --no-cache; then
        log_success "Playwright tests completed successfully"
        return 0
    else
        log_error "Playwright tests failed"
        return 1
    fi
}

# Run security tests
run_security_tests() {
    log_info "Running basic security tests..."
    
    # Test for common security headers
    headers_response=$(curl -s -I "$BASE_URL/api/v1/health")
    
    if echo "$headers_response" | grep -i "x-content-type-options" > /dev/null; then
        log_success "✓ X-Content-Type-Options header present"
    else
        log_warning "⚠ X-Content-Type-Options header missing"
    fi
    
    if echo "$headers_response" | grep -i "x-frame-options" > /dev/null; then
        log_success "✓ X-Frame-Options header present"
    else
        log_warning "⚠ X-Frame-Options header missing"
    fi
    
    # Test for CORS
    cors_response=$(curl -s -H "Origin: http://evil.com" -I "$BASE_URL/api/v1/health")
    
    if echo "$cors_response" | grep -i "access-control-allow-origin" > /dev/null; then
        log_info "CORS headers detected - verify they're properly configured"
    else
        log_info "No CORS headers detected"
    fi
}

# Performance tests
run_performance_tests() {
    log_info "Running basic performance tests..."
    
    # Simple response time test
    start_time=$(date +%s%N)
    curl -s "$BASE_URL/api/v1/health" > /dev/null
    end_time=$(date +%s%N)
    
    duration=$(( (end_time - start_time) / 1000000 )) # Convert to milliseconds
    
    log_info "Health endpoint response time: ${duration}ms"
    
    if [ $duration -lt 1000 ]; then
        log_success "✓ Response time is acceptable (<1000ms)"
    else
        log_warning "⚠ Response time is slow (>1000ms)"
    fi
}

# Generate test report
generate_report() {
    local test_status="$1"
    local report_file="$REPORTS_DIR/e2e-test-report.json"
    
    log_info "Generating test report: $report_file"
    
    cat > "$report_file" << EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "base_url": "$BASE_URL",
  "test_status": "$test_status",
  "environment": {
    "node_version": "$(node --version 2>/dev/null || echo 'not available')",
    "npm_version": "$(npm --version 2>/dev/null || echo 'not available')",
    "os": "$(uname -s)",
    "arch": "$(uname -m)"
  },
  "test_results": {
    "api_tests": "completed",
    "playwright_tests": "$test_status",
    "security_tests": "completed",
    "performance_tests": "completed"
  },
  "artifacts": {
    "screenshots": "$SCREENSHOTS_DIR",
    "logs": "$REPORTS_DIR"
  }
}
EOF
    
    log_success "Test report generated: $report_file"
}

# Main execution
main() {
    log_info "ollamamax E2E Test Suite"
    log_info "=========================="
    log_info "Base URL: $BASE_URL"
    log_info "Timeout: $TEST_TIMEOUT ms"
    log_info "Reports: $REPORTS_DIR"
    
    # Start test server if needed
    start_test_server
    
    # Wait for service to be ready
    if ! check_service_health "$BASE_URL"; then
        log_error "Service is not available. Exiting."
        generate_report "failed"
        exit 1
    fi
    
    # Run test suites
    test_failed=false
    
    # API Tests
    if ! run_api_tests; then
        test_failed=true
    fi
    
    # Playwright Tests
    if ! run_playwright_tests; then
        test_failed=true
    fi
    
    # Security Tests
    run_security_tests
    
    # Performance Tests
    run_performance_tests
    
    # Generate final report
    if [ "$test_failed" = true ]; then
        log_error "Some tests failed. Check the reports for details."
        generate_report "failed"
        exit 1
    else
        log_success "All tests completed successfully!"
        generate_report "passed"
        exit 0
    fi
}

# Run main function
main "$@"
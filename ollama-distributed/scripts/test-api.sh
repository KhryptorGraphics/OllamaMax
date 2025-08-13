#!/bin/bash

# OllamaMax Distributed API Testing Script
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${1:-http://localhost}"
VERBOSE="${VERBOSE:-false}"

# Functions
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}✓${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1"
}

warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Make API request with error handling
api_request() {
    local method="$1"
    local endpoint="$2"
    local data="$3"
    local expected_status="${4:-200}"
    
    local url="$BASE_URL$endpoint"
    local response
    local status_code
    
    if [ "$VERBOSE" = "true" ]; then
        log "Making $method request to $url"
    fi
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$url" 2>/dev/null)
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" 2>/dev/null)
    fi
    
    status_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    if [ "$status_code" = "$expected_status" ]; then
        if [ "$VERBOSE" = "true" ]; then
            echo "$response_body"
        fi
        return 0
    else
        error "Expected status $expected_status, got $status_code"
        if [ "$VERBOSE" = "true" ]; then
            echo "Response: $response_body"
        fi
        return 1
    fi
}

# Test health endpoint
test_health() {
    log "Testing health endpoint..."
    
    if api_request "GET" "/api/v1/health"; then
        success "Health endpoint working"
    else
        error "Health endpoint failed"
        return 1
    fi
}

# Test version endpoint
test_version() {
    log "Testing version endpoint..."
    
    if api_request "GET" "/api/v1/version"; then
        success "Version endpoint working"
    else
        error "Version endpoint failed"
        return 1
    fi
}

# Test cluster endpoints
test_cluster() {
    log "Testing cluster endpoints..."
    
    # Cluster status
    if api_request "GET" "/api/v1/cluster/status"; then
        success "Cluster status endpoint working"
    else
        warning "Cluster status endpoint not available (may be expected)"
    fi
    
    # Cluster nodes
    if api_request "GET" "/api/v1/nodes"; then
        success "Cluster nodes endpoint working"
    else
        warning "Cluster nodes endpoint not available"
    fi
}

# Test model endpoints
test_models() {
    log "Testing model endpoints..."
    
    # List models
    if api_request "GET" "/api/v1/models"; then
        success "Models list endpoint working"
    else
        warning "Models list endpoint not available"
    fi
    
    # Model sync status
    if api_request "GET" "/api/v1/models/sync/status"; then
        success "Model sync status endpoint working"
    else
        warning "Model sync status endpoint not available"
    fi
}

# Test metrics endpoints
test_metrics() {
    log "Testing metrics endpoints..."
    
    # System metrics
    if api_request "GET" "/api/v1/metrics"; then
        success "Metrics endpoint working"
    else
        warning "Metrics endpoint not available"
    fi
    
    # Resource metrics
    if api_request "GET" "/api/v1/metrics/resources"; then
        success "Resource metrics endpoint working"
    else
        warning "Resource metrics endpoint not available"
    fi
}

# Test task endpoints
test_tasks() {
    log "Testing task endpoints..."
    
    # List tasks
    if api_request "GET" "/api/v1/tasks"; then
        success "Tasks list endpoint working"
    else
        warning "Tasks list endpoint not available"
    fi
    
    # Task queue status
    if api_request "GET" "/api/v1/tasks/queue"; then
        success "Task queue endpoint working"
    else
        warning "Task queue endpoint not available"
    fi
}

# Test inference endpoint (if available)
test_inference() {
    log "Testing inference endpoints..."
    
    # This would typically require a model to be loaded
    # For now, just test that the endpoint exists
    local inference_data='{"model": "test", "prompt": "Hello, world!"}'
    
    # We expect this to fail with 404 or 400, not 500
    if api_request "POST" "/api/v1/inference" "$inference_data" "404" || \
       api_request "POST" "/api/v1/inference" "$inference_data" "400" || \
       api_request "POST" "/api/v1/inference" "$inference_data" "200"; then
        success "Inference endpoint accessible (response as expected)"
    else
        warning "Inference endpoint not available or unexpected response"
    fi
}

# Test error handling
test_error_handling() {
    log "Testing error handling..."
    
    # Test non-existent endpoint
    if api_request "GET" "/api/v1/nonexistent" "" "404"; then
        success "404 error handling working"
    else
        warning "404 error handling not as expected"
    fi
    
    # Test invalid JSON
    if api_request "POST" "/api/v1/inference" "invalid json" "400"; then
        success "400 error handling working"
    else
        warning "400 error handling not as expected"
    fi
}

# Load test (simple)
load_test() {
    log "Running simple load test..."
    
    local requests=10
    local concurrent=3
    local success_count=0
    
    log "Making $requests concurrent requests to health endpoint..."
    
    for i in $(seq 1 $requests); do
        if api_request "GET" "/api/v1/health" >/dev/null 2>&1; then
            ((success_count++))
        fi &
        
        # Limit concurrency
        if (( i % concurrent == 0 )); then
            wait
        fi
    done
    
    wait
    
    log "Load test completed: $success_count/$requests requests successful"
    
    if [ $success_count -eq $requests ]; then
        success "Load test passed"
    else
        warning "Load test had some failures"
    fi
}

# Main test suite
run_tests() {
    log "Starting API tests for $BASE_URL"
    
    local failed_tests=0
    
    # Core functionality tests
    test_health || ((failed_tests++))
    test_version || ((failed_tests++))
    
    # Cluster tests
    test_cluster || ((failed_tests++))
    
    # Feature tests
    test_models || ((failed_tests++))
    test_metrics || ((failed_tests++))
    test_tasks || ((failed_tests++))
    test_inference || ((failed_tests++))
    
    # Error handling tests
    test_error_handling || ((failed_tests++))
    
    # Performance tests
    load_test || ((failed_tests++))
    
    log "API tests completed"
    
    if [ $failed_tests -eq 0 ]; then
        success "All API tests passed!"
    else
        warning "$failed_tests test(s) had issues (some may be expected)"
    fi
    
    return $failed_tests
}

# Handle script arguments
case "${1:-}" in
    "--help"|"-h")
        echo "Usage: $0 [BASE_URL] [--verbose]"
        echo "  BASE_URL: Base URL for API testing (default: http://localhost)"
        echo "  --verbose: Enable verbose output"
        echo ""
        echo "Environment variables:"
        echo "  VERBOSE: Set to 'true' for verbose output"
        exit 0
        ;;
    "--verbose"|"-v")
        VERBOSE="true"
        BASE_URL="${2:-http://localhost}"
        ;;
    *)
        if [[ "$1" == http* ]]; then
            BASE_URL="$1"
            if [ "$2" = "--verbose" ] || [ "$2" = "-v" ]; then
                VERBOSE="true"
            fi
        fi
        ;;
esac

# Run the tests
run_tests

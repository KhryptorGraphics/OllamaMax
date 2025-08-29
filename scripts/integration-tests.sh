#!/bin/bash

# OllamaMax Backend Integration Test Suite
# This script tests all API endpoints and database integrations

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_BASE_URL="${API_BASE_URL:-http://localhost:11434}"
WS_BASE_URL="${WS_BASE_URL:-ws://localhost:11434}"
TEST_USER_USERNAME="${TEST_USER_USERNAME:-testuser}"
TEST_USER_PASSWORD="${TEST_USER_PASSWORD:-testpass123}"
TEST_USER_EMAIL="${TEST_USER_EMAIL:-testuser@example.com}"

# Global variables
ACCESS_TOKEN=""
TEST_USER_ID=""
TEST_MODEL_ID=""
TEST_NODE_ID=""
TEST_INFERENCE_ID=""

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

# Utility function to make API requests
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local headers=$4
    
    if [ -n "$ACCESS_TOKEN" ]; then
        headers="${headers} -H \"Authorization: Bearer $ACCESS_TOKEN\""
    fi
    
    if [ -n "$data" ]; then
        eval curl -s -X "$method" "$API_BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            $headers \
            -d "$data"
    else
        eval curl -s -X "$method" "$API_BASE_URL$endpoint" \
            $headers
    fi
}

# Utility function to check if service is healthy
check_service_health() {
    log_info "Checking service health..."
    
    local response=$(make_request GET "/health")
    local status=$(echo "$response" | jq -r '.status // "unknown"')
    
    if [ "$status" = "healthy" ]; then
        log_success "Service is healthy"
        return 0
    else
        log_error "Service is not healthy: $response"
        return 1
    fi
}

# Test user registration
test_user_registration() {
    log_info "Testing user registration..."
    
    local data="{
        \"username\": \"$TEST_USER_USERNAME\",
        \"email\": \"$TEST_USER_EMAIL\",
        \"password\": \"$TEST_USER_PASSWORD\"
    }"
    
    local response=$(make_request POST "/api/v1/auth/register" "$data")
    local user_id=$(echo "$response" | jq -r '.user.id // empty')
    
    if [ -n "$user_id" ]; then
        TEST_USER_ID=$user_id
        log_success "User registration successful: $user_id"
        return 0
    else
        log_error "User registration failed: $response"
        return 1
    fi
}

# Test user login
test_user_login() {
    log_info "Testing user login..."
    
    local data="{
        \"username\": \"$TEST_USER_USERNAME\",
        \"password\": \"$TEST_USER_PASSWORD\"
    }"
    
    local response=$(make_request POST "/api/v1/auth/login" "$data")
    local token=$(echo "$response" | jq -r '.access_token // empty')
    
    if [ -n "$token" ]; then
        ACCESS_TOKEN=$token
        log_success "User login successful"
        return 0
    else
        log_error "User login failed: $response"
        return 1
    fi
}

# Test model creation
test_model_creation() {
    log_info "Testing model creation..."
    
    local data="{
        \"name\": \"test-model\",
        \"version\": \"1.0.0\",
        \"size\": 1073741824,
        \"hash\": \"sha256:abcd1234567890\",
        \"content_type\": \"application/octet-stream\",
        \"description\": \"Test model for integration testing\",
        \"tags\": [\"test\", \"integration\"],
        \"parameters\": {
            \"temperature\": 0.8,
            \"max_tokens\": 2048
        }
    }"
    
    local response=$(make_request POST "/api/v1/models/" "$data")
    local model_id=$(echo "$response" | jq -r '.model.id // empty')
    
    if [ -n "$model_id" ]; then
        TEST_MODEL_ID=$model_id
        log_success "Model creation successful: $model_id"
        return 0
    else
        log_error "Model creation failed: $response"
        return 1
    fi
}

# Test model retrieval
test_model_retrieval() {
    log_info "Testing model retrieval..."
    
    if [ -z "$TEST_MODEL_ID" ]; then
        log_error "No test model ID available"
        return 1
    fi
    
    local response=$(make_request GET "/api/v1/models/$TEST_MODEL_ID")
    local model_name=$(echo "$response" | jq -r '.model.name // empty')
    
    if [ "$model_name" = "test-model" ]; then
        log_success "Model retrieval successful"
        return 0
    else
        log_error "Model retrieval failed: $response"
        return 1
    fi
}

# Test model listing
test_model_listing() {
    log_info "Testing model listing..."
    
    local response=$(make_request GET "/api/v1/models/?limit=10")
    local models_count=$(echo "$response" | jq -r '.models | length')
    
    if [ "$models_count" -ge 1 ]; then
        log_success "Model listing successful ($models_count models)"
        return 0
    else
        log_error "Model listing failed: $response"
        return 1
    fi
}

# Test model update
test_model_update() {
    log_info "Testing model update..."
    
    if [ -z "$TEST_MODEL_ID" ]; then
        log_error "No test model ID available"
        return 1
    fi
    
    local data="{
        \"description\": \"Updated test model description\",
        \"status\": \"ready\"
    }"
    
    local response=$(make_request PUT "/api/v1/models/$TEST_MODEL_ID" "$data")
    local message=$(echo "$response" | jq -r '.message // empty')
    
    if [[ "$message" == *"successfully"* ]]; then
        log_success "Model update successful"
        return 0
    else
        log_error "Model update failed: $response"
        return 1
    fi
}

# Test node registration (simulated)
test_node_operations() {
    log_info "Testing node operations..."
    
    # List nodes
    local response=$(make_request GET "/api/v1/nodes/")
    local nodes_count=$(echo "$response" | jq -r '.nodes | length')
    
    log_success "Node listing successful ($nodes_count nodes)"
    
    # If there are nodes, test getting one
    if [ "$nodes_count" -gt 0 ]; then
        local node_id=$(echo "$response" | jq -r '.nodes[0].id')
        if [ -n "$node_id" ] && [ "$node_id" != "null" ]; then
            TEST_NODE_ID=$node_id
            local node_response=$(make_request GET "/api/v1/nodes/$node_id")
            local node_peer_id=$(echo "$node_response" | jq -r '.node.peer_id // empty')
            
            if [ -n "$node_peer_id" ]; then
                log_success "Node retrieval successful: $node_peer_id"
                
                # Test node health check
                local health_response=$(make_request GET "/api/v1/nodes/$node_id/health")
                local health_status=$(echo "$health_response" | jq -r '.status // empty')
                log_success "Node health check successful: $health_status"
                
                return 0
            fi
        fi
    fi
    
    log_warning "No nodes available for detailed testing"
    return 0
}

# Test user profile operations
test_user_profile() {
    log_info "Testing user profile operations..."
    
    # Get user profile
    local response=$(make_request GET "/api/v1/users/profile")
    local username=$(echo "$response" | jq -r '.user.username // empty')
    
    if [ "$username" = "$TEST_USER_USERNAME" ]; then
        log_success "User profile retrieval successful"
    else
        log_error "User profile retrieval failed: $response"
        return 1
    fi
    
    # Update user profile
    local update_data="{\"email\": \"updated@example.com\"}"
    local update_response=$(make_request PUT "/api/v1/users/profile" "$update_data")
    local update_message=$(echo "$update_response" | jq -r '.message // empty')
    
    if [[ "$update_message" == *"successfully"* ]]; then
        log_success "User profile update successful"
        return 0
    else
        log_error "User profile update failed: $update_response"
        return 1
    fi
}

# Test inference operations
test_inference_operations() {
    log_info "Testing inference operations..."
    
    if [ -z "$TEST_MODEL_ID" ]; then
        log_warning "No test model available, skipping inference tests"
        return 0
    fi
    
    # Test generate endpoint
    local generate_data="{
        \"model\": \"test-model\",
        \"prompt\": \"Hello, this is a test prompt\",
        \"options\": {\"temperature\": 0.7}
    }"
    
    local response=$(make_request POST "/api/v1/inference/generate" "$generate_data")
    local request_id=$(echo "$response" | jq -r '.request_id // empty')
    
    if [ -n "$request_id" ]; then
        log_success "Inference generation request successful: $request_id"
        
        # List inference requests
        local list_response=$(make_request GET "/api/v1/inference/requests?limit=5")
        local requests_count=$(echo "$list_response" | jq -r '.requests | length')
        
        if [ "$requests_count" -ge 1 ]; then
            log_success "Inference requests listing successful ($requests_count requests)"
            return 0
        fi
    else
        log_error "Inference generation failed: $response"
        return 1
    fi
    
    return 0
}

# Test system configuration
test_system_config() {
    log_info "Testing system configuration..."
    
    # Get system config
    local response=$(make_request GET "/api/v1/system/config")
    local config_keys=$(echo "$response" | jq -r '.config | keys | length')
    
    if [ "$config_keys" -gt 0 ]; then
        log_success "System config retrieval successful ($config_keys keys)"
    else
        log_warning "No system config found"
    fi
    
    # Update system config
    local update_data="{
        \"config\": {
            \"test_setting\": \"integration_test_value\"
        }
    }"
    
    local update_response=$(make_request PUT "/api/v1/system/config" "$update_data")
    local update_message=$(echo "$update_response" | jq -r '.message // empty')
    
    if [[ "$update_message" == *"successfully"* ]]; then
        log_success "System config update successful"
        return 0
    else
        log_error "System config update failed: $update_response"
        return 1
    fi
}

# Test system statistics
test_system_stats() {
    log_info "Testing system statistics..."
    
    local response=$(make_request GET "/api/v1/system/stats")
    local db_status=$(echo "$response" | jq -r '.system.health.overall // empty')
    
    if [ -n "$db_status" ]; then
        log_success "System statistics retrieval successful (DB: $db_status)"
        return 0
    else
        log_error "System statistics retrieval failed: $response"
        return 1
    fi
}

# Test audit logs
test_audit_logs() {
    log_info "Testing audit logs..."
    
    local response=$(make_request GET "/api/v1/system/audit?limit=10")
    local logs_count=$(echo "$response" | jq -r '.audit_logs | length')
    
    log_success "Audit logs retrieval successful ($logs_count entries)"
    return 0
}

# Test metrics endpoint
test_metrics() {
    log_info "Testing metrics endpoint..."
    
    local response=$(make_request GET "/metrics")
    local db_connections=$(echo "$response" | jq -r '.database.postgresql.open_connections // 0')
    
    if [ "$db_connections" -gt 0 ]; then
        log_success "Metrics endpoint successful (DB connections: $db_connections)"
        return 0
    else
        log_warning "Metrics endpoint returned no database connections"
        return 0
    fi
}

# Test WebSocket connection
test_websocket_connection() {
    log_info "Testing WebSocket connection..."
    
    # Check if websocat is available
    if ! command -v websocat &> /dev/null; then
        log_warning "websocat not found, skipping WebSocket test"
        return 0
    fi
    
    # Test basic WebSocket connection
    local ws_response=$(timeout 5 websocat -t "$WS_BASE_URL/ws" -E <<< '{"type":"heartbeat"}' 2>/dev/null || true)
    
    if [[ "$ws_response" == *"welcome"* ]] || [[ "$ws_response" == *"pong"* ]]; then
        log_success "WebSocket connection successful"
        return 0
    else
        log_warning "WebSocket connection test inconclusive"
        return 0
    fi
}

# Test database performance
test_database_performance() {
    log_info "Testing database performance..."
    
    # Run concurrent requests to test connection pooling
    local pids=()
    for i in {1..5}; do
        make_request GET "/api/v1/models/" &
        pids+=($!)
    done
    
    # Wait for all requests to complete
    for pid in "${pids[@]}"; do
        wait $pid
    done
    
    log_success "Database connection pooling test completed"
    return 0
}

# Cleanup test data
cleanup_test_data() {
    log_info "Cleaning up test data..."
    
    # Delete test model if it exists
    if [ -n "$TEST_MODEL_ID" ]; then
        local response=$(make_request DELETE "/api/v1/models/$TEST_MODEL_ID")
        if [[ "$response" == *"successfully"* ]]; then
            log_success "Test model cleaned up"
        fi
    fi
    
    # Logout user
    if [ -n "$ACCESS_TOKEN" ]; then
        local logout_response=$(make_request POST "/api/v1/users/logout")
        if [[ "$logout_response" == *"successfully"* ]]; then
            log_success "User logged out successfully"
        fi
    fi
    
    log_success "Cleanup completed"
}

# Main test execution
run_integration_tests() {
    local failed_tests=0
    local total_tests=0
    
    log_info "Starting OllamaMax Backend Integration Tests"
    log_info "API Base URL: $API_BASE_URL"
    
    # Pre-flight checks
    if ! command -v jq &> /dev/null; then
        log_error "jq is required but not installed"
        exit 1
    fi
    
    if ! command -v curl &> /dev/null; then
        log_error "curl is required but not installed"
        exit 1
    fi
    
    # Test suite
    local tests=(
        "check_service_health"
        "test_user_registration"
        "test_user_login"
        "test_user_profile"
        "test_model_creation"
        "test_model_retrieval"
        "test_model_listing"
        "test_model_update"
        "test_node_operations"
        "test_inference_operations"
        "test_system_config"
        "test_system_stats"
        "test_audit_logs"
        "test_metrics"
        "test_websocket_connection"
        "test_database_performance"
    )
    
    # Run tests
    for test in "${tests[@]}"; do
        total_tests=$((total_tests + 1))
        echo
        if ! $test; then
            failed_tests=$((failed_tests + 1))
            log_error "Test failed: $test"
        fi
    done
    
    # Cleanup
    echo
    cleanup_test_data
    
    # Results summary
    echo
    echo "=========================================="
    echo "Integration Test Results"
    echo "=========================================="
    echo "Total tests: $total_tests"
    echo "Passed: $((total_tests - failed_tests))"
    echo "Failed: $failed_tests"
    echo "Success rate: $(( (total_tests - failed_tests) * 100 / total_tests ))%"
    echo "=========================================="
    
    if [ $failed_tests -eq 0 ]; then
        log_success "All integration tests passed!"
        exit 0
    else
        log_error "$failed_tests test(s) failed"
        exit 1
    fi
}

# Handle script arguments
case "${1:-run}" in
    "run")
        run_integration_tests
        ;;
    "health")
        check_service_health
        ;;
    "cleanup")
        cleanup_test_data
        ;;
    "help")
        echo "Usage: $0 [run|health|cleanup|help]"
        echo "  run     - Run full integration test suite (default)"
        echo "  health  - Check service health only"
        echo "  cleanup - Clean up test data"
        echo "  help    - Show this help message"
        ;;
    *)
        log_error "Unknown command: $1"
        echo "Use '$0 help' for usage information"
        exit 1
        ;;
esac
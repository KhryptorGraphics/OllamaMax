#!/bin/bash

# OllamaMax Integration Test Runner
# This script runs comprehensive integration tests for the OllamaMax system

set -e

echo "🧪 OllamaMax Integration Test Runner"
echo "===================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "SUCCESS")
            echo -e "${GREEN}✅ $message${NC}"
            ;;
        "ERROR")
            echo -e "${RED}❌ $message${NC}"
            ;;
        "WARNING")
            echo -e "${YELLOW}⚠️  $message${NC}"
            ;;
        "INFO")
            echo -e "${BLUE}ℹ️  $message${NC}"
            ;;
    esac
}

# Configuration
BINARY_PATH="./ollama-distributed"
TEST_TIMEOUT="5m"
VERBOSE=${VERBOSE:-false}

print_status "INFO" "Starting integration test suite..."

# Step 1: Check prerequisites
print_status "INFO" "Checking prerequisites..."

# Check if binary exists
if [ ! -f "$BINARY_PATH" ]; then
    print_status "ERROR" "Binary not found at $BINARY_PATH"
    echo "Please build the binary first:"
    echo "  go build -o ollama-distributed ./cmd/node"
    echo "  OR"
    echo "  docker-compose -f docker-compose.build.yml up ollama-build"
    exit 1
fi

print_status "SUCCESS" "Binary found at $BINARY_PATH"

# Check if Go is available for tests
if ! command -v go >/dev/null 2>&1; then
    print_status "ERROR" "Go not found, required for running tests"
    exit 1
fi

print_status "SUCCESS" "Go found: $(go version)"

# Step 2: Prepare test environment
print_status "INFO" "Preparing test environment..."

# Kill any existing processes on test ports
pkill -f "ollama-distributed.*start" || true
sleep 2

# Clean up any test artifacts
rm -f /tmp/integration-test-*.log

print_status "SUCCESS" "Test environment prepared"

# Step 3: Run unit tests first
print_status "INFO" "Running unit tests..."

if $VERBOSE; then
    go test -v ./cmd/node -run TestProxy
else
    go test ./cmd/node -run TestProxy
fi

if [ $? -eq 0 ]; then
    print_status "SUCCESS" "Unit tests passed"
else
    print_status "ERROR" "Unit tests failed"
    exit 1
fi

# Step 4: Run integration tests
print_status "INFO" "Running integration tests..."

# Set test environment variables
export INTEGRATION_TEST_BINARY="$BINARY_PATH"
export INTEGRATION_TEST_API_URL="http://localhost:8080"

# Run integration tests with timeout
if $VERBOSE; then
    timeout $TEST_TIMEOUT go test -v ./tests/integration -run TestComprehensiveIntegration
else
    timeout $TEST_TIMEOUT go test ./tests/integration -run TestComprehensiveIntegration
fi

INTEGRATION_RESULT=$?

if [ $INTEGRATION_RESULT -eq 0 ]; then
    print_status "SUCCESS" "Integration tests passed"
else
    print_status "ERROR" "Integration tests failed"
fi

# Step 5: Run user workflow tests
print_status "INFO" "Running user workflow tests..."

if $VERBOSE; then
    timeout $TEST_TIMEOUT go test -v ./tests/integration -run TestUserWorkflows
else
    timeout $TEST_TIMEOUT go test ./tests/integration -run TestUserWorkflows
fi

WORKFLOW_RESULT=$?

if [ $WORKFLOW_RESULT -eq 0 ]; then
    print_status "SUCCESS" "User workflow tests passed"
else
    print_status "ERROR" "User workflow tests failed"
fi

# Step 6: Run performance tests
print_status "INFO" "Running performance tests..."

if $VERBOSE; then
    timeout $TEST_TIMEOUT go test -v ./tests/integration -run BenchmarkProxyCommands -bench=.
else
    timeout $TEST_TIMEOUT go test ./tests/integration -run BenchmarkProxyCommands -bench=. -benchtime=5s
fi

PERFORMANCE_RESULT=$?

if [ $PERFORMANCE_RESULT -eq 0 ]; then
    print_status "SUCCESS" "Performance tests passed"
else
    print_status "WARNING" "Performance tests had issues (non-critical)"
fi

# Step 7: Generate test report
print_status "INFO" "Generating test report..."

TOTAL_TESTS=3
PASSED_TESTS=0

[ $INTEGRATION_RESULT -eq 0 ] && ((PASSED_TESTS++))
[ $WORKFLOW_RESULT -eq 0 ] && ((PASSED_TESTS++))
[ $PERFORMANCE_RESULT -eq 0 ] && ((PASSED_TESTS++))

SUCCESS_RATE=$((PASSED_TESTS * 100 / TOTAL_TESTS))

echo ""
echo "📊 Integration Test Report"
echo "========================="
echo "Integration Tests: $([ $INTEGRATION_RESULT -eq 0 ] && echo "✅ PASSED" || echo "❌ FAILED")"
echo "Workflow Tests:    $([ $WORKFLOW_RESULT -eq 0 ] && echo "✅ PASSED" || echo "❌ FAILED")"
echo "Performance Tests: $([ $PERFORMANCE_RESULT -eq 0 ] && echo "✅ PASSED" || echo "⚠️  ISSUES")"
echo ""
echo "Overall Success Rate: $PASSED_TESTS/$TOTAL_TESTS ($SUCCESS_RATE%)"

# Step 8: Cleanup
print_status "INFO" "Cleaning up..."

# Kill any test processes
pkill -f "ollama-distributed.*start" || true

print_status "SUCCESS" "Cleanup complete"

# Final result
echo ""
if [ $INTEGRATION_RESULT -eq 0 ] && [ $WORKFLOW_RESULT -eq 0 ]; then
    print_status "SUCCESS" "All critical tests passed! 🎉"
    echo ""
    echo "🚀 OllamaMax is ready for production use!"
    echo ""
    echo "Next steps:"
    echo "1. Deploy: Copy binary to target systems"
    echo "2. Start: ./ollama-distributed start"
    echo "3. Monitor: ./ollama-distributed proxy status"
    echo "4. Scale: Add more nodes to the cluster"
    exit 0
else
    print_status "ERROR" "Some tests failed"
    echo ""
    echo "🔧 Troubleshooting:"
    echo "1. Check build environment: ./scripts/verify-go-env.sh"
    echo "2. Review test logs above"
    echo "3. Try Docker build: docker-compose -f docker-compose.build.yml up"
    echo "4. Check documentation: BUILD_INSTRUCTIONS.md"
    exit 1
fi

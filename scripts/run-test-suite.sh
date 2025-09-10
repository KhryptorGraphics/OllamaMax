#!/bin/bash

# Comprehensive Test Suite Runner
# Executes all validation tests for the ollamamax project

set -e  # Exit on any error

echo "🚀 Starting Comprehensive Test Suite for OllamaMax"
echo "================================================="

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results tracking
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
WARNINGS=0

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
    ((PASSED_TESTS++))
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
    ((FAILED_TESTS++))
}

log_warning() {
    echo -e "${YELLOW}⚠️ $1${NC}"
    ((WARNINGS++))
}

log_info() {
    echo -e "${BLUE}ℹ️ $1${NC}"
}

run_test() {
    local test_name="$1"
    local test_command="$2"
    
    echo -e "\n${BLUE}🔍 Running: $test_name${NC}"
    ((TOTAL_TESTS++))
    
    if eval "$test_command" > /tmp/test_output.log 2>&1; then
        log_success "$test_name passed"
        return 0
    else
        log_error "$test_name failed"
        echo "Error output:"
        cat /tmp/test_output.log | head -20
        return 1
    fi
}

# Check prerequisites
echo -e "\n${BLUE}🔧 Checking Prerequisites${NC}"

# Check Node.js
if command -v node &> /dev/null; then
    NODE_VERSION=$(node --version)
    log_success "Node.js found: $NODE_VERSION"
else
    log_error "Node.js not found"
    exit 1
fi

# Check npm
if command -v npm &> /dev/null; then
    NPM_VERSION=$(npm --version)
    log_success "npm found: $NPM_VERSION"
else
    log_error "npm not found"
    exit 1
fi

# Check Go
if command -v go &> /dev/null; then
    GO_VERSION=$(go version)
    log_success "Go found: $GO_VERSION"
else
    log_warning "Go not found - Go tests will be skipped"
fi

# Install dependencies if node_modules doesn't exist
if [ ! -d "node_modules" ]; then
    echo -e "\n${BLUE}📦 Installing Dependencies${NC}"
    if npm install; then
        log_success "Dependencies installed"
    else
        log_error "Failed to install dependencies"
        exit 1
    fi
fi

# Run test suites
echo -e "\n${BLUE}🧪 Running Test Suites${NC}"

# 1. Comprehensive Fix Validation
run_test "Comprehensive Fix Validation" "node scripts/test-all-fixes.cjs"

# 2. Jest Unit Tests
run_test "Jest Unit Tests" "npm run test:unit"

# 3. Agent System Tests
run_test "Agent System Integration Tests" "npm run test:agents"

# 4. Configuration Tests
run_test "System Configuration Tests" "npm run test:fixes"

# 5. Go tests (if Go is available)
if command -v go &> /dev/null; then
    run_test "Go Module Tests" "go test ./... -v -timeout 30s"
else
    log_warning "Go tests skipped - Go not available"
fi

# 6. Docker validation (if Docker is available)
if command -v docker &> /dev/null; then
    run_test "Docker Configuration Validation" "docker build -t ollamamax-test . --dry-run || docker build -t ollamamax-test ."
else
    log_warning "Docker tests skipped - Docker not available"
fi

# 7. Linting (if golangci-lint is available)
if command -v golangci-lint &> /dev/null; then
    run_test "Go Linting" "golangci-lint run"
else
    log_warning "Go linting skipped - golangci-lint not available"
fi

# 8. Security scan (basic)
run_test "Basic Security Scan" "
    echo 'Checking for sensitive files...'
    ! find . -name '*.key' -o -name '*.pem' -o -name '.env' | grep -v '.env.example' | head -1
"

# 9. Performance check
run_test "Basic Performance Check" "
    echo 'Running performance benchmark...'
    timeout 30s node -e '
        const start = Date.now();
        const fs = require(\"fs\");
        const testData = \"x\".repeat(1000000);
        fs.writeFileSync(\"/tmp/perf-test.txt\", testData);
        const data = fs.readFileSync(\"/tmp/perf-test.txt\", \"utf8\");
        fs.unlinkSync(\"/tmp/perf-test.txt\");
        const duration = Date.now() - start;
        console.log(\`Performance test completed in \${duration}ms\`);
        process.exit(duration < 100 ? 0 : 1);
    '
"

# 10. Port configuration validation
run_test "Port Configuration Validation" "
    echo 'Checking port configurations...'
    ! grep -r 'port.*:[[:space:]]*[0-9]\\{1,4\\}[^0-9]' . --include='*.yml' --include='*.json' --include='*.js' --include='*.go' | grep -v node_modules | grep -E 'port.*:[[:space:]]*([0-9]{1,4})[^0-9]' | grep -v '1[0-9][0-9][0-9][0-9]' | head -1
"

# Generate final report
echo -e "\n${BLUE}📊 Test Results Summary${NC}"
echo "========================="
echo "Total Tests: $TOTAL_TESTS"
echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed: ${RED}$FAILED_TESTS${NC}"
echo -e "Warnings: ${YELLOW}$WARNINGS${NC}"

# Calculate success rate
if [ $TOTAL_TESTS -gt 0 ]; then
    SUCCESS_RATE=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    echo "Success Rate: $SUCCESS_RATE%"
    
    if [ $SUCCESS_RATE -ge 90 ]; then
        echo -e "\n${GREEN}🎉 Excellent! System is in great condition.${NC}"
        exit_code=0
    elif [ $SUCCESS_RATE -ge 70 ]; then
        echo -e "\n${YELLOW}⚠️ Good, but some areas need attention.${NC}"
        exit_code=0
    else
        echo -e "\n${RED}🚨 System needs significant improvements.${NC}"
        exit_code=1
    fi
else
    echo -e "\n${RED}🚨 No tests were executed.${NC}"
    exit_code=1
fi

# Cleanup
rm -f /tmp/test_output.log

echo -e "\n${BLUE}📋 Recommendations:${NC}"
if [ $FAILED_TESTS -gt 0 ]; then
    echo "- Address all failed tests before deployment"
fi
if [ $WARNINGS -gt 3 ]; then
    echo "- Review and resolve warning conditions"
fi
if [ $SUCCESS_RATE -lt 90 ]; then
    echo "- Run individual test suites for detailed diagnostics"
    echo "- Check logs for specific error details"
fi

echo -e "\n${BLUE}🔧 Available Test Commands:${NC}"
echo "- npm run test:comprehensive  # Run comprehensive fix validation"
echo "- npm run test:agents        # Test agent system integration"
echo "- npm run test:fixes         # Test configuration fixes"
echo "- npm run test:all           # Run all npm test suites"

exit $exit_code
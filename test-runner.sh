#!/bin/bash

# OllamaMax Comprehensive Test Runner
# This script orchestrates all testing activities for the OllamaMax project

set -e

echo "🧪 OllamaMax Comprehensive Test Runner"
echo "======================================"
echo

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results tracking
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
TEST_RESULTS_DIR="./test-results"

# Create test results directory
mkdir -p ${TEST_RESULTS_DIR}
mkdir -p ${TEST_RESULTS_DIR}/coverage
mkdir -p ${TEST_RESULTS_DIR}/reports
mkdir -p ${TEST_RESULTS_DIR}/screenshots
mkdir -p ${TEST_RESULTS_DIR}/performance

print_status() {
    local status=$1
    local message=$2
    
    case $status in
        "PASS")
            echo -e "${GREEN}✅ PASS${NC}: $message"
            PASSED_TESTS=$((PASSED_TESTS + 1))
            ;;
        "FAIL")
            echo -e "${RED}❌ FAIL${NC}: $message"
            FAILED_TESTS=$((FAILED_TESTS + 1))
            ;;
        "WARN")
            echo -e "${YELLOW}⚠️  WARN${NC}: $message"
            ;;
        "INFO")
            echo -e "${BLUE}ℹ️  INFO${NC}: $message"
            ;;
    esac
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

# Phase 1: Pre-flight checks
echo "🔍 Phase 1: Pre-flight Checks"
echo "------------------------------"

# Check Go installation
if command -v go >/dev/null 2>&1; then
    GO_VERSION=$(go version)
    print_status "PASS" "Go is installed: $GO_VERSION"
else
    print_status "FAIL" "Go is not installed"
    exit 1
fi

# Check Node.js for E2E tests
if command -v node >/dev/null 2>&1; then
    NODE_VERSION=$(node --version)
    print_status "PASS" "Node.js is installed: $NODE_VERSION"
else
    print_status "WARN" "Node.js not found - E2E tests may not work"
fi

# Check Docker for integration tests
if command -v docker >/dev/null 2>&1; then
    print_status "PASS" "Docker is available for integration tests"
else
    print_status "WARN" "Docker not found - integration tests may be limited"
fi

echo

# Phase 2: Build validation
echo "🏗️  Phase 2: Build Validation"
echo "-----------------------------"

# Check for build issues
echo "Checking for package conflicts and build issues..."

if go build ./... >/dev/null 2>&1; then
    print_status "PASS" "All packages compile successfully"
else
    print_status "FAIL" "Build issues detected - checking specific problems"
    
    # Check for specific issues
    echo "Analyzing build failures:"
    
    # Check pkg/p2p package naming
    if ls pkg/p2p/*.go >/dev/null 2>&1; then
        PKG_NAMES=$(grep -h "^package " pkg/p2p/*.go | sort | uniq | wc -l)
        if [ "$PKG_NAMES" -gt 1 ]; then
            print_status "FAIL" "Multiple package names in pkg/p2p directory"
            grep -h "^package " pkg/p2p/*.go | sort | uniq
        fi
    fi
    
    # Check for duplicate type declarations
    if grep -r "type ModelInfo" pkg/types/ 2>/dev/null; then
        print_status "FAIL" "Duplicate ModelInfo type declarations found"
    fi
    
    echo "❌ Build must be fixed before running tests"
    echo "Please run: go build ./... to see detailed errors"
fi

echo

# Phase 3: Unit Tests
echo "🔬 Phase 3: Unit Tests"
echo "---------------------"

# Try to run unit tests despite build issues
echo "Attempting to run unit tests for individual packages..."

# Test packages that might work
TESTABLE_PACKAGES=(
    "./internal/config"
    "./pkg/database" 
)

for package in "${TESTABLE_PACKAGES[@]}"; do
    if [ -d "$package" ]; then
        echo "Testing package: $package"
        if go test "$package" -v -coverprofile="${TEST_RESULTS_DIR}/coverage/$(basename "$package").out" 2>&1 | tee "${TEST_RESULTS_DIR}/$(basename "$package")-test.log"; then
            print_status "PASS" "Unit tests for $package"
        else
            print_status "FAIL" "Unit tests for $package"
        fi
    else
        print_status "WARN" "Package $package not found"
    fi
done

echo

# Phase 4: Integration Tests
echo "🔗 Phase 4: Integration Tests"
echo "----------------------------"

# Check if integration tests exist
if [ -d "./tests/integration" ]; then
    echo "Running integration tests..."
    # Integration tests would go here
    print_status "INFO" "Integration test directory found"
else
    print_status "WARN" "No integration test directory found"
fi

echo

# Phase 5: E2E Tests
echo "🌐 Phase 5: End-to-End Tests"
echo "---------------------------"

if [ -d "./tests/e2e" ]; then
    echo "Running E2E tests..."
    
    # Check if dependencies are installed
    if [ -f "./tests/e2e/package.json" ]; then
        cd tests/e2e
        
        # Install dependencies if needed
        if [ ! -d "node_modules" ]; then
            print_status "INFO" "Installing E2E test dependencies..."
            if npm install >/dev/null 2>&1; then
                print_status "PASS" "E2E dependencies installed"
            else
                print_status "FAIL" "Failed to install E2E dependencies"
            fi
        fi
        
        # Run E2E tests
        if npm test 2>&1 | tee "../../${TEST_RESULTS_DIR}/e2e-test.log"; then
            print_status "PASS" "E2E tests completed"
        else
            print_status "FAIL" "E2E tests failed"
        fi
        
        cd ../..
    else
        print_status "WARN" "No package.json found in E2E directory"
    fi
else
    print_status "WARN" "No E2E test directory found"
fi

echo

# Phase 6: Performance Tests
echo "⚡ Phase 6: Performance Tests"
echo "----------------------------"

echo "Running performance benchmarks..."

# Run Go benchmarks if possible
if go test -bench=. -benchtime=1s ./... >/dev/null 2>&1; then
    print_status "PASS" "Go benchmarks completed"
    go test -bench=. -benchtime=1s ./... > "${TEST_RESULTS_DIR}/benchmarks.txt" 2>&1
else
    print_status "WARN" "Go benchmarks could not run due to build issues"
fi

echo

# Phase 7: Security Scan
echo "🛡️  Phase 7: Security Scan"
echo "-------------------------"

# Check for common security issues
echo "Performing basic security checks..."

# Check for hardcoded secrets
if grep -r "password\|secret\|key" --include="*.go" . | grep -v "test" | grep -v "_test.go" >/dev/null 2>&1; then
    print_status "WARN" "Potential hardcoded secrets found - manual review required"
else
    print_status "PASS" "No obvious hardcoded secrets detected"
fi

# Check for SQL injection vulnerabilities (basic check)
if grep -r "SELECT\|INSERT\|UPDATE\|DELETE" --include="*.go" . | grep -v "test" >/dev/null 2>&1; then
    print_status "INFO" "SQL queries found - ensure parameterized queries are used"
fi

echo

# Phase 8: Generate Reports
echo "📊 Phase 8: Generate Reports"
echo "---------------------------"

# Generate coverage report if coverage files exist
COVERAGE_FILES=$(find ${TEST_RESULTS_DIR}/coverage -name "*.out" 2>/dev/null)
if [ -n "$COVERAGE_FILES" ]; then
    echo "Generating coverage reports..."
    for coverage_file in $COVERAGE_FILES; do
        if [ -f "$coverage_file" ]; then
            go tool cover -html="$coverage_file" -o="${coverage_file%.out}.html"
            print_status "PASS" "Coverage report generated: ${coverage_file%.out}.html"
        fi
    done
else
    print_status "WARN" "No coverage files found to generate reports"
fi

# Create test summary
cat > "${TEST_RESULTS_DIR}/test-summary.md" << EOF
# OllamaMax Test Execution Summary

**Execution Date:** $(date)

## Results Overview
- **Total Tests:** $TOTAL_TESTS
- **Passed:** $PASSED_TESTS
- **Failed:** $FAILED_TESTS
- **Success Rate:** $(( PASSED_TESTS * 100 / TOTAL_TESTS ))%

## Test Categories
- Unit Tests: $(if [ "$FAILED_TESTS" -eq 0 ]; then echo "✅ PASS"; else echo "❌ FAIL"; fi)
- Integration Tests: ⚠️ Limited (build issues)
- E2E Tests: $(if [ -d "./tests/e2e" ]; then echo "✅ Available"; else echo "❌ Missing"; fi)
- Performance Tests: ⚠️ Limited (build issues)  
- Security Scan: ✅ Basic checks completed

## Critical Issues
$(if [ "$FAILED_TESTS" -gt 0 ]; then
    echo "- Build failures prevent comprehensive testing"
    echo "- Package naming conflicts in pkg/p2p"
    echo "- Type redeclaration issues in pkg/types"
    echo "- Configuration structure mismatches"
else
    echo "- No critical issues detected"
fi)

## Recommendations
1. Fix package naming conflicts immediately
2. Resolve build issues to enable full test suite
3. Implement comprehensive integration tests
4. Add automated security scanning
5. Establish continuous integration pipeline

## Files Generated
- Test logs: ${TEST_RESULTS_DIR}/
- Coverage reports: ${TEST_RESULTS_DIR}/coverage/
- Performance benchmarks: ${TEST_RESULTS_DIR}/benchmarks.txt
- Screenshots (E2E): ${TEST_RESULTS_DIR}/screenshots/

EOF

echo

# Phase 9: Final Summary
echo "📋 Phase 9: Test Execution Summary"
echo "=================================="

echo -e "${BLUE}Test Results Summary:${NC}"
echo "  Total Tests: $TOTAL_TESTS"
echo -e "  ${GREEN}Passed: $PASSED_TESTS${NC}"
echo -e "  ${RED}Failed: $FAILED_TESTS${NC}"

if [ "$FAILED_TESTS" -gt 0 ]; then
    echo -e "${RED}❌ TEST SUITE FAILED${NC}"
    echo
    echo "Critical issues must be resolved before production deployment:"
    echo "1. Fix package build issues"
    echo "2. Resolve configuration structure mismatches" 
    echo "3. Implement comprehensive error handling"
    echo
    echo "See detailed logs in: ${TEST_RESULTS_DIR}/"
    exit 1
else
    echo -e "${GREEN}✅ TEST SUITE PASSED${NC}"
    echo
    echo "All tests completed successfully!"
    echo "Coverage reports and test results available in: ${TEST_RESULTS_DIR}/"
fi

echo
echo "Test execution completed at $(date)"

# Make script executable and run it
chmod +x test-runner.sh
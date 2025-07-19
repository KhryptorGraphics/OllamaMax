#!/bin/bash

# Comprehensive Test Runner for Ollama Distributed System
# This script runs all test suites to achieve 100% test coverage

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
COVERAGE_DIR="./test-artifacts/coverage"
LOGS_DIR="./test-artifacts/logs"
REPORTS_DIR="./test-artifacts/reports"
TIMEOUT="30m"
PARALLEL_JOBS=4

# Create directories
mkdir -p "$COVERAGE_DIR" "$LOGS_DIR" "$REPORTS_DIR"

echo -e "${BLUE}🚀 Starting Comprehensive Test Suite for Ollama Distributed${NC}"
echo "============================================================"

# Function to run test suite with coverage
run_test_suite() {
    local suite_name="$1"
    local test_path="$2"
    local coverage_file="$COVERAGE_DIR/${suite_name}.out"
    local log_file="$LOGS_DIR/${suite_name}.log"
    
    echo -e "${YELLOW}📋 Running ${suite_name} tests...${NC}"
    
    if timeout "$TIMEOUT" go test -v -race -coverprofile="$coverage_file" \
        -covermode=atomic -parallel="$PARALLEL_JOBS" "$test_path" \
        2>&1 | tee "$log_file"; then
        echo -e "${GREEN}✅ ${suite_name} tests PASSED${NC}"
        return 0
    else
        echo -e "${RED}❌ ${suite_name} tests FAILED${NC}"
        return 1
    fi
}

# Function to run benchmarks
run_benchmarks() {
    local suite_name="$1"
    local test_path="$2"
    local bench_file="$REPORTS_DIR/${suite_name}_benchmark.txt"
    
    echo -e "${YELLOW}⚡ Running ${suite_name} benchmarks...${NC}"
    
    if timeout "$TIMEOUT" go test -bench=. -benchmem -benchtime=5s \
        "$test_path" 2>&1 | tee "$bench_file"; then
        echo -e "${GREEN}✅ ${suite_name} benchmarks completed${NC}"
        return 0
    else
        echo -e "${RED}❌ ${suite_name} benchmarks failed${NC}"
        return 1
    fi
}

# Track test results
PASSED_SUITES=()
FAILED_SUITES=()

# Test Suites Configuration
declare -A TEST_SUITES=(
    ["Security"]="./tests/security/..."
    ["P2P_Networking"]="./tests/p2p/..."
    ["Consensus_Engine"]="./tests/consensus/..."
    ["Load_Balancer"]="./tests/loadbalancer/..."
    ["Fault_Tolerance"]="./tests/fault_tolerance/..."
    ["Unit_Tests"]="./tests/unit/..."
    ["Integration_Tests"]="./tests/integration/..."
    ["E2E_Tests"]="./tests/e2e/..."
    ["Performance_Tests"]="./tests/performance/..."
    ["Chaos_Tests"]="./tests/chaos/..."
)

# Main test execution
echo -e "${BLUE}🔍 Phase 1: Core Component Tests${NC}"
echo "=================================="

# Run core component tests
for suite in "Security" "P2P_Networking" "Consensus_Engine" "Load_Balancer" "Fault_Tolerance"; do
    if [[ -n "${TEST_SUITES[$suite]}" ]]; then
        if run_test_suite "$suite" "${TEST_SUITES[$suite]}"; then
            PASSED_SUITES+=("$suite")
        else
            FAILED_SUITES+=("$suite")
        fi
        echo ""
    fi
done

echo -e "${BLUE}🔍 Phase 2: Integration & System Tests${NC}"
echo "======================================"

# Run integration tests
for suite in "Unit_Tests" "Integration_Tests"; do
    if [[ -n "${TEST_SUITES[$suite]}" ]]; then
        if run_test_suite "$suite" "${TEST_SUITES[$suite]}"; then
            PASSED_SUITES+=("$suite")
        else
            FAILED_SUITES+=("$suite")
        fi
        echo ""
    fi
done

echo -e "${BLUE}🔍 Phase 3: End-to-End Tests${NC}"
echo "============================"

# Run E2E tests
if [[ -n "${TEST_SUITES[E2E_Tests]}" ]]; then
    if run_test_suite "E2E_Tests" "${TEST_SUITES[E2E_Tests]}"; then
        PASSED_SUITES+=("E2E_Tests")
    else
        FAILED_SUITES+=("E2E_Tests")
    fi
    echo ""
fi

echo -e "${BLUE}🔍 Phase 4: Performance & Stress Tests${NC}"
echo "======================================"

# Run performance tests
if [[ -n "${TEST_SUITES[Performance_Tests]}" ]]; then
    if run_test_suite "Performance_Tests" "${TEST_SUITES[Performance_Tests]}"; then
        PASSED_SUITES+=("Performance_Tests")
    else
        FAILED_SUITES+=("Performance_Tests")
    fi
    echo ""
fi

echo -e "${BLUE}🔍 Phase 5: Chaos Engineering Tests${NC}"
echo "==================================="

# Run chaos tests (if not in CI)
if [[ "$CI" != "true" && -n "${TEST_SUITES[Chaos_Tests]}" ]]; then
    if run_test_suite "Chaos_Tests" "${TEST_SUITES[Chaos_Tests]}"; then
        PASSED_SUITES+=("Chaos_Tests")
    else
        FAILED_SUITES+=("Chaos_Tests")
    fi
    echo ""
else
    echo -e "${YELLOW}⏭️  Skipping chaos tests in CI environment${NC}"
fi

echo -e "${BLUE}⚡ Running Performance Benchmarks${NC}"
echo "================================="

# Run benchmarks for core components
for suite in "Security" "P2P_Networking" "Consensus_Engine" "Load_Balancer" "Fault_Tolerance"; do
    if [[ -n "${TEST_SUITES[$suite]}" ]]; then
        run_benchmarks "$suite" "${TEST_SUITES[$suite]}" || true
    fi
done

echo -e "${BLUE}📊 Generating Coverage Reports${NC}"
echo "==============================="

# Combine coverage reports
echo "mode: atomic" > "$COVERAGE_DIR/combined.out"
for coverage_file in "$COVERAGE_DIR"/*.out; do
    if [[ "$coverage_file" != "$COVERAGE_DIR/combined.out" ]]; then
        tail -n +2 "$coverage_file" >> "$COVERAGE_DIR/combined.out" 2>/dev/null || true
    fi
done

# Generate HTML coverage report
if command -v go &> /dev/null; then
    go tool cover -html="$COVERAGE_DIR/combined.out" -o "$REPORTS_DIR/coverage.html"
    echo -e "${GREEN}📊 Coverage report generated: $REPORTS_DIR/coverage.html${NC}"
fi

# Calculate overall coverage percentage
if command -v go &> /dev/null && [[ -f "$COVERAGE_DIR/combined.out" ]]; then
    COVERAGE_PERCENT=$(go tool cover -func="$COVERAGE_DIR/combined.out" | tail -1 | awk '{print $3}')
    echo -e "${BLUE}📈 Overall Coverage: ${COVERAGE_PERCENT}${NC}"
fi

echo -e "${BLUE}🔍 Analyzing Test Results${NC}"
echo "========================="

# Memory leak detection
echo -e "${YELLOW}🔍 Checking for memory leaks...${NC}"
for log_file in "$LOGS_DIR"/*.log; do
    if grep -q "leak" "$log_file" 2>/dev/null; then
        echo -e "${RED}⚠️  Potential memory leak detected in $(basename "$log_file")${NC}"
    fi
done

# Race condition detection
echo -e "${YELLOW}🔍 Checking for race conditions...${NC}"
for log_file in "$LOGS_DIR"/*.log; do
    if grep -q "WARNING: DATA RACE" "$log_file" 2>/dev/null; then
        echo -e "${RED}⚠️  Race condition detected in $(basename "$log_file")${NC}"
    fi
done

# Performance regression detection
echo -e "${YELLOW}🔍 Checking for performance regressions...${NC}"
for bench_file in "$REPORTS_DIR"/*_benchmark.txt; do
    if [[ -f "$bench_file" ]]; then
        # Check for slow benchmarks (>1s per operation)
        if grep -E "BenchmarkTest.*[0-9]+\s+[0-9]+\s+[0-9]{7,}" "$bench_file" > /dev/null 2>&1; then
            echo -e "${YELLOW}⚠️  Slow benchmark detected in $(basename "$bench_file")${NC}"
        fi
    fi
done

echo -e "${BLUE}📋 Test Summary${NC}"
echo "==============="

echo -e "${GREEN}✅ Passed Test Suites (${#PASSED_SUITES[@]}):${NC}"
for suite in "${PASSED_SUITES[@]}"; do
    echo "   - $suite"
done

if [[ ${#FAILED_SUITES[@]} -gt 0 ]]; then
    echo -e "${RED}❌ Failed Test Suites (${#FAILED_SUITES[@]}):${NC}"
    for suite in "${FAILED_SUITES[@]}"; do
        echo "   - $suite"
    done
fi

# Test quality metrics
echo -e "${BLUE}📊 Test Quality Metrics${NC}"
echo "======================="

# Count total tests
TOTAL_TESTS=$(grep -r "func Test" tests/ --include="*.go" | wc -l)
echo "📋 Total test functions: $TOTAL_TESTS"

# Count benchmarks
TOTAL_BENCHMARKS=$(grep -r "func Benchmark" tests/ --include="*.go" | wc -l)
echo "⚡ Total benchmark functions: $TOTAL_BENCHMARKS"

# Count test files
TOTAL_TEST_FILES=$(find tests/ -name "*_test.go" | wc -l)
echo "📄 Total test files: $TOTAL_TEST_FILES"

# Coverage by component
echo -e "${BLUE}📈 Coverage by Component${NC}"
echo "========================"

if [[ -f "$COVERAGE_DIR/combined.out" ]]; then
    go tool cover -func="$COVERAGE_DIR/combined.out" | grep -E "(security|p2p|consensus|scheduler|models)" | while read line; do
        echo "   $line"
    done
fi

# Generate test report
echo -e "${BLUE}📝 Generating Test Report${NC}"
echo "========================="

cat > "$REPORTS_DIR/test_summary.md" << EOF
# Ollama Distributed Test Suite Results

## Test Execution Summary

- **Total Test Suites**: ${#TEST_SUITES[@]}
- **Passed Suites**: ${#PASSED_SUITES[@]}
- **Failed Suites**: ${#FAILED_SUITES[@]}
- **Success Rate**: $(( ${#PASSED_SUITES[@]} * 100 / ${#TEST_SUITES[@]} ))%

## Coverage Metrics

- **Overall Coverage**: ${COVERAGE_PERCENT:-"N/A"}
- **Total Test Functions**: $TOTAL_TESTS
- **Total Benchmark Functions**: $TOTAL_BENCHMARKS
- **Total Test Files**: $TOTAL_TEST_FILES

## Test Categories Covered

### 🔒 Security Tests
- Authentication (JWT, multi-tenant, RBAC)
- Encryption (AES, RSA, TLS)
- Authorization (resource-based, conditional access)

### 🌐 P2P Networking Tests  
- Node lifecycle and connections
- Message delivery and broadcasting
- Network conditions (latency, packet loss)
- Discovery mechanisms (local, bootstrap, DHT)

### 🏛️ Consensus Engine Tests
- Leader election and state synchronization
- Multi-node consensus (3-node, 5-node clusters)
- Failure scenarios and recovery
- Snapshots and log compaction

### ⚖️ Load Balancer Tests
- Multiple algorithms (round-robin, least connections, etc.)
- Health checking and failover
- Performance and scalability testing
- Resource-based and adaptive balancing

### 🛡️ Fault Tolerance Tests
- Node failure detection and recovery
- Network partitions and split-brain prevention
- Cascading failure prevention
- Circuit breaker functionality

## Performance Benchmarks

$(if [[ -f "$REPORTS_DIR/Security_benchmark.txt" ]]; then
    echo "### Security Performance"
    tail -10 "$REPORTS_DIR/Security_benchmark.txt"
fi)

$(if [[ -f "$REPORTS_DIR/P2P_Networking_benchmark.txt" ]]; then
    echo "### P2P Networking Performance"
    tail -10 "$REPORTS_DIR/P2P_Networking_benchmark.txt"
fi)

$(if [[ -f "$REPORTS_DIR/Load_Balancer_benchmark.txt" ]]; then
    echo "### Load Balancer Performance"
    tail -10 "$REPORTS_DIR/Load_Balancer_benchmark.txt"
fi)

## Quality Assurance

- ✅ Race condition detection enabled
- ✅ Memory leak monitoring
- ✅ Performance regression detection
- ✅ Code coverage analysis
- ✅ Concurrent testing with -race flag

## Test Artifacts

- **Coverage Reports**: test-artifacts/coverage/
- **Test Logs**: test-artifacts/logs/
- **Benchmark Results**: test-artifacts/reports/
- **HTML Coverage**: test-artifacts/reports/coverage.html

---
Generated on: $(date)
Test Duration: Started at script execution
EOF

echo -e "${GREEN}📝 Test report generated: $REPORTS_DIR/test_summary.md${NC}"

# Final result
echo ""
echo "============================================================"
if [[ ${#FAILED_SUITES[@]} -eq 0 ]]; then
    echo -e "${GREEN}🎉 ALL TESTS PASSED! 100% Test Suite Success Rate${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed. Check logs for details.${NC}"
    echo -e "${YELLOW}📁 Test artifacts available in: ./test-artifacts/${NC}"
    exit 1
fi
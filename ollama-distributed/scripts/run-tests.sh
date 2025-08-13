#!/bin/bash

# 🧪 Comprehensive Test Runner Script
# This script runs all tests with coverage, performance analysis, and reporting

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
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

log_step() {
    echo -e "${PURPLE}[STEP]${NC} $1"
}

# Configuration
TEST_TIMEOUT=${TEST_TIMEOUT:-30m}
COVERAGE_TARGET=${COVERAGE_TARGET:-80}
PARALLEL_TESTS=${PARALLEL_TESTS:-true}
GENERATE_REPORTS=${GENERATE_REPORTS:-true}
RUN_BENCHMARKS=${RUN_BENCHMARKS:-true}
RUN_INTEGRATION=${RUN_INTEGRATION:-true}
OUTPUT_DIR=${OUTPUT_DIR:-./test-results}

# Check if we're in the right directory
if [[ ! -f "go.mod" ]]; then
    log_error "This script must be run from the ollama-distributed root directory"
    exit 1
fi

# Create output directory
mkdir -p "$OUTPUT_DIR"

log_info "🧪 Starting comprehensive test suite..."
echo "Configuration:"
echo "  - Test timeout: $TEST_TIMEOUT"
echo "  - Coverage target: $COVERAGE_TARGET%"
echo "  - Parallel tests: $PARALLEL_TESTS"
echo "  - Generate reports: $GENERATE_REPORTS"
echo "  - Run benchmarks: $RUN_BENCHMARKS"
echo "  - Run integration tests: $RUN_INTEGRATION"
echo "  - Output directory: $OUTPUT_DIR"
echo

# Phase 1: Unit Tests
log_step "Phase 1: Running unit tests with coverage..."

UNIT_TEST_FLAGS="-v -race -timeout=$TEST_TIMEOUT"
if [[ "$PARALLEL_TESTS" == "true" ]]; then
    UNIT_TEST_FLAGS="$UNIT_TEST_FLAGS -parallel=4"
fi

# Run unit tests with coverage
log_info "Running unit tests..."
go test $UNIT_TEST_FLAGS -coverprofile="$OUTPUT_DIR/coverage.out" -covermode=atomic ./pkg/... 2>&1 | tee "$OUTPUT_DIR/unit-tests.log"

UNIT_EXIT_CODE=${PIPESTATUS[0]}

if [[ $UNIT_EXIT_CODE -eq 0 ]]; then
    log_success "Unit tests passed!"
else
    log_error "Unit tests failed with exit code $UNIT_EXIT_CODE"
fi

# Generate coverage report
if [[ -f "$OUTPUT_DIR/coverage.out" ]]; then
    log_info "Generating coverage report..."
    go tool cover -html="$OUTPUT_DIR/coverage.out" -o "$OUTPUT_DIR/coverage.html"
    
    # Calculate coverage percentage
    COVERAGE_PERCENT=$(go tool cover -func="$OUTPUT_DIR/coverage.out" | grep total | awk '{print $3}' | sed 's/%//')
    
    if [[ -n "$COVERAGE_PERCENT" ]]; then
        echo "Coverage: $COVERAGE_PERCENT%" > "$OUTPUT_DIR/coverage-summary.txt"
        
        if (( $(echo "$COVERAGE_PERCENT >= $COVERAGE_TARGET" | bc -l) )); then
            log_success "Coverage target met: $COVERAGE_PERCENT% >= $COVERAGE_TARGET%"
        else
            log_warning "Coverage target not met: $COVERAGE_PERCENT% < $COVERAGE_TARGET%"
        fi
    fi
fi

# Phase 2: Integration Tests
if [[ "$RUN_INTEGRATION" == "true" ]]; then
    log_step "Phase 2: Running integration tests..."
    
    # Set up test environment
    log_info "Setting up test environment..."
    
    # TODO: Start test services (Redis, databases, etc.)
    # docker-compose -f docker-compose.test.yml up -d
    
    # Run integration tests
    log_info "Running integration tests..."
    go test $UNIT_TEST_FLAGS -tags=integration ./tests/integration/... 2>&1 | tee "$OUTPUT_DIR/integration-tests.log"
    
    INTEGRATION_EXIT_CODE=${PIPESTATUS[0]}
    
    if [[ $INTEGRATION_EXIT_CODE -eq 0 ]]; then
        log_success "Integration tests passed!"
    else
        log_error "Integration tests failed with exit code $INTEGRATION_EXIT_CODE"
    fi
    
    # Clean up test environment
    log_info "Cleaning up test environment..."
    # docker-compose -f docker-compose.test.yml down
else
    log_info "Skipping integration tests (RUN_INTEGRATION=false)"
    INTEGRATION_EXIT_CODE=0
fi

# Phase 3: Benchmarks and Performance Tests
if [[ "$RUN_BENCHMARKS" == "true" ]]; then
    log_step "Phase 3: Running benchmarks and performance tests..."
    
    log_info "Running benchmarks..."
    go test -bench=. -benchmem -timeout=$TEST_TIMEOUT ./pkg/... 2>&1 | tee "$OUTPUT_DIR/benchmarks.log"
    
    BENCHMARK_EXIT_CODE=${PIPESTATUS[0]}
    
    if [[ $BENCHMARK_EXIT_CODE -eq 0 ]]; then
        log_success "Benchmarks completed!"
    else
        log_warning "Benchmarks completed with warnings (exit code $BENCHMARK_EXIT_CODE)"
    fi
    
    # Run memory profiling
    log_info "Running memory profiling..."
    go test -memprofile="$OUTPUT_DIR/mem.prof" -bench=BenchmarkMemory ./pkg/memory/... 2>/dev/null || true
    
    # Run CPU profiling
    log_info "Running CPU profiling..."
    go test -cpuprofile="$OUTPUT_DIR/cpu.prof" -bench=BenchmarkCPU ./pkg/... 2>/dev/null || true
else
    log_info "Skipping benchmarks (RUN_BENCHMARKS=false)"
    BENCHMARK_EXIT_CODE=0
fi

# Phase 4: Security Tests
log_step "Phase 4: Running security tests..."

log_info "Running security vulnerability scan..."
if command -v gosec >/dev/null 2>&1; then
    gosec -fmt json -out "$OUTPUT_DIR/security-report.json" ./... 2>/dev/null || true
    gosec ./... 2>&1 | tee "$OUTPUT_DIR/security-scan.log" || true
    log_success "Security scan completed"
else
    log_warning "gosec not installed, skipping security scan"
fi

# Check for common security issues
log_info "Checking for hardcoded secrets..."
if command -v git >/dev/null 2>&1; then
    git ls-files | xargs grep -l "password\|secret\|key\|token" | grep -v ".git" | grep -v "test" > "$OUTPUT_DIR/potential-secrets.txt" 2>/dev/null || true
    if [[ -s "$OUTPUT_DIR/potential-secrets.txt" ]]; then
        log_warning "Found files that may contain hardcoded secrets (see potential-secrets.txt)"
    else
        log_success "No obvious hardcoded secrets found"
    fi
fi

# Phase 5: Code Quality Analysis
log_step "Phase 5: Running code quality analysis..."

# Run go vet
log_info "Running go vet..."
go vet ./... 2>&1 | tee "$OUTPUT_DIR/vet-report.log"
VET_EXIT_CODE=${PIPESTATUS[0]}

if [[ $VET_EXIT_CODE -eq 0 ]]; then
    log_success "go vet passed!"
else
    log_warning "go vet found issues (exit code $VET_EXIT_CODE)"
fi

# Run golint if available
if command -v golint >/dev/null 2>&1; then
    log_info "Running golint..."
    golint ./... 2>&1 | tee "$OUTPUT_DIR/lint-report.log" || true
    log_success "golint completed"
else
    log_info "golint not available, skipping"
fi

# Run gofmt check
log_info "Checking code formatting..."
UNFORMATTED=$(gofmt -l . | grep -v vendor/ | grep -v .git/ || true)
if [[ -n "$UNFORMATTED" ]]; then
    echo "$UNFORMATTED" > "$OUTPUT_DIR/unformatted-files.txt"
    log_warning "Found unformatted files (see unformatted-files.txt)"
else
    log_success "All files are properly formatted"
fi

# Phase 6: Generate Reports
if [[ "$GENERATE_REPORTS" == "true" ]]; then
    log_step "Phase 6: Generating comprehensive reports..."
    
    # Create test summary
    cat > "$OUTPUT_DIR/test-summary.md" << EOF
# Test Summary Report

**Generated:** $(date)
**Duration:** $(date -d@$SECONDS -u +%H:%M:%S)

## Test Results

### Unit Tests
- Status: $([ $UNIT_EXIT_CODE -eq 0 ] && echo "✅ PASSED" || echo "❌ FAILED")
- Exit Code: $UNIT_EXIT_CODE

### Integration Tests
- Status: $([ $INTEGRATION_EXIT_CODE -eq 0 ] && echo "✅ PASSED" || echo "❌ FAILED")
- Exit Code: $INTEGRATION_EXIT_CODE

### Benchmarks
- Status: $([ $BENCHMARK_EXIT_CODE -eq 0 ] && echo "✅ COMPLETED" || echo "⚠️ WARNINGS")
- Exit Code: $BENCHMARK_EXIT_CODE

### Code Quality
- go vet: $([ $VET_EXIT_CODE -eq 0 ] && echo "✅ PASSED" || echo "⚠️ ISSUES")

## Coverage
EOF

    if [[ -f "$OUTPUT_DIR/coverage-summary.txt" ]]; then
        cat "$OUTPUT_DIR/coverage-summary.txt" >> "$OUTPUT_DIR/test-summary.md"
    else
        echo "Coverage: Not available" >> "$OUTPUT_DIR/test-summary.md"
    fi

    cat >> "$OUTPUT_DIR/test-summary.md" << EOF

## Files Generated
- Unit test log: unit-tests.log
- Integration test log: integration-tests.log
- Benchmark results: benchmarks.log
- Coverage report: coverage.html
- Security scan: security-scan.log
- Code quality: vet-report.log, lint-report.log

## Next Steps
1. Review failed tests and fix issues
2. Improve code coverage if below target
3. Address security vulnerabilities
4. Fix code quality issues
EOF

    log_success "Test summary generated: $OUTPUT_DIR/test-summary.md"
fi

# Final summary
echo
log_info "🏁 TEST SUITE COMPLETED"
echo "Results:"
echo "  - Unit tests: $([ $UNIT_EXIT_CODE -eq 0 ] && echo -e "${GREEN}PASSED${NC}" || echo -e "${RED}FAILED${NC}")"
echo "  - Integration tests: $([ $INTEGRATION_EXIT_CODE -eq 0 ] && echo -e "${GREEN}PASSED${NC}" || echo -e "${RED}FAILED${NC}")"
echo "  - Benchmarks: $([ $BENCHMARK_EXIT_CODE -eq 0 ] && echo -e "${GREEN}COMPLETED${NC}" || echo -e "${YELLOW}WARNINGS${NC}")"
echo "  - Code quality: $([ $VET_EXIT_CODE -eq 0 ] && echo -e "${GREEN}PASSED${NC}" || echo -e "${YELLOW}ISSUES${NC}")"

if [[ -f "$OUTPUT_DIR/coverage-summary.txt" ]]; then
    COVERAGE_PERCENT=$(cat "$OUTPUT_DIR/coverage-summary.txt" | grep -o '[0-9.]*')
    echo "  - Coverage: $COVERAGE_PERCENT%"
fi

echo "  - Reports: $OUTPUT_DIR/"
echo

# Exit with appropriate code
OVERALL_EXIT_CODE=0
if [[ $UNIT_EXIT_CODE -ne 0 ]]; then
    OVERALL_EXIT_CODE=1
fi
if [[ $INTEGRATION_EXIT_CODE -ne 0 ]]; then
    OVERALL_EXIT_CODE=1
fi

if [[ $OVERALL_EXIT_CODE -eq 0 ]]; then
    log_success "🎉 All critical tests passed!"
else
    log_error "💥 Some critical tests failed!"
fi

exit $OVERALL_EXIT_CODE

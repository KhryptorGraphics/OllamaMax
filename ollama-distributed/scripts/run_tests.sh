#!/bin/bash

# Ollama Distributed Test Automation Script
# This script runs comprehensive test suites and generates reports

set -euo pipefail

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEST_OUTPUT_DIR="${PROJECT_ROOT}/test_results"
COVERAGE_DIR="${TEST_OUTPUT_DIR}/coverage"
REPORTS_DIR="${TEST_OUTPUT_DIR}/reports"
LOG_DIR="${TEST_OUTPUT_DIR}/logs"

# Test configuration
RUN_UNIT_TESTS=true
RUN_INTEGRATION_TESTS=true
RUN_SECURITY_TESTS=true
RUN_P2P_TESTS=true
RUN_FAULT_TOLERANCE_TESTS=true
RUN_CHAOS_TESTS=false  # Default to false as they're resource intensive
RUN_PERFORMANCE_TESTS=false  # Default to false for CI environments
GENERATE_COVERAGE=true
GENERATE_REPORTS=true
PARALLEL_EXECUTION=true
TIMEOUT="30m"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "${LOG_DIR}/test_execution.log"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "${LOG_DIR}/test_execution.log"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "${LOG_DIR}/test_execution.log"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "${LOG_DIR}/test_execution.log"
}

# Setup functions
setup_test_environment() {
    log_info "Setting up test environment..."
    
    # Create output directories
    mkdir -p "${TEST_OUTPUT_DIR}" "${COVERAGE_DIR}" "${REPORTS_DIR}" "${LOG_DIR}"
    
    # Change to project root
    cd "${PROJECT_ROOT}"
    
    # Ensure Go modules are ready
    log_info "Downloading Go dependencies..."
    go mod download
    go mod tidy
    
    # Install test dependencies
    if ! command -v gotestsum &> /dev/null; then
        log_info "Installing gotestsum for enhanced test output..."
        go install gotest.tools/gotestsum@latest
    fi
    
    if ! command -v gocov &> /dev/null; then
        log_info "Installing gocov for coverage reporting..."
        go install github.com/axw/gocov/gocov@latest
    fi
    
    if ! command -v gocov-html &> /dev/null; then
        log_info "Installing gocov-html for HTML coverage reports..."
        go install github.com/matm/gocov-html@latest
    fi
    
    # Check if claude-flow is available for swarm testing
    if command -v npx &> /dev/null && npx claude-flow@alpha --version &> /dev/null; then
        log_info "Claude Flow MCP detected - enhanced coordination testing enabled"
        export CLAUDE_FLOW_ENABLED=true
    else
        log_warning "Claude Flow MCP not detected - running tests without swarm coordination"
        export CLAUDE_FLOW_ENABLED=false
    fi
    
    log_success "Test environment setup complete"
}

# Test execution functions
run_unit_tests() {
    if [ "$RUN_UNIT_TESTS" = false ]; then
        return 0
    fi
    
    log_info "Running unit tests..."
    
    local output_file="${REPORTS_DIR}/unit_tests.json"
    local coverage_file="${COVERAGE_DIR}/unit_coverage.out"
    
    if [ "$GENERATE_COVERAGE" = true ]; then
        local coverage_args="-coverprofile=${coverage_file} -covermode=atomic"
    else
        local coverage_args=""
    fi
    
    local parallel_args=""
    if [ "$PARALLEL_EXECUTION" = true ]; then
        parallel_args="-p=$(nproc)"
    fi
    
    if gotestsum \
        --format=pkgname \
        --jsonfile="${output_file}" \
        ${parallel_args} \
        -- \
        -timeout="${TIMEOUT}" \
        ${coverage_args} \
        ./tests/unit/...; then
        log_success "Unit tests passed"
        return 0
    else
        log_error "Unit tests failed"
        return 1
    fi
}

run_integration_tests() {
    if [ "$RUN_INTEGRATION_TESTS" = false ]; then
        return 0
    fi
    
    log_info "Running integration tests..."
    
    local output_file="${REPORTS_DIR}/integration_tests.json"
    local coverage_file="${COVERAGE_DIR}/integration_coverage.out"
    
    if [ "$GENERATE_COVERAGE" = true ]; then
        local coverage_args="-coverprofile=${coverage_file} -covermode=atomic"
    else
        local coverage_args=""
    fi
    
    # Integration tests run sequentially to avoid resource conflicts
    if gotestsum \
        --format=standard-verbose \
        --jsonfile="${output_file}" \
        -- \
        -timeout="${TIMEOUT}" \
        ${coverage_args} \
        ./tests/integration/...; then
        log_success "Integration tests passed"
        return 0
    else
        log_error "Integration tests failed"
        return 1
    fi
}

run_security_tests() {
    if [ "$RUN_SECURITY_TESTS" = false ]; then
        return 0
    fi
    
    log_info "Running security tests..."
    
    local output_file="${REPORTS_DIR}/security_tests.json"
    local coverage_file="${COVERAGE_DIR}/security_coverage.out"
    
    if [ "$GENERATE_COVERAGE" = true ]; then
        local coverage_args="-coverprofile=${coverage_file} -covermode=atomic"
    else
        local coverage_args=""
    fi
    
    if gotestsum \
        --format=standard-verbose \
        --jsonfile="${output_file}" \
        -- \
        -timeout="${TIMEOUT}" \
        ${coverage_args} \
        ./tests/security/...; then
        log_success "Security tests passed"
        return 0
    else
        log_error "Security tests failed"
        return 1
    fi
}

run_p2p_tests() {
    if [ "$RUN_P2P_TESTS" = false ]; then
        return 0
    fi
    
    log_info "Running P2P networking tests..."
    
    local output_file="${REPORTS_DIR}/p2p_tests.json"
    local coverage_file="${COVERAGE_DIR}/p2p_coverage.out"
    
    if [ "$GENERATE_COVERAGE" = true ]; then
        local coverage_args="-coverprofile=${coverage_file} -covermode=atomic"
    else
        local coverage_args=""
    fi
    
    if gotestsum \
        --format=standard-verbose \
        --jsonfile="${output_file}" \
        -- \
        -timeout="${TIMEOUT}" \
        ${coverage_args} \
        ./tests/p2p/...; then
        log_success "P2P tests passed"
        return 0
    else
        log_error "P2P tests failed"
        return 1
    fi
}

run_fault_tolerance_tests() {
    if [ "$RUN_FAULT_TOLERANCE_TESTS" = false ]; then
        return 0
    fi
    
    log_info "Running fault tolerance tests..."
    
    local output_file="${REPORTS_DIR}/fault_tolerance_tests.json"
    local coverage_file="${COVERAGE_DIR}/fault_tolerance_coverage.out"
    
    if [ "$GENERATE_COVERAGE" = true ]; then
        local coverage_args="-coverprofile=${coverage_file} -covermode=atomic"
    else
        local coverage_args=""
    fi
    
    if gotestsum \
        --format=standard-verbose \
        --jsonfile="${output_file}" \
        -- \
        -timeout="${TIMEOUT}" \
        ${coverage_args} \
        ./tests/fault_tolerance/...; then
        log_success "Fault tolerance tests passed"
        return 0
    else
        log_error "Fault tolerance tests failed"
        return 1
    fi
}

run_chaos_tests() {
    if [ "$RUN_CHAOS_TESTS" = false ]; then
        return 0
    fi
    
    log_info "Running chaos engineering tests..."
    log_warning "Chaos tests are resource intensive and may take significant time"
    
    local output_file="${REPORTS_DIR}/chaos_tests.json"
    
    if gotestsum \
        --format=standard-verbose \
        --jsonfile="${output_file}" \
        -- \
        -timeout="60m" \
        ./tests/chaos/...; then
        log_success "Chaos tests passed"
        return 0
    else
        log_error "Chaos tests failed"
        return 1
    fi
}

run_performance_tests() {
    if [ "$RUN_PERFORMANCE_TESTS" = false ]; then
        return 0
    fi
    
    log_info "Running performance tests..."
    
    local output_file="${REPORTS_DIR}/performance_tests.json"
    
    if gotestsum \
        --format=standard-verbose \
        --jsonfile="${output_file}" \
        -- \
        -timeout="${TIMEOUT}" \
        -bench=. \
        -benchmem \
        ./tests/performance/...; then
        log_success "Performance tests passed"
        return 0
    else
        log_error "Performance tests failed"
        return 1
    fi
}

# Coverage and reporting functions
generate_coverage_report() {
    if [ "$GENERATE_COVERAGE" = false ]; then
        return 0
    fi
    
    log_info "Generating coverage reports..."
    
    # Combine all coverage files
    local combined_coverage="${COVERAGE_DIR}/combined_coverage.out"
    
    # Create combined coverage header
    echo "mode: atomic" > "${combined_coverage}"
    
    # Combine all coverage files (skip header lines)
    for coverage_file in "${COVERAGE_DIR}"/*_coverage.out; do
        if [ -f "${coverage_file}" ]; then
            tail -n +2 "${coverage_file}" >> "${combined_coverage}"
        fi
    done
    
    # Generate HTML coverage report
    local html_report="${REPORTS_DIR}/coverage.html"
    if command -v gocov &> /dev/null && command -v gocov-html &> /dev/null; then
        gocov convert "${combined_coverage}" | gocov-html > "${html_report}"
        log_success "HTML coverage report generated: ${html_report}"
    fi
    
    # Generate text coverage summary
    local text_report="${REPORTS_DIR}/coverage_summary.txt"
    go tool cover -func="${combined_coverage}" > "${text_report}"
    
    # Extract overall coverage percentage
    local coverage_percentage=$(go tool cover -func="${combined_coverage}" | grep "total:" | awk '{print $3}')
    log_info "Overall test coverage: ${coverage_percentage}"
    
    # Check if coverage meets minimum threshold
    local min_coverage="70.0%"
    local coverage_numeric=$(echo "${coverage_percentage}" | sed 's/%//')
    local min_numeric=$(echo "${min_coverage}" | sed 's/%//')
    
    if (( $(echo "${coverage_numeric} >= ${min_numeric}" | bc -l) )); then
        log_success "Coverage ${coverage_percentage} meets minimum threshold of ${min_coverage}"
    else
        log_warning "Coverage ${coverage_percentage} below minimum threshold of ${min_coverage}"
    fi
    
    log_success "Coverage reports generated in ${REPORTS_DIR}/"
}

generate_test_reports() {
    if [ "$GENERATE_REPORTS" = false ]; then
        return 0
    fi
    
    log_info "Generating test reports..."
    
    # Generate consolidated test report
    local consolidated_report="${REPORTS_DIR}/test_summary.json"
    local html_report="${REPORTS_DIR}/test_report.html"
    
    # Create JSON summary
    cat > "${consolidated_report}" << EOF
{
    "timestamp": "$(date -Iseconds)",
    "environment": {
        "go_version": "$(go version)",
        "os": "$(uname -s)",
        "arch": "$(uname -m)",
        "claude_flow_enabled": "${CLAUDE_FLOW_ENABLED:-false}"
    },
    "test_suites": {
EOF

    local first=true
    for test_file in "${REPORTS_DIR}"/*_tests.json; do
        if [ -f "${test_file}" ]; then
            local suite_name=$(basename "${test_file}" _tests.json)
            
            if [ "$first" = false ]; then
                echo "," >> "${consolidated_report}"
            fi
            first=false
            
            echo "        \"${suite_name}\": $(cat "${test_file}")" >> "${consolidated_report}"
        fi
    done
    
    cat >> "${consolidated_report}" << EOF
    }
}
EOF

    # Generate HTML report
    generate_html_report "${consolidated_report}" "${html_report}"
    
    log_success "Test reports generated in ${REPORTS_DIR}/"
}

generate_html_report() {
    local json_file="$1"
    local html_file="$2"
    
    cat > "${html_file}" << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Ollama Distributed Test Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f0f0f0; padding: 20px; border-radius: 5px; }
        .suite { margin: 20px 0; padding: 15px; border: 1px solid #ddd; border-radius: 5px; }
        .passed { background: #d4edda; }
        .failed { background: #f8d7da; }
        .metrics { display: flex; gap: 20px; }
        .metric { text-align: center; padding: 10px; }
        .coverage-bar { width: 100%; height: 20px; background: #e9ecef; border-radius: 10px; overflow: hidden; }
        .coverage-fill { height: 100%; background: #28a745; }
        table { width: 100%; border-collapse: collapse; margin: 10px 0; }
        th, td { padding: 8px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Ollama Distributed System Test Report</h1>
        <p>Generated: <span id="timestamp"></span></p>
        <p>Environment: <span id="environment"></span></p>
    </div>
    
    <div class="metrics">
        <div class="metric">
            <h3>Overall Coverage</h3>
            <div class="coverage-bar">
                <div class="coverage-fill" id="coverage-fill" style="width: 0%"></div>
            </div>
            <span id="coverage-text">0%</span>
        </div>
        <div class="metric">
            <h3>Test Suites</h3>
            <div><span id="suites-passed">0</span> passed</div>
            <div><span id="suites-failed">0</span> failed</div>
        </div>
        <div class="metric">
            <h3>Total Tests</h3>
            <div><span id="tests-passed">0</span> passed</div>
            <div><span id="tests-failed">0</span> failed</div>
        </div>
    </div>
    
    <div id="test-suites"></div>
    
    <script>
        // This would contain JavaScript to parse the JSON and populate the HTML
        // For simplicity, showing structure only
        document.getElementById('timestamp').textContent = new Date().toISOString();
        document.getElementById('environment').textContent = 'Test Environment';
    </script>
</body>
</html>
EOF
}

# Cleanup functions
cleanup() {
    log_info "Cleaning up test environment..."
    
    # Kill any remaining test processes
    pkill -f "go test" || true
    pkill -f "claude-flow" || true
    
    # Clean up temporary files
    find /tmp -name "ollama-test-*" -type d -exec rm -rf {} + 2>/dev/null || true
    
    log_info "Cleanup complete"
}

# Main execution
main() {
    local start_time=$(date +%s)
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --unit)
                RUN_UNIT_TESTS=true
                ;;
            --no-unit)
                RUN_UNIT_TESTS=false
                ;;
            --integration)
                RUN_INTEGRATION_TESTS=true
                ;;
            --no-integration)
                RUN_INTEGRATION_TESTS=false
                ;;
            --security)
                RUN_SECURITY_TESTS=true
                ;;
            --no-security)
                RUN_SECURITY_TESTS=false
                ;;
            --p2p)
                RUN_P2P_TESTS=true
                ;;
            --no-p2p)
                RUN_P2P_TESTS=false
                ;;
            --fault-tolerance)
                RUN_FAULT_TOLERANCE_TESTS=true
                ;;
            --no-fault-tolerance)
                RUN_FAULT_TOLERANCE_TESTS=false
                ;;
            --chaos)
                RUN_CHAOS_TESTS=true
                ;;
            --performance)
                RUN_PERFORMANCE_TESTS=true
                ;;
            --all)
                RUN_UNIT_TESTS=true
                RUN_INTEGRATION_TESTS=true
                RUN_SECURITY_TESTS=true
                RUN_P2P_TESTS=true
                RUN_FAULT_TOLERANCE_TESTS=true
                RUN_CHAOS_TESTS=true
                RUN_PERFORMANCE_TESTS=true
                ;;
            --no-coverage)
                GENERATE_COVERAGE=false
                ;;
            --no-reports)
                GENERATE_REPORTS=false
                ;;
            --sequential)
                PARALLEL_EXECUTION=false
                ;;
            --timeout)
                TIMEOUT="$2"
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
        shift
    done
    
    # Setup trap for cleanup
    trap cleanup EXIT
    
    log_info "Starting Ollama Distributed test suite..."
    log_info "Test configuration:"
    log_info "  Unit tests: ${RUN_UNIT_TESTS}"
    log_info "  Integration tests: ${RUN_INTEGRATION_TESTS}"
    log_info "  Security tests: ${RUN_SECURITY_TESTS}"
    log_info "  P2P tests: ${RUN_P2P_TESTS}"
    log_info "  Fault tolerance tests: ${RUN_FAULT_TOLERANCE_TESTS}"
    log_info "  Chaos tests: ${RUN_CHAOS_TESTS}"
    log_info "  Performance tests: ${RUN_PERFORMANCE_TESTS}"
    log_info "  Generate coverage: ${GENERATE_COVERAGE}"
    log_info "  Parallel execution: ${PARALLEL_EXECUTION}"
    log_info "  Timeout: ${TIMEOUT}"
    
    # Setup test environment
    setup_test_environment
    
    # Run test suites
    local exit_code=0
    
    run_unit_tests || exit_code=$?
    run_integration_tests || exit_code=$?
    run_security_tests || exit_code=$?
    run_p2p_tests || exit_code=$?
    run_fault_tolerance_tests || exit_code=$?
    run_chaos_tests || exit_code=$?
    run_performance_tests || exit_code=$?
    
    # Generate reports
    generate_coverage_report
    generate_test_reports
    
    # Calculate total execution time
    local end_time=$(date +%s)
    local total_time=$((end_time - start_time))
    
    log_info "Test execution completed in ${total_time} seconds"
    
    if [ $exit_code -eq 0 ]; then
        log_success "All enabled test suites passed!"
        log_info "Reports available in: ${REPORTS_DIR}/"
    else
        log_error "Some test suites failed. Check logs and reports for details."
        log_info "Reports available in: ${REPORTS_DIR}/"
    fi
    
    exit $exit_code
}

show_help() {
    cat << EOF
Ollama Distributed Test Automation Script

Usage: $0 [OPTIONS]

Test Suite Options:
    --unit                  Run unit tests (default: true)
    --no-unit              Skip unit tests
    --integration          Run integration tests (default: true)
    --no-integration       Skip integration tests
    --security             Run security tests (default: true)
    --no-security          Skip security tests
    --p2p                  Run P2P networking tests (default: true)
    --no-p2p               Skip P2P tests
    --fault-tolerance      Run fault tolerance tests (default: true)
    --no-fault-tolerance   Skip fault tolerance tests
    --chaos                Run chaos engineering tests (default: false)
    --performance          Run performance tests (default: false)
    --all                  Run all test suites including chaos and performance

Execution Options:
    --no-coverage          Skip coverage generation
    --no-reports           Skip report generation
    --sequential           Run tests sequentially instead of parallel
    --timeout DURATION     Set test timeout (default: 30m)

Other Options:
    --help                 Show this help message

Examples:
    $0                     # Run default test suites
    $0 --all               # Run all test suites
    $0 --unit --security   # Run only unit and security tests
    $0 --chaos --timeout 60m  # Run chaos tests with 60 minute timeout

Output:
    Test results are saved to: test_results/
    - reports/: HTML and JSON reports
    - coverage/: Coverage data and reports
    - logs/: Execution logs
EOF
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
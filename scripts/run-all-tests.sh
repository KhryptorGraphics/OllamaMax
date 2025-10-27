#!/bin/bash
##
# Unified Test Execution Script
# Orchestrates all test suites with coverage aggregation
##

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COVERAGE_THRESHOLD=90
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ARTIFACTS_DIR="${PROJECT_ROOT}/test-artifacts"
COVERAGE_DIR="${ARTIFACTS_DIR}/coverage"
LOGS_DIR="${ARTIFACTS_DIR}/logs"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  OllamaMax Unified Test Execution${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo ""

# Create necessary directories
mkdir -p "${ARTIFACTS_DIR}" "${COVERAGE_DIR}" "${LOGS_DIR}"

# Function to log test execution
log_test() {
    local test_type=$1
    local status=$2
    local message=$3

    echo "[${TIMESTAMP}] ${test_type}: ${status} - ${message}" >> "${LOGS_DIR}/test-execution.log"
}

# Function to run tests with error handling
run_test_suite() {
    local test_name=$1
    local test_command=$2
    local log_file="${LOGS_DIR}/${test_name}-${TIMESTAMP}.log"

    echo -e "${YELLOW}▶ Running ${test_name}...${NC}"

    if eval "${test_command}" > "${log_file}" 2>&1; then
        echo -e "${GREEN}✅ ${test_name} passed${NC}"
        log_test "${test_name}" "PASSED" "All tests passed"
        return 0
    else
        echo -e "${RED}❌ ${test_name} failed${NC}"
        echo -e "${RED}   Check logs: ${log_file}${NC}"
        log_test "${test_name}" "FAILED" "Tests failed - see ${log_file}"
        return 1
    fi
}

# 1. Go Unit Tests
echo -e "\n${BLUE}━━━ Phase 1: Go Unit Tests ━━━${NC}"
if run_test_suite "Go Unit Tests" "cd '${PROJECT_ROOT}' && go test -v -coverprofile='${COVERAGE_DIR}/go-unit-coverage.out' -covermode=atomic ./pkg/... ./internal/..."; then
    ((PASSED_TESTS++))
else
    ((FAILED_TESTS++))
fi
((TOTAL_TESTS++))

# 2. Go Integration Tests
echo -e "\n${BLUE}━━━ Phase 2: Go Integration Tests ━━━${NC}"
if run_test_suite "Go Integration Tests" "cd '${PROJECT_ROOT}/ollama-distributed' && make test-integration"; then
    ((PASSED_TESTS++))
else
    ((FAILED_TESTS++))
fi
((TOTAL_TESTS++))

# 3. JavaScript Unit Tests
echo -e "\n${BLUE}━━━ Phase 3: JavaScript Unit Tests ━━━${NC}"
if run_test_suite "JavaScript Unit Tests" "cd '${PROJECT_ROOT}' && npm run test:coverage"; then
    ((PASSED_TESTS++))
else
    ((FAILED_TESTS++))
fi
((TOTAL_TESTS++))

# 4. E2E Tests (install Playwright browsers first and start backend)
echo -e "\n${BLUE}━━━ Phase 4: E2E Tests ━━━${NC}"
echo -e "${YELLOW}Installing Playwright browsers...${NC}"
npx playwright install --with-deps > "${LOGS_DIR}/playwright-install-${TIMESTAMP}.log" 2>&1 || echo "Playwright browser installation failed, but continuing..."

# Check if user wants to skip backend startup (for external backend)
if [ "${SKIP_START_BACKEND:-0}" = "1" ]; then
    echo -e "${YELLOW}⚠️  SKIP_START_BACKEND=1, using external backend${NC}"
    if run_test_suite "E2E Tests" "cd '${PROJECT_ROOT}' && npm run test:e2e"; then
        ((PASSED_TESTS++))
    else
        ((FAILED_TESTS++))
    fi
    ((TOTAL_TESTS++))
else
    # Build backend if needed
    if [ ! -f "${PROJECT_ROOT}/bin/ollamamax" ]; then
        echo -e "${YELLOW}Building backend binary...${NC}"
        GOOS=$(uname -s | tr '[:upper:]' '[:lower:]')
        CGO_ENABLED=0 GOOS=${GOOS} go build -o bin/ollamamax . >> "${LOGS_DIR}/backend-build-${TIMESTAMP}.log" 2>&1
    fi

    # Start backend server with proper environment
    echo -e "${YELLOW}Starting backend server for E2E tests...${NC}"
    DB_HOST=${DB_HOST:-localhost} \
    DB_PORT=${DB_PORT:-15432} \
    REDIS_HOST=${REDIS_HOST:-localhost} \
    REDIS_PORT=${REDIS_PORT:-16379} \
    OLLAMA_API_PORT=${OLLAMA_API_PORT:-11434} \
    "${PROJECT_ROOT}/bin/ollamamax" > "${LOGS_DIR}/backend-${TIMESTAMP}.log" 2>&1 &
    BACKEND_PID=$!
    echo "${BACKEND_PID}" > "${LOGS_DIR}/backend.pid"
    echo "Backend PID: $BACKEND_PID"

    # Setup trap to ensure backend cleanup on exit
    trap "kill \$(cat ${LOGS_DIR}/backend.pid 2>/dev/null) 2>/dev/null || true; rm -f ${LOGS_DIR}/backend.pid" EXIT

    # Wait for backend to be ready with exponential backoff
    echo -e "${YELLOW}Waiting for backend to be ready...${NC}"
    BACKEND_READY=false
    API_BASE_URL=${API_BASE_URL:-http://localhost:11434}
    MAX_ATTEMPTS=30
    for i in $(seq 1 $MAX_ATTEMPTS); do
        if curl -f "${API_BASE_URL}/api/v1/health" > /dev/null 2>&1; then
            echo -e "${GREEN}✅ Backend is ready (attempt $i/$MAX_ATTEMPTS)${NC}"
            BACKEND_READY=true
            break
        fi
        echo "Attempt $i/$MAX_ATTEMPTS: Waiting for backend at ${API_BASE_URL}/api/v1/health..."
        sleep 2
    done

    if [ "$BACKEND_READY" = false ]; then
        echo -e "${RED}❌ Backend failed to start within timeout${NC}"
        echo -e "${YELLOW}Backend logs:${NC}"
        tail -n 50 "${LOGS_DIR}/backend-${TIMESTAMP}.log" || true
        kill $BACKEND_PID 2>/dev/null || true
        ((FAILED_TESTS++))
        ((TOTAL_TESTS++))
    else
        # Run E2E tests
        if run_test_suite "E2E Tests" "cd '${PROJECT_ROOT}' && npm run test:e2e"; then
            ((PASSED_TESTS++))
        else
            ((FAILED_TESTS++))
        fi
        ((TOTAL_TESTS++))

        # Stop backend (trap will also handle this)
        echo -e "${YELLOW}Stopping backend server...${NC}"
        kill $BACKEND_PID 2>/dev/null || true
        wait $BACKEND_PID 2>/dev/null || true
        rm -f "${LOGS_DIR}/backend.pid"
        echo -e "${GREEN}✅ Backend stopped${NC}"
    fi
fi

# 5. Performance Tests
echo -e "\n${BLUE}━━━ Phase 5: Performance Tests ━━━${NC}"
if run_test_suite "Performance Tests" "cd '${PROJECT_ROOT}/ollama-distributed' && make test-performance"; then
    ((PASSED_TESTS++))
else
    ((FAILED_TESTS++))
fi
((TOTAL_TESTS++))

# Merge Go coverage files
echo -e "\n${BLUE}━━━ Merging Coverage Reports ━━━${NC}"
if [ -f "${COVERAGE_DIR}/go-unit-coverage.out" ]; then
    echo "mode: atomic" > "${COVERAGE_DIR}/go-merged-coverage.out"
    tail -n +2 "${COVERAGE_DIR}/go-unit-coverage.out" >> "${COVERAGE_DIR}/go-merged-coverage.out" 2>/dev/null || true

    # Merge integration coverage if available
    if [ -f "${PROJECT_ROOT}/ollama-distributed/test-artifacts/coverage/go-integration-coverage.out" ]; then
        echo -e "${YELLOW}Merging integration test coverage...${NC}"
        tail -n +2 "${PROJECT_ROOT}/ollama-distributed/test-artifacts/coverage/go-integration-coverage.out" >> "${COVERAGE_DIR}/go-merged-coverage.out" 2>/dev/null || true
        echo -e "${GREEN}✅ Integration coverage merged${NC}"
    fi

    # Generate HTML report
    go tool cover -html="${COVERAGE_DIR}/go-merged-coverage.out" -o "${COVERAGE_DIR}/go-coverage.html"

    # Calculate coverage percentage
    GO_COVERAGE=$(go tool cover -func="${COVERAGE_DIR}/go-merged-coverage.out" | grep total | grep -Eo '[0-9]+\.[0-9]+')
    echo -e "${BLUE}Go Coverage: ${GO_COVERAGE}%${NC}"
fi

# Validate coverage threshold
echo -e "\n${BLUE}━━━ Validating Coverage Thresholds ━━━${NC}"
COVERAGE_VALIDATION_PASSED=true

# Validate JavaScript coverage
if [ -f "${PROJECT_ROOT}/coverage/coverage-summary.json" ]; then
    if ! node "${PROJECT_ROOT}/scripts/validate-coverage.js"; then
        COVERAGE_VALIDATION_PASSED=false
    fi
fi

# Validate Go coverage using awk instead of bc
if [ -n "${GO_COVERAGE}" ]; then
    if awk "BEGIN {exit !(${GO_COVERAGE} < ${COVERAGE_THRESHOLD})}"; then
        echo -e "${RED}❌ Go coverage ${GO_COVERAGE}% is below threshold ${COVERAGE_THRESHOLD}%${NC}"
        COVERAGE_VALIDATION_PASSED=false
    else
        echo -e "${GREEN}✅ Go coverage ${GO_COVERAGE}% meets threshold ${COVERAGE_THRESHOLD}%${NC}"
    fi
fi

# Generate summary report
echo -e "\n${BLUE}═══════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Test Execution Summary${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo -e "Total Test Suites:  ${TOTAL_TESTS}"
echo -e "${GREEN}Passed:            ${PASSED_TESTS}${NC}"
echo -e "${RED}Failed:            ${FAILED_TESTS}${NC}"
echo -e "${YELLOW}Skipped:           ${SKIPPED_TESTS}${NC}"
echo -e "Success Rate:       $(( PASSED_TESTS * 100 / TOTAL_TESTS ))%"
echo -e ""
echo -e "Artifacts saved to: ${ARTIFACTS_DIR}"
echo -e "Coverage reports:   ${COVERAGE_DIR}"
echo -e "Logs:               ${LOGS_DIR}"
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"

# Exit with appropriate code
if [ ${FAILED_TESTS} -gt 0 ]; then
    echo -e "\n${RED}❌ Test execution failed with ${FAILED_TESTS} failed suite(s)${NC}\n"
    exit 1
elif [ "${COVERAGE_VALIDATION_PASSED}" = false ]; then
    echo -e "\n${RED}❌ Coverage validation failed - does not meet ${COVERAGE_THRESHOLD}% threshold${NC}\n"
    exit 1
else
    echo -e "\n${GREEN}✅ All tests passed and coverage meets requirements!${NC}\n"
    exit 0
fi

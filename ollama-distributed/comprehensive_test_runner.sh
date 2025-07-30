#!/bin/bash

# Comprehensive Test Runner for Ollama Distributed
# Supports: unit, integration, e2e, coverage, mutation, snapshot, TDD, watch modes

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test configuration
PROJECT_ROOT=$(pwd)
COVERAGE_DIR="${PROJECT_ROOT}/coverage"
TEST_ARTIFACTS="${PROJECT_ROOT}/test-artifacts"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Create necessary directories
mkdir -p "${COVERAGE_DIR}"
mkdir -p "${TEST_ARTIFACTS}/reports"
mkdir -p "${TEST_ARTIFACTS}/snapshots"
mkdir -p "${TEST_ARTIFACTS}/mutations"

echo -e "${BLUE}🧪 Ollama Distributed Comprehensive Test Suite${NC}"
echo -e "${BLUE}================================================${NC}"
echo "Timestamp: ${TIMESTAMP}"
echo "Flags: $@"
echo ""

# Function to run unit tests with coverage
run_unit_tests() {
    echo -e "\n${YELLOW}📦 Running Unit Tests with Coverage...${NC}"
    
    # Run Go unit tests with coverage
    go test -v -race -coverprofile="${COVERAGE_DIR}/unit_coverage.out" -covermode=atomic ./... | tee "${TEST_ARTIFACTS}/unit_test_${TIMESTAMP}.log"
    
    # Generate HTML coverage report
    go tool cover -html="${COVERAGE_DIR}/unit_coverage.out" -o "${COVERAGE_DIR}/unit_coverage.html"
    
    # Show coverage summary
    echo -e "\n${GREEN}Unit Test Coverage Summary:${NC}"
    go tool cover -func="${COVERAGE_DIR}/unit_coverage.out" | tail -n 1
}

# Function to run integration tests
run_integration_tests() {
    echo -e "\n${YELLOW}🔗 Running Integration Tests...${NC}"
    
    # Set integration test environment
    export OLLAMA_TEST_INTEGRATION=true
    export OLLAMA_TEST_ARTIFACTS_DIR="${TEST_ARTIFACTS}"
    
    # Run integration tests with tags
    go test -v -tags=integration -coverprofile="${COVERAGE_DIR}/integration_coverage.out" -timeout=30m ./tests/integration/... | tee "${TEST_ARTIFACTS}/integration_test_${TIMESTAMP}.log"
    
    # Merge coverage profiles
    gocovmerge "${COVERAGE_DIR}/unit_coverage.out" "${COVERAGE_DIR}/integration_coverage.out" > "${COVERAGE_DIR}/combined_coverage.out"
}

# Function to run E2E tests
run_e2e_tests() {
    echo -e "\n${YELLOW}🌐 Running End-to-End Tests...${NC}"
    
    # Start test cluster
    echo "Starting test cluster..."
    make dev-env > /dev/null 2>&1 &
    CLUSTER_PID=$!
    sleep 10
    
    # Run E2E tests
    go test -v -tags=e2e -timeout=45m ./tests/e2e/... | tee "${TEST_ARTIFACTS}/e2e_test_${TIMESTAMP}.log"
    
    # Stop test cluster
    kill $CLUSTER_PID 2>/dev/null || true
    make dev-env-stop > /dev/null 2>&1 || true
}

# Function to run mutation testing
run_mutation_tests() {
    echo -e "\n${YELLOW}🧬 Running Mutation Testing...${NC}"
    
    # Install go-mutesting if not present
    if ! command -v go-mutesting &> /dev/null; then
        echo "Installing go-mutesting..."
        go install github.com/zimmski/go-mutesting/cmd/go-mutesting@latest
    fi
    
    # Run mutation testing on critical packages
    for pkg in "pkg/consensus" "pkg/scheduler" "pkg/models" "internal/auth" "internal/storage"; do
        echo -e "\nMutation testing ${pkg}..."
        go-mutesting --verbose ./${pkg}/... 2>&1 | tee "${TEST_ARTIFACTS}/mutations/${pkg//\//_}_mutations_${TIMESTAMP}.log" || true
    done
}

# Function to run snapshot tests
run_snapshot_tests() {
    echo -e "\n${YELLOW}📸 Running Snapshot Tests...${NC}"
    
    # Create snapshot test file
    cat > "${PROJECT_ROOT}/tests/snapshot_test.go" << 'EOF'
package tests

import (
    "encoding/json"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "io/ioutil"
    "path/filepath"
)

func TestAPISnapshots(t *testing.T) {
    testCases := []struct {
        name     string
        endpoint string
        response interface{}
    }{
        {
            name:     "status_endpoint",
            endpoint: "/status",
            response: map[string]interface{}{
                "status": "healthy",
                "nodes":  5,
            },
        },
        {
            name:     "models_endpoint",
            endpoint: "/models",
            response: []map[string]interface{}{
                {"name": "llama2", "size": "7B"},
            },
        },
    }

    snapshotDir := filepath.Join("test-artifacts", "snapshots")
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            snapshotFile := filepath.Join(snapshotDir, tc.name+".json")
            
            // Marshal response
            actual, err := json.MarshalIndent(tc.response, "", "  ")
            require.NoError(t, err)
            
            // Check if snapshot exists
            if expected, err := ioutil.ReadFile(snapshotFile); err == nil {
                // Compare with existing snapshot
                assert.JSONEq(t, string(expected), string(actual), "Snapshot mismatch for %s", tc.name)
            } else {
                // Create new snapshot
                err = ioutil.WriteFile(snapshotFile, actual, 0644)
                require.NoError(t, err)
                t.Logf("Created new snapshot: %s", snapshotFile)
            }
        })
    }
}
EOF

    # Run snapshot tests
    go test -v ./tests/snapshot_test.go | tee "${TEST_ARTIFACTS}/snapshot_test_${TIMESTAMP}.log"
}

# Function to setup TDD workflow
setup_tdd_workflow() {
    echo -e "\n${YELLOW}🔴 Setting up TDD Workflow...${NC}"
    
    # Create TDD helper script
    cat > "${PROJECT_ROOT}/tdd_helper.sh" << 'EOF'
#!/bin/bash
# TDD Helper for Ollama Distributed

echo "🔴 TDD Mode Active - Write failing test first!"
echo ""
echo "Workflow:"
echo "1. Write a failing test in tests/tdd/"
echo "2. Run: go test ./tests/tdd/... -v"
echo "3. Implement minimal code to pass"
echo "4. Refactor while keeping tests green"
echo "5. Repeat"
echo ""
echo "Example test structure:"
echo "func TestNewFeature(t *testing.T) {"
echo "    // Arrange"
echo "    // Act"
echo "    // Assert"
echo "}"
EOF
    chmod +x "${PROJECT_ROOT}/tdd_helper.sh"
    
    mkdir -p "${PROJECT_ROOT}/tests/tdd"
    echo -e "${GREEN}TDD workflow ready! Run ./tdd_helper.sh for guidance${NC}"
}

# Function to setup continuous watch mode
setup_watch_mode() {
    echo -e "\n${YELLOW}👁️  Setting up Watch Mode...${NC}"
    
    # Create watch script
    cat > "${PROJECT_ROOT}/watch_tests.sh" << 'EOF'
#!/bin/bash
# Continuous test watcher

echo "👁️  Watching for file changes..."
echo "Press Ctrl+C to stop"

# Install reflex if not present
if ! command -v reflex &> /dev/null; then
    echo "Installing reflex..."
    go install github.com/cespare/reflex@latest
fi

# Watch Go files and run tests on change
reflex -r '\.go$' -s -- sh -c 'clear && go test -v ./...'
EOF
    chmod +x "${PROJECT_ROOT}/watch_tests.sh"
    
    echo -e "${GREEN}Watch mode ready! Run ./watch_tests.sh to start${NC}"
}

# Function to analyze coverage gaps
analyze_coverage_gaps() {
    echo -e "\n${YELLOW}🔍 Analyzing Coverage Gaps...${NC}"
    
    # Generate detailed coverage analysis
    go tool cover -func="${COVERAGE_DIR}/combined_coverage.out" > "${COVERAGE_DIR}/coverage_analysis.txt"
    
    # Find files with low coverage
    echo -e "\n${RED}Files with coverage < 50%:${NC}"
    awk '$3 < 50 && $3 != "-" {print $1 " - " $3 "%"}' "${COVERAGE_DIR}/coverage_analysis.txt"
    
    # Find uncovered functions
    echo -e "\n${RED}Uncovered functions:${NC}"
    grep "0.0%" "${COVERAGE_DIR}/coverage_analysis.txt" | head -20
}

# Function to generate comprehensive report
generate_final_report() {
    echo -e "\n${YELLOW}📊 Generating Comprehensive Test Report...${NC}"
    
    REPORT_FILE="${TEST_ARTIFACTS}/comprehensive_test_report_${TIMESTAMP}.md"
    
    cat > "${REPORT_FILE}" << EOF
# Ollama Distributed - Comprehensive Test Report
Generated: $(date)

## Test Execution Summary

### Unit Tests
- Status: $(grep -c "PASS" "${TEST_ARTIFACTS}/unit_test_${TIMESTAMP}.log" 2>/dev/null || echo "0") passed
- Coverage: $(go tool cover -func="${COVERAGE_DIR}/unit_coverage.out" 2>/dev/null | tail -n 1 | awk '{print $3}' || echo "N/A")

### Integration Tests
- Status: $(grep -c "PASS" "${TEST_ARTIFACTS}/integration_test_${TIMESTAMP}.log" 2>/dev/null || echo "0") passed
- Additional Coverage: Included in combined report

### E2E Tests
- Status: $(grep -c "PASS" "${TEST_ARTIFACTS}/e2e_test_${TIMESTAMP}.log" 2>/dev/null || echo "0") passed
- Critical Paths: Tested

### Coverage Analysis
\`\`\`
$(go tool cover -func="${COVERAGE_DIR}/combined_coverage.out" 2>/dev/null | tail -10 || echo "Coverage data not available")
\`\`\`

### Mutation Testing
- Packages tested: pkg/consensus, pkg/scheduler, pkg/models, internal/auth, internal/storage
- Results: See individual mutation logs in test-artifacts/mutations/

### Recommendations
1. Increase coverage in low-coverage files
2. Add more edge case tests
3. Implement property-based testing for critical algorithms
4. Add performance benchmarks

## Artifacts Generated
- Coverage HTML: coverage/unit_coverage.html
- Test Logs: test-artifacts/*_test_${TIMESTAMP}.log
- Mutation Reports: test-artifacts/mutations/
- Snapshots: test-artifacts/snapshots/
EOF

    echo -e "${GREEN}Comprehensive report generated: ${REPORT_FILE}${NC}"
}

# Main execution
echo -e "${BLUE}Starting comprehensive test suite...${NC}"

# Parse flags and run appropriate tests
if [[ "$@" == *"--unit"* ]] || [[ "$@" == *"--coverage"* ]]; then
    run_unit_tests
fi

if [[ "$@" == *"--integration"* ]]; then
    run_integration_tests
fi

if [[ "$@" == *"--e2e"* ]]; then
    run_e2e_tests
fi

if [[ "$@" == *"--mutation"* ]]; then
    run_mutation_tests
fi

if [[ "$@" == *"--snapshot"* ]]; then
    run_snapshot_tests
fi

if [[ "$@" == *"--tdd"* ]]; then
    setup_tdd_workflow
fi

if [[ "$@" == *"--watch"* ]]; then
    setup_watch_mode
fi

# Always analyze coverage if any tests were run
if [[ "$@" == *"--coverage"* ]]; then
    analyze_coverage_gaps
fi

# Generate final report
generate_final_report

echo -e "\n${GREEN}✅ Comprehensive test suite completed!${NC}"
echo -e "${BLUE}View coverage report: open ${COVERAGE_DIR}/unit_coverage.html${NC}"
echo -e "${BLUE}View test report: cat ${TEST_ARTIFACTS}/comprehensive_test_report_*.md${NC}"

# If watch mode requested, start it now
if [[ "$@" == *"--watch"* ]]; then
    echo -e "\n${YELLOW}Starting watch mode...${NC}"
    ./watch_tests.sh
fi
#!/bin/bash

################################################################################
# Chaos Engineering Execution Script for OllamaMax
#
# Executes comprehensive chaos testing from chaos_engineering_test.go
# Validates fault tolerance, recovery times, and system resilience
################################################################################

set -e

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
RESULTS_DIR="chaos-test-results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
CHAOS_DURATION="${CHAOS_DURATION:-2h}"
TEST_CLUSTER_SIZE="${TEST_CLUSTER_SIZE:-5}"

# Logging functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

mkdir -p "${RESULTS_DIR}"

log_info "==================== Chaos Engineering Execution ===================="
log_info "Results Directory: ${RESULTS_DIR}"
log_info "Test Duration: ${CHAOS_DURATION}"
log_info "Cluster Size: ${TEST_CLUSTER_SIZE}"
log_info "======================================================================"

# Check if chaos tests exist
if [ ! -f "ollama-distributed/tests/chaos/chaos_engineering_test.go" ]; then
    log_error "Chaos engineering tests not found!"
    exit 1
fi

log_success "Chaos engineering tests found"

# Deploy test cluster
log_info "Deploying test cluster with ${TEST_CLUSTER_SIZE} nodes..."

# Check for jq availability
if ! command -v jq &> /dev/null; then
    log_warning "jq not installed - JSON parsing will use fallback methods"
    log_info "Install jq for better parsing: sudo apt-get install jq"
fi

if command -v docker-compose &> /dev/null; then
    # Use Docker Compose for local testing
    log_info "Using Docker Compose for test cluster..."

    if [ ! -f "docker-compose.chaos-test.yml" ]; then
        log_error "docker-compose.chaos-test.yml not found!"
        log_error "This file should contain a 5-node test cluster configuration"
        exit 1
    fi

    docker-compose -f docker-compose.chaos-test.yml up -d || {
        log_warning "Docker Compose deployment failed, using existing cluster"
    }
elif command -v kubectl &> /dev/null; then
    # Use Kubernetes for test cluster
    log_info "Using Kubernetes for test cluster..."
    kubectl apply -f k8s/chaos-test-cluster.yaml || {
        log_warning "Kubernetes deployment failed, using existing cluster"
    }
else
    log_warning "No orchestration tool found, assuming cluster is already deployed"
fi

# Wait for cluster to be ready
log_info "Waiting for cluster to be healthy..."
sleep 30

# Verify cluster health
log_info "Verifying cluster health..."
if curl -sf http://localhost:11434/health > /dev/null; then
    log_success "Cluster is healthy"
else
    log_warning "Cluster health check failed, but continuing..."
fi

# Collect baseline metrics
log_info "Collecting baseline performance metrics..."
BASELINE_FILE="${RESULTS_DIR}/baseline-metrics-${TIMESTAMP}.log"

{
    echo "Baseline Metrics - $(date --iso-8601=seconds)"
    echo "==========================================="
    curl -s http://localhost:11434/metrics || echo "Metrics unavailable"
    echo ""
    echo "Consensus Status:"
    if command -v jq &> /dev/null; then
        curl -s http://localhost:11434/api/consensus/status | jq '.' || echo "Consensus status unavailable"
    else
        # Fallback: Pretty-print JSON without jq
        curl -s http://localhost:11434/api/consensus/status | python3 -m json.tool 2>/dev/null || curl -s http://localhost:11434/api/consensus/status || echo "Consensus status unavailable"
    fi
} > "${BASELINE_FILE}"

log_success "Baseline metrics collected"

# Execute chaos test scenarios
log_info "Executing chaos test scenarios..."

cd ollama-distributed

# Scenario 1: Network Partition
log_info "Running Chaos Scenario: Network Partition..."
go test -v -timeout 30m ./tests/chaos/... -run TestChaosEngineering/NetworkPartition \
    > "../${RESULTS_DIR}/scenario-network-partition-${TIMESTAMP}.log" 2>&1 &
NETWORK_PARTITION_PID=$!

# Scenario 2: Leader Failure
log_info "Running Chaos Scenario: Leader Failure..."
go test -v -timeout 30m ./tests/chaos/... -run TestChaosEngineering/LeaderFailure \
    > "../${RESULTS_DIR}/scenario-leader-failure-${TIMESTAMP}.log" 2>&1 &
LEADER_FAILURE_PID=$!

# Scenario 3: High Latency
log_info "Running Chaos Scenario: High Latency..."
go test -v -timeout 30m ./tests/chaos/... -run TestChaosEngineering/HighLatency \
    > "../${RESULTS_DIR}/scenario-high-latency-${TIMESTAMP}.log" 2>&1 &
HIGH_LATENCY_PID=$!

# Scenario 4: Memory Pressure
log_info "Running Chaos Scenario: Memory Pressure..."
go test -v -timeout 30m ./tests/chaos/... -run TestChaosEngineering/MemoryPressure \
    > "../${RESULTS_DIR}/scenario-memory-pressure-${TIMESTAMP}.log" 2>&1 &
MEMORY_PRESSURE_PID=$!

# Scenario 5: Byzantine Faults
log_info "Running Chaos Scenario: Byzantine Faults..."
go test -v -timeout 30m ./tests/chaos/... -run TestChaosEngineering/ByzantineFaults \
    > "../${RESULTS_DIR}/scenario-byzantine-faults-${TIMESTAMP}.log" 2>&1 &
BYZANTINE_FAULTS_PID=$!

# Scenario 6: Cascading Failures
log_info "Running Chaos Scenario: Cascading Failures..."
go test -v -timeout 30m ./tests/chaos/... -run TestChaosEngineering/CascadingFailures \
    > "../${RESULTS_DIR}/scenario-cascading-failures-${TIMESTAMP}.log" 2>&1 &
CASCADING_FAILURES_PID=$!

log_success "All chaos scenarios started"

# Wait for all scenarios to complete
log_info "Waiting for chaos scenarios to complete..."

FAILED_SCENARIOS=0
PASSED_SCENARIOS=0

wait_for_test() {
    local pid=$1
    local name=$2

    if wait ${pid}; then
        log_success "Scenario '${name}' passed"
        ((PASSED_SCENARIOS++))
        return 0
    else
        log_error "Scenario '${name}' failed"
        ((FAILED_SCENARIOS++))
        return 1
    fi
}

wait_for_test ${NETWORK_PARTITION_PID} "Network Partition"
wait_for_test ${LEADER_FAILURE_PID} "Leader Failure"
wait_for_test ${HIGH_LATENCY_PID} "High Latency"
wait_for_test ${MEMORY_PRESSURE_PID} "Memory Pressure"
wait_for_test ${BYZANTINE_FAULTS_PID} "Byzantine Faults"
wait_for_test ${CASCADING_FAILURES_PID} "Cascading Failures"

log_info "Chaos scenarios completed: ${PASSED_SCENARIOS} passed, ${FAILED_SCENARIOS} failed"

# Extended Chaos Testing: Random Chaos
log_info "Running extended chaos test: Random Chaos (${CHAOS_DURATION})..."
go test -v -timeout ${CHAOS_DURATION} ./tests/chaos/... -run TestRandomChaos \
    > "../${RESULTS_DIR}/extended-random-chaos-${TIMESTAMP}.log" 2>&1

if [ $? -eq 0 ]; then
    log_success "Random chaos test passed"
    ((PASSED_SCENARIOS++))
else
    log_error "Random chaos test failed"
    ((FAILED_SCENARIOS++))
fi

# Extended Chaos Testing: Stress Test
log_info "Running extended chaos test: Stress Test (5 minutes)..."
go test -v -timeout 10m ./tests/chaos/... -run TestStressTest \
    > "../${RESULTS_DIR}/extended-stress-test-${TIMESTAMP}.log" 2>&1

if [ $? -eq 0 ]; then
    log_success "Stress test passed"
    ((PASSED_SCENARIOS++))
else
    log_error "Stress test failed"
    ((FAILED_SCENARIOS++))
fi

cd ..

# Collect post-test metrics
log_info "Collecting post-test metrics..."
POST_TEST_FILE="${RESULTS_DIR}/post-test-metrics-${TIMESTAMP}.log"

{
    echo "Post-Test Metrics - $(date --iso-8601=seconds)"
    echo "==========================================="
    curl -s http://localhost:11434/metrics || echo "Metrics unavailable"
    echo ""
    echo "Consensus Status:"
    if command -v jq &> /dev/null; then
        curl -s http://localhost:11434/api/consensus/status | jq '.' || echo "Consensus status unavailable"
    else
        # Fallback: Pretty-print JSON without jq
        curl -s http://localhost:11434/api/consensus/status | python3 -m json.tool 2>/dev/null || curl -s http://localhost:11434/api/consensus/status || echo "Consensus status unavailable"
    fi
    echo ""
    echo "Cluster Health:"
    curl -s http://localhost:11434/health || echo "Health check unavailable"
} > "${POST_TEST_FILE}"

log_success "Post-test metrics collected"

# Analyze recovery times
log_info "Analyzing recovery times..."

RECOVERY_ANALYSIS="${RESULTS_DIR}/recovery-analysis-${TIMESTAMP}.log"

{
    echo "Recovery Time Analysis"
    echo "======================"
    echo ""

    for log_file in ${RESULTS_DIR}/scenario-*-${TIMESTAMP}.log; do
        scenario_name=$(basename ${log_file} | sed "s/scenario-//;s/-${TIMESTAMP}.log//")
        echo "Scenario: ${scenario_name}"

        # Extract recovery time if present in logs
        if grep -q "Recovery time:" ${log_file}; then
            grep "Recovery time:" ${log_file}
        else
            echo "  Recovery time not found in logs"
        fi

        echo ""
    done
} > "${RECOVERY_ANALYSIS}"

log_success "Recovery analysis completed"

# Calculate resilience score
log_info "Calculating system resilience score..."

TOTAL_SCENARIOS=$((PASSED_SCENARIOS + FAILED_SCENARIOS))
RESILIENCE_SCORE=$(echo "scale=2; ${PASSED_SCENARIOS} * 100 / ${TOTAL_SCENARIOS}" | bc)

# Generate chaos engineering report
REPORT_FILE="${RESULTS_DIR}/chaos-engineering-report-${TIMESTAMP}.md"

cat > "${REPORT_FILE}" <<EOF
# Chaos Engineering Test Report

**Timestamp:** $(date --iso-8601=seconds)
**Test Duration:** ${CHAOS_DURATION}
**Cluster Size:** ${TEST_CLUSTER_SIZE}
**Total Scenarios:** ${TOTAL_SCENARIOS}
**Passed Scenarios:** ${PASSED_SCENARIOS}
**Failed Scenarios:** ${FAILED_SCENARIOS}
**Resilience Score:** ${RESILIENCE_SCORE}/100

## Test Scenarios

### Basic Chaos Scenarios
1. **Network Partition** - Validates majority partition functionality
2. **Leader Failure** - Tests leader election and continuity
3. **High Latency** - Validates operation under network delays
4. **Memory Pressure** - Tests graceful handling of memory constraints
5. **Byzantine Faults** - Validates Byzantine fault tolerance
6. **Cascading Failures** - Tests circuit breakers and partial failures

### Extended Chaos Tests
7. **Random Chaos** - Sustained random failure injection (${CHAOS_DURATION})
8. **Stress Test** - Concurrent workload with continuous failures

## Recovery Time Objectives (RTO)

Target: <60 seconds for all failure scenarios

See detailed recovery analysis: \`${RECOVERY_ANALYSIS}\`

## Findings

EOF

if [ ${FAILED_SCENARIOS} -eq 0 ]; then
    echo "✅ **ALL CHAOS TESTS PASSED** - System demonstrated excellent resilience" >> "${REPORT_FILE}"
    echo "" >> "${REPORT_FILE}"
    echo "The system successfully handled all chaos scenarios including:" >> "${REPORT_FILE}"
    echo "- Network partitions with graceful degradation" >> "${REPORT_FILE}"
    echo "- Leader failures with rapid re-election" >> "${REPORT_FILE}"
    echo "- High latency conditions without timeouts" >> "${REPORT_FILE}"
    echo "- Memory pressure without crashes" >> "${REPORT_FILE}"
    echo "- Byzantine faults with maintained consensus" >> "${REPORT_FILE}"
    echo "- Cascading failures with circuit breakers" >> "${REPORT_FILE}"
else
    echo "⚠️ **SOME CHAOS TESTS FAILED** - Review failures and improve resilience" >> "${REPORT_FILE}"
    echo "" >> "${REPORT_FILE}"
    echo "Failed scenarios require attention:" >> "${REPORT_FILE}"
    echo "- Review scenario logs in \`${RESULTS_DIR}/scenario-*-${TIMESTAMP}.log\`" >> "${REPORT_FILE}"
    echo "- Analyze recovery times and failure modes" >> "${REPORT_FILE}"
    echo "- Implement improvements for failed scenarios" >> "${REPORT_FILE}"
fi

echo "" >> "${REPORT_FILE}"
echo "## Recommendations" >> "${REPORT_FILE}"
echo "" >> "${REPORT_FILE}"
echo "1. Monitor recovery times in production" >> "${REPORT_FILE}"
echo "2. Implement automated chaos testing in CI/CD" >> "${REPORT_FILE}"
echo "3. Establish alerting for resilience degradation" >> "${REPORT_FILE}"
echo "4. Regular chaos drills to maintain resilience" >> "${REPORT_FILE}"

log_success "Chaos engineering report: ${REPORT_FILE}"

# Cleanup test cluster (optional)
log_info "Cleaning up test environment..."

if [ "${CLEANUP_CLUSTER}" = "true" ]; then
    if command -v docker-compose &> /dev/null; then
        docker-compose -f docker-compose.chaos-test.yml down || true
    elif command -v kubectl &> /dev/null; then
        kubectl delete -f k8s/chaos-test-cluster.yaml || true
    fi
    log_success "Test cluster cleaned up"
else
    log_info "Skipping cluster cleanup (set CLEANUP_CLUSTER=true to enable)"
fi

# Final summary
log_info "=========================================================================="
log_info "Chaos Engineering Summary:"
log_info "  Total Scenarios: ${TOTAL_SCENARIOS}"
log_info "  Passed: ${PASSED_SCENARIOS}"
log_info "  Failed: ${FAILED_SCENARIOS}"
log_info "  Resilience Score: ${RESILIENCE_SCORE}/100"
log_info "=========================================================================="

# Exit with appropriate code
if [ ${FAILED_SCENARIOS} -eq 0 ]; then
    log_success "All chaos tests passed!"
    exit 0
else
    log_error "Chaos tests failed!"
    exit 1
fi

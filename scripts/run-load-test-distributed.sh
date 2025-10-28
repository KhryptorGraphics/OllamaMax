#!/bin/bash

################################################################################
# Distributed Load Test Orchestrator for OllamaMax
#
# Coordinates multiple k6 instances to achieve 100K+ RPS load testing
# Supports both local (Docker) and cloud (K8s) execution
################################################################################

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TARGET_RPS="${TARGET_RPS:-100000}"
BASE_URL="${BASE_URL:-http://localhost:11434}"
INSTANCES="${K6_INSTANCES:-10}"  # Number of k6 instances (10K RPS per instance)
EXECUTION_MODE="${EXECUTION_MODE:-local}"  # local or k8s
RESULTS_DIR="load-test-results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

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

# Create results directory
mkdir -p "${RESULTS_DIR}"

log_info "==================== Distributed Load Test Orchestrator ===================="
log_info "Target RPS: ${TARGET_RPS}"
log_info "Base URL: ${BASE_URL}"
log_info "K6 Instances: ${INSTANCES}"
log_info "Execution Mode: ${EXECUTION_MODE}"
log_info "Results Directory: ${RESULTS_DIR}"
log_info "=========================================================================="

# Pre-test validation
log_info "Running pre-test validation..."

# Check if k6 is installed
if ! command -v k6 &> /dev/null; then
    log_error "k6 not found. Please install k6: https://k6.io/docs/getting-started/installation/"
    exit 1
fi

log_success "k6 is installed: $(k6 version)"

# Check if target system is accessible
log_info "Checking target system accessibility: ${BASE_URL}"
if curl -s -f "${BASE_URL}/health" > /dev/null 2>&1; then
    log_success "Target system is accessible"
else
    log_warning "Target system health check failed, but continuing..."
fi

# Check system resources
log_info "Checking system resources..."
TOTAL_MEM=$(free -g | awk '/^Mem:/{print $2}')
AVAILABLE_MEM=$(free -g | awk '/^Mem:/{print $7}')
CPU_CORES=$(nproc)

log_info "  CPU Cores: ${CPU_CORES}"
log_info "  Total Memory: ${TOTAL_MEM}GB"
log_info "  Available Memory: ${AVAILABLE_MEM}GB"

if [ "${AVAILABLE_MEM}" -lt 8 ]; then
    log_warning "Low memory available (${AVAILABLE_MEM}GB). Recommended: 16GB+"
fi

if [ "${CPU_CORES}" -lt 8 ]; then
    log_warning "Low CPU cores (${CPU_CORES}). Recommended: 16+ cores for 100K RPS"
fi

# Check if load test script exists
if [ ! -f "load-test-distributed.js" ]; then
    log_error "load-test-distributed.js not found!"
    exit 1
fi

log_success "Pre-test validation completed"

# Function to run k6 instance
run_k6_instance() {
    local instance_id=$1
    local rps_per_instance=$((TARGET_RPS / INSTANCES))

    log_info "Starting k6 instance ${instance_id}/${INSTANCES} (Target: ${rps_per_instance} RPS)..."

    K6_INSTANCE_ID="${instance_id}" \
    K6_TOTAL_INSTANCES="${INSTANCES}" \
    TARGET_RPS="${TARGET_RPS}" \
    BASE_URL="${BASE_URL}" \
    k6 run \
        --out json="${RESULTS_DIR}/metrics-instance-${instance_id}-${TIMESTAMP}.json" \
        load-test-distributed.js \
        > "${RESULTS_DIR}/output-instance-${instance_id}-${TIMESTAMP}.log" 2>&1 &

    echo $!
}

# Function to monitor k6 instance
monitor_instance() {
    local pid=$1
    local instance_id=$2

    while kill -0 $pid 2>/dev/null; do
        sleep 10
        if [ -f "${RESULTS_DIR}/output-instance-${instance_id}-${TIMESTAMP}.log" ]; then
            # Check for errors
            if grep -q "ERRO" "${RESULTS_DIR}/output-instance-${instance_id}-${TIMESTAMP}.log"; then
                log_warning "Instance ${instance_id} encountered errors"
            fi
        fi
    done

    wait $pid
    local exit_code=$?

    if [ $exit_code -eq 0 ]; then
        log_success "Instance ${instance_id} completed successfully"
    else
        log_error "Instance ${instance_id} failed with exit code ${exit_code}"
    fi

    return $exit_code
}

# Function to collect resource metrics
collect_system_metrics() {
    local metrics_file="${RESULTS_DIR}/system-metrics-${TIMESTAMP}.log"

    log_info "Starting system metrics collection..."

    while true; do
        {
            echo "Timestamp: $(date --iso-8601=seconds)"
            echo "CPU Usage:"
            top -bn1 | grep "Cpu(s)" || mpstat 1 1
            echo ""
            echo "Memory Usage:"
            free -h
            echo ""
            echo "Network Stats:"
            ss -s || netstat -s | head -20
            echo ""
            echo "Disk I/O:"
            if command -v iostat &> /dev/null; then
                iostat -x 1 1
            elif [ -f /proc/diskstats ]; then
                # Fallback: Use /proc/diskstats for basic disk I/O metrics
                echo "Device:    read_ops  write_ops  read_MB  write_MB"
                awk '{printf "%-10s %8s  %8s  %7.2f  %7.2f\n", $3, $4, $8, $6/2048, $10/2048}' /proc/diskstats | grep -v "loop\|ram"
            else
                echo "iostat not available and /proc/diskstats not accessible"
                echo "Install sysstat for detailed I/O metrics: sudo apt-get install sysstat"
            fi
            echo "---"
            echo ""
        } >> "${metrics_file}"

        sleep 30
    done &

    echo $!
}

# Main execution
log_info "Starting distributed load test execution..."

# Start system metrics collection
METRICS_PID=$(collect_system_metrics)
log_info "System metrics collection started (PID: ${METRICS_PID})"

# Array to store k6 instance PIDs
declare -a K6_PIDS

# Start all k6 instances in parallel
for i in $(seq 1 ${INSTANCES}); do
    pid=$(run_k6_instance $i)
    K6_PIDS[$i]=$pid
    log_info "Started k6 instance $i with PID ${pid}"
    sleep 2  # Small delay between instance starts
done

log_success "All ${INSTANCES} k6 instances started"
log_info "Test duration: ~100 minutes (ramp-up, sustained, spike, peak, stress, extreme, ramp-down)"
log_info "Monitor progress in real-time: tail -f ${RESULTS_DIR}/output-instance-*-${TIMESTAMP}.log"

# Wait for all instances to complete
log_info "Waiting for all k6 instances to complete..."

FAILED_INSTANCES=0
for i in $(seq 1 ${INSTANCES}); do
    pid=${K6_PIDS[$i]}
    if ! monitor_instance $pid $i; then
        ((FAILED_INSTANCES++))
    fi
done

# Stop system metrics collection
kill ${METRICS_PID} 2>/dev/null || true
log_success "System metrics collection stopped"

log_info "All k6 instances completed. Failed instances: ${FAILED_INSTANCES}/${INSTANCES}"

# Aggregate results
log_info "Aggregating results from all instances..."

AGGREGATE_FILE="${RESULTS_DIR}/aggregate-results-${TIMESTAMP}.json"

# Initialize aggregate metrics
cat > "${AGGREGATE_FILE}" <<EOF
{
  "timestamp": "$(date --iso-8601=seconds)",
  "target_rps": ${TARGET_RPS},
  "instances": ${INSTANCES},
  "failed_instances": ${FAILED_INSTANCES},
  "metrics": {
    "total_requests": 0,
    "total_failed_requests": 0,
    "average_rps": 0,
    "average_duration_ms": 0,
    "p95_duration_ms": 0,
    "p99_duration_ms": 0,
    "error_rate": 0
  },
  "instances_results": []
}
EOF

# Parse and aggregate metrics from each instance
TOTAL_REQUESTS=0
TOTAL_FAILED=0
TOTAL_RPS=0
DURATIONS_SUM=0
P95_VALUES=()
P99_VALUES=()

for i in $(seq 1 ${INSTANCES}); do
    summary_file="${RESULTS_DIR}/summary-instance-${i}.json"

    if [ -f "${summary_file}" ]; then
        # Extract key metrics using jq (if available) or grep/awk fallback
        if command -v jq &> /dev/null; then
            requests=$(jq -r '.metrics.http_reqs.values.count // 0' "${summary_file}")
            failed=$(jq -r '.metrics.http_req_failed.values.passes // 0' "${summary_file}")
            rps=$(jq -r '.metrics.http_reqs.values.rate // 0' "${summary_file}")
            avg_duration=$(jq -r '.metrics.http_req_duration.values.avg // 0' "${summary_file}")
            p95=$(jq -r '.metrics.http_req_duration.values["p(95)"] // 0' "${summary_file}")
            p99=$(jq -r '.metrics.http_req_duration.values["p(99)"] // 0' "${summary_file}")

            TOTAL_REQUESTS=$((TOTAL_REQUESTS + requests))
            TOTAL_FAILED=$((TOTAL_FAILED + failed))
            TOTAL_RPS=$(echo "${TOTAL_RPS} + ${rps}" | bc)
            DURATIONS_SUM=$(echo "${DURATIONS_SUM} + ${avg_duration}" | bc)
            P95_VALUES+=("${p95}")
            P99_VALUES+=("${p99}")

            log_info "Instance ${i}: ${requests} requests, ${rps} RPS, P95: ${p95}ms, P99: ${p99}ms"
        else
            # Fallback: Use grep/awk for basic metric extraction
            log_warning "jq not available, using grep/awk fallback (install jq for better parsing: sudo apt-get install jq)"

            # Extract using grep and awk as fallback
            requests=$(grep -oP '"http_reqs".*?"count":\s*\K[0-9]+' "${summary_file}" 2>/dev/null | head -1 || echo "0")
            failed=$(grep -oP '"http_req_failed".*?"passes":\s*\K[0-9]+' "${summary_file}" 2>/dev/null | head -1 || echo "0")
            rps=$(grep -oP '"http_reqs".*?"rate":\s*\K[0-9.]+' "${summary_file}" 2>/dev/null | head -1 || echo "0")

            TOTAL_REQUESTS=$((TOTAL_REQUESTS + requests))
            TOTAL_FAILED=$((TOTAL_FAILED + failed))
            TOTAL_RPS=$(echo "${TOTAL_RPS} + ${rps}" | bc 2>/dev/null || echo "${TOTAL_RPS}")

            log_info "Instance ${i}: ${requests} requests (basic metrics, jq recommended for full details)"
        fi
    else
        log_warning "Summary file not found for instance ${i}: ${summary_file}"
    fi
done

# Calculate aggregate statistics
if [ ${INSTANCES} -gt 0 ]; then
    AVG_DURATION=$(echo "scale=2; ${DURATIONS_SUM} / ${INSTANCES}" | bc)
    ERROR_RATE=$(echo "scale=4; ${TOTAL_FAILED} / ${TOTAL_REQUESTS} * 100" | bc)

    # Calculate P95 and P99 across all instances (simple average for now)
    P95_SUM=0
    P99_SUM=0

    for val in "${P95_VALUES[@]}"; do
        P95_SUM=$(echo "${P95_SUM} + ${val}" | bc)
    done

    for val in "${P99_VALUES[@]}"; do
        P99_SUM=$(echo "${P99_SUM} + ${val}" | bc)
    done

    AVG_P95=$(echo "scale=2; ${P95_SUM} / ${INSTANCES}" | bc)
    AVG_P99=$(echo "scale=2; ${P99_SUM} / ${INSTANCES}" | bc)
fi

# Generate aggregate report
log_info "=========================================================================="
log_info "                    LOAD TEST RESULTS SUMMARY"
log_info "=========================================================================="
log_info "Target RPS: ${TARGET_RPS}"
log_info "Achieved RPS: ${TOTAL_RPS}"
log_info "Total Requests: ${TOTAL_REQUESTS}"
log_info "Failed Requests: ${TOTAL_FAILED}"
log_info "Error Rate: ${ERROR_RATE}%"
log_info "Average Duration: ${AVG_DURATION}ms"
log_info "P95 Duration: ${AVG_P95}ms"
log_info "P99 Duration: ${AVG_P99}ms"
log_info "Failed Instances: ${FAILED_INSTANCES}/${INSTANCES}"
log_info "=========================================================================="

# Generate HTML report
log_info "Generating HTML report..."

HTML_REPORT="${RESULTS_DIR}/load-test-report-${TIMESTAMP}.html"

cat > "${HTML_REPORT}" <<'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>OllamaMax Load Test Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        h1 { color: #333; border-bottom: 3px solid #4CAF50; padding-bottom: 10px; }
        h2 { color: #555; margin-top: 30px; }
        .metric { display: inline-block; margin: 15px; padding: 20px; background: #f9f9f9; border-radius: 5px; min-width: 200px; }
        .metric-label { font-size: 14px; color: #666; }
        .metric-value { font-size: 28px; font-weight: bold; color: #333; margin-top: 5px; }
        .success { color: #4CAF50; }
        .warning { color: #FF9800; }
        .error { color: #F44336; }
        table { width: 100%; border-collapse: collapse; margin-top: 20px; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background: #4CAF50; color: white; }
        tr:hover { background: #f5f5f5; }
    </style>
</head>
<body>
    <div class="container">
        <h1>OllamaMax Distributed Load Test Report</h1>
        <p><strong>Timestamp:</strong> TIMESTAMP_PLACEHOLDER</p>
        <p><strong>Target RPS:</strong> TARGET_RPS_PLACEHOLDER</p>
        <p><strong>K6 Instances:</strong> INSTANCES_PLACEHOLDER</p>

        <h2>Key Metrics</h2>
        <div>
            <div class="metric">
                <div class="metric-label">Total Requests</div>
                <div class="metric-value">TOTAL_REQUESTS_PLACEHOLDER</div>
            </div>
            <div class="metric">
                <div class="metric-label">Achieved RPS</div>
                <div class="metric-value ACHIEVED_RPS_CLASS_PLACEHOLDER">TOTAL_RPS_PLACEHOLDER</div>
            </div>
            <div class="metric">
                <div class="metric-label">Error Rate</div>
                <div class="metric-value ERROR_RATE_CLASS_PLACEHOLDER">ERROR_RATE_PLACEHOLDER%</div>
            </div>
            <div class="metric">
                <div class="metric-label">P95 Latency</div>
                <div class="metric-value P95_CLASS_PLACEHOLDER">AVG_P95_PLACEHOLDERms</div>
            </div>
            <div class="metric">
                <div class="metric-label">P99 Latency</div>
                <div class="metric-value P99_CLASS_PLACEHOLDER">AVG_P99_PLACEHOLDERms</div>
            </div>
        </div>

        <h2>Test Result: TEST_RESULT_PLACEHOLDER</h2>
        <p>RESULT_DESCRIPTION_PLACEHOLDER</p>

        <h2>Detailed Findings</h2>
        <ul>
            <li><strong>Performance:</strong> PERFORMANCE_FINDING_PLACEHOLDER</li>
            <li><strong>Reliability:</strong> RELIABILITY_FINDING_PLACEHOLDER</li>
            <li><strong>Scalability:</strong> SCALABILITY_FINDING_PLACEHOLDER</li>
        </ul>

        <h2>Recommendations</h2>
        <ul>
            RECOMMENDATIONS_PLACEHOLDER
        </ul>
    </div>
</body>
</html>
EOF

# Populate HTML report with actual values
sed -i "s/TIMESTAMP_PLACEHOLDER/$(date --iso-8601=seconds)/g" "${HTML_REPORT}"
sed -i "s/TARGET_RPS_PLACEHOLDER/${TARGET_RPS}/g" "${HTML_REPORT}"
sed -i "s/INSTANCES_PLACEHOLDER/${INSTANCES}/g" "${HTML_REPORT}"
sed -i "s/TOTAL_REQUESTS_PLACEHOLDER/${TOTAL_REQUESTS}/g" "${HTML_REPORT}"
sed -i "s/TOTAL_RPS_PLACEHOLDER/${TOTAL_RPS}/g" "${HTML_REPORT}"
sed -i "s/ERROR_RATE_PLACEHOLDER/${ERROR_RATE}/g" "${HTML_REPORT}"
sed -i "s/AVG_P95_PLACEHOLDER/${AVG_P95}/g" "${HTML_REPORT}"
sed -i "s/AVG_P99_PLACEHOLDER/${AVG_P99}/g" "${HTML_REPORT}"

# Determine test result
if (( $(echo "${TOTAL_RPS} >= ${TARGET_RPS}" | bc -l) )) && (( $(echo "${ERROR_RATE} < 0.1" | bc -l) )) && (( $(echo "${AVG_P95} < 500" | bc -l) )); then
    TEST_RESULT="✅ PASSED"
    RESULT_DESC="Load test successfully achieved target RPS with acceptable latency and error rate."
    sed -i "s/ACHIEVED_RPS_CLASS_PLACEHOLDER/success/g" "${HTML_REPORT}"
    sed -i "s/ERROR_RATE_CLASS_PLACEHOLDER/success/g" "${HTML_REPORT}"
    sed -i "s/P95_CLASS_PLACEHOLDER/success/g" "${HTML_REPORT}"
    sed -i "s/P99_CLASS_PLACEHOLDER/success/g" "${HTML_REPORT}"
else
    TEST_RESULT="❌ FAILED"
    RESULT_DESC="Load test did not meet all performance targets. Review detailed metrics below."
    sed -i "s/ACHIEVED_RPS_CLASS_PLACEHOLDER/error/g" "${HTML_REPORT}"
    sed -i "s/ERROR_RATE_CLASS_PLACEHOLDER/error/g" "${HTML_REPORT}"
    sed -i "s/P95_CLASS_PLACEHOLDER/warning/g" "${HTML_REPORT}"
    sed -i "s/P99_CLASS_PLACEHOLDER/warning/g" "${HTML_REPORT}"
fi

sed -i "s/TEST_RESULT_PLACEHOLDER/${TEST_RESULT}/g" "${HTML_REPORT}"
sed -i "s/RESULT_DESCRIPTION_PLACEHOLDER/${RESULT_DESC}/g" "${HTML_REPORT}"

# Add findings and recommendations
PERF_FINDING="System handled ${TOTAL_RPS} RPS with P95 latency of ${AVG_P95}ms"
REL_FINDING="Error rate of ${ERROR_RATE}% across ${TOTAL_REQUESTS} requests"
SCALE_FINDING="Test executed across ${INSTANCES} k6 instances with ${FAILED_INSTANCES} failures"

sed -i "s/PERFORMANCE_FINDING_PLACEHOLDER/${PERF_FINDING}/g" "${HTML_REPORT}"
sed -i "s/RELIABILITY_FINDING_PLACEHOLDER/${REL_FINDING}/g" "${HTML_REPORT}"
sed -i "s/SCALABILITY_FINDING_PLACEHOLDER/${SCALE_FINDING}/g" "${HTML_REPORT}"

RECOMMENDATIONS="<li>Review system resource utilization during peak load</li><li>Analyze error patterns and bottlenecks</li><li>Consider horizontal scaling if target RPS not achieved</li>"
sed -i "s|RECOMMENDATIONS_PLACEHOLDER|${RECOMMENDATIONS}|g" "${HTML_REPORT}"

log_success "HTML report generated: ${HTML_REPORT}"

# Cleanup and archival
log_info "Archiving results..."
tar -czf "${RESULTS_DIR}/load-test-archive-${TIMESTAMP}.tar.gz" \
    "${RESULTS_DIR}"/*-${TIMESTAMP}.* 2>/dev/null || true

log_success "Results archived to: ${RESULTS_DIR}/load-test-archive-${TIMESTAMP}.tar.gz"

# Final status
if [ ${FAILED_INSTANCES} -eq 0 ] && [ "${TEST_RESULT}" = "✅ PASSED" ]; then
    log_success "Load test completed successfully!"
    exit 0
else
    log_error "Load test completed with failures or did not meet targets"
    exit 1
fi

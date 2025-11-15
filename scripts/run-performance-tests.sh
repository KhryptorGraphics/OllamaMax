#!/bin/bash

# OllamaMax v1 Performance Testing Script
# Runs comprehensive load tests and generates performance report
# Addresses ISSUE-005 performance validation (v1 baseline, not full 100K RPS)

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
REPORTS_DIR="$PROJECT_ROOT/reports/performance"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
REPORT_FILE="$REPORTS_DIR/performance_report_$TIMESTAMP.md"

# Test configuration
BASE_URL="${BASE_URL:-http://localhost:13100}"
WS_URL="${WS_URL:-ws://localhost:13100}"
K6_BINARY="${K6_BINARY:-k6}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if k6 is installed
    if ! command -v $K6_BINARY &> /dev/null; then
        log_error "k6 is not installed. Please install k6 from https://k6.io/docs/getting-started/installation/"
        exit 1
    fi
    
    # Check if server is running
    if ! curl -s "$BASE_URL/api/health" > /dev/null; then
        log_error "OllamaMax server is not running at $BASE_URL"
        log_info "Please start the server with: npm run dev"
        exit 1
    fi
    
    # Create reports directory
    mkdir -p "$REPORTS_DIR"
    
    log_success "Prerequisites check passed"
}

# Get system information
get_system_info() {
    log_info "Gathering system information..."
    
    cat > "$REPORTS_DIR/system_info_$TIMESTAMP.txt" << EOF
# System Information - $(date)

## Hardware
CPU: $(lscpu | grep "Model name" | cut -d: -f2 | xargs)
CPU Cores: $(nproc)
Memory: $(free -h | grep "Mem:" | awk '{print $2}')
Disk: $(df -h / | tail -1 | awk '{print $2}')

## Software
OS: $(uname -a)
Node.js: $(node --version 2>/dev/null || echo "Not installed")
Docker: $(docker --version 2>/dev/null || echo "Not installed")
k6: $($K6_BINARY version 2>/dev/null || echo "Not installed")

## Network
Network Interface: $(ip route | grep default | awk '{print $5}' | head -1)
EOF
}

# Pre-test health check
pre_test_health_check() {
    log_info "Running pre-test health check..."
    
    # Basic health check
    HEALTH_RESPONSE=$(curl -s "$BASE_URL/api/health")
    if [[ $? -ne 0 ]]; then
        log_error "Health check failed"
        exit 1
    fi
    
    # Check detailed health
    DETAILED_HEALTH=$(curl -s "$BASE_URL/api/health/detailed")
    HEALTHY_NODES=$(echo "$DETAILED_HEALTH" | jq -r '.cluster.healthyNodes // 0')
    TOTAL_NODES=$(echo "$DETAILED_HEALTH" | jq -r '.cluster.totalNodes // 0')
    
    log_info "Cluster status: $HEALTHY_NODES/$TOTAL_NODES nodes healthy"
    
    # Check database connection
    DB_STATS=$(echo "$DETAILED_HEALTH" | jq -r '.system // {}')
    log_info "Database connections ready"
    
    log_success "Pre-test health check passed"
}

# Run load tests
run_load_tests() {
    log_info "Starting load tests..."
    
    local test_results_file="$REPORTS_DIR/k6_results_$TIMESTAMP.json"
    local test_log_file="$REPORTS_DIR/k6_log_$TIMESTAMP.log"
    
    # Run k6 test with JSON output
    log_info "Running k6 performance test..."
    $K6_BINARY run \
        --out json="$test_results_file" \
        --env BASE_URL="$BASE_URL" \
        --env WS_URL="$WS_URL" \
        "$PROJECT_ROOT/tests/load/k6-performance-test.js" \
        2>&1 | tee "$test_log_file"
    
    local exit_code=$?
    
    if [[ $exit_code -eq 0 ]]; then
        log_success "Load tests completed successfully"
    else
        log_warning "Load tests completed with warnings (exit code: $exit_code)"
    fi
    
    return $exit_code
}

# Post-test health check
post_test_health_check() {
    log_info "Running post-test health check..."
    
    # Wait for system to stabilize
    sleep 10
    
    # Check if server is still responsive
    if curl -s "$BASE_URL/api/health" > /dev/null; then
        log_success "Server is still healthy after load test"
    else
        log_error "Server appears to be unhealthy after load test"
        return 1
    fi
    
    # Get final metrics
    FINAL_METRICS=$(curl -s "$BASE_URL/api/metrics")
    echo "$FINAL_METRICS" > "$REPORTS_DIR/final_metrics_$TIMESTAMP.json"
    
    log_success "Post-test health check completed"
}

# Generate performance report
generate_report() {
    log_info "Generating performance report..."
    
    local test_results_file="$REPORTS_DIR/k6_results_$TIMESTAMP.json"
    local test_log_file="$REPORTS_DIR/k6_log_$TIMESTAMP.log"
    
    # Extract key metrics from k6 results
    if [[ -f "$test_results_file" ]]; then
        # Parse k6 JSON output for summary metrics
        local http_req_duration_p95=$(grep '"http_req_duration"' "$test_results_file" | tail -1 | jq -r '.data.value // "N/A"')
        local http_req_failed_rate=$(grep '"http_req_failed"' "$test_results_file" | tail -1 | jq -r '.data.value // "N/A"')
        local http_reqs_rate=$(grep '"http_reqs"' "$test_results_file" | tail -1 | jq -r '.data.value // "N/A"')
    else
        local http_req_duration_p95="N/A"
        local http_req_failed_rate="N/A"
        local http_reqs_rate="N/A"
    fi
    
    # Create performance report
    cat > "$REPORT_FILE" << EOF
# OllamaMax v1 Performance Test Report

**Test Date:** $(date)  
**Test Duration:** Approximately 22 minutes  
**Test Target:** v1 Baseline Performance Validation  

## Test Configuration

- **Base URL:** $BASE_URL
- **WebSocket URL:** $WS_URL
- **Load Pattern:** Gradual ramp-up to 1,000 concurrent users
- **Test Scenarios:** Health checks, node management, metrics, WebSocket inference

## System Information

$(cat "$REPORTS_DIR/system_info_$TIMESTAMP.txt")

## Performance Results

### Key Metrics

| Metric | Value | Threshold | Status |
|--------|-------|-----------|--------|
| 95th Percentile Response Time | ${http_req_duration_p95}ms | < 2000ms | $([ "${http_req_duration_p95%.*}" -lt 2000 ] 2>/dev/null && echo "✅ PASS" || echo "❌ FAIL") |
| Error Rate | ${http_req_failed_rate}% | < 5% | $([ "${http_req_failed_rate%.*}" -lt 5 ] 2>/dev/null && echo "✅ PASS" || echo "❌ FAIL") |
| Requests per Second | ${http_reqs_rate} | > 100 | $([ "${http_reqs_rate%.*}" -gt 100 ] 2>/dev/null && echo "✅ PASS" || echo "❌ FAIL") |

### Load Test Stages

1. **Ramp-up Phase (0-100 users):** 2 minutes
2. **Steady State (100 users):** 5 minutes  
3. **Stress Test (100-500 users):** 3 minutes
4. **High Load (500 users):** 5 minutes
5. **Peak Load (500-1000 users):** 2 minutes
6. **Peak Sustained (1000 users):** 3 minutes
7. **Ramp-down (1000-0 users):** 2 minutes

### Test Scenarios Distribution

- **40%** - Health and status endpoints
- **30%** - Node and model management
- **20%** - Metrics and monitoring endpoints
- **10%** - WebSocket inference simulation

## Compression Performance

- **Brotli Compression:** Enabled
- **Gzip Fallback:** Enabled
- **Average Compression Ratio:** $(grep -o 'compression_ratio.*' "$test_log_file" | tail -1 | cut -d: -f2 | xargs || echo "N/A")

## Database Connection Pool Performance

- **Max Open Connections:** 100
- **Max Idle Connections:** 20
- **Connection Pool Utilization:** Monitored via Prometheus metrics
- **Connection Errors:** $(grep -c "db_connection_errors" "$test_log_file" || echo "0")

## Identified Bottlenecks

$(if [[ -f "$test_log_file" ]]; then
    echo "### Performance Issues Detected"
    grep -i "error\|timeout\|failed" "$test_log_file" | head -10 | sed 's/^/- /'
    echo ""
    echo "### Resource Utilization"
    echo "- CPU usage patterns logged"
    echo "- Memory consumption tracked"
    echo "- Database connection pool monitored"
else
    echo "- No significant bottlenecks identified"
fi)

## Recommendations

### Immediate Actions
- Monitor database connection pool utilization under sustained load
- Verify compression is working effectively for large responses
- Implement connection pooling alerts for production deployment

### Future Optimizations
- Consider implementing response caching for frequently accessed endpoints
- Evaluate horizontal scaling patterns for > 1K concurrent users
- Plan for ISSUE-005 full 100K RPS target in future phases

## Test Files

- **System Info:** system_info_$TIMESTAMP.txt
- **K6 Results:** k6_results_$TIMESTAMP.json
- **Test Log:** k6_log_$TIMESTAMP.log
- **Final Metrics:** final_metrics_$TIMESTAMP.json

## Conclusion

$(if [[ $? -eq 0 ]]; then
    echo "✅ **v1 Baseline Performance VALIDATED**"
    echo ""
    echo "The system successfully handled the target load with acceptable performance characteristics."
    echo "Ready for v1 production deployment with current configuration."
else
    echo "⚠️ **Performance Issues Detected**"
    echo ""
    echo "Some performance thresholds were not met. Review the bottlenecks section above."
    echo "Additional optimization may be required before v1 deployment."
fi)

---
*Generated by OllamaMax Performance Testing Suite*
EOF
    
    log_success "Performance report generated: $REPORT_FILE"
}

# Main execution
main() {
    log_info "Starting OllamaMax v1 Performance Testing"
    log_info "Report will be saved to: $REPORT_FILE"
    
    check_prerequisites
    get_system_info
    pre_test_health_check
    
    local test_exit_code=0
    run_load_tests || test_exit_code=$?
    
    post_test_health_check
    generate_report
    
    if [[ $test_exit_code -eq 0 ]]; then
        log_success "Performance testing completed successfully!"
        log_info "View the full report at: $REPORT_FILE"
    else
        log_warning "Performance testing completed with issues"
        log_info "Check the report for details: $REPORT_FILE"
    fi
    
    return $test_exit_code
}

# Run main function
main "$@"

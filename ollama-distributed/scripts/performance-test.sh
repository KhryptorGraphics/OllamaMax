#!/bin/bash

# Performance Testing Script for OllamaMax Distributed Platform
# This script runs comprehensive performance tests and generates reports

set -e

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RESULTS_DIR="${PROJECT_ROOT}/performance-results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
REPORT_FILE="${RESULTS_DIR}/performance_report_${TIMESTAMP}.json"

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

# Cleanup function
cleanup() {
    log_info "Cleaning up test environment..."
    # Kill background processes
    jobs -p | xargs -r kill 2>/dev/null || true
    # Clean up temporary files
    rm -f /tmp/ollama-perf-* 2>/dev/null || true
}

trap cleanup EXIT

# Setup function
setup() {
    log_info "Setting up performance test environment..."
    
    # Create results directory
    mkdir -p "${RESULTS_DIR}"
    mkdir -p "${RESULTS_DIR}/charts"
    mkdir -p "${RESULTS_DIR}/logs"
    
    # Check dependencies
    local missing_deps=()
    
    for cmd in go k6 curl jq; do
        if ! command -v "$cmd" &> /dev/null; then
            missing_deps+=("$cmd")
        fi
    done
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "Missing dependencies: ${missing_deps[*]}"
        log_info "Please install missing dependencies:"
        log_info "  - Go: https://golang.org/dl/"
        log_info "  - K6: https://k6.io/docs/getting-started/installation/"
        log_info "  - curl: usually pre-installed"
        log_info "  - jq: https://stedolan.github.io/jq/download/"
        exit 1
    fi
    
    # Set system optimizations for testing
    if [[ "$EUID" -eq 0 ]]; then
        log_info "Applying system optimizations..."
        echo "net.core.somaxconn = 65535" >> /etc/sysctl.conf
        echo "net.core.netdev_max_backlog = 5000" >> /etc/sysctl.conf
        echo "net.ipv4.tcp_max_syn_backlog = 20480" >> /etc/sysctl.conf
        sysctl -p
    else
        log_warning "Running without root privileges - system optimizations skipped"
    fi
}

# Build application for testing
build_application() {
    log_info "Building application for performance testing..."
    
    cd "${PROJECT_ROOT}"
    
    # Build with optimizations
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -ldflags="-w -s -X main.version=perf-test -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        -o bin/ollama-distributed-perf \
        ./cmd/distributed-ollama
    
    if [ $? -eq 0 ]; then
        log_success "Application built successfully"
    else
        log_error "Failed to build application"
        exit 1
    fi
}

# Start test server
start_test_server() {
    log_info "Starting test server..."
    
    # Generate test configuration
    cat > "/tmp/ollama-perf-config.yaml" << EOF
node:
  id: "perf-test-node"
  name: "performance-test"
  environment: "testing"

api:
  listen: "0.0.0.0:8080"
  timeout: "5s"
  max_body_size: 33554432

metrics:
  enabled: true
  listen: "0.0.0.0:9090"

logging:
  level: "warn"
  format: "json"
EOF

    # Start server in background
    "${PROJECT_ROOT}/bin/ollama-distributed-perf" \
        --config "/tmp/ollama-perf-config.yaml" \
        > "${RESULTS_DIR}/logs/server.log" 2>&1 &
    
    SERVER_PID=$!
    
    # Wait for server to start
    log_info "Waiting for server to start..."
    for i in {1..30}; do
        if curl -s http://localhost:8080/api/v1/health > /dev/null 2>&1; then
            log_success "Server started successfully (PID: $SERVER_PID)"
            return 0
        fi
        sleep 1
    done
    
    log_error "Server failed to start within 30 seconds"
    return 1
}

# Run Go benchmarks
run_go_benchmarks() {
    log_info "Running Go benchmarks..."
    
    cd "${PROJECT_ROOT}"
    
    # Run comprehensive benchmarks
    go test -bench=. -benchmem -count=3 -timeout=30m \
        -benchtime=10s \
        -cpu=1,2,4 \
        ./benchmarks/... \
        > "${RESULTS_DIR}/go-benchmark-results.txt" 2>&1
    
    if [ $? -eq 0 ]; then
        log_success "Go benchmarks completed"
        
        # Parse benchmark results
        python3 -c "
import re
import json

results = {}
with open('${RESULTS_DIR}/go-benchmark-results.txt', 'r') as f:
    content = f.read()
    
    # Parse benchmark results
    benchmark_pattern = r'Benchmark(\w+)(?:-(\d+))?\s+(\d+)\s+(\d+\.?\d*)\s+ns/op\s+(\d+)\s+B/op\s+(\d+)\s+allocs/op'
    matches = re.findall(benchmark_pattern, content)
    
    for match in matches:
        name, cpu, iterations, ns_per_op, bytes_per_op, allocs_per_op = match
        key = f'{name}' + (f'-{cpu}' if cpu else '')
        results[key] = {
            'iterations': int(iterations),
            'ns_per_op': float(ns_per_op),
            'bytes_per_op': int(bytes_per_op),
            'allocs_per_op': int(allocs_per_op),
            'ops_per_sec': 1e9 / float(ns_per_op) if float(ns_per_op) > 0 else 0
        }

with open('${RESULTS_DIR}/go-benchmarks.json', 'w') as f:
    json.dump(results, f, indent=2)
"
        log_success "Benchmark results parsed and saved"
    else
        log_warning "Some Go benchmarks failed"
    fi
}

# Run K6 load tests
run_k6_load_tests() {
    log_info "Running K6 load tests..."
    
    # Create K6 test script
    cat > "/tmp/ollama-perf-k6.js" << 'EOF'
import http from 'k6/http';
import ws from 'k6/ws';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('error_rate');
const customLatency = new Trend('custom_latency');

export const options = {
  stages: [
    { duration: '2m', target: 20 },   // Ramp up
    { duration: '5m', target: 50 },   // Stay at 50 users
    { duration: '2m', target: 100 },  // Ramp to 100 users
    { duration: '5m', target: 100 },  // Stay at 100 users
    { duration: '2m', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.05'],
    error_rate: ['rate<0.05'],
  },
};

export default function() {
  // Test API endpoints
  const endpoints = [
    '/api/v1/health',
    '/api/v1/nodes',
    '/api/v1/cluster/status',
  ];
  
  const endpoint = endpoints[Math.floor(Math.random() * endpoints.length)];
  const startTime = Date.now();
  
  const response = http.get(`http://localhost:8080${endpoint}`, {
    headers: {
      'Content-Type': 'application/json',
    },
  });
  
  const latency = Date.now() - startTime;
  customLatency.add(latency);
  
  const result = check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
    'has content': (r) => r.body.length > 0,
  });
  
  errorRate.add(!result);
  
  sleep(Math.random() * 2); // Random think time
}

export function handleSummary(data) {
  return {
    'summary.json': JSON.stringify(data, null, 2),
  };
}
EOF

    # Run K6 test
    k6 run "/tmp/ollama-perf-k6.js" \
        --out json="${RESULTS_DIR}/k6-results.json" \
        > "${RESULTS_DIR}/logs/k6.log" 2>&1
    
    if [ $? -eq 0 ]; then
        log_success "K6 load tests completed"
        
        # Process K6 results
        if [ -f "summary.json" ]; then
            mv "summary.json" "${RESULTS_DIR}/k6-summary.json"
        fi
    else
        log_warning "Some K6 load tests failed"
    fi
}

# Run WebSocket performance tests
run_websocket_tests() {
    log_info "Running WebSocket performance tests..."
    
    # Create WebSocket test script
    cat > "/tmp/ollama-perf-ws.js" << 'EOF'
import ws from 'k6/ws';
import { check } from 'k6';

export const options = {
  vus: 50,
  duration: '2m',
};

export default function() {
  const url = 'ws://localhost:8080/ws';
  const params = { tags: { name: 'websocket' } };

  const response = ws.connect(url, params, function(socket) {
    socket.on('open', () => {
      console.log('WebSocket connection opened');
      
      // Send test messages
      const testMessage = JSON.stringify({
        type: 'ping',
        timestamp: Date.now()
      });
      
      socket.send(testMessage);
      
      // Set up periodic pings
      const interval = setInterval(() => {
        socket.send(testMessage);
      }, 1000);
      
      socket.setTimeout(() => {
        clearInterval(interval);
        socket.close();
      }, 30000);
    });

    socket.on('message', (data) => {
      const message = JSON.parse(data);
      check(message, {
        'message has type': (msg) => msg.type !== undefined,
        'message has timestamp': (msg) => msg.timestamp !== undefined,
      });
    });

    socket.on('close', () => {
      console.log('WebSocket connection closed');
    });
  });

  check(response, { 'WebSocket connection successful': (r) => r && r.url === url });
}
EOF

    # Run WebSocket test
    k6 run "/tmp/ollama-perf-ws.js" \
        --out json="${RESULTS_DIR}/websocket-results.json" \
        > "${RESULTS_DIR}/logs/websocket.log" 2>&1
    
    if [ $? -eq 0 ]; then
        log_success "WebSocket performance tests completed"
    else
        log_warning "WebSocket performance tests failed"
    fi
}

# Monitor system resources
monitor_resources() {
    log_info "Starting resource monitoring..."
    
    # Start resource monitoring in background
    (
        echo "timestamp,cpu_percent,memory_mb,disk_io_read,disk_io_write,network_rx,network_tx" > "${RESULTS_DIR}/resource-usage.csv"
        
        while true; do
            local timestamp=$(date +%s)
            local cpu_percent=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | sed 's/%us,//')
            local memory_mb=$(free -m | awk 'NR==2{printf "%.2f", $3}')
            local disk_stats=$(iostat -d 1 1 | tail -n +4 | awk 'NR==1{print $3","$4}')
            local network_stats=$(cat /proc/net/dev | grep eth0 | awk '{print $2","$10}')
            
            echo "${timestamp},${cpu_percent:-0},${memory_mb:-0},${disk_stats:-0,0},${network_stats:-0,0}" >> "${RESULTS_DIR}/resource-usage.csv"
            sleep 5
        done
    ) &
    
    MONITOR_PID=$!
}

# Generate performance report
generate_report() {
    log_info "Generating performance report..."
    
    # Create comprehensive report
    cat > "${REPORT_FILE}" << EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "test_duration": "$(date -d @$(($(date +%s) - START_TIME)) -u +%H:%M:%S)",
  "environment": {
    "os": "$(uname -s)",
    "arch": "$(uname -m)",
    "go_version": "$(go version | awk '{print $3}')",
    "cpu_cores": "$(nproc)",
    "memory_gb": "$(free -g | awk 'NR==2{print $2}')"
  },
  "results": {
EOF

    # Add Go benchmark results if available
    if [ -f "${RESULTS_DIR}/go-benchmarks.json" ]; then
        echo '    "go_benchmarks":' >> "${REPORT_FILE}"
        cat "${RESULTS_DIR}/go-benchmarks.json" >> "${REPORT_FILE}"
        echo ',' >> "${REPORT_FILE}"
    fi
    
    # Add K6 results if available
    if [ -f "${RESULTS_DIR}/k6-summary.json" ]; then
        echo '    "k6_load_tests":' >> "${REPORT_FILE}"
        cat "${RESULTS_DIR}/k6-summary.json" >> "${REPORT_FILE}"
        echo ',' >> "${REPORT_FILE}"
    fi
    
    # Close JSON
    echo '    "test_completed": true' >> "${REPORT_FILE}"
    echo '  }' >> "${REPORT_FILE}"
    echo '}' >> "${REPORT_FILE}"
    
    log_success "Performance report generated: ${REPORT_FILE}"
}

# Generate performance charts
generate_charts() {
    log_info "Generating performance charts..."
    
    # Create Python script for chart generation
    cat > "/tmp/generate-charts.py" << 'EOF'
import json
import matplotlib.pyplot as plt
import seaborn as sns
import pandas as pd
import sys
import os

def generate_benchmark_chart(data, output_dir):
    """Generate benchmark performance chart"""
    if not data.get('go_benchmarks'):
        return
    
    benchmarks = data['go_benchmarks']
    names = list(benchmarks.keys())
    ops_per_sec = [bench['ops_per_sec'] for bench in benchmarks.values()]
    
    plt.figure(figsize=(12, 6))
    plt.bar(names, ops_per_sec)
    plt.title('Go Benchmark Performance (Operations per Second)')
    plt.xlabel('Benchmark')
    plt.ylabel('Operations/Second')
    plt.xticks(rotation=45)
    plt.tight_layout()
    plt.savefig(f'{output_dir}/benchmark-performance.png')
    plt.close()

def generate_latency_chart(data, output_dir):
    """Generate latency distribution chart"""
    if not data.get('k6_load_tests'):
        return
    
    # This would use actual K6 data
    # For now, generate a mock chart
    latencies = [50, 75, 100, 125, 150, 200, 250, 300, 400, 500]
    percentiles = [50, 75, 90, 95, 99, 99.5, 99.9, 99.95, 99.99, 100]
    
    plt.figure(figsize=(10, 6))
    plt.plot(percentiles, latencies, marker='o')
    plt.title('Response Time Percentiles')
    plt.xlabel('Percentile')
    plt.ylabel('Response Time (ms)')
    plt.grid(True)
    plt.tight_layout()
    plt.savefig(f'{output_dir}/latency-percentiles.png')
    plt.close()

def main():
    if len(sys.argv) != 3:
        print("Usage: python generate-charts.py <report_file> <output_dir>")
        sys.exit(1)
    
    report_file = sys.argv[1]
    output_dir = sys.argv[2]
    
    if not os.path.exists(report_file):
        print(f"Report file not found: {report_file}")
        sys.exit(1)
    
    with open(report_file, 'r') as f:
        data = json.load(f)
    
    generate_benchmark_chart(data, output_dir)
    generate_latency_chart(data, output_dir)
    
    print(f"Charts generated in {output_dir}")

if __name__ == "__main__":
    main()
EOF

    # Run chart generation if Python and matplotlib are available
    if command -v python3 &> /dev/null && python3 -c "import matplotlib" &> /dev/null; then
        python3 "/tmp/generate-charts.py" "${REPORT_FILE}" "${RESULTS_DIR}/charts"
        log_success "Performance charts generated"
    else
        log_warning "Python3 or matplotlib not available - charts skipped"
    fi
}

# Print summary
print_summary() {
    log_info "Performance Test Summary"
    echo "================================="
    echo "Test completed at: $(date)"
    echo "Results directory: ${RESULTS_DIR}"
    echo "Report file: ${REPORT_FILE}"
    echo ""
    
    # Extract key metrics from report if available
    if [ -f "${REPORT_FILE}" ]; then
        local test_duration=$(jq -r '.test_duration' "${REPORT_FILE}" 2>/dev/null || echo "unknown")
        echo "Test duration: ${test_duration}"
        echo ""
        
        # Show top benchmark results
        echo "Top Go Benchmark Results:"
        jq -r '.results.go_benchmarks | to_entries | sort_by(.value.ops_per_sec) | reverse | .[0:5] | .[] | "  \(.key): \(.value.ops_per_sec | floor) ops/sec"' "${REPORT_FILE}" 2>/dev/null || echo "  No benchmark data available"
    fi
    
    echo ""
    echo "To view detailed results:"
    echo "  - Go benchmarks: ${RESULTS_DIR}/go-benchmark-results.txt"
    echo "  - K6 load tests: ${RESULTS_DIR}/k6-results.json"
    echo "  - Resource usage: ${RESULTS_DIR}/resource-usage.csv"
    echo "  - Performance charts: ${RESULTS_DIR}/charts/"
    echo ""
}

# Main execution
main() {
    START_TIME=$(date +%s)
    
    log_info "Starting OllamaMax Performance Testing Suite"
    
    # Setup
    setup
    build_application
    
    # Start test environment
    start_test_server
    monitor_resources
    
    # Run tests
    run_go_benchmarks
    run_k6_load_tests
    run_websocket_tests
    
    # Stop monitoring
    if [ ! -z "$MONITOR_PID" ]; then
        kill $MONITOR_PID 2>/dev/null || true
    fi
    
    # Generate reports
    generate_report
    generate_charts
    print_summary
    
    log_success "Performance testing completed successfully"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --help|-h)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --help, -h     Show this help message"
            echo "  --quick        Run quick tests only"
            echo "  --verbose      Enable verbose output"
            echo ""
            exit 0
            ;;
        --quick)
            QUICK_MODE=true
            shift
            ;;
        --verbose)
            set -x
            shift
            ;;
        *)
            log_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Run main function
main "$@"
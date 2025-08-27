#!/bin/bash

# Comprehensive Benchmark Runner Script
# Executes performance benchmarks and generates detailed reports

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
BENCHMARK_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$BENCHMARK_DIR")"
RESULTS_DIR="${BENCHMARK_DIR}/results"
REPORTS_DIR="${BENCHMARK_DIR}/reports"
CONFIG_FILE="${BENCHMARK_DIR}/benchmark_config.yaml"
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

log_header() {
    echo -e "${PURPLE}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${PURPLE}║${NC} $1"
    echo -e "${PURPLE}╚══════════════════════════════════════════════════════════════╝${NC}"
}

# Initialize benchmark environment
initialize_environment() {
    log_header "INITIALIZING BENCHMARK ENVIRONMENT"
    
    # Create necessary directories
    mkdir -p "$RESULTS_DIR" "$REPORTS_DIR"
    
    # Verify Go installation
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    # Verify project dependencies
    cd "$PROJECT_ROOT"
    if [[ ! -f "go.mod" ]]; then
        log_error "go.mod not found. Please run from project root."
        exit 1
    fi
    
    # Install/update dependencies
    log_info "Updating Go dependencies..."
    go mod tidy
    
    # Build benchmark tools
    log_info "Building benchmark tools..."
    go build -o "${BENCHMARK_DIR}/benchmark-runner" ./benchmarks/...
    
    log_success "Environment initialized successfully"
}

# System information collection
collect_system_info() {
    log_header "COLLECTING SYSTEM INFORMATION"
    
    local info_file="${RESULTS_DIR}/system_info_${TIMESTAMP}.txt"
    
    {
        echo "=== SYSTEM INFORMATION ==="
        echo "Date: $(date)"
        echo "Hostname: $(hostname)"
        echo "OS: $(uname -a)"
        echo "Go Version: $(go version)"
        echo "Git Commit: $(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
        echo "Git Branch: $(git branch --show-current 2>/dev/null || echo 'unknown')"
        echo ""
        
        echo "=== HARDWARE INFORMATION ==="
        echo "CPU Info:"
        if [[ -f "/proc/cpuinfo" ]]; then
            grep -E "(processor|model name|cpu cores|cache size)" /proc/cpuinfo | head -20
        elif command -v sysctl &> /dev/null; then
            sysctl -n machdep.cpu.brand_string 2>/dev/null || echo "CPU info not available"
            echo "CPU Cores: $(sysctl -n hw.ncpu 2>/dev/null || echo 'unknown')"
        fi
        echo ""
        
        echo "Memory Info:"
        if [[ -f "/proc/meminfo" ]]; then
            grep -E "(MemTotal|MemAvailable|SwapTotal)" /proc/meminfo
        elif command -v sysctl &> /dev/null; then
            echo "Memory: $(($(sysctl -n hw.memsize 2>/dev/null || echo 0) / 1024 / 1024 / 1024)) GB"
        fi
        echo ""
        
        echo "Disk Info:"
        df -h
        echo ""
        
        echo "Network Info:"
        if command -v ip &> /dev/null; then
            ip addr show | grep -E "(inet|link/ether)" | head -10
        elif command -v ifconfig &> /dev/null; then
            ifconfig | grep -E "(inet|ether)" | head -10
        fi
        echo ""
        
        echo "=== RUNTIME ENVIRONMENT ==="
        echo "Environment Variables:"
        env | grep -E "(GO|PATH|USER|HOME)" | sort
        echo ""
        
        echo "Ulimits:"
        ulimit -a
        echo ""
        
        echo "Load Average:"
        if [[ -f "/proc/loadavg" ]]; then
            cat /proc/loadavg
        elif command -v uptime &> /dev/null; then
            uptime
        fi
        
    } > "$info_file"
    
    log_success "System information collected: $info_file"
}

# Pre-benchmark validation
validate_environment() {
    log_header "VALIDATING BENCHMARK ENVIRONMENT"
    
    # Check available resources
    local available_memory
    if [[ -f "/proc/meminfo" ]]; then
        available_memory=$(grep MemAvailable /proc/meminfo | awk '{print $2}')
        available_memory=$((available_memory / 1024)) # Convert to MB
    else
        available_memory=8192 # Default assumption
    fi
    
    if [[ $available_memory -lt 2048 ]]; then
        log_warning "Low memory available: ${available_memory}MB. Benchmarks may be affected."
    fi
    
    # Check disk space
    local available_disk
    available_disk=$(df "$BENCHMARK_DIR" | awk 'NR==2 {print $4}')
    if [[ $available_disk -lt 1048576 ]]; then # Less than 1GB
        log_warning "Low disk space available. Benchmark results may be affected."
    fi
    
    # Check if other intensive processes are running
    if command -v pgrep &> /dev/null; then
        local intensive_procs
        intensive_procs=$(pgrep -f "(java|docker|kubectl|terraform)" | wc -l)
        if [[ $intensive_procs -gt 0 ]]; then
            log_warning "Found $intensive_procs intensive processes running. This may affect benchmark results."
        fi
    fi
    
    log_success "Environment validation completed"
}

# Run baseline benchmarks
run_baseline_benchmarks() {
    log_header "ESTABLISHING PERFORMANCE BASELINE"
    
    local baseline_file="${RESULTS_DIR}/baseline_${TIMESTAMP}.yaml"
    
    log_info "Running baseline consensus benchmarks..."
    go test -bench=BenchmarkConsensusOperations -benchtime=30s -count=3 \
        ./benchmarks/... > "${RESULTS_DIR}/baseline_consensus_${TIMESTAMP}.txt" 2>&1
    
    log_info "Running baseline P2P benchmarks..."
    go test -bench=BenchmarkP2PNetworking -benchtime=30s -count=3 \
        ./benchmarks/... > "${RESULTS_DIR}/baseline_p2p_${TIMESTAMP}.txt" 2>&1
    
    log_info "Running baseline API benchmarks..."
    go test -bench=BenchmarkAPIEndpoints -benchtime=30s -count=3 \
        ./benchmarks/... > "${RESULTS_DIR}/baseline_api_${TIMESTAMP}.txt" 2>&1
    
    log_info "Running baseline memory benchmarks..."
    go test -bench=BenchmarkMemoryUsage -benchtime=30s -count=3 \
        ./benchmarks/... > "${RESULTS_DIR}/baseline_memory_${TIMESTAMP}.txt" 2>&1
    
    log_success "Baseline benchmarks completed"
}

# Run comprehensive benchmarks
run_comprehensive_benchmarks() {
    log_header "RUNNING COMPREHENSIVE BENCHMARK SUITE"
    
    local comprehensive_file="${RESULTS_DIR}/comprehensive_${TIMESTAMP}.txt"
    
    log_info "Running comprehensive performance benchmarks..."
    log_info "This may take 15-30 minutes depending on system performance..."
    
    # Run with extended timeout and detailed output
    go test -bench=BenchmarkComprehensivePerformance -benchtime=5m -count=1 \
        -timeout=45m -v ./benchmarks/... > "$comprehensive_file" 2>&1 &
    
    local bench_pid=$!
    
    # Monitor benchmark progress
    while kill -0 $bench_pid 2>/dev/null; do
        log_info "Comprehensive benchmarks still running... (PID: $bench_pid)"
        sleep 30
    done
    
    wait $bench_pid
    local bench_result=$?
    
    if [[ $bench_result -eq 0 ]]; then
        log_success "Comprehensive benchmarks completed successfully"
    else
        log_error "Comprehensive benchmarks failed with exit code: $bench_result"
        return $bench_result
    fi
}

# Run scalability benchmarks
run_scalability_benchmarks() {
    log_header "RUNNING SCALABILITY BENCHMARKS"
    
    local scalability_file="${RESULTS_DIR}/scalability_${TIMESTAMP}.txt"
    
    log_info "Running scalability benchmarks across different cluster sizes..."
    
    go test -bench=BenchmarkScalability -benchtime=2m -count=1 \
        -timeout=20m ./benchmarks/... > "$scalability_file" 2>&1
    
    if [[ $? -eq 0 ]]; then
        log_success "Scalability benchmarks completed"
    else
        log_error "Scalability benchmarks failed"
        return 1
    fi
}

# Run resource utilization benchmarks
run_resource_benchmarks() {
    log_header "RUNNING RESOURCE UTILIZATION BENCHMARKS"
    
    local resource_file="${RESULTS_DIR}/resource_utilization_${TIMESTAMP}.txt"
    
    log_info "Running resource utilization benchmarks..."
    
    go test -bench=BenchmarkResourceUtilization -benchtime=1m -count=3 \
        ./benchmarks/... > "$resource_file" 2>&1
    
    if [[ $? -eq 0 ]]; then
        log_success "Resource utilization benchmarks completed"
    else
        log_error "Resource utilization benchmarks failed"
        return 1
    fi
}

# Run regression tests
run_regression_tests() {
    log_header "RUNNING REGRESSION DETECTION TESTS"
    
    local regression_file="${RESULTS_DIR}/regression_${TIMESTAMP}.txt"
    
    log_info "Running performance regression tests..."
    
    go test -run=TestPerformanceRegression -v \
        ./benchmarks/... > "$regression_file" 2>&1
    
    local regression_result=$?
    
    if [[ $regression_result -eq 0 ]]; then
        log_success "No performance regressions detected"
    else
        log_warning "Performance regressions may have been detected. Check: $regression_file"
    fi
    
    return $regression_result
}

# Generate benchmark reports
generate_reports() {
    log_header "GENERATING BENCHMARK REPORTS"
    
    local report_file="${REPORTS_DIR}/benchmark_report_${TIMESTAMP}.md"
    
    {
        echo "# Performance Benchmark Report"
        echo ""
        echo "**Generated:** $(date)"
        echo "**System:** $(hostname)"
        echo "**Git Commit:** $(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
        echo ""
        
        echo "## System Information"
        echo ""
        echo "\`\`\`"
        if [[ -f "${RESULTS_DIR}/system_info_${TIMESTAMP}.txt" ]]; then
            cat "${RESULTS_DIR}/system_info_${TIMESTAMP}.txt"
        else
            echo "System information not available"
        fi
        echo "\`\`\`"
        echo ""
        
        echo "## Benchmark Results Summary"
        echo ""
        
        # Process baseline results
        if [[ -f "${RESULTS_DIR}/baseline_consensus_${TIMESTAMP}.txt" ]]; then
            echo "### Consensus Performance"
            echo ""
            echo "\`\`\`"
            grep -E "(Benchmark|PASS|FAIL)" "${RESULTS_DIR}/baseline_consensus_${TIMESTAMP}.txt" | head -20
            echo "\`\`\`"
            echo ""
        fi
        
        # Process P2P results
        if [[ -f "${RESULTS_DIR}/baseline_p2p_${TIMESTAMP}.txt" ]]; then
            echo "### P2P Networking Performance"
            echo ""
            echo "\`\`\`"
            grep -E "(Benchmark|PASS|FAIL)" "${RESULTS_DIR}/baseline_p2p_${TIMESTAMP}.txt" | head -20
            echo "\`\`\`"
            echo ""
        fi
        
        # Process API results
        if [[ -f "${RESULTS_DIR}/baseline_api_${TIMESTAMP}.txt" ]]; then
            echo "### API Endpoint Performance"
            echo ""
            echo "\`\`\`"
            grep -E "(Benchmark|PASS|FAIL)" "${RESULTS_DIR}/baseline_api_${TIMESTAMP}.txt" | head -20
            echo "\`\`\`"
            echo ""
        fi
        
        # Process memory results
        if [[ -f "${RESULTS_DIR}/baseline_memory_${TIMESTAMP}.txt" ]]; then
            echo "### Memory Usage Analysis"
            echo ""
            echo "\`\`\`"
            grep -E "(Benchmark|PASS|FAIL)" "${RESULTS_DIR}/baseline_memory_${TIMESTAMP}.txt" | head -20
            echo "\`\`\`"
            echo ""
        fi
        
        echo "## Performance Analysis"
        echo ""
        echo "### Key Findings"
        echo ""
        
        # Extract key metrics
        local total_benchmarks=0
        local passed_benchmarks=0
        
        for result_file in "${RESULTS_DIR}"/*_${TIMESTAMP}.txt; do
            if [[ -f "$result_file" ]]; then
                local file_total=$(grep -c "Benchmark" "$result_file" 2>/dev/null || echo 0)
                local file_passed=$(grep -c "PASS" "$result_file" 2>/dev/null || echo 0)
                total_benchmarks=$((total_benchmarks + file_total))
                passed_benchmarks=$((passed_benchmarks + file_passed))
            fi
        done
        
        echo "- **Total Benchmarks:** $total_benchmarks"
        echo "- **Successful:** $passed_benchmarks"
        echo "- **Success Rate:** $(( passed_benchmarks * 100 / (total_benchmarks + 1) ))%"
        echo ""
        
        echo "### Recommendations"
        echo ""
        echo "1. Monitor memory usage patterns during peak loads"
        echo "2. Consider implementing connection pooling optimizations"
        echo "3. Evaluate consensus algorithm performance under high concurrency"
        echo "4. Optimize P2P networking for larger cluster deployments"
        echo "5. Implement automated performance monitoring in CI/CD"
        echo ""
        
        echo "## Raw Results"
        echo ""
        echo "Detailed benchmark results are available in the following files:"
        echo ""
        for result_file in "${RESULTS_DIR}"/*_${TIMESTAMP}.txt; do
            if [[ -f "$result_file" ]]; then
                echo "- $(basename "$result_file")"
            fi
        done
        
    } > "$report_file"
    
    log_success "Benchmark report generated: $report_file"
    
    # Generate summary statistics
    generate_summary_stats
}

# Generate summary statistics
generate_summary_stats() {
    local stats_file="${REPORTS_DIR}/benchmark_stats_${TIMESTAMP}.json"
    
    {
        echo "{"
        echo "  \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\","
        echo "  \"system\": {"
        echo "    \"hostname\": \"$(hostname)\","
        echo "    \"os\": \"$(uname -s)\","
        echo "    \"arch\": \"$(uname -m)\","
        echo "    \"go_version\": \"$(go version | awk '{print $3}')\""
        echo "  },"
        echo "  \"results\": {"
        echo "    \"total_benchmarks\": $(find "$RESULTS_DIR" -name "*_${TIMESTAMP}.txt" -exec grep -c "Benchmark" {} \; | awk '{sum += $1} END {print sum}'),"
        echo "    \"total_duration\": \"$(date -d@$(($(date +%s) - $(stat -c %Y "${RESULTS_DIR}/system_info_${TIMESTAMP}.txt" 2>/dev/null || echo $(date +%s))))) -u +%H:%M:%S)\""
        echo "  }"
        echo "}"
    } > "$stats_file"
    
    log_success "Summary statistics generated: $stats_file"
}

# Archive results
archive_results() {
    log_header "ARCHIVING BENCHMARK RESULTS"
    
    local archive_name="benchmark_results_${TIMESTAMP}.tar.gz"
    local archive_path="${BENCHMARK_DIR}/${archive_name}"
    
    tar -czf "$archive_path" -C "$BENCHMARK_DIR" \
        "results" "reports" 2>/dev/null
    
    if [[ $? -eq 0 ]]; then
        log_success "Results archived: $archive_path"
        log_info "Archive size: $(du -h "$archive_path" | cut -f1)"
    else
        log_warning "Failed to create archive"
    fi
}

# Cleanup old results
cleanup_old_results() {
    log_header "CLEANING UP OLD RESULTS"
    
    # Keep only last 10 result sets
    local old_results=$(find "$RESULTS_DIR" -name "*.txt" -type f -mtime +7 | head -50)
    local old_reports=$(find "$REPORTS_DIR" -name "*.md" -type f -mtime +7 | head -50)
    
    if [[ -n "$old_results" ]]; then
        echo "$old_results" | xargs rm -f
        log_info "Cleaned up old result files"
    fi
    
    if [[ -n "$old_reports" ]]; then
        echo "$old_reports" | xargs rm -f
        log_info "Cleaned up old report files"
    fi
    
    # Keep only last 5 archives
    local old_archives=$(find "$BENCHMARK_DIR" -name "benchmark_results_*.tar.gz" -type f | sort -r | tail -n +6)
    if [[ -n "$old_archives" ]]; then
        echo "$old_archives" | xargs rm -f
        log_info "Cleaned up old archives"
    fi
    
    log_success "Cleanup completed"
}

# Display final summary
display_summary() {
    log_header "BENCHMARK EXECUTION SUMMARY"
    
    echo -e "${CYAN}Benchmark Session: $TIMESTAMP${NC}"
    echo -e "${CYAN}Duration: $(date -d@$(($(date +%s) - START_TIME)) -u +%H:%M:%S)${NC}"
    echo ""
    
    echo "📊 Results Location:"
    echo "   • Results: $RESULTS_DIR"
    echo "   • Reports: $REPORTS_DIR"
    echo ""
    
    echo "📈 Key Files Generated:"
    for file in "${REPORTS_DIR}/benchmark_report_${TIMESTAMP}.md" \
                "${REPORTS_DIR}/benchmark_stats_${TIMESTAMP}.json" \
                "${BENCHMARK_DIR}/benchmark_results_${TIMESTAMP}.tar.gz"; do
        if [[ -f "$file" ]]; then
            echo "   • $(basename "$file")"
        fi
    done
    echo ""
    
    echo "🎯 Next Steps:"
    echo "   1. Review benchmark report for performance insights"
    echo "   2. Compare results with previous benchmarks"
    echo "   3. Address any performance regressions identified"
    echo "   4. Update performance baselines if improvements are confirmed"
    echo "   5. Share results with development team"
    echo ""
}

# Error handling
handle_error() {
    local exit_code=$?
    log_error "Benchmark execution failed with exit code: $exit_code"
    echo ""
    echo "🚨 Error occurred during benchmark execution"
    echo "   Check the logs above for details"
    echo "   Partial results may be available in: $RESULTS_DIR"
    echo ""
    exit $exit_code
}

# Main execution flow
main() {
    local START_TIME=$(date +%s)
    
    # Set error handling
    trap handle_error ERR
    
    log_header "COMPREHENSIVE PERFORMANCE BENCHMARK SUITE"
    echo -e "${CYAN}Session: $TIMESTAMP${NC}"
    echo -e "${CYAN}Project: ollama-distributed${NC}"
    echo ""
    
    # Execute benchmark phases
    initialize_environment
    collect_system_info
    validate_environment
    
    # Run different benchmark categories
    run_baseline_benchmarks
    run_comprehensive_benchmarks
    run_scalability_benchmarks  
    run_resource_benchmarks
    run_regression_tests
    
    # Generate reports and cleanup
    generate_reports
    archive_results
    cleanup_old_results
    
    # Display final summary
    display_summary
    
    log_success "🎉 BENCHMARK EXECUTION COMPLETED SUCCESSFULLY!"
}

# Script execution
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
#!/bin/bash
# 03-operations/health-monitoring-dashboard.sh
# Comprehensive health monitoring dashboard for Ollama Distributed Training

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# Configuration
API_BASE_URL="${API_BASE_URL:-http://127.0.0.1:8080}"
WEB_BASE_URL="${WEB_BASE_URL:-http://127.0.0.1:8081}"
UPDATE_INTERVAL="${UPDATE_INTERVAL:-5}"
LOG_FILE="${LOG_FILE:-/tmp/ollama-health-monitor.log}"

# Global variables
SCRIPT_PID=$$
MONITORING_ACTIVE=false
LAST_UPDATE=""

# Helper Functions
print_header() {
    local title="$1"
    local width=80
    local padding=$(( (width - ${#title}) / 2 ))
    
    echo -e "\n${BLUE}$(printf '═%.0s' $(seq 1 $width))${NC}"
    echo -e "${BLUE}$(printf '%*s' $padding)${WHITE}$title${BLUE}$(printf '%*s' $padding)${NC}"
    echo -e "${BLUE}$(printf '═%.0s' $(seq 1 $width))${NC}\n"
}

print_section() {
    echo -e "\n${CYAN}▶ $1${NC}"
    echo -e "${CYAN}$(printf '─%.0s' $(seq 1 40))${NC}"
}

print_status() {
    local status="$1"
    local message="$2"
    
    case "$status" in
        "healthy"|"online"|"ok"|"success")
            echo -e "  ${GREEN}✅ $message${NC}"
            ;;
        "warning"|"degraded"|"partial")
            echo -e "  ${YELLOW}⚠️  $message${NC}"
            ;;
        "error"|"offline"|"failed"|"critical")
            echo -e "  ${RED}❌ $message${NC}"
            ;;
        "info"|"unknown")
            echo -e "  ${BLUE}ℹ️  $message${NC}"
            ;;
        *)
            echo -e "  ${WHITE}• $message${NC}"
            ;;
    esac
}

log_event() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" >> "$LOG_FILE"
}

# API Helper Functions
make_api_request() {
    local endpoint="$1"
    local url="${API_BASE_URL}${endpoint}"
    
    curl -s -f -m 10 "$url" 2>/dev/null
}

check_api_endpoint() {
    local endpoint="$1"
    local description="$2"
    
    if response=$(make_api_request "$endpoint"); then
        print_status "healthy" "$description: Available"
        echo "$response"
    else
        print_status "error" "$description: Unavailable"
        echo "null"
    fi
}

# Health Check Functions
check_basic_health() {
    print_section "Basic Health Checks"
    
    # API Health Check
    if health_response=$(make_api_request "/health"); then
        if echo "$health_response" | jq -e '.status == "healthy"' > /dev/null 2>&1; then
            print_status "healthy" "API Server: Healthy"
            local uptime=$(echo "$health_response" | jq -r '.uptime // "unknown"')
            print_status "info" "Uptime: $uptime"
        else
            print_status "warning" "API Server: Running but not healthy"
        fi
    else
        print_status "error" "API Server: Not responding"
        return 1
    fi
    
    # Web Interface Check
    if curl -s -f -m 5 "$WEB_BASE_URL" > /dev/null 2>&1; then
        print_status "healthy" "Web Interface: Accessible"
    else
        print_status "warning" "Web Interface: Not accessible"
    fi
    
    return 0
}

check_cluster_status() {
    print_section "Cluster Status"
    
    # Get distributed status
    if cluster_response=$(make_api_request "/api/distributed/status"); then
        print_status "healthy" "Distributed System: Online"
        
        # Parse cluster information
        if command -v jq > /dev/null 2>&1; then
            local node_id=$(echo "$cluster_response" | jq -r '.node_id // "unknown"')
            local cluster_size=$(echo "$cluster_response" | jq -r '.cluster_size // 0')
            local connected_peers=$(echo "$cluster_response" | jq -r '.connected_peers // 0')
            local leader=$(echo "$cluster_response" | jq -r '.leader // "unknown"')
            
            print_status "info" "Node ID: $node_id"
            print_status "info" "Cluster Size: $cluster_size"
            print_status "info" "Connected Peers: $connected_peers"
            print_status "info" "Cluster Leader: $leader"
        else
            print_status "info" "Response: $cluster_response"
        fi
    else
        print_status "warning" "Distributed System: Status unavailable"
    fi
    
    # Check nodes
    if nodes_response=$(make_api_request "/api/distributed/nodes"); then
        print_status "healthy" "Node Discovery: Working"
        
        if command -v jq > /dev/null 2>&1; then
            local node_count=$(echo "$nodes_response" | jq '. | length')
            print_status "info" "Discovered Nodes: $node_count"
            
            # Show node details
            echo "$nodes_response" | jq -r '.[] | "  • Node: \(.id // "unknown") - Status: \(.status // "unknown")"' 2>/dev/null || true
        fi
    else
        print_status "warning" "Node Discovery: Unavailable"
    fi
}

check_model_status() {
    print_section "Model Management"
    
    # Check model list
    if models_response=$(make_api_request "/api/tags"); then
        print_status "healthy" "Model API: Available"
        
        if command -v jq > /dev/null 2>&1; then
            local model_count=$(echo "$models_response" | jq '.models | length' 2>/dev/null || echo "0")
            print_status "info" "Available Models: $model_count"
            
            # List models if any
            if [[ "$model_count" -gt 0 ]]; then
                echo "$models_response" | jq -r '.models[] | "  • \(.name) (\(.size))"' 2>/dev/null || true
            else
                print_status "info" "No models currently loaded"
            fi
        fi
    else
        print_status "warning" "Model API: Unavailable"
    fi
    
    # Check distributed models
    if dist_models_response=$(make_api_request "/api/distributed/models"); then
        print_status "healthy" "Distributed Models: Available"
    else
        print_status "warning" "Distributed Models: Unavailable"
    fi
}

check_performance_metrics() {
    print_section "Performance Metrics"
    
    # Get system metrics
    if metrics_response=$(make_api_request "/api/distributed/metrics"); then
        print_status "healthy" "Metrics Collection: Active"
        
        if command -v jq > /dev/null 2>&1; then
            # Parse metrics
            local cpu_usage=$(echo "$metrics_response" | jq -r '.cpu_usage // "unknown"')
            local memory_usage=$(echo "$metrics_response" | jq -r '.memory_usage // "unknown"')
            local disk_usage=$(echo "$metrics_response" | jq -r '.disk_usage // "unknown"')
            local network_rx=$(echo "$metrics_response" | jq -r '.network.rx_bytes // "unknown"')
            local network_tx=$(echo "$metrics_response" | jq -r '.network.tx_bytes // "unknown"')
            
            # Display metrics with status indicators
            if [[ "$cpu_usage" != "unknown" ]]; then
                local cpu_percent=$(echo "$cpu_usage" | sed 's/%//')
                if (( $(echo "$cpu_percent > 80" | bc -l 2>/dev/null || echo 0) )); then
                    print_status "warning" "CPU Usage: $cpu_usage"
                else
                    print_status "healthy" "CPU Usage: $cpu_usage"
                fi
            fi
            
            if [[ "$memory_usage" != "unknown" ]]; then
                print_status "info" "Memory Usage: $memory_usage"
            fi
            
            if [[ "$disk_usage" != "unknown" ]]; then
                print_status "info" "Disk Usage: $disk_usage"
            fi
            
            if [[ "$network_rx" != "unknown" && "$network_tx" != "unknown" ]]; then
                print_status "info" "Network RX/TX: $network_rx / $network_tx"
            fi
        fi
    else
        print_status "warning" "Metrics Collection: Unavailable"
    fi
    
    # Check Prometheus metrics if available
    if curl -s -f -m 5 "${API_BASE_URL}:9090/metrics" > /dev/null 2>&1; then
        print_status "healthy" "Prometheus Metrics: Available"
    else
        print_status "info" "Prometheus Metrics: Not configured"
    fi
}

check_network_connectivity() {
    print_section "Network Connectivity"
    
    # Check P2P connectivity
    local p2p_port="${P2P_PORT:-4001}"
    if netstat -ln 2>/dev/null | grep -q ":$p2p_port "; then
        print_status "healthy" "P2P Port ($p2p_port): Listening"
    else
        print_status "warning" "P2P Port ($p2p_port): Not listening"
    fi
    
    # Check API port
    local api_port=$(echo "$API_BASE_URL" | sed -n 's/.*:\([0-9]*\).*/\1/p')
    if netstat -ln 2>/dev/null | grep -q ":$api_port "; then
        print_status "healthy" "API Port ($api_port): Listening"
    else
        print_status "warning" "API Port ($api_port): Not listening"
    fi
    
    # Check Web port
    local web_port=$(echo "$WEB_BASE_URL" | sed -n 's/.*:\([0-9]*\).*/\1/p')
    if netstat -ln 2>/dev/null | grep -q ":$web_port "; then
        print_status "healthy" "Web Port ($web_port): Listening"
    else
        print_status "warning" "Web Port ($web_port): Not listening"
    fi
}

check_system_resources() {
    print_section "System Resources"
    
    # Check disk space
    local disk_usage
    disk_usage=$(df -h . | tail -1)
    local disk_percent=$(echo "$disk_usage" | awk '{print $5}' | sed 's/%//')
    
    if [[ "$disk_percent" -gt 90 ]]; then
        print_status "warning" "Disk Usage: $disk_percent% (Critical)"
    elif [[ "$disk_percent" -gt 80 ]]; then
        print_status "warning" "Disk Usage: $disk_percent% (High)"
    else
        print_status "healthy" "Disk Usage: $disk_percent%"
    fi
    
    # Check memory
    if command -v free > /dev/null 2>&1; then
        local memory_info
        memory_info=$(free -h | grep '^Mem:')
        local memory_used=$(echo "$memory_info" | awk '{print $3}')
        local memory_total=$(echo "$memory_info" | awk '{print $2}')
        print_status "info" "Memory Usage: $memory_used / $memory_total"
    fi
    
    # Check CPU load
    if [[ -f /proc/loadavg ]]; then
        local load_avg
        load_avg=$(cat /proc/loadavg | awk '{print $1}')
        local cpu_count=$(nproc 2>/dev/null || echo 1)
        local load_percent=$(echo "scale=0; $load_avg * 100 / $cpu_count" | bc 2>/dev/null || echo "unknown")
        
        if [[ "$load_percent" != "unknown" ]]; then
            if [[ "$load_percent" -gt 100 ]]; then
                print_status "warning" "CPU Load: ${load_percent}% (High)"
            else
                print_status "healthy" "CPU Load: ${load_percent}%"
            fi
        fi
    fi
}

check_log_health() {
    print_section "Log Health"
    
    # Check for recent log activity
    local log_dirs=(
        "./dev-data/logs"
        "./logs"
        "/var/log/ollama-distributed"
        "$HOME/.ollama-distributed/logs"
    )
    
    local found_logs=false
    for log_dir in "${log_dirs[@]}"; do
        if [[ -d "$log_dir" ]]; then
            local recent_logs
            recent_logs=$(find "$log_dir" -name "*.log" -mmin -10 2>/dev/null | wc -l)
            if [[ "$recent_logs" -gt 0 ]]; then
                print_status "healthy" "Recent Log Activity: $recent_logs files updated in last 10 minutes"
                found_logs=true
                break
            fi
        fi
    done
    
    if ! $found_logs; then
        print_status "info" "Log Activity: No recent log updates found"
    fi
    
    # Check for error patterns in logs
    local error_count=0
    for log_dir in "${log_dirs[@]}"; do
        if [[ -d "$log_dir" ]]; then
            error_count=$(find "$log_dir" -name "*.log" -exec grep -l "ERROR\|FATAL\|PANIC" {} \; 2>/dev/null | wc -l)
            if [[ "$error_count" -gt 0 ]]; then
                print_status "warning" "Error Logs: $error_count files contain errors"
                break
            fi
        fi
    done
    
    if [[ "$error_count" -eq 0 ]]; then
        print_status "healthy" "Error Logs: No recent errors detected"
    fi
}

# Continuous Monitoring Functions
run_continuous_monitor() {
    local interval="$1"
    
    print_header "Continuous Health Monitoring"
    echo -e "${YELLOW}Press Ctrl+C to stop monitoring${NC}\n"
    
    MONITORING_ACTIVE=true
    
    # Set up signal handlers
    trap 'MONITORING_ACTIVE=false; echo -e "\n${YELLOW}Stopping monitor...${NC}"; exit 0' INT TERM
    
    local iteration=1
    while $MONITORING_ACTIVE; do
        clear
        print_header "Ollama Distributed Health Dashboard"
        echo -e "${CYAN}Iteration: $iteration | Update Interval: ${interval}s | Last Update: $(date)${NC}"
        
        # Run all health checks
        check_basic_health
        check_cluster_status
        check_model_status
        check_performance_metrics
        check_network_connectivity
        check_system_resources
        
        # Show summary
        print_section "Summary"
        local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
        echo -e "  ${BLUE}Last Check: $timestamp${NC}"
        echo -e "  ${BLUE}Next Update: $(date -d "+${interval} seconds" '+%H:%M:%S')${NC}"
        
        # Wait for next iteration
        sleep "$interval"
        ((iteration++))
    done
}

# Alert Functions
check_critical_alerts() {
    local alerts=()
    
    # Check if API is down
    if ! make_api_request "/health" > /dev/null 2>&1; then
        alerts+=("API_DOWN")
    fi
    
    # Check disk space
    local disk_percent=$(df . | tail -1 | awk '{print $5}' | sed 's/%//')
    if [[ "$disk_percent" -gt 95 ]]; then
        alerts+=("DISK_CRITICAL")
    fi
    
    # Check memory if available
    if command -v free > /dev/null 2>&1; then
        local memory_percent=$(free | grep '^Mem:' | awk '{printf "%.0f", $3/$2 * 100}')
        if [[ "$memory_percent" -gt 95 ]]; then
            alerts+=("MEMORY_CRITICAL")
        fi
    fi
    
    # Return alerts
    if [[ ${#alerts[@]} -gt 0 ]]; then
        printf "%s\n" "${alerts[@]}"
        return 1
    fi
    return 0
}

send_alert() {
    local alert_type="$1"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    case "$alert_type" in
        "API_DOWN")
            log_event "ALERT: API server is not responding"
            ;;
        "DISK_CRITICAL")
            log_event "ALERT: Disk usage is critically high (>95%)"
            ;;
        "MEMORY_CRITICAL")
            log_event "ALERT: Memory usage is critically high (>95%)"
            ;;
        *)
            log_event "ALERT: Unknown alert type: $alert_type"
            ;;
    esac
}

# Report Generation
generate_health_report() {
    local output_file="${1:-/tmp/ollama-health-report-$(date +%Y%m%d-%H%M%S).txt}"
    
    {
        echo "Ollama Distributed Health Report"
        echo "Generated: $(date)"
        echo "========================================"
        echo
        
        # Redirect all check functions to capture output
        check_basic_health
        echo
        check_cluster_status
        echo
        check_model_status
        echo
        check_performance_metrics
        echo
        check_network_connectivity
        echo
        check_system_resources
        echo
        check_log_health
        echo
        
        echo "========================================"
        echo "Report generated by: $0"
        echo "System: $(uname -a)"
        echo "User: $(whoami)"
        echo "PWD: $(pwd)"
    } > "$output_file"
    
    echo -e "\n${GREEN}✅ Health report generated: $output_file${NC}"
}

# Usage and Help
show_usage() {
    cat << EOF
Ollama Distributed Health Monitoring Dashboard

Usage: $0 [command] [options]

Commands:
    check           Run one-time health check (default)
    monitor         Run continuous monitoring
    report          Generate health report
    alerts          Check for critical alerts
    help            Show this help

Options:
    --api-url URL       API base URL (default: http://127.0.0.1:8080)
    --web-url URL       Web base URL (default: http://127.0.0.1:8081)
    --interval N        Update interval in seconds for monitoring (default: 5)
    --output FILE       Output file for reports
    --log-file FILE     Log file location (default: /tmp/ollama-health-monitor.log)

Examples:
    $0                                    # Run one-time health check
    $0 check                              # Same as above
    $0 monitor                            # Start continuous monitoring
    $0 monitor --interval 10              # Monitor with 10-second updates
    $0 report --output health.txt         # Generate report to file
    $0 alerts                             # Check for critical alerts
    $0 --api-url http://server:8080 check # Check remote server

Environment Variables:
    API_BASE_URL        Default API URL
    WEB_BASE_URL        Default Web URL
    UPDATE_INTERVAL     Default monitoring interval
    LOG_FILE           Default log file location

EOF
}

# Main execution
main() {
    local command="${1:-check}"
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --api-url)
                API_BASE_URL="$2"
                shift 2
                ;;
            --web-url)
                WEB_BASE_URL="$2"
                shift 2
                ;;
            --interval)
                UPDATE_INTERVAL="$2"
                shift 2
                ;;
            --output)
                OUTPUT_FILE="$2"
                shift 2
                ;;
            --log-file)
                LOG_FILE="$2"
                shift 2
                ;;
            check|monitor|report|alerts|help)
                command="$1"
                shift
                ;;
            *)
                shift
                ;;
        esac
    done
    
    # Execute command
    case "$command" in
        "check")
            print_header "Ollama Distributed Health Check"
            check_basic_health
            check_cluster_status
            check_model_status
            check_performance_metrics
            check_network_connectivity
            check_system_resources
            check_log_health
            ;;
        "monitor")
            run_continuous_monitor "$UPDATE_INTERVAL"
            ;;
        "report")
            generate_health_report "$OUTPUT_FILE"
            ;;
        "alerts")
            print_header "Critical Alerts Check"
            if alerts=$(check_critical_alerts); then
                print_status "healthy" "No critical alerts"
            else
                print_status "error" "Critical alerts detected:"
                echo "$alerts" | while read -r alert; do
                    print_status "error" "Alert: $alert"
                    send_alert "$alert"
                done
            fi
            ;;
        "help")
            show_usage
            ;;
        *)
            echo -e "${RED}Unknown command: $command${NC}"
            show_usage
            exit 1
            ;;
    esac
}

# Initialize logging
mkdir -p "$(dirname "$LOG_FILE")"
log_event "Health monitoring script started with command: ${1:-check}"

# Run main function
main "$@"
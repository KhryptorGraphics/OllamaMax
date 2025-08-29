#!/bin/bash
# Comprehensive Health Check Script for OllamaMax Production System
# Validates all components and generates detailed health report

set -euo pipefail

# Configuration
NAMESPACE="${NAMESPACE:-ollama-system}"
DB_NAMESPACE="${DB_NAMESPACE:-database}"
MONITORING_NAMESPACE="${MONITORING_NAMESPACE:-monitoring}"
TIMEOUT="${TIMEOUT:-60}"
OUTPUT_FORMAT="${OUTPUT_FORMAT:-console}"  # console, json, prometheus
VERBOSE="${VERBOSE:-false}"

# Colors for console output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Health check results
declare -A HEALTH_RESULTS
OVERALL_HEALTH="healthy"

# Logging functions
log() {
    if [[ "$OUTPUT_FORMAT" == "console" ]]; then
        echo -e "${BLUE}[$(date +'%H:%M:%S')] INFO:${NC} $1" >&2
    fi
}

success() {
    if [[ "$OUTPUT_FORMAT" == "console" ]]; then
        echo -e "${GREEN}[$(date +'%H:%M:%S')] SUCCESS:${NC} $1" >&2
    fi
}

warning() {
    if [[ "$OUTPUT_FORMAT" == "console" ]]; then
        echo -e "${YELLOW}[$(date +'%H:%M:%S')] WARNING:${NC} $1" >&2
    fi
    if [[ "$OVERALL_HEALTH" == "healthy" ]]; then
        OVERALL_HEALTH="warning"
    fi
}

error() {
    if [[ "$OUTPUT_FORMAT" == "console" ]]; then
        echo -e "${RED}[$(date +'%H:%M:%S')] ERROR:${NC} $1" >&2
    fi
    OVERALL_HEALTH="unhealthy"
}

# Record health check result
record_result() {
    local component=$1
    local status=$2
    local message=$3
    local response_time=${4:-"N/A"}
    
    HEALTH_RESULTS["$component"]="$status,$message,$response_time"
    
    case $status in
        "healthy") success "$component: $message" ;;
        "warning") warning "$component: $message" ;;
        "unhealthy") error "$component: $message" ;;
    esac
}

# Check Kubernetes cluster connectivity
check_kubernetes() {
    log "Checking Kubernetes cluster connectivity..."
    
    local start_time=$(date +%s%3N)
    if kubectl cluster-info &>/dev/null; then
        local end_time=$(date +%s%3N)
        local response_time=$((end_time - start_time))
        record_result "kubernetes" "healthy" "Cluster accessible" "${response_time}ms"
        return 0
    else
        record_result "kubernetes" "unhealthy" "Cannot connect to cluster" "0"
        return 1
    fi
}

# Check namespace existence
check_namespaces() {
    log "Checking required namespaces..."
    
    local namespaces=("$NAMESPACE" "$DB_NAMESPACE" "$MONITORING_NAMESPACE")
    local all_healthy=true
    
    for ns in "${namespaces[@]}"; do
        if kubectl get namespace "$ns" &>/dev/null; then
            record_result "namespace-$ns" "healthy" "Namespace exists" "N/A"
        else
            record_result "namespace-$ns" "unhealthy" "Namespace missing" "N/A"
            all_healthy=false
        fi
    done
    
    return $([[ "$all_healthy" == "true" ]] && echo 0 || echo 1)
}

# Check application pods
check_application_pods() {
    log "Checking application pods..."
    
    local pod_count=$(kubectl get pods -n "$NAMESPACE" -l app.kubernetes.io/name=ollama-distributed --no-headers | wc -l)
    local ready_count=$(kubectl get pods -n "$NAMESPACE" -l app.kubernetes.io/name=ollama-distributed --no-headers | grep -c "Running" || true)
    
    if [[ $pod_count -eq 0 ]]; then
        record_result "app-pods" "unhealthy" "No application pods found" "N/A"
        return 1
    elif [[ $ready_count -eq $pod_count ]]; then
        record_result "app-pods" "healthy" "$ready_count/$pod_count pods running" "N/A"
        return 0
    else
        record_result "app-pods" "warning" "$ready_count/$pod_count pods running" "N/A"
        return 1
    fi
}

# Check database connectivity
check_database() {
    log "Checking database connectivity..."
    
    local start_time=$(date +%s%3N)
    if kubectl exec -n "$DB_NAMESPACE" deployment/postgres-primary -- pg_isready -U ollamamax &>/dev/null; then
        local end_time=$(date +%s%3N)
        local response_time=$((end_time - start_time))
        record_result "database" "healthy" "PostgreSQL accessible" "${response_time}ms"
        
        # Check database size and connections
        local db_info=$(kubectl exec -n "$DB_NAMESPACE" deployment/postgres-primary -- psql -U ollamamax -d ollamamax -t -c "SELECT count(*) FROM pg_stat_activity WHERE state = 'active';" 2>/dev/null | tr -d ' \n' || echo "unknown")
        record_result "database-connections" "healthy" "$db_info active connections" "N/A"
        
        return 0
    else
        record_result "database" "unhealthy" "PostgreSQL not accessible" "0"
        return 1
    fi
}

# Check Redis connectivity
check_redis() {
    log "Checking Redis connectivity..."
    
    local start_time=$(date +%s%3N)
    if kubectl exec -n "$DB_NAMESPACE" deployment/redis-master -- redis-cli ping | grep -q PONG; then
        local end_time=$(date +%s%3N)
        local response_time=$((end_time - start_time))
        record_result "redis" "healthy" "Redis accessible" "${response_time}ms"
        
        # Check Redis memory usage
        local memory_usage=$(kubectl exec -n "$DB_NAMESPACE" deployment/redis-master -- redis-cli info memory | grep "used_memory_human:" | cut -d: -f2 | tr -d '\r' || echo "unknown")
        record_result "redis-memory" "healthy" "Memory usage: $memory_usage" "N/A"
        
        return 0
    else
        record_result "redis" "unhealthy" "Redis not accessible" "0"
        return 1
    fi
}

# Check API endpoints
check_api_endpoints() {
    log "Checking API endpoints..."
    
    # Port forward to the service
    local port_forward_pid
    kubectl port-forward service/ollama-api 8080:8080 -n "$NAMESPACE" &>/dev/null &
    port_forward_pid=$!
    
    # Wait for port forward to be ready
    sleep 3
    
    local endpoints=("/health" "/ready" "/api/v1/models" "/metrics")
    local all_healthy=true
    
    for endpoint in "${endpoints[@]}"; do
        local start_time=$(date +%s%3N)
        if curl -f -s -m 10 "http://localhost:8080$endpoint" &>/dev/null; then
            local end_time=$(date +%s%3N)
            local response_time=$((end_time - start_time))
            record_result "api$endpoint" "healthy" "Endpoint accessible" "${response_time}ms"
        else
            record_result "api$endpoint" "unhealthy" "Endpoint not accessible" "0"
            all_healthy=false
        fi
    done
    
    # Clean up port forward
    kill $port_forward_pid 2>/dev/null || true
    
    return $([[ "$all_healthy" == "true" ]] && echo 0 || echo 1)
}

# Check monitoring services
check_monitoring() {
    log "Checking monitoring services..."
    
    # Check Prometheus
    if kubectl get pods -n "$MONITORING_NAMESPACE" -l app=prometheus --no-headers | grep -q "Running"; then
        record_result "prometheus" "healthy" "Prometheus running" "N/A"
    else
        record_result "prometheus" "unhealthy" "Prometheus not running" "N/A"
    fi
    
    # Check Grafana
    if kubectl get pods -n "$MONITORING_NAMESPACE" -l app=grafana --no-headers | grep -q "Running"; then
        record_result "grafana" "healthy" "Grafana running" "N/A"
    else
        record_result "grafana" "unhealthy" "Grafana not running" "N/A"
    fi
    
    # Check AlertManager
    if kubectl get pods -n "$MONITORING_NAMESPACE" -l app=alertmanager --no-headers | grep -q "Running"; then
        record_result "alertmanager" "healthy" "AlertManager running" "N/A"
    else
        record_result "alertmanager" "warning" "AlertManager not running" "N/A"
    fi
    
    # Check Jaeger
    if kubectl get pods -n "$MONITORING_NAMESPACE" -l app=jaeger --no-headers | grep -q "Running"; then
        record_result "jaeger" "healthy" "Jaeger running" "N/A"
    else
        record_result "jaeger" "warning" "Jaeger not running" "N/A"
    fi
}

# Check resource usage
check_resource_usage() {
    log "Checking resource usage..."
    
    # Get node resource usage
    local node_usage=$(kubectl top nodes --no-headers 2>/dev/null || echo "metrics-server-unavailable")
    
    if [[ "$node_usage" != "metrics-server-unavailable" ]]; then
        local cpu_usage=$(echo "$node_usage" | awk '{sum+=$3; count++} END {if(count>0) print sum/count; else print 0}' | sed 's/%//')
        local memory_usage=$(echo "$node_usage" | awk '{sum+=$5; count++} END {if(count>0) print sum/count; else print 0}' | sed 's/%//')
        
        # Check CPU usage
        if (( $(echo "$cpu_usage > 80" | bc -l 2>/dev/null || echo 0) )); then
            record_result "cpu-usage" "warning" "High CPU usage: ${cpu_usage}%" "N/A"
        elif (( $(echo "$cpu_usage > 95" | bc -l 2>/dev/null || echo 0) )); then
            record_result "cpu-usage" "unhealthy" "Critical CPU usage: ${cpu_usage}%" "N/A"
        else
            record_result "cpu-usage" "healthy" "CPU usage: ${cpu_usage}%" "N/A"
        fi
        
        # Check memory usage
        if (( $(echo "$memory_usage > 80" | bc -l 2>/dev/null || echo 0) )); then
            record_result "memory-usage" "warning" "High memory usage: ${memory_usage}%" "N/A"
        elif (( $(echo "$memory_usage > 95" | bc -l 2>/dev/null || echo 0) )); then
            record_result "memory-usage" "unhealthy" "Critical memory usage: ${memory_usage}%" "N/A"
        else
            record_result "memory-usage" "healthy" "Memory usage: ${memory_usage}%" "N/A"
        fi
    else
        record_result "resource-metrics" "warning" "Metrics server unavailable" "N/A"
    fi
}

# Check persistent volumes
check_storage() {
    log "Checking persistent volumes..."
    
    local pvs_status=$(kubectl get pv --no-headers | grep -c "Bound" || echo 0)
    local pvs_total=$(kubectl get pv --no-headers | wc -l || echo 0)
    
    if [[ $pvs_total -eq 0 ]]; then
        record_result "storage" "warning" "No persistent volumes found" "N/A"
    elif [[ $pvs_status -eq $pvs_total ]]; then
        record_result "storage" "healthy" "$pvs_status/$pvs_total PVs bound" "N/A"
    else
        record_result "storage" "warning" "$pvs_status/$pvs_total PVs bound" "N/A"
    fi
    
    # Check PVC status in application namespace
    local pvc_status=$(kubectl get pvc -n "$NAMESPACE" --no-headers | grep -c "Bound" || echo 0)
    local pvc_total=$(kubectl get pvc -n "$NAMESPACE" --no-headers | wc -l || echo 0)
    
    if [[ $pvc_total -gt 0 ]]; then
        if [[ $pvc_status -eq $pvc_total ]]; then
            record_result "app-storage" "healthy" "$pvc_status/$pvc_total PVCs bound" "N/A"
        else
            record_result "app-storage" "warning" "$pvc_status/$pvc_total PVCs bound" "N/A"
        fi
    fi
}

# Check ingress and external connectivity
check_ingress() {
    log "Checking ingress and external connectivity..."
    
    # Check ingress controller
    if kubectl get pods -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx --no-headers | grep -q "Running"; then
        record_result "ingress-controller" "healthy" "NGINX ingress controller running" "N/A"
    else
        record_result "ingress-controller" "unhealthy" "NGINX ingress controller not running" "N/A"
    fi
    
    # Check ingress resources
    local ingress_count=$(kubectl get ingress -n "$NAMESPACE" --no-headers | wc -l || echo 0)
    if [[ $ingress_count -gt 0 ]]; then
        record_result "ingress-resources" "healthy" "$ingress_count ingress resources found" "N/A"
    else
        record_result "ingress-resources" "warning" "No ingress resources found" "N/A"
    fi
    
    # Check external DNS resolution (if domain is configured)
    local ingress_hosts=$(kubectl get ingress -n "$NAMESPACE" -o jsonpath='{.items[*].spec.rules[*].host}' 2>/dev/null || echo "")
    if [[ -n "$ingress_hosts" ]]; then
        for host in $ingress_hosts; do
            if nslookup "$host" &>/dev/null; then
                record_result "dns-$host" "healthy" "DNS resolution working" "N/A"
            else
                record_result "dns-$host" "warning" "DNS resolution failed" "N/A"
            fi
        done
    fi
}

# Check certificates
check_certificates() {
    log "Checking TLS certificates..."
    
    # Check cert-manager
    if kubectl get pods -n cert-manager -l app=cert-manager --no-headers | grep -q "Running"; then
        record_result "cert-manager" "healthy" "Cert-manager running" "N/A"
    else
        record_result "cert-manager" "warning" "Cert-manager not running" "N/A"
    fi
    
    # Check certificate resources
    local cert_count=$(kubectl get certificates -A --no-headers | wc -l || echo 0)
    if [[ $cert_count -gt 0 ]]; then
        local ready_certs=$(kubectl get certificates -A --no-headers | grep -c "True" || echo 0)
        if [[ $ready_certs -eq $cert_count ]]; then
            record_result "certificates" "healthy" "$ready_certs/$cert_count certificates ready" "N/A"
        else
            record_result "certificates" "warning" "$ready_certs/$cert_count certificates ready" "N/A"
        fi
    else
        record_result "certificates" "warning" "No certificates found" "N/A"
    fi
}

# Generate JSON output
generate_json_output() {
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    echo "{"
    echo "  \"timestamp\": \"$timestamp\","
    echo "  \"overall_health\": \"$OVERALL_HEALTH\","
    echo "  \"checks\": {"
    
    local first=true
    for component in "${!HEALTH_RESULTS[@]}"; do
        if [[ "$first" == "false" ]]; then
            echo ","
        fi
        first=false
        
        IFS=',' read -r status message response_time <<< "${HEALTH_RESULTS[$component]}"
        echo -n "    \"$component\": {"
        echo -n "\"status\": \"$status\", "
        echo -n "\"message\": \"$message\", "
        echo -n "\"response_time\": \"$response_time\""
        echo -n "}"
    done
    
    echo ""
    echo "  }"
    echo "}"
}

# Generate Prometheus metrics output
generate_prometheus_output() {
    echo "# HELP ollama_health_check_status Health check status (1=healthy, 0.5=warning, 0=unhealthy)"
    echo "# TYPE ollama_health_check_status gauge"
    
    for component in "${!HEALTH_RESULTS[@]}"; do
        IFS=',' read -r status message response_time <<< "${HEALTH_RESULTS[$component]}"
        
        local value
        case $status in
            "healthy") value="1" ;;
            "warning") value="0.5" ;;
            "unhealthy") value="0" ;;
            *) value="0" ;;
        esac
        
        echo "ollama_health_check_status{component=\"$component\",status=\"$status\"} $value"
    done
    
    echo "# HELP ollama_health_check_response_time_ms Response time in milliseconds"
    echo "# TYPE ollama_health_check_response_time_ms gauge"
    
    for component in "${!HEALTH_RESULTS[@]}"; do
        IFS=',' read -r status message response_time <<< "${HEALTH_RESULTS[$component]}"
        
        if [[ "$response_time" != "N/A" && "$response_time" =~ ^[0-9]+ms$ ]]; then
            local time_value=${response_time%ms}
            echo "ollama_health_check_response_time_ms{component=\"$component\"} $time_value"
        fi
    done
    
    echo "# HELP ollama_overall_health_status Overall system health (1=healthy, 0.5=warning, 0=unhealthy)"
    echo "# TYPE ollama_overall_health_status gauge"
    local overall_value
    case $OVERALL_HEALTH in
        "healthy") overall_value="1" ;;
        "warning") overall_value="0.5" ;;
        "unhealthy") overall_value="0" ;;
        *) overall_value="0" ;;
    esac
    echo "ollama_overall_health_status $overall_value"
}

# Generate console summary
generate_console_summary() {
    echo
    echo "======================================"
    echo "      OllamaMax Health Check Summary  "
    echo "======================================"
    echo
    
    local healthy_count=0
    local warning_count=0
    local unhealthy_count=0
    
    for component in "${!HEALTH_RESULTS[@]}"; do
        IFS=',' read -r status message response_time <<< "${HEALTH_RESULTS[$component]}"
        case $status in
            "healthy") ((healthy_count++)) ;;
            "warning") ((warning_count++)) ;;
            "unhealthy") ((unhealthy_count++)) ;;
        esac
    done
    
    local total_checks=$((healthy_count + warning_count + unhealthy_count))
    
    echo -e "Overall Status: $(case $OVERALL_HEALTH in
        'healthy') echo -e "${GREEN}HEALTHY${NC}" ;;
        'warning') echo -e "${YELLOW}WARNING${NC}" ;;
        'unhealthy') echo -e "${RED}UNHEALTHY${NC}" ;;
    esac)"
    echo
    echo -e "✅ Healthy:   ${GREEN}$healthy_count${NC}/$total_checks"
    echo -e "⚠️  Warning:   ${YELLOW}$warning_count${NC}/$total_checks"
    echo -e "❌ Unhealthy: ${RED}$unhealthy_count${NC}/$total_checks"
    echo
    
    if [[ "$VERBOSE" == "true" ]]; then
        echo "Detailed Results:"
        echo "----------------"
        for component in $(printf '%s\n' "${!HEALTH_RESULTS[@]}" | sort); do
            IFS=',' read -r status message response_time <<< "${HEALTH_RESULTS[$component]}"
            local icon
            case $status in
                "healthy") icon="✅" ;;
                "warning") icon="⚠️ " ;;
                "unhealthy") icon="❌" ;;
            esac
            printf "%-30s %s %s" "$component" "$icon" "$message"
            if [[ "$response_time" != "N/A" ]]; then
                printf " (%s)" "$response_time"
            fi
            echo
        done
    fi
    
    echo "======================================"
}

# Main health check function
main() {
    log "Starting comprehensive health check..."
    
    # Run all health checks
    check_kubernetes || true
    check_namespaces || true
    check_application_pods || true
    check_database || true
    check_redis || true
    check_api_endpoints || true
    check_monitoring || true
    check_resource_usage || true
    check_storage || true
    check_ingress || true
    check_certificates || true
    
    # Generate output based on format
    case $OUTPUT_FORMAT in
        "json")
            generate_json_output
            ;;
        "prometheus")
            generate_prometheus_output
            ;;
        "console")
            generate_console_summary
            ;;
        *)
            error "Unknown output format: $OUTPUT_FORMAT"
            exit 1
            ;;
    esac
    
    # Exit with appropriate code
    case $OVERALL_HEALTH in
        "healthy") exit 0 ;;
        "warning") exit 1 ;;
        "unhealthy") exit 2 ;;
        *) exit 3 ;;
    esac
}

# Show help
show_help() {
    cat << EOF
OllamaMax Health Check Script

Usage: $0 [OPTIONS]

OPTIONS:
    -n, --namespace NAMESPACE        Application namespace (default: ollama-system)
    -d, --db-namespace NAMESPACE     Database namespace (default: database)
    -m, --monitoring-namespace NS    Monitoring namespace (default: monitoring)
    -t, --timeout SECONDS           Timeout for individual checks (default: 60)
    -o, --output FORMAT             Output format: console|json|prometheus (default: console)
    -v, --verbose                   Show detailed results
    -h, --help                      Show this help message

EXAMPLES:
    # Basic health check
    $0

    # JSON output for automation
    $0 --output json

    # Prometheus metrics format
    $0 --output prometheus

    # Verbose console output
    $0 --verbose

EXIT CODES:
    0 - All checks healthy
    1 - Some checks have warnings
    2 - Some checks are unhealthy
    3 - Script error
EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -n|--namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        -d|--db-namespace)
            DB_NAMESPACE="$2"
            shift 2
            ;;
        -m|--monitoring-namespace)
            MONITORING_NAMESPACE="$2"
            shift 2
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_FORMAT="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE="true"
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Run main function
main
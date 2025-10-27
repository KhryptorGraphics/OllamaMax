#!/bin/bash

# Validate Monitoring Stack Script
# Comprehensive validation of all OllamaMax monitoring components

set -euo pipefail

# Color codes for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

log_error() {
    echo -e "${RED}[✗]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_section() {
    echo -e "\n${CYAN}=== $1 ===${NC}\n"
}

# Test result tracking
CHECKS_PASSED=0
CHECKS_FAILED=0
FAILED_CHECKS=()

# Configuration with defaults
PROMETHEUS_URL="${PROMETHEUS_URL:-http://localhost:9090}"
GRAFANA_URL="${GRAFANA_URL:-http://localhost:3001}"
ALERTMANAGER_URL="${ALERTMANAGER_URL:-http://localhost:9093}"
JAEGER_URL="${JAEGER_URL:-http://localhost:16686}"
ELASTICSEARCH_URL="${ELASTICSEARCH_URL:-http://localhost:9200}"
LOGSTASH_URL="${LOGSTASH_URL:-http://localhost:9600}"
KIBANA_URL="${KIBANA_URL:-http://localhost:5601}"

# Timeout for health checks
HEALTH_CHECK_TIMEOUT=5
METRICS_QUERY_TIMEOUT=10

# Function to check service health
check_service_health() {
    local service_name=$1
    local health_url=$2
    local expected_response=${3:-200}

    log_info "Checking $service_name health..."

    local response=$(curl -s -o /dev/null -w "%{http_code}" \
        --max-time "$HEALTH_CHECK_TIMEOUT" \
        "$health_url" 2>&1)

    if [ "$response" = "$expected_response" ]; then
        log_success "$service_name is healthy (HTTP $response)"
        ((CHECKS_PASSED++))
        return 0
    else
        log_error "$service_name health check failed (HTTP $response)"
        FAILED_CHECKS+=("$service_name health")
        ((CHECKS_FAILED++))
        return 1
    fi
}

# Function to check Prometheus health
check_prometheus() {
    log_section "Prometheus Health Check"

    # Health endpoint
    if ! check_service_health "Prometheus" "$PROMETHEUS_URL/-/healthy"; then
        return 1
    fi

    # Ready endpoint
    log_info "Checking Prometheus ready status..."
    local ready=$(curl -s -o /dev/null -w "%{http_code}" \
        --max-time "$HEALTH_CHECK_TIMEOUT" \
        "$PROMETHEUS_URL/-/ready" 2>&1)

    if [ "$ready" = "200" ]; then
        log_success "Prometheus is ready"
        ((CHECKS_PASSED++))
    else
        log_error "Prometheus not ready (HTTP $ready)"
        FAILED_CHECKS+=("Prometheus ready")
        ((CHECKS_FAILED++))
    fi

    return 0
}

# Function to check Grafana health
check_grafana() {
    log_section "Grafana Health Check"

    if ! check_service_health "Grafana" "$GRAFANA_URL/api/health"; then
        return 1
    fi

    # Check Grafana metrics
    log_info "Checking Grafana metrics endpoint..."
    local metrics=$(curl -s -o /dev/null -w "%{http_code}" \
        --max-time "$HEALTH_CHECK_TIMEOUT" \
        "$GRAFANA_URL/metrics" 2>&1)

    if [ "$metrics" = "200" ]; then
        log_success "Grafana metrics endpoint accessible"
        ((CHECKS_PASSED++))
    else
        log_warning "Grafana metrics endpoint not accessible (HTTP $metrics)"
    fi

    return 0
}

# Function to check Alertmanager health
check_alertmanager() {
    log_section "Alertmanager Health Check"

    if ! check_service_health "Alertmanager" "$ALERTMANAGER_URL/-/healthy"; then
        return 1
    fi

    # Check Alertmanager ready
    log_info "Checking Alertmanager ready status..."
    local ready=$(curl -s -o /dev/null -w "%{http_code}" \
        --max-time "$HEALTH_CHECK_TIMEOUT" \
        "$ALERTMANAGER_URL/-/ready" 2>&1)

    if [ "$ready" = "200" ]; then
        log_success "Alertmanager is ready"
        ((CHECKS_PASSED++))
    else
        log_error "Alertmanager not ready (HTTP $ready)"
        FAILED_CHECKS+=("Alertmanager ready")
        ((CHECKS_FAILED++))
    fi

    # Check active alerts
    log_info "Checking active alerts..."
    local alerts_response=$(curl -s --max-time "$HEALTH_CHECK_TIMEOUT" \
        "$ALERTMANAGER_URL/api/v2/alerts" 2>&1)

    if [ $? -eq 0 ]; then
        local alert_count=$(echo "$alerts_response" | jq '. | length' 2>/dev/null || echo "0")
        log_success "Retrieved active alerts (count: $alert_count)"
        ((CHECKS_PASSED++))
    else
        log_error "Failed to retrieve active alerts"
        FAILED_CHECKS+=("Alertmanager alerts")
        ((CHECKS_FAILED++))
    fi

    return 0
}

# Function to check Jaeger health
check_jaeger() {
    log_section "Jaeger Health Check"

    if ! check_service_health "Jaeger" "$JAEGER_URL/"; then
        return 1
    fi

    return 0
}

# Function to check Elasticsearch health
check_elasticsearch() {
    log_section "Elasticsearch Health Check"

    # Cluster health
    log_info "Checking Elasticsearch cluster health..."
    local health_response=$(curl -s --max-time "$HEALTH_CHECK_TIMEOUT" \
        "$ELASTICSEARCH_URL/_cluster/health" 2>&1)

    if [ $? -eq 0 ]; then
        local cluster_status=$(echo "$health_response" | jq -r '.status' 2>/dev/null || echo "unknown")
        local node_count=$(echo "$health_response" | jq -r '.number_of_nodes' 2>/dev/null || echo "0")

        if [ "$cluster_status" = "green" ] || [ "$cluster_status" = "yellow" ]; then
            log_success "Elasticsearch cluster is $cluster_status (nodes: $node_count)"
            ((CHECKS_PASSED++))
        else
            log_error "Elasticsearch cluster health is $cluster_status"
            FAILED_CHECKS+=("Elasticsearch cluster health")
            ((CHECKS_FAILED++))
        fi
    else
        log_error "Failed to retrieve Elasticsearch cluster health"
        FAILED_CHECKS+=("Elasticsearch health")
        ((CHECKS_FAILED++))
        return 1
    fi

    # Check node stats
    log_info "Checking Elasticsearch node stats..."
    local stats=$(curl -s -o /dev/null -w "%{http_code}" \
        --max-time "$HEALTH_CHECK_TIMEOUT" \
        "$ELASTICSEARCH_URL/_nodes/stats" 2>&1)

    if [ "$stats" = "200" ]; then
        log_success "Elasticsearch node stats accessible"
        ((CHECKS_PASSED++))
    else
        log_warning "Elasticsearch node stats not accessible (HTTP $stats)"
    fi

    return 0
}

# Function to check Logstash health
check_logstash() {
    log_section "Logstash Health Check"

    # Node stats
    log_info "Checking Logstash node stats..."
    local stats_response=$(curl -s --max-time "$HEALTH_CHECK_TIMEOUT" \
        "$LOGSTASH_URL/_node/stats" 2>&1)

    if [ $? -eq 0 ]; then
        local pipeline_status=$(echo "$stats_response" | jq -r '.pipelines | keys[]' 2>/dev/null | head -1)
        if [ -n "$pipeline_status" ]; then
            log_success "Logstash is running (pipeline: ${pipeline_status:-default})"
            ((CHECKS_PASSED++))
        else
            log_warning "Logstash is running but no pipelines detected"
            ((CHECKS_PASSED++))
        fi
    else
        log_error "Failed to retrieve Logstash stats"
        FAILED_CHECKS+=("Logstash health")
        ((CHECKS_FAILED++))
        return 1
    fi

    # Node info
    log_info "Checking Logstash node info..."
    local info=$(curl -s -o /dev/null -w "%{http_code}" \
        --max-time "$HEALTH_CHECK_TIMEOUT" \
        "$LOGSTASH_URL/_node" 2>&1)

    if [ "$info" = "200" ]; then
        log_success "Logstash node info accessible"
        ((CHECKS_PASSED++))
    else
        log_warning "Logstash node info not accessible (HTTP $info)"
    fi

    return 0
}

# Function to check Kibana health
check_kibana() {
    log_section "Kibana Health Check"

    if ! check_service_health "Kibana" "$KIBANA_URL/api/status"; then
        return 1
    fi

    # Check detailed status
    log_info "Checking Kibana detailed status..."
    local status_response=$(curl -s --max-time "$HEALTH_CHECK_TIMEOUT" \
        "$KIBANA_URL/api/status" 2>&1)

    if [ $? -eq 0 ]; then
        local overall_status=$(echo "$status_response" | jq -r '.status.overall.state' 2>/dev/null || echo "unknown")
        if [ "$overall_status" = "green" ]; then
            log_success "Kibana overall status is $overall_status"
            ((CHECKS_PASSED++))
        else
            log_warning "Kibana overall status is $overall_status"
        fi
    fi

    return 0
}

# Function to query Prometheus metrics
query_prometheus_metrics() {
    log_section "Prometheus Metrics Validation"

    local metrics=(
        "http_requests_total"
        "db_connections_open"
        "p2p_connected_peers"
        "lb_requests_total"
    )

    for metric in "${metrics[@]}"; do
        log_info "Querying metric: $metric"

        local query_url="$PROMETHEUS_URL/api/v1/query?query=$metric"
        local response=$(curl -s --max-time "$METRICS_QUERY_TIMEOUT" "$query_url" 2>&1)

        if [ $? -eq 0 ]; then
            local status=$(echo "$response" | jq -r '.status' 2>/dev/null)
            local result_count=$(echo "$response" | jq '.data.result | length' 2>/dev/null || echo "0")

            if [ "$status" = "success" ]; then
                if [ "$result_count" -gt 0 ]; then
                    log_success "Metric $metric found ($result_count series)"
                    ((CHECKS_PASSED++))
                else
                    log_warning "Metric $metric query succeeded but returned no data"
                    ((CHECKS_PASSED++))
                fi
            else
                log_error "Failed to query metric $metric"
                FAILED_CHECKS+=("Prometheus metric: $metric")
                ((CHECKS_FAILED++))
            fi
        else
            log_error "Failed to query Prometheus for metric $metric"
            FAILED_CHECKS+=("Prometheus query: $metric")
            ((CHECKS_FAILED++))
        fi
    done
}

# Function to verify Jaeger traces
verify_jaeger_traces() {
    log_section "Jaeger Traces Verification"

    local service_name="${JAEGER_SERVICE_NAME:-ollamamax-api}"
    log_info "Checking traces for service: $service_name"

    local traces_url="$JAEGER_URL/api/traces?service=$service_name&limit=10"
    local response=$(curl -s --max-time "$METRICS_QUERY_TIMEOUT" "$traces_url" 2>&1)

    if [ $? -eq 0 ]; then
        local trace_count=$(echo "$response" | jq '.data | length' 2>/dev/null || echo "0")

        if [ "$trace_count" -gt 0 ]; then
            log_success "Found $trace_count traces for service $service_name"
            ((CHECKS_PASSED++))
        else
            log_warning "No traces found for service $service_name (this may be normal for new deployments)"
            ((CHECKS_PASSED++))
        fi
    else
        log_error "Failed to query Jaeger traces"
        FAILED_CHECKS+=("Jaeger traces")
        ((CHECKS_FAILED++))
    fi

    # Check Jaeger services
    log_info "Checking available services in Jaeger..."
    local services_url="$JAEGER_URL/api/services"
    local services=$(curl -s --max-time "$HEALTH_CHECK_TIMEOUT" "$services_url" 2>&1)

    if [ $? -eq 0 ]; then
        local service_count=$(echo "$services" | jq '.data | length' 2>/dev/null || echo "0")
        log_success "Jaeger has $service_count services tracked"
        ((CHECKS_PASSED++))
    else
        log_warning "Failed to retrieve Jaeger services list"
    fi
}

# Function to verify Elasticsearch logs
verify_elasticsearch_logs() {
    log_section "Elasticsearch Logs Verification"

    local index_pattern="${ELASTICSEARCH_INDEX_PATTERN:-ollamamax-logs-*}"
    log_info "Checking logs in index pattern: $index_pattern"

    local count_url="$ELASTICSEARCH_URL/$index_pattern/_count"
    local response=$(curl -s --max-time "$METRICS_QUERY_TIMEOUT" "$count_url" 2>&1)

    if [ $? -eq 0 ]; then
        local doc_count=$(echo "$response" | jq '.count' 2>/dev/null || echo "0")

        if [ "$doc_count" -gt 0 ]; then
            log_success "Found $doc_count log documents in $index_pattern"
            ((CHECKS_PASSED++))
        else
            log_warning "No log documents found in $index_pattern (this may be normal for new deployments)"
            ((CHECKS_PASSED++))
        fi
    else
        log_error "Failed to count documents in $index_pattern"
        FAILED_CHECKS+=("Elasticsearch logs count")
        ((CHECKS_FAILED++))
    fi

    # Check index health
    log_info "Checking index health for $index_pattern"
    local indices_url="$ELASTICSEARCH_URL/_cat/indices/$index_pattern?format=json"
    local indices=$(curl -s --max-time "$HEALTH_CHECK_TIMEOUT" "$indices_url" 2>&1)

    if [ $? -eq 0 ]; then
        local index_count=$(echo "$indices" | jq '. | length' 2>/dev/null || echo "0")
        if [ "$index_count" -gt 0 ]; then
            log_success "Found $index_count indices matching pattern $index_pattern"
            ((CHECKS_PASSED++))
        else
            log_warning "No indices found matching pattern $index_pattern"
        fi
    else
        log_warning "Failed to retrieve index information"
    fi
}

# Function to print summary report
print_summary() {
    log_section "Validation Summary Report"

    echo ""
    echo "╔════════════════════════════════════════╗"
    echo "║   Monitoring Stack Validation Report  ║"
    echo "╠════════════════════════════════════════╣"
    echo "║                                        ║"
    printf "║  Total Checks:       %-16s ║\n" "$((CHECKS_PASSED + CHECKS_FAILED))"
    printf "║  ${GREEN}Checks Passed:${NC}       %-16s ║\n" "$CHECKS_PASSED"
    printf "║  ${RED}Checks Failed:${NC}       %-16s ║\n" "$CHECKS_FAILED"
    echo "║                                        ║"
    echo "╚════════════════════════════════════════╝"
    echo ""

    if [ ${#FAILED_CHECKS[@]} -gt 0 ]; then
        echo -e "${RED}Failed Checks:${NC}"
        for check in "${FAILED_CHECKS[@]}"; do
            echo -e "  ${RED}✗${NC} $check"
        done
        echo ""
    fi

    # Component summary
    echo "Component Status:"
    echo "─────────────────────────────────────────"

    local components=(
        "Prometheus:$PROMETHEUS_URL"
        "Grafana:$GRAFANA_URL"
        "Alertmanager:$ALERTMANAGER_URL"
        "Jaeger:$JAEGER_URL"
        "Elasticsearch:$ELASTICSEARCH_URL"
        "Logstash:$LOGSTASH_URL"
        "Kibana:$KIBANA_URL"
    )

    for component in "${components[@]}"; do
        local name="${component%%:*}"
        local url="${component#*:}"
        echo "  $name: $url"
    done
    echo ""

    if [ $CHECKS_FAILED -eq 0 ]; then
        echo -e "${GREEN}✓ All monitoring stack validation checks passed!${NC}"
        echo ""
        return 0
    else
        echo -e "${RED}✗ Some monitoring stack validation checks failed!${NC}"
        echo ""
        return 1
    fi
}

# Main execution
main() {
    echo "╔════════════════════════════════════════╗"
    echo "║  OllamaMax Monitoring Stack Validator ║"
    echo "╚════════════════════════════════════════╝"
    echo ""
    echo "Validating monitoring infrastructure..."
    echo ""

    # Check for required commands
    if ! command -v curl &> /dev/null; then
        log_error "curl is required but not installed"
        exit 1
    fi

    if ! command -v jq &> /dev/null; then
        log_warning "jq is not installed, JSON parsing will be limited"
    fi

    # Run all validation checks
    check_prometheus || true
    check_grafana || true
    check_alertmanager || true
    check_jaeger || true
    check_elasticsearch || true
    check_logstash || true
    check_kibana || true

    # Query metrics and verify data
    query_prometheus_metrics || true
    verify_jaeger_traces || true
    verify_elasticsearch_logs || true

    # Print summary and exit with appropriate code
    if print_summary; then
        exit 0
    else
        exit 1
    fi
}

# Execute main function
main "$@"

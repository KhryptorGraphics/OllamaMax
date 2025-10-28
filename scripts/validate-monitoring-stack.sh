#!/bin/bash
set -e

echo "🔍 Validating OllamaMax Monitoring Stack"
echo "========================================"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

REPORT_FILE="docs/monitoring-validation-report.md"
FAILED_CHECKS=0

# Initialize report
mkdir -p docs
cat > "$REPORT_FILE" << 'EOFREPORT'
# OllamaMax Monitoring Stack Validation Report
Generated: TIMESTAMP_PLACEHOLDER

## Component Status

EOFREPORT

# Replace timestamp
sed -i "s/TIMESTAMP_PLACEHOLDER/$(date)/" "$REPORT_FILE"

# Check service health
check_service() {
    local service=$1
    local url=$2
    local expected_code=${3:-200}

    echo -e "\n${YELLOW}Checking $service...${NC}"

    response=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null || echo "000")

    if [ "$response" = "$expected_code" ]; then
        echo -e "${GREEN}✅ $service is healthy${NC}"
        echo "- [x] $service (HTTP $response)" >> "$REPORT_FILE"
        return 0
    else
        echo -e "${RED}❌ $service is unhealthy (HTTP $response)${NC}"
        echo "- [ ] $service (HTTP $response) ⚠️ FAILED" >> "$REPORT_FILE"
        ((FAILED_CHECKS++))
        return 1
    fi
}

# Check Prometheus metrics
check_prometheus_metrics() {
    echo -e "\n${YELLOW}Checking Prometheus metrics...${NC}"

    # Check for key OllamaMax metrics
    metrics=(
        "ollamamax_api_http_requests_total"
        "ollamamax_database_db_connections_open"
        "ollamamax_p2p_connected_peers"
        "ollamamax_loadbalancer_requests_total"
    )

    echo -e "\n### Prometheus Metrics" >> "$REPORT_FILE"

    local missing_metrics=0

    for metric in "${metrics[@]}"; do
        result=$(curl -s "http://localhost:9090/api/v1/query?query=$metric" 2>/dev/null | grep -o '"status":"success"' || true)

        if [ -n "$result" ]; then
            echo -e "${GREEN}✅ Metric found: $metric${NC}"
            echo "- [x] $metric" >> "$REPORT_FILE"
        else
            echo -e "${RED}❌ Metric missing: $metric${NC}"
            echo "- [ ] $metric ⚠️ MISSING" >> "$REPORT_FILE"
            ((missing_metrics++))
        fi
    done

    if [ $missing_metrics -gt 0 ]; then
        ((FAILED_CHECKS++))
    fi
}

# Check Jaeger traces
check_jaeger_traces() {
    echo -e "\n${YELLOW}Checking Jaeger traces...${NC}"

    services=$(curl -s "http://localhost:16686/api/services" 2>/dev/null | grep -o '"ollamamax-api"' || true)

    echo -e "\n### Jaeger Tracing" >> "$REPORT_FILE"

    if [ -n "$services" ]; then
        echo -e "${GREEN}✅ Jaeger has traces for ollamamax-api${NC}"
        echo "- [x] Service traces found" >> "$REPORT_FILE"
    else
        echo -e "${YELLOW}⚠️  No traces found in Jaeger (may need traffic)${NC}"
        echo "- [ ] Service traces found ⚠️ No traces" >> "$REPORT_FILE"
    fi
}

# Check Elasticsearch indices
check_elasticsearch_indices() {
    echo -e "\n${YELLOW}Checking Elasticsearch indices...${NC}"

    indices=$(curl -s "http://localhost:9200/_cat/indices/ollamamax-*?h=index" 2>/dev/null || echo "")

    echo -e "\n### Elasticsearch Indices" >> "$REPORT_FILE"

    if [ -n "$indices" ]; then
        echo -e "${GREEN}✅ Found Elasticsearch indices:${NC}"
        echo "$indices" | while read -r index; do
            if [ -n "$index" ]; then
                echo "  - $index"
                echo "- [x] $index" >> "$REPORT_FILE"
            fi
        done
    else
        echo -e "${YELLOW}⚠️  No OllamaMax indices found in Elasticsearch${NC}"
        echo "- [ ] No indices found ⚠️ Check Filebeat/Logstash" >> "$REPORT_FILE"
    fi
}

# Check alert rules
check_alert_rules() {
    echo -e "\n${YELLOW}Checking Prometheus alert rules...${NC}"

    rules=$(curl -s "http://localhost:9090/api/v1/rules" 2>/dev/null | grep -o '"name":"ollamamax-' || true)

    echo -e "\n### Alert Rules" >> "$REPORT_FILE"

    if [ -n "$rules" ]; then
        echo -e "${GREEN}✅ Alert rules loaded${NC}"
        echo "- [x] Rules loaded successfully" >> "$REPORT_FILE"
    else
        echo -e "${RED}❌ No alert rules found${NC}"
        echo "- [ ] Rules loaded ⚠️ FAILED" >> "$REPORT_FILE"
        ((FAILED_CHECKS++))
    fi
}

# Check Grafana datasources
check_grafana_datasources() {
    echo -e "\n${YELLOW}Checking Grafana datasources...${NC}"

    datasources=$(curl -s -u admin:admin "http://localhost:3001/api/datasources" 2>/dev/null || echo "[]")

    echo -e "\n### Grafana Datasources" >> "$REPORT_FILE"

    if echo "$datasources" | grep -q "Prometheus"; then
        echo -e "${GREEN}✅ Prometheus datasource configured${NC}"
        echo "- [x] Prometheus datasource" >> "$REPORT_FILE"
    else
        echo -e "${RED}❌ Prometheus datasource not found${NC}"
        echo "- [ ] Prometheus datasource ⚠️ MISSING" >> "$REPORT_FILE"
        ((FAILED_CHECKS++))
    fi

    if echo "$datasources" | grep -q "Jaeger"; then
        echo -e "${GREEN}✅ Jaeger datasource configured${NC}"
        echo "- [x] Jaeger datasource" >> "$REPORT_FILE"
    else
        echo -e "${YELLOW}⚠️  Jaeger datasource not found${NC}"
        echo "- [ ] Jaeger datasource ⚠️ MISSING" >> "$REPORT_FILE"
    fi
}

# Check Grafana dashboards
check_grafana_dashboards() {
    echo -e "\n${YELLOW}Checking Grafana dashboards...${NC}"

    dashboards=$(curl -s -u admin:admin "http://localhost:3001/api/search?type=dash-db" 2>/dev/null || echo "[]")

    echo -e "\n### Grafana Dashboards" >> "$REPORT_FILE"

    dashboard_count=$(echo "$dashboards" | grep -o '"title"' | wc -l)

    if [ "$dashboard_count" -gt 0 ]; then
        echo -e "${GREEN}✅ Found $dashboard_count dashboard(s)${NC}"
        echo "- [x] Dashboards loaded ($dashboard_count found)" >> "$REPORT_FILE"
    else
        echo -e "${YELLOW}⚠️  No dashboards found${NC}"
        echo "- [ ] Dashboards loaded ⚠️ NONE FOUND" >> "$REPORT_FILE"
    fi
}

# Check Docker containers
check_docker_containers() {
    echo -e "\n${YELLOW}Checking Docker containers...${NC}"

    echo -e "\n### Docker Containers" >> "$REPORT_FILE"

    containers=(
        "prometheus"
        "grafana"
        "alertmanager"
        "jaeger"
        "elasticsearch"
        "kibana"
    )

    for container in "${containers[@]}"; do
        if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "$container"; then
            status=$(docker ps --format '{{.Names}}: {{.Status}}' 2>/dev/null | grep "$container" || echo "unknown")
            echo -e "${GREEN}✅ Container running: $status${NC}"
            echo "- [x] $container (running)" >> "$REPORT_FILE"
        else
            echo -e "${RED}❌ Container not running: $container${NC}"
            echo "- [ ] $container ⚠️ NOT RUNNING" >> "$REPORT_FILE"
            ((FAILED_CHECKS++))
        fi
    done
}

# Check disk space for monitoring data
check_disk_space() {
    echo -e "\n${YELLOW}Checking disk space...${NC}"

    echo -e "\n### Disk Space" >> "$REPORT_FILE"

    available=$(df -h . | tail -1 | awk '{print $4}')
    used_percent=$(df -h . | tail -1 | awk '{print $5}' | tr -d '%')

    echo -e "${GREEN}✅ Available disk space: $available${NC}"
    echo "- Available: $available" >> "$REPORT_FILE"

    if [ "$used_percent" -gt 90 ]; then
        echo -e "${RED}❌ Disk usage is at ${used_percent}%${NC}"
        echo "- [ ] Disk usage: ${used_percent}% ⚠️ CRITICAL" >> "$REPORT_FILE"
        ((FAILED_CHECKS++))
    elif [ "$used_percent" -gt 80 ]; then
        echo -e "${YELLOW}⚠️  Disk usage is at ${used_percent}%${NC}"
        echo "- [x] Disk usage: ${used_percent}% ⚠️ WARNING" >> "$REPORT_FILE"
    else
        echo -e "${GREEN}✅ Disk usage is at ${used_percent}%${NC}"
        echo "- [x] Disk usage: ${used_percent}%" >> "$REPORT_FILE"
    fi
}

# Check monitoring data retention
check_data_retention() {
    echo -e "\n${YELLOW}Checking data retention settings...${NC}"

    echo -e "\n### Data Retention" >> "$REPORT_FILE"

    # Check Prometheus retention
    prom_retention=$(curl -s "http://localhost:9090/api/v1/status/runtimeinfo" 2>/dev/null | grep -o '"retentionTime":"[^"]*"' | cut -d'"' -f4 || echo "unknown")

    echo -e "${GREEN}✅ Prometheus retention: $prom_retention${NC}"
    echo "- Prometheus: $prom_retention" >> "$REPORT_FILE"

    # Check Elasticsearch index size
    es_size=$(curl -s "http://localhost:9200/_cat/indices/ollamamax-*?h=store.size" 2>/dev/null | awk '{sum += $1} END {print sum "gb"}' || echo "unknown")

    echo -e "${GREEN}✅ Elasticsearch data size: $es_size${NC}"
    echo "- Elasticsearch: $es_size" >> "$REPORT_FILE"
}

# Generate recommendations
generate_recommendations() {
    echo -e "\n### Recommendations" >> "$REPORT_FILE"

    if [ $FAILED_CHECKS -eq 0 ]; then
        cat >> "$REPORT_FILE" << 'EOFREC'

All monitoring components are functioning correctly. Consider these optimization steps:

1. **Performance Tuning**: Review dashboard queries for optimization opportunities
2. **Alert Tuning**: Adjust alert thresholds based on observed baseline metrics
3. **Retention Policies**: Verify data retention matches your compliance requirements
4. **Capacity Planning**: Monitor disk usage trends and plan for growth
5. **Dashboard Enhancement**: Add custom dashboards for application-specific metrics

EOFREC
    else
        cat >> "$REPORT_FILE" << 'EOFREC'

**Action Required**: Fix the following issues:

EOFREC
        if grep -q "NOT RUNNING" "$REPORT_FILE"; then
            echo "1. Start missing Docker containers with: \`docker-compose up -d\`" >> "$REPORT_FILE"
        fi

        if grep -q "MISSING" "$REPORT_FILE"; then
            echo "2. Configure missing datasources and verify service connectivity" >> "$REPORT_FILE"
        fi

        if grep -q "CRITICAL" "$REPORT_FILE"; then
            echo "3. Address disk space issues immediately to prevent data loss" >> "$REPORT_FILE"
        fi

        cat >> "$REPORT_FILE" << 'EOFREC'

After fixing issues, run this validation script again to verify.
EOFREC
    fi
}

# Main validation
main() {
    # Core services
    check_service "Prometheus" "http://localhost:9090/-/healthy"
    check_service "Grafana" "http://localhost:3001/api/health"
    check_service "Alertmanager" "http://localhost:9093/-/healthy"
    check_service "Jaeger UI" "http://localhost:16686"
    check_service "Elasticsearch" "http://localhost:9200/_cluster/health"
    check_service "Kibana" "http://localhost:5601/api/status"

    # Metrics and traces
    check_prometheus_metrics
    check_jaeger_traces
    check_elasticsearch_indices
    check_alert_rules

    # Grafana checks
    check_grafana_datasources
    check_grafana_dashboards

    # Infrastructure checks
    check_docker_containers
    check_disk_space
    check_data_retention

    # Generate recommendations
    generate_recommendations

    # Summary
    echo -e "\n======================================" | tee -a "$REPORT_FILE"
    echo -e "\n## Summary\n" >> "$REPORT_FILE"

    if [ $FAILED_CHECKS -eq 0 ]; then
        echo -e "${GREEN}✅ All monitoring components validated successfully!${NC}" | tee -a "$REPORT_FILE"
        echo -e "\nValidation: **PASSED** ✅" >> "$REPORT_FILE"
        echo -e "\n${GREEN}📄 Full report saved to $REPORT_FILE${NC}"
        exit 0
    else
        echo -e "${RED}❌ $FAILED_CHECKS validation check(s) failed${NC}" | tee -a "$REPORT_FILE"
        echo -e "\nValidation: **FAILED** ❌ ($FAILED_CHECKS issues)" >> "$REPORT_FILE"
        echo -e "\n${YELLOW}📄 Full report saved to $REPORT_FILE${NC}"
        exit 1
    fi
}

main

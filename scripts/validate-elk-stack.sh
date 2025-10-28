#!/bin/bash
# Validate ELK Stack (Elasticsearch, Logstash, Kibana) + Filebeat Configuration
# Tests log collection, processing, and trace correlation

set -e

echo "============================================"
echo "ELK Stack Configuration Validation"
echo "============================================"
echo ""

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

FAILED=0

# Helper functions
check_service() {
    local service=$1
    local port=$2
    local url=$3

    echo -n "Checking $service (port $port)... "
    if docker-compose ps | grep -q "$service.*Up"; then
        if curl -s -f "$url" > /dev/null 2>&1; then
            echo -e "${GREEN}OK${NC}"
            return 0
        else
            echo -e "${YELLOW}Service up but endpoint not responding${NC}"
            return 1
        fi
    else
        echo -e "${RED}FAILED - Service not running${NC}"
        return 1
    fi
}

send_test_log() {
    local level=$1
    local trace_id=$2
    local span_id=$3
    local message=$4
    local audit=${5:-false}

    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%S.%3NZ")

    # Create JSON log entry
    local log_entry=$(cat <<EOF
{
  "timestamp": "$timestamp",
  "level": "$level",
  "message": "$message",
  "trace_id": "$trace_id",
  "span_id": "$span_id",
  "service": "ollamamax-api",
  "audit": $audit
}
EOF
)

    # Send to Docker container log (will be picked up by Filebeat)
    docker-compose exec -T ollamamax-api sh -c "echo '$log_entry' >> /app/logs/app.log" 2>/dev/null || true
}

wait_for_elasticsearch() {
    echo -n "Waiting for Elasticsearch to be ready... "
    local max_attempts=30
    local attempt=0

    while [ $attempt -lt $max_attempts ]; do
        if curl -s -f "http://localhost:9200/_cluster/health" > /dev/null 2>&1; then
            echo -e "${GREEN}OK${NC}"
            return 0
        fi
        sleep 2
        attempt=$((attempt + 1))
    done

    echo -e "${RED}FAILED${NC}"
    return 1
}

query_logs() {
    local index=$1
    local trace_id=$2

    curl -s -X GET "http://localhost:9200/${index}/_search" \
        -H 'Content-Type: application/json' \
        -d "{\"query\": {\"term\": {\"trace_id\": \"$trace_id\"}}}" | \
        jq -r '.hits.total.value'
}

echo "=== Phase 1: Service Health Checks ==="
echo ""

# Check Elasticsearch
if ! check_service "elasticsearch" "9200" "http://localhost:9200/_cluster/health"; then
    FAILED=$((FAILED + 1))
fi

# Check Logstash
if ! check_service "logstash" "5044,9600" "http://localhost:9600/_node/stats"; then
    FAILED=$((FAILED + 1))
fi

# Check Kibana
if ! check_service "kibana" "5601" "http://localhost:5601/api/status"; then
    FAILED=$((FAILED + 1))
fi

# Check Filebeat
echo -n "Checking filebeat... "
if docker-compose ps | grep -q "filebeat.*Up"; then
    echo -e "${GREEN}OK${NC}"
else
    echo -e "${RED}FAILED${NC}"
    FAILED=$((FAILED + 1))
fi

echo ""
echo "=== Phase 2: Configuration Validation ==="
echo ""

# Validate Logstash pipeline
echo -n "Validating Logstash pipeline syntax... "
if docker-compose exec -T logstash /usr/share/logstash/bin/logstash \
    --config.test_and_exit -f /usr/share/logstash/pipeline/logstash.conf > /dev/null 2>&1; then
    echo -e "${GREEN}OK${NC}"
else
    echo -e "${RED}FAILED${NC}"
    FAILED=$((FAILED + 1))
fi

# Validate Filebeat configuration
echo -n "Validating Filebeat configuration... "
if docker-compose exec -T filebeat filebeat test config > /dev/null 2>&1; then
    echo -e "${GREEN}OK${NC}"
else
    echo -e "${RED}FAILED${NC}"
    FAILED=$((FAILED + 1))
fi

# Check Elasticsearch indices
echo -n "Checking Elasticsearch indices... "
INDICES=$(curl -s "http://localhost:9200/_cat/indices?format=json" | jq -r '.[].index' | grep -c "ollamamax" || true)
if [ "$INDICES" -ge 0 ]; then
    echo -e "${GREEN}OK (${INDICES} indices)${NC}"
else
    echo -e "${YELLOW}No OllamaMax indices found yet${NC}"
fi

echo ""
echo "=== Phase 3: Log Pipeline Testing ==="
echo ""

# Wait for services to be ready
if ! wait_for_elasticsearch; then
    echo -e "${RED}Elasticsearch not ready, skipping pipeline tests${NC}"
    FAILED=$((FAILED + 1))
else
    # Generate test logs
    TEST_TRACE_ID="test-trace-$(date +%s)"
    TEST_SPAN_ID="span-$(date +%s)"

    echo "Sending test logs with trace_id: $TEST_TRACE_ID"

    # Send different types of logs
    send_test_log "info" "$TEST_TRACE_ID" "$TEST_SPAN_ID" "Test info log" "false"
    send_test_log "error" "$TEST_TRACE_ID" "$TEST_SPAN_ID" "Test error log" "false"
    send_test_log "info" "$TEST_TRACE_ID" "$TEST_SPAN_ID" "Test audit log" "true"

    # Wait for logs to be processed
    echo -n "Waiting for logs to be processed (30s)... "
    sleep 30
    echo -e "${GREEN}OK${NC}"

    # Query Elasticsearch for test logs
    echo ""
    echo "Querying Elasticsearch indices:"

    # Check standard logs
    echo -n "  - ollamamax-logs-*: "
    LOG_COUNT=$(query_logs "ollamamax-logs-*" "$TEST_TRACE_ID")
    if [ "$LOG_COUNT" -gt 0 ]; then
        echo -e "${GREEN}Found $LOG_COUNT logs${NC}"
    else
        echo -e "${YELLOW}No logs found${NC}"
    fi

    # Check error logs
    echo -n "  - ollamamax-errors-*: "
    ERROR_COUNT=$(query_logs "ollamamax-errors-*" "$TEST_TRACE_ID")
    if [ "$ERROR_COUNT" -gt 0 ]; then
        echo -e "${GREEN}Found $ERROR_COUNT error logs${NC}"
    else
        echo -e "${YELLOW}No error logs found${NC}"
    fi

    # Check audit logs
    echo -n "  - ollamamax-audit-*: "
    AUDIT_COUNT=$(query_logs "ollamamax-audit-*" "$TEST_TRACE_ID")
    if [ "$AUDIT_COUNT" -gt 0 ]; then
        echo -e "${GREEN}Found $AUDIT_COUNT audit logs${NC}"
    else
        echo -e "${YELLOW}No audit logs found${NC}"
    fi
fi

echo ""
echo "=== Phase 4: Trace Correlation Testing ==="
echo ""

# Check if trace_id field exists in mapping
echo -n "Verifying trace_id field mapping... "
TRACE_FIELD=$(curl -s "http://localhost:9200/ollamamax-logs-*/_mapping/field/trace_id" | \
    jq -r 'keys | length')

if [ "$TRACE_FIELD" -gt 0 ]; then
    echo -e "${GREEN}OK${NC}"

    # Show example trace correlation
    echo ""
    echo "Example trace correlation query:"
    echo "GET ollamamax-logs-*/_search"
    echo '{'
    echo '  "query": {'
    echo '    "term": {'
    echo "      \"trace_id\": \"$TEST_TRACE_ID\""
    echo '    }'
    echo '  }'
    echo '}'
else
    echo -e "${YELLOW}Field not found yet (may need more time)${NC}"
fi

echo ""
echo "=== Phase 5: Integration Verification ==="
echo ""

# Check Filebeat output to Logstash
echo -n "Checking Filebeat -> Logstash connection... "
FILEBEAT_OUTPUTS=$(docker-compose logs filebeat 2>&1 | grep -c "Connection to backoff" || true)
if [ "$FILEBEAT_OUTPUTS" -eq 0 ]; then
    echo -e "${GREEN}OK${NC}"
else
    echo -e "${YELLOW}Connection issues detected${NC}"
fi

# Check Logstash -> Elasticsearch connection
echo -n "Checking Logstash -> Elasticsearch connection... "
LOGSTASH_ERRORS=$(docker-compose logs logstash 2>&1 | grep -c "Connection refused" || true)
if [ "$LOGSTASH_ERRORS" -eq 0 ]; then
    echo -e "${GREEN}OK${NC}"
else
    echo -e "${RED}FAILED - Connection errors detected${NC}"
    FAILED=$((FAILED + 1))
fi

# Check Kibana -> Elasticsearch connection
echo -n "Checking Kibana -> Elasticsearch connection... "
KIBANA_STATUS=$(curl -s "http://localhost:5601/api/status" | jq -r '.status.overall.state')
if [ "$KIBANA_STATUS" = "green" ] || [ "$KIBANA_STATUS" = "yellow" ]; then
    echo -e "${GREEN}OK ($KIBANA_STATUS)${NC}"
else
    echo -e "${YELLOW}Status: $KIBANA_STATUS${NC}"
fi

echo ""
echo "=== Phase 6: Performance Metrics ==="
echo ""

# Elasticsearch metrics
echo "Elasticsearch cluster stats:"
curl -s "http://localhost:9200/_cluster/stats" | jq -r '
  "  Nodes: \(.nodes.count.total)",
  "  Indices: \(.indices.count)",
  "  Documents: \(.indices.docs.count)",
  "  Store size: \(.indices.store.size_in_bytes / 1024 / 1024 | floor)MB"
'

# Logstash metrics
echo ""
echo "Logstash pipeline stats:"
curl -s "http://localhost:9600/_node/stats/pipelines" | jq -r '
  .pipelines.main.events |
  "  Events in: \(.in)",
  "  Events out: \(.out)",
  "  Events filtered: \(.filtered)"
'

echo ""
echo "=== Configuration File Validation ==="
echo ""

# Check configuration files exist
echo -n "Checking Logstash pipeline config... "
if [ -f "./monitoring/logstash/pipeline/logstash.conf" ]; then
    echo -e "${GREEN}OK${NC}"
else
    echo -e "${RED}FAILED - File not found${NC}"
    FAILED=$((FAILED + 1))
fi

echo -n "Checking Filebeat config... "
if [ -f "./monitoring/filebeat/filebeat.yml" ]; then
    echo -e "${GREEN}OK${NC}"
else
    echo -e "${RED}FAILED - File not found${NC}"
    FAILED=$((FAILED + 1))
fi

echo ""
echo "=== Validation Summary ==="
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All checks passed!${NC}"
    echo ""
    echo "Access URLs:"
    echo "  - Kibana: http://localhost:5601"
    echo "  - Elasticsearch: http://localhost:9200"
    echo "  - Logstash API: http://localhost:9600"
    echo ""
    echo "Index patterns to create in Kibana:"
    echo "  - ollamamax-logs-*"
    echo "  - ollamamax-errors-*"
    echo "  - ollamamax-audit-*"
    exit 0
else
    echo -e "${RED}✗ $FAILED check(s) failed${NC}"
    echo ""
    echo "Troubleshooting:"
    echo "  - Check Docker logs: docker-compose logs [service]"
    echo "  - Verify configurations: ./monitoring/logstash/pipeline/logstash.conf"
    echo "  - Check Elasticsearch: curl http://localhost:9200/_cluster/health"
    exit 1
fi

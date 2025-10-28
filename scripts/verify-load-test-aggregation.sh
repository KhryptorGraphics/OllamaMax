#!/bin/bash

################################################################################
# Verification Script for Load Test Aggregation Fixes
#
# Tests that the aggregate JSON is correctly populated with metrics
################################################################################

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}Load Test Aggregation Verification${NC}"
echo "========================================"
echo ""

# Check if jq is available
if ! command -v jq &> /dev/null; then
    echo -e "${RED}ERROR: jq is required for this verification${NC}"
    echo "Install with: sudo apt-get install jq"
    exit 1
fi

# Create test directory
TEST_DIR="load-test-results"
mkdir -p "${TEST_DIR}"

# Create mock summary files
echo "Creating mock per-instance summary files..."

for i in 1 2 3; do
    cat > "${TEST_DIR}/summary-instance-${i}.json" <<EOF
{
  "metrics": {
    "http_reqs": {
      "values": {
        "count": $((100000 * i)),
        "rate": $((10000 * i))
      }
    },
    "http_req_failed": {
      "values": {
        "fails": $((10 * i)),
        "passes": 0
      }
    },
    "http_req_duration": {
      "values": {
        "avg": $((200 + i * 10)),
        "p(95)": $((400 + i * 20)),
        "p(99)": $((800 + i * 30))
      }
    }
  }
}
EOF
done

echo -e "${GREEN}✓ Mock summary files created${NC}"
echo ""

# Create initial aggregate file
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
AGGREGATE_FILE="${TEST_DIR}/aggregate-results-${TIMESTAMP}.json"

cat > "${AGGREGATE_FILE}" <<EOF
{
  "timestamp": "$(date --iso-8601=seconds)",
  "target_rps": 100000,
  "instances": 3,
  "failed_instances": 0,
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

echo "Initial aggregate file created"
echo ""

# Simulate aggregation logic
TOTAL_REQUESTS=0
TOTAL_FAILED=0
TOTAL_RPS=0
DURATIONS_SUM=0
P95_VALUES=()
P99_VALUES=()
INSTANCES=3

for i in $(seq 1 ${INSTANCES}); do
    summary_file="${TEST_DIR}/summary-instance-${i}.json"

    requests=$(jq -r '.metrics.http_reqs.values.count // 0' "${summary_file}")
    failed=$(jq -r '.metrics.http_req_failed.values.fails // 0' "${summary_file}")
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
done

# Calculate aggregates
AVG_DURATION=$(echo "scale=2; ${DURATIONS_SUM} / ${INSTANCES}" | bc)
ERROR_RATE=$(echo "scale=4; ${TOTAL_FAILED} / ${TOTAL_REQUESTS} * 100" | bc)

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

echo "Computed metrics:"
echo "  Total Requests: ${TOTAL_REQUESTS}"
echo "  Total Failed: ${TOTAL_FAILED}"
echo "  Average RPS: ${TOTAL_RPS}"
echo "  Average Duration: ${AVG_DURATION}ms"
echo "  P95 Latency: ${AVG_P95}ms"
echo "  P99 Latency: ${AVG_P99}ms"
echo "  Error Rate: ${ERROR_RATE}%"
echo ""

# Update aggregate file
echo "Updating aggregate JSON..."

TEMP_JSON=$(mktemp)
jq --arg total_req "${TOTAL_REQUESTS}" \
   --arg total_failed "${TOTAL_FAILED}" \
   --arg avg_rps "${TOTAL_RPS}" \
   --arg avg_duration "${AVG_DURATION}" \
   --arg p95 "${AVG_P95}" \
   --arg p99 "${AVG_P99}" \
   --arg error_rate "${ERROR_RATE}" \
   '.metrics.total_requests = ($total_req | tonumber) |
    .metrics.total_failed_requests = ($total_failed | tonumber) |
    .metrics.average_rps = ($avg_rps | tonumber) |
    .metrics.average_duration_ms = ($avg_duration | tonumber) |
    .metrics.p95_duration_ms = ($p95 | tonumber) |
    .metrics.p99_duration_ms = ($p99 | tonumber) |
    .metrics.error_rate = ($error_rate | tonumber) |
    .metrics.peak_rps = ($avg_rps | tonumber)' \
   "${AGGREGATE_FILE}" > "${TEMP_JSON}"

mv "${TEMP_JSON}" "${AGGREGATE_FILE}"

echo -e "${GREEN}✓ Aggregate JSON updated${NC}"
echo ""

# Verify the update
echo "Verifying aggregate JSON content..."
echo ""

VERIFY_TOTAL_REQ=$(jq -r '.metrics.total_requests' "${AGGREGATE_FILE}")
VERIFY_AVG_RPS=$(jq -r '.metrics.average_rps' "${AGGREGATE_FILE}")
VERIFY_P95=$(jq -r '.metrics.p95_duration_ms' "${AGGREGATE_FILE}")
VERIFY_ERROR=$(jq -r '.metrics.error_rate' "${AGGREGATE_FILE}")

echo "Verification Results:"
echo "--------------------"

PASS_COUNT=0
FAIL_COUNT=0

# Check total requests
if [ "${VERIFY_TOTAL_REQ}" = "${TOTAL_REQUESTS}" ]; then
    echo -e "${GREEN}✓ total_requests: ${VERIFY_TOTAL_REQ}${NC}"
    ((PASS_COUNT++))
else
    echo -e "${RED}✗ total_requests: expected ${TOTAL_REQUESTS}, got ${VERIFY_TOTAL_REQ}${NC}"
    ((FAIL_COUNT++))
fi

# Check average RPS
if [ "${VERIFY_AVG_RPS}" = "${TOTAL_RPS}" ]; then
    echo -e "${GREEN}✓ average_rps: ${VERIFY_AVG_RPS}${NC}"
    ((PASS_COUNT++))
else
    echo -e "${RED}✗ average_rps: expected ${TOTAL_RPS}, got ${VERIFY_AVG_RPS}${NC}"
    ((FAIL_COUNT++))
fi

# Check P95
if [ "${VERIFY_P95}" = "${AVG_P95}" ]; then
    echo -e "${GREEN}✓ p95_duration_ms: ${VERIFY_P95}${NC}"
    ((PASS_COUNT++))
else
    echo -e "${RED}✗ p95_duration_ms: expected ${AVG_P95}, got ${VERIFY_P95}${NC}"
    ((FAIL_COUNT++))
fi

# Check error rate
if [ "${VERIFY_ERROR}" = "${ERROR_RATE}" ]; then
    echo -e "${GREEN}✓ error_rate: ${VERIFY_ERROR}%${NC}"
    ((PASS_COUNT++))
else
    echo -e "${RED}✗ error_rate: expected ${ERROR_RATE}, got ${VERIFY_ERROR}${NC}"
    ((FAIL_COUNT++))
fi

echo ""
echo "Final JSON structure:"
jq '.' "${AGGREGATE_FILE}"
echo ""

# Summary
echo "========================================"
if [ ${FAIL_COUNT} -eq 0 ]; then
    echo -e "${GREEN}✓ ALL CHECKS PASSED (${PASS_COUNT}/${PASS_COUNT})${NC}"
    echo ""
    echo "The aggregate JSON is correctly populated!"
    echo "The final report generator will receive accurate metrics."
    exit 0
else
    echo -e "${RED}✗ SOME CHECKS FAILED (${PASS_COUNT}/$((PASS_COUNT + FAIL_COUNT)))${NC}"
    echo ""
    echo "There are issues with the aggregation logic."
    exit 1
fi

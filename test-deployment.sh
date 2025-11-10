#!/bin/bash

# Comprehensive Deployment Test Script
# Tests all critical endpoints and functionality

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PASSED=0
FAILED=0
API_BASE="http://localhost:13000"

log_test() {
    echo -e "${BLUE}TEST:${NC} $1"
}

log_pass() {
    echo -e "${GREEN}✓ PASS:${NC} $1"
    ((PASSED++))
}

log_fail() {
    echo -e "${RED}✗ FAIL:${NC} $1"
    ((FAILED++))
}

log_header() {
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo ""
}

# Test 1: Health Check
log_header "Health & Status Endpoints"

log_test "GET /health"
if curl -s -f "$API_BASE/health" > /dev/null; then
    log_pass "Health endpoint responding"
else
    log_fail "Health endpoint not responding"
fi

log_test "GET /health/live"
if curl -s -f "$API_BASE/health/live" > /dev/null; then
    log_pass "Liveness check responding"
else
    log_fail "Liveness check not responding"
fi

log_test "GET /health/ready"
READY_STATUS=$(curl -s -w "%{http_code}" -o /dev/null "$API_BASE/health/ready")
if [ "$READY_STATUS" = "200" ] || [ "$READY_STATUS" = "503" ]; then
    log_pass "Readiness check responding (status: $READY_STATUS)"
else
    log_fail "Readiness check failed (status: $READY_STATUS)"
fi

# Test 2: Node Management
log_header "Node Management"

log_test "GET /api/nodes"
NODES_RESPONSE=$(curl -s "$API_BASE/api/nodes")
NODE_COUNT=$(echo "$NODES_RESPONSE" | jq -r '.nodes | length')
if [ "$NODE_COUNT" -gt 0 ]; then
    log_pass "Nodes endpoint returning $NODE_COUNT nodes"
else
    log_fail "No nodes found"
fi

# Test 3: Model Endpoints
log_header "Model Endpoints"

log_test "GET /v1/models"
MODELS_RESPONSE=$(curl -s "$API_BASE/v1/models")
MODEL_COUNT=$(echo "$MODELS_RESPONSE" | jq -r '.data | length')
if [ "$MODEL_COUNT" -gt 0 ]; then
    log_pass "Models endpoint returning $MODEL_COUNT models"
else
    log_fail "No models found"
fi

# Test 4: Documentation
log_header "Documentation"

log_test "GET /docs"
if curl -s -f "$API_BASE/docs" > /dev/null; then
    log_pass "Documentation endpoint responding"
else
    log_fail "Documentation endpoint not responding"
fi

log_test "GET /openapi.json"
if curl -s -f "$API_BASE/openapi.json" > /dev/null; then
    log_pass "OpenAPI spec available"
else
    log_fail "OpenAPI spec not available"
fi

# Test 5: Metrics
log_header "Metrics & Monitoring"

log_test "GET /metrics"
if curl -s -f "$API_BASE/metrics" > /dev/null; then
    log_pass "Metrics endpoint responding"
else
    log_fail "Metrics endpoint not responding"
fi

# Test 6: Static Files
log_header "Web Interface"

log_test "GET /index.html"
if curl -s -f "$API_BASE/index.html" > /dev/null; then
    log_pass "Web interface accessible"
else
    log_fail "Web interface not accessible"
fi

log_test "GET /auth.html"
if curl -s -f "$API_BASE/auth.html" > /dev/null; then
    log_pass "Auth page accessible"
else
    log_fail "Auth page not accessible"
fi

# Test 7: API Endpoints (without auth)
log_header "Public API Endpoints"

log_test "POST /v1/completions (should require auth)"
COMPLETION_STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X POST "$API_BASE/v1/completions" \
  -H "Content-Type: application/json" \
  -d '{"model":"llama-3.2-3b","prompt":"Hello"}')
if [ "$COMPLETION_STATUS" = "401" ] || [ "$COMPLETION_STATUS" = "403" ]; then
    log_pass "Completions endpoint properly protected (status: $COMPLETION_STATUS)"
else
    log_fail "Completions endpoint security issue (status: $COMPLETION_STATUS)"
fi

# Summary
log_header "Test Summary"

TOTAL=$((PASSED + FAILED))
PASS_RATE=$(awk "BEGIN {printf \"%.1f\", ($PASSED/$TOTAL)*100}")

echo ""
echo -e "Total Tests: ${BLUE}$TOTAL${NC}"
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo -e "Pass Rate: ${BLUE}$PASS_RATE%${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✨ All tests passed!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠️  Some tests failed${NC}"
    exit 1
fi


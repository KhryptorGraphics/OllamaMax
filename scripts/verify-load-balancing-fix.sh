#!/bin/bash
# Quick Verification Script for Load Balancing Fix
# This script verifies that the X-Served-By header is properly configured

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[PASS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[FAIL]${NC} $1"; }

LB_URL="${1:-http://localhost}"

echo -e "${BLUE}=== Load Balancing Fix Verification ===${NC}"
echo -e "${BLUE}Testing URL: ${LB_URL}${NC}\n"

# Test 1: Verify X-Served-By is PRESENT on proxied endpoints
echo -e "${BLUE}Test 1: Proxied Endpoints (should include X-Served-By)${NC}"

PROXIED_ENDPOINTS=(
    "/api/health"
    "/ollama/api/tags"
    "/"
)

for endpoint in "${PROXIED_ENDPOINTS[@]}"; do
    log_info "Testing ${endpoint}..."

    HEADER=$(curl -s -I "${LB_URL}${endpoint}" 2>/dev/null | grep -i "X-Served-By" | tr -d '\r' || true)

    if [ -n "$HEADER" ]; then
        BACKEND=$(echo "$HEADER" | awk '{print $2}' | tr -d '\r')
        log_success "X-Served-By header present: ${BACKEND}"
    else
        log_error "X-Served-By header MISSING (this is a problem)"
    fi
done

echo ""

# Test 2: Verify X-Served-By is ABSENT on non-proxied endpoints
echo -e "${BLUE}Test 2: Non-Proxied Endpoints (should NOT include X-Served-By)${NC}"

NON_PROXIED_ENDPOINTS=(
    "/health"
    "/nginx-health"
)

for endpoint in "${NON_PROXIED_ENDPOINTS[@]}"; do
    log_info "Testing ${endpoint}..."

    HEADER=$(curl -s -I "${LB_URL}${endpoint}" 2>/dev/null | grep -i "X-Served-By" || true)

    if [ -z "$HEADER" ]; then
        log_success "X-Served-By header correctly absent (expected)"
    else
        log_warning "X-Served-By header present (not expected for non-proxied endpoint)"
    fi
done

echo ""

# Test 3: Verify load balancing script has correct parameters
echo -e "${BLUE}Test 3: Load Balancing Script Configuration${NC}"

if [ -f "scripts/test-load-balancing.sh" ]; then
    log_info "Checking test script configuration..."

    if grep -q 'LB_PATH="${2:-/api/health}"' scripts/test-load-balancing.sh; then
        log_success "Script uses configurable proxied endpoint"
    else
        log_error "Script does not have LB_PATH parameter"
    fi

    if grep -q 'tr -d' scripts/test-load-balancing.sh; then
        log_success "Script includes CRLF trimming"
    else
        log_warning "Script may not trim CRLF properly"
    fi

    if grep -q "PROXIED endpoints" scripts/test-load-balancing.sh; then
        log_success "Script includes proxied endpoint documentation"
    else
        log_warning "Script missing documentation"
    fi
else
    log_error "test-load-balancing.sh not found"
fi

echo ""

# Test 4: Verify Nginx configuration
echo -e "${BLUE}Test 4: Nginx Configuration${NC}"

if [ -f "nginx/nginx-production.conf" ]; then
    log_info "Checking nginx configuration..."

    PROXY_COUNT=$(grep -c "add_header X-Served-By \$upstream_addr always" nginx/nginx-production.conf || echo "0")

    if [ "$PROXY_COUNT" -ge 3 ]; then
        log_success "X-Served-By header configured in ${PROXY_COUNT} proxied locations"
    else
        log_warning "X-Served-By header found in only ${PROXY_COUNT} locations (expected 3+)"
    fi

    if grep -q "location /health" nginx/nginx-production.conf && \
       ! grep -A5 "location /health" nginx/nginx-production.conf | grep -q "proxy_pass"; then
        log_success "/health endpoint is non-proxied (correct)"
    else
        log_warning "/health endpoint configuration may be incorrect"
    fi
else
    log_error "nginx-production.conf not found"
fi

echo ""

# Test 5: Documentation check
echo -e "${BLUE}Test 5: Documentation${NC}"

DOCS=(
    "docs/DEPLOYMENT_PROCEDURES.md"
    "docs/LOAD_BALANCING_TEST_CONFIGURATION.md"
)

for doc in "${DOCS[@]}"; do
    if [ -f "$doc" ]; then
        log_success "Documentation exists: $doc"

        if grep -q "X-Served-By" "$doc"; then
            log_success "  - Mentions X-Served-By header"
        fi

        if grep -q "proxied" "$doc" || grep -q "non-proxied" "$doc"; then
            log_success "  - Explains proxied vs non-proxied endpoints"
        fi
    else
        log_warning "Documentation missing: $doc"
    fi
done

echo ""
echo -e "${GREEN}=== Verification Complete ===${NC}"
echo -e "${BLUE}Next Steps:${NC}"
echo -e "1. Start services: docker compose up -d"
echo -e "2. Wait for ready: sleep 30"
echo -e "3. Run full test: bash scripts/test-load-balancing.sh http://localhost /api/health"
echo -e "4. Check report: cat deployment-results/loadbalancing-*.json"

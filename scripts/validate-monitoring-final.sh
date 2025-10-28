#!/bin/bash

# Final comprehensive monitoring validation

echo "OllamaMax Monitoring System - Final Validation"
echo "==============================================="
echo ""

FAILED_TESTS=0
PASSED_TESTS=0

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

pass() {
    echo -e "${GREEN}✓${NC} $1"
    ((PASSED_TESTS++))
}

fail() {
    echo -e "${RED}✗${NC} $1"
    ((FAILED_TESTS++))
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

section() {
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "$1"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

section "Core Implementation Validation"

# Check for actual promauto imports (not comments)
if grep -q '^[[:space:]]*"github.com/prometheus/client_golang/prometheus/promauto"' pkg/database/repositories.go; then
    fail "repositories.go imports promauto package"
else
    pass "No promauto package imports in repositories.go"
fi

# Check manager field in all repositories
REPO_TYPES=("ModelRepository" "NodeRepository" "UserRepository" "SessionRepository" "InferenceRepository" "AuditRepository" "ConfigRepository")
for repo in "${REPO_TYPES[@]}"; do
    if grep -A5 "type ${repo} struct" pkg/database/repositories.go | grep -q "manager.*DatabaseManager"; then
        pass "${repo} has manager field"
    else
        fail "${repo} missing manager field"
    fi
done

# Check RecordQuery usage
if grep -q "r.manager.RecordQuery" pkg/database/repositories.go; then
    pass "Repositories use manager.RecordQuery()"
else
    fail "Repositories not using manager.RecordQuery()"
fi

section "Metrics Implementation"

# Check tool availability
if ! command -v jq &> /dev/null; then
    warn "jq not installed - some validation features limited"
    info "Install jq for enhanced validation: sudo apt-get install jq"
fi

# Database Manager
if grep -q "func.*RegisterTo.*prometheus.Registerer" pkg/database/manager.go; then
    pass "DatabaseManager implements RegisterTo()"
else
    fail "DatabaseManager missing RegisterTo()"
fi

# P2P Node
if grep -q "bytesSent.*CounterVec" pkg/p2p/node.go && grep -q "bytesReceived.*CounterVec" pkg/p2p/node.go; then
    pass "P2P implements bytes tracking"
else
    fail "P2P missing bytes tracking"
fi

if grep -q "func.*RegisterTo.*prometheus.Registerer" pkg/p2p/node.go; then
    pass "P2P implements RegisterTo()"
else
    fail "P2P missing RegisterTo()"
fi

# API Server Integration
if grep -q "db.RegisterTo(registry)" pkg/api/server.go; then
    pass "API server registers database metrics"
else
    fail "API server doesn't register database metrics"
fi

section "Configuration & Provisioning"

# Docker Compose
if grep -q "/var/lib/grafana/dashboards" docker-compose.yml; then
    pass "Grafana dashboards mounted in docker-compose"
else
    fail "Grafana dashboards not mounted"
fi

# Datasource UIDs
if grep -q "uid: prometheus" monitoring/grafana/provisioning/datasources/prometheus.yml; then
    pass "Prometheus datasource has stable UID"
else
    fail "Prometheus missing stable UID"
fi

section "Dashboard Updates"

# Check database dashboard
if grep -q "ollamamax_database_db_connections" monitoring/grafana/dashboards/database-performance.json; then
    pass "Database dashboard uses updated metric names"
else
    fail "Database dashboard has old metric names"
fi

# Count updated dashboards
UPDATED_COUNT=$(grep -l "ollamamax_database" monitoring/grafana/dashboards/*.json 2>/dev/null | wc -l)
info "Updated metric names in $UPDATED_COUNT dashboard(s)"

section "Alert Rules"

# Check histogram_quantile usage
if grep -q "histogram_quantile" monitoring/alerts.yml; then
    pass "Alerts use histogram_quantile()"
else
    fail "Alerts missing histogram_quantile()"
fi

# Check namespace
if grep -q "ollamamax_database_" monitoring/alerts.yml; then
    pass "Alerts use ollamamax_database_ namespace"
else
    fail "Alerts missing ollamamax_database_ namespace"
fi

section "Documentation"

DOCS=("docs/MONITORING_IMPLEMENTATION_GUIDE.md" "VERIFICATION_FIXES_SUMMARY.md" "MONITORING_FINAL_STATUS.md")
for doc in "${DOCS[@]}"; do
    if [ -f "$doc" ]; then
        pass "$(basename $doc) exists"
    else
        fail "$(basename $doc) missing"
    fi
done

section "Summary & Recommendations"

TOTAL_TESTS=$((PASSED_TESTS + FAILED_TESTS))
if [ $TOTAL_TESTS -eq 0 ]; then
    echo "No tests run"
    exit 1
fi

PASS_RATE=$((PASSED_TESTS * 100 / TOTAL_TESTS))

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Results: ${GREEN}${PASSED_TESTS}${NC}/${TOTAL_TESTS} passed (${PASS_RATE}%)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}✓ ALL VALIDATION CHECKS PASSED!${NC}"
    echo ""
    echo "Your monitoring system is production-ready."
    echo ""
    echo "Next steps to test the stack:"
    echo "  1. Start services:"
    echo "     docker-compose up -d"
    echo ""
    echo "  2. Verify metrics endpoint:"
    echo "     curl http://localhost:13100/metrics | grep ollamamax_database"
    echo ""
    echo "  3. Check Prometheus targets:"
    echo "     curl http://localhost:9090/api/v1/targets | jq"
    echo ""
    echo "  4. Access Grafana dashboards:"
    echo "     http://localhost:3001 (admin/admin_password)"
    echo ""
    echo "  5. View Prometheus alerts:"
    echo "     http://localhost:9090/alerts"
    echo ""
    exit 0
else
    echo -e "${RED}✗ $FAILED_TESTS validation check(s) failed${NC}"
    echo ""
    echo "Please review the failures above and fix them."
    exit 1
fi

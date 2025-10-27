#!/bin/bash

# Verification script for Prometheus integration (Comment 1)
# This script validates that all required changes have been implemented

set -e

echo "============================================"
echo "Comment 1 Implementation Verification"
echo "============================================"
echo

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counters
PASSED=0
FAILED=0

# Helper function
check_file_exists() {
    if [ -f "$1" ]; then
        echo -e "${GREEN}✓${NC} File exists: $1"
        ((PASSED++))
        return 0
    else
        echo -e "${RED}✗${NC} File missing: $1"
        ((FAILED++))
        return 1
    fi
}

check_file_contains() {
    if grep -q "$2" "$1" 2>/dev/null; then
        echo -e "${GREEN}✓${NC} Found in $1: $2"
        ((PASSED++))
        return 0
    else
        echo -e "${RED}✗${NC} Not found in $1: $2"
        ((FAILED++))
        return 1
    fi
}

echo "Step 1: Checking file modifications..."
echo "----------------------------------------"

# Check server.go modifications
check_file_exists "pkg/api/server.go"
check_file_contains "pkg/api/server.go" "prometheus"
check_file_contains "pkg/api/server.go" "promhttp"
check_file_contains "pkg/api/server.go" "registry \*prometheus.Registry"
check_file_contains "pkg/api/server.go" "httpRequestsTotal"
check_file_contains "pkg/api/server.go" "httpRequestDuration"
check_file_contains "pkg/api/server.go" "httpRequestsInFlight"
check_file_contains "pkg/api/server.go" "prometheusMiddleware"
check_file_contains "pkg/api/server.go" "promhttp.HandlerFor"

echo

# Check handlers.go modifications
check_file_exists "pkg/api/handlers.go"
check_file_contains "pkg/api/handlers.go" "metricsJSONHandler"

echo
echo "Step 2: Checking new files..."
echo "----------------------------------------"

# Check test files
check_file_exists "pkg/api/prometheus_test.go"
check_file_exists "tests/prometheus-integration-test.go"

# Check documentation
check_file_exists "docs/PROMETHEUS_INTEGRATION.md"
check_file_exists "docs/COMMENT_1_IMPLEMENTATION.md"

echo
echo "Step 3: Checking Prometheus metrics implementation..."
echo "----------------------------------------"

# Check metric definitions
check_file_contains "pkg/api/server.go" "http_requests_total"
check_file_contains "pkg/api/server.go" "http_request_duration_seconds"
check_file_contains "pkg/api/server.go" "http_requests_in_flight"

# Check metric registration
check_file_contains "pkg/api/server.go" "registry.Register(httpRequestsTotal)"
check_file_contains "pkg/api/server.go" "registry.Register(httpRequestDuration)"
check_file_contains "pkg/api/server.go" "registry.Register(httpRequestsInFlight)"

echo
echo "Step 4: Checking endpoint configuration..."
echo "----------------------------------------"

# Check /metrics endpoint
check_file_contains "pkg/api/server.go" 'router.GET("/metrics"'

# Check /metrics.json endpoint
check_file_contains "pkg/api/server.go" 'router.GET("/metrics.json"'
check_file_contains "pkg/api/server.go" "metricsJSONHandler"

echo
echo "Step 5: Checking middleware integration..."
echo "----------------------------------------"

# Check middleware is registered
check_file_contains "pkg/api/server.go" "router.Use(s.prometheusMiddleware())"

# Check middleware implementation
check_file_contains "pkg/api/server.go" "httpRequestsInFlight.Inc()"
check_file_contains "pkg/api/server.go" "httpRequestsInFlight.Dec()"
check_file_contains "pkg/api/server.go" "httpRequestsTotal.WithLabelValues"
check_file_contains "pkg/api/server.go" "httpRequestDuration.WithLabelValues"

echo
echo "Step 6: Running integration test..."
echo "----------------------------------------"

if [ -f "tests/prometheus-integration-test.go" ]; then
    echo "Running Prometheus integration test..."
    if go run tests/prometheus-integration-test.go; then
        echo -e "${GREEN}✓${NC} Integration test passed"
        ((PASSED++))
    else
        echo -e "${RED}✗${NC} Integration test failed"
        ((FAILED++))
    fi
else
    echo -e "${YELLOW}⚠${NC} Integration test file not found, skipping..."
fi

echo
echo "Step 7: Checking documentation completeness..."
echo "----------------------------------------"

if [ -f "docs/PROMETHEUS_INTEGRATION.md" ]; then
    check_file_contains "docs/PROMETHEUS_INTEGRATION.md" "Prometheus Configuration"
    check_file_contains "docs/PROMETHEUS_INTEGRATION.md" "Example Queries"
    check_file_contains "docs/PROMETHEUS_INTEGRATION.md" "Grafana Dashboard"
    check_file_contains "docs/PROMETHEUS_INTEGRATION.md" "Testing"
    check_file_contains "docs/PROMETHEUS_INTEGRATION.md" "Security Considerations"
fi

echo
echo "Step 8: Checking backward compatibility..."
echo "----------------------------------------"

# Verify JSON metrics handler still exists
check_file_contains "pkg/api/handlers.go" "metricsJSONHandler"
check_file_contains "pkg/api/handlers.go" "s.db.Stats()"
check_file_contains "pkg/api/handlers.go" "timestamp"

echo
echo "============================================"
echo "Verification Summary"
echo "============================================"
echo -e "Passed: ${GREEN}${PASSED}${NC}"
echo -e "Failed: ${RED}${FAILED}${NC}"
echo

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All checks passed! Comment 1 implementation is complete.${NC}"
    echo
    echo "Next steps:"
    echo "  1. Review the implementation in pkg/api/server.go"
    echo "  2. Check documentation at docs/PROMETHEUS_INTEGRATION.md"
    echo "  3. Run integration test: go run tests/prometheus-integration-test.go"
    echo "  4. Configure Prometheus to scrape /metrics endpoint"
    echo "  5. Set up Grafana dashboards"
    echo
    exit 0
else
    echo -e "${RED}❌ Some checks failed. Please review the implementation.${NC}"
    echo
    exit 1
fi

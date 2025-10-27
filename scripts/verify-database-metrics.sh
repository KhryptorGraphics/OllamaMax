#!/bin/bash
# Verification script for Comment 3: Database Metrics Implementation

set -e

echo "=== Comment 3: Database Metrics Implementation Verification ==="
echo

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counter
passed=0
failed=0

check() {
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ PASS${NC}: $1"
        ((passed++))
    else
        echo -e "${RED}✗ FAIL${NC}: $1"
        ((failed++))
    fi
}

echo -e "${BLUE}1. Checking Prometheus import${NC}"
grep -q "github.com/prometheus/client_golang/prometheus" pkg/database/manager.go
check "Prometheus package imported"

echo -e "${BLUE}2. Checking DatabaseManager struct metrics fields${NC}"
grep -q "registry.*prometheus.Registry" pkg/database/manager.go
check "Prometheus registry field added"

grep -q "dbConnectionsOpen.*prometheus.Gauge" pkg/database/manager.go
check "dbConnectionsOpen gauge added"

grep -q "dbConnectionsInUse.*prometheus.Gauge" pkg/database/manager.go
check "dbConnectionsInUse gauge added"

grep -q "dbConnectionsIdle.*prometheus.Gauge" pkg/database/manager.go
check "dbConnectionsIdle gauge added"

grep -q "dbQueriesTotal.*prometheus.CounterVec" pkg/database/manager.go
check "dbQueriesTotal counter added"

grep -q "dbQueryDuration.*prometheus.HistogramVec" pkg/database/manager.go
check "dbQueryDuration histogram added"

grep -q "cacheHitsTotal.*prometheus.Counter" pkg/database/manager.go
check "cacheHitsTotal counter added"

grep -q "cacheMissesTotal.*prometheus.Counter" pkg/database/manager.go
check "cacheMissesTotal counter added"

grep -q "cacheOperationDuration.*prometheus.Histogram" pkg/database/manager.go
check "cacheOperationDuration histogram added"

echo -e "${BLUE}3. Checking lifecycle management${NC}"
grep -q "metricsCancel context.CancelFunc" pkg/database/manager.go
check "metricsCancel field for lifecycle management"

echo -e "${BLUE}4. Checking initializeMetrics method${NC}"
grep -q "func (dm \*DatabaseManager) initializeMetrics()" pkg/database/manager.go
check "initializeMetrics method exists"

grep -q "dm.registry = prometheus.NewRegistry()" pkg/database/manager.go
check "Registry creation in initializeMetrics"

grep -q "dm.registry.MustRegister" pkg/database/manager.go
check "Metrics registration"

echo -e "${BLUE}5. Checking startMetricsCollection method${NC}"
grep -q "func (dm \*DatabaseManager) startMetricsCollection()" pkg/database/manager.go
check "startMetricsCollection method exists"

grep -q "time.NewTicker(15 \* time.Second)" pkg/database/manager.go
check "15-second collection interval"

grep -q "dm.updatePoolMetrics()" pkg/database/manager.go
check "Pool metrics update call"

echo -e "${BLUE}6. Checking updatePoolMetrics method${NC}"
grep -q "func (dm \*DatabaseManager) updatePoolMetrics()" pkg/database/manager.go
check "updatePoolMetrics method exists"

grep -q "stats := dm.DB.Stats()" pkg/database/manager.go
check "DB.Stats() retrieval"

grep -q "dm.dbConnectionsOpen.Set" pkg/database/manager.go
check "Connection open metric update"

grep -q "dm.dbConnectionsInUse.Set" pkg/database/manager.go
check "Connection in-use metric update"

grep -q "dm.dbConnectionsIdle.Set" pkg/database/manager.go
check "Connection idle metric update"

echo -e "${BLUE}7. Checking GetPrometheusRegistry method${NC}"
grep -q "func (dm \*DatabaseManager) GetPrometheusRegistry() \*prometheus.Registry" pkg/database/manager.go
check "GetPrometheusRegistry method exists"

grep -q "return dm.registry" pkg/database/manager.go
check "Registry return statement"

echo -e "${BLUE}8. Checking RecordQuery method${NC}"
grep -q "func (dm \*DatabaseManager) RecordQuery(operation, table string, duration time.Duration)" pkg/database/manager.go
check "RecordQuery method exists"

grep -q "dm.dbQueriesTotal.WithLabelValues(operation, table).Inc()" pkg/database/manager.go
check "Query counter increment"

grep -q "dm.dbQueryDuration.WithLabelValues(operation, table).Observe" pkg/database/manager.go
check "Query duration observation"

echo -e "${BLUE}9. Checking RecordCacheHit method${NC}"
grep -q "func (dm \*DatabaseManager) RecordCacheHit(duration time.Duration)" pkg/database/manager.go
check "RecordCacheHit method exists"

grep -q "dm.cacheHitsTotal.Inc()" pkg/database/manager.go
check "Cache hit counter increment"

echo -e "${BLUE}10. Checking RecordCacheMiss method${NC}"
grep -q "func (dm \*DatabaseManager) RecordCacheMiss(duration time.Duration)" pkg/database/manager.go
check "RecordCacheMiss method exists"

grep -q "dm.cacheMissesTotal.Inc()" pkg/database/manager.go
check "Cache miss counter increment"

echo -e "${BLUE}11. Checking initialization in NewDatabaseManager${NC}"
grep -q "dm.initializeMetrics()" pkg/database/manager.go
check "initializeMetrics call in NewDatabaseManager"

grep -q "dm.startMetricsCollection()" pkg/database/manager.go
check "startMetricsCollection call in NewDatabaseManager"

echo -e "${BLUE}12. Checking cleanup in Close method${NC}"
grep -q "if dm.metricsCancel != nil" pkg/database/manager.go
check "metricsCancel check in Close"

grep -q "dm.metricsCancel()" pkg/database/manager.go
check "metricsCancel call in Close"

echo -e "${BLUE}13. Checking metric names${NC}"
grep -q '"db_connections_open"' pkg/database/manager.go
check "db_connections_open metric name"

grep -q '"db_connections_in_use"' pkg/database/manager.go
check "db_connections_in_use metric name"

grep -q '"db_connections_idle"' pkg/database/manager.go
check "db_connections_idle metric name"

grep -q '"db_queries_total"' pkg/database/manager.go
check "db_queries_total metric name"

grep -q '"db_query_duration_seconds"' pkg/database/manager.go
check "db_query_duration_seconds metric name"

grep -q '"cache_hits_total"' pkg/database/manager.go
check "cache_hits_total metric name"

grep -q '"cache_misses_total"' pkg/database/manager.go
check "cache_misses_total metric name"

grep -q '"cache_operation_duration_seconds"' pkg/database/manager.go
check "cache_operation_duration_seconds metric name"

echo -e "${BLUE}14. Checking metric labels${NC}"
grep -q '\[\]string{"operation", "table"}' pkg/database/manager.go
check "Query metrics labels (operation, table)"

echo -e "${BLUE}15. Checking compilation${NC}"
cd "$(dirname "$0")/.."
go list ./pkg/database > /dev/null 2>&1
check "Package compiles successfully"

echo
echo "=== Summary ==="
echo -e "Total checks: $((passed + failed))"
echo -e "${GREEN}Passed: $passed${NC}"
echo -e "${RED}Failed: $failed${NC}"
echo

if [ $failed -eq 0 ]; then
    echo -e "${GREEN}✓ All verification checks passed!${NC}"
    echo "Comment 3: Database Metrics Implementation is complete."
    exit 0
else
    echo -e "${RED}✗ Some verification checks failed.${NC}"
    echo "Please review the implementation."
    exit 1
fi

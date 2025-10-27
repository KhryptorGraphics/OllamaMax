# Database and Cache Metrics - Usage Examples

## Real-World Usage Examples

### Example 1: Monitoring Query Performance

**Scenario**: You want to monitor the performance of model retrieval operations.

**Steps**:
1. Make some requests to the API
2. Query the metrics endpoint
3. Analyze the results

```bash
# Generate some load
for i in {1..50}; do
  curl -s http://localhost:8080/api/v1/models > /dev/null
done

# View metrics
curl -s http://localhost:8080/metrics | grep 'db_query.*models'
```

**Expected Output**:
```
db_query_duration_seconds_bucket{operation="list",table="models",le="0.001"} 12
db_query_duration_seconds_bucket{operation="list",table="models",le="0.002"} 35
db_query_duration_seconds_bucket{operation="list",table="models",le="0.004"} 48
db_query_duration_seconds_bucket{operation="list",table="models",le="+Inf"} 50
db_query_duration_seconds_sum{operation="list",table="models"} 0.078
db_query_duration_seconds_count{operation="list",table="models"} 50
db_queries_total{operation="list",status="success",table="models"} 50
```

**Analysis**:
- Average latency: 0.078 / 50 = 1.56ms
- 96% of queries completed within 4ms (48/50)
- 100% success rate (no errors)

### Example 2: Cache Effectiveness Analysis

**Scenario**: Determine if the Redis cache is working effectively.

```bash
# First request (cache miss)
curl -s http://localhost:8080/api/v1/models/123e4567-e89b-12d3-a456-426614174000

# Second request (cache hit)
curl -s http://localhost:8080/api/v1/models/123e4567-e89b-12d3-a456-426614174000

# View cache metrics
curl -s http://localhost:8080/metrics | grep 'cache_.*models'
```

**Expected Output**:
```
cache_hits_total{cache_type="redis",table="models"} 1
cache_misses_total{cache_type="redis",table="models"} 1
cache_operation_duration_seconds_sum{cache_type="redis",operation="get",table="models"} 0.0023
cache_operation_duration_seconds_count{cache_type="redis",operation="get",table="models"} 2
cache_operation_duration_seconds_sum{cache_type="redis",operation="set",table="models"} 0.0008
cache_operation_duration_seconds_count{cache_type="redis",operation="set",table="models"} 1
```

**Analysis**:
- Hit rate: 1 / (1+1) = 50%
- Average GET latency: 1.15ms
- Average SET latency: 0.8ms
- Cache is working correctly

### Example 3: Error Rate Monitoring

**Scenario**: Monitor database errors during a load test.

```bash
# Run load test
ab -n 1000 -c 10 http://localhost:8080/api/v1/models

# Check error metrics
curl -s http://localhost:8080/metrics | grep 'db_queries_total.*error'
```

**Expected Output**:
```
db_queries_total{operation="list",status="error",table="models"} 5
db_queries_total{operation="list",status="success",table="models"} 995
```

**Analysis**:
- Error rate: 5 / 1000 = 0.5%
- System is stable under load
- Investigate the 5 errors in application logs

### Example 4: Performance Regression Detection

**Scenario**: Compare current performance with baseline.

```bash
# Export current metrics
curl -s http://localhost:8080/metrics > current_metrics.txt

# After code changes, export again
curl -s http://localhost:8080/metrics > new_metrics.txt

# Compare
diff current_metrics.txt new_metrics.txt | grep db_query_duration
```

**Example Diff**:
```diff
< db_query_duration_seconds_sum{operation="get",table="models"} 0.156
> db_query_duration_seconds_sum{operation="get",table="models"} 0.289
```

**Analysis**:
- Query time increased from ~3ms to ~6ms (85% slower)
- Performance regression detected
- Review recent code changes

### Example 5: Prometheus Query Examples

**Scenario**: Use Prometheus to analyze metrics over time.

#### Query 1: Average Latency Over Last Hour
```promql
rate(db_query_duration_seconds_sum[1h]) / rate(db_query_duration_seconds_count[1h])
```

**Result**: `0.0032` (3.2ms average)

#### Query 2: P95 Latency by Operation
```promql
histogram_quantile(0.95,
  sum(rate(db_query_duration_seconds_bucket[5m])) by (operation, le)
)
```

**Result**:
```
{operation="get"} 0.008
{operation="list"} 0.015
{operation="create"} 0.012
{operation="update"} 0.010
{operation="delete"} 0.006
```

**Analysis**: List operations are slowest at P95

#### Query 3: Cache Hit Rate Over Time
```promql
(
  rate(cache_hits_total[5m])
  /
  (rate(cache_hits_total[5m]) + rate(cache_misses_total[5m]))
) * 100
```

**Result**: `78.5%` (good cache effectiveness)

### Example 6: Grafana Dashboard Setup

**Scenario**: Create a comprehensive monitoring dashboard.

#### Panel 1: Query Rate
```json
{
  "title": "Database Queries per Second",
  "targets": [
    {
      "expr": "sum(rate(db_queries_total[5m])) by (operation)",
      "legendFormat": "{{operation}}"
    }
  ],
  "type": "graph"
}
```

#### Panel 2: Error Rate
```json
{
  "title": "Query Error Rate",
  "targets": [
    {
      "expr": "rate(db_queries_total{status=\"error\"}[5m]) / rate(db_queries_total[5m])",
      "legendFormat": "Error Rate"
    }
  ],
  "type": "graph",
  "alert": {
    "conditions": [
      {
        "evaluator": {
          "type": "gt",
          "params": [0.05]
        }
      }
    ]
  }
}
```

#### Panel 3: Cache Performance
```json
{
  "title": "Cache Hit Rate",
  "targets": [
    {
      "expr": "(rate(cache_hits_total[5m]) / (rate(cache_hits_total[5m]) + rate(cache_misses_total[5m]))) * 100",
      "legendFormat": "Hit Rate %"
    }
  ],
  "type": "gauge",
  "min": 0,
  "max": 100,
  "thresholds": [
    {"value": 70, "color": "red"},
    {"value": 85, "color": "yellow"},
    {"value": 95, "color": "green"}
  ]
}
```

### Example 7: Alert Configuration

**Scenario**: Set up proactive monitoring alerts.

#### Alert 1: High Latency
```yaml
groups:
  - name: database_performance
    interval: 30s
    rules:
      - alert: HighDatabaseLatency
        expr: |
          histogram_quantile(0.95,
            rate(db_query_duration_seconds_bucket[5m])
          ) > 0.1
        for: 5m
        labels:
          severity: warning
          component: database
        annotations:
          summary: "Database queries are slow"
          description: "P95 latency is {{ $value }}s for {{ $labels.operation }} on {{ $labels.table }}"
```

#### Alert 2: Low Cache Hit Rate
```yaml
      - alert: LowCacheHitRate
        expr: |
          (
            rate(cache_hits_total[10m])
            /
            (rate(cache_hits_total[10m]) + rate(cache_misses_total[10m]))
          ) < 0.7
        for: 10m
        labels:
          severity: warning
          component: cache
        annotations:
          summary: "Cache is ineffective"
          description: "Cache hit rate is {{ $value | humanizePercentage }} for {{ $labels.table }}"
```

#### Alert 3: High Error Rate
```yaml
      - alert: DatabaseErrors
        expr: |
          rate(db_queries_total{status="error"}[5m]) > 10
        for: 2m
        labels:
          severity: critical
          component: database
        annotations:
          summary: "High database error rate"
          description: "{{ $value }} errors/second on {{ $labels.table }}"
```

### Example 8: Load Testing Analysis

**Scenario**: Analyze metrics during load testing.

```bash
# Start monitoring in one terminal
watch -n 1 "curl -s http://localhost:8080/metrics | grep -E '(db_queries_total|cache_hits)'"

# Run load test in another terminal
ab -n 10000 -c 50 -t 60 http://localhost:8080/api/v1/models
```

**Metrics During Test**:
```
# T=0s (baseline)
db_queries_total{operation="list",status="success",table="models"} 100
cache_hits_total{cache_type="redis",table="models"} 50

# T=30s (under load)
db_queries_total{operation="list",status="success",table="models"} 5150
cache_hits_total{cache_type="redis",table="models"} 3875

# T=60s (peak load)
db_queries_total{operation="list",status="success",table="models"} 10100
cache_hits_total{cache_type="redis",table="models"} 7575
```

**Analysis**:
- Query rate: (10100-100)/60 = 166 qps
- Cache hit rate: (7575-50)/(10100-100) = 75.2%
- No errors observed (good stability)

### Example 9: Capacity Planning

**Scenario**: Determine system capacity limits.

#### Step 1: Collect Baseline Metrics
```bash
# Current load
curl -s http://localhost:8080/metrics | grep 'db_queries_total'
```

**Result**: 50 qps average

#### Step 2: Calculate P95 Latency
```promql
histogram_quantile(0.95, rate(db_query_duration_seconds_bucket[5m]))
```

**Result**: 8ms at 50 qps

#### Step 3: Stress Test
```bash
# Double the load
ab -n 20000 -c 100 http://localhost:8080/api/v1/models
```

#### Step 4: Analyze Impact
```promql
histogram_quantile(0.95, rate(db_query_duration_seconds_bucket[5m]))
```

**Result**: 15ms at 100 qps

**Analysis**:
- Latency increased 87.5% (8ms → 15ms)
- System can handle 2x load with acceptable degradation
- Plan for horizontal scaling at 150 qps

### Example 10: Debugging Slow Queries

**Scenario**: Identify and fix slow operations.

#### Step 1: Find Slow Operations
```promql
topk(5,
  histogram_quantile(0.99,
    sum(rate(db_query_duration_seconds_bucket[5m])) by (operation, table, le)
  )
)
```

**Result**:
```
{operation="list", table="models"} 0.045
{operation="get", table="users"} 0.032
{operation="list", table="nodes"} 0.028
{operation="create", table="models"} 0.025
{operation="update", table="users"} 0.022
```

**Analysis**: List operations on models table are slowest

#### Step 2: Check Query Count
```promql
rate(db_queries_total{operation="list", table="models"}[5m])
```

**Result**: 45 qps (high frequency)

#### Step 3: Investigate Cache
```promql
rate(cache_misses_total{table="models"}[5m])
/
(rate(cache_hits_total{table="models"}[5m]) + rate(cache_misses_total{table="models"}[5m]))
```

**Result**: 65% miss rate (low cache effectiveness)

**Action Items**:
1. Increase cache TTL for models
2. Add database index on frequently filtered columns
3. Consider query result pagination

### Example 11: Multi-Environment Comparison

**Scenario**: Compare metrics across dev, staging, and production.

```bash
# Export from each environment
for env in dev staging prod; do
  curl -s "https://${env}.example.com/metrics" > "${env}_metrics.txt"
done

# Compare query rates
for env in dev staging prod; do
  echo "=== $env ==="
  grep "db_queries_total" "${env}_metrics.txt" | grep "status=\"success\"" | head -3
done
```

**Output**:
```
=== dev ===
db_queries_total{operation="list",status="success",table="models"} 523

=== staging ===
db_queries_total{operation="list",status="success",table="models"} 4301

=== prod ===
db_queries_total{operation="list",status="success",table="models"} 45234
```

**Analysis**: Production has 10x more traffic than staging

### Example 12: Custom Aggregation

**Scenario**: Calculate custom business metrics from raw metrics.

#### Total Database Time Spent Per Second
```promql
sum(rate(db_query_duration_seconds_sum[5m]))
```

**Result**: `0.15` (150ms/second spent in database)

#### Average Queries Per Request
```promql
rate(db_queries_total[5m]) / rate(http_requests_total[5m])
```

**Result**: `2.3` (each API request makes ~2 DB queries)

#### Cache Efficiency Score (0-100)
```promql
(
  (rate(cache_hits_total[5m]) * 1.0)
  /
  (rate(cache_hits_total[5m]) + rate(cache_misses_total[5m]))
) * 100
```

**Result**: `76.5` (good cache efficiency)

## Testing Metrics Implementation

### Unit Test Example
```go
func TestQueryMetrics(t *testing.T) {
    // Reset metrics
    dbQueriesTotal.Reset()

    // Execute query with metrics
    err := recordQueryMetrics("get", "models", func() error {
        return nil // Simulated query
    })

    // Verify metrics
    metric := testutil.CollectAndCount(dbQueriesTotal)
    assert.Equal(t, 1, metric)
}
```

### Integration Test Example
```bash
# Start test server
go test -v -run TestMetricsEndpoint

# Make requests
curl http://localhost:8080/api/v1/models

# Verify metrics are exposed
metrics=$(curl -s http://localhost:8080/metrics)
echo "$metrics" | grep -q "db_queries_total" && echo "✓ DB metrics present"
echo "$metrics" | grep -q "cache_hits_total" && echo "✓ Cache metrics present"
```

## Troubleshooting Common Issues

### Issue 1: Metrics Not Appearing
```bash
# Check if metrics handler is registered
curl -I http://localhost:8080/metrics

# Should return: HTTP/1.1 200 OK
# If 404, check server initialization
```

### Issue 2: Incorrect Values
```bash
# Verify counter increments
before=$(curl -s http://localhost:8080/metrics | grep "db_queries_total.*success" | grep -oP '\d+$')
curl http://localhost:8080/api/v1/models > /dev/null
after=$(curl -s http://localhost:8080/metrics | grep "db_queries_total.*success" | grep -oP '\d+$')

echo "Queries before: $before"
echo "Queries after: $after"
echo "Increment: $((after - before))"
```

### Issue 3: High Cardinality
```bash
# Check unique label combinations
curl -s http://localhost:8080/metrics | grep "^db_query" | wc -l

# Should be reasonable (<1000 time series)
# If too high, review label values
```

## Best Practices Checklist

- [x] Monitor P95/P99 latencies, not just averages
- [x] Set up alerts for error rates
- [x] Track cache hit rates
- [x] Use recording rules for expensive queries
- [x] Create dashboards for each service component
- [x] Document baseline performance metrics
- [x] Review metrics weekly in team meetings
- [x] Set up alerting escalation policies
- [x] Test metrics in staging before production
- [x] Keep retention period appropriate (30-90 days)

## Additional Resources

- **Prometheus Best Practices**: https://prometheus.io/docs/practices/naming/
- **Grafana Dashboards**: https://grafana.com/grafana/dashboards/
- **PromQL Tutorial**: https://prometheus.io/docs/prometheus/latest/querying/basics/

## Support

For questions or issues with metrics implementation:
1. Review `/home/kp/OllamaMax/docs/COMMENT_4_IMPLEMENTATION.md`
2. Check `/home/kp/OllamaMax/docs/METRICS_QUICK_REFERENCE.md`
3. Consult application logs

## Version Information
- **Implementation Date**: 2025-10-27
- **Examples Version**: 1.0
- **Tested With**: Prometheus 2.40+, Grafana 9.0+

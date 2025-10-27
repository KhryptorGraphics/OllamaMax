# Database and Cache Metrics Quick Reference

## Quick Start

### View All Metrics
```bash
curl http://localhost:8080/metrics | grep -E "(db_query|cache_)"
```

### View Specific Metric
```bash
curl http://localhost:8080/metrics | grep "db_query_duration_seconds"
```

## Available Metrics

### Database Metrics

#### db_query_duration_seconds (Histogram)
Measures query execution time in seconds.

**Labels:**
- `operation`: get, list, create, update, delete
- `table`: models, users, nodes, model_replicas, audit_log_entries

**Buckets:** 0.001s, 0.002s, 0.004s, 0.008s, 0.016s, 0.032s, 0.064s, 0.128s, 0.256s, 0.512s, 1.024s, +Inf

**Example:**
```
db_query_duration_seconds_bucket{operation="get",table="models",le="0.001"} 45
db_query_duration_seconds_bucket{operation="get",table="models",le="0.002"} 48
db_query_duration_seconds_sum{operation="get",table="models"} 0.156
db_query_duration_seconds_count{operation="get",table="models"} 50
```

#### db_queries_total (Counter)
Total number of database queries executed.

**Labels:**
- `operation`: get, list, create, update, delete
- `table`: models, users, nodes, model_replicas, audit_log_entries
- `status`: success, error

**Example:**
```
db_queries_total{operation="get",status="success",table="models"} 50
db_queries_total{operation="get",status="error",table="models"} 2
```

### Cache Metrics

#### cache_hits_total (Counter)
Number of successful cache retrievals.

**Labels:**
- `cache_type`: redis
- `table`: models

**Example:**
```
cache_hits_total{cache_type="redis",table="models"} 35
```

#### cache_misses_total (Counter)
Number of failed cache retrievals (not found in cache).

**Labels:**
- `cache_type`: redis
- `table`: models

**Example:**
```
cache_misses_total{cache_type="redis",table="models"} 15
```

#### cache_operation_duration_seconds (Histogram)
Measures cache operation execution time in seconds.

**Labels:**
- `operation`: get, set, delete
- `cache_type`: redis
- `table`: models

**Buckets:** 0.0001s, 0.0002s, 0.0004s, 0.0008s, 0.0016s, 0.0032s, 0.0064s, 0.0128s, 0.0256s, 0.0512s, 0.1024s, +Inf

**Example:**
```
cache_operation_duration_seconds_bucket{cache_type="redis",operation="get",table="models",le="0.001"} 48
cache_operation_duration_seconds_sum{cache_type="redis",operation="get",table="models"} 0.012
cache_operation_duration_seconds_count{cache_type="redis",operation="get",table="models"} 50
```

## Common PromQL Queries

### Query Performance

#### Average Query Duration
```promql
rate(db_query_duration_seconds_sum[5m]) / rate(db_query_duration_seconds_count[5m])
```

#### Average Query Duration by Table
```promql
rate(db_query_duration_seconds_sum[5m]) by (table) / rate(db_query_duration_seconds_count[5m]) by (table)
```

#### Average Query Duration by Operation
```promql
rate(db_query_duration_seconds_sum[5m]) by (operation) / rate(db_query_duration_seconds_count[5m]) by (operation)
```

#### 95th Percentile Query Latency
```promql
histogram_quantile(0.95, rate(db_query_duration_seconds_bucket[5m]))
```

#### 99th Percentile Query Latency
```promql
histogram_quantile(0.99, rate(db_query_duration_seconds_bucket[5m]))
```

#### Top 5 Slowest Operations
```promql
topk(5, histogram_quantile(0.99, sum(rate(db_query_duration_seconds_bucket[5m])) by (operation, table, le)))
```

### Query Throughput

#### Queries per Second
```promql
rate(db_queries_total[5m])
```

#### Queries per Second by Table
```promql
sum(rate(db_queries_total[5m])) by (table)
```

#### Queries per Second by Operation
```promql
sum(rate(db_queries_total[5m])) by (operation)
```

### Error Monitoring

#### Query Error Rate
```promql
rate(db_queries_total{status="error"}[5m]) / rate(db_queries_total[5m])
```

#### Query Error Rate by Table
```promql
rate(db_queries_total{status="error"}[5m]) by (table) / rate(db_queries_total[5m]) by (table)
```

#### Total Errors in Last Hour
```promql
increase(db_queries_total{status="error"}[1h])
```

### Cache Performance

#### Cache Hit Rate
```promql
rate(cache_hits_total[5m]) / (rate(cache_hits_total[5m]) + rate(cache_misses_total[5m]))
```

#### Cache Hit Rate Percentage
```promql
(rate(cache_hits_total[5m]) / (rate(cache_hits_total[5m]) + rate(cache_misses_total[5m]))) * 100
```

#### Cache Miss Rate
```promql
rate(cache_misses_total[5m]) / (rate(cache_hits_total[5m]) + rate(cache_misses_total[5m]))
```

#### Average Cache Operation Latency
```promql
rate(cache_operation_duration_seconds_sum[5m]) / rate(cache_operation_duration_seconds_count[5m])
```

#### Average Cache Latency by Operation
```promql
rate(cache_operation_duration_seconds_sum[5m]) by (operation) / rate(cache_operation_duration_seconds_count[5m]) by (operation)
```

#### Cache Operations per Second
```promql
sum(rate(cache_operation_duration_seconds_count[5m])) by (operation)
```

## Grafana Panel Examples

### Query Duration Heatmap
```promql
sum(rate(db_query_duration_seconds_bucket[5m])) by (le, operation)
```
**Visualization:** Heatmap
**X-axis:** Time
**Y-axis:** Latency buckets

### Cache Hit Rate Gauge
```promql
(sum(rate(cache_hits_total[5m])) / (sum(rate(cache_hits_total[5m])) + sum(rate(cache_misses_total[5m])))) * 100
```
**Visualization:** Gauge
**Min:** 0
**Max:** 100
**Thresholds:** Red < 70%, Yellow < 85%, Green >= 85%

### Query Throughput Graph
```promql
sum(rate(db_queries_total[5m])) by (operation)
```
**Visualization:** Time series graph
**Legend:** {{operation}}

### Top Slow Queries Table
```promql
topk(10, histogram_quantile(0.99, sum(rate(db_query_duration_seconds_bucket[5m])) by (operation, table, le)))
```
**Visualization:** Table
**Columns:** Operation, Table, P99 Latency

### Error Rate Graph
```promql
rate(db_queries_total{status="error"}[5m])
```
**Visualization:** Time series graph
**Y-axis:** Errors per second

## Alert Rule Examples

### High Query Error Rate
```yaml
- alert: HighDatabaseErrorRate
  expr: rate(db_queries_total{status="error"}[5m]) / rate(db_queries_total[5m]) > 0.05
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Database error rate above 5%"
    description: "{{ $labels.table }} has {{ $value | humanizePercentage }} error rate"
```

### Slow Database Queries
```yaml
- alert: SlowDatabaseQueries
  expr: histogram_quantile(0.95, rate(db_query_duration_seconds_bucket[5m])) > 1
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "95th percentile query latency above 1s"
    description: "{{ $labels.operation }} on {{ $labels.table }} is slow: {{ $value }}s"
```

### Very Slow Database Queries
```yaml
- alert: VerySlowDatabaseQueries
  expr: histogram_quantile(0.99, rate(db_query_duration_seconds_bucket[5m])) > 5
  for: 2m
  labels:
    severity: critical
  annotations:
    summary: "99th percentile query latency above 5s"
    description: "{{ $labels.operation }} on {{ $labels.table }} is very slow: {{ $value }}s"
```

### Low Cache Hit Rate
```yaml
- alert: LowCacheHitRate
  expr: (rate(cache_hits_total[5m]) / (rate(cache_hits_total[5m]) + rate(cache_misses_total[5m]))) < 0.7
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "Cache hit rate below 70%"
    description: "{{ $labels.table }} cache hit rate is {{ $value | humanizePercentage }}"
```

### Very Low Cache Hit Rate
```yaml
- alert: VeryLowCacheHitRate
  expr: (rate(cache_hits_total[5m]) / (rate(cache_hits_total[5m]) + rate(cache_misses_total[5m]))) < 0.5
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Cache hit rate below 50%"
    description: "{{ $labels.table }} cache is ineffective: {{ $value | humanizePercentage }}"
```

### High Query Volume
```yaml
- alert: HighQueryVolume
  expr: rate(db_queries_total[5m]) > 1000
  for: 10m
  labels:
    severity: info
  annotations:
    summary: "Query volume above 1000 qps"
    description: "Current rate: {{ $value }} queries/second"
```

## Dashboard Layout Suggestion

### Row 1: Overview
- **Total QPS** (Stat panel)
- **Average Latency** (Stat panel)
- **Error Rate** (Stat panel)
- **Cache Hit Rate** (Gauge)

### Row 2: Query Performance
- **Query Latency Heatmap** (Heatmap)
- **P95/P99 Latency** (Time series)

### Row 3: Throughput
- **Queries by Operation** (Time series)
- **Queries by Table** (Time series)

### Row 4: Cache Performance
- **Cache Hit/Miss Rate** (Time series)
- **Cache Operation Latency** (Time series)

### Row 5: Errors
- **Error Rate by Table** (Time series)
- **Top Errors** (Table)

## Common Analysis Scenarios

### 1. Identify Slow Queries
```promql
# Find operations with P99 > 100ms
histogram_quantile(0.99, rate(db_query_duration_seconds_bucket[5m])) > 0.1
```

### 2. Find Most Frequent Operations
```promql
# Top 5 operations by query count
topk(5, rate(db_queries_total[5m]))
```

### 3. Analyze Cache Effectiveness
```promql
# Cache hit rate for each table
(rate(cache_hits_total[5m]) / (rate(cache_hits_total[5m]) + rate(cache_misses_total[5m]))) by (table)
```

### 4. Detect Performance Degradation
```promql
# Compare current vs 1 hour ago
rate(db_query_duration_seconds_sum[5m]) / rate(db_query_duration_seconds_count[5m])
/
rate(db_query_duration_seconds_sum[5m] offset 1h) / rate(db_query_duration_seconds_count[5m] offset 1h)
```

### 5. Calculate Total Database Time
```promql
# Total seconds spent in database queries per second
sum(rate(db_query_duration_seconds_sum[5m]))
```

## Exporting Metrics

### Prometheus Configuration
```yaml
scrape_configs:
  - job_name: 'ollamamax'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 15s
```

### cURL Export
```bash
# Export all metrics
curl -s http://localhost:8080/metrics > metrics.txt

# Export database metrics only
curl -s http://localhost:8080/metrics | grep "^db_" > db_metrics.txt

# Export cache metrics only
curl -s http://localhost:8080/metrics | grep "^cache_" > cache_metrics.txt
```

## Troubleshooting

### No Metrics Appearing
1. Check server is running: `curl http://localhost:8080/health`
2. Verify metrics endpoint: `curl http://localhost:8080/metrics`
3. Check logs for errors

### Metrics Not Updating
1. Trigger database operations
2. Wait for scrape interval (default 15s)
3. Check Prometheus targets page

### High Latency Values
1. Check database connection pool
2. Review query plans
3. Analyze cache hit rates
4. Check for lock contention

### Low Cache Hit Rate
1. Review cache TTL settings
2. Check Redis memory usage
3. Analyze access patterns
4. Consider increasing cache size

## Best Practices

1. **Set appropriate scrape intervals**: 15-30s for production
2. **Use recording rules**: Pre-aggregate expensive queries
3. **Set retention policies**: Balance storage vs. historical data
4. **Create dashboards**: Visualize key metrics
5. **Configure alerts**: Proactive monitoring
6. **Regular reviews**: Weekly metric analysis
7. **Document baselines**: Know normal behavior

## Related Documentation

- `/home/kp/OllamaMax/docs/COMMENT_4_IMPLEMENTATION.md` - Full implementation details
- `/home/kp/OllamaMax/COMMENT_4_COMPLETE.md` - Implementation summary

## Support

For issues or questions:
1. Check Prometheus documentation: https://prometheus.io/docs/
2. Review Grafana guides: https://grafana.com/docs/
3. Check application logs

## Version
- **Implementation Date**: 2025-10-27
- **Metrics Version**: 1.0
- **Compatible Prometheus Version**: 2.0+

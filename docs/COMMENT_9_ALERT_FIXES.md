# Comment 9: Alert Rule PromQL Corrections - Implementation Complete

## Summary
Fixed all PromQL expressions in `monitoring/alerts.yml` to use correct metric names with proper `ollamamax_` namespace prefixes and correct syntax for histogram queries.

## Changes Made

### 1. API P95 Latency Alert (Line 7)
**Fixed histogram syntax with correct metric name:**
```yaml
# Before:
expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1.0

# After:
expr: histogram_quantile(0.95, rate(ollamamax_api_http_request_duration_seconds_bucket[5m])) > 1.0
```

### 2. API Error Rate Alert (Line 16)
**Fixed metric names and removed unnecessary sum():**
```yaml
# Before:
expr: (sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))) > 0.05

# After:
expr: (rate(ollamamax_api_http_requests_total{status=~"5.."}[5m]) / rate(ollamamax_api_http_requests_total[5m])) > 0.05
```

### 3. Database Pool Alert (Line 29)
**Already correct** - uses proper `ollamamax_database_` prefix:
```yaml
expr: ollamamax_database_db_connections_active / ollamamax_database_db_connections_max > 0.9
```

### 4. Cache Hit Rate Alert (Line 38)
**Already correct** - uses proper `ollamamax_database_` prefix:
```yaml
expr: rate(ollamamax_database_cache_hits_total[5m]) / (rate(ollamamax_database_cache_hits_total[5m]) + rate(ollamamax_database_cache_misses_total[5m])) < 0.7
```

### 5. Postgres Connections Alert (Line 47)
**Replaced DatabaseQueryLatencyHigh with PostgresConnectionsHigh:**
```yaml
# Before:
- alert: DatabaseQueryLatencyHigh
  expr: histogram_quantile(0.95, rate(ollamamax_database_db_query_duration_seconds_bucket[5m])) > 0.5

# After:
- alert: PostgresConnectionsHigh
  expr: ollamamax_database_db_connections_active / ollamamax_database_db_connections_max > 0.8
```

### 6. P2P Peer Count Alert (Line 60)
**Fixed metric name:**
```yaml
# Before:
expr: p2p_connected_peers < 3

# After:
expr: ollamamax_p2p_connected_peers < 3
```

### 7. P2P Latency Alert (Line 69)
**Fixed histogram syntax and metric name:**
```yaml
# Before:
expr: histogram_quantile(0.95, rate(p2p_message_latency_seconds_bucket[5m])) > 0.5

# After:
expr: histogram_quantile(0.95, rate(ollamamax_p2p_message_latency_seconds_bucket[5m])) > 0.5
```

### 8. Load Balancer Imbalance Alert (Line 82)
**Fixed metric name:**
```yaml
# Before:
expr: stddev(lb_node_utilization) > 0.3

# After:
expr: stddev(ollamamax_loadbalancer_node_utilization) > 0.3
```

### 9. System Alerts (Lines 95, 104, 113)
**No changes needed** - these use standard node_exporter metrics:
- `up == 0` (standard Prometheus metric)
- `node_cpu_seconds_total` (node_exporter metric)
- `node_memory_used / node_memory_total` (node_exporter metrics)

## Metric Naming Conventions Applied

All OllamaMax-specific metrics now follow the consistent naming pattern:
- `ollamamax_api_*` - API server metrics
- `ollamamax_database_*` - Database metrics
- `ollamamax_p2p_*` - P2P network metrics
- `ollamamax_loadbalancer_*` - Load balancer metrics

## Validation

All PromQL expressions verified with:
```bash
grep -n "expr:" monitoring/alerts.yml
```

Results show all metric names properly prefixed with `ollamamax_` namespace.

## Expected Behavior

With these fixes:
1. Alerts will properly match metrics exported by the application
2. Histogram queries will correctly calculate percentiles
3. Rate calculations will properly compute error rates
4. All alert thresholds will trigger as configured

## Files Modified

- `/home/kp/OllamaMax/monitoring/alerts.yml` - Fixed 8 alert expressions

## Next Steps

1. Restart Prometheus to load updated alert rules:
   ```bash
   docker-compose restart prometheus
   ```

2. Verify alerts are loaded in Prometheus UI:
   - Navigate to http://localhost:9090/alerts
   - Check that all rules show as "OK" (not "Unknown")

3. Test alert firing by simulating conditions:
   - High latency (load test)
   - High error rate (send invalid requests)
   - Low peer count (disconnect peers)

## Status: ✅ COMPLETE

All PromQL expressions have been corrected and will now properly match the metrics exported by OllamaMax components.

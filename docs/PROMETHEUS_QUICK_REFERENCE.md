# Prometheus Integration - Quick Reference Card

## Endpoints

| Endpoint | Format | Purpose | Auth |
|----------|--------|---------|------|
| `/metrics` | Prometheus | Monitoring/scraping | No |
| `/metrics.json` | JSON | Legacy/debugging | No |

## Metrics

### http_requests_total
```promql
http_requests_total{method="GET",endpoint="/api/v1/health",status="200"}
```
- **Type**: Counter
- **Labels**: method, endpoint, status
- **Use**: Track total requests by endpoint and status

### http_request_duration_seconds
```promql
http_request_duration_seconds_bucket{method="GET",endpoint="/api/v1/health",status="200",le="0.01"}
```
- **Type**: Histogram
- **Labels**: method, endpoint, status
- **Buckets**: 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10
- **Use**: Measure request latency and percentiles

### http_requests_in_flight
```promql
http_requests_in_flight
```
- **Type**: Gauge
- **Labels**: None
- **Use**: Monitor current server load

## Common Queries

### Request Rate
```promql
# Requests per second
rate(http_requests_total[5m])

# By endpoint
sum(rate(http_requests_total[5m])) by (endpoint)

# By status code
sum(rate(http_requests_total[5m])) by (status)
```

### Latency
```promql
# 50th percentile (median)
histogram_quantile(0.50, rate(http_request_duration_seconds_bucket[5m]))

# 95th percentile
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# 99th percentile
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))

# Average latency
rate(http_request_duration_seconds_sum[5m]) / rate(http_request_duration_seconds_count[5m])
```

### Error Rate
```promql
# 4xx errors
sum(rate(http_requests_total{status=~"4.."}[5m]))

# 5xx errors
sum(rate(http_requests_total{status=~"5.."}[5m]))

# Error percentage
100 * sum(rate(http_requests_total{status=~"4..|5.."}[5m])) / sum(rate(http_requests_total[5m]))
```

### Traffic Analysis
```promql
# Top 10 endpoints by traffic
topk(10, sum(rate(http_requests_total[5m])) by (endpoint))

# Top 10 slowest endpoints
topk(10, histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) by (endpoint))

# Current load
http_requests_in_flight

# Peak requests (max over 1h)
max_over_time(http_requests_in_flight[1h])
```

## Prometheus Configuration

### Basic Setup
```yaml
scrape_configs:
  - job_name: 'ollamamax'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: /metrics
    scrape_interval: 15s
```

### With Authentication
```yaml
scrape_configs:
  - job_name: 'ollamamax'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: /metrics
    scrape_interval: 15s
    basic_auth:
      username: 'prometheus'
      password: 'secure_password'
```

### Service Discovery (Kubernetes)
```yaml
scrape_configs:
  - job_name: 'ollamamax'
    kubernetes_sd_configs:
      - role: pod
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app]
        action: keep
        regex: ollamamax
      - source_labels: [__meta_kubernetes_pod_container_port_number]
        action: keep
        regex: '8080'
```

## Alert Rules

### High Error Rate
```yaml
groups:
  - name: ollamamax
    rules:
      - alert: HighErrorRate
        expr: |
          100 * sum(rate(http_requests_total{status=~"5.."}[5m]))
          / sum(rate(http_requests_total[5m])) > 5
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate ({{ $value }}%)"
```

### High Latency
```yaml
      - alert: HighLatency
        expr: |
          histogram_quantile(0.95,
            rate(http_request_duration_seconds_bucket[5m])
          ) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "95th percentile latency > 1s"
```

### High Load
```yaml
      - alert: HighLoad
        expr: http_requests_in_flight > 100
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High number of concurrent requests"
```

## Grafana Dashboard Panels

### 1. Request Rate Graph
```json
{
  "targets": [
    {
      "expr": "sum(rate(http_requests_total[5m])) by (endpoint)"
    }
  ],
  "visualization": "timeseries"
}
```

### 2. Latency Heatmap
```json
{
  "targets": [
    {
      "expr": "rate(http_request_duration_seconds_bucket[5m])"
    }
  ],
  "visualization": "heatmap"
}
```

### 3. Error Rate
```json
{
  "targets": [
    {
      "expr": "100 * sum(rate(http_requests_total{status=~\"4..|5..\"}[5m])) / sum(rate(http_requests_total[5m]))"
    }
  ],
  "visualization": "stat",
  "thresholds": {
    "warning": 1,
    "critical": 5
  }
}
```

### 4. Active Requests Gauge
```json
{
  "targets": [
    {
      "expr": "http_requests_in_flight"
    }
  ],
  "visualization": "gauge",
  "max": 200
}
```

## Testing Commands

### Check Metrics Endpoint
```bash
curl http://localhost:8080/metrics
```

### Check JSON Endpoint
```bash
curl http://localhost:8080/metrics.json | jq
```

### Generate Test Traffic
```bash
# Simple load test
for i in {1..1000}; do
  curl -s http://localhost:8080/health > /dev/null
done

# With ab (Apache Bench)
ab -n 1000 -c 10 http://localhost:8080/health

# With wrk
wrk -t4 -c100 -d30s http://localhost:8080/health
```

### Query Prometheus
```bash
# Using Prometheus API
curl 'http://prometheus:9090/api/v1/query?query=http_requests_total'

# Get current in-flight requests
curl 'http://prometheus:9090/api/v1/query?query=http_requests_in_flight'
```

## Troubleshooting

### Issue: Metrics not appearing
```bash
# Check if endpoint is accessible
curl -I http://localhost:8080/metrics

# Check Prometheus targets
curl http://prometheus:9090/targets

# Check server logs
tail -f logs/ollamamax.log | grep metrics
```

### Issue: High cardinality
```bash
# Count unique metric series
curl -s http://localhost:8080/metrics | grep http_requests_total | wc -l

# Should be: endpoints × methods × status_codes
# Example: 50 endpoints × 4 methods × 10 statuses = 2000 series
```

### Issue: Incorrect values
```bash
# Check middleware is loaded
curl -s http://localhost:8080/metrics | grep http_requests_total

# Verify no errors in registration
grep "Failed to register" logs/ollamamax.log
```

## File Locations

| File | Purpose |
|------|---------|
| `pkg/api/server.go` | Main implementation |
| `pkg/api/handlers.go` | JSON metrics handler |
| `pkg/api/prometheus_test.go` | Unit tests |
| `tests/prometheus-integration-test.go` | Integration test |
| `docs/PROMETHEUS_INTEGRATION.md` | Full documentation |

## Dependencies

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)
```

Version: `v1.23.2` (already in go.mod)

## Quick Commands

```bash
# Run integration test
go run tests/prometheus-integration-test.go

# Run verification
bash scripts/verify-prometheus-implementation.sh

# Build and test
go build ./pkg/api && go test ./pkg/api

# Start server
go run cmd/ollamamax/main.go
```

## Best Practices

✅ **DO**:
- Use rate() for counters
- Use histogram_quantile() for percentiles
- Keep label cardinality low (< 10,000 series)
- Normalize dynamic path parameters
- Exclude metrics endpoints from tracking

❌ **DON'T**:
- Add user IDs or tokens as labels
- Track every unique URL
- Use gauges for counters
- Forget to register metrics
- Block in middleware

## Performance

- **Memory per metric**: ~100 bytes base + labels
- **CPU overhead**: < 1% with proper cardinality
- **Typical cardinality**: 500-2000 series
- **Max cardinality**: < 10,000 series recommended

## Support

- **Documentation**: `/docs/PROMETHEUS_INTEGRATION.md`
- **Integration Test**: `go run tests/prometheus-integration-test.go`
- **Prometheus Docs**: https://prometheus.io/docs/
- **Client Library**: https://github.com/prometheus/client_golang

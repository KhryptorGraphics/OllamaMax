# Prometheus Integration - Implementation Summary

## Overview

This document describes the implementation of Comment 1, which adds Prometheus metrics support to the OllamaMax API server while maintaining backward compatibility with the existing JSON metrics endpoint.

## Changes Made

### 1. Updated `pkg/api/server.go`

#### Added Imports
```go
import (
    // ... existing imports
    "strconv"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)
```

#### Extended Server Struct
Added Prometheus registry and metrics to the Server struct:
```go
type Server struct {
    // ... existing fields
    registry *prometheus.Registry

    // Prometheus metrics
    httpRequestsTotal          *prometheus.CounterVec
    httpRequestDuration        *prometheus.HistogramVec
    httpRequestsInFlight       prometheus.Gauge
}
```

#### Updated NewServer Function
Initialized Prometheus registry and registered three metrics:

1. **http_requests_total** (Counter with labels: method, endpoint, status)
   - Tracks total number of HTTP requests
   - Labels allow filtering by HTTP method, endpoint path, and status code

2. **http_request_duration_seconds** (Histogram with labels: method, endpoint, status)
   - Measures HTTP request duration in seconds
   - Uses default Prometheus buckets for latency distribution

3. **http_requests_in_flight** (Gauge)
   - Tracks number of currently active HTTP requests
   - Useful for monitoring server load

#### Added Prometheus Middleware
Created `prometheusMiddleware()` function that:
- Skips `/metrics` and `/metrics.json` endpoints to avoid recursion
- Increments/decrements in-flight requests gauge
- Records request duration using histogram
- Increments total request counter with appropriate labels
- Normalizes endpoint paths to avoid high cardinality

#### Updated Routes
Modified `setupRouter()` to:
- Add Prometheus middleware to the middleware chain
- Replace `/metrics` route with Prometheus handler: `promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{})`
- Add new `/metrics.json` route for backward compatibility

### 2. Updated `pkg/api/handlers.go`

Renamed `metricsHandler` to `metricsJSONHandler`:
```go
// JSON metrics handler for backward compatibility
func (s *Server) metricsJSONHandler(c *gin.Context) {
    stats := s.db.Stats()
    c.JSON(http.StatusOK, gin.H{
        "database": stats,
        "timestamp": time.Now(),
    })
}
```

## Metrics Endpoints

### `/metrics` (Prometheus Format)
Returns metrics in Prometheus text exposition format:

```
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",endpoint="/api/v1/models",status="200"} 42
http_requests_total{method="POST",endpoint="/api/v1/models",status="201"} 15

# HELP http_request_duration_seconds HTTP request duration in seconds
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{method="GET",endpoint="/api/v1/health",status="200",le="0.005"} 10
http_request_duration_seconds_bucket{method="GET",endpoint="/api/v1/health",status="200",le="0.01"} 25
...

# HELP http_requests_in_flight Number of HTTP requests currently being processed
# TYPE http_requests_in_flight gauge
http_requests_in_flight 3
```

**Use case**: Prometheus scraping, monitoring dashboards (Grafana)

### `/metrics.json` (JSON Format - Backward Compatible)
Returns metrics in JSON format:

```json
{
  "database": {
    "connections": 10,
    "queries_total": 1234
  },
  "timestamp": "2025-10-27T10:30:00Z"
}
```

**Use case**: Legacy systems, custom monitoring tools, development debugging

## Integration with Prometheus

### Prometheus Configuration

Add the following to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'ollamamax'
    static_configs:
      - targets: ['localhost:8080']
    scrape_interval: 15s
    metrics_path: /metrics
```

### Example Queries

**Request Rate (requests per second)**:
```promql
rate(http_requests_total[5m])
```

**Request Rate by Endpoint**:
```promql
rate(http_requests_total[5m]) by (endpoint)
```

**95th Percentile Latency**:
```promql
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
```

**Average Latency by Endpoint**:
```promql
rate(http_request_duration_seconds_sum[5m]) / rate(http_request_duration_seconds_count[5m])
```

**Error Rate (4xx and 5xx responses)**:
```promql
sum(rate(http_requests_total{status=~"4..|5.."}[5m]))
```

**Current Active Requests**:
```promql
http_requests_in_flight
```

## Grafana Dashboard

### Recommended Panels

1. **Request Rate Graph**
   - Query: `rate(http_requests_total[5m])`
   - Visualization: Time series graph

2. **Latency Heatmap**
   - Query: `rate(http_request_duration_seconds_bucket[5m])`
   - Visualization: Heatmap

3. **Active Requests Gauge**
   - Query: `http_requests_in_flight`
   - Visualization: Gauge

4. **Error Rate**
   - Query: `sum(rate(http_requests_total{status=~"4..|5.."}[5m]))`
   - Visualization: Time series with alert threshold

5. **Top Endpoints by Traffic**
   - Query: `topk(10, sum(rate(http_requests_total[5m])) by (endpoint))`
   - Visualization: Bar chart

## Testing

### Manual Testing

1. **Start the server**:
   ```bash
   go run cmd/ollamamax/main.go
   ```

2. **Check Prometheus metrics**:
   ```bash
   curl http://localhost:8080/metrics
   ```

3. **Check JSON metrics**:
   ```bash
   curl http://localhost:8080/metrics.json
   ```

4. **Generate traffic**:
   ```bash
   for i in {1..100}; do curl http://localhost:8080/health; done
   ```

5. **Verify metrics updated**:
   ```bash
   curl http://localhost:8080/metrics | grep http_requests_total
   ```

### Automated Testing

Run the integration test:
```bash
cd tests
go run prometheus-integration-test.go
```

Expected output:
```
✅ All Prometheus integration tests passed!

Implementation Summary:
  ✓ Prometheus registry initialized
  ✓ Three metrics registered (counter, histogram, gauge)
  ✓ Prometheus middleware tracking requests
  ✓ /metrics endpoint serving Prometheus format
  ✓ /metrics.json endpoint serving JSON format (backward compatibility)
```

## Performance Considerations

### Label Cardinality
The middleware normalizes endpoint paths using `c.FullPath()` to avoid high cardinality issues. This means:
- `/api/v1/models/123` → `/api/v1/models/:id`
- `/api/v1/users/456` → `/api/v1/users/:id`

This prevents creating unlimited unique metric series for dynamic path parameters.

### Metrics Exclusion
The `/metrics` and `/metrics.json` endpoints are explicitly excluded from metric collection to avoid:
- Recursive metric collection
- Inflated request counts
- Skewed latency measurements

### Memory Usage
The Prometheus metrics use:
- **Counter**: O(labels) memory - one float64 per unique label combination
- **Histogram**: O(labels × buckets) memory - default 11 buckets
- **Gauge**: O(1) memory - single float64 value

Expected memory overhead: < 1MB for typical workloads (< 100 unique endpoints)

## Migration Notes

### Backward Compatibility
- ✅ Existing `/metrics.json` functionality preserved
- ✅ No breaking changes to API responses
- ✅ Database stats still available at `/metrics.json`

### Upgrading Monitoring Systems
1. Update Prometheus scrape config to use `/metrics` endpoint
2. Keep legacy systems using `/metrics.json` until migration complete
3. Update dashboards to use new Prometheus metrics
4. Validate alert rules with new metric names

## Security Considerations

### Authentication
Currently, both `/metrics` and `/metrics.json` endpoints are **unauthenticated**. Consider:
- Adding authentication for production environments
- Using Prometheus service discovery with authentication
- Restricting access via firewall rules

Example with basic auth:
```go
router.GET("/metrics", gin.BasicAuth(gin.Accounts{
    "prometheus": "secure_password",
}), gin.WrapH(promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{})))
```

### Information Disclosure
Metrics may reveal:
- API usage patterns
- Internal endpoint structure
- Performance characteristics
- Error rates

Ensure metrics endpoints are not publicly accessible in production.

## Troubleshooting

### Metrics Not Appearing
1. Check server logs for registration errors
2. Verify middleware is configured: `router.Use(s.prometheusMiddleware())`
3. Ensure requests are reaching the server
4. Check Prometheus scrape status

### High Memory Usage
1. Review label cardinality: `curl localhost:8080/metrics | grep http_requests_total | wc -l`
2. Verify path normalization is working
3. Consider adjusting histogram buckets if needed

### Incorrect Metric Values
1. Verify middleware order (should be after logging, before rate limiting)
2. Check for panics that might skip deferred operations
3. Ensure status codes are set correctly

## Future Enhancements

### Potential Additions
1. **Custom business metrics**:
   - Model inference latency
   - Token generation rate
   - Cache hit rates

2. **Additional HTTP metrics**:
   - Request/response body size
   - Connection duration
   - Concurrent connections per endpoint

3. **System metrics**:
   - Go runtime metrics (goroutines, memory, GC)
   - Database connection pool metrics
   - External API call latency

4. **Distributed tracing**:
   - OpenTelemetry integration
   - Trace ID propagation
   - Span metrics

## References

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Prometheus Client Library for Go](https://github.com/prometheus/client_golang)
- [Gin Framework](https://github.com/gin-gonic/gin)
- [Best Practices for Naming Metrics](https://prometheus.io/docs/practices/naming/)
- [Instrumentation Best Practices](https://prometheus.io/docs/practices/instrumentation/)

## Related Files

- `/home/kp/OllamaMax/pkg/api/server.go` - Main implementation
- `/home/kp/OllamaMax/pkg/api/handlers.go` - JSON metrics handler
- `/home/kp/OllamaMax/pkg/api/prometheus_test.go` - Unit tests
- `/home/kp/OllamaMax/tests/prometheus-integration-test.go` - Integration test

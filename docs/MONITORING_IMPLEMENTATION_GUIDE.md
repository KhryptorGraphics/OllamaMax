# OllamaMax Monitoring Implementation Guide

## Overview

This document describes the unified monitoring architecture for OllamaMax, ensuring all application metrics are exposed via a single `/metrics` endpoint for Prometheus scraping.

## Architecture

### Unified Metrics Registry

All component metrics (API, Database, P2P, Load Balancer) are exposed through a single HTTP endpoint using a shared Prometheus registry:

```go
// In pkg/api/server.go - NewServer()
registry := prometheus.NewRegistry()

// Register API metrics
registry.Register(httpRequestsTotal)
registry.Register(httpRequestDuration)
registry.Register(httpRequestsInFlight)

// Register database metrics
if db != nil {
    db.RegisterTo(registry)
}

// Register P2P metrics (when available)
// Call server.RegisterP2PMetrics(p2pNode) after server creation

// Expose unified endpoint
router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(
    s.registry,
    promhttp.HandlerOpts{},
)))
```

### Component Metrics

#### 1. API Server Metrics
- **Namespace**: Default (no prefix)
- **Metrics**: `http_requests_total`, `http_request_duration_seconds`, `http_requests_in_flight`

#### 2. Database Metrics
- **Namespace**: `ollamamax_database_`
- **Pattern**: Repositories call `manager.RecordQuery()` instead of maintaining separate metrics
- **Key Metrics**: `db_connections_*`, `db_queries_total`, `db_query_duration_seconds`, `cache_*`

#### 3. P2P Network Metrics
- **Namespace**: `ollamamax_p2p_`
- **Registration Pattern**: Components expose `RegisterTo(prometheus.Registerer) error` method
- **Key Metrics**:
  - `ollamamax_p2p_connected_peers` - Current peer count
  - `ollamamax_p2p_messages_sent_total` - Messages sent by topic
  - `ollamamax_p2p_messages_received_total` - Messages received by topic
  - `ollamamax_p2p_bytes_sent_total` - Bytes sent by topic
  - `ollamamax_p2p_bytes_received_total` - Bytes received by topic
  - `ollamamax_p2p_message_latency_seconds` - Message processing latency histogram
  - `ollamamax_p2p_connection_errors_total` - Connection error count
- **Labels**: Keep low-cardinality (topic, direction)

#### 4. Load Balancer Metrics
- **Namespace**: `lb_`
- **Metrics**: `lb_requests_total`, `lb_node_utilization`

## Alert Rules - Correct PromQL Syntax

### Histogram Quantiles
```promql
# ✅ CORRECT
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1.0

# ❌ WRONG
http_request_duration_seconds{quantile="0.95"} > 1.0
```

### Ratios
```promql
# ✅ CORRECT
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) > 0.05
```

## Grafana Configuration

- **Datasource UIDs**: Stable UIDs added (`prometheus`, `jaeger`, `elasticsearch`)
- **Dashboard Provisioning**: Mounted `/var/lib/grafana/dashboards` in docker-compose

## Implementation Checklist

### ✅ Completed
1. Unified Prometheus registry (single shared registry)
2. Refactored database repositories to use manager metrics
3. Added P2P bytes counters with `ollamamax_p2p_` prefix
4. Fixed alert rules syntax
5. Added datasource UIDs
6. Mounted dashboard directory
7. Implemented `RegisterTo` pattern for P2P metrics
8. Added `Server.RegisterP2PMetrics()` method for optional P2P registration
9. Updated Grafana P2P dashboard to use `ollamamax_p2p_*` metric names
10. Documented metrics registration pattern and naming convention

### ⏳ Pending
11. Validate `/metrics` endpoint with P2P node running
12. Test Grafana P2P dashboard panels render data
13. Verify P2P alerts in `monitoring/alerts.yml` evaluate correctly

## Metrics Registration Pattern

### Standard Pattern for Subsystems

All subsystems (Database, P2P, Load Balancer, etc.) should follow this pattern:

1. **Define metrics with namespace prefix** in the component's constructor:
   ```go
   connectedPeers := prometheus.NewGauge(prometheus.GaugeOpts{
       Name: "ollamamax_p2p_connected_peers",
       Help: "Number of currently connected peers",
   })
   ```

2. **Register metrics to a local registry** for isolation:
   ```go
   registry := prometheus.NewRegistry()
   registry.MustRegister(connectedPeers)
   ```

3. **Expose `RegisterTo` method** for wiring to main registry:
   ```go
   func (n *BasicNode) RegisterTo(registerer prometheus.Registerer) error {
       collectors := []prometheus.Collector{
           n.connectedPeers,
           n.messagesSent,
           // ... other metrics
       }
       for _, collector := range collectors {
           if err := registerer.Register(collector); err != nil {
               return err
           }
       }
       return nil
   }
   ```

4. **Wire to main registry** at composition point:
   ```go
   // In main.go or server initialization
   apiServer, err := server.NewServer(cfg, db, logger)

   // If P2P node is available
   if p2pNode != nil {
       if err := apiServer.RegisterP2PMetrics(p2pNode); err != nil {
           logger.Warn("Failed to register P2P metrics", "error", err)
       }
   }
   ```

### Naming Convention

- **Prefix**: `ollamamax_<subsystem>_`
- **Examples**:
  - API: `ollamamax_api_http_requests_total`
  - Database: `ollamamax_database_connections_active`
  - P2P: `ollamamax_p2p_connected_peers`
  - Load Balancer: `ollamamax_lb_requests_total`

### Dashboard Alignment

When creating Grafana dashboards, always use the full metric name with prefix:

```json
{
  "expr": "ollamamax_p2p_connected_peers",
  "legendFormat": "Connected Peers"
}
```

**Never** use unprefixed names like `p2p_connected_peers` - they won't match exported metrics.

## Best Practices

- Use consistent namespaces per component (`ollamamax_<subsystem>_`)
- Keep label cardinality low (avoid high-cardinality labels like IDs)
- Inject dependencies instead of globals
- Use `histogram_quantile` for latency percentiles
- Register metrics exactly once to avoid duplicate collector errors
- Document all metrics in this guide with their purpose and labels

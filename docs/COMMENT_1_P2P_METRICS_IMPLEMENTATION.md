# Comment 1: P2P Metrics Visibility Implementation

## Summary

Implemented P2P metrics registration and Grafana dashboard alignment to make P2P metrics visible at the main `/metrics` endpoint and ensure dashboard queries match exported metric names.

## Problem Statement

- P2P metrics were defined with `ollamamax_p2p_*` prefix in `pkg/p2p/node.go`
- P2P collectors were not registered on the main Prometheus registry in `pkg/api/server.go`
- Grafana dashboard `monitoring/grafana/dashboards/p2p-detailed.json` queried unprefixed `p2p_*` series
- Result: P2P metrics were not scrapeable at `/metrics` and dashboard panels showed no data

## Implementation

### 1. API Server Changes (`pkg/api/server.go`)

#### Added P2P Metrics Registration Method

```go
// RegisterP2PMetrics registers P2P node metrics to the server's Prometheus registry
// This should be called after server creation if a P2P node is available
func (s *Server) RegisterP2PMetrics(p2pNode interface {
	RegisterTo(prometheus.Registerer) error
}) error {
	if err := p2pNode.RegisterTo(s.registry); err != nil {
		s.logger.Warn("Failed to register P2P metrics", "error", err)
		return fmt.Errorf("failed to register P2P metrics: %w", err)
	}
	s.logger.Info("P2P metrics registered successfully")
	return nil
}
```

#### Fixed Database Registration

Changed database metrics registration to check for nil:

```go
// Register database metrics to the main registry
if db != nil {
    if err := db.RegisterTo(registry); err != nil {
        logger.Warn("Failed to register database metrics", "error", err)
    }
}
```

### 2. Grafana Dashboard Updates (`monitoring/grafana/dashboards/p2p-detailed.json`)

Updated all metric queries to use the correct `ollamamax_p2p_` prefix:

| Panel | Old Query | New Query |
|-------|-----------|-----------|
| Connected Peers | `p2p_connected_peers` | `ollamamax_p2p_connected_peers` |
| Message Rate | `rate(p2p_messages_sent_total[5m])` | `rate(ollamamax_p2p_messages_sent_total[5m])` |
| Message Rate | `rate(p2p_messages_received_total[5m])` | `rate(ollamamax_p2p_messages_received_total[5m])` |
| Message Latency P95 | `histogram_quantile(0.95, rate(p2p_message_latency_seconds_bucket[5m]))` | `histogram_quantile(0.95, rate(ollamamax_p2p_message_latency_seconds_bucket[5m]))` |
| Messages by Topic | `sum by (topic) (p2p_messages_sent_total)` | `sum by (topic) (ollamamax_p2p_messages_sent_total)` |
| Connection Errors | `rate(p2p_connection_errors_total[5m])` | `rate(ollamamax_p2p_connection_errors_total[5m])` |
| Network Throughput | `rate(p2p_bytes_sent_total[5m])` | `rate(ollamamax_p2p_bytes_sent_total[5m])` |
| Network Throughput | `rate(p2p_bytes_received_total[5m])` | `rate(ollamamax_p2p_bytes_received_total[5m])` |

### 3. Documentation Updates (`docs/MONITORING_IMPLEMENTATION_GUIDE.md`)

#### Updated Architecture Section

Documented the unified registry pattern:

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

#### Documented P2P Metrics

Added comprehensive P2P metrics documentation:

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

#### Added Metrics Registration Pattern Section

Documented the standard pattern for all subsystems:

1. Define metrics with namespace prefix
2. Register metrics to a local registry for isolation
3. Expose `RegisterTo` method for wiring to main registry
4. Wire to main registry at composition point

#### Added Naming Convention

- **Prefix**: `ollamamax_<subsystem>_`
- **Examples**:
  - API: `ollamamax_api_http_requests_total`
  - Database: `ollamamax_database_connections_active`
  - P2P: `ollamamax_p2p_connected_peers`
  - Load Balancer: `ollamamax_lb_requests_total`

#### Updated Best Practices

- Use consistent namespaces per component (`ollamamax_<subsystem>_`)
- Keep label cardinality low (avoid high-cardinality labels like IDs)
- Register metrics exactly once to avoid duplicate collector errors
- Document all metrics with their purpose and labels

## Usage Example

For applications that create a P2P node (like `ollama-distributed/cmd/node/main.go`):

```go
// Initialize P2P node
p2pNode, err := p2p.NewNode(ctx, &cfg.P2P)
if err != nil {
    return fmt.Errorf("failed to create P2P node: %w", err)
}

// Initialize API server
apiServer, err := server.NewServer(cfg, db, logger)
if err != nil {
    return fmt.Errorf("failed to create API server: %w", err)
}

// Register P2P metrics to the API server's registry
if err := apiServer.RegisterP2PMetrics(p2pNode); err != nil {
    logger.Warn("Failed to register P2P metrics", "error", err)
}

// Start services...
```

## Verification Steps

1. **Start the stack** with P2P node enabled
2. **Hit `/metrics`** endpoint and confirm `ollamamax_p2p_*` series appear:
   ```bash
   curl http://localhost:8080/metrics | grep ollamamax_p2p_
   ```
3. **Check Grafana P2P dashboard** panels render data
4. **Verify alerts** in `monitoring/alerts.yml` evaluate (already using correct names)

## Files Modified

1. `pkg/api/server.go` - Added `RegisterP2PMetrics()` method, fixed database registration
2. `monitoring/grafana/dashboards/p2p-detailed.json` - Updated all 8 metric queries
3. `docs/MONITORING_IMPLEMENTATION_GUIDE.md` - Comprehensive documentation updates

## Alert Configuration

The alerts in `monitoring/alerts.yml` already use the correct metric names:

```yaml
- alert: P2PPeerCountLow
  expr: ollamamax_p2p_connected_peers < 3
  for: 5m
  labels:
    severity: warning

- alert: P2PHighLatency
  expr: histogram_quantile(0.95, rate(ollamamax_p2p_message_latency_seconds_bucket[5m])) > 0.5
  for: 5m
  labels:
    severity: warning
```

## Low-Cardinality Labels

P2P metrics use only low-cardinality labels to avoid metric explosion:

- `topic` - Message topic (limited set of known topics)
- `direction` - "sent" or "received" (2 values)

**Avoided**: peer IDs, timestamps, message IDs, or other high-cardinality values.

## Next Steps

1. Test with actual P2P node running
2. Validate dashboard panels populate with real data
3. Verify alert rules trigger correctly
4. Consider adding similar registration for load balancer metrics when available

## Notes

- The main `pkg/api/server.go` currently doesn't create a P2P node (that's done in `ollama-distributed/cmd/node/main.go`)
- The `RegisterP2PMetrics()` method allows optional P2P registration when a node is available
- The pattern is extensible to other subsystems (load balancer, consensus, etc.)
- All metrics follow the `ollamamax_<subsystem>_` naming convention for consistency


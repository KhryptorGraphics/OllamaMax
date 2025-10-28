# Monitoring Metrics Registry Pattern

## Overview

OllamaMax uses a unified Prometheus metrics registry pattern to ensure all application metrics are exposed at a single `/metrics` endpoint. This document explains the architecture and implementation.

## Architecture

### Single Registry Pattern

All metrics from different subsystems (API, Database, P2P, Load Balancer) are registered to a single Prometheus registry owned by the API server. This ensures:

1. **Single Scrape Target**: Prometheus only needs to scrape one endpoint
2. **Consistent Naming**: All metrics follow the `ollamamax_*` namespace convention
3. **No Hidden Metrics**: No metrics are trapped in private registries
4. **Simplified Configuration**: Grafana dashboards and alert rules reference one datasource

### Component Integration

#### Database Manager (`pkg/database/manager.go`)

**Metric Names**: `ollamamax_database_*`

```go
// Register database metrics to the main app registry
func (dm *DatabaseManager) RegisterTo(registerer prometheus.Registerer) error {
    collectors := []prometheus.Collector{
        dm.dbConnectionsOpen,
        dm.dbConnectionsInUse,
        dm.dbConnectionsIdle,
        dm.dbQueriesTotal,
        dm.dbQueryDuration,
        dm.redisCommandsTotal,
        dm.redisCommandDuration,
        dm.cacheHitsTotal,
        dm.cacheMissesTotal,
        // ...
    }

    for _, collector := range collectors {
        registerer.Register(collector)
    }
    return nil
}
```

**Repository Integration**:
- All repositories (ModelRepository, UserRepository, NodeRepository, etc.) call `manager.RecordQuery(operation, table, duration)`
- Cache operations call `manager.RecordCacheHit()` and `manager.RecordCacheMiss()`
- Redis commands call `manager.RecordRedisCommand(command, duration)`
- No `promauto` metrics in repositories

#### P2P Node (`pkg/p2p/node.go`)

**Metric Names**: `p2p_*`

```go
// Register P2P metrics to the main app registry
func (n *BasicNode) RegisterTo(registerer prometheus.Registerer) error {
    collectors := []prometheus.Collector{
        n.connectedPeers,
        n.messagesSent,
        n.messagesReceived,
        n.bytesSent,        // Added for network throughput
        n.bytesReceived,    // Added for network throughput
        n.messageLatency,
        n.connectionErrors,
    }

    for _, collector := range collectors {
        registerer.Register(collector)
    }
    return nil
}
```

**Metrics**:
- `p2p_connected_peers`: Current number of connected peers
- `p2p_messages_sent_total{topic}`: Messages sent by topic
- `p2p_messages_received_total{topic}`: Messages received by topic
- `p2p_bytes_sent_total{topic}`: Bytes sent by topic
- `p2p_bytes_received_total{topic}`: Bytes received by topic
- `p2p_message_latency_seconds`: Message processing latency histogram
- `p2p_connection_errors_total`: Connection error counter

#### Load Balancer (`pkg/distributed/load_balancer.go`)

**Metric Names**: `lb_*`

```go
// Create balancer with shared registry
func NewRoundRobinBalancerWithRegistry(registerer prometheus.Registerer) *RoundRobinBalancer {
    requestsTotal := prometheus.NewCounterVec(...)
    selectionDuration := prometheus.NewHistogram(...)
    nodeUtilization := prometheus.NewGaugeVec(...)

    if registerer != nil {
        registerer.MustRegister(requestsTotal, selectionDuration, nodeUtilization)
    }

    return &RoundRobinBalancer{...}
}
```

**Metrics**:
- `lb_requests_total{strategy,node_id}`: Load balancer requests by strategy and node
- `lb_node_selection_duration_seconds`: Node selection latency histogram
- `lb_node_utilization{node_id}`: Current node utilization gauge
- `lb_strategy_switches_total`: Strategy switch counter (SmartLoadBalancer)

#### API Server (`pkg/api/server.go`)

**Metric Names**: `http_*`

```go
func NewServer(cfg *config.Config, db *database.DatabaseManager, logger *slog.Logger) (*Server, error) {
    // Create main registry
    registry := prometheus.NewRegistry()

    // Register API metrics
    httpRequestsTotal := prometheus.NewCounterVec(...)
    httpRequestDuration := prometheus.NewHistogramVec(...)
    httpRequestsInFlight := prometheus.NewGauge(...)
    registry.MustRegister(httpRequestsTotal, httpRequestDuration, httpRequestsInFlight)

    // Register database metrics
    db.RegisterTo(registry)

    // Register P2P metrics (if P2P node is available)
    // p2pNode.RegisterTo(registry)

    // Register load balancer metrics (if load balancer is available)
    // loadBalancer.RegisterTo(registry)

    return &Server{registry: registry, ...}
}
```

## Metric Naming Conventions

### Database Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `ollamamax_database_db_connections_open` | Gauge | - | Open database connections |
| `ollamamax_database_db_connections_active` | Gauge | - | Active database connections |
| `ollamamax_database_db_connections_idle` | Gauge | - | Idle database connections |
| `ollamamax_database_db_connections_max` | Gauge | - | Maximum allowed connections |
| `ollamamax_database_db_queries_total` | Counter | operation, table | Total database queries |
| `ollamamax_database_db_query_duration_seconds` | Histogram | operation, table | Query execution time |
| `ollamamax_database_redis_commands_total` | Counter | command | Total Redis commands |
| `ollamamax_database_redis_command_duration_seconds` | Histogram | command | Redis command duration |
| `ollamamax_database_cache_hits_total` | Counter | - | Cache hits |
| `ollamamax_database_cache_misses_total` | Counter | - | Cache misses |

**Label Cardinality**:
- `operation`: select, insert, update, delete, list, authenticate, create
- `table`: models, users, nodes, sessions, audit_log_entries, model_replicas
- `command`: get, set, del, incr, expire

### API Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `http_requests_total` | Counter | method, endpoint, status | Total HTTP requests |
| `http_request_duration_seconds` | Histogram | method, endpoint, status | Request duration |
| `http_requests_in_flight` | Gauge | - | Current in-flight requests |

### P2P Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `p2p_connected_peers` | Gauge | - | Connected peer count |
| `p2p_messages_sent_total` | Counter | topic | Messages sent |
| `p2p_messages_received_total` | Counter | topic | Messages received |
| `p2p_bytes_sent_total` | Counter | topic | Bytes sent |
| `p2p_bytes_received_total` | Counter | topic | Bytes received |
| `p2p_message_latency_seconds` | Histogram | - | Message latency |
| `p2p_connection_errors_total` | Counter | - | Connection errors |

### Load Balancer Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `lb_requests_total` | Counter | strategy, node_id | Balancer requests |
| `lb_node_selection_duration_seconds` | Histogram | - | Selection time |
| `lb_node_utilization` | Gauge | node_id | Node utilization |
| `lb_strategy_switches_total` | Counter | - | Strategy switches |

## Grafana Configuration

### Datasource (`monitoring/grafana/provisioning/datasources/prometheus.yml`)

```yaml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    uid: prometheus  # Stable UID for dashboard references
    isDefault: true
    editable: true
```

### Dashboard Provisioning

Dashboards reference the stable `uid: prometheus` datasource:

```json
{
  "datasource": {
    "type": "prometheus",
    "uid": "prometheus"
  },
  "targets": [
    {
      "expr": "rate(ollamamax_database_db_queries_total[5m])",
      "legendFormat": "{{operation}} - {{table}}"
    }
  ]
}
```

## Alert Rules

### PromQL Examples

```yaml
# Database pool exhaustion
expr: ollamamax_database_db_connections_active / ollamamax_database_db_connections_max > 0.9

# Database query latency
expr: histogram_quantile(0.95, rate(ollamamax_database_db_query_duration_seconds_bucket[5m])) > 0.5

# API error rate
expr: (sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))) > 0.05

# P2P message latency
expr: histogram_quantile(0.95, rate(p2p_message_latency_seconds_bucket[5m])) > 0.5

# Load balancer imbalance
expr: stddev(lb_node_utilization) > 0.3
```

## Implementation Checklist

- [x] Database: Add `RegisterTo(registerer)` method
- [x] Database: Remove `promauto` from repositories
- [x] Database: Add `RecordRedisCommand()` helper
- [x] P2P: Add `RegisterTo(registerer)` method
- [x] P2P: Add bytes sent/received counters
- [x] Load Balancer: Accept `registerer` in constructors
- [x] API Server: Call `db.RegisterTo(registry)`
- [x] API Server: Expose single `/metrics` endpoint
- [x] Grafana: Set stable datasource UID
- [x] Grafana: Mount dashboard directory
- [x] Alerts: Update PromQL to use correct metric names
- [x] Alerts: Use histogram quantiles instead of summary quantiles

## Validation

### Check Metrics Exposure

```bash
# Verify all metrics appear at /metrics
curl http://localhost:13100/metrics | grep -E "^(ollamamax_database|p2p|lb|http)_"

# Check specific database metrics
curl http://localhost:13100/metrics | grep ollamamax_database_db_connections_active

# Check P2P metrics
curl http://localhost:13100/metrics | grep p2p_bytes_sent_total

# Check load balancer metrics
curl http://localhost:13100/metrics | grep lb_requests_total
```

### Verify in Prometheus

```promql
# Query Prometheus directly
up{job="ollamamax-api"}
rate(ollamamax_database_db_queries_total[5m])
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
p2p_connected_peers
lb_node_utilization
```

### Test Grafana Dashboards

1. Open Grafana at `http://localhost:3001`
2. Navigate to Dashboards
3. Check that panels populate with data
4. Verify datasource UID matches in all panels

### Test Alert Rules

```bash
# Reload Prometheus configuration
curl -X POST http://localhost:9090/-/reload

# Check alert rules
curl http://localhost:9090/api/v1/rules | jq '.data.groups[].rules[] | select(.type=="alerting")'

# Verify alert evaluation
curl http://localhost:9090/api/v1/alerts
```

## Migration Notes

### From Multiple Registries

**Before**:
```go
// Each component had its own registry
dbRegistry := db.GetPrometheusRegistry()
p2pRegistry := p2pNode.GetPrometheusRegistry()
lbRegistry := loadBalancer.GetPrometheusRegistry()

// Metrics exposed via custom gatherer
router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(
    prometheus.Gatherers{apiRegistry, dbRegistry, p2pRegistry, lbRegistry},
    promhttp.HandlerOpts{},
)))
```

**After**:
```go
// Single shared registry
registry := prometheus.NewRegistry()
db.RegisterTo(registry)
p2pNode.RegisterTo(registry)
loadBalancer.RegisterTo(registry)

// Metrics exposed from single registry
router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))
```

### Repository Metrics Migration

**Before** (promauto globals):
```go
var dbQueriesTotal = promauto.NewCounterVec(...)

func (r *ModelRepository) GetByID(ctx context.Context, id uuid.UUID) (*Model, error) {
    dbQueriesTotal.WithLabelValues("get", "models").Inc()
    // ...
}
```

**After** (manager methods):
```go
func (r *ModelRepository) GetByID(ctx context.Context, id uuid.UUID) (*Model, error) {
    start := time.Now()
    err := r.db.GetContext(ctx, &model, query, id)
    if r.manager != nil {
        r.manager.RecordQuery("get", "models", time.Since(start))
    }
    // ...
}
```

## Best Practices

1. **Always use RegisterTo()**: Never create private registries in components
2. **Inject registerer**: Pass `prometheus.Registerer` to constructors
3. **Low cardinality labels**: Keep label values bounded (operation, table, command)
4. **Consistent naming**: Use `subsystem_component_metric` pattern
5. **Document metrics**: Add help text describing each metric
6. **Test metrics**: Verify metrics appear at `/metrics` endpoint
7. **Monitor cardinality**: Track unique label combinations to avoid explosion

## Troubleshooting

### Metrics Not Appearing

1. Check if `RegisterTo()` was called during initialization
2. Verify no "already registered" errors in logs
3. Ensure metric names don't conflict across components
4. Check Prometheus scrape target is healthy

### Grafana Panels Empty

1. Verify datasource UID matches dashboard configuration
2. Check PromQL queries use correct metric names
3. Ensure time range includes recent data
4. Test queries directly in Prometheus UI

### Alert Rules Not Firing

1. Validate PromQL syntax in Prometheus UI
2. Check metric names match actual exposed metrics
3. Verify thresholds are appropriate for your environment
4. Ensure alertmanager is configured correctly

## References

- [Prometheus Go Client Documentation](https://prometheus.io/docs/guides/go-application/)
- [Prometheus Best Practices - Metric and Label Naming](https://prometheus.io/docs/practices/naming/)
- [Grafana Provisioning Documentation](https://grafana.com/docs/grafana/latest/administration/provisioning/)
- [Alerting Rules Documentation](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/)

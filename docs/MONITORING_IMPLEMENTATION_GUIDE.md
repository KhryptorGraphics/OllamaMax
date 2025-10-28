# OllamaMax Monitoring Implementation Guide

## Overview

This document describes the unified monitoring architecture for OllamaMax, ensuring all application metrics are exposed via a single `/metrics` endpoint for Prometheus scraping.

## Architecture

### Unified Metrics Registry

All component metrics (API, Database, P2P, Load Balancer) are exposed through a single HTTP endpoint using Prometheus `Gatherers`:

```go
router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(
    prometheus.Gatherers{
        s.registry,              // API metrics
        s.db.GetPrometheusRegistry(),  // Database metrics
    },
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
- **Namespace**: `p2p_`
- **New**: `p2p_bytes_sent_total`, `p2p_bytes_received_total`

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
1. Unified Prometheus registries via `Gatherers`
2. Refactored database repositories to use manager metrics
3. Added P2P bytes counters
4. Fixed alert rules syntax
5. Added datasource UIDs
6. Mounted dashboard directory

### ⏳ Pending
7. Update dashboard JSON metric names to `ollamamax_database_*`
8. Validate `/metrics` endpoint

## Best Practices

- Use consistent namespaces per component
- Keep label cardinality low
- Inject dependencies instead of globals
- Use `histogram_quantile` for latency percentiles

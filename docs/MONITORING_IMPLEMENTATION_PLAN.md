# Comprehensive Monitoring & Observability Implementation Plan

## Executive Summary

This document provides a detailed implementation plan for adding comprehensive monitoring and observability to OllamaMax using Prometheus metrics, Jaeger distributed tracing, and ELK stack for log aggregation.

## Current State

### ✅ Existing Infrastructure
- **Prometheus & Grafana**: Deployed in docker-compose.yml and k8s/monitoring-stack.yaml
- **Observability Code**: ollama-distributed/pkg/observability/ contains:
  - PrometheusExporter with Counter, Gauge, Histogram support
  - OpenTelemetryAdapter with Jaeger integration
  - ComponentTracer and DistributedTracingSystem
  - Structured logging with TraceID/SpanID
- **Basic Dashboards**: ollamamax-overview.json, p2p-network-monitoring.json, deployment-dashboard.json
- **Alert Rules**: monitoring/alerts.yml with basic performance and system alerts

### ⚠️ Gaps Identified
1. No Prometheus instrumentation in pkg/api and internal/server packages
2. No Jaeger deployment in infrastructure
3. No ELK stack (Elasticsearch, Logstash, Kibana) deployment
4. Missing specialized dashboards for API, Database, and detailed P2P metrics
5. Alert notification channels not configured
6. No log shipping to centralized logging
7. Metrics endpoints not exposed in API servers

## Implementation Phases

### Phase 1: Prometheus Metrics Instrumentation (Priority: HIGH)

#### Files to Modify:

**1. pkg/api/server.go**
```go
// Add imports
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/observability"
)

// Add to Server struct
type Server struct {
    ...
    promExporter *observability.PrometheusExporter
}

// In NewServer
promExporter, err := observability.NewPrometheusExporter("ollamamax", "api")
if err != nil {
    return nil, fmt.Errorf("failed to create Prometheus exporter: %w", err)
}

// Register metrics
promExporter.RegisterCounter("http_requests_total", "Total HTTP requests", "method", "endpoint", "status_code")
promExporter.RegisterHistogram("http_request_duration_seconds", "HTTP request duration", "method", "endpoint")
promExporter.RegisterGauge("http_requests_in_flight", "Current in-flight HTTP requests")

// Add /metrics endpoint in setupRouter
router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(promExporter.Registry(), promhttp.HandlerOpts{})))

// Update loggingMiddleware
func (s *Server) loggingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()

        // Increment in-flight requests
        s.promExporter.IncrementGauge("http_requests_in_flight", 1)
        defer s.promExporter.IncrementGauge("http_requests_in_flight", -1)

        c.Next()

        duration := time.Since(start)

        // Record metrics
        s.promExporter.IncrementCounter("http_requests_total", 1,
            c.Request.Method, c.Request.URL.Path, strconv.Itoa(c.Writer.Status()))
        s.promExporter.ObserveHistogram("http_request_duration_seconds", duration.Seconds(),
            c.Request.Method, c.Request.URL.Path)
    }
}
```

**2. internal/server/server.go**
```go
// Replace ServerMetrics struct with PrometheusExporter
type Server struct {
    ...
    promExporter *observability.PrometheusExporter
}

// Update NewServer
promExporter, err := observability.NewPrometheusExporter("ollamamax", "internal")
promExporter.RegisterCounter("requests_total", "Total requests", "method", "status")
promExporter.RegisterHistogram("request_duration_seconds", "Request duration")
promExporter.RegisterGauge("active_connections", "Active connections")

// Add /metrics endpoint
s.router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(s.promExporter.Registry(), promhttp.HandlerOpts{})))
```

**3. pkg/database/manager.go**
```go
type DatabaseManager struct {
    ...
    promExporter *observability.PrometheusExporter
}

// In NewDatabaseManager
promExporter, err := observability.NewPrometheusExporter("ollamamax", "database")
promExporter.RegisterGauge("db_connections_active", "Active PostgreSQL connections")
promExporter.RegisterGauge("db_connections_idle", "Idle PostgreSQL connections")
promExporter.RegisterCounter("db_connections_wait_count", "Connection wait count")
promExporter.RegisterHistogram("db_query_duration_seconds", "Query execution time", "operation", "table")
promExporter.RegisterCounter("db_queries_total", "Total queries executed", "operation", "table")
promExporter.RegisterCounter("redis_commands_total", "Total Redis commands", "command")
promExporter.RegisterHistogram("redis_command_duration_seconds", "Redis command duration", "command")
promExporter.RegisterCounter("cache_hits_total", "Cache hit count")
promExporter.RegisterCounter("cache_misses_total", "Cache miss count")

// Add goroutine to periodically export Stats()
go func() {
    ticker := time.NewTicker(15 * time.Second)
    defer ticker.Stop()
    for range ticker.C {
        stats := dm.DB.Stats()
        dm.promExporter.SetGauge("db_connections_active", float64(stats.InUse))
        dm.promExporter.SetGauge("db_connections_idle", float64(stats.Idle))
    }
}()
```

**4. pkg/p2p/node.go**
```go
type BasicNode struct {
    ...
    promExporter *observability.PrometheusExporter
}

// In NewBasicNode
promExporter, err := observability.NewPrometheusExporter("ollamamax", "p2p")
promExporter.RegisterGauge("p2p_connected_peers", "Number of connected peers")
promExporter.RegisterCounter("p2p_messages_sent_total", "Total messages sent", "topic", "message_type")
promExporter.RegisterCounter("p2p_messages_received_total", "Total messages received", "topic", "message_type")
promExporter.RegisterHistogram("p2p_message_latency_seconds", "Message round-trip latency", "topic")
promExporter.RegisterCounter("p2p_connection_errors_total", "Connection error count")
promExporter.RegisterCounter("p2p_bandwidth_bytes", "Bytes sent/received", "direction")

// Update Connect
if err := n.connect(ctx, peerAddr); err != nil {
    n.promExporter.IncrementCounter("p2p_connection_errors_total", 1)
    return err
}
n.promExporter.IncrementGauge("p2p_connected_peers", 1)

// Update Disconnect
n.promExporter.IncrementGauge("p2p_connected_peers", -1)

// Update Broadcast
n.promExporter.IncrementCounter("p2p_messages_sent_total", 1, topic, messageType)
n.promExporter.IncrementCounter("p2p_bandwidth_bytes", float64(len(data)), "sent")
```

**5. pkg/distributed/load_balancer.go**
```go
// Add PrometheusExporter to each balancer type
type RoundRobinBalancer struct {
    ...
    promExporter *observability.PrometheusExporter
}

// In constructors
promExporter, err := observability.NewPrometheusExporter("ollamamax", "loadbalancer")
promExporter.RegisterCounter("lb_requests_total", "Total requests", "strategy", "node_id")
promExporter.RegisterHistogram("lb_node_selection_duration_seconds", "Time to select node", "strategy")
promExporter.RegisterGauge("lb_node_utilization", "Current node utilization", "node_id")
promExporter.RegisterCounter("lb_strategy_switches_total", "Strategy switch count", "from_strategy", "to_strategy")

// In SelectNode methods
start := time.Now()
node, err := selectLogic()
duration := time.Since(start)
promExporter.ObserveHistogram("lb_node_selection_duration_seconds", duration.Seconds(), strategyName)
promExporter.IncrementCounter("lb_requests_total", 1, strategyName, node.ID)
```

### Phase 2: Jaeger Distributed Tracing (Priority: HIGH)

#### Files to Modify:

**1. pkg/api/server.go - Add Tracing**
```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

type Server struct {
    ...
    otelAdapter *observability.OpenTelemetryAdapter
    tracer      trace.Tracer
}

// In NewServer
otelAdapter, err := observability.NewOpenTelemetryAdapter(
    cfg.Observability.JaegerEndpoint,
    "ollamamax-api",
    cfg.Observability.SamplingRate,
)
s.tracer = otel.Tracer("ollamamax-api")

// Add tracing middleware in setupRouter
func (s *Server) tracingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, span := s.tracer.Start(c.Request.Context(),
            fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path))
        defer span.End()

        span.SetAttributes(
            attribute.String("http.method", c.Request.Method),
            attribute.String("http.url", c.Request.URL.Path),
            attribute.String("http.client_ip", c.ClientIP()),
        )

        c.Request = c.Request.WithContext(ctx)
        c.Next()

        span.SetAttributes(attribute.Int("http.status_code", c.Writer.Status()))
        if c.Writer.Status() >= 400 {
            span.RecordError(fmt.Errorf("HTTP %d", c.Writer.Status()))
        }
    }
}

// Update setupRouter
router.Use(s.tracingMiddleware())
```

**2. pkg/api/handlers.go - Add Spans**
```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
)

func (s *Server) loginHandler(c *gin.Context) {
    tracer := otel.Tracer("ollamamax-api-handlers")
    ctx, span := tracer.Start(c.Request.Context(), "loginHandler")
    defer span.End()

    span.SetAttributes(attribute.String("handler", "login"))

    var req struct { ... }
    if err := c.ShouldBindJSON(&req); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "Invalid request")
        return
    }

    span.SetAttributes(attribute.String("username", req.Username))

    // Authenticate user - pass ctx with trace
    user, err := s.db.Users.Authenticate(ctx, req.Username, req.Password)
    if err != nil {
        span.RecordError(err)
        return
    }

    span.SetStatus(codes.Ok, "Login successful")
}

// Apply similar pattern to all handlers:
// - getUserProfileHandler
// - listModelsHandler
// - createModelHandler
// - getModelHandler
```

### Phase 3: ELK Stack for Log Aggregation (Priority: MEDIUM)

#### Infrastructure Files to Create/Modify:

**1. docker-compose.yml - Add ELK Services**
```yaml
services:
  # Add Jaeger
  jaeger:
    image: jaegertracing/all-in-one:1.51
    ports:
      - "5775:5775/udp"
      - "6831:6831/udp"
      - "6832:6832/udp"
      - "5778:5778"
      - "16686:16686"
      - "14268:14268"
      - "14250:14250"
      - "9411:9411"
    environment:
      - COLLECTOR_ZIPKIN_HOST_PORT=:9411
      - COLLECTOR_OTLP_ENABLED=true
    networks:
      - ollama_network
    restart: unless-stopped

  # Add Elasticsearch
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.11.0
    ports:
      - "9200:9200"
      - "9300:9300"
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data
    networks:
      - ollama_network
    restart: unless-stopped

  # Add Logstash
  logstash:
    image: docker.elastic.co/logstash/logstash:8.11.0
    ports:
      - "5044:5044"
      - "9600:9600"
    volumes:
      - ./monitoring/logstash/pipeline:/usr/share/logstash/pipeline:ro
      - ./monitoring/logstash/config:/usr/share/logstash/config:ro
    depends_on:
      - elasticsearch
    networks:
      - ollama_network
    restart: unless-stopped

  # Add Kibana
  kibana:
    image: docker.elastic.co/kibana/kibana:8.11.0
    ports:
      - "5601:5601"
    environment:
      - ELASTICSEARCH_HOSTS=http://elasticsearch:9200
    depends_on:
      - elasticsearch
    networks:
      - ollama_network
    restart: unless-stopped

  # Add Filebeat
  filebeat:
    image: docker.elastic.co/beats/filebeat:8.11.0
    user: root
    volumes:
      - ./monitoring/filebeat/filebeat.yml:/usr/share/filebeat/filebeat.yml:ro
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
      - /var/run/docker.sock:/var/run/docker.sock:ro
    depends_on:
      - logstash
      - elasticsearch
    networks:
      - ollama_network
    restart: unless-stopped

volumes:
  elasticsearch_data:
    driver: local

# Update ollamamax-api environment
services:
  ollamamax-api:
    environment:
      - JAEGER_ENDPOINT=http://jaeger:14268/api/traces
      - JAEGER_SAMPLER_TYPE=probabilistic
      - JAEGER_SAMPLER_PARAM=0.1
```

**2. monitoring/logstash/pipeline/logstash.conf**
```conf
input {
  beats {
    port => 5044
  }
}

filter {
  json {
    source => "message"
  }

  mutate {
    add_field => {
      "[@metadata][index_name]" => "ollamamax-logs-%{+YYYY.MM.dd}"
    }
  }

  # Route error logs to separate index
  if [level] == "error" {
    mutate {
      replace => { "[@metadata][index_name]" => "ollamamax-errors-%{+YYYY.MM.dd}" }
    }
  }

  # Route audit logs
  if [component] == "audit" {
    mutate {
      replace => { "[@metadata][index_name]" => "ollamamax-audit-%{+YYYY.MM.dd}" }
    }
  }

  # Add geoip
  if [client_ip] {
    geoip {
      source => "client_ip"
    }
  }
}

output {
  elasticsearch {
    hosts => ["http://elasticsearch:9200"]
    index => "%{[@metadata][index_name]}"
    manage_template => true
  }
}
```

**3. monitoring/filebeat/filebeat.yml**
```yaml
filebeat.inputs:
- type: container
  paths:
    - /var/lib/docker/containers/*/*.log
  processors:
    - add_docker_metadata: ~
    - decode_json_fields:
        fields: ["message"]
        target: ""

output.logstash:
  hosts: ["logstash:5044"]
  bulk_max_size: 2048
  compression_level: 3

processors:
  - add_cloud_metadata: ~
  - add_host_metadata: ~
  - drop_fields:
      fields: ["agent.ephemeral_id", "agent.id", "ecs.version"]

logging.level: info
logging.to_files: false

monitoring.enabled: true
monitoring.elasticsearch:
  hosts: ["http://elasticsearch:9200"]
```

### Phase 4: Grafana Dashboards (Priority: MEDIUM)

#### Files to Create:

**1. monitoring/grafana/dashboards/api-performance.json**
```json
{
  "dashboard": {
    "title": "API Performance",
    "panels": [
      {
        "title": "Request Rate",
        "targets": [{
          "expr": "rate(ollamamax_api_http_requests_total[5m])",
          "legendFormat": "{{method}} {{endpoint}}"
        }]
      },
      {
        "title": "Response Time P95",
        "targets": [{
          "expr": "histogram_quantile(0.95, rate(ollamamax_api_http_request_duration_seconds_bucket[5m]))",
          "legendFormat": "{{endpoint}}"
        }]
      },
      {
        "title": "Error Rate",
        "targets": [{
          "expr": "rate(ollamamax_api_http_requests_total{status_code=~\"5..\"}[5m]) / rate(ollamamax_api_http_requests_total[5m]) * 100",
          "legendFormat": "Error Rate %"
        }]
      }
    ]
  }
}
```

**2. monitoring/grafana/dashboards/database-performance.json**
```json
{
  "dashboard": {
    "title": "Database Performance",
    "panels": [
      {
        "title": "Connection Pool Status",
        "targets": [
          {
            "expr": "ollamamax_database_db_connections_active",
            "legendFormat": "Active"
          },
          {
            "expr": "ollamamax_database_db_connections_idle",
            "legendFormat": "Idle"
          }
        ]
      },
      {
        "title": "Query Duration P95",
        "targets": [{
          "expr": "histogram_quantile(0.95, rate(ollamamax_database_db_query_duration_seconds_bucket[5m]))",
          "legendFormat": "{{operation}} {{table}}"
        }]
      },
      {
        "title": "Cache Hit Rate",
        "targets": [{
          "expr": "rate(ollamamax_database_cache_hits_total[5m]) / (rate(ollamamax_database_cache_hits_total[5m]) + rate(ollamamax_database_cache_misses_total[5m])) * 100",
          "legendFormat": "Hit Rate %"
        }]
      }
    ]
  }
}
```

### Phase 5: Alert Configuration (Priority: MEDIUM)

**1. monitoring/alerts.yml - Add New Alerts**
```yaml
groups:
- name: ollamamax-api
  rules:
  - alert: HighAPILatency
    expr: histogram_quantile(0.95, rate(ollamamax_api_http_request_duration_seconds_bucket[5m])) > 1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High API latency detected"
      description: "P95 latency is {{ $value }}s (threshold: 1s)"

  - alert: HighAPIErrorRate
    expr: rate(ollamamax_api_http_requests_total{status_code=~"5.."}[5m]) / rate(ollamamax_api_http_requests_total[5m]) > 0.05
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High API error rate"
      description: "Error rate is {{ $value | humanizePercentage }} (threshold: 5%)"

  - alert: APIEndpointDown
    expr: rate(ollamamax_api_http_requests_total[5m]) == 0
    for: 2m
    labels:
      severity: critical
    annotations:
      summary: "API endpoint not receiving requests"

- name: ollamamax-database
  rules:
  - alert: DatabaseConnectionPoolExhausted
    expr: ollamamax_database_db_connections_active / ollamamax_database_db_connections_max > 0.9
    for: 2m
    labels:
      severity: critical
    annotations:
      summary: "Database connection pool near exhaustion"
      description: "{{ $value | humanizePercentage }} of connections in use"

  - alert: SlowDatabaseQueries
    expr: rate(ollamamax_database_db_query_duration_seconds_sum[5m]) / rate(ollamamax_database_db_query_duration_seconds_count[5m]) > 1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Database queries are slow"
      description: "Average query time: {{ $value }}s"

  - alert: LowCacheHitRate
    expr: rate(ollamamax_database_cache_hits_total[5m]) / (rate(ollamamax_database_cache_hits_total[5m]) + rate(ollamamax_database_cache_misses_total[5m])) < 0.5
    for: 10m
    labels:
      severity: warning
    annotations:
      summary: "Low cache hit rate"
      description: "Hit rate: {{ $value | humanizePercentage }}"

- name: ollamamax-p2p
  rules:
  - alert: LowPeerCount
    expr: ollamamax_p2p_p2p_connected_peers < 2
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Low peer count in P2P network"
      description: "Only {{ $value }} peers connected"

  - alert: HighP2PLatency
    expr: histogram_quantile(0.95, rate(ollamamax_p2p_p2p_message_latency_seconds_bucket[5m])) > 1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High P2P network latency"
      description: "P95 latency: {{ $value }}s"

  - alert: P2PConnectionErrors
    expr: rate(ollamamax_p2p_p2p_connection_errors_total[5m]) > 1
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "High P2P connection error rate"
```

**2. monitoring/alertmanager/alertmanager.yml - Update Notification Channels**
```yaml
global:
  smtp_smarthost: '${SMTP_HOST}:${SMTP_PORT}'
  smtp_from: '${ALERT_EMAIL_FROM}'
  smtp_auth_username: '${SMTP_USER}'
  smtp_auth_password: '${SMTP_PASSWORD}'
  slack_api_url: '${SLACK_WEBHOOK_URL}'

route:
  receiver: 'default'
  group_by: ['alertname', 'cluster', 'service']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h
  routes:
  - match:
      severity: critical
    receiver: 'critical-alerts'
  - match:
      severity: warning
    receiver: 'warning-alerts'
  - match:
      component: p2p
    receiver: 'p2p-alerts'

receivers:
- name: 'default'
  email_configs:
  - to: '${DEFAULT_ALERT_EMAIL}'

- name: 'critical-alerts'
  email_configs:
  - to: '${CRITICAL_ALERT_EMAIL}'
  slack_configs:
  - channel: '${CRITICAL_SLACK_CHANNEL}'
    title: 'CRITICAL: {{ .GroupLabels.alertname }}'
    text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
  pagerduty_configs:
  - service_key: '${PAGERDUTY_SERVICE_KEY}'
    severity: critical
    description: '{{ .GroupLabels.alertname }}: {{ .Annotations.summary }}'

- name: 'warning-alerts'
  email_configs:
  - to: '${WARNING_ALERT_EMAIL}'
  slack_configs:
  - channel: '${WARNING_SLACK_CHANNEL}'

- name: 'p2p-alerts'
  email_configs:
  - to: '${P2P_ALERT_EMAIL}'
  slack_configs:
  - channel: '${P2P_SLACK_CHANNEL}'
```

### Phase 6: Validation & Testing (Priority: MEDIUM)

**1. scripts/validate-monitoring-stack.sh**
```bash
#!/bin/bash
set -e

echo "Validating Monitoring Stack..."

# Check Prometheus
echo "Checking Prometheus..."
curl -f http://localhost:9090/-/healthy || { echo "Prometheus is not healthy"; exit 1; }

# Check Grafana
echo "Checking Grafana..."
curl -f http://localhost:3001/api/health || { echo "Grafana is not healthy"; exit 1; }

# Check Alertmanager
echo "Checking Alertmanager..."
curl -f http://localhost:9093/-/healthy || { echo "Alertmanager is not healthy"; exit 1; }

# Check Jaeger
echo "Checking Jaeger..."
curl -f http://localhost:16686/ || { echo "Jaeger is not healthy"; exit 1; }

# Check Elasticsearch
echo "Checking Elasticsearch..."
curl -f http://localhost:9200/_cluster/health || { echo "Elasticsearch is not healthy"; exit 1; }

# Check Kibana
echo "Checking Kibana..."
curl -f http://localhost:5601/api/status || { echo "Kibana is not healthy"; exit 1; }

# Validate metrics
echo "Validating metrics..."
metrics=$(curl -s http://localhost:9090/api/v1/query?query=up)
if echo "$metrics" | grep -q '"status":"success"'; then
    echo "✓ Prometheus metrics available"
else
    echo "✗ Prometheus metrics not available"
    exit 1
fi

# Validate traces
echo "Validating traces..."
traces=$(curl -s http://localhost:16686/api/traces?service=ollamamax-api&limit=1)
if echo "$traces" | grep -q '"data"'; then
    echo "✓ Jaeger traces available"
else
    echo "✗ Jaeger traces not available"
    exit 1
fi

# Validate logs
echo "Validating logs..."
logs=$(curl -s http://localhost:9200/ollamamax-logs-*/_count)
if echo "$logs" | grep -q '"count"'; then
    echo "✓ Elasticsearch logs available"
else
    echo "✗ Elasticsearch logs not available"
    exit 1
fi

echo "✓ All monitoring services are healthy"
```

**2. scripts/test-alert-notifications.sh**
```bash
#!/bin/bash

# Test Slack notification
test_slack() {
    echo "Testing Slack notification..."
    curl -X POST "${SLACK_WEBHOOK_URL}" \
         -H 'Content-Type: application/json' \
         -d '{"text": "Test alert from OllamaMax monitoring stack"}' \
         && echo "✓ Slack notification sent" \
         || echo "✗ Slack notification failed"
}

# Test Email notification
test_email() {
    echo "Testing Email notification..."
    # Implementation depends on mail command availability
    echo "Email test requires manual verification"
}

# Test PagerDuty
test_pagerduty() {
    echo "Testing PagerDuty notification..."
    curl -X POST https://events.pagerduty.com/v2/enqueue \
         -H 'Content-Type: application/json' \
         -d "{
             \"routing_key\": \"${PAGERDUTY_SERVICE_KEY}\",
             \"event_action\": \"trigger\",
             \"payload\": {
                 \"summary\": \"Test alert from OllamaMax\",
                 \"severity\": \"info\",
                 \"source\": \"monitoring-test\"
             }
         }" \
         && echo "✓ PagerDuty notification sent" \
         || echo "✗ PagerDuty notification failed"
}

# Run tests
test_slack
test_email
test_pagerduty
```

### Phase 7: Configuration Updates

**1. .env.example - Add Monitoring Variables**
```env
# Jaeger Tracing
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
JAEGER_SAMPLER_TYPE=probabilistic
JAEGER_SAMPLER_PARAM=0.1
JAEGER_SERVICE_NAME=ollamamax

# Prometheus Metrics
PROMETHEUS_PORT=9090
METRICS_NAMESPACE=ollamamax
METRICS_ENABLED=true

# Slack Notifications
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK
CRITICAL_SLACK_CHANNEL=#ollamamax-critical
WARNING_SLACK_CHANNEL=#ollamamax-warnings
P2P_SLACK_CHANNEL=#ollamamax-p2p

# Email Notifications
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=alerts@company.com
SMTP_PASSWORD=your_smtp_password
ALERT_EMAIL_FROM=alerts@company.com
CRITICAL_ALERT_EMAIL=oncall@company.com
WARNING_ALERT_EMAIL=monitoring@company.com
P2P_ALERT_EMAIL=network@company.com

# PagerDuty Integration
PAGERDUTY_SERVICE_KEY=your_pagerduty_integration_key

# Elasticsearch
ELASTICSEARCH_HOST=elasticsearch
ELASTICSEARCH_PORT=9200

# Logstash
LOGSTASH_HOST=logstash
LOGSTASH_PORT=5044

# Kibana
KIBANA_HOST=kibana
KIBANA_PORT=5601
```

**2. package.json - Add Monitoring Scripts**
```json
{
  "scripts": {
    "monitoring:start": "docker-compose up -d prometheus grafana alertmanager jaeger elasticsearch logstash kibana filebeat",
    "monitoring:stop": "docker-compose stop prometheus grafana alertmanager jaeger elasticsearch logstash kibana filebeat",
    "monitoring:logs": "docker-compose logs -f prometheus grafana alertmanager jaeger",
    "monitoring:validate": "bash scripts/validate-monitoring-stack.sh",
    "alerts:test": "bash scripts/test-alert-notifications.sh --all",
    "alerts:test:slack": "bash scripts/test-alert-notifications.sh --slack",
    "alerts:test:email": "bash scripts/test-alert-notifications.sh --email",
    "alerts:test:pagerduty": "bash scripts/test-alert-notifications.sh --pagerduty"
  }
}
```

## Implementation Order

1. **Week 1**: Phase 1 (Prometheus Metrics) - Files: pkg/api/server.go, internal/server/server.go, pkg/database/manager.go
2. **Week 2**: Phase 1 continued - Files: pkg/p2p/node.go, pkg/distributed/load_balancer.go, pkg/database/repositories.go
3. **Week 3**: Phase 2 (Jaeger Tracing) - Files: pkg/api/server.go, pkg/api/handlers.go
4. **Week 4**: Phase 3 (ELK Stack) - docker-compose.yml, Logstash/Filebeat configs
5. **Week 5**: Phase 4 (Grafana Dashboards) - Dashboard JSON files
6. **Week 6**: Phase 5 (Alerts) - alerts.yml, alertmanager.yml
7. **Week 7**: Phase 6 (Validation) - Test scripts, CI/CD integration

## Success Metrics

- [ ] All API endpoints expose /metrics endpoint
- [ ] Prometheus scraping all metrics (up{job="ollamamax-api"} == 1)
- [ ] Grafana dashboards display real-time metrics
- [ ] Jaeger UI shows distributed traces
- [ ] Kibana displays application logs
- [ ] Alert notifications delivered to Slack/Email/PagerDuty
- [ ] Monitoring stack validation script passes
- [ ] CI/CD pipeline includes monitoring validation

## Rollback Plan

1. Keep existing monitoring configuration
2. Deploy new monitoring stack in parallel
3. Verify new stack for 1 week
4. Gradually migrate dashboards/alerts
5. Decommission old stack only after full verification

## Documentation

- Create MONITORING_IMPLEMENTATION_GUIDE.md with developer instructions
- Update production-monitoring.yaml with new architecture
- Document alert escalation procedures
- Create runbook for common monitoring issues

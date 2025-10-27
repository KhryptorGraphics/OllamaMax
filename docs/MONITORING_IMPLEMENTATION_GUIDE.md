# Monitoring Implementation Guide

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [API Compatibility](#api-compatibility)
4. [Metrics Catalog](#metrics-catalog)
5. [Installation](#installation)
6. [Configuration](#configuration)
7. [Dashboards](#dashboards)
8. [Alerts](#alerts)
9. [Tracing](#tracing)
10. [Logs](#logs)
11. [Testing](#testing)
12. [Troubleshooting](#troubleshooting)
13. [Performance](#performance)
14. [Security](#security)
15. [Best Practices](#best-practices)

---

## Overview

This monitoring implementation provides comprehensive observability for OllamaMax distributed inference platform using industry-standard tools:

- **Metrics**: Prometheus for collection, Grafana for visualization
- **Tracing**: Jaeger for distributed request tracing
- **Logging**: ELK Stack (Elasticsearch, Logstash, Kibana)
- **Alerting**: Prometheus Alertmanager with multi-channel notifications

### Key Features

- Real-time metrics collection (1-15s intervals)
- Distributed tracing across all services
- Centralized log aggregation and search
- Intelligent alerting with deduplication
- Multi-format API endpoints for compatibility
- Auto-scaling integration with HPA
- Production-grade security and authentication

### Architecture Goals

- **High Availability**: Redundant monitoring infrastructure
- **Low Overhead**: <2% performance impact on monitored services
- **Scalability**: Handles 10k+ requests/second
- **Flexibility**: Multiple data export formats
- **Security**: TLS, authentication, secrets management

---

## Architecture

### System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                        MONITORING STACK                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────────┐  │
│  │   OllamaMax  │      │  Load        │      │   Model      │  │
│  │   Instances  │─────▶│  Balancer    │─────▶│   Servers    │  │
│  └──────┬───────┘      └──────┬───────┘      └──────┬───────┘  │
│         │                     │                      │           │
│         │ /metrics           │ /metrics            │ /metrics   │
│         │ /metrics.json      │ /metrics.json       │ /metrics.json│
│         ▼                     ▼                      ▼           │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              Prometheus (Metrics Collection)              │  │
│  │  • Service Discovery                                      │  │
│  │  • Time-series Database                                   │  │
│  │  • PromQL Query Engine                                    │  │
│  │  • Retention: 15d (configurable)                          │  │
│  └──────────────┬───────────────────────────┬────────────────┘  │
│                 │                           │                    │
│                 ▼                           ▼                    │
│  ┌──────────────────────┐      ┌──────────────────────┐        │
│  │  Grafana Dashboards  │      │   Alertmanager       │        │
│  │  • Real-time Graphs  │      │  • Alert Routing     │        │
│  │  • Custom Queries    │      │  • Deduplication     │        │
│  │  • Multi-tenant      │      │  • Notifications     │        │
│  └──────────────────────┘      └──────────┬───────────┘        │
│                                            │                    │
│                                            ▼                    │
│                              ┌──────────────────────┐          │
│                              │  Notification        │          │
│                              │  Channels            │          │
│                              │  • Slack             │          │
│                              │  • Email             │          │
│                              │  • PagerDuty         │          │
│                              │  • Webhook           │          │
│                              └──────────────────────┘          │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              Jaeger (Distributed Tracing)                 │  │
│  │  • Span Collection                                        │  │
│  │  • Trace Aggregation                                      │  │
│  │  • Dependency Analysis                                    │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              ELK Stack (Centralized Logging)              │  │
│  │  • Elasticsearch: Log Storage & Search                    │  │
│  │  • Logstash: Log Processing & Enrichment                  │  │
│  │  • Kibana: Log Visualization & Analysis                   │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Component Descriptions

#### Prometheus
- **Purpose**: Time-series metrics database and query engine
- **Port**: 9090
- **Scrape Interval**: 15s (configurable)
- **Retention**: 15 days
- **Storage**: Local or remote (Thanos, Cortex)

#### Grafana
- **Purpose**: Metrics visualization and dashboards
- **Port**: 3000
- **Authentication**: OAuth2, LDAP, or built-in
- **Datasources**: Prometheus, Jaeger, Elasticsearch

#### Alertmanager
- **Purpose**: Alert routing and notification management
- **Port**: 9093
- **Features**: Grouping, inhibition, silencing
- **Integrations**: 20+ notification channels

#### Jaeger
- **Purpose**: Distributed tracing for request flows
- **Port**: 16686 (UI), 14268 (collector)
- **Storage**: Elasticsearch, Cassandra, Memory
- **Sampling**: Adaptive sampling for high-volume

#### ELK Stack
- **Elasticsearch**: 9200 (REST), 9300 (transport)
- **Logstash**: 5044 (Beats), 9600 (monitoring)
- **Kibana**: 5601 (UI)
- **Log Retention**: 7 days (configurable)

### Data Flow

1. **Metrics Collection**:
   - Services expose /metrics endpoints
   - Prometheus scrapes endpoints every 15s
   - Data stored in time-series database
   - Grafana queries Prometheus for visualization

2. **Distributed Tracing**:
   - Applications instrument code with OpenTelemetry
   - Spans sent to Jaeger collector
   - Traces stored and indexed
   - UI provides search and analysis

3. **Log Aggregation**:
   - Applications write structured logs
   - Filebeat ships logs to Logstash
   - Logstash enriches and forwards to Elasticsearch
   - Kibana provides search and dashboards

4. **Alerting**:
   - Prometheus evaluates alert rules
   - Alerts sent to Alertmanager
   - Alertmanager routes to notification channels
   - Incidents tracked and resolved

---

## API Compatibility

### Dual-Format Metrics Endpoints

OllamaMax provides two metrics endpoints to support both Prometheus scraping and legacy JSON consumers:

#### `/metrics` - Prometheus Text Format

**Purpose**: Primary endpoint for Prometheus scraping

**Format**: OpenMetrics/Prometheus exposition format (text-based)

**Content-Type**: `text/plain; version=0.0.4; charset=utf-8`

**Example Response**:
```prometheus
# HELP ollama_request_duration_seconds Request duration in seconds
# TYPE ollama_request_duration_seconds histogram
ollama_request_duration_seconds_bucket{method="POST",endpoint="/api/generate",le="0.1"} 145
ollama_request_duration_seconds_bucket{method="POST",endpoint="/api/generate",le="0.5"} 892
ollama_request_duration_seconds_bucket{method="POST",endpoint="/api/generate",le="1.0"} 1456
ollama_request_duration_seconds_bucket{method="POST",endpoint="/api/generate",le="+Inf"} 1500
ollama_request_duration_seconds_sum{method="POST",endpoint="/api/generate"} 678.4
ollama_request_duration_seconds_count{method="POST",endpoint="/api/generate"} 1500

# HELP ollama_active_requests Current number of active requests
# TYPE ollama_active_requests gauge
ollama_active_requests{instance="worker-1"} 12

# HELP ollama_requests_total Total number of requests
# TYPE ollama_requests_total counter
ollama_requests_total{method="POST",endpoint="/api/generate",status="200"} 1456
ollama_requests_total{method="POST",endpoint="/api/generate",status="500"} 44
```

**Usage**:
```yaml
# Prometheus scrape configuration
scrape_configs:
  - job_name: 'ollamamax'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

**Advantages**:
- Efficient text-based format
- Native Prometheus support
- Industry standard (OpenMetrics)
- Low parsing overhead

---

#### `/metrics.json` - JSON Format

**Purpose**: Backward compatibility for existing JSON consumers

**Format**: Structured JSON with nested metric objects

**Content-Type**: `application/json; charset=utf-8`

**Example Response**:
```json
{
  "timestamp": "2025-10-27T10:30:45Z",
  "instance": "worker-1",
  "metrics": {
    "request_duration": {
      "type": "histogram",
      "help": "Request duration in seconds",
      "values": {
        "POST:/api/generate": {
          "buckets": [
            {"le": 0.1, "count": 145},
            {"le": 0.5, "count": 892},
            {"le": 1.0, "count": 1456},
            {"le": "+Inf", "count": 1500}
          ],
          "sum": 678.4,
          "count": 1500
        }
      }
    },
    "active_requests": {
      "type": "gauge",
      "help": "Current number of active requests",
      "value": 12,
      "labels": {
        "instance": "worker-1"
      }
    },
    "requests_total": {
      "type": "counter",
      "help": "Total number of requests",
      "values": {
        "POST:/api/generate:200": 1456,
        "POST:/api/generate:500": 44
      }
    },
    "model_load_time_seconds": {
      "type": "gauge",
      "help": "Time taken to load model",
      "value": 2.34,
      "labels": {
        "model": "llama2:7b"
      }
    },
    "gpu_utilization": {
      "type": "gauge",
      "help": "GPU utilization percentage",
      "value": 78.5,
      "labels": {
        "gpu_id": "0"
      }
    },
    "memory_usage_bytes": {
      "type": "gauge",
      "help": "Current memory usage in bytes",
      "value": 8589934592,
      "labels": {
        "type": "heap"
      }
    }
  }
}
```

**Usage**:
```bash
# Fetch JSON metrics
curl -H "Accept: application/json" http://localhost:8080/metrics.json

# Parse with jq
curl -s http://localhost:8080/metrics.json | jq '.metrics.active_requests.value'
```

**Client Example**:
```javascript
// Node.js client
const axios = require('axios');

async function fetchMetrics() {
  const response = await axios.get('http://localhost:8080/metrics.json');
  const metrics = response.data.metrics;

  console.log(`Active Requests: ${metrics.active_requests.value}`);
  console.log(`GPU Utilization: ${metrics.gpu_utilization.value}%`);

  return metrics;
}
```

**Python Example**:
```python
import requests

def fetch_metrics():
    response = requests.get('http://localhost:8080/metrics.json')
    metrics = response.json()['metrics']

    print(f"Active Requests: {metrics['active_requests']['value']}")
    print(f"GPU Utilization: {metrics['gpu_utilization']['value']}%")

    return metrics
```

---

### Migration Guide

#### For Existing JSON Consumers

If you currently consume metrics in JSON format, follow these steps:

**1. Update Endpoint URL**:
```diff
- GET /metrics
+ GET /metrics.json
```

**2. Update Accept Header** (optional but recommended):
```javascript
// Before
fetch('/metrics')

// After
fetch('/metrics.json', {
  headers: { 'Accept': 'application/json' }
})
```

**3. Validate Response Structure**:
```javascript
// Example validation
async function validateMetrics(url) {
  const response = await fetch(url);
  const data = await response.json();

  if (!data.metrics || !data.timestamp) {
    throw new Error('Invalid metrics format');
  }

  return data;
}
```

**4. Update Monitoring Dashboards**:
- If using custom dashboards, update data source URLs
- Test queries against /metrics.json endpoint
- Verify all metrics are available

**5. Testing Checklist**:
```bash
# Verify endpoint availability
curl -f http://localhost:8080/metrics.json

# Compare metrics count
OLD_COUNT=$(curl -s http://localhost:8080/old-metrics | jq '.metrics | length')
NEW_COUNT=$(curl -s http://localhost:8080/metrics.json | jq '.metrics | length')

# Verify specific metrics
curl -s http://localhost:8080/metrics.json | jq '.metrics.active_requests'
```

#### For Prometheus Users

No changes required! Continue using `/metrics` endpoint:

```yaml
scrape_configs:
  - job_name: 'ollamamax'
    metrics_path: '/metrics'  # Keep using this
    static_configs:
      - targets: ['localhost:8080']
```

#### Choosing the Right Endpoint

| Use Case | Endpoint | Format | Best For |
|----------|----------|--------|----------|
| Prometheus scraping | `/metrics` | Text | Time-series monitoring |
| Custom dashboards | `/metrics.json` | JSON | Application integration |
| API consumers | `/metrics.json` | JSON | Programmatic access |
| Grafana | `/metrics` | Text | Standard dashboards |
| Log aggregation | `/metrics.json` | JSON | ELK/Splunk ingestion |
| Health checks | `/metrics.json` | JSON | Simple parsing |

#### Performance Considerations

- **Text Format**: 30-40% smaller response size
- **JSON Format**: Easier to parse in applications
- **Both formats** generated on-demand with minimal overhead (<1ms)
- Caching recommended for high-traffic scenarios

#### Content Negotiation

The server supports automatic format selection via Accept header:

```bash
# Request text format
curl -H "Accept: text/plain" http://localhost:8080/metrics

# Request JSON format
curl -H "Accept: application/json" http://localhost:8080/metrics
# Automatically redirects to /metrics.json

# Request both (prioritizes text)
curl -H "Accept: text/plain, application/json" http://localhost:8080/metrics
```

---

## Metrics Catalog

### Request Metrics

#### `ollama_request_duration_seconds`
- **Type**: Histogram
- **Description**: Request duration distribution
- **Labels**: `method`, `endpoint`, `status`
- **Buckets**: 0.1, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, +Inf
- **Use Case**: Identify slow endpoints, SLA tracking

**PromQL Examples**:
```promql
# 95th percentile latency
histogram_quantile(0.95, rate(ollama_request_duration_seconds_bucket[5m]))

# Average latency by endpoint
rate(ollama_request_duration_seconds_sum[5m]) / rate(ollama_request_duration_seconds_count[5m])

# Requests exceeding 5s SLA
sum(rate(ollama_request_duration_seconds_bucket{le="5.0"}[5m])) by (endpoint)
```

#### `ollama_requests_total`
- **Type**: Counter
- **Description**: Total number of requests processed
- **Labels**: `method`, `endpoint`, `status`
- **Use Case**: Traffic analysis, error rate calculation

**PromQL Examples**:
```promql
# Requests per second
rate(ollama_requests_total[1m])

# Error rate (4xx + 5xx)
sum(rate(ollama_requests_total{status=~"[45].."}[5m])) / sum(rate(ollama_requests_total[5m]))

# Success rate by endpoint
sum(rate(ollama_requests_total{status="200"}[5m])) by (endpoint)
```

#### `ollama_active_requests`
- **Type**: Gauge
- **Description**: Current number of in-flight requests
- **Labels**: `instance`, `worker_id`
- **Use Case**: Load monitoring, capacity planning

**PromQL Examples**:
```promql
# Total active requests across cluster
sum(ollama_active_requests)

# Max concurrent requests per instance
max(ollama_active_requests) by (instance)

# Instances approaching capacity
ollama_active_requests / ollama_max_concurrent_requests > 0.8
```

### Model Performance Metrics

#### `ollama_model_load_time_seconds`
- **Type**: Gauge
- **Description**: Time taken to load model into memory
- **Labels**: `model`, `instance`
- **Use Case**: Model startup optimization

**PromQL Examples**:
```promql
# Average load time by model
avg(ollama_model_load_time_seconds) by (model)

# Slowest model loads in last hour
topk(10, max_over_time(ollama_model_load_time_seconds[1h]))
```

#### `ollama_inference_duration_seconds`
- **Type**: Histogram
- **Description**: Time to generate model response
- **Labels**: `model`, `instance`
- **Buckets**: 0.5, 1.0, 2.0, 5.0, 10.0, 30.0, 60.0, +Inf
- **Use Case**: Model performance tracking

**PromQL Examples**:
```promql
# 99th percentile inference time
histogram_quantile(0.99, rate(ollama_inference_duration_seconds_bucket[5m]))

# Compare inference times across models
avg(rate(ollama_inference_duration_seconds_sum[5m])) by (model)
```

#### `ollama_tokens_processed_total`
- **Type**: Counter
- **Description**: Total tokens processed (input + output)
- **Labels**: `model`, `instance`, `type` (input/output)
- **Use Case**: Token throughput analysis

**PromQL Examples**:
```promql
# Tokens per second
rate(ollama_tokens_processed_total[1m])

# Input/output token ratio
sum(rate(ollama_tokens_processed_total{type="output"}[5m])) / sum(rate(ollama_tokens_processed_total{type="input"}[5m]))
```

### Resource Metrics

#### `ollama_gpu_utilization_percent`
- **Type**: Gauge
- **Description**: GPU utilization percentage (0-100)
- **Labels**: `gpu_id`, `instance`
- **Use Case**: GPU resource monitoring

**PromQL Examples**:
```promql
# Average GPU utilization across cluster
avg(ollama_gpu_utilization_percent)

# Underutilized GPUs (<50%)
ollama_gpu_utilization_percent < 50

# Peak GPU usage in last hour
max_over_time(ollama_gpu_utilization_percent[1h])
```

#### `ollama_gpu_memory_used_bytes`
- **Type**: Gauge
- **Description**: GPU memory currently in use
- **Labels**: `gpu_id`, `instance`
- **Use Case**: Memory capacity planning

**PromQL Examples**:
```promql
# GPU memory usage percentage
(ollama_gpu_memory_used_bytes / ollama_gpu_memory_total_bytes) * 100

# Available GPU memory
ollama_gpu_memory_total_bytes - ollama_gpu_memory_used_bytes
```

#### `ollama_memory_usage_bytes`
- **Type**: Gauge
- **Description**: System memory usage
- **Labels**: `type` (heap/rss/external), `instance`
- **Use Case**: Memory leak detection

**PromQL Examples**:
```promql
# Memory growth rate
rate(ollama_memory_usage_bytes{type="heap"}[10m])

# Instances using >80% memory
ollama_memory_usage_bytes / ollama_memory_total_bytes > 0.8
```

#### `ollama_cpu_usage_percent`
- **Type**: Gauge
- **Description**: CPU utilization percentage
- **Labels**: `instance`
- **Use Case**: CPU resource monitoring

### Queue Metrics

#### `ollama_queue_size`
- **Type**: Gauge
- **Description**: Current number of requests in queue
- **Labels**: `instance`, `priority`
- **Use Case**: Backlog monitoring

#### `ollama_queue_wait_time_seconds`
- **Type**: Histogram
- **Description**: Time requests spend in queue
- **Labels**: `instance`, `priority`
- **Buckets**: 0.1, 0.5, 1.0, 5.0, 10.0, 30.0, +Inf

### Load Balancing Metrics

#### `ollama_lb_backend_health`
- **Type**: Gauge
- **Description**: Backend health status (0=down, 1=up)
- **Labels**: `backend`, `region`
- **Use Case**: Backend availability monitoring

#### `ollama_lb_requests_routed_total`
- **Type**: Counter
- **Description**: Requests routed to each backend
- **Labels**: `backend`, `algorithm`
- **Use Case**: Load distribution analysis

### Cache Metrics

#### `ollama_cache_hits_total`
- **Type**: Counter
- **Description**: Number of cache hits
- **Labels**: `cache_type`, `instance`

#### `ollama_cache_size_bytes`
- **Type**: Gauge
- **Description**: Current cache size
- **Labels**: `cache_type`, `instance`

---

## Installation

### Prerequisites

- Docker 20.10+ or Kubernetes 1.20+
- 4GB RAM minimum for monitoring stack
- 20GB storage for metrics/logs retention
- Network connectivity to scrape targets

### Docker Compose Deployment

**1. Create monitoring directory**:
```bash
mkdir -p monitoring/{prometheus,grafana,alertmanager,jaeger,elk}
cd monitoring
```

**2. Download docker-compose.yml**:
```yaml
# docker-compose.yml
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:v2.47.0
    container_name: prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
      - ./prometheus/alerts.yml:/etc/prometheus/alerts.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--storage.tsdb.retention.time=15d'
      - '--web.enable-lifecycle'
    restart: unless-stopped
    networks:
      - monitoring

  grafana:
    image: grafana/grafana:10.1.5
    container_name: grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false
      - GF_INSTALL_PLUGINS=grafana-piechart-panel
    volumes:
      - ./grafana/provisioning:/etc/grafana/provisioning
      - ./grafana/dashboards:/var/lib/grafana/dashboards
      - grafana_data:/var/lib/grafana
    depends_on:
      - prometheus
    restart: unless-stopped
    networks:
      - monitoring

  alertmanager:
    image: prom/alertmanager:v0.26.0
    container_name: alertmanager
    ports:
      - "9093:9093"
    volumes:
      - ./alertmanager/alertmanager.yml:/etc/alertmanager/alertmanager.yml
      - alertmanager_data:/alertmanager
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
      - '--storage.path=/alertmanager'
    restart: unless-stopped
    networks:
      - monitoring

  jaeger:
    image: jaegertracing/all-in-one:1.50
    container_name: jaeger
    ports:
      - "16686:16686"  # UI
      - "14268:14268"  # Collector
      - "6831:6831/udp"  # Agent
    environment:
      - COLLECTOR_OTLP_ENABLED=true
      - SPAN_STORAGE_TYPE=elasticsearch
      - ES_SERVER_URLS=http://elasticsearch:9200
    depends_on:
      - elasticsearch
    restart: unless-stopped
    networks:
      - monitoring

  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.10.2
    container_name: elasticsearch
    environment:
      - discovery.type=single-node
      - "ES_JAVA_OPTS=-Xms2g -Xmx2g"
      - xpack.security.enabled=false
    ports:
      - "9200:9200"
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data
    restart: unless-stopped
    networks:
      - monitoring

  logstash:
    image: docker.elastic.co/logstash/logstash:8.10.2
    container_name: logstash
    ports:
      - "5044:5044"
      - "9600:9600"
    volumes:
      - ./logstash/logstash.conf:/usr/share/logstash/pipeline/logstash.conf
      - ./logstash/logstash.yml:/usr/share/logstash/config/logstash.yml
    depends_on:
      - elasticsearch
    restart: unless-stopped
    networks:
      - monitoring

  kibana:
    image: docker.elastic.co/kibana/kibana:8.10.2
    container_name: kibana
    ports:
      - "5601:5601"
    environment:
      - ELASTICSEARCH_HOSTS=http://elasticsearch:9200
    depends_on:
      - elasticsearch
    restart: unless-stopped
    networks:
      - monitoring

volumes:
  prometheus_data:
  grafana_data:
  alertmanager_data:
  elasticsearch_data:

networks:
  monitoring:
    driver: bridge
```

**3. Create Prometheus configuration**:
```yaml
# prometheus/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    cluster: 'ollamamax-prod'
    environment: 'production'

alerting:
  alertmanagers:
    - static_configs:
        - targets: ['alertmanager:9093']

rule_files:
  - 'alerts.yml'

scrape_configs:
  # OllamaMax instances
  - job_name: 'ollamamax'
    static_configs:
      - targets:
          - 'host.docker.internal:8080'
          - 'host.docker.internal:8081'
          - 'host.docker.internal:8082'
    metrics_path: '/metrics'
    scrape_interval: 15s

  # Load balancer
  - job_name: 'load-balancer'
    static_configs:
      - targets: ['host.docker.internal:8000']
    metrics_path: '/metrics'

  # Prometheus self-monitoring
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  # Node exporters (if deployed)
  - job_name: 'node-exporter'
    static_configs:
      - targets:
          - 'node-exporter-1:9100'
          - 'node-exporter-2:9100'
```

**4. Start monitoring stack**:
```bash
docker-compose up -d

# Verify services
docker-compose ps

# Check logs
docker-compose logs -f prometheus
docker-compose logs -f grafana
```

**5. Access UIs**:
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)
- Alertmanager: http://localhost:9093
- Jaeger: http://localhost:16686
- Kibana: http://localhost:5601

### Kubernetes Deployment

**1. Create namespace**:
```bash
kubectl create namespace monitoring
```

**2. Deploy Prometheus Operator**:
```bash
# Add Helm repo
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Install Prometheus Operator
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --set prometheus.prometheusSpec.retention=15d \
  --set prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage=50Gi \
  --set grafana.adminPassword=admin \
  --set alertmanager.enabled=true
```

**3. Deploy ServiceMonitor for OllamaMax**:
```yaml
# monitoring/servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: ollamamax
  namespace: monitoring
  labels:
    app: ollamamax
spec:
  selector:
    matchLabels:
      app: ollamamax
  endpoints:
    - port: metrics
      path: /metrics
      interval: 15s
```

```bash
kubectl apply -f monitoring/servicemonitor.yaml
```

**4. Deploy Jaeger**:
```bash
helm repo add jaegertracing https://jaegertracing.github.io/helm-charts
helm install jaeger jaegertracing/jaeger \
  --namespace monitoring \
  --set provisionDataStore.cassandra=false \
  --set storage.type=elasticsearch
```

**5. Deploy ELK Stack**:
```bash
helm repo add elastic https://helm.elastic.co
helm install elasticsearch elastic/elasticsearch --namespace monitoring
helm install kibana elastic/kibana --namespace monitoring
helm install filebeat elastic/filebeat --namespace monitoring
```

**6. Verify deployment**:
```bash
kubectl get pods -n monitoring
kubectl get svc -n monitoring

# Port-forward to access UIs
kubectl port-forward -n monitoring svc/prometheus-operated 9090:9090
kubectl port-forward -n monitoring svc/grafana 3000:80
```

---

## Configuration

### Environment Variables

#### OllamaMax Application

```bash
# Metrics configuration
METRICS_ENABLED=true
METRICS_PORT=8080
METRICS_PATH=/metrics
METRICS_JSON_PATH=/metrics.json

# Tracing configuration
TRACING_ENABLED=true
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
JAEGER_SAMPLER_TYPE=probabilistic
JAEGER_SAMPLER_PARAM=0.1

# Logging configuration
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=stdout

# Performance tuning
METRICS_BUFFER_SIZE=10000
METRICS_FLUSH_INTERVAL=10s
```

#### Prometheus

```bash
# Retention and storage
STORAGE_TSDB_RETENTION_TIME=15d
STORAGE_TSDB_RETENTION_SIZE=50GB

# Query performance
QUERY_TIMEOUT=2m
QUERY_MAX_SAMPLES=50000000
```

#### Grafana

```bash
# Authentication
GF_SECURITY_ADMIN_USER=admin
GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD}
GF_AUTH_ANONYMOUS_ENABLED=false

# Database
GF_DATABASE_TYPE=postgres
GF_DATABASE_HOST=postgres:5432
GF_DATABASE_NAME=grafana
GF_DATABASE_USER=grafana
GF_DATABASE_PASSWORD=${DB_PASSWORD}

# SMTP for alerts
GF_SMTP_ENABLED=true
GF_SMTP_HOST=smtp.example.com:587
GF_SMTP_USER=${SMTP_USER}
GF_SMTP_PASSWORD=${SMTP_PASSWORD}
```

### Secrets Management

**Using Docker Secrets**:
```bash
# Create secrets
echo "my-secure-password" | docker secret create grafana_password -
echo "smtp-password" | docker secret create smtp_password -

# Update docker-compose.yml
services:
  grafana:
    secrets:
      - grafana_password
    environment:
      - GF_SECURITY_ADMIN_PASSWORD_FILE=/run/secrets/grafana_password

secrets:
  grafana_password:
    external: true
```

**Using Kubernetes Secrets**:
```bash
# Create secret
kubectl create secret generic monitoring-secrets \
  --from-literal=grafana-password=my-secure-password \
  --from-literal=smtp-password=smtp-password \
  -n monitoring

# Reference in deployment
apiVersion: v1
kind: Pod
metadata:
  name: grafana
spec:
  containers:
    - name: grafana
      env:
        - name: GF_SECURITY_ADMIN_PASSWORD
          valueFrom:
            secretKeyRef:
              name: monitoring-secrets
              key: grafana-password
```

### Alert Routing Configuration

```yaml
# alertmanager/alertmanager.yml
global:
  resolve_timeout: 5m
  slack_api_url: 'https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK'

route:
  group_by: ['alertname', 'cluster', 'service']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h
  receiver: 'default'
  routes:
    # Critical alerts to PagerDuty
    - match:
        severity: critical
      receiver: 'pagerduty'
      continue: true

    # All alerts to Slack
    - match_re:
        severity: (warning|critical)
      receiver: 'slack'

receivers:
  - name: 'default'
    email_configs:
      - to: 'alerts@example.com'
        from: 'alertmanager@example.com'
        smarthost: 'smtp.example.com:587'
        auth_username: 'alerts@example.com'
        auth_password: '${SMTP_PASSWORD}'

  - name: 'slack'
    slack_configs:
      - channel: '#alerts'
        title: '{{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'
        send_resolved: true

  - name: 'pagerduty'
    pagerduty_configs:
      - service_key: '${PAGERDUTY_SERVICE_KEY}'
        description: '{{ .GroupLabels.alertname }}'

inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'instance']
```

---

## Dashboards

### Overview Dashboard

**Purpose**: High-level cluster health at a glance

**Key Panels**:
1. **Total Requests/sec**: Line graph showing request rate
2. **Error Rate**: Gauge showing percentage of failed requests
3. **Average Latency**: Single stat with sparkline
4. **Active Instances**: Table of healthy vs unhealthy instances
5. **GPU Utilization**: Heat map across all GPUs
6. **Queue Depth**: Bar chart of queued requests

**PromQL Queries**:
```promql
# Total requests/sec
sum(rate(ollama_requests_total[1m]))

# Error rate percentage
100 * sum(rate(ollama_requests_total{status=~"[45].."}[5m])) / sum(rate(ollama_requests_total[5m]))

# Average latency
avg(rate(ollama_request_duration_seconds_sum[5m])) / avg(rate(ollama_request_duration_seconds_count[5m]))

# Active instances
count(up{job="ollamamax"} == 1)

# Average GPU utilization
avg(ollama_gpu_utilization_percent)
```

**Screenshot Placeholder**:
```
┌─────────────────────────────────────────────────────────────┐
│  OllamaMax Cluster Overview                                 │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  [Requests/sec: 1,247 ↑]  [Error Rate: 0.2%]  [Latency: 450ms] │
│                                                             │
│  Active Instances: 8/10   Queue Depth: 12   GPU Util: 78%  │
│                                                             │
│  ┌───────────────────────┐  ┌───────────────────────┐      │
│  │  Request Rate         │  │  Error Distribution   │      │
│  │  [Line Graph]         │  │  [Pie Chart]          │      │
│  └───────────────────────┘  └───────────────────────┘      │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  GPU Utilization Heatmap                            │   │
│  │  [GPU0] ████████░░ 78%                              │   │
│  │  [GPU1] ██████████ 95%                              │   │
│  │  [GPU2] █████░░░░░ 52%                              │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### Request Performance Dashboard

**Purpose**: Deep dive into request latency and throughput

**Key Panels**:
1. **Latency Percentiles**: P50, P95, P99 over time
2. **Request Rate by Endpoint**: Stacked area chart
3. **Slow Requests**: Table of slowest requests (>5s)
4. **Request Duration Heatmap**: Distribution of latencies
5. **Cache Hit Rate**: Line graph of cache effectiveness

### Model Performance Dashboard

**Purpose**: Monitor individual model performance

**Key Panels**:
1. **Inference Time by Model**: Bar chart comparison
2. **Token Throughput**: Tokens/sec per model
3. **Model Load Times**: Gauge showing cold start performance
4. **Model Memory Usage**: Stacked bar chart
5. **Active Model Instances**: Table with status

### Resource Utilization Dashboard

**Purpose**: Infrastructure resource monitoring

**Key Panels**:
1. **CPU Usage**: Line graph per instance
2. **Memory Usage**: Stacked area chart (heap/rss)
3. **GPU Utilization**: Multi-series line graph
4. **GPU Memory**: Bar chart of used vs available
5. **Network I/O**: Upload/download rates
6. **Disk I/O**: Read/write rates

### Load Balancer Dashboard

**Purpose**: Monitor load distribution and backend health

**Key Panels**:
1. **Backend Health**: Status grid (green/red indicators)
2. **Request Distribution**: Pie chart by backend
3. **Backend Response Times**: Multi-series line graph
4. **Circuit Breaker Status**: Table with open/closed state
5. **Failover Events**: Timeline of backend failures

### Alerting Dashboard

**Purpose**: View active and historical alerts

**Key Panels**:
1. **Active Alerts**: Table with severity, time, description
2. **Alert History**: Timeline of fired/resolved alerts
3. **Alert Frequency**: Bar chart of most common alerts
4. **MTTD/MTTR**: Single stats for detection/resolution times
5. **Silenced Alerts**: Table of currently silenced alerts

---

## Alerts

### Alert Rules Configuration

```yaml
# prometheus/alerts.yml
groups:
  - name: ollamamax
    interval: 30s
    rules:
      # High error rate
      - alert: HighErrorRate
        expr: |
          100 * (
            sum(rate(ollama_requests_total{status=~"[45].."}[5m]))
            /
            sum(rate(ollama_requests_total[5m]))
          ) > 5
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value | humanizePercentage }} (threshold: 5%)"
          runbook: "https://docs.example.com/runbooks/high-error-rate"

      # Critical error rate
      - alert: CriticalErrorRate
        expr: |
          100 * (
            sum(rate(ollama_requests_total{status=~"[45].."}[5m]))
            /
            sum(rate(ollama_requests_total[5m]))
          ) > 10
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Critical error rate detected"
          description: "Error rate is {{ $value | humanizePercentage }} (threshold: 10%)"
          runbook: "https://docs.example.com/runbooks/critical-error-rate"

      # High latency
      - alert: HighLatency
        expr: |
          histogram_quantile(0.95,
            rate(ollama_request_duration_seconds_bucket[5m])
          ) > 5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High request latency"
          description: "95th percentile latency is {{ $value }}s (threshold: 5s)"

      # Instance down
      - alert: InstanceDown
        expr: up{job="ollamamax"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Instance {{ $labels.instance }} is down"
          description: "Instance has been unreachable for 1 minute"

      # High GPU temperature
      - alert: HighGPUTemperature
        expr: ollama_gpu_temperature_celsius > 85
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High GPU temperature on {{ $labels.instance }}"
          description: "GPU {{ $labels.gpu_id }} temperature is {{ $value }}°C"

      # GPU memory exhaustion
      - alert: GPUMemoryExhausted
        expr: |
          (ollama_gpu_memory_used_bytes / ollama_gpu_memory_total_bytes) > 0.95
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "GPU memory nearly exhausted"
          description: "GPU {{ $labels.gpu_id }} using {{ $value | humanizePercentage }} of available memory"

      # Queue backup
      - alert: QueueBackup
        expr: ollama_queue_size > 100
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Request queue backing up"
          description: "Queue size is {{ $value }} requests (threshold: 100)"

      # Slow model loading
      - alert: SlowModelLoad
        expr: ollama_model_load_time_seconds > 30
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "Model loading slowly"
          description: "Model {{ $labels.model }} took {{ $value }}s to load (threshold: 30s)"

      # Disk space low
      - alert: DiskSpaceLow
        expr: |
          (node_filesystem_avail_bytes / node_filesystem_size_bytes) < 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Disk space running low"
          description: "Only {{ $value | humanizePercentage }} disk space remaining"

      # Memory leak suspected
      - alert: MemoryLeakSuspected
        expr: |
          rate(ollama_memory_usage_bytes{type="heap"}[1h]) > 1048576
        for: 2h
        labels:
          severity: warning
        annotations:
          summary: "Possible memory leak detected"
          description: "Heap memory growing at {{ $value | humanize }}B/s for 2 hours"
```

### Alert Severity Levels

| Severity | Response Time | Examples | Notification |
|----------|--------------|----------|--------------|
| **Critical** | Immediate | Instance down, data loss, >10% error rate | PagerDuty + Slack |
| **Warning** | 15 minutes | High latency, queue backup, low disk space | Slack + Email |
| **Info** | None | Scaling events, deployments | Slack only |

### Notification Templates

**Slack Template**:
```yaml
slack_configs:
  - channel: '#alerts'
    title: ':fire: {{ .GroupLabels.alertname }}'
    text: |
      *Severity:* {{ .GroupLabels.severity }}
      *Summary:* {{ .CommonAnnotations.summary }}
      *Description:* {{ .CommonAnnotations.description }}

      *Affected Instances:*
      {{ range .Alerts }}
      • {{ .Labels.instance }} ({{ .Labels.job }})
      {{ end }}

      *Runbook:* {{ .CommonAnnotations.runbook }}
    actions:
      - type: button
        text: 'View in Prometheus'
        url: '{{ .ExternalURL }}'
      - type: button
        text: 'View Dashboard'
        url: 'http://grafana:3000/d/overview'
```

---

## Tracing

### Jaeger Integration

**Purpose**: Distributed tracing for request flows across services

### Instrumentation

**Go Application**:
```go
import (
    "github.com/opentelemetry/opentelemetry-go/trace"
    "github.com/opentelemetry/opentelemetry-go/exporters/jaeger"
)

func initTracer() {
    exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(
        jaeger.WithEndpoint("http://jaeger:14268/api/traces"),
    ))
    if err != nil {
        log.Fatal(err)
    }

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(resource.NewWithAttributes(
            semconv.ServiceNameKey.String("ollamamax"),
        )),
    )
    otel.SetTracerProvider(tp)
}

func handleRequest(ctx context.Context) {
    ctx, span := tracer.Start(ctx, "handleRequest")
    defer span.End()

    // Add attributes
    span.SetAttributes(
        attribute.String("model", "llama2:7b"),
        attribute.Int("tokens", 150),
    )

    // Call downstream services with context
    processInference(ctx)
}
```

**Node.js Application**:
```javascript
const { JaegerExporter } = require('@opentelemetry/exporter-jaeger');
const { NodeTracerProvider } = require('@opentelemetry/sdk-trace-node');

const provider = new NodeTracerProvider();
const exporter = new JaegerExporter({
  endpoint: 'http://jaeger:14268/api/traces',
});

provider.addSpanProcessor(new BatchSpanProcessor(exporter));
provider.register();

// Instrument request
const tracer = provider.getTracer('ollamamax');
app.post('/api/generate', async (req, res) => {
  const span = tracer.startSpan('generate_request');
  span.setAttribute('model', req.body.model);

  try {
    const result = await processRequest(req.body);
    res.json(result);
  } finally {
    span.end();
  }
});
```

### Trace Analysis

**Common Queries**:
- Find slow traces: Filter by duration > 5s
- Error traces: Filter by tag `error=true`
- Specific endpoint: Filter by operation `/api/generate`
- Service dependencies: View service graph

### Span Context Propagation

Ensure trace context is propagated across service boundaries:

```go
// HTTP client
req, _ := http.NewRequestWithContext(ctx, "POST", url, body)
otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
```

---

## Logs

### ELK Stack Configuration

### Logstash Pipeline

```conf
# logstash/logstash.conf
input {
  beats {
    port => 5044
  }
}

filter {
  if [type] == "ollamamax" {
    json {
      source => "message"
    }

    # Parse log level
    mutate {
      uppercase => [ "level" ]
    }

    # Extract request ID
    grok {
      match => { "message" => "request_id=%{UUID:request_id}" }
    }

    # Add geolocation for IP addresses
    geoip {
      source => "client_ip"
      target => "geoip"
    }
  }
}

output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "ollamamax-%{+YYYY.MM.dd}"
  }

  # Debug output
  stdout {
    codec => rubydebug
  }
}
```

### Filebeat Configuration

```yaml
# filebeat.yml
filebeat.inputs:
  - type: log
    enabled: true
    paths:
      - /var/log/ollamamax/*.log
    fields:
      type: ollamamax
      environment: production
    multiline.pattern: '^\{'
    multiline.negate: true
    multiline.match: after

output.logstash:
  hosts: ["logstash:5044"]

logging.level: info
```

### Kibana Dashboards

**Log Dashboard**:
1. **Log Volume**: Histogram of log events over time
2. **Log Levels**: Pie chart of ERROR/WARN/INFO distribution
3. **Top Errors**: Table of most frequent error messages
4. **Request Timeline**: Timeline visualization of requests
5. **Geographic Distribution**: Map of request origins

**Useful Queries**:
```
# Errors in last hour
level:ERROR AND @timestamp:[now-1h TO now]

# Slow requests
duration:>5000 AND endpoint:"/api/generate"

# Specific user session
request_id:"550e8400-e29b-41d4-a716-446655440000"

# Failed model loads
message:"failed to load model" AND level:ERROR
```

---

## Testing

### Validation Scripts

**1. Metrics Endpoint Test**:
```bash
#!/bin/bash
# scripts/test-metrics-endpoints.sh

set -e

INSTANCE="http://localhost:8080"

echo "Testing Prometheus /metrics endpoint..."
curl -f -s "${INSTANCE}/metrics" | grep -q "ollama_requests_total"
echo "✓ /metrics endpoint working"

echo "Testing JSON /metrics.json endpoint..."
METRICS_JSON=$(curl -f -s "${INSTANCE}/metrics.json")
echo "$METRICS_JSON" | jq -e '.metrics.active_requests' > /dev/null
echo "✓ /metrics.json endpoint working"

echo "Testing content negotiation..."
TEXT_METRICS=$(curl -s -H "Accept: text/plain" "${INSTANCE}/metrics")
echo "$TEXT_METRICS" | grep -q "# TYPE"
echo "✓ Content negotiation working"

echo "All metrics tests passed!"
```

**2. Prometheus Scraping Test**:
```bash
#!/bin/bash
# scripts/test-prometheus-scraping.sh

set -e

PROMETHEUS="http://localhost:9090"

echo "Checking Prometheus targets..."
TARGETS=$(curl -s "${PROMETHEUS}/api/v1/targets" | jq -r '.data.activeTargets[].health')

if echo "$TARGETS" | grep -q "down"; then
  echo "✗ Some targets are down"
  exit 1
fi

echo "✓ All targets healthy"

echo "Testing PromQL queries..."
QUERY="up{job='ollamamax'}"
RESULT=$(curl -s -G "${PROMETHEUS}/api/v1/query" --data-urlencode "query=$QUERY" | jq -r '.data.result[0].value[1]')

if [ "$RESULT" != "1" ]; then
  echo "✗ OllamaMax instance not reporting up"
  exit 1
fi

echo "✓ PromQL queries working"
echo "All Prometheus tests passed!"
```

**3. Alert Rules Test**:
```bash
#!/bin/bash
# scripts/test-alert-rules.sh

set -e

PROMETHEUS="http://localhost:9090"

echo "Checking alert rules..."
RULES=$(curl -s "${PROMETHEUS}/api/v1/rules" | jq '.data.groups[].rules[] | select(.type=="alerting")')

RULE_COUNT=$(echo "$RULES" | jq -s 'length')
echo "Found $RULE_COUNT alert rules"

if [ "$RULE_COUNT" -lt 10 ]; then
  echo "✗ Expected at least 10 alert rules"
  exit 1
fi

echo "✓ Alert rules loaded"

echo "Checking for firing alerts..."
FIRING=$(curl -s "${PROMETHEUS}/api/v1/alerts" | jq '.data.alerts[] | select(.state=="firing")')

if [ -n "$FIRING" ]; then
  echo "⚠ Warning: Some alerts are firing:"
  echo "$FIRING" | jq -r '.labels.alertname'
fi

echo "All alert tests passed!"
```

**4. Dashboard Validation**:
```bash
#!/bin/bash
# scripts/validate-dashboards.sh

set -e

GRAFANA="http://localhost:3000"
AUTH="admin:admin"

echo "Checking Grafana health..."
curl -f -u "$AUTH" "${GRAFANA}/api/health" > /dev/null
echo "✓ Grafana healthy"

echo "Listing dashboards..."
DASHBOARDS=$(curl -s -u "$AUTH" "${GRAFANA}/api/search?type=dash-db")
DASHBOARD_COUNT=$(echo "$DASHBOARDS" | jq 'length')

echo "Found $DASHBOARD_COUNT dashboards"

if [ "$DASHBOARD_COUNT" -lt 5 ]; then
  echo "✗ Expected at least 5 dashboards"
  exit 1
fi

echo "✓ Dashboards present"
echo "All dashboard tests passed!"
```

**5. End-to-End Monitoring Test**:
```bash
#!/bin/bash
# scripts/test-monitoring-e2e.sh

set -e

echo "Running end-to-end monitoring test..."

# Generate test traffic
echo "1. Generating test requests..."
for i in {1..100}; do
  curl -s -X POST http://localhost:8080/api/generate \
    -d '{"model":"llama2:7b","prompt":"test"}' > /dev/null &
done
wait

sleep 30  # Wait for metrics collection

# Verify metrics in Prometheus
echo "2. Verifying metrics in Prometheus..."
PROMETHEUS="http://localhost:9090"
QUERY="sum(rate(ollama_requests_total[1m]))"
RATE=$(curl -s -G "${PROMETHEUS}/api/v1/query" \
  --data-urlencode "query=$QUERY" | jq -r '.data.result[0].value[1]')

if [ "$RATE" == "null" ] || [ -z "$RATE" ]; then
  echo "✗ No request rate data found"
  exit 1
fi

echo "✓ Request rate: $RATE req/s"

# Verify dashboard data
echo "3. Verifying Grafana dashboard..."
GRAFANA="http://localhost:3000"
PANEL_DATA=$(curl -s -u "admin:admin" \
  "${GRAFANA}/api/dashboards/uid/overview" | jq -r '.dashboard.panels[0]')

if [ "$PANEL_DATA" == "null" ]; then
  echo "✗ Dashboard panels not found"
  exit 1
fi

echo "✓ Dashboard rendering"

echo "End-to-end test passed!"
```

### Running Tests

```bash
# Run all validation tests
./scripts/test-metrics-endpoints.sh
./scripts/test-prometheus-scraping.sh
./scripts/test-alert-rules.sh
./scripts/validate-dashboards.sh
./scripts/test-monitoring-e2e.sh

# Or run comprehensive test suite
npm run test:monitoring
```

---

## Troubleshooting

### Common Issues

#### Issue: Prometheus Not Scraping Targets

**Symptoms**:
- Targets show "down" in Prometheus UI
- No metrics data in Grafana

**Diagnosis**:
```bash
# Check target status
curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | select(.health=="down")'

# Test endpoint manually
curl http://localhost:8080/metrics

# Check Prometheus logs
docker logs prometheus
```

**Solutions**:
1. Verify network connectivity:
   ```bash
   docker exec prometheus ping host.docker.internal
   ```

2. Check firewall rules:
   ```bash
   sudo ufw status
   sudo ufw allow 8080/tcp
   ```

3. Verify service discovery:
   ```yaml
   # prometheus.yml
   scrape_configs:
     - job_name: 'ollamamax'
       static_configs:
         - targets: ['host.docker.internal:8080']  # Use correct host
   ```

4. Reload Prometheus config:
   ```bash
   curl -X POST http://localhost:9090/-/reload
   ```

---

#### Issue: High Cardinality Metrics

**Symptoms**:
- Prometheus using excessive memory
- Slow query performance
- OOM errors

**Diagnosis**:
```bash
# Check series count
curl http://localhost:9090/api/v1/status/tsdb | jq '.data.seriesCountByMetricName'

# Identify high-cardinality metrics
curl http://localhost:9090/api/v1/label/__name__/values | jq -r '.data[]' | \
  xargs -I {} sh -c 'echo -n "{}: "; curl -s "http://localhost:9090/api/v1/series?match[]={}" | jq ".data | length"'
```

**Solutions**:
1. Add metric relabeling to drop high-cardinality labels:
   ```yaml
   scrape_configs:
     - job_name: 'ollamamax'
       metric_relabel_configs:
         - source_labels: [user_id]
           action: labeldrop
   ```

2. Use recording rules for pre-aggregation:
   ```yaml
   groups:
     - name: aggregated
       rules:
         - record: job:ollama_requests_total:rate5m
           expr: sum(rate(ollama_requests_total[5m])) by (job)
   ```

3. Adjust retention:
   ```bash
   # Reduce retention period
   --storage.tsdb.retention.time=7d
   ```

---

#### Issue: Missing Metrics in Grafana

**Symptoms**:
- Dashboard panels show "No data"
- Queries return empty results

**Diagnosis**:
```bash
# Test Prometheus datasource
curl -u admin:admin http://localhost:3000/api/datasources

# Test query directly in Prometheus
curl -G http://localhost:9090/api/v1/query \
  --data-urlencode 'query=up{job="ollamamax"}'

# Check Grafana logs
docker logs grafana
```

**Solutions**:
1. Verify datasource configuration:
   - Go to Configuration > Data Sources
   - Test connection to Prometheus
   - Check URL is correct (http://prometheus:9090)

2. Check time range:
   - Ensure dashboard time range includes recent data
   - Try "Last 5 minutes" to verify live data

3. Verify query syntax:
   ```promql
   # Wrong
   ollama_requests_total

   # Right (must specify time range)
   rate(ollama_requests_total[5m])
   ```

---

#### Issue: Alerts Not Firing

**Symptoms**:
- Expected alerts not triggering
- No notifications received

**Diagnosis**:
```bash
# Check alert rules
curl http://localhost:9090/api/v1/rules | jq '.data.groups[].rules[] | select(.type=="alerting")'

# Check alert state
curl http://localhost:9090/api/v1/alerts | jq '.data.alerts[] | select(.state!="inactive")'

# Check Alertmanager
curl http://localhost:9093/api/v2/alerts
```

**Solutions**:
1. Verify rule syntax:
   ```yaml
   # Check for syntax errors
   promtool check rules alerts.yml
   ```

2. Adjust thresholds and duration:
   ```yaml
   - alert: HighErrorRate
     expr: error_rate > 5  # Lower threshold for testing
     for: 30s  # Shorter duration for testing
   ```

3. Check Alertmanager routing:
   ```yaml
   route:
     receiver: 'default'  # Ensure receiver exists
   ```

4. Test notification channel:
   ```bash
   # Send test alert
   curl -XPOST http://localhost:9093/api/v2/alerts -d '[
     {
       "labels": {"alertname": "Test", "severity": "warning"},
       "annotations": {"summary": "Test alert"}
     }
   ]'
   ```

---

#### Issue: Jaeger Traces Not Appearing

**Symptoms**:
- No traces in Jaeger UI
- Spans not being collected

**Diagnosis**:
```bash
# Check Jaeger collector health
curl http://localhost:14269/

# Check Jaeger logs
docker logs jaeger

# Verify application instrumentation
# Look for trace initialization in app logs
```

**Solutions**:
1. Verify exporter configuration:
   ```go
   exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(
       jaeger.WithEndpoint("http://jaeger:14268/api/traces"),
   ))
   ```

2. Check sampling rate:
   ```go
   // Increase sampling for testing
   trace.WithSampler(trace.AlwaysSample())
   ```

3. Test collector endpoint:
   ```bash
   curl -v http://localhost:14268/api/traces
   ```

---

#### Issue: Elasticsearch Disk Space Full

**Symptoms**:
- Logs not being stored
- Kibana showing read-only indices

**Diagnosis**:
```bash
# Check cluster health
curl http://localhost:9200/_cluster/health

# Check disk space
curl http://localhost:9200/_cat/allocation?v

# Check index sizes
curl http://localhost:9200/_cat/indices?v&s=store.size:desc
```

**Solutions**:
1. Delete old indices:
   ```bash
   # Delete indices older than 7 days
   curator_cli --host localhost --port 9200 delete_indices \
     --filter_list '[{"filtertype":"age","source":"name","direction":"older","timestring":"%Y.%m.%d","unit":"days","unit_count":7}]'
   ```

2. Adjust ILM policy:
   ```bash
   # Set shorter retention
   curl -X PUT "localhost:9200/_ilm/policy/ollamamax-policy" -H 'Content-Type: application/json' -d'
   {
     "policy": {
       "phases": {
         "delete": {
           "min_age": "3d",
           "actions": {"delete": {}}
         }
       }
     }
   }'
   ```

3. Increase disk space:
   ```bash
   # Expand volume
   docker volume inspect elasticsearch_data
   ```

---

## Performance

### Monitoring Overhead

Expected performance impact:

| Component | CPU Overhead | Memory Overhead | Network Overhead |
|-----------|--------------|-----------------|-------------------|
| Metrics collection | <0.5% | ~10MB per instance | ~1KB/s |
| Distributed tracing (1% sampling) | <0.1% | ~5MB | ~2KB/s |
| Log shipping | <0.2% | ~5MB | ~5KB/s |
| **Total** | **<1%** | **~20MB** | **~8KB/s** |

### Optimization Tips

**1. Adjust Scrape Intervals**:
```yaml
# Increase interval for less critical metrics
scrape_configs:
  - job_name: 'ollamamax-detailed'
    scrape_interval: 15s  # Detailed metrics

  - job_name: 'ollamamax-summary'
    scrape_interval: 60s  # Summary metrics
```

**2. Use Recording Rules**:
```yaml
# Pre-compute expensive queries
groups:
  - name: precomputed
    interval: 30s
    rules:
      - record: instance:request_rate:rate5m
        expr: rate(ollama_requests_total[5m])

      - record: instance:error_rate:rate5m
        expr: |
          rate(ollama_requests_total{status=~"[45].."}[5m])
          /
          rate(ollama_requests_total[5m])
```

**3. Enable Metric Caching**:
```go
// Cache metrics for 5 seconds
cache := prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "ollama_cached_metric",
        Help: "Cached metric to reduce computation",
    },
    []string{"label"},
)

// Update cache periodically
go func() {
    ticker := time.NewTicker(5 * time.Second)
    for range ticker.C {
        updateCachedMetrics()
    }
}()
```

**4. Sampling for Traces**:
```go
// Use adaptive sampling
trace.WithSampler(trace.ParentBased(
    trace.TraceIDRatioBased(0.01),  // 1% sampling
))

// Sample more aggressively for errors
if isError {
    trace.WithSampler(trace.AlwaysSample())
}
```

**5. Batch Log Shipping**:
```yaml
# filebeat.yml
filebeat.inputs:
  - type: log
    enabled: true
    paths:
      - /var/log/ollamamax/*.log

output.logstash:
  hosts: ["logstash:5044"]
  bulk_max_size: 2048  # Batch size
  compression_level: 3  # Compression
  worker: 2  # Parallel workers
```

### Resource Sizing

**Small Deployment** (1-10 instances):
- Prometheus: 2GB RAM, 10GB disk
- Grafana: 512MB RAM, 5GB disk
- Alertmanager: 256MB RAM, 1GB disk
- Jaeger: 1GB RAM, 10GB disk
- Elasticsearch: 4GB RAM, 50GB disk

**Medium Deployment** (10-50 instances):
- Prometheus: 8GB RAM, 100GB disk
- Grafana: 2GB RAM, 10GB disk
- Alertmanager: 1GB RAM, 5GB disk
- Jaeger: 4GB RAM, 100GB disk
- Elasticsearch: 16GB RAM, 500GB disk

**Large Deployment** (50+ instances):
- Prometheus: 32GB RAM, 500GB disk (or federated setup)
- Grafana: 4GB RAM, 20GB disk
- Alertmanager: 2GB RAM, 10GB disk
- Jaeger: 16GB RAM, 1TB disk
- Elasticsearch: 64GB RAM, 2TB disk (multi-node)

---

## Security

### Authentication

**Prometheus Basic Auth**:
```yaml
# prometheus.yml
basic_auth_users:
  admin: $2y$10$hashed_password

# Or use reverse proxy
# nginx.conf
location /prometheus {
    auth_basic "Restricted";
    auth_basic_user_file /etc/nginx/.htpasswd;
    proxy_pass http://prometheus:9090;
}
```

**Grafana OAuth2**:
```bash
# Environment variables
GF_AUTH_GENERIC_OAUTH_ENABLED=true
GF_AUTH_GENERIC_OAUTH_CLIENT_ID=your-client-id
GF_AUTH_GENERIC_OAUTH_CLIENT_SECRET=your-client-secret
GF_AUTH_GENERIC_OAUTH_AUTH_URL=https://accounts.google.com/o/oauth2/auth
GF_AUTH_GENERIC_OAUTH_TOKEN_URL=https://accounts.google.com/o/oauth2/token
```

### TLS/HTTPS

**Generate Certificates**:
```bash
# Self-signed certificate (development)
openssl req -x509 -newkey rsa:4096 \
  -keyout key.pem -out cert.pem \
  -days 365 -nodes \
  -subj "/CN=monitoring.example.com"

# Let's Encrypt (production)
certbot certonly --standalone -d monitoring.example.com
```

**Configure TLS in Prometheus**:
```yaml
# prometheus.yml
tls_server_config:
  cert_file: /etc/prometheus/cert.pem
  key_file: /etc/prometheus/key.pem
```

**Configure TLS in Grafana**:
```bash
# Environment variables
GF_SERVER_PROTOCOL=https
GF_SERVER_CERT_FILE=/etc/grafana/cert.pem
GF_SERVER_CERT_KEY=/etc/grafana/key.pem
```

### Network Security

**Firewall Rules**:
```bash
# Allow only from specific networks
sudo ufw allow from 10.0.0.0/8 to any port 9090 proto tcp  # Prometheus
sudo ufw allow from 10.0.0.0/8 to any port 3000 proto tcp  # Grafana

# Deny public access
sudo ufw deny 9090/tcp
sudo ufw deny 3000/tcp
```

**Kubernetes Network Policies**:
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: monitoring-network-policy
  namespace: monitoring
spec:
  podSelector:
    matchLabels:
      app: prometheus
  policyTypes:
    - Ingress
  ingress:
    - from:
      - namespaceSelector:
          matchLabels:
            name: monitoring
      ports:
        - protocol: TCP
          port: 9090
```

### Secrets Management

**Kubernetes Secrets**:
```bash
# Create secret
kubectl create secret generic monitoring-tls \
  --from-file=cert.pem \
  --from-file=key.pem \
  -n monitoring

# Mount in pod
volumes:
  - name: tls
    secret:
      secretName: monitoring-tls
volumeMounts:
  - name: tls
    mountPath: /etc/tls
    readOnly: true
```

**Vault Integration**:
```yaml
# Grafana with Vault
annotations:
  vault.hashicorp.com/agent-inject: "true"
  vault.hashicorp.com/role: "grafana"
  vault.hashicorp.com/agent-inject-secret-admin: "secret/grafana/admin"
  vault.hashicorp.com/agent-inject-template-admin: |
    {{ with secret "secret/grafana/admin" -}}
    export GF_SECURITY_ADMIN_PASSWORD="{{ .Data.password }}"
    {{- end }}
```

### Audit Logging

Enable audit logs for security compliance:

```yaml
# grafana.ini
[log]
mode = console file
level = info

[log.console]
format = json

[auditing]
enabled = true
log_dashboard_content = true
```

---

## Best Practices

### 1. Metric Naming Conventions

Follow Prometheus naming standards:
- Use `snake_case` for metric names
- Include units in metric names (`_seconds`, `_bytes`, `_total`)
- Counter metrics end with `_total`
- Use consistent label names across metrics

**Good Examples**:
```
ollama_request_duration_seconds
ollama_requests_total
ollama_memory_usage_bytes
ollama_gpu_utilization_percent
```

**Bad Examples**:
```
requestDuration  # camelCase
requests  # missing _total
memory  # missing unit
GPU-Util  # inconsistent case
```

### 2. Label Cardinality

Keep label cardinality low (<100 unique values per label):

**Good**:
```promql
ollama_requests_total{endpoint="/api/generate", status="200"}
```

**Bad**:
```promql
ollama_requests_total{user_id="12345", request_id="uuid", timestamp="..."}
# Creates millions of series!
```

### 3. Dashboard Organization

- Create role-specific dashboards (ops, dev, business)
- Use consistent time ranges and refresh rates
- Add descriptions and documentation links
- Use template variables for flexibility
- Group related panels together

### 4. Alert Hygiene

- Set appropriate thresholds based on historical data
- Use `for` clause to avoid flapping
- Include runbook links in annotations
- Test alerts in staging before production
- Regularly review and tune alert rules

### 5. Retention Policies

Balance storage costs with data needs:

**Prometheus**:
- Short-term: 15 days at full resolution
- Long-term: Use remote storage (Thanos, Cortex)

**Logs**:
- Hot: 7 days (searchable)
- Warm: 30 days (archived)
- Cold: 90 days (compliance)

### 6. Monitoring Monitoring

Monitor the monitoring stack itself:

```promql
# Prometheus scrape failures
up{job="prometheus"} == 0

# High query latency
prometheus_http_request_duration_seconds{handler="/api/v1/query", quantile="0.99"} > 1

# Disk space
prometheus_tsdb_storage_blocks_bytes / prometheus_tsdb_retention_limit_bytes > 0.9
```

### 7. Documentation

Maintain up-to-date documentation:
- Runbooks for each alert
- Dashboard usage guides
- Metrics catalog with examples
- Architecture diagrams
- Change log for monitoring changes

### 8. Testing and Validation

- Test metric collection in staging
- Validate alert rules with unit tests
- Perform chaos engineering exercises
- Conduct regular disaster recovery drills

### 9. Continuous Improvement

- Review dashboards quarterly
- Analyze alert fatigue metrics
- Gather feedback from on-call engineers
- Update based on incident post-mortems
- Stay current with tool updates

### 10. Cost Optimization

- Use recording rules to pre-aggregate expensive queries
- Implement metric dropping for unused data
- Tune retention based on actual needs
- Use sampling for high-volume traces
- Archive cold data to object storage

---

## Appendix

### Quick Reference Commands

```bash
# Prometheus
curl http://localhost:9090/api/v1/targets
curl http://localhost:9090/api/v1/alerts
curl -X POST http://localhost:9090/-/reload

# Grafana
curl -u admin:admin http://localhost:3000/api/health
curl -u admin:admin http://localhost:3000/api/dashboards/home

# Alertmanager
curl http://localhost:9093/api/v2/alerts
curl http://localhost:9093/api/v2/silences

# Jaeger
curl http://localhost:16686/api/services
curl http://localhost:16686/api/traces/{trace-id}

# Elasticsearch
curl http://localhost:9200/_cluster/health
curl http://localhost:9200/_cat/indices?v
```

### Useful PromQL Snippets

```promql
# Rate of change
rate(metric[5m])

# Increase over time
increase(metric[1h])

# Aggregations
sum(metric) by (label)
avg(metric) without (label)
topk(10, metric)
bottomk(5, metric)

# Joins
metric1 * on(label) metric2

# Time shifting
metric offset 1h

# Subqueries
rate(metric[5m])[1h:1m]
```

### Support Resources

- **Documentation**: https://prometheus.io/docs/
- **Grafana Dashboards**: https://grafana.com/grafana/dashboards/
- **Community**: Prometheus Slack, StackOverflow
- **Training**: Prometheus certification courses

---

**Document Version**: 1.0.0
**Last Updated**: 2025-10-27
**Authors**: OllamaMax Platform Team
**Review Cycle**: Quarterly
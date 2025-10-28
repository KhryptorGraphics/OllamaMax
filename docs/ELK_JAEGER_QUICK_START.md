# ELK Stack and Jaeger Quick Start Guide

## Deployment

### 1. Deploy the Stack
```bash
kubectl apply -f k8s/monitoring-stack.yaml
```

### 2. Verify Deployment
```bash
# Run validation script
./scripts/validate-elk-jaeger-deployment.sh

# Or manually check pods
kubectl get pods -n ollamamax-monitoring -w
```

### 3. Wait for All Pods to be Ready
Expected pods:
- `elasticsearch-0`, `elasticsearch-1`, `elasticsearch-2` (StatefulSet)
- `logstash-*` (Deployment)
- `kibana-*` (Deployment)
- `filebeat-*` (DaemonSet - one per node)
- `jaeger-*` (Deployment)

---

## Accessing the UIs

### Kibana (Log Visualization)

**Option 1: LoadBalancer (if available)**
```bash
# Get external IP
kubectl get svc -n ollamamax-monitoring kibana

# Access at: http://<EXTERNAL-IP>:5601
```

**Option 2: Port Forward**
```bash
kubectl port-forward -n ollamamax-monitoring svc/kibana 5601:5601

# Access at: http://localhost:5601
```

### Jaeger (Distributed Tracing)

**Option 1: LoadBalancer (if available)**
```bash
# Get external IP
kubectl get svc -n ollamamax-monitoring jaeger-ui

# Access at: http://<EXTERNAL-IP>:16686
```

**Option 2: Port Forward**
```bash
kubectl port-forward -n ollamamax-monitoring svc/jaeger-ui 16686:16686

# Access at: http://localhost:16686
```

---

## Initial Configuration

### Kibana Setup

1. **Access Kibana UI** (http://localhost:5601)

2. **Create Index Pattern:**
   - Navigate to: Management → Stack Management → Index Patterns
   - Click "Create index pattern"
   - Index pattern name: `ollamamax-logs-*`
   - Time field: `@timestamp`
   - Click "Create index pattern"

3. **View Logs:**
   - Navigate to: Analytics → Discover
   - Select `ollamamax-logs-*` index pattern
   - View and search logs

4. **Create Dashboards (Optional):**
   - Navigate to: Analytics → Dashboard
   - Create visualizations for:
     - Log volume over time
     - Top pods by log count
     - Error rate by component
     - Response time percentiles

---

## Using Elasticsearch

### Direct API Access

**Port Forward:**
```bash
kubectl port-forward -n ollamamax-monitoring svc/elasticsearch-http 9200:9200
```

**Check Cluster Health:**
```bash
curl http://localhost:9200/_cluster/health?pretty
```

**List Indices:**
```bash
curl http://localhost:9200/_cat/indices?v
```

**Search Logs:**
```bash
# Search last 100 logs
curl -X GET "http://localhost:9200/ollamamax-logs-*/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "match_all": {}
  },
  "size": 100,
  "sort": [
    {
      "@timestamp": {
        "order": "desc"
      }
    }
  ]
}
'
```

**Search for Errors:**
```bash
curl -X GET "http://localhost:9200/ollamamax-logs-*/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "match": {
      "level": "error"
    }
  },
  "size": 50
}
'
```

**Filter by Pod:**
```bash
curl -X GET "http://localhost:9200/ollamamax-logs-*/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "term": {
      "k8s_pod.keyword": "your-pod-name"
    }
  }
}
'
```

---

## Using Jaeger

### Send Traces from Applications

**Go Example (OpenTelemetry):**
```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func initTracer() (*sdktrace.TracerProvider, error) {
    // Create Jaeger exporter
    exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(
        jaeger.WithEndpoint("http://jaeger.ollamamax-monitoring.svc.cluster.local:14268/api/traces"),
    ))
    if err != nil {
        return nil, err
    }

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String("ollamamax-api"),
        )),
    )

    otel.SetTracerProvider(tp)
    return tp, nil
}
```

**Node.js Example (OpenTelemetry):**
```javascript
const { JaegerExporter } = require('@opentelemetry/exporter-jaeger');
const { NodeTracerProvider } = require('@opentelemetry/sdk-trace-node');
const { registerInstrumentations } = require('@opentelemetry/instrumentation');
const { HttpInstrumentation } = require('@opentelemetry/instrumentation-http');

const provider = new NodeTracerProvider();

const jaegerExporter = new JaegerExporter({
  endpoint: 'http://jaeger.ollamamax-monitoring.svc.cluster.local:14268/api/traces',
});

provider.addSpanProcessor(new BatchSpanProcessor(jaegerExporter));
provider.register();

registerInstrumentations({
  instrumentations: [
    new HttpInstrumentation(),
  ],
});
```

### View Traces

1. Access Jaeger UI (http://localhost:16686)
2. Select service from dropdown
3. Click "Find Traces"
4. Click on a trace to view details
5. Analyze spans, timing, and dependencies

---

## Log Format for Best Results

### Structured JSON Logging

**Go Example:**
```go
import "github.com/sirupsen/logrus"

log := logrus.New()
log.SetFormatter(&logrus.JSONFormatter{})
log.WithFields(logrus.Fields{
    "component": "api",
    "method": "GET",
    "path": "/models",
    "status": 200,
    "duration_ms": 42,
}).Info("Request processed")
```

**Expected Output:**
```json
{
  "component": "api",
  "duration_ms": 42,
  "level": "info",
  "method": "GET",
  "msg": "Request processed",
  "path": "/models",
  "status": 200,
  "time": "2025-10-27T10:30:45Z"
}
```

### Log Fields for Correlation

**Recommended Fields:**
- `trace_id`: Jaeger trace ID
- `span_id`: Jaeger span ID
- `request_id`: Unique request identifier
- `user_id`: User identifier
- `component`: Service/component name
- `operation`: Operation being performed
- `status`: HTTP status or operation result
- `duration_ms`: Operation duration

**Example with Trace Correlation:**
```json
{
  "level": "info",
  "time": "2025-10-27T10:30:45Z",
  "component": "ollamamax-api",
  "operation": "model_inference",
  "trace_id": "abc123def456",
  "span_id": "789ghi012",
  "request_id": "req-xyz-789",
  "model": "llama2-7b",
  "status": "success",
  "duration_ms": 234,
  "msg": "Model inference completed"
}
```

---

## Common Queries

### Kibana Query Language (KQL)

**Search for errors:**
```
level: error
```

**Search specific component:**
```
component: "ollamamax-api"
```

**Search by time range and status:**
```
@timestamp > "now-1h" AND status: 500
```

**Search for slow requests:**
```
duration_ms > 1000
```

**Combined query:**
```
component: "ollamamax-api" AND (level: error OR level: warning) AND @timestamp > "now-6h"
```

### Elasticsearch Query DSL

**Aggregation - Error Count by Component:**
```bash
curl -X GET "http://localhost:9200/ollamamax-logs-*/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "size": 0,
  "query": {
    "term": {
      "level": "error"
    }
  },
  "aggs": {
    "by_component": {
      "terms": {
        "field": "component.keyword",
        "size": 10
      }
    }
  }
}
'
```

**Aggregation - Average Response Time:**
```bash
curl -X GET "http://localhost:9200/ollamamax-logs-*/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "size": 0,
  "aggs": {
    "avg_duration": {
      "avg": {
        "field": "duration_ms"
      }
    }
  }
}
'
```

---

## Troubleshooting

### Elasticsearch Cluster Issues

**Check cluster health:**
```bash
kubectl exec -n ollamamax-monitoring elasticsearch-0 -- curl -s http://localhost:9200/_cluster/health?pretty
```

**Check node status:**
```bash
kubectl exec -n ollamamax-monitoring elasticsearch-0 -- curl -s http://localhost:9200/_cat/nodes?v
```

**Check shard allocation:**
```bash
kubectl exec -n ollamamax-monitoring elasticsearch-0 -- curl -s http://localhost:9200/_cat/shards?v
```

**Restart Elasticsearch pod:**
```bash
kubectl delete pod -n ollamamax-monitoring elasticsearch-0
# Wait for pod to restart
```

### Logstash Issues

**Check Logstash logs:**
```bash
kubectl logs -n ollamamax-monitoring -l app=logstash --tail=100
```

**Test Logstash pipeline:**
```bash
kubectl exec -n ollamamax-monitoring <logstash-pod> -- curl -s http://localhost:9600
```

**Restart Logstash:**
```bash
kubectl rollout restart deployment/logstash -n ollamamax-monitoring
```

### Filebeat Issues

**Check Filebeat logs on specific node:**
```bash
kubectl logs -n ollamamax-monitoring <filebeat-pod> --tail=100
```

**Test Logstash connectivity from Filebeat:**
```bash
kubectl exec -n ollamamax-monitoring <filebeat-pod> -- nc -zv logstash 5044
```

**Restart Filebeat on a node:**
```bash
kubectl delete pod -n ollamamax-monitoring <filebeat-pod>
```

### Kibana Issues

**Check Kibana logs:**
```bash
kubectl logs -n ollamamax-monitoring -l app=kibana --tail=100
```

**Check Elasticsearch connectivity:**
```bash
kubectl exec -n ollamamax-monitoring <kibana-pod> -- curl -s http://elasticsearch-http:9200
```

**Restart Kibana:**
```bash
kubectl rollout restart deployment/kibana -n ollamamax-monitoring
```

### Jaeger Issues

**Check Jaeger logs:**
```bash
kubectl logs -n ollamamax-monitoring -l app=jaeger --tail=100
```

**Test Jaeger collector:**
```bash
kubectl exec -n ollamamax-monitoring <jaeger-pod> -- curl -s http://localhost:14269
```

**Restart Jaeger:**
```bash
kubectl rollout restart deployment/jaeger -n ollamamax-monitoring
```

---

## Maintenance

### Index Management

**Delete old indices:**
```bash
# List indices by date
kubectl exec -n ollamamax-monitoring elasticsearch-0 -- curl -s "http://localhost:9200/_cat/indices/ollamamax-logs-*?v&s=index"

# Delete specific index
kubectl exec -n ollamamax-monitoring elasticsearch-0 -- curl -X DELETE "http://localhost:9200/ollamamax-logs-2025.10.01"

# Delete indices older than 30 days (requires curator or ILM)
```

**Configure Index Lifecycle Management (ILM):**
```bash
# Create ILM policy
curl -X PUT "http://localhost:9200/_ilm/policy/ollamamax-logs-policy" -H 'Content-Type: application/json' -d'
{
  "policy": {
    "phases": {
      "hot": {
        "actions": {
          "rollover": {
            "max_age": "1d",
            "max_size": "50gb"
          }
        }
      },
      "delete": {
        "min_age": "30d",
        "actions": {
          "delete": {}
        }
      }
    }
  }
}
'
```

### Backup and Restore

**Create snapshot repository:**
```bash
curl -X PUT "http://localhost:9200/_snapshot/backup_repo" -H 'Content-Type: application/json' -d'
{
  "type": "fs",
  "settings": {
    "location": "/backups"
  }
}
'
```

**Create snapshot:**
```bash
curl -X PUT "http://localhost:9200/_snapshot/backup_repo/snapshot_1"
```

**Restore snapshot:**
```bash
curl -X POST "http://localhost:9200/_snapshot/backup_repo/snapshot_1/_restore"
```

---

## Performance Tuning

### Elasticsearch

**Increase heap size (edit StatefulSet):**
```yaml
env:
- name: ES_JAVA_OPTS
  value: "-Xms2g -Xmx2g"  # Increase from 512m
```

**Optimize for write performance:**
```bash
curl -X PUT "http://localhost:9200/ollamamax-logs-*/_settings" -H 'Content-Type: application/json' -d'
{
  "index": {
    "refresh_interval": "30s",
    "number_of_replicas": 0
  }
}
'
```

### Logstash

**Increase workers:**
```yaml
env:
- name: LS_JAVA_OPTS
  value: "-Xmx1g -Xms1g"  # Increase heap
```

**Update pipeline config:**
```ruby
output {
  elasticsearch {
    hosts => ["http://elasticsearch-http:9200"]
    workers => 4  # Increase workers
    bulk_size => 5000
  }
}
```

### Filebeat

**Increase bulk size (edit ConfigMap):**
```yaml
output.logstash:
  hosts: ["logstash:5044"]
  bulk_max_size: 4096  # Increase from 2048
  worker: 4  # Increase workers
```

---

## Security Hardening (Production)

### Enable Elasticsearch Security

1. **Update ConfigMap to enable security:**
```yaml
- name: xpack.security.enabled
  value: "true"
- name: xpack.security.http.ssl.enabled
  value: "true"
```

2. **Set up passwords:**
```bash
kubectl exec -n ollamamax-monitoring elasticsearch-0 -- bin/elasticsearch-setup-passwords auto
```

3. **Update Kibana/Logstash/Filebeat with credentials**

### Network Policies

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: elk-stack-network-policy
  namespace: ollamamax-monitoring
spec:
  podSelector:
    matchLabels:
      app: elasticsearch
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: kibana
    - podSelector:
        matchLabels:
          app: logstash
    ports:
    - protocol: TCP
      port: 9200
```

---

## Monitoring the Monitoring Stack

### Prometheus Metrics

All components expose metrics that Prometheus can scrape:

- Elasticsearch: Port 9200 (metrics via plugin)
- Logstash: Port 9600
- Filebeat: Port 5066 (if enabled)
- Jaeger: Port 14269

**Add to Prometheus config:**
```yaml
scrape_configs:
  - job_name: 'elasticsearch'
    static_configs:
      - targets: ['elasticsearch-http:9200']

  - job_name: 'logstash'
    static_configs:
      - targets: ['logstash:9600']

  - job_name: 'jaeger'
    static_configs:
      - targets: ['jaeger:14269']
```

---

## Summary

**ELK Stack Flow:**
1. Filebeat collects logs from all pods
2. Logstash processes and enriches logs
3. Elasticsearch stores and indexes logs
4. Kibana visualizes logs

**Jaeger Flow:**
1. Application sends traces to Jaeger collector
2. Jaeger stores traces in memory
3. Jaeger UI displays traces

**Key URLs (with port-forward):**
- Kibana: http://localhost:5601
- Jaeger: http://localhost:16686
- Elasticsearch: http://localhost:9200

**Validation:**
```bash
./scripts/validate-elk-jaeger-deployment.sh
```

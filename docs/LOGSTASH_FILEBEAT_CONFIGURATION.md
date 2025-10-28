# Logstash and Filebeat Configuration

## Overview

Production-ready ELK stack configuration for comprehensive log aggregation, processing, and analysis with distributed tracing correlation.

## Architecture

```
Docker Containers -> Filebeat -> Logstash -> Elasticsearch -> Kibana
                                      ↓
                              Trace Correlation
                                      ↓
                                   Jaeger
```

## Components

### 1. Filebeat (Log Collection)

**Location**: `/monitoring/filebeat/filebeat.yml`

**Features**:
- Container log collection from Docker
- Docker metadata enrichment
- Kubernetes autodiscovery support
- JSON log parsing
- Load-balanced output to Logstash
- Service and environment tagging

**Configuration**:
```yaml
# Container logs input
- type: container
  paths:
    - '/var/lib/docker/containers/*/*.log'

# Output to Logstash
output.logstash:
  hosts: ["logstash:5044"]
  compression_level: 3
  worker: 2
  loadbalance: true
```

### 2. Logstash (Log Processing)

**Location**: `/monitoring/logstash/pipeline/logstash.conf`

**Features**:
- JSON message parsing
- Timestamp extraction and normalization
- Log level extraction and tagging
- Trace/span ID correlation for Jaeger
- Error and audit log tagging
- Kubernetes metadata enrichment
- Index routing (logs, errors, audit)

**Pipeline Flow**:
```
Input (Beats:5044) -> Filter (Parse/Enrich) -> Output (Elasticsearch)
```

### 3. Index Strategy

**Three index types**:

1. **ollamamax-logs-YYYY.MM.dd**
   - Standard application logs
   - INFO, DEBUG, WARN levels
   - General operations

2. **ollamamax-errors-YYYY.MM.dd**
   - ERROR and FATAL logs
   - Tagged with `error`
   - Priority alerting

3. **ollamamax-audit-YYYY.MM.dd**
   - Audit trail events
   - Security-sensitive operations
   - Compliance tracking
   - Tagged with `audit`

## Trace Correlation

### How It Works

Logstash extracts `trace_id` and `span_id` from logs and preserves them in Elasticsearch:

```
Application Log -> Filebeat -> Logstash -> Elasticsearch
    ↓
trace_id: abc123
span_id: def456
    ↓
Query Jaeger with trace_id -> Full trace visualization
```

### Log Format Requirements

To enable trace correlation, applications must log in JSON format:

```json
{
  "timestamp": "2025-01-27T10:30:45.123Z",
  "level": "info",
  "message": "Processing request",
  "trace_id": "abc123def456789",
  "span_id": "def456789",
  "service": "ollamamax-api"
}
```

## Field Mapping

### Extracted Fields

| Field | Source | Purpose |
|-------|--------|---------|
| `@timestamp` | `log.timestamp` | Event time |
| `level` | `log.level` | Log severity |
| `trace_id` | `log.trace_id` | Distributed trace ID |
| `span_id` | `log.span_id` | Trace span ID |
| `k8s_namespace` | `kubernetes.namespace` | K8s namespace |
| `k8s_pod` | `kubernetes.pod.name` | Pod name |
| `k8s_container` | `kubernetes.container.name` | Container name |

### Tags

- `error`: ERROR/FATAL logs
- `audit`: Audit events
- `forwarded`: From Filebeat

## Usage Examples

### Query Logs by Trace ID (Kibana)

```
GET ollamamax-logs-*/_search
{
  "query": {
    "term": {
      "trace_id": "abc123def456789"
    }
  }
}
```

### Query All Errors

```
GET ollamamax-errors-*/_search
{
  "query": {
    "match_all": {}
  },
  "sort": [
    { "@timestamp": "desc" }
  ]
}
```

### Query Audit Logs

```
GET ollamamax-audit-*/_search
{
  "query": {
    "range": {
      "@timestamp": {
        "gte": "now-24h"
      }
    }
  }
}
```

## Docker Compose Integration

### Services

1. **Filebeat**:
   - Port: None (internal)
   - Collects from: `/var/lib/docker/containers`
   - Sends to: Logstash:5044

2. **Logstash**:
   - Port: 5044 (Beats input)
   - Port: 9600 (Monitoring API)
   - Reads from: Filebeat
   - Writes to: Elasticsearch

3. **Elasticsearch**:
   - Port: 9200 (HTTP API)
   - Port: 9300 (Transport)
   - Storage: `elasticsearch_data` volume

4. **Kibana**:
   - Port: 5601 (Web UI)
   - Connects to: Elasticsearch

## Configuration Files

### Environment Variables

**Filebeat**:
```env
LOGSTASH_HOST=logstash
LOGSTASH_PORT=5044
ENVIRONMENT=production
```

**Logstash**:
```env
ELASTICSEARCH_HOST=elasticsearch
ELASTICSEARCH_PORT=9200
LS_JAVA_OPTS=-Xms256m -Xmx256m
```

### Volume Mounts

**Filebeat**:
- Config: `./monitoring/filebeat/filebeat.yml:/usr/share/filebeat/filebeat.yml:ro`
- Logs: `/var/lib/docker/containers:/var/lib/docker/containers:ro`
- Socket: `/var/run/docker.sock:/var/run/docker.sock:ro`

**Logstash**:
- Pipeline: `./monitoring/logstash/pipeline:/usr/share/logstash/pipeline:ro`

## Monitoring

### Health Checks

**Logstash**:
```bash
curl http://localhost:9600/_node/stats
```

**Elasticsearch**:
```bash
curl http://localhost:9200/_cluster/health
```

**Kibana**:
```bash
curl http://localhost:5601/api/status
```

### Metrics

Monitor via Prometheus:
- Filebeat: Logs shipped/sec
- Logstash: Events processed/sec
- Elasticsearch: Index size, query latency

## Troubleshooting

### Filebeat Not Collecting Logs

```bash
# Check Filebeat container
docker logs filebeat

# Verify permissions
docker exec filebeat ls -la /var/lib/docker/containers

# Test Logstash connection
docker exec filebeat ping logstash
```

### Logstash Pipeline Errors

```bash
# Check pipeline syntax
docker exec logstash /usr/share/logstash/bin/logstash --config.test_and_exit -f /usr/share/logstash/pipeline/logstash.conf

# View logs
docker logs logstash

# Check Elasticsearch connection
docker exec logstash curl -X GET "elasticsearch:9200/_cluster/health?pretty"
```

### Missing Trace IDs

1. Verify application logs contain `trace_id` field
2. Check Logstash filter extracts field correctly
3. Query Elasticsearch to confirm field exists:
   ```bash
   curl -X GET "localhost:9200/ollamamax-logs-*/_mapping/field/trace_id?pretty"
   ```

### High Memory Usage

**Logstash**:
```yaml
environment:
  - "LS_JAVA_OPTS=-Xms512m -Xmx512m"  # Increase heap
```

**Elasticsearch**:
```yaml
environment:
  - "ES_JAVA_OPTS=-Xms1g -Xmx1g"  # Increase heap
```

## Best Practices

1. **Index Lifecycle Management**:
   - Rotate daily indices
   - Delete old logs after 30 days
   - Archive to S3 for long-term storage

2. **Log Format**:
   - Use JSON structured logging
   - Include trace/span IDs
   - Add context fields (user_id, request_id)

3. **Performance**:
   - Use Filebeat compression
   - Configure Logstash workers
   - Monitor queue depths

4. **Security**:
   - Restrict Elasticsearch access
   - Use Kibana authentication
   - Audit log access

## Maintenance

### Index Cleanup

```bash
# Delete old indices
curl -X DELETE "localhost:9200/ollamamax-logs-2025.01.*"

# Create index template for retention
curl -X PUT "localhost:9200/_index_template/ollamamax-logs" -H 'Content-Type: application/json' -d'
{
  "index_patterns": ["ollamamax-logs-*"],
  "template": {
    "settings": {
      "index.lifecycle.name": "ollamamax-logs-policy",
      "index.lifecycle.rollover_alias": "ollamamax-logs"
    }
  }
}
'
```

### Pipeline Updates

```bash
# Reload Logstash configuration
docker exec logstash curl -XPOST 'localhost:9600/_node/pipeline/main/_reload'

# Restart Logstash
docker-compose restart logstash
```

## Integration with Jaeger

### Correlate Logs with Traces

1. **Find Trace ID in Kibana**:
   - Search logs by error or event
   - Copy `trace_id` field

2. **View Trace in Jaeger**:
   - Open Jaeger UI: http://localhost:16686
   - Search by trace ID
   - View complete request flow

3. **Create Kibana Link to Jaeger**:
   ```javascript
   // In Kibana Dashboard
   {
     "url": "http://localhost:16686/trace/{{trace_id}}",
     "label": "View Trace"
   }
   ```

## Performance Tuning

### Filebeat

```yaml
# Increase queue size
queue.mem:
  events: 4096
  flush.min_events: 512
  flush.timeout: 1s

# Increase workers
output.logstash:
  worker: 4
```

### Logstash

```conf
# Increase pipeline workers
pipeline.workers: 4

# Increase batch size
pipeline.batch.size: 125
```

### Elasticsearch

```yaml
# Increase refresh interval
index.refresh_interval: 30s

# Increase bulk queue
thread_pool.bulk.queue_size: 1000
```

## References

- [Filebeat Documentation](https://www.elastic.co/guide/en/beats/filebeat/current/index.html)
- [Logstash Documentation](https://www.elastic.co/guide/en/logstash/current/index.html)
- [Elasticsearch Documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/index.html)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)

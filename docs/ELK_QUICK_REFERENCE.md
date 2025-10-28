# ELK Stack Quick Reference

## Service URLs

```
Elasticsearch:  http://localhost:9200
Logstash API:   http://localhost:9600
Kibana:         http://localhost:5601
Jaeger:         http://localhost:16686
```

## Start Services

```bash
docker-compose up -d elasticsearch logstash kibana filebeat
```

## Validate Configuration

```bash
bash /home/kp/OllamaMax/scripts/validate-elk-stack.sh
```

## Health Checks

```bash
# Elasticsearch
curl http://localhost:9200/_cluster/health

# Logstash
curl http://localhost:9600/_node/stats

# Kibana
curl http://localhost:5601/api/status
```

## Query Logs by Trace ID

```bash
curl -X GET "localhost:9200/ollamamax-logs-*/_search" \
  -H 'Content-Type: application/json' \
  -d '{
    "query": {
      "term": { "trace_id": "YOUR_TRACE_ID" }
    }
  }'
```

## View Recent Errors

```bash
curl -X GET "localhost:9200/ollamamax-errors-*/_search" \
  -H 'Content-Type: application/json' \
  -d '{
    "size": 10,
    "sort": [{ "@timestamp": "desc" }]
  }' | jq '.hits.hits[]._source'
```

## Index Patterns (Kibana)

Create these index patterns in Kibana:
- `ollamamax-logs-*`
- `ollamamax-errors-*`
- `ollamamax-audit-*`

## Trace Correlation

1. Find trace_id in Kibana log
2. Open Jaeger: `http://localhost:16686/trace/{trace_id}`
3. View complete request trace

## Log Format (Application)

```json
{
  "timestamp": "2025-01-27T10:30:45.123Z",
  "level": "info",
  "message": "Your message",
  "trace_id": "abc123",
  "span_id": "def456",
  "service": "ollamamax-api",
  "audit": false
}
```

## Common Issues

### Filebeat not collecting
```bash
docker-compose logs filebeat
docker-compose restart filebeat
```

### Logstash pipeline errors
```bash
docker-compose logs logstash
docker-compose restart logstash
```

### Elasticsearch unhealthy
```bash
curl http://localhost:9200/_cluster/health?pretty
docker-compose restart elasticsearch
```

## Index Cleanup

```bash
# Delete old indices
curl -X DELETE "localhost:9200/ollamamax-logs-2025.01.*"
```

## Configuration Files

- Logstash: `/monitoring/logstash/pipeline/logstash.conf`
- Filebeat: `/monitoring/filebeat/filebeat.yml`
- Docker: `docker-compose.yml`

## Validation Script

```bash
cd /home/kp/OllamaMax
bash scripts/validate-elk-stack.sh
```

## Documentation

See `/docs/LOGSTASH_FILEBEAT_CONFIGURATION.md` for complete documentation.

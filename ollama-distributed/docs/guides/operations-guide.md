# Operations Guide - Ollama Distributed

Comprehensive operational guide for deploying, monitoring, and maintaining Ollama Distributed in production environments.

## Table of Contents

1. [Deployment Strategies](#deployment-strategies)
2. [Monitoring & Observability](#monitoring--observability)
3. [Disaster Recovery](#disaster-recovery)
4. [Scaling Operations](#scaling-operations)
5. [Maintenance Procedures](#maintenance-procedures)
6. [Security Operations](#security-operations)
7. [Troubleshooting Runbooks](#troubleshooting-runbooks)

## Deployment Strategies

### Production Deployment Architecture

#### High-Availability Setup
```
                    ┌─────────────────┐
                    │   Load Balancer │
                    │   (HAProxy/Nginx)│
                    └─────────┬───────┘
                              │
                    ┌─────────┴───────┐
                    │                 │
            ┌───────▼──────┐  ┌──────▼───────┐
            │  Ollama Node │  │  Ollama Node │
            │  (Leader)    │  │  (Follower)  │
            └──────┬───────┘  └──────┬───────┘
                   │                 │
            ┌──────▼────────┬────────▼──────┐
            │               │               │
    ┌───────▼──────┐ ┌─────▼──────┐ ┌──────▼───────┐
    │ Ollama Node  │ │ Ollama Node│ │  Ollama Node │
    │ (Follower)   │ │ (Follower) │ │  (Follower)  │
    └──────┬───────┘ └─────┬──────┘ └──────┬───────┘
           │               │               │
      ┌────▼──────────────▼──────────────▼────┐
      │        Shared Storage Layer         │
      │  (Distributed File System/S3)      │
      └─────────────────────────────────────┘
```

#### Multi-Region Deployment
```yaml
# deploy/production/regions.yaml
regions:
  us-east-1:
    nodes: 5
    models: [llama2, mistral, codellama]
    capacity: high
    
  us-west-1:
    nodes: 3
    models: [llama2, mistral]
    capacity: medium
    
  eu-west-1:
    nodes: 3
    models: [llama2]
    capacity: medium

replication:
  strategy: regional
  min_replicas: 2
  cross_region_sync: true
```

### Container Orchestration

#### Kubernetes Deployment
```yaml
# deploy/k8s/ollama-distributed.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: ollama-distributed
  labels:
    app: ollama-distributed
spec:
  serviceName: ollama-distributed-headless
  replicas: 5
  selector:
    matchLabels:
      app: ollama-distributed
  template:
    metadata:
      labels:
        app: ollama-distributed
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
      containers:
      - name: ollama-node
        image: ollama/distributed:latest
        ports:
        - containerPort: 8080
          name: http-api
        - containerPort: 8443
          name: https-api
        - containerPort: 7946
          name: p2p
        env:
        - name: OLLAMA_NODE_ID
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: OLLAMA_CLUSTER_PEERS
          value: "ollama-distributed-0.ollama-distributed-headless:8080,ollama-distributed-1.ollama-distributed-headless:8080"
        - name: OLLAMA_DATA_DIR
          value: "/data"
        - name: OLLAMA_MODELS_DIR
          value: "/data/models"
        volumeMounts:
        - name: data
          mountPath: /data
        - name: config
          mountPath: /etc/ollama
        resources:
          requests:
            memory: "8Gi"
            cpu: "2"
          limits:
            memory: "32Gi"
            cpu: "8"
        livenessProbe:
          httpGet:
            path: /api/v1/health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/v1/health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: ollama-config
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 500Gi
      storageClassName: fast-ssd

---
apiVersion: v1
kind: Service
metadata:
  name: ollama-distributed-headless
spec:
  clusterIP: None
  selector:
    app: ollama-distributed
  ports:
  - port: 8080
    targetPort: 8080
    name: http-api
  - port: 8443
    targetPort: 8443
    name: https-api
  - port: 7946
    targetPort: 7946
    name: p2p

---
apiVersion: v1
kind: Service
metadata:
  name: ollama-distributed-lb
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: nlb
spec:
  type: LoadBalancer
  selector:
    app: ollama-distributed
  ports:
  - port: 80
    targetPort: 8080
    name: http
  - port: 443
    targetPort: 8443
    name: https
```

#### Helm Chart
```yaml
# charts/ollama-distributed/values.yaml
replicaCount: 5

image:
  repository: ollama/distributed
  tag: latest
  pullPolicy: IfNotPresent

service:
  type: LoadBalancer
  port: 80
  targetPort: 8080

ingress:
  enabled: true
  className: nginx
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/websocket: "true"
  hosts:
    - host: ollama.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: ollama-tls
      hosts:
        - ollama.example.com

persistence:
  enabled: true
  size: 500Gi
  storageClass: fast-ssd

resources:
  limits:
    cpu: 8
    memory: 32Gi
  requests:
    cpu: 2
    memory: 8Gi

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80

nodeSelector:
  workload: ollama

tolerations:
  - key: "dedicated"
    operator: "Equal"
    value: "ollama"
    effect: "NoSchedule"

monitoring:
  enabled: true
  serviceMonitor:
    enabled: true
    interval: 30s

security:
  podSecurityPolicy:
    enabled: true
  networkPolicy:
    enabled: true
```

#### Docker Swarm Deployment
```yaml
# deploy/swarm/docker-compose.prod.yml
version: '3.8'

services:
  ollama-node:
    image: ollama/distributed:latest
    networks:
      - ollama-network
    ports:
      - "8080:8080"
      - "8443:8443"
    environment:
      - OLLAMA_CLUSTER_MODE=true
      - OLLAMA_NODE_ID={{.Node.ID}}
      - OLLAMA_CLUSTER_PEERS=tasks.ollama-node:8080
    volumes:
      - ollama-data:/data
      - ollama-models:/data/models
    deploy:
      replicas: 5
      placement:
        max_replicas_per_node: 1
        constraints:
          - node.role == worker
      resources:
        limits:
          cpus: '8.0'
          memory: 32G
        reservations:
          cpus: '2.0'
          memory: 8G
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
      update_config:
        parallelism: 1
        delay: 10s
        failure_action: rollback
        order: stop-first

  nginx-lb:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/ssl:ro
    networks:
      - ollama-network
    deploy:
      replicas: 2
      placement:
        constraints:
          - node.role == manager

volumes:
  ollama-data:
    driver: local
  ollama-models:
    driver: local

networks:
  ollama-network:
    driver: overlay
    attachable: true
```

### Cloud-Specific Deployments

#### AWS ECS with Fargate
```json
{
  "family": "ollama-distributed",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "2048",
  "memory": "8192",
  "executionRoleArn": "arn:aws:iam::account:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::account:role/ollamaTaskRole",
  "containerDefinitions": [
    {
      "name": "ollama-node",
      "image": "ollama/distributed:latest",
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "OLLAMA_CLUSTER_MODE",
          "value": "true"
        },
        {
          "name": "OLLAMA_DATA_DIR",
          "value": "/data"
        }
      ],
      "mountPoints": [
        {
          "sourceVolume": "ollama-data",
          "containerPath": "/data"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/ollama-distributed",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      },
      "healthCheck": {
        "command": [
          "CMD-SHELL",
          "curl -f http://localhost:8080/api/v1/health || exit 1"
        ],
        "interval": 30,
        "timeout": 5,
        "retries": 3,
        "startPeriod": 60
      }
    }
  ],
  "volumes": [
    {
      "name": "ollama-data",
      "efsVolumeConfiguration": {
        "fileSystemId": "fs-12345678",
        "rootDirectory": "/ollama"
      }
    }
  ]
}
```

## Monitoring & Observability

### Prometheus & Grafana Setup

#### Prometheus Configuration
```yaml
# deploy/monitoring/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "ollama_alerts.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

scrape_configs:
  - job_name: 'ollama-distributed'
    static_configs:
      - targets: 
        - 'ollama-node-1:8080'
        - 'ollama-node-2:8080'
        - 'ollama-node-3:8080'
    scrape_interval: 30s
    metrics_path: /metrics
    scrape_timeout: 10s
    
  - job_name: 'node-exporter'
    static_configs:
      - targets:
        - 'node-1:9100'
        - 'node-2:9100'
        - 'node-3:9100'
```

#### Grafana Dashboards
```json
{
  "dashboard": {
    "title": "Ollama Distributed - Cluster Overview",
    "panels": [
      {
        "title": "Cluster Health",
        "type": "stat",
        "targets": [
          {
            "expr": "up{job=\"ollama-distributed\"}",
            "legendFormat": "{{instance}}"
          }
        ]
      },
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(ollama_requests_total[5m])",
            "legendFormat": "Requests/sec"
          }
        ]
      },
      {
        "title": "Response Time P95",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(ollama_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          }
        ]
      },
      {
        "title": "Active Models",
        "type": "table",
        "targets": [
          {
            "expr": "ollama_models_loaded",
            "format": "table",
            "instant": true
          }
        ]
      }
    ]
  }
}
```

#### Alert Rules
```yaml
# deploy/monitoring/ollama_alerts.yml
groups:
  - name: ollama.rules
    rules:
      - alert: OllamaNodeDown
        expr: up{job="ollama-distributed"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Ollama node {{ $labels.instance }} is down"
          description: "Ollama node {{ $labels.instance }} has been down for more than 1 minute"

      - alert: HighRequestLatency
        expr: histogram_quantile(0.95, rate(ollama_request_duration_seconds_bucket[5m])) > 2
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High request latency on {{ $labels.instance }}"
          description: "95th percentile latency is {{ $value }}s"

      - alert: HighErrorRate
        expr: rate(ollama_requests_failed_total[5m]) / rate(ollama_requests_total[5m]) > 0.05
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "High error rate on {{ $labels.instance }}"
          description: "Error rate is {{ $value | humanizePercentage }}"

      - alert: LowDiskSpace
        expr: (node_filesystem_avail_bytes{mountpoint="/data"} / node_filesystem_size_bytes{mountpoint="/data"}) < 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Low disk space on {{ $labels.instance }}"
          description: "Only {{ $value | humanizePercentage }} disk space remaining"

      - alert: HighMemoryUsage
        expr: (1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) > 0.9
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage on {{ $labels.instance }}"
          description: "Memory usage is {{ $value | humanizePercentage }}"
```

### Logging Infrastructure

#### ELK Stack Configuration
```yaml
# deploy/logging/elasticsearch.yml
version: '3.8'
services:
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.11.0
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
      - "ES_JAVA_OPTS=-Xms2g -Xmx2g"
    volumes:
      - elasticsearch-data:/usr/share/elasticsearch/data
    ports:
      - "9200:9200"

  logstash:
    image: docker.elastic.co/logstash/logstash:8.11.0
    volumes:
      - ./logstash.conf:/usr/share/logstash/pipeline/logstash.conf:ro
    ports:
      - "5044:5044"
    depends_on:
      - elasticsearch

  kibana:
    image: docker.elastic.co/kibana/kibana:8.11.0
    ports:
      - "5601:5601"
    environment:
      - ELASTICSEARCH_HOSTS=http://elasticsearch:9200
    depends_on:
      - elasticsearch

volumes:
  elasticsearch-data:
```

```ruby
# deploy/logging/logstash.conf
input {
  beats {
    port => 5044
  }
}

filter {
  if [fields][service] == "ollama-distributed" {
    json {
      source => "message"
    }
    
    date {
      match => [ "timestamp", "ISO8601" ]
    }
    
    if [level] == "ERROR" {
      mutate {
        add_tag => [ "error" ]
      }
    }
    
    if [component] == "p2p" {
      mutate {
        add_field => { "category" => "networking" }
      }
    }
  }
}

output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "ollama-distributed-%{+YYYY.MM.dd}"
  }
}
```

#### Structured Logging Configuration
```yaml
# config/logging.yaml
logging:
  level: info
  format: json
  output: stdout
  
  components:
    p2p: debug
    consensus: info
    models: info
    api: warn
    
  structured_fields:
    - service: ollama-distributed
    - version: ${OLLAMA_VERSION}
    - node_id: ${OLLAMA_NODE_ID}
    - environment: ${ENVIRONMENT}
    
  sampling:
    enabled: true
    rate: 0.1  # Sample 10% of debug logs
    
  correlation:
    enabled: true
    header_name: X-Correlation-ID
```

### Distributed Tracing

#### Jaeger Configuration
```yaml
# deploy/tracing/jaeger.yml
version: '3.8'
services:
  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686"
      - "14268:14268"
    environment:
      - COLLECTOR_OTLP_ENABLED=true
```

#### OpenTelemetry Setup
```go
// pkg/telemetry/tracing.go
package telemetry

import (
    "context"
    
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/sdk/resource"
    tracesdk "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func InitTracing(serviceName, jaegerURL string) error {
    exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerURL)))
    if err != nil {
        return err
    }
    
    tp := tracesdk.NewTracerProvider(
        tracesdk.WithBatcher(exp),
        tracesdk.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String(serviceName),
            semconv.ServiceVersionKey.String("1.0.0"),
        )),
    )
    
    otel.SetTracerProvider(tp)
    return nil
}
```

## Disaster Recovery

### Backup Strategies

#### Automated Backup System
```bash
#!/bin/bash
# scripts/backup.sh

set -euo pipefail

BACKUP_DIR="/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=30

# Backup cluster configuration
echo "Backing up cluster configuration..."
./ollama-distributed config export > "${BACKUP_DIR}/config_${TIMESTAMP}.json"

# Backup model registry
echo "Backing up model registry..."
./ollama-distributed models export > "${BACKUP_DIR}/models_${TIMESTAMP}.json"

# Backup consensus state
echo "Backing up consensus state..."
tar -czf "${BACKUP_DIR}/consensus_${TIMESTAMP}.tar.gz" /data/consensus/

# Backup critical models
echo "Backing up critical models..."
./ollama-distributed models backup --critical-only --output "${BACKUP_DIR}/critical_models_${TIMESTAMP}.tar.gz"

# Upload to cloud storage
echo "Uploading to cloud storage..."
aws s3 sync "${BACKUP_DIR}/" "s3://ollama-backups/$(hostname)/" --delete

# Cleanup old backups
echo "Cleaning up old backups..."
find "${BACKUP_DIR}" -name "*.tar.gz" -mtime +${RETENTION_DAYS} -delete
find "${BACKUP_DIR}" -name "*.json" -mtime +${RETENTION_DAYS} -delete

echo "Backup completed successfully"
```

```cron
# Crontab entry for automated backups
0 2 * * * /opt/ollama/scripts/backup.sh >> /var/log/ollama-backup.log 2>&1
0 6 * * 0 /opt/ollama/scripts/backup-verify.sh >> /var/log/ollama-backup-verify.log 2>&1
```

#### Backup Verification
```bash
#!/bin/bash
# scripts/backup-verify.sh

BACKUP_DIR="/backups"
LATEST_BACKUP=$(ls -t ${BACKUP_DIR}/config_*.json | head -n1)

if [[ -z "$LATEST_BACKUP" ]]; then
    echo "ERROR: No backup found"
    exit 1
fi

# Verify backup integrity
echo "Verifying backup integrity..."
if ! jq empty "$LATEST_BACKUP" 2>/dev/null; then
    echo "ERROR: Invalid JSON in backup file"
    exit 1
fi

# Test restore (dry run)
echo "Testing restore process..."
./ollama-distributed config import --dry-run --file "$LATEST_BACKUP"

if [[ $? -eq 0 ]]; then
    echo "Backup verification successful"
    exit 0
else
    echo "ERROR: Backup verification failed"
    exit 1
fi
```

### Recovery Procedures

#### Complete Cluster Recovery
```bash
#!/bin/bash
# scripts/disaster-recovery.sh

set -euo pipefail

BACKUP_DATE=${1:-latest}
BACKUP_DIR="/backups"

echo "Starting disaster recovery process..."

# Stop all nodes
echo "Stopping all cluster nodes..."
for node in node-1 node-2 node-3; do
    ssh $node "systemctl stop ollama-distributed"
done

# Clear existing data
echo "Clearing existing data..."
for node in node-1 node-2 node-3; do
    ssh $node "rm -rf /data/consensus/* /data/models/* /data/cache/*"
done

# Restore configuration
echo "Restoring cluster configuration..."
if [[ "$BACKUP_DATE" == "latest" ]]; then
    BACKUP_FILE=$(ls -t ${BACKUP_DIR}/config_*.json | head -n1)
else
    BACKUP_FILE="${BACKUP_DIR}/config_${BACKUP_DATE}.json"
fi

./ollama-distributed config import --file "$BACKUP_FILE"

# Restore models
echo "Restoring critical models..."
MODEL_BACKUP=$(ls -t ${BACKUP_DIR}/critical_models_*.tar.gz | head -n1)
tar -xzf "$MODEL_BACKUP" -C /tmp/
./ollama-distributed models restore --from /tmp/models/

# Start bootstrap node
echo "Starting bootstrap node..."
ssh node-1 "systemctl start ollama-distributed"

# Wait for bootstrap
sleep 30

# Start remaining nodes
echo "Starting remaining nodes..."
for node in node-2 node-3; do
    ssh $node "systemctl start ollama-distributed"
    sleep 10
done

# Verify cluster health
echo "Verifying cluster health..."
./ollama-distributed cluster status

if [[ $? -eq 0 ]]; then
    echo "Disaster recovery completed successfully"
else
    echo "ERROR: Cluster not healthy after recovery"
    exit 1
fi
```

#### Single Node Recovery
```bash
#!/bin/bash
# scripts/node-recovery.sh

NODE_ID=${1:?Node ID required}
BACKUP_DATE=${2:-latest}

echo "Recovering node: $NODE_ID"

# Drain the node
echo "Draining node..."
./ollama-distributed node drain "$NODE_ID"

# Stop the node
echo "Stopping node..."
ssh "$NODE_ID" "systemctl stop ollama-distributed"

# Clear node data
echo "Clearing node data..."
ssh "$NODE_ID" "rm -rf /data/models/* /data/cache/*"

# Restore node-specific data if available
if [[ -f "/backups/node_${NODE_ID}_${BACKUP_DATE}.tar.gz" ]]; then
    echo "Restoring node-specific data..."
    scp "/backups/node_${NODE_ID}_${BACKUP_DATE}.tar.gz" "$NODE_ID:/tmp/"
    ssh "$NODE_ID" "tar -xzf /tmp/node_${NODE_ID}_${BACKUP_DATE}.tar.gz -C /data/"
fi

# Restart the node
echo "Restarting node..."
ssh "$NODE_ID" "systemctl start ollama-distributed"

# Wait for node to rejoin
echo "Waiting for node to rejoin cluster..."
timeout 300 bash -c "
    while ! ./ollama-distributed node status '$NODE_ID' | grep -q 'online'; do
        sleep 5
    done
"

# Undrain the node
echo "Undraining node..."
./ollama-distributed node undrain "$NODE_ID"

echo "Node recovery completed"
```

## Scaling Operations

### Horizontal Scaling

#### Auto-scaling Based on Metrics
```yaml
# deploy/autoscaling/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: ollama-distributed-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: StatefulSet
    name: ollama-distributed
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  - type: Pods
    pods:
      metric:
        name: requests_per_second
      target:
        type: AverageValue
        averageValue: "100"
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Pods
        value: 1
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Pods
        value: 2
        periodSeconds: 60
      - type: Percent
        value: 50
        periodSeconds: 60
```

#### Manual Scaling Operations
```bash
#!/bin/bash
# scripts/scale-cluster.sh

ACTION=${1:?Action required: scale-up|scale-down|rebalance}
TARGET_NODES=${2:-5}

case "$ACTION" in
    scale-up)
        echo "Scaling cluster up to $TARGET_NODES nodes..."
        
        # Add new nodes
        for i in $(seq $(( $(./ollama-distributed cluster size) + 1)) $TARGET_NODES); do
            echo "Adding node-$i..."
            ./scripts/provision-node.sh "node-$i"
            
            # Wait for node to join
            timeout 300 bash -c "
                while ! ./ollama-distributed node status 'node-$i' | grep -q 'online'; do
                    sleep 10
                done
            "
        done
        
        # Rebalance models
        echo "Rebalancing models across new nodes..."
        ./ollama-distributed models rebalance
        ;;
        
    scale-down)
        echo "Scaling cluster down to $TARGET_NODES nodes..."
        CURRENT_NODES=$(./ollama-distributed cluster size)
        
        for i in $(seq $TARGET_NODES $((CURRENT_NODES - 1))); do
            NODE_ID="node-$i"
            echo "Removing $NODE_ID..."
            
            # Drain the node
            ./ollama-distributed node drain "$NODE_ID"
            
            # Wait for models to be moved
            timeout 600 bash -c "
                while ./ollama-distributed node status '$NODE_ID' | grep -q 'has_models'; do
                    sleep 30
                done
            "
            
            # Remove from cluster
            ./ollama-distributed node remove "$NODE_ID"
        done
        ;;
        
    rebalance)
        echo "Rebalancing cluster..."
        ./ollama-distributed models rebalance --strategy even
        ./ollama-distributed cluster optimize
        ;;
        
    *)
        echo "Unknown action: $ACTION"
        echo "Usage: $0 {scale-up|scale-down|rebalance} [target-nodes]"
        exit 1
        ;;
esac

echo "Scaling operation completed"
```

### Vertical Scaling

#### Resource Adjustment
```bash
#!/bin/bash
# scripts/vertical-scale.sh

NODE_ID=${1:?Node ID required}
CPU_LIMIT=${2:?CPU limit required}
MEMORY_LIMIT=${3:?Memory limit required}

echo "Vertically scaling $NODE_ID to ${CPU_LIMIT} CPU, ${MEMORY_LIMIT} memory"

# Update resource limits in Kubernetes
kubectl patch statefulset ollama-distributed --patch "
spec:
  template:
    spec:
      containers:
      - name: ollama-node
        resources:
          limits:
            cpu: '${CPU_LIMIT}'
            memory: '${MEMORY_LIMIT}'
          requests:
            cpu: '$((CPU_LIMIT / 2))'
            memory: '$((MEMORY_LIMIT / 2))'
"

# Rolling restart to apply new limits
kubectl rollout restart statefulset/ollama-distributed

# Wait for rollout to complete
kubectl rollout status statefulset/ollama-distributed --timeout=600s

echo "Vertical scaling completed"
```

## Maintenance Procedures

### Rolling Updates

#### Zero-Downtime Update Process
```bash
#!/bin/bash
# scripts/rolling-update.sh

NEW_VERSION=${1:?New version required}
HEALTH_CHECK_URL="http://localhost:8080/api/v1/health"

echo "Starting rolling update to version $NEW_VERSION"

# Get list of nodes
NODES=($(./ollama-distributed cluster nodes --output names))

for NODE in "${NODES[@]}"; do
    echo "Updating node: $NODE"
    
    # Drain the node
    echo "Draining $NODE..."
    ./ollama-distributed node drain "$NODE"
    
    # Wait for requests to drain
    echo "Waiting for requests to drain..."
    sleep 60
    
    # Update the node
    echo "Updating $NODE to version $NEW_VERSION..."
    ssh "$NODE" "docker pull ollama/distributed:$NEW_VERSION"
    ssh "$NODE" "systemctl stop ollama-distributed"
    ssh "$NODE" "sed -i 's/ollama\/distributed:.*/ollama\/distributed:$NEW_VERSION/' /etc/systemd/system/ollama-distributed.service"
    ssh "$NODE" "systemctl daemon-reload"
    ssh "$NODE" "systemctl start ollama-distributed"
    
    # Health check
    echo "Waiting for $NODE to become healthy..."
    timeout 300 bash -c "
        while ! curl -sf http://$NODE:8080/api/v1/health > /dev/null; do
            sleep 10
        done
    "
    
    # Undrain the node
    echo "Undraining $NODE..."
    ./ollama-distributed node undrain "$NODE"
    
    # Wait before next node
    sleep 30
done

echo "Rolling update completed successfully"
```

### Routine Maintenance

#### Daily Maintenance Script
```bash
#!/bin/bash
# scripts/daily-maintenance.sh

LOG_FILE="/var/log/ollama-maintenance.log"
exec 1> >(tee -a "$LOG_FILE")
exec 2>&1

echo "Starting daily maintenance: $(date)"

# Health checks
echo "Running health checks..."
./ollama-distributed health check --verbose

# Cleanup operations
echo "Running cleanup operations..."
./ollama-distributed models cleanup --unused --older-than 7d
./ollama-distributed cache cleanup --size-limit 100GB
./ollama-distributed logs cleanup --older-than 30d

# Performance optimization
echo "Running performance optimization..."
./ollama-distributed optimize --rebalance-models
./ollama-distributed optimize --compact-storage

# Security updates
echo "Checking for security updates..."
./ollama-distributed security scan --fix-auto

# Metrics collection
echo "Collecting metrics..."
./ollama-distributed metrics collect --store /var/lib/ollama/metrics/daily/

# Backup verification
echo "Verifying backups..."
./scripts/backup-verify.sh

# Certificate renewal
echo "Checking certificate expiry..."
./ollama-distributed security cert-check --renew-threshold 30d

echo "Daily maintenance completed: $(date)"
```

#### Weekly Maintenance Script
```bash
#!/bin/bash
# scripts/weekly-maintenance.sh

LOG_FILE="/var/log/ollama-weekly-maintenance.log"
exec 1> >(tee -a "$LOG_FILE")
exec 2>&1

echo "Starting weekly maintenance: $(date)"

# Deep health check
echo "Running comprehensive health check..."
./ollama-distributed health check --comprehensive

# Performance analysis
echo "Running performance analysis..."
./ollama-distributed analyze performance --week
./ollama-distributed analyze usage-patterns --week

# Security audit
echo "Running security audit..."
./ollama-distributed security audit --comprehensive

# Storage optimization
echo "Optimizing storage..."
./ollama-distributed storage optimize --defragment
./ollama-distributed storage optimize --compress-logs

# Capacity planning
echo "Updating capacity planning..."
./ollama-distributed capacity-plan update
./ollama-distributed capacity-plan forecast --weeks 4

# Update checks
echo "Checking for updates..."
./ollama-distributed update check --security-only

# Generate weekly report
echo "Generating weekly report..."
./ollama-distributed report generate --type weekly --output /var/reports/

echo "Weekly maintenance completed: $(date)"
```

## Security Operations

### Security Monitoring

#### Security Event Detection
```bash
#!/bin/bash
# scripts/security-monitor.sh

ALERT_THRESHOLD=5
LOG_FILE="/var/log/ollama-security.log"

# Monitor failed authentication attempts
failed_auth=$(journalctl -u ollama-distributed --since "1 hour ago" | grep "authentication failed" | wc -l)
if [[ $failed_auth -gt $ALERT_THRESHOLD ]]; then
    echo "ALERT: $failed_auth failed authentication attempts in the last hour" | tee -a "$LOG_FILE"
    ./scripts/send-alert.sh "security" "High number of failed authentication attempts"
fi

# Monitor unusual API access patterns
unusual_patterns=$(./ollama-distributed security analyze --api-patterns --since "1 hour ago" | grep "suspicious" | wc -l)
if [[ $unusual_patterns -gt 0 ]]; then
    echo "ALERT: Suspicious API access patterns detected" | tee -a "$LOG_FILE"
    ./scripts/send-alert.sh "security" "Suspicious API access patterns"
fi

# Check for privilege escalation attempts
privilege_escalation=$(journalctl -u ollama-distributed --since "1 hour ago" | grep "privilege escalation" | wc -l)
if [[ $privilege_escalation -gt 0 ]]; then
    echo "CRITICAL: Privilege escalation attempt detected" | tee -a "$LOG_FILE"
    ./scripts/send-alert.sh "security-critical" "Privilege escalation attempt"
fi

# Monitor file integrity
if ! ./ollama-distributed security file-integrity-check; then
    echo "ALERT: File integrity check failed" | tee -a "$LOG_FILE"
    ./scripts/send-alert.sh "security" "File integrity compromised"
fi
```

#### Certificate Management
```bash
#!/bin/bash
# scripts/cert-management.sh

CERT_DIR="/etc/ollama/certs"
RENEWAL_THRESHOLD=30  # days

echo "Checking certificate expiry..."

for cert_file in "$CERT_DIR"/*.crt; do
    if [[ -f "$cert_file" ]]; then
        cert_name=$(basename "$cert_file" .crt)
        expiry_date=$(openssl x509 -enddate -noout -in "$cert_file" | cut -d= -f2)
        expiry_timestamp=$(date -d "$expiry_date" +%s)
        current_timestamp=$(date +%s)
        days_until_expiry=$(( (expiry_timestamp - current_timestamp) / 86400 ))
        
        echo "Certificate $cert_name expires in $days_until_expiry days"
        
        if [[ $days_until_expiry -lt $RENEWAL_THRESHOLD ]]; then
            echo "Renewing certificate: $cert_name"
            ./ollama-distributed security cert-renew "$cert_name"
            
            if [[ $? -eq 0 ]]; then
                echo "Certificate renewed successfully: $cert_name"
                ./scripts/rolling-restart.sh  # Apply new certificate
            else
                echo "ERROR: Failed to renew certificate: $cert_name"
                ./scripts/send-alert.sh "security-critical" "Certificate renewal failed: $cert_name"
            fi
        fi
    fi
done
```

## Troubleshooting Runbooks

### Common Issues Runbook

#### Node Connectivity Issues
```markdown
## Problem: Node Cannot Join Cluster

### Symptoms
- Node reports "failed to connect to cluster"
- Cluster doesn't see the new node
- Network timeout errors

### Diagnosis Steps
1. Check network connectivity:
   ```bash
   ping <cluster-leader-ip>
   telnet <cluster-leader-ip> 8080
   ```

2. Verify firewall rules:
   ```bash
   sudo iptables -L | grep 8080
   sudo ufw status
   ```

3. Check node configuration:
   ```bash
   ./ollama-distributed config validate
   ```

4. Review logs:
   ```bash
   journalctl -u ollama-distributed -f
   ```

### Resolution Steps
1. Fix network connectivity issues
2. Update firewall rules to allow required ports
3. Correct configuration errors
4. Restart the node service

### Prevention
- Use configuration validation in CI/CD
- Implement network connectivity monitoring
- Document firewall requirements
```

#### High Memory Usage
```markdown
## Problem: Node Running Out of Memory

### Symptoms
- OOM killer messages in system logs
- Node becomes unresponsive
- High swap usage

### Diagnosis Steps
1. Check memory usage:
   ```bash
   free -h
   ps aux --sort=-%mem | head -20
   ```

2. Check model memory usage:
   ```bash
   ./ollama-distributed models status --memory-usage
   ```

3. Review memory limits:
   ```bash
   ./ollama-distributed config get memory-limits
   ```

### Resolution Steps
1. Immediate: Restart the most memory-intensive models
2. Short-term: Reduce number of loaded models
3. Long-term: Scale horizontally or increase node memory

### Prevention
- Implement memory monitoring and alerting
- Set appropriate memory limits per model
- Use memory-efficient model loading strategies
```

#### Performance Degradation
```markdown
## Problem: Slow Response Times

### Symptoms
- High latency in API responses
- Request timeouts
- User complaints about slow performance

### Diagnosis Steps
1. Check system resources:
   ```bash
   top
   iostat -x 1
   ```

2. Analyze request patterns:
   ```bash
   ./ollama-distributed metrics requests --analyze
   ```

3. Check model distribution:
   ```bash
   ./ollama-distributed models status --distribution
   ```

### Resolution Steps
1. Identify bottlenecks (CPU, memory, disk, network)
2. Rebalance model distribution
3. Scale cluster if needed
4. Optimize model loading

### Prevention
- Implement performance monitoring
- Set up automated load balancing
- Regular performance testing
```

---

This operations guide provides comprehensive procedures for managing Ollama Distributed in production environments. Regular review and updates of these procedures ensure operational excellence and system reliability.

For additional operational support, refer to the [Security Guide](./security-guide.md) and [Performance Guide](./performance-guide.md).
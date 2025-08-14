# OllamaMax Production Deployment Guide

## Overview

OllamaMax is a production-ready, distributed version of Ollama with enterprise-grade features including P2P networking, Raft consensus, horizontal scaling, and comprehensive monitoring.

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Node 1        │    │   Node 2        │    │   Node 3        │
│                 │    │                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ API Server  │ │    │ │ API Server  │ │    │ │ API Server  │ │
│ │ :8080       │ │    │ │ :8080       │ │    │ │ :8080       │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ P2P Network │◄┼────┼►│ P2P Network │◄┼────┼►│ P2P Network │ │
│ │ :9000       │ │    │ │ :9000       │ │    │ │ :9000       │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ Raft        │ │    │ │ Raft        │ │    │ │ Raft        │ │
│ │ Consensus   │ │    │ │ Consensus   │ │    │ │ Consensus   │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Prerequisites

- **Go 1.21+** for building from source
- **Docker** (optional, for containerized deployment)
- **Linux/macOS** (Windows support experimental)
- **Minimum 4GB RAM** per node
- **Network connectivity** between nodes (P2P ports)

## Quick Start

### 1. Build from Source

```bash
git clone https://github.com/KhryptorGraphics/OllamaMax.git
cd OllamaMax/ollama-distributed
go build -o ollamamax ./cmd/ollamamax
```

### 2. Generate Configuration

```bash
./ollamamax init --config-dir ./config
```

### 3. Start First Node (Bootstrap)

```bash
./ollamamax start \
  --config ./config/node1.yaml \
  --bootstrap \
  --api-port 8080 \
  --p2p-port 9000
```

### 4. Start Additional Nodes

```bash
# Node 2
./ollamamax start \
  --config ./config/node2.yaml \
  --bootstrap-peer /ip4/NODE1_IP/tcp/9000/p2p/NODE1_PEER_ID \
  --api-port 8081 \
  --p2p-port 9001

# Node 3
./ollamamax start \
  --config ./config/node3.yaml \
  --bootstrap-peer /ip4/NODE1_IP/tcp/9000/p2p/NODE1_PEER_ID \
  --api-port 8082 \
  --p2p-port 9002
```

## Configuration

### Node Configuration (`config.yaml`)

```yaml
# Node identification
node:
  id: "node-1"
  data_dir: "/var/lib/ollamamax"

# API server configuration
api:
  listen: "0.0.0.0:8080"
  tls:
    enabled: true
    cert_file: "/etc/ollamamax/tls/server.crt"
    key_file: "/etc/ollamamax/tls/server.key"
  auth:
    enabled: true
    jwt_secret: "your-jwt-secret-here"
  rate_limit:
    enabled: true
    requests_per_minute: 1000

# P2P networking
p2p:
  listen: "/ip4/0.0.0.0/tcp/9000"
  bootstrap_peers:
    - "/ip4/10.0.1.10/tcp/9000/p2p/QmBootstrapPeerID"
  enable_dht: true
  enable_mdns: true

# Raft consensus
consensus:
  data_dir: "/var/lib/ollamamax/raft"
  bind_addr: "0.0.0.0:7000"
  advertise_addr: "10.0.1.10:7000"
  bootstrap: false
  bootstrap_expect: 3

# Scheduler configuration
scheduler:
  strategy: "load_balanced"
  health_check_interval: "30s"
  node_timeout: "60s"

# Logging
logging:
  level: "info"
  format: "json"
  output: "/var/log/ollamamax/ollamamax.log"

# Metrics
metrics:
  enabled: true
  listen: "0.0.0.0:9090"
  path: "/metrics"
```

## Security Configuration

### 1. TLS/HTTPS Setup

Generate TLS certificates:

```bash
# Self-signed (development)
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -days 365 -nodes

# Production (use Let's Encrypt or your CA)
certbot certonly --standalone -d your-domain.com
```

### 2. Authentication

Configure JWT authentication:

```yaml
api:
  auth:
    enabled: true
    jwt_secret: "your-secure-jwt-secret-256-bits"
    token_expiry: "24h"
```

### 3. Network Security

Configure firewall rules:

```bash
# API server
ufw allow 8080/tcp

# P2P networking
ufw allow 9000/tcp

# Raft consensus (internal only)
ufw allow from 10.0.1.0/24 to any port 7000

# Metrics (internal only)
ufw allow from 10.0.1.0/24 to any port 9090
```

## Docker Deployment

### 1. Build Docker Image

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o ollamamax ./cmd/ollamamax

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/ollamamax .
EXPOSE 8080 9000 7000 9090
CMD ["./ollamamax", "start"]
```

### 2. Docker Compose

```yaml
version: '3.8'
services:
  ollamamax-node1:
    build: .
    ports:
      - "8080:8080"
      - "9000:9000"
    environment:
      - NODE_ID=node-1
      - BOOTSTRAP=true
    volumes:
      - ./config/node1.yaml:/config.yaml
      - ollamamax-data-1:/var/lib/ollamamax

  ollamamax-node2:
    build: .
    ports:
      - "8081:8080"
      - "9001:9000"
    environment:
      - NODE_ID=node-2
      - BOOTSTRAP_PEER=/ip4/ollamamax-node1/tcp/9000/p2p/NODE1_PEER_ID
    volumes:
      - ./config/node2.yaml:/config.yaml
      - ollamamax-data-2:/var/lib/ollamamax
    depends_on:
      - ollamamax-node1

volumes:
  ollamamax-data-1:
  ollamamax-data-2:
```

## Kubernetes Deployment

### 1. StatefulSet

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: ollamamax
spec:
  serviceName: ollamamax
  replicas: 3
  selector:
    matchLabels:
      app: ollamamax
  template:
    metadata:
      labels:
        app: ollamamax
    spec:
      containers:
      - name: ollamamax
        image: ollamamax:latest
        ports:
        - containerPort: 8080
          name: api
        - containerPort: 9000
          name: p2p
        - containerPort: 7000
          name: raft
        volumeMounts:
        - name: data
          mountPath: /var/lib/ollamamax
        - name: config
          mountPath: /config.yaml
          subPath: config.yaml
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 100Gi
```

## Monitoring and Observability

### 1. Prometheus Configuration

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'ollamamax'
    static_configs:
      - targets: ['node1:9090', 'node2:9090', 'node3:9090']
    metrics_path: '/api/v1/metrics'
```

### 2. Grafana Dashboard

Key metrics to monitor:
- Node health and availability
- P2P peer connections
- Raft consensus state
- API request rates and latency
- Model inference performance
- Resource utilization (CPU, memory, disk)

### 3. Alerting Rules

```yaml
groups:
- name: ollamamax
  rules:
  - alert: NodeDown
    expr: up{job="ollamamax"} == 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "OllamaMax node is down"

  - alert: HighAPILatency
    expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High API latency detected"
```

## Operations

### Health Checks

```bash
# Node health
curl -X GET http://localhost:8080/api/v1/health

# Cluster status
curl -X GET http://localhost:8080/api/v1/cluster/status

# Metrics
curl -X GET http://localhost:8080/api/v1/metrics
```

### Node Management

```bash
# Drain node (stop accepting new requests)
curl -X POST http://localhost:8080/api/v1/nodes/{node-id}/drain

# Undrain node
curl -X POST http://localhost:8080/api/v1/nodes/{node-id}/undrain

# Join cluster
curl -X POST http://localhost:8080/api/v1/cluster/join \
  -H "Content-Type: application/json" \
  -d '{"node_id": "new-node", "address": "10.0.1.20:8080"}'
```

### Backup and Recovery

```bash
# Backup Raft data
tar -czf raft-backup-$(date +%Y%m%d).tar.gz /var/lib/ollamamax/raft/

# Backup models
tar -czf models-backup-$(date +%Y%m%d).tar.gz /var/lib/ollamamax/models/
```

## Troubleshooting

### Common Issues

1. **Nodes not discovering each other**
   - Check P2P port connectivity
   - Verify bootstrap peer addresses
   - Check firewall rules

2. **Raft consensus issues**
   - Ensure odd number of nodes (3, 5, 7)
   - Check network connectivity on Raft ports
   - Verify time synchronization (NTP)

3. **High memory usage**
   - Monitor model loading
   - Implement model eviction policies
   - Scale horizontally

### Log Analysis

```bash
# View logs
journalctl -u ollamamax -f

# Search for errors
grep -i error /var/log/ollamamax/ollamamax.log

# P2P connectivity issues
grep -i "peer\|connection" /var/log/ollamamax/ollamamax.log
```

## Performance Tuning

### Resource Allocation

- **CPU**: 4+ cores per node
- **Memory**: 8GB+ per node (depends on models)
- **Storage**: SSD recommended, 100GB+ per node
- **Network**: 1Gbps+ between nodes

### Optimization Tips

1. **Model Distribution**: Distribute large models across multiple nodes
2. **Load Balancing**: Use round-robin or least-connections strategies
3. **Caching**: Implement model caching for frequently used models
4. **Connection Pooling**: Optimize P2P connection management

## Support

- **Documentation**: [docs/](./docs/)
- **Issues**: [GitHub Issues](https://github.com/KhryptorGraphics/OllamaMax/issues)
- **Community**: [Discussions](https://github.com/KhryptorGraphics/OllamaMax/discussions)

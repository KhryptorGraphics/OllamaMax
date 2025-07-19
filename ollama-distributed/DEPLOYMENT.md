# Ollama Distributed - Deployment Guide

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or later
- Docker (optional)
- Kubernetes (optional)

### 1. Build the Binary

```bash
# Clone the repository
git clone https://github.com/ollama/ollama-distributed.git
cd ollama-distributed

# Build the binary
make build

# Or build for all platforms
make build-all
```

### 2. Start a Single Node

```bash
# Start with default configuration
./bin/ollama-distributed start

# Or with custom configuration
./bin/ollama-distributed start --config config/node.yaml
```

### 3. Access the Web Interface

Open your browser to `http://localhost:8080` to access the web control panel.

## 🏗️ Architecture Overview

### Core Components

```
┌─────────────────────────────────────────────────────────────────┐
│                    Ollama Distributed                          │
├─────────────────────────────────────────────────────────────────┤
│  Web Control Panel (React + WebSocket)                         │
├─────────────────────────────────────────────────────────────────┤
│  REST API Server (Gin + Gorilla WebSocket)                     │
├─────────────────────────────────────────────────────────────────┤
│  Distributed Scheduler (Load Balancing + Health Monitoring)    │
├─────────────────────────────────────────────────────────────────┤
│  Model Distribution (P2P Transfer + Checksumming)              │
├─────────────────────────────────────────────────────────────────┤
│  Consensus Engine (Raft + Fault Tolerance)                     │
├─────────────────────────────────────────────────────────────────┤
│  P2P Networking (libp2p + DHT + PubSub)                        │
└─────────────────────────────────────────────────────────────────┘
```

### Key Features

- **🌐 P2P Networking**: libp2p-based mesh networking with automatic peer discovery
- **🔒 Security**: JWT authentication, X.509 certificates, RBAC authorization
- **⚖️ Consensus**: Raft-based consensus for distributed coordination
- **📊 Load Balancing**: Multiple algorithms (round-robin, least-connections, random)
- **🔄 Model Distribution**: P2P model transfer with content verification
- **📈 Monitoring**: Real-time metrics and health monitoring
- **🕸️ Web UI**: Modern React-based control panel with WebSocket updates

## 🏃 Production Deployment

### 1. Kubernetes Deployment

```yaml
# k8s/ollama-distributed.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: ollama-distributed
spec:
  serviceName: ollama-distributed
  replicas: 3
  selector:
    matchLabels:
      app: ollama-distributed
  template:
    metadata:
      labels:
        app: ollama-distributed
    spec:
      containers:
      - name: ollama-distributed
        image: ollama-distributed:latest
        ports:
        - containerPort: 11434
          name: api
        - containerPort: 8080
          name: web
        - containerPort: 4001
          name: p2p
        - containerPort: 7000
          name: consensus
        env:
        - name: OLLAMA_API_LISTEN
          value: "0.0.0.0:11434"
        - name: OLLAMA_WEB_LISTEN
          value: "0.0.0.0:8080"
        - name: OLLAMA_P2P_LISTEN
          value: "/ip4/0.0.0.0/tcp/4001"
        - name: OLLAMA_CONSENSUS_BIND_ADDR
          value: "0.0.0.0:7000"
        volumeMounts:
        - name: data
          mountPath: /data
        - name: models
          mountPath: /models
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 10Gi
  - metadata:
      name: models
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 100Gi
```

### 2. Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  ollama-node-1:
    image: ollama-distributed:latest
    command: ["start", "--config", "/app/config/node.yaml"]
    ports:
      - "11434:11434"
      - "8080:8080"
    volumes:
      - ./data/node1:/data
      - ./models:/models
      - ./config:/app/config
    environment:
      - OLLAMA_NODE_NAME=node-1
      - OLLAMA_CONSENSUS_BOOTSTRAP=true

  ollama-node-2:
    image: ollama-distributed:latest
    command: ["start", "--config", "/app/config/node.yaml"]
    ports:
      - "11435:11434"
      - "8081:8080"
    volumes:
      - ./data/node2:/data
      - ./models:/models
      - ./config:/app/config
    environment:
      - OLLAMA_NODE_NAME=node-2
      - OLLAMA_P2P_BOOTSTRAP=/ip4/ollama-node-1/tcp/4001
    depends_on:
      - ollama-node-1

  ollama-node-3:
    image: ollama-distributed:latest
    command: ["start", "--config", "/app/config/node.yaml"]
    ports:
      - "11436:11434"
      - "8082:8080"
    volumes:
      - ./data/node3:/data
      - ./models:/models
      - ./config:/app/config
    environment:
      - OLLAMA_NODE_NAME=node-3
      - OLLAMA_P2P_BOOTSTRAP=/ip4/ollama-node-1/tcp/4001
    depends_on:
      - ollama-node-1
```

### 3. Bare Metal Deployment

```bash
# Node 1 (Bootstrap)
./bin/ollama-distributed start \
  --config config/node.yaml \
  --listen 0.0.0.0:11434 \
  --p2p-listen /ip4/0.0.0.0/tcp/4001 \
  --data-dir /opt/ollama/data \
  --enable-web

# Node 2 (Join cluster)
./bin/ollama-distributed start \
  --config config/node.yaml \
  --listen 0.0.0.0:11435 \
  --p2p-listen /ip4/0.0.0.0/tcp/4002 \
  --data-dir /opt/ollama/data2 \
  --bootstrap /ip4/NODE1_IP/tcp/4001 \
  --enable-web

# Node 3 (Join cluster)
./bin/ollama-distributed start \
  --config config/node.yaml \
  --listen 0.0.0.0:11436 \
  --p2p-listen /ip4/0.0.0.0/tcp/4003 \
  --data-dir /opt/ollama/data3 \
  --bootstrap /ip4/NODE1_IP/tcp/4001 \
  --enable-web
```

## 🔧 Configuration

### Environment Variables

```bash
# API Configuration
OLLAMA_API_LISTEN=0.0.0.0:11434
OLLAMA_API_TIMEOUT=30s
OLLAMA_API_MAX_BODY_SIZE=33554432

# P2P Configuration
OLLAMA_P2P_LISTEN=/ip4/0.0.0.0/tcp/4001
OLLAMA_P2P_BOOTSTRAP=
OLLAMA_P2P_ENABLE_DHT=true
OLLAMA_P2P_ENABLE_PUBSUB=true

# Consensus Configuration
OLLAMA_CONSENSUS_BIND_ADDR=0.0.0.0:7000
OLLAMA_CONSENSUS_BOOTSTRAP=false
OLLAMA_CONSENSUS_DATA_DIR=./data/consensus

# Security Configuration
OLLAMA_SECURITY_TLS_ENABLED=true
OLLAMA_SECURITY_AUTH_ENABLED=true
OLLAMA_SECURITY_AUTH_METHOD=jwt

# Web Interface
OLLAMA_WEB_ENABLED=true
OLLAMA_WEB_LISTEN=0.0.0.0:8080

# Storage Configuration
OLLAMA_STORAGE_DATA_DIR=./data
OLLAMA_STORAGE_MODEL_DIR=./models
OLLAMA_STORAGE_CACHE_DIR=./cache
```

### Configuration File

See `config/node.yaml` for a complete configuration example.

## 🔐 Security Setup

### 1. TLS Certificates

```bash
# Generate CA certificate
openssl req -x509 -newkey rsa:2048 -keyout ca-key.pem -out ca-cert.pem -days 365 -nodes

# Generate server certificate
openssl req -newkey rsa:2048 -keyout server-key.pem -out server-csr.pem -nodes
openssl x509 -req -in server-csr.pem -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -days 365
```

### 2. JWT Authentication

```bash
# Generate JWT secret
openssl rand -hex 32

# Update configuration
export OLLAMA_SECURITY_AUTH_SECRET_KEY="your-secret-key-here"
```

### 3. RBAC Configuration

```yaml
# Add to node.yaml
security:
  auth:
    enabled: true
    method: jwt
    roles:
      - name: admin
        permissions: ["*"]
      - name: user
        permissions: ["read", "inference"]
      - name: readonly
        permissions: ["read"]
```

## 📊 Monitoring & Metrics

### 1. Prometheus Integration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'ollama-distributed'
    static_configs:
      - targets: ['localhost:9090', 'localhost:9091', 'localhost:9092']
```

### 2. Grafana Dashboard

Import the pre-built dashboard from `monitoring/grafana-dashboard.json`.

### 3. Health Checks

```bash
# Check node health
curl http://localhost:11434/api/v1/health

# Check cluster status
curl http://localhost:11434/api/v1/cluster/status

# Check metrics
curl http://localhost:9090/metrics
```

## 🧪 Testing

### 1. Unit Tests

```bash
make test-unit
```

### 2. Integration Tests

```bash
make test-integration
```

### 3. End-to-End Tests

```bash
make test-e2e
```

### 4. Performance Tests

```bash
make benchmark
```

## 🎯 Performance Tuning

### 1. Consensus Settings

```yaml
consensus:
  heartbeat_timeout: 1s
  election_timeout: 1s
  commit_timeout: 50ms
  max_append_entries: 64
  snapshot_interval: 120s
  snapshot_threshold: 8192
```

### 2. Scheduler Settings

```yaml
scheduler:
  algorithm: "round_robin"
  load_balancing: "least_connections"
  worker_count: 10
  queue_size: 10000
```

### 3. P2P Settings

```yaml
p2p:
  conn_mgr_low: 50
  conn_mgr_high: 200
  max_streams: 1000
  dial_timeout: 30s
```

## 🔍 Troubleshooting

### Common Issues

1. **Nodes not connecting**: Check firewall settings and P2P listen addresses
2. **Consensus failures**: Ensure odd number of nodes (3, 5, 7, etc.)
3. **Model sync issues**: Verify network connectivity and disk space
4. **High latency**: Tune consensus and scheduler settings

### Debug Mode

```bash
# Enable debug logging
export OLLAMA_LOGGING_LEVEL=debug

# Start with verbose output
./bin/ollama-distributed start --config config/node.yaml --log-level debug
```

### Log Analysis

```bash
# View logs
tail -f logs/ollama-distributed.log

# Search for errors
grep -i error logs/ollama-distributed.log

# Monitor consensus
grep -i consensus logs/ollama-distributed.log
```

## 📈 Scaling Guidelines

### Small Deployments (1-10 nodes)
- Use round-robin load balancing
- 3-node consensus cluster
- Basic monitoring

### Medium Deployments (10-100 nodes)
- Use least-connections load balancing
- 5-node consensus cluster
- Enhanced monitoring with Prometheus/Grafana

### Large Deployments (100+ nodes)
- Use adaptive load balancing
- 7-node consensus cluster
- Full monitoring stack with alerting

## 🛡️ Security Best Practices

1. **Network Security**
   - Use TLS for all communications
   - Implement proper firewall rules
   - Isolate consensus network

2. **Authentication**
   - Enable JWT authentication
   - Use strong secret keys
   - Implement proper RBAC

3. **Data Protection**
   - Encrypt data at rest
   - Use secure model checksums
   - Implement audit logging

4. **Operational Security**
   - Regular security updates
   - Monitor access logs
   - Implement intrusion detection

## 🔄 Backup & Recovery

### 1. Data Backup

```bash
# Backup consensus data
tar -czf consensus-backup.tar.gz data/consensus/

# Backup models
tar -czf models-backup.tar.gz models/

# Backup configuration
tar -czf config-backup.tar.gz config/
```

### 2. Disaster Recovery

```bash
# Stop node
./bin/ollama-distributed stop

# Restore data
tar -xzf consensus-backup.tar.gz
tar -xzf models-backup.tar.gz
tar -xzf config-backup.tar.gz

# Restart node
./bin/ollama-distributed start
```

## 📞 Support

- **Documentation**: https://github.com/ollama/ollama-distributed/wiki
- **Issues**: https://github.com/ollama/ollama-distributed/issues
- **Discussions**: https://github.com/ollama/ollama-distributed/discussions
- **Security**: security@ollama.com

## 🏆 Success Metrics

Your deployment is successful when:

- ✅ All nodes are connected and healthy
- ✅ Consensus is working (exactly one leader)
- ✅ Models are distributed across nodes
- ✅ Load balancing is working
- ✅ Web interface is accessible
- ✅ API endpoints are responding
- ✅ Monitoring is operational
- ✅ Security is properly configured
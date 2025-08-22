# 🚀 Getting Started with OllamaMax

Welcome to **OllamaMax** - the enterprise-grade distributed version of Ollama! This guide will help you get up and running quickly.

## 🎯 What is OllamaMax?

OllamaMax transforms the single-node Ollama architecture into a **horizontally scalable, fault-tolerant platform** that can:

- **Scale across multiple nodes** for high availability
- **Distribute AI models** automatically across your cluster
- **Load balance requests** intelligently
- **Monitor performance** in real-time
- **Secure your AI infrastructure** with enterprise-grade security

## ⚡ Quick Start (2 minutes)

### **Option 1: Quick Start (Recommended for beginners)**

```bash
# 1. Clone and build
git clone https://github.com/KhryptorGraphics/OllamaMax.git
cd OllamaMax/ollama-distributed
make build

# 2. Start with default configuration
./bin/ollama-distributed start --config config/node.yaml

# 3. Access the interfaces
# API: http://localhost:8080/api/v1
# Web UI: http://localhost:8081 (if enabled)
# Metrics: http://localhost:9090/metrics
```

### **Option 2: Docker Deployment (Recommended for production)**

```bash
# 1. Using Docker Compose
git clone https://github.com/KhryptorGraphics/OllamaMax.git
cd OllamaMax/ollama-distributed

# 2. Start the cluster
docker-compose up -d

# 3. Access your nodes
# Node 1: http://localhost:8080
# Node 2: http://localhost:8082  
# Node 3: http://localhost:8084
# Web UIs: 8081, 8083, 8085
```

## 🌐 Web Interface

Once your node is running, access the **beautiful web interface**:

### **Dashboard Features:**
- 📊 **Real-time metrics** and performance monitoring
- 🖥️ **Node management** with health status
- 🧠 **Model management** with automatic distribution
- 🔄 **Transfer monitoring** for model synchronization
- 🔒 **Security dashboard** with threat detection
- 📈 **Analytics** with performance insights

### **Quick Actions:**
```bash
# Access API directly
curl http://localhost:8080/api/v1/health

# View cluster status
curl http://localhost:8080/api/v1/cluster/status

# List available models
curl -H "Authorization: Bearer <JWT_TOKEN>" \
  http://localhost:8080/api/v1/models

# Monitor cluster metrics
curl http://localhost:9090/metrics
```

## 🤖 Using AI Models

### **Pull and Use Models:**

```bash
# Download a model (requires authentication)
curl -X POST -H "Authorization: Bearer <JWT_TOKEN>" \
  http://localhost:8080/api/v1/models/llama2:7b/download

# List available models
curl -H "Authorization: Bearer <JWT_TOKEN>" \
  http://localhost:8080/api/v1/models

# Generate text with a model
curl -X POST -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  http://localhost:8080/api/v1/generate \
  -d '{
    "model": "llama2:7b",
    "prompt": "Why is the sky blue?",
    "stream": false
  }'
```

### **Model Management:**
- **Automatic distribution** across cluster nodes
- **Intelligent load balancing** for optimal performance
- **Health monitoring** and automatic failover
- **Version management** and updates

## 🏗️ Building a Cluster

### **Single Node (Development):**
```bash
# Start a standalone node
NODE_ID=dev-node BOOTSTRAP=true \
./ollama-distributed start --config config/node.yaml
```

### **Multi-Node Cluster (Production):**

**Node 1 (Bootstrap):**
```bash
NODE_ID=node-1 BOOTSTRAP=true \
API_LISTEN=0.0.0.0:8080 P2P_LISTEN=/ip4/0.0.0.0/tcp/9000 \
./ollama-distributed start --config config/node.yaml
```

**Node 2 (Join cluster):**
```bash
NODE_ID=node-2 BOOTSTRAP=false \
API_LISTEN=0.0.0.0:8082 P2P_LISTEN=/ip4/0.0.0.0/tcp/9001 \
P2P_BOOTSTRAP_PEERS=/ip4/NODE1_IP/tcp/9000 \
./ollama-distributed start --config config/node.yaml
```

**Node 3+ (Additional nodes):**
```bash
NODE_ID=node-3 BOOTSTRAP=false \
API_LISTEN=0.0.0.0:8084 P2P_LISTEN=/ip4/0.0.0.0/tcp/9002 \
P2P_BOOTSTRAP_PEERS=/ip4/NODE1_IP/tcp/9000 \
./ollama-distributed start --config config/node.yaml
```

## 📊 Monitoring and Management

### **Real-time Monitoring:**
```bash
# Check cluster status
curl http://localhost:8080/api/v1/cluster/status

# View node health
curl http://localhost:8080/api/v1/nodes

# Monitor performance
curl http://localhost:8080/api/v1/metrics

# Access Prometheus metrics
curl http://localhost:9090/metrics

# WebSocket real-time updates
wscat -c ws://localhost:8080/ws?token=<JWT_TOKEN>
```

### **Web Dashboard:**
- **Performance metrics** with real-time charts
- **Resource utilization** monitoring
- **Error tracking** and alerting
- **Security monitoring** and compliance

## 🔒 Security Features

OllamaMax includes **enterprise-grade security**:

### **Built-in Security:**
- 🔐 **JWT authentication** for API access
- 🛡️ **TLS encryption** for all communications
- 🔍 **Security scanning** and vulnerability detection
- 📋 **Compliance monitoring** (CIS benchmarks)
- 🚨 **Real-time threat detection**

### **Security Configuration:**
```bash
# Set JWT secret (required for authentication)
export OLLAMA_JWT_SECRET="your-secure-secret-here"

# Enable TLS (recommended for production)
export OLLAMA_TLS_CERT_FILE="/etc/ssl/certs/ollama.crt"
export OLLAMA_TLS_KEY_FILE="/etc/ssl/private/ollama.key"

# Configure in node.yaml
security:
  auth:
    enabled: true
    method: "jwt"
    secret_key: "${OLLAMA_JWT_SECRET}"
  tls:
    enabled: true
    cert_file: "${OLLAMA_TLS_CERT_FILE}"
    key_file: "${OLLAMA_TLS_KEY_FILE}"
```

## 🎛️ Configuration

### **Basic Configuration (config.yaml):**
```yaml
node:
  id: "my-ollama-node"
  name: "ollama-node"
  region: "us-west-2"

api:
  listen: "0.0.0.0:8080"
  timeout: "30s"
  cors:
    enabled: true
    allowed_origins: ["*"]

web:
  enabled: true
  listen: "0.0.0.0:8081"

storage:
  data_dir: "./data"
  model_dir: "./models"
  cache_dir: "./cache"

security:
  auth:
    enabled: true
    method: "jwt"
    secret_key: "${OLLAMA_JWT_SECRET}"
```

### **Advanced Configuration:**
```yaml
p2p:
  listen_address: "/ip4/0.0.0.0/tcp/9000"
  bootstrap_peers: ["peer1:9000", "peer2:9000"]

consensus:
  algorithm: "raft"
  election_timeout: "5s"

scheduler:
  algorithm: "round_robin"
  health_check_interval: "30s"

performance:
  optimization_enabled: true
  monitoring_enabled: true
```

## 🛠️ Common Tasks

### **Model Management:**
```bash
# Download models (requires authentication)
curl -X POST -H "Authorization: Bearer <JWT_TOKEN>" \
  http://localhost:8080/api/v1/models/llama2:7b/download

curl -X POST -H "Authorization: Bearer <JWT_TOKEN>" \
  http://localhost:8080/api/v1/models/codellama:13b/download

# List models
curl -H "Authorization: Bearer <JWT_TOKEN>" \
  http://localhost:8080/api/v1/models

# Delete models
curl -X DELETE -H "Authorization: Bearer <JWT_TOKEN>" \
  http://localhost:8080/api/v1/models/llama2:7b

# Check model status with details
curl -H "Authorization: Bearer <JWT_TOKEN>" \
  http://localhost:8080/api/v1/models/llama2:7b
```

### **Cluster Management:**
```bash
# Check cluster health
curl http://localhost:8080/api/v1/cluster/status

# View cluster nodes
curl -H "Authorization: Bearer <JWT_TOKEN>" \
  http://localhost:8080/api/v1/nodes

# Monitor transfers
curl -H "Authorization: Bearer <JWT_TOKEN>" \
  http://localhost:8080/api/v1/transfers

# Get cluster leader
curl -H "Authorization: Bearer <JWT_TOKEN>" \
  http://localhost:8080/api/v1/cluster/leader
```

### **Performance Monitoring:**
```bash
# View system metrics
curl -H "Authorization: Bearer <JWT_TOKEN>" \
  http://localhost:8080/api/v1/metrics

# Check dashboard data
curl -H "Authorization: Bearer <JWT_TOKEN>" \
  http://localhost:8080/api/v1/dashboard/metrics

# Access Prometheus metrics (no auth required)
curl http://localhost:9090/metrics

# Access Grafana dashboards (if deployed)
open http://localhost:3000
```

## 🔧 Troubleshooting

### **Common Issues:**

**Node won't start:**
```bash
# Check configuration format
yaml-lint config/node.yaml

# Check logs with debug level
LOG_LEVEL=debug ./ollama-distributed start --config config/node.yaml

# Verify ports are available
netstat -tulpn | grep -E ':(8080|8081|9000|7000|9090)'

# Check if directories exist
ls -la ./data ./models ./cache
```

**Can't connect to cluster:**
```bash
# Check network connectivity
ping <bootstrap-peer-ip>
telnet <bootstrap-peer-ip> 9000

# Verify bootstrap peers configuration
grep -A5 "p2p:" config/node.yaml

# Check P2P connectivity
curl http://localhost:8080/api/v1/cluster/status
```

**Models not syncing:**
```bash
# Check transfer status
curl -H "Authorization: Bearer <JWT_TOKEN>" \
  http://localhost:8080/api/v1/transfers

# Download model manually
curl -X POST -H "Authorization: Bearer <JWT_TOKEN>" \
  http://localhost:8080/api/v1/models/<model-name>/download

# Check storage space
df -h ./models

# Check replication status
curl -H "Authorization: Bearer <JWT_TOKEN>" \
  http://localhost:8080/api/v1/nodes
```

### **Getting Help:**
```bash
# Check available commands
./ollama-distributed --help

# View configuration options
./ollama-distributed start --help

# Health check (no auth required)
curl http://localhost:8080/api/v1/health

# Version information (no auth required)
curl http://localhost:8080/api/v1/version
```

## 📚 Next Steps

### **Production Deployment:**
1. **Review security settings** in [SECURITY_HARDENING.md](SECURITY_HARDENING.md)
2. **Set up monitoring** with [PERFORMANCE_MONITORING.md](PERFORMANCE_MONITORING.md)
3. **Configure CI/CD** using [CI_CD_PIPELINE.md](CI_CD_PIPELINE.md)
4. **Scale your cluster** following [SCALING_GUIDE.md](docs/SCALING_GUIDE.md)

### **Advanced Features:**
- **Custom load balancing** algorithms
- **Model versioning** and rollback
- **Multi-region deployment**
- **Custom security policies**
- **Performance optimization** tuning

### **Integration:**
- **Kubernetes deployment** with Helm charts
- **Docker Compose** for development
- **Terraform** for infrastructure as code
- **Prometheus/Grafana** for monitoring

## 🎉 Success!

You now have a **production-ready, distributed AI infrastructure** running! 

### **What you've achieved:**
✅ **Scalable AI platform** that grows with your needs  
✅ **High availability** with automatic failover  
✅ **Enterprise security** with compliance monitoring  
✅ **Real-time monitoring** and performance optimization  
✅ **Beautiful web interface** for easy management  
✅ **API compatibility** with existing Ollama workflows  

### **Join the Community:**
- 📖 **Documentation**: [Full documentation](docs/)
- 🐛 **Issues**: [Report bugs](https://github.com/KhryptorGraphics/OllamaMax/issues)
- 💬 **Discussions**: [Community forum](https://github.com/KhryptorGraphics/OllamaMax/discussions)
- 🚀 **Contributions**: [Contributing guide](CONTRIBUTING.md)

**Happy distributed AI computing!** 🤖✨

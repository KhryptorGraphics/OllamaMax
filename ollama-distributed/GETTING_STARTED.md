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
# 1. Download and extract OllamaMax
wget https://github.com/KhryptorGraphics/OllamaMax/releases/latest/download/ollama-distributed-linux.tar.gz
tar -xzf ollama-distributed-linux.tar.gz
cd ollama-distributed

# 2. Quick start with defaults
./ollama-distributed quickstart

# 3. Start your node
./ollama-distributed start --config quickstart-config.yaml

# 4. Access the Web UI
open http://localhost:8081
```

### **Option 2: Interactive Setup (Recommended for production)**

```bash
# 1. Run the setup wizard
./ollama-distributed setup

# 2. Follow the interactive prompts
# 3. Start your configured node
./ollama-distributed start --config config.yaml

# 4. Access your node
open http://localhost:8081  # Web UI
curl http://localhost:8080/api/v1/proxy/status  # API
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
# Access Web UI
http://localhost:8081

# View cluster status
http://localhost:8081/dashboard

# Manage models
http://localhost:8081/models

# Monitor nodes
http://localhost:8081/nodes
```

## 🤖 Using AI Models

### **Pull and Use Models:**

```bash
# Pull a model through the distributed system
./ollama-distributed proxy pull llama2

# List available models
./ollama-distributed proxy list

# Run a model
curl -X POST http://localhost:8080/api/generate \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama2",
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
./ollama-distributed quickstart --name dev-node
./ollama-distributed start --config quickstart-config.yaml
```

### **Multi-Node Cluster (Production):**

**Node 1 (Bootstrap):**
```bash
./ollama-distributed setup
# Configure as new cluster
./ollama-distributed start --config config.yaml
```

**Node 2 (Join cluster):**
```bash
./ollama-distributed setup
# Configure to join existing cluster
# Enter Node 1's address as bootstrap peer
./ollama-distributed start --config config.yaml
```

**Node 3+ (Additional nodes):**
```bash
# Same process as Node 2
./ollama-distributed setup
./ollama-distributed start --config config.yaml
```

## 📊 Monitoring and Management

### **Real-time Monitoring:**
```bash
# Check cluster status
./ollama-distributed proxy status

# View node health
./ollama-distributed proxy instances

# Monitor performance
curl http://localhost:8080/api/v1/metrics

# Access Prometheus metrics
curl http://localhost:9090/metrics
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
# Enable security during setup
./ollama-distributed setup
# Choose "Enable security features" = yes

# Or configure manually in config.yaml
security:
  enabled: true
  jwt_secret: "your-secret-key"
  session_timeout: "24h"
```

## 🎛️ Configuration

### **Basic Configuration (config.yaml):**
```yaml
node:
  name: "my-ollama-node"
  data_dir: "./data"

api:
  listen_address: ":8080"
  enable_cors: true

web:
  enabled: true
  listen_address: ":8081"

models:
  storage_path: "./models"
  cache_size: "1GB"
  auto_pull: true

security:
  enabled: true
  jwt_secret: "auto-generated"
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
# Pull models
./ollama-distributed proxy pull llama2
./ollama-distributed proxy pull codellama

# List models
./ollama-distributed proxy list

# Remove models
./ollama-distributed proxy rm llama2

# Check model status
curl http://localhost:8080/api/v1/models
```

### **Cluster Management:**
```bash
# Check cluster health
./ollama-distributed proxy status

# View cluster nodes
curl http://localhost:8080/api/v1/nodes

# Monitor transfers
curl http://localhost:8080/api/v1/transfers
```

### **Performance Monitoring:**
```bash
# View performance metrics
curl http://localhost:8080/api/v1/analytics/performance

# Check resource usage
curl http://localhost:8080/api/v1/proxy/metrics

# Access Grafana dashboards
open http://localhost:3000
```

## 🔧 Troubleshooting

### **Common Issues:**

**Node won't start:**
```bash
# Check configuration
./ollama-distributed validate config.yaml

# Check logs
./ollama-distributed start --config config.yaml --log-level debug

# Verify ports are available
netstat -tulpn | grep :8080
```

**Can't connect to cluster:**
```bash
# Check network connectivity
ping <bootstrap-peer-ip>
telnet <bootstrap-peer-ip> 9000

# Verify bootstrap peers in config
cat config.yaml | grep bootstrap_peers
```

**Models not syncing:**
```bash
# Check transfer status
curl http://localhost:8080/api/v1/transfers

# Force model sync
./ollama-distributed proxy pull <model-name>

# Check storage space
df -h ./models
```

### **Getting Help:**
```bash
# Command help
./ollama-distributed help
./ollama-distributed proxy help

# Configuration validation
./ollama-distributed validate config.yaml

# Health check
curl http://localhost:8080/health
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

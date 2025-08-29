# Getting Started Tutorial

Welcome to the Ollama Distributed getting started tutorial! In this interactive tutorial, you'll learn how to install, configure, and use Ollama Distributed to create your first AI inference cluster.

## Learning Objectives

By the end of this tutorial, you will be able to:
- Install Ollama Distributed on your system
- Configure and start your first node
- Create a multi-node cluster
- Download and deploy AI models
- Perform inference requests
- Monitor cluster health and performance

## Prerequisites

- Basic command line knowledge
- Docker (recommended) or Go 1.19+ installed
- At least 8GB RAM and 4 CPU cores
- Network connectivity between nodes

## Tutorial Steps

### Step 1: Installation

Choose your preferred installation method:

#### Option A: Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/ollama/ollama-distributed.git
cd ollama-distributed

# Start with Docker Compose
docker-compose up -d
```

#### Option B: From Source

```bash
# Clone and build
git clone https://github.com/ollama/ollama-distributed.git
cd ollama-distributed
go build -o bin/ollama-distributed cmd/distributed-ollama/main.go
```

**✅ Checkpoint**: Verify installation
```bash
# Check version
./bin/ollama-distributed --version
# Expected output: ollama-distributed version 1.0.0
```

### Step 2: Start Your First Node

Let's start with a single-node cluster:

```bash
# Create basic configuration
mkdir -p config data

# Start the bootstrap node
./bin/ollama-distributed start --bootstrap --config config/node.yaml

# In another terminal, check status
./bin/ollama-distributed status
```

**Expected Output:**
```json
{
  "node_id": "node-12345",
  "status": "healthy",
  "is_leader": true,
  "cluster_size": 1,
  "uptime": "running"
}
```

**✅ Checkpoint**: Your node should show as "healthy" and "is_leader": true

### Step 3: Access the Web Interface

Open your web browser and navigate to:
- **URL**: http://localhost:8080
- **Default credentials**: No authentication required for local setup

You should see the Ollama Distributed dashboard showing:
- ✅ 1 healthy node
- 📊 System metrics (CPU, memory, network)
- 🎯 No models loaded yet

### Step 4: Download Your First Model

Let's download the Llama2 model:

```bash
# Download Llama2 (7B parameters, ~4GB)
./bin/ollama-distributed model pull llama2

# Monitor download progress
./bin/ollama-distributed model status llama2
```

**Expected Progress:**
```
Downloading llama2: 45% (1.8GB/4.0GB) - ETA: 2m30s
```

**✅ Checkpoint**: Model status should show "available" when complete

### Step 5: Your First Inference Request

Now let's generate some text:

```bash
# Simple text generation
curl -X POST http://localhost:8080/api/v1/generate \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama2",
    "prompt": "Explain artificial intelligence in simple terms:",
    "stream": false
  }'
```

**Expected Response:**
```json
{
  "response": "Artificial intelligence (AI) is like teaching computers to think and make decisions similar to how humans do. It involves creating programs that can learn from data, recognize patterns, and solve problems without being explicitly programmed for every specific task...",
  "model": "llama2",
  "node_id": "node-12345"
}
```

**✅ Checkpoint**: You should receive a coherent text response

### Step 6: Create a Multi-Node Cluster

Let's add a second node to create a proper cluster:

```bash
# In a new terminal, start second node
./bin/ollama-distributed start --peers localhost:8080 --port 8081 --data-dir data2

# Check cluster status
./bin/ollama-distributed cluster status
```

**Expected Output:**
```json
{
  "leader": "node-12345",
  "nodes": 2,
  "cluster_health": "healthy",
  "consensus_state": "stable"
}
```

**✅ Checkpoint**: Cluster should show 2 nodes

### Step 7: Monitor Cluster Performance

Use the monitoring commands to observe your cluster:

```bash
# Real-time metrics
./bin/ollama-distributed metrics --watch

# Node performance
./bin/ollama-distributed proxy metrics --json

# Model distribution
./bin/ollama-distributed model list --distribution
```

**Key Metrics to Watch:**
- **CPU Usage**: Should be <50% during idle
- **Memory Usage**: Model loading will increase this
- **Request Latency**: Should be <200ms for simple requests
- **Error Rate**: Should be 0%

### Step 8: Test High Availability

Let's test fault tolerance:

```bash
# Stop one node (simulate failure)
docker stop ollama-distributed-node-1

# Verify cluster continues working
curl -X POST http://localhost:8080/api/v1/generate \
  -H "Content-Type: application/json" \
  -d '{"model": "llama2", "prompt": "Test message"}'

# Check cluster adapts
./bin/ollama-distributed cluster status
```

**Expected Behavior:**
- Requests should still work (routed to healthy node)
- Cluster status shows 1 healthy node
- Leader election occurs if leader was stopped

**✅ Checkpoint**: System remains functional with one node down

## Troubleshooting Common Issues

### Issue: Node Won't Start
**Symptoms**: Error messages about ports or permissions

**Solutions**:
```bash
# Check if port is in use
netstat -tuln | grep 8080

# Use different port
./bin/ollama-distributed start --port 8081

# Check permissions
chmod +x bin/ollama-distributed
```

### Issue: Model Download Fails
**Symptoms**: Download stalls or fails

**Solutions**:
```bash
# Check internet connectivity
curl -I https://ollama.ai

# Clear cache and retry
./bin/ollama-distributed model cleanup --cache
./bin/ollama-distributed model pull llama2 --retry 3
```

### Issue: High Memory Usage
**Symptoms**: System becomes slow, out of memory errors

**Solutions**:
```bash
# Check memory usage
./bin/ollama-distributed status --memory

# Unload unused models
./bin/ollama-distributed model unload mistral

# Increase swap space (Linux)
sudo fallocate -l 4G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
```

## Next Steps

Congratulations! You've successfully:
- ✅ Installed Ollama Distributed
- ✅ Created a multi-node cluster
- ✅ Downloaded and used an AI model
- ✅ Tested high availability
- ✅ Monitored performance

### Continue Learning

1. **Scaling**: [Learn advanced scaling techniques](./scaling-tutorial.md)
2. **Production**: [Deploy to production environments](../operations-guide/production-deployment.md)
3. **API Integration**: [Build applications with our APIs](../developer-guide/api-integration.md)
4. **Monitoring**: [Set up comprehensive monitoring](./monitoring-tutorial.md)

### Advanced Features to Explore

- **Multi-Model Deployment**: Run multiple models simultaneously
- **Custom Load Balancing**: Configure advanced routing strategies
- **Security**: Enable authentication and authorization
- **Plugins**: Develop custom functionality extensions

## Tutorial Completion

**🎉 Congratulations!** You've completed the Getting Started tutorial.

**Time Spent**: ~30-45 minutes
**Skills Gained**: Basic cluster setup, model management, monitoring
**Ready For**: Intermediate tutorials and production deployment

### Knowledge Check

Test your understanding:

1. What command starts a bootstrap node?
2. How do you check cluster health?
3. What's the difference between a leader and follower node?
4. How does the system handle node failures?

**Answers**:
1. `./bin/ollama-distributed start --bootstrap`
2. `./bin/ollama-distributed cluster status`
3. Leader coordinates cluster operations; followers replicate and serve requests
4. Automatic failover with leader re-election in <30 seconds

### Get Help

- 📖 [Full Documentation](../README.md)
- 💬 [Discord Community](https://discord.gg/ollama)
- 🐛 [Report Issues](https://github.com/ollama/ollama-distributed/issues)
- 🎓 [Advanced Training](../training/README.md)

---

**What's Next?** Try the [Scaling Tutorial](./scaling-tutorial.md) to learn how to scale your cluster to hundreds of nodes!
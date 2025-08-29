# Getting Started with Ollama Distributed

Welcome to Ollama Distributed! This guide will help you understand and get started with this distributed AI platform based on the actual implementation.

## What is Ollama Distributed?

Ollama Distributed is an ambitious distributed systems project that extends Ollama with clustering capabilities. Currently in active development, it provides:

**✅ Working Components:**
- **Professional CLI Interface**: Complete command-line tool with interactive setup
- **P2P Networking**: Real libp2p-based mesh networking between nodes
- **Distributed Architecture**: Sophisticated system design with schedulers and consensus
- **Configuration Management**: Comprehensive YAML-based configuration system  
- **API Compatibility**: Ollama-compatible HTTP endpoints (with placeholder responses)
- **Web Dashboard Framework**: Basic web interface structure

**🚧 In Development:**
- **Full Model Distribution**: P2P model synchronization (framework exists, integration incomplete)
- **Distributed Inference**: Model partitioning and distributed execution (placeholder responses)
- **Database Integration**: PostgreSQL/SQLite backend (package missing)
- **Production Monitoring**: Complete Prometheus/Grafana integration

## Current Capabilities

### What You Can Do Today

1. **✅ Install and Configure**: Use the professional CLI to set up nodes
2. **✅ Network Nodes**: Connect multiple nodes via P2P networking  
3. **✅ Monitor System**: Check cluster status and health
4. **✅ API Access**: Make API calls (returns placeholder responses)
5. **✅ Test Architecture**: Explore the distributed system design

### What's Coming Soon

1. **🔄 Model Sync**: P2P model distribution and replication
2. **🔄 Real Inference**: Distributed AI model execution
3. **🔄 Web Dashboard**: Complete React-based interface
4. **🔄 Production Features**: Full monitoring and management

## Quick Start (Development/Testing)

Get started with the current implementation:

```bash
# Build the project (if from source)
cd ollama-distributed
go build -o bin/ollama-distributed ./cmd/ollama-distributed

# Or use existing binary
./bin/ollama-distributed quickstart

# Your development cluster is now running
```

**Current Output:**
- API Server: http://localhost:8080 (placeholder responses)
- Web Interface: http://localhost:8081 (basic interface)
- Health Check: http://localhost:8080/health (working)

That's it! You now have the distributed framework running and can explore its architecture.

## Learning Objectives

By the end of this guide, you will:

1. ✅ Install Ollama Distributed on your system
2. ✅ Configure your first node
3. ✅ Create a basic cluster
4. ✅ Deploy your first AI model
5. ✅ Perform inference requests

## System Requirements

### Minimum Requirements
- **OS**: Linux, macOS, or Windows with WSL2
- **CPU**: 2 cores
- **RAM**: 4GB available
- **Storage**: 20GB free space
- **Network**: Internet connection for model downloads

### Recommended for Production
- **OS**: Linux (Ubuntu 20.04+, CentOS 8+)
- **CPU**: 8+ cores
- **RAM**: 16GB+ (32GB for large models)
- **Storage**: 100GB+ SSD
- **GPU**: NVIDIA GPU with CUDA support (optional)

## Installation Methods

### Method 1: Quick Install (Recommended)

This is the easiest way to get started:

```bash
# Download and run installer
curl -fsSL https://install.ollamamax.com/install.sh | bash

# Add to PATH (if not already done)
export PATH=$PATH:/usr/local/bin

# Verify installation
ollama-distributed --version
```

### Method 2: Manual Installation

For advanced users who want more control:

```bash
# Download the binary
wget https://github.com/ollamamax/releases/latest/download/ollama-distributed-linux-amd64.tar.gz

# Extract
tar -xzf ollama-distributed-linux-amd64.tar.gz

# Make executable and move to PATH
chmod +x ollama-distributed
sudo mv ollama-distributed /usr/local/bin/
```

### Method 3: Build from Source

```bash
# Clone the repository
git clone https://github.com/ollamamax/ollama-distributed.git
cd ollama-distributed

# Build
go build -o bin/ollama-distributed ./cmd/ollama-distributed

# Install
sudo cp bin/ollama-distributed /usr/local/bin/
```

## Configuration

### Option 1: Interactive Setup Wizard

The setup wizard will guide you through configuration:

```bash
ollama-distributed setup
```

This will create a configuration file at `~/.ollama-distributed/config.yaml` with optimal settings for your environment.

### Option 2: Auto-configuration

Let Ollama Distributed detect and configure your environment automatically:

```bash
ollama-distributed quickstart --auto-config
```

### Option 3: Manual Configuration

Create a configuration file manually:

```yaml
# ~/.ollama-distributed/config.yaml
server:
  host: "0.0.0.0"
  port: 8081
  
cluster:
  node_id: "node-1"
  bootstrap: true
  
models:
  store_path: "./models"
  
inference:
  max_concurrent: 10
```

## Creating Your First Cluster

### Single Node Cluster (Development)

Perfect for development and testing:

```bash
# Initialize and start
ollama-distributed quickstart

# Check status
ollama-distributed status
```

### Multi-Node Cluster (Production)

For production deployments:

**Node 1 (Bootstrap):**
```bash
# Start the first node
ollama-distributed start --bootstrap --node-id node-1 --port 8081

# Get the cluster join token
ollama-distributed cluster info --token
```

**Node 2 (Join):**
```bash
# Join the cluster
ollama-distributed start --join <join-token> --node-id node-2 --port 8082
```

**Node 3 (Join):**
```bash
# Add third node for high availability
ollama-distributed start --join <join-token> --node-id node-3 --port 8083
```

## Deploying Your First AI Model

### Method 1: Using the CLI

```bash
# Pull a model
ollama-distributed models pull llama2:7b

# List available models
ollama-distributed models list

# Get model info
ollama-distributed models show llama2:7b
```

### Method 2: Using the API

```bash
# Pull model via API
curl -X POST http://localhost:8081/api/pull \
  -H "Content-Type: application/json" \
  -d '{"name": "llama2:7b"}'
```

### Method 3: Import from Local Ollama

If you already have Ollama installed:

```bash
# Migrate existing models
ollama-distributed migrate --from-ollama --preserve-models

# Or sync specific models
ollama-distributed models sync llama2:7b
```

## Performing Inference Requests

### Using the CLI

```bash
# Interactive chat
ollama-distributed chat llama2:7b

# Single generation
ollama-distributed generate llama2:7b "Explain machine learning in simple terms"

# With streaming
ollama-distributed generate llama2:7b "Write a story" --stream
```

### Using the REST API

```bash
# Generate completion
curl -X POST http://localhost:8081/api/generate \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama2:7b",
    "prompt": "What is artificial intelligence?",
    "stream": false
  }'

# Streaming response
curl -X POST http://localhost:8081/api/generate \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama2:7b",
    "prompt": "Tell me a story",
    "stream": true
  }'
```

### Using SDKs

#### Python
```python
from ollama_distributed import Client

client = Client('http://localhost:8081')
response = client.generate(
    model='llama2:7b',
    prompt='Explain quantum computing'
)
print(response.text)
```

#### JavaScript
```javascript
import { OllamaClient } from 'ollama-distributed-js';

const client = new OllamaClient({ baseUrl: 'http://localhost:8081' });
const response = await client.generate({
  model: 'llama2:7b',
  prompt: 'What is blockchain?'
});
console.log(response.text);
```

## Monitoring Your Cluster

### Web Dashboard

Visit http://localhost:8081/dashboard to see:

- Real-time cluster status
- Node health and performance
- Model distribution
- Request metrics
- Resource utilization

### CLI Monitoring

```bash
# Cluster overview
ollama-distributed status

# Detailed node information
ollama-distributed cluster nodes

# Performance metrics
ollama-distributed metrics

# Live log monitoring
ollama-distributed logs --follow
```

### Metrics API

```bash
# Prometheus metrics
curl http://localhost:8081/metrics

# Health check
curl http://localhost:8081/health

# Cluster information
curl http://localhost:8081/api/cluster/status
```

## Next Steps

Congratulations! You now have a working Ollama Distributed cluster. Here's what to explore next:

1. **[Configuration Guide](./configuration.md)** - Advanced configuration options
2. **[API Reference](./api/overview.md)** - Complete API documentation
3. **[Deployment Guide](./deployment/overview.md)** - Production deployment strategies
4. **[Monitoring Guide](./monitoring.md)** - Set up comprehensive monitoring
5. **[Security Guide](./security.md)** - Secure your cluster

## Common Issues and Solutions

### Port Already in Use
```bash
# Check what's using the port
sudo lsof -i :8081

# Use a different port
ollama-distributed start --port 8082
```

### Model Download Issues
```bash
# Check connectivity
ollama-distributed validate --connectivity

# Use different model registry
ollama-distributed config set models.registry "https://huggingface.co"
```

### Memory Issues
```bash
# Check system resources
ollama-distributed validate --resources

# Reduce concurrent requests
ollama-distributed config set inference.max_concurrent 5
```

### Cluster Formation Problems
```bash
# Run diagnostics
ollama-distributed troubleshoot --cluster

# Reset cluster state
ollama-distributed cluster reset --confirm
```

## Getting Help

- **Documentation**: [https://docs.ollamamax.com](https://docs.ollamamax.com)
- **Community Forum**: [https://community.ollamamax.com](https://community.ollamamax.com)
- **GitHub Issues**: [https://github.com/ollamamax/ollama-distributed/issues](https://github.com/ollamamax/ollama-distributed/issues)
- **Discord**: [https://discord.gg/ollamamax](https://discord.gg/ollamamax)

## Examples and Tutorials

### Basic Examples
- [Simple Chat Application](./examples/chat-app.md)
- [REST API Integration](./examples/api-integration.md)
- [Batch Processing](./examples/batch-processing.md)

### Advanced Examples
- [Multi-Model Deployment](./examples/multi-model.md)
- [GPU Acceleration](./examples/gpu-acceleration.md)
- [Kubernetes Deployment](./examples/kubernetes.md)

---

**Ready to scale your AI workloads?** Ollama Distributed makes it simple to go from prototype to production with enterprise-grade distributed AI infrastructure.
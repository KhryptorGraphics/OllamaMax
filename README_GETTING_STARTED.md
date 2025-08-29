# 🚀 OllamaMax Getting Started Guide

Welcome to OllamaMax! This guide will get you up and running in just a few minutes.

## Quick Start (60 seconds)

For the fastest setup experience, use our quickstart command:

```bash
# Install OllamaMax
curl -fsSL https://raw.githubusercontent.com/KhryptorGraphics/OllamaMax/main/scripts/install.sh | bash

# Start immediately with defaults
ollama-distributed quickstart
```

That's it! 🎉 Your OllamaMax node is now running with:
- **Web Dashboard**: http://localhost:8081
- **API Endpoint**: http://localhost:8080
- **Health Check**: http://localhost:8080/health

## Installation Options

### 1. Automatic Installation (Recommended)

```bash
# Universal installer (Linux, macOS, Windows WSL)
curl -fsSL https://install.ollamamax.com | bash

# Or with options
curl -fsSL https://install.ollamamax.com | bash -s -- --enable-gpu --quick
```

### 2. Manual Installation

```bash
# Download and run installer
wget https://github.com/KhryptorGraphics/OllamaMax/raw/main/scripts/install.sh
chmod +x install.sh
./install.sh --help
```

### 3. From Source

```bash
git clone https://github.com/KhryptorGraphics/OllamaMax.git
cd OllamaMax
go build -o ollama-distributed ./ollama-distributed/cmd/node
```

## Getting Started Workflows

### 🏃‍♂️ Speed Runner (1 minute)
```bash
ollama-distributed quickstart
# ✅ Node running, web UI open, ready to go!
```

### 🛠️ Custom Setup (5 minutes)
```bash
ollama-distributed setup
# Interactive wizard guides you through configuration
# Choose ports, enable features, configure security
```

### 🔧 Advanced Setup (15 minutes)
```bash
# Generate custom configuration
./scripts/config-generator.sh --profile production --security

# Validate configuration
ollama-distributed validate --config production-config.yaml

# Start with custom config
ollama-distributed start --config production-config.yaml
```

## Your First AI Conversation

After starting OllamaMax, try these commands:

```bash
# Download a model
ollama-distributed proxy pull phi3:mini

# Chat via API
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "model": "phi3:mini",
    "messages": [{"role": "user", "content": "Hello! How are you?"}]
  }'

# Or use the beautiful web interface
open http://localhost:8081
```

## Essential Commands

| Command | Purpose | Example |
|---------|---------|---------|
| `quickstart` | Instant setup with defaults | `ollama-distributed quickstart` |
| `setup` | Interactive configuration wizard | `ollama-distributed setup` |
| `start` | Start the OllamaMax node | `ollama-distributed start` |
| `status` | Check system health | `ollama-distributed status --verbose` |
| `validate` | Validate configuration | `ollama-distributed validate --fix` |
| `proxy pull` | Download AI models | `ollama-distributed proxy pull llama2:7b` |
| `examples` | See usage examples | `ollama-distributed examples` |
| `tutorial` | Interactive tutorial | `ollama-distributed tutorial` |
| `troubleshoot` | Fix common issues | `ollama-distributed troubleshoot` |

## Configuration Profiles

Choose the right profile for your use case:

### 🔧 Development
```bash
./scripts/config-generator.sh --profile development
# • Debug logging, minimal security
# • Perfect for testing and development
# • Single node, low resource usage
```

### 🏭 Production  
```bash
./scripts/config-generator.sh --profile production --security
# • Optimized performance, security enabled
# • Structured logging, resource limits
# • Production-ready configuration
```

### 🌐 Cluster
```bash
./scripts/config-generator.sh --profile cluster --gpu
# • Multi-node distributed setup  
# • P2P networking, automatic failover
# • High availability and scalability
```

### ⚡ GPU-Accelerated
```bash
./scripts/config-generator.sh --profile gpu --enable-gpu
# • GPU acceleration enabled
# • Optimized for AI/ML workloads
# • Higher performance limits
```

## Integration with Existing Ollama

Already have Ollama installed? We've got you covered:

```bash
# Scan for existing installations
./scripts/ollama-integration.sh --scan

# Migrate your existing models and data
./scripts/ollama-integration.sh --mode migrate

# Run alongside existing Ollama (different ports)
./scripts/ollama-integration.sh --mode coexist
```

## Web Interface Features

Open http://localhost:8081 to access:

- 💬 **Chat Interface** - Interactive AI conversations
- 🤖 **Model Manager** - Download, update, and manage models  
- 📊 **Performance Dashboard** - Real-time metrics and monitoring
- ⚙️ **Configuration** - Visual configuration management
- 🔧 **System Tools** - Diagnostics and maintenance
- 📚 **API Explorer** - Test and explore REST APIs

## Troubleshooting

### Quick Fixes

```bash
# Check if everything is working
ollama-distributed status

# Fix common configuration issues  
ollama-distributed validate --fix

# Reset to defaults if needed
ollama-distributed quickstart --force

# Get diagnostic information
ollama-distributed troubleshoot
```

### Common Issues

#### Port Already in Use
```bash
# Use different ports
ollama-distributed setup
# Choose custom ports during setup

# Or specify ports directly  
ollama-distributed start --api-port 8090 --web-port 8091
```

#### Models Not Loading
```bash
# Check model directory
ollama-distributed status --verbose

# Re-download models
ollama-distributed proxy pull phi3:mini --force

# Check disk space
df -h
```

#### Permission Errors
```bash
# Fix data directory permissions
sudo chown -R $USER:$USER ~/.ollamamax

# Or use custom directory
ollama-distributed setup --data-dir ~/my-ollama-data
```

## Next Steps

### 📚 Learn More
- [Complete Tutorial](https://docs.ollamamax.com/tutorial) - Step-by-step learning
- [API Documentation](https://docs.ollamamax.com/api) - REST API reference  
- [Configuration Guide](https://docs.ollamamax.com/config) - Advanced configuration
- [Deployment Guide](https://docs.ollamamax.com/deploy) - Production deployment

### 🚀 Scale Up
- [Cluster Setup](https://docs.ollamamax.com/cluster) - Multi-node deployment
- [Performance Tuning](https://docs.ollamamax.com/performance) - Optimization guide
- [Monitoring](https://docs.ollamamax.com/monitoring) - Metrics and alerting
- [Security](https://docs.ollamamax.com/security) - Production security

### 🤝 Community
- [GitHub](https://github.com/KhryptorGraphics/OllamaMax) - Source code and issues
- [Discussions](https://github.com/KhryptorGraphics/OllamaMax/discussions) - Community forum
- [Discord](https://discord.gg/ollamamax) - Real-time chat
- [Examples](https://github.com/KhryptorGraphics/OllamaMax/examples) - Sample applications

## Support

Need help? We're here for you:

- 📖 **Documentation**: https://docs.ollamamax.com
- 🐛 **Bug Reports**: https://github.com/KhryptorGraphics/OllamaMax/issues
- 💬 **Community**: https://github.com/KhryptorGraphics/OllamaMax/discussions
- 📧 **Email**: support@ollamamax.com

---

**Welcome to the future of distributed AI!** 🌟

Ready to get started? Run: `ollama-distributed quickstart`
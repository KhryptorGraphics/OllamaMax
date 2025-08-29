# User Guide - Ollama Distributed

Welcome to the comprehensive user guide for Ollama Distributed, a powerful platform for distributed AI model inference across multiple nodes.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Basic Usage](#basic-usage)
3. [Web Interface](#web-interface)
4. [Model Management](#model-management)
5. [Troubleshooting](#troubleshooting)
6. [Best Practices](#best-practices)

## Getting Started

### Prerequisites

- **Hardware Requirements**:
  - Minimum: 4 CPU cores, 8GB RAM per node
  - Recommended: 8+ CPU cores, 32GB+ RAM per node
  - Storage: 100GB+ available space for models
  - Network: Gigabit Ethernet recommended

- **Software Requirements**:
  - Docker (recommended) or Go 1.19+
  - Operating System: Linux, macOS, or Windows
  - Network: Open ports 8080 (HTTP) and 8443 (HTTPS)

### Quick Start

#### Option 1: Using Docker (Recommended)

1. **Clone and Start**:
   ```bash
   git clone https://github.com/ollama/ollama-distributed.git
   cd ollama-distributed
   docker-compose up -d
   ```

2. **Access the Web Interface**:
   Open your browser to [http://localhost:8080](http://localhost:8080)

3. **Verify Installation**:
   ```bash
   curl http://localhost:8080/api/v1/health
   ```

#### Option 2: Native Installation

1. **Build from Source**:
   ```bash
   git clone https://github.com/ollama/ollama-distributed.git
   cd ollama-distributed
   go build -o bin/ollama-distributed cmd/distributed-ollama/main.go
   ```

2. **Start the Node**:
   ```bash
   ./bin/ollama-distributed start --config config/node.yaml
   ```

3. **Check Status**:
   ```bash
   ./bin/ollama-distributed status
   ```

## Basic Usage

### Starting Your First Cluster

1. **Initialize the First Node** (Bootstrap):
   ```bash
   # Start the bootstrap node
   ./ollama-distributed start --bootstrap
   
   # Check node status
   ./ollama-distributed status
   ```

2. **Add Additional Nodes**:
   ```bash
   # Join existing cluster (run on new nodes)
   ./ollama-distributed start --peers node1:8080,node2:8080
   
   # Verify cluster membership
   ./ollama-distributed cluster status
   ```

### Basic Commands

#### Node Management
```bash
# Start a node
./ollama-distributed start [options]

# Check node status
./ollama-distributed status

# Stop a node gracefully
./ollama-distributed stop

# View cluster information
./ollama-distributed cluster status
```

#### Model Operations
```bash
# Download a model
./ollama-distributed model pull llama2

# List available models
./ollama-distributed model list

# Remove a model
./ollama-distributed model rm llama2

# Check model status across cluster
./ollama-distributed model status llama2
```

#### Proxy Management
```bash
# Check proxy status
./ollama-distributed proxy status

# List all instances
./ollama-distributed proxy instances

# View performance metrics
./ollama-distributed proxy metrics

# Real-time monitoring
./ollama-distributed proxy metrics --watch
```

## Web Interface

### Dashboard Overview

The web interface provides a comprehensive control panel for managing your distributed Ollama cluster:

**Key Features:**
- **Real-time Cluster Status**: Live view of all nodes, their health, and performance
- **Model Management**: Visual interface for downloading, distributing, and managing models
- **Performance Monitoring**: Charts and graphs showing system metrics
- **Log Viewer**: Centralized log viewing with filtering capabilities
- **Settings Panel**: Configuration management interface

### Main Dashboard Sections

#### 1. Cluster Overview
- **Node Status Grid**: Visual representation of all cluster nodes
- **System Health**: Overall cluster health indicators
- **Active Models**: Currently loaded and available models
- **Request Statistics**: Real-time inference request metrics

#### 2. Model Management
- **Model Library**: Browse and search available models
- **Download Manager**: Monitor model download progress
- **Distribution Status**: See where models are replicated across nodes
- **Performance Metrics**: Per-model usage statistics

#### 3. Monitoring & Alerts
- **Performance Charts**: CPU, memory, network usage over time
- **Alert Center**: System alerts and notifications
- **Log Viewer**: Searchable, filterable log interface
- **Health Checks**: Automated system health monitoring

#### 4. Configuration
- **Node Settings**: Per-node configuration management
- **Cluster Settings**: Cluster-wide configuration options
- **Security Settings**: Authentication and access control
- **Backup & Recovery**: Data backup and restore options

### Navigation Tips

- **Dark/Light Mode**: Toggle in the top-right corner
- **Responsive Design**: Optimized for desktop, tablet, and mobile
- **Keyboard Shortcuts**: 
  - `Ctrl+K` / `Cmd+K`: Quick search
  - `Ctrl+/` / `Cmd+/`: Show help overlay
  - `Esc`: Close dialogs/modals
- **Real-time Updates**: Most data refreshes automatically every 5-10 seconds

## Model Management

### Downloading Models

#### Via Command Line
```bash
# Download popular models
./ollama-distributed model pull llama2
./ollama-distributed model pull mistral
./ollama-distributed model pull codellama

# Download specific model versions
./ollama-distributed model pull llama2:13b
./ollama-distributed model pull mistral:7b-instruct
```

#### Via Web Interface
1. Navigate to **Models** → **Library**
2. Browse or search for models
3. Click **Download** on desired model
4. Monitor progress in **Downloads** tab
5. Model becomes available when download completes

### Model Distribution

Models are automatically distributed across your cluster based on:

- **Available Storage**: Nodes with sufficient disk space
- **Resource Capacity**: CPU and memory availability  
- **Geographic Location**: Network proximity for optimal performance
- **Replication Factor**: Configurable redundancy level

#### Manual Distribution
```bash
# Force model distribution to specific nodes
./ollama-distributed model distribute llama2 --nodes node1,node2,node3

# Set replication factor
./ollama-distributed model configure llama2 --replicas 3

# Check distribution status
./ollama-distributed model status llama2 --verbose
```

### Model Usage

#### Text Generation
```bash
# Simple text generation
curl -X POST http://localhost:8080/api/v1/generate \
  -H "Content-Type: application/json" \
  -d '{"model": "llama2", "prompt": "Explain quantum computing"}'

# Streaming response
curl -X POST http://localhost:8080/api/v1/generate \
  -H "Content-Type: application/json" \
  -d '{"model": "llama2", "prompt": "Write a story", "stream": true}'
```

#### Chat Interface
```bash
# Chat completion
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama2",
    "messages": [
      {"role": "user", "content": "Hello, how are you?"}
    ]
  }'
```

#### Embeddings
```bash
# Generate embeddings
curl -X POST http://localhost:8080/api/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{"model": "llama2", "input": "Hello world"}'
```

## Troubleshooting

### Common Issues

#### 1. Node Connection Problems

**Symptoms**: Nodes can't join cluster or frequent disconnections

**Solutions**:
```bash
# Check network connectivity
ping <node-ip>
telnet <node-ip> 8080

# Verify firewall settings
sudo ufw status
sudo iptables -L

# Check node configuration
./ollama-distributed config validate

# Restart networking
./ollama-distributed restart --network-reset
```

#### 2. Model Download Failures

**Symptoms**: Models fail to download or get stuck

**Solutions**:
```bash
# Check disk space
df -h

# Verify internet connection
curl -I https://ollama.ai

# Clear download cache
./ollama-distributed model cleanup --cache

# Retry with verbose logging
./ollama-distributed model pull llama2 --verbose --retry 3
```

#### 3. High Memory Usage

**Symptoms**: Nodes running out of memory

**Solutions**:
```bash
# Check memory usage
./ollama-distributed status --memory

# Configure model limits
./ollama-distributed config set --max-models-per-node 2

# Enable memory monitoring
./ollama-distributed monitor --memory-alerts

# Adjust garbage collection
./ollama-distributed config set --gc-threshold 80
```

#### 4. Performance Issues

**Symptoms**: Slow inference or high latency

**Solutions**:
```bash
# Check system resources
./ollama-distributed metrics --detailed

# Optimize load balancing
./ollama-distributed config set --load-balance-algorithm round-robin

# Scale cluster horizontally
./ollama-distributed scale add-nodes --count 2

# Enable request caching
./ollama-distributed config set --enable-cache true
```

### Log Analysis

#### Accessing Logs
```bash
# View system logs
./ollama-distributed logs --follow

# Filter by log level
./ollama-distributed logs --level error

# Search logs
./ollama-distributed logs --search "connection failed"

# Export logs
./ollama-distributed logs --export /tmp/logs.json
```

#### Common Log Patterns

**Connection Issues**:
```
ERROR: Failed to connect to peer node-002: connection timeout
```
**Solution**: Check network connectivity and firewall rules

**Memory Warnings**:
```
WARN: Memory usage 95%, consider scaling or reducing model load
```
**Solution**: Add more nodes or optimize model distribution

**Model Errors**:
```
ERROR: Model llama2 failed to load: insufficient GPU memory
```
**Solution**: Move model to node with more resources

### Getting Help

#### Built-in Help System
```bash
# General help
./ollama-distributed help

# Command-specific help
./ollama-distributed help model
./ollama-distributed help cluster

# Configuration help
./ollama-distributed config help
```

#### Debug Mode
```bash
# Enable detailed debugging
./ollama-distributed --debug start

# Debug specific components
./ollama-distributed --debug-p2p --debug-consensus start

# Generate debug report
./ollama-distributed debug report --output debug-report.json
```

#### Community Support
- **GitHub Issues**: Report bugs and feature requests
- **Documentation**: Comprehensive guides and API reference
- **Community Forum**: Ask questions and share experiences
- **Discord**: Real-time chat and support

## Best Practices

### Cluster Planning

#### 1. Node Sizing
- **Small Cluster (3-5 nodes)**: 8GB RAM, 4 cores per node
- **Medium Cluster (5-20 nodes)**: 16GB RAM, 8 cores per node  
- **Large Cluster (20+ nodes)**: 32GB+ RAM, 16+ cores per node

#### 2. Network Design
- **Dedicated Network**: Use dedicated VLAN for cluster communication
- **Low Latency**: Keep nodes geographically close (<10ms latency)
- **High Bandwidth**: Gigabit or 10Gb connections recommended
- **Redundancy**: Multiple network paths for fault tolerance

#### 3. Storage Strategy
- **SSD Storage**: Use SSDs for model storage and caching
- **Distributed Storage**: Replicate models across multiple nodes
- **Backup Strategy**: Regular backups of critical data and configurations

### Performance Optimization

#### 1. Model Distribution
```bash
# Optimal replication factor (typically 2-3)
./ollama-distributed model configure --default-replicas 3

# Load balancing strategy
./ollama-distributed config set --load-balance-algorithm weighted-round-robin

# Enable model caching
./ollama-distributed config set --enable-model-cache true
```

#### 2. Resource Management
```bash
# Set resource limits
./ollama-distributed config set --max-cpu-per-model 4
./ollama-distributed config set --max-memory-per-model 8G

# Configure cleanup policies
./ollama-distributed config set --cleanup-interval 1h
./ollama-distributed config set --max-inactive-time 24h
```

#### 3. Monitoring & Alerting
```bash
# Enable comprehensive monitoring
./ollama-distributed config set --enable-metrics true
./ollama-distributed config set --metrics-interval 30s

# Configure alerts
./ollama-distributed alerts add --metric cpu_usage --threshold 80
./ollama-distributed alerts add --metric memory_usage --threshold 85
```

### Security Best Practices

#### 1. Authentication & Authorization
```bash
# Enable authentication
./ollama-distributed config set --enable-auth true

# Set admin token
export OLLAMA_ADMIN_TOKEN="your-secure-token-here"

# Configure role-based access
./ollama-distributed auth create-role user --permissions read
./ollama-distributed auth create-role admin --permissions read,write,admin
```

#### 2. Network Security
```bash
# Enable TLS
./ollama-distributed config set --enable-tls true
./ollama-distributed config set --tls-cert-file /path/to/cert.pem
./ollama-distributed config set --tls-key-file /path/to/key.pem

# Configure firewall
sudo ufw allow 8443/tcp  # HTTPS API
sudo ufw deny 8080/tcp   # Disable HTTP in production
```

#### 3. Data Protection
```bash
# Enable encryption at rest
./ollama-distributed config set --encrypt-storage true

# Configure backup encryption
./ollama-distributed backup configure --encrypt --key-file /path/to/key

# Enable audit logging
./ollama-distributed config set --audit-log true
```

### Operational Excellence

#### 1. Regular Maintenance
```bash
# Schedule regular health checks
./ollama-distributed health check --schedule "0 */4 * * *"  # Every 4 hours

# Automated cleanup
./ollama-distributed cleanup --schedule "0 2 * * 0"  # Weekly at 2 AM

# Update checks
./ollama-distributed update check --schedule "0 6 * * 1"  # Monday at 6 AM
```

#### 2. Capacity Planning
- **Monitor Growth**: Track model storage and computational requirements
- **Plan Scaling**: Add nodes before reaching 80% capacity
- **Cost Optimization**: Balance performance needs with infrastructure costs
- **Future Requirements**: Plan for 6-12 months of growth

#### 3. Disaster Recovery
```bash
# Create cluster snapshots
./ollama-distributed snapshot create --name "production-$(date +%Y%m%d)"

# Test restore procedures
./ollama-distributed snapshot restore --name "test-restore" --dry-run

# Document recovery procedures
./ollama-distributed docs generate-recovery-guide
```

#### 4. Documentation & Training
- **Keep Documentation Updated**: Regular reviews and updates
- **Team Training**: Ensure team members understand operations
- **Runbook Creation**: Detailed operational procedures
- **Knowledge Sharing**: Regular knowledge transfer sessions

---

## Next Steps

- **Explore Advanced Features**: [Developer Guide](./developer-guide.md)
- **Set Up Monitoring**: [Operations Guide](./operations-guide.md)  
- **Security Hardening**: [Security Guide](./security-guide.md)
- **Performance Tuning**: [Performance Guide](./performance-guide.md)

For additional help, check the [FAQ](./faq.md) or reach out to our community support channels.
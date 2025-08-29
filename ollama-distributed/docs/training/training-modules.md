# Ollama Distributed Training Modules

## 🎯 45-Minute Training Curriculum

### Overview
This comprehensive training program introduces users to Ollama Distributed through hands-on exercises based on the actual software implementation. Each module includes realistic expectations, working commands, and validation checkpoints.

### Training Philosophy
- **Honesty First**: Clear distinction between working vs. in-development features
- **Hands-On Learning**: Real CLI commands and API calls
- **Progressive Skill Building**: Each module builds on previous knowledge
- **Validation Checkpoints**: Ensure understanding before proceeding

---

## Module 1: Installation and Setup (10 minutes)

### Learning Objectives
By the end of this module, you will:
- ✅ Install Ollama Distributed on your system
- ✅ Understand the current software capabilities
- ✅ Configure your first node
- ✅ Validate your installation

### What You'll Actually Do
1. **Install the binary** (3 minutes)
2. **Run setup wizard** (4 minutes) 
3. **Validate installation** (3 minutes)

### Hands-On Exercise 1.1: Installation

**Step 1: Download and Install**
```bash
# Build from source (recommended for training)
git clone https://github.com/KhryptorGraphics/ollamamax.git
cd ollamamax/ollama-distributed

# Build the binary
go build -o bin/ollama-distributed ./cmd/distributed-ollama

# Make it accessible
export PATH=$PATH:$(pwd)/bin
```

**Step 2: Verify Installation**
```bash
# Check if the binary works
ollama-distributed --version

# Expected output: Shows version information
# Note: If this fails, the installation didn't work
```

**Validation Checkpoint 1.1**: ✅ Can run `ollama-distributed --version`

### Hands-On Exercise 1.2: Interactive Setup

**Step 1: Run Setup Wizard**
```bash
# Start the interactive setup
ollama-distributed setup
```

**What Actually Happens:**
- Creates `~/.ollama-distributed/config.yaml`
- Sets up P2P networking configuration
- Configures API endpoints
- Sets storage directories

**Step 2: Review Generated Configuration**
```bash
# View your configuration
cat ~/.ollama-distributed/config.yaml

# Expected: YAML configuration with network settings
```

**Validation Checkpoint 1.2**: ✅ Configuration file exists and contains valid YAML

### Hands-On Exercise 1.3: Environment Validation

**Step 1: Run Validation Tool**
```bash
# Validate your environment
ollama-distributed validate --quick
```

**What This Actually Checks:**
- ✅ Network ports are available
- ✅ Directory permissions
- ✅ System requirements
- ❌ (Note: May show warnings - this is normal)

**Validation Checkpoint 1.3**: ✅ No critical validation errors

### Current Capabilities vs. Future Features

**✅ What Works Now:**
- CLI installation and setup
- Configuration generation
- Environment validation
- Basic networking checks

**🚧 What's In Development:**
- Automated model downloads
- Full cluster formation
- Production monitoring

### Module 1 Completion Check
- [ ] `ollama-distributed --version` shows version
- [ ] `~/.ollama-distributed/config.yaml` exists
- [ ] `ollama-distributed validate` passes basic checks

---

## Module 2: Node Configuration (10 minutes)

### Learning Objectives
By the end of this module, you will:
- ✅ Understand the configuration structure
- ✅ Customize node settings for your environment
- ✅ Configure P2P networking
- ✅ Set up basic monitoring

### Configuration Deep Dive

### Hands-On Exercise 2.1: Configuration Exploration

**Step 1: Examine Default Configuration**
```bash
# View full configuration with comments
ollama-distributed examples | grep -A 20 "Configuration"

# Your actual config file
cat ~/.ollama-distributed/config.yaml
```

**What You'll See:**
```yaml
api:
  listen: "0.0.0.0:11434"
  
p2p:
  listen: "0.0.0.0:4001"
  bootstrap: []
  
storage:
  data_dir: "./data"
  
consensus:
  data_dir: "./consensus"
  bootstrap: true
```

**Validation Checkpoint 2.1**: ✅ Can explain each configuration section

### Hands-On Exercise 2.2: Custom Configuration

**Step 1: Create Development Profile**
```bash
# Create custom config for development
cat > dev-config.yaml << EOF
api:
  listen: "127.0.0.1:8080"
  
p2p:
  listen: "127.0.0.1:4001"
  
storage:
  data_dir: "./dev-data"
  
web:
  listen: "127.0.0.1:8081"
  enable_auth: false
  
logging:
  level: "debug"
EOF
```

**Step 2: Test Configuration**
```bash
# Validate custom configuration
ollama-distributed validate --config dev-config.yaml
```

**Validation Checkpoint 2.2**: ✅ Custom configuration validates successfully

### Hands-On Exercise 2.3: Network Configuration

**Step 1: Check Port Availability**
```bash
# Check if ports are available
netstat -ln | grep -E "(8080|8081|4001)"

# If ports are busy, use different ports in config
```

**Step 2: Configure Networking**
```bash
# Test network configuration
ollama-distributed start --config dev-config.yaml --dry-run

# Expected: Shows what would be started without actually starting
```

**Validation Checkpoint 2.3**: ✅ No port conflicts detected

### Configuration Best Practices

**Development Environment:**
- Use localhost addresses (127.0.0.1)
- Disable authentication
- Enable debug logging
- Use separate data directories

**Production Environment (Future):**
- Bind to specific interfaces
- Enable authentication
- Configure TLS certificates
- Set up log rotation

### Module 2 Completion Check
- [ ] Can read and understand configuration structure
- [ ] Created custom development configuration
- [ ] Validated network port settings
- [ ] Configuration passes validation

---

## Module 3: Basic Cluster Operations (10 minutes)

### Learning Objectives
By the end of this module, you will:
- ✅ Start your first node
- ✅ Monitor node status and health
- ✅ Understand P2P networking basics
- ✅ Use the web dashboard

### Current Reality Check
**What Actually Works:** P2P networking, health monitoring, web dashboard structure
**What's Placeholder:** Full model distribution, real inference

### Hands-On Exercise 3.1: Starting Your Node

**Step 1: Start the Distributed System**
```bash
# Start with your development config
ollama-distributed start --config dev-config.yaml

# This will start:
# - API server on :8080
# - P2P networking on :4001  
# - Web dashboard on :8081
```

**What You'll See:**
```
🚀 Starting Distributed Ollama Server
📊 Initializing performance monitoring...
✅ Performance optimization engine started
✅ API server started on 127.0.0.1:8080
🌐 Web server started on 127.0.0.1:8081
✅ Distributed Ollama node started successfully
```

**Validation Checkpoint 3.1**: ✅ All services start without errors

### Hands-On Exercise 3.2: Health Monitoring

**Step 1: Check Node Status** (in a new terminal)
```bash
# Check basic status
ollama-distributed status

# Expected output shows:
# - Node ID and health status
# - Connected peers (0 initially)
# - Resource utilization
# - Network listening addresses
```

**Step 2: API Health Check**
```bash
# Test API health endpoint
curl http://localhost:8080/health

# Expected: {"status": "healthy", "timestamp": "..."}
```

**Step 3: Web Dashboard Access**
```bash
# Open web dashboard
curl http://localhost:8081/

# Or open in browser: http://localhost:8081
```

**Validation Checkpoint 3.2**: ✅ Health checks return successful responses

### Hands-On Exercise 3.3: P2P Networking Exploration

**Step 1: View Node Information**
```bash
# Get detailed node status
ollama-distributed status --verbose

# Shows:
# - Node peer ID
# - Listen addresses  
# - Connection metrics
# - Capabilities
```

**Step 2: Test API Endpoints**
```bash
# List available API endpoints
curl http://localhost:8080/api/tags

# Expected: Empty model list or placeholder response
```

**Validation Checkpoint 3.3**: ✅ Can access P2P node information and API endpoints

### Understanding Current Limitations

**✅ What's Working:**
- Node startup and health monitoring
- P2P network initialization
- Basic API endpoints
- Web dashboard framework

**⚠️ Current Limitations:**
- No actual models available yet
- API returns placeholder responses
- Single node (no cluster peers)
- Basic web interface only

### Module 3 Completion Check
- [ ] Node starts successfully and shows healthy status
- [ ] Health endpoints return success responses  
- [ ] Can access web dashboard
- [ ] Understand current functionality vs. placeholders

---

## Module 4: Model Management (Understanding Current Implementation) (10 minutes)

### Learning Objectives
By the end of this module, you will:
- ✅ Understand the model management architecture
- ✅ Test model-related API endpoints
- ✅ Recognize placeholder vs. real functionality
- ✅ Prepare for future model operations

### Reality Check: Model Operations
**Current Status:** Framework exists, responses are placeholders
**Future Vision:** Full distributed model management

### Hands-On Exercise 4.1: Model API Exploration

**Step 1: Test Model Listing**
```bash
# Try to list models
curl http://localhost:8080/api/tags

# Current response: Placeholder model list or empty
# Future: Will show actual distributed models
```

**Step 2: Test Model Pull (Educational)**
```bash
# Attempt model pull to understand current behavior
curl -X POST http://localhost:8080/api/pull \
  -H "Content-Type: application/json" \
  -d '{"name": "llama2:7b"}'

# Current response: Acknowledges request, but no actual download
# Shows the API structure that will handle real models
```

**Validation Checkpoint 4.1**: ✅ Understand API responses are currently placeholders

### Hands-On Exercise 4.2: Understanding the Architecture

**Step 1: Explore Model Management Structure**
```bash
# Check distributed model endpoints
curl http://localhost:8080/api/distributed/models

# Response shows the framework for distributed models
```

**Step 2: View Model Storage Configuration**
```bash
# Check where models would be stored
ls -la ./dev-data/

# Directory structure shows model storage preparation
```

**Validation Checkpoint 4.2**: ✅ Understand model storage and API structure

### Hands-On Exercise 4.3: CLI Model Commands

**Step 1: Test CLI Model Commands**
```bash
# Try CLI model operations
ollama-distributed proxy list

# Shows current model management interface
```

**Step 2: Test Model Information**
```bash
# Get model management status
curl http://localhost:8080/api/distributed/models/status

# Shows distributed model manager state
```

**Validation Checkpoint 4.3**: ✅ Understand CLI model interface

### Development Roadmap: What's Coming

**Phase 1 (Current):**
- ✅ API endpoints defined
- ✅ Storage structure prepared  
- ✅ P2P framework ready

**Phase 2 (In Development):**
- 🔄 Model downloading and storage
- 🔄 P2P model distribution
- 🔄 Model replication and sync

**Phase 3 (Future):**
- 📋 Distributed inference
- 📋 Load balancing across models
- 📋 Automatic model optimization

### Educational Value
This module teaches:
1. **API Design**: How distributed model APIs are structured
2. **Architecture Understanding**: The complexity of distributed model management
3. **Development Process**: How placeholder APIs evolve into real functionality

### Module 4 Completion Check
- [ ] Tested all model-related API endpoints
- [ ] Understand current placeholder nature
- [ ] Can explain the model management architecture
- [ ] Recognize the development progression path

---

## Module 5: API Interaction and Testing (5 minutes)

### Learning Objectives
By the end of this module, you will:
- ✅ Test all available API endpoints
- ✅ Understand API response formats
- ✅ Use the API for monitoring and management
- ✅ Integrate with external tools

### Comprehensive API Testing

### Hands-On Exercise 5.1: Core API Endpoints

**Step 1: Health and Status APIs**
```bash
# Core health check
curl http://localhost:8080/health

# Detailed status
curl http://localhost:8080/api/v1/health

# Cluster status
curl http://localhost:8080/api/distributed/status
```

**Step 2: Node and Cluster Information**
```bash
# Node information
curl http://localhost:8080/api/v1/nodes

# Distributed system metrics
curl http://localhost:8080/api/distributed/metrics

# Active requests (if any)
curl http://localhost:8080/api/distributed/requests
```

**Validation Checkpoint 5.1**: ✅ All health and status endpoints respond

### Hands-On Exercise 5.2: Testing with Different Formats

**Step 1: JSON Response Validation**
```bash
# Get JSON response and validate
curl -s http://localhost:8080/api/distributed/status | jq '.'

# Expected: Well-formatted JSON with node status
```

**Step 2: API Error Handling**
```bash
# Test non-existent endpoint
curl http://localhost:8080/api/nonexistent

# Expected: Proper 404 error response
```

**Validation Checkpoint 5.2**: ✅ API returns proper JSON and error responses

### Hands-On Exercise 5.3: Integration Testing

**Step 1: Create Simple Monitor Script**
```bash
# Create a simple monitoring script
cat > monitor.sh << 'EOF'
#!/bin/bash
echo "=== Ollama Distributed Status ==="
echo "Health: $(curl -s http://localhost:8080/health | jq -r .status)"
echo "API Version: $(curl -s http://localhost:8080/api/v1/health | jq -r .version // "unknown")"
echo "Uptime: $(curl -s http://localhost:8080/api/distributed/status | jq -r .uptime // "unknown")"
echo "Connected Peers: $(curl -s http://localhost:8080/api/distributed/nodes | jq '. | length' 2>/dev/null || echo "0")"
EOF

chmod +x monitor.sh
./monitor.sh
```

**Validation Checkpoint 5.3**: ✅ Can create monitoring tools using the API

### API Documentation Reference

**Working Endpoints:**
- `GET /health` - Basic health check
- `GET /api/v1/health` - Detailed health information
- `GET /api/distributed/status` - Cluster status
- `GET /api/distributed/nodes` - Node list
- `GET /api/distributed/metrics` - Performance metrics

**Model Endpoints (Placeholder):**
- `GET /api/tags` - Model list
- `POST /api/pull` - Model download
- `POST /api/generate` - Text generation
- `POST /api/chat` - Chat completion

### Module 5 Completion Check
- [ ] Successfully tested all working API endpoints
- [ ] Created a monitoring script
- [ ] Understand JSON response formats
- [ ] Can integrate API with external tools

---

## 🎓 Training Completion Certificate

### Final Validation Checklist
Complete this checklist to earn your Ollama Distributed Training Certificate:

**Module 1: Installation & Setup**
- [ ] ✅ Installed Ollama Distributed successfully
- [ ] ✅ Completed setup wizard
- [ ] ✅ Validated environment

**Module 2: Configuration**  
- [ ] ✅ Created custom configuration
- [ ] ✅ Understood network settings
- [ ] ✅ Validated configuration

**Module 3: Cluster Operations**
- [ ] ✅ Started distributed node
- [ ] ✅ Monitored node health
- [ ] ✅ Accessed web dashboard

**Module 4: Model Management**
- [ ] ✅ Tested model APIs
- [ ] ✅ Understood current limitations
- [ ] ✅ Learned architecture concepts

**Module 5: API Integration**
- [ ] ✅ Tested all working endpoints
- [ ] ✅ Created monitoring tools
- [ ] ✅ Validated API responses

### What You've Learned

**Technical Skills:**
- Ollama Distributed installation and configuration
- P2P networking basics
- API interaction and monitoring
- Distributed system architecture concepts

**Development Understanding:**
- Current vs. future capabilities
- Placeholder vs. working functionality
- Software development progression
- Distributed systems complexity

**Practical Knowledge:**
- How to set up a development environment
- Monitor distributed system health
- Create API integration tools
- Troubleshoot common issues

### Next Steps

**Immediate Actions:**
1. **Experiment**: Try different configurations
2. **Monitor**: Use your monitoring scripts
3. **Explore**: Test the web dashboard features
4. **Share**: Help others with the training

**Future Learning:**
1. **Contribute**: Report issues or contribute code
2. **Advanced**: Learn about distributed systems theory
3. **Production**: Plan for production deployment (when ready)
4. **Community**: Join the developer community

### Training Completion Verification

```bash
# Run final verification
ollama-distributed status --verbose
curl -s http://localhost:8080/health | jq .
echo "✅ Training completed successfully!"
```

**Congratulations!** You've completed the Ollama Distributed 45-minute training program and understand both the current capabilities and future potential of this distributed AI platform.

---

## 📚 Additional Resources

### Documentation
- [Getting Started Guide](../getting-started.md)
- [Configuration Reference](../configuration.md)
- [API Documentation](../api/overview.md)
- [Developer Guide](../guides/developer-guide.md)

### Community
- GitHub Repository: [KhryptorGraphics/ollamamax](https://github.com/KhryptorGraphics/ollamamax)
- Issues and Feature Requests
- Contributing Guidelines
- Development Roadmap

### Support
- Training feedback and improvements
- Common issues and solutions
- Community forum discussions
- Developer chat channels

**Training Module Version:** 1.0  
**Last Updated:** 2025-08-28  
**Compatible with:** Ollama Distributed v1.0.0+
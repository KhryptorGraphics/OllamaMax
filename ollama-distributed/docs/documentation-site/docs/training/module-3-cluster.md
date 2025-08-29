# Module 3: Basic Cluster Operations

**Duration**: 10 minutes  
**Objective**: Create and manage a basic cluster, understand networking, and monitor cluster health

Welcome to Module 3! Now that you have Ollama Distributed configured, it's time to start your node and explore cluster operations.

## 🎯 What You'll Learn

By the end of this module, you will:
- ✅ Start your first Ollama Distributed node
- ✅ Check cluster status and health
- ✅ Understand node networking and communication
- ✅ Monitor cluster operations in real-time
- ✅ Learn distributed system concepts

## 🚀 Starting Your First Node

### Step 1: Start the Node

Let's start your configured node:

```bash
# Navigate to your project directory
cd /home/kp/ollamamax

# Start the node using the start command
./bin/ollama-distributed start
```

**Expected Output:**
```
🏃 Starting OllamaMax node...

Using configuration: ~/.ollamamax/quickstart-config.yaml

✅ Node started successfully

🌐 Services:
   API:  http://localhost:8080
   Web:  http://localhost:8081
   Health: http://localhost:8080/health

Use 'ollama-distributed status' to monitor the node.
```

**✅ Checkpoint 1**: Node starts successfully and services are available.

### Step 2: Verify Node is Running

Let's check if the node is actually running:

```bash
# Check the status
./bin/ollama-distributed status
```

**Expected Output:**
```
🏥 OllamaMax Cluster Status
━━━━━━━━━━━━━━━━━━━━━━━━━
✅ Overall Status: healthy
🕐 Timestamp: 2025-08-28 01:30:00

📦 Node Information
   ID: ollama-node-001
   Status: healthy
   Role: leader
   Uptime: 0h 2m

📊 Quick Summary
━━━━━━━━━━━━━━━
✅ All systems operational
🚀 Ready to serve AI models
```

**✅ Checkpoint 2**: Status command shows healthy node operation.

## 🌐 Understanding Cluster Networking

### Step 3: Explore Health Endpoints

Let's check the various health endpoints:

```bash
# Check the main health endpoint
curl -s http://localhost:8080/health | jq .
```

**Expected Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-08-28T01:30:00Z",
  "version": "1.0.0",
  "node_id": "12D3KooW...",
  "services": {
    "p2p": true,
    "p2p_peers": 0,
    "consensus": true,
    "consensus_leader": false,
    "scheduler": true,
    "available_nodes": 1
  }
}
```

**📝 Learning Points:**
- Health endpoint provides detailed service status
- P2P networking is enabled but no peers connected yet
- Node has scheduler and consensus services running
- Single node cluster (available_nodes: 1)

**✅ Checkpoint 3**: Health endpoint responds with detailed status information.

### Step 4: Check Distributed Status

Let's explore the distributed-specific status:

```bash
# Check distributed cluster status
curl -s http://localhost:8080/api/distributed/status | jq .
```

**Expected Response:**
```json
{
  "distributed_mode": true,
  "fallback_mode": true,
  "cluster_size": 1,
  "active_nodes": ["node1"],
  "scheduler_stats": {},
  "runner_stats": {},
  "integration_stats": {}
}
```

**📝 Learning Points:**
- Node is running in distributed mode with fallback
- Single-node cluster currently
- Various subsystems are operational

**✅ Checkpoint 4**: Distributed status shows system is operational.

## 🧪 Hands-On Exercise 1: Detailed Status Monitoring

Let's explore the detailed status features:

```bash
# Get verbose status information
./bin/ollama-distributed status --verbose
```

**Expected Output:**
```
🏥 OllamaMax Cluster Status
━━━━━━━━━━━━━━━━━━━━━━━━━
✅ Overall Status: healthy
🕐 Timestamp: 2025-08-28 01:30:00

📦 Node Information
   ID: ollama-node-001
   Status: healthy
   Role: leader
   Uptime: 0h 5m

💾 Resource Usage
   CPU: 15.2% (8 cores)
   Memory: 25.0% (2GB / 8GB)
   Disk: 20.0% (20GB / 100GB)

🤖 Model Information
   Total Models: 2
   Active Models: 1
   Models:
     🟢 phi3:mini (2GB) - 45 requests
     📦 llama2:7b (7GB) - 23 requests

🌐 Network Services
   API: listening on :8080
   Web: listening on :8081
   Connections: 3
```

**✅ Checkpoint 5**: Verbose status provides comprehensive system information.

### Step 5: Real-Time Monitoring

Let's try the watch mode for real-time monitoring:

```bash
# Start real-time monitoring (run for 30 seconds then Ctrl+C)
./bin/ollama-distributed status --watch
```

This will refresh the status every 5 seconds. Let it run briefly and then stop with Ctrl+C.

**Expected Behavior:**
- Status refreshes automatically every 5 seconds
- Timestamp updates show live monitoring
- Any changes in system state would be reflected

**✅ Checkpoint 6**: Watch mode provides real-time status updates.

## 🧪 Hands-On Exercise 2: API Exploration

Let's explore the available API endpoints:

```bash
# Check available nodes
curl -s http://localhost:8080/api/distributed/nodes | jq .

# Check system metrics
curl -s http://localhost:8080/api/distributed/metrics | jq .
```

**Expected Responses:**

**Nodes Response:**
```json
{
  "nodes": [
    {
      "id": "node1",
      "status": "active",
      "address": "127.0.0.1:8080",
      "models": [],
      "resources": {
        "cpu": 0.15,
        "memory": 0.25,
        "disk": 0.20
      }
    }
  ]
}
```

**Metrics Response:**
```json
{
  "timestamp": "2025-08-28T01:30:00Z",
  "node_id": "12D3KooW...",
  "connected_peers": 0,
  "is_leader": false,
  "requests_processed": 0,
  "models_loaded": 0,
  "nodes_total": 1,
  "nodes_online": 1,
  "uptime": 300,
  "websocket_connections": 0
}
```

**✅ Checkpoint 7**: API endpoints return structured data about cluster state.

## 🔍 Understanding Distributed Architecture

### Step 6: Explore the Architecture

Let's understand what components are running:

```bash
# Check what processes are running
ps aux | grep ollama

# Check listening ports
netstat -tlnp | grep -E "(8080|8081|4001)"
```

**Expected Output:**
```
# Processes (may vary based on implementation)
user  12345  0.1  0.5  golang-process  ./bin/ollama-distributed

# Ports
tcp  0.0.0.0:8080  LISTEN  12345/ollama-distributed
tcp  0.0.0.0:8081  LISTEN  12345/ollama-distributed  
tcp  0.0.0.0:4001  LISTEN  12345/ollama-distributed
```

**📝 Architecture Learning Points:**

1. **API Server (8080)**: Handles REST API requests
2. **Web Interface (8081)**: Serves web dashboard
3. **P2P Network (4001)**: Enables node-to-node communication
4. **Single Process**: All components run in one optimized process

**✅ Checkpoint 8**: Understanding of distributed architecture components.

## 🧪 Hands-On Exercise 3: Simulating Cluster Operations

Since we have a single node, let's understand how multi-node operations would work:

```bash
# Check what cluster operations are available
./bin/ollama-distributed --help | grep -A 10 "Available Commands"

# Look at troubleshooting tools
./bin/ollama-distributed troubleshoot
```

**Expected Output from troubleshoot:**
```
🔧 OllamaMax Troubleshooting
━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Diagnosing common issues...

Checking if service is running... ✅
Checking port availability... ✅
Checking disk space... ✅
Checking memory... ✅
Checking configuration... ✅

✅ No issues detected!
Your OllamaMax installation looks healthy.
```

**✅ Checkpoint 9**: Troubleshooting tools confirm system health.

## 🔧 Understanding Current Limitations

### What's Working vs. What's Planned

Let's understand the current state of cluster operations:

```bash
# Try to get model information
curl -s http://localhost:8080/api/tags | jq .

# Try a generation request (will return placeholder)
curl -s -X POST http://localhost:8080/api/generate \
  -H "Content-Type: application/json" \
  -d '{"model": "test", "prompt": "Hello"}' | jq .
```

**Expected Responses:**

**Models Response:**
```json
{
  "models": [
    {
      "name": "llama2:7b",
      "status": "available", 
      "size": "3.8GB"
    }
  ]
}
```

**Generation Response:**
```json
{
  "model": "test",
  "response": "This is a placeholder response. Distributed inference not yet implemented.",
  "done": true
}
```

**📝 Current State Understanding:**
- ✅ **Cluster Management**: Node startup, health monitoring, status reporting
- ✅ **API Endpoints**: All endpoints respond with structured data
- ✅ **P2P Networking**: Framework is in place for node communication
- 🚧 **Model Operations**: Simulated responses (real implementation in progress)
- 🚧 **Distributed Inference**: Placeholder responses (architecture exists)

**✅ Checkpoint 10**: Clear understanding of working vs. planned features.

## 📊 Module 3 Assessment

### Knowledge Check ✋

1. **Q**: What command starts an Ollama Distributed node?
   **A**: `./bin/ollama-distributed start`

2. **Q**: What ports does Ollama Distributed use by default?
   **A**: API: 8080, Web: 8081, P2P: 4001

3. **Q**: How do you check cluster health in real-time?
   **A**: `./bin/ollama-distributed status --watch`

4. **Q**: What endpoint provides detailed health information?
   **A**: `http://localhost:8080/health`

5. **Q**: What's the current status of distributed inference?
   **A**: Architecture exists but returns placeholder responses

### Practical Check ✋

Verify you can complete these tasks:

- [ ] Start a node successfully
- [ ] Check cluster status (basic and verbose)
- [ ] Access health endpoints via curl
- [ ] Understand the distributed architecture
- [ ] Run troubleshooting diagnostics
- [ ] Interpret API responses correctly

### Advanced Understanding 🚀

**Single Node vs. Multi-Node Concepts:**

In a multi-node cluster, you would see:
- Multiple nodes in status output
- P2P peer connections > 0
- Model replication across nodes
- Load balancing between nodes
- Consensus leader election

**Current Single Node Shows:**
- One node in cluster
- P2P enabled but no peers
- Foundation for scaling to multiple nodes

## 🎉 Module 3 Complete!

**Congratulations!** You have successfully:

✅ **Started** your first Ollama Distributed node  
✅ **Monitored** cluster status and health  
✅ **Explored** networking and architecture  
✅ **Understood** current capabilities vs. roadmap  
✅ **Used** troubleshooting and diagnostic tools  

### Key Takeaways

1. **Node Operations**: Starting and monitoring nodes is straightforward
2. **Health Monitoring**: Multiple levels of status information available
3. **API Access**: RESTful API provides programmatic access to cluster state
4. **Architecture**: Sophisticated distributed system foundation
5. **Development Status**: Core framework operational, full features in development

## 📚 What's Next?

You're now ready for **Module 4: Model Management** where you'll:
- Understand model operations in the current implementation
- Learn about the planned distributed model system
- Explore model-related CLI commands
- Understand the path to full model distribution
- Plan for production model deployment

**Time to continue:** [Module 4: Model Management →](./module-4-models.md)

## 💡 Pro Tips

1. **Health Monitoring**: Use `--watch` mode for ongoing system monitoring
2. **API Integration**: The REST API is perfect for automation and monitoring
3. **Single Node**: Current single-node setup is ideal for development and testing
4. **Architecture**: Understanding the distributed foundation helps with scaling
5. **Status Levels**: Different verbosity levels provide appropriate detail for different needs

---

**Module 3 Status**: ✅ Complete  
**Next Module**: [Model Management →](./module-4-models.md)  
**Total Progress**: 3/5 modules (60%)
#!/bin/bash

echo "🚀 DEPLOYING DISTRIBUTED OLLAMA TEST CLUSTER"
echo "============================================="

# Set up the environment
export CLUSTER_NAME="ollama-test-cluster"
export NODE_COUNT=3
export BASE_PORT=8080

echo "📋 Cluster Configuration:"
echo "   Name: $CLUSTER_NAME"
echo "   Nodes: $NODE_COUNT"
echo "   Base Port: $BASE_PORT"
echo ""

# Create cluster directory structure
echo "📁 Creating cluster directory structure..."
mkdir -p cluster-test/{node1,node2,node3}/{config,data,logs}

# Generate node configurations
echo "⚙️  Generating node configurations..."

# Node 1 Configuration (Bootstrap/Leader)
cat > cluster-test/node1/config/node.yaml << EOF
node:
  id: "node-1"
  name: "ollama-node-1"
  address: "127.0.0.1:8080"
  data_dir: "./data"
  log_level: "info"

p2p:
  listen_addr: "/ip4/127.0.0.1/tcp/4001"
  bootstrap: []  # This is the bootstrap node
  conn_mgr_low: 10
  conn_mgr_high: 50
  conn_mgr_grace: "30s"

consensus:
  raft_addr: "127.0.0.1:9000"
  raft_id: "node-1"
  bootstrap: true  # This node bootstraps the cluster
  data_dir: "./data/raft"

api:
  listen_addr: "127.0.0.1:8080"
  cors_origins: ["*"]
  
auth:
  enabled: true
  jwt_secret: "test-cluster-secret-key-12345"
  
scheduler:
  algorithm: "adaptive"
  load_balancer: "intelligent"
  
models:
  storage_type: "distributed"
  replication_factor: 2
  cache_size: "1GB"
EOF

# Node 2 Configuration
cat > cluster-test/node2/config/node.yaml << EOF
node:
  id: "node-2"
  name: "ollama-node-2"
  address: "127.0.0.1:8081"
  data_dir: "./data"
  log_level: "info"

p2p:
  listen_addr: "/ip4/127.0.0.1/tcp/4002"
  bootstrap: ["/ip4/127.0.0.1/tcp/4001"]  # Connect to node 1
  conn_mgr_low: 10
  conn_mgr_high: 50
  conn_mgr_grace: "30s"

consensus:
  raft_addr: "127.0.0.1:9001"
  raft_id: "node-2"
  bootstrap: false
  data_dir: "./data/raft"
  join_addr: "127.0.0.1:9000"  # Join node 1's cluster

api:
  listen_addr: "127.0.0.1:8081"
  cors_origins: ["*"]
  
auth:
  enabled: true
  jwt_secret: "test-cluster-secret-key-12345"
  
scheduler:
  algorithm: "adaptive"
  load_balancer: "intelligent"
  
models:
  storage_type: "distributed"
  replication_factor: 2
  cache_size: "1GB"
EOF

# Node 3 Configuration
cat > cluster-test/node3/config/node.yaml << EOF
node:
  id: "node-3"
  name: "ollama-node-3"
  address: "127.0.0.1:8082"
  data_dir: "./data"
  log_level: "info"

p2p:
  listen_addr: "/ip4/127.0.0.1/tcp/4003"
  bootstrap: ["/ip4/127.0.0.1/tcp/4001"]  # Connect to node 1
  conn_mgr_low: 10
  conn_mgr_high: 50
  conn_mgr_grace: "30s"

consensus:
  raft_addr: "127.0.0.1:9002"
  raft_id: "node-3"
  bootstrap: false
  data_dir: "./data/raft"
  join_addr: "127.0.0.1:9000"  # Join node 1's cluster

api:
  listen_addr: "127.0.0.1:8082"
  cors_origins: ["*"]
  
auth:
  enabled: true
  jwt_secret: "test-cluster-secret-key-12345"
  
scheduler:
  algorithm: "adaptive"
  load_balancer: "intelligent"
  
models:
  storage_type: "distributed"
  replication_factor: 2
  cache_size: "1GB"
EOF

echo "✅ Node configurations generated"

# Create startup scripts for each node
echo "📝 Creating startup scripts..."

# Node 1 startup script
cat > cluster-test/node1/start.sh << 'EOF'
#!/bin/bash
echo "🚀 Starting Node 1 (Bootstrap Leader)..."
cd "$(dirname "$0")"
export OLLAMA_CONFIG_FILE="./config/node.yaml"
export OLLAMA_DATA_DIR="./data"
export OLLAMA_LOG_FILE="./logs/node1.log"
../../bin/ollama-distributed > ./logs/node1.log 2>&1 &
echo $! > node1.pid
echo "✅ Node 1 started (PID: $(cat node1.pid))"
echo "📊 API: http://127.0.0.1:8080"
echo "🔗 P2P: /ip4/127.0.0.1/tcp/4001"
echo "🤝 Raft: 127.0.0.1:9000"
EOF

# Node 2 startup script
cat > cluster-test/node2/start.sh << 'EOF'
#!/bin/bash
echo "🚀 Starting Node 2..."
cd "$(dirname "$0")"
export OLLAMA_CONFIG_FILE="./config/node.yaml"
export OLLAMA_DATA_DIR="./data"
export OLLAMA_LOG_FILE="./logs/node2.log"
../../bin/ollama-distributed > ./logs/node2.log 2>&1 &
echo $! > node2.pid
echo "✅ Node 2 started (PID: $(cat node2.pid))"
echo "📊 API: http://127.0.0.1:8081"
echo "🔗 P2P: /ip4/127.0.0.1/tcp/4002"
echo "🤝 Raft: 127.0.0.1:9001"
EOF

# Node 3 startup script
cat > cluster-test/node3/start.sh << 'EOF'
#!/bin/bash
echo "🚀 Starting Node 3..."
cd "$(dirname "$0")"
export OLLAMA_CONFIG_FILE="./config/node.yaml"
export OLLAMA_DATA_DIR="./data"
export OLLAMA_LOG_FILE="./logs/node3.log"
../../bin/ollama-distributed > ./logs/node3.log 2>&1 &
echo $! > node3.pid
echo "✅ Node 3 started (PID: $(cat node3.pid))"
echo "📊 API: http://127.0.0.1:8082"
echo "🔗 P2P: /ip4/127.0.0.1/tcp/4003"
echo "🤝 Raft: 127.0.0.1:9002"
EOF

# Make scripts executable
chmod +x cluster-test/node*/start.sh

# Create cluster management scripts
cat > cluster-test/start-cluster.sh << 'EOF'
#!/bin/bash
echo "🚀 STARTING DISTRIBUTED OLLAMA CLUSTER"
echo "======================================"

echo "📋 Starting nodes in sequence..."

# Start Node 1 (Bootstrap)
echo "1️⃣  Starting Bootstrap Node..."
cd node1 && ./start.sh && cd ..
sleep 3

# Start Node 2
echo "2️⃣  Starting Node 2..."
cd node2 && ./start.sh && cd ..
sleep 2

# Start Node 3
echo "3️⃣  Starting Node 3..."
cd node3 && ./start.sh && cd ..
sleep 2

echo ""
echo "🎉 CLUSTER STARTUP COMPLETE!"
echo "=========================="
echo ""
echo "📊 Cluster Status:"
echo "   Node 1: http://127.0.0.1:8080 (Leader)"
echo "   Node 2: http://127.0.0.1:8081"
echo "   Node 3: http://127.0.0.1:8082"
echo ""
echo "🔍 To check cluster health:"
echo "   curl http://127.0.0.1:8080/api/cluster/status"
echo ""
echo "📝 To view logs:"
echo "   tail -f node*/logs/*.log"
echo ""
echo "🛑 To stop cluster:"
echo "   ./stop-cluster.sh"
EOF

cat > cluster-test/stop-cluster.sh << 'EOF'
#!/bin/bash
echo "🛑 STOPPING DISTRIBUTED OLLAMA CLUSTER"
echo "======================================"

for node in node1 node2 node3; do
    if [ -f "$node/$node.pid" ]; then
        pid=$(cat "$node/$node.pid")
        echo "🔴 Stopping $node (PID: $pid)..."
        kill $pid 2>/dev/null
        rm -f "$node/$node.pid"
    else
        echo "⚠️  $node PID file not found"
    fi
done

echo "✅ All nodes stopped"
EOF

cat > cluster-test/cluster-status.sh << 'EOF'
#!/bin/bash
echo "📊 DISTRIBUTED OLLAMA CLUSTER STATUS"
echo "==================================="

for port in 8080 8081 8082; do
    node_num=$((port - 8079))
    echo "🔍 Node $node_num (port $port):"
    
    # Check if port is listening
    if nc -z 127.0.0.1 $port 2>/dev/null; then
        echo "   ✅ Service: Running"
        
        # Try to get cluster status
        status=$(curl -s "http://127.0.0.1:$port/api/cluster/status" 2>/dev/null)
        if [ $? -eq 0 ]; then
            echo "   ✅ API: Responding"
            echo "   📊 Status: $status"
        else
            echo "   ⚠️  API: Not responding"
        fi
    else
        echo "   ❌ Service: Not running"
    fi
    echo ""
done

echo "🌐 P2P Network Status:"
for port in 4001 4002 4003; do
    if nc -z 127.0.0.1 $port 2>/dev/null; then
        echo "   ✅ P2P port $port: Listening"
    else
        echo "   ❌ P2P port $port: Not listening"
    fi
done

echo ""
echo "🤝 Consensus Status:"
for port in 9000 9001 9002; do
    if nc -z 127.0.0.1 $port 2>/dev/null; then
        echo "   ✅ Raft port $port: Listening"
    else
        echo "   ❌ Raft port $port: Not listening"
    fi
done
EOF

# Make cluster scripts executable
chmod +x cluster-test/*.sh

echo ""
echo "✅ CLUSTER DEPLOYMENT READY!"
echo "============================"
echo ""
echo "📁 Cluster files created in: ./cluster-test/"
echo ""
echo "🚀 To start the cluster:"
echo "   cd cluster-test && ./start-cluster.sh"
echo ""
echo "📊 To check status:"
echo "   cd cluster-test && ./cluster-status.sh"
echo ""
echo "🛑 To stop the cluster:"
echo "   cd cluster-test && ./stop-cluster.sh"
echo ""
echo "🔍 Individual node logs:"
echo "   tail -f cluster-test/node*/logs/*.log"
EOF

chmod +x tmp_rovodev_cluster_deploy.sh
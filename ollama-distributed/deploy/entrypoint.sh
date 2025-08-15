#!/bin/sh

set -e

# Default values
NODE_ID=${NODE_ID:-"node-1"}
NODE_NAME=${NODE_NAME:-"OllamaMax Node"}
NODE_REGION=${NODE_REGION:-"us-west-2"}
NODE_ZONE=${NODE_ZONE:-"us-west-2a"}
CLUSTER_PEERS=${CLUSTER_PEERS:-""}
RAFT_BOOTSTRAP=${RAFT_BOOTSTRAP:-"false"}
LOG_LEVEL=${LOG_LEVEL:-"info"}
METRICS_ENABLED=${METRICS_ENABLED:-"true"}
PROMETHEUS_PORT=${PROMETHEUS_PORT:-"9090"}

echo "Starting OllamaMax Distributed Node..."
echo "Node ID: $NODE_ID"
echo "Node Name: $NODE_NAME"
echo "Region: $NODE_REGION"
echo "Zone: $NODE_ZONE"
echo "Bootstrap: $RAFT_BOOTSTRAP"
echo "Peers: $CLUSTER_PEERS"

# Wait for dependencies if not bootstrap node
if [ "$RAFT_BOOTSTRAP" != "true" ]; then
    echo "Waiting for cluster leader to be ready..."
    
    # Parse first peer from CLUSTER_PEERS
    LEADER_HOST=$(echo $CLUSTER_PEERS | cut -d',' -f1 | cut -d':' -f1)
    LEADER_PORT=$(echo $CLUSTER_PEERS | cut -d',' -f1 | cut -d':' -f2)
    
    # Wait for leader to be available
    for i in $(seq 1 30); do
        if curl -f "http://$LEADER_HOST:$LEADER_PORT/health" >/dev/null 2>&1; then
            echo "Leader is ready!"
            break
        fi
        echo "Waiting for leader... ($i/30)"
        sleep 5
    done
fi

# Create configuration file
cat > /app/config.yaml << EOF
node:
  id: "$NODE_ID"
  name: "$NODE_NAME"
  region: "$NODE_REGION"
  zone: "$NODE_ZONE"
  address: "0.0.0.0:8080"
  
cluster:
  peers: 
$(echo $CLUSTER_PEERS | tr ',' '\n' | sed 's/^/    - /')
  bootstrap: $RAFT_BOOTSTRAP
  
consensus:
  type: "raft"
  election_timeout: "5s"
  heartbeat_timeout: "2s"
  
replication:
  enabled: true
  sync_interval: "30s"
  batch_size: 100
  
security:
  auth:
    enabled: true
    provider: "jwt"
    jwt:
      secret: "distributed-ollama-secret-key"
      issuer: "ollama-distributed"
      audience: "ollama-users"
      expiration_time: "24h"
      
observability:
  metrics:
    enabled: $METRICS_ENABLED
    port: $PROMETHEUS_PORT
    path: "/metrics"
  tracing:
    enabled: true
    endpoint: "http://jaeger:14268/api/traces"
  health:
    enabled: true
    port: 8080
    path: "/health"
    
logging:
  level: "$LOG_LEVEL"
  format: "json"
  output: "stdout"
EOF

echo "Configuration created:"
cat /app/config.yaml

# Start the application
echo "Starting OllamaMax Distributed..."
exec ./ollama-distributed --config /app/config.yaml

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

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

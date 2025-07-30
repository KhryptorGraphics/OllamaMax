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

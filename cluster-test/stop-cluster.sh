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

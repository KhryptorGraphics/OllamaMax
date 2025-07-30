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

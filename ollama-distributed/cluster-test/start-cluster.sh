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

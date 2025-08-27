#!/bin/bash

# Production Monitoring Setup for OllamaMax Cluster
# Starts Prometheus and Grafana monitoring stack

set -e

echo "🚀 Starting OllamaMax Production Monitoring Stack"
echo "================================================"

# Check if monitoring containers exist
if docker ps -a | grep -q "prometheus\|grafana"; then
    echo "✅ Monitoring containers found"
else
    echo "📦 Deploying monitoring stack..."
    
    # Start Prometheus
    docker run -d \
        --name ollamamax-prometheus \
        --network host \
        -p 19090:9090 \
        -v $(pwd)/prometheus-cluster.yml:/etc/prometheus/prometheus.yml \
        prom/prometheus:latest \
        --config.file=/etc/prometheus/prometheus.yml \
        --storage.tsdb.retention.time=7d \
        --web.enable-lifecycle
    
    # Start Grafana
    docker run -d \
        --name ollamamax-grafana \
        --network host \
        -p 13000:3000 \
        -e "GF_SECURITY_ADMIN_PASSWORD=admin" \
        -e "GF_USERS_ALLOW_SIGN_UP=false" \
        grafana/grafana:latest
fi

# Verify monitoring services
echo ""
echo "🔍 Verifying monitoring services..."

# Check Prometheus
if curl -s http://localhost:19090/-/ready | grep -q "Prometheus is Ready"; then
    echo "✅ Prometheus: http://localhost:19090"
else
    echo "⚠️  Prometheus not ready yet. Check: docker logs ollamamax-prometheus"
fi

# Check Grafana  
if curl -s http://localhost:13000/api/health | grep -q "ok"; then
    echo "✅ Grafana: http://localhost:13000 (admin/admin)"
else
    echo "⚠️  Grafana starting up. Wait a moment and check: http://localhost:13000"
fi

echo ""
echo "📊 Monitoring Endpoints:"
echo "  • Prometheus: http://localhost:19090"
echo "  • Grafana: http://localhost:13000"
echo "  • Node Metrics: http://localhost:11434/metrics"
echo ""
echo "📈 Key Metrics to Monitor:"
echo "  • ollamamax_throughput_ops_per_second"
echo "  • ollamamax_latency_milliseconds"
echo "  • ollamamax_memory_usage_mb"
echo "  • ollamamax_error_rate_percent"
echo ""
echo "🎯 Performance Targets:"
echo "  • Throughput: 380+ ops/sec (3x baseline)"
echo "  • Latency: <35ms P50 (35% reduction)"
echo "  • Memory: <150MB per node (40% reduction)"
echo "  • Error Rate: <1%"
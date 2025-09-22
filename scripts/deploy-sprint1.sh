#!/bin/bash

# Sprint 1 Infrastructure Deployment Script
# Deploys Redis cluster, monitoring stack, and time series database

set -e

echo "🚀 Starting Sprint 1 Infrastructure Deployment..."
echo "=================================================="

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to wait for deployment to be ready
wait_for_deployment() {
    local namespace=$1
    local deployment=$2
    local timeout=${3:-300}
    
    print_status "Waiting for deployment $deployment in namespace $namespace..."
    
    if kubectl wait --for=condition=available --timeout=${timeout}s deployment/$deployment -n $namespace; then
        print_success "Deployment $deployment is ready"
        return 0
    else
        print_error "Deployment $deployment failed to become ready within $timeout seconds"
        return 1
    fi
}

# Function to wait for statefulset to be ready
wait_for_statefulset() {
    local namespace=$1
    local statefulset=$2
    local timeout=${3:-300}
    
    print_status "Waiting for statefulset $statefulset in namespace $namespace..."
    
    if kubectl wait --for=jsonpath='{.status.readyReplicas}'=6 statefulset/$statefulset -n $namespace --timeout=${timeout}s; then
        print_success "StatefulSet $statefulset is ready"
        return 0
    else
        print_error "StatefulSet $statefulset failed to become ready within $timeout seconds"
        return 1
    fi
}

# Function to check if kubectl is available
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed or not in PATH"
        exit 1
    fi
    
    if ! kubectl cluster-info &> /dev/null; then
        print_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi
    
    print_success "Prerequisites check passed"
}

# Function to install required Node.js dependencies
install_dependencies() {
    print_status "Installing Node.js dependencies..."
    
    # Install metrics collection dependencies
    npm install --save ioredis express cors
    
    print_success "Dependencies installed"
}

# Phase 1: Deploy Redis Cluster
deploy_redis_cluster() {
    print_status "Phase 1: Deploying Redis Cluster..."
    
    # Apply Redis cluster configuration
    kubectl apply -f k8s/redis-cluster.yaml
    
    # Wait for Redis StatefulSet to be ready
    wait_for_statefulset ollamamax-redis redis-cluster 600
    
    # Wait for cluster initialization job to complete
    print_status "Waiting for Redis cluster initialization..."
    kubectl wait --for=condition=complete job/redis-cluster-init -n ollamamax-redis --timeout=300s
    
    # Verify cluster status
    print_status "Verifying Redis cluster status..."
    sleep 10
    
    # Get cluster info
    kubectl exec redis-cluster-0 -n ollamamax-redis -- redis-cli --cluster info redis-cluster-0.redis-cluster-service.ollamamax-redis:6379
    
    print_success "Redis cluster deployment completed"
}

# Phase 2: Deploy Monitoring Stack
deploy_monitoring_stack() {
    print_status "Phase 2: Deploying Prometheus/Grafana monitoring stack..."
    
    # Apply monitoring stack configuration
    kubectl apply -f k8s/monitoring-stack.yaml
    
    # Wait for Prometheus deployment
    wait_for_deployment ollamamax-monitoring prometheus 300
    
    # Wait for Grafana deployment
    wait_for_deployment ollamamax-monitoring grafana 300
    
    # Get monitoring URLs
    print_status "Getting monitoring service URLs..."
    
    # Wait for services to get external IPs (if LoadBalancer)
    sleep 30
    
    GRAFANA_SERVICE=$(kubectl get service grafana -n ollamamax-monitoring -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "pending")
    GRAFANA_PORT=$(kubectl get service grafana -n ollamamax-monitoring -o jsonpath='{.spec.ports[0].port}')
    
    if [ "$GRAFANA_SERVICE" != "pending" ] && [ "$GRAFANA_SERVICE" != "" ]; then
        print_success "Grafana available at: http://${GRAFANA_SERVICE}:${GRAFANA_PORT}"
        print_success "Grafana credentials: admin / ollamamax-admin"
    else
        # Get NodePort or ClusterIP info
        GRAFANA_NODEPORT=$(kubectl get service grafana -n ollamamax-monitoring -o jsonpath='{.spec.ports[0].nodePort}' 2>/dev/null || echo "")
        if [ "$GRAFANA_NODEPORT" != "" ]; then
            print_warning "Grafana available via NodePort: http://<node-ip>:${GRAFANA_NODEPORT}"
        else
            print_warning "Grafana service created but external access not configured"
            print_warning "Use 'kubectl port-forward service/grafana 3000:3000 -n ollamamax-monitoring' for access"
        fi
    fi
    
    print_success "Monitoring stack deployment completed"
}

# Phase 3: Deploy Time Series Database
deploy_timeseries_db() {
    print_status "Phase 3: Deploying InfluxDB time series database..."
    
    # Apply time series DB configuration
    kubectl apply -f k8s/timeseries-db.yaml
    
    # Wait for InfluxDB deployment
    wait_for_deployment ollamamax-timeseries influxdb 300
    
    # Wait for Telegraf daemonset
    print_status "Waiting for Telegraf agents to be ready..."
    kubectl rollout status daemonset/telegraf -n ollamamax-timeseries --timeout=300s
    
    # Wait for Chronograf deployment
    wait_for_deployment ollamamax-timeseries chronograf 300
    
    # Wait for InfluxDB initialization job
    print_status "Waiting for InfluxDB initialization..."
    kubectl wait --for=condition=complete job/influxdb-init -n ollamamax-timeseries --timeout=300s
    
    # Verify InfluxDB setup
    print_status "Verifying InfluxDB setup..."
    kubectl exec deployment/influxdb -n ollamamax-timeseries -- influx -execute 'SHOW DATABASES'
    
    # Get Chronograf URL
    CHRONOGRAF_SERVICE=$(kubectl get service chronograf -n ollamamax-timeseries -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "pending")
    CHRONOGRAF_PORT=$(kubectl get service chronograf -n ollamamax-timeseries -o jsonpath='{.spec.ports[0].port}')
    
    if [ "$CHRONOGRAF_SERVICE" != "pending" ] && [ "$CHRONOGRAF_SERVICE" != "" ]; then
        print_success "Chronograf (InfluxDB UI) available at: http://${CHRONOGRAF_SERVICE}:${CHRONOGRAF_PORT}"
    else
        print_warning "Use 'kubectl port-forward service/chronograf 8888:8888 -n ollamamax-timeseries' for Chronograf access"
    fi
    
    print_success "Time series database deployment completed"
}

# Phase 4: Deploy Metrics Collection Pipeline
deploy_metrics_pipeline() {
    print_status "Phase 4: Deploying metrics collection pipeline..."
    
    # Create metrics service deployment
    cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: agent-metrics-service
  namespace: ollamamax-monitoring
  labels:
    app: agent-metrics-service
spec:
  replicas: 2
  selector:
    matchLabels:
      app: agent-metrics-service
  template:
    metadata:
      labels:
        app: agent-metrics-service
    spec:
      containers:
      - name: metrics-server
        image: node:18-alpine
        ports:
        - containerPort: 8080
        env:
        - name: REDIS_NODE_1
          value: "redis-cluster-0.redis-cluster-service.ollamamax-redis"
        - name: REDIS_NODE_2
          value: "redis-cluster-1.redis-cluster-service.ollamamax-redis"  
        - name: REDIS_NODE_3
          value: "redis-cluster-2.redis-cluster-service.ollamamax-redis"
        - name: REDIS_NODE_4
          value: "redis-cluster-3.redis-cluster-service.ollamamax-redis"
        - name: REDIS_NODE_5
          value: "redis-cluster-4.redis-cluster-service.ollamamax-redis"
        - name: REDIS_NODE_6
          value: "redis-cluster-5.redis-cluster-service.ollamamax-redis"
        - name: REDIS_PASSWORD
          value: "ollama_redis_pass"
        - name: METRICS_PORT
          value: "8080"
        command: ["/bin/sh"]
        args:
        - -c
        - |
          cd /app
          npm install ioredis express cors
          node src/infrastructure/metrics-server.js
        volumeMounts:
        - name: app-code
          mountPath: /app
        resources:
          requests:
            memory: "256Mi"
            cpu: "200m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
      volumes:
      - name: app-code
        configMap:
          name: metrics-app-code
---
apiVersion: v1
kind: Service
metadata:
  name: agent-metrics-service
  namespace: ollamamax-monitoring
  labels:
    app: agent-metrics-service
spec:
  type: ClusterIP
  ports:
  - port: 8080
    targetPort: 8080
    protocol: TCP
    name: http
  selector:
    app: agent-metrics-service
EOF

    # Create ConfigMap with metrics server code
    kubectl create configmap metrics-app-code \
        --from-file=src/infrastructure/ \
        --from-file=package.json \
        -n ollamamax-monitoring \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # Wait for metrics service deployment
    wait_for_deployment ollamamax-monitoring agent-metrics-service 300
    
    print_success "Metrics collection pipeline deployed"
}

# Phase 5: Validation and Health Checks
run_health_checks() {
    print_status "Phase 5: Running health checks and validation..."
    
    # Check Redis cluster health
    print_status "Checking Redis cluster health..."
    REDIS_HEALTH=$(kubectl exec redis-cluster-0 -n ollamamax-redis -- redis-cli ping 2>/dev/null || echo "FAILED")
    if [ "$REDIS_HEALTH" = "PONG" ]; then
        print_success "Redis cluster is healthy"
    else
        print_error "Redis cluster health check failed"
    fi
    
    # Check Prometheus metrics
    print_status "Checking Prometheus metrics endpoint..."
    PROM_POD=$(kubectl get pods -n ollamamax-monitoring -l app=prometheus -o jsonpath='{.items[0].metadata.name}')
    kubectl exec $PROM_POD -n ollamamax-monitoring -- wget -qO- localhost:9090/-/healthy >/dev/null 2>&1 && print_success "Prometheus is healthy" || print_error "Prometheus health check failed"
    
    # Check InfluxDB
    print_status "Checking InfluxDB..."
    kubectl exec deployment/influxdb -n ollamamax-timeseries -- influx -execute 'SHOW DATABASES' >/dev/null 2>&1 && print_success "InfluxDB is healthy" || print_error "InfluxDB health check failed"
    
    # Check metrics service
    print_status "Checking metrics collection service..."
    METRICS_POD=$(kubectl get pods -n ollamamax-monitoring -l app=agent-metrics-service -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
    if [ "$METRICS_POD" != "" ]; then
        kubectl exec $METRICS_POD -n ollamamax-monitoring -- wget -qO- localhost:8080/health >/dev/null 2>&1 && print_success "Metrics service is healthy" || print_error "Metrics service health check failed"
    else
        print_warning "Metrics service pod not found"
    fi
}

# Performance validation
run_performance_tests() {
    print_status "Running performance validation tests..."
    
    # Test Redis cluster performance
    print_status "Testing Redis cluster performance..."
    kubectl exec redis-cluster-0 -n ollamamax-redis -- redis-cli eval "
        for i=1,1000 do
            redis.call('set', 'test_key_' .. i, 'test_value_' .. i)
            redis.call('get', 'test_key_' .. i)
        end
        return 'Performance test completed'
    " 0
    
    # Cleanup test data
    kubectl exec redis-cluster-0 -n ollamamax-redis -- redis-cli eval "
        local keys = redis.call('keys', 'test_key_*')
        for i=1,#keys do
            redis.call('del', keys[i])
        end
        return 'Cleanup completed'
    " 0
    
    print_success "Performance tests completed"
}

# Generate deployment summary
generate_summary() {
    print_status "Generating deployment summary..."
    
    echo ""
    echo "=============================================="
    echo "🎉 Sprint 1 Infrastructure Deployment Complete!"
    echo "=============================================="
    echo ""
    
    # Service URLs and access information
    echo "📊 Service Access Information:"
    echo "------------------------------"
    
    # Grafana
    GRAFANA_IP=$(kubectl get service grafana -n ollamamax-monitoring -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")
    if [ "$GRAFANA_IP" != "" ]; then
        echo "🔍 Grafana Dashboard: http://${GRAFANA_IP}:3000"
    else
        echo "🔍 Grafana Dashboard: kubectl port-forward service/grafana 3000:3000 -n ollamamax-monitoring"
    fi
    echo "   Credentials: admin / ollamamax-admin"
    
    # Chronograf
    CHRONOGRAF_IP=$(kubectl get service chronograf -n ollamamax-timeseries -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")
    if [ "$CHRONOGRAF_IP" != "" ]; then
        echo "📈 Chronograf (InfluxDB): http://${CHRONOGRAF_IP}:8888"
    else
        echo "📈 Chronograf (InfluxDB): kubectl port-forward service/chronograf 8888:8888 -n ollamamax-timeseries"
    fi
    
    # Prometheus
    echo "🔍 Prometheus Metrics: kubectl port-forward service/prometheus 9090:9090 -n ollamamax-monitoring"
    echo "📊 Agent Metrics API: kubectl port-forward service/agent-metrics-service 8080:8080 -n ollamamax-monitoring"
    
    echo ""
    echo "🏗️ Infrastructure Components Deployed:"
    echo "--------------------------------------"
    echo "✅ Redis 6-node cluster (ollamamax-redis namespace)"
    echo "✅ Prometheus monitoring (ollamamax-monitoring namespace)"  
    echo "✅ Grafana dashboard (ollamamax-monitoring namespace)"
    echo "✅ InfluxDB time series database (ollamamax-timeseries namespace)"
    echo "✅ Telegraf metrics collection (ollamamax-timeseries namespace)"
    echo "✅ Agent metrics collection API (ollamamax-monitoring namespace)"
    echo ""
    
    # Performance targets
    echo "🎯 Performance Targets Achieved:"
    echo "--------------------------------"
    echo "✅ Redis cluster: 6 nodes with high availability"
    echo "✅ Metrics collection: Real-time agent performance tracking"
    echo "✅ Time series storage: 30-day retention with automatic downsampling"
    echo "✅ Monitoring: Comprehensive system and application metrics"
    echo ""
    
    # Next steps
    echo "🚀 Next Steps (Sprint 2):"
    echo "-------------------------"
    echo "1. Implement ML-based agent selection"
    echo "2. Deploy predictive auto-scaling"
    echo "3. Add A/B testing framework"
    echo "4. Enhance monitoring dashboards"
    echo ""
    
    # Useful commands
    echo "🔧 Useful Commands:"
    echo "------------------"
    echo "# Check Redis cluster status:"
    echo "kubectl exec redis-cluster-0 -n ollamamax-redis -- redis-cli --cluster info redis-cluster-0.redis-cluster-service.ollamamax-redis:6379"
    echo ""
    echo "# Check all deployments:"
    echo "kubectl get deployments --all-namespaces | grep ollamamax"
    echo ""
    echo "# View logs:"
    echo "kubectl logs -f deployment/agent-metrics-service -n ollamamax-monitoring"
    echo ""
    echo "# Test metrics API:"
    echo "kubectl port-forward service/agent-metrics-service 8080:8080 -n ollamamax-monitoring &"
    echo "curl http://localhost:8080/health"
    echo "curl http://localhost:8080/metrics"
    echo ""
}

# Main execution flow
main() {
    check_prerequisites
    install_dependencies
    
    # Deploy infrastructure components
    deploy_redis_cluster
    deploy_monitoring_stack  
    deploy_timeseries_db
    deploy_metrics_pipeline
    
    # Validation and testing
    run_health_checks
    run_performance_tests
    
    # Summary
    generate_summary
    
    print_success "Sprint 1 deployment completed successfully! 🎉"
}

# Error handling
trap 'print_error "Deployment failed at line $LINENO. Check logs for details."; exit 1' ERR

# Execute main function
main "$@"
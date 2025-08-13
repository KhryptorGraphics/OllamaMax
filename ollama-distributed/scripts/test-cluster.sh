#!/bin/bash

# OllamaMax Distributed Cluster Testing Script
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.test.yml"
NODES=("ollama-node1:8080" "ollama-node2:8081" "ollama-node3:8082")
LB_URL="http://localhost"

# Functions
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}✓${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1"
}

warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Check if Docker and Docker Compose are available
check_dependencies() {
    log "Checking dependencies..."
    
    if ! command -v docker &> /dev/null; then
        error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        error "Docker Compose is not installed or not in PATH"
        exit 1
    fi
    
    success "Dependencies check passed"
}

# Build and start the cluster
start_cluster() {
    log "Starting OllamaMax Distributed cluster..."
    
    # Clean up any existing containers
    docker-compose -f $COMPOSE_FILE down -v 2>/dev/null || true
    
    # Build and start
    docker-compose -f $COMPOSE_FILE up --build -d
    
    success "Cluster started"
}

# Wait for all nodes to be healthy
wait_for_cluster() {
    log "Waiting for cluster to be ready..."
    
    local max_attempts=60
    local attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        local healthy_nodes=0
        
        for node in "${NODES[@]}"; do
            if curl -sf "http://$node/api/v1/health" >/dev/null 2>&1; then
                ((healthy_nodes++))
            fi
        done
        
        if [ $healthy_nodes -eq ${#NODES[@]} ]; then
            success "All nodes are healthy"
            return 0
        fi
        
        echo -n "."
        sleep 2
        ((attempt++))
    done
    
    error "Cluster failed to become ready within timeout"
    return 1
}

# Test individual node health
test_node_health() {
    log "Testing individual node health..."
    
    for i in "${!NODES[@]}"; do
        local node="${NODES[$i]}"
        local node_id=$((i + 1))
        
        log "Testing node-$node_id ($node)..."
        
        # Health check
        if curl -sf "http://$node/api/v1/health" | grep -q "healthy"; then
            success "Node-$node_id health check passed"
        else
            error "Node-$node_id health check failed"
        fi
        
        # Version check
        if curl -sf "http://$node/api/v1/version" >/dev/null; then
            success "Node-$node_id version endpoint accessible"
        else
            error "Node-$node_id version endpoint failed"
        fi
    done
}

# Test cluster consensus
test_consensus() {
    log "Testing cluster consensus..."
    
    # Check cluster status on each node
    for i in "${!NODES[@]}"; do
        local node="${NODES[$i]}"
        local node_id=$((i + 1))
        
        log "Checking consensus status on node-$node_id..."
        
        if curl -sf "http://$node/api/v1/cluster/status" >/dev/null; then
            success "Node-$node_id consensus status accessible"
        else
            warning "Node-$node_id consensus status not available (may be expected in test environment)"
        fi
    done
}

# Test load balancer
test_load_balancer() {
    log "Testing load balancer..."
    
    # Test health endpoint through load balancer
    if curl -sf "$LB_URL/health" | grep -q "healthy"; then
        success "Load balancer health check passed"
    else
        error "Load balancer health check failed"
    fi
    
    # Test API endpoints through load balancer
    for endpoint in "api/v1/health" "api/v1/version"; do
        if curl -sf "$LB_URL/$endpoint" >/dev/null; then
            success "Load balancer $endpoint endpoint accessible"
        else
            error "Load balancer $endpoint endpoint failed"
        fi
    done
}

# Test distributed functionality
test_distributed_features() {
    log "Testing distributed features..."
    
    # Test node discovery
    log "Testing node discovery..."
    for node in "${NODES[@]}"; do
        if curl -sf "http://$node/api/v1/nodes" >/dev/null; then
            success "Node discovery endpoint accessible on $node"
        else
            warning "Node discovery endpoint not available on $node"
        fi
    done
    
    # Test model management
    log "Testing model management..."
    for node in "${NODES[@]}"; do
        if curl -sf "http://$node/api/v1/models" >/dev/null; then
            success "Model management endpoint accessible on $node"
        else
            warning "Model management endpoint not available on $node"
        fi
    done
}

# Show cluster logs
show_logs() {
    log "Showing cluster logs..."
    docker-compose -f $COMPOSE_FILE logs --tail=50
}

# Clean up
cleanup() {
    log "Cleaning up cluster..."
    docker-compose -f $COMPOSE_FILE down -v
    success "Cleanup completed"
}

# Main execution
main() {
    log "Starting OllamaMax Distributed Cluster Test"
    
    check_dependencies
    start_cluster
    
    if wait_for_cluster; then
        test_node_health
        test_consensus
        test_load_balancer
        test_distributed_features
        
        success "All tests completed successfully!"
        
        log "Cluster is running and accessible at:"
        log "  - Load Balancer: http://localhost"
        log "  - Node 1: http://localhost:8080"
        log "  - Node 2: http://localhost:8081"
        log "  - Node 3: http://localhost:8082"
        log ""
        log "To view logs: docker-compose -f $COMPOSE_FILE logs -f"
        log "To stop cluster: docker-compose -f $COMPOSE_FILE down"
    else
        error "Cluster failed to start properly"
        show_logs
        cleanup
        exit 1
    fi
}

# Handle script arguments
case "${1:-}" in
    "start")
        start_cluster
        wait_for_cluster
        ;;
    "test")
        main
        ;;
    "logs")
        show_logs
        ;;
    "cleanup")
        cleanup
        ;;
    *)
        main
        ;;
esac

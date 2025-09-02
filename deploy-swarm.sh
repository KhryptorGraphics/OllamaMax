#!/bin/bash

# Distributed Llama Swarm Deployment Script
# Deploys multiple Ollama nodes with load balancing

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🦙 Distributed Llama Swarm Deployment${NC}"
echo -e "${BLUE}======================================${NC}"

# Functions
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Check Docker Swarm mode
check_swarm() {
    print_info "Checking Docker Swarm status..."
    
    if ! docker info | grep -q "Swarm: active"; then
        print_warning "Docker Swarm is not initialized"
        read -p "Initialize Docker Swarm? (y/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            docker swarm init --advertise-addr $(hostname -I | awk '{print $1}')
            print_status "Docker Swarm initialized"
        else
            print_error "Docker Swarm is required for deployment"
            exit 1
        fi
    else
        print_status "Docker Swarm is active"
    fi
}

# Deploy stack
deploy_stack() {
    print_info "Deploying Llama Swarm stack..."
    
    # Create necessary directories
    mkdir -p monitoring web-interface api-server
    
    # Deploy the stack
    docker stack deploy -c docker-swarm.yml llama-swarm
    
    print_status "Stack deployment initiated"
}

# Wait for services
wait_for_services() {
    print_info "Waiting for services to start..."
    
    local max_attempts=30
    local attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        local ready=$(docker service ls --filter "label=com.docker.stack.namespace=llama-swarm" --format "{{.Replicas}}" | grep -c "^[1-9]/[1-9]" || true)
        local total=$(docker service ls --filter "label=com.docker.stack.namespace=llama-swarm" | wc -l)
        
        if [ "$ready" -eq "$total" ] && [ "$total" -gt 0 ]; then
            print_status "All services are running"
            return 0
        fi
        
        echo -ne "\rWaiting for services... ($ready/$total ready) "
        sleep 5
        ((attempt++))
    done
    
    print_warning "Some services may not be fully ready"
}

# Scale Ollama nodes
scale_nodes() {
    local count=${1:-3}
    print_info "Scaling Ollama nodes to $count replicas..."
    
    docker service scale llama-swarm_ollama=$count
    print_status "Scaling command issued"
}

# Show service status
show_status() {
    print_info "Service Status:"
    echo
    docker service ls --filter "label=com.docker.stack.namespace=llama-swarm"
    echo
    print_info "Node Distribution:"
    docker service ps llama-swarm_ollama --format "table {{.Node}}\t{{.CurrentState}}"
}

# Load test models
load_models() {
    print_info "Loading AI models into Ollama nodes..."
    
    # Get Ollama container IDs
    local containers=$(docker ps --filter "label=com.docker.swarm.service.name=llama-swarm_ollama" -q)
    
    if [ -z "$containers" ]; then
        print_warning "No Ollama containers found"
        return
    fi
    
    # Load a small model for testing
    for container in $containers; do
        print_info "Loading model in container $container..."
        docker exec $container ollama pull llama2:7b 2>/dev/null || true
    done
    
    print_status "Model loading initiated"
}

# Test distributed inference
test_inference() {
    print_info "Testing distributed inference..."
    
    # Test WebSocket connection
    if command -v wscat &> /dev/null; then
        echo '{"type":"inference","content":"Hello, test message","model":"llama2","settings":{"streaming":false,"maxTokens":50}}' | \
        wscat -c ws://localhost:13000/chat -x exit 2>/dev/null || \
        print_warning "WebSocket test failed - wscat not available"
    fi
    
    # Test REST API
    response=$(curl -s -X GET http://localhost:13000/api/health 2>/dev/null || echo "{}")
    
    if echo "$response" | grep -q "healthy"; then
        print_status "API is healthy"
        echo "$response" | python3 -m json.tool 2>/dev/null || echo "$response"
    else
        print_warning "API health check failed"
    fi
}

# Display URLs
display_urls() {
    print_info "Service URLs:"
    echo
    echo -e "${GREEN}🌐 Main Services:${NC}"
    echo -e "  API Gateway:       http://localhost:13000"
    echo -e "  Web Interface:     http://localhost:13080"
    echo -e "  WebSocket:         ws://localhost:13000/chat"
    echo
    echo -e "${BLUE}📊 Monitoring:${NC}"
    echo -e "  Prometheus:        http://localhost:13090"
    echo -e "  Grafana:           http://localhost:13091"
    echo -e "  Node Metrics:      http://localhost:13092"
    echo
}

# Cleanup function
cleanup() {
    print_info "Removing Llama Swarm stack..."
    docker stack rm llama-swarm
    print_status "Stack removed"
}

# Main execution
case "${1:-deploy}" in
    "deploy")
        check_swarm
        deploy_stack
        wait_for_services
        show_status
        display_urls
        print_status "Deployment complete!"
        ;;
    "scale")
        scale_nodes ${2:-3}
        ;;
    "status")
        show_status
        ;;
    "test")
        test_inference
        ;;
    "models")
        load_models
        ;;
    "urls")
        display_urls
        ;;
    "cleanup")
        cleanup
        ;;
    "help")
        echo "Usage: $0 [command] [options]"
        echo
        echo "Commands:"
        echo "  deploy    - Deploy the swarm stack (default)"
        echo "  scale N   - Scale Ollama nodes to N replicas"
        echo "  status    - Show service status"
        echo "  test      - Test distributed inference"
        echo "  models    - Load test models"
        echo "  urls      - Display service URLs"
        echo "  cleanup   - Remove the swarm stack"
        echo "  help      - Show this help message"
        ;;
    *)
        print_error "Unknown command: $1"
        echo "Use '$0 help' for usage information"
        exit 1
        ;;
esac
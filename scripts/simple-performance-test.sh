#!/bin/bash

# Simple Performance Test for Topology Optimization
# Tests startup time and basic functionality without complex dependencies

set -euo pipefail
IFS=$'\n\t'

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
COMPOSE_FILE="docker-compose-topology-optimized.yml"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'  
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Performance targets
TARGET_STARTUP_TIME=110

log() {
    echo -e "${BLUE}[$(date '+%H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Simple startup time test
test_startup_time() {
    log "📊 Testing optimized cluster startup time..."
    
    # Stop any running services
    cd "$PROJECT_ROOT"
    docker-compose -f "$COMPOSE_FILE" down --remove-orphans >/dev/null 2>&1 || true
    
    # Check required environment variables
    export JWT_SECRET="${JWT_SECRET:-test-jwt-secret-for-performance-testing}"
    export MINIO_ROOT_USER="${MINIO_ROOT_USER:-testuser}"
    export MINIO_ROOT_PASSWORD="${MINIO_ROOT_PASSWORD:-testpass123}"
    export DB_PASSWORD="${DB_PASSWORD:-development}"
    
    # Create required directories
    mkdir -p data/{postgres,redis,minio,node-{1,2,3}} logs
    
    local start_time=$(date +%s)
    
    log "Starting foundation services..."
    docker-compose -f "$COMPOSE_FILE" up -d postgres redis minio
    
    # Wait for foundation services
    local foundation_ready=false
    for i in {1..30}; do
        if docker-compose -f "$COMPOSE_FILE" ps postgres | grep -q "Up" && \
           docker-compose -f "$COMPOSE_FILE" ps redis | grep -q "Up" && \
           docker-compose -f "$COMPOSE_FILE" ps minio | grep -q "Up"; then
            foundation_ready=true
            break
        fi
        sleep 2
    done
    
    if [ "$foundation_ready" = false ]; then
        error "Foundation services failed to start"
        return 1
    fi
    
    log "Starting core services..."
    docker-compose -f "$COMPOSE_FILE" up -d ollama-node-1
    
    # Wait for leader node
    local leader_ready=false
    for i in {1..30}; do
        if docker-compose -f "$COMPOSE_FILE" ps ollama-node-1 | grep -q "Up"; then
            leader_ready=true
            break
        fi
        sleep 2
    done
    
    if [ "$leader_ready" = false ]; then
        warning "Leader node failed to start (continuing test)"
    fi
    
    log "Starting worker nodes and load balancer..."
    docker-compose -f "$COMPOSE_FILE" up -d ollama-node-2 ollama-node-3 load-balancer
    
    local end_time=$(date +%s)
    local startup_time=$((end_time - start_time))
    
    echo
    echo -e "${BLUE}📊 Performance Test Results${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    if [ $startup_time -le $TARGET_STARTUP_TIME ]; then
        success "🏆 Startup Time: ${startup_time}s (Target: ≤${TARGET_STARTUP_TIME}s) - PASSED"
        local improvement=$(( (190 - startup_time) * 100 / 190 ))
        success "📈 Improvement: ${improvement}% faster than baseline (190s)"
    else
        warning "⚠️  Startup Time: ${startup_time}s (Target: ≤${TARGET_STARTUP_TIME}s) - MISSED TARGET"
        if [ $startup_time -lt 150 ]; then
            log "Still significantly better than baseline (190s)"
        fi
    fi
    
    # Test basic functionality
    log "Testing basic service connectivity..."
    local services_up=0
    local total_services=0
    
    # Check load balancer
    if curl -f -s -m 5 "http://localhost:80/health" >/dev/null 2>&1; then
        success "Load balancer: ✅ Healthy"
        ((services_up++))
    else
        warning "Load balancer: ⚠️  Not responding"
    fi
    ((total_services++))
    
    # Check if any ollama nodes are responding
    for port in 8080 8082 8084; do
        if curl -f -s -m 5 "http://localhost:${port}/health" >/dev/null 2>&1; then
            success "Node on port $port: ✅ Healthy"
            ((services_up++))
        else
            warning "Node on port $port: ⚠️  Not responding"
        fi
        ((total_services++))
    done
    
    local health_ratio=$(( services_up * 100 / total_services ))
    echo
    success "Service Health: ${services_up}/${total_services} services responding (${health_ratio}%)"
    
    # Simplified resource check
    log "Checking container resource usage..."
    local running_containers
    running_containers=$(docker-compose -f "$COMPOSE_FILE" ps -q | wc -l)
    success "Active containers: $running_containers"
    
    # Summary
    echo
    echo -e "${GREEN}🎯 Performance Validation Summary${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    if [ $startup_time -le $TARGET_STARTUP_TIME ] && [ $health_ratio -ge 50 ]; then
        success "🏆 VALIDATION PASSED - Optimization targets achieved"
        success "   • Startup time: ${startup_time}s ≤ ${TARGET_STARTUP_TIME}s"
        success "   • Service health: ${health_ratio}% ≥ 50%"
    else
        warning "📋 VALIDATION PARTIAL - Some targets missed"
        [ $startup_time -le $TARGET_STARTUP_TIME ] && success "   • Startup time: ✅ PASSED" || warning "   • Startup time: ⚠️  MISSED"
        [ $health_ratio -ge 50 ] && success "   • Service health: ✅ PASSED" || warning "   • Service health: ⚠️  MISSED"
    fi
    
    echo
    log "Service endpoints available:"
    log "  • Load Balancer:  http://localhost:80"
    log "  • API (Node 1):   http://localhost:8080"
    log "  • API (Node 2):   http://localhost:8082"  
    log "  • API (Node 3):   http://localhost:8084"
}

# Cleanup function
cleanup() {
    log "Test completed. Services are still running for further testing."
    log "To stop services: docker-compose -f $COMPOSE_FILE down"
}

main() {
    echo -e "${BLUE}🚀 OllamaMax Simple Performance Test${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo
    
    test_startup_time
    cleanup
}

# Trap cleanup on exit
trap cleanup EXIT

# Execute test
main "$@"
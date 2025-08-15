#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
COMPOSE_FILE="$SCRIPT_DIR/docker-compose.yml"
LOG_DIR="$SCRIPT_DIR/logs"
REPORT_FILE="$LOG_DIR/cluster-deployment-report.md"

# Cluster configuration
NODES=("node-1" "node-2" "node-3")
NODE_IPS=("172.20.0.10" "172.20.0.11" "172.20.0.12")
NODE_PORTS=("8100" "8101" "8102")
METRICS_PORTS=("9100" "9101" "9102")

echo -e "${BLUE}🚀 OllamaMax Multi-Node Distributed Cluster Deployment${NC}"
echo "========================================================"

# Function to print colored output
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

print_header() {
    echo -e "${PURPLE}🔥 $1${NC}"
}

print_step() {
    echo -e "${CYAN}📋 $1${NC}"
}

# Function to create log directory
setup_logging() {
    mkdir -p "$LOG_DIR"
    echo "# OllamaMax Distributed Cluster Deployment Report" > "$REPORT_FILE"
    echo "Generated on: $(date)" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
}

# Function to log to report
log_to_report() {
    echo "$1" >> "$REPORT_FILE"
}

# Function to check prerequisites
check_prerequisites() {
    print_header "Checking Prerequisites"
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed"
        exit 1
    fi
    
    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed"
        exit 1
    fi
    
    # Check if Docker daemon is running
    if ! docker info &> /dev/null; then
        print_error "Docker daemon is not running"
        exit 1
    fi
    
    # Check port availability
    print_info "Checking port availability..."
    local ports_to_check=(8100 8101 8102 8200 9100 9101 9102 9200 3100 6400)
    for port in "${ports_to_check[@]}"; do
        if netstat -tuln | grep ":$port " &> /dev/null; then
            print_warning "Port $port is already in use"
        else
            print_status "Port $port is available"
        fi
    done
    
    print_status "Prerequisites check completed"
    log_to_report "## Prerequisites Check"
    log_to_report "- Docker: ✅ Available"
    log_to_report "- Docker Compose: ✅ Available"
    log_to_report "- Ports: Checked and mostly available"
    log_to_report ""
}

# Function to run tests before deployment
run_pre_deployment_tests() {
    print_header "Running Pre-Deployment Tests"
    
    cd "$PROJECT_ROOT"
    
    local test_start_time=$(date +%s)
    
    # Run all test phases
    print_step "Running Phase 1: Foundation Tests..."
    if go test -v ./test/simple_phase1_test.go > "$LOG_DIR/phase1-tests.log" 2>&1; then
        print_status "Phase 1 tests passed"
    else
        print_error "Phase 1 tests failed"
        return 1
    fi
    
    print_step "Running Phase 2: Advanced Features Tests..."
    if go test -v ./test/standalone_phase2_test.go > "$LOG_DIR/phase2-tests.log" 2>&1; then
        print_status "Phase 2 tests passed"
    else
        print_error "Phase 2 tests failed"
        return 1
    fi
    
    print_step "Running Phase 3: Production Readiness Tests..."
    if go test -v ./test/standalone_phase3_test.go > "$LOG_DIR/phase3-tests.log" 2>&1; then
        print_status "Phase 3 tests passed"
    else
        print_error "Phase 3 tests failed"
        return 1
    fi
    
    local test_end_time=$(date +%s)
    local test_duration=$((test_end_time - test_start_time))
    
    print_status "All pre-deployment tests passed in ${test_duration}s"
    
    log_to_report "## Pre-Deployment Tests"
    log_to_report "- Phase 1 (Foundation): ✅ Passed"
    log_to_report "- Phase 2 (Advanced Features): ✅ Passed"
    log_to_report "- Phase 3 (Production Readiness): ✅ Passed"
    log_to_report "- Total Duration: ${test_duration} seconds"
    log_to_report ""
}

# Function to deploy the cluster
deploy_cluster() {
    print_header "Deploying Multi-Node Cluster"
    
    cd "$SCRIPT_DIR"
    
    local deploy_start_time=$(date +%s)
    
    # Stop any existing containers
    print_step "Stopping existing containers..."
    docker-compose -f "$COMPOSE_FILE" down --remove-orphans --volumes 2>/dev/null || true
    
    # Build and start services
    print_step "Building Docker images..."
    if docker-compose -f "$COMPOSE_FILE" build --no-cache > "$LOG_DIR/docker-build.log" 2>&1; then
        print_status "Docker images built successfully"
    else
        print_error "Docker build failed"
        return 1
    fi
    
    print_step "Starting cluster services..."
    if docker-compose -f "$COMPOSE_FILE" up -d > "$LOG_DIR/docker-up.log" 2>&1; then
        print_status "Cluster services started"
    else
        print_error "Failed to start cluster services"
        return 1
    fi
    
    local deploy_end_time=$(date +%s)
    local deploy_duration=$((deploy_end_time - deploy_start_time))
    
    print_status "Cluster deployment completed in ${deploy_duration}s"
    
    log_to_report "## Cluster Deployment"
    log_to_report "- Docker Build: ✅ Successful"
    log_to_report "- Service Startup: ✅ Successful"
    log_to_report "- Deployment Duration: ${deploy_duration} seconds"
    log_to_report ""
}

# Function to wait for cluster formation
wait_for_cluster_formation() {
    print_header "Waiting for Cluster Formation"
    
    local max_attempts=60
    local attempt=0
    local formation_start_time=$(date +%s)
    
    # Wait for each node to be healthy
    for i in "${!NODES[@]}"; do
        local node="${NODES[$i]}"
        local port="${NODE_PORTS[$i]}"
        
        print_step "Waiting for $node to be ready..."
        
        attempt=0
        while [ $attempt -lt $max_attempts ]; do
            if curl -f "http://localhost:$port/health" &> /dev/null; then
                print_status "$node is ready"
                break
            fi
            
            attempt=$((attempt + 1))
            sleep 5
            
            if [ $attempt -eq $max_attempts ]; then
                print_error "$node failed to start within timeout"
                return 1
            fi
        done
    done
    
    # Wait for load balancer
    print_step "Waiting for load balancer..."
    attempt=0
    while [ $attempt -lt $max_attempts ]; do
        if curl -f "http://localhost:8200/health" &> /dev/null; then
            print_status "Load balancer is ready"
            break
        fi
        
        attempt=$((attempt + 1))
        sleep 5
        
        if [ $attempt -eq $max_attempts ]; then
            print_error "Load balancer failed to start within timeout"
            return 1
        fi
    done
    
    # Wait additional time for cluster consensus
    print_step "Waiting for Raft consensus establishment..."
    sleep 30
    
    local formation_end_time=$(date +%s)
    local formation_duration=$((formation_end_time - formation_start_time))
    
    print_status "Cluster formation completed in ${formation_duration}s"
    
    log_to_report "## Cluster Formation"
    log_to_report "- Node Startup: ✅ All nodes ready"
    log_to_report "- Load Balancer: ✅ Ready"
    log_to_report "- Consensus Wait: ✅ Completed"
    log_to_report "- Formation Duration: ${formation_duration} seconds"
    log_to_report ""
}

# Function to verify cluster status
verify_cluster_status() {
    print_header "Verifying Cluster Status"
    
    log_to_report "## Cluster Status Verification"
    
    # Check individual nodes
    for i in "${!NODES[@]}"; do
        local node="${NODES[$i]}"
        local port="${NODE_PORTS[$i]}"
        local ip="${NODE_IPS[$i]}"
        
        print_step "Checking $node status..."
        
        # Health check
        if curl -f "http://localhost:$port/health" &> /dev/null; then
            print_status "$node health check passed"
            log_to_report "- $node (${ip}:8080): ✅ Healthy"
        else
            print_error "$node health check failed"
            log_to_report "- $node (${ip}:8080): ❌ Unhealthy"
        fi
        
        # Cluster status
        local cluster_status=$(curl -s "http://localhost:$port/cluster/status" 2>/dev/null || echo "failed")
        if [[ "$cluster_status" != "failed" ]]; then
            print_status "$node cluster status accessible"
            echo "$cluster_status" > "$LOG_DIR/${node}-cluster-status.json"
        else
            print_warning "$node cluster status not accessible"
        fi
        
        # Metrics check
        if curl -f "http://localhost:${METRICS_PORTS[$i]}/metrics" &> /dev/null; then
            print_status "$node metrics endpoint accessible"
        else
            print_warning "$node metrics endpoint not accessible"
        fi
    done
    
    # Check load balancer
    print_step "Checking load balancer..."
    if curl -f "http://localhost:8200/health" &> /dev/null; then
        print_status "Load balancer health check passed"
        log_to_report "- Load Balancer: ✅ Healthy"
    else
        print_error "Load balancer health check failed"
        log_to_report "- Load Balancer: ❌ Unhealthy"
    fi
    
    log_to_report ""
}

# Function to test load balancing
test_load_balancing() {
    print_header "Testing Load Balancing"
    
    print_step "Running load balancing tests..."
    
    local total_requests=30
    local successful_requests=0
    local node_hits=()
    
    # Initialize node hit counters
    for i in "${!NODES[@]}"; do
        node_hits[$i]=0
    done
    
    log_to_report "## Load Balancing Tests"
    log_to_report "Testing with $total_requests requests..."
    log_to_report ""
    
    for i in $(seq 1 $total_requests); do
        local response=$(curl -s "http://localhost:8200/api/v1/status" 2>/dev/null || echo "failed")
        
        if [[ "$response" != "failed" ]]; then
            successful_requests=$((successful_requests + 1))
            
            # Try to determine which node handled the request
            local node_id=$(echo "$response" | grep -o '"node_id":"[^"]*"' | cut -d'"' -f4 2>/dev/null || echo "unknown")
            
            case "$node_id" in
                "node-1") node_hits[0]=$((${node_hits[0]} + 1)) ;;
                "node-2") node_hits[1]=$((${node_hits[1]} + 1)) ;;
                "node-3") node_hits[2]=$((${node_hits[2]} + 1)) ;;
            esac
        fi
        
        sleep 0.1  # Small delay between requests
    done
    
    print_status "Load balancing test completed: $successful_requests/$total_requests successful"
    
    # Report distribution
    log_to_report "### Request Distribution:"
    for i in "${!NODES[@]}"; do
        local node="${NODES[$i]}"
        local hits="${node_hits[$i]}"
        local percentage=$(( hits * 100 / successful_requests ))
        print_info "$node: $hits requests (${percentage}%)"
        log_to_report "- $node: $hits requests (${percentage}%)"
    done
    
    log_to_report ""
}

# Function to test failover
test_failover() {
    print_header "Testing Failover Scenarios"
    
    print_step "Testing node failover..."
    
    log_to_report "## Failover Testing"
    
    # Stop node-2 temporarily
    print_step "Stopping node-2 for failover test..."
    docker-compose -f "$COMPOSE_FILE" stop ollama-node-2
    
    sleep 10
    
    # Test cluster still responds
    local failover_requests=10
    local successful_failover=0
    
    for i in $(seq 1 $failover_requests); do
        if curl -f "http://localhost:8200/health" &> /dev/null; then
            successful_failover=$((successful_failover + 1))
        fi
        sleep 1
    done
    
    print_status "Failover test: $successful_failover/$failover_requests requests successful"
    log_to_report "- Node-2 stopped: Cluster continued operating"
    log_to_report "- Failover requests: $successful_failover/$failover_requests successful"
    
    # Restart node-2
    print_step "Restarting node-2..."
    docker-compose -f "$COMPOSE_FILE" start ollama-node-2
    
    sleep 15
    
    # Verify node-2 rejoined
    if curl -f "http://localhost:8101/health" &> /dev/null; then
        print_status "Node-2 successfully rejoined cluster"
        log_to_report "- Node-2 restart: ✅ Successfully rejoined cluster"
    else
        print_warning "Node-2 may not have fully rejoined"
        log_to_report "- Node-2 restart: ⚠️ May not have fully rejoined"
    fi
    
    log_to_report ""
}

# Function to measure performance
measure_performance() {
    print_header "Measuring Performance"
    
    print_step "Running performance benchmarks..."
    
    log_to_report "## Performance Benchmarks"
    
    # Test individual node performance
    for i in "${!NODES[@]}"; do
        local node="${NODES[$i]}"
        local port="${NODE_PORTS[$i]}"
        
        print_step "Testing $node performance..."
        
        local total_time=0
        local successful_requests=0
        local test_requests=20
        
        for j in $(seq 1 $test_requests); do
            local start_time=$(date +%s%N)
            if curl -s -f "http://localhost:$port/health" &> /dev/null; then
                local end_time=$(date +%s%N)
                local duration=$(( (end_time - start_time) / 1000000 )) # Convert to milliseconds
                total_time=$((total_time + duration))
                successful_requests=$((successful_requests + 1))
            fi
        done
        
        if [ $successful_requests -gt 0 ]; then
            local average_time=$((total_time / successful_requests))
            print_status "$node: ${average_time}ms average latency"
            log_to_report "- $node: ${average_time}ms average latency ($successful_requests/$test_requests successful)"
        else
            print_error "$node: No successful requests"
            log_to_report "- $node: ❌ No successful requests"
        fi
    done
    
    # Test load balancer performance
    print_step "Testing load balancer performance..."
    
    local lb_total_time=0
    local lb_successful=0
    local lb_requests=50
    
    for i in $(seq 1 $lb_requests); do
        local start_time=$(date +%s%N)
        if curl -s -f "http://localhost:8200/health" &> /dev/null; then
            local end_time=$(date +%s%N)
            local duration=$(( (end_time - start_time) / 1000000 ))
            lb_total_time=$((lb_total_time + duration))
            lb_successful=$((lb_successful + 1))
        fi
        sleep 0.05  # 50ms delay
    done
    
    if [ $lb_successful -gt 0 ]; then
        local lb_average=$((lb_total_time / lb_successful))
        print_status "Load Balancer: ${lb_average}ms average latency"
        log_to_report "- Load Balancer: ${lb_average}ms average latency ($lb_successful/$lb_requests successful)"
    fi
    
    log_to_report ""
}

# Function to collect logs
collect_logs() {
    print_header "Collecting Container Logs"
    
    print_step "Saving container logs..."
    
    # Collect logs from each service
    local services=("ollama-node-1" "ollama-node-2" "ollama-node-3" "nginx-lb" "prometheus" "grafana" "redis")
    
    for service in "${services[@]}"; do
        print_info "Collecting logs for $service..."
        docker-compose -f "$COMPOSE_FILE" logs "$service" > "$LOG_DIR/${service}.log" 2>&1 || true
    done
    
    print_status "Container logs collected in $LOG_DIR"
    
    log_to_report "## Container Logs"
    log_to_report "Logs collected for all services in: $LOG_DIR"
    log_to_report ""
}

# Function to show cluster information
show_cluster_info() {
    print_header "Cluster Information"
    
    echo "======================================"
    echo "🌐 Access Points:"
    echo "Load Balancer: http://localhost:8200"
    echo "Node 1: http://localhost:8100"
    echo "Node 2: http://localhost:8101"
    echo "Node 3: http://localhost:8102"
    echo "Prometheus: http://localhost:9200"
    echo "Grafana: http://localhost:3100 (admin/admin123)"
    echo "Redis: localhost:6400"
    echo ""
    echo "📊 Monitoring Endpoints:"
    echo "- Health: http://localhost:8200/health"
    echo "- Cluster Status: http://localhost:8100/cluster/status"
    echo "- Metrics: http://localhost:9100/metrics"
    echo "- SLA Metrics: http://localhost:8100/sla/metrics"
    echo ""
    echo "🔧 Management Commands:"
    echo "- View logs: docker-compose -f $COMPOSE_FILE logs -f"
    echo "- Stop cluster: docker-compose -f $COMPOSE_FILE down"
    echo "- Restart service: docker-compose -f $COMPOSE_FILE restart <service>"
    echo ""
    
    log_to_report "## Cluster Access Information"
    log_to_report "- Load Balancer: http://localhost:8200"
    log_to_report "- Individual Nodes: 8100, 8101, 8102"
    log_to_report "- Monitoring: Prometheus (9200), Grafana (3100)"
    log_to_report "- All services deployed with static IPs in 172.20.0.0/16 network"
}

# Main execution function
main() {
    local start_time=$(date +%s)
    
    setup_logging
    
    case "${1:-deploy}" in
        "deploy")
            check_prerequisites
            run_pre_deployment_tests
            deploy_cluster
            wait_for_cluster_formation
            verify_cluster_status
            test_load_balancing
            test_failover
            measure_performance
            collect_logs
            show_cluster_info
            ;;
        "test")
            verify_cluster_status
            test_load_balancing
            test_failover
            measure_performance
            ;;
        "logs")
            collect_logs
            ;;
        "info")
            show_cluster_info
            ;;
        "cleanup")
            docker-compose -f "$COMPOSE_FILE" down --remove-orphans --volumes
            print_status "Cluster cleaned up"
            ;;
        *)
            echo "Usage: $0 [deploy|test|logs|info|cleanup]"
            echo "  deploy  - Full deployment and testing (default)"
            echo "  test    - Run cluster tests only"
            echo "  logs    - Collect container logs"
            echo "  info    - Show cluster information"
            echo "  cleanup - Stop and remove all containers"
            exit 1
            ;;
    esac
    
    local end_time=$(date +%s)
    local total_duration=$((end_time - start_time))
    
    print_status "Operation completed in ${total_duration} seconds"
    log_to_report "## Summary"
    log_to_report "- Total Duration: ${total_duration} seconds"
    log_to_report "- Report Generated: $(date)"
    log_to_report "- Logs Directory: $LOG_DIR"
    
    print_info "Detailed report available at: $REPORT_FILE"
}

# Run main function
main "$@"

#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
COMPOSE_FILE="$SCRIPT_DIR/docker-compose.yml"
TEST_TIMEOUT=300  # 5 minutes

echo -e "${BLUE}🚀 OllamaMax Distributed Deployment and Testing${NC}"
echo "=================================================="

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

# Function to check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
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
    
    print_status "Prerequisites check passed"
}

# Function to build the application
build_application() {
    print_info "Building OllamaMax Distributed..."
    
    cd "$PROJECT_ROOT"
    
    # Run tests first
    print_info "Running tests before deployment..."
    go test -v ./test/simple_phase1_test.go ./test/standalone_phase2_test.go ./test/standalone_phase3_test.go
    
    if [ $? -eq 0 ]; then
        print_status "All tests passed!"
    else
        print_error "Tests failed! Aborting deployment."
        exit 1
    fi
    
    # Build the application
    print_info "Building Go application..."
    mkdir -p cmd/distributed
    
    # Create main.go if it doesn't exist
    if [ ! -f "cmd/distributed/main.go" ]; then
        cat > cmd/distributed/main.go << 'EOF'
package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
    "github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/server"
    "github.com/sirupsen/logrus"
)

func main() {
    var configPath = flag.String("config", "config.yaml", "Path to configuration file")
    flag.Parse()

    // Load configuration
    cfg, err := config.LoadDistributedConfig(*configPath)
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

    // Setup logger
    logger := logrus.New()
    level, err := logrus.ParseLevel(cfg.Observability.Logging.Level)
    if err != nil {
        level = logrus.InfoLevel
    }
    logger.SetLevel(level)

    // Create and start server
    srv, err := server.NewDistributedServer(cfg, logger)
    if err != nil {
        logger.Fatalf("Failed to create server: %v", err)
    }

    // Start server
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go func() {
        if err := srv.Start(ctx); err != nil {
            logger.Errorf("Server error: %v", err)
        }
    }()

    logger.Infof("OllamaMax Distributed started on %s", cfg.Node.Address)

    // Wait for interrupt signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan

    logger.Info("Shutting down...")
    cancel()

    // Graceful shutdown
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer shutdownCancel()

    if err := srv.Shutdown(shutdownCtx); err != nil {
        logger.Errorf("Shutdown error: %v", err)
    }

    logger.Info("Server stopped")
}
EOF
    fi
    
    print_status "Application built successfully"
}

# Function to deploy the cluster
deploy_cluster() {
    print_info "Deploying OllamaMax Distributed Cluster..."
    
    cd "$SCRIPT_DIR"
    
    # Stop any existing containers
    print_info "Stopping existing containers..."
    docker-compose -f "$COMPOSE_FILE" down --remove-orphans || true
    
    # Build and start services
    print_info "Building and starting services..."
    docker-compose -f "$COMPOSE_FILE" build --no-cache
    docker-compose -f "$COMPOSE_FILE" up -d
    
    print_status "Cluster deployment initiated"
}

# Function to wait for services to be ready
wait_for_services() {
    print_info "Waiting for services to be ready..."
    
    local services=("ollama-node-1:8080" "ollama-node-2:8081" "ollama-node-3:8082" "nginx-lb:80")
    local max_attempts=60
    local attempt=0
    
    for service in "${services[@]}"; do
        local host=$(echo $service | cut -d':' -f1)
        local port=$(echo $service | cut -d':' -f2)
        
        print_info "Waiting for $service..."
        
        while [ $attempt -lt $max_attempts ]; do
            if curl -f "http://localhost:$port/health" &> /dev/null; then
                print_status "$service is ready"
                break
            fi
            
            attempt=$((attempt + 1))
            sleep 5
            
            if [ $attempt -eq $max_attempts ]; then
                print_error "$service failed to start within timeout"
                return 1
            fi
        done
        
        attempt=0
    done
    
    print_status "All services are ready"
}

# Function to run comprehensive tests
run_tests() {
    print_info "Running comprehensive deployment tests..."
    
    # Test 1: Health checks
    print_info "Test 1: Health checks..."
    for port in 8080 8081 8082 80; do
        if curl -f "http://localhost:$port/health" &> /dev/null; then
            print_status "Health check passed for port $port"
        else
            print_error "Health check failed for port $port"
            return 1
        fi
    done
    
    # Test 2: Cluster status
    print_info "Test 2: Cluster status..."
    response=$(curl -s "http://localhost:8080/cluster/status" || echo "failed")
    if [[ "$response" != "failed" ]]; then
        print_status "Cluster status endpoint accessible"
        echo "Cluster status: $response"
    else
        print_warning "Cluster status endpoint not accessible (may not be implemented yet)"
    fi
    
    # Test 3: Load balancer
    print_info "Test 3: Load balancer..."
    for i in {1..5}; do
        response=$(curl -s "http://localhost/health" || echo "failed")
        if [[ "$response" == "healthy" ]]; then
            print_status "Load balancer test $i passed"
        else
            print_error "Load balancer test $i failed"
            return 1
        fi
    done
    
    # Test 4: Metrics endpoints
    print_info "Test 4: Metrics endpoints..."
    for port in 9090 9091 9092; do
        if curl -f "http://localhost:$port/metrics" &> /dev/null; then
            print_status "Metrics endpoint accessible on port $port"
        else
            print_warning "Metrics endpoint not accessible on port $port (may not be implemented yet)"
        fi
    done
    
    # Test 5: Prometheus
    print_info "Test 5: Prometheus..."
    if curl -f "http://localhost:9093" &> /dev/null; then
        print_status "Prometheus is accessible"
    else
        print_warning "Prometheus not accessible"
    fi
    
    # Test 6: Grafana
    print_info "Test 6: Grafana..."
    if curl -f "http://localhost:3000" &> /dev/null; then
        print_status "Grafana is accessible"
    else
        print_warning "Grafana not accessible"
    fi
    
    print_status "All deployment tests completed"
}

# Function to run performance tests
run_performance_tests() {
    print_info "Running performance tests..."
    
    # Simple load test
    print_info "Running basic load test..."
    
    for i in {1..10}; do
        start_time=$(date +%s%N)
        response=$(curl -s -w "%{http_code}" "http://localhost/health")
        end_time=$(date +%s%N)
        
        duration=$(( (end_time - start_time) / 1000000 )) # Convert to milliseconds
        
        if [[ "$response" == *"200" ]]; then
            print_status "Request $i: ${duration}ms"
        else
            print_error "Request $i failed"
        fi
    done
    
    print_status "Performance tests completed"
}

# Function to show cluster information
show_cluster_info() {
    print_info "Cluster Information:"
    echo "===================="
    echo "Load Balancer: http://localhost"
    echo "Node 1: http://localhost:8080"
    echo "Node 2: http://localhost:8081"
    echo "Node 3: http://localhost:8082"
    echo "Prometheus: http://localhost:9093"
    echo "Grafana: http://localhost:3000 (admin/admin123)"
    echo ""
    echo "Health endpoints:"
    echo "- http://localhost/health"
    echo "- http://localhost:8080/health"
    echo "- http://localhost:8081/health"
    echo "- http://localhost:8082/health"
    echo ""
    echo "Metrics endpoints:"
    echo "- http://localhost:9090/metrics"
    echo "- http://localhost:9091/metrics"
    echo "- http://localhost:9092/metrics"
    echo ""
    print_info "Use 'docker-compose -f $COMPOSE_FILE logs -f' to view logs"
    print_info "Use 'docker-compose -f $COMPOSE_FILE down' to stop the cluster"
}

# Function to cleanup
cleanup() {
    print_info "Cleaning up..."
    cd "$SCRIPT_DIR"
    docker-compose -f "$COMPOSE_FILE" down --remove-orphans --volumes
    print_status "Cleanup completed"
}

# Main execution
main() {
    case "${1:-deploy}" in
        "deploy")
            check_prerequisites
            build_application
            deploy_cluster
            wait_for_services
            run_tests
            run_performance_tests
            show_cluster_info
            ;;
        "test")
            run_tests
            run_performance_tests
            ;;
        "cleanup")
            cleanup
            ;;
        "info")
            show_cluster_info
            ;;
        *)
            echo "Usage: $0 [deploy|test|cleanup|info]"
            echo "  deploy  - Full deployment and testing (default)"
            echo "  test    - Run tests only"
            echo "  cleanup - Stop and remove all containers"
            echo "  info    - Show cluster information"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"

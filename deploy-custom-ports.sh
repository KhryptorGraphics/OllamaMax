#!/bin/bash

# OllamaMax - Custom Port Deployment Script
# Deploys Docker services on ports 12925-12935

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.custom-ports.yml"
ENV_FILE=".env.custom-ports"
PROJECT_NAME="ollamamax-custom"

echo -e "${BLUE}🚀 OllamaMax - Custom Port Deployment (13000-13300)${NC}"
echo -e "${BLUE}=====================================================${NC}"

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

# Check if Docker is running
check_docker() {
    print_info "Checking Docker installation..."
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi

    if ! docker info &> /dev/null; then
        print_error "Docker is not running. Please start Docker first."
        exit 1
    fi
    
    print_status "Docker is running"
}

# Check port availability
check_ports() {
    print_info "Checking port availability (13000-13009)..."
    
    PORTS=(13000 13001 13002 13003 13004 13005 13006 13007 13008 13009)
    OCCUPIED_PORTS=()
    
    for port in "${PORTS[@]}"; do
        if netstat -tuln 2>/dev/null | grep ":$port " >/dev/null || \
           ss -tuln 2>/dev/null | grep ":$port " >/dev/null || \
           lsof -i ":$port" 2>/dev/null >/dev/null; then
            OCCUPIED_PORTS+=($port)
        fi
    done
    
    if [ ${#OCCUPIED_PORTS[@]} -gt 0 ]; then
        print_warning "Some ports are already in use: ${OCCUPIED_PORTS[*]}"
    else
        print_status "All required ports are available"
    fi
}

# Deploy services
deploy_services() {
    print_info "Building and starting services..."
    
    # Stop existing containers
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" down 2>/dev/null || true
    
    # Start core services
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" --env-file "$ENV_FILE" up -d redis
    sleep 5
    
    # Start Ollama engine
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" --env-file "$ENV_FILE" up -d ollama
    sleep 10
    
    # Start BMad dashboard
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" --env-file "$ENV_FILE" up -d bmad-dashboard
    sleep 5
    
    # Start infrastructure
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" --env-file "$ENV_FILE" up -d nginx prometheus grafana
    sleep 5
    
    # Start additional services
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" --env-file "$ENV_FILE" up -d redis-commander minio ollama-webui
    
    print_status "All services started"
}

# Display service URLs
display_urls() {
    print_info "Service URLs:"
    echo ""
    echo -e "${GREEN}🌐 Main Services:${NC}"
    echo -e "  Ollama Engine:         http://localhost:13000"
    echo -e "  Redis Cache:           localhost:13001"
    echo -e "  BMad Dashboard:        http://localhost:13002"
    echo -e "  Load Balancer:         http://localhost:13003"
    echo ""
    echo -e "${BLUE}📊 Monitoring:${NC}"
    echo -e "  Prometheus:            http://localhost:13004"
    echo -e "  Grafana:               http://localhost:13005"
    echo ""
    echo -e "${YELLOW}🔧 Management:${NC}"
    echo -e "  Redis Commander:       http://localhost:13006"
    echo -e "  MinIO API:             http://localhost:13007"
    echo -e "  MinIO Console:         http://localhost:13008"
    echo -e "  Ollama WebUI:          http://localhost:13009"
    echo ""
}

# Main deployment process
main() {
    check_docker
    check_ports
    deploy_services
    
    echo ""
    echo -e "${GREEN}🎉 Deployment completed!${NC}"
    echo ""
    
    display_urls
    
    print_status "OllamaMax is now running on custom ports 13000-13009"
}

# Handle script arguments
case "${1:-deploy}" in
    "deploy")
        main
        ;;
    "stop")
        print_info "Stopping all services..."
        docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" down
        print_status "All services stopped"
        ;;
    "status")
        print_info "Service status:"
        docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" ps
        ;;
    "logs")
        SERVICE=${2:-""}
        if [ -n "$SERVICE" ]; then
            docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" logs -f "$SERVICE"
        else
            docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" logs -f
        fi
        ;;
    "help")
        echo "Usage: $0 [deploy|stop|status|logs|help] [service_name]"
        echo ""
        echo "Commands:"
        echo "  deploy    - Deploy all services (default)"
        echo "  stop      - Stop all services"
        echo "  status    - Show service status"
        echo "  logs      - Show logs (optionally for specific service)"
        echo "  help      - Show this help message"
        ;;
    *)
        print_error "Unknown command: $1"
        echo "Use '$0 help' for usage information"
        exit 1
        ;;
esac
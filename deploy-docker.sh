#!/bin/bash

# Enhanced Ollama Distributed Inference Docker Deployment Script
# Auto-detects CUDA availability and deploys GPU or CPU stack accordingly

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Parse command line arguments
DEPLOYMENT_TYPE="auto"  # auto-detect by default
FORCE_REBUILD=false
PULL_MODELS=false
CUDA_AVAILABLE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --gpu|--cuda)
            DEPLOYMENT_TYPE="gpu"
            shift
            ;;
        --cpu|--no-gpu)
            DEPLOYMENT_TYPE="cpu"
            shift
            ;;
        --auto)
            DEPLOYMENT_TYPE="auto"
            shift
            ;;
        --dev|--development)
            DEPLOYMENT_TYPE="dev"
            shift
            ;;
        --rebuild)
            FORCE_REBUILD=true
            shift
            ;;
        --pull-models)
            PULL_MODELS=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --auto                 Auto-detect CUDA and deploy accordingly (default)"
            echo "  --gpu, --cuda          Force GPU deployment (requires CUDA)"
            echo "  --cpu, --no-gpu        Force CPU-only deployment"
            echo "  --dev, --development   Deploy lightweight development stack"
            echo "  --rebuild              Force rebuild of Docker images"
            echo "  --pull-models          Pull default models after deployment"
            echo "  -h, --help             Show this help message"
            exit 0
            ;;
        *)
            log_error "Unknown option $1"
            exit 1
            ;;
    esac
done

log_info "Starting Ollama Distributed Inference deployment..."

# Auto-detect CUDA availability if deployment type is auto
if [ "$DEPLOYMENT_TYPE" = "auto" ]; then
    log_info "Auto-detecting CUDA availability..."
    
    # Check for NVIDIA runtime
    if command -v nvidia-smi &> /dev/null && nvidia-smi &> /dev/null; then
        log_success "NVIDIA GPU detected with nvidia-smi"
        CUDA_AVAILABLE=true
        DEPLOYMENT_TYPE="gpu"
    elif docker info 2>/dev/null | grep -q nvidia; then
        log_success "NVIDIA Docker runtime detected"
        CUDA_AVAILABLE=true
        DEPLOYMENT_TYPE="gpu"
    elif [ -f /proc/driver/nvidia/version ]; then
        log_success "NVIDIA drivers detected"
        CUDA_AVAILABLE=true
        DEPLOYMENT_TYPE="gpu"
    else
        log_info "No CUDA/GPU support detected, using CPU-only deployment"
        CUDA_AVAILABLE=false
        DEPLOYMENT_TYPE="cpu"
    fi
fi

log_info "Deployment type: $DEPLOYMENT_TYPE"

# Check prerequisites
log_info "Checking prerequisites..."

if ! command -v docker &> /dev/null; then
    log_error "Docker is not installed"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    log_error "Docker Compose is not installed"
    exit 1
fi

# Validate GPU deployment requirements
if [ "$DEPLOYMENT_TYPE" = "gpu" ]; then
    log_info "Validating GPU deployment requirements..."
    
    if ! command -v nvidia-smi &> /dev/null; then
        log_warning "nvidia-smi not found. GPU acceleration may not work."
        read -p "Continue with GPU deployment anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Falling back to CPU deployment..."
            DEPLOYMENT_TYPE="cpu"
        fi
    elif ! docker info 2>/dev/null | grep -q nvidia; then
        log_warning "NVIDIA Docker runtime not detected. Installing nvidia-container-toolkit may be required."
        read -p "Continue with GPU deployment anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Falling back to CPU deployment..."
            DEPLOYMENT_TYPE="cpu"
        fi
    fi
fi

# Stop existing containers
log_info "Stopping existing containers..."
docker-compose -f docker-compose.distributed.yml down 2>/dev/null || true
docker-compose -f docker-compose.dev.yml down 2>/dev/null || true
docker-compose -f docker-compose.gpu.yml down 2>/dev/null || true
docker-compose -f docker-compose.cpu.yml down 2>/dev/null || true

# Clean up old containers if force rebuild
if [ "$FORCE_REBUILD" = true ]; then
    log_info "Force rebuilding Docker images..."
    case "$DEPLOYMENT_TYPE" in
        "gpu")
            docker-compose -f docker-compose.gpu.yml build --no-cache
            ;;
        "cpu")
            docker-compose -f docker-compose.cpu.yml build --no-cache
            ;;
        "dev")
            docker-compose -f docker-compose.dev.yml build --no-cache
            ;;
    esac
fi

# Start the appropriate deployment
log_info "Starting $DEPLOYMENT_TYPE deployment..."

case "$DEPLOYMENT_TYPE" in
    "gpu")
        log_info "Deploying GPU-accelerated distributed stack..."
        docker-compose -f docker-compose.gpu.yml up -d --build
        
        # Wait for services to be healthy
        log_info "Waiting for GPU workers to initialize..."
        sleep 45
        
        # Check service health
        log_info "Checking service health..."
        docker-compose -f docker-compose.gpu.yml ps
        
        log_success "GPU deployment completed!"
        log_info "Services available at:"
        log_info "  - Web Interface: http://localhost:13100"
        log_info "  - API Server: http://localhost:13100/api"
        log_info "  - Ollama Primary (GPU): http://localhost:13000"
        log_info "  - Ollama Worker 2 (GPU): http://localhost:13001"
        log_info "  - Ollama Worker 3 (GPU): http://localhost:13002"
        log_info "  - Prometheus: http://localhost:13092"
        log_info "  - Grafana: http://localhost:13093 (admin/ollama_grafana_pass)"
        log_info "  - Alertmanager: http://localhost:13094"
        log_info "  - MinIO Console: http://localhost:13091 (ollama/ollama_minio_pass)"
        ;;
        
    "cpu")
        log_info "Deploying CPU-only distributed stack..."
        docker-compose -f docker-compose.cpu.yml up -d --build
        
        # Wait for services to be healthy
        log_info "Waiting for CPU workers to initialize..."
        sleep 35
        
        # Check service health
        log_info "Checking service health..."
        docker-compose -f docker-compose.cpu.yml ps
        
        log_success "CPU deployment completed!"
        log_info "Services available at:"
        log_info "  - Web Interface: http://localhost:13100"
        log_info "  - API Server: http://localhost:13100/api"
        log_info "  - Ollama Primary (CPU): http://localhost:13000"
        log_info "  - Ollama Worker 2 (CPU): http://localhost:13001"
        log_info "  - Ollama Worker 3 (CPU): http://localhost:13002"
        log_info "  - Prometheus: http://localhost:13092"
        log_info "  - Grafana: http://localhost:13093 (admin/ollama_grafana_pass)"
        log_info "  - Alertmanager: http://localhost:13094"
        log_info "  - MinIO Console: http://localhost:13091 (ollama/ollama_minio_pass)"
        ;;
        
    "dev")
        log_info "Deploying lightweight development stack..."
        docker-compose -f docker-compose.dev.yml up -d --build
        
        # Wait for services to be ready
        log_info "Waiting for services to be ready..."
        sleep 20
        
        # Check service health
        log_info "Checking service health..."
        docker-compose -f docker-compose.dev.yml ps
        
        log_success "Development deployment completed!"
        log_info "Services available at:"
        log_info "  - Web Interface: http://localhost:13100"
        log_info "  - API Server: http://localhost:13100/api"
        log_info "  - MinIO Console: http://localhost:13191 (dev/dev_minio_pass)"
        log_warning "Note: Ollama workers should be started manually on ports 13000, 13001, 13002"
        ;;
esac

# Pull models if requested
if [ "$PULL_MODELS" = true ]; then
    log_info "Pulling default models..."
    
    if [ "$DEPLOYMENT_TYPE" = "prod" ]; then
        # Pull models to each worker
        docker exec ollama-primary ollama pull tinyllama:latest || log_warning "Failed to pull model to primary"
        docker exec ollama-worker-2 ollama pull tinyllama:latest || log_warning "Failed to pull model to worker 2"
        docker exec ollama-worker-3 ollama pull codellama:7b || log_warning "Failed to pull model to worker 3"
    else
        log_info "For development mode, pull models manually to your local Ollama instances"
    fi
fi

# Display final status
log_success "Deployment complete!"
log_info "To view logs: docker-compose -f docker-compose.$([ "$DEPLOYMENT_TYPE" = "prod" ] && echo "distributed" || echo "dev").yml logs -f"
log_info "To stop: docker-compose -f docker-compose.$([ "$DEPLOYMENT_TYPE" = "prod" ] && echo "distributed" || echo "dev").yml down"
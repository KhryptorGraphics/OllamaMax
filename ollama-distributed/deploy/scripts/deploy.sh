#!/bin/bash

# Ollamacron Deployment Script
# Automated deployment for various environments

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
DEPLOYMENT_TYPE="docker"
ENVIRONMENT="development"
NAMESPACE="ollamacron"
REPLICAS=3
DOMAIN=""
TLS_ENABLED=false
MONITORING_ENABLED=true
BACKUP_ENABLED=false
DEBUG=false

# Logging functions
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}" >&2
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $1${NC}"
}

# Show usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Ollamacron Deployment Script

Options:
    -t, --type TYPE         Deployment type (docker, kubernetes, local) [default: docker]
    -e, --environment ENV   Environment (development, staging, production) [default: development]
    -n, --namespace NS      Kubernetes namespace [default: ollamacron]
    -r, --replicas NUM      Number of replicas [default: 3]
    -d, --domain DOMAIN     Domain name for ingress
    --tls                   Enable TLS/SSL
    --no-monitoring         Disable monitoring
    --backup                Enable backup
    --debug                 Enable debug mode
    -h, --help              Show this help message

Examples:
    # Deploy with Docker Compose for development
    $0 -t docker -e development

    # Deploy to Kubernetes for production
    $0 -t kubernetes -e production -r 5 -d ollamacron.example.com --tls

    # Deploy locally for testing
    $0 -t local -e development --debug

EOF
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -t|--type)
                DEPLOYMENT_TYPE="$2"
                shift 2
                ;;
            -e|--environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -n|--namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            -r|--replicas)
                REPLICAS="$2"
                shift 2
                ;;
            -d|--domain)
                DOMAIN="$2"
                shift 2
                ;;
            --tls)
                TLS_ENABLED=true
                shift
                ;;
            --no-monitoring)
                MONITORING_ENABLED=false
                shift
                ;;
            --backup)
                BACKUP_ENABLED=true
                shift
                ;;
            --debug)
                DEBUG=true
                shift
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done
}

# Validate prerequisites
validate_prerequisites() {
    log "Validating prerequisites..."
    
    case $DEPLOYMENT_TYPE in
        docker)
            if ! command -v docker &> /dev/null; then
                error "Docker is not installed"
                exit 1
            fi
            
            if ! command -v docker-compose &> /dev/null; then
                error "Docker Compose is not installed"
                exit 1
            fi
            ;;
        kubernetes)
            if ! command -v kubectl &> /dev/null; then
                error "kubectl is not installed"
                exit 1
            fi
            
            if ! command -v helm &> /dev/null; then
                error "Helm is not installed"
                exit 1
            fi
            
            # Check if connected to cluster
            if ! kubectl cluster-info &> /dev/null; then
                error "Not connected to a Kubernetes cluster"
                exit 1
            fi
            ;;
        local)
            if ! command -v go &> /dev/null; then
                error "Go is not installed"
                exit 1
            fi
            ;;
        *)
            error "Unsupported deployment type: $DEPLOYMENT_TYPE"
            exit 1
            ;;
    esac
    
    log "Prerequisites validated"
}

# Deploy with Docker Compose
deploy_docker() {
    log "Deploying with Docker Compose..."
    
    cd "$(dirname "$0")/../docker/compose"
    
    # Choose compose file based on environment
    if [[ "$ENVIRONMENT" == "development" ]]; then
        COMPOSE_FILE="docker-compose.yml"
    else
        COMPOSE_FILE="docker-compose.cluster.yml"
    fi
    
    # Set environment variables
    export OLLAMACRON_ENVIRONMENT="$ENVIRONMENT"
    export OLLAMACRON_REPLICAS="$REPLICAS"
    export OLLAMACRON_DOMAIN="$DOMAIN"
    export OLLAMACRON_TLS_ENABLED="$TLS_ENABLED"
    export OLLAMACRON_MONITORING_ENABLED="$MONITORING_ENABLED"
    export OLLAMACRON_DEBUG="$DEBUG"
    
    # Create network if it doesn't exist
    docker network create ollamacron-network || true
    
    # Build and deploy
    docker-compose -f "$COMPOSE_FILE" build
    docker-compose -f "$COMPOSE_FILE" up -d
    
    # Wait for services to be ready
    log "Waiting for services to be ready..."
    sleep 30
    
    # Health check
    if curl -f http://localhost:8080/health &> /dev/null; then
        log "Ollamacron is running and healthy"
    else
        error "Ollamacron health check failed"
        exit 1
    fi
    
    log "Docker deployment completed"
    
    # Display access information
    echo
    echo -e "${GREEN}🎉 Ollamacron deployed successfully!${NC}"
    echo
    echo -e "${BLUE}Access URLs:${NC}"
    echo "  API:     http://localhost:8080"
    echo "  Health:  http://localhost:8081/health"
    echo "  Metrics: http://localhost:9090/metrics"
    if [[ "$MONITORING_ENABLED" == "true" ]]; then
        echo "  Grafana: http://localhost:3000 (admin/admin)"
    fi
}

# Deploy to Kubernetes
deploy_kubernetes() {
    log "Deploying to Kubernetes..."
    
    cd "$(dirname "$0")/../kubernetes/helm"
    
    # Create namespace if it doesn't exist
    kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
    
    # Prepare values file
    VALUES_FILE="/tmp/ollamacron-values.yaml"
    cat > "$VALUES_FILE" << EOF
ollamacron:
  replicaCount: $REPLICAS
  config:
    logging:
      level: $([ "$DEBUG" == "true" ] && echo "debug" || echo "info")
  
ingress:
  enabled: $([ -n "$DOMAIN" ] && echo "true" || echo "false")
  hosts:
    - host: $DOMAIN
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: ollamacron-tls
      hosts:
        - $DOMAIN

prometheus:
  enabled: $MONITORING_ENABLED
  
grafana:
  enabled: $MONITORING_ENABLED
  
redis:
  enabled: true

EOF
    
    # Deploy with Helm
    helm upgrade --install ollamacron ./ollamacron \
        --namespace "$NAMESPACE" \
        --values "$VALUES_FILE" \
        --wait \
        --timeout 600s
    
    # Wait for deployment to be ready
    log "Waiting for deployment to be ready..."
    kubectl wait --for=condition=available deployment/ollamacron --namespace "$NAMESPACE" --timeout=300s
    
    log "Kubernetes deployment completed"
    
    # Display access information
    echo
    echo -e "${GREEN}🎉 Ollamacron deployed successfully!${NC}"
    echo
    echo -e "${BLUE}Access information:${NC}"
    echo "  Namespace: $NAMESPACE"
    if [[ -n "$DOMAIN" ]]; then
        echo "  URL: https://$DOMAIN"
    else
        echo "  Use port-forward to access: kubectl port-forward -n $NAMESPACE svc/ollamacron 8080:8080"
    fi
    echo
    echo -e "${BLUE}Useful commands:${NC}"
    echo "  kubectl get pods -n $NAMESPACE"
    echo "  kubectl logs -n $NAMESPACE -l app.kubernetes.io/name=ollamacron"
    echo "  kubectl port-forward -n $NAMESPACE svc/ollamacron 8080:8080"
}

# Deploy locally
deploy_local() {
    log "Deploying locally..."
    
    cd "$(dirname "$0")/../../.."
    
    # Set environment variables
    export OLLAMACRON_ENVIRONMENT="$ENVIRONMENT"
    export OLLAMACRON_DEBUG="$DEBUG"
    export OLLAMACRON_CONFIG="config/environments/$ENVIRONMENT.yaml"
    
    # Build the binary
    log "Building Ollamacron..."
    go build -o bin/ollamacron ./cmd/node
    
    # Create directories
    mkdir -p data/models data/storage logs
    
    # Copy configuration
    cp "deploy/config/environments/$ENVIRONMENT.yaml" config.yaml
    
    # Start the service
    log "Starting Ollamacron..."
    ./bin/ollamacron server --config config.yaml &
    
    # Store PID
    echo $! > ollamacron.pid
    
    # Wait for service to be ready
    log "Waiting for service to be ready..."
    sleep 10
    
    # Health check
    if curl -f http://localhost:8080/health &> /dev/null; then
        log "Ollamacron is running and healthy"
    else
        error "Ollamacron health check failed"
        exit 1
    fi
    
    log "Local deployment completed"
    
    # Display access information
    echo
    echo -e "${GREEN}🎉 Ollamacron deployed successfully!${NC}"
    echo
    echo -e "${BLUE}Access URLs:${NC}"
    echo "  API:     http://localhost:8080"
    echo "  Health:  http://localhost:8081/health"
    echo "  Metrics: http://localhost:9090/metrics"
    echo
    echo -e "${BLUE}Control commands:${NC}"
    echo "  Stop:    kill \$(cat ollamacron.pid)"
    echo "  Logs:    tail -f logs/ollamacron.log"
}

# Cleanup function
cleanup() {
    log "Cleaning up..."
    
    case $DEPLOYMENT_TYPE in
        docker)
            cd "$(dirname "$0")/../docker/compose"
            docker-compose down
            ;;
        kubernetes)
            helm uninstall ollamacron --namespace "$NAMESPACE" || true
            kubectl delete namespace "$NAMESPACE" || true
            ;;
        local)
            if [[ -f ollamacron.pid ]]; then
                kill "$(cat ollamacron.pid)" || true
                rm ollamacron.pid
            fi
            ;;
    esac
    
    log "Cleanup completed"
}

# Main function
main() {
    # Parse arguments
    parse_args "$@"
    
    # Show configuration
    log "Deployment Configuration:"
    info "  Type: $DEPLOYMENT_TYPE"
    info "  Environment: $ENVIRONMENT"
    info "  Namespace: $NAMESPACE"
    info "  Replicas: $REPLICAS"
    info "  Domain: ${DOMAIN:-"<none>"}"
    info "  TLS: $TLS_ENABLED"
    info "  Monitoring: $MONITORING_ENABLED"
    info "  Backup: $BACKUP_ENABLED"
    info "  Debug: $DEBUG"
    
    # Validate prerequisites
    validate_prerequisites
    
    # Deploy based on type
    case $DEPLOYMENT_TYPE in
        docker)
            deploy_docker
            ;;
        kubernetes)
            deploy_kubernetes
            ;;
        local)
            deploy_local
            ;;
    esac
    
    log "Deployment completed successfully!"
}

# Handle script interruption
trap cleanup EXIT

# Run main function
main "$@"
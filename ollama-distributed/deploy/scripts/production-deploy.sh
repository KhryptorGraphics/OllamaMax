#!/bin/bash
# Production Deployment Script with Zero-Downtime
# Complete automated deployment with health checks and rollback

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
DEPLOY_DIR="$PROJECT_ROOT/deploy"

# Default values
ENVIRONMENT="${ENVIRONMENT:-production}"
CLUSTER_NAME="${CLUSTER_NAME:-ollama-production}"
NAMESPACE="${NAMESPACE:-ollama-system}"
IMAGE_TAG="${IMAGE_TAG:-latest}"
DEPLOYMENT_STRATEGY="${DEPLOYMENT_STRATEGY:-blue-green}"
DRY_RUN="${DRY_RUN:-false}"
SKIP_TESTS="${SKIP_TESTS:-false}"
TIMEOUT="${TIMEOUT:-1800}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] INFO:${NC} $1"
}

success() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] SUCCESS:${NC} $1"
}

warning() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING:${NC} $1"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR:${NC} $1"
}

# Help function
show_help() {
    cat << EOF
Production Deployment Script for OllamaMax Distributed System

Usage: $0 [OPTIONS]

OPTIONS:
    -e, --environment ENVIRONMENT    Deployment environment (default: production)
    -c, --cluster CLUSTER_NAME       Kubernetes cluster name (default: ollama-production)
    -n, --namespace NAMESPACE        Kubernetes namespace (default: ollama-system)
    -t, --image-tag TAG              Docker image tag to deploy (default: latest)
    -s, --strategy STRATEGY          Deployment strategy: blue-green|canary|rolling (default: blue-green)
    -d, --dry-run                    Perform a dry run without making changes
    --skip-tests                     Skip pre-deployment tests
    --timeout SECONDS                Deployment timeout in seconds (default: 1800)
    -h, --help                       Show this help message

EXAMPLES:
    # Production deployment with blue-green strategy
    $0 --environment production --strategy blue-green --image-tag v1.2.3

    # Canary deployment with 10% traffic
    $0 --strategy canary --image-tag v1.2.3

    # Dry run to validate configuration
    $0 --dry-run --image-tag v1.2.3

ENVIRONMENT VARIABLES:
    AWS_REGION                       AWS region for deployment
    KUBECONFIG                       Path to kubeconfig file
    GITHUB_TOKEN                     GitHub token for image registry
    SLACK_WEBHOOK_URL                Slack webhook for notifications
    DATADOG_API_KEY                  Datadog API key for monitoring
EOF
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -e|--environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -c|--cluster)
                CLUSTER_NAME="$2"
                shift 2
                ;;
            -n|--namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            -t|--image-tag)
                IMAGE_TAG="$2"
                shift 2
                ;;
            -s|--strategy)
                DEPLOYMENT_STRATEGY="$2"
                shift 2
                ;;
            -d|--dry-run)
                DRY_RUN="true"
                shift
                ;;
            --skip-tests)
                SKIP_TESTS="true"
                shift
                ;;
            --timeout)
                TIMEOUT="$2"
                shift 2
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# Validate prerequisites
validate_prerequisites() {
    log "Validating prerequisites..."

    # Check required tools
    local tools=("kubectl" "helm" "docker" "aws" "jq" "yq")
    for tool in "${tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            error "Required tool '$tool' is not installed"
            exit 1
        fi
    done

    # Check AWS CLI configuration
    if ! aws sts get-caller-identity &> /dev/null; then
        error "AWS CLI not configured or credentials invalid"
        exit 1
    fi

    # Check kubectl configuration
    if ! kubectl cluster-info &> /dev/null; then
        error "kubectl not configured or cluster unreachable"
        exit 1
    fi

    # Check if ArgoCD Rollouts is installed (for blue-green/canary)
    if [[ "$DEPLOYMENT_STRATEGY" =~ ^(blue-green|canary)$ ]]; then
        if ! kubectl get crd rollouts.argoproj.io &> /dev/null; then
            error "ArgoCD Rollouts CRD not found. Please install ArgoCD Rollouts."
            exit 1
        fi
    fi

    # Validate image exists
    if [[ "$DRY_RUN" != "true" ]]; then
        log "Validating Docker image: ghcr.io/khryptorgraphics/ollamamax:$IMAGE_TAG"
        if ! docker manifest inspect "ghcr.io/khryptorgraphics/ollamamax:$IMAGE_TAG" &> /dev/null; then
            error "Docker image ghcr.io/khryptorgraphics/ollamamax:$IMAGE_TAG not found"
            exit 1
        fi
    fi

    success "Prerequisites validated successfully"
}

# Run pre-deployment tests
run_pre_deployment_tests() {
    if [[ "$SKIP_TESTS" == "true" ]]; then
        warning "Skipping pre-deployment tests"
        return 0
    fi

    log "Running pre-deployment tests..."

    # Check cluster health
    log "Checking cluster health..."
    if ! kubectl get nodes | grep -q "Ready"; then
        error "Some cluster nodes are not in Ready state"
        return 1
    fi

    # Check resource quotas
    log "Checking resource quotas..."
    local cpu_available=$(kubectl describe nodes | grep -A 5 "Allocatable:" | grep "cpu:" | awk '{sum += $2} END {print sum}')
    local memory_available=$(kubectl describe nodes | grep -A 5 "Allocatable:" | grep "memory:" | awk '{sum += $2} END {print sum}')

    log "Available resources: CPU: ${cpu_available:-unknown}, Memory: ${memory_available:-unknown}"

    # Run connectivity tests
    log "Running connectivity tests..."
    if ! curl -f -s "https://api.github.com" > /dev/null; then
        error "External connectivity test failed"
        return 1
    fi

    # Test database connectivity
    if kubectl get pods -n database | grep -q postgres-primary; then
        log "Testing database connectivity..."
        if ! kubectl exec -n database deployment/postgres-primary -- pg_isready -U ollamamax; then
            error "Database connectivity test failed"
            return 1
        fi
    fi

    # Test Redis connectivity
    if kubectl get pods -n database | grep -q redis-master; then
        log "Testing Redis connectivity..."
        if ! kubectl exec -n database deployment/redis-master -- redis-cli ping | grep -q PONG; then
            error "Redis connectivity test failed"
            return 1
        fi
    fi

    success "Pre-deployment tests completed successfully"
}

# Deploy infrastructure components
deploy_infrastructure() {
    log "Deploying infrastructure components..."

    # Create namespaces
    local namespaces=("ollama-system" "database" "monitoring")
    for ns in "${namespaces[@]}"; do
        if [[ "$DRY_RUN" == "true" ]]; then
            log "[DRY-RUN] Would create namespace: $ns"
        else
            kubectl create namespace "$ns" --dry-run=client -o yaml | kubectl apply -f -
        fi
    done

    # Deploy database infrastructure
    log "Deploying database infrastructure..."
    if [[ "$DRY_RUN" == "true" ]]; then
        log "[DRY-RUN] Would deploy database infrastructure"
        kubectl apply -f "$DEPLOY_DIR/integration/database-deployment.yaml" --dry-run=client
    else
        kubectl apply -f "$DEPLOY_DIR/integration/database-deployment.yaml"
        kubectl wait --for=condition=ready pod -l app=postgres,role=primary -n database --timeout=300s
        kubectl wait --for=condition=ready pod -l app=redis,role=master -n database --timeout=300s
    fi

    # Deploy monitoring infrastructure
    log "Deploying monitoring infrastructure..."
    if [[ "$DRY_RUN" == "true" ]]; then
        log "[DRY-RUN] Would deploy monitoring infrastructure"
        kubectl apply -f "$DEPLOY_DIR/integration/monitoring-deployment.yaml" --dry-run=client
    else
        kubectl apply -f "$DEPLOY_DIR/integration/monitoring-deployment.yaml"
        kubectl wait --for=condition=ready pod -l app=prometheus -n monitoring --timeout=300s
        kubectl wait --for=condition=ready pod -l app=grafana -n monitoring --timeout=300s
    fi

    success "Infrastructure components deployed successfully"
}

# Deploy application using selected strategy
deploy_application() {
    log "Deploying application using $DEPLOYMENT_STRATEGY strategy..."

    case "$DEPLOYMENT_STRATEGY" in
        "blue-green")
            deploy_blue_green
            ;;
        "canary")
            deploy_canary
            ;;
        "rolling")
            deploy_rolling
            ;;
        *)
            error "Unknown deployment strategy: $DEPLOYMENT_STRATEGY"
            exit 1
            ;;
    esac
}

# Blue-Green deployment
deploy_blue_green() {
    log "Starting blue-green deployment..."

    # Update image tag in rollout spec
    local temp_file=$(mktemp)
    yq eval ".spec.template.spec.containers[0].image = \"ghcr.io/khryptorgraphics/ollamamax:$IMAGE_TAG\"" \
        "$DEPLOY_DIR/integration/blue-green-deployment.yaml" > "$temp_file"

    if [[ "$DRY_RUN" == "true" ]]; then
        log "[DRY-RUN] Would apply blue-green deployment"
        kubectl apply -f "$temp_file" --dry-run=client
    else
        kubectl apply -f "$temp_file"

        # Wait for rollout to be ready for promotion
        log "Waiting for blue-green rollout to be ready..."
        kubectl wait --for=condition=Paused rollout/ollama-distributed-rollout -n "$NAMESPACE" --timeout="${TIMEOUT}s"

        # Run deployment validation
        if ! validate_deployment; then
            error "Deployment validation failed, aborting rollout"
            kubectl argo rollouts abort ollama-distributed-rollout -n "$NAMESPACE"
            kubectl argo rollouts undo ollama-distributed-rollout -n "$NAMESPACE"
            exit 1
        fi

        # Promote the rollout
        log "Promoting blue-green deployment..."
        kubectl argo rollouts promote ollama-distributed-rollout -n "$NAMESPACE"

        # Wait for completion
        kubectl argo rollouts status ollama-distributed-rollout -n "$NAMESPACE" --timeout="${TIMEOUT}s"
    fi

    rm -f "$temp_file"
    success "Blue-green deployment completed successfully"
}

# Canary deployment
deploy_canary() {
    log "Starting canary deployment..."
    
    # Update image in main deployment
    local temp_file=$(mktemp)
    yq eval ".spec.template.spec.containers[0].image = \"ghcr.io/khryptorgraphics/ollamamax:$IMAGE_TAG\"" \
        "$DEPLOY_DIR/integration/production-deployment.yaml" > "$temp_file"

    if [[ "$DRY_RUN" == "true" ]]; then
        log "[DRY-RUN] Would apply canary deployment"
        kubectl apply -f "$temp_file" --dry-run=client
    else
        kubectl apply -f "$temp_file"

        # Enable canary with 5% traffic
        log "Enabling canary with 5% traffic..."
        kubectl patch ingress ollama-canary-ingress -n "$NAMESPACE" --type='json' \
            -p='[{"op": "replace", "path": "/metadata/annotations/nginx.ingress.kubernetes.io~1canary-weight", "value": "5"}]'

        # Monitor canary for 5 minutes
        log "Monitoring canary deployment for 5 minutes..."
        sleep 300

        # Check canary metrics
        if validate_canary_metrics; then
            # Gradually increase traffic
            for weight in 10 25 50 100; do
                log "Increasing canary traffic to $weight%..."
                kubectl patch ingress ollama-canary-ingress -n "$NAMESPACE" --type='json' \
                    -p="[{\"op\": \"replace\", \"path\": \"/metadata/annotations/nginx.ingress.kubernetes.io~1canary-weight\", \"value\": \"$weight\"}]"
                
                sleep 120  # Wait 2 minutes between traffic increases
                
                if ! validate_canary_metrics; then
                    error "Canary metrics validation failed at $weight% traffic"
                    rollback_canary
                    exit 1
                fi
            done

            # Disable canary (100% traffic to new version)
            log "Finalizing canary deployment..."
            kubectl patch ingress ollama-canary-ingress -n "$NAMESPACE" --type='json' \
                -p='[{"op": "replace", "path": "/metadata/annotations/nginx.ingress.kubernetes.io~1canary-weight", "value": "0"}]'
        else
            error "Canary validation failed, rolling back"
            rollback_canary
            exit 1
        fi
    fi

    rm -f "$temp_file"
    success "Canary deployment completed successfully"
}

# Rolling deployment
deploy_rolling() {
    log "Starting rolling deployment..."

    local temp_file=$(mktemp)
    yq eval ".spec.template.spec.containers[0].image = \"ghcr.io/khryptorgraphics/ollamamax:$IMAGE_TAG\"" \
        "$DEPLOY_DIR/integration/production-deployment.yaml" > "$temp_file"

    if [[ "$DRY_RUN" == "true" ]]; then
        log "[DRY-RUN] Would apply rolling deployment"
        kubectl apply -f "$temp_file" --dry-run=client
    else
        kubectl apply -f "$temp_file"
        kubectl rollout status deployment/ollama-distributed -n "$NAMESPACE" --timeout="${TIMEOUT}s"
    fi

    rm -f "$temp_file"
    success "Rolling deployment completed successfully"
}

# Validate deployment health
validate_deployment() {
    log "Validating deployment health..."

    # Check pod readiness
    if ! kubectl get pods -n "$NAMESPACE" -l app.kubernetes.io/name=ollama-distributed | grep -q "Running"; then
        error "No running pods found"
        return 1
    fi

    # Test API endpoints
    local service_name
    if [[ "$DEPLOYMENT_STRATEGY" == "blue-green" ]]; then
        service_name="ollama-api-preview"
    else
        service_name="ollama-api"
    fi

    # Port forward for testing
    local port_forward_pid
    kubectl port-forward service/"$service_name" 8080:8080 -n "$NAMESPACE" &
    port_forward_pid=$!
    sleep 5

    # Test health endpoint
    local health_check=false
    for i in {1..30}; do
        if curl -f -s "http://localhost:8080/health" > /dev/null; then
            health_check=true
            break
        fi
        sleep 2
    done

    # Clean up port forward
    kill $port_forward_pid 2>/dev/null || true

    if [[ "$health_check" != "true" ]]; then
        error "Health check failed"
        return 1
    fi

    # Check metrics
    if ! validate_metrics; then
        error "Metrics validation failed"
        return 1
    fi

    success "Deployment validation completed successfully"
    return 0
}

# Validate metrics
validate_metrics() {
    log "Validating deployment metrics..."

    # Query Prometheus for success rate
    local prometheus_url="http://prometheus.monitoring.svc.cluster.local:9090"
    local success_rate_query='sum(rate(ollama_http_requests_total{status!~"5.."}[5m])) / sum(rate(ollama_http_requests_total[5m]))'
    
    # Port forward to Prometheus
    local port_forward_pid
    kubectl port-forward service/prometheus 9090:9090 -n monitoring &
    port_forward_pid=$!
    sleep 5

    local success_rate=$(curl -s "http://localhost:9090/api/v1/query?query=${success_rate_query}" | jq -r '.data.result[0].value[1] // "0"')
    
    # Clean up port forward
    kill $port_forward_pid 2>/dev/null || true

    local success_rate_percent=$(echo "$success_rate * 100" | bc -l 2>/dev/null || echo "0")
    log "Current success rate: ${success_rate_percent}%"

    if (( $(echo "$success_rate < 0.99" | bc -l) )); then
        error "Success rate below 99%: ${success_rate_percent}%"
        return 1
    fi

    success "Metrics validation passed"
    return 0
}

# Validate canary metrics
validate_canary_metrics() {
    log "Validating canary metrics..."

    # Check error rate specifically for canary traffic
    local prometheus_url="http://prometheus.monitoring.svc.cluster.local:9090"
    local error_rate_query='sum(rate(ollama_http_requests_total{service="ollama-api-preview",status=~"5.."}[5m])) / sum(rate(ollama_http_requests_total{service="ollama-api-preview"}[5m]))'
    
    # Port forward to Prometheus
    local port_forward_pid
    kubectl port-forward service/prometheus 9090:9090 -n monitoring &
    port_forward_pid=$!
    sleep 5

    local error_rate=$(curl -s "http://localhost:9090/api/v1/query?query=${error_rate_query}" | jq -r '.data.result[0].value[1] // "0"')
    
    # Clean up port forward
    kill $port_forward_pid 2>/dev/null || true

    local error_rate_percent=$(echo "$error_rate * 100" | bc -l 2>/dev/null || echo "0")
    log "Canary error rate: ${error_rate_percent}%"

    if (( $(echo "$error_rate > 0.01" | bc -l) )); then
        error "Canary error rate above 1%: ${error_rate_percent}%"
        return 1
    fi

    success "Canary metrics validation passed"
    return 0
}

# Rollback canary deployment
rollback_canary() {
    log "Rolling back canary deployment..."
    
    # Set canary weight to 0
    kubectl patch ingress ollama-canary-ingress -n "$NAMESPACE" --type='json' \
        -p='[{"op": "replace", "path": "/metadata/annotations/nginx.ingress.kubernetes.io~1canary-weight", "value": "0"}]'
    
    success "Canary rollback completed"
}

# Send notifications
send_notifications() {
    local status=$1
    local message=$2

    if [[ -n "${SLACK_WEBHOOK_URL:-}" ]]; then
        local color
        case $status in
            "success") color="good" ;;
            "warning") color="warning" ;;
            "error") color="danger" ;;
            *) color="warning" ;;
        esac

        local payload=$(jq -n \
            --arg text "$message" \
            --arg color "$color" \
            --arg environment "$ENVIRONMENT" \
            --arg strategy "$DEPLOYMENT_STRATEGY" \
            --arg image_tag "$IMAGE_TAG" \
            '{
                "attachments": [{
                    "color": $color,
                    "title": "OllamaMax Deployment Notification",
                    "text": $text,
                    "fields": [
                        {"title": "Environment", "value": $environment, "short": true},
                        {"title": "Strategy", "value": $strategy, "short": true},
                        {"title": "Image Tag", "value": $image_tag, "short": true}
                    ],
                    "footer": "OllamaMax Deployment Bot",
                    "ts": now
                }]
            }')

        curl -s -X POST -H 'Content-type: application/json' \
            --data "$payload" "$SLACK_WEBHOOK_URL" > /dev/null || true
    fi
}

# Cleanup function
cleanup() {
    log "Performing cleanup..."
    
    # Kill any background port forwards
    pkill -f "kubectl port-forward" 2>/dev/null || true
    
    # Remove temporary files
    rm -f /tmp/ollama-deploy-*
}

# Main deployment function
main() {
    # Set up cleanup trap
    trap cleanup EXIT

    log "Starting OllamaMax production deployment"
    log "Environment: $ENVIRONMENT"
    log "Cluster: $CLUSTER_NAME"
    log "Namespace: $NAMESPACE"
    log "Image Tag: $IMAGE_TAG"
    log "Strategy: $DEPLOYMENT_STRATEGY"
    log "Dry Run: $DRY_RUN"

    # Validate prerequisites
    validate_prerequisites

    # Run pre-deployment tests
    run_pre_deployment_tests

    # Deploy infrastructure
    deploy_infrastructure

    # Deploy application
    deploy_application

    # Final validation
    if [[ "$DRY_RUN" != "true" ]]; then
        validate_deployment
    fi

    success "🎉 Production deployment completed successfully!"
    
    # Send success notification
    send_notifications "success" "✅ OllamaMax deployment to $ENVIRONMENT completed successfully with $DEPLOYMENT_STRATEGY strategy"
}

# Error handling
handle_error() {
    local line_number=$1
    local error_message=$2
    
    error "Deployment failed at line $line_number: $error_message"
    send_notifications "error" "❌ OllamaMax deployment to $ENVIRONMENT failed: $error_message"
    
    cleanup
    exit 1
}

# Set up error trap
trap 'handle_error ${LINENO} "$BASH_COMMAND"' ERR

# Parse arguments and run main function
parse_args "$@"
main
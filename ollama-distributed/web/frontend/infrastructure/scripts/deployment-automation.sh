#!/bin/bash
# Ollama Frontend - Deployment Automation Script
# Production-ready deployment with comprehensive validation and rollback

set -euo pipefail

# Configuration
NAMESPACE="ollama-frontend"
APP_NAME="ollama-frontend"
IMAGE_REGISTRY="ghcr.io"
IMAGE_NAME="ollamamax/frontend"
KUBECTL_TIMEOUT="600s"
HEALTH_CHECK_RETRIES=30
HEALTH_CHECK_INTERVAL=10

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# Usage function
usage() {
    cat << EOF
Ollama Frontend Deployment Automation

Usage: $0 [OPTIONS]

Options:
    -v, --version VERSION       Image version to deploy (required)
    -s, --strategy STRATEGY     Deployment strategy (blue-green|canary) [default: blue-green]
    -e, --environment ENV       Environment (staging|production) [default: staging]
    -n, --namespace NAMESPACE   Kubernetes namespace [default: ollama-frontend]
    -d, --dry-run              Perform dry run without actual deployment
    -f, --force                Force deployment without confirmation
    -r, --rollback             Rollback to previous version
    -h, --help                 Show this help message

Examples:
    $0 --version v1.2.3 --environment production
    $0 --version latest --strategy canary --dry-run
    $0 --rollback --environment production

EOF
}

# Parse command line arguments
parse_args() {
    VERSION=""
    STRATEGY="blue-green"
    ENVIRONMENT="staging"
    DRY_RUN=false
    FORCE=false
    ROLLBACK=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -s|--strategy)
                STRATEGY="$2"
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
            -d|--dry-run)
                DRY_RUN=true
                shift
                ;;
            -f|--force)
                FORCE=true
                shift
                ;;
            -r|--rollback)
                ROLLBACK=true
                shift
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done

    # Validation
    if [[ "$ROLLBACK" == "false" && -z "$VERSION" ]]; then
        log_error "Version is required unless performing rollback"
        usage
        exit 1
    fi

    if [[ "$STRATEGY" != "blue-green" && "$STRATEGY" != "canary" ]]; then
        log_error "Strategy must be 'blue-green' or 'canary'"
        exit 1
    fi

    if [[ "$ENVIRONMENT" != "staging" && "$ENVIRONMENT" != "production" ]]; then
        log_error "Environment must be 'staging' or 'production'"
        exit 1
    fi
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check required tools
    for tool in kubectl helm jq curl; do
        if ! command -v $tool &> /dev/null; then
            log_error "$tool is required but not installed"
            exit 1
        fi
    done

    # Check kubectl connection
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi

    # Check namespace exists
    if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
        log_error "Namespace '$NAMESPACE' does not exist"
        exit 1
    fi

    log_success "Prerequisites check passed"
}

# Get current deployment state
get_current_state() {
    log_info "Getting current deployment state..."

    # Get active color for blue-green deployments
    if [[ "$STRATEGY" == "blue-green" ]]; then
        ACTIVE_COLOR=$(kubectl get service ${APP_NAME}-active -n "$NAMESPACE" -o jsonpath='{.spec.selector.version}' 2>/dev/null || echo "blue")
        TARGET_COLOR="green"
        if [[ "$ACTIVE_COLOR" == "green" ]]; then
            TARGET_COLOR="blue"
        fi
        log_info "Current active: $ACTIVE_COLOR, Target: $TARGET_COLOR"
    fi

    # Get current image version
    CURRENT_VERSION=$(kubectl get deployment ${APP_NAME}-${ACTIVE_COLOR:-main} -n "$NAMESPACE" -o jsonpath='{.spec.template.spec.containers[0].image}' 2>/dev/null | cut -d':' -f2 || echo "unknown")
    log_info "Current version: $CURRENT_VERSION"
}

# Create deployment backup
create_backup() {
    log_info "Creating deployment backup..."

    BACKUP_NAME="pre-deployment-$(date +%Y%m%d-%H%M%S)"
    
    if command -v velero &> /dev/null; then
        if [[ "$DRY_RUN" == "false" ]]; then
            velero backup create "$BACKUP_NAME" \
                --include-namespaces "$NAMESPACE" \
                --wait || {
                log_warning "Backup creation failed, continuing anyway"
            }
        else
            log_info "[DRY-RUN] Would create backup: $BACKUP_NAME"
        fi
    else
        log_warning "Velero not available, skipping backup"
    fi
}

# Validate image exists
validate_image() {
    log_info "Validating image exists..."

    IMAGE_FULL="${IMAGE_REGISTRY}/${IMAGE_NAME}:${VERSION}"
    
    if [[ "$DRY_RUN" == "false" ]]; then
        if ! docker manifest inspect "$IMAGE_FULL" &> /dev/null; then
            log_error "Image $IMAGE_FULL does not exist or is not accessible"
            exit 1
        fi
        
        # Check for security vulnerabilities
        log_info "Running security scan on image..."
        if command -v trivy &> /dev/null; then
            trivy image --severity HIGH,CRITICAL --exit-code 1 "$IMAGE_FULL" || {
                log_error "Security scan failed - high/critical vulnerabilities found"
                if [[ "$FORCE" == "false" ]]; then
                    exit 1
                fi
                log_warning "Proceeding due to --force flag"
            }
        fi
    else
        log_info "[DRY-RUN] Would validate image: $IMAGE_FULL"
    fi
}

# Deploy using blue-green strategy
deploy_blue_green() {
    log_info "Starting blue-green deployment..."

    IMAGE_FULL="${IMAGE_REGISTRY}/${IMAGE_NAME}:${VERSION}"
    
    if [[ "$DRY_RUN" == "false" ]]; then
        # Update target deployment
        kubectl set image deployment/${APP_NAME}-${TARGET_COLOR} \
            ${APP_NAME}=${IMAGE_FULL} \
            -n "$NAMESPACE"
        
        # Wait for deployment to complete
        log_info "Waiting for deployment to complete..."
        kubectl rollout status deployment/${APP_NAME}-${TARGET_COLOR} \
            -n "$NAMESPACE" --timeout="$KUBECTL_TIMEOUT"
    else
        log_info "[DRY-RUN] Would update deployment: ${APP_NAME}-${TARGET_COLOR} with image: $IMAGE_FULL"
    fi
}

# Deploy using canary strategy
deploy_canary() {
    log_info "Starting canary deployment..."

    if [[ "$DRY_RUN" == "false" ]]; then
        # Use Argo Rollouts for canary deployment
        if ! kubectl get rollout ${APP_NAME}-rollout -n "$NAMESPACE" &> /dev/null; then
            log_error "Argo Rollouts not configured for canary deployment"
            exit 1
        fi
        
        # Update rollout image
        kubectl argo rollouts set image ${APP_NAME}-rollout \
            ${APP_NAME}=${IMAGE_REGISTRY}/${IMAGE_NAME}:${VERSION} \
            -n "$NAMESPACE"
        
        # Monitor rollout progress
        log_info "Monitoring canary rollout progress..."
        kubectl argo rollouts status ${APP_NAME}-rollout \
            -n "$NAMESPACE" --timeout="$KUBECTL_TIMEOUT"
    else
        log_info "[DRY-RUN] Would start canary deployment with image: ${IMAGE_REGISTRY}/${IMAGE_NAME}:${VERSION}"
    fi
}

# Run health checks
run_health_checks() {
    log_info "Running health checks..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY-RUN] Would run health checks"
        return 0
    fi

    local service_name
    if [[ "$STRATEGY" == "blue-green" ]]; then
        service_name="${APP_NAME}-${TARGET_COLOR}"
    else
        service_name="${APP_NAME}-stable"
    fi

    # Port forward to test the service
    log_info "Testing $service_name service..."
    kubectl port-forward service/$service_name 8080:3000 -n "$NAMESPACE" &
    PF_PID=$!
    
    # Cleanup function for port-forward
    cleanup_port_forward() {
        if kill -0 $PF_PID 2>/dev/null; then
            kill $PF_PID
        fi
    }
    trap cleanup_port_forward EXIT

    # Wait for port-forward to establish
    sleep 5

    # Run health checks
    local retry=0
    while [[ $retry -lt $HEALTH_CHECK_RETRIES ]]; do
        if curl -f -s http://localhost:8080/health > /dev/null 2>&1; then
            log_success "Health check passed"
            break
        fi
        
        retry=$((retry + 1))
        log_info "Health check attempt $retry/$HEALTH_CHECK_RETRIES failed, retrying in ${HEALTH_CHECK_INTERVAL}s..."
        sleep $HEALTH_CHECK_INTERVAL
    done

    if [[ $retry -eq $HEALTH_CHECK_RETRIES ]]; then
        log_error "Health checks failed after $HEALTH_CHECK_RETRIES attempts"
        cleanup_port_forward
        exit 1
    fi

    # Additional endpoint tests
    local endpoints=("/ready" "/api/v1/status")
    for endpoint in "${endpoints[@]}"; do
        if ! curl -f -s "http://localhost:8080$endpoint" > /dev/null 2>&1; then
            log_warning "Endpoint $endpoint check failed"
        else
            log_success "Endpoint $endpoint check passed"
        fi
    done

    cleanup_port_forward
}

# Switch traffic (blue-green only)
switch_traffic() {
    if [[ "$STRATEGY" != "blue-green" ]]; then
        return 0
    fi

    log_info "Switching traffic to $TARGET_COLOR..."

    if [[ "$DRY_RUN" == "false" ]]; then
        # Update active service
        kubectl patch service ${APP_NAME}-active -n "$NAMESPACE" \
            -p "{\"spec\":{\"selector\":{\"version\":\"$TARGET_COLOR\"}}}"
        
        # Verify the switch
        NEW_ACTIVE=$(kubectl get service ${APP_NAME}-active -n "$NAMESPACE" -o jsonpath='{.spec.selector.version}')
        if [[ "$NEW_ACTIVE" == "$TARGET_COLOR" ]]; then
            log_success "Traffic switched successfully to $TARGET_COLOR"
        else
            log_error "Traffic switch failed - active is still $NEW_ACTIVE"
            exit 1
        fi
    else
        log_info "[DRY-RUN] Would switch traffic from $ACTIVE_COLOR to $TARGET_COLOR"
    fi
}

# Run production validation tests
run_production_tests() {
    if [[ "$ENVIRONMENT" != "production" ]]; then
        return 0
    fi

    log_info "Running production validation tests..."

    if [[ "$DRY_RUN" == "false" ]]; then
        # Wait for DNS propagation
        sleep 30

        # Test public endpoints
        local base_url="https://ollama.example.com"
        if [[ "$ENVIRONMENT" == "staging" ]]; then
            base_url="https://staging.ollama.example.com"
        fi

        local endpoints=("/health" "/ready" "/api/v1/status")
        for endpoint in "${endpoints[@]}"; do
            if curl -f -s "$base_url$endpoint" > /dev/null 2>&1; then
                log_success "Public endpoint $endpoint is accessible"
            else
                log_error "Public endpoint $endpoint is not accessible"
                exit 1
            fi
        done

        # Test authentication flow
        log_info "Testing authentication flow..."
        local auth_response
        auth_response=$(curl -s -w "%{http_code}" -X POST "$base_url/api/auth/validate" \
            -H "Content-Type: application/json" \
            -d '{"token":"test"}' -o /dev/null)
        
        if [[ "$auth_response" == "401" ]] || [[ "$auth_response" == "400" ]]; then
            log_success "Authentication endpoint responding correctly"
        else
            log_warning "Authentication endpoint returned unexpected status: $auth_response"
        fi
    else
        log_info "[DRY-RUN] Would run production validation tests"
    fi
}

# Monitor deployment metrics
monitor_deployment() {
    log_info "Monitoring deployment metrics for 5 minutes..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY-RUN] Would monitor deployment metrics"
        return 0
    fi

    local prometheus_url="http://prometheus.monitoring.svc.cluster.local:9090"
    
    for i in {1..10}; do
        log_info "Monitoring... $(($i * 30)) seconds elapsed"
        
        # Check error rate
        local error_rate
        error_rate=$(curl -s "$prometheus_url/api/v1/query?query=sum(rate(http_requests_total{job=\"ollama-frontend\",code=~\"5..\"}[5m]))/sum(rate(http_requests_total{job=\"ollama-frontend\"}[5m]))" | \
            jq -r '.data.result[0].value[1]' 2>/dev/null || echo "0")
        
        # Check response time
        local response_time
        response_time=$(curl -s "$prometheus_url/api/v1/query?query=histogram_quantile(0.95,sum(rate(http_request_duration_seconds_bucket{job=\"ollama-frontend\"}[5m])))" | \
            jq -r '.data.result[0].value[1]' 2>/dev/null || echo "0")
        
        log_info "Metrics - Error rate: ${error_rate:-N/A}, P95 response time: ${response_time:-N/A}s"
        
        # Check for issues
        if [[ "$error_rate" != "0" ]] && (( $(echo "$error_rate > 0.05" | bc -l 2>/dev/null || echo 0) )); then
            log_error "High error rate detected: $error_rate"
            return 1
        fi
        
        if [[ "$response_time" != "0" ]] && (( $(echo "$response_time > 2.0" | bc -l 2>/dev/null || echo 0) )); then
            log_warning "High response time detected: ${response_time}s"
        fi
        
        sleep 30
    done

    log_success "Deployment monitoring completed - no issues detected"
    return 0
}

# Rollback deployment
rollback_deployment() {
    log_info "Starting rollback procedure..."

    if [[ "$STRATEGY" == "blue-green" ]]; then
        if [[ "$DRY_RUN" == "false" ]]; then
            # Switch back to previous color
            local previous_color="blue"
            if [[ "$ACTIVE_COLOR" == "blue" ]]; then
                previous_color="green"
            fi
            
            kubectl patch service ${APP_NAME}-active -n "$NAMESPACE" \
                -p "{\"spec\":{\"selector\":{\"version\":\"$previous_color\"}}}"
            
            log_success "Rolled back to $previous_color deployment"
        else
            log_info "[DRY-RUN] Would rollback to previous deployment"
        fi
    elif [[ "$STRATEGY" == "canary" ]]; then
        if [[ "$DRY_RUN" == "false" ]]; then
            kubectl argo rollouts abort ${APP_NAME}-rollout -n "$NAMESPACE"
            kubectl argo rollouts undo ${APP_NAME}-rollout -n "$NAMESPACE"
            log_success "Canary rollback initiated"
        else
            log_info "[DRY-RUN] Would abort canary and rollback"
        fi
    fi
}

# Update deployment annotations
update_deployment_record() {
    if [[ "$DRY_RUN" == "true" ]]; then
        return 0
    fi

    log_info "Updating deployment record..."

    local deployment_name
    if [[ "$STRATEGY" == "blue-green" ]]; then
        deployment_name="${APP_NAME}-${TARGET_COLOR}"
    else
        deployment_name="${APP_NAME}-rollout"
    fi

    kubectl annotate deployment "$deployment_name" -n "$NAMESPACE" \
        deployment.kubernetes.io/deployed-by="$(whoami)" \
        deployment.kubernetes.io/deployed-at="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        deployment.kubernetes.io/version="$VERSION" \
        deployment.kubernetes.io/strategy="$STRATEGY" \
        deployment.kubernetes.io/environment="$ENVIRONMENT" \
        --overwrite
}

# Send notifications
send_notification() {
    local status=$1
    local message=$2
    
    log_info "Sending notification: $status"
    
    if [[ -n "${SLACK_WEBHOOK_URL:-}" ]]; then
        local color="good"
        if [[ "$status" == "failed" ]]; then
            color="danger"
        elif [[ "$status" == "warning" ]]; then
            color="warning"
        fi
        
        local payload=$(cat << EOF
{
    "attachments": [
        {
            "color": "$color",
            "title": "Ollama Frontend Deployment $status",
            "fields": [
                {"title": "Environment", "value": "$ENVIRONMENT", "short": true},
                {"title": "Version", "value": "${VERSION:-N/A}", "short": true},
                {"title": "Strategy", "value": "$STRATEGY", "short": true},
                {"title": "Namespace", "value": "$NAMESPACE", "short": true}
            ],
            "text": "$message"
        }
    ]
}
EOF
        )
        
        curl -X POST "$SLACK_WEBHOOK_URL" \
            -H 'Content-type: application/json' \
            -d "$payload" || log_warning "Failed to send Slack notification"
    fi
}

# Cleanup function
cleanup() {
    local exit_code=$?
    
    if [[ $exit_code -ne 0 ]]; then
        log_error "Deployment failed with exit code $exit_code"
        send_notification "failed" "Deployment failed. Check logs for details."
        
        if [[ "$FORCE" == "false" && "$ROLLBACK" == "false" ]]; then
            read -p "Do you want to rollback? (y/N): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                rollback_deployment
            fi
        fi
    fi
    
    exit $exit_code
}

# Main execution function
main() {
    log_info "Starting Ollama Frontend deployment automation"
    log_info "Strategy: $STRATEGY, Environment: $ENVIRONMENT, Namespace: $NAMESPACE"
    
    if [[ "$ROLLBACK" == "true" ]]; then
        log_info "Rollback requested"
        check_prerequisites
        get_current_state
        rollback_deployment
        send_notification "success" "Rollback completed successfully"
        return 0
    fi

    log_info "Version: $VERSION"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN MODE - No actual changes will be made"
    fi

    # Confirmation for production
    if [[ "$ENVIRONMENT" == "production" && "$FORCE" == "false" && "$DRY_RUN" == "false" ]]; then
        echo
        log_warning "You are about to deploy to PRODUCTION environment!"
        log_warning "Version: $VERSION"
        log_warning "Strategy: $STRATEGY"
        read -p "Are you sure you want to continue? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Deployment cancelled"
            exit 0
        fi
    fi

    # Set trap for cleanup
    trap cleanup EXIT

    # Execute deployment steps
    check_prerequisites
    get_current_state
    create_backup
    validate_image

    if [[ "$STRATEGY" == "blue-green" ]]; then
        deploy_blue_green
    else
        deploy_canary
    fi

    run_health_checks
    switch_traffic
    run_production_tests

    # Monitor deployment and rollback if issues detected
    if ! monitor_deployment; then
        log_error "Deployment monitoring detected issues"
        if [[ "$FORCE" == "false" ]]; then
            rollback_deployment
            send_notification "failed" "Deployment rolled back due to monitoring issues"
            exit 1
        else
            log_warning "Continuing despite monitoring issues due to --force flag"
        fi
    fi

    update_deployment_record
    log_success "Deployment completed successfully!"
    send_notification "success" "Deployment completed successfully"
}

# Script entry point
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    parse_args "$@"
    main
fi
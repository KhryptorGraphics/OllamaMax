#!/bin/bash

# OllamaMax Rollback Script
# Automated rollback for failed deployments

set -e

echo "🔄 OllamaMax Rollback System"
echo "============================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
ENVIRONMENT=""
DEPLOYMENT_TYPE="kubernetes"
ROLLBACK_STEPS=1
DRY_RUN=false
FORCE=false
BACKUP_NAMESPACE="ollama-backup"

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "SUCCESS")
            echo -e "${GREEN}✅ $message${NC}"
            ;;
        "ERROR")
            echo -e "${RED}❌ $message${NC}"
            ;;
        "WARNING")
            echo -e "${YELLOW}⚠️  $message${NC}"
            ;;
        "INFO")
            echo -e "${BLUE}ℹ️  $message${NC}"
            ;;
    esac
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            --type)
                DEPLOYMENT_TYPE="$2"
                shift 2
                ;;
            --steps)
                ROLLBACK_STEPS="$2"
                shift 2
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --force)
                FORCE=true
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                echo "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# Show help
show_help() {
    cat << EOF
OllamaMax Rollback Script

Usage: $0 --environment ENV [OPTIONS]

Options:
    --environment ENV    Target environment (staging, production) [REQUIRED]
    --type TYPE          Deployment type (kubernetes, docker, local)
    --steps N            Number of rollback steps (default: 1)
    --dry-run            Show what would be done without executing
    --force              Force rollback without confirmation
    --help               Show this help message

Examples:
    $0 --environment staging
    $0 --environment production --type kubernetes --steps 2
    $0 --environment staging --dry-run
EOF
}

# Validate arguments
validate_args() {
    if [ -z "$ENVIRONMENT" ]; then
        print_status "ERROR" "Environment is required. Use --environment staging|production"
        exit 1
    fi

    if [[ ! "$ENVIRONMENT" =~ ^(staging|production)$ ]]; then
        print_status "ERROR" "Invalid environment: $ENVIRONMENT. Must be staging or production"
        exit 1
    fi

    if [[ ! "$DEPLOYMENT_TYPE" =~ ^(kubernetes|docker|local)$ ]]; then
        print_status "ERROR" "Invalid deployment type: $DEPLOYMENT_TYPE"
        exit 1
    fi
}

# Get current deployment info
get_current_deployment() {
    print_status "INFO" "Getting current deployment information..."

    case $DEPLOYMENT_TYPE in
        kubernetes)
            get_kubernetes_deployment
            ;;
        docker)
            get_docker_deployment
            ;;
        local)
            get_local_deployment
            ;;
    esac
}

# Get Kubernetes deployment info
get_kubernetes_deployment() {
    local namespace="ollama-$ENVIRONMENT"
    
    if ! kubectl get namespace "$namespace" >/dev/null 2>&1; then
        print_status "ERROR" "Kubernetes namespace $namespace not found"
        exit 1
    fi

    CURRENT_DEPLOYMENT=$(kubectl get deployment ollama-distributed -n "$namespace" -o jsonpath='{.metadata.name}' 2>/dev/null || echo "")
    CURRENT_IMAGE=$(kubectl get deployment ollama-distributed -n "$namespace" -o jsonpath='{.spec.template.spec.containers[0].image}' 2>/dev/null || echo "")
    CURRENT_REPLICAS=$(kubectl get deployment ollama-distributed -n "$namespace" -o jsonpath='{.spec.replicas}' 2>/dev/null || echo "0")

    print_status "INFO" "Current Kubernetes deployment:"
    echo "  Namespace: $namespace"
    echo "  Deployment: $CURRENT_DEPLOYMENT"
    echo "  Image: $CURRENT_IMAGE"
    echo "  Replicas: $CURRENT_REPLICAS"
}

# Get Docker deployment info
get_docker_deployment() {
    CURRENT_CONTAINER=$(docker ps --filter "name=ollama-distributed-$ENVIRONMENT" --format "{{.Names}}" | head -1)
    CURRENT_IMAGE=$(docker ps --filter "name=ollama-distributed-$ENVIRONMENT" --format "{{.Image}}" | head -1)

    print_status "INFO" "Current Docker deployment:"
    echo "  Container: $CURRENT_CONTAINER"
    echo "  Image: $CURRENT_IMAGE"
}

# Get local deployment info
get_local_deployment() {
    local pid_file="/var/run/ollama-distributed-$ENVIRONMENT.pid"
    
    if [ -f "$pid_file" ]; then
        CURRENT_PID=$(cat "$pid_file")
        print_status "INFO" "Current local deployment:"
        echo "  PID: $CURRENT_PID"
        echo "  PID file: $pid_file"
    else
        print_status "WARNING" "No local deployment found (no PID file)"
    fi
}

# Get rollback target
get_rollback_target() {
    print_status "INFO" "Determining rollback target..."

    case $DEPLOYMENT_TYPE in
        kubernetes)
            get_kubernetes_rollback_target
            ;;
        docker)
            get_docker_rollback_target
            ;;
        local)
            get_local_rollback_target
            ;;
    esac
}

# Get Kubernetes rollback target
get_kubernetes_rollback_target() {
    local namespace="ollama-$ENVIRONMENT"
    
    # Get rollout history
    local history
    history=$(kubectl rollout history deployment/ollama-distributed -n "$namespace" 2>/dev/null || echo "")
    
    if [ -z "$history" ]; then
        print_status "ERROR" "No rollout history found"
        exit 1
    fi

    # Get the revision to rollback to
    local current_revision
    current_revision=$(kubectl get deployment ollama-distributed -n "$namespace" -o jsonpath='{.metadata.annotations.deployment\.kubernetes\.io/revision}' 2>/dev/null || echo "1")
    
    ROLLBACK_REVISION=$((current_revision - ROLLBACK_STEPS))
    
    if [ $ROLLBACK_REVISION -lt 1 ]; then
        ROLLBACK_REVISION=1
    fi

    print_status "INFO" "Kubernetes rollback target:"
    echo "  Current revision: $current_revision"
    echo "  Rollback to revision: $ROLLBACK_REVISION"
    echo "  Rollback steps: $ROLLBACK_STEPS"
}

# Get Docker rollback target
get_docker_rollback_target() {
    # Look for backup images
    local backup_images
    backup_images=$(docker images --filter "label=ollama.backup=true" --filter "label=ollama.environment=$ENVIRONMENT" --format "{{.Repository}}:{{.Tag}}" | head -5)
    
    if [ -z "$backup_images" ]; then
        print_status "ERROR" "No backup Docker images found for rollback"
        exit 1
    fi

    ROLLBACK_IMAGE=$(echo "$backup_images" | head -1)
    
    print_status "INFO" "Docker rollback target:"
    echo "  Rollback image: $ROLLBACK_IMAGE"
    echo "  Available backups:"
    echo "$backup_images" | sed 's/^/    /'
}

# Get local rollback target
get_local_rollback_target() {
    local backup_dir="/opt/ollama-distributed/backups/$ENVIRONMENT"
    
    if [ ! -d "$backup_dir" ]; then
        print_status "ERROR" "No backup directory found: $backup_dir"
        exit 1
    fi

    ROLLBACK_BINARY=$(find "$backup_dir" -name "ollama-distributed-*" -type f -executable | sort -r | head -1)
    
    if [ -z "$ROLLBACK_BINARY" ]; then
        print_status "ERROR" "No backup binary found in $backup_dir"
        exit 1
    fi

    print_status "INFO" "Local rollback target:"
    echo "  Rollback binary: $ROLLBACK_BINARY"
}

# Confirm rollback
confirm_rollback() {
    if [ "$FORCE" = true ]; then
        return 0
    fi

    echo ""
    print_status "WARNING" "ROLLBACK CONFIRMATION REQUIRED"
    echo "Environment: $ENVIRONMENT"
    echo "Deployment Type: $DEPLOYMENT_TYPE"
    echo "Rollback Steps: $ROLLBACK_STEPS"
    
    case $DEPLOYMENT_TYPE in
        kubernetes)
            echo "Rollback to revision: $ROLLBACK_REVISION"
            ;;
        docker)
            echo "Rollback to image: $ROLLBACK_IMAGE"
            ;;
        local)
            echo "Rollback to binary: $ROLLBACK_BINARY"
            ;;
    esac

    echo ""
    read -p "Are you sure you want to proceed with rollback? (yes/no): " -r
    if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        print_status "INFO" "Rollback cancelled by user"
        exit 0
    fi
}

# Execute rollback
execute_rollback() {
    if [ "$DRY_RUN" = true ]; then
        print_status "INFO" "DRY RUN: Would execute rollback now"
        return 0
    fi

    print_status "INFO" "Executing rollback..."

    case $DEPLOYMENT_TYPE in
        kubernetes)
            execute_kubernetes_rollback
            ;;
        docker)
            execute_docker_rollback
            ;;
        local)
            execute_local_rollback
            ;;
    esac
}

# Execute Kubernetes rollback
execute_kubernetes_rollback() {
    local namespace="ollama-$ENVIRONMENT"
    
    print_status "INFO" "Rolling back Kubernetes deployment..."
    
    if kubectl rollout undo deployment/ollama-distributed -n "$namespace" --to-revision="$ROLLBACK_REVISION"; then
        print_status "SUCCESS" "Rollback command executed"
        
        # Wait for rollback to complete
        print_status "INFO" "Waiting for rollback to complete..."
        if kubectl rollout status deployment/ollama-distributed -n "$namespace" --timeout=300s; then
            print_status "SUCCESS" "Kubernetes rollback completed"
        else
            print_status "ERROR" "Kubernetes rollback timed out"
            return 1
        fi
    else
        print_status "ERROR" "Kubernetes rollback failed"
        return 1
    fi
}

# Execute Docker rollback
execute_docker_rollback() {
    print_status "INFO" "Rolling back Docker deployment..."
    
    # Stop current container
    if [ -n "$CURRENT_CONTAINER" ]; then
        print_status "INFO" "Stopping current container: $CURRENT_CONTAINER"
        docker stop "$CURRENT_CONTAINER" || true
        docker rm "$CURRENT_CONTAINER" || true
    fi
    
    # Start rollback container
    print_status "INFO" "Starting rollback container with image: $ROLLBACK_IMAGE"
    if docker run -d --name "ollama-distributed-$ENVIRONMENT" \
        --restart unless-stopped \
        -p 8080:8080 \
        -e ENVIRONMENT="$ENVIRONMENT" \
        "$ROLLBACK_IMAGE"; then
        print_status "SUCCESS" "Docker rollback completed"
    else
        print_status "ERROR" "Docker rollback failed"
        return 1
    fi
}

# Execute local rollback
execute_local_rollback() {
    print_status "INFO" "Rolling back local deployment..."
    
    # Stop current process
    local pid_file="/var/run/ollama-distributed-$ENVIRONMENT.pid"
    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        print_status "INFO" "Stopping current process: $pid"
        kill "$pid" 2>/dev/null || true
        sleep 5
        kill -9 "$pid" 2>/dev/null || true
        rm -f "$pid_file"
    fi
    
    # Start rollback binary
    print_status "INFO" "Starting rollback binary: $ROLLBACK_BINARY"
    if nohup "$ROLLBACK_BINARY" start --environment "$ENVIRONMENT" > "/var/log/ollama-distributed-$ENVIRONMENT.log" 2>&1 &
    then
        echo $! > "$pid_file"
        print_status "SUCCESS" "Local rollback completed"
    else
        print_status "ERROR" "Local rollback failed"
        return 1
    fi
}

# Verify rollback
verify_rollback() {
    print_status "INFO" "Verifying rollback..."
    
    # Wait a moment for services to start
    sleep 10
    
    # Run health check
    if [ -f "scripts/health-check.sh" ]; then
        chmod +x scripts/health-check.sh
        if ./scripts/health-check.sh --environment "$ENVIRONMENT" --timeout 120; then
            print_status "SUCCESS" "Rollback verification passed"
            return 0
        else
            print_status "ERROR" "Rollback verification failed"
            return 1
        fi
    else
        print_status "WARNING" "Health check script not found, skipping verification"
        return 0
    fi
}

# Main function
main() {
    parse_args "$@"
    validate_args
    
    print_status "INFO" "Starting rollback for environment: $ENVIRONMENT"
    
    get_current_deployment
    get_rollback_target
    confirm_rollback
    execute_rollback
    
    if verify_rollback; then
        print_status "SUCCESS" "Rollback completed successfully! 🎉"
        
        # Send notification
        echo "📧 Rollback notification:"
        echo "  Environment: $ENVIRONMENT"
        echo "  Type: $DEPLOYMENT_TYPE"
        echo "  Status: SUCCESS"
        echo "  Time: $(date)"
        
        exit 0
    else
        print_status "ERROR" "Rollback verification failed!"
        exit 1
    fi
}

# Run main function
main "$@"

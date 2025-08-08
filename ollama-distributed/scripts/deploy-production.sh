#!/bin/bash

# Production Deployment Script for OllamaMax
# Automated deployment with validation and rollback capabilities

set -e

echo "🚀 OllamaMax Production Deployment"
echo "=================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
CLOUD_PROVIDER=""
CLUSTER_NAME="ollama-production"
NAMESPACE="ollama-production"
IMAGE_TAG="latest"
DRY_RUN=false
SKIP_VALIDATION=false
ENABLE_MONITORING=true
BACKUP_BEFORE_DEPLOY=true
AUTO_APPROVE=false

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
            --cloud)
                CLOUD_PROVIDER="$2"
                shift 2
                ;;
            --cluster)
                CLUSTER_NAME="$2"
                shift 2
                ;;
            --namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            --image-tag)
                IMAGE_TAG="$2"
                shift 2
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --skip-validation)
                SKIP_VALIDATION=true
                shift
                ;;
            --no-monitoring)
                ENABLE_MONITORING=false
                shift
                ;;
            --no-backup)
                BACKUP_BEFORE_DEPLOY=false
                shift
                ;;
            --auto-approve)
                AUTO_APPROVE=true
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
OllamaMax Production Deployment Script

Usage: $0 --cloud PROVIDER [OPTIONS]

Required:
    --cloud PROVIDER         Cloud provider (aws, gcp, azure)

Options:
    --cluster NAME           Cluster name (default: ollama-production)
    --namespace NAME         Kubernetes namespace (default: ollama-production)
    --image-tag TAG          Container image tag (default: latest)
    --dry-run               Perform dry run without actual deployment
    --skip-validation       Skip pre-deployment validation
    --no-monitoring         Skip monitoring stack deployment
    --no-backup             Skip backup before deployment
    --auto-approve          Auto-approve all prompts
    --help                  Show this help message

Examples:
    $0 --cloud aws --cluster my-cluster --image-tag v1.0.0
    $0 --cloud gcp --dry-run --no-backup
    $0 --cloud azure --auto-approve --enable-monitoring
EOF
}

# Validate prerequisites
validate_prerequisites() {
    print_status "INFO" "Validating prerequisites..."

    # Check required tools
    local required_tools=("kubectl" "terraform" "helm" "docker")
    for tool in "${required_tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            print_status "ERROR" "Required tool not found: $tool"
            exit 1
        fi
    done

    # Check cloud-specific tools
    case $CLOUD_PROVIDER in
        aws)
            if ! command -v aws &> /dev/null; then
                print_status "ERROR" "AWS CLI not found"
                exit 1
            fi
            ;;
        gcp)
            if ! command -v gcloud &> /dev/null; then
                print_status "ERROR" "Google Cloud SDK not found"
                exit 1
            fi
            ;;
        azure)
            if ! command -v az &> /dev/null; then
                print_status "ERROR" "Azure CLI not found"
                exit 1
            fi
            ;;
        *)
            print_status "ERROR" "Unsupported cloud provider: $CLOUD_PROVIDER"
            exit 1
            ;;
    esac

    print_status "SUCCESS" "Prerequisites validated"
}

# Validate cluster access
validate_cluster_access() {
    print_status "INFO" "Validating cluster access..."

    # Test kubectl connectivity
    if ! kubectl cluster-info &> /dev/null; then
        print_status "ERROR" "Cannot connect to Kubernetes cluster"
        print_status "INFO" "Please configure kubectl for your cluster:"
        case $CLOUD_PROVIDER in
            aws)
                echo "  aws eks update-kubeconfig --region <region> --name $CLUSTER_NAME"
                ;;
            gcp)
                echo "  gcloud container clusters get-credentials $CLUSTER_NAME --region <region>"
                ;;
            azure)
                echo "  az aks get-credentials --resource-group <rg> --name $CLUSTER_NAME"
                ;;
        esac
        exit 1
    fi

    # Check cluster version
    local k8s_version=$(kubectl version --short --client | grep -o 'v[0-9]\+\.[0-9]\+')
    print_status "INFO" "Kubernetes client version: $k8s_version"

    # Check cluster nodes
    local node_count=$(kubectl get nodes --no-headers | wc -l)
    print_status "INFO" "Cluster nodes: $node_count"

    if [ "$node_count" -lt 3 ]; then
        print_status "WARNING" "Cluster has fewer than 3 nodes. Consider scaling up for production."
    fi

    print_status "SUCCESS" "Cluster access validated"
}

# Create backup
create_backup() {
    if [ "$BACKUP_BEFORE_DEPLOY" = false ]; then
        print_status "INFO" "Skipping backup (--no-backup specified)"
        return
    fi

    print_status "INFO" "Creating backup before deployment..."

    local backup_dir="./backups/$(date +%Y%m%d-%H%M%S)"
    mkdir -p "$backup_dir"

    # Backup existing resources
    if kubectl get namespace "$NAMESPACE" &> /dev/null; then
        print_status "INFO" "Backing up existing resources..."
        
        kubectl get all -n "$NAMESPACE" -o yaml > "$backup_dir/resources.yaml"
        kubectl get configmaps -n "$NAMESPACE" -o yaml > "$backup_dir/configmaps.yaml"
        kubectl get secrets -n "$NAMESPACE" -o yaml > "$backup_dir/secrets.yaml"
        kubectl get pvc -n "$NAMESPACE" -o yaml > "$backup_dir/pvc.yaml"
        
        print_status "SUCCESS" "Backup created: $backup_dir"
    else
        print_status "INFO" "Namespace $NAMESPACE does not exist, skipping backup"
    fi
}

# Deploy infrastructure
deploy_infrastructure() {
    print_status "INFO" "Deploying infrastructure with Terraform..."

    local terraform_dir="./infrastructure/terraform/$CLOUD_PROVIDER"
    
    if [ ! -d "$terraform_dir" ]; then
        print_status "ERROR" "Terraform configuration not found: $terraform_dir"
        exit 1
    fi

    cd "$terraform_dir"

    # Initialize Terraform
    print_status "INFO" "Initializing Terraform..."
    terraform init

    # Plan deployment
    print_status "INFO" "Planning infrastructure deployment..."
    terraform plan \
        -var="cluster_name=$CLUSTER_NAME" \
        -var="enable_monitoring=$ENABLE_MONITORING" \
        -out=tfplan

    if [ "$DRY_RUN" = true ]; then
        print_status "INFO" "Dry run mode - skipping infrastructure deployment"
        cd - > /dev/null
        return
    fi

    # Apply infrastructure
    if [ "$AUTO_APPROVE" = true ]; then
        terraform apply -auto-approve tfplan
    else
        echo
        read -p "Apply infrastructure changes? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            terraform apply tfplan
        else
            print_status "INFO" "Infrastructure deployment cancelled"
            cd - > /dev/null
            return
        fi
    fi

    cd - > /dev/null
    print_status "SUCCESS" "Infrastructure deployed"
}

# Deploy application
deploy_application() {
    print_status "INFO" "Deploying OllamaMax application..."

    # Create namespace if it doesn't exist
    if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
        print_status "INFO" "Creating namespace: $NAMESPACE"
        kubectl create namespace "$NAMESPACE"
    fi

    # Apply production configuration
    print_status "INFO" "Applying production configuration..."
    
    local kustomize_dir="./deploy/kubernetes/production"
    
    if [ ! -d "$kustomize_dir" ]; then
        print_status "ERROR" "Kustomize configuration not found: $kustomize_dir"
        exit 1
    fi

    # Update image tag in kustomization
    cd "$kustomize_dir"
    
    # Create temporary kustomization with correct image tag
    cp kustomization.yaml kustomization.yaml.bak
    sed -i "s|newTag:.*|newTag: $IMAGE_TAG|g" kustomization.yaml

    if [ "$DRY_RUN" = true ]; then
        print_status "INFO" "Dry run mode - showing what would be applied:"
        kubectl kustomize . --namespace="$NAMESPACE"
        mv kustomization.yaml.bak kustomization.yaml
        cd - > /dev/null
        return
    fi

    # Apply the configuration
    kubectl apply -k . --namespace="$NAMESPACE"
    
    # Restore original kustomization
    mv kustomization.yaml.bak kustomization.yaml
    cd - > /dev/null

    print_status "SUCCESS" "Application deployed"
}

# Deploy monitoring
deploy_monitoring() {
    if [ "$ENABLE_MONITORING" = false ]; then
        print_status "INFO" "Skipping monitoring deployment (--no-monitoring specified)"
        return
    fi

    print_status "INFO" "Deploying monitoring stack..."

    # Add Prometheus Helm repository
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    helm repo update

    if [ "$DRY_RUN" = true ]; then
        print_status "INFO" "Dry run mode - skipping monitoring deployment"
        return
    fi

    # Install Prometheus stack
    helm upgrade --install prometheus prometheus-community/kube-prometheus-stack \
        --namespace monitoring \
        --create-namespace \
        --set grafana.adminPassword=admin123 \
        --set prometheus.prometheusSpec.retention=30d \
        --set prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage=100Gi

    print_status "SUCCESS" "Monitoring stack deployed"
}

# Validate deployment
validate_deployment() {
    if [ "$SKIP_VALIDATION" = true ]; then
        print_status "INFO" "Skipping deployment validation (--skip-validation specified)"
        return
    fi

    if [ "$DRY_RUN" = true ]; then
        print_status "INFO" "Dry run mode - skipping validation"
        return
    fi

    print_status "INFO" "Validating deployment..."

    # Wait for pods to be ready
    print_status "INFO" "Waiting for pods to be ready..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=ollama-distributed \
        --namespace="$NAMESPACE" --timeout=300s

    # Check service endpoints
    print_status "INFO" "Checking service endpoints..."
    local service_ip=$(kubectl get service ollama-loadbalancer -n "$NAMESPACE" \
        -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")
    
    if [ -z "$service_ip" ]; then
        service_ip=$(kubectl get service ollama-loadbalancer -n "$NAMESPACE" \
            -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null || echo "pending")
    fi

    print_status "INFO" "Service endpoint: $service_ip"

    # Test health endpoint
    if [ "$service_ip" != "pending" ] && [ -n "$service_ip" ]; then
        print_status "INFO" "Testing health endpoint..."
        if curl -f "http://$service_ip/health" &> /dev/null; then
            print_status "SUCCESS" "Health check passed"
        else
            print_status "WARNING" "Health check failed - service may still be starting"
        fi
    fi

    print_status "SUCCESS" "Deployment validation completed"
}

# Print deployment summary
print_summary() {
    print_status "SUCCESS" "Deployment completed successfully!"
    echo
    echo "📋 Deployment Summary:"
    echo "====================="
    echo "Cloud Provider: $CLOUD_PROVIDER"
    echo "Cluster Name: $CLUSTER_NAME"
    echo "Namespace: $NAMESPACE"
    echo "Image Tag: $IMAGE_TAG"
    echo "Monitoring: $([ "$ENABLE_MONITORING" = true ] && echo "Enabled" || echo "Disabled")"
    echo
    echo "🔗 Access Information:"
    echo "====================="
    
    if [ "$DRY_RUN" = false ]; then
        local service_ip=$(kubectl get service ollama-loadbalancer -n "$NAMESPACE" \
            -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || \
            kubectl get service ollama-loadbalancer -n "$NAMESPACE" \
            -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null || echo "pending")
        
        echo "Web UI: http://$service_ip:8081"
        echo "API: http://$service_ip/api/v1/proxy/status"
        
        if [ "$ENABLE_MONITORING" = true ]; then
            echo "Grafana: kubectl port-forward svc/prometheus-grafana 3000:80 -n monitoring"
        fi
    else
        echo "Dry run completed - no actual deployment performed"
    fi
    
    echo
    echo "📚 Next Steps:"
    echo "============="
    echo "1. Verify all pods are running: kubectl get pods -n $NAMESPACE"
    echo "2. Check service status: kubectl get services -n $NAMESPACE"
    echo "3. View logs: kubectl logs -f statefulset/ollama-distributed -n $NAMESPACE"
    echo "4. Test API: curl http://\$SERVICE_IP/health"
    echo
}

# Main function
main() {
    parse_args "$@"
    
    if [ -z "$CLOUD_PROVIDER" ]; then
        print_status "ERROR" "Cloud provider is required. Use --cloud PROVIDER"
        show_help
        exit 1
    fi

    print_status "INFO" "Starting OllamaMax production deployment"
    print_status "INFO" "Cloud Provider: $CLOUD_PROVIDER"
    print_status "INFO" "Cluster: $CLUSTER_NAME"
    print_status "INFO" "Namespace: $NAMESPACE"
    print_status "INFO" "Image Tag: $IMAGE_TAG"
    
    if [ "$DRY_RUN" = true ]; then
        print_status "INFO" "DRY RUN MODE - No actual changes will be made"
    fi
    
    echo

    validate_prerequisites
    validate_cluster_access
    create_backup
    deploy_infrastructure
    deploy_application
    deploy_monitoring
    validate_deployment
    print_summary
}

# Run main function
main "$@"

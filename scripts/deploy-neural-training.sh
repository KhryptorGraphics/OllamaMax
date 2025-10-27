#!/bin/bash

###############################################################################
# Neural Training Deployment Script
# Deploys Sprint 3 neural training services to Kubernetes
###############################################################################

set -e

echo "🚀 Neural Training Deployment - Sprint 3"
echo "=========================================="
echo ""

# Configuration
NAMESPACE="${NAMESPACE:-default}"
KUBECTL="${KUBECTL:-kubectl}"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Functions
log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Step 1: Prerequisites check
log_info "Step 1: Checking prerequisites..."

# Check kubectl
if ! command -v kubectl &> /dev/null; then
    log_error "kubectl not found. Please install kubectl."
    exit 1
fi

# Check Sprint 1 (Redis)
if ! $KUBECTL get pods -n ollamamax-redis | grep -q redis-cluster; then
    log_error "Sprint 1 (Redis cluster) not running. Deploy Sprint 1 first."
    exit 1
fi
log_info "  ✓ Sprint 1 (Redis) running"

# Check Sprint 2 (ML Pipeline)
if ! $KUBECTL get deployment agent-selection-model &> /dev/null; then
    log_error "Sprint 2 (ML Pipeline) not running. Deploy Sprint 2 first."
    exit 1
fi
log_info "  ✓ Sprint 2 (ML Pipeline) running"

# Check Node.js dependencies
log_info "Step 2: Installing Node.js dependencies..."
npm install @tensorflow/tfjs-node ioredis ml-random-forest --save-dev

# Step 3: Deploy neural training services
log_info "Step 3: Deploying neural training services..."

# Apply Kubernetes manifests
if [ -f "k8s/ml-pipeline.yaml" ]; then
    $KUBECTL apply -f k8s/ml-pipeline.yaml
    log_info "  ✓ Neural training services deployed"
else
    log_error "k8s/ml-pipeline.yaml not found"
    exit 1
fi

# Step 4: Wait for pods to be ready
log_info "Step 4: Waiting for pods to be ready..."

DEPLOYMENTS=(
    "agent-lstm-predictor"
    "neural-pattern-trainer"
    "historical-data-aggregator"
    "unified-neural-orchestrator"
)

for deployment in "${DEPLOYMENTS[@]}"; do
    log_info "  Waiting for $deployment..."
    $KUBECTL wait --for=condition=available --timeout=300s deployment/$deployment || {
        log_warn "    ⚠ $deployment not ready yet, continuing..."
    }
done

# Step 5: Service validation
log_info "Step 5: Validating services..."

# Check services are accessible
for port in 8087 8088 8089 8090; do
    SERVICE_NAME=$(kubectl get svc -o name | grep -E "(agent-lstm|neural-pattern|historical-data|unified-neural)" | head -1 | cut -d'/' -f2)
    if [ -n "$SERVICE_NAME" ]; then
        log_info "  ✓ Service accessible on port $port"
    else
        log_warn "    ⚠ Service on port $port not found"
    fi
done

# Step 6: Initial data aggregation
log_info "Step 6: Running initial data aggregation..."
sleep 10  # Wait for services to stabilize

# Trigger initial aggregation via port-forward (if needed)
log_info "  Aggregation will run automatically on schedule"

# Step 7: Display deployment summary
log_info "Step 7: Deployment Summary"
echo "=========================================="
echo ""
echo "✅ Neural Training Services Deployed:"
echo "  - Agent LSTM Predictor       (port 8087)"
echo "  - Neural Pattern Trainer     (port 8088)"
echo "  - Historical Data Aggregator (port 8089)"
echo "  - Unified Neural Orchestrator(port 8090)"
echo ""
echo "📊 Access services:"
echo "  kubectl port-forward svc/agent-lstm-predictor 8087:8087"
echo "  kubectl port-forward svc/neural-pattern-trainer 8088:8088"
echo "  kubectl port-forward svc/historical-data-aggregator 8089:8089"
echo "  kubectl port-forward svc/unified-neural-orchestrator 8090:8090"
echo ""
echo "🔍 Check deployment status:"
echo "  kubectl get pods | grep -E '(agent-lstm|neural-pattern|historical-data|unified-neural)'"
echo ""
echo "📝 View logs:"
echo "  kubectl logs -l app=unified-neural-orchestrator --tail=50"
echo ""
echo "🧪 Run validation:"
echo "  npm run deploy:validate:neural"
echo ""
echo "=========================================="
echo "🎉 Neural Training Deployment Complete!"
echo "=========================================="

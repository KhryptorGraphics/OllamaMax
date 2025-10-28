#!/bin/bash
# Auto-Scaling Validation Script
# References: load-test.js, k8s/ml-pipeline.yaml

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
REPORT_DIR="deployment-results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
REPORT_FILE="${REPORT_DIR}/autoscaling-${TIMESTAMP}.json"
NAMESPACE="${1:-ollamamax}"
DEPLOYMENT="${2:-ollamamax-api}"

mkdir -p "${REPORT_DIR}"

echo -e "${BLUE}=== Auto-Scaling Validation ===${NC}"

# Helper functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[PASS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[FAIL]${NC} $1"; }

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    log_error "kubectl not found - testing K8s HPA requires kubectl"
    exit 1
fi

# Phase 1: Check HPA Configuration
echo -e "\n${BLUE}=== Phase 1: Checking HPA Configuration ===${NC}"

# Initialize HPA_COUNT to 0
HPA_COUNT=0

if kubectl get hpa -n "${NAMESPACE}" &> /dev/null; then
    HPAS=$(kubectl get hpa -n "${NAMESPACE}" --no-headers | awk '{print $1}')
    HPA_COUNT=$(echo "$HPAS" | wc -l)
    log_success "Found ${HPA_COUNT} HPA(s) in namespace ${NAMESPACE}"

    for HPA in $HPAS; do
        log_info "HPA: ${HPA}"
        kubectl get hpa "${HPA}" -n "${NAMESPACE}"
    done
else
    log_warning "No HPAs found in namespace ${NAMESPACE}"
fi

# Phase 2: Get Initial Replica Count
echo -e "\n${BLUE}=== Phase 2: Recording Initial State ===${NC}"

INITIAL_REPLICAS=$(kubectl get deployment "${DEPLOYMENT}" -n "${NAMESPACE}" -o jsonpath='{.spec.replicas}' 2>/dev/null || echo "0")
log_info "Initial replicas: ${INITIAL_REPLICAS}"

# Parse CPU values, stripping 'm' suffix and converting to millicores
INITIAL_CPU=$(kubectl top pods -n "${NAMESPACE}" -l "app=${DEPLOYMENT}" --no-headers 2>/dev/null | awk '{gsub(/m/,"",$2); sum+=$2} END {print sum}' || echo "0")
log_info "Initial CPU usage: ${INITIAL_CPU}m"

# Phase 3: Generate Load
echo -e "\n${BLUE}=== Phase 3: Generating Load ===${NC}"

log_info "Generating load to trigger scaling..."

# Resolve actual service name (default to known service or deployment name)
SERVICE_NAME="${DEPLOYMENT}-svc"
if ! kubectl get svc "${SERVICE_NAME}" -n "${NAMESPACE}" &> /dev/null; then
    # Try deployment name directly
    SERVICE_NAME="${DEPLOYMENT}"
    if ! kubectl get svc "${SERVICE_NAME}" -n "${NAMESPACE}" &> /dev/null; then
        log_warning "Service not found, using ${DEPLOYMENT} as fallback"
    fi
fi

log_info "Using service: ${SERVICE_NAME}"

# Use kubectl run to create a load generator pod
kubectl run load-generator \
    --image=busybox:latest \
    --restart=Never \
    --namespace="${NAMESPACE}" \
    --command -- /bin/sh -c "while true; do wget -q -O- http://${SERVICE_NAME}:8080/health; done" &> /dev/null &

LOAD_PID=$!
log_success "Load generator started (PID: ${LOAD_PID})"

# Phase 4: Monitor Scaling
echo -e "\n${BLUE}=== Phase 4: Monitoring Auto-Scaling ===${NC}"

SCALING_EVENTS=()
MAX_WAIT=300
ELAPSED=0
PREV_REPLICAS=$INITIAL_REPLICAS

log_info "Monitoring for ${MAX_WAIT}s..."

while [ $ELAPSED -lt $MAX_WAIT ]; do
    sleep 10
    ELAPSED=$((ELAPSED + 10))

    CURRENT_REPLICAS=$(kubectl get deployment "${DEPLOYMENT}" -n "${NAMESPACE}" -o jsonpath='{.spec.replicas}' 2>/dev/null || echo "0")
    CURRENT_CPU=$(kubectl top pods -n "${NAMESPACE}" -l "app=${DEPLOYMENT}" --no-headers 2>/dev/null | awk '{gsub(/m/,"",$2); sum+=$2} END {print sum}' || echo "0")

    log_info "[${ELAPSED}s] Replicas: ${CURRENT_REPLICAS}, CPU: ${CURRENT_CPU}m"

    # Check if scaled up (compare to previous value before updating)
    if [ "$CURRENT_REPLICAS" -gt "$PREV_REPLICAS" ]; then
        log_success "Scale-up detected: ${PREV_REPLICAS} -> ${CURRENT_REPLICAS}"
        SCALING_EVENTS+=("${ELAPSED}s:${PREV_REPLICAS}->${CURRENT_REPLICAS}")
        PREV_REPLICAS=$CURRENT_REPLICAS
        break
    fi

    # Record any scaling event
    if [ "$CURRENT_REPLICAS" != "$PREV_REPLICAS" ]; then
        SCALING_EVENTS+=("${ELAPSED}s:${PREV_REPLICAS}->${CURRENT_REPLICAS}")
        log_success "Scaling event detected: ${PREV_REPLICAS} -> ${CURRENT_REPLICAS}"
        PREV_REPLICAS=$CURRENT_REPLICAS
    fi
done

# Phase 5: Stop Load and Monitor Scale-Down
echo -e "\n${BLUE}=== Phase 5: Monitoring Scale-Down ===${NC}"

log_info "Stopping load generator..."
kubectl delete pod load-generator -n "${NAMESPACE}" --ignore-not-found=true &> /dev/null
kill $LOAD_PID &> /dev/null || true
log_success "Load generator stopped"

log_info "Monitoring scale-down for 120s..."
SCALEDOWN_WAIT=120
ELAPSED=0

while [ $ELAPSED -lt $SCALEDOWN_WAIT ]; do
    sleep 10
    ELAPSED=$((ELAPSED + 10))

    CURRENT_REPLICAS=$(kubectl get deployment "${DEPLOYMENT}" -n "${NAMESPACE}" -o jsonpath='{.spec.replicas}' 2>/dev/null || echo "0")
    log_info "[${ELAPSED}s] Replicas: ${CURRENT_REPLICAS}"

    if [ "$CURRENT_REPLICAS" -lt "$INITIAL_REPLICAS" ]; then
        SCALING_EVENTS+=("${ELAPSED}s:${INITIAL_REPLICAS}->${CURRENT_REPLICAS}")
        log_success "Scale-down detected: ${INITIAL_REPLICAS} -> ${CURRENT_REPLICAS}"
        break
    fi
done

# Phase 6: Test Scaling Limits
echo -e "\n${BLUE}=== Phase 6: Validating Scaling Constraints ===${NC}"

if kubectl get hpa -n "${NAMESPACE}" &> /dev/null; then
    for HPA in $HPAS; do
        MIN_REPLICAS=$(kubectl get hpa "${HPA}" -n "${NAMESPACE}" -o jsonpath='{.spec.minReplicas}')
        MAX_REPLICAS=$(kubectl get hpa "${HPA}" -n "${NAMESPACE}" -o jsonpath='{.spec.maxReplicas}')

        log_info "HPA ${HPA}: min=${MIN_REPLICAS}, max=${MAX_REPLICAS}"

        FINAL_REPLICAS=$(kubectl get deployment "${DEPLOYMENT}" -n "${NAMESPACE}" -o jsonpath='{.spec.replicas}' 2>/dev/null || echo "0")

        if [ "$FINAL_REPLICAS" -ge "$MIN_REPLICAS" ] && [ "$FINAL_REPLICAS" -le "$MAX_REPLICAS" ]; then
            log_success "Replica count within constraints: ${FINAL_REPLICAS}"
        else
            log_warning "Replica count outside constraints: ${FINAL_REPLICAS}"
        fi
    done
fi

# Generate Report
EVENTS_JSON=$(IFS=,; for event in "${SCALING_EVENTS[@]}"; do echo "\"$event\""; done | paste -sd,)

# Guard against uninitialized HPA_COUNT
HPA_COUNT=${HPA_COUNT:-0}

cat > "${REPORT_FILE}" <<EOF
{
  "timestamp": "${TIMESTAMP}",
  "namespace": "${NAMESPACE}",
  "deployment": "${DEPLOYMENT}",
  "hpa_count": ${HPA_COUNT},
  "scaling_events": [
    ${EVENTS_JSON}
  ],
  "total_events": ${#SCALING_EVENTS[@]},
  "status": "completed"
}
EOF

log_success "Report generated: ${REPORT_FILE}"
echo -e "\n${GREEN}Auto-scaling validation completed${NC}"

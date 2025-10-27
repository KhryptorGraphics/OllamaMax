#!/bin/bash
# Kubernetes Deployment Validation Script
# References: scripts/deploy-sprint1.sh, scripts/deploy-sprint2.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
K8S_DIR="${1:-k8s}"
TIMEOUT=600
REPORT_DIR="deployment-results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
REPORT_FILE="${REPORT_DIR}/k8s-validation-${TIMESTAMP}.json"
REPORT_MD="${REPORT_DIR}/k8s-validation-${TIMESTAMP}.md"

# Create report directory
mkdir -p "${REPORT_DIR}"

# Initialize report
VALIDATION_RESULTS=()
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNING_CHECKS=0

echo -e "${BLUE}=== Kubernetes Deployment Validation ===${NC}"
echo "K8s Directory: ${K8S_DIR}"
echo "Timestamp: ${TIMESTAMP}"
echo ""

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
    WARNING_CHECKS=$((WARNING_CHECKS + 1))
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
}

add_result() {
    local name="$1"
    local status="$2"
    local message="$3"
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    VALIDATION_RESULTS+=("{\"name\":\"$name\",\"status\":\"$status\",\"message\":\"$message\"}")
}

# Check CLI prerequisites
check_prerequisites() {
    local missing_tools=()

    # Required tools
    command -v kubectl &> /dev/null || missing_tools+=("kubectl")

    # Optional tools
    command -v jq &> /dev/null || log_warning "jq not available - some checks may be limited"

    if [ ${#missing_tools[@]} -gt 0 ]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        log_info "Please install missing tools before running this script"
        exit 1
    fi

    log_info "Required tools available"
}

check_prerequisites

# 1. Pre-Deployment Validation
echo -e "${BLUE}=== Phase 1: Pre-Deployment Validation ===${NC}"

# Check kubectl connectivity
log_info "Checking kubectl connectivity..."
if kubectl cluster-info &> /dev/null; then
    CLUSTER_INFO=$(kubectl cluster-info | head -1)
    log_success "Connected to cluster"
    add_result "Cluster Connectivity" "pass" "Connected to K8s cluster"
else
    log_error "Cannot connect to Kubernetes cluster"
    add_result "Cluster Connectivity" "fail" "Connection failed"
    exit 1
fi

# Check kubectl version
log_info "Checking kubectl version..."
KUBECTL_VERSION=$(kubectl version --client --short 2>/dev/null | awk '{print $3}')
log_success "kubectl version: ${KUBECTL_VERSION}"
add_result "kubectl Version" "pass" "kubectl ${KUBECTL_VERSION}"

# Check cluster resources
log_info "Checking cluster resources..."
NODES=$(kubectl get nodes --no-headers | wc -l)
READY_NODES=$(kubectl get nodes --no-headers | grep -c " Ready" || true)

if [ "$NODES" -eq "$READY_NODES" ]; then
    log_success "All ${NODES} node(s) ready"
    add_result "Node Status" "pass" "${NODES} nodes ready"
else
    log_warning "${READY_NODES}/${NODES} nodes ready"
    add_result "Node Status" "warning" "Only ${READY_NODES}/${NODES} nodes ready"
fi

# Validate manifest syntax
log_info "Validating manifest syntax..."
MANIFEST_ERRORS=0
for MANIFEST in "${K8S_DIR}"/*.yaml; do
    if [ -f "$MANIFEST" ]; then
        if kubectl apply --dry-run=client -f "$MANIFEST" &> /dev/null; then
            log_success "Valid manifest: $(basename "$MANIFEST")"
        else
            log_error "Invalid manifest: $(basename "$MANIFEST")"
            MANIFEST_ERRORS=$((MANIFEST_ERRORS + 1))
        fi
    fi
done

if [ "$MANIFEST_ERRORS" -eq 0 ]; then
    add_result "Manifest Validation" "pass" "All manifests valid"
else
    add_result "Manifest Validation" "fail" "${MANIFEST_ERRORS} invalid manifest(s)"
    exit 1
fi

# Check required namespaces
log_info "Checking namespaces..."
NAMESPACES=("ollamamax-redis" "ollamamax-timeseries" "ollamamax-ml" "ollamamax-monitoring" "ollamamax")
for NS in "${NAMESPACES[@]}"; do
    if kubectl get namespace "$NS" &> /dev/null; then
        log_success "Namespace exists: ${NS}"
        add_result "Namespace: ${NS}" "pass" "Exists"
    else
        log_info "Creating namespace: ${NS}"
        kubectl create namespace "$NS"
        add_result "Namespace: ${NS}" "pass" "Created"
    fi
done

# 2. Deployment Execution
echo -e "\n${BLUE}=== Phase 2: Deployment Execution ===${NC}"

START_TIME=$(date +%s)

# Apply manifests in order
log_info "Applying manifests..."
MANIFESTS_ORDER=(
    "redis-cluster.yaml"
    "timeseries-db.yaml"
    "ml-pipeline.yaml"
    "monitoring-stack.yaml"
)

for MANIFEST in "${MANIFESTS_ORDER[@]}"; do
    MANIFEST_PATH="${K8S_DIR}/${MANIFEST}"
    if [ -f "$MANIFEST_PATH" ]; then
        log_info "Applying ${MANIFEST}..."
        if kubectl apply -f "$MANIFEST_PATH"; then
            log_success "Applied ${MANIFEST}"
            add_result "Apply: ${MANIFEST}" "pass" "Applied successfully"
        else
            log_error "Failed to apply ${MANIFEST}"
            add_result "Apply: ${MANIFEST}" "fail" "Application failed"
        fi
    fi
done

# Wait for StatefulSets
log_info "Waiting for StatefulSets to be ready..."
STATEFULSETS=$(kubectl get statefulsets --all-namespaces -o jsonpath='{range .items[*]}{.metadata.namespace}/{.metadata.name}{"\n"}{end}')

for STS in $STATEFULSETS; do
    NS=$(echo "$STS" | cut -d/ -f1)
    NAME=$(echo "$STS" | cut -d/ -f2)

    log_info "Waiting for StatefulSet ${NAME} in namespace ${NS}..."
    if kubectl rollout status statefulset/"${NAME}" -n "${NS}" --timeout="${TIMEOUT}s" &> /dev/null; then
        log_success "StatefulSet ${NAME} ready"
        add_result "StatefulSet: ${NAME}" "pass" "Ready"
    else
        log_error "StatefulSet ${NAME} not ready"
        add_result "StatefulSet: ${NAME}" "fail" "Not ready after ${TIMEOUT}s"
    fi
done

# Wait for Deployments
log_info "Waiting for Deployments to be ready..."
DEPLOYMENTS=$(kubectl get deployments --all-namespaces -o jsonpath='{range .items[*]}{.metadata.namespace}/{.metadata.name}{"\n"}{end}')

for DEPLOY in $DEPLOYMENTS; do
    NS=$(echo "$DEPLOY" | cut -d/ -f1)
    NAME=$(echo "$DEPLOY" | cut -d/ -f2)

    log_info "Waiting for Deployment ${NAME} in namespace ${NS}..."
    if kubectl rollout status deployment/"${NAME}" -n "${NS}" --timeout="${TIMEOUT}s" &> /dev/null; then
        log_success "Deployment ${NAME} ready"
        add_result "Deployment: ${NAME}" "pass" "Ready"
    else
        log_error "Deployment ${NAME} not ready"
        add_result "Deployment: ${NAME}" "fail" "Not ready after ${TIMEOUT}s"
    fi
done

END_TIME=$(date +%s)
DEPLOYMENT_TIME=$((END_TIME - START_TIME))
log_info "Deployment time: ${DEPLOYMENT_TIME}s"

# 3. Post-Deployment Validation
echo -e "\n${BLUE}=== Phase 3: Post-Deployment Validation ===${NC}"

# Verify all pods are running
log_info "Verifying pod status..."
TOTAL_PODS=0
RUNNING_PODS=0
FAILED_PODS=0

for NS in "${NAMESPACES[@]}"; do
    PODS=$(kubectl get pods -n "$NS" --no-headers 2>/dev/null || true)
    if [ -n "$PODS" ]; then
        while IFS= read -r POD_LINE; do
            TOTAL_PODS=$((TOTAL_PODS + 1))
            POD_NAME=$(echo "$POD_LINE" | awk '{print $1}')
            POD_STATUS=$(echo "$POD_LINE" | awk '{print $3}')

            if [ "$POD_STATUS" = "Running" ]; then
                RUNNING_PODS=$((RUNNING_PODS + 1))
                log_success "Pod ${POD_NAME}: Running"
            else
                FAILED_PODS=$((FAILED_PODS + 1))
                log_error "Pod ${POD_NAME}: ${POD_STATUS}"
            fi
        done <<< "$PODS"
    fi
done

add_result "Pod Status" "pass" "${RUNNING_PODS}/${TOTAL_PODS} pods running"

# Check service endpoints
log_info "Checking service endpoints..."
SERVICES=$(kubectl get services --all-namespaces -o jsonpath='{range .items[*]}{.metadata.namespace}/{.metadata.name}{"\n"}{end}')

SERVICE_COUNT=0
READY_SERVICES=0

for SVC in $SERVICES; do
    NS=$(echo "$SVC" | cut -d/ -f1)
    NAME=$(echo "$SVC" | cut -d/ -f2)
    SERVICE_COUNT=$((SERVICE_COUNT + 1))

    ENDPOINTS=$(kubectl get endpoints "$NAME" -n "$NS" -o jsonpath='{.subsets[*].addresses[*].ip}' 2>/dev/null || echo "")

    if [ -n "$ENDPOINTS" ]; then
        READY_SERVICES=$((READY_SERVICES + 1))
        log_success "Service ${NAME} has endpoints"
    else
        log_warning "Service ${NAME} has no endpoints"
    fi
done

add_result "Service Endpoints" "pass" "${READY_SERVICES}/${SERVICE_COUNT} services have endpoints"

# Verify PVC status
log_info "Verifying PersistentVolumeClaims..."
PVCS=$(kubectl get pvc --all-namespaces -o jsonpath='{range .items[*]}{.metadata.namespace}/{.metadata.name}/{.status.phase}{"\n"}{end}')

PVC_COUNT=0
BOUND_PVCS=0

for PVC_INFO in $PVCS; do
    PVC_COUNT=$((PVC_COUNT + 1))
    NS=$(echo "$PVC_INFO" | cut -d/ -f1)
    NAME=$(echo "$PVC_INFO" | cut -d/ -f2)
    PHASE=$(echo "$PVC_INFO" | cut -d/ -f3)

    if [ "$PHASE" = "Bound" ]; then
        BOUND_PVCS=$((BOUND_PVCS + 1))
        log_success "PVC ${NAME}: Bound"
    else
        log_warning "PVC ${NAME}: ${PHASE}"
    fi
done

if [ "$PVC_COUNT" -eq "$BOUND_PVCS" ]; then
    add_result "PVC Status" "pass" "All ${PVC_COUNT} PVCs bound"
else
    add_result "PVC Status" "warning" "Only ${BOUND_PVCS}/${PVC_COUNT} PVCs bound"
fi

# 4. Integration Testing
echo -e "\n${BLUE}=== Phase 4: Integration Testing ===${NC}"

# Test service discovery (DNS)
log_info "Testing service discovery..."
TEST_POD=$(kubectl get pods -n ollamamax-redis --no-headers | head -1 | awk '{print $1}')

if [ -n "$TEST_POD" ]; then
    if kubectl exec -n ollamamax-redis "$TEST_POD" -- nslookup redis-cluster &> /dev/null; then
        log_success "Service discovery working"
        add_result "Service Discovery" "pass" "DNS resolution working"
    else
        log_warning "Service discovery may have issues"
        add_result "Service Discovery" "warning" "DNS resolution issues"
    fi
fi

# 5. Performance Validation
echo -e "\n${BLUE}=== Phase 5: Performance Validation ===${NC}"

# Check resource requests/limits
log_info "Verifying resource configuration..."
if command -v jq &> /dev/null; then
    PODS_WITH_LIMITS=$(kubectl get pods --all-namespaces -o json | jq '[.items[].spec.containers[] | select(.resources.limits != null)] | length' 2>/dev/null || echo "0")
    TOTAL_CONTAINERS=$(kubectl get pods --all-namespaces -o json | jq '[.items[].spec.containers[]] | length' 2>/dev/null || echo "0")

    if [ "$PODS_WITH_LIMITS" -gt 0 ] && [ "$TOTAL_CONTAINERS" -gt 0 ]; then
        log_success "Resource limits configured on ${PODS_WITH_LIMITS}/${TOTAL_CONTAINERS} containers"
        add_result "Resource Limits" "pass" "${PODS_WITH_LIMITS}/${TOTAL_CONTAINERS} containers have limits"
    else
        log_warning "No resource limits configured"
        add_result "Resource Limits" "warning" "No resource limits"
    fi
else
    log_warning "jq not available - skipping detailed resource limits check"
    # Fallback: Basic check using grep
    PODS_JSON=$(kubectl get pods --all-namespaces -o json 2>/dev/null)
    if echo "$PODS_JSON" | grep -q '"limits"'; then
        log_success "Resource limits configured (basic check)"
        add_result "Resource Limits" "pass" "Resource limits found (jq unavailable for detailed count)"
    else
        log_warning "No resource limits found in basic check"
        add_result "Resource Limits" "warning" "Resource limits check skipped (jq not available)"
    fi
fi

# Check HPA configurations
log_info "Checking HPA configurations..."
HPAS=$(kubectl get hpa --all-namespaces --no-headers 2>/dev/null | wc -l || echo "0")

if [ "$HPAS" -gt 0 ]; then
    log_success "${HPAS} HPA(s) configured"
    add_result "HPA Configuration" "pass" "${HPAS} HPAs configured"
else
    log_warning "No HPAs configured"
    add_result "HPA Configuration" "warning" "No HPAs configured"
fi

# 6. Generate Report
echo -e "\n${BLUE}=== Phase 6: Generating Report ===${NC}"

# Calculate validation score
VALIDATION_SCORE=$((PASSED_CHECKS * 100 / TOTAL_CHECKS))

# JSON Report
cat > "${REPORT_FILE}" <<EOF
{
  "timestamp": "${TIMESTAMP}",
  "k8s_directory": "${K8S_DIR}",
  "deployment_time_seconds": ${DEPLOYMENT_TIME},
  "summary": {
    "total_checks": ${TOTAL_CHECKS},
    "passed": ${PASSED_CHECKS},
    "failed": ${FAILED_CHECKS},
    "warnings": ${WARNING_CHECKS},
    "validation_score": ${VALIDATION_SCORE}
  },
  "cluster": {
    "nodes": ${NODES},
    "ready_nodes": ${READY_NODES}
  },
  "resources": {
    "total_pods": ${TOTAL_PODS},
    "running_pods": ${RUNNING_PODS},
    "failed_pods": ${FAILED_PODS},
    "services": ${SERVICE_COUNT},
    "ready_services": ${READY_SERVICES},
    "pvcs": ${PVC_COUNT},
    "bound_pvcs": ${BOUND_PVCS}
  },
  "results": [
    $(IFS=,; echo "${VALIDATION_RESULTS[*]}")
  ]
}
EOF

# Markdown Report
cat > "${REPORT_MD}" <<EOF
# Kubernetes Deployment Validation Report

**Generated:** ${TIMESTAMP}
**K8s Directory:** ${K8S_DIR}
**Deployment Time:** ${DEPLOYMENT_TIME}s

## Summary

- **Validation Score:** ${VALIDATION_SCORE}/100
- **Total Checks:** ${TOTAL_CHECKS}
- **Passed:** ${PASSED_CHECKS} ✅
- **Failed:** ${FAILED_CHECKS} ❌
- **Warnings:** ${WARNING_CHECKS} ⚠️

## Cluster Status

- **Nodes:** ${READY_NODES}/${NODES} ready

## Resource Status

- **Pods:** ${RUNNING_PODS}/${TOTAL_PODS} running
- **Services:** ${READY_SERVICES}/${SERVICE_COUNT} with endpoints
- **PVCs:** ${BOUND_PVCS}/${PVC_COUNT} bound

## Validation Results

$(if command -v jq &> /dev/null; then
    for result in "${VALIDATION_RESULTS[@]}"; do
        name=$(echo "$result" | jq -r '.name' 2>/dev/null || echo "unknown")
        status=$(echo "$result" | jq -r '.status' 2>/dev/null || echo "unknown")
        message=$(echo "$result" | jq -r '.message' 2>/dev/null || echo "")

        case "$status" in
            "pass") echo "✅ **${name}:** ${message}" ;;
            "fail") echo "❌ **${name}:** ${message}" ;;
            "warning") echo "⚠️  **${name}:** ${message}" ;;
        esac
    done
else
    # Fallback without jq - print raw JSON results
    echo "Raw validation results (jq not available for formatting):"
    echo '```json'
    printf '%s\n' "${VALIDATION_RESULTS[@]}"
    echo '```'
fi)

## Conclusion

$(if [ "$FAILED_CHECKS" -eq 0 ]; then
    echo "✅ **Deployment validation PASSED**"
else
    echo "❌ **Deployment validation FAILED** - ${FAILED_CHECKS} critical issue(s) found"
fi)

EOF

log_success "Reports generated:"
log_info "  JSON: ${REPORT_FILE}"
log_info "  Markdown: ${REPORT_MD}"

# Print summary
echo -e "\n${BLUE}=== Validation Summary ===${NC}"
echo "Total Checks: ${TOTAL_CHECKS}"
echo -e "Passed: ${GREEN}${PASSED_CHECKS}${NC}"
echo -e "Failed: ${RED}${FAILED_CHECKS}${NC}"
echo -e "Warnings: ${YELLOW}${WARNING_CHECKS}${NC}"
echo "Validation Score: ${VALIDATION_SCORE}/100"

# Exit with appropriate code
if [ "$FAILED_CHECKS" -gt 0 ]; then
    echo -e "\n${RED}Deployment validation FAILED${NC}"
    exit 1
else
    echo -e "\n${GREEN}Deployment validation PASSED${NC}"
    exit 0
fi

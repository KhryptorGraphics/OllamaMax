#!/bin/bash
# Master Deployment Validation Orchestrator
# References: scripts/run-training-tests.sh
#
# Usage: ./run-deployment-validation.sh [ENVIRONMENT] [DEPLOYMENT_TYPE] [OPTIONS]
#   ENVIRONMENT: local (default), staging, production
#   DEPLOYMENT_TYPE: docker, kubernetes, both (default)
#   OPTIONS:
#     --skip-phase PHASE_NAME  Skip specific phase (can be repeated)
#
# Example: ./run-deployment-validation.sh local both --skip-phase "Multi-Region Simulation" --skip-phase "Auto-Scaling Tests"

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
ENVIRONMENT="${1:-local}"
DEPLOYMENT_TYPE="${2:-both}"
shift 2 2>/dev/null || shift $# 2>/dev/null  # Remove first two args if they exist

# Parse skip-phase options
declare -a SKIP_PHASES
while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-phase)
            SKIP_PHASES+=("$2")
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            shift
            ;;
    esac
done

REPORT_DIR="deployment-results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
FINAL_REPORT="${REPORT_DIR}/deployment-validation-${TIMESTAMP}.md"

mkdir -p "${REPORT_DIR}"

echo -e "${BLUE}╔══════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Deployment Validation Orchestrator     ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════╝${NC}"
echo ""
echo "Environment: ${ENVIRONMENT}"
echo "Deployment Type: ${DEPLOYMENT_TYPE}"
echo "Timestamp: ${TIMESTAMP}"
if [ ${#SKIP_PHASES[@]} -gt 0 ]; then
    echo "Skipping phases: ${SKIP_PHASES[*]}"
fi
echo ""

# Helper functions
log_phase() {
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════${NC}"
}

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[PASS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[FAIL]${NC} $1"; }

# Track results
PHASES_PASSED=0
PHASES_FAILED=0
PHASES_SKIPPED=0
PHASE_RESULTS=()

execute_phase() {
    local phase_name="$1"
    local script_path="$2"
    local required="$3"

    # Check if phase should be skipped
    for skip in "${SKIP_PHASES[@]}"; do
        if [ "$skip" = "$phase_name" ]; then
            log_warning "Skipping phase (user requested): ${phase_name}"
            PHASES_SKIPPED=$((PHASES_SKIPPED + 1))
            PHASE_RESULTS+=("⏭️  ${phase_name}: SKIPPED (user requested)")
            return 0
        fi
    done

    log_phase "Phase: ${phase_name}"

    if [ ! -f "$script_path" ]; then
        log_warning "Script not found: ${script_path}"
        PHASES_SKIPPED=$((PHASES_SKIPPED + 1))
        PHASE_RESULTS+=("❌ ${phase_name}: SKIPPED (script not found)")
        return 1
    fi

    log_info "Executing: ${script_path}"

    if bash "$script_path"; then
        log_success "Phase completed: ${phase_name}"
        PHASES_PASSED=$((PHASES_PASSED + 1))
        PHASE_RESULTS+=("✅ ${phase_name}: PASSED")
        return 0
    else
        log_error "Phase failed: ${phase_name}"
        PHASES_FAILED=$((PHASES_FAILED + 1))
        PHASE_RESULTS+=("❌ ${phase_name}: FAILED")

        if [ "$required" = "true" ]; then
            log_error "Critical phase failed - stopping validation"
            return 1
        fi
        return 0
    fi
}

# Start validation
START_TIME=$(date +%s)

# Phase 1: Pre-deployment Checks
execute_phase "Pre-deployment Checks" "scripts/validate-docker-deployment-precheck.sh" "false" || true

# Phase 2: Docker Deployment Validation
if [ "$DEPLOYMENT_TYPE" = "docker" ] || [ "$DEPLOYMENT_TYPE" = "both" ]; then
    execute_phase "Docker Deployment" "scripts/validate-docker-deployment.sh" "false" || true
fi

# Phase 3: Kubernetes Deployment Validation
if [ "$DEPLOYMENT_TYPE" = "kubernetes" ] || [ "$DEPLOYMENT_TYPE" = "both" ]; then
    execute_phase "Kubernetes Deployment" "scripts/validate-k8s-deployment.sh" "false" || true
fi

# Phase 4: Multi-region Simulation
execute_phase "Multi-Region Simulation" "scripts/simulate-multi-region.sh" "false" || true

# Phase 5: Auto-scaling Tests
execute_phase "Auto-Scaling Tests" "scripts/test-autoscaling.sh" "false" || true

# Phase 6: Load Balancing Tests
execute_phase "Load Balancing Tests" "scripts/test-load-balancing.sh" "false" || true

# Phase 7: Security Validation
execute_phase "Security Validation" "scripts/validate-security.sh" "false" || true

# Phase 8: Health Checks
if [ -f "ollama-distributed/scripts/health-check.sh" ]; then
    execute_phase "Health Checks" "ollama-distributed/scripts/health-check.sh" "false" || true
fi

# Phase 9: Rollback Testing
execute_phase "Rollback Testing" "scripts/test-rollback.sh" "false" || true

# Calculate total time
END_TIME=$(date +%s)
TOTAL_TIME=$((END_TIME - START_TIME))

# Calculate readiness score
TOTAL_PHASES=$((PHASES_PASSED + PHASES_FAILED + PHASES_SKIPPED))
if [ $TOTAL_PHASES -gt 0 ]; then
    READINESS_SCORE=$((PHASES_PASSED * 100 / TOTAL_PHASES))
else
    READINESS_SCORE=0
fi

# Generate Final Report
cat > "${FINAL_REPORT}" <<EOF
# Deployment Validation Report

**Generated:** $(date)
**Environment:** ${ENVIRONMENT}
**Deployment Type:** ${DEPLOYMENT_TYPE}
**Execution Time:** ${TOTAL_TIME}s

## Executive Summary

- **Readiness Score:** ${READINESS_SCORE}/100
- **Phases Passed:** ${PHASES_PASSED}
- **Phases Failed:** ${PHASES_FAILED}
- **Phases Skipped:** ${PHASES_SKIPPED}

## Phase Results

$(for result in "${PHASE_RESULTS[@]}"; do
    echo "- $result"
done)

## Recommendation

$(if [ $READINESS_SCORE -ge 80 ]; then
    echo "✅ **READY FOR DEPLOYMENT** - All critical validations passed"
elif [ $READINESS_SCORE -ge 60 ]; then
    echo "⚠️  **DEPLOY WITH CAUTION** - Some validations failed, review issues"
else
    echo "❌ **NOT READY FOR DEPLOYMENT** - Critical issues detected"
fi)

## Detailed Results

Individual phase reports are available in the \`${REPORT_DIR}\` directory.

## Next Steps

$(if [ $PHASES_FAILED -gt 0 ]; then
    echo "1. Review failed phase reports"
    echo "2. Address critical issues"
    echo "3. Re-run validation"
    echo "4. Update deployment procedures"
else
    echo "1. Review individual phase reports"
    echo "2. Document any warnings"
    echo "3. Proceed with deployment"
    echo "4. Monitor deployment metrics"
fi)

---

**Report Location:** ${FINAL_REPORT}
**Artifacts Directory:** ${REPORT_DIR}
EOF

# Display Summary
echo ""
log_phase "Validation Summary"
echo ""
echo "Total Phases: ${TOTAL_PHASES}"
echo -e "Passed: ${GREEN}${PHASES_PASSED}${NC}"
echo -e "Failed: ${RED}${PHASES_FAILED}${NC}"
echo -e "Skipped: ${YELLOW}${PHASES_SKIPPED}${NC}"
echo ""
echo "Readiness Score: ${READINESS_SCORE}/100"
echo "Total Execution Time: ${TOTAL_TIME}s"
echo ""

# Display results
echo "Phase Results:"
for result in "${PHASE_RESULTS[@]}"; do
    echo "  $result"
done

echo ""
log_success "Final report generated: ${FINAL_REPORT}"
echo ""

# Exit code based on readiness score
if [ $READINESS_SCORE -ge 80 ]; then
    echo -e "${GREEN}✅ DEPLOYMENT VALIDATION PASSED${NC}"
    echo -e "${GREEN}   Ready for deployment${NC}"
    exit 0
elif [ $READINESS_SCORE -ge 60 ]; then
    echo -e "${YELLOW}⚠️  DEPLOYMENT VALIDATION PASSED WITH WARNINGS${NC}"
    echo -e "${YELLOW}   Review warnings before deployment${NC}"
    exit 0
else
    echo -e "${RED}❌ DEPLOYMENT VALIDATION FAILED${NC}"
    echo -e "${RED}   Address critical issues before deployment${NC}"
    exit 1
fi

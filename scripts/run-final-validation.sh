#!/bin/bash

################################################################################
# Master Final Validation Orchestrator for OllamaMax
#
# Coordinates all final validation activities with phase-by-phase execution
# Generates comprehensive production readiness assessment
################################################################################

set -e

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
LOG_FILE="final-validation-${TIMESTAMP}.log"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Command line options
PHASES_TO_RUN="all"
SKIP_PHASES=""
PARALLEL_MODE=false
DRY_RUN=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --phases)
            PHASES_TO_RUN="$2"
            shift 2
            ;;
        --skip)
            SKIP_PHASES="$2"
            shift 2
            ;;
        --parallel)
            PARALLEL_MODE=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--phases \"phase1,phase2\"] [--skip \"phase3\"] [--parallel] [--dry-run]"
            exit 1
            ;;
    esac
done

# Logging functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1" | tee -a "${LOG_FILE}"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "${LOG_FILE}"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "${LOG_FILE}"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1" | tee -a "${LOG_FILE}"; }
log_phase() { echo -e "${MAGENTA}[PHASE]${NC} $1" | tee -a "${LOG_FILE}"; }

log_info "==================== OllamaMax Final Validation Orchestrator ===================="
log_info "Timestamp: ${TIMESTAMP}"
log_info "Log File: ${LOG_FILE}"
log_info "Phases: ${PHASES_TO_RUN}"
log_info "Skip: ${SKIP_PHASES}"
log_info "Parallel Mode: ${PARALLEL_MODE}"
log_info "Dry Run: ${DRY_RUN}"
log_info "================================================================================="

# Phase tracking
declare -A PHASE_STATUS
declare -A PHASE_DURATION
PHASES=(
    "e2e:End-to-End Integration Tests:run-e2e-integration-tests.sh:30-45 min"
    "load:Distributed Load Tests:run-load-test-distributed.sh:2-3 hours"
    "chaos:Chaos Engineering Tests:execute-chaos-engineering.sh:3-4 hours"
    "security:Security Penetration Tests:execute-penetration-tests.sh:1-2 hours"
    "dr:Disaster Recovery Validation:validate-disaster-recovery.sh:2-3 hours"
    "deployment:Deployment Validation:run-deployment-validation.sh:1-2 hours"
)

# Check if phase should be run
should_run_phase() {
    local phase_id=$1

    # Check if skipped
    if [[ ",${SKIP_PHASES}," == *",${phase_id},"* ]]; then
        return 1
    fi

    # Check if in phases to run
    if [ "${PHASES_TO_RUN}" = "all" ]; then
        return 0
    elif [[ ",${PHASES_TO_RUN}," == *",${phase_id},"* ]]; then
        return 0
    else
        return 1
    fi
}

# Pre-validation checks
log_phase "Pre-Validation Checks"

# Check required tools
log_info "Checking required tools..."

REQUIRED_TOOLS=("curl" "jq" "docker" "git")
MISSING_TOOLS=()

for tool in "${REQUIRED_TOOLS[@]}"; do
    if ! command -v ${tool} &> /dev/null; then
        MISSING_TOOLS+=("${tool}")
        log_warning "${tool} not found"
    fi
done

if [ ${#MISSING_TOOLS[@]} -gt 0 ]; then
    log_warning "Missing tools: ${MISSING_TOOLS[*]}"
    log_warning "Some tests may be skipped"
fi

# Check system resources
log_info "Checking system resources..."

TOTAL_MEM=$(free -g | awk '/^Mem:/{print $2}')
AVAILABLE_MEM=$(free -g | awk '/^Mem:/{print $7}')
CPU_CORES=$(nproc)
DISK_SPACE=$(df -BG . | awk 'NR==2 {print $4}' | tr -d 'G')

log_info "  CPU Cores: ${CPU_CORES}"
log_info "  Total Memory: ${TOTAL_MEM}GB"
log_info "  Available Memory: ${AVAILABLE_MEM}GB"
log_info "  Available Disk: ${DISK_SPACE}GB"

if [ "${AVAILABLE_MEM}" -lt 16 ]; then
    log_warning "Low memory available (${AVAILABLE_MEM}GB). Recommended: 16GB+"
fi

if [ "${CPU_CORES}" -lt 8 ]; then
    log_warning "Low CPU cores (${CPU_CORES}). Recommended: 16+ cores"
fi

if [ "${DISK_SPACE}" -lt 50 ]; then
    log_warning "Low disk space (${DISK_SPACE}GB). Recommended: 100GB+"
fi

# Check target system
log_info "Checking target system availability..."

if curl -sf http://localhost:11434/health > /dev/null 2>&1; then
    log_success "Target system is accessible"
else
    log_warning "Target system not accessible at http://localhost:11434"
    log_warning "Some tests may fail or be skipped"
fi

log_success "Pre-validation checks completed"

# Dry run exit
if [ "${DRY_RUN}" = true ]; then
    log_info "Dry run mode - showing what would be executed:"
    for phase_spec in "${PHASES[@]}"; do
        IFS=':' read -r phase_id phase_name script duration <<< "${phase_spec}"

        if should_run_phase "${phase_id}"; then
            echo "  [RUN] Phase: ${phase_name} (${duration})"
            echo "        Script: ${script}"
        else
            echo "  [SKIP] Phase: ${phase_name}"
        fi
    done
    exit 0
fi

# Execute validation phases
log_info "Starting validation phase execution..."

TOTAL_PHASES=0
COMPLETED_PHASES=0
FAILED_PHASES=0
VALIDATION_START=$(date +%s)

for phase_spec in "${PHASES[@]}"; do
    IFS=':' read -r phase_id phase_name script duration <<< "${phase_spec}"

    if ! should_run_phase "${phase_id}"; then
        log_info "Skipping phase: ${phase_name}"
        PHASE_STATUS["${phase_id}"]="skipped"
        continue
    fi

    ((TOTAL_PHASES++))

    log_phase "Phase ${TOTAL_PHASES}: ${phase_name}"
    log_info "Estimated duration: ${duration}"
    log_info "Script: ${script}"

    PHASE_START=$(date +%s)

    if bash "${SCRIPT_DIR}/${script}" 2>&1 | tee -a "${LOG_FILE}"; then
        PHASE_END=$(date +%s)
        PHASE_TIME=$((PHASE_END - PHASE_START))
        PHASE_DURATION["${phase_id}"]=${PHASE_TIME}
        PHASE_STATUS["${phase_id}"]="passed"

        log_success "Phase ${phase_name} completed in ${PHASE_TIME}s"
        ((COMPLETED_PHASES++))
    else
        PHASE_END=$(date +%s)
        PHASE_TIME=$((PHASE_END - PHASE_START))
        PHASE_DURATION["${phase_id}"]=${PHASE_TIME}
        PHASE_STATUS["${phase_id}"]="failed"

        log_error "Phase ${phase_name} failed after ${PHASE_TIME}s"
        ((FAILED_PHASES++))

        # Continue with remaining phases to collect all results
        log_warning "Continuing with remaining phases to collect complete results..."
    fi

    echo "" | tee -a "${LOG_FILE}"
done

VALIDATION_END=$(date +%s)
TOTAL_DURATION=$((VALIDATION_END - VALIDATION_START))

# Generate final production readiness report
log_phase "Generating Final Production Readiness Report"

if bash "${SCRIPT_DIR}/generate-final-production-report.sh" 2>&1 | tee -a "${LOG_FILE}"; then
    log_success "Final production readiness report generated"
else
    log_error "Failed to generate final report"
fi

# Display validation summary
log_info "================================================================================="
log_info "                        FINAL VALIDATION SUMMARY"
log_info "================================================================================="
log_info "Total Duration: ${TOTAL_DURATION}s ($(echo "scale=2; ${TOTAL_DURATION} / 3600" | bc) hours)"
log_info "Total Phases: ${TOTAL_PHASES}"
log_info "Completed: ${COMPLETED_PHASES}"
log_info "Failed: ${FAILED_PHASES}"
log_info ""

for phase_spec in "${PHASES[@]}"; do
    IFS=':' read -r phase_id phase_name script duration <<< "${phase_spec}"

    status="${PHASE_STATUS[${phase_id}]:-not run}"
    phase_time="${PHASE_DURATION[${phase_id}]:-0}"

    case "${status}" in
        passed)
            log_success "✅ ${phase_name}: PASSED (${phase_time}s)"
            ;;
        failed)
            log_error "❌ ${phase_name}: FAILED (${phase_time}s)"
            ;;
        skipped)
            log_info "⊝ ${phase_name}: SKIPPED"
            ;;
        *)
            log_warning "⚠ ${phase_name}: NOT RUN"
            ;;
    esac
done

log_info ""
log_info "================================================================================="
log_info "Results Location:"
log_info "  Final Report: final-validation-results/FINAL_PRODUCTION_READINESS_REPORT.md"
log_info "  Known Issues: KNOWN_ISSUES.md"
log_info "  Log File: ${LOG_FILE}"
log_info "================================================================================="

# Exit with appropriate code
if [ ${FAILED_PHASES} -eq 0 ]; then
    log_success "All validation phases completed successfully!"
    exit 0
elif [ ${FAILED_PHASES} -le 1 ] && [ ${COMPLETED_PHASES} -ge 4 ]; then
    log_warning "Validation completed with ${FAILED_PHASES} failure(s)"
    exit 0
else
    log_error "Validation failed with ${FAILED_PHASES} critical failures"
    exit 1
fi

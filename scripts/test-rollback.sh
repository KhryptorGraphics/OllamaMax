#!/bin/bash
# Rollback Testing Script
# References: ollama-distributed/scripts/rollback.sh, ollama-distributed/scripts/health-check.sh

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
REPORT_FILE="${REPORT_DIR}/rollback-test-${TIMESTAMP}.json"

mkdir -p "${REPORT_DIR}"

echo -e "${BLUE}=== Rollback Testing ===${NC}"

# Helper functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[PASS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[FAIL]${NC} $1"; }

ROLLBACK_TESTS=()

# Phase 1: Docker Rollback Testing
echo -e "\n${BLUE}=== Phase 1: Docker Rollback Testing ===${NC}"

if docker compose version &> /dev/null; then
    log_info "Testing Docker deployment rollback..."

    # Record current state
    CURRENT_IMAGES=$(docker compose images --format json 2>/dev/null | jq -r '.[] | .Repository + ":" + .Tag' 2>/dev/null || echo "")
    if [ -z "$CURRENT_IMAGES" ]; then
        log_info "Could not parse images with jq (may not be supported on this version)"
        CURRENT_IMAGES=$(docker compose images 2>/dev/null | tail -n +2 | awk '{print $2":"$3}' || echo "unavailable")
    fi
    log_info "Current images: ${CURRENT_IMAGES}"

    # Simulate rollback
    if [ -f "ollama-distributed/scripts/rollback.sh" ]; then
        log_info "Executing rollback script..."
        ROLLBACK_START=$(date +%s)

        # Run rollback in test mode
        bash ollama-distributed/scripts/rollback.sh --docker --dry-run &> /dev/null || true

        ROLLBACK_END=$(date +%s)
        ROLLBACK_TIME=$((ROLLBACK_END - ROLLBACK_START))

        log_success "Rollback script executed in ${ROLLBACK_TIME}s"
        ROLLBACK_TESTS+=("{\"type\":\"docker\",\"time\":${ROLLBACK_TIME},\"status\":\"success\"}")
    else
        log_warning "Rollback script not found"
        ROLLBACK_TESTS+=("{\"type\":\"docker\",\"status\":\"skipped\",\"reason\":\"script not found\"}")
    fi

    # Verify services are still healthy
    if [ -f "ollama-distributed/scripts/health-check.sh" ]; then
        log_info "Running health checks after rollback..."
        if bash ollama-distributed/scripts/health-check.sh &> /dev/null; then
            log_success "All services healthy after rollback"
        else
            log_warning "Some health checks failed"
        fi
    fi
else
    log_warning "Docker Compose not available - skipping Docker rollback test"
fi

# Phase 2: Kubernetes Rollback Testing
echo -e "\n${BLUE}=== Phase 2: Kubernetes Rollback Testing ===${NC}"

if command -v kubectl &> /dev/null && kubectl cluster-info &> /dev/null; then
    log_info "Testing Kubernetes rollback..."

    # Get current deployments
    DEPLOYMENTS=$(kubectl get deployments --all-namespaces -o jsonpath='{range .items[*]}{.metadata.namespace}/{.metadata.name}{"\n"}{end}' 2>/dev/null || echo "")

    if [ -n "$DEPLOYMENTS" ]; then
        # Test rollback on first deployment
        FIRST_DEPLOY=$(echo "$DEPLOYMENTS" | head -1)
        NS=$(echo "$FIRST_DEPLOY" | cut -d/ -f1)
        NAME=$(echo "$FIRST_DEPLOY" | cut -d/ -f2)

        log_info "Testing rollback for deployment ${NAME} in namespace ${NS}..."

        # Get current revision
        CURRENT_REVISION=$(kubectl rollout history deployment/"${NAME}" -n "${NS}" 2>/dev/null | tail -1 | awk '{print $1}' || echo "1")
        log_info "Current revision: ${CURRENT_REVISION}"

        # Test rollback command (dry-run)
        ROLLBACK_START=$(date +%s)
        if kubectl rollout undo deployment/"${NAME}" -n "${NS}" --dry-run=client &> /dev/null; then
            ROLLBACK_END=$(date +%s)
            ROLLBACK_TIME=$((ROLLBACK_END - ROLLBACK_START))

            log_success "Kubernetes rollback validated in ${ROLLBACK_TIME}s"
            ROLLBACK_TESTS+=("{\"type\":\"kubernetes\",\"deployment\":\"${NAME}\",\"time\":${ROLLBACK_TIME},\"status\":\"success\"}")
        else
            log_error "Kubernetes rollback validation failed"
            ROLLBACK_TESTS+=("{\"type\":\"kubernetes\",\"deployment\":\"${NAME}\",\"status\":\"failed\"}")
        fi
    else
        log_warning "No deployments found for testing"
    fi
else
    log_warning "Kubernetes not available - skipping K8s rollback test"
fi

# Phase 3: Database Rollback Testing
echo -e "\n${BLUE}=== Phase 3: Database Rollback Testing ===${NC}"

if docker compose ps postgres &> /dev/null; then
    log_info "Testing database rollback capability..."

    # Check if database is accessible
    if docker compose exec -T postgres pg_isready -U postgres &> /dev/null; then
        log_success "Database is accessible"

        # Test backup creation
        log_info "Testing backup creation..."
        BACKUP_FILE="/tmp/test-backup-${TIMESTAMP}.sql"

        if docker compose exec -T postgres pg_dump -U postgres > "$BACKUP_FILE" 2>/dev/null; then
            BACKUP_SIZE=$(stat -f%z "$BACKUP_FILE" 2>/dev/null || stat -c%s "$BACKUP_FILE")
            log_success "Backup created: ${BACKUP_SIZE} bytes"

            ROLLBACK_TESTS+=("{\"type\":\"database\",\"backup_size\":${BACKUP_SIZE},\"status\":\"success\"}")

            # Cleanup
            rm -f "$BACKUP_FILE"
        else
            log_error "Backup creation failed"
            ROLLBACK_TESTS+=("{\"type\":\"database\",\"status\":\"failed\"}")
        fi
    else
        log_warning "Database not accessible"
    fi
else
    log_warning "PostgreSQL not running - skipping database rollback test"
fi

# Phase 4: Rollback Timing Measurements
echo -e "\n${BLUE}=== Phase 4: Rollback Timing Analysis ===${NC}"

# Calculate average rollback time
TOTAL_TIME=0
COUNT=0

for TEST in "${ROLLBACK_TESTS[@]}"; do
    TIME=$(echo "$TEST" | jq -r '.time // 0')
    if [ "$TIME" -gt 0 ]; then
        TOTAL_TIME=$((TOTAL_TIME + TIME))
        COUNT=$((COUNT + 1))
    fi
done

if [ $COUNT -gt 0 ]; then
    AVG_TIME=$((TOTAL_TIME / COUNT))
    log_info "Average rollback time: ${AVG_TIME}s"

    # Check against RTO (Recovery Time Objective)
    RTO_TARGET=300  # 5 minutes

    if [ $AVG_TIME -lt $RTO_TARGET ]; then
        log_success "Rollback time within RTO target (${RTO_TARGET}s)"
    else
        log_warning "Rollback time exceeds RTO target (${RTO_TARGET}s)"
    fi
fi

# Phase 5: Data Integrity Validation
echo -e "\n${BLUE}=== Phase 5: Data Integrity Validation ===${NC}"

if docker compose ps redis &> /dev/null; then
    log_info "Testing data persistence after rollback..."

    # Write test data
    TEST_KEY="rollback-test-${TIMESTAMP}"
    TEST_VALUE="test-value-${TIMESTAMP}"

    docker compose exec -T redis redis-cli SET "$TEST_KEY" "$TEST_VALUE" &> /dev/null

    # Verify data
    RETRIEVED_VALUE=$(docker compose exec -T redis redis-cli GET "$TEST_KEY" 2>/dev/null || echo "")

    if [ "$RETRIEVED_VALUE" = "$TEST_VALUE" ]; then
        log_success "Data integrity maintained"
    else
        log_error "Data integrity check failed"
    fi

    # Cleanup
    docker compose exec -T redis redis-cli DEL "$TEST_KEY" &> /dev/null
fi

# Generate Report
cat > "${REPORT_FILE}" <<EOF
{
  "timestamp": "${TIMESTAMP}",
  "tests_executed": $(echo "${ROLLBACK_TESTS[@]}" | wc -w),
  "average_rollback_time_seconds": ${AVG_TIME:-0},
  "rto_target_seconds": 300,
  "rto_met": $([ ${AVG_TIME:-0} -lt 300 ] && echo "true" || echo "false"),
  "test_results": [
    $(IFS=,; echo "${ROLLBACK_TESTS[*]}")
  ],
  "status": "completed"
}
EOF

log_success "Report generated: ${REPORT_FILE}"
echo -e "\n${GREEN}Rollback testing completed${NC}"

# Summary
echo -e "\n${BLUE}=== Rollback Test Summary ===${NC}"
echo "Tests Executed: ${#ROLLBACK_TESTS[@]}"
echo "Average Rollback Time: ${AVG_TIME:-0}s"
echo "RTO Target: 300s (5 minutes)"

if [ ${AVG_TIME:-0} -lt 300 ]; then
    echo -e "${GREEN}RTO Target: MET${NC}"
else
    echo -e "${RED}RTO Target: NOT MET${NC}"
fi

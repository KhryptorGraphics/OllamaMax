#!/bin/bash
# Docker Deployment Pre-Check Script
# This script performs lightweight validation before actual deployment
# References: validate-docker-deployment.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
COMPOSE_FILE="${1:-docker-compose.yml}"
REPORT_DIR="deployment-results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
REPORT_FILE="${REPORT_DIR}/docker-precheck-${TIMESTAMP}.json"

mkdir -p "${REPORT_DIR}"

VALIDATION_RESULTS=()
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNING_CHECKS=0

echo -e "${BLUE}=== Docker Deployment Pre-Check ===${NC}"
echo "Compose File: ${COMPOSE_FILE}"
echo "Timestamp: ${TIMESTAMP}"
echo ""

# Helper functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[PASS]${NC} $1"; PASSED_CHECKS=$((PASSED_CHECKS + 1)); }
log_warning() { echo -e "${YELLOW}[WARN]${NC} $1"; WARNING_CHECKS=$((WARNING_CHECKS + 1)); }
log_error() { echo -e "${RED}[FAIL]${NC} $1"; FAILED_CHECKS=$((FAILED_CHECKS + 1)); }

add_result() {
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    VALIDATION_RESULTS+=("{\"name\":\"$1\",\"status\":\"$2\",\"message\":\"$3\"}")
}

# Check Docker version
log_info "Checking Docker version..."
if docker --version &> /dev/null; then
    DOCKER_VERSION=$(docker --version | awk '{print $3}' | sed 's/,//')
    log_success "Docker version: ${DOCKER_VERSION}"
    add_result "Docker Version" "pass" "Docker ${DOCKER_VERSION} installed"
else
    log_error "Docker is not installed"
    add_result "Docker Version" "fail" "Docker not installed"
    exit 1
fi

# Check Docker Compose version
log_info "Checking Docker Compose version..."
if docker compose version &> /dev/null; then
    COMPOSE_VERSION=$(docker compose version | awk '{print $4}')
    log_success "Docker Compose version: ${COMPOSE_VERSION}"
    add_result "Docker Compose Version" "pass" "Docker Compose ${COMPOSE_VERSION} installed"
else
    log_error "Docker Compose is not installed"
    add_result "Docker Compose Version" "fail" "Docker Compose not installed"
    exit 1
fi

# Check system resources
log_info "Checking system resources..."
TOTAL_MEM=$(free -g | awk '/^Mem:/{print $2}')
AVAILABLE_MEM=$(free -g | awk '/^Mem:/{print $7}')
DISK_USAGE=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')

if [ "$TOTAL_MEM" -lt 8 ]; then
    log_warning "Low memory: ${TOTAL_MEM}GB (recommended: 8GB+)"
    add_result "System Memory" "warning" "Low memory: ${TOTAL_MEM}GB"
else
    log_success "Sufficient memory: ${TOTAL_MEM}GB"
    add_result "System Memory" "pass" "Sufficient memory: ${TOTAL_MEM}GB"
fi

if [ "$DISK_USAGE" -gt 80 ]; then
    log_warning "High disk usage: ${DISK_USAGE}%"
    add_result "Disk Space" "warning" "High disk usage: ${DISK_USAGE}%"
else
    log_success "Disk usage: ${DISK_USAGE}%"
    add_result "Disk Space" "pass" "Disk usage: ${DISK_USAGE}%"
fi

# Validate compose file syntax
log_info "Validating compose file syntax..."
if docker compose -f "${COMPOSE_FILE}" config > /dev/null 2>&1; then
    log_success "Compose file syntax valid"
    add_result "Compose File Syntax" "pass" "Valid YAML syntax"
else
    log_error "Compose file syntax invalid"
    add_result "Compose File Syntax" "fail" "Invalid YAML syntax"
    exit 1
fi

# Check for port conflicts
log_info "Checking for port conflicts..."
PORTS=$(docker compose -f "${COMPOSE_FILE}" config | grep -E "^\s+- \"[0-9]+:" | awk -F'"' '{print $2}' | cut -d: -f1 | sort -u)
PORT_CONFLICTS=0

# Check if ss is available, fallback to netstat
if command -v ss &> /dev/null; then
    for PORT in $PORTS; do
        if ss -tuln | grep -q ":${PORT} "; then
            log_warning "Port ${PORT} is already in use"
            PORT_CONFLICTS=$((PORT_CONFLICTS + 1))
        fi
    done
elif command -v netstat &> /dev/null; then
    for PORT in $PORTS; do
        if netstat -tuln 2>/dev/null | grep -q ":${PORT} "; then
            log_warning "Port ${PORT} is already in use"
            PORT_CONFLICTS=$((PORT_CONFLICTS + 1))
        fi
    done
else
    log_warning "Neither ss nor netstat available - skipping port conflict check"
    add_result "Port Conflicts" "warning" "Cannot check ports (no ss/netstat)"
    PORT_CONFLICTS=-1
fi

if [ "$PORT_CONFLICTS" -eq 0 ]; then
    log_success "No port conflicts detected"
    add_result "Port Conflicts" "pass" "No port conflicts"
elif [ "$PORT_CONFLICTS" -gt 0 ]; then
    log_warning "${PORT_CONFLICTS} port(s) in use"
    add_result "Port Conflicts" "warning" "${PORT_CONFLICTS} port conflicts detected"
fi

# Generate Report
VALIDATION_SCORE=$((PASSED_CHECKS * 100 / TOTAL_CHECKS))

cat > "${REPORT_FILE}" <<EOF
{
  "timestamp": "${TIMESTAMP}",
  "compose_file": "${COMPOSE_FILE}",
  "precheck": true,
  "summary": {
    "total_checks": ${TOTAL_CHECKS},
    "passed": ${PASSED_CHECKS},
    "failed": ${FAILED_CHECKS},
    "warnings": ${WARNING_CHECKS},
    "validation_score": ${VALIDATION_SCORE}
  },
  "results": [
    $(IFS=,; echo "${VALIDATION_RESULTS[*]}")
  ]
}
EOF

log_success "Pre-check report generated: ${REPORT_FILE}"

# Print summary
echo -e "\n${BLUE}=== Pre-Check Summary ===${NC}"
echo "Total Checks: ${TOTAL_CHECKS}"
echo -e "Passed: ${GREEN}${PASSED_CHECKS}${NC}"
echo -e "Failed: ${RED}${FAILED_CHECKS}${NC}"
echo -e "Warnings: ${YELLOW}${WARNING_CHECKS}${NC}"
echo "Validation Score: ${VALIDATION_SCORE}/100"

# Exit with appropriate code
if [ "$FAILED_CHECKS" -gt 0 ]; then
    echo -e "\n${RED}Pre-check FAILED${NC}"
    exit 1
else
    echo -e "\n${GREEN}Pre-check PASSED${NC}"
    exit 0
fi

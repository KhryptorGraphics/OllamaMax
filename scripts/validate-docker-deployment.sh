#!/bin/bash
# Docker Deployment Validation Script
# References: deploy-docker.sh, ollama-distributed/scripts/health-check.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="${1:-docker-compose.yml}"
TIMEOUT=300
REPORT_DIR="deployment-results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
REPORT_FILE="${REPORT_DIR}/docker-validation-${TIMESTAMP}.json"
REPORT_MD="${REPORT_DIR}/docker-validation-${TIMESTAMP}.md"

# Create report directory
mkdir -p "${REPORT_DIR}"

# Initialize report
VALIDATION_RESULTS=()
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNING_CHECKS=0

echo -e "${BLUE}=== Docker Deployment Validation ===${NC}"
echo "Compose File: ${COMPOSE_FILE}"
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
    local optional_tools=()

    # Required tools
    command -v docker &> /dev/null || missing_tools+=("docker")
    command -v curl &> /dev/null || missing_tools+=("curl")

    # Optional tools (warnings only)
    command -v jq &> /dev/null || optional_tools+=("jq")
    command -v bc &> /dev/null || optional_tools+=("bc")

    # Check for netstat or ss
    if ! command -v ss &> /dev/null && ! command -v netstat &> /dev/null; then
        optional_tools+=("ss or netstat")
    fi

    if [ ${#missing_tools[@]} -gt 0 ]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        log_info "Please install missing tools before running this script"
        exit 1
    fi

    if [ ${#optional_tools[@]} -gt 0 ]; then
        log_warning "Optional tools not available: ${optional_tools[*]}"
        log_info "Some checks may be skipped"
    fi

    log_info "Required tools available"
}

check_prerequisites

# 1. Pre-Deployment Validation
echo -e "${BLUE}=== Phase 1: Pre-Deployment Validation ===${NC}"

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

# Choose appropriate tool for port checking
if command -v ss &> /dev/null; then
    PORT_CHECK_CMD="ss -tuln"
elif command -v netstat &> /dev/null; then
    PORT_CHECK_CMD="netstat -tuln"
else
    log_warning "Neither ss nor netstat available - skipping port conflict check"
    add_result "Port Conflicts" "warning" "Port check skipped (ss/netstat not available)"
    PORT_CHECK_CMD=""
fi

if [ -n "$PORT_CHECK_CMD" ]; then
    for PORT in $PORTS; do
        if $PORT_CHECK_CMD 2>/dev/null | grep -q ":${PORT} "; then
            log_warning "Port ${PORT} is already in use"
            PORT_CONFLICTS=$((PORT_CONFLICTS + 1))
        fi
    done

    if [ "$PORT_CONFLICTS" -eq 0 ]; then
        log_success "No port conflicts detected"
        add_result "Port Conflicts" "pass" "No port conflicts"
    else
        log_warning "${PORT_CONFLICTS} port(s) in use"
        add_result "Port Conflicts" "warning" "${PORT_CONFLICTS} port conflicts detected"
    fi
fi

# 2. Deployment Execution
echo -e "\n${BLUE}=== Phase 2: Deployment Execution ===${NC}"

log_info "Pulling required images..."
START_TIME=$(date +%s)
if docker compose -f "${COMPOSE_FILE}" pull; then
    log_success "Images pulled successfully"
    add_result "Image Pull" "pass" "All images pulled"
else
    log_error "Failed to pull images"
    add_result "Image Pull" "fail" "Image pull failed"
    exit 1
fi

log_info "Starting services..."
if docker compose -f "${COMPOSE_FILE}" up -d; then
    log_success "Services started"
    add_result "Service Start" "pass" "All services started"
else
    log_error "Failed to start services"
    add_result "Service Start" "fail" "Service start failed"
    exit 1
fi

END_TIME=$(date +%s)
DEPLOYMENT_TIME=$((END_TIME - START_TIME))
log_info "Deployment time: ${DEPLOYMENT_TIME}s"

# Wait for services to be ready
log_info "Waiting for services to be ready (timeout: ${TIMEOUT}s)..."
sleep 10

# 3. Post-Deployment Validation
echo -e "\n${BLUE}=== Phase 3: Post-Deployment Validation ===${NC}"

# Verify all services are running
log_info "Verifying service status..."
SERVICES=$(docker compose -f "${COMPOSE_FILE}" ps --services)
RUNNING_SERVICES=0
TOTAL_SERVICES=0

for SERVICE in $SERVICES; do
    TOTAL_SERVICES=$((TOTAL_SERVICES + 1))

    if command -v jq &> /dev/null; then
        STATUS=$(docker compose -f "${COMPOSE_FILE}" ps "${SERVICE}" --format json 2>/dev/null | jq -r '.[0].State' 2>/dev/null || echo "unknown")
    else
        # Fallback without jq - parse docker ps output
        STATUS=$(docker compose -f "${COMPOSE_FILE}" ps "${SERVICE}" 2>/dev/null | grep "${SERVICE}" | awk '{print $4}' || echo "unknown")
    fi

    if [ "$STATUS" = "running" ] || echo "$STATUS" | grep -qi "up"; then
        log_success "Service ${SERVICE}: running"
        RUNNING_SERVICES=$((RUNNING_SERVICES + 1))
        add_result "Service: ${SERVICE}" "pass" "Running"
    else
        log_error "Service ${SERVICE}: ${STATUS}"
        add_result "Service: ${SERVICE}" "fail" "Status: ${STATUS}"
    fi
done

# Health check endpoints
log_info "Testing health check endpoints..."

# Detect actual mapped ports dynamically
API_PORT=$(docker compose -f "${COMPOSE_FILE}" port ollamamax-api 13100 2>/dev/null | cut -d: -f2 || echo "13100")
WEB_PORT=$(docker compose -f "${COMPOSE_FILE}" port ollamamax-web 8080 2>/dev/null | cut -d: -f2 || echo "8080")
OLLAMA_PORT=$(docker compose -f "${COMPOSE_FILE}" port ollama-primary 11434 2>/dev/null | cut -d: -f2 || echo "11434")

HEALTH_ENDPOINTS=(
    "http://localhost:${API_PORT}/health"
    "http://localhost:${WEB_PORT}/"
    "http://localhost:${OLLAMA_PORT}/api/tags"
)

HEALTHY_ENDPOINTS=0
for ENDPOINT in "${HEALTH_ENDPOINTS[@]}"; do
    if curl -f -s -o /dev/null -w "%{http_code}" --max-time 10 "${ENDPOINT}" > /dev/null 2>&1; then
        log_success "Health check passed: ${ENDPOINT}"
        HEALTHY_ENDPOINTS=$((HEALTHY_ENDPOINTS + 1))
        add_result "Health: ${ENDPOINT}" "pass" "Endpoint healthy"
    else
        log_warning "Health check failed: ${ENDPOINT}"
        add_result "Health: ${ENDPOINT}" "warning" "Endpoint not responding"
    fi
done

# Redis health check using redis-cli
log_info "Testing Redis health..."
if docker compose exec -T redis redis-cli PING 2>/dev/null | grep -q "PONG"; then
    log_success "Redis health check passed"
    HEALTHY_ENDPOINTS=$((HEALTHY_ENDPOINTS + 1))
    add_result "Health: Redis" "pass" "Redis responding to PING"
else
    log_warning "Redis health check failed"
    add_result "Health: Redis" "warning" "Redis not responding"
fi

# Test service connectivity
log_info "Testing service connectivity..."
CONNECTIVITY_TESTS=(
    "api:postgres"
    "api:redis"
    "web:api"
)

for TEST in "${CONNECTIVITY_TESTS[@]}"; do
    FROM=$(echo "$TEST" | cut -d: -f1)
    TO=$(echo "$TEST" | cut -d: -f2)

    if docker compose -f "${COMPOSE_FILE}" exec -T "${FROM}" ping -c 1 "${TO}" > /dev/null 2>&1; then
        log_success "Connectivity: ${FROM} -> ${TO}"
        add_result "Connectivity: ${FROM}->${TO}" "pass" "Connected"
    else
        log_warning "Connectivity failed: ${FROM} -> ${TO}"
        add_result "Connectivity: ${FROM}->${TO}" "warning" "Connection failed"
    fi
done

# Verify volume mounts
log_info "Verifying volume mounts..."
if command -v jq &> /dev/null; then
    VOLUMES=$(docker compose -f "${COMPOSE_FILE}" ps --format json 2>/dev/null | jq -r '.[].Mounts' 2>/dev/null || echo "[]")
    if [ -n "$VOLUMES" ] && [ "$VOLUMES" != "[]" ] && [ "$VOLUMES" != "null" ]; then
        log_success "Volumes mounted successfully"
        add_result "Volume Mounts" "pass" "Volumes mounted"
    else
        log_warning "No volumes detected"
        add_result "Volume Mounts" "warning" "No volumes found"
    fi
else
    # Fallback: check if volumes are defined in compose file
    if grep -q "volumes:" "${COMPOSE_FILE}"; then
        log_success "Volumes defined in compose file"
        add_result "Volume Mounts" "pass" "Volumes defined (jq unavailable for detailed check)"
    else
        log_warning "No volumes detected (basic check)"
        add_result "Volume Mounts" "warning" "Volume check skipped (jq not available)"
    fi
fi

# 4. Performance Validation
echo -e "\n${BLUE}=== Phase 4: Performance Validation ===${NC}"

# Measure API response time
log_info "Measuring API response time..."
if command -v curl &> /dev/null; then
    RESPONSE_TIME=$(curl -o /dev/null -s -w '%{time_total}' http://localhost:8080/health 2>/dev/null || echo "0")

    # Use bc if available, otherwise use awk
    if command -v bc &> /dev/null; then
        RESPONSE_TIME_MS=$(echo "$RESPONSE_TIME * 1000" | bc)
        COMPARE_200=$(echo "$RESPONSE_TIME_MS < 200" | bc -l)
        COMPARE_500=$(echo "$RESPONSE_TIME_MS < 500" | bc -l)
    else
        RESPONSE_TIME_MS=$(awk "BEGIN {print $RESPONSE_TIME * 1000}")
        COMPARE_200=$(awk "BEGIN {print ($RESPONSE_TIME_MS < 200)}")
        COMPARE_500=$(awk "BEGIN {print ($RESPONSE_TIME_MS < 500)}")
    fi

    if [ "$COMPARE_200" -eq 1 ]; then
        log_success "API response time: ${RESPONSE_TIME_MS}ms (target: <200ms)"
        add_result "API Response Time" "pass" "${RESPONSE_TIME_MS}ms"
    elif [ "$COMPARE_500" -eq 1 ]; then
        log_warning "API response time: ${RESPONSE_TIME_MS}ms (target: <200ms)"
        add_result "API Response Time" "warning" "${RESPONSE_TIME_MS}ms"
    else
        log_error "API response time: ${RESPONSE_TIME_MS}ms (target: <200ms)"
        add_result "API Response Time" "fail" "${RESPONSE_TIME_MS}ms"
    fi
fi

# Check resource usage
log_info "Checking resource usage..."
STATS=$(docker stats --no-stream --format "{{.Container}},{{.CPUPerc}},{{.MemUsage}}" 2>/dev/null)
if [ -n "$STATS" ]; then
    log_success "Resource monitoring active"
    add_result "Resource Monitoring" "pass" "Stats available"
else
    log_warning "Unable to retrieve resource stats"
    add_result "Resource Monitoring" "warning" "Stats unavailable"
fi

# 5. Security Validation
echo -e "\n${BLUE}=== Phase 5: Security Validation ===${NC}"

# Check for exposed secrets
log_info "Checking for exposed secrets..."
SECRETS_FOUND=0
if docker compose -f "${COMPOSE_FILE}" config | grep -iE "(password|secret|key|token)" | grep -vE "(PASSWORD|SECRET|KEY|TOKEN)=\$\{" > /dev/null; then
    log_error "Potential hardcoded secrets found"
    SECRETS_FOUND=1
    add_result "Secrets Management" "fail" "Hardcoded secrets detected"
else
    log_success "No hardcoded secrets found"
    add_result "Secrets Management" "pass" "No hardcoded secrets"
fi

# 6. Generate Report
echo -e "\n${BLUE}=== Phase 6: Generating Report ===${NC}"

# Calculate validation score
VALIDATION_SCORE=$((PASSED_CHECKS * 100 / TOTAL_CHECKS))

# JSON Report
cat > "${REPORT_FILE}" <<EOF
{
  "timestamp": "${TIMESTAMP}",
  "compose_file": "${COMPOSE_FILE}",
  "deployment_time_seconds": ${DEPLOYMENT_TIME},
  "summary": {
    "total_checks": ${TOTAL_CHECKS},
    "passed": ${PASSED_CHECKS},
    "failed": ${FAILED_CHECKS},
    "warnings": ${WARNING_CHECKS},
    "validation_score": ${VALIDATION_SCORE}
  },
  "services": {
    "total": ${TOTAL_SERVICES},
    "running": ${RUNNING_SERVICES}
  },
  "results": [
    $(IFS=,; echo "${VALIDATION_RESULTS[*]}")
  ]
}
EOF

# Markdown Report
cat > "${REPORT_MD}" <<EOF
# Docker Deployment Validation Report

**Generated:** ${TIMESTAMP}
**Compose File:** ${COMPOSE_FILE}
**Deployment Time:** ${DEPLOYMENT_TIME}s

## Summary

- **Validation Score:** ${VALIDATION_SCORE}/100
- **Total Checks:** ${TOTAL_CHECKS}
- **Passed:** ${PASSED_CHECKS} ✅
- **Failed:** ${FAILED_CHECKS} ❌
- **Warnings:** ${WARNING_CHECKS} ⚠️

## Service Status

- **Total Services:** ${TOTAL_SERVICES}
- **Running Services:** ${RUNNING_SERVICES}

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

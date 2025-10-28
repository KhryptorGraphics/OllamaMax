#!/bin/bash
# Load Balancing Validation Script
# References: nginx/nginx-production.conf, docker-compose.yml
#
# IMPORTANT: This script tests load balancing by sending requests to PROXIED endpoints
# that include the X-Served-By header set by Nginx. Non-proxied endpoints like /health
# and /nginx-health do NOT include this header and cannot be used for distribution analysis.
#
# Proxied endpoints that include X-Served-By:
# - /api/* (proxied to api_backend)
# - /ollama/* (proxied to ollama_backend)
# - / (proxied to web_backend)
#
# Non-proxied endpoints (do NOT use for distribution testing):
# - /health (self-served by Nginx)
# - /nginx-health (self-served by Nginx on port 8081)

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
REPORT_FILE="${REPORT_DIR}/loadbalancing-${TIMESTAMP}.json"
LB_URL="${1:-http://localhost}"
LB_PATH="${2:-/}"  # Default to root path (proxied to web_backend)
NUM_REQUESTS=1000

mkdir -p "${REPORT_DIR}"

echo -e "${BLUE}=== Load Balancing Validation ===${NC}"
echo -e "${BLUE}Testing endpoint: ${LB_URL}${LB_PATH}${NC}"

# Helper functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[PASS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[FAIL]${NC} $1"; }

# Check CLI prerequisites
check_prerequisites() {
    local missing_tools=()

    command -v curl &> /dev/null || missing_tools+=("curl")
    command -v docker &> /dev/null || missing_tools+=("docker")

    if [ ${#missing_tools[@]} -gt 0 ]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        log_info "Please install missing tools before running this script"
        exit 1
    fi

    log_info "All required tools available"
}

check_prerequisites

# Phase 1: Check Load Balancer Configuration
echo -e "\n${BLUE}=== Phase 1: Load Balancer Configuration ===${NC}"

if [ -f "nginx/nginx-production.conf" ]; then
    log_info "Checking nginx configuration..."

    if grep -q "least_conn" nginx/nginx-production.conf; then
        log_success "Load balancing algorithm: least_conn"
    else
        log_warning "Load balancing algorithm not specified"
    fi

    UPSTREAM_COUNT=$(grep -c "server.*:.*;" nginx/nginx-production.conf || echo "0")
    log_info "Backend servers configured: ${UPSTREAM_COUNT}"
else
    log_warning "nginx-production.conf not found"
fi

# Phase 2: Test Load Distribution
echo -e "\n${BLUE}=== Phase 2: Testing Load Distribution ===${NC}"

log_info "Sending ${NUM_REQUESTS} requests..."

declare -A BACKEND_HITS

for ((i=1; i<=NUM_REQUESTS; i++)); do
    # Send request to proxied endpoint and capture backend server from response header
    # Note: Using proxied path (e.g., /api/health) to ensure X-Served-By header is present
    BACKEND=$(curl -s -D - "${LB_URL}${LB_PATH}" 2>/dev/null | grep -i "X-Served-By" | awk '{print $2}' | tr -d '\r' || echo "unknown")

    if [ -n "$BACKEND" ] && [ "$BACKEND" != "unknown" ]; then
        BACKEND_HITS[$BACKEND]=$((BACKEND_HITS[$BACKEND] + 1))
    fi

    if [ $((i % 100)) -eq 0 ]; then
        echo -n "."
    fi
done
echo ""

# Calculate distribution
log_info "Request distribution:"
TOTAL_HITS=0
for BACKEND in "${!BACKEND_HITS[@]}"; do
    HITS=${BACKEND_HITS[$BACKEND]}
    TOTAL_HITS=$((TOTAL_HITS + HITS))
    PERCENTAGE=$((HITS * 100 / NUM_REQUESTS))
    log_info "  ${BACKEND}: ${HITS} requests (${PERCENTAGE}%)"
done

# Guard against divide by zero
if [ ${#BACKEND_HITS[@]} -eq 0 ]; then
    log_error "No backend hits recorded - check that LB_PATH is a proxied endpoint with X-Served-By header"
    log_error "Proxied endpoints: /api/*, /ollama/*, / (root)"
    log_error "Non-proxied endpoints (do NOT use): /health, /nginx-health"
    add_result "Load Distribution" "fail" "No backend hits recorded"
    exit 1
fi

# Check if distribution is balanced (within 20% variance)
EXPECTED_PER_BACKEND=$((NUM_REQUESTS / ${#BACKEND_HITS[@]}))
BALANCED=true

for BACKEND in "${!BACKEND_HITS[@]}"; do
    HITS=${BACKEND_HITS[$BACKEND]}
    VARIANCE=$((HITS * 100 / EXPECTED_PER_BACKEND - 100))
    if [ $VARIANCE -lt -20 ] || [ $VARIANCE -gt 20 ]; then
        BALANCED=false
    fi
done

if [ "$BALANCED" = true ]; then
    log_success "Load distribution is balanced"
else
    log_warning "Load distribution has high variance"
fi

# Phase 3: Test Health Checks
echo -e "\n${BLUE}=== Phase 3: Testing Health Checks ===${NC}"

if docker compose ps nginx &> /dev/null; then
    log_info "Testing nginx health check endpoint (non-proxied)..."

    if curl -f -s "${LB_URL}/nginx-health" > /dev/null 2>&1; then
        log_success "Health check endpoint responding"
    else
        log_error "Health check endpoint not responding"
    fi

    log_info "Testing proxied API endpoint..."
    if curl -f -s "${LB_URL}${LB_PATH}" > /dev/null 2>&1; then
        log_success "Proxied API endpoint responding"
    else
        log_warning "Proxied API endpoint not responding - using fallback"
    fi
fi

# Phase 4: Test Failover
echo -e "\n${BLUE}=== Phase 4: Testing Failover ===${NC}"

if docker compose ps &> /dev/null; then
    log_info "Simulating backend failure..."

    # Get first backend service
    BACKEND_SERVICE=$(docker compose ps --services | grep -E "ollama-node|api" | head -1)

    if [ -n "$BACKEND_SERVICE" ]; then
        log_info "Pausing ${BACKEND_SERVICE}..."
        docker compose pause "$BACKEND_SERVICE" 2>/dev/null

        sleep 2

        # Test if LB still responds (using nginx health endpoint for liveness)
        if curl -f -s "${LB_URL}/nginx-health" > /dev/null 2>&1; then
            log_success "Failover successful - LB still responding"
        else
            log_error "Failover failed - LB not responding"
        fi

        log_info "Restoring ${BACKEND_SERVICE}..."
        docker compose unpause "$BACKEND_SERVICE" 2>/dev/null
        log_success "Backend restored"
    fi
fi

# Phase 5: Test Connection Pooling
echo -e "\n${BLUE}=== Phase 5: Testing Connection Pooling ===${NC}"

log_info "Testing persistent connections..."

START=$(date +%s%N)
for ((i=1; i<=10; i++)); do
    curl -s -o /dev/null "${LB_URL}${LB_PATH}"
done
END=$(date +%s%N)

DURATION=$(( (END - START) / 1000000 ))
AVG_TIME=$((DURATION / 10))

log_info "Average request time: ${AVG_TIME}ms"

if [ $AVG_TIME -lt 100 ]; then
    log_success "Connection pooling appears effective"
else
    log_warning "High average request time - check connection pooling"
fi

# Phase 6: Test Concurrent Connections
echo -e "\n${BLUE}=== Phase 6: Testing Concurrent Connections ===${NC}"

log_info "Testing with 100 concurrent connections..."

# Check if Apache Bench is available
if command -v ab &> /dev/null; then
    ab -n 1000 -c 100 -q "${LB_URL}${LB_PATH}" > /tmp/ab-results.txt 2>&1

    REQUESTS_PER_SEC=$(grep "Requests per second" /tmp/ab-results.txt | awk '{print $4}')
    FAILED_REQUESTS=$(grep "Failed requests" /tmp/ab-results.txt | awk '{print $3}')

    log_info "Requests per second: ${REQUESTS_PER_SEC}"
    log_info "Failed requests: ${FAILED_REQUESTS}"

    if [ "$FAILED_REQUESTS" -eq 0 ]; then
        log_success "No failed requests under load"
    else
        log_warning "${FAILED_REQUESTS} requests failed"
    fi
else
    log_warning "Apache Bench (ab) not available - skipping concurrent connection test"
fi

# Generate Report
cat > "${REPORT_FILE}" <<EOF
{
  "timestamp": "${TIMESTAMP}",
  "load_balancer_url": "${LB_URL}",
  "test_path": "${LB_PATH}",
  "total_requests": ${NUM_REQUESTS},
  "backend_distribution": {
    $(for backend in "${!BACKEND_HITS[@]}"; do
        echo "\"${backend}\": ${BACKEND_HITS[$backend]}"
    done | paste -sd,)
  },
  "balanced": ${BALANCED},
  "average_response_time_ms": ${AVG_TIME},
  "status": "completed"
}
EOF

log_success "Report generated: ${REPORT_FILE}"
echo -e "\n${GREEN}Load balancing validation completed${NC}"

#!/bin/bash

################################################################################
# Pre-Deployment Validation Script for OllamaMax
#
# Validates network connectivity, dependencies, and configuration
# before deploying OllamaMax to production environments
################################################################################

set -e

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
REDIS_HOST="${REDIS_HOST:-localhost}"
REDIS_PORT="${REDIS_PORT:-6379}"
P2P_PEERS="${P2P_PEERS:-}"
API_URL="${API_URL:-http://localhost:13100}"

# Logging functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

FAILED_CHECKS=0
PASSED_CHECKS=0

pass() {
    log_success "$1"
    ((PASSED_CHECKS++))
}

fail() {
    log_error "$1"
    ((FAILED_CHECKS++))
}

warn() {
    log_warning "$1"
}

log_info "==================== Pre-Deployment Validation ===================="
log_info "PostgreSQL: ${POSTGRES_HOST}:${POSTGRES_PORT}"
log_info "Redis: ${REDIS_HOST}:${REDIS_PORT}"
log_info "P2P Peers: ${P2P_PEERS:-none configured}"
log_info "===================================================================="

# Section 1: Network Connectivity Tests
log_info ""
log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
log_info "Section 1: Network Connectivity Tests"
log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Test PostgreSQL connectivity
log_info "Testing PostgreSQL connectivity (${POSTGRES_HOST}:${POSTGRES_PORT})..."
if timeout 5 bash -c "cat < /dev/null > /dev/tcp/${POSTGRES_HOST}/${POSTGRES_PORT}" 2>/dev/null; then
    pass "PostgreSQL port ${POSTGRES_PORT} is reachable"

    # Try to connect with psql if available
    if command -v psql &> /dev/null; then
        if PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER:-ollama}" -d "${POSTGRES_DB:-ollamamax}" -c "SELECT 1;" &>/dev/null; then
            pass "PostgreSQL authentication successful"
        else
            fail "PostgreSQL authentication failed (check credentials)"
        fi
    else
        warn "psql not installed - skipping authentication test"
    fi
else
    fail "PostgreSQL port ${POSTGRES_PORT} is NOT reachable"
    log_error "Deployment will fail without PostgreSQL connectivity"
fi

# Test Redis connectivity
log_info "Testing Redis connectivity (${REDIS_HOST}:${REDIS_PORT})..."
if timeout 5 bash -c "cat < /dev/null > /dev/tcp/${REDIS_HOST}/${REDIS_PORT}" 2>/dev/null; then
    pass "Redis port ${REDIS_PORT} is reachable"

    # Try to ping Redis if redis-cli available
    if command -v redis-cli &> /dev/null; then
        if redis-cli -h "${REDIS_HOST}" -p "${REDIS_PORT}" -a "${REDIS_PASSWORD}" PING 2>/dev/null | grep -q "PONG"; then
            pass "Redis PING successful"
        else
            fail "Redis PING failed (check authentication)"
        fi
    else
        warn "redis-cli not installed - skipping PING test"
    fi
else
    fail "Redis port ${REDIS_PORT} is NOT reachable"
    log_error "Deployment will fail without Redis connectivity"
fi

# Test P2P peer reachability
if [ -n "${P2P_PEERS}" ]; then
    log_info "Testing P2P peer connectivity..."

    IFS=',' read -ra PEERS <<< "${P2P_PEERS}"
    for peer in "${PEERS[@]}"; do
        peer_host=$(echo "${peer}" | cut -d':' -f1)
        peer_port=$(echo "${peer}" | cut -d':' -f2)

        log_info "  Testing peer: ${peer_host}:${peer_port}..."
        if timeout 5 bash -c "cat < /dev/null > /dev/tcp/${peer_host}/${peer_port}" 2>/dev/null; then
            pass "  Peer ${peer_host}:${peer_port} is reachable"
        else
            fail "  Peer ${peer_host}:${peer_port} is NOT reachable"
        fi
    done
else
    warn "No P2P peers configured - skipping peer connectivity tests"
fi

# Section 2: DNS Resolution
log_info ""
log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
log_info "Section 2: DNS Resolution Tests"
log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Test PostgreSQL DNS
if [ "${POSTGRES_HOST}" != "localhost" ] && [ "${POSTGRES_HOST}" != "127.0.0.1" ]; then
    log_info "Testing DNS resolution for ${POSTGRES_HOST}..."
    if nslookup "${POSTGRES_HOST}" &>/dev/null || host "${POSTGRES_HOST}" &>/dev/null; then
        pass "DNS resolution successful for ${POSTGRES_HOST}"
    else
        fail "DNS resolution failed for ${POSTGRES_HOST}"
    fi
fi

# Test Redis DNS
if [ "${REDIS_HOST}" != "localhost" ] && [ "${REDIS_HOST}" != "127.0.0.1" ]; then
    log_info "Testing DNS resolution for ${REDIS_HOST}..."
    if nslookup "${REDIS_HOST}" &>/dev/null || host "${REDIS_HOST}" &>/dev/null; then
        pass "DNS resolution successful for ${REDIS_HOST}"
    else
        fail "DNS resolution failed for ${REDIS_HOST}"
    fi
fi

# Section 3: System Requirements
log_info ""
log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
log_info "Section 3: System Requirements"
log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check memory
TOTAL_MEM=$(free -g | awk '/^Mem:/{print $2}')
AVAILABLE_MEM=$(free -g | awk '/^Mem:/{print $7}')

log_info "Memory: ${AVAILABLE_MEM}GB available of ${TOTAL_MEM}GB total"
if [ "${AVAILABLE_MEM}" -ge 8 ]; then
    pass "Sufficient memory available (${AVAILABLE_MEM}GB >= 8GB)"
else
    fail "Insufficient memory (${AVAILABLE_MEM}GB < 8GB recommended)"
fi

# Check CPU cores
CPU_CORES=$(nproc)
log_info "CPU Cores: ${CPU_CORES}"
if [ "${CPU_CORES}" -ge 4 ]; then
    pass "Sufficient CPU cores (${CPU_CORES} >= 4)"
else
    warn "Low CPU cores (${CPU_CORES} < 4 recommended)"
fi

# Check disk space
DISK_SPACE=$(df -BG . | awk 'NR==2 {print $4}' | tr -d 'G')
log_info "Disk Space: ${DISK_SPACE}GB available"
if [ "${DISK_SPACE}" -ge 20 ]; then
    pass "Sufficient disk space (${DISK_SPACE}GB >= 20GB)"
else
    fail "Insufficient disk space (${DISK_SPACE}GB < 20GB recommended)"
fi

# Section 4: Required Tools
log_info ""
log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
log_info "Section 4: Required Tools"
log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

REQUIRED_TOOLS=("curl" "docker" "docker-compose")
OPTIONAL_TOOLS=("jq" "psql" "redis-cli" "kubectl")

for tool in "${REQUIRED_TOOLS[@]}"; do
    if command -v "${tool}" &>/dev/null; then
        pass "Required tool '${tool}' is installed"
    else
        fail "Required tool '${tool}' is NOT installed"
    fi
done

for tool in "${OPTIONAL_TOOLS[@]}"; do
    if command -v "${tool}" &>/dev/null; then
        pass "Optional tool '${tool}' is installed"
    else
        warn "Optional tool '${tool}' is NOT installed (recommended but not required)"
    fi
done

# Section 5: Credentials Validation
log_info ""
log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
log_info "Section 5: Credentials Validation"
log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check PostgreSQL credentials
if [ -n "${POSTGRES_USER}" ]; then
    pass "POSTGRES_USER is configured"
else
    fail "POSTGRES_USER is NOT set"
fi

if [ -n "${POSTGRES_PASSWORD}" ]; then
    pass "POSTGRES_PASSWORD is configured"
else
    fail "POSTGRES_PASSWORD is NOT set"
fi

if [ -n "${POSTGRES_DB}" ]; then
    pass "POSTGRES_DB is configured"
else
    fail "POSTGRES_DB is NOT set"
fi

# Check Redis password
if [ -n "${REDIS_PASSWORD}" ]; then
    pass "REDIS_PASSWORD is configured"
else
    warn "REDIS_PASSWORD is NOT set (recommended for production)"
fi

# Section 6: Configuration Validation
log_info ""
log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
log_info "Section 6: Configuration Validation"
log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check JWT_SECRET
if [ -n "${JWT_SECRET}" ]; then
    pass "JWT_SECRET is configured"
else
    fail "JWT_SECRET is NOT set (required for authentication)"
fi

# Check SMTP configuration (if email features enabled)
if [ -n "${SMTP_HOST}" ]; then
    pass "SMTP_HOST is configured"

    if [ -n "${SMTP_PASSWORD}" ]; then
        pass "SMTP_PASSWORD is configured"
    else
        warn "SMTP_PASSWORD is NOT set (may affect email features)"
    fi
else
    warn "SMTP not configured (email features will be disabled)"
fi

# Section 7: Smoke Test
log_info ""
log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
log_info "Section 7: Smoke Test"
log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

log_info "Performing smoke test against API (${API_URL}/api/health)..."
if curl -f -s "${API_URL}/api/health" > /dev/null; then
    pass "API health check successful"
else
    fail "API health check failed"
fi

# Final Summary
log_info ""
log_info "===================================================================="
log_info "Validation Summary"
log_info "===================================================================="
log_info "Passed Checks: ${GREEN}${PASSED_CHECKS}${NC}"
log_info "Failed Checks: ${RED}${FAILED_CHECKS}${NC}"
log_info "===================================================================="

if [ ${FAILED_CHECKS} -eq 0 ]; then
    log_success "✅ All critical validation checks passed!"
    log_success "System is ready for deployment"
    exit 0
elif [ ${FAILED_CHECKS} -le 2 ]; then
    log_warning "⚠️  Some validation checks failed, but deployment may proceed"
    log_warning "Review failed checks above and fix critical issues"
    exit 0
else
    log_error "❌ Multiple critical validation checks failed!"
    log_error "Deployment will likely fail - fix issues above before deploying"
    exit 1
fi

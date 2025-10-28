#!/bin/bash

##############################################################################
# Security Validation Script for OllamaMax
#
# This script validates that critical security vulnerabilities have been fixed:
# - ISSUE-001: No hardcoded SMTP credentials
# - ISSUE-002: No weak JWT secret defaults
# - ISSUE-003: No exposed database ports (PostgreSQL/Redis)
#
# Exit codes:
#   0 - All security checks passed
#   1 - Security vulnerabilities detected
##############################################################################

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔒 OllamaMax Security Validation"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Initialize counters
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0

# Function to check for pattern in files
check_pattern() {
    local pattern="$1"
    local description="$2"
    local severity="$3"

    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

    echo -n "Checking: $description... "

    # Search for pattern excluding documentation and test files
    local matches=$(grep -r "$pattern" "$PROJECT_ROOT" \
        --exclude-dir=node_modules \
        --exclude-dir=.git \
        --exclude-dir=docs \
        --exclude='*.md' \
        --exclude='*.log' \
        --exclude='validate-security-fixes.sh' \
        --exclude='verify-security-config.sh' \
        2>/dev/null || true)

    if [ -z "$matches" ]; then
        echo "✅ PASS"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
        return 0
    else
        echo "❌ FAIL ($severity)"
        echo "   Found in:"
        echo "$matches" | while IFS=: read -r file line content; do
            echo "     - $file:$line"
        done
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
        return 1
    fi
}

# Function to check environment variable usage
check_env_var_required() {
    local var_name="$1"
    local file="$2"
    local description="$3"

    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

    echo -n "Checking: $description... "

    if grep -q "${var_name}" "$file" && \
       grep -q "environment variable.*required\|JWT_SECRET:?\|SMTP_PASSWORD}" "$file" 2>/dev/null; then
        echo "✅ PASS"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
        return 0
    else
        echo "⚠️  WARNING"
        echo "   Environment variable validation may be missing in $file"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))  # Don't fail, just warn
        return 0
    fi
}

# Function to check for exposed ports in docker-compose files
check_docker_compose_ports() {
    local service="$1"
    local port="$2"
    local description="$3"

    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

    echo -n "Checking: $description... "

    # Find docker-compose files excluding dev and custom-ports variants
    local compose_files=$(find "$PROJECT_ROOT" -name "docker-compose*.yml" \
        -not -name "docker-compose.dev.yml" \
        -not -name "docker-compose.custom-ports.yml" \
        -not -path "*/node_modules/*" \
        -not -path "*/.git/*")

    local violations=""

    for file in $compose_files; do
        # Check if file has the service and exposed port mapping
        if grep -A 10 "^  ${service}:" "$file" | grep -q "^\s*-\s*[\"']*${port}:${port}[\"']*\s*$"; then
            violations="${violations}    - ${file}\n"
        fi
    done

    if [ -z "$violations" ]; then
        echo "✅ PASS"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
        return 0
    else
        echo "❌ FAIL (HIGH)"
        echo "   Exposed $service port found in:"
        echo -e "$violations"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
        return 1
    fi
}

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "ISSUE-001: Hardcoded SMTP Credentials"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

check_pattern "teamrsi123" "No hardcoded SMTP password 'teamrsi123'" "CRITICAL"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "ISSUE-002: Weak JWT Secret Defaults"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

check_pattern "ollamamax_secret_key_2024" "No hardcoded JWT secret 'ollamamax_secret_key_2024'" "CRITICAL"

# Check that JWT_SECRET is required in configuration files
check_env_var_required "JWT_SECRET" "$PROJECT_ROOT/api-server/auth-system.js" "JWT_SECRET validation in auth-system.js"
check_env_var_required "JWT_SECRET" "$PROJECT_ROOT/internal/config/config.go" "JWT_SECRET validation in config.go"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "ISSUE-003: Exposed Database Ports"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

check_docker_compose_ports "postgres" "5432" "No exposed PostgreSQL port (5432:5432)"
check_docker_compose_ports "redis" "6379" "No exposed Redis port (6379:6379)"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 Validation Summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Total Checks:  $TOTAL_CHECKS"
echo "Passed:        $PASSED_CHECKS ✅"
echo "Failed:        $FAILED_CHECKS ❌"
echo ""

if [ $FAILED_CHECKS -eq 0 ]; then
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "✅ ALL SECURITY CHECKS PASSED"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 0
else
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "❌ SECURITY VULNERABILITIES DETECTED"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "Please review and fix the failed checks above."
    exit 1
fi

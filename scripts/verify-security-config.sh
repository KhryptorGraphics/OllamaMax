#!/bin/bash

# Security Configuration Verification Script
# This script verifies that all critical security configurations are properly set

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counters
PASSED=0
FAILED=0
WARNINGS=0

echo "========================================="
echo "OllamaMax Security Configuration Check"
echo "========================================="
echo ""

# Function to check if variable is set and meets minimum requirements
check_required_var() {
    local var_name="$1"
    local min_length="$2"
    local var_value="${!var_name}"

    if [ -z "$var_value" ]; then
        echo -e "${RED}✗ FAILED${NC}: $var_name is not set"
        ((FAILED++))
        return 1
    fi

    if [ ${#var_value} -lt $min_length ]; then
        echo -e "${YELLOW}⚠ WARNING${NC}: $var_name is too short (${#var_value} chars, recommended: $min_length+)"
        ((WARNINGS++))
        return 2
    fi

    echo -e "${GREEN}✓ PASSED${NC}: $var_name is set and strong (${#var_value} chars)"
    ((PASSED++))
    return 0
}

# Function to check if variable is not using a known default
check_not_default() {
    local var_name="$1"
    shift
    local defaults=("$@")
    local var_value="${!var_name}"

    if [ -z "$var_value" ]; then
        echo -e "${RED}✗ FAILED${NC}: $var_name is not set"
        ((FAILED++))
        return 1
    fi

    for default in "${defaults[@]}"; do
        if [ "$var_value" = "$default" ]; then
            echo -e "${RED}✗ FAILED${NC}: $var_name is using insecure default value"
            ((FAILED++))
            return 1
        fi
    done

    echo -e "${GREEN}✓ PASSED${NC}: $var_name is not using default value"
    ((PASSED++))
    return 0
}

# Function to check CORS configuration
check_cors_production() {
    local var_value="${CORS_ALLOWED_ORIGINS}"

    if [ -z "$var_value" ]; then
        echo -e "${YELLOW}⚠ WARNING${NC}: CORS_ALLOWED_ORIGINS not set (will use localhost default)"
        ((WARNINGS++))
        return 2
    fi

    if echo "$var_value" | grep -q "localhost"; then
        echo -e "${YELLOW}⚠ WARNING${NC}: CORS_ALLOWED_ORIGINS contains localhost (development mode?)"
        ((WARNINGS++))
        return 2
    fi

    if echo "$var_value" | grep -q "\*"; then
        echo -e "${RED}✗ FAILED${NC}: CORS_ALLOWED_ORIGINS contains wildcard (insecure)"
        ((FAILED++))
        return 1
    fi

    echo -e "${GREEN}✓ PASSED${NC}: CORS_ALLOWED_ORIGINS configured for production"
    ((PASSED++))
    return 0
}

# Function to check file exists
check_file_exists() {
    local file_path="$1"
    local file_desc="$2"

    if [ ! -f "$file_path" ]; then
        echo -e "${RED}✗ FAILED${NC}: $file_desc not found at $file_path"
        ((FAILED++))
        return 1
    fi

    echo -e "${GREEN}✓ PASSED${NC}: $file_desc exists at $file_path"
    ((PASSED++))
    return 0
}

echo "=== Critical Security Variables ==="
echo ""

# Check JWT Secret
echo "Checking JWT_SECRET..."
check_required_var "JWT_SECRET" 32 || true
echo ""

# Alternative JWT_SECRET_KEY
if [ -z "$JWT_SECRET" ]; then
    echo "Checking JWT_SECRET_KEY (alternative)..."
    check_required_var "JWT_SECRET_KEY" 32 || true
    echo ""
fi

# Check SMTP Password
echo "Checking SMTP_PASSWORD..."
check_required_var "SMTP_PASSWORD" 16 || true
echo ""

# Check PostgreSQL Password
echo "Checking POSTGRES_PASSWORD..."
check_not_default "POSTGRES_PASSWORD" "secure_password" "" || true
check_required_var "POSTGRES_PASSWORD" 16 || true
echo ""

# Check Redis Password
echo "Checking REDIS_PASSWORD..."
check_not_default "REDIS_PASSWORD" "ollama_redis_pass" "" || true
check_required_var "REDIS_PASSWORD" 16 || true
echo ""

echo "=== CORS Configuration ==="
echo ""

# Check CORS Settings
echo "Checking CORS_ALLOWED_ORIGINS..."
check_cors_production || true
echo ""

echo "=== TLS/SSL Configuration ==="
echo ""

# Check TLS Enabled
if [ "$API_TLS_ENABLED" = "true" ]; then
    echo "Checking TLS certificates..."
    check_file_exists "${API_CERT_FILE:-}" "SSL Certificate" || true
    check_file_exists "${API_KEY_FILE:-}" "SSL Private Key" || true
else
    echo -e "${YELLOW}⚠ WARNING${NC}: API_TLS_ENABLED is not true (TLS disabled)"
    ((WARNINGS++))
fi
echo ""

echo "=== Additional Security Checks ==="
echo ""

# Check Grafana Password
echo "Checking GRAFANA_PASSWORD..."
check_not_default "GRAFANA_PASSWORD" "admin_password" "admin" "" || true
echo ""

# Check for hardcoded secrets in code
echo "Checking for hardcoded secrets in codebase..."
if grep -r "teamrsi123teamrsi123" . --include="*.js" --include="*.go" 2>/dev/null | grep -v "node_modules" | grep -v ".git" | grep -q .; then
    echo -e "${RED}✗ FAILED${NC}: Found hardcoded SMTP password in codebase"
    ((FAILED++))
else
    echo -e "${GREEN}✓ PASSED${NC}: No hardcoded SMTP passwords found"
    ((PASSED++))
fi
echo ""

echo "Checking for weak JWT defaults in codebase..."
if grep -r "ollamamax_secret_key_2024\|your-secret-key-change-this" . --include="*.js" --include="*.go" 2>/dev/null | grep -v "node_modules" | grep -v ".git" | grep -v "docs/" | grep -q .; then
    echo -e "${RED}✗ FAILED${NC}: Found weak JWT defaults in codebase"
    ((FAILED++))
else
    echo -e "${GREEN}✓ PASSED${NC}: No weak JWT defaults found"
    ((PASSED++))
fi
echo ""

# Check Docker Compose for exposed ports
if [ -f "docker-compose.yml" ]; then
    echo "Checking docker-compose.yml for exposed database ports..."
    if grep -A5 "postgres:" docker-compose.yml | grep -q "ports:"; then
        echo -e "${RED}✗ FAILED${NC}: PostgreSQL ports exposed in docker-compose.yml"
        ((FAILED++))
    else
        echo -e "${GREEN}✓ PASSED${NC}: PostgreSQL ports not exposed externally"
        ((PASSED++))
    fi

    if grep -A5 "redis:" docker-compose.yml | grep -q "ports:"; then
        echo -e "${RED}✗ FAILED${NC}: Redis ports exposed in docker-compose.yml"
        ((FAILED++))
    else
        echo -e "${GREEN}✓ PASSED${NC}: Redis ports not exposed externally"
        ((PASSED++))
    fi
fi
echo ""

# Summary
echo "========================================="
echo "Summary"
echo "========================================="
echo -e "${GREEN}Passed:${NC}   $PASSED"
echo -e "${YELLOW}Warnings:${NC} $WARNINGS"
echo -e "${RED}Failed:${NC}   $FAILED"
echo ""

if [ $FAILED -gt 0 ]; then
    echo -e "${RED}✗ SECURITY CHECK FAILED${NC}"
    echo "Please fix the failed checks before deploying to production."
    echo "See docs/ENVIRONMENT_VARIABLES.md for configuration guidance."
    exit 1
elif [ $WARNINGS -gt 0 ]; then
    echo -e "${YELLOW}⚠ SECURITY CHECK PASSED WITH WARNINGS${NC}"
    echo "Review the warnings above to ensure production readiness."
    exit 0
else
    echo -e "${GREEN}✓ ALL SECURITY CHECKS PASSED${NC}"
    echo "Configuration is ready for production deployment."
    exit 0
fi

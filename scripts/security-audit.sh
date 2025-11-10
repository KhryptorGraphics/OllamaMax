#!/bin/bash

# Security Audit Script for OllamaMax
# Performs comprehensive security checks

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

ISSUES_FOUND=0
WARNINGS_FOUND=0

log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
    ((WARNINGS_FOUND++))
}

log_error() {
    echo -e "${RED}✗${NC} $1"
    ((ISSUES_FOUND++))
}

log_header() {
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo ""
}

# Check if running from project root
if [ ! -f "package.json" ]; then
    log_error "Please run this script from the project root directory"
    exit 1
fi

log_header "OllamaMax Security Audit"

# 1. Environment Variables Check
log_header "1. Environment Variables Security"

if [ -f ".env" ]; then
    log_info "Checking .env file..."
    
    # Check for default/weak secrets
    if grep -q "dev-secret-key" .env; then
        log_error "Default JWT secret detected in .env - MUST change for production"
    else
        log_success "JWT secret appears to be customized"
    fi
    
    # Check JWT secret length
    JWT_SECRET=$(grep "^JWT_SECRET=" .env | cut -d'=' -f2)
    if [ ${#JWT_SECRET} -lt 32 ]; then
        log_error "JWT secret is too short (< 32 characters)"
    else
        log_success "JWT secret length is adequate"
    fi
    
    # Check if .env is in .gitignore
    if grep -q "^\.env$" .gitignore 2>/dev/null; then
        log_success ".env is in .gitignore"
    else
        log_error ".env is NOT in .gitignore - secrets may be committed!"
    fi
    
    # Check for exposed credentials
    if grep -qi "password.*=.*password\|password.*=.*123" .env; then
        log_warning "Weak or default passwords detected in .env"
    fi
else
    log_warning ".env file not found"
fi

# 2. Dependency Security
log_header "2. Dependency Security Audit"

log_info "Running npm audit..."
if npm audit --audit-level=moderate > /tmp/npm-audit.txt 2>&1; then
    log_success "No moderate or higher vulnerabilities found"
else
    VULN_COUNT=$(grep -c "vulnerabilities" /tmp/npm-audit.txt || echo "0")
    if [ "$VULN_COUNT" -gt 0 ]; then
        log_error "npm audit found vulnerabilities - run 'npm audit fix'"
        cat /tmp/npm-audit.txt
    fi
fi

# 3. File Permissions
log_header "3. File Permissions Check"

# Check sensitive files
if [ -f ".env" ]; then
    PERMS=$(stat -c "%a" .env 2>/dev/null || stat -f "%A" .env 2>/dev/null)
    if [ "$PERMS" = "600" ] || [ "$PERMS" = "400" ]; then
        log_success ".env has secure permissions ($PERMS)"
    else
        log_warning ".env permissions are $PERMS (should be 600 or 400)"
    fi
fi

# Check certificate files
if [ -d "certs" ]; then
    for file in certs/*.key; do
        if [ -f "$file" ]; then
            PERMS=$(stat -c "%a" "$file" 2>/dev/null || stat -f "%A" "$file" 2>/dev/null)
            if [ "$PERMS" = "600" ] || [ "$PERMS" = "400" ]; then
                log_success "$(basename $file) has secure permissions"
            else
                log_error "$(basename $file) permissions are $PERMS (should be 600)"
            fi
        fi
    done
fi

# 4. Code Security Patterns
log_header "4. Code Security Patterns"

log_info "Checking for security anti-patterns..."

# Check for hardcoded secrets
if grep -r "password.*=.*['\"]" src/ --include="*.js" | grep -v "process.env" | grep -q .; then
    log_error "Potential hardcoded passwords found in source code"
else
    log_success "No hardcoded passwords detected"
fi

# Check for SQL injection vulnerabilities
if grep -r "query.*+.*req\." src/ --include="*.js" | grep -q .; then
    log_warning "Potential SQL injection vulnerability (string concatenation in queries)"
else
    log_success "No obvious SQL injection patterns found"
fi

# Check for eval usage
if grep -r "eval(" src/ --include="*.js" | grep -q .; then
    log_error "eval() usage detected - security risk"
else
    log_success "No eval() usage found"
fi

# Check for console.log with sensitive data
if grep -r "console\.log.*password\|console\.log.*token\|console\.log.*secret" src/ --include="*.js" | grep -q .; then
    log_warning "Potential sensitive data logging detected"
fi

# 5. Authentication & Authorization
log_header "5. Authentication & Authorization"

log_info "Checking authentication implementation..."

# Check if JWT middleware is used
if grep -r "authMiddleware" src/server.js | grep -q .; then
    log_success "Authentication middleware is implemented"
else
    log_warning "Authentication middleware usage unclear"
fi

# Check for rate limiting
if grep -r "rate.*limit\|express-rate-limit" src/ --include="*.js" | grep -q .; then
    log_success "Rate limiting appears to be implemented"
else
    log_warning "Rate limiting not detected"
fi

# 6. HTTPS/TLS Configuration
log_header "6. HTTPS/TLS Configuration"

if [ -f "src/https-server.js" ]; then
    log_success "HTTPS server module exists"
    
    # Check cipher configuration
    if grep -q "ciphers" src/https-server.js; then
        log_success "Custom cipher configuration found"
    else
        log_warning "No custom cipher configuration"
    fi
else
    log_warning "HTTPS server module not found"
fi

# 7. Input Validation
log_header "7. Input Validation"

# Check for input validation
if grep -r "validator\|joi\|express-validator" package.json | grep -q .; then
    log_success "Input validation library detected"
else
    log_warning "No input validation library found"
fi

# 8. CORS Configuration
log_header "8. CORS Configuration"

if grep -r "cors" src/server.js | grep -q .; then
    log_success "CORS is configured"

    # Check if CORS is too permissive
    if grep -r "origin.*\*" src/server.js | grep -q .; then
        log_warning "CORS allows all origins (*) - should restrict in production"
    else
        log_success "CORS origin appears to be restricted"
    fi
else
    log_warning "CORS configuration not found"
fi

# 9. Security Headers
log_header "9. Security Headers"

if grep -r "helmet" src/server.js | grep -q .; then
    log_success "Helmet security headers middleware is used"
else
    log_error "Helmet middleware not detected - missing security headers"
fi

# 10. Database Security
log_header "10. Database Security"

# Check for parameterized queries
if grep -r "prepare\|placeholder\|\?" src/models/ --include="*.js" | grep -q .; then
    log_success "Parameterized queries detected"
else
    log_warning "Parameterized queries not clearly detected"
fi

# Check database file permissions
if [ -f "data/ollamamax.db" ]; then
    PERMS=$(stat -c "%a" data/ollamamax.db 2>/dev/null || stat -f "%A" data/ollamamax.db 2>/dev/null)
    if [ "$PERMS" = "600" ] || [ "$PERMS" = "640" ]; then
        log_success "Database file has secure permissions"
    else
        log_warning "Database file permissions are $PERMS (should be 600 or 640)"
    fi
fi

# 11. Logging Security
log_header "11. Logging Security"

# Check if logs directory exists and has proper permissions
if [ -d "logs" ]; then
    PERMS=$(stat -c "%a" logs 2>/dev/null || stat -f "%A" logs 2>/dev/null)
    if [ "$PERMS" = "750" ] || [ "$PERMS" = "755" ]; then
        log_success "Logs directory has appropriate permissions"
    else
        log_warning "Logs directory permissions are $PERMS"
    fi
fi

# 12. Docker Security (if using Docker)
log_header "12. Docker Security"

if [ -f "Dockerfile" ]; then
    log_info "Checking Dockerfile..."

    # Check if running as root
    if grep -q "USER" Dockerfile; then
        log_success "Dockerfile uses non-root user"
    else
        log_warning "Dockerfile may run as root - should use USER directive"
    fi

    # Check for latest tag
    if grep -q "FROM.*:latest" Dockerfile; then
        log_warning "Dockerfile uses :latest tag - should pin versions"
    else
        log_success "Dockerfile pins image versions"
    fi
fi

# Summary
log_header "Security Audit Summary"

echo ""
if [ $ISSUES_FOUND -eq 0 ] && [ $WARNINGS_FOUND -eq 0 ]; then
    log_success "No security issues or warnings found!"
    echo ""
    echo -e "${GREEN}✨ Security audit passed!${NC}"
else
    echo -e "${RED}Critical Issues: $ISSUES_FOUND${NC}"
    echo -e "${YELLOW}Warnings: $WARNINGS_FOUND${NC}"
    echo ""

    if [ $ISSUES_FOUND -gt 0 ]; then
        echo -e "${RED}⚠️  Please fix critical issues before deploying to production${NC}"
        exit 1
    else
        echo -e "${YELLOW}⚠️  Please review warnings and fix if applicable${NC}"
    fi
fi

echo ""
log_info "For more security best practices, see: docs/SECURITY.md"
echo ""



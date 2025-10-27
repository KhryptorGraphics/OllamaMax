#!/bin/bash
# Security Configuration Validation Script
# References: production-security.yaml, .github/workflows/ci-cd-pipeline.yml

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
REPORT_FILE="${REPORT_DIR}/security-validation-${TIMESTAMP}.json"

mkdir -p "${REPORT_DIR}"

TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNING_CHECKS=0
VALIDATION_RESULTS=()

echo -e "${BLUE}=== Security Configuration Validation ===${NC}"

# Helper functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
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
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    VALIDATION_RESULTS+=("{\"name\":\"$1\",\"status\":\"$2\",\"message\":\"$3\"}")
}

# Phase 1: Vulnerability Scanning
echo -e "\n${BLUE}=== Phase 1: Vulnerability Scanning ===${NC}"

# Trivy scan
if command -v trivy &> /dev/null; then
    log_info "Running Trivy filesystem scan..."
    if trivy fs --severity HIGH,CRITICAL --exit-code 1 . > /dev/null 2>&1; then
        log_success "Trivy scan passed"
        add_result "Trivy Scan" "pass" "No critical vulnerabilities"
    else
        log_error "Trivy scan found vulnerabilities"
        add_result "Trivy Scan" "fail" "Critical vulnerabilities found"
    fi
else
    log_warning "Trivy not installed - skipping scan"
    add_result "Trivy Scan" "warning" "Trivy not available"
fi

# Snyk scan
if command -v snyk &> /dev/null; then
    log_info "Running Snyk dependency scan..."
    if snyk test --severity-threshold=high > /dev/null 2>&1; then
        log_success "Snyk scan passed"
        add_result "Snyk Scan" "pass" "No high-severity issues"
    else
        log_error "Snyk scan found issues"
        add_result "Snyk Scan" "fail" "High-severity issues found"
    fi
else
    log_warning "Snyk not installed - skipping scan"
    add_result "Snyk Scan" "warning" "Snyk not available"
fi

# Phase 2: TLS/SSL Validation
echo -e "\n${BLUE}=== Phase 2: TLS/SSL Configuration ===${NC}"

# Check nginx SSL configuration
if [ -f "nginx/nginx-production.conf" ]; then
    log_info "Checking nginx SSL configuration..."

    if grep -q "ssl_protocols TLSv1.3" nginx/nginx-production.conf; then
        log_success "TLS 1.3 configured"
        add_result "TLS Version" "pass" "TLS 1.3 enforced"
    else
        log_error "TLS 1.3 not enforced"
        add_result "TLS Version" "fail" "TLS 1.3 not enforced"
    fi

    if grep -q "ssl_prefer_server_ciphers" nginx/nginx-production.conf; then
        log_success "Cipher preferences configured"
        add_result "SSL Ciphers" "pass" "Server cipher preferences set"
    else
        log_warning "Cipher preferences not configured"
        add_result "SSL Ciphers" "warning" "Cipher preferences missing"
    fi

    if grep -q "ssl_stapling on" nginx/nginx-production.conf; then
        log_success "OCSP stapling enabled"
        add_result "OCSP Stapling" "pass" "OCSP stapling enabled"
    else
        log_warning "OCSP stapling not enabled"
        add_result "OCSP Stapling" "warning" "OCSP stapling disabled"
    fi
fi

# Phase 3: Security Headers
echo -e "\n${BLUE}=== Phase 3: Security Headers ===${NC}"

if [ -f "nginx/nginx-production.conf" ]; then
    REQUIRED_HEADERS=(
        "Strict-Transport-Security"
        "X-Content-Type-Options"
        "X-Frame-Options"
        "X-XSS-Protection"
        "Content-Security-Policy"
    )

    for HEADER in "${REQUIRED_HEADERS[@]}"; do
        if grep -q "$HEADER" nginx/nginx-production.conf; then
            log_success "Header configured: ${HEADER}"
            add_result "Header: ${HEADER}" "pass" "Configured"
        else
            log_error "Header missing: ${HEADER}"
            add_result "Header: ${HEADER}" "fail" "Not configured"
        fi
    done
fi

# Phase 4: Authentication Testing
echo -e "\n${BLUE}=== Phase 4: Authentication Configuration ===${NC}"

# Check for JWT configuration
if [ -f ".env" ] || [ -f ".env.example" ]; then
    log_info "Checking authentication configuration..."

    if grep -qE "(JWT_SECRET|AUTH_SECRET)" .env .env.example 2>/dev/null; then
        log_success "Authentication secrets configured"
        add_result "Auth Secrets" "pass" "Secrets configured"
    else
        log_warning "Authentication secrets not found"
        add_result "Auth Secrets" "warning" "Secrets not configured"
    fi
fi

# Phase 5: Secrets Management
echo -e "\n${BLUE}=== Phase 5: Secrets Management ===${NC}"

log_info "Scanning for exposed secrets..."

# Check for hardcoded secrets in config files
EXPOSED_SECRETS=0
FILES_TO_CHECK=("docker-compose.yml" "docker-compose.prod.yml" ".env" "k8s/*.yaml")

for PATTERN in "${FILES_TO_CHECK[@]}"; do
    for FILE in $PATTERN; do
        if [ -f "$FILE" ]; then
            # Read file, search for sensitive keywords, then filter out template variables
            if grep -E "(password|secret|key|token)" "$FILE" 2>/dev/null | grep -vE '\$\{|\$\(' > /dev/null 2>&1; then
                EXPOSED_SECRETS=$((EXPOSED_SECRETS + 1))
            fi
        fi
    done
done

if [ $EXPOSED_SECRETS -eq 0 ]; then
    log_success "No exposed secrets detected"
    add_result "Secrets Exposure" "pass" "No hardcoded secrets"
else
    log_error "${EXPOSED_SECRETS} potential exposed secret(s)"
    add_result "Secrets Exposure" "fail" "${EXPOSED_SECRETS} exposed secrets"
fi

# Phase 6: Network Security
echo -e "\n${BLUE}=== Phase 6: Network Security ===${NC}"

# Check rate limiting configuration
if [ -f "nginx/nginx-production.conf" ]; then
    if grep -q "limit_req_zone" nginx/nginx-production.conf; then
        log_success "Rate limiting configured"
        add_result "Rate Limiting" "pass" "Rate limiting enabled"
    else
        log_error "Rate limiting not configured"
        add_result "Rate Limiting" "fail" "Rate limiting disabled"
    fi
fi

# Check Kubernetes network policies
if command -v kubectl &> /dev/null; then
    log_info "Checking Kubernetes network policies..."
    NETPOL_COUNT=$(kubectl get networkpolicies --all-namespaces --no-headers 2>/dev/null | wc -l || echo "0")

    if [ "$NETPOL_COUNT" -gt 0 ]; then
        log_success "${NETPOL_COUNT} network polic(ies) configured"
        add_result "Network Policies" "pass" "${NETPOL_COUNT} policies configured"
    else
        log_warning "No network policies configured"
        add_result "Network Policies" "warning" "No policies configured"
    fi
fi

# Phase 7: Docker Image Security
echo -e "\n${BLUE}=== Phase 7: Docker Image Security ===${NC}"

# Scan Docker images if available
IMAGES=$(docker images --format "{{.Repository}}:{{.Tag}}" | grep -E "ollama|api|web" || true)

for IMAGE in $IMAGES; do
    if command -v trivy &> /dev/null; then
        log_info "Scanning image: ${IMAGE}"
        if trivy image --severity HIGH,CRITICAL --exit-code 1 "$IMAGE" > /dev/null 2>&1; then
            log_success "Image ${IMAGE}: No critical vulnerabilities"
            add_result "Image: ${IMAGE}" "pass" "Secure"
        else
            log_error "Image ${IMAGE}: Vulnerabilities found"
            add_result "Image: ${IMAGE}" "fail" "Vulnerabilities detected"
        fi
    fi
done

# Phase 8: Audit Logging
echo -e "\n${BLUE}=== Phase 8: Audit Logging ===${NC}"

if [ -f "production-security.yaml" ]; then
    if grep -q "audit_logging" production-security.yaml; then
        log_success "Audit logging configuration found"
        add_result "Audit Logging" "pass" "Configured"
    else
        log_warning "Audit logging configuration not found"
        add_result "Audit Logging" "warning" "Not configured"
    fi
fi

# Phase 9: OWASP ZAP Baseline Scan
echo -e "\n${BLUE}=== Phase 9: OWASP ZAP Baseline Scan ===${NC}"

if command -v docker &> /dev/null; then
    log_info "Running OWASP ZAP baseline scan..."

    # Check if service is available
    if curl -f http://localhost:10080 > /dev/null 2>&1; then
        ZAP_REPORT="${REPORT_DIR}/zap-report.html"

        if docker run --rm -v "$(pwd)/${REPORT_DIR}:/zap/wrk:rw" \
            --network host \
            owasp/zap2docker-stable:latest \
            zap-baseline.py -t http://localhost:10080 -r zap-report.html > /dev/null 2>&1; then
            log_success "OWASP ZAP scan completed - no high alerts"
            add_result "OWASP ZAP Scan" "pass" "No high/medium alerts"
        else
            log_error "OWASP ZAP scan found vulnerabilities"
            add_result "OWASP ZAP Scan" "fail" "High/medium alerts detected"
        fi
    else
        log_warning "Application not available on localhost:10080 - skipping ZAP scan"
        add_result "OWASP ZAP Scan" "warning" "Application not available"
    fi
else
    log_warning "Docker not available - skipping OWASP ZAP scan"
    add_result "OWASP ZAP Scan" "warning" "Docker not available"
fi

# Generate Report
VALIDATION_SCORE=$((PASSED_CHECKS * 100 / TOTAL_CHECKS))

cat > "${REPORT_FILE}" <<EOF
{
  "timestamp": "${TIMESTAMP}",
  "summary": {
    "total_checks": ${TOTAL_CHECKS},
    "passed": ${PASSED_CHECKS},
    "failed": ${FAILED_CHECKS},
    "warnings": ${WARNING_CHECKS},
    "security_score": ${VALIDATION_SCORE}
  },
  "results": [
    $(IFS=,; echo "${VALIDATION_RESULTS[*]}")
  ]
}
EOF

log_success "Security report generated: ${REPORT_FILE}"

# Summary
echo -e "\n${BLUE}=== Security Validation Summary ===${NC}"
echo "Total Checks: ${TOTAL_CHECKS}"
echo -e "Passed: ${GREEN}${PASSED_CHECKS}${NC}"
echo -e "Failed: ${RED}${FAILED_CHECKS}${NC}"
echo -e "Warnings: ${YELLOW}${WARNING_CHECKS}${NC}"
echo "Security Score: ${VALIDATION_SCORE}/100"

if [ "$FAILED_CHECKS" -gt 0 ]; then
    echo -e "\n${RED}Security validation FAILED${NC}"
    exit 1
else
    echo -e "\n${GREEN}Security validation PASSED${NC}"
    exit 0
fi

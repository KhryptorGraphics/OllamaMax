#!/bin/bash

################################################################################
# Security Penetration Testing Execution Script for OllamaMax
#
# Executes comprehensive security testing from penetration_test.go
# Validates OWASP Top 10 compliance and security best practices
################################################################################

set -e

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
RESULTS_DIR="security-test-results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
TARGET_URL="${TARGET_URL:-http://localhost:11434}"

# Logging functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

mkdir -p "${RESULTS_DIR}"

log_info "==================== Security Penetration Testing ===================="
log_info "Results Directory: ${RESULTS_DIR}"
log_info "Target URL: ${TARGET_URL}"
log_info "======================================================================"

# Check if penetration tests exist
if [ ! -f "ollama-distributed/tests/security/penetration_test.go" ]; then
    log_error "Penetration tests not found!"
    exit 1
fi

log_success "Penetration tests found"

# Execute OWASP Top 10 Tests
log_info "Executing OWASP Top 10 security tests..."

cd ollama-distributed

OWASP_TESTS=(
    "A01_BrokenAccessControl"
    "A02_CryptographicFailures"
    "A03_Injection"
    "A04_InsecureDesign"
    "A05_SecurityMisconfiguration"
    "A06_VulnerableComponents"
    "A07_AuthenticationFailures"
    "A08_IntegrityFailures"
    "A09_LoggingFailures"
    "A10_SSRF"
)

PASSED_OWASP=0
FAILED_OWASP=0

for test in "${OWASP_TESTS[@]}"; do
    log_info "Running OWASP test: ${test}..."

    if go test -v -timeout 15m ./tests/security/... -run "TestOWASPTop10/${test}" \
        > "../${RESULTS_DIR}/owasp-${test}-${TIMESTAMP}.log" 2>&1; then
        log_success "OWASP ${test} passed"
        ((PASSED_OWASP++))
    else
        log_error "OWASP ${test} failed"
        ((FAILED_OWASP++))
    fi
done

# Execute comprehensive security penetration tests
log_info "Executing comprehensive security penetration tests..."

SECURITY_TESTS=(
    "Authentication"
    "Authorization"
    "InputValidation"
    "InjectionAttacks"
    "RateLimiting"
    "CORSSecurity"
    "TLSSecurity"
    "SessionSecurity"
    "InformationDisclosure"
    "DoSProtection"
)

PASSED_SECURITY=0
FAILED_SECURITY=0

for test in "${SECURITY_TESTS[@]}"; do
    log_info "Running security test: ${test}..."

    if go test -v -timeout 15m ./tests/security/... -run "TestSecurityPenetration.*${test}" \
        > "../${RESULTS_DIR}/security-${test}-${TIMESTAMP}.log" 2>&1; then
        log_success "Security ${test} passed"
        ((PASSED_SECURITY++))
    else
        log_error "Security ${test} failed"
        ((FAILED_SECURITY++))
    fi
done

cd ..

# Automated security scanning with OWASP ZAP (if available)
if command -v docker &> /dev/null; then
    log_info "Running OWASP ZAP baseline scan..."

    docker run --rm --network="host" \
        -v "$(pwd)/${RESULTS_DIR}:/zap/wrk:rw" \
        owasp/zap2docker-stable:latest \
        zap-baseline.py \
        -t "${TARGET_URL}" \
        -r "zap-baseline-report-${TIMESTAMP}.html" \
        -J "zap-baseline-report-${TIMESTAMP}.json" \
        > "${RESULTS_DIR}/zap-baseline-${TIMESTAMP}.log" 2>&1 || {
        log_warning "OWASP ZAP scan completed with warnings (exit code $?)"
    }

    log_success "OWASP ZAP scan completed"
else
    log_warning "Docker not available, skipping OWASP ZAP scan"
fi

# Vulnerability scanning with Snyk (if configured)
if command -v snyk &> /dev/null && [ -f "package.json" ]; then
    log_info "Running Snyk vulnerability scan..."

    snyk test --json > "${RESULTS_DIR}/snyk-vulnerabilities-${TIMESTAMP}.json" 2>&1 || {
        log_warning "Snyk scan found vulnerabilities (exit code $?)"
    }

    log_success "Snyk scan completed"
else
    log_info "Snyk not configured, skipping dependency vulnerability scan"
fi

# Container security scanning with Trivy (if available)
if command -v trivy &> /dev/null; then
    log_info "Running Trivy container security scan..."

    # Scan main application image if it exists
    if docker images | grep -q "ollamamax"; then
        trivy image --format json --output "${RESULTS_DIR}/trivy-scan-${TIMESTAMP}.json" \
            ollamamax:latest > "${RESULTS_DIR}/trivy-output-${TIMESTAMP}.log" 2>&1 || {
            log_warning "Trivy found vulnerabilities (exit code $?)"
        }
        log_success "Trivy scan completed"
    else
        log_info "OllamaMax image not found, skipping Trivy scan"
    fi
else
    log_info "Trivy not installed, skipping container security scan"
fi

# Static analysis with SonarQube (if configured)
if [ -n "${SONARQUBE_URL}" ] && [ -n "${SONARQUBE_TOKEN}" ]; then
    log_info "Running SonarQube security analysis..."

    if command -v sonar-scanner &> /dev/null; then
        sonar-scanner \
            -Dsonar.host.url="${SONARQUBE_URL}" \
            -Dsonar.login="${SONARQUBE_TOKEN}" \
            -Dsonar.projectKey=ollamamax \
            -Dsonar.sources=. \
            > "${RESULTS_DIR}/sonarqube-${TIMESTAMP}.log" 2>&1 || {
            log_warning "SonarQube analysis completed with issues"
        }
        log_success "SonarQube analysis completed"
    else
        log_warning "sonar-scanner not installed"
    fi
else
    log_info "SonarQube not configured, skipping static analysis"
fi

# Analyze and categorize vulnerabilities
log_info "Analyzing security findings..."

CRITICAL_VULNS=0
HIGH_VULNS=0
MEDIUM_VULNS=0
LOW_VULNS=0

# Count vulnerabilities from test results
for log_file in ${RESULTS_DIR}/*-${TIMESTAMP}.log; do
    if grep -qi "critical" "${log_file}"; then
        ((CRITICAL_VULNS++))
    fi
    if grep -qi "high" "${log_file}"; then
        ((HIGH_VULNS++))
    fi
    if grep -qi "medium" "${log_file}"; then
        ((MEDIUM_VULNS++))
    fi
    if grep -qi "low" "${log_file}"; then
        ((LOW_VULNS++))
    fi
done

# Calculate security score
TOTAL_TESTS=$((PASSED_OWASP + FAILED_OWASP + PASSED_SECURITY + FAILED_SECURITY))
PASSED_TESTS=$((PASSED_OWASP + PASSED_SECURITY))
SECURITY_SCORE=$(echo "scale=2; ${PASSED_TESTS} * 100 / ${TOTAL_TESTS}" | bc)

# Generate security assessment report
REPORT_FILE="${RESULTS_DIR}/security-assessment-report-${TIMESTAMP}.md"

cat > "${REPORT_FILE}" <<EOF
# Security Penetration Testing Report

**Timestamp:** $(date --iso-8601=seconds)
**Target URL:** ${TARGET_URL}
**Security Score:** ${SECURITY_SCORE}/100

## OWASP Top 10 Compliance

**Total Tests:** $((PASSED_OWASP + FAILED_OWASP))
**Passed:** ${PASSED_OWASP}
**Failed:** ${FAILED_OWASP}
**Compliance Rate:** $(echo "scale=2; ${PASSED_OWASP} * 100 / (${PASSED_OWASP} + ${FAILED_OWASP})" | bc)%

### Test Results by Category

EOF

for test in "${OWASP_TESTS[@]}"; do
    if grep -q "PASS" "${RESULTS_DIR}/owasp-${test}-${TIMESTAMP}.log" 2>/dev/null; then
        echo "- ✅ ${test}: PASSED" >> "${REPORT_FILE}"
    else
        echo "- ❌ ${test}: FAILED" >> "${REPORT_FILE}"
    fi
done

cat >> "${REPORT_FILE}" <<EOF

## Comprehensive Security Tests

**Total Tests:** $((PASSED_SECURITY + FAILED_SECURITY))
**Passed:** ${PASSED_SECURITY}
**Failed:** ${FAILED_SECURITY}

### Test Results

EOF

for test in "${SECURITY_TESTS[@]}"; do
    if grep -q "PASS" "${RESULTS_DIR}/security-${test}-${TIMESTAMP}.log" 2>/dev/null; then
        echo "- ✅ ${test}: PASSED" >> "${REPORT_FILE}"
    else
        echo "- ❌ ${test}: FAILED" >> "${REPORT_FILE}"
    fi
done

cat >> "${REPORT_FILE}" <<EOF

## Vulnerability Summary

- **Critical:** ${CRITICAL_VULNS}
- **High:** ${HIGH_VULNS}
- **Medium:** ${MEDIUM_VULNS}
- **Low:** ${LOW_VULNS}

## Automated Security Scans

EOF

if [ -f "${RESULTS_DIR}/zap-baseline-report-${TIMESTAMP}.html" ]; then
    echo "- ✅ OWASP ZAP baseline scan completed" >> "${REPORT_FILE}"
else
    echo "- ⚠️ OWASP ZAP scan not performed" >> "${REPORT_FILE}"
fi

if [ -f "${RESULTS_DIR}/snyk-vulnerabilities-${TIMESTAMP}.json" ]; then
    echo "- ✅ Snyk dependency scan completed" >> "${REPORT_FILE}"
else
    echo "- ⚠️ Snyk scan not performed" >> "${REPORT_FILE}"
fi

if [ -f "${RESULTS_DIR}/trivy-scan-${TIMESTAMP}.json" ]; then
    echo "- ✅ Trivy container scan completed" >> "${REPORT_FILE}"
else
    echo "- ⚠️ Trivy scan not performed" >> "${REPORT_FILE}"
fi

cat >> "${REPORT_FILE}" <<EOF

## Findings and Recommendations

### Critical Issues

EOF

if [ ${CRITICAL_VULNS} -eq 0 ]; then
    echo "✅ No critical vulnerabilities found" >> "${REPORT_FILE}"
else
    echo "❌ ${CRITICAL_VULNS} critical vulnerabilities require immediate attention" >> "${REPORT_FILE}"
    echo "" >> "${REPORT_FILE}"
    echo "Review detailed findings in test logs: \`${RESULTS_DIR}/*-${TIMESTAMP}.log\`" >> "${REPORT_FILE}"
fi

cat >> "${REPORT_FILE}" <<EOF

### High Priority Issues

EOF

if [ ${HIGH_VULNS} -eq 0 ]; then
    echo "✅ No high-priority vulnerabilities found" >> "${REPORT_FILE}"
else
    echo "⚠️ ${HIGH_VULNS} high-priority vulnerabilities should be addressed before production" >> "${REPORT_FILE}"
fi

cat >> "${REPORT_FILE}" <<EOF

### Recommendations

1. **Immediate Actions:**
   - Address all critical vulnerabilities before production deployment
   - Implement security patches for failed OWASP tests
   - Review and fix authentication/authorization issues

2. **Short-term Actions:**
   - Address high-priority vulnerabilities
   - Implement security monitoring and alerting
   - Conduct regular security assessments

3. **Long-term Actions:**
   - Establish security training program
   - Implement automated security testing in CI/CD
   - Regular penetration testing schedule

## Compliance Status

EOF

if [ ${CRITICAL_VULNS} -eq 0 ] && [ ${FAILED_OWASP} -eq 0 ]; then
    echo "✅ **SECURITY VALIDATED** - System meets OWASP Top 10 security standards" >> "${REPORT_FILE}"
elif [ ${CRITICAL_VULNS} -eq 0 ] && [ ${FAILED_OWASP} -le 2 ]; then
    echo "⚠️ **CONDITIONAL PASS** - Minor security issues identified, review recommended" >> "${REPORT_FILE}"
else
    echo "❌ **SECURITY CONCERNS** - Critical issues must be resolved before production" >> "${REPORT_FILE}"
fi

cat >> "${REPORT_FILE}" <<EOF

## Detailed Test Logs

All detailed test logs and vulnerability reports are available in:
\`${RESULTS_DIR}/*-${TIMESTAMP}.*\`

## Next Steps

1. Review all failed tests and vulnerability findings
2. Create remediation plan with priorities
3. Implement security fixes and retest
4. Update security documentation
5. Establish continuous security monitoring

---

**Report Generated:** $(date --iso-8601=seconds)
**Report Version:** 1.0
EOF

log_success "Security assessment report: ${REPORT_FILE}"

# Final summary
log_info "=========================================================================="
log_info "Security Testing Summary:"
log_info "  OWASP Top 10: ${PASSED_OWASP}/${FAILED_OWASP} passed"
log_info "  Security Tests: ${PASSED_SECURITY}/${FAILED_SECURITY} passed"
log_info "  Security Score: ${SECURITY_SCORE}/100"
log_info "  Critical Vulnerabilities: ${CRITICAL_VULNS}"
log_info "  High Vulnerabilities: ${HIGH_VULNS}"
log_info "=========================================================================="

# Exit with appropriate code
if [ ${CRITICAL_VULNS} -eq 0 ] && [ ${FAILED_OWASP} -eq 0 ]; then
    log_success "Security testing passed!"
    exit 0
else
    log_error "Security testing found critical issues!"
    exit 1
fi

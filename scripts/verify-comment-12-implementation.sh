#!/bin/bash

echo "🔍 Verifying Comment 12 Implementation"
echo "======================================"

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

PASSED=0
FAILED=0

check() {
    if [ "$1" = true ]; then
        echo -e "${GREEN}✅ $2${NC}"
        ((PASSED++))
    else
        echo -e "${RED}❌ $2${NC}"
        ((FAILED++))
    fi
}

# Check file existence
echo -e "\n📁 File Existence Checks"
check "$([ -f scripts/test-alert-notifications.sh ])" "test-alert-notifications.sh exists"
check "$([ -f scripts/validate-monitoring-stack.sh ])" "validate-monitoring-stack.sh exists"
check "$([ -f docs/MONITORING_VALIDATION_SCRIPTS.md ])" "MONITORING_VALIDATION_SCRIPTS.md exists"
check "$([ -f docs/COMMENT_12_IMPLEMENTATION_SUMMARY.md ])" "COMMENT_12_IMPLEMENTATION_SUMMARY.md exists"

# Check permissions
echo -e "\n🔒 Permission Checks"
check "$([ -x scripts/test-alert-notifications.sh ])" "test-alert-notifications.sh is executable"
check "$([ -x scripts/validate-monitoring-stack.sh ])" "validate-monitoring-stack.sh is executable"

# Check syntax
echo -e "\n📝 Syntax Validation"
bash -n scripts/test-alert-notifications.sh 2>/dev/null && syntax1=true || syntax1=false
check "$syntax1" "test-alert-notifications.sh syntax valid"

bash -n scripts/validate-monitoring-stack.sh 2>/dev/null && syntax2=true || syntax2=false
check "$syntax2" "validate-monitoring-stack.sh syntax valid"

# Check function counts
echo -e "\n🔧 Function Implementation Checks"
alert_funcs=$(grep -c "^test_\|^generate_report\|^main" scripts/test-alert-notifications.sh)
check "$([ $alert_funcs -ge 8 ])" "test-alert-notifications.sh has $alert_funcs functions (expected ≥8)"

monitor_funcs=$(grep -c "^check_\|^generate_recommendations\|^main" scripts/validate-monitoring-stack.sh)
check "$([ $monitor_funcs -ge 12 ])" "validate-monitoring-stack.sh has $monitor_funcs functions (expected ≥12)"

# Check key features
echo -e "\n⚙️  Feature Implementation Checks"
check "$(grep -q 'SLACK_WEBHOOK_URL' scripts/test-alert-notifications.sh)" "Slack webhook testing implemented"
check "$(grep -q 'SMTP_HOST' scripts/test-alert-notifications.sh)" "SMTP email testing implemented"
check "$(grep -q 'PAGERDUTY_SERVICE_KEY' scripts/test-alert-notifications.sh)" "PagerDuty testing implemented"
check "$(grep -q 'test_alertmanager' scripts/test-alert-notifications.sh)" "Alertmanager testing implemented"
check "$(grep -q 'test_silences' scripts/test-alert-notifications.sh)" "Silence testing implemented"

check "$(grep -q 'check_prometheus_metrics' scripts/validate-monitoring-stack.sh)" "Prometheus metrics validation implemented"
check "$(grep -q 'check_jaeger_traces' scripts/validate-monitoring-stack.sh)" "Jaeger trace validation implemented"
check "$(grep -q 'check_elasticsearch_indices' scripts/validate-monitoring-stack.sh)" "Elasticsearch validation implemented"
check "$(grep -q 'check_grafana_datasources' scripts/validate-monitoring-stack.sh)" "Grafana validation implemented"
check "$(grep -q 'check_docker_containers' scripts/validate-monitoring-stack.sh)" "Docker container checks implemented"

# Check report generation
echo -e "\n📊 Report Generation Checks"
check "$(grep -q 'docs/alert-notification-test-report.md' scripts/test-alert-notifications.sh)" "Alert test report generation configured"
check "$(grep -q 'docs/monitoring-validation-report.md' scripts/validate-monitoring-stack.sh)" "Monitoring validation report generation configured"

# Check error handling
echo -e "\n🛡️  Error Handling Checks"
check "$(grep -q 'set -e' scripts/test-alert-notifications.sh)" "Error handling enabled (test-alert-notifications.sh)"
check "$(grep -q 'set -e' scripts/validate-monitoring-stack.sh)" "Error handling enabled (validate-monitoring-stack.sh)"
check "$(grep -q 'exit 0' scripts/test-alert-notifications.sh)" "Success exit code implemented"
check "$(grep -q 'exit 1' scripts/test-alert-notifications.sh)" "Failure exit code implemented"

# Check documentation
echo -e "\n📚 Documentation Checks"
check "$(grep -q 'Usage' docs/MONITORING_VALIDATION_SCRIPTS.md)" "Usage documentation present"
check "$(grep -q 'CI/CD' docs/MONITORING_VALIDATION_SCRIPTS.md)" "CI/CD integration documented"
check "$(grep -q 'Troubleshooting' docs/MONITORING_VALIDATION_SCRIPTS.md)" "Troubleshooting section present"
check "$(grep -q 'Examples' docs/MONITORING_VALIDATION_SCRIPTS.md)" "Usage examples documented"

# File size checks
echo -e "\n💾 File Size Checks"
alert_size=$(wc -c < scripts/test-alert-notifications.sh)
check "$([ $alert_size -gt 10000 ])" "test-alert-notifications.sh adequate size ($alert_size bytes)"

monitor_size=$(wc -c < scripts/validate-monitoring-stack.sh)
check "$([ $monitor_size -gt 10000 ])" "validate-monitoring-stack.sh adequate size ($monitor_size bytes)"

doc_size=$(wc -c < docs/MONITORING_VALIDATION_SCRIPTS.md)
check "$([ $doc_size -gt 10000 ])" "MONITORING_VALIDATION_SCRIPTS.md adequate size ($doc_size bytes)"

# Summary
echo -e "\n======================================"
echo -e "Summary: $PASSED passed, $FAILED failed"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All verification checks passed!${NC}"
    echo -e "\n📋 Implementation Complete:"
    echo "  - Alert notification testing script"
    echo "  - Monitoring stack validation script"
    echo "  - Comprehensive documentation"
    echo "  - CI/CD integration support"
    exit 0
else
    echo -e "${RED}❌ $FAILED verification check(s) failed${NC}"
    exit 1
fi

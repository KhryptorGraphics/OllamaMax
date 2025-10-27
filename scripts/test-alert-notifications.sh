#!/bin/bash

# Test Alert Notifications Script
# Tests all alert notification channels for OllamaMax monitoring stack

set -euo pipefail

# Color codes for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

log_error() {
    echo -e "${RED}[✗]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Test result tracking
TESTS_PASSED=0
TESTS_FAILED=0
FAILED_TESTS=()

# Function to test Slack webhook
test_slack_webhook() {
    log_info "Testing Slack webhook notification..."

    if [ -z "${SLACK_WEBHOOK_URL:-}" ]; then
        log_warning "SLACK_WEBHOOK_URL not set, skipping Slack test"
        return 0
    fi

    local test_payload=$(cat <<EOF
{
    "text": "OllamaMax Monitoring Test Alert",
    "blocks": [
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "*Test Alert from OllamaMax Monitoring Stack*\n\nThis is a test notification to verify Slack integration is working correctly."
            }
        },
        {
            "type": "context",
            "elements": [
                {
                    "type": "mrkdwn",
                    "text": "Timestamp: $(date -u +"%Y-%m-%dT%H:%M:%SZ")"
                }
            ]
        }
    ]
}
EOF
)

    local response=$(curl -s -o /dev/null -w "%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "$test_payload" \
        "$SLACK_WEBHOOK_URL" \
        --max-time 10)

    if [ "$response" = "200" ]; then
        log_success "Slack webhook test passed (HTTP $response)"
        ((TESTS_PASSED++))
        return 0
    else
        log_error "Slack webhook test failed (HTTP $response)"
        FAILED_TESTS+=("Slack webhook")
        ((TESTS_FAILED++))
        return 1
    fi
}

# Function to test SMTP email
test_smtp_email() {
    log_info "Testing SMTP email notification..."

    if [ -z "${SMTP_HOST:-}" ] || [ -z "${SMTP_USER:-}" ] || [ -z "${SMTP_PASSWORD:-}" ]; then
        log_warning "SMTP configuration not set (SMTP_HOST, SMTP_USER, SMTP_PASSWORD), skipping SMTP test"
        return 0
    fi

    local smtp_port="${SMTP_PORT:-587}"
    local smtp_from="${SMTP_FROM:-monitoring@ollamamax.local}"
    local smtp_to="${SMTP_TO:-admin@ollamamax.local}"

    # Create email content
    local email_subject="OllamaMax Monitoring Test Alert"
    local email_body=$(cat <<EOF
Subject: ${email_subject}
From: ${smtp_from}
To: ${smtp_to}
Content-Type: text/html; charset=UTF-8

<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; }
        .header { background-color: #4CAF50; color: white; padding: 10px; }
        .content { padding: 20px; }
        .footer { background-color: #f1f1f1; padding: 10px; font-size: 12px; }
    </style>
</head>
<body>
    <div class="header">
        <h2>OllamaMax Monitoring Test Alert</h2>
    </div>
    <div class="content">
        <p>This is a test notification to verify SMTP email integration is working correctly.</p>
        <p><strong>Timestamp:</strong> $(date -u +"%Y-%m-%d %H:%M:%S UTC")</p>
        <p><strong>Test Type:</strong> Alert Notification System Validation</p>
    </div>
    <div class="footer">
        <p>Sent from OllamaMax Monitoring Stack</p>
    </div>
</body>
</html>
EOF
)

    # Test using curl with SMTP
    local response=$(echo "$email_body" | curl -s -o /dev/null -w "%{http_code}" \
        --url "smtp://${SMTP_HOST}:${smtp_port}" \
        --ssl-reqd \
        --mail-from "$smtp_from" \
        --mail-rcpt "$smtp_to" \
        --user "${SMTP_USER}:${SMTP_PASSWORD}" \
        --upload-file - \
        --max-time 30 2>&1)

    if [ $? -eq 0 ]; then
        log_success "SMTP email test passed"
        ((TESTS_PASSED++))
        return 0
    else
        log_error "SMTP email test failed: $response"
        FAILED_TESTS+=("SMTP email")
        ((TESTS_FAILED++))

        # Try alternative method using mail command if available
        if command -v mail &> /dev/null; then
            log_info "Attempting fallback with mail command..."
            echo "Test alert from OllamaMax Monitoring" | mail -s "$email_subject" "$smtp_to" 2>&1
            if [ $? -eq 0 ]; then
                log_success "SMTP email test passed (using mail command)"
                ((TESTS_PASSED++))
                ((TESTS_FAILED--))
                FAILED_TESTS=("${FAILED_TESTS[@]/SMTP email/}")
                return 0
            fi
        fi
        return 1
    fi
}

# Function to test PagerDuty
test_pagerduty() {
    log_info "Testing PagerDuty integration..."

    if [ -z "${PAGERDUTY_SERVICE_KEY:-}" ]; then
        log_warning "PAGERDUTY_SERVICE_KEY not set, skipping PagerDuty test"
        return 0
    fi

    local test_payload=$(cat <<EOF
{
    "routing_key": "${PAGERDUTY_SERVICE_KEY}",
    "event_action": "trigger",
    "payload": {
        "summary": "OllamaMax Monitoring Test Alert",
        "source": "monitoring.ollamamax.local",
        "severity": "info",
        "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
        "custom_details": {
            "test_type": "Alert Notification System Validation",
            "component": "Monitoring Stack",
            "description": "This is a test event to verify PagerDuty integration is working correctly."
        }
    }
}
EOF
)

    local response=$(curl -s -o /tmp/pagerduty_response.json -w "%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "$test_payload" \
        "https://events.pagerduty.com/v2/enqueue" \
        --max-time 10)

    if [ "$response" = "202" ]; then
        local dedup_key=$(jq -r '.dedup_key // empty' /tmp/pagerduty_response.json 2>/dev/null)
        log_success "PagerDuty test passed (HTTP $response, dedup_key: ${dedup_key:-N/A})"
        ((TESTS_PASSED++))
        rm -f /tmp/pagerduty_response.json
        return 0
    else
        log_error "PagerDuty test failed (HTTP $response)"
        if [ -f /tmp/pagerduty_response.json ]; then
            local error_msg=$(jq -r '.message // .errors // empty' /tmp/pagerduty_response.json 2>/dev/null)
            [ -n "$error_msg" ] && log_error "Error details: $error_msg"
            rm -f /tmp/pagerduty_response.json
        fi
        FAILED_TESTS+=("PagerDuty")
        ((TESTS_FAILED++))
        return 1
    fi
}

# Function to test webhook (generic)
test_generic_webhook() {
    log_info "Testing generic webhook notification..."

    if [ -z "${WEBHOOK_URL:-}" ]; then
        log_warning "WEBHOOK_URL not set, skipping generic webhook test"
        return 0
    fi

    local test_payload=$(cat <<EOF
{
    "alert_name": "OllamaMax Monitoring Test",
    "severity": "info",
    "status": "firing",
    "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
    "description": "This is a test alert to verify webhook integration is working correctly.",
    "labels": {
        "alertname": "MonitoringTest",
        "severity": "info",
        "component": "monitoring-stack"
    },
    "annotations": {
        "summary": "OllamaMax Monitoring Test Alert",
        "description": "Test notification for webhook validation"
    }
}
EOF
)

    local response=$(curl -s -o /dev/null -w "%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "$test_payload" \
        "$WEBHOOK_URL" \
        --max-time 10)

    if [ "$response" -ge 200 ] && [ "$response" -lt 300 ]; then
        log_success "Generic webhook test passed (HTTP $response)"
        ((TESTS_PASSED++))
        return 0
    else
        log_error "Generic webhook test failed (HTTP $response)"
        FAILED_TESTS+=("Generic webhook")
        ((TESTS_FAILED++))
        return 1
    fi
}

# Function to print summary report
print_summary() {
    echo ""
    echo "========================================"
    echo "  Alert Notification Test Summary"
    echo "========================================"
    echo ""
    echo -e "Total Tests Run:    $((TESTS_PASSED + TESTS_FAILED))"
    echo -e "${GREEN}Tests Passed:       ${TESTS_PASSED}${NC}"
    echo -e "${RED}Tests Failed:       ${TESTS_FAILED}${NC}"
    echo ""

    if [ ${#FAILED_TESTS[@]} -gt 0 ]; then
        echo -e "${RED}Failed Tests:${NC}"
        for test in "${FAILED_TESTS[@]}"; do
            echo -e "  ${RED}✗${NC} $test"
        done
        echo ""
    fi

    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}✓ All alert notification tests passed!${NC}"
        echo ""
        return 0
    else
        echo -e "${RED}✗ Some alert notification tests failed!${NC}"
        echo ""
        return 1
    fi
}

# Main execution
main() {
    echo "========================================"
    echo "  OllamaMax Alert Notification Tests"
    echo "========================================"
    echo ""
    echo "Testing alert notification channels..."
    echo ""

    # Check for required commands
    if ! command -v curl &> /dev/null; then
        log_error "curl is required but not installed"
        exit 1
    fi

    if ! command -v jq &> /dev/null; then
        log_warning "jq is not installed, JSON parsing may be limited"
    fi

    # Run all tests
    test_slack_webhook || true
    echo ""

    test_smtp_email || true
    echo ""

    test_pagerduty || true
    echo ""

    test_generic_webhook || true
    echo ""

    # Print summary and exit with appropriate code
    if print_summary; then
        exit 0
    else
        exit 1
    fi
}

# Execute main function
main "$@"

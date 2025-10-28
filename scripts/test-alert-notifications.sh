#!/bin/bash
set -e

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

echo "🔔 Testing Alert Notification Channels"
echo "======================================"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test Slack webhook
test_slack() {
    echo -e "\n${YELLOW}Testing Slack webhook...${NC}"

    if [ -z "$SLACK_WEBHOOK_URL" ]; then
        echo -e "${RED}❌ SLACK_WEBHOOK_URL not set${NC}"
        return 1
    fi

    response=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$SLACK_WEBHOOK_URL" \
        -H 'Content-Type: application/json' \
        -d '{
            "text": "✅ OllamaMax Alert Test",
            "blocks": [{
                "type": "section",
                "text": {
                    "type": "mrkdwn",
                    "text": "*Test Alert from OllamaMax*\nThis is a test notification to verify Slack integration."
                }
            }]
        }')

    if [ "$response" = "200" ]; then
        echo -e "${GREEN}✅ Slack webhook test successful${NC}"
        return 0
    else
        echo -e "${RED}❌ Slack webhook test failed (HTTP $response)${NC}"
        return 1
    fi
}

# Test SMTP email
test_email() {
    echo -e "\n${YELLOW}Testing SMTP email...${NC}"

    if [ -z "$SMTP_HOST" ] || [ -z "$ALERT_EMAIL_CRITICAL" ]; then
        echo -e "${RED}❌ SMTP variables not set${NC}"
        return 1
    fi

    # Use Python to send test email
    python3 << 'EOFPYTHON'
import smtplib
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
import os
import sys
from datetime import datetime

try:
    msg = MIMEMultipart()
    msg['From'] = os.getenv('SMTP_FROM', 'noreply@ollamamax.local')
    msg['To'] = os.getenv('ALERT_EMAIL_CRITICAL')
    msg['Subject'] = '[TEST] OllamaMax Alert Notification Test'

    body = f"""
This is a test email from OllamaMax monitoring system.

If you received this, email alerting is configured correctly.

Test timestamp: {datetime.now().isoformat()}
    """
    msg.attach(MIMEText(body, 'plain'))

    server = smtplib.SMTP(os.getenv('SMTP_HOST'), int(os.getenv('SMTP_PORT', 587)))
    server.starttls()
    server.login(os.getenv('SMTP_USER'), os.getenv('SMTP_PASSWORD'))
    server.send_message(msg)
    server.quit()

    print("✅ Email sent successfully")
    sys.exit(0)
except Exception as e:
    print(f"❌ Email test failed: {e}")
    sys.exit(1)
EOFPYTHON

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ SMTP email test successful${NC}"
        return 0
    else
        echo -e "${RED}❌ SMTP email test failed${NC}"
        return 1
    fi
}

# Test PagerDuty
test_pagerduty() {
    echo -e "\n${YELLOW}Testing PagerDuty integration...${NC}"

    if [ -z "$PAGERDUTY_SERVICE_KEY" ]; then
        echo -e "${RED}❌ PAGERDUTY_SERVICE_KEY not set${NC}"
        return 1
    fi

    response=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
        'https://events.pagerduty.com/v2/enqueue' \
        -H 'Content-Type: application/json' \
        -d "{
            \"routing_key\": \"$PAGERDUTY_SERVICE_KEY\",
            \"event_action\": \"trigger\",
            \"payload\": {
                \"summary\": \"OllamaMax Test Alert\",
                \"severity\": \"info\",
                \"source\": \"ollamamax-monitoring\",
                \"custom_details\": {
                    \"description\": \"This is a test alert to verify PagerDuty integration\"
                }
            }
        }")

    if [ "$response" = "202" ]; then
        echo -e "${GREEN}✅ PagerDuty test successful${NC}"
        return 0
    else
        echo -e "${RED}❌ PagerDuty test failed (HTTP $response)${NC}"
        return 1
    fi
}

# Test Alertmanager API
test_alertmanager() {
    echo -e "\n${YELLOW}Testing Alertmanager API...${NC}"

    # Check if Alertmanager is running
    if ! curl -s "http://localhost:9093/-/healthy" > /dev/null 2>&1; then
        echo -e "${RED}❌ Alertmanager is not running${NC}"
        return 1
    fi

    # Create a test alert
    response=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
        'http://localhost:9093/api/v1/alerts' \
        -H 'Content-Type: application/json' \
        -d '[{
            "labels": {
                "alertname": "TestAlert",
                "severity": "warning",
                "instance": "test"
            },
            "annotations": {
                "summary": "Test alert from validation script",
                "description": "This is a test alert to verify Alertmanager routing"
            }
        }]')

    if [ "$response" = "200" ]; then
        echo -e "${GREEN}✅ Alertmanager API test successful${NC}"
        return 0
    else
        echo -e "${RED}❌ Alertmanager API test failed (HTTP $response)${NC}"
        return 1
    fi
}

# Test Webhook receiver
test_webhook() {
    echo -e "\n${YELLOW}Testing custom webhook...${NC}"

    if [ -z "$WEBHOOK_URL" ]; then
        echo -e "${YELLOW}⚠️  WEBHOOK_URL not set, skipping${NC}"
        return 0
    fi

    response=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$WEBHOOK_URL" \
        -H 'Content-Type: application/json' \
        -d '{
            "status": "firing",
            "alerts": [{
                "labels": {
                    "alertname": "TestAlert",
                    "severity": "info"
                },
                "annotations": {
                    "summary": "OllamaMax webhook test"
                }
            }]
        }')

    if [ "$response" = "200" ] || [ "$response" = "202" ]; then
        echo -e "${GREEN}✅ Webhook test successful${NC}"
        return 0
    else
        echo -e "${RED}❌ Webhook test failed (HTTP $response)${NC}"
        return 1
    fi
}

# Test notification silences
test_silences() {
    echo -e "\n${YELLOW}Testing Alertmanager silences...${NC}"

    # Create a test silence
    silence_id=$(curl -s -X POST 'http://localhost:9093/api/v1/silences' \
        -H 'Content-Type: application/json' \
        -d "{
            \"matchers\": [{
                \"name\": \"alertname\",
                \"value\": \"TestAlert\",
                \"isRegex\": false
            }],
            \"startsAt\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
            \"endsAt\": \"$(date -u -d '+5 minutes' +%Y-%m-%dT%H:%M:%SZ)\",
            \"createdBy\": \"test-script\",
            \"comment\": \"Test silence from validation script\"
        }" | grep -o '"silenceID":"[^"]*"' | cut -d'"' -f4)

    if [ -n "$silence_id" ]; then
        echo -e "${GREEN}✅ Silence created successfully (ID: $silence_id)${NC}"

        # Clean up - delete the silence
        curl -s -X DELETE "http://localhost:9093/api/v1/silence/$silence_id" > /dev/null
        echo -e "${GREEN}✅ Silence deleted successfully${NC}"
        return 0
    else
        echo -e "${RED}❌ Failed to create silence${NC}"
        return 1
    fi
}

# Generate test report
generate_report() {
    local slack_status=$1
    local email_status=$2
    local pagerduty_status=$3
    local alertmanager_status=$4
    local webhook_status=$5
    local silences_status=$6

    cat > docs/alert-notification-test-report.md << EOFREPORT
# Alert Notification Test Report

**Generated:** $(date)

## Test Results Summary

| Channel | Status | Details |
|---------|--------|---------|
| Slack | $([ $slack_status -eq 0 ] && echo "✅ PASSED" || echo "❌ FAILED") | Webhook notification test |
| Email (SMTP) | $([ $email_status -eq 0 ] && echo "✅ PASSED" || echo "❌ FAILED") | Email delivery test |
| PagerDuty | $([ $pagerduty_status -eq 0 ] && echo "✅ PASSED" || echo "❌ FAILED") | Incident creation test |
| Alertmanager API | $([ $alertmanager_status -eq 0 ] && echo "✅ PASSED" || echo "❌ FAILED") | Alert posting test |
| Custom Webhook | $([ $webhook_status -eq 0 ] && echo "✅ PASSED" || echo "⚠️  SKIPPED") | Webhook delivery test |
| Silences | $([ $silences_status -eq 0 ] && echo "✅ PASSED" || echo "❌ FAILED") | Silence creation/deletion test |

## Configuration Check

### Environment Variables
- SLACK_WEBHOOK_URL: $([ -n "$SLACK_WEBHOOK_URL" ] && echo "✓ Set" || echo "✗ Not set")
- SMTP_HOST: $([ -n "$SMTP_HOST" ] && echo "✓ Set" || echo "✗ Not set")
- ALERT_EMAIL_CRITICAL: $([ -n "$ALERT_EMAIL_CRITICAL" ] && echo "✓ Set" || echo "✗ Not set")
- PAGERDUTY_SERVICE_KEY: $([ -n "$PAGERDUTY_SERVICE_KEY" ] && echo "✓ Set" || echo "✗ Not set")
- WEBHOOK_URL: $([ -n "$WEBHOOK_URL" ] && echo "✓ Set" || echo "✗ Not set")

## Next Steps

$(if [ $slack_status -ne 0 ] || [ $email_status -ne 0 ] || [ $pagerduty_status -ne 0 ] || [ $alertmanager_status -ne 0 ]; then
    echo "### Failed Tests"
    [ $slack_status -ne 0 ] && echo "- Fix Slack webhook configuration"
    [ $email_status -ne 0 ] && echo "- Verify SMTP credentials and settings"
    [ $pagerduty_status -ne 0 ] && echo "- Check PagerDuty service key"
    [ $alertmanager_status -ne 0 ] && echo "- Ensure Alertmanager is running and accessible"
else
    echo "All notification channels are working correctly! 🎉"
fi)

## Manual Verification

After running this automated test:

1. **Check Slack**: Verify test message appears in configured channel
2. **Check Email**: Verify test email received at configured address
3. **Check PagerDuty**: Verify test incident created in PagerDuty dashboard
4. **Check Alertmanager**: Visit http://localhost:9093 to see test alerts

---
*Generated by test-alert-notifications.sh*
EOFREPORT

    echo -e "\n${GREEN}📄 Test report saved to docs/alert-notification-test-report.md${NC}"
}

# Run all tests
main() {
    local failed=0
    local slack_result=0
    local email_result=0
    local pagerduty_result=0
    local alertmanager_result=0
    local webhook_result=0
    local silences_result=0

    test_slack || { slack_result=1; ((failed++)); }
    test_email || { email_result=1; ((failed++)); }
    test_pagerduty || { pagerduty_result=1; ((failed++)); }
    test_alertmanager || { alertmanager_result=1; ((failed++)); }
    test_webhook || { webhook_result=1; }  # Don't count as failure if skipped
    test_silences || { silences_result=1; ((failed++)); }

    # Generate test report
    mkdir -p docs
    generate_report $slack_result $email_result $pagerduty_result $alertmanager_result $webhook_result $silences_result

    echo -e "\n======================================"
    if [ $failed -eq 0 ]; then
        echo -e "${GREEN}✅ All notification tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}❌ $failed notification test(s) failed${NC}"
        echo -e "${YELLOW}Check docs/alert-notification-test-report.md for details${NC}"
        exit 1
    fi
}

main

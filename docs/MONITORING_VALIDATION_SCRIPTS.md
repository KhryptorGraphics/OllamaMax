# Monitoring Validation Scripts Documentation

## Overview

Two comprehensive validation scripts have been created to test and validate the OllamaMax monitoring stack:

1. **test-alert-notifications.sh** - Tests all alert notification channels
2. **validate-monitoring-stack.sh** - Validates entire monitoring infrastructure

## Script 1: test-alert-notifications.sh

### Purpose
Validates that all configured alert notification channels are working correctly.

### Location
`/home/kp/OllamaMax/scripts/test-alert-notifications.sh`

### Features

#### Test Coverage
- **Slack Webhook**: Tests Slack integration with formatted message blocks
- **SMTP Email**: Tests email delivery using Python smtplib
- **PagerDuty**: Tests incident creation via PagerDuty Events API
- **Alertmanager API**: Tests alert posting to Alertmanager
- **Custom Webhook**: Tests generic webhook receivers (optional)
- **Silences**: Tests Alertmanager silence creation and deletion

#### Environment Variables Required
```bash
# Slack
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL

# Email
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=alerts@example.com
SMTP_PASSWORD=your-password
SMTP_FROM=noreply@ollamamax.local
ALERT_EMAIL_CRITICAL=critical@example.com

# PagerDuty
PAGERDUTY_SERVICE_KEY=your-service-key

# Custom Webhook (optional)
WEBHOOK_URL=https://your-webhook-receiver.com/alerts
```

### Usage

```bash
# Basic usage
./scripts/test-alert-notifications.sh

# Run with custom .env file
env $(cat custom.env | xargs) ./scripts/test-alert-notifications.sh
```

### Output

#### Console Output
- Color-coded results (green for pass, red for fail, yellow for warnings)
- Real-time test progress
- Summary of passed/failed tests

#### Generated Report
- Location: `docs/alert-notification-test-report.md`
- Contains:
  - Test results summary table
  - Configuration status
  - Failed test remediation steps
  - Manual verification checklist

### Exit Codes
- `0`: All tests passed
- `1`: One or more tests failed

### Example Output

```
🔔 Testing Alert Notification Channels
======================================

Testing Slack webhook...
✅ Slack webhook test successful

Testing SMTP email...
✅ SMTP email test successful

Testing PagerDuty integration...
✅ PagerDuty test successful

Testing Alertmanager API...
✅ Alertmanager API test successful

Testing custom webhook...
⚠️  WEBHOOK_URL not set, skipping

Testing Alertmanager silences...
✅ Silence created successfully (ID: abc123)
✅ Silence deleted successfully

📄 Test report saved to docs/alert-notification-test-report.md

======================================
✅ All notification tests passed!
```

## Script 2: validate-monitoring-stack.sh

### Purpose
Comprehensive validation of the entire monitoring infrastructure including services, metrics, traces, and configurations.

### Location
`/home/kp/OllamaMax/scripts/validate-monitoring-stack.sh`

### Features

#### Validation Coverage

**1. Core Services Health**
- Prometheus (port 9090)
- Grafana (port 3001)
- Alertmanager (port 9093)
- Jaeger UI (port 16686)
- Elasticsearch (port 9200)
- Kibana (port 5601)

**2. Metrics Validation**
- `ollamamax_api_http_requests_total`
- `ollamamax_database_db_connections_open`
- `ollamamax_p2p_connected_peers`
- `ollamamax_loadbalancer_requests_total`

**3. Distributed Tracing**
- Jaeger service discovery
- Trace availability for ollamamax-api

**4. Log Aggregation**
- Elasticsearch indices for OllamaMax logs
- Index health and size

**5. Alert Rules**
- Prometheus rule loading
- Alert configuration validation

**6. Grafana Configuration**
- Datasource validation (Prometheus, Jaeger)
- Dashboard availability

**7. Infrastructure**
- Docker container status
- Disk space monitoring
- Data retention settings

### Usage

```bash
# Basic usage
./scripts/validate-monitoring-stack.sh

# Run as part of CI/CD
./scripts/validate-monitoring-stack.sh || exit 1
```

### Output

#### Console Output
```
🔍 Validating OllamaMax Monitoring Stack
========================================

Checking Prometheus...
✅ Prometheus is healthy

Checking Grafana...
✅ Grafana is healthy

Checking Prometheus metrics...
✅ Metric found: ollamamax_api_http_requests_total
✅ Metric found: ollamamax_database_db_connections_open
...

======================================
✅ All monitoring components validated successfully!
📄 Full report saved to docs/monitoring-validation-report.md
```

#### Generated Report
- Location: `docs/monitoring-validation-report.md`
- Contains:
  - Component status checklist
  - Metrics availability
  - Trace validation results
  - Configuration checks
  - Disk space analysis
  - Actionable recommendations

### Exit Codes
- `0`: All validations passed
- `1`: One or more validations failed

### Report Example

```markdown
# OllamaMax Monitoring Stack Validation Report
Generated: Mon Oct 27 17:56:23 UTC 2025

## Component Status

- [x] Prometheus (HTTP 200)
- [x] Grafana (HTTP 200)
- [x] Alertmanager (HTTP 200)
- [x] Jaeger UI (HTTP 200)
- [x] Elasticsearch (HTTP 200)
- [x] Kibana (HTTP 200)

### Prometheus Metrics
- [x] ollamamax_api_http_requests_total
- [x] ollamamax_database_db_connections_open
- [x] ollamamax_p2p_connected_peers
- [x] ollamamax_loadbalancer_requests_total

### Jaeger Tracing
- [x] Service traces found

### Alert Rules
- [x] Rules loaded successfully

### Grafana Datasources
- [x] Prometheus datasource
- [x] Jaeger datasource

### Docker Containers
- [x] prometheus (running)
- [x] grafana (running)
- [x] alertmanager (running)
- [x] jaeger (running)
- [x] elasticsearch (running)
- [x] kibana (running)

### Disk Space
- Available: 50G
- [x] Disk usage: 45%

### Data Retention
- Prometheus: 15d
- Elasticsearch: 2.3gb

### Recommendations

All monitoring components are functioning correctly. Consider these optimization steps:

1. **Performance Tuning**: Review dashboard queries for optimization opportunities
2. **Alert Tuning**: Adjust alert thresholds based on observed baseline metrics
3. **Retention Policies**: Verify data retention matches your compliance requirements
4. **Capacity Planning**: Monitor disk usage trends and plan for growth
5. **Dashboard Enhancement**: Add custom dashboards for application-specific metrics

## Summary

Validation: **PASSED** ✅
```

## Integration with CI/CD

### GitHub Actions Example

```yaml
name: Validate Monitoring Stack

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  schedule:
    - cron: '0 */6 * * *'  # Every 6 hours

jobs:
  validate-monitoring:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Start monitoring stack
        run: docker-compose up -d
      
      - name: Wait for services to be ready
        run: sleep 30
      
      - name: Validate monitoring stack
        run: ./scripts/validate-monitoring-stack.sh
      
      - name: Test alert notifications
        env:
          SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
          SMTP_HOST: ${{ secrets.SMTP_HOST }}
          SMTP_USER: ${{ secrets.SMTP_USER }}
          SMTP_PASSWORD: ${{ secrets.SMTP_PASSWORD }}
          ALERT_EMAIL_CRITICAL: ${{ secrets.ALERT_EMAIL }}
          PAGERDUTY_SERVICE_KEY: ${{ secrets.PAGERDUTY_KEY }}
        run: ./scripts/test-alert-notifications.sh
      
      - name: Upload validation reports
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: monitoring-reports
          path: |
            docs/monitoring-validation-report.md
            docs/alert-notification-test-report.md
```

### Docker Compose Integration

```yaml
# Add validation service
services:
  monitoring-validator:
    image: alpine:latest
    volumes:
      - ./scripts:/scripts
      - ./docs:/docs
    command: |
      sh -c "
        apk add --no-cache curl python3 py3-pip bash
        sleep 60  # Wait for services
        /scripts/validate-monitoring-stack.sh
        /scripts/test-alert-notifications.sh
      "
    depends_on:
      - prometheus
      - grafana
      - alertmanager
```

## Troubleshooting

### Common Issues

#### 1. Connection Refused Errors
```bash
# Check if services are running
docker-compose ps

# Check service logs
docker-compose logs prometheus
docker-compose logs grafana
```

#### 2. Missing Environment Variables
```bash
# Verify .env file exists
cat .env | grep -E "SLACK_|SMTP_|PAGERDUTY_"

# Export variables manually
export SLACK_WEBHOOK_URL="https://..."
```

#### 3. Permission Denied
```bash
# Ensure scripts are executable
chmod +x scripts/test-alert-notifications.sh
chmod +x scripts/validate-monitoring-stack.sh
```

#### 4. Python Not Found
```bash
# Install Python 3
sudo apt-get install python3 python3-pip  # Ubuntu/Debian
sudo yum install python3 python3-pip      # CentOS/RHEL
```

#### 5. Docker Not Available
```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Start Docker
sudo systemctl start docker
```

## Maintenance

### Update Notification Channels

To add a new notification channel:

1. Add test function to `test-alert-notifications.sh`:
```bash
test_new_channel() {
    echo -e "\n${YELLOW}Testing New Channel...${NC}"
    # Implementation
}
```

2. Update main function to call new test
3. Add to report generation
4. Update environment variables documentation

### Update Validation Checks

To add new validation checks:

1. Add check function to `validate-monitoring-stack.sh`:
```bash
check_new_component() {
    echo -e "\n${YELLOW}Checking New Component...${NC}"
    # Implementation
}
```

2. Call from main function
3. Update report generation
4. Document in this file

## Best Practices

1. **Run Regularly**: Schedule validation runs every 6 hours
2. **Monitor Reports**: Review generated reports in docs/
3. **Update Variables**: Keep environment variables current
4. **Version Control**: Commit report templates and configurations
5. **Alert on Failure**: Configure CI/CD to alert on validation failures
6. **Documentation**: Keep this guide updated with changes

## Related Documentation

- [Monitoring Implementation Guide](./MONITORING_IMPLEMENTATION_GUIDE.md)
- [Prometheus Integration](./PROMETHEUS_INTEGRATION.md)
- [Deployment Validation Report](./DEPLOYMENT_VALIDATION_REPORT.md)
- [Troubleshooting Guide](./TROUBLESHOOTING_GUIDE.md)

## Support

For issues or questions:
1. Check troubleshooting section above
2. Review generated reports for specific error messages
3. Check service logs: `docker-compose logs [service-name]`
4. Verify environment variables are set correctly

---
*Last Updated: 2025-10-27*
*Scripts Version: 1.0.0*

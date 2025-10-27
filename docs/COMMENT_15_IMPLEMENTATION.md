# Comment 15 Implementation: Monitoring Validation and Alert Notification Tests

## Implementation Summary

Successfully implemented comprehensive monitoring validation and alert notification test scripts for the OllamaMax monitoring stack.

## Files Created

### 1. `/scripts/test-alert-notifications.sh`
**Purpose**: Test all alert notification channels to ensure proper integration

**Features**:
- **Slack Webhook Testing**
  - Sends formatted test payload with blocks and context
  - Validates HTTP 200 response
  - Configurable via `SLACK_WEBHOOK_URL` environment variable

- **SMTP Email Testing**
  - Sends HTML-formatted test email via curl
  - Supports SMTP authentication with SSL/TLS
  - Fallback to `mail` command if available
  - Configuration via `SMTP_HOST`, `SMTP_USER`, `SMTP_PASSWORD`, `SMTP_PORT`, `SMTP_FROM`, `SMTP_TO`

- **PagerDuty Testing**
  - Uses PagerDuty Events API v2
  - Sends test trigger event
  - Validates 202 response and captures dedup_key
  - Configuration via `PAGERDUTY_SERVICE_KEY`

- **Generic Webhook Testing**
  - Supports custom webhook endpoints
  - Sends standardized alert payload
  - Configuration via `WEBHOOK_URL`

**Output Features**:
- Color-coded status indicators (✓ green for success, ✗ red for failure)
- Detailed logging with INFO, WARNING, ERROR levels
- Comprehensive summary report with pass/fail counts
- Lists failed tests for easy debugging
- Exit code 0 for success, 1 for any failures

**Usage**:
```bash
# Set environment variables
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
export SMTP_HOST="smtp.gmail.com"
export SMTP_USER="your-email@gmail.com"
export SMTP_PASSWORD="your-app-password"
export PAGERDUTY_SERVICE_KEY="your-service-key"

# Run tests
./scripts/test-alert-notifications.sh
```

### 2. `/scripts/validate-monitoring-stack.sh`
**Purpose**: Comprehensive validation of all monitoring components

**Components Validated**:

1. **Prometheus**
   - Health endpoint: `/-/healthy`
   - Ready status: `/-/ready`
   - Metrics queries for key metrics:
     - `http_requests_total`
     - `db_connections_open`
     - `p2p_connected_peers`
     - `lb_requests_total`

2. **Grafana**
   - Health API: `/api/health`
   - Metrics endpoint: `/metrics`

3. **Alertmanager**
   - Health endpoint: `/-/healthy`
   - Ready status: `/-/ready`
   - Active alerts count via API v2

4. **Jaeger**
   - UI endpoint health check
   - Traces verification for service
   - Services list retrieval

5. **Elasticsearch**
   - Cluster health: `/_cluster/health`
   - Node stats: `/_nodes/stats`
   - Log document count in index pattern
   - Index health verification

6. **Logstash**
   - Node stats: `/_node/stats`
   - Pipeline status detection
   - Node info retrieval

7. **Kibana**
   - Status API: `/api/status`
   - Overall status state verification

**Features**:
- Configurable URLs via environment variables
- Configurable timeouts (5s health, 10s metrics)
- Comprehensive metrics validation
- Trace and log data verification
- Color-coded output with section headers
- Detailed summary report with component URLs
- Box-drawing characters for attractive reporting
- Exit code 0 for all pass, 1 for any failures

**Configuration**:
```bash
# Optional: Override default URLs
export PROMETHEUS_URL="http://localhost:9090"
export GRAFANA_URL="http://localhost:3001"
export ALERTMANAGER_URL="http://localhost:9093"
export JAEGER_URL="http://localhost:16686"
export ELASTICSEARCH_URL="http://localhost:9200"
export LOGSTASH_URL="http://localhost:9600"
export KIBANA_URL="http://localhost:5601"

# Optional: Configure service name for Jaeger traces
export JAEGER_SERVICE_NAME="ollamamax-api"

# Optional: Configure Elasticsearch index pattern
export ELASTICSEARCH_INDEX_PATTERN="ollamamax-logs-*"
```

**Usage**:
```bash
# Run with defaults
./scripts/validate-monitoring-stack.sh

# Run with custom configuration
PROMETHEUS_URL="http://prometheus:9090" \
GRAFANA_URL="http://grafana:3000" \
./scripts/validate-monitoring-stack.sh
```

## Script Architecture

### Error Handling
- Both scripts use `set -euo pipefail` for strict error handling
- Graceful handling of missing environment variables (warnings, not errors)
- Timeout protection on all HTTP requests
- Fallback mechanisms (e.g., mail command for SMTP)

### Logging System
- **Color-coded output**: Green (success), Red (error), Yellow (warning), Blue (info), Cyan (sections)
- **Structured logging**: Consistent format with severity prefixes
- **Summary reporting**: Comprehensive pass/fail statistics
- **Component tracking**: Lists all failed checks for debugging

### Test Tracking
- Counters for passed/failed checks
- Arrays to store failed test names
- Summary report generation
- Exit codes for CI/CD integration

## Integration with CI/CD

Both scripts are designed for CI/CD pipeline integration:

```yaml
# Example GitHub Actions job
monitoring-validation:
  runs-on: ubuntu-latest
  steps:
    - name: Validate Monitoring Stack
      run: ./scripts/validate-monitoring-stack.sh
      env:
        PROMETHEUS_URL: ${{ secrets.PROMETHEUS_URL }}

    - name: Test Alert Notifications
      run: ./scripts/test-alert-notifications.sh
      env:
        SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
        SMTP_HOST: ${{ secrets.SMTP_HOST }}
        SMTP_USER: ${{ secrets.SMTP_USER }}
        SMTP_PASSWORD: ${{ secrets.SMTP_PASSWORD }}
        PAGERDUTY_SERVICE_KEY: ${{ secrets.PAGERDUTY_SERVICE_KEY }}
```

## Security Considerations

1. **Secrets Management**
   - All credentials passed via environment variables
   - No secrets hardcoded in scripts
   - Safe to commit scripts to version control

2. **Timeout Protection**
   - All HTTP requests have timeouts
   - Prevents hanging on unresponsive services

3. **Error Masking**
   - Sensitive error messages not logged
   - Exit codes indicate success/failure without exposing details

## Testing Strategy

### Alert Notification Tests
- **Unit-level**: Each notification channel tested independently
- **Integration-level**: End-to-end webhook delivery
- **Validation**: HTTP response codes and API responses

### Monitoring Stack Validation
- **Health checks**: Basic service availability
- **Functional checks**: API endpoints and data retrieval
- **Data validation**: Metrics, traces, and logs present
- **Performance checks**: Response time validation via timeouts

## Dependencies

### Required
- `curl`: HTTP client for all requests
- `bash`: Shell execution environment

### Optional
- `jq`: JSON parsing (gracefully degrades without it)
- `mail`: SMTP fallback for email testing

## Success Criteria

✅ **Implemented**:
1. Alert notification testing for Slack, SMTP, PagerDuty, generic webhooks
2. Comprehensive monitoring stack validation for all 7 components
3. Health checks for all services
4. Metrics query validation in Prometheus
5. Trace verification in Jaeger
6. Log verification in Elasticsearch
7. Color-coded output with clear status indicators
8. Detailed summary reports
9. Proper error handling and exit codes
10. Executable permissions set on both scripts

## Example Output

### test-alert-notifications.sh
```
========================================
  OllamaMax Alert Notification Tests
========================================

Testing alert notification channels...

[INFO] Testing Slack webhook notification...
[✓] Slack webhook test passed (HTTP 200)

[INFO] Testing SMTP email notification...
[✓] SMTP email test passed

[INFO] Testing PagerDuty integration...
[✓] PagerDuty test passed (HTTP 202, dedup_key: abc123)

========================================
  Alert Notification Test Summary
========================================

Total Tests Run:    3
Tests Passed:       3
Tests Failed:       0

✓ All alert notification tests passed!
```

### validate-monitoring-stack.sh
```
╔════════════════════════════════════════╗
║  OllamaMax Monitoring Stack Validator ║
╚════════════════════════════════════════╝

Validating monitoring infrastructure...

=== Prometheus Health Check ===

[INFO] Checking Prometheus health...
[✓] Prometheus is healthy (HTTP 200)
[INFO] Checking Prometheus ready status...
[✓] Prometheus is ready

=== Prometheus Metrics Validation ===

[INFO] Querying metric: http_requests_total
[✓] Metric http_requests_total found (5 series)

╔════════════════════════════════════════╗
║   Monitoring Stack Validation Report  ║
╠════════════════════════════════════════╣
║                                        ║
║  Total Checks:       24                ║
║  Checks Passed:      24                ║
║  Checks Failed:      0                 ║
║                                        ║
╚════════════════════════════════════════╝

✓ All monitoring stack validation checks passed!
```

## Next Steps

1. **Integration**: Add scripts to CI/CD pipeline
2. **Monitoring**: Set up scheduled runs (e.g., every 5 minutes)
3. **Alerting**: Configure alerts when validation scripts fail
4. **Expansion**: Add more notification channels as needed
5. **Documentation**: Update runbooks with script usage

## Conclusion

Comment 15 implementation is complete with two comprehensive, production-ready scripts for monitoring validation and alert notification testing. Both scripts feature:
- Robust error handling
- Beautiful color-coded output
- Comprehensive validation coverage
- CI/CD integration ready
- Secure credential handling
- Detailed reporting

The scripts are executable and ready for immediate use in validating the OllamaMax monitoring infrastructure.

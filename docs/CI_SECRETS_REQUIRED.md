# Required GitHub Secrets for CI/CD

## Monitoring Validation Job

The `monitoring-validation` job requires these secrets for testing alert notifications:

### Slack Integration (Optional)
- `TEST_SLACK_WEBHOOK`: Slack webhook URL for test notifications
  - Generate at: https://api.slack.com/messaging/webhooks
  - Format: `https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK`

### Email/SMTP (Optional)
- `TEST_SMTP_HOST`: SMTP server hostname (e.g., smtp.gmail.com)
- `TEST_SMTP_PORT`: SMTP port (typically 587 for TLS)
- `TEST_SMTP_USER`: SMTP username/email
- `TEST_SMTP_PASSWORD`: SMTP password or app-specific password

### PagerDuty (Optional)
- `TEST_PAGERDUTY_KEY`: PagerDuty integration key for test incidents
  - Generate at: https://support.pagerduty.com/docs/services-and-integrations

## Setting Secrets

In your GitHub repository:
1. Go to Settings → Secrets and variables → Actions
2. Click "New repository secret"
3. Add each secret with the exact name listed above
4. Secret values are encrypted and never exposed in logs

## Notes

- All secrets are optional - validation will skip tests for missing secrets
- Use dedicated test accounts/channels to avoid production impact
- Rotate secrets regularly for security

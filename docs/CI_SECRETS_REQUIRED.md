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

- **All secrets are optional** - validation will skip tests for missing secrets
- Use dedicated test accounts/channels to avoid production impact
- **Never use production credentials** in CI/CD secrets
- Rotate secrets regularly for security
- Test notifications will be clearly labeled as "[TEST]"

## Validation Behavior

The monitoring-validation job will:
- ✅ Run stack validation (health checks, metrics, logs, traces) - **Always runs**
- ⚠️  Skip Slack tests if `TEST_SLACK_WEBHOOK` is not set
- ⚠️  Skip email tests if SMTP secrets are not set
- ⚠️  Skip PagerDuty tests if `TEST_PAGERDUTY_KEY` is not set
- ✅ Continue with remaining tests even if optional tests are skipped

## Security Best Practices

1. **Use test-only accounts**: Create dedicated test email addresses and Slack channels
2. **Rotate regularly**: Change test credentials every 90 days
3. **Minimize permissions**: Test accounts should have minimal required permissions
4. **Monitor usage**: Review secret usage in Actions logs
5. **Audit access**: Regularly review who can access repository secrets
6. **Document clearly**: Label all test credentials as "[TEST]" or "[CI]"

## Troubleshooting

### Slack webhook fails
- Verify webhook URL format is correct
- Check webhook hasn't expired or been revoked
- Test webhook manually using `curl`

### Email tests fail
- Verify SMTP credentials are correct
- For Gmail, ensure "Less secure app access" is enabled OR use app password
- Check SMTP port (587 for TLS, 465 for SSL)
- Verify firewall isn't blocking outbound SMTP

### PagerDuty tests fail
- Verify integration key is valid
- Check service is active in PagerDuty
- Ensure API v2 endpoint is accessible
- Review PagerDuty integration status

## Example Secret Values

```bash
# Slack (example)
TEST_SLACK_WEBHOOK=https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX

# Email (example)
TEST_SMTP_HOST=smtp.gmail.com
TEST_SMTP_PORT=587
TEST_SMTP_USER=test-alerts@example.com
TEST_SMTP_PASSWORD=abcd-efgh-ijkl-mnop

# PagerDuty (example)
TEST_PAGERDUTY_KEY=a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
```

**Never commit these values to version control!**

# OllamaMax E2E Testing Suite

Comprehensive end-to-end testing infrastructure for the OllamaMax distributed AI platform using Playwright.

## 🚀 Features

### Core Testing Capabilities
- **Multi-browser testing** (Chromium, Firefox, Safari)
- **Cross-device testing** (Mobile, Tablet, Desktop)
- **Performance monitoring** and Core Web Vitals tracking
- **Security testing** including XSS, SQL injection, and header validation
- **Load testing** for concurrent users and high-traffic scenarios
- **API integration testing** for distributed inference endpoints
- **Real-time WebSocket testing**
- **Visual regression testing** with screenshot comparison

### Specialized Test Suites
- **Core Functionality**: System health, model management, distributed nodes
- **Distributed Inference**: AI model testing, load balancing, failover
- **Security Testing**: Authentication, input validation, information disclosure
- **Performance Testing**: Load times, memory usage, network analysis
- **Load Testing**: Concurrent users, throughput measurement

## 📁 Project Structure

```
tests/e2e/
├── playwright.config.ts      # Main Playwright configuration
├── global-setup.ts          # Global test setup and environment validation
├── global-teardown.ts       # Cleanup and report generation
├── package.json            # Dependencies and scripts
├── tests/                  # Test specifications
│   ├── core-functionality.spec.ts
│   ├── distributed-inference.spec.ts
│   └── security.spec.ts
├── helpers/               # Test utilities and helpers
│   ├── performance-helper.ts
│   ├── load-test-helper.ts
│   ├── security-helper.ts
│   ├── screenshot-helper.ts
│   └── metrics-helper.ts
└── reports/              # Generated test reports
    ├── playwright-report/
    ├── screenshots/
    ├── performance/
    ├── load-tests/
    └── security/
```

## 🛠️ Setup and Installation

### Prerequisites
- Node.js 16+ 
- npm or yarn
- OllamaMax application running locally or accessible via URL

### Install Dependencies
```bash
cd tests/e2e
npm install
```

### Install Playwright Browsers
```bash
npm run install-browsers
```

### Environment Configuration
Create `.env` file or set environment variables:
```bash
BASE_URL=http://localhost:8080  # Your OllamaMax instance URL
NODE_ENV=test
```

## 🎯 Running Tests

### Quick Start
```bash
# Run all tests
npm test

# Run with UI (interactive mode)
npm run test:ui

# Run specific test suite
npm run test:core
npm run test:inference
npm run test:security
npm run test:performance
```

### Browser-Specific Testing
```bash
# Run on specific browsers
npm run test:chromium
npm run test:firefox
npm run test:webkit

# Mobile testing
npm run test:mobile
```

### Advanced Testing
```bash
# Run with debugging
npm run test:debug

# Run in headed mode (visible browser)
npm run test:headed

# Run load tests
npm run test:load
```

## 📊 Test Categories

### 1. Core Functionality Tests
**Location**: `tests/core-functionality.spec.ts`

Tests fundamental platform features:
- System health and dashboard loading
- API endpoint validation
- Model management interface
- Distributed node monitoring
- Real-time WebSocket connectivity
- Performance metrics collection
- Responsive design validation
- Error handling and recovery

```bash
npm run test:core
```

### 2. Distributed Inference Tests
**Location**: `tests/distributed-inference.spec.ts`

Tests AI inference capabilities:
- Model inference API endpoints
- Distributed load balancing
- Concurrent request handling
- Streaming inference capability
- Model management and switching
- Performance under load
- Error handling and recovery
- Failover mechanisms

```bash
npm run test:inference
```

### 3. Security Tests
**Location**: `tests/security.spec.ts`

Comprehensive security validation:
- Security headers validation
- XSS prevention testing
- SQL injection prevention
- Authentication bypass attempts
- Rate limiting and DoS protection
- Information disclosure prevention
- CORS configuration validation
- File upload security
- Session management

```bash
npm run test:security
```

### 4. Performance Tests
**Project**: `performance`

Performance and load testing:
- Core Web Vitals measurement
- Memory usage monitoring
- Network performance analysis
- Resource loading optimization
- Cross-device performance
- Load testing scenarios

```bash
npm run test:performance
```

## 📈 Reports and Analysis

### Viewing Reports
```bash
# Open HTML report
npm run report

# Generate Allure report
npm run report:allure
```

### Report Types

1. **HTML Reports**: Interactive Playwright reports with screenshots and videos
2. **Screenshot Gallery**: Visual documentation of test execution
3. **Performance Reports**: Detailed performance metrics and trends
4. **Security Reports**: Vulnerability assessments and security findings
5. **Load Test Reports**: Concurrency and throughput analysis

### Report Locations
- `reports/playwright-report/` - Main HTML reports
- `reports/screenshots/` - Visual documentation
- `reports/performance/` - Performance metrics
- `reports/load-tests/` - Load testing results
- `reports/security/` - Security scan results

## 🔧 Configuration

### Playwright Configuration
Edit `playwright.config.ts` to customize:
- Browser selection
- Test timeouts
- Retry policies
- Reporter configuration
- Global test settings

### Environment Variables
```bash
BASE_URL=http://localhost:8080    # Target application URL
NODE_ENV=test                     # Environment mode
CI=true                          # CI mode flag
HEADLESS=true                    # Run headless browsers
```

### Test Data Configuration
Test data is automatically configured in `global-setup.ts`:
- Test user accounts
- Model configurations
- Performance baselines
- Security test payloads

## 🚨 Troubleshooting

### Common Issues

1. **Application Not Available**
   ```bash
   Error: Services not ready for testing
   ```
   Solution: Ensure OllamaMax is running at the specified BASE_URL

2. **Browser Installation Issues**
   ```bash
   npm run install-browsers
   ```

3. **Memory Issues on CI**
   Set resource limits in playwright.config.ts

4. **Network Timeouts**
   Increase timeout values for slower environments

### Debug Mode
```bash
# Run single test with debug
npm run test:debug -- tests/core-functionality.spec.ts

# Run with trace viewer
playwright test --trace on
```

### Logging
Enable verbose logging:
```bash
DEBUG=pw:* npm test
```

## 🔄 CI/CD Integration

### GitHub Actions Example
```yaml
name: E2E Tests
on: [push, pull_request]
jobs:
  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
      - name: Install dependencies
        run: |
          cd tests/e2e
          npm ci
      - name: Install Playwright browsers
        run: |
          cd tests/e2e
          npm run install-browsers
      - name: Run E2E tests
        run: |
          cd tests/e2e
          npm test
        env:
          BASE_URL: http://localhost:8080
      - uses: actions/upload-artifact@v3
        if: always()
        with:
          name: test-reports
          path: tests/e2e/reports/
```

## 📝 Writing Custom Tests

### Basic Test Structure
```typescript
import { test, expect } from '@playwright/test';

test.describe('My Feature', () => {
  test('should work correctly', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('h1')).toBeVisible();
  });
});
```

### Using Helpers
```typescript
import { PerformanceHelper } from '../helpers/performance-helper';
import { ScreenshotHelper } from '../helpers/screenshot-helper';

test('performance test', async ({ page }) => {
  const perfHelper = new PerformanceHelper(page);
  const screenshotHelper = new ScreenshotHelper(page);
  
  await page.goto('/');
  const metrics = await perfHelper.collectMetrics();
  await screenshotHelper.captureFullPage('my-test');
  
  expect(metrics.loadTime).toBeLessThan(5000);
});
```

## 🧹 Maintenance

### Cleanup Commands
```bash
# Clean test artifacts
npm run clean

# Clean all reports and artifacts
npm run clean:all

# Clean old screenshots (7+ days)
# Automatic cleanup in screenshot-helper.ts
```

### Code Quality
```bash
# Lint tests
npm run lint

# Fix linting issues
npm run lint:fix

# Type check
npm run type-check

# Validate all
npm run validate
```

## 🎯 Best Practices

### Test Organization
- Group related tests in describe blocks
- Use descriptive test names
- Keep tests independent and atomic
- Use page object models for complex UIs

### Performance
- Use `waitForLoadState('networkidle')` for dynamic content
- Minimize unnecessary waits
- Run tests in parallel when possible
- Clean up resources after tests

### Reliability
- Use proper selectors (data-testid preferred)
- Handle dynamic content with proper waits
- Implement retry logic for flaky operations
- Use soft assertions where appropriate

### Security Testing
- Never commit real credentials
- Use test-specific data
- Validate input sanitization
- Test authentication boundaries

## 📚 Resources

- [Playwright Documentation](https://playwright.dev/)
- [OllamaMax Platform Docs](../../../README.md)
- [Test Writing Guide](https://playwright.dev/docs/writing-tests)
- [Best Practices](https://playwright.dev/docs/best-practices)

## 🤝 Contributing

1. Follow existing code style and patterns
2. Write comprehensive tests with good coverage
3. Update documentation for new features
4. Run full test suite before submitting
5. Include performance and security considerations

## 📞 Support

For issues and questions:
- Check existing test reports and logs
- Review troubleshooting section
- Open GitHub issue with reproduction steps
- Include environment details and error logs
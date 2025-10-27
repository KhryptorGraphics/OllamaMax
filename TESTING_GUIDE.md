# OllamaMax Testing Guide

Comprehensive testing guide for the OllamaMax distributed AI platform. This guide covers all test types, setup instructions, commands, coverage requirements, and CI/CD integration.

## Table of Contents

1. [Overview](#overview)
2. [Setup](#setup)
3. [Running Tests Locally](#running-tests-locally)
4. [Coverage Generation and Interpretation](#coverage-generation-and-interpretation)
5. [CI/CD Integration](#cicd-integration)
6. [Troubleshooting](#troubleshooting)
7. [Maintenance](#maintenance)

---

## Overview

### Testing Philosophy

OllamaMax follows a comprehensive testing strategy that emphasizes:

- **Test-First Development**: Write tests before implementation (TDD)
- **High Coverage Requirements**: 90%+ coverage for both Go and JavaScript code
- **Multi-Layer Testing**: Unit, integration, E2E, performance, and security tests
- **Continuous Validation**: Automated testing in CI/CD pipelines
- **Quality Gates**: Coverage thresholds enforced at build time

### Test Types

| Test Type | Purpose | Tools | Location |
|-----------|---------|-------|----------|
| **Unit Tests** | Test individual functions and modules in isolation | Jest (JS), Go test | `tests/unit/`, Go packages |
| **Integration Tests** | Test component interactions and distributed systems | Go test, Playwright | `tests/integration/`, `ollama-distributed/tests/integration/` |
| **E2E Tests** | Test complete user workflows across the full stack | Playwright | `tests/e2e/` |
| **Performance Tests** | Measure response times, throughput, and resource usage | Jest, Go benchmarks, k6 | `tests/performance/`, `ollama-distributed/tests/performance/` |
| **Security Tests** | Validate authentication, input sanitization, and vulnerability scanning | Playwright, Snyk, OWASP ZAP | `tests/e2e/tests/security.spec.ts` |

### Coverage Requirements

**Strict 90% coverage threshold applied to:**

- **Go Code**: All packages under `pkg/`, `internal/`, `cmd/`
- **JavaScript Code**: All files under `api-server/`, `src/agents/`, `critical-fixes/`

Coverage validation occurs:
- Locally via scripts (`scripts/run-all-tests.sh`, `scripts/coverage-report.sh`)
- In CI via workflows (`.github/workflows/ci-cd-pipeline.yml`, `.github/workflows/production-pipeline.yml`)
- In Go builds via Makefile (`ollama-distributed/Makefile`)

**Build fails if coverage falls below 90%.**

---

## Setup

### Prerequisites

**Required Software:**

- **Go**: Version `1.21` (specified in `.github/workflows/ci-cd-pipeline.yml:77` and `go.mod:3`)
- **Node.js**: Version `18` or `20` (CI uses `18`, production pipeline uses `20`)
- **npm**: For JavaScript dependency management
- **PostgreSQL**: Version `15` (for database tests)
- **Redis**: Version `7` (for cache tests)

**Verify Installations:**

```bash
go version        # Should show go1.21 or higher
node --version    # Should show v18.x or v20.x
npm --version     # Should show 9.x or higher
psql --version    # Should show 15.x
redis-cli --version  # Should show 7.x
```

### Installing Dependencies

**Go Dependencies:**

```bash
# From project root
go mod download
go mod tidy

# From ollama-distributed directory
cd ollama-distributed
go mod download
```

**JavaScript Dependencies:**

```bash
# From project root
npm ci
```

**Playwright Browser Installation:**

```bash
# Install browsers with system dependencies
npx playwright install --with-deps
```

### Environment Variables

**Required Environment Variables:**

```bash
# Database Configuration
export DB_HOST=localhost
export DB_PORT=5432              # or 15432 for CI non-standard port
export DB_NAME=ollamamax_test
export DB_USER=test_user
export DB_PASSWORD=test_password

# Redis Configuration
export REDIS_HOST=localhost
export REDIS_PORT=6379           # or 16379 for CI non-standard port
export REDIS_PASSWORD=test_redis_password

# Application Configuration
export BASE_URL=http://localhost:8080
export JWT_SECRET_KEY=test-jwt-secret-key-for-testing-only
export AUTH_ENABLED=true
export NODE_ENV=test

# CI Configuration (optional, for CI mode)
export CI=true
export HEADLESS=true
```

**Create `.env` File (Optional):**

```bash
cat > .env << 'EOF'
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ollamamax_test
DB_USER=test_user
DB_PASSWORD=test_password
REDIS_HOST=localhost
REDIS_PORT=6379
BASE_URL=http://localhost:8080
JWT_SECRET_KEY=test-jwt-secret-key-for-testing-only
NODE_ENV=test
EOF
```

### Starting Required Services

**PostgreSQL:**

```bash
# Start PostgreSQL (if not running)
sudo systemctl start postgresql

# Create test database
psql -U postgres -c "CREATE DATABASE ollamamax_test;"
psql -U postgres -c "CREATE USER test_user WITH PASSWORD 'test_password';"
psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE ollamamax_test TO test_user;"
```

**Redis:**

```bash
# Start Redis (if not running)
sudo systemctl start redis

# Verify Redis is running
redis-cli ping
# Expected output: PONG
```

**Application Server (for E2E tests):**

```bash
# Build and start the application
npm run build
./bin/ollamamax &

# Or use the web interface test server
cd web-interface
python3 -m http.server 8080 &
cd ..
```

---

## Running Tests Locally

### Go Unit Tests with Coverage

**Command:**

```bash
go test -v -coverprofile=test-artifacts/coverage/go-unit-coverage.out -covermode=atomic ./pkg/... ./internal/...
```

**Expected Output:**

```
=== RUN   TestExample
--- PASS: TestExample (0.00s)
PASS
coverage: 92.3% of statements
ok      github.com/khryptorgraphics/ollamamax/pkg/api    0.123s  coverage: 92.3% of statements
```

**Artifacts Generated:**

- `test-artifacts/coverage/go-unit-coverage.out` - Coverage profile

### Go Integration Tests with Coverage

**Using Makefile:**

```bash
make -C ollama-distributed test-integration
```

**Direct Command:**

```bash
cd ollama-distributed
OLLAMA_TEST_ARTIFACTS_DIR=./test-artifacts \
OLLAMA_TEST_NODE_COUNT=3 \
go test -v -timeout=30m -coverprofile=test-artifacts/coverage/go-integration-coverage.out -covermode=atomic -tags=integration ./tests/integration/...
```

**Expected Output:**

```
=== RUN   TestDistributedConsensus
=== RUN   TestDistributedConsensus/Raft_consensus_with_3_nodes
--- PASS: TestDistributedConsensus (5.23s)
    --- PASS: TestDistributedConsensus/Raft_consensus_with_3_nodes (5.23s)
PASS
coverage: 91.7% of statements
ok      github.com/khryptorgraphics/ollamamax/ollama-distributed/tests/integration    5.234s
```

**Artifacts Generated:**

- `ollama-distributed/test-artifacts/coverage/go-integration-coverage.out` - Integration coverage profile
- `ollama-distributed/test-artifacts/logs/` - Test execution logs

### JavaScript Unit Tests with Coverage

**Command:**

```bash
npm run test:coverage
```

**Expected Output:**

```
PASS tests/critical-fixes/prewarming.test.cjs
  ✓ should validate prewarming configuration (123ms)

Test Suites: 1 passed, 1 total
Tests:       5 passed, 5 total
Snapshots:   0 total
Time:        2.345s

=============================== Coverage summary ===============================
Statements   : 92.45% ( 148/160 )
Branches     : 91.23% ( 104/114 )
Functions    : 93.75% ( 30/32 )
Lines        : 92.45% ( 148/160 )
================================================================================
```

**Artifacts Generated:**

- `coverage/coverage-summary.json` - Coverage summary in JSON format
- `coverage/lcov.info` - Coverage data in LCOV format
- `coverage/lcov-report/index.html` - HTML coverage report

### Playwright E2E Tests

**Command:**

```bash
BASE_URL=http://localhost:8080 npm run test:e2e
```

**With Specific Browser:**

```bash
# Chromium only
npx playwright test --project=chromium

# Firefox only
npx playwright test --project=firefox

# WebKit (Safari) only
npx playwright test --project=webkit
```

**Expected Output:**

```
Running 12 tests using 2 workers

  ✓  [chromium] › core-functionality.spec.ts:15:1 › Core Functionality › Dashboard loads successfully (2s)
  ✓  [chromium] › core-functionality.spec.ts:23:1 › Core Functionality › API health endpoint responds (1s)

12 passed (45s)

To open last HTML report run:
  npx playwright show-report test-results/playwright-report
```

**Artifacts Generated:**

- `test-results/playwright-report/` - HTML report with screenshots and videos
- `test-results/playwright/results.json` - Test results in JSON format

### Performance Tests

**Go Performance Benchmarks:**

```bash
make -C ollama-distributed test-performance
```

**Direct Command:**

```bash
cd ollama-distributed
OLLAMA_TEST_ARTIFACTS_DIR=./test-artifacts \
OLLAMA_TEST_NODE_COUNT=5 \
go test -v -timeout=60m -bench=. -benchmem ./tests/performance/...
```

**Jest Performance Tests:**

```bash
jest tests/performance-*.test.js --config=jest.config.cjs
```

**Expected Output:**

```
BenchmarkInferenceLatency-8         1000    1234567 ns/op    4096 B/op    32 allocs/op
BenchmarkThroughput-8               5000    234567 ns/op     2048 B/op    16 allocs/op
PASS
ok      github.com/khryptorgraphics/ollamamax/ollama-distributed/tests/performance    12.345s
```

### Unified Test Execution

**Run All Tests in One Command:**

```bash
scripts/run-all-tests.sh
```

**What This Script Does:**

1. **Phase 1**: Go unit tests with coverage
2. **Phase 2**: Go integration tests (via Makefile)
3. **Phase 3**: JavaScript unit tests with coverage
4. **Phase 4**: E2E tests with Playwright
5. **Phase 5**: Performance tests
6. **Coverage Merging**: Combines Go coverage profiles
7. **Validation**: Checks 90% threshold for both Go and JS

**Expected Summary Output:**

```
═══════════════════════════════════════════════════════
  Test Execution Summary
═══════════════════════════════════════════════════════
Total Test Suites:  5
Passed:            5
Failed:            0
Skipped:           0
Success Rate:       100%

Artifacts saved to: /home/kp/OllamaMax/test-artifacts
Coverage reports:   /home/kp/OllamaMax/test-artifacts/coverage
Logs:               /home/kp/OllamaMax/test-artifacts/logs
═══════════════════════════════════════════════════════

✅ All tests passed and coverage meets requirements!
```

**Artifacts Generated:**

- `test-artifacts/coverage/go-merged-coverage.out` - Merged Go coverage profile
- `test-artifacts/coverage/go-coverage.html` - HTML visualization of Go coverage
- `test-artifacts/logs/` - All test execution logs with timestamps
- `coverage/` - JavaScript coverage reports

---

## Coverage Generation and Interpretation

### How Coverage Merging Works

**Go Coverage Merging (`scripts/coverage-report.sh`):**

```bash
# Merges all .out files from test-artifacts/coverage/
echo "mode: atomic" > test-artifacts/coverage/merged-coverage.out
tail -n +2 test-artifacts/coverage/go-unit-coverage.out >> test-artifacts/coverage/merged-coverage.out
tail -n +2 ollama-distributed/test-artifacts/coverage/go-integration-coverage.out >> test-artifacts/coverage/merged-coverage.out

# Generate HTML report
go tool cover -html=test-artifacts/coverage/merged-coverage.out -o test-artifacts/coverage/go-coverage.html
```

**JavaScript Coverage (`coverage-summary.json`):**

```bash
# Coverage automatically aggregated by Jest during test run
npm run test:coverage

# Reads from coverage/coverage-summary.json
node scripts/validate-coverage.js
```

### Coverage Thresholds

**Jest Configuration (`jest.config.cjs:17-24`):**

```javascript
coverageThreshold: {
  global: {
    branches: 90,
    functions: 90,
    lines: 90,
    statements: 90
  }
}
```

**Go Coverage Gate (Makefile):**

From `ollama-distributed/Makefile:200-208`:

```makefile
COVERAGE_THRESHOLD ?= 90

COVERAGE=$$(go tool cover -func=$(COVERAGE_DIR)/go-merged-coverage.out | grep total | grep -Eo '[0-9]+\.[0-9]+' || echo "0"); \
if [ -z "$$COVERAGE" ] || [ "$$(awk "BEGIN {print ($$COVERAGE < $(COVERAGE_THRESHOLD))}")" -eq 1 ]; then \
    echo "❌ Coverage $$COVERAGE% is below threshold $(COVERAGE_THRESHOLD)%"; \
    exit 1; \
else \
    echo "✅ Coverage $$COVERAGE% meets threshold $(COVERAGE_THRESHOLD)%"; \
fi
```

**CI Workflows:**

- `.github/workflows/ci-cd-pipeline.yml:118-126` - Go coverage validation
- `.github/workflows/ci-cd-pipeline.yml:131-132` - JS coverage validation via `npm run validate:coverage`
- `.github/workflows/production-pipeline.yml:71-86` - Both Go and JS coverage validation

### What Failure Output Looks Like

**Go Coverage Failure:**

```
COVERAGE=$(go tool cover -func=coverage.out | grep total | grep -Eo '[0-9]+\.[0-9]+')
echo "Go Coverage: 87.3%"
❌ Go coverage 87.3% is below threshold 90%
exit code 1
```

**JavaScript Coverage Failure:**

```
Jest: "global" coverage threshold for statements (90%) not met: 87.45%
Jest: "global" coverage threshold for branches (90%) not met: 85.23%
FAIL
```

### Reading Coverage Reports

**Go Coverage - Command Line:**

```bash
# Show coverage by function
go tool cover -func=test-artifacts/coverage/go-merged-coverage.out

# Sample output:
# github.com/khryptorgraphics/ollamamax/pkg/api/server.go:45:  HandleRequest   92.3%
# github.com/khryptorgraphics/ollamamax/pkg/api/server.go:78:  ValidateAuth    88.5%
# ...
# total:                                                        (statements)    91.7%
```

**Go Coverage - HTML Report:**

```bash
# Open in browser
open test-artifacts/coverage/go-coverage.html

# Or for merged coverage
open test-artifacts/coverage/coverage.html
```

**JavaScript Coverage - HTML Report:**

```bash
# Open in browser
open coverage/lcov-report/index.html
```

**Locating Low-Coverage Files:**

```bash
# Find Go files below 90% (from coverage report script)
go tool cover -func=test-artifacts/coverage/go-merged-coverage.out | grep -v "100.0%" | grep -v "total:" | head -n 20

# Sample output:
# pkg/database/connection.go:123:  Connect         75.0%
# internal/cache/redis.go:45:      GetValue        82.3%
```

---

## CI/CD Integration

### CI/CD Pipeline (`ci-cd-pipeline.yml`)

**Triggered On:**
- Push to `main` or `develop` branches
- Pull requests to `main`

**Key Jobs:**

1. **Test Job** (`.github/workflows/ci-cd-pipeline.yml:38-164`):
   - Go version: `1.21`
   - Node version: `18`
   - Services: PostgreSQL (port 15432), Redis (port 16379)
   - Steps:
     - Go tests with coverage (line 108-115)
     - Go coverage validation 90% (line 118-126)
     - JS tests with coverage (line 129)
     - JS coverage validation (line 131-132)
     - Playwright E2E tests (line 134-141)
     - Artifact upload: `playwright-report` (line 143-149)

2. **Build and Test Docker** (line 165-230):
   - Build Docker images
   - Test service connectivity on non-standard ports
   - Run integration tests

3. **Security Scan** (line 232-258):
   - Trivy vulnerability scanning
   - Docker security scan

**Coverage Validation Commands:**

```yaml
# Go Coverage (lines 118-126)
- name: Validate Go coverage threshold (90%)
  run: |
    COVERAGE=$(go tool cover -func=coverage.out | grep total | grep -Eo '[0-9]+\.[0-9]+')
    echo "Go Coverage: ${COVERAGE}%"
    if [ "$(awk "BEGIN {print ($COVERAGE < 90)}")" -eq 1 ]; then
      echo "❌ Go coverage ${COVERAGE}% is below threshold 90%"
      exit 1
    else
      echo "✅ Go coverage ${COVERAGE}% meets threshold 90%"
    fi

# JavaScript Coverage (lines 131-132)
- name: Validate JavaScript coverage threshold (90%)
  run: npm run validate:coverage
```

### Production Pipeline (`production-pipeline.yml`)

**Triggered On:**
- Push to `main` or `production` branches
- Pull requests to `main` or `production`

**Key Jobs:**

1. **Code Quality** (line 35-95):
   - Go linting with golangci-lint
   - Go tests with coverage and 90% validation (line 65-80)
   - JS tests with coverage and 90% validation (line 82-86)
   - Coverage artifact upload

2. **Performance Testing** (line 96-131):
   - Performance benchmarks
   - Jest performance tests (line 124)
   - Artifact upload: `performance-reports` to `test-results/performance/`

3. **Browser Testing** (line 132-159):
   - Matrix strategy: `[chromium, firefox, webkit]`
   - Playwright browser-specific tests

4. **Load Testing** (line 161-193):
   - k6 load testing with 100 VUs for 5 minutes

### Downloading CI Artifacts

**Via GitHub Web UI:**

1. Navigate to Actions tab
2. Select the workflow run
3. Scroll to "Artifacts" section
4. Download:
   - `playwright-report` (E2E test results)
   - `coverage-reports` (Go coverage HTML and profiles)
   - `performance-reports` (Performance test results)
   - `test-results-chromium`, `test-results-firefox`, `test-results-webkit` (Browser-specific results)

**Via GitHub CLI:**

```bash
# List artifacts for a run
gh run view <run-id> --log

# Download all artifacts
gh run download <run-id>

# Download specific artifact
gh run download <run-id> -n playwright-report
```

### Artifact Upload Paths

| Artifact Name | Path | Workflow |
|---------------|------|----------|
| `playwright-report` | `test-results/playwright-report/` | ci-cd-pipeline.yml:148 |
| `coverage-reports` | `coverage.out`, `coverage.html` | production-pipeline.yml:92 |
| `performance-reports` | `test-results/performance/` | production-pipeline.yml:130 |
| `test-results-{browser}` | `test-results/` | production-pipeline.yml:158 |
| `load-test-results` | `load-test-report.json` | production-pipeline.yml:192 |

---

## Troubleshooting

### Service Readiness Issues

**Problem: "Services not ready for testing"**

```
Error: connect ECONNREFUSED 127.0.0.1:5432
```

**Solutions:**

```bash
# 1. Check if PostgreSQL is running
sudo systemctl status postgresql
sudo systemctl start postgresql

# 2. Verify database exists
psql -U postgres -l | grep ollamamax_test

# 3. Test connection
psql -U test_user -d ollamamax_test -h localhost -p 5432
```

**Problem: "Redis connection refused"**

```
Error: connect ECONNREFUSED 127.0.0.1:6379
```

**Solutions:**

```bash
# 1. Check if Redis is running
sudo systemctl status redis
sudo systemctl start redis

# 2. Test connection
redis-cli ping

# 3. Check port
netstat -tlnp | grep 6379
```

**Problem: "BASE_URL not accessible"**

```
Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:8080
```

**Solutions:**

```bash
# 1. Ensure application is running
ps aux | grep ollamamax

# 2. Start application
npm run build
./bin/ollamamax &

# 3. Verify health endpoint
curl http://localhost:8080/health

# 4. Check port availability
lsof -i :8080
```

### Playwright Browser Installation Errors

**Problem: "Executable doesn't exist"**

```
Error: browserType.launch: Executable doesn't exist at /home/user/.cache/ms-playwright/chromium-1084/chrome-linux/chrome
```

**Solutions:**

```bash
# Install browsers with system dependencies
npx playwright install --with-deps

# If on CI/Docker, ensure proper base image
# Playwright provides official Docker images:
# mcr.microsoft.com/playwright:v1.40.0-focal
```

**Problem: "Missing dependencies"**

```
Error: Host system is missing dependencies to run browsers.
```

**Solutions:**

```bash
# Ubuntu/Debian
npx playwright install-deps

# Or manually
sudo apt-get update
sudo apt-get install -y \
  libnss3 libnspr4 libatk1.0-0 libatk-bridge2.0-0 \
  libcups2 libdrm2 libxkbcommon0 libxcomposite1 \
  libxdamage1 libxfixes3 libxrandr2 libgbm1 \
  libasound2
```

### Test Timeouts and Flakes

**Problem: "Test timeout of 30000ms exceeded"**

**Solutions:**

```bash
# 1. Increase timeout in playwright.config.js
# Line 7: timeout: 60000

# 2. For specific tests
test('slow test', async ({ page }) => {
  test.setTimeout(60000);
  // ...
});

# 3. Check for network issues
# Ensure BASE_URL is accessible and responsive
curl -w "@curl-format.txt" http://localhost:8080/
```

**Problem: "Flaky tests"**

**Solutions:**

```bash
# 1. Enable retries (already configured)
# playwright.config.js:10: retries: process.env.CI ? 2 : 1

# 2. Add explicit waits
await page.waitForLoadState('networkidle');
await page.waitForSelector('[data-testid="element"]');

# 3. Run tests sequentially
# playwright.config.js:8: fullyParallel: false
```

### Coverage Validation Failures

**Problem: "Coverage below threshold"**

```
❌ Go coverage 87.3% is below threshold 90%
```

**Solutions:**

```bash
# 1. Identify low-coverage files
go tool cover -func=test-artifacts/coverage/go-merged-coverage.out | grep -v "100.0%" | grep -v "total:" | head -n 20

# 2. Open HTML report to see uncovered lines
open test-artifacts/coverage/go-coverage.html

# 3. Add tests for uncovered code paths
# Focus on:
# - Error handling paths
# - Edge cases
# - Conditional branches

# 4. Run specific package tests to verify improvement
go test -v -coverprofile=coverage.out ./pkg/database/...
go tool cover -func=coverage.out
```

**Problem: "No coverage data found"**

**Solutions:**

```bash
# 1. Ensure test artifacts directory exists
mkdir -p test-artifacts/coverage

# 2. Run tests with correct flags
go test -coverprofile=test-artifacts/coverage/coverage.out -covermode=atomic ./...

# 3. Check file permissions
ls -la test-artifacts/coverage/
chmod -R 755 test-artifacts/
```

### Checking Logs

**Test Execution Logs:**

```bash
# Unified test script logs
cat test-artifacts/logs/test-execution.log

# Phase-specific logs
cat test-artifacts/logs/Go\ Unit\ Tests-*.log
cat test-artifacts/logs/E2E\ Tests-*.log

# Playwright logs
ls -la test-results/playwright-report/
```

**Playwright HTML Reports:**

```bash
# Open report
npx playwright show-report test-results/playwright-report

# View specific test traces
# Navigate to failed test and click "Trace"
```

---

## Maintenance

### Adding New Tests

**1. Unit Tests (Go):**

```go
// pkg/api/server_test.go
package api

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestNewFeature(t *testing.T) {
    assert.NotNil(t, NewFeature())
}
```

**2. Unit Tests (JavaScript):**

```javascript
// tests/unit/feature.test.cjs
describe('New Feature', () => {
  test('should work correctly', () => {
    expect(newFeature()).toBeTruthy();
  });
});
```

**3. E2E Tests (Playwright):**

```typescript
// tests/e2e/tests/new-feature.spec.ts
import { test, expect } from '@playwright/test';

test('new feature works end-to-end', async ({ page }) => {
  await page.goto('/');
  await page.click('[data-testid="new-feature"]');
  await expect(page.locator('.result')).toBeVisible();
});
```

### Updating Coverage Thresholds

**Jest (`jest.config.cjs`):**

```javascript
coverageThreshold: {
  global: {
    branches: 95,     // Increased from 90
    functions: 95,
    lines: 95,
    statements: 95
  }
}
```

**Makefile (`ollama-distributed/Makefile`):**

```makefile
COVERAGE_THRESHOLD ?= 95  # Increased from 90
```

**CI Workflows:**

Update threshold checks in:
- `.github/workflows/ci-cd-pipeline.yml:121` - Change `90` to new value
- `.github/workflows/production-pipeline.yml:75` - Change `90` to new value

### Keeping This Guide Synced

**When Scripts Change:**

- Update commands in "Running Tests Locally" section
- Update artifact paths if output locations change
- Update expected output examples

**When Workflows Change:**

- Update job descriptions in "CI/CD Integration" section
- Update artifact names and paths table
- Update environment variable lists

**When Coverage Requirements Change:**

- Update "Coverage Requirements" section
- Update all threshold references (90% → new value)
- Update code snippets showing threshold validation

**When New Test Types Added:**

- Add new section under "Test Types" table
- Add new commands under "Running Tests Locally"
- Update unified test script documentation

---

## Quick Reference

### Essential Commands

```bash
# Run all tests
scripts/run-all-tests.sh

# Go unit tests
go test -v -coverprofile=test-artifacts/coverage/go-unit-coverage.out -covermode=atomic ./pkg/... ./internal/...

# Go integration tests
make -C ollama-distributed test-integration

# JavaScript tests
npm run test:coverage

# E2E tests
npm run test:e2e

# Performance tests
make -C ollama-distributed test-performance

# View coverage
open test-artifacts/coverage/go-coverage.html
open coverage/lcov-report/index.html
```

### Key Files

| File | Purpose |
|------|---------|
| `scripts/run-all-tests.sh` | Unified test execution |
| `scripts/coverage-report.sh` | Coverage report generation |
| `jest.config.cjs` | JavaScript test configuration |
| `playwright.config.js` | E2E test configuration |
| `ollama-distributed/Makefile` | Go test targets and coverage gates |
| `.github/workflows/ci-cd-pipeline.yml` | Main CI/CD workflow |
| `.github/workflows/production-pipeline.yml` | Production quality gates |

### Artifact Locations

```
test-artifacts/
├── coverage/
│   ├── go-unit-coverage.out
│   ├── go-merged-coverage.out
│   ├── go-coverage.html
│   └── coverage-badge.json
├── logs/
│   └── test-execution.log
coverage/
├── coverage-summary.json
└── lcov-report/
    └── index.html
test-results/
├── playwright-report/
└── performance/
```

---

**For questions or issues, refer to:**
- [TEST_STRATEGY.md](TEST_STRATEGY.md) - Overall testing strategy
- [tests/e2e/README.md](tests/e2e/README.md) - E2E testing details
- [CI/CD Workflows](.github/workflows/) - Pipeline configurations

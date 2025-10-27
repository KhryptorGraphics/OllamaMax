# OllamaMax Testing Guide

**Comprehensive guide for running, maintaining, and understanding the OllamaMax test suite.**

---

## Table of Contents

1. [Overview](#overview)
2. [Setup and Prerequisites](#setup-and-prerequisites)
3. [Test Categories](#test-categories)
4. [Executing Tests](#executing-tests)
5. [Coverage Generation and Interpretation](#coverage-generation-and-interpretation)
6. [CI Workflow Integration](#ci-workflow-integration)
7. [Troubleshooting](#troubleshooting)
8. [Best Practices](#best-practices)

---

## Overview

OllamaMax employs a comprehensive, multi-layered testing strategy covering:

- **Unit Tests**: Go (pkg/*, internal/*) and JavaScript (Jest)
- **Integration Tests**: Go distributed components testing
- **E2E Tests**: Playwright-based browser automation
- **Performance Tests**: Load testing and benchmarking

### Testing Philosophy

- **Test-Driven Development**: Write tests before implementation
- **High Coverage**: Maintain >90% code coverage across all languages
- **Automated CI/CD**: All tests run automatically on every commit
- **Production Readiness**: Comprehensive validation before deployment

### Key Metrics

- **Current Coverage Target**: 90% (both Go and JavaScript)
- **Test Execution Time**: ~5-10 minutes (full suite)
- **CI Integration**: GitHub Actions with parallel execution
- **Test Count**: 100+ unit tests, 50+ integration tests, 30+ E2E tests

---

## Setup and Prerequisites

### System Requirements

- **Node.js**: 18+ (for JavaScript tests and Playwright)
- **Go**: 1.21+ (for Go tests)
- **npm**: Latest version
- **Git**: For version control
- **Docker**: Optional, for integration test environments

### Initial Setup

```bash
# Clone the repository
git clone <repository-url>
cd OllamaMax

# Install Node dependencies
npm ci

# Install Go dependencies
go mod download

# Install Playwright browsers (for E2E tests)
npx playwright install --with-deps

# Install development tools
cd ollama-distributed
make deps-dev
```

### Environment Variables

Create a `.env` file in the project root (for local testing):

```bash
# API Configuration
BASE_URL=http://localhost:8080
API_URL=http://localhost:11434

# Database (for integration tests)
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ollamamax_test
DB_USER=test_user
DB_PASSWORD=test_password

# Redis (for integration tests)
REDIS_HOST=localhost
REDIS_PORT=6379

# Test Configuration
NODE_ENV=test
CI=false
HEADLESS=true
```

---

## Test Categories

### 1. Unit Tests

#### Go Unit Tests

**Location**: `pkg/**/*_test.go`, `internal/**/*_test.go`

**Purpose**: Test individual Go packages and functions in isolation

**Execution**:
```bash
# Run all Go unit tests
go test -v ./pkg/... ./internal/...

# With coverage
go test -v -coverprofile=coverage.out ./pkg/... ./internal/...

# Specific package
go test -v ./pkg/api/...
```

**Examples**:
- API handler tests (`pkg/api/server_test.go`)
- Configuration validation tests
- Database operation tests
- P2P network tests

#### JavaScript Unit Tests

**Location**: `tests/**/*.test.cjs`, `tests/**/*.test.js`

**Purpose**: Test JavaScript modules, API servers, and utilities

**Execution**:
```bash
# Run all Jest tests
npm test

# With coverage
npm run test:coverage

# Watch mode (for development)
npm run test:watch

# Specific test file
npm test tests/critical-fixes/prewarming.test.cjs
```

**Test Files**:
- `tests/critical-fixes/prewarming.test.cjs` - Agent pool prewarming
- API server tests
- WebSocket connection tests
- Configuration parsing tests

### 2. Integration Tests

**Location**: `ollama-distributed/tests/integration/`

**Purpose**: Test interaction between distributed components

**Execution**:
```bash
cd ollama-distributed

# Run integration tests
make test-integration

# With specific node count
OLLAMA_TEST_NODE_COUNT=5 make test-integration

# With coverage (now enabled!)
make test-integration  # Coverage enabled by default
```

**Test Scenarios**:
- Multi-node cluster communication
- Load balancing across nodes
- Failover and recovery
- Model synchronization
- P2P network discovery

### 3. E2E (End-to-End) Tests

**Location**: `tests/e2e/tests/`

**Purpose**: Browser-based testing of complete user workflows

**Execution**:
```bash
# Run all E2E tests
npm run test:e2e

# UI mode (interactive)
npm run test:e2e:ui

# Specific browser
npx playwright test --project=chromium
npx playwright test --project=firefox
npx playwright test --project=webkit

# Debug mode
npm run test:e2e:debug

# Headed mode (visible browser)
npx playwright test --headed
```

**Test Suites**:
- `core-functionality.spec.ts` - System health, dashboard, API
- `distributed-inference.spec.ts` - AI model inference, load balancing
- `security.spec.ts` - XSS, SQL injection, authentication

**Reports**:
```bash
# View HTML report
npm run test:e2e:report

# Reports location
test-results/playwright-report/index.html
```

### 4. Performance Tests

**Location**: `tests/performance/`, `ollama-distributed/tests/performance/`

**Purpose**: Benchmark and validate performance under load

**Execution**:
```bash
# Go performance benchmarks
cd ollama-distributed
make test-performance

# JavaScript performance tests (now wired into CI!)
jest tests/performance-*.test.js --config=jest.config.cjs

# Run all performance tests
npm run test:performance
```

**Metrics Tracked**:
- Request latency
- Throughput (requests/second)
- Memory usage
- CPU utilization
- Concurrent user handling

---

## Executing Tests

### Running Complete Test Suite

#### Using Unified Script (Recommended)

```bash
# Run all tests with coverage aggregation
./scripts/run-all-tests.sh
```

This script executes:
1. Go unit tests with coverage
2. Go integration tests with coverage (now included!)
3. JavaScript unit tests with coverage
4. E2E tests (Playwright)
5. Performance benchmarks
6. Coverage validation and merging

#### Using Makefile (Go-specific)

```bash
cd ollama-distributed

# Quick test (unit only)
make quick-test

# All tests except performance/chaos
make test-all

# Full test suite (including performance/chaos)
make test-full

# CI mode
make test-ci
```

#### Using npm Scripts

```bash
# JavaScript unit tests
npm test

# E2E tests
npm run test:e2e

# Coverage generation
npm run test:coverage

# All JavaScript tests
npm run test:all
```

### Running Specific Test Suites

#### By Test Type

```bash
# Go unit tests only
go test ./pkg/... ./internal/...

# Go integration tests only
cd ollama-distributed && make test-integration

# E2E core functionality only
npx playwright test tests/core-functionality.spec.ts

# Performance tests only
cd ollama-distributed && make test-performance
```

#### By Component

```bash
# API tests
go test ./pkg/api/...

# Database tests
go test ./internal/database/...

# P2P network tests
cd ollama-distributed && make test-p2p

# Consensus tests
cd ollama-distributed && make test-consensus
```

### Test Filtering

```bash
# Go: Run specific test
go test -v -run TestServerCreation ./pkg/api/

# Jest: Run specific test
npm test -- --testNamePattern="AgentPoolManager"

# Playwright: Run specific test
npx playwright test --grep "should load dashboard"
```

---

## Coverage Generation and Interpretation

### Go Coverage

#### Generate Coverage Report

```bash
# Unit tests coverage
go test -coverprofile=coverage.out ./pkg/... ./internal/...

# Integration tests coverage (now included!)
cd ollama-distributed
make test-integration  # Generates go-integration-coverage.out

# Merge coverage files
./scripts/run-all-tests.sh  # Automatically merges to go-merged-coverage.out

# HTML report
go tool cover -html=coverage.out -o coverage.html
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

#### Interpret Go Coverage

```bash
# Function-level coverage summary
go tool cover -func=coverage.out

# Example output:
# github.com/khryptorgraphics/ollamamax/pkg/api/server.go:42:    NewServer       100.0%
# github.com/khryptorgraphics/ollamamax/pkg/api/server.go:67:    Start           85.7%
# total:                                                         (statements)    92.3%
```

**Coverage Threshold**: 90% (enforced in CI and Makefile)

### JavaScript Coverage

#### Generate Coverage Report

```bash
# Run tests with coverage
npm run test:coverage

# Coverage reports generated in:
# - coverage/lcov-report/index.html (HTML)
# - coverage/coverage-summary.json (JSON)
# - coverage/lcov.info (LCOV format)
```

#### Interpret JavaScript Coverage

Open `coverage/lcov-report/index.html` in a browser to see:

- **Line Coverage**: Percentage of lines executed
- **Branch Coverage**: Percentage of conditional branches taken
- **Function Coverage**: Percentage of functions called
- **Statement Coverage**: Percentage of statements executed

**Coverage Validation**:
```bash
# Automatic validation against 90% threshold
npm run validate:coverage

# Uses scripts/validate-coverage.js
```

### Merged Coverage Reports

The unified test script (`./scripts/run-all-tests.sh`) now merges:

1. **Go unit coverage** → `coverage/go-unit-coverage.out`
2. **Go integration coverage** → `coverage/go-integration-coverage.out` ✅ NEW!
3. **Merged Go coverage** → `coverage/go-merged-coverage.out`
4. **JavaScript coverage** → `coverage/` directory

**Access Merged Reports**:
```bash
# Go merged HTML report
open test-artifacts/coverage/go-coverage.html

# JavaScript HTML report
open coverage/lcov-report/index.html

# Check logs
cat test-artifacts/logs/test-execution.log
```

---

## CI Workflow Integration

### GitHub Actions Workflows

#### 1. Main CI/CD Pipeline (`.github/workflows/ci-cd-pipeline.yml`)

**Triggers**: Push to `main`/`develop`, Pull Requests to `main`

**Jobs**:
- **test**: Runs all tests with coverage validation
  - ✅ NEW: Playwright E2E tests now included!
- **build-and-test-docker**: Docker image build and integration tests
- **security-scan**: Trivy vulnerability scanning
- **deploy-staging**: Auto-deploy to staging (develop branch)
- **deploy-production**: Auto-deploy to production (main branch)

**Key Steps**:
```yaml
- Run Go tests with coverage
- Validate Go coverage threshold (90%)
- Run JavaScript tests with coverage
- Validate JavaScript coverage threshold (90%)
- Install Playwright browsers  # ✅ NEW!
- Run Playwright E2E tests  # ✅ NEW!
- Upload Playwright report  # ✅ NEW!
- Run linting
- Build application
- Run security audit
```

#### 2. Production Pipeline (`.github/workflows/production-pipeline.yml`)

**Triggers**: Push/PR to `main`/`production`

**Jobs**:
- **security-scan**: Snyk, OWASP ZAP
- **code-quality**: Linting, coverage validation
- **performance-testing**: Playwright performance tests + Jest benchmarks ✅ NEW!
- **browser-testing**: Multi-browser E2E tests
- **load-testing**: k6 load tests
- **database-validation**: PostgreSQL/Redis tests
- **docker-build**: Docker image creation

**Performance Tests** (now wired in!):
```yaml
- name: Run performance tests
  run: |
    npm run test:performance
    npm run test:performance:stress

- name: Run Jest performance benchmarks  # ✅ NEW!
  run: jest tests/performance-*.test.js --config=jest.config.cjs
```

#### 3. Coverage Gate Workflow (`.github/workflows/coverage-gate.yml`)

**Purpose**: Enforce 90% coverage threshold

**Execution**:
```yaml
- Validate Go coverage >= 90%
- Validate JavaScript coverage >= 90%
- Fail build if thresholds not met
```

### Running Tests Locally Like CI

```bash
# Simulate CI environment
CI=true npm test

# Run with same parallelism as CI
CI=true npm run test:coverage

# Go tests in CI mode
cd ollama-distributed
OLLAMA_TEST_CI=true make test-ci
```

### CI Artifacts

**Generated Artifacts** (downloadable from GitHub Actions):
- Coverage reports (HTML and LCOV)
- Test execution logs
- Playwright test reports and videos ✅ NEW!
- Performance benchmark results
- Security scan results

---

## Troubleshooting

### Common Issues

#### 1. Go Tests Failing: "package not found"

**Symptom**: `package github.com/khryptorgraphics/ollamamax/... is not in GOROOT`

**Solution**: ✅ FIXED! Import paths corrected in `pkg/api/server_test.go`
```bash
# Verify go.mod module path
cat go.mod  # Should show: module github.com/khryptorgraphics/ollamamax

# Re-download dependencies
go mod download
go mod tidy

# Clear module cache if persistent
go clean -modcache
go mod download
```

#### 2. Integration Tests Timing Out

**Symptom**: Tests exceed 30-minute timeout

**Solution**:
```bash
# Increase timeout
cd ollama-distributed
TEST_TIMEOUT=60m make test-integration

# Reduce node count for faster execution
OLLAMA_TEST_NODE_COUNT=2 make test-integration
```

#### 3. Playwright Browser Installation Issues

**Symptom**: `Executable doesn't exist at /path/to/browser`

**Solution**:
```bash
# Reinstall browsers with system dependencies
npx playwright install --with-deps

# Or specific browser
npx playwright install chromium --with-deps
```

#### 4. Coverage Below Threshold

**Symptom**: `❌ Coverage X% is below threshold 90%`

**Solution**: ✅ IMPROVED! Coverage validation now uses merged file with proper error handling
```bash
# Identify uncovered code
go tool cover -html=coverage.out  # Go
open coverage/lcov-report/index.html  # JavaScript

# Add tests for uncovered lines/functions
# Run coverage again to verify
```

#### 5. E2E Tests Failing: "Application Not Available"

**Symptom**: `Error: Services not ready for testing`

**Solution**:
```bash
# Ensure application is running
npm run start &  # Or appropriate start command
sleep 10  # Wait for startup

# Verify BASE_URL is correct
export BASE_URL=http://localhost:8080
npm run test:e2e

# Check webServer config in playwright.config.js
```

#### 6. Flaky Prewarming Tests

**Symptom**: Intermittent failures in `prewarming.test.cjs`

**Solution**: ✅ FIXED! Tests now use Jest fake timers for deterministic timing
```javascript
// Tests use jest.useFakeTimers() for deterministic timing
// No real delays, faster and more reliable execution
beforeEach(() => {
  jest.useFakeTimers();
});

afterEach(() => {
  jest.clearAllTimers();
  jest.useRealTimers();
});
```

### Debug Mode

#### Go Tests Debug
```bash
# Verbose output
go test -v ./pkg/api/...

# With race detection
go test -race ./pkg/api/...

# With trace
go test -trace=trace.out ./pkg/api/...
go tool trace trace.out
```

#### Jest Tests Debug
```bash
# Run single test with debug
node --inspect-brk node_modules/.bin/jest tests/critical-fixes/prewarming.test.cjs

# Verbose output
npm test -- --verbose

# No coverage (faster debugging)
npm test -- --no-coverage
```

#### Playwright Tests Debug
```bash
# Debug mode with inspector
npm run test:e2e:debug

# Step through with UI
npx playwright test --ui

# Headed mode (watch browser)
npx playwright test --headed --project=chromium
```

### Logging

Enable verbose logging for detailed diagnostics:

```bash
# Go tests
OLLAMA_TEST_LOG_LEVEL=debug go test -v ./...

# Playwright
DEBUG=pw:* npm run test:e2e

# Check test execution logs
cat test-artifacts/logs/test-execution.log
```

---

## Best Practices

### Writing Tests

1. **Follow AAA Pattern**: Arrange, Act, Assert
   ```javascript
   test('should do something', () => {
     // Arrange: Setup test data
     const input = createTestData();

     // Act: Execute functionality
     const result = functionUnderTest(input);

     // Assert: Verify expectations
     expect(result).toBe(expected);
   });
   ```

2. **Keep Tests Independent**: Each test should run in isolation
3. **Use Descriptive Names**: Test names should explain what is being tested
4. **Test Edge Cases**: Include boundary conditions and error scenarios
5. **Mock External Dependencies**: Use mocks for databases, APIs, network calls
6. **Use Fake Timers**: For tests involving setTimeout/setInterval (see prewarming.test.cjs)

### Maintaining Tests

1. **Run Tests Before Commits**: Ensure all tests pass locally
2. **Update Tests with Code Changes**: Keep tests in sync with implementation
3. **Monitor Coverage Trends**: Don't let coverage drop below 90%
4. **Review Test Failures**: Investigate and fix flaky tests immediately
5. **Clean Up Obsolete Tests**: Remove or update tests for removed features

### Performance Optimization

1. **Parallelize Where Possible**: Use Jest `--maxWorkers` and Playwright projects
2. **Use Test Fixtures**: Reuse setup/teardown code
3. **Minimize I/O**: Mock file system operations in unit tests
4. **Optimize Test Data**: Use minimal, realistic test data
5. **Profile Slow Tests**: Identify and optimize long-running tests
6. **Use Fake Timers**: Avoid real delays in tests (faster, more reliable)

### CI/CD Integration

1. **Fast Feedback**: Prioritize fast-running tests in CI
2. **Fail Fast**: Configure CI to stop on first failure (where appropriate)
3. **Artifact Retention**: Save test reports and logs for debugging
4. **Threshold Enforcement**: Use coverage gates to maintain quality
5. **Regular Maintenance**: Update dependencies and tools regularly

---

## Summary of Recent Improvements

### ✅ Comment 1: TESTING_GUIDE.md Created
Comprehensive testing guide with setup, execution, coverage, CI integration, and troubleshooting.

### ✅ Comment 2: Integration Coverage Added
- Makefile updated: `test-integration` now generates coverage with `-coverprofile`
- `run-all-tests.sh` merges integration coverage into `go-merged-coverage.out`

### ✅ Comment 3: E2E Tests in CI
- Added dedicated Playwright E2E job to `ci-cd-pipeline.yml`
- Installs browsers, runs tests, uploads reports as artifacts

### ✅ Comment 4: Performance Tests in CI
- `production-pipeline.yml` now runs `jest tests/performance-*.test.js`
- Performance benchmarks integrated into CI workflow

### ✅ Comment 5: Go Import Paths Fixed
- `pkg/api/server_test.go` updated to use correct module path
- Changed from `internal/config` to `ollama-distributed/internal/config`

### ✅ Comment 6: Fake Timers for Prewarming Tests
- `prewarming.test.cjs` switched to `jest.useFakeTimers()`
- Deterministic timing, faster execution, no flakiness

### ✅ Comment 7: Coverage Gating Normalized
- Makefile `test-coverage` target merges unit + integration coverage
- Uses `set -o pipefail` for proper error handling
- Single merged file validation with explicit fallbacks

---

## Additional Resources

- **Test Strategy**: [TEST_STRATEGY.md](../TEST_STRATEGY.md) - Overall testing approach
- **E2E Testing**: [tests/e2e/README.md](../tests/e2e/README.md) - Playwright setup and usage
- **Coverage Scripts**: [scripts/validate-coverage.js](../scripts/validate-coverage.js) - Coverage validation logic
- **Makefile Targets**: [ollama-distributed/Makefile](../ollama-distributed/Makefile) - Go test commands

### External Documentation

- **Jest**: https://jestjs.io/docs/getting-started
- **Playwright**: https://playwright.dev/docs/intro
- **Go Testing**: https://golang.org/pkg/testing/
- **Coverage Tools**: https://go.dev/blog/cover

---

## Pre-Commit Checklist

Before committing code, ensure:

- [ ] All tests pass locally (`./scripts/run-all-tests.sh`)
- [ ] Coverage is ≥90% (Go and JavaScript)
- [ ] New features have corresponding tests
- [ ] Tests follow naming conventions
- [ ] No hardcoded credentials or secrets in tests
- [ ] E2E tests pass in headless mode
- [ ] Performance benchmarks show no regressions
- [ ] CI pipeline passes on PR
- [ ] Fake timers used for time-dependent tests
- [ ] Integration coverage included in merged reports

**Happy Testing! 🧪**

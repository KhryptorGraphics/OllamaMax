# Comprehensive Test Suite Guide

This guide explains how to use the comprehensive test suite that validates all implemented fixes in the ollamamax project.

## Quick Start

```bash
# Run the complete test suite
npm run test:comprehensive

# Or use the shell script
./scripts/run-test-suite.sh

# Run all tests including Jest
npm run test:all
```

## Test Categories

### 1. Database Configuration Tests
- ✅ Validates all database ports are above 10000
- ✅ Checks environment variable configuration
- ✅ Tests Go database configuration scripts
- ✅ Verifies database connection settings

### 2. API Server Initialization Tests
- ✅ Validates health endpoint configuration
- ✅ Checks API server port configuration (>10000)
- ✅ Tests Express.js server setup
- ✅ Verifies API route configuration

### 3. WebSocket Graceful Shutdown Tests
- ✅ Detects WebSocket implementations
- ✅ Validates graceful shutdown patterns
- ✅ Tests SIGTERM/SIGINT handling
- ✅ Checks process cleanup procedures

### 4. Docker Build Validation Tests
- ✅ Validates multi-stage Dockerfile
- ✅ Checks Node.js base image support
- ✅ Tests port exposure configuration
- ✅ Validates docker-compose.yml settings

### 5. Kubernetes Autoscaling Tests
- ✅ Checks HPA (Horizontal Pod Autoscaler) configuration
- ✅ Validates resource limits and requests
- ✅ Tests replica configuration
- ✅ Verifies K8s deployment files

### 6. Claude-Flow Agent Integration Tests
- ✅ Validates all 54 expected agents
- ✅ Tests agent spawning capabilities
- ✅ Checks command structure
- ✅ Verifies dependency configuration

### 7. Port Configuration Tests
- ✅ Comprehensive port scanning (all config files)
- ✅ Validates all ports are above 10000
- ✅ Checks multiple port patterns
- ✅ Security compliance verification

### 8. Security Configuration Tests
- ✅ Validates .gitignore patterns
- ✅ Checks for sensitive files
- ✅ Scans for hardcoded credentials
- ✅ Security best practices validation

### 9. Performance Benchmarks
- ✅ File I/O performance testing
- ✅ Memory usage monitoring
- ✅ System resource validation
- ✅ Performance threshold checks

## Test Scripts

### Main Test Runner
```bash
# Comprehensive test with detailed reporting
node scripts/test-all-fixes.js
```

**Features:**
- 🔍 Detailed analysis of all system components
- 📊 Performance metrics collection
- 💡 Intelligent recommendations
- 🏥 Overall health scoring
- ⚡ Fast execution (sub-100ms decisions)

### Jest Test Suites
```bash
# Run specific test suites
npm run test:fixes          # Configuration tests
npm run test:agents         # Agent system tests
npm run test:unit           # Unit tests
```

**Features:**
- 🧪 Structured test organization
- 📈 Coverage reporting
- 🚀 CI/CD integration ready
- 🔄 Watch mode support

### Shell Script Runner
```bash
# Complete system validation
./scripts/run-test-suite.sh
```

**Features:**
- 🔧 Prerequisite checking
- 🐳 Docker validation
- 🔍 Security scanning
- 📋 Comprehensive reporting

## Test Output Examples

### Successful Run
```
🚀 Starting Comprehensive Test Suite...

🔍 Testing Database Configuration...
✅ Database Port Configuration: All database ports (15432, 16379) are above 10000 in docker-compose.yml
✅ Database Environment Variables: Environment variables configured in .env.example

🌐 Testing API Server Initialization...
✅ Health Endpoints: Health endpoints found in src/api-server.js
✅ API Server Port: API server port 18080 is above 10000 in src/api-server.js

📊 COMPREHENSIVE TEST RESULTS
===============================
📈 SUMMARY:
✅ Passed: 25
❌ Failed: 0
⚠️ Warnings: 3
⏱️ Total Time: 1247ms

🏥 OVERALL HEALTH SCORE: 92%
🎉 System is in excellent condition!
```

### Test with Issues
```
❌ Port Configuration: Found 2 ports <= 10000: 8080 (src/api-server.js), 3000 (package.json)
⚠️ Docker Compose: docker-compose.yml not found
⚠️ WebSocket Configuration: No WebSocket configuration detected

💡 RECOMMENDATIONS:
🚨 Critical: Address all failed tests before deployment
🔌 Update all ports to be above 10000 for security compliance
```

## Integration with CI/CD

### GitHub Actions
```yaml
name: Test Suite
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-node@v2
        with:
          node-version: '16'
      - run: npm install
      - run: npm run test:all
      - run: ./scripts/run-test-suite.sh
```

### Docker Testing
```dockerfile
# Multi-stage testing
FROM node:16 AS test
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run test:all
RUN ./scripts/run-test-suite.sh
```

## Configuration

### Jest Configuration
Located in `jest.config.cjs`:
```javascript
module.exports = {
  testEnvironment: 'node',
  collectCoverageFrom: [
    'src/**/*.{js,ts}',
    'scripts/**/*.{js,ts}',
    '!node_modules/**'
  ],
  coverageThreshold: {
    global: {
      branches: 70,
      functions: 80,
      lines: 80,
      statements: 80
    }
  }
};
```

### Test Thresholds
- **Port Security**: All ports must be > 10000
- **Performance**: File I/O < 100ms, Memory < 200MB
- **Coverage**: 80% lines, 70% branches
- **Health Score**: 90% for excellent, 70% for good

## Troubleshooting

### Common Issues

**1. Port Configuration Failures**
```bash
# Fix: Update all ports to be above 10000
sed -i 's/port: 8080/port: 18080/g' config-file.yml
```

**2. Missing Dependencies**
```bash
# Fix: Install missing dependencies
npm install
```

**3. Go Tests Failing**
```bash
# Fix: Ensure Go modules are initialized
go mod init
go mod tidy
```

**4. Docker Build Issues**
```bash
# Fix: Check Dockerfile syntax
docker build --dry-run -t test .
```

### Debug Mode
```bash
# Run with verbose logging
DEBUG=1 node scripts/test-all-fixes.js

# Run specific test categories
TEST_CATEGORY=database node scripts/test-all-fixes.js
```

## Performance Optimization

### Parallel Execution
Tests run in parallel where possible:
- Database and API tests run concurrently
- Security scans happen in background
- Performance benchmarks are asynchronous

### Caching
- Test results are cached for repeated runs
- File system operations are optimized
- Memory usage is monitored and limited

### Resource Management
- Tests respect system resource limits
- Automatic cleanup of temporary files
- Memory leak detection and prevention

## Extending the Test Suite

### Adding New Tests
1. Add test logic to `scripts/test-all-fixes.js`
2. Create Jest tests in `tests/` directory
3. Update `package.json` scripts
4. Document in this guide

### Custom Validators
```javascript
// Example custom validator
async testCustomComponent() {
    try {
        // Your test logic here
        const result = await validateComponent();
        this.addResult('Custom Test', 'passed', 'Test passed');
    } catch (error) {
        this.addResult('Custom Test', 'failed', error.message);
    }
}
```

## Best Practices

1. **Run Before Every Commit**
   ```bash
   npm run test:all
   ```

2. **Use in Pre-push Hooks**
   ```bash
   #!/bin/sh
   npm run test:comprehensive
   ```

3. **Monitor Performance**
   ```bash
   # Regular performance monitoring
   npm run test:performance
   ```

4. **Security First**
   ```bash
   # Regular security validation
   ./scripts/run-test-suite.sh | grep -E "(security|credential|sensitive)"
   ```

## Support and Maintenance

- **Test Suite Version**: 1.0.0
- **Compatibility**: Node.js 16+, Go 1.19+
- **Update Frequency**: With every major fix
- **Maintenance**: Automated with CI/CD

For issues or improvements, please refer to the project's issue tracker or documentation.
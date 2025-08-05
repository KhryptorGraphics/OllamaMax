# OllamaMax Integration Testing Guide

## 🎯 Overview

This guide covers the comprehensive integration testing framework for OllamaMax, ensuring all components work together correctly in real-world scenarios.

## 🧪 Test Suite Components

### 1. **End-to-End Tests**
- **Complete system workflow** from CLI to distributed system
- **Real server startup** and API interaction
- **User journey validation** from discovery to automation

### 2. **Integration Test Framework**
- **Reusable test utilities** for consistent testing
- **Server lifecycle management** (start, ready check, cleanup)
- **Performance and stress testing** capabilities

### 3. **User Workflow Tests**
- **Discovery workflow**: How users find proxy commands
- **Monitoring workflow**: Status, instances, metrics checking
- **Automation workflow**: JSON output and scripting

## 🚀 Quick Start

### Run All Tests
```bash
# Make script executable
chmod +x scripts/run-integration-tests.sh

# Run complete test suite
./scripts/run-integration-tests.sh

# Run with verbose output
VERBOSE=true ./scripts/run-integration-tests.sh
```

### Run Specific Tests
```bash
# Integration tests only
go test ./tests/integration -run TestComprehensiveIntegration

# User workflow tests
go test ./tests/integration -run TestUserWorkflows

# Performance benchmarks
go test ./tests/integration -bench=. -benchtime=10s
```

## 📋 Test Categories

### 1. **Basic Functionality Tests**
```bash
# CLI help and command discovery
./ollama-distributed --help
./ollama-distributed proxy --help

# Command validation
./ollama-distributed proxy status --help
./ollama-distributed proxy instances --help
./ollama-distributed proxy metrics --help
```

### 2. **Proxy Command Tests**
```bash
# Status monitoring
./ollama-distributed proxy status --api-url http://localhost:8080
./ollama-distributed proxy status --json

# Instance management
./ollama-distributed proxy instances --api-url http://localhost:8080
./ollama-distributed proxy instances --json

# Performance metrics
./ollama-distributed proxy metrics --api-url http://localhost:8080
./ollama-distributed proxy metrics --json
```

### 3. **Error Handling Tests**
```bash
# Invalid commands
./ollama-distributed invalid-command
./ollama-distributed proxy invalid-subcommand

# Network errors
./ollama-distributed proxy status --api-url http://invalid:9999

# Missing parameters
./ollama-distributed proxy status --api-url
```

### 4. **JSON Output Validation**
```bash
# Validate JSON structure
./ollama-distributed proxy status --json | jq '.'
./ollama-distributed proxy instances --json | jq '.instances'
./ollama-distributed proxy metrics --json | jq '.metrics'
```

## 🔧 Test Environment Setup

### Prerequisites
```bash
# 1. Build the binary
go build -o ollama-distributed ./cmd/node

# 2. Verify Go environment
./scripts/verify-go-env.sh

# 3. Install test dependencies
go mod download
```

### Manual Setup
```bash
# Start server manually
./ollama-distributed start --port 8080 &

# Wait for server to be ready
curl http://localhost:8080/health

# Run tests
go test ./tests/integration

# Cleanup
pkill -f "ollama-distributed.*start"
```

### Docker Setup
```bash
# Build and test in Docker
docker-compose -f docker-compose.build.yml up ollama-test

# Development environment
docker-compose -f docker-compose.build.yml up -d ollama-dev
docker exec -it ollama-dev bash
```

## 📊 Test Results Interpretation

### Success Indicators
```
✅ All tests passed! 🎉
✅ Integration tests passed
✅ User workflow tests passed
✅ Performance tests passed
Overall Success Rate: 3/3 (100%)
```

### Failure Analysis
```
❌ Integration tests failed
⚠️  Performance tests had issues
Overall Success Rate: 1/3 (33%)
```

**Common Failure Causes:**
1. **Binary not found**: Run `go build ./cmd/node` first
2. **Server startup failed**: Check port availability (8080)
3. **Network issues**: Verify localhost connectivity
4. **Timeout errors**: Increase test timeout values

## 🏃 Performance Testing

### Command Performance
```bash
# Test command execution speed
go test ./tests/integration -run TestPerformance

# Benchmark specific commands
go test ./tests/integration -bench=BenchmarkProxyCommands
```

### Stress Testing
```bash
# Concurrent command execution
go test ./tests/integration -run StressTest

# Custom stress test
for i in {1..100}; do
  ./ollama-distributed proxy --help &
done
wait
```

### Performance Expectations
- **CLI Help Commands**: < 100ms
- **API Status Calls**: < 500ms
- **JSON Processing**: < 200ms
- **Concurrent Execution**: 10+ commands/second

## 🔍 Debugging Tests

### Verbose Output
```bash
# Detailed test output
go test -v ./tests/integration

# With race detection
go test -race ./tests/integration

# With coverage
go test -cover ./tests/integration
```

### Manual Debugging
```bash
# Start server with debug logging
./ollama-distributed start --log-level debug

# Test individual commands
./ollama-distributed proxy status --api-url http://localhost:8080

# Check API directly
curl http://localhost:8080/api/v1/proxy/status
```

### Log Analysis
```bash
# Server logs
tail -f /var/log/ollama-distributed.log

# Test logs
go test ./tests/integration 2>&1 | tee test-output.log

# API logs
curl -v http://localhost:8080/health
```

## 🎯 Test Scenarios

### Scenario 1: New User Experience
1. User runs `./ollama-distributed --help`
2. Discovers `proxy` command
3. Explores `./ollama-distributed proxy --help`
4. Tries `./ollama-distributed proxy status`

### Scenario 2: System Monitoring
1. Admin starts system: `./ollama-distributed start`
2. Checks status: `./ollama-distributed proxy status`
3. Lists instances: `./ollama-distributed proxy instances`
4. Monitors metrics: `./ollama-distributed proxy metrics --watch`

### Scenario 3: Automation Integration
1. Script gets JSON status: `./ollama-distributed proxy status --json`
2. Processes with jq: `| jq '.status'`
3. Alerts on issues: `if [ "$status" != "running" ]; then alert; fi`

## 📚 Test Documentation

### Test Files
- **`end_to_end_test.go`**: Complete system testing
- **`integration_test_framework.go`**: Reusable test utilities
- **`comprehensive_integration_test.go`**: Full test suite

### Test Scripts
- **`run-integration-tests.sh`**: Main test runner
- **`verify-go-env.sh`**: Environment verification
- **`setup-build-env.sh`**: Environment setup

### Documentation
- **`INTEGRATION_TESTING.md`**: This guide
- **`BUILD_INSTRUCTIONS.md`**: Build and setup guide
- **`CLI_REFERENCE.md`**: Command-line reference

## ✅ Success Criteria

### Critical Tests (Must Pass)
- [ ] Binary builds successfully
- [ ] Server starts and responds to health checks
- [ ] All proxy CLI commands work
- [ ] JSON output is valid
- [ ] Error handling works correctly

### Quality Tests (Should Pass)
- [ ] Performance meets expectations
- [ ] Stress tests complete successfully
- [ ] User workflows are intuitive
- [ ] Documentation examples work

### Production Readiness
- [ ] All critical tests pass
- [ ] Performance is acceptable
- [ ] Error messages are user-friendly
- [ ] Documentation is complete and accurate

## 🚀 Next Steps

After successful integration testing:

1. **Deploy to staging**: Test in staging environment
2. **User acceptance testing**: Get feedback from real users
3. **Performance tuning**: Optimize based on test results
4. **Production deployment**: Deploy to production systems
5. **Monitoring setup**: Implement production monitoring

The integration testing framework ensures OllamaMax is ready for real-world use with confidence in its reliability and performance.

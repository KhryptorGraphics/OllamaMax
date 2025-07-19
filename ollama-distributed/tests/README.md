# Ollama Distributed - Comprehensive Test Suite

This directory contains a comprehensive test suite for the Ollama Distributed system, designed to ensure reliability, performance, and fault tolerance across all components.

## Test Architecture

The test suite is organized into five main categories:

### 1. Unit Tests (`/tests/unit/`)
**Purpose**: Test individual components in isolation
**Speed**: Fast (< 5 minutes)
**Scope**: Single components, functions, and classes

- **discovery_test.go**: P2P discovery engine tests
- **scheduler_test.go**: Distributed scheduler tests
- **model_sync_test.go**: Model synchronization tests
- **api_test.go**: API layer tests

### 2. Integration Tests (`/tests/integration/`)
**Purpose**: Test component interactions and workflows
**Speed**: Medium (5-15 minutes)
**Scope**: Component interactions, data flow, basic clustering

- **distributed_inference_test.go**: Distributed inference workflows
- **cluster_test.go**: Basic cluster functionality
- **test_cluster.go**: Test cluster infrastructure

### 3. End-to-End Tests (`/tests/e2e/`)
**Purpose**: Test complete user workflows
**Speed**: Slow (15-45 minutes)
**Scope**: Full system workflows, user scenarios

- **complete_workflow_test.go**: Complete user workflows from start to finish

### 4. Performance Tests (`/tests/performance/`)
**Purpose**: Benchmark and load testing
**Speed**: Very Slow (30-90 minutes)
**Scope**: Performance characteristics, scalability limits

- **benchmark_test.go**: Comprehensive performance benchmarks

### 5. Chaos Tests (`/tests/chaos/`)
**Purpose**: Fault injection and resilience testing
**Speed**: Very Slow (30-90 minutes)
**Scope**: System resilience, fault tolerance, edge cases

- **chaos_test.go**: General chaos engineering tests
- **byzantine_test.go**: Byzantine failure scenarios

## Running Tests

### Prerequisites

```bash
# Install dependencies
make deps deps-dev deps-test

# Setup test environment
make setup-test-env
```

### Quick Testing

```bash
# Run unit tests only (fast)
make test-unit

# Run smoke tests
make smoke-test

# Run quick tests
make quick-test
```

### Comprehensive Testing

```bash
# Run all tests except performance and chaos
make test-all

# Run complete test suite including performance and chaos
make test-full

# Run specific test categories
make test-p2p
make test-consensus  
make test-scheduler
make test-models
make test-api
```

### CI/CD Testing

```bash
# Run CI test suite
make test-ci

# Run test runner in CI mode
make test-runner-ci
```

### Performance Testing

```bash
# Run performance tests
make test-performance

# Run benchmarks
make test-bench

# Run load balancing tests
make test-load-balancing

# Run scalability tests
make test-scalability
```

### Chaos Testing

```bash
# Run chaos tests
make test-chaos

# Run Byzantine failure tests
make test-byzantine

# Run fault tolerance tests
make test-fault-tolerance
```

## Test Configuration

### Environment Variables

- `OLLAMA_TEST_NODE_COUNT`: Number of nodes in test cluster (default: 3)
- `OLLAMA_TEST_TIMEOUT`: Test timeout duration (default: 30m)
- `OLLAMA_TEST_ARTIFACTS_DIR`: Directory for test artifacts (default: ./test-artifacts)
- `OLLAMA_TEST_CI`: Enable CI mode (default: false)
- `OLLAMA_TEST_USE_DOCKER`: Use Docker for testing (default: false)

### Test Flags

- `-v`: Verbose output
- `-race`: Enable race detection
- `-short`: Run only short tests
- `-timeout`: Set test timeout
- `-parallel`: Set parallel test count
- `-bench`: Run benchmarks
- `-coverprofile`: Generate coverage report

### Examples

```bash
# Run unit tests with race detection
go test -v -race ./tests/unit/...

# Run integration tests with custom timeout
OLLAMA_TEST_TIMEOUT=45m go test -v -timeout=45m ./tests/integration/...

# Run performance tests with 5 nodes
OLLAMA_TEST_NODE_COUNT=5 go test -v -bench=. ./tests/performance/...

# Run chaos tests with Docker
OLLAMA_TEST_USE_DOCKER=true go test -v -tags=chaos ./tests/chaos/...
```

## Test Infrastructure

### TestCluster

The `TestCluster` in `/tests/integration/test_cluster.go` provides a unified testing infrastructure:

```go
// Create a test cluster with 3 nodes
cluster, err := integration.NewTestCluster(3)
require.NoError(t, err)
defer cluster.Shutdown()

// Start the cluster
err = cluster.Start()
require.NoError(t, err)

// Get cluster components
leader := cluster.GetLeader()
nodes := cluster.GetActiveNodes()
```

### Test Utilities

- **Mock Services**: Comprehensive mocking for external dependencies
- **Test Data**: Realistic test data generators
- **Assertions**: Custom assertions for distributed systems
- **Helpers**: Utility functions for common test patterns

## Key Test Scenarios

### 1. Basic Functionality
- Node startup and shutdown
- P2P discovery and connection
- Consensus operations
- Model loading and inference
- API endpoints

### 2. Distributed Operations
- Multi-node inference
- Model synchronization
- Load balancing
- Fault detection and recovery
- Leader election

### 3. Performance Characteristics
- Throughput benchmarks
- Latency measurements  
- Memory usage profiling
- Network overhead analysis
- Scalability testing

### 4. Fault Tolerance
- Node failures
- Network partitions
- Resource exhaustion
- Byzantine failures
- Cascading failures

### 5. Edge Cases
- Concurrent operations
- Resource limits
- Configuration errors
- Network issues
- Data corruption

## Coverage Requirements

- **Unit Tests**: > 90% line coverage
- **Integration Tests**: > 80% path coverage
- **E2E Tests**: > 70% user scenario coverage
- **Performance Tests**: Key metrics benchmarked
- **Chaos Tests**: Critical failure modes tested

## Performance Benchmarks

### Baseline Expectations

- **Inference Latency**: < 5s for small models, < 30s for large models
- **Throughput**: > 10 req/s for concurrent requests
- **Memory Usage**: < 4GB per node for standard configurations
- **Network Overhead**: < 10% of total bandwidth
- **Startup Time**: < 30s for cluster initialization

### Scalability Targets

- **Node Count**: Support 10+ nodes in cluster
- **Model Size**: Handle models up to 70B parameters
- **Concurrent Requests**: Support 100+ concurrent requests
- **Network Partitions**: Survive 30% node failures
- **Data Consistency**: Maintain consistency across all scenarios

## Continuous Integration

The test suite integrates with GitHub Actions for automated testing:

- **Unit Tests**: Run on every PR and push
- **Integration Tests**: Run on every PR and push
- **E2E Tests**: Run on every PR and push
- **Performance Tests**: Run nightly or on-demand
- **Chaos Tests**: Run nightly or on-demand

### Test Matrix

- **Go Versions**: 1.21, 1.22
- **Operating Systems**: Linux, macOS, Windows
- **Architectures**: amd64, arm64
- **Configurations**: Various cluster sizes and configurations

## Test Artifacts

All test runs generate artifacts in `./test-artifacts/`:

- **Logs**: Detailed test execution logs
- **Coverage**: HTML coverage reports
- **Benchmarks**: Performance benchmark results
- **Metrics**: System metrics during tests
- **Screenshots**: E2E test screenshots (if applicable)

## Debugging Tests

### Debug Mode

```bash
# Run tests with debugging
make test-debug

# Run with tracing
make test-trace

# Run with profiling
make test-profile
```

### Log Analysis

```bash
# View recent test logs
make test-logs

# Check test status
make test-status

# View test artifacts
make test-artifacts
```

### Interactive Debugging

```bash
# Run specific test with debugging
go test -v -run=TestSpecificTest ./tests/unit/... -args -test.debug=true

# Run with delve debugger
dlv test ./tests/unit/... -- -test.run=TestSpecificTest
```

## Contributing to Tests

### Test Guidelines

1. **Test Naming**: Use descriptive test names that explain the scenario
2. **Test Structure**: Follow Arrange-Act-Assert pattern
3. **Test Data**: Use realistic test data
4. **Assertions**: Use meaningful assertion messages
5. **Cleanup**: Always clean up resources

### Adding New Tests

1. Choose the appropriate test category
2. Follow existing patterns and conventions
3. Add test documentation
4. Update CI configuration if needed
5. Ensure proper cleanup and resource management

### Test Review Checklist

- [ ] Test covers the intended functionality
- [ ] Test is deterministic and repeatable
- [ ] Test includes proper error handling
- [ ] Test cleans up resources properly
- [ ] Test documentation is clear and complete
- [ ] Test follows project conventions

## Common Issues and Solutions

### Test Flakiness

- **Root Cause**: Timing issues, resource contention, external dependencies
- **Solution**: Add proper timeouts, use test-specific resources, mock external services

### Resource Exhaustion

- **Root Cause**: Tests not cleaning up resources
- **Solution**: Use defer statements, implement proper cleanup, limit resource usage

### Network Issues

- **Root Cause**: Port conflicts, network configuration
- **Solution**: Use dynamic port allocation, proper network isolation

### CI/CD Failures

- **Root Cause**: Environment differences, resource limits
- **Solution**: Use consistent environments, proper resource limits, retry logic

## Monitoring and Metrics

### Test Metrics

- Test execution time
- Test success/failure rates
- Coverage percentages
- Resource utilization
- Performance benchmarks

### Alerting

- Test failures in CI/CD
- Performance regressions
- Coverage drops
- Resource exhaustion

## Future Enhancements

### Planned Improvements

1. **Property-Based Testing**: Add property-based tests for edge cases
2. **Mutation Testing**: Add mutation testing for test quality
3. **Visual Testing**: Add visual regression testing for UI components
4. **Security Testing**: Enhanced security and penetration testing
5. **Load Testing**: More comprehensive load testing scenarios

### Test Automation

1. **Auto-generated Tests**: Generate tests from API specifications
2. **Smart Test Selection**: Run only relevant tests based on changes
3. **Test Parallelization**: Improve test execution speed
4. **Test Orchestration**: Better test dependency management

## Resources

### Documentation

- [Go Testing Documentation](https://golang.org/doc/testing)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Chaos Engineering Principles](https://principlesofchaos.org/)

### Tools

- [golangci-lint](https://golangci-lint.run/): Linting and static analysis
- [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck): Security vulnerability scanning
- [pprof](https://golang.org/pkg/net/http/pprof/): Performance profiling
- [race detector](https://golang.org/doc/articles/race_detector.html): Race condition detection

### Best Practices

- [Google Testing Blog](https://testing.googleblog.com/)
- [Effective Go Testing](https://golang.org/doc/effective_go.html#testing)
- [Test-Driven Development](https://en.wikipedia.org/wiki/Test-driven_development)

---

This comprehensive test suite ensures the reliability, performance, and fault tolerance of the Ollama Distributed system. Regular testing and continuous improvement of the test suite are essential for maintaining system quality and reliability.
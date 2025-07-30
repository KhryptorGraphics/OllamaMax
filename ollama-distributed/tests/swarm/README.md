# Swarm Operations Test Suite

## Overview

This comprehensive test suite validates swarm operations, coordination mechanisms, file operations, and performance characteristics for the Ollama Distributed System. It provides a complete testing framework for ensuring the reliability and performance of swarm-based coordination.

## Test Components

### 1. Swarm Operations Tests (`swarm_operations_test.go`)
- **SwarmTestHarness**: Complete test infrastructure for swarm operations
- **Initialization Testing**: Multiple topology support (mesh, hierarchical, star, ring)
- **Scaling Tests**: Dynamic agent scaling under load
- **Coordination Tests**: Message passing, task distribution, load balancing
- **Fault Recovery**: Agent failure simulation and replacement
- **Performance Tests**: Throughput, memory efficiency, network latency
- **Security Tests**: Authentication, encryption, access control

### 2. Validation Framework (`validation_routines.go`)
- **ValidationSuite**: Comprehensive validation management
- **Multiple Validators**:
  - SwarmHealthValidator: Overall swarm connectivity and health
  - AgentCoordinationValidator: Message passing and coordination
  - PerformanceValidator: Resource utilization and response times
  - SecurityValidator: Authentication and encryption
  - DataIntegrityValidator: Message and task integrity
- **Configurable Validation Levels**: Basic, Standard, Strict, Paranoid
- **Parallel/Sequential Execution**: Optimized for performance
- **Detailed Reporting**: Comprehensive validation reports

### 3. File Operations Coordination (`file_operations_coordination_test.go`)
- **FileOperationCoordinator**: Centralized file operation management
- **FileLockManager**: Distributed file locking with read/write/exclusive modes
- **FileChangeTracker**: Change tracking and conflict detection
- **BackupManager**: Automatic backup and recovery
- **Supported Operations**: Create, Update, Delete, Read, Append, Copy, Move
- **Concurrency Testing**: Multi-agent file operations
- **Performance Testing**: High-concurrency scenarios
- **Memory Efficiency**: Resource usage validation

### 4. Performance Measurement Tools (`performance_measurement_tools.go`)
- **PerformanceMeter**: Comprehensive performance monitoring
- **Multiple Collectors**:
  - TimingCollector: Operation latency and response times
  - ThroughputCollector: Operations per second and success rates
  - ResourceCollector: Memory, CPU, and goroutine monitoring
  - NetworkCollector: Network latency and communication metrics
  - CoordinationCollector: Lock contention and coordination events
- **Real-time Monitoring**: Continuous performance tracking
- **Statistical Analysis**: Percentiles, averages, min/max calculations
- **Atomic Operations**: Thread-safe metrics collection

### 5. Integration Tests (`integration_test.go`)
- **SwarmIntegrationTestSuite**: End-to-end testing framework
- **Complete Workflow Testing**: Full swarm lifecycle validation
- **Failure Recovery Testing**: Comprehensive fault tolerance validation
- **Security Validation**: Authentication, encryption, access control
- **Performance Under Load**: Multi-scenario load testing
- **Benchmark Suite**: Performance benchmarking utilities

### 6. Test Runner (`run_swarm_tests.go`)
- **SwarmTestRunner**: Orchestrated test execution
- **Configurable Execution**: Selective test suite execution
- **Coverage Reporting**: Integrated code coverage analysis
- **Parallel Execution**: Optimized test performance
- **Comprehensive Reporting**: Detailed markdown reports
- **Artifact Management**: Test logs and coverage files

## Usage

### Running All Tests
```bash
# Build the test runner
go build -o swarm_test_runner run_swarm_tests.go

# Run all tests with coverage
./swarm_test_runner -coverage -v

# Run specific test types
./swarm_test_runner -integration -v
./swarm_test_runner -performance -v
./swarm_test_runner -validation -v
```

### Running Individual Test Suites
```bash
# Swarm operations tests
go test -v ./swarm_operations_test.go

# File operations coordination
go test -v ./file_operations_coordination_test.go

# Integration tests
go test -v ./integration_test.go
```

### Configuration Options

The test runner supports extensive configuration:

```bash
./swarm_test_runner [options]

Options:
  -v, --verbose          Verbose output
  -coverage              Enable coverage reporting (default: true)
  -bench                 Run benchmarks
  -integration           Run only integration tests
  -performance           Run only performance tests
  -validation            Run only validation tests
  -parallel              Enable parallel execution (default: true)
  -race                  Enable race detection (default: true)
  -timeout duration      Global timeout (default: 30m)
  -output string         Output directory (default: ./swarm-test-output)
  -workers int           Max parallel workers (default: CPU count)
```

## Test Architecture

### Coordination Hooks Integration
All tests integrate with Claude Flow coordination hooks:
- **pre-task**: Initialize coordination context
- **post-edit**: Track file operations and store progress
- **notify**: Custom notifications for test events
- **post-task**: Complete task coordination and analyze performance

### Memory Coordination
Tests store coordination data in shared memory:
- Implementation progress tracking
- Test results and metrics
- Coordination decisions and findings
- Performance baselines and trends

### Performance Baselines

Expected performance characteristics:
- **Throughput**: Minimum 10 operations/second under normal load
- **Response Time**: Average under 1 second for standard operations
- **Memory Usage**: Under 80% of configured threshold
- **Network Latency**: Average under 100ms for coordination messages
- **Error Rate**: Under 5% for all operations
- **Coverage**: Minimum 80% code coverage across test suites

### Security Testing

Comprehensive security validation:
- **Authentication**: Valid/invalid credential testing
- **Authorization**: Role-based access control validation
- **Encryption**: Message encryption/decryption testing
- **Integrity**: Message tampering detection
- **Access Control**: Operation permission validation

## Output and Reporting

### Test Artifacts
- **Logs**: Individual test suite logs in `{output-dir}/*.log`
- **Coverage**: Coverage reports in `{output-dir}/*_coverage.out`
- **Reports**: Detailed markdown report in `{output-dir}/swarm_test_report.md`
- **Performance**: Performance metrics and trends

### Report Format
```
📊 SWARM TEST REPORT
====================
TEST SUITE               STATUS     DURATION     COVERAGE   ERROR
------------------------------------------------------------------------
swarm_operations         ✅ PASS    2m15s        85.2%      
validation_routines      ✅ PASS    1m32s        92.1%      
file_operations_coord    ✅ PASS    3m08s        78.9%      
performance_measurement  ✅ PASS    1m45s        88.7%      
integration             ✅ PASS    4m22s        N/A        
------------------------------------------------------------------------

📈 SUMMARY:
   Total Test Suites: 5
   Successful: 5 (100.0%)
   Failed: 0 (0.0%)
   Total Duration: 12m42s
   Average Coverage: 86.2%
```

## Troubleshooting

### Common Issues

1. **Build Failures**: Ensure Go modules are properly initialized with `go mod tidy`
2. **Memory Issues**: Adjust memory thresholds in test configuration
3. **Timeout Issues**: Increase timeout values for slower systems
4. **Race Conditions**: Use `-race=false` if encountering race detection issues
5. **Coverage Issues**: Ensure all source files are in GOPATH

### Debug Mode
Run with maximum verbosity for debugging:
```bash
./swarm_test_runner -v -timeout=60m -workers=1
```

### Performance Tuning
For optimal performance on different systems:
- **Low-resource systems**: Use `-workers=2 -parallel=false`
- **High-performance systems**: Use `-workers=8 -parallel=true`
- **CI/CD environments**: Use `-timeout=45m -race=false`

## Integration with Claude Flow

This test suite is designed to work seamlessly with Claude Flow coordination:

1. **Initialization**: Tests use `pre-task` hooks to establish coordination context
2. **Execution**: Tests store progress using `post-edit` hooks and memory operations
3. **Reporting**: Tests use `notify` hooks for status updates
4. **Completion**: Tests use `post-task` hooks for performance analysis

### Memory Storage Keys
- `hive/coder/implementation_plan`: Initial implementation strategy
- `hive/coder/validation_implementation`: Validation framework progress
- `hive/coder/integration_tests`: Integration test implementation
- `hive/coder/final_implementation`: Complete deliverables summary

## Contributing

When adding new tests:
1. Follow the established patterns for test harnesses and coordination
2. Include comprehensive validation and performance measurement
3. Integrate with Claude Flow hooks for coordination
4. Update this README with new test descriptions
5. Ensure proper error handling and cleanup
6. Add appropriate benchmarks for performance-critical code

## License

This test suite is part of the Ollama Distributed System and follows the same licensing terms.
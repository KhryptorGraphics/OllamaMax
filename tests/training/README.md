# Training Tests README

## Overview

This directory contains comprehensive tests for the OllamaMax Distributed Training System, covering all 5 training modules plus certification assessment.

## Test Files

### Go Test Files

#### `training_test_suite.go` (868 lines)
Complete test suite for all 5 modules plus certification
- Module 1: Installation and Setup
- Module 2: Configuration Management
- Module 3: Basic Operations
- Module 4: API Integration  
- Module 5: Validation and Testing
- Certification Assessment

**Key Features:**
- Comprehensive test result tracking
- JSON result export
- Module completion recording
- Error handling and reporting

#### `training_module_tests.go` (670 lines)
Module-specific tests with learning objectives
- Prerequisites validation
- Build validation
- Configuration structure
- Network configuration
- Cluster operations concepts
- API architecture understanding
- Performance benchmarking

**Key Features:**
- Learning objective validation
- Practical skills testing
- Concept verification
- Certification readiness checks

#### `training_performance_benchmarks_test.go` (754 lines)
Performance benchmarks for all training scenarios
- Module execution performance
- API endpoint performance
- Configuration loading benchmarks
- Validation script performance
- Concurrent user handling

**Key Features:**
- Detailed performance metrics
- Memory usage tracking
- Throughput measurement
- Scalability testing

#### `certification_tests.go` (843 lines)
Full certification assessment framework
- Prerequisites assessment
- Practical skills validation
- Knowledge assessment
- Hands-on exercises
- Final certification determination

**Key Features:**
- Comprehensive assessment coverage
- Skill verification
- Knowledge testing
- Pass/fail determination

### Bash Validation Scripts

#### `validation_scripts_enhanced.sh` (1711 lines)
Comprehensive bash validation with 7 test categories
1. Prerequisites check
2. Installation validation
3. Configuration tests
4. API validation
5. Security tests
6. Performance tests
7. Deployment checks

**Key Features:**
- Multi-phase validation
- Detailed logging
- Error handling
- Summary reporting

#### `integration_test.go`
Integration tests for training system components
- End-to-end workflows
- Component integration
- System-level validation

## Running Tests

### Run All Training Tests
```bash
# Using Make (recommended)
cd ollama-distributed
make test-training

# Using npm
npm run test:training

# Direct execution
bash scripts/run-training-tests.sh
```

### Run Module Tests Only
```bash
# Using Make
make test-training-modules

# Using npm
npm run test:training:modules

# Direct Go test
cd tests/training
go test -v -run TestTrainingModule ./...
```

### Run Performance Benchmarks
```bash
# Using Make
make test-training-performance

# Using npm
npm run test:training:performance

# Direct Go test
cd tests/training
go test -bench=. -benchmem ./...
```

### Run Certification Tests
```bash
# Using Make
make test-certification

# Using npm
npm run test:training:certification

# Direct Go test
cd tests/training
go test -v -run TestCertification ./...
```

### Run Validation Scripts
```bash
# Using Make
make test-training-validation

# Using npm
npm run test:training:validation

# Direct execution
bash tests/training/validation_scripts_enhanced.sh full
```

### Run Complete Suite with Reports
```bash
# Using Make (generates metrics + dashboard)
make test-training-all

# Using npm
npm run test:training:all
```

## Test Environment Setup

### Required Tools
- Go 1.21 or later
- Bash 4.0 or later
- curl
- jq
- git

### Environment Variables
```bash
export PROJECT_ROOT=/path/to/OllamaMax
export OLLAMA_PROJECT_ROOT=/path/to/OllamaMax
export TRAINING_ROOT=/path/to/OllamaMax/ollama-distributed/training
```

### Port Requirements
The tests may use these ports (configurable):
- 8080, 8081 - API servers
- 4001, 4002 - P2P networking
- 11434 - Ollama API

### Directory Structure
```
tests/training/
├── README.md                               # This file
├── training_test_suite.go                  # Main test suite
├── training_module_tests.go                # Module tests
├── training_performance_benchmarks_test.go # Performance benchmarks
├── certification_tests.go                  # Certification tests
├── validation_scripts_enhanced.sh          # Validation scripts
└── integration_test.go                     # Integration tests
```

## Test Coverage

### Target Coverage
- Overall: ≥ 90%
- Per-module: ≥ 85%

### Current Coverage
Run tests to measure:
```bash
cd tests/training
go test -v -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

### Coverage Report
Generate HTML report:
```bash
go tool cover -html=coverage.out -o coverage.html
open coverage.html
```

## Validation Checkpoints

### Module 1: Installation and Setup
- ✅ Binary build successful
- ✅ Help command works
- ✅ Version command works
- ✅ Go version compatible
- ✅ Disk space sufficient
- ✅ Ports available

### Module 2: Configuration Management
- ✅ Configuration directory created
- ✅ Configuration manager executes
- ✅ Profile files valid
- ✅ Configuration syntax correct
- ✅ Dry run successful

### Module 3: Basic Operations
- ✅ Health monitoring script works
- ✅ Service startup validated
- ✅ Multi-node configuration created
- ✅ Port conflict validation passes

### Module 4: API Integration
- ✅ API client compiles
- ✅ API client help works
- ✅ Custom tool template valid

### Module 5: Validation and Testing
- ✅ Validation suite compiles
- ✅ All validation categories covered
- ✅ Test extension framework available

### Certification
- ✅ System requirements met
- ✅ Training modules completed
- ✅ Build and deploy capabilities
- ✅ Configuration management skills
- ✅ Troubleshooting abilities

## Troubleshooting

### Common Issues

#### Path Configuration Problems
**Symptom:** Tests can't find project files
**Solution:** Set environment variables correctly
```bash
export PROJECT_ROOT=$(pwd)
export OLLAMA_PROJECT_ROOT=$(pwd)
```

#### Port Conflicts
**Symptom:** Tests fail with "port already in use"
**Solution:** Find and kill process using the port
```bash
lsof -i :8080
kill -9 <PID>
```

#### Permission Issues
**Symptom:** Tests fail with permission denied
**Solution:** Ensure scripts are executable
```bash
chmod +x tests/training/validation_scripts_enhanced.sh
chmod +x scripts/run-training-tests.sh
```

#### Test Timeout
**Symptom:** Tests hang or timeout
**Solution:** Increase timeout or check for blocking operations
```bash
go test -v -timeout 30m ./...
```

### Getting Help
- Check test output logs in `test-results/training/`
- Review validation script output
- Consult training documentation
- Open issue on GitHub

## Contributing

### Adding New Tests
1. Create test function with descriptive name
2. Use `ts.runTrainingTest()` for consistent result tracking
3. Add validation checkpoints
4. Update this README with new test information

### Test Naming Conventions
- Test functions: `TestTrainingModuleX_Feature`
- Benchmark functions: `BenchmarkTrainingModuleX_Operation`
- Helper functions: `test<Feature>` (lowercase start)

### Documentation Requirements
- Add function comments explaining purpose
- Document test prerequisites
- Include expected outcomes
- Note any special setup requirements

## Pull Request Process
1. Run all training tests locally
2. Ensure tests pass with ≥ 90% coverage
3. Update documentation if needed
4. Submit PR with test results

## Additional Resources

### Documentation
- [Training Implementation Summary](../../TRAINING_IMPLEMENTATION_SUMMARY.md)
- [Training Quality Dashboard](../../docs/TRAINING_QUALITY_DASHBOARD.md)
- Training quality metrics are generated dynamically in test-results/training/metrics.json
- Historical validation reports are available in test-results/training/

### Related Files
- [Makefile](../../ollama-distributed/Makefile) - Build and test targets
- [CI/CD Pipeline](../../.github/workflows/ci-cd-pipeline.yml) - Automated testing
- [package.json](../../package.json) - npm test scripts

---

**Last Updated:** 2025-10-27
**Maintainer:** Training Quality Team

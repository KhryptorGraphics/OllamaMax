# Training Exercise Templates for Ollama Distributed

This document provides structured exercise templates for each training module, complete with objectives, step-by-step instructions, validation criteria, and solution examples.

## Module 1: Installation and Setup Exercises

### Exercise 1.1: Complete Installation from Source

**Objective**: Build Ollama Distributed from source code and verify the installation.

**Prerequisites**: 
- Go 1.21+ installed
- Git available
- 2GB free disk space

**Time Estimate**: 15 minutes

**Instructions**:
1. Clone the repository (if not already done)
2. Navigate to the project directory
3. Build the main binary
4. Verify the build completed successfully
5. Test basic CLI commands

**Step-by-Step**:
```bash
# Step 1: Navigate to project
cd /home/kp/ollamamax/ollama-distributed

# Step 2: Build binary
go build -o bin/ollama-distributed ./cmd/node

# Step 3: Verify executable
ls -la bin/ollama-distributed

# Step 4: Test help command
./bin/ollama-distributed --help

# Step 5: Test version command
./bin/ollama-distributed --version
```

**Validation Criteria**:
- [ ] Binary file exists and is executable
- [ ] Help command displays available commands
- [ ] Version command returns version information
- [ ] No build errors occurred

**Common Issues & Solutions**:
- **Build fails**: Check Go version with `go version`
- **Permission denied**: Run `chmod +x bin/ollama-distributed`
- **Command not found**: Use full path `./bin/ollama-distributed`

**Extension Challenges**:
1. Build additional tools in the `cmd/` directory
2. Create a custom build script with error handling
3. Set up automated build with Make or CI/CD

---

### Exercise 1.2: Environment Validation

**Objective**: Validate that the system meets all requirements for running Ollama Distributed.

**Time Estimate**: 10 minutes

**Instructions**:
Run comprehensive system validation and resolve any issues found.

**Step-by-Step**:
```bash
# Use the provided installation script
chmod +x training/code-examples/01-installation/install-and-build.sh
./training/code-examples/01-installation/install-and-build.sh validate-only

# Or run manual checks:
go version                    # Check Go version
df -h .                      # Check disk space
free -h                      # Check memory
netstat -ln | grep -E "(8080|8081|4001)"  # Check ports
```

**Validation Criteria**:
- [ ] Go version 1.21 or higher
- [ ] At least 2GB free disk space
- [ ] At least 4GB RAM available
- [ ] Required ports (8080, 8081, 4001) are free
- [ ] Git and curl are available

**Troubleshooting Guide**:
- **Go version too old**: Install latest Go from https://golang.org/dl/
- **Insufficient disk space**: Clean up files or move to larger disk
- **Port conflicts**: Either stop conflicting services or use different ports
- **Missing tools**: Install using system package manager

---

## Module 2: Configuration Management Exercises

### Exercise 2.1: Create Custom Configuration Profiles

**Objective**: Create and validate custom configuration profiles for different environments.

**Time Estimate**: 20 minutes

**Instructions**:
1. Create development configuration
2. Create testing configuration
3. Create production template
4. Validate all configurations

**Step-by-Step**:
```bash
# Step 1: Create configuration directory
mkdir -p ~/.ollama-distributed/profiles

# Step 2: Run configuration manager
cd training/code-examples/02-configuration
go run configuration-manager.go

# Step 3: Verify profiles were created
ls -la ~/.ollama-distributed/profiles/

# Step 4: Examine configuration content
cat ~/.ollama-distributed/profiles/development.yaml
```

**Configuration Requirements**:
- **Development**: Debug enabled, local addresses, minimal security
- **Testing**: Production-like settings, isolated ports
- **Production**: Security enabled, proper TLS, restrictive settings

**Validation Criteria**:
- [ ] All three profile files exist
- [ ] YAML syntax is valid for each profile
- [ ] Port assignments don't conflict
- [ ] Security settings appropriate for environment
- [ ] Data directories are properly configured

**Advanced Challenges**:
1. Add environment-specific model configurations
2. Create profile inheritance system
3. Add configuration validation rules
4. Implement configuration hot-reloading

---

### Exercise 2.2: Configuration Validation and Testing

**Objective**: Validate configuration files and test different configuration scenarios.

**Time Estimate**: 15 minutes

**Step-by-Step**:
```bash
# Step 1: Validate configuration syntax
./bin/ollama-distributed validate --config ~/.ollama-distributed/profiles/development.yaml

# Step 2: Test dry run with configuration
./bin/ollama-distributed start --config ~/.ollama-distributed/profiles/development.yaml --dry-run

# Step 3: Compare configurations
diff ~/.ollama-distributed/profiles/development.yaml ~/.ollama-distributed/profiles/testing.yaml

# Step 4: Test port availability
netstat -ln | grep -E "(8080|9080|8081|9081)"
```

**Validation Criteria**:
- [ ] Configuration validates without errors
- [ ] Dry run completes successfully
- [ ] Port assignments are unique across profiles
- [ ] Data directories exist and are writable
- [ ] Log directories can be created

---

## Module 3: Basic Operations Exercises

### Exercise 3.1: Start and Monitor Distributed Node

**Objective**: Start a distributed node and monitor its health and status.

**Time Estimate**: 25 minutes

**Step-by-Step**:
```bash
# Step 1: Start node in background
./bin/ollama-distributed start --config ~/.ollama-distributed/profiles/development.yaml &
NODE_PID=$!

# Step 2: Wait for startup
sleep 10

# Step 3: Check health
curl http://localhost:8080/health

# Step 4: Run health monitoring
chmod +x training/code-examples/03-operations/health-monitoring-dashboard.sh
./training/code-examples/03-operations/health-monitoring-dashboard.sh check

# Step 5: Generate health report
./training/code-examples/03-operations/health-monitoring-dashboard.sh report --output health-report.txt

# Step 6: Clean shutdown
kill $NODE_PID
```

**Validation Criteria**:
- [ ] Node starts without errors
- [ ] Health endpoint returns "healthy" status
- [ ] All services are listening on expected ports
- [ ] Health monitoring dashboard shows green status
- [ ] Node shuts down gracefully

**Monitoring Checklist**:
- [ ] API server responding
- [ ] Web interface accessible
- [ ] P2P networking initialized
- [ ] Metrics collection active
- [ ] Log files being written

---

### Exercise 3.2: Multi-Node Cluster Setup

**Objective**: Set up a multi-node cluster and verify peer discovery.

**Time Estimate**: 30 minutes

**Instructions**:
Create a local multi-node cluster using different ports and verify connectivity.

**Step-by-Step**:
```bash
# Step 1: Create node2 configuration
cp ~/.ollama-distributed/profiles/development.yaml ~/.ollama-distributed/profiles/node2.yaml

# Step 2: Update node2 ports (edit the file)
sed -i 's/8080/8082/g' ~/.ollama-distributed/profiles/node2.yaml
sed -i 's/8081/8083/g' ~/.ollama-distributed/profiles/node2.yaml  
sed -i 's/4001/4002/g' ~/.ollama-distributed/profiles/node2.yaml

# Step 3: Start first node
./bin/ollama-distributed start --config ~/.ollama-distributed/profiles/development.yaml &
NODE1_PID=$!

# Step 4: Start second node
./bin/ollama-distributed start --config ~/.ollama-distributed/profiles/node2.yaml &
NODE2_PID=$!

# Step 5: Wait for startup
sleep 15

# Step 6: Check both nodes
curl http://localhost:8080/health
curl http://localhost:8082/health

# Step 7: Check cluster status
curl http://localhost:8080/api/distributed/status
curl http://localhost:8082/api/distributed/status

# Step 8: Clean shutdown
kill $NODE1_PID $NODE2_PID
```

**Validation Criteria**:
- [ ] Both nodes start successfully
- [ ] Nodes discover each other
- [ ] Cluster status shows multiple nodes
- [ ] P2P networking is functional
- [ ] No port conflicts occur

---

## Module 4: API Integration Exercises

### Exercise 4.1: Comprehensive API Client Testing

**Objective**: Build and test a comprehensive API client for all endpoints.

**Time Estimate**: 30 minutes

**Step-by-Step**:
```bash
# Step 1: Start the service
./bin/ollama-distributed start --config ~/.ollama-distributed/profiles/development.yaml &
SERVICE_PID=$!
sleep 10

# Step 2: Test API client
cd training/code-examples/04-api-integration
go run comprehensive-api-client.go health
go run comprehensive-api-client.go basic
go run comprehensive-api-client.go monitor

# Step 3: Test different commands
go run comprehensive-api-client.go generate  # May fail if no models
go run comprehensive-api-client.go stream    # May fail if no models

# Step 4: Build client as tool
go build -o api-client comprehensive-api-client.go
./api-client monitor

# Step 5: Cleanup
kill $SERVICE_PID
```

**API Testing Checklist**:
- [ ] Health endpoint responds correctly
- [ ] Cluster status endpoint returns data
- [ ] Node listing works
- [ ] Metrics endpoint provides system data
- [ ] Error handling works for invalid requests
- [ ] Client properly handles timeouts

**Advanced Challenges**:
1. Add authentication to API client
2. Implement retry logic with exponential backoff
3. Add streaming response handling
4. Create client SDK for different languages

---

### Exercise 4.2: Custom Integration Development

**Objective**: Develop custom integration tools using the API.

**Time Estimate**: 45 minutes

**Requirements**:
Build one or more of the following tools:
1. **Health Dashboard**: Real-time monitoring interface
2. **Load Tester**: Concurrent API request testing
3. **Configuration Manager**: API-based config management
4. **Alert System**: Monitoring with threshold alerts

**Template Structure**:
```go
package main

import (
    // Your imports here
)

type CustomTool struct {
    client *OllamaDistributedClient
    config ToolConfig
}

func (t *CustomTool) Initialize() error {
    // Initialize your tool
    return nil
}

func (t *CustomTool) Run() error {
    // Main tool logic
    return nil
}

func main() {
    // Tool implementation
}
```

**Validation Criteria**:
- [ ] Tool compiles and runs without errors
- [ ] Proper error handling implemented
- [ ] Tool provides useful functionality
- [ ] Code follows Go best practices
- [ ] Tool includes help/usage information

---

## Module 5: Validation and Testing Exercises

### Exercise 5.1: Complete Training Validation

**Objective**: Run the comprehensive training validation suite.

**Time Estimate**: 20 minutes

**Step-by-Step**:
```bash
# Step 1: Run validation suite
cd training/code-examples/05-validation-testing
go run training-validation-suite.go --output validation-results.json

# Step 2: Review results
cat validation-results.json | jq '.'

# Step 3: Generate detailed report
go run training-validation-suite.go --output detailed-report.json --config custom-validation.yaml

# Step 4: Address any failed tests
# (Follow remediation steps based on failures)
```

**Validation Categories**:
1. **Prerequisites**: System requirements
2. **Installation**: Binary and dependencies
3. **Configuration**: Profile validation
4. **Startup**: Service initialization  
5. **API**: Endpoint functionality
6. **Performance**: Response times and load
7. **Security**: Basic security checks

**Success Criteria**:
- [ ] All prerequisite tests pass
- [ ] Installation tests pass
- [ ] Configuration tests pass  
- [ ] At least 80% of API tests pass
- [ ] Performance within acceptable limits
- [ ] No critical security issues

---

### Exercise 5.2: Custom Test Development

**Objective**: Develop custom tests for specific scenarios.

**Time Estimate**: 35 minutes

**Requirements**:
Extend the validation suite with custom tests for:
1. **Model Management**: Test model loading/unloading
2. **Cluster Operations**: Test node joining/leaving
3. **Failover**: Test leader election and failover
4. **Load Balancing**: Test request distribution

**Template**:
```go
func (ts *TestSuite) validateCustomScenario() {
    category := "Custom"
    
    ts.runTest("Your Test Name", category, func() error {
        // Your test implementation
        return nil
    })
}
```

**Validation Criteria**:
- [ ] Custom tests integrate with existing suite
- [ ] Tests cover edge cases and error conditions
- [ ] Tests provide meaningful feedback
- [ ] Tests are repeatable and reliable

---

## Certification Exercises

### Final Certification Challenge

**Objective**: Complete end-to-end deployment and operation scenario.

**Time Estimate**: 60 minutes

**Scenario**: Deploy a complete Ollama Distributed system with:
1. Multi-node cluster (3 nodes minimum)
2. Load balancing configuration
3. Health monitoring
4. API integration tools
5. Automated testing

**Deliverables**:
1. **Configuration Files**: All node configurations
2. **Deployment Scripts**: Automated setup scripts
3. **Monitoring Tools**: Custom health monitoring
4. **API Client**: Working API client with all endpoints
5. **Test Suite**: Custom validation tests
6. **Documentation**: Complete setup and operation guide

**Assessment Criteria**:
- [ ] System deploys successfully
- [ ] All nodes communicate properly
- [ ] Health monitoring is functional
- [ ] API endpoints respond correctly
- [ ] Load balancing works as expected
- [ ] System handles node failures gracefully
- [ ] Documentation is complete and accurate

**Bonus Challenges**:
1. Implement automated model distribution
2. Add Prometheus/Grafana monitoring
3. Create Docker-based deployment
4. Implement rolling updates
5. Add comprehensive logging and alerting

---

## Exercise Solutions Repository

All exercise solutions are provided in the `training/solutions/` directory with:
- Complete working code
- Detailed explanations
- Alternative approaches
- Performance optimizations
- Security considerations

### Solution Structure
```
training/solutions/
├── module-1-installation/
│   ├── exercise-1.1-solution.sh
│   ├── exercise-1.2-solution.sh
│   └── README.md
├── module-2-configuration/
│   ├── exercise-2.1-solution.go
│   ├── exercise-2.2-solution.yaml
│   └── README.md
├── module-3-operations/
│   ├── exercise-3.1-solution.sh
│   ├── exercise-3.2-solution.sh
│   └── README.md
├── module-4-api-integration/
│   ├── exercise-4.1-solution.go
│   ├── exercise-4.2-solution.go
│   └── README.md
└── module-5-validation/
    ├── exercise-5.1-solution.go
    ├── exercise-5.2-solution.go
    └── README.md
```

### Using Solutions

Solutions are provided for:
- **Learning**: Understand different approaches
- **Reference**: Complete working examples  
- **Troubleshooting**: Fix common issues
- **Extension**: Advanced implementations

**Best Practices**:
1. Attempt exercises before viewing solutions
2. Compare your approach with provided solutions
3. Understand why solutions work
4. Experiment with modifications
5. Use solutions as starting points for extensions

---

## Assessment and Certification

### Exercise Completion Tracking

Use the provided tracking sheet to monitor progress:

```
Module 1 - Installation and Setup
[ ] Exercise 1.1: Complete Installation from Source
[ ] Exercise 1.2: Environment Validation

Module 2 - Configuration Management  
[ ] Exercise 2.1: Create Custom Configuration Profiles
[ ] Exercise 2.2: Configuration Validation and Testing

Module 3 - Basic Operations
[ ] Exercise 3.1: Start and Monitor Distributed Node
[ ] Exercise 3.2: Multi-Node Cluster Setup

Module 4 - API Integration
[ ] Exercise 4.1: Comprehensive API Client Testing
[ ] Exercise 4.2: Custom Integration Development

Module 5 - Validation and Testing
[ ] Exercise 5.1: Complete Training Validation
[ ] Exercise 5.2: Custom Test Development

Final Certification
[ ] Certification Challenge Completed
[ ] All Deliverables Submitted
```

### Certification Requirements

To receive Ollama Distributed Training Certification:
1. Complete all required exercises (80% minimum)
2. Pass final certification challenge
3. Submit all required deliverables
4. Demonstrate practical competency

**Certificate includes**:
- Completion verification
- Skill assessment results
- Practical competency demonstration
- Date of completion
- Valid for 2 years
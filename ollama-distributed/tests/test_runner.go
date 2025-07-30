package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
)

// TestRunner manages test execution across different test types
type TestRunner struct {
	config     *TestConfig
	testSuites []TestSuite
	results    []TestResult
}

// TestConfig holds configuration for test execution
type TestConfig struct {
	// Test execution settings
	Timeout         time.Duration `json:"timeout"`
	Parallel        bool          `json:"parallel"`
	Verbose         bool          `json:"verbose"`
	FailFast        bool          `json:"fail_fast"`
	
	// Test selection
	TestTypes       []string      `json:"test_types"`
	TestPackages    []string      `json:"test_packages"`
	TestPattern     string        `json:"test_pattern"`
	
	// Environment settings
	NodeCount       int           `json:"node_count"`
	UseDocker       bool          `json:"use_docker"`
	CleanupOnExit   bool          `json:"cleanup_on_exit"`
	
	// Resource limits
	CPULimit        string        `json:"cpu_limit"`
	MemoryLimit     string        `json:"memory_limit"`
	DiskLimit       string        `json:"disk_limit"`
	
	// CI/CD settings
	CIMode          bool          `json:"ci_mode"`
	JUnitOutput     string        `json:"junit_output"`
	CoverageOutput  string        `json:"coverage_output"`
	ArtifactsDir    string        `json:"artifacts_dir"`
}

// TestSuite represents a test suite
type TestSuite struct {
	Name        string
	Path        string
	Type        TestType
	Description string
	Tags        []string
	Timeout     time.Duration
	Requires    []string
}

// TestType represents different types of tests
type TestType string

const (
	UnitTest        TestType = "unit"
	IntegrationTest TestType = "integration"
	E2ETest         TestType = "e2e"
	PerformanceTest TestType = "performance"
	ChaosTest       TestType = "chaos"
)

// TestResult represents test execution results
type TestResult struct {
	Suite     string        `json:"suite"`
	Package   string        `json:"package"`
	Status    TestStatus    `json:"status"`
	Duration  time.Duration `json:"duration"`
	Output    string        `json:"output"`
	Error     string        `json:"error,omitempty"`
	Coverage  float64       `json:"coverage,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}

// TestStatus represents test execution status
type TestStatus string

const (
	StatusPass    TestStatus = "pass"
	StatusFail    TestStatus = "fail"
	StatusSkip    TestStatus = "skip"
	StatusTimeout TestStatus = "timeout"
)

// NewTestRunner creates a new test runner
func NewTestRunner(config *TestConfig) *TestRunner {
	if config == nil {
		config = DefaultTestConfig()
	}

	return &TestRunner{
		config:     config,
		testSuites: make([]TestSuite, 0),
		results:    make([]TestResult, 0),
	}
}

// DefaultTestConfig returns default test configuration
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		Timeout:       30 * time.Minute,
		Parallel:      true,
		Verbose:       false,
		FailFast:      false,
		TestTypes:     []string{"unit", "integration"},
		NodeCount:     3,
		UseDocker:     false,
		CleanupOnExit: true,
		CPULimit:      "2",
		MemoryLimit:   "4Gi",
		DiskLimit:     "10Gi",
		CIMode:        false,
		ArtifactsDir:  "./test-artifacts",
	}
}

// RegisterTestSuite registers a test suite
func (tr *TestRunner) RegisterTestSuite(suite TestSuite) {
	tr.testSuites = append(tr.testSuites, suite)
}

// RegisterDefaultSuites registers default test suites
func (tr *TestRunner) RegisterDefaultSuites() {
	suites := []TestSuite{
		{
			Name:        "Unit Tests",
			Path:        "./tests/unit",
			Type:        UnitTest,
			Description: "Fast unit tests for individual components",
			Tags:        []string{"unit", "fast"},
			Timeout:     5 * time.Minute,
			Requires:    []string{},
		},
		{
			Name:        "Integration Tests",
			Path:        "./tests/integration",
			Type:        IntegrationTest,
			Description: "Integration tests for component interactions",
			Tags:        []string{"integration", "slow"},
			Timeout:     15 * time.Minute,
			Requires:    []string{"cluster"},
		},
		{
			Name:        "E2E Tests",
			Path:        "./tests/e2e",
			Type:        E2ETest,
			Description: "End-to-end workflow tests",
			Tags:        []string{"e2e", "slow"},
			Timeout:     30 * time.Minute,
			Requires:    []string{"cluster", "external"},
		},
		{
			Name:        "Performance Tests",
			Path:        "./tests/performance",
			Type:        PerformanceTest,
			Description: "Performance benchmarks and load tests",
			Tags:        []string{"performance", "benchmark"},
			Timeout:     60 * time.Minute,
			Requires:    []string{"cluster", "resources"},
		},
		{
			Name:        "Chaos Tests",
			Path:        "./tests/chaos",
			Type:        ChaosTest,
			Description: "Chaos engineering and fault injection tests",
			Tags:        []string{"chaos", "fault", "slow"},
			Timeout:     45 * time.Minute,
			Requires:    []string{"cluster", "privileges"},
		},
	}

	for _, suite := range suites {
		tr.RegisterTestSuite(suite)
	}
}

// RunAllTests runs all registered test suites
func (tr *TestRunner) RunAllTests(ctx context.Context) error {
	fmt.Printf("🚀 Starting test execution with %d test suites\n", len(tr.testSuites))
	
	// Setup test environment
	if err := tr.setupTestEnvironment(); err != nil {
		return fmt.Errorf("failed to setup test environment: %w", err)
	}

	// Create artifacts directory
	if err := os.MkdirAll(tr.config.ArtifactsDir, 0755); err != nil {
		return fmt.Errorf("failed to create artifacts directory: %w", err)
	}

	// Run test suites
	for _, suite := range tr.testSuites {
		if !tr.shouldRunSuite(suite) {
			fmt.Printf("⏭️  Skipping test suite: %s\n", suite.Name)
			continue
		}

		fmt.Printf("🧪 Running test suite: %s\n", suite.Name)
		
		result := tr.runTestSuite(ctx, suite)
		tr.results = append(tr.results, result)
		
		if result.Status == StatusFail && tr.config.FailFast {
			fmt.Printf("❌ Test suite failed and fail-fast is enabled. Stopping.\n")
			break
		}
	}

	// Generate reports
	if err := tr.generateReports(); err != nil {
		fmt.Printf("⚠️  Failed to generate reports: %v\n", err)
	}

	// Cleanup
	if tr.config.CleanupOnExit {
		if err := tr.cleanup(); err != nil {
			fmt.Printf("⚠️  Failed to cleanup: %v\n", err)
		}
	}

	// Check overall results
	return tr.checkResults()
}

// runTestSuite runs a single test suite
func (tr *TestRunner) runTestSuite(ctx context.Context, suite TestSuite) TestResult {
	startTime := time.Now()
	
	result := TestResult{
		Suite:     suite.Name,
		Package:   suite.Path,
		Status:    StatusPass,
		Timestamp: startTime,
	}

	// Create suite-specific context with timeout
	suiteCtx, cancel := context.WithTimeout(ctx, suite.Timeout)
	defer cancel()

	// Check requirements
	if err := tr.checkRequirements(suite); err != nil {
		result.Status = StatusSkip
		result.Error = fmt.Sprintf("Requirements not met: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}

	// Build test command
	cmd := tr.buildTestCommand(suite)
	cmd.Dir = suite.Path

	// Run the test
	output, err := tr.runCommand(suiteCtx, cmd)
	result.Output = output
	result.Duration = time.Since(startTime)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result.Status = StatusTimeout
		} else {
			result.Status = StatusFail
		}
		result.Error = err.Error()
	}

	// Parse coverage if available
	if tr.config.CoverageOutput != "" {
		result.Coverage = tr.parseCoverage(suite)
	}

	// Log result
	tr.logResult(result)

	return result
}

// buildTestCommand builds the test command for a suite
func (tr *TestRunner) buildTestCommand(suite TestSuite) *exec.Cmd {
	args := []string{"test"}
	
	if tr.config.Verbose {
		args = append(args, "-v")
	}
	
	if tr.config.Parallel {
		args = append(args, "-parallel", fmt.Sprintf("%d", runtime.NumCPU()))
	}

	if tr.config.TestPattern != "" {
		args = append(args, "-run", tr.config.TestPattern)
	}

	if tr.config.CoverageOutput != "" {
		coverageFile := filepath.Join(tr.config.ArtifactsDir, fmt.Sprintf("%s-coverage.out", suite.Name))
		args = append(args, "-coverprofile", coverageFile)
	}

	if tr.config.Timeout > 0 {
		args = append(args, "-timeout", tr.config.Timeout.String())
	}

	// Add test type specific flags
	switch suite.Type {
	case PerformanceTest:
		args = append(args, "-bench", ".", "-benchmem")
	case ChaosTest:
		args = append(args, "-tags", "chaos")
	}

	// Add package path
	args = append(args, "./...")

	return exec.Command("go", args...)
}

// runCommand runs a command with context
func (tr *TestRunner) runCommand(ctx context.Context, cmd *exec.Cmd) (string, error) {
	// Set environment variables
	cmd.Env = append(os.Environ(), tr.buildEnvironment()...)

	// Capture output
	output, err := cmd.CombinedOutput()
	
	if ctx.Err() == context.DeadlineExceeded {
		return string(output), context.DeadlineExceeded
	}

	return string(output), err
}

// buildEnvironment builds environment variables for test execution
func (tr *TestRunner) buildEnvironment() []string {
	env := []string{
		fmt.Sprintf("OLLAMA_TEST_NODE_COUNT=%d", tr.config.NodeCount),
		fmt.Sprintf("OLLAMA_TEST_TIMEOUT=%s", tr.config.Timeout.String()),
		fmt.Sprintf("OLLAMA_TEST_ARTIFACTS_DIR=%s", tr.config.ArtifactsDir),
	}

	if tr.config.CIMode {
		env = append(env, "OLLAMA_TEST_CI=true")
	}

	if tr.config.UseDocker {
		env = append(env, "OLLAMA_TEST_USE_DOCKER=true")
	}

	return env
}

// shouldRunSuite checks if a test suite should be run
func (tr *TestRunner) shouldRunSuite(suite TestSuite) bool {
	// Check if test type is enabled
	if len(tr.config.TestTypes) > 0 {
		found := false
		for _, testType := range tr.config.TestTypes {
			if strings.EqualFold(testType, string(suite.Type)) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check if package is enabled
	if len(tr.config.TestPackages) > 0 {
		found := false
		for _, pkg := range tr.config.TestPackages {
			if strings.Contains(suite.Path, pkg) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// checkRequirements checks if suite requirements are met
func (tr *TestRunner) checkRequirements(suite TestSuite) error {
	for _, req := range suite.Requires {
		switch req {
		case "cluster":
			if tr.config.NodeCount < 1 {
				return fmt.Errorf("cluster tests require at least 1 node")
			}
		case "external":
			// Check external dependencies
			if tr.config.CIMode {
				return fmt.Errorf("external tests not supported in CI mode")
			}
		case "resources":
			// Check resource availability
			if !tr.hasResources() {
				return fmt.Errorf("insufficient resources for performance tests")
			}
		case "privileges":
			// Check if running with required privileges
			if !tr.hasPrivileges() {
				return fmt.Errorf("chaos tests require elevated privileges")
			}
		}
	}
	return nil
}

// hasResources checks if system has sufficient resources
func (tr *TestRunner) hasResources() bool {
	// Simple check - can be enhanced with actual resource monitoring
	return runtime.NumCPU() >= 2 && runtime.GOMAXPROCS(0) >= 2
}

// hasPrivileges checks if running with required privileges
func (tr *TestRunner) hasPrivileges() bool {
	// Simple check - can be enhanced with actual privilege checking
	return os.Getuid() == 0 || tr.config.UseDocker
}

// setupTestEnvironment sets up the test environment
func (tr *TestRunner) setupTestEnvironment() error {
	fmt.Println("🔧 Setting up test environment...")

	// Create test data directories
	testDirs := []string{
		filepath.Join(tr.config.ArtifactsDir, "logs"),
		filepath.Join(tr.config.ArtifactsDir, "coverage"),
		filepath.Join(tr.config.ArtifactsDir, "reports"),
	}

	for _, dir := range testDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Setup Docker environment if needed
	if tr.config.UseDocker {
		if err := tr.setupDockerEnvironment(); err != nil {
			return fmt.Errorf("failed to setup Docker environment: %w", err)
		}
	}

	return nil
}

// setupDockerEnvironment sets up Docker environment for testing
func (tr *TestRunner) setupDockerEnvironment() error {
	fmt.Println("🐳 Setting up Docker environment...")

	// Check if Docker is available
	cmd := exec.Command("docker", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker not available: %w", err)
	}

	// Pull required images
	images := []string{
		"golang:1.21-alpine",
		"redis:7-alpine",
		"postgres:15-alpine",
	}

	for _, image := range images {
		fmt.Printf("📥 Pulling Docker image: %s\n", image)
		cmd := exec.Command("docker", "pull", image)
		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️  Failed to pull image %s: %v\n", image, err)
		}
	}

	return nil
}

// parseCoverage parses coverage information
func (tr *TestRunner) parseCoverage(suite TestSuite) float64 {
	coverageFile := filepath.Join(tr.config.ArtifactsDir, fmt.Sprintf("%s-coverage.out", suite.Name))
	
	if _, err := os.Stat(coverageFile); os.IsNotExist(err) {
		return 0.0
	}

	// Simple coverage parsing - can be enhanced
	cmd := exec.Command("go", "tool", "cover", "-func", coverageFile)
	output, err := cmd.Output()
	if err != nil {
		return 0.0
	}

	// Parse coverage percentage from output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "total:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				coverageStr := strings.TrimSuffix(parts[2], "%")
				if coverage, err := fmt.Sscanf(coverageStr, "%f", &coverage); err == nil {
					return coverage
				}
			}
		}
	}

	return 0.0
}

// logResult logs test result
func (tr *TestRunner) logResult(result TestResult) {
	statusIcon := "✅"
	switch result.Status {
	case StatusFail:
		statusIcon = "❌"
	case StatusSkip:
		statusIcon = "⏭️"
	case StatusTimeout:
		statusIcon = "⏱️"
	}

	fmt.Printf("%s %s (%s)\n", statusIcon, result.Suite, result.Duration.Round(time.Millisecond))
	
	if result.Error != "" {
		fmt.Printf("   Error: %s\n", result.Error)
	}
	
	if result.Coverage > 0 {
		fmt.Printf("   Coverage: %.1f%%\n", result.Coverage)
	}
}

// generateReports generates test reports
func (tr *TestRunner) generateReports() error {
	fmt.Println("📊 Generating test reports...")

	// Generate summary report
	if err := tr.generateSummaryReport(); err != nil {
		return fmt.Errorf("failed to generate summary report: %w", err)
	}

	// Generate JUnit XML report if requested
	if tr.config.JUnitOutput != "" {
		if err := tr.generateJUnitReport(); err != nil {
			return fmt.Errorf("failed to generate JUnit report: %w", err)
		}
	}

	return nil
}

// generateSummaryReport generates a summary report
func (tr *TestRunner) generateSummaryReport() error {
	reportFile := filepath.Join(tr.config.ArtifactsDir, "test-summary.txt")
	
	file, err := os.Create(reportFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write summary
	fmt.Fprintf(file, "Test Execution Summary\n")
	fmt.Fprintf(file, "=====================\n\n")
	
	totalTests := len(tr.results)
	passedTests := 0
	failedTests := 0
	skippedTests := 0
	timeoutTests := 0
	
	for _, result := range tr.results {
		switch result.Status {
		case StatusPass:
			passedTests++
		case StatusFail:
			failedTests++
		case StatusSkip:
			skippedTests++
		case StatusTimeout:
			timeoutTests++
		}
	}

	fmt.Fprintf(file, "Total Tests: %d\n", totalTests)
	fmt.Fprintf(file, "Passed: %d\n", passedTests)
	fmt.Fprintf(file, "Failed: %d\n", failedTests)
	fmt.Fprintf(file, "Skipped: %d\n", skippedTests)
	fmt.Fprintf(file, "Timeout: %d\n", timeoutTests)
	fmt.Fprintf(file, "\nSuccess Rate: %.1f%%\n", float64(passedTests)/float64(totalTests)*100)
	
	fmt.Fprintf(file, "\nDetailed Results:\n")
	fmt.Fprintf(file, "================\n\n")
	
	for _, result := range tr.results {
		fmt.Fprintf(file, "Suite: %s\n", result.Suite)
		fmt.Fprintf(file, "Status: %s\n", result.Status)
		fmt.Fprintf(file, "Duration: %s\n", result.Duration)
		if result.Coverage > 0 {
			fmt.Fprintf(file, "Coverage: %.1f%%\n", result.Coverage)
		}
		if result.Error != "" {
			fmt.Fprintf(file, "Error: %s\n", result.Error)
		}
		fmt.Fprintf(file, "\n")
	}

	fmt.Printf("📄 Summary report written to: %s\n", reportFile)
	return nil
}

// generateJUnitReport generates a JUnit XML report
func (tr *TestRunner) generateJUnitReport() error {
	// JUnit XML generation implementation
	// This would generate XML in JUnit format for CI/CD systems
	fmt.Printf("📋 JUnit report would be generated at: %s\n", tr.config.JUnitOutput)
	return nil
}

// checkResults checks overall test results
func (tr *TestRunner) checkResults() error {
	failedCount := 0
	for _, result := range tr.results {
		if result.Status == StatusFail {
			failedCount++
		}
	}

	if failedCount > 0 {
		return fmt.Errorf("%d test suite(s) failed", failedCount)
	}

	fmt.Println("🎉 All tests passed!")
	return nil
}

// cleanup cleans up test resources
func (tr *TestRunner) cleanup() error {
	fmt.Println("🧹 Cleaning up test environment...")

	// Clean up Docker containers if used
	if tr.config.UseDocker {
		cmd := exec.Command("docker", "system", "prune", "-f")
		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️  Failed to cleanup Docker resources: %v\n", err)
		}
	}

	// Clean up temporary files
	tempDirs := []string{
		"/tmp/ollama-test-*",
		"/tmp/test-cluster-*",
	}

	for _, pattern := range tempDirs {
		cmd := exec.Command("sh", "-c", fmt.Sprintf("rm -rf %s", pattern))
		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️  Failed to cleanup temp files %s: %v\n", pattern, err)
		}
	}

	return nil
}

// Main function to run tests
func main() {
	// Parse command line arguments
	testConfig := parseArgs()
	
	// Create test runner
	runner := NewTestRunner(testConfig)
	
	// Register test suites
	runner.RegisterDefaultSuites()
	
	// Run tests
	ctx := context.Background()
	if err := runner.RunAllTests(ctx); err != nil {
		fmt.Printf("❌ Test execution failed: %v\n", err)
		os.Exit(1)
	}
}

// parseArgs parses command line arguments
func parseArgs() *TestConfig {
	config := DefaultTestConfig()
	
	// Simple argument parsing - can be enhanced with proper CLI library
	args := os.Args[1:]
	
	for i, arg := range args {
		switch arg {
		case "-v", "--verbose":
			config.Verbose = true
		case "-p", "--parallel":
			config.Parallel = true
		case "--ci":
			config.CIMode = true
		case "--docker":
			config.UseDocker = true
		case "--fail-fast":
			config.FailFast = true
		case "--timeout":
			if i+1 < len(args) {
				if timeout, err := time.ParseDuration(args[i+1]); err == nil {
					config.Timeout = timeout
				}
			}
		case "--types":
			if i+1 < len(args) {
				config.TestTypes = strings.Split(args[i+1], ",")
			}
		case "--nodes":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &config.NodeCount)
			}
		}
	}
	
	return config
}
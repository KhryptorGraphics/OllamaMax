package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// TestRunner configuration
type TestRunnerConfig struct {
	Verbose          bool
	Coverage         bool
	Benchmark        bool
	Integration      bool
	Performance      bool
	Validation       bool
	Parallel         bool
	Race             bool
	Timeout          time.Duration
	OutputDir        string
	ReportFormat     string
	MaxWorkers       int
}

// SwarmTestRunner manages execution of all swarm tests
type SwarmTestRunner struct {
	config     *TestRunnerConfig
	workingDir string
	testSuites []TestSuite
}

// TestSuite represents a group of related tests
type TestSuite struct {
	Name        string
	Path        string
	Description string
	Tags        []string
	Priority    int
	Timeout     time.Duration
}

// NewSwarmTestRunner creates a new test runner
func NewSwarmTestRunner(config *TestRunnerConfig) *SwarmTestRunner {
	workingDir, _ := os.Getwd()
	
	runner := &SwarmTestRunner{
		config:     config,
		workingDir: workingDir,
	}
	
	runner.initializeTestSuites()
	return runner
}

// initializeTestSuites sets up all test suites
func (str *SwarmTestRunner) initializeTestSuites() {
	str.testSuites = []TestSuite{
		{
			Name:        "swarm_operations",
			Path:        "./swarm_operations_test.go",
			Description: "Core swarm operations and coordination tests",
			Tags:        []string{"unit", "swarm", "coordination"},
			Priority:    1,
			Timeout:     5 * time.Minute,
		},
		{
			Name:        "validation_routines",
			Path:        "./validation_routines.go",
			Description: "Validation framework and health checks",
			Tags:        []string{"unit", "validation", "health"},
			Priority:    2,
			Timeout:     3 * time.Minute,
		},
		{
			Name:        "file_operations_coordination",
			Path:        "./file_operations_coordination_test.go",
			Description: "File operations and coordination tests",
			Tags:        []string{"unit", "file_ops", "coordination"},
			Priority:    2,
			Timeout:     4 * time.Minute,
		},
		{
			Name:        "performance_measurement",
			Path:        "./performance_measurement_tools.go",
			Description: "Performance measurement and monitoring tools",
			Tags:        []string{"unit", "performance", "monitoring"},
			Priority:    3,
			Timeout:     3 * time.Minute,
		},
		{
			Name:        "integration",
			Path:        "./integration_test.go",
			Description: "Comprehensive integration tests",
			Tags:        []string{"integration", "comprehensive"},
			Priority:    4,
			Timeout:     10 * time.Minute,
		},
	}
}

// RunAllTests executes all configured test suites
func (str *SwarmTestRunner) RunAllTests(ctx context.Context) error {
	fmt.Println("🧪 Starting Swarm Test Suite")
	fmt.Println("============================")
	fmt.Printf("Configuration:\n")
	fmt.Printf("  Verbose: %t\n", str.config.Verbose)
	fmt.Printf("  Coverage: %t\n", str.config.Coverage)
	fmt.Printf("  Parallel: %t\n", str.config.Parallel)
	fmt.Printf("  Race Detection: %t\n", str.config.Race)
	fmt.Printf("  Timeout: %v\n", str.config.Timeout)
	fmt.Printf("  Output Directory: %s\n", str.config.OutputDir)
	fmt.Printf("  Max Workers: %d\n", str.config.MaxWorkers)
	fmt.Println()

	// Create output directory
	if err := os.MkdirAll(str.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Filter test suites based on configuration
	suitesToRun := str.filterTestSuites()
	
	if len(suitesToRun) == 0 {
		return fmt.Errorf("no test suites selected for execution")
	}

	fmt.Printf("📋 Running %d test suites:\n", len(suitesToRun))
	for _, suite := range suitesToRun {
		fmt.Printf("  - %s: %s\n", suite.Name, suite.Description)
	}
	fmt.Println()

	startTime := time.Now()
	var totalResults []TestResult
	
	for _, suite := range suitesToRun {
		fmt.Printf("🚀 Executing test suite: %s\n", suite.Name)
		
		result, err := str.runTestSuite(ctx, suite)
		if err != nil {
			fmt.Printf("❌ Test suite %s failed: %v\n", suite.Name, err)
			result.Success = false
			result.Error = err.Error()
		} else if result.Success {
			fmt.Printf("✅ Test suite %s completed successfully\n", suite.Name)
		} else {
			fmt.Printf("⚠️  Test suite %s completed with issues\n", suite.Name)
		}
		
		totalResults = append(totalResults, result)
		fmt.Println()
	}

	totalDuration := time.Since(startTime)
	
	// Generate comprehensive report
	str.generateReport(totalResults, totalDuration)
	
	// Check overall success
	return str.checkOverallResults(totalResults)
}

// filterTestSuites filters test suites based on configuration
func (str *SwarmTestRunner) filterTestSuites() []TestSuite {
	var filtered []TestSuite
	
	for _, suite := range str.testSuites {
		include := true
		
		// Filter by type
		if str.config.Integration && !containsTag(suite.Tags, "integration") {
			include = false
		}
		if str.config.Performance && !containsTag(suite.Tags, "performance") {
			include = false
		}
		if str.config.Validation && !containsTag(suite.Tags, "validation") {
			include = false
		}
		
		if include {
			filtered = append(filtered, suite)
		}
	}
	
	return filtered
}

// runTestSuite executes a single test suite
func (str *SwarmTestRunner) runTestSuite(ctx context.Context, suite TestSuite) (TestResult, error) {
	result := TestResult{
		Suite:     suite.Name,
		StartTime: time.Now(),
	}
	
	// Build test command
	cmd := str.buildTestCommand(suite)
	
	// Set up output capture
	outputFile := filepath.Join(str.config.OutputDir, fmt.Sprintf("%s.log", suite.Name))
	output, err := os.Create(outputFile)
	if err != nil {
		return result, fmt.Errorf("failed to create output file: %w", err)
	}
	defer output.Close()
	
	if str.config.Verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = output
		cmd.Stderr = output
	}
	
	// Execute with timeout
	suiteCtx, cancel := context.WithTimeout(ctx, suite.Timeout)
	defer cancel()
	
	cmd = exec.CommandContext(suiteCtx, cmd.Args[0], cmd.Args[1:]...)
	
	err = cmd.Run()
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Success = err == nil
	
	if err != nil {
		result.Error = err.Error()
	}
	
	// Extract coverage if enabled
	if str.config.Coverage {
		result.Coverage = str.extractCoverage(suite.Name)
	}
	
	return result, nil
}

// buildTestCommand builds the go test command for a suite
func (str *SwarmTestRunner) buildTestCommand(suite TestSuite) *exec.Cmd {
	args := []string{"test"}
	
	if str.config.Verbose {
		args = append(args, "-v")
	}
	
	if str.config.Race {
		args = append(args, "-race")
	}
	
	if str.config.Coverage {
		coverageFile := filepath.Join(str.config.OutputDir, fmt.Sprintf("%s_coverage.out", suite.Name))
		args = append(args, "-coverprofile", coverageFile, "-covermode=atomic")
	}
	
	if str.config.Benchmark {
		args = append(args, "-bench=.")
	}
	
	args = append(args, "-timeout", suite.Timeout.String())
	
	if len(suite.Tags) > 0 {
		args = append(args, "-tags", strings.Join(suite.Tags, ","))
	}
	
	args = append(args, suite.Path)
	
	return exec.Command("go", args...)
}

// extractCoverage extracts coverage percentage from coverage file
func (str *SwarmTestRunner) extractCoverage(suiteName string) float64 {
	coverageFile := filepath.Join(str.config.OutputDir, fmt.Sprintf("%s_coverage.out", suiteName))
	
	cmd := exec.Command("go", "tool", "cover", "-func", coverageFile)
	output, err := cmd.Output()
	if err != nil {
		return 0.0
	}
	
	// Parse coverage output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "total:") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				coverageStr := strings.TrimSuffix(fields[2], "%")
				var coverage float64
				fmt.Sscanf(coverageStr, "%f", &coverage)
				return coverage
			}
		}
	}
	
	return 0.0
}

// generateReport generates a comprehensive test report
func (str *SwarmTestRunner) generateReport(results []TestResult, totalDuration time.Duration) {
	fmt.Println("📊 SWARM TEST REPORT")
	fmt.Println("====================")
	
	successCount := 0
	failureCount := 0
	totalCoverage := 0.0
	coverageCount := 0
	
	fmt.Printf("%-25s %-10s %-12s %-10s %s\n", "TEST SUITE", "STATUS", "DURATION", "COVERAGE", "ERROR")
	fmt.Println(strings.Repeat("-", 80))
	
	for _, result := range results {
		status := "✅ PASS"
		errorMsg := ""
		
		if !result.Success {
			status = "❌ FAIL"
			failureCount++
			errorMsg = result.Error
			if len(errorMsg) > 30 {
				errorMsg = errorMsg[:27] + "..."
			}
		} else {
			successCount++
		}
		
		coverage := "N/A"
		if result.Coverage > 0 {
			coverage = fmt.Sprintf("%.1f%%", result.Coverage)
			totalCoverage += result.Coverage
			coverageCount++
		}
		
		fmt.Printf("%-25s %-10s %-12s %-10s %s\n",
			result.Suite,
			status,
			result.Duration.Round(time.Second),
			coverage,
			errorMsg)
	}
	
	fmt.Println(strings.Repeat("-", 80))
	
	// Summary
	total := len(results)
	successRate := float64(successCount) / float64(total) * 100
	
	fmt.Printf("📈 SUMMARY:\n")
	fmt.Printf("   Total Test Suites: %d\n", total)
	fmt.Printf("   Successful: %d (%.1f%%)\n", successCount, successRate)
	fmt.Printf("   Failed: %d (%.1f%%)\n", failureCount, 100-successRate)
	fmt.Printf("   Total Duration: %v\n", totalDuration.Round(time.Second))
	
	if coverageCount > 0 {
		avgCoverage := totalCoverage / float64(coverageCount)
		fmt.Printf("   Average Coverage: %.1f%%\n", avgCoverage)
	}
	
	fmt.Printf("\n📁 ARTIFACTS:\n")
	fmt.Printf("   Test Logs: %s/*.log\n", str.config.OutputDir)
	if str.config.Coverage {
		fmt.Printf("   Coverage Reports: %s/*_coverage.out\n", str.config.OutputDir)
	}
	
	// Generate detailed report file
	str.writeDetailedReport(results, totalDuration)
}

// writeDetailedReport writes a detailed report to file
func (str *SwarmTestRunner) writeDetailedReport(results []TestResult, totalDuration time.Duration) {
	reportFile := filepath.Join(str.config.OutputDir, "swarm_test_report.md")
	
	file, err := os.Create(reportFile)
	if err != nil {
		log.Printf("Failed to create detailed report: %v", err)
		return
	}
	defer file.Close()
	
	fmt.Fprintf(file, "# Swarm Test Suite Report\n\n")
	fmt.Fprintf(file, "Generated: %s\n\n", time.Now().Format(time.RFC3339))
	
	fmt.Fprintf(file, "## Configuration\n\n")
	fmt.Fprintf(file, "- Verbose: %t\n", str.config.Verbose)
	fmt.Fprintf(file, "- Coverage: %t\n", str.config.Coverage)
	fmt.Fprintf(file, "- Parallel: %t\n", str.config.Parallel)
	fmt.Fprintf(file, "- Race Detection: %t\n", str.config.Race)
	fmt.Fprintf(file, "- Total Duration: %v\n", totalDuration)
	fmt.Fprintf(file, "\n")
	
	fmt.Fprintf(file, "## Test Results\n\n")
	fmt.Fprintf(file, "| Test Suite | Status | Duration | Coverage | Error |\n")
	fmt.Fprintf(file, "|------------|--------|----------|----------|-------|\n")
	
	for _, result := range results {
		status := "✅ PASS"
		if !result.Success {
			status = "❌ FAIL"
		}
		
		coverage := "N/A"
		if result.Coverage > 0 {
			coverage = fmt.Sprintf("%.1f%%", result.Coverage)
		}
		
		errorMsg := result.Error
		if len(errorMsg) > 50 {
			errorMsg = errorMsg[:47] + "..."
		}
		
		fmt.Fprintf(file, "| %s | %s | %v | %s | %s |\n",
			result.Suite, status, result.Duration.Round(time.Second), coverage, errorMsg)
	}
	
	fmt.Printf("📄 Detailed report written to: %s\n", reportFile)
}

// checkOverallResults checks if all tests passed
func (str *SwarmTestRunner) checkOverallResults(results []TestResult) error {
	for _, result := range results {
		if !result.Success {
			return fmt.Errorf("test suite %s failed", result.Suite)
		}
	}
	return nil
}

// TestResult represents the result of a test suite execution
type TestResult struct {
	Suite     string
	Success   bool
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Coverage  float64
	Error     string
}

// Utility functions

func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

// Main function
func main() {
	// Parse command line flags
	config := &TestRunnerConfig{}
	
	flag.BoolVar(&config.Verbose, "v", false, "Verbose output")
	flag.BoolVar(&config.Coverage, "coverage", true, "Enable coverage reporting")
	flag.BoolVar(&config.Benchmark, "bench", false, "Run benchmarks")
	flag.BoolVar(&config.Integration, "integration", false, "Run only integration tests")
	flag.BoolVar(&config.Performance, "performance", false, "Run only performance tests")
	flag.BoolVar(&config.Validation, "validation", false, "Run only validation tests")
	flag.BoolVar(&config.Parallel, "parallel", true, "Enable parallel test execution")
	flag.BoolVar(&config.Race, "race", true, "Enable race detection")
	flag.DurationVar(&config.Timeout, "timeout", 30*time.Minute, "Global timeout for all tests")
	flag.StringVar(&config.OutputDir, "output", "./swarm-test-output", "Output directory for test artifacts")
	flag.StringVar(&config.ReportFormat, "format", "text", "Report format (text, json, xml)")
	flag.IntVar(&config.MaxWorkers, "workers", runtime.NumCPU(), "Maximum number of parallel workers")
	
	flag.Parse()
	
	// Create test runner
	runner := NewSwarmTestRunner(config)
	
	// Run tests
	ctx := context.Background()
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}
	
	if err := runner.RunAllTests(ctx); err != nil {
		fmt.Printf("❌ Test execution failed: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("✅ All swarm tests completed successfully!")
}
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// TestSuite represents a test suite configuration
type TestSuite struct {
	Name        string
	Path        string
	Timeout     time.Duration
	Parallel    bool
	Tags        []string
	Coverage    bool
	Race        bool
	Verbose     bool
	Environment map[string]string
}

// TestResult represents the result of a test execution
type TestResult struct {
	Suite     string
	Success   bool
	Duration  time.Duration
	Output    string
	Error     error
	Coverage  float64
}

// TestRunner manages comprehensive test execution
type TestRunner struct {
	config      *TestConfig
	results     []*TestResult
	resultsMu   sync.Mutex
	coverageDir string
	artifactsDir string
}

// TestConfig holds configuration for test execution
type TestConfig struct {
	Verbose        bool
	Parallel       bool
	Race           bool
	Coverage       bool
	FailFast       bool
	MaxWorkers     int
	Timeout        time.Duration
	SuiteFilter    []string
	TagFilter      []string
	OutputFormat   string
	ArtifactsDir   string
	NodeCount      int
	CI             bool
}

// NewTestRunner creates a new test runner
func NewTestRunner(config *TestConfig) *TestRunner {
	if config.ArtifactsDir == "" {
		config.ArtifactsDir = "./test-artifacts"
	}

	runner := &TestRunner{
		config:       config,
		results:      make([]*TestResult, 0),
		artifactsDir: config.ArtifactsDir,
		coverageDir:  filepath.Join(config.ArtifactsDir, "coverage"),
	}

	// Create artifacts directories
	os.MkdirAll(runner.artifactsDir, 0755)
	os.MkdirAll(runner.coverageDir, 0755)
	os.MkdirAll(filepath.Join(runner.artifactsDir, "logs"), 0755)

	return runner
}

// GetTestSuites returns all available test suites
func (tr *TestRunner) GetTestSuites() []*TestSuite {
	suites := []*TestSuite{
		{
			Name:     "unit",
			Path:     "./tests/unit/...",
			Timeout:  5 * time.Minute,
			Parallel: true,
			Tags:     []string{},
			Coverage: true,
			Race:     true,
			Verbose:  tr.config.Verbose,
			Environment: map[string]string{
				"GOMAXPROCS": strconv.Itoa(runtime.NumCPU()),
			},
		},
		{
			Name:     "integration",
			Path:     "./tests/integration/...",
			Timeout:  15 * time.Minute,
			Parallel: true,
			Tags:     []string{"integration"},
			Coverage: true,
			Race:     false, // Race detection can be flaky in integration tests
			Verbose:  tr.config.Verbose,
			Environment: map[string]string{
				"OLLAMA_TEST_NODE_COUNT":      strconv.Itoa(tr.config.NodeCount),
				"OLLAMA_TEST_ARTIFACTS_DIR":   tr.artifactsDir,
				"OLLAMA_TEST_CI":             strconv.FormatBool(tr.config.CI),
				"OLLAMA_TEST_USE_DOCKER":     "false",
				"GOMAXPROCS":                 strconv.Itoa(runtime.NumCPU()),
			},
		},
		{
			Name:     "e2e",
			Path:     "./tests/e2e/...",
			Timeout:  45 * time.Minute,
			Parallel: false, // E2E tests often can't run in parallel
			Tags:     []string{"e2e"},
			Coverage: false,
			Race:     false,
			Verbose:  tr.config.Verbose,
			Environment: map[string]string{
				"OLLAMA_TEST_NODE_COUNT":      strconv.Itoa(tr.config.NodeCount),
				"OLLAMA_TEST_ARTIFACTS_DIR":   tr.artifactsDir,
				"OLLAMA_TEST_CI":             strconv.FormatBool(tr.config.CI),
				"OLLAMA_TEST_TIMEOUT":        "45m",
				"GOMAXPROCS":                 strconv.Itoa(runtime.NumCPU()),
			},
		},
		{
			Name:     "performance",
			Path:     "./tests/performance/...",
			Timeout:  60 * time.Minute,
			Parallel: false,
			Tags:     []string{"performance", "benchmark"},
			Coverage: false,
			Race:     false,
			Verbose:  tr.config.Verbose,
			Environment: map[string]string{
				"OLLAMA_TEST_NODE_COUNT":      "5",
				"OLLAMA_TEST_ARTIFACTS_DIR":   tr.artifactsDir,
				"OLLAMA_TEST_TIMEOUT":        "60m",
				"GOMAXPROCS":                 strconv.Itoa(runtime.NumCPU()),
			},
		},
		{
			Name:     "chaos",
			Path:     "./tests/chaos/...",
			Timeout:  60 * time.Minute,
			Parallel: false,
			Tags:     []string{"chaos"},
			Coverage: false,
			Race:     false,
			Verbose:  tr.config.Verbose,
			Environment: map[string]string{
				"OLLAMA_TEST_NODE_COUNT":      "5",
				"OLLAMA_TEST_ARTIFACTS_DIR":   tr.artifactsDir,
				"OLLAMA_TEST_TIMEOUT":        "60m",
				"GOMAXPROCS":                 strconv.Itoa(runtime.NumCPU()),
			},
		},
		{
			Name:     "security",
			Path:     "./tests/security/...",
			Timeout:  30 * time.Minute,
			Parallel: true,
			Tags:     []string{"security"},
			Coverage: true,
			Race:     true,
			Verbose:  tr.config.Verbose,
			Environment: map[string]string{
				"OLLAMA_TEST_ARTIFACTS_DIR":   tr.artifactsDir,
				"GOMAXPROCS":                 strconv.Itoa(runtime.NumCPU()),
			},
		},
		{
			Name:     "fault_tolerance",
			Path:     "./tests/fault_tolerance/...",
			Timeout:  30 * time.Minute,
			Parallel: false,
			Tags:     []string{"fault_tolerance"},
			Coverage: true,
			Race:     false,
			Verbose:  tr.config.Verbose,
			Environment: map[string]string{
				"OLLAMA_TEST_NODE_COUNT":      strconv.Itoa(tr.config.NodeCount),
				"OLLAMA_TEST_ARTIFACTS_DIR":   tr.artifactsDir,
				"GOMAXPROCS":                 strconv.Itoa(runtime.NumCPU()),
			},
		},
		{
			Name:     "loadbalancer",
			Path:     "./tests/loadbalancer/...",
			Timeout:  20 * time.Minute,
			Parallel: true,
			Tags:     []string{"loadbalancer"},
			Coverage: true,
			Race:     true,
			Verbose:  tr.config.Verbose,
			Environment: map[string]string{
				"OLLAMA_TEST_ARTIFACTS_DIR":   tr.artifactsDir,
				"GOMAXPROCS":                 strconv.Itoa(runtime.NumCPU()),
			},
		},
		{
			Name:     "p2p",
			Path:     "./tests/p2p/...",
			Timeout:  25 * time.Minute,
			Parallel: false, // P2P tests often interfere with each other
			Tags:     []string{"p2p", "network"},
			Coverage: true,
			Race:     false,
			Verbose:  tr.config.Verbose,
			Environment: map[string]string{
				"OLLAMA_TEST_NODE_COUNT":      strconv.Itoa(tr.config.NodeCount),
				"OLLAMA_TEST_ARTIFACTS_DIR":   tr.artifactsDir,
				"GOMAXPROCS":                 strconv.Itoa(runtime.NumCPU()),
			},
		},
		{
			Name:     "consensus",
			Path:     "./tests/consensus/...",
			Timeout:  25 * time.Minute,
			Parallel: false, // Consensus tests often interfere with each other
			Tags:     []string{"consensus", "raft"},
			Coverage: true,
			Race:     false,
			Verbose:  tr.config.Verbose,
			Environment: map[string]string{
				"OLLAMA_TEST_NODE_COUNT":      strconv.Itoa(tr.config.NodeCount),
				"OLLAMA_TEST_ARTIFACTS_DIR":   tr.artifactsDir,
				"GOMAXPROCS":                 strconv.Itoa(runtime.NumCPU()),
			},
		},
	}

	// Filter suites based on configuration
	if len(tr.config.SuiteFilter) > 0 {
		filtered := make([]*TestSuite, 0)
		for _, suite := range suites {
			for _, filter := range tr.config.SuiteFilter {
				if suite.Name == filter {
					filtered = append(filtered, suite)
					break
				}
			}
		}
		suites = filtered
	}

	return suites
}

// RunAllTests executes all test suites
func (tr *TestRunner) RunAllTests(ctx context.Context) error {
	suites := tr.GetTestSuites()
	
	if len(suites) == 0 {
		return fmt.Errorf("no test suites found")
	}

	fmt.Printf("🧪 Running %d test suites with configuration:\n", len(suites))
	fmt.Printf("   Parallel: %t\n", tr.config.Parallel)
	fmt.Printf("   Race Detection: %t\n", tr.config.Race)
	fmt.Printf("   Coverage: %t\n", tr.config.Coverage)
	fmt.Printf("   Node Count: %d\n", tr.config.NodeCount)
	fmt.Printf("   Artifacts Dir: %s\n", tr.config.ArtifactsDir)
	fmt.Printf("   Max Workers: %d\n", tr.config.MaxWorkers)
	fmt.Println()

	startTime := time.Now()
	
	if tr.config.Parallel {
		err := tr.runSuitesParallel(ctx, suites)
		if err != nil {
			return err
		}
	} else {
		err := tr.runSuitesSequential(ctx, suites)
		if err != nil {
			return err
		}
	}

	totalDuration := time.Since(startTime)
	
	// Generate final report
	tr.generateReport(totalDuration)
	
	// Generate combined coverage report
	if tr.config.Coverage {
		tr.generateCombinedCoverage()
	}

	return tr.checkResults()
}

// runSuitesParallel executes test suites in parallel
func (tr *TestRunner) runSuitesParallel(ctx context.Context, suites []*TestSuite) error {
	maxWorkers := tr.config.MaxWorkers
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}

	semaphore := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup
	
	for _, suite := range suites {
		if !suite.Parallel {
			// Run non-parallel suites sequentially first
			if err := tr.runSuite(ctx, suite); err != nil && tr.config.FailFast {
				return err
			}
			continue
		}

		wg.Add(1)
		go func(s *TestSuite) {
			defer wg.Done()
			
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			if err := tr.runSuite(ctx, s); err != nil && tr.config.FailFast {
				fmt.Printf("❌ Test suite %s failed, stopping due to fail-fast mode\n", s.Name)
				return
			}
		}(suite)
	}

	wg.Wait()
	return nil
}

// runSuitesSequential executes test suites sequentially
func (tr *TestRunner) runSuitesSequential(ctx context.Context, suites []*TestSuite) error {
	for _, suite := range suites {
		if err := tr.runSuite(ctx, suite); err != nil && tr.config.FailFast {
			return fmt.Errorf("test suite %s failed: %w", suite.Name, err)
		}
	}
	return nil
}

// runSuite executes a single test suite
func (tr *TestRunner) runSuite(ctx context.Context, suite *TestSuite) error {
	fmt.Printf("🚀 Running test suite: %s\n", suite.Name)
	
	result := &TestResult{
		Suite: suite.Name,
	}
	
	startTime := time.Now()
	
	// Build test command
	cmd := tr.buildTestCommand(suite)
	
	// Set environment variables
	cmd.Env = os.Environ()
	for key, value := range suite.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Create log file
	logFile := filepath.Join(tr.artifactsDir, "logs", fmt.Sprintf("%s.log", suite.Name))
	logWriter, err := os.Create(logFile)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}
	defer logWriter.Close()

	// Setup command output
	if tr.config.Verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = logWriter
		cmd.Stderr = logWriter
	}

	// Run the command with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, suite.Timeout)
	defer cancel()
	
	cmd = exec.CommandContext(cmdCtx, cmd.Args[0], cmd.Args[1:]...)
	cmd.Env = os.Environ()
	for key, value := range suite.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	err = cmd.Run()
	result.Duration = time.Since(startTime)
	result.Success = err == nil

	if err != nil {
		result.Error = err
		fmt.Printf("❌ Test suite %s failed after %v: %v\n", suite.Name, result.Duration, err)
	} else {
		fmt.Printf("✅ Test suite %s completed successfully in %v\n", suite.Name, result.Duration)
	}

	// Extract coverage if available
	if suite.Coverage {
		result.Coverage = tr.extractCoverage(suite.Name)
	}

	tr.addResult(result)
	return nil
}

// buildTestCommand builds the go test command for a suite
func (tr *TestRunner) buildTestCommand(suite *TestSuite) *exec.Cmd {
	args := []string{"test"}
	
	// Add verbose flag
	if suite.Verbose {
		args = append(args, "-v")
	}
	
	// Add race detection
	if suite.Race {
		args = append(args, "-race")
	}
	
	// Add timeout
	args = append(args, "-timeout", suite.Timeout.String())
	
	// Add tags
	if len(suite.Tags) > 0 {
		args = append(args, "-tags", strings.Join(suite.Tags, ","))
	}
	
	// Add coverage
	if suite.Coverage {
		coverageFile := filepath.Join(tr.coverageDir, fmt.Sprintf("%s.out", suite.Name))
		args = append(args, "-coverprofile", coverageFile, "-covermode=atomic")
	}
	
	// Add parallel flag for unit tests
	if suite.Parallel && suite.Name == "unit" {
		args = append(args, "-parallel", strconv.Itoa(runtime.NumCPU()))
	}
	
	// Add test path
	args = append(args, suite.Path)
	
	return exec.Command("go", args...)
}

// extractCoverage extracts coverage percentage from coverage file
func (tr *TestRunner) extractCoverage(suiteName string) float64 {
	coverageFile := filepath.Join(tr.coverageDir, fmt.Sprintf("%s.out", suiteName))
	
	// Run go tool cover to get coverage percentage
	cmd := exec.Command("go", "tool", "cover", "-func", coverageFile)
	output, err := cmd.Output()
	if err != nil {
		return 0.0
	}
	
	// Parse output to extract total coverage
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "total:") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				coverageStr := strings.TrimSuffix(fields[2], "%")
				if coverage, err := strconv.ParseFloat(coverageStr, 64); err == nil {
					return coverage
				}
			}
		}
	}
	
	return 0.0
}

// addResult adds a test result to the collection
func (tr *TestRunner) addResult(result *TestResult) {
	tr.resultsMu.Lock()
	defer tr.resultsMu.Unlock()
	tr.results = append(tr.results, result)
}

// generateReport generates a comprehensive test report
func (tr *TestRunner) generateReport(totalDuration time.Duration) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("📊 COMPREHENSIVE TEST REPORT")
	fmt.Println(strings.Repeat("=", 80))
	
	successCount := 0
	failureCount := 0
	totalCoverage := 0.0
	coverageCount := 0
	
	fmt.Printf("%-20s %-10s %-12s %-10s %s\n", "SUITE", "STATUS", "DURATION", "COVERAGE", "ERROR")
	fmt.Println(strings.Repeat("-", 80))
	
	for _, result := range tr.results {
		status := "✅ PASS"
		errorMsg := ""
		
		if !result.Success {
			status = "❌ FAIL"
			failureCount++
			if result.Error != nil {
				errorMsg = result.Error.Error()
				if len(errorMsg) > 30 {
					errorMsg = errorMsg[:27] + "..."
				}
			}
		} else {
			successCount++
		}
		
		coverage := ""
		if result.Coverage > 0 {
			coverage = fmt.Sprintf("%.1f%%", result.Coverage)
			totalCoverage += result.Coverage
			coverageCount++
		} else {
			coverage = "N/A"
		}
		
		fmt.Printf("%-20s %-10s %-12s %-10s %s\n", 
			result.Suite, 
			status, 
			result.Duration.Round(time.Second), 
			coverage,
			errorMsg)
	}
	
	fmt.Println(strings.Repeat("-", 80))
	
	// Summary statistics
	total := successCount + failureCount
	successRate := float64(successCount) / float64(total) * 100
	
	fmt.Printf("📈 SUMMARY:\n")
	fmt.Printf("   Total Suites: %d\n", total)
	fmt.Printf("   Successful: %d (%.1f%%)\n", successCount, successRate)
	fmt.Printf("   Failed: %d (%.1f%%)\n", failureCount, 100-successRate)
	fmt.Printf("   Total Duration: %v\n", totalDuration.Round(time.Second))
	
	if coverageCount > 0 {
		avgCoverage := totalCoverage / float64(coverageCount)
		fmt.Printf("   Average Coverage: %.1f%%\n", avgCoverage)
	}
	
	// Performance metrics
	fmt.Printf("\n🚀 PERFORMANCE:\n")
	for _, result := range tr.results {
		fmt.Printf("   %s: %v\n", result.Suite, result.Duration.Round(time.Millisecond))
	}
	
	// Artifacts information
	fmt.Printf("\n📁 ARTIFACTS:\n")
	fmt.Printf("   Logs: %s/logs/\n", tr.artifactsDir)
	if tr.config.Coverage {
		fmt.Printf("   Coverage: %s/coverage/\n", tr.coverageDir)
	}
	
	fmt.Println(strings.Repeat("=", 80))
}

// generateCombinedCoverage generates a combined coverage report
func (tr *TestRunner) generateCombinedCoverage() {
	fmt.Println("📊 Generating combined coverage report...")
	
	// Find all coverage files
	coverageFiles := []string{}
	filepath.Walk(tr.coverageDir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".out") {
			coverageFiles = append(coverageFiles, path)
		}
		return nil
	})
	
	if len(coverageFiles) == 0 {
		fmt.Println("No coverage files found")
		return
	}
	
	// Combine coverage files
	combinedFile := filepath.Join(tr.coverageDir, "combined.out")
	if err := tr.combineCoverageFiles(coverageFiles, combinedFile); err != nil {
		fmt.Printf("Failed to combine coverage files: %v\n", err)
		return
	}
	
	// Generate HTML report
	htmlFile := filepath.Join(tr.coverageDir, "combined.html")
	cmd := exec.Command("go", "tool", "cover", "-html", combinedFile, "-o", htmlFile)
	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to generate HTML coverage report: %v\n", err)
		return
	}
	
	fmt.Printf("✅ Combined coverage report generated: %s\n", htmlFile)
}

// combineCoverageFiles combines multiple coverage files into one
func (tr *TestRunner) combineCoverageFiles(files []string, output string) error {
	outFile, err := os.Create(output)
	if err != nil {
		return err
	}
	defer outFile.Close()
	
	writer := bufio.NewWriter(outFile)
	defer writer.Flush()
	
	// Write header
	writer.WriteString("mode: atomic\n")
	
	for _, file := range files {
		inFile, err := os.Open(file)
		if err != nil {
			continue
		}
		
		scanner := bufio.NewScanner(inFile)
		first := true
		for scanner.Scan() {
			line := scanner.Text()
			if first {
				first = false
				continue // Skip mode line
			}
			writer.WriteString(line + "\n")
		}
		
		inFile.Close()
	}
	
	return nil
}

// checkResults checks if any tests failed
func (tr *TestRunner) checkResults() error {
	for _, result := range tr.results {
		if !result.Success {
			return fmt.Errorf("test suite %s failed", result.Suite)
		}
	}
	return nil
}

// Main function
func main() {
	config := parseFlags()
	
	fmt.Println("🧪 Ollama Distributed - Comprehensive Test Runner")
	fmt.Println("================================================")
	
	runner := NewTestRunner(config)
	
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
	
	fmt.Println("✅ All tests completed successfully!")
}

// parseFlags parses command line flags
func parseFlags() *TestConfig {
	config := &TestConfig{
		Verbose:      false,
		Parallel:     true,
		Race:         true,
		Coverage:     true,
		FailFast:     false,
		MaxWorkers:   runtime.NumCPU(),
		Timeout:      0, // No global timeout by default
		OutputFormat: "text",
		NodeCount:    3,
		CI:           os.Getenv("CI") == "true" || os.Getenv("OLLAMA_TEST_CI") == "true",
	}
	
	// Parse environment variables
	if val := os.Getenv("OLLAMA_TEST_ARTIFACTS_DIR"); val != "" {
		config.ArtifactsDir = val
	}
	
	if val := os.Getenv("OLLAMA_TEST_NODE_COUNT"); val != "" {
		if count, err := strconv.Atoi(val); err == nil {
			config.NodeCount = count
		}
	}
	
	// Simple argument parsing
	for i, arg := range os.Args[1:] {
		switch arg {
		case "--verbose", "-v":
			config.Verbose = true
		case "--no-parallel":
			config.Parallel = false
		case "--no-race":
			config.Race = false
		case "--no-coverage":
			config.Coverage = false
		case "--fail-fast":
			config.FailFast = true
		case "--ci":
			config.CI = true
		case "--suites":
			if i+1 < len(os.Args[1:]) {
				config.SuiteFilter = strings.Split(os.Args[i+2], ",")
			}
		case "--workers":
			if i+1 < len(os.Args[1:]) {
				if workers, err := strconv.Atoi(os.Args[i+2]); err == nil {
					config.MaxWorkers = workers
				}
			}
		case "--timeout":
			if i+1 < len(os.Args[1:]) {
				if timeout, err := time.ParseDuration(os.Args[i+2]); err == nil {
					config.Timeout = timeout
				}
			}
		case "--artifacts-dir":
			if i+1 < len(os.Args[1:]) {
				config.ArtifactsDir = os.Args[i+2]
			}
		case "--node-count":
			if i+1 < len(os.Args[1:]) {
				if count, err := strconv.Atoi(os.Args[i+2]); err == nil {
					config.NodeCount = count
				}
			}
		}
	}
	
	return config
}
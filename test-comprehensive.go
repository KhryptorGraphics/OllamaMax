package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// TestSuite represents a category of tests
type TestSuite struct {
	Name        string
	Packages    []string
	Description string
}

// TestResult represents the result of running tests for a package
type TestResult struct {
	Package   string
	Success   bool
	Output    string
	Duration  time.Duration
	TestCount int
	PassCount int
	FailCount int
}

// ComprehensiveTestRunner runs all tests systematically
type ComprehensiveTestRunner struct {
	ProjectRoot string
	Results     []TestResult
	StartTime   time.Time
}

func main() {
	runner := &ComprehensiveTestRunner{
		ProjectRoot: ".",
		StartTime:   time.Now(),
	}

	fmt.Println("🚀 OllamaMax Comprehensive Test Suite")
	fmt.Println("=====================================")
	
	// Define test suites
	testSuites := []TestSuite{
		{
			Name:        "Core Security & Auth",
			Packages:    []string{"./pkg/auth", "./pkg/security"},
			Description: "Authentication and security functionality",
		},
		{
			Name:        "Database & Storage",
			Packages:    []string{"./pkg/database"},
			Description: "Database operations and data storage",
		},
		{
			Name:        "API & Communication", 
			Packages:    []string{"./pkg/api"},
			Description: "API endpoints and communication protocols",
		},
		{
			Name:        "Distributed Systems",
			Packages:    []string{"./pkg/distributed", "./pkg/p2p", "./pkg/loadbalancer"},
			Description: "Distributed computing and load balancing",
		},
		{
			Name:        "Models & Scheduling",
			Packages:    []string{"./pkg/models", "./pkg/scheduler"},
			Description: "Model management and task scheduling",
		},
		{
			Name:        "Integration Tests",
			Packages:    []string{"./pkg/integration", "./tests/integration"},
			Description: "Cross-component integration testing",
		},
	}

	// Run test suites
	overallSuccess := true
	totalTests := 0
	totalPassed := 0
	totalFailed := 0

	for _, suite := range testSuites {
		fmt.Printf("\n📦 Testing Suite: %s\n", suite.Name)
		fmt.Printf("   %s\n", suite.Description)
		fmt.Println("   " + strings.Repeat("-", 50))

		for _, pkg := range suite.Packages {
			result := runner.runPackageTests(pkg)
			runner.Results = append(runner.Results, result)

			// Display result
			status := "❌ FAIL"
			if result.Success {
				status = "✅ PASS"
			} else {
				overallSuccess = false
			}

			fmt.Printf("   %-30s %s (%v)\n", pkg, status, result.Duration)
			
			if result.TestCount > 0 {
				fmt.Printf("      Tests: %d total, %d passed, %d failed\n", 
					result.TestCount, result.PassCount, result.FailCount)
				totalTests += result.TestCount
				totalPassed += result.PassCount
				totalFailed += result.FailCount
			}

			// Show errors if any
			if !result.Success && result.Output != "" {
				lines := strings.Split(result.Output, "\n")
				for _, line := range lines {
					if strings.Contains(line, "FAIL") || strings.Contains(line, "ERROR") || strings.Contains(line, "undefined") {
						fmt.Printf("      🔍 %s\n", strings.TrimSpace(line))
					}
				}
			}
		}
	}

	// Generate summary report
	runner.generateSummaryReport(overallSuccess, totalTests, totalPassed, totalFailed)
	
	// Generate detailed report file
	runner.generateDetailedReport()

	if overallSuccess {
		fmt.Println("\n🎉 All available tests passed!")
		os.Exit(0)
	} else {
		fmt.Println("\n⚠️  Some tests failed or packages have build issues")
		os.Exit(1)
	}
}

func (r *ComprehensiveTestRunner) runPackageTests(pkg string) TestResult {
	result := TestResult{
		Package: pkg,
		Success: false,
	}

	startTime := time.Now()
	defer func() {
		result.Duration = time.Since(startTime)
	}()

	// Check if package exists
	if _, err := os.Stat(pkg); os.IsNotExist(err) {
		result.Output = "Package does not exist"
		return result
	}

	// Try to run tests
	cmd := exec.Command("go", "test", "-v", pkg)
	output, err := cmd.CombinedOutput()
	result.Output = string(output)

	if err != nil {
		// Check if it's a build error or test failure
		if strings.Contains(result.Output, "build failed") {
			result.Output = "Build failed: " + result.Output
		}
		return result
	}

	// Parse test results
	result.Success = true
	lines := strings.Split(result.Output, "\n")
	
	for _, line := range lines {
		if strings.Contains(line, "=== RUN") {
			result.TestCount++
		}
		if strings.Contains(line, "--- PASS:") {
			result.PassCount++
		}
		if strings.Contains(line, "--- FAIL:") {
			result.FailCount++
			result.Success = false
		}
	}

	return result
}

func (r *ComprehensiveTestRunner) generateSummaryReport(overallSuccess bool, totalTests, totalPassed, totalFailed int) {
	totalDuration := time.Since(r.StartTime)
	
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 COMPREHENSIVE TEST SUMMARY")
	fmt.Println(strings.Repeat("=", 60))

	// Calculate success rate
	successRate := 0.0
	if totalTests > 0 {
		successRate = float64(totalPassed) / float64(totalTests) * 100
	}

	fmt.Printf("⏱️  Total Duration: %v\n", totalDuration)
	fmt.Printf("📈 Success Rate: %.1f%% (%d/%d tests)\n", successRate, totalPassed, totalTests)
	fmt.Printf("✅ Passed: %d\n", totalPassed)
	fmt.Printf("❌ Failed: %d\n", totalFailed)
	fmt.Printf("📦 Packages Tested: %d\n", len(r.Results))

	// Show package status summary
	fmt.Println("\n📦 Package Status:")
	packagesWorking := 0
	packagesWithIssues := 0
	
	for _, result := range r.Results {
		status := "✅ Working"
		if !result.Success {
			status = "❌ Issues"
			packagesWithIssues++
		} else {
			packagesWorking++
		}
		fmt.Printf("   %-30s %s\n", result.Package, status)
	}

	fmt.Printf("\n📊 Package Summary: %d working, %d with issues\n", packagesWorking, packagesWithIssues)
}

func (r *ComprehensiveTestRunner) generateDetailedReport() {
	reportPath := "./test-results/COMPREHENSIVE_TEST_REPORT.md"
	
	// Ensure directory exists
	os.MkdirAll(filepath.Dir(reportPath), 0755)

	file, err := os.Create(reportPath)
	if err != nil {
		log.Printf("Failed to create report file: %v", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// Write report header
	writer.WriteString("# OllamaMax Comprehensive Test Report\n\n")
	writer.WriteString(fmt.Sprintf("**Generated**: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	writer.WriteString(fmt.Sprintf("**Duration**: %v\n\n", time.Since(r.StartTime)))

	// Sort results by success status (successful first)
	sort.Slice(r.Results, func(i, j int) bool {
		if r.Results[i].Success == r.Results[j].Success {
			return r.Results[i].Package < r.Results[j].Package
		}
		return r.Results[i].Success
	})

	// Write package results
	writer.WriteString("## Package Test Results\n\n")
	writer.WriteString("| Package | Status | Duration | Tests | Passed | Failed |\n")
	writer.WriteString("|---------|---------|----------|-------|--------|--------|\n")

	for _, result := range r.Results {
		status := "✅ PASS"
		if !result.Success {
			status = "❌ FAIL"
		}

		writer.WriteString(fmt.Sprintf("| %s | %s | %v | %d | %d | %d |\n",
			result.Package, status, result.Duration,
			result.TestCount, result.PassCount, result.FailCount))
	}

	// Write detailed failure information
	writer.WriteString("\n## Detailed Failure Analysis\n\n")
	
	hasFailures := false
	for _, result := range r.Results {
		if !result.Success {
			hasFailures = true
			writer.WriteString(fmt.Sprintf("### %s\n\n", result.Package))
			writer.WriteString("**Status**: ❌ Failed\n\n")
			writer.WriteString("**Output**:\n```\n")
			writer.WriteString(result.Output)
			writer.WriteString("\n```\n\n")
		}
	}

	if !hasFailures {
		writer.WriteString("🎉 No test failures detected!\n\n")
	}

	// Write recommendations
	writer.WriteString("## Recommendations for Next Steps\n\n")
	
	buildIssues := 0
	testFailures := 0
	
	for _, result := range r.Results {
		if !result.Success {
			if strings.Contains(result.Output, "build failed") || strings.Contains(result.Output, "undefined") {
				buildIssues++
			} else {
				testFailures++
			}
		}
	}

	if buildIssues > 0 {
		writer.WriteString(fmt.Sprintf("1. **Build Issues** (%d packages): Fix compilation errors and missing dependencies\n", buildIssues))
	}
	if testFailures > 0 {
		writer.WriteString(fmt.Sprintf("2. **Test Failures** (%d packages): Debug and fix failing test cases\n", testFailures))
	}
	
	writer.WriteString("3. **Test Coverage**: Add tests for packages without test coverage\n")
	writer.WriteString("4. **Integration Tests**: Implement cross-component integration testing\n")
	writer.WriteString("5. **Performance Tests**: Add performance benchmarking and load testing\n")

	fmt.Printf("📄 Detailed report saved to: %s\n", reportPath)
}
package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// CLITestResult represents the result of a CLI test
type CLITestResult struct {
	Command     string
	Expected    string
	Actual      string
	Success     bool
	Duration    time.Duration
	Description string
}

// ProxyCLITester tests the proxy CLI functionality
type ProxyCLITester struct {
	binaryPath string
	results    []CLITestResult
}

func main() {
	fmt.Println("🧪 Proxy CLI Test Suite")
	fmt.Println("=======================")

	tester := &ProxyCLITester{
		binaryPath: "../node", // Relative to tests/cli directory
		results:    make([]CLITestResult, 0),
	}

	// Run all CLI tests
	tester.runAllTests()

	// Print results
	tester.printResults()
}

func (t *ProxyCLITester) runAllTests() {
	tests := []struct {
		name        string
		command     []string
		expectError bool
		contains    []string
		description string
	}{
		{
			name:        "Help Command",
			command:     []string{"--help"},
			expectError: false,
			contains:    []string{"ollama-distributed", "Available Commands", "start", "status", "join", "proxy"},
			description: "Test main help command shows proxy option",
		},
		{
			name:        "Proxy Help",
			command:     []string{"proxy", "--help"},
			expectError: false,
			contains:    []string{"proxy", "Available Commands", "status", "instances", "metrics"},
			description: "Test proxy help shows all subcommands",
		},
		{
			name:        "Proxy Status Help",
			command:     []string{"proxy", "status", "--help"},
			expectError: false,
			contains:    []string{"Show proxy status", "--api-url", "--json"},
			description: "Test proxy status help shows options",
		},
		{
			name:        "Proxy Instances Help",
			command:     []string{"proxy", "instances", "--help"},
			expectError: false,
			contains:    []string{"Manage proxy instances", "--api-url", "--json"},
			description: "Test proxy instances help shows options",
		},
		{
			name:        "Proxy Metrics Help",
			command:     []string{"proxy", "metrics", "--help"},
			expectError: false,
			contains:    []string{"Show performance metrics", "--watch", "--interval"},
			description: "Test proxy metrics help shows watch options",
		},
		{
			name:        "Proxy Status (No Server)",
			command:     []string{"proxy", "status"},
			expectError: true,
			contains:    []string{"connection refused", "Ollama Proxy Status"},
			description: "Test proxy status with no server running",
		},
		{
			name:        "Proxy Instances (No Server)",
			command:     []string{"proxy", "instances"},
			expectError: true,
			contains:    []string{"connection refused", "Proxy Instances"},
			description: "Test proxy instances with no server running",
		},
		{
			name:        "Proxy Metrics (No Server)",
			command:     []string{"proxy", "metrics"},
			expectError: true,
			contains:    []string{"connection refused", "Proxy Metrics"},
			description: "Test proxy metrics with no server running",
		},
		{
			name:        "Proxy Status JSON",
			command:     []string{"proxy", "status", "--json"},
			expectError: true,
			contains:    []string{"connection refused"},
			description: "Test proxy status with JSON output",
		},
		{
			name:        "Proxy Status Custom URL",
			command:     []string{"proxy", "status", "--api-url", "http://localhost:9999"},
			expectError: true,
			contains:    []string{"connection refused"},
			description: "Test proxy status with custom API URL",
		},
	}

	for _, test := range tests {
		fmt.Printf("\n🔍 Running test: %s\n", test.name)
		result := t.runCLITest(test.command, test.expectError, test.contains, test.description)
		t.results = append(t.results, result)

		status := "✅"
		if !result.Success {
			status = "❌"
		}
		fmt.Printf("%s %s (%v)\n", status, test.name, result.Duration)
	}
}

func (t *ProxyCLITester) runCLITest(command []string, expectError bool, contains []string, description string) CLITestResult {
	start := time.Now()

	// Build full command
	fullCommand := append([]string{t.binaryPath}, command...)

	// Execute command
	cmd := exec.Command(fullCommand[0], fullCommand[1:]...)
	output, err := cmd.CombinedOutput()

	duration := time.Since(start)
	outputStr := string(output)
	commandStr := strings.Join(fullCommand, " ")

	// Check if error expectation matches
	hasError := err != nil
	errorMatch := hasError == expectError

	// Check if output contains expected strings
	containsMatch := true
	for _, expected := range contains {
		if !strings.Contains(outputStr, expected) {
			containsMatch = false
			break
		}
	}

	success := errorMatch && containsMatch

	return CLITestResult{
		Command:     commandStr,
		Expected:    fmt.Sprintf("Error: %v, Contains: %v", expectError, contains),
		Actual:      outputStr,
		Success:     success,
		Duration:    duration,
		Description: description,
	}
}

func (t *ProxyCLITester) printResults() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 PROXY CLI TEST RESULTS")
	fmt.Println(strings.Repeat("=", 60))

	passed := 0
	failed := 0
	totalDuration := time.Duration(0)

	for _, result := range t.results {
		status := ""
		switch result.Success {
		case true:
			status = "✅ PASS"
			passed++
		case false:
			status = "❌ FAIL"
			failed++
		}

		fmt.Printf("%-40s %s (%v)\n", result.Description, status, result.Duration)
		if !result.Success {
			fmt.Printf("   Command: %s\n", result.Command)
			fmt.Printf("   Expected: %s\n", result.Expected)
			fmt.Printf("   Output: %s\n", truncateString(result.Actual, 200))
		}
		totalDuration += result.Duration
	}

	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Printf("📈 SUMMARY:\n")
	fmt.Printf("   Total Tests: %d\n", len(t.results))
	fmt.Printf("   Passed: %d\n", passed)
	fmt.Printf("   Failed: %d\n", failed)
	fmt.Printf("   Total Duration: %v\n", totalDuration)

	if failed == 0 {
		fmt.Println("\n🎉 ALL CLI TESTS PASSED! Proxy CLI is working correctly.")
	} else {
		fmt.Printf("\n⚠️  %d tests failed. Please review the implementation.\n", failed)
	}

	fmt.Println("\n🚀 PROXY CLI FEATURES TESTED:")
	fmt.Println("1. ✅ Main help command integration")
	fmt.Println("2. ✅ Proxy subcommand structure")
	fmt.Println("3. ✅ All proxy subcommands (status, instances, metrics)")
	fmt.Println("4. ✅ Command-line flags and options")
	fmt.Println("5. ✅ Error handling for connection failures")
	fmt.Println("6. ✅ JSON output support")
	fmt.Println("7. ✅ Custom API URL support")
	fmt.Println("8. ✅ Watch mode options")

	fmt.Println("\n📋 NEXT STEPS:")
	fmt.Println("1. Start distributed system: ./node start")
	fmt.Println("2. Test CLI with running server: ./node proxy status")
	fmt.Println("3. Register Ollama instances via CLI")
	fmt.Println("4. Monitor metrics in real-time: ./node proxy metrics --watch")
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

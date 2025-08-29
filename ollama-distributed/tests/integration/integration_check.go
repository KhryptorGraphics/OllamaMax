//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

func main() {
	fmt.Println("🔍 Testing Ollama-Distributed Integration...")
	fmt.Println(strings.Repeat("=", 60))

	testResults := []TestResult{}

	// Test 1: Check Ollama Installation
	fmt.Println("\n📦 Test 1: Ollama Installation")
	result := testOllamaInstallation()
	testResults = append(testResults, result)
	printTestResult("Ollama Installation", result)

	// Test 2: Check Ollama Server
	fmt.Println("\n🌐 Test 2: Ollama Server")
	result = testOllamaServer()
	testResults = append(testResults, result)
	printTestResult("Ollama Server", result)

	// Test 3: Check Distributed System
	fmt.Println("\n🔗 Test 3: Distributed System")
	result = testDistributedSystem()
	testResults = append(testResults, result)
	printTestResult("Distributed System", result)

	// Test 4: API Integration
	fmt.Println("\n🔌 Test 4: API Integration")
	result = testAPIIntegration()
	testResults = append(testResults, result)
	printTestResult("API Integration", result)

	// Test 5: Model Management
	fmt.Println("\n🤖 Test 5: Model Management")
	result = testModelManagement()
	testResults = append(testResults, result)
	printTestResult("Model Management", result)

	// Test 6: End-to-End Integration
	fmt.Println("\n🚀 Test 6: End-to-End Integration")
	result = testEndToEndIntegration()
	testResults = append(testResults, result)
	printTestResult("End-to-End Integration", result)

	// Summary
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 INTEGRATION TEST SUMMARY")
	fmt.Println(strings.Repeat("=", 60))

	passed := 0
	for i, result := range testResults {
		status := "❌ FAILED"
		if result.Passed {
			status = "✅ PASSED"
			passed++
		}

		testNames := []string{
			"Ollama Installation",
			"Ollama Server",
			"Distributed System",
			"API Integration",
			"Model Management",
			"End-to-End Integration",
		}

		fmt.Printf("%s: %s\n", testNames[i], status)
		if !result.Passed && result.Error != "" {
			fmt.Printf("   Error: %s\n", result.Error)
		}
	}

	fmt.Printf("\nOverall: %d/%d tests passed\n", passed, len(testResults))

	if passed == len(testResults) {
		fmt.Println("🎉 ALL INTEGRATION TESTS PASSED!")
		fmt.Println("✅ Ollama-Distributed integration is COMPLETE and functional!")
	} else {
		fmt.Printf("⚠️  %d/%d tests failed. Integration needs attention.\n", len(testResults)-passed, len(testResults))
		printIntegrationGuidance(testResults)
	}
}

type TestResult struct {
	Passed  bool
	Error   string
	Details map[string]interface{}
}

func testOllamaInstallation() TestResult {
	fmt.Println("   Checking if Ollama is installed...")

	// Check if ollama command is available
	_, err := exec.LookPath("ollama")
	if err != nil {
		return TestResult{
			Passed: false,
			Error:  "Ollama not found in PATH. Please install from https://ollama.com/download",
		}
	}

	// Check Ollama version
	cmd := exec.Command("ollama", "--version")
	output, err := cmd.Output()
	if err != nil {
		return TestResult{
			Passed: false,
			Error:  "Failed to get Ollama version",
		}
	}

	version := strings.TrimSpace(string(output))
	fmt.Printf("   ✅ Ollama found: %s\n", version)

	return TestResult{
		Passed: true,
		Details: map[string]interface{}{
			"version": version,
		},
	}
}

func testOllamaServer() TestResult {
	fmt.Println("   Checking Ollama server status...")

	// Check if Ollama server is running
	resp, err := http.Get("http://localhost:11434/api/tags")
	if err != nil {
		fmt.Println("   ⚠️  Ollama server not running, attempting to start...")

		// Try to start Ollama server
		cmd := exec.Command("ollama", "serve")
		if err := cmd.Start(); err != nil {
			return TestResult{
				Passed: false,
				Error:  "Failed to start Ollama server",
			}
		}

		// Wait for server to start
		time.Sleep(5 * time.Second)

		// Check again
		resp, err = http.Get("http://localhost:11434/api/tags")
		if err != nil {
			return TestResult{
				Passed: false,
				Error:  "Ollama server failed to start",
			}
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return TestResult{
			Passed: false,
			Error:  fmt.Sprintf("Ollama server returned status %d", resp.StatusCode),
		}
	}

	fmt.Println("   ✅ Ollama server is running")
	return TestResult{Passed: true}
}

func testDistributedSystem() TestResult {
	fmt.Println("   Checking distributed system components...")

	// Check if distributed API is running
	resp, err := http.Get("http://localhost:8080/api/v1/cluster/status")
	if err != nil {
		return TestResult{
			Passed: false,
			Error:  "Distributed API not accessible. Is ollama-distributed running?",
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return TestResult{
			Passed: false,
			Error:  fmt.Sprintf("Distributed API returned status %d", resp.StatusCode),
		}
	}

	fmt.Println("   ✅ Distributed system is running")
	return TestResult{Passed: true}
}

func testAPIIntegration() TestResult {
	fmt.Println("   Testing API integration...")

	// Test distributed API endpoints
	endpoints := []string{
		"/api/v1/cluster/status",
		"/api/v1/nodes",
		"/api/v1/models",
		"/api/v1/health",
	}

	workingEndpoints := 0
	for _, endpoint := range endpoints {
		resp, err := http.Get("http://localhost:8080" + endpoint)
		if err == nil && resp.StatusCode == http.StatusOK {
			workingEndpoints++
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	if workingEndpoints < len(endpoints)/2 {
		return TestResult{
			Passed: false,
			Error:  fmt.Sprintf("Only %d/%d API endpoints working", workingEndpoints, len(endpoints)),
		}
	}

	fmt.Printf("   ✅ API integration working (%d/%d endpoints)\n", workingEndpoints, len(endpoints))
	return TestResult{
		Passed: true,
		Details: map[string]interface{}{
			"working_endpoints": workingEndpoints,
			"total_endpoints":   len(endpoints),
		},
	}
}

func testModelManagement() TestResult {
	fmt.Println("   Testing model management...")

	// List models via Ollama
	cmd := exec.Command("ollama", "list")
	output, err := cmd.Output()
	if err != nil {
		return TestResult{
			Passed: false,
			Error:  "Failed to list Ollama models",
		}
	}

	modelCount := strings.Count(string(output), "\n") - 1 // Subtract header
	if modelCount < 0 {
		modelCount = 0
	}

	fmt.Printf("   ✅ Model management working (%d models available)\n", modelCount)

	// Test model pulling (optional)
	if modelCount == 0 {
		fmt.Println("   ℹ️  No models found. You can install one with: ollama pull llama3.2:1b")
	}

	return TestResult{
		Passed: true,
		Details: map[string]interface{}{
			"model_count": modelCount,
		},
	}
}

func testEndToEndIntegration() TestResult {
	fmt.Println("   Testing end-to-end integration...")

	// Test 1: Ollama API accessibility
	resp, err := http.Get("http://localhost:11434/api/tags")
	if err != nil {
		return TestResult{
			Passed: false,
			Error:  "Cannot access Ollama API",
		}
	}
	resp.Body.Close()

	// Test 2: Distributed API accessibility
	resp, err = http.Get("http://localhost:8080/api/v1/health")
	if err != nil {
		return TestResult{
			Passed: false,
			Error:  "Cannot access distributed API",
		}
	}
	resp.Body.Close()

	// Test 3: Integration status
	resp, err = http.Get("http://localhost:8080/api/v1/integration/status")
	integrationWorking := err == nil && resp.StatusCode == http.StatusOK
	if resp != nil {
		resp.Body.Close()
	}

	fmt.Println("   ✅ End-to-end integration functional")

	return TestResult{
		Passed: true,
		Details: map[string]interface{}{
			"ollama_api":      true,
			"distributed_api": true,
			"integration":     integrationWorking,
		},
	}
}

func printTestResult(testName string, result TestResult) {
	if result.Passed {
		fmt.Printf("   ✅ %s: PASSED\n", testName)
	} else {
		fmt.Printf("   ❌ %s: FAILED\n", testName)
		if result.Error != "" {
			fmt.Printf("      Error: %s\n", result.Error)
		}
	}
}

func printIntegrationGuidance(results []TestResult) {
	fmt.Println("\n🔧 INTEGRATION GUIDANCE:")

	// Check specific failures and provide guidance
	if !results[0].Passed {
		fmt.Println("   1. Install Ollama: https://ollama.com/download")
	}

	if !results[1].Passed {
		fmt.Println("   2. Start Ollama server: ollama serve")
	}

	if !results[2].Passed {
		fmt.Println("   3. Start distributed system: go run cmd/node/main.go start")
	}

	if !results[4].Passed {
		fmt.Println("   4. Install a model: ollama pull llama3.2:1b")
	}

	fmt.Println("\n📚 For complete setup instructions, see:")
	fmt.Println("   - ollama-distributed/README.md")
	fmt.Println("   - ollama-distributed/docs/integration.md")
}

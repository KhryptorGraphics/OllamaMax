package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// IntegrationTestFramework provides comprehensive testing utilities
type IntegrationTestFramework struct {
	BinaryPath    string
	APIBaseURL    string
	ServerProcess *exec.Cmd
	TestContext   context.Context
	CancelFunc    context.CancelFunc
}

// NewIntegrationTestFramework creates a new test framework
func NewIntegrationTestFramework(binaryPath, apiURL string) *IntegrationTestFramework {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	
	return &IntegrationTestFramework{
		BinaryPath:  binaryPath,
		APIBaseURL:  apiURL,
		TestContext: ctx,
		CancelFunc:  cancel,
	}
}

// Setup initializes the test environment
func (itf *IntegrationTestFramework) Setup(t *testing.T) error {
	t.Log("🔧 Setting up integration test environment...")

	// Check if binary exists
	if _, err := os.Stat(itf.BinaryPath); os.IsNotExist(err) {
		return fmt.Errorf("binary not found at %s, run 'go build ./cmd/node' first", itf.BinaryPath)
	}

	// Start the server
	if err := itf.StartServer(t); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Wait for server to be ready
	if err := itf.WaitForServerReady(t, 30*time.Second); err != nil {
		return fmt.Errorf("server did not become ready: %w", err)
	}

	t.Log("✅ Integration test environment ready")
	return nil
}

// Teardown cleans up the test environment
func (itf *IntegrationTestFramework) Teardown(t *testing.T) {
	t.Log("🧹 Cleaning up integration test environment...")

	if itf.ServerProcess != nil {
		if err := itf.ServerProcess.Process.Kill(); err != nil {
			t.Logf("Warning: Failed to kill server process: %v", err)
		}
		itf.ServerProcess.Wait()
	}

	if itf.CancelFunc != nil {
		itf.CancelFunc()
	}

	t.Log("✅ Integration test environment cleaned up")
}

// StartServer starts the OllamaMax server
func (itf *IntegrationTestFramework) StartServer(t *testing.T) error {
	t.Log("🚀 Starting OllamaMax server...")

	itf.ServerProcess = exec.CommandContext(
		itf.TestContext,
		itf.BinaryPath,
		"start",
		"--port", "8080",
		"--log-level", "debug",
	)

	// Capture output for debugging
	itf.ServerProcess.Stdout = os.Stdout
	itf.ServerProcess.Stderr = os.Stderr

	if err := itf.ServerProcess.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	t.Log("✅ Server started")
	return nil
}

// WaitForServerReady waits for the server to be ready
func (itf *IntegrationTestFramework) WaitForServerReady(t *testing.T, timeout time.Duration) error {
	t.Log("⏳ Waiting for server to be ready...")

	start := time.Now()
	for time.Since(start) < timeout {
		if itf.IsServerReady() {
			t.Log("✅ Server is ready")
			return nil
		}
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("server did not become ready within %v", timeout)
}

// IsServerReady checks if the server is ready
func (itf *IntegrationTestFramework) IsServerReady() bool {
	resp, err := http.Get(itf.APIBaseURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// RunCLICommand runs a CLI command and returns the output
func (itf *IntegrationTestFramework) RunCLICommand(args ...string) (string, error) {
	cmd := exec.Command(itf.BinaryPath, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// RunCLICommandWithTimeout runs a CLI command with timeout
func (itf *IntegrationTestFramework) RunCLICommandWithTimeout(timeout time.Duration, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, itf.BinaryPath, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// TestProxyCommand tests a proxy command with various scenarios
func (itf *IntegrationTestFramework) TestProxyCommand(t *testing.T, command string) {
	testCases := []struct {
		name        string
		args        []string
		expectError bool
		contains    []string
		notContains []string
	}{
		{
			name:        fmt.Sprintf("%s Help", strings.Title(command)),
			args:        []string{"proxy", command, "--help"},
			expectError: false,
			contains:    []string{"--json", "--api-url"},
		},
		{
			name:        fmt.Sprintf("%s Basic", strings.Title(command)),
			args:        []string{"proxy", command, "--api-url", itf.APIBaseURL},
			expectError: false,
			contains:    []string{strings.Title(command)},
		},
		{
			name:        fmt.Sprintf("%s JSON", strings.Title(command)),
			args:        []string{"proxy", command, "--json", "--api-url", itf.APIBaseURL},
			expectError: false,
			contains:    []string{"{", "}"},
		},
		{
			name:        fmt.Sprintf("%s Invalid URL", strings.Title(command)),
			args:        []string{"proxy", command, "--api-url", "http://invalid:9999"},
			expectError: true,
			contains:    []string{"Error"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := itf.RunCLICommandWithTimeout(10*time.Second, tc.args...)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but command succeeded")
				}
			} else {
				if err != nil {
					t.Errorf("Command failed: %v\nOutput: %s", err, output)
				}
			}

			for _, contains := range tc.contains {
				if !strings.Contains(output, contains) {
					t.Errorf("Output should contain '%s'\nOutput: %s", contains, output)
				}
			}

			for _, notContains := range tc.notContains {
				if strings.Contains(output, notContains) {
					t.Errorf("Output should not contain '%s'\nOutput: %s", notContains, output)
				}
			}
		})
	}
}

// ValidateJSONOutput validates that output is valid JSON
func (itf *IntegrationTestFramework) ValidateJSONOutput(t *testing.T, output string) map[string]interface{} {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		t.Fatalf("Output is not valid JSON: %v\nOutput: %s", err, output)
	}
	return data
}

// TestAPIEndpoint tests an API endpoint directly
func (itf *IntegrationTestFramework) TestAPIEndpoint(t *testing.T, endpoint string, expectedStatus int) {
	resp, err := http.Get(itf.APIBaseURL + endpoint)
	if err != nil {
		t.Fatalf("Failed to call API endpoint %s: %v", endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d but got %d for endpoint %s", expectedStatus, resp.StatusCode, endpoint)
	}
}

// PerformanceTest runs performance tests on CLI commands
func (itf *IntegrationTestFramework) PerformanceTest(t *testing.T, command []string, iterations int, maxDuration time.Duration) {
	t.Logf("🏃 Running performance test: %v (%d iterations)", command, iterations)

	start := time.Now()
	for i := 0; i < iterations; i++ {
		_, err := itf.RunCLICommandWithTimeout(10*time.Second, command...)
		if err != nil {
			t.Fatalf("Command failed on iteration %d: %v", i+1, err)
		}
	}
	duration := time.Since(start)

	avgDuration := duration / time.Duration(iterations)
	t.Logf("✅ Performance test complete: %d iterations in %v (avg: %v per command)", 
		iterations, duration, avgDuration)

	if duration > maxDuration {
		t.Errorf("Performance test took too long: %v > %v", duration, maxDuration)
	}
}

// StressTest runs stress tests with concurrent commands
func (itf *IntegrationTestFramework) StressTest(t *testing.T, command []string, concurrency int, duration time.Duration) {
	t.Logf("💪 Running stress test: %v (concurrency: %d, duration: %v)", command, concurrency, duration)

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	results := make(chan error, concurrency)
	
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			for {
				select {
				case <-ctx.Done():
					results <- nil
					return
				default:
					_, err := itf.RunCLICommandWithTimeout(5*time.Second, command...)
					if err != nil {
						results <- fmt.Errorf("worker %d failed: %w", id, err)
						return
					}
				}
			}
		}(i)
	}

	// Wait for all workers to complete
	for i := 0; i < concurrency; i++ {
		if err := <-results; err != nil {
			t.Errorf("Stress test failed: %v", err)
		}
	}

	t.Log("✅ Stress test completed successfully")
}

// GenerateTestReport generates a comprehensive test report
func (itf *IntegrationTestFramework) GenerateTestReport(t *testing.T, results map[string]bool) {
	t.Log("📊 Generating test report...")

	passed := 0
	failed := 0
	
	for testName, success := range results {
		if success {
			t.Logf("✅ %s: PASSED", testName)
			passed++
		} else {
			t.Logf("❌ %s: FAILED", testName)
			failed++
		}
	}

	total := passed + failed
	successRate := float64(passed) / float64(total) * 100

	t.Logf("📈 Test Summary: %d/%d tests passed (%.1f%% success rate)", passed, total, successRate)

	if failed > 0 {
		t.Errorf("Some tests failed: %d failures out of %d tests", failed, total)
	}
}

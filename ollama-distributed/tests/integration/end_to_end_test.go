package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// EndToEndTestSuite tests the complete OllamaMax system
type EndToEndTestSuite struct {
	binaryPath  string
	apiURL      string
	serverCmd   *exec.Cmd
	serverReady bool
}

// TestEndToEndWorkflow tests the complete user workflow
func TestEndToEndWorkflow(t *testing.T) {
	suite := &EndToEndTestSuite{
		binaryPath: "./ollama-distributed",
		apiURL:     "http://localhost:8080",
	}

	// Skip if binary doesn't exist
	if _, err := os.Stat(suite.binaryPath); os.IsNotExist(err) {
		t.Skip("Binary not found, run 'go build ./cmd/node' first")
	}

	t.Run("StartSystem", suite.testStartSystem)
	t.Run("VerifySystemHealth", suite.testVerifySystemHealth)
	t.Run("TestProxyCLI", suite.testProxyCLI)
	t.Run("TestProxyStatus", suite.testProxyStatus)
	t.Run("TestProxyInstances", suite.testProxyInstances)
	t.Run("TestProxyMetrics", suite.testProxyMetrics)
	t.Run("TestJSONOutput", suite.testJSONOutput)
	t.Run("TestErrorHandling", suite.testErrorHandling)
	t.Run("StopSystem", suite.testStopSystem)
}

// testStartSystem starts the distributed system
func (suite *EndToEndTestSuite) testStartSystem(t *testing.T) {
	t.Log("🚀 Starting OllamaMax distributed system...")

	// Start the server in background
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	suite.serverCmd = exec.CommandContext(ctx, suite.binaryPath, "start", "--port", "8080")
	suite.serverCmd.Stdout = os.Stdout
	suite.serverCmd.Stderr = os.Stderr

	err := suite.serverCmd.Start()
	require.NoError(t, err, "Failed to start server")

	// Wait for server to be ready
	suite.waitForServerReady(t, 30*time.Second)
	suite.serverReady = true

	t.Log("✅ System started successfully")
}

// testVerifySystemHealth verifies the system is healthy
func (suite *EndToEndTestSuite) testVerifySystemHealth(t *testing.T) {
	if !suite.serverReady {
		t.Skip("Server not ready")
	}

	t.Log("🔍 Verifying system health...")

	// Test API endpoint
	resp, err := http.Get(suite.apiURL + "/health")
	require.NoError(t, err, "Health check failed")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Health check returned non-200 status")

	t.Log("✅ System health verified")
}

// testProxyCLI tests basic proxy CLI functionality
func (suite *EndToEndTestSuite) testProxyCLI(t *testing.T) {
	t.Log("🎛️ Testing proxy CLI commands...")

	// Test proxy help
	cmd := exec.Command(suite.binaryPath, "proxy", "--help")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Proxy help command failed")

	outputStr := string(output)
	assert.Contains(t, outputStr, "status", "Help should mention status command")
	assert.Contains(t, outputStr, "instances", "Help should mention instances command")
	assert.Contains(t, outputStr, "metrics", "Help should mention metrics command")

	t.Log("✅ Proxy CLI help working")
}

// testProxyStatus tests proxy status command
func (suite *EndToEndTestSuite) testProxyStatus(t *testing.T) {
	if !suite.serverReady {
		t.Skip("Server not ready")
	}

	t.Log("📊 Testing proxy status command...")

	// Test basic status
	cmd := exec.Command(suite.binaryPath, "proxy", "status", "--api-url", suite.apiURL)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Proxy status command failed: %s", string(output))

	outputStr := string(output)
	assert.Contains(t, outputStr, "Proxy Status", "Status output should contain header")

	t.Log("✅ Proxy status command working")
}

// testProxyInstances tests proxy instances command
func (suite *EndToEndTestSuite) testProxyInstances(t *testing.T) {
	if !suite.serverReady {
		t.Skip("Server not ready")
	}

	t.Log("🔗 Testing proxy instances command...")

	// Test instances command
	cmd := exec.Command(suite.binaryPath, "proxy", "instances", "--api-url", suite.apiURL)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Proxy instances command failed: %s", string(output))

	outputStr := string(output)
	assert.Contains(t, outputStr, "Instances", "Instances output should contain header")

	t.Log("✅ Proxy instances command working")
}

// testProxyMetrics tests proxy metrics command
func (suite *EndToEndTestSuite) testProxyMetrics(t *testing.T) {
	if !suite.serverReady {
		t.Skip("Server not ready")
	}

	t.Log("📈 Testing proxy metrics command...")

	// Test metrics command
	cmd := exec.Command(suite.binaryPath, "proxy", "metrics", "--api-url", suite.apiURL)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Proxy metrics command failed: %s", string(output))

	outputStr := string(output)
	assert.Contains(t, outputStr, "Metrics", "Metrics output should contain header")

	t.Log("✅ Proxy metrics command working")
}

// testJSONOutput tests JSON output functionality
func (suite *EndToEndTestSuite) testJSONOutput(t *testing.T) {
	if !suite.serverReady {
		t.Skip("Server not ready")
	}

	t.Log("📋 Testing JSON output...")

	// Test JSON status
	cmd := exec.Command(suite.binaryPath, "proxy", "status", "--json", "--api-url", suite.apiURL)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "JSON status command failed: %s", string(output))

	// Verify it's valid JSON
	var statusData map[string]interface{}
	err = json.Unmarshal(output, &statusData)
	require.NoError(t, err, "Status output is not valid JSON")

	t.Log("✅ JSON output working")
}

// testErrorHandling tests error handling scenarios
func (suite *EndToEndTestSuite) testErrorHandling(t *testing.T) {
	t.Log("🚨 Testing error handling...")

	// Test with invalid API URL
	cmd := exec.Command(suite.binaryPath, "proxy", "status", "--api-url", "http://invalid:9999")
	output, err := cmd.CombinedOutput()

	// Command should fail gracefully
	assert.Error(t, err, "Command should fail with invalid URL")

	outputStr := string(output)
	assert.Contains(t, outputStr, "Error", "Error output should contain error message")

	t.Log("✅ Error handling working")
}

// testStopSystem stops the distributed system
func (suite *EndToEndTestSuite) testStopSystem(t *testing.T) {
	if suite.serverCmd != nil && suite.serverCmd.Process != nil {
		t.Log("🛑 Stopping system...")

		err := suite.serverCmd.Process.Kill()
		if err != nil {
			t.Logf("Warning: Failed to kill server process: %v", err)
		}

		suite.serverCmd.Wait()
		suite.serverReady = false

		t.Log("✅ System stopped")
	}
}

// waitForServerReady waits for the server to be ready
func (suite *EndToEndTestSuite) waitForServerReady(t *testing.T, timeout time.Duration) {
	t.Log("⏳ Waiting for server to be ready...")

	start := time.Now()
	for time.Since(start) < timeout {
		resp, err := http.Get(suite.apiURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			t.Log("✅ Server is ready")
			return
		}
		if resp != nil {
			resp.Body.Close()
		}

		time.Sleep(1 * time.Second)
	}

	t.Fatalf("Server did not become ready within %v", timeout)
}

// TestProxyWorkflow tests a complete proxy workflow
func TestProxyWorkflow(t *testing.T) {
	// This test simulates a real user workflow
	t.Log("🎯 Testing complete proxy workflow...")

	binaryPath := "./ollama-distributed"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("Binary not found, run 'go build ./cmd/node' first")
	}

	// Test workflow steps
	testCases := []struct {
		name        string
		args        []string
		expectError bool
		contains    []string
	}{
		{
			name:        "Help Command",
			args:        []string{"--help"},
			expectError: false,
			contains:    []string{"OllamaMax", "proxy"},
		},
		{
			name:        "Proxy Help",
			args:        []string{"proxy", "--help"},
			expectError: false,
			contains:    []string{"status", "instances", "metrics"},
		},
		{
			name:        "Status Help",
			args:        []string{"proxy", "status", "--help"},
			expectError: false,
			contains:    []string{"--json", "--api-url"},
		},
		{
			name:        "Invalid Command",
			args:        []string{"invalid-command"},
			expectError: true,
			contains:    []string{"Error", "unknown"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tc.args...)
			output, err := cmd.CombinedOutput()

			if tc.expectError {
				assert.Error(t, err, "Command should have failed")
			} else {
				assert.NoError(t, err, "Command should have succeeded: %s", string(output))
			}

			outputStr := string(output)
			for _, contains := range tc.contains {
				assert.Contains(t, outputStr, contains, "Output should contain: %s", contains)
			}
		})
	}

	t.Log("✅ Proxy workflow test complete")
}

// BenchmarkProxyCommands benchmarks proxy command performance
func BenchmarkProxyCommands(b *testing.B) {
	binaryPath := "./ollama-distributed"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		b.Skip("Binary not found")
	}

	b.Run("ProxyHelp", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cmd := exec.Command(binaryPath, "proxy", "--help")
			_, err := cmd.CombinedOutput()
			if err != nil {
				b.Fatalf("Command failed: %v", err)
			}
		}
	})

	b.Run("StatusHelp", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cmd := exec.Command(binaryPath, "proxy", "status", "--help")
			_, err := cmd.CombinedOutput()
			if err != nil {
				b.Fatalf("Command failed: %v", err)
			}
		}
	})
}

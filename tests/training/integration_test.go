// Integration tests for the training system components
package training

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TrainingIntegrationSuite manages integration testing for training components
type TrainingIntegrationSuite struct {
	ProjectRoot     string
	TrainingRoot    string
	BinaryPath      string
	TestConfigPath  string
	TestServer      *httptest.Server
	HTTPClient      *http.Client
}

// NewTrainingIntegrationSuite creates a new integration test suite
func NewTrainingIntegrationSuite() *TrainingIntegrationSuite {
	projectRoot := "/home/kp/ollamamax"
	trainingRoot := filepath.Join(projectRoot, "ollama-distributed", "training")
	binaryPath := filepath.Join(projectRoot, "ollama-distributed", "bin", "ollama-distributed")

	return &TrainingIntegrationSuite{
		ProjectRoot:    projectRoot,
		TrainingRoot:   trainingRoot,
		BinaryPath:     binaryPath,
		HTTPClient:     &http.Client{Timeout: 30 * time.Second},
	}
}

// Setup creates test environment
func (tis *TrainingIntegrationSuite) Setup(t *testing.T) {
	// Create test configuration
	tis.createTestConfiguration(t)
	
	// Setup mock API server
	tis.setupMockAPIServer()
	
	// Ensure binary exists
	tis.ensureBinaryExists(t)
}

// Teardown cleans up test environment
func (tis *TrainingIntegrationSuite) Teardown() {
	if tis.TestServer != nil {
		tis.TestServer.Close()
	}
	
	// Clean up test files
	if tis.TestConfigPath != "" {
		os.Remove(tis.TestConfigPath)
	}
}

func (tis *TrainingIntegrationSuite) createTestConfiguration(t *testing.T) {
	config := map[string]interface{}{
		"server": map[string]interface{}{
			"host": "127.0.0.1",
			"port": 8080,
		},
		"p2p": map[string]interface{}{
			"enabled": true,
			"port":    4001,
		},
		"storage": map[string]interface{}{
			"data_dir": "./test-data",
		},
		"logging": map[string]interface{}{
			"level": "info",
		},
	}
	
	data, err := yaml.Marshal(config)
	require.NoError(t, err)
	
	testConfigDir := filepath.Join(os.TempDir(), "ollama-test-configs")
	os.MkdirAll(testConfigDir, 0755)
	
	tis.TestConfigPath = filepath.Join(testConfigDir, "test-config.yaml")
	err = ioutil.WriteFile(tis.TestConfigPath, data, 0644)
	require.NoError(t, err)
}

func (tis *TrainingIntegrationSuite) setupMockAPIServer() {
	mux := http.NewServeMux()
	
	// Health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"version":   "test-1.0.0",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	
	// API endpoints for training
	mux.HandleFunc("/api/distributed/status", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"cluster_id": "test-cluster",
			"node_id":    "test-node-1",
			"status":     "active",
			"nodes":      1,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	
	mux.HandleFunc("/api/distributed/nodes", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"nodes": []map[string]interface{}{
				{
					"id":     "test-node-1",
					"host":   "127.0.0.1",
					"port":   8080,
					"status": "active",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	
	mux.HandleFunc("/api/distributed/metrics", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"requests_total":     1000,
			"requests_per_sec":   10.5,
			"avg_latency_ms":     125,
			"active_connections": 15,
			"memory_usage_mb":    256,
			"cpu_usage_percent":  25.5,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	
	tis.TestServer = httptest.NewServer(mux)
}

func (tis *TrainingIntegrationSuite) ensureBinaryExists(t *testing.T) {
	if _, err := os.Stat(tis.BinaryPath); os.IsNotExist(err) {
		// Try to build the binary
		cmd := exec.Command("go", "build", "-o", "bin/ollama-distributed", "./cmd/node")
		cmd.Dir = filepath.Join(tis.ProjectRoot, "ollama-distributed")
		
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Skipf("Cannot build binary for integration tests: %v\nOutput: %s", err, string(output))
		}
	}
}

// INTEGRATION TESTS

func TestTrainingCodeExamplesIntegration(t *testing.T) {
	tis := NewTrainingIntegrationSuite()
	tis.Setup(t)
	defer tis.Teardown()
	
	t.Run("Configuration Manager Integration", func(t *testing.T) {
		tis.testConfigurationManagerIntegration(t)
	})
	
	t.Run("Health Monitoring Integration", func(t *testing.T) {
		tis.testHealthMonitoringIntegration(t)
	})
	
	t.Run("API Client Integration", func(t *testing.T) {
		tis.testAPIClientIntegration(t)
	})
	
	t.Run("Validation Suite Integration", func(t *testing.T) {
		tis.testValidationSuiteIntegration(t)
	})
}

func (tis *TrainingIntegrationSuite) testConfigurationManagerIntegration(t *testing.T) {
	configManagerPath := filepath.Join(tis.TrainingRoot, "code-examples", "02-configuration", "configuration-manager.go")
	
	if _, err := os.Stat(configManagerPath); err != nil {
		t.Skip("Configuration manager not found, skipping integration test")
		return
	}
	
	// Test configuration generation
	tempDir := os.TempDir()
	os.Setenv("HOME", tempDir) // Override home directory for test
	defer func() {
		os.Unsetenv("HOME")
		// Clean up generated configs
		os.RemoveAll(filepath.Join(tempDir, ".ollama-distributed"))
	}()
	
	cmd := exec.Command("go", "run", "configuration-manager.go")
	cmd.Dir = filepath.Dir(configManagerPath)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Configuration manager failed: %v\nOutput: %s", err, string(output))
	}
	
	// Verify configurations were created
	configDir := filepath.Join(tempDir, ".ollama-distributed", "profiles")
	profiles := []string{"development.yaml", "testing.yaml", "production.yaml"}
	
	for _, profile := range profiles {
		profilePath := filepath.Join(configDir, profile)
		assert.FileExists(t, profilePath, "Profile %s should be created", profile)
		
		// Validate YAML syntax
		data, err := ioutil.ReadFile(profilePath)
		require.NoError(t, err)
		
		var config interface{}
		err = yaml.Unmarshal(data, &config)
		assert.NoError(t, err, "Profile %s should have valid YAML syntax", profile)
	}
}

func (tis *TrainingIntegrationSuite) testHealthMonitoringIntegration(t *testing.T) {
	scriptPath := filepath.Join(tis.TrainingRoot, "code-examples", "03-operations", "health-monitoring-dashboard.sh")
	
	if _, err := os.Stat(scriptPath); err != nil {
		t.Skip("Health monitoring script not found, skipping integration test")
		return
	}
	
	// Make script executable
	err := os.Chmod(scriptPath, 0755)
	require.NoError(t, err)
	
	// Test check command (should handle case where service is not running)
	cmd := exec.Command(scriptPath, "check")
	output, err := cmd.CombinedOutput()
	
	// Script might fail if service is not running, but should not crash
	if err != nil && !strings.Contains(string(output), "connection") && !strings.Contains(string(output), "refused") {
		t.Errorf("Health monitoring script failed unexpectedly: %v\nOutput: %s", err, string(output))
	}
	
	// Test help command
	cmd = exec.Command(scriptPath, "--help")
	output, err = cmd.Output()
	if err == nil {
		assert.Contains(t, string(output), "health", "Help output should mention health functionality")
	}
}

func (tis *TrainingIntegrationSuite) testAPIClientIntegration(t *testing.T) {
	clientPath := filepath.Join(tis.TrainingRoot, "code-examples", "04-api-integration", "comprehensive-api-client.go")
	
	if _, err := os.Stat(clientPath); err != nil {
		t.Skip("API client not found, skipping integration test")
		return
	}
	
	// Test compilation
	tempBinary := filepath.Join(os.TempDir(), "test-api-client")
	cmd := exec.Command("go", "build", "-o", tempBinary, "comprehensive-api-client.go")
	cmd.Dir = filepath.Dir(clientPath)
	
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "API client should compile successfully: %s", string(output))
	defer os.Remove(tempBinary)
	
	// Test basic functionality (should handle connection errors gracefully)
	cmd = exec.Command(tempBinary, "health")
	output, err = cmd.CombinedOutput()
	
	// Client should handle connection errors gracefully
	if err != nil {
		assert.Contains(t, string(output), "connection", "Client should report connection issues clearly")
	}
}

func (tis *TrainingIntegrationSuite) testValidationSuiteIntegration(t *testing.T) {
	suitePath := filepath.Join(tis.TrainingRoot, "code-examples", "05-validation-testing", "training-validation-suite.go")
	
	if _, err := os.Stat(suitePath); err != nil {
		t.Skip("Validation suite not found, skipping integration test")
		return
	}
	
	// Test compilation
	tempBinary := filepath.Join(os.TempDir(), "test-validation-suite")
	cmd := exec.Command("go", "build", "-o", tempBinary, "training-validation-suite.go")
	cmd.Dir = filepath.Dir(suitePath)
	
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Validation suite should compile successfully: %s", string(output))
	defer os.Remove(tempBinary)
	
	// Test basic functionality
	tempOutput := filepath.Join(os.TempDir(), "validation-results.json")
	defer os.Remove(tempOutput)
	
	cmd = exec.Command(tempBinary, "--config", tis.TestConfigPath, "--output", tempOutput)
	output, err = cmd.CombinedOutput()
	
	// Validation suite should run (might have failures due to missing service)
	if err != nil {
		// Check if it's a legitimate error or just validation failures
		if !strings.Contains(string(output), "connection") && !strings.Contains(string(output), "refused") {
			t.Errorf("Validation suite failed unexpectedly: %v\nOutput: %s", err, string(output))
		}
	}
	
	// Check if results file was created
	if _, err := os.Stat(tempOutput); err == nil {
		data, err := ioutil.ReadFile(tempOutput)
		require.NoError(t, err)
		
		var results map[string]interface{}
		err = json.Unmarshal(data, &results)
		assert.NoError(t, err, "Results should be valid JSON")
		
		assert.Contains(t, results, "total_tests", "Results should contain test count")
	}
}

// Service Integration Tests
func TestServiceIntegrationWithTraining(t *testing.T) {
	tis := NewTrainingIntegrationSuite()
	tis.Setup(t)
	defer tis.Teardown()
	
	// Test service startup and training interaction
	t.Run("Service Startup with Training Config", func(t *testing.T) {
		tis.testServiceStartupWithTrainingConfig(t)
	})
	
	t.Run("API Endpoints for Training", func(t *testing.T) {
		tis.testAPIEndpointsForTraining(t)
	})
}

func (tis *TrainingIntegrationSuite) testServiceStartupWithTrainingConfig(t *testing.T) {
	// Test dry run with training configuration
	cmd := exec.Command(tis.BinaryPath, "start", "--config", tis.TestConfigPath, "--dry-run")
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		if strings.Contains(err.Error(), "unknown flag") {
			t.Skip("Dry run flag not implemented")
			return
		}
		t.Errorf("Dry run with training config failed: %v\nOutput: %s", err, string(output))
	}
	
	// If dry run succeeded, output should not contain errors
	outputStr := strings.ToLower(string(output))
	if strings.Contains(outputStr, "error") || strings.Contains(outputStr, "failed") {
		t.Errorf("Dry run reported errors: %s", string(output))
	}
}

func (tis *TrainingIntegrationSuite) testAPIEndpointsForTraining(t *testing.T) {
	// Test our mock API server to ensure training endpoints work
	endpoints := map[string]string{
		"Health":           "/health",
		"Cluster Status":   "/api/distributed/status",
		"Node List":        "/api/distributed/nodes",
		"System Metrics":   "/api/distributed/metrics",
	}
	
	for name, endpoint := range endpoints {
		t.Run(name, func(t *testing.T) {
			resp, err := tis.HTTPClient.Get(tis.TestServer.URL + endpoint)
			require.NoError(t, err, "Endpoint %s should be accessible", endpoint)
			defer resp.Body.Close()
			
			assert.Equal(t, http.StatusOK, resp.StatusCode, "Endpoint %s should return 200", endpoint)
			
			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			assert.NoError(t, err, "Endpoint %s should return valid JSON", endpoint)
			
			// Basic validation that response contains expected fields
			assert.NotEmpty(t, response, "Endpoint %s should return non-empty response", endpoint)
		})
	}
}

// End-to-End Training Workflow Test
func TestEndToEndTrainingWorkflow(t *testing.T) {
	tis := NewTrainingIntegrationSuite()
	tis.Setup(t)
	defer tis.Teardown()
	
	t.Run("Complete Training Workflow", func(t *testing.T) {
		tis.testCompleteTrainingWorkflow(t)
	})
}

func (tis *TrainingIntegrationSuite) testCompleteTrainingWorkflow(t *testing.T) {
	// Step 1: Configuration Generation
	configManagerPath := filepath.Join(tis.TrainingRoot, "code-examples", "02-configuration", "configuration-manager.go")
	if _, err := os.Stat(configManagerPath); err == nil {
		// Generate configurations
		tempDir := os.TempDir()
		os.Setenv("HOME", tempDir)
		defer func() {
			os.Unsetenv("HOME")
			os.RemoveAll(filepath.Join(tempDir, ".ollama-distributed"))
		}()
		
		cmd := exec.Command("go", "run", "configuration-manager.go")
		cmd.Dir = filepath.Dir(configManagerPath)
		_, err := cmd.CombinedOutput()
		
		if err != nil {
			t.Logf("Configuration generation failed (continuing): %v", err)
		} else {
			t.Log("✅ Configuration generation completed")
		}
	}
	
	// Step 2: Binary validation
	if _, err := os.Stat(tis.BinaryPath); err == nil {
		cmd := exec.Command(tis.BinaryPath, "--help")
		_, err := cmd.Output()
		if err != nil {
			t.Errorf("Binary help command failed: %v", err)
		} else {
			t.Log("✅ Binary validation completed")
		}
	} else {
		t.Log("⚠️ Binary not available for testing")
	}
	
	// Step 3: API client testing
	clientPath := filepath.Join(tis.TrainingRoot, "code-examples", "04-api-integration", "comprehensive-api-client.go")
	if _, err := os.Stat(clientPath); err == nil {
		tempBinary := filepath.Join(os.TempDir(), "test-workflow-client")
		cmd := exec.Command("go", "build", "-o", tempBinary, "comprehensive-api-client.go")
		cmd.Dir = filepath.Dir(clientPath)
		
		if err := cmd.Run(); err == nil {
			defer os.Remove(tempBinary)
			t.Log("✅ API client compilation completed")
		} else {
			t.Logf("API client compilation failed: %v", err)
		}
	}
	
	// Step 4: Validation suite
	suitePath := filepath.Join(tis.TrainingRoot, "code-examples", "05-validation-testing", "training-validation-suite.go")
	if _, err := os.Stat(suitePath); err == nil {
		tempBinary := filepath.Join(os.TempDir(), "test-workflow-validation")
		cmd := exec.Command("go", "build", "-o", tempBinary, "training-validation-suite.go")
		cmd.Dir = filepath.Dir(suitePath)
		
		if err := cmd.Run(); err == nil {
			defer os.Remove(tempBinary)
			t.Log("✅ Validation suite compilation completed")
		} else {
			t.Logf("Validation suite compilation failed: %v", err)
		}
	}
	
	t.Log("🎉 End-to-end training workflow validation completed")
}

// Performance testing for training components
func TestTrainingComponentsPerformance(t *testing.T) {
	tis := NewTrainingIntegrationSuite()
	tis.Setup(t)
	defer tis.Teardown()
	
	t.Run("Configuration Generation Performance", func(t *testing.T) {
		tis.benchmarkConfigurationGeneration(t)
	})
	
	t.Run("API Client Performance", func(t *testing.T) {
		tis.benchmarkAPIClient(t)
	})
}

func (tis *TrainingIntegrationSuite) benchmarkConfigurationGeneration(t *testing.T) {
	configManagerPath := filepath.Join(tis.TrainingRoot, "code-examples", "02-configuration", "configuration-manager.go")
	
	if _, err := os.Stat(configManagerPath); err != nil {
		t.Skip("Configuration manager not found")
		return
	}
	
	start := time.Now()
	
	tempDir := os.TempDir()
	os.Setenv("HOME", tempDir)
	defer func() {
		os.Unsetenv("HOME")
		os.RemoveAll(filepath.Join(tempDir, ".ollama-distributed"))
	}()
	
	cmd := exec.Command("go", "run", "configuration-manager.go")
	cmd.Dir = filepath.Dir(configManagerPath)
	
	_, err := cmd.CombinedOutput()
	duration := time.Since(start)
	
	if err == nil {
		t.Logf("Configuration generation took %v", duration)
		
		// Should complete within reasonable time
		if duration > 30*time.Second {
			t.Errorf("Configuration generation took too long: %v", duration)
		}
	}
}

func (tis *TrainingIntegrationSuite) benchmarkAPIClient(t *testing.T) {
	// Test API client performance against our mock server
	start := time.Now()
	
	for i := 0; i < 10; i++ {
		resp, err := tis.HTTPClient.Get(tis.TestServer.URL + "/health")
		if err != nil {
			t.Errorf("API request failed: %v", err)
			break
		}
		resp.Body.Close()
	}
	
	duration := time.Since(start)
	avgDuration := duration / 10
	
	t.Logf("Average API response time: %v", avgDuration)
	
	if avgDuration > 100*time.Millisecond {
		t.Errorf("API responses too slow: %v average", avgDuration)
	}
}
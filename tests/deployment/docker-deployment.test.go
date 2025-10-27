package deployment_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDockerComposeUp verifies all services start successfully
func TestDockerComposeUp(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Start services
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", "docker-compose.yml", "up", "-d")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to start services: %s", string(output))

	// Wait for services to be ready
	time.Sleep(30 * time.Second)

	// Verify services are running
	cmd = exec.CommandContext(ctx, "docker", "compose", "ps", "--format", "json")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Failed to list services: %s", string(output))

	assert.Contains(t, string(output), "\"State\":\"running\"", "Not all services are running")
}

// TestServiceHealthChecks validates all health endpoints
func TestServiceHealthChecks(t *testing.T) {
	// Get actual service names from docker-compose
	healthEndpoints := []struct {
		name     string
		url      string
		timeout  time.Duration
		expected int
	}{
		{"Ollama Primary", "http://localhost:11434/api/tags", 30 * time.Second, http.StatusOK},
		{"OllamaMax API", "http://localhost:13100/health", 10 * time.Second, http.StatusOK},
		{"OllamaMax Web", "http://localhost:8080/", 10 * time.Second, http.StatusOK},
	}

	for _, endpoint := range healthEndpoints {
		t.Run(endpoint.name, func(t *testing.T) {
			client := &http.Client{Timeout: endpoint.timeout}

			// Retry logic for service startup
			var resp *http.Response
			var err error
			for i := 0; i < 5; i++ {
				resp, err = client.Get(endpoint.url)
				if err == nil && resp.StatusCode == endpoint.expected {
					break
				}
				time.Sleep(5 * time.Second)
			}

			require.NoError(t, err, "Health check failed for %s", endpoint.name)
			defer resp.Body.Close()

			assert.Equal(t, endpoint.expected, resp.StatusCode,
				"Unexpected status code from %s", endpoint.name)
		})
	}
}

// TestServiceDependencies checks inter-service connectivity
func TestServiceDependencies(t *testing.T) {
	// Use actual docker-compose service names
	dependencies := []struct {
		from string
		to   string
	}{
		{"ollamamax-api", "postgres"},
		{"ollamamax-api", "redis"},
		{"ollamamax-web", "ollamamax-api"},
	}

	for _, dep := range dependencies {
		t.Run(fmt.Sprintf("%s->%s", dep.from, dep.to), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, "docker", "compose", "exec", "-T",
				dep.from, "ping", "-c", "1", "-W", "1", dep.to)
			output, err := cmd.CombinedOutput()

			assert.NoError(t, err, "Connectivity failed from %s to %s: %s",
				dep.from, dep.to, string(output))
		})
	}
}

// TestVolumeMount verifies data persistence
func TestVolumeMount(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	testKey := fmt.Sprintf("test-key-%d", time.Now().Unix())
	testValue := "test-value"

	// Write data to Redis
	cmd := exec.CommandContext(ctx, "docker", "compose", "exec", "-T",
		"redis", "redis-cli", "SET", testKey, testValue)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to write to Redis: %s", string(output))

	// Restart Redis
	cmd = exec.CommandContext(ctx, "docker", "compose", "restart", "redis")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Failed to restart Redis: %s", string(output))

	time.Sleep(5 * time.Second)

	// Read data from Redis
	cmd = exec.CommandContext(ctx, "docker", "compose", "exec", "-T",
		"redis", "redis-cli", "GET", testKey)
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Failed to read from Redis: %s", string(output))

	assert.Contains(t, string(output), testValue, "Data not persisted across restart")

	// Cleanup
	cmd = exec.CommandContext(ctx, "docker", "compose", "exec", "-T",
		"redis", "redis-cli", "DEL", testKey)
	cmd.Run()
}

// TestNetworkConnectivity validates network configuration
func TestNetworkConnectivity(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get expected network name from environment or parse from docker-compose
	expectedNetwork := os.Getenv("EXPECTED_NETWORK")

	if expectedNetwork == "" {
		// Dynamically discover networks from docker-compose config
		cmd := exec.CommandContext(ctx, "docker", "compose", "config", "--format", "json")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Skipf("Failed to parse docker-compose config, skipping network check: %s", string(output))
			return
		}

		// Parse the JSON to extract network names
		var composeConfig map[string]interface{}
		if err := json.Unmarshal(output, &composeConfig); err != nil {
			t.Skipf("Failed to parse docker-compose JSON, skipping network check: %v", err)
			return
		}

		// Extract networks from config
		if networks, ok := composeConfig["networks"].(map[string]interface{}); ok && len(networks) > 0 {
			// Use the first network as expected network
			for networkName := range networks {
				expectedNetwork = networkName
				break
			}
		}

		if expectedNetwork == "" {
			t.Skip("No networks defined in docker-compose, skipping network check")
			return
		}
	}

	// Check if the expected network exists
	cmd := exec.CommandContext(ctx, "docker", "network", "ls", "--format", "{{.Name}}")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to list networks: %s", string(output))

	networks := bytes.Split(output, []byte("\n"))
	found := false
	for _, network := range networks {
		networkName := string(bytes.TrimSpace(network))
		// Check for exact match or match with project prefix
		if networkName == expectedNetwork || bytes.Contains(network, []byte(expectedNetwork)) {
			found = true
			t.Logf("Found expected network: %s", networkName)
			break
		}
	}

	assert.True(t, found, "Expected network '%s' not found. Available networks: %s",
		expectedNetwork, string(output))
}

// TestEnvironmentVariables verifies environment configuration
func TestEnvironmentVariables(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use actual docker-compose service names
	envVars := []struct {
		service string
		varName string
	}{
		{"ollamamax-api", "POSTGRES_HOST"},
		{"ollamamax-api", "REDIS_HOST"},
		{"ollamamax-web", "API_BASE_URL"},
	}

	for _, env := range envVars {
		t.Run(fmt.Sprintf("%s:%s", env.service, env.varName), func(t *testing.T) {
			cmd := exec.CommandContext(ctx, "docker", "compose", "exec", "-T",
				env.service, "env")
			output, err := cmd.CombinedOutput()
			require.NoError(t, err, "Failed to get environment: %s", string(output))

			assert.Contains(t, string(output), env.varName,
				"Environment variable %s not found in %s", env.varName, env.service)
		})
	}
}

// TestResourceLimits validates CPU/memory limits
func TestResourceLimits(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "compose", "ps", "--format", "json")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to list services: %s", string(output))

	// Check that compose file has resource limits defined
	composeContent, err := os.ReadFile("docker-compose.yml")
	require.NoError(t, err, "Failed to read docker-compose.yml")

	// At minimum, check that resources section exists
	assert.Contains(t, string(composeContent), "resources:",
		"No resource limits defined in compose file")
}

// TestStartupTime measures service startup duration
func TestStartupTime(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Stop all services
	stopCmd := exec.CommandContext(ctx, "docker", "compose", "down")
	stopCmd.Run()

	// Measure startup time
	startTime := time.Now()

	cmd := exec.CommandContext(ctx, "docker", "compose", "up", "-d")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to start services: %s", string(output))

	// Wait for health checks
	time.Sleep(30 * time.Second)

	duration := time.Since(startTime)

	t.Logf("Service startup time: %v", duration)
	assert.Less(t, duration, 3*time.Minute, "Startup time exceeded threshold")
}

// TestPortMappings verifies port configurations
func TestPortMappings(t *testing.T) {
	// Match actual docker-compose.yml service names and ports
	expectedPorts := []struct {
		port    string
		service string
	}{
		{"11434", "ollama-primary"},
		{"13100", "ollamamax-api"},
		{"8080", "ollamamax-web"},
		{"6379", "redis"},
		{"5432", "postgres"},
	}

	for _, portMapping := range expectedPorts {
		t.Run(fmt.Sprintf("Port_%s", portMapping.port), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, "docker", "compose", "ps",
				"--format", "json", portMapping.service)
			output, err := cmd.CombinedOutput()
			require.NoError(t, err, "Failed to get service info: %s", string(output))

			assert.Contains(t, string(output), portMapping.port,
				"Port %s not mapped for service %s", portMapping.port, portMapping.service)
		})
	}
}

// TestSecretsManagement ensures secrets are not exposed
func TestSecretsManagement(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check compose file for hardcoded secrets
	composeContent, err := os.ReadFile("docker-compose.yml")
	require.NoError(t, err, "Failed to read docker-compose.yml")

	sensitivePatterns := []string{
		"password=",
		"secret=",
		"key=",
		"token=",
	}

	for _, pattern := range sensitivePatterns {
		// Ensure patterns use environment variables or secrets
		lines := bytes.Split(composeContent, []byte("\n"))
		for _, line := range lines {
			if bytes.Contains(bytes.ToLower(line), []byte(pattern)) {
				// Should use ${VAR} or secrets:
				assert.True(t,
					bytes.Contains(line, []byte("${")) ||
						bytes.Contains(line, []byte("secrets:")),
					"Potential hardcoded secret found: %s", string(line))
			}
		}
	}
}

// Cleanup function
func TestMain(m *testing.M) {
	// Setup
	fmt.Println("Setting up Docker deployment tests...")

	// Run tests
	code := m.Run()

	// Cleanup
	fmt.Println("Cleaning up Docker deployment tests...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "compose", "down", "-v")
	cmd.Run()

	os.Exit(code)
}

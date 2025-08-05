package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Mock HTTP server responses for different proxy endpoints
var mockProxyResponses = map[string]map[int]string{
	"/api/v1/proxy/status": {
		200: `{
			"status": "running",
			"instance_count": 3,
			"healthy_instances": 2,
			"load_balancer": true
		}`,
		503: `{
			"error": "Proxy not available",
			"message": "Ollama proxy is not initialized",
			"status": "disabled"
		}`,
	},
	"/api/v1/proxy/instances": {
		200: `{
			"instances": [
				{
					"id": "instance-1",
					"node_id": "node-1",
					"endpoint": "http://localhost:11434",
					"status": "healthy",
					"last_seen": "2024-01-01T12:00:00Z",
					"request_count": 150,
					"error_count": 2
				},
				{
					"id": "instance-2", 
					"node_id": "node-2",
					"endpoint": "http://localhost:11435",
					"status": "unhealthy",
					"last_seen": "2024-01-01T11:55:00Z",
					"request_count": 89,
					"error_count": 15
				}
			]
		}`,
		503: `{
			"error": "Proxy not available",
			"message": "Ollama proxy is not initialized"
		}`,
	},
	"/api/v1/proxy/metrics": {
		200: `{
			"total_requests": 1250,
			"successful_requests": 1180,
			"failed_requests": 70,
			"average_latency": 125000000,
			"requests_per_second": 12.5,
			"load_balancing": {
				"decisions": 1250,
				"errors": 15
			},
			"instance_metrics": {
				"instance-1": {
					"requests": 750,
					"errors": 10,
					"average_latency": 120000000,
					"last_request": "2024-01-01T12:00:00Z"
				},
				"instance-2": {
					"requests": 500,
					"errors": 60,
					"average_latency": 180000000,
					"last_request": "2024-01-01T11:58:00Z"
				}
			}
		}`,
		503: `{
			"error": "Proxy not available",
			"message": "Ollama proxy is not initialized"
		}`,
	},
}

// createMockServer creates a test HTTP server with predefined responses
func createMockServer(statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response, exists := mockProxyResponses[r.URL.Path][statusCode]
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "Not found"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		w.Write([]byte(response))
	}))
}

// captureOutput captures stdout during command execution
func captureOutput(fn func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = oldStdout
	output, _ := io.ReadAll(r)
	return string(output)
}

func TestProxyStatusCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		serverStatus   int
		expectedError  bool
		expectedOutput []string
		description    string
	}{
		{
			name:          "successful status",
			args:          []string{"status"},
			serverStatus:  200,
			expectedError: false,
			expectedOutput: []string{
				"🔄 Ollama Proxy Status",
				"=====================",
				"running",
				"instance_count",
				"healthy_instances",
			},
			description: "Test successful proxy status retrieval",
		},
		{
			name:          "status with JSON output",
			args:          []string{"status", "--json"},
			serverStatus:  200,
			expectedError: false,
			expectedOutput: []string{
				"status",
				"instance_count",
				"healthy_instances",
				"load_balancer",
			},
			description: "Test proxy status with JSON output",
		},
		{
			name:          "proxy unavailable",
			args:          []string{"status"},
			serverStatus:  503,
			expectedError: true,
			expectedOutput: []string{
				"🔄 Ollama Proxy Status",
				"HTTP 503",
			},
			description: "Test proxy status when service unavailable",
		},
		{
			name:          "custom API URL",
			args:          []string{"status", "--api-url", "MOCK_SERVER_URL"},
			serverStatus:  200,
			expectedError: false,
			expectedOutput: []string{
				"🔄 Ollama Proxy Status",
				"API URL: MOCK_SERVER_URL",
			},
			description: "Test proxy status with custom API URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			mockServer := createMockServer(tt.serverStatus)
			defer mockServer.Close()

			// Replace MOCK_SERVER_URL placeholder with actual server URL
			for i, arg := range tt.args {
				if arg == "MOCK_SERVER_URL" {
					tt.args[i] = mockServer.URL
				}
			}

			// Create command
			cmd := proxyStatusCmd()
			cmd.SetArgs(tt.args)

			// Set API URL if not provided in args
			if !contains(tt.args, "--api-url") {
				cmd.Flags().Set("api-url", mockServer.URL)
			}

			// Capture output and execute
			var output string
			var err error

			if contains(tt.expectedOutput, "API URL:") {
				// For tests that check API URL in output, capture stdout
				output = captureOutput(func() {
					err = cmd.Execute()
				})
			} else {
				// For other tests, execute normally
				err = cmd.Execute()
			}

			// Check error expectation
			if tt.expectedError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Check output contains expected strings
			if output != "" {
				for _, expected := range tt.expectedOutput {
					if !strings.Contains(output, expected) {
						t.Errorf("output missing expected string %q\nOutput: %s", expected, output)
					}
				}
			}
		})
	}
}

func TestProxyInstancesCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		serverStatus   int
		expectedError  bool
		expectedOutput []string
		description    string
	}{
		{
			name:          "successful instances list",
			args:          []string{"instances"},
			serverStatus:  200,
			expectedError: false,
			expectedOutput: []string{
				"🖥️  Proxy Instances",
				"==================",
			},
			description: "Test successful proxy instances retrieval",
		},
		{
			name:          "instances with JSON output",
			args:          []string{"instances", "--json"},
			serverStatus:  200,
			expectedError: false,
			expectedOutput: []string{
				"instances",
				"instance-1",
				"instance-2",
			},
			description: "Test proxy instances with JSON output",
		},
		{
			name:          "proxy unavailable",
			args:          []string{"instances"},
			serverStatus:  503,
			expectedError: true,
			expectedOutput: []string{
				"🖥️  Proxy Instances",
				"HTTP 503",
			},
			description: "Test proxy instances when service unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			mockServer := createMockServer(tt.serverStatus)
			defer mockServer.Close()

			// Create command
			cmd := proxyInstancesCmd()
			cmd.SetArgs(tt.args)
			cmd.Flags().Set("api-url", mockServer.URL)

			// Execute command
			err := cmd.Execute()

			// Check error expectation
			if tt.expectedError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestProxyMetricsCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		serverStatus   int
		expectedError  bool
		expectedOutput []string
		description    string
	}{
		{
			name:          "successful metrics",
			args:          []string{"metrics"},
			serverStatus:  200,
			expectedError: false,
			expectedOutput: []string{
				"📊 Proxy Metrics",
				"================",
			},
			description: "Test successful proxy metrics retrieval",
		},
		{
			name:          "metrics with JSON output",
			args:          []string{"metrics", "--json"},
			serverStatus:  200,
			expectedError: false,
			expectedOutput: []string{
				"total_requests",
				"successful_requests",
				"failed_requests",
			},
			description: "Test proxy metrics with JSON output",
		},
		{
			name:          "proxy unavailable",
			args:          []string{"metrics"},
			serverStatus:  503,
			expectedError: true,
			expectedOutput: []string{
				"📊 Proxy Metrics",
				"HTTP 503",
			},
			description: "Test proxy metrics when service unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			mockServer := createMockServer(tt.serverStatus)
			defer mockServer.Close()

			// Create command
			cmd := proxyMetricsCmd()
			cmd.SetArgs(tt.args)
			cmd.Flags().Set("api-url", mockServer.URL)

			// Execute command
			err := cmd.Execute()

			// Check error expectation
			if tt.expectedError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func TestMakeHTTPRequest(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		serverResponse string
		serverStatus   int
		requestBody    interface{}
		expectedError  bool
		description    string
	}{
		{
			name:           "successful GET request",
			method:         "GET",
			serverResponse: `{"status": "ok"}`,
			serverStatus:   200,
			requestBody:    nil,
			expectedError:  false,
			description:    "Test successful GET request",
		},
		{
			name:           "successful POST request with body",
			method:         "POST",
			serverResponse: `{"created": true}`,
			serverStatus:   201,
			requestBody:    map[string]string{"key": "value"},
			expectedError:  false,
			description:    "Test successful POST request with JSON body",
		},
		{
			name:           "HTTP 404 error",
			method:         "GET",
			serverResponse: `{"error": "Not found"}`,
			serverStatus:   404,
			requestBody:    nil,
			expectedError:  true,
			description:    "Test HTTP 404 error handling",
		},
		{
			name:           "HTTP 500 error",
			method:         "GET",
			serverResponse: `{"error": "Internal server error"}`,
			serverStatus:   500,
			requestBody:    nil,
			expectedError:  true,
			description:    "Test HTTP 500 error handling",
		},
		{
			name:           "HTTP 503 service unavailable",
			method:         "GET",
			serverResponse: `{"error": "Service unavailable"}`,
			serverStatus:   503,
			requestBody:    nil,
			expectedError:  true,
			description:    "Test HTTP 503 service unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify method
				if r.Method != tt.method {
					t.Errorf("expected method %s, got %s", tt.method, r.Method)
				}

				// Verify content type for POST requests with body
				if tt.requestBody != nil && r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
				}

				// Verify request body for POST requests
				if tt.requestBody != nil {
					body, err := io.ReadAll(r.Body)
					if err != nil {
						t.Errorf("failed to read request body: %v", err)
					}

					var requestData map[string]string
					if err := json.Unmarshal(body, &requestData); err != nil {
						t.Errorf("failed to unmarshal request body: %v", err)
					}
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			// Make HTTP request
			response, err := makeHTTPRequest(tt.method, server.URL, tt.requestBody)

			// Check error expectation
			if tt.expectedError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Check response for successful requests
			if !tt.expectedError && response != nil {
				if !strings.Contains(string(response), "status") && !strings.Contains(string(response), "created") {
					t.Errorf("unexpected response: %s", string(response))
				}
			}
		})
	}
}

func TestProxyCommandFlags(t *testing.T) {
	tests := []struct {
		name          string
		command       func() *cobra.Command
		args          []string
		expectedFlags map[string]string
		description   string
	}{
		{
			name:    "proxy status default flags",
			command: proxyStatusCmd,
			args:    []string{},
			expectedFlags: map[string]string{
				"api-url": "http://localhost:8080",
				"json":    "false",
			},
			description: "Test proxy status command default flag values",
		},
		{
			name:    "proxy status custom flags",
			command: proxyStatusCmd,
			args:    []string{"--api-url", "http://localhost:9999", "--json"},
			expectedFlags: map[string]string{
				"api-url": "http://localhost:9999",
				"json":    "true",
			},
			description: "Test proxy status command with custom flag values",
		},
		{
			name:    "proxy metrics default flags",
			command: proxyMetricsCmd,
			args:    []string{},
			expectedFlags: map[string]string{
				"api-url":  "http://localhost:8080",
				"json":     "false",
				"watch":    "false",
				"interval": "5",
			},
			description: "Test proxy metrics command default flag values",
		},
		{
			name:    "proxy metrics custom flags",
			command: proxyMetricsCmd,
			args:    []string{"--api-url", "http://localhost:7777", "--json", "--watch", "--interval", "10"},
			expectedFlags: map[string]string{
				"api-url":  "http://localhost:7777",
				"json":     "true",
				"watch":    "true",
				"interval": "10",
			},
			description: "Test proxy metrics command with custom flag values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.command()
			cmd.SetArgs(tt.args)

			// Parse flags
			err := cmd.ParseFlags(tt.args)
			if err != nil {
				t.Errorf("failed to parse flags: %v", err)
			}

			// Check flag values
			for flagName, expectedValue := range tt.expectedFlags {
				actualValue, err := cmd.Flags().GetString(flagName)
				if err != nil {
					// Try as bool for boolean flags
					if boolVal, boolErr := cmd.Flags().GetBool(flagName); boolErr == nil {
						actualValue = fmt.Sprintf("%t", boolVal)
					} else if intVal, intErr := cmd.Flags().GetInt(flagName); intErr == nil {
						actualValue = fmt.Sprintf("%d", intVal)
					} else {
						t.Errorf("failed to get flag %s: %v", flagName, err)
						continue
					}
				}

				if actualValue != expectedValue {
					t.Errorf("flag %s: expected %s, got %s", flagName, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestProxyCommandStructure(t *testing.T) {
	// Test main proxy command
	proxyCmd := proxyCmd()

	if proxyCmd.Use != "proxy" {
		t.Errorf("expected proxy command Use to be 'proxy', got %s", proxyCmd.Use)
	}

	if !strings.Contains(proxyCmd.Short, "proxy") {
		t.Errorf("expected proxy command Short to contain 'proxy', got %s", proxyCmd.Short)
	}

	// Test subcommands exist
	subcommands := proxyCmd.Commands()
	expectedSubcommands := []string{"status", "instances", "metrics"}

	if len(subcommands) != len(expectedSubcommands) {
		t.Errorf("expected %d subcommands, got %d", len(expectedSubcommands), len(subcommands))
	}

	for _, expected := range expectedSubcommands {
		found := false
		for _, cmd := range subcommands {
			if cmd.Use == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand %s not found", expected)
		}
	}
}

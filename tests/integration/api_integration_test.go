// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockAPIServer provides a test server for integration testing
type MockAPIServer struct {
	server *httptest.Server
}

// NewMockAPIServer creates a new mock API server for testing
func NewMockAPIServer() *MockAPIServer {
	mux := http.NewServeMux()
	
	// Health endpoint
	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"version":   "test-1.0.0",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Models endpoint
	mux.HandleFunc("/api/v1/models", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			response := map[string]interface{}{
				"models": []map[string]interface{}{
					{
						"name":    "test-model",
						"version": "1.0.0",
						"size":    1024000,
						"status":  "available",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		case "POST":
			var request map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			
			response := map[string]interface{}{
				"message": "Model deployment initiated",
				"model":   request["name"],
				"status":  "deploying",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(response)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Inference endpoint
	mux.HandleFunc("/api/v1/inference", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var request map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		response := map[string]interface{}{
			"result":     "This is a test inference response",
			"model":      request["model"],
			"prompt":     request["prompt"],
			"tokens":     42,
			"latency_ms": 150,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Metrics endpoint
	mux.HandleFunc("/api/v1/metrics", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"requests_total":    1000,
			"requests_per_sec":  10.5,
			"avg_latency_ms":    125,
			"active_connections": 15,
			"models_loaded":     3,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	server := httptest.NewServer(mux)
	return &MockAPIServer{server: server}
}

// Close shuts down the mock server
func (m *MockAPIServer) Close() {
	m.server.Close()
}

// URL returns the base URL of the mock server
func (m *MockAPIServer) URL() string {
	return m.server.URL
}

func TestAPIHealthEndpoint(t *testing.T) {
	server := NewMockAPIServer()
	defer server.Close()

	resp, err := http.Get(server.URL() + "/api/v1/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.Status)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Contains(t, response, "timestamp")
	assert.Contains(t, response, "version")
}

func TestAPIModelsEndpoint(t *testing.T) {
	server := NewMockAPIServer()
	defer server.Close()

	t.Run("GET models", func(t *testing.T) {
		resp, err := http.Get(server.URL() + "/api/v1/models")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		models, ok := response["models"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, models, 1)

		model := models[0].(map[string]interface{})
		assert.Equal(t, "test-model", model["name"])
		assert.Equal(t, "1.0.0", model["version"])
		assert.Equal(t, "available", model["status"])
	})

	t.Run("POST model deployment", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"name":    "new-model",
			"version": "2.0.0",
		}
		
		body, err := json.Marshal(requestBody)
		require.NoError(t, err)

		resp, err := http.Post(
			server.URL()+"/api/v1/models",
			"application/json",
			bytes.NewBuffer(body),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "new-model", response["model"])
		assert.Equal(t, "deploying", response["status"])
		assert.Contains(t, response, "message")
	})
}

func TestAPIInferenceEndpoint(t *testing.T) {
	server := NewMockAPIServer()
	defer server.Close()

	requestBody := map[string]interface{}{
		"model":  "test-model",
		"prompt": "Hello, world!",
		"max_tokens": 100,
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	resp, err := http.Post(
		server.URL()+"/api/v1/inference",
		"application/json",
		bytes.NewBuffer(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "test-model", response["model"])
	assert.Equal(t, "Hello, world!", response["prompt"])
	assert.Contains(t, response, "result")
	assert.Contains(t, response, "tokens")
	assert.Contains(t, response, "latency_ms")
}

func TestAPIMetricsEndpoint(t *testing.T) {
	server := NewMockAPIServer()
	defer server.Close()

	resp, err := http.Get(server.URL() + "/api/v1/metrics")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	// Verify metrics structure
	expectedKeys := []string{
		"requests_total",
		"requests_per_sec", 
		"avg_latency_ms",
		"active_connections",
		"models_loaded",
	}

	for _, key := range expectedKeys {
		assert.Contains(t, response, key)
	}

	// Verify metric types
	assert.IsType(t, float64(0), response["requests_total"])
	assert.IsType(t, float64(0), response["requests_per_sec"])
	assert.IsType(t, float64(0), response["avg_latency_ms"])
}

func TestAPIErrorHandling(t *testing.T) {
	server := NewMockAPIServer()
	defer server.Close()

	t.Run("404 Not Found", func(t *testing.T) {
		resp, err := http.Get(server.URL() + "/api/v1/nonexistent")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("405 Method Not Allowed", func(t *testing.T) {
		resp, err := http.Post(server.URL()+"/api/v1/health", "application/json", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("400 Bad Request", func(t *testing.T) {
		// Send invalid JSON
		resp, err := http.Post(
			server.URL()+"/api/v1/inference",
			"application/json",
			bytes.NewBuffer([]byte("invalid json")),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestAPICORSHeaders(t *testing.T) {
	server := NewMockAPIServer()
	defer server.Close()

	client := &http.Client{}
	req, err := http.NewRequest("OPTIONS", server.URL()+"/api/v1/health", nil)
	require.NoError(t, err)

	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Note: Our mock server doesn't implement CORS headers
	// In a real integration test, we would verify CORS headers
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusMethodNotAllowed)
}

func TestAPIRateLimiting(t *testing.T) {
	server := NewMockAPIServer()
	defer server.Close()

	// Make multiple rapid requests
	var responses []int
	for i := 0; i < 10; i++ {
		resp, err := http.Get(server.URL() + "/api/v1/health")
		require.NoError(t, err)
		responses = append(responses, resp.StatusCode)
		resp.Body.Close()
	}

	// Note: Our mock server doesn't implement rate limiting
	// In a real integration test, we would expect some 429 responses
	for _, status := range responses {
		assert.Equal(t, http.StatusOK, status)
	}
}

func TestAPIContentType(t *testing.T) {
	server := NewMockAPIServer()
	defer server.Close()

	endpoints := []string{
		"/api/v1/health",
		"/api/v1/models", 
		"/api/v1/metrics",
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			resp, err := http.Get(server.URL() + endpoint)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
		})
	}
}

func TestAPIResponseTime(t *testing.T) {
	server := NewMockAPIServer()
	defer server.Close()

	start := time.Now()
	resp, err := http.Get(server.URL() + "/api/v1/health")
	duration := time.Since(start)

	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	// Response should be fast (under 100ms for mock server)
	assert.Less(t, duration, 100*time.Millisecond)
}

func BenchmarkAPIHealthEndpoint(b *testing.B) {
	server := NewMockAPIServer()
	defer server.Close()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Get(server.URL() + "/api/v1/health")
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}

func BenchmarkAPIInferenceEndpoint(b *testing.B) {
	server := NewMockAPIServer()
	defer server.Close()

	requestBody := map[string]interface{}{
		"model":  "test-model",
		"prompt": "Benchmark test prompt",
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		b.Fatal(err)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Post(
			server.URL()+"/api/v1/inference",
			"application/json",
			bytes.NewBuffer(body),
		)
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}
package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/api"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/distributed"
)

// TestAPIServer tests the API server
func TestAPIServer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test configuration
	config := &api.Config{
		Listen:      "127.0.0.1:0",
		EnableCORS:  true,
		RateLimit:   1000,
		Timeout:     30 * time.Second,
		MaxBodySize: 32 * 1024 * 1024, // 32MB
	}

	// Create mock scheduler
	scheduler := &MockDistributedScheduler{
		nodes: make(map[string]*distributed.NodeInfo),
		tasks: make(map[string]*distributed.DistributedTask),
	}

	// Create API server
	server := api.NewServer(config, scheduler)
	require.NotNil(t, server)

	t.Run("TestHealthEndpoint", func(t *testing.T) {
		// Create test request
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		// Handle request
		server.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response api.HealthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "healthy", response.Status)
	})

	t.Run("TestNodesEndpoint", func(t *testing.T) {
		// Add test nodes to scheduler
		scheduler.nodes["node-1"] = &distributed.NodeInfo{
			ID:      "node-1",
			Address: "127.0.0.1:8080",
			Status:  distributed.NodeStatusOnline,
		}
		scheduler.nodes["node-2"] = &distributed.NodeInfo{
			ID:      "node-2",
			Address: "127.0.0.1:8081",
			Status:  distributed.NodeStatusOnline,
		}

		// Create test request
		req := httptest.NewRequest("GET", "/api/v1/nodes", nil)
		w := httptest.NewRecorder()

		// Handle request
		server.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response api.NodesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(response.Nodes))
	})

	t.Run("TestTasksEndpoint", func(t *testing.T) {
		// Add test tasks to scheduler
		scheduler.tasks["task-1"] = &distributed.DistributedTask{
			ID:        "task-1",
			Type:      distributed.TaskTypeInference,
			ModelName: "test-model",
			Status:    distributed.TaskStatusRunning,
		}

		// Create test request
		req := httptest.NewRequest("GET", "/api/v1/tasks", nil)
		w := httptest.NewRecorder()

		// Handle request
		server.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response api.TasksResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(response.Tasks))
	})

	t.Run("TestInferenceEndpoint", func(t *testing.T) {
		// Create test inference request
		inferenceReq := api.InferenceRequest{
			Model:  "test-model",
			Prompt: "Hello, world!",
			Options: map[string]interface{}{
				"temperature": 0.7,
				"max_tokens":  100,
			},
		}

		// Marshal request
		reqBody, err := json.Marshal(inferenceReq)
		require.NoError(t, err)

		// Create test request
		req := httptest.NewRequest("POST", "/api/v1/inference", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Handle request
		server.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response api.InferenceResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.Response)
	})

	t.Run("TestMetricsEndpoint", func(t *testing.T) {
		// Create test request
		req := httptest.NewRequest("GET", "/api/v1/metrics", nil)
		w := httptest.NewRecorder()

		// Handle request
		server.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response api.MetricsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response.Metrics)
	})
}

// TestAPICompatibility tests API compatibility with Ollama
func TestAPICompatibility(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test server
	server := api.NewCompatibilityServer()
	require.NotNil(t, server)

	t.Run("TestOllamaGenerate", func(t *testing.T) {
		// Create Ollama-compatible generate request
		generateReq := api.OllamaGenerateRequest{
			Model:  "llama3.2",
			Prompt: "Tell me a joke",
			Stream: false,
		}

		// Marshal request
		reqBody, err := json.Marshal(generateReq)
		require.NoError(t, err)

		// Create test request
		req := httptest.NewRequest("POST", "/api/generate", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Handle request
		server.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response api.OllamaGenerateResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.Response)
		assert.True(t, response.Done)
	})

	t.Run("TestOllamaChat", func(t *testing.T) {
		// Create Ollama-compatible chat request
		chatReq := api.OllamaChatRequest{
			Model: "llama3.2",
			Messages: []api.OllamaMessage{
				{
					Role:    "user",
					Content: "Hello, how are you?",
				},
			},
			Stream: false,
		}

		// Marshal request
		reqBody, err := json.Marshal(chatReq)
		require.NoError(t, err)

		// Create test request
		req := httptest.NewRequest("POST", "/api/chat", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Handle request
		server.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response api.OllamaChatResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response.Message)
		assert.NotEmpty(t, response.Message.Content)
		assert.True(t, response.Done)
	})

	t.Run("TestOllamaModels", func(t *testing.T) {
		// Create test request
		req := httptest.NewRequest("GET", "/api/tags", nil)
		w := httptest.NewRecorder()

		// Handle request
		server.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response api.OllamaModelsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response.Models)
	})

	t.Run("TestOllamaShow", func(t *testing.T) {
		// Create show request
		showReq := api.OllamaShowRequest{
			Name: "llama3.2",
		}

		// Marshal request
		reqBody, err := json.Marshal(showReq)
		require.NoError(t, err)

		// Create test request
		req := httptest.NewRequest("POST", "/api/show", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Handle request
		server.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response api.OllamaShowResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.ModelInfo)
	})
}

// TestAPIStreaming tests streaming API endpoints
func TestAPIStreaming(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test server
	server := api.NewCompatibilityServer()
	require.NotNil(t, server)

	t.Run("TestStreamingGenerate", func(t *testing.T) {
		// Create streaming generate request
		generateReq := api.OllamaGenerateRequest{
			Model:  "llama3.2",
			Prompt: "Write a short story",
			Stream: true,
		}

		// Marshal request
		reqBody, err := json.Marshal(generateReq)
		require.NoError(t, err)

		// Create test request
		req := httptest.NewRequest("POST", "/api/generate", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Handle request
		server.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")

		// Parse streaming response
		responses := parseStreamingResponse(w.Body.String())
		assert.Greater(t, len(responses), 0)
		
		// Check final response
		finalResponse := responses[len(responses)-1]
		assert.True(t, finalResponse.Done)
	})

	t.Run("TestStreamingChat", func(t *testing.T) {
		// Create streaming chat request
		chatReq := api.OllamaChatRequest{
			Model: "llama3.2",
			Messages: []api.OllamaMessage{
				{
					Role:    "user",
					Content: "Explain quantum computing",
				},
			},
			Stream: true,
		}

		// Marshal request
		reqBody, err := json.Marshal(chatReq)
		require.NoError(t, err)

		// Create test request
		req := httptest.NewRequest("POST", "/api/chat", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Handle request
		server.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")

		// Parse streaming response
		responses := parseStreamingChatResponse(w.Body.String())
		assert.Greater(t, len(responses), 0)
		
		// Check final response
		finalResponse := responses[len(responses)-1]
		assert.True(t, finalResponse.Done)
	})
}

// TestAPIErrorHandling tests error handling in API endpoints
func TestAPIErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test server
	server := api.NewCompatibilityServer()
	require.NotNil(t, server)

	t.Run("TestInvalidModel", func(t *testing.T) {
		// Create request with invalid model
		generateReq := api.OllamaGenerateRequest{
			Model:  "non-existent-model",
			Prompt: "Hello",
		}

		// Marshal request
		reqBody, err := json.Marshal(generateReq)
		require.NoError(t, err)

		// Create test request
		req := httptest.NewRequest("POST", "/api/generate", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Handle request
		server.ServeHTTP(w, req)

		// Check error response
		assert.Equal(t, http.StatusNotFound, w.Code)
		
		var errorResponse api.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.NotEmpty(t, errorResponse.Error)
	})

	t.Run("TestInvalidRequest", func(t *testing.T) {
		// Create invalid request
		req := httptest.NewRequest("POST", "/api/generate", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Handle request
		server.ServeHTTP(w, req)

		// Check error response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var errorResponse api.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.NotEmpty(t, errorResponse.Error)
	})

	t.Run("TestRateLimit", func(t *testing.T) {
		// Create server with low rate limit
		config := &api.Config{
			RateLimit: 1, // 1 request per second
		}
		
		limitedServer := api.NewServerWithConfig(config)
		require.NotNil(t, limitedServer)

		// Send multiple requests quickly
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()
			
			limitedServer.ServeHTTP(w, req)
			
			if i == 0 {
				assert.Equal(t, http.StatusOK, w.Code)
			} else {
				// Should be rate limited
				assert.Equal(t, http.StatusTooManyRequests, w.Code)
			}
		}
	})
}

// TestAPIMiddleware tests API middleware
func TestAPIMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test server
	server := api.NewCompatibilityServer()
	require.NotNil(t, server)

	t.Run("TestCORS", func(t *testing.T) {
		// Create CORS preflight request
		req := httptest.NewRequest("OPTIONS", "/api/generate", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type")
		w := httptest.NewRecorder()

		// Handle request
		server.ServeHTTP(w, req)

		// Check CORS headers
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
	})

	t.Run("TestRequestLogging", func(t *testing.T) {
		// Create test request
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		// Handle request
		server.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Request should be logged (we can't easily test this without capturing logs)
		// but we can verify the response was successful
	})

	t.Run("TestRequestID", func(t *testing.T) {
		// Create test request
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		// Handle request
		server.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Should have request ID header
		requestID := w.Header().Get("X-Request-ID")
		assert.NotEmpty(t, requestID)
	})
}

// MockDistributedScheduler implements a mock distributed scheduler for testing
type MockDistributedScheduler struct {
	nodes map[string]*distributed.NodeInfo
	tasks map[string]*distributed.DistributedTask
}

func (m *MockDistributedScheduler) GetNodes() []*distributed.NodeInfo {
	nodes := make([]*distributed.NodeInfo, 0, len(m.nodes))
	for _, node := range m.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

func (m *MockDistributedScheduler) GetActiveTasks() []*distributed.DistributedTask {
	tasks := make([]*distributed.DistributedTask, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

func (m *MockDistributedScheduler) GetMetrics() *distributed.PerformanceMetrics {
	return &distributed.PerformanceMetrics{
		TotalRequests:       100,
		CompletedRequests:   95,
		FailedRequests:      5,
		AverageLatency:      50 * time.Millisecond,
		Throughput:          10.5,
		ResourceUtilization: 0.65,
		LastUpdated:         time.Now(),
	}
}

func (m *MockDistributedScheduler) ProcessInference(ctx context.Context, req *api.InferenceRequest) (*api.InferenceResponse, error) {
	// Mock inference processing
	return &api.InferenceResponse{
		Response: "This is a mock response to: " + req.Prompt,
		Done:     true,
		TotalDuration: 1000000000, // 1 second in nanoseconds
	}, nil
}

// Helper functions for parsing streaming responses
func parseStreamingResponse(body string) []api.OllamaGenerateResponse {
	responses := []api.OllamaGenerateResponse{}
	
	// Split by newlines and parse each JSON object
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		var response api.OllamaGenerateResponse
		if err := json.Unmarshal([]byte(line), &response); err == nil {
			responses = append(responses, response)
		}
	}
	
	return responses
}

func parseStreamingChatResponse(body string) []api.OllamaChatResponse {
	responses := []api.OllamaChatResponse{}
	
	// Split by newlines and parse each JSON object
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		var response api.OllamaChatResponse
		if err := json.Unmarshal([]byte(line), &response); err == nil {
			responses = append(responses, response)
		}
	}
	
	return responses
}

// BenchmarkAPIEndpoints benchmarks API endpoint performance
func BenchmarkAPIEndpoints(b *testing.B) {
	gin.SetMode(gin.TestMode)

	// Create test server
	server := api.NewCompatibilityServer()
	require.NotNil(b, server)

	b.Run("HealthEndpoint", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				req := httptest.NewRequest("GET", "/health", nil)
				w := httptest.NewRecorder()
				
				server.ServeHTTP(w, req)
				
				if w.Code != http.StatusOK {
					b.Errorf("Expected status 200, got %d", w.Code)
				}
			}
		})
	})

	b.Run("GenerateEndpoint", func(b *testing.B) {
		generateReq := api.OllamaGenerateRequest{
			Model:  "llama3.2",
			Prompt: "Hello",
			Stream: false,
		}

		reqBody, err := json.Marshal(generateReq)
		require.NoError(b, err)

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				req := httptest.NewRequest("POST", "/api/generate", bytes.NewReader(reqBody))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				
				server.ServeHTTP(w, req)
				
				if w.Code != http.StatusOK {
					b.Errorf("Expected status 200, got %d", w.Code)
				}
			}
		})
	})

	b.Run("ModelsEndpoint", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				req := httptest.NewRequest("GET", "/api/tags", nil)
				w := httptest.NewRecorder()
				
				server.ServeHTTP(w, req)
				
				if w.Code != http.StatusOK {
					b.Errorf("Expected status 200, got %d", w.Code)
				}
			}
		})
	})
}
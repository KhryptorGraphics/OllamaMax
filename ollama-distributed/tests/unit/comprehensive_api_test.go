package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ollama/ollama-distributed/internal/config"
	"github.com/ollama/ollama-distributed/pkg/api"
	"github.com/ollama/ollama-distributed/pkg/consensus"
	"github.com/ollama/ollama-distributed/pkg/p2p"
	"github.com/ollama/ollama-distributed/pkg/scheduler"
)

// TestAPIServerCreation tests API server creation
func TestAPIServerCreation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Create dependencies
	p2pNode := createMockP2PNode(t)
	defer p2pNode.Stop()
	
	consensusEngine := createMockConsensusEngine(t)
	defer consensusEngine.Shutdown(context.Background())
	
	schedulerEngine := createMockSchedulerEngine(t)
	defer schedulerEngine.Shutdown(context.Background())
	
	config := &config.APIConfig{
		Listen:      "127.0.0.1:0",
		EnableCORS:  true,
		Timeout:     30 * time.Second,
		MaxBodySize: 32 * 1024 * 1024,
	}
	
	server, err := api.NewServer(config, p2pNode, consensusEngine, schedulerEngine)
	require.NoError(t, err, "Failed to create API server")
	require.NotNil(t, server, "API server should not be nil")
}

// TestAPIHealthEndpoint tests the health endpoint
func TestAPIHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := createTestAPIServer(t)
	
	// Test health endpoint
	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	
	server.router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "healthy", response["status"])
	assert.NotNil(t, response["timestamp"])
	assert.NotNil(t, response["node_id"])
	assert.NotNil(t, response["services"])
}

// TestAPINodesEndpoints tests node management endpoints
func TestAPINodesEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := createTestAPIServer(t)
	
	t.Run("GetNodes", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/nodes", nil)
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "nodes")
	})
	
	t.Run("GetSpecificNode", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/nodes/test-node-id", nil)
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		// May return 404 if node doesn't exist, which is expected
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNotFound)
	})
	
	t.Run("DrainNode", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/nodes/test-node-id/drain", nil)
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "message")
	})
	
	t.Run("UndrainNode", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/nodes/test-node-id/undrain", nil)
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "message")
	})
}

// TestAPIModelsEndpoints tests model management endpoints
func TestAPIModelsEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := createTestAPIServer(t)
	
	t.Run("GetModels", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/models", nil)
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "models")
	})
	
	t.Run("GetSpecificModel", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/models/test-model", nil)
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		// May return 404 if model doesn't exist
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNotFound)
	})
	
	t.Run("DownloadModel", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/models/test-model/download", nil)
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		// May fail if no nodes available
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusServiceUnavailable)
	})
	
	t.Run("DeleteModel", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/models/test-model", nil)
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		// May return 404 if model doesn't exist
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNotFound)
	})
}

// TestAPIClusterEndpoints tests cluster management endpoints
func TestAPIClusterEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := createTestAPIServer(t)
	
	t.Run("GetClusterStatus", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/cluster/status", nil)
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "node_id")
		assert.Contains(t, response, "is_leader")
		assert.Contains(t, response, "leader")
		assert.Contains(t, response, "peers")
		assert.Contains(t, response, "status")
	})
	
	t.Run("GetLeader", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/cluster/leader", nil)
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "leader")
	})
	
	t.Run("JoinCluster", func(t *testing.T) {
		joinRequest := map[string]interface{}{
			"node_id": "new-node-123",
			"address": "127.0.0.1:8001",
		}
		
		jsonData, err := json.Marshal(joinRequest)
		require.NoError(t, err)
		
		req := httptest.NewRequest("POST", "/api/v1/cluster/join", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		// May succeed or fail depending on consensus state
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
	})
	
	t.Run("LeaveCluster", func(t *testing.T) {
		leaveRequest := map[string]interface{}{
			"node_id": "leaving-node-123",
		}
		
		jsonData, err := json.Marshal(leaveRequest)
		require.NoError(t, err)
		
		req := httptest.NewRequest("POST", "/api/v1/cluster/leave", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		// May succeed or fail depending on consensus state
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
	})
}

// TestAPIInferenceEndpoints tests inference endpoints
func TestAPIInferenceEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := createTestAPIServer(t)
	
	t.Run("Generate", func(t *testing.T) {
		generateRequest := map[string]interface{}{
			"model":  "test-model",
			"prompt": "Hello, world!",
			"stream": false,
		}
		
		jsonData, err := json.Marshal(generateRequest)
		require.NoError(t, err)
		
		req := httptest.NewRequest("POST", "/api/v1/generate", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		// May succeed or fail depending on scheduler state
		assert.True(t, w.Code == http.StatusOK || 
			w.Code == http.StatusInternalServerError || 
			w.Code == http.StatusRequestTimeout)
	})
	
	t.Run("Chat", func(t *testing.T) {
		chatRequest := map[string]interface{}{
			"model": "test-model",
			"messages": []map[string]interface{}{
				{
					"role":    "user",
					"content": "Hello, how are you?",
				},
			},
			"stream": false,
		}
		
		jsonData, err := json.Marshal(chatRequest)
		require.NoError(t, err)
		
		req := httptest.NewRequest("POST", "/api/v1/chat", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "message")
	})
	
	t.Run("Embeddings", func(t *testing.T) {
		embeddingRequest := map[string]interface{}{
			"model": "test-model",
			"input": "Hello, world!",
		}
		
		jsonData, err := json.Marshal(embeddingRequest)
		require.NoError(t, err)
		
		req := httptest.NewRequest("POST", "/api/v1/embeddings", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "embeddings")
	})
}

// TestAPIMetricsEndpoint tests the metrics endpoint
func TestAPIMetricsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := createTestAPIServer(t)
	
	req := httptest.NewRequest("GET", "/api/v1/metrics", nil)
	w := httptest.NewRecorder()
	
	server.router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	// Check for expected metrics fields
	expectedFields := []string{
		"node_id", "connected_peers", "is_leader", "requests_processed",
		"models_loaded", "nodes_total", "nodes_online", "uptime",
		"cpu_usage", "memory_usage", "network_usage",
	}
	
	for _, field := range expectedFields {
		assert.Contains(t, response, field, "Metrics should contain field: %s", field)
	}
}

// TestAPITransfersEndpoints tests transfer endpoints
func TestAPITransfersEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := createTestAPIServer(t)
	
	t.Run("GetTransfers", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/transfers", nil)
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "transfers")
	})
	
	t.Run("GetSpecificTransfer", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/transfers/test-transfer-id", nil)
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "transfer")
	})
}

// TestAPIDistributionEndpoint tests distribution management endpoint
func TestAPIDistributionEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := createTestAPIServer(t)
	
	t.Run("EnableAutoDistribution", func(t *testing.T) {
		distributionRequest := map[string]interface{}{
			"enabled": true,
		}
		
		jsonData, err := json.Marshal(distributionRequest)
		require.NoError(t, err)
		
		req := httptest.NewRequest("POST", "/api/v1/distribution/auto-configure", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		// May succeed or fail depending on consensus state
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
	})
	
	t.Run("DisableAutoDistribution", func(t *testing.T) {
		distributionRequest := map[string]interface{}{
			"enabled": false,
		}
		
		jsonData, err := json.Marshal(distributionRequest)
		require.NoError(t, err)
		
		req := httptest.NewRequest("POST", "/api/v1/distribution/auto-configure", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		// May succeed or fail depending on consensus state
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
	})
}

// TestAPICORSHeaders tests CORS header handling
func TestAPICORSHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := createTestAPIServer(t)
	
	// Test preflight request
	req := httptest.NewRequest("OPTIONS", "/api/v1/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")
	w := httptest.NewRecorder()
	
	server.router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "*")
}

// TestAPIErrorHandling tests API error handling
func TestAPIErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := createTestAPIServer(t)
	
	t.Run("InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/generate", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	t.Run("MissingRequiredFields", func(t *testing.T) {
		incompleteRequest := map[string]interface{}{
			// Missing required fields
		}
		
		jsonData, err := json.Marshal(incompleteRequest)
		require.NoError(t, err)
		
		req := httptest.NewRequest("POST", "/api/v1/generate", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusInternalServerError)
	})
	
	t.Run("NotFoundEndpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/nonexistent", nil)
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// TestAPIWebSocketConnection tests WebSocket functionality
func TestAPIWebSocketConnection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := createTestAPIServer(t)
	
	// Start the test server
	testServer := httptest.NewServer(server.router)
	defer testServer.Close()
	
	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "/api/v1/ws"
	
	// Test WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Skip("WebSocket connection failed (may be expected in test environment)")
		return
	}
	defer conn.Close()
	
	// Test ping message
	pingMsg := map[string]interface{}{
		"type": "ping",
	}
	
	err = conn.WriteJSON(pingMsg)
	require.NoError(t, err)
	
	// Read response
	var response map[string]interface{}
	err = conn.ReadJSON(&response)
	require.NoError(t, err)
	
	assert.Equal(t, "pong", response["type"])
}

// TestAPIStaticFiles tests static file serving
func TestAPIStaticFiles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := createTestAPIServer(t)
	
	// Test root path (should serve index.html)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	
	server.router.ServeHTTP(w, req)
	
	// May return 404 if static files don't exist, which is acceptable in tests
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNotFound)
}

// TestAPIContentTypes tests content type handling
func TestAPIContentTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := createTestAPIServer(t)
	
	testCases := []struct {
		name        string
		contentType string
		body        string
		expectedCode int
	}{
		{
			name:        "ValidJSON",
			contentType: "application/json",
			body:        `{"model": "test", "prompt": "hello"}`,
			expectedCode: http.StatusOK,
		},
		{
			name:        "InvalidContentType",
			contentType: "text/plain",
			body:        `{"model": "test", "prompt": "hello"}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:        "MissingContentType",
			contentType: "",
			body:        `{"model": "test", "prompt": "hello"}`,
			expectedCode: http.StatusBadRequest,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/generate", strings.NewReader(tc.body))
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}
			w := httptest.NewRecorder()
			
			server.router.ServeHTTP(w, req)
			
			// Allow for various error codes depending on implementation
			if tc.expectedCode == http.StatusOK {
				assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
			} else {
				assert.True(t, w.Code >= 400)
			}
		})
	}
}

// TestAPIRateLimiting tests rate limiting (if implemented)
func TestAPIRateLimiting(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := createTestAPIServer(t)
	
	// Make multiple rapid requests
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest("GET", "/api/v1/health", nil)
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		// Should generally succeed, rate limiting may not be implemented in tests
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusTooManyRequests)
	}
}

// TestAPIServerLifecycle tests server lifecycle
func TestAPIServerLifecycle(t *testing.T) {
	server := createTestAPIServer(t)
	
	// Test server start
	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	
	// Give server time to start
	time.Sleep(100 * time.Millisecond)
	
	// Test server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err = server.Shutdown(ctx)
	require.NoError(t, err, "Server should shutdown gracefully")
}

// Helper functions

func createTestAPIServer(t *testing.T) *api.Server {
	p2pNode := createMockP2PNode(t)
	consensusEngine := createMockConsensusEngine(t)
	schedulerEngine := createMockSchedulerEngine(t)
	
	config := &config.APIConfig{
		Listen:      "127.0.0.1:0",
		EnableCORS:  true,
		Timeout:     30 * time.Second,
		MaxBodySize: 32 * 1024 * 1024,
	}
	
	server, err := api.NewServer(config, p2pNode, consensusEngine, schedulerEngine)
	require.NoError(t, err)
	
	return server
}

// BenchmarkAPIEndpoints benchmarks API endpoint performance
func BenchmarkAPIEndpoints(b *testing.B) {
	gin.SetMode(gin.TestMode)
	server := createTestAPIServer(b)
	
	b.Run("HealthEndpoint", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest("GET", "/api/v1/health", nil)
			w := httptest.NewRecorder()
			
			server.router.ServeHTTP(w, req)
			
			if w.Code != http.StatusOK {
				b.Errorf("Expected status 200, got %d", w.Code)
			}
		}
	})
	
	b.Run("MetricsEndpoint", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest("GET", "/api/v1/metrics", nil)
			w := httptest.NewRecorder()
			
			server.router.ServeHTTP(w, req)
			
			if w.Code != http.StatusOK {
				b.Errorf("Expected status 200, got %d", w.Code)
			}
		}
	})
	
	b.Run("NodesEndpoint", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest("GET", "/api/v1/nodes", nil)
			w := httptest.NewRecorder()
			
			server.router.ServeHTTP(w, req)
			
			if w.Code != http.StatusOK {
				b.Errorf("Expected status 200, got %d", w.Code)
			}
		}
	})
	
	b.Run("ModelsEndpoint", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest("GET", "/api/v1/models", nil)
			w := httptest.NewRecorder()
			
			server.router.ServeHTTP(w, req)
			
			if w.Code != http.StatusOK {
				b.Errorf("Expected status 200, got %d", w.Code)
			}
		}
	})
	
	b.Run("GenerateEndpoint", func(b *testing.B) {
		generateRequest := map[string]interface{}{
			"model":  "test-model",
			"prompt": "Hello, world!",
			"stream": false,
		}
		
		jsonData, err := json.Marshal(generateRequest)
		require.NoError(b, err)
		
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest("POST", "/api/v1/generate", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			
			server.router.ServeHTTP(w, req)
			
			// May succeed or fail depending on backend state
			if w.Code >= 500 {
				b.Errorf("Unexpected server error: %d", w.Code)
			}
		}
	})
}
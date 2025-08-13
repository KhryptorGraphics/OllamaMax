package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/consensus"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/monitoring"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/messaging"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for testing

// Create actual instances instead of mocks for testing
func createMockP2PNode(t *testing.T) *p2p.Node {
	ctx := context.Background()
	nodeConfig := &config.NodeConfig{
		Listen:       []string{"/ip4/127.0.0.1/tcp/0"},
		EnableNoise:  true,
		ConnMgrLow:   5,
		ConnMgrHigh:  20,
		ConnMgrGrace: 30 * time.Second,
	}

	node, err := p2p.NewP2PNode(ctx, nodeConfig)
	if err != nil {
		t.Fatalf("Failed to create mock P2P node: %v", err)
	}

	return node
}

func createMockConsensusEngine(t *testing.T) *consensus.Engine {
	ctx := context.Background()

	// Create a minimal P2P node for consensus
	p2pNode := createMockP2PNode(t)

	// Create consensus configuration
	consensusConfig := &config.ConsensusConfig{
		DataDir:   t.TempDir(),
		NodeID:    "test-node",
		Bootstrap: true,
	}

	// Create mock message router and network monitor
	messageRouter := &messaging.MessageRouter{}
	networkMonitor := &monitoring.NetworkMonitor{}

	engine, err := consensus.NewEngine(consensusConfig, p2pNode, messageRouter, networkMonitor)
	if err != nil {
		t.Fatalf("Failed to create mock consensus engine: %v", err)
	}

	return engine
}

func createMockSchedulerEngine(t *testing.T) *scheduler.Engine {
	// Create dependencies
	p2pNode := createMockP2PNode(t)
	consensusEngine := createMockConsensusEngine(t)

	// Create scheduler configuration
	schedulerConfig := &config.SchedulerConfig{
		Algorithm:           "round_robin",
		LoadBalancing:       "least_connections",
		HealthCheckInterval: 30 * time.Second,
		MaxWorkers:          4,
		QueueSize:           1000,
	}

	engine, err := scheduler.NewEngine(schedulerConfig, p2pNode, consensusEngine)
	if err != nil {
		t.Fatalf("Failed to create mock scheduler engine: %v", err)
	}

	return engine
}

type MockConsensusEngine struct {
	mock.Mock
}

func (m *MockConsensusEngine) IsLeader() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockConsensusEngine) Leader() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockConsensusEngine) Apply(key string, value interface{}, metadata map[string]interface{}) error {
	args := m.Called(key, value, metadata)
	return args.Error(0)
}

func (m *MockConsensusEngine) AddVoter(id, address string) error {
	args := m.Called(id, address)
	return args.Error(0)
}

func (m *MockConsensusEngine) RemoveServer(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

type MockSchedulerEngine struct {
	mock.Mock
}

func (m *MockSchedulerEngine) GetNodes() map[string]*scheduler.NodeInfo {
	args := m.Called()
	return args.Get(0).(map[string]*scheduler.NodeInfo)
}

func (m *MockSchedulerEngine) GetAllModels() map[string]*scheduler.ModelInfo {
	args := m.Called()
	return args.Get(0).(map[string]*scheduler.ModelInfo)
}

func (m *MockSchedulerEngine) GetModel(name string) (*scheduler.ModelInfo, bool) {
	args := m.Called(name)
	return args.Get(0).(*scheduler.ModelInfo), args.Bool(1)
}

func (m *MockSchedulerEngine) GetAvailableNodes() []*scheduler.NodeInfo {
	args := m.Called()
	return args.Get(0).([]*scheduler.NodeInfo)
}

func (m *MockSchedulerEngine) RegisterModel(name string, size int64, checksum, nodeID string) error {
	args := m.Called(name, size, checksum, nodeID)
	return args.Error(0)
}

func (m *MockSchedulerEngine) DeleteModel(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockSchedulerEngine) Schedule(req *scheduler.Request) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockSchedulerEngine) GetModelCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockSchedulerEngine) GetOnlineNodeCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockSchedulerEngine) GetStats() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

// Helper functions

func setupTestServer() (*Server, *MockP2PNode, *MockConsensusEngine, *MockSchedulerEngine) {
	gin.SetMode(gin.TestMode)

	mockP2P := &MockP2PNode{}
	mockConsensus := &MockConsensusEngine{}
	mockScheduler := &MockSchedulerEngine{}

	config := &config.APIConfig{
		Listen:      ":0",
		MaxBodySize: 1024 * 1024,
		RateLimit: config.RateLimitConfig{
			RPS: 100,
		},
		Cors: config.CorsConfig{
			AllowedOrigins:   []string{"http://localhost:8080"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
			AllowedHeaders:   []string{"Content-Type", "Authorization"},
			AllowCredentials: true,
			MaxAge:           3600,
		},
	}

	server, err := NewServer(config, mockP2P, mockConsensus, mockScheduler)
	if err != nil {
		panic(err)
	}

	return server, mockP2P, mockConsensus, mockScheduler
}

func createValidJWT() string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  "test-user",
		"username": "testuser",
		"exp":      time.Now().Add(time.Hour).Unix(),
	})

	tokenString, _ := token.SignedString([]byte("your-secret-key"))
	return tokenString
}

// Test cases

func TestNewServer(t *testing.T) {
	server, _, _, _ := setupTestServer()
	assert.NotNil(t, server)
	assert.NotNil(t, server.router)
	assert.NotNil(t, server.wsHub)
}

func TestServer_HealthCheck(t *testing.T) {
	server, mockP2P, _, _ := setupTestServer()

	mockP2P.On("ID").Return("test-node-id")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	req.Header.Set("Authorization", "Bearer "+createValidJWT())

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "test-node-id", response["node_id"])

	mockP2P.AssertExpectations(t)
}

func TestServer_GetClusterStatus(t *testing.T) {
	server, mockP2P, mockConsensus, _ := setupTestServer()

	mockP2P.On("ID").Return("test-node-id")
	mockP2P.On("ConnectedPeers").Return([]string{"peer1", "peer2"})
	mockConsensus.On("IsLeader").Return(true)
	mockConsensus.On("Leader").Return("test-node-id")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/cluster/status", nil)
	req.Header.Set("Authorization", "Bearer "+createValidJWT())

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-node-id", response["node_id"])
	assert.Equal(t, true, response["is_leader"])
	assert.Equal(t, "test-node-id", response["leader"])
	assert.Equal(t, float64(2), response["peers"])

	mockP2P.AssertExpectations(t)
	mockConsensus.AssertExpectations(t)
}

func TestServer_GetNodes(t *testing.T) {
	server, _, _, mockScheduler := setupTestServer()

	testNodes := map[string]*scheduler.NodeInfo{
		"node1": {
			ID:     "node1",
			Status: "online",
		},
		"node2": {
			ID:     "node2",
			Status: "offline",
		},
	}

	mockScheduler.On("GetNodes").Return(testNodes)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/nodes", nil)
	req.Header.Set("Authorization", "Bearer "+createValidJWT())

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "nodes")

	mockScheduler.AssertExpectations(t)
}

func TestServer_GetNode(t *testing.T) {
	server, _, _, mockScheduler := setupTestServer()

	testNode := &scheduler.NodeInfo{
		ID:     "test-node",
		Status: "online",
	}

	testNodes := map[string]*scheduler.NodeInfo{
		"test-node": testNode,
	}

	mockScheduler.On("GetNodes").Return(testNodes)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/nodes/test-node", nil)
	req.Header.Set("Authorization", "Bearer "+createValidJWT())

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "node")

	mockScheduler.AssertExpectations(t)
}

func TestServer_GetNodeNotFound(t *testing.T) {
	server, _, _, mockScheduler := setupTestServer()

	testNodes := map[string]*scheduler.NodeInfo{}
	mockScheduler.On("GetNodes").Return(testNodes)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/nodes/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+createValidJWT())

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Node not found", response["error"])

	mockScheduler.AssertExpectations(t)
}

func TestServer_GetModels(t *testing.T) {
	server, _, _, mockScheduler := setupTestServer()

	testModels := map[string]*scheduler.ModelInfo{
		"llama2": {
			Name: "llama2",
			Size: 7000000000,
		},
		"mixtral": {
			Name: "mixtral",
			Size: 8000000000,
		},
	}

	mockScheduler.On("GetAllModels").Return(testModels)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/models", nil)
	req.Header.Set("Authorization", "Bearer "+createValidJWT())

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "models")

	mockScheduler.AssertExpectations(t)
}

func TestServer_DownloadModel(t *testing.T) {
	server, _, _, mockScheduler := setupTestServer()

	testNode := &scheduler.NodeInfo{
		ID:     "test-node",
		Status: "online",
	}

	mockScheduler.On("GetAvailableNodes").Return([]*scheduler.NodeInfo{testNode})
	mockScheduler.On("RegisterModel", "llama2", int64(0), "", "test-node").Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/models/llama2/download", nil)
	req.Header.Set("Authorization", "Bearer "+createValidJWT())

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["message"], "llama2")
	assert.Equal(t, "test-node", response["target_node"])

	mockScheduler.AssertExpectations(t)
}

func TestServer_DownloadModelNoNodes(t *testing.T) {
	server, _, _, mockScheduler := setupTestServer()

	mockScheduler.On("GetAvailableNodes").Return([]*scheduler.NodeInfo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/models/llama2/download", nil)
	req.Header.Set("Authorization", "Bearer "+createValidJWT())

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "No available nodes")

	mockScheduler.AssertExpectations(t)
}

func TestServer_DeleteModel(t *testing.T) {
	server, _, _, mockScheduler := setupTestServer()

	testModel := &scheduler.ModelInfo{
		Name:      "llama2",
		Locations: []string{"node1", "node2"},
	}

	mockScheduler.On("GetModel", "llama2").Return(testModel, true)
	mockScheduler.On("DeleteModel", "llama2").Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/models/llama2", nil)
	req.Header.Set("Authorization", "Bearer "+createValidJWT())

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["message"], "llama2")

	mockScheduler.AssertExpectations(t)
}

func TestServer_Generate(t *testing.T) {
	server, _, _, mockScheduler := setupTestServer()

	// Mock scheduler response
	mockScheduler.On("Schedule", mock.AnythingOfType("*scheduler.Request")).Run(func(args mock.Arguments) {
		req := args.Get(0).(*scheduler.Request)
		go func() {
			req.ResponseCh <- &scheduler.Response{
				Success: true,
				NodeID:  "test-node",
			}
		}()
	}).Return(nil)

	requestBody := map[string]interface{}{
		"model":  "llama2",
		"prompt": "Hello, how are you?",
		"stream": false,
	}

	body, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/generate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+createValidJWT())

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "response")
	assert.Equal(t, "test-node", response["node_id"])

	mockScheduler.AssertExpectations(t)
}

func TestServer_JoinCluster(t *testing.T) {
	server, _, mockConsensus, _ := setupTestServer()

	mockConsensus.On("AddVoter", "new-node", "127.0.0.1:8080").Return(nil)

	requestBody := map[string]interface{}{
		"node_id": "new-node",
		"address": "127.0.0.1:8080",
	}

	body, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/cluster/join", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+createValidJWT())

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Node joined cluster", response["message"])

	mockConsensus.AssertExpectations(t)
}

func TestServer_LeaveCluster(t *testing.T) {
	server, _, mockConsensus, _ := setupTestServer()

	mockConsensus.On("RemoveServer", "old-node").Return(nil)

	requestBody := map[string]interface{}{
		"node_id": "old-node",
	}

	body, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/cluster/leave", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+createValidJWT())

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Node left cluster", response["message"])

	mockConsensus.AssertExpectations(t)
}

func TestServer_GetMetrics(t *testing.T) {
	server, mockP2P, mockConsensus, mockScheduler := setupTestServer()

	mockP2P.On("ID").Return("test-node-id")
	mockP2P.On("ConnectedPeers").Return([]string{"peer1"})
	mockConsensus.On("IsLeader").Return(true)
	mockScheduler.On("GetModelCount").Return(5)
	mockScheduler.On("GetNodes").Return(map[string]*scheduler.NodeInfo{"node1": {}})
	mockScheduler.On("GetOnlineNodeCount").Return(1)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/metrics", nil)
	req.Header.Set("Authorization", "Bearer "+createValidJWT())

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-node-id", response["node_id"])
	assert.Equal(t, float64(1), response["connected_peers"])
	assert.Equal(t, true, response["is_leader"])
	assert.Equal(t, float64(5), response["models_loaded"])

	mockP2P.AssertExpectations(t)
	mockConsensus.AssertExpectations(t)
	mockScheduler.AssertExpectations(t)
}

func TestServer_AuthMiddleware(t *testing.T) {
	server, _, _, _ := setupTestServer()

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "no auth header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid auth format",
			authHeader:     "InvalidFormat",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "valid token",
			authHeader:     "Bearer " + createValidJWT(),
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/health", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			server.router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestServer_CORSMiddleware(t *testing.T) {
	server, _, _, _ := setupTestServer()

	tests := []struct {
		name           string
		origin         string
		method         string
		expectedStatus int
		expectCORS     bool
	}{
		{
			name:           "allowed origin",
			origin:         "http://localhost:8080",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectCORS:     true,
		},
		{
			name:           "disallowed origin",
			origin:         "http://evil.com",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectCORS:     false,
		},
		{
			name:           "preflight request",
			origin:         "http://localhost:8080",
			method:         "OPTIONS",
			expectedStatus: http.StatusNoContent,
			expectCORS:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, "/api/v1/health", nil)
			req.Header.Set("Origin", tt.origin)
			if tt.method != "OPTIONS" {
				req.Header.Set("Authorization", "Bearer "+createValidJWT())
			}

			server.router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectCORS {
				assert.Equal(t, tt.origin, w.Header().Get("Access-Control-Allow-Origin"))
			} else {
				assert.Equal(t, "", w.Header().Get("Access-Control-Allow-Origin"))
			}
		})
	}
}

func TestServer_RateLimitMiddleware(t *testing.T) {
	server, _, _, _ := setupTestServer()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	req.Header.Set("Authorization", "Bearer "+createValidJWT())

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "100", w.Header().Get("X-RateLimit-Limit"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
}

func TestServer_InputValidationMiddleware(t *testing.T) {
	server, _, _, _ := setupTestServer()

	tests := []struct {
		name           string
		method         string
		contentType    string
		bodySize       int64
		expectedStatus int
	}{
		{
			name:           "valid JSON request",
			method:         "POST",
			contentType:    "application/json",
			bodySize:       100,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unsupported content type",
			method:         "POST",
			contentType:    "text/plain",
			bodySize:       100,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "GET request no validation",
			method:         "GET",
			contentType:    "",
			bodySize:       0,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if tt.bodySize > 0 {
				body = make([]byte, tt.bodySize)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, "/api/v1/health", bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer "+createValidJWT())
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			server.router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check security headers
			assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
			assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
			assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
		})
	}
}

func TestServer_AutoDistribution(t *testing.T) {
	server, _, mockConsensus, _ := setupTestServer()

	mockConsensus.On("Apply", "auto_distribution_enabled", true, mock.Anything).Return(nil)

	requestBody := map[string]interface{}{
		"enabled": true,
	}

	body, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/distribution/auto-configure", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+createValidJWT())

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["message"], "enabled")
	assert.Equal(t, true, response["enabled"])

	mockConsensus.AssertExpectations(t)
}

func TestServer_WSHub(t *testing.T) {
	server, _, _, _ := setupTestServer()

	// Test WebSocket hub functionality
	hub := server.wsHub
	assert.NotNil(t, hub)
	assert.NotNil(t, hub.clients)
	assert.NotNil(t, hub.broadcast)
	assert.NotNil(t, hub.register)
	assert.NotNil(t, hub.unregister)

	// Test broadcast functionality
	testMessage := map[string]interface{}{
		"type": "test",
		"data": "test data",
	}

	// This should not block even if no clients are connected
	hub.Broadcast(testMessage)
}

func TestServer_ErrorHandling(t *testing.T) {
	server, _, _, mockScheduler := setupTestServer()

	t.Run("invalid JSON request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/cluster/join", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+createValidJWT())

		server.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("scheduler error", func(t *testing.T) {
		mockScheduler.On("GetAvailableNodes").Return([]*scheduler.NodeInfo{})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/models/test/download", nil)
		req.Header.Set("Authorization", "Bearer "+createValidJWT())

		server.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		mockScheduler.AssertExpectations(t)
	})
}

func TestServer_StartShutdown(t *testing.T) {
	server, _, _, _ := setupTestServer()

	// Test start
	err := server.Start()
	assert.NoError(t, err)

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Shutdown(ctx)
	assert.NoError(t, err)
}

// Benchmark tests

func BenchmarkServer_HealthCheck(b *testing.B) {
	server, mockP2P, _, _ := setupTestServer()
	mockP2P.On("ID").Return("test-node-id")

	token := createValidJWT()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/health", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			server.router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				b.Fatal("Expected 200, got", w.Code)
			}
		}
	})
}

func BenchmarkServer_GetMetrics(b *testing.B) {
	server, mockP2P, mockConsensus, mockScheduler := setupTestServer()

	mockP2P.On("ID").Return("test-node-id")
	mockP2P.On("ConnectedPeers").Return([]string{})
	mockConsensus.On("IsLeader").Return(false)
	mockScheduler.On("GetModelCount").Return(0)
	mockScheduler.On("GetNodes").Return(map[string]*scheduler.NodeInfo{})
	mockScheduler.On("GetOnlineNodeCount").Return(0)

	token := createValidJWT()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/metrics", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			server.router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				b.Fatal("Expected 200, got", w.Code)
			}
		}
	})
}

func BenchmarkServer_AuthMiddleware(b *testing.B) {
	server, _, _, _ := setupTestServer()
	token := createValidJWT()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/health", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			server.router.ServeHTTP(w, req)
		}
	})
}

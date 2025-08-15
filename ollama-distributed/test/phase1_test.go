package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/api"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/security"
)

// TestPhase1BackendFrontendIntegration tests the Phase 1 implementation
func TestPhase1BackendFrontendIntegration(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Test configuration loading and validation
	t.Run("ConfigurationSystem", func(t *testing.T) {
		testConfigurationSystem(t)
	})

	// Test authentication endpoints
	t.Run("AuthenticationEndpoints", func(t *testing.T) {
		testAuthenticationEndpoints(t)
	})

	// Test dashboard data endpoints
	t.Run("DashboardEndpoints", func(t *testing.T) {
		testDashboardEndpoints(t)
	})

	// Test WebSocket functionality
	t.Run("WebSocketIntegration", func(t *testing.T) {
		testWebSocketIntegration(t)
	})

	// Test JWT token generation
	t.Run("JWTTokenGeneration", func(t *testing.T) {
		testJWTTokenGeneration(t)
	})
}

func testConfigurationSystem(t *testing.T) {
	// Test configuration structure
	config := &config.DistributedConfig{}
	config.SetDefaults()

	// Verify default values are set
	assert.Equal(t, "0.0.0.0", config.API.Host)
	assert.Equal(t, 8080, config.API.Port)
	assert.Equal(t, 9000, config.P2P.Port)
	assert.Equal(t, "raft", config.Consensus.Algorithm)
	assert.Equal(t, "development", config.Node.Environment)

	// Test validation
	config.Node.ID = "test-node-1"
	err := config.Validate()
	assert.NoError(t, err)

	// Test invalid configuration
	config.API.Port = 0
	err = config.Validate()
	assert.Error(t, err)
}

func testAuthenticationEndpoints(t *testing.T) {
	// Create a test server
	server := createTestServer(t)
	router := server.SetupRoutes()

	// Test login endpoint
	t.Run("Login", func(t *testing.T) {
		loginData := map[string]string{
			"email":    "admin@ollamamax.com",
			"password": "admin123",
		}
		body, _ := json.Marshal(loginData)

		req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "user")
		assert.Contains(t, response, "token")
	})

	// Test register endpoint
	t.Run("Register", func(t *testing.T) {
		registerData := map[string]string{
			"email":     "test@example.com",
			"password":  "password123",
			"firstName": "Test",
			"lastName":  "User",
		}
		body, _ := json.Marshal(registerData)

		req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "user")
		assert.Contains(t, response, "message")
	})

	// Test invalid login
	t.Run("InvalidLogin", func(t *testing.T) {
		loginData := map[string]string{
			"email":    "invalid@example.com",
			"password": "wrongpassword",
		}
		body, _ := json.Marshal(loginData)

		req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func testDashboardEndpoints(t *testing.T) {
	server := createTestServer(t)
	router := server.SetupRoutes()

	// Get a valid token first
	token := getValidToken(t, router)

	// Test dashboard data endpoint
	t.Run("DashboardData", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/dashboard/data", nil)
		req.Header.Set("Authorization", token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify dashboard data structure
		assert.Contains(t, response, "clusterStatus")
		assert.Contains(t, response, "nodeCount")
		assert.Contains(t, response, "activeModels")
		assert.Contains(t, response, "nodes")
		assert.Contains(t, response, "timestamp")
	})

	// Test dashboard metrics endpoint
	t.Run("DashboardMetrics", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/dashboard/metrics", nil)
		req.Header.Set("Authorization", token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify metrics data structure
		assert.Contains(t, response, "requestsPerSecond")
		assert.Contains(t, response, "responseTime")
		assert.Contains(t, response, "errorRate")
		assert.Contains(t, response, "resourceUsage")
	})

	// Test unauthorized access
	t.Run("UnauthorizedAccess", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/dashboard/data", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func testWebSocketIntegration(t *testing.T) {
	// Note: WebSocket testing requires more complex setup
	// For now, we'll test that the WebSocket endpoint exists
	server := createTestServer(t)
	router := server.SetupRoutes()

	// Test WebSocket endpoint exists (will fail upgrade but endpoint should exist)
	req := httptest.NewRequest("GET", "/api/v1/ws", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// WebSocket upgrade will fail in test, but endpoint should be found
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}

func testJWTTokenGeneration(t *testing.T) {
	// Test JWT token generation
	token, err := security.GenerateJWT("test@example.com", "user")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Test admin token generation
	adminToken, err := security.GenerateJWT("admin@example.com", "admin")
	assert.NoError(t, err)
	assert.NotEmpty(t, adminToken)
	assert.NotEqual(t, token, adminToken)
}

// Helper functions

func createTestServer(t *testing.T) *api.Server {
	// Create test configuration
	cfg := &config.DistributedConfig{}
	cfg.SetDefaults()
	cfg.Node.ID = "test-node"

	// Create test server
	server := api.NewServer(cfg)
	require.NotNil(t, server)

	return server
}

func getValidToken(t *testing.T, router *gin.Engine) string {
	loginData := map[string]string{
		"email":    "admin@ollamamax.com",
		"password": "admin123",
	}
	body, _ := json.Marshal(loginData)

	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	token, ok := response["token"].(string)
	require.True(t, ok)
	require.NotEmpty(t, token)

	return "Bearer " + token
}

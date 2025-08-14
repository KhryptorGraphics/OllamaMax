package unit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAPIServer tests the API server
func TestAPIServer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Use the helper function to create a test API server
	server := createTestAPIServer(t)
	require.NotNil(t, server)

	t.Run("TestHealthEndpoint", func(t *testing.T) {
		// Create test request
		req := httptest.NewRequest("GET", "/api/v1/health", nil)
		w := httptest.NewRecorder()

		// Handle request
		server.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("TestVersionEndpoint", func(t *testing.T) {
		// Create test request
		req := httptest.NewRequest("GET", "/api/v1/version", nil)
		w := httptest.NewRecorder()

		// Handle request
		server.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)
	})

}

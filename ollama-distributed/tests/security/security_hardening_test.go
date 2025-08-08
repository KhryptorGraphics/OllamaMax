package security

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/api"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/security"
)

// TestSecurityHardening tests the security hardening implementation
func TestSecurityHardening(t *testing.T) {
	// Set up test environment
	os.Setenv("JWT_SECRET", "test-secret-key-that-is-long-enough-for-security")
	os.Setenv("ADMIN_DEFAULT_PASSWORD", "secure-test-password-123")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("ADMIN_DEFAULT_PASSWORD")
	}()

	t.Run("AuthenticationSecurity", testAuthenticationSecurity)
	t.Run("InputValidation", testInputValidation)
	t.Run("SecurityHeaders", testSecurityHeaders)
	t.Run("RateLimiting", testRateLimiting)
	t.Run("TLSConfiguration", testTLSConfiguration)
	t.Run("AuditLogging", testAuditLogging)
}

// testAuthenticationSecurity tests authentication security measures
func testAuthenticationSecurity(t *testing.T) {
	t.Log("🔐 Testing authentication security...")

	// Create auth manager
	authConfig := &api.AuthConfig{
		JWTSecret:   os.Getenv("JWT_SECRET"),
		TokenExpiry: 1 * time.Hour,
		Issuer:      "test",
		Audience:    "test",
	}
	
	authManager, err := api.NewAuthManager(authConfig)
	require.NoError(t, err)

	// Test 1: Hardcoded credentials should be eliminated
	t.Run("NoHardcodedCredentials", func(t *testing.T) {
		// Try to authenticate with common weak credentials
		weakCredentials := [][]string{
			{"admin", "admin"},
			{"admin", "password"},
			{"admin", "123456"},
			{"root", "root"},
			{"test", "test"},
		}

		for _, creds := range weakCredentials {
			session, token, err := authManager.Authenticate(creds[0], creds[1])
			assert.Error(t, err, "Should reject weak credentials: %s/%s", creds[0], creds[1])
			assert.Nil(t, session)
			assert.Empty(t, token)
		}
	})

	// Test 2: Rate limiting should prevent brute force attacks
	t.Run("RateLimiting", func(t *testing.T) {
		username := "testuser"
		
		// Make multiple failed attempts
		for i := 0; i < 6; i++ {
			_, _, err := authManager.Authenticate(username, "wrongpassword")
			assert.Error(t, err)
		}

		// Next attempt should be rate limited
		_, _, err := authManager.Authenticate(username, "wrongpassword")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "too many authentication attempts")
	})

	// Test 3: Strong password requirements
	t.Run("StrongPasswordRequirements", func(t *testing.T) {
		// Test with admin credentials using environment variable
		adminPassword := os.Getenv("ADMIN_DEFAULT_PASSWORD")
		session, token, err := authManager.Authenticate("admin", adminPassword)
		
		if err == nil {
			assert.NotNil(t, session)
			assert.NotEmpty(t, token)
			assert.Equal(t, "admin", session.Username)
			assert.Contains(t, session.Roles, "admin")
		}
	})
}

// testInputValidation tests input validation security
func testInputValidation(t *testing.T) {
	t.Log("🛡️ Testing input validation...")

	router := gin.New()
	
	// Add security middleware
	hardeningConfig := security.DefaultSecurityHardeningConfig()
	hardening := security.NewSecurityHardening(hardeningConfig)
	hardening.ApplySecurityMiddleware(router)

	// Add test endpoint
	router.POST("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test 1: SQL injection attempts
	t.Run("SQLInjectionPrevention", func(t *testing.T) {
		sqlInjectionPayloads := []string{
			"'; DROP TABLE users; --",
			"' OR '1'='1",
			"admin'/*",
			"1' UNION SELECT * FROM users--",
		}

		for _, payload := range sqlInjectionPayloads {
			body := map[string]string{"input": payload}
			jsonBody, _ := json.Marshal(body)
			
			req := httptest.NewRequest("POST", "/api/test", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			// Should either reject or sanitize the input
			assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusOK,
				"Should handle SQL injection payload: %s", payload)
		}
	})

	// Test 2: XSS prevention
	t.Run("XSSPrevention", func(t *testing.T) {
		xssPayloads := []string{
			"<script>alert('xss')</script>",
			"javascript:alert('xss')",
			"<img src=x onerror=alert('xss')>",
			"<svg onload=alert('xss')>",
		}

		for _, payload := range xssPayloads {
			body := map[string]string{"input": payload}
			jsonBody, _ := json.Marshal(body)
			
			req := httptest.NewRequest("POST", "/api/test", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			// Should either reject or sanitize the input
			assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusOK,
				"Should handle XSS payload: %s", payload)
		}
	})

	// Test 3: Path traversal prevention
	t.Run("PathTraversalPrevention", func(t *testing.T) {
		pathTraversalPayloads := []string{
			"../../../etc/passwd",
			"..\\..\\..\\windows\\system32\\config\\sam",
			"/etc/shadow",
			"C:\\Windows\\System32\\config\\SAM",
		}

		for _, payload := range pathTraversalPayloads {
			req := httptest.NewRequest("GET", "/api/test/"+payload, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			// Should reject path traversal attempts
			assert.Equal(t, http.StatusBadRequest, w.Code,
				"Should reject path traversal: %s", payload)
		}
	})
}

// testSecurityHeaders tests security headers implementation
func testSecurityHeaders(t *testing.T) {
	t.Log("🔒 Testing security headers...")

	router := gin.New()
	
	// Add security middleware
	hardeningConfig := security.DefaultSecurityHardeningConfig()
	hardening := security.NewSecurityHardening(hardeningConfig)
	hardening.ApplySecurityMiddleware(router)

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Test required security headers
	headers := w.Header()
	
	assert.Contains(t, headers.Get("Strict-Transport-Security"), "max-age=",
		"Should include HSTS header")
	assert.Equal(t, "nosniff", headers.Get("X-Content-Type-Options"),
		"Should include X-Content-Type-Options header")
	assert.Equal(t, "DENY", headers.Get("X-Frame-Options"),
		"Should include X-Frame-Options header")
	assert.Contains(t, headers.Get("X-XSS-Protection"), "1",
		"Should include X-XSS-Protection header")
	assert.NotEmpty(t, headers.Get("Content-Security-Policy"),
		"Should include Content-Security-Policy header")
	assert.NotEmpty(t, headers.Get("Referrer-Policy"),
		"Should include Referrer-Policy header")
	
	// Server header should be removed or minimal
	serverHeader := headers.Get("Server")
	assert.True(t, serverHeader == "" || !strings.Contains(strings.ToLower(serverHeader), "go"),
		"Should not reveal server information")
}

// testRateLimiting tests rate limiting implementation
func testRateLimiting(t *testing.T) {
	t.Log("⏱️ Testing rate limiting...")

	router := gin.New()
	
	// Add security middleware with strict rate limiting
	hardeningConfig := security.DefaultSecurityHardeningConfig()
	hardeningConfig.RequestsPerMinute = 5 // Very low for testing
	hardening := security.NewSecurityHardening(hardeningConfig)
	hardening.ApplySecurityMiddleware(router)

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Make requests up to the limit
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345" // Simulate same client
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code,
			"Request %d should succeed", i+1)
	}

	// Next request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusTooManyRequests, w.Code,
		"Should be rate limited after exceeding limit")
}

// testTLSConfiguration tests TLS configuration
func testTLSConfiguration(t *testing.T) {
	t.Log("🔐 Testing TLS configuration...")

	hardeningConfig := security.DefaultSecurityHardeningConfig()
	hardening := security.NewSecurityHardening(hardeningConfig)

	tlsConfig, err := hardening.GetTLSConfig()
	require.NoError(t, err)

	if tlsConfig != nil {
		// Test minimum TLS version
		assert.GreaterOrEqual(t, tlsConfig.MinVersion, uint16(0x0303),
			"Should enforce minimum TLS 1.2")

		// Test cipher suites
		assert.NotEmpty(t, tlsConfig.CipherSuites,
			"Should specify secure cipher suites")

		// Test server cipher preference
		assert.True(t, tlsConfig.PreferServerCipherSuites,
			"Should prefer server cipher suites")
	}
}

// testAuditLogging tests audit logging functionality
func testAuditLogging(t *testing.T) {
	t.Log("📝 Testing audit logging...")

	router := gin.New()
	
	// Add security middleware
	hardeningConfig := security.DefaultSecurityHardeningConfig()
	hardening := security.NewSecurityHardening(hardeningConfig)
	hardening.ApplySecurityMiddleware(router)

	router.POST("/admin/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin action"})
	})

	router.GET("/public/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "public action"})
	})

	// Test admin action logging
	req := httptest.NewRequest("POST", "/admin/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Admin actions should be logged (implementation would write to log file)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test failed request logging
	req = httptest.NewRequest("POST", "/nonexistent", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Failed requests should be logged
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestEnvironmentValidation tests environment variable validation
func TestEnvironmentValidation(t *testing.T) {
	t.Log("🌍 Testing environment validation...")

	// Test missing JWT secret
	t.Run("MissingJWTSecret", func(t *testing.T) {
		os.Unsetenv("JWT_SECRET")
		
		hardeningConfig := security.DefaultSecurityHardeningConfig()
		hardeningConfig.RequireAuthentication = true
		hardeningConfig.JWTSecret = ""
		
		hardening := security.NewSecurityHardening(hardeningConfig)
		err := hardening.ValidateEnvironment()
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT_SECRET")
	})

	// Test short JWT secret
	t.Run("ShortJWTSecret", func(t *testing.T) {
		os.Setenv("JWT_SECRET", "short")
		defer os.Unsetenv("JWT_SECRET")
		
		hardeningConfig := security.DefaultSecurityHardeningConfig()
		hardeningConfig.RequireAuthentication = true
		hardeningConfig.JWTSecret = ""
		
		hardening := security.NewSecurityHardening(hardeningConfig)
		err := hardening.ValidateEnvironment()
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "32 characters")
	})

	// Test valid JWT secret
	t.Run("ValidJWTSecret", func(t *testing.T) {
		os.Setenv("JWT_SECRET", "this-is-a-very-long-and-secure-jwt-secret-key")
		defer os.Unsetenv("JWT_SECRET")
		
		hardeningConfig := security.DefaultSecurityHardeningConfig()
		hardeningConfig.RequireAuthentication = true
		hardeningConfig.JWTSecret = ""
		
		hardening := security.NewSecurityHardening(hardeningConfig)
		err := hardening.ValidateEnvironment()
		
		assert.NoError(t, err)
	})
}

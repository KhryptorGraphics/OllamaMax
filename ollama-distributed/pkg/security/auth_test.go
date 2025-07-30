package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSecurityManager(t *testing.T) {
	cfg := &config.SecurityConfig{
		TLS: config.TLSConfig{
			Enabled:  true,
			CertFile: "test-cert.pem",
			KeyFile:  "test-key.pem",
		},
		Auth: config.AuthConfig{
			Enabled:     true,
			Method:      "jwt",
			TokenExpiry: 24 * time.Hour,
			SecretKey:   "test-secret-key",
		},
	}

	manager, err := NewSecurityManager(cfg)
	require.NoError(t, err)
	require.NotNil(t, manager)
	
	defer manager.Close()
}

func TestGenerateSelfSignedCert(t *testing.T) {
	manager := &SecurityManager{}
	
	cert, key, err := manager.GenerateSelfSignedCert("localhost", []string{"127.0.0.1"}, 365*24*time.Hour)
	require.NoError(t, err)
	require.NotEmpty(t, cert)
	require.NotEmpty(t, key)
	
	// Verify certificate can be parsed
	block, _ := pem.Decode(cert)
	require.NotNil(t, block)
	
	parsedCert, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)
	assert.Equal(t, "localhost", parsedCert.Subject.CommonName)
}

func TestJWTMiddleware(t *testing.T) {
	cfg := &config.SecurityConfig{
		Auth: config.AuthConfig{
			Enabled:     true,
			Method:      "jwt",
			TokenExpiry: time.Hour,
			SecretKey:   "test-secret-key",
		},
	}

	manager, err := NewSecurityManager(cfg)
	require.NoError(t, err)
	defer manager.Close()

	// Generate test token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "test-user",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte("test-secret-key"))
	require.NoError(t, err)

	// Create test router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(manager.JWTMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test with valid token
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test with invalid token
	req = httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Test without token
	req = httptest.NewRequest("GET", "/protected", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAPIKeyMiddleware(t *testing.T) {
	cfg := &config.SecurityConfig{
		Auth: config.AuthConfig{
			Enabled: true,
			Method:  "api_key",
		},
	}

	manager, err := NewSecurityManager(cfg)
	require.NoError(t, err)
	defer manager.Close()

	// Add test API key
	testAPIKey := "test-api-key-12345"
	manager.AddAPIKey(testAPIKey, "test-user", []string{"read", "write"})

	// Create test router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(manager.APIKeyMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test with valid API key (header)
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("X-API-Key", testAPIKey)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test with valid API key (query param)
	req = httptest.NewRequest("GET", "/protected?api_key="+testAPIKey, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test with invalid API key
	req = httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("X-API-Key", "invalid-key")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRateLimitMiddleware(t *testing.T) {
	cfg := &config.SecurityConfig{
		RateLimit: config.RateLimitConfig{
			Enabled:       true,
			RequestsPerMinute: 2,
			BurstSize:     1,
		},
	}

	manager, err := NewSecurityManager(cfg)
	require.NoError(t, err)
	defer manager.Close()

	// Create test router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(manager.RateLimitMiddleware())
	router.GET("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// First request should succeed
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Second request should succeed (within burst)
	req = httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Third request should be rate limited
	req = httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestSecurityHeaders(t *testing.T) {
	cfg := &config.SecurityConfig{
		Headers: config.SecurityHeadersConfig{
			Enabled: true,
			CSP:     "default-src 'self'",
			HSTS:    "max-age=31536000",
		},
	}

	manager, err := NewSecurityManager(cfg)
	require.NoError(t, err)
	defer manager.Close()

	// Create test router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(manager.SecurityHeadersMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "default-src 'self'", w.Header().Get("Content-Security-Policy"))
	assert.Equal(t, "max-age=31536000", w.Header().Get("Strict-Transport-Security"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
}

func TestEncryptDecrypt(t *testing.T) {
	manager := &SecurityManager{}
	
	plaintext := "This is a test message for encryption"
	key := []byte("32-byte-key-for-aes256-encryption!!")
	
	// Test encryption
	encrypted, err := manager.Encrypt([]byte(plaintext), key)
	require.NoError(t, err)
	require.NotEmpty(t, encrypted)
	assert.NotEqual(t, plaintext, string(encrypted))
	
	// Test decryption
	decrypted, err := manager.Decrypt(encrypted, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, string(decrypted))
}

func TestHashPassword(t *testing.T) {
	manager := &SecurityManager{}
	
	password := "test-password-123"
	
	// Test hashing
	hash, err := manager.HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
	
	// Test verification with correct password
	valid := manager.VerifyPassword(password, hash)
	assert.True(t, valid)
	
	// Test verification with incorrect password
	invalid := manager.VerifyPassword("wrong-password", hash)
	assert.False(t, invalid)
}

func TestGenerateAPIKey(t *testing.T) {
	manager := &SecurityManager{}
	
	// Test API key generation
	apiKey, err := manager.GenerateAPIKey("test-user", []string{"read", "write"})
	require.NoError(t, err)
	require.NotEmpty(t, apiKey)
	
	// API key should be reasonable length
	assert.GreaterOrEqual(t, len(apiKey), 32)
	
	// Generate another key and ensure it's different
	apiKey2, err := manager.GenerateAPIKey("test-user-2", []string{"read"})
	require.NoError(t, err)
	assert.NotEqual(t, apiKey, apiKey2)
}

func TestValidateJWT(t *testing.T) {
	manager := &SecurityManager{
		jwtSecret: []byte("test-secret-key"),
	}
	
	// Create valid token
	claims := jwt.MapClaims{
		"sub": "test-user",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(manager.jwtSecret)
	require.NoError(t, err)
	
	// Test validation
	parsedClaims, err := manager.ValidateJWT(tokenString)
	require.NoError(t, err)
	require.NotNil(t, parsedClaims)
	assert.Equal(t, "test-user", parsedClaims["sub"])
	
	// Test with invalid token
	_, err = manager.ValidateJWT("invalid.jwt.token")
	assert.Error(t, err)
	
	// Test with expired token
	expiredClaims := jwt.MapClaims{
		"sub": "test-user",
		"exp": time.Now().Add(-time.Hour).Unix(),
		"iat": time.Now().Add(-2*time.Hour).Unix(),
	}
	
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredTokenString, err := expiredToken.SignedString(manager.jwtSecret)
	require.NoError(t, err)
	
	_, err = manager.ValidateJWT(expiredTokenString)
	assert.Error(t, err)
}

func TestCORSMiddleware(t *testing.T) {
	cfg := &config.SecurityConfig{
		CORS: config.CORSConfig{
			Enabled:      true,
			AllowOrigins: []string{"https://example.com"},
			AllowMethods: []string{"GET", "POST"},
			AllowHeaders: []string{"Content-Type", "Authorization"},
		},
	}

	manager, err := NewSecurityManager(cfg)
	require.NoError(t, err)
	defer manager.Close()

	// Create test router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(manager.CORSMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test preflight request
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")

	// Test actual request
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
}
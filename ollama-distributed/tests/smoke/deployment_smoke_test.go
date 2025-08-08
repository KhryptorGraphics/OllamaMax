package smoke

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	stagingURL    = flag.String("staging-url", "", "Staging environment URL")
	productionURL = flag.String("production-url", "", "Production environment URL")
	timeout       = flag.Duration("timeout", 30*time.Second, "Request timeout")
)

// SmokeTestSuite runs basic smoke tests against deployed environments
type SmokeTestSuite struct {
	baseURL string
	client  *http.Client
}

// NewSmokeTestSuite creates a new smoke test suite
func NewSmokeTestSuite(baseURL string) *SmokeTestSuite {
	return &SmokeTestSuite{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client: &http.Client{
			Timeout: *timeout,
		},
	}
}

// TestStagingDeployment tests staging environment
func TestStagingDeployment(t *testing.T) {
	if *stagingURL == "" {
		t.Skip("Staging URL not provided, skipping staging smoke tests")
	}

	suite := NewSmokeTestSuite(*stagingURL)
	t.Logf("🧪 Running staging smoke tests against: %s", *stagingURL)

	suite.runSmokeTests(t, "staging")
}

// TestProductionDeployment tests production environment
func TestProductionDeployment(t *testing.T) {
	if *productionURL == "" {
		t.Skip("Production URL not provided, skipping production smoke tests")
	}

	suite := NewSmokeTestSuite(*productionURL)
	t.Logf("🧪 Running production smoke tests against: %s", *productionURL)

	suite.runSmokeTests(t, "production")
}

// runSmokeTests executes the complete smoke test suite
func (s *SmokeTestSuite) runSmokeTests(t *testing.T, environment string) {
	t.Run("HealthCheck", func(t *testing.T) {
		s.testHealthEndpoint(t)
	})

	t.Run("APIEndpoints", func(t *testing.T) {
		s.testAPIEndpoints(t)
	})

	t.Run("ProxyFunctionality", func(t *testing.T) {
		s.testProxyFunctionality(t)
	})

	t.Run("SecurityHeaders", func(t *testing.T) {
		s.testSecurityHeaders(t)
	})

	t.Run("Performance", func(t *testing.T) {
		s.testPerformance(t)
	})

	if environment == "production" {
		t.Run("ProductionSpecific", func(t *testing.T) {
			s.testProductionSpecific(t)
		})
	}
}

// testHealthEndpoint tests the health endpoint
func (s *SmokeTestSuite) testHealthEndpoint(t *testing.T) {
	t.Log("🏥 Testing health endpoint...")

	resp, err := s.client.Get(s.baseURL + "/health")
	require.NoError(t, err, "Health endpoint should be accessible")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Health endpoint should return 200")

	// Check response time
	assert.True(t, resp.Header.Get("Date") != "", "Response should have Date header")
}

// testAPIEndpoints tests core API endpoints
func (s *SmokeTestSuite) testAPIEndpoints(t *testing.T) {
	t.Log("🔌 Testing API endpoints...")

	endpoints := []struct {
		path           string
		expectedStatus int
		description    string
	}{
		{"/api/v1/proxy/status", http.StatusOK, "Proxy status endpoint"},
		{"/api/v1/proxy/instances", http.StatusOK, "Proxy instances endpoint"},
		{"/api/v1/proxy/metrics", http.StatusOK, "Proxy metrics endpoint"},
		{"/api/v1/health", http.StatusOK, "API health endpoint"},
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.description, func(t *testing.T) {
			resp, err := s.client.Get(s.baseURL + endpoint.path)
			require.NoError(t, err, "Endpoint %s should be accessible", endpoint.path)
			defer resp.Body.Close()

			// Accept both success and auth required (401) as valid responses
			validStatuses := []int{endpoint.expectedStatus, http.StatusUnauthorized}
			assert.Contains(t, validStatuses, resp.StatusCode,
				"Endpoint %s should return valid status", endpoint.path)
		})
	}
}

// testProxyFunctionality tests proxy-specific functionality
func (s *SmokeTestSuite) testProxyFunctionality(t *testing.T) {
	t.Log("🎛️ Testing proxy functionality...")

	// Test proxy status endpoint with JSON response
	resp, err := s.client.Get(s.baseURL + "/api/v1/proxy/status")
	require.NoError(t, err, "Proxy status should be accessible")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var statusData map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&statusData)
		assert.NoError(t, err, "Status response should be valid JSON")

		// Check for expected fields
		assert.Contains(t, statusData, "status", "Status response should contain status field")
	} else if resp.StatusCode == http.StatusUnauthorized {
		t.Log("ℹ️ Proxy status requires authentication (expected in production)")
	} else {
		t.Errorf("Unexpected status code: %d", resp.StatusCode)
	}
}

// testSecurityHeaders tests security headers
func (s *SmokeTestSuite) testSecurityHeaders(t *testing.T) {
	t.Log("🔒 Testing security headers...")

	resp, err := s.client.Get(s.baseURL + "/health")
	require.NoError(t, err, "Should be able to make request for header check")
	defer resp.Body.Close()

	securityHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1",
	}

	for header, expectedValue := range securityHeaders {
		headerValue := resp.Header.Get(header)
		if headerValue != "" {
			assert.Contains(t, headerValue, expectedValue,
				"Security header %s should contain expected value", header)
		} else {
			t.Logf("⚠️ Security header %s is missing", header)
		}
	}

	// Check for HSTS header (should be present in production)
	hstsHeader := resp.Header.Get("Strict-Transport-Security")
	if hstsHeader != "" {
		assert.Contains(t, hstsHeader, "max-age", "HSTS header should contain max-age")
		t.Log("✅ HSTS header present")
	} else {
		t.Log("ℹ️ HSTS header not present (may be expected in development)")
	}
}

// testPerformance tests basic performance characteristics
func (s *SmokeTestSuite) testPerformance(t *testing.T) {
	t.Log("⚡ Testing performance...")

	// Test response time
	start := time.Now()
	resp, err := s.client.Get(s.baseURL + "/health")
	duration := time.Since(start)

	require.NoError(t, err, "Performance test request should succeed")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Performance test should return 200")

	// Response time should be reasonable
	maxResponseTime := 5 * time.Second
	assert.True(t, duration < maxResponseTime,
		"Response time should be less than %v, got %v", maxResponseTime, duration)

	t.Logf("✅ Response time: %v", duration)

	// Test multiple concurrent requests
	concurrentRequests := 5
	results := make(chan time.Duration, concurrentRequests)

	for i := 0; i < concurrentRequests; i++ {
		go func() {
			start := time.Now()
			resp, err := s.client.Get(s.baseURL + "/health")
			duration := time.Since(start)
			if err == nil {
				resp.Body.Close()
			}
			results <- duration
		}()
	}

	// Collect results
	var totalDuration time.Duration
	for i := 0; i < concurrentRequests; i++ {
		duration := <-results
		totalDuration += duration
	}

	avgDuration := totalDuration / time.Duration(concurrentRequests)
	t.Logf("✅ Average concurrent response time: %v", avgDuration)

	assert.True(t, avgDuration < maxResponseTime,
		"Average concurrent response time should be reasonable")
}

// testProductionSpecific tests production-specific requirements
func (s *SmokeTestSuite) testProductionSpecific(t *testing.T) {
	t.Log("🏭 Testing production-specific requirements...")

	// Test HTTPS enforcement
	if strings.HasPrefix(s.baseURL, "https://") {
		t.Log("✅ Using HTTPS")
	} else {
		t.Error("❌ Production should use HTTPS")
	}

	// Test that debug endpoints are not exposed
	debugEndpoints := []string{
		"/debug/pprof/",
		"/debug/vars",
		"/metrics", // Raw metrics should be protected
	}

	for _, endpoint := range debugEndpoints {
		resp, err := s.client.Get(s.baseURL + endpoint)
		if err == nil {
			defer resp.Body.Close()
			assert.NotEqual(t, http.StatusOK, resp.StatusCode,
				"Debug endpoint %s should not be publicly accessible in production", endpoint)
		}
	}

	// Test authentication is required for sensitive endpoints
	sensitiveEndpoints := []string{
		"/api/v1/admin/",
		"/api/v1/config/",
	}

	for _, endpoint := range sensitiveEndpoints {
		resp, err := s.client.Get(s.baseURL + endpoint)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				t.Errorf("Sensitive endpoint %s should require authentication", endpoint)
			}
		}
	}
}

// TestEnvironmentVariables tests environment-specific configuration
func TestEnvironmentVariables(t *testing.T) {
	t.Log("🌍 Testing environment variables...")

	// Check required environment variables for deployment
	requiredEnvVars := []string{
		"JWT_SECRET",
	}

	for _, envVar := range requiredEnvVars {
		value := os.Getenv(envVar)
		if value == "" {
			t.Logf("⚠️ Environment variable %s is not set", envVar)
		} else {
			t.Logf("✅ Environment variable %s is set", envVar)
		}
	}
}

// TestDeploymentInfo tests deployment information endpoints
func TestDeploymentInfo(t *testing.T) {
	baseURL := ""
	if *stagingURL != "" {
		baseURL = *stagingURL
	} else if *productionURL != "" {
		baseURL = *productionURL
	} else {
		t.Skip("No deployment URL provided")
	}

	suite := NewSmokeTestSuite(baseURL)
	t.Logf("🔍 Testing deployment info for: %s", baseURL)

	// Test version endpoint if available
	resp, err := suite.client.Get(baseURL + "/api/v1/version")
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			var versionData map[string]interface{}
			if json.NewDecoder(resp.Body).Decode(&versionData) == nil {
				t.Logf("✅ Version info: %+v", versionData)
			}
		}
	}

	// Test build info endpoint if available
	resp, err = suite.client.Get(baseURL + "/api/v1/build")
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			var buildData map[string]interface{}
			if json.NewDecoder(resp.Body).Decode(&buildData) == nil {
				t.Logf("✅ Build info: %+v", buildData)
			}
		}
	}
}

// BenchmarkHealthEndpoint benchmarks the health endpoint
func BenchmarkHealthEndpoint(b *testing.B) {
	baseURL := ""
	if *stagingURL != "" {
		baseURL = *stagingURL
	} else if *productionURL != "" {
		baseURL = *productionURL
	} else {
		b.Skip("No deployment URL provided")
	}

	suite := NewSmokeTestSuite(baseURL)
	healthURL := baseURL + "/health"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := suite.client.Get(healthURL)
			if err != nil {
				b.Error(err)
				continue
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				b.Errorf("Expected 200, got %d", resp.StatusCode)
			}
		}
	})
}

// Helper function to check if URL is accessible
func isURLAccessible(url string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode < 500
}

// TestMain sets up test environment
func TestMain(m *testing.M) {
	flag.Parse()

	// Validate URLs if provided
	if *stagingURL != "" && !isURLAccessible(*stagingURL+"/health") {
		fmt.Printf("⚠️ Warning: Staging URL %s may not be accessible\n", *stagingURL)
	}

	if *productionURL != "" && !isURLAccessible(*productionURL+"/health") {
		fmt.Printf("⚠️ Warning: Production URL %s may not be accessible\n", *productionURL)
	}

	// Run tests
	code := m.Run()
	os.Exit(code)
}

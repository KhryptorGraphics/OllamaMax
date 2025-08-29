package proxy

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
)

func TestNewOllamaProxy(t *testing.T) {
	cfg := &config.Config{
		Proxy: &config.ProxyConfig{
			Strategy:        "round_robin",
			HealthCheckPath: "/api/health",
			MaxRetries:      3,
			RetryDelay:      time.Second,
		},
	}

	backends := []string{
		"http://localhost:8081",
		"http://localhost:8082",
	}

	proxy, err := NewOllamaProxy(cfg, backends)
	require.NoError(t, err)
	require.NotNil(t, proxy)

	assert.Equal(t, cfg.Proxy.Strategy, proxy.config.Strategy)
	assert.Len(t, proxy.backends, 2)
	assert.NotNil(t, proxy.router)
	assert.NotNil(t, proxy.healthChecker)
}

func TestOllamaProxy_ServeHTTP_RoundRobin(t *testing.T) {
	// Create test backend servers
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("backend1"))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("backend2"))
	}))
	defer backend2.Close()

	cfg := &config.Config{
		Proxy: &config.ProxyConfig{
			Strategy:        "round_robin",
			HealthCheckPath: "/api/health",
			MaxRetries:      1,
			RetryDelay:      100 * time.Millisecond,
		},
	}

	proxy, err := NewOllamaProxy(cfg, []string{backend1.URL, backend2.URL})
	require.NoError(t, err)

	// Mark backends as healthy
	proxy.healthChecker.SetBackendHealth(backend1.URL, true)
	proxy.healthChecker.SetBackendHealth(backend2.URL, true)

	// Test round robin distribution
	responses := make([]string, 4)
	for i := 0; i < 4; i++ {
		req := httptest.NewRequest("GET", "/api/generate", nil)
		w := httptest.NewRecorder()

		proxy.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		responses[i] = w.Body.String()
	}

	// Should alternate between backends
	assert.Equal(t, "backend1", responses[0])
	assert.Equal(t, "backend2", responses[1])
	assert.Equal(t, "backend1", responses[2])
	assert.Equal(t, "backend2", responses[3])
}

func TestOllamaProxy_ServeHTTP_WeightedRoundRobin(t *testing.T) {
	// Create test backend servers
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("backend1"))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("backend2"))
	}))
	defer backend2.Close()

	cfg := &config.Config{
		Proxy: &config.ProxyConfig{
			Strategy:        "weighted_round_robin",
			HealthCheckPath: "/api/health",
			MaxRetries:      1,
			RetryDelay:      100 * time.Millisecond,
		},
	}

	proxy, err := NewOllamaProxy(cfg, []string{backend1.URL, backend2.URL})
	require.NoError(t, err)

	// Set different weights (backend1: 3, backend2: 1)
	proxy.router.SetBackendWeight(backend1.URL, 3)
	proxy.router.SetBackendWeight(backend2.URL, 1)

	// Mark backends as healthy
	proxy.healthChecker.SetBackendHealth(backend1.URL, true)
	proxy.healthChecker.SetBackendHealth(backend2.URL, true)

	// Test weighted distribution over 8 requests
	responses := make([]string, 8)
	for i := 0; i < 8; i++ {
		req := httptest.NewRequest("GET", "/api/generate", nil)
		w := httptest.NewRecorder()

		proxy.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		responses[i] = w.Body.String()
	}

	// Count responses from each backend
	backend1Count := 0
	backend2Count := 0
	for _, resp := range responses {
		if resp == "backend1" {
			backend1Count++
		} else if resp == "backend2" {
			backend2Count++
		}
	}

	// Backend1 should get ~3x more requests than backend2
	assert.Greater(t, backend1Count, backend2Count)
	assert.True(t, backend1Count >= 6) // Should get at least 6 out of 8 requests
}

func TestOllamaProxy_ServeHTTP_LeastConnections(t *testing.T) {
	// Create test backend servers with delays
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("backend1"))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("backend2"))
	}))
	defer backend2.Close()

	cfg := &config.Config{
		Proxy: &config.ProxyConfig{
			Strategy:        "least_connections",
			HealthCheckPath: "/api/health",
			MaxRetries:      1,
			RetryDelay:      100 * time.Millisecond,
		},
	}

	proxy, err := NewOllamaProxy(cfg, []string{backend1.URL, backend2.URL})
	require.NoError(t, err)

	// Mark backends as healthy
	proxy.healthChecker.SetBackendHealth(backend1.URL, true)
	proxy.healthChecker.SetBackendHealth(backend2.URL, true)

	// Make concurrent requests
	done := make(chan string, 3)
	for i := 0; i < 3; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/api/generate", nil)
			w := httptest.NewRecorder()
			proxy.ServeHTTP(w, req)
			done <- w.Body.String()
		}()
	}

	responses := make([]string, 3)
	for i := 0; i < 3; i++ {
		responses[i] = <-done
	}

	// Backend2 (faster) should get more requests due to least connections
	backend2Count := 0
	for _, resp := range responses {
		if resp == "backend2" {
			backend2Count++
		}
	}
	assert.GreaterOrEqual(t, backend2Count, 2)
}

func TestOllamaProxy_ServeHTTP_UnhealthyBackend(t *testing.T) {
	// Create test backend servers
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("backend1"))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("backend2"))
	}))
	defer backend2.Close()

	cfg := &config.Config{
		Proxy: &config.ProxyConfig{
			Strategy:        "round_robin",
			HealthCheckPath: "/api/health",
			MaxRetries:      2,
			RetryDelay:      100 * time.Millisecond,
		},
	}

	proxy, err := NewOllamaProxy(cfg, []string{backend1.URL, backend2.URL})
	require.NoError(t, err)

	// Mark only backend2 as healthy
	proxy.healthChecker.SetBackendHealth(backend1.URL, false)
	proxy.healthChecker.SetBackendHealth(backend2.URL, true)

	// All requests should go to backend2
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/api/generate", nil)
		w := httptest.NewRecorder()

		proxy.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "backend2", w.Body.String())
	}
}

func TestOllamaProxy_ServeHTTP_AllBackendsUnhealthy(t *testing.T) {
	cfg := &config.Config{
		Proxy: &config.ProxyConfig{
			Strategy:        "round_robin",
			HealthCheckPath: "/api/health",
			MaxRetries:      1,
			RetryDelay:      100 * time.Millisecond,
		},
	}

	proxy, err := NewOllamaProxy(cfg, []string{"http://localhost:9999"})
	require.NoError(t, err)

	// Mark backend as unhealthy
	proxy.healthChecker.SetBackendHealth("http://localhost:9999", false)

	req := httptest.NewRequest("GET", "/api/generate", nil)
	w := httptest.NewRecorder()

	proxy.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "no healthy backends available")
}

func TestOllamaProxy_ServeHTTP_RetryLogic(t *testing.T) {
	// Create a backend that fails first, then succeeds
	failCount := 0
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if failCount < 2 {
			failCount++
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer backend.Close()

	cfg := &config.Config{
		Proxy: &config.ProxyConfig{
			Strategy:        "round_robin",
			HealthCheckPath: "/api/health",
			MaxRetries:      3,
			RetryDelay:      50 * time.Millisecond,
		},
	}

	proxy, err := NewOllamaProxy(cfg, []string{backend.URL})
	require.NoError(t, err)

	// Mark backend as healthy
	proxy.healthChecker.SetBackendHealth(backend.URL, true)

	req := httptest.NewRequest("GET", "/api/generate", nil)
	w := httptest.NewRecorder()

	proxy.ServeHTTP(w, req)

	// Should eventually succeed after retries
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "success", w.Body.String())
}

func TestOllamaProxy_ServeHTTP_StreamingRequest(t *testing.T) {
	// Create backend that handles streaming
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request body
		body, _ := io.ReadAll(r.Body)
		assert.Contains(t, string(body), "stream")

		// Set streaming headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		w.WriteHeader(http.StatusOK)

		// Write streaming response
		flusher := w.(http.Flusher)
		for i := 0; i < 3; i++ {
			w.Write([]byte("data: chunk " + string(rune('1'+i)) + "\n\n"))
			flusher.Flush()
			time.Sleep(10 * time.Millisecond)
		}
	}))
	defer backend.Close()

	cfg := &config.Config{
		Proxy: &config.ProxyConfig{
			Strategy:        "round_robin",
			HealthCheckPath: "/api/health",
			MaxRetries:      1,
			RetryDelay:      100 * time.Millisecond,
		},
	}

	proxy, err := NewOllamaProxy(cfg, []string{backend.URL})
	require.NoError(t, err)

	// Mark backend as healthy
	proxy.healthChecker.SetBackendHealth(backend.URL, true)

	// Create streaming request
	body := strings.NewReader(`{"model": "test", "stream": true}`)
	req := httptest.NewRequest("POST", "/api/generate", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	proxy.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "chunk 1")
	assert.Contains(t, w.Body.String(), "chunk 2")
	assert.Contains(t, w.Body.String(), "chunk 3")
}

func TestOllamaProxy_WebSocketUpgrade(t *testing.T) {
	// Create backend that handles WebSocket upgrade
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for WebSocket upgrade headers
		if r.Header.Get("Upgrade") == "websocket" {
			w.Header().Set("Upgrade", "websocket")
			w.Header().Set("Connection", "Upgrade")
			w.WriteHeader(http.StatusSwitchingProtocols)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer backend.Close()

	cfg := &config.Config{
		Proxy: &config.ProxyConfig{
			Strategy:        "round_robin",
			HealthCheckPath: "/api/health",
			MaxRetries:      1,
			RetryDelay:      100 * time.Millisecond,
		},
	}

	proxy, err := NewOllamaProxy(cfg, []string{backend.URL})
	require.NoError(t, err)

	// Mark backend as healthy
	proxy.healthChecker.SetBackendHealth(backend.URL, true)

	// Create WebSocket upgrade request
	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Key", "test-key")
	req.Header.Set("Sec-WebSocket-Version", "13")

	w := httptest.NewRecorder()
	proxy.ServeHTTP(w, req)

	assert.Equal(t, http.StatusSwitchingProtocols, w.Code)
	assert.Equal(t, "websocket", w.Header().Get("Upgrade"))
	assert.Equal(t, "Upgrade", w.Header().Get("Connection"))
}

func TestOllamaProxy_GetStats(t *testing.T) {
	cfg := &config.Config{
		Proxy: &config.ProxyConfig{
			Strategy:        "round_robin",
			HealthCheckPath: "/api/health",
			MaxRetries:      1,
			RetryDelay:      100 * time.Millisecond,
		},
	}

	proxy, err := NewOllamaProxy(cfg, []string{
		"http://localhost:8081",
		"http://localhost:8082",
	})
	require.NoError(t, err)

	stats := proxy.GetStats()
	assert.NotNil(t, stats)
	assert.Len(t, stats.Backends, 2)
	assert.GreaterOrEqual(t, stats.TotalRequests, uint64(0))
	assert.GreaterOrEqual(t, stats.FailedRequests, uint64(0))
}

func TestOllamaProxy_CircuitBreaker(t *testing.T) {
	// Create backend that always fails
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer backend.Close()

	cfg := &config.Config{
		Proxy: &config.ProxyConfig{
			Strategy:               "round_robin",
			HealthCheckPath:        "/api/health",
			MaxRetries:             1,
			RetryDelay:             10 * time.Millisecond,
			CircuitBreakerEnabled:  true,
			CircuitBreakerFailures: 3,
			CircuitBreakerTimeout:  100 * time.Millisecond,
		},
	}

	proxy, err := NewOllamaProxy(cfg, []string{backend.URL})
	require.NoError(t, err)

	// Mark backend as healthy initially
	proxy.healthChecker.SetBackendHealth(backend.URL, true)

	// Make requests until circuit breaker opens
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/api/generate", nil)
		w := httptest.NewRecorder()
		proxy.ServeHTTP(w, req)

		if i < 3 {
			// First 3 requests should get 500 from backend
			assert.Equal(t, http.StatusInternalServerError, w.Code)
		} else {
			// After circuit opens, should get 503
			assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		}
	}

	// Wait for circuit breaker timeout
	time.Sleep(150 * time.Millisecond)

	// Circuit should be half-open now, allowing one request through
	req := httptest.NewRequest("GET", "/api/generate", nil)
	w := httptest.NewRecorder()
	proxy.ServeHTTP(w, req)

	// Should still fail but circuit is testing
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRequestRouter_SelectBackend_Sticky(t *testing.T) {
	cfg := &config.Config{
		Proxy: &config.ProxyConfig{
			Strategy: "sticky",
		},
	}

	backends := []string{
		"http://backend1:8080",
		"http://backend2:8080",
		"http://backend3:8080",
	}

	router, err := NewRequestRouter(cfg.Proxy, backends)
	require.NoError(t, err)

	// Mark all backends as healthy
	for _, backend := range backends {
		router.SetBackendHealth(backend, true)
	}

	// Create request with session cookie
	req1 := httptest.NewRequest("GET", "/api/generate", nil)
	req1.AddCookie(&http.Cookie{
		Name:  "session",
		Value: "user123",
	})

	backend1 := router.SelectBackend(req1)
	require.NotEmpty(t, backend1)

	// Same session should get same backend
	req2 := httptest.NewRequest("GET", "/api/chat", nil)
	req2.AddCookie(&http.Cookie{
		Name:  "session",
		Value: "user123",
	})

	backend2 := router.SelectBackend(req2)
	assert.Equal(t, backend1, backend2)

	// Different session should potentially get different backend
	req3 := httptest.NewRequest("GET", "/api/generate", nil)
	req3.AddCookie(&http.Cookie{
		Name:  "session",
		Value: "user456",
	})

	backend3 := router.SelectBackend(req3)
	require.NotEmpty(t, backend3)

	// user456 should consistently get the same backend
	req4 := httptest.NewRequest("GET", "/api/chat", nil)
	req4.AddCookie(&http.Cookie{
		Name:  "session",
		Value: "user456",
	})

	backend4 := router.SelectBackend(req4)
	assert.Equal(t, backend3, backend4)
}

func TestHealthChecker_StartStop(t *testing.T) {
	backends := []string{
		"http://localhost:9999", // Non-existent backend
	}

	checker := NewHealthChecker(backends, "/api/health", 100*time.Millisecond, 3)

	// Start health checking
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go checker.Start(ctx)

	// Wait for health check
	time.Sleep(150 * time.Millisecond)

	// Backend should be marked as unhealthy
	assert.False(t, checker.IsHealthy("http://localhost:9999"))

	// Stop health checker
	cancel()
	time.Sleep(50 * time.Millisecond) // Allow goroutine to stop
}
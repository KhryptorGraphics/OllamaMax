package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/internal/config"
)

func TestNewServer(t *testing.T) {
	cfg := config.DefaultConfig()

	server, err := NewServer(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	if server == nil {
		t.Fatal("NewServer() returned nil server")
	}

	if server.config != cfg {
		t.Error("Server config not set correctly")
	}
}

func TestServerStart(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.API.Listen = ":0" // Use random available port

	server, err := NewServer(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Start(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Stop server
	cancel()

	select {
	case err := <-errChan:
		if err != nil && err != context.Canceled && err != http.ErrServerClosed {
			t.Errorf("Server.Start() unexpected error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Server did not stop within timeout")
	}
}

func TestServerShutdown(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.API.Listen = ":0"

	server, err := NewServer(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	ctx := context.Background()
	go server.Start(ctx)

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		t.Errorf("Server.Shutdown() error = %v", err)
	}
}

func TestHealthCheckHandler(t *testing.T) {
	cfg := config.DefaultConfig()
	server, err := NewServer(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	// Test canonical health endpoint
	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusPartialContent {
		t.Errorf("Health check status = %v, want %v or %v", w.Code, http.StatusOK, http.StatusPartialContent)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %v, want application/json", contentType)
	}

	// Test legacy health endpoint
	req2 := httptest.NewRequest("GET", "/health", nil)
	w2 := httptest.NewRecorder()
	server.router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK && w2.Code != http.StatusPartialContent {
		t.Errorf("Legacy health check status = %v, want %v or %v", w2.Code, http.StatusOK, http.StatusPartialContent)
	}
}

func TestAPIVersionHandler(t *testing.T) {
	cfg := config.DefaultConfig()
	server, err := NewServer(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	req := httptest.NewRequest("GET", "/api/version", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Version endpoint status = %v, want %v", w.Code, http.StatusOK)
	}
}

func TestCORSMiddleware(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.API.Cors.Enabled = true
	cfg.API.Cors.AllowedOrigins = []string{"http://localhost:3000"}

	server, err := NewServer(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	req := httptest.NewRequest("OPTIONS", "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "http://localhost:3000" {
		t.Errorf("Access-Control-Allow-Origin = %v, want http://localhost:3000", allowOrigin)
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.API.RateLimit.Enabled = true
	cfg.API.RateLimit.RequestsPer = 2
	cfg.API.RateLimit.Duration = time.Second
	cfg.API.RateLimit.BurstSize = 1

	server, err := NewServer(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	// Make requests up to the limit using canonical endpoint
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/api/v1/health", nil)
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		if i < 2 {
			if w.Code != http.StatusOK && w.Code != http.StatusPartialContent {
				t.Errorf("Request %d status = %v, want %v or %v", i, w.Code, http.StatusOK, http.StatusPartialContent)
			}
		}
	}

	// Next request should be rate limited
	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Rate limited request status = %v, want %v", w.Code, http.StatusTooManyRequests)
	}
}

func TestAuthMiddleware(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Auth.Enabled = true

	server, err := NewServer(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

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
			name:           "invalid token",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestNotFoundHandler(t *testing.T) {
	cfg := config.DefaultConfig()
	server, err := NewServer(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	req := httptest.NewRequest("GET", "/non-existent-endpoint", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Not found status = %v, want %v", w.Code, http.StatusNotFound)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	cfg := config.DefaultConfig()
	server, err := NewServer(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	// Assuming /health only allows GET
	req := httptest.NewRequest("POST", "/health", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed && w.Code != http.StatusOK {
		t.Errorf("Method not allowed status = %v", w.Code)
	}
}

func TestServerMetrics(t *testing.T) {
	cfg := config.DefaultConfig()
	server, err := NewServer(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	// Make some requests to generate metrics using canonical endpoint
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/api/v1/health", nil)
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
	}

	metrics := server.GetMetrics()
	if metrics.RequestCount < 5 {
		t.Errorf("RequestCount = %v, want >= 5", metrics.RequestCount)
	}
}

func TestConcurrentRequests(t *testing.T) {
	cfg := config.DefaultConfig()
	server, err := NewServer(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	const numRequests = 100
	done := make(chan bool, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/api/v1/health", nil)
			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, req)

			if w.Code != http.StatusOK && w.Code != http.StatusPartialContent {
				t.Errorf("Concurrent request failed with status %v", w.Code)
			}

			done <- true
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		<-done
	}
}

package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/khryptorgraphics/ollamamax/internal/config"
	"github.com/khryptorgraphics/ollamamax/pkg/database"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"
)

// ServerMetrics holds Prometheus metrics for the server
type ServerMetrics struct {
	// Prometheus metrics
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	httpRequestsInFlight prometheus.Gauge

	// Legacy fields for backward compatibility
	StartTime time.Time
	mu        sync.RWMutex

	// Prometheus registry
	registry *prometheus.Registry
}

// Server represents a simplified API server for OllamaMax
type Server struct {
	config   *config.Config
	db       *database.DatabaseManager
	logger   *slog.Logger
	router   *gin.Engine
	server   *http.Server
	shutdown chan struct{}
	wg       sync.WaitGroup
	metrics  *ServerMetrics

	// SECURITY: Rate limiters for authentication endpoints
	authRateLimiters map[string]*authRateLimiter
	authRLMutex      sync.RWMutex
}

// authRateLimiter holds rate limiter per IP for authentication endpoints
type authRateLimiter struct {
	limiter   *rate.Limiter
	lastSeen  time.Time
	attempts  int
}

// NewServer creates a new simplified API server
// This constructor is exported for use in tests
func NewServer(cfg *config.Config, db *database.DatabaseManager, logger *slog.Logger) (*Server, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Use default logger if none provided
	if logger == nil {
		logger = slog.Default()
	}

	// Initialize Prometheus registry
	registry := prometheus.NewRegistry()

	// Initialize Prometheus metrics
	metrics := &ServerMetrics{
		StartTime: time.Now(),
		registry:  registry,
	}

	// HTTP requests total counter
	metrics.httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ollamamax_api_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTP request duration histogram
	metrics.httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ollamamax_api_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	// HTTP requests in flight gauge
	metrics.httpRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ollamamax_api_http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
	)

	// Register metrics with the registry
	registry.MustRegister(metrics.httpRequestsTotal)
	registry.MustRegister(metrics.httpRequestDuration)
	registry.MustRegister(metrics.httpRequestsInFlight)

	s := &Server{
		config:           cfg,
		db:               db,
		logger:           logger,
		shutdown:         make(chan struct{}),
		metrics:          metrics,
		authRateLimiters: make(map[string]*authRateLimiter),
	}

	s.setupRouter()

	// Start cleanup goroutine for expired rate limiters
	go s.cleanupExpiredRateLimiters()

	return s, nil
}

// GetMetrics returns current server metrics (legacy compatibility)
func (s *Server) GetMetrics() ServerMetrics {
	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()

	return ServerMetrics{
		StartTime: s.metrics.StartTime,
		registry:  s.metrics.registry,
	}
}

// Shutdown is an alias for Stop for backward compatibility
func (s *Server) Shutdown(ctx context.Context) error {
	return s.Stop()
}

// setupRouter configures the Gin router
func (s *Server) setupRouter() {
	gin.SetMode(gin.ReleaseMode)
	s.router = gin.New()

	// Add middleware
	s.router.Use(gin.Recovery())
	s.router.Use(s.loggingMiddleware())
	s.router.Use(s.corsMiddleware())

	// Metrics endpoint
	s.router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(s.metrics.registry, promhttp.HandlerOpts{})))

	// Health endpoint - canonical path at /api/v1/health
	s.router.GET("/api/v1/health", s.healthHandler)
	s.router.GET("/health", s.healthHandler) // Legacy support

	// Basic API endpoints
	api := s.router.Group("/api")
	{
		api.GET("/version", s.versionHandler) // Canonical: /api/version
		api.GET("/v1/version", s.versionHandler) // Legacy support
		api.GET("/v1/status", s.statusHandler)
	}

	// SECURITY FIX (ISSUE-007): Add rate limiting to authentication endpoints
	// These endpoints are common attack vectors
	authGroup := s.router.Group("/api/v1/auth")
	authGroup.Use(s.authRateLimitMiddleware())
	{
		// Placeholder endpoints - actual implementation would go here
		// authGroup.POST("/login", s.loginHandler)
		// authGroup.POST("/register", s.registerHandler)
		// authGroup.POST("/reset-password", s.resetPasswordHandler)
	}
}

// loggingMiddleware provides basic request logging and metrics tracking
func (s *Server) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		method := c.Request.Method

		// Skip metrics recording for /metrics endpoint to avoid recursion
		skipMetrics := path == "/metrics"

		// Increment in-flight requests
		if !skipMetrics {
			s.metrics.httpRequestsInFlight.Inc()
		}

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		statusCode := c.Writer.Status()

		// Record Prometheus metrics
		if !skipMetrics {
			statusStr := fmt.Sprintf("%d", statusCode)

			// Normalize path to avoid high cardinality
			normalizedPath := s.normalizePath(path)

			// Record metrics with labels
			s.metrics.httpRequestsTotal.WithLabelValues(method, normalizedPath, statusStr).Inc()
			s.metrics.httpRequestDuration.WithLabelValues(method, normalizedPath, statusStr).Observe(latency.Seconds())
			s.metrics.httpRequestsInFlight.Dec()
		}

		if raw != "" {
			path = path + "?" + raw
		}

		s.logger.Info("HTTP request",
			"method", method,
			"path", path,
			"status", statusCode,
			"latency", latency,
			"client_ip", clientIP,
		)
	}
}

// normalizePath normalizes request paths to reduce cardinality in metrics
func (s *Server) normalizePath(path string) string {
	// Map known paths to normalized versions
	knownPaths := map[string]string{
		"/api/v1/health": "/api/v1/health",
		"/health":        "/health",
		"/api/version":   "/api/version",
		"/api/v1/version": "/api/v1/version",
		"/api/v1/status": "/api/v1/status",
		"/metrics":       "/metrics",
	}

	if normalized, ok := knownPaths[path]; ok {
		return normalized
	}

	// For unknown paths, use a generic label
	return "/other"
}

// corsMiddleware provides CORS support with configurable allowed origins
func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// SECURITY FIX (ISSUE-006): Use specific allowed origins instead of "*"
		// Get allowed origins from config or use localhost defaults for development
		allowedOrigins := s.config.API.Cors.AllowedOrigins
		if len(allowedOrigins) == 0 {
			// Default to localhost for development if not configured
			allowedOrigins = []string{"http://localhost:3000", "http://localhost:8080"}
		}

		origin := c.GetHeader("Origin")
		allowed := false

		// Check if origin is in allowed list
		for _, allowedOrigin := range allowedOrigins {
			if allowedOrigin == origin || allowedOrigin == "*" {
				allowed = true
				c.Header("Access-Control-Allow-Origin", allowedOrigin)
				break
			}
		}

		// If origin not allowed, use first allowed origin (for OPTIONS preflight)
		if !allowed && len(allowedOrigins) > 0 {
			c.Header("Access-Control-Allow-Origin", allowedOrigins[0])
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "3600")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// healthHandler handles health check requests
func (s *Server) healthHandler(c *gin.Context) {
	ctx := c.Request.Context()

	responseData := gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
	}

	// Check database health if available
	if s.db != nil {
		health, err := s.db.Health(ctx)
		if err != nil {
			responseData["status"] = "degraded"
			responseData["database"] = gin.H{
				"status": "unhealthy",
				"error":  err.Error(),
			}
		} else {
			responseData["database"] = health
		}
	} else {
		responseData["status"] = "degraded"
		responseData["database"] = gin.H{
			"status": "unavailable",
			"error":  "database not connected",
		}
	}

	statusCode := http.StatusOK
	if responseData["status"] == "degraded" {
		statusCode = http.StatusPartialContent
	}

	c.JSON(statusCode, responseData)
}

// versionHandler handles version requests
func (s *Server) versionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version":   "1.0.0",
		"service":   "OllamaMax",
		"timestamp": time.Now(),
	})
}

// statusHandler handles status requests
func (s *Server) statusHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":     "running",
		"service":    "OllamaMax",
		"timestamp":  time.Now(),
		"version":    "1.0.0",
		"listen":     s.config.API.Listen,
	})
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	s.server = &http.Server{
		Addr:         s.config.API.Listen,
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	s.logger.Info("Starting HTTP server", "listen", s.config.API.Listen)

	// Start server in goroutine
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		var err error
		if s.config.API.TLSEnabled {
			err = s.server.ListenAndServeTLS(s.config.API.CertFile, s.config.API.KeyFile)
		} else {
			err = s.server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server error", "error", err)
			close(s.shutdown)
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case <-ctx.Done():
		s.logger.Info("Shutdown signal received")
		return s.Stop()
	case <-s.shutdown:
		s.logger.Error("Server shutdown due to error")
		return fmt.Errorf("server error")
	}
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop() error {
	if s.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.logger.Info("Shutting down HTTP server")

	// Shutdown server
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("Error shutting down server", "error", err)
		return err
	}

	// Wait for goroutines to finish
	s.wg.Wait()

	s.logger.Info("HTTP server stopped")
	return nil
}

// GetRouter returns the Gin router for testing
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

// authRateLimitMiddleware implements strict rate limiting for authentication endpoints
// SECURITY FIX (ISSUE-007): Protects against brute force attacks
func (s *Server) authRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		path := c.Request.URL.Path

		// Get rate limit config from environment or use defaults
		var requestsPerMinute rate.Limit
		var burstSize int

		switch {
		case c.Request.URL.Path == "/api/v1/auth/login":
			// Login: 5 attempts per minute
			requestsPerMinute = rate.Limit(s.config.API.RateLimit.LoginRequestsPer) / 60.0
			burstSize = 2
		case c.Request.URL.Path == "/api/v1/auth/register":
			// Register: 3 attempts per minute
			requestsPerMinute = rate.Limit(s.config.API.RateLimit.RegisterRequestsPer) / 60.0
			burstSize = 1
		case c.Request.URL.Path == "/api/v1/auth/reset-password":
			// Password reset: 3 attempts per minute
			requestsPerMinute = rate.Limit(s.config.API.RateLimit.ResetPasswordRequestsPer) / 60.0
			burstSize = 1
		default:
			// Other auth endpoints: 10 attempts per minute
			requestsPerMinute = 10.0 / 60.0
			burstSize = 3
		}

		s.authRLMutex.Lock()
		limiter, exists := s.authRateLimiters[clientIP]
		if !exists {
			limiter = &authRateLimiter{
				limiter:  rate.NewLimiter(requestsPerMinute, burstSize),
				lastSeen: time.Now(),
				attempts: 0,
			}
			s.authRateLimiters[clientIP] = limiter
		}
		limiter.lastSeen = time.Now()
		limiter.attempts++
		s.authRLMutex.Unlock()

		if !limiter.limiter.Allow() {
			s.logger.Warn("Auth rate limit exceeded",
				"client_ip", clientIP,
				"path", path,
				"attempts", limiter.attempts,
			)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate_limit_exceeded",
				"message":     "Too many authentication attempts. Please try again later.",
				"retry_after": 60, // seconds
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// cleanupExpiredRateLimiters periodically removes expired rate limiters
func (s *Server) cleanupExpiredRateLimiters() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.authRLMutex.Lock()
			now := time.Now()
			for ip, limiter := range s.authRateLimiters {
				// Remove limiters that haven't been used in 30 minutes
				if now.Sub(limiter.lastSeen) > 30*time.Minute {
					delete(s.authRateLimiters, ip)
				}
			}
			s.authRLMutex.Unlock()
		case <-s.shutdown:
			return
		}
	}
}

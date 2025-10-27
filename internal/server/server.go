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
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTP request duration histogram
	metrics.httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	// HTTP requests in flight gauge
	metrics.httpRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
	)

	// Register metrics with the registry
	registry.MustRegister(metrics.httpRequestsTotal)
	registry.MustRegister(metrics.httpRequestDuration)
	registry.MustRegister(metrics.httpRequestsInFlight)

	s := &Server{
		config:   cfg,
		db:       db,
		logger:   logger,
		shutdown: make(chan struct{}),
		metrics:  metrics,
	}

	s.setupRouter()
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

// corsMiddleware provides CORS support
func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

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

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
)

// Server represents a simplified API server for OllamaMax
type Server struct {
	config   *config.Config
	db       *database.DatabaseManager
	logger   *slog.Logger
	router   *gin.Engine
	server   *http.Server
	shutdown chan struct{}
	wg       sync.WaitGroup
}

// NewServer creates a new simplified API server
func NewServer(cfg *config.Config, db *database.DatabaseManager, logger *slog.Logger) *Server {
	s := &Server{
		config:   cfg,
		db:       db,
		logger:   logger,
		shutdown: make(chan struct{}),
	}

	s.setupRouter()
	return s
}

// setupRouter configures the Gin router
func (s *Server) setupRouter() {
	gin.SetMode(gin.ReleaseMode)
	s.router = gin.New()

	// Add middleware
	s.router.Use(gin.Recovery())
	s.router.Use(s.loggingMiddleware())
	s.router.Use(s.corsMiddleware())

	// Health endpoint
	s.router.GET("/health", s.healthHandler)
	s.router.GET("/api/v1/health", s.healthHandler)

	// Basic API endpoints
	api := s.router.Group("/api/v1")
	{
		api.GET("/version", s.versionHandler)
		api.GET("/status", s.statusHandler)
	}
}

// loggingMiddleware provides basic request logging
func (s *Server) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

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
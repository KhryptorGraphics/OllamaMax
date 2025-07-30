package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// HTTPServerConfig configures the HTTP server
type HTTPServerConfig struct {
	Listen       string
	TLSEnabled   bool
	CertFile     string
	KeyFile      string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// HTTPMetrics tracks HTTP server performance
type HTTPMetrics struct {
	RequestsTotal       int64         `json:"requests_total"`
	RequestsSuccess     int64         `json:"requests_success"`
	RequestsError       int64         `json:"requests_error"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	ConnectionsActive   int64         `json:"connections_active"`
	ConnectionsTotal    int64         `json:"connections_total"`
	LastUpdated         time.Time     `json:"last_updated"`
	mu                  sync.RWMutex
}

// NewHTTPServer creates a new HTTP server
func NewHTTPServer(config *HTTPServerConfig) (*HTTPServer, error) {
	if config == nil {
		config = &HTTPServerConfig{
			Listen:       ":8080",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		}
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	server := &HTTPServer{
		config: config,
		router: gin.New(),
		metrics: &HTTPMetrics{
			LastUpdated: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}
	
	// Setup basic middleware
	server.setupMiddleware()
	
	return server, nil
}

// setupMiddleware sets up basic middleware
func (hs *HTTPServer) setupMiddleware() {
	hs.router.Use(gin.Logger())
	hs.router.Use(gin.Recovery())
	hs.router.Use(hs.metricsMiddleware())
}

// metricsMiddleware tracks request metrics
func (hs *HTTPServer) metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Update active connections
		hs.metrics.mu.Lock()
		hs.metrics.ConnectionsActive++
		hs.metrics.ConnectionsTotal++
		hs.metrics.mu.Unlock()
		
		c.Next()
		
		// Update metrics after request
		duration := time.Since(start)
		
		hs.metrics.mu.Lock()
		hs.metrics.ConnectionsActive--
		hs.metrics.RequestsTotal++
		
		if c.Writer.Status() >= 200 && c.Writer.Status() < 400 {
			hs.metrics.RequestsSuccess++
		} else {
			hs.metrics.RequestsError++
		}
		
		// Update average response time
		if hs.metrics.RequestsTotal == 1 {
			hs.metrics.AverageResponseTime = duration
		} else {
			hs.metrics.AverageResponseTime = (hs.metrics.AverageResponseTime + duration) / 2
		}
		
		hs.metrics.LastUpdated = time.Now()
		hs.metrics.mu.Unlock()
	}
}

// Start starts the HTTP server
func (hs *HTTPServer) Start() error {
	hs.server = &http.Server{
		Addr:         hs.config.Listen,
		Handler:      hs.router,
		ReadTimeout:  hs.config.ReadTimeout,
		WriteTimeout: hs.config.WriteTimeout,
		IdleTimeout:  hs.config.IdleTimeout,
	}
	
	go func() {
		var err error
		if hs.config.TLSEnabled {
			err = hs.server.ListenAndServeTLS(hs.config.CertFile, hs.config.KeyFile)
		} else {
			err = hs.server.ListenAndServe()
		}
		
		if err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()
	
	return nil
}

// Stop stops the HTTP server
func (hs *HTTPServer) Stop() error {
	hs.cancel()
	
	if hs.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return hs.server.Shutdown(ctx)
	}
	
	return nil
}

// GetRouter returns the Gin router
func (hs *HTTPServer) GetRouter() *gin.Engine {
	return hs.router
}

// GetMetrics returns HTTP server metrics
func (hs *HTTPServer) GetMetrics() *HTTPMetrics {
	hs.metrics.mu.RLock()
	defer hs.metrics.mu.RUnlock()
	
	// Create a copy
	metrics := *hs.metrics
	return &metrics
}

// AddRoute adds a route to the HTTP server
func (hs *HTTPServer) AddRoute(method, path string, handler gin.HandlerFunc) {
	switch method {
	case "GET":
		hs.router.GET(path, handler)
	case "POST":
		hs.router.POST(path, handler)
	case "PUT":
		hs.router.PUT(path, handler)
	case "DELETE":
		hs.router.DELETE(path, handler)
	case "PATCH":
		hs.router.PATCH(path, handler)
	default:
		hs.router.Any(path, handler)
	}
}

// AddMiddleware adds middleware to the HTTP server
func (hs *HTTPServer) AddMiddleware(middleware gin.HandlerFunc) {
	hs.router.Use(middleware)
}

// AddRouteGroup adds a route group to the HTTP server
func (hs *HTTPServer) AddRouteGroup(path string, middleware ...gin.HandlerFunc) *gin.RouterGroup {
	group := hs.router.Group(path)
	for _, mw := range middleware {
		group.Use(mw)
	}
	return group
}

// ServeStatic serves static files
func (hs *HTTPServer) ServeStatic(relativePath, root string) {
	hs.router.Static(relativePath, root)
}

// ServeStaticFile serves a single static file
func (hs *HTTPServer) ServeStaticFile(relativePath, filepath string) {
	hs.router.StaticFile(relativePath, filepath)
}

// SetNoRoute sets the handler for unmatched routes
func (hs *HTTPServer) SetNoRoute(handler gin.HandlerFunc) {
	hs.router.NoRoute(handler)
}

// SetNoMethod sets the handler for unmatched methods
func (hs *HTTPServer) SetNoMethod(handler gin.HandlerFunc) {
	hs.router.NoMethod(handler)
}

// GetServer returns the underlying HTTP server
func (hs *HTTPServer) GetServer() *http.Server {
	return hs.server
}

// IsRunning returns whether the server is running
func (hs *HTTPServer) IsRunning() bool {
	return hs.server != nil
}

// GetListenAddress returns the listen address
func (hs *HTTPServer) GetListenAddress() string {
	return hs.config.Listen
}

// IsTLSEnabled returns whether TLS is enabled
func (hs *HTTPServer) IsTLSEnabled() bool {
	return hs.config.TLSEnabled
}

// UpdateConfig updates the server configuration
func (hs *HTTPServer) UpdateConfig(config *HTTPServerConfig) error {
	if hs.IsRunning() {
		return fmt.Errorf("cannot update config while server is running")
	}
	
	hs.config = config
	return nil
}

// GetConfig returns the server configuration
func (hs *HTTPServer) GetConfig() *HTTPServerConfig {
	return hs.config
}

// Health returns the health status of the HTTP server
func (hs *HTTPServer) Health() map[string]interface{} {
	hs.metrics.mu.RLock()
	defer hs.metrics.mu.RUnlock()
	
	return map[string]interface{}{
		"status":             "healthy",
		"listen_address":     hs.config.Listen,
		"tls_enabled":        hs.config.TLSEnabled,
		"active_connections": hs.metrics.ConnectionsActive,
		"total_requests":     hs.metrics.RequestsTotal,
		"success_rate":       float64(hs.metrics.RequestsSuccess) / float64(hs.metrics.RequestsTotal),
		"average_response_time": hs.metrics.AverageResponseTime.String(),
		"last_updated":       hs.metrics.LastUpdated,
	}
}

// Reset resets the HTTP server metrics
func (hs *HTTPServer) Reset() {
	hs.metrics.mu.Lock()
	defer hs.metrics.mu.Unlock()
	
	hs.metrics.RequestsTotal = 0
	hs.metrics.RequestsSuccess = 0
	hs.metrics.RequestsError = 0
	hs.metrics.ConnectionsTotal = 0
	hs.metrics.AverageResponseTime = 0
	hs.metrics.LastUpdated = time.Now()
}

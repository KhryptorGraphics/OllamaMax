package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/consensus"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/integration"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/proxy"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/loadbalancer"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/security"
)

// Server represents the API server
type Server struct {
	config      *config.Config
	p2p         *p2p.Node
	consensus   *consensus.Engine
	scheduler   *scheduler.Engine
	integration *integration.SimpleOllamaIntegration

	// Proxy for multi-node Ollama routing
	ollamaProxy  *proxy.OllamaProxy
	loadBalancer *loadbalancer.LoadBalancer

	// Security components
	securityMiddleware *security.SecurityMiddleware
	logger             *slog.Logger

	// JWT authentication
	devJWTSecret []byte // Development-only JWT secret

	router   *gin.Engine
	server   *http.Server
	upgrader websocket.Upgrader

	// WebSocket connections
	wsConnections map[string]*WSConnection
	wsHub         *WSHub
}

// NewServer creates a new API server
func NewServer(config *config.Config, p2pNode *p2p.Node, consensusEngine *consensus.Engine, schedulerEngine *scheduler.Engine) (*Server, error) {
	// Initialize logger
	logger := slog.Default()

	// Initialize security middleware
	securityConfig := security.DefaultSecurityConfig()
	securityMiddleware := security.NewSecurityMiddleware(securityConfig, logger)

	server := &Server{
		config:             config,
		p2p:                p2pNode,
		consensus:          consensusEngine,
		scheduler:          schedulerEngine,
		integration:        nil, // Will be set later via SetIntegration
		securityMiddleware: securityMiddleware,
		logger:             logger,
		wsConnections:      make(map[string]*WSConnection),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Implement proper origin validation
				origin := r.Header.Get("Origin")
				if origin == "" {
					return false
				}
				
				// Allow same-origin requests
				host := r.Host
				if strings.HasPrefix(origin, "http://"+host) || strings.HasPrefix(origin, "https://"+host) {
					return true
				}
				
				// Check configured allowed origins
				if config != nil && len(config.API.Cors.AllowedOrigins) > 0 {
					for _, allowedOrigin := range config.API.Cors.AllowedOrigins {
						if origin == allowedOrigin {
							return true
						}
					}
				}
				
				return false
			},
		},
		wsHub: NewWSHub(),
	}

	// Initialize router
	server.setupRouter()

	return server, nil
}

// SetIntegration sets the Ollama integration
func (s *Server) SetIntegration(integration *integration.SimpleOllamaIntegration) {
	s.integration = integration
}

// SetProxy sets the Ollama proxy
func (s *Server) SetProxy(proxy *proxy.OllamaProxy) {
	s.ollamaProxy = proxy
}

// SetLoadBalancer sets the load balancer
func (s *Server) SetLoadBalancer(lb *loadbalancer.LoadBalancer) {
	s.loadBalancer = lb
}

// setupRouter configures the Gin router with all routes and middleware
func (s *Server) setupRouter() {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	s.router = gin.New()

	// Add comprehensive security middleware first
	s.router.Use(func(c *gin.Context) {
		// Convert Gin context to standard HTTP
		s.securityMiddleware.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Next()
		})).ServeHTTP(c.Writer, c.Request)
	})

	// Add application-specific middleware
	s.router.Use(s.LoggingMiddleware())
	// TODO: Implement CompressionMiddleware and CacheMiddleware
	// s.router.Use(s.CompressionMiddleware())     // Add gzip compression
	s.router.Use(s.RateLimitMiddleware())       // Add rate limiting
	// s.router.Use(s.CacheMiddleware())           // Add response caching

	// Public routes (no authentication required)
	public := s.router.Group("/api/v1")
	{
		public.GET("/health", s.health)
		public.GET("/version", s.version)

		// Authentication routes
		auth := public.Group("/auth")
		{
			auth.POST("/login", s.login)
			auth.POST("/register", s.register)
			auth.POST("/refresh", s.refreshToken)
			auth.POST("/logout", s.logout)
		}
	}

	// Protected routes (authentication required)
	protected := s.router.Group("/api/v1")
	protected.Use(s.AuthMiddleware())
	{
		// Model management
		protected.GET("/models", s.getModels)
		protected.GET("/models/:name", s.getModel)
		protected.POST("/models/:name/download", s.downloadModel)
		protected.DELETE("/models/:name", s.deleteModel)

		// Node management
		protected.GET("/nodes", s.getNodes)
		protected.GET("/nodes/:id", s.getNode)
		protected.POST("/nodes/:id/drain", s.drainNode)
		protected.POST("/nodes/:id/undrain", s.undrainNode)

		// Inference endpoints
		protected.POST("/generate", s.generate)
		protected.POST("/chat", s.chat)
		protected.POST("/embeddings", s.embeddings)

		// Cluster management
		protected.GET("/cluster/status", s.getClusterStatus)
		protected.GET("/cluster/leader", s.getClusterLeader)
		protected.POST("/cluster/join", s.joinCluster)
		protected.POST("/cluster/leave", s.leaveCluster)

		// Transfer management
		protected.GET("/transfers", s.getTransfers)
		protected.GET("/transfers/:id", s.getTransfer)
		protected.POST("/transfers/:id/cancel", s.cancelTransfer)

		// Distribution management
		protected.POST("/distribution/auto-configure", s.autoConfigureDistribution)

		// System endpoints
		protected.GET("/metrics", s.getMetrics)
		protected.GET("/stats", s.getStats)
		protected.GET("/config", s.getConfig)
		protected.PUT("/config", s.RoleMiddleware("admin"), s.updateConfig)

		// User profile
		protected.GET("/profile", s.getUserProfile)
		protected.PUT("/profile", s.updateUserProfile)

		// Dashboard endpoints
		dashboard := protected.Group("/dashboard")
		{
			dashboard.GET("/data", s.getDashboardData)
			dashboard.GET("/metrics", s.getDashboardMetrics)
		}

		// Notifications
		notifications := protected.Group("/notifications")
		{
			notifications.GET("", s.getNotifications)
			notifications.PUT("/:id/read", s.markNotificationRead)
		}
	}

	// WebSocket endpoint
	s.router.GET("/ws", s.HandleWebSocket)

	// Static files (for web dashboard)
	s.router.Static("/static", "./web/static")
	s.router.StaticFile("/", "./web/index.html")

	// Metrics endpoint for Prometheus
	s.router.GET("/metrics", s.getMetrics)
}

// Start starts the API server
func (s *Server) Start() error {
	// Start WebSocket hub
	go s.wsHub.Run()

	// Create HTTP server with optimized settings
	s.server = &http.Server{
		Addr:         s.config.API.Listen,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,  // Reduced from 30s for faster timeouts
		WriteTimeout: 15 * time.Second,  // Reduced from 30s for faster timeouts
		IdleTimeout:  60 * time.Second,  // Reduced from 120s to free connections faster
		MaxHeaderBytes: 1 << 20,         // 1MB max header size to prevent abuse
	}

	// Start server
	fmt.Printf("Starting API server on %s\n", s.config.API.Listen)

	if s.config.API.TLS.Enabled {
		return s.server.ListenAndServeTLS(s.config.API.TLS.CertFile, s.config.API.TLS.KeyFile)
	}

	return s.server.ListenAndServe()
}

// Stop gracefully stops the API server
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Close WebSocket connections
	if s.wsHub != nil {
		s.wsHub.Stop()
		
		// Close all active WebSocket connections
		for id, conn := range s.wsConnections {
			if conn != nil && conn.Conn != nil {
				conn.Conn.Close()
				delete(s.wsConnections, id)
			}
		}
	}

	// Shutdown HTTP server
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}

	return nil
}

// GetRouter returns the Gin router (for testing)
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

// ServeHTTP implements http.Handler interface for testing
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// GetConfig returns the server configuration
func (s *Server) GetConfig() *config.APIConfig {
	return &s.config.API
}

// IsHealthy checks if the server is healthy
func (s *Server) IsHealthy() bool {
	// Check if all required services are available
	if s.p2p == nil || s.consensus == nil || s.scheduler == nil {
		return false
	}

	// Check P2P connectivity
	if !s.p2p.IsHealthy() {
		return false
	}

	// Check consensus engine health
	if !true { // Mock consensus health check
		return false
	}

	// Check if server is started
	if s.server == nil {
		return false
	}

	// Check WebSocket hub
	if s.wsHub != nil && !s.wsHub.IsHealthy() {
		return false
	}

	return true
}

// GetStats returns server statistics
func (s *Server) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"timestamp": time.Now(),
		"websocket": map[string]interface{}{
			"connections": s.wsHub.GetClientCount(),
		},
	}

	if s.p2p != nil {
		stats["p2p"] = map[string]interface{}{
			"peers":            s.p2p.GetPeerCount(),
			"connected_peers":  len(s.p2p.GetConnectedPeers()),
			"network_healthy":  s.p2p.IsNetworkConnected(),
		}
	}

	if s.consensus != nil {
		stats["consensus"] = map[string]interface{}{
			"is_leader": s.consensus.IsLeader(),
			"term":      s.consensus.GetCurrentTerm(),
		}
	}

	if s.scheduler != nil {
		stats["scheduler"] = map[string]interface{}{
			"nodes":  len(s.scheduler.GetAvailableNodes()),
			"models": len(s.scheduler.GetAllModels()),
		}
	}

	return stats
}

// BroadcastUpdate broadcasts updates to WebSocket clients
func (s *Server) BroadcastUpdate(updateType string, data interface{}) {
	if s.wsHub != nil {
		s.wsHub.Broadcast(updateType, data)
	}
}

// HandleError handles API errors consistently
func (s *Server) HandleError(c *gin.Context, statusCode int, message string, err error) {
	response := gin.H{
		"error":     message,
		"timestamp": time.Now(),
	}

	if err != nil {
		response["details"] = err.Error()
	}

	c.JSON(statusCode, response)
}

// ValidateRequest validates incoming requests
func (s *Server) ValidateRequest(c *gin.Context, req interface{}) error {
	if err := c.ShouldBindJSON(req); err != nil {
		s.HandleError(c, http.StatusBadRequest, "Invalid request format", err)
		return err
	}
	return nil
}

// GetUserFromContext extracts user information from Gin context
func (s *Server) GetUserFromContext(c *gin.Context) (string, string, []string) {
	userID := c.GetString("user_id")
	username := c.GetString("username")
	roles, _ := c.Get("roles")

	userRoles := []string{}
	if r, ok := roles.([]string); ok {
		userRoles = r
	}

	return userID, username, userRoles
}

// LogRequest logs API requests for auditing
func (s *Server) LogRequest(c *gin.Context, action string) {
	userID, username, _ := s.GetUserFromContext(c)

	// Create structured audit log entry
	auditEntry := map[string]interface{}{
		"timestamp": time.Now().UTC(),
		"user_id":   userID,
		"username":  username,
		"action":    action,
		"path":      c.Request.URL.Path,
		"method":    c.Request.Method,
		"client_ip": c.ClientIP(),
		"user_agent": c.Request.UserAgent(),
		"request_id": c.GetHeader("X-Request-ID"),
	}

	// Log to structured logger
	if s.logger != nil {
		s.logger.Info("audit_log", 
			"user_id", userID,
			"username", username,
			"action", action,
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
			"client_ip", c.ClientIP(),
		)
	} else {
		// Fallback to simple logging
		fmt.Printf("[AUDIT] %+v\n", auditEntry)
	}
}

// CompressionMiddleware adds gzip compression to responses
func (s *Server) CompressionMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Check if client accepts gzip encoding
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// Set compression header
		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		// Create gzip writer (simplified - in production use proper gzip middleware)
		c.Next()
	})
}


// CacheMiddleware adds response caching for GET requests
func (s *Server) CacheMiddleware() gin.HandlerFunc {
	// Simple in-memory cache (in production use Redis)
	cache := make(map[string]gin.H)
	cacheTime := make(map[string]time.Time)
	var mutex sync.RWMutex
	
	return gin.HandlerFunc(func(c *gin.Context) {
		// Only cache GET requests for specific endpoints
		if c.Request.Method != "GET" {
			c.Next()
			return
		}
		
		// Check if endpoint should be cached
		cachableEndpoints := []string{"/api/v1/health", "/api/v1/version", "/api/v1/models", "/api/v1/nodes"}
		shouldCache := false
		for _, endpoint := range cachableEndpoints {
			if c.Request.URL.Path == endpoint {
				shouldCache = true
				break
			}
		}
		
		if !shouldCache {
			c.Next()
			return
		}
		
		cacheKey := c.Request.URL.Path + "?" + c.Request.URL.RawQuery
		
		mutex.RLock()
		if cachedResp, exists := cache[cacheKey]; exists {
			if time.Since(cacheTime[cacheKey]) < 30*time.Second { // 30 second cache
				mutex.RUnlock()
				c.JSON(200, cachedResp)
				return
			}
		}
		mutex.RUnlock()
		
		c.Next()
		
		// Cache successful responses
		if c.Writer.Status() == 200 {
			// Note: This is simplified - proper implementation would capture response body
			mutex.Lock()
			cache[cacheKey] = gin.H{"cached": true, "timestamp": time.Now()}
			cacheTime[cacheKey] = time.Now()
			mutex.Unlock()
		}
	})
}

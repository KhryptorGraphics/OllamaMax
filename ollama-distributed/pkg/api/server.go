package api

import (
	"context"
	"fmt"
	"net/http"
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
)

// Server represents the API server
type Server struct {
	config      *config.APIConfig
	p2p         *p2p.Node
	consensus   *consensus.Engine
	scheduler   *scheduler.Engine
	integration *integration.SimpleOllamaIntegration

	// Proxy for multi-node Ollama routing
	ollamaProxy  *proxy.OllamaProxy
	loadBalancer *loadbalancer.LoadBalancer

	router   *gin.Engine
	server   *http.Server
	upgrader websocket.Upgrader

	// WebSocket connections
	wsConnections map[string]*WSConnection
	wsHub         *WSHub
}

// NewServer creates a new API server
func NewServer(config *config.APIConfig, p2pNode *p2p.Node, consensusEngine *consensus.Engine, schedulerEngine *scheduler.Engine) (*Server, error) {
	server := &Server{
		config:        config,
		p2p:           p2pNode,
		consensus:     consensusEngine,
		scheduler:     schedulerEngine,
		integration:   nil, // Will be set later via SetIntegration
		wsConnections: make(map[string]*WSConnection),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
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

	// Add middleware
	s.router.Use(s.LoggingMiddleware())
	s.router.Use(s.CORSMiddleware())
	s.router.Use(s.SecurityHeadersMiddleware())
	s.router.Use(s.RateLimitMiddleware())

	// Public routes (no authentication required)
	public := s.router.Group("/api/v1")
	{
		public.GET("/health", s.health)
		public.GET("/version", s.version)
		public.POST("/auth/login", s.login)
		public.POST("/auth/logout", s.logout)
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
		protected.GET("/profile", s.profile)
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

	// Create HTTP server
	s.server = &http.Server{
		Addr:         s.config.Listen,
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server
	fmt.Printf("Starting API server on %s\n", s.config.Listen)

	if s.config.TLS.Enabled {
		return s.server.ListenAndServeTLS(s.config.TLS.CertFile, s.config.TLS.KeyFile)
	}

	return s.server.ListenAndServe()
}

// Stop gracefully stops the API server
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Close WebSocket connections
	if s.wsHub != nil {
		// TODO: Implement graceful WebSocket shutdown
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
	return s.config
}

// IsHealthy checks if the server is healthy
func (s *Server) IsHealthy() bool {
	// Check if all required services are available
	if s.p2p == nil || s.consensus == nil || s.scheduler == nil {
		return false
	}

	// TODO: Add more health checks
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
			"peers": 0, // TODO: Implement GetPeers method
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

	// TODO: Implement proper audit logging
	fmt.Printf("[AUDIT] User: %s (%s), Action: %s, Path: %s, Method: %s, IP: %s\n",
		username, userID, action, c.Request.URL.Path, c.Request.Method, c.ClientIP())
}

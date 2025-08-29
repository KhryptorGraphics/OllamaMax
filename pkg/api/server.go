package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/khryptorgraphics/ollamamax/internal/config"
	"github.com/khryptorgraphics/ollamamax/pkg/auth"
	"github.com/khryptorgraphics/ollamamax/pkg/database"
)

// Server represents the API server
type Server struct {
	config   *config.Config
	db       *database.DatabaseManager
	jwtSvc   *auth.JWTService
	logger   *slog.Logger
	server   *http.Server
	websocket *WebSocketHub
}

// NewServer creates a new API server instance
func NewServer(cfg *config.Config, db *database.DatabaseManager, logger *slog.Logger) (*Server, error) {
	// Initialize JWT service
	jwtSvc, err := auth.NewJWTService(&cfg.Auth)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT service: %w", err)
	}

	// Initialize WebSocket hub
	websocketHub := NewWebSocketHub(logger)

	server := &Server{
		config:    cfg,
		db:        db,
		jwtSvc:    jwtSvc,
		logger:    logger,
		websocket: websocketHub,
	}

	return server, nil
}

// Start starts the API server
func (s *Server) Start(ctx context.Context) error {
	// Create Gin router
	router := s.setupRouter()

	// Create HTTP server
	s.server = &http.Server{
		Addr:         s.config.API.Listen,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start WebSocket hub
	go s.websocket.Run()

	s.logger.Info("Starting API server",
		"address", s.config.API.Listen,
		"tls_enabled", s.config.API.TLSEnabled)

	// Start server
	if s.config.API.TLSEnabled {
		return s.server.ListenAndServeTLS(s.config.API.CertFile, s.config.API.KeyFile)
	}
	return s.server.ListenAndServe()
}

// Stop gracefully stops the API server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping API server")

	// Stop WebSocket hub
	s.websocket.Stop()

	// Shutdown HTTP server
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// setupRouter configures the Gin router with middleware and routes
func (s *Server) setupRouter() *gin.Engine {
	// Set Gin mode based on environment
	if s.logger.Enabled(context.Background(), slog.LevelDebug) {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(s.loggingMiddleware())
	router.Use(gin.Recovery())
	router.Use(s.corsMiddleware())
	router.Use(s.securityMiddleware())

	// Rate limiting middleware
	if s.config.API.RateLimit.Enabled {
		router.Use(s.rateLimitMiddleware())
	}

	// Health check endpoint (no auth required)
	router.GET("/health", s.healthHandler)
	router.GET("/metrics", s.metricsHandler)

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Public endpoints
		auth := v1.Group("/auth")
		{
			auth.POST("/login", s.loginHandler)
			auth.POST("/register", s.registerHandler)
			auth.POST("/refresh", s.refreshTokenHandler)
		}

		// Protected endpoints (require authentication)
		protected := v1.Group("/")
		protected.Use(auth.JWTAuthMiddleware(s.jwtSvc))
		{
			// User management
			users := protected.Group("/users")
			{
				users.GET("/profile", s.getUserProfileHandler)
				users.PUT("/profile", s.updateUserProfileHandler)
				users.POST("/logout", s.logoutHandler)
			}

			// Model management
			models := protected.Group("/models")
			{
				models.GET("/", s.listModelsHandler)
				models.POST("/", s.createModelHandler)
				models.GET("/:id", s.getModelHandler)
				models.PUT("/:id", s.updateModelHandler)
				models.DELETE("/:id", s.deleteModelHandler)
				models.GET("/:id/replicas", s.getModelReplicasHandler)
			}

			// Node management
			nodes := protected.Group("/nodes")
			{
				nodes.GET("/", s.listNodesHandler)
				nodes.GET("/:id", s.getNodeHandler)
				nodes.PUT("/:id", s.updateNodeHandler)
				nodes.DELETE("/:id", s.deleteNodeHandler)
				nodes.GET("/:id/health", s.getNodeHealthHandler)
			}

			// Inference endpoints
			inference := protected.Group("/inference")
			{
				inference.POST("/chat", s.chatHandler)
				inference.POST("/generate", s.generateHandler)
				inference.GET("/requests", s.listInferenceRequestsHandler)
				inference.GET("/requests/:id", s.getInferenceRequestHandler)
			}

			// System configuration
			system := protected.Group("/system")
			{
				system.GET("/config", s.getSystemConfigHandler)
				system.PUT("/config", s.updateSystemConfigHandler)
				system.GET("/stats", s.getSystemStatsHandler)
				system.GET("/audit", s.getAuditLogsHandler)
			}
		}
	}

	// WebSocket endpoints
	router.GET("/ws", s.websocketHandler)
	router.GET("/ws/inference/:id", s.inferenceWebsocketHandler)

	// Static file serving (for admin dashboard)
	router.Static("/static", "./web/dist")
	router.StaticFile("/", "./web/dist/index.html")

	return router
}

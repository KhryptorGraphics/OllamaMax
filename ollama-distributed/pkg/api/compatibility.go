package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ollama/ollama-distributed/pkg/integration"
	"github.com/ollama/ollama-distributed/pkg/models"
	"github.com/ollama/ollama-distributed/pkg/scheduler"
)

// DistributedServerWrapper wraps the original Ollama server with distributed capabilities
type DistributedServerWrapper struct {
	// Original server components (using interface for compatibility)
	originalServer interface{}
	
	// Distributed components
	distributedRoutes *DistributedRoutes
	scheduler         *scheduler.Engine
	modelDistribution *models.Distribution
	
	// Configuration
	distributedEnabled bool
	fallbackEnabled    bool
	localAddr          string
}

// NewDistributedServerWrapper creates a new wrapper for the original server
func NewDistributedServerWrapper(originalServer interface{}, scheduler *scheduler.Engine, modelDist *models.Distribution, localAddr string) (*DistributedServerWrapper, error) {
	// Create distributed routes
	distributedRoutes, err := NewDistributedRoutes(scheduler, modelDist, localAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create distributed routes: %w", err)
	}
	
	// Set original server for fallback
	distributedRoutes.SetOriginalServer(originalServer)
	
	return &DistributedServerWrapper{
		originalServer:     originalServer,
		distributedRoutes:  distributedRoutes,
		scheduler:          scheduler,
		modelDistribution:  modelDist,
		distributedEnabled: true,
		fallbackEnabled:    true,
		localAddr:          localAddr,
	}, nil
}

// GenerateRoutesWithDistributed generates routes with distributed capabilities
func (dsw *DistributedServerWrapper) GenerateRoutesWithDistributed(rc *integration.Registry) (http.Handler, error) {
	// Create a stub handler for original routes (would integrate with actual ollama server)
	originalHandler := http.NewServeMux()
	originalHandler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Original ollama server integration not available in stub mode"))
	})
	
	// Create a new router with distributed capabilities
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	
	// Add distributed middleware
	router.Use(dsw.distributedMiddleware())
	
	// Setup distributed routes
	dsw.distributedRoutes.SetupRoutes(router)
	
	// Add fallback handler for any routes not handled by distributed system
	router.NoRoute(func(c *gin.Context) {
		// If distributed mode is disabled or fallback is needed, use original handler
		if !dsw.distributedEnabled || dsw.shouldFallback(c) {
			// Serve using original handler
			originalHandler.ServeHTTP(c.Writer, c.Request)
		} else {
			// Return 404 for unknown distributed routes
			c.JSON(http.StatusNotFound, gin.H{"error": "endpoint not found"})
		}
	})
	
	return router, nil
}

// distributedMiddleware adds distributed-specific middleware
func (dsw *DistributedServerWrapper) distributedMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add distributed headers
		c.Header("X-Ollama-Distributed-Wrapper", "true")
		c.Header("X-Ollama-Distributed-Enabled", fmt.Sprintf("%v", dsw.distributedEnabled))
		c.Header("X-Ollama-Fallback-Enabled", fmt.Sprintf("%v", dsw.fallbackEnabled))
		
		// Check if we should enable distributed mode for this request
		if dsw.shouldUseDistributed(c) {
			c.Header("X-Ollama-Route-Mode", "distributed")
		} else {
			c.Header("X-Ollama-Route-Mode", "local")
		}
		
		c.Next()
	}
}

// shouldUseDistributed determines if a request should use distributed routing
func (dsw *DistributedServerWrapper) shouldUseDistributed(c *gin.Context) bool {
	// Check if distributed mode is enabled
	if !dsw.distributedEnabled {
		return false
	}
	
	// Check for override headers
	if c.GetHeader("X-Ollama-Force-Local") == "true" {
		return false
	}
	
	if c.GetHeader("X-Ollama-Force-Distributed") == "true" {
		return true
	}
	
	// Check request path
	path := c.Request.URL.Path
	
	// Always use distributed for inference endpoints
	if strings.HasPrefix(path, "/api/generate") ||
		strings.HasPrefix(path, "/api/chat") ||
		strings.HasPrefix(path, "/api/embed") ||
		strings.HasPrefix(path, "/v1/chat/completions") ||
		strings.HasPrefix(path, "/v1/completions") ||
		strings.HasPrefix(path, "/v1/embeddings") {
		return true
	}
	
	// Use distributed for model management if model is distributed
	if strings.HasPrefix(path, "/api/show") ||
		strings.HasPrefix(path, "/api/tags") ||
		strings.HasPrefix(path, "/api/pull") {
		return true
	}
	
	// Default to local for other endpoints
	return false
}

// shouldFallback determines if a request should fallback to local
func (dsw *DistributedServerWrapper) shouldFallback(c *gin.Context) bool {
	// Check if fallback is enabled
	if !dsw.fallbackEnabled {
		return false
	}
	
	// Check for fallback indicators in headers
	if c.GetHeader("X-Ollama-Distributed-Error") != "" {
		return true
	}
	
	// Check cluster health
	if dsw.scheduler.GetClusterSize() < 2 {
		return true
	}
	
	// Check if any nodes are available
	if dsw.scheduler.GetActiveNodes() == 0 {
		return true
	}
	
	return false
}

// Start starts the distributed server wrapper
func (dsw *DistributedServerWrapper) Start(ctx context.Context) error {
	slog.Info("Starting distributed server wrapper")
	
	// Start distributed routes
	if err := dsw.distributedRoutes.Start(ctx); err != nil {
		return fmt.Errorf("failed to start distributed routes: %w", err)
	}
	
	return nil
}

// Shutdown gracefully shuts down the distributed server wrapper
func (dsw *DistributedServerWrapper) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down distributed server wrapper")
	
	// Shutdown distributed routes
	if err := dsw.distributedRoutes.Shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown distributed routes", "error", err)
		return err
	}
	
	return nil
}

// GetDistributedRoutes returns the distributed routes handler
func (dsw *DistributedServerWrapper) GetDistributedRoutes() *DistributedRoutes {
	return dsw.distributedRoutes
}

// SetDistributedEnabled enables or disables distributed mode
func (dsw *DistributedServerWrapper) SetDistributedEnabled(enabled bool) {
	dsw.distributedEnabled = enabled
}

// SetFallbackEnabled enables or disables fallback mode
func (dsw *DistributedServerWrapper) SetFallbackEnabled(enabled bool) {
	dsw.fallbackEnabled = enabled
}

// IsDistributedEnabled returns whether distributed mode is enabled
func (dsw *DistributedServerWrapper) IsDistributedEnabled() bool {
	return dsw.distributedEnabled
}

// IsFallbackEnabled returns whether fallback mode is enabled
func (dsw *DistributedServerWrapper) IsFallbackEnabled() bool {
	return dsw.fallbackEnabled
}

// GetStats returns wrapper statistics
func (dsw *DistributedServerWrapper) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"distributed_enabled": dsw.distributedEnabled,
		"fallback_enabled":    dsw.fallbackEnabled,
		"local_addr":          dsw.localAddr,
		"cluster_size":        dsw.scheduler.GetClusterSize(),
		"active_nodes":        dsw.scheduler.GetActiveNodes(),
		"distributed_routes":  dsw.distributedRoutes.GetIntegrationLayer().GetStats(),
	}
}

// DistributedServerCompatibility provides compatibility functions for integration
type DistributedServerCompatibility struct {
	wrapper *DistributedServerWrapper
}

// NewDistributedServerCompatibility creates a new compatibility layer
func NewDistributedServerCompatibility(wrapper *DistributedServerWrapper) *DistributedServerCompatibility {
	return &DistributedServerCompatibility{
		wrapper: wrapper,
	}
}

// WrapGenerateRoutes wraps the original GenerateRoutes function with distributed capabilities
func (dsc *DistributedServerCompatibility) WrapGenerateRoutes(originalServer interface{}) func(*integration.Registry) (http.Handler, error) {
	return func(rc *integration.Registry) (http.Handler, error) {
		// Check if distributed mode is enabled
		if os.Getenv("OLLAMA_DISTRIBUTED") == "false" {
			// Return stub handler for original server
			handler := http.NewServeMux()
			handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Original ollama server (stub mode)"))
			})
			return handler, nil
		}
		
		// Use distributed wrapper
		return dsc.wrapper.GenerateRoutesWithDistributed(rc)
	}
}

// GetDistributedHandlers returns distributed-specific handlers
func (dsc *DistributedServerCompatibility) GetDistributedHandlers() map[string]gin.HandlerFunc {
	distributedRoutes := dsc.wrapper.GetDistributedRoutes()
	
	return map[string]gin.HandlerFunc{
		"generate":    distributedRoutes.handleGenerate,
		"chat":        distributedRoutes.handleChat,
		"embed":       distributedRoutes.handleEmbed,
		"embeddings":  distributedRoutes.handleEmbeddings,
		"pull":        distributedRoutes.handlePull,
		"show":        distributedRoutes.handleShow,
		"tags":        distributedRoutes.handleTags,
		"ps":          distributedRoutes.handlePs,
		"version":     distributedRoutes.handleVersion,
		"health":      distributedRoutes.handleHealth,
		"metrics":     distributedRoutes.handleMetrics,
	}
}

// InjectDistributedHandlers injects distributed handlers into existing routes
func (dsc *DistributedServerCompatibility) InjectDistributedHandlers(router *gin.Engine) {
	handlers := dsc.GetDistributedHandlers()
	
	// Replace existing handlers with distributed versions
	api := router.Group("/api")
	{
		api.POST("/generate", handlers["generate"])
		api.POST("/chat", handlers["chat"])
		api.POST("/embed", handlers["embed"])
		api.POST("/embeddings", handlers["embeddings"])
		api.POST("/pull", handlers["pull"])
		api.POST("/show", handlers["show"])
		api.GET("/tags", handlers["tags"])
		api.GET("/ps", handlers["ps"])
		api.GET("/version", handlers["version"])
	}
	
	// Add health and metrics
	router.GET("/health", handlers["health"])
	router.GET("/metrics", handlers["metrics"])
}

// CreateDistributedMiddleware creates middleware for distributed functionality
func (dsc *DistributedServerCompatibility) CreateDistributedMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request should be handled by distributed system
		if dsc.wrapper.shouldUseDistributed(c) {
			// Mark as distributed
			c.Set("distributed", true)
			c.Header("X-Ollama-Distributed-Request", "true")
			
			// Check if fallback is needed
			if dsc.wrapper.shouldFallback(c) {
				c.Header("X-Ollama-Fallback-Available", "true")
			}
		} else {
			// Mark as local
			c.Set("distributed", false)
			c.Header("X-Ollama-Distributed-Request", "false")
		}
		
		c.Next()
	}
}

// GetDistributedStatus returns the status of distributed components
func (dsc *DistributedServerCompatibility) GetDistributedStatus() map[string]interface{} {
	return map[string]interface{}{
		"wrapper_enabled":     dsc.wrapper.IsDistributedEnabled(),
		"fallback_enabled":    dsc.wrapper.IsFallbackEnabled(),
		"cluster_size":        dsc.wrapper.scheduler.GetClusterSize(),
		"active_nodes":        dsc.wrapper.scheduler.GetActiveNodes(),
		"scheduler_healthy":   dsc.wrapper.scheduler.IsHealthy(),
		"routes_initialized":  dsc.wrapper.distributedRoutes != nil,
		"integration_stats":   dsc.wrapper.distributedRoutes.GetIntegrationLayer().GetStats(),
	}
}

// EnableDistributedMode enables distributed mode with configuration
func (dsc *DistributedServerCompatibility) EnableDistributedMode(config map[string]interface{}) error {
	// Enable distributed mode
	dsc.wrapper.SetDistributedEnabled(true)
	
	// Configure based on provided config
	if fallback, ok := config["fallback"].(bool); ok {
		dsc.wrapper.SetFallbackEnabled(fallback)
	}
	
	// Start distributed components if not already started
	ctx := context.Background()
	if err := dsc.wrapper.Start(ctx); err != nil {
		return fmt.Errorf("failed to start distributed mode: %w", err)
	}
	
	slog.Info("Distributed mode enabled", "config", config)
	return nil
}

// DisableDistributedMode disables distributed mode
func (dsc *DistributedServerCompatibility) DisableDistributedMode() error {
	// Disable distributed mode
	dsc.wrapper.SetDistributedEnabled(false)
	
	// Shutdown distributed components
	ctx := context.Background()
	if err := dsc.wrapper.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown distributed mode: %w", err)
	}
	
	slog.Info("Distributed mode disabled")
	return nil
}

// GetIntegrationLayer returns the integration layer for external use
func (dsc *DistributedServerCompatibility) GetIntegrationLayer() *IntegrationLayer {
	return dsc.wrapper.GetDistributedRoutes().GetIntegrationLayer()
}

// GetDistributedRunner returns the distributed runner for external use
func (dsc *DistributedServerCompatibility) GetDistributedRunner() *DistributedRunner {
	return dsc.wrapper.GetDistributedRoutes().GetDistributedRunner()
}

// RegisterDistributedEndpoints registers additional distributed endpoints
func (dsc *DistributedServerCompatibility) RegisterDistributedEndpoints(router *gin.Engine) {
	// Distributed status endpoint
	router.GET("/distributed/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, dsc.GetDistributedStatus())
	})
	
	// Distributed control endpoints
	admin := router.Group("/distributed/admin")
	{
		admin.POST("/enable", func(c *gin.Context) {
			var config map[string]interface{}
			if err := c.ShouldBindJSON(&config); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			
			if err := dsc.EnableDistributedMode(config); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			
			c.JSON(http.StatusOK, gin.H{"message": "Distributed mode enabled"})
		})
		
		admin.POST("/disable", func(c *gin.Context) {
			if err := dsc.DisableDistributedMode(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			
			c.JSON(http.StatusOK, gin.H{"message": "Distributed mode disabled"})
		})
	}
}
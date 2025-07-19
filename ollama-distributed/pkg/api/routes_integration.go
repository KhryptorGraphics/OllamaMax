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

// RouteIntegration provides seamless integration with existing Ollama routes
type RouteIntegration struct {
	distributedWrapper *DistributedServerWrapper
	compatibility      *DistributedServerCompatibility
	fallbackManager    *FallbackManager
	standalone         *StandaloneMode
	
	// Integration state
	enabled     bool
	initialized bool
}

// NewRouteIntegration creates a new route integration
func NewRouteIntegration(scheduler *scheduler.Engine, modelDist *models.Distribution, localAddr string) (*RouteIntegration, error) {
	// Create fallback manager
	fallbackMgr, err := NewFallbackManager(localAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create fallback manager: %w", err)
	}
	
	// Create standalone mode
	standalone := NewStandaloneMode(fallbackMgr)
	
	return &RouteIntegration{
		fallbackManager: fallbackMgr,
		standalone:      standalone,
		enabled:         os.Getenv("OLLAMA_DISTRIBUTED") != "false",
		initialized:     false,
	}, nil
}

// Initialize initializes the route integration with an existing server
func (ri *RouteIntegration) Initialize(originalServer interface{}, scheduler *scheduler.Engine, modelDist *models.Distribution, localAddr string) error {
	if ri.initialized {
		return nil
	}
	
	// Create distributed wrapper
	wrapper, err := NewDistributedServerWrapper(originalServer, scheduler, modelDist, localAddr)
	if err != nil {
		return fmt.Errorf("failed to create distributed wrapper: %w", err)
	}
	
	// Create compatibility layer
	compatibility := NewDistributedServerCompatibility(wrapper)
	
	ri.distributedWrapper = wrapper
	ri.compatibility = compatibility
	ri.initialized = true
	
	slog.Info("Route integration initialized", "enabled", ri.enabled)
	return nil
}

// WrapGenerateRoutes wraps the original GenerateRoutes function
func (ri *RouteIntegration) WrapGenerateRoutes(originalServer interface{}, scheduler *scheduler.Engine, modelDist *models.Distribution, localAddr string) func(*integration.Registry) (http.Handler, error) {
	return func(rc *integration.Registry) (http.Handler, error) {
		// Initialize if not done
		if !ri.initialized {
			if err := ri.Initialize(originalServer, scheduler, modelDist, localAddr); err != nil {
				slog.Error("Failed to initialize route integration", "error", err)
				// Fall back to original
				return originalServer.GenerateRoutes(rc)
			}
		}
		
		// Check if distributed mode is enabled
		if !ri.enabled {
			slog.Info("Distributed mode disabled, using original routes")
			return originalServer.GenerateRoutes(rc)
		}
		
		// Check if we should use standalone mode
		if ri.shouldUseStandaloneMode() {
			slog.Info("Using standalone mode")
			return ri.generateStandaloneRoutes(rc)
		}
		
		// Use distributed routes
		slog.Info("Using distributed routes")
		return ri.distributedWrapper.GenerateRoutesWithDistributed(rc)
	}
}

// shouldUseStandaloneMode determines if standalone mode should be used
func (ri *RouteIntegration) shouldUseStandaloneMode() bool {
	// Check if explicitly enabled
	if os.Getenv("OLLAMA_STANDALONE") == "true" {
		ri.standalone.Enable("environment-variable")
		return true
	}
	
	// Check if no distributed components are available
	if ri.distributedWrapper == nil {
		ri.standalone.Enable("no-distributed-components")
		return true
	}
	
	// Check if local instance is unhealthy
	if !ri.fallbackManager.IsLocalHealthy() {
		ri.standalone.Enable("local-unhealthy")
		return true
	}
	
	return ri.standalone.IsEnabled()
}

// generateStandaloneRoutes generates routes for standalone mode
func (ri *RouteIntegration) generateStandaloneRoutes(rc *integration.Registry) (http.Handler, error) {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	
	// Add standalone middleware
	router.Use(ri.standaloneMiddleware())
	
	// All routes go through standalone handler
	router.Any("/*path", ri.standalone.HandleRequest)
	
	return router, nil
}

// standaloneMiddleware adds standalone-specific middleware
func (ri *RouteIntegration) standaloneMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Ollama-Mode", "standalone")
		c.Header("X-Ollama-Standalone", "true")
		c.Header("X-Ollama-Standalone-Reason", ri.standalone.GetReason())
		c.Next()
	}
}

// CreateIntegratedRouter creates a router with integrated distributed capabilities
func (ri *RouteIntegration) CreateIntegratedRouter(originalServer interface{}, rc *integration.Registry) (*gin.Engine, error) {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	
	// Add integration middleware
	router.Use(ri.integrationMiddleware())
	
	// Get original routes
	originalHandler, err := originalServer.GenerateRoutes(rc)
	if err != nil {
		return nil, fmt.Errorf("failed to get original routes: %w", err)
	}
	
	// Add distributed routes if enabled
	if ri.enabled && ri.compatibility != nil {
		// Inject distributed handlers
		ri.compatibility.InjectDistributedHandlers(router)
		
		// Register distributed endpoints
		ri.compatibility.RegisterDistributedEndpoints(router)
	}
	
	// Add fallback handler
	router.NoRoute(func(c *gin.Context) {
		// Check if we should use distributed
		if ri.shouldUseDistributed(c) {
			// Try distributed first
			if ri.compatibility != nil {
				handlers := ri.compatibility.GetDistributedHandlers()
				if handler, exists := handlers[ri.getHandlerName(c.Request.URL.Path)]; exists {
					handler(c)
					return
				}
			}
		}
		
		// Fall back to original
		originalHandler.ServeHTTP(c.Writer, c.Request)
	})
	
	return router, nil
}

// integrationMiddleware adds integration-specific middleware
func (ri *RouteIntegration) integrationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Ollama-Integration", "true")
		c.Header("X-Ollama-Integration-Enabled", fmt.Sprintf("%v", ri.enabled))
		c.Header("X-Ollama-Integration-Initialized", fmt.Sprintf("%v", ri.initialized))
		
		// Add health status
		if ri.fallbackManager != nil {
			c.Header("X-Ollama-Local-Healthy", fmt.Sprintf("%v", ri.fallbackManager.IsLocalHealthy()))
		}
		
		c.Next()
	}
}

// shouldUseDistributed determines if a request should use distributed routing
func (ri *RouteIntegration) shouldUseDistributed(c *gin.Context) bool {
	if !ri.enabled || ri.compatibility == nil {
		return false
	}
	
	// Check path
	path := c.Request.URL.Path
	
	// Use distributed for inference endpoints
	if strings.Contains(path, "/generate") ||
		strings.Contains(path, "/chat") ||
		strings.Contains(path, "/embed") ||
		strings.Contains(path, "/completions") {
		return true
	}
	
	// Use distributed for model management
	if strings.Contains(path, "/show") ||
		strings.Contains(path, "/tags") ||
		strings.Contains(path, "/pull") {
		return true
	}
	
	return false
}

// getHandlerName extracts handler name from path
func (ri *RouteIntegration) getHandlerName(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 3 {
		return parts[len(parts)-1]
	}
	return ""
}

// EnableDistributedMode enables distributed mode
func (ri *RouteIntegration) EnableDistributedMode() error {
	ri.enabled = true
	
	if ri.compatibility != nil {
		config := map[string]interface{}{
			"fallback": true,
		}
		return ri.compatibility.EnableDistributedMode(config)
	}
	
	return nil
}

// DisableDistributedMode disables distributed mode
func (ri *RouteIntegration) DisableDistributedMode() error {
	ri.enabled = false
	
	if ri.compatibility != nil {
		return ri.compatibility.DisableDistributedMode()
	}
	
	return nil
}

// EnableStandaloneMode enables standalone mode
func (ri *RouteIntegration) EnableStandaloneMode(reason string) {
	ri.standalone.Enable(reason)
}

// DisableStandaloneMode disables standalone mode
func (ri *RouteIntegration) DisableStandaloneMode() {
	ri.standalone.Disable()
}

// IsDistributedMode returns whether distributed mode is enabled
func (ri *RouteIntegration) IsDistributedMode() bool {
	return ri.enabled
}

// IsStandaloneMode returns whether standalone mode is enabled
func (ri *RouteIntegration) IsStandaloneMode() bool {
	return ri.standalone.IsEnabled()
}

// GetStatus returns integration status
func (ri *RouteIntegration) GetStatus() map[string]interface{} {
	status := map[string]interface{}{
		"enabled":     ri.enabled,
		"initialized": ri.initialized,
		"mode":        ri.getMode(),
	}
	
	if ri.fallbackManager != nil {
		status["fallback"] = ri.fallbackManager.GetStats()
	}
	
	if ri.standalone != nil {
		status["standalone"] = ri.standalone.GetStats()
	}
	
	if ri.compatibility != nil {
		status["distributed"] = ri.compatibility.GetDistributedStatus()
	}
	
	return status
}

// getMode returns the current mode
func (ri *RouteIntegration) getMode() string {
	if ri.standalone.IsEnabled() {
		return "standalone"
	}
	
	if ri.enabled {
		return "distributed"
	}
	
	return "local"
}

// Start starts the route integration
func (ri *RouteIntegration) Start(ctx context.Context) error {
	if ri.distributedWrapper != nil {
		return ri.distributedWrapper.Start(ctx)
	}
	return nil
}

// Shutdown shuts down the route integration
func (ri *RouteIntegration) Shutdown(ctx context.Context) error {
	if ri.distributedWrapper != nil {
		return ri.distributedWrapper.Shutdown(ctx)
	}
	return nil
}

// GetFallbackManager returns the fallback manager
func (ri *RouteIntegration) GetFallbackManager() *FallbackManager {
	return ri.fallbackManager
}

// GetDistributedWrapper returns the distributed wrapper
func (ri *RouteIntegration) GetDistributedWrapper() *DistributedServerWrapper {
	return ri.distributedWrapper
}

// GetCompatibility returns the compatibility layer
func (ri *RouteIntegration) GetCompatibility() *DistributedServerCompatibility {
	return ri.compatibility
}

// Global integration instance
var globalIntegration *RouteIntegration

// InitializeGlobalIntegration initializes the global integration
func InitializeGlobalIntegration(scheduler *scheduler.Engine, modelDist *models.Distribution, localAddr string) error {
	integration, err := NewRouteIntegration(scheduler, modelDist, localAddr)
	if err != nil {
		return fmt.Errorf("failed to create global integration: %w", err)
	}
	
	globalIntegration = integration
	return nil
}

// GetGlobalIntegration returns the global integration
func GetGlobalIntegration() *RouteIntegration {
	return globalIntegration
}

// WrapServerGenerateRoutes wraps the server's GenerateRoutes method
func WrapServerGenerateRoutes(originalServer interface{}, scheduler *scheduler.Engine, modelDist *models.Distribution, localAddr string) func(*integration.Registry) (http.Handler, error) {
	// Initialize global integration if not done
	if globalIntegration == nil {
		if err := InitializeGlobalIntegration(scheduler, modelDist, localAddr); err != nil {
			slog.Error("Failed to initialize global integration", "error", err)
			// Return original function
			return originalServer.GenerateRoutes
		}
	}
	
	// Return wrapped function
	return globalIntegration.WrapGenerateRoutes(originalServer, scheduler, modelDist, localAddr)
}

// Helper functions for external integration

// EnableDistributed enables distributed mode globally
func EnableDistributed() error {
	if globalIntegration == nil {
		return fmt.Errorf("global integration not initialized")
	}
	return globalIntegration.EnableDistributedMode()
}

// DisableDistributed disables distributed mode globally
func DisableDistributed() error {
	if globalIntegration == nil {
		return fmt.Errorf("global integration not initialized")
	}
	return globalIntegration.DisableDistributedMode()
}

// EnableStandalone enables standalone mode globally
func EnableStandalone(reason string) {
	if globalIntegration != nil {
		globalIntegration.EnableStandaloneMode(reason)
	}
}

// DisableStandalone disables standalone mode globally
func DisableStandalone() {
	if globalIntegration != nil {
		globalIntegration.DisableStandaloneMode()
	}
}

// GetIntegrationStatus returns global integration status
func GetIntegrationStatus() map[string]interface{} {
	if globalIntegration == nil {
		return map[string]interface{}{
			"initialized": false,
			"error":       "global integration not initialized",
		}
	}
	return globalIntegration.GetStatus()
}
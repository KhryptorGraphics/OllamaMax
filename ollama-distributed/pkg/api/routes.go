package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ollama/ollama-distributed/pkg/integration"
	"github.com/ollama/ollama-distributed/pkg/models"
	"github.com/ollama/ollama-distributed/pkg/scheduler"
)

// DistributedRoutes provides distributed routing capability for Ollama
type DistributedRoutes struct {
	scheduler         *scheduler.Engine
	integrationLayer  *IntegrationLayer
	distributedRunner *DistributedRunner
	modelDistribution *models.Distribution
	
	// Original server for fallback
	originalServer integration.Server
	
	// Configuration
	distributedMode bool
	fallbackMode    bool
	localAddr       string
}

// NewDistributedRoutes creates a new distributed routes handler
func NewDistributedRoutes(scheduler *scheduler.Engine, modelDist *models.Distribution, localAddr string) (*DistributedRoutes, error) {
	// Create integration layer
	integrationLayer, err := NewIntegrationLayer(scheduler, localAddr, modelDist)
	if err != nil {
		return nil, fmt.Errorf("failed to create integration layer: %w", err)
	}
	
	// Create distributed runner
	distributedRunner := NewDistributedRunner(scheduler, integrationLayer)
	
	return &DistributedRoutes{
		scheduler:         scheduler,
		integrationLayer:  integrationLayer,
		distributedRunner: distributedRunner,
		modelDistribution: modelDist,
		distributedMode:   true,
		fallbackMode:      true,
		localAddr:         localAddr,
	}, nil
}

// Start starts the distributed routes system
func (dr *DistributedRoutes) Start(ctx context.Context) error {
	// Start distributed runner
	if err := dr.distributedRunner.Start(ctx); err != nil {
		return fmt.Errorf("failed to start distributed runner: %w", err)
	}
	
	slog.Info("Distributed routes started", "localAddr", dr.localAddr)
	return nil
}

// SetupRoutes sets up routes with distributed capabilities
func (dr *DistributedRoutes) SetupRoutes(router *gin.Engine) {
	// Add distributed middleware
	router.Use(dr.distributedMiddleware())
	
	// API routes with distributed handling
	api := router.Group("/api")
	{
		// Inference endpoints
		api.POST("/generate", dr.handleGenerate)
		api.POST("/chat", dr.handleChat)
		api.POST("/embed", dr.handleEmbed)
		api.POST("/embeddings", dr.handleEmbeddings)
		
		// Model management
		api.POST("/pull", dr.handlePull)
		api.POST("/push", dr.handlePush)
		api.POST("/show", dr.handleShow)
		api.GET("/tags", dr.handleTags)
		api.DELETE("/delete", dr.handleDelete)
		api.POST("/copy", dr.handleCopy)
		api.POST("/create", dr.handleCreate)
		
		// System endpoints
		api.GET("/ps", dr.handlePs)
		api.GET("/version", dr.handleVersion)
		
		// Distributed-specific endpoints
		api.GET("/distributed/status", dr.handleDistributedStatus)
		api.GET("/distributed/nodes", dr.handleDistributedNodes)
		api.GET("/distributed/models", dr.handleDistributedModels)
		api.POST("/distributed/rebalance", dr.handleRebalance)
		api.POST("/distributed/migrate", dr.handleMigrate)
	}
	
	// OpenAI compatibility endpoints
	v1 := router.Group("/v1")
	{
		v1.POST("/chat/completions", dr.handleOpenAIChat)
		v1.POST("/completions", dr.handleOpenAICompletion)
		v1.POST("/embeddings", dr.handleOpenAIEmbeddings)
		v1.GET("/models", dr.handleOpenAIModels)
		v1.GET("/models/:model", dr.handleOpenAIModel)
	}
	
	// Health and metrics
	router.GET("/health", dr.handleHealth)
	router.GET("/metrics", dr.handleMetrics)
	
	// Distributed admin endpoints
	admin := router.Group("/admin")
	{
		admin.Use(dr.adminAuthMiddleware())
		admin.POST("/mode", dr.handleSetMode)
		admin.POST("/fallback", dr.handleSetFallback)
		admin.POST("/rebalance", dr.handleForceRebalance)
		admin.GET("/stats", dr.handleStats)
	}
}

// Middleware

func (dr *DistributedRoutes) distributedMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add distributed headers
		c.Header("X-Ollama-Distributed", "true")
		c.Header("X-Ollama-Version", "distributed")
		c.Header("X-Ollama-Cluster-Size", fmt.Sprintf("%d", dr.scheduler.GetClusterSize()))
		c.Header("X-Ollama-Mode", dr.getMode())
		
		// Add request ID for tracing
		requestID := fmt.Sprintf("req_%d", time.Now().UnixNano())
		c.Header("X-Request-ID", requestID)
		c.Set("requestID", requestID)
		
		// Continue processing
		c.Next()
	}
}

func (dr *DistributedRoutes) adminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Simple token-based auth for admin endpoints
		token := c.GetHeader("Authorization")
		if token == "" || token != "Bearer "+os.Getenv("OLLAMA_ADMIN_TOKEN") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// Core API handlers

func (dr *DistributedRoutes) handleGenerate(c *gin.Context) {
	if dr.distributedMode {
		dr.integrationLayer.HandleRequest(c)
	} else {
		dr.fallbackToOriginal(c)
	}
}

func (dr *DistributedRoutes) handleChat(c *gin.Context) {
	if dr.distributedMode {
		dr.integrationLayer.HandleRequest(c)
	} else {
		dr.fallbackToOriginal(c)
	}
}

func (dr *DistributedRoutes) handleEmbed(c *gin.Context) {
	if dr.distributedMode {
		dr.integrationLayer.HandleRequest(c)
	} else {
		dr.fallbackToOriginal(c)
	}
}

func (dr *DistributedRoutes) handleEmbeddings(c *gin.Context) {
	if dr.distributedMode {
		dr.integrationLayer.HandleRequest(c)
	} else {
		dr.fallbackToOriginal(c)
	}
}

func (dr *DistributedRoutes) handlePull(c *gin.Context) {
	if dr.distributedMode {
		dr.integrationLayer.HandleRequest(c)
	} else {
		dr.fallbackToOriginal(c)
	}
}

func (dr *DistributedRoutes) handlePush(c *gin.Context) {
	if dr.distributedMode {
		dr.integrationLayer.HandleRequest(c)
	} else {
		dr.fallbackToOriginal(c)
	}
}

func (dr *DistributedRoutes) handleShow(c *gin.Context) {
	if dr.distributedMode {
		dr.integrationLayer.HandleRequest(c)
	} else {
		dr.fallbackToOriginal(c)
	}
}

func (dr *DistributedRoutes) handleTags(c *gin.Context) {
	if dr.distributedMode {
		dr.integrationLayer.HandleRequest(c)
	} else {
		dr.fallbackToOriginal(c)
	}
}

func (dr *DistributedRoutes) handleDelete(c *gin.Context) {
	if dr.distributedMode {
		dr.integrationLayer.HandleRequest(c)
	} else {
		dr.fallbackToOriginal(c)
	}
}

func (dr *DistributedRoutes) handleCopy(c *gin.Context) {
	if dr.distributedMode {
		dr.integrationLayer.HandleRequest(c)
	} else {
		dr.fallbackToOriginal(c)
	}
}

func (dr *DistributedRoutes) handleCreate(c *gin.Context) {
	if dr.distributedMode {
		dr.integrationLayer.HandleRequest(c)
	} else {
		dr.fallbackToOriginal(c)
	}
}

func (dr *DistributedRoutes) handlePs(c *gin.Context) {
	if dr.distributedMode {
		dr.integrationLayer.HandleRequest(c)
	} else {
		dr.fallbackToOriginal(c)
	}
}

func (dr *DistributedRoutes) handleVersion(c *gin.Context) {
	if dr.distributedMode {
		dr.integrationLayer.HandleRequest(c)
	} else {
		dr.fallbackToOriginal(c)
	}
}

// OpenAI compatibility handlers

func (dr *DistributedRoutes) handleOpenAIChat(c *gin.Context) {
	if dr.distributedMode {
		dr.integrationLayer.HandleRequest(c)
	} else {
		dr.fallbackToOriginal(c)
	}
}

func (dr *DistributedRoutes) handleOpenAICompletion(c *gin.Context) {
	if dr.distributedMode {
		dr.integrationLayer.HandleRequest(c)
	} else {
		dr.fallbackToOriginal(c)
	}
}

func (dr *DistributedRoutes) handleOpenAIEmbeddings(c *gin.Context) {
	if dr.distributedMode {
		dr.integrationLayer.HandleRequest(c)
	} else {
		dr.fallbackToOriginal(c)
	}
}

func (dr *DistributedRoutes) handleOpenAIModels(c *gin.Context) {
	if dr.distributedMode {
		dr.integrationLayer.HandleRequest(c)
	} else {
		dr.fallbackToOriginal(c)
	}
}

func (dr *DistributedRoutes) handleOpenAIModel(c *gin.Context) {
	if dr.distributedMode {
		dr.integrationLayer.HandleRequest(c)
	} else {
		dr.fallbackToOriginal(c)
	}
}

// Distributed-specific handlers

func (dr *DistributedRoutes) handleDistributedStatus(c *gin.Context) {
	status := map[string]interface{}{
		"distributed_mode": dr.distributedMode,
		"fallback_mode":    dr.fallbackMode,
		"cluster_size":     dr.scheduler.GetClusterSize(),
		"active_nodes":     dr.scheduler.GetActiveNodes(),
		"scheduler_stats":  dr.scheduler.GetStats(),
		"runner_stats":     dr.distributedRunner.GetStats(),
		"integration_stats": dr.integrationLayer.GetStats(),
	}
	
	c.JSON(http.StatusOK, status)
}

func (dr *DistributedRoutes) handleDistributedNodes(c *gin.Context) {
	nodes := dr.scheduler.GetNodes()
	c.JSON(http.StatusOK, gin.H{"nodes": nodes})
}

func (dr *DistributedRoutes) handleDistributedModels(c *gin.Context) {
	models := dr.modelDistribution.GetDistributedModels()
	c.JSON(http.StatusOK, gin.H{"models": models})
}

func (dr *DistributedRoutes) handleRebalance(c *gin.Context) {
	if err := dr.modelDistribution.Rebalance(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Rebalance initiated"})
}

func (dr *DistributedRoutes) handleMigrate(c *gin.Context) {
	var req struct {
		ModelName string `json:"model_name"`
		FromNode  string `json:"from_node"`
		ToNode    string `json:"to_node"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := dr.modelDistribution.MigrateModel(req.ModelName, req.FromNode, req.ToNode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Migration initiated"})
}

// System handlers

func (dr *DistributedRoutes) handleHealth(c *gin.Context) {
	health := map[string]interface{}{
		"status":       "healthy",
		"timestamp":    time.Now().Unix(),
		"distributed":  dr.distributedMode,
		"cluster_size": dr.scheduler.GetClusterSize(),
		"services": map[string]interface{}{
			"scheduler":   dr.scheduler.IsHealthy(),
			"runner":      len(dr.distributedRunner.GetActiveRunners()) > 0,
			"integration": dr.integrationLayer != nil,
		},
	}
	
	c.JSON(http.StatusOK, health)
}

func (dr *DistributedRoutes) handleMetrics(c *gin.Context) {
	metrics := map[string]interface{}{
		"scheduler":   dr.scheduler.GetStats(),
		"runner":      dr.distributedRunner.GetStats(),
		"integration": dr.integrationLayer.GetStats(),
		"models":      dr.modelDistribution.GetStats(),
	}
	
	c.JSON(http.StatusOK, metrics)
}

// Admin handlers

func (dr *DistributedRoutes) handleSetMode(c *gin.Context) {
	var req struct {
		Mode string `json:"mode"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	switch req.Mode {
	case "distributed":
		dr.distributedMode = true
		dr.integrationLayer.SetDistributedMode(true)
	case "local":
		dr.distributedMode = false
		dr.integrationLayer.SetDistributedMode(false)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mode"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Mode set to %s", req.Mode)})
}

func (dr *DistributedRoutes) handleSetFallback(c *gin.Context) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	dr.fallbackMode = req.Enabled
	dr.integrationLayer.SetFallbackMode(req.Enabled)
	
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Fallback mode set to %v", req.Enabled)})
}

func (dr *DistributedRoutes) handleForceRebalance(c *gin.Context) {
	if err := dr.modelDistribution.ForceRebalance(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Force rebalance initiated"})
}

func (dr *DistributedRoutes) handleStats(c *gin.Context) {
	stats := map[string]interface{}{
		"mode":             dr.getMode(),
		"uptime":           time.Since(time.Now()).String(),
		"cluster_size":     dr.scheduler.GetClusterSize(),
		"active_nodes":     dr.scheduler.GetActiveNodes(),
		"total_models":     dr.modelDistribution.GetTotalModels(),
		"distributed_models": dr.modelDistribution.GetDistributedModelCount(),
		"active_runners":   len(dr.distributedRunner.GetActiveRunners()),
		"scheduler_stats":  dr.scheduler.GetStats(),
		"runner_stats":     dr.distributedRunner.GetStats(),
		"integration_stats": dr.integrationLayer.GetStats(),
	}
	
	c.JSON(http.StatusOK, stats)
}

// Helper methods

func (dr *DistributedRoutes) getMode() string {
	if dr.distributedMode {
		if dr.fallbackMode {
			return "distributed-with-fallback"
		}
		return "distributed"
	}
	return "local"
}

func (dr *DistributedRoutes) fallbackToOriginal(c *gin.Context) {
	if dr.originalServer != nil {
		// TODO: Implement fallback to original server
		c.Header("X-Ollama-Fallback", "original-server")
	}
	
	// For now, proxy to local
	dr.integrationLayer.proxyToLocal(c)
}

// SetOriginalServer sets the original server for fallback
func (dr *DistributedRoutes) SetOriginalServer(server integration.Server) {
	dr.originalServer = server
}

// GetIntegrationLayer returns the integration layer
func (dr *DistributedRoutes) GetIntegrationLayer() *IntegrationLayer {
	return dr.integrationLayer
}

// GetDistributedRunner returns the distributed runner
func (dr *DistributedRoutes) GetDistributedRunner() *DistributedRunner {
	return dr.distributedRunner
}

// IsDistributedMode returns whether distributed mode is enabled
func (dr *DistributedRoutes) IsDistributedMode() bool {
	return dr.distributedMode
}

// IsFallbackMode returns whether fallback mode is enabled
func (dr *DistributedRoutes) IsFallbackMode() bool {
	return dr.fallbackMode
}

// Shutdown gracefully shuts down the distributed routes
func (dr *DistributedRoutes) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down distributed routes")
	
	// Shutdown distributed runner
	if err := dr.distributedRunner.Shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown distributed runner", "error", err)
		return err
	}
	
	return nil
}
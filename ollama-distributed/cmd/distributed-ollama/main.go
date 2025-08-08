package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/api"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/inference"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/models"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/distributed"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/orchestration"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/partitioning"
)

// DistributedOllamaServer represents the main distributed Ollama server
type DistributedOllamaServer struct {
	// Core components
	p2pNode         *p2p.Node
	modelManager    *models.DistributedModelManager
	inferenceEngine *inference.DistributedInferenceEngine
	scheduler       *distributed.DistributedScheduler
	integration     *api.DistributedOllamaIntegration

	// HTTP server
	httpServer *http.Server
	router     *gin.Engine

	// Configuration
	config *config.DistributedConfig
	logger *slog.Logger

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
}

func main() {
	// Parse command line flags
	var (
		configPath = flag.String("config", "config.yaml", "Path to configuration file")
		port       = flag.Int("port", 11434, "HTTP server port")
		p2pPort    = flag.Int("p2p-port", 4001, "P2P network port")
		logLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
		bootstrap  = flag.String("bootstrap", "", "Bootstrap peer address")
	)
	flag.Parse()

	// Setup logging
	logger := setupLogging(*logLevel)

	logger.Info("Starting Distributed Ollama Server",
		"version", "1.0.0",
		"port", *port,
		"p2p_port", *p2pPort)

	// Load configuration
	cfg, err := config.LoadDistributedConfig(*configPath)
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Override config with command line flags
	if *port != 11434 {
		cfg.API.Port = *port
	}
	// P2P port will be handled when creating the P2P node

	// Create main context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create and start the distributed server
	server, err := NewDistributedOllamaServer(ctx, cfg, logger)
	if err != nil {
		logger.Error("Failed to create distributed server", "error", err)
		os.Exit(1)
	}

	// Add bootstrap peer if specified
	if *bootstrap != "" {
		logger.Info("Adding bootstrap peer", "address", *bootstrap)
		// This would connect to the bootstrap peer
		// server.p2pNode.ConnectToPeer(*bootstrap)
	}

	// Start the server
	if err := server.Start(); err != nil {
		logger.Error("Failed to start server", "error", err)
		os.Exit(1)
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	logger.Info("Received shutdown signal, stopping server...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Stop(shutdownCtx); err != nil {
		logger.Error("Error during shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("Server stopped successfully")
}

// NewDistributedOllamaServer creates a new distributed Ollama server
func NewDistributedOllamaServer(
	ctx context.Context,
	cfg *config.DistributedConfig,
	logger *slog.Logger,
) (*DistributedOllamaServer, error) {
	serverCtx, cancel := context.WithCancel(ctx)

	// Initialize P2P node with default config for now
	// In a real implementation, this would use the proper config conversion
	p2pNode, err := p2p.NewNode(serverCtx, nil)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create P2P node: %w", err)
	}

	// Initialize model manager with basic config
	// Create a minimal config for the model manager
	basicConfig := &config.DistributedConfig{}
	basicConfig.Models.StoragePath = "./models"
	basicConfig.Models.CacheSize = "1GB"

	modelManager, err := models.NewDistributedModelManager(basicConfig, p2pNode, logger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create model manager: %w", err)
	}

	// Initialize partition manager
	partitionManager := partitioning.NewPartitionManager(&partitioning.Config{
		DefaultStrategy: "layerwise",
		LayerThreshold:  10,
		BatchSizeLimit:  1024,
	})

	// Initialize orchestration engine
	orchestrator := orchestration.NewOrchestrationEngine(&orchestration.Config{
		MaxConcurrentTasks: 100,
		TaskTimeout:        5 * time.Minute,
	})

	// Initialize distributed scheduler with nil for now
	// In a real implementation, this would use proper config and consensus engine
	scheduler, err := distributed.NewDistributedScheduler(nil, nil, p2pNode, nil)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	// Initialize distributed inference engine
	inferenceConfig := &inference.DistributedInferenceConfig{
		MaxConcurrentInferences: 10,
		InferenceTimeout:        5 * time.Minute,
		PartitionStrategy:       "layerwise",
		AggregationStrategy:     "concat",
		MinNodesRequired:        2,
		LoadBalancingEnabled:    true,
		FaultToleranceEnabled:   true,
	}

	inferenceEngine := inference.NewDistributedInferenceEngine(
		p2pNode,
		modelManager,
		partitionManager,
		orchestrator,
		inferenceConfig,
	)

	// Initialize distributed integration
	integrationConfig := &api.DistributedIntegrationConfig{
		MinModelSizeForDistribution: 4 * 1024 * 1024 * 1024, // 4GB
		MinNodesForDistribution:     2,
		MaxConcurrentRequests:       10,
		RequestTimeout:              5 * time.Minute,
		DefaultStrategy:             "layerwise",
		EnableLoadBalancing:         true,
		EnableFaultTolerance:        true,
		EnableCaching:               true,
		CacheSize:                   100,
		EnablePrefetching:           false,
	}

	integration := api.NewDistributedOllamaIntegration(
		inferenceEngine,
		modelManager,
		scheduler,
		p2pNode,
		integrationConfig,
		logger,
	)

	// Setup HTTP router
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	server := &DistributedOllamaServer{
		p2pNode:         p2pNode,
		modelManager:    modelManager,
		inferenceEngine: inferenceEngine,
		scheduler:       scheduler,
		integration:     integration,
		router:          router,
		config:          cfg,
		logger:          logger,
		ctx:             serverCtx,
		cancel:          cancel,
	}

	// Setup HTTP routes
	server.setupRoutes()

	// Create HTTP server
	server.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.API.Port),
		Handler: router,
	}

	return server, nil
}

// Start starts the distributed Ollama server
func (s *DistributedOllamaServer) Start() error {
	s.logger.Info("Starting distributed Ollama server components")

	// Start P2P node
	if err := s.p2pNode.Start(); err != nil {
		return fmt.Errorf("failed to start P2P node: %w", err)
	}

	// Start model manager
	if err := s.modelManager.Start(); err != nil {
		return fmt.Errorf("failed to start model manager: %w", err)
	}

	// Start scheduler
	if err := s.scheduler.Start(); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	// Start integration
	if err := s.integration.Start(); err != nil {
		return fmt.Errorf("failed to start integration: %w", err)
	}

	// Start HTTP server
	go func() {
		s.logger.Info("Starting HTTP server", "address", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server error", "error", err)
		}
	}()

	s.logger.Info("Distributed Ollama server started successfully")
	return nil
}

// Stop stops the distributed Ollama server
func (s *DistributedOllamaServer) Stop(ctx context.Context) error {
	s.logger.Info("Stopping distributed Ollama server")

	// Stop HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("Error stopping HTTP server", "error", err)
	}

	// Stop integration
	if err := s.integration.Stop(); err != nil {
		s.logger.Error("Error stopping integration", "error", err)
	}

	// Stop scheduler
	if err := s.scheduler.Shutdown(ctx); err != nil {
		s.logger.Error("Error stopping scheduler", "error", err)
	}

	// Stop model manager
	if err := s.modelManager.Shutdown(ctx); err != nil {
		s.logger.Error("Error stopping model manager", "error", err)
	}

	// Stop P2P node
	if err := s.p2pNode.Stop(); err != nil {
		s.logger.Error("Error stopping P2P node", "error", err)
	}

	// Cancel context
	s.cancel()

	s.logger.Info("Distributed Ollama server stopped")
	return nil
}

// setupRoutes sets up HTTP routes
func (s *DistributedOllamaServer) setupRoutes() {
	// Ollama-compatible API routes
	api := s.router.Group("/api")
	{
		api.POST("/generate", s.handleGenerate)
		api.POST("/chat", s.handleChat)
		api.GET("/tags", s.handleListModels)
		api.POST("/pull", s.handlePullModel)
		api.DELETE("/delete", s.handleDeleteModel)
		api.POST("/push", s.handlePushModel)
		api.POST("/create", s.handleCreateModel)
		api.POST("/copy", s.handleCopyModel)
		api.POST("/show", s.handleShowModel)
		api.POST("/embed", s.handleEmbed)
	}

	// Distributed-specific API routes
	distributed := s.router.Group("/api/distributed")
	{
		distributed.GET("/status", s.handleDistributedStatus)
		distributed.GET("/nodes", s.handleListNodes)
		distributed.GET("/models", s.handleDistributedModels)
		distributed.GET("/metrics", s.handleMetrics)
		distributed.GET("/requests", s.handleActiveRequests)
	}

	// Health check
	s.router.GET("/health", s.handleHealth)
}

// setupLogging sets up structured logging
func setupLogging(level string) *slog.Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})

	return slog.New(handler)
}

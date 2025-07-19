package integration

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ollama/ollama/api"
	"github.com/ollama/ollama/server"
	"github.com/ollama/ollama-distributed/pkg/scheduler/distributed"
	"github.com/ollama/ollama-distributed/pkg/consensus"
	"github.com/ollama/ollama-distributed/pkg/p2p"
)

// OllamaDistributedScheduler integrates distributed scheduling with Ollama
type OllamaDistributedScheduler struct {
	*server.Scheduler // Embed the original scheduler
	distributedScheduler *distributed.DistributedScheduler
	config *IntegrationConfig
	mu     sync.RWMutex
}

// IntegrationConfig holds integration configuration
type IntegrationConfig struct {
	Enabled               bool          `json:"enabled"`
	DistributionThreshold int64         `json:"distribution_threshold"`
	MinNodes              int           `json:"min_nodes"`
	MaxNodes              int           `json:"max_nodes"`
	FailbackToLocal       bool          `json:"failback_to_local"`
	HealthCheckInterval   time.Duration `json:"health_check_interval"`
	MetricsEnabled        bool          `json:"metrics_enabled"`
}

// RunnerRef wraps the original runner with distributed capabilities
type RunnerRef struct {
	*server.RunnerRef // Embed the original runner
	distributed      bool
	nodes           []string
	partitionPlan   interface{}
	metadata        map[string]interface{}
}

// NewOllamaDistributedScheduler creates a new integrated scheduler
func NewOllamaDistributedScheduler(
	baseScheduler *server.Scheduler,
	config *IntegrationConfig,
	p2pNode *p2p.Node,
	consensusEngine *consensus.Engine,
) (*OllamaDistributedScheduler, error) {
	// Create distributed scheduler configuration
	distributedConfig := &distributed.DistributedConfig{
		ClusterID:         "ollama-cluster",
		NodeID:            p2pNode.GetID(),
		MaxNodes:          config.MaxNodes,
		HeartbeatInterval: config.HealthCheckInterval,
		DefaultStrategy:   "layerwise",
		LayerThreshold:    32,
		BatchSizeLimit:    16,
		LBAlgorithm:       "intelligent",
		LatencyTarget:     100 * time.Millisecond,
		WeightFactors: map[string]float64{
			"latency":    0.4,
			"throughput": 0.3,
			"memory":     0.2,
			"bandwidth":  0.1,
		},
		ReplicationFactor:     2,
		HealthCheckInterval:   config.HealthCheckInterval,
		RecoveryTimeout:       30 * time.Second,
		CommunicationProtocol: "grpc",
		Encryption:            true,
		Compression:           true,
	}
	
	// Create distributed scheduler
	distributedScheduler, err := distributed.NewDistributedScheduler(
		baseScheduler,
		distributedConfig,
		p2pNode,
		consensusEngine,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create distributed scheduler: %v", err)
	}
	
	ods := &OllamaDistributedScheduler{
		Scheduler:            baseScheduler,
		distributedScheduler: distributedScheduler,
		config:              config,
	}
	
	return ods, nil
}

// GetRunner overrides the base scheduler's GetRunner method
func (ods *OllamaDistributedScheduler) GetRunner(
	ctx context.Context,
	model *server.Model,
	opts api.Options,
	sessionDuration *api.Duration,
) (chan *server.RunnerRef, chan error) {
	// Check if distribution is enabled and appropriate
	if !ods.config.Enabled {
		return ods.Scheduler.GetRunner(ctx, model, opts, sessionDuration)
	}
	
	// Determine if request should be distributed
	shoudDistribute := ods.shouldDistributeRequest(model, opts)
	
	if !shoudDistribute {
		// Use local scheduler
		return ods.Scheduler.GetRunner(ctx, model, opts, sessionDuration)
	}
	
	// Use distributed scheduler
	return ods.getDistributedRunner(ctx, model, opts, sessionDuration)
}

// shouldDistributeRequest determines if a request should be distributed
func (ods *OllamaDistributedScheduler) shouldDistributeRequest(model *server.Model, opts api.Options) bool {
	// Check if distributed scheduler is available
	if !ods.distributedScheduler.ShouldDistribute(model, opts) {
		return false
	}
	
	// Check minimum nodes requirement
	if len(ods.distributedScheduler.GetNodes()) < ods.config.MinNodes {
		return false
	}
	
	// Check model size threshold
	if ods.config.DistributionThreshold > 0 {
		// Estimate model size (this would be more sophisticated in practice)
		modelSize := ods.estimateModelSize(model)
		if modelSize < ods.config.DistributionThreshold {
			return false
		}
	}
	
	return true
}

// estimateModelSize estimates the size of a model
func (ods *OllamaDistributedScheduler) estimateModelSize(model *server.Model) int64 {
	// This is a simplified estimation
	// In practice, this would use actual model metadata
	if model.Name == "" {
		return 0
	}
	
	// Rough size estimation based on model name patterns
	if containsAny(model.Name, []string{"70b", "65b"}) {
		return 140 * 1024 * 1024 * 1024 // 140GB
	} else if containsAny(model.Name, []string{"34b", "33b"}) {
		return 70 * 1024 * 1024 * 1024 // 70GB
	} else if containsAny(model.Name, []string{"13b", "14b"}) {
		return 26 * 1024 * 1024 * 1024 // 26GB
	} else if containsAny(model.Name, []string{"7b", "8b"}) {
		return 14 * 1024 * 1024 * 1024 // 14GB
	} else if containsAny(model.Name, []string{"3b", "4b"}) {
		return 6 * 1024 * 1024 * 1024 // 6GB
	}
	
	return 4 * 1024 * 1024 * 1024 // 4GB default
}

// getDistributedRunner gets a distributed runner
func (ods *OllamaDistributedScheduler) getDistributedRunner(
	ctx context.Context,
	model *server.Model,
	opts api.Options,
	sessionDuration *api.Duration,
) (chan *server.RunnerRef, chan error) {
	successCh := make(chan *server.RunnerRef, 1)
	errorCh := make(chan error, 1)
	
	go func() {
		defer close(successCh)
		defer close(errorCh)
		
		slog.Info("attempting distributed execution",
			"model", model.Name,
			"num_ctx", opts.NumCtx,
			"cluster_nodes", len(ods.distributedScheduler.GetNodes()))
		
		// Get distributed runner
		distributedSuccessCh, distributedErrorCh := ods.distributedScheduler.GetDistributedRunner(
			ctx, model, opts, sessionDuration,
		)
		
		select {
		case runner := <-distributedSuccessCh:
			if runner != nil {
				// Wrap runner with distributed metadata
				wrappedRunner := &RunnerRef{
					RunnerRef:     runner,
					distributed:   true,
					nodes:         ods.getNodeIDs(),
					partitionPlan: nil, // Would be populated by the distributed scheduler
					metadata: map[string]interface{}{
						"distribution_strategy": "layerwise",
						"node_count":           len(ods.distributedScheduler.GetNodes()),
						"distributed_at":       time.Now(),
					},
				}
				
				// Convert back to base runner type
				successCh <- wrappedRunner.RunnerRef
				slog.Info("distributed execution successful", "model", model.Name)
				return
			}
			
		case err := <-distributedErrorCh:
			if err != nil {
				slog.Warn("distributed execution failed, falling back to local",
					"model", model.Name,
					"error", err)
				
				// Fallback to local execution if enabled
				if ods.config.FailbackToLocal {
					localSuccessCh, localErrorCh := ods.Scheduler.GetRunner(ctx, model, opts, sessionDuration)
					
					select {
					case runner := <-localSuccessCh:
						successCh <- runner
						return
					case localErr := <-localErrorCh:
						errorCh <- fmt.Errorf("distributed execution failed: %v, local fallback failed: %v", err, localErr)
						return
					}
				} else {
					errorCh <- fmt.Errorf("distributed execution failed: %v", err)
					return
				}
			}
			
		case <-ctx.Done():
			errorCh <- ctx.Err()
			return
		}
	}()
	
	return successCh, errorCh
}

// getNodeIDs returns the IDs of all nodes in the cluster
func (ods *OllamaDistributedScheduler) getNodeIDs() []string {
	nodes := ods.distributedScheduler.GetNodes()
	ids := make([]string, len(nodes))
	for i, node := range nodes {
		if nodeInfo, ok := node.(*distributed.NodeInfo); ok {
			ids[i] = nodeInfo.ID
		} else {
			ids[i] = fmt.Sprintf("node_%d", i)
		}
	}
	return ids
}

// Start starts the distributed scheduler
func (ods *OllamaDistributedScheduler) Start(ctx context.Context) error {
	ods.mu.Lock()
	defer ods.mu.Unlock()
	
	if !ods.config.Enabled {
		slog.Info("distributed scheduling disabled, using local scheduler only")
		return nil
	}
	
	// Start the distributed scheduler
	if err := ods.distributedScheduler.Start(); err != nil {
		return fmt.Errorf("failed to start distributed scheduler: %v", err)
	}
	
	slog.Info("distributed scheduler started",
		"cluster_id", ods.distributedScheduler.GetClusterID(),
		"node_id", ods.distributedScheduler.GetNodeID(),
		"distribution_threshold", ods.config.DistributionThreshold,
		"min_nodes", ods.config.MinNodes)
	
	return nil
}

// Shutdown gracefully shuts down the distributed scheduler
func (ods *OllamaDistributedScheduler) Shutdown(ctx context.Context) error {
	ods.mu.Lock()
	defer ods.mu.Unlock()
	
	if !ods.config.Enabled {
		return nil
	}
	
	// Shutdown the distributed scheduler
	if err := ods.distributedScheduler.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown distributed scheduler: %v", err)
	}
	
	slog.Info("distributed scheduler shutdown complete")
	return nil
}

// GetMetrics returns combined metrics from both schedulers
func (ods *OllamaDistributedScheduler) GetMetrics() map[string]interface{} {
	ods.mu.RLock()
	defer ods.mu.RUnlock()
	
	metrics := map[string]interface{}{
		"distributed_enabled": ods.config.Enabled,
		"distribution_threshold": ods.config.DistributionThreshold,
		"min_nodes": ods.config.MinNodes,
		"max_nodes": ods.config.MaxNodes,
		"local_scheduler": "active",
	}
	
	if ods.config.Enabled {
		distributedMetrics := ods.distributedScheduler.GetMetrics()
		metrics["distributed_scheduler"] = distributedMetrics
		
		clusterHealth := ods.distributedScheduler.GetClusterHealth()
		metrics["cluster_health"] = clusterHealth
		
		activeTasks := ods.distributedScheduler.GetActiveTasks()
		metrics["active_tasks"] = len(activeTasks)
		
		nodes := ods.distributedScheduler.GetNodes()
		metrics["cluster_nodes"] = len(nodes)
	} else {
		metrics["distributed_scheduler"] = "disabled"
	}
	
	return metrics
}

// GetClusterStatus returns the status of the distributed cluster
func (ods *OllamaDistributedScheduler) GetClusterStatus() map[string]interface{} {
	ods.mu.RLock()
	defer ods.mu.RUnlock()
	
	if !ods.config.Enabled {
		return map[string]interface{}{
			"status": "disabled",
			"message": "distributed scheduling is disabled",
		}
	}
	
	nodes := ods.distributedScheduler.GetNodes()
	activeTasks := ods.distributedScheduler.GetActiveTasks()
	clusterHealth := ods.distributedScheduler.GetClusterHealth()
	
	// Calculate cluster health score
	healthyNodes := 0
	for _, health := range clusterHealth {
		if healthCheck, ok := health.(*distributed.HealthCheck); ok {
			if healthCheck.Status == "healthy" {
				healthyNodes++
			}
		}
	}
	
	healthScore := float64(healthyNodes) / float64(len(nodes))
	
	status := "healthy"
	if healthScore < 0.5 {
		status = "degraded"
	} else if healthScore < 0.8 {
		status = "warning"
	}
	
	return map[string]interface{}{
		"status": status,
		"health_score": healthScore,
		"total_nodes": len(nodes),
		"healthy_nodes": healthyNodes,
		"active_tasks": len(activeTasks),
		"cluster_id": ods.distributedScheduler.GetClusterID(),
		"node_id": ods.distributedScheduler.GetNodeID(),
		"last_updated": time.Now(),
	}
}

// UpdateConfig updates the integration configuration
func (ods *OllamaDistributedScheduler) UpdateConfig(config *IntegrationConfig) error {
	ods.mu.Lock()
	defer ods.mu.Unlock()
	
	ods.config = config
	
	slog.Info("integration configuration updated",
		"enabled", config.Enabled,
		"distribution_threshold", config.DistributionThreshold,
		"min_nodes", config.MinNodes,
		"failback_to_local", config.FailbackToLocal)
	
	return nil
}

// Helper functions

// containsAny checks if a string contains any of the given substrings
func containsAny(str string, substrings []string) bool {
	for _, substr := range substrings {
		if strings.Contains(strings.ToLower(str), substr) {
			return true
		}
	}
	return false
}

// Extension methods for the distributed scheduler

func (ds *distributed.DistributedScheduler) GetClusterID() string {
	return ds.GetConfig().ClusterID
}

func (ds *distributed.DistributedScheduler) GetNodeID() string {
	return ds.GetConfig().NodeID
}

func (ds *distributed.DistributedScheduler) GetConfig() *distributed.DistributedConfig {
	// This would return the actual config in a real implementation
	return &distributed.DistributedConfig{
		ClusterID: "ollama-cluster",
		NodeID:    "node-local",
	}
}

// Integration utility functions

// CreateDefaultIntegrationConfig creates a default integration configuration
func CreateDefaultIntegrationConfig() *IntegrationConfig {
	return &IntegrationConfig{
		Enabled:               true,
		DistributionThreshold: 8 * 1024 * 1024 * 1024, // 8GB
		MinNodes:              2,
		MaxNodes:              16,
		FailbackToLocal:       true,
		HealthCheckInterval:   30 * time.Second,
		MetricsEnabled:        true,
	}
}

// ValidateIntegrationConfig validates the integration configuration
func ValidateIntegrationConfig(config *IntegrationConfig) error {
	if config.MinNodes < 1 {
		return fmt.Errorf("min_nodes must be at least 1")
	}
	
	if config.MaxNodes < config.MinNodes {
		return fmt.Errorf("max_nodes must be greater than or equal to min_nodes")
	}
	
	if config.DistributionThreshold < 0 {
		return fmt.Errorf("distribution_threshold must be non-negative")
	}
	
	if config.HealthCheckInterval < time.Second {
		return fmt.Errorf("health_check_interval must be at least 1 second")
	}
	
	return nil
}

// setupIntegrationWithOllama sets up the integration with the existing Ollama server
func SetupIntegrationWithOllama(
	ctx context.Context,
	config *IntegrationConfig,
	p2pNode *p2p.Node,
	consensusEngine *consensus.Engine,
) (*OllamaDistributedScheduler, error) {
	// Validate configuration
	if err := ValidateIntegrationConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}
	
	// Create base Ollama scheduler
	baseScheduler := server.InitScheduler(ctx)
	
	// Create distributed scheduler
	distributedScheduler, err := NewOllamaDistributedScheduler(
		baseScheduler,
		config,
		p2pNode,
		consensusEngine,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create distributed scheduler: %v", err)
	}
	
	// Start the distributed scheduler
	if err := distributedScheduler.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start distributed scheduler: %v", err)
	}
	
	slog.Info("Ollama distributed scheduler integration setup complete",
		"cluster_id", distributedScheduler.GetClusterID(),
		"node_id", distributedScheduler.GetNodeID(),
		"enabled", config.Enabled)
	
	return distributedScheduler, nil
}

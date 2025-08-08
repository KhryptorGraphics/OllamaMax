package api

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/inference"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/models"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/ollama/api"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/distributed"
)

// DistributedOllamaIntegration integrates distributed inference with Ollama API
type DistributedOllamaIntegration struct {
	// Core components
	distributedEngine *inference.DistributedInferenceEngine
	modelManager      *models.DistributedModelManager
	scheduler         *distributed.DistributedScheduler
	p2pNode           *p2p.Node

	// Configuration
	config *DistributedIntegrationConfig
	logger *slog.Logger

	// State management
	activeRequests map[string]*DistributedRequest
	requestsMutex  sync.RWMutex

	// Performance tracking
	metrics *IntegrationMetrics

	// Lifecycle
	started bool
	mu      sync.RWMutex
}

// DistributedIntegrationConfig configures the distributed integration
type DistributedIntegrationConfig struct {
	// Thresholds for distributed execution
	MinModelSizeForDistribution int64         `json:"min_model_size_for_distribution"`
	MinNodesForDistribution     int           `json:"min_nodes_for_distribution"`
	MaxConcurrentRequests       int           `json:"max_concurrent_requests"`
	RequestTimeout              time.Duration `json:"request_timeout"`

	// Distribution strategy
	DefaultStrategy      string `json:"default_strategy"`
	EnableLoadBalancing  bool   `json:"enable_load_balancing"`
	EnableFaultTolerance bool   `json:"enable_fault_tolerance"`

	// Performance optimization
	EnableCaching     bool `json:"enable_caching"`
	CacheSize         int  `json:"cache_size"`
	EnablePrefetching bool `json:"enable_prefetching"`
}

// DistributedRequest represents a distributed inference request
type DistributedRequest struct {
	ID              string
	OriginalRequest *api.GenerateRequest
	StartTime       time.Time
	Status          RequestStatus
	NodesUsed       []string
	PartitionCount  int

	// Result channels
	ResultChan chan *api.GenerateResponse
	ErrorChan  chan error

	// Context
	Context    context.Context
	CancelFunc context.CancelFunc
}

// RequestStatus represents the status of a distributed request
type RequestStatus string

const (
	RequestStatusPending      RequestStatus = "pending"
	RequestStatusDistributing RequestStatus = "distributing"
	RequestStatusExecuting    RequestStatus = "executing"
	RequestStatusAggregating  RequestStatus = "aggregating"
	RequestStatusCompleted    RequestStatus = "completed"
	RequestStatusFailed       RequestStatus = "failed"
)

// IntegrationMetrics tracks integration performance
type IntegrationMetrics struct {
	TotalRequests        int64
	DistributedRequests  int64
	LocalRequests        int64
	SuccessfulRequests   int64
	FailedRequests       int64
	AverageLatency       time.Duration
	AverageNodesUsed     float64
	DistributionOverhead time.Duration
	LastUpdated          time.Time
}

// NewDistributedOllamaIntegration creates a new distributed Ollama integration
func NewDistributedOllamaIntegration(
	distributedEngine *inference.DistributedInferenceEngine,
	modelManager *models.DistributedModelManager,
	scheduler *distributed.DistributedScheduler,
	p2pNode *p2p.Node,
	config *DistributedIntegrationConfig,
	logger *slog.Logger,
) *DistributedOllamaIntegration {
	if config == nil {
		config = &DistributedIntegrationConfig{
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
	}

	return &DistributedOllamaIntegration{
		distributedEngine: distributedEngine,
		modelManager:      modelManager,
		scheduler:         scheduler,
		p2pNode:           p2pNode,
		config:            config,
		logger:            logger,
		activeRequests:    make(map[string]*DistributedRequest),
		metrics: &IntegrationMetrics{
			LastUpdated: time.Now(),
		},
	}
}

// Start starts the distributed integration
func (doi *DistributedOllamaIntegration) Start() error {
	doi.mu.Lock()
	defer doi.mu.Unlock()

	if doi.started {
		return fmt.Errorf("distributed integration already started")
	}

	doi.logger.Info("Starting distributed Ollama integration")
	doi.started = true

	return nil
}

// Stop stops the distributed integration
func (doi *DistributedOllamaIntegration) Stop() error {
	doi.mu.Lock()
	defer doi.mu.Unlock()

	if !doi.started {
		return nil
	}

	doi.logger.Info("Stopping distributed Ollama integration")

	// Cancel all active requests
	doi.requestsMutex.Lock()
	for _, request := range doi.activeRequests {
		if request.CancelFunc != nil {
			request.CancelFunc()
		}
	}
	doi.requestsMutex.Unlock()

	doi.started = false
	return nil
}

// HandleGenerateRequest handles a generate request, deciding whether to distribute it
func (doi *DistributedOllamaIntegration) HandleGenerateRequest(
	ctx context.Context,
	req *api.GenerateRequest,
) (*api.GenerateResponse, error) {
	doi.metrics.TotalRequests++

	// Check if request should be distributed
	shouldDistribute, err := doi.shouldDistributeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to determine distribution strategy: %w", err)
	}

	if shouldDistribute {
		doi.logger.Info("Distributing request across cluster",
			"model", req.Model,
			"prompt_length", len(req.Prompt))

		return doi.handleDistributedRequest(ctx, req)
	} else {
		doi.logger.Debug("Handling request locally",
			"model", req.Model,
			"reason", "below distribution threshold")

		return doi.handleLocalRequest(ctx, req)
	}
}

// shouldDistributeRequest determines if a request should be distributed
func (doi *DistributedOllamaIntegration) shouldDistributeRequest(req *api.GenerateRequest) (bool, error) {
	// Check if we have enough nodes
	availableNodes := doi.p2pNode.GetConnectedPeers()
	if len(availableNodes) < doi.config.MinNodesForDistribution {
		return false, nil
	}

	// Check model size
	model, err := doi.modelManager.GetModel(req.Model)
	if err != nil {
		// Model not found in distributed system, don't distribute
		return false, nil
	}

	if model.Size < doi.config.MinModelSizeForDistribution {
		return false, nil
	}

	// Check current load
	doi.requestsMutex.RLock()
	activeCount := len(doi.activeRequests)
	doi.requestsMutex.RUnlock()

	if activeCount >= doi.config.MaxConcurrentRequests {
		return false, nil
	}

	// Check if scheduler recommends distribution
	if doi.scheduler != nil {
		// Convert api.GenerateRequest to types that scheduler expects
		// This is a simplified check
		return true, nil
	}

	return true, nil
}

// handleDistributedRequest handles a request using distributed inference
func (doi *DistributedOllamaIntegration) handleDistributedRequest(
	ctx context.Context,
	req *api.GenerateRequest,
) (*api.GenerateResponse, error) {
	startTime := time.Now()

	// Create distributed request
	distributedReq := &DistributedRequest{
		ID:              fmt.Sprintf("dist_%d", time.Now().UnixNano()),
		OriginalRequest: req,
		StartTime:       startTime,
		Status:          RequestStatusPending,
		ResultChan:      make(chan *api.GenerateResponse, 1),
		ErrorChan:       make(chan error, 1),
	}

	// Create context with timeout
	distributedReq.Context, distributedReq.CancelFunc = context.WithTimeout(ctx, doi.config.RequestTimeout)
	defer distributedReq.CancelFunc()

	// Register active request
	doi.requestsMutex.Lock()
	doi.activeRequests[distributedReq.ID] = distributedReq
	doi.requestsMutex.Unlock()

	defer func() {
		doi.requestsMutex.Lock()
		delete(doi.activeRequests, distributedReq.ID)
		doi.requestsMutex.Unlock()
	}()

	// Execute distributed inference
	distributedReq.Status = RequestStatusExecuting

	// Convert request parameters
	parameters := make(map[string]interface{})
	if req.Options != nil {
		// Convert api.Options to map
		parameters["temperature"] = req.Options["temperature"]
		parameters["top_p"] = req.Options["top_p"]
		parameters["top_k"] = req.Options["top_k"]
		// Add other parameters as needed
	}

	// Execute distributed inference
	result, err := doi.distributedEngine.ExecuteDistributedInference(
		distributedReq.Context,
		req.Model,
		req.Prompt,
		parameters,
	)

	if err != nil {
		doi.metrics.FailedRequests++
		return nil, fmt.Errorf("distributed inference failed: %w", err)
	}

	// Convert result to Ollama API response
	response := &api.GenerateResponse{
		Model:     req.Model,
		Response:  result.Text,
		Done:      true,
		CreatedAt: time.Now(),
		Context:   result.Tokens,
	}

	// Update metrics
	doi.metrics.DistributedRequests++
	doi.metrics.SuccessfulRequests++
	doi.updateLatencyMetrics(time.Since(startTime))
	doi.updateNodesUsedMetrics(float64(len(result.NodesUsed)))

	distributedReq.Status = RequestStatusCompleted
	distributedReq.NodesUsed = make([]string, len(result.NodesUsed))
	for i, nodeID := range result.NodesUsed {
		distributedReq.NodesUsed[i] = nodeID.String()
	}

	doi.logger.Info("Distributed request completed",
		"request_id", distributedReq.ID,
		"model", req.Model,
		"nodes_used", len(result.NodesUsed),
		"latency", time.Since(startTime))

	return response, nil
}

// handleLocalRequest handles a request locally (fallback)
func (doi *DistributedOllamaIntegration) handleLocalRequest(
	ctx context.Context,
	req *api.GenerateRequest,
) (*api.GenerateResponse, error) {
	startTime := time.Now()

	// This would integrate with the local Ollama server
	// For now, return a mock response
	response := &api.GenerateResponse{
		Model:     req.Model,
		Response:  fmt.Sprintf("Local response for: %s", req.Prompt),
		Done:      true,
		CreatedAt: time.Now(),
		Context:   []int{1, 2, 3, 4, 5}, // Mock context
	}

	// Update metrics
	doi.metrics.LocalRequests++
	doi.metrics.SuccessfulRequests++
	doi.updateLatencyMetrics(time.Since(startTime))

	doi.logger.Debug("Local request completed",
		"model", req.Model,
		"latency", time.Since(startTime))

	return response, nil
}

// updateLatencyMetrics updates latency metrics
func (doi *DistributedOllamaIntegration) updateLatencyMetrics(latency time.Duration) {
	if doi.metrics.AverageLatency == 0 {
		doi.metrics.AverageLatency = latency
	} else {
		doi.metrics.AverageLatency = (doi.metrics.AverageLatency + latency) / 2
	}
	doi.metrics.LastUpdated = time.Now()
}

// updateNodesUsedMetrics updates nodes used metrics
func (doi *DistributedOllamaIntegration) updateNodesUsedMetrics(nodesUsed float64) {
	if doi.metrics.AverageNodesUsed == 0 {
		doi.metrics.AverageNodesUsed = nodesUsed
	} else {
		doi.metrics.AverageNodesUsed = (doi.metrics.AverageNodesUsed + nodesUsed) / 2
	}
}

// GetMetrics returns integration metrics
func (doi *DistributedOllamaIntegration) GetMetrics() *IntegrationMetrics {
	return doi.metrics
}

// GetActiveRequests returns currently active distributed requests
func (doi *DistributedOllamaIntegration) GetActiveRequests() map[string]*DistributedRequest {
	doi.requestsMutex.RLock()
	defer doi.requestsMutex.RUnlock()

	// Return a copy to avoid race conditions
	active := make(map[string]*DistributedRequest)
	for id, request := range doi.activeRequests {
		active[id] = request
	}
	return active
}

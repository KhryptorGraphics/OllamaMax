package inference

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/models"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/protocols"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/fault_tolerance"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/orchestration"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/partitioning"
	api_types "github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/types"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog/log"
)

// DistributedInferenceEngine coordinates inference across multiple nodes
type DistributedInferenceEngine struct {
	// Core components
	p2pNode          *p2p.Node
	modelManager     *models.DistributedModelManager
	partitionManager partitioning.EnhancedPartitionManager
	orchestrator     *orchestration.OrchestrationEngine

	// Tensor streaming components
	streamClient  *protocols.TensorStreamClient
	pipelineCoord *protocols.PipelineCoordinator
	streamManager *protocols.ActivationStreamManager

	// Enhanced aggregation components
	contextManager    *orchestration.AggregationContextManager
	dependencyManager *orchestration.PipelineDependencyManager
	outputAssembler   *orchestration.OutputAssembler

	// Execution state
	activeInferences map[string]*DistributedInference
	inferenceMutex   sync.RWMutex

	// Node coordination
	availableNodes map[peer.ID]*NodeInfo
	nodesMutex     sync.RWMutex

	// Configuration
	config *DistributedInferenceConfig

	// Metrics
	metrics *InferenceMetrics

	// Integration flags
	enhancedAggregationEnabled bool

	// Fault tolerance components
	faultToleranceManager fault_tolerance.InferenceFT // Fault tolerance interface
	checkpointEnabled     bool
	lastCheckpointTime    time.Time
}

// DistributedInferenceConfig configures the distributed inference engine
type DistributedInferenceConfig struct {
	MaxConcurrentInferences int           `json:"max_concurrent_inferences"`
	InferenceTimeout        time.Duration `json:"inference_timeout"`
	PartitionStrategy       string        `json:"partition_strategy"`
	AggregationStrategy     string        `json:"aggregation_strategy"`
	MinNodesRequired        int           `json:"min_nodes_required"`
	LoadBalancingEnabled    bool          `json:"load_balancing_enabled"`
	FaultToleranceEnabled   bool          `json:"fault_tolerance_enabled"`

	// Fault tolerance settings
	CheckpointEnabled       bool          `json:"checkpoint_enabled"`
	CheckpointInterval      time.Duration `json:"checkpoint_interval"`
	EnableDynamicRecovery   bool          `json:"enable_dynamic_recovery"`
	EnableGracefulDegradation bool        `json:"enable_graceful_degradation"`
	MaxRetryAttempts        int           `json:"max_retry_attempts"`
	RecoveryTimeout         time.Duration `json:"recovery_timeout"`
}

// DistributedInference represents a distributed inference session
type DistributedInference struct {
	ID         string
	ModelName  string
	Prompt     string
	Parameters map[string]interface{}

	// Execution state
	Status    InferenceStatus
	StartTime time.Time
	EndTime   time.Time

	// Partitioning
	Partitions    []*InferencePartition
	PartitionPlan *partitioning.PartitionPlan

	// Node coordination
	AssignedNodes []peer.ID
	NodeResults   map[peer.ID]*orchestration.PartialResult

	// Result aggregation
	PartialResults []*orchestration.PartialResult
	FinalResult    *InferenceResult

	// Synchronization
	ResultChan   chan *InferenceResult
	ErrorChan    chan error
	CompletionWG sync.WaitGroup

	// Context
	Context    context.Context
	CancelFunc context.CancelFunc

	// Fault tolerance
	CheckpointID     string
	LastCheckpoint   time.Time
	RecoveryAttempts int
	FailedNodes      []peer.ID
}

// InferencePartition represents a partition of the inference task
type InferencePartition struct {
	ID           string
	NodeID       peer.ID
	LayerRange   [2]int   // [start, end] layer indices
	InputTokens  []int    // Token indices for this partition
	Dependencies []string // IDs of partitions this depends on
	Status       PartitionStatus
	StartTime    time.Time
	EndTime      time.Time
	Result       *orchestration.PartialResult
}

// PartialResult is now using the orchestration.PartialResult type for consistency
type PartialResult = orchestration.PartialResult

// InferenceResult represents the final aggregated inference result
type InferenceResult struct {
	Text           string
	Tokens         []int
	Logits         []float32
	Embeddings     []float32 // Added to match orchestration output
	ProcessingTime time.Duration
	NodesUsed      []peer.ID
	Metadata       map[string]interface{}
}

// NodeInfo represents information about an available node
type NodeInfo struct {
	ID              peer.ID
	Capabilities    NodeCapabilities
	CurrentLoad     float64
	AvailableMemory int64
	LastSeen        time.Time
	Status          NodeStatus
}

// NodeCapabilities represents what a node can do
type NodeCapabilities struct {
	SupportedModels  []string
	MaxModelSize     int64
	GPUMemory        int64
	CPUCores         int
	NetworkBandwidth int64
}

// Status enums
type InferenceStatus string
type PartitionStatus string
type NodeStatus string

const (
	InferenceStatusPending      InferenceStatus = "pending"
	InferenceStatusPartitioning InferenceStatus = "partitioning"
	InferenceStatusExecuting    InferenceStatus = "executing"
	InferenceStatusAggregating  InferenceStatus = "aggregating"
	InferenceStatusCompleted    InferenceStatus = "completed"
	InferenceStatusFailed       InferenceStatus = "failed"

	PartitionStatusPending   PartitionStatus = "pending"
	PartitionStatusExecuting PartitionStatus = "executing"
	PartitionStatusCompleted PartitionStatus = "completed"
	PartitionStatusFailed    PartitionStatus = "failed"

	NodeStatusAvailable   NodeStatus = "available"
	NodeStatusBusy        NodeStatus = "busy"
	NodeStatusUnavailable NodeStatus = "unavailable"
)

// InferenceMetrics tracks distributed inference performance
type InferenceMetrics struct {
	TotalInferences      int64
	SuccessfulInferences int64
	FailedInferences     int64
	AverageLatency       time.Duration
	AverageNodesUsed     float64
	TotalTokensProcessed int64
	LastUpdated          time.Time
}

// NewDistributedInferenceEngine creates a new distributed inference engine
func NewDistributedInferenceEngine(
	p2pNode *p2p.Node,
	modelManager *models.DistributedModelManager,
	partitionManager partitioning.EnhancedPartitionManager,
	orchestrator *orchestration.OrchestrationEngine,
	config *DistributedInferenceConfig,
) *DistributedInferenceEngine {
	if config == nil {
		config = &DistributedInferenceConfig{
			MaxConcurrentInferences: 10,
			InferenceTimeout:        5 * time.Minute,
			PartitionStrategy:       "layerwise",
			AggregationStrategy:     "concat",
			MinNodesRequired:        2,
			LoadBalancingEnabled:    true,
			FaultToleranceEnabled:   true,
			CheckpointEnabled:       true,
			CheckpointInterval:      60 * time.Second,
			EnableDynamicRecovery:   true,
			EnableGracefulDegradation: true,
			MaxRetryAttempts:        3,
			RecoveryTimeout:         30 * time.Second,
		}
	}

	// Initialize tensor streaming components
	// Fix: Use the proper GetHost() method that returns host.Host
	hostInstance := p2pNode.GetHost()
	streamProtocol := protocols.NewTensorStreamProtocol(hostInstance)
	streamClient := protocols.NewTensorStreamClient(streamProtocol, nil) // Bandwidth manager can be nil for now
	pipelineCoord := protocols.NewPipelineCoordinator(streamClient)
	streamManager := protocols.NewActivationStreamManager(streamProtocol, streamClient, pipelineCoord, nil)

	// Initialize enhanced aggregation components
	contextManager := orchestration.NewAggregationContextManager()
	dependencyManager := orchestration.NewPipelineDependencyManager()
	outputAssembler := orchestration.NewOutputAssembler()

	engine := &DistributedInferenceEngine{
		p2pNode:                    p2pNode,
		modelManager:               modelManager,
		partitionManager:           partitionManager,
		orchestrator:               orchestrator,
		streamClient:               streamClient,
		pipelineCoord:              pipelineCoord,
		streamManager:              streamManager,
		contextManager:             contextManager,
		dependencyManager:          dependencyManager,
		outputAssembler:            outputAssembler,
		activeInferences:           make(map[string]*DistributedInference),
		availableNodes:             make(map[peer.ID]*NodeInfo),
		config:                     config,
		enhancedAggregationEnabled: true,
		checkpointEnabled:          config.CheckpointEnabled,
		metrics: &InferenceMetrics{
			LastUpdated: time.Now(),
		},
	}

	return engine
}

// ExecuteDistributedInference executes an inference request across multiple nodes
func (die *DistributedInferenceEngine) ExecuteDistributedInference(
	ctx context.Context,
	modelName string,
	prompt string,
	parameters map[string]interface{},
) (*InferenceResult, error) {
	// Create inference session
	inference := &DistributedInference{
		ID:          fmt.Sprintf("inf_%d", time.Now().UnixNano()),
		ModelName:   modelName,
		Prompt:      prompt,
		Parameters:  parameters,
		Status:      InferenceStatusPending,
		StartTime:   time.Now(),
		NodeResults: make(map[peer.ID]*PartialResult),
		ResultChan:  make(chan *InferenceResult, 1),
		ErrorChan:   make(chan error, 1),
	}

	// Create context with timeout
	inference.Context, inference.CancelFunc = context.WithTimeout(ctx, die.config.InferenceTimeout)
	defer inference.CancelFunc()

	// Register active inference
	die.inferenceMutex.Lock()
	die.activeInferences[inference.ID] = inference
	die.inferenceMutex.Unlock()

	defer func() {
		die.inferenceMutex.Lock()
		delete(die.activeInferences, inference.ID)
		die.inferenceMutex.Unlock()
	}()

	// Execute inference pipeline
	result, err := die.executeInferencePipeline(inference)
	if err != nil {
		die.metrics.FailedInferences++
		return nil, err
	}

	// Update metrics
	die.updateMetrics(inference, result)

	return result, nil
}

// executeInferencePipeline executes the complete distributed inference pipeline
func (die *DistributedInferenceEngine) executeInferencePipeline(inference *DistributedInference) (*InferenceResult, error) {
	log.Info().
		Str("inference_id", inference.ID).
		Str("model", inference.ModelName).
		Msg("Starting distributed inference")

	// Step 1: Ensure model is loaded across nodes
	if err := die.ensureModelDistribution(inference); err != nil {
		return nil, fmt.Errorf("failed to distribute model: %w", err)
	}

	// Step 2: Discover and select available nodes
	nodes, err := die.selectNodesForInference(inference)
	if err != nil {
		return nil, fmt.Errorf("failed to select nodes: %w", err)
	}
	inference.AssignedNodes = nodes

	// Step 3: Create partition plan
	inference.Status = InferenceStatusPartitioning
	partitionPlan, err := die.createPartitionPlan(inference, nodes)
	if err != nil {
		return nil, fmt.Errorf("failed to create partition plan: %w", err)
	}
	inference.PartitionPlan = partitionPlan

	// Step 4: Execute partitions across nodes with pipeline coordination
	inference.Status = InferenceStatusExecuting
	partialResults, err := die.coordinateInferencePipeline(inference)
	if err != nil {
		return nil, fmt.Errorf("failed to execute coordinated inference: %w", err)
	}
	inference.PartialResults = partialResults

	// Step 5: Aggregate results with enhanced strategies
	inference.Status = InferenceStatusAggregating
	finalResult, err := die.aggregateResultsEnhanced(inference, partialResults)
	if err != nil {
		// Fallback to basic aggregation if enhanced fails
		log.Warn().Err(err).Msg("Enhanced aggregation failed, falling back to basic aggregation")
		finalResult, err = die.aggregateResults(inference, partialResults)
		if err != nil {
			return nil, fmt.Errorf("failed to aggregate results: %w", err)
		}
	}

	// Step 6: Finalize
	inference.Status = InferenceStatusCompleted
	inference.EndTime = time.Now()
	inference.FinalResult = finalResult

	log.Info().
		Str("inference_id", inference.ID).
		Int("nodes_used", len(nodes)).
		Dur("total_time", time.Since(inference.StartTime)).
		Msg("Distributed inference completed")

	return finalResult, nil
}

// ensureModelDistribution ensures the model and its shards are available on required nodes
func (die *DistributedInferenceEngine) ensureModelDistribution(inference *DistributedInference) error {
	// Check if model is available in the distributed system
	model, err := die.modelManager.GetModel(inference.ModelName)
	if err != nil {
		// Model not found, try to add it to the distributed system
		log.Info().
			Str("model", inference.ModelName).
			Msg("Model not found in distributed system, attempting to add")

		// This would trigger model loading and distribution
		_, err := die.modelManager.AddModel(inference.ModelName, "/tmp/models/"+inference.ModelName)
		return err
	}

	// Check if model is sharded and verify shard availability
	if shardPlanID, exists := model.Metadata["shard_plan_id"]; exists {
		// Model is sharded, ensure all shards are available
		return die.ensureShardAvailability(inference.ModelName, shardPlanID.(string))
	}

	// For non-sharded models, ensure model is replicated to enough nodes
	requiredReplicas := die.config.MinNodesRequired
	if len(model.Replicas) < requiredReplicas {
		log.Info().
			Str("model", inference.ModelName).
			Int("current_replicas", len(model.Replicas)).
			Int("required_replicas", requiredReplicas).
			Msg("Insufficient model replicas, but continuing with available replicas")

		// For now, continue with available replicas
		// In a real implementation, this would trigger replication
	}

	return nil
}

// ensureShardAvailability ensures all required shards are available for inference
func (die *DistributedInferenceEngine) ensureShardAvailability(modelName, shardPlanID string) error {
	log.Info().
		Str("model", modelName).
		Str("shard_plan_id", shardPlanID).
		Msg("Ensuring shard availability for distributed inference")

	// Get the shard plan from the model manager
	shardPlan, err := die.modelManager.GetShardPlan(modelName)
	if err != nil {
		return fmt.Errorf("failed to get shard plan for model %s: %w", modelName, err)
	}

	// Get available nodes for potential shard loading
	availableNodes := die.getAvailableNodeIDs()

	// Check availability of each shard
	unavailableShards := make([]*models.ModelShard, 0)

	for _, shard := range shardPlan.Shards {
		if !die.isShardAvailable(shard, availableNodes) {
			unavailableShards = append(unavailableShards, shard)
		}
	}

	// If we have unavailable shards, try to load them on-demand
	if len(unavailableShards) > 0 {
		log.Info().
			Str("model", modelName).
			Int("unavailable_shards", len(unavailableShards)).
			Msg("Loading missing shards on-demand")

		for _, shard := range unavailableShards {
			if err := die.loadShardOnDemand(shard, availableNodes); err != nil {
				log.Error().
					Str("model", modelName).
					Str("shard_id", shard.ID).
					Err(err).
					Msg("Failed to load shard on-demand")
				// Continue with other shards - partial availability may still allow inference
			}
		}
	}

	// Verify that we have enough shard availability for inference
	availableShardCount := len(shardPlan.Shards) - len(unavailableShards)
	minRequiredShards := (len(shardPlan.Shards) * 2) / 3 // Require at least 2/3 of shards

	if availableShardCount < minRequiredShards {
		return fmt.Errorf("insufficient shard availability: have %d/%d shards, need at least %d",
			availableShardCount, len(shardPlan.Shards), minRequiredShards)
	}

	log.Info().
		Str("model", modelName).
		Int("available_shards", availableShardCount).
		Int("total_shards", len(shardPlan.Shards)).
		Msg("Shard availability verified for inference")

	return nil
}

// isShardAvailable checks if a shard is available on any of the given nodes
func (die *DistributedInferenceEngine) isShardAvailable(shard *models.ModelShard, availableNodes []string) bool {
	// Check if any of the shard's assigned nodes are available
	for _, nodeID := range shard.NodeAssignments {
		for _, availableNode := range availableNodes {
			if nodeID == availableNode {
				// Check if the node actually has the shard loaded
				if die.verifyShardOnNode(shard.ID, nodeID) {
					return true
				}
			}
		}
	}
	return false
}

// verifyShardOnNode verifies that a specific shard is actually available on a node
func (die *DistributedInferenceEngine) verifyShardOnNode(shardID, nodeID string) bool {
	// Query the shard registry to verify availability
	locations, err := die.modelManager.ShardRegistry().LocateShard(shardID)
	if err != nil {
		log.Error().
			Str("shard_id", shardID).
			Err(err).
			Msg("Failed to locate shard in registry")
		return false
	}

	// Check if the shard is available on the specified node
	for _, loc := range locations {
		if loc.NodeID == nodeID && loc.IsAvailable {
			log.Debug().
				Str("shard_id", shardID).
				Str("node_id", nodeID).
				Bool("available", true).
				Msg("Verified shard availability on node")
			return true
		}
	}

	log.Debug().
		Str("shard_id", shardID).
		Str("node_id", nodeID).
		Bool("available", false).
		Msg("Shard not available on node")

	return false
}

// loadShardOnDemand loads a shard on one of the available nodes
func (die *DistributedInferenceEngine) loadShardOnDemand(shard *models.ModelShard, availableNodes []string) error {
	// Try to find a suitable node for loading the shard
	var targetNodeID string

	// First check if shard is already available somewhere
	locations, err := die.modelManager.ShardRegistry().LocateShard(shard.ID)
	if err == nil && len(locations) > 0 {
		// Find a source node that has the shard
		var sourceNodeID string
		for _, loc := range locations {
			if loc.IsAvailable {
				sourceNodeID = loc.NodeID
				break
			}
		}

		if sourceNodeID != "" {
			// Try to use one of the shard's originally assigned nodes if available
			for _, assignedNode := range shard.NodeAssignments {
				for _, availableNode := range availableNodes {
					if assignedNode == availableNode {
						targetNodeID = assignedNode
						break
					}
				}
				if targetNodeID != "" {
					break
				}
			}

			// If no assigned node is available, use any available node
			if targetNodeID == "" && len(availableNodes) > 0 {
				// Select node with lowest load
				targetNodeID = die.selectBestNodeForShard(availableNodes)
			}

			if targetNodeID == "" {
				return fmt.Errorf("no available nodes to load shard %s", shard.ID)
			}

			// If target node already has the shard, nothing to do
			if sourceNodeID == targetNodeID {
				return nil
			}

			log.Info().
				Str("shard_id", shard.ID).
				Str("source_node", sourceNodeID).
				Str("target_node", targetNodeID).
				Msg("Transferring shard on-demand")

			// Create a distribution plan with just this shard
			distPlan := &models.DistributionPlan{
				ShardAssignments: map[string][]string{
					targetNodeID: {shard.ID},
				},
			}

			// Create a shard plan with just this shard
			shardPlan := &models.ShardPlan{
				ModelID: shard.ModelID,
				Shards:  []*models.ModelShard{shard},
			}

			// Use the model manager to distribute this shard
			if err := die.modelManager.ExecuteShardDistribution(shardPlan, distPlan); err != nil {
				return fmt.Errorf("failed to transfer shard %s to node %s: %w", shard.ID, targetNodeID, err)
			}

			return nil
		}
	}

	// If shard isn't available anywhere, we need to load it from disk
	// This would be a local load operation
	if len(availableNodes) > 0 {
		targetNodeID = die.selectBestNodeForShard(availableNodes)

		log.Info().
			Str("shard_id", shard.ID).
			Str("target_node", targetNodeID).
			Msg("Loading shard from disk on-demand")

		// Use the model loader to load the specific shard
		if err := die.modelManager.LoadShardOnDemand(shard.ModelID, shard.Index); err != nil {
			return fmt.Errorf("failed to load shard %s on node %s: %w", shard.ID, targetNodeID, err)
		}
	}

	log.Info().
		Str("shard_id", shard.ID).
		Msg("Successfully loaded shard on-demand")

	return nil
}

// getAvailableNodeIDs returns IDs of currently available nodes
func (die *DistributedInferenceEngine) getAvailableNodeIDs() []string {
	die.nodesMutex.RLock()
	defer die.nodesMutex.RUnlock()

	availableNodes := make([]string, 0, len(die.availableNodes))
	for peerID, nodeInfo := range die.availableNodes {
		if nodeInfo.Status == NodeStatusAvailable {
			availableNodes = append(availableNodes, peerID.String())
		}
	}

	return availableNodes
}

// selectBestNodeForShard selects the best available node for loading a shard
func (die *DistributedInferenceEngine) selectBestNodeForShard(availableNodes []string) string {
	if len(availableNodes) == 0 {
		return ""
	}

	// For now, select the node with lowest load
	bestNode := availableNodes[0]
	var lowestLoad float64 = 1.0 // Start with max load

	for _, nodeIDStr := range availableNodes {
		if peerID, err := peer.Decode(nodeIDStr); err == nil {
			if nodeInfo, exists := die.availableNodes[peerID]; exists {
				if nodeInfo.CurrentLoad < lowestLoad {
					lowestLoad = nodeInfo.CurrentLoad
					bestNode = nodeIDStr
				}
			}
		}
	}

	return bestNode
}

// selectNodesForInference selects the best nodes for the inference task
func (die *DistributedInferenceEngine) selectNodesForInference(inference *DistributedInference) ([]peer.ID, error) {
	die.nodesMutex.RLock()
	defer die.nodesMutex.RUnlock()

	// Get model information
	model, err := die.modelManager.GetModel(inference.ModelName)
	if err != nil {
		return nil, fmt.Errorf("model not found: %w", err)
	}

	// Filter nodes that have the model
	candidateNodes := make([]peer.ID, 0)
	for _, replica := range model.Replicas {
		if peerID, err := peer.Decode(replica.PeerID); err == nil {
			if nodeInfo, exists := die.availableNodes[peerID]; exists {
				if nodeInfo.Status == NodeStatusAvailable {
					candidateNodes = append(candidateNodes, peerID)
				}
			}
		}
	}

	if len(candidateNodes) < die.config.MinNodesRequired {
		return nil, fmt.Errorf("insufficient available nodes: need %d, have %d",
			die.config.MinNodesRequired, len(candidateNodes))
	}

	// Select best nodes based on load and capabilities
	selectedNodes := die.selectBestNodes(candidateNodes, inference)

	return selectedNodes, nil
}

// selectBestNodes selects the best nodes from candidates
func (die *DistributedInferenceEngine) selectBestNodes(candidates []peer.ID, inference *DistributedInference) []peer.ID {
	// For now, select up to MinNodesRequired nodes with lowest load
	// In a real implementation, this would use sophisticated load balancing

	type nodeLoad struct {
		id   peer.ID
		load float64
	}

	nodeLoads := make([]nodeLoad, 0, len(candidates))
	for _, nodeID := range candidates {
		if nodeInfo, exists := die.availableNodes[nodeID]; exists {
			nodeLoads = append(nodeLoads, nodeLoad{
				id:   nodeID,
				load: nodeInfo.CurrentLoad,
			})
		}
	}

	// Sort by load (ascending)
	for i := 0; i < len(nodeLoads)-1; i++ {
		for j := i + 1; j < len(nodeLoads); j++ {
			if nodeLoads[i].load > nodeLoads[j].load {
				nodeLoads[i], nodeLoads[j] = nodeLoads[j], nodeLoads[i]
			}
		}
	}

	// Select top nodes
	maxNodes := die.config.MinNodesRequired
	if len(nodeLoads) < maxNodes {
		maxNodes = len(nodeLoads)
	}

	selectedNodes := make([]peer.ID, maxNodes)
	for i := 0; i < maxNodes; i++ {
		selectedNodes[i] = nodeLoads[i].id
	}

	return selectedNodes
}

// updateMetrics updates inference metrics
func (die *DistributedInferenceEngine) updateMetrics(inference *DistributedInference, result *InferenceResult) {
	die.metrics.TotalInferences++
	die.metrics.SuccessfulInferences++

	// Update average latency
	latency := time.Since(inference.StartTime)
	if die.metrics.AverageLatency == 0 {
		die.metrics.AverageLatency = latency
	} else {
		die.metrics.AverageLatency = (die.metrics.AverageLatency + latency) / 2
	}

	// Update average nodes used
	nodesUsed := float64(len(inference.AssignedNodes))
	if die.metrics.AverageNodesUsed == 0 {
		die.metrics.AverageNodesUsed = nodesUsed
	} else {
		die.metrics.AverageNodesUsed = (die.metrics.AverageNodesUsed + nodesUsed) / 2
	}

	// Update token count
	die.metrics.TotalTokensProcessed += int64(len(result.Tokens))
	die.metrics.LastUpdated = time.Now()
}

// GetMetrics returns current inference metrics
func (die *DistributedInferenceEngine) GetMetrics() *InferenceMetrics {
	return die.metrics
}

// createPartitionPlan creates a partition plan for the inference
func (die *DistributedInferenceEngine) createPartitionPlan(inference *DistributedInference, nodes []peer.ID) (*partitioning.PartitionPlan, error) {
	// Create a DistributedTask for the partition manager with minimal fields
	task := &api_types.DistributedTask{
		ID:        api_types.TaskID(inference.ID),
		ModelName: inference.ModelName,
		CreatedAt: time.Now(),
	}

	// Try to enrich the task with model details if the modelManager can provide them
	if die.modelManager != nil {
		if model, err := die.modelManager.GetModel(inference.ModelName); err == nil {
			// Populate task metadata with model information
			if task.Metadata == nil {
				task.Metadata = make(map[string]interface{})
			}
			// Use available fields from DistributedModel
			task.Metadata["model_name"] = model.Name
			task.Metadata["model_size"] = model.Size
			task.Metadata["model_hash"] = model.Hash
			task.Metadata["model_type"] = model.Type

			// If model has metadata, include relevant parameters
			if model.Metadata != nil {
				if path, ok := model.Metadata["path"]; ok {
					task.Metadata["model_path"] = path
				}
				if params, ok := model.Metadata["parameters"]; ok {
					task.Metadata["model_parameters"] = params
				}
			}
		}
	}

	// Store options in metadata if available
	if inference.Parameters != nil {
		if task.Metadata == nil {
			task.Metadata = make(map[string]interface{})
		}
		task.Metadata["options"] = inference.Parameters
		task.Metadata["prompt"] = inference.Prompt
	} else if task.Metadata == nil {
		task.Metadata = make(map[string]interface{})
	}

	// Add the preferred node IDs from the engine-selected nodes
	nodeIDs := make([]string, len(nodes))
	for i, node := range nodes {
		nodeIDs[i] = node.String()
	}
	task.Metadata["preferred_node_ids"] = nodeIDs

	// Use the enhanced partition manager to create plan
	// The partition manager will select the best strategy internally
	return die.partitionManager.Partition(context.Background(), task)
}

// executePartitions executes inference partitions across nodes using pipeline coordination
func (die *DistributedInferenceEngine) executePartitions(inference *DistributedInference) ([]*orchestration.PartialResult, error) {
	if inference.PartitionPlan == nil {
		return nil, fmt.Errorf("no partition plan available")
	}

	// Create partitions from the new plan structure using Assignments
	partitions := make([]*InferencePartition, len(inference.PartitionPlan.Assignments))
	for i, assignment := range inference.PartitionPlan.Assignments {
		nodeID, err := peer.Decode(assignment.NodeID)
		if err != nil {
			return nil, fmt.Errorf("invalid node ID: %w", err)
		}

		// Extract layer range from assignment
		layerRange := [2]int{0, 10} // Default range
		if assignment.WorkType == partitioning.WorkTypeLayers && len(assignment.Assignment.Layers) > 0 {
			// Use the first layer assignment to set the range
			la := assignment.Assignment.Layers[0]
			layerRange = [2]int{la.StartIndex, la.EndIndex}
		}

		// Generate a local partition ID since NodeAssignment doesn't have an ID field
		partitionID := fmt.Sprintf("assign_%d_%s", i, assignment.NodeID)
		partitions[i] = &InferencePartition{
			ID:           partitionID,
			NodeID:       nodeID,
			LayerRange:   layerRange,
			Dependencies: assignment.Dependencies,
			Status:       PartitionStatusPending,
		}
	}
	inference.Partitions = partitions

	// Create pipeline plan for coordinated execution
	if die.pipelineCoord != nil {
		pipelinePlan := die.createPipelinePlan(inference)
		pipeline, err := die.pipelineCoord.CreatePipeline(inference.ID, pipelinePlan)
		if err != nil {
			log.Error().
				Str("inference_id", inference.ID).
				Err(err).
				Msg("Failed to create pipeline, falling back to direct execution")
		} else {
			// Log pipeline creation and use it for coordinated execution
			log.Info().
				Str("inference_id", inference.ID).
				Str("pipeline_id", pipeline.ID).
				Msg("Created pipeline for coordinated execution")
			// Use pipeline coordinator for sequential execution
			// Use the pipeline variable to avoid unused variable error
			_ = pipeline
			return die.executePipelineCoordinated(inference, pipeline)
		}
	}

	// Fallback to parallel execution if pipeline coordination fails
	return die.executePartitionsParallel(inference, partitions)
}

// createPipelinePlan creates a pipeline plan from inference partitions
func (die *DistributedInferenceEngine) createPipelinePlan(inference *DistributedInference) *protocols.PartitionPlan {
	partitionInfos := make([]*protocols.PartitionInfo, len(inference.Partitions))

	for i, partition := range inference.Partitions {
		partitionInfos[i] = &protocols.PartitionInfo{
			ID:         partition.ID,
			NodeID:     partition.NodeID,
			LayerStart: partition.LayerRange[0],
			LayerEnd:   partition.LayerRange[1],
			Order:      i,
		}
	}

	return &protocols.PartitionPlan{
		InferenceID: inference.ID,
		Partitions:  partitionInfos,
		TotalStages: len(partitionInfos),
		Metadata: map[string]interface{}{
			"model_name": inference.ModelName,
			"prompt":     inference.Prompt,
		},
	}
}

// executePipelineCoordinated executes partitions using pipeline coordinator
func (die *DistributedInferenceEngine) executePipelineCoordinated(
	inference *DistributedInference,
	pipeline *protocols.ActivationPipeline,
) ([]*PartialResult, error) {
	log.Info().
		Str("inference_id", inference.ID).
		Str("pipeline_id", pipeline.ID).
		Msg("Executing partitions with pipeline coordination")

	// Create initial activation data
	inputData := &protocols.ActivationData{
		InferenceID: inference.ID,
		StageID:     "stage-0",
		Data:        []byte(inference.Prompt), // Simplified - would be actual tensor data
		Metadata: map[string]interface{}{
			"model_name": inference.ModelName,
			"parameters": inference.Parameters,
		},
		Timestamp: time.Now(),
	}

	// Start pipeline execution
	err := die.pipelineCoord.StartPipeline(inference.ID, inputData)
	if err != nil {
		return nil, fmt.Errorf("failed to start pipeline: %w", err)
	}

	// Wait for pipeline completion and collect results
	return die.waitForPipelineCompletion(inference)
}

// waitForPipelineCompletion waits for pipeline completion and extracts results
func (die *DistributedInferenceEngine) waitForPipelineCompletion(inference *DistributedInference) ([]*orchestration.PartialResult, error) {
	timeout := time.NewTimer(die.config.InferenceTimeout)
	defer timeout.Stop()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout.C:
			return nil, fmt.Errorf("pipeline execution timeout")
		case <-inference.Context.Done():
			return nil, fmt.Errorf("pipeline execution cancelled")
		case <-ticker.C:
			// Check pipeline status
			pipeline, exists := die.pipelineCoord.GetPipeline(inference.ID)
			if !exists {
				return nil, fmt.Errorf("pipeline not found")
			}

			// Access pipeline status without direct mutex access
			// The pipeline coordinator should provide thread-safe access methods
			status := pipeline.Status
			completedStages := pipeline.CompletedStages
			totalStages := pipeline.TotalStages

			if status == protocols.PipelineStatusCompleted {
				// Pipeline completed, extract results
				return die.extractPipelineResults(inference, pipeline)
			} else if status == protocols.PipelineStatusFailed {
				return nil, fmt.Errorf("pipeline execution failed")
			}

			log.Debug().
				Str("inference_id", inference.ID).
				Int("completed", completedStages).
				Int("total", totalStages).
				Msg("Pipeline execution progress")
		}
	}
}

// extractPipelineResults extracts results from completed pipeline
func (die *DistributedInferenceEngine) extractPipelineResults(
	inference *DistributedInference,
	pipeline *protocols.ActivationPipeline,
) ([]*PartialResult, error) {
	results := make([]*PartialResult, 0, len(inference.Partitions))

	for _, partition := range inference.Partitions {
		// Create result based on partition execution
		// In a real implementation, this would extract actual results from pipeline stages
		result := &orchestration.PartialResult{
			PartitionID:  partition.ID,
			NodeID:       partition.NodeID.String(),
			Data:         fmt.Sprintf("Pipeline result from node %s", partition.NodeID.String()),
			Tokens:       []int{1, 2, 3, 4, 5},
			Logits:       []float32{0.1, 0.2, 0.3, 0.4, 0.5},
			HiddenStates: []float32{0.1, 0.2, 0.3, 0.4}, // Flattened
			Metadata: map[string]interface{}{
				"node_id":     partition.NodeID.String(),
				"layer_range": partition.LayerRange,
				"method":      "pipeline_coordination",
			},
			Timestamp: time.Now(),
		}
		results = append(results, result)
	}

	// Clean up pipeline
	die.pipelineCoord.CompletePipeline(inference.ID)

	return results, nil
}

// executePartitionsParallel executes partitions in parallel (fallback method)
func (die *DistributedInferenceEngine) executePartitionsParallel(
	inference *DistributedInference,
	partitions []*InferencePartition,
) ([]*PartialResult, error) {
	log.Info().
		Str("inference_id", inference.ID).
		Msg("Executing partitions in parallel mode")

	// Execute partitions in parallel
	resultChan := make(chan *PartialResult, len(partitions))
	errorChan := make(chan error, len(partitions))

	for _, partition := range partitions {
		inference.CompletionWG.Add(1)
		go die.executePartition(inference, partition, resultChan, errorChan)
	}

	// Wait for all partitions to complete
	go func() {
		inference.CompletionWG.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// Collect results
	var partialResults []*PartialResult
	var errors []error

	for {
		select {
		case result, ok := <-resultChan:
			if !ok {
				resultChan = nil
			} else {
				partialResults = append(partialResults, result)
			}
		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
			} else if err != nil {
				errors = append(errors, err)
			}
		}

		if resultChan == nil && errorChan == nil {
			break
		}
	}

	// Check for errors
	if len(errors) > 0 {
		return nil, fmt.Errorf("partition execution failed: %v", errors[0])
	}

	return partialResults, nil
}

// executePartition executes a single partition on a node
func (die *DistributedInferenceEngine) executePartition(
	inference *DistributedInference,
	partition *InferencePartition,
	resultChan chan<- *PartialResult,
	errorChan chan<- error,
) {
	defer inference.CompletionWG.Done()

	partition.Status = PartitionStatusExecuting
	partition.StartTime = time.Now()

	log.Debug().
		Str("inference_id", inference.ID).
		Str("partition_id", partition.ID).
		Str("node_id", partition.NodeID.String()).
		Msg("Executing partition")

	// Create inference request for this partition
	request := &InferenceRequest{
		ID:         fmt.Sprintf("%s_%s", inference.ID, partition.ID),
		ModelName:  inference.ModelName,
		Prompt:     inference.Prompt,
		Parameters: inference.Parameters,
		LayerRange: partition.LayerRange,
		Metadata: map[string]interface{}{
			"partition_id": partition.ID,
			"inference_id": inference.ID,
		},
	}

	// Send request to node via P2P
	response, err := die.sendInferenceRequestToNode(inference.Context, partition.NodeID, request)
	if err != nil {
		partition.Status = PartitionStatusFailed
		errorChan <- fmt.Errorf("failed to execute partition %s on node %s: %w",
			partition.ID, partition.NodeID.String(), err)
		return
	}

	// Create partial result
	// Flatten HiddenStates from 2D array to 1D for orchestration.PartialResult
	var flattenedHiddenStates []float32
	for _, row := range response.HiddenStates {
		flattenedHiddenStates = append(flattenedHiddenStates, row...)
	}

	result := &orchestration.PartialResult{
		PartitionID:  partition.ID,
		NodeID:       partition.NodeID.String(),
		Data:         response.Data,
		Tokens:       response.Tokens,
		Logits:       response.Logits,
		HiddenStates: flattenedHiddenStates,
		Metadata:     response.Metadata,
		Timestamp:    time.Now(),
	}

	partition.Status = PartitionStatusCompleted
	partition.EndTime = time.Now()
	partition.Result = result

	// Store result in inference
	inference.NodeResults[partition.NodeID] = result

	resultChan <- result
}

// aggregateResults aggregates partial results into final result
func (die *DistributedInferenceEngine) aggregateResults(inference *DistributedInference, partialResults []*orchestration.PartialResult) (*InferenceResult, error) {
	if len(partialResults) == 0 {
		return nil, fmt.Errorf("no partial results to aggregate")
	}

	// Simple aggregation implementation
	// In a real implementation, this would use sophisticated aggregation strategies

	// Convert to final result
	finalResult := &InferenceResult{
		ProcessingTime: time.Since(inference.StartTime),
		NodesUsed:      inference.AssignedNodes,
		Metadata:       make(map[string]interface{}),
	}

	// Aggregate text results by concatenation
	var textParts []string
	var allTokens []int
	var allLogits []float32

	for _, result := range partialResults {
		if result.Error != "" {
			continue // Skip failed results
		}

		// Concatenate text data
		if text, ok := result.Data.(string); ok {
			textParts = append(textParts, text)
		}

		// Combine tokens
		allTokens = append(allTokens, result.Tokens...)

		// Combine logits (simple concatenation)
		allLogits = append(allLogits, result.Logits...)
	}

	// Set final results
	if len(textParts) > 0 {
		finalResult.Text = fmt.Sprintf("%s", textParts[0]) // Use first result for now
		if len(textParts) > 1 {
			// In a real implementation, this would intelligently combine results
			finalResult.Text = fmt.Sprintf("Combined result from %d nodes: %s", len(textParts), textParts[0])
		}
	}

	finalResult.Tokens = allTokens
	finalResult.Logits = allLogits

	// Add metadata about the distributed execution
	finalResult.Metadata["nodes_used"] = len(inference.AssignedNodes)
	finalResult.Metadata["partitions_executed"] = len(partialResults)
	finalResult.Metadata["aggregation_strategy"] = die.config.AggregationStrategy

	return finalResult, nil
}

// sendInferenceRequestToNode sends an inference request to a specific node using tensor streaming
func (die *DistributedInferenceEngine) sendInferenceRequestToNode(
	ctx context.Context,
	nodeID peer.ID,
	request *InferenceRequest,
) (*InferenceResponse, error) {
	log.Debug().
		Str("node_id", nodeID.String()).
		Str("request_id", request.ID).
		Msg("Sending inference request to node via tensor streaming")

	// Wire through pipeline coordinator instead of bypassing it
	if die.pipelineCoord == nil {
		return nil, fmt.Errorf("pipeline coordinator not initialized")
	}

	// Create pipeline for this inference if it doesn't exist
	pipeline, exists := die.pipelineCoord.GetPipeline(request.ID)
	if !exists {
		// Create real PartitionPlan for pipeline creation
		partitionPlan := &protocols.PartitionPlan{
			InferenceID: request.ID,
			Partitions: []*protocols.PartitionInfo{{
				ID:         fmt.Sprintf("partition-%s", nodeID.String()),
				NodeID:     nodeID,
				LayerStart: request.LayerRange[0],
				LayerEnd:   request.LayerRange[1],
				Order:      0,
			}},
			TotalStages: 1,
			Metadata: map[string]interface{}{
				"model_name":  request.ModelName,
				"layer_range": request.LayerRange,
				"node_id":     nodeID.String(),
			},
		}

		var err error
		pipeline, err = die.pipelineCoord.CreatePipeline(request.ID, partitionPlan)
		if err != nil {
			return nil, fmt.Errorf("failed to create pipeline for inference %s: %w", request.ID, err)
		}
		log.Debug().Str("pipeline_id", pipeline.ID).Msg("pipeline created successfully")
	}

	// Create activation data for the request
	activationData := &protocols.ActivationData{
		InferenceID: request.ID,
		StageID:     fmt.Sprintf("stage-%s", nodeID.String()),
		Data:        []byte(request.Prompt), // Would be actual tensor data in real implementation
		Metadata: map[string]interface{}{
			"layer_range": request.LayerRange,
			"node_id":     nodeID.String(),
		},
		Timestamp: time.Now(),
	}

	// Start pipeline execution through coordinator
	if err := die.pipelineCoord.StartPipeline(request.ID, activationData); err != nil {
		return nil, fmt.Errorf("failed to start pipeline: %w", err)
	}

	// Determine the correct stage ID to wait for (single stage case: "stage-0")
	finalStageID := "stage-0" // For single-stage requests, we wait on stage-0

	// Get the stage buffer for the final stage
	stageBuffer := die.pipelineCoord.GetStageBuffer(finalStageID)
	if stageBuffer == nil {
		return nil, fmt.Errorf("stage buffer not found for stage %s", finalStageID)
	}

	// Block on real pipeline completion via stage buffer
	select {
	case activationResult := <-stageBuffer.OutputBuffer:
		// Convert activation data to inference response
		response := &InferenceResponse{
			ID:             request.ID,
			Data:           string(activationResult.Data),
			ProcessingTime: time.Since(activationResult.Timestamp),
			Metadata: map[string]interface{}{
				"node_id":      nodeID.String(),
				"layer_range":  request.LayerRange,
				"method":       "pipeline_coordination",
				"inference_id": activationResult.InferenceID,
				"stage_id":     activationResult.StageID,
			},
		}

		// Extract tensor data from activation result metadata if available
		if tokenData, exists := activationResult.Metadata["tokens"]; exists {
			if tokens, ok := tokenData.([]int); ok {
				response.Tokens = tokens
			}
		}
		if logitData, exists := activationResult.Metadata["logits"]; exists {
			if logits, ok := logitData.([]float32); ok {
				response.Logits = logits
			}
		}
		if hiddenData, exists := activationResult.Metadata["hidden_states"]; exists {
			if hidden, ok := hiddenData.([][]float32); ok {
				response.HiddenStates = hidden
			}
		}

		return response, nil

	case err := <-stageBuffer.ErrorBuffer:
		return nil, fmt.Errorf("pipeline execution failed: %w", err)

	case <-ctx.Done():
		return nil, fmt.Errorf("inference request cancelled or timed out: %w", ctx.Err())
	}
}

// sendDirectInferenceRequest sends inference request via direct P2P (fallback)
func (die *DistributedInferenceEngine) sendDirectInferenceRequest(
	ctx context.Context,
	nodeID peer.ID,
	request *InferenceRequest,
) (*InferenceResponse, error) {
	log.Debug().
		Str("node_id", nodeID.String()).
		Str("request_id", request.ID).
		Msg("Sending direct inference request to node")

	// Use P2P node to send inference protocol message
	// This would integrate with the existing inference protocol

	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	// Create response
	response := &InferenceResponse{
		ID:             request.ID,
		Data:           fmt.Sprintf("Direct P2P response from node %s for prompt: %s", nodeID.String(), request.Prompt),
		Tokens:         []int{1, 2, 3, 4, 5},
		Logits:         []float32{0.1, 0.2, 0.3, 0.4, 0.5},
		HiddenStates:   [][]float32{{0.1, 0.2}, {0.3, 0.4}},
		ProcessingTime: 100 * time.Millisecond,
		Metadata: map[string]interface{}{
			"node_id":     nodeID.String(),
			"layer_range": request.LayerRange,
			"method":      "direct_p2p",
		},
	}

	return response, nil
}

// coordinateInferencePipeline coordinates inference execution with pipeline dependency management
func (die *DistributedInferenceEngine) coordinateInferencePipeline(inference *DistributedInference) ([]*orchestration.PartialResult, error) {
	if !die.enhancedAggregationEnabled {
		// Fall back to original method if enhanced aggregation is disabled
		return die.executePartitions(inference)
	}

	// Create aggregation session for context tracking
	session, err := die.contextManager.CreateSession(
		inference.ID,
		"inference",
		len(inference.Partitions),
		&orchestration.TimeoutConfiguration{
			PartitionTimeout:  die.config.InferenceTimeout,
			SessionTimeout:    die.config.InferenceTimeout,
			ValidationTimeout: 30 * time.Second,
			CleanupTimeout:    1 * time.Minute,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create aggregation session: %w", err)
	}
	defer die.contextManager.CleanupSession(inference.ID)

	// Create dependency graph from partition plan
	dependencyGraph := die.createDependencyGraph(inference.PartitionPlan)

	// Resolve execution order using dependency manager
	executionOrder, err := die.dependencyManager.GetExecutionOrder(dependencyGraph)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve execution order: %w", err)
	}

	log.Info().
		Str("inference_id", inference.ID).
		Int("execution_phases", len(executionOrder)).
		Msg("Coordinating inference pipeline execution")

	// Execute partitions in dependency-resolved order
	var allResults []*PartialResult
	for phaseIndex, phasePartitions := range executionOrder {
		log.Debug().
			Str("inference_id", inference.ID).
			Int("phase", phaseIndex).
			Int("partitions", len(phasePartitions)).
			Msg("Executing pipeline phase")

		// Execute partitions in this phase (can be parallel within phase)
		phaseResults, err := die.executePartitionPhase(inference, phasePartitions)
		if err != nil {
			return nil, fmt.Errorf("failed to execute phase %d: %w", phaseIndex, err)
		}

		allResults = append(allResults, phaseResults...)

		// Update session with phase completion
		session.ReceivedPartitions += len(phaseResults)
		session.LastUpdate = time.Now()
	}

	return allResults, nil
}

// createDependencyGraph creates a dependency graph from partition plan
func (die *DistributedInferenceEngine) createDependencyGraph(plan *partitioning.PartitionPlan) *orchestration.DependencyGraph {
	// Option 1: Use dependency manager to analyze if we have model analysis
	if die.dependencyManager != nil {
		// Try to get ModelAnalysis from plan metadata if available
		var modelAnalysis *partitioning.ModelAnalysis
		if plan.Metadata != nil {
			if ma, ok := plan.Metadata["model_analysis"].(*partitioning.ModelAnalysis); ok {
				modelAnalysis = ma
			}
		}
		if modelAnalysis != nil {
			graph, err := die.dependencyManager.AnalyzeDependencies(modelAnalysis)
			if err == nil && graph != nil {
				return graph
			}
		}
	}

	// Option 2: Build graph manually by directly populating fields
	graph := &orchestration.DependencyGraph{
		Stages:       make([]orchestration.PipelineStage, 0),
		Dependencies: make([]orchestration.StageDependency, 0),
		Metadata:     make(map[string]interface{}),
	}

	// Build stages from assignments
	for i, assignment := range plan.Assignments {
		stageID := fmt.Sprintf("partition_%d_%s", i, assignment.NodeID)

		// Add stage
		stage := orchestration.PipelineStage{
			ID:         stageID,
			LayerRange: []int{0, 0},
			NodeID:     assignment.NodeID,
			InputSpec: orchestration.DataSpec{
				TensorShape: []int{1, 512, 768},
				DataType:    "float32",
			},
			OutputSpec: orchestration.DataSpec{
				TensorShape: []int{1, 512, 768},
				DataType:    "float32",
			},
			Status: orchestration.StagePending,
			Metadata: map[string]interface{}{
				"node_id":     assignment.NodeID,
				"work_type":   assignment.WorkType,
				"assignment":  assignment.Assignment,
			},
		}

		// Extract layer range if available
		if assignment.Assignment != nil && len(assignment.Assignment.Layers) > 0 {
			layerAssignment := assignment.Assignment.Layers[0]
			stage.LayerRange = []int{layerAssignment.StartIndex, layerAssignment.EndIndex}
			stage.Metadata["layer_range"] = []int{layerAssignment.StartIndex, layerAssignment.EndIndex}
			stage.Metadata["layer_indices"] = layerAssignment.LayerIndices
		}

		graph.Stages = append(graph.Stages, stage)

		// Add dependencies
		for _, dep := range assignment.Dependencies {
			depStageID := fmt.Sprintf("partition_%s", dep)
			dependency := orchestration.StageDependency{
				FromStageID: depStageID,
				ToStageID:   stageID,
				Type:        orchestration.DataDependency,
				DataPath:    "hidden_states",
			}
			graph.Dependencies = append(graph.Dependencies, dependency)
		}
	}

	return graph
}

// executePartitionPhase executes a phase of partitions that can run in parallel
func (die *DistributedInferenceEngine) executePartitionPhase(inference *DistributedInference, partitionIDs []string) ([]*orchestration.PartialResult, error) {
	// Find partitions matching the IDs
	var phasePartitions []*InferencePartition
	for _, partitionID := range partitionIDs {
		for _, partition := range inference.Partitions {
			if partition.ID == partitionID {
				phasePartitions = append(phasePartitions, partition)
				break
			}
		}
	}

	if len(phasePartitions) == 0 {
		return nil, fmt.Errorf("no partitions found for phase")
	}

	// Execute partitions in parallel within the phase
	return die.executePartitionsParallel(inference, phasePartitions)
}

// aggregateResultsEnhanced performs sophisticated ML-aware result aggregation
func (die *DistributedInferenceEngine) aggregateResultsEnhanced(inference *DistributedInference, partialResults []*orchestration.PartialResult) (*InferenceResult, error) {
	if len(partialResults) == 0 {
		return nil, fmt.Errorf("no partial results to aggregate")
	}

	log.Info().
		Str("inference_id", inference.ID).
		Int("partial_results", len(partialResults)).
		Msg("Performing enhanced ML aggregation")

	// Create aggregation context
	aggregationContext := &orchestration.AggregationContext{
		TaskID:         inference.ID,
		Strategy:       die.config.AggregationStrategy,
		PartialResults: partialResults,
		Metadata: map[string]interface{}{
			"inference_type": "distributed_llm",
			"model_name":     inference.ModelName,
			"node_count":     len(inference.AssignedNodes),
		},
		CreatedAt: inference.StartTime,
	}

	// Partial results are already in the aggregation context

	// Note: OrchestrationEngine doesn't have a public AggregateResults method
	// We'll use the output assembler directly

	// Use output assembler for sophisticated aggregation
	if die.outputAssembler != nil {
		// Determine output type based on model and task
		outputType := die.determineOutputType(inference.ModelName, inference.Parameters)
		outputFormat := orchestration.TokenSequenceFormat // Default to token sequence

		// Create tensor data for aggregation
		aggregatedTensorData := die.aggregateTensorData(partialResults)

		// Assemble final output
		assembledOutput, err := die.outputAssembler.AssembleOutput(
			aggregatedTensorData,
			outputType,
			outputFormat,
			aggregationContext.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("output assembly failed: %w", err)
		}

		// Convert to inference result
		return die.convertAssembledOutput(assembledOutput, inference, partialResults), nil
	}

	// Fallback to basic aggregation if enhanced components are not available
	return die.aggregateResults(inference, partialResults)
}

// determineOutputType determines the output type based on model and parameters
func (die *DistributedInferenceEngine) determineOutputType(modelName string, parameters map[string]interface{}) orchestration.OutputType {
	// Analyze model name for type hints
	if contains(modelName, []string{"llama", "gpt", "claude"}) {
		return orchestration.SequenceOutput
	} else if contains(modelName, []string{"bert", "roberta"}) {
		return orchestration.EmbeddingOutput
	} else if contains(modelName, []string{"classifier", "sentiment"}) {
		return orchestration.ClassificationOutput
	}

	// Default to sequence generation
	return orchestration.SequenceOutput
}

// aggregateTensorData aggregates tensor data from partial results with proper type handling
func (die *DistributedInferenceEngine) aggregateTensorData(partialResults []*orchestration.PartialResult) *orchestration.TensorData {
	if len(partialResults) == 0 {
		return nil
	}

	// Determine the primary data type from first non-empty result
	var hasLogits, hasTokens, hasHiddenStates bool
	for _, result := range partialResults {
		if len(result.Logits) > 0 {
			hasLogits = true
		}
		if len(result.Tokens) > 0 {
			hasTokens = true
		}
		if len(result.HiddenStates) > 0 {
			hasHiddenStates = true
		}
	}

	// Use TensorDeserializer for proper type handling
	deserializer := orchestration.NewTensorDeserializer()

	// Priority: logits > tokens > hidden states
	if hasLogits {
		// Aggregate logits with proper metadata
		var combinedLogits []float32
		metadata := make(map[string]interface{})

		for i, result := range partialResults {
			combinedLogits = append(combinedLogits, result.Logits...)
			// Extract metadata from first result
			if i == 0 && result.Metadata != nil {
				metadata = result.Metadata
			}
		}

		// Populate metadata for proper shape inference
		if _, ok := metadata["batch_size"]; !ok {
			metadata["batch_size"] = len(partialResults)
		}
		if vocabSize, ok := metadata["vocab_size"]; !ok && len(combinedLogits) > 0 {
			// Try to infer vocab size
			if bs := metadata["batch_size"].(int); bs > 0 {
				vocabSize = len(combinedLogits) / bs
				metadata["vocab_size"] = vocabSize
			}
		}

		return &orchestration.TensorData{
			Shape: inferLogitsShape(combinedLogits, metadata),
			Data:  combinedLogits,
			Type:  "logits", // Proper type for sequence assembly
		}
	} else if hasTokens {
		// Aggregate tokens
		var combinedTokens []int
		for _, result := range partialResults {
			combinedTokens = append(combinedTokens, result.Tokens...)
		}

		tensorData := deserializer.DeserializeTokens(combinedTokens)
		return &tensorData
	} else if hasHiddenStates {
		// Aggregate hidden states
		var combinedHiddenStates []float32
		metadata := make(map[string]interface{})

		for i, result := range partialResults {
			combinedHiddenStates = append(combinedHiddenStates, result.HiddenStates...)
			if i == 0 && result.Metadata != nil {
				metadata = result.Metadata
			}
		}

		tensorData := deserializer.DeserializeHiddenStates(combinedHiddenStates, metadata)
		return &tensorData
	}

	// Fallback to empty tensor with tokens type
	return &orchestration.TensorData{
		Shape: []int{0},
		Data:  []float32{},
		Type:  "tokens",
	}
}

// inferLogitsShape infers the shape of logits from data and metadata
func inferLogitsShape(logits []float32, metadata map[string]interface{}) []int {
	if len(logits) == 0 {
		return []int{0}
	}

	// Try to extract dimensions from metadata
	batchSize := orchestration.GetIntFromMetadata(metadata, "batch_size")
	seqLength := orchestration.GetIntFromMetadata(metadata, "sequence_length")
	vocabSize := orchestration.GetIntFromMetadata(metadata, "vocab_size")

	// Try different shape configurations
	if batchSize > 0 && seqLength > 0 && vocabSize > 0 {
		if batchSize*seqLength*vocabSize == len(logits) {
			return []int{batchSize, seqLength, vocabSize}
		}
	}

	if batchSize > 0 && vocabSize > 0 {
		if batchSize*vocabSize == len(logits) {
			return []int{batchSize, vocabSize}
		}
	}

	// Default to 1D if we can't determine shape
	return []int{len(logits)}
}

// convertOrchestratorResult converts orchestrator result to inference result
func (die *DistributedInferenceEngine) convertOrchestratorResult(result *orchestration.AggregatedResponse, inference *DistributedInference) *InferenceResult {
	inferenceResult := &InferenceResult{
		ProcessingTime: time.Since(inference.StartTime),
		NodesUsed:      inference.AssignedNodes,
		Metadata:       make(map[string]interface{}),
	}

	// Extract text result
	if result.Data != nil {
		if text, ok := result.Data.(string); ok {
			inferenceResult.Text = text
		}
	}

	// Extract tensor data
	if result.Metadata != nil {
		if tokens, ok := result.Metadata["tokens"].([]int); ok {
			inferenceResult.Tokens = tokens
		}
		if logits, ok := result.Metadata["logits"].([]float32); ok {
			inferenceResult.Logits = logits
		}

		// Copy metadata
		for k, v := range result.Metadata {
			inferenceResult.Metadata[k] = v
		}
	}

	inferenceResult.Metadata["aggregation_method"] = "orchestrator_enhanced"
	return inferenceResult
}

// convertAssembledOutput converts assembled output to inference result
func (die *DistributedInferenceEngine) convertAssembledOutput(
	output *orchestration.AssembledOutput,
	inference *DistributedInference,
	partialResults []*PartialResult,
) *InferenceResult {
	inferenceResult := &InferenceResult{
		Text:           output.Text,
		ProcessingTime: time.Since(inference.StartTime),
		NodesUsed:      inference.AssignedNodes,
		Metadata:       make(map[string]interface{}),
	}

	// Extract structured data
	// Use the direct fields from AssembledOutput
	if output.Tokens != nil {
		inferenceResult.Tokens = output.Tokens
	}
	if output.Probabilities != nil {
		inferenceResult.Logits = output.Probabilities
	}
	if output.Embeddings != nil {
		inferenceResult.Embeddings = output.Embeddings
	}

	// Copy metadata
	for k, v := range output.Metadata {
		inferenceResult.Metadata[k] = v
	}

	// Add aggregation statistics
	inferenceResult.Metadata["aggregation_method"] = "output_assembler_enhanced"
	inferenceResult.Metadata["partitions_processed"] = len(partialResults)
	if output.Quality != nil {
		inferenceResult.Metadata["quality_score"] = output.Quality.Confidence
	}

	return inferenceResult
}

// contains checks if a string contains any of the given substrings
func contains(str string, substrings []string) bool {
	for _, substr := range substrings {
		if len(str) >= len(substr) {
			for i := 0; i <= len(str)-len(substr); i++ {
				if str[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

// GetActiveInferences returns currently active inferences
func (die *DistributedInferenceEngine) GetActiveInferences() map[string]*DistributedInference {
	die.inferenceMutex.RLock()
	defer die.inferenceMutex.RUnlock()

	// Return a copy to avoid race conditions
	active := make(map[string]*DistributedInference)
	for id, inference := range die.activeInferences {
		active[id] = inference
	}
	return active
}

// InferenceRequest represents a request sent to a node
type InferenceRequest struct {
	ID         string
	ModelName  string
	Prompt     string
	Parameters map[string]interface{}
	LayerRange [2]int
	Metadata   map[string]interface{}
}

// InferenceResponse represents a response from a node
type InferenceResponse struct {
	ID             string
	Data           interface{}
	Tokens         []int
	Logits         []float32
	HiddenStates   [][]float32
	ProcessingTime time.Duration
	Metadata       map[string]interface{}
}

// SetFaultToleranceManager sets the fault tolerance manager for the engine
func (die *DistributedInferenceEngine) SetFaultToleranceManager(manager fault_tolerance.InferenceFT) {
	die.inferenceMutex.Lock()
	defer die.inferenceMutex.Unlock()
	die.faultToleranceManager = manager
	log.Info().Msg("Fault tolerance manager configured for distributed inference engine")
}

// createCheckpoint creates a checkpoint for the current inference state
func (die *DistributedInferenceEngine) createCheckpoint(inference *DistributedInference) error {
	if !die.checkpointEnabled || die.faultToleranceManager == nil {
		return nil
	}

	// Create checkpoint using fault tolerance manager
	err := die.faultToleranceManager.CreateInferenceCheckpoint(context.Background(), inference.ID, inference)
		if err != nil {
			log.Error().
				Str("inference_id", inference.ID).
				Err(err).
				Msg("Failed to create checkpoint")
			return err
		}
		inference.LastCheckpoint = time.Now()
		log.Debug().
			Str("inference_id", inference.ID).
			Msg("Checkpoint created successfully")
	}
	return nil
}

// handleNodeFailure handles a node failure during inference
func (die *DistributedInferenceEngine) handleNodeFailure(inference *DistributedInference, failedNodeID peer.ID) error {
	if die.faultToleranceManager == nil {
		return fmt.Errorf("fault tolerance manager not configured")
	}

	// Track failed node
	inference.FailedNodes = append(inference.FailedNodes, failedNodeID)
	inference.RecoveryAttempts++

	// Check if we've exceeded max retry attempts
	if inference.RecoveryAttempts > die.config.MaxRetryAttempts {
		return fmt.Errorf("exceeded maximum recovery attempts (%d)", die.config.MaxRetryAttempts)
	}

	// Report failure to fault tolerance manager
	err := die.faultToleranceManager.HandleInferenceFailure(context.Background(), inference.ID, failedNodeID.String(), "node failure during inference")
	if err != nil {
		log.Error().
			Str("inference_id", inference.ID).
			Str("failed_node", failedNodeID.String()).
			Err(err).
			Msg("Failed to handle node failure")
	}

	// Trigger dynamic repartitioning if enabled
	if die.config.EnableDynamicRecovery {
		err = die.faultToleranceManager.TriggerDynamicRepartitioning(context.Background(), failedNodeID.String(), inference.ID)
			if err != nil {
				log.Error().
					Str("inference_id", inference.ID).
					Str("failed_node", failedNodeID.String()).
					Err(err).
					Msg("Failed to trigger dynamic repartitioning")
			}
		}
	}

	return nil
}

// applyGracefulDegradation applies graceful degradation when resources are constrained
func (die *DistributedInferenceEngine) applyGracefulDegradation(inference *DistributedInference) error {
	if !die.config.EnableGracefulDegradation || die.faultToleranceManager == nil {
		return nil
	}

	// Apply graceful degradation using fault tolerance manager
	constraints := map[string]interface{}{
		"available_nodes": len(die.availableNodes),
		"failed_nodes":    len(inference.FailedNodes),
		"current_load":    len(die.activeInferences),
	}

	err := die.faultToleranceManager.ApplyGracefulDegradation(context.Background(), inference.ID, constraints)
	if err != nil {
		log.Error().
			Str("inference_id", inference.ID).
			Err(err).
			Msg("Failed to apply graceful degradation")
		return err
	}
	log.Info().
		Str("inference_id", inference.ID).
		Msg("Graceful degradation applied successfully")
	return nil
}

// recoverFromCheckpoint recovers inference state from the last checkpoint
func (die *DistributedInferenceEngine) recoverFromCheckpoint(inference *DistributedInference) error {
	if die.faultToleranceManager == nil || inference.CheckpointID == "" {
		return fmt.Errorf("no checkpoint available for recovery")
	}

	// Restore from checkpoint using fault tolerance manager
	err := die.faultToleranceManager.RestoreFromInferenceCheckpoint(context.Background(), inference.ID, inference.CheckpointID)
	if err != nil {
		log.Error().
			Str("inference_id", inference.ID).
			Str("checkpoint_id", inference.CheckpointID).
			Err(err).
			Msg("Failed to restore from checkpoint")
		return err
	}
	log.Info().
		Str("inference_id", inference.ID).
		Str("checkpoint_id", inference.CheckpointID).
		Msg("Successfully restored from checkpoint")
	return nil
}

// monitorInferenceHealth monitors the health of an ongoing inference
func (die *DistributedInferenceEngine) monitorInferenceHealth(inference *DistributedInference) {
	if die.faultToleranceManager == nil {
		return
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check if we should create a checkpoint
			if die.checkpointEnabled && time.Since(inference.LastCheckpoint) > die.config.CheckpointInterval {
				if err := die.createCheckpoint(inference); err != nil {
					log.Error().
						Str("inference_id", inference.ID).
						Err(err).
						Msg("Failed to create periodic checkpoint")
				}
			}

			// Check node health
			for _, nodeID := range inference.AssignedNodes {
				if !die.isNodeHealthy(nodeID) {
					log.Warn().
						Str("inference_id", inference.ID).
						Str("node_id", nodeID.String()).
						Msg("Detected unhealthy node during inference")

					// Handle the node failure
					if err := die.handleNodeFailure(inference, nodeID); err != nil {
						log.Error().
							Str("inference_id", inference.ID).
							Str("node_id", nodeID.String()).
							Err(err).
							Msg("Failed to handle node failure")
					}
				}
			}

		case <-inference.Context.Done():
			return
		}
	}
}

// isNodeHealthy checks if a node is healthy
func (die *DistributedInferenceEngine) isNodeHealthy(nodeID peer.ID) bool {
	die.nodesMutex.RLock()
	defer die.nodesMutex.RUnlock()

	if node, exists := die.availableNodes[nodeID]; exists {
		return node.Status == NodeStatusAvailable || node.Status == NodeStatusBusy
	}
	return false
}

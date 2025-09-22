package models

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/partitioning"
)

// ShardingStrategy defines how models are split across nodes
type ShardingStrategy string

const (
	ShardingStrategyLayerWise  ShardingStrategy = "layer_wise"
	ShardingStrategyTensorWise ShardingStrategy = "tensor_wise"
	ShardingStrategyHybrid     ShardingStrategy = "hybrid"
	ShardingStrategyPipeline   ShardingStrategy = "pipeline"
)

// ModelShard represents a single shard of a distributed model
type ModelShard struct {
	ID           string                 `json:"id"`
	ModelID      string                 `json:"model_id"`
	Index        int                    `json:"index"`        // Shard index in sequence
	Offset       int64                  `json:"offset"`       // Byte offset in original model
	Size         int64                  `json:"size"`         // Size in bytes
	Checksum     string                 `json:"checksum"`     // SHA256 checksum
	LayerRange   []int                  `json:"layer_range"`  // [start, end] layers
	TensorNames  []string               `json:"tensor_names"` // Specific tensors in this shard
	NodeAssignments []string            `json:"node_assignments"`
	Replicas     int                    `json:"replicas"`
	CreatedAt    time.Time              `json:"created_at"`
	LastAccessed time.Time              `json:"last_accessed"`
	AccessCount  int64                  `json:"access_count"`
	Priority     int                    `json:"priority"` // Higher priority shards get more replicas
	Metadata     map[string]interface{} `json:"metadata"`
}

// ShardPlan defines the complete sharding strategy for a model
type ShardPlan struct {
	ID                string                       `json:"id"`
	ModelID           string                       `json:"model_id"`
	ModelName         string                       `json:"model_name"`
	Strategy          ShardingStrategy             `json:"strategy"`
	TotalShards       int                          `json:"total_shards"`
	Shards            []*ModelShard                `json:"shards"`
	MemoryRequirements map[string]int64            `json:"memory_requirements"` // NodeID -> memory needed
	NodeAssignments   map[string][]int             `json:"node_assignments"`    // NodeID -> shard indices
	CommunicationTopology map[string][]string      `json:"communication_topology"` // Node communication patterns
	ReplicationFactor int                          `json:"replication_factor"`
	CreatedAt         time.Time                    `json:"created_at"`
	OptimizationHints map[string]interface{}      `json:"optimization_hints"`
	EstimatedTransferTime time.Duration            `json:"estimated_transfer_time"`
	TotalModelSize    int64                        `json:"total_model_size"`
}

// NodeCapabilities describes what a node can handle
type NodeCapabilities struct {
	NodeID          string  `json:"node_id"`
	AvailableMemory int64   `json:"available_memory"`
	TotalMemory     int64   `json:"total_memory"`
	GPUMemory       int64   `json:"gpu_memory"`
	CPUCores        int     `json:"cpu_cores"`
	GPUCount        int     `json:"gpu_count"`
	NetworkBandwidth int64  `json:"network_bandwidth"` // bytes/sec
	StorageCapacity int64   `json:"storage_capacity"`
	ComputeCapability float32 `json:"compute_capability"` // CUDA compute capability
	IsHealthy       bool    `json:"is_healthy"`
}

// ModelShardManager manages the lifecycle of model shards
type ModelShardManager struct {
	mu              sync.RWMutex
	shardPlans      map[string]*ShardPlan  // modelID -> ShardPlan
	activeShards    map[string]*ModelShard // shardID -> ModelShard
	nodeCapabilities map[string]*NodeCapabilities
	memoryOptimizer *partitioning.MemoryOptimizer
	modelAnalyzer   *partitioning.ModelAnalyzer
	config          *ShardingConfig
}

// ShardingConfig contains configuration for sharding operations
type ShardingConfig struct {
	DefaultStrategy      ShardingStrategy
	MaxShardSize         int64   // Maximum size per shard in bytes
	MinShardSize         int64   // Minimum size per shard in bytes
	ReplicationFactor    int     // Default replication factor
	MemorySafetyMargin   float64 // Reserve this fraction of memory (e.g., 0.1 = 10%)
	EnableCompression    bool
	CompressionRatio     float64 // Expected compression ratio
	EnableGradientCheckpointing bool
	MaxConcurrentTransfers int
	TransferChunkSize    int64
	VerifyChecksums      bool
}

// NewModelShardManager creates a new shard manager
func NewModelShardManager(config *ShardingConfig) *ModelShardManager {
	if config == nil {
		config = DefaultShardingConfig()
	}

	return &ModelShardManager{
		shardPlans:       make(map[string]*ShardPlan),
		activeShards:     make(map[string]*ModelShard),
		nodeCapabilities: make(map[string]*NodeCapabilities),
		memoryOptimizer:  partitioning.NewMemoryOptimizer(),
		modelAnalyzer:    partitioning.NewModelAnalyzer(),
		config:           config,
	}
}

// DefaultShardingConfig returns default sharding configuration
func DefaultShardingConfig() *ShardingConfig {
	return &ShardingConfig{
		DefaultStrategy:      ShardingStrategyLayerWise,
		MaxShardSize:         2 * 1024 * 1024 * 1024, // 2GB
		MinShardSize:         100 * 1024 * 1024,      // 100MB
		ReplicationFactor:    2,
		MemorySafetyMargin:   0.15, // 15% safety margin
		EnableCompression:    true,
		CompressionRatio:     0.7,
		EnableGradientCheckpointing: true,
		MaxConcurrentTransfers: 4,
		TransferChunkSize:    16 * 1024 * 1024, // 16MB chunks
		VerifyChecksums:      true,
	}
}

// CreateShardPlan analyzes model and node capabilities to create optimal shard distribution
func (m *ModelShardManager) CreateShardPlan(modelID, modelPath string, modelSize int64,
	nodeCapabilities []*NodeCapabilities) (*ShardPlan, error) {

	m.mu.Lock()
	defer m.mu.Unlock()

	// Update node capabilities
	for _, cap := range nodeCapabilities {
		m.nodeCapabilities[cap.NodeID] = cap
	}

	// Analyze model to determine sharding strategy
	strategy := m.determineStrategy(modelSize, nodeCapabilities)

	// Calculate number of shards based on model size and node capabilities
	numShards := m.calculateOptimalShards(modelSize, nodeCapabilities)

	// Create shard plan
	plan := &ShardPlan{
		ID:                fmt.Sprintf("plan-%s-%d", modelID, time.Now().Unix()),
		ModelID:           modelID,
		Strategy:          strategy,
		TotalShards:       numShards,
		Shards:            make([]*ModelShard, 0, numShards),
		MemoryRequirements: make(map[string]int64),
		NodeAssignments:   make(map[string][]int),
		CommunicationTopology: make(map[string][]string),
		ReplicationFactor: m.config.ReplicationFactor,
		CreatedAt:         time.Now(),
		TotalModelSize:    modelSize,
	}

	// Create individual shards
	shardSize := modelSize / int64(numShards)
	for i := 0; i < numShards; i++ {
		offset := int64(i) * shardSize
		size := shardSize

		// Adjust last shard size for remainder
		if i == numShards-1 {
			size = modelSize - offset
		}

		checksum, err := m.generateChecksum(modelPath, offset, size)
		if err != nil {
			return nil, fmt.Errorf("failed to generate checksum for shard %d: %w", i, err)
		}

		shard := &ModelShard{
			ID:        fmt.Sprintf("%s-shard-%d", modelID, i),
			ModelID:   modelID,
			Index:     i,
			Offset:    offset,
			Size:      size,
			Checksum:  checksum,
			Priority:  m.calculateShardPriority(i, numShards),
			CreatedAt: time.Now(),
			Metadata:  make(map[string]interface{}),
		}

		// Calculate layer range for layer-wise sharding
		if strategy == ShardingStrategyLayerWise {
			shard.LayerRange = m.calculateLayerRange(i, numShards)
		}

		plan.Shards = append(plan.Shards, shard)
		m.activeShards[shard.ID] = shard
	}

	// Assign shards to nodes
	if err := m.assignShardsToNodes(plan, nodeCapabilities); err != nil {
		return nil, fmt.Errorf("failed to assign shards to nodes: %w", err)
	}

	// Calculate communication topology
	m.calculateCommunicationTopology(plan)

	// Estimate transfer time
	plan.EstimatedTransferTime = m.estimateTransferTime(plan, nodeCapabilities)

	// Store the plan
	m.shardPlans[modelID] = plan

	return plan, nil
}

// ValidateShardPlan ensures memory constraints are met
func (m *ModelShardManager) ValidateShardPlan(plan *ShardPlan) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check each node's memory constraints
	for nodeID, shardIndices := range plan.NodeAssignments {
		cap, exists := m.nodeCapabilities[nodeID]
		if !exists {
			return fmt.Errorf("node %s not found in capabilities", nodeID)
		}

		totalMemoryNeeded := int64(0)
		for _, idx := range shardIndices {
			if idx >= len(plan.Shards) {
				return fmt.Errorf("invalid shard index %d", idx)
			}
			shard := plan.Shards[idx]

			// Account for model weights + activations + overhead
			memoryNeeded := shard.Size
			if m.config.EnableGradientCheckpointing {
				memoryNeeded = int64(float64(memoryNeeded) * 1.3) // 30% overhead for activations
			} else {
				memoryNeeded = int64(float64(memoryNeeded) * 1.5) // 50% overhead without checkpointing
			}

			totalMemoryNeeded += memoryNeeded
		}

		// Apply safety margin
		availableMemory := int64(float64(cap.AvailableMemory) * (1 - m.config.MemorySafetyMargin))

		if totalMemoryNeeded > availableMemory {
			return fmt.Errorf("node %s needs %d bytes but only has %d available",
				nodeID, totalMemoryNeeded, availableMemory)
		}

		plan.MemoryRequirements[nodeID] = totalMemoryNeeded
	}

	// Validate replication factor
	for _, shard := range plan.Shards {
		if len(shard.NodeAssignments) < plan.ReplicationFactor {
			return fmt.Errorf("shard %s has insufficient replicas: %d < %d",
				shard.ID, len(shard.NodeAssignments), plan.ReplicationFactor)
		}
	}

	return nil
}

// ExecuteShardPlan coordinates the actual distribution process
func (m *ModelShardManager) ExecuteShardPlan(plan *ShardPlan) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate the plan first
	if err := m.ValidateShardPlan(plan); err != nil {
		return fmt.Errorf("shard plan validation failed: %w", err)
	}

	// Mark shards as being distributed
	for _, shard := range plan.Shards {
		shard.Metadata["status"] = "distributing"
		shard.Metadata["started_at"] = time.Now()
	}

	// The actual distribution will be handled by ChunkTransferOrchestrator
	// This method sets up the plan for execution

	return nil
}

// determineStrategy selects the best sharding strategy based on model and cluster characteristics
func (m *ModelShardManager) determineStrategy(modelSize int64, nodes []*NodeCapabilities) ShardingStrategy {
	// Calculate total available memory
	totalMemory := int64(0)
	totalGPUs := 0
	for _, node := range nodes {
		totalMemory += node.AvailableMemory
		totalGPUs += node.GPUCount
	}

	// For very large models (>100GB), use hybrid strategy
	if modelSize > 100*1024*1024*1024 {
		return ShardingStrategyHybrid
	}

	// If we have many GPUs, use pipeline parallelism
	if totalGPUs > 8 {
		return ShardingStrategyPipeline
	}

	// For models that fit in memory with room to spare, use layer-wise
	if float64(modelSize)*1.5 < float64(totalMemory) {
		return ShardingStrategyLayerWise
	}

	// Otherwise use tensor-wise for fine-grained distribution
	return ShardingStrategyTensorWise
}

// calculateOptimalShards determines the optimal number of shards
func (m *ModelShardManager) calculateOptimalShards(modelSize int64, nodes []*NodeCapabilities) int {
	// Don't create shards smaller than MinShardSize
	maxShards := int(modelSize / m.config.MinShardSize)

	// Don't create shards larger than MaxShardSize
	minShards := int(math.Ceil(float64(modelSize) / float64(m.config.MaxShardSize)))

	// Consider number of nodes and their capabilities
	nodeCount := len(nodes)
	optimalShards := nodeCount * 2 // Aim for 2 shards per node for load balancing

	// Clamp to the valid range
	if optimalShards < minShards {
		optimalShards = minShards
	}
	if optimalShards > maxShards {
		optimalShards = maxShards
	}

	// Round to nearest power of 2 for better alignment
	return nearestPowerOf2(optimalShards)
}

// assignShardsToNodes distributes shards across available nodes
func (m *ModelShardManager) assignShardsToNodes(plan *ShardPlan, nodes []*NodeCapabilities) error {
	// Sort nodes by available memory (descending)
	sortedNodes := make([]*NodeCapabilities, len(nodes))
	copy(sortedNodes, nodes)
	sort.Slice(sortedNodes, func(i, j int) bool {
		return sortedNodes[i].AvailableMemory > sortedNodes[j].AvailableMemory
	})

	// Simple round-robin with memory awareness
	nodeIndex := 0
	for _, shard := range plan.Shards {
		assignedNodes := make([]string, 0, plan.ReplicationFactor)

		// Assign primary and replicas
		for replica := 0; replica < plan.ReplicationFactor; replica++ {
			// Find next suitable node
			assigned := false
			attempts := 0
			for !assigned && attempts < len(sortedNodes) {
				node := sortedNodes[(nodeIndex+attempts)%len(sortedNodes)]

				// Check if node has enough memory
				memoryNeeded := shard.Size
				if m.config.EnableGradientCheckpointing {
					memoryNeeded = int64(float64(memoryNeeded) * 1.3)
				} else {
					memoryNeeded = int64(float64(memoryNeeded) * 1.5)
				}

				currentUsage := plan.MemoryRequirements[node.NodeID]
				availableMemory := int64(float64(node.AvailableMemory) * (1 - m.config.MemorySafetyMargin))

				if currentUsage+memoryNeeded <= availableMemory {
					assignedNodes = append(assignedNodes, node.NodeID)
					plan.NodeAssignments[node.NodeID] = append(plan.NodeAssignments[node.NodeID], shard.Index)
					plan.MemoryRequirements[node.NodeID] = currentUsage + memoryNeeded
					assigned = true
					nodeIndex = (nodeIndex + attempts + 1) % len(sortedNodes)
				}

				attempts++
			}

			if !assigned {
				return fmt.Errorf("unable to assign shard %d replica %d: insufficient memory", shard.Index, replica)
			}
		}

		shard.NodeAssignments = assignedNodes
		shard.Replicas = len(assignedNodes)
	}

	return nil
}

// calculateCommunicationTopology determines optimal communication patterns
func (m *ModelShardManager) calculateCommunicationTopology(plan *ShardPlan) {
	// Build communication graph based on shard dependencies
	for nodeID, shardIndices := range plan.NodeAssignments {
		neighbors := make(map[string]bool)

		for _, shardIdx := range shardIndices {
			// For pipeline parallelism, communicate with nodes having adjacent shards
			if plan.Strategy == ShardingStrategyPipeline {
				// Previous shard
				if shardIdx > 0 {
					prevShard := plan.Shards[shardIdx-1]
					for _, prevNode := range prevShard.NodeAssignments {
						if prevNode != nodeID {
							neighbors[prevNode] = true
						}
					}
				}

				// Next shard
				if shardIdx < len(plan.Shards)-1 {
					nextShard := plan.Shards[shardIdx+1]
					for _, nextNode := range nextShard.NodeAssignments {
						if nextNode != nodeID {
							neighbors[nextNode] = true
						}
					}
				}
			}
		}

		// Convert map to slice
		topology := make([]string, 0, len(neighbors))
		for neighbor := range neighbors {
			topology = append(topology, neighbor)
		}

		plan.CommunicationTopology[nodeID] = topology
	}
}

// estimateTransferTime calculates expected time to distribute all shards
func (m *ModelShardManager) estimateTransferTime(plan *ShardPlan, nodes []*NodeCapabilities) time.Duration {
	// Find minimum bandwidth among nodes
	minBandwidth := int64(math.MaxInt64)
	for _, node := range nodes {
		if node.NetworkBandwidth > 0 && node.NetworkBandwidth < minBandwidth {
			minBandwidth = node.NetworkBandwidth
		}
	}

	// Default to 100 Mbps if not specified
	if minBandwidth == int64(math.MaxInt64) {
		minBandwidth = 100 * 1024 * 1024 / 8 // 100 Mbps in bytes/sec
	}

	// Calculate total data to transfer (including replicas)
	totalBytes := int64(0)
	for _, shard := range plan.Shards {
		totalBytes += shard.Size * int64(plan.ReplicationFactor)
	}

	// Account for parallel transfers
	parallelFactor := m.config.MaxConcurrentTransfers
	if parallelFactor > len(nodes) {
		parallelFactor = len(nodes)
	}

	effectiveBandwidth := minBandwidth * int64(parallelFactor)

	// Add overhead for coordination and verification
	overhead := 1.2 // 20% overhead

	seconds := float64(totalBytes) / float64(effectiveBandwidth) * overhead
	return time.Duration(seconds) * time.Second
}

// calculateLayerRange determines which layers belong to a shard
func (m *ModelShardManager) calculateLayerRange(shardIndex, totalShards int) []int {
	// Assume 100 layers for a large model (this would come from model analysis)
	totalLayers := 100
	layersPerShard := totalLayers / totalShards

	start := shardIndex * layersPerShard
	end := start + layersPerShard - 1

	// Last shard gets remaining layers
	if shardIndex == totalShards-1 {
		end = totalLayers - 1
	}

	return []int{start, end}
}

// calculateShardPriority determines replication priority for a shard
func (m *ModelShardManager) calculateShardPriority(shardIndex, totalShards int) int {
	// Early layers (embedding, initial processing) get higher priority
	// as they're used more frequently
	priority := totalShards - shardIndex

	// First and last shards are most critical
	if shardIndex == 0 || shardIndex == totalShards-1 {
		priority += 10
	}

	return priority
}

// generateChecksum creates a checksum for shard verification using actual model file data
func (m *ModelShardManager) generateChecksum(modelPath string, offset, size int64) (string, error) {
	// Open the model file for reading
	file, err := os.Open(modelPath)
	if err != nil {
		return "", fmt.Errorf("failed to open model file %s: %w", modelPath, err)
	}
	defer file.Close()

	// Seek to the shard offset
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return "", fmt.Errorf("failed to seek to offset %d: %w", offset, err)
	}

	// Create a hasher and read the specified amount of data
	hasher := sha256.New()
	_, err = io.CopyN(hasher, file, size)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read shard data at offset %d, size %d: %w", offset, size, err)
	}

	// Calculate and return the checksum
	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash), nil
}

// nearestPowerOf2 rounds to the nearest power of 2
func nearestPowerOf2(n int) int {
	if n <= 1 {
		return 1
	}

	power := 1
	for power < n {
		power <<= 1
	}

	// Check if the previous power of 2 is closer
	if power-n > n-power/2 && power > 1 {
		power >>= 1
	}

	return power
}

// GetShardPlan returns the shard plan for a model
func (m *ModelShardManager) GetShardPlan(modelID string) (*ShardPlan, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plan, exists := m.shardPlans[modelID]
	if !exists {
		return nil, fmt.Errorf("shard plan not found for model %s", modelID)
	}

	return plan, nil
}

// GetShard returns a specific shard
func (m *ModelShardManager) GetShard(shardID string) (*ModelShard, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	shard, exists := m.activeShards[shardID]
	if !exists {
		return nil, fmt.Errorf("shard %s not found", shardID)
	}

	// Update access metadata
	shard.LastAccessed = time.Now()
	shard.AccessCount++

	return shard, nil
}

// UpdateNodeCapabilities updates the capabilities of a node
func (m *ModelShardManager) UpdateNodeCapabilities(cap *NodeCapabilities) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nodeCapabilities[cap.NodeID] = cap
}

// RebalanceShards rebalances shards when nodes join or leave
func (m *ModelShardManager) RebalanceShards(modelID string, currentNodes []*NodeCapabilities) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plan, exists := m.shardPlans[modelID]
	if !exists {
		return fmt.Errorf("shard plan not found for model %s", modelID)
	}

	// Clear current assignments
	plan.NodeAssignments = make(map[string][]int)
	plan.MemoryRequirements = make(map[string]int64)

	// Reassign shards to current nodes
	if err := m.assignShardsToNodes(plan, currentNodes); err != nil {
		return fmt.Errorf("failed to rebalance shards: %w", err)
	}

	// Recalculate communication topology
	m.calculateCommunicationTopology(plan)

	return nil
}
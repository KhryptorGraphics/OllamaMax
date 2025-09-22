package models

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/partitioning"
)

// MemoryPressureLevel indicates the memory pressure on a node
type MemoryPressureLevel int

const (
	MemoryPressureNormal MemoryPressureLevel = iota
	MemoryPressureModerate
	MemoryPressureHigh
	MemoryPressureCritical
)

// DistributionPlan describes how model shards should be distributed
type DistributionPlan struct {
	ID                string                         `json:"id"`
	ModelID           string                         `json:"model_id"`
	ShardAssignments  map[string][]string           `json:"shard_assignments"`  // NodeID -> ShardIDs
	MemoryAllocations map[string]*MemoryAllocation `json:"memory_allocations"`
	LoadSequence      []LoadStep                    `json:"load_sequence"`
	OptimalTopology   string                        `json:"optimal_topology"`
	EstimatedLoadTime time.Duration                 `json:"estimated_load_time"`
	TotalMemoryUsed   int64                         `json:"total_memory_used"`
	Constraints       []string                      `json:"constraints"`
	CreatedAt         time.Time                     `json:"created_at"`
}

// MemoryAllocation tracks memory usage for a node
type MemoryAllocation struct {
	NodeID            string  `json:"node_id"`
	ModelWeights      int64   `json:"model_weights"`      // Memory for model parameters
	Activations       int64   `json:"activations"`         // Memory for activations
	Gradients         int64   `json:"gradients"`          // Memory for gradients (if training)
	KVCache           int64   `json:"kv_cache"`           // Memory for KV cache (transformers)
	SystemOverhead    int64   `json:"system_overhead"`     // OS and framework overhead
	TotalAllocated    int64   `json:"total_allocated"`
	AvailableMemory   int64   `json:"available_memory"`
	UtilizationRatio  float64 `json:"utilization_ratio"`
}

// LoadStep defines a step in the model loading sequence
type LoadStep struct {
	Order      int      `json:"order"`
	NodeID     string   `json:"node_id"`
	ShardIDs   []string `json:"shard_ids"`
	Parallel   bool     `json:"parallel"`    // Can be done in parallel with same order
	DependsOn  []string `json:"depends_on"`  // ShardIDs that must be loaded first
	MemoryPeak int64    `json:"memory_peak"` // Peak memory during this step
}

// MemoryMonitoringData contains real-time memory stats
type MemoryMonitoringData struct {
	NodeID          string              `json:"node_id"`
	Timestamp       time.Time           `json:"timestamp"`
	UsedMemory      int64               `json:"used_memory"`
	FreeMemory      int64               `json:"free_memory"`
	CachedMemory    int64               `json:"cached_memory"`
	SwapUsed        int64               `json:"swap_used"`
	PressureLevel   MemoryPressureLevel `json:"pressure_level"`
	AllocationTrend float64             `json:"allocation_trend"` // Rate of change
}

// MemoryAwareDistributor calculates optimal model distribution based on memory
type MemoryAwareDistributor struct {
	mu                sync.RWMutex
	memoryOptimizer   *partitioning.MemoryOptimizer
	modelAnalyzer     *partitioning.ModelAnalyzer
	nodeMemoryStats   map[string]*MemoryMonitoringData
	distributionPlans map[string]*DistributionPlan
	config            *DistributorConfig

	// Memory pressure callbacks
	pressureCallbacks []func(nodeID string, level MemoryPressureLevel)
}

// DistributorConfig contains configuration for the distributor
type DistributorConfig struct {
	MemorySafetyMargin        float64       // Reserve this fraction of memory
	EnableGradientCheckpointing bool
	EnableActivationCompression bool
	CompressionRatio          float64       // Expected compression ratio
	KVCacheMultiplier         float64       // KV cache size relative to model size
	MaxMemoryUtilization      float64       // Maximum allowed memory utilization
	MemoryMonitorInterval     time.Duration // How often to check memory
	PressureThresholds        map[MemoryPressureLevel]float64
	EnableDynamicRebalancing  bool
	RebalanceThreshold        float64       // Trigger rebalance when imbalance exceeds this
}

// NewMemoryAwareDistributor creates a new memory-aware distributor
func NewMemoryAwareDistributor(config *DistributorConfig) *MemoryAwareDistributor {
	if config == nil {
		config = DefaultDistributorConfig()
	}

	return &MemoryAwareDistributor{
		memoryOptimizer:   partitioning.NewMemoryOptimizer(),
		modelAnalyzer:     partitioning.NewModelAnalyzer(),
		nodeMemoryStats:   make(map[string]*MemoryMonitoringData),
		distributionPlans: make(map[string]*DistributionPlan),
		config:            config,
		pressureCallbacks: make([]func(string, MemoryPressureLevel), 0),
	}
}

// DefaultDistributorConfig returns default configuration
func DefaultDistributorConfig() *DistributorConfig {
	return &DistributorConfig{
		MemorySafetyMargin:        0.15, // 15% safety margin
		EnableGradientCheckpointing: true,
		EnableActivationCompression: false,
		CompressionRatio:          0.7,
		KVCacheMultiplier:         0.1, // KV cache is ~10% of model size
		MaxMemoryUtilization:      0.85,
		MemoryMonitorInterval:     5 * time.Second,
		PressureThresholds: map[MemoryPressureLevel]float64{
			MemoryPressureNormal:   0.60,
			MemoryPressureModerate: 0.75,
			MemoryPressureHigh:     0.85,
			MemoryPressureCritical: 0.95,
		},
		EnableDynamicRebalancing: true,
		RebalanceThreshold:       0.2, // 20% imbalance triggers rebalance
	}
}

// CalculateDistributionPlan creates an optimal distribution plan
func (d *MemoryAwareDistributor) CalculateDistributionPlan(
	modelAnalysis *partitioning.ModelAnalysis,
	shardPlan *ShardPlan,
	nodeCapabilities []*NodeCapabilities) (*DistributionPlan, error) {

	d.mu.Lock()
	defer d.mu.Unlock()

	// Create distribution plan
	plan := &DistributionPlan{
		ID:                fmt.Sprintf("dist-%s-%d", shardPlan.ModelID, time.Now().Unix()),
		ModelID:           shardPlan.ModelID,
		ShardAssignments:  make(map[string][]string),
		MemoryAllocations: make(map[string]*MemoryAllocation),
		LoadSequence:      make([]LoadStep, 0),
		CreatedAt:         time.Now(),
	}

	// Calculate memory requirements for each shard
	shardMemoryReqs := d.calculateShardMemoryRequirements(modelAnalysis, shardPlan)

	// Sort nodes by available memory (descending)
	sortedNodes := make([]*NodeCapabilities, len(nodeCapabilities))
	copy(sortedNodes, nodeCapabilities)
	sort.Slice(sortedNodes, func(i, j int) bool {
		return sortedNodes[i].AvailableMemory > sortedNodes[j].AvailableMemory
	})

	// Assign shards to nodes using best-fit decreasing algorithm
	nodeAllocations := make(map[string]*MemoryAllocation)
	for _, node := range sortedNodes {
		nodeAllocations[node.NodeID] = &MemoryAllocation{
			NodeID:          node.NodeID,
			AvailableMemory: node.AvailableMemory,
		}
	}

	// Sort shards by memory requirement (descending)
	sortedShards := make([]*ModelShard, len(shardPlan.Shards))
	copy(sortedShards, shardPlan.Shards)
	sort.Slice(sortedShards, func(i, j int) bool {
		return shardMemoryReqs[sortedShards[i].ID] > shardMemoryReqs[sortedShards[j].ID]
	})

	// Assign shards using best-fit algorithm
	for _, shard := range sortedShards {
		memReq := shardMemoryReqs[shard.ID]
		assigned := false

		// Find best-fit node (smallest node that can fit the shard)
		var bestNode *NodeCapabilities
		var bestAllocation *MemoryAllocation
		minWaste := int64(math.MaxInt64)

		for _, node := range sortedNodes {
			allocation := nodeAllocations[node.NodeID]
			effectiveAvailable := d.getEffectiveAvailableMemory(node, allocation)

			if effectiveAvailable >= memReq {
				waste := effectiveAvailable - memReq
				if waste < minWaste {
					minWaste = waste
					bestNode = node
					bestAllocation = allocation
				}
			}
		}

		if bestNode != nil {
			// Assign shard to best-fit node
			plan.ShardAssignments[bestNode.NodeID] = append(
				plan.ShardAssignments[bestNode.NodeID], shard.ID)

			// Update memory allocation
			d.updateMemoryAllocation(bestAllocation, shard, memReq, modelAnalysis)
			assigned = true
		}

		if !assigned {
			return nil, fmt.Errorf("unable to assign shard %s: insufficient memory across all nodes", shard.ID)
		}
	}

	// Store allocations in plan
	for nodeID, allocation := range nodeAllocations {
		if allocation.TotalAllocated > 0 {
			plan.MemoryAllocations[nodeID] = allocation
			plan.TotalMemoryUsed += allocation.TotalAllocated
		}
	}

	// Calculate optimal loading sequence
	plan.LoadSequence = d.calculateLoadSequence(plan, shardPlan)

	// Determine optimal topology based on distribution
	plan.OptimalTopology = d.determineOptimalTopology(plan, shardPlan)

	// Estimate load time
	plan.EstimatedLoadTime = d.estimateLoadTime(plan, nodeCapabilities)

	// Add distribution constraints
	plan.Constraints = d.identifyConstraints(plan, modelAnalysis)

	// Store the plan
	d.distributionPlans[plan.ModelID] = plan

	return plan, nil
}

// ValidateMemoryRequirements ensures each node can handle assigned shards
func (d *MemoryAwareDistributor) ValidateMemoryRequirements(plan *DistributionPlan) error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	for nodeID, allocation := range plan.MemoryAllocations {
		// Check if total allocation exceeds available memory
		if allocation.TotalAllocated > allocation.AvailableMemory {
			return fmt.Errorf("node %s: allocated %d exceeds available %d",
				nodeID, allocation.TotalAllocated, allocation.AvailableMemory)
		}

		// Check utilization ratio
		utilization := float64(allocation.TotalAllocated) / float64(allocation.AvailableMemory)
		if utilization > d.config.MaxMemoryUtilization {
			return fmt.Errorf("node %s: utilization %.2f exceeds max %.2f",
				nodeID, utilization, d.config.MaxMemoryUtilization)
		}

		allocation.UtilizationRatio = utilization
	}

	return nil
}

// MonitorMemoryPressure monitors real-time memory usage
func (d *MemoryAwareDistributor) MonitorMemoryPressure(nodeID string, stats *MemoryMonitoringData) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Update stats
	stats.Timestamp = time.Now()
	d.nodeMemoryStats[nodeID] = stats

	// Calculate pressure level
	usedRatio := float64(stats.UsedMemory) / float64(stats.UsedMemory + stats.FreeMemory)

	oldLevel := stats.PressureLevel
	stats.PressureLevel = d.calculatePressureLevel(usedRatio)

	// Trigger callbacks if pressure level changed
	if oldLevel != stats.PressureLevel {
		for _, callback := range d.pressureCallbacks {
			go callback(nodeID, stats.PressureLevel)
		}
	}

	// Check if rebalancing is needed
	if d.config.EnableDynamicRebalancing && stats.PressureLevel >= MemoryPressureHigh {
		go d.triggerRebalance(nodeID)
	}
}

// RebalanceOnMemoryPressure redistributes shards when nodes experience pressure
func (d *MemoryAwareDistributor) RebalanceOnMemoryPressure(modelID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	plan, exists := d.distributionPlans[modelID]
	if !exists {
		return fmt.Errorf("distribution plan not found for model %s", modelID)
	}

	// Identify nodes under pressure
	pressuredNodes := make([]string, 0)
	availableNodes := make([]string, 0)

	for nodeID, stats := range d.nodeMemoryStats {
		if stats.PressureLevel >= MemoryPressureHigh {
			pressuredNodes = append(pressuredNodes, nodeID)
		} else if stats.PressureLevel <= MemoryPressureModerate {
			availableNodes = append(availableNodes, nodeID)
		}
	}

	if len(pressuredNodes) == 0 || len(availableNodes) == 0 {
		return nil // No rebalancing needed or possible
	}

	// Calculate shards to move
	shardsToMove := make([]string, 0)
	for _, nodeID := range pressuredNodes {
		shards := plan.ShardAssignments[nodeID]
		if len(shards) > 1 {
			// Move some shards from this node
			numToMove := len(shards) / 3 // Move up to 1/3 of shards
			if numToMove == 0 {
				numToMove = 1
			}
			shardsToMove = append(shardsToMove, shards[:numToMove]...)
		}
	}

	// Redistribute shards to available nodes
	shardIndex := 0
	for _, targetNode := range availableNodes {
		if shardIndex >= len(shardsToMove) {
			break
		}

		// Check if target node can accommodate the shard
		allocation := plan.MemoryAllocations[targetNode]
		if allocation.UtilizationRatio < 0.7 { // Only if under 70% utilization
			shardID := shardsToMove[shardIndex]

			// Move shard assignment
			d.moveShardAssignment(plan, shardID, targetNode)
			shardIndex++
		}
	}

	return nil
}

// calculateShardMemoryRequirements calculates memory needs for each shard
func (d *MemoryAwareDistributor) calculateShardMemoryRequirements(
	analysis *partitioning.ModelAnalysis,
	shardPlan *ShardPlan) map[string]int64 {

	requirements := make(map[string]int64)

	for _, shard := range shardPlan.Shards {
		// Base memory for weights
		memReq := shard.Size

		// Add activation memory
		if d.config.EnableGradientCheckpointing {
			// With gradient checkpointing, we store fewer activations
			memReq += int64(float64(shard.Size) * 0.3)
		} else {
			// Without checkpointing, we need more activation memory
			memReq += int64(float64(shard.Size) * 0.5)
		}

		// Add KV cache memory for transformer models
		if analysis != nil && analysis.Architecture == "transformer" {
			kvCacheSize := int64(float64(shard.Size) * d.config.KVCacheMultiplier)
			memReq += kvCacheSize
		}

		// Apply compression if enabled
		if d.config.EnableActivationCompression {
			memReq = int64(float64(memReq) * d.config.CompressionRatio)
		}

		// Add system overhead
		overhead := int64(float64(memReq) * 0.1) // 10% overhead
		memReq += overhead

		requirements[shard.ID] = memReq
	}

	return requirements
}

// getEffectiveAvailableMemory calculates usable memory considering safety margin
func (d *MemoryAwareDistributor) getEffectiveAvailableMemory(
	node *NodeCapabilities,
	allocation *MemoryAllocation) int64 {

	totalAvailable := node.AvailableMemory

	// Apply safety margin
	safeAvailable := int64(float64(totalAvailable) * (1 - d.config.MemorySafetyMargin))

	// Subtract already allocated
	remaining := safeAvailable - allocation.TotalAllocated

	if remaining < 0 {
		return 0
	}

	return remaining
}

// updateMemoryAllocation updates allocation tracking for a node
func (d *MemoryAwareDistributor) updateMemoryAllocation(
	allocation *MemoryAllocation,
	shard *ModelShard,
	memReq int64,
	analysis *partitioning.ModelAnalysis) {

	// Break down memory allocation
	allocation.ModelWeights += shard.Size

	if d.config.EnableGradientCheckpointing {
		allocation.Activations += int64(float64(shard.Size) * 0.3)
	} else {
		allocation.Activations += int64(float64(shard.Size) * 0.5)
	}

	if analysis != nil && analysis.Architecture == "transformer" {
		allocation.KVCache += int64(float64(shard.Size) * d.config.KVCacheMultiplier)
	}

	allocation.SystemOverhead += int64(float64(memReq) * 0.1)
	allocation.TotalAllocated += memReq
}

// calculateLoadSequence determines optimal order for loading shards
func (d *MemoryAwareDistributor) calculateLoadSequence(
	plan *DistributionPlan,
	shardPlan *ShardPlan) []LoadStep {

	sequence := make([]LoadStep, 0)
	order := 0

	// Group shards by node for parallel loading
	for nodeID, shardIDs := range plan.ShardAssignments {
		step := LoadStep{
			Order:    order,
			NodeID:   nodeID,
			ShardIDs: shardIDs,
			Parallel: true, // Shards on same node can load in parallel
		}

		// Calculate peak memory for this step
		peakMemory := int64(0)
		for _, shardID := range shardIDs {
			for _, shard := range shardPlan.Shards {
				if shard.ID == shardID {
					peakMemory += shard.Size
					break
				}
			}
		}
		step.MemoryPeak = peakMemory

		sequence = append(sequence, step)
	}

	// Sort by memory peak (load smaller first to avoid OOM)
	sort.Slice(sequence, func(i, j int) bool {
		return sequence[i].MemoryPeak < sequence[j].MemoryPeak
	})

	// Update order after sorting
	for i := range sequence {
		sequence[i].Order = i
	}

	return sequence
}

// determineOptimalTopology selects best communication pattern
func (d *MemoryAwareDistributor) determineOptimalTopology(
	plan *DistributionPlan,
	shardPlan *ShardPlan) string {

	numNodes := len(plan.ShardAssignments)

	// For small clusters, use all-to-all
	if numNodes <= 4 {
		return "all-to-all"
	}

	// For pipeline parallelism, use ring topology
	if shardPlan.Strategy == ShardingStrategyPipeline {
		return "ring"
	}

	// For large clusters with layer-wise sharding, use hierarchical
	if numNodes > 8 && shardPlan.Strategy == ShardingStrategyLayerWise {
		return "hierarchical"
	}

	// Default to mesh for flexibility
	return "mesh"
}

// estimateLoadTime calculates expected time to load all shards
func (d *MemoryAwareDistributor) estimateLoadTime(
	plan *DistributionPlan,
	nodes []*NodeCapabilities) time.Duration {

	// Find slowest disk I/O speed
	minIOSpeed := int64(100 * 1024 * 1024) // Default 100 MB/s

	for _, node := range nodes {
		if node.StorageCapacity > 0 {
			// Estimate I/O speed based on storage type
			// This is simplified; real implementation would query actual speeds
			ioSpeed := int64(500 * 1024 * 1024) // Assume SSD
			if ioSpeed < minIOSpeed {
				minIOSpeed = ioSpeed
			}
		}
	}

	// Calculate total data to load
	maxLoadPerNode := int64(0)
	for _, allocation := range plan.MemoryAllocations {
		if allocation.ModelWeights > maxLoadPerNode {
			maxLoadPerNode = allocation.ModelWeights
		}
	}

	// Estimate time based on slowest node
	seconds := float64(maxLoadPerNode) / float64(minIOSpeed)

	// Add overhead for initialization and verification
	seconds *= 1.2 // 20% overhead

	return time.Duration(seconds) * time.Second
}

// identifyConstraints lists any constraints affecting distribution
func (d *MemoryAwareDistributor) identifyConstraints(
	plan *DistributionPlan,
	analysis *partitioning.ModelAnalysis) []string {

	constraints := make([]string, 0)

	// Check for memory constraints
	for nodeID, allocation := range plan.MemoryAllocations {
		if allocation.UtilizationRatio > 0.8 {
			constraints = append(constraints,
				fmt.Sprintf("Node %s at high memory utilization (%.1f%%)",
					nodeID, allocation.UtilizationRatio*100))
		}
	}

	// Check for uneven distribution
	minShards := math.MaxInt32
	maxShards := 0
	for _, shardIDs := range plan.ShardAssignments {
		if len(shardIDs) < minShards {
			minShards = len(shardIDs)
		}
		if len(shardIDs) > maxShards {
			maxShards = len(shardIDs)
		}
	}

	if maxShards > minShards*2 {
		constraints = append(constraints,
			fmt.Sprintf("Uneven shard distribution: %d-%d shards per node", minShards, maxShards))
	}

	// Check for architecture-specific constraints
	if analysis != nil {
		if analysis.Architecture == "transformer" && !d.config.EnableGradientCheckpointing {
			constraints = append(constraints,
				"Transformer model without gradient checkpointing requires more memory")
		}
	}

	return constraints
}

// calculatePressureLevel determines memory pressure level from usage ratio
func (d *MemoryAwareDistributor) calculatePressureLevel(usedRatio float64) MemoryPressureLevel {
	for level := MemoryPressureCritical; level >= MemoryPressureNormal; level-- {
		threshold, exists := d.config.PressureThresholds[level]
		if exists && usedRatio >= threshold {
			return level
		}
	}
	return MemoryPressureNormal
}

// triggerRebalance initiates rebalancing for a model
func (d *MemoryAwareDistributor) triggerRebalance(nodeID string) {
	// Find models with shards on the pressured node
	modelsToRebalance := make([]string, 0)

	for modelID, plan := range d.distributionPlans {
		if _, hasShards := plan.ShardAssignments[nodeID]; hasShards {
			modelsToRebalance = append(modelsToRebalance, modelID)
		}
	}

	// Rebalance each affected model
	for _, modelID := range modelsToRebalance {
		_ = d.RebalanceOnMemoryPressure(modelID)
	}
}

// moveShardAssignment moves a shard from one node to another
func (d *MemoryAwareDistributor) moveShardAssignment(
	plan *DistributionPlan,
	shardID string,
	targetNode string) {

	// Find the shard and current node
	var sourceNode string
	var movedShard *ModelShard

	// First find the shard details from the distribution plans
	for modelID, storedPlan := range d.distributionPlans {
		if storedPlan.ModelID == plan.ModelID {
			// Find the shard in the model's shard plan
			shardPlan, err := d.getShardPlanForModel(modelID)
			if err != nil {
				continue // Skip if shard plan not found
			}

			for _, shard := range shardPlan.Shards {
				if shard.ID == shardID {
					movedShard = shard
					break
				}
			}
			break
		}
	}

	// Find current node and remove shard
	for nodeID, shards := range plan.ShardAssignments {
		for i, sID := range shards {
			if sID == shardID {
				sourceNode = nodeID
				// Remove from source
				plan.ShardAssignments[nodeID] = append(shards[:i], shards[i+1:]...)
				break
			}
		}
		if sourceNode != "" {
			break
		}
	}

	// Add to target
	plan.ShardAssignments[targetNode] = append(plan.ShardAssignments[targetNode], shardID)

	// Update memory allocations for both nodes
	if movedShard != nil {
		// Calculate memory requirement for the moved shard
		memReq := d.calculateSingleShardMemoryRequirement(movedShard)

		// Update source node allocation (reduce)
		if sourceAllocation, exists := plan.MemoryAllocations[sourceNode]; exists {
			d.decreaseMemoryAllocation(sourceAllocation, movedShard, memReq)
		}

		// Update target node allocation (increase)
		if targetAllocation, exists := plan.MemoryAllocations[targetNode]; exists {
			d.increaseMemoryAllocation(targetAllocation, movedShard, memReq)
		} else {
			// Create new allocation for target node if it doesn't exist
			plan.MemoryAllocations[targetNode] = &MemoryAllocation{
				NodeID: targetNode,
			}
			d.increaseMemoryAllocation(plan.MemoryAllocations[targetNode], movedShard, memReq)
		}

		// Update total memory used
		plan.TotalMemoryUsed = 0
		for _, allocation := range plan.MemoryAllocations {
			plan.TotalMemoryUsed += allocation.TotalAllocated
		}
	}
}

// getShardPlanForModel retrieves the shard plan for a given model ID
func (d *MemoryAwareDistributor) getShardPlanForModel(modelID string) (*ShardPlan, error) {
	// In a real implementation, this would query the shard manager
	// For now, we'll return an error to indicate this needs to be connected
	return nil, fmt.Errorf("shard plan not found for model %s", modelID)
}

// calculateSingleShardMemoryRequirement calculates memory requirement for a single shard
func (d *MemoryAwareDistributor) calculateSingleShardMemoryRequirement(shard *ModelShard) int64 {
	// Base memory for weights
	memReq := shard.Size

	// Add activation memory
	if d.config.EnableGradientCheckpointing {
		memReq += int64(float64(shard.Size) * 0.3)
	} else {
		memReq += int64(float64(shard.Size) * 0.5)
	}

	// Add KV cache memory (assuming transformer architecture)
	kvCacheSize := int64(float64(shard.Size) * d.config.KVCacheMultiplier)
	memReq += kvCacheSize

	// Apply compression if enabled
	if d.config.EnableActivationCompression {
		memReq = int64(float64(memReq) * d.config.CompressionRatio)
	}

	// Add system overhead
	overhead := int64(float64(memReq) * 0.1) // 10% overhead
	memReq += overhead

	return memReq
}

// decreaseMemoryAllocation reduces memory allocation for a node when removing a shard
func (d *MemoryAwareDistributor) decreaseMemoryAllocation(
	allocation *MemoryAllocation,
	shard *ModelShard,
	memReq int64) {

	// Reduce model weights
	allocation.ModelWeights -= shard.Size

	// Reduce activations
	if d.config.EnableGradientCheckpointing {
		allocation.Activations -= int64(float64(shard.Size) * 0.3)
	} else {
		allocation.Activations -= int64(float64(shard.Size) * 0.5)
	}

	// Reduce KV cache
	allocation.KVCache -= int64(float64(shard.Size) * d.config.KVCacheMultiplier)

	// Reduce system overhead
	allocation.SystemOverhead -= int64(float64(memReq) * 0.1)

	// Update total
	allocation.TotalAllocated -= memReq

	// Ensure values don't go negative
	if allocation.ModelWeights < 0 {
		allocation.ModelWeights = 0
	}
	if allocation.Activations < 0 {
		allocation.Activations = 0
	}
	if allocation.KVCache < 0 {
		allocation.KVCache = 0
	}
	if allocation.SystemOverhead < 0 {
		allocation.SystemOverhead = 0
	}
	if allocation.TotalAllocated < 0 {
		allocation.TotalAllocated = 0
	}

	// Recalculate utilization ratio
	if allocation.AvailableMemory > 0 {
		allocation.UtilizationRatio = float64(allocation.TotalAllocated) / float64(allocation.AvailableMemory)
	}
}

// increaseMemoryAllocation increases memory allocation for a node when adding a shard
func (d *MemoryAwareDistributor) increaseMemoryAllocation(
	allocation *MemoryAllocation,
	shard *ModelShard,
	memReq int64) {

	// Add model weights
	allocation.ModelWeights += shard.Size

	// Add activations
	if d.config.EnableGradientCheckpointing {
		allocation.Activations += int64(float64(shard.Size) * 0.3)
	} else {
		allocation.Activations += int64(float64(shard.Size) * 0.5)
	}

	// Add KV cache
	allocation.KVCache += int64(float64(shard.Size) * d.config.KVCacheMultiplier)

	// Add system overhead
	allocation.SystemOverhead += int64(float64(memReq) * 0.1)

	// Update total
	allocation.TotalAllocated += memReq

	// Recalculate utilization ratio
	if allocation.AvailableMemory > 0 {
		allocation.UtilizationRatio = float64(allocation.TotalAllocated) / float64(allocation.AvailableMemory)
	}
}

// RegisterPressureCallback adds a callback for memory pressure events
func (d *MemoryAwareDistributor) RegisterPressureCallback(
	callback func(nodeID string, level MemoryPressureLevel)) {

	d.mu.Lock()
	defer d.mu.Unlock()

	d.pressureCallbacks = append(d.pressureCallbacks, callback)
}

// GetDistributionPlan returns the distribution plan for a model
func (d *MemoryAwareDistributor) GetDistributionPlan(modelID string) (*DistributionPlan, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	plan, exists := d.distributionPlans[modelID]
	if !exists {
		return nil, fmt.Errorf("distribution plan not found for model %s", modelID)
	}

	return plan, nil
}

// GetNodeMemoryStats returns current memory stats for a node
func (d *MemoryAwareDistributor) GetNodeMemoryStats(nodeID string) (*MemoryMonitoringData, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	stats, exists := d.nodeMemoryStats[nodeID]
	if !exists {
		return nil, fmt.Errorf("memory stats not found for node %s", nodeID)
	}

	return stats, nil
}
package partitioning

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"
)

// DataSplitStrategy implements data-split partitioning
type DataSplitStrategy struct {
	name    string
	metrics *StrategyMetrics
	config  *DataSplitConfig
}

// DataSplitConfig holds configuration for data-split partitioning
type DataSplitConfig struct {
	MinBatchSize     int     `json:"min_batch_size"`
	MaxBatchSize     int     `json:"max_batch_size"`
	MergeStrategy    string  `json:"merge_strategy"`
	LoadBalanceWeight float64 `json:"load_balance_weight"`
	EfficiencyTarget float64 `json:"efficiency_target"`
}

// DataSplitPartition represents a data-split partition
type DataSplitPartition struct {
	BatchSize      int                    `json:"batch_size"`
	PartitionSize  int                    `json:"partition_size"`
	Nodes          []string               `json:"nodes"`
	MergeStrategy  string                 `json:"merge_strategy"`
	DataDistribution map[string]interface{} `json:"data_distribution"`
	LoadBalancing  *LoadBalancingInfo     `json:"load_balancing"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// LoadBalancingInfo represents load balancing information
type LoadBalancingInfo struct {
	Strategy         string             `json:"strategy"`
	Weights          map[string]float64 `json:"weights"`
	CapacityFactors  map[string]float64 `json:"capacity_factors"`
	UtilizationFactors map[string]float64 `json:"utilization_factors"`
}

// DataPartition represents a single data partition
type DataPartition struct {
	ID            string      `json:"id"`
	NodeID        string      `json:"node_id"`
	DataSlice     interface{} `json:"data_slice"`
	BatchSize     int         `json:"batch_size"`
	Weight        float64     `json:"weight"`
	Priority      int         `json:"priority"`
	EstimatedTime time.Duration `json:"estimated_time"`
}

// NewDataSplitStrategy creates a new data-split partitioning strategy
func NewDataSplitStrategy() *DataSplitStrategy {
	return &DataSplitStrategy{
		name: "data_split",
		metrics: &StrategyMetrics{
			LastUsed: time.Now(),
		},
		config: &DataSplitConfig{
			MinBatchSize:     1,
			MaxBatchSize:     32,
			MergeStrategy:    "concat",
			LoadBalanceWeight: 0.7,
			EfficiencyTarget: 0.85,
		},
	}
}

// GetName returns the name of the strategy
func (ds *DataSplitStrategy) GetName() string {
	return ds.name
}

// GetMetrics returns the metrics for the strategy
func (ds *DataSplitStrategy) GetMetrics() *StrategyMetrics {
	return ds.metrics
}

// CanHandle checks if the strategy can handle the given task
func (ds *DataSplitStrategy) CanHandle(task *PartitionTask) bool {
	// Check if we have multiple nodes
	if len(task.Nodes) < 2 {
		return false
	}
	
	// Check if task supports batch processing
	if !ds.supportsBatchProcessing(task) {
		return false
	}
	
	// Check if nodes have sufficient capacity
	sufficientNodes := 0
	for _, node := range task.Nodes {
		if ds.hasCapacityForBatch(node) {
			sufficientNodes++
		}
	}
	
	return sufficientNodes >= 2
}

// supportsBatchProcessing checks if a task supports batch processing
func (ds *DataSplitStrategy) supportsBatchProcessing(task *PartitionTask) bool {
	// Check task type
	if task.Type == "embedding" || task.Type == "classification" {
		return true
	}
	
	// Check if inference can be batched
	if task.Type == "inference" {
		// Check if model supports batching
		if batchable, exists := task.Metadata["batchable"]; exists {
			if b, ok := batchable.(bool); ok {
				return b
			}
		}
		// Default: assume inference can be batched
		return true
	}
	
	return false
}

// hasCapacityForBatch checks if a node has capacity for batch processing
func (ds *DataSplitStrategy) hasCapacityForBatch(node *NodeInfo) bool {
	// Check memory utilization
	if node.Usage.MemoryUtilization > 0.8 {
		return false
	}
	
	// Check GPU utilization
	if node.Usage.GPUUtilization > 0.9 {
		return false
	}
	
	// Check active requests
	if node.Usage.ActiveRequests > 5 {
		return false
	}
	
	return true
}

// Partition performs data-split partitioning
func (ds *DataSplitStrategy) Partition(ctx context.Context, task *PartitionTask) (*PartitionPlan, error) {
	start := time.Now()
	
	// Update metrics
	ds.metrics.TotalPartitions++
	ds.metrics.LastUsed = time.Now()
	
	// Analyze task requirements
	batchRequirements := ds.analyzeBatchRequirements(task)
	
	// Analyze node capacities
	nodeCapacities := ds.analyzeNodeCapacities(task.Nodes)
	
	// Calculate optimal batch distribution
	batchDistribution, err := ds.calculateBatchDistribution(batchRequirements, nodeCapacities)
	if err != nil {
		ds.metrics.FailedPartitions++
		return nil, fmt.Errorf("failed to calculate batch distribution: %v", err)
	}
	
	// Create data partitions
	dataPartitions := ds.createDataPartitions(batchDistribution, task.Nodes)
	
	// Create partitions
	partitions := ds.createPartitions(dataPartitions)
	
	// Estimate performance
	estimatedLatency := ds.estimateLatency(dataPartitions, nodeCapacities)
	estimatedThroughput := ds.estimateThroughput(dataPartitions, nodeCapacities)
	
	// Create partition plan
	plan := &PartitionPlan{
		ID:                  fmt.Sprintf("data_split_%d", time.Now().UnixNano()),
		Strategy:            ds.name,
		Partitions:          partitions,
		EstimatedLatency:    estimatedLatency,
		EstimatedThroughput: estimatedThroughput,
		CreatedAt:           time.Now(),
		Metadata: map[string]interface{}{
			"batch_distribution": batchDistribution,
			"data_partitions":    dataPartitions,
			"node_capacities":    nodeCapacities,
			"merge_strategy":     ds.config.MergeStrategy,
			"partitioning_time":  time.Since(start),
		},
	}
	
	ds.metrics.SuccessfulPartitions++
	
	slog.Info("data-split partitioning completed",
		"task_id", task.ID,
		"data_partitions", len(dataPartitions),
		"estimated_latency", estimatedLatency,
		"estimated_throughput", estimatedThroughput,
		"merge_strategy", ds.config.MergeStrategy)
	
	return plan, nil
}

// BatchRequirements represents batch processing requirements
type BatchRequirements struct {
	TotalBatchSize    int     `json:"total_batch_size"`
	MinPartitionSize  int     `json:"min_partition_size"`
	MaxPartitionSize  int     `json:"max_partition_size"`
	MemoryPerItem     int64   `json:"memory_per_item"`
	ComputePerItem    float64 `json:"compute_per_item"`
	Parallelizable    bool    `json:"parallelizable"`
}

// analyzeBatchRequirements analyzes the batch processing requirements
func (ds *DataSplitStrategy) analyzeBatchRequirements(task *PartitionTask) *BatchRequirements {
	// Extract batch size from task metadata
	totalBatchSize := 1
	if batchSize, exists := task.Metadata["batch_size"]; exists {
		if bs, ok := batchSize.(int); ok {
			totalBatchSize = bs
		}
	}
	
	// Estimate memory per item
	memoryPerItem := ds.estimateMemoryPerItem(task)
	
	// Estimate compute per item
	computePerItem := ds.estimateComputePerItem(task)
	
	// Determine parallelizability
	parallelizable := ds.isParallelizable(task)
	
	return &BatchRequirements{
		TotalBatchSize:    totalBatchSize,
		MinPartitionSize:  ds.config.MinBatchSize,
		MaxPartitionSize:  ds.config.MaxBatchSize,
		MemoryPerItem:     memoryPerItem,
		ComputePerItem:    computePerItem,
		Parallelizable:    parallelizable,
	}
}

// estimateMemoryPerItem estimates memory requirement per batch item
func (ds *DataSplitStrategy) estimateMemoryPerItem(task *PartitionTask) int64 {
	// Base memory estimation
	baseMemory := int64(1024 * 1024) // 1MB base
	
	// Adjust based on context length
	contextMemory := int64(task.Options.NumCtx * 4 * 1024) // 4KB per context token
	
	// Adjust based on task type
	var typeMultiplier float64
	switch task.Type {
	case "embedding":
		typeMultiplier = 0.5
	case "classification":
		typeMultiplier = 0.3
	case "inference":
		typeMultiplier = 1.0
	default:
		typeMultiplier = 1.0
	}
	
	return int64(float64(baseMemory+contextMemory) * typeMultiplier)
}

// estimateComputePerItem estimates compute requirement per batch item
func (ds *DataSplitStrategy) estimateComputePerItem(task *PartitionTask) float64 {
	// Base compute estimation (in GFLOPS)
	baseCompute := 1.0
	
	// Adjust based on model complexity
	if task.GGML != nil {
		modelSize := float64(task.GGML.Size())
		baseCompute = modelSize / (1024 * 1024 * 1024) // GFLOPS based on model size
	}
	
	// Adjust based on context length
	contextMultiplier := math.Log(float64(task.Options.NumCtx)) / math.Log(2048)
	
	return baseCompute * contextMultiplier
}

// isParallelizable checks if a task is parallelizable
func (ds *DataSplitStrategy) isParallelizable(task *PartitionTask) bool {
	// Most batch operations are parallelizable
	switch task.Type {
	case "embedding", "classification":
		return true
	case "inference":
		// Check if model supports parallel inference
		if parallel, exists := task.Metadata["parallel_inference"]; exists {
			if p, ok := parallel.(bool); ok {
				return p
			}
		}
		return true // Default assumption
	default:
		return false
	}
}

// NodeCapacity represents the capacity of a node for batch processing
type NodeCapacity struct {
	Node            *NodeInfo `json:"node"`
	AvailableMemory int64     `json:"available_memory"`
	AvailableCompute float64  `json:"available_compute"`
	Throughput      float64   `json:"throughput"`
	Latency         time.Duration `json:"latency"`
	CapacityScore   float64   `json:"capacity_score"`
}

// analyzeNodeCapacities analyzes the capacity of nodes for batch processing
func (ds *DataSplitStrategy) analyzeNodeCapacities(nodes []*NodeInfo) []*NodeCapacity {
	capacities := make([]*NodeCapacity, len(nodes))
	
	for i, node := range nodes {
		// Calculate available memory
		availableMemory := int64(float64(node.Capacity.MemoryBytes) * (1.0 - node.Usage.MemoryUtilization))
		
		// Calculate available compute
		availableCompute := float64(node.Capacity.CPUCores) * (1.0 - node.Usage.CPUUtilization)
		
		// Estimate throughput
		throughput := ds.estimateNodeThroughput(node)
		
		// Calculate capacity score
		capacityScore := ds.calculateCapacityScore(node, availableMemory, availableCompute)
		
		capacities[i] = &NodeCapacity{
			Node:            node,
			AvailableMemory: availableMemory,
			AvailableCompute: availableCompute,
			Throughput:      throughput,
			Latency:         node.Latency,
			CapacityScore:   capacityScore,
		}
	}
	
	return capacities
}

// estimateNodeThroughput estimates the throughput of a node
func (ds *DataSplitStrategy) estimateNodeThroughput(node *NodeInfo) float64 {
	// Base throughput on compute capacity
	baseThroughput := float64(node.Capacity.CPUCores) * 10.0 // 10 ops per core per second
	
	// Adjust for GPU if available
	if node.Capacity.GPUCount > 0 {
		gpuThroughput := float64(node.Capacity.GPUCount) * 100.0 // 100 ops per GPU per second
		baseThroughput += gpuThroughput
	}
	
	// Adjust for current utilization
	utilizationFactor := 1.0 - (node.Usage.CPUUtilization+node.Usage.GPUUtilization)/2.0
	
	// Apply capacity-specific multiplier
	if node.Capacity.ComputeScore > 0 {
		baseThroughput *= node.Capacity.ComputeScore
	}
	
	return baseThroughput * utilizationFactor
}

// calculateCapacityScore calculates a capacity score for a node
func (ds *DataSplitStrategy) calculateCapacityScore(node *NodeInfo, availableMemory int64, availableCompute float64) float64 {
	// Memory score (0-1)
	memoryScore := math.Min(float64(availableMemory)/(4*1024*1024*1024), 1.0) // 4GB reference
	
	// Compute score (0-1)
	computeScore := math.Min(availableCompute/8.0, 1.0) // 8 cores reference
	
	// Throughput score (0-1)
	throughput := ds.estimateNodeThroughput(node)
	throughputScore := math.Min(throughput/1000.0, 1.0) // 1000 ops/sec reference
	
	// Latency score (0-1)
	latencyScore := math.Max(0.0, 1.0-float64(node.Latency)/float64(100*time.Millisecond))
	
	// Weighted combination
	capacityScore := 0.3*memoryScore + 0.3*computeScore + 0.3*throughputScore + 0.1*latencyScore
	
	return math.Min(capacityScore, 1.0)
}

// BatchDistribution represents the distribution of batch items across nodes
type BatchDistribution struct {
	TotalItems     int                   `json:"total_items"`
	NodeAllocation map[string]int        `json:"node_allocation"`
	Weights        map[string]float64    `json:"weights"`
	LoadBalancing  *LoadBalancingInfo    `json:"load_balancing"`
	Efficiency     float64               `json:"efficiency"`
}

// calculateBatchDistribution calculates the optimal batch distribution
func (ds *DataSplitStrategy) calculateBatchDistribution(requirements *BatchRequirements, capacities []*NodeCapacity) (*BatchDistribution, error) {
	if len(capacities) == 0 {
		return nil, fmt.Errorf("no nodes available for batch distribution")
	}
	
	// Calculate total capacity
	totalCapacity := 0.0
	for _, capacity := range capacities {
		totalCapacity += capacity.CapacityScore
	}
	
	if totalCapacity == 0 {
		return nil, fmt.Errorf("no available capacity for batch distribution")
	}
	
	// Calculate weights based on capacity
	weights := make(map[string]float64)
	for _, capacity := range capacities {
		weights[capacity.Node.ID] = capacity.CapacityScore / totalCapacity
	}
	
	// Calculate node allocation
	nodeAllocation := make(map[string]int)
	remainingItems := requirements.TotalBatchSize
	
	for i, capacity := range capacities {
		weight := weights[capacity.Node.ID]
		allocation := int(float64(requirements.TotalBatchSize) * weight)
		
		// Ensure minimum and maximum constraints
		if allocation < requirements.MinPartitionSize {
			allocation = requirements.MinPartitionSize
		}
		if allocation > requirements.MaxPartitionSize {
			allocation = requirements.MaxPartitionSize
		}
		
		// Adjust for remaining items
		if i == len(capacities)-1 && remainingItems > 0 {
			// Give remaining items to the last node
			allocation = remainingItems
		} else if allocation > remainingItems {
			allocation = remainingItems
		}
		
		nodeAllocation[capacity.Node.ID] = allocation
		remainingItems -= allocation
		
		if remainingItems <= 0 {
			break
		}
	}
	
	// Calculate efficiency
	efficiency := ds.calculateDistributionEfficiency(nodeAllocation, weights)
	
	// Create load balancing info
	loadBalancing := &LoadBalancingInfo{
		Strategy:           "capacity_weighted",
		Weights:            weights,
		CapacityFactors:    make(map[string]float64),
		UtilizationFactors: make(map[string]float64),
	}
	
	for _, capacity := range capacities {
		loadBalancing.CapacityFactors[capacity.Node.ID] = capacity.CapacityScore
		loadBalancing.UtilizationFactors[capacity.Node.ID] = (capacity.Node.Usage.CPUUtilization + capacity.Node.Usage.GPUUtilization) / 2.0
	}
	
	return &BatchDistribution{
		TotalItems:     requirements.TotalBatchSize,
		NodeAllocation: nodeAllocation,
		Weights:        weights,
		LoadBalancing:  loadBalancing,
		Efficiency:     efficiency,
	}, nil
}

// calculateDistributionEfficiency calculates the efficiency of a distribution
func (ds *DataSplitStrategy) calculateDistributionEfficiency(nodeAllocation map[string]int, weights map[string]float64) float64 {
	// Calculate how well the allocation matches the weights
	totalAllocation := 0
	for _, allocation := range nodeAllocation {
		totalAllocation += allocation
	}
	
	if totalAllocation == 0 {
		return 0.0
	}
	
	// Calculate efficiency as the correlation between allocation and weights
	efficiency := 0.0
	for nodeID, allocation := range nodeAllocation {
		actualWeight := float64(allocation) / float64(totalAllocation)
		expectedWeight := weights[nodeID]
		
		// Penalty for deviation from expected weight
		deviation := math.Abs(actualWeight - expectedWeight)
		efficiency += (1.0 - deviation) * expectedWeight
	}
	
	return efficiency
}

// createDataPartitions creates data partitions from batch distribution
func (ds *DataSplitStrategy) createDataPartitions(distribution *BatchDistribution, nodes []*NodeInfo) []*DataPartition {
	partitions := make([]*DataPartition, 0)
	
	for _, node := range nodes {
		allocation := distribution.NodeAllocation[node.ID]
		if allocation > 0 {
			partition := &DataPartition{
				ID:            fmt.Sprintf("data_%s", node.ID),
				NodeID:        node.ID,
				DataSlice:     ds.createDataSlice(allocation),
				BatchSize:     allocation,
				Weight:        distribution.Weights[node.ID],
				Priority:      ds.calculatePriority(node, allocation),
				EstimatedTime: ds.estimatePartitionTime(node, allocation),
			}
			
			partitions = append(partitions, partition)
		}
	}
	
	return partitions
}

// createDataSlice creates a data slice representation
func (ds *DataSplitStrategy) createDataSlice(size int) interface{} {
	return map[string]interface{}{
		"size":        size,
		"start_index": 0,
		"end_index":   size - 1,
		"type":        "batch_slice",
	}
}

// calculatePriority calculates the priority for a partition
func (ds *DataSplitStrategy) calculatePriority(node *NodeInfo, allocation int) int {
	// Higher priority for nodes with lower utilization
	utilization := (node.Usage.CPUUtilization + node.Usage.GPUUtilization) / 2.0
	priority := int((1.0 - utilization) * 10.0)
	
	// Adjust based on allocation size
	if allocation > 16 {
		priority += 2 // Higher priority for larger allocations
	}
	
	return priority
}

// estimatePartitionTime estimates the time to process a partition
func (ds *DataSplitStrategy) estimatePartitionTime(node *NodeInfo, allocation int) time.Duration {
	// Base processing time per item
	baseTimePerItem := 100 * time.Millisecond
	
	// Adjust based on node capacity
	if node.Capacity.ComputeScore > 0 {
		baseTimePerItem = time.Duration(float64(baseTimePerItem) / node.Capacity.ComputeScore)
	}
	
	// Adjust based on current utilization
	utilization := (node.Usage.CPUUtilization + node.Usage.GPUUtilization) / 2.0
	adjustmentFactor := 1.0 + utilization
	
	// Total time = base time * items * adjustment factor
	totalTime := time.Duration(float64(baseTimePerItem) * float64(allocation) * adjustmentFactor)
	
	return totalTime
}

// createPartitions creates partitions from data partitions
func (ds *DataSplitStrategy) createPartitions(dataPartitions []*DataPartition) []*Partition {
	partitions := make([]*Partition, len(dataPartitions))
	
	for i, dataPartition := range dataPartitions {
		partitions[i] = &Partition{
			ID:               dataPartition.ID,
			NodeID:           dataPartition.NodeID,
			Type:             PartitionTypeData,
			Data:             dataPartition,
			Dependencies:     []string{}, // Data partitions are independent
			EstimatedLatency: dataPartition.EstimatedTime,
			EstimatedMemory:  ds.estimatePartitionMemory(dataPartition),
			Metadata: map[string]interface{}{
				"batch_size":     dataPartition.BatchSize,
				"weight":         dataPartition.Weight,
				"priority":       dataPartition.Priority,
				"data_slice":     dataPartition.DataSlice,
				"merge_strategy": ds.config.MergeStrategy,
			},
		}
	}
	
	return partitions
}

// estimatePartitionMemory estimates memory usage for a data partition
func (ds *DataSplitStrategy) estimatePartitionMemory(dataPartition *DataPartition) int64 {
	// Base memory per item
	baseMemoryPerItem := int64(1024 * 1024) // 1MB per item
	
	// Total memory = base memory * batch size * overhead factor
	overheadFactor := 1.2 // 20% overhead
	totalMemory := int64(float64(baseMemoryPerItem) * float64(dataPartition.BatchSize) * overheadFactor)
	
	return totalMemory
}

// estimateLatency estimates the latency for data-split partitioning
func (ds *DataSplitStrategy) estimateLatency(dataPartitions []*DataPartition, capacities []*NodeCapacity) time.Duration {
	// Find the maximum estimated time (bottleneck)
	maxTime := time.Duration(0)
	for _, partition := range dataPartitions {
		if partition.EstimatedTime > maxTime {
			maxTime = partition.EstimatedTime
		}
	}
	
	// Add merge overhead
	mergeOverhead := time.Duration(float64(len(dataPartitions)) * 10.0 * float64(time.Millisecond))
	
	return maxTime + mergeOverhead
}

// estimateThroughput estimates the throughput for data-split partitioning
func (ds *DataSplitStrategy) estimateThroughput(dataPartitions []*DataPartition, capacities []*NodeCapacity) float64 {
	// Calculate total throughput (sum of all partitions)
	totalThroughput := 0.0
	for _, partition := range dataPartitions {
		// Throughput = batch size / estimated time
		partitionThroughput := float64(partition.BatchSize) / partition.EstimatedTime.Seconds()
		totalThroughput += partitionThroughput
	}
	
	// Apply efficiency factor
	efficiencyFactor := 0.9 // 90% efficiency for data-split
	
	return totalThroughput * efficiencyFactor
}

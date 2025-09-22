package partitioning

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"
)

// TensorPartitionStrategy implements tensor parallelism partitioning
type TensorPartitionStrategy struct {
	name           string
	analyzer       *ModelAnalyzer
	memoryOptimizer *MemoryOptimizer
	metrics        *StrategyMetrics
	config         *TensorStrategyConfig
}

// TensorStrategyConfig contains configuration for tensor partitioning
type TensorStrategyConfig struct {
	MinTensorSizeForSplit   int64   `json:"min_tensor_size_for_split"`   // Minimum tensor size to consider splitting
	MaxSplitDimensions      int     `json:"max_split_dimensions"`        // Maximum dimensions to split across
	PreferAttentionSplit    bool    `json:"prefer_attention_split"`      // Prefer splitting attention tensors
	AllowAsymmetricSplit    bool    `json:"allow_asymmetric_split"`      // Allow uneven tensor splits
	CommunicationOverhead   float64 `json:"communication_overhead"`      // Communication overhead factor
	AggregationStrategy     string  `json:"aggregation_strategy"`        // all-reduce, all-gather, etc.
	SplitThreshold          float64 `json:"split_threshold"`             // Threshold for tensor splitting decision
}

// NewTensorPartitionStrategy creates a new tensor partition strategy
func NewTensorPartitionStrategy(analyzer *ModelAnalyzer, optimizer *MemoryOptimizer) *TensorPartitionStrategy {
	return &TensorPartitionStrategy{
		name:            "tensor_parallelism",
		analyzer:        analyzer,
		memoryOptimizer: optimizer,
		metrics: &StrategyMetrics{
			UsageCount:   0,
			SuccessCount: 0,
			FailureCount: 0,
			Performance:  &PerformanceMetrics{},
			CustomMetrics: make(map[string]interface{}),
		},
		config: &TensorStrategyConfig{
			MinTensorSizeForSplit: 64 * 1024 * 1024, // 64MB
			MaxSplitDimensions:    4,
			PreferAttentionSplit:  true,
			AllowAsymmetricSplit:  false,
			CommunicationOverhead: 0.1,              // 10% overhead
			AggregationStrategy:   "all_reduce",
			SplitThreshold:        0.7,              // 70% memory utilization threshold
		},
	}
}

// GetName returns the strategy name
func (tps *TensorPartitionStrategy) GetName() string {
	return tps.name
}

// GetMetrics returns strategy metrics
func (tps *TensorPartitionStrategy) GetMetrics() *StrategyMetrics {
	return tps.metrics
}

// CanHandle determines if this strategy can handle the given request
func (tps *TensorPartitionStrategy) CanHandle(req *PartitionRequest) bool {
	if req == nil || req.Model == nil {
		return false
	}

	// Tensor parallelism is suitable for:
	// 1. Large models where individual tensors are significant
	// 2. Models with attention mechanisms that can benefit from Q/K/V splitting
	// 3. High-bandwidth interconnect environments
	// 4. When memory per node is limited but aggregate memory is sufficient

	modelSize := req.Model.Parameters
	if modelSize < 7_000_000_000 { // Less than 7B parameters
		return false // Not large enough to benefit from tensor parallelism
	}

	// Check if we have tensor information
	if req.ModelAnalysis != nil && req.ModelAnalysis.TensorInfo != nil {
		tensorInfo := req.ModelAnalysis.TensorInfo
		if len(tensorInfo.SplittableTensors) == 0 {
			return false // No tensors that can be split
		}

		// Check if any tensors are large enough to split
		hasSplittableTensors := false
		for _, tensorIdx := range tensorInfo.SplittableTensors {
			// Check in CombinedTensors if available, otherwise fall back to old approach
			if len(tensorInfo.CombinedTensors) > 0 {
				if tensorIdx >= 0 && tensorIdx < len(tensorInfo.CombinedTensors) {
					if tensorInfo.CombinedTensors[tensorIdx].Size >= tps.config.MinTensorSizeForSplit {
						hasSplittableTensors = true
						break
					}
				}
			} else {
				// Fallback to old approach for backwards compatibility
				if tensorIdx < len(tensorInfo.AttentionTensors) {
					if tensorInfo.AttentionTensors[tensorIdx].Size >= tps.config.MinTensorSizeForSplit {
						hasSplittableTensors = true
						break
					}
				} else {
					// Check WeightTensors with offset index
					weightIdx := tensorIdx - len(tensorInfo.AttentionTensors)
					if weightIdx >= 0 && weightIdx < len(tensorInfo.WeightTensors) {
						if tensorInfo.WeightTensors[weightIdx].Size >= tps.config.MinTensorSizeForSplit {
							hasSplittableTensors = true
							break
						}
					}
				}
			}
		}
		if !hasSplittableTensors {
			return false
		}
	}

	return true
}

// Partition creates a partition plan using tensor parallelism
func (tps *TensorPartitionStrategy) Partition(ctx context.Context, req *PartitionRequest, nodes []*NodeInfo) (*PartitionPlan, error) {
	startTime := time.Now()
	tps.metrics.UsageCount++

	// Validate inputs
	if err := tps.validateInputs(req, nodes); err != nil {
		tps.metrics.FailureCount++
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// Analyze model if not already done
	modelAnalysis := req.ModelAnalysis
	if modelAnalysis == nil {
		var err error
		modelAnalysis, err = tps.analyzer.AnalyzeModel(req.Model.Path, req.Model.Details)
		if err != nil {
			tps.metrics.FailureCount++
			return nil, fmt.Errorf("model analysis failed: %w", err)
		}
	}

	// Filter and rank nodes for tensor parallelism
	suitableNodes, err := tps.selectSuitableNodes(nodes, modelAnalysis, req.Options)
	if err != nil {
		tps.metrics.FailureCount++
		return nil, fmt.Errorf("node selection failed: %w", err)
	}

	// Create tensor assignment plan
	assignments, err := tps.createTensorAssignments(modelAnalysis, suitableNodes, req.Options)
	if err != nil {
		tps.metrics.FailureCount++
		return nil, fmt.Errorf("tensor assignment failed: %w", err)
	}

	// Create communication plan for tensor parallelism
	communicationPlan := tps.createCommunicationPlan(assignments, modelAnalysis)

	// Calculate resource costs
	estimatedCost := tps.calculateResourceCost(assignments, suitableNodes, modelAnalysis)

	// Create the final partition plan
	plan := &PartitionPlan{
		ID:            generateTensorPartitionPlanID(),
		TaskID:        req.TaskID,
		Strategy:      tps.name,
		Assignments:   assignments,
		Communication: communicationPlan,
		EstimatedCost: estimatedCost,
		CreatedAt:     time.Now(),
		Metadata: map[string]interface{}{
			"total_tensors":        len(modelAnalysis.TensorInfo.AttentionTensors) + len(modelAnalysis.TensorInfo.WeightTensors),
			"splittable_tensors":   len(modelAnalysis.TensorInfo.SplittableTensors),
			"nodes_used":          len(assignments),
			"split_strategy":      tps.config.AggregationStrategy,
			"optimization_target": req.Options.OptimizeFor,
			"execution_time_ms":   float64(time.Since(startTime).Nanoseconds()) / 1e6,
		},
	}

	// Update metrics
	tps.metrics.SuccessCount++
	tps.metrics.LastUsed = time.Now()
	tps.updatePerformanceMetrics(time.Since(startTime), len(suitableNodes), modelAnalysis)

	return plan, nil
}

// validateInputs validates the partition request and available nodes
func (tps *TensorPartitionStrategy) validateInputs(req *PartitionRequest, nodes []*NodeInfo) error {
	if err := ValidatePartitionRequest(req); err != nil {
		return err
	}

	if len(nodes) == 0 {
		return fmt.Errorf("no nodes available")
	}

	// Tensor parallelism requires nodes with good interconnect
	validNodes := 0
	for _, node := range nodes {
		if err := ValidateNodeInfo(node); err != nil {
			continue
		}
		
		// Check for sufficient network bandwidth
		if node.Capabilities.Network != nil && node.Capabilities.Network.Bandwidth >= 10.0 {
			validNodes++
		}
	}

	if validNodes < 2 {
		return fmt.Errorf("tensor parallelism requires at least 2 nodes with high-speed interconnect")
	}

	return nil
}

// selectSuitableNodes filters and ranks nodes for tensor parallelism
func (tps *TensorPartitionStrategy) selectSuitableNodes(nodes []*NodeInfo, analysis *ModelAnalysis, options *PartitionOptions) ([]*NodeInfo, error) {
	var suitableNodes []*NodeInfo

	// For tensor parallelism, we need nodes with:
	// 1. Sufficient memory for tensor splits
	// 2. High-bandwidth interconnect (for frequent communication)
	// 3. Similar computational capabilities (for balanced execution)

	minMemoryPerNode := analysis.MemoryReqs.TotalRequired / int64(len(nodes))
	if minMemoryPerNode < analysis.MemoryReqs.MinNodeMemory {
		minMemoryPerNode = analysis.MemoryReqs.MinNodeMemory
	}

	for _, node := range nodes {
		if node.Status != NodeStatusActive {
			continue
		}

		// Skip nodes without capability information
		if node.Capabilities == nil || node.Capabilities.Memory == nil || node.Capabilities.Network == nil {
			continue
		}

		// Check memory requirements
		if node.Capabilities.Memory.AvailableBytes < minMemoryPerNode {
			continue
		}

		// Check network bandwidth (critical for tensor parallelism)
		if node.Capabilities.Network.Bandwidth < 10.0 {
			continue // Need at least 10 Gbps for tensor parallelism
		}

		// Check current load
		if node.CurrentLoad != nil {
			if node.CurrentLoad.MemoryUtilization > 0.7 || 
			   node.CurrentLoad.NetworkUtilization > 0.8 {
				continue // Too much load for tensor parallelism
			}
		}

		suitableNodes = append(suitableNodes, node)
	}

	if len(suitableNodes) < 2 {
		return nil, fmt.Errorf("insufficient suitable nodes: need at least 2, found %d", len(suitableNodes))
	}

	// Rank nodes for tensor parallelism
	tps.rankNodesForTensorParallelism(suitableNodes, analysis)

	// Select optimal number of nodes for tensor parallelism
	optimalNodes := tps.calculateOptimalNodeCount(suitableNodes, analysis, options)
	
	if len(suitableNodes) > optimalNodes {
		suitableNodes = suitableNodes[:optimalNodes]
	}

	return suitableNodes, nil
}

// rankNodesForTensorParallelism ranks nodes based on their suitability for tensor parallelism
func (tps *TensorPartitionStrategy) rankNodesForTensorParallelism(nodes []*NodeInfo, analysis *ModelAnalysis) {
	sort.Slice(nodes, func(i, j int) bool {
		scoreI := tps.calculateTensorParallelismScore(nodes[i], analysis)
		scoreJ := tps.calculateTensorParallelismScore(nodes[j], analysis)
		return scoreI > scoreJ
	})
}

// calculateTensorParallelismScore calculates suitability score for tensor parallelism
func (tps *TensorPartitionStrategy) calculateTensorParallelismScore(node *NodeInfo, analysis *ModelAnalysis) float64 {
	score := 0.0

	// Guard against nil capabilities
	if node.Capabilities == nil {
		return score
	}

	// Network bandwidth (50% weight - critical for tensor parallelism)
	if node.Capabilities.Network != nil {
		bandwidth := node.Capabilities.Network.Bandwidth
		if bandwidth > 0 {
			networkScore := math.Min(bandwidth/100.0, 1.0) // Normalize to 100 Gbps
			score += networkScore * 0.5
		}
	}

	// Memory capacity (25% weight)
	if node.Capabilities.Memory != nil && node.Capabilities.Memory.AvailableBytes > 0 {
		minNodeMemory := analysis.MemoryReqs.MinNodeMemory
		if minNodeMemory <= 0 {
			minNodeMemory = 1 // Prevent division by zero
		}
		memoryRatio := float64(node.Capabilities.Memory.AvailableBytes) / float64(minNodeMemory)
		memoryScore := math.Min(memoryRatio/2.0, 1.0) // Normalize to 2x required
		score += memoryScore * 0.25
	}

	// GPU capabilities (15% weight - important for tensor operations)
	if node.Capabilities.GPU != nil {
		gpuCount := node.Capabilities.GPU.Count
		if gpuCount > 0 {
			gpuScore := math.Min(float64(gpuCount)/4.0, 1.0) // Normalize to 4 GPUs
			score += gpuScore * 0.15
		}
	}

	// Network latency (10% weight - lower is better)
	if node.Capabilities.Network != nil {
		latency := node.Capabilities.Network.Latency
		// Handle negative latency values (treat as 0)
		if latency < 0 {
			latency = 0
		}
		latencyScore := math.Max(0, 1.0-(latency/10.0)) // Penalize latency > 10ms
		score += latencyScore * 0.10
	}

	return score
}

// calculateOptimalNodeCount calculates the optimal number of nodes for tensor parallelism
func (tps *TensorPartitionStrategy) calculateOptimalNodeCount(nodes []*NodeInfo, analysis *ModelAnalysis, options *PartitionOptions) int {
	// For tensor parallelism, more nodes isn't always better due to communication overhead
	_ = len(analysis.TensorInfo.AttentionTensors) + len(analysis.TensorInfo.WeightTensors) // Total tensors (might be used later)
	splittableTensors := len(analysis.TensorInfo.SplittableTensors)

	// Heuristic: optimal node count based on model size and tensor count
	var optimal int
	switch {
	case analysis.ParameterSize >= 70: // 70B+ parameters
		optimal = 8
	case analysis.ParameterSize >= 30: // 30-70B parameters
		optimal = 4
	case analysis.ParameterSize >= 13: // 13-30B parameters
		optimal = 4
	case analysis.ParameterSize >= 7:  // 7-13B parameters
		optimal = 2
	default:
		optimal = 2
	}

	// Limit by splittable tensors
	if splittableTensors > 0 && splittableTensors < optimal {
		optimal = splittableTensors
	}

	// Apply user constraints
	if options.MaxNodes > 0 && optimal > options.MaxNodes {
		optimal = options.MaxNodes
	}
	if options.MinNodes > 0 && optimal < options.MinNodes {
		optimal = options.MinNodes
	}

	// Ensure we don't exceed available nodes
	if optimal > len(nodes) {
		optimal = len(nodes)
	}

	return optimal
}

// createTensorAssignments creates tensor assignments for nodes
func (tps *TensorPartitionStrategy) createTensorAssignments(analysis *ModelAnalysis, nodes []*NodeInfo, options *PartitionOptions) ([]*NodeAssignment, error) {
	tensorInfo := analysis.TensorInfo
	
	// Group tensors by layer and type for optimal distribution
	tensorGroups := tps.groupTensorsByLayer(tensorInfo)
	
	// Create assignments based on tensor splits
	assignments, err := tps.distributeTensorGroups(tensorGroups, nodes, analysis)
	if err != nil {
		return nil, fmt.Errorf("tensor distribution failed: %w", err)
	}

	return assignments, nil
}

// groupTensorsByLayer groups tensors by their layer index for coherent splitting
func (tps *TensorPartitionStrategy) groupTensorsByLayer(tensorInfo *TensorAnalysis) map[int][]*TensorInfo {
	groups := make(map[int][]*TensorInfo)

	// Group attention tensors by layer
	for i := range tensorInfo.AttentionTensors {
		tensor := &tensorInfo.AttentionTensors[i]
		layerIdx := tensor.LayerIndex
		groups[layerIdx] = append(groups[layerIdx], tensor)
	}

	// Group weight tensors by layer
	for i := range tensorInfo.WeightTensors {
		tensor := &tensorInfo.WeightTensors[i]
		layerIdx := tensor.LayerIndex
		groups[layerIdx] = append(groups[layerIdx], tensor)
	}

	return groups
}

// distributeTensorGroups distributes tensor groups across nodes
func (tps *TensorPartitionStrategy) distributeTensorGroups(groups map[int][]*TensorInfo, nodes []*NodeInfo, analysis *ModelAnalysis) ([]*NodeAssignment, error) {
	var assignments []*NodeAssignment

	// Calculate total tensor memory and processing requirements
	totalTensorMemory := int64(0)
	for _, tensors := range groups {
		for _, tensor := range tensors {
			totalTensorMemory += tensor.Size
		}
	}

	// Create assignments for each node
	for i, node := range nodes {
		assignment, err := tps.createTensorAssignment(node, i, nodes, groups, analysis, totalTensorMemory)
		if err != nil {
			return nil, fmt.Errorf("failed to create tensor assignment for node %s: %w", node.ID, err)
		}
		assignments = append(assignments, assignment)
	}

	return assignments, nil
}

// createTensorAssignment creates a tensor assignment for a specific node
func (tps *TensorPartitionStrategy) createTensorAssignment(node *NodeInfo, nodeIndex int, nodes []*NodeInfo, groups map[int][]*TensorInfo, analysis *ModelAnalysis, totalMemory int64) (*NodeAssignment, error) {
	totalNodes := len(nodes)
	var tensorAssignments []TensorAssignment
	totalMemoryForNode := int64(0)
	totalComputeWeight := 0.0

	// Ensure we have a deterministic tensor index map
	indexMap := analysis.TensorInfo.IndexMap
	if indexMap == nil {
		// Build index map from CombinedTensors for deterministic ordering
		indexMap = make(map[string]int)
		for i, tensor := range analysis.TensorInfo.CombinedTensors {
			indexMap[tensor.Name] = i
		}
	}

	// Distribute tensor splits across nodes
	for _, tensors := range groups {
		for _, tensor := range tensors {
			// Get the deterministic tensor index from IndexMap
			tensorIndex, ok := indexMap[tensor.Name]
			if !ok {
				// This should not happen if CombinedTensors is properly built
				return nil, fmt.Errorf("tensor %s not found in index map", tensor.Name)
			}
			
			if !tensor.Splittable {
				// Assign entire tensor to first node if not splittable
				if nodeIndex == 0 {
					splitInfo := []TensorSplitInfo{
						{
							TensorIndex:   tensorIndex, // Now using deterministic index
							TensorName:    tensor.Name,
							SplitDimension: -1, // No split
							SplitStart:    0,
							SplitEnd:      len(tensor.Shape),
							OriginalShape: tensor.Shape,
							SplitShape:    tensor.Shape,
						},
					}
					
					tensorAssignment := TensorAssignment{
						TensorIndices:  []int{tensorIndex},
						TensorNames:    []string{tensor.Name},
						SplitInfo:      splitInfo,
						MemoryRequired: tensor.Size,
						ComputeWeight:  1.0, // Assume unit compute weight
					}
					
					tensorAssignments = append(tensorAssignments, tensorAssignment)
					totalMemoryForNode += tensor.Size
					totalComputeWeight += 1.0
				}
				continue
			}

			// Split tensor across nodes
			splitInfo := tps.calculateTensorSplit(tensor, nodeIndex, len(nodes))
			if splitInfo.SplitEnd > splitInfo.SplitStart {
				// Use the deterministic tensor index in splitInfo
				splitInfo.TensorIndex = tensorIndex
				
				tensorAssignment := TensorAssignment{
					TensorIndices:  []int{tensorIndex},
					TensorNames:    []string{tensor.Name},
					SplitInfo:      []TensorSplitInfo{*splitInfo},
					MemoryRequired: tensor.Size / int64(totalNodes), // Approximate split size
					ComputeWeight:  1.0 / float64(totalNodes),       // Distributed compute
				}
				
				tensorAssignments = append(tensorAssignments, tensorAssignment)
				totalMemoryForNode += tensorAssignment.MemoryRequired
				totalComputeWeight += tensorAssignment.ComputeWeight
			}
		}
	}

	// Create work assignment
	workAssignment := &WorkAssignment{
		Tensors: tensorAssignments,
	}

	// Calculate resource requirements
	resources := tps.calculateTensorResourceRequirements(totalMemoryForNode, totalComputeWeight, node.Capabilities)

	// Create dependencies (tensor parallelism requires all nodes to coordinate)
	var dependencies []string
	for i, otherNode := range nodes {
		if i != nodeIndex {
			dependencies = append(dependencies, otherNode.ID)
		}
	}

	// Create the node assignment
	assignment := &NodeAssignment{
		NodeID:       node.ID,
		Role:         RoleWorker,
		WorkType:     WorkTypeTensors,
		Assignment:   workAssignment,
		Resources:    resources,
		Dependencies: dependencies,
	}

	return assignment, nil
}

// calculateTensorSplit calculates how to split a tensor for a specific node
func (tps *TensorPartitionStrategy) calculateTensorSplit(tensor *TensorInfo, nodeIndex, totalNodes int) *TensorSplitInfo {
	// Determine the best dimension to split
	splitDim := tps.chooseSplitDimension(tensor.Shape)
	
	if splitDim == -1 {
		// Cannot split this tensor
		return &TensorSplitInfo{
			TensorName:     tensor.Name,
			SplitDimension: -1,
			SplitStart:     0,
			SplitEnd:       0,
			OriginalShape:  tensor.Shape,
			SplitShape:     tensor.Shape,
		}
	}

	dimSize := tensor.Shape[splitDim]
	splitSize := dimSize / totalNodes
	remainder := dimSize % totalNodes

	// Calculate split bounds for this node
	splitStart := nodeIndex * splitSize
	splitEnd := splitStart + splitSize

	// Distribute remainder among first nodes
	if nodeIndex < remainder {
		splitStart += nodeIndex
		splitEnd += nodeIndex + 1
	} else {
		splitStart += remainder
		splitEnd += remainder
	}

	// Calculate split shape
	splitShape := make([]int, len(tensor.Shape))
	copy(splitShape, tensor.Shape)
	splitShape[splitDim] = splitEnd - splitStart

	return &TensorSplitInfo{
		TensorName:     tensor.Name,
		SplitDimension: splitDim,
		SplitStart:     splitStart,
		SplitEnd:       splitEnd,
		OriginalShape:  tensor.Shape,
		SplitShape:     splitShape,
	}
}

// chooseSplitDimension chooses the best dimension to split a tensor
func (tps *TensorPartitionStrategy) chooseSplitDimension(shape []int) int {
	if len(shape) < 2 {
		return -1 // Can't split 1D tensors effectively
	}

	// For attention tensors, prefer splitting the last dimension (head dimension)
	// For weight tensors, prefer splitting the largest dimension
	
	maxDim := -1
	maxSize := 0
	
	for i, size := range shape {
		if size > maxSize && size >= 4 { // Must be at least 4 to split effectively
			maxSize = size
			maxDim = i
		}
	}

	return maxDim
}

// calculateTensorResourceRequirements calculates resource requirements for tensor assignment
func (tps *TensorPartitionStrategy) calculateTensorResourceRequirements(memoryRequired int64, computeWeight float64, nodeCapabilities *NodeCapabilities) *ResourceRequirements {
	// CPU requirement - tensor operations are often compute intensive with nil guard
	var cpuReq *CPURequirement
	if nodeCapabilities.CPU != nil {
		cpuCores := int(math.Ceil(computeWeight * 4)) // Assume 4 cores per compute unit
		if cpuCores > nodeCapabilities.CPU.Cores {
			cpuCores = nodeCapabilities.CPU.Cores
		}
		
		cpuReq = &CPURequirement{
			Cores:       cpuCores,
			Utilization: math.Min(computeWeight, 1.0),
		}
	} else {
		// Default CPU requirements if nil
		cpuReq = &CPURequirement{
			Cores:       1,
			Utilization: 0.0,
		}
	}

	// Memory requirement with overhead for tensor operations and nil guard
	memoryWithOverhead := int64(float64(memoryRequired) * 1.3) // 30% overhead for tensor operations
	var memoryReq *MemoryRequirement
	if nodeCapabilities.Memory != nil && nodeCapabilities.Memory.TotalBytes > 0 {
		memoryReq = &MemoryRequirement{
			Bytes:       memoryWithOverhead,
			Utilization: float64(memoryWithOverhead) / float64(nodeCapabilities.Memory.TotalBytes),
		}
	} else {
		// Default memory requirements if nil
		memoryReq = &MemoryRequirement{
			Bytes:       memoryWithOverhead,
			Utilization: 0.0,
		}
	}

	// GPU requirement (tensor operations benefit significantly from GPU)
	var gpuReq *GPURequirement
	if nodeCapabilities.GPU != nil && nodeCapabilities.GPU.Count > 0 {
		gpuReq = &GPURequirement{
			Count:       1,
			MemoryBytes: memoryRequired,
			Utilization: computeWeight,
		}
	}

	// Network requirement (critical for tensor parallelism due to frequent all-reduce operations)
	networkReq := &NetworkRequirement{
		BandwidthGbps: computeWeight * 10.0, // High bandwidth needed
		LatencyMs:     0.5,                  // Low latency critical
	}

	// Storage requirement
	storageReq := &StorageRequirement{
		Bytes:     memoryRequired / 5, // Some temporary storage for splits
		IOPSRead:  int(computeWeight * 200),
		IOPSWrite: int(computeWeight * 100),
	}

	return &ResourceRequirements{
		CPU:     cpuReq,
		Memory:  memoryReq,
		GPU:     gpuReq,
		Network: networkReq,
		Storage: storageReq,
	}
}

// createCommunicationPlan creates a communication plan for tensor parallelism
func (tps *TensorPartitionStrategy) createCommunicationPlan(assignments []*NodeAssignment, analysis *ModelAnalysis) *CommunicationPlan {
	var connections []NodeConnection

	// For tensor parallelism, nodes need all-to-all communication for gradients and activations
	for i := 0; i < len(assignments); i++ {
		for j := 0; j < len(assignments); j++ {
			if i != j {
				connection := NodeConnection{
					From:           assignments[i].NodeID,
					To:             assignments[j].NodeID,
					ConnectionType: ConnectionTypeGradient,
					Parameters: map[string]interface{}{
						"operation":      tps.config.AggregationStrategy,
						"tensor_splits":  true,
						"high_priority":  true,
					},
				}
				connections = append(connections, connection)
			}
		}
	}

	return &CommunicationPlan{
		Topology:    TopologyAllToAll,
		Connections: connections,
		Parameters: map[string]interface{}{
			"aggregation_strategy": tps.config.AggregationStrategy,
			"communication_pattern": "all_to_all",
			"overlap_computation": true,
		},
	}
}

// calculateResourceCost calculates the estimated cost for tensor parallelism
func (tps *TensorPartitionStrategy) calculateResourceCost(assignments []*NodeAssignment, nodes []*NodeInfo, analysis *ModelAnalysis) *ResourceCost {
	computeCost := 0.0
	memoryCost := 0.0
	networkCost := 0.0
	storageCost := 0.0

	for _, assignment := range assignments {
		if assignment.Resources == nil {
			continue
		}

		// CPU cost
		if assignment.Resources.CPU != nil {
			computeCost += float64(assignment.Resources.CPU.Cores) * assignment.Resources.CPU.Utilization * 0.15 // Higher cost due to tensor ops
		}

		// Memory cost
		if assignment.Resources.Memory != nil {
			memoryCost += float64(assignment.Resources.Memory.Bytes) / (1024 * 1024 * 1024) * 0.02 // Higher memory cost
		}

		// GPU cost (significant for tensor parallelism)
		if assignment.Resources.GPU != nil {
			computeCost += float64(assignment.Resources.GPU.Count) * assignment.Resources.GPU.Utilization * 1.5 // Higher GPU cost
		}

		// Network cost (high due to frequent communication)
		if assignment.Resources.Network != nil {
			networkCost += assignment.Resources.Network.BandwidthGbps * 0.15 // Higher network cost
		}

		// Storage cost
		if assignment.Resources.Storage != nil {
			storageCost += float64(assignment.Resources.Storage.Bytes) / (1024 * 1024 * 1024) * 0.002
		}
	}

	// Add communication overhead cost
	communicationOverheadCost := networkCost * tps.config.CommunicationOverhead

	totalCost := computeCost + memoryCost + networkCost + storageCost + communicationOverheadCost

	return &ResourceCost{
		ComputeCost: computeCost,
		MemoryCost:  memoryCost,
		NetworkCost: networkCost + communicationOverheadCost,
		StorageCost: storageCost,
		TotalCost:   totalCost,
		Details: map[string]interface{}{
			"strategy":              tps.name,
			"nodes_used":           len(assignments),
			"communication_overhead": communicationOverheadCost,
			"aggregation_strategy": tps.config.AggregationStrategy,
		},
	}
}

// updatePerformanceMetrics updates performance metrics for the strategy
func (tps *TensorPartitionStrategy) updatePerformanceMetrics(executionTime time.Duration, nodeCount int, analysis *ModelAnalysis) {
	if tps.metrics.Performance == nil {
		tps.metrics.Performance = &PerformanceMetrics{
			ExecutionTimeMs:   []float64{},
			MemoryUsageBytes:  []int64{},
			NetworkBandwidth:  []float64{},
			QualityScore:      0.75,
			EfficiencyScore:   0.8,
		}
	}

	execTimeMs := float64(executionTime.Nanoseconds()) / 1e6
	tps.metrics.Performance.ExecutionTimeMs = append(tps.metrics.Performance.ExecutionTimeMs, execTimeMs)
	tps.metrics.Performance.MemoryUsageBytes = append(tps.metrics.Performance.MemoryUsageBytes, analysis.MemoryReqs.TotalRequired)
	
	// Estimate network bandwidth usage (high for tensor parallelism)
	estimatedBandwidth := float64(analysis.TensorInfo.TotalTensorSize) / (1024 * 1024 * 1024) // GB
	tps.metrics.Performance.NetworkBandwidth = append(tps.metrics.Performance.NetworkBandwidth, estimatedBandwidth)

	// Calculate average latency
	totalTime := 0.0
	for _, t := range tps.metrics.Performance.ExecutionTimeMs {
		totalTime += t
	}
	tps.metrics.AverageLatency = time.Duration(totalTime/float64(len(tps.metrics.Performance.ExecutionTimeMs))) * time.Millisecond

	// Update quality score based on tensor utilization
	splittableTensors := len(analysis.TensorInfo.SplittableTensors)
	totalTensors := len(analysis.TensorInfo.AttentionTensors) + len(analysis.TensorInfo.WeightTensors)
	if totalTensors > 0 {
		tensorUtilization := float64(splittableTensors) / float64(totalTensors)
		tps.metrics.Performance.QualityScore = 0.6 + (tensorUtilization * 0.4)
	}

	// Efficiency score based on parallelization effectiveness
	parallelizationFactor := float64(nodeCount) / float64(analysis.MemoryReqs.RecommendedNodes)
	tps.metrics.Performance.EfficiencyScore = math.Min(parallelizationFactor, 1.0)

	// Keep only last 100 measurements
	if len(tps.metrics.Performance.ExecutionTimeMs) > 100 {
		tps.metrics.Performance.ExecutionTimeMs = tps.metrics.Performance.ExecutionTimeMs[len(tps.metrics.Performance.ExecutionTimeMs)-100:]
		tps.metrics.Performance.MemoryUsageBytes = tps.metrics.Performance.MemoryUsageBytes[len(tps.metrics.Performance.MemoryUsageBytes)-100:]
		tps.metrics.Performance.NetworkBandwidth = tps.metrics.Performance.NetworkBandwidth[len(tps.metrics.Performance.NetworkBandwidth)-100:]
	}
}

// Helper function to generate tensor partition plan IDs
func generateTensorPartitionPlanID() string {
	return fmt.Sprintf("tensor_%d", time.Now().UnixNano())
}
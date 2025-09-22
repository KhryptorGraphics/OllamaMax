package partitioning

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"
)

// LayerPartitionStrategy implements layer-wise partitioning
type LayerPartitionStrategy struct {
	name           string
	analyzer       *ModelAnalyzer
	memoryOptimizer *MemoryOptimizer
	metrics        *StrategyMetrics
	config         *LayerStrategyConfig
}

// LayerStrategyConfig contains configuration for layer partitioning
type LayerStrategyConfig struct {
	AllowPartialLayers    bool    `json:"allow_partial_layers"`    // Allow splitting layers across nodes
	MemoryBuffer         float64 `json:"memory_buffer"`           // Memory buffer percentage (0.1 = 10%)
	LoadBalanceThreshold float64 `json:"load_balance_threshold"`  // Threshold for load balancing
	PreferSequentialExecution bool `json:"prefer_sequential"`      // Prefer sequential execution
	OptimizeForLatency   bool    `json:"optimize_for_latency"`    // Optimize for latency vs throughput
}

// NewLayerPartitionStrategy creates a new layer partition strategy
func NewLayerPartitionStrategy(analyzer *ModelAnalyzer, optimizer *MemoryOptimizer) *LayerPartitionStrategy {
	return &LayerPartitionStrategy{
		name:            "layer_parallelism",
		analyzer:        analyzer,
		memoryOptimizer: optimizer,
		metrics: &StrategyMetrics{
			UsageCount:   0,
			SuccessCount: 0,
			FailureCount: 0,
			Performance:  &PerformanceMetrics{},
			CustomMetrics: make(map[string]interface{}),
		},
		config: &LayerStrategyConfig{
			AllowPartialLayers:    false,
			MemoryBuffer:         0.15, // 15% buffer
			LoadBalanceThreshold: 0.2,  // 20% imbalance threshold
			PreferSequentialExecution: true,
			OptimizeForLatency:   false,
		},
	}
}

// GetName returns the strategy name
func (lps *LayerPartitionStrategy) GetName() string {
	return lps.name
}

// GetMetrics returns strategy metrics
func (lps *LayerPartitionStrategy) GetMetrics() *StrategyMetrics {
	return lps.metrics
}

// CanHandle determines if this strategy can handle the given request
func (lps *LayerPartitionStrategy) CanHandle(req *PartitionRequest) bool {
	if req == nil || req.Model == nil {
		return false
	}

	// Layer partitioning is suitable for:
	// 1. Models with clear layer structure (transformer models)
	// 2. Medium to large models that benefit from distribution
	// 3. When memory constraints require distribution but tensor parallelism isn't needed
	
	modelSize := req.Model.Parameters
	if modelSize < 1_000_000_000 { // Less than 1B parameters
		return false // Too small for layer parallelism
	}

	// Check if we have layer information
	if req.ModelAnalysis != nil && req.ModelAnalysis.LayerInfo != nil {
		layerInfo := req.ModelAnalysis.LayerInfo
		if layerInfo.TotalLayers < 4 { // Need sufficient layers to distribute
			return false
		}
	}

	return true
}

// Partition creates a partition plan using layer-wise distribution
func (lps *LayerPartitionStrategy) Partition(ctx context.Context, req *PartitionRequest, nodes []*NodeInfo) (*PartitionPlan, error) {
	startTime := time.Now()
	lps.metrics.UsageCount++

	// Validate inputs
	if err := lps.validateInputs(req, nodes); err != nil {
		lps.metrics.FailureCount++
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// Analyze model if not already done
	modelAnalysis := req.ModelAnalysis
	if modelAnalysis == nil {
		var err error
		modelAnalysis, err = lps.analyzer.AnalyzeModel(req.Model.Path, req.Model.Details)
		if err != nil {
			lps.metrics.FailureCount++
			return nil, fmt.Errorf("model analysis failed: %w", err)
		}
	}

	// Filter and rank nodes
	suitableNodes, err := lps.selectSuitableNodes(nodes, modelAnalysis, req.Options)
	if err != nil {
		lps.metrics.FailureCount++
		return nil, fmt.Errorf("node selection failed: %w", err)
	}

	// Create layer assignment plan
	assignments, err := lps.createLayerAssignments(modelAnalysis, suitableNodes, req.Options)
	if err != nil {
		lps.metrics.FailureCount++
		return nil, fmt.Errorf("layer assignment failed: %w", err)
	}

	// Create communication plan
	communicationPlan := lps.createCommunicationPlan(assignments)

	// Calculate resource costs
	estimatedCost := lps.calculateResourceCost(assignments, suitableNodes)

	// Create the final partition plan
	plan := &PartitionPlan{
		ID:            generatePartitionPlanID(),
		TaskID:        req.TaskID,
		Strategy:      lps.name,
		Assignments:   assignments,
		Communication: communicationPlan,
		EstimatedCost: estimatedCost,
		CreatedAt:     time.Now(),
		Metadata: map[string]interface{}{
			"total_layers":       modelAnalysis.LayerInfo.TotalLayers,
			"nodes_used":        len(assignments),
			"optimization_target": req.Options.OptimizeFor,
			"execution_time_ms": float64(time.Since(startTime).Nanoseconds()) / 1e6,
		},
	}

	// Update metrics
	lps.metrics.SuccessCount++
	lps.metrics.LastUsed = time.Now()
	lps.updatePerformanceMetrics(time.Since(startTime), len(suitableNodes), modelAnalysis)

	return plan, nil
}

// validateInputs validates the partition request and available nodes
func (lps *LayerPartitionStrategy) validateInputs(req *PartitionRequest, nodes []*NodeInfo) error {
	if err := ValidatePartitionRequest(req); err != nil {
		return err
	}

	if len(nodes) == 0 {
		return fmt.Errorf("no nodes available")
	}

	// Validate node capabilities
	for i, node := range nodes {
		if err := ValidateNodeInfo(node); err != nil {
			return fmt.Errorf("invalid node %d: %w", i, err)
		}
	}

	return nil
}

// selectSuitableNodes filters and ranks nodes based on their suitability for layer partitioning
func (lps *LayerPartitionStrategy) selectSuitableNodes(nodes []*NodeInfo, analysis *ModelAnalysis, options *PartitionOptions) ([]*NodeInfo, error) {
	var suitableNodes []*NodeInfo

	minMemoryRequired := analysis.MemoryReqs.MinNodeMemory
	if minMemoryRequired == 0 {
		// Fallback calculation if not available
		minMemoryRequired = analysis.MemoryReqs.TotalRequired / int64(len(nodes))
	}

	// Filter nodes based on basic requirements
	for _, node := range nodes {
		if node.Status != NodeStatusActive {
			continue
		}

		// Skip nodes without capability information
		if node.Capabilities == nil || node.Capabilities.Memory == nil {
			continue
		}

		// Check memory requirements
		availableMemory := node.Capabilities.Memory.AvailableBytes
		if availableMemory == 0 {
			continue // No available memory
		}
		
		memoryUtilization := float64(minMemoryRequired) / float64(availableMemory)
		
		if memoryUtilization > (1.0 - lps.config.MemoryBuffer) {
			continue // Not enough memory
		}

		// Check if node is overloaded
		if node.CurrentLoad != nil {
			if node.CurrentLoad.MemoryUtilization > 0.8 || 
			   node.CurrentLoad.CPUUtilization > 0.9 {
				continue // Node is overloaded
			}
		}

		suitableNodes = append(suitableNodes, node)
	}

	if len(suitableNodes) == 0 {
		return nil, fmt.Errorf("no suitable nodes found")
	}

	// Rank nodes based on capabilities and current load
	lps.rankNodesBySuitability(suitableNodes, analysis)

	// Limit nodes based on options
	maxNodes := options.MaxNodes
	if maxNodes == 0 || maxNodes > len(suitableNodes) {
		maxNodes = len(suitableNodes)
	}

	if options.MinNodes > 0 && len(suitableNodes) < options.MinNodes {
		return nil, fmt.Errorf("insufficient nodes: have %d, need at least %d", len(suitableNodes), options.MinNodes)
	}

	return suitableNodes[:maxNodes], nil
}

// rankNodesBySuitability ranks nodes based on their suitability for layer partitioning
func (lps *LayerPartitionStrategy) rankNodesBySuitability(nodes []*NodeInfo, analysis *ModelAnalysis) {
	sort.Slice(nodes, func(i, j int) bool {
		scoreI := lps.calculateNodeScore(nodes[i], analysis)
		scoreJ := lps.calculateNodeScore(nodes[j], analysis)
		return scoreI > scoreJ
	})
}

// calculateNodeScore calculates a suitability score for a node
func (lps *LayerPartitionStrategy) calculateNodeScore(node *NodeInfo, analysis *ModelAnalysis) float64 {
	score := 0.0

	// Memory score (40% weight)
	if node.Capabilities != nil && node.Capabilities.Memory != nil && node.Capabilities.Memory.AvailableBytes > 0 {
		minNodeMemory := analysis.MemoryReqs.MinNodeMemory
		if minNodeMemory <= 0 {
			minNodeMemory = 1 // Prevent division by zero
		}
		memoryRatio := float64(node.Capabilities.Memory.AvailableBytes) / float64(minNodeMemory)
		memoryScore := math.Min(memoryRatio, 2.0) / 2.0 // Cap at 2x required memory
		score += memoryScore * 0.4
	}

	// CPU score (25% weight)
	if node.Capabilities != nil && node.Capabilities.CPU != nil {
		cpuCores := node.Capabilities.CPU.Cores
		if cpuCores <= 0 {
			cpuCores = 1 // Guard against zero/negative cores
		}
		cpuScore := math.Min(float64(cpuCores)/8.0, 1.0) // Normalize to 8 cores
		score += cpuScore * 0.25
	}

	// Network score (20% weight)
	if node.Capabilities != nil && node.Capabilities.Network != nil && node.Capabilities.Network.Bandwidth > 0 {
		networkScore := math.Min(node.Capabilities.Network.Bandwidth/10.0, 1.0) // Normalize to 10 Gbps
		score += networkScore * 0.20
	}

	// Current load penalty (15% weight)
	if node.CurrentLoad != nil {
		// Clamp utilization values to [0, 1] range
		memUtil := math.Max(0, math.Min(1, node.CurrentLoad.MemoryUtilization))
		cpuUtil := math.Max(0, math.Min(1, node.CurrentLoad.CPUUtilization))
		loadPenalty := (memUtil + cpuUtil) / 2.0
		score += (1.0 - loadPenalty) * 0.15
	} else {
		score += 0.15 // Full score if no load information
	}

	return score
}

// createLayerAssignments creates layer assignments for nodes
func (lps *LayerPartitionStrategy) createLayerAssignments(analysis *ModelAnalysis, nodes []*NodeInfo, options *PartitionOptions) ([]*NodeAssignment, error) {
	layerInfo := analysis.LayerInfo
	totalLayers := layerInfo.TotalLayers

	// Calculate optimal layer distribution
	distribution := lps.calculateLayerDistribution(totalLayers, len(nodes), layerInfo.LayerSizes, nodes)

	// Structure to track assignment details
	type assignmentDetail struct {
		node   *NodeInfo
		start  int
		end    int
	}
	
	var assignmentDetails []assignmentDetail
	var assignments []*NodeAssignment
	var previousNodeID string

	// Build assignment details first
	layerIndex := 0
	for i, node := range nodes {
		if i >= len(distribution) {
			break
		}

		layerCount := distribution[i]
		if layerCount == 0 {
			continue
		}

		// Track this assignment's details
		detail := assignmentDetail{
			node:  node,
			start: layerIndex,
			end:   layerIndex + layerCount - 1,
		}
		assignmentDetails = append(assignmentDetails, detail)
		layerIndex += layerCount
	}

	// Validate that all layers are assigned
	if layerIndex != totalLayers {
		return nil, fmt.Errorf("layer assignment mismatch: assigned %d of %d layers", layerIndex, totalLayers)
	}

	// Now create assignments with correct isFirst/isLast flags
	for i, detail := range assignmentDetails {
		isFirst := i == 0
		isLast := i == len(assignmentDetails)-1
		
		// Create layer assignment with proper dependency tracking
		assignment := lps.createLayerAssignmentWithDependency(
			detail.node, 
			detail.start, 
			detail.end, 
			layerInfo, 
			isFirst, 
			isLast, 
			previousNodeID,
		)
		assignments = append(assignments, assignment)
		
		// Track previous node ID for dependency chain
		previousNodeID = detail.node.ID
	}

	// Balance load if needed
	if options.LoadBalance {
		lps.balanceLayerAssignments(assignments, layerInfo)
	}

	return assignments, nil
}

// calculateLayerDistribution calculates how to distribute layers across nodes
func (lps *LayerPartitionStrategy) calculateLayerDistribution(totalLayers, numNodes int, layerSizes []int64, nodes []*NodeInfo) []int {
	distribution := make([]int, numNodes)

	if numNodes == 1 {
		distribution[0] = totalLayers
		return distribution
	}

	// Calculate total memory capacity with nil guards
	totalCapacity := int64(0)
	nodeCapacities := make([]int64, numNodes)
	
	for i, node := range nodes {
		if i < numNodes && node.Capabilities != nil && node.Capabilities.Memory != nil {
			capacity := node.Capabilities.Memory.AvailableBytes
			if capacity <= 0 {
				capacity = 1 // Minimum capacity to avoid zero
			}
			nodeCapacities[i] = capacity
			totalCapacity += capacity
		}
	}
	
	// Guard against zero total capacity
	if totalCapacity <= 0 {
		// Fallback to equal distribution if no capacity info
		for i := 0; i < numNodes; i++ {
			distribution[i] = totalLayers / numNodes
			if i < totalLayers%numNodes {
				distribution[i]++
			}
		}
		return distribution
	}

	// Distribute layers based on memory capacity
	remainingLayers := totalLayers
	totalMemory := int64(0)
	for _, size := range layerSizes {
		totalMemory += size
	}

	for i := 0; i < numNodes && remainingLayers > 0; i++ {
		if i == numNodes-1 {
			// Last node gets all remaining layers
			distribution[i] = remainingLayers
		} else {
			// Distribute proportionally to memory capacity with guard
			var ratio float64
			if totalCapacity > 0 {
				ratio = float64(nodeCapacities[i]) / float64(totalCapacity)
			} else {
				ratio = 1.0 / float64(numNodes) // Equal distribution fallback
			}
			layerCount := int(math.Round(float64(totalLayers) * ratio))
			
			// Ensure at least one layer per node (except possibly the last)
			if layerCount == 0 && remainingLayers > numNodes-i {
				layerCount = 1
			}
			
			// Don't exceed remaining layers
			if layerCount > remainingLayers {
				layerCount = remainingLayers
			}
			
			distribution[i] = layerCount
			remainingLayers -= layerCount
		}
	}

	return distribution
}

// createLayerAssignmentWithDependency creates a layer assignment for a node with proper dependency tracking
func (lps *LayerPartitionStrategy) createLayerAssignmentWithDependency(node *NodeInfo, startLayer, endLayer int, layerInfo *LayerAnalysis, isFirst, isLast bool, previousNodeID string) *NodeAssignment {
	layerCount := endLayer - startLayer + 1
	
	// Calculate memory and compute requirements
	memoryRequired := int64(0)
	computeWeight := 0.0
	layerIndices := make([]int, layerCount)
	layerNames := make([]string, layerCount)

	for i := 0; i < layerCount; i++ {
		layerIdx := startLayer + i
		layerIndices[i] = layerIdx
		layerNames[i] = fmt.Sprintf("layer_%d", layerIdx)
		
		if layerIdx < len(layerInfo.LayerSizes) {
			memoryRequired += layerInfo.LayerSizes[layerIdx]
		}
		if layerIdx < len(layerInfo.LayerWeights) {
			computeWeight += layerInfo.LayerWeights[layerIdx]
		}
	}

	// Create work assignment
	layerAssignment := &LayerAssignment{
		LayerIndices:   layerIndices,
		LayerNames:     layerNames,
		TotalLayers:    layerCount,
		StartIndex:     startLayer,
		EndIndex:       endLayer,
		MemoryRequired: memoryRequired,
		ComputeWeight:  computeWeight,
	}

	workAssignment := &WorkAssignment{
		Layers: []LayerAssignment{*layerAssignment},
	}

	// Calculate resource requirements
	resources := lps.calculateResourceRequirements(memoryRequired, computeWeight, node.Capabilities)

	// Determine dependencies
	var dependencies []string
	if !isFirst && previousNodeID != "" {
		// Depends on previous node's assignment
		dependencies = append(dependencies, previousNodeID)
	}

	// Create the node assignment
	assignment := &NodeAssignment{
		NodeID:      node.ID,
		Role:        RoleWorker,
		WorkType:    WorkTypeLayers,
		Assignment:  workAssignment,
		Resources:   resources,
		Dependencies: dependencies,
	}

	return assignment
}

// createLayerAssignment creates a layer assignment for a node (backward compatibility)
func (lps *LayerPartitionStrategy) createLayerAssignment(node *NodeInfo, startLayer, endLayer int, layerInfo *LayerAnalysis, isFirst, isLast bool) *NodeAssignment {
	// Call the new method with empty previous node ID for backward compatibility
	return lps.createLayerAssignmentWithDependency(node, startLayer, endLayer, layerInfo, isFirst, isLast, "")
}

// balanceLayerAssignments balances the computational load across assignments
func (lps *LayerPartitionStrategy) balanceLayerAssignments(assignments []*NodeAssignment, layerInfo *LayerAnalysis) {
	if len(assignments) <= 1 {
		return
	}

	// Calculate current load distribution
	loads := make([]float64, len(assignments))
	totalLoad := 0.0
	
	for i, assignment := range assignments {
		if len(assignment.Assignment.Layers) > 0 {
			loads[i] = assignment.Assignment.Layers[0].ComputeWeight
			totalLoad += loads[i]
		}
	}

	// Check if balancing is needed
	avgLoad := totalLoad / float64(len(assignments))
	maxImbalance := 0.0
	for _, load := range loads {
		imbalance := math.Abs(load-avgLoad) / avgLoad
		if imbalance > maxImbalance {
			maxImbalance = imbalance
		}
	}

	if maxImbalance <= lps.config.LoadBalanceThreshold {
		return // Already balanced
	}

	// Perform load balancing by adjusting layer boundaries
	// This is a simplified version - in practice, would need more sophisticated algorithms
	lps.adjustLayerBoundaries(assignments, layerInfo, avgLoad)
}

// adjustLayerBoundaries adjusts layer boundaries to balance load
func (lps *LayerPartitionStrategy) adjustLayerBoundaries(assignments []*NodeAssignment, layerInfo *LayerAnalysis, targetLoad float64) {
	// Simple greedy approach to balance load
	for i := 0; i < len(assignments)-1; i++ {
		currentAssignment := assignments[i]
		nextAssignment := assignments[i+1]

		if len(currentAssignment.Assignment.Layers) == 0 || len(nextAssignment.Assignment.Layers) == 0 {
			continue
		}

		currentLayer := &currentAssignment.Assignment.Layers[0]
		nextLayer := &nextAssignment.Assignment.Layers[0]

		currentLoad := currentLayer.ComputeWeight
		nextLoad := nextLayer.ComputeWeight

		// If current node is overloaded and next is underloaded, transfer layers
		if currentLoad > targetLoad*1.2 && nextLoad < targetLoad*0.8 {
			// Transfer one layer from current to next (simplified)
			if currentLayer.EndIndex > currentLayer.StartIndex {
				// Move the last layer from current to next
				transferLayerIdx := currentLayer.EndIndex
				currentLayer.EndIndex--
				currentLayer.TotalLayers--
				
				// Update memory and compute weight
				if transferLayerIdx < len(layerInfo.LayerSizes) {
					transferMemory := layerInfo.LayerSizes[transferLayerIdx]
					transferWeight := 0.0
					if transferLayerIdx < len(layerInfo.LayerWeights) {
						transferWeight = layerInfo.LayerWeights[transferLayerIdx]
					}
					
					currentLayer.MemoryRequired -= transferMemory
					currentLayer.ComputeWeight -= transferWeight
					
					nextLayer.StartIndex = transferLayerIdx
					nextLayer.TotalLayers++
					nextLayer.MemoryRequired += transferMemory
					nextLayer.ComputeWeight += transferWeight
				}

				// Update layer indices and names
				currentLayer.LayerIndices = currentLayer.LayerIndices[:len(currentLayer.LayerIndices)-1]
				currentLayer.LayerNames = currentLayer.LayerNames[:len(currentLayer.LayerNames)-1]
				
				nextLayer.LayerIndices = append([]int{transferLayerIdx}, nextLayer.LayerIndices...)
				nextLayer.LayerNames = append([]string{fmt.Sprintf("layer_%d", transferLayerIdx)}, nextLayer.LayerNames...)
			}
		}
	}
}

// calculateResourceRequirements calculates resource requirements for a layer assignment
func (lps *LayerPartitionStrategy) calculateResourceRequirements(memoryRequired int64, computeWeight float64, nodeCapabilities *NodeCapabilities) *ResourceRequirements {
	// CPU requirement based on compute weight with nil guard
	var cpuReq *CPURequirement
	if nodeCapabilities.CPU != nil {
		cpuCores := int(math.Ceil(computeWeight))
		if cpuCores > nodeCapabilities.CPU.Cores {
			cpuCores = nodeCapabilities.CPU.Cores
		}
		
		cpuReq = &CPURequirement{
			Cores:       cpuCores,
			Utilization: math.Min(computeWeight/float64(nodeCapabilities.CPU.Cores), 1.0),
		}
	} else {
		// Default CPU requirements if nil
		cpuReq = &CPURequirement{
			Cores:       1,
			Utilization: 0.0,
		}
	}

	// Memory requirement with buffer and nil guard
	memoryWithBuffer := int64(float64(memoryRequired) * (1.0 + lps.config.MemoryBuffer))
	var memoryReq *MemoryRequirement
	if nodeCapabilities.Memory != nil && nodeCapabilities.Memory.TotalBytes > 0 {
		memoryReq = &MemoryRequirement{
			Bytes:       memoryWithBuffer,
			Utilization: float64(memoryWithBuffer) / float64(nodeCapabilities.Memory.TotalBytes),
		}
	} else {
		// Default memory requirements if nil
		memoryReq = &MemoryRequirement{
			Bytes:       memoryWithBuffer,
			Utilization: 0.0,
		}
	}

	// GPU requirement (if available)
	var gpuReq *GPURequirement
	if nodeCapabilities.GPU != nil && nodeCapabilities.GPU.Count > 0 {
		gpuMemoryNeeded := memoryRequired / 2 // Assume half the memory goes to GPU
		gpuReq = &GPURequirement{
			Count:       1,
			MemoryBytes: gpuMemoryNeeded,
			Utilization: computeWeight * 0.8, // Assume 80% of compute goes to GPU
		}
	}

	// Network requirement
	networkReq := &NetworkRequirement{
		BandwidthGbps: computeWeight * 0.5, // Rough estimate
		LatencyMs:     1.0,                  // Target 1ms latency
	}

	// Storage requirement
	storageReq := &StorageRequirement{
		Bytes:     memoryRequired / 10, // Assume some temporary storage needed
		IOPSRead:  int(computeWeight * 100),
		IOPSWrite: int(computeWeight * 50),
	}

	return &ResourceRequirements{
		CPU:     cpuReq,
		Memory:  memoryReq,
		GPU:     gpuReq,
		Network: networkReq,
		Storage: storageReq,
	}
}

// createCommunicationPlan creates a communication plan for layer-wise partitioning
func (lps *LayerPartitionStrategy) createCommunicationPlan(assignments []*NodeAssignment) *CommunicationPlan {
	var connections []NodeConnection

	// For layer parallelism, communication is sequential between adjacent layers
	for i := 0; i < len(assignments)-1; i++ {
		connection := NodeConnection{
			From:           assignments[i].NodeID,
			To:             assignments[i+1].NodeID,
			ConnectionType: ConnectionTypePipeline,
			Parameters: map[string]interface{}{
				"order":    i,
				"dataflow": "sequential",
			},
		}
		connections = append(connections, connection)
	}

	return &CommunicationPlan{
		Topology:    TopologyPointToPoint,
		Connections: connections,
		Parameters: map[string]interface{}{
			"execution_order": "sequential",
			"buffer_size":     "auto",
		},
	}
}

// calculateResourceCost calculates the estimated cost of the partition plan
func (lps *LayerPartitionStrategy) calculateResourceCost(assignments []*NodeAssignment, nodes []*NodeInfo) *ResourceCost {
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
			computeCost += float64(assignment.Resources.CPU.Cores) * assignment.Resources.CPU.Utilization * 0.1 // $0.1 per core-hour
		}

		// Memory cost
		if assignment.Resources.Memory != nil {
			memoryCost += float64(assignment.Resources.Memory.Bytes) / (1024 * 1024 * 1024) * 0.01 // $0.01 per GB-hour
		}

		// GPU cost
		if assignment.Resources.GPU != nil {
			computeCost += float64(assignment.Resources.GPU.Count) * assignment.Resources.GPU.Utilization * 1.0 // $1.0 per GPU-hour
		}

		// Network cost
		if assignment.Resources.Network != nil {
			networkCost += assignment.Resources.Network.BandwidthGbps * 0.05 // $0.05 per Gbps-hour
		}

		// Storage cost
		if assignment.Resources.Storage != nil {
			storageCost += float64(assignment.Resources.Storage.Bytes) / (1024 * 1024 * 1024) * 0.001 // $0.001 per GB-hour
		}
	}

	totalCost := computeCost + memoryCost + networkCost + storageCost

	return &ResourceCost{
		ComputeCost: computeCost,
		MemoryCost:  memoryCost,
		NetworkCost: networkCost,
		StorageCost: storageCost,
		TotalCost:   totalCost,
		Details: map[string]interface{}{
			"strategy":      lps.name,
			"nodes_used":   len(assignments),
			"cost_per_node": totalCost / float64(len(assignments)),
		},
	}
}

// updatePerformanceMetrics updates the performance metrics
func (lps *LayerPartitionStrategy) updatePerformanceMetrics(executionTime time.Duration, nodeCount int, analysis *ModelAnalysis) {
	if lps.metrics.Performance == nil {
		lps.metrics.Performance = &PerformanceMetrics{
			ExecutionTimeMs:   []float64{},
			MemoryUsageBytes:  []int64{},
			NetworkBandwidth:  []float64{},
			QualityScore:      0.8,
			EfficiencyScore:   0.7,
		}
	}

	execTimeMs := float64(executionTime.Nanoseconds()) / 1e6
	lps.metrics.Performance.ExecutionTimeMs = append(lps.metrics.Performance.ExecutionTimeMs, execTimeMs)
	lps.metrics.Performance.MemoryUsageBytes = append(lps.metrics.Performance.MemoryUsageBytes, analysis.MemoryReqs.TotalRequired)
	
	// Calculate average latency
	totalTime := 0.0
	for _, t := range lps.metrics.Performance.ExecutionTimeMs {
		totalTime += t
	}
	lps.metrics.AverageLatency = time.Duration(totalTime/float64(len(lps.metrics.Performance.ExecutionTimeMs))) * time.Millisecond

	// Update quality score based on node utilization and balance
	efficiencyFactor := math.Min(float64(nodeCount)/float64(analysis.MemoryReqs.RecommendedNodes), 1.0)
	lps.metrics.Performance.QualityScore = 0.7 + (efficiencyFactor * 0.3)
	lps.metrics.Performance.EfficiencyScore = efficiencyFactor

	// Keep only last 100 measurements
	if len(lps.metrics.Performance.ExecutionTimeMs) > 100 {
		lps.metrics.Performance.ExecutionTimeMs = lps.metrics.Performance.ExecutionTimeMs[len(lps.metrics.Performance.ExecutionTimeMs)-100:]
		lps.metrics.Performance.MemoryUsageBytes = lps.metrics.Performance.MemoryUsageBytes[len(lps.metrics.Performance.MemoryUsageBytes)-100:]
	}
}

// Helper function to generate partition plan IDs
func generatePartitionPlanID() string {
	return fmt.Sprintf("layer_%d", time.Now().UnixNano())
}
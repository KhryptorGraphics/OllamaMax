package partitioning

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"
)

// PipelinePartitionStrategy implements pipeline parallelism partitioning
type PipelinePartitionStrategy struct {
	name           string
	analyzer       *ModelAnalyzer
	memoryOptimizer *MemoryOptimizer
	metrics        *StrategyMetrics
	config         *PipelineStrategyConfig
}

// PipelineStrategyConfig contains configuration for pipeline partitioning
type PipelineStrategyConfig struct {
	MinStages                int     `json:"min_stages"`                 // Minimum pipeline stages
	MaxStages                int     `json:"max_stages"`                 // Maximum pipeline stages
	StageBalanceThreshold    float64 `json:"stage_balance_threshold"`    // Threshold for stage balance
	MicroBatchSize           int     `json:"micro_batch_size"`           // Micro-batch size for pipeline
	BubbleOptimization       bool    `json:"bubble_optimization"`        // Enable pipeline bubble optimization
	OverlapCommunication     bool    `json:"overlap_communication"`      // Overlap computation and communication
	GradientAccumulation     int     `json:"gradient_accumulation"`      // Gradient accumulation steps
	CheckpointActivations    bool    `json:"checkpoint_activations"`     // Checkpoint activations to save memory
	AdaptiveBatching         bool    `json:"adaptive_batching"`          // Adaptive micro-batch sizing
}

// PipelineStage represents a pipeline stage
type PipelineStage struct {
	StageIndex       int       `json:"stage_index"`
	LayerRange       []int     `json:"layer_range"`     // [start, end]
	ExpectedMemory   int64     `json:"expected_memory"`
	ExpectedCompute  float64   `json:"expected_compute"`
	InputActivations []string  `json:"input_activations"`
	OutputActivations []string `json:"output_activations"`
	Dependencies     []int     `json:"dependencies"`    // Dependent stage indices
	IsFirst          bool      `json:"is_first"`
	IsLast           bool      `json:"is_last"`
}

// NewPipelinePartitionStrategy creates a new pipeline partition strategy
func NewPipelinePartitionStrategy(analyzer *ModelAnalyzer, optimizer *MemoryOptimizer) *PipelinePartitionStrategy {
	return &PipelinePartitionStrategy{
		name:            "pipeline_parallelism",
		analyzer:        analyzer,
		memoryOptimizer: optimizer,
		metrics: &StrategyMetrics{
			UsageCount:   0,
			SuccessCount: 0,
			FailureCount: 0,
			Performance:  &PerformanceMetrics{},
			CustomMetrics: make(map[string]interface{}),
		},
		config: &PipelineStrategyConfig{
			MinStages:                2,
			MaxStages:                16,
			StageBalanceThreshold:    0.15, // 15% imbalance tolerance
			MicroBatchSize:           4,
			BubbleOptimization:       true,
			OverlapCommunication:     true,
			GradientAccumulation:     1,
			CheckpointActivations:    true,
			AdaptiveBatching:         false,
		},
	}
}

// GetName returns the strategy name
func (pps *PipelinePartitionStrategy) GetName() string {
	return pps.name
}

// GetMetrics returns strategy metrics
func (pps *PipelinePartitionStrategy) GetMetrics() *StrategyMetrics {
	return pps.metrics
}

// CanHandle determines if this strategy can handle the given request
func (pps *PipelinePartitionStrategy) CanHandle(req *PartitionRequest) bool {
	if req == nil || req.Model == nil {
		return false
	}

	// Pipeline parallelism is suitable for:
	// 1. Large models with many sequential layers
	// 2. Models that don't fit in memory of a single node
	// 3. When you have multiple nodes but limited high-speed interconnect
	// 4. Models with clear sequential structure (like transformers)

	modelSize := req.Model.Parameters
	if modelSize < 3_000_000_000 { // Less than 3B parameters
		return false // Too small to benefit from pipeline parallelism
	}

	// Check if we have layer information for staging
	if req.ModelAnalysis != nil && req.ModelAnalysis.LayerInfo != nil {
		layerInfo := req.ModelAnalysis.LayerInfo
		if layerInfo.TotalLayers < 6 { // Need sufficient layers for meaningful stages
			return false
		}

		// Check if we have stage information
		if req.ModelAnalysis.StageInfo != nil {
			stageInfo := req.ModelAnalysis.StageInfo
			if stageInfo.OptimalStages < pps.config.MinStages {
				return false
			}
		}
	}

	return true
}

// Partition creates a partition plan using pipeline parallelism
func (pps *PipelinePartitionStrategy) Partition(ctx context.Context, req *PartitionRequest, nodes []*NodeInfo) (*PartitionPlan, error) {
	startTime := time.Now()
	pps.metrics.UsageCount++

	// Validate inputs
	if err := pps.validateInputs(req, nodes); err != nil {
		pps.metrics.FailureCount++
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// Analyze model if not already done
	modelAnalysis := req.ModelAnalysis
	if modelAnalysis == nil {
		var err error
		modelAnalysis, err = pps.analyzer.AnalyzeModel(req.Model.Path, req.Model.Details)
		if err != nil {
			pps.metrics.FailureCount++
			return nil, fmt.Errorf("model analysis failed: %w", err)
		}
	}

	// Select and rank nodes for pipeline parallelism
	suitableNodes, err := pps.selectSuitableNodes(nodes, modelAnalysis, req.Options)
	if err != nil {
		pps.metrics.FailureCount++
		return nil, fmt.Errorf("node selection failed: %w", err)
	}

	// Create pipeline stages
	stages, err := pps.createPipelineStages(modelAnalysis, len(suitableNodes))
	if err != nil {
		pps.metrics.FailureCount++
		return nil, fmt.Errorf("pipeline stage creation failed: %w", err)
	}

	// Optimize pipeline for bubble reduction
	if pps.config.BubbleOptimization {
		pps.optimizePipelineStages(stages, modelAnalysis)
	}

	// Create stage assignments for nodes
	assignments, err := pps.createStageAssignments(stages, suitableNodes, modelAnalysis)
	if err != nil {
		pps.metrics.FailureCount++
		return nil, fmt.Errorf("stage assignment failed: %w", err)
	}

	// Recreate communication plan using the updated stages after optimization
	// This ensures the communication topology matches the final stage boundaries
	communicationPlan := pps.createPipelineCommunicationPlan(assignments, stages)

	// Calculate resource costs
	estimatedCost := pps.calculatePipelineResourceCost(assignments, suitableNodes, stages)

	// Create the final partition plan
	plan := &PartitionPlan{
		ID:            generatePipelinePartitionPlanID(),
		TaskID:        req.TaskID,
		Strategy:      pps.name,
		Assignments:   assignments,
		Communication: communicationPlan,
		EstimatedCost: estimatedCost,
		CreatedAt:     time.Now(),
		Metadata: map[string]interface{}{
			"pipeline_stages":       len(stages),
			"nodes_used":           len(assignments),
			"micro_batch_size":     pps.config.MicroBatchSize,
			"bubble_optimization":  pps.config.BubbleOptimization,
			"optimization_target":  req.Options.OptimizeFor,
			"execution_time_ms":    float64(time.Since(startTime).Nanoseconds()) / 1e6,
			"gradient_accumulation": pps.config.GradientAccumulation,
		},
	}

	// Update metrics
	pps.metrics.SuccessCount++
	pps.metrics.LastUsed = time.Now()
	pps.updatePerformanceMetrics(time.Since(startTime), len(suitableNodes), modelAnalysis)

	return plan, nil
}

// validateInputs validates the partition request and available nodes
func (pps *PipelinePartitionStrategy) validateInputs(req *PartitionRequest, nodes []*NodeInfo) error {
	if err := ValidatePartitionRequest(req); err != nil {
		return err
	}

	if len(nodes) < pps.config.MinStages {
		return fmt.Errorf("insufficient nodes for pipeline: have %d, need at least %d", len(nodes), pps.config.MinStages)
	}

	// Validate node capabilities
	for i, node := range nodes {
		if err := ValidateNodeInfo(node); err != nil {
			return fmt.Errorf("invalid node %d: %w", i, err)
		}
	}

	return nil
}

// selectSuitableNodes filters and ranks nodes for pipeline parallelism
func (pps *PipelinePartitionStrategy) selectSuitableNodes(nodes []*NodeInfo, analysis *ModelAnalysis, options *PartitionOptions) ([]*NodeInfo, error) {
	var suitableNodes []*NodeInfo

	// For pipeline parallelism, we need nodes with:
	// 1. Sufficient memory for individual stages
	// 2. Reasonable network connectivity (not as critical as tensor parallelism)
	// 3. Similar compute capabilities for balanced pipeline stages

	// Add defensive nil check for StageInfo to avoid rare nil deref
	var stageMemoryEstimate int64
	if analysis.StageInfo != nil {
		stageMemoryEstimate = analysis.MemoryReqs.TotalRequired / int64(analysis.StageInfo.OptimalStages)
	} else {
		// Fallback: recompute or use default estimate if StageInfo is nil
		stageMemoryEstimate = analysis.MemoryReqs.TotalRequired / int64(pps.config.MinStages)
	}

	for _, node := range nodes {
		if node.Status != NodeStatusActive {
			continue
		}

		// Check memory requirements for at least one stage
		if node.Capabilities.Memory.AvailableBytes < stageMemoryEstimate {
			continue
		}

		// Check current load
		if node.CurrentLoad != nil {
			if node.CurrentLoad.MemoryUtilization > 0.8 || 
			   node.CurrentLoad.CPUUtilization > 0.85 {
				continue // Node too loaded for pipeline stage
			}
		}

		suitableNodes = append(suitableNodes, node)
	}

	if len(suitableNodes) < pps.config.MinStages {
		return nil, fmt.Errorf("insufficient suitable nodes: need at least %d, found %d", pps.config.MinStages, len(suitableNodes))
	}

	// Rank nodes for pipeline suitability
	pps.rankNodesForPipeline(suitableNodes, analysis)

	// Select optimal number of nodes (pipeline stages)
	optimalStages := pps.calculateOptimalStageCount(analysis, len(suitableNodes), options)
	
	if len(suitableNodes) > optimalStages {
		suitableNodes = suitableNodes[:optimalStages]
	}

	return suitableNodes, nil
}

// rankNodesForPipeline ranks nodes based on their suitability for pipeline parallelism
func (pps *PipelinePartitionStrategy) rankNodesForPipeline(nodes []*NodeInfo, analysis *ModelAnalysis) {
	sort.Slice(nodes, func(i, j int) bool {
		scoreI := pps.calculatePipelineScore(nodes[i], analysis)
		scoreJ := pps.calculatePipelineScore(nodes[j], analysis)
		return scoreI > scoreJ
	})
}

// calculatePipelineScore calculates suitability score for pipeline parallelism
func (pps *PipelinePartitionStrategy) calculatePipelineScore(node *NodeInfo, analysis *ModelAnalysis) float64 {
	score := 0.0

	// Memory capacity (40% weight)
	if node.Capabilities.Memory != nil {
		var stageMemoryEstimate int64
		if analysis.StageInfo != nil {
			stageMemoryEstimate = analysis.MemoryReqs.TotalRequired / int64(analysis.StageInfo.OptimalStages)
		} else {
			// Fallback estimate if StageInfo is nil
			stageMemoryEstimate = analysis.MemoryReqs.TotalRequired / int64(2) // Default to 2 stages
		}
		memoryRatio := float64(node.Capabilities.Memory.AvailableBytes) / float64(stageMemoryEstimate)
		memoryScore := math.Min(memoryRatio/2.0, 1.0) // Normalize to 2x required
		score += memoryScore * 0.4
	}

	// CPU capability (30% weight)
	if node.Capabilities.CPU != nil {
		cpuScore := math.Min(float64(node.Capabilities.CPU.Cores)/8.0, 1.0) // Normalize to 8 cores
		score += cpuScore * 0.3
	}

	// Network capability (20% weight - less critical than tensor parallelism)
	if node.Capabilities.Network != nil {
		networkScore := math.Min(node.Capabilities.Network.Bandwidth/25.0, 1.0) // Normalize to 25 Gbps
		score += networkScore * 0.2
	}

	// Current load (10% weight)
	if node.CurrentLoad != nil {
		loadPenalty := (node.CurrentLoad.MemoryUtilization + node.CurrentLoad.CPUUtilization) / 2.0
		score += (1.0 - loadPenalty) * 0.1
	} else {
		score += 0.1
	}

	return score
}

// calculateOptimalStageCount calculates optimal number of pipeline stages
func (pps *PipelinePartitionStrategy) calculateOptimalStageCount(analysis *ModelAnalysis, availableNodes int, options *PartitionOptions) int {
	// Start with model analysis suggestion, with nil guard
	var optimal int
	if analysis.StageInfo != nil {
		optimal = analysis.StageInfo.OptimalStages
	} else {
		// Fallback calculation based on model size if StageInfo is nil
		switch {
		case analysis.ParameterSize >= 70: // 70B+ parameters
			optimal = 8
		case analysis.ParameterSize >= 30: // 30-70B parameters
			optimal = 6
		case analysis.ParameterSize >= 13: // 13-30B parameters
			optimal = 4
		default:
			optimal = pps.config.MinStages
		}
	}

	// Consider model size for adjustment
	switch {
	case analysis.ParameterSize >= 70: // 70B+ parameters
		optimal = int(math.Max(float64(optimal), 8))
	case analysis.ParameterSize >= 30: // 30-70B parameters
		optimal = int(math.Max(float64(optimal), 6))
	case analysis.ParameterSize >= 13: // 13-30B parameters
		optimal = int(math.Max(float64(optimal), 4))
	}

	// Apply constraints
	if optimal > pps.config.MaxStages {
		optimal = pps.config.MaxStages
	}
	if optimal < pps.config.MinStages {
		optimal = pps.config.MinStages
	}

	// Apply user options
	if options.MaxNodes > 0 && optimal > options.MaxNodes {
		optimal = options.MaxNodes
	}
	if options.MinNodes > 0 && optimal < options.MinNodes {
		optimal = options.MinNodes
	}

	// Ensure we don't exceed available nodes
	if optimal > availableNodes {
		optimal = availableNodes
	}

	return optimal
}

// createPipelineStages creates pipeline stages based on model analysis
func (pps *PipelinePartitionStrategy) createPipelineStages(analysis *ModelAnalysis, numStages int) ([]*PipelineStage, error) {
	layerInfo := analysis.LayerInfo
	stageInfo := analysis.StageInfo

	var stages []*PipelineStage

	// Use pre-calculated stage boundaries if available
	stageBoundaries := stageInfo.StageBoundaries
	if len(stageBoundaries) != numStages+1 {
		// Recalculate boundaries for the desired number of stages
		stageBoundaries = pps.recalculateStageBoundaries(layerInfo.TotalLayers, numStages, layerInfo.LayerWeights)
	}

	// Create stages
	for i := 0; i < numStages; i++ {
		start := stageBoundaries[i]
		end := stageBoundaries[i+1]

		// Calculate expected memory and compute for this stage
		expectedMemory := int64(0)
		expectedCompute := 0.0

		for j := start; j < end && j < len(layerInfo.LayerSizes); j++ {
			expectedMemory += layerInfo.LayerSizes[j]
			if j < len(layerInfo.LayerWeights) {
				expectedCompute += layerInfo.LayerWeights[j]
			}
		}

		// Create input/output activation specs
		inputActivations := pps.createActivationSpecs(start, layerInfo, "input")
		outputActivations := pps.createActivationSpecs(end-1, layerInfo, "output")

		// Determine dependencies
		var dependencies []int
		if i > 0 {
			dependencies = append(dependencies, i-1)
		}

		stage := &PipelineStage{
			StageIndex:        i,
			LayerRange:        []int{start, end - 1},
			ExpectedMemory:    expectedMemory,
			ExpectedCompute:   expectedCompute,
			InputActivations:  inputActivations,
			OutputActivations: outputActivations,
			Dependencies:      dependencies,
			IsFirst:           i == 0,
			IsLast:            i == numStages-1,
		}

		stages = append(stages, stage)
	}

	return stages, nil
}

// recalculateStageBoundaries recalculates stage boundaries for a specific stage count
func (pps *PipelinePartitionStrategy) recalculateStageBoundaries(totalLayers, numStages int, layerWeights []float64) []int {
	boundaries := make([]int, numStages+1)
	boundaries[0] = 0
	boundaries[numStages] = totalLayers

	if numStages == 1 {
		return boundaries
	}

	// Calculate total compute weight
	totalWeight := 0.0
	for _, weight := range layerWeights {
		totalWeight += weight
	}

	// Distribute layers based on compute weight balance
	currentWeight := 0.0
	targetWeightPerStage := totalWeight / float64(numStages)
	stageIndex := 1

	for i := 0; i < totalLayers && stageIndex < numStages; i++ {
		if i < len(layerWeights) {
			currentWeight += layerWeights[i]
		}

		// If we've reached the target weight for this stage, set boundary
		if currentWeight >= targetWeightPerStage*float64(stageIndex) {
			boundaries[stageIndex] = i + 1
			stageIndex++
		}
	}

	return boundaries
}

// createActivationSpecs creates activation specifications for a layer
func (pps *PipelinePartitionStrategy) createActivationSpecs(layerIndex int, layerInfo *LayerAnalysis, specType string) []string {
	var specs []string

	if layerType, ok := layerInfo.LayerTypes[layerIndex]; ok {
		switch layerType {
		case LayerTypeAttention:
			if specType == "input" {
				specs = []string{
					fmt.Sprintf("attention_input_%d", layerIndex),
					fmt.Sprintf("position_embeddings_%d", layerIndex),
				}
			} else {
				specs = []string{
					fmt.Sprintf("attention_output_%d", layerIndex),
					fmt.Sprintf("attention_weights_%d", layerIndex),
				}
			}
		case LayerTypeFFN:
			if specType == "input" {
				specs = []string{fmt.Sprintf("ffn_input_%d", layerIndex)}
			} else {
				specs = []string{fmt.Sprintf("ffn_output_%d", layerIndex)}
			}
		default:
			if specType == "input" {
				specs = []string{fmt.Sprintf("layer_input_%d", layerIndex)}
			} else {
				specs = []string{fmt.Sprintf("layer_output_%d", layerIndex)}
			}
		}
	}

	return specs
}

// optimizePipelineStages optimizes pipeline stages to reduce bubble time
func (pps *PipelinePartitionStrategy) optimizePipelineStages(stages []*PipelineStage, analysis *ModelAnalysis) {
	if len(stages) <= 1 {
		return
	}

	// Calculate current stage balance
	computeWeights := make([]float64, len(stages))
	totalCompute := 0.0

	for i, stage := range stages {
		computeWeights[i] = stage.ExpectedCompute
		totalCompute += stage.ExpectedCompute
	}

	avgCompute := totalCompute / float64(len(stages))

	// Check if rebalancing is needed
	maxImbalance := 0.0
	for _, weight := range computeWeights {
		imbalance := math.Abs(weight-avgCompute) / avgCompute
		if imbalance > maxImbalance {
			maxImbalance = imbalance
		}
	}

	if maxImbalance <= pps.config.StageBalanceThreshold {
		return // Already balanced
	}

	// Rebalance stages by adjusting boundaries
	pps.rebalancePipelineStages(stages, analysis.LayerInfo, avgCompute)

	// After rebalancing, recompute stage interfaces and dependencies
	pps.recomputeStageInterfaces(stages, analysis.LayerInfo)
}

// rebalancePipelineStages rebalances pipeline stages for better load distribution
func (pps *PipelinePartitionStrategy) rebalancePipelineStages(stages []*PipelineStage, layerInfo *LayerAnalysis, targetCompute float64) {
	// Simple greedy rebalancing algorithm
	for i := 0; i < len(stages)-1; i++ {
		currentStage := stages[i]
		nextStage := stages[i+1]

		currentLoad := currentStage.ExpectedCompute
		nextLoad := nextStage.ExpectedCompute

		// If current stage is overloaded and next is underloaded
		if currentLoad > targetCompute*1.3 && nextLoad < targetCompute*0.7 {
			// Move layers from current to next stage
			layersToMove := pps.calculateLayersToMove(currentStage, nextStage, layerInfo, targetCompute)
			if layersToMove > 0 {
				// Adjust stage boundaries to maintain contiguous, non-overlapping ranges
				currentStage.LayerRange[1] -= layersToMove
				nextStage.LayerRange[0] = currentStage.LayerRange[1] + 1

				// Recalculate expected resources
				pps.recalculateStageResources(currentStage, layerInfo)
				pps.recalculateStageResources(nextStage, layerInfo)

				// Ensure stage boundaries stay consistent for subsequent stages
				for j := i + 2; j < len(stages); j++ {
					if stages[j].LayerRange[0] <= nextStage.LayerRange[1] {
						offset := nextStage.LayerRange[1] - stages[j].LayerRange[0] + 1
						stages[j].LayerRange[0] += offset
						stages[j].LayerRange[1] += offset
					}
				}
			}
		}
	}
}

// calculateLayersToMove calculates how many layers to move between stages
func (pps *PipelinePartitionStrategy) calculateLayersToMove(fromStage, toStage *PipelineStage, layerInfo *LayerAnalysis, targetCompute float64) int {
	layersToMove := 0
	currentFromLoad := fromStage.ExpectedCompute
	currentToLoad := toStage.ExpectedCompute

	// Calculate how many layers we need to move to balance
	excessLoad := currentFromLoad - targetCompute
	
	// Move layers one by one until balanced
	start := fromStage.LayerRange[0]
	end := fromStage.LayerRange[1]
	
	for i := end; i > start && excessLoad > 0; i-- {
		if i < len(layerInfo.LayerWeights) {
			layerWeight := layerInfo.LayerWeights[i]
			if excessLoad >= layerWeight && (currentToLoad + layerWeight) <= targetCompute*1.2 {
				layersToMove++
				excessLoad -= layerWeight
				currentFromLoad -= layerWeight
				currentToLoad += layerWeight
			}
		}
	}

	return layersToMove
}

// recalculateStageResources recalculates resources for a stage after boundary changes
func (pps *PipelinePartitionStrategy) recalculateStageResources(stage *PipelineStage, layerInfo *LayerAnalysis) {
	start := stage.LayerRange[0]
	end := stage.LayerRange[1]

	expectedMemory := int64(0)
	expectedCompute := 0.0

	for i := start; i <= end && i < len(layerInfo.LayerSizes); i++ {
		expectedMemory += layerInfo.LayerSizes[i]
		if i < len(layerInfo.LayerWeights) {
			expectedCompute += layerInfo.LayerWeights[i]
		}
	}

	stage.ExpectedMemory = expectedMemory
	stage.ExpectedCompute = expectedCompute
}

// recomputeStageInterfaces recomputes activation specs and dependencies after rebalancing
func (pps *PipelinePartitionStrategy) recomputeStageInterfaces(stages []*PipelineStage, layerInfo *LayerAnalysis) {
	for i, stage := range stages {
		// Rebuild input/output activations based on final LayerRange
		if stage.LayerRange[0] >= 0 && stage.LayerRange[0] < layerInfo.TotalLayers {
			stage.InputActivations = pps.createActivationSpecs(stage.LayerRange[0], layerInfo, "input")
		}
		if stage.LayerRange[1] >= 0 && stage.LayerRange[1] < layerInfo.TotalLayers {
			stage.OutputActivations = pps.createActivationSpecs(stage.LayerRange[1], layerInfo, "output")
		}

		// Re-evaluate IsFirst/IsLast flags
		stage.IsFirst = (i == 0)
		stage.IsLast = (i == len(stages) - 1)

		// Rebuild dependencies according to neighboring stages
		stage.Dependencies = []int{}
		if i > 0 {
			stage.Dependencies = append(stage.Dependencies, i-1)
		}
	}

	// Validate non-overlapping contiguous ranges
	for i := 0; i < len(stages)-1; i++ {
		currentEnd := stages[i].LayerRange[1]
		nextStart := stages[i+1].LayerRange[0]
		if nextStart != currentEnd + 1 {
			// Fix gap or overlap
			stages[i+1].LayerRange[0] = currentEnd + 1
		}
	}
}

// createStageAssignments creates node assignments for pipeline stages
func (pps *PipelinePartitionStrategy) createStageAssignments(stages []*PipelineStage, nodes []*NodeInfo, analysis *ModelAnalysis) ([]*NodeAssignment, error) {
	var assignments []*NodeAssignment

	if len(stages) != len(nodes) {
		return nil, fmt.Errorf("stage count (%d) does not match node count (%d)", len(stages), len(nodes))
	}

	for i, stage := range stages {
		node := nodes[i]
		
		assignment, err := pps.createPipelineStageAssignment(stage, node, analysis)
		if err != nil {
			return nil, fmt.Errorf("failed to create assignment for stage %d: %w", i, err)
		}
		
		assignments = append(assignments, assignment)
	}

	return assignments, nil
}

// createPipelineStageAssignment creates a node assignment for a pipeline stage
func (pps *PipelinePartitionStrategy) createPipelineStageAssignment(stage *PipelineStage, node *NodeInfo, analysis *ModelAnalysis) (*NodeAssignment, error) {
	// Create pipeline stage assignment details
	stageAssignment := &PipelineStageAssignment{
		StageIndex:     stage.StageIndex,
		LayerRange:     stage.LayerRange,
		MemoryRequired: stage.ExpectedMemory,
		ComputeWeight:  stage.ExpectedCompute,
		IsFirstStage:   stage.IsFirst,
		IsLastStage:    stage.IsLast,
		InputSpecs:     pps.convertActivationsToTensorSpecs(stage.InputActivations),
		OutputSpecs:    pps.convertActivationsToTensorSpecs(stage.OutputActivations),
	}

	// Create work assignment
	workAssignment := &WorkAssignment{
		PipelineStage: stageAssignment,
	}

	// Calculate resource requirements
	resources := pps.calculatePipelineResourceRequirements(stage, node.Capabilities)

	// Create dependencies
	var dependencies []string
	for _, depStage := range stage.Dependencies {
		dependencies = append(dependencies, fmt.Sprintf("pipeline_stage_%d", depStage))
	}

	// Create the node assignment
	assignment := &NodeAssignment{
		NodeID:       node.ID,
		Role:         RoleWorker,
		WorkType:     WorkTypePipelineStage,
		Assignment:   workAssignment,
		Resources:    resources,
		Dependencies: dependencies,
	}

	return assignment, nil
}

// convertActivationsToTensorSpecs converts activation names to tensor specs
func (pps *PipelinePartitionStrategy) convertActivationsToTensorSpecs(activations []string) []TensorSpec {
	specs := make([]TensorSpec, len(activations))
	
	for i, activation := range activations {
		specs[i] = TensorSpec{
			Name:  activation,
			Shape: []int{1, 4096}, // Default shape - would be refined in real implementation
			Type:  "float32",
		}
	}
	
	return specs
}

// calculatePipelineResourceRequirements calculates resource requirements for a pipeline stage
func (pps *PipelinePartitionStrategy) calculatePipelineResourceRequirements(stage *PipelineStage, nodeCapabilities *NodeCapabilities) *ResourceRequirements {
	// CPU requirement based on stage compute load
	cpuCores := int(math.Ceil(stage.ExpectedCompute))
	if cpuCores > nodeCapabilities.CPU.Cores {
		cpuCores = nodeCapabilities.CPU.Cores
	}
	
	cpuReq := &CPURequirement{
		Cores:       cpuCores,
		Utilization: math.Min(stage.ExpectedCompute/float64(nodeCapabilities.CPU.Cores), 1.0),
	}

	// Memory requirement with checkpoint optimization
	memoryRequired := stage.ExpectedMemory
	if pps.config.CheckpointActivations {
		// Reduce memory requirement when using activation checkpointing
		memoryRequired = int64(float64(memoryRequired) * 0.7) // 30% reduction
	}

	memoryReq := &MemoryRequirement{
		Bytes:       memoryRequired,
		Utilization: float64(memoryRequired) / float64(nodeCapabilities.Memory.TotalBytes),
	}

	// GPU requirement
	var gpuReq *GPURequirement
	if nodeCapabilities.GPU != nil && nodeCapabilities.GPU.Count > 0 {
		gpuMemoryNeeded := memoryRequired
		if gpuMemoryNeeded > nodeCapabilities.GPU.TotalMemory {
			gpuMemoryNeeded = nodeCapabilities.GPU.TotalMemory
		}
		
		gpuReq = &GPURequirement{
			Count:       1,
			MemoryBytes: gpuMemoryNeeded,
			Utilization: math.Min(stage.ExpectedCompute, 1.0),
		}
	}

	// Network requirement (moderate for pipeline - mainly point-to-point)
	networkReq := &NetworkRequirement{
		BandwidthGbps: stage.ExpectedCompute * 2.0, // Moderate bandwidth
		LatencyMs:     2.0,                         // Acceptable latency for pipeline
	}

	// Storage requirement
	storageReq := &StorageRequirement{
		Bytes:     memoryRequired / 8, // Some storage for checkpoints
		IOPSRead:  int(stage.ExpectedCompute * 50),
		IOPSWrite: int(stage.ExpectedCompute * 25),
	}

	return &ResourceRequirements{
		CPU:     cpuReq,
		Memory:  memoryReq,
		GPU:     gpuReq,
		Network: networkReq,
		Storage: storageReq,
	}
}

// createPipelineCommunicationPlan creates communication plan for pipeline parallelism
func (pps *PipelinePartitionStrategy) createPipelineCommunicationPlan(assignments []*NodeAssignment, stages []*PipelineStage) *CommunicationPlan {
	// Ensure stages have been properly recomputed before creating communication plan
	// This ensures node connections mirror the final boundaries
	var connections []NodeConnection

	// Pipeline communication is sequential between adjacent stages
	for i := 0; i < len(assignments)-1; i++ {
		// Forward pass connection
		forwardConnection := NodeConnection{
			From:           assignments[i].NodeID,
			To:             assignments[i+1].NodeID,
			ConnectionType: ConnectionTypePipeline,
			Parameters: map[string]interface{}{
				"direction":         "forward",
				"stage_index":       i,
				"micro_batch_size":  pps.config.MicroBatchSize,
				"overlap_comm":      pps.config.OverlapCommunication,
			},
		}
		connections = append(connections, forwardConnection)

		// Backward pass connection (for gradients)
		backwardConnection := NodeConnection{
			From:           assignments[i+1].NodeID,
			To:             assignments[i].NodeID,
			ConnectionType: ConnectionTypeGradient,
			Parameters: map[string]interface{}{
				"direction":     "backward",
				"stage_index":   i + 1,
				"gradient_accumulation": pps.config.GradientAccumulation,
			},
		}
		connections = append(connections, backwardConnection)
	}

	return &CommunicationPlan{
		Topology:    TopologyPointToPoint,
		Connections: connections,
		Parameters: map[string]interface{}{
			"execution_pattern":     "pipeline",
			"micro_batch_size":      pps.config.MicroBatchSize,
			"bubble_optimization":   pps.config.BubbleOptimization,
			"overlap_communication": pps.config.OverlapCommunication,
			"checkpoint_activations": pps.config.CheckpointActivations,
		},
	}
}

// calculatePipelineResourceCost calculates resource cost for pipeline parallelism
func (pps *PipelinePartitionStrategy) calculatePipelineResourceCost(assignments []*NodeAssignment, nodes []*NodeInfo, stages []*PipelineStage) *ResourceCost {
	computeCost := 0.0
	memoryCost := 0.0
	networkCost := 0.0
	storageCost := 0.0

	// Calculate pipeline efficiency factor (accounts for bubble time)
	bubbleFactor := pps.calculatePipelineBubbleFactor(stages)

	for _, assignment := range assignments {
		if assignment.Resources == nil {
			continue
		}

		// CPU cost (adjusted for pipeline bubbles)
		if assignment.Resources.CPU != nil {
			cpuCostBase := float64(assignment.Resources.CPU.Cores) * assignment.Resources.CPU.Utilization * 0.12
			computeCost += cpuCostBase * bubbleFactor
		}

		// Memory cost
		if assignment.Resources.Memory != nil {
			memoryCost += float64(assignment.Resources.Memory.Bytes) / (1024 * 1024 * 1024) * 0.015
		}

		// GPU cost (adjusted for pipeline efficiency)
		if assignment.Resources.GPU != nil {
			gpuCostBase := float64(assignment.Resources.GPU.Count) * assignment.Resources.GPU.Utilization * 1.2
			computeCost += gpuCostBase * bubbleFactor
		}

		// Network cost (lower than tensor parallelism)
		if assignment.Resources.Network != nil {
			networkCost += assignment.Resources.Network.BandwidthGbps * 0.08
		}

		// Storage cost
		if assignment.Resources.Storage != nil {
			storageCost += float64(assignment.Resources.Storage.Bytes) / (1024 * 1024 * 1024) * 0.003
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
			"strategy":       pps.name,
			"stages":         len(stages),
			"bubble_factor":  bubbleFactor,
			"micro_batches":  pps.config.MicroBatchSize,
		},
	}
}

// calculatePipelineBubbleFactor calculates the efficiency factor accounting for pipeline bubbles
func (pps *PipelinePartitionStrategy) calculatePipelineBubbleFactor(stages []*PipelineStage) float64 {
	if len(stages) <= 1 {
		return 1.0
	}

	// Calculate stage balance
	totalCompute := 0.0
	maxCompute := 0.0
	minCompute := math.MaxFloat64

	for _, stage := range stages {
		totalCompute += stage.ExpectedCompute
		if stage.ExpectedCompute > maxCompute {
			maxCompute = stage.ExpectedCompute
		}
		if stage.ExpectedCompute < minCompute {
			minCompute = stage.ExpectedCompute
		}
	}

	avgCompute := totalCompute / float64(len(stages))
	
	// Calculate imbalance factor
	imbalanceFactor := 1.0
	if avgCompute > 0 {
		imbalanceFactor = maxCompute / avgCompute
	}

	// Pipeline bubble factor increases with imbalance and number of stages
	bubbleFactor := 1.0 + (imbalanceFactor-1.0)*0.3 + float64(len(stages))*0.05

	// Optimization reduces bubble factor
	if pps.config.BubbleOptimization {
		bubbleFactor *= 0.85 // 15% improvement with optimization
	}

	return bubbleFactor
}

// updatePerformanceMetrics updates performance metrics for the strategy
func (pps *PipelinePartitionStrategy) updatePerformanceMetrics(executionTime time.Duration, nodeCount int, analysis *ModelAnalysis) {
	if pps.metrics.Performance == nil {
		pps.metrics.Performance = &PerformanceMetrics{
			ExecutionTimeMs:   []float64{},
			MemoryUsageBytes:  []int64{},
			NetworkBandwidth:  []float64{},
			QualityScore:      0.8,
			EfficiencyScore:   0.75,
		}
	}

	execTimeMs := float64(executionTime.Nanoseconds()) / 1e6
	pps.metrics.Performance.ExecutionTimeMs = append(pps.metrics.Performance.ExecutionTimeMs, execTimeMs)
	pps.metrics.Performance.MemoryUsageBytes = append(pps.metrics.Performance.MemoryUsageBytes, analysis.MemoryReqs.TotalRequired)
	
	// Estimate network bandwidth (moderate for pipeline)
	estimatedBandwidth := float64(analysis.MemoryReqs.TotalRequired) / (1024 * 1024 * 1024) * 0.5 // Lower than tensor parallelism
	pps.metrics.Performance.NetworkBandwidth = append(pps.metrics.Performance.NetworkBandwidth, estimatedBandwidth)

	// Calculate average latency
	totalTime := 0.0
	for _, t := range pps.metrics.Performance.ExecutionTimeMs {
		totalTime += t
	}
	pps.metrics.AverageLatency = time.Duration(totalTime/float64(len(pps.metrics.Performance.ExecutionTimeMs))) * time.Millisecond

	// Update quality score based on stage balance
	stageBalance := 1.0
	if analysis.StageInfo != nil && len(analysis.StageInfo.StageWeights) > 1 {
		maxWeight := 0.0
		minWeight := math.MaxFloat64
		for _, weight := range analysis.StageInfo.StageWeights {
			if weight > maxWeight {
				maxWeight = weight
			}
			if weight < minWeight {
				minWeight = weight
			}
		}
		if minWeight > 0 {
			stageBalance = minWeight / maxWeight
		}
	}
	pps.metrics.Performance.QualityScore = 0.6 + (stageBalance * 0.4)

	// Efficiency score based on pipeline utilization
	pipelineUtilization := 1.0 / (1.0 + float64(nodeCount)*0.1) // Decreases with more stages
	pps.metrics.Performance.EfficiencyScore = pipelineUtilization

	// Keep only last 100 measurements
	if len(pps.metrics.Performance.ExecutionTimeMs) > 100 {
		pps.metrics.Performance.ExecutionTimeMs = pps.metrics.Performance.ExecutionTimeMs[len(pps.metrics.Performance.ExecutionTimeMs)-100:]
		pps.metrics.Performance.MemoryUsageBytes = pps.metrics.Performance.MemoryUsageBytes[len(pps.metrics.Performance.MemoryUsageBytes)-100:]
		pps.metrics.Performance.NetworkBandwidth = pps.metrics.Performance.NetworkBandwidth[len(pps.metrics.Performance.NetworkBandwidth)-100:]
	}
}

// Helper function to generate pipeline partition plan IDs
func generatePipelinePartitionPlanID() string {
	return fmt.Sprintf("pipeline_%d", time.Now().UnixNano())
}
package partitioning

import (
	"context"
	"fmt"
	"math"
	"time"
)

// HybridPartitionStrategy implements hybrid parallelism combining multiple strategies
type HybridPartitionStrategy struct {
	name              string
	analyzer          *ModelAnalyzer
	memoryOptimizer   *MemoryOptimizer
	metrics           *StrategyMetrics
	config            *HybridStrategyConfig
	layerStrategy     *LayerPartitionStrategy
	tensorStrategy    *TensorPartitionStrategy
	pipelineStrategy  *PipelinePartitionStrategy
}

// HybridStrategyConfig contains configuration for hybrid partitioning
type HybridStrategyConfig struct {
	PreferredCombinations []StrategyCombination `json:"preferred_combinations"` // Preferred strategy combinations
	MaxStrategies         int                   `json:"max_strategies"`         // Maximum number of strategies to combine
	BalanceThreshold      float64               `json:"balance_threshold"`      // Threshold for load balancing
	DynamicAdjustment     bool                  `json:"dynamic_adjustment"`     // Enable dynamic strategy adjustment
	ConflictResolution    string                `json:"conflict_resolution"`    // Method for resolving strategy conflicts
	OptimizationIterations int                  `json:"optimization_iterations"` // Number of optimization iterations
	AdaptiveWeighting     bool                  `json:"adaptive_weighting"`     // Use adaptive strategy weighting
}

// StrategyCombination defines a combination of partitioning strategies
type StrategyCombination struct {
	Name               string             `json:"name"`
	PrimaryStrategy    string             `json:"primary_strategy"`    // Primary partitioning strategy
	SecondaryStrategies []string          `json:"secondary_strategies"` // Additional strategies to apply
	Conditions         *CombinationConditions `json:"conditions"`      // Conditions for applying this combination
	Weight             float64            `json:"weight"`             // Weight/priority of this combination
	MinModelSize       int64              `json:"min_model_size"`     // Minimum model size for this combination
	MaxModelSize       int64              `json:"max_model_size"`     // Maximum model size for this combination
	MinNodes           int                `json:"min_nodes"`          // Minimum nodes required
	MaxNodes           int                `json:"max_nodes"`          // Maximum nodes supported
}

// CombinationConditions defines conditions for applying a strategy combination
type CombinationConditions struct {
	NetworkBandwidthMin  float64            `json:"network_bandwidth_min"`  // Minimum network bandwidth required
	MemoryPerNodeMin     int64              `json:"memory_per_node_min"`    // Minimum memory per node
	GPURequired          bool               `json:"gpu_required"`           // Whether GPUs are required
	ModelFamilies        []string           `json:"model_families"`         // Applicable model families
	OptimizationTargets  []OptimizationTarget `json:"optimization_targets"` // Applicable optimization targets
}

// HybridPartitionPlan extends PartitionPlan with hybrid-specific information
type HybridPartitionPlan struct {
	*PartitionPlan
	StrategyCombination  *StrategyCombination  `json:"strategy_combination"`
	PrimaryAssignments   []*NodeAssignment     `json:"primary_assignments"`
	SecondaryAssignments []*NodeAssignment     `json:"secondary_assignments"`
	CoordinationPlan     *CoordinationPlan     `json:"coordination_plan"`
}

// CoordinationPlan defines how different strategies coordinate
type CoordinationPlan struct {
	CoordinationType     CoordinationType      `json:"coordination_type"`
	SynchronizationPoints []SyncPoint          `json:"synchronization_points"`
	ConflictResolution   ConflictResolution    `json:"conflict_resolution"`
	ResourceSharing      *ResourceSharingPlan  `json:"resource_sharing"`
}

// CoordinationType defines how strategies are coordinated
type CoordinationType string

const (
	CoordinationNested     CoordinationType = "nested"      // One strategy nested within another
	CoordinationLayered    CoordinationType = "layered"     // Strategies applied in layers
	CoordinationHierarchical CoordinationType = "hierarchical" // Hierarchical strategy application
	CoordinationAdaptive   CoordinationType = "adaptive"    // Dynamic coordination based on conditions
)

// SyncPoint defines synchronization points between strategies
type SyncPoint struct {
	Name        string            `json:"name"`
	Type        SyncType          `json:"type"`
	Participants []string         `json:"participants"` // Strategy or node participants
	Condition   string            `json:"condition"`    // Condition for synchronization
	Timeout     time.Duration     `json:"timeout"`      // Timeout for synchronization
	Parameters  map[string]interface{} `json:"parameters"`
}

// SyncType defines types of synchronization
type SyncType string

const (
	SyncBarrier    SyncType = "barrier"    // All participants must reach sync point
	SyncMajority   SyncType = "majority"   // Majority of participants must reach sync point
	SyncAny        SyncType = "any"        // Any participant can trigger sync point
	SyncConditional SyncType = "conditional" // Sync based on condition
)

// ConflictResolution defines how to resolve conflicts between strategies
type ConflictResolution struct {
	Method      ConflictResolutionMethod `json:"method"`
	Priority    []string                 `json:"priority"`    // Strategy priority order
	Fallback    string                   `json:"fallback"`    // Fallback strategy
	Parameters  map[string]interface{}   `json:"parameters"`
}

// ConflictResolutionMethod defines methods for resolving conflicts
type ConflictResolutionMethod string

const (
	ConflictPriority  ConflictResolutionMethod = "priority"  // Use priority-based resolution
	ConflictNegotiate ConflictResolutionMethod = "negotiate" // Negotiate resource allocation
	ConflictPartition ConflictResolutionMethod = "partition" // Partition resources between strategies
	ConflictFallback  ConflictResolutionMethod = "fallback"  // Fall back to single strategy
)

// ResourceSharingPlan defines how resources are shared between strategies
type ResourceSharingPlan struct {
	SharedResources    []SharedResource     `json:"shared_resources"`
	IsolatedResources  []IsolatedResource   `json:"isolated_resources"`
	SharingPolicies    []SharingPolicy      `json:"sharing_policies"`
}

// SharedResource defines a resource shared between strategies
type SharedResource struct {
	ResourceType   string            `json:"resource_type"`
	ResourceID     string            `json:"resource_id"`
	Strategies     []string          `json:"strategies"`     // Strategies sharing this resource
	SharingRatio   map[string]float64 `json:"sharing_ratio"` // Ratio of resource allocation per strategy
	AccessPattern  string            `json:"access_pattern"` // Access pattern (exclusive, shared, etc.)
}

// IsolatedResource defines a resource isolated to a specific strategy
type IsolatedResource struct {
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	Strategy     string `json:"strategy"`
	Exclusive    bool   `json:"exclusive"`
}

// SharingPolicy defines policies for resource sharing
type SharingPolicy struct {
	PolicyName  string                 `json:"policy_name"`
	Conditions  []string               `json:"conditions"`
	Actions     []string               `json:"actions"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// NewHybridPartitionStrategy creates a new hybrid partition strategy
func NewHybridPartitionStrategy(analyzer *ModelAnalyzer, optimizer *MemoryOptimizer) *HybridPartitionStrategy {
	hps := &HybridPartitionStrategy{
		name:            "hybrid_parallelism",
		analyzer:        analyzer,
		memoryOptimizer: optimizer,
		metrics: &StrategyMetrics{
			UsageCount:   0,
			SuccessCount: 0,
			FailureCount: 0,
			Performance:  &PerformanceMetrics{},
			CustomMetrics: make(map[string]interface{}),
		},
		config: &HybridStrategyConfig{
			MaxStrategies:         3,
			BalanceThreshold:      0.2,
			DynamicAdjustment:     true,
			ConflictResolution:    "priority",
			OptimizationIterations: 3,
			AdaptiveWeighting:     true,
		},
	}

	// Initialize sub-strategies
	hps.layerStrategy = NewLayerPartitionStrategy(analyzer, optimizer)
	hps.tensorStrategy = NewTensorPartitionStrategy(analyzer, optimizer)
	hps.pipelineStrategy = NewPipelinePartitionStrategy(analyzer, optimizer)

	// Initialize default strategy combinations
	hps.initializeDefaultCombinations()

	return hps
}

// initializeDefaultCombinations initializes default strategy combinations
func (hps *HybridPartitionStrategy) initializeDefaultCombinations() {
	hps.config.PreferredCombinations = []StrategyCombination{
		{
			Name:               "pipeline_tensor",
			PrimaryStrategy:    "pipeline_parallelism",
			SecondaryStrategies: []string{"tensor_parallelism"},
			Weight:             0.9,
			MinModelSize:       30_000_000_000, // 30B+
			MinNodes:           4,
			MaxNodes:           32,
			Conditions: &CombinationConditions{
				NetworkBandwidthMin: 25.0, // 25 Gbps
				MemoryPerNodeMin:    32 * 1024 * 1024 * 1024, // 32GB
				GPURequired:         true,
				ModelFamilies:       []string{"llama", "gpt"},
				OptimizationTargets: []OptimizationTarget{OptimizeBalance, OptimizeThroughput},
			},
		},
		{
			Name:               "pipeline_layer",
			PrimaryStrategy:    "pipeline_parallelism",
			SecondaryStrategies: []string{"layer_parallelism"},
			Weight:             0.7,
			MinModelSize:       7_000_000_000, // 7B+
			MinNodes:           2,
			MaxNodes:           16,
			Conditions: &CombinationConditions{
				NetworkBandwidthMin: 10.0, // 10 Gbps
				MemoryPerNodeMin:    16 * 1024 * 1024 * 1024, // 16GB
				GPURequired:         false,
				ModelFamilies:       []string{"llama", "gpt", "bert"},
				OptimizationTargets: []OptimizationTarget{OptimizeLatency, OptimizeBalance},
			},
		},
		{
			Name:               "tensor_layer",
			PrimaryStrategy:    "tensor_parallelism",
			SecondaryStrategies: []string{"layer_parallelism"},
			Weight:             0.6,
			MinModelSize:       13_000_000_000, // 13B+
			MinNodes:           2,
			MaxNodes:           8,
			Conditions: &CombinationConditions{
				NetworkBandwidthMin: 40.0, // 40 Gbps
				MemoryPerNodeMin:    24 * 1024 * 1024 * 1024, // 24GB
				GPURequired:         true,
				ModelFamilies:       []string{"llama", "gpt"},
				OptimizationTargets: []OptimizationTarget{OptimizeLatency, OptimizeThroughput},
			},
		},
	}
}

// GetName returns the strategy name
func (hps *HybridPartitionStrategy) GetName() string {
	return hps.name
}

// GetMetrics returns strategy metrics
func (hps *HybridPartitionStrategy) GetMetrics() *StrategyMetrics {
	return hps.metrics
}

// CanHandle determines if this strategy can handle the given request
func (hps *HybridPartitionStrategy) CanHandle(req *PartitionRequest) bool {
	if req == nil || req.Model == nil {
		return false
	}

	// Hybrid parallelism is suitable for:
	// 1. Very large models that benefit from multiple parallelization strategies
	// 2. Complex deployment scenarios with varied node capabilities
	// 3. When optimization requires combining different approaches

	modelSize := req.Model.Parameters
	if modelSize < 7_000_000_000 { // Less than 7B parameters
		return false // Too small to benefit from hybrid parallelism
	}

	// Check if any combination strategy can handle this request
	for _, combination := range hps.config.PreferredCombinations {
		if hps.canApplyCombination(&combination, req) {
			return true
		}
	}

	return false
}

// canApplyCombination checks if a strategy combination can be applied
func (hps *HybridPartitionStrategy) canApplyCombination(combination *StrategyCombination, req *PartitionRequest) bool {
	// Check model size constraints
	modelSize := req.Model.Parameters
	if combination.MinModelSize > 0 && modelSize < combination.MinModelSize {
		return false
	}
	if combination.MaxModelSize > 0 && modelSize > combination.MaxModelSize {
		return false
	}

	// Check if primary strategy can handle the request
	switch combination.PrimaryStrategy {
	case "pipeline_parallelism":
		if !hps.pipelineStrategy.CanHandle(req) {
			return false
		}
	case "tensor_parallelism":
		if !hps.tensorStrategy.CanHandle(req) {
			return false
		}
	case "layer_parallelism":
		if !hps.layerStrategy.CanHandle(req) {
			return false
		}
	}

	// Check secondary strategies
	for _, secondaryStrategy := range combination.SecondaryStrategies {
		switch secondaryStrategy {
		case "pipeline_parallelism":
			if !hps.pipelineStrategy.CanHandle(req) {
				return false
			}
		case "tensor_parallelism":
			if !hps.tensorStrategy.CanHandle(req) {
				return false
			}
		case "layer_parallelism":
			if !hps.layerStrategy.CanHandle(req) {
				return false
			}
		}
	}

	return true
}

// Partition creates a partition plan using hybrid parallelism
func (hps *HybridPartitionStrategy) Partition(ctx context.Context, req *PartitionRequest, nodes []*NodeInfo) (*PartitionPlan, error) {
	startTime := time.Now()
	hps.metrics.UsageCount++

	// Validate inputs
	if err := hps.validateInputs(req, nodes); err != nil {
		hps.metrics.FailureCount++
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// Analyze model if not already done
	modelAnalysis := req.ModelAnalysis
	if modelAnalysis == nil {
		var err error
		modelAnalysis, err = hps.analyzer.AnalyzeModel(req.Model.Path, req.Model.Details)
		if err != nil {
			hps.metrics.FailureCount++
			return nil, fmt.Errorf("model analysis failed: %w", err)
		}
	}

	// Select the best strategy combination
	combination, err := hps.selectBestCombination(req, nodes, modelAnalysis)
	if err != nil {
		hps.metrics.FailureCount++
		return nil, fmt.Errorf("strategy combination selection failed: %w", err)
	}

	// Create hybrid partition plan
	hybridPlan, err := hps.createHybridPartitionPlan(ctx, req, nodes, combination, modelAnalysis)
	if err != nil {
		hps.metrics.FailureCount++
		return nil, fmt.Errorf("hybrid partition plan creation failed: %w", err)
	}

	// Optimize the hybrid plan
	if hps.config.OptimizationIterations > 0 {
		err = hps.optimizeHybridPlan(hybridPlan, modelAnalysis)
		if err != nil {
			return nil, fmt.Errorf("hybrid plan optimization failed: %w", err)
		}
	}

	// Convert to standard partition plan
	plan := hybridPlan.PartitionPlan
	plan.Metadata["strategy_combination"] = combination.Name
	plan.Metadata["primary_strategy"] = combination.PrimaryStrategy
	plan.Metadata["secondary_strategies"] = combination.SecondaryStrategies
	plan.Metadata["execution_time_ms"] = float64(time.Since(startTime).Nanoseconds()) / 1e6

	// Update metrics
	hps.metrics.SuccessCount++
	hps.metrics.LastUsed = time.Now()
	hps.updatePerformanceMetrics(time.Since(startTime), len(nodes), modelAnalysis)

	return plan, nil
}

// validateInputs validates the partition request and available nodes
func (hps *HybridPartitionStrategy) validateInputs(req *PartitionRequest, nodes []*NodeInfo) error {
	if err := ValidatePartitionRequest(req); err != nil {
		return err
	}

	if len(nodes) < 2 {
		return fmt.Errorf("hybrid parallelism requires at least 2 nodes")
	}

	// Validate node capabilities
	for i, node := range nodes {
		if err := ValidateNodeInfo(node); err != nil {
			return fmt.Errorf("invalid node %d: %w", i, err)
		}
	}

	return nil
}

// selectBestCombination selects the best strategy combination for the request
func (hps *HybridPartitionStrategy) selectBestCombination(req *PartitionRequest, nodes []*NodeInfo, analysis *ModelAnalysis) (*StrategyCombination, error) {
	var bestCombination *StrategyCombination
	bestScore := -1.0

	for i := range hps.config.PreferredCombinations {
		combination := &hps.config.PreferredCombinations[i]
		
		if !hps.canApplyCombination(combination, req) {
			continue
		}

		// Check node constraints
		if combination.MinNodes > 0 && len(nodes) < combination.MinNodes {
			continue
		}
		if combination.MaxNodes > 0 && len(nodes) > combination.MaxNodes {
			continue
		}

		// Check conditions
		if combination.Conditions != nil && !hps.checkCombinationConditions(combination.Conditions, req, nodes) {
			continue
		}

		// Calculate combination score
		score := hps.calculateCombinationScore(combination, req, nodes, analysis)
		if score > bestScore {
			bestScore = score
			bestCombination = combination
		}
	}

	if bestCombination == nil {
		return nil, fmt.Errorf("no suitable strategy combination found")
	}

	return bestCombination, nil
}

// checkCombinationConditions checks if combination conditions are met
func (hps *HybridPartitionStrategy) checkCombinationConditions(conditions *CombinationConditions, req *PartitionRequest, nodes []*NodeInfo) bool {
	// Check network bandwidth
	if conditions.NetworkBandwidthMin > 0 {
		sufficientNodes := 0
		for _, node := range nodes {
			if node.Capabilities.Network != nil && node.Capabilities.Network.Bandwidth >= conditions.NetworkBandwidthMin {
				sufficientNodes++
			}
		}
		if sufficientNodes < 2 { // Need at least 2 nodes with sufficient bandwidth
			return false
		}
	}

	// Check memory per node
	if conditions.MemoryPerNodeMin > 0 {
		for _, node := range nodes {
			if node.Capabilities.Memory.AvailableBytes < conditions.MemoryPerNodeMin {
				return false
			}
		}
	}

	// Check GPU requirement
	if conditions.GPURequired {
		gpuNodes := 0
		for _, node := range nodes {
			if node.Capabilities.GPU != nil && node.Capabilities.GPU.Count > 0 {
				gpuNodes++
			}
		}
		if gpuNodes == 0 {
			return false
		}
	}

	// Check model family
	if len(conditions.ModelFamilies) > 0 {
		modelFamily := req.Model.Family
		familyMatched := false
		for _, family := range conditions.ModelFamilies {
			if modelFamily == family {
				familyMatched = true
				break
			}
		}
		if !familyMatched {
			return false
		}
	}

	// Check optimization targets
	if len(conditions.OptimizationTargets) > 0 {
		targetMatched := false
		for _, target := range conditions.OptimizationTargets {
			if req.Options.OptimizeFor == target {
				targetMatched = true
				break
			}
		}
		if !targetMatched {
			return false
		}
	}

	return true
}

// calculateCombinationScore calculates a score for a strategy combination
func (hps *HybridPartitionStrategy) calculateCombinationScore(combination *StrategyCombination, req *PartitionRequest, nodes []*NodeInfo, analysis *ModelAnalysis) float64 {
	score := combination.Weight

	// Adjust score based on model size fit
	modelSize := req.Model.Parameters
	sizeScore := 1.0
	if combination.MinModelSize > 0 && combination.MaxModelSize > 0 {
		midSize := (combination.MinModelSize + combination.MaxModelSize) / 2
		sizeDiff := math.Abs(float64(modelSize - midSize)) / float64(midSize)
		sizeScore = math.Max(0.1, 1.0-sizeDiff)
	}
	score *= sizeScore

	// Adjust score based on node count fit
	nodeScore := 1.0
	if combination.MinNodes > 0 && combination.MaxNodes > 0 {
		if len(nodes) >= combination.MinNodes && len(nodes) <= combination.MaxNodes {
			midNodes := (combination.MinNodes + combination.MaxNodes) / 2
			nodeDiff := math.Abs(float64(len(nodes) - midNodes)) / float64(midNodes)
			nodeScore = math.Max(0.1, 1.0-nodeDiff)
		} else {
			nodeScore = 0.1 // Out of optimal range
		}
	}
	score *= nodeScore

	// Adjust score based on optimization target match
	if req.Options.OptimizeFor != "" && combination.Conditions != nil {
		targetMatched := false
		for _, target := range combination.Conditions.OptimizationTargets {
			if req.Options.OptimizeFor == target {
				targetMatched = true
				break
			}
		}
		if targetMatched {
			score *= 1.2 // Bonus for target match
		}
	}

	return score
}

// createHybridPartitionPlan creates a hybrid partition plan
func (hps *HybridPartitionStrategy) createHybridPartitionPlan(ctx context.Context, req *PartitionRequest, nodes []*NodeInfo, combination *StrategyCombination, analysis *ModelAnalysis) (*HybridPartitionPlan, error) {
	// Create primary strategy plan
	primaryPlan, err := hps.createPrimaryStrategyPlan(ctx, req, nodes, combination.PrimaryStrategy, analysis)
	if err != nil {
		return nil, fmt.Errorf("primary strategy plan creation failed: %w", err)
	}

	// Create secondary strategy plans
	secondaryPlans, err := hps.createSecondaryStrategyPlans(ctx, req, nodes, combination.SecondaryStrategies, primaryPlan, analysis)
	if err != nil {
		return nil, fmt.Errorf("secondary strategy plans creation failed: %w", err)
	}

	// Merge plans into hybrid plan
	mergedPlan, err := hps.mergePlans(primaryPlan, secondaryPlans, combination)
	if err != nil {
		return nil, fmt.Errorf("plan merging failed: %w", err)
	}

	// Create coordination plan
	coordinationPlan := hps.createCoordinationPlan(combination, primaryPlan, secondaryPlans)

	// Create hybrid partition plan
	hybridPlan := &HybridPartitionPlan{
		PartitionPlan:       mergedPlan,
		StrategyCombination: combination,
		PrimaryAssignments:  primaryPlan.Assignments,
		SecondaryAssignments: hps.extractSecondaryAssignments(secondaryPlans),
		CoordinationPlan:    coordinationPlan,
	}

	return hybridPlan, nil
}

// createPrimaryStrategyPlan creates a plan using the primary strategy
func (hps *HybridPartitionStrategy) createPrimaryStrategyPlan(ctx context.Context, req *PartitionRequest, nodes []*NodeInfo, primaryStrategy string, analysis *ModelAnalysis) (*PartitionPlan, error) {
	switch primaryStrategy {
	case "pipeline_parallelism":
		return hps.pipelineStrategy.Partition(ctx, req, nodes)
	case "tensor_parallelism":
		return hps.tensorStrategy.Partition(ctx, req, nodes)
	case "layer_parallelism":
		return hps.layerStrategy.Partition(ctx, req, nodes)
	default:
		return nil, fmt.Errorf("unsupported primary strategy: %s", primaryStrategy)
	}
}

// createSecondaryStrategyPlans creates plans for secondary strategies
func (hps *HybridPartitionStrategy) createSecondaryStrategyPlans(ctx context.Context, req *PartitionRequest, nodes []*NodeInfo, secondaryStrategies []string, primaryPlan *PartitionPlan, analysis *ModelAnalysis) ([]*PartitionPlan, error) {
	var secondaryPlans []*PartitionPlan

	for _, strategy := range secondaryStrategies {
		// Create a modified request for secondary strategy
		secondaryReq := hps.createSecondaryRequest(req, primaryPlan, strategy)
		
		// Get available nodes for secondary strategy
		availableNodes := hps.getAvailableNodesForSecondary(nodes, primaryPlan)

		var plan *PartitionPlan
		var err error

		switch strategy {
		case "pipeline_parallelism":
			plan, err = hps.pipelineStrategy.Partition(ctx, secondaryReq, availableNodes)
		case "tensor_parallelism":
			plan, err = hps.tensorStrategy.Partition(ctx, secondaryReq, availableNodes)
		case "layer_parallelism":
			plan, err = hps.layerStrategy.Partition(ctx, secondaryReq, availableNodes)
		default:
			err = fmt.Errorf("unsupported secondary strategy: %s", strategy)
		}

		if err != nil {
			return nil, fmt.Errorf("secondary strategy %s failed: %w", strategy, err)
		}

		secondaryPlans = append(secondaryPlans, plan)
	}

	return secondaryPlans, nil
}

// createSecondaryRequest creates a modified request for secondary strategies
func (hps *HybridPartitionStrategy) createSecondaryRequest(originalReq *PartitionRequest, primaryPlan *PartitionPlan, secondaryStrategy string) *PartitionRequest {
	// Create a copy of the original request
	secondaryReq := &PartitionRequest{
		TaskID:        originalReq.TaskID + "_" + secondaryStrategy,
		Model:         originalReq.Model,
		ModelAnalysis: originalReq.ModelAnalysis,
		Options:       originalReq.Options,
		Constraints:   originalReq.Constraints,
	}

	// Modify constraints based on primary plan resource usage
	if secondaryReq.Constraints == nil {
		secondaryReq.Constraints = &PartitionConstraints{}
	}

	// Adjust resource constraints to account for primary strategy usage
	// This is a simplified implementation - in practice, would need more sophisticated resource accounting
	
	return secondaryReq
}

// getAvailableNodesForSecondary gets nodes available for secondary strategies
func (hps *HybridPartitionStrategy) getAvailableNodesForSecondary(allNodes []*NodeInfo, primaryPlan *PartitionPlan) []*NodeInfo {
	// For this implementation, we allow sharing nodes between strategies
	// In practice, you might want more sophisticated resource partitioning
	return allNodes
}

// mergePlans merges primary and secondary plans into a unified hybrid plan
func (hps *HybridPartitionStrategy) mergePlans(primaryPlan *PartitionPlan, secondaryPlans []*PartitionPlan, combination *StrategyCombination) (*PartitionPlan, error) {
	// Start with the primary plan
	mergedPlan := &PartitionPlan{
		ID:            generateHybridPartitionPlanID(),
		TaskID:        primaryPlan.TaskID,
		Strategy:      hps.name,
		Assignments:   []*NodeAssignment{},
		Communication: &CommunicationPlan{},
		EstimatedCost: &ResourceCost{},
		CreatedAt:     time.Now(),
		Metadata:      make(map[string]interface{}),
	}

	// Merge assignments
	mergedPlan.Assignments = append(mergedPlan.Assignments, primaryPlan.Assignments...)
	
	// Merge secondary assignments
	for _, secondaryPlan := range secondaryPlans {
		// Adjust secondary assignments to work with primary
		adjustedAssignments := hps.adjustSecondaryAssignments(secondaryPlan.Assignments, primaryPlan.Assignments)
		mergedPlan.Assignments = append(mergedPlan.Assignments, adjustedAssignments...)
	}

	// Merge communication plans
	mergedPlan.Communication = hps.mergeCommunicationPlans(primaryPlan.Communication, secondaryPlans)

	// Merge costs
	mergedPlan.EstimatedCost = hps.mergeCosts(primaryPlan.EstimatedCost, secondaryPlans)

	// Merge metadata
	mergedPlan.Metadata["primary_plan_id"] = primaryPlan.ID
	mergedPlan.Metadata["secondary_plan_count"] = len(secondaryPlans)
	mergedPlan.Metadata["combination_name"] = combination.Name

	return mergedPlan, nil
}

// adjustSecondaryAssignments adjusts secondary assignments to work with primary assignments
func (hps *HybridPartitionStrategy) adjustSecondaryAssignments(secondaryAssignments, primaryAssignments []*NodeAssignment) []*NodeAssignment {
	var adjustedAssignments []*NodeAssignment

	for _, assignment := range secondaryAssignments {
		// Create adjusted assignment
		adjustedAssignment := &NodeAssignment{
			NodeID:       assignment.NodeID,
			Role:         RoleSecondary,
			WorkType:     assignment.WorkType,
			Assignment:   assignment.Assignment,
			Resources:    hps.adjustResourceRequirements(assignment.Resources, primaryAssignments),
			Dependencies: append(assignment.Dependencies, hps.getPrimaryDependencies(assignment.NodeID, primaryAssignments)...),
		}

		adjustedAssignments = append(adjustedAssignments, adjustedAssignment)
	}

	return adjustedAssignments
}

// adjustResourceRequirements adjusts resource requirements considering primary strategy usage
func (hps *HybridPartitionStrategy) adjustResourceRequirements(resources *ResourceRequirements, primaryAssignments []*NodeAssignment) *ResourceRequirements {
	if resources == nil {
		return nil
	}

	// Create adjusted resource requirements
	adjustedResources := &ResourceRequirements{
		CPU:     resources.CPU,
		Memory:  resources.Memory,
		GPU:     resources.GPU,
		Network: resources.Network,
		Storage: resources.Storage,
	}

	// Find primary assignment for the same node and adjust accordingly
	// This is simplified - in practice, you'd need more sophisticated resource coordination
	
	return adjustedResources
}

// getPrimaryDependencies gets dependencies from primary assignments
func (hps *HybridPartitionStrategy) getPrimaryDependencies(nodeID string, primaryAssignments []*NodeAssignment) []string {
	var dependencies []string

	for _, assignment := range primaryAssignments {
		if assignment.NodeID == nodeID {
			dependencies = append(dependencies, "primary_"+assignment.NodeID)
			break
		}
	}

	return dependencies
}

// mergeCommunicationPlans merges communication plans from multiple strategies
func (hps *HybridPartitionStrategy) mergeCommunicationPlans(primaryComm *CommunicationPlan, secondaryPlans []*PartitionPlan) *CommunicationPlan {
	mergedComm := &CommunicationPlan{
		Topology:    TopologyMesh, // Hybrid typically uses mesh topology
		Connections: []NodeConnection{},
		Parameters:  make(map[string]interface{}),
	}

	// Add primary connections
	if primaryComm != nil {
		mergedComm.Connections = append(mergedComm.Connections, primaryComm.Connections...)
		for k, v := range primaryComm.Parameters {
			mergedComm.Parameters["primary_"+k] = v
		}
	}

	// Add secondary connections
	for i, secondaryPlan := range secondaryPlans {
		if secondaryPlan.Communication != nil {
			// Adjust secondary connections to avoid conflicts
			adjustedConnections := hps.adjustSecondaryConnections(secondaryPlan.Communication.Connections, i)
			mergedComm.Connections = append(mergedComm.Connections, adjustedConnections...)
			
			for k, v := range secondaryPlan.Communication.Parameters {
				mergedComm.Parameters[fmt.Sprintf("secondary_%d_%s", i, k)] = v
			}
		}
	}

	return mergedComm
}

// adjustSecondaryConnections adjusts secondary connections to work with primary
func (hps *HybridPartitionStrategy) adjustSecondaryConnections(connections []NodeConnection, secondaryIndex int) []NodeConnection {
	var adjustedConnections []NodeConnection

	for _, conn := range connections {
		adjustedConn := conn
		adjustedConn.Parameters = make(map[string]interface{})
		for k, v := range conn.Parameters {
			adjustedConn.Parameters[k] = v
		}
		adjustedConn.Parameters["secondary_index"] = secondaryIndex
		adjustedConn.Parameters["priority"] = "secondary"
		
		adjustedConnections = append(adjustedConnections, adjustedConn)
	}

	return adjustedConnections
}

// mergeCosts merges costs from multiple strategies
func (hps *HybridPartitionStrategy) mergeCosts(primaryCost *ResourceCost, secondaryPlans []*PartitionPlan) *ResourceCost {
	mergedCost := &ResourceCost{
		ComputeCost: 0.0,
		MemoryCost:  0.0,
		NetworkCost: 0.0,
		StorageCost: 0.0,
		Details:     make(map[string]interface{}),
	}

	// Add primary costs
	if primaryCost != nil {
		mergedCost.ComputeCost += primaryCost.ComputeCost
		mergedCost.MemoryCost += primaryCost.MemoryCost
		mergedCost.NetworkCost += primaryCost.NetworkCost
		mergedCost.StorageCost += primaryCost.StorageCost
		mergedCost.Details["primary_cost"] = primaryCost.TotalCost
	}

	// Add secondary costs
	secondaryTotalCost := 0.0
	for i, secondaryPlan := range secondaryPlans {
		if secondaryPlan.EstimatedCost != nil {
			// Apply coordination overhead
			overhead := 1.2 // 20% overhead for coordination
			
			mergedCost.ComputeCost += secondaryPlan.EstimatedCost.ComputeCost * overhead
			mergedCost.MemoryCost += secondaryPlan.EstimatedCost.MemoryCost * overhead
			mergedCost.NetworkCost += secondaryPlan.EstimatedCost.NetworkCost * overhead
			mergedCost.StorageCost += secondaryPlan.EstimatedCost.StorageCost * overhead
			
			secondaryCost := secondaryPlan.EstimatedCost.TotalCost * overhead
			secondaryTotalCost += secondaryCost
			mergedCost.Details[fmt.Sprintf("secondary_%d_cost", i)] = secondaryCost
		}
	}

	mergedCost.Details["secondary_total_cost"] = secondaryTotalCost
	mergedCost.TotalCost = mergedCost.ComputeCost + mergedCost.MemoryCost + mergedCost.NetworkCost + mergedCost.StorageCost

	return mergedCost
}

// extractSecondaryAssignments extracts assignments from secondary plans
func (hps *HybridPartitionStrategy) extractSecondaryAssignments(secondaryPlans []*PartitionPlan) []*NodeAssignment {
	var assignments []*NodeAssignment

	for _, plan := range secondaryPlans {
		assignments = append(assignments, plan.Assignments...)
	}

	return assignments
}

// createCoordinationPlan creates a coordination plan for hybrid strategies
func (hps *HybridPartitionStrategy) createCoordinationPlan(combination *StrategyCombination, primaryPlan *PartitionPlan, secondaryPlans []*PartitionPlan) *CoordinationPlan {
	coordinationPlan := &CoordinationPlan{
		CoordinationType: CoordinationNested,
		SynchronizationPoints: []SyncPoint{
			{
				Name:         "initialization",
				Type:         SyncBarrier,
				Participants: []string{"primary"},
				Condition:    "primary_ready",
				Timeout:      30 * time.Second,
			},
		},
		ConflictResolution: ConflictResolution{
			Method:   ConflictPriority,
			Priority: []string{combination.PrimaryStrategy},
			Fallback: combination.PrimaryStrategy,
		},
		ResourceSharing: &ResourceSharingPlan{
			SharedResources:   []SharedResource{},
			IsolatedResources: []IsolatedResource{},
			SharingPolicies:   []SharingPolicy{},
		},
	}

	// Add secondary strategies to priority and sync points
	for _, strategy := range combination.SecondaryStrategies {
		coordinationPlan.ConflictResolution.Priority = append(coordinationPlan.ConflictResolution.Priority, strategy)
		
		syncPoint := SyncPoint{
			Name:         strategy + "_sync",
			Type:         SyncBarrier,
			Participants: []string{"primary", strategy},
			Condition:    strategy + "_ready",
			Timeout:      15 * time.Second,
		}
		coordinationPlan.SynchronizationPoints = append(coordinationPlan.SynchronizationPoints, syncPoint)
	}

	return coordinationPlan
}

// optimizeHybridPlan optimizes the hybrid partition plan
func (hps *HybridPartitionStrategy) optimizeHybridPlan(hybridPlan *HybridPartitionPlan, analysis *ModelAnalysis) error {
	for i := 0; i < hps.config.OptimizationIterations; i++ {
		// Optimize resource allocation
		err := hps.optimizeResourceAllocation(hybridPlan)
		if err != nil {
			return fmt.Errorf("resource allocation optimization failed: %w", err)
		}

		// Optimize communication patterns
		err = hps.optimizeCommunicationPatterns(hybridPlan)
		if err != nil {
			return fmt.Errorf("communication pattern optimization failed: %w", err)
		}

		// Balance load across strategies
		err = hps.balanceStrategicLoad(hybridPlan, analysis)
		if err != nil {
			return fmt.Errorf("strategic load balancing failed: %w", err)
		}
	}

	return nil
}

// optimizeResourceAllocation optimizes resource allocation across strategies
func (hps *HybridPartitionStrategy) optimizeResourceAllocation(hybridPlan *HybridPartitionPlan) error {
	// Implement resource allocation optimization
	// This is a simplified implementation
	return nil
}

// optimizeCommunicationPatterns optimizes communication patterns
func (hps *HybridPartitionStrategy) optimizeCommunicationPatterns(hybridPlan *HybridPartitionPlan) error {
	// Implement communication pattern optimization
	// This is a simplified implementation
	return nil
}

// balanceStrategicLoad balances load across different strategies
func (hps *HybridPartitionStrategy) balanceStrategicLoad(hybridPlan *HybridPartitionPlan, analysis *ModelAnalysis) error {
	// Implement strategic load balancing
	// This is a simplified implementation
	return nil
}

// updatePerformanceMetrics updates performance metrics for hybrid strategy
func (hps *HybridPartitionStrategy) updatePerformanceMetrics(executionTime time.Duration, nodeCount int, analysis *ModelAnalysis) {
	if hps.metrics.Performance == nil {
		hps.metrics.Performance = &PerformanceMetrics{
			ExecutionTimeMs:   []float64{},
			MemoryUsageBytes:  []int64{},
			NetworkBandwidth:  []float64{},
			QualityScore:      0.85,
			EfficiencyScore:   0.8,
		}
	}

	execTimeMs := float64(executionTime.Nanoseconds()) / 1e6
	hps.metrics.Performance.ExecutionTimeMs = append(hps.metrics.Performance.ExecutionTimeMs, execTimeMs)
	hps.metrics.Performance.MemoryUsageBytes = append(hps.metrics.Performance.MemoryUsageBytes, analysis.MemoryReqs.TotalRequired)
	
	// Higher network usage for hybrid strategies
	estimatedBandwidth := float64(analysis.TensorInfo.TotalTensorSize) / (1024 * 1024 * 1024) * 1.5
	hps.metrics.Performance.NetworkBandwidth = append(hps.metrics.Performance.NetworkBandwidth, estimatedBandwidth)

	// Calculate average latency
	totalTime := 0.0
	for _, t := range hps.metrics.Performance.ExecutionTimeMs {
		totalTime += t
	}
	hps.metrics.AverageLatency = time.Duration(totalTime/float64(len(hps.metrics.Performance.ExecutionTimeMs))) * time.Millisecond

	// Quality score based on strategy combination effectiveness
	hps.metrics.Performance.QualityScore = 0.8 + (0.2 * math.Min(1.0, float64(nodeCount)/8.0))

	// Efficiency score accounting for coordination overhead
	baseEfficiency := 0.9
	coordinationOverhead := 0.1 * float64(len(hps.config.PreferredCombinations)) / 3.0
	hps.metrics.Performance.EfficiencyScore = math.Max(0.5, baseEfficiency-coordinationOverhead)

	// Keep only last 100 measurements
	if len(hps.metrics.Performance.ExecutionTimeMs) > 100 {
		hps.metrics.Performance.ExecutionTimeMs = hps.metrics.Performance.ExecutionTimeMs[len(hps.metrics.Performance.ExecutionTimeMs)-100:]
		hps.metrics.Performance.MemoryUsageBytes = hps.metrics.Performance.MemoryUsageBytes[len(hps.metrics.Performance.MemoryUsageBytes)-100:]
		hps.metrics.Performance.NetworkBandwidth = hps.metrics.Performance.NetworkBandwidth[len(hps.metrics.Performance.NetworkBandwidth)-100:]
	}
}

// Helper function to generate hybrid partition plan IDs
func generateHybridPartitionPlanID() string {
	return fmt.Sprintf("hybrid_%d", time.Now().UnixNano())
}
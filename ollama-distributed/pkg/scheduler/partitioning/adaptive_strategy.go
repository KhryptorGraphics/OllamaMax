package partitioning

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"
)

// AdaptivePartitionStrategy implements adaptive partitioning with dynamic strategy selection
type AdaptivePartitionStrategy struct {
	name              string
	analyzer          *ModelAnalyzer
	memoryOptimizer   *MemoryOptimizer
	metrics           *StrategyMetrics
	config            *AdaptiveStrategyConfig
	strategies        map[string]PartitionStrategy
	decisionEngine    *DecisionEngine
	learningEngine    *LearningEngine
}

// AdaptiveStrategyConfig contains configuration for adaptive partitioning
type AdaptiveStrategyConfig struct {
	DecisionCriteria        []DecisionCriterion `json:"decision_criteria"`        // Criteria for strategy selection
	LearningEnabled         bool                `json:"learning_enabled"`         // Enable learning from past decisions
	RealTimeAdaptation      bool                `json:"real_time_adaptation"`     // Enable real-time strategy switching
	FallbackStrategy        string              `json:"fallback_strategy"`        // Fallback strategy if selection fails
	ConfidenceThreshold     float64             `json:"confidence_threshold"`     // Minimum confidence for strategy selection
	ExplorationRate         float64             `json:"exploration_rate"`         // Rate of exploration vs exploitation
	AdaptationInterval      time.Duration       `json:"adaptation_interval"`      // Interval for checking adaptation
	PerformanceWindow       int                 `json:"performance_window"`       // Window size for performance history
	MinSampleSize           int                 `json:"min_sample_size"`          // Minimum samples before learning
}

// DecisionCriterion defines a criterion for strategy decision making
type DecisionCriterion struct {
	Name        string                 `json:"name"`
	Weight      float64                `json:"weight"`
	Type        CriterionType          `json:"type"`
	Conditions  []ConditionRule        `json:"conditions"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// CriterionType defines types of decision criteria
type CriterionType string

const (
	CriterionModelSize      CriterionType = "model_size"
	CriterionNodeCount      CriterionType = "node_count"
	CriterionNetworkSpeed   CriterionType = "network_speed"
	CriterionMemoryRatio    CriterionType = "memory_ratio"
	CriterionGPUAvailable   CriterionType = "gpu_available"
	CriterionOptimizationTarget CriterionType = "optimization_target"
	CriterionHistoricalPerformance CriterionType = "historical_performance"
	CriterionResourceUtilization CriterionType = "resource_utilization"
)

// ConditionRule defines a condition rule for decision making
type ConditionRule struct {
	Field     string      `json:"field"`
	Operator  string      `json:"operator"`  // >, <, ==, >=, <=, !=
	Value     interface{} `json:"value"`
	Strategy  string      `json:"strategy"`  // Recommended strategy if condition is met
	Score     float64     `json:"score"`     // Score boost if condition is met
}

// DecisionEngine handles strategy selection decisions
type DecisionEngine struct {
	criteria         []DecisionCriterion
	historicalData   []*DecisionRecord
	strategyRankings map[string]float64
	contextWeights   map[string]float64
}

// LearningEngine handles learning from past decisions and performance
type LearningEngine struct {
	enabled          bool
	performanceHistory []*PerformanceRecord
	strategySuccessRates map[string]float64
	adaptationRules    []*AdaptationRule
	explorationRate    float64
}

// DecisionRecord records a partitioning decision and its outcome
type DecisionRecord struct {
	Timestamp        time.Time              `json:"timestamp"`
	ModelInfo        *ModelInfo             `json:"model_info"`
	NodeCount        int                    `json:"node_count"`
	SelectedStrategy string                 `json:"selected_strategy"`
	AlternativeStrategies []string          `json:"alternative_strategies"`
	DecisionFactors  map[string]float64     `json:"decision_factors"`
	Confidence       float64                `json:"confidence"`
	Context          *DecisionContext       `json:"context"`
	Outcome          *DecisionOutcome       `json:"outcome,omitempty"`
}

// PerformanceRecord records performance outcomes for learning
type PerformanceRecord struct {
	Timestamp        time.Time              `json:"timestamp"`
	Strategy         string                 `json:"strategy"`
	ModelSize        int64                  `json:"model_size"`
	NodeCount        int                    `json:"node_count"`
	ExecutionTime    time.Duration          `json:"execution_time"`
	ResourceUtilization float64             `json:"resource_utilization"`
	ThroughputTPS    float64                `json:"throughput_tps"`
	Success          bool                   `json:"success"`
	ErrorType        string                 `json:"error_type,omitempty"`
	Context          map[string]interface{} `json:"context"`
}

// AdaptationRule defines rules for adaptive strategy adjustment
type AdaptationRule struct {
	Name        string                 `json:"name"`
	Trigger     *AdaptationTrigger     `json:"trigger"`
	Action      *AdaptationAction      `json:"action"`
	Priority    int                    `json:"priority"`
	Enabled     bool                   `json:"enabled"`
}

// AdaptationTrigger defines triggers for adaptation
type AdaptationTrigger struct {
	Type       TriggerType            `json:"type"`
	Condition  string                 `json:"condition"`
	Threshold  float64                `json:"threshold"`
	Parameters map[string]interface{} `json:"parameters"`
}

// TriggerType defines types of adaptation triggers
type TriggerType string

const (
	TriggerPerformanceDrop TriggerType = "performance_drop"
	TriggerResourcePressure TriggerType = "resource_pressure"
	TriggerTimeInterval    TriggerType = "time_interval"
	TriggerErrorRate       TriggerType = "error_rate"
	TriggerContextChange   TriggerType = "context_change"
)

// AdaptationAction defines actions to take during adaptation
type AdaptationAction struct {
	Type       ActionType             `json:"type"`
	Target     string                 `json:"target"`
	Parameters map[string]interface{} `json:"parameters"`
}

// ActionType defines types of adaptation actions
type ActionType string

const (
	ActionSwitchStrategy    ActionType = "switch_strategy"
	ActionAdjustParameters  ActionType = "adjust_parameters"
	ActionScaleResources    ActionType = "scale_resources"
	ActionRebalanceLoad     ActionType = "rebalance_load"
)

// DecisionContext contains context information for decision making
type DecisionContext struct {
	Timestamp           time.Time              `json:"timestamp"`
	RequestID           string                 `json:"request_id"`
	ClusterState        *ClusterState          `json:"cluster_state"`
	ResourcePressure    float64                `json:"resource_pressure"`
	NetworkConditions   *NetworkConditions     `json:"network_conditions"`
	RecentPerformance   []*PerformanceRecord   `json:"recent_performance"`
	SystemLoad          float64                `json:"system_load"`
}

// DecisionOutcome contains the outcome of a partitioning decision
type DecisionOutcome struct {
	Success          bool                   `json:"success"`
	ExecutionTime    time.Duration          `json:"execution_time"`
	ResourceUsage    *ResourceUsageMetrics  `json:"resource_usage"`
	Performance      *PerformanceMetrics    `json:"performance"`
	ErrorMessage     string                 `json:"error_message,omitempty"`
	QualityScore     float64                `json:"quality_score"`
}

// ClusterState represents the current state of the cluster
type ClusterState struct {
	TotalNodes       int                    `json:"total_nodes"`
	ActiveNodes      int                    `json:"active_nodes"`
	AverageLoad      float64                `json:"average_load"`
	MemoryUtilization float64               `json:"memory_utilization"`
	NetworkUtilization float64              `json:"network_utilization"`
	NodeCapabilities map[string]*NodeCapabilities `json:"node_capabilities"`
}

// NetworkConditions represents current network conditions
type NetworkConditions struct {
	AverageBandwidth float64 `json:"average_bandwidth"`
	AverageLatency   float64 `json:"average_latency"`
	PacketLoss       float64 `json:"packet_loss"`
	Stability        float64 `json:"stability"` // 0.0-1.0
}

// ResourceUsageMetrics contains resource usage metrics
type ResourceUsageMetrics struct {
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  float64 `json:"memory_usage"`
	GPUUsage     float64 `json:"gpu_usage"`
	NetworkUsage float64 `json:"network_usage"`
	StorageUsage float64 `json:"storage_usage"`
}

// NewAdaptivePartitionStrategy creates a new adaptive partition strategy
func NewAdaptivePartitionStrategy(analyzer *ModelAnalyzer, optimizer *MemoryOptimizer) *AdaptivePartitionStrategy {
	aps := &AdaptivePartitionStrategy{
		name:            "adaptive_parallelism",
		analyzer:        analyzer,
		memoryOptimizer: optimizer,
		metrics: &StrategyMetrics{
			UsageCount:   0,
			SuccessCount: 0,
			FailureCount: 0,
			Performance:  &PerformanceMetrics{},
			CustomMetrics: make(map[string]interface{}),
		},
		strategies: make(map[string]PartitionStrategy),
		config: &AdaptiveStrategyConfig{
			LearningEnabled:      true,
			RealTimeAdaptation:   false,
			FallbackStrategy:     "layer_parallelism",
			ConfidenceThreshold:  0.7,
			ExplorationRate:      0.1,
			AdaptationInterval:   5 * time.Minute,
			PerformanceWindow:    100,
			MinSampleSize:        10,
		},
	}

	// Initialize strategies
	aps.initializeStrategies()

	// Initialize decision engine
	aps.initializeDecisionEngine()

	// Initialize learning engine
	aps.initializeLearningEngine()

	return aps
}

// initializeStrategies initializes the available strategies
func (aps *AdaptivePartitionStrategy) initializeStrategies() {
	aps.strategies["layer_parallelism"] = NewLayerPartitionStrategy(aps.analyzer, aps.memoryOptimizer)
	aps.strategies["tensor_parallelism"] = NewTensorPartitionStrategy(aps.analyzer, aps.memoryOptimizer)
	aps.strategies["pipeline_parallelism"] = NewPipelinePartitionStrategy(aps.analyzer, aps.memoryOptimizer)
	aps.strategies["hybrid_parallelism"] = NewHybridPartitionStrategy(aps.analyzer, aps.memoryOptimizer)
}

// initializeDecisionEngine initializes the decision engine with default criteria
func (aps *AdaptivePartitionStrategy) initializeDecisionEngine() {
	aps.decisionEngine = &DecisionEngine{
		historicalData:   []*DecisionRecord{},
		strategyRankings: make(map[string]float64),
		contextWeights:   make(map[string]float64),
	}

	// Initialize default decision criteria
	aps.config.DecisionCriteria = []DecisionCriterion{
		{
			Name:   "model_size_criterion",
			Weight: 0.3,
			Type:   CriterionModelSize,
			Conditions: []ConditionRule{
				{Field: "parameters", Operator: "<", Value: 3000000000, Strategy: "layer_parallelism", Score: 0.8},
				{Field: "parameters", Operator: ">=", Value: 3000000000, Strategy: "pipeline_parallelism", Score: 0.7},
				{Field: "parameters", Operator: ">=", Value: 13000000000, Strategy: "tensor_parallelism", Score: 0.9},
				{Field: "parameters", Operator: ">=", Value: 30000000000, Strategy: "hybrid_parallelism", Score: 1.0},
			},
		},
		{
			Name:   "node_count_criterion",
			Weight: 0.25,
			Type:   CriterionNodeCount,
			Conditions: []ConditionRule{
				{Field: "node_count", Operator: "<=", Value: 2, Strategy: "layer_parallelism", Score: 0.9},
				{Field: "node_count", Operator: ">", Value: 2, Strategy: "pipeline_parallelism", Score: 0.7},
				{Field: "node_count", Operator: ">=", Value: 4, Strategy: "tensor_parallelism", Score: 0.8},
				{Field: "node_count", Operator: ">=", Value: 8, Strategy: "hybrid_parallelism", Score: 1.0},
			},
		},
		{
			Name:   "network_speed_criterion",
			Weight: 0.2,
			Type:   CriterionNetworkSpeed,
			Conditions: []ConditionRule{
				{Field: "avg_bandwidth", Operator: "<", Value: 10.0, Strategy: "layer_parallelism", Score: 0.8},
				{Field: "avg_bandwidth", Operator: ">=", Value: 10.0, Strategy: "pipeline_parallelism", Score: 0.7},
				{Field: "avg_bandwidth", Operator: ">=", Value: 25.0, Strategy: "tensor_parallelism", Score: 0.9},
				{Field: "avg_bandwidth", Operator: ">=", Value: 40.0, Strategy: "hybrid_parallelism", Score: 1.0},
			},
		},
		{
			Name:   "optimization_target_criterion",
			Weight: 0.15,
			Type:   CriterionOptimizationTarget,
			Conditions: []ConditionRule{
				{Field: "optimize_for", Operator: "==", Value: "latency", Strategy: "tensor_parallelism", Score: 0.9},
				{Field: "optimize_for", Operator: "==", Value: "throughput", Strategy: "pipeline_parallelism", Score: 0.8},
				{Field: "optimize_for", Operator: "==", Value: "memory", Strategy: "layer_parallelism", Score: 0.8},
				{Field: "optimize_for", Operator: "==", Value: "balance", Strategy: "hybrid_parallelism", Score: 1.0},
			},
		},
		{
			Name:   "historical_performance_criterion",
			Weight: 0.1,
			Type:   CriterionHistoricalPerformance,
		},
	}

	// Initialize strategy rankings
	aps.decisionEngine.strategyRankings = map[string]float64{
		"layer_parallelism":    0.7,
		"tensor_parallelism":   0.8,
		"pipeline_parallelism": 0.75,
		"hybrid_parallelism":   0.9,
	}
}

// initializeLearningEngine initializes the learning engine
func (aps *AdaptivePartitionStrategy) initializeLearningEngine() {
	aps.learningEngine = &LearningEngine{
		enabled:             aps.config.LearningEnabled,
		performanceHistory:  []*PerformanceRecord{},
		strategySuccessRates: make(map[string]float64),
		adaptationRules:     []*AdaptationRule{},
		explorationRate:     aps.config.ExplorationRate,
	}

	// Initialize success rates
	for strategyName := range aps.strategies {
		aps.learningEngine.strategySuccessRates[strategyName] = 0.5 // Start with neutral
	}

	// Initialize adaptation rules
	aps.initializeAdaptationRules()
}

// initializeAdaptationRules initializes default adaptation rules
func (aps *AdaptivePartitionStrategy) initializeAdaptationRules() {
	aps.learningEngine.adaptationRules = []*AdaptationRule{
		{
			Name:     "performance_drop_rule",
			Priority: 1,
			Enabled:  true,
			Trigger: &AdaptationTrigger{
				Type:      TriggerPerformanceDrop,
				Condition: "average_performance < threshold",
				Threshold: 0.7,
			},
			Action: &AdaptationAction{
				Type:   ActionSwitchStrategy,
				Target: "best_performing_alternative",
			},
		},
		{
			Name:     "resource_pressure_rule",
			Priority: 2,
			Enabled:  true,
			Trigger: &AdaptationTrigger{
				Type:      TriggerResourcePressure,
				Condition: "memory_utilization > threshold",
				Threshold: 0.9,
			},
			Action: &AdaptationAction{
				Type:   ActionSwitchStrategy,
				Target: "memory_efficient_strategy",
			},
		},
	}
}

// GetName returns the strategy name
func (aps *AdaptivePartitionStrategy) GetName() string {
	return aps.name
}

// GetMetrics returns strategy metrics
func (aps *AdaptivePartitionStrategy) GetMetrics() *StrategyMetrics {
	return aps.metrics
}

// CanHandle determines if this strategy can handle the given request
func (aps *AdaptivePartitionStrategy) CanHandle(req *PartitionRequest) bool {
	// Adaptive strategy can handle any request by selecting appropriate sub-strategies
	if req == nil || req.Model == nil {
		return false
	}

	// Check if at least one sub-strategy can handle the request
	for _, strategy := range aps.strategies {
		if strategy.CanHandle(req) {
			return true
		}
	}

	return false
}

// Partition creates a partition plan using adaptive strategy selection
func (aps *AdaptivePartitionStrategy) Partition(ctx context.Context, req *PartitionRequest, nodes []*NodeInfo) (*PartitionPlan, error) {
	startTime := time.Now()
	aps.metrics.UsageCount++

	// Validate inputs
	if err := aps.validateInputs(req, nodes); err != nil {
		aps.metrics.FailureCount++
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// Analyze model if not already done
	modelAnalysis := req.ModelAnalysis
	if modelAnalysis == nil {
		var err error
		modelAnalysis, err = aps.analyzer.AnalyzeModel(req.Model.Path, req.Model.Details)
		if err != nil {
			aps.metrics.FailureCount++
			return nil, fmt.Errorf("model analysis failed: %w", err)
		}
	}

	// Gather decision context
	decisionContext := aps.gatherDecisionContext(req, nodes, modelAnalysis)

	// Select the best strategy using the decision engine
	selectedStrategy, confidence, err := aps.selectStrategy(req, nodes, modelAnalysis, decisionContext)
	if err != nil {
		aps.metrics.FailureCount++
		return nil, fmt.Errorf("strategy selection failed: %w", err)
	}

	// Record the decision
	decisionRecord := aps.recordDecision(req, nodes, selectedStrategy, confidence, decisionContext)

	// Execute the selected strategy
	strategy, exists := aps.strategies[selectedStrategy]
	if !exists {
		aps.metrics.FailureCount++
		return nil, fmt.Errorf("selected strategy %s not found", selectedStrategy)
	}

	plan, err := strategy.Partition(ctx, req, nodes)
	if err != nil {
		// Record failure and potentially try fallback
		aps.recordDecisionOutcome(decisionRecord, false, time.Since(startTime), err)
		
		if selectedStrategy != aps.config.FallbackStrategy {
			fallbackStrategy := aps.strategies[aps.config.FallbackStrategy]
			plan, err = fallbackStrategy.Partition(ctx, req, nodes)
			if err != nil {
				aps.metrics.FailureCount++
				return nil, fmt.Errorf("fallback strategy also failed: %w", err)
			}
			selectedStrategy = aps.config.FallbackStrategy
		} else {
			aps.metrics.FailureCount++
			return nil, fmt.Errorf("selected strategy failed: %w", err)
		}
	}

	// Update plan metadata
	plan.Metadata["adaptive_strategy"] = selectedStrategy
	plan.Metadata["selection_confidence"] = confidence
	plan.Metadata["decision_factors"] = decisionRecord.DecisionFactors
	plan.Metadata["execution_time_ms"] = float64(time.Since(startTime).Nanoseconds()) / 1e6

	// Record successful outcome
	aps.recordDecisionOutcome(decisionRecord, true, time.Since(startTime), nil)

	// Learn from this decision if learning is enabled
	if aps.learningEngine.enabled {
		aps.learnFromDecision(decisionRecord, plan)
	}

	// Update metrics
	aps.metrics.SuccessCount++
	aps.metrics.LastUsed = time.Now()
	aps.updatePerformanceMetrics(time.Since(startTime), len(nodes), modelAnalysis)

	return plan, nil
}

// validateInputs validates the partition request and available nodes
func (aps *AdaptivePartitionStrategy) validateInputs(req *PartitionRequest, nodes []*NodeInfo) error {
	if err := ValidatePartitionRequest(req); err != nil {
		return err
	}

	if len(nodes) == 0 {
		return fmt.Errorf("no nodes available")
	}

	return nil
}

// gatherDecisionContext gathers context information for decision making
func (aps *AdaptivePartitionStrategy) gatherDecisionContext(req *PartitionRequest, nodes []*NodeInfo, analysis *ModelAnalysis) *DecisionContext {
	clusterState := aps.analyzeClusterState(nodes)
	networkConditions := aps.analyzeNetworkConditions(nodes)
	recentPerformance := aps.getRecentPerformance(10) // Last 10 records

	return &DecisionContext{
		Timestamp:         time.Now(),
		RequestID:         req.TaskID,
		ClusterState:      clusterState,
		ResourcePressure:  aps.calculateResourcePressure(nodes),
		NetworkConditions: networkConditions,
		RecentPerformance: recentPerformance,
		SystemLoad:        clusterState.AverageLoad,
	}
}

// analyzeClusterState analyzes the current state of the cluster
func (aps *AdaptivePartitionStrategy) analyzeClusterState(nodes []*NodeInfo) *ClusterState {
	totalNodes := len(nodes)
	activeNodes := 0
	totalLoad := 0.0
	totalMemoryUtil := 0.0
	totalNetworkUtil := 0.0
	nodeCapabilities := make(map[string]*NodeCapabilities)

	for _, node := range nodes {
		nodeCapabilities[node.ID] = node.Capabilities
		
		if node.Status == NodeStatusActive {
			activeNodes++
			if node.CurrentLoad != nil {
				totalLoad += (node.CurrentLoad.CPUUtilization + node.CurrentLoad.MemoryUtilization) / 2.0
				totalMemoryUtil += node.CurrentLoad.MemoryUtilization
				totalNetworkUtil += node.CurrentLoad.NetworkUtilization
			}
		}
	}

	avgLoad := 0.0
	avgMemoryUtil := 0.0
	avgNetworkUtil := 0.0
	
	if activeNodes > 0 {
		avgLoad = totalLoad / float64(activeNodes)
		avgMemoryUtil = totalMemoryUtil / float64(activeNodes)
		avgNetworkUtil = totalNetworkUtil / float64(activeNodes)
	}

	return &ClusterState{
		TotalNodes:         totalNodes,
		ActiveNodes:        activeNodes,
		AverageLoad:        avgLoad,
		MemoryUtilization:  avgMemoryUtil,
		NetworkUtilization: avgNetworkUtil,
		NodeCapabilities:   nodeCapabilities,
	}
}

// analyzeNetworkConditions analyzes network conditions
func (aps *AdaptivePartitionStrategy) analyzeNetworkConditions(nodes []*NodeInfo) *NetworkConditions {
	totalBandwidth := 0.0
	totalLatency := 0.0
	validNodes := 0

	for _, node := range nodes {
		if node.Capabilities.Network != nil {
			totalBandwidth += node.Capabilities.Network.Bandwidth
			totalLatency += node.Capabilities.Network.Latency
			validNodes++
		}
	}

	avgBandwidth := 0.0
	avgLatency := 0.0
	
	if validNodes > 0 {
		avgBandwidth = totalBandwidth / float64(validNodes)
		avgLatency = totalLatency / float64(validNodes)
	}

	// Simplified stability calculation
	stability := 1.0 - (avgLatency / 100.0) // Assume 100ms is very unstable
	if stability < 0 {
		stability = 0
	}

	return &NetworkConditions{
		AverageBandwidth: avgBandwidth,
		AverageLatency:   avgLatency,
		PacketLoss:       0.0, // Would need actual measurement
		Stability:        stability,
	}
}

// calculateResourcePressure calculates overall resource pressure
func (aps *AdaptivePartitionStrategy) calculateResourcePressure(nodes []*NodeInfo) float64 {
	totalPressure := 0.0
	validNodes := 0

	for _, node := range nodes {
		if node.CurrentLoad != nil {
			nodePressure := (node.CurrentLoad.CPUUtilization + 
							node.CurrentLoad.MemoryUtilization + 
							node.CurrentLoad.GPUUtilization + 
							node.CurrentLoad.NetworkUtilization) / 4.0
			totalPressure += nodePressure
			validNodes++
		}
	}

	if validNodes == 0 {
		return 0.0
	}

	return totalPressure / float64(validNodes)
}

// getRecentPerformance gets recent performance records
func (aps *AdaptivePartitionStrategy) getRecentPerformance(limit int) []*PerformanceRecord {
	if !aps.learningEngine.enabled || len(aps.learningEngine.performanceHistory) == 0 {
		return []*PerformanceRecord{}
	}

	history := aps.learningEngine.performanceHistory
	start := len(history) - limit
	if start < 0 {
		start = 0
	}

	return history[start:]
}

// selectStrategy selects the best strategy using the decision engine
func (aps *AdaptivePartitionStrategy) selectStrategy(req *PartitionRequest, nodes []*NodeInfo, analysis *ModelAnalysis, context *DecisionContext) (string, float64, error) {
	// Calculate scores for each strategy
	strategyScores := make(map[string]float64)
	
	for strategyName, strategy := range aps.strategies {
		if !strategy.CanHandle(req) {
			continue
		}
		
		score := aps.calculateStrategyScore(strategyName, req, nodes, analysis, context)
		strategyScores[strategyName] = score
	}

	if len(strategyScores) == 0 {
		return "", 0.0, fmt.Errorf("no suitable strategies found")
	}

	// Select strategy based on scores
	selectedStrategy, confidence := aps.selectBasedOnScores(strategyScores, context)
	
	if confidence < aps.config.ConfidenceThreshold {
		// Use fallback strategy if confidence is too low
		selectedStrategy = aps.config.FallbackStrategy
		confidence = aps.config.ConfidenceThreshold
	}

	return selectedStrategy, confidence, nil
}

// calculateStrategyScore calculates a score for a strategy
func (aps *AdaptivePartitionStrategy) calculateStrategyScore(strategyName string, req *PartitionRequest, nodes []*NodeInfo, analysis *ModelAnalysis, context *DecisionContext) float64 {
	totalScore := 0.0

	// Evaluate each decision criterion
	for _, criterion := range aps.config.DecisionCriteria {
		criterionScore := aps.evaluateCriterion(criterion, strategyName, req, nodes, analysis, context)
		totalScore += criterionScore * criterion.Weight
	}

	// Add historical performance bonus
	if aps.learningEngine.enabled {
		historicalBonus := aps.getHistoricalPerformanceBonus(strategyName, req, analysis)
		totalScore += historicalBonus * 0.1 // 10% weight for historical performance
	}

	// Add exploration bonus
	if aps.learningEngine.enabled {
		explorationBonus := aps.getExplorationBonus(strategyName)
		totalScore += explorationBonus
	}

	return totalScore
}

// evaluateCriterion evaluates a single decision criterion
func (aps *AdaptivePartitionStrategy) evaluateCriterion(criterion DecisionCriterion, strategyName string, req *PartitionRequest, nodes []*NodeInfo, analysis *ModelAnalysis, context *DecisionContext) float64 {
	switch criterion.Type {
	case CriterionModelSize:
		return aps.evaluateModelSizeCriterion(criterion, strategyName, req.Model.Parameters)
	case CriterionNodeCount:
		return aps.evaluateNodeCountCriterion(criterion, strategyName, len(nodes))
	case CriterionNetworkSpeed:
		return aps.evaluateNetworkSpeedCriterion(criterion, strategyName, context.NetworkConditions.AverageBandwidth)
	case CriterionOptimizationTarget:
		return aps.evaluateOptimizationTargetCriterion(criterion, strategyName, req.Options.OptimizeFor)
	case CriterionHistoricalPerformance:
		return aps.evaluateHistoricalPerformanceCriterion(criterion, strategyName, req, analysis)
	case CriterionResourceUtilization:
		return aps.evaluateResourceUtilizationCriterion(criterion, strategyName, context.ResourcePressure)
	default:
		return 0.0
	}
}

// evaluateModelSizeCriterion evaluates model size criterion
func (aps *AdaptivePartitionStrategy) evaluateModelSizeCriterion(criterion DecisionCriterion, strategyName string, modelSize int64) float64 {
	for _, condition := range criterion.Conditions {
		if condition.Strategy == strategyName && aps.evaluateCondition(condition, "parameters", modelSize) {
			return condition.Score
		}
	}
	return 0.0
}

// evaluateNodeCountCriterion evaluates node count criterion
func (aps *AdaptivePartitionStrategy) evaluateNodeCountCriterion(criterion DecisionCriterion, strategyName string, nodeCount int) float64 {
	for _, condition := range criterion.Conditions {
		if condition.Strategy == strategyName && aps.evaluateCondition(condition, "node_count", nodeCount) {
			return condition.Score
		}
	}
	return 0.0
}

// evaluateNetworkSpeedCriterion evaluates network speed criterion
func (aps *AdaptivePartitionStrategy) evaluateNetworkSpeedCriterion(criterion DecisionCriterion, strategyName string, avgBandwidth float64) float64 {
	for _, condition := range criterion.Conditions {
		if condition.Strategy == strategyName && aps.evaluateCondition(condition, "avg_bandwidth", avgBandwidth) {
			return condition.Score
		}
	}
	return 0.0
}

// evaluateOptimizationTargetCriterion evaluates optimization target criterion
func (aps *AdaptivePartitionStrategy) evaluateOptimizationTargetCriterion(criterion DecisionCriterion, strategyName string, target OptimizationTarget) float64 {
	for _, condition := range criterion.Conditions {
		if condition.Strategy == strategyName && aps.evaluateCondition(condition, "optimize_for", target) {
			return condition.Score
		}
	}
	return 0.0
}

// evaluateHistoricalPerformanceCriterion evaluates historical performance criterion
func (aps *AdaptivePartitionStrategy) evaluateHistoricalPerformanceCriterion(criterion DecisionCriterion, strategyName string, req *PartitionRequest, analysis *ModelAnalysis) float64 {
	if !aps.learningEngine.enabled {
		return 0.0
	}

	successRate, exists := aps.learningEngine.strategySuccessRates[strategyName]
	if !exists {
		return 0.0
	}

	return successRate * 0.5 // Scale to reasonable range
}

// evaluateResourceUtilizationCriterion evaluates resource utilization criterion
func (aps *AdaptivePartitionStrategy) evaluateResourceUtilizationCriterion(criterion DecisionCriterion, strategyName string, resourcePressure float64) float64 {
	// Prefer strategies that work well under current resource pressure
	// This is a simplified implementation
	if resourcePressure > 0.8 {
		// High pressure - prefer memory-efficient strategies
		if strategyName == "layer_parallelism" {
			return 0.8
		}
		return 0.3
	}
	
	return 0.5 // Neutral score under normal conditions
}

// evaluateCondition evaluates a single condition rule
func (aps *AdaptivePartitionStrategy) evaluateCondition(condition ConditionRule, field string, value interface{}) bool {
	if condition.Field != field {
		return false
	}

	switch condition.Operator {
	case ">":
		return aps.compareValues(value, condition.Value) > 0
	case "<":
		return aps.compareValues(value, condition.Value) < 0
	case ">=":
		return aps.compareValues(value, condition.Value) >= 0
	case "<=":
		return aps.compareValues(value, condition.Value) <= 0
	case "==":
		return aps.compareValues(value, condition.Value) == 0
	case "!=":
		return aps.compareValues(value, condition.Value) != 0
	default:
		return false
	}
}

// compareValues compares two values
func (aps *AdaptivePartitionStrategy) compareValues(a, b interface{}) int {
	switch av := a.(type) {
	case int64:
		if bv, ok := b.(int64); ok {
			if av > bv {
				return 1
			} else if av < bv {
				return -1
			}
			return 0
		}
		if bv, ok := b.(int); ok {
			bv64 := int64(bv)
			if av > bv64 {
				return 1
			} else if av < bv64 {
				return -1
			}
			return 0
		}
	case int:
		if bv, ok := b.(int); ok {
			if av > bv {
				return 1
			} else if av < bv {
				return -1
			}
			return 0
		}
	case float64:
		if bv, ok := b.(float64); ok {
			if av > bv {
				return 1
			} else if av < bv {
				return -1
			}
			return 0
		}
	case string:
		if bv, ok := b.(string); ok {
			if av > bv {
				return 1
			} else if av < bv {
				return -1
			}
			return 0
		}
	case OptimizationTarget:
		if bv, ok := b.(OptimizationTarget); ok {
			if av == bv {
				return 0
			}
			return 1
		}
		if bv, ok := b.(string); ok {
			if string(av) == bv {
				return 0
			}
			return 1
		}
	}
	return 0
}

// selectBasedOnScores selects strategy based on calculated scores
func (aps *AdaptivePartitionStrategy) selectBasedOnScores(scores map[string]float64, context *DecisionContext) (string, float64) {
	// Sort strategies by score
	type strategyScore struct {
		name  string
		score float64
	}
	
	var sortedStrategies []strategyScore
	for name, score := range scores {
		sortedStrategies = append(sortedStrategies, strategyScore{name, score})
	}
	
	sort.Slice(sortedStrategies, func(i, j int) bool {
		return sortedStrategies[i].score > sortedStrategies[j].score
	})

	if len(sortedStrategies) == 0 {
		return aps.config.FallbackStrategy, aps.config.ConfidenceThreshold
	}

	best := sortedStrategies[0]
	
	// Calculate confidence based on score difference
	confidence := best.score
	if len(sortedStrategies) > 1 {
		second := sortedStrategies[1]
		scoreDifference := best.score - second.score
		confidence = math.Min(1.0, best.score + scoreDifference*0.1)
	}

	return best.name, confidence
}

// getHistoricalPerformanceBonus gets performance bonus based on historical data
func (aps *AdaptivePartitionStrategy) getHistoricalPerformanceBonus(strategyName string, req *PartitionRequest, analysis *ModelAnalysis) float64 {
	if len(aps.learningEngine.performanceHistory) < aps.config.MinSampleSize {
		return 0.0
	}

	// Find similar past requests and calculate average performance
	similarRequests := aps.findSimilarRequests(req, analysis)
	if len(similarRequests) == 0 {
		return 0.0
	}

	totalPerformance := 0.0
	strategyCount := 0
	
	for _, record := range similarRequests {
		if record.Strategy == strategyName && record.Success {
			performance := aps.calculatePerformanceScore(record)
			totalPerformance += performance
			strategyCount++
		}
	}

	if strategyCount == 0 {
		return 0.0
	}

	avgPerformance := totalPerformance / float64(strategyCount)
	return avgPerformance - 0.5 // Normalize around 0
}

// findSimilarRequests finds similar requests in performance history
func (aps *AdaptivePartitionStrategy) findSimilarRequests(req *PartitionRequest, analysis *ModelAnalysis) []*PerformanceRecord {
	var similar []*PerformanceRecord
	
	targetSize := req.Model.Parameters
	sizeThreshold := int64(float64(targetSize) * 0.3) // 30% size tolerance
	
	for _, record := range aps.learningEngine.performanceHistory {
		sizeDiff := math.Abs(float64(record.ModelSize - targetSize))
		if sizeDiff <= float64(sizeThreshold) {
			similar = append(similar, record)
		}
	}
	
	return similar
}

// calculatePerformanceScore calculates a normalized performance score
func (aps *AdaptivePartitionStrategy) calculatePerformanceScore(record *PerformanceRecord) float64 {
	// Combine multiple performance factors
	timeScore := math.Max(0, 1.0 - float64(record.ExecutionTime.Milliseconds())/10000.0) // Normalize to 10s
	utilizationScore := record.ResourceUtilization
	throughputScore := math.Min(1.0, record.ThroughputTPS/100.0) // Normalize to 100 TPS
	
	return (timeScore + utilizationScore + throughputScore) / 3.0
}

// getExplorationBonus gets exploration bonus for less-used strategies
func (aps *AdaptivePartitionStrategy) getExplorationBonus(strategyName string) float64 {
	if !aps.learningEngine.enabled {
		return 0.0
	}

	_ = aps.learningEngine.strategySuccessRates[strategyName] // Will be used for more sophisticated scoring
	usageCount := 0.0
	
	// Count recent usage
	for _, record := range aps.learningEngine.performanceHistory {
		if record.Strategy == strategyName {
			usageCount++
		}
	}
	
	// Bonus for less-used strategies (exploration)
	if usageCount == 0 {
		return aps.learningEngine.explorationRate * 0.5
	}
	
	usageRatio := usageCount / float64(len(aps.learningEngine.performanceHistory))
	explorationBonus := (1.0 - usageRatio) * aps.learningEngine.explorationRate * 0.1
	
	return explorationBonus
}

// recordDecision records a partitioning decision
func (aps *AdaptivePartitionStrategy) recordDecision(req *PartitionRequest, nodes []*NodeInfo, selectedStrategy string, confidence float64, context *DecisionContext) *DecisionRecord {
	alternatives := []string{}
	decisionFactors := make(map[string]float64)
	
	// Collect alternative strategies
	for strategyName, strategy := range aps.strategies {
		if strategyName != selectedStrategy && strategy.CanHandle(req) {
			alternatives = append(alternatives, strategyName)
		}
	}
	
	// Record decision factors (simplified)
	decisionFactors["model_size"] = float64(req.Model.Parameters)
	decisionFactors["node_count"] = float64(len(nodes))
	decisionFactors["network_bandwidth"] = context.NetworkConditions.AverageBandwidth
	decisionFactors["resource_pressure"] = context.ResourcePressure
	
	record := &DecisionRecord{
		Timestamp:             time.Now(),
		ModelInfo:             req.Model,
		NodeCount:             len(nodes),
		SelectedStrategy:      selectedStrategy,
		AlternativeStrategies: alternatives,
		DecisionFactors:       decisionFactors,
		Confidence:            confidence,
		Context:               context,
	}
	
	aps.decisionEngine.historicalData = append(aps.decisionEngine.historicalData, record)
	
	// Keep only recent decisions
	if len(aps.decisionEngine.historicalData) > aps.config.PerformanceWindow {
		aps.decisionEngine.historicalData = aps.decisionEngine.historicalData[len(aps.decisionEngine.historicalData)-aps.config.PerformanceWindow:]
	}
	
	return record
}

// recordDecisionOutcome records the outcome of a decision
func (aps *AdaptivePartitionStrategy) recordDecisionOutcome(record *DecisionRecord, success bool, executionTime time.Duration, err error) {
	outcome := &DecisionOutcome{
		Success:       success,
		ExecutionTime: executionTime,
		QualityScore:  0.5,
	}
	
	if err != nil {
		outcome.ErrorMessage = err.Error()
	}
	
	if success {
		outcome.QualityScore = 0.8 + (0.2 * record.Confidence)
	} else {
		outcome.QualityScore = 0.2
	}
	
	record.Outcome = outcome
}

// learnFromDecision learns from a partitioning decision and its outcome
func (aps *AdaptivePartitionStrategy) learnFromDecision(record *DecisionRecord, plan *PartitionPlan) {
	if record.Outcome == nil {
		return
	}
	
	// Update strategy success rates
	strategyName := record.SelectedStrategy
	currentRate := aps.learningEngine.strategySuccessRates[strategyName]
	
	// Use exponential moving average
	alpha := 0.1
	if record.Outcome.Success {
		newRate := currentRate + alpha*(1.0-currentRate)
		aps.learningEngine.strategySuccessRates[strategyName] = newRate
	} else {
		newRate := currentRate + alpha*(0.0-currentRate)
		aps.learningEngine.strategySuccessRates[strategyName] = newRate
	}
	
	// Add to performance history
	performanceRecord := &PerformanceRecord{
		Timestamp:           record.Timestamp,
		Strategy:           strategyName,
		ModelSize:          record.ModelInfo.Parameters,
		NodeCount:          record.NodeCount,
		ExecutionTime:      record.Outcome.ExecutionTime,
		ResourceUtilization: 0.7, // Would be calculated from actual metrics
		ThroughputTPS:      10.0,  // Would be calculated from actual metrics
		Success:            record.Outcome.Success,
		Context:            make(map[string]interface{}),
	}
	
	if !record.Outcome.Success {
		performanceRecord.ErrorType = "execution_failed"
	}
	
	aps.learningEngine.performanceHistory = append(aps.learningEngine.performanceHistory, performanceRecord)
	
	// Keep only recent performance history
	if len(aps.learningEngine.performanceHistory) > aps.config.PerformanceWindow {
		aps.learningEngine.performanceHistory = aps.learningEngine.performanceHistory[len(aps.learningEngine.performanceHistory)-aps.config.PerformanceWindow:]
	}
	
	// Update decision engine rankings based on outcomes
	aps.updateStrategyRankings()
}

// updateStrategyRankings updates strategy rankings based on historical performance
func (aps *AdaptivePartitionStrategy) updateStrategyRankings() {
	for strategyName, successRate := range aps.learningEngine.strategySuccessRates {
		// Combine success rate with current ranking
		currentRanking := aps.decisionEngine.strategyRankings[strategyName]
		newRanking := 0.7*currentRanking + 0.3*successRate
		aps.decisionEngine.strategyRankings[strategyName] = newRanking
	}
}

// updatePerformanceMetrics updates performance metrics for the adaptive strategy
func (aps *AdaptivePartitionStrategy) updatePerformanceMetrics(executionTime time.Duration, nodeCount int, analysis *ModelAnalysis) {
	if aps.metrics.Performance == nil {
		aps.metrics.Performance = &PerformanceMetrics{
			ExecutionTimeMs:   []float64{},
			MemoryUsageBytes:  []int64{},
			NetworkBandwidth:  []float64{},
			QualityScore:      0.8,
			EfficiencyScore:   0.85,
		}
	}

	execTimeMs := float64(executionTime.Nanoseconds()) / 1e6
	aps.metrics.Performance.ExecutionTimeMs = append(aps.metrics.Performance.ExecutionTimeMs, execTimeMs)
	aps.metrics.Performance.MemoryUsageBytes = append(aps.metrics.Performance.MemoryUsageBytes, analysis.MemoryReqs.TotalRequired)
	
	// Network usage depends on selected strategy (estimated)
	estimatedBandwidth := float64(analysis.TensorInfo.TotalTensorSize) / (1024 * 1024 * 1024) * 0.8
	aps.metrics.Performance.NetworkBandwidth = append(aps.metrics.Performance.NetworkBandwidth, estimatedBandwidth)

	// Calculate average latency
	totalTime := 0.0
	for _, t := range aps.metrics.Performance.ExecutionTimeMs {
		totalTime += t
	}
	aps.metrics.AverageLatency = time.Duration(totalTime/float64(len(aps.metrics.Performance.ExecutionTimeMs))) * time.Millisecond

	// Quality score based on decision confidence and learning
	avgConfidence := 0.0
	if len(aps.decisionEngine.historicalData) > 0 {
		for _, record := range aps.decisionEngine.historicalData {
			avgConfidence += record.Confidence
		}
		avgConfidence /= float64(len(aps.decisionEngine.historicalData))
	}
	aps.metrics.Performance.QualityScore = 0.7 + (avgConfidence * 0.3)

	// Efficiency score based on success rate
	totalSuccessRate := 0.0
	strategyCount := 0
	for _, successRate := range aps.learningEngine.strategySuccessRates {
		totalSuccessRate += successRate
		strategyCount++
	}
	if strategyCount > 0 {
		avgSuccessRate := totalSuccessRate / float64(strategyCount)
		aps.metrics.Performance.EfficiencyScore = 0.5 + (avgSuccessRate * 0.5)
	}

	// Keep only last 100 measurements
	if len(aps.metrics.Performance.ExecutionTimeMs) > 100 {
		aps.metrics.Performance.ExecutionTimeMs = aps.metrics.Performance.ExecutionTimeMs[len(aps.metrics.Performance.ExecutionTimeMs)-100:]
		aps.metrics.Performance.MemoryUsageBytes = aps.metrics.Performance.MemoryUsageBytes[len(aps.metrics.Performance.MemoryUsageBytes)-100:]
		aps.metrics.Performance.NetworkBandwidth = aps.metrics.Performance.NetworkBandwidth[len(aps.metrics.Performance.NetworkBandwidth)-100:]
	}
}
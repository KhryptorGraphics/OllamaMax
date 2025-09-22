package partitioning

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	api_types "github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/types"
)

// EnhancedPartitionManagerImpl implements the EnhancedPartitionManager interface
type EnhancedPartitionManagerImpl struct {
	baseManager      PartitionManager
	strategies       map[string]PartitionStrategy
	strategyAliases  map[string]string // Maps aliases to canonical names
	analyzer         *ModelAnalyzer
	memoryOptimizer  *MemoryOptimizer
	strategyMetrics  map[string]*StrategyMetrics
	selectionHistory []*StrategySelectionRecord
	recommendations  []*PartitionRecommendation
	mu               sync.RWMutex
	config           *EnhancedManagerConfig
}

// EnhancedManagerConfig contains configuration for the enhanced partition manager
type EnhancedManagerConfig struct {
	MaxHistorySize      int           `json:"max_history_size"`      // Maximum size of selection history
	MetricsRetention    time.Duration `json:"metrics_retention"`     // How long to retain metrics
	RecommendationLimit int           `json:"recommendation_limit"`  // Maximum number of recommendations
	PerformanceWindow   int           `json:"performance_window"`    // Window for performance calculations
	AdaptationEnabled   bool          `json:"adaptation_enabled"`    // Enable adaptive improvements
	LearningRate        float64       `json:"learning_rate"`         // Learning rate for adaptations
	AutoOptimization    bool          `json:"auto_optimization"`     // Enable automatic optimizations
}

// BasePartitionManagerImpl implements the basic PartitionManager interface
type BasePartitionManagerImpl struct {
	strategies map[string]PartitionStrategy
	analyzer   *ModelAnalyzer
	optimizer  *MemoryOptimizer
	mu         sync.RWMutex
}

// NewEnhancedPartitionManager creates a new enhanced partition manager
func NewEnhancedPartitionManager() EnhancedPartitionManager {
	analyzer := NewModelAnalyzer()
	optimizer := NewMemoryOptimizer()
	
	// Create base manager
	baseManager := &BasePartitionManagerImpl{
		strategies: make(map[string]PartitionStrategy),
		analyzer:   analyzer,
		optimizer:  optimizer,
	}
	
	// Initialize strategies in base manager
	baseManager.initializeStrategies()
	
	enhanced := &EnhancedPartitionManagerImpl{
		baseManager:      baseManager,
		strategies:       baseManager.strategies,
		strategyAliases:  make(map[string]string),
		analyzer:         analyzer,
		memoryOptimizer:  optimizer,
		strategyMetrics:  make(map[string]*StrategyMetrics),
		selectionHistory: []*StrategySelectionRecord{},
		recommendations:  []*PartitionRecommendation{},
		config: &EnhancedManagerConfig{
			MaxHistorySize:      1000,
			MetricsRetention:    24 * time.Hour,
			RecommendationLimit: 10,
			PerformanceWindow:   50,
			AdaptationEnabled:   true,
			LearningRate:        0.1,
			AutoOptimization:    true,
		},
	}
	
	// Initialize strategy aliases for consistent naming
	enhanced.initializeStrategyAliases()
	
	// Initialize strategy metrics
	enhanced.initializeStrategyMetrics()
	
	return enhanced
}

// initializeStrategies initializes all available strategies
func (bpm *BasePartitionManagerImpl) initializeStrategies() {
	analyzer := bpm.analyzer
	optimizer := bpm.optimizer
	
	bpm.strategies["layer_parallelism"] = NewLayerPartitionStrategy(analyzer, optimizer)
	bpm.strategies["tensor_parallelism"] = NewTensorPartitionStrategy(analyzer, optimizer)
	bpm.strategies["pipeline_parallelism"] = NewPipelinePartitionStrategy(analyzer, optimizer)
	bpm.strategies["hybrid_parallelism"] = NewHybridPartitionStrategy(analyzer, optimizer)
	bpm.strategies["adaptive_parallelism"] = NewAdaptivePartitionStrategy(analyzer, optimizer)
}

// initializeStrategyAliases sets up strategy name aliases for consistency
func (epm *EnhancedPartitionManagerImpl) initializeStrategyAliases() {
	epm.mu.Lock()
	defer epm.mu.Unlock()
	
	// Map short names to canonical names
	epm.strategyAliases["layer"] = "layer_parallelism"
	epm.strategyAliases["layerwise"] = "layer_parallelism"
	epm.strategyAliases["tensor"] = "tensor_parallelism"
	epm.strategyAliases["pipeline"] = "pipeline_parallelism"
	epm.strategyAliases["hybrid"] = "hybrid_parallelism"
	epm.strategyAliases["adaptive"] = "adaptive_parallelism"
	
	// Long-form aliases for compatibility
	epm.strategyAliases["layerwise_parallelism"] = "layer_parallelism"
	epm.strategyAliases["tensor_parallel"] = "tensor_parallelism"
	epm.strategyAliases["pipeline_parallel"] = "pipeline_parallelism"
	epm.strategyAliases["hybrid_parallel"] = "hybrid_parallelism"
	epm.strategyAliases["adaptive_partitioning"] = "adaptive_parallelism"
}

// initializeStrategyMetrics initializes metrics for all strategies
func (epm *EnhancedPartitionManagerImpl) initializeStrategyMetrics() {
	epm.mu.Lock()
	defer epm.mu.Unlock()
	
	for name, strategy := range epm.strategies {
		epm.strategyMetrics[name] = strategy.GetMetrics()
	}
}

// Partition creates a partition plan for the given task
func (epm *EnhancedPartitionManagerImpl) Partition(ctx context.Context, task *api_types.DistributedTask) (*PartitionPlan, error) {
	// Convert DistributedTask to PartitionRequest
	req, err := epm.convertTaskToRequest(task)
	if err != nil {
		return nil, fmt.Errorf("failed to convert task to request: %w", err)
	}
	
	// Get available nodes
	nodes, err := epm.getAvailableNodes(task)
	if err != nil {
		return nil, fmt.Errorf("failed to get available nodes: %w", err)
	}
	
	// Select best strategy
	strategyName, err := epm.SelectStrategy(task, req.Model, req.Options)
	if err != nil {
		return nil, fmt.Errorf("strategy selection failed: %w", err)
	}
	
	// Get the selected strategy
	strategy, exists := epm.strategies[strategyName]
	if !exists {
		return nil, fmt.Errorf("selected strategy %s not found", strategyName)
	}
	
	// Execute partitioning
	startTime := time.Now()
	plan, err := strategy.Partition(ctx, req, nodes)
	executionTime := time.Since(startTime)
	
	// Record the execution result
	result := &PartitionExecutionResult{
		TaskID:        string(task.ID),
		Strategy:      strategyName,
		Success:       err == nil,
		ExecutionTime: executionTime,
		Metrics:       strategy.GetMetrics().Performance,
	}
	
	// Only set PlanID if plan is not nil
	if err == nil && plan != nil {
		result.PlanID = plan.ID
	}
	
	if err != nil {
		result.ErrorMessage = err.Error()
	}
	
	// Update metrics
	epm.UpdateMetrics(strategyName, result)
	
	// Record selection history
	epm.recordStrategySelection(task, strategyName, req.Model, nodes, err == nil)
	
	// Generate recommendations if auto-optimization is enabled
	if epm.config.AutoOptimization {
		go epm.generateRecommendationsAsync(ctx, req)
	}
	
	return plan, err
}

// SelectStrategy selects the best strategy for the given task
func (epm *EnhancedPartitionManagerImpl) SelectStrategy(task *api_types.DistributedTask, model *ModelInfo, options *PartitionOptions) (string, error) {
	epm.mu.RLock()
	defer epm.mu.RUnlock()
	
	// Normalize the strategy name if one is specified in options
	if options != nil && options.Strategy != "" {
		options.Strategy = NormalizeStrategyName(options.Strategy)
	}
	
	// Get available nodes to consider their constraints
	nodes, err := epm.getAvailableNodes(task)
	if err != nil {
		return "", fmt.Errorf("failed to get available nodes: %w", err)
	}
	
	// Get model analysis
	var analysis *ModelAnalysis
	
	if model != nil && model.Path != "" {
		analysis, err = epm.analyzer.AnalyzeModel(model.Path, model.Details)
		if err != nil {
			return "", fmt.Errorf("model analysis failed: %w", err)
		}
	}
	
	// Create partition request for evaluation
	req := &PartitionRequest{
		TaskID:        string(task.ID),
		Model:         model,
		ModelAnalysis: analysis,
		Options:       options,
	}
	
	// Score all available strategies
	strategyScores := make(map[string]float64)
	
	for strategyName, strategy := range epm.strategies {
		if !strategy.CanHandle(req) {
			continue
		}
		
		score := epm.calculateStrategyScore(strategyName, req, task, nodes)
		strategyScores[strategyName] = score
	}
	
	if len(strategyScores) == 0 {
		return "", fmt.Errorf("no suitable strategies found")
	}
	
	// Select strategy with highest score
	bestStrategy := ""
	bestScore := -1.0
	
	for strategyName, score := range strategyScores {
		if score > bestScore {
			bestScore = score
			bestStrategy = strategyName
		}
	}
	
	return bestStrategy, nil
}

// calculateStrategyScore calculates a score for a strategy
func (epm *EnhancedPartitionManagerImpl) calculateStrategyScore(strategyName string, req *PartitionRequest, task *api_types.DistributedTask, nodes []*NodeInfo) float64 {
	score := 0.0
	
	// Base score from strategy metrics
	if metrics, exists := epm.strategyMetrics[strategyName]; exists {
		successRate := float64(metrics.SuccessCount) / float64(metrics.UsageCount+1)
		score += successRate * 0.4
		
		if metrics.Performance != nil {
			score += metrics.Performance.QualityScore * 0.3
			score += metrics.Performance.EfficiencyScore * 0.3
		}
	}
	
	// Model size considerations
	if req.Model != nil && req.ModelAnalysis != nil {
		modelSize := req.Model.Parameters
		
		switch strategyName {
		case "layer_parallelism":
			if modelSize < 7_000_000_000 {
				score += 0.2 // Good for smaller models
			}
		case "tensor_parallelism":
			if modelSize >= 7_000_000_000 {
				score += 0.3 // Good for larger models
			}
		case "pipeline_parallelism":
			if modelSize >= 13_000_000_000 {
				score += 0.25 // Good for very large models
			}
		case "hybrid_parallelism":
			if modelSize >= 30_000_000_000 {
				score += 0.4 // Best for extremely large models
			}
		case "adaptive_parallelism":
			score += 0.1 // Always gets a small bonus for adaptivity
		}
	}
	
	// Node constraint scoring - penalize strategies that can't work with available nodes
	nodeCount := len(nodes)
	if nodeCount > 0 {
		switch strategyName {
		case "tensor_parallelism":
			// Tensor parallelism needs high bandwidth between nodes
			if nodeCount < 2 {
				score *= 0.1 // Heavy penalty for insufficient nodes
			}
			// Check for sufficient bandwidth
			highBandwidthNodes := 0
			for _, node := range nodes {
				if node.Capabilities != nil && node.Capabilities.Network != nil && node.Capabilities.Network.Bandwidth >= 10.0 {
					highBandwidthNodes++
				}
			}
			if highBandwidthNodes < 2 {
				score *= 0.5 // Penalty for insufficient bandwidth
			}
		case "pipeline_parallelism":
			if nodeCount < 3 {
				score *= 0.3 // Pipeline needs at least 3 nodes to be effective
			}
		case "hybrid_parallelism":
			if nodeCount < 4 {
				score *= 0.2 // Hybrid needs even more nodes
			}
		}
	}
	
	// Historical performance bonus
	score += epm.getHistoricalPerformanceBonus(strategyName, req)
	
	return score
}

// getHistoricalPerformanceBonus gets bonus score based on historical performance
func (epm *EnhancedPartitionManagerImpl) getHistoricalPerformanceBonus(strategyName string, req *PartitionRequest) float64 {
	if len(epm.selectionHistory) < 10 {
		return 0.0
	}
	
	recentSuccess := 0
	recentTotal := 0
	
	// Look at recent history
	start := len(epm.selectionHistory) - 20
	if start < 0 {
		start = 0
	}
	
	for i := start; i < len(epm.selectionHistory); i++ {
		record := epm.selectionHistory[i]
		if record.SelectedStrategy == strategyName {
			recentTotal++
			if record.Success {
				recentSuccess++
			}
		}
	}
	
	if recentTotal == 0 {
		return 0.0
	}
	
	recentSuccessRate := float64(recentSuccess) / float64(recentTotal)
	return (recentSuccessRate - 0.5) * 0.2 // Bonus/penalty based on recent performance
}

// ValidatePartition validates a partition plan
func (epm *EnhancedPartitionManagerImpl) ValidatePartition(ctx context.Context, plan *PartitionPlan, nodes []*NodeInfo) (*PartitionValidationResult, error) {
	result := &PartitionValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
		Score:    1.0,
		Details:  make(map[string]interface{}),
	}
	
	// Validate assignments
	if len(plan.Assignments) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:     "no_assignments",
			Message:  "Partition plan has no node assignments",
			Severity: SeverityCritical,
		})
	}
	
	// Validate node capacity
	nodeMap := make(map[string]*NodeInfo)
	for _, node := range nodes {
		nodeMap[node.ID] = node
	}
	
	for _, assignment := range plan.Assignments {
		node, exists := nodeMap[assignment.NodeID]
		if !exists {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Code:     "node_not_found",
				Message:  fmt.Sprintf("Node %s not found", assignment.NodeID),
				NodeID:   assignment.NodeID,
				Severity: SeverityCritical,
			})
			continue
		}
		
		// Check resource requirements
		if assignment.Resources != nil {
			if assignment.Resources.Memory != nil {
				required := assignment.Resources.Memory.Bytes
				available := node.Capabilities.Memory.AvailableBytes
				
				if required > available {
					result.Valid = false
					result.Errors = append(result.Errors, ValidationError{
						Code:     "insufficient_memory",
						Message:  fmt.Sprintf("Node %s requires %d bytes but only has %d available", node.ID, required, available),
						NodeID:   node.ID,
						Severity: SeverityHigh,
					})
				} else if float64(required)/float64(available) > 0.9 {
					result.Warnings = append(result.Warnings, ValidationWarning{
						Code:       "high_memory_usage",
						Message:    fmt.Sprintf("Node %s will use >90%% of available memory", node.ID),
						NodeID:     node.ID,
						Suggestion: "Consider reducing memory allocation or adding more nodes",
					})
				}
			}
		}
	}
	
	// Calculate quality score
	score := 1.0
	score -= float64(len(result.Errors)) * 0.3
	score -= float64(len(result.Warnings)) * 0.1
	result.Score = score
	
	result.Details["total_nodes"] = len(plan.Assignments)
	result.Details["strategy"] = plan.Strategy
	result.Details["validation_time"] = time.Now()
	
	return result, nil
}

// GetAvailableStrategies returns available partitioning strategies
func (epm *EnhancedPartitionManagerImpl) GetAvailableStrategies() []string {
	epm.mu.RLock()
	defer epm.mu.RUnlock()
	
	strategies := make([]string, 0, len(epm.strategies))
	for name := range epm.strategies {
		strategies = append(strategies, name)
	}
	
	sort.Strings(strategies)
	return strategies
}

// RegisterStrategy registers a new partitioning strategy
func (epm *EnhancedPartitionManagerImpl) RegisterStrategy(name string, strategy PartitionStrategy) error {
	epm.mu.Lock()
	defer epm.mu.Unlock()
	
	// Normalize the name through alias resolution
	canonicalName := name
	if alias, hasAlias := epm.strategyAliases[name]; hasAlias {
		canonicalName = alias
	}
	
	if _, exists := epm.strategies[canonicalName]; exists {
		return fmt.Errorf("strategy %s already exists", canonicalName)
	}
	
	epm.strategies[canonicalName] = strategy
	epm.strategyMetrics[canonicalName] = strategy.GetMetrics()
	
	return nil
}

// GetStrategy returns a strategy by name
func (epm *EnhancedPartitionManagerImpl) GetStrategy(name string) (PartitionStrategy, error) {
	epm.mu.RLock()
	defer epm.mu.RUnlock()
	
	// Normalize the name through alias resolution
	canonicalName := name
	if alias, hasAlias := epm.strategyAliases[name]; hasAlias {
		canonicalName = alias
	}
	
	strategy, exists := epm.strategies[canonicalName]
	if !exists {
		return nil, fmt.Errorf("strategy %s not found", name)
	}
	
	return strategy, nil
}

// GetStrategyMetrics returns metrics for a specific strategy
func (epm *EnhancedPartitionManagerImpl) GetStrategyMetrics(strategyName string) (*StrategyMetrics, error) {
	epm.mu.RLock()
	defer epm.mu.RUnlock()
	
	metrics, exists := epm.strategyMetrics[strategyName]
	if !exists {
		return nil, fmt.Errorf("metrics for strategy %s not found", strategyName)
	}
	
	return metrics, nil
}

// GetSelectionHistory returns the history of strategy selections
func (epm *EnhancedPartitionManagerImpl) GetSelectionHistory(limit int) ([]*StrategySelectionRecord, error) {
	epm.mu.RLock()
	defer epm.mu.RUnlock()
	
	if limit <= 0 || limit > len(epm.selectionHistory) {
		limit = len(epm.selectionHistory)
	}
	
	start := len(epm.selectionHistory) - limit
	history := make([]*StrategySelectionRecord, limit)
	copy(history, epm.selectionHistory[start:])
	
	return history, nil
}

// OptimizeStrategy optimizes strategy selection based on historical data
func (epm *EnhancedPartitionManagerImpl) OptimizeStrategy(ctx context.Context, req *PartitionRequest) (string, error) {
	// This is a simplified implementation
	// In practice, this would use machine learning or more sophisticated optimization
	
	strategyScores := make(map[string]float64)
	
	for strategyName, strategy := range epm.strategies {
		if !strategy.CanHandle(req) {
			continue
		}
		
		// Calculate score based on historical performance
		score := epm.calculateOptimizedScore(strategyName, req)
		strategyScores[strategyName] = score
	}
	
	if len(strategyScores) == 0 {
		return "", fmt.Errorf("no suitable strategies found")
	}
	
	// Select best strategy
	bestStrategy := ""
	bestScore := -1.0
	
	for strategyName, score := range strategyScores {
		if score > bestScore {
			bestScore = score
			bestStrategy = strategyName
		}
	}
	
	return bestStrategy, nil
}

// calculateOptimizedScore calculates an optimized score for a strategy
func (epm *EnhancedPartitionManagerImpl) calculateOptimizedScore(strategyName string, req *PartitionRequest) float64 {
	// Base score (pass empty nodes array when not available)
	score := epm.calculateStrategyScore(strategyName, req, nil, []*NodeInfo{})
	
	// Learning-based adjustments
	if epm.config.AdaptationEnabled {
		adjustment := epm.calculateLearningAdjustment(strategyName, req)
		score += adjustment * epm.config.LearningRate
	}
	
	return score
}

// calculateLearningAdjustment calculates learning-based score adjustments
func (epm *EnhancedPartitionManagerImpl) calculateLearningAdjustment(strategyName string, req *PartitionRequest) float64 {
	// Find similar requests in history
	similarRequests := epm.findSimilarHistoricalRequests(req)
	if len(similarRequests) == 0 {
		return 0.0
	}
	
	successCount := 0
	totalCount := 0
	
	for _, record := range similarRequests {
		if record.SelectedStrategy == strategyName {
			totalCount++
			if record.Success {
				successCount++
			}
		}
	}
	
	if totalCount == 0 {
		return 0.0
	}
	
	successRate := float64(successCount) / float64(totalCount)
	return successRate - 0.5 // Adjustment relative to 50% baseline
}

// findSimilarHistoricalRequests finds similar requests in history
func (epm *EnhancedPartitionManagerImpl) findSimilarHistoricalRequests(req *PartitionRequest) []*StrategySelectionRecord {
	var similar []*StrategySelectionRecord
	
	if req.Model == nil {
		return similar
	}
	
	targetSize := req.Model.Parameters
	tolerance := targetSize * 30 / 100 // 30% tolerance
	
	for _, record := range epm.selectionHistory {
		if record.ModelInfo != nil {
			sizeDiff := record.ModelInfo.Parameters - targetSize
			if sizeDiff < 0 {
				sizeDiff = -sizeDiff
			}
			
			if sizeDiff <= tolerance {
				similar = append(similar, record)
			}
		}
	}
	
	return similar
}

// UpdateMetrics updates metrics for a strategy based on execution results
func (epm *EnhancedPartitionManagerImpl) UpdateMetrics(strategyName string, result *PartitionExecutionResult) error {
	epm.mu.Lock()
	defer epm.mu.Unlock()
	
	metrics, exists := epm.strategyMetrics[strategyName]
	if !exists {
		return fmt.Errorf("metrics for strategy %s not found", strategyName)
	}
	
	// NOTE: Do NOT increment UsageCount, SuccessCount, or FailureCount here
	// as the strategies already update these counts themselves
	
	// Update latency (moving average)
	if metrics.AverageLatency == 0 {
		metrics.AverageLatency = result.ExecutionTime
	} else {
		// Exponential moving average
		alpha := 0.1
		metrics.AverageLatency = time.Duration(float64(metrics.AverageLatency)*(1-alpha) + float64(result.ExecutionTime)*alpha)
	}
	
	// Update last used time
	metrics.LastUsed = time.Now()
	
	// Copy any performance metrics from the result if available
	if result.Metrics != nil {
		// Initialize performance metrics if nil
		if metrics.Performance == nil {
			metrics.Performance = &PerformanceMetrics{}
		}
		
		// Append arrays from result.Metrics into metrics.Performance
		metrics.Performance.ExecutionTimeMs = append(metrics.Performance.ExecutionTimeMs, result.Metrics.ExecutionTimeMs...)
		metrics.Performance.MemoryUsageBytes = append(metrics.Performance.MemoryUsageBytes, result.Metrics.MemoryUsageBytes...)
		metrics.Performance.NetworkBandwidth = append(metrics.Performance.NetworkBandwidth, result.Metrics.NetworkBandwidth...)
		
		// Update quality and efficiency scores
		metrics.Performance.QualityScore = result.Metrics.QualityScore
		metrics.Performance.EfficiencyScore = result.Metrics.EfficiencyScore
		
		// Cap history to 100 entries
		if len(metrics.Performance.ExecutionTimeMs) > 100 {
			metrics.Performance.ExecutionTimeMs = metrics.Performance.ExecutionTimeMs[len(metrics.Performance.ExecutionTimeMs)-100:]
		}
		if len(metrics.Performance.MemoryUsageBytes) > 100 {
			metrics.Performance.MemoryUsageBytes = metrics.Performance.MemoryUsageBytes[len(metrics.Performance.MemoryUsageBytes)-100:]
		}
		if len(metrics.Performance.NetworkBandwidth) > 100 {
			metrics.Performance.NetworkBandwidth = metrics.Performance.NetworkBandwidth[len(metrics.Performance.NetworkBandwidth)-100:]
		}
		
		// Update metrics.AverageLatency from the average of ExecutionTimeMs
		if len(metrics.Performance.ExecutionTimeMs) > 0 {
			totalTime := 0.0
			for _, execTime := range metrics.Performance.ExecutionTimeMs {
				totalTime += execTime
			}
			avgTimeMs := totalTime / float64(len(metrics.Performance.ExecutionTimeMs))
			metrics.AverageLatency = time.Duration(avgTimeMs) * time.Millisecond
		}
	}
	
	return nil
}

// GetRecommendations returns recommendations for improving partitioning
func (epm *EnhancedPartitionManagerImpl) GetRecommendations(ctx context.Context, req *PartitionRequest) ([]*PartitionRecommendation, error) {
	epm.mu.RLock()
	defer epm.mu.RUnlock()
	
	// Return existing recommendations
	recommendations := make([]*PartitionRecommendation, len(epm.recommendations))
	copy(recommendations, epm.recommendations)
	
	return recommendations, nil
}

// generateRecommendationsAsync generates recommendations asynchronously
func (epm *EnhancedPartitionManagerImpl) generateRecommendationsAsync(ctx context.Context, req *PartitionRequest) {
	recommendations := epm.generateRecommendations(req)
	
	epm.mu.Lock()
	defer epm.mu.Unlock()
	
	// Add new recommendations
	epm.recommendations = append(epm.recommendations, recommendations...)
	
	// Keep only recent recommendations
	if len(epm.recommendations) > epm.config.RecommendationLimit {
		epm.recommendations = epm.recommendations[len(epm.recommendations)-epm.config.RecommendationLimit:]
	}
}

// generateRecommendations generates recommendations for a request
func (epm *EnhancedPartitionManagerImpl) generateRecommendations(req *PartitionRequest) []*PartitionRecommendation {
	var recommendations []*PartitionRecommendation
	
	// Strategy recommendations
	if req.Model != nil && req.Model.Parameters >= 30_000_000_000 {
		recommendations = append(recommendations, &PartitionRecommendation{
			Type:        RecommendationStrategy,
			Priority:    RecommendationPriorityHigh,
			Title:       "Consider Hybrid Parallelism",
			Description: "For very large models (30B+ parameters), hybrid parallelism often provides the best performance",
			Impact:      "20-40% improvement in throughput",
			Actions:     []string{"Enable hybrid_parallelism strategy", "Ensure high-speed interconnect"},
		})
	}
	
	// Resource recommendations
	lowMemoryStrategies := []string{}
	for strategyName, metrics := range epm.strategyMetrics {
		if metrics.FailureCount > 0 {
			recentFailureRate := float64(metrics.FailureCount) / float64(metrics.UsageCount)
			if recentFailureRate > 0.2 { // >20% failure rate
				lowMemoryStrategies = append(lowMemoryStrategies, strategyName)
			}
		}
	}
	
	if len(lowMemoryStrategies) > 0 {
		recommendations = append(recommendations, &PartitionRecommendation{
			Type:        RecommendationResources,
			Priority:    RecommendationPriorityMedium,
			Title:       "Memory Optimization Needed",
			Description: "Some strategies are failing due to memory constraints",
			Impact:      "Reduce failure rate and improve reliability",
			Actions:     []string{"Increase node memory", "Enable gradient checkpointing", "Consider memory-efficient strategies"},
		})
	}
	
	return recommendations
}

// convertTaskToRequest converts an DistributedTask to a PartitionRequest
func (epm *EnhancedPartitionManagerImpl) convertTaskToRequest(task *api_types.DistributedTask) (*PartitionRequest, error) {
	// Extract model information
	var modelInfo *ModelInfo
	if task.ModelName != "" {
		modelInfo = &ModelInfo{
			Name:       task.ModelName,
			Path:       task.ModelName, // Using model name as path for now
			Size:       0,              // Size would need to be fetched from elsewhere
			Parameters: extractParameterCount(task.ModelName), // Heuristic extraction
			Family:     extractModelFamily(task.ModelName),
		}
	}
	
	// Create partition options
	options := &PartitionOptions{
		MaxNodes:         10, // Default
		MinNodes:         1,
		MemoryThreshold:  0.8,
		LoadBalance:      true,
		OptimizeFor:      OptimizeBalance, // Default optimization target
		CustomParams:     make(map[string]interface{}),
	}
	
	// Create the partition request
	req := &PartitionRequest{
		TaskID:      string(task.ID),
		Model:       modelInfo,
		Options:     options,
		Constraints: &PartitionConstraints{},
	}
	
	return req, nil
}

// getAvailableNodes gets available nodes for partitioning
// Expected metadata shape:
// - preferred_node_ids: []string - List of preferred node IDs
// - preferred_nodes_capabilities: []map[string]interface{} or []any - Node capability snapshots
//   Each capability map contains: {id, cpu_cores, mem_total, mem_available, gpu_count, net_bw, net_latency}
func (epm *EnhancedPartitionManagerImpl) getAvailableNodes(task *api_types.DistributedTask) ([]*NodeInfo, error) {
	// Check if metadata is available
	if task.Metadata != nil {
		// First, try to get node capabilities from metadata
		if caps, ok := task.Metadata["preferred_nodes_capabilities"]; ok {
			// Handle both []map[string]interface{} and []any types
			var capsList []map[string]interface{}

			switch v := caps.(type) {
			case []map[string]interface{}:
				capsList = v
			case []interface{}:
				// Convert []interface{} to []map[string]interface{}
				for _, item := range v {
					if capMap, ok := item.(map[string]interface{}); ok {
						capsList = append(capsList, capMap)
					}
				}
			}

			if len(capsList) > 0 {
				nodes := make([]*NodeInfo, 0, len(capsList))

				// Also get preferred_node_ids if available for ordering
				var preferredNodeIDs []string
				if ids, ok := task.Metadata["preferred_node_ids"].([]string); ok {
					preferredNodeIDs = ids
				}

				// Build NodeInfo from capability snapshots
				for _, capMap := range capsList {
					// Parse node ID
					nodeID := ""
					if id, ok := capMap["id"].(string); ok {
						nodeID = id
					}
					if nodeID == "" {
						continue // Skip entries without ID
					}

					// Skip if we have preferred_node_ids and this ID is not in the list
					if len(preferredNodeIDs) > 0 {
						found := false
						for _, prefID := range preferredNodeIDs {
							if prefID == nodeID {
								found = true
								break
							}
						}
						if !found {
							continue
						}
					}

					// Parse CPU cores with type safety
					cpuCores := 8 // Default
					switch v := capMap["cpu_cores"].(type) {
					case int:
						cpuCores = v
					case float64:
						cpuCores = int(v)
					case int64:
						cpuCores = int(v)
					}
					if cpuCores < 1 {
						cpuCores = 1 // Clamp to minimum
					}

					// Parse memory total (bytes)
					memTotal := int64(32 * 1024 * 1024 * 1024) // Default 32GB
					switch v := capMap["mem_total"].(type) {
					case int64:
						memTotal = v
					case float64:
						memTotal = int64(v)
					case int:
						memTotal = int64(v)
					}
					if memTotal < 1024*1024*1024 { // Min 1GB
						memTotal = 1024 * 1024 * 1024
					}

					// Parse memory available (bytes)
					memAvailable := int64(24 * 1024 * 1024 * 1024) // Default 24GB
					switch v := capMap["mem_available"].(type) {
					case int64:
						memAvailable = v
					case float64:
						memAvailable = int64(v)
					case int:
						memAvailable = int64(v)
					}
					if memAvailable < 0 || memAvailable > memTotal {
						memAvailable = int64(float64(memTotal) * 0.75) // Default to 75% of total
					}

					// Parse GPU count
					gpuCount := 0 // Default
					switch v := capMap["gpu_count"].(type) {
					case int:
						gpuCount = v
					case float64:
						gpuCount = int(v)
					case int64:
						gpuCount = int(v)
					}
					if gpuCount < 0 {
						gpuCount = 0
					}

					// Parse network bandwidth (Gbps)
					netBandwidth := 25.0 // Default 25 Gbps
					switch v := capMap["net_bw"].(type) {
					case float64:
						netBandwidth = v
					case float32:
						netBandwidth = float64(v)
					case int:
						netBandwidth = float64(v)
					}
					if netBandwidth < 0.1 {
						netBandwidth = 1.0 // Min 1 Gbps
					}

					// Parse network latency (ms)
					netLatency := 1.0 // Default 1ms
					switch v := capMap["net_latency"].(type) {
					case float64:
						netLatency = v
					case float32:
						netLatency = float64(v)
					case int:
						netLatency = float64(v)
					}
					if netLatency < 0.01 {
						netLatency = 1.0 // Default to 1ms if too low
					}

					// Build NodeInfo with parsed capabilities
					node := &NodeInfo{
						ID:      nodeID,
						Address: fmt.Sprintf("%s:8080", nodeID), // Address can be resolved later
						Status:  NodeStatusActive,
						Capabilities: &NodeCapabilities{
							CPU: &CPUCapabilities{
								Cores:          cpuCores,
								ThreadsPerCore: 2, // Default
								Architecture:   "x86_64",
								Frequency:      3.2, // Default GHz
								CacheSize:      16 * 1024 * 1024, // Default 16MB
							},
							Memory: &MemoryCapabilities{
								TotalBytes:     memTotal,
								AvailableBytes: memAvailable,
								Type:           "DDR4",
								Speed:          3200,
								Bandwidth:      25.6, // GB/s default
							},
							Network: &NetworkCapabilities{
								Bandwidth:        netBandwidth,
								Latency:          netLatency,
								InterconnectType: "Ethernet",
								Topology:         "star",
							},
						},
						CurrentLoad: &NodeLoad{
							CPUUtilization:     0.3, // Default estimates
							MemoryUtilization:  1.0 - (float64(memAvailable) / float64(memTotal)),
							GPUUtilization:     0.2,
							NetworkUtilization: 0.1,
							ActiveTasks:        0,
							QueuedTasks:        0,
						},
					}

					// Add GPU capabilities if GPUs are present
					if gpuCount > 0 {
						node.Capabilities.GPU = &GPUCapabilities{
							Count:             gpuCount,
							Model:             "NVIDIA GPU", // Generic model
							MemoryPerGPU:      24 * 1024 * 1024 * 1024, // Default 24GB per GPU
							TotalMemory:       int64(gpuCount) * 24 * 1024 * 1024 * 1024,
							ComputeCapability: "8.0",
							CudaCores:         10000, // Default estimate
						}
					}

					// Add storage capabilities with defaults
					node.Capabilities.Storage = &StorageCapabilities{
						TotalBytes:     1024 * 1024 * 1024 * 1024, // 1TB default
						AvailableBytes: 512 * 1024 * 1024 * 1024,  // 512GB available
						Type:           "SSD",
						ReadSpeed:      2000,  // MB/s
						WriteSpeed:     1500,  // MB/s
					}

					nodes = append(nodes, node)
				}

				// If we got valid nodes from capabilities, return them
				if len(nodes) > 0 {
					// Maintain order based on preferred_node_ids if provided
					if len(preferredNodeIDs) > 0 {
						orderedNodes := make([]*NodeInfo, 0, len(nodes))
						for _, prefID := range preferredNodeIDs {
							for _, node := range nodes {
								if node.ID == prefID {
									orderedNodes = append(orderedNodes, node)
									break
								}
							}
						}
						return orderedNodes, nil
					}
					return nodes, nil
				}
			}
		}

		// Fall back to preferred_node_ids with synthesized capabilities
		if preferredNodeIDs, ok := task.Metadata["preferred_node_ids"].([]string); ok && len(preferredNodeIDs) > 0 {
			// Use the preferred node IDs from the engine
			nodes := make([]*NodeInfo, 0, len(preferredNodeIDs))
			for i, nodeID := range preferredNodeIDs {
				nodes = append(nodes, &NodeInfo{
					ID:      nodeID,
					Address: fmt.Sprintf("node-%d:8080", i+1), // Address can be resolved later if needed
					Status:  NodeStatusActive,
					Capabilities: &NodeCapabilities{
						CPU: &CPUCapabilities{
							Cores:        8,
							ThreadsPerCore: 2,
							Architecture: "x86_64",
							Frequency:    3.2,
							CacheSize:    16 * 1024 * 1024, // 16MB
						},
						Memory: &MemoryCapabilities{
							TotalBytes:     32 * 1024 * 1024 * 1024, // 32GB
							AvailableBytes: 24 * 1024 * 1024 * 1024, // 24GB available
							Type:          "DDR4",
							Speed:         3200,
							Bandwidth:     25.6, // GB/s
						},
						GPU: &GPUCapabilities{
							Count:         1,
							Model:         "NVIDIA RTX 4090",
							MemoryPerGPU:  24 * 1024 * 1024 * 1024, // 24GB
							TotalMemory:   24 * 1024 * 1024 * 1024,
							ComputeCapability: "8.9",
							CudaCores:     16384,
						},
						Network: &NetworkCapabilities{
							Bandwidth:        25.0, // 25 Gbps
							Latency:          1.0,   // 1ms
							InterconnectType: "Ethernet",
							Topology:        "star",
						},
						Storage: &StorageCapabilities{
							TotalBytes:     1024 * 1024 * 1024 * 1024, // 1TB
							AvailableBytes: 512 * 1024 * 1024 * 1024,  // 512GB available
							Type:          "NVMe SSD",
							ReadSpeed:     3500,  // MB/s
							WriteSpeed:    3000,  // MB/s
						},
					},
					CurrentLoad: &NodeLoad{
						CPUUtilization:    0.3,
						MemoryUtilization: 0.4,
						GPUUtilization:    0.2,
						NetworkUtilization: 0.1,
						ActiveTasks:       2,
						QueuedTasks:       0,
					},
				})
			}
			return nodes, nil
		}
	}
	
	// Fallback to default implementation if no preferred nodes provided
	// This is a simplified implementation
	// In practice, this would query the node manager or cluster state
	
	nodes := []*NodeInfo{
		{
			ID:      "node-1",
			Address: "192.168.1.10:8080",
			Status:  NodeStatusActive,
			Capabilities: &NodeCapabilities{
				CPU: &CPUCapabilities{
					Cores:        8,
					ThreadsPerCore: 2,
					Architecture: "x86_64",
					Frequency:    3.2,
					CacheSize:    16 * 1024 * 1024, // 16MB
				},
				Memory: &MemoryCapabilities{
					TotalBytes:     32 * 1024 * 1024 * 1024, // 32GB
					AvailableBytes: 24 * 1024 * 1024 * 1024, // 24GB available
					Type:          "DDR4",
					Speed:         3200,
					Bandwidth:     25.6, // GB/s
				},
				GPU: &GPUCapabilities{
					Count:         1,
					Model:         "NVIDIA RTX 4090",
					MemoryPerGPU:  24 * 1024 * 1024 * 1024, // 24GB
					TotalMemory:   24 * 1024 * 1024 * 1024,
					ComputeCapability: "8.9",
					CudaCores:     16384,
				},
				Network: &NetworkCapabilities{
					Bandwidth:        25.0, // 25 Gbps
					Latency:          1.0,   // 1ms
					InterconnectType: "Ethernet",
					Topology:        "star",
				},
				Storage: &StorageCapabilities{
					TotalBytes:     1024 * 1024 * 1024 * 1024, // 1TB
					AvailableBytes: 512 * 1024 * 1024 * 1024,  // 512GB available
					Type:          "NVMe SSD",
					ReadSpeed:     3500,  // MB/s
					WriteSpeed:    3000,  // MB/s
				},
			},
			CurrentLoad: &NodeLoad{
				CPUUtilization:    0.3,
				MemoryUtilization: 0.4,
				GPUUtilization:    0.2,
				NetworkUtilization: 0.1,
				ActiveTasks:       2,
				QueuedTasks:       0,
			},
		},
	}
	
	return nodes, nil
}

// recordStrategySelection records a strategy selection in history
func (epm *EnhancedPartitionManagerImpl) recordStrategySelection(task *api_types.DistributedTask, strategy string, model *ModelInfo, nodes []*NodeInfo, success bool) {
	epm.mu.Lock()
	defer epm.mu.Unlock()
	
	record := &StrategySelectionRecord{
		ID:               fmt.Sprintf("selection_%d", time.Now().UnixNano()),
		Timestamp:        time.Now(),
		TaskID:           string(task.ID),
		SelectedStrategy: strategy,
		ModelInfo:        model,
		NodeCount:        len(nodes),
		Success:          success,
		Metadata:         make(map[string]interface{}),
	}
	
	// Get alternative strategies
	for strategyName, strat := range epm.strategies {
		if strategyName != strategy {
			req := &PartitionRequest{Model: model}
			if strat.CanHandle(req) {
				record.AlternativeStrategies = append(record.AlternativeStrategies, strategyName)
			}
		}
	}
	
	epm.selectionHistory = append(epm.selectionHistory, record)
	
	// Keep only recent history
	if len(epm.selectionHistory) > epm.config.MaxHistorySize {
		epm.selectionHistory = epm.selectionHistory[len(epm.selectionHistory)-epm.config.MaxHistorySize:]
	}
}

// Helper functions

// extractParameterCount extracts parameter count from model name
func extractParameterCount(modelName string) int64 {
	// This is a heuristic extraction - in practice would be more sophisticated
	if contains(modelName, "70b") || contains(modelName, "70B") {
		return 70_000_000_000
	}
	if contains(modelName, "33b") || contains(modelName, "33B") {
		return 33_000_000_000
	}
	if contains(modelName, "13b") || contains(modelName, "13B") {
		return 13_000_000_000
	}
	if contains(modelName, "7b") || contains(modelName, "7B") {
		return 7_000_000_000
	}
	if contains(modelName, "3b") || contains(modelName, "3B") {
		return 3_000_000_000
	}
	return 7_000_000_000 // Default
}

// extractModelFamily extracts model family from model name
func extractModelFamily(modelName string) string {
	modelNameLower := strings.ToLower(modelName)
	if contains(modelNameLower, "llama") {
		return "llama"
	}
	if contains(modelNameLower, "gpt") {
		return "gpt"
	}
	if contains(modelNameLower, "bert") {
		return "bert"
	}
	if contains(modelNameLower, "claude") {
		return "claude"
	}
	return "generic"
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && 
		(strings.Contains(s, substr))))
}

// Base partition manager methods
func (bpm *BasePartitionManagerImpl) Partition(ctx context.Context, task *api_types.DistributedTask) (*PartitionPlan, error) {
	// Basic implementation - would be enhanced
	return nil, fmt.Errorf("base partition manager partition not implemented")
}

func (bpm *BasePartitionManagerImpl) SelectStrategy(task *api_types.DistributedTask, model *ModelInfo, options *PartitionOptions) (string, error) {
	// Basic strategy selection
	if model != nil && model.Parameters > 30_000_000_000 {
		return "hybrid_parallelism", nil
	}
	if model != nil && model.Parameters > 7_000_000_000 {
		return "pipeline_parallelism", nil
	}
	return "layer_parallelism", nil
}

func (bpm *BasePartitionManagerImpl) ValidatePartition(ctx context.Context, plan *PartitionPlan, nodes []*NodeInfo) (*PartitionValidationResult, error) {
	// Basic validation
	return &PartitionValidationResult{Valid: true, Score: 1.0}, nil
}

func (bpm *BasePartitionManagerImpl) GetAvailableStrategies() []string {
	bpm.mu.RLock()
	defer bpm.mu.RUnlock()
	
	strategies := make([]string, 0, len(bpm.strategies))
	for name := range bpm.strategies {
		strategies = append(strategies, name)
	}
	return strategies
}

func (bpm *BasePartitionManagerImpl) RegisterStrategy(name string, strategy PartitionStrategy) error {
	bpm.mu.Lock()
	defer bpm.mu.Unlock()
	
	bpm.strategies[name] = strategy
	return nil
}

func (bpm *BasePartitionManagerImpl) GetStrategy(name string) (PartitionStrategy, error) {
	bpm.mu.RLock()
	defer bpm.mu.RUnlock()
	
	strategy, exists := bpm.strategies[name]
	if !exists {
		return nil, fmt.Errorf("strategy %s not found", name)
	}
	return strategy, nil
}
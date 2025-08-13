package partitioning

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"time"
)

// EnhancedPartitionManager extends the partition manager with advanced features
type EnhancedPartitionManager struct {
	*PartitionManager // Embed base manager

	// Enhanced strategies
	enhancedStrategies map[string]PartitionStrategy

	// Performance tracking
	strategyPerformance map[string]*StrategyPerformance

	// Adaptive selection
	selectionHistory []*StrategySelection

	// Metrics
	metrics *EnhancedPartitionMetrics

	// Lifecycle
	mu      sync.RWMutex
	started bool
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// EnhancedPartitionMetrics tracks enhanced partitioning metrics
type EnhancedPartitionMetrics struct {
	// Performance tracking
	PerformanceHistorySize     int64   `json:"performance_history_size"`
	AveragePerformanceScore    float64 `json:"average_performance_score"`
	PerformanceTrackingEnabled bool    `json:"performance_tracking_enabled"`

	// Adaptive selection
	SelectionHistorySize int64         `json:"selection_history_size"`
	SelectionSuccessRate float64       `json:"selection_success_rate"`
	AverageSelectionTime time.Duration `json:"average_selection_time"`

	// Strategy weights
	StrategyWeights map[string]float64 `json:"strategy_weights"`

	// Learning metrics
	LearningRate float64   `json:"learning_rate"`
	Accuracy     float64   `json:"accuracy"`
	LastUpdated  time.Time `json:"last_updated"`

	// Timestamps
	LastSelection         *time.Time `json:"last_selection,omitempty"`
	LastPerformanceUpdate *time.Time `json:"last_performance_update,omitempty"`
}

// StrategyPerformance tracks performance metrics for partitioning strategies
type StrategyPerformance struct {
	TotalExecutions      int64         `json:"total_executions"`
	SuccessfulExecutions int64         `json:"successful_executions"`
	FailedExecutions     int64         `json:"failed_executions"`
	AverageLatency       time.Duration `json:"average_latency"`
	AverageThroughput    float64       `json:"average_throughput"`
	LastUsed             time.Time     `json:"last_used"`
	SuccessRate          float64       `json:"success_rate"`
	ErrorRate            float64       `json:"error_rate"`
	PerformanceScore     float64       `json:"performance_score"`
}

// StrategySelection represents a strategy selection decision
type StrategySelection struct {
	ID                  string                 `json:"id"`
	Timestamp           time.Time              `json:"timestamp"`
	StrategyName        string                 `json:"strategy_name"`
	TaskID              string                 `json:"task_id"`
	ModelName           string                 `json:"model_name"`
	SelectedAt          time.Time              `json:"selected_at"`
	ExecutionLatency    time.Duration          `json:"execution_latency"`
	ExecutionThroughput float64                `json:"execution_throughput"`
	Success             bool                   `json:"success"`
	Metadata            map[string]interface{} `json:"metadata"`
}

// PipelineParallelismStrategy implements pipeline parallelism for sequential models
type PipelineParallelismStrategy struct {
	name    string
	metrics *StrategyMetrics
}

// TensorParallelismStrategy implements tensor parallelism for intra-layer operations
type TensorParallelismStrategy struct {
	name    string
	metrics *StrategyMetrics
}

// HybridParallelismStrategy combines pipeline and tensor parallelism
type HybridParallelismStrategy struct {
	name    string
	metrics *StrategyMetrics
}

// AdaptivePartitioningStrategy adapts partitioning based on workload analysis
type AdaptivePartitioningStrategy struct {
	name       string
	metrics    *StrategyMetrics
	thresholds map[string]float64
	learning   bool
	accuracy   float64
	history    []*PartitionResult
	historyMu  sync.RWMutex
}

// NewEnhancedPartitionManager creates a new enhanced partition manager
func NewEnhancedPartitionManager(baseManager *PartitionManager) *EnhancedPartitionManager {
	ctx, cancel := context.WithCancel(context.Background())

	// Create enhanced manager
	epm := &EnhancedPartitionManager{
		PartitionManager:    baseManager,
		enhancedStrategies:  make(map[string]PartitionStrategy),
		strategyPerformance: make(map[string]*StrategyPerformance),
		selectionHistory:    make([]*StrategySelection, 0),
		metrics: &EnhancedPartitionMetrics{
			LastUpdated: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize components
	epm.initializeComponents()

	return epm
}

// initializeComponents initializes enhanced partition manager components
func (epm *EnhancedPartitionManager) initializeComponents() {
	// Initialize enhanced strategies
	epm.registerEnhancedStrategies()

	// Initialize performance tracking
	epm.initializePerformanceTracking()

	// Initialize metrics
	epm.initializeMetrics()
}

// registerEnhancedStrategies registers enhanced partitioning strategies
func (epm *EnhancedPartitionManager) registerEnhancedStrategies() {
	// Register pipeline parallelism strategy
	epm.enhancedStrategies["pipeline_parallel"] = NewPipelineParallelismStrategy()

	// Register tensor parallelism strategy
	epm.enhancedStrategies["tensor_parallel"] = NewTensorParallelismStrategy()

	// Register hybrid parallelism strategy
	epm.enhancedStrategies["hybrid_parallel"] = NewHybridParallelismStrategy()

	// Register adaptive partitioning strategy
	epm.enhancedStrategies["adaptive"] = NewAdaptivePartitioningStrategy()

	// Initialize strategy performance tracking
	for name := range epm.enhancedStrategies {
		epm.strategyPerformance[name] = &StrategyPerformance{
			LastUsed: time.Now(),
		}
	}
}

// initializePerformanceTracking initializes performance tracking
func (epm *EnhancedPartitionManager) initializePerformanceTracking() {
	// Initialize performance tracking settings
	epm.metrics.PerformanceTrackingEnabled = true
	epm.metrics.PerformanceHistorySize = 1000
	epm.metrics.AveragePerformanceScore = 0.7 // Initial score

	// Initialize selection history settings
	epm.metrics.SelectionHistorySize = 1000

	// Initialize adaptive selection settings
	epm.metrics.SelectionSuccessRate = 0.8                   // Initial success rate
	epm.metrics.AverageSelectionTime = 50 * time.Millisecond // Initial average

	// Initialize strategy weights
	epm.metrics.StrategyWeights = map[string]float64{
		"layerwise":          0.25,
		"data_split":         0.20,
		"task_parallel":      0.15,
		"sequence_parallel":  0.10,
		"attention_parallel": 0.10,
		"pipeline_parallel":  0.10,
		"tensor_parallel":    0.05,
		"hybrid_parallel":    0.03,
		"adaptive":           0.02,
	}

	// Initialize learning settings
	epm.metrics.LearningRate = 0.1
	epm.metrics.Accuracy = 0.7 // Initial accuracy

	// Initialize timestamps
	epm.metrics.LastUpdated = time.Now()

	// Initialize performance tracking
	if epm.metrics.PerformanceTrackingEnabled {
		epm.wg.Add(1)
		go epm.performanceTrackingTask()
	}

	// Initialize adaptive selection
	epm.wg.Add(1)
	go epm.adaptiveSelectionTask()
}

// initializeMetrics initializes enhanced partitioning metrics
func (epm *EnhancedPartitionManager) initializeMetrics() {
	// Initialize enhanced metrics
	epm.metrics.LastUpdated = time.Now()

	// Initialize strategy metrics
	epm.metrics.StrategyWeights = make(map[string]float64)

	// Initialize learning metrics
	epm.metrics.LearningRate = 0.1
	epm.metrics.Accuracy = 0.7 // Initial accuracy

	// Initialize timestamps
	epm.metrics.LastUpdated = time.Now()
}

// performanceTrackingTask tracks performance metrics
func (epm *EnhancedPartitionManager) performanceTrackingTask() {
	defer epm.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-epm.ctx.Done():
			return
		case <-ticker.C:
			epm.trackPerformance()
		}
	}
}

// trackPerformance tracks performance metrics
func (epm *EnhancedPartitionManager) trackPerformance() {
	epm.mu.Lock()
	defer epm.mu.Unlock()

	now := time.Now()

	// Update metrics
	epm.metrics.LastPerformanceUpdate = &now
	epm.metrics.LastUpdated = now

	// Calculate performance score based on recent selections
	if len(epm.selectionHistory) > 0 {
		recentSelections := epm.selectionHistory
		if len(recentSelections) > 100 {
			recentSelections = recentSelections[len(recentSelections)-100:]
		}

		totalSelections := len(recentSelections)
		successfulSelections := 0
		totalLatency := time.Duration(0)
		totalThroughput := 0.0

		for _, selection := range recentSelections {
			if selection.Success {
				successfulSelections++
				totalLatency += selection.ExecutionLatency
				totalThroughput += selection.ExecutionThroughput
			}
		}

		if totalSelections > 0 {
			epm.metrics.SelectionHistorySize = int64(totalSelections)
			epm.metrics.SelectionSuccessRate = float64(successfulSelections) / float64(totalSelections)
		}

		if successfulSelections > 0 {
			epm.metrics.AverageSelectionTime = totalLatency / time.Duration(successfulSelections)
			epm.metrics.AveragePerformanceScore = totalThroughput / float64(successfulSelections)
		}
	}
}

// adaptiveSelectionTask performs adaptive selection
func (epm *EnhancedPartitionManager) adaptiveSelectionTask() {
	defer epm.wg.Done()

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-epm.ctx.Done():
			return
		case <-ticker.C:
			epm.adaptSelection()
		}
	}
}

// adaptSelection adapts selection based on performance
func (epm *EnhancedPartitionManager) adaptSelection() {
	epm.mu.Lock()
	defer epm.mu.Unlock()

	start := time.Now()

	// Update metrics
	epm.metrics.LastUpdated = time.Now()

	// Calculate performance score
	if len(epm.selectionHistory) > 10 {
		// Calculate recent performance
		recentSelections := epm.selectionHistory[len(epm.selectionHistory)-10:]

		totalLatency := time.Duration(0)
		totalThroughput := 0.0
		successfulSelections := 0

		for _, selection := range recentSelections {
			if selection.Success {
				successfulSelections++
				totalLatency += selection.ExecutionLatency
				totalThroughput += selection.ExecutionThroughput
			}
		}

		if successfulSelections > 0 {
			avgLatency := totalLatency / time.Duration(successfulSelections)
			avgThroughput := totalThroughput / float64(successfulSelections)

			// Update accuracy with exponential moving average
			alpha := epm.metrics.LearningRate
			performanceScore := 1.0 - (float64(avgLatency)/float64(100*time.Millisecond)+
				(1.0-avgThroughput/100.0))/2.0

			epm.metrics.Accuracy = alpha*performanceScore + (1-alpha)*epm.metrics.Accuracy

			// Update strategy weights based on performance
			epm.updateStrategyWeights(performanceScore)
		}
	}

	// Update timestamps
	now := time.Now()
	epm.metrics.LastUpdated = now
	epm.metrics.AverageSelectionTime = time.Since(start)
}

// updateStrategyWeights updates strategy weights based on performance
func (epm *EnhancedPartitionManager) updateStrategyWeights(performanceScore float64) {
	// Get recent selection history
	recentHistory := epm.selectionHistory
	if len(recentHistory) > 100 {
		recentHistory = recentHistory[len(recentHistory)-100:]
	}

	// Calculate performance for each strategy
	strategyPerformance := make(map[string]float64)
	strategyCounts := make(map[string]int)

	for _, selection := range recentHistory {
		strategyPerformance[selection.StrategyName] += performanceScore
		strategyCounts[selection.StrategyName]++
	}

	// Calculate average performance for each strategy
	for strategy, totalPerformance := range strategyPerformance {
		count := strategyCounts[strategy]
		if count > 0 {
			avgPerformance := totalPerformance / float64(count)

			// Update weight with exponential moving average
			alpha := epm.metrics.LearningRate
			currentWeight := epm.metrics.StrategyWeights[strategy]
			newWeight := alpha*avgPerformance + (1-alpha)*currentWeight

			// Clamp weight between 0.01 and 0.99
			if newWeight < 0.01 {
				newWeight = 0.01
			}
			if newWeight > 0.99 {
				newWeight = 0.99
			}

			epm.metrics.StrategyWeights[strategy] = newWeight
		}
	}

	// Normalize weights to sum to 1.0
	totalWeight := 0.0
	for _, weight := range epm.metrics.StrategyWeights {
		totalWeight += weight
	}

	if totalWeight > 0 {
		for strategy, weight := range epm.metrics.StrategyWeights {
			epm.metrics.StrategyWeights[strategy] = weight / totalWeight
		}
	}
}

// GetAvailableStrategies returns all available strategies
func (epm *EnhancedPartitionManager) GetAvailableStrategies() []string {
	epm.mu.RLock()
	defer epm.mu.RUnlock()

	// Get base strategies
	baseStrategies := epm.PartitionManager.GetAvailableStrategies()

	// Get enhanced strategies
	enhancedStrategies := make([]string, 0, len(baseStrategies)+len(epm.enhancedStrategies))

	// Add base strategies
	enhancedStrategies = append(enhancedStrategies, baseStrategies...)

	// Add enhanced strategies
	for name := range epm.enhancedStrategies {
		enhancedStrategies = append(enhancedStrategies, name)
	}

	return enhancedStrategies
}

// GetStrategyMetrics returns strategy metrics
func (epm *EnhancedPartitionManager) GetStrategyMetrics() map[string]*StrategyMetrics {
	epm.mu.RLock()
	defer epm.mu.RUnlock()

	// Get base metrics
	baseMetrics := make(map[string]*StrategyMetrics)
	for _, strategy := range epm.PartitionManager.strategies {
		baseMetrics[strategy.GetName()] = strategy.GetMetrics()
	}

	// Get enhanced metrics
	enhancedMetrics := make(map[string]*StrategyMetrics)

	// Add base metrics
	for name, metrics := range baseMetrics {
		enhancedMetrics[name] = metrics
	}

	// Add enhanced strategy metrics
	for name, strategy := range epm.enhancedStrategies {
		enhancedMetrics[name] = strategy.GetMetrics()
	}

	return enhancedMetrics
}

// GetSelectionHistory returns selection history
func (epm *EnhancedPartitionManager) GetSelectionHistory() []*StrategySelection {
	epm.mu.RLock()
	defer epm.mu.RUnlock()

	// Create a copy to avoid race conditions
	history := make([]*StrategySelection, len(epm.selectionHistory))
	copy(history, epm.selectionHistory)

	return history
}

// GetEnhancedMetrics returns enhanced partitioning metrics
func (epm *EnhancedPartitionManager) GetEnhancedMetrics() *EnhancedPartitionMetrics {
	epm.mu.RLock()
	defer epm.mu.RUnlock()

	// Update enhanced metrics
	epm.metrics.LastUpdated = time.Now()

	// Update performance tracking metrics
	epm.metrics.PerformanceHistorySize = int64(len(epm.selectionHistory))

	// Update selection history metrics
	if len(epm.selectionHistory) > 0 {
		recentSelections := epm.selectionHistory
		if len(recentSelections) > 1000 {
			recentSelections = recentSelections[len(recentSelections)-1000:]
		}

		totalSelections := len(recentSelections)
		successfulSelections := 0
		totalLatency := time.Duration(0)
		totalThroughput := 0.0

		for _, selection := range recentSelections {
			if selection.Success {
				successfulSelections++
				totalLatency += selection.ExecutionLatency
				totalThroughput += selection.ExecutionThroughput
			}
		}

		if totalSelections > 0 {
			epm.metrics.SelectionHistorySize = int64(totalSelections)
			epm.metrics.SelectionSuccessRate = float64(successfulSelections) / float64(totalSelections)
		}

		if successfulSelections > 0 {
			epm.metrics.AverageSelectionTime = totalLatency / time.Duration(successfulSelections)
			epm.metrics.AveragePerformanceScore = totalThroughput / float64(successfulSelections)
		}
	}

	// Create a copy to avoid race conditions
	metrics := &EnhancedPartitionMetrics{
		PerformanceHistorySize:     epm.metrics.PerformanceHistorySize,
		AveragePerformanceScore:    epm.metrics.AveragePerformanceScore,
		PerformanceTrackingEnabled: epm.metrics.PerformanceTrackingEnabled,
		SelectionHistorySize:       epm.metrics.SelectionHistorySize,
		SelectionSuccessRate:       epm.metrics.SelectionSuccessRate,
		AverageSelectionTime:       epm.metrics.AverageSelectionTime,
		StrategyWeights:            make(map[string]float64),
		LearningRate:               epm.metrics.LearningRate,
		Accuracy:                   epm.metrics.Accuracy,
		LastUpdated:                epm.metrics.LastUpdated,
		LastSelection:              epm.metrics.LastSelection,
		LastPerformanceUpdate:      epm.metrics.LastPerformanceUpdate,
	}

	// Copy strategy weights
	for k, v := range epm.metrics.StrategyWeights {
		metrics.StrategyWeights[k] = v
	}

	return metrics
}

// Shutdown gracefully shuts down the enhanced partition manager
func (epm *EnhancedPartitionManager) Shutdown(ctx context.Context) error {
	epm.mu.Lock()
	defer epm.mu.Unlock()

	if !epm.started {
		return nil
	}

	slog.Info("shutting down enhanced partition manager")

	// Cancel context
	epm.cancel()

	// Wait for background tasks
	epm.wg.Wait()

	// Shutdown base manager
	// Shutdown is handled through context cancellation
	// epm.PartitionManager doesn't have a Shutdown method in the base implementation
	epm.started = false

	return nil
}

// NewPipelineParallelismStrategy creates a new pipeline parallelism strategy
func NewPipelineParallelismStrategy() *PipelineParallelismStrategy {
	return &PipelineParallelismStrategy{
		name: "pipeline_parallel",
		metrics: &StrategyMetrics{
			LastUsed: time.Now(),
		},
	}
}

// GetName returns the strategy name
func (pps *PipelineParallelismStrategy) GetName() string {
	return pps.name
}

// GetMetrics returns strategy metrics
func (pps *PipelineParallelismStrategy) GetMetrics() *StrategyMetrics {
	return pps.metrics
}

// CanHandle checks if this strategy can handle the task
func (pps *PipelineParallelismStrategy) CanHandle(task *PartitionTask) bool {
	// This strategy works well for models with many layers
		if layers := kv.Uint("llm.layers"); layers > 0 {
			return int(layers) > 20 // Only for models with many layers
		}
	}
	return false
}

// Partition implements pipeline parallelism partitioning
func (pps *PipelineParallelismStrategy) Partition(ctx context.Context, task *PartitionTask) (*PartitionPlan, error) {
	start := time.Now()

	// Get number of layers
	layerCount := 0
		if layers := kv.Uint("llm.layers"); layers > 0 {
			layerCount = int(layers)
		}
	}

	if layerCount == 0 {
		return nil, fmt.Errorf("unable to determine layer count")
	}

	nodeCount := len(task.Nodes)
	if nodeCount == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	// Calculate layers per stage
	layersPerStage := int(math.Ceil(float64(layerCount) / float64(nodeCount)))

	// Create partitions
	partitions := make([]*Partition, 0)
	stageIndex := 0

	for i := 0; i < layerCount; i += layersPerStage {
		end := i + layersPerStage
		if end > layerCount {
			end = layerCount
		}

		// Assign to node in round-robin fashion
		nodeIndex := stageIndex % nodeCount
		nodeID := task.Nodes[nodeIndex].ID

		partition := &Partition{
			ID:     fmt.Sprintf("partition_%s_stage_%d", task.ID, stageIndex),
			NodeID: nodeID,
			Type:   PartitionTypeLayer,
			Data: map[string]interface{}{
				"start_layer": i,
				"end_layer":   end,
				"layer_count": end - i,
			},
			Dependencies: []string{}, // Will be set later
			Metadata: map[string]interface{}{
				"stage": stageIndex,
			},
			EstimatedLatency: time.Duration((end-i)*10) * time.Millisecond, // Rough estimate
			EstimatedMemory:  int64((end - i) * 100 * 1024 * 1024),         // Rough estimate (100MB per layer)
		}

		// Set dependencies (each stage depends on the previous one)
		if stageIndex > 0 {
			partition.Dependencies = append(partition.Dependencies, fmt.Sprintf("partition_%s_stage_%d", task.ID, stageIndex-1))
		}

		partitions = append(partitions, partition)
		stageIndex++
	}

	// Create plan
	plan := &PartitionPlan{
		ID:                  fmt.Sprintf("plan_%s", task.ID),
		Strategy:            pps.GetName(),
		Partitions:          partitions,
		Metadata:            make(map[string]interface{}),
		CreatedAt:           time.Now(),
		EstimatedLatency:    time.Duration(layerCount*10) * time.Millisecond,
		EstimatedThroughput: 1.0, // Placeholder
		OptimizationScore:   0.8, // Placeholder
	}

	// Update metrics
	pps.metrics.TotalPartitions += int64(len(partitions))
	pps.metrics.SuccessfulPartitions += int64(len(partitions))
	pps.metrics.LastUsed = time.Now()
	pps.metrics.AverageLatency = (pps.metrics.AverageLatency*time.Duration(pps.metrics.SuccessfulPartitions-1) +
		time.Since(start)) / time.Duration(pps.metrics.SuccessfulPartitions)

	return plan, nil
}

// NewTensorParallelismStrategy creates a new tensor parallelism strategy
func NewTensorParallelismStrategy() *TensorParallelismStrategy {
	return &TensorParallelismStrategy{
		name: "tensor_parallel",
		metrics: &StrategyMetrics{
			LastUsed: time.Now(),
		},
	}
}

// GetName returns the strategy name
func (tps *TensorParallelismStrategy) GetName() string {
	return tps.name
}

// GetMetrics returns strategy metrics
func (tps *TensorParallelismStrategy) GetMetrics() *StrategyMetrics {
	return tps.metrics
}

// CanHandle checks if this strategy can handle the task
func (tps *TensorParallelismStrategy) CanHandle(task *PartitionTask) bool {
	// This strategy works for models with large context windows (>2048 tokens)
	contextLength := task.GetNumCtx()
	return contextLength > 2048 // Only for large context models
}

// Partition implements tensor parallelism partitioning
func (tps *TensorParallelismStrategy) Partition(ctx context.Context, task *PartitionTask) (*PartitionPlan, error) {
	start := time.Now()

	nodeCount := len(task.Nodes)
	if nodeCount == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	// Get context length
	contextLength := task.GetNumCtx()
	if contextLength == 0 {
		contextLength = 2048 // Default context length
	}

	// For tensor parallelism, we split the computation across nodes
	// rather than splitting layers
	partitions := make([]*Partition, nodeCount)

	// Split the context across nodes
	contextPerNode := contextLength / nodeCount
	remainder := contextLength % nodeCount

	for i := 0; i < nodeCount; i++ {
		startToken := i * contextPerNode
		endToken := startToken + contextPerNode
		if i < remainder {
			startToken += i
			endToken += i + 1
		} else {
			startToken += remainder
			endToken += remainder
		}

		partition := &Partition{
			ID:     fmt.Sprintf("partition_%s_tensor_%d", task.ID, i),
			NodeID: task.Nodes[i].ID,
			Type:   PartitionTypeData,
			Data: map[string]interface{}{
				"start_token": startToken,
				"end_token":   endToken,
				"token_count": endToken - startToken,
			},
			Dependencies: []string{}, // All partitions can run in parallel
			Metadata: map[string]interface{}{
				"tensor_split": i,
			},
			EstimatedLatency: time.Duration((endToken-startToken)*5) * time.Millisecond, // Rough estimate
			EstimatedMemory:  int64((endToken - startToken) * 2 * 1024),                 // Rough estimate (2KB per token)
		}

		partitions[i] = partition
	}

	// Create plan
	plan := &PartitionPlan{
		ID:                  fmt.Sprintf("plan_%s", task.ID),
		Strategy:            tps.GetName(),
		Partitions:          partitions,
		Metadata:            make(map[string]interface{}),
		CreatedAt:           time.Now(),
		EstimatedLatency:    time.Duration(contextLength*5) * time.Millisecond / time.Duration(nodeCount),
		EstimatedThroughput: float64(nodeCount), // Placeholder
		OptimizationScore:   0.7,                // Placeholder
	}

	// Update metrics
	tps.metrics.TotalPartitions += int64(len(partitions))
	tps.metrics.SuccessfulPartitions += int64(len(partitions))
	tps.metrics.LastUsed = time.Now()
	tps.metrics.AverageLatency = (tps.metrics.AverageLatency*time.Duration(tps.metrics.SuccessfulPartitions-1) +
		time.Since(start)) / time.Duration(tps.metrics.SuccessfulPartitions)

	return plan, nil
}

// NewHybridParallelismStrategy creates a new hybrid parallelism strategy
func NewHybridParallelismStrategy() *HybridParallelismStrategy {
	return &HybridParallelismStrategy{
		name: "hybrid_parallel",
		metrics: &StrategyMetrics{
			LastUsed: time.Now(),
		},
	}
}

// GetName returns the strategy name
func (hps *HybridParallelismStrategy) GetName() string {
	return hps.name
}

// GetMetrics returns strategy metrics
func (hps *HybridParallelismStrategy) GetMetrics() *StrategyMetrics {
	return hps.metrics
}

// CanHandle checks if this strategy can handle the task
func (hps *HybridParallelismStrategy) CanHandle(task *PartitionTask) bool {
	// This strategy works for large models with both many layers and large context
	layerCount := 0
		if layers := kv.Uint("llm.layers"); layers > 0 {
			layerCount = int(layers)
		}
	}

	contextLength := task.GetNumCtx()
	if contextLength == 0 {
		contextLength = 2048 // Default context length
	}

	return layerCount > 20 && contextLength > 2048
}

// Partition implements hybrid parallelism partitioning
func (hps *HybridParallelismStrategy) Partition(ctx context.Context, task *PartitionTask) (*PartitionPlan, error) {
	start := time.Now()

	// Get number of layers
	layerCount := 0
		if layers := kv.Uint("llm.layers"); layers > 0 {
			layerCount = int(layers)
		}
	}

	if layerCount == 0 {
		return nil, fmt.Errorf("unable to determine layer count")
	}

	nodeCount := len(task.Nodes)
	if nodeCount == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	// For hybrid approach, we divide nodes into pipeline stages
	// and split context within each stage

	// Determine pipeline stages (sqrt of nodes for balanced approach)
	pipelineStages := int(math.Sqrt(float64(nodeCount)))
	if pipelineStages < 2 {
		pipelineStages = 2
	}
	if pipelineStages > layerCount {
		pipelineStages = layerCount
	}

	nodesPerStage := nodeCount / pipelineStages
	if nodesPerStage == 0 {
		nodesPerStage = 1
	}

	// Calculate layers per stage
	layersPerStage := layerCount / pipelineStages
	if layersPerStage == 0 {
		layersPerStage = 1
	}

	partitions := make([]*Partition, 0)
	stageIndex := 0

	// Create pipeline stages
	for i := 0; i < layerCount; i += layersPerStage {
		endLayer := i + layersPerStage
		if endLayer > layerCount {
			endLayer = layerCount
		}

		// For each pipeline stage, split context across nodes in that stage
		stageNodes := make([]*NodeInfo, 0)
		for j := 0; j < nodesPerStage && (stageIndex*nodesPerStage+j) < nodeCount; j++ {
			nodeIndex := stageIndex*nodesPerStage + j
			if nodeIndex < len(task.Nodes) {
				stageNodes = append(stageNodes, task.Nodes[nodeIndex])
			}
		}

		if len(stageNodes) == 0 {
			continue
		}

		// Split context across nodes in this stage
		contextPerNode := task.GetNumCtx() / len(stageNodes)
		remainder := task.GetNumCtx() % len(stageNodes)

		for j, node := range stageNodes {
			startToken := j * contextPerNode
			endToken := startToken + contextPerNode
			if j < remainder {
				startToken += j
				endToken += j + 1
			} else {
				startToken += remainder
				endToken += remainder
			}

			partition := &Partition{
				ID:     fmt.Sprintf("partition_%s_hybrid_%d_%d", task.ID, stageIndex, j),
				NodeID: node.ID,
				Type:   PartitionTypeLayer,
				Data: map[string]interface{}{
					"start_layer": i,
					"end_layer":   endLayer,
					"layer_count": endLayer - i,
					"start_token": startToken,
					"end_token":   endToken,
					"token_count": endToken - startToken,
				},
				Dependencies: []string{}, // Will be set later
				Metadata: map[string]interface{}{
					"pipeline_stage": stageIndex,
					"tensor_split":   j,
				},
				EstimatedLatency: time.Duration((endLayer-i)*(endToken-startToken)*2) * time.Millisecond,
				EstimatedMemory:  int64((endLayer - i) * (endToken - startToken) * 2 * 1024), // Rough estimate
			}

			// Set dependencies (depends on previous pipeline stage)
			if stageIndex > 0 {
				// Depend on all partitions from previous stage
				for k := 0; k < len(stageNodes); k++ {
					partition.Dependencies = append(partition.Dependencies,
						fmt.Sprintf("partition_%s_hybrid_%d_%d", task.ID, stageIndex-1, k))
				}
			}

			partitions = append(partitions, partition)
		}

		stageIndex++
	}

	// Create plan
	plan := &PartitionPlan{
		ID:                  fmt.Sprintf("plan_%s", task.ID),
		Strategy:            hps.GetName(),
		Partitions:          partitions,
		Metadata:            make(map[string]interface{}),
		CreatedAt:           time.Now(),
		EstimatedLatency:    time.Duration(layerCount*task.GetNumCtx()*2) * time.Millisecond / time.Duration(nodeCount),
		EstimatedThroughput: float64(nodeCount), // Placeholder
		OptimizationScore:   0.9,                // High score for hybrid approach
	}

	// Update metrics
	hps.metrics.TotalPartitions += int64(len(partitions))
	hps.metrics.SuccessfulPartitions += int64(len(partitions))
	hps.metrics.LastUsed = time.Now()
	hps.metrics.AverageLatency = (hps.metrics.AverageLatency*time.Duration(hps.metrics.SuccessfulPartitions-1) +
		time.Since(start)) / time.Duration(hps.metrics.SuccessfulPartitions)

	return plan, nil
}

// NewAdaptivePartitioningStrategy creates a new adaptive partitioning strategy
func NewAdaptivePartitioningStrategy() *AdaptivePartitioningStrategy {
	return &AdaptivePartitioningStrategy{
		name: "adaptive",
		metrics: &StrategyMetrics{
			LastUsed: time.Now(),
		},
		thresholds: map[string]float64{
			"large_model":      5.0 * 1024 * 1024 * 1024, // 5GB
			"large_context":    2048,
			"many_layers":      20,
			"high_parallelism": 0.8,
		},
		learning: true,
		accuracy: 0.7, // Initial accuracy
		history:  make([]*PartitionResult, 0),
	}
}

// GetName returns the strategy name
func (aps *AdaptivePartitioningStrategy) GetName() string {
	return aps.name
}

// GetMetrics returns strategy metrics
func (aps *AdaptivePartitioningStrategy) GetMetrics() *StrategyMetrics {
	return aps.metrics
}

// CanHandle checks if this strategy can handle the task
func (aps *AdaptivePartitioningStrategy) CanHandle(task *PartitionTask) bool {
	// This strategy can handle any task
	return true
}

// Partition implements adaptive partitioning based on workload analysis
func (aps *AdaptivePartitioningStrategy) Partition(ctx context.Context, task *PartitionTask) (*PartitionPlan, error) {
	start := time.Now()

	// Analyze workload characteristics
	modelSize := aps.estimateModelSize(task)
	contextLength := task.GetNumCtx()
	layerCount := aps.estimateLayerCount(task)
	parallelizability := aps.estimateParallelizability(task)
	nodeCount := len(task.Nodes)

	// Select the best strategy based on workload analysis
	var plan *PartitionPlan
	var err error

	// For very large models, use pipeline parallelism
	if modelSize > aps.thresholds["large_model"] && layerCount > int(aps.thresholds["many_layers"]) {
		strategy := NewPipelineParallelismStrategy()
		plan, err = strategy.Partition(ctx, task)
	} else if contextLength > int(aps.thresholds["large_context"]) && parallelizability > aps.thresholds["high_parallelism"] {
		// For large context with high parallelizability, use tensor parallelism
		strategy := NewTensorParallelismStrategy()
		plan, err = strategy.Partition(ctx, task)
	} else if modelSize > aps.thresholds["large_model"] && contextLength > int(aps.thresholds["large_context"]) {
		// For both large model and large context, use hybrid parallelism
		strategy := NewHybridParallelismStrategy()
		plan, err = strategy.Partition(ctx, task)
	} else if nodeCount > 1 && layerCount > int(aps.thresholds["many_layers"]) {
		// For multi-node setups with sufficient layers, use pipeline parallelism
		strategy := NewPipelineParallelismStrategy()
		plan, err = strategy.Partition(ctx, task)
	} else {
		// Default to layerwise partitioning
		strategy := NewLayerwiseStrategy()
		plan, err = strategy.Partition(ctx, task)
	}

	if err != nil {
		aps.metrics.FailedPartitions++
		return nil, fmt.Errorf("failed to partition task: %w", err)
	}

	// Update metrics
	aps.metrics.TotalPartitions += int64(len(plan.Partitions))
	aps.metrics.SuccessfulPartitions += int64(len(plan.Partitions))
	aps.metrics.LastUsed = time.Now()
	aps.metrics.AverageLatency = (aps.metrics.AverageLatency*time.Duration(aps.metrics.SuccessfulPartitions-1) +
		time.Since(start)) / time.Duration(aps.metrics.SuccessfulPartitions)

	return plan, nil
}

// estimateModelSize estimates the size of a model
func (aps *AdaptivePartitioningStrategy) estimateModelSize(task *PartitionTask) float64 {
	}
	// Fallback estimation based on model name patterns
	return 4.0 * 1024 * 1024 * 1024 // 4GB default
}

// estimateLayerCount estimates the number of layers in a model
func (aps *AdaptivePartitioningStrategy) estimateLayerCount(task *PartitionTask) int {
		if layers := kv.Uint("llm.layers"); layers > 0 {
			return int(layers)
		}
	}
	// Fallback estimation
	return 24 // Default for many transformer models
}

// estimateParallelizability estimates how parallelizable a task is
func (aps *AdaptivePartitioningStrategy) estimateParallelizability(task *PartitionTask) float64 {
	// Factors that affect parallelizability:
	// 1. Model architecture (transformers are more parallelizable)
	// 2. Context length (longer contexts are more parallelizable)
	// 3. Batch size (larger batches are more parallelizable)

	contextLength := float64(task.GetNumCtx())
	batchSize := float64(1) // Default batch size

	// Base parallelizability on context length and batch size
	parallelizability := math.Min((contextLength/2048.0)*(batchSize/4.0), 1.0)

	// Adjust based on model type
	if task.Model != nil {
		// Check if model is a transformer (more parallelizable)
		if isTransformerModel(task.Model) {
			parallelizability *= 1.2
		}
	}

	return math.Min(parallelizability, 1.0)
}

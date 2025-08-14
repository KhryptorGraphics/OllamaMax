package partitioning

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/types"
)

// PartitionManager manages workload partitioning strategies
type PartitionManager struct {
	config     *Config
	strategies map[string]PartitionStrategy
	optimizer  *PartitionOptimizer
	analyzer   *WorkloadAnalyzer
}

// Config holds partitioning configuration
type Config struct {
	DefaultStrategy string `json:"default_strategy"`
	LayerThreshold  int    `json:"layer_threshold"`
	BatchSizeLimit  int    `json:"batch_size_limit"`
}

// PartitionStrategy defines the interface for partitioning strategies
type PartitionStrategy interface {
	Partition(ctx context.Context, task *PartitionTask) (*PartitionPlan, error)
	GetName() string
	GetMetrics() *StrategyMetrics
	CanHandle(task *PartitionTask) bool
}

// PartitionTask represents a task to be partitioned
type PartitionTask struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Model     *types.OllamaModel     `json:"model"`
	Options   map[string]interface{} `json:"options"`
	Nodes     []*NodeInfo            `json:"nodes"`
	Metadata  map[string]interface{} `json:"metadata"`
	Priority  int                    `json:"priority"`
	Timeout   time.Duration          `json:"timeout"`
	CreatedAt time.Time              `json:"created_at"`
}

// NodeInfo represents node information for partitioning
type NodeInfo struct {
	ID           string                 `json:"id"`
	Address      string                 `json:"address"`
	Capacity     *ResourceCapacity      `json:"capacity"`
	Usage        *ResourceUsage         `json:"usage"`
	GPUs         []GPUInfo              `json:"gpus"`
	Latency      time.Duration          `json:"latency"`
	Bandwidth    int64                  `json:"bandwidth"`
	Capabilities []string               `json:"capabilities"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// GPUInfo represents GPU information
type GPUInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Memory    int64  `json:"memory"`
	Compute   string `json:"compute"`
	Driver    string `json:"driver"`
	Available bool   `json:"available"`
}

// Helper functions for safe options access
func (pt *PartitionTask) GetNumCtx() int {
	if val, ok := pt.Options["num_ctx"]; ok {
		if intVal, ok := val.(int); ok {
			return intVal
		}
		if floatVal, ok := val.(float64); ok {
			return int(floatVal)
		}
	}
	return 2048 // default context length
}

func (pt *PartitionTask) GetBatchSize() int {
	if val, ok := pt.Options["batch_size"]; ok {
		if intVal, ok := val.(int); ok {
			return intVal
		}
		if floatVal, ok := val.(float64); ok {
			return int(floatVal)
		}
	}
	return 1 // default batch size
}

// ResourceCapacity represents node resource capacity
type ResourceCapacity struct {
	CPUCores         int64   `json:"cpu_cores"`
	MemoryBytes      int64   `json:"memory_bytes"`
	GPUCount         int     `json:"gpu_count"`
	GPUMemoryBytes   int64   `json:"gpu_memory_bytes"`
	NetworkBandwidth int64   `json:"network_bandwidth"`
	ComputeScore     float64 `json:"compute_score"`
}

// ResourceUsage represents node resource usage
type ResourceUsage struct {
	CPUUtilization     float64 `json:"cpu_utilization"`
	MemoryUtilization  float64 `json:"memory_utilization"`
	GPUUtilization     float64 `json:"gpu_utilization"`
	NetworkUtilization float64 `json:"network_utilization"`
	ActiveRequests     int     `json:"active_requests"`
}

// PartitionPlan represents a partitioning plan
type PartitionPlan struct {
	ID                  string                 `json:"id"`
	Strategy            string                 `json:"strategy"`
	Partitions          []*Partition           `json:"partitions"`
	Metadata            map[string]interface{} `json:"metadata"`
	EstimatedLatency    time.Duration          `json:"estimated_latency"`
	EstimatedThroughput float64                `json:"estimated_throughput"`
	OptimizationScore   float64                `json:"optimization_score"`
	CreatedAt           time.Time              `json:"created_at"`
}

// Partition represents a single partition
type Partition struct {
	ID               string                 `json:"id"`
	NodeID           string                 `json:"node_id"`
	Type             PartitionType          `json:"type"`
	Data             interface{}            `json:"data"`
	Dependencies     []string               `json:"dependencies"`
	Metadata         map[string]interface{} `json:"metadata"`
	EstimatedLatency time.Duration          `json:"estimated_latency"`
	EstimatedMemory  int64                  `json:"estimated_memory"`
}

// PartitionType represents the type of partition
type PartitionType string

const (
	PartitionTypeLayer     PartitionType = "layer"
	PartitionTypeData      PartitionType = "data"
	PartitionTypeTask      PartitionType = "task"
	PartitionTypeSequence  PartitionType = "sequence"
	PartitionTypeAttention PartitionType = "attention"
	PartitionTypeEmbedding PartitionType = "embedding"
)

// StrategyMetrics represents metrics for a partitioning strategy
type StrategyMetrics struct {
	TotalPartitions      int64         `json:"total_partitions"`
	SuccessfulPartitions int64         `json:"successful_partitions"`
	FailedPartitions     int64         `json:"failed_partitions"`
	AverageLatency       time.Duration `json:"average_latency"`
	AverageThroughput    float64       `json:"average_throughput"`
	LastUsed             time.Time     `json:"last_used"`
}

// PartitionOptimizer optimizes partitioning decisions
type PartitionOptimizer struct {
	history             []*PartitionResult
	learningRate        float64
	optimizationWeights map[string]float64
}

// PartitionResult represents the result of a partitioning decision
type PartitionResult struct {
	Plan             *PartitionPlan `json:"plan"`
	ActualLatency    time.Duration  `json:"actual_latency"`
	ActualThroughput float64        `json:"actual_throughput"`
	Success          bool           `json:"success"`
	Timestamp        time.Time      `json:"timestamp"`
}

// WorkloadAnalyzer analyzes workloads to select optimal partitioning strategies
type WorkloadAnalyzer struct {
	profiles map[string]*WorkloadProfile
}

// WorkloadProfile represents a workload profile
type WorkloadProfile struct {
	ModelSize          int64   `json:"model_size"`
	ContextLength      int     `json:"context_length"`
	BatchSize          int     `json:"batch_size"`
	Complexity         float64 `json:"complexity"`
	MemoryRequirement  int64   `json:"memory_requirement"`
	ComputeRequirement float64 `json:"compute_requirement"`
	Parallelizability  float64 `json:"parallelizability"`
}

// NewPartitionManager creates a new partition manager
func NewPartitionManager(config *Config) *PartitionManager {
	pm := &PartitionManager{
		config:     config,
		strategies: make(map[string]PartitionStrategy),
		optimizer: &PartitionOptimizer{
			history:      make([]*PartitionResult, 0),
			learningRate: 0.1,
			optimizationWeights: map[string]float64{
				"latency":    0.4,
				"throughput": 0.3,
				"memory":     0.2,
				"bandwidth":  0.1,
			},
		},
		analyzer: &WorkloadAnalyzer{
			profiles: make(map[string]*WorkloadProfile),
		},
	}

	// Register strategies
	pm.RegisterStrategy(NewLayerwiseStrategy())
	pm.RegisterStrategy(NewDataSplitStrategy())
	pm.RegisterStrategy(NewTaskParallelismStrategy())
	pm.RegisterStrategy(NewSequenceParallelismStrategy())
	pm.RegisterStrategy(NewAttentionParallelismStrategy())

	// Register advanced strategies (will be implemented in enhanced_partitioning.go)

	return pm
}

// RegisterStrategy registers a partitioning strategy
func (pm *PartitionManager) RegisterStrategy(strategy PartitionStrategy) {
	pm.strategies[strategy.GetName()] = strategy
}

// SelectStrategy selects the best partitioning strategy for a task
func (pm *PartitionManager) SelectStrategy(task interface{}, model *types.OllamaModel, opts map[string]interface{}) (string, error) {
	// Create partition task
	partitionTask := &PartitionTask{
		ID:        fmt.Sprintf("task_%d", time.Now().UnixNano()),
		Type:      "inference",
		Model:     model,
		Options:   opts,
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
	}

	// Analyze workload
	profile := pm.analyzer.AnalyzeWorkload(partitionTask)

	// Select strategy based on workload profile
	if profile.ModelSize > 10*1024*1024*1024 { // 10GB+
		if profile.Parallelizability > 0.8 {
			return "layerwise", nil
		}
		return "data_split", nil
	}

	if profile.ContextLength > 2048 {
		return "sequence_parallel", nil
	}

	if profile.BatchSize > 1 {
		return "data_split", nil
	}

	// Default strategy
	return pm.config.DefaultStrategy, nil
}

// Partition partitions a task using the specified strategy
func (pm *PartitionManager) Partition(ctx context.Context, task *PartitionTask, strategyName string) (*PartitionPlan, error) {
	strategy, exists := pm.strategies[strategyName]
	if !exists {
		return nil, fmt.Errorf("strategy not found: %s", strategyName)
	}

	if !strategy.CanHandle(task) {
		return nil, fmt.Errorf("strategy %s cannot handle task", strategyName)
	}

	// Execute partitioning
	plan, err := strategy.Partition(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("partitioning failed: %v", err)
	}

	// Optimize plan
	optimizedPlan := pm.optimizer.OptimizePlan(plan)

	return optimizedPlan, nil
}

// GetAvailableStrategies returns all available partitioning strategies
func (pm *PartitionManager) GetAvailableStrategies() []string {
	strategies := make([]string, 0, len(pm.strategies))
	for name := range pm.strategies {
		strategies = append(strategies, name)
	}
	return strategies
}

// GetStrategyMetrics returns metrics for a specific strategy
func (pm *PartitionManager) GetStrategyMetrics(strategyName string) (*StrategyMetrics, error) {
	strategy, exists := pm.strategies[strategyName]
	if !exists {
		return nil, fmt.Errorf("strategy not found: %s", strategyName)
	}

	return strategy.GetMetrics(), nil
}

// WorkloadAnalyzer methods

// AnalyzeWorkload analyzes a workload and returns its profile
func (wa *WorkloadAnalyzer) AnalyzeWorkload(task *PartitionTask) *WorkloadProfile {
	profile := &WorkloadProfile{
		ModelSize:          wa.estimateModelSize(task),
		ContextLength:      task.GetNumCtx(),
		BatchSize:          wa.estimateBatchSize(task),
		Complexity:         wa.estimateComplexity(task),
		MemoryRequirement:  wa.estimateMemoryRequirement(task),
		ComputeRequirement: wa.estimateComputeRequirement(task),
		Parallelizability:  wa.estimateParallelizability(task),
	}

	// Cache profile
	profileKey := fmt.Sprintf("%s_%d_%d", task.Model.Name, task.GetNumCtx(), wa.estimateBatchSize(task))
	wa.profiles[profileKey] = profile

	return profile
}

// estimateModelSize estimates the size of a model
func (wa *WorkloadAnalyzer) estimateModelSize(task *PartitionTask) int64 {
	// Use model size from metadata if available
	if task.Model != nil && task.Model.Size > 0 {
		return task.Model.Size
	}
	// Fallback estimation based on model name patterns
	return 4 * 1024 * 1024 * 1024 // 4GB default
}

// estimateBatchSize estimates the batch size for a task
func (wa *WorkloadAnalyzer) estimateBatchSize(task *PartitionTask) int {
	// Extract batch size from task metadata or options
	if batchSize, exists := task.Metadata["batch_size"]; exists {
		if bs, ok := batchSize.(int); ok {
			return bs
		}
	}
	return 1 // Default batch size
}

// estimateComplexity estimates the computational complexity
func (wa *WorkloadAnalyzer) estimateComplexity(task *PartitionTask) float64 {
	// Base complexity on model size and context length
	modelSize := float64(wa.estimateModelSize(task))
	contextLength := float64(task.GetNumCtx())

	// Complexity grows with model size and context length
	complexity := math.Log(modelSize) * math.Log(contextLength)

	// Normalize to 0-1 range
	return math.Min(complexity/100.0, 1.0)
}

// estimateMemoryRequirement estimates memory requirements
func (wa *WorkloadAnalyzer) estimateMemoryRequirement(task *PartitionTask) int64 {
	modelSize := wa.estimateModelSize(task)
	contextLength := int64(task.GetNumCtx())

	// Memory requirement = model size + context memory + overhead
	contextMemory := contextLength * 4 * 1024 // 4KB per context token
	overhead := modelSize / 10                // 10% overhead

	return modelSize + contextMemory + overhead
}

// estimateComputeRequirement estimates compute requirements
func (wa *WorkloadAnalyzer) estimateComputeRequirement(task *PartitionTask) float64 {
	modelSize := float64(wa.estimateModelSize(task))
	contextLength := float64(task.GetNumCtx())

	// Compute requirement based on model parameters and context
	// This is a simplified estimation
	params := modelSize / 4                  // Assume 4 bytes per parameter
	operations := params * contextLength * 2 // Forward and backward pass

	// Return in GFLOPS
	return operations / 1e9
}

// estimateParallelizability estimates how parallelizable a task is
func (wa *WorkloadAnalyzer) estimateParallelizability(task *PartitionTask) float64 {
	// Factors that affect parallelizability:
	// 1. Model architecture (transformers are more parallelizable)
	// 2. Context length (longer contexts are more parallelizable)
	// 3. Batch size (larger batches are more parallelizable)

	contextLength := float64(task.GetNumCtx())
	batchSize := float64(wa.estimateBatchSize(task))

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

// isTransformerModel checks if a model is a transformer-based model
func isTransformerModel(model *types.OllamaModel) bool {
	// Check model name patterns for transformer architectures
	modelName := strings.ToLower(model.Name)
	transformerPatterns := []string{"llama", "mistral", "gpt", "bert", "t5", "transformer"}

	for _, pattern := range transformerPatterns {
		if strings.Contains(modelName, pattern) {
			return true
		}
	}
	return false
}

// PartitionOptimizer methods

// OptimizePlan optimizes a partition plan
func (po *PartitionOptimizer) OptimizePlan(plan *PartitionPlan) *PartitionPlan {
	// Calculate optimization score
	plan.OptimizationScore = po.calculateOptimizationScore(plan)

	// Apply optimizations
	optimizedPlan := po.applyOptimizations(plan)

	// Recalculate score
	optimizedPlan.OptimizationScore = po.calculateOptimizationScore(optimizedPlan)

	return optimizedPlan
}

// calculateOptimizationScore calculates an optimization score for a plan
func (po *PartitionOptimizer) calculateOptimizationScore(plan *PartitionPlan) float64 {
	// Weighted score based on latency, throughput, and resource usage
	latencyScore := 1.0 - (float64(plan.EstimatedLatency) / float64(time.Second))
	throughputScore := plan.EstimatedThroughput / 1000.0 // Normalize to 1000 ops/sec
	memoryScore := 1.0                                   // Placeholder for memory efficiency
	bandwidthScore := 1.0                                // Placeholder for bandwidth efficiency

	// Weighted combination
	score := po.optimizationWeights["latency"]*latencyScore +
		po.optimizationWeights["throughput"]*throughputScore +
		po.optimizationWeights["memory"]*memoryScore +
		po.optimizationWeights["bandwidth"]*bandwidthScore

	return math.Max(0.0, math.Min(1.0, score))
}

// applyOptimizations applies optimizations to a partition plan
func (po *PartitionOptimizer) applyOptimizations(plan *PartitionPlan) *PartitionPlan {
	// Create optimized copy
	optimized := &PartitionPlan{
		ID:                  plan.ID + "_optimized",
		Strategy:            plan.Strategy,
		Partitions:          make([]*Partition, len(plan.Partitions)),
		Metadata:            make(map[string]interface{}),
		EstimatedLatency:    plan.EstimatedLatency,
		EstimatedThroughput: plan.EstimatedThroughput,
		CreatedAt:           time.Now(),
	}

	// Copy partitions
	copy(optimized.Partitions, plan.Partitions)

	// Copy metadata
	for k, v := range plan.Metadata {
		optimized.Metadata[k] = v
	}

	// Apply specific optimizations
	optimized = po.optimizePartitionPlacement(optimized)
	optimized = po.optimizeResourceUsage(optimized)
	optimized = po.optimizeCommunication(optimized)

	return optimized
}

// optimizePartitionPlacement optimizes the placement of partitions
func (po *PartitionOptimizer) optimizePartitionPlacement(plan *PartitionPlan) *PartitionPlan {
	// Analyze partition dependencies and optimize placement
	// This is a simplified implementation
	for i, partition := range plan.Partitions {
		// Add optimization metadata
		if partition.Metadata == nil {
			partition.Metadata = make(map[string]interface{})
		}
		partition.Metadata["optimized_placement"] = true
		partition.Metadata["optimization_step"] = i
	}

	return plan
}

// optimizeResourceUsage optimizes resource usage across partitions
func (po *PartitionOptimizer) optimizeResourceUsage(plan *PartitionPlan) *PartitionPlan {
	// Balance resource usage across partitions
	// This is a simplified implementation
	for _, partition := range plan.Partitions {
		if partition.Metadata == nil {
			partition.Metadata = make(map[string]interface{})
		}
		partition.Metadata["resource_optimized"] = true
	}

	return plan
}

// optimizeCommunication optimizes communication between partitions
func (po *PartitionOptimizer) optimizeCommunication(plan *PartitionPlan) *PartitionPlan {
	// Minimize communication overhead between partitions
	// This is a simplified implementation
	for _, partition := range plan.Partitions {
		if partition.Metadata == nil {
			partition.Metadata = make(map[string]interface{})
		}
		partition.Metadata["communication_optimized"] = true
	}

	return plan
}

// RecordResult records the result of a partitioning decision for learning
func (po *PartitionOptimizer) RecordResult(result *PartitionResult) {
	po.history = append(po.history, result)

	// Keep only last 1000 results
	if len(po.history) > 1000 {
		po.history = po.history[len(po.history)-1000:]
	}

	// Learn from result
	po.learnFromResult(result)
}

// learnFromResult learns from a partitioning result
func (po *PartitionOptimizer) learnFromResult(result *PartitionResult) {
	// Adjust optimization weights based on result
	if result.Success {
		// Positive reinforcement
		if result.ActualLatency < result.Plan.EstimatedLatency {
			po.optimizationWeights["latency"] *= (1.0 + po.learningRate)
		}
		if result.ActualThroughput > result.Plan.EstimatedThroughput {
			po.optimizationWeights["throughput"] *= (1.0 + po.learningRate)
		}
	} else {
		// Negative reinforcement
		po.optimizationWeights["latency"] *= (1.0 - po.learningRate)
		po.optimizationWeights["throughput"] *= (1.0 - po.learningRate)
	}

	// Normalize weights
	po.normalizeWeights()
}

// normalizeWeights normalizes optimization weights to sum to 1.0
func (po *PartitionOptimizer) normalizeWeights() {
	sum := 0.0
	for _, weight := range po.optimizationWeights {
		sum += weight
	}

	if sum > 0 {
		for key, weight := range po.optimizationWeights {
			po.optimizationWeights[key] = weight / sum
		}
	}
}

// GetOptimizationHistory returns the optimization history
func (po *PartitionOptimizer) GetOptimizationHistory() []*PartitionResult {
	return po.history
}

// GetOptimizationWeights returns the current optimization weights
func (po *PartitionOptimizer) GetOptimizationWeights() map[string]float64 {
	weights := make(map[string]float64)
	for k, v := range po.optimizationWeights {
		weights[k] = v
	}
	return weights
}

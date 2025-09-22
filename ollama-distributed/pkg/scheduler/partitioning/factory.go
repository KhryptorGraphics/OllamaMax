package partitioning

// Factory wrapper functions that return concrete implemented strategies
// These functions provide test-expected names while delegating to actual implementations

// NewLayerwiseStrategy returns a layer partitioning strategy
func NewLayerwiseStrategy() PartitionStrategy {
	return NewLayerPartitionStrategy(NewModelAnalyzer(), NewMemoryOptimizer())
}

// NewTensorParallelismStrategy returns a tensor partitioning strategy
func NewTensorParallelismStrategy() PartitionStrategy {
	return NewTensorPartitionStrategy(NewModelAnalyzer(), NewMemoryOptimizer())
}

// NewPipelineParallelismStrategy returns a pipeline partitioning strategy
func NewPipelineParallelismStrategy() PartitionStrategy {
	return NewPipelinePartitionStrategy(NewModelAnalyzer(), NewMemoryOptimizer())
}

// NewHybridParallelismStrategy returns a hybrid partitioning strategy
func NewHybridParallelismStrategy() PartitionStrategy {
	return NewHybridPartitionStrategy(NewModelAnalyzer(), NewMemoryOptimizer())
}

// NewAdaptivePartitioningStrategy returns an adaptive partitioning strategy
func NewAdaptivePartitioningStrategy() PartitionStrategy {
	return NewAdaptivePartitionStrategy(NewModelAnalyzer(), NewMemoryOptimizer())
}

// Additional factory functions for legacy compatibility
// These are required by stub.go for compatibility

// NewDataSplitStrategy returns a data split strategy (uses tensor parallelism internally)
func NewDataSplitStrategy() PartitionStrategy {
	return NewTensorPartitionStrategy(NewModelAnalyzer(), NewMemoryOptimizer())
}

// NewTaskParallelismStrategy returns a task parallelism strategy (uses hybrid internally)
func NewTaskParallelismStrategy() PartitionStrategy {
	return NewHybridPartitionStrategy(NewModelAnalyzer(), NewMemoryOptimizer())
}

// NewSequenceParallelismStrategy returns a sequence parallelism strategy (uses pipeline internally)
func NewSequenceParallelismStrategy() PartitionStrategy {
	return NewPipelinePartitionStrategy(NewModelAnalyzer(), NewMemoryOptimizer())
}

// NewAttentionParallelismStrategy returns an attention parallelism strategy (uses tensor internally)
func NewAttentionParallelismStrategy() PartitionStrategy {
	return NewTensorPartitionStrategy(NewModelAnalyzer(), NewMemoryOptimizer())
}

// GetAvailableStrategies returns a list of all available partitioning strategies
func GetAvailableStrategies() []string {
	// Return all known strategy aliases
	strategies := make([]string, 0, len(StrategyAliases))
	for alias := range StrategyAliases {
		strategies = append(strategies, alias)
	}
	return strategies
}

// GetStrategyDescription returns a human-readable description of a strategy
func GetStrategyDescription(strategyName string) string {
	descriptions := map[string]string{
		"layer":    "Distributes model layers across nodes sequentially, optimal for memory-constrained environments",
		"tensor":   "Splits tensors across nodes with all-reduce communication, best for high-bandwidth networks",
		"pipeline": "Creates pipeline stages with micro-batching, ideal for balanced compute workloads",
		"hybrid":   "Combines multiple strategies with intelligent coordination for complex scenarios",
		"adaptive": "Learns optimal strategy selection based on historical performance and current conditions",
		// Long-form names
		"layerwise": "Distributes model layers across nodes sequentially, optimal for memory-constrained environments",
		"tensor_parallelism": "Splits tensors across nodes with all-reduce communication, best for high-bandwidth networks",
		"pipeline_parallelism": "Creates pipeline stages with micro-batching, ideal for balanced compute workloads",
		"hybrid_parallelism": "Combines multiple strategies with intelligent coordination for complex scenarios",
		"adaptive_partitioning": "Learns optimal strategy selection based on historical performance and current conditions",
	}
	
	if desc, exists := descriptions[strategyName]; exists {
		return desc
	}
	return "Unknown strategy"
}
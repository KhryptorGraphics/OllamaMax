package partitioning

import (
	"fmt"
	"math"
	"sort"
)

// MemoryOptimizer provides memory optimization utilities for partitioning strategies
type MemoryOptimizer struct {
	config *MemoryOptimizerConfig
}

// MemoryOptimizerConfig contains configuration for memory optimization
type MemoryOptimizerConfig struct {
	SafetyBuffer           float64 `json:"safety_buffer"`            // Safety buffer percentage (e.g., 0.1 for 10%)
	FragmentationThreshold float64 `json:"fragmentation_threshold"`  // Threshold for memory fragmentation
	AllowOvercommit        bool    `json:"allow_overcommit"`         // Allow memory overcommit
	OvercommitRatio        float64 `json:"overcommit_ratio"`         // Overcommit ratio if allowed
	GradientCheckpointing  bool    `json:"gradient_checkpointing"`   // Enable gradient checkpointing
	ActivationCheckpointing bool   `json:"activation_checkpointing"` // Enable activation checkpointing
	SwapToStorage          bool    `json:"swap_to_storage"`          // Allow swapping to storage
	CompressionRatio       float64 `json:"compression_ratio"`        // Expected compression ratio
}

// MemoryAllocation represents a memory allocation plan
type MemoryAllocation struct {
	NodeID         string              `json:"node_id"`
	TotalRequired  int64               `json:"total_required"`  // Total memory required
	Available      int64               `json:"available"`       // Available memory
	Allocations    []MemoryBlock       `json:"allocations"`     // Specific memory allocations
	Utilization    float64             `json:"utilization"`     // Memory utilization percentage
	Fragmentation  float64             `json:"fragmentation"`   // Memory fragmentation score
	OptimizationHints []OptimizationHint `json:"optimization_hints"` // Optimization suggestions
}

// MemoryBlock represents a block of allocated memory
type MemoryBlock struct {
	ID          string      `json:"id"`
	Purpose     string      `json:"purpose"`      // model_weights, activations, gradients, etc.
	Size        int64       `json:"size"`         // Size in bytes
	Priority    MemoryPriority `json:"priority"`  // Memory priority
	Compressible bool       `json:"compressible"` // Can this block be compressed
	Swappable   bool        `json:"swappable"`    // Can this block be swapped to storage
	Shareable   bool        `json:"shareable"`    // Can this block be shared between processes
}

// MemoryPriority defines memory allocation priorities
type MemoryPriority int

const (
	PriorityLow MemoryPriority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

// OptimizationHint provides memory optimization suggestions
type OptimizationHint struct {
	Type        HintType `json:"type"`
	Description string   `json:"description"`
	Impact      string   `json:"impact"`       // Expected impact of applying this hint
	Effort      string   `json:"effort"`       // Effort required to implement
	Savings     int64    `json:"savings"`      // Expected memory savings in bytes
}

// HintType defines types of optimization hints
type HintType string

const (
	HintCheckpointing    HintType = "checkpointing"
	HintCompression      HintType = "compression"
	HintSwapping         HintType = "swapping"
	HintReallocation     HintType = "reallocation"
	HintFragmentation    HintType = "fragmentation"
	HintOvercommit       HintType = "overcommit"
)

// MemoryPressureAnalysis contains memory pressure analysis results
type MemoryPressureAnalysis struct {
	OverallPressure    float64           `json:"overall_pressure"`     // 0.0-1.0
	NodePressures      map[string]float64 `json:"node_pressures"`      // Per-node pressure
	CriticalNodes      []string          `json:"critical_nodes"`       // Nodes under critical pressure
	BottleneckType     BottleneckType    `json:"bottleneck_type"`      // Type of memory bottleneck
	Recommendations    []string          `json:"recommendations"`      // Recommended actions
	PredictedFailures  []PredictedFailure `json:"predicted_failures"`  // Predicted memory failures
}

// BottleneckType defines types of memory bottlenecks
type BottleneckType string

const (
	BottleneckCapacity      BottleneckType = "capacity"      // Insufficient total memory
	BottleneckFragmentation BottleneckType = "fragmentation" // Memory fragmentation issues
	BottleneckBandwidth     BottleneckType = "bandwidth"     // Memory bandwidth limitations
	BottleneckLatency       BottleneckType = "latency"       // Memory access latency
	BottleneckNone          BottleneckType = "none"          // No significant bottleneck
)

// PredictedFailure represents a predicted memory allocation failure
type PredictedFailure struct {
	NodeID      string  `json:"node_id"`
	Probability float64 `json:"probability"` // 0.0-1.0
	TimeToFailure string `json:"time_to_failure"`
	Cause       string  `json:"cause"`
	Mitigation  string  `json:"mitigation"`
}

// NewMemoryOptimizer creates a new memory optimizer
func NewMemoryOptimizer() *MemoryOptimizer {
	return &MemoryOptimizer{
		config: &MemoryOptimizerConfig{
			SafetyBuffer:            0.15, // 15% safety buffer
			FragmentationThreshold:  0.3,  // 30% fragmentation threshold
			AllowOvercommit:         false,
			OvercommitRatio:         1.2,  // 20% overcommit if allowed
			GradientCheckpointing:   true,
			ActivationCheckpointing: true,
			SwapToStorage:          false,
			CompressionRatio:       0.7,   // 30% compression expected
		},
	}
}

// CalculateMemoryRequirements calculates memory requirements for different partition types
func (mo *MemoryOptimizer) CalculateMemoryRequirements(partitionType string, modelAnalysis *ModelAnalysis, nodeCount int) (int64, error) {
	// Default to inference mode for backward compatibility
	return mo.CalculateMemoryRequirementsWithMode(partitionType, modelAnalysis, nodeCount, true)
}

// CalculateMemoryRequirementsWithMode calculates memory requirements with inference/training mode
func (mo *MemoryOptimizer) CalculateMemoryRequirementsWithMode(partitionType string, modelAnalysis *ModelAnalysis, nodeCount int, inferenceMode bool) (int64, error) {
	if modelAnalysis == nil || modelAnalysis.MemoryReqs == nil {
		return 0, fmt.Errorf("model analysis or memory requirements not available")
	}

	baseMemory := modelAnalysis.MemoryReqs.TotalRequired

	switch partitionType {
	case "layer_parallelism":
		return mo.calculateLayerParallelismMemory(baseMemory, modelAnalysis.LayerInfo, nodeCount, inferenceMode)
	case "tensor_parallelism":
		return mo.calculateTensorParallelismMemory(baseMemory, modelAnalysis.TensorInfo, nodeCount, inferenceMode)
	case "pipeline_parallelism":
		return mo.calculatePipelineParallelismMemory(baseMemory, modelAnalysis.StageInfo, nodeCount, inferenceMode)
	case "hybrid_parallelism":
		return mo.calculateHybridParallelismMemory(baseMemory, modelAnalysis, nodeCount, inferenceMode)
	default:
		return baseMemory, nil
	}
}

// calculateLayerParallelismMemory calculates memory for layer parallelism
func (mo *MemoryOptimizer) calculateLayerParallelismMemory(baseMemory int64, layerInfo *LayerAnalysis, nodeCount int, inferenceMode bool) (int64, error) {
	if layerInfo == nil {
		return baseMemory / int64(nodeCount), nil
	}

	// Each node gets a subset of layers
	memoryPerNode := baseMemory / int64(nodeCount)

	// Add activation memory for forward/backward passes
	var activationMemory int64
	if inferenceMode {
		activationMemory = memoryPerNode / 8 // Inference: less activation memory needed
	} else {
		activationMemory = memoryPerNode / 4 // Training: 25% of model memory for activations
	}

	// Add gradient memory if training
	var gradientMemory int64
	if !inferenceMode {
		gradientMemory = memoryPerNode // Same size as model weights for gradients
	}

	// Add optimizer memory if training
	var optimizerMemory int64
	if !inferenceMode {
		optimizerMemory = memoryPerNode * 2 // Adam optimizer needs ~2x model weights
	}

	totalMemoryPerNode := memoryPerNode + activationMemory + gradientMemory + optimizerMemory
	
	// Apply optimizations
	if mo.config.GradientCheckpointing {
		totalMemoryPerNode = int64(float64(totalMemoryPerNode) * 0.8) // 20% savings
	}
	
	// Add safety buffer
	totalMemoryPerNode = int64(float64(totalMemoryPerNode) * (1.0 + mo.config.SafetyBuffer))
	
	return totalMemoryPerNode, nil
}

// calculateTensorParallelismMemory calculates memory for tensor parallelism
func (mo *MemoryOptimizer) calculateTensorParallelismMemory(baseMemory int64, tensorInfo *TensorAnalysis, nodeCount int, inferenceMode bool) (int64, error) {
	// For tensor parallelism, each node has all layers but split tensors
	// Memory is reduced mainly by splitting large tensors

	reductionFactor := 1.0 / float64(nodeCount)
	tensorMemory := int64(float64(baseMemory) * reductionFactor)

	// Activations are also split
	var activationMemory int64
	if inferenceMode {
		activationMemory = tensorMemory / 5 // Inference: less activation memory
	} else {
		activationMemory = tensorMemory / 3 // Training: more activation memory
	}

	// Gradients only for training
	var gradientMemory int64
	if !inferenceMode {
		gradientMemory = tensorMemory
	}

	// Optimizer states only for training
	var optimizerMemory int64
	if !inferenceMode {
		optimizerMemory = tensorMemory * 2 // Adam optimizer
	}
	
	// Communication buffers needed for all-reduce operations
	communicationBuffers := tensorMemory / 10 // 10% for communication
	
	totalMemoryPerNode := tensorMemory + activationMemory + gradientMemory + optimizerMemory + communicationBuffers
	
	// Apply safety buffer
	totalMemoryPerNode = int64(float64(totalMemoryPerNode) * (1.0 + mo.config.SafetyBuffer))
	
	return totalMemoryPerNode, nil
}

// calculatePipelineParallelismMemory calculates memory for pipeline parallelism
func (mo *MemoryOptimizer) calculatePipelineParallelismMemory(baseMemory int64, stageInfo *PipelineStageAnalysis, nodeCount int, inferenceMode bool) (int64, error) {
	// Each node gets a pipeline stage
	stageMemory := baseMemory / int64(nodeCount)

	// Pipeline needs activation buffers for multiple micro-batches
	microBatches := 4 // Typical micro-batch count
	var activationBuffers int64
	if inferenceMode {
		activationBuffers = stageMemory / 16 * int64(microBatches) // Inference: smaller buffers
	} else {
		activationBuffers = stageMemory / 8 * int64(microBatches) // Training: larger buffers
	}

	// Gradient accumulation buffers only for training
	var gradientBuffers int64
	if !inferenceMode {
		gradientBuffers = stageMemory
	}

	// Optimizer states only for training
	var optimizerBuffers int64
	if !inferenceMode {
		optimizerBuffers = stageMemory * 2 // Adam optimizer
	}
	
	totalMemoryPerNode := stageMemory + activationBuffers + gradientBuffers + optimizerBuffers
	
	// Apply activation checkpointing optimization
	if mo.config.ActivationCheckpointing {
		totalMemoryPerNode = int64(float64(totalMemoryPerNode) * 0.7) // 30% savings
	}
	
	// Add safety buffer
	totalMemoryPerNode = int64(float64(totalMemoryPerNode) * (1.0 + mo.config.SafetyBuffer))
	
	return totalMemoryPerNode, nil
}

// calculateHybridParallelismMemory calculates memory for hybrid parallelism
func (mo *MemoryOptimizer) calculateHybridParallelismMemory(baseMemory int64, analysis *ModelAnalysis, nodeCount int, inferenceMode bool) (int64, error) {
	// Hybrid is more complex - combine different strategies
	// Assume pipeline + tensor parallelism combination
	
	pipelineStages := nodeCount / 2 // Half the nodes for pipeline stages
	if pipelineStages < 2 {
		pipelineStages = 2
	}
	tensorParallelNodes := nodeCount / pipelineStages
	if tensorParallelNodes < 1 {
		tensorParallelNodes = 1
	}
	
	// Calculate per-stage memory
	stageMemory := baseMemory / int64(pipelineStages)
	
	// Apply tensor parallelism within each stage
	tensorReduction := 1.0 / float64(tensorParallelNodes)
	memoryPerNode := int64(float64(stageMemory) * tensorReduction)
	
	// Add buffers and overheads
	activationBuffers := memoryPerNode / 4
	gradientBuffers := memoryPerNode / 2
	communicationBuffers := memoryPerNode / 8 // For both pipeline and tensor communication
	
	totalMemoryPerNode := memoryPerNode + activationBuffers + gradientBuffers + communicationBuffers
	
	// Apply all optimizations
	if mo.config.GradientCheckpointing {
		totalMemoryPerNode = int64(float64(totalMemoryPerNode) * 0.85) // 15% savings
	}
	if mo.config.ActivationCheckpointing {
		totalMemoryPerNode = int64(float64(totalMemoryPerNode) * 0.9) // 10% additional savings
	}
	
	// Add safety buffer
	totalMemoryPerNode = int64(float64(totalMemoryPerNode) * (1.0 + mo.config.SafetyBuffer))
	
	return totalMemoryPerNode, nil
}

// ValidateNodeCapacity validates if nodes have sufficient memory capacity
func (mo *MemoryOptimizer) ValidateNodeCapacity(nodes []*NodeInfo, requiredMemoryPerNode int64) (*MemoryPressureAnalysis, error) {
	analysis := &MemoryPressureAnalysis{
		NodePressures:     make(map[string]float64),
		CriticalNodes:     []string{},
		Recommendations:   []string{},
		PredictedFailures: []PredictedFailure{},
	}
	
	totalPressure := 0.0
	criticalCount := 0
	
	for _, node := range nodes {
		if node.Capabilities.Memory == nil {
			continue
		}
		
		availableMemory := node.Capabilities.Memory.AvailableBytes
		pressure := float64(requiredMemoryPerNode) / float64(availableMemory)
		
		analysis.NodePressures[node.ID] = pressure
		totalPressure += pressure
		
		// Check for critical pressure (>90% utilization)
		if pressure > 0.9 {
			analysis.CriticalNodes = append(analysis.CriticalNodes, node.ID)
			criticalCount++
			
			// Predict failure
			failure := PredictedFailure{
				NodeID:        node.ID,
				Probability:   math.Min(1.0, pressure-0.5), // Higher pressure = higher probability
				TimeToFailure: "immediate",
				Cause:         "insufficient_memory",
				Mitigation:    "increase_node_memory_or_reduce_allocation",
			}
			analysis.PredictedFailures = append(analysis.PredictedFailures, failure)
		}
	}
	
	if len(nodes) > 0 {
		analysis.OverallPressure = totalPressure / float64(len(nodes))
	}
	
	// Determine bottleneck type
	if analysis.OverallPressure > 0.9 {
		analysis.BottleneckType = BottleneckCapacity
	} else if criticalCount > len(nodes)/2 {
		analysis.BottleneckType = BottleneckFragmentation
	} else {
		analysis.BottleneckType = BottleneckNone
	}
	
	// Generate recommendations
	analysis.Recommendations = mo.generateMemoryRecommendations(analysis)
	
	return analysis, nil
}

// generateMemoryRecommendations generates memory optimization recommendations
func (mo *MemoryOptimizer) generateMemoryRecommendations(analysis *MemoryPressureAnalysis) []string {
	var recommendations []string
	
	if analysis.OverallPressure > 0.8 {
		recommendations = append(recommendations, "Consider enabling gradient checkpointing to reduce memory usage")
		recommendations = append(recommendations, "Enable activation checkpointing to save memory")
		
		if mo.config.AllowOvercommit {
			recommendations = append(recommendations, "Consider using memory overcommit with caution")
		}
		
		if mo.config.SwapToStorage {
			recommendations = append(recommendations, "Enable swapping less critical data to storage")
		}
	}
	
	if len(analysis.CriticalNodes) > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Critical memory pressure on %d nodes - consider rebalancing", len(analysis.CriticalNodes)))
		recommendations = append(recommendations, "Add more nodes or upgrade memory capacity on critical nodes")
	}
	
	if analysis.BottleneckType == BottleneckFragmentation {
		recommendations = append(recommendations, "Memory fragmentation detected - consider memory compaction")
		recommendations = append(recommendations, "Use larger, contiguous memory allocations where possible")
	}
	
	return recommendations
}

// OptimizeMemoryDistribution optimizes memory distribution across nodes
func (mo *MemoryOptimizer) OptimizeMemoryDistribution(nodes []*NodeInfo, memoryBlocks []MemoryBlock) ([]*MemoryAllocation, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes provided")
	}
	
	// Sort nodes by available memory (descending)
	sortedNodes := make([]*NodeInfo, len(nodes))
	copy(sortedNodes, nodes)
	sort.Slice(sortedNodes, func(i, j int) bool {
		return sortedNodes[i].Capabilities.Memory.AvailableBytes > sortedNodes[j].Capabilities.Memory.AvailableBytes
	})
	
	// Sort memory blocks by priority and size
	sortedBlocks := make([]MemoryBlock, len(memoryBlocks))
	copy(sortedBlocks, memoryBlocks)
	sort.Slice(sortedBlocks, func(i, j int) bool {
		if sortedBlocks[i].Priority != sortedBlocks[j].Priority {
			return sortedBlocks[i].Priority > sortedBlocks[j].Priority
		}
		return sortedBlocks[i].Size > sortedBlocks[j].Size
	})
	
	// Initialize allocations
	allocations := make([]*MemoryAllocation, len(sortedNodes))
	for i, node := range sortedNodes {
		allocations[i] = &MemoryAllocation{
			NodeID:      node.ID,
			Available:   node.Capabilities.Memory.AvailableBytes,
			Allocations: []MemoryBlock{},
		}
	}
	
	// Distribute blocks using best-fit decreasing algorithm
	for _, block := range sortedBlocks {
		bestNodeIndex := mo.findBestFitNode(allocations, block)
		if bestNodeIndex >= 0 {
			allocations[bestNodeIndex].Allocations = append(allocations[bestNodeIndex].Allocations, block)
			allocations[bestNodeIndex].TotalRequired += block.Size
		} else {
			// Cannot fit block - apply optimizations
			optimizedBlock := mo.optimizeMemoryBlock(block)
			bestNodeIndex = mo.findBestFitNode(allocations, optimizedBlock)
			if bestNodeIndex >= 0 {
				allocations[bestNodeIndex].Allocations = append(allocations[bestNodeIndex].Allocations, optimizedBlock)
				allocations[bestNodeIndex].TotalRequired += optimizedBlock.Size
			} else {
				return nil, fmt.Errorf("cannot allocate memory block %s (size: %d)", block.ID, block.Size)
			}
		}
	}
	
	// Calculate utilization and fragmentation for each allocation
	for _, allocation := range allocations {
		allocation.Utilization = float64(allocation.TotalRequired) / float64(allocation.Available)
		allocation.Fragmentation = mo.calculateFragmentation(allocation)
		allocation.OptimizationHints = mo.generateOptimizationHints(allocation)
	}
	
	return allocations, nil
}

// findBestFitNode finds the best node to fit a memory block
func (mo *MemoryOptimizer) findBestFitNode(allocations []*MemoryAllocation, block MemoryBlock) int {
	bestIndex := -1
	bestWastage := int64(math.MaxInt64)
	
	for i, allocation := range allocations {
		remainingSpace := allocation.Available - allocation.TotalRequired
		if remainingSpace >= block.Size {
			wastage := remainingSpace - block.Size
			if wastage < bestWastage {
				bestWastage = wastage
				bestIndex = i
			}
		}
	}
	
	return bestIndex
}

// optimizeMemoryBlock applies optimizations to a memory block
func (mo *MemoryOptimizer) optimizeMemoryBlock(block MemoryBlock) MemoryBlock {
	optimizedBlock := block
	
	// Apply compression if possible
	if block.Compressible && mo.config.CompressionRatio > 0 {
		optimizedBlock.Size = int64(float64(block.Size) * mo.config.CompressionRatio)
	}
	
	// Other optimizations can be added here
	
	return optimizedBlock
}

// calculateFragmentation calculates memory fragmentation for an allocation
func (mo *MemoryOptimizer) calculateFragmentation(allocation *MemoryAllocation) float64 {
	if len(allocation.Allocations) <= 1 {
		return 0.0
	}
	
	// Simple fragmentation calculation based on number of blocks
	// More blocks = higher fragmentation
	totalBlocks := len(allocation.Allocations)
	avgBlockSize := allocation.TotalRequired / int64(totalBlocks)
	
	variance := 0.0
	for _, block := range allocation.Allocations {
		diff := float64(block.Size - avgBlockSize)
		variance += diff * diff
	}
	variance /= float64(totalBlocks)
	
	// Normalize fragmentation score (0.0 to 1.0)
	fragmentation := math.Sqrt(variance) / float64(avgBlockSize)
	return math.Min(1.0, fragmentation)
}

// generateOptimizationHints generates optimization hints for a memory allocation
func (mo *MemoryOptimizer) generateOptimizationHints(allocation *MemoryAllocation) []OptimizationHint {
	var hints []OptimizationHint
	
	// High utilization hints
	if allocation.Utilization > 0.9 {
		hints = append(hints, OptimizationHint{
			Type:        HintCheckpointing,
			Description: "Enable gradient checkpointing to reduce memory usage",
			Impact:      "10-20% memory reduction",
			Effort:      "Low",
			Savings:     int64(float64(allocation.TotalRequired) * 0.15),
		})
	}
	
	// Fragmentation hints
	if allocation.Fragmentation > mo.config.FragmentationThreshold {
		hints = append(hints, OptimizationHint{
			Type:        HintFragmentation,
			Description: "Memory fragmentation detected - consider consolidating allocations",
			Impact:      "5-10% efficiency improvement",
			Effort:      "Medium",
			Savings:     int64(float64(allocation.TotalRequired) * 0.08),
		})
	}
	
	// Compression hints
	compressibleSize := int64(0)
	for _, block := range allocation.Allocations {
		if block.Compressible {
			compressibleSize += block.Size
		}
	}
	
	if compressibleSize > 0 {
		expectedSavings := int64(float64(compressibleSize) * (1.0 - mo.config.CompressionRatio))
		hints = append(hints, OptimizationHint{
			Type:        HintCompression,
			Description: "Enable compression for applicable memory blocks",
			Impact:      fmt.Sprintf("Up to %d MB savings", expectedSavings/(1024*1024)),
			Effort:      "Low",
			Savings:     expectedSavings,
		})
	}
	
	// Swapping hints
	if mo.config.SwapToStorage && allocation.Utilization > 0.95 {
		swappableSize := int64(0)
		for _, block := range allocation.Allocations {
			if block.Swappable && block.Priority <= PriorityNormal {
				swappableSize += block.Size
			}
		}
		
		if swappableSize > 0 {
			hints = append(hints, OptimizationHint{
				Type:        HintSwapping,
				Description: "Consider swapping low-priority data to storage",
				Impact:      fmt.Sprintf("Free up %d MB in memory", swappableSize/(1024*1024)),
				Effort:      "Medium",
				Savings:     swappableSize,
			})
		}
	}
	
	return hints
}

// PredictMemoryUsage predicts memory usage patterns
func (mo *MemoryOptimizer) PredictMemoryUsage(allocations []*MemoryAllocation, timeHorizon string) (map[string]float64, error) {
	predictions := make(map[string]float64)
	
	for _, allocation := range allocations {
		// Simple prediction based on current utilization and growth patterns
		currentUtil := allocation.Utilization
		
		// Apply growth factors based on time horizon
		var growthFactor float64
		switch timeHorizon {
		case "1h":
			growthFactor = 1.05 // 5% growth
		case "24h":
			growthFactor = 1.15 // 15% growth
		case "7d":
			growthFactor = 1.3  // 30% growth
		default:
			growthFactor = 1.1  // 10% default growth
		}
		
		predictedUtil := math.Min(1.0, currentUtil*growthFactor)
		predictions[allocation.NodeID] = predictedUtil
	}
	
	return predictions, nil
}

// GetMemoryInsights provides insights about memory usage patterns
func (mo *MemoryOptimizer) GetMemoryInsights(allocations []*MemoryAllocation) map[string]interface{} {
	insights := make(map[string]interface{})
	
	totalMemory := int64(0)
	totalAllocated := int64(0)
	highUtilNodes := 0
	fragmentedNodes := 0
	
	for _, allocation := range allocations {
		totalMemory += allocation.Available
		totalAllocated += allocation.TotalRequired
		
		if allocation.Utilization > 0.8 {
			highUtilNodes++
		}
		
		if allocation.Fragmentation > mo.config.FragmentationThreshold {
			fragmentedNodes++
		}
	}
	
	overallUtilization := 0.0
	if totalMemory > 0 {
		overallUtilization = float64(totalAllocated) / float64(totalMemory)
	}
	
	insights["overall_utilization"] = overallUtilization
	insights["total_memory_gb"] = float64(totalMemory) / (1024 * 1024 * 1024)
	insights["total_allocated_gb"] = float64(totalAllocated) / (1024 * 1024 * 1024)
	insights["high_utilization_nodes"] = highUtilNodes
	insights["fragmented_nodes"] = fragmentedNodes
	insights["optimization_potential"] = mo.calculateOptimizationPotential(allocations)
	
	return insights
}

// calculateOptimizationPotential calculates the potential for memory optimization
func (mo *MemoryOptimizer) calculateOptimizationPotential(allocations []*MemoryAllocation) float64 {
	totalPotential := 0.0
	totalMemory := int64(0)
	
	for _, allocation := range allocations {
		totalMemory += allocation.TotalRequired
		
		for _, hint := range allocation.OptimizationHints {
			totalPotential += float64(hint.Savings)
		}
	}
	
	if totalMemory == 0 {
		return 0.0
	}
	
	return totalPotential / float64(totalMemory)
}
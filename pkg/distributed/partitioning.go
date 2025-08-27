package distributed

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// PartitionStrategy defines interface for different partitioning approaches
type PartitionStrategy interface {
	Partition(ctx context.Context, request *InferenceRequest, nodes []NodeInfo) (*PartitionPlan, error)
	Validate(plan *PartitionPlan) error
	EstimateLatency(plan *PartitionPlan) time.Duration
	EstimateMemoryUsage(plan *PartitionPlan) int64
}

// PartitionPlan defines how to partition work across nodes
type PartitionPlan struct {
	ID          string
	RequestID   string
	Strategy    string
	Partitions  []*Partition
	Coordinator string
	CreatedAt   time.Time
}

// Partition represents a work partition
type Partition struct {
	ID        string
	NodeID    string
	StartIdx  int
	EndIdx    int
	ModelPart string
	Data      []byte
}

// LayerPartitionStrategy partitions by model layers
type LayerPartitionStrategy struct {
	layerMapping map[string][]LayerInfo
	mu           sync.RWMutex
}

// LayerInfo contains layer information
type LayerInfo struct {
	Name        string
	Type        string
	Parameters  int64
	MemoryUsage int64
	ComputeCost float64
}

// TensorPartitionStrategy partitions by tensors
type TensorPartitionStrategy struct {
	tensorMapping map[string][]TensorInfo
	mu            sync.RWMutex
}

// TensorInfo contains tensor information
type TensorInfo struct {
	Name   string
	Shape  []int
	Size   int64
	Device string
}

// PipelinePartitionStrategy implements pipeline parallelism
type PipelinePartitionStrategy struct {
	stageMapping map[string][]PipelineStage
	mu           sync.RWMutex
}

// PipelineStage represents a pipeline stage
type PipelineStage struct {
	ID          string
	Name        string
	Layers      []string
	Dependencies []string
	MemoryReq   int64
	ComputeReq  float64
}

// DataPartitionStrategy partitions input data
type DataPartitionStrategy struct {
	chunkSize int
	mu        sync.RWMutex
}

// NewLayerPartitionStrategy creates layer-based partitioning strategy
func NewLayerPartitionStrategy() *LayerPartitionStrategy {
	return &LayerPartitionStrategy{
		layerMapping: make(map[string][]LayerInfo),
	}
}

// Partition partitions work by model layers
func (lps *LayerPartitionStrategy) Partition(ctx context.Context, request *InferenceRequest, nodes []NodeInfo) (*PartitionPlan, error) {
	lps.mu.RLock()
	layers, exists := lps.layerMapping[request.Model]
	lps.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("no layer mapping found for model %s", request.Model)
	}
	
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}
	
	// Distribute layers across nodes
	partitions := make([]*Partition, 0, len(nodes))
	layersPerNode := int(math.Ceil(float64(len(layers)) / float64(len(nodes))))
	
	for i, node := range nodes {
		startIdx := i * layersPerNode
		endIdx := startIdx + layersPerNode
		if endIdx > len(layers) {
			endIdx = len(layers)
		}
		
		if startIdx < len(layers) {
			partition := &Partition{
				ID:        fmt.Sprintf("layer_%s_%d", node.ID, i),
				NodeID:    node.ID,
				StartIdx:  startIdx,
				EndIdx:    endIdx,
				ModelPart: fmt.Sprintf("layers_%d_%d", startIdx, endIdx),
			}
			partitions = append(partitions, partition)
		}
	}
	
	plan := &PartitionPlan{
		ID:          fmt.Sprintf("layer_plan_%s", request.ID),
		RequestID:   request.ID,
		Strategy:    "layer",
		Partitions:  partitions,
		Coordinator: nodes[0].ID, // First node coordinates
		CreatedAt:   time.Now(),
	}
	
	return plan, nil
}

// Validate validates a layer partition plan
func (lps *LayerPartitionStrategy) Validate(plan *PartitionPlan) error {
	if plan == nil {
		return fmt.Errorf("plan is nil")
	}
	
	if len(plan.Partitions) == 0 {
		return fmt.Errorf("no partitions in plan")
	}
	
	// Check for overlapping partitions
	for i, p1 := range plan.Partitions {
		for j, p2 := range plan.Partitions {
			if i != j && p1.NodeID == p2.NodeID && p1.StartIdx < p2.EndIdx && p2.StartIdx < p1.EndIdx {
				return fmt.Errorf("overlapping partitions on node %s", p1.NodeID)
			}
		}
	}
	
	return nil
}

// EstimateLatency estimates execution latency
func (lps *LayerPartitionStrategy) EstimateLatency(plan *PartitionPlan) time.Duration {
	if plan == nil || len(plan.Partitions) == 0 {
		return time.Hour // High latency for invalid plans
	}
	
	// Estimate based on largest partition (pipeline bottleneck)
	maxLayers := 0
	for _, partition := range plan.Partitions {
		layers := partition.EndIdx - partition.StartIdx
		if layers > maxLayers {
			maxLayers = layers
		}
	}
	
	// Rough estimate: 100ms per layer
	return time.Duration(maxLayers*100) * time.Millisecond
}

// EstimateMemoryUsage estimates memory usage
func (lps *LayerPartitionStrategy) EstimateMemoryUsage(plan *PartitionPlan) int64 {
	if plan == nil || len(plan.Partitions) == 0 {
		return 0
	}
	
	// Sum memory across all partitions
	var totalMemory int64
	for _, partition := range plan.Partitions {
		layers := partition.EndIdx - partition.StartIdx
		// Rough estimate: 1GB per layer
		totalMemory += int64(layers) * 1024 * 1024 * 1024
	}
	
	return totalMemory
}

// NewTensorPartitionStrategy creates tensor-based partitioning strategy
func NewTensorPartitionStrategy() *TensorPartitionStrategy {
	return &TensorPartitionStrategy{
		tensorMapping: make(map[string][]TensorInfo),
	}
}

// Partition partitions work by tensors
func (tps *TensorPartitionStrategy) Partition(ctx context.Context, request *InferenceRequest, nodes []NodeInfo) (*PartitionPlan, error) {
	tps.mu.RLock()
	tensors, exists := tps.tensorMapping[request.Model]
	tps.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("no tensor mapping found for model %s", request.Model)
	}
	
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}
	
	// Distribute tensors across nodes by size
	partitions := make([]*Partition, 0, len(nodes))
	
	// Sort tensors by size (largest first)
	sortedTensors := make([]TensorInfo, len(tensors))
	copy(sortedTensors, tensors)
	
	// Simple round-robin distribution
	for i, tensor := range sortedTensors {
		nodeIdx := i % len(nodes)
		node := nodes[nodeIdx]
		
		partition := &Partition{
			ID:        fmt.Sprintf("tensor_%s_%s", node.ID, tensor.Name),
			NodeID:    node.ID,
			StartIdx:  i,
			EndIdx:    i + 1,
			ModelPart: tensor.Name,
		}
		partitions = append(partitions, partition)
	}
	
	plan := &PartitionPlan{
		ID:          fmt.Sprintf("tensor_plan_%s", request.ID),
		RequestID:   request.ID,
		Strategy:    "tensor",
		Partitions:  partitions,
		Coordinator: nodes[0].ID,
		CreatedAt:   time.Now(),
	}
	
	return plan, nil
}

// Validate validates a tensor partition plan
func (tps *TensorPartitionStrategy) Validate(plan *PartitionPlan) error {
	if plan == nil {
		return fmt.Errorf("plan is nil")
	}
	
	if len(plan.Partitions) == 0 {
		return fmt.Errorf("no partitions in plan")
	}
	
	return nil
}

// EstimateLatency estimates tensor partition latency
func (tps *TensorPartitionStrategy) EstimateLatency(plan *PartitionPlan) time.Duration {
	if plan == nil || len(plan.Partitions) == 0 {
		return time.Hour
	}
	
	// Tensor operations can be parallelized better
	// Estimate: 50ms base + 10ms per partition
	baseLatency := 50 * time.Millisecond
	partitionLatency := time.Duration(len(plan.Partitions)*10) * time.Millisecond
	
	return baseLatency + partitionLatency
}

// EstimateMemoryUsage estimates tensor memory usage
func (tps *TensorPartitionStrategy) EstimateMemoryUsage(plan *PartitionPlan) int64 {
	if plan == nil || len(plan.Partitions) == 0 {
		return 0
	}
	
	// Estimate: 500MB per tensor partition
	return int64(len(plan.Partitions)) * 500 * 1024 * 1024
}

// NewPipelinePartitionStrategy creates pipeline partitioning strategy
func NewPipelinePartitionStrategy() *PipelinePartitionStrategy {
	return &PipelinePartitionStrategy{
		stageMapping: make(map[string][]PipelineStage),
	}
}

// Partition partitions work using pipeline stages
func (pps *PipelinePartitionStrategy) Partition(ctx context.Context, request *InferenceRequest, nodes []NodeInfo) (*PartitionPlan, error) {
	pps.mu.RLock()
	stages, exists := pps.stageMapping[request.Model]
	pps.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("no pipeline mapping found for model %s", request.Model)
	}
	
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}
	
	// Assign stages to nodes
	partitions := make([]*Partition, 0, len(stages))
	
	for i, stage := range stages {
		nodeIdx := i % len(nodes)
		node := nodes[nodeIdx]
		
		partition := &Partition{
			ID:        fmt.Sprintf("stage_%s_%s", node.ID, stage.ID),
			NodeID:    node.ID,
			StartIdx:  i,
			EndIdx:    i + 1,
			ModelPart: stage.Name,
		}
		partitions = append(partitions, partition)
	}
	
	plan := &PartitionPlan{
		ID:          fmt.Sprintf("pipeline_plan_%s", request.ID),
		RequestID:   request.ID,
		Strategy:    "pipeline",
		Partitions:  partitions,
		Coordinator: nodes[0].ID,
		CreatedAt:   time.Now(),
	}
	
	return plan, nil
}

// Validate validates a pipeline partition plan
func (pps *PipelinePartitionStrategy) Validate(plan *PartitionPlan) error {
	if plan == nil {
		return fmt.Errorf("plan is nil")
	}
	
	if len(plan.Partitions) == 0 {
		return fmt.Errorf("no partitions in plan")
	}
	
	return nil
}

// EstimateLatency estimates pipeline latency
func (pps *PipelinePartitionStrategy) EstimateLatency(plan *PartitionPlan) time.Duration {
	if plan == nil || len(plan.Partitions) == 0 {
		return time.Hour
	}
	
	// Pipeline latency is sum of stage latencies
	// Estimate: 80ms per stage
	return time.Duration(len(plan.Partitions)*80) * time.Millisecond
}

// EstimateMemoryUsage estimates pipeline memory usage
func (pps *PipelinePartitionStrategy) EstimateMemoryUsage(plan *PartitionPlan) int64 {
	if plan == nil || len(plan.Partitions) == 0 {
		return 0
	}
	
	// Pipeline stages share memory more efficiently
	// Estimate: 2GB base + 200MB per stage
	baseMemory := int64(2 * 1024 * 1024 * 1024)
	stageMemory := int64(len(plan.Partitions)) * 200 * 1024 * 1024
	
	return baseMemory + stageMemory
}

// NewDataPartitionStrategy creates data partitioning strategy
func NewDataPartitionStrategy(chunkSize int) *DataPartitionStrategy {
	return &DataPartitionStrategy{
		chunkSize: chunkSize,
	}
}

// Partition partitions input data
func (dps *DataPartitionStrategy) Partition(ctx context.Context, request *InferenceRequest, nodes []NodeInfo) (*PartitionPlan, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}
	
	// For data partitioning, we split the input
	inputLength := len(request.Prompt)
	if inputLength == 0 {
		return nil, fmt.Errorf("no input data to partition")
	}
	
	partitions := make([]*Partition, 0, len(nodes))
	chunkSize := dps.chunkSize
	if chunkSize <= 0 {
		chunkSize = inputLength / len(nodes)
		if chunkSize == 0 {
			chunkSize = 1
		}
	}
	
	for i, node := range nodes {
		startIdx := i * chunkSize
		endIdx := startIdx + chunkSize
		if endIdx > inputLength {
			endIdx = inputLength
		}
		
		if startIdx < inputLength {
			partition := &Partition{
				ID:        fmt.Sprintf("data_%s_%d", node.ID, i),
				NodeID:    node.ID,
				StartIdx:  startIdx,
				EndIdx:    endIdx,
				ModelPart: fmt.Sprintf("chunk_%d_%d", startIdx, endIdx),
				Data:      []byte(request.Prompt[startIdx:endIdx]),
			}
			partitions = append(partitions, partition)
		}
	}
	
	plan := &PartitionPlan{
		ID:          fmt.Sprintf("data_plan_%s", request.ID),
		RequestID:   request.ID,
		Strategy:    "data",
		Partitions:  partitions,
		Coordinator: nodes[0].ID,
		CreatedAt:   time.Now(),
	}
	
	return plan, nil
}

// Validate validates a data partition plan
func (dps *DataPartitionStrategy) Validate(plan *PartitionPlan) error {
	if plan == nil {
		return fmt.Errorf("plan is nil")
	}
	
	if len(plan.Partitions) == 0 {
		return fmt.Errorf("no partitions in plan")
	}
	
	return nil
}

// EstimateLatency estimates data partition latency
func (dps *DataPartitionStrategy) EstimateLatency(plan *PartitionPlan) time.Duration {
	if plan == nil || len(plan.Partitions) == 0 {
		return time.Hour
	}
	
	// Data partitioning has parallel processing benefits
	// Estimate: 30ms base + 5ms per partition
	baseLatency := 30 * time.Millisecond
	partitionLatency := time.Duration(len(plan.Partitions)*5) * time.Millisecond
	
	return baseLatency + partitionLatency
}

// EstimateMemoryUsage estimates data partition memory usage
func (dps *DataPartitionStrategy) EstimateMemoryUsage(plan *PartitionPlan) int64 {
	if plan == nil || len(plan.Partitions) == 0 {
		return 0
	}
	
	var totalMemory int64
	for _, partition := range plan.Partitions {
		if partition.Data != nil {
			totalMemory += int64(len(partition.Data))
		}
	}
	
	// Add base memory overhead: 100MB
	totalMemory += 100 * 1024 * 1024
	
	return totalMemory
}
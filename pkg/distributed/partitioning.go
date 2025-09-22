package distributed

import (
	"context"
	"fmt"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/partitioning"
	api_types "github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/types"
)

// PartitionStrategy defines interface for different partitioning approaches
// This is now a thin wrapper around the enhanced partitioning system
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

// LayerInfo contains layer information (kept for backward compatibility)
type LayerInfo struct {
	Name        string
	Type        string
	Parameters  int64
	MemoryUsage int64
	ComputeCost float64
}

// TensorInfo contains tensor information (kept for backward compatibility)
type TensorInfo struct {
	Name   string
	Shape  []int
	Size   int64
	Device string
}

// PipelineStage represents a pipeline stage (kept for backward compatibility)
type PipelineStage struct {
	ID          string
	Name        string
	Layers      []string
	Dependencies []string
	MemoryReq   int64
	ComputeReq  float64
}

// EnhancedPartitioningAdapter wraps the new enhanced partitioning system
type EnhancedPartitioningAdapter struct {
	manager partitioning.EnhancedPartitionManager
}

// NewEnhancedPartitioningAdapter creates a new adapter that uses the enhanced partitioning system
func NewEnhancedPartitioningAdapter() *EnhancedPartitioningAdapter {
	return &EnhancedPartitioningAdapter{
		manager: partitioning.NewEnhancedPartitionManager(),
	}
}

// Partition delegates to the enhanced partition manager
func (epa *EnhancedPartitioningAdapter) Partition(ctx context.Context, request *InferenceRequest, nodes []NodeInfo) (*PartitionPlan, error) {
	// Convert InferenceRequest to DistributedTask (minimal required fields)
	task := &api_types.DistributedTask{
		ID:        api_types.TaskID(request.ID),
		ModelName: request.Model,
		CreatedAt: time.Now(),
	}
	
	// Extract preferred node IDs and capabilities from the nodes argument
	var preferredNodeIDs []string
	var nodeCapabilities []map[string]interface{}
	for _, node := range nodes {
		preferredNodeIDs = append(preferredNodeIDs, node.ID)

		// Build capability snapshot for each node
		capSnapshot := map[string]interface{}{
			"id":           node.ID,
			"cpu_cores":    node.CPUCores,
			"mem_total":    node.TotalMemory,
			"mem_available": node.AvailableMemory,
			"gpu_count":    node.GPUCount,
			"net_bw":       node.NetworkBandwidth,
			"net_latency":  node.NetworkLatency,
		}
		nodeCapabilities = append(nodeCapabilities, capSnapshot)
	}

	// Store options, preferred node IDs, and capabilities in metadata
	if task.Metadata == nil {
		task.Metadata = make(map[string]interface{})
	}

	if request.Options != nil {
		task.Metadata["options"] = request.Options
	}

	// Set preferred node IDs for the enhanced manager
	if len(preferredNodeIDs) > 0 {
		task.Metadata["preferred_node_ids"] = preferredNodeIDs
	}

	// Pass node capabilities through the adapter
	if len(nodeCapabilities) > 0 {
		task.Metadata["preferred_nodes_capabilities"] = nodeCapabilities
	}

	// Use the enhanced partition manager
	enhancedPlan, err := epa.manager.Partition(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("enhanced partitioning failed: %w", err)
	}

	// Convert back to legacy PartitionPlan
	return convertEnhancedPlanToLegacy(enhancedPlan, request.ID), nil
}

// Validate validates the partition plan
func (epa *EnhancedPartitioningAdapter) Validate(plan *PartitionPlan) error {
	if plan == nil {
		return fmt.Errorf("partition plan is nil")
	}
	if len(plan.Partitions) == 0 {
		return fmt.Errorf("no partitions in plan")
	}
	return nil
}

// EstimateLatency estimates execution latency
func (epa *EnhancedPartitioningAdapter) EstimateLatency(plan *PartitionPlan) time.Duration {
	// Simple estimation based on partition count
	return time.Duration(len(plan.Partitions)*100) * time.Millisecond
}

// EstimateMemoryUsage estimates memory usage
func (epa *EnhancedPartitioningAdapter) EstimateMemoryUsage(plan *PartitionPlan) int64 {
	// Simple estimation
	return int64(len(plan.Partitions)) * 1_000_000_000 // 1GB per partition
}

// convertEnhancedPlanToLegacy converts an enhanced partition plan to legacy format
func convertEnhancedPlanToLegacy(enhancedPlan *partitioning.PartitionPlan, requestID string) *PartitionPlan {
	if enhancedPlan == nil {
		return nil
	}

	plan := &PartitionPlan{
		ID:        enhancedPlan.ID,
		RequestID: requestID,
		Strategy:  enhancedPlan.Strategy,
		CreatedAt: enhancedPlan.CreatedAt,
	}

	// Convert assignments to partitions
	for i, assignment := range enhancedPlan.Assignments {
		partition := &Partition{
			ID:     fmt.Sprintf("partition_%d", i),
			NodeID: assignment.NodeID,
		}

		// Handle different work types
		switch assignment.WorkType {
		case partitioning.WorkTypePipelineStage:
			// Handle pipeline stage ranges
			if assignment.Assignment != nil && assignment.Assignment.PipelineStage != nil {
				stage := assignment.Assignment.PipelineStage
				if len(stage.LayerRange) >= 2 {
					partition.StartIdx = stage.LayerRange[0]
					partition.EndIdx = stage.LayerRange[1]
					partition.ModelPart = fmt.Sprintf("pipeline_stage_%d_%d", stage.LayerRange[0], stage.LayerRange[1])
				} else {
					// Fallback to default range if LayerRange is incomplete
					partition.StartIdx = 0
					partition.EndIdx = 10
					partition.ModelPart = "pipeline_stage_0_10"
				}
			} else {
				// Default values for pipeline stage if no stage info available
				partition.StartIdx = 0
				partition.EndIdx = 10
				partition.ModelPart = "pipeline_stage_0_10"
			}
		case partitioning.WorkTypeTensors:
			// For tensor work type, don't use layer indices
			partition.StartIdx = -1
			partition.EndIdx = -1
			partition.ModelPart = "tensors"
		case partitioning.WorkTypeHybrid:
			// For hybrid work type, don't use layer indices
			partition.StartIdx = -1
			partition.EndIdx = -1
			partition.ModelPart = "hybrid"
		case partitioning.WorkTypeLayers:
			// Extract layer indices for layer-based assignments
			if assignment.Assignment != nil && len(assignment.Assignment.Layers) > 0 {
				layer := assignment.Assignment.Layers[0]
				partition.StartIdx = layer.StartIndex
				partition.EndIdx = layer.EndIndex
				partition.ModelPart = fmt.Sprintf("layers_%d_%d", layer.StartIndex, layer.EndIndex)
			} else {
				// Default values if no layer info available
				partition.StartIdx = 0
				partition.EndIdx = 10
				partition.ModelPart = "layers_0_10"
			}
		default:
			// For unknown work types, use -1 to indicate not applicable
			partition.StartIdx = -1
			partition.EndIdx = -1
			partition.ModelPart = string(assignment.WorkType)
		}

		plan.Partitions = append(plan.Partitions, partition)
	}

	// Set coordinator as the first node
	if len(enhancedPlan.Assignments) > 0 {
		plan.Coordinator = enhancedPlan.Assignments[0].NodeID
	}

	return plan
}

// Legacy strategy implementations that delegate to the enhanced system

// LayerPartitionStrategy partitions by model layers (delegates to enhanced system)
type LayerPartitionStrategy struct {
	adapter *EnhancedPartitioningAdapter
}

// NewLayerPartitionStrategy creates layer-based partitioning strategy
func NewLayerPartitionStrategy() *LayerPartitionStrategy {
	return &LayerPartitionStrategy{
		adapter: NewEnhancedPartitioningAdapter(),
	}
}

// Partition partitions work by model layers
func (lps *LayerPartitionStrategy) Partition(ctx context.Context, request *InferenceRequest, nodes []NodeInfo) (*PartitionPlan, error) {
	return lps.adapter.Partition(ctx, request, nodes)
}

// Validate validates the partition plan
func (lps *LayerPartitionStrategy) Validate(plan *PartitionPlan) error {
	return lps.adapter.Validate(plan)
}

// EstimateLatency estimates execution latency
func (lps *LayerPartitionStrategy) EstimateLatency(plan *PartitionPlan) time.Duration {
	return lps.adapter.EstimateLatency(plan)
}

// EstimateMemoryUsage estimates memory usage
func (lps *LayerPartitionStrategy) EstimateMemoryUsage(plan *PartitionPlan) int64 {
	return lps.adapter.EstimateMemoryUsage(plan)
}

// TensorPartitionStrategy partitions by tensors (delegates to enhanced system)
type TensorPartitionStrategy struct {
	adapter *EnhancedPartitioningAdapter
}

// NewTensorPartitionStrategy creates tensor-based partitioning strategy
func NewTensorPartitionStrategy() *TensorPartitionStrategy {
	return &TensorPartitionStrategy{
		adapter: NewEnhancedPartitioningAdapter(),
	}
}

// Partition partitions work by tensors
func (tps *TensorPartitionStrategy) Partition(ctx context.Context, request *InferenceRequest, nodes []NodeInfo) (*PartitionPlan, error) {
	return tps.adapter.Partition(ctx, request, nodes)
}

// Validate validates the partition plan
func (tps *TensorPartitionStrategy) Validate(plan *PartitionPlan) error {
	return tps.adapter.Validate(plan)
}

// EstimateLatency estimates execution latency
func (tps *TensorPartitionStrategy) EstimateLatency(plan *PartitionPlan) time.Duration {
	return tps.adapter.EstimateLatency(plan)
}

// EstimateMemoryUsage estimates memory usage
func (tps *TensorPartitionStrategy) EstimateMemoryUsage(plan *PartitionPlan) int64 {
	return tps.adapter.EstimateMemoryUsage(plan)
}

// PipelinePartitionStrategy implements pipeline parallelism (delegates to enhanced system)
type PipelinePartitionStrategy struct {
	adapter *EnhancedPartitioningAdapter
}

// NewPipelinePartitionStrategy creates pipeline-based partitioning strategy
func NewPipelinePartitionStrategy() *PipelinePartitionStrategy {
	return &PipelinePartitionStrategy{
		adapter: NewEnhancedPartitioningAdapter(),
	}
}

// Partition partitions work in pipeline stages
func (pps *PipelinePartitionStrategy) Partition(ctx context.Context, request *InferenceRequest, nodes []NodeInfo) (*PartitionPlan, error) {
	return pps.adapter.Partition(ctx, request, nodes)
}

// Validate validates the partition plan
func (pps *PipelinePartitionStrategy) Validate(plan *PartitionPlan) error {
	return pps.adapter.Validate(plan)
}

// EstimateLatency estimates execution latency
func (pps *PipelinePartitionStrategy) EstimateLatency(plan *PartitionPlan) time.Duration {
	return pps.adapter.EstimateLatency(plan)
}

// EstimateMemoryUsage estimates memory usage
func (pps *PipelinePartitionStrategy) EstimateMemoryUsage(plan *PartitionPlan) int64 {
	return pps.adapter.EstimateMemoryUsage(plan)
}

// DataPartitionStrategy partitions input data (simple implementation, not using enhanced system)
type DataPartitionStrategy struct {
	chunkSize int
}

// NewDataPartitionStrategy creates data-based partitioning strategy
func NewDataPartitionStrategy(chunkSize int) *DataPartitionStrategy {
	if chunkSize <= 0 {
		chunkSize = 1024 // Default chunk size
	}
	return &DataPartitionStrategy{
		chunkSize: chunkSize,
	}
}

// Partition partitions work by data chunks
func (dps *DataPartitionStrategy) Partition(ctx context.Context, request *InferenceRequest, nodes []NodeInfo) (*PartitionPlan, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	plan := &PartitionPlan{
		ID:          fmt.Sprintf("plan_%s_%d", request.ID, time.Now().Unix()),
		RequestID:   request.ID,
		Strategy:    "data",
		Partitions:  make([]*Partition, 0),
		Coordinator: nodes[0].ID,
		CreatedAt:   time.Now(),
	}

	// Simple round-robin data partitioning
	dataLen := len(request.Prompt)
	nodeCount := len(nodes)
	chunkSize := dataLen / nodeCount
	if chunkSize < 1 {
		chunkSize = 1
	}

	for i, node := range nodes {
		startIdx := i * chunkSize
		endIdx := startIdx + chunkSize
		if i == nodeCount-1 {
			endIdx = dataLen
		}

		partition := &Partition{
			ID:        fmt.Sprintf("partition_%d", i),
			NodeID:    node.ID,
			StartIdx:  startIdx,
			EndIdx:    endIdx,
			ModelPart: "full",
		}

		if startIdx < dataLen {
			partition.Data = []byte(request.Prompt[startIdx:min(endIdx, dataLen)])
		}

		plan.Partitions = append(plan.Partitions, partition)
	}

	return plan, nil
}

// Validate validates the partition plan
func (dps *DataPartitionStrategy) Validate(plan *PartitionPlan) error {
	if plan == nil {
		return fmt.Errorf("partition plan is nil")
	}
	if len(plan.Partitions) == 0 {
		return fmt.Errorf("no partitions in plan")
	}
	return nil
}

// EstimateLatency estimates execution latency
func (dps *DataPartitionStrategy) EstimateLatency(plan *PartitionPlan) time.Duration {
	return time.Duration(len(plan.Partitions)*50) * time.Millisecond
}

// EstimateMemoryUsage estimates memory usage
func (dps *DataPartitionStrategy) EstimateMemoryUsage(plan *PartitionPlan) int64 {
	var totalSize int64
	for _, partition := range plan.Partitions {
		totalSize += int64(len(partition.Data))
	}
	return totalSize * 2 // Estimate 2x for processing overhead
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
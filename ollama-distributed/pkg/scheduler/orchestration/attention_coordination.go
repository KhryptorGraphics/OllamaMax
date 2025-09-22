package orchestration

import (
	"fmt"
	"errors"
	"sort"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/partitioning"
)

// Feature flag for distributed attention computation
var EnableDistributedAttention bool = false

// AttentionOutput represents the output from a single attention computation
type AttentionOutput struct {
	Values    []float32                // Attention values
	Heads     int                      // Number of attention heads
	SeqLen    int                      // Sequence length
	HeadDim   int                      // Dimension per head
	LayerID   int                      // Layer identifier
	PartitionID string                 // Partition identifier
	Metadata  map[string]interface{}   // Additional metadata
}

// AttentionState represents the state of attention computation
type AttentionState struct {
	QueryStates  []float32  // Query vectors
	KeyStates    []float32  // Key vectors
	ValueStates  []float32  // Value vectors
	AttentionMask []bool     // Attention mask
	LayerID      int        // Layer identifier
	HeadID       int        // Head identifier
}

// AttentionCoordinator manages attention mechanism coordination
type AttentionCoordinator struct {
	partitionManager *AttentionPartitionManager
	aggregator       *AttentionAggregator
	stateManager     *AttentionStateManager
	modelAnalyzer    *partitioning.ModelAnalyzer
}

func NewAttentionCoordinator() *AttentionCoordinator {
	return &AttentionCoordinator{
		partitionManager: NewAttentionPartitionManager(),
		aggregator:       NewAttentionAggregator(),
		stateManager:     NewAttentionStateManager(),
		modelAnalyzer:    partitioning.NewModelAnalyzer(),
	}
}

// AggregateAttentionHeads combines attention outputs from different heads/nodes
func (ac *AttentionCoordinator) AggregateAttentionHeads(outputs []AttentionOutput) (*TensorData, error) {
	if len(outputs) == 0 {
		return nil, errors.New("no attention outputs to aggregate")
	}

	// Sort outputs by layer and head for consistent aggregation
	sort.Slice(outputs, func(i, j int) bool {
		if outputs[i].LayerID != outputs[j].LayerID {
			return outputs[i].LayerID < outputs[j].LayerID
		}
		return outputs[i].PartitionID < outputs[j].PartitionID
	})

	// Group outputs by layer
	layerOutputs := make(map[int][]AttentionOutput)
	for _, output := range outputs {
		layerOutputs[output.LayerID] = append(layerOutputs[output.LayerID], output)
	}

	// Aggregate each layer separately
	aggregatedLayers := make([]TensorData, 0)
	for layerID, layerOutputs := range layerOutputs {
		aggregated, err := ac.aggregator.AggregateLayer(layerOutputs)
		if err != nil {
			return nil, fmt.Errorf("failed to aggregate layer %d: %v", layerID, err)
		}
		aggregatedLayers = append(aggregatedLayers, *aggregated)
	}

	// Combine all layers
	if len(aggregatedLayers) == 1 {
		return &aggregatedLayers[0], nil
	}

	// Stack layers along layer dimension
	tensorAggregator := NewTensorAggregator()
	result, err := tensorAggregator.ConcatenateTensors(aggregatedLayers, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to combine layers: %v", err)
	}

	return result, nil
}

// CoordinateAttentionComputation coordinates attention across partitions
func (ac *AttentionCoordinator) CoordinateAttentionComputation(
	states []AttentionState,
	modelAnalysis *partitioning.ModelAnalysis,
) (*AttentionOutput, error) {

	// Partition attention computation across available nodes
	partitionPlan, err := ac.partitionManager.CreateAttentionPartitionPlan(states, modelAnalysis)
	if err != nil {
		return nil, fmt.Errorf("failed to create attention partition plan: %v", err)
	}

	// Execute attention computation on partitions
	partitionResults := make([]AttentionOutput, len(partitionPlan.Partitions))
	for i, partition := range partitionPlan.Partitions {
		result, err := ac.executeAttentionPartition(partition, states)
		if err != nil {
			return nil, fmt.Errorf("failed to execute attention partition %s (index %d): %v", partition.ID, i, err)
		}
		partitionResults[i] = *result
	}

	// Aggregate results from all partitions
	aggregated, err := ac.aggregator.AggregatePartitions(partitionResults)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate attention partitions: %v", err)
	}

	return aggregated, nil
}

// AttentionPartitionManager distributes attention computation across nodes
type AttentionPartitionManager struct{}

func NewAttentionPartitionManager() *AttentionPartitionManager {
	return &AttentionPartitionManager{}
}

// AttentionPartitionPlan defines how attention computation is distributed
type AttentionPartitionPlan struct {
	Partitions    []AttentionPartition
	Strategy      string
	TotalHeads    int
	HeadsPerNode  int
	ModelAnalysis *partitioning.ModelAnalysis
}

// AttentionPartition represents a partition of attention computation
type AttentionPartition struct {
	ID           string
	NodeID       string
	HeadRange    []int  // Range of attention heads [start, end)
	LayerRange   []int  // Range of layers [start, end)
	States       []AttentionState
	Metadata     map[string]interface{}
}

// CreateAttentionPartitionPlan creates a plan for distributing attention computation
func (apm *AttentionPartitionManager) CreateAttentionPartitionPlan(
	states []AttentionState,
	modelAnalysis *partitioning.ModelAnalysis,
) (*AttentionPartitionPlan, error) {

	if modelAnalysis == nil {
		return nil, errors.New("model analysis is required")
	}

	// Extract attention configuration from model analysis with proper type handling
	totalHeads := 8 // Default
	totalLayers := 0
	if modelAnalysis.LayerInfo != nil {
		totalLayers = modelAnalysis.LayerInfo.TotalLayers
	}
	// For now, use defaults as we don't have additional attention info
	// In the future, this could come from model-specific metadata

	// Calculate optimal partitioning
	nodeCount := 3 // Should come from available nodes
	headsPerNode := totalHeads / nodeCount
	if totalHeads%nodeCount != 0 {
		headsPerNode++
	}

	partitions := make([]AttentionPartition, nodeCount)
	for i := 0; i < nodeCount; i++ {
		startHead := i * headsPerNode
		endHead := min((i+1)*headsPerNode, totalHeads)

		// Filter states for this partition
		partitionStates := make([]AttentionState, 0)
		for _, state := range states {
			if state.HeadID >= startHead && state.HeadID < endHead {
				partitionStates = append(partitionStates, state)
			}
		}

		partitions[i] = AttentionPartition{
			ID:         fmt.Sprintf("attention_partition_%d", i),
			NodeID:     fmt.Sprintf("node_%d", i),
			HeadRange:  []int{startHead, endHead},
			LayerRange: []int{0, totalLayers},
			States:     partitionStates,
			Metadata: map[string]interface{}{
				"heads_per_partition": endHead - startHead,
				"total_heads":         totalHeads,
			},
		}
	}

	return &AttentionPartitionPlan{
		Partitions:    partitions,
		Strategy:      "head_parallel",
		TotalHeads:    totalHeads,
		HeadsPerNode:  headsPerNode,
		ModelAnalysis: modelAnalysis,
	}, nil
}

// DistributeAttentionHeads distributes attention heads across nodes
func (apm *AttentionPartitionManager) DistributeAttentionHeads(
	totalHeads int,
	nodeCount int,
	modelAnalysis *partitioning.ModelAnalysis,
) ([][]int, error) {

	if nodeCount <= 0 {
		return nil, errors.New("node count must be positive")
	}

	if totalHeads <= 0 {
		return nil, errors.New("total heads must be positive")
	}

	headsPerNode := totalHeads / nodeCount
	remainder := totalHeads % nodeCount

	distribution := make([][]int, nodeCount)
	currentHead := 0

	for i := 0; i < nodeCount; i++ {
		headsForThisNode := headsPerNode
		if i < remainder {
			headsForThisNode++
		}

		distribution[i] = []int{currentHead, currentHead + headsForThisNode}
		currentHead += headsForThisNode
	}

	return distribution, nil
}

// AttentionAggregator combines attention outputs
type AttentionAggregator struct {
	tensorAggregator *TensorAggregator
}

func NewAttentionAggregator() *AttentionAggregator {
	return &AttentionAggregator{
		tensorAggregator: NewTensorAggregator(),
	}
}

// AggregateLayer aggregates attention outputs from a single layer
func (aa *AttentionAggregator) AggregateLayer(outputs []AttentionOutput) (*TensorData, error) {
	if len(outputs) == 0 {
		return nil, errors.New("no outputs to aggregate")
	}

	// Convert attention outputs to tensor data
	tensors := make([]TensorData, len(outputs))
	for i, output := range outputs {
		tensors[i] = TensorData{
			Shape: []int{output.SeqLen, output.Heads * output.HeadDim},
			Data:  output.Values,
			Type:  "attention",
		}
	}

	// Concatenate along head dimension (dimension 1)
	result, err := aa.tensorAggregator.ConcatenateTensors(tensors, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to concatenate attention tensors: %v", err)
	}

	// Reshape to final attention output format
	totalHeads := 0
	for _, output := range outputs {
		totalHeads += output.Heads
	}

	if len(outputs) > 0 {
		seqLen := outputs[0].SeqLen
		headDim := outputs[0].HeadDim
		result.Shape = []int{seqLen, totalHeads, headDim}
	}

	return result, nil
}

// AggregatePartitions aggregates attention outputs from different partitions
func (aa *AttentionAggregator) AggregatePartitions(outputs []AttentionOutput) (*AttentionOutput, error) {
	if len(outputs) == 0 {
		return nil, errors.New("no outputs to aggregate")
	}

	// Aggregate tensor data
	aggregated, err := aa.AggregateLayer(outputs)
	if err != nil {
		return nil, err
	}

	// Combine metadata
	totalHeads := 0
	for _, output := range outputs {
		totalHeads += output.Heads
	}

	return &AttentionOutput{
		Values:      aggregated.Data,
		Heads:       totalHeads,
		SeqLen:      outputs[0].SeqLen,
		HeadDim:     outputs[0].HeadDim,
		LayerID:     outputs[0].LayerID,
		PartitionID: "aggregated",
		Metadata: map[string]interface{}{
			"aggregated_from": len(outputs),
			"total_heads":     totalHeads,
		},
	}, nil
}

// CombineAttentionHeads combines multiple attention heads into final output
func (aa *AttentionAggregator) CombineAttentionHeads(
	heads [][]float32,
	seqLen int,
	headDim int,
) ([]float32, error) {

	if len(heads) == 0 {
		return nil, errors.New("no attention heads to combine")
	}

	numHeads := len(heads)
	totalDim := numHeads * headDim
	output := make([]float32, seqLen*totalDim)

	// Interleave attention heads
	for seq := 0; seq < seqLen; seq++ {
		for head := 0; head < numHeads; head++ {
			for dim := 0; dim < headDim; dim++ {
				srcIdx := seq*headDim + dim
				dstIdx := seq*totalDim + head*headDim + dim

				if srcIdx < len(heads[head]) && dstIdx < len(output) {
					output[dstIdx] = heads[head][srcIdx]
				}
			}
		}
	}

	return output, nil
}

// NormalizeAttentionWeights applies softmax normalization to attention weights
func (aa *AttentionAggregator) NormalizeAttentionWeights(weights []float32, seqLen int) []float32 {
	// If per-sequence normalization is needed, slice and apply shared normalization
	if seqLen > 0 && len(weights) >= seqLen*seqLen {
		normalized := make([]float32, len(weights))

		// Apply softmax for each sequence position using shared utility
		for seq := 0; seq < seqLen; seq++ {
			start := seq * seqLen
			end := start + seqLen

			if end <= len(weights) {
				// Use shared normalization on this sequence segment
				segment := weights[start:end]
				normalizedSegment := SharedAttentionNormalization(segment)
				copy(normalized[start:end], normalizedSegment)
			}
		}

		return normalized
	}

	// Fallback to direct shared normalization for entire weight array
	return SharedAttentionNormalization(weights)
}

// AttentionStateManager tracks attention computation state
type AttentionStateManager struct {
	states map[string]*AttentionComputationSession
}

func NewAttentionStateManager() *AttentionStateManager {
	return &AttentionStateManager{
		states: make(map[string]*AttentionComputationSession),
	}
}

// AttentionComputationSession represents a session of attention computation
type AttentionComputationSession struct {
	ID              string
	LayerStates     map[int][]AttentionState
	Progress        map[string]float32
	Dependencies    map[int][]int  // Layer dependencies
	CompletedLayers []int
	Metadata        map[string]interface{}
}

// CreateSession creates a new attention computation session
func (asm *AttentionStateManager) CreateSession(
	sessionID string,
	modelAnalysis *partitioning.ModelAnalysis,
) *AttentionComputationSession {

	session := &AttentionComputationSession{
		ID:              sessionID,
		LayerStates:     make(map[int][]AttentionState),
		Progress:        make(map[string]float32),
		Dependencies:    make(map[int][]int),
		CompletedLayers: make([]int, 0),
		Metadata:        make(map[string]interface{}),
	}

	// Initialize layer dependencies from model analysis
	if modelAnalysis != nil && modelAnalysis.LayerInfo != nil {
		totalLayers := modelAnalysis.LayerInfo.TotalLayers
		for i := 0; i < totalLayers; i++ {
			// Simple sequential dependency for transformer layers
			if i > 0 {
				session.Dependencies[i] = []int{i - 1}
			}
		}
	}

	asm.states[sessionID] = session
	return session
}

// UpdateLayerState updates the state for a specific layer
func (asm *AttentionStateManager) UpdateLayerState(
	sessionID string,
	layerID int,
	state AttentionState,
) error {

	session, exists := asm.states[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	if session.LayerStates[layerID] == nil {
		session.LayerStates[layerID] = make([]AttentionState, 0)
	}

	session.LayerStates[layerID] = append(session.LayerStates[layerID], state)

	// Update progress
	progressKey := fmt.Sprintf("layer_%d", layerID)
	session.Progress[progressKey] = 1.0 // Mark as completed

	return nil
}

// CanProcessLayer checks if a layer's dependencies are satisfied
func (asm *AttentionStateManager) CanProcessLayer(sessionID string, layerID int) (bool, error) {
	session, exists := asm.states[sessionID]
	if !exists {
		return false, fmt.Errorf("session %s not found", sessionID)
	}

	dependencies, exists := session.Dependencies[layerID]
	if !exists {
		return true, nil // No dependencies
	}

	// Check if all dependencies are completed
	for _, depLayerID := range dependencies {
		completed := false
		for _, completedLayer := range session.CompletedLayers {
			if completedLayer == depLayerID {
				completed = true
				break
			}
		}
		if !completed {
			return false, nil
		}
	}

	return true, nil
}

// MarkLayerCompleted marks a layer as completed
func (asm *AttentionStateManager) MarkLayerCompleted(sessionID string, layerID int) error {
	session, exists := asm.states[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	// Add to completed layers if not already there
	for _, completed := range session.CompletedLayers {
		if completed == layerID {
			return nil // Already marked as completed
		}
	}

	session.CompletedLayers = append(session.CompletedLayers, layerID)
	return nil
}

// GetSessionProgress returns the overall progress of the session
func (asm *AttentionStateManager) GetSessionProgress(sessionID string) (float32, error) {
	session, exists := asm.states[sessionID]
	if !exists {
		return 0, fmt.Errorf("session %s not found", sessionID)
	}

	if len(session.Progress) == 0 {
		return 0, nil
	}

	var total float32
	for _, progress := range session.Progress {
		total += progress
	}

	return total / float32(len(session.Progress)), nil
}

// CleanupSession removes a completed session
func (asm *AttentionStateManager) CleanupSession(sessionID string) {
	delete(asm.states, sessionID)
}

// Helper function to execute attention partition with safe handling
func (ac *AttentionCoordinator) executeAttentionPartition(
	partition AttentionPartition,
	states []AttentionState,
) (*AttentionOutput, error) {

	// Check feature flag first
	if !EnableDistributedAttention {
		// Check if metadata explicitly enables distributed attention
		if partition.Metadata == nil || partition.Metadata["enable_distributed_attention"] != true {
			return nil, errors.New("distributed attention partition execution not implemented")
		}
	}

	// Extract and validate dimensions from metadata
	seqLen := GetIntFromMetadata(partition.Metadata, "sequence_length")
	headDim := GetIntFromMetadata(partition.Metadata, "head_dim")
	headsInPartition := GetIntFromMetadata(partition.Metadata, "heads_in_partition")

	// If not in metadata, try to calculate from head range
	if headsInPartition == 0 && len(partition.HeadRange) == 2 {
		headsInPartition = partition.HeadRange[1] - partition.HeadRange[0]
	}

	// Validate required dimensions
	if seqLen <= 0 || headDim <= 0 || headsInPartition <= 0 {
		return nil, fmt.Errorf("invalid dimensions for partition %s: seqLen=%d, headDim=%d, headsInPartition=%d",
			partition.ID, seqLen, headDim, headsInPartition)
	}

	// Minimal attention computation
	if len(states) == 0 {
		return nil, fmt.Errorf("no attention states provided for partition %s", partition.ID)
	}

	// Build Q, K, V tensors from states
	var allQ, allK, allV []float32
	for _, state := range states {
		// Validate state dimensions
		expectedSize := seqLen * headDim
		if len(state.QueryStates) != expectedSize || len(state.KeyStates) != expectedSize || len(state.ValueStates) != expectedSize {
			return nil, fmt.Errorf("state dimension mismatch for head %d in partition %s: expected %d, got Q=%d, K=%d, V=%d",
				state.HeadID, partition.ID, expectedSize, len(state.QueryStates), len(state.KeyStates), len(state.ValueStates))
		}
		allQ = append(allQ, state.QueryStates...)
		allK = append(allK, state.KeyStates...)
		allV = append(allV, state.ValueStates...)
	}

	// Compute attention scores: scores = softmax(Q @ K^T)
	outputSize := seqLen * headsInPartition * headDim
	values := make([]float32, outputSize)

	// Simple attention computation per head
	for h := 0; h < headsInPartition; h++ {
		for s := 0; s < seqLen; s++ {
			// Calculate attention scores for this sequence position
			scores := make([]float32, seqLen)
			for k := 0; k < seqLen; k++ {
				// Dot product between query and key
				var score float32
				for d := 0; d < headDim; d++ {
					qIdx := h*seqLen*headDim + s*headDim + d
					kIdx := h*seqLen*headDim + k*headDim + d
					if qIdx < len(allQ) && kIdx < len(allK) {
						score += allQ[qIdx] * allK[kIdx]
					}
				}
				scores[k] = score
			}

			// Apply softmax normalization
			normalizedScores := SharedAttentionNormalization(scores)

			// Multiply by values to get output
			for d := 0; d < headDim; d++ {
				var output float32
				for k := 0; k < seqLen; k++ {
					vIdx := h*seqLen*headDim + k*headDim + d
					if vIdx < len(allV) {
						output += normalizedScores[k] * allV[vIdx]
					}
				}
				outIdx := h*seqLen*headDim + s*headDim + d
				if outIdx < len(values) {
					values[outIdx] = output
				}
			}
		}
	}

	return &AttentionOutput{
		Values:      values,
		Heads:       headsInPartition,
		SeqLen:      seqLen,
		HeadDim:     headDim,
		LayerID:     partition.LayerRange[0],
		PartitionID: partition.ID,
		Metadata:    partition.Metadata,
	}, nil
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getIntFromMetadata is deprecated - use GetIntFromMetadata from tensor_util.go instead
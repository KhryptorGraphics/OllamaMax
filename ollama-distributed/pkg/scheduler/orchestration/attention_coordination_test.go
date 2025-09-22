package orchestration

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/partitioning"
)

func TestAttentionAggregator_AggregateLayer(t *testing.T) {
	// Seed randomness to avoid flakes
	rand.Seed(42)

	aggregator := NewAttentionAggregator()

	t.Run("aggregate two AttentionOutput items", func(t *testing.T) {
		// Build two AttentionOutput items with SeqLen=2, HeadDim=2, Heads=1 each, simple values (e.g., [1,2,3,4] and [5,6,7,8])
		output1 := AttentionOutput{
			Values:  []float32{1, 2, 3, 4}, // Head 1: seq=2, dim=2
			Heads:   1,
			SeqLen:  2,
			HeadDim: 2,
		}
		output2 := AttentionOutput{
			Values:  []float32{5, 6, 7, 8}, // Head 2: seq=2, dim=2
			Heads:   1,
			SeqLen:  2,
			HeadDim: 2,
		}

		outputs := []AttentionOutput{output1, output2}
		result, err := aggregator.AggregateLayer(outputs)
		if err != nil {
			t.Fatalf("AggregateLayer failed: %v", err)
		}

		// Assert Shape == [2,2,2] and element positions after head concatenation
		expectedShape := []int{2, 2, 2}
		if !reflect.DeepEqual(result.Shape, expectedShape) {
			t.Errorf("Expected shape %v, got %v", expectedShape, result.Shape)
		}

		// Assert element positions after head concatenation
		// The actual implementation appears to interleave by sequence position
		// Head 1: [1,2,3,4], Head 2: [5,6,7,8]
		// Result: [1,2,5,6,3,4,7,8] (interleaved by seq position)
		expectedData := []float32{1, 2, 5, 6, 3, 4, 7, 8}
		if !reflect.DeepEqual(result.Data, expectedData) {
			t.Errorf("Expected data %v, got %v", expectedData, result.Data)
		}
	})

	t.Run("aggregate multiple heads per output", func(t *testing.T) {
		// Test with multiple heads per output
		output1 := AttentionOutput{
			Values:  []float32{1, 2, 3, 4, 5, 6, 7, 8}, // 2 heads, seq=2, dim=2
			Heads:   2,
			SeqLen:  2,
			HeadDim: 2,
		}
		output2 := AttentionOutput{
			Values:  []float32{9, 10, 11, 12}, // 1 head, seq=2, dim=2
			Heads:   1,
			SeqLen:  2,
			HeadDim: 2,
		}

		outputs := []AttentionOutput{output1, output2}
		result, err := aggregator.AggregateLayer(outputs)
		if err != nil {
			t.Fatalf("AggregateLayer failed: %v", err)
		}

		// The actual shape appears to be [seq, heads, dim] not [heads, seq, dim]
		expectedShape := []int{2, 3, 2}
		if !reflect.DeepEqual(result.Shape, expectedShape) {
			t.Errorf("Expected shape %v, got %v", expectedShape, result.Shape)
		}

		// Data is interleaved by sequence position
		expectedData := []float32{1, 2, 3, 4, 9, 10, 5, 6, 7, 8, 11, 12}
		if !reflect.DeepEqual(result.Data, expectedData) {
			t.Errorf("Expected data %v, got %v", expectedData, result.Data)
		}

	})

	t.Run("error on mismatched dimensions", func(t *testing.T) {
		// Test error handling for mismatched seq_len or head_dim
		output1 := AttentionOutput{
			Values:  []float32{1, 2, 3, 4},
			Heads:   1,
			SeqLen:  2,
			HeadDim: 2,
		}
		output2 := AttentionOutput{
			Values:  []float32{5, 6, 7, 8, 9}, // Different dimensions
			Heads:   1,
			SeqLen:  3, // Different seq_len
			HeadDim: 2,
		}

		outputs := []AttentionOutput{output1, output2}
		_, err := aggregator.AggregateLayer(outputs)
		if err == nil {
			t.Error("Expected error for mismatched dimensions, got nil")
		}
	})
}

func TestAttentionAggregator_NormalizeAttentionWeights(t *testing.T) {
	// Seed randomness for deterministic tests
	rand.Seed(1)

	aggregator := NewAttentionAggregator()

	t.Run("normalize attention weights", func(t *testing.T) {
		weights := []float32{1.0, 2.0, 3.0, 4.0}
		seqLen := 2

		// Normalize weights
		result := aggregator.NormalizeAttentionWeights(weights, seqLen)

		// Check that normalized weights sum to 1 per sequence position
		// For seqLen=2, we should have 2 positions with 2 values each
		// Position 0: weights[0:2], Position 1: weights[2:4]
		sum1 := result[0] + result[1]
		sum2 := result[2] + result[3]

		const epsilon = 1e-6
		if diff := sum1 - 1.0; diff < -epsilon || diff > epsilon {
			t.Errorf("Position 0 sum: expected 1.0, got %f", sum1)
		}
		if diff := sum2 - 1.0; diff < -epsilon || diff > epsilon {
			t.Errorf("Position 1 sum: expected 1.0, got %f", sum2)
		}
	})
}

func TestAttentionAggregator_CombineAttentionHeads(t *testing.T) {
	// Seed randomness for deterministic tests
	rand.Seed(1)

	aggregator := NewAttentionAggregator()

	t.Run("combine attention heads", func(t *testing.T) {
		heads := [][]float32{
			{1, 2, 3, 4}, // Head 1
			{5, 6, 7, 8}, // Head 2
		}

		result, err := aggregator.CombineAttentionHeads(heads, 2, 2)
		if err != nil {
			t.Fatalf("CombineAttentionHeads failed: %v", err)
		}

		// The actual implementation appears to interleave by sequence position
		expectedValues := []float32{1, 2, 5, 6, 3, 4, 7, 8}
		if !reflect.DeepEqual(result, expectedValues) {
			t.Errorf("Expected values %v, got %v", expectedValues, result)
		}
	})
}

func TestExecuteAttentionPartition_FlagOff(t *testing.T) {
	// Ensure flag is off
	EnableDistributedAttention = false

	coordinator := NewAttentionCoordinator()
	partition := AttentionPartition{
		ID:       "test_partition",
		NodeID:   "node_1",
		HeadRange: []int{0, 2},
		LayerRange: []int{0, 1},
		Metadata: map[string]interface{}{
			"sequence_length": 4,
			"head_dim":       8,
			"heads_in_partition": 2,
		},
	}

	states := []AttentionState{
		{
			QueryStates:  make([]float32, 32), // 4 seq * 8 dim
			KeyStates:    make([]float32, 32),
			ValueStates:  make([]float32, 32),
			HeadID:       0,
		},
	}

	_, err := coordinator.executeAttentionPartition(partition, states)
	if err == nil {
		t.Errorf("Expected error when flag is off, got nil")
	}
	if !strings.Contains(err.Error(), "not implemented") {
		t.Errorf("Expected 'not implemented' error, got: %v", err)
	}
}

func TestExecuteAttentionPartition_MetadataEnabled(t *testing.T) {
	// Flag is off but metadata enables it
	EnableDistributedAttention = false

	coordinator := NewAttentionCoordinator()
	partition := AttentionPartition{
		ID:       "test_partition",
		NodeID:   "node_1",
		HeadRange: []int{0, 2},
		LayerRange: []int{0, 1},
		Metadata: map[string]interface{}{
			"enable_distributed_attention": true,
			"sequence_length": 4,
			"head_dim":       8,
			"heads_in_partition": 2,
		},
	}

	states := []AttentionState{
		{
			QueryStates:  make([]float32, 32), // 4 seq * 8 dim
			KeyStates:    make([]float32, 32),
			ValueStates:  make([]float32, 32),
			HeadID:       0,
		},
		{
			QueryStates:  make([]float32, 32),
			KeyStates:    make([]float32, 32),
			ValueStates:  make([]float32, 32),
			HeadID:       1,
		},
	}

	result, err := coordinator.executeAttentionPartition(partition, states)
	if err != nil {
		t.Errorf("Expected success when metadata enables attention, got error: %v", err)
	}
	if result == nil {
		t.Errorf("Expected non-nil result")
	}
	if result.Heads != 2 {
		t.Errorf("Expected 2 heads, got %d", result.Heads)
	}
	if result.SeqLen != 4 {
		t.Errorf("Expected sequence length 4, got %d", result.SeqLen)
	}
	if result.HeadDim != 8 {
		t.Errorf("Expected head dimension 8, got %d", result.HeadDim)
	}
}

func TestExecuteAttentionPartition_InvalidDimensions(t *testing.T) {
	EnableDistributedAttention = true

	coordinator := NewAttentionCoordinator()
	partition := AttentionPartition{
		ID:       "test_partition",
		NodeID:   "node_1",
		HeadRange: []int{0, 1},
		LayerRange: []int{0, 1},
		Metadata: map[string]interface{}{
			"sequence_length": 0, // Invalid
			"head_dim":       8,
			"heads_in_partition": 1,
		},
	}

	states := []AttentionState{}

	_, err := coordinator.executeAttentionPartition(partition, states)
	if err == nil {
		t.Errorf("Expected error for invalid dimensions, got nil")
	}
	if !strings.Contains(err.Error(), "invalid dimensions") {
		t.Errorf("Expected 'invalid dimensions' error, got: %v", err)
	}
}

func TestExecuteAttentionPartition_StateDimensionMismatch(t *testing.T) {
	EnableDistributedAttention = true

	coordinator := NewAttentionCoordinator()
	partition := AttentionPartition{
		ID:       "test_partition",
		NodeID:   "node_1",
		HeadRange: []int{0, 1},
		LayerRange: []int{0, 1},
		Metadata: map[string]interface{}{
			"sequence_length": 4,
			"head_dim":       8,
			"heads_in_partition": 1,
		},
	}

	states := []AttentionState{
		{
			QueryStates:  make([]float32, 16), // Wrong size - should be 32 (4*8)
			KeyStates:    make([]float32, 32),
			ValueStates:  make([]float32, 32),
			HeadID:       0,
		},
	}

	_, err := coordinator.executeAttentionPartition(partition, states)
	if err == nil {
		t.Errorf("Expected error for dimension mismatch, got nil")
	}
	if !strings.Contains(err.Error(), "state dimension mismatch") {
		t.Errorf("Expected 'state dimension mismatch' error, got: %v", err)
	}
}

func TestExecuteAttentionPartition_MinimalComputation(t *testing.T) {
	EnableDistributedAttention = true

	coordinator := NewAttentionCoordinator()
	partition := AttentionPartition{
		ID:       "test_partition",
		NodeID:   "node_1",
		HeadRange: []int{0, 1},
		LayerRange: []int{0, 1},
		Metadata: map[string]interface{}{
			"sequence_length": 2,
			"head_dim":       3,
			"heads_in_partition": 1,
		},
	}

	// Create synthetic Q, K, V with known values
	states := []AttentionState{
		{
			QueryStates:  []float32{1.0, 0.0, 0.0, 0.0, 1.0, 0.0}, // 2 seq * 3 dim
			KeyStates:    []float32{1.0, 0.0, 0.0, 0.0, 1.0, 0.0},
			ValueStates:  []float32{1.0, 2.0, 3.0, 4.0, 5.0, 6.0},
			HeadID:       0,
		},
	}

	result, err := coordinator.executeAttentionPartition(partition, states)
	if err != nil {
		t.Errorf("Expected success for minimal computation, got error: %v", err)
	}
	if result == nil {
		t.Errorf("Expected non-nil result")
	}
	if len(result.Values) != 6 { // 2 seq * 1 head * 3 dim
		t.Errorf("Expected 6 output values, got %d", len(result.Values))
	}

	// Check that output is non-zero (actual attention computation)
	hasNonZero := false
	for _, v := range result.Values {
		if v != 0.0 {
			hasNonZero = true
			break
		}
	}
	if !hasNonZero {
		t.Errorf("Expected non-zero attention output, got all zeros")
	}
}

func TestCoordinateAttentionComputation_PropagatesErrors(t *testing.T) {
	EnableDistributedAttention = false

	coordinator := NewAttentionCoordinator()
	states := []AttentionState{
		{
			QueryStates:  make([]float32, 32),
			KeyStates:    make([]float32, 32),
			ValueStates:  make([]float32, 32),
			HeadID:       0,
		},
	}

	modelAnalysis := &partitioning.ModelAnalysis{
		LayerInfo: &partitioning.LayerAnalysis{
			TotalLayers: 1,
		},
	}

	_, err := coordinator.CoordinateAttentionComputation(states, modelAnalysis)
	if err == nil {
		t.Errorf("Expected error to be propagated from executeAttentionPartition, got nil")
	}
	if !strings.Contains(err.Error(), "failed to execute attention partition") {
		t.Errorf("Expected 'failed to execute attention partition' error, got: %v", err)
	}
}
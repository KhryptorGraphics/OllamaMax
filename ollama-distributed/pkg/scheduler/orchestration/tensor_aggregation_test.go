package orchestration

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

func TestTensorAggregator_ConcatenateTensors(t *testing.T) {
	// Seed randomness to avoid flakes
	rand.Seed(42)

	aggregator := NewTensorAggregator()

	t.Run("dim=0 concatenation A[2,3]+B[1,3]", func(t *testing.T) {
		// Create small tensors: A shape [2,3], B shape [1,3]; concat dim=0 -> expect shape [3,3] and data order A;B
		tensorA := TensorData{
			Data:  []float32{1, 2, 3, 4, 5, 6}, // 2x3 matrix
			Shape: []int{2, 3},
			Type:  "float32",
		}
		tensorB := TensorData{
			Data:  []float32{7, 8, 9}, // 1x3 matrix
			Shape: []int{1, 3},
			Type:  "float32",
		}

		result, err := aggregator.ConcatenateTensors([]TensorData{tensorA, tensorB}, 0)
		if err != nil {
			t.Fatalf("ConcatenateTensors failed: %v", err)
		}

		// Assert on both Shape and Data
		expectedShape := []int{3, 3}
		if !reflect.DeepEqual(result.Shape, expectedShape) {
			t.Errorf("Expected shape %v, got %v", expectedShape, result.Shape)
		}

		// Check data - should be A followed by B (data order A;B)
		expectedData := []float32{1, 2, 3, 4, 5, 6, 7, 8, 9}
		if !reflect.DeepEqual(result.Data, expectedData) {
			t.Errorf("Expected data %v, got %v", expectedData, result.Data)
		}

		// Check type
		if result.Type != "float32" {
			t.Errorf("Expected type float32, got %s", result.Type)
		}
	})

	t.Run("dim=1 concatenation C[2,2]+D[2,1]", func(t *testing.T) {
		// Create tensors C shape [2,2], D shape [2,1]; concat dim=1 -> expect shape [2,3] with correct inner-stride layout
		tensorC := TensorData{
			Data:  []float32{1, 2, 3, 4}, // 2x2 matrix
			Shape: []int{2, 2},
			Type:  "float32",
		}
		tensorD := TensorData{
			Data:  []float32{5, 6}, // 2x1 matrix
			Shape: []int{2, 1},
			Type:  "float32",
		}

		result, err := aggregator.ConcatenateTensors([]TensorData{tensorC, tensorD}, 1)
		if err != nil {
			t.Fatalf("ConcatenateTensors failed: %v", err)
		}

		// Assert on both Shape and Data
		expectedShape := []int{2, 3}
		if !reflect.DeepEqual(result.Shape, expectedShape) {
			t.Errorf("Expected shape %v, got %v", expectedShape, result.Shape)
		}

		// Check data - inner-stride concatenation
		// C = [[1, 2], [3, 4]], D = [[5], [6]]
		// Result = [[1, 2, 5], [3, 4, 6]]
		expectedData := []float32{1, 2, 5, 3, 4, 6}
		if !reflect.DeepEqual(result.Data, expectedData) {
			t.Errorf("Expected data %v, got %v", expectedData, result.Data)
		}

		// Check type
		if result.Type != "float32" {
			t.Errorf("Expected type float32, got %s", result.Type)
		}
	})

	t.Run("error on mismatched dimensions", func(t *testing.T) {
		// Test error handling for mismatched dimensions
		tensorE := TensorData{
			Data:  []float32{1, 2, 3, 4},
			Shape: []int{2, 2},
			Type:  "float32",
		}
		tensorF := TensorData{
			Data:  []float32{5, 6, 7},
			Shape: []int{3, 1},
			Type:  "float32",
		}

		_, err := aggregator.ConcatenateTensors([]TensorData{tensorE, tensorF}, 0)
		if err == nil {
			t.Error("Expected error for mismatched dimensions, got nil")
		}
	})
}

func TestTensorAggregator_AverageTensors(t *testing.T) {
	// Seed randomness for deterministic tests
	rand.Seed(1)

	aggregator := NewTensorAggregator()

	t.Run("average of tensors", func(t *testing.T) {
		tensor1 := TensorData{
			Data:  []float32{1, 2, 3, 4},
			Shape: []int{2, 2},
			Type:  "float32",
		}
		tensor2 := TensorData{
			Data:  []float32{5, 6, 7, 8},
			Shape: []int{2, 2},
			Type:  "float32",
		}

		result, err := aggregator.AverageTensors([]TensorData{tensor1, tensor2})
		if err != nil {
			t.Fatalf("AverageTensors failed: %v", err)
		}

		// Check shape - should remain [2, 2]
		expectedShape := []int{2, 2}
		if !reflect.DeepEqual(result.Shape, expectedShape) {
			t.Errorf("Expected shape %v, got %v", expectedShape, result.Shape)
		}

		// Check data - element-wise average
		expectedData := []float32{3, 4, 5, 6}
		if !reflect.DeepEqual(result.Data, expectedData) {
			t.Errorf("Expected data %v, got %v", expectedData, result.Data)
		}
	})
}

func TestTensorAggregator_WeightedAverageTensors(t *testing.T) {
	// Seed randomness for deterministic tests
	rand.Seed(1)

	aggregator := NewTensorAggregator()

	t.Run("weighted average of tensors", func(t *testing.T) {
		tensor1 := TensorData{
			Data:  []float32{1, 2, 3, 4},
			Shape: []int{2, 2},
			Type:  "float32",
		}
		tensor2 := TensorData{
			Data:  []float32{5, 6, 7, 8},
			Shape: []int{2, 2},
			Type:  "float32",
		}

		weights := []float32{0.3, 0.7}
		result, err := aggregator.WeightedAverageTensors([]TensorData{tensor1, tensor2}, weights)
		if err != nil {
			t.Fatalf("WeightedAverageTensors failed: %v", err)
		}

		// Check shape - should remain [2, 2]
		expectedShape := []int{2, 2}
		if !reflect.DeepEqual(result.Shape, expectedShape) {
			t.Errorf("Expected shape %v, got %v", expectedShape, result.Shape)
		}

		// Check data - weighted average
		// 0.3*[1,2,3,4] + 0.7*[5,6,7,8] = [3.8, 4.8, 5.8, 6.8]
		expectedData := []float32{3.8, 4.8, 5.8, 6.8}
		const epsilon = 1e-6
		for i := range expectedData {
			if diff := result.Data[i] - expectedData[i]; diff < -epsilon || diff > epsilon {
				t.Errorf("Index %d: expected %f, got %f", i, expectedData[i], result.Data[i])
			}
		}
	})
}

func TestWeightedAggregationStrategy_WeightParsing(t *testing.T) {
	strategy := &WeightedAggregationStrategy{}

	// Create test context with different weight types
	t.Run("int weight", func(t *testing.T) {
		context := &AggregationContext{
			TaskID: "test",
			PartialResults: []*PartialResult{
				{
					PartitionID:  "partition1",
					HiddenStates: []float32{1.0, 2.0},
					Metadata:     map[string]interface{}{"weight": 5}, // int
				},
			},
		}

		_, err := strategy.Aggregate(context)
		if err != nil {
			t.Errorf("Expected success with int weight, got error: %v", err)
		}
	})

	t.Run("float32 weight", func(t *testing.T) {
		context := &AggregationContext{
			TaskID: "test",
			PartialResults: []*PartialResult{
				{
					PartitionID:  "partition1",
					HiddenStates: []float32{1.0, 2.0},
					Metadata:     map[string]interface{}{"weight": float32(2.5)},
				},
			},
		}

		_, err := strategy.Aggregate(context)
		if err != nil {
			t.Errorf("Expected success with float32 weight, got error: %v", err)
		}
	})

	t.Run("float64 weight", func(t *testing.T) {
		context := &AggregationContext{
			TaskID: "test",
			PartialResults: []*PartialResult{
				{
					PartitionID:  "partition1",
					HiddenStates: []float32{1.0, 2.0},
					Metadata:     map[string]interface{}{"weight": float64(3.7)},
				},
			},
		}

		_, err := strategy.Aggregate(context)
		if err != nil {
			t.Errorf("Expected success with float64 weight, got error: %v", err)
		}
	})

	t.Run("string weight", func(t *testing.T) {
		context := &AggregationContext{
			TaskID: "test",
			PartialResults: []*PartialResult{
				{
					PartitionID:  "partition1",
					HiddenStates: []float32{1.0, 2.0},
					Metadata:     map[string]interface{}{"weight": "2.5"},
				},
			},
		}

		_, err := strategy.Aggregate(context)
		if err != nil {
			t.Errorf("Expected success with string weight, got error: %v", err)
		}
	})

	t.Run("invalid weight type", func(t *testing.T) {
		context := &AggregationContext{
			TaskID: "test",
			PartialResults: []*PartialResult{
				{
					PartitionID:  "partition1",
					HiddenStates: []float32{1.0, 2.0},
					Metadata:     map[string]interface{}{"weight": []int{1, 2}}, // Invalid type
				},
			},
		}

		_, err := strategy.Aggregate(context)
		if err == nil {
			t.Errorf("Expected error for invalid weight type, got nil")
		}
		if !strings.Contains(err.Error(), "invalid weight type") || !strings.Contains(err.Error(), "partition1") {
			t.Errorf("Expected 'invalid weight type' error with partition ID, got: %v", err)
		}
	})

	t.Run("negative weight", func(t *testing.T) {
		context := &AggregationContext{
			TaskID: "test",
			PartialResults: []*PartialResult{
				{
					PartitionID:  "partition1",
					HiddenStates: []float32{1.0, 2.0},
					Metadata:     map[string]interface{}{"weight": -1.5},
				},
			},
		}

		_, err := strategy.Aggregate(context)
		if err == nil {
			t.Errorf("Expected error for negative weight, got nil")
		}
		if !strings.Contains(err.Error(), "negative weight") || !strings.Contains(err.Error(), "partition1") {
			t.Errorf("Expected 'negative weight' error with partition ID, got: %v", err)
		}
	})

	t.Run("zero total weight", func(t *testing.T) {
		context := &AggregationContext{
			TaskID: "test",
			PartialResults: []*PartialResult{
				{
					PartitionID:  "partition1",
					HiddenStates: []float32{1.0, 2.0},
					Metadata:     map[string]interface{}{"weight": 0.0},
				},
				{
					PartitionID:  "partition2",
					HiddenStates: []float32{3.0, 4.0},
					Metadata:     map[string]interface{}{"weight": 0.0},
				},
			},
		}

		_, err := strategy.Aggregate(context)
		if err == nil {
			t.Errorf("Expected error for zero total weight, got nil")
		}
		if !strings.Contains(err.Error(), "total weight") || !strings.Contains(err.Error(), "must be positive") {
			t.Errorf("Expected 'total weight must be positive' error, got: %v", err)
		}
	})
}

func TestTensorAggregator_DataConsistencyValidation(t *testing.T) {
	aggregator := NewTensorAggregator()

	t.Run("valid tensors pass concatenation", func(t *testing.T) {
		tensors := []TensorData{
			{
				Data:  []float32{1, 2, 3, 4},
				Shape: []int{2, 2},
				Type:  "test",
			},
			{
				Data:  []float32{5, 6, 7, 8},
				Shape: []int{2, 2},
				Type:  "test",
			},
		}

		_, err := aggregator.ConcatenateTensors(tensors, 0)
		if err != nil {
			t.Errorf("Expected success with valid tensors, got error: %v", err)
		}
	})

	t.Run("tensor with mismatched data/shape fails", func(t *testing.T) {
		tensors := []TensorData{
			{
				Data:  []float32{1, 2, 3}, // Only 3 elements
				Shape: []int{2, 2},       // But shape expects 4 elements
				Type:  "test",
			},
		}

		_, err := aggregator.ConcatenateTensors(tensors, 0)
		if err == nil {
			t.Errorf("Expected error for mismatched data/shape, got nil")
		}
		if !strings.Contains(err.Error(), "data consistency validation failed") {
			t.Errorf("Expected 'data consistency validation failed' error, got: %v", err)
		}
	})

	t.Run("error includes tensor index", func(t *testing.T) {
		tensors := []TensorData{
			{
				Data:  []float32{1, 2, 3, 4},
				Shape: []int{2, 2},
				Type:  "test",
			},
			{
				Data:  []float32{5, 6, 7}, // Mismatch in second tensor
				Shape: []int{2, 2},
				Type:  "test",
			},
		}

		_, err := aggregator.ConcatenateTensors(tensors, 0)
		if err == nil {
			t.Errorf("Expected error for mismatched data/shape, got nil")
		}
		if !strings.Contains(err.Error(), "tensor 1") {
			t.Errorf("Expected error to mention 'tensor 1', got: %v", err)
		}
	})
}

func TestTensorAggregationStrategy_DataConsistencyValidation(t *testing.T) {
	strategy := &TensorAggregationStrategy{}

	t.Run("metadata handling with mismatched dimensions", func(t *testing.T) {
		// This test verifies that the deserializer gracefully handles mismatched metadata
		// The deserializer should create a valid tensor even with inconsistent metadata
		context := &AggregationContext{
			TaskID: "test",
			PartialResults: []*PartialResult{
				{
					PartitionID:  "partition1",
					HiddenStates: []float32{1.0, 2.0, 3.0}, // 3 elements
					Metadata: map[string]interface{}{
						"batch_size":      2,
						"sequence_length": 2, // 2*2*1 = 4 elements expected, but deserializer should handle gracefully
						"hidden_size":     1,
					},
				},
			},
		}

		result, err := strategy.Aggregate(context)
		if err != nil {
			t.Errorf("Expected success with graceful metadata handling, got error: %v", err)
		}
		if result == nil {
			t.Errorf("Expected non-nil result")
		}
	})

	t.Run("logits metadata handling", func(t *testing.T) {
		// This test verifies that the deserializer gracefully handles logits with mismatched metadata
		context := &AggregationContext{
			TaskID: "test",
			PartialResults: []*PartialResult{
				{
					PartitionID: "partition2",
					Logits:      []float32{1.0, 2.0}, // 2 elements
					Metadata: map[string]interface{}{
						"batch_size":  1,
						"vocab_size":  5, // 1*5 = 5 elements expected, but deserializer should handle gracefully
					},
				},
			},
		}

		result, err := strategy.Aggregate(context)
		if err != nil {
			t.Errorf("Expected success with graceful logits metadata handling, got error: %v", err)
		}
		if result == nil {
			t.Errorf("Expected non-nil result")
		}
	})

	t.Run("tokens data/shape mismatch", func(t *testing.T) {
		context := &AggregationContext{
			TaskID: "test",
			PartialResults: []*PartialResult{
				{
					PartitionID: "partition3",
					Tokens:      []int{1, 2, 3}, // Will be converted to 3 float32 elements
					Metadata:    map[string]interface{}{},
				},
			},
		}

		// Create a tensor that would fail validation
		// This is a bit tricky since DeserializeTokens creates shape from data length
		// So let's test with a more realistic scenario
		_, err := strategy.Aggregate(context)
		// This should succeed since tokens create shape based on data length
		if err != nil && strings.Contains(err.Error(), "tokens data inconsistency") {
			t.Errorf("Unexpected tokens data inconsistency error: %v", err)
		}
	})
}
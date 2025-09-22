package orchestration

import (
	"fmt"
	"math"
	"time"
)

// TensorAggregationStrategy handles tensor-based aggregation for ML model outputs
type TensorAggregationStrategy struct {
	name string
}

func (tas *TensorAggregationStrategy) GetName() string {
	return "tensor"
}

func (tas *TensorAggregationStrategy) Aggregate(context *AggregationContext) (*AggregatedResponse, error) {
	start := time.Now()

	aggregator := NewTensorAggregator()
	deserializer := NewTensorDeserializer()
	validator := NewTensorValidator()

	// Separate tensor types for proper aggregation
	hiddenStatesTensors := make([]TensorData, 0)
	logitsTensors := make([]TensorData, 0)
	tokensTensors := make([]TensorData, 0)

	// Partition PartialResults by tensor type
	for _, partial := range context.PartialResults {
		if partial.Error == "" {
			if partial.HiddenStates != nil {
				tensor := deserializer.DeserializeHiddenStates(partial.HiddenStates, partial.Metadata)
				// Validate data consistency
				if err := validator.ValidateDataConsistency(tensor); err != nil {
					return nil, fmt.Errorf("hidden states data inconsistency in partition %s: %v", partial.PartitionID, err)
				}
				hiddenStatesTensors = append(hiddenStatesTensors, tensor)
			}
			if partial.Logits != nil {
				tensor := deserializer.DeserializeLogits(partial.Logits, partial.Metadata)
				// Validate data consistency
				if err := validator.ValidateDataConsistency(tensor); err != nil {
					return nil, fmt.Errorf("logits data inconsistency in partition %s: %v", partial.PartitionID, err)
				}
				logitsTensors = append(logitsTensors, tensor)
			}
			if partial.Tokens != nil {
				tensor := deserializer.DeserializeTokens(partial.Tokens)
				// Validate data consistency
				if err := validator.ValidateDataConsistency(tensor); err != nil {
					return nil, fmt.Errorf("tokens data inconsistency in partition %s: %v", partial.PartitionID, err)
				}
				tokensTensors = append(tokensTensors, tensor)
			}
		}
	}

	// Aggregate each type separately
	aggregatedData := make(map[string]interface{})
	shapes := make(map[string][]int)

	if len(hiddenStatesTensors) > 0 {
		result, err := aggregator.ConcatenateTensors(hiddenStatesTensors, 0)
		if err != nil {
			return nil, fmt.Errorf("hidden states aggregation failed: %v", err)
		}
		aggregatedData["hidden_states"] = result
		shapes["hidden_states"] = result.Shape
	}

	if len(logitsTensors) > 0 {
		result, err := aggregator.ConcatenateTensors(logitsTensors, 0)
		if err != nil {
			return nil, fmt.Errorf("logits aggregation failed: %v", err)
		}
		aggregatedData["logits"] = result
		shapes["logits"] = result.Shape
	}

	if len(tokensTensors) > 0 {
		result, err := aggregator.ConcatenateTensors(tokensTensors, 0)
		if err != nil {
			return nil, fmt.Errorf("tokens aggregation failed: %v", err)
		}
		aggregatedData["tokens"] = result
		shapes["tokens"] = result.Shape
	}

	return &AggregatedResponse{
		TaskID:   context.TaskID,
		Strategy: tas.GetName(),
		Data:     aggregatedData,
		Metadata: map[string]interface{}{
			"hidden_states_count": len(hiddenStatesTensors),
			"logits_count":        len(logitsTensors),
			"tokens_count":        len(tokensTensors),
			"total_partitions":    len(context.PartialResults),
			"shapes":              shapes,
		},
		Latency:   time.Since(start),
		Timestamp: time.Now(),
	}, nil
}

// AttentionAggregationStrategy handles multi-head attention aggregation
type AttentionAggregationStrategy struct {
	name string
	coordinator *AttentionCoordinator
}

func NewAttentionAggregationStrategy() *AttentionAggregationStrategy {
	return &AttentionAggregationStrategy{
		name:        "attention",
		coordinator: NewAttentionCoordinator(),
	}
}

func (aas *AttentionAggregationStrategy) GetName() string {
	return "attention"
}

func (aas *AttentionAggregationStrategy) Aggregate(context *AggregationContext) (*AggregatedResponse, error) {
	start := time.Now()

	// Extract attention data from partial results
	attentionOutputs := make([]AttentionOutput, 0)
	for _, partial := range context.PartialResults {
		if partial.Error == "" && partial.HiddenStates != nil {
			// Extract dimensions from metadata
			heads := extractHeadCount(partial.Metadata)
			seqLen := extractSequenceLength(partial.Metadata)
			headDim := GetIntFromMetadata(partial.Metadata, "head_dim")

			// If HeadDim not in metadata, try to infer it
			if headDim == 0 && heads > 0 && seqLen > 0 {
				totalElements := len(partial.HiddenStates)
				// HiddenStates should be heads * seqLen * headDim
				if totalElements > 0 && totalElements%(heads*seqLen) == 0 {
					headDim = totalElements / (heads * seqLen)
				}
			}

			// Validate dimensions
			if heads > 0 && seqLen > 0 && headDim > 0 {
				if heads*seqLen*headDim != len(partial.HiddenStates) {
					return nil, fmt.Errorf("dimension mismatch: heads(%d) * seqLen(%d) * headDim(%d) = %d, but got %d values",
						heads, seqLen, headDim, heads*seqLen*headDim, len(partial.HiddenStates))
				}
			}

			// Convert hidden states to attention output format
			attentionOutput := AttentionOutput{
				Values:      partial.HiddenStates,
				Heads:       heads,
				SeqLen:      seqLen,
				HeadDim:     headDim,
				LayerID:     GetIntFromMetadata(partial.Metadata, "layer_id"),
				PartitionID: partial.PartitionID,
				Metadata:    partial.Metadata,
			}
			attentionOutputs = append(attentionOutputs, attentionOutput)
		}
	}

	// Coordinate attention aggregation
	result, err := aas.coordinator.AggregateAttentionHeads(attentionOutputs)
	if err != nil {
		return nil, fmt.Errorf("attention aggregation failed: %v", err)
	}

	return &AggregatedResponse{
		TaskID:   context.TaskID,
		Strategy: aas.GetName(),
		Data:     result,
		Metadata: map[string]interface{}{
			"attention_heads":  len(attentionOutputs),
			"total_partitions": len(context.PartialResults),
		},
		Latency:   time.Since(start),
		Timestamp: time.Now(),
	}, nil
}

// EmbeddingAggregationStrategy handles embedding output aggregation
type EmbeddingAggregationStrategy struct {
	name string
}

func (eas *EmbeddingAggregationStrategy) GetName() string {
	return "embedding"
}

func (eas *EmbeddingAggregationStrategy) Aggregate(context *AggregationContext) (*AggregatedResponse, error) {
	start := time.Now()

	// Use deserializer to get proper shape information
	aggregator := NewTensorAggregator()
	deserializer := NewTensorDeserializer()
	validator := NewTensorValidator()

	tensorData := make([]TensorData, 0)
	for _, partial := range context.PartialResults {
		if partial.Error == "" && partial.HiddenStates != nil {
			// Use deserializer for proper shape handling
			tensor := deserializer.DeserializeHiddenStates(partial.HiddenStates, partial.Metadata)
			tensor.Type = "embedding"

			// Validate data consistency
			if err := validator.ValidateDataConsistency(tensor); err != nil {
				return nil, fmt.Errorf("embedding data inconsistency in partition %s: %v", partial.PartitionID, err)
			}

			tensorData = append(tensorData, tensor)
		}
	}

	if len(tensorData) == 0 {
		return nil, fmt.Errorf("no valid embeddings to aggregate")
	}

	// Use dim=1 only if tensors are 2D, otherwise use dim=0
	concatDim := 0
	if len(tensorData[0].Shape) > 1 {
		concatDim = 1
	}
	result, err := aggregator.ConcatenateTensors(tensorData, concatDim)
	if err != nil {
		return nil, fmt.Errorf("embedding aggregation failed: %v", err)
	}

	return &AggregatedResponse{
		TaskID:   context.TaskID,
		Strategy: eas.GetName(),
		Data:     result.Data,
		Metadata: map[string]interface{}{
			"embedding_count":  len(tensorData),
			"total_partitions": len(context.PartialResults),
			"embedding_dim":    result.Shape[len(result.Shape)-1],
		},
		Latency:   time.Since(start),
		Timestamp: time.Now(),
	}, nil
}

// LogitsAggregationStrategy handles logits and probability distribution aggregation
type LogitsAggregationStrategy struct {
	name string
}

func (las *LogitsAggregationStrategy) GetName() string {
	return "logits"
}

func (las *LogitsAggregationStrategy) Aggregate(context *AggregationContext) (*AggregatedResponse, error) {
	start := time.Now()
	validator := NewTensorValidator()
	aggregator := NewTensorAggregator()

	logitsTensors := make([]TensorData, 0)
	for _, partial := range context.PartialResults {
		if partial.Error == "" && partial.Logits != nil {
			tensor := TensorData{
				Shape: extractShapeFromMetadata(partial.Metadata, len(partial.Logits)),
				Data:  partial.Logits,
				Type:  "logits",
			}

			// Validate data consistency
			if err := validator.ValidateDataConsistency(tensor); err != nil {
				return nil, fmt.Errorf("logits data inconsistency in partition %s: %v", partial.PartitionID, err)
			}

			logitsTensors = append(logitsTensors, tensor)
		}
	}

	if len(logitsTensors) == 0 {
		return nil, fmt.Errorf("no valid logits to aggregate")
	}

	// Determine aggregation strategy based on shape compatibility
	var aggregatedLogits []float32
	var finalShape []int
	var aggregationType string

	// Check if we have vocab-sharded logits (same rank, different last dimension)
	if las.isVocabSharded(logitsTensors) {
		// Concatenate along vocab dimension
		aggregationType = "vocab_concatenation"
		concatenated, err := aggregator.ConcatenateTensors(logitsTensors, len(logitsTensors[0].Shape)-1)
		if err != nil {
			return nil, fmt.Errorf("failed to concatenate vocab-sharded logits: %v", err)
		}
		aggregatedLogits = concatenated.Data
		finalShape = concatenated.Shape
	} else if las.areShapesIdentical(logitsTensors) {
		// Average logits across partitions for probability distribution
		aggregationType = "averaging"
		logitSize := len(logitsTensors[0].Data)
		aggregatedLogits = make([]float32, logitSize)

		for i := range aggregatedLogits {
			var sum float32
			for _, tensor := range logitsTensors {
				if i < len(tensor.Data) {
					sum += tensor.Data[i]
				}
			}
			aggregatedLogits[i] = sum / float32(len(logitsTensors))
		}
		finalShape = logitsTensors[0].Shape
	} else {
		return nil, fmt.Errorf("incompatible logits shapes for aggregation: cannot average or concatenate")
	}

	// Apply softmax for probability distribution
	probabilities := las.applySoftmaxWithShape(aggregatedLogits, finalShape)

	return &AggregatedResponse{
		TaskID:   context.TaskID,
		Strategy: las.GetName(),
		Data: map[string]interface{}{
			"logits":        aggregatedLogits,
			"probabilities": probabilities,
		},
		Metadata: map[string]interface{}{
			"logits_count":     len(logitsTensors),
			"total_partitions": len(context.PartialResults),
			"vocab_size":       len(aggregatedLogits),
			"output_shape":     finalShape,
			"aggregation_type": aggregationType,
		},
		Latency:   time.Since(start),
		Timestamp: time.Now(),
	}, nil
}

// isVocabSharded checks if logits are vocab-dimension sharded
func (las *LogitsAggregationStrategy) isVocabSharded(tensors []TensorData) bool {
	if len(tensors) < 2 {
		return false
	}

	firstRank := len(tensors[0].Shape)
	if firstRank == 0 {
		return false
	}

	// Check if all tensors have same rank and all dimensions equal except last
	for i := 1; i < len(tensors); i++ {
		if len(tensors[i].Shape) != firstRank {
			return false
		}

		// Check all dimensions except last
		for dim := 0; dim < firstRank-1; dim++ {
			if tensors[0].Shape[dim] != tensors[i].Shape[dim] {
				return false
			}
		}
	}

	// Check if last dimensions are different (indicating sharding)
	for i := 1; i < len(tensors); i++ {
		if tensors[0].Shape[firstRank-1] == tensors[i].Shape[firstRank-1] {
			return false // Same vocab size means not sharded
		}
	}

	return true
}

// areShapesIdentical checks if all tensors have identical shapes
func (las *LogitsAggregationStrategy) areShapesIdentical(tensors []TensorData) bool {
	if len(tensors) < 2 {
		return true
	}

	validator := NewTensorValidator()
	for i := 1; i < len(tensors); i++ {
		if !validator.ShapesEqual(tensors[0].Shape, tensors[i].Shape) {
			return false
		}
	}
	return true
}

// applySoftmaxWithShape applies softmax considering tensor shape
func (las *LogitsAggregationStrategy) applySoftmaxWithShape(logits []float32, shape []int) []float32 {
	if len(shape) == 1 {
		// 1D tensor - simple softmax
		return applySoftmax(logits)
	} else if len(shape) == 2 {
		// 2D tensor [batch, vocab] - apply softmax per batch
		batchSize := shape[0]
		vocabSize := shape[1]
		result := make([]float32, len(logits))

		for b := 0; b < batchSize; b++ {
			start := b * vocabSize
			end := start + vocabSize
			batchLogits := logits[start:end]
			batchProbs := applySoftmax(batchLogits)
			copy(result[start:end], batchProbs)
		}
		return result
	} else if len(shape) == 3 {
		// 3D tensor [batch, seq, vocab] - apply softmax per sequence position
		batchSize := shape[0]
		seqLen := shape[1]
		vocabSize := shape[2]
		result := make([]float32, len(logits))

		for b := 0; b < batchSize; b++ {
			for s := 0; s < seqLen; s++ {
				start := (b*seqLen+s)*vocabSize
				end := start + vocabSize
				posLogits := logits[start:end]
				posProbs := applySoftmax(posLogits)
				copy(result[start:end], posProbs)
			}
		}
		return result
	}

	// Fallback to simple softmax for unknown shapes
	return applySoftmax(logits)
}

// HiddenStateAggregationStrategy handles intermediate layer output aggregation
type HiddenStateAggregationStrategy struct {
	name string
}

func (hsas *HiddenStateAggregationStrategy) GetName() string {
	return "hidden_states"
}

func (hsas *HiddenStateAggregationStrategy) Aggregate(context *AggregationContext) (*AggregatedResponse, error) {
	start := time.Now()

	// Use deserializer to get proper shape information
	aggregator := NewTensorAggregator()
	deserializer := NewTensorDeserializer()
	validator := NewTensorValidator()

	tensorData := make([]TensorData, 0)
	sequenceInfo := make([]map[string]interface{}, 0)

	for _, partial := range context.PartialResults {
		if partial.Error == "" && partial.HiddenStates != nil {
			// Use deserializer for proper shape handling
			tensor := deserializer.DeserializeHiddenStates(partial.HiddenStates, partial.Metadata)
			tensor.Type = "hidden_state"

			// Validate data consistency
			if err := validator.ValidateDataConsistency(tensor); err != nil {
				return nil, fmt.Errorf("hidden state data inconsistency in partition %s: %v", partial.PartitionID, err)
			}

			tensorData = append(tensorData, tensor)
			sequenceInfo = append(sequenceInfo, partial.Metadata)
		}
	}

	if len(tensorData) == 0 {
		return nil, fmt.Errorf("no valid hidden states to aggregate")
	}

	// Use dim=1 only if tensors are 2D, otherwise use dim=0
	concatDim := 0
	if len(tensorData[0].Shape) > 1 {
		concatDim = 1
	}
	result, err := aggregator.ConcatenateTensors(tensorData, concatDim)
	if err != nil {
		return nil, fmt.Errorf("hidden state aggregation failed: %v", err)
	}

	return &AggregatedResponse{
		TaskID:   context.TaskID,
		Strategy: hsas.GetName(),
		Data:     result.Data,
		Metadata: map[string]interface{}{
			"hidden_state_count": len(tensorData),
			"total_partitions":   len(context.PartialResults),
			"output_shape":       result.Shape,
			"sequence_info":      sequenceInfo,
		},
		Latency:   time.Since(start),
		Timestamp: time.Now(),
	}, nil
}

// Helper functions
func extractShape(data []float32) []int {
	if data == nil {
		return []int{}
	}
	// Simple heuristic - assume 1D for now, can be enhanced based on metadata
	return []int{len(data)}
}

func extractHeadCount(metadata map[string]interface{}) int {
	if heads, ok := metadata["attention_heads"]; ok {
		headCount := ConvertToInt(heads)
		if headCount > 0 {
			return headCount
		}
	}
	return 8 // Default attention heads
}

func extractSequenceLength(metadata map[string]interface{}) int {
	if seqLen, ok := metadata["sequence_length"]; ok {
		length := ConvertToInt(seqLen)
		if length > 0 {
			return length
		}
	}
	return 512 // Default sequence length
}

func extractShapeFromMetadata(metadata map[string]interface{}, dataLength int) []int {
	// Try to extract shape information from metadata
	if batchSize, ok := metadata["batch_size"]; ok {
		if bs := ConvertToInt(batchSize); bs > 0 {
			if seqLen, ok := metadata["sequence_length"]; ok {
				if sl := ConvertToInt(seqLen); sl > 0 {
					hiddenSize := dataLength / (bs * sl)
					if hiddenSize > 0 && bs*sl*hiddenSize == dataLength {
						return []int{bs, sl, hiddenSize}
					}
				}
			}
			// Try batch only
			featureSize := dataLength / bs
			if featureSize > 0 && bs*featureSize == dataLength {
				return []int{bs, featureSize}
			}
		}
	}

	// Default to 1D shape if no metadata available
	return []int{dataLength}
}

// convertToInt removed - use ConvertToInt from tensor_util.go instead

func applySoftmax(logits []float32) []float32 {
	// Find max for numerical stability
	var max float32 = logits[0]
	for _, v := range logits[1:] {
		if v > max {
			max = v
		}
	}

	// Calculate exponentials and sum
	var sum float32
	exp := make([]float32, len(logits))
	for i, v := range logits {
		exp[i] = float32(math.Exp(float64(v - max)))
		sum += exp[i]
	}

	// Normalize to get probabilities
	probabilities := make([]float32, len(logits))
	for i, v := range exp {
		probabilities[i] = v / sum
	}

	return probabilities
}


// Legacy strategies for backward compatibility

// ConcatAggregationStrategy concatenates partial results (legacy)
type ConcatAggregationStrategy struct {
	name string
}

func (cas *ConcatAggregationStrategy) GetName() string {
	return "concat"
}

func (cas *ConcatAggregationStrategy) Aggregate(context *AggregationContext) (*AggregatedResponse, error) {
	// Use tensor aggregation for better handling
	tensorStrategy := &TensorAggregationStrategy{}
	return tensorStrategy.Aggregate(context)
}

// AverageAggregationStrategy averages partial results (legacy)
type AverageAggregationStrategy struct {
	name string
}

func (aas *AverageAggregationStrategy) GetName() string {
	return "average"
}

func (aas *AverageAggregationStrategy) Aggregate(context *AggregationContext) (*AggregatedResponse, error) {
	// For ML outputs, use logits aggregation which includes averaging
	logitsStrategy := &LogitsAggregationStrategy{}
	return logitsStrategy.Aggregate(context)
}

// WeightedAggregationStrategy performs weighted aggregation (legacy)
type WeightedAggregationStrategy struct {
	name string
}

func (was *WeightedAggregationStrategy) GetName() string {
	return "weighted"
}

func (was *WeightedAggregationStrategy) Aggregate(context *AggregationContext) (*AggregatedResponse, error) {
	start := time.Now()

	// Perform weighted tensor aggregation
	aggregator := NewTensorAggregator()

	tensors := make([]TensorData, 0)
	weights := make([]float32, 0)

	for _, partial := range context.PartialResults {
		if partial.Error == "" {
			weight := float32(1.0)
			if w, exists := partial.Metadata["weight"]; exists {
				var ok bool
				weight, ok = ToFloat32(w)
				if !ok {
					return nil, fmt.Errorf("invalid weight type for partition %s: expected numeric, got %T", partial.PartitionID, w)
				}
				if weight < 0 {
					return nil, fmt.Errorf("negative weight %f for partition %s: weights must be non-negative", weight, partial.PartitionID)
				}
			}

			if partial.HiddenStates != nil {
				tensor := TensorData{
					Shape: extractShape(partial.HiddenStates),
					Data:  partial.HiddenStates,
					Type:  "hidden_states",
				}
				tensors = append(tensors, tensor)
				weights = append(weights, weight)
			}
		}
	}

	// Validate total weight is positive
	var totalWeight float32
	for _, w := range weights {
		totalWeight += w
	}
	if totalWeight <= 0 {
		return nil, fmt.Errorf("total weight %f must be positive", totalWeight)
	}

	result, err := aggregator.WeightedAverageTensors(tensors, weights)
	if err != nil {
		return nil, fmt.Errorf("weighted aggregation failed: %v", err)
	}

	return &AggregatedResponse{
		TaskID:   context.TaskID,
		Strategy: was.GetName(),
		Data:     result,
		Metadata: map[string]interface{}{
			"tensor_count":     len(tensors),
			"total_partitions": len(context.PartialResults),
		},
		Latency:   time.Since(start),
		Timestamp: time.Now(),
	}, nil
}

// Partitioning strategies have been moved to the partitioning package
// See: pkg/scheduler/partitioning/strategies.go


package orchestration

import (
	"fmt"
	"math"
	"errors"
)

// TensorData represents tensor information extracted from PartialResult
type TensorData struct {
	Shape []int     // Tensor dimensions
	Data  []float32 // Raw tensor data
	Type  string    // Type of tensor (hidden_states, logits, embeddings)
}

// TensorMetadata contains information about tensor characteristics
type TensorMetadata struct {
	BatchSize    int
	SeqLength    int
	HiddenSize   int
	VocabSize    int
	AttentionHeads int
}

// TensorDeserializer converts PartialResult fields to standardized tensor format
type TensorDeserializer struct{}

func NewTensorDeserializer() *TensorDeserializer {
	return &TensorDeserializer{}
}

// DeserializeHiddenStates converts PartialResult.HiddenStates to TensorData
func (td *TensorDeserializer) DeserializeHiddenStates(hiddenStates []float32, metadata map[string]interface{}) TensorData {
	shape := extractTensorShapeFromMetadata(metadata, len(hiddenStates))
	return TensorData{
		Shape: shape,
		Data:  hiddenStates,
		Type:  "hidden_states",
	}
}

// DeserializeLogits converts PartialResult.Logits to TensorData with enhanced shape validation
func (td *TensorDeserializer) DeserializeLogits(logits []float32, metadata map[string]interface{}) TensorData {
	shape := []int{len(logits)} // Default to 1D logits vector

	// Enhanced shape computation with validation
	if bs := GetIntFromMetadata(metadata, "batch_size"); bs > 0 {
		if sl := GetIntFromMetadata(metadata, "sequence_length"); sl > 0 {
			if vs := GetIntFromMetadata(metadata, "vocab_size"); vs > 0 {
				if bs*sl*vs == len(logits) {
					shape = []int{bs, sl, vs}
				} else if bs*vs == len(logits) {
					shape = []int{bs, vs}
				}
			}
		} else if vs := len(logits) / bs; vs > 0 && bs*vs == len(logits) {
			shape = []int{bs, vs}
		}
	}

	return TensorData{
		Shape: shape,
		Data:  logits,
		Type:  "logits",
	}
}

// DeserializeTokens converts PartialResult.Tokens to TensorData
func (td *TensorDeserializer) DeserializeTokens(tokens []int) TensorData {
	// Convert int tokens to float32 for consistency
	data := make([]float32, len(tokens))
	for i, token := range tokens {
		data[i] = float32(token)
	}

	return TensorData{
		Shape: []int{len(tokens)},
		Data:  data,
		Type:  "tokens",
	}
}

// TensorAggregator handles mathematical operations for tensor combination
type TensorAggregator struct {
	validator *TensorValidator
}

func NewTensorAggregator() *TensorAggregator {
	return &TensorAggregator{
		validator: NewTensorValidator(),
	}
}

// ConcatenateTensors concatenates tensors along a specified dimension
func (ta *TensorAggregator) ConcatenateTensors(tensors []TensorData, dim int) (*TensorData, error) {
	if len(tensors) == 0 {
		return nil, errors.New("no tensors to concatenate")
	}

	// Validate data consistency for each tensor
	for i, tensor := range tensors {
		if err := ta.validator.ValidateDataConsistency(tensor); err != nil {
			return nil, fmt.Errorf("tensor %d data consistency validation failed: %v", i, err)
		}
	}

	// Validate tensor compatibility
	if err := ta.validator.ValidateCompatibility(tensors, dim); err != nil {
		return nil, fmt.Errorf("tensor compatibility validation failed: %v", err)
	}

	// Calculate output shape
	outputShape := make([]int, len(tensors[0].Shape))
	copy(outputShape, tensors[0].Shape)

	// Sum up the concatenation dimension
	for i := 1; i < len(tensors); i++ {
		outputShape[dim] += tensors[i].Shape[dim]
	}

	// Handle arbitrary dimension concatenation
	totalSize := calculateTensorSize(outputShape)
	outputData := make([]float32, totalSize)

	// For arbitrary dimension concatenation, we need to properly handle the layout
	if dim == 0 {
		// Simple concatenation along first dimension
		offset := 0
		for _, tensor := range tensors {
			copy(outputData[offset:offset+len(tensor.Data)], tensor.Data)
			offset += len(tensor.Data)
		}
	} else {
		// Complex concatenation along arbitrary dimension
		if err := ta.concatenateArbitraryDimension(outputData, tensors, dim); err != nil {
			return nil, fmt.Errorf("arbitrary dimension concatenation failed: %v", err)
		}
	}

	return &TensorData{
		Shape: outputShape,
		Data:  outputData,
		Type:  tensors[0].Type,
	}, nil
}

// ElementWiseAdd performs element-wise addition of tensors
func (ta *TensorAggregator) ElementWiseAdd(tensors []TensorData) (*TensorData, error) {
	if len(tensors) == 0 {
		return nil, errors.New("no tensors to add")
	}

	// Validate all tensors have same shape
	for i := 1; i < len(tensors); i++ {
		if !ta.validator.ShapesEqual(tensors[0].Shape, tensors[i].Shape) {
			return nil, errors.New("tensors must have same shape for element-wise addition")
		}
	}

	result := make([]float32, len(tensors[0].Data))
	copy(result, tensors[0].Data)

	// Add remaining tensors
	for i := 1; i < len(tensors); i++ {
		for j := range result {
			result[j] += tensors[i].Data[j]
		}
	}

	return &TensorData{
		Shape: tensors[0].Shape,
		Data:  result,
		Type:  tensors[0].Type,
	}, nil
}

// AverageTensors computes element-wise average of tensors
func (ta *TensorAggregator) AverageTensors(tensors []TensorData) (*TensorData, error) {
	sumResult, err := ta.ElementWiseAdd(tensors)
	if err != nil {
		return nil, err
	}

	// Divide by number of tensors
	count := float32(len(tensors))
	for i := range sumResult.Data {
		sumResult.Data[i] /= count
	}

	return sumResult, nil
}

// WeightedAverageTensors computes weighted average of tensors
func (ta *TensorAggregator) WeightedAverageTensors(tensors []TensorData, weights []float32) (*TensorData, error) {
	if len(tensors) != len(weights) {
		return nil, errors.New("number of tensors must match number of weights")
	}

	if len(tensors) == 0 {
		return nil, errors.New("no tensors to average")
	}

	// Validate all tensors have same shape
	for i := 1; i < len(tensors); i++ {
		if !ta.validator.ShapesEqual(tensors[0].Shape, tensors[i].Shape) {
			return nil, errors.New("tensors must have same shape for weighted averaging")
		}
	}

	// Calculate weighted sum
	result := make([]float32, len(tensors[0].Data))
	var totalWeight float32

	for i, tensor := range tensors {
		weight := weights[i]
		totalWeight += weight

		for j, val := range tensor.Data {
			result[j] += val * weight
		}
	}

	// Normalize by total weight
	if totalWeight == 0 {
		return nil, errors.New("total weight cannot be zero")
	}

	for i := range result {
		result[i] /= totalWeight
	}

	return &TensorData{
		Shape: tensors[0].Shape,
		Data:  result,
		Type:  tensors[0].Type,
	}, nil
}

// SequenceAssembly assembles sequence tokens from distributed partitions
func (ta *TensorAggregator) SequenceAssembly(tokenTensors []TensorData) (*TensorData, error) {
	if len(tokenTensors) == 0 {
		return nil, errors.New("no token tensors to assemble")
	}

	// Sort by sequence position if metadata is available
	// For now, just concatenate in order
	return ta.ConcatenateTensors(tokenTensors, 0)
}

// AttentionWeightAggregation aggregates attention weights with proper normalization
func (ta *TensorAggregator) AttentionWeightAggregation(attentionTensors []TensorData) (*TensorData, error) {
	if len(attentionTensors) == 0 {
		return nil, errors.New("no attention tensors to aggregate")
	}

	// Concatenate attention heads
	result, err := ta.ConcatenateTensors(attentionTensors, 1) // Concatenate along head dimension
	if err != nil {
		return nil, err
	}

	// Apply attention normalization (softmax across heads)
	result.Data = ta.normalizeAttentionWeights(result.Data)

	return result, nil
}

// TensorValidator ensures tensor compatibility and validates operations
type TensorValidator struct{}

func NewTensorValidator() *TensorValidator {
	return &TensorValidator{}
}

// ValidateCompatibility checks if tensors can be concatenated along specified dimension
func (tv *TensorValidator) ValidateCompatibility(tensors []TensorData, dim int) error {
	if len(tensors) == 0 {
		return errors.New("no tensors to validate")
	}

	refShape := tensors[0].Shape
	if dim >= len(refShape) || dim < 0 {
		return fmt.Errorf("invalid concatenation dimension %d for shape %v", dim, refShape)
	}

	for i := 1; i < len(tensors); i++ {
		tensor := tensors[i]
		if len(tensor.Shape) != len(refShape) {
			return fmt.Errorf("tensor %d has different rank: %v vs %v", i, tensor.Shape, refShape)
		}

		// Check all dimensions except concatenation dimension
		for j := 0; j < len(refShape); j++ {
			if j != dim && tensor.Shape[j] != refShape[j] {
				return fmt.Errorf("tensor %d has incompatible shape at dimension %d: %d vs %d",
					i, j, tensor.Shape[j], refShape[j])
			}
		}
	}

	return nil
}

// ShapesEqual checks if two tensor shapes are identical
func (tv *TensorValidator) ShapesEqual(shape1, shape2 []int) bool {
	if len(shape1) != len(shape2) {
		return false
	}
	for i := range shape1 {
		if shape1[i] != shape2[i] {
			return false
		}
	}
	return true
}

// ValidateDataConsistency ensures tensor data is consistent with shape
func (tv *TensorValidator) ValidateDataConsistency(tensor TensorData) error {
	expectedSize := calculateTensorSize(tensor.Shape)
	actualSize := len(tensor.Data)

	if actualSize != expectedSize {
		return fmt.Errorf("tensor data size %d doesn't match shape %v (expected %d)",
			actualSize, tensor.Shape, expectedSize)
	}

	return nil
}

// ValidatePositiveShape ensures all shape dimensions are positive
func (tv *TensorValidator) ValidatePositiveShape(shape []int) error {
	for i, dim := range shape {
		if dim <= 0 {
			return fmt.Errorf("invalid dimension %d at position %d: must be positive", dim, i)
		}
	}
	return nil
}

// Helper functions

func calculateTensorSize(shape []int) int {
	if len(shape) == 0 {
		return 0
	}
	size := 1
	for _, dim := range shape {
		size *= dim
	}
	return size
}

func extractTensorShapeFromMetadata(metadata map[string]interface{}, dataLength int) []int {
	// Try to extract shape information from metadata with robust type handling
	if bs := GetIntFromMetadata(metadata, "batch_size"); bs > 0 {
		if sl := GetIntFromMetadata(metadata, "sequence_length"); sl > 0 {
			if hs := GetIntFromMetadata(metadata, "hidden_size"); hs > 0 {
				if bs*sl*hs == dataLength {
					return []int{bs, sl, hs}
				}
			} else {
				hs := dataLength / (bs * sl)
				if hs > 0 && bs*sl*hs == dataLength {
					return []int{bs, sl, hs}
				}
			}
		} else {
			fs := dataLength / bs
			if fs > 0 && bs*fs == dataLength {
				return []int{bs, fs}
			}
		}
	}

	// Default to 1D shape if no metadata available
	return []int{dataLength}
}


func (ta *TensorAggregator) normalizeAttentionWeights(weights []float32) []float32 {
	// Use shared attention normalization utility
	return SharedAttentionNormalization(weights)
}

// concatenateArbitraryDimension handles concatenation along non-leading dimensions
func (ta *TensorAggregator) concatenateArbitraryDimension(outputData []float32, tensors []TensorData, dim int) error {
	refShape := tensors[0].Shape
	if dim < 0 || dim >= len(refShape) {
		return fmt.Errorf("invalid dimension %d for shape %v", dim, refShape)
	}

	// Compute strides for arbitrary dimension concatenation
	outer := 1
	for i := 0; i < dim; i++ {
		outer *= refShape[i]
	}

	inner := 1
	for i := dim + 1; i < len(refShape); i++ {
		inner *= refShape[i]
	}

	// Copy data with proper stride handling
	outputOffset := 0
	for outerIdx := 0; outerIdx < outer; outerIdx++ {
		for _, tensor := range tensors {
			dimSize := tensor.Shape[dim]
			stride := dimSize * inner

			// Calculate input offset for this tensor
			inputOffset := outerIdx * tensor.Shape[dim] * inner

			// Copy this tensor's data for current outer block
			if inputOffset+stride <= len(tensor.Data) && outputOffset+stride <= len(outputData) {
				copy(outputData[outputOffset:outputOffset+stride], tensor.Data[inputOffset:inputOffset+stride])
				outputOffset += stride
			} else {
				return fmt.Errorf("index out of bounds during concatenation")
			}
		}
	}

	return nil
}

// SharedAttentionNormalization provides unified attention normalization
func SharedAttentionNormalization(weights []float32) []float32 {
	if len(weights) == 0 {
		return weights
	}

	// Apply softmax normalization with math.Exp
	max := weights[0]
	for _, w := range weights[1:] {
		if w > max {
			max = w
		}
	}

	// Calculate exp and sum
	sum := float32(0)
	normalized := make([]float32, len(weights))
	for i, w := range weights {
		exp := float32(math.Exp(float64(w - max)))
		normalized[i] = exp
		sum += exp
	}

	// Normalize
	if sum > 0 {
		for i := range normalized {
			normalized[i] /= sum
		}
	}

	return normalized
}

// TensorMemoryManager handles efficient memory management for large tensors
type TensorMemoryManager struct {
	maxBufferSize int
	buffers       map[string][]float32
}

func NewTensorMemoryManager(maxBufferSize int) *TensorMemoryManager {
	return &TensorMemoryManager{
		maxBufferSize: maxBufferSize,
		buffers:       make(map[string][]float32),
	}
}

// GetBuffer retrieves or creates a reusable buffer for tensor operations
func (tmm *TensorMemoryManager) GetBuffer(key string, size int) []float32 {
	if buffer, exists := tmm.buffers[key]; exists && len(buffer) >= size {
		return buffer[:size]
	}

	// Create new buffer
	buffer := make([]float32, size)
	if size <= tmm.maxBufferSize {
		tmm.buffers[key] = buffer
	}

	return buffer
}

// ReleaseBuffer releases a buffer for reuse
func (tmm *TensorMemoryManager) ReleaseBuffer(key string) {
	delete(tmm.buffers, key)
}

// ClearBuffers clears all buffers to free memory
func (tmm *TensorMemoryManager) ClearBuffers() {
	tmm.buffers = make(map[string][]float32)
}
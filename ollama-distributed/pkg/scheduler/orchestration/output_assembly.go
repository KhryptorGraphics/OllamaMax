package orchestration

import (
	"fmt"
	"errors"
	"sort"
	"math"
	"math/rand"
)

// OutputType represents different types of model outputs
type OutputType string

const (
	SequenceOutput      OutputType = "sequence"
	ClassificationOutput OutputType = "classification"
	EmbeddingOutput     OutputType = "embedding"
	GenerativeOutput    OutputType = "generative"
	MultiModalOutput    OutputType = "multimodal"
)

// OutputFormat specifies the format of assembled output
type OutputFormat string

const (
	TokenSequenceFormat    OutputFormat = "token_sequence"
	ProbabilityFormat      OutputFormat = "probability"
	EmbeddingVectorFormat  OutputFormat = "embedding_vector"
	StructuredFormat       OutputFormat = "structured"
	StreamingFormat        OutputFormat = "streaming"
)

// AssembledOutput represents the final assembled output
type AssembledOutput struct {
	Type        OutputType
	Format      OutputFormat
	Data        interface{}
	Tokens      []int
	Probabilities []float32
	Embeddings  []float32
	Text        string
	Metadata    map[string]interface{}
	Quality     *OutputQuality
}

// OutputQuality contains quality metrics for the assembled output
type OutputQuality struct {
	Confidence    float32
	Coherence     float32
	Completeness  float32
	Consistency   float32
	ValidationScore float32
	Metrics       map[string]float32
}

// OutputAssembler coordinates final assembly of distributed inference results
type OutputAssembler struct {
	sequenceAssembler      *SequenceOutputAssembler
	classificationAssembler *ClassificationOutputAssembler
	embeddingAssembler     *EmbeddingOutputAssembler
	generativeAssembler    *GenerativeOutputAssembler
	validator              *OutputValidator
}

func NewOutputAssembler() *OutputAssembler {
	return &OutputAssembler{
		sequenceAssembler:       NewSequenceOutputAssembler(),
		classificationAssembler: NewClassificationOutputAssembler(),
		embeddingAssembler:      NewEmbeddingOutputAssembler(),
		generativeAssembler:     NewGenerativeOutputAssembler(),
		validator:               NewOutputValidator(),
	}
}

// AssembleOutput assembles final output based on aggregated partial results
func (oa *OutputAssembler) AssembleOutput(
	aggregatedData *TensorData,
	outputType OutputType,
	outputFormat OutputFormat,
	metadata map[string]interface{},
) (*AssembledOutput, error) {

	// Validate input
	if aggregatedData == nil {
		return nil, errors.New("aggregated data is required")
	}

	// Route to appropriate assembler based on output type
	var result *AssembledOutput
	var err error

	switch outputType {
	case SequenceOutput:
		result, err = oa.sequenceAssembler.AssembleSequence(aggregatedData, outputFormat, metadata)
	case ClassificationOutput:
		result, err = oa.classificationAssembler.AssembleClassification(aggregatedData, outputFormat, metadata)
	case EmbeddingOutput:
		result, err = oa.embeddingAssembler.AssembleEmbedding(aggregatedData, outputFormat, metadata)
	case GenerativeOutput:
		result, err = oa.generativeAssembler.AssembleGenerative(aggregatedData, outputFormat, metadata)
	default:
		return nil, fmt.Errorf("unsupported output type: %s", outputType)
	}

	if err != nil {
		return nil, fmt.Errorf("output assembly failed: %v", err)
	}

	// Validate assembled output
	quality, err := oa.validator.ValidateOutput(result)
	if err != nil {
		return nil, fmt.Errorf("output validation failed: %v", err)
	}

	result.Quality = quality
	return result, nil
}

// AssembleStreamingOutput handles streaming output assembly
func (oa *OutputAssembler) AssembleStreamingOutput(
	partialResults []*TensorData,
	outputType OutputType,
	callback func(*AssembledOutput) error,
) error {

	// Process partial results as they arrive
	for i, partial := range partialResults {
		assembled, err := oa.AssembleOutput(partial, outputType, StreamingFormat, map[string]interface{}{
			"streaming_index": i,
			"is_partial":      i < len(partialResults)-1,
		})
		if err != nil {
			return fmt.Errorf("streaming assembly failed at index %d: %v", i, err)
		}

		// Send assembled output to callback
		if err := callback(assembled); err != nil {
			return fmt.Errorf("streaming callback failed: %v", err)
		}
	}

	return nil
}

// SequenceOutputAssembler handles language model token generation and sequence completion
type SequenceOutputAssembler struct {
	tokenizer *TokenProcessor
}

func NewSequenceOutputAssembler() *SequenceOutputAssembler {
	return &SequenceOutputAssembler{
		tokenizer: NewTokenProcessor(),
	}
}

// AssembleSequence assembles token sequences from distributed partitions
func (soa *SequenceOutputAssembler) AssembleSequence(
	aggregatedData *TensorData,
	format OutputFormat,
	metadata map[string]interface{},
) (*AssembledOutput, error) {

	if aggregatedData.Type == "tokens" {
		// Direct token sequence assembly
		tokens := make([]int, len(aggregatedData.Data))
		for i, val := range aggregatedData.Data {
			tokens[i] = int(val)
		}

		text, err := soa.tokenizer.DetokenizeSequence(tokens)
		if err != nil {
			return nil, fmt.Errorf("detokenization failed: %v", err)
		}

		return &AssembledOutput{
			Type:   SequenceOutput,
			Format: format,
			Data:   tokens,
			Tokens: tokens,
			Text:   text,
			Metadata: map[string]interface{}{
				"sequence_length": len(tokens),
				"token_count":     len(tokens),
			},
		}, nil
	}

	// Handle logits to token sequence conversion
	if aggregatedData.Type == "logits" {
		return soa.assembleFromLogits(aggregatedData, format, metadata)
	}

	return nil, errors.New("unsupported data type for sequence assembly")
}

// assembleFromLogits converts logits to token sequences
func (soa *SequenceOutputAssembler) assembleFromLogits(
	logitsData *TensorData,
	format OutputFormat,
	metadata map[string]interface{},
) (*AssembledOutput, error) {

	// Apply sampling strategy based on metadata
	samplingStrategy := "greedy" // Default
	if strategy, ok := metadata["sampling_strategy"]; ok {
		if s, ok := strategy.(string); ok {
			samplingStrategy = s
		}
	}

	tokens, probabilities, err := soa.sampleTokensWithShape(logitsData, samplingStrategy, metadata)
	if err != nil {
		return nil, fmt.Errorf("token sampling failed: %v", err)
	}

	text, err := soa.tokenizer.DetokenizeSequence(tokens)
	if err != nil {
		return nil, fmt.Errorf("detokenization failed: %v", err)
	}

	return &AssembledOutput{
		Type:          SequenceOutput,
		Format:        format,
		Data:          tokens,
		Tokens:        tokens,
		Probabilities: probabilities,
		Text:          text,
		Metadata: map[string]interface{}{
			"sampling_strategy": samplingStrategy,
			"sequence_length":   len(tokens),
		},
	}, nil
}

// sampleTokensWithShape applies sampling strategy using shape-aware processing
func (soa *SequenceOutputAssembler) sampleTokensWithShape(
	logitsData *TensorData,
	strategy string,
	metadata map[string]interface{},
) ([]int, []float32, error) {
	logits := logitsData.Data
	shape := logitsData.Shape

	// Use shape from TensorData if available, otherwise fall back to metadata
	var batchSize, seqLen, vocabSize int

	if len(shape) > 0 {
		// Derive dimensions from TensorData.Shape
		if len(shape) == 3 {
			// [batch, seq, vocab]
			batchSize = shape[0]
			seqLen = shape[1]
			vocabSize = shape[2]
		} else if len(shape) == 2 {
			// [batch, vocab]
			batchSize = shape[0]
			seqLen = 1
			vocabSize = shape[1]
		} else if len(shape) == 1 {
			// [vocab]
			batchSize = 1
			seqLen = 1
			vocabSize = shape[0]
		}
	} else {
		// Fallback to metadata if shape is empty
		batchSize = 1
		seqLen = 1
		vocabSize = len(logits)

		if bs := GetIntFromMetadata(metadata, "batch_size"); bs > 0 {
			batchSize = bs
		}
		if sl := GetIntFromMetadata(metadata, "sequence_length"); sl > 0 {
			seqLen = sl
		}
		if vs := GetIntFromMetadata(metadata, "vocab_size"); vs > 0 {
			vocabSize = vs
		}
	}

	// Update metadata with resolved shape
	metadata["resolved_batch_size"] = batchSize
	metadata["resolved_seq_length"] = seqLen
	metadata["resolved_vocab_size"] = vocabSize

	// Handle different logits shapes
	if batchSize*seqLen*vocabSize == len(logits) {
		// [batch, seq, vocab] - sample for each time step
		return soa.sampleBatchedSequenceLogits(logits, batchSize, seqLen, vocabSize, strategy, metadata)
	} else if batchSize*vocabSize == len(logits) {
		// [batch, vocab] - sample for each batch
		return soa.sampleBatchedLogits(logits, batchSize, vocabSize, strategy, metadata)
	} else if vocabSize == len(logits) {
		// [vocab] - single sample
		// Forward to original implementation for single sampling
		return soa.sampleTokens(logits, strategy, metadata)
	}

	return nil, nil, fmt.Errorf("shape mismatch: shape %v doesn't match data length %d", shape, len(logits))
}

// sampleTokens applies sampling strategy to convert logits to tokens with batched/sequence support
func (soa *SequenceOutputAssembler) sampleTokens(
	logits []float32,
	strategy string,
	metadata map[string]interface{},
) ([]int, []float32, error) {

	// Determine logits shape from metadata or infer from data
	batchSize := 1
	seqLen := 1
	vocabSize := len(logits)

	// Extract shape information from metadata
	if bs := GetIntFromMetadata(metadata, "batch_size"); bs > 0 {
		batchSize = bs
	}
	if sl := GetIntFromMetadata(metadata, "sequence_length"); sl > 0 {
		seqLen = sl
	}
	if vs := GetIntFromMetadata(metadata, "vocab_size"); vs > 0 {
		vocabSize = vs
	}

	// Handle different logits shapes
	if batchSize*seqLen*vocabSize == len(logits) {
		// [batch, seq, vocab] - sample for each time step
		return soa.sampleBatchedSequenceLogits(logits, batchSize, seqLen, vocabSize, strategy, metadata)
	} else if batchSize*vocabSize == len(logits) {
		// [batch, vocab] - sample for each batch
		return soa.sampleBatchedLogits(logits, batchSize, vocabSize, strategy, metadata)
	} else {
		// [vocab] - single sample
		switch strategy {
		case "greedy":
			return soa.greedySampling(logits)
		case "top_k":
			k := 50
			if kVal, ok := metadata["top_k"]; ok {
				if kInt, ok := kVal.(int); ok {
					k = kInt
				}
			}
			return soa.topKSampling(logits, k)
		case "nucleus":
			p := float32(0.9)
			if pVal, ok := metadata["nucleus_p"]; ok {
				if pFloat, ok := pVal.(float32); ok {
					p = pFloat
				}
			}
			return soa.nucleusSampling(logits, p)
		default:
			return soa.greedySampling(logits)
		}
	}
}

func (soa *SequenceOutputAssembler) greedySampling(logits []float32) ([]int, []float32, error) {
	if len(logits) == 0 {
		return nil, nil, errors.New("empty logits")
	}

	maxIdx := 0
	maxVal := logits[0]
	for i, val := range logits[1:] {
		if val > maxVal {
			maxVal = val
			maxIdx = i + 1
		}
	}

	// Apply softmax to get probabilities
	probabilities := applySoftmax(logits)

	return []int{maxIdx}, []float32{probabilities[maxIdx]}, nil
}

func (soa *SequenceOutputAssembler) topKSampling(logits []float32, k int) ([]int, []float32, error) {
	if k <= 0 || k > len(logits) {
		k = len(logits)
	}

	// Get top-k indices
	indices := make([]int, len(logits))
	for i := range indices {
		indices[i] = i
	}

	sort.Slice(indices, func(i, j int) bool {
		return logits[indices[i]] > logits[indices[j]]
	})

	topKIndices := indices[:k]
	topKLogits := make([]float32, k)
	for i, idx := range topKIndices {
		topKLogits[i] = logits[idx]
	}

	// Apply softmax to top-k
	probabilities := applySoftmax(topKLogits)

	// Sample from top-k distribution
	selectedIdx := soa.sampleFromDistribution(probabilities)
	if selectedIdx >= len(topKIndices) {
		selectedIdx = 0
	}

	token := topKIndices[selectedIdx]
	prob := probabilities[selectedIdx]

	return []int{token}, []float32{prob}, nil
}

func (soa *SequenceOutputAssembler) nucleusSampling(logits []float32, p float32) ([]int, []float32, error) {
	// Sort logits in descending order
	indices := make([]int, len(logits))
	for i := range indices {
		indices[i] = i
	}

	sort.Slice(indices, func(i, j int) bool {
		return logits[indices[i]] > logits[indices[j]]
	})

	// Apply softmax
	sortedLogits := make([]float32, len(logits))
	for i, idx := range indices {
		sortedLogits[i] = logits[idx]
	}
	probabilities := applySoftmax(sortedLogits)

	// Find nucleus (cumulative probability >= p)
	var cumSum float32
	nucleusSize := 0
	for i, prob := range probabilities {
		cumSum += prob
		nucleusSize = i + 1
		if cumSum >= p {
			break
		}
	}

	// Renormalize nucleus probabilities
	nucleusProbs := make([]float32, nucleusSize)
	var nucleusSum float32
	for i := 0; i < nucleusSize; i++ {
		nucleusProbs[i] = probabilities[i]
		nucleusSum += nucleusProbs[i]
	}

	if nucleusSum > 0 {
		for i := range nucleusProbs {
			nucleusProbs[i] /= nucleusSum
		}
	}

	// Sample from nucleus
	selectedIdx := soa.sampleFromDistribution(nucleusProbs)
	if selectedIdx >= len(indices) {
		selectedIdx = 0
	}

	token := indices[selectedIdx]
	prob := nucleusProbs[selectedIdx]

	return []int{token}, []float32{prob}, nil
}

func (soa *SequenceOutputAssembler) sampleFromDistribution(probs []float32) int {
	if len(probs) == 0 {
		return 0
	}

	// Generate random number [0, 1)
	r := rand.Float32()

	// Find cumulative probability bucket
	var cumSum float32
	for i, prob := range probs {
		cumSum += prob
		if r <= cumSum {
			return i
		}
	}

	return len(probs) - 1
}

// sampleBatchedSequenceLogits handles [batch, seq, vocab] shaped logits
func (soa *SequenceOutputAssembler) sampleBatchedSequenceLogits(
	logits []float32,
	batchSize, seqLen, vocabSize int,
	strategy string,
	metadata map[string]interface{},
) ([]int, []float32, error) {

	tokens := make([]int, batchSize*seqLen)
	probabilities := make([]float32, batchSize*seqLen)

	// Process each batch and sequence position
	for b := 0; b < batchSize; b++ {
		for s := 0; s < seqLen; s++ {
			// Extract logits for this position: [vocab]
			offset := (b*seqLen + s) * vocabSize
			posLogits := logits[offset : offset+vocabSize]

			// Sample single token
			posTokens, posProbs, err := soa.sampleSinglePosition(posLogits, strategy, metadata)
			if err != nil {
				return nil, nil, fmt.Errorf("sampling failed at batch %d, seq %d: %v", b, s, err)
			}

			// Store results
			idx := b*seqLen + s
			tokens[idx] = posTokens[0]
			probabilities[idx] = posProbs[0]
		}
	}

	return tokens, probabilities, nil
}

// sampleBatchedLogits handles [batch, vocab] shaped logits
func (soa *SequenceOutputAssembler) sampleBatchedLogits(
	logits []float32,
	batchSize, vocabSize int,
	strategy string,
	metadata map[string]interface{},
) ([]int, []float32, error) {

	tokens := make([]int, batchSize)
	probabilities := make([]float32, batchSize)

	// Process each batch
	for b := 0; b < batchSize; b++ {
		// Extract logits for this batch: [vocab]
		offset := b * vocabSize
		batchLogits := logits[offset : offset+vocabSize]

		// Sample single token
		batchTokens, batchProbs, err := soa.sampleSinglePosition(batchLogits, strategy, metadata)
		if err != nil {
			return nil, nil, fmt.Errorf("sampling failed at batch %d: %v", b, err)
		}

		// Store results
		tokens[b] = batchTokens[0]
		probabilities[b] = batchProbs[0]
	}

	return tokens, probabilities, nil
}

// sampleSinglePosition applies sampling strategy to a single vocab distribution
func (soa *SequenceOutputAssembler) sampleSinglePosition(
	logits []float32,
	strategy string,
	metadata map[string]interface{},
) ([]int, []float32, error) {

	switch strategy {
	case "greedy":
		return soa.greedySampling(logits)
	case "top_k":
		k := 50
		if kVal := GetIntFromMetadata(metadata, "top_k"); kVal > 0 {
			k = kVal
		}
		return soa.topKSampling(logits, k)
	case "nucleus":
		p := float32(0.9)
		if pVal, ok := metadata["nucleus_p"]; ok {
			if pFloat, ok := pVal.(float32); ok {
				p = pFloat
			}
		}
		return soa.nucleusSampling(logits, p)
	default:
		return soa.greedySampling(logits)
	}
}


// ClassificationOutputAssembler handles classification model logits and probability distributions
type ClassificationOutputAssembler struct{}

func NewClassificationOutputAssembler() *ClassificationOutputAssembler {
	return &ClassificationOutputAssembler{}
}

// AssembleClassification assembles classification outputs with proper probability handling
func (coa *ClassificationOutputAssembler) AssembleClassification(
	aggregatedData *TensorData,
	format OutputFormat,
	metadata map[string]interface{},
) (*AssembledOutput, error) {

	if aggregatedData.Type != "logits" {
		return nil, fmt.Errorf("expected logits for classification, got %s", aggregatedData.Type)
	}

	// Apply softmax to get probabilities
	probabilities := applySoftmax(aggregatedData.Data)

	// Find predicted class
	predictedClass := 0
	maxProb := probabilities[0]
	for i, prob := range probabilities[1:] {
		if prob > maxProb {
			maxProb = prob
			predictedClass = i + 1
		}
	}

	// Get top-k predictions if requested
	topK := 5 // Default
	if k, ok := metadata["top_k"]; ok {
		if kInt, ok := k.(int); ok {
			topK = kInt
		}
	}

	topPredictions := coa.getTopKPredictions(probabilities, topK)

	return &AssembledOutput{
		Type:          ClassificationOutput,
		Format:        format,
		Data:          predictedClass,
		Probabilities: probabilities,
		Metadata: map[string]interface{}{
			"predicted_class":   predictedClass,
			"confidence":        maxProb,
			"top_predictions":   topPredictions,
			"num_classes":       len(probabilities),
		},
	}, nil
}

func (coa *ClassificationOutputAssembler) getTopKPredictions(probs []float32, k int) []map[string]interface{} {
	type prediction struct {
		classID int
		prob    float32
	}

	predictions := make([]prediction, len(probs))
	for i, prob := range probs {
		predictions[i] = prediction{classID: i, prob: prob}
	}

	// Sort by probability descending
	sort.Slice(predictions, func(i, j int) bool {
		return predictions[i].prob > predictions[j].prob
	})

	// Take top-k
	if k > len(predictions) {
		k = len(predictions)
	}

	result := make([]map[string]interface{}, k)
	for i := 0; i < k; i++ {
		result[i] = map[string]interface{}{
			"class_id":    predictions[i].classID,
			"probability": predictions[i].prob,
		}
	}

	return result
}

// EmbeddingOutputAssembler handles embedding model outputs
type EmbeddingOutputAssembler struct{}

func NewEmbeddingOutputAssembler() *EmbeddingOutputAssembler {
	return &EmbeddingOutputAssembler{}
}

// AssembleEmbedding assembles embedding vectors with normalization
func (eoa *EmbeddingOutputAssembler) AssembleEmbedding(
	aggregatedData *TensorData,
	format OutputFormat,
	metadata map[string]interface{},
) (*AssembledOutput, error) {

	embeddings := aggregatedData.Data

	// Apply normalization if requested
	normalize := true // Default
	if norm, ok := metadata["normalize"]; ok {
		if normBool, ok := norm.(bool); ok {
			normalize = normBool
		}
	}

	if normalize {
		embeddings = eoa.normalizeEmbedding(embeddings)
	}

	// Calculate embedding statistics
	magnitude := eoa.calculateMagnitude(embeddings)
	dimension := len(embeddings)

	return &AssembledOutput{
		Type:       EmbeddingOutput,
		Format:     format,
		Data:       embeddings,
		Embeddings: embeddings,
		Metadata: map[string]interface{}{
			"dimension":  dimension,
			"magnitude":  magnitude,
			"normalized": normalize,
		},
	}, nil
}

func (eoa *EmbeddingOutputAssembler) normalizeEmbedding(embedding []float32) []float32 {
	// L2 normalization
	var sumSquares float32
	for _, val := range embedding {
		sumSquares += val * val
	}

	magnitude := float32(math.Sqrt(float64(sumSquares)))
	if magnitude == 0 {
		return embedding
	}

	normalized := make([]float32, len(embedding))
	for i, val := range embedding {
		normalized[i] = val / magnitude
	}

	return normalized
}

func (eoa *EmbeddingOutputAssembler) calculateMagnitude(embedding []float32) float32 {
	var sumSquares float32
	for _, val := range embedding {
		sumSquares += val * val
	}
	return float32(math.Sqrt(float64(sumSquares)))
}

// GenerativeOutputAssembler handles generative model outputs with beam search
type GenerativeOutputAssembler struct {
	sequenceAssembler *SequenceOutputAssembler
}

func NewGenerativeOutputAssembler() *GenerativeOutputAssembler {
	return &GenerativeOutputAssembler{
		sequenceAssembler: NewSequenceOutputAssembler(),
	}
}

// AssembleGenerative assembles generative outputs with advanced sampling
func (goa *GenerativeOutputAssembler) AssembleGenerative(
	aggregatedData *TensorData,
	format OutputFormat,
	metadata map[string]interface{},
) (*AssembledOutput, error) {

	// Check if beam search is requested
	useBeamSearch := false
	beamSize := 1
	if beam, ok := metadata["beam_search"]; ok {
		if beamBool, ok := beam.(bool); ok {
			useBeamSearch = beamBool
		}
	}
	if size, ok := metadata["beam_size"]; ok {
		if sizeInt, ok := size.(int); ok {
			beamSize = sizeInt
		}
	}

	if useBeamSearch && beamSize > 1 {
		return goa.assembleWithBeamSearch(aggregatedData, format, metadata, beamSize)
	}

	// Use standard sequence assembly for single sequence generation
	return goa.sequenceAssembler.AssembleSequence(aggregatedData, format, metadata)
}

func (goa *GenerativeOutputAssembler) assembleWithBeamSearch(
	aggregatedData *TensorData,
	format OutputFormat,
	metadata map[string]interface{},
	beamSize int,
) (*AssembledOutput, error) {

	// Simplified beam search implementation
	// In practice, this would maintain multiple hypothesis paths

	if aggregatedData.Type != "logits" {
		return nil, fmt.Errorf("beam search requires logits, got %s", aggregatedData.Type)
	}

	// For demonstration, generate top-k sequences
	probabilities := applySoftmax(aggregatedData.Data)

	// Get top beam_size tokens
	indices := make([]int, len(probabilities))
	for i := range indices {
		indices[i] = i
	}

	sort.Slice(indices, func(i, j int) bool {
		return probabilities[indices[i]] > probabilities[indices[j]]
	})

	// Generate multiple sequences (simplified)
	sequences := make([][]int, beamSize)
	sequenceProbs := make([]float32, beamSize)

	for i := 0; i < beamSize && i < len(indices); i++ {
		sequences[i] = []int{indices[i]}
		sequenceProbs[i] = probabilities[indices[i]]
	}

	// Return best sequence
	bestSequence := sequences[0]
	bestProb := sequenceProbs[0]

	text, err := goa.sequenceAssembler.tokenizer.DetokenizeSequence(bestSequence)
	if err != nil {
		text = fmt.Sprintf("tokens_%v", bestSequence)
	}

	return &AssembledOutput{
		Type:          GenerativeOutput,
		Format:        format,
		Data:          bestSequence,
		Tokens:        bestSequence,
		Probabilities: []float32{bestProb},
		Text:          text,
		Metadata: map[string]interface{}{
			"beam_search":    true,
			"beam_size":      beamSize,
			"best_sequence":  bestSequence,
			"all_sequences":  sequences,
			"sequence_probs": sequenceProbs,
		},
	}, nil
}

// OutputValidator validates assembled outputs for quality and correctness
type OutputValidator struct{}

func NewOutputValidator() *OutputValidator {
	return &OutputValidator{}
}

// ValidateOutput performs comprehensive validation of assembled output
func (ov *OutputValidator) ValidateOutput(output *AssembledOutput) (*OutputQuality, error) {
	quality := &OutputQuality{
		Metrics: make(map[string]float32),
	}

	// Validate based on output type
	switch output.Type {
	case SequenceOutput:
		return ov.validateSequenceOutput(output, quality)
	case ClassificationOutput:
		return ov.validateClassificationOutput(output, quality)
	case EmbeddingOutput:
		return ov.validateEmbeddingOutput(output, quality)
	case GenerativeOutput:
		return ov.validateGenerativeOutput(output, quality)
	default:
		quality.ValidationScore = 0.5 // Neutral score for unknown types
	}

	return quality, nil
}

func (ov *OutputValidator) validateSequenceOutput(output *AssembledOutput, quality *OutputQuality) (*OutputQuality, error) {
	// Validate token sequence
	if len(output.Tokens) == 0 {
		quality.Completeness = 0.0
	} else {
		quality.Completeness = 1.0
	}

	// Check for valid tokens (non-negative)
	validTokens := 0
	for _, token := range output.Tokens {
		if token >= 0 {
			validTokens++
		}
	}

	quality.Consistency = float32(validTokens) / float32(len(output.Tokens))
	quality.Confidence = 0.8 // Default confidence for sequences

	// Overall validation score
	quality.ValidationScore = (quality.Completeness + quality.Consistency + quality.Confidence) / 3.0

	quality.Metrics["token_count"] = float32(len(output.Tokens))
	quality.Metrics["valid_token_ratio"] = quality.Consistency

	return quality, nil
}

func (ov *OutputValidator) validateClassificationOutput(output *AssembledOutput, quality *OutputQuality) (*OutputQuality, error) {
	// Check probability distribution
	if len(output.Probabilities) == 0 {
		quality.Completeness = 0.0
		quality.ValidationScore = 0.0
		return quality, nil
	}

	// Validate probability sum (should be ~1.0)
	var probSum float32
	maxProb := float32(0)
	for _, prob := range output.Probabilities {
		probSum += prob
		if prob > maxProb {
			maxProb = prob
		}
	}

	quality.Completeness = 1.0
	quality.Consistency = 1.0 - float32(math.Abs(float64(probSum-1.0))) // Close to 1.0 is better
	quality.Confidence = maxProb // Highest probability as confidence

	quality.ValidationScore = (quality.Completeness + quality.Consistency + quality.Confidence) / 3.0

	quality.Metrics["probability_sum"] = probSum
	quality.Metrics["max_probability"] = maxProb
	quality.Metrics["num_classes"] = float32(len(output.Probabilities))

	return quality, nil
}

func (ov *OutputValidator) validateEmbeddingOutput(output *AssembledOutput, quality *OutputQuality) (*OutputQuality, error) {
	if len(output.Embeddings) == 0 {
		quality.Completeness = 0.0
		quality.ValidationScore = 0.0
		return quality, nil
	}

	// Check for NaN or infinite values
	validValues := 0
	var magnitude float32
	for _, val := range output.Embeddings {
		if !math.IsNaN(float64(val)) && !math.IsInf(float64(val), 0) {
			validValues++
		}
		magnitude += val * val
	}
	magnitude = float32(math.Sqrt(float64(magnitude)))

	quality.Completeness = 1.0
	quality.Consistency = float32(validValues) / float32(len(output.Embeddings))
	quality.Confidence = 0.9 // High confidence for valid embeddings

	quality.ValidationScore = (quality.Completeness + quality.Consistency + quality.Confidence) / 3.0

	quality.Metrics["embedding_dimension"] = float32(len(output.Embeddings))
	quality.Metrics["valid_value_ratio"] = quality.Consistency
	quality.Metrics["magnitude"] = magnitude

	return quality, nil
}

func (ov *OutputValidator) validateGenerativeOutput(output *AssembledOutput, quality *OutputQuality) (*OutputQuality, error) {
	// Use sequence validation as base
	quality, err := ov.validateSequenceOutput(output, quality)
	if err != nil {
		return quality, err
	}

	// Additional validation for generative outputs
	if output.Text != "" {
		quality.Coherence = ov.assessTextCoherence(output.Text)
		quality.ValidationScore = (quality.ValidationScore + quality.Coherence) / 2.0
	}

	quality.Metrics["text_length"] = float32(len(output.Text))

	return quality, nil
}

func (ov *OutputValidator) assessTextCoherence(text string) float32 {
	// Simplified coherence assessment
	if len(text) == 0 {
		return 0.0
	}

	// Check for reasonable length and character distribution
	coherenceScore := float32(0.8) // Base score

	// Penalize very short texts
	if len(text) < 10 {
		coherenceScore -= 0.3
	}

	// Bonus for reasonable length
	if len(text) >= 50 && len(text) <= 1000 {
		coherenceScore += 0.1
	}

	// Ensure score is in [0, 1] range
	if coherenceScore > 1.0 {
		coherenceScore = 1.0
	}
	if coherenceScore < 0.0 {
		coherenceScore = 0.0
	}

	return coherenceScore
}

// TokenProcessor handles tokenization and detokenization
type TokenProcessor struct{}

func NewTokenProcessor() *TokenProcessor {
	return &TokenProcessor{}
}

// DetokenizeSequence converts token IDs back to text
func (tp *TokenProcessor) DetokenizeSequence(tokens []int) (string, error) {
	// Simplified detokenization - in practice would use proper tokenizer
	if len(tokens) == 0 {
		return "", nil
	}

	// Mock detokenization - replace with actual tokenizer
	words := make([]string, len(tokens))
	for i, token := range tokens {
		if token < 0 {
			words[i] = "<UNK>"
		} else if token == 0 {
			words[i] = "<PAD>"
		} else if token == 1 {
			words[i] = "<BOS>"
		} else if token == 2 {
			words[i] = "<EOS>"
		} else {
			words[i] = fmt.Sprintf("token_%d", token)
		}
	}

	text := ""
	for i, word := range words {
		if i > 0 {
			text += " "
		}
		text += word
	}

	return text, nil
}
package orchestration

import (
	"fmt"
	"errors"
	"strings"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/partitioning"
)

// ModelFamily represents different families of ML models
type ModelFamily string

const (
	LLaMAFamily       ModelFamily = "llama"
	GPTFamily         ModelFamily = "gpt"
	BERTFamily        ModelFamily = "bert"
	T5Family          ModelFamily = "t5"
	TransformerFamily ModelFamily = "transformer"
	UnknownFamily     ModelFamily = "unknown"
)

// TransformerAggregationStrategy handles transformer-based model aggregation
type TransformerAggregationStrategy struct {
	name                string
	attentionCoordinator *AttentionCoordinator
	tensorAggregator    *TensorAggregator
}

func NewTransformerAggregationStrategy() *TransformerAggregationStrategy {
	return &TransformerAggregationStrategy{
		name:                "transformer",
		attentionCoordinator: NewAttentionCoordinator(),
		tensorAggregator:    NewTensorAggregator(),
	}
}

func (tas *TransformerAggregationStrategy) GetName() string {
	return tas.name
}

func (tas *TransformerAggregationStrategy) Aggregate(context *AggregationContext) (*AggregatedResponse, error) {
	// Extract model analysis from context
	modelAnalysis := extractModelAnalysis(context.Metadata)
	if modelAnalysis == nil {
		return nil, errors.New("transformer aggregation requires model analysis")
	}

	// Separate results by layer type for proper handling
	layerResults := tas.separateByLayerType(context.PartialResults, modelAnalysis)

	// Process attention layers first
	attentionResults, err := tas.processAttentionLayers(layerResults["attention"])
	if err != nil {
		return nil, fmt.Errorf("failed to process attention layers: %v", err)
	}

	// Process feed-forward layers
	feedForwardResults, err := tas.processFeedForwardLayers(layerResults["feed_forward"])
	if err != nil {
		return nil, fmt.Errorf("failed to process feed-forward layers: %v", err)
	}

	// Process normalization layers
	normResults, err := tas.processNormalizationLayers(layerResults["layer_norm"])
	if err != nil {
		return nil, fmt.Errorf("failed to process normalization layers: %v", err)
	}

	// Combine all layer results in proper transformer order
	finalResult, err := tas.combineTransformerLayers(attentionResults, feedForwardResults, normResults, modelAnalysis)
	if err != nil {
		return nil, fmt.Errorf("failed to combine transformer layers: %v", err)
	}

	return &AggregatedResponse{
		TaskID:   context.TaskID,
		Strategy: tas.name,
		Data:     finalResult,
		Metadata: map[string]interface{}{
			"model_family":      "transformer",
			"layer_count":       modelAnalysis.LayerInfo.TotalLayers,
			"attention_layers":  len(layerResults["attention"]),
			"feedforward_layers": len(layerResults["feed_forward"]),
		},
	}, nil
}

// LLMInferenceAggregationStrategy handles large language model inference
type LLMInferenceAggregationStrategy struct {
	name             string
	sequenceAssembler *SequenceOutputAssembler
	tensorAggregator *TensorAggregator
}

func NewLLMInferenceAggregationStrategy() *LLMInferenceAggregationStrategy {
	return &LLMInferenceAggregationStrategy{
		name:              "llm_inference",
		sequenceAssembler: NewSequenceOutputAssembler(),
		tensorAggregator:  NewTensorAggregator(),
	}
}

func (lias *LLMInferenceAggregationStrategy) GetName() string {
	return lias.name
}

func (lias *LLMInferenceAggregationStrategy) Aggregate(context *AggregationContext) (*AggregatedResponse, error) {
	// Determine model family for family-specific optimizations
	modelFamily := determineModelFamily(context.Metadata)

	// Apply family-specific aggregation logic
	switch modelFamily {
	case LLaMAFamily:
		return lias.aggregateLLaMA(context)
	case GPTFamily:
		return lias.aggregateGPT(context)
	case T5Family:
		return lias.aggregateT5(context)
	default:
		return lias.aggregateGenericLLM(context)
	}
}

func (lias *LLMInferenceAggregationStrategy) aggregateLLaMA(context *AggregationContext) (*AggregatedResponse, error) {
	// LLaMA-specific aggregation with RMSNorm handling
	return lias.aggregateWithRMSNorm(context)
}

func (lias *LLMInferenceAggregationStrategy) aggregateGPT(context *AggregationContext) (*AggregatedResponse, error) {
	// GPT-specific aggregation with LayerNorm
	return lias.aggregateWithLayerNorm(context)
}

func (lias *LLMInferenceAggregationStrategy) aggregateT5(context *AggregationContext) (*AggregatedResponse, error) {
	// T5-specific aggregation with encoder-decoder handling
	return lias.aggregateEncoderDecoder(context)
}

func (lias *LLMInferenceAggregationStrategy) aggregateGenericLLM(context *AggregationContext) (*AggregatedResponse, error) {
	// Generic LLM aggregation using standard transformer patterns
	transformerStrategy := NewTransformerAggregationStrategy()
	return transformerStrategy.Aggregate(context)
}

// MultiModalAggregationStrategy handles multi-modal model aggregation
type MultiModalAggregationStrategy struct {
	name              string
	textProcessor     *TextModalityProcessor
	imageProcessor    *ImageModalityProcessor
	audioProcessor    *AudioModalityProcessor
	fusionProcessor   *ModalityFusionProcessor
}

func NewMultiModalAggregationStrategy() *MultiModalAggregationStrategy {
	return &MultiModalAggregationStrategy{
		name:              "multimodal",
		textProcessor:     NewTextModalityProcessor(),
		imageProcessor:    NewImageModalityProcessor(),
		audioProcessor:    NewAudioModalityProcessor(),
		fusionProcessor:   NewModalityFusionProcessor(),
	}
}

func (mmas *MultiModalAggregationStrategy) GetName() string {
	return mmas.name
}

func (mmas *MultiModalAggregationStrategy) Aggregate(context *AggregationContext) (*AggregatedResponse, error) {
	// Separate results by modality
	modalityResults := mmas.separateByModality(context.PartialResults)

	// Process each modality
	var processedModalities []ModalityResult

	// Process text modality
	if textResults, exists := modalityResults["text"]; exists {
		textResult, err := mmas.textProcessor.ProcessModality(textResults)
		if err != nil {
			return nil, fmt.Errorf("text modality processing failed: %v", err)
		}
		processedModalities = append(processedModalities, *textResult)
	}

	// Process image modality
	if imageResults, exists := modalityResults["image"]; exists {
		imageResult, err := mmas.imageProcessor.ProcessModality(imageResults)
		if err != nil {
			return nil, fmt.Errorf("image modality processing failed: %v", err)
		}
		processedModalities = append(processedModalities, *imageResult)
	}

	// Process audio modality
	if audioResults, exists := modalityResults["audio"]; exists {
		audioResult, err := mmas.audioProcessor.ProcessModality(audioResults)
		if err != nil {
			return nil, fmt.Errorf("audio modality processing failed: %v", err)
		}
		processedModalities = append(processedModalities, *audioResult)
	}

	// Fuse modalities
	fusedResult, err := mmas.fusionProcessor.FuseModalities(processedModalities)
	if err != nil {
		return nil, fmt.Errorf("modality fusion failed: %v", err)
	}

	return &AggregatedResponse{
		TaskID:   context.TaskID,
		Strategy: mmas.name,
		Data:     fusedResult,
		Metadata: map[string]interface{}{
			"modalities_processed": len(processedModalities),
			"fusion_method":        mmas.fusionProcessor.GetFusionMethod(),
		},
	}, nil
}

// DistributedTrainingAggregationStrategy handles distributed training aggregation
type DistributedTrainingAggregationStrategy struct {
	name               string
	gradientAggregator *GradientAggregator
	optimizerSync      *OptimizerSynchronizer
}

func NewDistributedTrainingAggregationStrategy() *DistributedTrainingAggregationStrategy {
	return &DistributedTrainingAggregationStrategy{
		name:               "distributed_training",
		gradientAggregator: NewGradientAggregator(),
		optimizerSync:      NewOptimizerSynchronizer(),
	}
}

func (dtas *DistributedTrainingAggregationStrategy) GetName() string {
	return dtas.name
}

func (dtas *DistributedTrainingAggregationStrategy) Aggregate(context *AggregationContext) (*AggregatedResponse, error) {
	// Extract gradients from partial results
	gradients := dtas.extractGradients(context.PartialResults)
	if len(gradients) == 0 {
		return nil, errors.New("no gradients found in partial results")
	}

	// Aggregate gradients (AllReduce operation)
	aggregatedGradients, err := dtas.gradientAggregator.AllReduce(gradients)
	if err != nil {
		return nil, fmt.Errorf("gradient aggregation failed: %v", err)
	}

	// Synchronize optimizer states if present
	if dtas.hasOptimizerStates(context.PartialResults) {
		err = dtas.optimizerSync.SynchronizeStates(context.PartialResults)
		if err != nil {
			return nil, fmt.Errorf("optimizer synchronization failed: %v", err)
		}
	}

	return &AggregatedResponse{
		TaskID:   context.TaskID,
		Strategy: dtas.name,
		Data:     aggregatedGradients,
		Metadata: map[string]interface{}{
			"gradient_count":     len(gradients),
			"optimizer_synced":   dtas.hasOptimizerStates(context.PartialResults),
			"aggregation_method": "allreduce",
		},
	}, nil
}

// Helper structures and processors

// ModalityResult represents processed result from a single modality
type ModalityResult struct {
	Modality string
	Data     []float32
	Shape    []int
	Metadata map[string]interface{}
}

// TextModalityProcessor processes text modality
type TextModalityProcessor struct{}

func NewTextModalityProcessor() *TextModalityProcessor {
	return &TextModalityProcessor{}
}

func (tmp *TextModalityProcessor) ProcessModality(results []*PartialResult) (*ModalityResult, error) {
	// Combine text embeddings/hidden states
	aggregator := NewTensorAggregator()
	tensors := make([]TensorData, len(results))

	for i, result := range results {
		if len(result.HiddenStates) == 0 {
			return nil, fmt.Errorf("no text data in result %d", i)
		}
		tensors[i] = TensorData{
			Shape: []int{len(result.HiddenStates)},
			Data:  result.HiddenStates,
			Type:  "text_embedding",
		}
	}

	combined, err := aggregator.ConcatenateTensors(tensors, 0)
	if err != nil {
		return nil, err
	}

	return &ModalityResult{
		Modality: "text",
		Data:     combined.Data,
		Shape:    combined.Shape,
		Metadata: map[string]interface{}{
			"sequence_length": combined.Shape[0],
		},
	}, nil
}

// ImageModalityProcessor processes image modality
type ImageModalityProcessor struct{}

func NewImageModalityProcessor() *ImageModalityProcessor {
	return &ImageModalityProcessor{}
}

func (imp *ImageModalityProcessor) ProcessModality(results []*PartialResult) (*ModalityResult, error) {
	// Process image features
	aggregator := NewTensorAggregator()
	tensors := make([]TensorData, len(results))

	for i, result := range results {
		// Assume image features are stored in HiddenStates
		if len(result.HiddenStates) == 0 {
			return nil, fmt.Errorf("no image data in result %d", i)
		}
		tensors[i] = TensorData{
			Shape: []int{len(result.HiddenStates)},
			Data:  result.HiddenStates,
			Type:  "image_features",
		}
	}

	combined, err := aggregator.AverageTensors(tensors) // Average for image features
	if err != nil {
		return nil, err
	}

	return &ModalityResult{
		Modality: "image",
		Data:     combined.Data,
		Shape:    combined.Shape,
		Metadata: map[string]interface{}{
			"feature_dim": combined.Shape[0],
		},
	}, nil
}

// AudioModalityProcessor processes audio modality
type AudioModalityProcessor struct{}

func NewAudioModalityProcessor() *AudioModalityProcessor {
	return &AudioModalityProcessor{}
}

func (amp *AudioModalityProcessor) ProcessModality(results []*PartialResult) (*ModalityResult, error) {
	// Process audio features
	aggregator := NewTensorAggregator()
	tensors := make([]TensorData, len(results))

	for i, result := range results {
		if len(result.HiddenStates) == 0 {
			return nil, fmt.Errorf("no audio data in result %d", i)
		}
		tensors[i] = TensorData{
			Shape: []int{len(result.HiddenStates)},
			Data:  result.HiddenStates,
			Type:  "audio_features",
		}
	}

	combined, err := aggregator.ConcatenateTensors(tensors, 0) // Concatenate for temporal audio
	if err != nil {
		return nil, err
	}

	return &ModalityResult{
		Modality: "audio",
		Data:     combined.Data,
		Shape:    combined.Shape,
		Metadata: map[string]interface{}{
			"temporal_length": combined.Shape[0],
		},
	}, nil
}

// ModalityFusionProcessor fuses different modalities
type ModalityFusionProcessor struct {
	fusionMethod string
}

func NewModalityFusionProcessor() *ModalityFusionProcessor {
	return &ModalityFusionProcessor{
		fusionMethod: "concatenation", // Default fusion method
	}
}

func (mfp *ModalityFusionProcessor) FuseModalities(modalities []ModalityResult) (*TensorData, error) {
	if len(modalities) == 0 {
		return nil, errors.New("no modalities to fuse")
	}

	switch mfp.fusionMethod {
	case "concatenation":
		return mfp.fuseConcatenation(modalities)
	case "attention":
		return mfp.fuseAttention(modalities)
	case "weighted":
		return mfp.fuseWeighted(modalities)
	default:
		return mfp.fuseConcatenation(modalities) // Default
	}
}

func (mfp *ModalityFusionProcessor) GetFusionMethod() string {
	return mfp.fusionMethod
}

func (mfp *ModalityFusionProcessor) fuseConcatenation(modalities []ModalityResult) (*TensorData, error) {
	// Simple concatenation fusion
	totalLength := 0
	for _, mod := range modalities {
		totalLength += len(mod.Data)
	}

	fusedData := make([]float32, totalLength)
	offset := 0

	for _, mod := range modalities {
		copy(fusedData[offset:offset+len(mod.Data)], mod.Data)
		offset += len(mod.Data)
	}

	return &TensorData{
		Shape: []int{totalLength},
		Data:  fusedData,
		Type:  "fused_multimodal",
	}, nil
}

func (mfp *ModalityFusionProcessor) fuseAttention(modalities []ModalityResult) (*TensorData, error) {
	// Attention-based fusion (simplified)
	return mfp.fuseConcatenation(modalities) // Fallback to concatenation for now
}

func (mfp *ModalityFusionProcessor) fuseWeighted(modalities []ModalityResult) (*TensorData, error) {
	// Weighted fusion (simplified)
	return mfp.fuseConcatenation(modalities) // Fallback to concatenation for now
}

// GradientAggregator handles gradient aggregation for distributed training
type GradientAggregator struct{}

func NewGradientAggregator() *GradientAggregator {
	return &GradientAggregator{}
}

func (ga *GradientAggregator) AllReduce(gradients [][]float32) ([]float32, error) {
	if len(gradients) == 0 {
		return nil, errors.New("no gradients to aggregate")
	}

	// Check all gradients have same length
	gradientLength := len(gradients[0])
	for i, grad := range gradients[1:] {
		if len(grad) != gradientLength {
			return nil, fmt.Errorf("gradient %d has different length: %d vs %d", i+1, len(grad), gradientLength)
		}
	}

	// Average gradients (AllReduce with average)
	aggregated := make([]float32, gradientLength)
	for i := 0; i < gradientLength; i++ {
		sum := float32(0)
		for _, grad := range gradients {
			sum += grad[i]
		}
		aggregated[i] = sum / float32(len(gradients))
	}

	return aggregated, nil
}

// OptimizerSynchronizer synchronizes optimizer states
type OptimizerSynchronizer struct{}

func NewOptimizerSynchronizer() *OptimizerSynchronizer {
	return &OptimizerSynchronizer{}
}

func (os *OptimizerSynchronizer) SynchronizeStates(results []*PartialResult) error {
	// Placeholder for optimizer state synchronization
	// In practice, would synchronize momentum, learning rate schedules, etc.
	return nil
}

// Helper methods

func extractModelAnalysis(metadata map[string]interface{}) *partitioning.ModelAnalysis {
	if metadata == nil {
		return nil
	}

	if analysis, ok := metadata["model_analysis"]; ok {
		if modelAnalysis, ok := analysis.(*partitioning.ModelAnalysis); ok {
			return modelAnalysis
		}
	}

	return nil
}

func determineModelFamily(metadata map[string]interface{}) ModelFamily {
	if metadata == nil {
		return UnknownFamily
	}

	// Check for explicit model family
	if family, ok := metadata["model_family"]; ok {
		if familyStr, ok := family.(string); ok {
			switch strings.ToLower(familyStr) {
			case "llama":
				return LLaMAFamily
			case "gpt":
				return GPTFamily
			case "bert":
				return BERTFamily
			case "t5":
				return T5Family
			case "transformer":
				return TransformerFamily
			}
		}
	}

	// Infer from model name if available
	if modelName, ok := metadata["model_name"]; ok {
		if nameStr, ok := modelName.(string); ok {
			nameLower := strings.ToLower(nameStr)
			if strings.Contains(nameLower, "llama") {
				return LLaMAFamily
			}
			if strings.Contains(nameLower, "gpt") {
				return GPTFamily
			}
			if strings.Contains(nameLower, "bert") {
				return BERTFamily
			}
			if strings.Contains(nameLower, "t5") {
				return T5Family
			}
		}
	}

	return TransformerFamily // Default to transformer family
}

func (tas *TransformerAggregationStrategy) separateByLayerType(
	results []*PartialResult,
	modelAnalysis *partitioning.ModelAnalysis,
) map[string][]*PartialResult {

	layerResults := make(map[string][]*PartialResult)
	layerResults["attention"] = make([]*PartialResult, 0)
	layerResults["feed_forward"] = make([]*PartialResult, 0)
	layerResults["layer_norm"] = make([]*PartialResult, 0)

	// Separate results based on layer information from metadata
	for _, result := range results {
		layerType := "feed_forward" // Default
		if result.Metadata != nil {
			if lt, ok := result.Metadata["layer_type"]; ok {
				if ltStr, ok := lt.(string); ok {
					layerType = ltStr
				}
			}
		}

		if _, exists := layerResults[layerType]; !exists {
			layerResults[layerType] = make([]*PartialResult, 0)
		}
		layerResults[layerType] = append(layerResults[layerType], result)
	}

	return layerResults
}

func (tas *TransformerAggregationStrategy) processAttentionLayers(results []*PartialResult) (*TensorData, error) {
	if len(results) == 0 {
		return &TensorData{Data: []float32{}, Shape: []int{}, Type: "attention"}, nil
	}

	// Get model analysis from first result's metadata
	var modelAnalysis *partitioning.ModelAnalysis
	if len(results) > 0 && results[0].Metadata != nil {
		modelAnalysis = extractModelAnalysis(results[0].Metadata)
	}

	// Convert to attention outputs
	attentionOutputs := make([]AttentionOutput, len(results))
	for i, result := range results {
		// Extract attention dimensions from metadata or model analysis
		heads, seqLen, headDim, err := extractAttentionDims(result.Metadata, modelAnalysis)
		if err != nil {
			return nil, fmt.Errorf("failed to extract attention dimensions for partition %s: %v",
				result.PartitionID, err)
		}

		// Validate dimensions match data size
		expectedSize := heads * seqLen * headDim
		if len(result.HiddenStates) != expectedSize {
			return nil, fmt.Errorf("attention data size mismatch for partition %s: got %d, expected %d (heads=%d, seqLen=%d, headDim=%d)",
				result.PartitionID, len(result.HiddenStates), expectedSize, heads, seqLen, headDim)
		}

		attentionOutputs[i] = AttentionOutput{
			Values:  result.HiddenStates,
			Heads:   heads,
			SeqLen:  seqLen,
			HeadDim: headDim,
		}
	}

	// Aggregate attention
	return tas.attentionCoordinator.AggregateAttentionHeads(attentionOutputs)
}

func (tas *TransformerAggregationStrategy) processFeedForwardLayers(results []*PartialResult) (*TensorData, error) {
	if len(results) == 0 {
		return &TensorData{Data: []float32{}, Shape: []int{}, Type: "feed_forward"}, nil
	}

	// Simple tensor aggregation for feed-forward layers
	tensors := make([]TensorData, len(results))
	for i, result := range results {
		tensors[i] = TensorData{
			Shape: []int{len(result.HiddenStates)},
			Data:  result.HiddenStates,
			Type:  "feed_forward",
		}
	}

	return tas.tensorAggregator.ConcatenateTensors(tensors, 0)
}

func (tas *TransformerAggregationStrategy) processNormalizationLayers(results []*PartialResult) (*TensorData, error) {
	if len(results) == 0 {
		return &TensorData{Data: []float32{}, Shape: []int{}, Type: "layer_norm"}, nil
	}

	// Average normalization outputs
	tensors := make([]TensorData, len(results))
	for i, result := range results {
		tensors[i] = TensorData{
			Shape: []int{len(result.HiddenStates)},
			Data:  result.HiddenStates,
			Type:  "layer_norm",
		}
	}

	return tas.tensorAggregator.AverageTensors(tensors)
}

func (tas *TransformerAggregationStrategy) combineTransformerLayers(
	attention, feedForward, norm *TensorData,
	modelAnalysis *partitioning.ModelAnalysis,
) (*TensorData, error) {

	// Combine all layer types in proper transformer order
	layers := []*TensorData{attention, feedForward, norm}
	validLayers := make([]TensorData, 0)

	for _, layer := range layers {
		if len(layer.Data) > 0 {
			validLayers = append(validLayers, *layer)
		}
	}

	if len(validLayers) == 0 {
		return &TensorData{Data: []float32{}, Shape: []int{}, Type: "transformer"}, nil
	}

	return tas.tensorAggregator.ConcatenateTensors(validLayers, 0)
}

// Family-specific aggregation methods

func (lias *LLMInferenceAggregationStrategy) aggregateWithRMSNorm(context *AggregationContext) (*AggregatedResponse, error) {
	// LLaMA uses RMSNorm instead of LayerNorm
	transformerStrategy := NewTransformerAggregationStrategy()
	result, err := transformerStrategy.Aggregate(context)
	if err != nil {
		return nil, err
	}

	// Apply RMSNorm-specific post-processing
	result.Metadata["normalization"] = "rmsnorm"
	return result, nil
}

func (lias *LLMInferenceAggregationStrategy) aggregateWithLayerNorm(context *AggregationContext) (*AggregatedResponse, error) {
	// Standard transformer with LayerNorm
	transformerStrategy := NewTransformerAggregationStrategy()
	result, err := transformerStrategy.Aggregate(context)
	if err != nil {
		return nil, err
	}

	result.Metadata["normalization"] = "layernorm"
	return result, nil
}

func (lias *LLMInferenceAggregationStrategy) aggregateEncoderDecoder(context *AggregationContext) (*AggregatedResponse, error) {
	// T5 encoder-decoder architecture
	transformerStrategy := NewTransformerAggregationStrategy()
	result, err := transformerStrategy.Aggregate(context)
	if err != nil {
		return nil, err
	}

	result.Metadata["architecture"] = "encoder_decoder"
	return result, nil
}

func (mmas *MultiModalAggregationStrategy) separateByModality(results []*PartialResult) map[string][]*PartialResult {
	modalityResults := make(map[string][]*PartialResult)

	for _, result := range results {
		modality := "text" // Default
		if result.Metadata != nil {
			if mod, ok := result.Metadata["modality"]; ok {
				if modStr, ok := mod.(string); ok {
					modality = modStr
				}
			}
		}

		if modalityResults[modality] == nil {
			modalityResults[modality] = make([]*PartialResult, 0)
		}
		modalityResults[modality] = append(modalityResults[modality], result)
	}

	return modalityResults
}

func (dtas *DistributedTrainingAggregationStrategy) extractGradients(results []*PartialResult) [][]float32 {
	gradients := make([][]float32, 0)

	for _, result := range results {
		if result.Metadata != nil {
			if grad, ok := result.Metadata["gradients"]; ok {
				if gradSlice, ok := grad.([]float32); ok {
					gradients = append(gradients, gradSlice)
				}
			}
		}
	}

	return gradients
}

func (dtas *DistributedTrainingAggregationStrategy) hasOptimizerStates(results []*PartialResult) bool {
	for _, result := range results {
		if result.Metadata != nil {
			if _, ok := result.Metadata["optimizer_state"]; ok {
				return true
			}
		}
	}
	return false
}

// extractAttentionDims extracts attention dimensions from metadata or model analysis
func extractAttentionDims(metadata map[string]interface{}, analysis *partitioning.ModelAnalysis) (heads, seqLen, headDim int, err error) {
	// Try to extract from metadata first
	if metadata != nil {
		// Get attention heads
		heads = GetIntFromMetadata(metadata, "attention_heads")

		// Get sequence length
		seqLen = GetIntFromMetadata(metadata, "sequence_length")

		// Get head dimension
		headDim = GetIntFromMetadata(metadata, "head_dim")
	}

	// Fall back to model analysis if values are still zero
	// Use defaults if not found in metadata
	// TODO: Add AttentionHeads, SequenceLength, HeadDimension fields to LayerAnalysis
	if analysis != nil && analysis.LayerInfo != nil {
		// Default fallback values
		if heads == 0 {
			heads = 1 // Default to single head
		}
		if seqLen == 0 {
			seqLen = 1 // Default sequence length
		}
		if headDim == 0 {
			headDim = 64 // Common head dimension
		}
	}

	// Use coordinator defaults as last resort
	if heads == 0 {
		heads = 8 // Default number of attention heads
	}
	if seqLen == 0 {
		seqLen = 512 // Default sequence length
	}
	if headDim == 0 {
		headDim = 64 // Default head dimension
	}

	return heads, seqLen, headDim, nil
}


package partitioning

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/khryptorgraphics/ollamamax/pkg/types"
)

// ModelAnalysis contains detailed information about a model's architecture
type ModelAnalysis struct {
	ModelName     string                 `json:"model_name"`
	ParameterSize int64                  `json:"parameter_size"` // Total parameters in billions
	LayerInfo     *LayerAnalysis         `json:"layer_info"`
	TensorInfo    *TensorAnalysis        `json:"tensor_info"`
	StageInfo     *PipelineStageAnalysis `json:"stage_info"`
	MemoryReqs    *MemoryRequirements    `json:"memory_reqs"`
}

// LayerAnalysis contains layer-specific information
type LayerAnalysis struct {
	TotalLayers       int                    `json:"total_layers"`
	TransformerBlocks int                    `json:"transformer_blocks"`
	AttentionLayers   int                    `json:"attention_layers"`
	FFNLayers         int                    `json:"ffn_layers"`
	LayerSizes        []int64                `json:"layer_sizes"`        // Memory per layer in bytes
	LayerWeights      []float64              `json:"layer_weights"`      // Computational weight per layer
	Dependencies      [][]int                `json:"dependencies"`       // Layer dependencies
	LayerTypes        map[int]LayerType      `json:"layer_types"`
}

// TensorAnalysis contains tensor distribution information
type TensorAnalysis struct {
	AttentionTensors  []TensorInfo       `json:"attention_tensors"`
	WeightTensors     []TensorInfo       `json:"weight_tensors"`
	ActivationTensors []TensorInfo       `json:"activation_tensors"`
	CombinedTensors   []TensorInfo       `json:"combined_tensors"`   // All tensors in encounter order
	IndexMap          map[string]int     `json:"index_map"`          // Name to combined index mapping
	TotalTensorSize   int64              `json:"total_tensor_size"`
	SplittableTensors []int              `json:"splittable_tensors"` // Indices into CombinedTensors
}

// PipelineStageAnalysis contains pipeline stage information
type PipelineStageAnalysis struct {
	OptimalStages    int              `json:"optimal_stages"`
	StageBoundaries  []int            `json:"stage_boundaries"`
	StageWeights     []float64        `json:"stage_weights"`     // Computational weight per stage
	StageMemory      []int64          `json:"stage_memory"`      // Memory requirement per stage
	CommunicationMap map[int][]int    `json:"communication_map"` // Stage to stage communication
}

// MemoryRequirements contains memory usage estimates
type MemoryRequirements struct {
	ModelWeights     int64 `json:"model_weights"`     // Memory for model weights
	Activations      int64 `json:"activations"`       // Memory for activations
	Gradients        int64 `json:"gradients"`         // Memory for gradients (training)
	Optimizer        int64 `json:"optimizer"`         // Memory for optimizer states
	Overhead         int64 `json:"overhead"`          // Memory overhead
	TotalRequired    int64 `json:"total_required"`    // Total memory required
	MinNodeMemory    int64 `json:"min_node_memory"`   // Minimum memory per node
	RecommendedNodes int   `json:"recommended_nodes"` // Recommended number of nodes
}

// TensorInfo contains information about individual tensors
type TensorInfo struct {
	Name        string    `json:"name"`
	Shape       []int     `json:"shape"`
	Size        int64     `json:"size"`        // Size in bytes
	Type        TensorType `json:"type"`
	Splittable  bool      `json:"splittable"`  // Can this tensor be split?
	LayerIndex  int       `json:"layer_index"` // Which layer this tensor belongs to
}

// LayerType defines different types of layers
type LayerType string

const (
	LayerTypeEmbedding   LayerType = "embedding"
	LayerTypeAttention   LayerType = "attention"
	LayerTypeFFN         LayerType = "ffn"
	LayerTypeNormalization LayerType = "normalization"
	LayerTypeOutput      LayerType = "output"
)

// TensorType defines different types of tensors
type TensorType string

const (
	TensorTypeWeight     TensorType = "weight"
	TensorTypeBias       TensorType = "bias"
	TensorTypeActivation TensorType = "activation"
	TensorTypeAttention  TensorType = "attention"
)

// ModelAnalyzer provides model analysis capabilities
type ModelAnalyzer struct {
	// Cache for analyzed models
	analysisCache map[string]*ModelAnalysis
}

// NewModelAnalyzer creates a new model analyzer
func NewModelAnalyzer() *ModelAnalyzer {
	return &ModelAnalyzer{
		analysisCache: make(map[string]*ModelAnalysis),
	}
}

// AnalyzeModel analyzes a model and returns detailed architecture information
func (ma *ModelAnalyzer) AnalyzeModel(modelPath string, modelDetails *types.OllamaModelDetails) (*ModelAnalysis, error) {
	return ma.AnalyzeModelWithOptions(modelPath, modelDetails, nil)
}

// AnalyzeModelWithOptions analyzes a model with custom options
func (ma *ModelAnalyzer) AnalyzeModelWithOptions(modelPath string, modelDetails *types.OllamaModelDetails, options *PartitionOptions) (*ModelAnalysis, error) {
	if modelPath == "" {
		return nil, fmt.Errorf("model path cannot be empty")
	}

	// Check cache first
	if analysis, exists := ma.analysisCache[modelPath]; exists {
		return analysis, nil
	}

	modelName := extractModelName(modelPath)
	parameterSize := estimateParameterSize(modelName, modelDetails)

	// Perform comprehensive analysis
	// Determine model family for dimension inference
	family := determineModelFamily(modelName)

	layerInfo, err := ma.analyzeLayers(modelName, parameterSize)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze layers: %w", err)
	}

	tensorInfo := ma.analyzeTensors(layerInfo, family, parameterSize, options)
	stageInfo := ma.analyzePipelineStages(layerInfo)
	memoryReqs := ma.calculateMemoryRequirements(parameterSize, layerInfo, modelDetails, options)

	analysis := &ModelAnalysis{
		ModelName:     modelName,
		ParameterSize: parameterSize,
		LayerInfo:     layerInfo,
		TensorInfo:    tensorInfo,
		StageInfo:     stageInfo,
		MemoryReqs:    memoryReqs,
	}

	// Cache the analysis
	ma.analysisCache[modelPath] = analysis
	return analysis, nil
}

// analyzeLayers analyzes the layer structure of a model
func (ma *ModelAnalyzer) analyzeLayers(modelName string, parameterSize int64) (*LayerAnalysis, error) {
	// Use heuristics based on model family and size
	family := determineModelFamily(modelName)
	
	var totalLayers, transformerBlocks int
	var layerTypes map[int]LayerType

	switch family {
	case "llama":
		totalLayers = estimateLlamaLayers(parameterSize)
		transformerBlocks = totalLayers - 2 // Exclude embedding and output layers
		layerTypes = generateLlamaLayerTypes(totalLayers)
	case "gpt":
		totalLayers = estimateGPTLayers(parameterSize)
		transformerBlocks = totalLayers - 2
		layerTypes = generateGPTLayerTypes(totalLayers)
	case "bert":
		totalLayers = estimateBertLayers(parameterSize)
		transformerBlocks = totalLayers - 2
		layerTypes = generateBertLayerTypes(totalLayers)
	default:
		// Generic transformer estimation
		totalLayers = estimateGenericLayers(parameterSize)
		transformerBlocks = totalLayers - 2
		layerTypes = generateGenericLayerTypes(totalLayers)
	}

	attentionLayers := transformerBlocks // Each transformer block has attention
	ffnLayers := transformerBlocks       // Each transformer block has FFN

	// Calculate layer sizes and weights
	layerSizes := calculateLayerSizes(totalLayers, parameterSize, layerTypes)
	layerWeights := calculateLayerWeights(totalLayers, layerTypes)
	dependencies := generateLayerDependencies(totalLayers)

	return &LayerAnalysis{
		TotalLayers:       totalLayers,
		TransformerBlocks: transformerBlocks,
		AttentionLayers:   attentionLayers,
		FFNLayers:         ffnLayers,
		LayerSizes:        layerSizes,
		LayerWeights:      layerWeights,
		Dependencies:      dependencies,
		LayerTypes:        layerTypes,
	}, nil
}

// analyzeTensors analyzes tensor distribution possibilities
func (ma *ModelAnalyzer) analyzeTensors(layerInfo *LayerAnalysis, family string, paramSize int64, options *PartitionOptions) *TensorAnalysis {
	// Infer model dimensions based on family and parameters
	hiddenDim, ffnDim, _ := inferModelDims(family, paramSize, options)
	var attentionTensors, weightTensors, activationTensors []TensorInfo
	var combinedTensors []TensorInfo
	indexMap := make(map[string]int)
	var totalSize int64
	var splittableTensors []int

	combinedIndex := 0
	
	// Generate tensor information for each layer
	for i := 0; i < layerInfo.TotalLayers; i++ {
		layerType := layerInfo.LayerTypes[i]
		layerSize := layerInfo.LayerSizes[i]
		
		switch layerType {
		case LayerTypeAttention:
			// Query, Key, Value tensors using inferred dimensions
			qTensor := TensorInfo{
				Name:       fmt.Sprintf("layer_%d_attention_q", i),
				Shape:      []int{hiddenDim, hiddenDim},
				Size:       layerSize / 4,
				Type:       TensorTypeAttention,
				Splittable: true,
				LayerIndex: i,
			}
			kTensor := TensorInfo{
				Name:       fmt.Sprintf("layer_%d_attention_k", i),
				Shape:      []int{hiddenDim, hiddenDim},
				Size:       layerSize / 4,
				Type:       TensorTypeAttention,
				Splittable: true,
				LayerIndex: i,
			}
			vTensor := TensorInfo{
				Name:       fmt.Sprintf("layer_%d_attention_v", i),
				Shape:      []int{hiddenDim, hiddenDim},
				Size:       layerSize / 4,
				Type:       TensorTypeAttention,
				Splittable: true,
				LayerIndex: i,
			}
			oTensor := TensorInfo{
				Name:       fmt.Sprintf("layer_%d_attention_o", i),
				Shape:      []int{hiddenDim, hiddenDim},
				Size:       layerSize / 4,
				Type:       TensorTypeAttention,
				Splittable: true,
				LayerIndex: i,
			}
			
			attentionTensors = append(attentionTensors, qTensor, kTensor, vTensor, oTensor)
			totalSize += qTensor.Size + kTensor.Size + vTensor.Size + oTensor.Size
			
			// Add to combined tensors and update index map
			combinedTensors = append(combinedTensors, qTensor, kTensor, vTensor, oTensor)
			indexMap[qTensor.Name] = combinedIndex
			indexMap[kTensor.Name] = combinedIndex + 1
			indexMap[vTensor.Name] = combinedIndex + 2
			indexMap[oTensor.Name] = combinedIndex + 3
			
			// All attention tensors are splittable
			for j := 0; j < 4; j++ {
				splittableTensors = append(splittableTensors, combinedIndex+j)
			}
			combinedIndex += 4
			
		case LayerTypeFFN:
			// Feed-forward network tensors using inferred dimensions
			w1Tensor := TensorInfo{
				Name:       fmt.Sprintf("layer_%d_ffn_w1", i),
				Shape:      []int{hiddenDim, ffnDim},
				Size:       layerSize / 2,
				Type:       TensorTypeWeight,
				Splittable: true,
				LayerIndex: i,
			}
			w2Tensor := TensorInfo{
				Name:       fmt.Sprintf("layer_%d_ffn_w2", i),
				Shape:      []int{ffnDim, hiddenDim},
				Size:       layerSize / 2,
				Type:       TensorTypeWeight,
				Splittable: true,
				LayerIndex: i,
			}
			
			weightTensors = append(weightTensors, w1Tensor, w2Tensor)
			totalSize += w1Tensor.Size + w2Tensor.Size
			
			// Add to combined tensors and update index map
			combinedTensors = append(combinedTensors, w1Tensor, w2Tensor)
			indexMap[w1Tensor.Name] = combinedIndex
			indexMap[w2Tensor.Name] = combinedIndex + 1
			
			// FFN weights are splittable
			splittableTensors = append(splittableTensors, combinedIndex, combinedIndex+1)
			combinedIndex += 2
			
		default:
			// Other layer types (embedding, normalization, output)
			layerTensor := TensorInfo{
				Name:       fmt.Sprintf("layer_%d_weight", i),
				Shape:      []int{hiddenDim, hiddenDim},
				Size:       layerSize,
				Type:       TensorTypeWeight,
				Splittable: layerType != LayerTypeNormalization, // Normalization layers usually shouldn't be split
				LayerIndex: i,
			}
			
			weightTensors = append(weightTensors, layerTensor)
			totalSize += layerTensor.Size
			
			// Add to combined tensors and update index map
			combinedTensors = append(combinedTensors, layerTensor)
			indexMap[layerTensor.Name] = combinedIndex
			
			if layerTensor.Splittable {
				splittableTensors = append(splittableTensors, combinedIndex)
			}
			combinedIndex++
		}
		
		// Add activation tensors for each layer (intermediate activations)
		// These represent the intermediate values during forward pass
		activationTensor := TensorInfo{
			Name:       fmt.Sprintf("layer_%d_activation", i),
			Shape:      []int{1, hiddenDim}, // Using inferred hidden dimension
			Size:       int64(hiddenDim * 4), // 4 bytes per float32
			Type:       TensorTypeActivation,
			Splittable: false, // Activations typically not splittable
			LayerIndex: i,
		}
		activationTensors = append(activationTensors, activationTensor)
		totalSize += activationTensor.Size
		
		// Add to combined tensors and update index map
		combinedTensors = append(combinedTensors, activationTensor)
		indexMap[activationTensor.Name] = combinedIndex
		combinedIndex++
		
		// Add additional activation for attention layers
		if layerType == LayerTypeAttention {
			attentionActivation := TensorInfo{
				Name:       fmt.Sprintf("layer_%d_attention_activation", i),
				Shape:      []int{1, hiddenDim},
				Size:       int64(hiddenDim * 4),
				Type:       TensorTypeActivation,
				Splittable: false,
				LayerIndex: i,
			}
			activationTensors = append(activationTensors, attentionActivation)
			totalSize += attentionActivation.Size
			
			// Add to combined tensors and update index map
			combinedTensors = append(combinedTensors, attentionActivation)
			indexMap[attentionActivation.Name] = combinedIndex
			combinedIndex++
		}
	}

	return &TensorAnalysis{
		AttentionTensors:  attentionTensors,
		WeightTensors:     weightTensors,
		ActivationTensors: activationTensors,
		CombinedTensors:   combinedTensors,
		IndexMap:          indexMap,
		TotalTensorSize:   totalSize,
		SplittableTensors: splittableTensors,
	}
}

// analyzePipelineStages analyzes optimal pipeline stage configuration
func (ma *ModelAnalyzer) analyzePipelineStages(layerInfo *LayerAnalysis) *PipelineStageAnalysis {
	totalLayers := layerInfo.TotalLayers
	
	// Calculate optimal number of stages based on layer count
	optimalStages := calculateOptimalStages(totalLayers)
	
	// Generate stage boundaries
	stageBoundaries := generateStageBoundaries(totalLayers, optimalStages)
	
	// Calculate stage weights (computational load)
	stageWeights := calculateStageWeights(stageBoundaries, layerInfo.LayerWeights)
	
	// Calculate memory requirements per stage
	stageMemory := calculateStageMemory(stageBoundaries, layerInfo.LayerSizes)
	
	// Generate communication map
	communicationMap := generateCommunicationMap(optimalStages)

	return &PipelineStageAnalysis{
		OptimalStages:    optimalStages,
		StageBoundaries:  stageBoundaries,
		StageWeights:     stageWeights,
		StageMemory:      stageMemory,
		CommunicationMap: communicationMap,
	}
}

// calculateMemoryRequirements estimates memory requirements for the model
func (ma *ModelAnalyzer) calculateMemoryRequirements(parameterSize int64, layerInfo *LayerAnalysis, details *types.OllamaModelDetails, options *PartitionOptions) *MemoryRequirements {
	// Infer bytes per parameter based on precision
	bytesPerParam := inferBytesPerParam(details, options)

	// Check if we're in inference or training mode
	isInferenceMode := true // Default to inference
	if options != nil && options.CustomParams != nil {
		if mode, ok := options.CustomParams["mode"].(string); ok {
			isInferenceMode = (mode != "training")
		}
	}

	// Base memory for model weights
	modelWeights := parameterSize * 1_000_000_000 * bytesPerParam // Convert billions to actual count, then to bytes
	
	// Activation memory (estimate based on model size)
	// For inference, activations are smaller as we don't need to store intermediate values for backprop
	var activations int64
	if isInferenceMode {
		activations = modelWeights / 8 // Inference: activations typically 1/8 of weights
	} else {
		activations = modelWeights / 4 // Training: activations typically 1/4 of weights
	}
	
	// Gradients (only needed for training)
	var gradients int64
	if !isInferenceMode {
		gradients = modelWeights // Same size as weights for training
	}
	
	// Optimizer states (only needed for training)
	var optimizer int64
	if !isInferenceMode {
		optimizer = modelWeights * 2 // Adam requires ~2x model weights
	}
	
	// System overhead (10% of total)
	overhead := (modelWeights + activations + gradients + optimizer) / 10
	
	// Total memory required
	totalRequired := modelWeights + activations + gradients + optimizer + overhead
	
	// Minimum memory per node (should handle at least one layer)
	minLayerSize := int64(0)
	if len(layerInfo.LayerSizes) > 0 {
		minLayerSize = layerInfo.LayerSizes[0]
		for _, size := range layerInfo.LayerSizes {
			if size > minLayerSize {
				minLayerSize = size
			}
		}
	}
	minNodeMemory := minLayerSize * 2 // 2x largest layer for safety
	
	// Recommended number of nodes
	recommendedNodes := int((totalRequired / (8 * 1024 * 1024 * 1024)) + 1) // Assume 8GB per node
	if recommendedNodes < 1 {
		recommendedNodes = 1
	}

	return &MemoryRequirements{
		ModelWeights:     modelWeights,
		Activations:      activations,
		Gradients:        gradients,
		Optimizer:        optimizer,
		Overhead:         overhead,
		TotalRequired:    totalRequired,
		MinNodeMemory:    minNodeMemory,
		RecommendedNodes: recommendedNodes,
	}
}

// Helper functions for model analysis

func extractModelName(modelPath string) string {
	base := filepath.Base(modelPath)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	return strings.ToLower(name)
}

func estimateParameterSize(modelName string, details *types.OllamaModelDetails) int64 {
	// Helper function to extract number from string
	extractNumber := func(s string) (int64, bool) {
		// Remove whitespace and convert to lowercase
		s = strings.ToLower(strings.TrimSpace(s))
		
		// Try various patterns: "70B", "70b", "70 B", "70 b", "70B params", etc.
		patterns := []string{
			`(\d+)\s*[bB]`, // e.g., "70B", "70 B"
			`(\d+)\s*billion`, // e.g., "70 billion"
			`(\d+)\s*[mM]`, // e.g., "350M", "350 M"  
			`(\d+)\s*million`, // e.g., "350 million"
			`(\d+)\s*[kK]`, // e.g., "125K", "125 K"
			`(\d+)\s*thousand`, // e.g., "125 thousand"
			`(\d+)\.(\d+)\s*[bB]`, // e.g., "1.5B", "6.7B"
		}
		
		for _, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			matches := re.FindStringSubmatch(s)
			if len(matches) > 1 {
				// Handle decimal patterns (e.g., "6.7B")
				if len(matches) > 2 && matches[2] != "" {
					whole, _ := strconv.ParseInt(matches[1], 10, 64)
					decimal, _ := strconv.ParseInt(matches[2], 10, 64)
					// Approximate decimal billions (e.g., 6.7B ≈ 7B)
					if decimal >= 5 {
						whole++
					}
					return whole, true
				}
				
				// Handle integer patterns
				if val, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					// Convert based on suffix
					if strings.Contains(s, "m") || strings.Contains(s, "million") {
						return val / 1000, true // Convert millions to billions (approximation)
					} else if strings.Contains(s, "k") || strings.Contains(s, "thousand") {
						return val / 1000000, true // Convert thousands to billions (approximation)
					}
					return val, true // Already in billions
				}
			}
		}
		return 0, false
	}
	
	// First try details if available
	if details != nil && details.ParameterSize != "" {
		if size, found := extractNumber(details.ParameterSize); found {
			return size
		}
	}
	
	// Fallback to name-based estimation with improved pattern matching
	name := strings.ToLower(modelName)
	
	// Try to extract from model name using improved patterns
	if size, found := extractNumber(name); found {
		return size
	}
	
	// Legacy fallback patterns for common model names
	switch {
	case strings.Contains(name, "70b") || strings.Contains(name, "llama-2-70") || strings.Contains(name, "llama2-70"):
		return 70
	case strings.Contains(name, "33b") || strings.Contains(name, "vicuna-33"):
		return 33
	case strings.Contains(name, "13b") || strings.Contains(name, "llama-2-13") || strings.Contains(name, "llama2-13"):
		return 13
	case strings.Contains(name, "7b") || strings.Contains(name, "llama-2-7") || strings.Contains(name, "llama2-7") || strings.Contains(name, "mistral-7"):
		return 7
	case strings.Contains(name, "3b") || strings.Contains(name, "phi-3"):
		return 3
	case strings.Contains(name, "1.5b") || strings.Contains(name, "phi-1.5"):
		return 2 // Round up
	case strings.Contains(name, "1b") || strings.Contains(name, "phi-1"):
		return 1
	default:
		return 7 // Default assumption for unknown models
	}
}

func determineModelFamily(modelName string) string {
	name := strings.ToLower(modelName)
	switch {
	case strings.Contains(name, "llama"):
		return "llama"
	case strings.Contains(name, "gpt"):
		return "gpt"
	case strings.Contains(name, "bert"):
		return "bert"
	default:
		return "generic"
	}
}

func estimateLlamaLayers(parameterSize int64) int {
	// LLaMA layer count estimation based on parameter size
	switch {
	case parameterSize >= 65:
		return 80
	case parameterSize >= 30:
		return 60
	case parameterSize >= 13:
		return 40
	case parameterSize >= 7:
		return 32
	default:
		return 24
	}
}

func estimateGPTLayers(parameterSize int64) int {
	// GPT layer count estimation
	switch {
	case parameterSize >= 175:
		return 96
	case parameterSize >= 13:
		return 40
	case parameterSize >= 6:
		return 28
	default:
		return 12
	}
}

func estimateBertLayers(parameterSize int64) int {
	// BERT typically has fewer layers
	return 24
}

func estimateGenericLayers(parameterSize int64) int {
	// Generic estimation
	return int(parameterSize*4 + 12)
}

func generateLlamaLayerTypes(totalLayers int) map[int]LayerType {
	types := make(map[int]LayerType)
	types[0] = LayerTypeEmbedding
	
	for i := 1; i < totalLayers-1; i += 2 {
		types[i] = LayerTypeAttention
		if i+1 < totalLayers-1 {
			types[i+1] = LayerTypeFFN
		}
	}
	
	types[totalLayers-1] = LayerTypeOutput
	return types
}

func generateGPTLayerTypes(totalLayers int) map[int]LayerType {
	types := make(map[int]LayerType)
	types[0] = LayerTypeEmbedding
	
	for i := 1; i < totalLayers-1; i += 2 {
		types[i] = LayerTypeAttention
		if i+1 < totalLayers-1 {
			types[i+1] = LayerTypeFFN
		}
	}
	
	types[totalLayers-1] = LayerTypeOutput
	return types
}

func generateBertLayerTypes(totalLayers int) map[int]LayerType {
	return generateGenericLayerTypes(totalLayers)
}

func generateGenericLayerTypes(totalLayers int) map[int]LayerType {
	types := make(map[int]LayerType)
	types[0] = LayerTypeEmbedding
	
	for i := 1; i < totalLayers-1; i += 2 {
		types[i] = LayerTypeAttention
		if i+1 < totalLayers-1 {
			types[i+1] = LayerTypeFFN
		}
	}
	
	types[totalLayers-1] = LayerTypeOutput
	return types
}

func calculateLayerSizes(totalLayers int, parameterSize int64, layerTypes map[int]LayerType) []int64 {
	sizes := make([]int64, totalLayers)
	// Default to FP16 (2 bytes) for layer size calculation
	totalParams := parameterSize * 1_000_000_000 * 2 // Convert billions of parameters to bytes
	
	// Distribute parameters across layers based on type
	embeddingRatio := 0.1
	outputRatio := 0.1
	transformerRatio := 0.8
	
	embeddingSize := int64(float64(totalParams) * embeddingRatio)
	outputSize := int64(float64(totalParams) * outputRatio)
	
	transformerCount := 0
	for i := 0; i < totalLayers; i++ {
		if layerTypes[i] == LayerTypeAttention || layerTypes[i] == LayerTypeFFN {
			transformerCount++
		}
	}
	
	transformerLayerSize := int64(0)
	if transformerCount > 0 {
		transformerLayerSize = int64(float64(totalParams) * transformerRatio / float64(transformerCount))
	}
	
	for i := 0; i < totalLayers; i++ {
		switch layerTypes[i] {
		case LayerTypeEmbedding:
			sizes[i] = embeddingSize
		case LayerTypeOutput:
			sizes[i] = outputSize
		default:
			sizes[i] = transformerLayerSize
		}
	}
	
	return sizes
}

func calculateLayerWeights(totalLayers int, layerTypes map[int]LayerType) []float64 {
	weights := make([]float64, totalLayers)
	
	for i := 0; i < totalLayers; i++ {
		switch layerTypes[i] {
		case LayerTypeEmbedding:
			weights[i] = 0.5
		case LayerTypeAttention:
			weights[i] = 1.0
		case LayerTypeFFN:
			weights[i] = 0.8
		case LayerTypeNormalization:
			weights[i] = 0.1
		case LayerTypeOutput:
			weights[i] = 0.5
		default:
			weights[i] = 0.5
		}
	}
	
	return weights
}

func generateLayerDependencies(totalLayers int) [][]int {
	deps := make([][]int, totalLayers)
	
	for i := 0; i < totalLayers; i++ {
		if i > 0 {
			deps[i] = []int{i - 1} // Each layer depends on the previous one
		} else {
			deps[i] = []int{} // First layer has no dependencies
		}
	}
	
	return deps
}

func calculateOptimalStages(totalLayers int) int {
	// Heuristic for optimal stage count
	switch {
	case totalLayers >= 80:
		return 8
	case totalLayers >= 60:
		return 6
	case totalLayers >= 40:
		return 4
	case totalLayers >= 20:
		return 3
	default:
		return 2
	}
}

func generateStageBoundaries(totalLayers, stages int) []int {
	boundaries := make([]int, stages+1)
	boundaries[0] = 0
	
	layersPerStage := totalLayers / stages
	remainder := totalLayers % stages
	
	for i := 1; i <= stages; i++ {
		boundaries[i] = boundaries[i-1] + layersPerStage
		if i <= remainder {
			boundaries[i]++
		}
	}
	
	return boundaries
}

func calculateStageWeights(boundaries []int, layerWeights []float64) []float64 {
	stages := len(boundaries) - 1
	weights := make([]float64, stages)
	
	for i := 0; i < stages; i++ {
		start := boundaries[i]
		end := boundaries[i+1]
		
		for j := start; j < end && j < len(layerWeights); j++ {
			weights[i] += layerWeights[j]
		}
	}
	
	return weights
}

func calculateStageMemory(boundaries []int, layerSizes []int64) []int64 {
	stages := len(boundaries) - 1
	memory := make([]int64, stages)
	
	for i := 0; i < stages; i++ {
		start := boundaries[i]
		end := boundaries[i+1]
		
		for j := start; j < end && j < len(layerSizes); j++ {
			memory[i] += layerSizes[j]
		}
	}
	
	return memory
}

func generateCommunicationMap(stages int) map[int][]int {
	comm := make(map[int][]int)
	
	for i := 0; i < stages; i++ {
		if i > 0 {
			comm[i] = append(comm[i], i-1) // Receive from previous stage
		}
		if i < stages-1 {
			comm[i] = append(comm[i], i+1) // Send to next stage
		}
	}
	
	return comm
}

// inferBytesPerParam infers the number of bytes per parameter based on model details and options
func inferBytesPerParam(details *types.OllamaModelDetails, options *PartitionOptions) int64 {
	// Default to FP16 (2 bytes) for most models
	defaultBytes := int64(2)

	// Check options first for explicit override
	if options != nil && options.CustomParams != nil {
		// Check various parameter names for precision specification
		if dtype, ok := options.CustomParams["dtype"].(string); ok {
			return mapDtypeToBytes(dtype)
		}
		if precision, ok := options.CustomParams["precision"].(string); ok {
			return mapDtypeToBytes(precision)
		}
		if bytesPerParam, ok := options.CustomParams["bytes_per_param"].(int64); ok {
			return bytesPerParam
		}
		if bytesPerParam, ok := options.CustomParams["bytes_per_param"].(int); ok {
			return int64(bytesPerParam)
		}
	}

	// Check model details for quantization information
	if details != nil {
		// Check for quantization in format field (common location)
		if details.Format != "" && strings.Contains(strings.ToLower(details.Format), "q") {
			// Extract quantization info from format (e.g., "q4_0", "q8_0")
			if strings.Contains(strings.ToLower(details.Format), "q4") {
				return 4
			} else if strings.Contains(strings.ToLower(details.Format), "q8") {
				return 8
			}
		}
		// Could also check other fields like Families for hints
	}

	return defaultBytes
}

// mapDtypeToBytes maps data type strings to bytes per parameter
func mapDtypeToBytes(dtype string) int64 {
	switch strings.ToLower(dtype) {
	case "fp32", "float32", "f32":
		return 4
	case "fp16", "float16", "f16", "half":
		return 2
	case "bf16", "bfloat16":
		return 2
	case "int8", "i8":
		return 1
	case "int4", "i4":
		return 1 // Approximation, actually 0.5 bytes
	default:
		return 2 // Default to FP16
	}
}

// mapQuantizationToBytes maps quantization strings to bytes per parameter
func mapQuantizationToBytes(quantization string) int64 {
	q := strings.ToLower(quantization)
	switch {
	case strings.Contains(q, "q8"):
		return 1
	case strings.Contains(q, "q4"):
		return 1 // Approximation for 4-bit
	case strings.Contains(q, "q5"):
		return 1 // Approximation for 5-bit
	case strings.Contains(q, "q6"):
		return 1 // Approximation for 6-bit
	case strings.Contains(q, "f16") || strings.Contains(q, "fp16"):
		return 2
	case strings.Contains(q, "f32") || strings.Contains(q, "fp32"):
		return 4
	default:
		return 2 // Default to FP16
	}
}

// inferModelDims infers model dimensions based on family and parameter size
func inferModelDims(family string, paramB int64, options *PartitionOptions) (hidden int, ffn int, heads int) {
	// Check for overrides in options
	if options != nil && options.CustomParams != nil {
		if h, ok := options.CustomParams["hidden_dim"].(int); ok {
			hidden = h
		}
		if f, ok := options.CustomParams["ffn_dim"].(int); ok {
			ffn = f
		}
		if n, ok := options.CustomParams["num_heads"].(int); ok {
			heads = n
		}
		if hidden > 0 && ffn > 0 && heads > 0 {
			return hidden, ffn, heads
		}
	}

	// Family-based defaults
	switch family {
	case "llama":
		return inferLlamaDims(paramB)
	case "gpt":
		return inferGPTDims(paramB)
	case "bert":
		return inferBertDims(paramB)
	default:
		return inferGenericDims(paramB)
	}
}

// inferLlamaDims infers dimensions for LLaMA models
func inferLlamaDims(paramB int64) (hidden int, ffn int, heads int) {
	switch {
	case paramB >= 65: // 65B/70B models
		return 8192, 22016, 64
	case paramB >= 30: // 30B/33B models
		return 6656, 17920, 52
	case paramB >= 13: // 13B models
		return 5120, 13824, 40
	case paramB >= 7: // 7B models
		return 4096, 11008, 32
	case paramB >= 3: // 3B models
		return 3072, 8192, 24
	default: // Smaller models
		return 2048, 5504, 16
	}
}

// inferGPTDims infers dimensions for GPT models
func inferGPTDims(paramB int64) (hidden int, ffn int, heads int) {
	switch {
	case paramB >= 175: // GPT-3 175B
		return 12288, 49152, 96
	case paramB >= 13: // GPT-3 13B
		return 5120, 20480, 40
	case paramB >= 6: // GPT-J 6B
		return 4096, 16384, 16
	case paramB >= 2: // GPT-2 XL
		return 1600, 6400, 25
	default:
		return 768, 3072, 12
	}
}

// inferBertDims infers dimensions for BERT models
func inferBertDims(paramB int64) (hidden int, ffn int, heads int) {
	if paramB >= 1 {
		return 1024, 4096, 16 // BERT-Large
	}
	return 768, 3072, 12 // BERT-Base
}

// inferGenericDims provides generic dimension inference
func inferGenericDims(paramB int64) (hidden int, ffn int, heads int) {
	// Simple heuristic based on parameter count
	switch {
	case paramB >= 50:
		return 8192, 22016, 64
	case paramB >= 20:
		return 5120, 13824, 40
	case paramB >= 10:
		return 4096, 11008, 32
	case paramB >= 5:
		return 3072, 8192, 24
	case paramB >= 1:
		return 2048, 5504, 16
	default:
		return 1024, 2816, 8
	}
}
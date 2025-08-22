package multimodal

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MultiModalEngine handles multi-modal inference requests
type MultiModalEngine struct {
	textProcessor  TextProcessor
	imageProcessor ImageProcessor
	audioProcessor AudioProcessor
	videoProcessor VideoProcessor
	fusionEngine   *ModalityFusionEngine
	routingEngine  *ModalityRoutingEngine
	preprocessors  map[ModalityType]Preprocessor
	postprocessors map[ModalityType]Postprocessor
	ctx            context.Context
	cancel         context.CancelFunc
}

// ModalityType defines supported input modalities
type ModalityType string

const (
	ModalityText  ModalityType = "text"
	ModalityImage ModalityType = "image"
	ModalityAudio ModalityType = "audio"
	ModalityVideo ModalityType = "video"
)

// MultiModalRequest represents a multi-modal inference request
type MultiModalRequest struct {
	RequestID    string                   `json:"request_id"`
	Inputs       map[ModalityType][]Input `json:"inputs"`
	ModelID      string                   `json:"model_id"`
	Task         string                   `json:"task"`
	Parameters   map[string]interface{}   `json:"parameters"`
	FusionMode   FusionMode               `json:"fusion_mode"`
	OutputFormat string                   `json:"output_format"`
	Priority     int                      `json:"priority"`
	Timeout      time.Duration            `json:"timeout"`
	Metadata     map[string]interface{}   `json:"metadata"`
}

// Input represents a single input of any modality
type Input struct {
	Type       ModalityType           `json:"type"`
	Data       []byte                 `json:"data"`
	Format     string                 `json:"format"`
	Encoding   string                 `json:"encoding"`
	Metadata   map[string]interface{} `json:"metadata"`
	Timestamp  time.Time              `json:"timestamp"`
	SequenceID int                    `json:"sequence_id"`
}

// MultiModalResponse represents the inference response
type MultiModalResponse struct {
	RequestID      string                    `json:"request_id"`
	Outputs        map[ModalityType][]Output `json:"outputs"`
	FusedOutput    *Output                   `json:"fused_output,omitempty"`
	ProcessingTime time.Duration             `json:"processing_time"`
	ModelUsed      string                    `json:"model_used"`
	Confidence     float64                   `json:"confidence"`
	Metadata       map[string]interface{}    `json:"metadata"`
	Error          string                    `json:"error,omitempty"`
}

// Output represents inference output
type Output struct {
	Type       ModalityType           `json:"type"`
	Data       []byte                 `json:"data"`
	Format     string                 `json:"format"`
	Confidence float64                `json:"confidence"`
	Metadata   map[string]interface{} `json:"metadata"`
	Timestamp  time.Time              `json:"timestamp"`
}

// FusionMode defines how modalities are combined
type FusionMode string

const (
	FusionEarly  FusionMode = "early"  // Combine features before inference
	FusionLate   FusionMode = "late"   // Combine outputs after inference
	FusionHybrid FusionMode = "hybrid" // Adaptive fusion strategy
	FusionNone   FusionMode = "none"   // Process modalities separately
)

// Processor interfaces for different modalities
type TextProcessor interface {
	Process(ctx context.Context, inputs []Input, params map[string]interface{}) ([]Output, error)
	GetSupportedFormats() []string
	GetCapabilities() []string
}

type ImageProcessor interface {
	Process(ctx context.Context, inputs []Input, params map[string]interface{}) ([]Output, error)
	GetSupportedFormats() []string
	GetCapabilities() []string
}

type AudioProcessor interface {
	Process(ctx context.Context, inputs []Input, params map[string]interface{}) ([]Output, error)
	GetSupportedFormats() []string
	GetCapabilities() []string
}

type VideoProcessor interface {
	Process(ctx context.Context, inputs []Input, params map[string]interface{}) ([]Output, error)
	GetSupportedFormats() []string
	GetCapabilities() []string
}

// Preprocessor interface for input preprocessing
type Preprocessor interface {
	Preprocess(input *Input) (*Input, error)
	GetName() string
}

// Postprocessor interface for output postprocessing
type Postprocessor interface {
	Postprocess(output *Output) (*Output, error)
	GetName() string
}

// ModalityFusionEngine handles fusion of different modalities
type ModalityFusionEngine struct {
	strategies map[FusionMode]FusionStrategy
	weights    map[ModalityType]float64
	learner    *FusionLearner
}

// FusionStrategy interface for fusion algorithms
type FusionStrategy interface {
	Fuse(outputs map[ModalityType][]Output, weights map[ModalityType]float64) (*Output, error)
	GetName() string
	GetRequiredModalities() []ModalityType
}

// FusionLearner learns optimal fusion weights
type FusionLearner struct {
	model        FusionModel
	trainingData []*FusionTrainingExample
	weights      map[ModalityType]float64
	mutex        sync.RWMutex
}

// FusionModel interface for learning fusion weights
type FusionModel interface {
	Train(examples []*FusionTrainingExample) error
	Predict(features []float64) (map[ModalityType]float64, error)
	GetAccuracy() float64
}

// FusionTrainingExample represents training data for fusion learning
type FusionTrainingExample struct {
	Inputs      map[ModalityType][]Input
	GroundTruth *Output
	Context     map[string]interface{}
	Timestamp   time.Time
}

// ModalityRoutingEngine routes requests to appropriate processors
type ModalityRoutingEngine struct {
	routes        map[string]*ProcessingRoute
	loadBalancer  *LoadBalancer
	healthChecker *HealthChecker
}

// ProcessingRoute defines routing for a specific task/model combination
type ProcessingRoute struct {
	TaskType   string
	ModelID    string
	Modalities []ModalityType
	Processors map[ModalityType][]ProcessorEndpoint
	FusionMode FusionMode
	Priority   int
	Timeout    time.Duration
}

// ProcessorEndpoint represents a processing endpoint
type ProcessorEndpoint struct {
	ID           string
	Address      string
	Modality     ModalityType
	Capabilities []string
	Load         float64
	Healthy      bool
	LastCheck    time.Time
}

// LoadBalancer balances load across processor endpoints
type LoadBalancer struct {
	strategy LoadBalancingStrategy
	metrics  *LoadMetrics
}

// LoadBalancingStrategy interface for load balancing algorithms
type LoadBalancingStrategy interface {
	SelectEndpoint(endpoints []ProcessorEndpoint, request *MultiModalRequest) (*ProcessorEndpoint, error)
	GetName() string
}

// LoadMetrics tracks load balancing metrics
type LoadMetrics struct {
	RequestCount   map[string]int64
	ResponseTime   map[string]time.Duration
	ErrorRate      map[string]float64
	ThroughputRate map[string]float64
	mutex          sync.RWMutex
}

// HealthChecker monitors processor endpoint health
type HealthChecker struct {
	endpoints     map[string]*ProcessorEndpoint
	checkInterval time.Duration
	timeout       time.Duration
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewMultiModalEngine creates a new multi-modal inference engine
func NewMultiModalEngine(config *MultiModalConfig) (*MultiModalEngine, error) {
	ctx, cancel := context.WithCancel(context.Background())

	engine := &MultiModalEngine{
		textProcessor:  NewTextProcessor(config.TextConfig),
		imageProcessor: NewImageProcessor(config.ImageConfig),
		audioProcessor: NewAudioProcessor(config.AudioConfig),
		videoProcessor: NewVideoProcessor(config.VideoConfig),
		fusionEngine:   NewModalityFusionEngine(config.FusionConfig),
		routingEngine:  NewModalityRoutingEngine(config.RoutingConfig),
		preprocessors:  make(map[ModalityType]Preprocessor),
		postprocessors: make(map[ModalityType]Postprocessor),
		ctx:            ctx,
		cancel:         cancel,
	}

	// Initialize preprocessors and postprocessors
	engine.initializeProcessors(config)

	// Start background tasks
	go engine.healthCheckLoop()
	go engine.metricsCollectionLoop()

	return engine, nil
}

// ProcessRequest processes a multi-modal inference request
func (mme *MultiModalEngine) ProcessRequest(ctx context.Context, request *MultiModalRequest) (*MultiModalResponse, error) {
	startTime := time.Now()

	response := &MultiModalResponse{
		RequestID: request.RequestID,
		Outputs:   make(map[ModalityType][]Output),
		Metadata:  make(map[string]interface{}),
	}

	// Validate request
	if err := mme.validateRequest(request); err != nil {
		response.Error = err.Error()
		return response, err
	}

	// Preprocess inputs
	preprocessedInputs, err := mme.preprocessInputs(request.Inputs)
	if err != nil {
		response.Error = err.Error()
		return response, err
	}

	// Route and process each modality
	modalityOutputs := make(map[ModalityType][]Output)

	for modality, inputs := range preprocessedInputs {
		outputs, err := mme.processModality(ctx, modality, inputs, request.Parameters)
		if err != nil {
			response.Error = fmt.Sprintf("failed to process %s: %v", modality, err)
			return response, err
		}
		modalityOutputs[modality] = outputs
	}

	// Apply fusion if required
	if request.FusionMode != FusionNone && len(modalityOutputs) > 1 {
		fusedOutput, err := mme.fusionEngine.FuseOutputs(modalityOutputs, request.FusionMode)
		if err != nil {
			response.Error = fmt.Sprintf("fusion failed: %v", err)
			return response, err
		}
		response.FusedOutput = fusedOutput
	}

	// Postprocess outputs
	for modality, outputs := range modalityOutputs {
		postprocessedOutputs, err := mme.postprocessOutputs(modality, outputs)
		if err != nil {
			response.Error = fmt.Sprintf("postprocessing failed for %s: %v", modality, err)
			return response, err
		}
		response.Outputs[modality] = postprocessedOutputs
	}

	response.ProcessingTime = time.Since(startTime)
	response.ModelUsed = request.ModelID
	response.Confidence = mme.calculateOverallConfidence(response.Outputs)

	return response, nil
}

// validateRequest validates the multi-modal request
func (mme *MultiModalEngine) validateRequest(request *MultiModalRequest) error {
	if request.RequestID == "" {
		return fmt.Errorf("request ID is required")
	}

	if len(request.Inputs) == 0 {
		return fmt.Errorf("at least one input modality is required")
	}

	if request.ModelID == "" {
		return fmt.Errorf("model ID is required")
	}

	// Validate each input
	for modality, inputs := range request.Inputs {
		if len(inputs) == 0 {
			return fmt.Errorf("empty input list for modality %s", modality)
		}

		for i, input := range inputs {
			if len(input.Data) == 0 {
				return fmt.Errorf("empty data for input %d in modality %s", i, modality)
			}
		}
	}

	return nil
}

// preprocessInputs applies preprocessing to all inputs
func (mme *MultiModalEngine) preprocessInputs(inputs map[ModalityType][]Input) (map[ModalityType][]Input, error) {
	preprocessed := make(map[ModalityType][]Input)

	for modality, inputList := range inputs {
		preprocessor, exists := mme.preprocessors[modality]
		if !exists {
			// No preprocessor, use inputs as-is
			preprocessed[modality] = inputList
			continue
		}

		var processedList []Input
		for _, input := range inputList {
			processed, err := preprocessor.Preprocess(&input)
			if err != nil {
				return nil, fmt.Errorf("preprocessing failed for %s: %w", modality, err)
			}
			processedList = append(processedList, *processed)
		}
		preprocessed[modality] = processedList
	}

	return preprocessed, nil
}

// processModality processes inputs for a specific modality
func (mme *MultiModalEngine) processModality(ctx context.Context, modality ModalityType, inputs []Input, params map[string]interface{}) ([]Output, error) {
	switch modality {
	case ModalityText:
		return mme.textProcessor.Process(ctx, inputs, params)
	case ModalityImage:
		return mme.imageProcessor.Process(ctx, inputs, params)
	case ModalityAudio:
		return mme.audioProcessor.Process(ctx, inputs, params)
	case ModalityVideo:
		return mme.videoProcessor.Process(ctx, inputs, params)
	default:
		return nil, fmt.Errorf("unsupported modality: %s", modality)
	}
}

// postprocessOutputs applies postprocessing to outputs
func (mme *MultiModalEngine) postprocessOutputs(modality ModalityType, outputs []Output) ([]Output, error) {
	postprocessor, exists := mme.postprocessors[modality]
	if !exists {
		// No postprocessor, return outputs as-is
		return outputs, nil
	}

	var processedOutputs []Output
	for _, output := range outputs {
		processed, err := postprocessor.Postprocess(&output)
		if err != nil {
			return nil, fmt.Errorf("postprocessing failed for %s: %w", modality, err)
		}
		processedOutputs = append(processedOutputs, *processed)
	}

	return processedOutputs, nil
}

// calculateOverallConfidence calculates overall confidence from all outputs
func (mme *MultiModalEngine) calculateOverallConfidence(outputs map[ModalityType][]Output) float64 {
	var totalConfidence float64
	var count int

	for _, outputList := range outputs {
		for _, output := range outputList {
			totalConfidence += output.Confidence
			count++
		}
	}

	if count == 0 {
		return 0.0
	}

	return totalConfidence / float64(count)
}

// initializeProcessors initializes preprocessors and postprocessors
func (mme *MultiModalEngine) initializeProcessors(config *MultiModalConfig) {
	// Initialize preprocessors
	mme.preprocessors[ModalityText] = NewTextPreprocessor()
	mme.preprocessors[ModalityImage] = NewImagePreprocessor()
	mme.preprocessors[ModalityAudio] = NewAudioPreprocessor()
	mme.preprocessors[ModalityVideo] = NewVideoPreprocessor()

	// Initialize postprocessors
	mme.postprocessors[ModalityText] = NewTextPostprocessor()
	mme.postprocessors[ModalityImage] = NewImagePostprocessor()
	mme.postprocessors[ModalityAudio] = NewAudioPostprocessor()
	mme.postprocessors[ModalityVideo] = NewVideoPostprocessor()
}

// healthCheckLoop performs periodic health checks
func (mme *MultiModalEngine) healthCheckLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-mme.ctx.Done():
			return
		case <-ticker.C:
			mme.routingEngine.healthChecker.CheckAllEndpoints()
		}
	}
}

// metricsCollectionLoop collects performance metrics
func (mme *MultiModalEngine) metricsCollectionLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-mme.ctx.Done():
			return
		case <-ticker.C:
			mme.collectMetrics()
		}
	}
}

// collectMetrics collects and reports performance metrics
func (mme *MultiModalEngine) collectMetrics() {
	// TODO: Implement metrics collection
}

// Configuration types
type MultiModalConfig struct {
	TextConfig    *ProcessorConfig
	ImageConfig   *ProcessorConfig
	AudioConfig   *ProcessorConfig
	VideoConfig   *ProcessorConfig
	FusionConfig  *FusionConfig
	RoutingConfig *RoutingConfig
}

type ProcessorConfig struct {
	Enabled    bool
	ModelPath  string
	BatchSize  int
	Timeout    time.Duration
	MaxWorkers int
}

type FusionConfig struct {
	DefaultMode     FusionMode
	LearningEnabled bool
	WeightDecay     float64
	UpdateInterval  time.Duration
}

type RoutingConfig struct {
	LoadBalancing string
	HealthCheck   bool
	CheckInterval time.Duration
	Timeout       time.Duration
}

// Factory functions
func NewTextProcessor(config *ProcessorConfig) TextProcessor {
	return NewTextProcessorImpl(config)
}

func NewImageProcessor(config *ProcessorConfig) ImageProcessor {
	return NewImageProcessorImpl(config)
}

func NewAudioProcessor(config *ProcessorConfig) AudioProcessor {
	return NewAudioProcessorImpl(config)
}

func NewVideoProcessor(config *ProcessorConfig) VideoProcessor {
	return &DefaultVideoProcessor{} // Keep placeholder for now
}

func NewModalityFusionEngine(config *FusionConfig) *ModalityFusionEngine {
	engine := &ModalityFusionEngine{
		strategies: make(map[FusionMode]FusionStrategy),
		weights:    make(map[ModalityType]float64),
		learner:    NewFusionLearner(),
	}

	// Initialize default strategies
	engine.strategies[FusionLate] = &LateFusionStrategy{}
	engine.strategies[FusionEarly] = &EarlyFusionStrategy{}
	engine.strategies[FusionHybrid] = &HybridFusionStrategy{}

	// Initialize default weights
	engine.weights[ModalityText] = 0.4
	engine.weights[ModalityImage] = 0.3
	engine.weights[ModalityAudio] = 0.2
	engine.weights[ModalityVideo] = 0.1

	return engine
}

func NewModalityRoutingEngine(config *RoutingConfig) *ModalityRoutingEngine {
	return &ModalityRoutingEngine{
		routes:        make(map[string]*ProcessingRoute),
		loadBalancer:  NewLoadBalancer(),
		healthChecker: NewHealthChecker(),
	}
}

func NewFusionLearner() *FusionLearner {
	return &FusionLearner{
		weights: make(map[ModalityType]float64),
	}
}

func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		strategy: &RoundRobinStrategy{},
		metrics:  &LoadMetrics{},
	}
}

func NewHealthChecker() *HealthChecker {
	ctx, cancel := context.WithCancel(context.Background())
	return &HealthChecker{
		endpoints:     make(map[string]*ProcessorEndpoint),
		checkInterval: 30 * time.Second,
		timeout:       5 * time.Second,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Placeholder processor implementations
type DefaultTextProcessor struct{}

func (p *DefaultTextProcessor) Process(ctx context.Context, inputs []Input, params map[string]interface{}) ([]Output, error) {
	return []Output{}, nil
}
func (p *DefaultTextProcessor) GetSupportedFormats() []string { return []string{"text/plain"} }
func (p *DefaultTextProcessor) GetCapabilities() []string     { return []string{"text_generation"} }

type DefaultImageProcessor struct{}

func (p *DefaultImageProcessor) Process(ctx context.Context, inputs []Input, params map[string]interface{}) ([]Output, error) {
	return []Output{}, nil
}
func (p *DefaultImageProcessor) GetSupportedFormats() []string {
	return []string{"image/jpeg", "image/png"}
}
func (p *DefaultImageProcessor) GetCapabilities() []string { return []string{"image_classification"} }

type DefaultAudioProcessor struct{}

func (p *DefaultAudioProcessor) Process(ctx context.Context, inputs []Input, params map[string]interface{}) ([]Output, error) {
	return []Output{}, nil
}
func (p *DefaultAudioProcessor) GetSupportedFormats() []string {
	return []string{"audio/wav", "audio/mp3"}
}
func (p *DefaultAudioProcessor) GetCapabilities() []string { return []string{"speech_recognition"} }

type DefaultVideoProcessor struct{}

func (p *DefaultVideoProcessor) Process(ctx context.Context, inputs []Input, params map[string]interface{}) ([]Output, error) {
	return []Output{}, nil
}
func (p *DefaultVideoProcessor) GetSupportedFormats() []string { return []string{"video/mp4"} }
func (p *DefaultVideoProcessor) GetCapabilities() []string     { return []string{"video_analysis"} }

// Preprocessor implementations
func NewTextPreprocessor() Preprocessor  { return &TextPreprocessor{} }
func NewImagePreprocessor() Preprocessor { return &ImagePreprocessor{} }
func NewAudioPreprocessor() Preprocessor { return &AudioPreprocessor{} }
func NewVideoPreprocessor() Preprocessor { return &VideoPreprocessor{} }

func NewTextPostprocessor() Postprocessor  { return &TextPostprocessor{} }
func NewImagePostprocessor() Postprocessor { return &ImagePostprocessor{} }
func NewAudioPostprocessor() Postprocessor { return &AudioPostprocessor{} }
func NewVideoPostprocessor() Postprocessor { return &VideoPostprocessor{} }

type TextPreprocessor struct{}

func (p *TextPreprocessor) Preprocess(input *Input) (*Input, error) { return input, nil }
func (p *TextPreprocessor) GetName() string                         { return "text_preprocessor" }

type ImagePreprocessor struct{}

func (p *ImagePreprocessor) Preprocess(input *Input) (*Input, error) { return input, nil }
func (p *ImagePreprocessor) GetName() string                         { return "image_preprocessor" }

type AudioPreprocessor struct{}

func (p *AudioPreprocessor) Preprocess(input *Input) (*Input, error) { return input, nil }
func (p *AudioPreprocessor) GetName() string                         { return "audio_preprocessor" }

type VideoPreprocessor struct{}

func (p *VideoPreprocessor) Preprocess(input *Input) (*Input, error) { return input, nil }
func (p *VideoPreprocessor) GetName() string                         { return "video_preprocessor" }

type TextPostprocessor struct{}

func (p *TextPostprocessor) Postprocess(output *Output) (*Output, error) { return output, nil }
func (p *TextPostprocessor) GetName() string                             { return "text_postprocessor" }

type ImagePostprocessor struct{}

func (p *ImagePostprocessor) Postprocess(output *Output) (*Output, error) { return output, nil }
func (p *ImagePostprocessor) GetName() string                             { return "image_postprocessor" }

type AudioPostprocessor struct{}

func (p *AudioPostprocessor) Postprocess(output *Output) (*Output, error) { return output, nil }
func (p *AudioPostprocessor) GetName() string                             { return "audio_postprocessor" }

type VideoPostprocessor struct{}

func (p *VideoPostprocessor) Postprocess(output *Output) (*Output, error) { return output, nil }
func (p *VideoPostprocessor) GetName() string                             { return "video_postprocessor" }

// Load balancing strategy
type RoundRobinStrategy struct {
	counter int
}

func (s *RoundRobinStrategy) SelectEndpoint(endpoints []ProcessorEndpoint, request *MultiModalRequest) (*ProcessorEndpoint, error) {
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("no endpoints available")
	}

	s.counter = (s.counter + 1) % len(endpoints)
	return &endpoints[s.counter], nil
}

func (s *RoundRobinStrategy) GetName() string {
	return "round_robin"
}

// ModalityFusionEngine methods
func (mfe *ModalityFusionEngine) FuseOutputs(outputs map[ModalityType][]Output, mode FusionMode) (*Output, error) {
	strategy, exists := mfe.strategies[mode]
	if !exists {
		return nil, fmt.Errorf("fusion strategy not found for mode: %s", mode)
	}

	return strategy.Fuse(outputs, mfe.weights)
}

// HealthChecker methods
func (hc *HealthChecker) CheckAllEndpoints() {
	for _, endpoint := range hc.endpoints {
		hc.checkEndpoint(endpoint)
	}
}

func (hc *HealthChecker) checkEndpoint(endpoint *ProcessorEndpoint) {
	// TODO: Implement actual health check
	endpoint.Healthy = true
	endpoint.LastCheck = time.Now()
}

// LateFusionStrategy implements late fusion of outputs
type LateFusionStrategy struct{}

func (lfs *LateFusionStrategy) Fuse(outputs map[ModalityType][]Output, weights map[ModalityType]float64) (*Output, error) {
	if len(outputs) == 0 {
		return nil, fmt.Errorf("no outputs to fuse")
	}

	// Simple late fusion: combine text outputs with weighted confidence
	var combinedText string
	var totalConfidence float64
	var totalWeight float64

	for modality, outputList := range outputs {
		weight := weights[modality]
		if weight == 0 {
			weight = 1.0 / float64(len(outputs)) // Equal weight if not specified
		}

		for _, output := range outputList {
			if output.Type == ModalityText {
				combinedText += fmt.Sprintf("[%s]: %s\n", modality, string(output.Data))
				totalConfidence += output.Confidence * weight
				totalWeight += weight
			}
		}
	}

	if totalWeight == 0 {
		totalWeight = 1.0
	}

	fusedOutput := &Output{
		Type:       ModalityText,
		Data:       []byte(combinedText),
		Format:     "text/plain",
		Confidence: totalConfidence / totalWeight,
		Metadata: map[string]interface{}{
			"fusion_strategy":  "late",
			"modalities_fused": len(outputs),
		},
		Timestamp: time.Now(),
	}

	return fusedOutput, nil
}

func (lfs *LateFusionStrategy) GetName() string {
	return "late_fusion"
}

func (lfs *LateFusionStrategy) GetRequiredModalities() []ModalityType {
	return []ModalityType{} // Can work with any modalities
}

// EarlyFusionStrategy implements early fusion (placeholder)
type EarlyFusionStrategy struct{}

func (efs *EarlyFusionStrategy) Fuse(outputs map[ModalityType][]Output, weights map[ModalityType]float64) (*Output, error) {
	// For now, just delegate to late fusion
	lfs := &LateFusionStrategy{}
	return lfs.Fuse(outputs, weights)
}

func (efs *EarlyFusionStrategy) GetName() string {
	return "early_fusion"
}

func (efs *EarlyFusionStrategy) GetRequiredModalities() []ModalityType {
	return []ModalityType{}
}

// HybridFusionStrategy implements hybrid fusion (placeholder)
type HybridFusionStrategy struct{}

func (hfs *HybridFusionStrategy) Fuse(outputs map[ModalityType][]Output, weights map[ModalityType]float64) (*Output, error) {
	// For now, just delegate to late fusion
	lfs := &LateFusionStrategy{}
	return lfs.Fuse(outputs, weights)
}

func (hfs *HybridFusionStrategy) GetName() string {
	return "hybrid_fusion"
}

func (hfs *HybridFusionStrategy) GetRequiredModalities() []ModalityType {
	return []ModalityType{}
}

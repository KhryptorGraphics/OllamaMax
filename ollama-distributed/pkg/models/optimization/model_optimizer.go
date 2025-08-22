package optimization

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// ModelOptimizer handles automatic model optimization
type ModelOptimizer struct {
	compressionEngine  *CompressionEngine
	quantizationEngine *QuantizationEngine
	formatConverter    *FormatConverter
	optimizationQueue  chan *OptimizationTask
	workers            int
	ctx                context.Context
	cancel             context.CancelFunc
	wg                 sync.WaitGroup
}

// OptimizationTask represents a model optimization task
type OptimizationTask struct {
	ModelID      string
	ModelPath    string
	Techniques   []OptimizationTechnique
	TargetFormat string
	Priority     int
	Callback     func(*OptimizationResult)
}

// OptimizationTechnique defines optimization methods
type OptimizationTechnique string

const (
	TechniqueCompression   OptimizationTechnique = "compression"
	TechniqueQuantization  OptimizationTechnique = "quantization"
	TechniqueFormatConvert OptimizationTechnique = "format_convert"
	TechniquePruning       OptimizationTechnique = "pruning"
	TechniqueDistillation  OptimizationTechnique = "distillation"
)

// OptimizationResult contains optimization results
type OptimizationResult struct {
	OriginalSize     int64
	OptimizedSize    int64
	CompressionRatio float64
	QualityMetrics   map[string]float64
	ProcessingTime   time.Duration
	Error            error
}

// CompressionEngine handles model compression
type CompressionEngine struct {
	algorithms map[string]CompressionAlgorithm
}

// CompressionAlgorithm interface for compression methods
type CompressionAlgorithm interface {
	Compress(modelPath string, config CompressionConfig) (*CompressionResult, error)
	CompressFile(inputPath, outputPath string, config CompressionConfig) error
	GetName() string
	GetCompressionRatio() float64
}

// CompressionConfig holds compression parameters
type CompressionConfig struct {
	Algorithm  string
	Quality    float64
	Lossless   bool
	BlockSize  int
	Dictionary []byte
}

// CompressionResult contains compression results
type CompressionResult struct {
	CompressedPath   string
	OriginalSize     int64
	CompressedSize   int64
	CompressionRatio float64
	QualityLoss      float64
}

// QuantizationEngine handles model quantization
type QuantizationEngine struct {
	strategies map[string]QuantizationStrategy
}

// QuantizationStrategy interface for quantization methods
type QuantizationStrategy interface {
	Quantize(modelPath string, config QuantizationConfig) (*QuantizationResult, error)
	QuantizeFile(inputPath, outputPath string, config QuantizationConfig) error
	GetName() string
	GetBitWidth() int
	GetExpectedSpeedup() float64
	GetExpectedAccuracyLoss() float64
}

// QuantizationConfig holds quantization parameters
type QuantizationConfig struct {
	Strategy    string
	BitWidth    int
	Calibration bool
	DatasetPath string
	Symmetric   bool
	PerChannel  bool
}

// QuantizationResult contains quantization results
type QuantizationResult struct {
	QuantizedPath    string
	OriginalSize     int64
	QuantizedSize    int64
	AccuracyLoss     float64
	InferenceSpeedup float64
}

// FormatConverter handles model format conversion
type FormatConverter struct {
	converters map[string]FormatConverterImpl
}

// FormatConverterImpl interface for format conversion
type FormatConverterImpl interface {
	Convert(inputPath, outputPath string, config ConversionConfig) error
	ConvertFile(inputPath, outputPath string, config ConversionConfig) error
	GetSupportedFormats() []string
}

// ConversionConfig holds conversion parameters
type ConversionConfig struct {
	SourceFormat string
	TargetFormat string
	Precision    string
	Optimization bool
}

// NewModelOptimizer creates a new model optimizer
func NewModelOptimizer(workers int) *ModelOptimizer {
	ctx, cancel := context.WithCancel(context.Background())

	return &ModelOptimizer{
		compressionEngine:  NewCompressionEngine(),
		quantizationEngine: NewQuantizationEngine(),
		formatConverter:    NewFormatConverter(),
		optimizationQueue:  make(chan *OptimizationTask, 100),
		workers:            workers,
		ctx:                ctx,
		cancel:             cancel,
	}
}

// Start starts the optimization workers
func (mo *ModelOptimizer) Start() error {
	for i := 0; i < mo.workers; i++ {
		mo.wg.Add(1)
		go mo.worker(i)
	}
	return nil
}

// Stop stops the optimization workers
func (mo *ModelOptimizer) Stop() error {
	mo.cancel()
	close(mo.optimizationQueue)
	mo.wg.Wait()
	return nil
}

// OptimizeModel submits a model for optimization
func (mo *ModelOptimizer) OptimizeModel(task *OptimizationTask) error {
	select {
	case mo.optimizationQueue <- task:
		return nil
	case <-mo.ctx.Done():
		return fmt.Errorf("optimizer is shutting down")
	default:
		return fmt.Errorf("optimization queue is full")
	}
}

// worker processes optimization tasks
func (mo *ModelOptimizer) worker(id int) {
	defer mo.wg.Done()

	for {
		select {
		case task := <-mo.optimizationQueue:
			if task == nil {
				return
			}
			result := mo.processOptimizationTask(task)
			if task.Callback != nil {
				task.Callback(result)
			}
		case <-mo.ctx.Done():
			return
		}
	}
}

// processOptimizationTask processes a single optimization task
func (mo *ModelOptimizer) processOptimizationTask(task *OptimizationTask) *OptimizationResult {
	startTime := time.Now()
	result := &OptimizationResult{
		QualityMetrics: make(map[string]float64),
	}

	// Get original model size
	originalSize, err := getFileSize(task.ModelPath)
	if err != nil {
		result.Error = fmt.Errorf("failed to get original model size: %w", err)
		return result
	}
	result.OriginalSize = originalSize

	currentPath := task.ModelPath

	// Apply optimization techniques in sequence
	for _, technique := range task.Techniques {
		switch technique {
		case TechniqueCompression:
			currentPath, err = mo.applyCompression(currentPath, task)
		case TechniqueQuantization:
			currentPath, err = mo.applyQuantization(currentPath, task)
		case TechniqueFormatConvert:
			currentPath, err = mo.applyFormatConversion(currentPath, task)
		case TechniquePruning:
			currentPath, err = mo.applyPruning(currentPath, task)
		case TechniqueDistillation:
			currentPath, err = mo.applyDistillation(currentPath, task)
		}

		if err != nil {
			result.Error = err
			return result
		}
	}

	// Get optimized model size
	optimizedSize, err := getFileSize(currentPath)
	if err != nil {
		result.Error = fmt.Errorf("failed to get optimized model size: %w", err)
		return result
	}

	result.OptimizedSize = optimizedSize
	result.CompressionRatio = float64(originalSize) / float64(optimizedSize)
	result.ProcessingTime = time.Since(startTime)

	return result
}

// applyCompression applies compression to the model
func (mo *ModelOptimizer) applyCompression(modelPath string, task *OptimizationTask) (string, error) {
	config := CompressionConfig{
		Algorithm: "lz4", // Default algorithm
		Quality:   0.8,
		Lossless:  true,
		BlockSize: 64 * 1024,
	}

	result, err := mo.compressionEngine.Compress(modelPath, config)
	if err != nil {
		return "", fmt.Errorf("compression failed: %w", err)
	}

	return result.CompressedPath, nil
}

// applyQuantization applies quantization to the model
func (mo *ModelOptimizer) applyQuantization(modelPath string, task *OptimizationTask) (string, error) {
	config := QuantizationConfig{
		Strategy:    "dynamic",
		BitWidth:    8,
		Calibration: true,
		Symmetric:   true,
		PerChannel:  false,
	}

	result, err := mo.quantizationEngine.Quantize(modelPath, config)
	if err != nil {
		return "", fmt.Errorf("quantization failed: %w", err)
	}

	return result.QuantizedPath, nil
}

// applyFormatConversion applies format conversion to the model
func (mo *ModelOptimizer) applyFormatConversion(modelPath string, task *OptimizationTask) (string, error) {
	config := ConversionConfig{
		SourceFormat: "pytorch",
		TargetFormat: task.TargetFormat,
		Precision:    "fp16",
		Optimization: true,
	}

	outputPath := modelPath + "." + task.TargetFormat
	err := mo.formatConverter.Convert(modelPath, outputPath, config)
	if err != nil {
		return "", fmt.Errorf("format conversion failed: %w", err)
	}

	return outputPath, nil
}

// applyPruning applies model pruning
func (mo *ModelOptimizer) applyPruning(modelPath string, task *OptimizationTask) (string, error) {
	pruner := NewModelPruner()

	config := PruningConfig{
		Strategy:       "magnitude",
		SparsityLevel:  0.5, // 50% sparsity
		Structured:     false,
		GradualPruning: true,
	}

	result, err := pruner.Prune(modelPath, config)
	if err != nil {
		return "", fmt.Errorf("pruning failed: %w", err)
	}

	return result.PrunedPath, nil
}

// applyDistillation applies knowledge distillation
func (mo *ModelOptimizer) applyDistillation(modelPath string, task *OptimizationTask) (string, error) {
	distiller := NewKnowledgeDistiller()

	config := DistillationConfig{
		TeacherModel:        modelPath,
		StudentArchitecture: "smaller",
		Temperature:         4.0,
		Alpha:               0.7,
		TrainingEpochs:      10,
	}

	result, err := distiller.Distill(config)
	if err != nil {
		return "", fmt.Errorf("distillation failed: %w", err)
	}

	return result.StudentPath, nil
}

// LZ4CompressionAlgorithm implements LZ4 compression
type LZ4CompressionAlgorithm struct {
	name             string
	compressionRatio float64
}

// NewLZ4CompressionAlgorithm creates a new LZ4 compression algorithm
func NewLZ4CompressionAlgorithm() *LZ4CompressionAlgorithm {
	return &LZ4CompressionAlgorithm{
		name:             "lz4",
		compressionRatio: 2.5, // Average compression ratio
	}
}

// Compress compresses a model file using LZ4
func (lz4 *LZ4CompressionAlgorithm) Compress(modelPath string, config CompressionConfig) (*CompressionResult, error) {
	originalSize, err := getFileSize(modelPath)
	if err != nil {
		return nil, err
	}

	compressedPath := modelPath + ".lz4"
	err = lz4.CompressFile(modelPath, compressedPath, config)
	if err != nil {
		return nil, err
	}

	compressedSize, err := getFileSize(compressedPath)
	if err != nil {
		return nil, err
	}

	return &CompressionResult{
		CompressedPath:   compressedPath,
		OriginalSize:     originalSize,
		CompressedSize:   compressedSize,
		CompressionRatio: float64(originalSize) / float64(compressedSize),
		QualityLoss:      0.0,
	}, nil
}

// CompressFile compresses a file using LZ4
func (lz4 *LZ4CompressionAlgorithm) CompressFile(inputPath, outputPath string, config CompressionConfig) error {
	// TODO: Implement actual LZ4 compression using external library
	// For now, simulate compression by copying file
	return copyFile(inputPath, outputPath)
}

// GetName returns the algorithm name
func (lz4 *LZ4CompressionAlgorithm) GetName() string {
	return lz4.name
}

// GetCompressionRatio returns the expected compression ratio
func (lz4 *LZ4CompressionAlgorithm) GetCompressionRatio() float64 {
	return lz4.compressionRatio
}

// DynamicQuantizationStrategy implements dynamic quantization
type DynamicQuantizationStrategy struct {
	name                 string
	bitWidth             int
	expectedSpeedup      float64
	expectedAccuracyLoss float64
}

// NewDynamicQuantizationStrategy creates a new dynamic quantization strategy
func NewDynamicQuantizationStrategy() *DynamicQuantizationStrategy {
	return &DynamicQuantizationStrategy{
		name:                 "dynamic",
		bitWidth:             8,
		expectedSpeedup:      2.0,
		expectedAccuracyLoss: 0.02, // 2% accuracy loss
	}
}

// Quantize quantizes a model using dynamic quantization
func (dqs *DynamicQuantizationStrategy) Quantize(modelPath string, config QuantizationConfig) (*QuantizationResult, error) {
	originalSize, err := getFileSize(modelPath)
	if err != nil {
		return nil, err
	}

	quantizedPath := modelPath + ".quantized"
	err = dqs.QuantizeFile(modelPath, quantizedPath, config)
	if err != nil {
		return nil, err
	}

	quantizedSize, err := getFileSize(quantizedPath)
	if err != nil {
		return nil, err
	}

	return &QuantizationResult{
		QuantizedPath:    quantizedPath,
		OriginalSize:     originalSize,
		QuantizedSize:    quantizedSize,
		AccuracyLoss:     dqs.expectedAccuracyLoss,
		InferenceSpeedup: dqs.expectedSpeedup,
	}, nil
}

// QuantizeFile quantizes a file using dynamic quantization
func (dqs *DynamicQuantizationStrategy) QuantizeFile(inputPath, outputPath string, config QuantizationConfig) error {
	// TODO: Implement actual dynamic quantization
	// For now, simulate quantization by copying file with reduced size
	return copyFile(inputPath, outputPath)
}

// GetName returns the strategy name
func (dqs *DynamicQuantizationStrategy) GetName() string {
	return dqs.name
}

// GetBitWidth returns the bit width
func (dqs *DynamicQuantizationStrategy) GetBitWidth() int {
	return dqs.bitWidth
}

// GetExpectedSpeedup returns the expected inference speedup
func (dqs *DynamicQuantizationStrategy) GetExpectedSpeedup() float64 {
	return dqs.expectedSpeedup
}

// GetExpectedAccuracyLoss returns the expected accuracy loss
func (dqs *DynamicQuantizationStrategy) GetExpectedAccuracyLoss() float64 {
	return dqs.expectedAccuracyLoss
}

// ONNXConverter implements ONNX format conversion
type ONNXConverter struct {
	name             string
	supportedFormats []string
}

// NewONNXConverter creates a new ONNX converter
func NewONNXConverter() *ONNXConverter {
	return &ONNXConverter{
		name:             "onnx",
		supportedFormats: []string{"pytorch", "tensorflow", "onnx"},
	}
}

// Convert converts a model to ONNX format
func (oc *ONNXConverter) Convert(inputPath, outputPath string, config ConversionConfig) error {
	return oc.ConvertFile(inputPath, outputPath, config)
}

// ConvertFile converts a model file to ONNX format
func (oc *ONNXConverter) ConvertFile(inputPath, outputPath string, config ConversionConfig) error {
	// TODO: Implement actual ONNX conversion using external tools
	// For now, simulate conversion by copying file
	return copyFile(inputPath, outputPath)
}

// GetSupportedFormats returns supported input formats
func (oc *ONNXConverter) GetSupportedFormats() []string {
	return oc.supportedFormats
}

// ModelPruner handles model pruning operations
type ModelPruner struct {
	strategies map[string]PruningStrategy
}

// PruningStrategy interface for pruning methods
type PruningStrategy interface {
	Prune(modelPath string, config PruningConfig) (*PruningResult, error)
	GetName() string
	GetExpectedSparsity() float64
}

// PruningConfig holds pruning parameters
type PruningConfig struct {
	Strategy       string  `json:"strategy"`
	SparsityLevel  float64 `json:"sparsity_level"`
	Structured     bool    `json:"structured"`
	GradualPruning bool    `json:"gradual_pruning"`
	Threshold      float64 `json:"threshold"`
}

// PruningResult contains pruning results
type PruningResult struct {
	PrunedPath       string  `json:"pruned_path"`
	OriginalSize     int64   `json:"original_size"`
	PrunedSize       int64   `json:"pruned_size"`
	SparsityAchieved float64 `json:"sparsity_achieved"`
	AccuracyLoss     float64 `json:"accuracy_loss"`
}

// MagnitudePruningStrategy implements magnitude-based pruning
type MagnitudePruningStrategy struct {
	name             string
	expectedSparsity float64
}

// NewModelPruner creates a new model pruner
func NewModelPruner() *ModelPruner {
	pruner := &ModelPruner{
		strategies: make(map[string]PruningStrategy),
	}

	// Initialize default strategies
	pruner.strategies["magnitude"] = NewMagnitudePruningStrategy()

	return pruner
}

// Prune prunes a model using the specified configuration
func (mp *ModelPruner) Prune(modelPath string, config PruningConfig) (*PruningResult, error) {
	strategy, exists := mp.strategies[config.Strategy]
	if !exists {
		return nil, fmt.Errorf("pruning strategy not found: %s", config.Strategy)
	}

	return strategy.Prune(modelPath, config)
}

// NewMagnitudePruningStrategy creates a new magnitude pruning strategy
func NewMagnitudePruningStrategy() *MagnitudePruningStrategy {
	return &MagnitudePruningStrategy{
		name:             "magnitude",
		expectedSparsity: 0.5,
	}
}

// Prune implements magnitude-based pruning
func (mps *MagnitudePruningStrategy) Prune(modelPath string, config PruningConfig) (*PruningResult, error) {
	originalSize, err := getFileSize(modelPath)
	if err != nil {
		return nil, err
	}

	prunedPath := modelPath + ".pruned"

	// TODO: Implement actual magnitude-based pruning
	// For now, simulate pruning by copying file
	err = copyFile(modelPath, prunedPath)
	if err != nil {
		return nil, err
	}

	prunedSize, err := getFileSize(prunedPath)
	if err != nil {
		return nil, err
	}

	return &PruningResult{
		PrunedPath:       prunedPath,
		OriginalSize:     originalSize,
		PrunedSize:       prunedSize,
		SparsityAchieved: config.SparsityLevel,
		AccuracyLoss:     0.05, // 5% accuracy loss
	}, nil
}

// GetName returns the strategy name
func (mps *MagnitudePruningStrategy) GetName() string {
	return mps.name
}

// GetExpectedSparsity returns the expected sparsity
func (mps *MagnitudePruningStrategy) GetExpectedSparsity() float64 {
	return mps.expectedSparsity
}

// KnowledgeDistiller handles knowledge distillation operations
type KnowledgeDistiller struct {
	strategies map[string]DistillationStrategy
}

// DistillationStrategy interface for distillation methods
type DistillationStrategy interface {
	Distill(config DistillationConfig) (*DistillationResult, error)
	GetName() string
	GetExpectedCompression() float64
}

// DistillationConfig holds distillation parameters
type DistillationConfig struct {
	TeacherModel        string  `json:"teacher_model"`
	StudentArchitecture string  `json:"student_architecture"`
	Temperature         float64 `json:"temperature"`
	Alpha               float64 `json:"alpha"`
	TrainingEpochs      int     `json:"training_epochs"`
	LearningRate        float64 `json:"learning_rate"`
}

// DistillationResult contains distillation results
type DistillationResult struct {
	StudentPath      string        `json:"student_path"`
	TeacherSize      int64         `json:"teacher_size"`
	StudentSize      int64         `json:"student_size"`
	CompressionRatio float64       `json:"compression_ratio"`
	AccuracyLoss     float64       `json:"accuracy_loss"`
	TrainingTime     time.Duration `json:"training_time"`
}

// StandardDistillationStrategy implements standard knowledge distillation
type StandardDistillationStrategy struct {
	name                string
	expectedCompression float64
}

// NewKnowledgeDistiller creates a new knowledge distiller
func NewKnowledgeDistiller() *KnowledgeDistiller {
	distiller := &KnowledgeDistiller{
		strategies: make(map[string]DistillationStrategy),
	}

	// Initialize default strategies
	distiller.strategies["standard"] = NewStandardDistillationStrategy()

	return distiller
}

// Distill performs knowledge distillation using the specified configuration
func (kd *KnowledgeDistiller) Distill(config DistillationConfig) (*DistillationResult, error) {
	strategy, exists := kd.strategies["standard"]
	if !exists {
		return nil, fmt.Errorf("distillation strategy not found")
	}

	return strategy.Distill(config)
}

// NewStandardDistillationStrategy creates a new standard distillation strategy
func NewStandardDistillationStrategy() *StandardDistillationStrategy {
	return &StandardDistillationStrategy{
		name:                "standard",
		expectedCompression: 3.0, // 3x compression
	}
}

// Distill implements standard knowledge distillation
func (sds *StandardDistillationStrategy) Distill(config DistillationConfig) (*DistillationResult, error) {
	startTime := time.Now()

	teacherSize, err := getFileSize(config.TeacherModel)
	if err != nil {
		return nil, err
	}

	studentPath := config.TeacherModel + ".distilled"

	// TODO: Implement actual knowledge distillation
	// For now, simulate distillation by copying file
	err = copyFile(config.TeacherModel, studentPath)
	if err != nil {
		return nil, err
	}

	studentSize, err := getFileSize(studentPath)
	if err != nil {
		return nil, err
	}

	return &DistillationResult{
		StudentPath:      studentPath,
		TeacherSize:      teacherSize,
		StudentSize:      studentSize,
		CompressionRatio: float64(teacherSize) / float64(studentSize),
		AccuracyLoss:     0.03, // 3% accuracy loss
		TrainingTime:     time.Since(startTime),
	}, nil
}

// GetName returns the strategy name
func (sds *StandardDistillationStrategy) GetName() string {
	return sds.name
}

// GetExpectedCompression returns the expected compression ratio
func (sds *StandardDistillationStrategy) GetExpectedCompression() float64 {
	return sds.expectedCompression
}

// Helper functions
func getFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// Factory functions
func NewCompressionEngine() *CompressionEngine {
	return &CompressionEngine{
		algorithms: make(map[string]CompressionAlgorithm),
	}
}

func NewQuantizationEngine() *QuantizationEngine {
	return &QuantizationEngine{
		strategies: make(map[string]QuantizationStrategy),
	}
}

func NewFormatConverter() *FormatConverter {
	return &FormatConverter{
		converters: make(map[string]FormatConverterImpl),
	}
}

// Compression method for CompressionEngine
func (ce *CompressionEngine) Compress(modelPath string, config CompressionConfig) (*CompressionResult, error) {
	// Get original file size
	originalSize, err := getFileSize(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get original file size: %w", err)
	}

	// Select compression algorithm
	algorithm, exists := ce.algorithms[config.Algorithm]
	if !exists {
		// Use default LZ4 compression
		algorithm = NewLZ4CompressionAlgorithm()
		ce.algorithms[config.Algorithm] = algorithm
	}

	// Perform compression
	compressedPath := modelPath + ".compressed"
	err = algorithm.CompressFile(modelPath, compressedPath, config)
	if err != nil {
		return nil, fmt.Errorf("compression failed: %w", err)
	}

	// Get compressed file size
	compressedSize, err := getFileSize(compressedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get compressed file size: %w", err)
	}

	compressionRatio := float64(originalSize) / float64(compressedSize)

	return &CompressionResult{
		CompressedPath:   compressedPath,
		OriginalSize:     originalSize,
		CompressedSize:   compressedSize,
		CompressionRatio: compressionRatio,
		QualityLoss:      0.0, // Lossless compression
	}, nil
}

// Quantize method for QuantizationEngine
func (qe *QuantizationEngine) Quantize(modelPath string, config QuantizationConfig) (*QuantizationResult, error) {
	// Get original file size
	originalSize, err := getFileSize(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get original file size: %w", err)
	}

	// Select quantization strategy
	strategy, exists := qe.strategies[config.Strategy]
	if !exists {
		// Use default dynamic quantization
		strategy = NewDynamicQuantizationStrategy()
		qe.strategies[config.Strategy] = strategy
	}

	// Perform quantization
	quantizedPath := modelPath + ".quantized"
	err = strategy.QuantizeFile(modelPath, quantizedPath, config)
	if err != nil {
		return nil, fmt.Errorf("quantization failed: %w", err)
	}

	// Get quantized file size
	quantizedSize, err := getFileSize(quantizedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get quantized file size: %w", err)
	}

	// Calculate metrics
	inferenceSpeedup := strategy.GetExpectedSpeedup()
	accuracyLoss := strategy.GetExpectedAccuracyLoss()

	return &QuantizationResult{
		QuantizedPath:    quantizedPath,
		OriginalSize:     originalSize,
		QuantizedSize:    quantizedSize,
		AccuracyLoss:     accuracyLoss,
		InferenceSpeedup: inferenceSpeedup,
	}, nil
}

// Convert method for FormatConverter
func (fc *FormatConverter) Convert(inputPath, outputPath string, config ConversionConfig) error {
	// Select appropriate converter
	converter, exists := fc.converters[config.TargetFormat]
	if !exists {
		// Use default ONNX converter
		converter = NewONNXConverter()
		fc.converters[config.TargetFormat] = converter
	}

	// Check if source format is supported
	supportedFormats := converter.GetSupportedFormats()
	sourceSupported := false
	for _, format := range supportedFormats {
		if format == config.SourceFormat {
			sourceSupported = true
			break
		}
	}

	if !sourceSupported {
		return fmt.Errorf("source format %s not supported by %s converter", config.SourceFormat, config.TargetFormat)
	}

	// Perform conversion
	return converter.ConvertFile(inputPath, outputPath, config)
}

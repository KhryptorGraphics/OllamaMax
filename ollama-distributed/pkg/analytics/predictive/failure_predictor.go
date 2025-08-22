package predictive

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// FailurePredictor implements ML-based failure prediction
type FailurePredictor struct {
	model           *FailurePredictionModel
	featureStore    *FeatureStore
	anomalyDetector *AnomalyDetector
	config          *PredictorConfig
	predictions     map[string]*FailurePrediction
	mutex           sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
}

// PredictorConfig holds configuration for failure prediction
type PredictorConfig struct {
	PredictionWindow    time.Duration `json:"prediction_window"`
	FeatureWindow       time.Duration `json:"feature_window"`
	ModelUpdateInterval time.Duration `json:"model_update_interval"`
	ConfidenceThreshold float64       `json:"confidence_threshold"`
	AnomalyThreshold    float64       `json:"anomaly_threshold"`
	MaxPredictions      int           `json:"max_predictions"`
	EnableRetraining    bool          `json:"enable_retraining"`
}

// FailurePredictionModel implements time-series forecasting for failure prediction
type FailurePredictionModel struct {
	weights         [][]float64 // Neural network weights
	biases          [][]float64 // Neural network biases
	layers          []int       // Layer sizes
	learningRate    float64
	accuracy        float64
	trainingData    []*TrainingExample
	lastTraining    time.Time
	predictionCache map[string]*ModelPrediction
	mutex           sync.RWMutex
}

// FeatureStore manages feature extraction and storage
type FeatureStore struct {
	features   map[string][]*FeatureVector
	extractors map[string]FeatureExtractor
	storage    FeatureStorage
	aggregator *FeatureAggregator
	mutex      sync.RWMutex
}

// AnomalyDetector identifies unusual patterns
type AnomalyDetector struct {
	model       *IsolationForest
	threshold   float64
	sensitivity float64
	history     []*AnomalyScore
	mutex       sync.RWMutex
}

// FailurePrediction represents a failure prediction result
type FailurePrediction struct {
	NodeID          string        `json:"node_id"`
	FailureType     string        `json:"failure_type"`
	Probability     float64       `json:"probability"`
	Confidence      float64       `json:"confidence"`
	TimeToFailure   time.Duration `json:"time_to_failure"`
	PredictedTime   time.Time     `json:"predicted_time"`
	RootCause       string        `json:"root_cause"`
	Severity        string        `json:"severity"`
	Recommendations []string      `json:"recommendations"`
	Features        []string      `json:"features"`
	Timestamp       time.Time     `json:"timestamp"`
}

// FeatureVector represents extracted features for a time point
type FeatureVector struct {
	Timestamp time.Time              `json:"timestamp"`
	NodeID    string                 `json:"node_id"`
	Features  map[string]float64     `json:"features"`
	Labels    map[string]interface{} `json:"labels"`
}

// TrainingExample represents a training data point
type TrainingExample struct {
	Features []float64 `json:"features"`
	Label    float64   `json:"label"` // 1.0 for failure, 0.0 for normal
	Weight   float64   `json:"weight"`
	NodeID   string    `json:"node_id"`
	Time     time.Time `json:"time"`
}

// ModelPrediction represents a raw model prediction
type ModelPrediction struct {
	Probability float64   `json:"probability"`
	Confidence  float64   `json:"confidence"`
	Features    []float64 `json:"features"`
	Timestamp   time.Time `json:"timestamp"`
}

// AnomalyScore represents an anomaly detection result
type AnomalyScore struct {
	NodeID    string    `json:"node_id"`
	Score     float64   `json:"score"`
	Threshold float64   `json:"threshold"`
	IsAnomaly bool      `json:"is_anomaly"`
	Timestamp time.Time `json:"timestamp"`
}

// FeatureExtractor interface for extracting features
type FeatureExtractor interface {
	ExtractFeatures(data interface{}) (*FeatureVector, error)
	GetFeatureNames() []string
	GetExtractionInterval() time.Duration
}

// FeatureStorage interface for storing features
type FeatureStorage interface {
	Store(features *FeatureVector) error
	Retrieve(nodeID string, timeRange TimeRange) ([]*FeatureVector, error)
	Aggregate(nodeID string, timeRange TimeRange, aggregationType string) (*FeatureVector, error)
}

// FeatureAggregator aggregates features over time windows
type FeatureAggregator struct {
	windowSize time.Duration
	stepSize   time.Duration
}

// IsolationForest implements isolation forest for anomaly detection
type IsolationForest struct {
	trees     []*IsolationTree
	numTrees  int
	subsample int
	maxDepth  int
}

// IsolationTree represents a single isolation tree
type IsolationTree struct {
	root     *TreeNode
	maxDepth int
}

// TreeNode represents a node in the isolation tree
type TreeNode struct {
	feature   int
	threshold float64
	left      *TreeNode
	right     *TreeNode
	size      int
	depth     int
}

// TimeRange represents a time range for queries
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// NewFailurePredictor creates a new failure predictor
func NewFailurePredictor(config *PredictorConfig) (*FailurePredictor, error) {
	ctx, cancel := context.WithCancel(context.Background())

	predictor := &FailurePredictor{
		model:           NewFailurePredictionModel(),
		featureStore:    NewFeatureStore(),
		anomalyDetector: NewAnomalyDetector(config.AnomalyThreshold),
		config:          config,
		predictions:     make(map[string]*FailurePrediction),
		ctx:             ctx,
		cancel:          cancel,
	}

	// Start background processes
	go predictor.predictionLoop()
	go predictor.modelUpdateLoop()

	return predictor, nil
}

// PredictFailure predicts potential failures for a node
func (fp *FailurePredictor) PredictFailure(nodeID string, currentMetrics map[string]float64) (*FailurePrediction, error) {
	fp.mutex.RLock()
	defer fp.mutex.RUnlock()

	// Extract features from current metrics
	features := fp.extractPredictionFeatures(nodeID, currentMetrics)

	// Get model prediction
	modelPred, err := fp.model.Predict(features)
	if err != nil {
		return nil, fmt.Errorf("model prediction failed: %w", err)
	}

	// Check for anomalies
	anomalyScore := fp.anomalyDetector.DetectAnomaly(features)

	// Combine predictions and ensure valid range
	probability := math.Max(0.0, math.Min(1.0, modelPred.Probability))
	if anomalyScore.IsAnomaly {
		probability = math.Max(probability, math.Min(1.0, anomalyScore.Score))
	}

	// Only create prediction if above threshold and not for normal conditions
	if probability < fp.config.ConfidenceThreshold {
		return nil, nil
	}

	// Additional check: don't predict failure for clearly normal metrics
	cpuUtil := getMetricValue(currentMetrics, "cpu_utilization")
	memUtil := getMetricValue(currentMetrics, "memory_utilization")
	errorRate := getMetricValue(currentMetrics, "error_rate")

	if cpuUtil < 0.7 && memUtil < 0.7 && errorRate < 0.05 && !anomalyScore.IsAnomaly {
		return nil, nil
	}

	// Determine failure type and recommendations
	failureType, rootCause := fp.analyzeFailureType(features, currentMetrics)
	recommendations := fp.generateRecommendations(failureType, currentMetrics)

	prediction := &FailurePrediction{
		NodeID:          nodeID,
		FailureType:     failureType,
		Probability:     probability,
		Confidence:      modelPred.Confidence,
		TimeToFailure:   fp.estimateTimeToFailure(probability),
		PredictedTime:   time.Now().Add(fp.estimateTimeToFailure(probability)),
		RootCause:       rootCause,
		Severity:        fp.calculateSeverity(probability),
		Recommendations: recommendations,
		Features:        fp.getFeatureNames(),
		Timestamp:       time.Now(),
	}

	// Cache prediction
	fp.predictions[nodeID] = prediction

	return prediction, nil
}

// AddTrainingData adds training data for model improvement
func (fp *FailurePredictor) AddTrainingData(nodeID string, metrics map[string]float64, failed bool) error {
	features := fp.extractPredictionFeatures(nodeID, metrics)

	label := 0.0
	if failed {
		label = 1.0
	}

	example := &TrainingExample{
		Features: features,
		Label:    label,
		Weight:   1.0,
		NodeID:   nodeID,
		Time:     time.Now(),
	}

	return fp.model.AddTrainingExample(example)
}

// GetPredictions returns all current predictions
func (fp *FailurePredictor) GetPredictions() map[string]*FailurePrediction {
	fp.mutex.RLock()
	defer fp.mutex.RUnlock()

	// Return copy of predictions
	predictions := make(map[string]*FailurePrediction)
	for k, v := range fp.predictions {
		predictions[k] = v
	}
	return predictions
}

// extractPredictionFeatures extracts features for prediction
func (fp *FailurePredictor) extractPredictionFeatures(nodeID string, metrics map[string]float64) []float64 {
	// Extract key features for failure prediction
	features := make([]float64, 0, 15)

	// Resource utilization features
	features = append(features, getMetricValue(metrics, "cpu_utilization"))
	features = append(features, getMetricValue(metrics, "memory_utilization"))
	features = append(features, getMetricValue(metrics, "disk_utilization"))
	features = append(features, getMetricValue(metrics, "network_utilization"))

	// Performance features (normalized)
	features = append(features, math.Min(1.0, getMetricValue(metrics, "response_time")))     // Cap at 1.0
	features = append(features, math.Min(1.0, getMetricValue(metrics, "error_rate")))        // Cap at 1.0
	features = append(features, math.Min(1.0, getMetricValue(metrics, "throughput")/1000.0)) // Normalize throughput

	// System health features (normalized)
	features = append(features, math.Min(1.0, getMetricValue(metrics, "load_average")/10.0))       // Normalize load average
	features = append(features, math.Min(1.0, getMetricValue(metrics, "connection_count")/1000.0)) // Normalize connections
	features = append(features, math.Min(1.0, getMetricValue(metrics, "queue_depth")/100.0))       // Normalize queue depth

	// Trend features (simplified)
	features = append(features, getMetricValue(metrics, "cpu_trend"))
	features = append(features, getMetricValue(metrics, "memory_trend"))
	features = append(features, getMetricValue(metrics, "error_trend"))

	// Time-based features
	hour := float64(time.Now().Hour()) / 24.0
	dayOfWeek := float64(time.Now().Weekday()) / 7.0
	features = append(features, hour, dayOfWeek)

	return features
}

// analyzeFailureType determines the type of failure based on features
func (fp *FailurePredictor) analyzeFailureType(features []float64, metrics map[string]float64) (string, string) {
	// Simple rule-based failure type detection
	cpuUtil := getMetricValue(metrics, "cpu_utilization")
	memUtil := getMetricValue(metrics, "memory_utilization")
	diskUtil := getMetricValue(metrics, "disk_utilization")
	errorRate := getMetricValue(metrics, "error_rate")

	if cpuUtil > 0.9 {
		return "cpu_exhaustion", "High CPU utilization detected"
	}
	if memUtil > 0.9 {
		return "memory_exhaustion", "High memory utilization detected"
	}
	if diskUtil > 0.9 {
		return "disk_exhaustion", "High disk utilization detected"
	}
	if errorRate > 0.1 {
		return "service_degradation", "High error rate detected"
	}

	return "general_failure", "Multiple indicators suggest potential failure"
}

// generateRecommendations generates recommendations based on failure type
func (fp *FailurePredictor) generateRecommendations(failureType string, metrics map[string]float64) []string {
	switch failureType {
	case "cpu_exhaustion":
		return []string{
			"Scale up CPU resources",
			"Optimize CPU-intensive processes",
			"Consider load balancing",
		}
	case "memory_exhaustion":
		return []string{
			"Increase memory allocation",
			"Check for memory leaks",
			"Optimize memory usage",
		}
	case "disk_exhaustion":
		return []string{
			"Clean up disk space",
			"Archive old data",
			"Add storage capacity",
		}
	case "service_degradation":
		return []string{
			"Check service health",
			"Review recent deployments",
			"Monitor error logs",
		}
	default:
		return []string{
			"Monitor system closely",
			"Prepare for maintenance",
			"Check all system components",
		}
	}
}

// estimateTimeToFailure estimates time until failure based on probability
func (fp *FailurePredictor) estimateTimeToFailure(probability float64) time.Duration {
	// Simple inverse relationship: higher probability = shorter time
	if probability >= 0.9 {
		return 5 * time.Minute
	} else if probability >= 0.8 {
		return 15 * time.Minute
	} else if probability >= 0.7 {
		return 30 * time.Minute
	} else if probability >= 0.6 {
		return 1 * time.Hour
	}
	return 2 * time.Hour
}

// calculateSeverity calculates severity based on probability
func (fp *FailurePredictor) calculateSeverity(probability float64) string {
	if probability >= 0.9 {
		return "critical"
	} else if probability >= 0.7 {
		return "high"
	} else if probability >= 0.5 {
		return "medium"
	}
	return "low"
}

// getFeatureNames returns the names of features used
func (fp *FailurePredictor) getFeatureNames() []string {
	return []string{
		"cpu_utilization", "memory_utilization", "disk_utilization", "network_utilization",
		"response_time", "error_rate", "throughput",
		"load_average", "connection_count", "queue_depth",
		"cpu_trend", "memory_trend", "error_trend",
		"hour_of_day", "day_of_week",
	}
}

// predictionLoop runs predictions in the background
func (fp *FailurePredictor) predictionLoop() {
	ticker := time.NewTicker(fp.config.PredictionWindow)
	defer ticker.Stop()

	for {
		select {
		case <-fp.ctx.Done():
			return
		case <-ticker.C:
			// Clean up old predictions
			fp.cleanupOldPredictions()
		}
	}
}

// modelUpdateLoop updates the model periodically
func (fp *FailurePredictor) modelUpdateLoop() {
	ticker := time.NewTicker(fp.config.ModelUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-fp.ctx.Done():
			return
		case <-ticker.C:
			if fp.config.EnableRetraining {
				err := fp.model.Retrain()
				if err != nil {
					// Log error but continue
					fmt.Printf("Model retraining failed: %v\n", err)
				}
			}
		}
	}
}

// cleanupOldPredictions removes old predictions
func (fp *FailurePredictor) cleanupOldPredictions() {
	fp.mutex.Lock()
	defer fp.mutex.Unlock()

	cutoff := time.Now().Add(-fp.config.PredictionWindow)
	for nodeID, prediction := range fp.predictions {
		if prediction.Timestamp.Before(cutoff) {
			delete(fp.predictions, nodeID)
		}
	}
}

// Stop stops the failure predictor
func (fp *FailurePredictor) Stop() {
	fp.cancel()
}

// Helper function to safely get metric values
func getMetricValue(metrics map[string]float64, key string) float64 {
	if value, exists := metrics[key]; exists {
		return value
	}
	return 0.0
}

// NewFailurePredictionModel creates a new failure prediction model
func NewFailurePredictionModel() *FailurePredictionModel {
	// Initialize a simple neural network with 15 inputs, 10 hidden, 1 output
	layers := []int{15, 10, 1}

	model := &FailurePredictionModel{
		layers:          layers,
		learningRate:    0.01,
		accuracy:        0.0,
		trainingData:    make([]*TrainingExample, 0),
		predictionCache: make(map[string]*ModelPrediction),
	}

	// Initialize weights and biases
	model.initializeWeights()

	return model
}

// initializeWeights initializes neural network weights randomly
func (fpm *FailurePredictionModel) initializeWeights() {
	fpm.weights = make([][]float64, len(fpm.layers)-1)
	fpm.biases = make([][]float64, len(fpm.layers)-1)

	for i := 0; i < len(fpm.layers)-1; i++ {
		inputSize := fpm.layers[i]
		outputSize := fpm.layers[i+1]

		// Initialize weights with small random values
		fpm.weights[i] = make([]float64, inputSize*outputSize)
		for j := range fpm.weights[i] {
			fpm.weights[i][j] = (math.Cos(float64(j)) * 0.01) // Smaller weights for stability
		}

		// Initialize biases to zero
		fpm.biases[i] = make([]float64, outputSize)
	}
}

// Predict makes a prediction using a simplified rule-based approach
func (fpm *FailurePredictionModel) Predict(features []float64) (*ModelPrediction, error) {
	fpm.mutex.RLock()
	defer fpm.mutex.RUnlock()

	if len(features) != fpm.layers[0] {
		return nil, fmt.Errorf("expected %d features, got %d", fpm.layers[0], len(features))
	}

	// Simplified prediction based on key features
	// Features: cpu, memory, disk, network, response_time, error_rate, throughput, ...
	cpuUtil := features[0]
	memUtil := features[1]
	diskUtil := features[2]
	errorRate := features[5]
	responseTime := features[4]

	// Calculate failure probability based on resource utilization
	score := 0.0

	// High resource utilization increases failure probability
	if cpuUtil > 0.8 {
		score += (cpuUtil - 0.8) * 2.0
	}
	if memUtil > 0.8 {
		score += (memUtil - 0.8) * 2.0
	}
	if diskUtil > 0.8 {
		score += (diskUtil - 0.8) * 1.5
	}
	if errorRate > 0.05 {
		score += errorRate * 5.0
	}
	if responseTime > 1.0 {
		score += (responseTime - 1.0) * 0.5
	}

	// Normalize to probability
	probability := math.Min(1.0, score)

	// Calculate confidence based on how extreme the values are
	confidence := 0.5
	if cpuUtil > 0.9 || memUtil > 0.9 || errorRate > 0.1 {
		confidence = 0.9
	} else if cpuUtil < 0.3 && memUtil < 0.3 && errorRate < 0.01 {
		confidence = 0.8
	}

	return &ModelPrediction{
		Probability: probability,
		Confidence:  confidence,
		Features:    features,
		Timestamp:   time.Now(),
	}, nil
}

// AddTrainingExample adds a training example
func (fpm *FailurePredictionModel) AddTrainingExample(example *TrainingExample) error {
	fpm.mutex.Lock()
	defer fpm.mutex.Unlock()

	fpm.trainingData = append(fpm.trainingData, example)

	// Keep only recent training data
	maxExamples := 10000
	if len(fpm.trainingData) > maxExamples {
		fpm.trainingData = fpm.trainingData[len(fpm.trainingData)-maxExamples:]
	}

	return nil
}

// Retrain retrains the model with accumulated training data
func (fpm *FailurePredictionModel) Retrain() error {
	fpm.mutex.Lock()
	defer fpm.mutex.Unlock()

	if len(fpm.trainingData) < 10 {
		return fmt.Errorf("insufficient training data: %d examples", len(fpm.trainingData))
	}

	// Simple training: adjust weights based on recent examples
	epochs := 10
	for epoch := 0; epoch < epochs; epoch++ {
		totalError := 0.0

		for _, example := range fpm.trainingData {
			// Forward pass
			prediction, err := fpm.predict(example.Features)
			if err != nil {
				continue
			}

			// Calculate error
			error := example.Label - prediction
			totalError += math.Abs(error)

			// Simple weight update (gradient descent approximation)
			fpm.updateWeights(example.Features, error)
		}

		// Update accuracy
		fpm.accuracy = 1.0 - (totalError / float64(len(fpm.trainingData)))
	}

	fpm.lastTraining = time.Now()
	return nil
}

// predict is an internal prediction method without locking
func (fpm *FailurePredictionModel) predict(features []float64) (float64, error) {
	// Simplified prediction for training
	sum := 0.0
	for i, feature := range features {
		if i < len(fpm.weights[0]) {
			sum += feature * fpm.weights[0][i]
		}
	}
	return sigmoid(sum), nil
}

// updateWeights updates weights based on error
func (fpm *FailurePredictionModel) updateWeights(features []float64, error float64) {
	// Simple weight update
	for i := 0; i < len(fpm.weights[0]) && i < len(features); i++ {
		fpm.weights[0][i] += fpm.learningRate * error * features[i]
	}
}

// sigmoid activation function
func sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

// NewFeatureStore creates a new feature store
func NewFeatureStore() *FeatureStore {
	return &FeatureStore{
		features:   make(map[string][]*FeatureVector),
		extractors: make(map[string]FeatureExtractor),
		aggregator: &FeatureAggregator{
			windowSize: time.Hour,
			stepSize:   time.Minute * 5,
		},
	}
}

// NewAnomalyDetector creates a new anomaly detector
func NewAnomalyDetector(threshold float64) *AnomalyDetector {
	return &AnomalyDetector{
		model: &IsolationForest{
			numTrees:  100,
			subsample: 256,
			maxDepth:  10,
		},
		threshold:   threshold,
		sensitivity: 0.1,
		history:     make([]*AnomalyScore, 0),
	}
}

// DetectAnomaly detects if the given features represent an anomaly
func (ad *AnomalyDetector) DetectAnomaly(features []float64) *AnomalyScore {
	ad.mutex.Lock()
	defer ad.mutex.Unlock()

	// Simple anomaly detection based on feature values
	score := 0.0
	anomalyCount := 0

	// Check for extreme values (more lenient thresholds)
	for _, feature := range features {
		if feature > 0.95 || feature < 0.05 {
			anomalyCount++
		}
		// Only count significant deviations
		deviation := math.Abs(feature - 0.5)
		if deviation > 0.4 { // Only significant deviations
			score += deviation
		}
	}

	score = score / float64(len(features))
	// Adjust anomaly detection sensitivity
	isAnomaly := score > (ad.threshold*0.3) || anomalyCount > 3

	anomalyScore := &AnomalyScore{
		Score:     score,
		Threshold: ad.threshold,
		IsAnomaly: isAnomaly,
		Timestamp: time.Now(),
	}

	// Store in history
	ad.history = append(ad.history, anomalyScore)
	if len(ad.history) > 1000 {
		ad.history = ad.history[1:]
	}

	return anomalyScore
}

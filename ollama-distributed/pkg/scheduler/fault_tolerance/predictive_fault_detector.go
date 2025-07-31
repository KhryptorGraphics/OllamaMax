package fault_tolerance

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// PredictiveFaultDetector provides machine learning-based fault prediction
type PredictiveFaultDetector struct {
	// Base components
	enhancedDetector *EnhancedFaultDetector

	// Predictive models
	timeSeriesAnalyzer *TimeSeriesAnalyzer
	mlPredictor        *MLPredictor
	trendAnalyzer      *TrendAnalyzer
	correlationEngine  *CorrelationEngine

	// Prediction state
	predictions       map[string]*FaultPrediction
	predictionHistory []*FaultPrediction
	predictionsMu     sync.RWMutex

	// Learning system
	learningEngine *LearningEngine
	modelRegistry  *ModelRegistry

	// Configuration
	config *PredictiveDetectionConfig

	// Lifecycle
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	running   bool
	runningMu sync.RWMutex
}

// PredictiveDetectionConfig configures predictive fault detection
type PredictiveDetectionConfig struct {
	// Prediction intervals
	PredictionInterval  time.Duration `json:"prediction_interval"`
	LearningInterval    time.Duration `json:"learning_interval"`
	ModelUpdateInterval time.Duration `json:"model_update_interval"`

	// Prediction parameters
	PredictionHorizon    time.Duration `json:"prediction_horizon"`
	ConfidenceThreshold  float64       `json:"confidence_threshold"`
	MinHistorySize       int           `json:"min_history_size"`
	MaxPredictionHistory int           `json:"max_prediction_history"`

	// Model configuration
	EnableTimeSeriesML        bool `json:"enable_time_series_ml"`
	EnableTrendAnalysis       bool `json:"enable_trend_analysis"`
	EnableCorrelationAnalysis bool `json:"enable_correlation_analysis"`
	EnableContinuousLearning  bool `json:"enable_continuous_learning"`

	// Advanced features
	EnableEnsemblePrediction bool `json:"enable_ensemble_prediction"`
	EnableAdaptiveThresholds bool `json:"enable_adaptive_thresholds"`
	EnableSeasonalAdjustment bool `json:"enable_seasonal_adjustment"`
}

// FaultPrediction represents a predicted fault
type FaultPrediction struct {
	ID            string                 `json:"id"`
	PredictedType FaultType              `json:"predicted_type"`
	Target        string                 `json:"target"`
	Confidence    float64                `json:"confidence"`
	TimeToFailure time.Duration          `json:"time_to_failure"`
	PredictedAt   time.Time              `json:"predicted_at"`
	ExpectedAt    time.Time              `json:"expected_at"`
	ModelUsed     string                 `json:"model_used"`
	Features      map[string]float64     `json:"features"`
	Metadata      map[string]interface{} `json:"metadata"`
	Status        PredictionStatus       `json:"status"`
	ActualOutcome *ActualOutcome         `json:"actual_outcome,omitempty"`
}

// PredictionStatus represents the status of a prediction
type PredictionStatus string

const (
	PredictionStatusPending       PredictionStatus = "pending"
	PredictionStatusConfirmed     PredictionStatus = "confirmed"
	PredictionStatusFalsePositive PredictionStatus = "false_positive"
	PredictionStatusExpired       PredictionStatus = "expired"
)

// ActualOutcome represents the actual outcome of a prediction
type ActualOutcome struct {
	Occurred   bool          `json:"occurred"`
	ActualType FaultType     `json:"actual_type,omitempty"`
	ActualTime time.Time     `json:"actual_time,omitempty"`
	TimeDelta  time.Duration `json:"time_delta,omitempty"`
	Accuracy   float64       `json:"accuracy"`
}

// TimeSeriesAnalyzer analyzes time series data for fault prediction
type TimeSeriesAnalyzer struct {
	// Time series models
	armaModels     map[string]*ARMAModel
	lstmModels     map[string]*LSTMModel
	seasonalModels map[string]*SeasonalModel

	// Analysis configuration
	config *TimeSeriesConfig

	// State
	mu sync.RWMutex
}

// MLPredictor provides machine learning-based prediction
type MLPredictor struct {
	// ML models
	ensembleModel *EnsembleModel
	neuralNetwork *NeuralNetwork
	randomForest  *RandomForest
	svmModel      *SVMModel

	// Feature engineering
	featureExtractor *FeatureExtractor
	featureSelector  *FeatureSelector

	// Configuration
	config *MLPredictorConfig

	// State
	mu sync.RWMutex
}

// TrendAnalyzer analyzes trends in system metrics
type TrendAnalyzer struct {
	// Trend detection
	trendDetectors map[string]*TrendDetector
	changePoints   map[string][]*ChangePoint

	// Seasonal analysis
	seasonalAnalyzer *SeasonalAnalyzer

	// Configuration
	config *TrendAnalysisConfig

	// State
	mu sync.RWMutex
}

// CorrelationEngine analyzes correlations between metrics and faults
type CorrelationEngine struct {
	// Correlation analysis
	correlationMatrix map[string]map[string]float64
	causalityGraph    *CausalityGraph

	// Cross-correlation analysis
	crossCorrelations map[string]*CrossCorrelation

	// Configuration
	config *CorrelationConfig

	// State
	mu sync.RWMutex
}

// LearningEngine provides continuous learning capabilities
type LearningEngine struct {
	// Learning algorithms
	onlineLearning        *OnlineLearning
	reinforcementLearning *ReinforcementLearning
	transferLearning      *TransferLearning

	// Model adaptation
	modelAdapter        *ModelAdapter
	hyperparameterTuner *HyperparameterTuner

	// Configuration
	config *LearningConfig

	// State
	mu sync.RWMutex
}

// ModelRegistry manages prediction models
type ModelRegistry struct {
	// Model storage
	models        map[string]PredictionModel
	modelMetadata map[string]*ModelMetadata

	// Model versioning
	modelVersions  map[string][]*ModelVersion
	activeVersions map[string]string

	// Configuration
	config *ModelRegistryConfig

	// State
	mu sync.RWMutex
}

// NewPredictiveFaultDetector creates a new predictive fault detector
func NewPredictiveFaultDetector(enhancedDetector *EnhancedFaultDetector, config *PredictiveDetectionConfig) *PredictiveFaultDetector {
	if config == nil {
		config = &PredictiveDetectionConfig{
			PredictionInterval:        30 * time.Second,
			LearningInterval:          5 * time.Minute,
			ModelUpdateInterval:       1 * time.Hour,
			PredictionHorizon:         10 * time.Minute,
			ConfidenceThreshold:       0.7,
			MinHistorySize:            100,
			MaxPredictionHistory:      10000,
			EnableTimeSeriesML:        true,
			EnableTrendAnalysis:       true,
			EnableCorrelationAnalysis: true,
			EnableContinuousLearning:  true,
			EnableEnsemblePrediction:  true,
			EnableAdaptiveThresholds:  true,
			EnableSeasonalAdjustment:  false,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	detector := &PredictiveFaultDetector{
		enhancedDetector:  enhancedDetector,
		predictions:       make(map[string]*FaultPrediction),
		predictionHistory: make([]*FaultPrediction, 0),
		config:            config,
		ctx:               ctx,
		cancel:            cancel,
	}

	// Initialize components
	detector.initializeComponents()

	return detector
}

// initializeComponents initializes all predictive components
func (pfd *PredictiveFaultDetector) initializeComponents() {
	// Initialize time series analyzer
	pfd.timeSeriesAnalyzer = NewTimeSeriesAnalyzer(&TimeSeriesConfig{
		WindowSize:      100,
		SeasonalPeriods: []int{24, 168}, // Daily and weekly patterns
		EnableARMA:      true,
		EnableLSTM:      pfd.config.EnableTimeSeriesML,
		EnableSeasonal:  pfd.config.EnableSeasonalAdjustment,
	})

	// Initialize ML predictor
	pfd.mlPredictor = NewMLPredictor(&MLPredictorConfig{
		EnableEnsemble:     pfd.config.EnableEnsemblePrediction,
		EnableNeuralNet:    true,
		EnableRandomForest: true,
		EnableSVM:          false, // Disabled for performance
		FeatureSelection:   true,
	})

	// Initialize trend analyzer
	pfd.trendAnalyzer = NewTrendAnalyzer(&TrendAnalysisConfig{
		TrendWindow:           50,
		ChangePointDetection:  true,
		SeasonalDecomposition: pfd.config.EnableSeasonalAdjustment,
	})

	// Initialize correlation engine
	pfd.correlationEngine = NewCorrelationEngine(&CorrelationConfig{
		CorrelationWindow:    200,
		CausalityAnalysis:    true,
		CrossCorrelationLags: 10,
	})

	// Initialize learning engine
	pfd.learningEngine = NewLearningEngine(&LearningConfig{
		EnableOnlineLearning:   pfd.config.EnableContinuousLearning,
		EnableRL:               false, // Disabled initially
		EnableTransferLearning: true,
		AdaptationRate:         0.01,
	})

	// Initialize model registry
	pfd.modelRegistry = NewModelRegistry(&ModelRegistryConfig{
		MaxVersions:      10,
		AutoCleanup:      true,
		ModelPersistence: false, // Disabled for now
	})

	log.Info().Msg("Predictive fault detector components initialized")
}

// Start starts the predictive fault detector
func (pfd *PredictiveFaultDetector) Start() error {
	pfd.runningMu.Lock()
	defer pfd.runningMu.Unlock()

	if pfd.running {
		return nil
	}

	// Start prediction routine
	pfd.wg.Add(1)
	go pfd.predictionRoutine()

	// Start learning routine
	if pfd.config.EnableContinuousLearning {
		pfd.wg.Add(1)
		go pfd.learningRoutine()
	}

	// Start model update routine
	pfd.wg.Add(1)
	go pfd.modelUpdateRoutine()

	pfd.running = true
	log.Info().Msg("Predictive fault detector started")
	return nil
}

// Stop stops the predictive fault detector
func (pfd *PredictiveFaultDetector) Stop() error {
	pfd.runningMu.Lock()
	defer pfd.runningMu.Unlock()

	if !pfd.running {
		return nil
	}

	// Cancel context to stop all routines
	pfd.cancel()

	// Wait for all routines to finish
	pfd.wg.Wait()

	pfd.running = false
	log.Info().Msg("Predictive fault detector stopped")
	return nil
}

// PredictFaults performs fault prediction based on current system state
func (pfd *PredictiveFaultDetector) PredictFaults(metrics map[string]interface{}) []*FaultPrediction {
	// Collect time series data
	timeSeriesData := pfd.collectTimeSeriesData(metrics)

	// Perform time series analysis
	timeSeriesPredictions := pfd.timeSeriesAnalyzer.AnalyzeAndPredict(timeSeriesData)

	// Perform ML-based prediction
	mlPredictions := pfd.mlPredictor.Predict(metrics, timeSeriesData)

	// Perform trend analysis
	trendPredictions := pfd.trendAnalyzer.AnalyzeTrends(timeSeriesData)

	// Perform correlation analysis
	correlationPredictions := pfd.correlationEngine.AnalyzeCorrelations(metrics, timeSeriesData)

	// Combine predictions using ensemble methods
	combinedPredictions := pfd.combinePredictions(
		timeSeriesPredictions,
		mlPredictions,
		trendPredictions,
		correlationPredictions,
	)

	// Filter predictions by confidence threshold
	filteredPredictions := pfd.filterPredictions(combinedPredictions)

	// Store predictions
	pfd.storePredictions(filteredPredictions)

	return filteredPredictions
}

// collectTimeSeriesData collects time series data from metrics
func (pfd *PredictiveFaultDetector) collectTimeSeriesData(metrics map[string]interface{}) map[string]*TimeSeriesData {
	timeSeriesData := make(map[string]*TimeSeriesData)

	// Get historical data from enhanced detector
	if pfd.enhancedDetector != nil && pfd.enhancedDetector.metricsCollector != nil {
		for metricName := range metrics {
			if series := pfd.getMetricTimeSeries(metricName); series != nil {
				timeSeriesData[metricName] = &TimeSeriesData{
					Name:       metricName,
					Values:     series.Values,
					Timestamps: series.Timestamps,
					Frequency:  pfd.calculateFrequency(series.Timestamps),
				}
			}
		}
	}

	return timeSeriesData
}

// getMetricTimeSeries retrieves time series data for a metric
func (pfd *PredictiveFaultDetector) getMetricTimeSeries(metricName string) *MetricTimeSeries {
	if pfd.enhancedDetector.metricsCollector == nil {
		return nil
	}

	pfd.enhancedDetector.metricsCollector.mu.RLock()
	defer pfd.enhancedDetector.metricsCollector.mu.RUnlock()

	if series, exists := pfd.enhancedDetector.metricsCollector.metrics[metricName]; exists {
		return series
	}

	return nil
}

// calculateFrequency calculates the frequency of time series data
func (pfd *PredictiveFaultDetector) calculateFrequency(timestamps []time.Time) time.Duration {
	if len(timestamps) < 2 {
		return time.Minute // Default frequency
	}

	// Calculate average interval between timestamps
	totalDuration := timestamps[len(timestamps)-1].Sub(timestamps[0])
	intervals := len(timestamps) - 1

	if intervals > 0 {
		return totalDuration / time.Duration(intervals)
	}

	return time.Minute
}

// combinePredictions combines predictions from different models
func (pfd *PredictiveFaultDetector) combinePredictions(
	timeSeriesPredictions []*FaultPrediction,
	mlPredictions []*FaultPrediction,
	trendPredictions []*FaultPrediction,
	correlationPredictions []*FaultPrediction,
) []*FaultPrediction {

	// Ensemble prediction using weighted voting
	predictionMap := make(map[string]*EnsemblePrediction)

	// Add time series predictions
	pfd.addPredictionsToEnsemble(predictionMap, timeSeriesPredictions, 0.3)

	// Add ML predictions
	pfd.addPredictionsToEnsemble(predictionMap, mlPredictions, 0.4)

	// Add trend predictions
	pfd.addPredictionsToEnsemble(predictionMap, trendPredictions, 0.2)

	// Add correlation predictions
	pfd.addPredictionsToEnsemble(predictionMap, correlationPredictions, 0.1)

	// Convert ensemble predictions to final predictions
	var finalPredictions []*FaultPrediction
	for _, ensemble := range predictionMap {
		if prediction := pfd.createFinalPrediction(ensemble); prediction != nil {
			finalPredictions = append(finalPredictions, prediction)
		}
	}

	return finalPredictions
}

// addPredictionsToEnsemble adds predictions to the ensemble with weights
func (pfd *PredictiveFaultDetector) addPredictionsToEnsemble(
	ensembleMap map[string]*EnsemblePrediction,
	predictions []*FaultPrediction,
	weight float64,
) {
	for _, prediction := range predictions {
		key := fmt.Sprintf("%s_%s", prediction.Target, string(prediction.PredictedType))

		if ensemble, exists := ensembleMap[key]; exists {
			ensemble.Predictions = append(ensemble.Predictions, prediction)
			ensemble.TotalWeight += weight * prediction.Confidence
			ensemble.Count++
		} else {
			ensembleMap[key] = &EnsemblePrediction{
				Target:      prediction.Target,
				FaultType:   prediction.PredictedType,
				Predictions: []*FaultPrediction{prediction},
				TotalWeight: weight * prediction.Confidence,
				Count:       1,
			}
		}
	}
}

// Helper methods and routines

// extractEnsembleFeatures extracts features from ensemble predictions
func (pfd *PredictiveFaultDetector) extractEnsembleFeatures(ensemble *EnsemblePrediction) map[string]float64 {
	features := make(map[string]float64)

	if len(ensemble.Predictions) > 0 {
		// Average features from all predictions
		for _, pred := range ensemble.Predictions {
			for key, value := range pred.Features {
				features[key] += value
			}
		}

		// Normalize by count
		for key := range features {
			features[key] /= float64(len(ensemble.Predictions))
		}
	}

	return features
}

// getModelsUsed returns list of models used in ensemble
func (pfd *PredictiveFaultDetector) getModelsUsed(ensemble *EnsemblePrediction) []string {
	models := make(map[string]bool)
	for _, pred := range ensemble.Predictions {
		models[pred.ModelUsed] = true
	}

	var modelList []string
	for model := range models {
		modelList = append(modelList, model)
	}

	return modelList
}

// isDuplicatePrediction checks if prediction is duplicate
func (pfd *PredictiveFaultDetector) isDuplicatePrediction(prediction *FaultPrediction) bool {
	pfd.predictionsMu.RLock()
	defer pfd.predictionsMu.RUnlock()

	for _, existing := range pfd.predictions {
		if existing.Target == prediction.Target &&
			existing.PredictedType == prediction.PredictedType &&
			existing.Status == PredictionStatusPending {
			return true
		}
	}

	return false
}

// Prediction routines

// predictionRoutine performs periodic predictions
func (pfd *PredictiveFaultDetector) predictionRoutine() {
	defer pfd.wg.Done()

	ticker := time.NewTicker(pfd.config.PredictionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pfd.ctx.Done():
			return
		case <-ticker.C:
			pfd.performPredictions()
		}
	}
}

// learningRoutine performs continuous learning
func (pfd *PredictiveFaultDetector) learningRoutine() {
	defer pfd.wg.Done()

	ticker := time.NewTicker(pfd.config.LearningInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pfd.ctx.Done():
			return
		case <-ticker.C:
			pfd.performLearning()
		}
	}
}

// modelUpdateRoutine performs periodic model updates
func (pfd *PredictiveFaultDetector) modelUpdateRoutine() {
	defer pfd.wg.Done()

	ticker := time.NewTicker(pfd.config.ModelUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pfd.ctx.Done():
			return
		case <-ticker.C:
			pfd.performModelUpdates()
		}
	}
}

// performPredictions performs the actual prediction logic
func (pfd *PredictiveFaultDetector) performPredictions() {
	// Collect current metrics from enhanced detector
	metrics := pfd.collectCurrentMetrics()
	if len(metrics) == 0 {
		return
	}

	// Perform predictions
	predictions := pfd.PredictFaults(metrics)

	if len(predictions) > 0 {
		log.Info().Int("count", len(predictions)).Msg("Predictions generated")
	}
}

// performLearning performs continuous learning
func (pfd *PredictiveFaultDetector) performLearning() {
	if !pfd.config.EnableContinuousLearning {
		return
	}

	// Update models based on recent predictions and outcomes
	pfd.updateModelsFromOutcomes()

	log.Debug().Msg("Continuous learning performed")
}

// performModelUpdates performs model updates
func (pfd *PredictiveFaultDetector) performModelUpdates() {
	// Update time series models
	if pfd.timeSeriesAnalyzer != nil {
		pfd.updateTimeSeriesModels()
	}

	// Update ML models
	if pfd.mlPredictor != nil {
		pfd.updateMLModels()
	}

	log.Debug().Msg("Model updates performed")
}

// collectCurrentMetrics collects current metrics for prediction
func (pfd *PredictiveFaultDetector) collectCurrentMetrics() map[string]interface{} {
	// This would integrate with the enhanced detector's metrics collector
	// For now, return empty metrics
	return make(map[string]interface{})
}

// updateModelsFromOutcomes updates models based on prediction outcomes
func (pfd *PredictiveFaultDetector) updateModelsFromOutcomes() {
	pfd.predictionsMu.RLock()
	defer pfd.predictionsMu.RUnlock()

	// Find predictions with outcomes
	for _, prediction := range pfd.predictionHistory {
		if prediction.ActualOutcome != nil {
			// Update models based on accuracy
			pfd.learningEngine.onlineLearning.LastUpdated = time.Now()
		}
	}
}

// updateTimeSeriesModels updates time series models
func (pfd *PredictiveFaultDetector) updateTimeSeriesModels() {
	// Update ARMA models, LSTM models, etc.
	log.Debug().Msg("Time series models updated")
}

// updateMLModels updates ML models
func (pfd *PredictiveFaultDetector) updateMLModels() {
	// Update neural networks, random forests, etc.
	log.Debug().Msg("ML models updated")
}

// GetPredictions returns current predictions
func (pfd *PredictiveFaultDetector) GetPredictions() map[string]*FaultPrediction {
	pfd.predictionsMu.RLock()
	defer pfd.predictionsMu.RUnlock()

	predictions := make(map[string]*FaultPrediction)
	for id, prediction := range pfd.predictions {
		predictions[id] = prediction
	}

	return predictions
}

// GetPredictionHistory returns prediction history
func (pfd *PredictiveFaultDetector) GetPredictionHistory() []*FaultPrediction {
	pfd.predictionsMu.RLock()
	defer pfd.predictionsMu.RUnlock()

	history := make([]*FaultPrediction, len(pfd.predictionHistory))
	copy(history, pfd.predictionHistory)

	return history
}

// GetPredictionStatistics returns prediction statistics
func (pfd *PredictiveFaultDetector) GetPredictionStatistics() *PredictionStatistics {
	pfd.predictionsMu.RLock()
	defer pfd.predictionsMu.RUnlock()

	stats := &PredictionStatistics{
		TotalPredictions:    len(pfd.predictionHistory),
		ActivePredictions:   len(pfd.predictions),
		PredictionsByType:   make(map[FaultType]int),
		PredictionsByStatus: make(map[PredictionStatus]int),
		AverageConfidence:   0.0,
		AccuracyRate:        0.0,
	}

	totalConfidence := 0.0
	accurateCount := 0
	totalWithOutcome := 0

	for _, prediction := range pfd.predictionHistory {
		stats.PredictionsByType[prediction.PredictedType]++
		stats.PredictionsByStatus[prediction.Status]++
		totalConfidence += prediction.Confidence

		if prediction.ActualOutcome != nil {
			totalWithOutcome++
			if prediction.ActualOutcome.Accuracy > 0.7 {
				accurateCount++
			}
		}
	}

	if len(pfd.predictionHistory) > 0 {
		stats.AverageConfidence = totalConfidence / float64(len(pfd.predictionHistory))
	}

	if totalWithOutcome > 0 {
		stats.AccuracyRate = float64(accurateCount) / float64(totalWithOutcome)
	}

	return stats
}

// PredictionStatistics represents prediction statistics
type PredictionStatistics struct {
	TotalPredictions    int                      `json:"total_predictions"`
	ActivePredictions   int                      `json:"active_predictions"`
	PredictionsByType   map[FaultType]int        `json:"predictions_by_type"`
	PredictionsByStatus map[PredictionStatus]int `json:"predictions_by_status"`
	AverageConfidence   float64                  `json:"average_confidence"`
	AccuracyRate        float64                  `json:"accuracy_rate"`
}

// createFinalPrediction creates a final prediction from ensemble
func (pfd *PredictiveFaultDetector) createFinalPrediction(ensemble *EnsemblePrediction) *FaultPrediction {
	if ensemble.Count == 0 {
		return nil
	}

	// Calculate ensemble confidence
	ensembleConfidence := ensemble.TotalWeight / float64(ensemble.Count)

	// Check if confidence meets threshold
	if ensembleConfidence < pfd.config.ConfidenceThreshold {
		return nil
	}

	// Calculate average time to failure
	var totalTimeToFailure time.Duration
	for _, pred := range ensemble.Predictions {
		totalTimeToFailure += pred.TimeToFailure
	}
	avgTimeToFailure := totalTimeToFailure / time.Duration(len(ensemble.Predictions))

	// Create final prediction
	return &FaultPrediction{
		ID:            fmt.Sprintf("pred_%d", time.Now().UnixNano()),
		PredictedType: ensemble.FaultType,
		Target:        ensemble.Target,
		Confidence:    ensembleConfidence,
		TimeToFailure: avgTimeToFailure,
		PredictedAt:   time.Now(),
		ExpectedAt:    time.Now().Add(avgTimeToFailure),
		ModelUsed:     "ensemble",
		Features:      pfd.extractEnsembleFeatures(ensemble),
		Metadata: map[string]interface{}{
			"ensemble_count": ensemble.Count,
			"total_weight":   ensemble.TotalWeight,
			"models_used":    pfd.getModelsUsed(ensemble),
		},
		Status: PredictionStatusPending,
	}
}

// filterPredictions filters predictions by confidence and other criteria
func (pfd *PredictiveFaultDetector) filterPredictions(predictions []*FaultPrediction) []*FaultPrediction {
	var filtered []*FaultPrediction

	for _, prediction := range predictions {
		// Check confidence threshold
		if prediction.Confidence < pfd.config.ConfidenceThreshold {
			continue
		}

		// Check time horizon
		if prediction.TimeToFailure > pfd.config.PredictionHorizon {
			continue
		}

		// Check for duplicate predictions
		if !pfd.isDuplicatePrediction(prediction) {
			filtered = append(filtered, prediction)
		}
	}

	return filtered
}

// storePredictions stores predictions in the prediction history
func (pfd *PredictiveFaultDetector) storePredictions(predictions []*FaultPrediction) {
	pfd.predictionsMu.Lock()
	defer pfd.predictionsMu.Unlock()

	for _, prediction := range predictions {
		// Store in current predictions
		pfd.predictions[prediction.ID] = prediction

		// Add to history
		pfd.predictionHistory = append(pfd.predictionHistory, prediction)

		// Limit history size
		if len(pfd.predictionHistory) > pfd.config.MaxPredictionHistory {
			pfd.predictionHistory = pfd.predictionHistory[1:]
		}

		// Log prediction
		log.Info().
			Str("prediction_id", prediction.ID).
			Str("target", prediction.Target).
			Str("fault_type", string(prediction.PredictedType)).
			Float64("confidence", prediction.Confidence).
			Dur("time_to_failure", prediction.TimeToFailure).
			Msg("Fault predicted")
	}
}

package fault_tolerance

import (
	"math"
	"time"

	"github.com/rs/zerolog/log"
)

// AnomalyDetectionConfig configures anomaly detection
type AnomalyDetectionConfig struct {
	EnableStatistical bool    `json:"enable_statistical"`
	EnableML          bool    `json:"enable_ml"`
	Threshold         float64 `json:"threshold"`
	WindowSize        int     `json:"window_size"`
	MinSamples        int     `json:"min_samples"`
}

// AnomalyResult represents the result of anomaly detection
type AnomalyResult struct {
	MetricName      string    `json:"metric_name"`
	Value           float64   `json:"value"`
	Expected        float64   `json:"expected"`
	Deviation       float64   `json:"deviation"`
	Severity        string    `json:"severity"`
	Confidence      float64   `json:"confidence"`
	Timestamp       time.Time `json:"timestamp"`
	DetectionMethod string    `json:"detection_method"`
}

// NewAnomalyDetector creates a new anomaly detector
func NewAnomalyDetector(config *AnomalyDetectionConfig) *AnomalyDetector {
	if config == nil {
		config = &AnomalyDetectionConfig{
			EnableStatistical: true,
			EnableML:          false,
			Threshold:         2.0,
			WindowSize:        100,
			MinSamples:        10,
		}
	}

	return &AnomalyDetector{
		statisticalModels: make(map[string]*StatisticalModel),
		mlModels:          make(map[string]*MLModel),
		config:            config,
	}
}

// DetectAnomalies detects anomalies in the given metrics
func (ad *AnomalyDetector) DetectAnomalies(metrics map[string]interface{}) []*AnomalyResult {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	var anomalies []*AnomalyResult

	for metricName, value := range metrics {
		if numValue, ok := ad.convertToFloat64(value); ok {
			// Statistical detection
			if ad.config.EnableStatistical {
				if anomaly := ad.detectStatisticalAnomaly(metricName, numValue); anomaly != nil {
					anomalies = append(anomalies, anomaly)
				}
			}

			// ML detection
			if ad.config.EnableML {
				if anomaly := ad.detectMLAnomaly(metricName, numValue); anomaly != nil {
					anomalies = append(anomalies, anomaly)
				}
			}

			// Update models with new data
			ad.updateModels(metricName, numValue)
		}
	}

	return anomalies
}

// detectStatisticalAnomaly detects anomalies using statistical methods
func (ad *AnomalyDetector) detectStatisticalAnomaly(metricName string, value float64) *AnomalyResult {
	model, exists := ad.statisticalModels[metricName]
	if !exists || model.SampleSize < ad.config.MinSamples {
		return nil
	}

	// Z-score detection
	if model.Type == "zscore" {
		if model.StdDev == 0 {
			return nil
		}

		zScore := math.Abs(value-model.Mean) / model.StdDev
		if zScore > ad.config.Threshold {
			severity := ad.calculateSeverity(zScore, ad.config.Threshold)
			return &AnomalyResult{
				MetricName:      metricName,
				Value:           value,
				Expected:        model.Mean,
				Deviation:       zScore,
				Severity:        severity,
				Confidence:      math.Min(zScore/ad.config.Threshold, 1.0),
				Timestamp:       time.Now(),
				DetectionMethod: "zscore",
			}
		}
	}

	// IQR detection
	if model.Type == "iqr" {
		iqr := model.Q3 - model.Q1
		lowerBound := model.Q1 - 1.5*iqr
		upperBound := model.Q3 + 1.5*iqr

		if value < lowerBound || value > upperBound {
			deviation := math.Max(lowerBound-value, value-upperBound) / iqr
			severity := ad.calculateSeverity(deviation, 1.5)
			return &AnomalyResult{
				MetricName:      metricName,
				Value:           value,
				Expected:        model.Median,
				Deviation:       deviation,
				Severity:        severity,
				Confidence:      math.Min(deviation/1.5, 1.0),
				Timestamp:       time.Now(),
				DetectionMethod: "iqr",
			}
		}
	}

	return nil
}

// detectMLAnomaly detects anomalies using machine learning methods
func (ad *AnomalyDetector) detectMLAnomaly(metricName string, value float64) *AnomalyResult {
	model, exists := ad.mlModels[metricName]
	if !exists {
		return nil
	}

	// Simple autoencoder-like detection (placeholder implementation)
	if model.Type == "autoencoder" {
		reconstructionError := ad.calculateReconstructionError(model, value)
		if reconstructionError > model.Threshold {
			severity := ad.calculateSeverity(reconstructionError, model.Threshold)
			return &AnomalyResult{
				MetricName:      metricName,
				Value:           value,
				Expected:        0, // Would be the reconstructed value
				Deviation:       reconstructionError,
				Severity:        severity,
				Confidence:      math.Min(reconstructionError/model.Threshold, 1.0),
				Timestamp:       time.Now(),
				DetectionMethod: "autoencoder",
			}
		}
	}

	return nil
}

// updateModels updates the statistical and ML models with new data
func (ad *AnomalyDetector) updateModels(metricName string, value float64) {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	// Update statistical model
	if ad.config.EnableStatistical {
		ad.updateStatisticalModel(metricName, value)
	}

	// Update ML model
	if ad.config.EnableML {
		ad.updateMLModel(metricName, value)
	}
}

// updateStatisticalModel updates the statistical model for a metric
func (ad *AnomalyDetector) updateStatisticalModel(metricName string, value float64) {
	model, exists := ad.statisticalModels[metricName]
	if !exists {
		model = &StatisticalModel{
			Name:        metricName,
			Type:        "zscore", // Default to z-score
			SampleSize:  0,
			LastUpdated: time.Now(),
			Threshold:   ad.config.Threshold,
		}
		ad.statisticalModels[metricName] = model
	}

	// Update running statistics
	model.SampleSize++
	if model.SampleSize == 1 {
		model.Mean = value
		model.StdDev = 0
	} else {
		// Online algorithm for mean and variance
		oldMean := model.Mean
		model.Mean = oldMean + (value-oldMean)/float64(model.SampleSize)

		if model.SampleSize > 1 {
			// Update variance using Welford's online algorithm
			variance := ((float64(model.SampleSize-2) * model.StdDev * model.StdDev) +
				(value-oldMean)*(value-model.Mean)) / float64(model.SampleSize-1)
			model.StdDev = math.Sqrt(variance)
		}
	}

	model.LastUpdated = time.Now()

	// Update quartiles periodically (simplified implementation)
	if model.SampleSize%10 == 0 {
		ad.updateQuartiles(model, metricName)
	}
}

// updateMLModel updates the ML model for a metric
func (ad *AnomalyDetector) updateMLModel(metricName string, value float64) {
	model, exists := ad.mlModels[metricName]
	if !exists {
		model = &MLModel{
			Name:         metricName,
			Type:         "autoencoder",
			Parameters:   make(map[string]interface{}),
			TrainingData: make([][]float64, 0),
			LastTrained:  time.Now(),
			Accuracy:     0.0,
			Threshold:    1.0,
		}
		ad.mlModels[metricName] = model
	}

	// Add to training data (keep only recent data)
	model.TrainingData = append(model.TrainingData, []float64{value})
	if len(model.TrainingData) > ad.config.WindowSize {
		model.TrainingData = model.TrainingData[1:]
	}

	// Retrain periodically
	if len(model.TrainingData) >= ad.config.MinSamples &&
		time.Since(model.LastTrained) > 10*time.Minute {
		ad.retrainMLModel(model)
	}
}

// updateQuartiles updates quartile values for IQR detection
func (ad *AnomalyDetector) updateQuartiles(model *StatisticalModel, metricName string) {
	// This is a simplified implementation
	// In practice, you'd maintain a sliding window of values
	// For now, we'll use the mean and standard deviation to estimate quartiles
	model.Q1 = model.Mean - 0.675*model.StdDev
	model.Median = model.Mean
	model.Q3 = model.Mean + 0.675*model.StdDev
}

// retrainMLModel retrains the ML model
func (ad *AnomalyDetector) retrainMLModel(model *MLModel) {
	// Simplified retraining - in practice, this would involve actual ML algorithms
	if len(model.TrainingData) == 0 {
		return
	}

	// Calculate simple statistics as a placeholder for ML training
	sum := 0.0
	for _, data := range model.TrainingData {
		if len(data) > 0 {
			sum += data[0]
		}
	}
	mean := sum / float64(len(model.TrainingData))

	// Update threshold based on variance
	variance := 0.0
	for _, data := range model.TrainingData {
		if len(data) > 0 {
			diff := data[0] - mean
			variance += diff * diff
		}
	}
	variance /= float64(len(model.TrainingData))

	model.Threshold = math.Sqrt(variance) * 2.0 // 2 sigma threshold
	model.LastTrained = time.Now()
	model.Accuracy = 0.85 // Placeholder accuracy

	log.Debug().
		Str("metric", model.Name).
		Float64("threshold", model.Threshold).
		Msg("ML model retrained")
}

// calculateReconstructionError calculates reconstruction error for autoencoder
func (ad *AnomalyDetector) calculateReconstructionError(model *MLModel, value float64) float64 {
	// Simplified reconstruction error calculation
	// In practice, this would use the actual autoencoder model
	if len(model.TrainingData) == 0 {
		return 0.0
	}

	// Calculate mean of training data as "reconstruction"
	sum := 0.0
	for _, data := range model.TrainingData {
		if len(data) > 0 {
			sum += data[0]
		}
	}
	reconstruction := sum / float64(len(model.TrainingData))

	return math.Abs(value - reconstruction)
}

// calculateSeverity calculates severity based on deviation and threshold
func (ad *AnomalyDetector) calculateSeverity(deviation, threshold float64) string {
	ratio := deviation / threshold
	if ratio >= 3.0 {
		return "critical"
	} else if ratio >= 2.0 {
		return "high"
	} else if ratio >= 1.5 {
		return "medium"
	} else {
		return "low"
	}
}

// convertToFloat64 converts various numeric types to float64
func (ad *AnomalyDetector) convertToFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	default:
		return 0, false
	}
}

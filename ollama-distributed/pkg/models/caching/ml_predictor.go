package caching

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// LinearRegressionModel implements a simple linear regression ML model
type LinearRegressionModel struct {
	weights    []float64
	bias       float64
	learningRate float64
	iterations int
	accuracy   float64
	mutex      sync.RWMutex
}

// NewLinearRegressionModel creates a new linear regression model
func NewLinearRegressionModel(featureCount int) *LinearRegressionModel {
	return &LinearRegressionModel{
		weights:      make([]float64, featureCount),
		bias:         0.0,
		learningRate: 0.01,
		iterations:   1000,
		accuracy:     0.0,
	}
}

// Predict predicts usage probability based on features
func (lr *LinearRegressionModel) Predict(features []float64) (float64, error) {
	lr.mutex.RLock()
	defer lr.mutex.RUnlock()
	
	if len(features) != len(lr.weights) {
		return 0, fmt.Errorf("feature count mismatch: expected %d, got %d", len(lr.weights), len(features))
	}
	
	prediction := lr.bias
	for i, feature := range features {
		prediction += lr.weights[i] * feature
	}
	
	// Apply sigmoid to get probability between 0 and 1
	return sigmoid(prediction), nil
}

// Train trains the model using gradient descent
func (lr *LinearRegressionModel) Train(data []TrainingExample) error {
	lr.mutex.Lock()
	defer lr.mutex.Unlock()
	
	if len(data) == 0 {
		return fmt.Errorf("no training data provided")
	}
	
	// Initialize weights if not done
	if len(lr.weights) == 0 {
		lr.weights = make([]float64, len(data[0].Features))
	}
	
	// Gradient descent training
	for iter := 0; iter < lr.iterations; iter++ {
		totalLoss := 0.0
		
		// Calculate gradients
		weightGradients := make([]float64, len(lr.weights))
		biasGradient := 0.0
		
		for _, example := range data {
			// Forward pass
			prediction := lr.bias
			for i, feature := range example.Features {
				prediction += lr.weights[i] * feature
			}
			prediction = sigmoid(prediction)
			
			// Calculate loss (binary cross-entropy)
			loss := -example.Label*math.Log(prediction+1e-15) - (1-example.Label)*math.Log(1-prediction+1e-15)
			totalLoss += loss * example.Weight
			
			// Calculate gradients
			error := prediction - example.Label
			biasGradient += error * example.Weight
			
			for i, feature := range example.Features {
				weightGradients[i] += error * feature * example.Weight
			}
		}
		
		// Update weights
		lr.bias -= lr.learningRate * biasGradient / float64(len(data))
		for i := range lr.weights {
			lr.weights[i] -= lr.learningRate * weightGradients[i] / float64(len(data))
		}
		
		// Calculate accuracy every 100 iterations
		if iter%100 == 0 {
			lr.accuracy = lr.calculateAccuracy(data)
		}
	}
	
	return nil
}

// GetAccuracy returns the model's accuracy
func (lr *LinearRegressionModel) GetAccuracy() float64 {
	lr.mutex.RLock()
	defer lr.mutex.RUnlock()
	return lr.accuracy
}

// Save saves the model to a file
func (lr *LinearRegressionModel) Save(path string) error {
	// TODO: Implement model serialization
	return nil
}

// Load loads the model from a file
func (lr *LinearRegressionModel) Load(path string) error {
	// TODO: Implement model deserialization
	return nil
}

// calculateAccuracy calculates the model's accuracy on the given data
func (lr *LinearRegressionModel) calculateAccuracy(data []TrainingExample) float64 {
	correct := 0
	total := len(data)
	
	for _, example := range data {
		prediction := lr.bias
		for i, feature := range example.Features {
			prediction += lr.weights[i] * feature
		}
		prediction = sigmoid(prediction)
		
		// Convert to binary prediction
		binaryPrediction := 0.0
		if prediction > 0.5 {
			binaryPrediction = 1.0
		}
		
		if binaryPrediction == example.Label {
			correct++
		}
	}
	
	return float64(correct) / float64(total)
}

// AdvancedFeatureExtractor extracts features for ML prediction
type AdvancedFeatureExtractor struct {
	timeFeatures     []string
	contextFeatures  []string
	usageFeatures    []string
	seasonalFeatures []string
	featureCache     map[string][]float64
	cacheMutex       sync.RWMutex
}

// NewAdvancedFeatureExtractor creates a new feature extractor
func NewAdvancedFeatureExtractor() *AdvancedFeatureExtractor {
	return &AdvancedFeatureExtractor{
		timeFeatures:     []string{"hour", "day_of_week", "month", "is_weekend"},
		contextFeatures:  []string{"user_id_hash", "session_length", "request_type"},
		usageFeatures:    []string{"access_frequency", "last_access_hours", "total_accesses"},
		seasonalFeatures: []string{"seasonal_index", "trend", "weekly_pattern"},
		featureCache:     make(map[string][]float64),
	}
}

// ExtractFeatures extracts features from usage pattern
func (afe *AdvancedFeatureExtractor) ExtractFeatures(pattern *UsagePattern) []float64 {
	cacheKey := fmt.Sprintf("%s_%d", pattern.ModelID, pattern.Timestamp.Unix())
	
	// Check cache first
	afe.cacheMutex.RLock()
	if cached, exists := afe.featureCache[cacheKey]; exists {
		afe.cacheMutex.RUnlock()
		return cached
	}
	afe.cacheMutex.RUnlock()
	
	features := make([]float64, 0, 20)
	
	// Time features
	features = append(features, afe.extractTimeFeatures(pattern.Timestamp)...)
	
	// Context features
	features = append(features, afe.extractContextFeatures(pattern.UserContext)...)
	
	// Usage features
	features = append(features, afe.extractUsageFeatures(pattern)...)
	
	// Seasonal features
	features = append(features, afe.extractSeasonalFeatures(pattern.Timestamp)...)
	
	// Cache the result
	afe.cacheMutex.Lock()
	afe.featureCache[cacheKey] = features
	afe.cacheMutex.Unlock()
	
	return features
}

// extractTimeFeatures extracts time-based features
func (afe *AdvancedFeatureExtractor) extractTimeFeatures(timestamp time.Time) []float64 {
	features := make([]float64, 4)
	
	// Hour of day (normalized to 0-1)
	features[0] = float64(timestamp.Hour()) / 24.0
	
	// Day of week (normalized to 0-1)
	features[1] = float64(timestamp.Weekday()) / 7.0
	
	// Month (normalized to 0-1)
	features[2] = float64(timestamp.Month()) / 12.0
	
	// Is weekend (binary)
	if timestamp.Weekday() == time.Saturday || timestamp.Weekday() == time.Sunday {
		features[3] = 1.0
	} else {
		features[3] = 0.0
	}
	
	return features
}

// extractContextFeatures extracts context-based features
func (afe *AdvancedFeatureExtractor) extractContextFeatures(context map[string]interface{}) []float64 {
	features := make([]float64, 3)
	
	// User ID hash (normalized)
	if userID, ok := context["user_id"].(string); ok {
		features[0] = float64(simpleHash(userID)) / float64(math.MaxUint32)
	}
	
	// Session length (normalized to hours)
	if sessionLength, ok := context["session_length"].(time.Duration); ok {
		features[1] = sessionLength.Hours() / 24.0 // Normalize to days
	}
	
	// Request type (encoded)
	if requestType, ok := context["request_type"].(string); ok {
		features[2] = encodeRequestType(requestType)
	}
	
	return features
}

// extractUsageFeatures extracts usage-based features
func (afe *AdvancedFeatureExtractor) extractUsageFeatures(pattern *UsagePattern) []float64 {
	features := make([]float64, 3)
	
	// Access frequency (normalized)
	features[0] = math.Min(float64(pattern.AccessCount)/100.0, 1.0)
	
	// Hours since last access (normalized to days)
	hoursSinceAccess := time.Since(pattern.Timestamp).Hours()
	features[1] = math.Min(hoursSinceAccess/24.0, 7.0) / 7.0 // Max 7 days
	
	// Total accesses (log normalized)
	if pattern.AccessCount > 0 {
		features[2] = math.Log(float64(pattern.AccessCount)) / 10.0 // Normalize log scale
	}
	
	return features
}

// extractSeasonalFeatures extracts seasonal features
func (afe *AdvancedFeatureExtractor) extractSeasonalFeatures(timestamp time.Time) []float64 {
	features := make([]float64, 3)
	
	// Seasonal index (based on day of year)
	dayOfYear := timestamp.YearDay()
	features[0] = math.Sin(2 * math.Pi * float64(dayOfYear) / 365.0)
	
	// Trend (based on hour of day)
	hourOfDay := timestamp.Hour()
	features[1] = math.Sin(2 * math.Pi * float64(hourOfDay) / 24.0)
	
	// Weekly pattern (based on day of week)
	dayOfWeek := int(timestamp.Weekday())
	features[2] = math.Sin(2 * math.Pi * float64(dayOfWeek) / 7.0)
	
	return features
}

// PredictiveEvictionPolicy implements ML-based eviction
type PredictiveEvictionPolicy struct {
	predictor    *UsagePredictor
	threshold    float64
	minAge       time.Duration
	maxAge       time.Duration
}

// NewPredictiveEvictionPolicy creates a new predictive eviction policy
func NewPredictiveEvictionPolicy(predictor *UsagePredictor) *PredictiveEvictionPolicy {
	return &PredictiveEvictionPolicy{
		predictor: predictor,
		threshold: 0.3, // Evict if prediction score < 0.3
		minAge:    time.Hour,
		maxAge:    24 * time.Hour,
	}
}

// ShouldEvict determines if an entry should be evicted
func (pep *PredictiveEvictionPolicy) ShouldEvict(entry *CacheEntry, cacheSize int64, maxSize int64) bool {
	// Always evict if cache is over capacity and entry is old enough
	if cacheSize > maxSize && time.Since(entry.LastAccessed) > pep.minAge {
		return true
	}
	
	// Evict if predicted score is low and entry is old
	if entry.PredictedScore < pep.threshold && time.Since(entry.LastAccessed) > pep.minAge {
		return true
	}
	
	// Force evict very old entries
	if time.Since(entry.LastAccessed) > pep.maxAge {
		return true
	}
	
	return false
}

// SelectVictim selects the best candidate for eviction
func (pep *PredictiveEvictionPolicy) SelectVictim(entries []*CacheEntry) *CacheEntry {
	if len(entries) == 0 {
		return nil
	}
	
	// Sort entries by eviction score (lower is better for eviction)
	sort.Slice(entries, func(i, j int) bool {
		scoreI := pep.calculateEvictionScore(entries[i])
		scoreJ := pep.calculateEvictionScore(entries[j])
		return scoreI < scoreJ
	})
	
	return entries[0]
}

// GetName returns the policy name
func (pep *PredictiveEvictionPolicy) GetName() string {
	return "predictive_eviction"
}

// calculateEvictionScore calculates a score for eviction (lower = more likely to evict)
func (pep *PredictiveEvictionPolicy) calculateEvictionScore(entry *CacheEntry) float64 {
	// Base score from prediction
	score := entry.PredictedScore
	
	// Adjust for recency (more recent = higher score)
	hoursSinceAccess := time.Since(entry.LastAccessed).Hours()
	recencyBonus := math.Exp(-hoursSinceAccess / 24.0) // Exponential decay over days
	score += recencyBonus * 0.3
	
	// Adjust for access frequency
	frequencyBonus := math.Log(float64(entry.AccessCount + 1)) / 10.0
	score += frequencyBonus * 0.2
	
	return score
}

// Helper functions
func sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

func simpleHash(s string) uint32 {
	hash := uint32(0)
	for _, c := range s {
		hash = hash*31 + uint32(c)
	}
	return hash
}

func encodeRequestType(requestType string) float64 {
	switch requestType {
	case "inference":
		return 1.0
	case "training":
		return 0.8
	case "evaluation":
		return 0.6
	case "download":
		return 0.4
	default:
		return 0.2
	}
}

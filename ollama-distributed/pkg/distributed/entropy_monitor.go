package distributed

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// EntropyMonitor implements gradient/activation entropy monitoring
// Based on EDGC research (arXiv:2511.10333v1): Entropy-Driven Dynamic Gradient Compression
// Monitors entropy evolution to dynamically adjust compression rank
type EntropyMonitor struct {
	mu sync.RWMutex

	// Entropy tracking
	gradientEntropy   map[string]*EntropyHistory
	activationEntropy map[string]*EntropyHistory

	// Configuration
	config *EntropyMonitorConfig

	// Metrics
	metrics *EntropyMetrics
}

// EntropyHistory tracks entropy evolution over time
type EntropyHistory struct {
	LayerName      string
	EntropyValues  []float64
	Timestamps     []time.Time
	CurrentEntropy float64
	AvgEntropy     float64
	StdDevEntropy  float64
	TrendSlope     float64 // Positive = increasing, negative = decreasing
}

// EntropyMonitorConfig configures entropy monitoring behavior
type EntropyMonitorConfig struct {
	HistorySize        int           // Number of entropy values to keep
	UpdateInterval     time.Duration // How often to calculate entropy
	EntropyThreshold   float64       // Threshold for high entropy (triggers compression)
	TrendWindowSize    int           // Window size for trend calculation
	EnableAdaptive     bool          // Enable adaptive threshold adjustment
	MinEntropyForComp  float64       // Minimum entropy to enable compression
}

// EntropyMetrics tracks entropy monitoring performance
type EntropyMetrics struct {
	mu                    sync.RWMutex
	TotalCalculations     int64
	HighEntropyDetections int64
	LowEntropyDetections  int64
	AverageEntropy        float64
	CompressionTriggered  int64
}

// DefaultEntropyMonitorConfig returns default configuration
func DefaultEntropyMonitorConfig() *EntropyMonitorConfig {
	return &EntropyMonitorConfig{
		HistorySize:        100,
		UpdateInterval:     100 * time.Millisecond,
		EntropyThreshold:   0.7,
		TrendWindowSize:    10,
		EnableAdaptive:     true,
		MinEntropyForComp:  0.3,
	}
}

// NewEntropyMonitor creates a new entropy monitor
func NewEntropyMonitor(config *EntropyMonitorConfig) *EntropyMonitor {
	if config == nil {
		config = DefaultEntropyMonitorConfig()
	}

	return &EntropyMonitor{
		gradientEntropy:   make(map[string]*EntropyHistory),
		activationEntropy: make(map[string]*EntropyHistory),
		config:            config,
		metrics:           &EntropyMetrics{},
	}
}

// CalculateGradientEntropy calculates entropy for gradient values
// Entropy H(X) = -Σ p(x) * log2(p(x))
func (em *EntropyMonitor) CalculateGradientEntropy(layerName string, gradients []float64) (float64, error) {
	if len(gradients) == 0 {
		return 0, fmt.Errorf("empty gradient array")
	}

	entropy := em.calculateEntropy(gradients)

	// Update history
	em.mu.Lock()
	defer em.mu.Unlock()

	history, exists := em.gradientEntropy[layerName]
	if !exists {
		history = &EntropyHistory{
			LayerName:     layerName,
			EntropyValues: make([]float64, 0, em.config.HistorySize),
			Timestamps:    make([]time.Time, 0, em.config.HistorySize),
		}
		em.gradientEntropy[layerName] = history
	}

	em.updateHistory(history, entropy)
	em.updateMetrics(entropy)

	return entropy, nil
}

// CalculateActivationEntropy calculates entropy for activation values
func (em *EntropyMonitor) CalculateActivationEntropy(layerName string, activations []float64) (float64, error) {
	if len(activations) == 0 {
		return 0, fmt.Errorf("empty activation array")
	}

	entropy := em.calculateEntropy(activations)

	// Update history
	em.mu.Lock()
	defer em.mu.Unlock()

	history, exists := em.activationEntropy[layerName]
	if !exists {
		history = &EntropyHistory{
			LayerName:     layerName,
			EntropyValues: make([]float64, 0, em.config.HistorySize),
			Timestamps:    make([]time.Time, 0, em.config.HistorySize),
		}
		em.activationEntropy[layerName] = history
	}

	em.updateHistory(history, entropy)
	em.updateMetrics(entropy)

	return entropy, nil
}

// calculateEntropy computes Shannon entropy for a set of values
func (em *EntropyMonitor) calculateEntropy(values []float64) float64 {
	// Normalize values to probabilities
	sum := 0.0
	for _, v := range values {
		sum += math.Abs(v)
	}

	if sum == 0 {
		return 0
	}

	// Calculate entropy
	entropy := 0.0
	for _, v := range values {
		p := math.Abs(v) / sum
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}

	// Normalize to [0, 1]
	maxEntropy := math.Log2(float64(len(values)))
	if maxEntropy > 0 {
		entropy /= maxEntropy
	}

	return entropy
}

// updateHistory updates entropy history for a layer
func (em *EntropyMonitor) updateHistory(history *EntropyHistory, entropy float64) {
	history.EntropyValues = append(history.EntropyValues, entropy)
	history.Timestamps = append(history.Timestamps, time.Now())

	// Maintain history size
	if len(history.EntropyValues) > em.config.HistorySize {
		history.EntropyValues = history.EntropyValues[1:]
		history.Timestamps = history.Timestamps[1:]
	}

	history.CurrentEntropy = entropy

	// Calculate statistics
	history.AvgEntropy = em.calculateAverage(history.EntropyValues)
	history.StdDevEntropy = em.calculateStdDev(history.EntropyValues, history.AvgEntropy)
	history.TrendSlope = em.calculateTrend(history.EntropyValues)
}

// calculateAverage calculates the average of values
func (em *EntropyMonitor) calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculateStdDev calculates standard deviation
func (em *EntropyMonitor) calculateStdDev(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 0
	}

	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(values))

	return math.Sqrt(variance)
}

// calculateTrend calculates the trend slope using linear regression
func (em *EntropyMonitor) calculateTrend(values []float64) float64 {
	n := len(values)
	if n < 2 {
		return 0
	}

	// Use recent window for trend
	windowSize := em.config.TrendWindowSize
	if windowSize > n {
		windowSize = n
	}

	recentValues := values[n-windowSize:]

	// Simple linear regression
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0
	for i, y := range recentValues {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	nFloat := float64(len(recentValues))
	slope := (nFloat*sumXY - sumX*sumY) / (nFloat*sumX2 - sumX*sumX)

	return slope
}

// ShouldCompress determines if compression should be applied based on entropy
func (em *EntropyMonitor) ShouldCompress(layerName string, isGradient bool) bool {
	em.mu.RLock()
	defer em.mu.RUnlock()

	var history *EntropyHistory
	if isGradient {
		history = em.gradientEntropy[layerName]
	} else {
		history = em.activationEntropy[layerName]
	}

	if history == nil {
		return false
	}

	// Check if entropy is above threshold
	if history.CurrentEntropy < em.config.MinEntropyForComp {
		return false
	}

	return history.CurrentEntropy >= em.config.EntropyThreshold
}

// updateMetrics updates performance metrics
func (em *EntropyMonitor) updateMetrics(entropy float64) {
	em.metrics.mu.Lock()
	defer em.metrics.mu.Unlock()

	em.metrics.TotalCalculations++

	if entropy >= em.config.EntropyThreshold {
		em.metrics.HighEntropyDetections++
	} else {
		em.metrics.LowEntropyDetections++
	}

	// Update average entropy (exponential moving average)
	alpha := 0.1
	em.metrics.AverageEntropy = em.metrics.AverageEntropy*(1-alpha) + entropy*alpha
}

// GetMetrics returns current metrics
func (em *EntropyMonitor) GetMetrics() *EntropyMetrics {
	em.metrics.mu.RLock()
	defer em.metrics.mu.RUnlock()

	return &EntropyMetrics{
		TotalCalculations:     em.metrics.TotalCalculations,
		HighEntropyDetections: em.metrics.HighEntropyDetections,
		LowEntropyDetections:  em.metrics.LowEntropyDetections,
		AverageEntropy:        em.metrics.AverageEntropy,
		CompressionTriggered:  em.metrics.CompressionTriggered,
	}
}


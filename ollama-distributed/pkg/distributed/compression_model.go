package distributed

import (
	"math"
	"sync"
)

// CompressionModel implements dynamic rank adjustment and compression ratio prediction
// Based on EDGC research: adapts compression based on entropy evolution
type CompressionModel struct {
	mu sync.RWMutex

	// Entropy monitor integration
	entropyMonitor *EntropyMonitor

	// Compression state
	layerCompressionRank  map[string]int
	layerCompressionRatio map[string]float64

	// Configuration
	config *CompressionModelConfig

	// Metrics
	metrics *CompressionMetrics
}

// CompressionModelConfig configures compression model behavior
type CompressionModelConfig struct {
	MinRank            int     // Minimum compression rank
	MaxRank            int     // Maximum compression rank
	DefaultRank        int     // Default compression rank
	RankAdjustmentStep int     // How much to adjust rank per update
	TargetRatio        float64 // Target compression ratio (0.0-1.0)
	AdaptiveThreshold  float64 // Threshold for adaptive adjustment
}

// CompressionMetrics tracks compression performance
type CompressionMetrics struct {
	mu                      sync.RWMutex
	TotalCompressions       int64
	AverageCompressionRatio float64
	TotalBytesSaved         uint64
	RankAdjustments         int64
}

// DefaultCompressionModelConfig returns default configuration
func DefaultCompressionModelConfig() *CompressionModelConfig {
	return &CompressionModelConfig{
		MinRank:            2,
		MaxRank:            64,
		DefaultRank:        16,
		RankAdjustmentStep: 2,
		TargetRatio:        0.5, // 50% compression
		AdaptiveThreshold:  0.1,
	}
}

// NewCompressionModel creates a new compression model
func NewCompressionModel(config *CompressionModelConfig, entropyMonitor *EntropyMonitor) *CompressionModel {
	if config == nil {
		config = DefaultCompressionModelConfig()
	}

	return &CompressionModel{
		entropyMonitor:        entropyMonitor,
		layerCompressionRank:  make(map[string]int),
		layerCompressionRatio: make(map[string]float64),
		config:                config,
		metrics:               &CompressionMetrics{},
	}
}

// GetCompressionRank returns the current compression rank for a layer
func (cm *CompressionModel) GetCompressionRank(layerName string) int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	rank, exists := cm.layerCompressionRank[layerName]
	if !exists {
		return cm.config.DefaultRank
	}

	return rank
}

// UpdateCompressionRank updates compression rank based on entropy
func (cm *CompressionModel) UpdateCompressionRank(layerName string, entropy float64) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	currentRank, exists := cm.layerCompressionRank[layerName]
	if !exists {
		currentRank = cm.config.DefaultRank
	}

	// Adjust rank based on entropy
	// High entropy → increase rank (less compression)
	// Low entropy → decrease rank (more compression)
	newRank := currentRank

	if entropy > 0.7 {
		// High entropy: increase rank
		newRank += cm.config.RankAdjustmentStep
	} else if entropy < 0.3 {
		// Low entropy: decrease rank
		newRank -= cm.config.RankAdjustmentStep
	}

	// Clamp to valid range
	if newRank < cm.config.MinRank {
		newRank = cm.config.MinRank
	}
	if newRank > cm.config.MaxRank {
		newRank = cm.config.MaxRank
	}

	// Update rank
	cm.layerCompressionRank[layerName] = newRank

	// Update metrics
	if newRank != currentRank {
		cm.metrics.mu.Lock()
		cm.metrics.RankAdjustments++
		cm.metrics.mu.Unlock()
	}

	return nil
}

// PredictCompressionRatio predicts compression ratio for given parameters
func (cm *CompressionModel) PredictCompressionRatio(originalSize uint64, rank int) float64 {
	// Compression ratio model: ratio = 1 - (rank / maxRank)
	// Higher rank = less compression
	ratio := 1.0 - (float64(rank) / float64(cm.config.MaxRank))

	// Ensure ratio is in valid range
	if ratio < 0.1 {
		ratio = 0.1
	}
	if ratio > 0.9 {
		ratio = 0.9
	}

	return ratio
}

// CompressData compresses data using the current compression rank
func (cm *CompressionModel) CompressData(layerName string, data []float64) ([]byte, error) {
	rank := cm.GetCompressionRank(layerName)

	// Perform compression (placeholder - actual implementation would use SVD or similar)
	compressed := cm.performCompression(data, rank)

	// Update metrics
	originalSize := uint64(len(data) * 8) // 8 bytes per float64
	compressedSize := uint64(len(compressed))
	ratio := float64(compressedSize) / float64(originalSize)

	cm.updateMetrics(originalSize, compressedSize, ratio)

	return compressed, nil
}

// performCompression performs actual compression (placeholder)
func (cm *CompressionModel) performCompression(data []float64, rank int) []byte {
	// Placeholder: Simple quantization
	// Actual implementation would use low-rank approximation (SVD)

	// Quantize to rank levels
	quantized := make([]byte, len(data))

	// Find min/max for normalization
	min, max := data[0], data[0]
	for _, v := range data {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// Quantize
	scale := float64(rank-1) / (max - min)
	for i, v := range data {
		normalized := (v - min) * scale
		quantized[i] = byte(math.Round(normalized))
	}

	return quantized
}

// DecompressData decompresses data
func (cm *CompressionModel) DecompressData(layerName string, compressed []byte) ([]float64, error) {
	rank := cm.GetCompressionRank(layerName)

	// Perform decompression (placeholder)
	decompressed := cm.performDecompression(compressed, rank)

	return decompressed, nil
}

// performDecompression performs actual decompression (placeholder)
func (cm *CompressionModel) performDecompression(compressed []byte, rank int) []float64 {
	// Placeholder: Reverse quantization
	decompressed := make([]float64, len(compressed))

	// Dequantize (would need to store min/max from compression)
	scale := 1.0 / float64(rank-1)
	for i, v := range compressed {
		decompressed[i] = float64(v) * scale
	}

	return decompressed
}

// OptimizeCompressionRank optimizes compression rank for a layer
func (cm *CompressionModel) OptimizeCompressionRank(layerName string, targetRatio float64) (int, error) {
	// Binary search for optimal rank
	minRank := cm.config.MinRank
	maxRank := cm.config.MaxRank

	for maxRank-minRank > 1 {
		midRank := (minRank + maxRank) / 2
		predictedRatio := cm.PredictCompressionRatio(1024, midRank)

		if predictedRatio < targetRatio {
			maxRank = midRank
		} else {
			minRank = midRank
		}
	}

	optimalRank := minRank

	// Update layer rank
	cm.mu.Lock()
	cm.layerCompressionRank[layerName] = optimalRank
	cm.mu.Unlock()

	return optimalRank, nil
}

// updateMetrics updates compression metrics
func (cm *CompressionModel) updateMetrics(originalSize, compressedSize uint64, ratio float64) {
	cm.metrics.mu.Lock()
	defer cm.metrics.mu.Unlock()

	cm.metrics.TotalCompressions++
	cm.metrics.TotalBytesSaved += originalSize - compressedSize

	// Update average compression ratio (exponential moving average)
	alpha := 0.1
	cm.metrics.AverageCompressionRatio = cm.metrics.AverageCompressionRatio*(1-alpha) + ratio*alpha
}

// GetMetrics returns current metrics
func (cm *CompressionModel) GetMetrics() *CompressionMetrics {
	cm.metrics.mu.RLock()
	defer cm.metrics.mu.RUnlock()

	return &CompressionMetrics{
		TotalCompressions:       cm.metrics.TotalCompressions,
		AverageCompressionRatio: cm.metrics.AverageCompressionRatio,
		TotalBytesSaved:         cm.metrics.TotalBytesSaved,
		RankAdjustments:         cm.metrics.RankAdjustments,
	}
}

// GetCompressionRatio returns the current compression ratio for a layer
func (cm *CompressionModel) GetCompressionRatio(layerName string) float64 {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	ratio, exists := cm.layerCompressionRatio[layerName]
	if !exists {
		return cm.config.TargetRatio
	}

	return ratio
}

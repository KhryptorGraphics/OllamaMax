package distributed

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewCompressionModel tests compression model creation
func TestNewCompressionModel(t *testing.T) {
	config := DefaultCompressionModelConfig()
	entropyMonitor := NewEntropyMonitor(DefaultEntropyMonitorConfig())
	model := NewCompressionModel(config, entropyMonitor)

	assert.NotNil(t, model)
	assert.NotNil(t, model.entropyMonitor)
}

// TestPredictCompressionRatio tests compression ratio prediction
func TestPredictCompressionRatio(t *testing.T) {
	config := DefaultCompressionModelConfig()
	entropyMonitor := NewEntropyMonitor(DefaultEntropyMonitorConfig())
	model := NewCompressionModel(config, entropyMonitor)

	ratio := model.PredictCompressionRatio(1000, 16)

	assert.Greater(t, ratio, 0.0)
	assert.Less(t, ratio, 1.0)
}

// TestUpdateCompressionRank tests rank adjustment
func TestUpdateCompressionRank(t *testing.T) {
	config := DefaultCompressionModelConfig()
	entropyMonitor := NewEntropyMonitor(DefaultEntropyMonitorConfig())
	model := NewCompressionModel(config, entropyMonitor)

	// Update with high entropy (should increase rank)
	err := model.UpdateCompressionRank("layer_0", 0.8)
	require.NoError(t, err)

	rank := model.GetCompressionRank("layer_0")
	assert.Greater(t, rank, config.DefaultRank)
}

// TestGetCompressionRank tests getting compression rank
func TestGetCompressionRank(t *testing.T) {
	config := DefaultCompressionModelConfig()
	entropyMonitor := NewEntropyMonitor(DefaultEntropyMonitorConfig())
	model := NewCompressionModel(config, entropyMonitor)

	rank := model.GetCompressionRank("layer_0")
	assert.Equal(t, config.DefaultRank, rank)
}

// TestCompressData tests data compression
func TestCompressData(t *testing.T) {
	config := DefaultCompressionModelConfig()
	entropyMonitor := NewEntropyMonitor(DefaultEntropyMonitorConfig())
	model := NewCompressionModel(config, entropyMonitor)

	data := []float64{0.1, 0.2, 0.3, 0.4, 0.5}
	compressed, err := model.CompressData("layer_0", data)
	require.NoError(t, err)
	assert.NotNil(t, compressed)
	assert.Greater(t, len(compressed), 0)
}

// TestAdaptiveRankAdjustment tests adaptive rank adjustment
func TestAdaptiveRankAdjustment(t *testing.T) {
	config := DefaultCompressionModelConfig()
	entropyMonitor := NewEntropyMonitor(DefaultEntropyMonitorConfig())
	model := NewCompressionModel(config, entropyMonitor)

	// Test increasing rank with high entropy
	err := model.UpdateCompressionRank("layer_0", 0.9)
	require.NoError(t, err)
	rank1 := model.GetCompressionRank("layer_0")
	assert.Greater(t, rank1, config.DefaultRank)

	// Test decreasing rank with low entropy
	err = model.UpdateCompressionRank("layer_0", 0.1)
	require.NoError(t, err)
	rank2 := model.GetCompressionRank("layer_0")
	assert.Less(t, rank2, rank1)
}

// TestCompressionRatioEstimation tests compression ratio estimation
func TestCompressionRatioEstimation(t *testing.T) {
	config := DefaultCompressionModelConfig()
	entropyMonitor := NewEntropyMonitor(DefaultEntropyMonitorConfig())
	model := NewCompressionModel(config, entropyMonitor)

	// Estimate ratio for different ranks
	// Formula: ratio = 1 - (rank / maxRank)
	// Lower rank = higher ratio (less compression)
	// Higher rank = lower ratio (more compression)
	ratio1 := model.PredictCompressionRatio(1000, 32) // Higher rank = lower ratio
	ratio2 := model.PredictCompressionRatio(1000, 8)  // Lower rank = higher ratio

	assert.Greater(t, ratio1, 0.0)
	assert.Less(t, ratio1, 1.0)
	assert.Greater(t, ratio2, 0.0)
	assert.Less(t, ratio2, 1.0)
	// Lower rank = higher ratio
	assert.Greater(t, ratio2, ratio1)
}

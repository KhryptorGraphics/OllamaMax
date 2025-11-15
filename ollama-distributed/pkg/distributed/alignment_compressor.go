package distributed

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AlignmentCompressor implements dynamic alignment compression
// Integrates entropy monitoring and compression model for adaptive compression
// Based on EDGC research: 46.45% reduction in communication latency
type AlignmentCompressor struct {
	mu sync.RWMutex

	// Component integration
	entropyMonitor   *EntropyMonitor
	compressionModel *CompressionModel
	blockSync        *BlockSynchronizer

	// Compression state
	layerCompressionState map[string]*CompressionState

	// Configuration
	config *AlignmentCompressorConfig

	// Metrics
	metrics *AlignmentMetrics

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// CompressionState tracks compression state for a layer
type CompressionState struct {
	LayerName          string
	IsCompressed       bool
	CompressionRank    int
	CompressionRatio   float64
	OriginalSize       uint64
	CompressedSize     uint64
	LastCompression    time.Time
	CompressionCount   int64
	DecompressionCount int64
}

// AlignmentCompressorConfig configures alignment compressor behavior
type AlignmentCompressorConfig struct {
	EnableCompression     bool
	CompressionInterval   time.Duration
	MinCompressionSize    uint64  // Minimum size to compress
	MaxCompressionLatency time.Duration
	TargetLatencyReduction float64 // Target: 46.45% reduction
}

// AlignmentMetrics tracks alignment compressor performance
type AlignmentMetrics struct {
	mu                     sync.RWMutex
	TotalCompressions      int64
	TotalDecompressions    int64
	TotalBytesCompressed   uint64
	TotalBytesDecompressed uint64
	AverageCompressionTime time.Duration
	LatencyReduction       float64 // Actual latency reduction achieved
	CommunicationSavings   uint64  // Total bytes saved
}

// DefaultAlignmentCompressorConfig returns default configuration
func DefaultAlignmentCompressorConfig() *AlignmentCompressorConfig {
	return &AlignmentCompressorConfig{
		EnableCompression:      true,
		CompressionInterval:    100 * time.Millisecond,
		MinCompressionSize:     1024, // 1KB
		MaxCompressionLatency:  50 * time.Millisecond,
		TargetLatencyReduction: 0.4645, // 46.45%
	}
}

// NewAlignmentCompressor creates a new alignment compressor
func NewAlignmentCompressor(
	config *AlignmentCompressorConfig,
	entropyMonitor *EntropyMonitor,
	compressionModel *CompressionModel,
	blockSync *BlockSynchronizer,
) *AlignmentCompressor {
	if config == nil {
		config = DefaultAlignmentCompressorConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	ac := &AlignmentCompressor{
		entropyMonitor:        entropyMonitor,
		compressionModel:      compressionModel,
		blockSync:             blockSync,
		layerCompressionState: make(map[string]*CompressionState),
		config:                config,
		metrics:               &AlignmentMetrics{},
		ctx:                   ctx,
		cancel:                cancel,
	}

	return ac
}

// CompressActivations compresses activations for a layer
func (ac *AlignmentCompressor) CompressActivations(layerName string, activations []float64) ([]byte, error) {
	if !ac.config.EnableCompression {
		// Return uncompressed (convert to bytes)
		return ac.floatsToBytes(activations), nil
	}

	startTime := time.Now()

	// Check if compression should be applied
	shouldCompress := ac.entropyMonitor.ShouldCompress(layerName, false)
	if !shouldCompress {
		return ac.floatsToBytes(activations), nil
	}

	// Calculate entropy
	entropy, err := ac.entropyMonitor.CalculateActivationEntropy(layerName, activations)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate entropy: %w", err)
	}

	// Update compression rank based on entropy
	err = ac.compressionModel.UpdateCompressionRank(layerName, entropy)
	if err != nil {
		return nil, fmt.Errorf("failed to update compression rank: %w", err)
	}

	// Compress data
	compressed, err := ac.compressionModel.CompressData(layerName, activations)
	if err != nil {
		return nil, fmt.Errorf("failed to compress data: %w", err)
	}

	// Update state
	ac.updateCompressionState(layerName, uint64(len(activations)*8), uint64(len(compressed)))

	// Update metrics
	compressionTime := time.Since(startTime)
	ac.updateMetrics(uint64(len(activations)*8), uint64(len(compressed)), compressionTime)

	return compressed, nil
}

// DecompressActivations decompresses activations for a layer
func (ac *AlignmentCompressor) DecompressActivations(layerName string, compressed []byte) ([]float64, error) {
	if !ac.config.EnableCompression {
		// Return as-is (convert from bytes)
		return ac.bytesToFloats(compressed), nil
	}

	// Decompress data
	decompressed, err := ac.compressionModel.DecompressData(layerName, compressed)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	// Update metrics
	ac.metrics.mu.Lock()
	ac.metrics.TotalDecompressions++
	ac.metrics.TotalBytesDecompressed += uint64(len(decompressed) * 8)
	ac.metrics.mu.Unlock()

	return decompressed, nil
}

// CompressGradients compresses gradients for a layer
func (ac *AlignmentCompressor) CompressGradients(layerName string, gradients []float64) ([]byte, error) {
	if !ac.config.EnableCompression {
		return ac.floatsToBytes(gradients), nil
	}

	// Check if compression should be applied
	shouldCompress := ac.entropyMonitor.ShouldCompress(layerName, true)
	if !shouldCompress {
		return ac.floatsToBytes(gradients), nil
	}

	// Calculate entropy
	entropy, err := ac.entropyMonitor.CalculateGradientEntropy(layerName, gradients)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate entropy: %w", err)
	}

	// Update compression rank
	err = ac.compressionModel.UpdateCompressionRank(layerName, entropy)
	if err != nil {
		return nil, fmt.Errorf("failed to update compression rank: %w", err)
	}

	// Compress data
	compressed, err := ac.compressionModel.CompressData(layerName, gradients)
	if err != nil {
		return nil, fmt.Errorf("failed to compress data: %w", err)
	}

	return compressed, nil
}

// updateCompressionState updates compression state for a layer
func (ac *AlignmentCompressor) updateCompressionState(layerName string, originalSize, compressedSize uint64) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	state, exists := ac.layerCompressionState[layerName]
	if !exists {
		state = &CompressionState{
			LayerName: layerName,
		}
		ac.layerCompressionState[layerName] = state
	}

	state.IsCompressed = true
	state.CompressionRank = ac.compressionModel.GetCompressionRank(layerName)
	state.CompressionRatio = float64(compressedSize) / float64(originalSize)
	state.OriginalSize = originalSize
	state.CompressedSize = compressedSize
	state.LastCompression = time.Now()
	state.CompressionCount++
}

// updateMetrics updates performance metrics
func (ac *AlignmentCompressor) updateMetrics(originalSize, compressedSize uint64, compressionTime time.Duration) {
	ac.metrics.mu.Lock()
	defer ac.metrics.mu.Unlock()

	ac.metrics.TotalCompressions++
	ac.metrics.TotalBytesCompressed += originalSize
	ac.metrics.CommunicationSavings += originalSize - compressedSize

	// Update average compression time (exponential moving average)
	alpha := 0.1
	ac.metrics.AverageCompressionTime = time.Duration(
		float64(ac.metrics.AverageCompressionTime)*(1-alpha) + float64(compressionTime)*alpha,
	)

	// Calculate latency reduction
	if originalSize > 0 {
		reduction := float64(originalSize-compressedSize) / float64(originalSize)
		ac.metrics.LatencyReduction = ac.metrics.LatencyReduction*(1-alpha) + reduction*alpha
	}
}

// floatsToBytes converts float64 slice to bytes (placeholder)
func (ac *AlignmentCompressor) floatsToBytes(floats []float64) []byte {
	bytes := make([]byte, len(floats)*8)
	// Placeholder: actual implementation would use binary encoding
	return bytes
}

// bytesToFloats converts bytes to float64 slice (placeholder)
func (ac *AlignmentCompressor) bytesToFloats(bytes []byte) []float64 {
	floats := make([]float64, len(bytes)/8)
	// Placeholder: actual implementation would use binary decoding
	return floats
}

// GetMetrics returns current metrics
func (ac *AlignmentCompressor) GetMetrics() *AlignmentMetrics {
	ac.metrics.mu.RLock()
	defer ac.metrics.mu.RUnlock()

	return &AlignmentMetrics{
		TotalCompressions:      ac.metrics.TotalCompressions,
		TotalDecompressions:    ac.metrics.TotalDecompressions,
		TotalBytesCompressed:   ac.metrics.TotalBytesCompressed,
		TotalBytesDecompressed: ac.metrics.TotalBytesDecompressed,
		AverageCompressionTime: ac.metrics.AverageCompressionTime,
		LatencyReduction:       ac.metrics.LatencyReduction,
		CommunicationSavings:   ac.metrics.CommunicationSavings,
	}
}

// Stop stops the alignment compressor
func (ac *AlignmentCompressor) Stop() {
	ac.cancel()
	ac.wg.Wait()
}


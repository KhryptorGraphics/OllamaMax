package distributed

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// ChunkedCommunicator implements chunked non-blocking communication
// Overlaps communication with computation to hide latency
type ChunkedCommunicator struct {
	mu sync.RWMutex

	// Configuration
	config *ChunkedCommConfig

	// Active transfers
	activeTransfers map[string]*ChunkedTransfer

	// Completion channels
	completionChans map[string]chan *TransferResult

	// Performance metrics
	metrics *ChunkedCommMetrics
}

// ChunkedTransfer represents an active chunked transfer
type ChunkedTransfer struct {
	TransferID    string
	SourcePeer    peer.ID
	DestPeer      peer.ID
	TotalSize     uint64
	ChunkSize     uint64
	TotalChunks   int
	SentChunks    int
	ReceivedChunks int
	StartTime     time.Time
	LastChunkTime time.Time
	Status        TransferStatus
	mu            sync.RWMutex
}

// TransferStatus represents the status of a transfer
type TransferStatus int

const (
	TransferPending TransferStatus = iota
	TransferInProgress
	TransferComplete
	TransferFailed
)

// TransferResult contains the result of a transfer
type TransferResult struct {
	TransferID string
	Success    bool
	Data       []byte
	Error      error
	Duration   time.Duration
}

// ChunkedCommConfig configures chunked communication behavior
type ChunkedCommConfig struct {
	ChunkSize           uint64        // Size of each chunk (64KB recommended)
	MaxConcurrentChunks int           // Maximum concurrent chunks in flight
	TimeoutPerChunk     time.Duration // Timeout for each chunk
	EnableCompression   bool          // Enable compression for chunks
	EnablePipelining    bool          // Enable pipelined chunk sending
	RetryAttempts       int           // Number of retry attempts for failed chunks
}

// ChunkedCommMetrics tracks performance metrics
type ChunkedCommMetrics struct {
	mu                   sync.RWMutex
	TotalTransfers       int64
	CompletedTransfers   int64
	FailedTransfers      int64
	TotalBytesTransferred uint64
	AverageChunkLatency  time.Duration
	AverageThroughput    float64 // MB/s
}

// DefaultChunkedCommConfig returns default configuration
func DefaultChunkedCommConfig() *ChunkedCommConfig {
	return &ChunkedCommConfig{
		ChunkSize:           64 * 1024, // 64KB
		MaxConcurrentChunks: 8,
		TimeoutPerChunk:     5 * time.Second,
		EnableCompression:   true,
		EnablePipelining:    true,
		RetryAttempts:       3,
	}
}

// NewChunkedCommunicator creates a new chunked communicator
func NewChunkedCommunicator(config *ChunkedCommConfig) *ChunkedCommunicator {
	if config == nil {
		config = DefaultChunkedCommConfig()
	}

	return &ChunkedCommunicator{
		config:          config,
		activeTransfers: make(map[string]*ChunkedTransfer),
		completionChans: make(map[string]chan *TransferResult),
		metrics:         &ChunkedCommMetrics{},
	}
}

// SendChunked sends data in chunks to a peer
func (cc *ChunkedCommunicator) SendChunked(
	ctx context.Context,
	destPeer peer.ID,
	data []byte,
	transferID string,
) (<-chan *TransferResult, error) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	// Create transfer
	totalChunks := (len(data) + int(cc.config.ChunkSize) - 1) / int(cc.config.ChunkSize)
	transfer := &ChunkedTransfer{
		TransferID:  transferID,
		DestPeer:    destPeer,
		TotalSize:   uint64(len(data)),
		ChunkSize:   cc.config.ChunkSize,
		TotalChunks: totalChunks,
		SentChunks:  0,
		StartTime:   time.Now(),
		Status:      TransferPending,
	}

	cc.activeTransfers[transferID] = transfer

	// Create completion channel
	completionChan := make(chan *TransferResult, 1)
	cc.completionChans[transferID] = completionChan

	// Start chunked transfer in background
	go cc.performChunkedTransfer(ctx, transfer, data, completionChan)

	return completionChan, nil
}

// performChunkedTransfer performs the actual chunked transfer
func (cc *ChunkedCommunicator) performChunkedTransfer(
	ctx context.Context,
	transfer *ChunkedTransfer,
	data []byte,
	resultChan chan<- *TransferResult,
) {
	transfer.mu.Lock()
	transfer.Status = TransferInProgress
	transfer.mu.Unlock()

	startTime := time.Now()

	// Send chunks
	for i := 0; i < transfer.TotalChunks; i++ {
		select {
		case <-ctx.Done():
			resultChan <- &TransferResult{
				TransferID: transfer.TransferID,
				Success:    false,
				Error:      ctx.Err(),
			}
			return
		default:
		}

		// Calculate chunk boundaries
		start := i * int(cc.config.ChunkSize)
		end := min(start+int(cc.config.ChunkSize), len(data))
		chunk := data[start:end]

		// Send chunk (placeholder - actual implementation would use libp2p)
		err := cc.sendChunk(ctx, transfer.DestPeer, chunk, i)
		if err != nil {
			resultChan <- &TransferResult{
				TransferID: transfer.TransferID,
				Success:    false,
				Error:      fmt.Errorf("failed to send chunk %d: %w", i, err),
			}
			return
		}

		transfer.mu.Lock()
		transfer.SentChunks++
		transfer.LastChunkTime = time.Now()
		transfer.mu.Unlock()
	}

	// Transfer complete
	duration := time.Since(startTime)
	transfer.mu.Lock()
	transfer.Status = TransferComplete
	transfer.mu.Unlock()

	// Update metrics
	cc.updateMetrics(transfer, duration, true)

	resultChan <- &TransferResult{
		TransferID: transfer.TransferID,
		Success:    true,
		Duration:   duration,
	}
}

// sendChunk sends a single chunk (placeholder implementation)
func (cc *ChunkedCommunicator) sendChunk(ctx context.Context, destPeer peer.ID, chunk []byte, chunkIndex int) error {
	// Placeholder - actual implementation would use libp2p streams
	// Apply compression if enabled
	if cc.config.EnableCompression {
		// Compress chunk
		_ = chunk
	}

	// Simulate network delay
	time.Sleep(1 * time.Millisecond)

	return nil
}

// ReceiveChunked receives data in chunks from a peer
func (cc *ChunkedCommunicator) ReceiveChunked(
	ctx context.Context,
	transferID string,
	expectedSize uint64,
) ([]byte, error) {
	// Placeholder implementation
	// Actual implementation would receive chunks and reassemble
	return make([]byte, expectedSize), nil
}

// GetTransferStatus returns the status of a transfer
func (cc *ChunkedCommunicator) GetTransferStatus(transferID string) (*ChunkedTransfer, error) {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	transfer, exists := cc.activeTransfers[transferID]
	if !exists {
		return nil, fmt.Errorf("transfer %s not found", transferID)
	}

	return transfer, nil
}

// updateMetrics updates performance metrics
func (cc *ChunkedCommunicator) updateMetrics(transfer *ChunkedTransfer, duration time.Duration, success bool) {
	cc.metrics.mu.Lock()
	defer cc.metrics.mu.Unlock()

	cc.metrics.TotalTransfers++
	if success {
		cc.metrics.CompletedTransfers++
	} else {
		cc.metrics.FailedTransfers++
	}

	cc.metrics.TotalBytesTransferred += transfer.TotalSize

	// Calculate throughput (MB/s)
	throughput := float64(transfer.TotalSize) / (1024 * 1024) / duration.Seconds()
	
	// Update average throughput (exponential moving average)
	alpha := 0.3
	cc.metrics.AverageThroughput = cc.metrics.AverageThroughput*(1-alpha) + throughput*alpha

	// Update average chunk latency
	chunkLatency := duration / time.Duration(transfer.TotalChunks)
	cc.metrics.AverageChunkLatency = time.Duration(
		float64(cc.metrics.AverageChunkLatency)*(1-alpha) + float64(chunkLatency)*alpha,
	)
}

// GetMetrics returns current performance metrics
func (cc *ChunkedCommunicator) GetMetrics() *ChunkedCommMetrics {
	cc.metrics.mu.RLock()
	defer cc.metrics.mu.RUnlock()

	return &ChunkedCommMetrics{
		TotalTransfers:        cc.metrics.TotalTransfers,
		CompletedTransfers:    cc.metrics.CompletedTransfers,
		FailedTransfers:       cc.metrics.FailedTransfers,
		TotalBytesTransferred: cc.metrics.TotalBytesTransferred,
		AverageChunkLatency:   cc.metrics.AverageChunkLatency,
		AverageThroughput:     cc.metrics.AverageThroughput,
	}
}


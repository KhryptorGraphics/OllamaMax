package protocols

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

// TensorStreamClient handles high-performance tensor streaming between nodes
type TensorStreamClient struct {
	protocol       *TensorStreamProtocol
	connectionPool map[peer.ID]*StreamConnection
	poolMutex      sync.RWMutex
	activeStreams  map[string]*ClientStream
	streamsMutex   sync.RWMutex
	bandwidthMgr   *host.BandwidthManager
	compressor     *TensorCompressor
	retryConfig    *RetryConfig
	metrics        *StreamingMetrics
}

// StreamConnection represents a pooled connection to a peer
type StreamConnection struct {
	PeerID      peer.ID
	LastUsed    time.Time
	InUse       bool
	StreamCount int
	MaxStreams  int
	Health      ConnectionHealth
	mutex       sync.RWMutex
}

// ConnectionHealth tracks the health of a stream connection
type ConnectionHealth struct {
	Latency    time.Duration
	PacketLoss float64
	Bandwidth  int64
	ErrorCount int
	LastCheck  time.Time
	Status     HealthStatus
}

// ClientStream tracks a client-side tensor stream
type ClientStream struct {
	ID           string
	PeerID       peer.ID
	Header       ActivationHeader
	ChunksSent   int32
	TotalChunks  int32
	BytesSent    int64
	StartTime    time.Time
	LastActivity time.Time
	Status       StreamStatus
	Error        error
	mutex        sync.RWMutex
}

// StreamStatus represents the status of a client stream
type StreamStatus uint8

const (
	StreamStatusInitialized StreamStatus = iota
	StreamStatusStreaming
	StreamStatusCompleted
	StreamStatusFailed
	StreamStatusCancelled
)

// RetryConfig configures retry behavior for failed transfers
type RetryConfig struct {
	MaxRetries      int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	RetryableErrors []string
}

// StreamingMetrics tracks performance metrics for tensor streaming
type StreamingMetrics struct {
	TotalStreams     int64
	ActiveStreams    int64
	CompletedStreams int64
	FailedStreams    int64
	TotalBytes       int64
	AverageLatency   time.Duration
	Throughput       float64
	CompressionRatio float64
	mutex            sync.RWMutex
}

// NewTensorStreamClient creates a new tensor streaming client
func NewTensorStreamClient(protocol *TensorStreamProtocol, bandwidthMgr *host.BandwidthManager) *TensorStreamClient {
	return &TensorStreamClient{
		protocol:       protocol,
		connectionPool: make(map[peer.ID]*StreamConnection),
		activeStreams:  make(map[string]*ClientStream),
		bandwidthMgr:   bandwidthMgr,
		compressor:     NewTensorCompressor(),
		retryConfig:    NewDefaultRetryConfig(),
		metrics:        NewStreamingMetrics(),
	}
}

// NewDefaultRetryConfig creates a default retry configuration
func NewDefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:    3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
		RetryableErrors: []string{
			"connection timeout",
			"temporary failure",
			"network unreachable",
		},
	}
}

// NewStreamingMetrics creates a new streaming metrics tracker
func NewStreamingMetrics() *StreamingMetrics {
	return &StreamingMetrics{}
}

// StreamActivation streams activation data to a target node
func (tsc *TensorStreamClient) StreamActivation(targetPeer peer.ID, data *ActivationData) (string, int32, error) {
	ctx := context.Background()

	// Get or create connection
	_, err := tsc.getConnection(targetPeer)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get connection to peer %s: %w", targetPeer, err)
	}

	// Create simple metadata for compression
	metadata := &TensorMetadata{
		Shape:           []int64{int64(len(data.Data))},
		DType:           DTypeFloat32,
		CompressionType: CompressionGzip,
		OriginalSize:    int64(len(data.Data)),
	}

	// Compress data
	compressedData, err := tsc.compressor.CompressActivation(data.Data, metadata)
	if err != nil {
		return "", 0, fmt.Errorf("failed to compress activation: %w", err)
	}

	// Generate sequence number for consistent ID generation
	sequenceNum := tsc.generateSequenceNumber()

	// Create activation header
	header := ActivationHeader{
		InferenceID: data.InferenceID,
		PartitionID: data.StageID,
		TensorShape: metadata.Shape,
		DType:       metadata.DType,
		Compression: metadata.CompressionType,
		TotalSize:   int64(len(compressedData)),
		ChunkCount:  tsc.calculateChunkCount(int64(len(compressedData))),
		SequenceNum: sequenceNum,
		Timestamp:   time.Now(),
		SourcePeer:  "", // Will be set by the protocol handler
		TargetPeer:  targetPeer,
		Priority:    1, // Default priority
	}

	// Create client stream with consistent stream ID format
	streamID := fmt.Sprintf("%s-%s-%d", header.InferenceID, header.PartitionID, header.SequenceNum)
	clientStream := &ClientStream{
		ID:           streamID,
		PeerID:       targetPeer,
		Header:       header,
		TotalChunks:  header.ChunkCount,
		StartTime:    time.Now(),
		LastActivity: time.Now(),
		Status:       StreamStatusInitialized,
	}

	tsc.streamsMutex.Lock()
	tsc.activeStreams[streamID] = clientStream
	tsc.streamsMutex.Unlock()

	// Update metrics
	tsc.updateMetrics(func(m *StreamingMetrics) {
		m.TotalStreams++
		m.ActiveStreams++
	})

	// Send activation_start message
	startMsg := &ActivationStreamMessage{
		Type:      MsgTypeActivationStart,
		Header:    &header,
		Timestamp: time.Now(),
	}

	err = tsc.protocol.SendMessage(ctx, targetPeer, startMsg)
	if err != nil {
		return "", 0, fmt.Errorf("failed to send activation start: %w", err)
	}

	// Stream the data chunks
	err = tsc.streamTensorChunks(clientStream, compressedData)
	if err != nil {
		return "", 0, err
	}

	// Send activation_complete message
	completeMsg := &ActivationStreamMessage{
		Type:      MsgTypeActivationComplete,
		Header:    &header,
		Timestamp: time.Now(),
	}

	err = tsc.protocol.SendMessage(ctx, targetPeer, completeMsg)
	if err != nil {
		return "", 0, err
	}

	// Return streamID and sequenceNum for aligned identification
	return streamID, header.SequenceNum, nil
}

// getConnection gets or creates a connection to a peer
func (tsc *TensorStreamClient) getConnection(peerID peer.ID) (*StreamConnection, error) {
	tsc.poolMutex.Lock()
	defer tsc.poolMutex.Unlock()

	// Check if connection exists and is healthy
	if conn, exists := tsc.connectionPool[peerID]; exists {
		conn.mutex.Lock()
		defer conn.mutex.Unlock()

		if conn.Health.Status == HealthStatusHealthy && !conn.InUse {
			conn.InUse = true
			conn.LastUsed = time.Now()
			return conn, nil
		}
	}

	// Create new connection
	conn := &StreamConnection{
		PeerID:      peerID,
		LastUsed:    time.Now(),
		InUse:       true,
		StreamCount: 0,
		MaxStreams:  10,
		Health: ConnectionHealth{
			Status:    HealthStatusHealthy,
			LastCheck: time.Now(),
		},
	}

	tsc.connectionPool[peerID] = conn
	return conn, nil
}

// calculateChunkCount calculates the number of chunks needed for data
func (tsc *TensorStreamClient) calculateChunkCount(dataSize int64) int32 {
	chunkSize := int64(tsc.protocol.chunkSize)
	chunks := (dataSize + chunkSize - 1) / chunkSize
	return int32(chunks)
}

// generateSequenceNumber generates a unique sequence number
func (tsc *TensorStreamClient) generateSequenceNumber() int32 {
	return int32(time.Now().UnixNano() % 1000000)
}

// streamTensorChunks streams tensor data in chunks
func (tsc *TensorStreamClient) streamTensorChunks(stream *ClientStream, data []byte) error {
	stream.mutex.Lock()
	stream.Status = StreamStatusStreaming
	stream.mutex.Unlock()

	chunkSize := int(tsc.protocol.chunkSize)
	totalSize := len(data)

	for i := 0; i < int(stream.TotalChunks); i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > totalSize {
			end = totalSize
		}

		chunk := &ActivationChunk{
			Header:     stream.Header,
			ChunkIndex: int32(i),
			ChunkSize:  int32(end - start),
			Data:       data[start:end],
			Checksum:   tsc.calculateChecksum(data[start:end]),
		}

		// Send chunk with retry logic
		err := tsc.sendChunkWithRetry(stream, chunk)
		if err != nil {
			stream.mutex.Lock()
			stream.Status = StreamStatusFailed
			stream.Error = err
			stream.mutex.Unlock()

			tsc.updateMetrics(func(m *StreamingMetrics) {
				m.FailedStreams++
				m.ActiveStreams--
			})

			return fmt.Errorf("failed to send chunk %d: %w", i, err)
		}

		stream.mutex.Lock()
		stream.ChunksSent++
		stream.BytesSent += int64(end - start)
		stream.LastActivity = time.Now()
		stream.mutex.Unlock()

		// Apply bandwidth throttling
		if tsc.bandwidthMgr != nil {
			// Check bandwidth availability for tensor streaming
			tsc.bandwidthMgr.CheckBandwidth(stream.PeerID, "tensor-stream", int64(end-start))
		}
	}

	// Mark stream as completed
	stream.mutex.Lock()
	stream.Status = StreamStatusCompleted
	stream.mutex.Unlock()

	tsc.updateMetrics(func(m *StreamingMetrics) {
		m.CompletedStreams++
		m.ActiveStreams--
		m.TotalBytes += stream.BytesSent
	})

	log.Printf("Successfully streamed %d chunks (%d bytes) to peer %s",
		stream.ChunksSent, stream.BytesSent, stream.PeerID)

	return nil
}

// sendChunkWithRetry sends a chunk with retry logic
func (tsc *TensorStreamClient) sendChunkWithRetry(stream *ClientStream, chunk *ActivationChunk) error {
	var lastErr error
	delay := tsc.retryConfig.InitialDelay

	for attempt := 0; attempt <= tsc.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(delay)
			delay = time.Duration(float64(delay) * tsc.retryConfig.BackoffFactor)
			if delay > tsc.retryConfig.MaxDelay {
				delay = tsc.retryConfig.MaxDelay
			}
		}

		err := tsc.sendChunk(stream, chunk)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !tsc.isRetryableError(err) {
			break
		}

		log.Printf("Attempt %d failed for chunk %d to peer %s: %v",
			attempt+1, chunk.ChunkIndex, stream.PeerID, err)
	}

	return fmt.Errorf("failed after %d attempts: %w", tsc.retryConfig.MaxRetries+1, lastErr)
}

// sendChunk sends a single chunk to the peer
func (tsc *TensorStreamClient) sendChunk(stream *ClientStream, chunk *ActivationChunk) error {
	ctx := context.Background()

	// Create activation stream message
	msg := &ActivationStreamMessage{
		Type:      MsgTypeActivationChunk,
		Chunk:     chunk,
		Timestamp: time.Now(),
	}

	// Send via protocol using the new signature
	return tsc.protocol.SendMessage(ctx, stream.PeerID, msg)
}

// calculateChecksum calculates a checksum for chunk data
func (tsc *TensorStreamClient) calculateChecksum(data []byte) string {
	// Simple checksum implementation - in production use SHA256 or similar
	sum := uint32(0)
	for _, b := range data {
		sum += uint32(b)
	}
	return fmt.Sprintf("%08x", sum)
}

// isRetryableError checks if an error is retryable
func (tsc *TensorStreamClient) isRetryableError(err error) bool {
	errStr := err.Error()
	for _, retryableErr := range tsc.retryConfig.RetryableErrors {
		if strings.Contains(errStr, retryableErr) {
			return true
		}
	}
	return false
}

// updateMetrics updates streaming metrics
func (tsc *TensorStreamClient) updateMetrics(updateFunc func(*StreamingMetrics)) {
	tsc.metrics.mutex.Lock()
	defer tsc.metrics.mutex.Unlock()
	updateFunc(tsc.metrics)
}

// GetStreamingMetrics returns current streaming metrics
func (tsc *TensorStreamClient) GetStreamingMetrics() *StreamingMetrics {
	tsc.metrics.mutex.RLock()
	defer tsc.metrics.mutex.RUnlock()

	// Return a copy to avoid race conditions
	return &StreamingMetrics{
		TotalStreams:     tsc.metrics.TotalStreams,
		ActiveStreams:    tsc.metrics.ActiveStreams,
		CompletedStreams: tsc.metrics.CompletedStreams,
		FailedStreams:    tsc.metrics.FailedStreams,
		TotalBytes:       tsc.metrics.TotalBytes,
		AverageLatency:   tsc.metrics.AverageLatency,
		Throughput:       tsc.metrics.Throughput,
		CompressionRatio: tsc.metrics.CompressionRatio,
	}
}

// CleanupCompletedStreams removes completed streams from memory
func (tsc *TensorStreamClient) CleanupCompletedStreams() {
	tsc.streamsMutex.Lock()
	defer tsc.streamsMutex.Unlock()

	for streamID, stream := range tsc.activeStreams {
		stream.mutex.RLock()
		if stream.Status == StreamStatusCompleted || stream.Status == StreamStatusFailed {
			delete(tsc.activeStreams, streamID)
		}
		stream.mutex.RUnlock()
	}
}

// Close closes the tensor stream client and cleans up resources
func (tsc *TensorStreamClient) Close() error {
	// Cancel all active streams
	tsc.streamsMutex.Lock()
	for _, stream := range tsc.activeStreams {
		stream.mutex.Lock()
		stream.Status = StreamStatusCancelled
		stream.mutex.Unlock()
	}
	tsc.streamsMutex.Unlock()

	// Close all connections
	tsc.poolMutex.Lock()
	for peerID := range tsc.connectionPool {
		delete(tsc.connectionPool, peerID)
	}
	tsc.poolMutex.Unlock()

	return nil
}

// RegisterResponseChannel registers a response channel for the given stream ID
func (tsc *TensorStreamClient) RegisterResponseChannel(streamID string) (chan *ActivationData, func()) {
	return tsc.protocol.RegisterResponseChannel(streamID)
}

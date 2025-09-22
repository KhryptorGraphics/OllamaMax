package protocols

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

// ActivationStreamManager coordinates all aspects of tensor streaming during distributed inference
type ActivationStreamManager struct {
	streamProtocol *TensorStreamProtocol
	streamClient   *TensorStreamClient
	pipelineCoord  *PipelineCoordinator
	bandwidthMgr   *host.BandwidthManager
	activeStreams  map[string]*ManagedStream
	streamsMutex   sync.RWMutex
	streamPools    map[peer.ID]*StreamPool
	poolsMutex     sync.RWMutex
	healthMonitor  *StreamHealthMonitor
	performanceOpt *StreamPerformanceOptimizer
	ctx            context.Context
	cancel         context.CancelFunc
	config         *StreamManagerConfig
}

// ManagedStream represents a managed tensor stream with lifecycle tracking
type ManagedStream struct {
	ID                string
	InferenceID       string
	SourcePeer        peer.ID
	TargetPeer        peer.ID
	StreamType        StreamType
	Priority          StreamPriority
	Status            ManagedStreamStatus
	StartTime         time.Time
	LastActivity      time.Time
	BytesTransferred  int64
	ChunksTransferred int32
	ErrorCount        int
	RetryCount        int
	Health            StreamHealth
	Performance       StreamPerformance
	mutex             sync.RWMutex
}

// StreamType defines the type of tensor stream
type StreamType uint8

const (
	StreamTypeActivation StreamType = iota
	StreamTypeGradient
	StreamTypeParameter
	StreamTypeMetadata
)

// StreamPriority defines the priority level of a stream
type StreamPriority uint8

const (
	StreamPriorityLow StreamPriority = iota
	StreamPriorityNormal
	StreamPriorityHigh
	StreamPriorityCritical
)

// ManagedStreamStatus represents the status of a managed stream
type ManagedStreamStatus uint8

const (
	ManagedStreamStatusCreated ManagedStreamStatus = iota
	ManagedStreamStatusActive
	ManagedStreamStatusPaused
	ManagedStreamStatusCompleted
	ManagedStreamStatusFailed
	ManagedStreamStatusCancelled
)

// StreamHealth tracks the health metrics of a stream
type StreamHealth struct {
	Latency         time.Duration
	Throughput      float64
	ErrorRate       float64
	PacketLoss      float64
	LastHealthCheck time.Time
	HealthScore     float64
}

// StreamPerformance tracks performance metrics
type StreamPerformance struct {
	AverageChunkTime   time.Duration
	CompressionRatio   float64
	NetworkUtilization float64
	CPUUtilization     float64
	MemoryUsage        int64
}

// StreamPool manages multiple streams to a specific peer
type StreamPool struct {
	PeerID         peer.ID
	MaxStreams     int
	ActiveStreams  map[string]*ManagedStream
	AvailableSlots int
	TotalBandwidth int64
	UsedBandwidth  int64
	LastOptimized  time.Time
	mutex          sync.RWMutex
}

// StreamHealthMonitor monitors the health of all streams
type StreamHealthMonitor struct {
	streams         map[string]*ManagedStream
	mutex           sync.RWMutex
	checkInterval   time.Duration
	alertThresholds *HealthThresholds
}

// HealthThresholds defines thresholds for health alerts
type HealthThresholds struct {
	MaxLatency     time.Duration
	MinThroughput  float64
	MaxErrorRate   float64
	MaxPacketLoss  float64
	MinHealthScore float64
}

// StreamPerformanceOptimizer optimizes stream performance
type StreamPerformanceOptimizer struct {
	optimizationRules map[string]*OptimizationRule
	lastOptimization  time.Time
	mutex             sync.RWMutex
}

// OptimizationRule defines a performance optimization rule
type OptimizationRule struct {
	Name      string
	Condition func(*ManagedStream) bool
	Action    func(*ManagedStream) error
	Priority  int
	Enabled   bool
}

// StreamManagerConfig configures the stream manager
type StreamManagerConfig struct {
	MaxConcurrentStreams int
	StreamTimeout        time.Duration
	HealthCheckInterval  time.Duration
	OptimizationInterval time.Duration
	RetryAttempts        int
	BufferSize           int
	CompressionEnabled   bool
	PriorityScheduling   bool
}

// NewActivationStreamManager creates a new activation stream manager
func NewActivationStreamManager(
	streamProtocol *TensorStreamProtocol,
	streamClient *TensorStreamClient,
	pipelineCoord *PipelineCoordinator,
	bandwidthMgr *host.BandwidthManager,
) *ActivationStreamManager {
	ctx, cancel := context.WithCancel(context.Background())

	config := &StreamManagerConfig{
		MaxConcurrentStreams: 100,
		StreamTimeout:        30 * time.Second,
		HealthCheckInterval:  5 * time.Second,
		OptimizationInterval: 10 * time.Second,
		RetryAttempts:        3,
		BufferSize:           1024 * 1024, // 1MB
		CompressionEnabled:   true,
		PriorityScheduling:   true,
	}

	asm := &ActivationStreamManager{
		streamProtocol: streamProtocol,
		streamClient:   streamClient,
		pipelineCoord:  pipelineCoord,
		bandwidthMgr:   bandwidthMgr,
		activeStreams:  make(map[string]*ManagedStream),
		streamPools:    make(map[peer.ID]*StreamPool),
		healthMonitor:  NewStreamHealthMonitor(),
		performanceOpt: NewStreamPerformanceOptimizer(),
		ctx:            ctx,
		cancel:         cancel,
		config:         config,
	}

	// Start background workers
	go asm.healthMonitorWorker()
	go asm.performanceOptimizerWorker()
	go asm.cleanupWorker()

	return asm
}

// NewStreamHealthMonitor creates a new stream health monitor
func NewStreamHealthMonitor() *StreamHealthMonitor {
	return &StreamHealthMonitor{
		streams:       make(map[string]*ManagedStream),
		checkInterval: 5 * time.Second,
		alertThresholds: &HealthThresholds{
			MaxLatency:     500 * time.Millisecond,
			MinThroughput:  10.0, // MB/s
			MaxErrorRate:   0.05, // 5%
			MaxPacketLoss:  0.01, // 1%
			MinHealthScore: 0.7,  // 70%
		},
	}
}

// NewStreamPerformanceOptimizer creates a new performance optimizer
func NewStreamPerformanceOptimizer() *StreamPerformanceOptimizer {
	optimizer := &StreamPerformanceOptimizer{
		optimizationRules: make(map[string]*OptimizationRule),
	}

	// Add default optimization rules
	optimizer.addDefaultRules()

	return optimizer
}

// addDefaultRules adds default optimization rules
func (spo *StreamPerformanceOptimizer) addDefaultRules() {
	// Rule: Increase compression for slow connections
	spo.optimizationRules["compress_slow_connections"] = &OptimizationRule{
		Name: "Compress Slow Connections",
		Condition: func(stream *ManagedStream) bool {
			return stream.Health.Throughput < 5.0 // Less than 5 MB/s
		},
		Action: func(stream *ManagedStream) error {
			log.Printf("Applying compression optimization to stream %s", stream.ID)
			// Implementation would adjust compression settings
			return nil
		},
		Priority: 1,
		Enabled:  true,
	}

	// Rule: Reduce chunk size for high latency connections
	spo.optimizationRules["reduce_chunk_size_high_latency"] = &OptimizationRule{
		Name: "Reduce Chunk Size for High Latency",
		Condition: func(stream *ManagedStream) bool {
			return stream.Health.Latency > 200*time.Millisecond
		},
		Action: func(stream *ManagedStream) error {
			log.Printf("Reducing chunk size for high latency stream %s", stream.ID)
			// Implementation would adjust chunk size
			return nil
		},
		Priority: 2,
		Enabled:  true,
	}
}

// CreateStream creates a new managed stream (overloaded method for compatibility)
func (asm *ActivationStreamManager) CreateStream(streamID string, inferenceID string, sourcePeer, targetPeer peer.ID) (*ManagedStream, error) {
	return asm.createStreamWithOptions(streamID, inferenceID, sourcePeer, targetPeer, StreamTypeActivation, StreamPriorityNormal)
}

// createStreamWithOptions creates a new managed stream with full options
func (asm *ActivationStreamManager) createStreamWithOptions(
	streamID string,
	inferenceID string,
	sourcePeer, targetPeer peer.ID,
	streamType StreamType,
	priority StreamPriority,
) (*ManagedStream, error) {

	// Check if we can create more streams
	if len(asm.activeStreams) >= asm.config.MaxConcurrentStreams {
		return nil, fmt.Errorf("maximum concurrent streams reached")
	}

	// Use provided streamID or generate one if empty
	if streamID == "" {
		streamID = fmt.Sprintf("stream-%s-%d", inferenceID, time.Now().UnixNano())
	}

	stream := &ManagedStream{
		ID:           streamID,
		InferenceID:  inferenceID,
		SourcePeer:   sourcePeer,
		TargetPeer:   targetPeer,
		StreamType:   streamType,
		Priority:     priority,
		Status:       ManagedStreamStatusCreated,
		StartTime:    time.Now(),
		LastActivity: time.Now(),
		Health: StreamHealth{
			LastHealthCheck: time.Now(),
			HealthScore:     1.0,
		},
	}

	asm.streamsMutex.Lock()
	asm.activeStreams[streamID] = stream
	asm.streamsMutex.Unlock()

	// Add to stream pool
	asm.addToStreamPool(targetPeer, stream)

	// Register with health monitor
	asm.healthMonitor.RegisterStream(stream)

	log.Printf("Created managed stream %s for inference %s", streamID, inferenceID)
	return stream, nil
}

// addToStreamPool adds a stream to the appropriate stream pool
func (asm *ActivationStreamManager) addToStreamPool(peerID peer.ID, stream *ManagedStream) {
	asm.poolsMutex.Lock()
	defer asm.poolsMutex.Unlock()

	pool, exists := asm.streamPools[peerID]
	if !exists {
		pool = &StreamPool{
			PeerID:         peerID,
			MaxStreams:     10,
			ActiveStreams:  make(map[string]*ManagedStream),
			AvailableSlots: 10,
			TotalBandwidth: 100 * 1024 * 1024, // 100 MB/s default
		}
		asm.streamPools[peerID] = pool
	}

	pool.mutex.Lock()
	pool.ActiveStreams[stream.ID] = stream
	pool.AvailableSlots--
	pool.mutex.Unlock()
}

// StartStream starts a managed stream
func (asm *ActivationStreamManager) StartStream(streamID string, data *ActivationData) error {
	asm.streamsMutex.RLock()
	stream, exists := asm.activeStreams[streamID]
	asm.streamsMutex.RUnlock()

	if !exists {
		return fmt.Errorf("stream %s not found", streamID)
	}

	stream.mutex.Lock()
	stream.Status = ManagedStreamStatusActive
	stream.LastActivity = time.Now()
	stream.mutex.Unlock()

	// Use stream client to perform the actual streaming
	_, _, err := asm.streamClient.StreamActivation(stream.TargetPeer, data)
	return err
}

// UpdateStreamStatus updates the status of a managed stream
func (asm *ActivationStreamManager) UpdateStreamStatus(streamID string, status ManagedStreamStatus) error {
	asm.streamsMutex.RLock()
	stream, exists := asm.activeStreams[streamID]
	asm.streamsMutex.RUnlock()

	if !exists {
		return fmt.Errorf("stream %s not found", streamID)
	}

	stream.mutex.Lock()
	stream.Status = status
	stream.LastActivity = time.Now()
	stream.mutex.Unlock()

	log.Printf("Updated stream %s status to %d", streamID, status)
	return nil
}

// RegisterStream registers a stream with the health monitor
func (shm *StreamHealthMonitor) RegisterStream(stream *ManagedStream) {
	shm.mutex.Lock()
	defer shm.mutex.Unlock()

	shm.streams[stream.ID] = stream
}

// healthMonitorWorker runs the health monitoring loop
func (asm *ActivationStreamManager) healthMonitorWorker() {
	ticker := time.NewTicker(asm.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-asm.ctx.Done():
			return
		case <-ticker.C:
			asm.performHealthChecks()
		}
	}
}

// performHealthChecks performs health checks on all active streams
func (asm *ActivationStreamManager) performHealthChecks() {
	asm.streamsMutex.RLock()
	streams := make([]*ManagedStream, 0, len(asm.activeStreams))
	for _, stream := range asm.activeStreams {
		streams = append(streams, stream)
	}
	asm.streamsMutex.RUnlock()

	for _, stream := range streams {
		asm.checkStreamHealth(stream)
	}
}

// checkStreamHealth checks the health of a specific stream
func (asm *ActivationStreamManager) checkStreamHealth(stream *ManagedStream) {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()

	now := time.Now()

	// Check for timeout
	if now.Sub(stream.LastActivity) > asm.config.StreamTimeout {
		stream.Status = ManagedStreamStatusFailed
		log.Printf("Stream %s timed out", stream.ID)
		return
	}

	// Update health metrics (simplified)
	stream.Health.LastHealthCheck = now

	// Calculate health score based on various factors
	healthScore := 1.0
	if stream.Health.Latency > asm.healthMonitor.alertThresholds.MaxLatency {
		healthScore -= 0.2
	}
	if stream.Health.Throughput < asm.healthMonitor.alertThresholds.MinThroughput {
		healthScore -= 0.3
	}
	if stream.Health.ErrorRate > asm.healthMonitor.alertThresholds.MaxErrorRate {
		healthScore -= 0.4
	}

	stream.Health.HealthScore = healthScore

	// Trigger alerts if health is poor
	if healthScore < asm.healthMonitor.alertThresholds.MinHealthScore {
		log.Printf("ALERT: Stream %s health score is low: %.2f", stream.ID, healthScore)
	}
}

// performanceOptimizerWorker runs the performance optimization loop
func (asm *ActivationStreamManager) performanceOptimizerWorker() {
	ticker := time.NewTicker(asm.config.OptimizationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-asm.ctx.Done():
			return
		case <-ticker.C:
			asm.optimizePerformance()
		}
	}
}

// optimizePerformance applies performance optimizations to streams
func (asm *ActivationStreamManager) optimizePerformance() {
	asm.streamsMutex.RLock()
	streams := make([]*ManagedStream, 0, len(asm.activeStreams))
	for _, stream := range asm.activeStreams {
		if stream.Status == ManagedStreamStatusActive {
			streams = append(streams, stream)
		}
	}
	asm.streamsMutex.RUnlock()

	for _, stream := range streams {
		asm.performanceOpt.OptimizeStream(stream)
	}
}

// OptimizeStream applies optimization rules to a stream
func (spo *StreamPerformanceOptimizer) OptimizeStream(stream *ManagedStream) {
	spo.mutex.RLock()
	defer spo.mutex.RUnlock()

	for _, rule := range spo.optimizationRules {
		if rule.Enabled && rule.Condition(stream) {
			if err := rule.Action(stream); err != nil {
				log.Printf("Failed to apply optimization rule %s to stream %s: %v",
					rule.Name, stream.ID, err)
			}
		}
	}
}

// cleanupWorker runs the cleanup loop for completed streams
func (asm *ActivationStreamManager) cleanupWorker() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-asm.ctx.Done():
			return
		case <-ticker.C:
			asm.cleanupCompletedStreams()
		}
	}
}

// cleanupCompletedStreams removes completed or failed streams
func (asm *ActivationStreamManager) cleanupCompletedStreams() {
	asm.streamsMutex.Lock()
	defer asm.streamsMutex.Unlock()

	for streamID, stream := range asm.activeStreams {
		stream.mutex.RLock()
		shouldCleanup := stream.Status == ManagedStreamStatusCompleted ||
			stream.Status == ManagedStreamStatusFailed ||
			stream.Status == ManagedStreamStatusCancelled
		stream.mutex.RUnlock()

		if shouldCleanup {
			delete(asm.activeStreams, streamID)
			asm.removeFromStreamPool(stream.TargetPeer, streamID)
			log.Printf("Cleaned up stream %s", streamID)
		}
	}
}

// removeFromStreamPool removes a stream from its pool
func (asm *ActivationStreamManager) removeFromStreamPool(peerID peer.ID, streamID string) {
	asm.poolsMutex.Lock()
	defer asm.poolsMutex.Unlock()

	if pool, exists := asm.streamPools[peerID]; exists {
		pool.mutex.Lock()
		delete(pool.ActiveStreams, streamID)
		pool.AvailableSlots++
		pool.mutex.Unlock()
	}
}

// GetStreamStats returns statistics about managed streams
func (asm *ActivationStreamManager) GetStreamStats() map[string]interface{} {
	asm.streamsMutex.RLock()
	defer asm.streamsMutex.RUnlock()

	stats := map[string]interface{}{
		"total_streams":     len(asm.activeStreams),
		"active_streams":    0,
		"completed_streams": 0,
		"failed_streams":    0,
	}

	for _, stream := range asm.activeStreams {
		stream.mutex.RLock()
		switch stream.Status {
		case ManagedStreamStatusActive:
			stats["active_streams"] = stats["active_streams"].(int) + 1
		case ManagedStreamStatusCompleted:
			stats["completed_streams"] = stats["completed_streams"].(int) + 1
		case ManagedStreamStatusFailed:
			stats["failed_streams"] = stats["failed_streams"].(int) + 1
		}
		stream.mutex.RUnlock()
	}

	return stats
}

// Close shuts down the activation stream manager
func (asm *ActivationStreamManager) Close() error {
	asm.cancel()

	// Cancel all active streams
	asm.streamsMutex.Lock()
	for _, stream := range asm.activeStreams {
		stream.mutex.Lock()
		stream.Status = ManagedStreamStatusCancelled
		stream.mutex.Unlock()
	}
	asm.streamsMutex.Unlock()

	return nil
}

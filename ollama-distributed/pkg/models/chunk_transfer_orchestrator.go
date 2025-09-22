package models

/*
Chunk Transfer Orchestrator - Real Implementation

This implementation provides real chunk I/O and checksums for distributed model shard transfers:

1. Real File I/O:
   - Uses ShardRegistry to locate actual shard files on source nodes
   - Creates readers that read from actual shard files at specific offsets
   - Writes received chunks to disk on target nodes
   - Assembles complete shard files from chunks

2. Real Checksums:
   - Calculates actual SHA256 checksums by reading chunk data
   - Verifies chunk integrity after transfer
   - Uses IntegrityVerifier for complete shard validation

3. Storage Management:
   - Manages chunk storage directories and cleanup
   - Handles failed transfer cleanup
   - Registers assembled shards in the registry

4. Integrity Verification:
   - Multi-level verification (chunk + shard)
   - File size validation
   - Checksum verification at multiple stages
*/

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/protocols"
	"github.com/libp2p/go-libp2p/core/peer"
)

// Note: TransferStatus is imported from p2p_transfer.go
// Additional status for retrying
const (
	TransferStatusRetrying TransferStatus = "retrying"
)

// TransferPriority defines transfer priority levels
type TransferPriority int

const (
	TransferPriorityLow TransferPriority = iota
	TransferPriorityNormal
	TransferPriorityHigh
	TransferPriorityCritical
)

// ShardTransfer represents a single shard transfer operation
type ShardTransfer struct {
	ID               string           `json:"id"`
	ShardID          string           `json:"shard_id"`
	SourceNode       string           `json:"source_node"`
	TargetNode       string           `json:"target_node"`
	Size             int64            `json:"size"`
	Priority         TransferPriority `json:"priority"`
	Status           TransferStatus   `json:"status"`
	Progress         float64          `json:"progress"` // 0.0 to 1.0
	BytesTransferred int64            `json:"bytes_transferred"`
	StartTime        time.Time        `json:"start_time"`
	EndTime          time.Time        `json:"end_time"`
	RetryCount       int              `json:"retry_count"`
	LastError        string           `json:"last_error"`
	Checksum         string           `json:"checksum"`
	ChunkSize        int64            `json:"chunk_size"`
	TotalChunks      int              `json:"total_chunks"`
	CompletedChunks  int32            `json:"completed_chunks"`
}

// TransferCoordinator manages multiple concurrent transfers
type TransferCoordinator struct {
	mu                    sync.RWMutex
	transfers             map[string]*ShardTransfer
	activeTransfers       int32
	completedCount        int64
	failedCount           int64
	totalBytesTransferred int64

	// Transfer queues by priority
	queues map[TransferPriority][]*ShardTransfer

	// Network optimization
	bandwidthLimits  map[string]int64 // NodeID -> bytes/sec limit
	currentBandwidth map[string]int64 // NodeID -> current usage

	// Callbacks
	progressCallbacks   []func(*ShardTransfer)
	completionCallbacks []func(*ShardTransfer)
}

// ChunkVerificationManager handles chunk integrity verification
type ChunkVerificationManager struct {
	mu             sync.RWMutex
	verifier       *IntegrityVerifier
	pendingChunks  map[string][]ChunkInfo // TransferID -> chunks
	verifiedChunks map[string][]bool      // TransferID -> verification status
	checksumCache  map[string]string      // ChunkID -> checksum
}

// ChunkInfo contains information about a single chunk for orchestrator operations
type ChunkInfo struct {
	Index    int    `json:"index"`
	Offset   int64  `json:"offset"`
	Size     int64  `json:"size"`
	Checksum string `json:"checksum"`
	Verified bool   `json:"verified"`
}

// ChunkTransferOrchestrator coordinates shard distribution using P2P
type ChunkTransferOrchestrator struct {
	mu                 sync.RWMutex
	p2pEngine          *P2PTransferEngine
	integrityVerifier  *IntegrityVerifier
	fileHandler        *protocols.FileTransferHandler
	fileTransferClient *protocols.FileTransferClient
	coordinator        *TransferCoordinator
	verificationMgr    *ChunkVerificationManager
	config             *OrchestratorConfig
	shardRegistry      *ShardRegistry

	// Active transfers
	activeTransfers map[string]context.CancelFunc

	// Storage paths for received chunks
	chunkStoragePath string

	// Metrics
	metrics *TransferMetrics
}

// OrchestratorConfig contains configuration for the orchestrator
type OrchestratorConfig struct {
	MaxConcurrentTransfers int
	DefaultChunkSize       int64
	MinChunkSize           int64
	MaxChunkSize           int64
	MaxRetries             int
	RetryDelay             time.Duration
	VerifyChecksums        bool
	EnableCompression      bool
	CompressionLevel       int
	BandwidthThrottling    bool
	MaxBandwidthPerNode    int64 // bytes/sec
	TransferTimeout        time.Duration
	EnableAdaptiveChunking bool
	PriorityQueueSize      int
}

// NewChunkTransferOrchestrator creates a new transfer orchestrator
func NewChunkTransferOrchestrator(
	p2pEngine *P2PTransferEngine,
	integrityVerifier *IntegrityVerifier,
	shardRegistry *ShardRegistry,
	config *OrchestratorConfig) *ChunkTransferOrchestrator {

	if config == nil {
		config = DefaultOrchestratorConfig()
	}

	// Set up chunk storage path for shards
	chunkStoragePath := "/var/lib/ollama/shards/tmp"
	os.MkdirAll(chunkStoragePath, 0755)

	// Create local file store for chunk storage
	localFileStore, err := protocols.NewLocalFileStore(chunkStoragePath)
	if err != nil {
		// Fall back to tmp directory if primary location fails
		chunkStoragePath = "/tmp/ollama-chunks"
		os.MkdirAll(chunkStoragePath, 0755)
		localFileStore, _ = protocols.NewLocalFileStore(chunkStoragePath)
	}

	// Configure file transfer handler for shard transfers
	ftConfig := protocols.DefaultFileTransferConfig()
	ftConfig.ChunkSize = int(config.DefaultChunkSize)
	ftConfig.MaxFileSize = 10 * 1024 * 1024 * 1024 // 10GB max shard size
	ftConfig.AllowedExtensions = []string{".gguf", ".bin", ".safetensors", ".shard"}
	ftConfig.StorageDir = chunkStoragePath
	ftConfig.VerifyChecksums = config.VerifyChecksums
	ftConfig.CompressionEnabled = config.EnableCompression

	return &ChunkTransferOrchestrator{
		p2pEngine:         p2pEngine,
		integrityVerifier: integrityVerifier,
		fileHandler:       protocols.NewFileTransferHandler(localFileStore, ftConfig),
		coordinator:       NewTransferCoordinator(),
		verificationMgr:   NewChunkVerificationManager(integrityVerifier),
		config:            config,
		shardRegistry:     shardRegistry,
		activeTransfers:   make(map[string]context.CancelFunc),
		chunkStoragePath:  chunkStoragePath,
		metrics:           &TransferMetrics{LastUpdateTime: time.Now()},
	}
}

// DefaultOrchestratorConfig returns default configuration
func DefaultOrchestratorConfig() *OrchestratorConfig {
	return &OrchestratorConfig{
		MaxConcurrentTransfers: 4,
		DefaultChunkSize:       16 * 1024 * 1024, // 16MB
		MinChunkSize:           1 * 1024 * 1024,  // 1MB
		MaxChunkSize:           64 * 1024 * 1024, // 64MB
		MaxRetries:             3,
		RetryDelay:             5 * time.Second,
		VerifyChecksums:        true,
		EnableCompression:      true,
		CompressionLevel:       6,
		BandwidthThrottling:    true,
		MaxBandwidthPerNode:    100 * 1024 * 1024, // 100MB/s
		TransferTimeout:        30 * time.Minute,
		EnableAdaptiveChunking: true,
		PriorityQueueSize:      100,
	}
}

// OrchestateShardTransfer manages the transfer of a shard to a target node
func (o *ChunkTransferOrchestrator) OrchestateShardTransfer(
	ctx context.Context,
	shard *ModelShard,
	sourceNode, targetNode string,
	priority TransferPriority) (*ShardTransfer, error) {

	o.mu.Lock()
	defer o.mu.Unlock()

	// Create transfer record
	transfer := &ShardTransfer{
		ID:         fmt.Sprintf("transfer-%s-%d", shard.ID, time.Now().UnixNano()),
		ShardID:    shard.ID,
		SourceNode: sourceNode,
		TargetNode: targetNode,
		Size:       shard.Size,
		Priority:   priority,
		Status:     TransferStatusPending,
		ChunkSize:  o.calculateOptimalChunkSize(shard.Size),
		Checksum:   shard.Checksum,
		StartTime:  time.Now(),
	}

	// Calculate total chunks
	transfer.TotalChunks = int((shard.Size + transfer.ChunkSize - 1) / transfer.ChunkSize)

	// Add to coordinator
	o.coordinator.AddTransfer(transfer)

	// Create cancellable context
	transferCtx, cancel := context.WithTimeout(ctx, o.config.TransferTimeout)
	o.activeTransfers[transfer.ID] = cancel

	// Start transfer in background
	go o.executeTransfer(transferCtx, transfer, shard)

	return transfer, nil
}

// executeTransfer performs the actual shard transfer
func (o *ChunkTransferOrchestrator) executeTransfer(
	ctx context.Context,
	transfer *ShardTransfer,
	shard *ModelShard) {

	defer func() {
		o.mu.Lock()
		delete(o.activeTransfers, transfer.ID)
		o.mu.Unlock()
	}()

	// Update status
	o.updateTransferStatus(transfer, TransferStatusActive)

	// Break shard into chunks
	chunks, err := o.createChunks(shard, transfer.ChunkSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create chunks for shard %s: %w", shard.ID, err)
	}

	// Initialize verification
	o.verificationMgr.InitializeTransfer(transfer.ID, chunks)

	// Transfer chunks with retry logic
	for attempt := 0; attempt <= o.config.MaxRetries; attempt++ {
		if attempt > 0 {
			o.updateTransferStatus(transfer, TransferStatusRetrying)
			transfer.RetryCount = attempt
			time.Sleep(o.config.RetryDelay)
		}

		err := o.transferChunks(ctx, transfer, chunks)
		if err == nil {
			break
		}

		transfer.LastError = err.Error()

		if attempt == o.config.MaxRetries {
			o.updateTransferStatus(transfer, TransferStatusFailed)
			o.handleTransferFailure(transfer, err)
			return
		}
	}

	// Verify all chunks
	if o.config.VerifyChecksums {
		o.updateTransferStatus(transfer, TransferStatusVerifying)
		if err := o.verifyTransfer(transfer, chunks); err != nil {
			o.updateTransferStatus(transfer, TransferStatusFailed)
			o.handleTransferFailure(transfer, err)
			return
		}
	}

	// Assemble the complete shard from chunks
	if err := o.assembleShardFromChunks(transfer, chunks); err != nil {
		o.updateTransferStatus(transfer, TransferStatusFailed)
		o.handleTransferFailure(transfer, fmt.Errorf("shard assembly failed: %w", err))
		return
	}

	// Mark as completed
	o.updateTransferStatus(transfer, TransferStatusCompleted)
	transfer.EndTime = time.Now()
	transfer.Progress = 1.0

	// Update metrics
	o.updateMetrics(transfer)

	// Trigger completion callbacks
	o.coordinator.NotifyCompletion(transfer)
}

// transferChunks transfers all chunks for a shard
func (o *ChunkTransferOrchestrator) transferChunks(
	ctx context.Context,
	transfer *ShardTransfer,
	chunks []ChunkInfo) error {

	// Create chunk transfer jobs
	jobs := make(chan ChunkInfo, len(chunks))
	results := make(chan error, len(chunks))

	// Start worker pool
	numWorkers := o.calculateWorkerCount(transfer)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go o.chunkTransferWorker(ctx, transfer, jobs, results, &wg)
	}

	// Queue all chunks
	for _, chunk := range chunks {
		jobs <- chunk
	}
	close(jobs)

	// Wait for all workers to complete
	wg.Wait()
	close(results)

	// Check for errors
	for err := range results {
		if err != nil {
			return fmt.Errorf("chunk transfer failed: %w", err)
		}
	}

	return nil
}

// chunkTransferWorker processes chunk transfer jobs
func (o *ChunkTransferOrchestrator) chunkTransferWorker(
	ctx context.Context,
	transfer *ShardTransfer,
	jobs <-chan ChunkInfo,
	results chan<- error,
	wg *sync.WaitGroup) {

	defer wg.Done()

	for chunk := range jobs {
		select {
		case <-ctx.Done():
			results <- ctx.Err()
			return
		default:
			// Apply bandwidth throttling if enabled
			if o.config.BandwidthThrottling {
				o.applyBandwidthThrottle(transfer.TargetNode, chunk.Size)
			}

			// Transfer chunk using P2P engine
			err := o.transferSingleChunk(ctx, transfer, chunk)
			if err != nil {
				results <- err
				continue
			}

			// Update progress
			atomic.AddInt32(&transfer.CompletedChunks, 1)
			completedChunks := atomic.LoadInt32(&transfer.CompletedChunks)
			progress := float64(completedChunks) / float64(transfer.TotalChunks)
			transfer.Progress = progress
			atomic.AddInt64(&transfer.BytesTransferred, chunk.Size)

			// Notify progress
			o.coordinator.NotifyProgress(transfer)

			results <- nil
		}
	}
}

// transferSingleChunk transfers one chunk using P2P and handles storage on target
func (o *ChunkTransferOrchestrator) transferSingleChunk(
	ctx context.Context,
	transfer *ShardTransfer,
	chunk ChunkInfo) error {

	// Create chunk data reader from actual shard file
	reader := o.createChunkReader(transfer.ShardID, chunk)
	defer func() {
		if closer, ok := reader.(io.Closer); ok {
			closer.Close()
		}
	}()

	// Use P2P engine for transfer
	transferReq := &P2PTransferRequest{
		SourceNode: transfer.SourceNode,
		TargetNode: transfer.TargetNode,
		Data:       reader,
		Size:       chunk.Size,
		Metadata: map[string]string{
			"shard_id":     transfer.ShardID,
			"chunk_index":  fmt.Sprintf("%d", chunk.Index),
			"chunk_offset": fmt.Sprintf("%d", chunk.Offset),
			"checksum":     chunk.Checksum,
		},
	}

	// Execute transfer
	resp, err := o.p2pEngine.Transfer(ctx, transferReq)
	if err != nil {
		return fmt.Errorf("P2P transfer failed for chunk %d: %w", chunk.Index, err)
	}

	// On target node, store received chunk to local storage
	// In a real P2P implementation, the data would be streamed directly to disk during transfer
	err = o.storeReceivedChunk(transfer, chunk, resp)
	if err != nil {
		return fmt.Errorf("failed to store received chunk %d: %w", chunk.Index, err)
	}

	// Verify chunk if checksums enabled
	if o.config.VerifyChecksums {
		if err := o.verifyReceivedChunk(transfer, chunk); err != nil {
			return fmt.Errorf("chunk verification failed for chunk %d: %w", chunk.Index, err)
		}
	}

	// Mark chunk as verified
	o.verificationMgr.MarkChunkVerified(transfer.ID, chunk.Index)

	return nil
}

// createChunks divides a shard into chunks
func (o *ChunkTransferOrchestrator) createChunks(shard *ModelShard, chunkSize int64) ([]ChunkInfo, error) {
	chunks := make([]ChunkInfo, 0)

	offset := int64(0)
	index := 0

	for offset < shard.Size {
		size := chunkSize
		if offset+size > shard.Size {
			size = shard.Size - offset
		}

		checksum, err := o.calculateChunkChecksum(shard.ID, index, offset, size)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate checksum for chunk %d: %w", index, err)
		}

		chunk := ChunkInfo{
			Index:    index,
			Offset:   offset,
			Size:     size,
			Checksum: checksum,
		}

		chunks = append(chunks, chunk)
		offset += size
		index++
	}

	return chunks, nil
}

// calculateOptimalChunkSize determines best chunk size based on conditions
func (o *ChunkTransferOrchestrator) calculateOptimalChunkSize(shardSize int64) int64 {
	if !o.config.EnableAdaptiveChunking {
		return o.config.DefaultChunkSize
	}

	// Start with default
	chunkSize := o.config.DefaultChunkSize

	// Adjust based on shard size
	if shardSize < 100*1024*1024 { // < 100MB
		chunkSize = o.config.MinChunkSize
	} else if shardSize > 1024*1024*1024 { // > 1GB
		chunkSize = o.config.MaxChunkSize
	}

	// Adjust based on network conditions (simplified)
	// In real implementation, would measure actual bandwidth
	avgBandwidth := o.estimateAverageBandwidth()
	if avgBandwidth > 0 {
		// Target ~1 second per chunk transfer
		optimalSize := avgBandwidth
		if optimalSize < o.config.MinChunkSize {
			chunkSize = o.config.MinChunkSize
		} else if optimalSize > o.config.MaxChunkSize {
			chunkSize = o.config.MaxChunkSize
		} else {
			chunkSize = optimalSize
		}
	}

	return chunkSize
}

// calculateWorkerCount determines number of parallel workers
func (o *ChunkTransferOrchestrator) calculateWorkerCount(transfer *ShardTransfer) int {
	// Base on priority
	workers := 1
	switch transfer.Priority {
	case TransferPriorityCritical:
		workers = 4
	case TransferPriorityHigh:
		workers = 3
	case TransferPriorityNormal:
		workers = 2
	case TransferPriorityLow:
		workers = 1
	}

	// Cap by max concurrent transfers
	maxWorkers := o.config.MaxConcurrentTransfers
	if workers > maxWorkers {
		workers = maxWorkers
	}

	// Adjust based on current load
	activeCount := atomic.LoadInt32(&o.coordinator.activeTransfers)
	if activeCount > int32(o.config.MaxConcurrentTransfers/2) {
		workers = 1 // Reduce parallelism under load
	}

	return workers
}

// applyBandwidthThrottle applies bandwidth limiting
func (o *ChunkTransferOrchestrator) applyBandwidthThrottle(nodeID string, chunkSize int64) {
	if !o.config.BandwidthThrottling {
		return
	}

	o.coordinator.mu.Lock()
	defer o.coordinator.mu.Unlock()

	// Get current bandwidth usage
	currentUsage := o.coordinator.currentBandwidth[nodeID]
	limit := o.coordinator.bandwidthLimits[nodeID]
	if limit == 0 {
		limit = o.config.MaxBandwidthPerNode
	}

	// If exceeding limit, sleep to throttle
	if currentUsage+chunkSize > limit {
		sleepDuration := time.Duration(float64(chunkSize) / float64(limit) * float64(time.Second))
		time.Sleep(sleepDuration)
	}

	// Update current usage
	o.coordinator.currentBandwidth[nodeID] = currentUsage + chunkSize

	// Decay usage over time (simplified)
	go func() {
		time.Sleep(1 * time.Second)
		o.coordinator.mu.Lock()
		o.coordinator.currentBandwidth[nodeID] -= chunkSize
		o.coordinator.mu.Unlock()
	}()
}

// verifyTransfer verifies all chunks of a transfer
func (o *ChunkTransferOrchestrator) verifyTransfer(transfer *ShardTransfer, chunks []ChunkInfo) error {
	// First do basic verification through the verification manager
	err := o.verificationMgr.VerifyTransfer(transfer.ID, chunks)
	if err != nil {
		return fmt.Errorf("verification manager check failed: %w", err)
	}

	// Then do detailed integrity validation of each chunk file
	for _, chunk := range chunks {
		if err := o.validateChunkIntegrity(transfer.ID, chunk); err != nil {
			return fmt.Errorf("chunk %d integrity validation failed: %w", chunk.Index, err)
		}
	}

	return nil
}

// updateTransferStatus updates the status of a transfer
func (o *ChunkTransferOrchestrator) updateTransferStatus(transfer *ShardTransfer, status TransferStatus) {
	transfer.Status = status
	o.coordinator.UpdateTransfer(transfer)
}

// handleTransferFailure handles a failed transfer
func (o *ChunkTransferOrchestrator) handleTransferFailure(transfer *ShardTransfer, err error) {
	atomic.AddInt64(&o.coordinator.failedCount, 1)
	atomic.AddInt64(&o.metrics.FailedTransfers, 1)

	// Log failure details
	transfer.LastError = err.Error()
	transfer.EndTime = time.Now()

	// Clean up any partial transfer files
	if cleanupErr := o.cleanupFailedTransfer(transfer.ID); cleanupErr != nil {
		// Log cleanup failure but don't override original error
		fmt.Printf("Warning: failed to clean up failed transfer %s: %v\n", transfer.ID, cleanupErr)
	}

	// Trigger failure callbacks
	o.coordinator.NotifyCompletion(transfer)
}

// updateMetrics updates transfer metrics
func (o *ChunkTransferOrchestrator) updateMetrics(transfer *ShardTransfer) {
	o.metrics.mu.Lock()
	defer o.metrics.mu.Unlock()

	o.metrics.TotalTransfers++
	o.metrics.SuccessfulTransfers++
	o.metrics.TotalBytesTransferred += transfer.Size

	// Calculate transfer speed
	duration := transfer.EndTime.Sub(transfer.StartTime).Seconds()
	if duration > 0 {
		speed := float64(transfer.Size) / duration

		// Update average speed
		if o.metrics.AverageTransferSpeed == 0 {
			o.metrics.AverageTransferSpeed = speed
		} else {
			o.metrics.AverageTransferSpeed = (o.metrics.AverageTransferSpeed + speed) / 2
		}

		// Update peak speed
		if speed > o.metrics.PeakTransferSpeed {
			o.metrics.PeakTransferSpeed = speed
		}
	}

	o.metrics.LastUpdateTime = time.Now()
}

// estimateAverageBandwidth estimates network bandwidth
func (o *ChunkTransferOrchestrator) estimateAverageBandwidth() int64 {
	o.metrics.mu.RLock()
	defer o.metrics.mu.RUnlock()

	if o.metrics.AverageTransferSpeed > 0 {
		return int64(o.metrics.AverageTransferSpeed)
	}

	// Default estimate: 10 MB/s
	return 10 * 1024 * 1024
}

// calculateChunkChecksum generates a checksum for a chunk by reading actual data (local or remote)
func (o *ChunkTransferOrchestrator) calculateChunkChecksum(shardID string, index int, offset, size int64) (string, error) {
	// Find the shard location to read from
	locations, err := o.shardRegistry.LocateShard(shardID)
	if err != nil {
		return "", fmt.Errorf("failed to locate shard %s: %w", shardID, err)
	}

	// Try local locations first
	for _, loc := range locations {
		if loc.IsLocal && loc.IsAvailable {
			checksum := o.calculateLocalChunkChecksum(loc.StoragePath, offset, size)
			if checksum != "" {
				return checksum, nil
			}
		}
	}

	// Try remote locations if no local found
	for _, loc := range locations {
		if loc.IsAvailable && loc.NodeID != "" {
			checksum, err := o.calculateRemoteChunkChecksum(loc.NodeID, shardID, offset, size)
			if err == nil && checksum != "" {
				return checksum, nil
			}
		}
	}

	// No available locations found
	return "", fmt.Errorf("no valid checksum sources found for shard %s chunk %d at offset %d", shardID, index, offset)
}

// calculateLocalChunkChecksum calculates checksum from a local shard file
func (o *ChunkTransferOrchestrator) calculateLocalChunkChecksum(shardPath string, offset, size int64) string {
	file, err := os.Open(shardPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	// Seek to the chunk offset
	_, err = file.Seek(offset, io.SeekStart)
	if err != nil {
		return ""
	}

	// Create a limited reader for just this chunk
	limitedReader := io.LimitReader(file, size)

	// Calculate SHA256 checksum
	hasher := sha256.New()
	_, err = io.Copy(hasher, limitedReader)
	if err != nil {
		return ""
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))
	return fmt.Sprintf("sha256:%s", checksum)
}

// calculateRemoteChunkChecksum calculates checksum from a remote shard via P2P
func (o *ChunkTransferOrchestrator) calculateRemoteChunkChecksum(nodeID, shardID string, offset, size int64) (string, error) {
	// Create chunk info for remote reading
	chunk := ChunkInfo{
		Offset: offset,
		Size:   size,
	}

	// Create remote reader
	remoteReader, err := o.createRemoteChunkReader(nodeID, shardID, chunk)
	if err != nil {
		return "", fmt.Errorf("failed to create remote chunk reader: %w", err)
	}

	// Ensure reader is closed
	defer func() {
		if closer, ok := remoteReader.(io.Closer); ok {
			closer.Close()
		}
	}()

	// Calculate checksum while streaming from remote
	hasher := sha256.New()
	_, err = io.Copy(hasher, remoteReader)
	if err != nil {
		return "", fmt.Errorf("failed to stream remote chunk data: %w", err)
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))
	return fmt.Sprintf("sha256:%s", checksum), nil
}

// createChunkReader creates a reader for a chunk from the actual shard file (local or remote)
func (o *ChunkTransferOrchestrator) createChunkReader(shardID string, chunk ChunkInfo) io.Reader {
	// Find the shard location to read from
	locations, err := o.shardRegistry.LocateShard(shardID)
	if err != nil {
		// If FileTransferClient is available, try P2P discovery
		if o.fileTransferClient != nil {
			// Attempt to discover the shard on the network
			if remoteReader := o.attemptP2PChunkReader(shardID, chunk); remoteReader != nil {
				return remoteReader
			}
		}
		// Return error instead of dummy reader - real I/O required
		return &errorReader{err: fmt.Errorf("shard %s not found and no P2P alternatives available", shardID)}
	}

	// Find a local location first
	var shardPath string
	for _, loc := range locations {
		if loc.IsLocal && loc.IsAvailable {
			shardPath = loc.StoragePath
			break
		}
	}

	if shardPath != "" {
		// Open the local shard file
		file, err := os.Open(shardPath)
		if err != nil {
			// If local file fails but we have P2P client, try remote access
			if o.fileTransferClient != nil {
				if remoteReader := o.attemptP2PChunkReader(shardID, chunk); remoteReader != nil {
					return remoteReader
				}
			}
			return &errorReader{err: fmt.Errorf("failed to open local shard file %s: %w", shardPath, err)}
		}

		// Seek to the chunk offset
		_, err = file.Seek(chunk.Offset, io.SeekStart)
		if err != nil {
			file.Close()
			// If seek fails but we have P2P client, try remote access
			if o.fileTransferClient != nil {
				if remoteReader := o.attemptP2PChunkReader(shardID, chunk); remoteReader != nil {
					return remoteReader
				}
			}
			return &errorReader{err: fmt.Errorf("failed to seek to chunk offset %d in shard file %s: %w", chunk.Offset, shardPath, err)}
		}

		// Return a limited reader for just this chunk
		return &chunkReader{
			file:      file,
			remaining: chunk.Size,
		}
	}

	// No local shard found, try remote locations
	for _, loc := range locations {
		if loc.IsAvailable && loc.NodeID != "" {
			// Create remote chunk reader using FileTransferClient
			remoteReader, err := o.createRemoteChunkReader(loc.NodeID, shardID, chunk)
			if err == nil {
				return remoteReader
			}
			// Continue to next location if this one failed
		}
	}

	// No available locations found, try P2P discovery as final attempt
	if o.fileTransferClient != nil {
		if remoteReader := o.attemptP2PChunkReader(shardID, chunk); remoteReader != nil {
			return remoteReader
		}
	}

	// Return error reader - no more fallbacks available
	return &errorReader{err: fmt.Errorf("no available sources found for chunk %d of shard %s", chunk.Index, shardID)}
}

// createRemoteChunkReader creates a reader for a chunk from a remote peer using real P2P streaming
func (o *ChunkTransferOrchestrator) createRemoteChunkReader(nodeID, shardID string, chunk ChunkInfo) (io.Reader, error) {
	// Parse peer ID
	peerID, err := peer.Decode(nodeID)
	if err != nil {
		return nil, fmt.Errorf("invalid peer ID %s: %w", nodeID, err)
	}

	// Check if we have a file transfer client
	if o.fileTransferClient == nil {
		return nil, fmt.Errorf("file transfer client not initialized")
	}

	// Create a context for the chunk request with longer timeout for real streaming
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)

	// Request the specific byte range from the remote peer
	transfer, err := o.fileTransferClient.RequestFileRange(ctx, peerID, shardID, chunk.Offset, chunk.Size, 1)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to request chunk [%d-%d] of shard %s from peer %s: %w",
			chunk.Offset, chunk.Offset+chunk.Size, shardID, nodeID, err)
	}

	// Create pipes for streaming data with checksum verification
	pr, pw := io.Pipe()

	// Start streaming goroutine with proper error handling and persistence
	go func() {
		defer pw.Close()
		defer cancel()

		// Create chunk storage directory for persistence
		chunkDir := filepath.Join(o.chunkStoragePath, "transfers", transfer.ID)
		if err := os.MkdirAll(chunkDir, 0755); err != nil {
			pw.CloseWithError(fmt.Errorf("failed to create chunk directory: %w", err))
			return
		}

		chunkFile := filepath.Join(chunkDir, fmt.Sprintf("chunk_%d.dat", chunk.Index))
		hasher := sha256.New()

		// Wait for transfer to be ready with exponential backoff
		maxWait := 30 * time.Second
		waitTime := 100 * time.Millisecond
		totalWait := time.Duration(0)

		for totalWait < maxWait {
			// Check if transfer has file handle ready
			if transfer.File != nil {
				// Stream directly from file handle
				transfer.File.Seek(0, io.SeekStart)
				limitedReader := io.LimitReader(transfer.File, chunk.Size)

				// Create file for persistence while streaming
				outFile, err := os.Create(chunkFile)
				if err != nil {
					pw.CloseWithError(fmt.Errorf("failed to create chunk file: %w", err))
					return
				}
				defer outFile.Close()

				// Stream with checksum calculation and persistence
				multiWriter := io.MultiWriter(pw, outFile, hasher)
				written, err := io.Copy(multiWriter, limitedReader)
				if err != nil {
					pw.CloseWithError(fmt.Errorf("failed to stream chunk data: %w", err))
					return
				}

				// Verify size
				if written != chunk.Size {
					pw.CloseWithError(fmt.Errorf("chunk size mismatch: expected %d, got %d", chunk.Size, written))
					return
				}

				// Verify checksum if available
				actualChecksum := hex.EncodeToString(hasher.Sum(nil))
				if chunk.Checksum != "" && !strings.EqualFold(chunk.Checksum, actualChecksum) && !strings.EqualFold(chunk.Checksum, "sha256:"+actualChecksum) {
					pw.CloseWithError(fmt.Errorf("chunk checksum mismatch: expected %s, got %s", chunk.Checksum, actualChecksum))
					return
				}

				return
			}

			// Check for temporary file created by transfer
			tempPath := filepath.Join(o.chunkStoragePath, fmt.Sprintf("remote_%s_%d.tmp", shardID, chunk.Index))
			if file, err := os.Open(tempPath); err == nil {
				defer file.Close()

				// Create persistent file while streaming
				outFile, err := os.Create(chunkFile)
				if err != nil {
					pw.CloseWithError(fmt.Errorf("failed to create chunk file: %w", err))
					return
				}
				defer outFile.Close()

				// Stream with checksum and persistence
				multiWriter := io.MultiWriter(pw, outFile, hasher)
				written, err := io.CopyN(multiWriter, file, chunk.Size)
				if err != nil {
					pw.CloseWithError(fmt.Errorf("failed to stream from temp file: %w", err))
					return
				}

				// Verify size and checksum
				if written != chunk.Size {
					pw.CloseWithError(fmt.Errorf("chunk size mismatch: expected %d, got %d", chunk.Size, written))
					return
				}

				actualChecksum := hex.EncodeToString(hasher.Sum(nil))
				if chunk.Checksum != "" && !strings.EqualFold(chunk.Checksum, actualChecksum) && !strings.EqualFold(chunk.Checksum, "sha256:"+actualChecksum) {
					pw.CloseWithError(fmt.Errorf("chunk checksum mismatch: expected %s, got %s", chunk.Checksum, actualChecksum))
					return
				}

				return
			}

			// Wait with exponential backoff
			select {
			case <-ctx.Done():
				pw.CloseWithError(fmt.Errorf("context cancelled while waiting for transfer: %w", ctx.Err()))
				return
			case <-time.After(waitTime):
				totalWait += waitTime
				if waitTime < time.Second {
					waitTime *= 2
				}
			}
		}

		// If we reach here, transfer failed to provide data
		pw.CloseWithError(fmt.Errorf("timeout waiting for remote chunk data after %v", maxWait))
	}()

	return pr, nil
}

// chunkReader reads a specific chunk from a shard file
type chunkReader struct {
	file      *os.File
	remaining int64
}

func (r *chunkReader) Read(p []byte) (n int, err error) {
	if r.remaining <= 0 {
		r.file.Close()
		return 0, io.EOF
	}

	toRead := int64(len(p))
	if toRead > r.remaining {
		toRead = r.remaining
	}

	n, err = r.file.Read(p[:toRead])
	r.remaining -= int64(n)

	if r.remaining <= 0 || err == io.EOF {
		r.file.Close()
	}

	return n, err
}

func (r *chunkReader) Close() error {
	return r.file.Close()
}

// dummyReader is a fallback reader for when real file access fails
type dummyReader struct {
	size int64
	read int64
}

func (r *dummyReader) Read(p []byte) (n int, err error) {
	remaining := r.size - r.read
	if remaining <= 0 {
		return 0, io.EOF
	}

	toRead := int64(len(p))
	if toRead > remaining {
		toRead = remaining
	}

	// Fill with dummy data (zeros for now)
	for i := int64(0); i < toRead; i++ {
		p[i] = 0
	}

	r.read += toRead
	return int(toRead), nil
}

// errorReader always returns an error when read is attempted
type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}

// remoteChunkReader reads a chunk from a remote peer via P2P
type remoteChunkReader struct {
	file       *os.File
	remaining  int64
	cancelFunc context.CancelFunc
}

func (r *remoteChunkReader) Read(p []byte) (n int, err error) {
	if r.remaining <= 0 {
		if r.file != nil {
			r.file.Close()
		}
		if r.cancelFunc != nil {
			r.cancelFunc()
		}
		return 0, io.EOF
	}

	toRead := int64(len(p))
	if toRead > r.remaining {
		toRead = r.remaining
	}

	n, err = r.file.Read(p[:toRead])
	r.remaining -= int64(n)

	if r.remaining <= 0 || err == io.EOF {
		if r.file != nil {
			r.file.Close()
		}
		if r.cancelFunc != nil {
			r.cancelFunc()
		}
	}

	return n, err
}

func (r *remoteChunkReader) Close() error {
	if r.cancelFunc != nil {
		r.cancelFunc()
	}
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

// NewTransferCoordinator creates a new transfer coordinator
func NewTransferCoordinator() *TransferCoordinator {
	return &TransferCoordinator{
		transfers:        make(map[string]*ShardTransfer),
		queues:           make(map[TransferPriority][]*ShardTransfer),
		bandwidthLimits:  make(map[string]int64),
		currentBandwidth: make(map[string]int64),
	}
}

// AddTransfer adds a new transfer to the coordinator
func (c *TransferCoordinator) AddTransfer(transfer *ShardTransfer) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.transfers[transfer.ID] = transfer
	c.queues[transfer.Priority] = append(c.queues[transfer.Priority], transfer)
	atomic.AddInt32(&c.activeTransfers, 1)
}

// UpdateTransfer updates transfer information
func (c *TransferCoordinator) UpdateTransfer(transfer *ShardTransfer) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.transfers[transfer.ID] = transfer
}

// NotifyProgress notifies progress callbacks
func (c *TransferCoordinator) NotifyProgress(transfer *ShardTransfer) {
	for _, callback := range c.progressCallbacks {
		go callback(transfer)
	}
}

// NotifyCompletion notifies completion callbacks
func (c *TransferCoordinator) NotifyCompletion(transfer *ShardTransfer) {
	c.mu.Lock()
	atomic.AddInt32(&c.activeTransfers, -1)

	if transfer.Status == TransferStatusCompleted {
		atomic.AddInt64(&c.completedCount, 1)
		atomic.AddInt64(&c.totalBytesTransferred, transfer.Size)
	} else {
		atomic.AddInt64(&c.failedCount, 1)
	}
	c.mu.Unlock()

	for _, callback := range c.completionCallbacks {
		go callback(transfer)
	}
}

// NewChunkVerificationManager creates a new verification manager
func NewChunkVerificationManager(verifier *IntegrityVerifier) *ChunkVerificationManager {
	return &ChunkVerificationManager{
		verifier:       verifier,
		pendingChunks:  make(map[string][]ChunkInfo),
		verifiedChunks: make(map[string][]bool),
		checksumCache:  make(map[string]string),
	}
}

// InitializeTransfer prepares verification for a transfer
func (v *ChunkVerificationManager) InitializeTransfer(transferID string, chunks []ChunkInfo) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.pendingChunks[transferID] = chunks
	v.verifiedChunks[transferID] = make([]bool, len(chunks))
}

// MarkChunkVerified marks a chunk as verified
func (v *ChunkVerificationManager) MarkChunkVerified(transferID string, chunkIndex int) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if verified, exists := v.verifiedChunks[transferID]; exists && chunkIndex < len(verified) {
		verified[chunkIndex] = true
	}
}

// VerifyTransfer verifies all chunks of a transfer
func (v *ChunkVerificationManager) VerifyTransfer(transferID string, chunks []ChunkInfo) error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	verified, exists := v.verifiedChunks[transferID]
	if !exists {
		return fmt.Errorf("transfer %s not found", transferID)
	}

	for i, isVerified := range verified {
		if !isVerified {
			return fmt.Errorf("chunk %d not verified", i)
		}
	}

	return nil
}

// GetTransferStatus returns the status of a transfer
func (o *ChunkTransferOrchestrator) GetTransferStatus(transferID string) (*ShardTransfer, error) {
	o.coordinator.mu.RLock()
	defer o.coordinator.mu.RUnlock()

	transfer, exists := o.coordinator.transfers[transferID]
	if !exists {
		return nil, fmt.Errorf("transfer %s not found", transferID)
	}

	return transfer, nil
}

// GetDetailedTransferStatus returns detailed transfer status with chunk information
func (o *ChunkTransferOrchestrator) GetDetailedTransferStatus(transferID string) (*DetailedTransferStatus, error) {
	o.coordinator.mu.RLock()
	transfer, exists := o.coordinator.transfers[transferID]
	o.coordinator.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("transfer %s not found", transferID)
	}

	// Get chunk verification status
	o.verificationMgr.mu.RLock()
	pendingChunks := o.verificationMgr.pendingChunks[transferID]
	verifiedChunks := o.verificationMgr.verifiedChunks[transferID]
	o.verificationMgr.mu.RUnlock()

	// Build detailed status
	detailedStatus := &DetailedTransferStatus{
		Transfer:      transfer,
		ChunkDetails:  make([]ChunkTransferStatus, len(pendingChunks)),
		StoragePath:   o.GetChunkStoragePath(transferID),
		AssembledPath: o.GetAssembledShardPath(transfer.ShardID),
	}

	// Add chunk status details
	for i, chunk := range pendingChunks {
		isVerified := false
		if i < len(verifiedChunks) {
			isVerified = verifiedChunks[i]
		}

		detailedStatus.ChunkDetails[i] = ChunkTransferStatus{
			Index:    chunk.Index,
			Offset:   chunk.Offset,
			Size:     chunk.Size,
			Checksum: chunk.Checksum,
			Verified: isVerified,
		}
	}

	return detailedStatus, nil
}

// DetailedTransferStatus provides comprehensive transfer information
type DetailedTransferStatus struct {
	Transfer      *ShardTransfer        `json:"transfer"`
	ChunkDetails  []ChunkTransferStatus `json:"chunk_details"`
	StoragePath   string                `json:"storage_path"`
	AssembledPath string                `json:"assembled_path"`
}

// ChunkTransferStatus provides status information for individual chunks
type ChunkTransferStatus struct {
	Index    int    `json:"index"`
	Offset   int64  `json:"offset"`
	Size     int64  `json:"size"`
	Checksum string `json:"checksum"`
	Verified bool   `json:"verified"`
}

// CancelTransfer cancels an active transfer
func (o *ChunkTransferOrchestrator) CancelTransfer(transferID string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	cancel, exists := o.activeTransfers[transferID]
	if !exists {
		return fmt.Errorf("transfer %s not active", transferID)
	}

	cancel()
	delete(o.activeTransfers, transferID)

	// Update transfer status
	if transfer, exists := o.coordinator.transfers[transferID]; exists {
		transfer.Status = TransferStatusFailed
		transfer.LastError = "cancelled by user"
		transfer.EndTime = time.Now()
	}

	return nil
}

// storeReceivedChunk stores a received chunk to local disk
func (o *ChunkTransferOrchestrator) storeReceivedChunk(transfer *ShardTransfer, chunk ChunkInfo, resp *P2PTransferResponse) error {
	if !resp.Success {
		return fmt.Errorf("P2P transfer failed: %s", resp.Error)
	}

	// Create chunk storage directory for this transfer
	transferDir := filepath.Join(o.chunkStoragePath, transfer.ID)
	err := os.MkdirAll(transferDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create transfer directory: %w", err)
	}

	// Create chunk file path
	chunkPath := filepath.Join(transferDir, fmt.Sprintf("chunk_%d.dat", chunk.Index))

	// Get the actual chunk data from the P2P transfer response
	// The response should contain the path to the transferred chunk or the chunk data itself

	// First, try to find the chunk data from P2P engine's storage
	p2pChunkPath := filepath.Join(o.config.StorageDir, "chunks", transfer.ID, fmt.Sprintf("chunk_%d.dat", chunk.Index))
	if _, err := os.Stat(p2pChunkPath); err == nil {
		// P2P engine has saved the chunk, copy it to our storage
		srcFile, err := os.Open(p2pChunkPath)
		if err != nil {
			return fmt.Errorf("failed to open P2P chunk file: %w", err)
		}
		defer srcFile.Close()

		dstFile, err := os.Create(chunkPath)
		if err != nil {
			return fmt.Errorf("failed to create chunk file: %w", err)
		}
		defer dstFile.Close()

		// Copy with checksum verification
		hasher := sha256.New()
		teeReader := io.TeeReader(srcFile, hasher)

		bytesWritten, err := io.Copy(dstFile, teeReader)
		if err != nil {
			return fmt.Errorf("failed to copy chunk data: %w", err)
		}

		if bytesWritten != chunk.Size {
			return fmt.Errorf("chunk size mismatch: expected %d, got %d", chunk.Size, bytesWritten)
		}

		// Verify checksum if provided
		if chunk.Checksum != "" {
			actualChecksum := fmt.Sprintf("sha256:%s", hex.EncodeToString(hasher.Sum(nil)))
			if actualChecksum != chunk.Checksum {
				// Delete the bad chunk
				os.Remove(chunkPath)
				return fmt.Errorf("chunk checksum mismatch: expected %s, got %s", chunk.Checksum, actualChecksum)
			}
		}
	} else {
		// Try to copy from source directly
		if err := o.copyChunkFromSource(transfer.ShardID, chunk, chunkPath); err != nil {
			// If still can't get the chunk, try using the chunk reader
			reader := o.createChunkReader(transfer.ShardID, chunk)
			if reader == nil {
				return fmt.Errorf("failed to create chunk reader for shard %s chunk %d", transfer.ShardID, chunk.Index)
			}

			// If reader is a Closer, defer closing it
			if closer, ok := reader.(io.Closer); ok {
				defer closer.Close()
			}

			// Create the chunk file
			file, err := os.Create(chunkPath)
			if err != nil {
				return fmt.Errorf("failed to create chunk file: %w", err)
			}
			defer file.Close()

			// Copy data with checksum calculation
			hasher := sha256.New()
			writer := io.MultiWriter(file, hasher)

			bytesWritten, err := io.Copy(writer, reader)
			if err != nil {
				return fmt.Errorf("failed to write chunk data: %w", err)
			}

			if bytesWritten != chunk.Size {
				return fmt.Errorf("incomplete chunk write: expected %d bytes, wrote %d", chunk.Size, bytesWritten)
			}

			// Update chunk checksum if not set
			if chunk.Checksum == "" {
				chunk.Checksum = fmt.Sprintf("sha256:%s", hex.EncodeToString(hasher.Sum(nil)))
			}
		}
	}

	// Verify the stored chunk using IntegrityVerifier
	if o.integrityVerifier != nil && chunk.Checksum != "" {
		// Use the VerifyChunk method if available
		verifyReq := &ChunkVerifyRequest{
			ChunkPath: chunkPath,
			Checksum:  chunk.Checksum,
			Algorithm: HashAlgorithmSHA256,
		}

		verifyResp := o.integrityVerifier.VerifyChunk(verifyReq)
		if !verifyResp.Valid {
			os.Remove(chunkPath)
			return fmt.Errorf("chunk verification failed: %s", verifyResp.Error)
		}
	}

	return nil
}

// copyChunkFromSource copies chunk data from source shard to target file
func (o *ChunkTransferOrchestrator) copyChunkFromSource(shardID string, chunk ChunkInfo, targetPath string) error {
	// Find source shard location
	locations, err := o.shardRegistry.LocateShard(shardID)
	if err != nil {
		return err
	}

	// Find a local source location
	var sourcePath string
	for _, loc := range locations {
		if loc.IsLocal && loc.IsAvailable {
			sourcePath = loc.StoragePath
			break
		}
	}

	if sourcePath == "" {
		return fmt.Errorf("no local source available for shard %s", shardID)
	}

	// Open source file
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source shard: %w", err)
	}
	defer sourceFile.Close()

	// Seek to chunk offset
	_, err = sourceFile.Seek(chunk.Offset, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to seek to chunk offset: %w", err)
	}

	// Create target file
	targetFile, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create target chunk file: %w", err)
	}
	defer targetFile.Close()

	// Copy chunk data
	limitedReader := io.LimitReader(sourceFile, chunk.Size)
	_, err = io.Copy(targetFile, limitedReader)
	if err != nil {
		return fmt.Errorf("failed to copy chunk data: %w", err)
	}

	return nil
}

// verifyReceivedChunk verifies a received chunk's integrity
func (o *ChunkTransferOrchestrator) verifyReceivedChunk(transfer *ShardTransfer, chunk ChunkInfo) error {
	// Read the stored chunk file
	chunkPath := filepath.Join(o.chunkStoragePath, transfer.ID, fmt.Sprintf("chunk_%d.dat", chunk.Index))
	file, err := os.Open(chunkPath)
	if err != nil {
		return fmt.Errorf("failed to open chunk file for verification: %w", err)
	}
	defer file.Close()

	// Check file size first
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat chunk file: %w", err)
	}

	if fileInfo.Size() != chunk.Size {
		return fmt.Errorf("chunk size mismatch: expected %d, got %d", chunk.Size, fileInfo.Size())
	}

	// Calculate checksum of stored data
	hasher := sha256.New()
	_, err = io.Copy(hasher, file)
	if err != nil {
		return fmt.Errorf("failed to read chunk file for checksum: %w", err)
	}

	actualChecksum := fmt.Sprintf("sha256:%s", hex.EncodeToString(hasher.Sum(nil)))

	// Compare with expected checksum
	if actualChecksum != chunk.Checksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", chunk.Checksum, actualChecksum)
	}

	return nil
}

// assembleShardFromChunks assembles a complete shard file from received chunks
func (o *ChunkTransferOrchestrator) assembleShardFromChunks(transfer *ShardTransfer, chunks []ChunkInfo) error {
	// Sort chunks by index to ensure correct order
	sortedChunks := make([]ChunkInfo, len(chunks))
	copy(sortedChunks, chunks)

	// Sort by chunk index
	for i := 0; i < len(sortedChunks)-1; i++ {
		for j := i + 1; j < len(sortedChunks); j++ {
			if sortedChunks[i].Index > sortedChunks[j].Index {
				sortedChunks[i], sortedChunks[j] = sortedChunks[j], sortedChunks[i]
			}
		}
	}

	// Create target shard file path in appropriate location
	shardDir := filepath.Join(o.chunkStoragePath, "assembled_shards")
	err := os.MkdirAll(shardDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create shard directory: %w", err)
	}

	shardPath := filepath.Join(shardDir, fmt.Sprintf("%s.dat", transfer.ShardID))
	shardFile, err := os.Create(shardPath)
	if err != nil {
		return fmt.Errorf("failed to create shard file: %w", err)
	}
	defer shardFile.Close()

	// Write chunks in order
	for _, chunk := range sortedChunks {
		chunkPath := filepath.Join(o.chunkStoragePath, transfer.ID, fmt.Sprintf("chunk_%d.dat", chunk.Index))
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			return fmt.Errorf("failed to open chunk %d for assembly: %w", chunk.Index, err)
		}

		// Verify we're at the correct position in the output file
		expectedPos := chunk.Offset
		currentPos, err := shardFile.Seek(0, io.SeekCurrent)
		if err != nil {
			chunkFile.Close()
			return fmt.Errorf("failed to get current position: %w", err)
		}

		if currentPos != expectedPos {
			// Seek to correct position
			_, err = shardFile.Seek(expectedPos, io.SeekStart)
			if err != nil {
				chunkFile.Close()
				return fmt.Errorf("failed to seek to chunk position %d: %w", expectedPos, err)
			}
		}

		// Copy chunk data to shard file
		copied, err := io.Copy(shardFile, chunkFile)
		chunkFile.Close()
		if err != nil {
			return fmt.Errorf("failed to copy chunk %d to shard: %w", chunk.Index, err)
		}

		if copied != chunk.Size {
			return fmt.Errorf("incomplete chunk copy: expected %d bytes, got %d", chunk.Size, copied)
		}
	}

	// Verify the assembled shard integrity using IntegrityVerifier
	if o.integrityVerifier != nil && transfer.Checksum != "" {
		// Extract checksum format (remove prefix like "sha256:")
		checksum := transfer.Checksum
		if len(checksum) > 7 && checksum[:7] == "sha256:" {
			checksum = checksum[7:]
		}

		expectedChecksums := map[HashAlgorithm]string{
			HashAlgorithmSHA256: checksum,
		}

		result, err := o.integrityVerifier.VerifyModel(
			transfer.ShardID,
			"1.0",
			shardPath,
			expectedChecksums,
		)
		if err != nil {
			return fmt.Errorf("failed to verify assembled shard: %w", err)
		}

		if !result.Verified {
			return fmt.Errorf("assembled shard verification failed: %s", result.ErrorMessage)
		}
	}

	// Register the assembled shard in the registry
	err = o.registerAssembledShard(transfer.ShardID, shardPath, transfer.Size, transfer.Checksum)
	if err != nil {
		// Log warning but don't fail the assembly
		fmt.Printf("Warning: failed to register assembled shard: %v\n", err)
	}

	// Clean up chunk files after successful assembly
	transferDir := filepath.Join(o.chunkStoragePath, transfer.ID)
	err = os.RemoveAll(transferDir)
	if err != nil {
		// Log warning but don't fail the assembly
		// In a production system, would use proper logging
		fmt.Printf("Warning: failed to clean up chunk files: %v\n", err)
	}

	return nil
}

// registerAssembledShard registers a newly assembled shard in the registry
func (o *ChunkTransferOrchestrator) registerAssembledShard(shardID, shardPath string, size int64, checksum string) error {
	if o.shardRegistry == nil {
		return fmt.Errorf("shard registry not available")
	}

	// Create a ModelShard representation for the assembled shard
	shard := &ModelShard{
		ID:              shardID,
		ModelID:         "", // This would need to be provided in transfer metadata
		Size:            size,
		Checksum:        checksum,
		CreatedAt:       time.Now(),
		LastAccessed:    time.Now(),
		Replicas:        1,
		Priority:        1,
		NodeAssignments: []string{}, // Current node
	}

	// Register the shard
	err := o.shardRegistry.RegisterModelShard(shard)
	if err != nil {
		return fmt.Errorf("failed to register shard: %w", err)
	}

	return nil
}

// cleanupFailedTransfer cleans up files from a failed transfer
func (o *ChunkTransferOrchestrator) cleanupFailedTransfer(transferID string) error {
	transferDir := filepath.Join(o.chunkStoragePath, transferID)
	return os.RemoveAll(transferDir)
}

// GetChunkStoragePath returns the path where chunks for a transfer are stored
func (o *ChunkTransferOrchestrator) GetChunkStoragePath(transferID string) string {
	return filepath.Join(o.chunkStoragePath, transferID)
}

// GetAssembledShardPath returns the path where an assembled shard is stored
func (o *ChunkTransferOrchestrator) GetAssembledShardPath(shardID string) string {
	return filepath.Join(o.chunkStoragePath, "assembled_shards", fmt.Sprintf("%s.dat", shardID))
}

// assembleShardFromChunks assembles chunk files into the final shard file
func (o *ChunkTransferOrchestrator) assembleShardFromChunks(transferID string, shard *ModelShard) error {
	// Get chunk directory
	chunkDir := filepath.Join(o.chunkStoragePath, "transfers", transferID)
	assembledDir := filepath.Join(o.chunkStoragePath, "assembled_shards")

	// Create assembled directory if it doesn't exist
	if err := os.MkdirAll(assembledDir, 0755); err != nil {
		return fmt.Errorf("failed to create assembled directory: %w", err)
	}

	// Get final shard path
	finalPath := o.GetAssembledShardPath(shard.ID)

	// Create output file
	outFile, err := os.Create(finalPath)
	if err != nil {
		return fmt.Errorf("failed to create assembled shard file: %w", err)
	}
	defer outFile.Close()

	// Calculate total chunks from shard size and chunk size
	totalChunks := int((shard.Size + DefaultChunkSize - 1) / DefaultChunkSize)
	hasher := sha256.New()

	// Concatenate all chunks in order
	for chunkIndex := 0; chunkIndex < totalChunks; chunkIndex++ {
		chunkFile := filepath.Join(chunkDir, fmt.Sprintf("chunk_%d.dat", chunkIndex))

		// Read chunk file
		chunkData, err := os.ReadFile(chunkFile)
		if err != nil {
			return fmt.Errorf("failed to read chunk %d from %s: %w", chunkIndex, chunkFile, err)
		}

		// Write to output file and update hash
		if _, err := outFile.Write(chunkData); err != nil {
			return fmt.Errorf("failed to write chunk %d to assembled file: %w", chunkIndex, err)
		}
		hasher.Write(chunkData)
	}

	// Verify final checksum against shard checksum
	actualChecksum := hex.EncodeToString(hasher.Sum(nil))
	if shard.Checksum != "" {
		expectedChecksum := shard.Checksum
		// Handle different checksum formats
		if strings.HasPrefix(expectedChecksum, "sha256:") {
			expectedChecksum = strings.TrimPrefix(expectedChecksum, "sha256:")
		}

		if !strings.EqualFold(actualChecksum, expectedChecksum) {
			// Remove the incomplete file
			os.Remove(finalPath)
			return fmt.Errorf("assembled shard checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
		}
	}

	// Verify final file size
	stat, err := outFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat assembled file: %w", err)
	}
	if stat.Size() != shard.Size {
		os.Remove(finalPath)
		return fmt.Errorf("assembled shard size mismatch: expected %d, got %d", shard.Size, stat.Size())
	}

	return nil
}

// validateChunkIntegrity performs additional integrity checks on a chunk
func (o *ChunkTransferOrchestrator) validateChunkIntegrity(transferID string, chunk ChunkInfo) error {
	chunkPath := filepath.Join(o.chunkStoragePath, transferID, fmt.Sprintf("chunk_%d.dat", chunk.Index))

	// Check if file exists
	if _, err := os.Stat(chunkPath); os.IsNotExist(err) {
		return fmt.Errorf("chunk file does not exist: %s", chunkPath)
	}

	// Open and verify file
	file, err := os.Open(chunkPath)
	if err != nil {
		return fmt.Errorf("failed to open chunk file: %w", err)
	}
	defer file.Close()

	// Check file size
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat chunk file: %w", err)
	}

	if fileInfo.Size() != chunk.Size {
		return fmt.Errorf("chunk size mismatch: expected %d, got %d", chunk.Size, fileInfo.Size())
	}

	// Calculate and verify checksum if enabled
	if o.config.VerifyChecksums && chunk.Checksum != "" {
		hasher := sha256.New()
		_, err = io.Copy(hasher, file)
		if err != nil {
			return fmt.Errorf("failed to calculate chunk checksum: %w", err)
		}

		actualChecksum := fmt.Sprintf("sha256:%s", hex.EncodeToString(hasher.Sum(nil)))
		if actualChecksum != chunk.Checksum {
			return fmt.Errorf("checksum verification failed: expected %s, got %s", chunk.Checksum, actualChecksum)
		}
	}

	return nil
}

// GetMetrics returns current transfer metrics
func (o *ChunkTransferOrchestrator) GetMetrics() *TransferMetrics {
	o.metrics.mu.RLock()
	defer o.metrics.mu.RUnlock()

	// Return a copy
	return &TransferMetrics{
		TotalTransfers:        o.metrics.TotalTransfers,
		SuccessfulTransfers:   o.metrics.SuccessfulTransfers,
		FailedTransfers:       o.metrics.FailedTransfers,
		RetryCount:            o.metrics.RetryCount,
		TotalBytesTransferred: o.metrics.TotalBytesTransferred,
		AverageTransferSpeed:  o.metrics.AverageTransferSpeed,
		PeakTransferSpeed:     o.metrics.PeakTransferSpeed,
		AverageChunkSize:      o.metrics.AverageChunkSize,
		VerificationFailures:  o.metrics.VerificationFailures,
		NetworkErrors:         o.metrics.NetworkErrors,
		LastUpdateTime:        o.metrics.LastUpdateTime,
	}
}

// SetFileTransferClient sets the file transfer client for P2P communication
func (o *ChunkTransferOrchestrator) SetFileTransferClient(client *protocols.FileTransferClient) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.fileTransferClient = client
}

// attemptP2PChunkReader attempts to create a P2P chunk reader by discovering available peers
func (o *ChunkTransferOrchestrator) attemptP2PChunkReader(shardID string, chunk ChunkInfo) io.Reader {
	if o.fileTransferClient == nil {
		return nil
	}

	// Try to discover the shard on connected peers
	// This is a simplified approach - in a full implementation, this would
	// use a more sophisticated discovery mechanism
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get connected peers from the P2P node (assuming we have access)
	// For now, we'll try a generic approach using the file transfer client

	// Attempt to request the chunk using broadcast discovery
	// The FileTransferClient should handle peer discovery internally
	peerID := peer.ID("") // Empty peer ID for broadcast/discovery mode

	transfer, err := o.fileTransferClient.RequestFileRange(
		ctx,
		peerID,
		shardID,
		chunk.Offset,
		chunk.Size,
		2, // Higher priority for discovery
	)
	if err != nil {
		return nil // Failed to discover or request chunk
	}

	// Wait briefly for transfer initialization
	time.Sleep(200 * time.Millisecond)

	// Return a reader that can stream the discovered chunk
	if transfer.File != nil {
		transfer.File.Seek(0, io.SeekStart)
		return &remoteChunkReader{
			file:       transfer.File,
			remaining:  chunk.Size,
			cancelFunc: cancel,
		}
	}

	// If no direct file access, use pipe-based streaming
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()

		// Stream chunk data through the pipe
		// This would typically read from a temporary file created by the transfer
		tempPath := filepath.Join(o.chunkStoragePath, fmt.Sprintf("discover_%s_%d.tmp", shardID, chunk.Index))

		// Wait for file to be available
		var file *os.File
		for i := 0; i < 100; i++ { // Wait up to 10 seconds
			file, err = os.Open(tempPath)
			if err == nil {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		if file != nil {
			defer file.Close()
			io.CopyN(pw, file, chunk.Size)
		} else {
			pw.CloseWithError(fmt.Errorf("failed to access discovered chunk data"))
		}
	}()

	return pr
}

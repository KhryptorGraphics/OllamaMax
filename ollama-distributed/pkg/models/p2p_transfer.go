package models

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	// Protocol IDs for P2P model transfer
	ModelTransferProtocol     = "/ollama/model-transfer/1.0.0"
	ModelChunkProtocol        = "/ollama/model-chunk/1.0.0"
	ModelMetadataProtocol     = "/ollama/model-metadata/1.0.0"
	ModelVerificationProtocol = "/ollama/model-verify/1.0.0"

	// Transfer constants
	DefaultChunkSize    = 1024 * 1024 // 1MB chunks
	MaxConcurrentChunks = 10
	ChunkRetryAttempts  = 3
	TransferTimeout     = 30 * time.Minute
	VerificationTimeout = 5 * time.Minute
)

// P2PTransferEngine handles P2P model transfers with chunking and verification
type P2PTransferEngine struct {
	mu sync.RWMutex

	// Transfer management
	activeTransfers map[string]*P2PTransfer
	chunkCache      map[string]*ModelChunk

	// Configuration
	config *TransferConfig

	// Metrics
	metrics *TransferMetrics

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// P2PTransfer represents an active P2P model transfer
type P2PTransfer struct {
	TransferID   string  `json:"transfer_id"`
	ModelName    string  `json:"model_name"`
	ModelVersion string  `json:"model_version"`
	SourcePeer   peer.ID `json:"source_peer"`
	TargetPeer   peer.ID `json:"target_peer"`

	// Transfer metadata
	TotalSize   int64 `json:"total_size"`
	ChunkSize   int64 `json:"chunk_size"`
	TotalChunks int   `json:"total_chunks"`

	// Progress tracking
	CompletedChunks  int     `json:"completed_chunks"`
	BytesTransferred int64   `json:"bytes_transferred"`
	Progress         float64 `json:"progress"`

	// Chunk management
	Chunks     map[int]*ChunkTransfer `json:"chunks"`
	ChunkQueue chan int               `json:"-"`

	// Transfer state
	Status    TransferStatus `json:"status"`
	StartTime time.Time      `json:"start_time"`
	EndTime   time.Time      `json:"end_time,omitempty"`
	Duration  time.Duration  `json:"duration,omitempty"`

	// Error handling
	ErrorCount int    `json:"error_count"`
	LastError  string `json:"last_error,omitempty"`

	// Verification
	ExpectedChecksum string `json:"expected_checksum"`
	ActualChecksum   string `json:"actual_checksum,omitempty"`
	Verified         bool   `json:"verified"`

	// Synchronization
	mu         sync.RWMutex  `json:"-"`
	completeCh chan struct{} `json:"-"`
}

// ChunkTransfer represents the transfer of a single chunk
type ChunkTransfer struct {
	ChunkIndex   int           `json:"chunk_index"`
	Offset       int64         `json:"offset"`
	Size         int64         `json:"size"`
	Checksum     string        `json:"checksum"`
	Status       ChunkStatus   `json:"status"`
	RetryCount   int           `json:"retry_count"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      time.Time     `json:"end_time,omitempty"`
	Duration     time.Duration `json:"duration,omitempty"`
	ErrorMessage string        `json:"error_message,omitempty"`
}

// ModelChunk represents a chunk of model data
type ModelChunk struct {
	ModelName  string    `json:"model_name"`
	ChunkIndex int       `json:"chunk_index"`
	Offset     int64     `json:"offset"`
	Size       int64     `json:"size"`
	Data       []byte    `json:"data"`
	Checksum   string    `json:"checksum"`
	CreatedAt  time.Time `json:"created_at"`
}

// TransferConfig configures P2P transfers
type TransferConfig struct {
	ChunkSize           int64
	MaxConcurrentChunks int
	RetryAttempts       int
	TransferTimeout     time.Duration
	VerificationTimeout time.Duration
	EnableCompression   bool
	EnableEncryption    bool
	CacheChunks         bool
	MaxCacheSize        int64
}

// TransferMetrics tracks transfer performance
type TransferMetrics struct {
	TotalTransfers        int64         `json:"total_transfers"`
	SuccessfulTransfers   int64         `json:"successful_transfers"`
	FailedTransfers       int64         `json:"failed_transfers"`
	TotalBytesTransferred int64         `json:"total_bytes_transferred"`
	AverageTransferSpeed  float64       `json:"average_transfer_speed"` // bytes per second
	AverageTransferTime   time.Duration `json:"average_transfer_time"`
	ChunkRetries          int64         `json:"chunk_retries"`
	VerificationFailures  int64         `json:"verification_failures"`
	LastUpdated           time.Time     `json:"last_updated"`
}

// Transfer and chunk status enums
type TransferStatus string

const (
	TransferStatusPending   TransferStatus = "pending"
	TransferStatusActive    TransferStatus = "active"
	TransferStatusCompleted TransferStatus = "completed"
	TransferStatusFailed    TransferStatus = "failed"
	TransferStatusCancelled TransferStatus = "cancelled"
	TransferStatusVerifying TransferStatus = "verifying"
)

type ChunkStatus string

const (
	ChunkStatusPending   ChunkStatus = "pending"
	ChunkStatusActive    ChunkStatus = "active"
	ChunkStatusCompleted ChunkStatus = "completed"
	ChunkStatusFailed    ChunkStatus = "failed"
	ChunkStatusRetrying  ChunkStatus = "retrying"
)

// NewP2PTransferEngine creates a new P2P transfer engine
func NewP2PTransferEngine(config *TransferConfig) *P2PTransferEngine {
	ctx, cancel := context.WithCancel(context.Background())

	if config == nil {
		config = &TransferConfig{
			ChunkSize:           DefaultChunkSize,
			MaxConcurrentChunks: MaxConcurrentChunks,
			RetryAttempts:       ChunkRetryAttempts,
			TransferTimeout:     TransferTimeout,
			VerificationTimeout: VerificationTimeout,
			EnableCompression:   true,
			EnableEncryption:    true,
			CacheChunks:         true,
			MaxCacheSize:        100 * 1024 * 1024, // 100MB cache
		}
	}

	engine := &P2PTransferEngine{
		activeTransfers: make(map[string]*P2PTransfer),
		chunkCache:      make(map[string]*ModelChunk),
		config:          config,
		metrics:         &TransferMetrics{},
		ctx:             ctx,
		cancel:          cancel,
	}

	// Start background tasks
	engine.wg.Add(1)
	go engine.metricsLoop()

	return engine
}

// StartTransfer initiates a P2P model transfer
func (e *P2PTransferEngine) StartTransfer(modelName, modelVersion string, sourcePeer, targetPeer peer.ID, totalSize int64, expectedChecksum string) (*P2PTransfer, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	transferID := fmt.Sprintf("transfer_%s_%s_%d", modelName, modelVersion, time.Now().UnixNano())

	// Calculate chunk information
	chunkSize := e.config.ChunkSize
	totalChunks := int(math.Ceil(float64(totalSize) / float64(chunkSize)))

	// Create transfer
	transfer := &P2PTransfer{
		TransferID:       transferID,
		ModelName:        modelName,
		ModelVersion:     modelVersion,
		SourcePeer:       sourcePeer,
		TargetPeer:       targetPeer,
		TotalSize:        totalSize,
		ChunkSize:        chunkSize,
		TotalChunks:      totalChunks,
		Chunks:           make(map[int]*ChunkTransfer),
		ChunkQueue:       make(chan int, totalChunks),
		Status:           TransferStatusPending,
		StartTime:        time.Now(),
		ExpectedChecksum: expectedChecksum,
		completeCh:       make(chan struct{}),
	}

	// Initialize chunks
	for i := 0; i < totalChunks; i++ {
		offset := int64(i) * chunkSize
		size := chunkSize
		if i == totalChunks-1 {
			// Last chunk might be smaller
			size = totalSize - offset
		}

		chunk := &ChunkTransfer{
			ChunkIndex: i,
			Offset:     offset,
			Size:       size,
			Status:     ChunkStatusPending,
		}

		transfer.Chunks[i] = chunk
		transfer.ChunkQueue <- i
	}

	// Store transfer
	e.activeTransfers[transferID] = transfer

	// Start transfer processing
	go e.processTransfer(transfer)

	e.metrics.TotalTransfers++

	return transfer, nil
}

// processTransfer processes a P2P transfer with concurrent chunk downloads
func (e *P2PTransferEngine) processTransfer(transfer *P2PTransfer) {
	defer func() {
		close(transfer.completeCh)
		e.finalizeTransfer(transfer)
	}()

	transfer.mu.Lock()
	transfer.Status = TransferStatusActive
	transfer.mu.Unlock()

	// Create worker pool for concurrent chunk transfers
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, e.config.MaxConcurrentChunks)

	// Process chunks concurrently
	for i := 0; i < transfer.TotalChunks; i++ {
		select {
		case chunkIndex := <-transfer.ChunkQueue:
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				semaphore <- struct{}{}        // Acquire semaphore
				defer func() { <-semaphore }() // Release semaphore

				e.transferChunk(transfer, idx)
			}(chunkIndex)

		case <-time.After(e.config.TransferTimeout):
			transfer.mu.Lock()
			transfer.Status = TransferStatusFailed
			transfer.LastError = "transfer timeout"
			transfer.mu.Unlock()
			return

		case <-e.ctx.Done():
			transfer.mu.Lock()
			transfer.Status = TransferStatusCancelled
			transfer.mu.Unlock()
			return
		}
	}

	// Wait for all chunks to complete
	wg.Wait()

	// Check if all chunks completed successfully
	if e.allChunksCompleted(transfer) {
		// Verify the complete model
		if e.verifyTransfer(transfer) {
			transfer.mu.Lock()
			transfer.Status = TransferStatusCompleted
			transfer.Verified = true
			transfer.mu.Unlock()
			e.metrics.SuccessfulTransfers++
		} else {
			transfer.mu.Lock()
			transfer.Status = TransferStatusFailed
			transfer.LastError = "verification failed"
			transfer.mu.Unlock()
			e.metrics.FailedTransfers++
			e.metrics.VerificationFailures++
		}
	} else {
		transfer.mu.Lock()
		transfer.Status = TransferStatusFailed
		transfer.LastError = "incomplete transfer"
		transfer.mu.Unlock()
		e.metrics.FailedTransfers++
	}
}

// transferChunk transfers a single chunk with retry logic
func (e *P2PTransferEngine) transferChunk(transfer *P2PTransfer, chunkIndex int) {
	chunk := transfer.Chunks[chunkIndex]

	for attempt := 0; attempt < e.config.RetryAttempts; attempt++ {
		chunk.StartTime = time.Now()
		chunk.Status = ChunkStatusActive
		chunk.RetryCount = attempt

		// Check cache first
		if e.config.CacheChunks {
			cacheKey := fmt.Sprintf("%s_%d", transfer.ModelName, chunkIndex)
			if cachedChunk, exists := e.chunkCache[cacheKey]; exists {
				// Use cached chunk
				chunk.Status = ChunkStatusCompleted
				chunk.EndTime = time.Now()
				chunk.Duration = chunk.EndTime.Sub(chunk.StartTime)
				chunk.Checksum = cachedChunk.Checksum

				e.updateTransferProgress(transfer)
				return
			}
		}

		// Attempt to transfer chunk
		if err := e.downloadChunk(transfer, chunk); err != nil {
			chunk.Status = ChunkStatusFailed
			chunk.ErrorMessage = err.Error()

			if attempt < e.config.RetryAttempts-1 {
				chunk.Status = ChunkStatusRetrying
				e.metrics.ChunkRetries++
				time.Sleep(time.Duration(attempt+1) * time.Second) // Exponential backoff
				continue
			}

			// Final attempt failed
			transfer.mu.Lock()
			transfer.ErrorCount++
			transfer.LastError = fmt.Sprintf("chunk %d failed: %v", chunkIndex, err)
			transfer.mu.Unlock()
			return
		}

		// Success
		chunk.Status = ChunkStatusCompleted
		chunk.EndTime = time.Now()
		chunk.Duration = chunk.EndTime.Sub(chunk.StartTime)

		e.updateTransferProgress(transfer)
		return
	}
}

// downloadChunk downloads a single chunk from the source peer
func (e *P2PTransferEngine) downloadChunk(transfer *P2PTransfer, chunk *ChunkTransfer) error {
	// In a real implementation, this would:
	// 1. Open a stream to the source peer
	// 2. Send a chunk request with offset and size
	// 3. Receive the chunk data
	// 4. Verify the chunk checksum
	// 5. Store the chunk data

	// For now, simulate the download
	time.Sleep(time.Duration(chunk.Size/1024/1024) * 100 * time.Millisecond) // Simulate based on size

	// Generate a mock checksum
	data := make([]byte, chunk.Size)
	hash := sha256.Sum256(data)
	chunk.Checksum = hex.EncodeToString(hash[:])

	// Cache the chunk if enabled
	if e.config.CacheChunks {
		cacheKey := fmt.Sprintf("%s_%d", transfer.ModelName, chunk.ChunkIndex)
		e.chunkCache[cacheKey] = &ModelChunk{
			ModelName:  transfer.ModelName,
			ChunkIndex: chunk.ChunkIndex,
			Offset:     chunk.Offset,
			Size:       chunk.Size,
			Data:       data,
			Checksum:   chunk.Checksum,
			CreatedAt:  time.Now(),
		}

		// Manage cache size
		e.manageCacheSize()
	}

	return nil
}

// updateTransferProgress updates the overall transfer progress
func (e *P2PTransferEngine) updateTransferProgress(transfer *P2PTransfer) {
	transfer.mu.Lock()
	defer transfer.mu.Unlock()

	completedChunks := 0
	bytesTransferred := int64(0)

	for _, chunk := range transfer.Chunks {
		if chunk.Status == ChunkStatusCompleted {
			completedChunks++
			bytesTransferred += chunk.Size
		}
	}

	transfer.CompletedChunks = completedChunks
	transfer.BytesTransferred = bytesTransferred
	transfer.Progress = float64(completedChunks) / float64(transfer.TotalChunks)
}

// allChunksCompleted checks if all chunks have been completed
func (e *P2PTransferEngine) allChunksCompleted(transfer *P2PTransfer) bool {
	transfer.mu.RLock()
	defer transfer.mu.RUnlock()

	for _, chunk := range transfer.Chunks {
		if chunk.Status != ChunkStatusCompleted {
			return false
		}
	}
	return true
}

// verifyTransfer verifies the integrity of the complete transfer
func (e *P2PTransferEngine) verifyTransfer(transfer *P2PTransfer) bool {
	transfer.mu.Lock()
	transfer.Status = TransferStatusVerifying
	transfer.mu.Unlock()

	// Calculate checksum of all chunks combined
	hash := sha256.New()

	for i := 0; i < transfer.TotalChunks; i++ {
		chunk := transfer.Chunks[i]
		if chunk.Status != ChunkStatusCompleted {
			return false
		}

		// In a real implementation, you would read the actual chunk data
		// For now, use the chunk checksum as part of the overall hash
		hash.Write([]byte(chunk.Checksum))
	}

	actualChecksum := hex.EncodeToString(hash.Sum(nil))

	transfer.mu.Lock()
	transfer.ActualChecksum = actualChecksum
	transfer.mu.Unlock()

	// Compare with expected checksum
	return actualChecksum == transfer.ExpectedChecksum
}

// finalizeTransfer finalizes a transfer and updates metrics
func (e *P2PTransferEngine) finalizeTransfer(transfer *P2PTransfer) {
	transfer.mu.Lock()
	transfer.EndTime = time.Now()
	transfer.Duration = transfer.EndTime.Sub(transfer.StartTime)
	transfer.mu.Unlock()

	// Update metrics
	e.mu.Lock()
	e.metrics.TotalBytesTransferred += transfer.BytesTransferred

	if transfer.Status == TransferStatusCompleted {
		// Update average transfer speed
		if transfer.Duration > 0 {
			speed := float64(transfer.BytesTransferred) / transfer.Duration.Seconds()
			e.metrics.AverageTransferSpeed = (e.metrics.AverageTransferSpeed + speed) / 2.0
		}

		// Update average transfer time
		e.metrics.AverageTransferTime = (e.metrics.AverageTransferTime + transfer.Duration) / 2
	}

	e.metrics.LastUpdated = time.Now()
	e.mu.Unlock()
}

// manageCacheSize manages the chunk cache size
func (e *P2PTransferEngine) manageCacheSize() {
	if !e.config.CacheChunks {
		return
	}

	// Calculate current cache size
	currentSize := int64(0)
	for _, chunk := range e.chunkCache {
		currentSize += chunk.Size
	}

	// Remove oldest chunks if over limit
	if currentSize > e.config.MaxCacheSize {
		// Convert to slice and sort by creation time
		type cacheEntry struct {
			key   string
			chunk *ModelChunk
		}

		entries := make([]cacheEntry, 0, len(e.chunkCache))
		for key, chunk := range e.chunkCache {
			entries = append(entries, cacheEntry{key, chunk})
		}

		// Sort by creation time (oldest first)
		for i := 0; i < len(entries)-1; i++ {
			for j := i + 1; j < len(entries); j++ {
				if entries[i].chunk.CreatedAt.After(entries[j].chunk.CreatedAt) {
					entries[i], entries[j] = entries[j], entries[i]
				}
			}
		}

		// Remove oldest entries until under limit
		for _, entry := range entries {
			if currentSize <= e.config.MaxCacheSize {
				break
			}
			currentSize -= entry.chunk.Size
			delete(e.chunkCache, entry.key)
		}
	}
}

// GetTransfer returns a transfer by ID
func (e *P2PTransferEngine) GetTransfer(transferID string) (*P2PTransfer, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	transfer, exists := e.activeTransfers[transferID]
	return transfer, exists
}

// GetActiveTransfers returns all active transfers
func (e *P2PTransferEngine) GetActiveTransfers() []*P2PTransfer {
	e.mu.RLock()
	defer e.mu.RUnlock()

	transfers := make([]*P2PTransfer, 0, len(e.activeTransfers))
	for _, transfer := range e.activeTransfers {
		transfers = append(transfers, transfer)
	}

	return transfers
}

// GetMetrics returns transfer metrics
func (e *P2PTransferEngine) GetMetrics() *TransferMetrics {
	e.mu.RLock()
	defer e.mu.RUnlock()

	metrics := *e.metrics
	return &metrics
}

// metricsLoop periodically updates metrics
func (e *P2PTransferEngine) metricsLoop() {
	defer e.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.updateMetrics()
		}
	}
}

// updateMetrics updates transfer metrics
func (e *P2PTransferEngine) updateMetrics() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.metrics.LastUpdated = time.Now()
}

// Close closes the P2P transfer engine
func (e *P2PTransferEngine) Close() error {
	e.cancel()
	e.wg.Wait()

	// Cancel all active transfers
	e.mu.Lock()
	for _, transfer := range e.activeTransfers {
		transfer.mu.Lock()
		if transfer.Status == TransferStatusActive || transfer.Status == TransferStatusPending {
			transfer.Status = TransferStatusCancelled
		}
		transfer.mu.Unlock()
	}
	e.mu.Unlock()

	return nil
}

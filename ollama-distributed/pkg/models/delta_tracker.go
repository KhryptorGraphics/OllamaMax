package models

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// DeltaTracker manages delta operations for incremental model synchronization
type DeltaTracker struct {
	deltaDir string
	logger   *slog.Logger

	// Delta storage
	deltas     map[string][]*Delta
	deltaMutex sync.RWMutex

	// Delta operations
	pendingOps map[string]*DeltaOperation
	opsMutex   sync.RWMutex

	// Compression settings
	compressionEnabled bool
	compressionLevel   int

	ctx    context.Context
	cancel context.CancelFunc
}

// Delta represents a single delta operation
type Delta struct {
	ID         string    `json:"id"`
	ModelName  string    `json:"model_name"`
	Type       DeltaType `json:"type"`
	Offset     int64     `json:"offset"`
	Size       int64     `json:"size"`
	Hash       string    `json:"hash"`
	Data       []byte    `json:"data,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	Compressed bool      `json:"compressed"`
}

// DeltaType represents the type of delta operation
type DeltaType string

const (
	DeltaTypeInsert DeltaType = "insert"
	DeltaTypeUpdate DeltaType = "update"
	DeltaTypeDelete DeltaType = "delete"
)

// DeltaOperation represents a complete delta operation
type DeltaOperation struct {
	ID          string          `json:"id"`
	ModelName   string          `json:"model_name"`
	SourceHash  string          `json:"source_hash"`
	TargetHash  string          `json:"target_hash"`
	Deltas      []*Delta        `json:"deltas"`
	Status      OperationStatus `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	CompletedAt time.Time       `json:"completed_at"`
	Size        int64           `json:"size"`
	Error       string          `json:"error,omitempty"`
}

// OperationStatus represents the status of a delta operation
type OperationStatus string

const (
	OperationStatusPending    OperationStatus = "pending"
	OperationStatusInProgress OperationStatus = "in_progress"
	OperationStatusCompleted  OperationStatus = "completed"
	OperationStatusFailed     OperationStatus = "failed"
)

// DeltaMetadata contains metadata about a delta set
type DeltaMetadata struct {
	ModelName        string    `json:"model_name"`
	FromVersion      string    `json:"from_version"`
	ToVersion        string    `json:"to_version"`
	DeltaCount       int       `json:"delta_count"`
	TotalSize        int64     `json:"total_size"`
	CompressedSize   int64     `json:"compressed_size"`
	CreatedAt        time.Time `json:"created_at"`
	CompressionRatio float64   `json:"compression_ratio"`
}

// NewDeltaTracker creates a new delta tracker
func NewDeltaTracker(deltaDir string, logger *slog.Logger) (*DeltaTracker, error) {
	if err := os.MkdirAll(deltaDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create delta directory: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	dt := &DeltaTracker{
		deltaDir:           deltaDir,
		logger:             logger,
		deltas:             make(map[string][]*Delta),
		pendingOps:         make(map[string]*DeltaOperation),
		compressionEnabled: true,
		compressionLevel:   6,
		ctx:                ctx,
		cancel:             cancel,
	}

	// Load existing deltas
	if err := dt.loadDeltas(); err != nil {
		return nil, fmt.Errorf("failed to load deltas: %w", err)
	}

	return dt, nil
}

// CreateDelta creates a delta between two model versions
func (dt *DeltaTracker) CreateDelta(modelName, sourceFile, targetFile string) (*DeltaOperation, error) {
	dt.logger.Info("creating delta", "model", modelName, "source", sourceFile, "target", targetFile)

	// Calculate hashes
	sourceHash, err := dt.calculateFileHash(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate source hash: %w", err)
	}

	targetHash, err := dt.calculateFileHash(targetFile)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate target hash: %w", err)
	}

	// Create operation
	opID := fmt.Sprintf("%s_%s_%s", modelName, sourceHash[:8], targetHash[:8])
	op := &DeltaOperation{
		ID:         opID,
		ModelName:  modelName,
		SourceHash: sourceHash,
		TargetHash: targetHash,
		Status:     OperationStatusPending,
		CreatedAt:  time.Now(),
	}

	// Store pending operation
	dt.opsMutex.Lock()
	dt.pendingOps[opID] = op
	dt.opsMutex.Unlock()

	// Generate deltas
	deltas, err := dt.generateDeltas(sourceFile, targetFile)
	if err != nil {
		op.Status = OperationStatusFailed
		op.Error = err.Error()
		return op, fmt.Errorf("failed to generate deltas: %w", err)
	}

	op.Deltas = deltas
	op.Status = OperationStatusCompleted
	op.CompletedAt = time.Now()

	// Calculate total size
	var totalSize int64
	for _, delta := range deltas {
		totalSize += delta.Size
	}
	op.Size = totalSize

	// Store deltas
	dt.deltaMutex.Lock()
	dt.deltas[modelName] = append(dt.deltas[modelName], deltas...)
	dt.deltaMutex.Unlock()

	// Save to disk
	if err := dt.saveDelta(op); err != nil {
		dt.logger.Error("failed to save delta", "operation", opID, "error", err)
	}

	dt.logger.Info("delta created", "operation", opID, "deltas", len(deltas), "size", totalSize)

	return op, nil
}

// ApplyDelta applies a delta operation to a model file
func (dt *DeltaTracker) ApplyDelta(targetFile string, op *DeltaOperation) error {
	dt.logger.Info("applying delta", "operation", op.ID, "target", targetFile)

	// Open target file
	file, err := os.OpenFile(targetFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open target file: %w", err)
	}
	defer file.Close()

	// Apply deltas in order
	sort.Slice(op.Deltas, func(i, j int) bool {
		return op.Deltas[i].Offset < op.Deltas[j].Offset
	})

	for _, delta := range op.Deltas {
		if err := dt.applyDelta(file, delta); err != nil {
			return fmt.Errorf("failed to apply delta %s: %w", delta.ID, err)
		}
	}

	// Verify result
	resultHash, err := dt.calculateFileHash(targetFile)
	if err != nil {
		return fmt.Errorf("failed to calculate result hash: %w", err)
	}

	if resultHash != op.TargetHash {
		return fmt.Errorf("delta application failed: hash mismatch (expected %s, got %s)", op.TargetHash, resultHash)
	}

	dt.logger.Info("delta applied successfully", "operation", op.ID)
	return nil
}

// GetDeltas returns all deltas for a model
func (dt *DeltaTracker) GetDeltas(modelName string) []*Delta {
	dt.deltaMutex.RLock()
	defer dt.deltaMutex.RUnlock()

	deltas := dt.deltas[modelName]
	result := make([]*Delta, len(deltas))
	copy(result, deltas)

	return result
}

// GetDeltaOperation returns a delta operation by ID
func (dt *DeltaTracker) GetDeltaOperation(opID string) (*DeltaOperation, bool) {
	dt.opsMutex.RLock()
	defer dt.opsMutex.RUnlock()

	op, exists := dt.pendingOps[opID]
	return op, exists
}

// generateDeltas generates deltas between two files
func (dt *DeltaTracker) generateDeltas(sourceFile, targetFile string) ([]*Delta, error) {
	sourceData, err := os.ReadFile(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read source file: %w", err)
	}

	targetData, err := os.ReadFile(targetFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read target file: %w", err)
	}

	return dt.generateDeltasFromData(sourceData, targetData)
}

// generateDeltasFromData generates deltas from byte data
func (dt *DeltaTracker) generateDeltasFromData(sourceData, targetData []byte) ([]*Delta, error) {
	var deltas []*Delta

	// Simple delta generation algorithm
	// In a real implementation, this would use a more sophisticated algorithm
	// like rsync's rolling hash or binary diff algorithms

	chunkSize := 1024 * 4 // 4KB chunks

	sourceOffset := 0
	targetOffset := 0

	for targetOffset < len(targetData) {
		chunkEnd := targetOffset + chunkSize
		if chunkEnd > len(targetData) {
			chunkEnd = len(targetData)
		}

		targetChunk := targetData[targetOffset:chunkEnd]

		// Check if this chunk exists in source
		found := false
		for i := sourceOffset; i < len(sourceData)-len(targetChunk); i++ {
			if dt.compareChunks(sourceData[i:i+len(targetChunk)], targetChunk) {
				// Chunk exists, no delta needed
				found = true
				sourceOffset = i + len(targetChunk)
				break
			}
		}

		if !found {
			// Chunk is new or modified, create delta
			deltaID := fmt.Sprintf("delta_%d_%d", time.Now().UnixNano(), targetOffset)
			hash := sha256.Sum256(targetChunk)

			delta := &Delta{
				ID:        deltaID,
				Type:      DeltaTypeInsert,
				Offset:    int64(targetOffset),
				Size:      int64(len(targetChunk)),
				Hash:      hex.EncodeToString(hash[:]),
				Data:      targetChunk,
				Timestamp: time.Now(),
			}

			// Apply compression if enabled
			if dt.compressionEnabled {
				compressed, err := dt.compressData(targetChunk)
				if err == nil && len(compressed) < len(targetChunk) {
					delta.Data = compressed
					delta.Compressed = true
				}
			}

			deltas = append(deltas, delta)
		}

		targetOffset = chunkEnd
	}

	return deltas, nil
}

// applyDelta applies a single delta to a file
func (dt *DeltaTracker) applyDelta(file *os.File, delta *Delta) error {
	// Seek to position
	if _, err := file.Seek(delta.Offset, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek to position: %w", err)
	}

	data := delta.Data

	// Decompress if needed
	if delta.Compressed {
		decompressed, err := dt.decompressData(data)
		if err != nil {
			return fmt.Errorf("failed to decompress data: %w", err)
		}
		data = decompressed
	}

	// Apply delta based on type
	switch delta.Type {
	case DeltaTypeInsert, DeltaTypeUpdate:
		if _, err := file.Write(data); err != nil {
			return fmt.Errorf("failed to write data: %w", err)
		}
	case DeltaTypeDelete:
		// For delete operations, we would need to handle file truncation
		// This is simplified for demonstration
		return fmt.Errorf("delete operations not fully implemented")
	}

	return nil
}

// compareChunks compares two byte slices
func (dt *DeltaTracker) compareChunks(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// calculateFileHash calculates SHA256 hash of a file
func (dt *DeltaTracker) calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// compressData compresses data using a simple compression algorithm
func (dt *DeltaTracker) compressData(data []byte) ([]byte, error) {
	// In a real implementation, this would use a proper compression library
	// like gzip, lz4, or zstd
	// For demonstration, we'll return the original data
	return data, nil
}

// decompressData decompresses data
func (dt *DeltaTracker) decompressData(data []byte) ([]byte, error) {
	// In a real implementation, this would use a proper decompression library
	// For demonstration, we'll return the original data
	return data, nil
}

// loadDeltas loads existing deltas from disk
func (dt *DeltaTracker) loadDeltas() error {
	return filepath.Walk(dt.deltaDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".delta" {
			if err := dt.loadDeltaFile(path); err != nil {
				dt.logger.Error("failed to load delta file", "path", path, "error", err)
			}
		}

		return nil
	})
}

// loadDeltaFile loads a single delta file
func (dt *DeltaTracker) loadDeltaFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read delta file: %w", err)
	}

	var op DeltaOperation
	if err := json.Unmarshal(data, &op); err != nil {
		return fmt.Errorf("failed to unmarshal delta operation: %w", err)
	}

	// Store operation
	dt.opsMutex.Lock()
	dt.pendingOps[op.ID] = &op
	dt.opsMutex.Unlock()

	// Store deltas
	dt.deltaMutex.Lock()
	dt.deltas[op.ModelName] = append(dt.deltas[op.ModelName], op.Deltas...)
	dt.deltaMutex.Unlock()

	return nil
}

// saveDelta saves a delta operation to disk
func (dt *DeltaTracker) saveDelta(op *DeltaOperation) error {
	filename := fmt.Sprintf("%s.delta", op.ID)
	path := filepath.Join(dt.deltaDir, filename)

	data, err := json.MarshalIndent(op, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal delta operation: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// GetDeltaMetadata returns metadata about deltas for a model
func (dt *DeltaTracker) GetDeltaMetadata(modelName string) *DeltaMetadata {
	dt.deltaMutex.RLock()
	defer dt.deltaMutex.RUnlock()

	deltas := dt.deltas[modelName]
	if len(deltas) == 0 {
		return nil
	}

	var totalSize, compressedSize int64
	for _, delta := range deltas {
		totalSize += delta.Size
		if delta.Compressed {
			compressedSize += int64(len(delta.Data))
		} else {
			compressedSize += delta.Size
		}
	}

	compressionRatio := 1.0
	if totalSize > 0 {
		compressionRatio = float64(compressedSize) / float64(totalSize)
	}

	return &DeltaMetadata{
		ModelName:        modelName,
		DeltaCount:       len(deltas),
		TotalSize:        totalSize,
		CompressedSize:   compressedSize,
		CompressionRatio: compressionRatio,
	}
}

// Cleanup removes old delta files
func (dt *DeltaTracker) Cleanup(maxAge time.Duration) error {
	dt.deltaMutex.Lock()
	defer dt.deltaMutex.Unlock()

	cutoff := time.Now().Add(-maxAge)

	for modelName, deltas := range dt.deltas {
		var remaining []*Delta

		for _, delta := range deltas {
			if delta.Timestamp.After(cutoff) {
				remaining = append(remaining, delta)
			}
		}

		dt.deltas[modelName] = remaining
	}

	return nil
}

// Close closes the delta tracker
func (dt *DeltaTracker) Close() error {
	dt.cancel()
	return nil
}

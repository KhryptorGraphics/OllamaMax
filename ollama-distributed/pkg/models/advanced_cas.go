package models

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AdvancedCAS implements an advanced content-addressed storage system with deduplication and compression
type AdvancedCAS struct {
	mu sync.RWMutex

	// Storage configuration
	config *CASConfig

	// Storage backends
	primaryStore    StorageBackend
	secondaryStores []StorageBackend

	// Content management
	contentIndex map[string]*ContentEntry
	chunkIndex   map[string]*ChunkEntry

	// Deduplication
	deduplicator *Deduplicator

	// Compression
	compressor *Compressor

	// Storage optimization
	optimizer *StorageOptimizer

	// Metrics and monitoring
	metrics *CASMetrics

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// ContentEntry represents a content entry in the CAS
type ContentEntry struct {
	Hash           string   `json:"hash"`
	Size           int64    `json:"size"`
	CompressedSize int64    `json:"compressed_size"`
	ChunkHashes    []string `json:"chunk_hashes"`

	// Content metadata
	ContentType string          `json:"content_type"`
	Encoding    string          `json:"encoding"`
	Compression CompressionType `json:"compression"`

	// Storage information
	StorageBackend string `json:"storage_backend"`
	StoragePath    string `json:"storage_path"`

	// Access tracking
	RefCount     int       `json:"ref_count"`
	CreatedAt    time.Time `json:"created_at"`
	LastAccessed time.Time `json:"last_accessed"`
	AccessCount  int64     `json:"access_count"`

	// Optimization
	HotData          bool    `json:"hot_data"`
	CompressionRatio float64 `json:"compression_ratio"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata"`
}

// ChunkEntry represents a chunk in the CAS
type ChunkEntry struct {
	Hash           string `json:"hash"`
	Size           int64  `json:"size"`
	CompressedSize int64  `json:"compressed_size"`

	// Storage information
	StorageBackend string `json:"storage_backend"`
	StoragePath    string `json:"storage_path"`

	// Deduplication
	RefCount    int      `json:"ref_count"`
	ContentRefs []string `json:"content_refs"`

	// Access tracking
	CreatedAt    time.Time `json:"created_at"`
	LastAccessed time.Time `json:"last_accessed"`
	AccessCount  int64     `json:"access_count"`

	// Optimization
	Compression      CompressionType `json:"compression"`
	CompressionRatio float64         `json:"compression_ratio"`
}

// CASConfig configures the advanced CAS
type CASConfig struct {
	// Storage settings
	PrimaryStoragePath    string
	SecondaryStoragePaths []string
	MaxStorageSize        int64
	MaxContentEntries     int

	// Chunking settings
	ChunkSize              int64
	MinChunkSize           int64
	MaxChunkSize           int64
	EnableVariableChunking bool

	// Compression settings
	EnableCompression    bool
	CompressionLevel     int
	CompressionThreshold int64
	CompressionAlgorithm CompressionType

	// Deduplication settings
	EnableDeduplication    bool
	DeduplicationThreshold int64
	ChunkDeduplication     bool

	// Optimization settings
	EnableOptimization   bool
	OptimizationInterval time.Duration
	HotDataThreshold     int64
	ColdDataThreshold    time.Duration

	// Cleanup settings
	EnableCleanup   bool
	CleanupInterval time.Duration
	MaxAge          time.Duration
	MinRefCount     int

	// Performance settings
	MaxConcurrentOperations int
	IOBufferSize            int
	EnableAsyncWrites       bool
}

// CASMetrics tracks CAS performance and usage
type CASMetrics struct {
	// Storage metrics
	TotalContentEntries  int64 `json:"total_content_entries"`
	TotalChunkEntries    int64 `json:"total_chunk_entries"`
	TotalStorageUsed     int64 `json:"total_storage_used"`
	TotalStorageCapacity int64 `json:"total_storage_capacity"`

	// Deduplication metrics
	DeduplicationRatio   float64 `json:"deduplication_ratio"`
	DuplicateChunksFound int64   `json:"duplicate_chunks_found"`
	SpaceSavedByDedup    int64   `json:"space_saved_by_dedup"`

	// Compression metrics
	CompressionRatio        float64 `json:"compression_ratio"`
	CompressedEntries       int64   `json:"compressed_entries"`
	SpaceSavedByCompression int64   `json:"space_saved_by_compression"`

	// Performance metrics
	AverageReadLatency  time.Duration `json:"average_read_latency"`
	AverageWriteLatency time.Duration `json:"average_write_latency"`
	TotalReads          int64         `json:"total_reads"`
	TotalWrites         int64         `json:"total_writes"`

	// Access patterns
	HotDataAccesses  int64   `json:"hot_data_accesses"`
	ColdDataAccesses int64   `json:"cold_data_accesses"`
	CacheHitRatio    float64 `json:"cache_hit_ratio"`

	// Storage backend metrics
	BackendMetrics map[string]*BackendMetrics `json:"backend_metrics"`

	// Last updated
	LastUpdated time.Time `json:"last_updated"`
}

// BackendMetrics tracks metrics for a storage backend
type BackendMetrics struct {
	BackendName     string        `json:"backend_name"`
	StorageUsed     int64         `json:"storage_used"`
	StorageCapacity int64         `json:"storage_capacity"`
	ReadOperations  int64         `json:"read_operations"`
	WriteOperations int64         `json:"write_operations"`
	AverageLatency  time.Duration `json:"average_latency"`
	ErrorCount      int64         `json:"error_count"`
}

// Deduplicator handles content deduplication
type Deduplicator struct {
	enabled      bool
	threshold    int64
	chunkSize    int64
	algorithm    DeduplicationAlgorithm
	fingerprints map[string]string
	mu           sync.RWMutex
}

// Compressor handles content compression
type Compressor struct {
	enabled   bool
	algorithm CompressionType
	level     int
	threshold int64
	mu        sync.RWMutex
}

// StorageOptimizer optimizes storage layout and access patterns
type StorageOptimizer struct {
	enabled             bool
	hotThreshold        int64
	coldThreshold       time.Duration
	lastOptimization    time.Time
	optimizationHistory []*OptimizationResult
}

// StorageBackend interface for different storage backends
type StorageBackend interface {
	Name() string
	Store(hash string, data io.Reader) error
	Retrieve(hash string) (io.ReadCloser, error)
	Delete(hash string) error
	Exists(hash string) bool
	Size(hash string) (int64, error)
	List() ([]string, error)
	GetMetrics() *BackendMetrics
	Close() error
}

// Enums and constants
type CompressionType string

const (
	CompressionTypeNone   CompressionType = "none"
	CompressionTypeGzip   CompressionType = "gzip"
	CompressionTypeLZ4    CompressionType = "lz4"
	CompressionTypeZstd   CompressionType = "zstd"
	CompressionTypeBrotli CompressionType = "brotli"
)

type DeduplicationAlgorithm string

const (
	DeduplicationAlgorithmSHA256  DeduplicationAlgorithm = "sha256"
	DeduplicationAlgorithmBlake2b DeduplicationAlgorithm = "blake2b"
	DeduplicationAlgorithmRolling DeduplicationAlgorithm = "rolling"
)

// NewAdvancedCAS creates a new advanced content-addressed storage system
func NewAdvancedCAS(config *CASConfig) (*AdvancedCAS, error) {
	ctx, cancel := context.WithCancel(context.Background())

	if config == nil {
		config = &CASConfig{
			PrimaryStoragePath:      "./cas_storage",
			MaxStorageSize:          100 * 1024 * 1024 * 1024, // 100GB
			MaxContentEntries:       1000000,
			ChunkSize:               1024 * 1024,      // 1MB
			MinChunkSize:            64 * 1024,        // 64KB
			MaxChunkSize:            16 * 1024 * 1024, // 16MB
			EnableVariableChunking:  true,
			EnableCompression:       true,
			CompressionLevel:        6,
			CompressionThreshold:    1024, // 1KB
			CompressionAlgorithm:    CompressionTypeGzip,
			EnableDeduplication:     true,
			DeduplicationThreshold:  1024, // 1KB
			ChunkDeduplication:      true,
			EnableOptimization:      true,
			OptimizationInterval:    time.Hour,
			HotDataThreshold:        100,
			ColdDataThreshold:       7 * 24 * time.Hour, // 7 days
			EnableCleanup:           true,
			CleanupInterval:         24 * time.Hour,
			MaxAge:                  30 * 24 * time.Hour, // 30 days
			MinRefCount:             1,
			MaxConcurrentOperations: 10,
			IOBufferSize:            64 * 1024, // 64KB
			EnableAsyncWrites:       true,
		}
	}

	cas := &AdvancedCAS{
		config:       config,
		contentIndex: make(map[string]*ContentEntry),
		chunkIndex:   make(map[string]*ChunkEntry),
		metrics: &CASMetrics{
			BackendMetrics: make(map[string]*BackendMetrics),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize storage backends
	if err := cas.initializeStorageBackends(); err != nil {
		return nil, fmt.Errorf("failed to initialize storage backends: %w", err)
	}

	// Initialize deduplicator
	if config.EnableDeduplication {
		cas.deduplicator = &Deduplicator{
			enabled:      true,
			threshold:    config.DeduplicationThreshold,
			chunkSize:    config.ChunkSize,
			algorithm:    DeduplicationAlgorithmSHA256,
			fingerprints: make(map[string]string),
		}
	}

	// Initialize compressor
	if config.EnableCompression {
		cas.compressor = &Compressor{
			enabled:   true,
			algorithm: config.CompressionAlgorithm,
			level:     config.CompressionLevel,
			threshold: config.CompressionThreshold,
		}
	}

	// Initialize optimizer
	if config.EnableOptimization {
		cas.optimizer = &StorageOptimizer{
			enabled:             true,
			hotThreshold:        config.HotDataThreshold,
			coldThreshold:       config.ColdDataThreshold,
			optimizationHistory: make([]*OptimizationResult, 0),
		}
	}

	// Start background tasks
	cas.wg.Add(3)
	go cas.optimizationLoop()
	go cas.cleanupLoop()
	go cas.metricsLoop()

	return cas, nil
}

// initializeStorageBackends initializes storage backends
func (cas *AdvancedCAS) initializeStorageBackends() error {
	// Initialize primary storage backend
	primaryBackend, err := NewFileSystemBackend(cas.config.PrimaryStoragePath)
	if err != nil {
		return fmt.Errorf("failed to initialize primary storage backend: %w", err)
	}
	cas.primaryStore = primaryBackend

	// Initialize secondary storage backends
	for _, path := range cas.config.SecondaryStoragePaths {
		backend, err := NewFileSystemBackend(path)
		if err != nil {
			return fmt.Errorf("failed to initialize secondary storage backend %s: %w", path, err)
		}
		cas.secondaryStores = append(cas.secondaryStores, backend)
	}

	return nil
}

// Store stores content in the CAS with deduplication and compression
func (cas *AdvancedCAS) Store(content io.Reader, metadata map[string]interface{}) (*ContentEntry, error) {
	startTime := time.Now()

	// Read content into memory (for small files) or create temporary file (for large files)
	data, err := io.ReadAll(content)
	if err != nil {
		return nil, fmt.Errorf("failed to read content: %w", err)
	}

	// Calculate content hash
	hash := cas.calculateHash(data)

	cas.mu.Lock()
	defer cas.mu.Unlock()

	// Check if content already exists (deduplication)
	if existing, exists := cas.contentIndex[hash]; exists {
		existing.RefCount++
		existing.LastAccessed = time.Now()
		existing.AccessCount++
		cas.metrics.DuplicateChunksFound++
		return existing, nil
	}

	// Create content entry
	entry := &ContentEntry{
		Hash:         hash,
		Size:         int64(len(data)),
		RefCount:     1,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
		AccessCount:  1,
		Metadata:     metadata,
	}

	// Apply compression if enabled and beneficial
	var processedData []byte
	if cas.shouldCompress(data) {
		compressed, err := cas.compress(data)
		if err == nil && len(compressed) < len(data) {
			processedData = compressed
			entry.CompressedSize = int64(len(compressed))
			entry.Compression = cas.config.CompressionAlgorithm
			entry.CompressionRatio = float64(len(data)) / float64(len(compressed))
			cas.metrics.CompressedEntries++
			cas.metrics.SpaceSavedByCompression += int64(len(data) - len(compressed))
		} else {
			processedData = data
			entry.Compression = CompressionTypeNone
		}
	} else {
		processedData = data
		entry.Compression = CompressionTypeNone
	}

	// Store in primary backend
	if err := cas.primaryStore.Store(hash, bytes.NewReader(processedData)); err != nil {
		return nil, fmt.Errorf("failed to store content: %w", err)
	}

	entry.StorageBackend = cas.primaryStore.Name()
	entry.StoragePath = hash

	// Store in content index
	cas.contentIndex[hash] = entry

	// Update metrics
	cas.metrics.TotalContentEntries++
	cas.metrics.TotalWrites++
	cas.metrics.TotalStorageUsed += entry.Size

	// Update write latency
	latency := time.Since(startTime)
	if cas.metrics.TotalWrites == 1 {
		cas.metrics.AverageWriteLatency = latency
	} else {
		cas.metrics.AverageWriteLatency = (cas.metrics.AverageWriteLatency + latency) / 2
	}

	return entry, nil
}

// Retrieve retrieves content from the CAS
func (cas *AdvancedCAS) Retrieve(hash string) (io.ReadCloser, error) {
	startTime := time.Now()

	cas.mu.RLock()
	entry, exists := cas.contentIndex[hash]
	cas.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("content not found: %s", hash)
	}

	// Update access tracking
	cas.mu.Lock()
	entry.LastAccessed = time.Now()
	entry.AccessCount++
	cas.mu.Unlock()

	// Retrieve from storage backend
	reader, err := cas.primaryStore.Retrieve(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve content: %w", err)
	}

	// Decompress if necessary
	if entry.Compression != CompressionTypeNone {
		decompressed, err := cas.decompress(reader, entry.Compression)
		if err != nil {
			reader.Close()
			return nil, fmt.Errorf("failed to decompress content: %w", err)
		}
		reader.Close()
		reader = decompressed
	}

	// Update metrics
	cas.mu.Lock()
	cas.metrics.TotalReads++
	latency := time.Since(startTime)
	if cas.metrics.TotalReads == 1 {
		cas.metrics.AverageReadLatency = latency
	} else {
		cas.metrics.AverageReadLatency = (cas.metrics.AverageReadLatency + latency) / 2
	}
	cas.mu.Unlock()

	return reader, nil
}

// calculateHash calculates the hash of content
func (cas *AdvancedCAS) calculateHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// shouldCompress determines if content should be compressed
func (cas *AdvancedCAS) shouldCompress(data []byte) bool {
	if cas.compressor == nil || !cas.compressor.enabled {
		return false
	}

	return int64(len(data)) >= cas.compressor.threshold
}

// compress compresses data using the configured algorithm
func (cas *AdvancedCAS) compress(data []byte) ([]byte, error) {
	if cas.compressor == nil {
		return data, nil
	}

	switch cas.compressor.algorithm {
	case CompressionTypeGzip:
		return cas.compressGzip(data)
	default:
		return data, fmt.Errorf("unsupported compression algorithm: %s", cas.compressor.algorithm)
	}
}

// compressGzip compresses data using gzip
func (cas *AdvancedCAS) compressGzip(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)

	if _, err := writer.Write(data); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// decompress decompresses data using the specified algorithm
func (cas *AdvancedCAS) decompress(reader io.ReadCloser, algorithm CompressionType) (io.ReadCloser, error) {
	switch algorithm {
	case CompressionTypeGzip:
		return gzip.NewReader(reader)
	default:
		return reader, nil
	}
}

// optimizationLoop performs periodic storage optimization
func (cas *AdvancedCAS) optimizationLoop() {
	defer cas.wg.Done()

	if cas.optimizer == nil || !cas.optimizer.enabled {
		return
	}

	ticker := time.NewTicker(cas.config.OptimizationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-cas.ctx.Done():
			return
		case <-ticker.C:
			cas.performOptimization()
		}
	}
}

// performOptimization performs storage optimization
func (cas *AdvancedCAS) performOptimization() {
	// Implementation would optimize storage layout, move hot data, etc.
	// For now, this is a placeholder
}

// cleanupLoop performs periodic cleanup
func (cas *AdvancedCAS) cleanupLoop() {
	defer cas.wg.Done()

	if !cas.config.EnableCleanup {
		return
	}

	ticker := time.NewTicker(cas.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-cas.ctx.Done():
			return
		case <-ticker.C:
			cas.performCleanup()
		}
	}
}

// performCleanup performs cleanup of old and unused content
func (cas *AdvancedCAS) performCleanup() {
	cas.mu.Lock()
	defer cas.mu.Unlock()

	cutoff := time.Now().Add(-cas.config.MaxAge)

	for hash, entry := range cas.contentIndex {
		if entry.RefCount <= cas.config.MinRefCount && entry.LastAccessed.Before(cutoff) {
			// Delete from storage
			cas.primaryStore.Delete(hash)

			// Remove from index
			delete(cas.contentIndex, hash)

			// Update metrics
			cas.metrics.TotalContentEntries--
			cas.metrics.TotalStorageUsed -= entry.Size
		}
	}
}

// metricsLoop updates metrics periodically
func (cas *AdvancedCAS) metricsLoop() {
	defer cas.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-cas.ctx.Done():
			return
		case <-ticker.C:
			cas.updateMetrics()
		}
	}
}

// updateMetrics updates CAS metrics
func (cas *AdvancedCAS) updateMetrics() {
	cas.mu.Lock()
	defer cas.mu.Unlock()

	// Calculate deduplication ratio
	if cas.metrics.TotalContentEntries > 0 {
		cas.metrics.DeduplicationRatio = float64(cas.metrics.DuplicateChunksFound) / float64(cas.metrics.TotalContentEntries)
	}

	// Calculate compression ratio
	if cas.metrics.CompressedEntries > 0 {
		cas.metrics.CompressionRatio = float64(cas.metrics.SpaceSavedByCompression) / float64(cas.metrics.TotalStorageUsed)
	}

	cas.metrics.LastUpdated = time.Now()
}

// GetMetrics returns CAS metrics
func (cas *AdvancedCAS) GetMetrics() *CASMetrics {
	cas.mu.RLock()
	defer cas.mu.RUnlock()

	metrics := *cas.metrics
	return &metrics
}

// Close closes the advanced CAS
func (cas *AdvancedCAS) Close() error {
	cas.cancel()
	cas.wg.Wait()

	// Close storage backends
	if cas.primaryStore != nil {
		cas.primaryStore.Close()
	}

	for _, backend := range cas.secondaryStores {
		backend.Close()
	}

	return nil
}

// FileSystemBackend implements a file system storage backend
type FileSystemBackend struct {
	basePath string
	metrics  *BackendMetrics
	mu       sync.RWMutex
}

// NewFileSystemBackend creates a new file system storage backend
func NewFileSystemBackend(basePath string) (*FileSystemBackend, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &FileSystemBackend{
		basePath: basePath,
		metrics: &BackendMetrics{
			BackendName: "filesystem",
		},
	}, nil
}

// Name returns the backend name
func (fs *FileSystemBackend) Name() string {
	return "filesystem"
}

// Store stores data in the file system
func (fs *FileSystemBackend) Store(hash string, data io.Reader) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	startTime := time.Now()

	// Create file path
	filePath := fs.getFilePath(hash)

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		fs.metrics.ErrorCount++
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		fs.metrics.ErrorCount++
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy data to file
	written, err := io.Copy(file, data)
	if err != nil {
		fs.metrics.ErrorCount++
		return fmt.Errorf("failed to write data: %w", err)
	}

	// Update metrics
	fs.metrics.WriteOperations++
	fs.metrics.StorageUsed += written

	// Update average latency
	latency := time.Since(startTime)
	if fs.metrics.WriteOperations == 1 {
		fs.metrics.AverageLatency = latency
	} else {
		fs.metrics.AverageLatency = (fs.metrics.AverageLatency + latency) / 2
	}

	return nil
}

// Retrieve retrieves data from the file system
func (fs *FileSystemBackend) Retrieve(hash string) (io.ReadCloser, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	startTime := time.Now()

	// Get file path
	filePath := fs.getFilePath(hash)

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		fs.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Update metrics
	fs.metrics.ReadOperations++

	// Update average latency
	latency := time.Since(startTime)
	if fs.metrics.ReadOperations == 1 {
		fs.metrics.AverageLatency = latency
	} else {
		fs.metrics.AverageLatency = (fs.metrics.AverageLatency + latency) / 2
	}

	return file, nil
}

// Delete deletes data from the file system
func (fs *FileSystemBackend) Delete(hash string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Get file path
	filePath := fs.getFilePath(hash)

	// Get file size before deletion
	if info, err := os.Stat(filePath); err == nil {
		fs.metrics.StorageUsed -= info.Size()
	}

	// Delete file
	if err := os.Remove(filePath); err != nil {
		fs.metrics.ErrorCount++
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// Exists checks if data exists in the file system
func (fs *FileSystemBackend) Exists(hash string) bool {
	filePath := fs.getFilePath(hash)
	_, err := os.Stat(filePath)
	return err == nil
}

// Size returns the size of stored data
func (fs *FileSystemBackend) Size(hash string) (int64, error) {
	filePath := fs.getFilePath(hash)
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to get file info: %w", err)
	}
	return info.Size(), nil
}

// List returns a list of all stored hashes
func (fs *FileSystemBackend) List() ([]string, error) {
	var hashes []string

	err := filepath.Walk(fs.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			// Extract hash from file path
			relPath, err := filepath.Rel(fs.basePath, path)
			if err != nil {
				return err
			}

			// Remove directory separators to get hash
			hash := filepath.Base(relPath)
			hashes = append(hashes, hash)
		}

		return nil
	})

	return hashes, err
}

// GetMetrics returns backend metrics
func (fs *FileSystemBackend) GetMetrics() *BackendMetrics {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	metrics := *fs.metrics
	return &metrics
}

// Close closes the backend
func (fs *FileSystemBackend) Close() error {
	// Nothing to close for file system backend
	return nil
}

// getFilePath returns the file path for a hash
func (fs *FileSystemBackend) getFilePath(hash string) string {
	// Use first two characters as subdirectory for better distribution
	if len(hash) >= 2 {
		return filepath.Join(fs.basePath, hash[:2], hash)
	}
	return filepath.Join(fs.basePath, hash)
}

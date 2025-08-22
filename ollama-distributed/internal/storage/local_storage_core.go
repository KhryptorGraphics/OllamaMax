package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"log/slog"
)

// LocalStorage implements Storage interface for local filesystem storage
type LocalStorage struct {
	basePath string
	metaPath string
	logger   *slog.Logger

	// Configuration
	maxSize     int64
	compression bool
	encryption  bool

	// Caching and performance
	metadataCache map[string]*ObjectMetadata
	cacheMutex    sync.RWMutex
	cacheSize     int
	maxCacheSize  int

	// Statistics
	stats      *StorageStats
	statsMutex sync.RWMutex

	// File locks for concurrent access
	fileLocks  map[string]*sync.RWMutex
	locksMutex sync.RWMutex

	// Health monitoring
	lastHealthCheck time.Time
	healthy         bool

	// Background tasks
	ctx     context.Context
	cancel  context.CancelFunc
	started bool
	mu      sync.RWMutex
}

// LocalStorageConfig contains configuration for local storage
type LocalStorageConfig struct {
	BasePath     string        `json:"base_path"`
	MaxSize      int64         `json:"max_size"`
	Compression  bool          `json:"compression"`
	Encryption   bool          `json:"encryption"`
	MaxCacheSize int           `json:"max_cache_size"`
	CleanupAge   time.Duration `json:"cleanup_age"`
	SyncWrites   bool          `json:"sync_writes"`
}

// NewLocalStorage creates a new local storage instance
func NewLocalStorage(config *LocalStorageConfig, logger *slog.Logger) (*LocalStorage, error) {
	if config.BasePath == "" {
		return nil, &StorageError{
			Code:    ErrCodeInvalidArgument,
			Message: "base path cannot be empty",
		}
	}

	if logger == nil {
		logger = slog.Default()
	}

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(config.BasePath, 0755); err != nil {
		return nil, &StorageError{
			Code:    ErrCodePermissionDenied,
			Message: fmt.Sprintf("failed to create base directory: %v", err),
		}
	}

	metaPath := filepath.Join(config.BasePath, ".metadata")
	if err := os.MkdirAll(metaPath, 0755); err != nil {
		return nil, &StorageError{
			Code:    ErrCodePermissionDenied,
			Message: fmt.Sprintf("failed to create metadata directory: %v", err),
		}
	}

	ls := &LocalStorage{
		basePath:      config.BasePath,
		metaPath:      metaPath,
		logger:        logger,
		maxSize:       config.MaxSize,
		compression:   config.Compression,
		encryption:    config.Encryption,
		metadataCache: make(map[string]*ObjectMetadata),
		maxCacheSize:  config.MaxCacheSize,
		fileLocks:     make(map[string]*sync.RWMutex),
		stats:         NewStorageStats(),
		healthy:       true,
	}

	return ls, nil
}

// Start starts the local storage
func (ls *LocalStorage) Start(ctx context.Context) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	if ls.started {
		return nil
	}

	ls.ctx, ls.cancel = context.WithCancel(ctx)
	ls.started = true

	// Start background tasks
	go ls.backgroundCleanup()
	go ls.backgroundHealthCheck()

	ls.logger.Info("local storage started", "base_path", ls.basePath)

	return nil
}

// Stop stops the local storage
func (ls *LocalStorage) Stop(ctx context.Context) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	if !ls.started {
		return nil
	}

	ls.cancel()
	ls.started = false

	ls.logger.Info("local storage stopped")
	return nil
}

// Close closes the local storage
func (ls *LocalStorage) Close() error {
	return ls.Stop(context.Background())
}

// Store stores an object in local storage
func (ls *LocalStorage) Store(ctx context.Context, key string, data io.Reader, metadata *ObjectMetadata) error {
	start := time.Now()
	defer func() {
		ls.updateLatencyStats("write", time.Since(start))
		ls.incrementOperationCount("store")
	}()

	if err := ls.validateKey(key); err != nil {
		return err
	}

	if metadata == nil {
		metadata = &ObjectMetadata{
			Key:       key,
			CreatedAt: time.Now(),
		}
	}

	// Get file lock
	fileLock := ls.getFileLock(key)
	fileLock.Lock()
	defer fileLock.Unlock()

	// Create object path
	objectPath := ls.getObjectPath(key)
	if err := os.MkdirAll(filepath.Dir(objectPath), 0755); err != nil {
		return &StorageError{
			Code:    ErrCodePermissionDenied,
			Message: fmt.Sprintf("failed to create object directory: %v", err),
		}
	}

	// Create temporary file
	tempPath := objectPath + ".tmp"
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return &StorageError{
			Code:    ErrCodePermissionDenied,
			Message: fmt.Sprintf("failed to create temporary file: %v", err),
		}
	}
	defer tempFile.Close()

	// Calculate hash while writing
	hasher := sha256.New()
	multiWriter := io.MultiWriter(tempFile, hasher)

	size, err := io.Copy(multiWriter, data)
	if err != nil {
		os.Remove(tempPath)
		return &StorageError{
			Code:    ErrCodeIOError,
			Message: fmt.Sprintf("failed to write data: %v", err),
		}
	}

	// Check size limits
	if ls.maxSize > 0 && size > ls.maxSize {
		os.Remove(tempPath)
		return &StorageError{
			Code:    ErrCodeQuotaExceeded,
			Message: fmt.Sprintf("object size %d exceeds maximum %d", size, ls.maxSize),
		}
	}

	// Sync to disk if configured
	if err := tempFile.Sync(); err != nil {
		os.Remove(tempPath)
		return &StorageError{
			Code:    ErrCodeIOError,
			Message: fmt.Sprintf("failed to sync file: %v", err),
		}
	}

	// Atomic rename
	if err := os.Rename(tempPath, objectPath); err != nil {
		os.Remove(tempPath)
		return &StorageError{
			Code:    ErrCodeIOError,
			Message: fmt.Sprintf("failed to rename file: %v", err),
		}
	}

	// Update metadata
	metadata.Size = size
	metadata.Hash = hex.EncodeToString(hasher.Sum(nil))
	metadata.UpdatedAt = time.Now()

	// Store metadata
	if err := ls.storeMetadata(key, metadata); err != nil {
		ls.logger.Warn("failed to store metadata", "key", key, "error", err)
	}

	// Update cache
	ls.updateCache(key, metadata)

	// Update statistics
	ls.incrementStorageSize(size)

	ls.logger.Debug("object stored", "key", key, "size", size, "hash", metadata.Hash)
	return nil
}

// Retrieve retrieves an object from local storage
func (ls *LocalStorage) Retrieve(ctx context.Context, key string) (io.ReadCloser, *ObjectMetadata, error) {
	start := time.Now()
	defer func() {
		ls.updateLatencyStats("read", time.Since(start))
		ls.incrementOperationCount("retrieve")
	}()

	if err := ls.validateKey(key); err != nil {
		return nil, nil, err
	}

	// Get file lock
	fileLock := ls.getFileLock(key)
	fileLock.RLock()
	defer fileLock.RUnlock()

	// Check if object exists
	objectPath := ls.getObjectPath(key)
	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		return nil, nil, &StorageError{
			Code:    ErrCodeNotFound,
			Message: fmt.Sprintf("object not found: %s", key),
		}
	}

	// Open file
	file, err := os.Open(objectPath)
	if err != nil {
		return nil, nil, &StorageError{
			Code:    ErrCodeIOError,
			Message: fmt.Sprintf("failed to open file: %v", err),
		}
	}

	// Get metadata
	metadata, err := ls.getMetadata(key)
	if err != nil {
		file.Close()
		return nil, nil, err
	}

	// Update access time
	metadata.AccessedAt = time.Now()
	go ls.storeMetadata(key, metadata) // Async update

	return file, metadata, nil
}

// Delete deletes an object from local storage
func (ls *LocalStorage) Delete(ctx context.Context, key string) error {
	start := time.Now()
	defer func() {
		ls.updateLatencyStats("delete", time.Since(start))
		ls.incrementOperationCount("delete")
	}()

	if err := ls.validateKey(key); err != nil {
		return err
	}

	// Get file lock
	fileLock := ls.getFileLock(key)
	fileLock.Lock()
	defer fileLock.Unlock()

	// Get metadata before deletion for statistics
	metadata, err := ls.getMetadata(key)
	if err != nil {
		if IsNotFoundError(err) {
			return nil // Already deleted
		}
		return err
	}

	// Delete object file
	objectPath := ls.getObjectPath(key)
	if err := os.Remove(objectPath); err != nil && !os.IsNotExist(err) {
		return &StorageError{
			Code:    ErrCodeIOError,
			Message: fmt.Sprintf("failed to delete object file: %v", err),
		}
	}

	// Delete metadata file
	metadataPath := ls.getMetadataPath(key)
	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		ls.logger.Warn("failed to delete metadata file", "key", key, "error", err)
	}

	// Remove from cache
	ls.removeFromCache(key)

	// Update statistics
	ls.decrementStorageSize(metadata.Size)

	ls.logger.Debug("object deleted", "key", key, "size", metadata.Size)
	return nil
}

// Exists checks if an object exists in local storage
func (ls *LocalStorage) Exists(ctx context.Context, key string) (bool, error) {
	if err := ls.validateKey(key); err != nil {
		return false, err
	}

	objectPath := ls.getObjectPath(key)
	_, err := os.Stat(objectPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, &StorageError{
			Code:    ErrCodeIOError,
			Message: fmt.Sprintf("failed to check file existence: %v", err),
		}
	}

	return true, nil
}

// getObjectPath returns the filesystem path for an object
func (ls *LocalStorage) getObjectPath(key string) string {
	// Use first two characters of key for directory sharding
	if len(key) >= 2 {
		return filepath.Join(ls.basePath, key[:2], key)
	}
	return filepath.Join(ls.basePath, key)
}

// getMetadataPath returns the filesystem path for metadata
func (ls *LocalStorage) getMetadataPath(key string) string {
	return filepath.Join(ls.metaPath, key+".meta")
}

// getFileLock returns a file lock for the given key
func (ls *LocalStorage) getFileLock(key string) *sync.RWMutex {
	ls.locksMutex.Lock()
	defer ls.locksMutex.Unlock()

	if lock, exists := ls.fileLocks[key]; exists {
		return lock
	}

	lock := &sync.RWMutex{}
	ls.fileLocks[key] = lock
	return lock
}

// validateKey validates an object key
func (ls *LocalStorage) validateKey(key string) error {
	if key == "" {
		return &StorageError{
			Code:    ErrCodeInvalidArgument,
			Message: "key cannot be empty",
		}
	}

	if len(key) > 255 {
		return &StorageError{
			Code:    ErrCodeInvalidArgument,
			Message: "key too long (max 255 characters)",
		}
	}

	// Check for invalid characters
	for _, char := range key {
		if char < 32 || char == 127 {
			return &StorageError{
				Code:    ErrCodeInvalidArgument,
				Message: "key contains invalid characters",
			}
		}
	}

	return nil
}

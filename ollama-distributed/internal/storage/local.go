package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"log/slog"
)

// LocalStorage implements Storage interface for local filesystem storage
type LocalStorage struct {
	basePath    string
	metaPath    string
	logger      *slog.Logger
	
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
	stats       *StorageStats
	statsMutex  sync.RWMutex
	
	// File locks for concurrent access
	fileLocks   map[string]*sync.RWMutex
	locksMutex  sync.RWMutex
	
	// Health monitoring
	lastHealthCheck time.Time
	healthy         bool
	
	// Background tasks
	ctx         context.Context
	cancel      context.CancelFunc
	started     bool
	mu          sync.RWMutex
}

// LocalStorageConfig contains configuration for local storage
type LocalStorageConfig struct {
	BasePath      string `json:"base_path"`
	MaxSize       int64  `json:"max_size"`
	Compression   bool   `json:"compression"`
	Encryption    bool   `json:"encryption"`
	MaxCacheSize  int    `json:"max_cache_size"`
	CleanupAge    time.Duration `json:"cleanup_age"`
	SyncWrites    bool   `json:"sync_writes"`
}

// NewLocalStorage creates a new local storage instance
func NewLocalStorage(config *LocalStorageConfig, logger *slog.Logger) (*LocalStorage, error) {
	if config.BasePath == "" {
		return nil, &StorageError{
			Code:    ErrCodeInvalidArgument,
			Message: "base path cannot be empty",
		}
	}
	
	// Create base directories
	if err := os.MkdirAll(config.BasePath, 0755); err != nil {
		return nil, &StorageError{
			Code:    ErrCodeInternal,
			Message: "failed to create base directory",
			Cause:   err,
		}
	}
	
	metaPath := filepath.Join(config.BasePath, "metadata")
	if err := os.MkdirAll(metaPath, 0755); err != nil {
		return nil, &StorageError{
			Code:    ErrCodeInternal,
			Message: "failed to create metadata directory",
			Cause:   err,
		}
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
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
		stats: &StorageStats{
			OperationCounts: make(map[string]int64),
			Performance: &PerformanceStats{
				ReadLatency:   &LatencyStats{},
				WriteLatency:  &LatencyStats{},
				DeleteLatency: &LatencyStats{},
				Throughput:    &Throughput{},
			},
		},
		ctx:     ctx,
		cancel:  cancel,
		healthy: true,
	}
	
	// Load existing metadata into cache
	if err := ls.loadMetadataCache(); err != nil {
		logger.Warn("failed to load metadata cache", "error", err)
	}
	
	return ls, nil
}

// Start starts the local storage
func (ls *LocalStorage) Start(ctx context.Context) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	
	if ls.started {
		return &StorageError{
			Code:    ErrCodeInternal,
			Message: "storage already started",
		}
	}
	
	// Start background cleanup routine
	go ls.cleanupRoutine()
	
	// Start health monitoring
	go ls.healthMonitorRoutine()
	
	// Start statistics collection
	go ls.statsCollectionRoutine()
	
	ls.started = true
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
	
	// Save metadata cache
	if err := ls.saveMetadataCache(); err != nil {
		ls.logger.Error("failed to save metadata cache", "error", err)
	}
	
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
	
	// Get file lock
	lock := ls.getFileLock(key)
	lock.Lock()
	defer lock.Unlock()
	
	// Create object path
	objPath := ls.getObjectPath(key)
	if err := os.MkdirAll(filepath.Dir(objPath), 0755); err != nil {
		return &StorageError{
			Code:      ErrCodeInternal,
			Message:   "failed to create object directory",
			Operation: "store",
			Key:       key,
			Cause:     err,
		}
	}
	
	// Create temporary file for atomic write
	tempPath := objPath + ".tmp"
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return &StorageError{
			Code:      ErrCodeInternal,
			Message:   "failed to create temporary file",
			Operation: "store",
			Key:       key,
			Cause:     err,
		}
	}
	defer os.Remove(tempPath) // Cleanup on error
	
	// Copy data and calculate hash
	hash := sha256.New()
	size, err := io.Copy(io.MultiWriter(tempFile, hash), data)
	if err != nil {
		tempFile.Close()
		return &StorageError{
			Code:      ErrCodeInternal,
			Message:   "failed to write data",
			Operation: "store",
			Key:       key,
			Cause:     err,
		}
	}
	
	if err := tempFile.Sync(); err != nil {
		tempFile.Close()
		return &StorageError{
			Code:      ErrCodeInternal,
			Message:   "failed to sync file",
			Operation: "store",
			Key:       key,
			Cause:     err,
		}
	}
	tempFile.Close()
	
	// Check size limits
	if ls.maxSize > 0 && size > ls.maxSize {
		return &StorageError{
			Code:      ErrCodeQuotaExceeded,
			Message:   "object size exceeds maximum allowed size",
			Operation: "store",
			Key:       key,
		}
	}
	
	// Prepare metadata
	now := time.Now()
	if metadata == nil {
		metadata = &ObjectMetadata{}
	}
	metadata.Key = key
	metadata.Size = size
	metadata.Hash = hex.EncodeToString(hash.Sum(nil))
	metadata.CreatedAt = now
	metadata.UpdatedAt = now
	metadata.AccessedAt = now
	
	// Atomic move
	if err := os.Rename(tempPath, objPath); err != nil {
		return &StorageError{
			Code:      ErrCodeInternal,
			Message:   "failed to move temporary file",
			Operation: "store",
			Key:       key,
			Cause:     err,
		}
	}
	
	// Store metadata
	if err := ls.storeMetadata(key, metadata); err != nil {
		// Try to cleanup object file
		os.Remove(objPath)
		return err
	}
	
	// Update cache
	ls.updateMetadataCache(key, metadata)
	
	// Update statistics
	ls.statsMutex.Lock()
	ls.stats.TotalObjects++
	ls.stats.TotalSize += size
	ls.stats.UsedSpace += size
	ls.statsMutex.Unlock()
	
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
	
	// Get metadata first
	metadata, err := ls.GetMetadata(ctx, key)
	if err != nil {
		return nil, nil, err
	}
	
	// Open object file
	objPath := ls.getObjectPath(key)
	file, err := os.Open(objPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, &StorageError{
				Code:      ErrCodeNotFound,
				Message:   "object not found",
				Operation: "retrieve",
				Key:       key,
			}
		}
		return nil, nil, &StorageError{
			Code:      ErrCodeInternal,
			Message:   "failed to open object file",
			Operation: "retrieve",
			Key:       key,
			Cause:     err,
		}
	}
	
	// Update access time
	metadata.AccessedAt = time.Now()
	ls.updateMetadataCache(key, metadata)
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
	lock := ls.getFileLock(key)
	lock.Lock()
	defer lock.Unlock()
	
	// Get metadata for size accounting
	metadata, err := ls.GetMetadata(ctx, key)
	if err != nil {
		if isNotFoundError(err) {
			return nil // Already deleted
		}
		return err
	}
	
	// Delete object file
	objPath := ls.getObjectPath(key)
	if err := os.Remove(objPath); err != nil && !os.IsNotExist(err) {
		return &StorageError{
			Code:      ErrCodeInternal,
			Message:   "failed to delete object file",
			Operation: "delete",
			Key:       key,
			Cause:     err,
		}
	}
	
	// Delete metadata
	metaPath := ls.getMetadataPath(key)
	if err := os.Remove(metaPath); err != nil && !os.IsNotExist(err) {
		ls.logger.Warn("failed to delete metadata file", "key", key, "error", err)
	}
	
	// Remove from cache
	ls.removeFromMetadataCache(key)
	
	// Update statistics
	ls.statsMutex.Lock()
	ls.stats.TotalObjects--
	ls.stats.TotalSize -= metadata.Size
	ls.stats.UsedSpace -= metadata.Size
	ls.statsMutex.Unlock()
	
	ls.logger.Debug("object deleted", "key", key, "size", metadata.Size)
	return nil
}

// Exists checks if an object exists in local storage
func (ls *LocalStorage) Exists(ctx context.Context, key string) (bool, error) {
	if err := ls.validateKey(key); err != nil {
		return false, err
	}
	
	objPath := ls.getObjectPath(key)
	_, err := os.Stat(objPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, &StorageError{
			Code:      ErrCodeInternal,
			Message:   "failed to check object existence",
			Operation: "exists",
			Key:       key,
			Cause:     err,
		}
	}
	
	return true, nil
}

// GetMetadata retrieves metadata for an object
func (ls *LocalStorage) GetMetadata(ctx context.Context, key string) (*ObjectMetadata, error) {
	if err := ls.validateKey(key); err != nil {
		return nil, err
	}
	
	// Check cache first
	if metadata := ls.getFromMetadataCache(key); metadata != nil {
		return metadata, nil
	}
	
	// Load from disk
	metaPath := ls.getMetadataPath(key)
	data, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &StorageError{
				Code:      ErrCodeNotFound,
				Message:   "metadata not found",
				Operation: "get_metadata",
				Key:       key,
			}
		}
		return nil, &StorageError{
			Code:      ErrCodeInternal,
			Message:   "failed to read metadata file",
			Operation: "get_metadata",
			Key:       key,
			Cause:     err,
		}
	}
	
	var metadata ObjectMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, &StorageError{
			Code:      ErrCodeInternal,
			Message:   "failed to parse metadata",
			Operation: "get_metadata",
			Key:       key,
			Cause:     err,
		}
	}
	
	// Update cache
	ls.updateMetadataCache(key, &metadata)
	
	return &metadata, nil
}

// SetMetadata sets metadata for an object
func (ls *LocalStorage) SetMetadata(ctx context.Context, key string, metadata *ObjectMetadata) error {
	if err := ls.validateKey(key); err != nil {
		return err
	}
	
	metadata.Key = key
	metadata.UpdatedAt = time.Now()
	
	if err := ls.storeMetadata(key, metadata); err != nil {
		return err
	}
	
	ls.updateMetadataCache(key, metadata)
	return nil
}

// UpdateMetadata updates specific metadata fields
func (ls *LocalStorage) UpdateMetadata(ctx context.Context, key string, updates map[string]interface{}) error {
	metadata, err := ls.GetMetadata(ctx, key)
	if err != nil {
		return err
	}
	
	// Apply updates
	if metadata.Attributes == nil {
		metadata.Attributes = make(map[string]interface{})
	}
	
	for field, value := range updates {
		switch field {
		case "content_type":
			if v, ok := value.(string); ok {
				metadata.ContentType = v
			}
		case "version":
			if v, ok := value.(string); ok {
				metadata.Version = v
			}
		default:
			metadata.Attributes[field] = value
		}
	}
	
	return ls.SetMetadata(ctx, key, metadata)
}

// BatchStore performs batch store operations
func (ls *LocalStorage) BatchStore(ctx context.Context, operations []BatchStoreOperation) error {
	var errors []error
	
	for _, op := range operations {
		if err := ls.Store(ctx, op.Key, op.Data, op.Metadata); err != nil {
			errors = append(errors, fmt.Errorf("failed to store %s: %w", op.Key, err))
		}
	}
	
	if len(errors) > 0 {
		return &StorageError{
			Code:    ErrCodeInternal,
			Message: fmt.Sprintf("batch store failed with %d errors", len(errors)),
		}
	}
	
	return nil
}

// BatchDelete performs batch delete operations
func (ls *LocalStorage) BatchDelete(ctx context.Context, keys []string) error {
	var errors []error
	
	for _, key := range keys {
		if err := ls.Delete(ctx, key); err != nil {
			errors = append(errors, fmt.Errorf("failed to delete %s: %w", key, err))
		}
	}
	
	if len(errors) > 0 {
		return &StorageError{
			Code:    ErrCodeInternal,
			Message: fmt.Sprintf("batch delete failed with %d errors", len(errors)),
		}
	}
	
	return nil
}

// List lists objects with optional prefix and pagination
func (ls *LocalStorage) List(ctx context.Context, prefix string, options *ListOptions) (*ListResult, error) {
	if options == nil {
		options = &ListOptions{Limit: 1000}
	}
	
	var items []*ObjectMetadata
	
	// Walk the metadata directory
	err := filepath.Walk(ls.metaPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			return nil
		}
		
		// Extract key from path
		relPath, err := filepath.Rel(ls.metaPath, path)
		if err != nil {
			return err
		}
		
		key := strings.ReplaceAll(relPath, string(filepath.Separator), "/")
		key = strings.TrimSuffix(key, ".meta")
		
		// Check prefix filter
		if prefix != "" && !strings.HasPrefix(key, prefix) {
			return nil
		}
		
		// Load metadata
		metadata, err := ls.GetMetadata(ctx, key)
		if err != nil {
			ls.logger.Warn("failed to load metadata during list", "key", key, "error", err)
			return nil
		}
		
		items = append(items, metadata)
		
		// Check limit
		if options.Limit > 0 && len(items) >= options.Limit {
			return filepath.SkipDir
		}
		
		return nil
	})
	
	if err != nil {
		return nil, &StorageError{
			Code:      ErrCodeInternal,
			Message:   "failed to list objects",
			Operation: "list",
			Cause:     err,
		}
	}
	
	// Sort results
	if options.SortBy == "name" {
		sort.Slice(items, func(i, j int) bool {
			if options.SortOrder == "desc" {
				return items[i].Key > items[j].Key
			}
			return items[i].Key < items[j].Key
		})
	} else if options.SortBy == "size" {
		sort.Slice(items, func(i, j int) bool {
			if options.SortOrder == "desc" {
				return items[i].Size > items[j].Size
			}
			return items[i].Size < items[j].Size
		})
	} else if options.SortBy == "modified" {
		sort.Slice(items, func(i, j int) bool {
			if options.SortOrder == "desc" {
				return items[i].UpdatedAt.After(items[j].UpdatedAt)
			}
			return items[i].UpdatedAt.Before(items[j].UpdatedAt)
		})
	}
	
	return &ListResult{
		Items:   items,
		Total:   int64(len(items)),
		HasMore: false,
	}, nil
}

// ListKeys lists object keys with optional prefix
func (ls *LocalStorage) ListKeys(ctx context.Context, prefix string) ([]string, error) {
	result, err := ls.List(ctx, prefix, &ListOptions{})
	if err != nil {
		return nil, err
	}
	
	keys := make([]string, len(result.Items))
	for i, item := range result.Items {
		keys[i] = item.Key
	}
	
	return keys, nil
}

// HealthCheck performs a health check on local storage
func (ls *LocalStorage) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	checks := make(map[string]CheckResult)
	healthy := true
	
	// Check disk space
	start := time.Now()
	stat, err := ls.getDiskUsage()
	if err != nil {
		checks["disk_space"] = CheckResult{
			Status:  "error",
			Message: "failed to get disk usage",
			Latency: time.Since(start).Milliseconds(),
			Time:    time.Now(),
		}
		healthy = false
	} else {
		status := "ok"
		message := fmt.Sprintf("%.1f%% used", float64(stat.Used)/float64(stat.Total)*100)
		
		if float64(stat.Used)/float64(stat.Total) > 0.9 {
			status = "warning"
			message = "disk space running low"
		}
		
		checks["disk_space"] = CheckResult{
			Status:  status,
			Message: message,
			Latency: time.Since(start).Milliseconds(),
			Time:    time.Now(),
		}
		
		if status != "ok" {
			healthy = false
		}
	}
	
	// Check write performance
	start = time.Now()
	testKey := "health_check_test"
	testData := strings.NewReader("health check test data")
	err = ls.Store(ctx, testKey, testData, nil)
	writeLatency := time.Since(start).Milliseconds()
	
	if err != nil {
		checks["write_test"] = CheckResult{
			Status:  "error",
			Message: "write test failed",
			Latency: writeLatency,
			Time:    time.Now(),
		}
		healthy = false
	} else {
		// Cleanup test object
		ls.Delete(ctx, testKey)
		
		status := "ok"
		message := "write test passed"
		if writeLatency > 1000 {
			status = "warning"
			message = "slow write performance"
		}
		
		checks["write_test"] = CheckResult{
			Status:  status,
			Message: message,
			Latency: writeLatency,
			Time:    time.Now(),
		}
	}
	
	ls.lastHealthCheck = time.Now()
	ls.healthy = healthy
	
	status := "healthy"
	if !healthy {
		status = "unhealthy"
	}
	
	return &HealthStatus{
		Status:    status,
		Healthy:   healthy,
		LastCheck: ls.lastHealthCheck,
		Checks:    checks,
	}, nil
}

// GetStats returns storage statistics
func (ls *LocalStorage) GetStats(ctx context.Context) (*StorageStats, error) {
	ls.statsMutex.RLock()
	defer ls.statsMutex.RUnlock()
	
	// Update disk usage
	diskStat, err := ls.getDiskUsage()
	if err != nil {
		ls.logger.Warn("failed to get disk usage", "error", err)
	} else {
		ls.stats.AvailableSpace = diskStat.Available
	}
	
	// Create a copy of stats
	stats := *ls.stats
	stats.OperationCounts = make(map[string]int64)
	for k, v := range ls.stats.OperationCounts {
		stats.OperationCounts[k] = v
	}
	
	// Copy performance stats
	if ls.stats.Performance != nil {
		perf := *ls.stats.Performance
		if ls.stats.Performance.ReadLatency != nil {
			readLat := *ls.stats.Performance.ReadLatency
			perf.ReadLatency = &readLat
		}
		if ls.stats.Performance.WriteLatency != nil {
			writeLat := *ls.stats.Performance.WriteLatency
			perf.WriteLatency = &writeLat
		}
		if ls.stats.Performance.DeleteLatency != nil {
			deleteLat := *ls.stats.Performance.DeleteLatency
			perf.DeleteLatency = &deleteLat
		}
		if ls.stats.Performance.Throughput != nil {
			throughput := *ls.stats.Performance.Throughput
			perf.Throughput = &throughput
		}
		stats.Performance = &perf
	}
	
	return &stats, nil
}

// Helper methods

func (ls *LocalStorage) validateKey(key string) error {
	if key == "" {
		return &StorageError{
			Code:    ErrCodeInvalidArgument,
			Message: "key cannot be empty",
		}
	}
	
	if strings.Contains(key, "..") {
		return &StorageError{
			Code:    ErrCodeInvalidArgument,
			Message: "key cannot contain '..'",
		}
	}
	
	return nil
}

func (ls *LocalStorage) getObjectPath(key string) string {
	// Create a safe path by replacing path separators
	safePath := strings.ReplaceAll(key, "/", string(filepath.Separator))
	return filepath.Join(ls.basePath, "objects", safePath)
}

func (ls *LocalStorage) getMetadataPath(key string) string {
	safePath := strings.ReplaceAll(key, "/", string(filepath.Separator))
	return filepath.Join(ls.metaPath, safePath+".meta")
}

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

func (ls *LocalStorage) storeMetadata(key string, metadata *ObjectMetadata) error {
	metaPath := ls.getMetadataPath(key)
	if err := os.MkdirAll(filepath.Dir(metaPath), 0755); err != nil {
		return &StorageError{
			Code:      ErrCodeInternal,
			Message:   "failed to create metadata directory",
			Operation: "store_metadata",
			Key:       key,
			Cause:     err,
		}
	}
	
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return &StorageError{
			Code:      ErrCodeInternal,
			Message:   "failed to serialize metadata",
			Operation: "store_metadata",
			Key:       key,
			Cause:     err,
		}
	}
	
	tempPath := metaPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return &StorageError{
			Code:      ErrCodeInternal,
			Message:   "failed to write metadata file",
			Operation: "store_metadata",
			Key:       key,
			Cause:     err,
		}
	}
	
	if err := os.Rename(tempPath, metaPath); err != nil {
		os.Remove(tempPath)
		return &StorageError{
			Code:      ErrCodeInternal,
			Message:   "failed to move metadata file",
			Operation: "store_metadata",
			Key:       key,
			Cause:     err,
		}
	}
	
	return nil
}

// Cache management methods

func (ls *LocalStorage) getFromMetadataCache(key string) *ObjectMetadata {
	ls.cacheMutex.RLock()
	defer ls.cacheMutex.RUnlock()
	
	if metadata, exists := ls.metadataCache[key]; exists {
		// Create a copy to avoid race conditions
		copy := *metadata
		return &copy
	}
	
	return nil
}

func (ls *LocalStorage) updateMetadataCache(key string, metadata *ObjectMetadata) {
	ls.cacheMutex.Lock()
	defer ls.cacheMutex.Unlock()
	
	// Check cache size and evict if necessary
	if len(ls.metadataCache) >= ls.maxCacheSize {
		ls.evictFromCache()
	}
	
	// Create a copy to store in cache
	copy := *metadata
	ls.metadataCache[key] = &copy
	ls.cacheSize++
}

func (ls *LocalStorage) removeFromMetadataCache(key string) {
	ls.cacheMutex.Lock()
	defer ls.cacheMutex.Unlock()
	
	if _, exists := ls.metadataCache[key]; exists {
		delete(ls.metadataCache, key)
		ls.cacheSize--
	}
}

func (ls *LocalStorage) evictFromCache() {
	// Simple LRU eviction - remove oldest accessed
	var oldestKey string
	var oldestTime time.Time = time.Now()
	
	for key, metadata := range ls.metadataCache {
		if metadata.AccessedAt.Before(oldestTime) {
			oldestTime = metadata.AccessedAt
			oldestKey = key
		}
	}
	
	if oldestKey != "" {
		delete(ls.metadataCache, oldestKey)
		ls.cacheSize--
	}
}

func (ls *LocalStorage) loadMetadataCache() error {
	// Load frequently accessed metadata into cache
	return filepath.Walk(ls.metaPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		
		if !strings.HasSuffix(path, ".meta") {
			return nil
		}
		
		// Extract key
		relPath, err := filepath.Rel(ls.metaPath, path)
		if err != nil {
			return err
		}
		key := strings.TrimSuffix(relPath, ".meta")
		key = strings.ReplaceAll(key, string(filepath.Separator), "/")
		
		// Load metadata
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		
		var metadata ObjectMetadata
		if err := json.Unmarshal(data, &metadata); err != nil {
			return err
		}
		
		// Add to cache if recently accessed
		if time.Since(metadata.AccessedAt) < 24*time.Hour {
			ls.cacheMutex.Lock()
			if len(ls.metadataCache) < ls.maxCacheSize {
				ls.metadataCache[key] = &metadata
				ls.cacheSize++
			}
			ls.cacheMutex.Unlock()
		}
		
		return nil
	})
}

func (ls *LocalStorage) saveMetadataCache() error {
	// No need to explicitly save as metadata is persisted on writes
	return nil
}

// Statistics and monitoring methods

func (ls *LocalStorage) updateLatencyStats(operation string, latency time.Duration) {
	ls.statsMutex.Lock()
	defer ls.statsMutex.Unlock()
	
	latencyMs := latency.Milliseconds()
	
	var stats *LatencyStats
	switch operation {
	case "read":
		stats = ls.stats.Performance.ReadLatency
	case "write":
		stats = ls.stats.Performance.WriteLatency
	case "delete":
		stats = ls.stats.Performance.DeleteLatency
	default:
		return
	}
	
	// Update statistics
	if stats.Samples == 0 {
		stats.Min = latencyMs
		stats.Max = latencyMs
		stats.Mean = float64(latencyMs)
		stats.Median = latencyMs
	} else {
		if latencyMs < stats.Min {
			stats.Min = latencyMs
		}
		if latencyMs > stats.Max {
			stats.Max = latencyMs
		}
		
		// Update mean
		stats.Mean = (stats.Mean*float64(stats.Samples) + float64(latencyMs)) / float64(stats.Samples+1)
	}
	
	stats.Samples++
}

func (ls *LocalStorage) incrementOperationCount(operation string) {
	ls.statsMutex.Lock()
	defer ls.statsMutex.Unlock()
	
	ls.stats.OperationCounts[operation]++
}

// Background routines

func (ls *LocalStorage) cleanupRoutine() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for {
		select {
		case <-ls.ctx.Done():
			return
		case <-ticker.C:
			ls.performCleanup()
		}
	}
}

func (ls *LocalStorage) healthMonitorRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ls.ctx.Done():
			return
		case <-ticker.C:
			ls.HealthCheck(ls.ctx)
		}
	}
}

func (ls *LocalStorage) statsCollectionRoutine() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ls.ctx.Done():
			return
		case <-ticker.C:
			ls.collectStats()
		}
	}
}

func (ls *LocalStorage) performCleanup() {
	// Clean up temporary files
	err := filepath.Walk(ls.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if strings.HasSuffix(path, ".tmp") && time.Since(info.ModTime()) > time.Hour {
			os.Remove(path)
		}
		
		return nil
	})
	
	if err != nil {
		ls.logger.Error("cleanup failed", "error", err)
	}
}

func (ls *LocalStorage) collectStats() {
	// Update throughput stats based on operation counts
	// This is a simplified implementation
	ls.statsMutex.Lock()
	defer ls.statsMutex.Unlock()
	
	// Calculate ops per second (simplified)
	if ls.stats.Performance != nil && ls.stats.Performance.Throughput != nil {
		totalOps := int64(0)
		for _, count := range ls.stats.OperationCounts {
			totalOps += count
		}
		
		// Simple calculation - in practice this would be more sophisticated
		ls.stats.Performance.Throughput.ReadOpsPerSec = float64(ls.stats.OperationCounts["retrieve"]) / 60.0
		ls.stats.Performance.Throughput.WriteOpsPerSec = float64(ls.stats.OperationCounts["store"]) / 60.0
		ls.stats.Performance.Throughput.DeleteOpsPerSec = float64(ls.stats.OperationCounts["delete"]) / 60.0
	}
}

// Disk usage calculation (platform-specific implementations would be needed)
type diskStat struct {
	Total     int64
	Used      int64
	Available int64
}

func (ls *LocalStorage) getDiskUsage() (*diskStat, error) {
	// This is a simplified implementation
	// In practice, this would use platform-specific syscalls
	return &diskStat{
		Total:     100 * 1024 * 1024 * 1024, // 100GB
		Used:      ls.stats.UsedSpace,
		Available: (100 * 1024 * 1024 * 1024) - ls.stats.UsedSpace,
	}, nil
}

// Helper function to check if error is not found
func isNotFoundError(err error) bool {
	if storageErr, ok := err.(*StorageError); ok {
		return storageErr.Code == ErrCodeNotFound
	}
	return false
}
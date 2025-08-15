package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// MetadataManager manages metadata storage and operations
type MetadataManager struct {
	logger *slog.Logger

	// Storage backends
	levelDB    *leveldb.DB
	fileSystem *FileSystemMetadata

	// Configuration
	config *MetadataConfig

	// Caching
	cache      map[string]*CachedMetadata
	cacheMutex sync.RWMutex
	cacheSize  int
	maxCache   int

	// Indexing
	indexes    map[string]*MetadataIndex
	indexMutex sync.RWMutex

	// Statistics
	stats      *MetadataStats
	statsMutex sync.RWMutex

	// Background tasks
	ctx     context.Context
	cancel  context.CancelFunc
	started bool
	mu      sync.RWMutex
}

// NewMetadataManager creates a new metadata manager
func NewMetadataManager(config *MetadataConfig, logger *slog.Logger) (*MetadataManager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	mm := &MetadataManager{
		logger:   logger,
		config:   config,
		cache:    make(map[string]*CachedMetadata),
		maxCache: config.CacheSize,
		indexes:  make(map[string]*MetadataIndex),
		stats:    NewMetadataStats(),
		ctx:      ctx,
		cancel:   cancel,
	}

	// Initialize backend
	if err := mm.initializeBackend(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize backend: %w", err)
	}

	// Create default indexes if search is enabled
	if config.EnableSearch {
		mm.createDefaultIndexes()
	}

	return mm, nil
}

// Start starts the metadata manager
func (mm *MetadataManager) Start(ctx context.Context) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if mm.started {
		return fmt.Errorf("metadata manager already started")
	}

	// Start background routines
	go mm.syncRoutine()
	go mm.compactionRoutine()
	go mm.cacheMaintenanceRoutine()
	go mm.statsCollectionRoutine()

	mm.started = true
	return nil
}

// Stop stops the metadata manager
func (mm *MetadataManager) Stop(ctx context.Context) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if !mm.started {
		return nil
	}

	mm.cancel()

	// Close backend
	if mm.levelDB != nil {
		mm.levelDB.Close()
	}

	mm.started = false
	return nil
}

// Store stores metadata for an object
func (mm *MetadataManager) Store(ctx context.Context, key string, metadata *ObjectMetadata) error {
	start := time.Now()
	defer func() {
		mm.updateLatencyStats("set", time.Since(start))
		mm.incrementOperationCount("store")
	}()

	// Validate input
	if key == "" {
		return NewStorageError(ErrCodeInvalidKey, "key cannot be empty", key)
	}
	if metadata == nil {
		return NewStorageError(ErrCodeInvalidValue, "metadata cannot be nil", key)
	}

	// Set metadata key and timestamps
	metadata.Key = key
	metadata.UpdatedAt = time.Now()
	if metadata.CreatedAt.IsZero() {
		metadata.CreatedAt = metadata.UpdatedAt
	}

	// Store in backend
	if err := mm.storeInBackend(key, metadata); err != nil {
		return fmt.Errorf("failed to store in backend: %w", err)
	}

	// Update cache
	mm.updateCache(key, metadata)

	// Update indexes
	if mm.config.EnableSearch {
		mm.updateIndexes(key, metadata)
	}

	// Update statistics
	mm.statsMutex.Lock()
	mm.stats.TotalObjects++
	mm.stats.TotalSize += metadata.Size
	mm.statsMutex.Unlock()

	return nil
}

// Get retrieves metadata for an object
func (mm *MetadataManager) Get(ctx context.Context, key string) (*ObjectMetadata, error) {
	start := time.Now()
	defer func() {
		mm.updateLatencyStats("get", time.Since(start))
		mm.incrementOperationCount("get")
	}()

	// Check cache first
	if cached := mm.getFromCache(key); cached != nil {
		mm.incrementCacheHits()
		cached.Metadata.AccessedAt = time.Now()
		return cached.Metadata, nil
	}

	// Load from backend
	metadata, err := mm.loadFromBackend(key)
	if err != nil {
		mm.incrementCacheMisses()
		return nil, err
	}

	// Update cache
	mm.updateCache(key, metadata)
	mm.incrementCacheMisses()

	return metadata, nil
}

// Update updates specific metadata fields
func (mm *MetadataManager) Update(ctx context.Context, key string, updates map[string]interface{}) error {
	// Get current metadata
	metadata, err := mm.Get(ctx, key)
	if err != nil {
		return err
	}

	// Apply updates
	for field, value := range updates {
		switch field {
		case "attributes":
			if attrs, ok := value.(map[string]interface{}); ok {
				metadata.Attributes = attrs
			}
		case "content_type":
			if contentType, ok := value.(string); ok {
				metadata.ContentType = contentType
			}
		case "version":
			if version, ok := value.(string); ok {
				metadata.Version = version
			}
		default:
			return NewStorageError(ErrCodeInvalidValue, fmt.Sprintf("unsupported field: %s", field), key)
		}
	}

	return mm.Store(ctx, key, metadata)
}

// Delete deletes metadata for an object
func (mm *MetadataManager) Delete(ctx context.Context, key string) error {
	start := time.Now()
	defer func() {
		mm.updateLatencyStats("delete", time.Since(start))
		mm.incrementOperationCount("delete")
	}()

	// Get metadata for index removal
	metadata, err := mm.Get(ctx, key)
	if err != nil && !IsNotFound(err) {
		return err
	}

	// Remove from backend
	if err := mm.deleteFromBackend(key); err != nil {
		return fmt.Errorf("failed to delete from backend: %w", err)
	}

	// Remove from cache
	mm.removeFromCache(key)

	// Remove from indexes
	if metadata != nil && mm.config.EnableSearch {
		mm.removeFromIndexes(key, metadata)
	}

	// Update statistics
	if metadata != nil {
		mm.statsMutex.Lock()
		mm.stats.TotalObjects--
		mm.stats.TotalSize -= metadata.Size
		mm.statsMutex.Unlock()
	}

	return nil
}

// List lists metadata with optional filtering
func (mm *MetadataManager) List(ctx context.Context, prefix string, options *ListOptions) ([]*ObjectMetadata, error) {
	start := time.Now()
	defer func() {
		mm.updateLatencyStats("list", time.Since(start))
		mm.incrementOperationCount("list")
	}()

	return mm.listFromBackend(prefix, options)
}

// GetStats returns metadata statistics
func (mm *MetadataManager) GetStats(ctx context.Context) (*MetadataStats, error) {
	mm.statsMutex.RLock()
	defer mm.statsMutex.RUnlock()

	// Update cache hit rate
	totalRequests := mm.stats.CacheHits + mm.stats.CacheMisses
	if totalRequests > 0 {
		mm.stats.CacheHitRate = float64(mm.stats.CacheHits) / float64(totalRequests)
	}

	// Create a copy to avoid race conditions
	stats := *mm.stats
	if stats.OperationCounts == nil {
		stats.OperationCounts = make(map[string]int64)
	}

	return &stats, nil
}

// Backend initialization methods

func (mm *MetadataManager) initializeBackend() error {
	switch mm.config.Backend {
	case "leveldb":
		return mm.initializeLevelDB()
	case "filesystem":
		return mm.initializeFileSystem()
	case "memory":
		return nil // Memory backend doesn't need initialization
	default:
		return fmt.Errorf("unsupported metadata backend: %s", mm.config.Backend)
	}
}

func (mm *MetadataManager) initializeLevelDB() error {
	dbPath := filepath.Join(mm.config.DataDir, "metadata.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return err
	}

	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return err
	}

	mm.levelDB = db
	return nil
}

func (mm *MetadataManager) initializeFileSystem() error {
	metaPath := filepath.Join(mm.config.DataDir, "metadata")
	if err := os.MkdirAll(metaPath, 0755); err != nil {
		return err
	}

	mm.fileSystem = &FileSystemMetadata{
		basePath: metaPath,
		logger:   mm.logger,
	}

	return nil
}

// Backend operations

func (mm *MetadataManager) storeInBackend(key string, metadata *ObjectMetadata) error {
	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	switch mm.config.Backend {
	case "leveldb":
		return mm.levelDB.Put([]byte(key), data, nil)
	case "filesystem":
		return mm.fileSystem.store(key, data)
	case "memory":
		return nil // Memory backend handled by cache
	default:
		return fmt.Errorf("unsupported backend: %s", mm.config.Backend)
	}
}

func (mm *MetadataManager) loadFromBackend(key string) (*ObjectMetadata, error) {
	var data []byte
	var err error

	switch mm.config.Backend {
	case "leveldb":
		data, err = mm.levelDB.Get([]byte(key), nil)
		if err != nil {
			if err == leveldb.ErrNotFound {
				return nil, NewStorageError(ErrCodeNotFound, "metadata not found", key)
			}
			return nil, err
		}
	case "filesystem":
		data, err = mm.fileSystem.load(key)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, NewStorageError(ErrCodeNotFound, "metadata not found", key)
			}
			return nil, err
		}
	case "memory":
		return nil, NewStorageError(ErrCodeNotFound, "metadata not found", key)
	default:
		return nil, fmt.Errorf("unsupported backend: %s", mm.config.Backend)
	}

	var metadata ObjectMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func (mm *MetadataManager) deleteFromBackend(key string) error {
	switch mm.config.Backend {
	case "leveldb":
		return mm.levelDB.Delete([]byte(key), nil)
	case "filesystem":
		return mm.fileSystem.delete(key)
	case "memory":
		return nil // Memory backend handled by cache
	default:
		return fmt.Errorf("unsupported backend: %s", mm.config.Backend)
	}
}

func (mm *MetadataManager) listFromBackend(prefix string, options *ListOptions) ([]*ObjectMetadata, error) {
	var results []*ObjectMetadata

	switch mm.config.Backend {
	case "leveldb":
		iter := mm.levelDB.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
		defer iter.Release()

		count := 0

		for iter.Next() {
			var metadata ObjectMetadata
			if err := json.Unmarshal(iter.Value(), &metadata); err != nil {
				continue
			}

			results = append(results, &metadata)
			count++

			if options != nil && options.Limit > 0 && count >= options.Limit {
				break
			}
		}

		return results, iter.Error()
	case "filesystem":
		return mm.fileSystem.list(prefix, options)
	case "memory":
		// List from cache for memory backend
		mm.cacheMutex.RLock()
		defer mm.cacheMutex.RUnlock()

		for key, cached := range mm.cache {
			if strings.HasPrefix(key, prefix) {
				results = append(results, cached.Metadata)
			}
		}
		return results, nil
	default:
		return nil, fmt.Errorf("unsupported backend: %s", mm.config.Backend)
	}
}

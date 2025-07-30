package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
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

// MetadataConfig contains configuration for metadata management
type MetadataConfig struct {
	Backend       string        `json:"backend"`        // leveldb, filesystem, memory
	DataDir       string        `json:"data_dir"`
	IndexingMode  string        `json:"indexing_mode"`  // eager, lazy, disabled
	CacheSize     int           `json:"cache_size"`
	SyncInterval  time.Duration `json:"sync_interval"`
	CompactInterval time.Duration `json:"compact_interval"`
	EnableSearch  bool          `json:"enable_search"`
	EnableVersioning bool       `json:"enable_versioning"`
}

// FileSystemMetadata implements filesystem-based metadata storage
type FileSystemMetadata struct {
	basePath string
	logger   *slog.Logger
}

// CachedMetadata represents cached metadata with additional information
type CachedMetadata struct {
	Metadata   *ObjectMetadata `json:"metadata"`
	CachedAt   time.Time       `json:"cached_at"`
	AccessCount int            `json:"access_count"`
	LastAccess time.Time       `json:"last_access"`
}

// MetadataIndex represents an index for fast metadata queries
type MetadataIndex struct {
	Name      string                     `json:"name"`
	Type      string                     `json:"type"` // btree, hash, text
	Fields    []string                   `json:"fields"`
	Values    map[string][]string        `json:"values"` // value -> keys
	CreatedAt time.Time                  `json:"created_at"`
	UpdatedAt time.Time                  `json:"updated_at"`
	Stats     *IndexStats                `json:"stats"`
	mutex     sync.RWMutex               `json:"-"` // Added for thread safety
}

// IndexStats contains statistics about an index
type IndexStats struct {
	TotalEntries  int64     `json:"total_entries"`
	UniqueValues  int64     `json:"unique_values"`
	LastUpdated   time.Time `json:"last_updated"`
	UpdateCount   int64     `json:"update_count"`
	QueryCount    int64     `json:"query_count"`
	AverageDepth  float64   `json:"average_depth"`
}

// MetadataStats contains statistics about metadata operations
type MetadataStats struct {
	TotalObjects     int64             `json:"total_objects"`
	TotalSize        int64             `json:"total_size"`
	CacheHitRate     float64           `json:"cache_hit_rate"`
	CacheHits        int64             `json:"cache_hits"`
	CacheMisses      int64             `json:"cache_misses"`
	IndexQueries     int64             `json:"index_queries"`
	OperationCounts  map[string]int64  `json:"operation_counts"`
	Performance      *MetadataPerformance `json:"performance"`
	LastCompaction   time.Time         `json:"last_compaction"`
	LastSync         time.Time         `json:"last_sync"`
}

// MetadataPerformance contains performance metrics
type MetadataPerformance struct {
	GetLatency    *LatencyStats `json:"get_latency"`
	SetLatency    *LatencyStats `json:"set_latency"`
	SearchLatency *LatencyStats `json:"search_latency"`
	IndexLatency  *LatencyStats `json:"index_latency"`
}

// MetadataQuery represents a metadata query
type MetadataQuery struct {
	Fields     map[string]interface{} `json:"fields"`
	Conditions []*QueryCondition      `json:"conditions"`
	Sort       *SortOptions           `json:"sort"`
	Limit      int                    `json:"limit"`
	Offset     int                    `json:"offset"`
	FullText   string                 `json:"full_text"`
}

// QueryCondition represents a query condition
type QueryCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, gt, lt, gte, lte, in, like, regex
	Value    interface{} `json:"value"`
	LogicalOp string     `json:"logical_op"` // and, or, not
}

// SortOptions represents sorting options
type SortOptions struct {
	Field string `json:"field"`
	Order string `json:"order"` // asc, desc
}

// MetadataQueryResult represents the result of a metadata query
type MetadataQueryResult struct {
	Objects      []*ObjectMetadata `json:"objects"`
	Total        int64             `json:"total"`
	QueryTime    time.Duration     `json:"query_time"`
	IndexUsed    string            `json:"index_used"`
	Explanation  string            `json:"explanation"`
}

// NewMetadataManager creates a new metadata manager
func NewMetadataManager(config *MetadataConfig, logger *slog.Logger) (*MetadataManager, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	mm := &MetadataManager{
		logger:     logger,
		config:     config,
		cache:      make(map[string]*CachedMetadata),
		maxCache:   config.CacheSize,
		indexes:    make(map[string]*MetadataIndex),
		stats: &MetadataStats{
			OperationCounts: make(map[string]int64),
			Performance: &MetadataPerformance{
				GetLatency:    &LatencyStats{},
				SetLatency:    &LatencyStats{},
				SearchLatency: &LatencyStats{},
				IndexLatency:  &LatencyStats{},
			},
		},
		ctx:    ctx,
		cancel: cancel,
	}
	
	// Initialize storage backend
	if err := mm.initializeBackend(); err != nil {
		return nil, fmt.Errorf("failed to initialize metadata backend: %w", err)
	}
	
	// Create default indexes
	if config.IndexingMode != "disabled" {
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
	mm.logger.Info("metadata manager started", "backend", mm.config.Backend)
	
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
	
	// Close storage backends
	if mm.levelDB != nil {
		mm.levelDB.Close()
	}
	
	mm.started = false
	mm.logger.Info("metadata manager stopped")
	
	return nil
}

// Store stores metadata for an object
func (mm *MetadataManager) Store(ctx context.Context, key string, metadata *ObjectMetadata) error {
	start := time.Now()
	defer func() {
		mm.updateLatencyStats("set", time.Since(start))
		mm.incrementOperationCount("store")
	}()
	
	// Prepare metadata
	metadata.Key = key
	metadata.UpdatedAt = time.Now()
	if metadata.CreatedAt.IsZero() {
		metadata.CreatedAt = time.Now()
	}
	
	// Store in backend
	if err := mm.storeInBackend(key, metadata); err != nil {
		return err
	}
	
	// Update cache
	mm.updateCache(key, metadata)
	
	// Update indexes
	if mm.config.IndexingMode == "eager" {
		mm.updateIndexes(key, metadata)
	}
	
	// Update statistics
	mm.statsMutex.Lock()
	mm.stats.TotalObjects++
	if metadata.Size > 0 {
		mm.stats.TotalSize += metadata.Size
	}
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
		return cached.Metadata, nil
	}
	
	mm.incrementCacheMisses()
	
	// Load from backend
	metadata, err := mm.loadFromBackend(key)
	if err != nil {
		return nil, err
	}
	
	// Update cache
	mm.updateCache(key, metadata)
	
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
		case "size":
			if v, ok := value.(int64); ok {
				metadata.Size = v
			}
		case "hash":
			if v, ok := value.(string); ok {
				metadata.Hash = v
			}
		default:
			metadata.Attributes[field] = value
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
	
	// Get metadata for size accounting
	metadata, err := mm.Get(ctx, key)
	if err != nil && !isNotFoundError(err) {
		return err
	}
	
	// Delete from backend
	if err := mm.deleteFromBackend(key); err != nil {
		return err
	}
	
	// Remove from cache
	mm.removeFromCache(key)
	
	// Remove from indexes
	if mm.config.IndexingMode != "disabled" {
		mm.removeFromIndexes(key, metadata)
	}
	
	// Update statistics
	if metadata != nil {
		mm.statsMutex.Lock()
		mm.stats.TotalObjects--
		if metadata.Size > 0 {
			mm.stats.TotalSize -= metadata.Size
		}
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

// Search performs advanced metadata search
func (mm *MetadataManager) Search(ctx context.Context, query *MetadataQuery) (*MetadataQueryResult, error) {
	start := time.Now()
	defer func() {
		mm.updateLatencyStats("search", time.Since(start))
		mm.incrementOperationCount("search")
	}()
	
	if !mm.config.EnableSearch {
		return nil, &StorageError{
			Code:    ErrCodeUnavailable,
			Message: "search not enabled",
		}
	}
	
	// Determine best index to use
	indexName := mm.selectBestIndex(query)
	
	var results []*ObjectMetadata
	var err error
	
	if indexName != "" {
		results, err = mm.searchWithIndex(ctx, query, indexName)
	} else {
		results, err = mm.searchWithoutIndex(ctx, query)
	}
	
	if err != nil {
		return nil, err
	}
	
	// Apply sorting and pagination
	results = mm.applySortingAndPagination(results, query)
	
	return &MetadataQueryResult{
		Objects:     results,
		Total:       int64(len(results)),
		QueryTime:   time.Since(start),
		IndexUsed:   indexName,
		Explanation: mm.explainQuery(query, indexName),
	}, nil
}

// CreateIndex creates a new metadata index
func (mm *MetadataManager) CreateIndex(ctx context.Context, name string, fields []string, indexType string) error {
	start := time.Now()
	defer func() {
		mm.updateLatencyStats("index", time.Since(start))
		mm.incrementOperationCount("create_index")
	}()
	
	mm.indexMutex.Lock()
	defer mm.indexMutex.Unlock()
	
	if _, exists := mm.indexes[name]; exists {
		return &StorageError{
			Code:    ErrCodeAlreadyExists,
			Message: "index already exists",
		}
	}
	
	index := &MetadataIndex{
		Name:      name,
		Type:      indexType,
		Fields:    fields,
		Values:    make(map[string][]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Stats:     &IndexStats{},
	}
	
	mm.indexes[name] = index
	
	// Build index from existing metadata
	go mm.buildIndex(ctx, index)
	
	mm.logger.Info("metadata index created", "name", name, "fields", fields, "type", indexType)
	
	return nil
}

// DropIndex drops a metadata index
func (mm *MetadataManager) DropIndex(ctx context.Context, name string) error {
	mm.indexMutex.Lock()
	defer mm.indexMutex.Unlock()
	
	if _, exists := mm.indexes[name]; !exists {
		return &StorageError{
			Code:    ErrCodeNotFound,
			Message: "index not found",
		}
	}
	
	delete(mm.indexes, name)
	
	mm.logger.Info("metadata index dropped", "name", name)
	
	return nil
}

// GetIndexes returns all metadata indexes
func (mm *MetadataManager) GetIndexes(ctx context.Context) ([]*MetadataIndex, error) {
	mm.indexMutex.RLock()
	defer mm.indexMutex.RUnlock()
	
	indexes := make([]*MetadataIndex, 0, len(mm.indexes))
	for _, index := range mm.indexes {
		// Create a copy
		indexCopy := *index
		indexCopy.Values = make(map[string][]string)
		for k, v := range index.Values {
			indexCopy.Values[k] = append([]string{}, v...)
		}
		indexes = append(indexes, &indexCopy)
	}
	
	return indexes, nil
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
	
	// Create a copy
	stats := *mm.stats
	stats.OperationCounts = make(map[string]int64)
	for k, v := range mm.stats.OperationCounts {
		stats.OperationCounts[k] = v
	}
	
	return &stats, nil
}

// Backend initialization

func (mm *MetadataManager) initializeBackend() error {
	switch mm.config.Backend {
	case "leveldb":
		return mm.initializeLevelDB()
	case "filesystem":
		return mm.initializeFileSystem()
	case "memory":
		return nil // Memory backend is the default cache
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
		return nil // Only stored in cache
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
		if err == leveldb.ErrNotFound {
			return nil, &StorageError{
				Code:    ErrCodeNotFound,
				Message: "metadata not found",
			}
		}
	case "filesystem":
		data, err = mm.fileSystem.load(key)
	case "memory":
		return nil, &StorageError{
			Code:    ErrCodeNotFound,
			Message: "metadata not found in memory",
		}
	default:
		return nil, fmt.Errorf("unsupported backend: %s", mm.config.Backend)
	}
	
	if err != nil {
		return nil, err
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
		return nil // Only deleted from cache
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
		
		for iter.Next() {
			var metadata ObjectMetadata
			if err := json.Unmarshal(iter.Value(), &metadata); err != nil {
				continue
			}
			results = append(results, &metadata)
			
			if options != nil && options.Limit > 0 && len(results) >= options.Limit {
				break
			}
		}
		
		return results, iter.Error()
		
	case "filesystem":
		return mm.fileSystem.list(prefix, options)
		
	case "memory":
		// List from cache
		mm.cacheMutex.RLock()
		defer mm.cacheMutex.RUnlock()
		
		for key, cached := range mm.cache {
			if strings.HasPrefix(key, prefix) {
				results = append(results, cached.Metadata)
				if options != nil && options.Limit > 0 && len(results) >= options.Limit {
					break
				}
			}
		}
		
		return results, nil
		
	default:
		return nil, fmt.Errorf("unsupported backend: %s", mm.config.Backend)
	}
}

// Cache management

func (mm *MetadataManager) getFromCache(key string) *CachedMetadata {
	mm.cacheMutex.RLock()
	defer mm.cacheMutex.RUnlock()
	
	if cached, exists := mm.cache[key]; exists {
		cached.LastAccess = time.Now()
		cached.AccessCount++
		return cached
	}
	
	return nil
}

func (mm *MetadataManager) updateCache(key string, metadata *ObjectMetadata) {
	mm.cacheMutex.Lock()
	defer mm.cacheMutex.Unlock()
	
	// Check cache size
	if len(mm.cache) >= mm.maxCache {
		mm.evictFromCache()
	}
	
	mm.cache[key] = &CachedMetadata{
		Metadata:    metadata,
		CachedAt:    time.Now(),
		AccessCount: 1,
		LastAccess:  time.Now(),
	}
	mm.cacheSize++
}

func (mm *MetadataManager) removeFromCache(key string) {
	mm.cacheMutex.Lock()
	defer mm.cacheMutex.Unlock()
	
	if _, exists := mm.cache[key]; exists {
		delete(mm.cache, key)
		mm.cacheSize--
	}
}

func (mm *MetadataManager) evictFromCache() {
	// LRU eviction
	var oldestKey string
	var oldestTime time.Time = time.Now()
	
	for key, cached := range mm.cache {
		if cached.LastAccess.Before(oldestTime) {
			oldestTime = cached.LastAccess
			oldestKey = key
		}
	}
	
	if oldestKey != "" {
		delete(mm.cache, oldestKey)
		mm.cacheSize--
	}
}

// Index management

func (mm *MetadataManager) createDefaultIndexes() {
	// Create indexes for common fields
	commonIndexes := map[string][]string{
		"size_index":         {"size"},
		"type_index":         {"content_type"},
		"created_index":      {"created_at"},
		"updated_index":      {"updated_at"},
		"hash_index":         {"hash"},
	}
	
	for name, fields := range commonIndexes {
		if err := mm.CreateIndex(context.Background(), name, fields, "btree"); err != nil {
			mm.logger.Warn("failed to create default index", "name", name, "error", err)
		}
	}
}

func (mm *MetadataManager) updateIndexes(key string, metadata *ObjectMetadata) {
	mm.indexMutex.Lock()
	defer mm.indexMutex.Unlock()
	
	for _, index := range mm.indexes {
		mm.updateIndex(index, key, metadata)
	}
}

func (mm *MetadataManager) updateIndex(index *MetadataIndex, key string, metadata *ObjectMetadata) {
	for _, field := range index.Fields {
		value := mm.extractFieldValue(metadata, field)
		if value != "" {
			if keys, exists := index.Values[value]; exists {
				// Check if key already exists
				found := false
				for _, k := range keys {
					if k == key {
						found = true
						break
					}
				}
				if !found {
					index.Values[value] = append(keys, key)
				}
			} else {
				index.Values[value] = []string{key}
			}
		}
	}
	
	index.UpdatedAt = time.Now()
	index.Stats.UpdateCount++
}

func (mm *MetadataManager) removeFromIndexes(key string, metadata *ObjectMetadata) {
	if metadata == nil {
		return
	}
	
	mm.indexMutex.Lock()
	defer mm.indexMutex.Unlock()
	
	for _, index := range mm.indexes {
		mm.removeFromIndex(index, key, metadata)
	}
}

func (mm *MetadataManager) removeFromIndex(index *MetadataIndex, key string, metadata *ObjectMetadata) {
	for _, field := range index.Fields {
		value := mm.extractFieldValue(metadata, field)
		if value != "" {
			if keys, exists := index.Values[value]; exists {
				// Remove key from slice
				var newKeys []string
				for _, k := range keys {
					if k != key {
						newKeys = append(newKeys, k)
					}
				}
				
				if len(newKeys) == 0 {
					delete(index.Values, value)
				} else {
					index.Values[value] = newKeys
				}
			}
		}
	}
	
	index.UpdatedAt = time.Now()
}

func (mm *MetadataManager) buildIndex(ctx context.Context, index *MetadataIndex) {
	mm.logger.Info("building metadata index", "name", index.Name)
	
	// Get all metadata from backend
	allMetadata, err := mm.listFromBackend("", nil)
	if err != nil {
		mm.logger.Error("failed to load metadata for index building", "error", err)
		return
	}
	
	// Build index
	for _, metadata := range allMetadata {
		mm.updateIndex(index, metadata.Key, metadata)
	}
	
	index.Stats.TotalEntries = int64(len(allMetadata))
	index.Stats.UniqueValues = int64(len(index.Values))
	index.Stats.LastUpdated = time.Now()
	
	mm.logger.Info("metadata index built", "name", index.Name, "entries", index.Stats.TotalEntries)
}

// Search implementation

func (mm *MetadataManager) selectBestIndex(query *MetadataQuery) string {
	mm.indexMutex.RLock()
	defer mm.indexMutex.RUnlock()
	
	var bestIndex string
	var bestScore int
	
	for name, index := range mm.indexes {
		score := mm.calculateIndexScore(index, query)
		if score > bestScore {
			bestScore = score
			bestIndex = name
		}
	}
	
	return bestIndex
}

func (mm *MetadataManager) calculateIndexScore(index *MetadataIndex, query *MetadataQuery) int {
	score := 0
	
	// Check if index fields match query conditions
	for _, condition := range query.Conditions {
		for _, field := range index.Fields {
			if field == condition.Field {
				score += 10
				
				// Bonus for exact match operators
				if condition.Operator == "eq" {
					score += 5
				}
			}
		}
	}
	
	// Check sorting field
	if query.Sort != nil {
		for _, field := range index.Fields {
			if field == query.Sort.Field {
				score += 3
			}
		}
	}
	
	return score
}

func (mm *MetadataManager) searchWithIndex(ctx context.Context, query *MetadataQuery, indexName string) ([]*ObjectMetadata, error) {
	mm.indexMutex.RLock()
	index, exists := mm.indexes[indexName]
	mm.indexMutex.RUnlock()
	
	if !exists {
		return mm.searchWithoutIndex(ctx, query)
	}
	
	var candidateKeys []string
	
	// Find candidate keys using index
	for _, condition := range query.Conditions {
		if mm.isFieldIndexed(index, condition.Field) {
			keys := mm.getKeysFromIndex(index, condition)
			if candidateKeys == nil {
				candidateKeys = keys
			} else {
				candidateKeys = mm.intersectKeys(candidateKeys, keys)
			}
		}
	}
	
	// Load metadata for candidate keys
	var results []*ObjectMetadata
	for _, key := range candidateKeys {
		metadata, err := mm.Get(ctx, key)
		if err != nil {
			continue
		}
		
		// Apply additional filtering
		if mm.matchesQuery(metadata, query) {
			results = append(results, metadata)
		}
	}
	
	index.Stats.QueryCount++
	
	return results, nil
}

func (mm *MetadataManager) searchWithoutIndex(ctx context.Context, query *MetadataQuery) ([]*ObjectMetadata, error) {
	// Full scan
	allMetadata, err := mm.listFromBackend("", nil)
	if err != nil {
		return nil, err
	}
	
	var results []*ObjectMetadata
	for _, metadata := range allMetadata {
		if mm.matchesQuery(metadata, query) {
			results = append(results, metadata)
		}
	}
	
	return results, nil
}

func (mm *MetadataManager) applySortingAndPagination(results []*ObjectMetadata, query *MetadataQuery) []*ObjectMetadata {
	// Apply sorting
	if query.Sort != nil {
		sort.Slice(results, func(i, j int) bool {
			return mm.compareMetadata(results[i], results[j], query.Sort)
		})
	}
	
	// Apply pagination
	start := query.Offset
	if start < 0 {
		start = 0
	}
	if start >= len(results) {
		return []*ObjectMetadata{}
	}
	
	end := start + query.Limit
	if query.Limit <= 0 || end > len(results) {
		end = len(results)
	}
	
	return results[start:end]
}

// Helper methods

func (mm *MetadataManager) extractFieldValue(metadata *ObjectMetadata, field string) string {
	switch field {
	case "key":
		return metadata.Key
	case "size":
		return fmt.Sprintf("%d", metadata.Size)
	case "content_type":
		return metadata.ContentType
	case "hash":
		return metadata.Hash
	case "version":
		return metadata.Version
	case "created_at":
		return metadata.CreatedAt.Format(time.RFC3339)
	case "updated_at":
		return metadata.UpdatedAt.Format(time.RFC3339)
	case "accessed_at":
		return metadata.AccessedAt.Format(time.RFC3339)
	default:
		if metadata.Attributes != nil {
			if value, exists := metadata.Attributes[field]; exists {
				return fmt.Sprintf("%v", value)
			}
		}
		return ""
	}
}

func (mm *MetadataManager) isFieldIndexed(index *MetadataIndex, field string) bool {
	for _, indexField := range index.Fields {
		if indexField == field {
			return true
		}
	}
	return false
}

func (mm *MetadataManager) getKeysFromIndex(index *MetadataIndex, condition *QueryCondition) []string {
	value := fmt.Sprintf("%v", condition.Value)
	
	switch condition.Operator {
	case "eq":
		if keys, exists := index.Values[value]; exists {
			return keys
		}
		return []string{}
	case "ne":
		var allKeys []string
		for v, keys := range index.Values {
			if v != value {
				allKeys = append(allKeys, keys...)
			}
		}
		return allKeys
	default:
		// For other operators, return all keys for full evaluation
		var allKeys []string
		for _, keys := range index.Values {
			allKeys = append(allKeys, keys...)
		}
		return allKeys
	}
}

func (mm *MetadataManager) intersectKeys(keys1, keys2 []string) []string {
	keyMap := make(map[string]bool)
	for _, key := range keys1 {
		keyMap[key] = true
	}
	
	var result []string
	for _, key := range keys2 {
		if keyMap[key] {
			result = append(result, key)
		}
	}
	
	return result
}

func (mm *MetadataManager) matchesQuery(metadata *ObjectMetadata, query *MetadataQuery) bool {
	for _, condition := range query.Conditions {
		if !mm.matchesCondition(metadata, condition) {
			return false
		}
	}
	
	// TODO: Implement full-text search
	if query.FullText != "" {
		// Simplified full-text search
		searchText := strings.ToLower(query.FullText)
		if !strings.Contains(strings.ToLower(metadata.Key), searchText) &&
		   !strings.Contains(strings.ToLower(metadata.ContentType), searchText) {
			return false
		}
	}
	
	return true
}

func (mm *MetadataManager) matchesCondition(metadata *ObjectMetadata, condition *QueryCondition) bool {
	fieldValue := mm.extractFieldValue(metadata, condition.Field)
	conditionValue := fmt.Sprintf("%v", condition.Value)
	
	switch condition.Operator {
	case "eq":
		return fieldValue == conditionValue
	case "ne":
		return fieldValue != conditionValue
	case "like":
		return strings.Contains(strings.ToLower(fieldValue), strings.ToLower(conditionValue))
	case "gt", "gte", "lt", "lte":
		// Numeric comparison (simplified)
		return mm.compareValues(fieldValue, conditionValue, condition.Operator)
	case "in":
		// TODO: Implement IN operator
		return false
	default:
		return false
	}
}

func (mm *MetadataManager) compareValues(value1, value2, operator string) bool {
	// Simplified comparison (would need proper type handling)
	switch operator {
	case "gt":
		return value1 > value2
	case "gte":
		return value1 >= value2
	case "lt":
		return value1 < value2
	case "lte":
		return value1 <= value2
	default:
		return false
	}
}

func (mm *MetadataManager) compareMetadata(m1, m2 *ObjectMetadata, sort *SortOptions) bool {
	value1 := mm.extractFieldValue(m1, sort.Field)
	value2 := mm.extractFieldValue(m2, sort.Field)
	
	if sort.Order == "desc" {
		return value1 > value2
	}
	return value1 < value2
}

func (mm *MetadataManager) explainQuery(query *MetadataQuery, indexUsed string) string {
	if indexUsed != "" {
		return fmt.Sprintf("Used index: %s", indexUsed)
	}
	return "Full table scan"
}

// Statistics methods

func (mm *MetadataManager) updateLatencyStats(operation string, latency time.Duration) {
	mm.statsMutex.Lock()
	defer mm.statsMutex.Unlock()
	
	latencyMs := latency.Milliseconds()
	
	var stats *LatencyStats
	switch operation {
	case "get":
		stats = mm.stats.Performance.GetLatency
	case "set":
		stats = mm.stats.Performance.SetLatency
	case "search":
		stats = mm.stats.Performance.SearchLatency
	case "index":
		stats = mm.stats.Performance.IndexLatency
	default:
		return
	}
	
	if stats.Samples == 0 {
		stats.Min = latencyMs
		stats.Max = latencyMs
		stats.Mean = float64(latencyMs)
	} else {
		if latencyMs < stats.Min {
			stats.Min = latencyMs
		}
		if latencyMs > stats.Max {
			stats.Max = latencyMs
		}
		stats.Mean = (stats.Mean*float64(stats.Samples) + float64(latencyMs)) / float64(stats.Samples+1)
	}
	
	stats.Samples++
}

func (mm *MetadataManager) incrementOperationCount(operation string) {
	mm.statsMutex.Lock()
	defer mm.statsMutex.Unlock()
	
	mm.stats.OperationCounts[operation]++
}

func (mm *MetadataManager) incrementCacheHits() {
	mm.statsMutex.Lock()
	defer mm.statsMutex.Unlock()
	
	mm.stats.CacheHits++
}

func (mm *MetadataManager) incrementCacheMisses() {
	mm.statsMutex.Lock()
	defer mm.statsMutex.Unlock()
	
	mm.stats.CacheMisses++
}

// Background routines

func (mm *MetadataManager) syncRoutine() {
	if mm.config.SyncInterval <= 0 {
		return
	}
	
	ticker := time.NewTicker(mm.config.SyncInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-mm.ctx.Done():
			return
		case <-ticker.C:
			mm.performSync()
		}
	}
}

func (mm *MetadataManager) compactionRoutine() {
	if mm.config.CompactInterval <= 0 {
		return
	}
	
	ticker := time.NewTicker(mm.config.CompactInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-mm.ctx.Done():
			return
		case <-ticker.C:
			mm.performCompaction()
		}
	}
}

func (mm *MetadataManager) cacheMaintenanceRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-mm.ctx.Done():
			return
		case <-ticker.C:
			mm.performCacheMaintenance()
		}
	}
}

func (mm *MetadataManager) statsCollectionRoutine() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-mm.ctx.Done():
			return
		case <-ticker.C:
			mm.collectStats()
		}
	}
}

func (mm *MetadataManager) performSync() {
	if mm.levelDB != nil {
		// Force sync to disk
		// LevelDB doesn't have explicit sync, but compaction helps
	}
	
	mm.statsMutex.Lock()
	mm.stats.LastSync = time.Now()
	mm.statsMutex.Unlock()
}

func (mm *MetadataManager) performCompaction() {
	if mm.levelDB != nil {
		// Compact database
		mm.levelDB.CompactRange(util.Range{})
	}
	
	mm.statsMutex.Lock()
	mm.stats.LastCompaction = time.Now()
	mm.statsMutex.Unlock()
}

func (mm *MetadataManager) performCacheMaintenance() {
	mm.cacheMutex.Lock()
	defer mm.cacheMutex.Unlock()
	
	// Remove stale cache entries
	cutoff := time.Now().Add(-1 * time.Hour)
	for key, cached := range mm.cache {
		if cached.LastAccess.Before(cutoff) && cached.AccessCount < 5 {
			delete(mm.cache, key)
			mm.cacheSize--
		}
	}
}

func (mm *MetadataManager) collectStats() {
	// Update cache statistics
	mm.cacheMutex.RLock()
	cacheSize := len(mm.cache)
	mm.cacheMutex.RUnlock()
	
	mm.logger.Debug("metadata stats", "cache_size", cacheSize, "total_objects", mm.stats.TotalObjects)
}

// FileSystemMetadata implementation

func (fsm *FileSystemMetadata) store(key string, data []byte) error {
	safePath := strings.ReplaceAll(key, "/", string(filepath.Separator))
	metaPath := filepath.Join(fsm.basePath, safePath+".meta")
	
	if err := os.MkdirAll(filepath.Dir(metaPath), 0755); err != nil {
		return err
	}
	
	return os.WriteFile(metaPath, data, 0644)
}

func (fsm *FileSystemMetadata) load(key string) ([]byte, error) {
	safePath := strings.ReplaceAll(key, "/", string(filepath.Separator))
	metaPath := filepath.Join(fsm.basePath, safePath+".meta")
	
	data, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &StorageError{
				Code:    ErrCodeNotFound,
				Message: "metadata not found",
			}
		}
		return nil, err
	}
	
	return data, nil
}

func (fsm *FileSystemMetadata) delete(key string) error {
	safePath := strings.ReplaceAll(key, "/", string(filepath.Separator))
	metaPath := filepath.Join(fsm.basePath, safePath+".meta")
	
	err := os.Remove(metaPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	
	return nil
}

func (fsm *FileSystemMetadata) list(prefix string, options *ListOptions) ([]*ObjectMetadata, error) {
	var results []*ObjectMetadata
	
	err := filepath.Walk(fsm.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() || !strings.HasSuffix(path, ".meta") {
			return nil
		}
		
		// Extract key
		relPath, err := filepath.Rel(fsm.basePath, path)
		if err != nil {
			return err
		}
		
		key := strings.TrimSuffix(relPath, ".meta")
		key = strings.ReplaceAll(key, string(filepath.Separator), "/")
		
		if prefix != "" && !strings.HasPrefix(key, prefix) {
			return nil
		}
		
		// Load metadata
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		
		var metadata ObjectMetadata
		if err := json.Unmarshal(data, &metadata); err != nil {
			return err
		}
		
		results = append(results, &metadata)
		
		if options != nil && options.Limit > 0 && len(results) >= options.Limit {
			return filepath.SkipDir
		}
		
		return nil
	})
	
	return results, err
}
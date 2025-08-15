package storage

import (
	"sync"
	"time"
)

// MetadataConfig contains configuration for metadata management
type MetadataConfig struct {
	Backend          string        `json:"backend"` // leveldb, filesystem, memory
	DataDir          string        `json:"data_dir"`
	IndexingMode     string        `json:"indexing_mode"` // eager, lazy, disabled
	CacheSize        int           `json:"cache_size"`
	SyncInterval     time.Duration `json:"sync_interval"`
	CompactInterval  time.Duration `json:"compact_interval"`
	EnableSearch     bool          `json:"enable_search"`
	EnableVersioning bool          `json:"enable_versioning"`
}

// FileSystemMetadata implements filesystem-based metadata storage
type FileSystemMetadata struct {
	basePath string
	logger   interface{} // Using interface{} to avoid import cycle
}

// CachedMetadata represents cached metadata with additional information
type CachedMetadata struct {
	Metadata    *ObjectMetadata `json:"metadata"`
	CachedAt    time.Time       `json:"cached_at"`
	AccessCount int             `json:"access_count"`
	LastAccess  time.Time       `json:"last_access"`
}

// MetadataIndex represents an index for fast metadata queries
type MetadataIndex struct {
	Name      string              `json:"name"`
	Type      string              `json:"type"` // btree, hash, text
	Fields    []string            `json:"fields"`
	Values    map[string][]string `json:"values"` // value -> keys
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
	Stats     *IndexStats         `json:"stats"`
	mutex     *sync.RWMutex       `json:"-"` // Thread safety for index operations
}

// IndexStats contains statistics about an index
type IndexStats struct {
	TotalEntries int64     `json:"total_entries"`
	UniqueValues int64     `json:"unique_values"`
	LastUpdated  time.Time `json:"last_updated"`
	UpdateCount  int64     `json:"update_count"`
	QueryCount   int64     `json:"query_count"`
	AverageDepth float64   `json:"average_depth"`
}

// MetadataStats contains statistics about metadata operations
type MetadataStats struct {
	TotalObjects    int64                `json:"total_objects"`
	TotalSize       int64                `json:"total_size"`
	CacheHitRate    float64              `json:"cache_hit_rate"`
	CacheHits       int64                `json:"cache_hits"`
	CacheMisses     int64                `json:"cache_misses"`
	OperationCounts map[string]int64     `json:"operation_counts"`
	Performance     *MetadataPerformance `json:"performance"`
	LastCompaction  time.Time            `json:"last_compaction"`
	LastSync        time.Time            `json:"last_sync"`
}

// MetadataPerformance contains performance metrics
type MetadataPerformance struct {
	GetLatency    *LatencyStats `json:"get_latency"`
	SetLatency    *LatencyStats `json:"set_latency"`
	SearchLatency *LatencyStats `json:"search_latency"`
	IndexLatency  *LatencyStats `json:"index_latency"`
}

// Note: LatencyStats is defined in interface.go

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
	Field     string      `json:"field"`
	Operator  string      `json:"operator"` // eq, ne, gt, lt, gte, lte, in, like, regex
	Value     interface{} `json:"value"`
	LogicalOp string      `json:"logical_op"` // and, or, not
}

// SortOptions represents sorting options
type SortOptions struct {
	Field string `json:"field"`
	Order string `json:"order"` // asc, desc
}

// MetadataQueryResult represents the result of a metadata query
type MetadataQueryResult struct {
	Objects     []*ObjectMetadata `json:"objects"`
	Total       int64             `json:"total"`
	QueryTime   time.Duration     `json:"query_time"`
	IndexUsed   string            `json:"index_used"`
	Explanation string            `json:"explanation"`
}

// Note: ListOptions is defined in interface.go

// Note: ObjectMetadata is defined in interface.go
// We don't redefine it here to avoid conflicts

// Note: StorageError and error codes are defined in interface.go
// We don't redefine them here to avoid conflicts

// Additional error codes for metadata operations
const (
	ErrCodeInvalidKey   = "INVALID_KEY"
	ErrCodeInvalidValue = "INVALID_VALUE"
	ErrCodeBackendError = "BACKEND_ERROR"
	ErrCodeIndexError   = "INDEX_ERROR"
	ErrCodeCacheError   = "CACHE_ERROR"
)

// NewStorageError creates a new storage error
func NewStorageError(code, message, key string) *StorageError {
	return &StorageError{
		Code:    code,
		Message: message,
		Key:     key,
	}
}

// IsNotFound checks if the error is a not found error
func IsNotFound(err error) bool {
	if storageErr, ok := err.(*StorageError); ok {
		return storageErr.Code == ErrCodeNotFound
	}
	return false
}

// IsAlreadyExists checks if the error is an already exists error
func IsAlreadyExists(err error) bool {
	if storageErr, ok := err.(*StorageError); ok {
		return storageErr.Code == ErrCodeAlreadyExists
	}
	return false
}

// DefaultMetadataConfig returns a default metadata configuration
func DefaultMetadataConfig() *MetadataConfig {
	return &MetadataConfig{
		Backend:          "leveldb",
		DataDir:          "./data",
		IndexingMode:     "eager",
		CacheSize:        1000,
		SyncInterval:     30 * time.Second,
		CompactInterval:  5 * time.Minute,
		EnableSearch:     true,
		EnableVersioning: true,
	}
}

// NewLatencyStats creates a new latency stats instance
func NewLatencyStats() *LatencyStats {
	return &LatencyStats{
		Min:     0,
		Max:     0,
		Mean:    0.0,
		Median:  0,
		P95:     0,
		P99:     0,
		Samples: 0,
	}
}

// NewMetadataPerformance creates a new metadata performance instance
func NewMetadataPerformance() *MetadataPerformance {
	return &MetadataPerformance{
		GetLatency:    NewLatencyStats(),
		SetLatency:    NewLatencyStats(),
		SearchLatency: NewLatencyStats(),
		IndexLatency:  NewLatencyStats(),
	}
}

// NewMetadataStats creates a new metadata stats instance
func NewMetadataStats() *MetadataStats {
	return &MetadataStats{
		TotalObjects:    0,
		TotalSize:       0,
		CacheHitRate:    0.0,
		CacheHits:       0,
		CacheMisses:     0,
		OperationCounts: make(map[string]int64),
		Performance:     NewMetadataPerformance(),
		LastCompaction:  time.Time{},
		LastSync:        time.Time{},
	}
}

// NewIndexStats creates a new index stats instance
func NewIndexStats() *IndexStats {
	return &IndexStats{
		TotalEntries: 0,
		UniqueValues: 0,
		LastUpdated:  time.Now(),
		UpdateCount:  0,
		QueryCount:   0,
		AverageDepth: 0.0,
	}
}

package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/syndtr/goleveldb/leveldb/util"
)

// Cache management methods

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

// Statistics methods

func (mm *MetadataManager) updateLatencyStats(operation string, latency time.Duration) {
	mm.statsMutex.Lock()
	defer mm.statsMutex.Unlock()

	latencyMs := latency.Milliseconds()

	var stats *LatencyStats
	switch operation {
	case "get":
		if mm.stats.Performance.GetLatency == nil {
			mm.stats.Performance.GetLatency = NewLatencyStats()
		}
		stats = mm.stats.Performance.GetLatency
	case "set":
		if mm.stats.Performance.SetLatency == nil {
			mm.stats.Performance.SetLatency = NewLatencyStats()
		}
		stats = mm.stats.Performance.SetLatency
	case "search":
		if mm.stats.Performance.SearchLatency == nil {
			mm.stats.Performance.SearchLatency = NewLatencyStats()
		}
		stats = mm.stats.Performance.SearchLatency
	case "index":
		if mm.stats.Performance.IndexLatency == nil {
			mm.stats.Performance.IndexLatency = NewLatencyStats()
		}
		stats = mm.stats.Performance.IndexLatency
	default:
		return
	}

	// Update latency statistics
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
		if cached.LastAccess.Before(cutoff) {
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
			return nil, NewStorageError(ErrCodeNotFound, "metadata file not found", key)
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

		if !strings.HasSuffix(path, ".meta") {
			return nil
		}

		// Extract key from path
		relPath, err := filepath.Rel(fsm.basePath, path)
		if err != nil {
			return err
		}

		key := strings.TrimSuffix(relPath, ".meta")
		key = strings.ReplaceAll(key, string(filepath.Separator), "/")

		if !strings.HasPrefix(key, prefix) {
			return nil
		}

		// Read metadata
		data, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip files that can't be read
		}

		var metadata ObjectMetadata
		if err := json.Unmarshal(data, &metadata); err != nil {
			return nil // Skip invalid metadata files
		}

		results = append(results, &metadata)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Apply pagination if specified
	if options != nil && options.Limit > 0 {
		if options.Limit < len(results) {
			results = results[:options.Limit]
		}
	}

	return results, nil
}

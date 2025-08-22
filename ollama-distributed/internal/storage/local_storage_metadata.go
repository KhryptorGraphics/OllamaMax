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
	"time"
)

// storeMetadata stores metadata for an object
func (ls *LocalStorage) storeMetadata(key string, metadata *ObjectMetadata) error {
	metadataPath := ls.getMetadataPath(key)

	// Create metadata directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(metadataPath), 0755); err != nil {
		return &StorageError{
			Code:    ErrCodePermissionDenied,
			Message: fmt.Sprintf("failed to create metadata directory: %v", err),
		}
	}

	// Marshal metadata to JSON
	data, err := json.Marshal(metadata)
	if err != nil {
		return &StorageError{
			Code:    ErrCodeInternalError,
			Message: fmt.Sprintf("failed to marshal metadata: %v", err),
		}
	}

	// Write to temporary file first
	tempPath := metadataPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return &StorageError{
			Code:    ErrCodeIOError,
			Message: fmt.Sprintf("failed to write metadata: %v", err),
		}
	}

	// Atomic rename
	if err := os.Rename(tempPath, metadataPath); err != nil {
		os.Remove(tempPath)
		return &StorageError{
			Code:    ErrCodeIOError,
			Message: fmt.Sprintf("failed to rename metadata file: %v", err),
		}
	}

	return nil
}

// getMetadata retrieves metadata for an object
func (ls *LocalStorage) getMetadata(key string) (*ObjectMetadata, error) {
	// Check cache first
	if metadata := ls.getFromCache(key); metadata != nil {
		return metadata, nil
	}

	metadataPath := ls.getMetadataPath(key)

	// Check if metadata file exists
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		// Try to generate metadata from object file
		return ls.generateMetadata(key)
	}

	// Read metadata file
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, &StorageError{
			Code:    ErrCodeIOError,
			Message: fmt.Sprintf("failed to read metadata: %v", err),
		}
	}

	// Unmarshal metadata
	var metadata ObjectMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, &StorageError{
			Code:    ErrCodeInternalError,
			Message: fmt.Sprintf("failed to unmarshal metadata: %v", err),
		}
	}

	// Update cache
	ls.updateCache(key, &metadata)

	return &metadata, nil
}

// generateMetadata generates metadata from an object file
func (ls *LocalStorage) generateMetadata(key string) (*ObjectMetadata, error) {
	objectPath := ls.getObjectPath(key)

	// Get file info
	fileInfo, err := os.Stat(objectPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &StorageError{
				Code:    ErrCodeNotFound,
				Message: fmt.Sprintf("object not found: %s", key),
			}
		}
		return nil, &StorageError{
			Code:    ErrCodeIOError,
			Message: fmt.Sprintf("failed to stat object file: %v", err),
		}
	}

	// Create basic metadata
	metadata := &ObjectMetadata{
		Key:       key,
		Size:      fileInfo.Size(),
		CreatedAt: fileInfo.ModTime(),
		UpdatedAt: fileInfo.ModTime(),
	}

	// Calculate hash if needed
	if metadata.Hash == "" {
		hash, err := ls.calculateFileHash(objectPath)
		if err != nil {
			ls.logger.Warn("failed to calculate file hash", "key", key, "error", err)
		} else {
			metadata.Hash = hash
		}
	}

	// Store generated metadata
	if err := ls.storeMetadata(key, metadata); err != nil {
		ls.logger.Warn("failed to store generated metadata", "key", key, "error", err)
	}

	return metadata, nil
}

// calculateFileHash calculates the SHA256 hash of a file
func (ls *LocalStorage) calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// updateCache updates the metadata cache
func (ls *LocalStorage) updateCache(key string, metadata *ObjectMetadata) {
	ls.cacheMutex.Lock()
	defer ls.cacheMutex.Unlock()

	// Check cache size limit
	if ls.cacheSize >= ls.maxCacheSize && ls.maxCacheSize > 0 {
		ls.evictOldestCacheEntry()
	}

	// Add to cache
	if _, exists := ls.metadataCache[key]; !exists {
		ls.cacheSize++
	}
	ls.metadataCache[key] = metadata
}

// getFromCache retrieves metadata from cache
func (ls *LocalStorage) getFromCache(key string) *ObjectMetadata {
	ls.cacheMutex.RLock()
	defer ls.cacheMutex.RUnlock()

	return ls.metadataCache[key]
}

// removeFromCache removes metadata from cache
func (ls *LocalStorage) removeFromCache(key string) {
	ls.cacheMutex.Lock()
	defer ls.cacheMutex.Unlock()

	if _, exists := ls.metadataCache[key]; exists {
		delete(ls.metadataCache, key)
		ls.cacheSize--
	}
}

// evictOldestCacheEntry evicts the oldest cache entry
func (ls *LocalStorage) evictOldestCacheEntry() {
	if len(ls.metadataCache) == 0 {
		return
	}

	var oldestKey string
	var oldestTime time.Time

	for key, metadata := range ls.metadataCache {
		if oldestKey == "" || metadata.AccessedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = metadata.AccessedAt
		}
	}

	if oldestKey != "" {
		delete(ls.metadataCache, oldestKey)
		ls.cacheSize--
	}
}

// clearCache clears the metadata cache
func (ls *LocalStorage) clearCache() {
	ls.cacheMutex.Lock()
	defer ls.cacheMutex.Unlock()

	ls.metadataCache = make(map[string]*ObjectMetadata)
	ls.cacheSize = 0
}

// List lists objects with optional prefix filtering
func (ls *LocalStorage) List(ctx context.Context, prefix string, limit int, marker string) ([]string, string, error) {
	var keys []string
	var nextMarker string

	// Walk through the base directory
	err := filepath.Walk(ls.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and metadata directory
		if info.IsDir() || strings.Contains(path, ".metadata") {
			return nil
		}

		// Get relative path as key
		relPath, err := filepath.Rel(ls.basePath, path)
		if err != nil {
			return err
		}

		// Normalize path separators
		key := filepath.ToSlash(relPath)

		// Apply prefix filter
		if prefix != "" && !strings.HasPrefix(key, prefix) {
			return nil
		}

		// Apply marker filter (for pagination)
		if marker != "" && key <= marker {
			return nil
		}

		keys = append(keys, key)
		return nil
	})

	if err != nil {
		return nil, "", &StorageError{
			Code:    ErrCodeIOError,
			Message: fmt.Sprintf("failed to list objects: %v", err),
		}
	}

	// Sort keys
	sort.Strings(keys)

	// Apply limit
	if limit > 0 && len(keys) > limit {
		nextMarker = keys[limit-1]
		keys = keys[:limit]
	}

	return keys, nextMarker, nil
}

// GetMetadata returns metadata for an object
func (ls *LocalStorage) GetMetadata(ctx context.Context, key string) (*ObjectMetadata, error) {
	if err := ls.validateKey(key); err != nil {
		return nil, err
	}

	return ls.getMetadata(key)
}

// SetMetadata sets metadata for an object
func (ls *LocalStorage) SetMetadata(ctx context.Context, key string, metadata *ObjectMetadata) error {
	if err := ls.validateKey(key); err != nil {
		return err
	}

	// Check if object exists
	exists, err := ls.Exists(ctx, key)
	if err != nil {
		return err
	}
	if !exists {
		return &StorageError{
			Code:    ErrCodeNotFound,
			Message: fmt.Sprintf("object not found: %s", key),
		}
	}

	// Update metadata
	metadata.Key = key
	metadata.UpdatedAt = time.Now()

	// Store metadata
	if err := ls.storeMetadata(key, metadata); err != nil {
		return err
	}

	// Update cache
	ls.updateCache(key, metadata)

	return nil
}

// ListMetadata lists metadata for objects with optional prefix filtering
func (ls *LocalStorage) ListMetadata(ctx context.Context, prefix string, limit int, marker string) ([]*ObjectMetadata, string, error) {
	keys, nextMarker, err := ls.List(ctx, prefix, limit, marker)
	if err != nil {
		return nil, "", err
	}

	var metadataList []*ObjectMetadata
	for _, key := range keys {
		metadata, err := ls.getMetadata(key)
		if err != nil {
			ls.logger.Warn("failed to get metadata for key", "key", key, "error", err)
			continue
		}
		metadataList = append(metadataList, metadata)
	}

	return metadataList, nextMarker, nil
}

// GetCacheStats returns cache statistics
func (ls *LocalStorage) GetCacheStats() map[string]interface{} {
	ls.cacheMutex.RLock()
	defer ls.cacheMutex.RUnlock()

	return map[string]interface{}{
		"size":     ls.cacheSize,
		"max_size": ls.maxCacheSize,
		"entries":  len(ls.metadataCache),
	}
}

// RefreshCache refreshes the metadata cache
func (ls *LocalStorage) RefreshCache(ctx context.Context) error {
	ls.clearCache()

	// Reload frequently accessed metadata
	keys, _, err := ls.List(ctx, "", 100, "")
	if err != nil {
		return err
	}

	for _, key := range keys {
		if _, err := ls.getMetadata(key); err != nil {
			ls.logger.Warn("failed to refresh metadata for key", "key", key, "error", err)
		}
	}

	return nil
}

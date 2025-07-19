package models

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ContentAddressedStore implements a content-addressed storage system for models
type ContentAddressedStore struct {
	storeDir    string
	logger      *slog.Logger
	
	// Object tracking
	objects     map[string]*StoredObject
	objectsMutex sync.RWMutex
	
	// Reference counting
	refCounts   map[string]int
	refMutex    sync.RWMutex
	
	// Storage configuration
	maxObjects  int
	maxSize     int64
	
	// Cleanup settings
	cleanupInterval time.Duration
	maxAge         time.Duration
	
	// Statistics
	stats       *StoreStats
	statsMutex  sync.RWMutex
}

// StoredObject represents an object in the content-addressed store
type StoredObject struct {
	Hash      string            `json:"hash"`
	Size      int64             `json:"size"`
	Path      string            `json:"path"`
	RefCount  int               `json:"ref_count"`
	CreatedAt time.Time         `json:"created_at"`
	AccessedAt time.Time        `json:"accessed_at"`
	Metadata  map[string]string `json:"metadata"`
}

// StoreStats contains statistics about the store
type StoreStats struct {
	TotalObjects    int64 `json:"total_objects"`
	TotalSize       int64 `json:"total_size"`
	TotalReferences int64 `json:"total_references"`
	HitCount        int64 `json:"hit_count"`
	MissCount       int64 `json:"miss_count"`
	LastCleanup     time.Time `json:"last_cleanup"`
}

// NewContentAddressedStore creates a new content-addressed store
func NewContentAddressedStore(storeDir string, logger *slog.Logger) (*ContentAddressedStore, error) {
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create store directory: %w", err)
	}
	
	cas := &ContentAddressedStore{
		storeDir:        storeDir,
		logger:          logger,
		objects:         make(map[string]*StoredObject),
		refCounts:       make(map[string]int),
		maxObjects:      10000,
		maxSize:         100 * 1024 * 1024 * 1024, // 100GB
		cleanupInterval: time.Hour,
		maxAge:          24 * time.Hour * 7, // 1 week
		stats:           &StoreStats{},
	}
	
	// Load existing objects
	if err := cas.loadObjects(); err != nil {
		return nil, fmt.Errorf("failed to load objects: %w", err)
	}
	
	// Start cleanup routine
	go cas.cleanupRoutine()
	
	return cas, nil
}

// Store stores data in the content-addressed store
func (cas *ContentAddressedStore) Store(hash string, sourcePath string) error {
	cas.logger.Info("storing object", "hash", hash, "source", sourcePath)
	
	// Check if object already exists
	cas.objectsMutex.RLock()
	if obj, exists := cas.objects[hash]; exists {
		cas.objectsMutex.RUnlock()
		
		// Update access time and increment reference count
		obj.AccessedAt = time.Now()
		cas.incrementReference(hash)
		
		cas.statsMutex.Lock()
		cas.stats.HitCount++
		cas.statsMutex.Unlock()
		
		cas.logger.Debug("object already exists", "hash", hash, "ref_count", obj.RefCount)
		return nil
	}
	cas.objectsMutex.RUnlock()
	
	// Calculate actual hash to verify
	actualHash, err := cas.calculateHash(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to calculate hash: %w", err)
	}
	
	if actualHash != hash {
		return fmt.Errorf("hash mismatch: expected %s, got %s", hash, actualHash)
	}
	
	// Get file info
	info, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}
	
	// Create storage path
	storagePath := cas.getStoragePath(hash)
	if err := os.MkdirAll(filepath.Dir(storagePath), 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}
	
	// Copy file to storage
	if err := cas.copyFile(sourcePath, storagePath); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}
	
	// Create object entry
	obj := &StoredObject{
		Hash:       hash,
		Size:       info.Size(),
		Path:       storagePath,
		RefCount:   1,
		CreatedAt:  time.Now(),
		AccessedAt: time.Now(),
		Metadata:   make(map[string]string),
	}
	
	// Store object
	cas.objectsMutex.Lock()
	cas.objects[hash] = obj
	cas.objectsMutex.Unlock()
	
	cas.refMutex.Lock()
	cas.refCounts[hash] = 1
	cas.refMutex.Unlock()
	
	// Update statistics
	cas.statsMutex.Lock()
	cas.stats.TotalObjects++
	cas.stats.TotalSize += info.Size()
	cas.stats.TotalReferences++
	cas.stats.MissCount++
	cas.statsMutex.Unlock()
	
	cas.logger.Info("object stored", "hash", hash, "size", info.Size())
	return nil
}

// Get retrieves an object from the store
func (cas *ContentAddressedStore) Get(hash string) (*StoredObject, error) {
	cas.objectsMutex.RLock()
	obj, exists := cas.objects[hash]
	cas.objectsMutex.RUnlock()
	
	if !exists {
		cas.statsMutex.Lock()
		cas.stats.MissCount++
		cas.statsMutex.Unlock()
		
		return nil, fmt.Errorf("object not found: %s", hash)
	}
	
	// Update access time
	obj.AccessedAt = time.Now()
	
	cas.statsMutex.Lock()
	cas.stats.HitCount++
	cas.statsMutex.Unlock()
	
	return obj, nil
}

// GetReader returns a reader for an object
func (cas *ContentAddressedStore) GetReader(hash string) (io.ReadCloser, error) {
	obj, err := cas.Get(hash)
	if err != nil {
		return nil, err
	}
	
	return os.Open(obj.Path)
}

// Exists checks if an object exists in the store
func (cas *ContentAddressedStore) Exists(hash string) bool {
	cas.objectsMutex.RLock()
	defer cas.objectsMutex.RUnlock()
	
	_, exists := cas.objects[hash]
	return exists
}

// IncrementReference increments the reference count for an object
func (cas *ContentAddressedStore) IncrementReference(hash string) error {
	cas.objectsMutex.RLock()
	obj, exists := cas.objects[hash]
	cas.objectsMutex.RUnlock()
	
	if !exists {
		return fmt.Errorf("object not found: %s", hash)
	}
	
	cas.incrementReference(hash)
	
	cas.logger.Debug("incremented reference", "hash", hash, "ref_count", obj.RefCount)
	return nil
}

// DecrementReference decrements the reference count for an object
func (cas *ContentAddressedStore) DecrementReference(hash string) error {
	cas.objectsMutex.RLock()
	obj, exists := cas.objects[hash]
	cas.objectsMutex.RUnlock()
	
	if !exists {
		return fmt.Errorf("object not found: %s", hash)
	}
	
	cas.decrementReference(hash)
	
	cas.logger.Debug("decremented reference", "hash", hash, "ref_count", obj.RefCount)
	return nil
}

// incrementReference increments the reference count
func (cas *ContentAddressedStore) incrementReference(hash string) {
	cas.refMutex.Lock()
	cas.refCounts[hash]++
	cas.refMutex.Unlock()
	
	cas.objectsMutex.Lock()
	if obj, exists := cas.objects[hash]; exists {
		obj.RefCount++
	}
	cas.objectsMutex.Unlock()
	
	cas.statsMutex.Lock()
	cas.stats.TotalReferences++
	cas.statsMutex.Unlock()
}

// decrementReference decrements the reference count
func (cas *ContentAddressedStore) decrementReference(hash string) {
	cas.refMutex.Lock()
	if cas.refCounts[hash] > 0 {
		cas.refCounts[hash]--
	}
	refCount := cas.refCounts[hash]
	cas.refMutex.Unlock()
	
	cas.objectsMutex.Lock()
	if obj, exists := cas.objects[hash]; exists {
		obj.RefCount = refCount
	}
	cas.objectsMutex.Unlock()
	
	cas.statsMutex.Lock()
	if cas.stats.TotalReferences > 0 {
		cas.stats.TotalReferences--
	}
	cas.statsMutex.Unlock()
}

// ListObjects returns all objects in the store
func (cas *ContentAddressedStore) ListObjects() []*StoredObject {
	cas.objectsMutex.RLock()
	defer cas.objectsMutex.RUnlock()
	
	objects := make([]*StoredObject, 0, len(cas.objects))
	for _, obj := range cas.objects {
		objects = append(objects, obj)
	}
	
	return objects
}

// GetStats returns store statistics
func (cas *ContentAddressedStore) GetStats() *StoreStats {
	cas.statsMutex.RLock()
	defer cas.statsMutex.RUnlock()
	
	stats := *cas.stats
	return &stats
}

// calculateHash calculates the SHA256 hash of a file
func (cas *ContentAddressedStore) calculateHash(filePath string) (string, error) {
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

// getStoragePath returns the storage path for a hash
func (cas *ContentAddressedStore) getStoragePath(hash string) string {
	// Create a directory structure like: store/ab/cd/abcd...
	dir1 := hash[:2]
	dir2 := hash[2:4]
	return filepath.Join(cas.storeDir, dir1, dir2, hash)
}

// copyFile copies a file from source to destination
func (cas *ContentAddressedStore) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	
	return dstFile.Sync()
}

// loadObjects loads existing objects from the store
func (cas *ContentAddressedStore) loadObjects() error {
	return filepath.Walk(cas.storeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			return nil
		}
		
		// Extract hash from filename
		hash := filepath.Base(path)
		if len(hash) != 64 { // SHA256 hash length
			return nil
		}
		
		// Verify hash
		actualHash, err := cas.calculateHash(path)
		if err != nil {
			cas.logger.Error("failed to verify object hash", "path", path, "error", err)
			return nil
		}
		
		if actualHash != hash {
			cas.logger.Error("hash mismatch for stored object", "path", path, "expected", hash, "actual", actualHash)
			return nil
		}
		
		// Create object entry
		obj := &StoredObject{
			Hash:       hash,
			Size:       info.Size(),
			Path:       path,
			RefCount:   1, // Default reference count
			CreatedAt:  info.ModTime(),
			AccessedAt: info.ModTime(),
			Metadata:   make(map[string]string),
		}
		
		cas.objectsMutex.Lock()
		cas.objects[hash] = obj
		cas.objectsMutex.Unlock()
		
		cas.refMutex.Lock()
		cas.refCounts[hash] = 1
		cas.refMutex.Unlock()
		
		// Update statistics
		cas.statsMutex.Lock()
		cas.stats.TotalObjects++
		cas.stats.TotalSize += info.Size()
		cas.stats.TotalReferences++
		cas.statsMutex.Unlock()
		
		return nil
	})
}

// cleanupRoutine runs periodic cleanup
func (cas *ContentAddressedStore) cleanupRoutine() {
	ticker := time.NewTicker(cas.cleanupInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		if err := cas.cleanup(); err != nil {
			cas.logger.Error("cleanup failed", "error", err)
		}
	}
}

// cleanup removes unreferenced and old objects
func (cas *ContentAddressedStore) cleanup() error {
	cas.logger.Info("starting cleanup")
	
	var removed int
	var freedSize int64
	
	cas.objectsMutex.Lock()
	defer cas.objectsMutex.Unlock()
	
	cas.refMutex.Lock()
	defer cas.refMutex.Unlock()
	
	cutoff := time.Now().Add(-cas.maxAge)
	
	for hash, obj := range cas.objects {
		shouldRemove := false
		
		// Remove if no references
		if obj.RefCount == 0 {
			shouldRemove = true
		}
		
		// Remove if too old and not recently accessed
		if obj.AccessedAt.Before(cutoff) && obj.RefCount <= 1 {
			shouldRemove = true
		}
		
		if shouldRemove {
			// Remove file
			if err := os.Remove(obj.Path); err != nil {
				cas.logger.Error("failed to remove object file", "hash", hash, "path", obj.Path, "error", err)
				continue
			}
			
			// Remove from maps
			delete(cas.objects, hash)
			delete(cas.refCounts, hash)
			
			removed++
			freedSize += obj.Size
			
			cas.logger.Debug("removed object", "hash", hash, "size", obj.Size)
		}
	}
	
	// Update statistics
	cas.statsMutex.Lock()
	cas.stats.TotalObjects -= int64(removed)
	cas.stats.TotalSize -= freedSize
	cas.stats.LastCleanup = time.Now()
	cas.statsMutex.Unlock()
	
	cas.logger.Info("cleanup completed", "removed", removed, "freed_size", freedSize)
	
	return nil
}

// Verify verifies the integrity of all objects in the store
func (cas *ContentAddressedStore) Verify() error {
	cas.logger.Info("starting verification")
	
	var errors []string
	
	cas.objectsMutex.RLock()
	objects := make([]*StoredObject, 0, len(cas.objects))
	for _, obj := range cas.objects {
		objects = append(objects, obj)
	}
	cas.objectsMutex.RUnlock()
	
	for _, obj := range objects {
		// Check if file exists
		if _, err := os.Stat(obj.Path); err != nil {
			errors = append(errors, fmt.Sprintf("object file missing: %s (%s)", obj.Hash, obj.Path))
			continue
		}
		
		// Verify hash
		actualHash, err := cas.calculateHash(obj.Path)
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to calculate hash for %s: %v", obj.Hash, err))
			continue
		}
		
		if actualHash != obj.Hash {
			errors = append(errors, fmt.Sprintf("hash mismatch for %s: expected %s, got %s", obj.Path, obj.Hash, actualHash))
		}
	}
	
	if len(errors) > 0 {
		cas.logger.Error("verification failed", "errors", len(errors))
		for _, err := range errors {
			cas.logger.Error("verification error", "error", err)
		}
		return fmt.Errorf("verification failed with %d errors", len(errors))
	}
	
	cas.logger.Info("verification completed successfully")
	return nil
}

// Close closes the content-addressed store
func (cas *ContentAddressedStore) Close() error {
	cas.logger.Info("closing content-addressed store")
	return nil
}
package routing

import (
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// ContentStore manages local and remote content references
type ContentStore struct {
	// Local content
	localContent map[string]*ContentMetadata
	localMux     sync.RWMutex

	// Remote references
	remoteContent map[string]*RemoteContent
	remoteMux     sync.RWMutex

	// Caching
	cache *ContentCache

	// Storage backend
	storage Storage

	// Indexing
	index *ContentIndex
}

// ContentMetadata represents content metadata
type ContentMetadata struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Size     int64  `json:"size"`
	Checksum string `json:"checksum"`
	Version  string `json:"version"`

	// Content location
	Path string `json:"path,omitempty"`
	URL  string `json:"url,omitempty"`

	// Timestamps
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	AccessedAt time.Time `json:"accessed_at"`

	// Content properties
	MimeType    string `json:"mime_type"`
	Encoding    string `json:"encoding"`
	Compression string `json:"compression"`

	// Metadata
	Description string            `json:"description"`
	Author      string            `json:"author"`
	Tags        map[string]string `json:"tags"`
	Labels      []string          `json:"labels"`
}

// RemoteContent represents remote content reference
type RemoteContent struct {
	Metadata      *ContentMetadata
	Providers     []peer.ID
	LastUpdated   time.Time
	RetrievalCost int64
	Availability  float64
}

// ContentCache manages cached content
type ContentCache struct {
	cache   map[string]*CacheEntry
	maxSize int
	ttl     time.Duration
	mu      sync.RWMutex
}

// CacheEntry represents a cache entry
type CacheEntry struct {
	Content   *RemoteContent
	ExpiresAt time.Time
}

// Storage defines the storage interface
type Storage interface {
	Get(key string) ([]byte, error)
	Put(key string, value []byte) error
	Delete(key string) error
	Has(key string) bool
}

// MemoryStorage implements in-memory storage
type MemoryStorage struct {
	data map[string][]byte
	mu   sync.RWMutex
}

// NewContentStore creates a new content store
func NewContentStore(config *ContentRouterConfig) (*ContentStore, error) {
	cache := NewContentCache(config.CacheSize, config.CacheTTL)
	storage := NewMemoryStorage()
	index := NewContentIndex()

	return &ContentStore{
		localContent:  make(map[string]*ContentMetadata),
		remoteContent: make(map[string]*RemoteContent),
		cache:         cache,
		storage:       storage,
		index:         index,
	}, nil
}

// NewContentCache creates a new content cache
func NewContentCache(maxSize int, ttl time.Duration) *ContentCache {
	return &ContentCache{
		cache:   make(map[string]*CacheEntry),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

// NewMemoryStorage creates a new memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string][]byte),
	}
}

// ContentStore methods

// StoreLocal stores content locally
func (cs *ContentStore) StoreLocal(content *ContentMetadata) {
	cs.localMux.Lock()
	defer cs.localMux.Unlock()

	cs.localContent[content.ID] = content

	// Index the content
	if cs.index != nil {
		cs.index.AddContent(content)
	}
}

// GetLocal retrieves local content
func (cs *ContentStore) GetLocal(contentID string) (*ContentMetadata, bool) {
	cs.localMux.RLock()
	defer cs.localMux.RUnlock()

	content, exists := cs.localContent[contentID]
	return content, exists
}

// CacheRemote caches remote content
func (cs *ContentStore) CacheRemote(contentID string, metadata *ContentMetadata, providers []peer.ID) {
	cs.remoteMux.Lock()
	defer cs.remoteMux.Unlock()

	remote := &RemoteContent{
		Metadata:      metadata,
		Providers:     providers,
		LastUpdated:   time.Now(),
		RetrievalCost: 0,
		Availability:  1.0,
	}

	cs.remoteContent[contentID] = remote

	// Also cache in the cache layer
	if cs.cache != nil {
		cs.cache.Put(contentID, remote)
	}
}

// GetCached retrieves cached content
func (cs *ContentStore) GetCached(contentID string) (*RemoteContent, bool) {
	if cs.cache != nil {
		return cs.cache.Get(contentID)
	}

	cs.remoteMux.RLock()
	defer cs.remoteMux.RUnlock()

	content, exists := cs.remoteContent[contentID]
	return content, exists
}

// CleanupCache cleans up the cache
func (cs *ContentStore) CleanupCache() {
	if cs.cache != nil {
		cs.cache.Cleanup()
	}
}

// GetAllLocal returns all local content
func (cs *ContentStore) GetAllLocal() map[string]*ContentMetadata {
	cs.localMux.RLock()
	defer cs.localMux.RUnlock()

	result := make(map[string]*ContentMetadata)
	for k, v := range cs.localContent {
		result[k] = v
	}
	return result
}

// RemoveLocal removes local content
func (cs *ContentStore) RemoveLocal(contentID string) {
	cs.localMux.Lock()
	defer cs.localMux.Unlock()

	delete(cs.localContent, contentID)

	// Remove from index
	if cs.index != nil {
		cs.index.RemoveContent(contentID)
	}
}

// ContentCache methods

// Get retrieves content from cache
func (cc *ContentCache) Get(contentID string) (*RemoteContent, bool) {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	entry, exists := cc.cache[contentID]
	if !exists || time.Now().After(entry.ExpiresAt) {
		return nil, false
	}
	return entry.Content, true
}

// Put stores content in cache
func (cc *ContentCache) Put(contentID string, content *RemoteContent) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.cache[contentID] = &CacheEntry{
		Content:   content,
		ExpiresAt: time.Now().Add(cc.ttl),
	}

	// Enforce cache size limit
	if len(cc.cache) > cc.maxSize {
		cc.evictOldest()
	}
}

// Delete removes content from cache
func (cc *ContentCache) Delete(contentID string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	delete(cc.cache, contentID)
}

// Cleanup removes expired entries
func (cc *ContentCache) Cleanup() {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	now := time.Now()
	for contentID, entry := range cc.cache {
		if now.After(entry.ExpiresAt) {
			delete(cc.cache, contentID)
		}
	}
}

// evictOldest evicts the oldest cache entry
func (cc *ContentCache) evictOldest() {
	var oldestID string
	var oldestTime time.Time

	for contentID, entry := range cc.cache {
		if oldestID == "" || entry.ExpiresAt.Before(oldestTime) {
			oldestID = contentID
			oldestTime = entry.ExpiresAt
		}
	}

	if oldestID != "" {
		delete(cc.cache, oldestID)
	}
}

// GetSize returns the current cache size
func (cc *ContentCache) GetSize() int {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return len(cc.cache)
}

// GetMaxSize returns the maximum cache size
func (cc *ContentCache) GetMaxSize() int {
	return cc.maxSize
}

// SetMaxSize sets the maximum cache size
func (cc *ContentCache) SetMaxSize(maxSize int) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.maxSize = maxSize

	// Evict entries if necessary
	for len(cc.cache) > cc.maxSize {
		cc.evictOldest()
	}
}

// MemoryStorage methods

// Get retrieves data from storage
func (ms *MemoryStorage) Get(key string) ([]byte, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	data, exists := ms.data[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}
	return data, nil
}

// Put stores data in storage
func (ms *MemoryStorage) Put(key string, value []byte) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.data[key] = value
	return nil
}

// Delete removes data from storage
func (ms *MemoryStorage) Delete(key string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	delete(ms.data, key)
	return nil
}

// Has checks if key exists in storage
func (ms *MemoryStorage) Has(key string) bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	_, exists := ms.data[key]
	return exists
}

// GetSize returns the number of items in storage
func (ms *MemoryStorage) GetSize() int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return len(ms.data)
}

// Clear removes all data from storage
func (ms *MemoryStorage) Clear() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.data = make(map[string][]byte)
}

// GetKeys returns all keys in storage
func (ms *MemoryStorage) GetKeys() []string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	keys := make([]string, 0, len(ms.data))
	for key := range ms.data {
		keys = append(keys, key)
	}
	return keys
}

// Background task for ContentRouter

// cacheCleanupTask cleans up expired cache entries
func (cr *ContentRouter) cacheCleanupTask() {
	defer cr.wg.Done()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-cr.ctx.Done():
			return
		case <-ticker.C:
			if cr.contentStore != nil {
				cr.contentStore.CleanupCache()
			}
		}
	}
}

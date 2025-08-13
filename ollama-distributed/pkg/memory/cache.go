package memory

import (
	"container/list"
	"sync"
	"time"
)

// Cache interface defines cache operations
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	Delete(key string)
	Clear()
	Cleanup()
	Stats() CacheStats
}

// CacheStats holds cache statistics
type CacheStats struct {
	Size     int           `json:"size"`
	Capacity int           `json:"capacity"`
	Hits     int64         `json:"hits"`
	Misses   int64         `json:"misses"`
	HitRatio float64       `json:"hit_ratio"`
	TTL      time.Duration `json:"ttl"`
}

// LRUCache implements a thread-safe LRU cache with TTL
type LRUCache struct {
	capacity int
	ttl      time.Duration

	// Cache storage
	items map[string]*list.Element
	order *list.List

	// Statistics
	hits   int64
	misses int64

	// Thread safety
	mu sync.RWMutex
}

// cacheItem represents an item in the cache
type cacheItem struct {
	key       string
	value     interface{}
	timestamp time.Time
}

// NewLRUCache creates a new LRU cache
func NewLRUCache(capacity int, ttl time.Duration) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		ttl:      ttl,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

// Get retrieves a value from the cache
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	element, exists := c.items[key]
	if !exists {
		c.misses++
		return nil, false
	}

	item := element.Value.(*cacheItem)

	// Check TTL
	if c.ttl > 0 && time.Since(item.timestamp) > c.ttl {
		c.removeElement(element)
		c.misses++
		return nil, false
	}

	// Move to front (most recently used)
	c.order.MoveToFront(element)
	c.hits++

	return item.value, true
}

// Set adds or updates a value in the cache
func (c *LRUCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if key already exists
	if element, exists := c.items[key]; exists {
		// Update existing item
		item := element.Value.(*cacheItem)
		item.value = value
		item.timestamp = time.Now()
		c.order.MoveToFront(element)
		return
	}

	// Add new item
	item := &cacheItem{
		key:       key,
		value:     value,
		timestamp: time.Now(),
	}

	element := c.order.PushFront(item)
	c.items[key] = element

	// Check capacity and evict if necessary
	if c.order.Len() > c.capacity {
		c.evictOldest()
	}
}

// Delete removes a key from the cache
func (c *LRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, exists := c.items[key]; exists {
		c.removeElement(element)
	}
}

// Clear removes all items from the cache
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.order.Init()
}

// Cleanup removes expired items
func (c *LRUCache) Cleanup() {
	if c.ttl <= 0 {
		return // No TTL configured
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	// Iterate from back (oldest) to front
	for element := c.order.Back(); element != nil; {
		item := element.Value.(*cacheItem)

		if now.Sub(item.timestamp) > c.ttl {
			prev := element.Prev()
			c.removeElement(element)
			element = prev
		} else {
			// Items are ordered by access time, so we can stop here
			break
		}
	}
}

// Stats returns cache statistics
func (c *LRUCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	hitRatio := 0.0
	if total > 0 {
		hitRatio = float64(c.hits) / float64(total)
	}

	return CacheStats{
		Size:     len(c.items),
		Capacity: c.capacity,
		Hits:     c.hits,
		Misses:   c.misses,
		HitRatio: hitRatio,
		TTL:      c.ttl,
	}
}

// removeElement removes an element from the cache
func (c *LRUCache) removeElement(element *list.Element) {
	item := element.Value.(*cacheItem)
	delete(c.items, item.key)
	c.order.Remove(element)
}

// evictOldest removes the oldest item from the cache
func (c *LRUCache) evictOldest() {
	element := c.order.Back()
	if element != nil {
		c.removeElement(element)
	}
}

// TTLCache implements a simple TTL-based cache
type TTLCache struct {
	items map[string]*ttlItem
	ttl   time.Duration
	mu    sync.RWMutex
}

// ttlItem represents an item with TTL
type ttlItem struct {
	value     interface{}
	timestamp time.Time
}

// NewTTLCache creates a new TTL cache
func NewTTLCache(ttl time.Duration) *TTLCache {
	return &TTLCache{
		items: make(map[string]*ttlItem),
		ttl:   ttl,
	}
}

// Get retrieves a value from the TTL cache
func (c *TTLCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Check TTL
	if time.Since(item.timestamp) > c.ttl {
		// Item expired, remove it
		c.mu.RUnlock()
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		c.mu.RLock()
		return nil, false
	}

	return item.value, true
}

// Set adds a value to the TTL cache
func (c *TTLCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &ttlItem{
		value:     value,
		timestamp: time.Now(),
	}
}

// Delete removes a key from the TTL cache
func (c *TTLCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Clear removes all items from the TTL cache
func (c *TTLCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*ttlItem)
}

// Cleanup removes expired items from the TTL cache
func (c *TTLCache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if now.Sub(item.timestamp) > c.ttl {
			delete(c.items, key)
		}
	}
}

// Stats returns TTL cache statistics
func (c *TTLCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return CacheStats{
		Size:     len(c.items),
		Capacity: -1, // No capacity limit
		Hits:     0,  // Not tracked in TTL cache
		Misses:   0,  // Not tracked in TTL cache
		HitRatio: 0,  // Not tracked in TTL cache
		TTL:      c.ttl,
	}
}

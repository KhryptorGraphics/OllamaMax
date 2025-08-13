package resources

import (
	"sync"
	"time"
)

// ResourceQuery represents a query for discovering resources
type ResourceQuery struct {
	ID             string            `json:"id"`
	ModelTypes     []string          `json:"model_types,omitempty"`
	MinCPU         int               `json:"min_cpu,omitempty"`
	MinMemory      int64             `json:"min_memory,omitempty"`
	RequiredGPU    bool              `json:"required_gpu,omitempty"`
	MaxLatency     time.Duration     `json:"max_latency,omitempty"`
	MaxPrice       float64           `json:"max_price,omitempty"`
	PreferredZones []string          `json:"preferred_zones,omitempty"`
	Tags           map[string]string `json:"tags,omitempty"`
	Timestamp      time.Time         `json:"timestamp"`
}

// DiscoveryResponse represents a response to a resource discovery query
type DiscoveryResponse struct {
	QueryID   string           `json:"query_id"`
	Results   []*Advertisement `json:"results"`
	Timestamp time.Time        `json:"timestamp"`
	Source    string           `json:"source"`
}

// DiscoveryCache manages a cache of discovered resources
type DiscoveryCache struct {
	cache     map[string]*Advertisement
	mutex     sync.RWMutex
	maxSize   int
	ttl       time.Duration
	lastClean time.Time
}

// NewDiscoveryCache creates a new discovery cache
func NewDiscoveryCache(maxSize int, ttl time.Duration) *DiscoveryCache {
	return &DiscoveryCache{
		cache:     make(map[string]*Advertisement),
		maxSize:   maxSize,
		ttl:       ttl,
		lastClean: time.Now(),
	}
}

// Store stores an advertisement in the cache
func (dc *DiscoveryCache) Store(ad *Advertisement) {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()

	// Check if cache is full
	if len(dc.cache) >= dc.maxSize {
		// Remove oldest entries
		dc.evictOldest()
	}

	dc.cache[ad.ID] = ad
}

// Find searches for advertisements matching a query
func (dc *DiscoveryCache) Find(query *ResourceQuery) []*Advertisement {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()

	var results []*Advertisement
	now := time.Now()

	for _, ad := range dc.cache {
		// Check if advertisement is still valid
		if now.Sub(ad.Timestamp) > dc.ttl {
			continue
		}

		// Check if advertisement matches query
		if dc.matchesQuery(ad, query) {
			results = append(results, ad)
		}
	}

	return results
}

// matchesQuery checks if an advertisement matches a query (internal helper)
func (dc *DiscoveryCache) matchesQuery(ad *Advertisement, query *ResourceQuery) bool {
	// Check model types
	if len(query.ModelTypes) > 0 {
		hasModel := false
		for _, modelType := range query.ModelTypes {
			for _, supportedModel := range ad.Capabilities.SupportedModels {
				if supportedModel == modelType {
					hasModel = true
					break
				}
			}
			if hasModel {
				break
			}
		}
		if !hasModel {
			return false
		}
	}

	// Check CPU requirements
	if query.MinCPU > 0 && ad.Capabilities.CPUCores < query.MinCPU {
		return false
	}

	// Check memory requirements
	if query.MinMemory > 0 && ad.Capabilities.Memory < query.MinMemory {
		return false
	}

	// Check GPU requirements
	if query.RequiredGPU && len(ad.Capabilities.GPUs) == 0 {
		return false
	}

	// Check latency requirements
	if query.MaxLatency > 0 && ad.Capabilities.Latency > query.MaxLatency {
		return false
	}

	// Check price requirements
	if query.MaxPrice > 0 && ad.Capabilities.PricePerToken > query.MaxPrice {
		return false
	}

	return true
}

// Cleanup removes expired entries from the cache
func (dc *DiscoveryCache) Cleanup() {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()

	now := time.Now()

	for id, ad := range dc.cache {
		if now.Sub(ad.Timestamp) > dc.ttl {
			delete(dc.cache, id)
		}
	}

	dc.lastClean = now
}

// evictOldest removes the oldest entries when cache is full
func (dc *DiscoveryCache) evictOldest() {
	if len(dc.cache) == 0 {
		return
	}

	// Find oldest entry
	var oldestID string
	var oldestTime time.Time

	for id, ad := range dc.cache {
		if oldestID == "" || ad.Timestamp.Before(oldestTime) {
			oldestID = id
			oldestTime = ad.Timestamp
		}
	}

	// Remove oldest entry
	if oldestID != "" {
		delete(dc.cache, oldestID)
	}
}

// GetSize returns the current cache size
func (dc *DiscoveryCache) GetSize() int {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()
	return len(dc.cache)
}

// Clear empties the cache
func (dc *DiscoveryCache) Clear() {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	dc.cache = make(map[string]*Advertisement)
}

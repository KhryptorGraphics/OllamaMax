package models

import (
	"sync"
	"time"
)

// ObjectPool provides efficient memory management for replication objects
type ObjectPool struct {
	replicaInfoPool     sync.Pool
	replicationTaskPool sync.Pool
	policyPool          sync.Pool
}

// NewObjectPool creates a new object pool for efficient memory management
func NewObjectPool() *ObjectPool {
	return &ObjectPool{
		replicaInfoPool: sync.Pool{
			New: func() interface{} {
				return &ReplicaInfo{
					Metadata: make(map[string]string),
				}
			},
		},
		replicationTaskPool: sync.Pool{
			New: func() interface{} {
				return &ReplicationTask{
					ResponseChan: make(chan error, 1),
				}
			},
		},
		policyPool: sync.Pool{
			New: func() interface{} {
				return &ReplicationPolicy{
					Constraints: make(map[string]string),
				}
			},
		},
	}
}

// GetReplicaInfo gets a ReplicaInfo from the pool
func (op *ObjectPool) GetReplicaInfo() *ReplicaInfo {
	replica := op.replicaInfoPool.Get().(*ReplicaInfo)
	
	// Reset the replica info
	replica.ModelName = ""
	replica.PeerID = ""
	replica.Status = ""
	replica.LastSync = time.Time{}
	replica.SyncAttempts = 0
	replica.Health = ""
	replica.CreatedAt = time.Time{}
	replica.UpdatedAt = time.Time{}
	
	// Clear the metadata map
	for k := range replica.Metadata {
		delete(replica.Metadata, k)
	}
	
	return replica
}

// PutReplicaInfo returns a ReplicaInfo to the pool
func (op *ObjectPool) PutReplicaInfo(replica *ReplicaInfo) {
	if replica != nil {
		op.replicaInfoPool.Put(replica)
	}
}

// GetReplicationTask gets a ReplicationTask from the pool
func (op *ObjectPool) GetReplicationTask() *ReplicationTask {
	task := op.replicationTaskPool.Get().(*ReplicationTask)
	
	// Reset the task
	task.ID = ""
	task.Type = ""
	task.ModelName = ""
	task.SourcePeer = ""
	task.TargetPeer = ""
	task.Status = "pending"
	task.Progress = 0.0
	task.Error = ""
	task.CreatedAt = time.Time{}
	task.UpdatedAt = time.Time{}
	task.CompletedAt = nil
	
	// Clear the metadata map
	for k := range task.Metadata {
		delete(task.Metadata, k)
	}
	
	// Ensure channel is ready
	select {
	case <-task.ResponseChan:
		// Clear any existing value
	default:
		// Channel is already empty
	}
	
	return task
}

// PutReplicationTask returns a ReplicationTask to the pool
func (op *ObjectPool) PutReplicationTask(task *ReplicationTask) {
	if task != nil {
		// Close response channel if needed
		if task.ResponseChan != nil {
			close(task.ResponseChan)
		}
		op.replicationTaskPool.Put(task)
	}
}

// GetReplicationPolicy gets a ReplicationPolicy from the pool
func (op *ObjectPool) GetReplicationPolicy() *ReplicationPolicy {
	policy := op.policyPool.Get().(*ReplicationPolicy)
	
	// Reset the policy
	policy.ModelName = ""
	policy.MinReplicas = 0
	policy.MaxReplicas = 0
	policy.PreferredPeers = policy.PreferredPeers[:0] // Clear slice but keep capacity
	policy.ExcludedPeers = policy.ExcludedPeers[:0]   // Clear slice but keep capacity
	policy.ReplicationFactor = 0
	policy.SyncInterval = 0
	policy.Priority = 0
	policy.CreatedAt = time.Time{}
	policy.UpdatedAt = time.Time{}
	
	// Clear the constraints map
	for k := range policy.Constraints {
		delete(policy.Constraints, k)
	}
	
	return policy
}

// PutReplicationPolicy returns a ReplicationPolicy to the pool
func (op *ObjectPool) PutReplicationPolicy(policy *ReplicationPolicy) {
	if policy != nil {
		op.policyPool.Put(policy)
	}
}

// BufferPool provides reusable byte buffers for efficient memory usage
type BufferPool struct {
	pool sync.Pool
}

// NewBufferPool creates a new buffer pool
func NewBufferPool(initialSize int) *BufferPool {
	return &BufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, initialSize)
			},
		},
	}
}

// Get gets a buffer from the pool
func (bp *BufferPool) Get() []byte {
	return bp.pool.Get().([]byte)
}

// Put returns a buffer to the pool
func (bp *BufferPool) Put(buf []byte) {
	if buf != nil {
		// Reset length but keep capacity
		buf = buf[:0]
		bp.pool.Put(buf)
	}
}

// ConnectionCache provides connection caching for improved performance
type ConnectionCache struct {
	cache   map[string]*CachedConnection
	mutex   sync.RWMutex
	maxSize int
	ttl     time.Duration
}

// CachedConnection represents a cached connection
type CachedConnection struct {
	Connection interface{}
	CreatedAt  time.Time
	LastUsed   time.Time
	UseCount   int64
}

// NewConnectionCache creates a new connection cache
func NewConnectionCache(maxSize int, ttl time.Duration) *ConnectionCache {
	cache := &ConnectionCache{
		cache:   make(map[string]*CachedConnection),
		maxSize: maxSize,
		ttl:     ttl,
	}
	
	// Start cleanup routine
	go cache.cleanup()
	
	return cache
}

// Get retrieves a connection from the cache
func (cc *ConnectionCache) Get(key string) (interface{}, bool) {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()
	
	conn, exists := cc.cache[key]
	if !exists {
		return nil, false
	}
	
	// Check if connection is expired
	if time.Since(conn.CreatedAt) > cc.ttl {
		go cc.invalidate(key) // Async cleanup
		return nil, false
	}
	
	// Update last used time
	conn.LastUsed = time.Now()
	conn.UseCount++
	
	return conn.Connection, true
}

// Set stores a connection in the cache
func (cc *ConnectionCache) Set(key string, connection interface{}) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	
	// Check cache size limit
	if len(cc.cache) >= cc.maxSize {
		cc.evictLRU()
	}
	
	cc.cache[key] = &CachedConnection{
		Connection: connection,
		CreatedAt:  time.Now(),
		LastUsed:   time.Now(),
		UseCount:   1,
	}
}

// invalidate removes a connection from the cache
func (cc *ConnectionCache) invalidate(key string) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	delete(cc.cache, key)
}

// evictLRU removes the least recently used connection
func (cc *ConnectionCache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time
	
	first := true
	for key, conn := range cc.cache {
		if first || conn.LastUsed.Before(oldestTime) {
			oldestKey = key
			oldestTime = conn.LastUsed
			first = false
		}
	}
	
	if oldestKey != "" {
		delete(cc.cache, oldestKey)
	}
}

// cleanup periodically removes expired connections
func (cc *ConnectionCache) cleanup() {
	ticker := time.NewTicker(cc.ttl / 2)
	defer ticker.Stop()
	
	for range ticker.C {
		cc.mutex.Lock()
		now := time.Now()
		
		for key, conn := range cc.cache {
			if now.Sub(conn.CreatedAt) > cc.ttl {
				delete(cc.cache, key)
			}
		}
		
		cc.mutex.Unlock()
	}
}

// Size returns the current cache size
func (cc *ConnectionCache) Size() int {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()
	return len(cc.cache)
}
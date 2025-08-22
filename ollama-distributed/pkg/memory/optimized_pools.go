package memory

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// OptimizedMemoryManager provides high-performance memory management with bounded pools
type OptimizedMemoryManager struct {
	// Channel pools organized by capacity
	channelPools map[int]*ChannelPool
	channelMu    sync.RWMutex

	// Buffer pools with size classes (powers of 2)
	bufferPools map[int]*BufferPool
	bufferMu    sync.RWMutex

	// Object pools for frequent allocations
	requestPool  *ObjectPool[interface{}]
	responsePool *ObjectPool[interface{}]

	// Memory pressure monitoring
	memoryPressure int64 // atomic
	gcOptimizer    *GCOptimizer

	// Performance metrics
	metrics *PoolMetrics

	// Configuration
	config *OptimizedConfig
}

// OptimizedConfig holds configuration for optimized memory management
type OptimizedConfig struct {
	// Pool limits
	MaxChannelPools    int `yaml:"max_channel_pools" json:"max_channel_pools"`
	MaxBufferPools     int `yaml:"max_buffer_pools" json:"max_buffer_pools"`
	MaxObjectsPerPool  int `yaml:"max_objects_per_pool" json:"max_objects_per_pool"`

	// GC optimization
	EnableGCOptimization bool          `yaml:"enable_gc_optimization" json:"enable_gc_optimization"`
	GCTargetPercent      int           `yaml:"gc_target_percent" json:"gc_target_percent"`
	GCPressureThreshold  int64         `yaml:"gc_pressure_threshold" json:"gc_pressure_threshold"`

	// Monitoring
	EnableMetrics       bool          `yaml:"enable_metrics" json:"enable_metrics"`
	MetricsInterval     time.Duration `yaml:"metrics_interval" json:"metrics_interval"`
	CleanupInterval     time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"`
}

// ChannelPool manages a pool of channels with specific capacity
type ChannelPool struct {
	capacity int
	pool     sync.Pool
	created  int64 // atomic
	reused   int64 // atomic
}

// BufferPool manages a pool of byte slices with specific size
type BufferPool struct {
	size int
	pool sync.Pool
	hits int64 // atomic
	miss int64 // atomic
}

// ObjectPool manages a pool of typed objects
type ObjectPool[T any] struct {
	pool sync.Pool
	new  func() T
	hits int64 // atomic
	miss int64 // atomic
}

// PoolMetrics tracks memory pool performance
type PoolMetrics struct {
	// Channel pool metrics
	ChannelPoolHits  int64 `json:"channel_pool_hits"`
	ChannelPoolMiss  int64 `json:"channel_pool_miss"`
	ChannelReuseRate float64 `json:"channel_reuse_rate"`

	// Buffer pool metrics
	BufferPoolHits  int64 `json:"buffer_pool_hits"`
	BufferPoolMiss  int64 `json:"buffer_pool_miss"`
	BufferReuseRate float64 `json:"buffer_reuse_rate"`

	// Memory pressure
	MemoryPressure int64 `json:"memory_pressure"`
	GCCount        int64 `json:"gc_count"`

	// Timestamp
	LastUpdated time.Time `json:"last_updated"`
}

// DefaultOptimizedConfig returns optimized default configuration
func DefaultOptimizedConfig() *OptimizedConfig {
	return &OptimizedConfig{
		MaxChannelPools:      20,
		MaxBufferPools:       15,
		MaxObjectsPerPool:    1000,
		EnableGCOptimization: true,
		GCTargetPercent:      100,
		GCPressureThreshold:  80, // 80% memory usage
		EnableMetrics:        true,
		MetricsInterval:      30 * time.Second,
		CleanupInterval:      5 * time.Minute,
	}
}

// NewOptimizedMemoryManager creates a new optimized memory manager
func NewOptimizedMemoryManager(config *OptimizedConfig) *OptimizedMemoryManager {
	if config == nil {
		config = DefaultOptimizedConfig()
	}

	manager := &OptimizedMemoryManager{
		channelPools: make(map[int]*ChannelPool),
		bufferPools:  make(map[int]*BufferPool),
		config:       config,
		metrics:      &PoolMetrics{LastUpdated: time.Now()},
	}

	// Initialize object pools
	manager.requestPool = NewObjectPool(func() interface{} { return make(map[string]interface{}) })
	manager.responsePool = NewObjectPool(func() interface{} { return make(map[string]interface{}) })

	// Initialize GC optimizer if enabled
	if config.EnableGCOptimization {
		manager.gcOptimizer = NewGCOptimizer(config)
	}

	return manager
}

// GetChannel returns a channel from the pool or creates a new one
func (m *OptimizedMemoryManager) GetChannel(capacity int) chan interface{} {
	pool := m.getChannelPool(capacity)
	
	if ch := pool.pool.Get(); ch != nil {
		atomic.AddInt64(&pool.reused, 1)
		atomic.AddInt64(&m.metrics.ChannelPoolHits, 1)
		return ch.(chan interface{})
	}

	// Create new channel
	atomic.AddInt64(&pool.created, 1)
	atomic.AddInt64(&m.metrics.ChannelPoolMiss, 1)
	return make(chan interface{}, capacity)
}

// ReturnChannel returns a channel to the pool
func (m *OptimizedMemoryManager) ReturnChannel(ch chan interface{}, capacity int) {
	// Clear the channel before returning to pool
	for len(ch) > 0 {
		<-ch
	}

	pool := m.getChannelPool(capacity)
	pool.pool.Put(ch)
}

// GetBuffer returns a buffer from the pool or creates a new one
func (m *OptimizedMemoryManager) GetBuffer(size int) []byte {
	// Round up to next power of 2 for better pooling
	poolSize := nextPowerOf2(size)
	pool := m.getBufferPool(poolSize)

	if buf := pool.pool.Get(); buf != nil {
		atomic.AddInt64(&pool.hits, 1)
		atomic.AddInt64(&m.metrics.BufferPoolHits, 1)
		return buf.([]byte)[:size] // Return slice with requested size
	}

	// Create new buffer
	atomic.AddInt64(&pool.miss, 1)
	atomic.AddInt64(&m.metrics.BufferPoolMiss, 1)
	return make([]byte, size, poolSize)
}

// ReturnBuffer returns a buffer to the pool
func (m *OptimizedMemoryManager) ReturnBuffer(buf []byte) {
	if buf == nil || len(buf) == 0 {
		return
	}

	poolSize := cap(buf)
	pool := m.getBufferPool(poolSize)
	
	// Reset buffer content for security
	for i := range buf {
		buf[i] = 0
	}
	
	pool.pool.Put(buf[:poolSize]) // Return with full capacity
}

// GetRequest returns a request object from the pool
func (m *OptimizedMemoryManager) GetRequest() interface{} {
	return m.requestPool.Get()
}

// ReturnRequest returns a request object to the pool
func (m *OptimizedMemoryManager) ReturnRequest(req interface{}) {
	m.requestPool.Put(req)
}

// GetResponse returns a response object from the pool
func (m *OptimizedMemoryManager) GetResponse() interface{} {
	return m.responsePool.Get()
}

// ReturnResponse returns a response object to the pool
func (m *OptimizedMemoryManager) ReturnResponse(resp interface{}) {
	m.responsePool.Put(resp)
}

// getChannelPool gets or creates a channel pool for the given capacity
func (m *OptimizedMemoryManager) getChannelPool(capacity int) *ChannelPool {
	m.channelMu.RLock()
	if pool, exists := m.channelPools[capacity]; exists {
		m.channelMu.RUnlock()
		return pool
	}
	m.channelMu.RUnlock()

	// Create new pool
	m.channelMu.Lock()
	defer m.channelMu.Unlock()

	// Double-check after acquiring write lock
	if pool, exists := m.channelPools[capacity]; exists {
		return pool
	}

	pool := &ChannelPool{
		capacity: capacity,
		pool: sync.Pool{
			New: func() interface{} {
				return make(chan interface{}, capacity)
			},
		},
	}

	m.channelPools[capacity] = pool
	return pool
}

// getBufferPool gets or creates a buffer pool for the given size
func (m *OptimizedMemoryManager) getBufferPool(size int) *BufferPool {
	m.bufferMu.RLock()
	if pool, exists := m.bufferPools[size]; exists {
		m.bufferMu.RUnlock()
		return pool
	}
	m.bufferMu.RUnlock()

	// Create new pool
	m.bufferMu.Lock()
	defer m.bufferMu.Unlock()

	// Double-check after acquiring write lock
	if pool, exists := m.bufferPools[size]; exists {
		return pool
	}

	pool := &BufferPool{
		size: size,
		pool: sync.Pool{
			New: func() interface{} {
				return make([]byte, size)
			},
		},
	}

	m.bufferPools[size] = pool
	return pool
}

// NewObjectPool creates a new typed object pool
func NewObjectPool[T any](newFunc func() T) *ObjectPool[T] {
	return &ObjectPool[T]{
		new: newFunc,
		pool: sync.Pool{
			New: func() interface{} {
				return newFunc()
			},
		},
	}
}

// Get returns an object from the pool
func (p *ObjectPool[T]) Get() T {
	if obj := p.pool.Get(); obj != nil {
		atomic.AddInt64(&p.hits, 1)
		return obj.(T)
	}
	atomic.AddInt64(&p.miss, 1)
	return p.new()
}

// Put returns an object to the pool
func (p *ObjectPool[T]) Put(obj T) {
	p.pool.Put(obj)
}

// GetMetrics returns current pool metrics
func (m *OptimizedMemoryManager) GetMetrics() *PoolMetrics {
	// Update calculated metrics
	channelTotal := atomic.LoadInt64(&m.metrics.ChannelPoolHits) + atomic.LoadInt64(&m.metrics.ChannelPoolMiss)
	if channelTotal > 0 {
		m.metrics.ChannelReuseRate = float64(atomic.LoadInt64(&m.metrics.ChannelPoolHits)) / float64(channelTotal)
	}

	bufferTotal := atomic.LoadInt64(&m.metrics.BufferPoolHits) + atomic.LoadInt64(&m.metrics.BufferPoolMiss)
	if bufferTotal > 0 {
		m.metrics.BufferReuseRate = float64(atomic.LoadInt64(&m.metrics.BufferPoolHits)) / float64(bufferTotal)
	}

	m.metrics.MemoryPressure = atomic.LoadInt64(&m.memoryPressure)
	m.metrics.LastUpdated = time.Now()

	return m.metrics
}

// UpdateMemoryPressure updates the current memory pressure
func (m *OptimizedMemoryManager) UpdateMemoryPressure() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	// Calculate memory pressure as percentage of allocated vs system memory
	if memStats.Sys > 0 {
		pressure := int64((memStats.Alloc * 100) / memStats.Sys)
		atomic.StoreInt64(&m.memoryPressure, pressure)
		
		// Trigger GC optimization if pressure is high
		if m.gcOptimizer != nil && pressure > m.config.GCPressureThreshold {
			m.gcOptimizer.OptimizeGC()
		}
	}
}

// Cleanup performs periodic cleanup of pools
func (m *OptimizedMemoryManager) Cleanup() {
	// Force GC to clean up unused pool objects
	runtime.GC()
	
	// Update memory pressure
	m.UpdateMemoryPressure()
}

// nextPowerOf2 returns the next power of 2 greater than or equal to n
func nextPowerOf2(n int) int {
	if n <= 0 {
		return 1
	}
	
	// If n is already a power of 2, return it
	if (n & (n - 1)) == 0 {
		return n
	}
	
	// Find the next power of 2
	power := 1
	for power < n {
		power <<= 1
	}
	return power
}
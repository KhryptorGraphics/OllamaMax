package memory

import (
	"sync"
	"time"
)

// Pool manages a pool of reusable memory buffers
type Pool struct {
	itemSize int
	pool     sync.Pool
	stats    PoolStats
	mu       sync.RWMutex
}

// PoolStats holds memory pool statistics
type PoolStats struct {
	ItemSize    int     `json:"item_size"`
	Gets        int64   `json:"gets"`
	Puts        int64   `json:"puts"`
	Hits        int64   `json:"hits"`
	Misses      int64   `json:"misses"`
	HitRatio    float64 `json:"hit_ratio"`
	CurrentSize int     `json:"current_size"`
}

// MultiBufferPool manages multiple pools of different buffer sizes
type MultiBufferPool struct {
	pools map[int]*Pool
	mu    sync.RWMutex
}

// NewPool creates a new memory pool
func NewPool(itemSize int) *Pool {
	p := &Pool{
		itemSize: itemSize,
	}

	p.pool.New = func() interface{} {
		p.mu.Lock()
		p.stats.Misses++
		p.mu.Unlock()
		return make([]byte, itemSize)
	}

	return p
}

// Get retrieves a buffer from the pool
func (p *Pool) Get() []byte {
	p.mu.Lock()
	p.stats.Gets++
	p.mu.Unlock()

	buf := p.pool.Get().([]byte)

	// Check if we got a buffer from the pool (hit) or created new (miss)
	if len(buf) == p.itemSize {
		p.mu.Lock()
		p.stats.Hits++
		p.mu.Unlock()
	}

	// Reset buffer
	for i := range buf {
		buf[i] = 0
	}

	return buf
}

// Put returns a buffer to the pool
func (p *Pool) Put(buf []byte) {
	if len(buf) != p.itemSize {
		return // Wrong size, don't put back
	}

	p.mu.Lock()
	p.stats.Puts++
	p.mu.Unlock()

	p.pool.Put(buf)
}

// Stats returns pool statistics
func (p *Pool) Stats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := p.stats
	if stats.Gets > 0 {
		stats.HitRatio = float64(stats.Hits) / float64(stats.Gets)
	}

	return stats
}

// Clear clears the pool
func (p *Pool) Clear() {
	// Create a new pool to clear the old one
	p.pool = sync.Pool{
		New: func() interface{} {
			p.mu.Lock()
			p.stats.Misses++
			p.mu.Unlock()
			return make([]byte, p.itemSize)
		},
	}
}

// Cleanup performs pool maintenance (no-op for sync.Pool)
func (p *Pool) Cleanup() {
	// sync.Pool automatically manages cleanup
	// This method is here for interface compatibility
}

// BufferPool manages pools of different buffer sizes
type LegacyBufferPool struct {
	pools map[int]*Pool
	mu    sync.RWMutex
}

// NewLegacyBufferPool creates a new buffer pool manager
func NewLegacyBufferPool() *LegacyBufferPool {
	return &LegacyBufferPool{
		pools: make(map[int]*Pool),
	}
}

// GetBuffer gets a buffer of the specified size
func (bp *LegacyBufferPool) GetBuffer(size int) []byte {
	// Round up to nearest power of 2 for better pool utilization
	poolSize := legacyNextPowerOf2(size)

	bp.mu.RLock()
	pool, exists := bp.pools[poolSize]
	bp.mu.RUnlock()

	if !exists {
		bp.mu.Lock()
		// Double-check after acquiring write lock
		if pool, exists = bp.pools[poolSize]; !exists {
			pool = NewPool(poolSize)
			bp.pools[poolSize] = pool
		}
		bp.mu.Unlock()
	}

	buf := pool.Get()
	return buf[:size] // Return slice of requested size
}

// PutBuffer returns a buffer to the appropriate pool
func (bp *LegacyBufferPool) PutBuffer(buf []byte) {
	if cap(buf) == 0 {
		return
	}

	poolSize := cap(buf)

	bp.mu.RLock()
	pool, exists := bp.pools[poolSize]
	bp.mu.RUnlock()

	if exists {
		// Restore full capacity before putting back
		fullBuf := buf[:cap(buf)]
		pool.Put(fullBuf)
	}
}

// GetStats returns statistics for all pools
func (bp *MultiBufferPool) GetStats() map[int]PoolStats {
	bp.mu.RLock()
	defer bp.mu.RUnlock()

	stats := make(map[int]PoolStats)
	for size, pool := range bp.pools {
		stats[size] = pool.Stats()
	}

	return stats
}

// Clear clears all pools
func (bp *MultiBufferPool) Clear() {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	for _, pool := range bp.pools {
		pool.Clear()
	}
}

// legacyNextPowerOf2 returns the next power of 2 greater than or equal to n
func legacyNextPowerOf2(n int) int {
	if n <= 0 {
		return 1
	}

	// Handle powers of 2
	if n&(n-1) == 0 {
		return n
	}

	// Find the next power of 2
	power := 1
	for power < n {
		power <<= 1
	}

	return power
}

// SimpleObjectPool manages a pool of reusable objects
type SimpleObjectPool struct {
	pool      sync.Pool
	newFunc   func() interface{}
	resetFunc func(interface{})
	stats     ObjectPoolStats
	mu        sync.RWMutex
}

// ObjectPoolStats holds object pool statistics
type ObjectPoolStats struct {
	Gets     int64   `json:"gets"`
	Puts     int64   `json:"puts"`
	Hits     int64   `json:"hits"`
	Misses   int64   `json:"misses"`
	HitRatio float64 `json:"hit_ratio"`
}

// NewSimpleObjectPool creates a new object pool
func NewSimpleObjectPool(newFunc func() interface{}, resetFunc func(interface{})) *SimpleObjectPool {
	op := &SimpleObjectPool{
		newFunc:   newFunc,
		resetFunc: resetFunc,
	}

	op.pool.New = func() interface{} {
		op.mu.Lock()
		op.stats.Misses++
		op.mu.Unlock()
		return newFunc()
	}

	return op
}

// Get retrieves an object from the pool
func (op *SimpleObjectPool) Get() interface{} {
	op.mu.Lock()
	op.stats.Gets++
	op.mu.Unlock()

	obj := op.pool.Get()

	// This is a hit since we got an object from the pool
	op.mu.Lock()
	op.stats.Hits++
	op.mu.Unlock()

	return obj
}

// Put returns an object to the pool
func (op *SimpleObjectPool) Put(obj interface{}) {
	if op.resetFunc != nil {
		op.resetFunc(obj)
	}

	op.mu.Lock()
	op.stats.Puts++
	op.mu.Unlock()

	op.pool.Put(obj)
}

// Stats returns object pool statistics
func (op *SimpleObjectPool) Stats() ObjectPoolStats {
	op.mu.RLock()
	defer op.mu.RUnlock()

	stats := op.stats
	if stats.Gets > 0 {
		stats.HitRatio = float64(stats.Hits) / float64(stats.Gets)
	}

	return stats
}

// Clear clears the object pool
func (op *SimpleObjectPool) Clear() {
	op.pool = sync.Pool{
		New: func() interface{} {
			op.mu.Lock()
			op.stats.Misses++
			op.mu.Unlock()
			return op.newFunc()
		},
	}
}

// TimedPool manages objects with automatic cleanup
type TimedPool struct {
	pool   *SimpleObjectPool
	items  map[interface{}]time.Time
	maxAge time.Duration
	mu     sync.RWMutex
}

// NewTimedPool creates a new timed object pool
func NewTimedPool(newFunc func() interface{}, resetFunc func(interface{}), maxAge time.Duration) *TimedPool {
	return &TimedPool{
		pool:   NewSimpleObjectPool(newFunc, resetFunc),
		items:  make(map[interface{}]time.Time),
		maxAge: maxAge,
	}
}

// Get retrieves an object from the timed pool
func (tp *TimedPool) Get() interface{} {
	obj := tp.pool.Get()

	tp.mu.Lock()
	tp.items[obj] = time.Now()
	tp.mu.Unlock()

	return obj
}

// Put returns an object to the timed pool
func (tp *TimedPool) Put(obj interface{}) {
	tp.mu.Lock()
	delete(tp.items, obj)
	tp.mu.Unlock()

	tp.pool.Put(obj)
}

// Cleanup removes old objects from the pool
func (tp *TimedPool) Cleanup() {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	now := time.Now()
	for obj, timestamp := range tp.items {
		if now.Sub(timestamp) > tp.maxAge {
			delete(tp.items, obj)
			// Don't put back to pool, let it be garbage collected
		}
	}
}

// Stats returns timed pool statistics
func (tp *TimedPool) Stats() ObjectPoolStats {
	return tp.pool.Stats()
}

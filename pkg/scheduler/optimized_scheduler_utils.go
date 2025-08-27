package scheduler

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/khryptorgraphics/ollamamax/pkg/types"
)

// ParallelNodeFilter applies constraints to nodes in parallel for improved performance
type ParallelNodeFilter struct {
	workers int
	mu      sync.RWMutex
}

// NewParallelNodeFilter creates a new parallel node filter
func NewParallelNodeFilter() *ParallelNodeFilter {
	return &ParallelNodeFilter{
		workers: runtime.NumCPU() * 2,
	}
}

// ApplyConstraintsParallel filters nodes that match constraints using parallel processing
// This replaces the O(n) sequential processing with concurrent evaluation
func (pf *ParallelNodeFilter) ApplyConstraintsParallel(
	nodes []*IntelligentNode,
	constraints *TaskConstraints,
	matcher func(*IntelligentNode, *TaskConstraints) bool,
) []*IntelligentNode {
	if len(nodes) == 0 {
		return nodes
	}

	// For small node counts, sequential is faster
	if len(nodes) < 10 {
		return pf.applyConstraintsSequential(nodes, constraints, matcher)
	}

	// Channel for collecting matching nodes
	resultCh := make(chan *IntelligentNode, len(nodes))
	var wg sync.WaitGroup

	// Create worker pool
	workCh := make(chan *IntelligentNode, len(nodes))
	
	// Start workers
	for i := 0; i < pf.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for node := range workCh {
				if matcher(node, constraints) {
					resultCh <- node
				}
			}
		}()
	}

	// Send work to workers
	for _, node := range nodes {
		workCh <- node
	}
	close(workCh)

	// Wait for workers to complete
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results
	var candidateNodes []*IntelligentNode
	for node := range resultCh {
		candidateNodes = append(candidateNodes, node)
	}

	return candidateNodes
}

// applyConstraintsSequential is the fallback for small node counts
func (pf *ParallelNodeFilter) applyConstraintsSequential(
	nodes []*IntelligentNode,
	constraints *TaskConstraints,
	matcher func(*IntelligentNode, *TaskConstraints) bool,
) []*IntelligentNode {
	var candidateNodes []*IntelligentNode
	for _, node := range nodes {
		if matcher(node, constraints) {
			candidateNodes = append(candidateNodes, node)
		}
	}
	return candidateNodes
}

// CircularTaskHistory implements a ring buffer for task history
// This eliminates the O(n) slice operations and reduces GC pressure
type CircularTaskHistory struct {
	buffer []*TaskExecutionRecord
	size   int
	head   int
	tail   int
	cap    int
	mu     sync.RWMutex
}

// NewCircularTaskHistory creates a new circular task history buffer
func NewCircularTaskHistory(capacity int) *CircularTaskHistory {
	return &CircularTaskHistory{
		buffer: make([]*TaskExecutionRecord, capacity),
		cap:    capacity,
		size:   0,
		head:   0,
		tail:   0,
	}
}

// Add adds a new task execution record to the history
func (cth *CircularTaskHistory) Add(record *TaskExecutionRecord) {
	cth.mu.Lock()
	defer cth.mu.Unlock()

	cth.buffer[cth.tail] = record
	cth.tail = (cth.tail + 1) % cth.cap

	if cth.size < cth.cap {
		cth.size++
	} else {
		// Overwrite oldest entry
		cth.head = (cth.head + 1) % cth.cap
	}
}

// GetAll returns all records in chronological order
func (cth *CircularTaskHistory) GetAll() []*TaskExecutionRecord {
	cth.mu.RLock()
	defer cth.mu.RUnlock()

	if cth.size == 0 {
		return nil
	}

	result := make([]*TaskExecutionRecord, cth.size)
	for i := 0; i < cth.size; i++ {
		idx := (cth.head + i) % cth.cap
		result[i] = cth.buffer[idx]
	}
	return result
}

// GetRecent returns the most recent n records
func (cth *CircularTaskHistory) GetRecent(n int) []*TaskExecutionRecord {
	cth.mu.RLock()
	defer cth.mu.RUnlock()

	if n > cth.size {
		n = cth.size
	}
	if n == 0 {
		return nil
	}

	result := make([]*TaskExecutionRecord, n)
	start := (cth.tail - n + cth.cap) % cth.cap
	for i := 0; i < n; i++ {
		idx := (start + i) % cth.cap
		result[i] = cth.buffer[idx]
	}
	return result
}

// Size returns the current number of records
func (cth *CircularTaskHistory) Size() int {
	cth.mu.RLock()
	defer cth.mu.RUnlock()
	return cth.size
}

// TaskExecutionPool provides object pooling for TaskExecutionRecord
// This reduces GC pressure from frequent allocations
var taskExecutionPool = sync.Pool{
	New: func() interface{} {
		return &TaskExecutionRecord{
			Metadata: make(map[string]interface{}, 8),
		}
	},
}

// GetTaskExecutionRecord gets a record from the pool
func GetTaskExecutionRecord() *TaskExecutionRecord {
	return taskExecutionPool.Get().(*TaskExecutionRecord)
}

// PutTaskExecutionRecord returns a record to the pool
func PutTaskExecutionRecord(record *TaskExecutionRecord) {
	// Clear the record before returning to pool
	record.TaskID = ""
	record.NodeID = ""
	record.StartTime = time.Time{}
	record.EndTime = time.Time{}
	record.Status = ""
	record.Error = nil
	record.ResourceUsage = types.ResourceUsage{}
	
	// Clear metadata
	for k := range record.Metadata {
		delete(record.Metadata, k)
	}
	
	taskExecutionPool.Put(record)
}

// AsyncTaskProgressMonitor handles task progress updates asynchronously
// This eliminates the synchronous blocking in the original implementation
type AsyncTaskProgressMonitor struct {
	updateCh chan *progressUpdate
	stopCh   chan struct{}
	wg       sync.WaitGroup
	workers  int
}

type progressUpdate struct {
	task   *ScheduledTask
	status string
	progress float64
}

// NewAsyncTaskProgressMonitor creates a new async progress monitor
func NewAsyncTaskProgressMonitor(workers int) *AsyncTaskProgressMonitor {
	if workers <= 0 {
		workers = 4
	}
	
	monitor := &AsyncTaskProgressMonitor{
		updateCh: make(chan *progressUpdate, 1000),
		stopCh:   make(chan struct{}),
		workers:  workers,
	}
	
	monitor.start()
	return monitor
}

// start initializes the worker goroutines
func (atpm *AsyncTaskProgressMonitor) start() {
	for i := 0; i < atpm.workers; i++ {
		atpm.wg.Add(1)
		go atpm.worker()
	}
}

// worker processes progress updates
func (atpm *AsyncTaskProgressMonitor) worker() {
	defer atpm.wg.Done()
	
	for {
		select {
		case update := <-atpm.updateCh:
			if update != nil {
				// Process the update asynchronously
				atpm.processUpdate(update)
			}
		case <-atpm.stopCh:
			return
		}
	}
}

// UpdateProgress sends a progress update asynchronously
func (atpm *AsyncTaskProgressMonitor) UpdateProgress(task *ScheduledTask, status string, progress float64) {
	select {
	case atpm.updateCh <- &progressUpdate{
		task:     task,
		status:   status,
		progress: progress,
	}:
	default:
		// Channel full, drop update rather than block
		// This prevents blocking the main execution path
	}
}

// processUpdate handles the actual progress update
func (atpm *AsyncTaskProgressMonitor) processUpdate(update *progressUpdate) {
	// Update task progress without blocking
	// This would typically update metrics, logs, or state
	atomic.StoreUint64((*uint64)(&update.task.Progress), uint64(update.progress*100))
	
	// Log progress if needed
	if update.progress == 1.0 || update.status == "completed" {
		// Task completed
		update.task.Status = TaskStatusCompleted
	}
}

// Stop gracefully shuts down the monitor
func (atpm *AsyncTaskProgressMonitor) Stop() {
	close(atpm.stopCh)
	atpm.wg.Wait()
	close(atpm.updateCh)
}

// NodeSelectionCache provides caching for node selection decisions
// This reduces repeated expensive computations
type NodeSelectionCache struct {
	cache map[string]*cacheEntry
	mu    sync.RWMutex
	ttl   time.Duration
}

type cacheEntry struct {
	node      *IntelligentNode
	timestamp time.Time
	score     float64
}

// NewNodeSelectionCache creates a new node selection cache
func NewNodeSelectionCache(ttl time.Duration) *NodeSelectionCache {
	cache := &NodeSelectionCache{
		cache: make(map[string]*cacheEntry),
		ttl:   ttl,
	}
	
	// Start cleanup goroutine
	go cache.cleanup()
	
	return cache
}

// Get retrieves a cached node selection
func (nsc *NodeSelectionCache) Get(key string) (*IntelligentNode, float64, bool) {
	nsc.mu.RLock()
	defer nsc.mu.RUnlock()
	
	entry, exists := nsc.cache[key]
	if !exists {
		return nil, 0, false
	}
	
	// Check if entry is still valid
	if time.Since(entry.timestamp) > nsc.ttl {
		return nil, 0, false
	}
	
	return entry.node, entry.score, true
}

// Set stores a node selection in the cache
func (nsc *NodeSelectionCache) Set(key string, node *IntelligentNode, score float64) {
	nsc.mu.Lock()
	defer nsc.mu.Unlock()
	
	nsc.cache[key] = &cacheEntry{
		node:      node,
		timestamp: time.Now(),
		score:     score,
	}
}

// cleanup periodically removes expired entries
func (nsc *NodeSelectionCache) cleanup() {
	ticker := time.NewTicker(nsc.ttl)
	defer ticker.Stop()
	
	for range ticker.C {
		nsc.mu.Lock()
		now := time.Now()
		for key, entry := range nsc.cache {
			if now.Sub(entry.timestamp) > nsc.ttl {
				delete(nsc.cache, key)
			}
		}
		nsc.mu.Unlock()
	}
}
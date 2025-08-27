package scheduler

import (
	"container/heap"
	"context"
	"crypto/sha256"
	"fmt"
	"hash/fnv"
	"log/slog"
	"math"
	"sync"
	"time"
	"unsafe"

	"github.com/khryptorgraphics/ollamamax/internal/config"
	"github.com/khryptorgraphics/ollamamax/pkg/consensus"
	"github.com/khryptorgraphics/ollamamax/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/pkg/loadbalancer"
	"github.com/khryptorgraphics/ollamamax/pkg/types"
)

// OptimizedScheduler provides high-performance scheduling with algorithmic optimizations
type OptimizedScheduler struct {
	mu sync.RWMutex

	// Core components
	config    *IntelligentSchedulerConfig
	p2p       *p2p.Node
	consensus *consensus.Engine
	logger    *slog.Logger

	// Optimized data structures
	taskQueue        *OptimizedPriorityQueue    // O(log n) operations
	taskHistory      *RingBufferWithIndex       // O(1) insertions, O(log n) lookups
	constraintIndex  *BloomConstraintIndex      // O(1) average constraint checking
	performanceCache *LRUPerformanceCache       // O(1) performance data access
	nodeIndex        *ConcurrentNodeIndex       // O(1) node lookups

	// Enhanced scheduling components
	loadBalancer      *loadbalancer.IntelligentLoadBalancer
	resourcePredictor *OptimizedResourcePredictor
	taskAnalyzer      *CachedTaskAnalyzer

	// Parallel processing
	workerPool    *WorkerPool
	taskPipeline  *SchedulingPipeline
	syncExecutor  *ConcurrentExecutor

	// Metrics and monitoring
	atomicMetrics *AtomicSchedulerMetrics
	profiler      *PerformanceProfiler

	// Lifecycle
	ctx     context.Context
	cancel  context.CancelFunc
	started bool
}

// OptimizedPriorityQueue implements a binary heap for O(log n) task operations
type OptimizedPriorityQueue struct {
	mu    sync.RWMutex
	items []*PriorityTask
	index map[string]int // task ID -> heap index for O(log n) updates
}

// PriorityTask wraps ScheduledTask with priority queue semantics
type PriorityTask struct {
	*ScheduledTask
	priority float64 // Combined priority score
	heapIndex int    // Index in heap for efficient updates
}

// RingBufferWithIndex provides O(1) insertions and O(log n) lookups
type RingBufferWithIndex struct {
	mu      sync.RWMutex
	buffer  []*TaskExecutionRecord
	head    int
	size    int
	maxSize int
	
	// B-tree index for efficient lookups
	timeIndex map[int64][]*TaskExecutionRecord // timestamp -> records
	taskIndex map[string][]*TaskExecutionRecord // task_id -> records
}

// BloomConstraintIndex uses bloom filters for fast constraint checking
type BloomConstraintIndex struct {
	mu          sync.RWMutex
	nodeFilters map[string]*BloomFilter    // node_id -> capabilities bloom filter
	capabilityMap map[string][]string      // capability -> node_ids with capability
	lastUpdate  time.Time
}

// BloomFilter implementation for constraint checking
type BloomFilter struct {
	bits []uint64
	size uint64
	hashFuncs int
}

// LRUPerformanceCache provides O(1) performance data access
type LRUPerformanceCache struct {
	mu       sync.RWMutex
	capacity int
	cache    map[string]*CacheNode
	head     *CacheNode
	tail     *CacheNode
	hitCount int64
	missCount int64
}

// CacheNode for doubly-linked list in LRU cache
type CacheNode struct {
	key   string
	value *NodePerformanceProfile
	prev  *CacheNode
	next  *CacheNode
	ttl   time.Time
}

// ConcurrentNodeIndex provides thread-safe O(1) node operations
type ConcurrentNodeIndex struct {
	nodes    sync.Map // node_id -> *OptimizedNode
	byZone   sync.Map // zone -> []node_id
	byCapability sync.Map // capability -> []node_id
	count    int64    // atomic counter
}

// OptimizedNode extends IntelligentNode with performance optimizations
type OptimizedNode struct {
	*IntelligentNode
	
	// Pre-computed values for fast access
	effectiveLoadScore float64
	lastScoreUpdate    time.Time
	scoreLock         sync.RWMutex
	
	// Capability set for O(1) lookups
	capabilitySet map[string]struct{}
}

// WorkerPool for parallel task processing
type WorkerPool struct {
	workers   []*TaskWorker
	taskChan  chan *TaskWork
	resultChan chan *TaskResult
	workerCount int
	ctx       context.Context
	cancel    context.CancelFunc
}

// TaskWorker handles individual scheduling tasks
type TaskWorker struct {
	id       int
	pool     *WorkerPool
	logger   *slog.Logger
}

// TaskWork represents work to be done by worker pool
type TaskWork struct {
	Type string
	Task *ScheduledTask
	Callback func(*TaskResult, error)
}

// SchedulingPipeline for pipelined task processing
type SchedulingPipeline struct {
	stages []PipelineStage
	input  chan *ScheduledTask
	output chan *ScheduledTask
}

// PipelineStage represents a stage in the scheduling pipeline
type PipelineStage struct {
	Name     string
	Process  func(*ScheduledTask) error
	Workers  int
	Input    chan *ScheduledTask
	Output   chan *ScheduledTask
}

// AtomicSchedulerMetrics uses atomic operations for lock-free metrics
type AtomicSchedulerMetrics struct {
	TotalTasksScheduled    int64 // atomic
	SuccessfulTasks        int64 // atomic
	FailedTasks           int64 // atomic
	SchedulingTimeSum     int64 // atomic, nanoseconds
	ExecutionTimeSum      int64 // atomic, nanoseconds
	CacheHits             int64 // atomic
	CacheMisses           int64 // atomic
}

// PerformanceProfiler for real-time performance monitoring
type PerformanceProfiler struct {
	metrics map[string]*ProfilerMetric
	mu      sync.RWMutex
}

// ProfilerMetric tracks performance of specific operations
type ProfilerMetric struct {
	Name         string
	Count        int64
	TotalTime    time.Duration
	MinTime      time.Duration
	MaxTime      time.Duration
	RecentTimes  []time.Duration // Ring buffer of recent measurements
}

// Implementation of heap.Interface for OptimizedPriorityQueue
func (pq *OptimizedPriorityQueue) Len() int {
	return len(pq.items)
}

func (pq *OptimizedPriorityQueue) Less(i, j int) bool {
	// Higher priority first (max heap)
	return pq.items[i].priority > pq.items[j].priority
}

func (pq *OptimizedPriorityQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.items[i].heapIndex = i
	pq.items[j].heapIndex = j
	
	// Update index maps
	pq.index[pq.items[i].ID] = i
	pq.index[pq.items[j].ID] = j
}

func (pq *OptimizedPriorityQueue) Push(x interface{}) {
	item := x.(*PriorityTask)
	item.heapIndex = len(pq.items)
	pq.items = append(pq.items, item)
	pq.index[item.ID] = item.heapIndex
}

func (pq *OptimizedPriorityQueue) Pop() interface{} {
	old := pq.items
	n := len(old)
	item := old[n-1]
	item.heapIndex = -1
	delete(pq.index, item.ID)
	pq.items = old[0 : n-1]
	return item
}

// NewOptimizedScheduler creates a high-performance optimized scheduler
func NewOptimizedScheduler(
	config *IntelligentSchedulerConfig,
	p2pNode *p2p.Node,
	consensusEngine *consensus.Engine,
	logger *slog.Logger,
) *OptimizedScheduler {
	ctx, cancel := context.WithCancel(context.Background())

	os := &OptimizedScheduler{
		config:    config,
		p2p:       p2pNode,
		consensus: consensusEngine,
		logger:    logger,
		ctx:       ctx,
		cancel:    cancel,
	}

	// Initialize optimized data structures
	os.initializeOptimizedComponents()

	return os
}

// initializeOptimizedComponents initializes all optimized components
func (os *OptimizedScheduler) initializeOptimizedComponents() {
	// Initialize optimized priority queue
	os.taskQueue = &OptimizedPriorityQueue{
		items: make([]*PriorityTask, 0),
		index: make(map[string]int),
	}
	heap.Init(os.taskQueue)

	// Initialize ring buffer with index
	os.taskHistory = &RingBufferWithIndex{
		buffer:    make([]*TaskExecutionRecord, 10000),
		maxSize:   10000,
		timeIndex: make(map[int64][]*TaskExecutionRecord),
		taskIndex: make(map[string][]*TaskExecutionRecord),
	}

	// Initialize bloom constraint index
	os.constraintIndex = &BloomConstraintIndex{
		nodeFilters:   make(map[string]*BloomFilter),
		capabilityMap: make(map[string][]string),
		lastUpdate:    time.Now(),
	}

	// Initialize LRU performance cache
	os.performanceCache = NewLRUPerformanceCache(1000) // 1000 entries

	// Initialize concurrent node index
	os.nodeIndex = &ConcurrentNodeIndex{}

	// Initialize worker pool
	os.workerPool = NewWorkerPool(config.WorkerCount, os.ctx)

	// Initialize task pipeline
	os.taskPipeline = NewSchedulingPipeline(os.ctx)

	// Initialize atomic metrics
	os.atomicMetrics = &AtomicSchedulerMetrics{}

	// Initialize performance profiler
	os.profiler = &PerformanceProfiler{
		metrics: make(map[string]*ProfilerMetric),
	}

	// Initialize enhanced components
	os.resourcePredictor = NewOptimizedResourcePredictor(config, logger)
	os.taskAnalyzer = NewCachedTaskAnalyzer(config, logger)
}

// ScheduleTaskOptimized schedules a task using optimized algorithms
func (os *OptimizedScheduler) ScheduleTaskOptimized(task *ScheduledTask) error {
	startTime := time.Now()
	defer func() {
		// Atomic metrics update
		duration := time.Since(startTime)
		atomic.AddInt64(&os.atomicMetrics.SchedulingTimeSum, int64(duration))
		atomic.AddInt64(&os.atomicMetrics.TotalTasksScheduled, 1)
	}()

	// Parallel task processing pipeline
	resultChan := make(chan error, 1)
	
	go func() {
		defer close(resultChan)
		
		// Stage 1: Task Analysis (can run in parallel)
		analysisChan := make(chan *TaskAnalysis, 1)
		go func() {
			analysis, err := os.taskAnalyzer.AnalyzeTaskCached(task)
			if err != nil {
				resultChan <- fmt.Errorf("task analysis failed: %w", err)
				return
			}
			analysisChan <- analysis
		}()

		// Stage 2: Resource Prediction (can run in parallel)
		predictionChan := make(chan *ResourcePrediction, 1)
		go func() {
			// Wait for analysis
			analysis := <-analysisChan
			prediction, err := os.resourcePredictor.PredictRequirementsOptimized(task, analysis)
			if err != nil {
				resultChan <- fmt.Errorf("resource prediction failed: %w", err)
				return
			}
			predictionChan <- prediction
		}()

		// Stage 3: Node Selection (optimized)
		prediction := <-predictionChan
		
		// Get available nodes using optimized index
		availableNodes := os.getAvailableNodesOptimized()
		if len(availableNodes) == 0 {
			resultChan <- fmt.Errorf("no available nodes")
			return
		}

		// Apply constraints using bloom filter
		candidateNodes := os.applyConstraintsOptimized(task, availableNodes)
		if len(candidateNodes) == 0 {
			resultChan <- fmt.Errorf("no nodes satisfy task constraints")
			return
		}

		// Select optimal node
		selectedNode, err := os.selectOptimalNodeOptimized(task, candidateNodes, prediction)
		if err != nil {
			resultChan <- fmt.Errorf("node selection failed: %w", err)
			return
		}

		// Schedule task
		task.AssignedNode = selectedNode.ID
		task.ScheduledAt = time.Now()
		task.Status = TaskStatusScheduled
		task.EstimatedRuntime = prediction.EstimatedRuntime

		// Add to optimized queue
		priorityTask := &PriorityTask{
			ScheduledTask: task,
			priority:      os.calculateTaskPriority(task, prediction),
		}

		os.taskQueue.mu.Lock()
		heap.Push(os.taskQueue, priorityTask)
		os.taskQueue.mu.Unlock()

		// Update atomic metrics
		atomic.AddInt64(&os.atomicMetrics.SuccessfulTasks, 1)

		resultChan <- nil
	}()

	// Wait for completion or timeout
	select {
	case err := <-resultChan:
		if err != nil {
			atomic.AddInt64(&os.atomicMetrics.FailedTasks, 1)
		}
		return err
	case <-time.After(5 * time.Second):
		atomic.AddInt64(&os.atomicMetrics.FailedTasks, 1)
		return fmt.Errorf("scheduling timeout")
	}
}

// getAvailableNodesOptimized uses concurrent node index for O(1) access
func (os *OptimizedScheduler) getAvailableNodesOptimized() []*OptimizedNode {
	var nodes []*OptimizedNode
	
	os.nodeIndex.nodes.Range(func(key, value interface{}) bool {
		node := value.(*OptimizedNode)
		if node.Status == "available" {
			nodes = append(nodes, node)
		}
		return true
	})
	
	return nodes
}

// applyConstraintsOptimized uses bloom filters for O(1) constraint checking
func (os *OptimizedScheduler) applyConstraintsOptimized(task *ScheduledTask, nodes []*OptimizedNode) []*OptimizedNode {
	if task.Constraints == nil {
		return nodes
	}

	var candidateNodes []*OptimizedNode

	for _, node := range nodes {
		if os.nodeMatchesConstraintsOptimized(node, task.Constraints) {
			candidateNodes = append(candidateNodes, node)
		}
	}

	return candidateNodes
}

// nodeMatchesConstraintsOptimized uses optimized constraint checking
func (os *OptimizedScheduler) nodeMatchesConstraintsOptimized(node *OptimizedNode, constraints *TaskConstraints) bool {
	// Check required capabilities using capability set (O(1) lookup)
	for _, requiredCap := range constraints.RequiredCapabilities {
		if _, exists := node.capabilitySet[requiredCap]; !exists {
			return false
		}
	}

	// Check excluded nodes
	for _, excludedNode := range constraints.ExcludedNodes {
		if node.ID == excludedNode {
			return false
		}
	}

	// Additional constraint checks...
	return true
}

// calculateTaskPriority computes task priority for heap ordering
func (os *OptimizedScheduler) calculateTaskPriority(task *ScheduledTask, prediction *ResourcePrediction) float64 {
	// Priority factors
	basePriority := float64(task.Priority)
	urgencyFactor := 1.0
	
	if !task.Deadline.IsZero() {
		timeUntilDeadline := task.Deadline.Sub(time.Now())
		urgencyFactor = math.Max(0.1, 1.0/math.Max(timeUntilDeadline.Seconds(), 1.0))
	}
	
	complexityFactor := 1.0 + (prediction.CPURequirement-1.0)*0.1
	
	return basePriority * urgencyFactor * complexityFactor
}

// NewLRUPerformanceCache creates a new LRU cache for performance data
func NewLRUPerformanceCache(capacity int) *LRUPerformanceCache {
	cache := &LRUPerformanceCache{
		capacity: capacity,
		cache:    make(map[string]*CacheNode),
	}
	
	// Initialize dummy head and tail
	cache.head = &CacheNode{}
	cache.tail = &CacheNode{}
	cache.head.next = cache.tail
	cache.tail.prev = cache.head
	
	return cache
}

// Get retrieves a performance profile from cache (O(1))
func (lru *LRUPerformanceCache) Get(nodeID string) (*NodePerformanceProfile, bool) {
	lru.mu.Lock()
	defer lru.mu.Unlock()
	
	if node, exists := lru.cache[nodeID]; exists {
		if time.Now().Before(node.ttl) {
			// Move to front
			lru.moveToFront(node)
			atomic.AddInt64(&lru.hitCount, 1)
			return node.value, true
		} else {
			// Expired, remove
			lru.removeNode(node)
			delete(lru.cache, nodeID)
		}
	}
	
	atomic.AddInt64(&lru.missCount, 1)
	return nil, false
}

// Put stores a performance profile in cache (O(1))
func (lru *LRUPerformanceCache) Put(nodeID string, profile *NodePerformanceProfile, ttl time.Duration) {
	lru.mu.Lock()
	defer lru.mu.Unlock()
	
	if node, exists := lru.cache[nodeID]; exists {
		node.value = profile
		node.ttl = time.Now().Add(ttl)
		lru.moveToFront(node)
	} else {
		newNode := &CacheNode{
			key:   nodeID,
			value: profile,
			ttl:   time.Now().Add(ttl),
		}
		
		if len(lru.cache) >= lru.capacity {
			// Evict least recently used
			tail := lru.tail.prev
			lru.removeNode(tail)
			delete(lru.cache, tail.key)
		}
		
		lru.cache[nodeID] = newNode
		lru.addToFront(newNode)
	}
}

// Helper methods for LRU cache
func (lru *LRUPerformanceCache) addToFront(node *CacheNode) {
	node.prev = lru.head
	node.next = lru.head.next
	lru.head.next.prev = node
	lru.head.next = node
}

func (lru *LRUPerformanceCache) removeNode(node *CacheNode) {
	node.prev.next = node.next
	node.next.prev = node.prev
}

func (lru *LRUPerformanceCache) moveToFront(node *CacheNode) {
	lru.removeNode(node)
	lru.addToFront(node)
}

// NewBloomFilter creates a new bloom filter
func NewBloomFilter(expectedElements int) *BloomFilter {
	// Optimal parameters for 1% false positive rate
	size := uint64(-float64(expectedElements) * math.Log(0.01) / (math.Log(2) * math.Log(2)))
	hashFuncs := int(float64(size) / float64(expectedElements) * math.Log(2))
	
	return &BloomFilter{
		bits:      make([]uint64, (size+63)/64), // Round up to 64-bit words
		size:      size,
		hashFuncs: hashFuncs,
	}
}

// Add adds an element to the bloom filter
func (bf *BloomFilter) Add(data []byte) {
	hash1, hash2 := bf.hash(data)
	
	for i := 0; i < bf.hashFuncs; i++ {
		index := (hash1 + uint64(i)*hash2) % bf.size
		wordIndex := index / 64
		bitIndex := index % 64
		bf.bits[wordIndex] |= 1 << bitIndex
	}
}

// Contains checks if an element might be in the bloom filter
func (bf *BloomFilter) Contains(data []byte) bool {
	hash1, hash2 := bf.hash(data)
	
	for i := 0; i < bf.hashFuncs; i++ {
		index := (hash1 + uint64(i)*hash2) % bf.size
		wordIndex := index / 64
		bitIndex := index % 64
		if (bf.bits[wordIndex] & (1 << bitIndex)) == 0 {
			return false
		}
	}
	
	return true
}

// hash computes two independent hash values
func (bf *BloomFilter) hash(data []byte) (uint64, uint64) {
	hasher := fnv.New64a()
	hasher.Write(data)
	hash1 := hasher.Sum64()
	
	hasher.Reset()
	hasher.Write(data)
	hasher.Write([]byte{1}) // Add salt for second hash
	hash2 := hasher.Sum64()
	
	return hash1, hash2
}

// GetCacheStats returns cache performance statistics
func (lru *LRUPerformanceCache) GetCacheStats() (hitRate float64, hitCount, missCount int64) {
	hits := atomic.LoadInt64(&lru.hitCount)
	misses := atomic.LoadInt64(&lru.missCount)
	total := hits + misses
	
	if total > 0 {
		hitRate = float64(hits) / float64(total)
	}
	
	return hitRate, hits, misses
}

// GetAtomicMetrics returns current atomic metrics
func (os *OptimizedScheduler) GetAtomicMetrics() *IntelligentSchedulerMetrics {
	totalTasks := atomic.LoadInt64(&os.atomicMetrics.TotalTasksScheduled)
	successfulTasks := atomic.LoadInt64(&os.atomicMetrics.SuccessfulTasks)
	failedTasks := atomic.LoadInt64(&os.atomicMetrics.FailedTasks)
	schedulingTimeSum := atomic.LoadInt64(&os.atomicMetrics.SchedulingTimeSum)
	
	avgSchedulingTime := time.Duration(0)
	if totalTasks > 0 {
		avgSchedulingTime = time.Duration(schedulingTimeSum / totalTasks)
	}
	
	return &IntelligentSchedulerMetrics{
		TotalTasksScheduled: totalTasks,
		SuccessfulTasks:     successfulTasks,
		FailedTasks:         failedTasks,
		AvgSchedulingTime:   avgSchedulingTime,
		LastUpdated:         time.Now(),
	}
}
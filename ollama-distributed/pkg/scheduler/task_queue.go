package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// QueueMetrics tracks queue performance metrics
type QueueMetrics struct {
	TotalEnqueued       int64     `json:"total_enqueued"`
	TotalDequeued       int64     `json:"total_dequeued"`
	CurrentSize         int64     `json:"current_size"`
	HighPrioritySize    int64     `json:"high_priority_size"`
	NormalPrioritySize  int64     `json:"normal_priority_size"`
	LowPrioritySize     int64     `json:"low_priority_size"`
	AverageWaitTime     time.Duration `json:"average_wait_time"`
	MaxWaitTime         time.Duration `json:"max_wait_time"`
	LastUpdated         time.Time `json:"last_updated"`
	mu                  sync.RWMutex
}

// TaskExecutionMetrics tracks task execution metrics
type TaskExecutionMetrics struct {
	StartTime           time.Time     `json:"start_time"`
	EndTime             time.Time     `json:"end_time"`
	Duration            time.Duration `json:"duration"`
	QueueTime           time.Duration `json:"queue_time"`
	ExecutionTime       time.Duration `json:"execution_time"`
	CPUUsage            float64       `json:"cpu_usage"`
	MemoryUsage         int64         `json:"memory_usage"`
	NetworkUsage        int64         `json:"network_usage"`
	Success             bool          `json:"success"`
	ErrorCount          int           `json:"error_count"`
}

// NewTaskQueue creates a new task queue
func NewTaskQueue(config *TaskQueueConfig) (*TaskQueue, error) {
	if config == nil {
		config = &TaskQueueConfig{
			MaxSize:             10000,
			Timeout:             30 * time.Second,
			EnablePriority:      true,
			HighPriorityRatio:   0.3,
			NormalPriorityRatio: 0.5,
			LowPriorityRatio:    0.2,
		}
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	// Calculate queue sizes based on ratios
	highSize := int(float64(config.MaxSize) * config.HighPriorityRatio)
	normalSize := int(float64(config.MaxSize) * config.NormalPriorityRatio)
	lowSize := int(float64(config.MaxSize) * config.LowPriorityRatio)
	
	queue := &TaskQueue{
		config:              config,
		highPriorityQueue:   make(chan *Task, highSize),
		normalPriorityQueue: make(chan *Task, normalSize),
		lowPriorityQueue:    make(chan *Task, lowSize),
		metrics: &QueueMetrics{
			LastUpdated: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}
	
	return queue, nil
}

// Start starts the task queue
func (tq *TaskQueue) Start() error {
	// Start metrics collection
	tq.wg.Add(1)
	go tq.metricsLoop()
	
	return nil
}

// Stop stops the task queue
func (tq *TaskQueue) Stop() error {
	tq.cancel()
	tq.wg.Wait()
	
	// Close channels
	close(tq.highPriorityQueue)
	close(tq.normalPriorityQueue)
	close(tq.lowPriorityQueue)
	
	return nil
}

// Enqueue adds a task to the appropriate priority queue
func (tq *TaskQueue) Enqueue(task *Task) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}
	
	// Select queue based on priority
	var targetQueue chan *Task
	switch task.Priority {
	case TaskPriorityCritical, TaskPriorityHigh:
		targetQueue = tq.highPriorityQueue
	case TaskPriorityNormal:
		targetQueue = tq.normalPriorityQueue
	case TaskPriorityLow:
		targetQueue = tq.lowPriorityQueue
	default:
		targetQueue = tq.normalPriorityQueue
	}
	
	// Try to enqueue with timeout
	select {
	case targetQueue <- task:
		task.Status = TaskStatusQueued
		
		// Update metrics
		tq.metrics.mu.Lock()
		tq.metrics.TotalEnqueued++
		tq.metrics.CurrentSize++
		tq.updateQueueSizes()
		tq.metrics.LastUpdated = time.Now()
		tq.metrics.mu.Unlock()
		
		return nil
		
	case <-time.After(tq.config.Timeout):
		return fmt.Errorf("queue timeout: failed to enqueue task")
		
	case <-tq.ctx.Done():
		return fmt.Errorf("queue stopped")
	}
}

// Dequeue removes and returns the highest priority task
func (tq *TaskQueue) Dequeue() (*Task, error) {
	// Try high priority queue first
	select {
	case task := <-tq.highPriorityQueue:
		return tq.processDequeue(task), nil
	default:
	}
	
	// Try normal priority queue
	select {
	case task := <-tq.normalPriorityQueue:
		return tq.processDequeue(task), nil
	default:
	}
	
	// Try low priority queue
	select {
	case task := <-tq.lowPriorityQueue:
		return tq.processDequeue(task), nil
	default:
	}
	
	// No tasks available
	return nil, fmt.Errorf("no tasks available")
}

// processDequeue processes a dequeued task
func (tq *TaskQueue) processDequeue(task *Task) *Task {
	// Calculate wait time
	waitTime := time.Since(task.CreatedAt)
	
	// Update metrics
	tq.metrics.mu.Lock()
	tq.metrics.TotalDequeued++
	tq.metrics.CurrentSize--
	tq.updateQueueSizes()
	
	// Update wait time metrics
	if waitTime > tq.metrics.MaxWaitTime {
		tq.metrics.MaxWaitTime = waitTime
	}
	
	// Update average wait time (simple moving average)
	if tq.metrics.TotalDequeued == 1 {
		tq.metrics.AverageWaitTime = waitTime
	} else {
		tq.metrics.AverageWaitTime = (tq.metrics.AverageWaitTime + waitTime) / 2
	}
	
	tq.metrics.LastUpdated = time.Now()
	tq.metrics.mu.Unlock()
	
	return task
}

// updateQueueSizes updates the individual queue size metrics
func (tq *TaskQueue) updateQueueSizes() {
	tq.metrics.HighPrioritySize = int64(len(tq.highPriorityQueue))
	tq.metrics.NormalPrioritySize = int64(len(tq.normalPriorityQueue))
	tq.metrics.LowPrioritySize = int64(len(tq.lowPriorityQueue))
}

// GetMetrics returns current queue metrics
func (tq *TaskQueue) GetMetrics() *QueueMetrics {
	tq.metrics.mu.RLock()
	defer tq.metrics.mu.RUnlock()
	
	// Create a copy
	metrics := *tq.metrics
	return &metrics
}

// Size returns the current total queue size
func (tq *TaskQueue) Size() int64 {
	tq.metrics.mu.RLock()
	defer tq.metrics.mu.RUnlock()
	
	return tq.metrics.CurrentSize
}

// IsEmpty returns whether the queue is empty
func (tq *TaskQueue) IsEmpty() bool {
	return tq.Size() == 0
}

// IsFull returns whether the queue is full
func (tq *TaskQueue) IsFull() bool {
	return tq.Size() >= int64(tq.config.MaxSize)
}

// metricsLoop runs the metrics collection loop
func (tq *TaskQueue) metricsLoop() {
	defer tq.wg.Done()
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-tq.ctx.Done():
			return
		case <-ticker.C:
			tq.collectMetrics()
		}
	}
}

// collectMetrics collects and updates metrics
func (tq *TaskQueue) collectMetrics() {
	tq.metrics.mu.Lock()
	defer tq.metrics.mu.Unlock()
	
	// Update queue sizes
	tq.updateQueueSizes()
	tq.metrics.LastUpdated = time.Now()
}

// Clear clears all tasks from the queue
func (tq *TaskQueue) Clear() {
	// Drain all queues
	for {
		select {
		case <-tq.highPriorityQueue:
		default:
			goto drainNormal
		}
	}
	
drainNormal:
	for {
		select {
		case <-tq.normalPriorityQueue:
		default:
			goto drainLow
		}
	}
	
drainLow:
	for {
		select {
		case <-tq.lowPriorityQueue:
		default:
			goto updateMetrics
		}
	}
	
updateMetrics:
	// Reset metrics
	tq.metrics.mu.Lock()
	tq.metrics.CurrentSize = 0
	tq.metrics.HighPrioritySize = 0
	tq.metrics.NormalPrioritySize = 0
	tq.metrics.LowPrioritySize = 0
	tq.metrics.LastUpdated = time.Now()
	tq.metrics.mu.Unlock()
}

// GetQueueSizes returns the sizes of individual priority queues
func (tq *TaskQueue) GetQueueSizes() (high, normal, low int64) {
	tq.metrics.mu.RLock()
	defer tq.metrics.mu.RUnlock()
	
	return tq.metrics.HighPrioritySize, tq.metrics.NormalPrioritySize, tq.metrics.LowPrioritySize
}

// SetPriorityRatios updates the priority ratios (requires restart to take effect)
func (tq *TaskQueue) SetPriorityRatios(high, normal, low float64) error {
	if high+normal+low != 1.0 {
		return fmt.Errorf("priority ratios must sum to 1.0")
	}
	
	tq.config.HighPriorityRatio = high
	tq.config.NormalPriorityRatio = normal
	tq.config.LowPriorityRatio = low
	
	return nil
}

package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// TaskMetrics tracks task execution metrics
type TaskMetrics struct {
	TotalTasks          int64         `json:"total_tasks"`
	ActiveTasks         int64         `json:"active_tasks"`
	CompletedTasks      int64         `json:"completed_tasks"`
	FailedTasks         int64         `json:"failed_tasks"`
	CancelledTasks      int64         `json:"cancelled_tasks"`
	AverageExecutionTime time.Duration `json:"average_execution_time"`
	AverageQueueTime    time.Duration `json:"average_queue_time"`
	SuccessRate         float64       `json:"success_rate"`
	LastUpdated         time.Time     `json:"last_updated"`
	mu                  sync.RWMutex
}

// NewTaskTracker creates a new task tracker
func NewTaskTracker(config *TaskTrackerConfig) (*TaskTracker, error) {
	if config == nil {
		config = &TaskTrackerConfig{
			MaxActiveTasks:   10000,
			TaskTimeout:      30 * time.Minute,
			ResultBufferSize: 1000,
			CleanupInterval:  5 * time.Minute,
		}
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	tracker := &TaskTracker{
		config:      config,
		activeTasks: make(map[string]*TrackedTask),
		results:     make(chan *TaskResult, config.ResultBufferSize),
		metrics: &TaskMetrics{
			LastUpdated: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}
	
	return tracker, nil
}

// Start starts the task tracker
func (tt *TaskTracker) Start() error {
	// Start result processing
	tt.wg.Add(1)
	go tt.resultProcessor()
	
	// Start cleanup routine
	tt.wg.Add(1)
	go tt.cleanupLoop()
	
	// Start metrics collection
	tt.wg.Add(1)
	go tt.metricsLoop()
	
	return nil
}

// Stop stops the task tracker
func (tt *TaskTracker) Stop() error {
	tt.cancel()
	tt.wg.Wait()
	
	// Close results channel
	close(tt.results)
	
	return nil
}

// TrackTask starts tracking a task
func (tt *TaskTracker) TrackTask(task *Task, worker *WorkerNode) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}
	
	if worker == nil {
		return fmt.Errorf("worker cannot be nil")
	}
	
	tt.activeTasksMu.Lock()
	defer tt.activeTasksMu.Unlock()
	
	// Check if we've reached the maximum number of active tasks
	if len(tt.activeTasks) >= tt.config.MaxActiveTasks {
		return fmt.Errorf("maximum number of active tasks reached")
	}
	
	// Create tracked task
	trackedTask := &TrackedTask{
		Task:       task,
		Worker:     worker,
		StartTime:  time.Now(),
		LastUpdate: time.Now(),
		Progress:   0.0,
		Status:     TaskStatusRunning,
		Heartbeats: make([]time.Time, 0),
	}
	
	// Add to active tasks
	tt.activeTasks[task.ID] = trackedTask
	
	// Update task status
	task.Status = TaskStatusRunning
	task.StartedAt = time.Now()
	
	// Update metrics
	tt.updateMetrics()
	
	return nil
}

// UntrackTask stops tracking a task
func (tt *TaskTracker) UntrackTask(taskID string) error {
	tt.activeTasksMu.Lock()
	defer tt.activeTasksMu.Unlock()
	
	_, exists := tt.activeTasks[taskID]
	if !exists {
		return fmt.Errorf("task not being tracked")
	}
	
	delete(tt.activeTasks, taskID)
	
	// Update metrics
	tt.updateMetrics()
	
	return nil
}

// UpdateTaskProgress updates the progress of a tracked task
func (tt *TaskTracker) UpdateTaskProgress(taskID string, progress float64) error {
	tt.activeTasksMu.Lock()
	defer tt.activeTasksMu.Unlock()
	
	trackedTask, exists := tt.activeTasks[taskID]
	if !exists {
		return fmt.Errorf("task not being tracked")
	}
	
	trackedTask.Progress = progress
	trackedTask.LastUpdate = time.Now()
	
	// Add heartbeat
	trackedTask.Heartbeats = append(trackedTask.Heartbeats, time.Now())
	
	// Keep only recent heartbeats (last 10)
	if len(trackedTask.Heartbeats) > 10 {
		trackedTask.Heartbeats = trackedTask.Heartbeats[len(trackedTask.Heartbeats)-10:]
	}
	
	return nil
}

// CompleteTask marks a task as completed
func (tt *TaskTracker) CompleteTask(taskID string, result []byte) error {
	tt.activeTasksMu.Lock()
	trackedTask, exists := tt.activeTasks[taskID]
	if !exists {
		tt.activeTasksMu.Unlock()
		return fmt.Errorf("task not being tracked")
	}
	
	// Remove from active tasks
	delete(tt.activeTasks, taskID)
	tt.activeTasksMu.Unlock()
	
	// Create task result
	taskResult := &TaskResult{
		TaskID:      taskID,
		WorkerID:    trackedTask.Worker.ID,
		Status:      TaskStatusCompleted,
		Result:      result,
		CompletedAt: time.Now(),
		Duration:    time.Since(trackedTask.StartTime),
		Metrics: &TaskExecutionMetrics{
			StartTime:     trackedTask.StartTime,
			EndTime:       time.Now(),
			Duration:      time.Since(trackedTask.StartTime),
			QueueTime:     trackedTask.Task.ScheduledAt.Sub(trackedTask.Task.CreatedAt),
			ExecutionTime: time.Since(trackedTask.StartTime),
			Success:       true,
		},
	}
	
	// Update task status
	trackedTask.Task.Status = TaskStatusCompleted
	trackedTask.Task.CompletedAt = time.Now()
	
	// Send result
	select {
	case tt.results <- taskResult:
	case <-time.After(5 * time.Second):
		// Result buffer full, log warning but don't fail
	}
	
	// Update metrics
	tt.updateMetrics()
	
	return nil
}

// FailTask marks a task as failed
func (tt *TaskTracker) FailTask(taskID string, errorMsg string) error {
	tt.activeTasksMu.Lock()
	trackedTask, exists := tt.activeTasks[taskID]
	if !exists {
		tt.activeTasksMu.Unlock()
		return fmt.Errorf("task not being tracked")
	}
	
	// Remove from active tasks
	delete(tt.activeTasks, taskID)
	tt.activeTasksMu.Unlock()
	
	// Create task result
	taskResult := &TaskResult{
		TaskID:      taskID,
		WorkerID:    trackedTask.Worker.ID,
		Status:      TaskStatusFailed,
		Error:       errorMsg,
		CompletedAt: time.Now(),
		Duration:    time.Since(trackedTask.StartTime),
		Metrics: &TaskExecutionMetrics{
			StartTime:     trackedTask.StartTime,
			EndTime:       time.Now(),
			Duration:      time.Since(trackedTask.StartTime),
			QueueTime:     trackedTask.Task.ScheduledAt.Sub(trackedTask.Task.CreatedAt),
			ExecutionTime: time.Since(trackedTask.StartTime),
			Success:       false,
			ErrorCount:    1,
		},
	}
	
	// Update task status
	trackedTask.Task.Status = TaskStatusFailed
	trackedTask.Task.Error = errorMsg
	trackedTask.Task.CompletedAt = time.Now()
	
	// Send result
	select {
	case tt.results <- taskResult:
	case <-time.After(5 * time.Second):
		// Result buffer full, log warning but don't fail
	}
	
	// Update metrics
	tt.updateMetrics()
	
	return nil
}

// GetTrackedTask returns a tracked task by ID
func (tt *TaskTracker) GetTrackedTask(taskID string) (*TrackedTask, bool) {
	tt.activeTasksMu.RLock()
	defer tt.activeTasksMu.RUnlock()
	
	task, exists := tt.activeTasks[taskID]
	return task, exists
}

// GetAllTrackedTasks returns all currently tracked tasks
func (tt *TaskTracker) GetAllTrackedTasks() []*TrackedTask {
	tt.activeTasksMu.RLock()
	defer tt.activeTasksMu.RUnlock()
	
	tasks := make([]*TrackedTask, 0, len(tt.activeTasks))
	for _, task := range tt.activeTasks {
		tasks = append(tasks, task)
	}
	
	return tasks
}

// GetTasksByWorker returns all tasks assigned to a specific worker
func (tt *TaskTracker) GetTasksByWorker(workerID peer.ID) []*TrackedTask {
	tt.activeTasksMu.RLock()
	defer tt.activeTasksMu.RUnlock()
	
	var tasks []*TrackedTask
	for _, task := range tt.activeTasks {
		if task.Worker.ID == workerID {
			tasks = append(tasks, task)
		}
	}
	
	return tasks
}

// GetMetrics returns current task metrics
func (tt *TaskTracker) GetMetrics() *TaskMetrics {
	tt.metrics.mu.RLock()
	defer tt.metrics.mu.RUnlock()
	
	// Create a copy
	metrics := *tt.metrics
	return &metrics
}

// GetResults returns the results channel for consuming task results
func (tt *TaskTracker) GetResults() <-chan *TaskResult {
	return tt.results
}

// resultProcessor processes task results
func (tt *TaskTracker) resultProcessor() {
	defer tt.wg.Done()
	
	for {
		select {
		case <-tt.ctx.Done():
			return
		case result := <-tt.results:
			if result != nil {
				tt.processResult(result)
			}
		}
	}
}

// processResult processes a single task result
func (tt *TaskTracker) processResult(result *TaskResult) {
	// Update metrics based on result
	tt.metrics.mu.Lock()
	defer tt.metrics.mu.Unlock()
	
	switch result.Status {
	case TaskStatusCompleted:
		tt.metrics.CompletedTasks++
	case TaskStatusFailed:
		tt.metrics.FailedTasks++
	case TaskStatusCancelled:
		tt.metrics.CancelledTasks++
	}
	
	// Update average execution time
	if result.Metrics != nil {
		if tt.metrics.CompletedTasks == 1 {
			tt.metrics.AverageExecutionTime = result.Metrics.ExecutionTime
			tt.metrics.AverageQueueTime = result.Metrics.QueueTime
		} else {
			tt.metrics.AverageExecutionTime = (tt.metrics.AverageExecutionTime + result.Metrics.ExecutionTime) / 2
			tt.metrics.AverageQueueTime = (tt.metrics.AverageQueueTime + result.Metrics.QueueTime) / 2
		}
	}
	
	// Update success rate
	totalCompleted := tt.metrics.CompletedTasks + tt.metrics.FailedTasks + tt.metrics.CancelledTasks
	if totalCompleted > 0 {
		tt.metrics.SuccessRate = float64(tt.metrics.CompletedTasks) / float64(totalCompleted)
	}
	
	tt.metrics.LastUpdated = time.Now()
}

// cleanupLoop runs the cleanup routine
func (tt *TaskTracker) cleanupLoop() {
	defer tt.wg.Done()
	
	ticker := time.NewTicker(tt.config.CleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-tt.ctx.Done():
			return
		case <-ticker.C:
			tt.cleanupTimedOutTasks()
		}
	}
}

// cleanupTimedOutTasks removes tasks that have timed out
func (tt *TaskTracker) cleanupTimedOutTasks() {
	tt.activeTasksMu.Lock()
	defer tt.activeTasksMu.Unlock()
	
	now := time.Now()
	var timedOutTasks []string
	
	for taskID, trackedTask := range tt.activeTasks {
		if now.Sub(trackedTask.StartTime) > tt.config.TaskTimeout {
			timedOutTasks = append(timedOutTasks, taskID)
		}
	}
	
	// Remove timed out tasks
	for _, taskID := range timedOutTasks {
		trackedTask := tt.activeTasks[taskID]
		delete(tt.activeTasks, taskID)
		
		// Create timeout result
		result := &TaskResult{
			TaskID:      taskID,
			WorkerID:    trackedTask.Worker.ID,
			Status:      TaskStatusFailed,
			Error:       "task timeout",
			CompletedAt: now,
			Duration:    now.Sub(trackedTask.StartTime),
		}
		
		// Send timeout result
		select {
		case tt.results <- result:
		default:
			// Result buffer full, skip
		}
	}
}

// updateMetrics updates task metrics
func (tt *TaskTracker) updateMetrics() {
	tt.metrics.mu.Lock()
	defer tt.metrics.mu.Unlock()
	
	tt.metrics.ActiveTasks = int64(len(tt.activeTasks))
	tt.metrics.TotalTasks = tt.metrics.CompletedTasks + tt.metrics.FailedTasks + tt.metrics.CancelledTasks + tt.metrics.ActiveTasks
	tt.metrics.LastUpdated = time.Now()
}

// metricsLoop runs the metrics collection loop
func (tt *TaskTracker) metricsLoop() {
	defer tt.wg.Done()
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-tt.ctx.Done():
			return
		case <-ticker.C:
			tt.updateMetrics()
		}
	}
}

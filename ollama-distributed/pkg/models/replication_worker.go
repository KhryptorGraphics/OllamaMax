package models

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"
)

// ReplicationWorkerImpl handles replication tasks (implementation)
type ReplicationWorkerImpl struct {
	ID       int
	manager  *ReplicationManager
	stopChan chan struct{}
	logger   *slog.Logger
	busy     int32 // atomic flag for busy state
}

// NewReplicationWorker creates a new replication worker
func NewReplicationWorker(id int, manager *ReplicationManager, logger *slog.Logger) *ReplicationWorkerImpl {
	return &ReplicationWorkerImpl{
		ID:       id,
		manager:  manager,
		stopChan: make(chan struct{}),
		logger:   logger.With("worker_id", id),
	}
}

// Start begins processing replication tasks
func (rw *ReplicationWorker) Start(ctx context.Context) {
	rw.logger.Info("Replication worker started")
	
	for {
		select {
		case <-ctx.Done():
			rw.logger.Info("Replication worker stopped due to context cancellation")
			return
		case <-rw.stopChan:
			rw.logger.Info("Replication worker stopped")
			return
		case task := <-rw.manager.workQueue:
			if task != nil {
				rw.processTaskImpl(ctx, task)
			}
		}
	}
}

// Stop stops the replication worker
func (rw *ReplicationWorker) Stop() {
	close(rw.stopChan)
}

// IsBusy returns whether the worker is currently processing a task
func (rw *ReplicationWorker) IsBusy() bool {
	return atomic.LoadInt32(&rw.busy) == 1
}

// processTaskImpl processes a replication task (implementation)
func (rw *ReplicationWorkerImpl) processTaskImpl(ctx context.Context, task *ReplicationTask) {
	if task == nil {
		return
	}

	// Set busy flag
	atomic.StoreInt32(&rw.busy, 1)
	defer atomic.StoreInt32(&rw.busy, 0)

	rw.logger.Info("Processing replication task",
		"task_id", task.ID,
		"type", task.Type,
		"model", task.ModelName,
		"source", task.SourcePeer,
		"target", task.TargetPeer,
	)

	// Update task status
	task.Status = TaskStatusRunning
	task.UpdatedAt = time.Now()

	// Process the task based on its type
	var err error
	switch task.Type {
	case TaskTypeSync:
		err = rw.syncReplica(ctx, task)
	case TaskTypeCopy:
		err = rw.copyReplica(ctx, task)
	case TaskTypeDelete:
		err = rw.deleteReplica(ctx, task)
	case TaskTypeVerify:
		err = rw.verifyReplica(ctx, task)
	default:
		err = fmt.Errorf("unknown task type: %s", task.Type)
	}

	// Update task completion status
	now := time.Now()
	if err != nil {
		task.Status = TaskStatusFailed
		task.Error = err.Error()
		rw.logger.Error("Replication task failed",
			"task_id", task.ID,
			"error", err,
		)
		
		// Update failure metrics
		atomic.AddInt64(&rw.manager.failedReplications, 1)
	} else {
		task.Status = TaskStatusCompleted
		task.Progress = 100.0
		task.CompletedAt = &now
		rw.logger.Info("Replication task completed successfully",
			"task_id", task.ID,
		)
		
		// Update success metrics
		atomic.AddInt64(&rw.manager.successfulReplications, 1)
	}
	
	task.UpdatedAt = now

	// Send response if channel is provided
	if task.ResponseChan != nil {
		select {
		case task.ResponseChan <- err:
		default:
			// Channel might be closed or full
		}
	}
}

// syncReplica synchronizes a replica between peers
func (rw *ReplicationWorker) syncReplica(ctx context.Context, task *ReplicationTask) error {
	rw.logger.Debug("Syncing replica",
		"model", task.ModelName,
		"source", task.SourcePeer,
		"target", task.TargetPeer,
	)

	// Implementation would perform actual synchronization
	// For now, simulate work and return success
	
	// Simulate sync work
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(100 * time.Millisecond):
		// Simulated sync completed
	}

	task.Progress = 100.0
	return nil
}

// copyReplica copies a replica to a new peer
func (rw *ReplicationWorker) copyReplica(ctx context.Context, task *ReplicationTask) error {
	rw.logger.Debug("Copying replica",
		"model", task.ModelName,
		"source", task.SourcePeer,
		"target", task.TargetPeer,
	)

	// Implementation would perform actual copying
	// For now, simulate work and return success
	
	// Simulate copy work with progress updates
	for progress := 0.0; progress <= 100.0; progress += 25.0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
			task.Progress = progress
		}
	}

	return nil
}

// deleteReplica deletes a replica from a peer
func (rw *ReplicationWorker) deleteReplica(ctx context.Context, task *ReplicationTask) error {
	rw.logger.Debug("Deleting replica",
		"model", task.ModelName,
		"target", task.TargetPeer,
	)

	// Implementation would perform actual deletion
	// For now, simulate work and return success
	
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(50 * time.Millisecond):
		// Simulated deletion completed
	}

	task.Progress = 100.0
	return nil
}

// verifyReplica verifies the integrity of a replica
func (rw *ReplicationWorker) verifyReplica(ctx context.Context, task *ReplicationTask) error {
	rw.logger.Debug("Verifying replica",
		"model", task.ModelName,
		"target", task.TargetPeer,
	)

	// Implementation would perform actual verification
	// For now, simulate work and return success
	
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(75 * time.Millisecond):
		// Simulated verification completed
	}

	task.Progress = 100.0
	return nil
}
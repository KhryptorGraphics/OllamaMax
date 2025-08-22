package models

import (
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// ModelLifecycle manages the lifecycle of distributed models
type ModelLifecycle struct {
	events      chan *LifecycleEvent
	eventsMutex sync.RWMutex

	// Lifecycle stages
	stages      map[string]*LifecycleStage
	stagesMutex sync.RWMutex

	// Event hooks
	hooks      map[LifecycleEventType][]LifecycleHook
	hooksMutex sync.RWMutex

	logger *slog.Logger
}

// LifecycleEvent represents a model lifecycle event
type LifecycleEvent struct {
	Type      LifecycleEventType     `json:"type"`
	ModelName string                 `json:"model_name"`
	PeerID    string                 `json:"peer_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// LifecycleEventType represents the type of lifecycle event
type LifecycleEventType string

const (
	EventModelAdded      LifecycleEventType = "model_added"
	EventModelUpdated    LifecycleEventType = "model_updated"
	EventModelDeleted    LifecycleEventType = "model_deleted"
	EventModelSynced     LifecycleEventType = "model_synced"
	EventModelReplicated LifecycleEventType = "model_replicated"
	EventModelAccessed   LifecycleEventType = "model_accessed"
	EventModelCorrupted  LifecycleEventType = "model_corrupted"
	EventModelHealed     LifecycleEventType = "model_healed"
)

// LifecycleStage represents a stage in the model lifecycle
type LifecycleStage struct {
	Name      string                 `json:"name"`
	ModelName string                 `json:"model_name"`
	Status    StageStatus            `json:"status"`
	StartTime time.Time              `json:"start_time"`
	EndTime   time.Time              `json:"end_time"`
	Duration  time.Duration          `json:"duration"`
	Progress  float64                `json:"progress"`
	Metadata  map[string]interface{} `json:"metadata"`
	Errors    []string               `json:"errors"`
}

// StageStatus represents the status of a lifecycle stage
type StageStatus string

const (
	StageStatusPending    StageStatus = "pending"
	StageStatusInProgress StageStatus = "in_progress"
	StageStatusCompleted  StageStatus = "completed"
	StageStatusFailed     StageStatus = "failed"
	StageStatusSkipped    StageStatus = "skipped"
)

// LifecycleHook represents a hook function for lifecycle events
type LifecycleHook func(event *LifecycleEvent) error

// NewModelLifecycle creates a new model lifecycle manager
func NewModelLifecycle(logger *slog.Logger) *ModelLifecycle {
	return &ModelLifecycle{
		events: make(chan *LifecycleEvent, 1000),
		stages: make(map[string]*LifecycleStage),
		hooks:  make(map[LifecycleEventType][]LifecycleHook),
		logger: logger,
	}
}

// start starts the lifecycle manager
func (ml *ModelLifecycle) start() {
	ml.logger.Info("model lifecycle manager started")

	for event := range ml.events {
		ml.processEvent(event)
	}
}

// processEvent processes a lifecycle event
func (ml *ModelLifecycle) processEvent(event *LifecycleEvent) {
	ml.logger.Debug("processing lifecycle event",
		"type", event.Type,
		"model", event.ModelName,
		"peer", event.PeerID)

	// Execute hooks
	ml.hooksMutex.RLock()
	hooks := ml.hooks[event.Type]
	ml.hooksMutex.RUnlock()

	for _, hook := range hooks {
		if err := hook(event); err != nil {
			ml.logger.Error("lifecycle hook failed",
				"type", event.Type,
				"model", event.ModelName,
				"error", err)
		}
	}

	// Update stages based on event
	ml.updateStagesFromEvent(event)
}

// AddHook adds a lifecycle hook for a specific event type
func (ml *ModelLifecycle) AddHook(eventType LifecycleEventType, hook LifecycleHook) {
	ml.hooksMutex.Lock()
	defer ml.hooksMutex.Unlock()

	ml.hooks[eventType] = append(ml.hooks[eventType], hook)
}

// RemoveHook removes all hooks for a specific event type
func (ml *ModelLifecycle) RemoveHook(eventType LifecycleEventType) {
	ml.hooksMutex.Lock()
	defer ml.hooksMutex.Unlock()

	delete(ml.hooks, eventType)
}

// StartStage starts a new lifecycle stage for a model
func (ml *ModelLifecycle) StartStage(modelName, stageName string, metadata map[string]interface{}) *LifecycleStage {
	ml.stagesMutex.Lock()
	defer ml.stagesMutex.Unlock()

	stageKey := modelName + ":" + stageName
	stage := &LifecycleStage{
		Name:      stageName,
		ModelName: modelName,
		Status:    StageStatusInProgress,
		StartTime: time.Now(),
		Progress:  0.0,
		Metadata:  metadata,
		Errors:    []string{},
	}

	ml.stages[stageKey] = stage
	ml.logger.Info("lifecycle stage started",
		"model", modelName,
		"stage", stageName)

	return stage
}

// CompleteStage marks a lifecycle stage as completed
func (ml *ModelLifecycle) CompleteStage(modelName, stageName string) error {
	ml.stagesMutex.Lock()
	defer ml.stagesMutex.Unlock()

	stageKey := modelName + ":" + stageName
	stage, exists := ml.stages[stageKey]
	if !exists {
		return fmt.Errorf("stage not found: %s", stageKey)
	}

	stage.Status = StageStatusCompleted
	stage.EndTime = time.Now()
	stage.Duration = stage.EndTime.Sub(stage.StartTime)
	stage.Progress = 1.0

	ml.logger.Info("lifecycle stage completed",
		"model", modelName,
		"stage", stageName,
		"duration", stage.Duration)

	return nil
}

// FailStage marks a lifecycle stage as failed
func (ml *ModelLifecycle) FailStage(modelName, stageName string, err error) error {
	ml.stagesMutex.Lock()
	defer ml.stagesMutex.Unlock()

	stageKey := modelName + ":" + stageName
	stage, exists := ml.stages[stageKey]
	if !exists {
		return fmt.Errorf("stage not found: %s", stageKey)
	}

	stage.Status = StageStatusFailed
	stage.EndTime = time.Now()
	stage.Duration = stage.EndTime.Sub(stage.StartTime)
	stage.Errors = append(stage.Errors, err.Error())

	ml.logger.Error("lifecycle stage failed",
		"model", modelName,
		"stage", stageName,
		"duration", stage.Duration,
		"error", err)

	return nil
}

// UpdateStageProgress updates the progress of a lifecycle stage
func (ml *ModelLifecycle) UpdateStageProgress(modelName, stageName string, progress float64) error {
	ml.stagesMutex.Lock()
	defer ml.stagesMutex.Unlock()

	stageKey := modelName + ":" + stageName
	stage, exists := ml.stages[stageKey]
	if !exists {
		return fmt.Errorf("stage not found: %s", stageKey)
	}

	stage.Progress = progress
	ml.logger.Debug("lifecycle stage progress updated",
		"model", modelName,
		"stage", stageName,
		"progress", progress)

	return nil
}

// GetStage retrieves a lifecycle stage
func (ml *ModelLifecycle) GetStage(modelName, stageName string) (*LifecycleStage, error) {
	ml.stagesMutex.RLock()
	defer ml.stagesMutex.RUnlock()

	stageKey := modelName + ":" + stageName
	stage, exists := ml.stages[stageKey]
	if !exists {
		return nil, fmt.Errorf("stage not found: %s", stageKey)
	}

	// Return a copy to avoid race conditions
	stageCopy := *stage
	return &stageCopy, nil
}

// GetModelStages retrieves all stages for a specific model
func (ml *ModelLifecycle) GetModelStages(modelName string) []*LifecycleStage {
	ml.stagesMutex.RLock()
	defer ml.stagesMutex.RUnlock()

	var stages []*LifecycleStage
	for _, stage := range ml.stages {
		if stage.ModelName == modelName {
			stageCopy := *stage
			stages = append(stages, &stageCopy)
		}
	}

	return stages
}

// GetAllStages retrieves all lifecycle stages
func (ml *ModelLifecycle) GetAllStages() []*LifecycleStage {
	ml.stagesMutex.RLock()
	defer ml.stagesMutex.RUnlock()

	stages := make([]*LifecycleStage, 0, len(ml.stages))
	for _, stage := range ml.stages {
		stageCopy := *stage
		stages = append(stages, &stageCopy)
	}

	return stages
}

// CleanupCompletedStages removes completed stages older than the specified duration
func (ml *ModelLifecycle) CleanupCompletedStages(maxAge time.Duration) {
	ml.stagesMutex.Lock()
	defer ml.stagesMutex.Unlock()

	cutoff := time.Now().Add(-maxAge)
	var removedCount int

	for key, stage := range ml.stages {
		if (stage.Status == StageStatusCompleted || stage.Status == StageStatusFailed) &&
			stage.EndTime.Before(cutoff) {
			delete(ml.stages, key)
			removedCount++
		}
	}

	if removedCount > 0 {
		ml.logger.Info("cleaned up completed lifecycle stages",
			"removed_count", removedCount,
			"max_age", maxAge)
	}
}

// GetStageStatistics returns statistics about lifecycle stages
func (ml *ModelLifecycle) GetStageStatistics() map[string]interface{} {
	ml.stagesMutex.RLock()
	defer ml.stagesMutex.RUnlock()

	stats := map[string]interface{}{
		"total_stages":     len(ml.stages),
		"pending_stages":   0,
		"in_progress":      0,
		"completed_stages": 0,
		"failed_stages":    0,
		"skipped_stages":   0,
	}

	for _, stage := range ml.stages {
		switch stage.Status {
		case StageStatusPending:
			stats["pending_stages"] = stats["pending_stages"].(int) + 1
		case StageStatusInProgress:
			stats["in_progress"] = stats["in_progress"].(int) + 1
		case StageStatusCompleted:
			stats["completed_stages"] = stats["completed_stages"].(int) + 1
		case StageStatusFailed:
			stats["failed_stages"] = stats["failed_stages"].(int) + 1
		case StageStatusSkipped:
			stats["skipped_stages"] = stats["skipped_stages"].(int) + 1
		}
	}

	return stats
}

// updateStagesFromEvent updates stages based on lifecycle events
func (ml *ModelLifecycle) updateStagesFromEvent(event *LifecycleEvent) {
	switch event.Type {
	case EventModelAdded:
		ml.StartStage(event.ModelName, "model_added", event.Data)
		ml.CompleteStage(event.ModelName, "model_added")

	case EventModelSynced:
		ml.StartStage(event.ModelName, "model_sync", event.Data)
		ml.CompleteStage(event.ModelName, "model_sync")

	case EventModelReplicated:
		ml.StartStage(event.ModelName, "model_replication", event.Data)
		ml.CompleteStage(event.ModelName, "model_replication")

	case EventModelDeleted:
		ml.StartStage(event.ModelName, "model_deletion", event.Data)
		ml.CompleteStage(event.ModelName, "model_deletion")

	case EventModelCorrupted:
		stage := ml.StartStage(event.ModelName, "model_corruption", event.Data)
		ml.FailStage(event.ModelName, "model_corruption", fmt.Errorf("model corrupted"))
		_ = stage

	case EventModelHealed:
		ml.StartStage(event.ModelName, "model_healing", event.Data)
		ml.CompleteStage(event.ModelName, "model_healing")
	}
}

// EmitEvent emits a lifecycle event
func (ml *ModelLifecycle) EmitEvent(eventType LifecycleEventType, modelName, peerID string, data map[string]interface{}) {
	event := &LifecycleEvent{
		Type:      eventType,
		ModelName: modelName,
		PeerID:    peerID,
		Timestamp: time.Now(),
		Data:      data,
	}

	select {
	case ml.events <- event:
	default:
		ml.logger.Warn("lifecycle event queue full",
			"event", eventType,
			"model", modelName)
	}
}

// GetEventHistory returns recent lifecycle events (simplified implementation)
func (ml *ModelLifecycle) GetEventHistory(limit int) []*LifecycleEvent {
	// In a real implementation, this would maintain an event history
	// For now, return empty slice
	return []*LifecycleEvent{}
}

// IsStageActive checks if a stage is currently active (in progress)
func (ml *ModelLifecycle) IsStageActive(modelName, stageName string) bool {
	stage, err := ml.GetStage(modelName, stageName)
	if err != nil {
		return false
	}
	return stage.Status == StageStatusInProgress
}

// GetActiveStages returns all currently active stages
func (ml *ModelLifecycle) GetActiveStages() []*LifecycleStage {
	ml.stagesMutex.RLock()
	defer ml.stagesMutex.RUnlock()

	var activeStages []*LifecycleStage
	for _, stage := range ml.stages {
		if stage.Status == StageStatusInProgress {
			stageCopy := *stage
			activeStages = append(activeStages, &stageCopy)
		}
	}

	return activeStages
}

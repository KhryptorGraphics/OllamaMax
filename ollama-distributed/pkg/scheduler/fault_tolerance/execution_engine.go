package fault_tolerance

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// ExecutionEngine executes recovery plans
type ExecutionEngine struct {
	config           *ExecutionEngineConfig
	activeExecutions map[string]*StepExecution
	executionPool    chan struct{}
	mu               sync.RWMutex
}

// ExecutionEngineConfig configures the execution engine
type ExecutionEngineConfig struct {
	MaxConcurrent           int           `json:"max_concurrent"`
	ExecutionTimeout        time.Duration `json:"execution_timeout"`
	EnableParallelExecution bool          `json:"enable_parallel_execution"`
	EnableProgressTracking  bool          `json:"enable_progress_tracking"`
}

// StepExecution represents an active step execution
type StepExecution struct {
	StepID    string                 `json:"step_id"`
	Status    StepStatus             `json:"status"`
	StartTime time.Time              `json:"start_time"`
	EndTime   time.Time              `json:"end_time"`
	Result    interface{}            `json:"result,omitempty"`
	Error     error                  `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// NewExecutionEngine creates a new execution engine
func NewExecutionEngine(config *ExecutionEngineConfig) *ExecutionEngine {
	if config == nil {
		config = &ExecutionEngineConfig{
			MaxConcurrent:           5,
			ExecutionTimeout:        8 * time.Minute,
			EnableParallelExecution: true,
			EnableProgressTracking:  true,
		}
	}

	return &ExecutionEngine{
		config:           config,
		activeExecutions: make(map[string]*StepExecution),
		executionPool:    make(chan struct{}, config.MaxConcurrent),
	}
}

// ExecutePlan executes a recovery plan
func (ee *ExecutionEngine) ExecutePlan(ctx context.Context, plan *RecoveryPlan, progress *RecoveryProgress) (*RecoveryResult, error) {
	log.Info().
		Str("plan_id", plan.ID).
		Int("steps", len(plan.Steps)).
		Msg("Starting plan execution")

	// Initialize execution context
	executionCtx, cancel := context.WithTimeout(ctx, ee.config.ExecutionTimeout)
	defer cancel()

	// Execute steps based on dependencies and parallelization
	if ee.config.EnableParallelExecution {
		return ee.executeParallel(executionCtx, plan, progress)
	} else {
		return ee.executeSequential(executionCtx, plan, progress)
	}
}

// ExecuteRollback executes a rollback plan
func (ee *ExecutionEngine) ExecuteRollback(ctx context.Context, rollback *RollbackPlan) (*RecoveryResult, error) {
	log.Info().
		Str("rollback_id", rollback.ID).
		Int("steps", len(rollback.Steps)).
		Msg("Starting rollback execution")

	// Execute rollback steps sequentially
	successfulSteps := 0
	var lastError error

	for _, step := range rollback.Steps {
		stepCtx, cancel := context.WithTimeout(ctx, step.Timeout)

		result, err := ee.executeStep(stepCtx, step)
		cancel()

		if err != nil {
			lastError = err
			log.Error().
				Err(err).
				Str("step_id", step.ID).
				Msg("Rollback step failed")
			break
		} else {
			successfulSteps++
			log.Debug().
				Str("step_id", step.ID).
				Msg("Rollback step completed")
		}

		// Update step result
		step.Result = result
		step.Status = StepStatusCompleted
	}

	// Create rollback result
	result := &RecoveryResult{
		FaultID:    rollback.ID,
		Strategy:   "rollback",
		Successful: lastError == nil,
		Duration:   rollback.Timeout,
		Metadata: map[string]interface{}{
			"rollback_steps":    len(rollback.Steps),
			"successful_steps":  successfulSteps,
			"rollback_complete": lastError == nil,
		},
		Timestamp: time.Now(),
	}

	if lastError != nil {
		result.Error = lastError.Error()
	}

	return result, lastError
}

// executeParallel executes steps in parallel where possible
func (ee *ExecutionEngine) executeParallel(ctx context.Context, plan *RecoveryPlan, progress *RecoveryProgress) (*RecoveryResult, error) {
	// Build dependency graph
	dependencyGraph := ee.buildStepDependencyGraph(plan.Steps)

	// Execute steps in topological order with parallelization
	executed := make(map[string]bool)
	executing := make(map[string]bool)
	results := make(map[string]interface{})
	errors := make(map[string]error)

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Continue until all steps are executed
	for len(executed) < len(plan.Steps) {
		// Find steps ready for execution
		readySteps := ee.findReadySteps(plan.Steps, dependencyGraph, executed, executing)

		if len(readySteps) == 0 {
			// Check if we're stuck due to errors
			if len(errors) > 0 {
				break
			}
			// Wait a bit and retry
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Execute ready steps in parallel
		for _, step := range readySteps {
			// Acquire execution slot
			select {
			case ee.executionPool <- struct{}{}:
				// Got slot, proceed
			case <-ctx.Done():
				return nil, ctx.Err()
			}

			mu.Lock()
			executing[step.ID] = true
			mu.Unlock()

			wg.Add(1)
			go func(s *RecoveryStep) {
				defer wg.Done()
				defer func() { <-ee.executionPool }() // Release slot

				stepCtx, cancel := context.WithTimeout(ctx, s.Timeout)
				defer cancel()

				result, err := ee.executeStep(stepCtx, s)

				mu.Lock()
				defer mu.Unlock()

				delete(executing, s.ID)
				executed[s.ID] = true

				if err != nil {
					errors[s.ID] = err
					s.Status = StepStatusFailed
					s.Error = err
				} else {
					results[s.ID] = result
					s.Status = StepStatusCompleted
					s.Result = result
				}

				// Update progress
				if ee.config.EnableProgressTracking && progress != nil {
					ee.updateProgress(progress, plan.Steps)
				}
			}(step)
		}

		// Wait for current batch to complete
		wg.Wait()

		// Check for critical step failures
		for stepID, err := range errors {
			step := ee.findStepByID(plan.Steps, stepID)
			if step != nil && step.Critical {
				return ee.createFailureResult(plan, fmt.Errorf("critical step %s failed: %w", stepID, err)), err
			}
		}
	}

	// Create final result
	return ee.createSuccessResult(plan, results), nil
}

// executeSequential executes steps sequentially
func (ee *ExecutionEngine) executeSequential(ctx context.Context, plan *RecoveryPlan, progress *RecoveryProgress) (*RecoveryResult, error) {
	results := make(map[string]interface{})

	for _, step := range plan.Steps {
		// Check dependencies
		if !ee.areDependenciesSatisfied(step, plan.Steps) {
			return ee.createFailureResult(plan, fmt.Errorf("dependencies not satisfied for step %s", step.ID)),
				fmt.Errorf("dependencies not satisfied")
		}

		// Execute step
		stepCtx, cancel := context.WithTimeout(ctx, step.Timeout)
		result, err := ee.executeStep(stepCtx, step)
		cancel()

		if err != nil {
			step.Status = StepStatusFailed
			step.Error = err

			if step.Critical {
				return ee.createFailureResult(plan, fmt.Errorf("critical step %s failed: %w", step.ID, err)), err
			}

			log.Warn().
				Err(err).
				Str("step_id", step.ID).
				Msg("Non-critical step failed, continuing")
		} else {
			step.Status = StepStatusCompleted
			step.Result = result
			results[step.ID] = result
		}

		// Update progress
		if ee.config.EnableProgressTracking && progress != nil {
			ee.updateProgress(progress, plan.Steps)
		}
	}

	return ee.createSuccessResult(plan, results), nil
}

// executeStep executes a single recovery step
func (ee *ExecutionEngine) executeStep(ctx context.Context, step *RecoveryStep) (interface{}, error) {
	// Track execution
	execution := &StepExecution{
		StepID:    step.ID,
		Status:    StepStatusRunning,
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	ee.mu.Lock()
	ee.activeExecutions[step.ID] = execution
	ee.mu.Unlock()

	defer func() {
		execution.EndTime = time.Now()
		ee.mu.Lock()
		delete(ee.activeExecutions, step.ID)
		ee.mu.Unlock()
	}()

	step.Status = StepStatusRunning
	step.StartTime = time.Now()

	log.Info().
		Str("step_id", step.ID).
		Str("action", step.Action).
		Str("target", step.Target).
		Msg("Executing recovery step")

	// Execute step with retries
	var result interface{}
	var err error

	for attempt := 0; attempt <= step.Retries; attempt++ {
		if attempt > 0 {
			log.Info().
				Str("step_id", step.ID).
				Int("attempt", attempt).
				Msg("Retrying step execution")

			// Exponential backoff
			backoff := time.Duration(attempt) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		result, err = ee.performStepAction(ctx, step)
		if err == nil {
			break
		}

		log.Warn().
			Err(err).
			Str("step_id", step.ID).
			Int("attempt", attempt).
			Msg("Step execution attempt failed")
	}

	step.EndTime = time.Now()
	step.Duration = step.EndTime.Sub(step.StartTime)

	if err != nil {
		execution.Status = StepStatusFailed
		execution.Error = err
		return nil, fmt.Errorf("step %s failed after %d attempts: %w", step.ID, step.Retries+1, err)
	}

	execution.Status = StepStatusCompleted
	execution.Result = result

	log.Info().
		Str("step_id", step.ID).
		Dur("duration", step.Duration).
		Msg("Step execution completed")

	return result, nil
}

// performStepAction performs the actual step action
func (ee *ExecutionEngine) performStepAction(ctx context.Context, step *RecoveryStep) (interface{}, error) {
	// Simulate step execution based on action type
	switch step.Action {
	case "prepare_recovery":
		return ee.performPreparation(ctx, step)
	case "validate_system_state":
		return ee.performValidation(ctx, step)
	case "failover_node":
		return ee.performFailover(ctx, step)
	case "restart_service":
		return ee.performServiceRestart(ctx, step)
	case "scale_resources":
		return ee.performResourceScaling(ctx, step)
	case "recover_partition":
		return ee.performPartitionRecovery(ctx, step)
	case "verify_recovery":
		return ee.performVerification(ctx, step)
	case "cleanup_recovery":
		return ee.performCleanup(ctx, step)
	default:
		return ee.performGenericAction(ctx, step)
	}
}

// Step action implementations (simplified)

func (ee *ExecutionEngine) performPreparation(ctx context.Context, step *RecoveryStep) (interface{}, error) {
	// Simulate preparation
	time.Sleep(100 * time.Millisecond)
	return map[string]interface{}{
		"prepared":            true,
		"resources_allocated": true,
	}, nil
}

func (ee *ExecutionEngine) performValidation(ctx context.Context, step *RecoveryStep) (interface{}, error) {
	// Simulate validation
	time.Sleep(200 * time.Millisecond)
	return map[string]interface{}{
		"validation_passed": true,
		"system_state":      "ready",
	}, nil
}

func (ee *ExecutionEngine) performFailover(ctx context.Context, step *RecoveryStep) (interface{}, error) {
	// Simulate failover
	time.Sleep(2 * time.Second)
	return map[string]interface{}{
		"failover_completed": true,
		"new_primary":        "backup_node_1",
		"data_migrated":      true,
	}, nil
}

func (ee *ExecutionEngine) performServiceRestart(ctx context.Context, step *RecoveryStep) (interface{}, error) {
	// Simulate service restart
	time.Sleep(1 * time.Second)
	return map[string]interface{}{
		"service_restarted":   true,
		"health_check_passed": true,
	}, nil
}

func (ee *ExecutionEngine) performResourceScaling(ctx context.Context, step *RecoveryStep) (interface{}, error) {
	// Simulate resource scaling
	time.Sleep(1500 * time.Millisecond)
	return map[string]interface{}{
		"scaling_completed": true,
		"new_capacity":      "150%",
	}, nil
}

func (ee *ExecutionEngine) performPartitionRecovery(ctx context.Context, step *RecoveryStep) (interface{}, error) {
	// Simulate partition recovery
	time.Sleep(2500 * time.Millisecond)
	return map[string]interface{}{
		"partition_healed":  true,
		"nodes_reconnected": true,
		"data_synchronized": true,
	}, nil
}

func (ee *ExecutionEngine) performVerification(ctx context.Context, step *RecoveryStep) (interface{}, error) {
	// Simulate verification
	time.Sleep(500 * time.Millisecond)
	return map[string]interface{}{
		"verification_passed": true,
		"system_healthy":      true,
	}, nil
}

func (ee *ExecutionEngine) performCleanup(ctx context.Context, step *RecoveryStep) (interface{}, error) {
	// Simulate cleanup
	time.Sleep(300 * time.Millisecond)
	return map[string]interface{}{
		"cleanup_completed":    true,
		"temp_resources_freed": true,
	}, nil
}

func (ee *ExecutionEngine) performGenericAction(ctx context.Context, step *RecoveryStep) (interface{}, error) {
	// Simulate generic action
	time.Sleep(1 * time.Second)
	return map[string]interface{}{
		"action_completed": true,
		"action_type":      step.Action,
	}, nil
}

// Helper methods

func (ee *ExecutionEngine) buildStepDependencyGraph(steps []*RecoveryStep) map[string][]string {
	graph := make(map[string][]string)
	for _, step := range steps {
		graph[step.ID] = step.Dependencies
	}
	return graph
}

func (ee *ExecutionEngine) findReadySteps(steps []*RecoveryStep, graph map[string][]string, executed, executing map[string]bool) []*RecoveryStep {
	var ready []*RecoveryStep

	for _, step := range steps {
		if executed[step.ID] || executing[step.ID] {
			continue
		}

		// Check if all dependencies are satisfied
		allSatisfied := true
		for _, depID := range step.Dependencies {
			if !executed[depID] {
				allSatisfied = false
				break
			}
		}

		if allSatisfied {
			ready = append(ready, step)
		}
	}

	return ready
}

func (ee *ExecutionEngine) findStepByID(steps []*RecoveryStep, stepID string) *RecoveryStep {
	for _, step := range steps {
		if step.ID == stepID {
			return step
		}
	}
	return nil
}

func (ee *ExecutionEngine) areDependenciesSatisfied(step *RecoveryStep, allSteps []*RecoveryStep) bool {
	for _, depID := range step.Dependencies {
		depStep := ee.findStepByID(allSteps, depID)
		if depStep == nil || depStep.Status != StepStatusCompleted {
			return false
		}
	}
	return true
}

func (ee *ExecutionEngine) updateProgress(progress *RecoveryProgress, steps []*RecoveryStep) {
	completed := 0
	failed := 0

	for _, step := range steps {
		switch step.Status {
		case StepStatusCompleted:
			completed++
		case StepStatusFailed:
			failed++
		}
	}

	progress.CompletedSteps = completed
	progress.FailedSteps = failed
	progress.PercentComplete = float64(completed) / float64(len(steps)) * 100
	progress.LastUpdated = time.Now()
}

func (ee *ExecutionEngine) createSuccessResult(plan *RecoveryPlan, results map[string]interface{}) *RecoveryResult {
	return &RecoveryResult{
		FaultID:    plan.ID,
		Strategy:   "orchestrated_recovery",
		Successful: true,
		Duration:   time.Since(plan.CreatedAt),
		Metadata: map[string]interface{}{
			"plan_id":        plan.ID,
			"steps_executed": len(results),
			"results":        results,
		},
		Timestamp: time.Now(),
	}
}

func (ee *ExecutionEngine) createFailureResult(plan *RecoveryPlan, err error) *RecoveryResult {
	return &RecoveryResult{
		FaultID:    plan.ID,
		Strategy:   "orchestrated_recovery",
		Successful: false,
		Duration:   time.Since(plan.CreatedAt),
		Error:      err.Error(),
		Metadata: map[string]interface{}{
			"plan_id":        plan.ID,
			"failure_reason": err.Error(),
		},
		Timestamp: time.Now(),
	}
}

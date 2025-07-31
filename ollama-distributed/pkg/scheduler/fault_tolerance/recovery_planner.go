package fault_tolerance

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

// RecoveryPlanner creates recovery plans for faults
type RecoveryPlanner struct {
	config *RecoveryPlannerConfig
}

// RecoveryPlannerConfig configures the recovery planner
type RecoveryPlannerConfig struct {
	PlanningTimeout        time.Duration `json:"planning_timeout"`
	EnableOptimization     bool          `json:"enable_optimization"`
	EnableResourcePlanning bool          `json:"enable_resource_planning"`
}

// NewRecoveryPlanner creates a new recovery planner
func NewRecoveryPlanner(config *RecoveryPlannerConfig) *RecoveryPlanner {
	if config == nil {
		config = &RecoveryPlannerConfig{
			PlanningTimeout:        2 * time.Minute,
			EnableOptimization:     true,
			EnableResourcePlanning: true,
		}
	}

	return &RecoveryPlanner{
		config: config,
	}
}

// CreateSingleFaultPlan creates a recovery plan for a single fault
func (rp *RecoveryPlanner) CreateSingleFaultPlan(fault *FaultDetection) (*RecoveryPlan, error) {
	plan := &RecoveryPlan{
		ID:          fmt.Sprintf("plan_%d", time.Now().UnixNano()),
		Name:        fmt.Sprintf("Recovery for %s", fault.Type),
		Description: fmt.Sprintf("Recovery plan for fault %s on %s", fault.ID, fault.Target),
		FaultIDs:    []string{fault.ID},
		NodeIDs:     []string{fault.Target},
		Steps:       make([]*RecoveryStep, 0),
		Dependencies: make([]*RecoveryDependency, 0),
		Priority:    rp.calculatePlanPriority(fault),
		EstimatedTime: rp.estimateRecoveryTime(fault),
		CreatedAt:   time.Now(),
		CreatedBy:   "recovery_planner",
		Metadata:    make(map[string]interface{}),
	}

	// Create recovery steps based on fault type
	steps, err := rp.createStepsForFault(fault)
	if err != nil {
		return nil, fmt.Errorf("failed to create recovery steps: %w", err)
	}
	plan.Steps = steps

	// Add resource requirements if enabled
	if rp.config.EnableResourcePlanning {
		plan.Resources = rp.calculateResourceRequirements(fault)
	}

	// Add constraints
	plan.Constraints = rp.createConstraints(fault)

	// Create rollback plan
	plan.Rollback = rp.createRollbackPlan(plan)

	log.Info().
		Str("plan_id", plan.ID).
		Str("fault_id", fault.ID).
		Int("steps", len(plan.Steps)).
		Msg("Created single fault recovery plan")

	return plan, nil
}

// CreateMultiNodePlan creates a recovery plan for multiple faults across nodes
func (rp *RecoveryPlanner) CreateMultiNodePlan(faults []*FaultDetection) (*RecoveryPlan, error) {
	if len(faults) == 0 {
		return nil, fmt.Errorf("no faults provided")
	}

	// Collect fault and node IDs
	faultIDs := make([]string, len(faults))
	nodeIDs := make(map[string]bool)
	for i, fault := range faults {
		faultIDs[i] = fault.ID
		nodeIDs[fault.Target] = true
	}

	nodeList := make([]string, 0, len(nodeIDs))
	for node := range nodeIDs {
		nodeList = append(nodeList, node)
	}

	plan := &RecoveryPlan{
		ID:          fmt.Sprintf("multiplan_%d", time.Now().UnixNano()),
		Name:        "Multi-Node Recovery Plan",
		Description: fmt.Sprintf("Recovery plan for %d faults across %d nodes", len(faults), len(nodeList)),
		FaultIDs:    faultIDs,
		NodeIDs:     nodeList,
		Steps:       make([]*RecoveryStep, 0),
		Dependencies: make([]*RecoveryDependency, 0),
		Priority:    rp.calculateMultiFaultPriority(faults),
		EstimatedTime: rp.estimateMultiRecoveryTime(faults),
		CreatedAt:   time.Now(),
		CreatedBy:   "recovery_planner",
		Metadata: map[string]interface{}{
			"fault_count": len(faults),
			"node_count":  len(nodeList),
		},
	}

	// Create coordinated recovery steps
	steps, err := rp.createCoordinatedSteps(faults)
	if err != nil {
		return nil, fmt.Errorf("failed to create coordinated recovery steps: %w", err)
	}
	plan.Steps = steps

	// Add resource requirements
	if rp.config.EnableResourcePlanning {
		plan.Resources = rp.calculateMultiResourceRequirements(faults)
	}

	// Add constraints
	plan.Constraints = rp.createMultiConstraints(faults)

	// Create rollback plan
	plan.Rollback = rp.createRollbackPlan(plan)

	log.Info().
		Str("plan_id", plan.ID).
		Int("faults", len(faults)).
		Int("nodes", len(nodeList)).
		Int("steps", len(plan.Steps)).
		Msg("Created multi-node recovery plan")

	return plan, nil
}

// createStepsForFault creates recovery steps for a specific fault
func (rp *RecoveryPlanner) createStepsForFault(fault *FaultDetection) ([]*RecoveryStep, error) {
	var steps []*RecoveryStep

	// Preparation step
	steps = append(steps, &RecoveryStep{
		ID:           fmt.Sprintf("prep_%s", fault.ID),
		Name:         "Preparation",
		Type:         StepTypePreparation,
		Action:       "prepare_recovery",
		Target:       fault.Target,
		Parameters:   map[string]interface{}{"fault_id": fault.ID},
		Timeout:      30 * time.Second,
		Retries:      1,
		Critical:     false,
		Parallel:     false,
		Status:       StepStatusPending,
		Metadata:     make(map[string]interface{}),
	})

	// Validation step
	steps = append(steps, &RecoveryStep{
		ID:           fmt.Sprintf("validate_%s", fault.ID),
		Name:         "Validation",
		Type:         StepTypeValidation,
		Action:       "validate_system_state",
		Target:       fault.Target,
		Parameters:   map[string]interface{}{"fault_type": fault.Type},
		Dependencies: []string{steps[0].ID},
		Timeout:      60 * time.Second,
		Retries:      2,
		Critical:     true,
		Parallel:     false,
		Status:       StepStatusPending,
		Metadata:     make(map[string]interface{}),
	})

	// Recovery execution steps based on fault type
	executionSteps := rp.createExecutionSteps(fault)
	steps = append(steps, executionSteps...)

	// Verification step
	steps = append(steps, &RecoveryStep{
		ID:         fmt.Sprintf("verify_%s", fault.ID),
		Name:       "Verification",
		Type:       StepTypeVerification,
		Action:     "verify_recovery",
		Target:     fault.Target,
		Parameters: map[string]interface{}{"expected_state": "healthy"},
		Dependencies: rp.getExecutionStepIDs(executionSteps),
		Timeout:    120 * time.Second,
		Retries:    3,
		Critical:   true,
		Parallel:   false,
		Status:     StepStatusPending,
		Metadata:   make(map[string]interface{}),
	})

	// Cleanup step
	steps = append(steps, &RecoveryStep{
		ID:           fmt.Sprintf("cleanup_%s", fault.ID),
		Name:         "Cleanup",
		Type:         StepTypeCleanup,
		Action:       "cleanup_recovery",
		Target:       fault.Target,
		Parameters:   map[string]interface{}{"cleanup_temp_resources": true},
		Dependencies: []string{steps[len(steps)-1].ID},
		Timeout:      60 * time.Second,
		Retries:      1,
		Critical:     false,
		Parallel:     false,
		Status:       StepStatusPending,
		Metadata:     make(map[string]interface{}),
	})

	return steps, nil
}

// createExecutionSteps creates execution steps based on fault type
func (rp *RecoveryPlanner) createExecutionSteps(fault *FaultDetection) []*RecoveryStep {
	var steps []*RecoveryStep

	switch fault.Type {
	case FaultTypeNodeFailure:
		steps = append(steps, &RecoveryStep{
			ID:       fmt.Sprintf("failover_%s", fault.ID),
			Name:     "Node Failover",
			Type:     StepTypeExecution,
			Action:   "failover_node",
			Target:   fault.Target,
			Parameters: map[string]interface{}{
				"backup_node": "auto_select",
				"migrate_data": true,
			},
			Timeout:  5 * time.Minute,
			Retries:  2,
			Critical: true,
			Parallel: false,
			Status:   StepStatusPending,
			Metadata: make(map[string]interface{}),
		})

	case FaultTypeServiceUnavailable:
		steps = append(steps, &RecoveryStep{
			ID:       fmt.Sprintf("restart_%s", fault.ID),
			Name:     "Service Restart",
			Type:     StepTypeExecution,
			Action:   "restart_service",
			Target:   fault.Target,
			Parameters: map[string]interface{}{
				"graceful": true,
				"timeout":  30,
			},
			Timeout:  2 * time.Minute,
			Retries:  3,
			Critical: true,
			Parallel: false,
			Status:   StepStatusPending,
			Metadata: make(map[string]interface{}),
		})

	case FaultTypeResourceExhaustion:
		steps = append(steps, &RecoveryStep{
			ID:       fmt.Sprintf("scale_%s", fault.ID),
			Name:     "Resource Scaling",
			Type:     StepTypeExecution,
			Action:   "scale_resources",
			Target:   fault.Target,
			Parameters: map[string]interface{}{
				"scale_factor": 1.5,
				"resource_type": "auto_detect",
			},
			Timeout:  3 * time.Minute,
			Retries:  2,
			Critical: true,
			Parallel: false,
			Status:   StepStatusPending,
			Metadata: make(map[string]interface{}),
		})

	case FaultTypeNetworkPartition:
		steps = append(steps, &RecoveryStep{
			ID:       fmt.Sprintf("partition_%s", fault.ID),
			Name:     "Partition Recovery",
			Type:     StepTypeExecution,
			Action:   "recover_partition",
			Target:   fault.Target,
			Parameters: map[string]interface{}{
				"reconnect_nodes": true,
				"sync_data": true,
			},
			Timeout:  4 * time.Minute,
			Retries:  2,
			Critical: true,
			Parallel: false,
			Status:   StepStatusPending,
			Metadata: make(map[string]interface{}),
		})

	default:
		// Generic recovery step
		steps = append(steps, &RecoveryStep{
			ID:       fmt.Sprintf("generic_%s", fault.ID),
			Name:     "Generic Recovery",
			Type:     StepTypeExecution,
			Action:   "generic_recovery",
			Target:   fault.Target,
			Parameters: map[string]interface{}{
				"fault_type": fault.Type,
			},
			Timeout:  2 * time.Minute,
			Retries:  2,
			Critical: true,
			Parallel: false,
			Status:   StepStatusPending,
			Metadata: make(map[string]interface{}),
		})
	}

	return steps
}

// createCoordinatedSteps creates coordinated steps for multiple faults
func (rp *RecoveryPlanner) createCoordinatedSteps(faults []*FaultDetection) ([]*RecoveryStep, error) {
	var steps []*RecoveryStep

	// Global preparation step
	steps = append(steps, &RecoveryStep{
		ID:       fmt.Sprintf("global_prep_%d", time.Now().UnixNano()),
		Name:     "Global Preparation",
		Type:     StepTypePreparation,
		Action:   "prepare_multi_recovery",
		Target:   "cluster",
		Parameters: map[string]interface{}{
			"fault_count": len(faults),
			"coordination_mode": "parallel",
		},
		Timeout:  60 * time.Second,
		Retries:  1,
		Critical: false,
		Parallel: false,
		Status:   StepStatusPending,
		Metadata: make(map[string]interface{}),
	})

	// Create individual recovery steps for each fault
	for _, fault := range faults {
		faultSteps, err := rp.createStepsForFault(fault)
		if err != nil {
			return nil, fmt.Errorf("failed to create steps for fault %s: %w", fault.ID, err)
		}

		// Make fault-specific steps depend on global preparation
		if len(faultSteps) > 0 {
			faultSteps[0].Dependencies = []string{steps[0].ID}
		}

		steps = append(steps, faultSteps...)
	}

	// Global verification step
	executionStepIDs := make([]string, 0)
	for _, step := range steps {
		if step.Type == StepTypeVerification {
			executionStepIDs = append(executionStepIDs, step.ID)
		}
	}

	steps = append(steps, &RecoveryStep{
		ID:           fmt.Sprintf("global_verify_%d", time.Now().UnixNano()),
		Name:         "Global Verification",
		Type:         StepTypeVerification,
		Action:       "verify_cluster_health",
		Target:       "cluster",
		Parameters:   map[string]interface{}{"check_all_nodes": true},
		Dependencies: executionStepIDs,
		Timeout:      3 * time.Minute,
		Retries:      3,
		Critical:     true,
		Parallel:     false,
		Status:       StepStatusPending,
		Metadata:     make(map[string]interface{}),
	})

	return steps, nil
}

// Helper methods

// calculatePlanPriority calculates priority for a recovery plan
func (rp *RecoveryPlanner) calculatePlanPriority(fault *FaultDetection) int {
	priority := 5

	switch fault.Severity {
	case FaultSeverityCritical:
		priority = 10
	case FaultSeverityHigh:
		priority = 8
	case FaultSeverityMedium:
		priority = 6
	case FaultSeverityLow:
		priority = 4
	}

	return priority
}

// calculateMultiFaultPriority calculates priority for multiple faults
func (rp *RecoveryPlanner) calculateMultiFaultPriority(faults []*FaultDetection) int {
	maxPriority := 0
	for _, fault := range faults {
		priority := rp.calculatePlanPriority(fault)
		if priority > maxPriority {
			maxPriority = priority
		}
	}
	return maxPriority
}

// estimateRecoveryTime estimates recovery time for a fault
func (rp *RecoveryPlanner) estimateRecoveryTime(fault *FaultDetection) time.Duration {
	baseTime := 2 * time.Minute

	switch fault.Type {
	case FaultTypeNodeFailure:
		return baseTime * 3
	case FaultTypeNetworkPartition:
		return baseTime * 2
	case FaultTypeServiceUnavailable:
		return baseTime
	case FaultTypeResourceExhaustion:
		return baseTime * 2
	default:
		return baseTime
	}
}

// estimateMultiRecoveryTime estimates recovery time for multiple faults
func (rp *RecoveryPlanner) estimateMultiRecoveryTime(faults []*FaultDetection) time.Duration {
	totalTime := time.Duration(0)
	for _, fault := range faults {
		totalTime += rp.estimateRecoveryTime(fault)
	}
	// Add coordination overhead
	return totalTime + time.Duration(len(faults))*30*time.Second
}

// getExecutionStepIDs gets IDs of execution steps
func (rp *RecoveryPlanner) getExecutionStepIDs(steps []*RecoveryStep) []string {
	ids := make([]string, len(steps))
	for i, step := range steps {
		ids[i] = step.ID
	}
	return ids
}

// calculateResourceRequirements calculates resource requirements
func (rp *RecoveryPlanner) calculateResourceRequirements(fault *FaultDetection) *ResourceRequirements {
	return &ResourceRequirements{
		CPU:    1.0,
		Memory: 1024 * 1024 * 1024, // 1GB
		Disk:   1024 * 1024 * 1024, // 1GB
		Network: 100 * 1024 * 1024, // 100MB
		Nodes:   []string{fault.Target},
		Services: []string{"recovery_service"},
		Constraints: make(map[string]interface{}),
	}
}

// calculateMultiResourceRequirements calculates resource requirements for multiple faults
func (rp *RecoveryPlanner) calculateMultiResourceRequirements(faults []*FaultDetection) *ResourceRequirements {
	nodeSet := make(map[string]bool)
	for _, fault := range faults {
		nodeSet[fault.Target] = true
	}

	nodes := make([]string, 0, len(nodeSet))
	for node := range nodeSet {
		nodes = append(nodes, node)
	}

	return &ResourceRequirements{
		CPU:    float64(len(faults)) * 1.0,
		Memory: int64(len(faults)) * 1024 * 1024 * 1024,
		Disk:   int64(len(faults)) * 1024 * 1024 * 1024,
		Network: int64(len(faults)) * 100 * 1024 * 1024,
		Nodes:   nodes,
		Services: []string{"recovery_service", "coordination_service"},
		Constraints: make(map[string]interface{}),
	}
}

// createConstraints creates recovery constraints
func (rp *RecoveryPlanner) createConstraints(fault *FaultDetection) *RecoveryConstraints {
	return &RecoveryConstraints{
		MaxDuration:     10 * time.Minute,
		MaxRetries:      3,
		RequiredNodes:   []string{fault.Target},
		ExcludedNodes:   []string{},
		MaintenanceMode: false,
		Constraints:     make(map[string]interface{}),
	}
}

// createMultiConstraints creates constraints for multiple faults
func (rp *RecoveryPlanner) createMultiConstraints(faults []*FaultDetection) *RecoveryConstraints {
	nodeSet := make(map[string]bool)
	for _, fault := range faults {
		nodeSet[fault.Target] = true
	}

	nodes := make([]string, 0, len(nodeSet))
	for node := range nodeSet {
		nodes = append(nodes, node)
	}

	return &RecoveryConstraints{
		MaxDuration:     time.Duration(len(faults)) * 10 * time.Minute,
		MaxRetries:      5,
		RequiredNodes:   nodes,
		ExcludedNodes:   []string{},
		MaintenanceMode: false,
		Constraints:     make(map[string]interface{}),
	}
}

// createRollbackPlan creates a rollback plan
func (rp *RecoveryPlanner) createRollbackPlan(plan *RecoveryPlan) *RollbackPlan {
	rollbackSteps := make([]*RecoveryStep, 0)

	// Create rollback steps in reverse order
	for i := len(plan.Steps) - 1; i >= 0; i-- {
		step := plan.Steps[i]
		if step.Type == StepTypeExecution {
			rollbackStep := &RecoveryStep{
				ID:       fmt.Sprintf("rollback_%s", step.ID),
				Name:     fmt.Sprintf("Rollback %s", step.Name),
				Type:     StepTypeRollback,
				Action:   fmt.Sprintf("rollback_%s", step.Action),
				Target:   step.Target,
				Parameters: map[string]interface{}{
					"original_step_id": step.ID,
					"rollback_mode":    "safe",
				},
				Timeout:  step.Timeout,
				Retries:  2,
				Critical: false,
				Parallel: false,
				Status:   StepStatusPending,
				Metadata: make(map[string]interface{}),
			}
			rollbackSteps = append(rollbackSteps, rollbackStep)
		}
	}

	return &RollbackPlan{
		ID:         fmt.Sprintf("rollback_%s", plan.ID),
		Steps:      rollbackSteps,
		Conditions: []string{"execution_failed", "manual_trigger"},
		Automatic:  true,
		Timeout:    plan.EstimatedTime,
		Metadata:   make(map[string]interface{}),
	}
}

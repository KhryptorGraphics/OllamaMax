package fault_tolerance

import "context"

// InferenceFT defines the fault tolerance interface for distributed inference
type InferenceFT interface {
	// Checkpoint operations
	CreateInferenceCheckpoint(ctx context.Context, sessionID string, state interface{}) error
	RestoreFromInferenceCheckpoint(ctx context.Context, sessionID, checkpointID string) error

	// Failure handling
	HandleInferenceFailure(ctx context.Context, sessionID, nodeID, errorMsg string) error

	// Dynamic operations
	TriggerDynamicRepartitioning(ctx context.Context, failedNodeID, sessionID string) error
	ApplyGracefulDegradation(ctx context.Context, sessionID string, constraints map[string]interface{}) error
}
package types

import (
	"fmt"
	"time"
)

// Error types for the distributed system

// DistributedError represents an error in the distributed system
type DistributedError struct {
	Code      ErrorCode `json:"code"`
	Message   string    `json:"message"`
	NodeID    NodeID    `json:"node_id,omitempty"`
	TaskID    TaskID    `json:"task_id,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Cause     error     `json:"-"`
}

func (e *DistributedError) Error() string {
	if e.NodeID != "" {
		return fmt.Sprintf("[%s] %s: %s (node: %s)", e.Code, e.Message, e.Timestamp.Format(time.RFC3339), e.NodeID)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Timestamp.Format(time.RFC3339))
}

func (e *DistributedError) Unwrap() error {
	return e.Cause
}

// ErrorCode represents different types of errors
type ErrorCode string

const (
	// Node errors
	ErrorCodeNodeNotFound    ErrorCode = "NODE_NOT_FOUND"
	ErrorCodeNodeUnavailable ErrorCode = "NODE_UNAVAILABLE"
	ErrorCodeNodeOverloaded  ErrorCode = "NODE_OVERLOADED"

	// Task errors
	ErrorCodeTaskNotFound  ErrorCode = "TASK_NOT_FOUND"
	ErrorCodeTaskFailed    ErrorCode = "TASK_FAILED"
	ErrorCodeTaskTimeout   ErrorCode = "TASK_TIMEOUT"
	ErrorCodeTaskCancelled ErrorCode = "TASK_CANCELLED"

	// Model errors
	ErrorCodeModelNotFound      ErrorCode = "MODEL_NOT_FOUND"
	ErrorCodeModelLoadFailed    ErrorCode = "MODEL_LOAD_FAILED"
	ErrorCodeModelCorrupted     ErrorCode = "MODEL_CORRUPTED"
	ErrorCodeModelNotReplicated ErrorCode = "MODEL_NOT_REPLICATED"

	// Network errors
	ErrorCodeNetworkPartition ErrorCode = "NETWORK_PARTITION"
	ErrorCodeConnectionFailed ErrorCode = "CONNECTION_FAILED"
	ErrorCodeTimeout          ErrorCode = "TIMEOUT"

	// Consensus errors
	ErrorCodeConsensusFailure  ErrorCode = "CONSENSUS_FAILURE"
	ErrorCodeLeaderElection    ErrorCode = "LEADER_ELECTION_FAILED"
	ErrorCodeStateInconsistent ErrorCode = "STATE_INCONSISTENT"

	// Resource errors
	ErrorCodeInsufficientResources ErrorCode = "INSUFFICIENT_RESOURCES"
	ErrorCodeResourceExhausted     ErrorCode = "RESOURCE_EXHAUSTED"

	// Configuration errors
	ErrorCodeInvalidConfig  ErrorCode = "INVALID_CONFIG"
	ErrorCodeConfigNotFound ErrorCode = "CONFIG_NOT_FOUND"

	// Authentication errors
	ErrorCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrorCodeForbidden    ErrorCode = "FORBIDDEN"
	ErrorCodeTokenExpired ErrorCode = "TOKEN_EXPIRED"

	// Generic errors
	ErrorCodeInternalError  ErrorCode = "INTERNAL_ERROR"
	ErrorCodeNotImplemented ErrorCode = "NOT_IMPLEMENTED"
	ErrorCodeInvalidRequest ErrorCode = "INVALID_REQUEST"
)

// Error constructors

func NewNodeNotFoundError(nodeID NodeID) *DistributedError {
	return &DistributedError{
		Code:      ErrorCodeNodeNotFound,
		Message:   fmt.Sprintf("Node not found: %s", nodeID),
		NodeID:    nodeID,
		Timestamp: time.Now(),
	}
}

func NewTaskNotFoundError(taskID TaskID) *DistributedError {
	return &DistributedError{
		Code:      ErrorCodeTaskNotFound,
		Message:   fmt.Sprintf("Task not found: %s", taskID),
		TaskID:    taskID,
		Timestamp: time.Now(),
	}
}

func NewModelNotFoundError(modelName string) *DistributedError {
	return &DistributedError{
		Code:      ErrorCodeModelNotFound,
		Message:   fmt.Sprintf("Model not found: %s", modelName),
		Timestamp: time.Now(),
	}
}

func NewInsufficientResourcesError(nodeID NodeID, required *ResourceRequirements) *DistributedError {
	return &DistributedError{
		Code:      ErrorCodeInsufficientResources,
		Message:   fmt.Sprintf("Insufficient resources on node %s", nodeID),
		NodeID:    nodeID,
		Timestamp: time.Now(),
	}
}

func NewNetworkPartitionError(nodeID NodeID) *DistributedError {
	return &DistributedError{
		Code:      ErrorCodeNetworkPartition,
		Message:   fmt.Sprintf("Network partition detected for node %s", nodeID),
		NodeID:    nodeID,
		Timestamp: time.Now(),
	}
}

func NewConsensusFailureError(message string) *DistributedError {
	return &DistributedError{
		Code:      ErrorCodeConsensusFailure,
		Message:   message,
		Timestamp: time.Now(),
	}
}

func NewUnauthorizedError(message string) *DistributedError {
	return &DistributedError{
		Code:      ErrorCodeUnauthorized,
		Message:   message,
		Timestamp: time.Now(),
	}
}

func NewInternalError(message string, cause error) *DistributedError {
	return &DistributedError{
		Code:      ErrorCodeInternalError,
		Message:   message,
		Timestamp: time.Now(),
		Cause:     cause,
	}
}

func NewNotImplementedError(feature string) *DistributedError {
	return &DistributedError{
		Code:      ErrorCodeNotImplemented,
		Message:   fmt.Sprintf("Feature not implemented: %s", feature),
		Timestamp: time.Now(),
	}
}

// Error checking utilities

func IsNodeError(err error) bool {
	if de, ok := err.(*DistributedError); ok {
		return de.Code == ErrorCodeNodeNotFound ||
			de.Code == ErrorCodeNodeUnavailable ||
			de.Code == ErrorCodeNodeOverloaded
	}
	return false
}

func IsTaskError(err error) bool {
	if de, ok := err.(*DistributedError); ok {
		return de.Code == ErrorCodeTaskNotFound ||
			de.Code == ErrorCodeTaskFailed ||
			de.Code == ErrorCodeTaskTimeout ||
			de.Code == ErrorCodeTaskCancelled
	}
	return false
}

func IsModelError(err error) bool {
	if de, ok := err.(*DistributedError); ok {
		return de.Code == ErrorCodeModelNotFound ||
			de.Code == ErrorCodeModelLoadFailed ||
			de.Code == ErrorCodeModelCorrupted ||
			de.Code == ErrorCodeModelNotReplicated
	}
	return false
}

func IsNetworkError(err error) bool {
	if de, ok := err.(*DistributedError); ok {
		return de.Code == ErrorCodeNetworkPartition ||
			de.Code == ErrorCodeConnectionFailed ||
			de.Code == ErrorCodeTimeout
	}
	return false
}

func IsConsensusError(err error) bool {
	if de, ok := err.(*DistributedError); ok {
		return de.Code == ErrorCodeConsensusFailure ||
			de.Code == ErrorCodeLeaderElection ||
			de.Code == ErrorCodeStateInconsistent
	}
	return false
}

func IsResourceError(err error) bool {
	if de, ok := err.(*DistributedError); ok {
		return de.Code == ErrorCodeInsufficientResources ||
			de.Code == ErrorCodeResourceExhausted
	}
	return false
}

func IsAuthError(err error) bool {
	if de, ok := err.(*DistributedError); ok {
		return de.Code == ErrorCodeUnauthorized ||
			de.Code == ErrorCodeForbidden ||
			de.Code == ErrorCodeTokenExpired
	}
	return false
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) ErrorCode {
	if de, ok := err.(*DistributedError); ok {
		return de.Code
	}
	return ErrorCodeInternalError
}

// GetNodeID extracts the node ID from an error if present
func GetNodeID(err error) NodeID {
	if de, ok := err.(*DistributedError); ok {
		return de.NodeID
	}
	return ""
}

// GetTaskID extracts the task ID from an error if present
func GetTaskID(err error) TaskID {
	if de, ok := err.(*DistributedError); ok {
		return de.TaskID
	}
	return ""
}

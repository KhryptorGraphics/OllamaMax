package fault_tolerance

import (
	"context"
	"time"
)

// SelfHealingRecoveryAdapter adapts SelfHealingEngine to RecoveryStrategy for the RecoveryEngine
// This allows the RecoveryEngine to invoke self-healing flows as a recovery method
// without directly depending on enhanced components.
type SelfHealingRecoveryAdapter struct {
	name   string
	healer *SelfHealingEngine
}

func NewSelfHealingRecoveryAdapter(healer *SelfHealingEngine) *SelfHealingRecoveryAdapter {
	return &SelfHealingRecoveryAdapter{
		name:   "self_healing",
		healer: healer,
	}
}

func (s *SelfHealingRecoveryAdapter) GetName() string { return s.name }

func (s *SelfHealingRecoveryAdapter) CanHandle(fault *FaultDetection) bool {
	// Be conservative; use self-healing for performance/resource/service issues and as fallback
	if s.healer == nil || fault == nil {
		return false
	}
	switch fault.Type {
	case FaultTypePerformanceAnomaly, FaultTypeResourceExhaustion, FaultTypeServiceUnavailable, FaultTypeNetworkPartition:
		return true
	default:
		// Allow as fallback for node failures at low priority
		return fault.Type == FaultTypeNodeFailure
	}
}

func (s *SelfHealingRecoveryAdapter) Recover(ctx context.Context, fault *FaultDetection) (*RecoveryResult, error) {
	if s.healer == nil {
		return &RecoveryResult{
			FaultID:    fault.ID,
			Strategy:   s.name,
			Successful: false,
			Duration:   0,
			Error:      "self-healing engine not available",
			Metadata:   map[string]interface{}{"reason": "nil_healer"},
			Timestamp:  time.Now(),
		}, nil
	}

	start := time.Now()
	res, err := s.healer.HealFault(ctx, fault)
	duration := time.Since(start)

	rr := &RecoveryResult{
		FaultID:    fault.ID,
		Strategy:   s.name,
		Successful: err == nil && res != nil && res.Success,
		Duration:   duration,
		Metadata:   map[string]interface{}{},
		Timestamp:  time.Now(),
	}
	if err != nil {
		rr.Error = err.Error()
	}
	if res != nil {
		// propagate some details
		rr.Metadata["actions"] = res.Actions
		rr.Metadata["confidence"] = res.Confidence
		rr.Metadata["healing_duration"] = res.Duration
	}
	return rr, nil
}


package fault_tolerance

import (
	"time"
)

// UpdateFaultStatus updates the status of a fault detection
func (fd *FaultDetector) UpdateFaultStatus(faultID string, status FaultStatus) {
	fd.detectionsMu.Lock()
	defer fd.detectionsMu.Unlock()
	
	if fault, exists := fd.detections[faultID]; exists {
		fault.Status = status
		if status == FaultStatusResolved {
			now := time.Now()
			fault.ResolvedAt = &now
		}
	}
}
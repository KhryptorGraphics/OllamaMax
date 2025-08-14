package fault_tolerance

import (
	"context"
	"testing"
	"time"
)

// mockNode simulates a minimal node structure compatible with predictive_detection's NodeInfo
type mockNode struct {
	ID      string
	Metrics map[string]interface{}
}

// TestEnhancedFTM_NodeProvider_Wiring ensures SetNodeProvider is used by GetAvailableNodes
func TestEnhancedFTM_NodeProvider_Wiring(t *testing.T) {
	base := NewFaultToleranceManager(&Config{HealthCheckInterval: time.Second})
	cfg := NewEnhancedFaultToleranceConfig(&Config{HealthCheckInterval: time.Second})
	eftm := NewEnhancedFaultToleranceManager(cfg, base)

	// Provide two mock nodes via provider
	provided := []interface{}{
		&NodeInfo{ID: "node-A", Metrics: map[string]interface{}{"cpu_utilization": 0.5}},
		map[string]interface{}{"id": "node-B", "cpu_utilization": 0.7},
	}

	eftm.SetNodeProvider(func() []interface{} { return provided })

	got := eftm.GetAvailableNodes()
	if len(got) != len(provided) {
		t.Fatalf("expected %d nodes from provider, got %d", len(provided), len(got))
	}
}

// TestFaultPredictor_UsesNodeProvider validates predictive path consumes nodes from provider and runs without panic
func TestFaultPredictor_UsesNodeProvider(t *testing.T) {
	base := NewFaultToleranceManager(&Config{HealthCheckInterval: time.Millisecond * 10})
	cfg := NewEnhancedFaultToleranceConfig(&Config{HealthCheckInterval: time.Millisecond * 10})

	// Create EFTM and enable prediction
	eftm := NewEnhancedFaultToleranceManager(cfg, base)

	// Inject deterministic nodes with metrics that should trigger some predictions
	provided := []interface{}{
		&NodeInfo{ID: "n1", Metrics: map[string]interface{}{
			"cpu_utilization":     0.95,
			"memory_utilization":  0.90,
			"disk_utilization":    0.80,
			"network_utilization": 0.70,
			"temperature":         85.0,
			"error_rate":          0.10,
			"latency":             250.0,
			"throughput":          10.0,
			"active_requests":     150.0,
			"queued_requests":     100.0,
			"gpu_utilization":     0.80,
			"active_processes":    200.0,
		}},
		map[string]interface{}{
			"id":                  "n2",
			"cpu_utilization":     0.80,
			"memory_utilization":  0.85,
			"disk_utilization":    0.70,
			"network_utilization": 0.60,
			"temperature":         75.0,
			"error_rate":          0.05,
			"latency":             200.0,
			"throughput":          12.0,
			"active_requests":     120.0,
			"queued_requests":     80.0,
			"gpu_utilization":     0.50,
			"active_processes":    150.0,
		},
	}
	eftm.SetNodeProvider(func() []interface{} { return provided })

	// Access the predictor via public API on EFTM: Start -> it will spawn predictor loop if learning enabled
	if err := eftm.Start(); err != nil {
		t.Fatalf("failed to start EFTM: %v", err)
	}
	defer eftm.Shutdown(context.Background())

	// Manually invoke a single prediction tick by calling predictFaults through the predictor
	// We can't call unexported methods directly; instead, simulate by creating a new predictor that uses eftm
	pred := NewFaultPredictor(cfg, eftm.FaultToleranceManager)
	// Make sure the predictor points to our eftm so GetAvailableNodes uses our provider
	pred.manager = eftm

	// Lower threshold to ensure we register at least one prediction
	pred.threshold = 0.1

	// Run predict once; ensure it does not panic and updates metrics/history
	before := pred.metrics.PredictionsMade
	pred.predictFaults()
	after := pred.metrics.PredictionsMade

	if after <= before {
		t.Fatalf("expected predictions to increase, before=%d after=%d", before, after)
	}
}

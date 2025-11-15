package scheduler

import (
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/test"
)

func TestPeerPerformanceTracker(t *testing.T) {
	// Create a new peer performance tracker
	config := DefaultPeerSelectionConfig()
	tracker := NewPeerPerformanceTracker(config)
	defer tracker.Stop()

	// Create test peers
	peer1 := test.RandPeerIDFatal(t)
	peer2 := test.RandPeerIDFatal(t)

	// Test updating peer metrics
	metrics1 := &PeerMetrics{
		EffectiveFLOPs:      10.0,
		AvailableMemory:     8 * 1024 * 1024 * 1024, // 8GB
		Latency:             50 * time.Millisecond,
		LatencyPenalty:      1.2,
		EstimatedCongestion: 0.8,
		LastUtilityScore:    0.0, // Will be calculated
	}

	metrics2 := &PeerMetrics{
		EffectiveFLOPs:      15.0,
		AvailableMemory:     16 * 1024 * 1024 * 1024, // 16GB
		Latency:             30 * time.Millisecond,
		LatencyPenalty:      1.1,
		EstimatedCongestion: 0.5,
		LastUtilityScore:    0.0,
	}

	tracker.UpdatePeerMetrics(peer1, metrics1)
	tracker.UpdatePeerMetrics(peer2, metrics2)

	// Verify metrics were updated
	peer1Metrics := tracker.GetPeerMetrics(peer1)
	if peer1Metrics == nil {
		t.Fatal("Failed to get metrics for peer1")
	}

	if peer1Metrics.EffectiveFLOPs != 10.0 {
		t.Errorf("Expected EffectiveFLOPs 10.0, got %f", peer1Metrics.EffectiveFLOPs)
	}

	// Test utility score calculation
	// U = (effective_FLOPs × available_memory) / (latency_penalty × estimated_congestion)
	// expectedScore1 := (10.0 * float64(8*1024*1024*1024)) / (1.2 * 0.8)
	// expectedScore2 := (15.0 * float64(16*1024*1024*1024)) / (1.1 * 0.5)

	if peer1Metrics.LastUtilityScore <= 0 {
		t.Errorf("Expected positive utility score for peer1, got %f", peer1Metrics.LastUtilityScore)
	}

	// Test peer scores retrieval
	allScores := tracker.GetAllPeerScores()
	if len(allScores) != 2 {
		t.Errorf("Expected 2 peer scores, got %d", len(allScores))
	}

	// Verify sorting (peer2 should have higher score)
	if allScores[0].PeerID != peer2 {
		t.Error("Expected peer2 to have highest utility score")
	}
}

func TestUtilityScoreCalculation(t *testing.T) {
	tracker := &PeerPerformanceTracker{
		config: DefaultPeerSelectionConfig(),
	}

	// Test utility score calculation
	metrics := &PeerMetrics{
		EffectiveFLOPs:      10.0,
		AvailableMemory:     8 * 1024 * 1024 * 1024, // 8GB
		LatencyPenalty:      1.2,
		EstimatedCongestion: 0.8,
	}

	tracker.calculateUtilityScore(metrics)

	expectedScore := (10.0 * float64(8*1024*1024*1024)) / (1.2 * 0.8)
	if metrics.LastUtilityScore != expectedScore {
		t.Errorf("Expected utility score %f, got %f", expectedScore, metrics.LastUtilityScore)
	}

	// Test with zero denominator (should default to 1.0)
	metrics2 := &PeerMetrics{
		EffectiveFLOPs:      10.0,
		AvailableMemory:     8 * 1024 * 1024 * 1024,
		LatencyPenalty:      0.0,
		EstimatedCongestion: 0.0,
	}

	tracker.calculateUtilityScore(metrics2)

	expectedScore2 := (10.0 * float64(8*1024*1024*1024)) / 1.0 // Denominator defaults to 1.0
	if metrics2.LastUtilityScore != expectedScore2 {
		t.Errorf("Expected utility score %f with zero denominator, got %f", expectedScore2, metrics2.LastUtilityScore)
	}
}

func TestDynamicKSelection(t *testing.T) {
	tracker := &PeerPerformanceTracker{
		config: DefaultPeerSelectionConfig(),
	}

	// Create test peer scores with diminishing returns
	peerScores := []*PeerUtilityScore{
		{UtilityScore: 100.0},
		{UtilityScore: 90.0},
		{UtilityScore: 85.0},
		{UtilityScore: 84.0},
		{UtilityScore: 83.5},
	}

	// Test optimal K calculation
	maxK := 5
	optimalK := tracker.calculateOptimalK(peerScores, maxK)

	// Should stop at K=3 due to diminishing returns (5% threshold)
	if optimalK > 3 {
		t.Errorf("Expected optimal K <= 3 due to diminishing returns, got %d", optimalK)
	}

	// Test with linear decrease (no diminishing returns)
	linearScores := []*PeerUtilityScore{
		{UtilityScore: 100.0},
		{UtilityScore: 80.0},
		{UtilityScore: 60.0},
		{UtilityScore: 40.0},
		{UtilityScore: 20.0},
	}

	optimalK2 := tracker.calculateOptimalK(linearScores, maxK)
	if optimalK2 < maxK {
		t.Errorf("Expected optimal K = %d for linear decrease, got %d", maxK, optimalK2)
	}
}

func TestLatencyCalculations(t *testing.T) {
	tracker := &PeerPerformanceTracker{
		config: DefaultPeerSelectionConfig(),
	}

	// Test average latency calculation
	latencyHistory := []time.Duration{
		50 * time.Millisecond,
		60 * time.Millisecond,
		40 * time.Millisecond,
		70 * time.Millisecond,
	}

	avgLatency := tracker.calculateAverageLatency(latencyHistory)
	expectedAvg := (50 + 60 + 40 + 70) * int64(time.Millisecond) / 4

	if int64(avgLatency) != expectedAvg {
		t.Errorf("Expected average latency %d, got %d", expectedAvg, avgLatency)
	}

	// Test latency penalty calculation
	latency := 100 * time.Millisecond
	penalty := tracker.calculateLatencyPenalty(latency)

	// With default penalty factor of 1.5:
	// penalty = (100/100)^1.5 = 1.0
	if penalty < 1.0 {
		t.Errorf("Expected penalty >= 1.0, got %f", penalty)
	}

	// Test with higher latency
	highLatency := 500 * time.Millisecond
	highPenalty := tracker.calculateLatencyPenalty(highLatency)

	if highPenalty <= penalty {
		t.Errorf("Expected higher penalty for higher latency, got %f <= %f", highPenalty, penalty)
	}
}

func TestTaskCompletionRecording(t *testing.T) {
	tracker := NewPeerPerformanceTracker(nil)
	defer tracker.Stop()

	peerID := test.RandPeerIDFatal(t)

	// Initialize peer metrics
	initialMetrics := &PeerMetrics{
		EffectiveFLOPs:      5.0,
		AvailableMemory:     4 * 1024 * 1024 * 1024,
		Latency:             100 * time.Millisecond,
		LatencyPenalty:      1.5,
		EstimatedCongestion: 1.0,
	}
	tracker.UpdatePeerMetrics(peerID, initialMetrics)

	// Record successful task completion
	tracker.RecordTaskCompletion(peerID, true, 100*time.Millisecond, 1000000)

	// Verify metrics were updated
	metrics := tracker.GetPeerMetrics(peerID)
	if metrics == nil {
		t.Fatal("Failed to get peer metrics")
	}

	if metrics.TotalTasks != 1 {
		t.Errorf("Expected TotalTasks = 1, got %d", metrics.TotalTasks)
	}

	if metrics.SuccessfulTasks != 1 {
		t.Errorf("Expected SuccessfulTasks = 1, got %d", metrics.SuccessfulTasks)
	}

	if metrics.SuccessRate != 1.0 {
		t.Errorf("Expected SuccessRate = 1.0, got %f", metrics.SuccessRate)
	}

	// Record failed task completion
	tracker.RecordTaskCompletion(peerID, false, 50*time.Millisecond, 500000)

	// Verify updated metrics
	metrics = tracker.GetPeerMetrics(peerID)
	if metrics.TotalTasks != 2 {
		t.Errorf("Expected TotalTasks = 2, got %d", metrics.TotalTasks)
	}

	if metrics.SuccessfulTasks != 1 {
		t.Errorf("Expected SuccessfulTasks = 1, got %d", metrics.SuccessfulTasks)
	}

	if metrics.FailedTasks != 1 {
		t.Errorf("Expected FailedTasks = 1, got %d", metrics.FailedTasks)
	}

	if metrics.SuccessRate != 0.5 {
		t.Errorf("Expected SuccessRate = 0.5, got %f", metrics.SuccessRate)
	}
}
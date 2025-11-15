package scheduler

import (
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBackPressureManager(t *testing.T) {
	config := DefaultBackPressureConfig()
	bpm := NewBackPressureManager(config)

	assert.NotNil(t, bpm)
	assert.NotNil(t, bpm.peerQueues)
	assert.NotNil(t, bpm.throttledPeers)
	assert.Equal(t, config.MaxQueueDepth, bpm.config.MaxQueueDepth)
}

func TestUpdateQueueDepth(t *testing.T) {
	bpm := NewBackPressureManager(DefaultBackPressureConfig())
	defer bpm.Stop()

	peerID := peer.ID("test-peer-1")
	
	// Update queue depth
	bpm.UpdateQueueDepth(peerID, 50)

	// Verify queue was created and updated
	depth, err := bpm.GetQueueDepth(peerID)
	require.NoError(t, err)
	assert.Equal(t, 50, depth)
}

func TestThrottlingOnHighQueueDepth(t *testing.T) {
	config := &BackPressureConfig{
		MaxQueueDepth:       100,
		QueueDepthThreshold: 0.8,
		ThrottleDuration:    5 * time.Second,
		MonitorInterval:     100 * time.Millisecond,
		AdaptiveThrottling:  false,
		MinProcessingRate:   1.0,
	}
	bpm := NewBackPressureManager(config)
	defer bpm.Stop()

	peerID := peer.ID("test-peer-1")
	
	// Update with high queue depth (above threshold)
	bpm.UpdateQueueDepth(peerID, 85)

	// Verify peer is throttled
	assert.True(t, bpm.IsThrottled(peerID))
}

func TestNoThrottlingOnLowQueueDepth(t *testing.T) {
	config := &BackPressureConfig{
		MaxQueueDepth:       100,
		QueueDepthThreshold: 0.8,
		ThrottleDuration:    5 * time.Second,
		MonitorInterval:     100 * time.Millisecond,
		AdaptiveThrottling:  false,
		MinProcessingRate:   1.0,
	}
	bpm := NewBackPressureManager(config)
	defer bpm.Stop()

	peerID := peer.ID("test-peer-1")
	
	// Update with low queue depth
	bpm.UpdateQueueDepth(peerID, 50)

	// Verify peer is not throttled
	assert.False(t, bpm.IsThrottled(peerID))
}

func TestThrottleExpiration(t *testing.T) {
	config := &BackPressureConfig{
		MaxQueueDepth:       100,
		QueueDepthThreshold: 0.8,
		ThrottleDuration:    100 * time.Millisecond,
		MonitorInterval:     50 * time.Millisecond,
		AdaptiveThrottling:  false,
		MinProcessingRate:   1.0,
	}
	bpm := NewBackPressureManager(config)
	defer bpm.Stop()

	peerID := peer.ID("test-peer-1")
	
	// Trigger throttling
	bpm.UpdateQueueDepth(peerID, 85)
	assert.True(t, bpm.IsThrottled(peerID))

	// Wait for throttle to expire
	time.Sleep(150 * time.Millisecond)

	// Verify throttle expired
	assert.False(t, bpm.IsThrottled(peerID))
}

func TestAdaptiveThrottling(t *testing.T) {
	config := &BackPressureConfig{
		MaxQueueDepth:       100,
		QueueDepthThreshold: 0.8,
		ThrottleDuration:    5 * time.Second,
		MonitorInterval:     100 * time.Millisecond,
		AdaptiveThrottling:  true,
		MinProcessingRate:   5.0,
	}
	bpm := NewBackPressureManager(config)
	defer bpm.Stop()

	peerID := peer.ID("test-peer-1")
	
	// Simulate slow processing rate
	for i := 0; i < 10; i++ {
		bpm.UpdateQueueDepth(peerID, 50+i) // Queue growing slowly
		time.Sleep(10 * time.Millisecond)
	}

	// Should be throttled due to low processing rate
	assert.True(t, bpm.IsThrottled(peerID))
}

func TestProcessingRateCalculation(t *testing.T) {
	bpm := NewBackPressureManager(DefaultBackPressureConfig())
	defer bpm.Stop()

	peerID := peer.ID("test-peer-1")
	
	// Simulate queue draining (good processing rate)
	depths := []int{100, 95, 90, 85, 80, 75, 70, 65, 60, 55}
	for _, depth := range depths {
		bpm.UpdateQueueDepth(peerID, depth)
		time.Sleep(10 * time.Millisecond)
	}

	// Get queue and check processing rate
	bpm.mu.RLock()
	queue := bpm.peerQueues[peerID]
	bpm.mu.RUnlock()

	require.NotNil(t, queue)
	// Processing rate should be positive (queue draining)
	assert.Greater(t, queue.ProcessingRate, 0.0)
}

func TestMultiplePeers(t *testing.T) {
	bpm := NewBackPressureManager(DefaultBackPressureConfig())
	defer bpm.Stop()

	peer1 := peer.ID("peer-1")
	peer2 := peer.ID("peer-2")
	peer3 := peer.ID("peer-3")

	// Update different peers
	bpm.UpdateQueueDepth(peer1, 50)
	bpm.UpdateQueueDepth(peer2, 85)
	bpm.UpdateQueueDepth(peer3, 30)

	// Verify individual states
	assert.False(t, bpm.IsThrottled(peer1))
	assert.True(t, bpm.IsThrottled(peer2))
	assert.False(t, bpm.IsThrottled(peer3))
}

func TestGetQueueDepthError(t *testing.T) {
	bpm := NewBackPressureManager(DefaultBackPressureConfig())
	defer bpm.Stop()

	peerID := peer.ID("nonexistent-peer")
	
	_, err := bpm.GetQueueDepth(peerID)
	assert.Error(t, err)
}

func TestUnthrottlePeer(t *testing.T) {
	config := &BackPressureConfig{
		MaxQueueDepth:       100,
		QueueDepthThreshold: 0.8,
		ThrottleDuration:    5 * time.Second,
		MonitorInterval:     100 * time.Millisecond,
		AdaptiveThrottling:  false,
		MinProcessingRate:   1.0,
	}
	bpm := NewBackPressureManager(config)
	defer bpm.Stop()

	peerID := peer.ID("test-peer-1")
	
	// Trigger throttling
	bpm.UpdateQueueDepth(peerID, 85)
	assert.True(t, bpm.IsThrottled(peerID))

	// Reduce queue depth below threshold
	bpm.UpdateQueueDepth(peerID, 50)

	// Should be unthrottled
	assert.False(t, bpm.IsThrottled(peerID))
}

func TestQueueDepthHistory(t *testing.T) {
	bpm := NewBackPressureManager(DefaultBackPressureConfig())
	defer bpm.Stop()

	peerID := peer.ID("test-peer-1")
	
	// Update multiple times
	for i := 0; i < 10; i++ {
		bpm.UpdateQueueDepth(peerID, i*10)
	}

	// Verify history is maintained
	bpm.mu.RLock()
	queue := bpm.peerQueues[peerID]
	bpm.mu.RUnlock()

	require.NotNil(t, queue)
	assert.Len(t, queue.DepthHistory, 10)
}


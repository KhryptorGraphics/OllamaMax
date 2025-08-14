package consensus

import (
	"testing"
	"time"

	"github.com/hashicorp/raft"
	"github.com/stretchr/testify/assert"
)

func TestFSM_BasicOperations(t *testing.T) {
	fsm := &FSM{
		state:   make(map[string]interface{}),
		applyCh: make(chan *ApplyEvent, 10),
	}

	// Test initial state
	assert.Empty(t, fsm.state)

	// Test setting a value
	fsm.state["test-key"] = "test-value"
	assert.Equal(t, "test-value", fsm.state["test-key"])

	// Test getting a value
	value, exists := fsm.state["test-key"]
	assert.True(t, exists)
	assert.Equal(t, "test-value", value)

	// Test deleting a value
	delete(fsm.state, "test-key")
	_, exists = fsm.state["test-key"]
	assert.False(t, exists)
}

func TestApplyEvent_Creation(t *testing.T) {
	event := &ApplyEvent{
		Type:      "set",
		Key:       "test-key",
		Value:     "test-value",
		Metadata:  map[string]interface{}{"author": "test"},
		Timestamp: time.Now(),
	}

	assert.Equal(t, "set", event.Type)
	assert.Equal(t, "test-key", event.Key)
	assert.Equal(t, "test-value", event.Value)
	assert.Equal(t, "test", event.Metadata["author"])
	assert.False(t, event.Timestamp.IsZero())
}

func TestNodeCapability_BasicFunctionality(t *testing.T) {
	capability := &NodeCapability{
		NodeID:            "test-node",
		CPUCores:          4,
		MemoryGB:          8,
		StorageGB:         100,
		NetworkBandwidth:  1000,
		Latency:           10 * time.Millisecond,
		Throughput:        1000.0,
		Reliability:       0.99,
		Uptime:            24 * time.Hour,
		LoadAverage:       0.5,
		ActiveConnections: 10,
		Region:            "us-east-1",
		Zone:              "us-east-1a",
		Priority:          0.8,
		LastUpdated:       time.Now(),
	}

	// Test initial state
	assert.Equal(t, "test-node", string(capability.NodeID))
	assert.Equal(t, 4, capability.CPUCores)
	assert.Equal(t, 8, capability.MemoryGB)
	assert.Equal(t, 100, capability.StorageGB)
	assert.Equal(t, int64(1000), capability.NetworkBandwidth)
	assert.Equal(t, 10*time.Millisecond, capability.Latency)
	assert.Equal(t, 1000.0, capability.Throughput)
	assert.Equal(t, 0.99, capability.Reliability)
	assert.Equal(t, 24*time.Hour, capability.Uptime)
	assert.Equal(t, 0.5, capability.LoadAverage)
	assert.Equal(t, 10, capability.ActiveConnections)
	assert.Equal(t, "us-east-1", capability.Region)
	assert.Equal(t, "us-east-1a", capability.Zone)
	assert.Equal(t, 0.8, capability.Priority)
}

func TestConflict_BasicOperations(t *testing.T) {
	conflict := &Conflict{
		ID:           "conflict-1",
		Key:          "test-key",
		ConflictType: ConflictTypeValue,
		Values: []*ConflictValue{
			{
				Value:      "value1",
				NodeID:     "node1",
				Timestamp:  time.Now(),
				Version:    1,
				Confidence: 0.9,
			},
			{
				Value:      "value2",
				NodeID:     "node2",
				Timestamp:  time.Now(),
				Version:    2,
				Confidence: 0.8,
			},
		},
		DetectedAt: time.Now(),
		Priority:   PriorityNormal,
		Status:     "pending",
		Metadata:   make(map[string]interface{}),
	}

	// Test initial state
	assert.Equal(t, "conflict-1", conflict.ID)
	assert.Equal(t, "test-key", conflict.Key)
	assert.Equal(t, ConflictTypeValue, conflict.ConflictType)
	assert.Len(t, conflict.Values, 2)
	assert.Equal(t, PriorityNormal, conflict.Priority)
	assert.Equal(t, ConflictStatus("pending"), conflict.Status)
}

func TestResolutionRule_Validation(t *testing.T) {
	tests := []struct {
		name    string
		rule    *ResolutionRule
		isValid bool
	}{
		{
			name: "valid rule",
			rule: &ResolutionRule{
				ID:          "rule-1",
				Name:        "Test Rule",
				Description: "A test rule",
				Strategy:    StrategyLastWriteWins,
				Priority:    PriorityNormal,
				Enabled:     true,
				CreatedAt:   time.Now(),
			},
			isValid: true,
		},
		{
			name: "empty ID",
			rule: &ResolutionRule{
				ID:          "",
				Name:        "Test Rule",
				Description: "A test rule",
				Strategy:    StrategyLastWriteWins,
				Priority:    PriorityNormal,
				Enabled:     true,
				CreatedAt:   time.Now(),
			},
			isValid: false,
		},
		{
			name: "empty name",
			rule: &ResolutionRule{
				ID:          "rule-1",
				Name:        "",
				Description: "A test rule",
				Strategy:    StrategyLastWriteWins,
				Priority:    PriorityNormal,
				Enabled:     true,
				CreatedAt:   time.Now(),
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.rule.ID != "" && tt.rule.Name != ""
			assert.Equal(t, tt.isValid, isValid)
		})
	}
}

func TestStateVersion_BasicOperations(t *testing.T) {
	stateVersion := &StateVersion{
		Version:         1,
		Hash:            "abc123",
		Timestamp:       time.Now(),
		Size:            1024,
		Checksum:        "def456",
		Metadata:        make(map[string]interface{}),
		PreviousVersion: 0,
		DeltaSize:       512,
		Changes:         make([]*StateChange, 0),
	}

	// Test initial state
	assert.Equal(t, int64(1), stateVersion.Version)
	assert.Equal(t, "abc123", stateVersion.Hash)
	assert.Equal(t, int64(1024), stateVersion.Size)
	assert.Equal(t, "def456", stateVersion.Checksum)
	assert.Equal(t, int64(0), stateVersion.PreviousVersion)
	assert.Equal(t, int64(512), stateVersion.DeltaSize)
	assert.NotNil(t, stateVersion.Metadata)
	assert.NotNil(t, stateVersion.Changes)
}

func TestStateChange_Validation(t *testing.T) {
	change := &StateChange{
		Type:      ChangeTypeSet,
		Key:       "test-key",
		OldValue:  "old-value",
		NewValue:  "new-value",
		Timestamp: time.Now(),
	}

	assert.Equal(t, ChangeTypeSet, change.Type)
	assert.Equal(t, "test-key", change.Key)
	assert.Equal(t, "old-value", change.OldValue)
	assert.Equal(t, "new-value", change.NewValue)
	assert.False(t, change.Timestamp.IsZero())
}

func TestElectionEvent_Creation(t *testing.T) {
	event := &ElectionEvent{
		Timestamp:  time.Now(),
		EventType:  ElectionStarted,
		OldLeader:  "old-leader",
		NewLeader:  "new-leader",
		Candidates: []raft.ServerID{"candidate1", "candidate2"},
		Reason:     "test election",
		Duration:   100 * time.Millisecond,
	}

	assert.False(t, event.Timestamp.IsZero())
	assert.Equal(t, ElectionStarted, event.EventType)
	assert.Equal(t, "old-leader", string(event.OldLeader))
	assert.Equal(t, "new-leader", string(event.NewLeader))
	assert.Len(t, event.Candidates, 2)
	assert.Equal(t, "test election", event.Reason)
	assert.Equal(t, 100*time.Millisecond, event.Duration)
}

func TestConflictTypes(t *testing.T) {
	// Test that conflict type constants are defined
	assert.Equal(t, ConflictType("value"), ConflictTypeValue)
	assert.Equal(t, ConflictType("version"), ConflictTypeVersion)
	assert.Equal(t, ConflictType("timestamp"), ConflictTypeTimestamp)
	assert.Equal(t, ConflictType("structural"), ConflictTypeStructural)
	assert.Equal(t, ConflictType("permission"), ConflictTypePermission)
}

func TestResolutionStrategies(t *testing.T) {
	// Test that resolution strategy constants are defined
	assert.Equal(t, ResolutionStrategy("last_write_wins"), StrategyLastWriteWins)
	assert.Equal(t, ResolutionStrategy("highest_version"), StrategyHighestVersion)
	assert.Equal(t, ResolutionStrategy("majority_vote"), StrategyMajorityVote)
	assert.Equal(t, ResolutionStrategy("highest_confidence"), StrategyHighestConfidence)
	assert.Equal(t, ResolutionStrategy("custom_rule"), StrategyCustomRule)
	assert.Equal(t, ResolutionStrategy("manual_review"), StrategyManualReview)
}

func TestConflictPriorities(t *testing.T) {
	// Test that priority constants are defined
	assert.Equal(t, ConflictPriority("low"), PriorityLow)
	assert.Equal(t, ConflictPriority("normal"), PriorityNormal)
	assert.Equal(t, ConflictPriority("high"), PriorityHigh)
	assert.Equal(t, ConflictPriority("critical"), PriorityCritical)
}

func TestElectionEventTypes(t *testing.T) {
	// Test that election event type constants are defined
	assert.Equal(t, ElectionEventType("election_started"), ElectionStarted)
	assert.Equal(t, ElectionEventType("election_completed"), ElectionCompleted)
	assert.Equal(t, ElectionEventType("leader_changed"), LeaderChanged)
	assert.Equal(t, ElectionEventType("leader_failed"), LeaderFailed)
}

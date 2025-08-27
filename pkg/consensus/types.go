package consensus

import (
	"context"
	"time"
)

// ConsensusEngine interface for consensus operations
type ConsensusEngine interface {
	Start(ctx context.Context) error
	Stop() error
	IsLeader() bool
	GetLeader() string
	ProposeChange(data []byte) error
}

// State represents consensus state
type State struct {
	Term     uint64    `json:"term"`
	Leader   string    `json:"leader"`
	LastApplied uint64 `json:"last_applied"`
	Updated  time.Time `json:"updated"`
}

// LogEntry represents a consensus log entry
type LogEntry struct {
	Index uint64 `json:"index"`
	Term  uint64 `json:"term"`
	Data  []byte `json:"data"`
}

// MockConsensusEngine is a simple mock implementation
type MockConsensusEngine struct {
	isLeader bool
	leader   string
}

func NewMockConsensusEngine() *MockConsensusEngine {
	return &MockConsensusEngine{
		isLeader: true,
		leader:   "self",
	}
}

func (m *MockConsensusEngine) Start(ctx context.Context) error {
	return nil
}

func (m *MockConsensusEngine) Stop() error {
	return nil
}

func (m *MockConsensusEngine) IsLeader() bool {
	return m.isLeader
}

func (m *MockConsensusEngine) GetLeader() string {
	return m.leader
}

func (m *MockConsensusEngine) ProposeChange(data []byte) error {
	return nil
}
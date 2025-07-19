package consensus

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/ollama/ollama-distributed/internal/config"
	"github.com/ollama/ollama-distributed/pkg/p2p"
)

// Engine represents the consensus engine using Raft
type Engine struct {
	config *config.ConsensusConfig
	p2p    *p2p.Node
	
	raft     *raft.Raft
	fsm      *FSM
	store    *raftboltdb.BoltStore
	snapshots raft.SnapshotStore
	transport *raft.NetworkTransport
	
	// Leadership tracking
	isLeader     bool
	leadershipMu sync.RWMutex
	leaderCh     chan bool
	
	// State management
	state   map[string]interface{}
	stateMu sync.RWMutex
	
	// Event channels
	applyCh chan *ApplyEvent
	
	started bool
	mu      sync.RWMutex
}

// ApplyEvent represents a state change event
type ApplyEvent struct {
	Type      string                 `json:"type"`
	Key       string                 `json:"key"`
	Value     interface{}            `json:"value"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// FSM implements the Raft finite state machine
type FSM struct {
	state   map[string]interface{}
	stateMu sync.RWMutex
	applyCh chan *ApplyEvent
}

// NewEngine creates a new consensus engine
func NewEngine(config *config.ConsensusConfig, p2pNode *p2p.Node) (*Engine, error) {
	engine := &Engine{
		config:    config,
		p2p:       p2pNode,
		state:     make(map[string]interface{}),
		leaderCh:  make(chan bool, 1),
		applyCh:   make(chan *ApplyEvent, 1000),
	}
	
	// Create FSM
	engine.fsm = &FSM{
		state:   make(map[string]interface{}),
		applyCh: engine.applyCh,
	}
	
	if err := engine.initRaft(); err != nil {
		return nil, fmt.Errorf("failed to initialize Raft: %w", err)
	}
	
	return engine, nil
}

// initRaft initializes the Raft consensus system
func (e *Engine) initRaft() error {
	// Create data directory
	if err := os.MkdirAll(e.config.DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}
	
	// Create Raft configuration
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(e.p2p.ID().String())
	raftConfig.LogLevel = e.config.LogLevel
	raftConfig.HeartbeatTimeout = e.config.HeartbeatTimeout
	raftConfig.ElectionTimeout = e.config.ElectionTimeout
	raftConfig.CommitTimeout = e.config.CommitTimeout
	raftConfig.MaxAppendEntries = e.config.MaxAppendEntries
	raftConfig.SnapshotInterval = e.config.SnapshotInterval
	raftConfig.SnapshotThreshold = e.config.SnapshotThreshold
	
	// Create log store
	logStore, err := raftboltdb.NewBoltStore(filepath.Join(e.config.DataDir, "raft-log.db"))
	if err != nil {
		return fmt.Errorf("failed to create log store: %w", err)
	}
	e.store = logStore
	
	// Create stable store
	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(e.config.DataDir, "raft-stable.db"))
	if err != nil {
		return fmt.Errorf("failed to create stable store: %w", err)
	}
	
	// Create snapshot store
	snapshots, err := raft.NewFileSnapshotStore(e.config.DataDir, 3, os.Stderr)
	if err != nil {
		return fmt.Errorf("failed to create snapshot store: %w", err)
	}
	e.snapshots = snapshots
	
	// Create transport
	addr, err := net.ResolveTCPAddr("tcp", e.config.BindAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve bind address: %w", err)
	}
	
	transport, err := raft.NewTCPTransport(e.config.BindAddr, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return fmt.Errorf("failed to create transport: %w", err)
	}
	e.transport = transport
	
	// Create Raft instance
	ra, err := raft.NewRaft(raftConfig, e.fsm, logStore, stableStore, snapshots, transport)
	if err != nil {
		return fmt.Errorf("failed to create Raft instance: %w", err)
	}
	e.raft = ra
	
	// Start leadership monitoring
	go e.monitorLeadership()
	
	return nil
}

// monitorLeadership monitors leadership changes
func (e *Engine) monitorLeadership() {
	for {
		select {
		case isLeader := <-e.raft.LeaderCh():
			e.leadershipMu.Lock()
			e.isLeader = isLeader
			e.leadershipMu.Unlock()
			
			// Notify leadership change
			select {
			case e.leaderCh <- isLeader:
			default:
			}
		}
	}
}

// Start starts the consensus engine
func (e *Engine) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if e.started {
		return fmt.Errorf("consensus engine already started")
	}
	
	// Bootstrap if configured
	if e.config.Bootstrap {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raft.ServerID(e.p2p.ID().String()),
					Address: e.transport.LocalAddr(),
				},
			},
		}
		
		e.raft.BootstrapCluster(configuration)
	}
	
	// Start event processing
	go e.processEvents()
	
	e.started = true
	return nil
}

// processEvents processes apply events
func (e *Engine) processEvents() {
	for event := range e.applyCh {
		// Update local state
		e.stateMu.Lock()
		e.state[event.Key] = event.Value
		e.stateMu.Unlock()
		
		// TODO: Notify subscribers
	}
}

// Apply applies a state change through Raft consensus
func (e *Engine) Apply(key string, value interface{}, metadata map[string]interface{}) error {
	if !e.IsLeader() {
		return fmt.Errorf("not leader, cannot apply changes")
	}
	
	event := &ApplyEvent{
		Type:      "set",
		Key:       key,
		Value:     value,
		Timestamp: time.Now(),
		Metadata:  metadata,
	}
	
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	
	future := e.raft.Apply(data, 10*time.Second)
	if err := future.Error(); err != nil {
		return fmt.Errorf("failed to apply change: %w", err)
	}
	
	return nil
}

// Get gets a value from the state
func (e *Engine) Get(key string) (interface{}, bool) {
	e.stateMu.RLock()
	defer e.stateMu.RUnlock()
	
	value, exists := e.state[key]
	return value, exists
}

// GetAll gets all state values
func (e *Engine) GetAll() map[string]interface{} {
	e.stateMu.RLock()
	defer e.stateMu.RUnlock()
	
	state := make(map[string]interface{})
	for k, v := range e.state {
		state[k] = v
	}
	
	return state
}

// Delete deletes a key from the state
func (e *Engine) Delete(key string) error {
	if !e.IsLeader() {
		return fmt.Errorf("not leader, cannot delete")
	}
	
	event := &ApplyEvent{
		Type:      "delete",
		Key:       key,
		Timestamp: time.Now(),
	}
	
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	
	future := e.raft.Apply(data, 10*time.Second)
	if err := future.Error(); err != nil {
		return fmt.Errorf("failed to apply delete: %w", err)
	}
	
	return nil
}

// IsLeader returns true if this node is the leader
func (e *Engine) IsLeader() bool {
	e.leadershipMu.RLock()
	defer e.leadershipMu.RUnlock()
	return e.isLeader
}

// Leader returns the current leader address
func (e *Engine) Leader() string {
	return string(e.raft.Leader())
}

// AddVoter adds a voting member to the cluster
func (e *Engine) AddVoter(id string, address string) error {
	if !e.IsLeader() {
		return fmt.Errorf("not leader, cannot add voter")
	}
	
	future := e.raft.AddVoter(raft.ServerID(id), raft.ServerAddress(address), 0, 10*time.Second)
	return future.Error()
}

// RemoveServer removes a server from the cluster
func (e *Engine) RemoveServer(id string) error {
	if !e.IsLeader() {
		return fmt.Errorf("not leader, cannot remove server")
	}
	
	future := e.raft.RemoveServer(raft.ServerID(id), 0, 10*time.Second)
	return future.Error()
}

// GetConfiguration returns the current cluster configuration
func (e *Engine) GetConfiguration() (*raft.Configuration, error) {
	future := e.raft.GetConfiguration()
	if err := future.Error(); err != nil {
		return nil, err
	}
	
	config := future.Configuration()
	return &config, nil
}

// LeadershipChanges returns a channel that receives leadership changes
func (e *Engine) LeadershipChanges() <-chan bool {
	return e.leaderCh
}

// Stats returns Raft statistics
func (e *Engine) Stats() map[string]string {
	return e.raft.Stats()
}

// Shutdown gracefully shuts down the consensus engine
func (e *Engine) Shutdown(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if !e.started {
		return nil
	}
	
	// Close apply channel
	close(e.applyCh)
	
	// Shutdown Raft
	if e.raft != nil {
		future := e.raft.Shutdown()
		if err := future.Error(); err != nil {
			return fmt.Errorf("failed to shutdown Raft: %w", err)
		}
	}
	
	// Close stores
	if e.store != nil {
		e.store.Close()
	}
	
	// Close transport
	if e.transport != nil {
		e.transport.Close()
	}
	
	e.started = false
	return nil
}

// FSM Methods

// Apply applies a log entry to the FSM
func (f *FSM) Apply(log *raft.Log) interface{} {
	var event ApplyEvent
	if err := json.Unmarshal(log.Data, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}
	
	// Use atomic operations to prevent race conditions
	f.stateMu.Lock()
	defer f.stateMu.Unlock()
	
	// Validate event before applying
	if err := f.validateEvent(&event); err != nil {
		return fmt.Errorf("invalid event: %w", err)
	}
	
	// Apply state changes atomically
	switch event.Type {
	case "set":
		f.state[event.Key] = event.Value
	case "delete":
		delete(f.state, event.Key)
	case "update":
		// Handle updates atomically
		if _, exists := f.state[event.Key]; exists {
			f.state[event.Key] = event.Value
		} else {
			return fmt.Errorf("cannot update non-existent key: %s", event.Key)
		}
	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}
	
	// Send to apply channel with timeout to prevent blocking
	select {
	case f.applyCh <- &event:
	case <-time.After(1 * time.Second):
		// Log warning but continue - don't fail the consensus operation
		log.Printf("Warning: apply channel full, dropping event notification for key %s", event.Key)
	}
	
	return nil
}

// validateEvent validates an event before applying it
func (f *FSM) validateEvent(event *ApplyEvent) error {
	if event.Key == "" {
		return fmt.Errorf("event key cannot be empty")
	}
	
	if event.Type == "" {
		return fmt.Errorf("event type cannot be empty")
	}
	
	// Validate timestamp is not too old
	if time.Since(event.Timestamp) > 5*time.Minute {
		return fmt.Errorf("event timestamp too old")
	}
	
	return nil
}

// Snapshot creates a snapshot of the FSM state
func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	f.stateMu.RLock()
	defer f.stateMu.RUnlock()
	
	// Clone state
	state := make(map[string]interface{})
	for k, v := range f.state {
		state[k] = v
	}
	
	return &fsmSnapshot{state: state}, nil
}

// Restore restores the FSM from a snapshot
func (f *FSM) Restore(snapshot io.ReadCloser) error {
	defer snapshot.Close()
	
	var state map[string]interface{}
	if err := json.NewDecoder(snapshot).Decode(&state); err != nil {
		return fmt.Errorf("failed to decode snapshot: %w", err)
	}
	
	f.stateMu.Lock()
	defer f.stateMu.Unlock()
	
	f.state = state
	return nil
}

// fsmSnapshot implements the raft.FSMSnapshot interface
type fsmSnapshot struct {
	state map[string]interface{}
}

// Persist persists the snapshot to the given sink
func (s *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	err := json.NewEncoder(sink).Encode(s.state)
	if err != nil {
		sink.Cancel()
		return fmt.Errorf("failed to encode snapshot: %w", err)
	}
	
	return sink.Close()
}

// Release releases the snapshot resources
func (s *fsmSnapshot) Release() {
	// Nothing to release
}
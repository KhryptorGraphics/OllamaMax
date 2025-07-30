package consensus

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/messaging"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/monitoring"
)

// Config is an alias for config.ConsensusConfig for backward compatibility
type Config = config.ConsensusConfig

// Engine represents the consensus engine using Raft
type Engine struct {
	config *config.ConsensusConfig
	p2p    *p2p.Node

	raft      *raft.Raft
	fsm       *FSM
	store     *raftboltdb.BoltStore
	snapshots raft.SnapshotStore
	transport raft.Transport

	// P2P integration
	messageRouter  *messaging.MessageRouter
	p2pTransport   *P2PTransport
	networkMonitor *monitoring.NetworkMonitor

	// Leadership tracking (atomic for thread safety)
	isLeader int64 // Use atomic operations
	leaderCh chan bool

	// Advanced leader election
	leaderElection *LeaderElectionManager

	// State synchronization
	stateSynchronizer *StateSynchronizer

	// Conflict resolution
	conflictResolver *ConflictResolver

	// State management
	state   map[string]interface{}
	stateMu sync.RWMutex

	// Event channels
	applyCh    chan *ApplyEvent
	shutdown   bool
	shutdownMu sync.RWMutex

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
	state      map[string]interface{}
	stateMu    sync.RWMutex
	applyCh    chan *ApplyEvent
	shutdown   bool
	shutdownMu sync.RWMutex
}

// NewEngine creates a new consensus engine
func NewEngine(config *config.ConsensusConfig, p2pNode *p2p.Node, messageRouter *messaging.MessageRouter, networkMonitor *monitoring.NetworkMonitor) (*Engine, error) {
	engine := &Engine{
		config:         config,
		p2p:            p2pNode,
		messageRouter:  messageRouter,
		networkMonitor: networkMonitor,
		state:          make(map[string]interface{}),
		leaderCh:       make(chan bool, 1),
		applyCh:        make(chan *ApplyEvent, 1000),
	}

	// Create FSM
	engine.fsm = &FSM{
		state:   make(map[string]interface{}),
		applyCh: engine.applyCh,
	}

	if err := engine.initRaft(); err != nil {
		return nil, fmt.Errorf("failed to initialize Raft: %w", err)
	}

	// Initialize leader election manager
	engine.leaderElection = NewLeaderElectionManager(engine, nil)

	// Initialize state synchronizer
	engine.stateSynchronizer = NewStateSynchronizer(engine, nil)

	// Initialize conflict resolver
	engine.conflictResolver = NewConflictResolver(engine, nil)

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
	nodeID := e.GetNodeID()
	raftConfig.LocalID = raft.ServerID(nodeID)
	raftConfig.LogLevel = e.config.LogLevel
	raftConfig.HeartbeatTimeout = e.config.HeartbeatTimeout
	raftConfig.ElectionTimeout = e.config.ElectionTimeout
	raftConfig.CommitTimeout = e.config.CommitTimeout
	raftConfig.MaxAppendEntries = e.config.MaxAppendEntries
	raftConfig.SnapshotInterval = e.config.SnapshotInterval
	raftConfig.SnapshotThreshold = e.config.SnapshotThreshold

	// Fix: Set LeaderLeaseTimeout to be less than HeartbeatTimeout
	// This prevents leadership oscillation issues
	raftConfig.LeaderLeaseTimeout = e.config.HeartbeatTimeout / 2

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

	// Create P2P transport
	localAddr := raft.ServerAddress(e.p2p.ID().String())
	p2pTransport, err := NewP2PTransport(nil, e.messageRouter, e.p2p.ID(), localAddr)
	if err != nil {
		return fmt.Errorf("failed to create P2P transport: %w", err)
	}
	e.p2pTransport = p2pTransport
	e.transport = p2pTransport

	// Create Raft instance
	ra, err := raft.NewRaft(raftConfig, e.fsm, logStore, stableStore, snapshots, e.transport)
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
			// Use atomic operations to prevent race conditions
			var leaderVal int64
			if isLeader {
				leaderVal = 1
			}
			atomic.StoreInt64(&e.isLeader, leaderVal)

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
					ID:      raft.ServerID(e.GetNodeID()),
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
		// Update local state based on event type
		e.stateMu.Lock()
		switch event.Type {
		case "set", "update":
			e.state[event.Key] = event.Value
		case "delete":
			delete(e.state, event.Key)
		}
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
	// First check FSM state (authoritative)
	e.fsm.stateMu.RLock()
	value, exists := e.fsm.state[key]
	e.fsm.stateMu.RUnlock()

	if exists {
		return value, true
	}

	// Fallback to engine state
	e.stateMu.RLock()
	defer e.stateMu.RUnlock()

	value, exists = e.state[key]
	return value, exists
}

// GetState gets all state values (alias for GetAll)
func (e *Engine) GetState() map[string]interface{} {
	return e.GetAll()
}

// GetAll gets all state values
func (e *Engine) GetAll() map[string]interface{} {
	// First get FSM state (authoritative)
	e.fsm.stateMu.RLock()
	fsmState := make(map[string]interface{})
	for k, v := range e.fsm.state {
		fsmState[k] = v
	}
	e.fsm.stateMu.RUnlock()

	// Merge with engine state (for any additional data)
	e.stateMu.RLock()
	defer e.stateMu.RUnlock()

	for k, v := range e.state {
		if _, exists := fsmState[k]; !exists {
			fsmState[k] = v
		}
	}

	return fsmState
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
	return atomic.LoadInt64(&e.isLeader) == 1
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

// GetCurrentTerm returns the current Raft term
func (e *Engine) GetCurrentTerm() uint64 {
	stats := e.raft.Stats()
	if termStr, exists := stats["term"]; exists {
		if term, err := strconv.ParseUint(termStr, 10, 64); err == nil {
			return term
		}
	}
	return 0
}

// Shutdown gracefully shuts down the consensus engine
func (e *Engine) Shutdown(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.started {
		return nil
	}

	// Set shutdown flag to prevent new channel sends
	e.fsm.shutdownMu.Lock()
	e.fsm.shutdown = true
	e.fsm.shutdownMu.Unlock()

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

	// Close P2P transport
	if e.p2pTransport != nil {
		e.p2pTransport.Close()
	}

	// Close leader election manager
	if e.leaderElection != nil {
		e.leaderElection.Close()
	}

	// Close state synchronizer
	if e.stateSynchronizer != nil {
		e.stateSynchronizer.Close()
	}

	// Close conflict resolver
	if e.conflictResolver != nil {
		e.conflictResolver.Close()
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
	// Use goroutine to prevent blocking the FSM apply operation
	// Check if FSM is shutting down before sending
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Channel was closed, ignore the panic - this is expected during shutdown
				fmt.Printf("Debug: apply channel closed during event send for key %s (expected during shutdown)\n", event.Key)
			}
		}()

		// Check if shutdown is in progress
		f.shutdownMu.RLock()
		isShutdown := f.shutdown
		f.shutdownMu.RUnlock()

		if isShutdown {
			// Skip sending to closed channel during shutdown
			return
		}

		select {
		case f.applyCh <- &event:
			// Successfully sent
		case <-time.After(1 * time.Second):
			// Log warning but continue - don't fail the consensus operation
			fmt.Printf("Warning: apply channel full, dropping event notification for key %s\n", event.Key)
		}
	}()

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

// GetNodeID returns the node ID
func (e *Engine) GetNodeID() string {
	if e.config.NodeID != "" {
		return e.config.NodeID
	}
	if e.p2p != nil {
		return e.p2p.ID().String()
	}
	return "unknown"
}

// IsStarted returns true if the engine is started
func (e *Engine) IsStarted() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.started
}

// Close gracefully shuts down the consensus engine (alias for Shutdown)
func (e *Engine) Close() error {
	return e.Shutdown(context.Background())
}

// Set sets a value in the state (alias for Apply)
func (e *Engine) Set(key string, value interface{}) error {
	return e.Apply(key, value, nil)
}

// SetWithContext sets a value in the state with context
func (e *Engine) SetWithContext(ctx context.Context, key string, value interface{}) error {
	return e.Apply(key, value, nil)
}

// IsRunning returns true if the engine is running (alias for IsStarted)
func (e *Engine) IsRunning() bool {
	return e.IsStarted()
}

// CreateSnapshot creates a snapshot of the current state
func (e *Engine) CreateSnapshot() error {
	if e.raft == nil {
		return fmt.Errorf("raft instance not initialized")
	}
	future := e.raft.Snapshot()
	return future.Error()
}

// ListSnapshots returns a list of available snapshots
func (e *Engine) ListSnapshots() ([]string, error) {
	// This is a simplified implementation
	// In practice, you would scan the snapshot directory
	return []string{}, nil
}

// RestoreSnapshot restores from a snapshot
func (e *Engine) RestoreSnapshot(snapshotID string) error {
	// This is a simplified implementation
	// In practice, you would restore from the specified snapshot
	return fmt.Errorf("snapshot restoration not implemented")
}

// GetLogSize returns the current size of the log
func (e *Engine) GetLogSize() (uint64, error) {
	if e.raft == nil {
		return 0, fmt.Errorf("raft instance not initialized")
	}
	stats := e.raft.Stats()
	// Try to parse the log size from stats
	if lastLogStr, exists := stats["last_log_index"]; exists {
		if lastLog, err := strconv.ParseUint(lastLogStr, 10, 64); err == nil {
			return lastLog, nil
		}
	}
	return 0, nil
}

// GetBindAddr returns the bind address
func (e *Engine) GetBindAddr() string {
	return e.config.BindAddr
}

// SetPeers sets the peers for the cluster
func (e *Engine) SetPeers(peers []string) error {
	// This is a simplified implementation
	// In practice, you would update the cluster configuration
	e.config.Peers = peers
	return nil
}

// GetAllKeys returns all keys in the state
func (e *Engine) GetAllKeys() []string {
	e.stateMu.RLock()
	defer e.stateMu.RUnlock()

	keys := make([]string, 0, len(e.state))
	for k := range e.state {
		keys = append(keys, k)
	}
	return keys
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

// GetLeaderElectionManager returns the leader election manager
func (e *Engine) GetLeaderElectionManager() *LeaderElectionManager {
	return e.leaderElection
}

// UpdateNodeCapability updates the capability information for this node
func (e *Engine) UpdateNodeCapability(capability *NodeCapability) {
	if e.leaderElection != nil {
		e.leaderElection.UpdateNodeCapability(capability)
	}
}

// GetBestLeaderCandidate returns the best candidate for leadership
func (e *Engine) GetBestLeaderCandidate() *NodeCapability {
	if e.leaderElection != nil {
		return e.leaderElection.GetBestLeaderCandidate()
	}
	return nil
}

// GetLeadershipRanking returns nodes ranked by leadership priority
func (e *Engine) GetLeadershipRanking() []*NodeCapability {
	if e.leaderElection != nil {
		return e.leaderElection.GetLeadershipRanking()
	}
	return nil
}

// GetElectionMetrics returns election metrics
func (e *Engine) GetElectionMetrics() *ElectionMetrics {
	if e.leaderElection != nil {
		return e.leaderElection.GetElectionMetrics()
	}
	return &ElectionMetrics{}
}

// GetStateSynchronizer returns the state synchronizer
func (e *Engine) GetStateSynchronizer() *StateSynchronizer {
	return e.stateSynchronizer
}

// CreateStateVersion creates a new state version snapshot
func (e *Engine) CreateStateVersion() (*StateVersion, error) {
	if e.stateSynchronizer != nil {
		return e.stateSynchronizer.CreateStateVersion()
	}
	return nil, fmt.Errorf("state synchronizer not initialized")
}

// SyncWithPeer synchronizes state with a specific peer
func (e *Engine) SyncWithPeer(peerID raft.ServerID, syncType SyncType, priority SyncPriority) (*SyncRequest, error) {
	if e.stateSynchronizer != nil {
		return e.stateSynchronizer.SyncWithPeer(peerID, syncType, priority)
	}
	return nil, fmt.Errorf("state synchronizer not initialized")
}

// GetSyncMetrics returns synchronization metrics
func (e *Engine) GetSyncMetrics() *SyncMetrics {
	if e.stateSynchronizer != nil {
		return e.stateSynchronizer.GetSyncMetrics()
	}
	return &SyncMetrics{}
}

// GetConflictResolver returns the conflict resolver
func (e *Engine) GetConflictResolver() *ConflictResolver {
	return e.conflictResolver
}

// DetectConflict detects and registers a new conflict
func (e *Engine) DetectConflict(key string, values []*ConflictValue) *Conflict {
	if e.conflictResolver != nil {
		return e.conflictResolver.DetectConflict(key, values)
	}
	return nil
}

// ResolveConflict resolves a specific conflict
func (e *Engine) ResolveConflict(conflictID string) error {
	if e.conflictResolver != nil {
		return e.conflictResolver.ResolveConflict(conflictID)
	}
	return fmt.Errorf("conflict resolver not initialized")
}

// GetActiveConflicts returns all active conflicts
func (e *Engine) GetActiveConflicts() []*Conflict {
	if e.conflictResolver != nil {
		return e.conflictResolver.GetActiveConflicts()
	}
	return nil
}

// GetConflictMetrics returns conflict resolution metrics
func (e *Engine) GetConflictMetrics() *ConflictMetrics {
	if e.conflictResolver != nil {
		return e.conflictResolver.GetConflictMetrics()
	}
	return &ConflictMetrics{}
}

package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

// ConsensusHandler handles consensus protocol messages
type ConsensusHandler struct {
	nodeID    peer.ID
	callbacks map[string]ConsensusCallback
}

// SchedulerHandler handles scheduler protocol messages
type SchedulerHandler struct {
	nodeID    peer.ID
	callbacks map[string]SchedulerCallback
}

// ModelHandler handles model management protocol messages
type ModelHandler struct {
	nodeID    peer.ID
	callbacks map[string]ModelCallback
}

// DiscoveryHandler handles discovery protocol messages
type DiscoveryHandler struct {
	nodeID    peer.ID
	callbacks map[string]DiscoveryCallback
}

// HealthHandler handles health check protocol messages
type HealthHandler struct {
	nodeID    peer.ID
	callbacks map[string]HealthCallback
}

// DataHandler handles data transfer protocol messages
type DataHandler struct {
	nodeID    peer.ID
	callbacks map[string]DataCallback
}

// Callback function types
type ConsensusCallback func(ctx context.Context, msg *ConsensusMessage) error
type SchedulerCallback func(ctx context.Context, msg *SchedulerMessage) error
type ModelCallback func(ctx context.Context, msg *ModelMessage) error
type DiscoveryCallback func(ctx context.Context, msg *DiscoveryMessage) error
type HealthCallback func(ctx context.Context, msg *HealthMessage) error
type DataCallback func(ctx context.Context, msg *DataMessage) error

// Message payload structures

// ConsensusMessage represents a consensus protocol message
type ConsensusMessage struct {
	Type         ConsensusMessageType `json:"type"`
	Term         uint64               `json:"term"`
	LeaderID     peer.ID              `json:"leader_id"`
	CandidateID  peer.ID              `json:"candidate_id,omitempty"`
	VoteGranted  bool                 `json:"vote_granted,omitempty"`
	PrevLogIndex uint64               `json:"prev_log_index,omitempty"`
	PrevLogTerm  uint64               `json:"prev_log_term,omitempty"`
	Entries      []LogEntry           `json:"entries,omitempty"`
	LeaderCommit uint64               `json:"leader_commit,omitempty"`
	Success      bool                 `json:"success,omitempty"`
}

// SchedulerMessage represents a scheduler protocol message
type SchedulerMessage struct {
	Type      SchedulerMessageType  `json:"type"`
	TaskID    string                `json:"task_id"`
	JobID     string                `json:"job_id"`
	WorkerID  peer.ID               `json:"worker_id"`
	Task      *Task                 `json:"task,omitempty"`
	Result    *TaskResult           `json:"result,omitempty"`
	Status    TaskStatus            `json:"status,omitempty"`
	Priority  int                   `json:"priority,omitempty"`
	Deadline  time.Time             `json:"deadline,omitempty"`
	Resources *ResourceRequirements `json:"resources,omitempty"`
}

// ModelMessage represents a model management protocol message
type ModelMessage struct {
	Type        ModelMessageType `json:"type"`
	ModelID     string           `json:"model_id"`
	Version     string           `json:"version"`
	ChunkID     string           `json:"chunk_id,omitempty"`
	ChunkIndex  int              `json:"chunk_index,omitempty"`
	TotalChunks int              `json:"total_chunks,omitempty"`
	Data        []byte           `json:"data,omitempty"`
	Checksum    string           `json:"checksum,omitempty"`
	Metadata    *ModelMetadata   `json:"metadata,omitempty"`
	Status      ModelStatus      `json:"status,omitempty"`
}

// DiscoveryMessage represents a discovery protocol message
type DiscoveryMessage struct {
	Type         DiscoveryMessageType `json:"type"`
	NodeID       peer.ID              `json:"node_id"`
	Capabilities *NodeCapabilities    `json:"capabilities,omitempty"`
	Services     []ServiceInfo        `json:"services,omitempty"`
	Location     *GeographicLocation  `json:"location,omitempty"`
	Timestamp    time.Time            `json:"timestamp"`
	TTL          time.Duration        `json:"ttl"`
}

// HealthMessage represents a health check protocol message
type HealthMessage struct {
	Type      HealthMessageType `json:"type"`
	NodeID    peer.ID           `json:"node_id"`
	Status    HealthStatus      `json:"status"`
	Metrics   *HealthMetrics    `json:"metrics,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	RequestID string            `json:"request_id,omitempty"`
}

// DataMessage represents a data transfer protocol message
type DataMessage struct {
	Type        DataMessageType `json:"type"`
	TransferID  string          `json:"transfer_id"`
	SequenceNum uint64          `json:"sequence_num"`
	Data        []byte          `json:"data"`
	Checksum    string          `json:"checksum"`
	IsLast      bool            `json:"is_last"`
	Compressed  bool            `json:"compressed"`
}

// Supporting data structures

type LogEntry struct {
	Index   uint64      `json:"index"`
	Term    uint64      `json:"term"`
	Command interface{} `json:"command"`
}

type Task struct {
	ID        string                `json:"id"`
	Type      string                `json:"type"`
	Payload   []byte                `json:"payload"`
	Resources *ResourceRequirements `json:"resources"`
	Deadline  time.Time             `json:"deadline"`
	Priority  int                   `json:"priority"`
	Metadata  map[string]string     `json:"metadata"`
}

type TaskResult struct {
	TaskID      string             `json:"task_id"`
	Status      TaskStatus         `json:"status"`
	Result      []byte             `json:"result"`
	Error       string             `json:"error,omitempty"`
	Metrics     map[string]float64 `json:"metrics"`
	CompletedAt time.Time          `json:"completed_at"`
}

type ResourceRequirements struct {
	CPU       float64 `json:"cpu"`
	Memory    int64   `json:"memory"`
	GPU       int     `json:"gpu"`
	Storage   int64   `json:"storage"`
	Bandwidth int64   `json:"bandwidth"`
}

type ModelMetadata struct {
	Name       string            `json:"name"`
	Version    string            `json:"version"`
	Size       int64             `json:"size"`
	Format     string            `json:"format"`
	Checksum   string            `json:"checksum"`
	Tags       []string          `json:"tags"`
	Parameters map[string]string `json:"parameters"`
	CreatedAt  time.Time         `json:"created_at"`
}

type NodeCapabilities struct {
	CPU       float64           `json:"cpu"`
	Memory    int64             `json:"memory"`
	GPU       int               `json:"gpu"`
	Storage   int64             `json:"storage"`
	Bandwidth int64             `json:"bandwidth"`
	Models    []string          `json:"models"`
	Features  []string          `json:"features"`
	Metadata  map[string]string `json:"metadata"`
}

type ServiceInfo struct {
	Name     string            `json:"name"`
	Version  string            `json:"version"`
	Endpoint string            `json:"endpoint"`
	Protocol string            `json:"protocol"`
	Status   string            `json:"status"`
	Metadata map[string]string `json:"metadata"`
}

type GeographicLocation struct {
	Country   string  `json:"country"`
	Region    string  `json:"region"`
	City      string  `json:"city"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
}

type HealthMetrics struct {
	CPUUsage    float64            `json:"cpu_usage"`
	MemoryUsage float64            `json:"memory_usage"`
	DiskUsage   float64            `json:"disk_usage"`
	NetworkIO   NetworkIOMetrics   `json:"network_io"`
	Uptime      time.Duration      `json:"uptime"`
	LoadAverage []float64          `json:"load_average"`
	Custom      map[string]float64 `json:"custom"`
}

type NetworkIOMetrics struct {
	BytesIn    int64 `json:"bytes_in"`
	BytesOut   int64 `json:"bytes_out"`
	PacketsIn  int64 `json:"packets_in"`
	PacketsOut int64 `json:"packets_out"`
	ErrorsIn   int64 `json:"errors_in"`
	ErrorsOut  int64 `json:"errors_out"`
}

// Enums
type ConsensusMessageType string

const (
	ConsensusRequestVote     ConsensusMessageType = "request_vote"
	ConsensusVoteResponse    ConsensusMessageType = "vote_response"
	ConsensusAppendEntries   ConsensusMessageType = "append_entries"
	ConsensusAppendResponse  ConsensusMessageType = "append_response"
	ConsensusHeartbeat       ConsensusMessageType = "heartbeat"
	ConsensusInstallSnapshot ConsensusMessageType = "install_snapshot"
)

type SchedulerMessageType string

const (
	SchedulerTaskAssignment  SchedulerMessageType = "task_assignment"
	SchedulerTaskResult      SchedulerMessageType = "task_result"
	SchedulerTaskStatus      SchedulerMessageType = "task_status"
	SchedulerResourceUpdate  SchedulerMessageType = "resource_update"
	SchedulerJobCancel       SchedulerMessageType = "job_cancel"
	SchedulerWorkerRegister  SchedulerMessageType = "worker_register"
	SchedulerWorkerHeartbeat SchedulerMessageType = "worker_heartbeat"
)

type ModelMessageType string

const (
	ModelRequest     ModelMessageType = "model_request"
	ModelResponse    ModelMessageType = "model_response"
	ModelChunk       ModelMessageType = "model_chunk"
	ModelChunkAck    ModelMessageType = "model_chunk_ack"
	ModelMetadataMsg ModelMessageType = "model_metadata"
	ModelSync        ModelMessageType = "model_sync"
	ModelReplication ModelMessageType = "model_replication"
	ModelDelete      ModelMessageType = "model_delete"
)

type DiscoveryMessageType string

const (
	DiscoveryAnnounce  DiscoveryMessageType = "announce"
	DiscoveryQuery     DiscoveryMessageType = "query"
	DiscoveryResponse  DiscoveryMessageType = "response"
	DiscoveryHeartbeat DiscoveryMessageType = "heartbeat"
	DiscoveryGoodbye   DiscoveryMessageType = "goodbye"
)

type HealthMessageType string

const (
	HealthPing            HealthMessageType = "ping"
	HealthPong            HealthMessageType = "pong"
	HealthStatusMsg       HealthMessageType = "status"
	HealthMetricsRequest  HealthMessageType = "metrics_request"
	HealthMetricsResponse HealthMessageType = "metrics_response"
)

type DataMessageType string

const (
	DataTransferStart    DataMessageType = "transfer_start"
	DataTransferChunk    DataMessageType = "transfer_chunk"
	DataTransferAck      DataMessageType = "transfer_ack"
	DataTransferComplete DataMessageType = "transfer_complete"
	DataTransferError    DataMessageType = "transfer_error"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

type ModelStatus string

const (
	ModelStatusAvailable    ModelStatus = "available"
	ModelStatusTransferring ModelStatus = "transferring"
	ModelStatusCorrupted    ModelStatus = "corrupted"
	ModelStatusMissing      ModelStatus = "missing"
)

type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// Protocol handler implementations

// NewConsensusHandler creates a new consensus protocol handler
func NewConsensusHandler(nodeID peer.ID) *ConsensusHandler {
	return &ConsensusHandler{
		nodeID:    nodeID,
		callbacks: make(map[string]ConsensusCallback),
	}
}

func (ch *ConsensusHandler) HandleMessage(ctx context.Context, msg *Message) error {
	var consensusMsg ConsensusMessage
	if err := json.Unmarshal(msg.Payload, &consensusMsg); err != nil {
		return fmt.Errorf("failed to unmarshal consensus message: %w", err)
	}

	callback, exists := ch.callbacks[string(consensusMsg.Type)]
	if !exists {
		return fmt.Errorf("no callback registered for consensus message type: %s", consensusMsg.Type)
	}

	return callback(ctx, &consensusMsg)
}

func (ch *ConsensusHandler) GetProtocol() protocol.ID {
	return ProtocolConsensus
}

func (ch *ConsensusHandler) GetMessageTypes() []MessageType {
	return []MessageType{MessageTypeConsensus}
}

func (ch *ConsensusHandler) RegisterCallback(msgType ConsensusMessageType, callback ConsensusCallback) {
	ch.callbacks[string(msgType)] = callback
}

// NewSchedulerHandler creates a new scheduler protocol handler
func NewSchedulerHandler(nodeID peer.ID) *SchedulerHandler {
	return &SchedulerHandler{
		nodeID:    nodeID,
		callbacks: make(map[string]SchedulerCallback),
	}
}

func (sh *SchedulerHandler) HandleMessage(ctx context.Context, msg *Message) error {
	var schedulerMsg SchedulerMessage
	if err := json.Unmarshal(msg.Payload, &schedulerMsg); err != nil {
		return fmt.Errorf("failed to unmarshal scheduler message: %w", err)
	}

	callback, exists := sh.callbacks[string(schedulerMsg.Type)]
	if !exists {
		return fmt.Errorf("no callback registered for scheduler message type: %s", schedulerMsg.Type)
	}

	return callback(ctx, &schedulerMsg)
}

func (sh *SchedulerHandler) GetProtocol() protocol.ID {
	return ProtocolScheduler
}

func (sh *SchedulerHandler) GetMessageTypes() []MessageType {
	return []MessageType{MessageTypeScheduler}
}

func (sh *SchedulerHandler) RegisterCallback(msgType SchedulerMessageType, callback SchedulerCallback) {
	sh.callbacks[string(msgType)] = callback
}

// NewModelHandler creates a new model protocol handler
func NewModelHandler(nodeID peer.ID) *ModelHandler {
	return &ModelHandler{
		nodeID:    nodeID,
		callbacks: make(map[string]ModelCallback),
	}
}

func (mh *ModelHandler) HandleMessage(ctx context.Context, msg *Message) error {
	var modelMsg ModelMessage
	if err := json.Unmarshal(msg.Payload, &modelMsg); err != nil {
		return fmt.Errorf("failed to unmarshal model message: %w", err)
	}

	callback, exists := mh.callbacks[string(modelMsg.Type)]
	if !exists {
		return fmt.Errorf("no callback registered for model message type: %s", modelMsg.Type)
	}

	return callback(ctx, &modelMsg)
}

func (mh *ModelHandler) GetProtocol() protocol.ID {
	return ProtocolModel
}

func (mh *ModelHandler) GetMessageTypes() []MessageType {
	return []MessageType{MessageTypeModel}
}

func (mh *ModelHandler) RegisterCallback(msgType ModelMessageType, callback ModelCallback) {
	mh.callbacks[string(msgType)] = callback
}

// NewDiscoveryHandler creates a new discovery protocol handler
func NewDiscoveryHandler(nodeID peer.ID) *DiscoveryHandler {
	return &DiscoveryHandler{
		nodeID:    nodeID,
		callbacks: make(map[string]DiscoveryCallback),
	}
}

func (dh *DiscoveryHandler) HandleMessage(ctx context.Context, msg *Message) error {
	var discoveryMsg DiscoveryMessage
	if err := json.Unmarshal(msg.Payload, &discoveryMsg); err != nil {
		return fmt.Errorf("failed to unmarshal discovery message: %w", err)
	}

	callback, exists := dh.callbacks[string(discoveryMsg.Type)]
	if !exists {
		return fmt.Errorf("no callback registered for discovery message type: %s", discoveryMsg.Type)
	}

	return callback(ctx, &discoveryMsg)
}

func (dh *DiscoveryHandler) GetProtocol() protocol.ID {
	return ProtocolDiscovery
}

func (dh *DiscoveryHandler) GetMessageTypes() []MessageType {
	return []MessageType{MessageTypeDiscovery}
}

func (dh *DiscoveryHandler) RegisterCallback(msgType DiscoveryMessageType, callback DiscoveryCallback) {
	dh.callbacks[string(msgType)] = callback
}

// NewHealthHandler creates a new health protocol handler
func NewHealthHandler(nodeID peer.ID) *HealthHandler {
	return &HealthHandler{
		nodeID:    nodeID,
		callbacks: make(map[string]HealthCallback),
	}
}

func (hh *HealthHandler) HandleMessage(ctx context.Context, msg *Message) error {
	var healthMsg HealthMessage
	if err := json.Unmarshal(msg.Payload, &healthMsg); err != nil {
		return fmt.Errorf("failed to unmarshal health message: %w", err)
	}

	callback, exists := hh.callbacks[string(healthMsg.Type)]
	if !exists {
		return fmt.Errorf("no callback registered for health message type: %s", healthMsg.Type)
	}

	return callback(ctx, &healthMsg)
}

func (hh *HealthHandler) GetProtocol() protocol.ID {
	return ProtocolHealth
}

func (hh *HealthHandler) GetMessageTypes() []MessageType {
	return []MessageType{MessageTypeHealth}
}

func (hh *HealthHandler) RegisterCallback(msgType HealthMessageType, callback HealthCallback) {
	hh.callbacks[string(msgType)] = callback
}

// NewDataHandler creates a new data protocol handler
func NewDataHandler(nodeID peer.ID) *DataHandler {
	return &DataHandler{
		nodeID:    nodeID,
		callbacks: make(map[string]DataCallback),
	}
}

func (dh *DataHandler) HandleMessage(ctx context.Context, msg *Message) error {
	var dataMsg DataMessage
	if err := json.Unmarshal(msg.Payload, &dataMsg); err != nil {
		return fmt.Errorf("failed to unmarshal data message: %w", err)
	}

	callback, exists := dh.callbacks[string(dataMsg.Type)]
	if !exists {
		return fmt.Errorf("no callback registered for data message type: %s", dataMsg.Type)
	}

	return callback(ctx, &dataMsg)
}

func (dh *DataHandler) GetProtocol() protocol.ID {
	return ProtocolData
}

func (dh *DataHandler) GetMessageTypes() []MessageType {
	return []MessageType{MessageTypeData}
}

func (dh *DataHandler) RegisterCallback(msgType DataMessageType, callback DataCallback) {
	dh.callbacks[string(msgType)] = callback
}

// Helper functions for creating messages

// CreateConsensusMessage creates a consensus message
func CreateConsensusMessage(msgType ConsensusMessageType, source, dest peer.ID, payload *ConsensusMessage) (*Message, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal consensus message: %w", err)
	}

	return &Message{
		ID:          generateMessageID(),
		Type:        MessageTypeConsensus,
		Protocol:    ProtocolConsensus,
		Source:      source,
		Destination: dest,
		Payload:     data,
		Headers:     make(map[string]string),
		Timestamp:   time.Now(),
		TTL:         30 * time.Second,
		Priority:    PriorityHigh,
		RequiresAck: true,
	}, nil
}

// CreateSchedulerMessage creates a scheduler message
func CreateSchedulerMessage(msgType SchedulerMessageType, source, dest peer.ID, payload *SchedulerMessage) (*Message, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal scheduler message: %w", err)
	}

	return &Message{
		ID:          generateMessageID(),
		Type:        MessageTypeScheduler,
		Protocol:    ProtocolScheduler,
		Source:      source,
		Destination: dest,
		Payload:     data,
		Headers:     make(map[string]string),
		Timestamp:   time.Now(),
		TTL:         60 * time.Second,
		Priority:    PriorityNormal,
		RequiresAck: true,
	}, nil
}

// CreateModelMessage creates a model message
func CreateModelMessage(msgType ModelMessageType, source, dest peer.ID, payload *ModelMessage) (*Message, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal model message: %w", err)
	}

	return &Message{
		ID:          generateMessageID(),
		Type:        MessageTypeModel,
		Protocol:    ProtocolModel,
		Source:      source,
		Destination: dest,
		Payload:     data,
		Headers:     make(map[string]string),
		Timestamp:   time.Now(),
		TTL:         300 * time.Second, // Longer TTL for model transfers
		Priority:    PriorityNormal,
		RequiresAck: true,
	}, nil
}

// CreateDiscoveryMessage creates a discovery message
func CreateDiscoveryMessage(msgType DiscoveryMessageType, source, dest peer.ID, payload *DiscoveryMessage) (*Message, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal discovery message: %w", err)
	}

	return &Message{
		ID:          generateMessageID(),
		Type:        MessageTypeDiscovery,
		Protocol:    ProtocolDiscovery,
		Source:      source,
		Destination: dest,
		Payload:     data,
		Headers:     make(map[string]string),
		Timestamp:   time.Now(),
		TTL:         30 * time.Second,
		Priority:    PriorityNormal,
		RequiresAck: false, // Discovery messages don't need acks
	}, nil
}

// CreateHealthMessage creates a health message
func CreateHealthMessage(msgType HealthMessageType, source, dest peer.ID, payload *HealthMessage) (*Message, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal health message: %w", err)
	}

	return &Message{
		ID:          generateMessageID(),
		Type:        MessageTypeHealth,
		Protocol:    ProtocolHealth,
		Source:      source,
		Destination: dest,
		Payload:     data,
		Headers:     make(map[string]string),
		Timestamp:   time.Now(),
		TTL:         10 * time.Second,
		Priority:    PriorityLow,
		RequiresAck: false, // Health messages don't need acks
	}, nil
}

// CreateDataMessage creates a data message
func CreateDataMessage(msgType DataMessageType, source, dest peer.ID, payload *DataMessage) (*Message, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data message: %w", err)
	}

	return &Message{
		ID:          generateMessageID(),
		Type:        MessageTypeData,
		Protocol:    ProtocolData,
		Source:      source,
		Destination: dest,
		Payload:     data,
		Headers:     make(map[string]string),
		Timestamp:   time.Now(),
		TTL:         120 * time.Second,
		Priority:    PriorityNormal,
		RequiresAck: true,
	}, nil
}

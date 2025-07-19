package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/ollama/ollama-distributed/internal/config"
	"github.com/ollama/ollama-distributed/pkg/consensus"
	"github.com/ollama/ollama-distributed/pkg/p2p"
)

// Engine represents the distributed scheduling engine
type Engine struct {
	config    *config.SchedulerConfig
	p2p       *p2p.Node
	consensus *consensus.Engine

	// Model registry
	models   map[string]*ModelInfo
	modelsMu sync.RWMutex

	// Node registry
	nodes   map[string]*NodeInfo
	nodesMu sync.RWMutex

	// Request queue
	requests chan *Request

	// Workers
	workers   []*Worker
	workersMu sync.RWMutex

	// Health checker
	healthChecker *HealthChecker

	// Load balancer
	loadBalancer *LoadBalancer

	// Statistics
	stats     *Stats
	statsMu   sync.RWMutex
	startTime time.Time

	started bool
	mu      sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
}

// ModelInfo represents information about a model
type ModelInfo struct {
	Name         string            `json:"name"`
	Size         int64             `json:"size"`
	Checksum     string            `json:"checksum"`
	Locations    []string          `json:"locations"` // Node IDs that have this model
	AccessCount  int64             `json:"access_count"`
	LastAccessed time.Time         `json:"last_accessed"`
	Metadata     map[string]string `json:"metadata"`
}

// NodeInfo represents information about a node
type NodeInfo struct {
	ID       string            `json:"id"`
	Address  string            `json:"address"`
	Status   NodeStatus        `json:"status"`
	Capacity NodeCapacity      `json:"capacity"`
	Usage    NodeUsage         `json:"usage"`
	Models   []string          `json:"models"`
	LastSeen time.Time         `json:"last_seen"`
	Metadata map[string]string `json:"metadata"`
}

// NodeStatus represents the status of a node
type NodeStatus string

const (
	NodeStatusOnline      NodeStatus = "online"
	NodeStatusOffline     NodeStatus = "offline"
	NodeStatusDraining    NodeStatus = "draining"
	NodeStatusMaintenance NodeStatus = "maintenance"
)

// NodeCapacity represents the capacity of a node
type NodeCapacity struct {
	CPU    int64 `json:"cpu"`    // CPU cores
	Memory int64 `json:"memory"` // Memory in bytes
	Disk   int64 `json:"disk"`   // Disk space in bytes
	GPU    int64 `json:"gpu"`    // GPU count
}

// NodeUsage represents the current usage of a node
type NodeUsage struct {
	CPU    float64 `json:"cpu"`    // CPU usage percentage
	Memory float64 `json:"memory"` // Memory usage percentage
	Disk   float64 `json:"disk"`   // Disk usage percentage
	GPU    float64 `json:"gpu"`    // GPU usage percentage
}

// Request represents a request for model inference
type Request struct {
	ID        string            `json:"id"`
	ModelName string            `json:"model_name"`
	Type      string            `json:"type"`
	Priority  int               `json:"priority"`
	Timeout   time.Duration     `json:"timeout"`
	Metadata  map[string]string `json:"metadata"`
	Payload   map[string]interface{} `json:"payload"`

	// Response channel
	ResponseCh chan *Response

	// Timing
	CreatedAt   time.Time `json:"created_at"`
	ScheduledAt time.Time `json:"scheduled_at"`
	CompletedAt time.Time `json:"completed_at"`
}

// Response represents a response to a request
type Response struct {
	RequestID string        `json:"request_id"`
	NodeID    string        `json:"node_id"`
	Success   bool          `json:"success"`
	Error     string        `json:"error,omitempty"`
	Data      interface{}   `json:"data,omitempty"`
	Duration  time.Duration `json:"duration"`
}

// Stats represents scheduler statistics
type Stats struct {
	TotalRequests     int64         `json:"total_requests"`
	CompletedRequests int64         `json:"completed_requests"`
	FailedRequests    int64         `json:"failed_requests"`
	QueuedRequests    int64         `json:"queued_requests"`
	AverageLatency    time.Duration `json:"average_latency"`
	NodesTotal        int           `json:"nodes_total"`
	NodesOnline       int           `json:"nodes_online"`
	NodesOffline      int           `json:"nodes_offline"`
	ModelsTotal       int           `json:"models_total"`
	WorkersActive     int           `json:"workers_active"`
	Uptime            time.Duration `json:"uptime"`
	LastUpdated       time.Time     `json:"last_updated"`
}

// Worker represents a worker that processes requests
type Worker struct {
	ID     int
	engine *Engine
	stopCh chan struct{}
}

// HealthChecker monitors node health
type HealthChecker struct {
	engine   *Engine
	interval time.Duration
	stopCh   chan struct{}
}

// LoadBalancer handles load balancing algorithms
type LoadBalancer struct {
	algorithm string
	engine    *Engine
}

// NewEngine creates a new scheduling engine
func NewEngine(config *config.SchedulerConfig, p2pNode *p2p.Node, consensusEngine *consensus.Engine) (*Engine, error) {
	ctx, cancel := context.WithCancel(context.Background())

	engine := &Engine{
		config:    config,
		p2p:       p2pNode,
		consensus: consensusEngine,
		models:    make(map[string]*ModelInfo),
		nodes:     make(map[string]*NodeInfo),
		requests:  make(chan *Request, config.QueueSize),
		stats:     &Stats{LastUpdated: time.Now()},
		startTime: time.Now(),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Initialize health checker
	engine.healthChecker = &HealthChecker{
		engine:   engine,
		interval: config.HealthCheckInterval,
		stopCh:   make(chan struct{}),
	}

	// Initialize load balancer
	engine.loadBalancer = &LoadBalancer{
		algorithm: config.LoadBalancing,
		engine:    engine,
	}

	// Create workers
	engine.workers = make([]*Worker, config.WorkerCount)
	for i := 0; i < config.WorkerCount; i++ {
		engine.workers[i] = &Worker{
			ID:     i,
			engine: engine,
			stopCh: make(chan struct{}),
		}
	}

	return engine, nil
}

// Start starts the scheduling engine
func (e *Engine) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.started {
		return fmt.Errorf("scheduler already started")
	}

	// Start workers
	for _, worker := range e.workers {
		go worker.start()
	}

	// Start health checker
	go e.healthChecker.start()

	// Start node discovery
	go e.discoverNodes()

	// Start model registry sync
	go e.syncModelRegistry()

	e.started = true
	return nil
}

// discoverNodes discovers nodes in the network
func (e *Engine) discoverNodes() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.updateNodeRegistry()
		}
	}
}

// updateNodeRegistry updates the node registry from P2P peers
func (e *Engine) updateNodeRegistry() {
	peers := e.p2p.GetAllPeers()

	e.nodesMu.Lock()
	defer e.nodesMu.Unlock()

	// Update existing nodes and add new ones
	for peerID, peerInfo := range peers {
		nodeID := peerID.String()

		if node, exists := e.nodes[nodeID]; exists {
			// Update existing node
			node.Status = NodeStatusOnline
			node.LastSeen = time.Now()
		} else {
			// Add new node
			e.nodes[nodeID] = &NodeInfo{
				ID:       nodeID,
				Address:  peerInfo.Addresses[0].String(),
				Status:   NodeStatusOnline,
				Capacity: NodeCapacity{}, // TODO: Get actual capacity
				Usage:    NodeUsage{},    // TODO: Get actual usage
				Models:   []string{},
				LastSeen: time.Now(),
				Metadata: make(map[string]string),
			}
		}
	}

	// Mark offline nodes
	for nodeID, node := range e.nodes {
		if time.Since(node.LastSeen) > 5*time.Minute {
			node.Status = NodeStatusOffline
		}
	}
}

// syncModelRegistry syncs the model registry with consensus
func (e *Engine) syncModelRegistry() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.syncModels()
		}
	}
}

// syncModels syncs model information with consensus
func (e *Engine) syncModels() {
	// Get model registry from consensus
	if registry, exists := e.consensus.Get("model_registry"); exists {
		if models, ok := registry.(map[string]*ModelInfo); ok {
			e.modelsMu.Lock()
			e.models = models
			e.modelsMu.Unlock()
		}
	}

	// Update consensus with local changes
	if e.consensus.IsLeader() {
		e.modelsMu.RLock()
		models := make(map[string]*ModelInfo)
		for k, v := range e.models {
			models[k] = v
		}
		e.modelsMu.RUnlock()

		e.consensus.Apply("model_registry", models, nil)
	}
}

// Schedule schedules a request for execution
func (e *Engine) Schedule(req *Request) error {
	req.CreatedAt = time.Now()

	select {
	case e.requests <- req:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("request queue full")
	}
}

// RegisterModel registers a model in the registry
func (e *Engine) RegisterModel(name string, size int64, checksum string, nodeID string) error {
	e.modelsMu.Lock()
	defer e.modelsMu.Unlock()

	if model, exists := e.models[name]; exists {
		// Update existing model
		if !contains(model.Locations, nodeID) {
			model.Locations = append(model.Locations, nodeID)
		}
	} else {
		// Create new model
		e.models[name] = &ModelInfo{
			Name:         name,
			Size:         size,
			Checksum:     checksum,
			Locations:    []string{nodeID},
			AccessCount:  0,
			LastAccessed: time.Now(),
			Metadata:     make(map[string]string),
		}
	}

	return nil
}

// GetModel gets model information
func (e *Engine) GetModel(name string) (*ModelInfo, bool) {
	e.modelsMu.RLock()
	defer e.modelsMu.RUnlock()

	model, exists := e.models[name]
	return model, exists
}

// GetAllModels returns all registered models
func (e *Engine) GetAllModels() map[string]*ModelInfo {
	e.modelsMu.RLock()
	defer e.modelsMu.RUnlock()

	models := make(map[string]*ModelInfo)
	for k, v := range e.models {
		models[k] = v
	}

	return models
}

// DeleteModel removes a model from the registry
func (e *Engine) DeleteModel(name string) error {
	e.modelsMu.Lock()
	defer e.modelsMu.Unlock()

	if _, exists := e.models[name]; !exists {
		return fmt.Errorf("model %s not found", name)
	}

	delete(e.models, name)
	return nil
}

// GetModelCount returns the number of registered models
func (e *Engine) GetModelCount() int {
	e.modelsMu.RLock()
	defer e.modelsMu.RUnlock()

	return len(e.models)
}

// GetOnlineNodeCount returns the number of online nodes
func (e *Engine) GetOnlineNodeCount() int {
	e.nodesMu.RLock()
	defer e.nodesMu.RUnlock()

	count := 0
	for _, node := range e.nodes {
		if node.Status == NodeStatusOnline {
			count++
		}
	}

	return count
}

// GetNodes returns all registered nodes
func (e *Engine) GetNodes() map[string]*NodeInfo {
	e.nodesMu.RLock()
	defer e.nodesMu.RUnlock()

	nodes := make(map[string]*NodeInfo)
	for k, v := range e.nodes {
		nodes[k] = v
	}

	return nodes
}

// GetAvailableNodes returns nodes that are online and available
func (e *Engine) GetAvailableNodes() []*NodeInfo {
	e.nodesMu.RLock()
	defer e.nodesMu.RUnlock()

	var available []*NodeInfo
	for _, node := range e.nodes {
		if node.Status == NodeStatusOnline {
			available = append(available, node)
		}
	}

	return available
}

// GetClusterSize returns the total number of nodes in the cluster
func (e *Engine) GetClusterSize() int {
	e.nodesMu.RLock()
	defer e.nodesMu.RUnlock()

	return len(e.nodes)
}

// GetActiveNodes returns the count of active (online) nodes
func (e *Engine) GetActiveNodes() int {
	e.nodesMu.RLock()
	defer e.nodesMu.RUnlock()

	count := 0
	for _, node := range e.nodes {
		if node.Status == NodeStatusOnline {
			count++
		}
	}

	return count
}

// GetStats returns current scheduler statistics
func (e *Engine) GetStats() *Stats {
	// Get current counts without holding locks together
	nodesTotal := e.GetClusterSize()
	nodesOnline := e.GetActiveNodes()
	modelsTotal := e.GetModelCount()

	e.statsMu.Lock()
	defer e.statsMu.Unlock()

	// Update current stats
	e.stats.NodesTotal = nodesTotal
	e.stats.NodesOnline = nodesOnline
	e.stats.NodesOffline = nodesTotal - nodesOnline
	e.stats.ModelsTotal = modelsTotal
	e.stats.WorkersActive = len(e.workers)
	e.stats.QueuedRequests = int64(len(e.requests))
	e.stats.Uptime = time.Since(e.startTime)
	e.stats.LastUpdated = time.Now()

	// Return a copy of the stats
	return &Stats{
		TotalRequests:     e.stats.TotalRequests,
		CompletedRequests: e.stats.CompletedRequests,
		FailedRequests:    e.stats.FailedRequests,
		QueuedRequests:    e.stats.QueuedRequests,
		AverageLatency:    e.stats.AverageLatency,
		NodesTotal:        e.stats.NodesTotal,
		NodesOnline:       e.stats.NodesOnline,
		NodesOffline:      e.stats.NodesOffline,
		ModelsTotal:       e.stats.ModelsTotal,
		WorkersActive:     e.stats.WorkersActive,
		Uptime:            e.stats.Uptime,
		LastUpdated:       e.stats.LastUpdated,
	}
}

// IsHealthy returns true if the scheduler is healthy
func (e *Engine) IsHealthy() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Check if scheduler is started
	if !e.started {
		return false
	}

	// Check if we have at least one online node
	if e.GetActiveNodes() == 0 {
		return false
	}

	// Check if workers are running
	if len(e.workers) == 0 {
		return false
	}

	// Check if request queue is not completely full
	if len(e.requests) >= cap(e.requests) {
		return false
	}

	return true
}

// Shutdown gracefully shuts down the scheduling engine
func (e *Engine) Shutdown(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.started {
		return nil
	}

	// Stop workers
	for _, worker := range e.workers {
		close(worker.stopCh)
	}

	// Stop health checker
	close(e.healthChecker.stopCh)

	// Cancel context
	e.cancel()

	e.started = false
	return nil
}

// Worker methods

// start starts the worker
func (w *Worker) start() {
	for {
		select {
		case <-w.stopCh:
			return
		case req := <-w.engine.requests:
			w.processRequest(req)
		}
	}
}

// processRequest processes a single request
func (w *Worker) processRequest(req *Request) {
	req.ScheduledAt = time.Now()

	// Find the best node for this request
	node, err := w.engine.loadBalancer.SelectNode(req)
	if err != nil {
		w.sendResponse(req, &Response{
			RequestID: req.ID,
			Success:   false,
			Error:     fmt.Sprintf("failed to select node: %v", err),
			Duration:  time.Since(req.CreatedAt),
		})
		return
	}

	// Execute the request on the selected node
	response := w.executeRequest(req, node)
	w.sendResponse(req, response)
}

// executeRequest executes a request on a specific node
func (w *Worker) executeRequest(req *Request, node *NodeInfo) *Response {
	start := time.Now()

	// Execute request via P2P communication
	ctx, cancel := context.WithTimeout(context.Background(), req.Timeout)
	defer cancel()

	// Prepare request payload
	payload := map[string]interface{}{
		"id":         req.ID,
		"model_name": req.ModelName,
		"type":       req.Type,
		"priority":   req.Priority,
		"payload":    req.Payload,
		"created_at": req.CreatedAt,
	}

	// Send request to node via P2P
	responseData, err := w.sendP2PRequest(ctx, node, payload)
	if err != nil {
		return &Response{
			RequestID: req.ID,
			NodeID:    node.ID,
			Success:   false,
			Error:     fmt.Sprintf("P2P request failed: %v", err),
			Duration:  time.Since(start),
		}
	}

	// Parse response
	if responseMap, ok := responseData.(map[string]interface{}); ok {
		if success, exists := responseMap["success"]; exists {
			if successBool, ok := success.(bool); ok && successBool {
				return &Response{
					RequestID: req.ID,
					NodeID:    node.ID,
					Success:   true,
					Data:      responseMap["data"],
					Duration:  time.Since(start),
				}
			}
		}
		
		if errorMsg, exists := responseMap["error"]; exists {
			return &Response{
				RequestID: req.ID,
				NodeID:    node.ID,
				Success:   false,
				Error:     fmt.Sprintf("%v", errorMsg),
				Duration:  time.Since(start),
			}
		}
	}

	// Successful response
	return &Response{
		RequestID: req.ID,
		NodeID:    node.ID,
		Success:   true,
		Data:      responseData,
		Duration:  time.Since(start),
	}
}

// sendResponse sends a response back to the requester
func (w *Worker) sendResponse(req *Request, response *Response) {
	req.CompletedAt = time.Now()

	// Update statistics
	w.engine.statsMu.Lock()
	w.engine.stats.TotalRequests++
	if response.Success {
		w.engine.stats.CompletedRequests++

		// Update average latency for successful requests only
		if w.engine.stats.CompletedRequests == 1 {
			w.engine.stats.AverageLatency = response.Duration
		} else {
			totalLatency := w.engine.stats.AverageLatency * time.Duration(w.engine.stats.CompletedRequests-1)
			w.engine.stats.AverageLatency = (totalLatency + response.Duration) / time.Duration(w.engine.stats.CompletedRequests)
		}
	} else {
		w.engine.stats.FailedRequests++
	}
	w.engine.statsMu.Unlock()

	select {
	case req.ResponseCh <- response:
	case <-time.After(5 * time.Second):
		// Response channel blocked or closed
	}
}

// HealthChecker methods

// start starts the health checker
func (h *HealthChecker) start() {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-h.stopCh:
			return
		case <-ticker.C:
			h.checkHealth()
		}
	}
}

// checkHealth checks the health of all nodes
func (h *HealthChecker) checkHealth() {
	nodes := h.engine.GetAvailableNodes()

	for _, node := range nodes {
		go h.checkNodeHealth(node)
	}
}

// checkNodeHealth checks the health of a specific node
func (h *HealthChecker) checkNodeHealth(node *NodeInfo) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Send health check ping
	ping := map[string]interface{}{
		"type":      "health_check",
		"timestamp": start.Unix(),
		"node_id":   h.engine.p2p.ID().String(),
	}

	// Attempt to send ping via P2P
	response, err := h.sendHealthPing(ctx, node, ping)
	
	h.engine.nodesMu.Lock()
	defer h.engine.nodesMu.Unlock()

	if err != nil {
		// Health check failed
		if time.Since(node.LastSeen) > 2*time.Minute {
			node.Status = NodeStatusOffline
		} else {
			node.Status = NodeStatusDraining
		}
		return
	}

	// Parse health response
	if healthData, ok := response.(map[string]interface{}); ok {
		// Update node capacity and usage from health response
		if capacity, exists := healthData["capacity"]; exists {
			if capacityMap, ok := capacity.(map[string]interface{}); ok {
				h.updateNodeCapacity(node, capacityMap)
			}
		}
		
		if usage, exists := healthData["usage"]; exists {
			if usageMap, ok := usage.(map[string]interface{}); ok {
				h.updateNodeUsage(node, usageMap)
			}
		}
		
		if models, exists := healthData["models"]; exists {
			if modelSlice, ok := models.([]interface{}); ok {
				node.Models = make([]string, len(modelSlice))
				for i, model := range modelSlice {
					if modelStr, ok := model.(string); ok {
						node.Models[i] = modelStr
					}
				}
			}
		}
	}

	// Health check successful
	node.Status = NodeStatusOnline
	node.LastSeen = time.Now()
}

// LoadBalancer methods

// SelectNode selects the best node for a request
func (lb *LoadBalancer) SelectNode(req *Request) (*NodeInfo, error) {
	nodes := lb.engine.GetAvailableNodes()

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no available nodes")
	}

	// Check if any nodes have the required model
	var candidateNodes []*NodeInfo
	for _, node := range nodes {
		if contains(node.Models, req.ModelName) {
			candidateNodes = append(candidateNodes, node)
		}
	}

	// If no nodes have the model, use all available nodes
	if len(candidateNodes) == 0 {
		candidateNodes = nodes
	}

	// Apply load balancing algorithm
	switch lb.algorithm {
	case "round_robin":
		return lb.roundRobin(candidateNodes)
	case "least_connections":
		return lb.leastConnections(candidateNodes)
	case "random":
		return lb.random(candidateNodes)
	default:
		return lb.roundRobin(candidateNodes)
	}
}

// roundRobin implements round-robin load balancing
func (lb *LoadBalancer) roundRobin(nodes []*NodeInfo) (*NodeInfo, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	// Get or create round-robin state
	state := lb.getRoundRobinState()
	
	// Select next node in rotation
	currentIndex := state.currentIndex
	selectedNode := nodes[currentIndex]
	
	// Update state for next request
	state.currentIndex = (currentIndex + 1) % len(nodes)
	
	return selectedNode, nil
}

// leastConnections implements least connections load balancing
func (lb *LoadBalancer) leastConnections(nodes []*NodeInfo) (*NodeInfo, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	// Find node with least connections/load
	var selectedNode *NodeInfo
	lowestLoad := float64(100) // Start with max load

	for _, node := range nodes {
		// Calculate current load based on CPU and memory usage
		currentLoad := (node.Usage.CPU + node.Usage.Memory) / 2
		
		// Prefer nodes with lower load
		if currentLoad < lowestLoad {
			lowestLoad = currentLoad
			selectedNode = node
		}
	}

	if selectedNode == nil {
		// Fallback to first node if no suitable node found
		selectedNode = nodes[0]
	}

	return selectedNode, nil
}

// random implements random load balancing
func (lb *LoadBalancer) random(nodes []*NodeInfo) (*NodeInfo, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	// Use current time as seed for randomness
	seed := time.Now().UnixNano()
	randomIndex := int(seed % int64(len(nodes)))
	
	// Ensure index is within bounds
	if randomIndex < 0 {
		randomIndex = 0
	}
	if randomIndex >= len(nodes) {
		randomIndex = len(nodes) - 1
	}

	return nodes[randomIndex], nil
}

// Helper functions

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RoundRobinState tracks round-robin load balancing state
type RoundRobinState struct {
	currentIndex int
	mu           sync.Mutex
}

var globalRoundRobinState = &RoundRobinState{}

// getRoundRobinState returns the global round-robin state
func (lb *LoadBalancer) getRoundRobinState() *RoundRobinState {
	return globalRoundRobinState
}

// sendP2PRequest sends a request to a node via P2P
func (w *Worker) sendP2PRequest(ctx context.Context, node *NodeInfo, payload map[string]interface{}) (interface{}, error) {
	// Convert node ID to peer ID
	peerID, err := peer.Decode(node.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid peer ID: %w", err)
	}

	// Check if connected to peer
	if !w.engine.p2p.IsConnected(peerID) {
		// Try to connect
		peerInfo := peer.AddrInfo{
			ID: peerID,
			// Note: In a real implementation, we'd have the multiaddrs
		}
		if err := w.engine.p2p.ConnectToPeer(ctx, peerInfo); err != nil {
			return nil, fmt.Errorf("failed to connect to peer: %w", err)
		}
	}

	// Send request via P2P stream
	// This is a simplified implementation
	// In practice, you'd use libp2p streams for communication
	
	// For now, simulate successful communication
	response := map[string]interface{}{
		"success": true,
		"data":    "processed successfully",
		"node_id": node.ID,
	}

	return response, nil
}

// sendHealthPing sends a health check ping to a node
func (h *HealthChecker) sendHealthPing(ctx context.Context, node *NodeInfo, ping map[string]interface{}) (interface{}, error) {
	// Convert node ID to peer ID
	peerID, err := peer.Decode(node.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid peer ID: %w", err)
	}

	// Check if connected to peer
	if !h.engine.p2p.IsConnected(peerID) {
		return nil, fmt.Errorf("peer not connected")
	}

	// Send health ping via P2P
	// This is a simplified implementation
	
	// Simulate health response
	response := map[string]interface{}{
		"status": "healthy",
		"capacity": map[string]interface{}{
			"cpu":    8,
			"memory": 16 * 1024 * 1024 * 1024, // 16GB
			"disk":   1024 * 1024 * 1024 * 1024, // 1TB
			"gpu":    1,
		},
		"usage": map[string]interface{}{
			"cpu":    30.5,
			"memory": 45.2,
			"disk":   25.8,
			"gpu":    0.0,
		},
		"models": []string{"llama2", "codellama"},
	}

	return response, nil
}

// updateNodeCapacity updates node capacity from health response
func (h *HealthChecker) updateNodeCapacity(node *NodeInfo, capacity map[string]interface{}) {
	if cpu, ok := capacity["cpu"].(float64); ok {
		node.Capacity.CPU = int64(cpu)
	}
	if memory, ok := capacity["memory"].(float64); ok {
		node.Capacity.Memory = int64(memory)
	}
	if disk, ok := capacity["disk"].(float64); ok {
		node.Capacity.Disk = int64(disk)
	}
	if gpu, ok := capacity["gpu"].(float64); ok {
		node.Capacity.GPU = int64(gpu)
	}
}

// updateNodeUsage updates node usage from health response
func (h *HealthChecker) updateNodeUsage(node *NodeInfo, usage map[string]interface{}) {
	if cpu, ok := usage["cpu"].(float64); ok {
		node.Usage.CPU = cpu
	}
	if memory, ok := usage["memory"].(float64); ok {
		node.Usage.Memory = memory
	}
	if disk, ok := usage["disk"].(float64); ok {
		node.Usage.Disk = disk
	}
	if gpu, ok := usage["gpu"].(float64); ok {
		node.Usage.GPU = gpu
	}
}

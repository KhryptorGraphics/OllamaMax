package protocols

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

// HealthCheckHandler handles health check requests and node capability exchanges
type HealthCheckHandler struct {
	// Node information
	nodeInfo     *NodeHealth
	capabilities *NodeCapabilities

	// Health monitoring
	healthMonitor HealthMonitor

	// Configuration
	config *HealthConfig

	// Metrics
	metrics *HealthMetrics

	// Peer health tracking
	peerHealth    map[peer.ID]*PeerHealth
	peerHealthMux sync.RWMutex
}

// HealthConfig configures health check behavior
type HealthConfig struct {
	HealthCheckInterval    time.Duration `json:"health_check_interval"`
	HealthTimeout          time.Duration `json:"health_timeout"`
	MaxFailures            int           `json:"max_failures"`
	RecoveryThreshold      int           `json:"recovery_threshold"`
	EnableMetrics          bool          `json:"enable_metrics"`
	EnableResourceMonitor  bool          `json:"enable_resource_monitor"`
	ResourceUpdateInterval time.Duration `json:"resource_update_interval"`
}

// NodeHealth represents the health status of a node
type NodeHealth struct {
	NodeID   peer.ID       `json:"node_id"`
	Status   HealthStatus  `json:"status"`
	Uptime   time.Duration `json:"uptime"`
	LastSeen time.Time     `json:"last_seen"`

	// Resource utilization
	Resources *ResourceMetrics `json:"resources"`

	// Service health
	Services map[string]ServiceHealth `json:"services"`

	// Network connectivity
	ConnectedPeers int           `json:"connected_peers"`
	NetworkLatency time.Duration `json:"network_latency"`

	// Model information
	LoadedModels    []string `json:"loaded_models"`
	AvailableModels []string `json:"available_models"`

	// Performance metrics
	RequestRate     float64       `json:"request_rate"`
	ErrorRate       float64       `json:"error_rate"`
	AvgResponseTime time.Duration `json:"avg_response_time"`
}

// HealthStatus represents the health status of a node
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// ServiceHealth represents the health of a specific service
type ServiceHealth struct {
	Name       string        `json:"name"`
	Status     HealthStatus  `json:"status"`
	LastCheck  time.Time     `json:"last_check"`
	ErrorCount int           `json:"error_count"`
	Uptime     time.Duration `json:"uptime"`
	Message    string        `json:"message,omitempty"`
}

// ResourceMetrics represents system resource utilization
type ResourceMetrics struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	NetworkRx   int64   `json:"network_rx"`
	NetworkTx   int64   `json:"network_tx"`

	// Detailed memory information
	MemoryTotal     int64 `json:"memory_total"`
	MemoryFree      int64 `json:"memory_free"`
	MemoryAvailable int64 `json:"memory_available"`

	// GPU information (if available)
	GPUCount  int     `json:"gpu_count"`
	GPUUsage  float64 `json:"gpu_usage"`
	GPUMemory int64   `json:"gpu_memory"`

	Timestamp time.Time `json:"timestamp"`
}

// NodeCapabilities represents the capabilities of a node
type NodeCapabilities struct {
	NodeID   peer.ID `json:"node_id"`
	NodeType string  `json:"node_type"`
	Version  string  `json:"version"`

	// Hardware capabilities
	CPUCores    int       `json:"cpu_cores"`
	TotalMemory int64     `json:"total_memory"`
	TotalDisk   int64     `json:"total_disk"`
	HasGPU      bool      `json:"has_gpu"`
	GPUInfo     []GPUInfo `json:"gpu_info,omitempty"`

	// Software capabilities
	SupportedModels       []string `json:"supported_models"`
	MaxConcurrentRequests int      `json:"max_concurrent_requests"`
	MaxModelSize          int64    `json:"max_model_size"`

	// Network capabilities
	MaxConnections int   `json:"max_connections"`
	BandwidthUp    int64 `json:"bandwidth_up"`
	BandwidthDown  int64 `json:"bandwidth_down"`

	// Features
	Features  []string `json:"features"`
	Protocols []string `json:"protocols"`

	// Location and metadata
	Location *GeoLocation      `json:"location,omitempty"`
	Metadata map[string]string `json:"metadata"`

	LastUpdated time.Time `json:"last_updated"`
}

// GPUInfo represents information about a GPU
type GPUInfo struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Memory       int64  `json:"memory"`
	ComputeUnits int    `json:"compute_units"`
	Driver       string `json:"driver"`
}

// GeoLocation represents geographical location information
type GeoLocation struct {
	Country   string  `json:"country"`
	Region    string  `json:"region"`
	City      string  `json:"city"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// PeerHealth tracks health information for a peer
type PeerHealth struct {
	PeerID              peer.ID           `json:"peer_id"`
	LastHealthCheck     time.Time         `json:"last_health_check"`
	ConsecutiveFailures int               `json:"consecutive_failures"`
	Health              *NodeHealth       `json:"health"`
	Capabilities        *NodeCapabilities `json:"capabilities"`
	Status              HealthStatus      `json:"status"`
	RTT                 time.Duration     `json:"rtt"`
}

// HealthMetrics tracks health check metrics
type HealthMetrics struct {
	TotalChecks      int64         `json:"total_checks"`
	SuccessfulChecks int64         `json:"successful_checks"`
	FailedChecks     int64         `json:"failed_checks"`
	AverageRTT       time.Duration `json:"average_rtt"`
	PeersMonitored   int           `json:"peers_monitored"`
	UnhealthyPeers   int           `json:"unhealthy_peers"`
	LastCheck        time.Time     `json:"last_check"`

	// Per-peer metrics
	PeerMetrics map[peer.ID]*PeerHealthMetrics `json:"peer_metrics"`

	mu sync.RWMutex
}

// PeerHealthMetrics tracks metrics for individual peers
type PeerHealthMetrics struct {
	CheckCount          int64         `json:"check_count"`
	SuccessCount        int64         `json:"success_count"`
	FailureCount        int64         `json:"failure_count"`
	AverageRTT          time.Duration `json:"average_rtt"`
	LastCheck           time.Time     `json:"last_check"`
	ConsecutiveFailures int           `json:"consecutive_failures"`
}

// HealthMonitor defines the interface for monitoring node health
type HealthMonitor interface {
	GetNodeHealth() *NodeHealth
	GetNodeCapabilities() *NodeCapabilities
	GetResourceMetrics() *ResourceMetrics
	IsServiceHealthy(serviceName string) bool
	GetServiceHealth(serviceName string) *ServiceHealth
	UpdateResourceMetrics() error
}

// NewHealthCheckHandler creates a new health check handler
func NewHealthCheckHandler(monitor HealthMonitor, config *HealthConfig) *HealthCheckHandler {
	if config == nil {
		config = DefaultHealthConfig()
	}

	return &HealthCheckHandler{
		healthMonitor: monitor,
		config:        config,
		metrics: &HealthMetrics{
			PeerMetrics: make(map[peer.ID]*PeerHealthMetrics),
		},
		peerHealth: make(map[peer.ID]*PeerHealth),
	}
}

// HandleMessage handles health check protocol messages
func (hh *HealthCheckHandler) HandleMessage(ctx context.Context, stream network.Stream, msg *Message) error {
	switch msg.Type {
	case MsgTypeHealthPing:
		return hh.handleHealthPing(ctx, stream, msg)
	case MsgTypeCapabilitiesRequest:
		return hh.handleCapabilitiesRequest(ctx, stream, msg)
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// handleHealthPing handles incoming health ping requests
func (hh *HealthCheckHandler) handleHealthPing(ctx context.Context, stream network.Stream, msg *Message) error {
	start := time.Now()
	peerID := stream.Conn().RemotePeer()

	// Get current health status
	health := hh.healthMonitor.GetNodeHealth()
	if health == nil {
		return hh.sendErrorResponse(stream, msg.ID, "health_unavailable", "Node health information unavailable")
	}

	// Create health pong response
	response := &Message{
		Type:      MsgTypeHealthPong,
		ID:        generateMessageID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"request_id":        msg.ID,
			"node_id":           health.NodeID.String(),
			"status":            string(health.Status),
			"uptime":            health.Uptime.String(),
			"resources":         health.Resources,
			"services":          health.Services,
			"connected_peers":   health.ConnectedPeers,
			"network_latency":   health.NetworkLatency.String(),
			"loaded_models":     health.LoadedModels,
			"available_models":  health.AvailableModels,
			"request_rate":      health.RequestRate,
			"error_rate":        health.ErrorRate,
			"avg_response_time": health.AvgResponseTime.String(),
			"timestamp":         time.Now(),
		},
	}

	// Send response
	handler := NewProtocolHandler(HealthCheckProtocol)
	if err := handler.SendMessage(stream, response); err != nil {
		hh.updateFailureMetrics(peerID)
		return fmt.Errorf("failed to send health pong: %w", err)
	}

	// Update metrics
	rtt := time.Since(start)
	hh.updateSuccessMetrics(peerID, rtt)

	log.Printf("Responded to health ping from peer %s (RTT: %v)", peerID, rtt)
	return nil
}

// handleCapabilitiesRequest handles capabilities requests
func (hh *HealthCheckHandler) handleCapabilitiesRequest(ctx context.Context, stream network.Stream, msg *Message) error {
	peerID := stream.Conn().RemotePeer()

	// Get current capabilities
	capabilities := hh.healthMonitor.GetNodeCapabilities()
	if capabilities == nil {
		return hh.sendErrorResponse(stream, msg.ID, "capabilities_unavailable", "Node capabilities unavailable")
	}

	// Create capabilities response
	response := &Message{
		Type:      MsgTypeCapabilitiesResponse,
		ID:        generateMessageID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"request_id":              msg.ID,
			"node_id":                 capabilities.NodeID.String(),
			"node_type":               capabilities.NodeType,
			"version":                 capabilities.Version,
			"cpu_cores":               capabilities.CPUCores,
			"total_memory":            capabilities.TotalMemory,
			"total_disk":              capabilities.TotalDisk,
			"has_gpu":                 capabilities.HasGPU,
			"gpu_info":                capabilities.GPUInfo,
			"supported_models":        capabilities.SupportedModels,
			"max_concurrent_requests": capabilities.MaxConcurrentRequests,
			"max_model_size":          capabilities.MaxModelSize,
			"max_connections":         capabilities.MaxConnections,
			"bandwidth_up":            capabilities.BandwidthUp,
			"bandwidth_down":          capabilities.BandwidthDown,
			"features":                capabilities.Features,
			"protocols":               capabilities.Protocols,
			"location":                capabilities.Location,
			"metadata":                capabilities.Metadata,
			"last_updated":            capabilities.LastUpdated,
		},
	}

	// Send response
	handler := NewProtocolHandler(HealthCheckProtocol)
	if err := handler.SendMessage(stream, response); err != nil {
		return fmt.Errorf("failed to send capabilities response: %w", err)
	}

	log.Printf("Sent capabilities to peer %s", peerID)
	return nil
}

// sendErrorResponse sends an error response
func (hh *HealthCheckHandler) sendErrorResponse(stream network.Stream, requestID, errorCode, errorMessage string) error {
	errorMsg := &Message{
		Type:      "error",
		ID:        generateMessageID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"request_id":    requestID,
			"error_code":    errorCode,
			"error_message": errorMessage,
		},
	}

	handler := NewProtocolHandler(HealthCheckProtocol)
	return handler.SendMessage(stream, errorMsg)
}

// UpdatePeerHealth updates health information for a peer
func (hh *HealthCheckHandler) UpdatePeerHealth(peerID peer.ID, health *NodeHealth, rtt time.Duration) {
	hh.peerHealthMux.Lock()
	defer hh.peerHealthMux.Unlock()

	if hh.peerHealth[peerID] == nil {
		hh.peerHealth[peerID] = &PeerHealth{
			PeerID: peerID,
		}
	}

	peerHealth := hh.peerHealth[peerID]
	peerHealth.LastHealthCheck = time.Now()
	peerHealth.Health = health
	peerHealth.RTT = rtt

	// Determine status based on health
	if health != nil {
		peerHealth.Status = health.Status
		peerHealth.ConsecutiveFailures = 0
	} else {
		peerHealth.ConsecutiveFailures++
		if peerHealth.ConsecutiveFailures >= hh.config.MaxFailures {
			peerHealth.Status = HealthStatusUnhealthy
		} else {
			peerHealth.Status = HealthStatusDegraded
		}
	}
}

// GetPeerHealth returns health information for a peer
func (hh *HealthCheckHandler) GetPeerHealth(peerID peer.ID) (*PeerHealth, bool) {
	hh.peerHealthMux.RLock()
	defer hh.peerHealthMux.RUnlock()

	health, exists := hh.peerHealth[peerID]
	return health, exists
}

// GetAllPeerHealth returns health information for all monitored peers
func (hh *HealthCheckHandler) GetAllPeerHealth() map[peer.ID]*PeerHealth {
	hh.peerHealthMux.RLock()
	defer hh.peerHealthMux.RUnlock()

	result := make(map[peer.ID]*PeerHealth)
	for peerID, health := range hh.peerHealth {
		// Create copy
		healthCopy := *health
		result[peerID] = &healthCopy
	}

	return result
}

// GetHealthyPeers returns a list of healthy peers
func (hh *HealthCheckHandler) GetHealthyPeers() []peer.ID {
	hh.peerHealthMux.RLock()
	defer hh.peerHealthMux.RUnlock()

	var healthyPeers []peer.ID
	for peerID, health := range hh.peerHealth {
		if health.Status == HealthStatusHealthy {
			healthyPeers = append(healthyPeers, peerID)
		}
	}

	return healthyPeers
}

// CleanupStaleEntries removes old peer health entries
func (hh *HealthCheckHandler) CleanupStaleEntries(maxAge time.Duration) {
	hh.peerHealthMux.Lock()
	defer hh.peerHealthMux.Unlock()

	cutoff := time.Now().Add(-maxAge)
	for peerID, health := range hh.peerHealth {
		if health.LastHealthCheck.Before(cutoff) {
			delete(hh.peerHealth, peerID)
			log.Printf("Removed stale health entry for peer %s", peerID)
		}
	}
}

// Metrics update methods

func (hh *HealthCheckHandler) updateSuccessMetrics(peerID peer.ID, rtt time.Duration) {
	hh.metrics.mu.Lock()
	defer hh.metrics.mu.Unlock()

	hh.metrics.TotalChecks++
	hh.metrics.SuccessfulChecks++
	hh.metrics.LastCheck = time.Now()

	// Update average RTT
	if hh.metrics.SuccessfulChecks > 0 {
		totalRTT := hh.metrics.AverageRTT * time.Duration(hh.metrics.SuccessfulChecks-1)
		hh.metrics.AverageRTT = (totalRTT + rtt) / time.Duration(hh.metrics.SuccessfulChecks)
	} else {
		hh.metrics.AverageRTT = rtt
	}

	// Update peer-specific metrics
	if hh.metrics.PeerMetrics[peerID] == nil {
		hh.metrics.PeerMetrics[peerID] = &PeerHealthMetrics{}
	}

	peerMetrics := hh.metrics.PeerMetrics[peerID]
	peerMetrics.CheckCount++
	peerMetrics.SuccessCount++
	peerMetrics.LastCheck = time.Now()
	peerMetrics.ConsecutiveFailures = 0

	// Update peer average RTT
	if peerMetrics.SuccessCount > 0 {
		totalRTT := peerMetrics.AverageRTT * time.Duration(peerMetrics.SuccessCount-1)
		peerMetrics.AverageRTT = (totalRTT + rtt) / time.Duration(peerMetrics.SuccessCount)
	} else {
		peerMetrics.AverageRTT = rtt
	}
}

func (hh *HealthCheckHandler) updateFailureMetrics(peerID peer.ID) {
	hh.metrics.mu.Lock()
	defer hh.metrics.mu.Unlock()

	hh.metrics.TotalChecks++
	hh.metrics.FailedChecks++
	hh.metrics.LastCheck = time.Now()

	// Update peer-specific metrics
	if hh.metrics.PeerMetrics[peerID] == nil {
		hh.metrics.PeerMetrics[peerID] = &PeerHealthMetrics{}
	}

	peerMetrics := hh.metrics.PeerMetrics[peerID]
	peerMetrics.CheckCount++
	peerMetrics.FailureCount++
	peerMetrics.ConsecutiveFailures++
	peerMetrics.LastCheck = time.Now()
}

// GetMetrics returns a copy of current metrics
func (hh *HealthCheckHandler) GetMetrics() *HealthMetrics {
	hh.metrics.mu.RLock()
	defer hh.metrics.mu.RUnlock()

	// Count current peer status
	hh.peerHealthMux.RLock()
	peersMonitored := len(hh.peerHealth)
	unhealthyPeers := 0
	for _, health := range hh.peerHealth {
		if health.Status == HealthStatusUnhealthy {
			unhealthyPeers++
		}
	}
	hh.peerHealthMux.RUnlock()

	// Create deep copy
	metricsCopy := &HealthMetrics{
		TotalChecks:      hh.metrics.TotalChecks,
		SuccessfulChecks: hh.metrics.SuccessfulChecks,
		FailedChecks:     hh.metrics.FailedChecks,
		AverageRTT:       hh.metrics.AverageRTT,
		PeersMonitored:   peersMonitored,
		UnhealthyPeers:   unhealthyPeers,
		LastCheck:        hh.metrics.LastCheck,
		PeerMetrics:      make(map[peer.ID]*PeerHealthMetrics),
	}

	// Copy peer metrics
	for peerID, peerMetrics := range hh.metrics.PeerMetrics {
		metricsCopy.PeerMetrics[peerID] = &PeerHealthMetrics{
			CheckCount:          peerMetrics.CheckCount,
			SuccessCount:        peerMetrics.SuccessCount,
			FailureCount:        peerMetrics.FailureCount,
			AverageRTT:          peerMetrics.AverageRTT,
			LastCheck:           peerMetrics.LastCheck,
			ConsecutiveFailures: peerMetrics.ConsecutiveFailures,
		}
	}

	return metricsCopy
}

// DefaultHealthConfig returns default health check configuration
func DefaultHealthConfig() *HealthConfig {
	return &HealthConfig{
		HealthCheckInterval:    30 * time.Second,
		HealthTimeout:          10 * time.Second,
		MaxFailures:            3,
		RecoveryThreshold:      2,
		EnableMetrics:          true,
		EnableResourceMonitor:  true,
		ResourceUpdateInterval: 15 * time.Second,
	}
}

// HealthClient provides client-side health check functionality
type HealthClient struct {
	protocolClient *ProtocolClient
}

// NewHealthClient creates a new health check client
func NewHealthClient(dialer StreamDialer, timeout time.Duration) *HealthClient {
	return &HealthClient{
		protocolClient: NewProtocolClient(dialer, HealthCheckProtocol, timeout),
	}
}

// PingPeer sends a health ping to a peer
func (hc *HealthClient) PingPeer(ctx context.Context, peerID peer.ID) (*NodeHealth, time.Duration, error) {
	start := time.Now()

	// Create ping message
	pingMsg := CreateRequestMessage(MsgTypeHealthPing, map[string]interface{}{
		"timestamp": start,
		"ping_id":   generateMessageID(),
	})

	// Send ping and wait for pong
	pongMsg, err := hc.protocolClient.SendRequest(ctx, peerID, pingMsg)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to send health ping: %w", err)
	}

	rtt := time.Since(start)

	// Handle error response
	if pongMsg.Type == "error" {
		errorCode, _ := pongMsg.Data["error_code"].(string)
		errorMessage, _ := pongMsg.Data["error_message"].(string)
		return nil, rtt, fmt.Errorf("health ping error [%s]: %s", errorCode, errorMessage)
	}

	// Parse health response
	health, err := hc.parseHealthResponse(pongMsg)
	if err != nil {
		return nil, rtt, fmt.Errorf("failed to parse health response: %w", err)
	}

	return health, rtt, nil
}

// GetCapabilities requests capabilities from a peer
func (hc *HealthClient) GetCapabilities(ctx context.Context, peerID peer.ID) (*NodeCapabilities, error) {
	// Create capabilities request
	reqMsg := CreateRequestMessage(MsgTypeCapabilitiesRequest, map[string]interface{}{
		"timestamp": time.Now(),
	})

	// Send request and wait for response
	respMsg, err := hc.protocolClient.SendRequest(ctx, peerID, reqMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to send capabilities request: %w", err)
	}

	// Handle error response
	if respMsg.Type == "error" {
		errorCode, _ := respMsg.Data["error_code"].(string)
		errorMessage, _ := respMsg.Data["error_message"].(string)
		return nil, fmt.Errorf("capabilities request error [%s]: %s", errorCode, errorMessage)
	}

	// Parse capabilities response
	return hc.parseCapabilitiesResponse(respMsg)
}

// parseHealthResponse parses a health pong response
func (hc *HealthClient) parseHealthResponse(msg *Message) (*NodeHealth, error) {
	data := msg.Data

	health := &NodeHealth{
		Services: make(map[string]ServiceHealth),
	}

	if nodeIDStr, ok := data["node_id"].(string); ok {
		if nodeID, err := peer.Decode(nodeIDStr); err == nil {
			health.NodeID = nodeID
		}
	}

	if statusStr, ok := data["status"].(string); ok {
		health.Status = HealthStatus(statusStr)
	}

	if uptimeStr, ok := data["uptime"].(string); ok {
		if uptime, err := time.ParseDuration(uptimeStr); err == nil {
			health.Uptime = uptime
		}
	}

	if lastSeenStr, ok := data["timestamp"].(string); ok {
		if lastSeen, err := time.Parse(time.RFC3339, lastSeenStr); err == nil {
			health.LastSeen = lastSeen
		}
	}

	// Parse resources
	if resourcesData, ok := data["resources"]; ok {
		if resourcesMap, ok := resourcesData.(map[string]interface{}); ok {
			health.Resources = parseResourceMetrics(resourcesMap)
		}
	}

	// Parse simple fields
	if connectedPeers, ok := data["connected_peers"].(float64); ok {
		health.ConnectedPeers = int(connectedPeers)
	}

	if networkLatencyStr, ok := data["network_latency"].(string); ok {
		if latency, err := time.ParseDuration(networkLatencyStr); err == nil {
			health.NetworkLatency = latency
		}
	}

	if loadedModels, ok := data["loaded_models"].([]interface{}); ok {
		health.LoadedModels = make([]string, len(loadedModels))
		for i, model := range loadedModels {
			if modelStr, ok := model.(string); ok {
				health.LoadedModels[i] = modelStr
			}
		}
	}

	if availableModels, ok := data["available_models"].([]interface{}); ok {
		health.AvailableModels = make([]string, len(availableModels))
		for i, model := range availableModels {
			if modelStr, ok := model.(string); ok {
				health.AvailableModels[i] = modelStr
			}
		}
	}

	if requestRate, ok := data["request_rate"].(float64); ok {
		health.RequestRate = requestRate
	}

	if errorRate, ok := data["error_rate"].(float64); ok {
		health.ErrorRate = errorRate
	}

	if avgResponseTimeStr, ok := data["avg_response_time"].(string); ok {
		if avgResponseTime, err := time.ParseDuration(avgResponseTimeStr); err == nil {
			health.AvgResponseTime = avgResponseTime
		}
	}

	return health, nil
}

// parseCapabilitiesResponse parses a capabilities response
func (hc *HealthClient) parseCapabilitiesResponse(msg *Message) (*NodeCapabilities, error) {
	data := msg.Data

	capabilities := &NodeCapabilities{
		Metadata: make(map[string]string),
	}

	if nodeIDStr, ok := data["node_id"].(string); ok {
		if nodeID, err := peer.Decode(nodeIDStr); err == nil {
			capabilities.NodeID = nodeID
		}
	}

	if nodeType, ok := data["node_type"].(string); ok {
		capabilities.NodeType = nodeType
	}

	if version, ok := data["version"].(string); ok {
		capabilities.Version = version
	}

	if cpuCores, ok := data["cpu_cores"].(float64); ok {
		capabilities.CPUCores = int(cpuCores)
	}

	if totalMemory, ok := data["total_memory"].(float64); ok {
		capabilities.TotalMemory = int64(totalMemory)
	}

	if totalDisk, ok := data["total_disk"].(float64); ok {
		capabilities.TotalDisk = int64(totalDisk)
	}

	if hasGPU, ok := data["has_gpu"].(bool); ok {
		capabilities.HasGPU = hasGPU
	}

	// Parse supported models
	if supportedModels, ok := data["supported_models"].([]interface{}); ok {
		capabilities.SupportedModels = make([]string, len(supportedModels))
		for i, model := range supportedModels {
			if modelStr, ok := model.(string); ok {
				capabilities.SupportedModels[i] = modelStr
			}
		}
	}

	if maxConcurrentRequests, ok := data["max_concurrent_requests"].(float64); ok {
		capabilities.MaxConcurrentRequests = int(maxConcurrentRequests)
	}

	if maxModelSize, ok := data["max_model_size"].(float64); ok {
		capabilities.MaxModelSize = int64(maxModelSize)
	}

	// Parse features
	if features, ok := data["features"].([]interface{}); ok {
		capabilities.Features = make([]string, len(features))
		for i, feature := range features {
			if featureStr, ok := feature.(string); ok {
				capabilities.Features[i] = featureStr
			}
		}
	}

	// Parse protocols
	if protocols, ok := data["protocols"].([]interface{}); ok {
		capabilities.Protocols = make([]string, len(protocols))
		for i, protocol := range protocols {
			if protocolStr, ok := protocol.(string); ok {
				capabilities.Protocols[i] = protocolStr
			}
		}
	}

	if lastUpdatedStr, ok := data["last_updated"].(string); ok {
		if lastUpdated, err := time.Parse(time.RFC3339, lastUpdatedStr); err == nil {
			capabilities.LastUpdated = lastUpdated
		}
	}

	return capabilities, nil
}

// parseResourceMetrics parses resource metrics from message data
func parseResourceMetrics(data map[string]interface{}) *ResourceMetrics {
	resources := &ResourceMetrics{}

	if cpuUsage, ok := data["cpu_usage"].(float64); ok {
		resources.CPUUsage = cpuUsage
	}

	if memoryUsage, ok := data["memory_usage"].(float64); ok {
		resources.MemoryUsage = memoryUsage
	}

	if diskUsage, ok := data["disk_usage"].(float64); ok {
		resources.DiskUsage = diskUsage
	}

	if networkRx, ok := data["network_rx"].(float64); ok {
		resources.NetworkRx = int64(networkRx)
	}

	if networkTx, ok := data["network_tx"].(float64); ok {
		resources.NetworkTx = int64(networkTx)
	}

	if memoryTotal, ok := data["memory_total"].(float64); ok {
		resources.MemoryTotal = int64(memoryTotal)
	}

	if memoryFree, ok := data["memory_free"].(float64); ok {
		resources.MemoryFree = int64(memoryFree)
	}

	if memoryAvailable, ok := data["memory_available"].(float64); ok {
		resources.MemoryAvailable = int64(memoryAvailable)
	}

	if gpuCount, ok := data["gpu_count"].(float64); ok {
		resources.GPUCount = int(gpuCount)
	}

	if gpuUsage, ok := data["gpu_usage"].(float64); ok {
		resources.GPUUsage = gpuUsage
	}

	if gpuMemory, ok := data["gpu_memory"].(float64); ok {
		resources.GPUMemory = int64(gpuMemory)
	}

	if timestampStr, ok := data["timestamp"].(string); ok {
		if timestamp, err := time.Parse(time.RFC3339, timestampStr); err == nil {
			resources.Timestamp = timestamp
		}
	}

	return resources
}

// GetClientMetrics returns client metrics
func (hc *HealthClient) GetClientMetrics() *ClientMetrics {
	return hc.protocolClient.GetClientMetrics()
}

// BasicHealthMonitor provides a basic implementation of HealthMonitor
type BasicHealthMonitor struct {
	nodeID      peer.ID
	startTime   time.Time
	services    map[string]*ServiceHealth
	servicesMux sync.RWMutex
}

// NewBasicHealthMonitor creates a new basic health monitor
func NewBasicHealthMonitor(nodeID peer.ID) *BasicHealthMonitor {
	return &BasicHealthMonitor{
		nodeID:    nodeID,
		startTime: time.Now(),
		services:  make(map[string]*ServiceHealth),
	}
}

// GetNodeHealth returns current node health
func (bhm *BasicHealthMonitor) GetNodeHealth() *NodeHealth {
	resources := bhm.GetResourceMetrics()

	// Determine overall status based on resource usage
	status := HealthStatusHealthy
	if resources.CPUUsage > 90 || resources.MemoryUsage > 90 {
		status = HealthStatusDegraded
	}
	if resources.CPUUsage > 95 || resources.MemoryUsage > 95 {
		status = HealthStatusUnhealthy
	}

	bhm.servicesMux.RLock()
	services := make(map[string]ServiceHealth)
	for name, service := range bhm.services {
		services[name] = *service
	}
	bhm.servicesMux.RUnlock()

	return &NodeHealth{
		NodeID:          bhm.nodeID,
		Status:          status,
		Uptime:          time.Since(bhm.startTime),
		LastSeen:        time.Now(),
		Resources:       resources,
		Services:        services,
		ConnectedPeers:  0, // Would be updated by network layer
		LoadedModels:    []string{},
		AvailableModels: []string{},
	}
}

// GetNodeCapabilities returns node capabilities
func (bhm *BasicHealthMonitor) GetNodeCapabilities() *NodeCapabilities {
	return &NodeCapabilities{
		NodeID:                bhm.nodeID,
		NodeType:              "ollama-node",
		Version:               "1.0.0",
		CPUCores:              runtime.NumCPU(),
		TotalMemory:           8 * 1024 * 1024 * 1024,   // 8GB default
		TotalDisk:             100 * 1024 * 1024 * 1024, // 100GB default
		HasGPU:                false,
		SupportedModels:       []string{},
		MaxConcurrentRequests: 10,
		MaxModelSize:          10 * 1024 * 1024 * 1024, // 10GB
		MaxConnections:        100,
		Features:              []string{"inference", "model-loading"},
		Protocols:             []string{"inference", "health", "file-transfer"},
		Metadata:              make(map[string]string),
		LastUpdated:           time.Now(),
	}
}

// GetResourceMetrics returns current resource metrics
func (bhm *BasicHealthMonitor) GetResourceMetrics() *ResourceMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &ResourceMetrics{
		CPUUsage:        float64(runtime.NumGoroutine()) / float64(runtime.NumCPU()) * 10, // Rough estimate
		MemoryUsage:     float64(m.Alloc) / float64(8*1024*1024*1024) * 100,               // % of 8GB
		DiskUsage:       20.0,                                                             // 20% default
		NetworkRx:       1024 * 1024,                                                      // 1MB/s default
		NetworkTx:       1024 * 1024,                                                      // 1MB/s default
		MemoryTotal:     8 * 1024 * 1024 * 1024,                                           // 8GB
		MemoryFree:      int64(8*1024*1024*1024) - int64(m.Alloc),
		MemoryAvailable: int64(8*1024*1024*1024) - int64(m.Alloc),
		GPUCount:        0,
		GPUUsage:        0,
		GPUMemory:       0,
		Timestamp:       time.Now(),
	}
}

// IsServiceHealthy checks if a service is healthy
func (bhm *BasicHealthMonitor) IsServiceHealthy(serviceName string) bool {
	bhm.servicesMux.RLock()
	defer bhm.servicesMux.RUnlock()

	if service, exists := bhm.services[serviceName]; exists {
		return service.Status == HealthStatusHealthy
	}
	return false
}

// GetServiceHealth returns service health information
func (bhm *BasicHealthMonitor) GetServiceHealth(serviceName string) *ServiceHealth {
	bhm.servicesMux.RLock()
	defer bhm.servicesMux.RUnlock()

	if service, exists := bhm.services[serviceName]; exists {
		serviceCopy := *service
		return &serviceCopy
	}
	return nil
}

// UpdateResourceMetrics updates resource metrics (no-op for basic monitor)
func (bhm *BasicHealthMonitor) UpdateResourceMetrics() error {
	return nil
}

// RegisterService registers a service for health monitoring
func (bhm *BasicHealthMonitor) RegisterService(name string, status HealthStatus) {
	bhm.servicesMux.Lock()
	defer bhm.servicesMux.Unlock()

	bhm.services[name] = &ServiceHealth{
		Name:      name,
		Status:    status,
		LastCheck: time.Now(),
		Uptime:    time.Since(bhm.startTime),
	}
}

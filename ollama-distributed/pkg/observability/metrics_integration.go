package observability

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// MetricsIntegration provides a unified interface for components to report metrics
type MetricsIntegration struct {
	registry *MetricsRegistry

	// Component-specific integrators
	schedulerIntegrator      *SchedulerIntegrator
	consensusIntegrator      *ConsensusIntegrator
	p2pIntegrator            *P2PIntegrator
	apiIntegrator            *APIIntegrator
	faultToleranceIntegrator *FaultToleranceIntegrator
	modelIntegrator          *ModelIntegrator

	// Lifecycle
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	started bool
	mu      sync.RWMutex
}

// ComponentMetricsReporter interface for components to report metrics
type ComponentMetricsReporter interface {
	ReportMetrics(integration *MetricsIntegration)
	GetComponentName() string
}

// SchedulerIntegrator integrates scheduler metrics with Prometheus
type SchedulerIntegrator struct {
	metrics *SchedulerMetrics
	nodeID  string
}

// ConsensusIntegrator integrates consensus metrics with Prometheus
type ConsensusIntegrator struct {
	metrics *ConsensusMetrics
	nodeID  string
}

// P2PIntegrator integrates P2P metrics with Prometheus
type P2PIntegrator struct {
	metrics *P2PMetrics
	nodeID  string
}

// APIIntegrator integrates API gateway metrics with Prometheus
type APIIntegrator struct {
	metrics *APIMetrics
	nodeID  string
}

// FaultToleranceIntegrator integrates fault tolerance metrics with Prometheus
type FaultToleranceIntegrator struct {
	metrics *FaultToleranceMetrics
	nodeID  string
}

// ModelIntegrator integrates model management metrics with Prometheus
type ModelIntegrator struct {
	metrics *ModelMetrics
	nodeID  string
}

// NewMetricsIntegration creates a new metrics integration
func NewMetricsIntegration(registry *MetricsRegistry, nodeID string) *MetricsIntegration {
	ctx, cancel := context.WithCancel(context.Background())

	mi := &MetricsIntegration{
		registry: registry,
		ctx:      ctx,
		cancel:   cancel,
	}

	// Initialize component integrators
	mi.schedulerIntegrator = &SchedulerIntegrator{
		metrics: registry.GetSchedulerMetrics(),
		nodeID:  nodeID,
	}

	mi.consensusIntegrator = &ConsensusIntegrator{
		metrics: registry.GetConsensusMetrics(),
		nodeID:  nodeID,
	}

	mi.p2pIntegrator = &P2PIntegrator{
		metrics: registry.GetP2PMetrics(),
		nodeID:  nodeID,
	}

	mi.apiIntegrator = &APIIntegrator{
		metrics: registry.GetAPIMetrics(),
		nodeID:  nodeID,
	}

	mi.faultToleranceIntegrator = &FaultToleranceIntegrator{
		metrics: registry.GetFaultToleranceMetrics(),
		nodeID:  nodeID,
	}

	mi.modelIntegrator = &ModelIntegrator{
		metrics: registry.GetModelMetrics(),
		nodeID:  nodeID,
	}

	return mi
}

// Start starts the metrics integration
func (mi *MetricsIntegration) Start() error {
	mi.mu.Lock()
	defer mi.mu.Unlock()

	if mi.started {
		return nil
	}

	// Start metrics collection loop
	mi.wg.Add(1)
	go mi.metricsCollectionLoop()

	mi.started = true
	log.Info().Msg("Metrics integration started")
	return nil
}

// Stop stops the metrics integration
func (mi *MetricsIntegration) Stop() error {
	mi.mu.Lock()
	defer mi.mu.Unlock()

	if !mi.started {
		return nil
	}

	mi.cancel()
	mi.wg.Wait()

	mi.started = false
	log.Info().Msg("Metrics integration stopped")
	return nil
}

// GetSchedulerIntegrator returns the scheduler integrator
func (mi *MetricsIntegration) GetSchedulerIntegrator() *SchedulerIntegrator {
	return mi.schedulerIntegrator
}

// GetConsensusIntegrator returns the consensus integrator
func (mi *MetricsIntegration) GetConsensusIntegrator() *ConsensusIntegrator {
	return mi.consensusIntegrator
}

// GetP2PIntegrator returns the P2P integrator
func (mi *MetricsIntegration) GetP2PIntegrator() *P2PIntegrator {
	return mi.p2pIntegrator
}

// GetAPIIntegrator returns the API integrator
func (mi *MetricsIntegration) GetAPIIntegrator() *APIIntegrator {
	return mi.apiIntegrator
}

// GetFaultToleranceIntegrator returns the fault tolerance integrator
func (mi *MetricsIntegration) GetFaultToleranceIntegrator() *FaultToleranceIntegrator {
	return mi.faultToleranceIntegrator
}

// GetModelIntegrator returns the model integrator
func (mi *MetricsIntegration) GetModelIntegrator() *ModelIntegrator {
	return mi.modelIntegrator
}

// metricsCollectionLoop runs the metrics collection loop
func (mi *MetricsIntegration) metricsCollectionLoop() {
	defer mi.wg.Done()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-mi.ctx.Done():
			return
		case <-ticker.C:
			// Metrics are reported by components when they update
			// This loop can be used for periodic health checks or cleanup
		}
	}
}

// Scheduler Integration Methods

// ReportTaskScheduled reports a task scheduling event
func (si *SchedulerIntegrator) ReportTaskScheduled(taskType, status string) {
	si.metrics.TasksTotal.WithLabelValues(status, si.nodeID, taskType).Inc()
}

// ReportTaskActive reports active task count
func (si *SchedulerIntegrator) ReportTaskActive(taskType string, count float64) {
	si.metrics.TasksActive.WithLabelValues(si.nodeID, taskType).Set(count)
}

// ReportTaskDuration reports task execution duration
func (si *SchedulerIntegrator) ReportTaskDuration(taskType string, duration time.Duration) {
	si.metrics.TaskDuration.WithLabelValues(taskType, si.nodeID).Observe(duration.Seconds())
}

// ReportTaskError reports a task error
func (si *SchedulerIntegrator) ReportTaskError(taskType, errorType string) {
	si.metrics.TaskErrors.WithLabelValues(errorType, si.nodeID, taskType).Inc()
}

// ReportLoadBalancerRequest reports a load balancer request
func (si *SchedulerIntegrator) ReportLoadBalancerRequest(strategy string) {
	si.metrics.LoadBalancerRequests.WithLabelValues(strategy, si.nodeID).Inc()
}

// ReportNodeUtilization reports node resource utilization
func (si *SchedulerIntegrator) ReportNodeUtilization(resourceType string, utilization float64) {
	si.metrics.NodeUtilization.WithLabelValues(si.nodeID, resourceType).Set(utilization)
}

// Consensus Integration Methods

// ReportLeaderElection reports a leader election event
func (ci *ConsensusIntegrator) ReportLeaderElection(result string) {
	ci.metrics.LeaderElections.WithLabelValues(ci.nodeID, result).Inc()
}

// ReportLogEntry reports a log entry event
func (ci *ConsensusIntegrator) ReportLogEntry(entryType string, count float64) {
	ci.metrics.LogEntries.WithLabelValues(ci.nodeID, entryType).Add(count)
}

// ReportCommitLatency reports consensus commit latency
func (ci *ConsensusIntegrator) ReportCommitLatency(latency time.Duration) {
	ci.metrics.CommitLatency.WithLabelValues(ci.nodeID).Observe(latency.Seconds())
}

// ReportQuorumStatus reports quorum status
func (ci *ConsensusIntegrator) ReportQuorumStatus(clusterID string, hasQuorum bool) {
	status := 0.0
	if hasQuorum {
		status = 1.0
	}
	ci.metrics.QuorumStatus.WithLabelValues(clusterID).Set(status)
}

// ReportNodeState reports consensus node state
func (ci *ConsensusIntegrator) ReportNodeState(state string) {
	var stateValue float64
	switch state {
	case "follower":
		stateValue = 0
	case "candidate":
		stateValue = 1
	case "leader":
		stateValue = 2
	default:
		stateValue = -1
	}
	ci.metrics.NodeStates.WithLabelValues(ci.nodeID).Set(stateValue)
}

// ReportConsensusError reports a consensus error
func (ci *ConsensusIntegrator) ReportConsensusError(errorType string) {
	ci.metrics.ConsensusErrors.WithLabelValues(errorType, ci.nodeID).Inc()
}

// P2P Integration Methods

// ReportConnection reports a P2P connection event
func (pi *P2PIntegrator) ReportConnection(direction, peerID string) {
	pi.metrics.ConnectionsTotal.WithLabelValues(direction, peerID).Inc()
}

// ReportActiveConnections reports active connection count
func (pi *P2PIntegrator) ReportActiveConnections(protocol string, count float64) {
	pi.metrics.ConnectionsActive.WithLabelValues(protocol).Set(count)
}

// ReportMessageSent reports a sent message
func (pi *P2PIntegrator) ReportMessageSent(messageType, peerID string) {
	pi.metrics.MessagesSent.WithLabelValues(messageType, peerID).Inc()
}

// ReportMessageReceived reports a received message
func (pi *P2PIntegrator) ReportMessageReceived(messageType, peerID string) {
	pi.metrics.MessagesReceived.WithLabelValues(messageType, peerID).Inc()
}

// ReportNetworkLatency reports network latency
func (pi *P2PIntegrator) ReportNetworkLatency(peerID string, latency time.Duration) {
	pi.metrics.NetworkLatency.WithLabelValues(peerID).Observe(latency.Seconds())
}

// ReportBandwidthUsage reports bandwidth usage
func (pi *P2PIntegrator) ReportBandwidthUsage(direction, peerID string, bytesPerSecond float64) {
	pi.metrics.BandwidthUsage.WithLabelValues(direction, peerID).Set(bytesPerSecond)
}

// ReportPeerDiscovery reports a peer discovery event
func (pi *P2PIntegrator) ReportPeerDiscovery(discoveryType, result string) {
	pi.metrics.PeerDiscovery.WithLabelValues(discoveryType, result).Inc()
}

// API Integration Methods

// ReportAPIRequest reports an API request
func (ai *APIIntegrator) ReportAPIRequest(method, endpoint, statusCode string) {
	ai.metrics.RequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
}

// ReportRequestDuration reports API request duration
func (ai *APIIntegrator) ReportRequestDuration(method, endpoint string, duration time.Duration) {
	ai.metrics.RequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// ReportResponseSize reports API response size
func (ai *APIIntegrator) ReportResponseSize(method, endpoint string, sizeBytes float64) {
	ai.metrics.ResponseSize.WithLabelValues(method, endpoint).Observe(sizeBytes)
}

// ReportActiveConnections reports active connection count
func (ai *APIIntegrator) ReportActiveConnections(connectionType string, count float64) {
	ai.metrics.ActiveConnections.WithLabelValues(connectionType).Set(count)
}

// ReportWebSocketConnections reports WebSocket connection count
func (ai *APIIntegrator) ReportWebSocketConnections(endpoint string, count float64) {
	ai.metrics.WebSocketConnections.WithLabelValues(endpoint).Set(count)
}

// ReportRateLimitHit reports a rate limit hit
func (ai *APIIntegrator) ReportRateLimitHit(endpoint, clientID string) {
	ai.metrics.RateLimitHits.WithLabelValues(endpoint, clientID).Inc()
}

// Fault Tolerance Integration Methods

// ReportFaultDetected reports a detected fault
func (fti *FaultToleranceIntegrator) ReportFaultDetected(faultType, component, severity string) {
	fti.metrics.FaultsDetected.WithLabelValues(faultType, component, severity).Inc()
}

// ReportRecoveryAttempt reports a recovery attempt
func (fti *FaultToleranceIntegrator) ReportRecoveryAttempt(recoveryType, component string) {
	fti.metrics.RecoveryAttempts.WithLabelValues(recoveryType, component).Inc()
}

// ReportRecoverySuccess reports a successful recovery
func (fti *FaultToleranceIntegrator) ReportRecoverySuccess(recoveryType, component string) {
	fti.metrics.RecoverySuccess.WithLabelValues(recoveryType, component).Inc()
}

// ReportPredictionAccuracy reports prediction accuracy
func (fti *FaultToleranceIntegrator) ReportPredictionAccuracy(modelType, component string, accuracy float64) {
	fti.metrics.PredictionAccuracy.WithLabelValues(modelType, component).Set(accuracy)
}

// ReportHealingOperation reports a healing operation
func (fti *FaultToleranceIntegrator) ReportHealingOperation(healingType, component string) {
	fti.metrics.HealingOperations.WithLabelValues(healingType, component).Inc()
}

// ReportSystemHealth reports system health score
func (fti *FaultToleranceIntegrator) ReportSystemHealth(component, subsystem string, healthScore float64) {
	fti.metrics.SystemHealth.WithLabelValues(component, subsystem).Set(healthScore)
}

// Model Integration Methods

// ReportModelLoaded reports a loaded model
func (mi *ModelIntegrator) ReportModelLoaded(modelName string, count float64) {
	mi.metrics.ModelsLoaded.WithLabelValues(modelName, mi.nodeID).Set(count)
}

// ReportModelRequest reports a model inference request
func (mi *ModelIntegrator) ReportModelRequest(modelName, status string) {
	mi.metrics.ModelRequests.WithLabelValues(modelName, mi.nodeID, status).Inc()
}

// ReportModelLatency reports model inference latency
func (mi *ModelIntegrator) ReportModelLatency(modelName string, latency time.Duration) {
	mi.metrics.ModelLatency.WithLabelValues(modelName, mi.nodeID).Observe(latency.Seconds())
}

// ReportModelError reports a model error
func (mi *ModelIntegrator) ReportModelError(modelName, errorType string) {
	mi.metrics.ModelErrors.WithLabelValues(modelName, errorType, mi.nodeID).Inc()
}

// ReportReplicationOperation reports a replication operation
func (mi *ModelIntegrator) ReportReplicationOperation(operationType, modelName, status string) {
	mi.metrics.ReplicationOperations.WithLabelValues(operationType, modelName, status).Inc()
}

// ReportStorageUsage reports storage usage
func (mi *ModelIntegrator) ReportStorageUsage(modelName, storageType string, usageBytes float64) {
	mi.metrics.StorageUsage.WithLabelValues(modelName, mi.nodeID, storageType).Set(usageBytes)
}

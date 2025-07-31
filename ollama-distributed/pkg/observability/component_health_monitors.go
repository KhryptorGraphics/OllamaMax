package observability

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// SchedulerHealthMonitor monitors scheduler component health
type SchedulerHealthMonitor struct {
	name          string
	scheduler     SchedulerHealthChecker
	healthHistory []*ComponentHealthStatus
	historyMu     sync.RWMutex
	lastCheck     time.Time
	maxHistory    int
}

// ConsensusHealthMonitor monitors consensus component health
type ConsensusHealthMonitor struct {
	name          string
	consensus     ConsensusHealthChecker
	healthHistory []*ComponentHealthStatus
	historyMu     sync.RWMutex
	lastCheck     time.Time
	maxHistory    int
}

// P2PHealthMonitor monitors P2P component health
type P2PHealthMonitor struct {
	name          string
	p2pNode       P2PHealthChecker
	healthHistory []*ComponentHealthStatus
	historyMu     sync.RWMutex
	lastCheck     time.Time
	maxHistory    int
}

// APIGatewayHealthMonitor monitors API gateway component health
type APIGatewayHealthMonitor struct {
	name          string
	apiGateway    APIGatewayHealthChecker
	healthHistory []*ComponentHealthStatus
	historyMu     sync.RWMutex
	lastCheck     time.Time
	maxHistory    int
}

// FaultToleranceHealthMonitor monitors fault tolerance component health
type FaultToleranceHealthMonitor struct {
	name           string
	faultTolerance FaultToleranceHealthChecker
	healthHistory  []*ComponentHealthStatus
	historyMu      sync.RWMutex
	lastCheck      time.Time
	maxHistory     int
}

// ModelManagerHealthMonitor monitors model manager component health
type ModelManagerHealthMonitor struct {
	name          string
	modelManager  ModelManagerHealthChecker
	healthHistory []*ComponentHealthStatus
	historyMu     sync.RWMutex
	lastCheck     time.Time
	maxHistory    int
}

// SystemHealthMonitor monitors overall system health
type SystemHealthMonitor struct {
	name          string
	healthHistory []*ComponentHealthStatus
	historyMu     sync.RWMutex
	lastCheck     time.Time
	maxHistory    int
}

// Health checker interfaces for different components
type SchedulerHealthChecker interface {
	IsHealthy() bool
	GetClusterSize() int
	GetActiveTaskCount() int
	GetQueuedTaskCount() int
	GetWorkerCount() int
	GetLastActivity() time.Time
}

type ConsensusHealthChecker interface {
	IsHealthy() bool
	IsLeader() bool
	GetNodeState() string
	GetClusterSize() int
	GetLastCommit() time.Time
	HasQuorum() bool
}

type P2PHealthChecker interface {
	IsHealthy() bool
	GetConnectedPeerCount() int
	GetNetworkLatency() time.Duration
	GetLastActivity() time.Time
	IsNetworkConnected() bool
}

type APIGatewayHealthChecker interface {
	IsHealthy() bool
	GetActiveConnections() int
	GetRequestRate() float64
	GetErrorRate() float64
	GetLastRequest() time.Time
}

type FaultToleranceHealthChecker interface {
	IsHealthy() bool
	GetFaultCount() int
	GetRecoveryCount() int
	GetSystemHealth() float64
	GetLastFault() time.Time
}

type ModelManagerHealthChecker interface {
	IsHealthy() bool
	GetLoadedModelCount() int
	GetActiveInferences() int
	GetStorageUsage() float64
	GetLastActivity() time.Time
}

// NewSchedulerHealthMonitor creates a new scheduler health monitor
func NewSchedulerHealthMonitor(scheduler SchedulerHealthChecker) *SchedulerHealthMonitor {
	return &SchedulerHealthMonitor{
		name:          "scheduler",
		scheduler:     scheduler,
		healthHistory: make([]*ComponentHealthStatus, 0),
		maxHistory:    50,
	}
}

// GetComponentName returns the component name
func (shm *SchedulerHealthMonitor) GetComponentName() string {
	return shm.name
}

// CheckHealth checks the scheduler health
func (shm *SchedulerHealthMonitor) CheckHealth(ctx context.Context) *ComponentHealthStatus {
	start := time.Now()

	status := &ComponentHealthStatus{
		ComponentName: shm.name,
		Timestamp:     start,
		Metadata:      make(map[string]interface{}),
		Checks:        make([]*HealthCheck, 0),
	}

	// Check if scheduler is healthy
	isHealthy := shm.scheduler.IsHealthy()
	clusterSize := shm.scheduler.GetClusterSize()
	activeTasks := shm.scheduler.GetActiveTaskCount()
	queuedTasks := shm.scheduler.GetQueuedTaskCount()
	workerCount := shm.scheduler.GetWorkerCount()
	lastActivity := shm.scheduler.GetLastActivity()

	// Add metadata
	status.Metadata["cluster_size"] = clusterSize
	status.Metadata["active_tasks"] = activeTasks
	status.Metadata["queued_tasks"] = queuedTasks
	status.Metadata["worker_count"] = workerCount
	status.Metadata["last_activity"] = lastActivity.Unix()

	// Perform health checks
	checks := []*HealthCheck{
		{
			Name:    "scheduler_responsive",
			Status:  HealthStatusHealthy,
			Message: "Scheduler is responsive",
		},
		{
			Name:    "cluster_connectivity",
			Status:  shm.getClusterConnectivityStatus(clusterSize),
			Message: fmt.Sprintf("Cluster size: %d", clusterSize),
		},
		{
			Name:    "task_processing",
			Status:  shm.getTaskProcessingStatus(activeTasks, queuedTasks),
			Message: fmt.Sprintf("Active: %d, Queued: %d", activeTasks, queuedTasks),
		},
		{
			Name:    "worker_availability",
			Status:  shm.getWorkerAvailabilityStatus(workerCount),
			Message: fmt.Sprintf("Workers: %d", workerCount),
		},
	}

	status.Checks = checks

	// Determine overall status
	if !isHealthy {
		status.Status = HealthStatusUnhealthy
		status.Message = "Scheduler reports unhealthy status"
	} else {
		status.Status = shm.aggregateCheckStatus(checks)
		status.Message = shm.getStatusMessage(status.Status)
	}

	status.Latency = time.Since(start)
	shm.lastCheck = time.Now()

	// Add to history
	shm.addToHistory(status)

	return status
}

// GetHealthHistory returns the health history
func (shm *SchedulerHealthMonitor) GetHealthHistory() []*ComponentHealthStatus {
	shm.historyMu.RLock()
	defer shm.historyMu.RUnlock()

	history := make([]*ComponentHealthStatus, len(shm.healthHistory))
	copy(history, shm.healthHistory)
	return history
}

// IsHealthy returns whether the component is healthy
func (shm *SchedulerHealthMonitor) IsHealthy() bool {
	return shm.scheduler.IsHealthy()
}

// GetLastCheck returns the last check time
func (shm *SchedulerHealthMonitor) GetLastCheck() time.Time {
	return shm.lastCheck
}

// Helper methods for scheduler health monitor

func (shm *SchedulerHealthMonitor) getClusterConnectivityStatus(clusterSize int) HealthStatus {
	if clusterSize == 0 {
		return HealthStatusUnhealthy
	} else if clusterSize < 3 {
		return HealthStatusDegraded
	}
	return HealthStatusHealthy
}

func (shm *SchedulerHealthMonitor) getTaskProcessingStatus(activeTasks, queuedTasks int) HealthStatus {
	if queuedTasks > 100 {
		return HealthStatusDegraded
	} else if queuedTasks > 500 {
		return HealthStatusUnhealthy
	}
	return HealthStatusHealthy
}

func (shm *SchedulerHealthMonitor) getWorkerAvailabilityStatus(workerCount int) HealthStatus {
	if workerCount == 0 {
		return HealthStatusUnhealthy
	} else if workerCount < 2 {
		return HealthStatusDegraded
	}
	return HealthStatusHealthy
}

func (shm *SchedulerHealthMonitor) aggregateCheckStatus(checks []*HealthCheck) HealthStatus {
	hasUnhealthy := false
	hasDegraded := false

	for _, check := range checks {
		switch check.Status {
		case HealthStatusUnhealthy:
			hasUnhealthy = true
		case HealthStatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return HealthStatusUnhealthy
	} else if hasDegraded {
		return HealthStatusDegraded
	}
	return HealthStatusHealthy
}

func (shm *SchedulerHealthMonitor) getStatusMessage(status HealthStatus) string {
	switch status {
	case HealthStatusHealthy:
		return "Scheduler is operating normally"
	case HealthStatusDegraded:
		return "Scheduler is experiencing some issues"
	case HealthStatusUnhealthy:
		return "Scheduler is experiencing critical issues"
	default:
		return "Scheduler status unknown"
	}
}

func (shm *SchedulerHealthMonitor) addToHistory(status *ComponentHealthStatus) {
	shm.historyMu.Lock()
	defer shm.historyMu.Unlock()

	shm.healthHistory = append(shm.healthHistory, status)

	if len(shm.healthHistory) > shm.maxHistory {
		shm.healthHistory = shm.healthHistory[1:]
	}
}

// NewSystemHealthMonitor creates a new system health monitor
func NewSystemHealthMonitor() *SystemHealthMonitor {
	return &SystemHealthMonitor{
		name:          "system",
		healthHistory: make([]*ComponentHealthStatus, 0),
		maxHistory:    50,
	}
}

// GetComponentName returns the component name
func (sysm *SystemHealthMonitor) GetComponentName() string {
	return sysm.name
}

// CheckHealth checks the system health
func (sysm *SystemHealthMonitor) CheckHealth(ctx context.Context) *ComponentHealthStatus {
	start := time.Now()

	status := &ComponentHealthStatus{
		ComponentName: sysm.name,
		Timestamp:     start,
		Metadata:      make(map[string]interface{}),
		Checks:        make([]*HealthCheck, 0),
	}

	// Get system metrics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	cpuCount := runtime.NumCPU()
	goroutineCount := runtime.NumGoroutine()

	// Add metadata
	status.Metadata["cpu_count"] = cpuCount
	status.Metadata["goroutine_count"] = goroutineCount
	status.Metadata["memory_alloc"] = memStats.Alloc
	status.Metadata["memory_sys"] = memStats.Sys
	status.Metadata["gc_count"] = memStats.NumGC

	// Perform health checks
	checks := []*HealthCheck{
		{
			Name:    "memory_usage",
			Status:  sysm.getMemoryStatus(memStats),
			Message: fmt.Sprintf("Allocated: %d MB", memStats.Alloc/1024/1024),
		},
		{
			Name:    "goroutine_count",
			Status:  sysm.getGoroutineStatus(goroutineCount),
			Message: fmt.Sprintf("Goroutines: %d", goroutineCount),
		},
		{
			Name:    "gc_performance",
			Status:  sysm.getGCStatus(memStats),
			Message: fmt.Sprintf("GC cycles: %d", memStats.NumGC),
		},
	}

	status.Checks = checks
	status.Status = sysm.aggregateCheckStatus(checks)
	status.Message = sysm.getStatusMessage(status.Status)
	status.Latency = time.Since(start)
	sysm.lastCheck = time.Now()

	// Add to history
	sysm.addToHistory(status)

	return status
}

// GetHealthHistory returns the health history
func (sysm *SystemHealthMonitor) GetHealthHistory() []*ComponentHealthStatus {
	sysm.historyMu.RLock()
	defer sysm.historyMu.RUnlock()

	history := make([]*ComponentHealthStatus, len(sysm.healthHistory))
	copy(history, sysm.healthHistory)
	return history
}

// IsHealthy returns whether the component is healthy
func (sysm *SystemHealthMonitor) IsHealthy() bool {
	// Simple system health check
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Consider system healthy if memory usage is reasonable
	return memStats.Alloc < 1024*1024*1024 // Less than 1GB
}

// GetLastCheck returns the last check time
func (sysm *SystemHealthMonitor) GetLastCheck() time.Time {
	return sysm.lastCheck
}

// Helper methods for system health monitor

func (sysm *SystemHealthMonitor) getMemoryStatus(memStats runtime.MemStats) HealthStatus {
	allocMB := memStats.Alloc / 1024 / 1024

	if allocMB > 1024 { // More than 1GB
		return HealthStatusUnhealthy
	} else if allocMB > 512 { // More than 512MB
		return HealthStatusDegraded
	}
	return HealthStatusHealthy
}

func (sysm *SystemHealthMonitor) getGoroutineStatus(count int) HealthStatus {
	if count > 10000 {
		return HealthStatusUnhealthy
	} else if count > 5000 {
		return HealthStatusDegraded
	}
	return HealthStatusHealthy
}

func (sysm *SystemHealthMonitor) getGCStatus(memStats runtime.MemStats) HealthStatus {
	// Simple GC health check based on pause time
	if memStats.PauseTotalNs > 1000000000 { // More than 1 second total pause
		return HealthStatusDegraded
	}
	return HealthStatusHealthy
}

func (sysm *SystemHealthMonitor) aggregateCheckStatus(checks []*HealthCheck) HealthStatus {
	hasUnhealthy := false
	hasDegraded := false

	for _, check := range checks {
		switch check.Status {
		case HealthStatusUnhealthy:
			hasUnhealthy = true
		case HealthStatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return HealthStatusUnhealthy
	} else if hasDegraded {
		return HealthStatusDegraded
	}
	return HealthStatusHealthy
}

func (sysm *SystemHealthMonitor) getStatusMessage(status HealthStatus) string {
	switch status {
	case HealthStatusHealthy:
		return "System is operating normally"
	case HealthStatusDegraded:
		return "System is experiencing some performance issues"
	case HealthStatusUnhealthy:
		return "System is experiencing critical issues"
	default:
		return "System status unknown"
	}
}

func (sysm *SystemHealthMonitor) addToHistory(status *ComponentHealthStatus) {
	sysm.historyMu.Lock()
	defer sysm.historyMu.Unlock()

	sysm.healthHistory = append(sysm.healthHistory, status)

	if len(sysm.healthHistory) > sysm.maxHistory {
		sysm.healthHistory = sysm.healthHistory[1:]
	}
}

// NewP2PHealthMonitor creates a new P2P health monitor
func NewP2PHealthMonitor(p2pNode P2PHealthChecker) *P2PHealthMonitor {
	return &P2PHealthMonitor{
		name:          "p2p",
		p2pNode:       p2pNode,
		healthHistory: make([]*ComponentHealthStatus, 0),
		maxHistory:    50,
	}
}

// GetComponentName returns the component name
func (p2pm *P2PHealthMonitor) GetComponentName() string {
	return p2pm.name
}

// CheckHealth checks the P2P health
func (p2pm *P2PHealthMonitor) CheckHealth(ctx context.Context) *ComponentHealthStatus {
	start := time.Now()

	status := &ComponentHealthStatus{
		ComponentName: p2pm.name,
		Timestamp:     start,
		Metadata:      make(map[string]interface{}),
		Checks:        make([]*HealthCheck, 0),
	}

	// Check if P2P node is healthy
	isHealthy := p2pm.p2pNode.IsHealthy()
	connectedPeers := p2pm.p2pNode.GetConnectedPeerCount()
	networkLatency := p2pm.p2pNode.GetNetworkLatency()
	lastActivity := p2pm.p2pNode.GetLastActivity()
	isConnected := p2pm.p2pNode.IsNetworkConnected()

	// Add metadata
	status.Metadata["connected_peers"] = connectedPeers
	status.Metadata["network_latency_ms"] = networkLatency.Milliseconds()
	status.Metadata["last_activity"] = lastActivity.Unix()
	status.Metadata["is_connected"] = isConnected

	// Perform health checks
	checks := []*HealthCheck{
		{
			Name:    "p2p_responsive",
			Status:  HealthStatusHealthy,
			Message: "P2P node is responsive",
		},
		{
			Name:    "peer_connectivity",
			Status:  p2pm.getPeerConnectivityStatus(connectedPeers),
			Message: fmt.Sprintf("Connected peers: %d", connectedPeers),
		},
		{
			Name:    "network_latency",
			Status:  p2pm.getNetworkLatencyStatus(networkLatency),
			Message: fmt.Sprintf("Latency: %v", networkLatency),
		},
		{
			Name:    "connection_status",
			Status:  p2pm.getConnectionStatus(isConnected),
			Message: fmt.Sprintf("Connected: %v", isConnected),
		},
	}

	status.Checks = checks

	// Determine overall status
	if !isHealthy {
		status.Status = HealthStatusUnhealthy
		status.Message = "P2P node reports unhealthy status"
	} else {
		status.Status = p2pm.aggregateCheckStatus(checks)
		status.Message = p2pm.getStatusMessage(status.Status)
	}

	status.Latency = time.Since(start)
	p2pm.lastCheck = time.Now()

	// Add to history
	p2pm.addToHistory(status)

	return status
}

// GetHealthHistory returns the health history
func (p2pm *P2PHealthMonitor) GetHealthHistory() []*ComponentHealthStatus {
	p2pm.historyMu.RLock()
	defer p2pm.historyMu.RUnlock()

	history := make([]*ComponentHealthStatus, len(p2pm.healthHistory))
	copy(history, p2pm.healthHistory)
	return history
}

// IsHealthy returns whether the component is healthy
func (p2pm *P2PHealthMonitor) IsHealthy() bool {
	return p2pm.p2pNode.IsHealthy()
}

// GetLastCheck returns the last check time
func (p2pm *P2PHealthMonitor) GetLastCheck() time.Time {
	return p2pm.lastCheck
}

// Helper methods for P2P health monitor

func (p2pm *P2PHealthMonitor) getPeerConnectivityStatus(connectedPeers int) HealthStatus {
	if connectedPeers == 0 {
		return HealthStatusUnhealthy
	} else if connectedPeers < 2 {
		return HealthStatusDegraded
	}
	return HealthStatusHealthy
}

func (p2pm *P2PHealthMonitor) getNetworkLatencyStatus(latency time.Duration) HealthStatus {
	if latency > 1*time.Second {
		return HealthStatusUnhealthy
	} else if latency > 500*time.Millisecond {
		return HealthStatusDegraded
	}
	return HealthStatusHealthy
}

func (p2pm *P2PHealthMonitor) getConnectionStatus(isConnected bool) HealthStatus {
	if !isConnected {
		return HealthStatusUnhealthy
	}
	return HealthStatusHealthy
}

func (p2pm *P2PHealthMonitor) aggregateCheckStatus(checks []*HealthCheck) HealthStatus {
	hasUnhealthy := false
	hasDegraded := false

	for _, check := range checks {
		switch check.Status {
		case HealthStatusUnhealthy:
			hasUnhealthy = true
		case HealthStatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return HealthStatusUnhealthy
	} else if hasDegraded {
		return HealthStatusDegraded
	}
	return HealthStatusHealthy
}

func (p2pm *P2PHealthMonitor) getStatusMessage(status HealthStatus) string {
	switch status {
	case HealthStatusHealthy:
		return "P2P node is operating normally"
	case HealthStatusDegraded:
		return "P2P node is experiencing some issues"
	case HealthStatusUnhealthy:
		return "P2P node is experiencing critical issues"
	default:
		return "P2P node status unknown"
	}
}

func (p2pm *P2PHealthMonitor) addToHistory(status *ComponentHealthStatus) {
	p2pm.historyMu.Lock()
	defer p2pm.historyMu.Unlock()

	p2pm.healthHistory = append(p2pm.healthHistory, status)

	if len(p2pm.healthHistory) > p2pm.maxHistory {
		p2pm.healthHistory = p2pm.healthHistory[1:]
	}
}

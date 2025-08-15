package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusStarting  HealthStatus = "starting"
)

// ComponentHealth represents the health of a single component
type ComponentHealth struct {
	Name        string                 `json:"name"`
	Status      HealthStatus           `json:"status"`
	Message     string                 `json:"message,omitempty"`
	LastChecked time.Time              `json:"last_checked"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// OverallHealth represents the overall system health
type OverallHealth struct {
	Status     HealthStatus                `json:"status"`
	Timestamp  time.Time                   `json:"timestamp"`
	Components map[string]ComponentHealth  `json:"components"`
	Metadata   map[string]interface{}      `json:"metadata,omitempty"`
}

// HealthChecker interface for component health checks
type HealthChecker interface {
	Name() string
	Check(ctx context.Context) ComponentHealth
}

// KubernetesHealthManager manages health checks for Kubernetes compatibility
type KubernetesHealthManager struct {
	checkers    map[string]HealthChecker
	mu          sync.RWMutex
	logger      *logrus.Logger
	startupTime time.Time
	ready       bool
	live        bool
}

// NewKubernetesHealthManager creates a new health manager
func NewKubernetesHealthManager(logger *logrus.Logger) *KubernetesHealthManager {
	return &KubernetesHealthManager{
		checkers:    make(map[string]HealthChecker),
		logger:      logger,
		startupTime: time.Now(),
		ready:       false,
		live:        true,
	}
}

// RegisterChecker registers a health checker
func (hm *KubernetesHealthManager) RegisterChecker(checker HealthChecker) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.checkers[checker.Name()] = checker
}

// SetReady marks the service as ready
func (hm *KubernetesHealthManager) SetReady(ready bool) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.ready = ready
	
	correlationID := uuid.New().String()
	hm.logger.WithFields(logrus.Fields{
		"correlation_id": correlationID,
		"component":      "health_manager",
		"event":          "readiness_changed",
		"ready":          ready,
		"timestamp":      time.Now().UTC(),
	}).Info("Service readiness status changed")
}

// SetLive marks the service as live
func (hm *KubernetesHealthManager) SetLive(live bool) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.live = live
	
	correlationID := uuid.New().String()
	hm.logger.WithFields(logrus.Fields{
		"correlation_id": correlationID,
		"component":      "health_manager",
		"event":          "liveness_changed",
		"live":           live,
		"timestamp":      time.Now().UTC(),
	}).Info("Service liveness status changed")
}

// CheckHealth performs all health checks
func (hm *KubernetesHealthManager) CheckHealth(ctx context.Context) OverallHealth {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	correlationID := uuid.New().String()
	startTime := time.Now()

	components := make(map[string]ComponentHealth)
	overallStatus := HealthStatusHealthy

	// Run all health checks
	for name, checker := range hm.checkers {
		checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		componentHealth := checker.Check(checkCtx)
		cancel()

		components[name] = componentHealth

		// Determine overall status
		switch componentHealth.Status {
		case HealthStatusUnhealthy:
			overallStatus = HealthStatusUnhealthy
		case HealthStatusDegraded:
			if overallStatus == HealthStatusHealthy {
				overallStatus = HealthStatusDegraded
			}
		case HealthStatusStarting:
			if overallStatus == HealthStatusHealthy {
				overallStatus = HealthStatusStarting
			}
		}
	}

	health := OverallHealth{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Components: components,
		Metadata: map[string]interface{}{
			"startup_time":    hm.startupTime,
			"uptime_seconds":  time.Since(hm.startupTime).Seconds(),
			"correlation_id":  correlationID,
			"check_duration":  time.Since(startTime).Milliseconds(),
		},
	}

	// Log health check result
	hm.logger.WithFields(logrus.Fields{
		"correlation_id":     correlationID,
		"component":          "health_manager",
		"event":              "health_check_completed",
		"overall_status":     string(overallStatus),
		"components_checked": len(components),
		"check_duration_ms":  time.Since(startTime).Milliseconds(),
		"timestamp":          time.Now().UTC(),
	}).Info("Health check completed")

	return health
}

// LivenessHandler handles Kubernetes liveness probes
func (hm *KubernetesHealthManager) LivenessHandler(w http.ResponseWriter, r *http.Request) {
	correlationID := uuid.New().String()
	
	hm.mu.RLock()
	live := hm.live
	hm.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Correlation-ID", correlationID)

	response := map[string]interface{}{
		"status":        "ok",
		"live":          live,
		"timestamp":     time.Now().UTC(),
		"correlation_id": correlationID,
	}

	if live {
		w.WriteHeader(http.StatusOK)
		hm.logger.WithFields(logrus.Fields{
			"correlation_id": correlationID,
			"component":      "health_manager",
			"event":          "liveness_check",
			"result":         "healthy",
			"timestamp":      time.Now().UTC(),
		}).Debug("Liveness check passed")
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		response["status"] = "unhealthy"
		hm.logger.WithFields(logrus.Fields{
			"correlation_id": correlationID,
			"component":      "health_manager",
			"event":          "liveness_check",
			"result":         "unhealthy",
			"timestamp":      time.Now().UTC(),
		}).Warn("Liveness check failed")
	}

	json.NewEncoder(w).Encode(response)
}

// ReadinessHandler handles Kubernetes readiness probes
func (hm *KubernetesHealthManager) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	correlationID := uuid.New().String()
	ctx := context.WithValue(r.Context(), "correlation_id", correlationID)
	
	hm.mu.RLock()
	ready := hm.ready
	hm.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Correlation-ID", correlationID)

	// Perform quick health checks for readiness
	health := hm.CheckHealth(ctx)
	
	isReady := ready && (health.Status == HealthStatusHealthy || health.Status == HealthStatusDegraded)

	response := map[string]interface{}{
		"status":         "ok",
		"ready":          isReady,
		"overall_health": string(health.Status),
		"timestamp":      time.Now().UTC(),
		"correlation_id": correlationID,
		"components":     len(health.Components),
	}

	if isReady {
		w.WriteHeader(http.StatusOK)
		hm.logger.WithFields(logrus.Fields{
			"correlation_id":   correlationID,
			"component":        "health_manager",
			"event":            "readiness_check",
			"result":           "ready",
			"overall_health":   string(health.Status),
			"components_count": len(health.Components),
			"timestamp":        time.Now().UTC(),
		}).Debug("Readiness check passed")
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		response["status"] = "not_ready"
		hm.logger.WithFields(logrus.Fields{
			"correlation_id":   correlationID,
			"component":        "health_manager",
			"event":            "readiness_check",
			"result":           "not_ready",
			"ready_flag":       ready,
			"overall_health":   string(health.Status),
			"components_count": len(health.Components),
			"timestamp":        time.Now().UTC(),
		}).Warn("Readiness check failed")
	}

	json.NewEncoder(w).Encode(response)
}

// StartupHandler handles Kubernetes startup probes
func (hm *KubernetesHealthManager) StartupHandler(w http.ResponseWriter, r *http.Request) {
	correlationID := uuid.New().String()
	ctx := context.WithValue(r.Context(), "correlation_id", correlationID)
	
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Correlation-ID", correlationID)

	health := hm.CheckHealth(ctx)
	uptime := time.Since(hm.startupTime)
	
	// Consider started if not in starting state and uptime > 30 seconds
	isStarted := health.Status != HealthStatusStarting && uptime > 30*time.Second

	response := map[string]interface{}{
		"status":         "ok",
		"started":        isStarted,
		"overall_health": string(health.Status),
		"uptime_seconds": uptime.Seconds(),
		"timestamp":      time.Now().UTC(),
		"correlation_id": correlationID,
	}

	if isStarted {
		w.WriteHeader(http.StatusOK)
		hm.logger.WithFields(logrus.Fields{
			"correlation_id": correlationID,
			"component":      "health_manager",
			"event":          "startup_check",
			"result":         "started",
			"uptime_seconds": uptime.Seconds(),
			"overall_health": string(health.Status),
			"timestamp":      time.Now().UTC(),
		}).Debug("Startup check passed")
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		response["status"] = "starting"
		hm.logger.WithFields(logrus.Fields{
			"correlation_id": correlationID,
			"component":      "health_manager",
			"event":          "startup_check",
			"result":         "starting",
			"uptime_seconds": uptime.Seconds(),
			"overall_health": string(health.Status),
			"timestamp":      time.Now().UTC(),
		}).Info("Startup check - still starting")
	}

	json.NewEncoder(w).Encode(response)
}

// HealthHandler provides detailed health information
func (hm *KubernetesHealthManager) HealthHandler(w http.ResponseWriter, r *http.Request) {
	correlationID := uuid.New().String()
	ctx := context.WithValue(r.Context(), "correlation_id", correlationID)
	
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Correlation-ID", correlationID)

	health := hm.CheckHealth(ctx)

	switch health.Status {
	case HealthStatusHealthy, HealthStatusDegraded:
		w.WriteHeader(http.StatusOK)
	case HealthStatusStarting:
		w.WriteHeader(http.StatusAccepted)
	default:
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	hm.logger.WithFields(logrus.Fields{
		"correlation_id":   correlationID,
		"component":        "health_manager",
		"event":            "detailed_health_check",
		"overall_status":   string(health.Status),
		"components_count": len(health.Components),
		"timestamp":        time.Now().UTC(),
	}).Info("Detailed health check requested")

	json.NewEncoder(w).Encode(health)
}

// FaultToleranceHealthChecker checks fault tolerance system health
type FaultToleranceHealthChecker struct {
	manager interface {
		GetHealthStatus() (bool, string, map[string]interface{})
	}
}

// NewFaultToleranceHealthChecker creates a new fault tolerance health checker
func NewFaultToleranceHealthChecker(manager interface {
	GetHealthStatus() (bool, string, map[string]interface{})
}) *FaultToleranceHealthChecker {
	return &FaultToleranceHealthChecker{manager: manager}
}

// Name returns the checker name
func (c *FaultToleranceHealthChecker) Name() string {
	return "fault_tolerance"
}

// Check performs the health check
func (c *FaultToleranceHealthChecker) Check(ctx context.Context) ComponentHealth {
	healthy, message, details := c.manager.GetHealthStatus()
	
	status := HealthStatusHealthy
	if !healthy {
		status = HealthStatusUnhealthy
	}

	return ComponentHealth{
		Name:        c.Name(),
		Status:      status,
		Message:     message,
		LastChecked: time.Now(),
		Details:     details,
	}
}

// ClusterHealthChecker checks cluster consensus health
type ClusterHealthChecker struct {
	consensus interface {
		IsLeader() bool
		GetClusterSize() int
		GetHealthyNodes() int
	}
}

// NewClusterHealthChecker creates a new cluster health checker
func NewClusterHealthChecker(consensus interface {
	IsLeader() bool
	GetClusterSize() int
	GetHealthyNodes() int
}) *ClusterHealthChecker {
	return &ClusterHealthChecker{consensus: consensus}
}

// Name returns the checker name
func (c *ClusterHealthChecker) Name() string {
	return "cluster_consensus"
}

// Check performs the health check
func (c *ClusterHealthChecker) Check(ctx context.Context) ComponentHealth {
	clusterSize := c.consensus.GetClusterSize()
	healthyNodes := c.consensus.GetHealthyNodes()
	isLeader := c.consensus.IsLeader()

	status := HealthStatusHealthy
	message := fmt.Sprintf("Cluster healthy: %d/%d nodes", healthyNodes, clusterSize)

	if healthyNodes < clusterSize/2+1 {
		status = HealthStatusUnhealthy
		message = fmt.Sprintf("Cluster unhealthy: %d/%d nodes (no quorum)", healthyNodes, clusterSize)
	} else if healthyNodes < clusterSize {
		status = HealthStatusDegraded
		message = fmt.Sprintf("Cluster degraded: %d/%d nodes", healthyNodes, clusterSize)
	}

	details := map[string]interface{}{
		"cluster_size":   clusterSize,
		"healthy_nodes":  healthyNodes,
		"is_leader":      isLeader,
		"has_quorum":     healthyNodes >= clusterSize/2+1,
	}

	return ComponentHealth{
		Name:        c.Name(),
		Status:      status,
		Message:     message,
		LastChecked: time.Now(),
		Details:     details,
	}
}

package observability

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// HealthCheckManager manages health checks across all system components
type HealthCheckManager struct {
	config *HealthCheckConfig

	// Component health monitors
	componentMonitors map[string]ComponentHealthMonitor

	// Dependency health checkers
	dependencyCheckers map[string]DependencyHealthChecker

	// Health status aggregation
	healthAggregator *HealthAggregator

	// Metrics integration
	metricsIntegration *MetricsIntegration

	// HTTP server for health endpoints
	httpServer *http.Server

	// Lifecycle
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	started bool
	mu      sync.RWMutex
}

// HealthCheckConfig configures the health check system
type HealthCheckConfig struct {
	// HTTP server settings
	ListenAddress string `json:"listen_address"`

	// Health check intervals
	ComponentCheckInterval  time.Duration `json:"component_check_interval"`
	DependencyCheckInterval time.Duration `json:"dependency_check_interval"`

	// Timeouts
	HealthCheckTimeout time.Duration `json:"health_check_timeout"`

	// Thresholds
	UnhealthyThreshold int `json:"unhealthy_threshold"`
	DegradedThreshold  int `json:"degraded_threshold"`

	// Features
	EnableMetricsIntegration bool `json:"enable_metrics_integration"`
	EnableKubernetesProbes   bool `json:"enable_kubernetes_probes"`
	EnableDependencyChecks   bool `json:"enable_dependency_checks"`
}

// ComponentHealthMonitor interface for component health monitoring
type ComponentHealthMonitor interface {
	GetComponentName() string
	CheckHealth(ctx context.Context) *ComponentHealthStatus
	GetHealthHistory() []*ComponentHealthStatus
	IsHealthy() bool
	GetLastCheck() time.Time
}

// DependencyHealthChecker interface for dependency health checking
type DependencyHealthChecker interface {
	GetDependencyName() string
	CheckDependency(ctx context.Context) *DependencyHealthStatus
	GetDependencyType() DependencyType
	IsRequired() bool
}

// ComponentHealthStatus represents the health status of a component
type ComponentHealthStatus struct {
	ComponentName string                 `json:"component_name"`
	Status        HealthStatus           `json:"status"`
	Message       string                 `json:"message"`
	Timestamp     time.Time              `json:"timestamp"`
	Latency       time.Duration          `json:"latency"`
	Metadata      map[string]interface{} `json:"metadata"`
	Checks        []*HealthCheck         `json:"checks"`
}

// DependencyHealthStatus represents the health status of a dependency
type DependencyHealthStatus struct {
	DependencyName string                 `json:"dependency_name"`
	Type           DependencyType         `json:"type"`
	Status         HealthStatus           `json:"status"`
	Message        string                 `json:"message"`
	Timestamp      time.Time              `json:"timestamp"`
	Latency        time.Duration          `json:"latency"`
	Required       bool                   `json:"required"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// HealthCheck represents an individual health check
type HealthCheck struct {
	Name     string                 `json:"name"`
	Status   HealthStatus           `json:"status"`
	Message  string                 `json:"message"`
	Latency  time.Duration          `json:"latency"`
	Metadata map[string]interface{} `json:"metadata"`
}

// HealthStatus represents the health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// DependencyType represents the type of dependency
type DependencyType string

const (
	DependencyTypeDatabase DependencyType = "database"
	DependencyTypeCache    DependencyType = "cache"
	DependencyTypeStorage  DependencyType = "storage"
	DependencyTypeService  DependencyType = "service"
	DependencyTypeNetwork  DependencyType = "network"
	DependencyTypeExternal DependencyType = "external"
)

// OverallHealthStatus represents the overall system health
type OverallHealthStatus struct {
	Status       HealthStatus                       `json:"status"`
	Message      string                             `json:"message"`
	Timestamp    time.Time                          `json:"timestamp"`
	Components   map[string]*ComponentHealthStatus  `json:"components"`
	Dependencies map[string]*DependencyHealthStatus `json:"dependencies"`
	Summary      *HealthSummary                     `json:"summary"`
}

// HealthSummary provides a summary of health status
type HealthSummary struct {
	TotalComponents       int `json:"total_components"`
	HealthyComponents     int `json:"healthy_components"`
	DegradedComponents    int `json:"degraded_components"`
	UnhealthyComponents   int `json:"unhealthy_components"`
	TotalDependencies     int `json:"total_dependencies"`
	HealthyDependencies   int `json:"healthy_dependencies"`
	DegradedDependencies  int `json:"degraded_dependencies"`
	UnhealthyDependencies int `json:"unhealthy_dependencies"`
}

// NewHealthCheckManager creates a new health check manager
func NewHealthCheckManager(config *HealthCheckConfig, metricsIntegration *MetricsIntegration) *HealthCheckManager {
	if config == nil {
		config = DefaultHealthCheckConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	hcm := &HealthCheckManager{
		config:             config,
		componentMonitors:  make(map[string]ComponentHealthMonitor),
		dependencyCheckers: make(map[string]DependencyHealthChecker),
		metricsIntegration: metricsIntegration,
		ctx:                ctx,
		cancel:             cancel,
	}

	// Create health aggregator
	hcm.healthAggregator = NewHealthAggregator(config)

	return hcm
}

// DefaultHealthCheckConfig returns default health check configuration
func DefaultHealthCheckConfig() *HealthCheckConfig {
	return &HealthCheckConfig{
		ListenAddress:            ":8081",
		ComponentCheckInterval:   30 * time.Second,
		DependencyCheckInterval:  60 * time.Second,
		HealthCheckTimeout:       10 * time.Second,
		UnhealthyThreshold:       3,
		DegradedThreshold:        1,
		EnableMetricsIntegration: true,
		EnableKubernetesProbes:   true,
		EnableDependencyChecks:   true,
	}
}

// Start starts the health check manager
func (hcm *HealthCheckManager) Start() error {
	hcm.mu.Lock()
	defer hcm.mu.Unlock()

	if hcm.started {
		return nil
	}

	// Start health check loops
	hcm.wg.Add(2)
	go hcm.componentHealthCheckLoop()
	go hcm.dependencyHealthCheckLoop()

	// Start HTTP server for health endpoints
	if err := hcm.startHTTPServer(); err != nil {
		return fmt.Errorf("failed to start health check HTTP server: %w", err)
	}

	hcm.started = true
	log.Info().Str("address", hcm.config.ListenAddress).Msg("Health check manager started")
	return nil
}

// Stop stops the health check manager
func (hcm *HealthCheckManager) Stop() error {
	hcm.mu.Lock()
	defer hcm.mu.Unlock()

	if !hcm.started {
		return nil
	}

	// Stop HTTP server
	if hcm.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := hcm.httpServer.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to shutdown health check HTTP server gracefully")
		}
	}

	// Stop health check loops
	hcm.cancel()
	hcm.wg.Wait()

	hcm.started = false
	log.Info().Msg("Health check manager stopped")
	return nil
}

// RegisterComponentMonitor registers a component health monitor
func (hcm *HealthCheckManager) RegisterComponentMonitor(monitor ComponentHealthMonitor) {
	hcm.mu.Lock()
	defer hcm.mu.Unlock()

	componentName := monitor.GetComponentName()
	hcm.componentMonitors[componentName] = monitor

	log.Info().Str("component", componentName).Msg("Component health monitor registered")
}

// RegisterDependencyChecker registers a dependency health checker
func (hcm *HealthCheckManager) RegisterDependencyChecker(checker DependencyHealthChecker) {
	hcm.mu.Lock()
	defer hcm.mu.Unlock()

	dependencyName := checker.GetDependencyName()
	hcm.dependencyCheckers[dependencyName] = checker

	log.Info().Str("dependency", dependencyName).Msg("Dependency health checker registered")
}

// GetOverallHealth returns the overall system health status
func (hcm *HealthCheckManager) GetOverallHealth() *OverallHealthStatus {
	return hcm.healthAggregator.GetOverallHealth(hcm.componentMonitors, hcm.dependencyCheckers)
}

// IsHealthy returns whether the system is healthy
func (hcm *HealthCheckManager) IsHealthy() bool {
	overallHealth := hcm.GetOverallHealth()
	return overallHealth.Status == HealthStatusHealthy
}

// IsReady returns whether the system is ready (for Kubernetes readiness probes)
func (hcm *HealthCheckManager) IsReady() bool {
	overallHealth := hcm.GetOverallHealth()

	// System is ready if all required dependencies are healthy
	for _, depStatus := range overallHealth.Dependencies {
		if depStatus.Required && depStatus.Status == HealthStatusUnhealthy {
			return false
		}
	}

	// System is ready if core components are at least degraded
	for componentName, compStatus := range overallHealth.Components {
		if hcm.isCoreComponent(componentName) && compStatus.Status == HealthStatusUnhealthy {
			return false
		}
	}

	return true
}

// IsLive returns whether the system is live (for Kubernetes liveness probes)
func (hcm *HealthCheckManager) IsLive() bool {
	// System is live if at least one core component is responsive
	overallHealth := hcm.GetOverallHealth()

	for componentName, compStatus := range overallHealth.Components {
		if hcm.isCoreComponent(componentName) && compStatus.Status != HealthStatusUnknown {
			return true
		}
	}

	return false
}

// isCoreComponent determines if a component is core to system operation
func (hcm *HealthCheckManager) isCoreComponent(componentName string) bool {
	coreComponents := []string{"scheduler", "consensus", "p2p", "api_gateway"}
	for _, core := range coreComponents {
		if componentName == core {
			return true
		}
	}
	return false
}

// componentHealthCheckLoop runs the component health check loop
func (hcm *HealthCheckManager) componentHealthCheckLoop() {
	defer hcm.wg.Done()

	ticker := time.NewTicker(hcm.config.ComponentCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-hcm.ctx.Done():
			return
		case <-ticker.C:
			hcm.checkAllComponents()
		}
	}
}

// dependencyHealthCheckLoop runs the dependency health check loop
func (hcm *HealthCheckManager) dependencyHealthCheckLoop() {
	defer hcm.wg.Done()

	ticker := time.NewTicker(hcm.config.DependencyCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-hcm.ctx.Done():
			return
		case <-ticker.C:
			if hcm.config.EnableDependencyChecks {
				hcm.checkAllDependencies()
			}
		}
	}
}

// checkAllComponents checks health of all registered components
func (hcm *HealthCheckManager) checkAllComponents() {
	hcm.mu.RLock()
	monitors := make(map[string]ComponentHealthMonitor)
	for name, monitor := range hcm.componentMonitors {
		monitors[name] = monitor
	}
	hcm.mu.RUnlock()

	for name, monitor := range monitors {
		go func(name string, monitor ComponentHealthMonitor) {
			ctx, cancel := context.WithTimeout(hcm.ctx, hcm.config.HealthCheckTimeout)
			defer cancel()

			status := monitor.CheckHealth(ctx)
			hcm.reportComponentHealth(name, status)
		}(name, monitor)
	}
}

// checkAllDependencies checks health of all registered dependencies
func (hcm *HealthCheckManager) checkAllDependencies() {
	hcm.mu.RLock()
	checkers := make(map[string]DependencyHealthChecker)
	for name, checker := range hcm.dependencyCheckers {
		checkers[name] = checker
	}
	hcm.mu.RUnlock()

	for name, checker := range checkers {
		go func(name string, checker DependencyHealthChecker) {
			ctx, cancel := context.WithTimeout(hcm.ctx, hcm.config.HealthCheckTimeout)
			defer cancel()

			status := checker.CheckDependency(ctx)
			hcm.reportDependencyHealth(name, status)
		}(name, checker)
	}
}

// reportComponentHealth reports component health to metrics
func (hcm *HealthCheckManager) reportComponentHealth(componentName string, status *ComponentHealthStatus) {
	if hcm.config.EnableMetricsIntegration && hcm.metricsIntegration != nil {
		// Report health status to metrics
		healthValue := hcm.healthStatusToFloat(status.Status)

		// This would integrate with the metrics system to report health status
		// Implementation depends on the specific metrics integration
		log.Debug().
			Str("component", componentName).
			Str("status", string(status.Status)).
			Dur("latency", status.Latency).
			Float64("health_value", healthValue).
			Msg("Component health reported")
	}
}

// reportDependencyHealth reports dependency health to metrics
func (hcm *HealthCheckManager) reportDependencyHealth(dependencyName string, status *DependencyHealthStatus) {
	if hcm.config.EnableMetricsIntegration && hcm.metricsIntegration != nil {
		// Report dependency health status to metrics
		healthValue := hcm.healthStatusToFloat(status.Status)

		log.Debug().
			Str("dependency", dependencyName).
			Str("status", string(status.Status)).
			Dur("latency", status.Latency).
			Float64("health_value", healthValue).
			Msg("Dependency health reported")
	}
}

// healthStatusToFloat converts health status to float for metrics
func (hcm *HealthCheckManager) healthStatusToFloat(status HealthStatus) float64 {
	switch status {
	case HealthStatusHealthy:
		return 1.0
	case HealthStatusDegraded:
		return 0.5
	case HealthStatusUnhealthy:
		return 0.0
	default:
		return -1.0
	}
}

// startHTTPServer starts the HTTP server for health endpoints
func (hcm *HealthCheckManager) startHTTPServer() error {
	router := gin.New()
	router.Use(gin.Recovery())

	// Health endpoints
	router.GET("/health", hcm.handleHealth)
	router.GET("/ready", hcm.handleReady)
	router.GET("/live", hcm.handleLive)

	// Detailed health endpoints
	router.GET("/health/detailed", hcm.handleDetailedHealth)
	router.GET("/health/components", hcm.handleComponentsHealth)
	router.GET("/health/dependencies", hcm.handleDependenciesHealth)

	hcm.httpServer = &http.Server{
		Addr:    hcm.config.ListenAddress,
		Handler: router,
	}

	// Start server in background
	go func() {
		if err := hcm.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("Health check HTTP server failed")
		}
	}()

	return nil
}

// handleHealth handles the basic health endpoint
func (hcm *HealthCheckManager) handleHealth(c *gin.Context) {
	overallHealth := hcm.GetOverallHealth()

	statusCode := http.StatusOK
	if overallHealth.Status == HealthStatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	} else if overallHealth.Status == HealthStatusDegraded {
		statusCode = http.StatusOK // Still return 200 for degraded
	}

	response := gin.H{
		"status":    string(overallHealth.Status),
		"timestamp": overallHealth.Timestamp.Unix(),
		"message":   overallHealth.Message,
	}

	c.JSON(statusCode, response)
}

// handleReady handles the Kubernetes readiness probe endpoint
func (hcm *HealthCheckManager) handleReady(c *gin.Context) {
	if hcm.IsReady() {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ready",
			"timestamp": time.Now().Unix(),
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "not_ready",
			"timestamp": time.Now().Unix(),
		})
	}
}

// handleLive handles the Kubernetes liveness probe endpoint
func (hcm *HealthCheckManager) handleLive(c *gin.Context) {
	if hcm.IsLive() {
		c.JSON(http.StatusOK, gin.H{
			"status":    "live",
			"timestamp": time.Now().Unix(),
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "not_live",
			"timestamp": time.Now().Unix(),
		})
	}
}

// handleDetailedHealth handles the detailed health endpoint
func (hcm *HealthCheckManager) handleDetailedHealth(c *gin.Context) {
	overallHealth := hcm.GetOverallHealth()

	statusCode := http.StatusOK
	if overallHealth.Status == HealthStatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, overallHealth)
}

// handleComponentsHealth handles the components health endpoint
func (hcm *HealthCheckManager) handleComponentsHealth(c *gin.Context) {
	overallHealth := hcm.GetOverallHealth()

	c.JSON(http.StatusOK, gin.H{
		"components": overallHealth.Components,
		"summary": gin.H{
			"total":     overallHealth.Summary.TotalComponents,
			"healthy":   overallHealth.Summary.HealthyComponents,
			"degraded":  overallHealth.Summary.DegradedComponents,
			"unhealthy": overallHealth.Summary.UnhealthyComponents,
		},
	})
}

// handleDependenciesHealth handles the dependencies health endpoint
func (hcm *HealthCheckManager) handleDependenciesHealth(c *gin.Context) {
	overallHealth := hcm.GetOverallHealth()

	c.JSON(http.StatusOK, gin.H{
		"dependencies": overallHealth.Dependencies,
		"summary": gin.H{
			"total":     overallHealth.Summary.TotalDependencies,
			"healthy":   overallHealth.Summary.HealthyDependencies,
			"degraded":  overallHealth.Summary.DegradedDependencies,
			"unhealthy": overallHealth.Summary.UnhealthyDependencies,
		},
	})
}

package observability

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// HealthAggregator aggregates health status from multiple components and dependencies
type HealthAggregator struct {
	config *HealthCheckConfig
	
	// Cached health status
	lastOverallHealth *OverallHealthStatus
	lastUpdate        time.Time
	cacheMu           sync.RWMutex
	
	// Health history
	healthHistory []*OverallHealthStatus
	historyMu     sync.RWMutex
	maxHistory    int
}

// NewHealthAggregator creates a new health aggregator
func NewHealthAggregator(config *HealthCheckConfig) *HealthAggregator {
	return &HealthAggregator{
		config:        config,
		healthHistory: make([]*OverallHealthStatus, 0),
		maxHistory:    100, // Keep last 100 health status records
	}
}

// GetOverallHealth returns the overall system health status
func (ha *HealthAggregator) GetOverallHealth(
	componentMonitors map[string]ComponentHealthMonitor,
	dependencyCheckers map[string]DependencyHealthChecker,
) *OverallHealthStatus {
	
	// Check if we have a recent cached result
	ha.cacheMu.RLock()
	if ha.lastOverallHealth != nil && time.Since(ha.lastUpdate) < 5*time.Second {
		cached := ha.lastOverallHealth
		ha.cacheMu.RUnlock()
		return cached
	}
	ha.cacheMu.RUnlock()
	
	// Collect component health status
	components := make(map[string]*ComponentHealthStatus)
	for name, monitor := range componentMonitors {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		status := monitor.CheckHealth(ctx)
		cancel()
		
		if status != nil {
			components[name] = status
		} else {
			// Create unknown status if check failed
			components[name] = &ComponentHealthStatus{
				ComponentName: name,
				Status:        HealthStatusUnknown,
				Message:       "Health check failed",
				Timestamp:     time.Now(),
			}
		}
	}
	
	// Collect dependency health status
	dependencies := make(map[string]*DependencyHealthStatus)
	for name, checker := range dependencyCheckers {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		status := checker.CheckDependency(ctx)
		cancel()
		
		if status != nil {
			dependencies[name] = status
		} else {
			// Create unknown status if check failed
			dependencies[name] = &DependencyHealthStatus{
				DependencyName: name,
				Type:           checker.GetDependencyType(),
				Status:         HealthStatusUnknown,
				Message:        "Dependency check failed",
				Timestamp:      time.Now(),
				Required:       checker.IsRequired(),
			}
		}
	}
	
	// Calculate overall health status
	overallStatus := ha.calculateOverallStatus(components, dependencies)
	
	// Create summary
	summary := ha.createHealthSummary(components, dependencies)
	
	// Create overall health status
	overallHealth := &OverallHealthStatus{
		Status:       overallStatus.Status,
		Message:      overallStatus.Message,
		Timestamp:    time.Now(),
		Components:   components,
		Dependencies: dependencies,
		Summary:      summary,
	}
	
	// Cache the result
	ha.cacheMu.Lock()
	ha.lastOverallHealth = overallHealth
	ha.lastUpdate = time.Now()
	ha.cacheMu.Unlock()
	
	// Add to history
	ha.addToHistory(overallHealth)
	
	return overallHealth
}

// OverallStatusResult represents the calculated overall status
type OverallStatusResult struct {
	Status  HealthStatus
	Message string
}

// calculateOverallStatus calculates the overall system health status
func (ha *HealthAggregator) calculateOverallStatus(
	components map[string]*ComponentHealthStatus,
	dependencies map[string]*DependencyHealthStatus,
) *OverallStatusResult {
	
	// Check for critical failures first
	
	// 1. Check required dependencies
	for _, depStatus := range dependencies {
		if depStatus.Required && depStatus.Status == HealthStatusUnhealthy {
			return &OverallStatusResult{
				Status:  HealthStatusUnhealthy,
				Message: "Critical dependency failure: " + depStatus.DependencyName,
			}
		}
	}
	
	// 2. Check core components
	coreComponents := []string{"scheduler", "consensus", "p2p", "api_gateway"}
	unhealthyCoreComponents := 0
	degradedCoreComponents := 0
	
	for _, coreComponent := range coreComponents {
		if compStatus, exists := components[coreComponent]; exists {
			switch compStatus.Status {
			case HealthStatusUnhealthy:
				unhealthyCoreComponents++
			case HealthStatusDegraded:
				degradedCoreComponents++
			}
		}
	}
	
	// If more than half of core components are unhealthy, system is unhealthy
	if unhealthyCoreComponents > len(coreComponents)/2 {
		return &OverallStatusResult{
			Status:  HealthStatusUnhealthy,
			Message: "Majority of core components are unhealthy",
		}
	}
	
	// If any core component is unhealthy, system is degraded
	if unhealthyCoreComponents > 0 {
		return &OverallStatusResult{
			Status:  HealthStatusDegraded,
			Message: "Some core components are unhealthy",
		}
	}
	
	// Check for degraded dependencies
	degradedRequiredDeps := 0
	for _, depStatus := range dependencies {
		if depStatus.Required && depStatus.Status == HealthStatusDegraded {
			degradedRequiredDeps++
		}
	}
	
	// If core components are degraded or required dependencies are degraded
	if degradedCoreComponents > 0 || degradedRequiredDeps > 0 {
		return &OverallStatusResult{
			Status:  HealthStatusDegraded,
			Message: "Some components or dependencies are degraded",
		}
	}
	
	// Check overall component health
	totalComponents := len(components)
	unhealthyComponents := 0
	degradedComponents := 0
	
	for _, compStatus := range components {
		switch compStatus.Status {
		case HealthStatusUnhealthy:
			unhealthyComponents++
		case HealthStatusDegraded:
			degradedComponents++
		}
	}
	
	// If more than 25% of components are unhealthy, system is degraded
	if totalComponents > 0 && float64(unhealthyComponents)/float64(totalComponents) > 0.25 {
		return &OverallStatusResult{
			Status:  HealthStatusDegraded,
			Message: "High number of unhealthy components",
		}
	}
	
	// If any components are degraded, system is degraded
	if degradedComponents > 0 {
		return &OverallStatusResult{
			Status:  HealthStatusDegraded,
			Message: "Some components are degraded",
		}
	}
	
	// System is healthy
	return &OverallStatusResult{
		Status:  HealthStatusHealthy,
		Message: "All systems operational",
	}
}

// createHealthSummary creates a health summary
func (ha *HealthAggregator) createHealthSummary(
	components map[string]*ComponentHealthStatus,
	dependencies map[string]*DependencyHealthStatus,
) *HealthSummary {
	
	summary := &HealthSummary{
		TotalComponents:   len(components),
		TotalDependencies: len(dependencies),
	}
	
	// Count component health status
	for _, compStatus := range components {
		switch compStatus.Status {
		case HealthStatusHealthy:
			summary.HealthyComponents++
		case HealthStatusDegraded:
			summary.DegradedComponents++
		case HealthStatusUnhealthy:
			summary.UnhealthyComponents++
		}
	}
	
	// Count dependency health status
	for _, depStatus := range dependencies {
		switch depStatus.Status {
		case HealthStatusHealthy:
			summary.HealthyDependencies++
		case HealthStatusDegraded:
			summary.DegradedDependencies++
		case HealthStatusUnhealthy:
			summary.UnhealthyDependencies++
		}
	}
	
	return summary
}

// addToHistory adds a health status to the history
func (ha *HealthAggregator) addToHistory(overallHealth *OverallHealthStatus) {
	ha.historyMu.Lock()
	defer ha.historyMu.Unlock()
	
	// Add to history
	ha.healthHistory = append(ha.healthHistory, overallHealth)
	
	// Trim history if it exceeds max size
	if len(ha.healthHistory) > ha.maxHistory {
		ha.healthHistory = ha.healthHistory[1:]
	}
	
	log.Debug().
		Str("status", string(overallHealth.Status)).
		Int("components", overallHealth.Summary.TotalComponents).
		Int("dependencies", overallHealth.Summary.TotalDependencies).
		Msg("Health status added to history")
}

// GetHealthHistory returns the health history
func (ha *HealthAggregator) GetHealthHistory() []*OverallHealthStatus {
	ha.historyMu.RLock()
	defer ha.historyMu.RUnlock()
	
	// Return a copy of the history
	history := make([]*OverallHealthStatus, len(ha.healthHistory))
	copy(history, ha.healthHistory)
	
	return history
}

// GetHealthTrend returns the health trend over time
func (ha *HealthAggregator) GetHealthTrend(duration time.Duration) *HealthTrend {
	ha.historyMu.RLock()
	defer ha.historyMu.RUnlock()
	
	cutoff := time.Now().Add(-duration)
	
	trend := &HealthTrend{
		Duration:    duration,
		StartTime:   cutoff,
		EndTime:     time.Now(),
		DataPoints:  make([]*HealthDataPoint, 0),
	}
	
	for _, health := range ha.healthHistory {
		if health.Timestamp.After(cutoff) {
			dataPoint := &HealthDataPoint{
				Timestamp:           health.Timestamp,
				Status:              health.Status,
				HealthyComponents:   health.Summary.HealthyComponents,
				DegradedComponents:  health.Summary.DegradedComponents,
				UnhealthyComponents: health.Summary.UnhealthyComponents,
			}
			trend.DataPoints = append(trend.DataPoints, dataPoint)
		}
	}
	
	return trend
}

// HealthTrend represents health trend over time
type HealthTrend struct {
	Duration   time.Duration      `json:"duration"`
	StartTime  time.Time          `json:"start_time"`
	EndTime    time.Time          `json:"end_time"`
	DataPoints []*HealthDataPoint `json:"data_points"`
}

// HealthDataPoint represents a single health data point
type HealthDataPoint struct {
	Timestamp           time.Time    `json:"timestamp"`
	Status              HealthStatus `json:"status"`
	HealthyComponents   int          `json:"healthy_components"`
	DegradedComponents  int          `json:"degraded_components"`
	UnhealthyComponents int          `json:"unhealthy_components"`
}

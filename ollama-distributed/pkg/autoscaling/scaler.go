package autoscaling

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AutoScaler manages automatic scaling of the distributed system
type AutoScaler struct {
	config *Config

	// Metrics collection
	metricsCollector MetricsCollector

	// Scaling policies
	policies []ScalingPolicy

	// Current state
	currentReplicas int
	targetReplicas  int
	lastScaleTime   time.Time

	// Scaling executor
	executor ScalingExecutor

	// Statistics
	stats *ScalingStats

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.RWMutex
}

// Config holds auto-scaling configuration
type Config struct {
	// Scaling limits
	MinReplicas int `yaml:"min_replicas"`
	MaxReplicas int `yaml:"max_replicas"`

	// Scaling behavior
	ScaleUpCooldown   time.Duration `yaml:"scale_up_cooldown"`
	ScaleDownCooldown time.Duration `yaml:"scale_down_cooldown"`

	// Evaluation settings
	EvaluationInterval  time.Duration `yaml:"evaluation_interval"`
	StabilizationWindow time.Duration `yaml:"stabilization_window"`

	// Scaling policies
	CPUThreshold    float64 `yaml:"cpu_threshold"`
	MemoryThreshold float64 `yaml:"memory_threshold"`
	QueueThreshold  int     `yaml:"queue_threshold"`

	// Advanced settings
	ScaleUpPercent   int  `yaml:"scale_up_percent"`
	ScaleDownPercent int  `yaml:"scale_down_percent"`
	EnablePredictive bool `yaml:"enable_predictive"`
}

// DefaultConfig returns default auto-scaling configuration
func DefaultConfig() *Config {
	return &Config{
		MinReplicas:         1,
		MaxReplicas:         10,
		ScaleUpCooldown:     3 * time.Minute,
		ScaleDownCooldown:   5 * time.Minute,
		EvaluationInterval:  30 * time.Second,
		StabilizationWindow: 5 * time.Minute,
		CPUThreshold:        70.0,
		MemoryThreshold:     80.0,
		QueueThreshold:      100,
		ScaleUpPercent:      50,
		ScaleDownPercent:    25,
		EnablePredictive:    false,
	}
}

// ScalingStats holds auto-scaling statistics
type ScalingStats struct {
	// Scaling events
	ScaleUpEvents   int64 `json:"scale_up_events"`
	ScaleDownEvents int64 `json:"scale_down_events"`

	// Current state
	CurrentReplicas int       `json:"current_replicas"`
	TargetReplicas  int       `json:"target_replicas"`
	LastScaleTime   time.Time `json:"last_scale_time"`

	// Performance metrics
	AverageResponseTime time.Duration `json:"average_response_time"`
	ThroughputPerSecond float64       `json:"throughput_per_second"`

	// Resource utilization
	CPUUtilization    float64 `json:"cpu_utilization"`
	MemoryUtilization float64 `json:"memory_utilization"`
	QueueSize         int     `json:"queue_size"`

	// Timestamps
	LastEvaluation time.Time `json:"last_evaluation"`
	StartTime      time.Time `json:"start_time"`
}

// MetricsCollector interface for collecting scaling metrics
type MetricsCollector interface {
	GetCPUUtilization() float64
	GetMemoryUtilization() float64
	GetQueueSize() int
	GetResponseTime() time.Duration
	GetThroughput() float64
	GetActiveConnections() int
}

// ScalingExecutor interface for executing scaling actions
type ScalingExecutor interface {
	ScaleUp(replicas int) error
	ScaleDown(replicas int) error
	GetCurrentReplicas() (int, error)
}

// ScalingPolicy defines a scaling policy
type ScalingPolicy interface {
	Evaluate(metrics *Metrics) *ScalingDecision
	Name() string
}

// Metrics holds current system metrics
type Metrics struct {
	CPUUtilization    float64
	MemoryUtilization float64
	QueueSize         int
	ResponseTime      time.Duration
	Throughput        float64
	ActiveConnections int
	Timestamp         time.Time
}

// ScalingDecision represents a scaling decision
type ScalingDecision struct {
	Action         ScalingAction
	TargetReplicas int
	Reason         string
	Confidence     float64
	Priority       int
}

// ScalingAction represents the type of scaling action
type ScalingAction int

const (
	NoAction ScalingAction = iota
	ScaleUp
	ScaleDown
)

// NewAutoScaler creates a new auto-scaler
func NewAutoScaler(config *Config, metricsCollector MetricsCollector, executor ScalingExecutor) *AutoScaler {
	if config == nil {
		config = DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	scaler := &AutoScaler{
		config:           config,
		metricsCollector: metricsCollector,
		executor:         executor,
		stats:            &ScalingStats{StartTime: time.Now()},
		ctx:              ctx,
		cancel:           cancel,
	}

	// Initialize scaling policies
	scaler.initializePolicies()

	return scaler
}

// Start starts the auto-scaler
func (as *AutoScaler) Start() error {
	// Get initial replica count
	replicas, err := as.executor.GetCurrentReplicas()
	if err != nil {
		return fmt.Errorf("failed to get current replicas: %w", err)
	}

	as.mu.Lock()
	as.currentReplicas = replicas
	as.targetReplicas = replicas
	as.stats.CurrentReplicas = replicas
	as.stats.TargetReplicas = replicas
	as.mu.Unlock()

	// Start evaluation loop
	as.wg.Add(1)
	go as.runEvaluationLoop()

	fmt.Printf("Auto-scaler started with %d replicas\n", replicas)
	return nil
}

// Stop stops the auto-scaler
func (as *AutoScaler) Stop() error {
	as.cancel()
	as.wg.Wait()
	return nil
}

// initializePolicies initializes default scaling policies
func (as *AutoScaler) initializePolicies() {
	as.policies = []ScalingPolicy{
		NewCPUPolicy(as.config.CPUThreshold),
		NewMemoryPolicy(as.config.MemoryThreshold),
		NewQueuePolicy(as.config.QueueThreshold),
		NewResponseTimePolicy(5 * time.Second),
	}
}

// runEvaluationLoop runs the scaling evaluation loop
func (as *AutoScaler) runEvaluationLoop() {
	defer as.wg.Done()

	ticker := time.NewTicker(as.config.EvaluationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-as.ctx.Done():
			return
		case <-ticker.C:
			as.evaluate()
		}
	}
}

// evaluate evaluates scaling policies and makes scaling decisions
func (as *AutoScaler) evaluate() {
	// Collect current metrics
	metrics := &Metrics{
		CPUUtilization:    as.metricsCollector.GetCPUUtilization(),
		MemoryUtilization: as.metricsCollector.GetMemoryUtilization(),
		QueueSize:         as.metricsCollector.GetQueueSize(),
		ResponseTime:      as.metricsCollector.GetResponseTime(),
		Throughput:        as.metricsCollector.GetThroughput(),
		ActiveConnections: as.metricsCollector.GetActiveConnections(),
		Timestamp:         time.Now(),
	}

	// Update statistics
	as.updateStats(metrics)

	// Evaluate all policies
	decisions := make([]*ScalingDecision, 0, len(as.policies))
	for _, policy := range as.policies {
		if decision := policy.Evaluate(metrics); decision != nil {
			decisions = append(decisions, decision)
		}
	}

	// Make final scaling decision
	finalDecision := as.makeFinalDecision(decisions)
	if finalDecision != nil {
		as.executeScalingDecision(finalDecision)
	}
}

// makeFinalDecision combines multiple policy decisions into a final decision
func (as *AutoScaler) makeFinalDecision(decisions []*ScalingDecision) *ScalingDecision {
	if len(decisions) == 0 {
		return nil
	}

	// Check cooldown periods
	as.mu.RLock()
	timeSinceLastScale := time.Since(as.lastScaleTime)
	as.mu.RUnlock()

	// Find the highest priority decision
	var finalDecision *ScalingDecision
	for _, decision := range decisions {
		if finalDecision == nil || decision.Priority > finalDecision.Priority {
			finalDecision = decision
		}
	}

	// Apply cooldown logic
	if finalDecision.Action == ScaleUp && timeSinceLastScale < as.config.ScaleUpCooldown {
		return nil
	}

	if finalDecision.Action == ScaleDown && timeSinceLastScale < as.config.ScaleDownCooldown {
		return nil
	}

	// Apply scaling limits
	as.mu.RLock()
	currentReplicas := as.currentReplicas
	as.mu.RUnlock()

	if finalDecision.TargetReplicas < as.config.MinReplicas {
		finalDecision.TargetReplicas = as.config.MinReplicas
	}

	if finalDecision.TargetReplicas > as.config.MaxReplicas {
		finalDecision.TargetReplicas = as.config.MaxReplicas
	}

	// Don't scale if already at target
	if finalDecision.TargetReplicas == currentReplicas {
		return nil
	}

	return finalDecision
}

// executeScalingDecision executes a scaling decision
func (as *AutoScaler) executeScalingDecision(decision *ScalingDecision) {
	as.mu.Lock()
	currentReplicas := as.currentReplicas
	as.mu.Unlock()

	var err error

	switch decision.Action {
	case ScaleUp:
		err = as.executor.ScaleUp(decision.TargetReplicas)
		if err == nil {
			as.mu.Lock()
			as.stats.ScaleUpEvents++
			as.mu.Unlock()
			fmt.Printf("Scaled up from %d to %d replicas: %s\n",
				currentReplicas, decision.TargetReplicas, decision.Reason)
		}

	case ScaleDown:
		err = as.executor.ScaleDown(decision.TargetReplicas)
		if err == nil {
			as.mu.Lock()
			as.stats.ScaleDownEvents++
			as.mu.Unlock()
			fmt.Printf("Scaled down from %d to %d replicas: %s\n",
				currentReplicas, decision.TargetReplicas, decision.Reason)
		}
	}

	if err != nil {
		fmt.Printf("Scaling failed: %v\n", err)
		return
	}

	// Update state
	as.mu.Lock()
	as.currentReplicas = decision.TargetReplicas
	as.targetReplicas = decision.TargetReplicas
	as.lastScaleTime = time.Now()
	as.stats.CurrentReplicas = decision.TargetReplicas
	as.stats.TargetReplicas = decision.TargetReplicas
	as.stats.LastScaleTime = as.lastScaleTime
	as.mu.Unlock()
}

// updateStats updates scaling statistics
func (as *AutoScaler) updateStats(metrics *Metrics) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.stats.CPUUtilization = metrics.CPUUtilization
	as.stats.MemoryUtilization = metrics.MemoryUtilization
	as.stats.QueueSize = metrics.QueueSize
	as.stats.AverageResponseTime = metrics.ResponseTime
	as.stats.ThroughputPerSecond = metrics.Throughput
	as.stats.LastEvaluation = metrics.Timestamp
}

// GetStats returns current scaling statistics
func (as *AutoScaler) GetStats() ScalingStats {
	as.mu.RLock()
	defer as.mu.RUnlock()

	return *as.stats
}

// SetMinReplicas updates the minimum replica count
func (as *AutoScaler) SetMinReplicas(min int) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.config.MinReplicas = min
}

// SetMaxReplicas updates the maximum replica count
func (as *AutoScaler) SetMaxReplicas(max int) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.config.MaxReplicas = max
}

// ForceScale forces scaling to a specific replica count
func (as *AutoScaler) ForceScale(replicas int) error {
	if replicas < as.config.MinReplicas || replicas > as.config.MaxReplicas {
		return fmt.Errorf("replica count %d outside allowed range [%d, %d]",
			replicas, as.config.MinReplicas, as.config.MaxReplicas)
	}

	as.mu.RLock()
	currentReplicas := as.currentReplicas
	as.mu.RUnlock()

	if replicas == currentReplicas {
		return nil
	}

	var err error
	if replicas > currentReplicas {
		err = as.executor.ScaleUp(replicas)
	} else {
		err = as.executor.ScaleDown(replicas)
	}

	if err != nil {
		return err
	}

	// Update state
	as.mu.Lock()
	as.currentReplicas = replicas
	as.targetReplicas = replicas
	as.lastScaleTime = time.Now()
	as.stats.CurrentReplicas = replicas
	as.stats.TargetReplicas = replicas
	as.stats.LastScaleTime = as.lastScaleTime
	as.mu.Unlock()

	fmt.Printf("Force scaled to %d replicas\n", replicas)
	return nil
}

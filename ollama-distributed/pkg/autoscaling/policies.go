package autoscaling

import (
	"fmt"
	"time"
)

// CPUPolicy implements CPU-based scaling policy
type CPUPolicy struct {
	threshold float64
}

// NewCPUPolicy creates a new CPU-based scaling policy
func NewCPUPolicy(threshold float64) *CPUPolicy {
	return &CPUPolicy{
		threshold: threshold,
	}
}

// Name returns the policy name
func (p *CPUPolicy) Name() string {
	return "CPU"
}

// Evaluate evaluates the CPU policy
func (p *CPUPolicy) Evaluate(metrics *Metrics) *ScalingDecision {
	if metrics.CPUUtilization > p.threshold {
		// Scale up if CPU usage is high
		targetReplicas := int(float64(1) * (metrics.CPUUtilization / p.threshold))
		if targetReplicas < 1 {
			targetReplicas = 1
		}

		return &ScalingDecision{
			Action:         ScaleUp,
			TargetReplicas: targetReplicas + 1, // Add one more replica
			Reason:         fmt.Sprintf("CPU utilization %.1f%% > %.1f%%", metrics.CPUUtilization, p.threshold),
			Confidence:     0.8,
			Priority:       2,
		}
	}

	if metrics.CPUUtilization < p.threshold*0.5 {
		// Scale down if CPU usage is very low
		return &ScalingDecision{
			Action:         ScaleDown,
			TargetReplicas: 1, // Scale down by one replica
			Reason:         fmt.Sprintf("CPU utilization %.1f%% < %.1f%%", metrics.CPUUtilization, p.threshold*0.5),
			Confidence:     0.6,
			Priority:       1,
		}
	}

	return nil
}

// MemoryPolicy implements memory-based scaling policy
type MemoryPolicy struct {
	threshold float64
}

// NewMemoryPolicy creates a new memory-based scaling policy
func NewMemoryPolicy(threshold float64) *MemoryPolicy {
	return &MemoryPolicy{
		threshold: threshold,
	}
}

// Name returns the policy name
func (p *MemoryPolicy) Name() string {
	return "Memory"
}

// Evaluate evaluates the memory policy
func (p *MemoryPolicy) Evaluate(metrics *Metrics) *ScalingDecision {
	if metrics.MemoryUtilization > p.threshold {
		// Scale up if memory usage is high
		targetReplicas := int(float64(1) * (metrics.MemoryUtilization / p.threshold))
		if targetReplicas < 1 {
			targetReplicas = 1
		}

		return &ScalingDecision{
			Action:         ScaleUp,
			TargetReplicas: targetReplicas + 1,
			Reason:         fmt.Sprintf("Memory utilization %.1f%% > %.1f%%", metrics.MemoryUtilization, p.threshold),
			Confidence:     0.9,
			Priority:       3, // Higher priority than CPU
		}
	}

	if metrics.MemoryUtilization < p.threshold*0.4 {
		// Scale down if memory usage is very low
		return &ScalingDecision{
			Action:         ScaleDown,
			TargetReplicas: 1,
			Reason:         fmt.Sprintf("Memory utilization %.1f%% < %.1f%%", metrics.MemoryUtilization, p.threshold*0.4),
			Confidence:     0.7,
			Priority:       2,
		}
	}

	return nil
}

// QueuePolicy implements queue-based scaling policy
type QueuePolicy struct {
	threshold int
}

// NewQueuePolicy creates a new queue-based scaling policy
func NewQueuePolicy(threshold int) *QueuePolicy {
	return &QueuePolicy{
		threshold: threshold,
	}
}

// Name returns the policy name
func (p *QueuePolicy) Name() string {
	return "Queue"
}

// Evaluate evaluates the queue policy
func (p *QueuePolicy) Evaluate(metrics *Metrics) *ScalingDecision {
	if metrics.QueueSize > p.threshold {
		// Scale up if queue is too large
		// Calculate target replicas based on queue size
		targetReplicas := (metrics.QueueSize / p.threshold) + 1

		return &ScalingDecision{
			Action:         ScaleUp,
			TargetReplicas: targetReplicas,
			Reason:         fmt.Sprintf("Queue size %d > %d", metrics.QueueSize, p.threshold),
			Confidence:     0.95,
			Priority:       4, // Highest priority - queue backlog is critical
		}
	}

	if metrics.QueueSize < p.threshold/4 {
		// Scale down if queue is very small
		return &ScalingDecision{
			Action:         ScaleDown,
			TargetReplicas: 1,
			Reason:         fmt.Sprintf("Queue size %d < %d", metrics.QueueSize, p.threshold/4),
			Confidence:     0.5,
			Priority:       1,
		}
	}

	return nil
}

// ResponseTimePolicy implements response time-based scaling policy
type ResponseTimePolicy struct {
	threshold time.Duration
}

// NewResponseTimePolicy creates a new response time-based scaling policy
func NewResponseTimePolicy(threshold time.Duration) *ResponseTimePolicy {
	return &ResponseTimePolicy{
		threshold: threshold,
	}
}

// Name returns the policy name
func (p *ResponseTimePolicy) Name() string {
	return "ResponseTime"
}

// Evaluate evaluates the response time policy
func (p *ResponseTimePolicy) Evaluate(metrics *Metrics) *ScalingDecision {
	if metrics.ResponseTime > p.threshold {
		// Scale up if response time is too high
		multiplier := float64(metrics.ResponseTime) / float64(p.threshold)
		targetReplicas := int(multiplier) + 1

		return &ScalingDecision{
			Action:         ScaleUp,
			TargetReplicas: targetReplicas,
			Reason:         fmt.Sprintf("Response time %v > %v", metrics.ResponseTime, p.threshold),
			Confidence:     0.85,
			Priority:       3,
		}
	}

	if metrics.ResponseTime < p.threshold/3 {
		// Scale down if response time is very low
		return &ScalingDecision{
			Action:         ScaleDown,
			TargetReplicas: 1,
			Reason:         fmt.Sprintf("Response time %v < %v", metrics.ResponseTime, p.threshold/3),
			Confidence:     0.4,
			Priority:       1,
		}
	}

	return nil
}

// ThroughputPolicy implements throughput-based scaling policy
type ThroughputPolicy struct {
	minThroughput float64
	maxThroughput float64
}

// NewThroughputPolicy creates a new throughput-based scaling policy
func NewThroughputPolicy(minThroughput, maxThroughput float64) *ThroughputPolicy {
	return &ThroughputPolicy{
		minThroughput: minThroughput,
		maxThroughput: maxThroughput,
	}
}

// Name returns the policy name
func (p *ThroughputPolicy) Name() string {
	return "Throughput"
}

// Evaluate evaluates the throughput policy
func (p *ThroughputPolicy) Evaluate(metrics *Metrics) *ScalingDecision {
	if metrics.Throughput > p.maxThroughput {
		// Scale up if throughput is too high (system is overloaded)
		targetReplicas := int(metrics.Throughput/p.maxThroughput) + 1

		return &ScalingDecision{
			Action:         ScaleUp,
			TargetReplicas: targetReplicas,
			Reason:         fmt.Sprintf("Throughput %.1f > %.1f", metrics.Throughput, p.maxThroughput),
			Confidence:     0.7,
			Priority:       2,
		}
	}

	if metrics.Throughput < p.minThroughput {
		// Scale down if throughput is too low
		return &ScalingDecision{
			Action:         ScaleDown,
			TargetReplicas: 1,
			Reason:         fmt.Sprintf("Throughput %.1f < %.1f", metrics.Throughput, p.minThroughput),
			Confidence:     0.6,
			Priority:       1,
		}
	}

	return nil
}

// CompositePolicy combines multiple policies with weights
type CompositePolicy struct {
	policies []WeightedPolicy
	name     string
}

// WeightedPolicy represents a policy with a weight
type WeightedPolicy struct {
	Policy ScalingPolicy
	Weight float64
}

// NewCompositePolicy creates a new composite policy
func NewCompositePolicy(name string, policies []WeightedPolicy) *CompositePolicy {
	return &CompositePolicy{
		policies: policies,
		name:     name,
	}
}

// Name returns the policy name
func (p *CompositePolicy) Name() string {
	return p.name
}

// Evaluate evaluates the composite policy
func (p *CompositePolicy) Evaluate(metrics *Metrics) *ScalingDecision {
	var scaleUpScore, scaleDownScore float64
	var scaleUpReasons, scaleDownReasons []string
	var maxTargetReplicas int

	// Evaluate all sub-policies
	for _, wp := range p.policies {
		decision := wp.Policy.Evaluate(metrics)
		if decision == nil {
			continue
		}

		switch decision.Action {
		case ScaleUp:
			scaleUpScore += decision.Confidence * wp.Weight
			scaleUpReasons = append(scaleUpReasons, fmt.Sprintf("%s: %s", wp.Policy.Name(), decision.Reason))
			if decision.TargetReplicas > maxTargetReplicas {
				maxTargetReplicas = decision.TargetReplicas
			}

		case ScaleDown:
			scaleDownScore += decision.Confidence * wp.Weight
			scaleDownReasons = append(scaleDownReasons, fmt.Sprintf("%s: %s", wp.Policy.Name(), decision.Reason))
		}
	}

	// Make final decision based on scores
	if scaleUpScore > scaleDownScore && scaleUpScore > 0.5 {
		return &ScalingDecision{
			Action:         ScaleUp,
			TargetReplicas: maxTargetReplicas,
			Reason:         fmt.Sprintf("Composite decision (score: %.2f): %v", scaleUpScore, scaleUpReasons),
			Confidence:     scaleUpScore,
			Priority:       5, // Highest priority for composite decisions
		}
	}

	if scaleDownScore > scaleUpScore && scaleDownScore > 0.3 {
		return &ScalingDecision{
			Action:         ScaleDown,
			TargetReplicas: 1,
			Reason:         fmt.Sprintf("Composite decision (score: %.2f): %v", scaleDownScore, scaleDownReasons),
			Confidence:     scaleDownScore,
			Priority:       3,
		}
	}

	return nil
}

// PredictivePolicy implements predictive scaling based on historical data
type PredictivePolicy struct {
	history    []MetricsSnapshot
	maxHistory int
	threshold  float64
}

// MetricsSnapshot represents a point-in-time metrics snapshot
type MetricsSnapshot struct {
	Metrics   *Metrics
	Timestamp time.Time
}

// NewPredictivePolicy creates a new predictive scaling policy
func NewPredictivePolicy(maxHistory int, threshold float64) *PredictivePolicy {
	return &PredictivePolicy{
		history:    make([]MetricsSnapshot, 0, maxHistory),
		maxHistory: maxHistory,
		threshold:  threshold,
	}
}

// Name returns the policy name
func (p *PredictivePolicy) Name() string {
	return "Predictive"
}

// Evaluate evaluates the predictive policy
func (p *PredictivePolicy) Evaluate(metrics *Metrics) *ScalingDecision {
	// Add current metrics to history
	p.addToHistory(metrics)

	if len(p.history) < 10 {
		// Not enough data for prediction
		return nil
	}

	// Simple trend analysis
	trend := p.calculateTrend()

	if trend > p.threshold {
		// Predict scale up needed
		return &ScalingDecision{
			Action:         ScaleUp,
			TargetReplicas: 2, // Conservative prediction
			Reason:         fmt.Sprintf("Predictive scaling: trend %.2f > %.2f", trend, p.threshold),
			Confidence:     0.6,
			Priority:       1, // Lower priority for predictions
		}
	}

	if trend < -p.threshold {
		// Predict scale down possible
		return &ScalingDecision{
			Action:         ScaleDown,
			TargetReplicas: 1,
			Reason:         fmt.Sprintf("Predictive scaling: trend %.2f < %.2f", trend, -p.threshold),
			Confidence:     0.4,
			Priority:       1,
		}
	}

	return nil
}

// addToHistory adds a metrics snapshot to the history
func (p *PredictivePolicy) addToHistory(metrics *Metrics) {
	snapshot := MetricsSnapshot{
		Metrics:   metrics,
		Timestamp: time.Now(),
	}

	p.history = append(p.history, snapshot)

	// Keep only the most recent entries
	if len(p.history) > p.maxHistory {
		p.history = p.history[1:]
	}
}

// calculateTrend calculates the trend in resource utilization
func (p *PredictivePolicy) calculateTrend() float64 {
	if len(p.history) < 2 {
		return 0
	}

	// Simple linear trend calculation
	n := len(p.history)
	recent := p.history[n-1].Metrics.CPUUtilization
	older := p.history[0].Metrics.CPUUtilization

	return recent - older
}

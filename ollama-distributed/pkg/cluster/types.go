package cluster

import (
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/types"
)

// NodeInfo represents detailed information about a cluster node
type NodeInfo struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Address      string            `json:"address"`
	Region       string            `json:"region"`
	Zone         string            `json:"zone"`
	Status       NodeStatus        `json:"status"`
	Capabilities NodeCapabilities  `json:"capabilities"`
	Resources    ResourceInfo      `json:"resources"`
	Metadata     map[string]string `json:"metadata"`
	LastSeen     time.Time         `json:"last_seen"`
	JoinedAt     time.Time         `json:"joined_at"`
}

// NodeStatus represents the current status of a node
type NodeStatus string

const (
	NodeStatusHealthy     NodeStatus = "healthy"
	NodeStatusDegraded    NodeStatus = "degraded"
	NodeStatusUnhealthy   NodeStatus = "unhealthy"
	NodeStatusUnavailable NodeStatus = "unavailable"
	NodeStatusJoining     NodeStatus = "joining"
	NodeStatusLeaving     NodeStatus = "leaving"
)

// NodeCapabilities defines what a node can do
type NodeCapabilities struct {
	Inference    bool     `json:"inference"`
	Storage      bool     `json:"storage"`
	Coordination bool     `json:"coordination"`
	Gateway      bool     `json:"gateway"`
	Models       []string `json:"models"`
}

// ResourceInfo contains resource usage information
type ResourceInfo struct {
	CPU     ResourceUsage `json:"cpu"`
	Memory  ResourceUsage `json:"memory"`
	GPU     ResourceUsage `json:"gpu"`
	Disk    ResourceUsage `json:"disk"`
	Network NetworkUsage  `json:"network"`
}

// ResourceUsage represents usage of a specific resource
type ResourceUsage struct {
	Used      float64 `json:"used"`
	Available float64 `json:"available"`
	Total     float64 `json:"total"`
	Percent   float64 `json:"percent"`
}

// NetworkUsage represents network usage statistics
type NetworkUsage struct {
	BytesIn    uint64 `json:"bytes_in"`
	BytesOut   uint64 `json:"bytes_out"`
	PacketsIn  uint64 `json:"packets_in"`
	PacketsOut uint64 `json:"packets_out"`
}

// DiscoveryStrategy defines how nodes are discovered
type DiscoveryStrategy interface {
	Discover() ([]*NodeInfo, error)
	GetName() string
}

// HealthCheck represents a health check configuration
type HealthCheck struct {
	Name       string        `json:"name"`
	Endpoint   string        `json:"endpoint"`
	Interval   time.Duration `json:"interval"`
	Timeout    time.Duration `json:"timeout"`
	Retries    int           `json:"retries"`
	Enabled    bool          `json:"enabled"`
	LastResult *HealthResult `json:"last_result"`
}

// HealthResult represents the result of a health check
type HealthResult struct {
	Success    bool          `json:"success"`
	Latency    time.Duration `json:"latency"`
	Error      string        `json:"error,omitempty"`
	Timestamp  time.Time     `json:"timestamp"`
	StatusCode int           `json:"status_code,omitempty"`
	Response   string        `json:"response,omitempty"`
}

// AlertManager handles health-related alerts
type AlertManager struct {
	alerts   []*Alert
	channels []AlertChannel
}

// Alert represents a system alert
type Alert struct {
	ID         string            `json:"id"`
	Type       AlertType         `json:"type"`
	Severity   AlertSeverity     `json:"severity"`
	Title      string            `json:"title"`
	Message    string            `json:"message"`
	NodeID     string            `json:"node_id,omitempty"`
	Metadata   map[string]string `json:"metadata"`
	CreatedAt  time.Time         `json:"created_at"`
	ResolvedAt *time.Time        `json:"resolved_at,omitempty"`
}

// AlertType defines the type of alert
type AlertType string

const (
	AlertTypeNodeDown        AlertType = "node_down"
	AlertTypeHighLatency     AlertType = "high_latency"
	AlertTypeResourceExhaust AlertType = "resource_exhaustion"
	AlertTypeScalingEvent    AlertType = "scaling_event"
	AlertTypePerformance     AlertType = "performance_degradation"
)

// AlertSeverity defines the severity of an alert
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityError    AlertSeverity = "error"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertChannel defines how alerts are delivered
type AlertChannel interface {
	SendAlert(alert *Alert) error
	GetName() string
}

// LoadBalancingStrategy defines how load is balanced
type LoadBalancingStrategy interface {
	SelectNode(nodes []*NodeInfo, request *RequestContext) (*NodeInfo, error)
	GetName() string
}

// LoadMetrics tracks load for a specific node
type LoadMetrics struct {
	NodeID            string    `json:"node_id"`
	RequestsPerSecond float64   `json:"requests_per_second"`
	AverageLatency    float64   `json:"average_latency"`
	ErrorRate         float64   `json:"error_rate"`
	CPUUtilization    float64   `json:"cpu_utilization"`
	MemoryUtilization float64   `json:"memory_utilization"`
	ActiveConnections int       `json:"active_connections"`
	QueueLength       int       `json:"queue_length"`
	LastUpdated       time.Time `json:"last_updated"`
}

// RequestContext provides context for load balancing decisions
type RequestContext struct {
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Headers   map[string]string `json:"headers"`
	ModelName string            `json:"model_name,omitempty"`
	Priority  int               `json:"priority"`
	Timeout   time.Duration     `json:"timeout"`
	Metadata  map[string]string `json:"metadata"`
}

// ScalingPolicy defines when and how to scale
type ScalingPolicy struct {
	Name     string           `json:"name"`
	Enabled  bool             `json:"enabled"`
	Triggers []ScalingTrigger `json:"triggers"`
	Actions  []ScalingAction  `json:"actions"`
	Cooldown time.Duration    `json:"cooldown"`
	MinNodes int              `json:"min_nodes"`
	MaxNodes int              `json:"max_nodes"`
}

// ScalingTrigger defines conditions that trigger scaling
type ScalingTrigger struct {
	Metric    string        `json:"metric"`
	Operator  string        `json:"operator"` // >, <, >=, <=, ==
	Threshold float64       `json:"threshold"`
	Duration  time.Duration `json:"duration"`
}

// ScalingAction defines what action to take when scaling
type ScalingAction struct {
	Type     ScalingActionType `json:"type"`
	Count    int               `json:"count"`
	NodeType string            `json:"node_type,omitempty"`
	Region   string            `json:"region,omitempty"`
	Zone     string            `json:"zone,omitempty"`
}

// ScalingActionType defines the type of scaling action
type ScalingActionType string

const (
	ScalingActionScaleUp   ScalingActionType = "scale_up"
	ScalingActionScaleDown ScalingActionType = "scale_down"
)

// PerformanceMetrics contains current performance data
type PerformanceMetrics struct {
	Timestamp          time.Time `json:"timestamp"`
	TotalRequests      uint64    `json:"total_requests"`
	RequestsPerSecond  float64   `json:"requests_per_second"`
	AverageLatency     float64   `json:"average_latency"`
	P95Latency         float64   `json:"p95_latency"`
	P99Latency         float64   `json:"p99_latency"`
	ErrorRate          float64   `json:"error_rate"`
	ThroughputMBps     float64   `json:"throughput_mbps"`
	ActiveConnections  int       `json:"active_connections"`
	ClusterUtilization float64   `json:"cluster_utilization"`
}

// PerformanceHistory stores historical performance data
type PerformanceHistory struct {
	Metrics    []*PerformanceMetrics `json:"metrics"`
	MaxEntries int                   `json:"max_entries"`
}

// PredictionModel represents a machine learning model for predictions
type PredictionModel struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Accuracy    float64   `json:"accuracy"`
	LastTrained time.Time `json:"last_trained"`
	Features    []string  `json:"features"`
}

// ScalingPrediction represents a scaling prediction
type ScalingPrediction struct {
	Timestamp        time.Time     `json:"timestamp"`
	PredictedLoad    float64       `json:"predicted_load"`
	RecommendedNodes int           `json:"recommended_nodes"`
	Confidence       float64       `json:"confidence"`
	Horizon          time.Duration `json:"horizon"`
	Reasoning        string        `json:"reasoning"`
}

// RegionInfo contains information about a region
type RegionInfo struct {
	Name        string             `json:"name"`
	Nodes       []*NodeInfo        `json:"nodes"`
	Status      RegionStatus       `json:"status"`
	Latency     map[string]float64 `json:"latency"` // Latency to other regions
	Capacity    ResourceInfo       `json:"capacity"`
	Utilization ResourceInfo       `json:"utilization"`
}

// RegionStatus represents the status of a region
type RegionStatus string

const (
	RegionStatusHealthy  RegionStatus = "healthy"
	RegionStatusDegraded RegionStatus = "degraded"
	RegionStatusIsolated RegionStatus = "isolated"
)

// ReplicationState tracks cross-region replication
type ReplicationState struct {
	SourceRegion string        `json:"source_region"`
	TargetRegion string        `json:"target_region"`
	Status       string        `json:"status"`
	Progress     float64       `json:"progress"`
	LastSync     time.Time     `json:"last_sync"`
	Lag          time.Duration `json:"lag"`
}

// EnhancedClusterStatus provides comprehensive cluster status
type EnhancedClusterStatus struct {
	BasicStatus        *types.ClusterState           `json:"basic_status"`
	NodeHealth         map[string]float64            `json:"node_health"`
	LoadDistribution   map[string]*LoadMetrics       `json:"load_distribution"`
	PerformanceMetrics *PerformanceMetrics           `json:"performance_metrics"`
	ScalingState       *ScalingState                 `json:"scaling_state"`
	RegionStatus       map[string]*RegionInfo        `json:"region_status"`
	Predictions        map[string]*ScalingPrediction `json:"predictions"`
}

// ScalingState represents the current scaling state
type ScalingState struct {
	CurrentNodes      int       `json:"current_nodes"`
	TargetNodes       int       `json:"target_nodes"`
	LastScaleAction   time.Time `json:"last_scale_action"`
	ScalingInProgress bool      `json:"scaling_in_progress"`
	CooldownUntil     time.Time `json:"cooldown_until"`
}

// PerformanceInsights provides performance analysis
type PerformanceInsights struct {
	OverallHealth      float64            `json:"overall_health"`
	Bottlenecks        []string           `json:"bottlenecks"`
	Recommendations    []string           `json:"recommendations"`
	TrendAnalysis      *TrendAnalysis     `json:"trend_analysis"`
	ResourceEfficiency map[string]float64 `json:"resource_efficiency"`
	PredictedIssues    []*PredictedIssue  `json:"predicted_issues"`
}

// TrendAnalysis provides trend analysis of performance metrics
type TrendAnalysis struct {
	LatencyTrend     string  `json:"latency_trend"` // "improving", "stable", "degrading"
	ThroughputTrend  string  `json:"throughput_trend"`
	ErrorRateTrend   string  `json:"error_rate_trend"`
	UtilizationTrend string  `json:"utilization_trend"`
	Confidence       float64 `json:"confidence"`
}

// PredictedIssue represents a potential future issue
type PredictedIssue struct {
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	ETA         time.Time `json:"eta"`
	Confidence  float64   `json:"confidence"`
	Mitigation  string    `json:"mitigation"`
}

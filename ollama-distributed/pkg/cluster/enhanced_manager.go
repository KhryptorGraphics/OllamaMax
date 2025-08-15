package cluster

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/monitoring"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/distributed"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/types"
)

// EnhancedManager provides advanced cluster management capabilities
type EnhancedManager struct {
	*distributed.ClusterManager // Embed the base cluster manager

	config           *config.DistributedConfig
	logger           *logrus.Logger
	monitoringSystem *monitoring.MonitoringSystem

	// Advanced features
	nodeDiscovery  *NodeDiscovery
	healthMonitor  *HealthMonitor
	loadBalancer   *LoadBalancer
	scalingManager *ScalingManager

	// Performance tracking
	performanceTracker *PerformanceTracker
	predictiveScaler   *PredictiveScaler

	// Cross-region support
	regionManager *RegionManager

	// Metrics
	clusterSizeGauge   prometheus.Gauge
	nodeHealthGauge    *prometheus.GaugeVec
	requestLatencyHist *prometheus.HistogramVec
	throughputCounter  *prometheus.CounterVec

	mu sync.RWMutex
}

// NodeDiscovery handles intelligent node discovery
type NodeDiscovery struct {
	config *config.DistributedConfig
	logger *logrus.Logger
	nodes  map[string]*NodeInfo
	mu     sync.RWMutex

	// Discovery strategies
	strategies []DiscoveryStrategy
}

// HealthMonitor provides comprehensive health monitoring
type HealthMonitor struct {
	config       *config.DistributedConfig
	logger       *logrus.Logger
	healthChecks map[string]*HealthCheck
	alertManager *AlertManager

	// Health metrics
	healthScores map[string]float64
	lastChecked  map[string]time.Time
	mu           sync.RWMutex
}

// LoadBalancer provides intelligent load balancing
type LoadBalancer struct {
	config     *config.DistributedConfig
	logger     *logrus.Logger
	strategies map[string]LoadBalancingStrategy

	// Load tracking
	nodeLoads map[string]*LoadMetrics
	mu        sync.RWMutex
}

// ScalingManager handles automatic scaling decisions
type ScalingManager struct {
	config          *config.DistributedConfig
	logger          *logrus.Logger
	scalingPolicies []*ScalingPolicy

	// Scaling state
	lastScaleAction time.Time
	scalingCooldown time.Duration
	mu              sync.RWMutex
}

// PerformanceTracker monitors cluster performance
type PerformanceTracker struct {
	config  *config.DistributedConfig
	logger  *logrus.Logger
	metrics *PerformanceMetrics

	// Performance history
	history *PerformanceHistory
	mu      sync.RWMutex
}

// PredictiveScaler uses ML for scaling predictions
type PredictiveScaler struct {
	config *config.DistributedConfig
	logger *logrus.Logger
	model  *PredictionModel

	// Prediction state
	predictions map[string]*ScalingPrediction
	mu          sync.RWMutex
}

// RegionManager handles cross-region operations
type RegionManager struct {
	config  *config.DistributedConfig
	logger  *logrus.Logger
	regions map[string]*RegionInfo

	// Cross-region state
	replicationState map[string]*ReplicationState
	mu               sync.RWMutex
}

// NewEnhancedManager creates a new enhanced cluster manager
func NewEnhancedManager(cfg *config.DistributedConfig, baseManager *distributed.ClusterManager, logger *logrus.Logger) (*EnhancedManager, error) {
	em := &EnhancedManager{
		ClusterManager: baseManager,
		config:         cfg,
		logger:         logger,
	}

	// Initialize metrics
	if err := em.initializeMetrics(); err != nil {
		return nil, fmt.Errorf("failed to initialize metrics: %w", err)
	}

	// Initialize components
	if err := em.initializeComponents(); err != nil {
		return nil, fmt.Errorf("failed to initialize components: %w", err)
	}

	return em, nil
}

// initializeMetrics sets up Prometheus metrics
func (em *EnhancedManager) initializeMetrics() error {
	em.clusterSizeGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ollama_cluster_size",
		Help: "Current number of nodes in the cluster",
	})

	em.nodeHealthGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ollama_node_health_score",
		Help: "Health score of cluster nodes",
	}, []string{"node_id", "region", "zone"})

	em.requestLatencyHist = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ollama_request_duration_seconds",
		Help:    "Request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"node_id", "method", "status"})

	em.throughputCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ollama_requests_total",
		Help: "Total number of requests processed",
	}, []string{"node_id", "method", "status"})

	// Register metrics
	prometheus.MustRegister(em.clusterSizeGauge)
	prometheus.MustRegister(em.nodeHealthGauge)
	prometheus.MustRegister(em.requestLatencyHist)
	prometheus.MustRegister(em.throughputCounter)

	return nil
}

// initializeComponents initializes all enhanced components
func (em *EnhancedManager) initializeComponents() error {

	// Initialize node discovery
	em.nodeDiscovery = &NodeDiscovery{
		config: em.config,
		logger: em.logger,
		nodes:  make(map[string]*NodeInfo),
		strategies: []DiscoveryStrategy{
			&MDNSDiscovery{config: em.config, logger: em.logger},
			&P2PDiscovery{config: em.config, logger: em.logger},
		},
	}

	// Initialize health monitor
	em.healthMonitor = &HealthMonitor{
		config:       em.config,
		logger:       em.logger,
		healthChecks: make(map[string]*HealthCheck),
		alertManager: &AlertManager{
			alerts:   make([]*Alert, 0),
			channels: make([]AlertChannel, 0),
		},
		healthScores: make(map[string]float64),
		lastChecked:  make(map[string]time.Time),
	}

	// Initialize load balancer
	em.loadBalancer = &LoadBalancer{
		config: em.config,
		logger: em.logger,
		strategies: map[string]LoadBalancingStrategy{
			"round_robin":  &RoundRobinStrategy{},
			"least_loaded": &LeastLoadedStrategy{},
			"affinity":     &AffinityStrategy{},
		},
		nodeLoads: make(map[string]*LoadMetrics),
	}

	// Initialize scaling manager
	em.scalingManager = &ScalingManager{
		config:          em.config,
		logger:          em.logger,
		scalingPolicies: make([]*ScalingPolicy, 0),
		scalingCooldown: 10 * time.Minute,
	}

	// Initialize performance tracker
	em.performanceTracker = &PerformanceTracker{
		config: em.config,
		logger: em.logger,
		metrics: &PerformanceMetrics{
			Timestamp: time.Now(),
		},
		history: &PerformanceHistory{
			Metrics:    make([]*PerformanceMetrics, 0),
			MaxEntries: 1000,
		},
	}

	// Initialize predictive scaler
	em.predictiveScaler = &PredictiveScaler{
		config: em.config,
		logger: em.logger,
		model: &PredictionModel{
			Name:     "linear_regression",
			Version:  "1.0",
			Accuracy: 0.85,
		},
		predictions: make(map[string]*ScalingPrediction),
	}

	// Initialize region manager
	em.regionManager = &RegionManager{
		config:           em.config,
		logger:           em.logger,
		regions:          make(map[string]*RegionInfo),
		replicationState: make(map[string]*ReplicationState),
	}

	return nil
}

// Start begins all enhanced cluster management operations
func (em *EnhancedManager) Start(ctx context.Context) error {
	em.logger.Info("Starting enhanced cluster manager")

	// Start base manager
	if err := em.ClusterManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start base manager: %w", err)
	}

	// Start enhanced components
	go em.nodeDiscovery.Start(ctx)
	go em.healthMonitor.Start(ctx)
	go em.loadBalancer.Start(ctx)
	go em.scalingManager.Start(ctx)
	go em.performanceTracker.Start(ctx)
	go em.predictiveScaler.Start(ctx)
	go em.regionManager.Start(ctx)

	// Start metrics collection
	go em.collectMetrics(ctx)

	em.logger.Info("Enhanced cluster manager started successfully")
	return nil
}

// collectMetrics periodically collects and updates metrics
func (em *EnhancedManager) collectMetrics(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			em.updateMetrics()
		}
	}
}

// updateMetrics updates all Prometheus metrics
func (em *EnhancedManager) updateMetrics() {
	em.mu.RLock()
	defer em.mu.RUnlock()

	// Update cluster size
	clusterSize := float64(len(em.GetNodes()))
	em.clusterSizeGauge.Set(clusterSize)

	// Update node health scores
	for nodeID, node := range em.GetNodes() {
		healthScore := em.healthMonitor.GetHealthScore(nodeID)
		em.nodeHealthGauge.WithLabelValues(
			nodeID,
			node.Region,
			node.Zone,
		).Set(healthScore)
	}
}

// GetEnhancedClusterStatus returns comprehensive cluster status
func (em *EnhancedManager) GetEnhancedClusterStatus() *EnhancedClusterStatus {
	em.mu.RLock()
	defer em.mu.RUnlock()

	// Get basic status and convert to ClusterState if possible
	basicStatus := em.GetStatus()
	var clusterState *types.ClusterState
	if cs, ok := basicStatus.(*types.ClusterState); ok {
		clusterState = cs
	} else {
		// Create a basic ClusterState from the interface
		clusterState = &types.ClusterState{
			Status:      types.ClusterStatusHealthy,
			LastUpdated: time.Now(),
			Metadata:    make(map[string]interface{}),
		}
		if statusMap, ok := basicStatus.(map[string]interface{}); ok {
			clusterState.Metadata = statusMap
		}
	}

	return &EnhancedClusterStatus{
		BasicStatus:        clusterState,
		NodeHealth:         em.healthMonitor.GetAllHealthScores(),
		LoadDistribution:   em.loadBalancer.GetLoadDistribution(),
		PerformanceMetrics: em.performanceTracker.GetCurrentMetrics(),
		ScalingState:       em.scalingManager.GetScalingState(),
		RegionStatus:       em.regionManager.GetRegionStatus(),
		Predictions:        em.predictiveScaler.GetPredictions(),
	}
}

// GetOptimalNode returns the best node for a given request
func (em *EnhancedManager) GetOptimalNode(request *RequestContext) (*NodeInfo, error) {
	return em.loadBalancer.SelectNode(request)
}

// TriggerScaling manually triggers scaling evaluation
func (em *EnhancedManager) TriggerScaling() error {
	return em.scalingManager.EvaluateScaling()
}

// GetPerformanceInsights returns performance analysis
func (em *EnhancedManager) GetPerformanceInsights() *PerformanceInsights {
	return em.performanceTracker.GetInsights()
}

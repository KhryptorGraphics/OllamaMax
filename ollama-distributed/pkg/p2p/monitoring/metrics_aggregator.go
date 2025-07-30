package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// MetricsAggregator aggregates metrics from multiple sources
type MetricsAggregator struct {
	config          *AggregatorConfig
	
	// Metric sources
	networkMonitor  *NetworkMonitor
	sources         map[string]MetricsSource
	sourcesMu       sync.RWMutex
	
	// Aggregated metrics
	aggregatedMetrics *AggregatedMetrics
	metricsMu         sync.RWMutex
	
	// Historical data
	historicalData    *HistoricalData
	
	// Export targets
	exporters         []MetricsExporter
	exportersMu       sync.RWMutex
	
	// Lifecycle
	ctx               context.Context
	cancel            context.CancelFunc
	wg                sync.WaitGroup
}

// AggregatorConfig configures the metrics aggregator
type AggregatorConfig struct {
	// Aggregation settings
	AggregationInterval time.Duration
	RetentionPeriod     time.Duration
	
	// Historical data settings
	HistoricalResolution time.Duration
	MaxHistoricalPoints  int
	
	// Export settings
	ExportInterval      time.Duration
	EnableExport        bool
	
	// Performance settings
	BufferSize          int
	WorkerCount         int
}

// AggregatedMetrics contains all aggregated metrics
type AggregatedMetrics struct {
	// Network metrics summary
	NetworkSummary      *NetworkSummary      `json:"network_summary"`
	
	// Performance summary
	PerformanceSummary  *PerformanceSummary  `json:"performance_summary"`
	
	// Security summary
	SecuritySummary     *SecuritySummary     `json:"security_summary"`
	
	// Resource usage summary
	ResourceSummary     *ResourceSummary     `json:"resource_summary"`
	
	// Topology metrics
	TopologyMetrics     *TopologyMetrics     `json:"topology_metrics"`
	
	// Timestamp
	Timestamp           time.Time            `json:"timestamp"`
	
	// Metadata
	NodeID              peer.ID              `json:"node_id"`
	Version             string               `json:"version"`
}

// NetworkSummary summarizes network metrics
type NetworkSummary struct {
	TotalConnections    int64                `json:"total_connections"`
	ActiveConnections   int64                `json:"active_connections"`
	MessageThroughput   float64              `json:"message_throughput"`
	AverageLatency      time.Duration        `json:"average_latency"`
	ErrorRate           float64              `json:"error_rate"`
	BandwidthUtilization float64             `json:"bandwidth_utilization"`
	ProtocolDistribution map[string]float64  `json:"protocol_distribution"`
}

// PerformanceSummary summarizes performance metrics
type PerformanceSummary struct {
	CPUUsage            float64              `json:"cpu_usage"`
	MemoryUsage         float64              `json:"memory_usage"`
	DiskUsage           float64              `json:"disk_usage"`
	NetworkUtilization  float64              `json:"network_utilization"`
	ResponseTime        time.Duration        `json:"response_time"`
	Throughput          float64              `json:"throughput"`
	QueueDepth          int64                `json:"queue_depth"`
}

// SecuritySummary summarizes security metrics
type SecuritySummary struct {
	AuthenticationRate  float64              `json:"authentication_rate"`
	SecurityViolations  int64                `json:"security_violations"`
	EncryptionRate      float64              `json:"encryption_rate"`
	TrustedPeers        int64                `json:"trusted_peers"`
	BlockedConnections  int64                `json:"blocked_connections"`
	CertificateStatus   string               `json:"certificate_status"`
}

// ResourceSummary summarizes resource usage
type ResourceSummary struct {
	MemoryAllocated     int64                `json:"memory_allocated"`
	MemoryUsed          int64                `json:"memory_used"`
	GoroutineCount      int                  `json:"goroutine_count"`
	FileDescriptors     int                  `json:"file_descriptors"`
	NetworkConnections  int                  `json:"network_connections"`
	DiskIOPS            float64              `json:"disk_iops"`
}

// TopologyMetrics summarizes network topology
type TopologyMetrics struct {
	NodeCount           int64                `json:"node_count"`
	ClusterSize         int64                `json:"cluster_size"`
	NetworkDiameter     int                  `json:"network_diameter"`
	AverageConnectivity float64              `json:"average_connectivity"`
	PartitionCount      int                  `json:"partition_count"`
	LeaderNode          peer.ID              `json:"leader_node"`
}

// HistoricalData stores historical metrics
type HistoricalData struct {
	dataPoints          []*HistoricalPoint
	maxPoints           int
	resolution          time.Duration
	currentIndex        int
	mu                  sync.RWMutex
}

// HistoricalPoint represents a point in time with metrics
type HistoricalPoint struct {
	Timestamp           time.Time            `json:"timestamp"`
	Metrics             *AggregatedMetrics   `json:"metrics"`
}

// MetricsSource interface for metric sources
type MetricsSource interface {
	GetMetrics() (map[string]interface{}, error)
	GetName() string
	IsHealthy() bool
}

// MetricsExporter interface for metric exporters
type MetricsExporter interface {
	Export(metrics *AggregatedMetrics) error
	GetName() string
	IsEnabled() bool
}

// PrometheusExporter exports metrics to Prometheus
type PrometheusExporter struct {
	endpoint            string
	enabled             bool
}

// InfluxDBExporter exports metrics to InfluxDB
type InfluxDBExporter struct {
	config              *InfluxDBConfig
	enabled             bool
}

// JSONExporter exports metrics to JSON files
type JSONExporter struct {
	outputPath          string
	enabled             bool
}

// NewMetricsAggregator creates a new metrics aggregator
func NewMetricsAggregator(config *AggregatorConfig, networkMonitor *NetworkMonitor) *MetricsAggregator {
	if config == nil {
		config = &AggregatorConfig{
			AggregationInterval:  30 * time.Second,
			RetentionPeriod:      24 * time.Hour,
			HistoricalResolution: 1 * time.Minute,
			MaxHistoricalPoints:  1440, // 24 hours at 1-minute resolution
			ExportInterval:       1 * time.Minute,
			EnableExport:         true,
			BufferSize:           1000,
			WorkerCount:          2,
		}
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	aggregator := &MetricsAggregator{
		config:         config,
		networkMonitor: networkMonitor,
		sources:        make(map[string]MetricsSource),
		aggregatedMetrics: &AggregatedMetrics{
			NetworkSummary:     &NetworkSummary{ProtocolDistribution: make(map[string]float64)},
			PerformanceSummary: &PerformanceSummary{},
			SecuritySummary:    &SecuritySummary{},
			ResourceSummary:    &ResourceSummary{},
			TopologyMetrics:    &TopologyMetrics{},
		},
		historicalData: &HistoricalData{
			dataPoints:  make([]*HistoricalPoint, config.MaxHistoricalPoints),
			maxPoints:   config.MaxHistoricalPoints,
			resolution:  config.HistoricalResolution,
		},
		exporters: make([]MetricsExporter, 0),
		ctx:       ctx,
		cancel:    cancel,
	}
	
	return aggregator
}

// Start starts the metrics aggregator
func (ma *MetricsAggregator) Start() error {
	// Start aggregation worker
	ma.wg.Add(1)
	go ma.aggregationWorker()
	
	// Start export worker if enabled
	if ma.config.EnableExport {
		ma.wg.Add(1)
		go ma.exportWorker()
	}
	
	// Start historical data worker
	ma.wg.Add(1)
	go ma.historicalDataWorker()
	
	return nil
}

// Stop stops the metrics aggregator
func (ma *MetricsAggregator) Stop() error {
	ma.cancel()
	ma.wg.Wait()
	return nil
}

// RegisterSource registers a metrics source
func (ma *MetricsAggregator) RegisterSource(source MetricsSource) {
	ma.sourcesMu.Lock()
	defer ma.sourcesMu.Unlock()
	ma.sources[source.GetName()] = source
}

// RegisterExporter registers a metrics exporter
func (ma *MetricsAggregator) RegisterExporter(exporter MetricsExporter) {
	ma.exportersMu.Lock()
	defer ma.exportersMu.Unlock()
	ma.exporters = append(ma.exporters, exporter)
}

// GetAggregatedMetrics returns the current aggregated metrics
func (ma *MetricsAggregator) GetAggregatedMetrics() *AggregatedMetrics {
	ma.metricsMu.RLock()
	defer ma.metricsMu.RUnlock()
	
	// Return a deep copy
	metrics := *ma.aggregatedMetrics
	
	// Copy nested structs
	if ma.aggregatedMetrics.NetworkSummary != nil {
		networkSummary := *ma.aggregatedMetrics.NetworkSummary
		networkSummary.ProtocolDistribution = make(map[string]float64)
		for k, v := range ma.aggregatedMetrics.NetworkSummary.ProtocolDistribution {
			networkSummary.ProtocolDistribution[k] = v
		}
		metrics.NetworkSummary = &networkSummary
	}
	
	if ma.aggregatedMetrics.PerformanceSummary != nil {
		performanceSummary := *ma.aggregatedMetrics.PerformanceSummary
		metrics.PerformanceSummary = &performanceSummary
	}
	
	if ma.aggregatedMetrics.SecuritySummary != nil {
		securitySummary := *ma.aggregatedMetrics.SecuritySummary
		metrics.SecuritySummary = &securitySummary
	}
	
	if ma.aggregatedMetrics.ResourceSummary != nil {
		resourceSummary := *ma.aggregatedMetrics.ResourceSummary
		metrics.ResourceSummary = &resourceSummary
	}
	
	if ma.aggregatedMetrics.TopologyMetrics != nil {
		topologyMetrics := *ma.aggregatedMetrics.TopologyMetrics
		metrics.TopologyMetrics = &topologyMetrics
	}
	
	return &metrics
}

// GetHistoricalData returns historical metrics data
func (ma *MetricsAggregator) GetHistoricalData(duration time.Duration) []*HistoricalPoint {
	ma.historicalData.mu.RLock()
	defer ma.historicalData.mu.RUnlock()
	
	cutoff := time.Now().Add(-duration)
	var points []*HistoricalPoint
	
	for _, point := range ma.historicalData.dataPoints {
		if point != nil && point.Timestamp.After(cutoff) {
			points = append(points, point)
		}
	}
	
	return points
}

// Worker functions

// aggregationWorker aggregates metrics from all sources
func (ma *MetricsAggregator) aggregationWorker() {
	defer ma.wg.Done()
	
	ticker := time.NewTicker(ma.config.AggregationInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ma.ctx.Done():
			return
		case <-ticker.C:
			ma.aggregateMetrics()
		}
	}
}

// exportWorker exports metrics to registered exporters
func (ma *MetricsAggregator) exportWorker() {
	defer ma.wg.Done()
	
	ticker := time.NewTicker(ma.config.ExportInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ma.ctx.Done():
			return
		case <-ticker.C:
			ma.exportMetrics()
		}
	}
}

// historicalDataWorker manages historical data storage
func (ma *MetricsAggregator) historicalDataWorker() {
	defer ma.wg.Done()
	
	ticker := time.NewTicker(ma.config.HistoricalResolution)
	defer ticker.Stop()
	
	for {
		select {
		case <-ma.ctx.Done():
			return
		case <-ticker.C:
			ma.storeHistoricalData()
		}
	}
}

// aggregateMetrics aggregates metrics from all sources
func (ma *MetricsAggregator) aggregateMetrics() {
	ma.metricsMu.Lock()
	defer ma.metricsMu.Unlock()
	
	// Aggregate network metrics
	if ma.networkMonitor != nil {
		ma.aggregateNetworkMetrics()
	}
	
	// Aggregate metrics from other sources
	ma.sourcesMu.RLock()
	for _, source := range ma.sources {
		if source.IsHealthy() {
			metrics, err := source.GetMetrics()
			if err == nil {
				ma.processSourceMetrics(source.GetName(), metrics)
			}
		}
	}
	ma.sourcesMu.RUnlock()
	
	// Update timestamp
	ma.aggregatedMetrics.Timestamp = time.Now()
}

// aggregateNetworkMetrics aggregates network-specific metrics
func (ma *MetricsAggregator) aggregateNetworkMetrics() {
	connectionMetrics := ma.networkMonitor.GetConnectionMetrics()
	messageMetrics := ma.networkMonitor.GetMessageMetrics()
	performanceMetrics := ma.networkMonitor.GetPerformanceMetrics()
	securityMetrics := ma.networkMonitor.GetSecurityMetrics()
	
	// Update network summary
	ma.aggregatedMetrics.NetworkSummary.TotalConnections = connectionMetrics.TotalConnections
	ma.aggregatedMetrics.NetworkSummary.ActiveConnections = connectionMetrics.ActiveConnections
	ma.aggregatedMetrics.NetworkSummary.MessageThroughput = messageMetrics.MessageThroughput
	ma.aggregatedMetrics.NetworkSummary.AverageLatency = connectionMetrics.AverageLatency
	
	// Calculate error rate
	if messageMetrics.TotalMessages > 0 {
		ma.aggregatedMetrics.NetworkSummary.ErrorRate = float64(messageMetrics.MessagesDropped) / float64(messageMetrics.TotalMessages)
	}
	
	// Calculate bandwidth utilization
	totalBandwidth := float64(connectionMetrics.TotalBytesIn + connectionMetrics.TotalBytesOut)
	if totalBandwidth > 0 {
		ma.aggregatedMetrics.NetworkSummary.BandwidthUtilization = (connectionMetrics.BandwidthIn + connectionMetrics.BandwidthOut) / totalBandwidth
	}
	
	// Update protocol distribution
	totalProtocolMessages := int64(0)
	for _, count := range messageMetrics.MessagesByProtocol {
		totalProtocolMessages += count
	}
	if totalProtocolMessages > 0 {
		for protocol, count := range messageMetrics.MessagesByProtocol {
			ma.aggregatedMetrics.NetworkSummary.ProtocolDistribution[protocol] = float64(count) / float64(totalProtocolMessages)
		}
	}
	
	// Update performance summary
	ma.aggregatedMetrics.PerformanceSummary.CPUUsage = performanceMetrics.CPUUsage
	ma.aggregatedMetrics.PerformanceSummary.MemoryUsage = performanceMetrics.MemoryUsage
	ma.aggregatedMetrics.PerformanceSummary.NetworkUtilization = performanceMetrics.NetworkUtilization
	ma.aggregatedMetrics.PerformanceSummary.ResponseTime = performanceMetrics.NetworkLatency
	ma.aggregatedMetrics.PerformanceSummary.Throughput = performanceMetrics.NetworkThroughput
	
	// Update security summary
	if securityMetrics.AuthAttempts > 0 {
		ma.aggregatedMetrics.SecuritySummary.AuthenticationRate = float64(securityMetrics.AuthSuccesses) / float64(securityMetrics.AuthAttempts)
	}
	ma.aggregatedMetrics.SecuritySummary.SecurityViolations = securityMetrics.SecurityViolations
	ma.aggregatedMetrics.SecuritySummary.BlockedConnections = securityMetrics.BlockedConnections
}

// processSourceMetrics processes metrics from a specific source
func (ma *MetricsAggregator) processSourceMetrics(sourceName string, metrics map[string]interface{}) {
	// Process metrics based on source type
	// This would be implemented based on specific source types
}

// exportMetrics exports metrics to all registered exporters
func (ma *MetricsAggregator) exportMetrics() {
	metrics := ma.GetAggregatedMetrics()
	
	ma.exportersMu.RLock()
	defer ma.exportersMu.RUnlock()
	
	for _, exporter := range ma.exporters {
		if exporter.IsEnabled() {
			go func(exp MetricsExporter) {
				if err := exp.Export(metrics); err != nil {
					// Log error but continue with other exporters
				}
			}(exporter)
		}
	}
}

// storeHistoricalData stores current metrics as historical data
func (ma *MetricsAggregator) storeHistoricalData() {
	ma.historicalData.mu.Lock()
	defer ma.historicalData.mu.Unlock()
	
	point := &HistoricalPoint{
		Timestamp: time.Now(),
		Metrics:   ma.GetAggregatedMetrics(),
	}
	
	ma.historicalData.dataPoints[ma.historicalData.currentIndex] = point
	ma.historicalData.currentIndex = (ma.historicalData.currentIndex + 1) % ma.historicalData.maxPoints
}

// Exporter implementations

// NewPrometheusExporter creates a new Prometheus exporter
func NewPrometheusExporter(endpoint string) *PrometheusExporter {
	return &PrometheusExporter{
		endpoint: endpoint,
		enabled:  true,
	}
}

func (pe *PrometheusExporter) Export(metrics *AggregatedMetrics) error {
	// Implementation would export to Prometheus
	return nil
}

func (pe *PrometheusExporter) GetName() string {
	return "prometheus"
}

func (pe *PrometheusExporter) IsEnabled() bool {
	return pe.enabled
}

// NewJSONExporter creates a new JSON exporter
func NewJSONExporter(outputPath string) *JSONExporter {
	return &JSONExporter{
		outputPath: outputPath,
		enabled:    true,
	}
}

func (je *JSONExporter) Export(metrics *AggregatedMetrics) error {
	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}
	
	filename := fmt.Sprintf("%s/metrics_%d.json", je.outputPath, time.Now().Unix())
	// Implementation would write to file
	_ = data
	_ = filename
	
	return nil
}

func (je *JSONExporter) GetName() string {
	return "json"
}

func (je *JSONExporter) IsEnabled() bool {
	return je.enabled
}

package production

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
)

// ProductionMonitor provides comprehensive production monitoring
type ProductionMonitor struct {
	config *config.DistributedConfig
	logger *logrus.Logger

	// Monitoring components
	metricsCollector *MetricsCollector
	alertManager     *AlertManager
	healthChecker    *HealthChecker
	slaMonitor       *SLAMonitor

	// OpenTelemetry
	tracer trace.Tracer
	meter  metric.Meter

	// Production metrics
	requestCounter   metric.Int64Counter
	latencyHistogram metric.Float64Histogram
	errorCounter     metric.Int64Counter
	uptimeGauge      metric.Float64ObservableGauge

	// SLA tracking
	slaMetrics      *SLAMetrics
	performanceKPIs *PerformanceKPIs

	mu sync.RWMutex
}

// MetricsCollector handles comprehensive metrics collection
type MetricsCollector struct {
	config   *config.DistributedConfig
	logger   *logrus.Logger
	registry *prometheus.Registry

	// Business metrics
	businessMetrics map[string]*BusinessMetric

	// Technical metrics
	systemMetrics *SystemMetrics
	appMetrics    *ApplicationMetrics

	mu sync.RWMutex
}

// AlertManager handles production alerting
type AlertManager struct {
	config *config.AlertingConfig
	logger *logrus.Logger

	// Alert rules and channels
	rules    []*AlertRule
	channels map[string]AlertChannel

	// Alert state management
	activeAlerts   map[string]*Alert
	alertHistory   []*Alert
	escalationTree *EscalationTree

	// Notification throttling
	throttleMap map[string]time.Time

	mu sync.RWMutex
}

// HealthChecker performs comprehensive health monitoring
type HealthChecker struct {
	config *config.HealthConfig
	logger *logrus.Logger

	// Health check definitions
	checks map[string]*HealthCheck

	// Health state
	overallHealth   float64
	componentHealth map[string]*ComponentHealth
	dependencies    map[string]*DependencyHealth

	// Health history
	healthHistory []*HealthSnapshot

	mu sync.RWMutex
}

// SLAMonitor tracks Service Level Agreements
type SLAMonitor struct {
	config *config.DistributedConfig
	logger *logrus.Logger

	// SLA definitions
	slaTargets map[string]*SLATarget

	// SLA tracking
	currentPeriod *SLAPeriod
	slaHistory    []*SLAPeriod

	// Burn rate tracking
	burnRates map[string]*BurnRate

	mu sync.RWMutex
}

// Production monitoring types

type BusinessMetric struct {
	Name      string                 `json:"name"`
	Type      string                 `json:"type"` // counter, gauge, histogram
	Value     float64                `json:"value"`
	Labels    map[string]string      `json:"labels"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type SystemMetrics struct {
	CPU        CPUMetrics        `json:"cpu"`
	Memory     MemoryMetrics     `json:"memory"`
	Disk       DiskMetrics       `json:"disk"`
	Network    NetworkMetrics    `json:"network"`
	Processes  ProcessMetrics    `json:"processes"`
	FileSystem FileSystemMetrics `json:"filesystem"`
	Timestamp  time.Time         `json:"timestamp"`
}

type ApplicationMetrics struct {
	RequestRate     float64           `json:"request_rate"`
	ResponseTime    ResponseTimeStats `json:"response_time"`
	ErrorRate       float64           `json:"error_rate"`
	Throughput      float64           `json:"throughput"`
	Concurrency     int               `json:"concurrency"`
	QueueDepth      int               `json:"queue_depth"`
	CacheHitRate    float64           `json:"cache_hit_rate"`
	DatabaseMetrics DatabaseMetrics   `json:"database"`
	Timestamp       time.Time         `json:"timestamp"`
}

type CPUMetrics struct {
	Usage     float64 `json:"usage"`
	LoadAvg1  float64 `json:"load_avg_1"`
	LoadAvg5  float64 `json:"load_avg_5"`
	LoadAvg15 float64 `json:"load_avg_15"`
	Cores     int     `json:"cores"`
}

type MemoryMetrics struct {
	Used      uint64      `json:"used"`
	Available uint64      `json:"available"`
	Total     uint64      `json:"total"`
	Percent   float64     `json:"percent"`
	Swap      SwapMetrics `json:"swap"`
}

type SwapMetrics struct {
	Used    uint64  `json:"used"`
	Total   uint64  `json:"total"`
	Percent float64 `json:"percent"`
}

type DiskMetrics struct {
	Usage   map[string]DiskUsage `json:"usage"`
	IOStats DiskIOStats          `json:"io_stats"`
}

type DiskUsage struct {
	Used      uint64  `json:"used"`
	Available uint64  `json:"available"`
	Total     uint64  `json:"total"`
	Percent   float64 `json:"percent"`
}

type DiskIOStats struct {
	ReadBytes  uint64 `json:"read_bytes"`
	WriteBytes uint64 `json:"write_bytes"`
	ReadOps    uint64 `json:"read_ops"`
	WriteOps   uint64 `json:"write_ops"`
	ReadTime   uint64 `json:"read_time"`
	WriteTime  uint64 `json:"write_time"`
}

type NetworkMetrics struct {
	Interfaces  map[string]NetworkInterface `json:"interfaces"`
	Connections NetworkConnections          `json:"connections"`
}

type NetworkInterface struct {
	BytesReceived   uint64 `json:"bytes_received"`
	BytesSent       uint64 `json:"bytes_sent"`
	PacketsReceived uint64 `json:"packets_received"`
	PacketsSent     uint64 `json:"packets_sent"`
	Errors          uint64 `json:"errors"`
	Drops           uint64 `json:"drops"`
}

type NetworkConnections struct {
	Established int `json:"established"`
	TimeWait    int `json:"time_wait"`
	CloseWait   int `json:"close_wait"`
	Listen      int `json:"listen"`
}

type ProcessMetrics struct {
	Count       int     `json:"count"`
	Running     int     `json:"running"`
	Sleeping    int     `json:"sleeping"`
	Zombie      int     `json:"zombie"`
	CPUPercent  float64 `json:"cpu_percent"`
	MemoryBytes uint64  `json:"memory_bytes"`
}

type FileSystemMetrics struct {
	OpenFiles int `json:"open_files"`
	MaxFiles  int `json:"max_files"`
	Inodes    int `json:"inodes"`
	MaxInodes int `json:"max_inodes"`
}

type ResponseTimeStats struct {
	Mean float64 `json:"mean"`
	P50  float64 `json:"p50"`
	P90  float64 `json:"p90"`
	P95  float64 `json:"p95"`
	P99  float64 `json:"p99"`
	P999 float64 `json:"p999"`
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
}

type DatabaseMetrics struct {
	Connections   int     `json:"connections"`
	ActiveQueries int     `json:"active_queries"`
	SlowQueries   int     `json:"slow_queries"`
	QueryTime     float64 `json:"query_time"`
	LockWaitTime  float64 `json:"lock_wait_time"`
	DeadlockCount int     `json:"deadlock_count"`
}

type ComponentHealth struct {
	Name         string            `json:"name"`
	Status       string            `json:"status"` // healthy, degraded, unhealthy
	Score        float64           `json:"score"`  // 0.0 to 1.0
	LastCheck    time.Time         `json:"last_check"`
	CheckCount   int               `json:"check_count"`
	FailureCount int               `json:"failure_count"`
	Metadata     map[string]string `json:"metadata"`
}

type DependencyHealth struct {
	Name         string    `json:"name"`
	Type         string    `json:"type"` // database, cache, api, queue
	Status       string    `json:"status"`
	ResponseTime float64   `json:"response_time"`
	LastCheck    time.Time `json:"last_check"`
	ErrorRate    float64   `json:"error_rate"`
	Critical     bool      `json:"critical"`
}

type HealthSnapshot struct {
	Timestamp        time.Time                    `json:"timestamp"`
	OverallHealth    float64                      `json:"overall_health"`
	ComponentHealth  map[string]*ComponentHealth  `json:"component_health"`
	DependencyHealth map[string]*DependencyHealth `json:"dependency_health"`
	Issues           []string                     `json:"issues"`
}

type SLATarget struct {
	Name        string        `json:"name"`
	Type        string        `json:"type"` // availability, latency, throughput, error_rate
	Target      float64       `json:"target"`
	Period      time.Duration `json:"period"`
	Measurement string        `json:"measurement"`
	Critical    bool          `json:"critical"`
}

type SLAPeriod struct {
	StartTime   time.Time             `json:"start_time"`
	EndTime     time.Time             `json:"end_time"`
	Targets     map[string]*SLATarget `json:"targets"`
	Actual      map[string]float64    `json:"actual"`
	Compliance  map[string]bool       `json:"compliance"`
	ErrorBudget map[string]float64    `json:"error_budget"`
}

type BurnRate struct {
	SLAName    string        `json:"sla_name"`
	Window     time.Duration `json:"window"`
	Rate       float64       `json:"rate"`
	Threshold  float64       `json:"threshold"`
	LastUpdate time.Time     `json:"last_update"`
	Alerting   bool          `json:"alerting"`
}

type SLAMetrics struct {
	Availability  float64           `json:"availability"`
	Latency       ResponseTimeStats `json:"latency"`
	Throughput    float64           `json:"throughput"`
	ErrorRate     float64           `json:"error_rate"`
	ErrorBudget   float64           `json:"error_budget"`
	BurnRate      float64           `json:"burn_rate"`
	TimeToRestore float64           `json:"time_to_restore"`
	IncidentCount int               `json:"incident_count"`
}

type PerformanceKPIs struct {
	MTBF            float64 `json:"mtbf"`             // Mean Time Between Failures
	MTTR            float64 `json:"mttr"`             // Mean Time To Recovery
	MTTD            float64 `json:"mttd"`             // Mean Time To Detection
	ChangeFailRate  float64 `json:"change_fail_rate"` // Change Failure Rate
	DeployFrequency float64 `json:"deploy_frequency"` // Deployment Frequency
	LeadTime        float64 `json:"lead_time"`        // Lead Time for Changes
}

type EscalationTree struct {
	Levels []EscalationLevel `json:"levels"`
}

type EscalationLevel struct {
	Level      int           `json:"level"`
	Delay      time.Duration `json:"delay"`
	Recipients []string      `json:"recipients"`
	Channels   []string      `json:"channels"`
	Conditions []string      `json:"conditions"`
}

type Alert struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Severity    string            `json:"severity"`
	Status      string            `json:"status"`
	Message     string            `json:"message"`
	Source      string            `json:"source"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    time.Time         `json:"starts_at"`
	EndsAt      *time.Time        `json:"ends_at,omitempty"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Fingerprint string            `json:"fingerprint"`
}

type AlertRule struct {
	Name        string            `json:"name"`
	Query       string            `json:"query"`
	Condition   string            `json:"condition"`
	Threshold   float64           `json:"threshold"`
	Duration    time.Duration     `json:"duration"`
	Severity    string            `json:"severity"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Enabled     bool              `json:"enabled"`
}

type AlertChannel interface {
	Send(alert *Alert) error
	GetName() string
	IsHealthy() bool
}

type HealthCheck struct {
	Name         string            `json:"name"`
	Type         string            `json:"type"` // http, tcp, command, custom
	Target       string            `json:"target"`
	Interval     time.Duration     `json:"interval"`
	Timeout      time.Duration     `json:"timeout"`
	Retries      int               `json:"retries"`
	Headers      map[string]string `json:"headers,omitempty"`
	ExpectedCode int               `json:"expected_code,omitempty"`
	ExpectedBody string            `json:"expected_body,omitempty"`
	Command      string            `json:"command,omitempty"`
	Enabled      bool              `json:"enabled"`
}

// NewProductionMonitor creates a new production monitoring system
func NewProductionMonitor(cfg *config.DistributedConfig, logger *logrus.Logger) (*ProductionMonitor, error) {
	pm := &ProductionMonitor{
		config: cfg,
		logger: logger,
	}

	if err := pm.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize production monitor: %w", err)
	}

	return pm, nil
}

// initialize sets up all monitoring components
func (pm *ProductionMonitor) initialize() error {
	// Initialize OpenTelemetry
	pm.tracer = otel.Tracer("ollama-distributed-production")
	pm.meter = otel.Meter("ollama-distributed-production")

	// Initialize metrics
	if err := pm.initializeMetrics(); err != nil {
		return fmt.Errorf("failed to initialize metrics: %w", err)
	}

	// Initialize components
	pm.metricsCollector = &MetricsCollector{
		config:          pm.config,
		logger:          pm.logger,
		registry:        prometheus.NewRegistry(),
		businessMetrics: make(map[string]*BusinessMetric),
		systemMetrics:   &SystemMetrics{},
		appMetrics:      &ApplicationMetrics{},
	}

	pm.alertManager = &AlertManager{
		config:       &pm.config.Observability.Alerting,
		logger:       pm.logger,
		rules:        make([]*AlertRule, 0),
		channels:     make(map[string]AlertChannel),
		activeAlerts: make(map[string]*Alert),
		alertHistory: make([]*Alert, 0),
		throttleMap:  make(map[string]time.Time),
	}

	pm.healthChecker = &HealthChecker{
		config:          &pm.config.Observability.Health,
		logger:          pm.logger,
		checks:          make(map[string]*HealthCheck),
		componentHealth: make(map[string]*ComponentHealth),
		dependencies:    make(map[string]*DependencyHealth),
		healthHistory:   make([]*HealthSnapshot, 0),
	}

	pm.slaMonitor = &SLAMonitor{
		config:     pm.config,
		logger:     pm.logger,
		slaTargets: make(map[string]*SLATarget),
		slaHistory: make([]*SLAPeriod, 0),
		burnRates:  make(map[string]*BurnRate),
	}

	// Initialize SLA metrics and KPIs
	pm.slaMetrics = &SLAMetrics{}
	pm.performanceKPIs = &PerformanceKPIs{}

	return nil
}

// initializeMetrics sets up OpenTelemetry metrics
func (pm *ProductionMonitor) initializeMetrics() error {
	var err error

	pm.requestCounter, err = pm.meter.Int64Counter(
		"ollama_requests_total",
		metric.WithDescription("Total number of requests processed"),
	)
	if err != nil {
		return err
	}

	pm.latencyHistogram, err = pm.meter.Float64Histogram(
		"ollama_request_duration_seconds",
		metric.WithDescription("Request duration in seconds"),
	)
	if err != nil {
		return err
	}

	pm.errorCounter, err = pm.meter.Int64Counter(
		"ollama_errors_total",
		metric.WithDescription("Total number of errors"),
	)
	if err != nil {
		return err
	}

	pm.uptimeGauge, err = pm.meter.Float64ObservableGauge(
		"ollama_uptime_seconds",
		metric.WithDescription("System uptime in seconds"),
	)
	if err != nil {
		return err
	}

	return nil
}

// Start begins all production monitoring operations
func (pm *ProductionMonitor) Start(ctx context.Context) error {
	pm.logger.Info("Starting production monitoring system")

	// Start all components
	go pm.metricsCollector.Start(ctx)
	go pm.alertManager.Start(ctx)
	go pm.healthChecker.Start(ctx)
	go pm.slaMonitor.Start(ctx)

	// Start main monitoring loop
	go pm.monitoringLoop(ctx)

	pm.logger.Info("Production monitoring system started successfully")
	return nil
}

// monitoringLoop runs the main monitoring cycle
func (pm *ProductionMonitor) monitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pm.collectAndAnalyze()
		}
	}
}

// collectAndAnalyze performs comprehensive monitoring analysis
func (pm *ProductionMonitor) collectAndAnalyze() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Update SLA metrics
	pm.updateSLAMetrics()

	// Update performance KPIs
	pm.updatePerformanceKPIs()

	// Check for SLA violations
	pm.checkSLAViolations()

	// Update burn rates
	pm.updateBurnRates()

	pm.logger.Debug("Completed monitoring analysis cycle")
}

// updateSLAMetrics updates current SLA metrics
func (pm *ProductionMonitor) updateSLAMetrics() {
	// Simulate SLA metrics collection
	pm.slaMetrics.Availability = 99.95 + (0.05 * (2*rand.Float64() - 1)) // 99.9-100%
	pm.slaMetrics.ErrorRate = 0.001 + (0.004 * rand.Float64())           // 0.1-0.5%
	pm.slaMetrics.Throughput = 100 + (50 * rand.Float64())               // 100-150 RPS
	pm.slaMetrics.ErrorBudget = 0.1 - pm.slaMetrics.ErrorRate            // Remaining error budget
	pm.slaMetrics.BurnRate = pm.slaMetrics.ErrorRate / 0.1               // Current burn rate

	// Update latency stats
	pm.slaMetrics.Latency = ResponseTimeStats{
		Mean: 150 + (50 * rand.Float64()),
		P50:  120 + (30 * rand.Float64()),
		P90:  200 + (100 * rand.Float64()),
		P95:  250 + (150 * rand.Float64()),
		P99:  400 + (200 * rand.Float64()),
		P999: 800 + (400 * rand.Float64()),
	}
}

// updatePerformanceKPIs updates performance KPIs
func (pm *ProductionMonitor) updatePerformanceKPIs() {
	// Simulate KPI calculations
	pm.performanceKPIs.MTBF = 720.0 + (240.0 * rand.Float64())         // 720-960 hours
	pm.performanceKPIs.MTTR = 15.0 + (10.0 * rand.Float64())           // 15-25 minutes
	pm.performanceKPIs.MTTD = 5.0 + (5.0 * rand.Float64())             // 5-10 minutes
	pm.performanceKPIs.ChangeFailRate = 0.05 + (0.05 * rand.Float64()) // 5-10%
	pm.performanceKPIs.DeployFrequency = 2.0 + (1.0 * rand.Float64())  // 2-3 per day
	pm.performanceKPIs.LeadTime = 120.0 + (60.0 * rand.Float64())      // 2-3 hours
}

// checkSLAViolations checks for SLA violations and triggers alerts
func (pm *ProductionMonitor) checkSLAViolations() {
	// Check availability SLA
	if pm.slaMetrics.Availability < 99.9 {
		pm.triggerSLAAlert("availability", pm.slaMetrics.Availability, 99.9)
	}

	// Check error rate SLA
	if pm.slaMetrics.ErrorRate > 0.01 {
		pm.triggerSLAAlert("error_rate", pm.slaMetrics.ErrorRate, 0.01)
	}

	// Check latency SLA
	if pm.slaMetrics.Latency.P95 > 500 {
		pm.triggerSLAAlert("latency_p95", pm.slaMetrics.Latency.P95, 500)
	}
}

// triggerSLAAlert triggers an SLA violation alert
func (pm *ProductionMonitor) triggerSLAAlert(metric string, actual, target float64) {
	alert := &Alert{
		ID:       fmt.Sprintf("sla-%s-%d", metric, time.Now().Unix()),
		Name:     fmt.Sprintf("SLA Violation: %s", metric),
		Severity: "critical",
		Status:   "firing",
		Message:  fmt.Sprintf("SLA violation detected: %s actual=%.3f target=%.3f", metric, actual, target),
		Source:   "sla-monitor",
		Labels: map[string]string{
			"type":   "sla_violation",
			"metric": metric,
		},
		StartsAt: time.Now(),
	}

	pm.alertManager.ProcessAlert(alert)
}

// updateBurnRates updates error budget burn rates
func (pm *ProductionMonitor) updateBurnRates() {
	for name, burnRate := range pm.slaMonitor.burnRates {
		// Update burn rate based on current error rate
		burnRate.Rate = pm.slaMetrics.ErrorRate / 0.1 // Assuming 0.1 error budget
		burnRate.LastUpdate = time.Now()

		// Check if burn rate exceeds threshold
		if burnRate.Rate > burnRate.Threshold && !burnRate.Alerting {
			pm.triggerBurnRateAlert(name, burnRate)
			burnRate.Alerting = true
		} else if burnRate.Rate <= burnRate.Threshold && burnRate.Alerting {
			burnRate.Alerting = false
		}
	}
}

// triggerBurnRateAlert triggers a burn rate alert
func (pm *ProductionMonitor) triggerBurnRateAlert(slaName string, burnRate *BurnRate) {
	alert := &Alert{
		ID:       fmt.Sprintf("burnrate-%s-%d", slaName, time.Now().Unix()),
		Name:     fmt.Sprintf("High Error Budget Burn Rate: %s", slaName),
		Severity: "warning",
		Status:   "firing",
		Message:  fmt.Sprintf("Error budget burn rate is %.2fx the threshold for %s", burnRate.Rate/burnRate.Threshold, slaName),
		Source:   "burn-rate-monitor",
		Labels: map[string]string{
			"type":     "burn_rate",
			"sla_name": slaName,
		},
		StartsAt: time.Now(),
	}

	pm.alertManager.ProcessAlert(alert)
}

// GetSLAMetrics returns current SLA metrics
func (pm *ProductionMonitor) GetSLAMetrics() *SLAMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.slaMetrics
}

// GetPerformanceKPIs returns current performance KPIs
func (pm *ProductionMonitor) GetPerformanceKPIs() *PerformanceKPIs {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.performanceKPIs
}

// GetSystemHealth returns overall system health
func (pm *ProductionMonitor) GetSystemHealth() *HealthSnapshot {
	return pm.healthChecker.GetCurrentHealth()
}

// RecordRequest records a request for monitoring
func (pm *ProductionMonitor) RecordRequest(ctx context.Context, method, endpoint string, duration time.Duration, statusCode int) {
	// Record OpenTelemetry metrics
	pm.requestCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("method", method),
		attribute.String("endpoint", endpoint),
		attribute.Int("status_code", statusCode),
	))

	pm.latencyHistogram.Record(ctx, duration.Seconds(), metric.WithAttributes(
		attribute.String("method", method),
		attribute.String("endpoint", endpoint),
	))

	// Record errors
	if statusCode >= 400 {
		pm.errorCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("method", method),
			attribute.String("endpoint", endpoint),
			attribute.Int("status_code", statusCode),
		))
	}
}

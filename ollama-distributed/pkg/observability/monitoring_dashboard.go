package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// MonitoringDashboard provides real-time monitoring visualization
type MonitoringDashboard struct {
	config          *DashboardConfig
	metricsRegistry *MetricsRegistry
	healthManager   *HealthCheckManager
	notificationSys *NotificationSystem

	// WebSocket connections for real-time updates
	connections map[string]*websocket.Conn
	connMutex   sync.RWMutex

	// Data aggregation
	dataAggregator *DataAggregator

	ctx    context.Context
	cancel context.CancelFunc
}

// DashboardConfig configures the monitoring dashboard
type DashboardConfig struct {
	Enabled         bool          `json:"enabled"`
	Port            int           `json:"port"`
	UpdateInterval  time.Duration `json:"update_interval"`
	RetentionPeriod time.Duration `json:"retention_period"`

	// Real-time features
	EnableWebSocket     bool `json:"enable_websocket"`
	EnableAlerts        bool `json:"enable_alerts"`
	EnableNotifications bool `json:"enable_notifications"`

	// Dashboard customization
	Theme            string   `json:"theme"`
	DefaultPanels    []string `json:"default_panels"`
	CustomDashboards []string `json:"custom_dashboards"`
}

// DataAggregator aggregates metrics and health data for visualization
type DataAggregator struct {
	metricsData map[string]*MetricSeries
	healthData  map[string]*HealthSeries
	alertData   map[string]*AlertSeries
	mutex       sync.RWMutex
}

// MetricSeries represents time-series metric data
type MetricSeries struct {
	Name       string                 `json:"name"`
	Labels     map[string]string      `json:"labels"`
	Values     []MetricValue          `json:"values"`
	LastUpdate time.Time              `json:"last_update"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// MetricValue represents a single metric value with timestamp
type MetricValue struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// HealthSeries represents time-series health data
type HealthSeries struct {
	Component  string        `json:"component"`
	Status     string        `json:"status"`
	Values     []HealthValue `json:"values"`
	LastUpdate time.Time     `json:"last_update"`
}

// HealthValue represents a single health check result
type HealthValue struct {
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
}

// AlertSeries represents time-series alert data
type AlertSeries struct {
	AlertName  string       `json:"alert_name"`
	Severity   string       `json:"severity"`
	Values     []AlertValue `json:"values"`
	LastUpdate time.Time    `json:"last_update"`
}

// AlertValue represents a single alert event
type AlertValue struct {
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"` // firing, resolved
	Message   string    `json:"message"`
}

// DashboardData represents the complete dashboard data
type DashboardData struct {
	Timestamp time.Time                `json:"timestamp"`
	Metrics   map[string]*MetricSeries `json:"metrics"`
	Health    map[string]*HealthSeries `json:"health"`
	Alerts    map[string]*AlertSeries  `json:"alerts"`
	Summary   *SystemSummary           `json:"summary"`
}

// SystemSummary provides high-level system status
type SystemSummary struct {
	OverallHealth string    `json:"overall_health"`
	ActiveAlerts  int       `json:"active_alerts"`
	TotalNodes    int       `json:"total_nodes"`
	HealthyNodes  int       `json:"healthy_nodes"`
	RequestRate   float64   `json:"request_rate"`
	ErrorRate     float64   `json:"error_rate"`
	LastUpdate    time.Time `json:"last_update"`
}

// NewMonitoringDashboard creates a new monitoring dashboard
func NewMonitoringDashboard(config *DashboardConfig, metricsRegistry *MetricsRegistry, healthManager *HealthCheckManager, notificationSys *NotificationSystem) *MonitoringDashboard {
	if config == nil {
		config = DefaultDashboardConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	dashboard := &MonitoringDashboard{
		config:          config,
		metricsRegistry: metricsRegistry,
		healthManager:   healthManager,
		notificationSys: notificationSys,
		connections:     make(map[string]*websocket.Conn),
		dataAggregator:  NewDataAggregator(),
		ctx:             ctx,
		cancel:          cancel,
	}

	return dashboard
}

// NewDataAggregator creates a new data aggregator
func NewDataAggregator() *DataAggregator {
	return &DataAggregator{
		metricsData: make(map[string]*MetricSeries),
		healthData:  make(map[string]*HealthSeries),
		alertData:   make(map[string]*AlertSeries),
	}
}

// Start starts the monitoring dashboard
func (md *MonitoringDashboard) Start() error {
	if !md.config.Enabled {
		log.Info().Msg("Monitoring dashboard disabled")
		return nil
	}

	// Start data collection
	go md.startDataCollection()

	// Start HTTP server for dashboard
	go md.startHTTPServer()

	// Start WebSocket server for real-time updates
	if md.config.EnableWebSocket {
		go md.startWebSocketUpdates()
	}

	log.Info().
		Int("port", md.config.Port).
		Msg("Monitoring dashboard started")

	return nil
}

// startDataCollection starts collecting metrics and health data
func (md *MonitoringDashboard) startDataCollection() {
	ticker := time.NewTicker(md.config.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-md.ctx.Done():
			return
		case <-ticker.C:
			md.collectData()
		}
	}
}

// collectData collects current metrics and health data
func (md *MonitoringDashboard) collectData() {
	md.dataAggregator.mutex.Lock()
	defer md.dataAggregator.mutex.Unlock()

	now := time.Now()

	// Collect metrics data
	if md.metricsRegistry != nil {
		metrics := md.metricsRegistry.GetAllMetrics()
		for name, metric := range metrics {
			series, exists := md.dataAggregator.metricsData[name]
			if !exists {
				series = &MetricSeries{
					Name:   name,
					Labels: make(map[string]string),
					Values: make([]MetricValue, 0),
				}
				md.dataAggregator.metricsData[name] = series
			}

			// Add new value
			value := MetricValue{
				Timestamp: now,
				Value:     metric.Value,
			}
			series.Values = append(series.Values, value)
			series.LastUpdate = now

			// Trim old values
			md.trimOldValues(series)
		}
	}

	// Collect health data
	if md.healthManager != nil {
		healthStatus := md.healthManager.GetOverallHealth()
		for component, status := range healthStatus.Components {
			series, exists := md.dataAggregator.healthData[component]
			if !exists {
				series = &HealthSeries{
					Component: component,
					Values:    make([]HealthValue, 0),
				}
				md.dataAggregator.healthData[component] = series
			}

			// Add new value
			value := HealthValue{
				Timestamp: now,
				Status:    string(status.Status),
				Message:   status.Message,
			}
			series.Values = append(series.Values, value)
			series.Status = string(status.Status)
			series.LastUpdate = now

			// Trim old values
			md.trimOldHealthValues(series)
		}
	}
}

// trimOldValues removes values older than retention period
func (md *MonitoringDashboard) trimOldValues(series *MetricSeries) {
	cutoff := time.Now().Add(-md.config.RetentionPeriod)
	var newValues []MetricValue

	for _, value := range series.Values {
		if value.Timestamp.After(cutoff) {
			newValues = append(newValues, value)
		}
	}

	series.Values = newValues
}

// trimOldHealthValues removes health values older than retention period
func (md *MonitoringDashboard) trimOldHealthValues(series *HealthSeries) {
	cutoff := time.Now().Add(-md.config.RetentionPeriod)
	var newValues []HealthValue

	for _, value := range series.Values {
		if value.Timestamp.After(cutoff) {
			newValues = append(newValues, value)
		}
	}

	series.Values = newValues
}

// startHTTPServer starts the HTTP server for the dashboard
func (md *MonitoringDashboard) startHTTPServer() {
	mux := http.NewServeMux()

	// Dashboard API endpoints
	mux.HandleFunc("/api/dashboard/data", md.handleDashboardData)
	mux.HandleFunc("/api/dashboard/summary", md.handleSystemSummary)
	mux.HandleFunc("/api/dashboard/metrics", md.handleMetricsData)
	mux.HandleFunc("/api/dashboard/health", md.handleHealthData)
	mux.HandleFunc("/api/dashboard/alerts", md.handleAlertsData)

	// WebSocket endpoint
	if md.config.EnableWebSocket {
		mux.HandleFunc("/ws/dashboard", md.handleWebSocket)
	}

	// Static files (dashboard UI)
	mux.Handle("/", http.FileServer(http.Dir("./web/dashboard/")))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", md.config.Port),
		Handler: mux,
	}

	log.Info().
		Int("port", md.config.Port).
		Msg("Dashboard HTTP server listening")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error().Err(err).Msg("Dashboard HTTP server error")
	}
}

// handleDashboardData handles requests for complete dashboard data
func (md *MonitoringDashboard) handleDashboardData(w http.ResponseWriter, r *http.Request) {
	md.dataAggregator.mutex.RLock()
	defer md.dataAggregator.mutex.RUnlock()

	data := &DashboardData{
		Timestamp: time.Now(),
		Metrics:   md.dataAggregator.metricsData,
		Health:    md.dataAggregator.healthData,
		Alerts:    md.dataAggregator.alertData,
		Summary:   md.generateSystemSummary(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// handleSystemSummary handles requests for system summary
func (md *MonitoringDashboard) handleSystemSummary(w http.ResponseWriter, r *http.Request) {
	summary := md.generateSystemSummary()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// generateSystemSummary generates a system summary
func (md *MonitoringDashboard) generateSystemSummary() *SystemSummary {
	summary := &SystemSummary{
		LastUpdate: time.Now(),
	}

	// Calculate overall health
	healthyComponents := 0
	totalComponents := 0

	for _, healthSeries := range md.dataAggregator.healthData {
		totalComponents++
		if healthSeries.Status == "healthy" {
			healthyComponents++
		}
	}

	if totalComponents > 0 {
		healthRatio := float64(healthyComponents) / float64(totalComponents)
		if healthRatio >= 0.9 {
			summary.OverallHealth = "healthy"
		} else if healthRatio >= 0.7 {
			summary.OverallHealth = "degraded"
		} else {
			summary.OverallHealth = "unhealthy"
		}
	} else {
		summary.OverallHealth = "unknown"
	}

	summary.TotalNodes = totalComponents
	summary.HealthyNodes = healthyComponents

	// Count active alerts
	activeAlerts := 0
	for _, alertSeries := range md.dataAggregator.alertData {
		if len(alertSeries.Values) > 0 {
			lastAlert := alertSeries.Values[len(alertSeries.Values)-1]
			if lastAlert.Status == "firing" {
				activeAlerts++
			}
		}
	}
	summary.ActiveAlerts = activeAlerts

	return summary
}

// handleMetricsData handles requests for metrics data
func (md *MonitoringDashboard) handleMetricsData(w http.ResponseWriter, r *http.Request) {
	md.dataAggregator.mutex.RLock()
	defer md.dataAggregator.mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(md.dataAggregator.metricsData)
}

// handleHealthData handles requests for health data
func (md *MonitoringDashboard) handleHealthData(w http.ResponseWriter, r *http.Request) {
	md.dataAggregator.mutex.RLock()
	defer md.dataAggregator.mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(md.dataAggregator.healthData)
}

// handleAlertsData handles requests for alerts data
func (md *MonitoringDashboard) handleAlertsData(w http.ResponseWriter, r *http.Request) {
	md.dataAggregator.mutex.RLock()
	defer md.dataAggregator.mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(md.dataAggregator.alertData)
}

// handleWebSocket handles WebSocket connections for real-time updates
func (md *MonitoringDashboard) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins in development
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade WebSocket connection")
		return
	}

	clientID := fmt.Sprintf("client-%d", time.Now().UnixNano())

	md.connMutex.Lock()
	md.connections[clientID] = conn
	md.connMutex.Unlock()

	log.Info().Str("client_id", clientID).Msg("WebSocket client connected")

	// Handle client disconnection
	defer func() {
		md.connMutex.Lock()
		delete(md.connections, clientID)
		md.connMutex.Unlock()
		conn.Close()
		log.Info().Str("client_id", clientID).Msg("WebSocket client disconnected")
	}()

	// Keep connection alive and handle messages
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Err(err).Msg("WebSocket error")
			}
			break
		}
	}
}

// startWebSocketUpdates starts sending real-time updates to WebSocket clients
func (md *MonitoringDashboard) startWebSocketUpdates() {
	ticker := time.NewTicker(md.config.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-md.ctx.Done():
			return
		case <-ticker.C:
			md.broadcastUpdate()
		}
	}
}

// broadcastUpdate broadcasts dashboard updates to all connected WebSocket clients
func (md *MonitoringDashboard) broadcastUpdate() {
	md.dataAggregator.mutex.RLock()
	data := &DashboardData{
		Timestamp: time.Now(),
		Metrics:   md.dataAggregator.metricsData,
		Health:    md.dataAggregator.healthData,
		Alerts:    md.dataAggregator.alertData,
		Summary:   md.generateSystemSummary(),
	}
	md.dataAggregator.mutex.RUnlock()

	message, err := json.Marshal(data)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal dashboard data")
		return
	}

	md.connMutex.RLock()
	defer md.connMutex.RUnlock()

	for clientID, conn := range md.connections {
		err := conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Error().
				Err(err).
				Str("client_id", clientID).
				Msg("Failed to send WebSocket message")
		}
	}
}

// Shutdown gracefully shuts down the monitoring dashboard
func (md *MonitoringDashboard) Shutdown() error {
	md.cancel()

	// Close all WebSocket connections
	md.connMutex.Lock()
	for clientID, conn := range md.connections {
		conn.Close()
		log.Info().Str("client_id", clientID).Msg("Closed WebSocket connection")
	}
	md.connMutex.Unlock()

	log.Info().Msg("Monitoring dashboard stopped")
	return nil
}

// DefaultDashboardConfig returns a default dashboard configuration
func DefaultDashboardConfig() *DashboardConfig {
	return &DashboardConfig{
		Enabled:         true,
		Port:            8080,
		UpdateInterval:  5 * time.Second,
		RetentionPeriod: time.Hour,
		EnableWebSocket: true,
		EnableAlerts:    true,
		Theme:           "dark",
		DefaultPanels:   []string{"overview", "metrics", "health", "alerts"},
	}
}

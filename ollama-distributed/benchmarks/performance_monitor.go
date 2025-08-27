package benchmarks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// PerformanceMonitor provides real-time performance monitoring and alerting
type PerformanceMonitor struct {
	metrics          *RealTimeMetrics
	alertRules       []AlertRule
	subscribers      map[string]*websocket.Conn
	subscribersMutex sync.RWMutex
	stopCh           chan struct{}
	logger           Logger
}

// RealTimeMetrics holds current system performance metrics
type RealTimeMetrics struct {
	mutex sync.RWMutex
	
	// Throughput metrics
	RequestsPerSecond   float64    `json:"requests_per_second"`
	OperationsPerSecond float64    `json:"operations_per_second"`
	ErrorRate          float64    `json:"error_rate"`
	
	// Latency metrics (milliseconds)
	LatencyP50  float64 `json:"latency_p50"`
	LatencyP95  float64 `json:"latency_p95"`
	LatencyP99  float64 `json:"latency_p99"`
	LatencyMean float64 `json:"latency_mean"`
	
	// Resource utilization
	CPUUsagePercent    float64 `json:"cpu_usage_percent"`
	MemoryUsageMB      float64 `json:"memory_usage_mb"`
	MemoryUsagePercent float64 `json:"memory_usage_percent"`
	GoroutineCount     int     `json:"goroutine_count"`
	
	// Network metrics
	NetworkInMBps  float64 `json:"network_in_mbps"`
	NetworkOutMBps float64 `json:"network_out_mbps"`
	
	// Custom metrics
	CustomMetrics map[string]interface{} `json:"custom_metrics"`
	
	// Timestamps
	LastUpdate    time.Time `json:"last_update"`
	Uptime        float64   `json:"uptime_seconds"`
	
	// Historical data (last 100 points)
	History struct {
		Timestamps []time.Time `json:"timestamps"`
		CPU        []float64   `json:"cpu_history"`
		Memory     []float64   `json:"memory_history"`
		Throughput []float64   `json:"throughput_history"`
		Latency    []float64   `json:"latency_history"`
	} `json:"history"`
}

// AlertRule defines performance alerting rules
type AlertRule struct {
	Name        string                 `json:"name"`
	Condition   AlertCondition         `json:"condition"`
	Threshold   float64                `json:"threshold"`
	Duration    time.Duration          `json:"duration"`
	Severity    AlertSeverity          `json:"severity"`
	Callback    func(Alert)            `json:"-"`
	LastFired   time.Time              `json:"last_fired"`
	IsActive    bool                   `json:"is_active"`
}

// AlertCondition defines the type of alert condition
type AlertCondition string

const (
	AlertConditionGreaterThan AlertCondition = "greater_than"
	AlertConditionLessThan    AlertCondition = "less_than"
	AlertConditionEquals      AlertCondition = "equals"
	AlertConditionChange      AlertCondition = "change_percent"
)

// AlertSeverity defines alert severity levels
type AlertSeverity string

const (
	AlertSeverityCritical AlertSeverity = "critical"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityInfo     AlertSeverity = "info"
)

// Alert represents a fired alert
type Alert struct {
	RuleName    string        `json:"rule_name"`
	MetricName  string        `json:"metric_name"`
	Value       float64       `json:"value"`
	Threshold   float64       `json:"threshold"`
	Severity    AlertSeverity `json:"severity"`
	Timestamp   time.Time     `json:"timestamp"`
	Description string        `json:"description"`
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(logger Logger) *PerformanceMonitor {
	monitor := &PerformanceMonitor{
		metrics: &RealTimeMetrics{
			CustomMetrics: make(map[string]interface{}),
		},
		alertRules:  make([]AlertRule, 0),
		subscribers: make(map[string]*websocket.Conn),
		stopCh:      make(chan struct{}),
		logger:      logger,
	}
	
	// Initialize history slices
	monitor.metrics.History.Timestamps = make([]time.Time, 0, 100)
	monitor.metrics.History.CPU = make([]float64, 0, 100)
	monitor.metrics.History.Memory = make([]float64, 0, 100)
	monitor.metrics.History.Throughput = make([]float64, 0, 100)
	monitor.metrics.History.Latency = make([]float64, 0, 100)
	
	// Add default alert rules
	monitor.addDefaultAlertRules()
	
	return monitor
}

// Start begins performance monitoring
func (pm *PerformanceMonitor) Start(ctx context.Context) error {
	pm.logger.Info("Starting performance monitor")
	
	// Start metrics collection
	go pm.collectMetrics(ctx)
	
	// Start alert processing
	go pm.processAlerts(ctx)
	
	// Start HTTP server for dashboard
	go pm.startHTTPServer(ctx)
	
	return nil
}

// Stop stops the performance monitor
func (pm *PerformanceMonitor) Stop() error {
	pm.logger.Info("Stopping performance monitor")
	
	close(pm.stopCh)
	
	// Close all WebSocket connections
	pm.subscribersMutex.Lock()
	for _, conn := range pm.subscribers {
		conn.Close()
	}
	pm.subscribersMutex.Unlock()
	
	return nil
}

// collectMetrics collects system metrics periodically
func (pm *PerformanceMonitor) collectMetrics(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	startTime := time.Now()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-pm.stopCh:
			return
		case <-ticker.C:
			pm.updateMetrics(startTime)
			pm.broadcastMetrics()
		}
	}
}

// updateMetrics updates the current metrics
func (pm *PerformanceMonitor) updateMetrics(startTime time.Time) {
	pm.metrics.mutex.Lock()
	defer pm.metrics.mutex.Unlock()
	
	now := time.Now()
	
	// Update runtime metrics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	pm.metrics.MemoryUsageMB = float64(memStats.Alloc) / 1024 / 1024
	pm.metrics.MemoryUsagePercent = float64(memStats.Alloc) / float64(memStats.Sys) * 100
	pm.metrics.GoroutineCount = runtime.NumGoroutine()
	pm.metrics.CPUUsagePercent = pm.estimateCPUUsage() // Simplified CPU usage estimation
	
	// Update timestamps
	pm.metrics.LastUpdate = now
	pm.metrics.Uptime = now.Sub(startTime).Seconds()
	
	// Update history (keep last 100 points)
	if len(pm.metrics.History.Timestamps) >= 100 {
		// Remove oldest point
		pm.metrics.History.Timestamps = pm.metrics.History.Timestamps[1:]
		pm.metrics.History.CPU = pm.metrics.History.CPU[1:]
		pm.metrics.History.Memory = pm.metrics.History.Memory[1:]
		pm.metrics.History.Throughput = pm.metrics.History.Throughput[1:]
		pm.metrics.History.Latency = pm.metrics.History.Latency[1:]
	}
	
	// Add current point
	pm.metrics.History.Timestamps = append(pm.metrics.History.Timestamps, now)
	pm.metrics.History.CPU = append(pm.metrics.History.CPU, pm.metrics.CPUUsagePercent)
	pm.metrics.History.Memory = append(pm.metrics.History.Memory, pm.metrics.MemoryUsagePercent)
	pm.metrics.History.Throughput = append(pm.metrics.History.Throughput, pm.metrics.RequestsPerSecond)
	pm.metrics.History.Latency = append(pm.metrics.History.Latency, pm.metrics.LatencyMean)
}

// estimateCPUUsage provides a simple CPU usage estimation
func (pm *PerformanceMonitor) estimateCPUUsage() float64 {
	// This is a simplified estimation
	// In a real implementation, you'd use system calls or third-party libraries
	return float64(runtime.NumGoroutine()) * 0.1 // Rough approximation
}

// processAlerts checks alert rules and fires alerts
func (pm *PerformanceMonitor) processAlerts(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-pm.stopCh:
			return
		case <-ticker.C:
			pm.checkAlertRules()
		}
	}
}

// checkAlertRules evaluates all alert rules
func (pm *PerformanceMonitor) checkAlertRules() {
	pm.metrics.mutex.RLock()
	currentMetrics := *pm.metrics
	pm.metrics.mutex.RUnlock()
	
	for i := range pm.alertRules {
		rule := &pm.alertRules[i]
		shouldFire := pm.evaluateAlertRule(rule, &currentMetrics)
		
		if shouldFire && (!rule.IsActive || time.Since(rule.LastFired) > rule.Duration) {
			pm.fireAlert(rule, &currentMetrics)
		} else if !shouldFire && rule.IsActive {
			rule.IsActive = false
		}
	}
}

// evaluateAlertRule checks if an alert rule should fire
func (pm *PerformanceMonitor) evaluateAlertRule(rule *AlertRule, metrics *RealTimeMetrics) bool {
	var value float64
	
	// Get metric value based on rule name
	switch rule.Name {
	case "high_cpu":
		value = metrics.CPUUsagePercent
	case "high_memory":
		value = metrics.MemoryUsagePercent
	case "high_latency":
		value = metrics.LatencyP95
	case "high_error_rate":
		value = metrics.ErrorRate
	case "low_throughput":
		value = metrics.RequestsPerSecond
	default:
		return false
	}
	
	// Evaluate condition
	switch rule.Condition {
	case AlertConditionGreaterThan:
		return value > rule.Threshold
	case AlertConditionLessThan:
		return value < rule.Threshold
	case AlertConditionEquals:
		return value == rule.Threshold
	default:
		return false
	}
}

// fireAlert fires an alert
func (pm *PerformanceMonitor) fireAlert(rule *AlertRule, metrics *RealTimeMetrics) {
	rule.IsActive = true
	rule.LastFired = time.Now()
	
	alert := Alert{
		RuleName:    rule.Name,
		MetricName:  rule.Name,
		Threshold:   rule.Threshold,
		Severity:    rule.Severity,
		Timestamp:   time.Now(),
		Description: fmt.Sprintf("Alert: %s triggered", rule.Name),
	}
	
	// Get current value
	switch rule.Name {
	case "high_cpu":
		alert.Value = metrics.CPUUsagePercent
	case "high_memory":
		alert.Value = metrics.MemoryUsagePercent
	case "high_latency":
		alert.Value = metrics.LatencyP95
	case "high_error_rate":
		alert.Value = metrics.ErrorRate
	case "low_throughput":
		alert.Value = metrics.RequestsPerSecond
	}
	
	pm.logger.Error("Performance alert fired", 
		"rule", rule.Name, 
		"value", alert.Value, 
		"threshold", rule.Threshold,
		"severity", rule.Severity)
	
	// Execute callback if available
	if rule.Callback != nil {
		go rule.Callback(alert)
	}
	
	// Broadcast alert to subscribers
	pm.broadcastAlert(alert)
}

// addDefaultAlertRules adds default performance alert rules
func (pm *PerformanceMonitor) addDefaultAlertRules() {
	pm.alertRules = []AlertRule{
		{
			Name:      "high_cpu",
			Condition: AlertConditionGreaterThan,
			Threshold: 80.0,
			Duration:  30 * time.Second,
			Severity:  AlertSeverityWarning,
		},
		{
			Name:      "high_memory",
			Condition: AlertConditionGreaterThan,
			Threshold: 85.0,
			Duration:  30 * time.Second,
			Severity:  AlertSeverityWarning,
		},
		{
			Name:      "high_latency",
			Condition: AlertConditionGreaterThan,
			Threshold: 1000.0, // 1 second
			Duration:  10 * time.Second,
			Severity:  AlertSeverityCritical,
		},
		{
			Name:      "high_error_rate",
			Condition: AlertConditionGreaterThan,
			Threshold: 5.0, // 5%
			Duration:  10 * time.Second,
			Severity:  AlertSeverityCritical,
		},
		{
			Name:      "low_throughput",
			Condition: AlertConditionLessThan,
			Threshold: 10.0, // 10 req/s
			Duration:  60 * time.Second,
			Severity:  AlertSeverityWarning,
		},
	}
}

// startHTTPServer starts the HTTP server for the performance dashboard
func (pm *PerformanceMonitor) startHTTPServer(ctx context.Context) {
	mux := http.NewServeMux()
	
	// Serve static dashboard
	mux.HandleFunc("/", pm.handleDashboard)
	mux.HandleFunc("/api/metrics", pm.handleMetricsAPI)
	mux.HandleFunc("/api/alerts", pm.handleAlertsAPI)
	mux.HandleFunc("/ws", pm.handleWebSocket)
	
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	
	pm.logger.Info("Starting performance dashboard on :8080")
	
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			pm.logger.Error("Dashboard server error", "error", err)
		}
	}()
	
	// Shutdown server when context is done
	<-ctx.Done()
	server.Shutdown(context.Background())
}

// handleDashboard serves the performance dashboard HTML
func (pm *PerformanceMonitor) handleDashboard(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Performance Dashboard</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; }
        .metrics-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; margin-bottom: 30px; }
        .metric-card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .metric-value { font-size: 2em; font-weight: bold; color: #333; }
        .metric-label { color: #666; margin-top: 5px; }
        .chart-container { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); margin-bottom: 20px; }
        .alerts { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .alert { padding: 10px; margin-bottom: 10px; border-radius: 4px; }
        .alert.critical { background-color: #fee; border-left: 4px solid #f00; }
        .alert.warning { background-color: #ffeaa7; border-left: 4px solid #fdcb6e; }
        .status-indicator { display: inline-block; width: 10px; height: 10px; border-radius: 50%; margin-right: 10px; }
        .status-good { background-color: #00b894; }
        .status-warning { background-color: #fdcb6e; }
        .status-critical { background-color: #e84393; }
        h1 { color: #333; margin-bottom: 30px; }
        h2 { color: #333; margin-bottom: 20px; }
    </style>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>
<body>
    <div class="container">
        <h1>🚀 Performance Dashboard - Ollama Distributed</h1>
        
        <div class="metrics-grid">
            <div class="metric-card">
                <div class="metric-value" id="throughput">--</div>
                <div class="metric-label">Requests/Second</div>
            </div>
            <div class="metric-card">
                <div class="metric-value" id="latency">--</div>
                <div class="metric-label">Avg Latency (ms)</div>
            </div>
            <div class="metric-card">
                <div class="metric-value" id="cpu">--</div>
                <div class="metric-label">CPU Usage (%)</div>
            </div>
            <div class="metric-card">
                <div class="metric-value" id="memory">--</div>
                <div class="metric-label">Memory Usage (%)</div>
            </div>
            <div class="metric-card">
                <div class="metric-value" id="goroutines">--</div>
                <div class="metric-label">Active Goroutines</div>
            </div>
            <div class="metric-card">
                <div class="metric-value" id="errors">--</div>
                <div class="metric-label">Error Rate (%)</div>
            </div>
        </div>
        
        <div class="chart-container">
            <h2>Performance Trends</h2>
            <canvas id="performanceChart" width="400" height="200"></canvas>
        </div>
        
        <div class="alerts">
            <h2>Active Alerts</h2>
            <div id="alertsList">No active alerts</div>
        </div>
    </div>

    <script>
        // WebSocket connection
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const ws = new WebSocket(protocol + '//' + window.location.host + '/ws');
        
        // Chart setup
        const ctx = document.getElementById('performanceChart').getContext('2d');
        const chart = new Chart(ctx, {
            type: 'line',
            data: {
                labels: [],
                datasets: [{
                    label: 'CPU %',
                    data: [],
                    borderColor: 'rgb(255, 99, 132)',
                    tension: 0.1
                }, {
                    label: 'Memory %',
                    data: [],
                    borderColor: 'rgb(54, 162, 235)',
                    tension: 0.1
                }, {
                    label: 'Throughput',
                    data: [],
                    borderColor: 'rgb(75, 192, 192)',
                    tension: 0.1,
                    yAxisID: 'y1'
                }]
            },
            options: {
                responsive: true,
                scales: {
                    y: {
                        type: 'linear',
                        display: true,
                        position: 'left',
                    },
                    y1: {
                        type: 'linear',
                        display: true,
                        position: 'right',
                        grid: {
                            drawOnChartArea: false,
                        },
                    }
                }
            }
        });
        
        // Update metrics display
        function updateMetrics(data) {
            document.getElementById('throughput').textContent = data.requests_per_second.toFixed(1);
            document.getElementById('latency').textContent = data.latency_mean.toFixed(1);
            document.getElementById('cpu').textContent = data.cpu_usage_percent.toFixed(1);
            document.getElementById('memory').textContent = data.memory_usage_percent.toFixed(1);
            document.getElementById('goroutines').textContent = data.goroutine_count;
            document.getElementById('errors').textContent = data.error_rate.toFixed(2);
            
            // Update chart
            if (data.history && data.history.timestamps.length > 0) {
                const labels = data.history.timestamps.map(ts => new Date(ts).toLocaleTimeString());
                chart.data.labels = labels.slice(-30); // Show last 30 points
                chart.data.datasets[0].data = data.history.cpu_history.slice(-30);
                chart.data.datasets[1].data = data.history.memory_history.slice(-30);
                chart.data.datasets[2].data = data.history.throughput_history.slice(-30);
                chart.update('none');
            }
        }
        
        // WebSocket message handling
        ws.onmessage = function(event) {
            const data = JSON.parse(event.data);
            if (data.type === 'metrics') {
                updateMetrics(data.data);
            } else if (data.type === 'alert') {
                addAlert(data.data);
            }
        };
        
        // Add alert to display
        function addAlert(alert) {
            const alertsList = document.getElementById('alertsList');
            const alertDiv = document.createElement('div');
            alertDiv.className = 'alert ' + alert.severity;
            alertDiv.innerHTML = 
                '<span class="status-indicator status-' + alert.severity + '"></span>' +
                '<strong>' + alert.rule_name + '</strong>: ' + alert.description +
                ' (Value: ' + alert.value.toFixed(2) + ', Threshold: ' + alert.threshold + ')' +
                ' <small>' + new Date(alert.timestamp).toLocaleString() + '</small>';
            alertsList.appendChild(alertDiv);
            
            // Remove old alerts (keep last 10)
            while (alertsList.children.length > 10) {
                alertsList.removeChild(alertsList.firstChild);
            }
        }
        
        // Initial load
        fetch('/api/metrics')
            .then(response => response.json())
            .then(data => updateMetrics(data))
            .catch(err => console.error('Failed to load initial metrics:', err));
    </script>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// handleMetricsAPI serves current metrics as JSON
func (pm *PerformanceMonitor) handleMetricsAPI(w http.ResponseWriter, r *http.Request) {
	pm.metrics.mutex.RLock()
	defer pm.metrics.mutex.RUnlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pm.metrics)
}

// handleAlertsAPI serves current alert rules as JSON
func (pm *PerformanceMonitor) handleAlertsAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pm.alertRules)
}

// handleWebSocket handles WebSocket connections for real-time updates
func (pm *PerformanceMonitor) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		pm.logger.Error("WebSocket upgrade failed", "error", err)
		return
	}
	
	// Generate unique subscriber ID
	subscriberID := fmt.Sprintf("sub_%d", time.Now().UnixNano())
	
	pm.subscribersMutex.Lock()
	pm.subscribers[subscriberID] = conn
	pm.subscribersMutex.Unlock()
	
	pm.logger.Info("WebSocket client connected", "id", subscriberID)
	
	// Handle client disconnect
	defer func() {
		pm.subscribersMutex.Lock()
		delete(pm.subscribers, subscriberID)
		pm.subscribersMutex.Unlock()
		conn.Close()
		pm.logger.Info("WebSocket client disconnected", "id", subscriberID)
	}()
	
	// Keep connection alive and handle client messages
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// broadcastMetrics broadcasts current metrics to all subscribers
func (pm *PerformanceMonitor) broadcastMetrics() {
	pm.metrics.mutex.RLock()
	metricsData := *pm.metrics
	pm.metrics.mutex.RUnlock()
	
	message := map[string]interface{}{
		"type": "metrics",
		"data": metricsData,
	}
	
	pm.broadcastMessage(message)
}

// broadcastAlert broadcasts an alert to all subscribers
func (pm *PerformanceMonitor) broadcastAlert(alert Alert) {
	message := map[string]interface{}{
		"type": "alert",
		"data": alert,
	}
	
	pm.broadcastMessage(message)
}

// broadcastMessage broadcasts a message to all WebSocket subscribers
func (pm *PerformanceMonitor) broadcastMessage(message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		pm.logger.Error("Failed to marshal broadcast message", "error", err)
		return
	}
	
	pm.subscribersMutex.RLock()
	subscribers := make(map[string]*websocket.Conn)
	for id, conn := range pm.subscribers {
		subscribers[id] = conn
	}
	pm.subscribersMutex.RUnlock()
	
	for id, conn := range subscribers {
		err := conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			pm.logger.Error("Failed to send WebSocket message", "subscriber", id, "error", err)
			
			// Remove failed connection
			pm.subscribersMutex.Lock()
			delete(pm.subscribers, id)
			pm.subscribersMutex.Unlock()
			
			conn.Close()
		}
	}
}

// UpdateMetric updates a specific metric value
func (pm *PerformanceMonitor) UpdateMetric(name string, value float64) {
	pm.metrics.mutex.Lock()
	defer pm.metrics.mutex.Unlock()
	
	switch name {
	case "requests_per_second":
		pm.metrics.RequestsPerSecond = value
	case "operations_per_second":
		pm.metrics.OperationsPerSecond = value
	case "error_rate":
		pm.metrics.ErrorRate = value
	case "latency_p50":
		pm.metrics.LatencyP50 = value
	case "latency_p95":
		pm.metrics.LatencyP95 = value
	case "latency_p99":
		pm.metrics.LatencyP99 = value
	case "latency_mean":
		pm.metrics.LatencyMean = value
	case "network_in_mbps":
		pm.metrics.NetworkInMBps = value
	case "network_out_mbps":
		pm.metrics.NetworkOutMBps = value
	default:
		pm.metrics.CustomMetrics[name] = value
	}
}

// GetCurrentMetrics returns the current metrics snapshot
func (pm *PerformanceMonitor) GetCurrentMetrics() RealTimeMetrics {
	pm.metrics.mutex.RLock()
	defer pm.metrics.mutex.RUnlock()
	
	return *pm.metrics
}

// AddAlertRule adds a new alert rule
func (pm *PerformanceMonitor) AddAlertRule(rule AlertRule) {
	pm.alertRules = append(pm.alertRules, rule)
	pm.logger.Info("Added alert rule", "name", rule.Name, "threshold", rule.Threshold)
}
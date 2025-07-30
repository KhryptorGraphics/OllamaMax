package observability

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MetricType represents different types of metrics
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

// Metric represents a single metric
type Metric struct {
	Name        string                 `json:"name"`
	Type        MetricType             `json:"type"`
	Value       float64                `json:"value"`
	Labels      map[string]string      `json:"labels"`
	Timestamp   time.Time              `json:"timestamp"`
	Description string                 `json:"description"`
}

// MetricsCollector collects and manages metrics
type MetricsCollector struct {
	metrics     map[string]*Metric
	counters    map[string]*Counter
	gauges      map[string]*Gauge
	histograms  map[string]*Histogram
	summaries   map[string]*Summary
	
	config      *MetricsConfig
	mu          sync.RWMutex
	
	// Background collection
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// MetricsConfig configures the metrics collector
type MetricsConfig struct {
	Namespace       string
	Subsystem       string
	CollectionInterval time.Duration
	RetentionPeriod    time.Duration
	MaxMetrics         int
	EnableAutoCleanup  bool
	
	// Export configuration
	EnableExport       bool
	ExportInterval     time.Duration
	ExportFormat       string
	ExportEndpoint     string
}

// Counter represents a monotonically increasing counter
type Counter struct {
	name        string
	value       float64
	labels      map[string]string
	description string
	mu          sync.Mutex
}

// Gauge represents a value that can go up and down
type Gauge struct {
	name        string
	value       float64
	labels      map[string]string
	description string
	mu          sync.Mutex
}

// Histogram represents a distribution of values
type Histogram struct {
	name        string
	buckets     []float64
	counts      []uint64
	sum         float64
	count       uint64
	labels      map[string]string
	description string
	mu          sync.Mutex
}

// Summary represents a summary of observations
type Summary struct {
	name        string
	quantiles   map[float64]float64
	sum         float64
	count       uint64
	labels      map[string]string
	description string
	mu          sync.Mutex
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(config *MetricsConfig) *MetricsCollector {
	if config == nil {
		config = &MetricsConfig{
			Namespace:          "ollama_distributed",
			CollectionInterval: 10 * time.Second,
			RetentionPeriod:    24 * time.Hour,
			MaxMetrics:         10000,
			EnableAutoCleanup:  true,
			EnableExport:       false,
			ExportInterval:     60 * time.Second,
			ExportFormat:       "prometheus",
		}
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	mc := &MetricsCollector{
		metrics:    make(map[string]*Metric),
		counters:   make(map[string]*Counter),
		gauges:     make(map[string]*Gauge),
		histograms: make(map[string]*Histogram),
		summaries:  make(map[string]*Summary),
		config:     config,
		ctx:        ctx,
		cancel:     cancel,
	}
	
	// Start background tasks
	mc.wg.Add(2)
	go mc.collectionLoop()
	go mc.cleanupLoop()
	
	if config.EnableExport {
		mc.wg.Add(1)
		go mc.exportLoop()
	}
	
	return mc
}

// NewCounter creates a new counter metric
func (mc *MetricsCollector) NewCounter(name, description string, labels map[string]string) *Counter {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	fullName := mc.getFullName(name)
	
	counter := &Counter{
		name:        fullName,
		labels:      labels,
		description: description,
	}
	
	mc.counters[fullName] = counter
	return counter
}

// NewGauge creates a new gauge metric
func (mc *MetricsCollector) NewGauge(name, description string, labels map[string]string) *Gauge {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	fullName := mc.getFullName(name)
	
	gauge := &Gauge{
		name:        fullName,
		labels:      labels,
		description: description,
	}
	
	mc.gauges[fullName] = gauge
	return gauge
}

// NewHistogram creates a new histogram metric
func (mc *MetricsCollector) NewHistogram(name, description string, buckets []float64, labels map[string]string) *Histogram {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	fullName := mc.getFullName(name)
	
	if buckets == nil {
		buckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}
	}
	
	histogram := &Histogram{
		name:        fullName,
		buckets:     buckets,
		counts:      make([]uint64, len(buckets)+1),
		labels:      labels,
		description: description,
	}
	
	mc.histograms[fullName] = histogram
	return histogram
}

// NewSummary creates a new summary metric
func (mc *MetricsCollector) NewSummary(name, description string, quantiles []float64, labels map[string]string) *Summary {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	fullName := mc.getFullName(name)
	
	if quantiles == nil {
		quantiles = []float64{0.5, 0.9, 0.95, 0.99}
	}
	
	quantileMap := make(map[float64]float64)
	for _, q := range quantiles {
		quantileMap[q] = 0
	}
	
	summary := &Summary{
		name:        fullName,
		quantiles:   quantileMap,
		labels:      labels,
		description: description,
	}
	
	mc.summaries[fullName] = summary
	return summary
}

// getFullName constructs the full metric name
func (mc *MetricsCollector) getFullName(name string) string {
	if mc.config.Subsystem != "" {
		return fmt.Sprintf("%s_%s_%s", mc.config.Namespace, mc.config.Subsystem, name)
	}
	return fmt.Sprintf("%s_%s", mc.config.Namespace, name)
}

// GetAllMetrics returns all current metrics
func (mc *MetricsCollector) GetAllMetrics() map[string]*Metric {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	result := make(map[string]*Metric)
	for name, metric := range mc.metrics {
		result[name] = metric
	}
	
	return result
}

// collectionLoop periodically collects metrics
func (mc *MetricsCollector) collectionLoop() {
	defer mc.wg.Done()
	
	ticker := time.NewTicker(mc.config.CollectionInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-mc.ctx.Done():
			return
		case <-ticker.C:
			mc.collectMetrics()
		}
	}
}

// collectMetrics collects current metric values
func (mc *MetricsCollector) collectMetrics() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	now := time.Now()
	
	// Collect counter metrics
	for name, counter := range mc.counters {
		counter.mu.Lock()
		mc.metrics[name] = &Metric{
			Name:        name,
			Type:        MetricTypeCounter,
			Value:       counter.value,
			Labels:      counter.labels,
			Timestamp:   now,
			Description: counter.description,
		}
		counter.mu.Unlock()
	}
	
	// Collect gauge metrics
	for name, gauge := range mc.gauges {
		gauge.mu.Lock()
		mc.metrics[name] = &Metric{
			Name:        name,
			Type:        MetricTypeGauge,
			Value:       gauge.value,
			Labels:      gauge.labels,
			Timestamp:   now,
			Description: gauge.description,
		}
		gauge.mu.Unlock()
	}
	
	// Collect histogram metrics
	for name, histogram := range mc.histograms {
		histogram.mu.Lock()
		mc.metrics[name] = &Metric{
			Name:        name,
			Type:        MetricTypeHistogram,
			Value:       histogram.sum,
			Labels:      histogram.labels,
			Timestamp:   now,
			Description: histogram.description,
		}
		histogram.mu.Unlock()
	}
	
	// Collect summary metrics
	for name, summary := range mc.summaries {
		summary.mu.Lock()
		mc.metrics[name] = &Metric{
			Name:        name,
			Type:        MetricTypeSummary,
			Value:       summary.sum,
			Labels:      summary.labels,
			Timestamp:   now,
			Description: summary.description,
		}
		summary.mu.Unlock()
	}
}

// cleanupLoop periodically cleans up old metrics
func (mc *MetricsCollector) cleanupLoop() {
	defer mc.wg.Done()
	
	if !mc.config.EnableAutoCleanup {
		return
	}
	
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	
	for {
		select {
		case <-mc.ctx.Done():
			return
		case <-ticker.C:
			mc.cleanupOldMetrics()
		}
	}
}

// cleanupOldMetrics removes old metrics
func (mc *MetricsCollector) cleanupOldMetrics() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	cutoff := time.Now().Add(-mc.config.RetentionPeriod)
	
	for name, metric := range mc.metrics {
		if metric.Timestamp.Before(cutoff) {
			delete(mc.metrics, name)
		}
	}
}

// exportLoop periodically exports metrics
func (mc *MetricsCollector) exportLoop() {
	defer mc.wg.Done()
	
	ticker := time.NewTicker(mc.config.ExportInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-mc.ctx.Done():
			return
		case <-ticker.C:
			mc.exportMetrics()
		}
	}
}

// exportMetrics exports metrics to external systems
func (mc *MetricsCollector) exportMetrics() {
	// Implementation would export metrics to Prometheus, InfluxDB, etc.
	// For now, this is a placeholder
}

// Close closes the metrics collector
func (mc *MetricsCollector) Close() error {
	mc.cancel()
	mc.wg.Wait()
	return nil
}

// Counter methods

// Inc increments the counter by 1
func (c *Counter) Inc() {
	c.Add(1)
}

// Add adds the given value to the counter
func (c *Counter) Add(value float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value += value
}

// Get returns the current counter value
func (c *Counter) Get() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

// Gauge methods

// Set sets the gauge to the given value
func (g *Gauge) Set(value float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value = value
}

// Inc increments the gauge by 1
func (g *Gauge) Inc() {
	g.Add(1)
}

// Dec decrements the gauge by 1
func (g *Gauge) Dec() {
	g.Add(-1)
}

// Add adds the given value to the gauge
func (g *Gauge) Add(value float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value += value
}

// Get returns the current gauge value
func (g *Gauge) Get() float64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.value
}

// Histogram methods

// Observe adds an observation to the histogram
func (h *Histogram) Observe(value float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	h.sum += value
	h.count++
	
	// Find the appropriate bucket
	for i, bucket := range h.buckets {
		if value <= bucket {
			h.counts[i]++
			return
		}
	}
	
	// Value is greater than all buckets
	h.counts[len(h.buckets)]++
}

// GetSum returns the sum of all observations
func (h *Histogram) GetSum() float64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.sum
}

// GetCount returns the count of all observations
func (h *Histogram) GetCount() uint64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.count
}

// Summary methods

// Observe adds an observation to the summary
func (s *Summary) Observe(value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.sum += value
	s.count++
	
	// Update quantiles (simplified implementation)
	// In production, you'd use a more sophisticated algorithm like t-digest
	for q := range s.quantiles {
		s.quantiles[q] = value // Simplified
	}
}

// GetSum returns the sum of all observations
func (s *Summary) GetSum() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sum
}

// GetCount returns the count of all observations
func (s *Summary) GetCount() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.count
}

// GetQuantile returns the value for a given quantile
func (s *Summary) GetQuantile(quantile float64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.quantiles[quantile]
}

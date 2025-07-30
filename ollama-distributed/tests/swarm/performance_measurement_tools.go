package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// PerformanceMeter provides comprehensive performance measurement for swarm operations
type PerformanceMeter struct {
	startTime    time.Time
	metrics      *PerformanceMetrics
	collectors   map[string]MetricsCollector
	mu           sync.RWMutex
	isRecording  int32 // atomic flag
}

// PerformanceMetrics holds comprehensive performance data
type PerformanceMetrics struct {
	// Timing metrics
	TotalDuration       time.Duration
	AverageResponseTime time.Duration
	MinResponseTime     time.Duration
	MaxResponseTime     time.Duration
	P95ResponseTime     time.Duration
	P99ResponseTime     time.Duration

	// Throughput metrics
	TotalOperations     int64
	SuccessfulOps       int64
	FailedOps          int64
	OperationsPerSecond float64

	// Resource metrics
	PeakMemoryUsage     uint64
	AverageMemoryUsage  uint64
	CPUUtilization      float64
	GoroutineCount      int

	// Network metrics
	NetworkLatency      time.Duration
	MessagesSent        int64
	MessagesReceived    int64
	BytesSent          int64
	BytesReceived      int64

	// Coordination metrics
	LockContentions     int64
	LockWaitTime       time.Duration
	CoordinationEvents  int64
	ConflictResolutions int64

	// Error metrics
	ErrorRate          float64
	TimeoutCount       int64
	RetryCount         int64

	// Custom metrics
	CustomMetrics map[string]interface{}
}

// MetricsCollector interface for different types of metrics collectors
type MetricsCollector interface {
	Name() string
	Start(ctx context.Context) error
	Stop() error
	GetMetrics() map[string]interface{}
	Reset()
}

// TimingCollector collects timing and latency metrics
type TimingCollector struct {
	measurements []time.Duration
	startTimes   map[string]time.Time
	mu           sync.RWMutex
	isActive     bool
}

// ThroughputCollector collects throughput and operation metrics
type ThroughputCollector struct {
	operations    int64
	successful    int64
	failed        int64
	startTime     time.Time
	mu            sync.RWMutex
	isActive      bool
}

// ResourceCollector collects system resource metrics
type ResourceCollector struct {
	memorySnapshots []uint64
	cpuUsage        []float64
	goroutineCount  []int
	ticker          *time.Ticker
	stopChan        chan struct{}
	mu              sync.RWMutex
	isActive        bool
}

// NetworkCollector collects network and communication metrics
type NetworkCollector struct {
	latencies       []time.Duration
	messagesSent    int64
	messagesRecv    int64
	bytesSent       int64
	bytesRecv       int64
	mu              sync.RWMutex
	isActive        bool
}

// CoordinationCollector collects coordination and synchronization metrics
type CoordinationCollector struct {
	lockContentions    int64
	lockWaitTimes      []time.Duration
	coordEvents        int64
	conflicts          int64
	mu                 sync.RWMutex
	isActive           bool
}

// NewPerformanceMeter creates a new performance meter
func NewPerformanceMeter() *PerformanceMeter {
	pm := &PerformanceMeter{
		metrics:     &PerformanceMetrics{CustomMetrics: make(map[string]interface{})},
		collectors:  make(map[string]MetricsCollector),
	}

	// Register default collectors
	pm.RegisterCollector(&TimingCollector{
		measurements: make([]time.Duration, 0),
		startTimes:   make(map[string]time.Time),
	})
	pm.RegisterCollector(&ThroughputCollector{})
	pm.RegisterCollector(&ResourceCollector{
		memorySnapshots: make([]uint64, 0),
		cpuUsage:        make([]float64, 0),
		goroutineCount:  make([]int, 0),
		stopChan:        make(chan struct{}),
	})
	pm.RegisterCollector(&NetworkCollector{
		latencies: make([]time.Duration, 0),
	})
	pm.RegisterCollector(&CoordinationCollector{
		lockWaitTimes: make([]time.Duration, 0),
	})

	return pm
}

// RegisterCollector registers a new metrics collector
func (pm *PerformanceMeter) RegisterCollector(collector MetricsCollector) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.collectors[collector.Name()] = collector
}

// StartRecording starts performance recording
func (pm *PerformanceMeter) StartRecording(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&pm.isRecording, 0, 1) {
		return fmt.Errorf("performance recording already active")
	}

	pm.startTime = time.Now()

	// Start all collectors
	for _, collector := range pm.collectors {
		if err := collector.Start(ctx); err != nil {
			return fmt.Errorf("failed to start collector %s: %w", collector.Name(), err)
		}
	}

	return nil
}

// StopRecording stops performance recording and calculates final metrics
func (pm *PerformanceMeter) StopRecording() *PerformanceMetrics {
	if !atomic.CompareAndSwapInt32(&pm.isRecording, 1, 0) {
		return pm.metrics
	}

	pm.metrics.TotalDuration = time.Since(pm.startTime)

	// Stop all collectors and aggregate metrics
	for _, collector := range pm.collectors {
		collector.Stop()
		pm.aggregateMetrics(collector)
	}

	pm.calculateDerivedMetrics()
	return pm.metrics
}

// RecordOperation records a single operation for timing
func (pm *PerformanceMeter) RecordOperation(operationID string, duration time.Duration, success bool) {
	if atomic.LoadInt32(&pm.isRecording) == 0 {
		return
	}

	// Update timing collector
	if timing, ok := pm.collectors["timing"].(*TimingCollector); ok {
		timing.RecordMeasurement(duration)
	}

	// Update throughput collector
	if throughput, ok := pm.collectors["throughput"].(*ThroughputCollector); ok {
		throughput.RecordOperation(success)
	}
}

// RecordNetworkLatency records network latency
func (pm *PerformanceMeter) RecordNetworkLatency(latency time.Duration) {
	if atomic.LoadInt32(&pm.isRecording) == 0 {
		return
	}

	if network, ok := pm.collectors["network"].(*NetworkCollector); ok {
		network.RecordLatency(latency)
	}
}

// RecordLockContention records lock contention events
func (pm *PerformanceMeter) RecordLockContention(waitTime time.Duration) {
	if atomic.LoadInt32(&pm.isRecording) == 0 {
		return
	}

	if coord, ok := pm.collectors["coordination"].(*CoordinationCollector); ok {
		coord.RecordLockContention(waitTime)
	}
}

// aggregateMetrics aggregates metrics from a collector
func (pm *PerformanceMeter) aggregateMetrics(collector MetricsCollector) {
	metrics := collector.GetMetrics()

	switch collector.Name() {
	case "timing":
		if measurements, ok := metrics["measurements"].([]time.Duration); ok && len(measurements) > 0 {
			pm.metrics.AverageResponseTime = calculateAverage(measurements)
			pm.metrics.MinResponseTime = calculateMin(measurements)
			pm.metrics.MaxResponseTime = calculateMax(measurements)
			pm.metrics.P95ResponseTime = calculatePercentile(measurements, 0.95)
			pm.metrics.P99ResponseTime = calculatePercentile(measurements, 0.99)
		}

	case "throughput":
		if total, ok := metrics["total"].(int64); ok {
			pm.metrics.TotalOperations = total
		}
		if successful, ok := metrics["successful"].(int64); ok {
			pm.metrics.SuccessfulOps = successful
		}
		if failed, ok := metrics["failed"].(int64); ok {
			pm.metrics.FailedOps = failed
		}

	case "resource":
		if memory, ok := metrics["peak_memory"].(uint64); ok {
			pm.metrics.PeakMemoryUsage = memory
		}
		if avgMemory, ok := metrics["avg_memory"].(uint64); ok {
			pm.metrics.AverageMemoryUsage = avgMemory
		}
		if cpu, ok := metrics["cpu_usage"].(float64); ok {
			pm.metrics.CPUUtilization = cpu
		}
		if goroutines, ok := metrics["goroutines"].(int); ok {
			pm.metrics.GoroutineCount = goroutines
		}

	case "network":
		if latency, ok := metrics["avg_latency"].(time.Duration); ok {
			pm.metrics.NetworkLatency = latency
		}
		if sent, ok := metrics["messages_sent"].(int64); ok {
			pm.metrics.MessagesSent = sent
		}
		if recv, ok := metrics["messages_received"].(int64); ok {
			pm.metrics.MessagesReceived = recv
		}

	case "coordination":
		if contentions, ok := metrics["lock_contentions"].(int64); ok {
			pm.metrics.LockContentions = contentions
		}
		if waitTime, ok := metrics["avg_lock_wait"].(time.Duration); ok {
			pm.metrics.LockWaitTime = waitTime
		}
		if events, ok := metrics["coordination_events"].(int64); ok {
			pm.metrics.CoordinationEvents = events
		}
	}
}

// calculateDerivedMetrics calculates derived metrics
func (pm *PerformanceMeter) calculateDerivedMetrics() {
	if pm.metrics.TotalDuration > 0 {
		pm.metrics.OperationsPerSecond = float64(pm.metrics.TotalOperations) / pm.metrics.TotalDuration.Seconds()
	}

	if pm.metrics.TotalOperations > 0 {
		pm.metrics.ErrorRate = float64(pm.metrics.FailedOps) / float64(pm.metrics.TotalOperations) * 100
	}
}

// TimingCollector implementation

func (tc *TimingCollector) Name() string {
	return "timing"
}

func (tc *TimingCollector) Start(ctx context.Context) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.isActive = true
	return nil
}

func (tc *TimingCollector) Stop() error {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.isActive = false
	return nil
}

func (tc *TimingCollector) GetMetrics() map[string]interface{} {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	
	return map[string]interface{}{
		"measurements": tc.measurements,
		"count":        len(tc.measurements),
	}
}

func (tc *TimingCollector) Reset() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.measurements = make([]time.Duration, 0)
	tc.startTimes = make(map[string]time.Time)
}

func (tc *TimingCollector) RecordMeasurement(duration time.Duration) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if tc.isActive {
		tc.measurements = append(tc.measurements, duration)
	}
}

func (tc *TimingCollector) StartOperation(operationID string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if tc.isActive {
		tc.startTimes[operationID] = time.Now()
	}
}

func (tc *TimingCollector) EndOperation(operationID string) time.Duration {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	
	if startTime, exists := tc.startTimes[operationID]; exists {
		duration := time.Since(startTime)
		tc.measurements = append(tc.measurements, duration)
		delete(tc.startTimes, operationID)
		return duration
	}
	
	return 0
}

// ThroughputCollector implementation

func (tpc *ThroughputCollector) Name() string {
	return "throughput"
}

func (tpc *ThroughputCollector) Start(ctx context.Context) error {
	tpc.mu.Lock()
	defer tpc.mu.Unlock()
	tpc.startTime = time.Now()
	tpc.isActive = true
	return nil
}

func (tpc *ThroughputCollector) Stop() error {
	tpc.mu.Lock()
	defer tpc.mu.Unlock()
	tpc.isActive = false
	return nil
}

func (tpc *ThroughputCollector) GetMetrics() map[string]interface{} {
	tpc.mu.RLock()
	defer tpc.mu.RUnlock()
	
	return map[string]interface{}{
		"total":      atomic.LoadInt64(&tpc.operations),
		"successful": atomic.LoadInt64(&tpc.successful),
		"failed":     atomic.LoadInt64(&tpc.failed),
		"duration":   time.Since(tpc.startTime),
	}
}

func (tpc *ThroughputCollector) Reset() {
	atomic.StoreInt64(&tpc.operations, 0)
	atomic.StoreInt64(&tpc.successful, 0)
	atomic.StoreInt64(&tpc.failed, 0)
	tpc.startTime = time.Now()
}

func (tpc *ThroughputCollector) RecordOperation(success bool) {
	if tpc.isActive {
		atomic.AddInt64(&tpc.operations, 1)
		if success {
			atomic.AddInt64(&tpc.successful, 1)
		} else {
			atomic.AddInt64(&tpc.failed, 1)
		}
	}
}

// ResourceCollector implementation

func (rc *ResourceCollector) Name() string {
	return "resource"
}

func (rc *ResourceCollector) Start(ctx context.Context) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	
	rc.isActive = true
	rc.ticker = time.NewTicker(100 * time.Millisecond)
	
	go rc.collectMetrics(ctx)
	
	return nil
}

func (rc *ResourceCollector) Stop() error {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	
	rc.isActive = false
	if rc.ticker != nil {
		rc.ticker.Stop()
	}
	
	select {
	case rc.stopChan <- struct{}{}:
	default:
	}
	
	return nil
}

func (rc *ResourceCollector) GetMetrics() map[string]interface{} {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	
	var peakMemory, avgMemory uint64
	if len(rc.memorySnapshots) > 0 {
		peakMemory = rc.memorySnapshots[0]
		var totalMemory uint64
		for _, mem := range rc.memorySnapshots {
			if mem > peakMemory {
				peakMemory = mem
			}
			totalMemory += mem
		}
		avgMemory = totalMemory / uint64(len(rc.memorySnapshots))
	}
	
	var avgCPU float64
	if len(rc.cpuUsage) > 0 {
		var totalCPU float64
		for _, cpu := range rc.cpuUsage {
			totalCPU += cpu
		}
		avgCPU = totalCPU / float64(len(rc.cpuUsage))
	}
	
	currentGoroutines := runtime.NumGoroutine()
	
	return map[string]interface{}{
		"peak_memory":  peakMemory,
		"avg_memory":   avgMemory,
		"cpu_usage":    avgCPU,
		"goroutines":   currentGoroutines,
	}
}

func (rc *ResourceCollector) Reset() {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.memorySnapshots = make([]uint64, 0)
	rc.cpuUsage = make([]float64, 0)
	rc.goroutineCount = make([]int, 0)
}

func (rc *ResourceCollector) collectMetrics(ctx context.Context) {
	for {
		select {
		case <-rc.ticker.C:
			rc.mu.Lock()
			if rc.isActive {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				rc.memorySnapshots = append(rc.memorySnapshots, m.Alloc)
				rc.goroutineCount = append(rc.goroutineCount, runtime.NumGoroutine())
				// CPU usage would require more complex implementation
				rc.cpuUsage = append(rc.cpuUsage, 0.0) // Placeholder
			}
			rc.mu.Unlock()
		case <-rc.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

// NetworkCollector implementation

func (nc *NetworkCollector) Name() string {
	return "network"
}

func (nc *NetworkCollector) Start(ctx context.Context) error {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	nc.isActive = true
	return nil
}

func (nc *NetworkCollector) Stop() error {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	nc.isActive = false
	return nil
}

func (nc *NetworkCollector) GetMetrics() map[string]interface{} {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	
	var avgLatency time.Duration
	if len(nc.latencies) > 0 {
		avgLatency = calculateAverage(nc.latencies)
	}
	
	return map[string]interface{}{
		"avg_latency":        avgLatency,
		"messages_sent":      atomic.LoadInt64(&nc.messagesSent),
		"messages_received":  atomic.LoadInt64(&nc.messagesRecv),
		"bytes_sent":         atomic.LoadInt64(&nc.bytesSent),
		"bytes_received":     atomic.LoadInt64(&nc.bytesRecv),
	}
}

func (nc *NetworkCollector) Reset() {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	nc.latencies = make([]time.Duration, 0)
	atomic.StoreInt64(&nc.messagesSent, 0)
	atomic.StoreInt64(&nc.messagesRecv, 0)
	atomic.StoreInt64(&nc.bytesSent, 0)
	atomic.StoreInt64(&nc.bytesRecv, 0)
}

func (nc *NetworkCollector) RecordLatency(latency time.Duration) {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	if nc.isActive {
		nc.latencies = append(nc.latencies, latency)
	}
}

func (nc *NetworkCollector) RecordMessage(sent bool, bytes int64) {
	if nc.isActive {
		if sent {
			atomic.AddInt64(&nc.messagesSent, 1)
			atomic.AddInt64(&nc.bytesSent, bytes)
		} else {
			atomic.AddInt64(&nc.messagesRecv, 1)
			atomic.AddInt64(&nc.bytesRecv, bytes)
		}
	}
}

// CoordinationCollector implementation

func (cc *CoordinationCollector) Name() string {
	return "coordination"
}

func (cc *CoordinationCollector) Start(ctx context.Context) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.isActive = true
	return nil
}

func (cc *CoordinationCollector) Stop() error {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.isActive = false
	return nil
}

func (cc *CoordinationCollector) GetMetrics() map[string]interface{} {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	
	var avgWaitTime time.Duration
	if len(cc.lockWaitTimes) > 0 {
		avgWaitTime = calculateAverage(cc.lockWaitTimes)
	}
	
	return map[string]interface{}{
		"lock_contentions":     atomic.LoadInt64(&cc.lockContentions),
		"avg_lock_wait":        avgWaitTime,
		"coordination_events":  atomic.LoadInt64(&cc.coordEvents),
		"conflicts":            atomic.LoadInt64(&cc.conflicts),
	}
}

func (cc *CoordinationCollector) Reset() {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.lockWaitTimes = make([]time.Duration, 0)
	atomic.StoreInt64(&cc.lockContentions, 0)
	atomic.StoreInt64(&cc.coordEvents, 0)
	atomic.StoreInt64(&cc.conflicts, 0)
}

func (cc *CoordinationCollector) RecordLockContention(waitTime time.Duration) {
	if cc.isActive {
		cc.mu.Lock()
		cc.lockWaitTimes = append(cc.lockWaitTimes, waitTime)
		cc.mu.Unlock()
		atomic.AddInt64(&cc.lockContentions, 1)
	}
}

func (cc *CoordinationCollector) RecordCoordinationEvent() {
	if cc.isActive {
		atomic.AddInt64(&cc.coordEvents, 1)
	}
}

func (cc *CoordinationCollector) RecordConflict() {
	if cc.isActive {
		atomic.AddInt64(&cc.conflicts, 1)
	}
}

// Utility functions for statistical calculations

func calculateAverage(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	
	var total time.Duration
	for _, d := range durations {
		total += d
	}
	return total / time.Duration(len(durations))
}

func calculateMin(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	
	min := durations[0]
	for _, d := range durations[1:] {
		if d < min {
			min = d
		}
	}
	return min
}

func calculateMax(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	
	max := durations[0]
	for _, d := range durations[1:] {
		if d > max {
			max = d
		}
	}
	return max
}

func calculatePercentile(durations []time.Duration, percentile float64) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	
	// Simple percentile calculation - for production use, consider a proper sorting algorithm
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	
	// Basic bubble sort for simplicity
	for i := 0; i < len(sorted); i++ {
		for j := 0; j < len(sorted)-1; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}
	
	index := int(float64(len(sorted)-1) * percentile)
	return sorted[index]
}
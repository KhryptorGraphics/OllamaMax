package memory

import (
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

// Monitor tracks memory usage and provides statistics
type Monitor struct {
	config *Config
	stats  *Stats
	mu     sync.RWMutex
}

// Stats holds memory statistics
type Stats struct {
	// Memory usage
	AllocMB      int64 `json:"alloc_mb"`
	TotalAllocMB int64 `json:"total_alloc_mb"`
	SysMB        int64 `json:"sys_mb"`
	UsedMB       int64 `json:"used_mb"`

	// GC statistics
	NumGC        uint32    `json:"num_gc"`
	GCCPUPercent float64   `json:"gc_cpu_percent"`
	LastGC       time.Time `json:"last_gc"`

	// Heap statistics
	HeapAllocMB    int64  `json:"heap_alloc_mb"`
	HeapSysMB      int64  `json:"heap_sys_mb"`
	HeapIdleMB     int64  `json:"heap_idle_mb"`
	HeapInuseMB    int64  `json:"heap_inuse_mb"`
	HeapReleasedMB int64  `json:"heap_released_mb"`
	HeapObjects    uint64 `json:"heap_objects"`

	// Stack statistics
	StackInuseMB int64 `json:"stack_inuse_mb"`
	StackSysMB   int64 `json:"stack_sys_mb"`

	// Goroutine statistics
	NumGoroutine int `json:"num_goroutine"`

	// Timestamp
	Timestamp time.Time `json:"timestamp"`
}

// NewMonitor creates a new memory monitor
func NewMonitor(config *Config) *Monitor {
	return &Monitor{
		config: config,
		stats:  &Stats{},
	}
}

// Update updates memory statistics
func (m *Monitor) Update() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Convert bytes to MB
	const bytesToMB = 1024 * 1024

	m.stats.AllocMB = int64(memStats.Alloc / bytesToMB)
	m.stats.TotalAllocMB = int64(memStats.TotalAlloc / bytesToMB)
	m.stats.SysMB = int64(memStats.Sys / bytesToMB)
	m.stats.UsedMB = m.stats.AllocMB

	// GC statistics
	m.stats.NumGC = memStats.NumGC
	m.stats.GCCPUPercent = memStats.GCCPUFraction * 100
	if memStats.LastGC > 0 {
		m.stats.LastGC = time.Unix(0, int64(memStats.LastGC))
	}

	// Heap statistics
	m.stats.HeapAllocMB = int64(memStats.HeapAlloc / bytesToMB)
	m.stats.HeapSysMB = int64(memStats.HeapSys / bytesToMB)
	m.stats.HeapIdleMB = int64(memStats.HeapIdle / bytesToMB)
	m.stats.HeapInuseMB = int64(memStats.HeapInuse / bytesToMB)
	m.stats.HeapReleasedMB = int64(memStats.HeapReleased / bytesToMB)
	m.stats.HeapObjects = memStats.HeapObjects

	// Stack statistics
	m.stats.StackInuseMB = int64(memStats.StackInuse / bytesToMB)
	m.stats.StackSysMB = int64(memStats.StackSys / bytesToMB)

	// Goroutine statistics
	m.stats.NumGoroutine = runtime.NumGoroutine()

	// Timestamp
	m.stats.Timestamp = time.Now()
}

// GetStats returns current memory statistics
func (m *Monitor) GetStats() *Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to avoid race conditions
	statsCopy := *m.stats
	return &statsCopy
}

// IsMemoryHigh checks if memory usage is above warning threshold
func (m *Monitor) IsMemoryHigh() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.stats.UsedMB > m.config.WarningThresholdMB
}

// IsMemoryCritical checks if memory usage is above critical threshold
func (m *Monitor) IsMemoryCritical() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.stats.UsedMB > m.config.CriticalThresholdMB
}

// GetMemoryPressure returns memory pressure as a percentage (0-100)
func (m *Monitor) GetMemoryPressure() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.config.MaxMemoryMB <= 0 {
		return 0
	}

	pressure := float64(m.stats.UsedMB) / float64(m.config.MaxMemoryMB) * 100
	if pressure > 100 {
		pressure = 100
	}

	return pressure
}

// SimpleGCOptimizer optimizes garbage collection based on memory usage
type SimpleGCOptimizer struct {
	config        *Config
	lastGCPercent int
	mu            sync.Mutex
}

// NewSimpleGCOptimizer creates a new GC optimizer
func NewSimpleGCOptimizer(config *Config) *SimpleGCOptimizer {
	return &SimpleGCOptimizer{
		config:        config,
		lastGCPercent: config.GCTargetPercent,
	}
}

// Optimize adjusts GC settings based on current memory usage
func (gc *SimpleGCOptimizer) Optimize() {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	const bytesToMB = 1024 * 1024
	currentUsageMB := int64(memStats.Alloc / bytesToMB)

	var newGCPercent int

	// Adjust GC target based on memory pressure
	if currentUsageMB > gc.config.CriticalThresholdMB {
		// Critical memory usage - aggressive GC
		newGCPercent = 50
	} else if currentUsageMB > gc.config.WarningThresholdMB {
		// High memory usage - more frequent GC
		newGCPercent = 75
	} else {
		// Normal memory usage - standard GC
		newGCPercent = gc.config.GCTargetPercent
	}

	// Only update if there's a significant change
	if abs(newGCPercent-gc.lastGCPercent) >= 10 {
		debug.SetGCPercent(newGCPercent)
		gc.lastGCPercent = newGCPercent
	}
}

// GetCurrentGCPercent returns the current GC target percentage
func (gc *SimpleGCOptimizer) GetCurrentGCPercent() int {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	return gc.lastGCPercent
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// MemoryProfiler provides memory profiling capabilities
type MemoryProfiler struct {
	enabled bool
	samples []ProfileSample
	mu      sync.RWMutex
}

// ProfileSample represents a memory profile sample
type ProfileSample struct {
	Timestamp time.Time `json:"timestamp"`
	Stats     Stats     `json:"stats"`
}

// NewMemoryProfiler creates a new memory profiler
func NewMemoryProfiler(enabled bool) *MemoryProfiler {
	return &MemoryProfiler{
		enabled: enabled,
		samples: make([]ProfileSample, 0),
	}
}

// Sample takes a memory profile sample
func (mp *MemoryProfiler) Sample(stats *Stats) {
	if !mp.enabled {
		return
	}

	mp.mu.Lock()
	defer mp.mu.Unlock()

	sample := ProfileSample{
		Timestamp: time.Now(),
		Stats:     *stats,
	}

	mp.samples = append(mp.samples, sample)

	// Keep only last 1000 samples
	if len(mp.samples) > 1000 {
		mp.samples = mp.samples[len(mp.samples)-1000:]
	}
}

// GetSamples returns all profile samples
func (mp *MemoryProfiler) GetSamples() []ProfileSample {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	// Return a copy
	samples := make([]ProfileSample, len(mp.samples))
	copy(samples, mp.samples)

	return samples
}

// GetSamplesInRange returns profile samples within a time range
func (mp *MemoryProfiler) GetSamplesInRange(start, end time.Time) []ProfileSample {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	var result []ProfileSample
	for _, sample := range mp.samples {
		if sample.Timestamp.After(start) && sample.Timestamp.Before(end) {
			result = append(result, sample)
		}
	}

	return result
}

// Clear clears all profile samples
func (mp *MemoryProfiler) Clear() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.samples = mp.samples[:0]
}

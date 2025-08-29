package performance

import (
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// GCOptimizer optimizes garbage collection for performance
type GCOptimizer struct {
	config *OptimizerConfig

	// GC settings
	currentGCPercent int
	memoryLimit      int64
	gcPauseTarget    time.Duration

	// Statistics
	stats *GCStats
	mu    sync.RWMutex
}

// GCStats tracks garbage collection statistics
type GCStats struct {
	// GC frequency
	GCCycles         uint32        `json:"gc_cycles"`
	GCRate           float64       `json:"gc_rate"`          // cycles per second
	LastGCTime       time.Time     `json:"last_gc_time"`

	// GC pause times
	AveragePause     time.Duration `json:"average_pause"`
	MaxPause         time.Duration `json:"max_pause"`
	P95Pause         time.Duration `json:"p95_pause"`
	TotalPauseTime   time.Duration `json:"total_pause_time"`

	// Memory statistics
	HeapSize         uint64        `json:"heap_size"`
	HeapInUse        uint64        `json:"heap_in_use"`
	HeapReleased     uint64        `json:"heap_released"`
	NextGCThreshold  uint64        `json:"next_gc_threshold"`

	// GC efficiency
	GCOverhead       float64       `json:"gc_overhead"`      // GC time / total time
	AllocationRate   float64       `json:"allocation_rate"`  // bytes per second

	LastUpdated      time.Time     `json:"last_updated"`
}

// NewGCOptimizer creates a new garbage collection optimizer
func NewGCOptimizer(config *OptimizerConfig) *GCOptimizer {
	optimizer := &GCOptimizer{
		config:        config,
		gcPauseTarget: config.GCMaxPause,
		memoryLimit:   config.GCMemoryLimit,
		stats:         &GCStats{LastUpdated: time.Now()},
	}

	// Set initial GC parameters
	optimizer.setOptimalGCPercent()
	optimizer.setMemoryLimit()

	return optimizer
}

// Optimize performs garbage collection optimization
func (gco *GCOptimizer) Optimize() {
	gco.mu.Lock()
	defer gco.mu.Unlock()

	gco.updateGCStats()
	gco.adjustGCParameters()
}

// AdjustForLowLatency adjusts GC settings for low latency requirements
func (gco *GCOptimizer) AdjustForLowLatency() {
	gco.mu.Lock()
	defer gco.mu.Unlock()

	// Reduce GC target percentage for more frequent, smaller collections
	newPercent := gco.currentGCPercent - 10
	if newPercent < 20 {
		newPercent = 20 // Don't go below 20%
	}

	if newPercent != gco.currentGCPercent {
		gco.currentGCPercent = newPercent
		debug.SetGCPercent(newPercent)
		
		log.Info().
			Int("new_gc_percent", newPercent).
			Msg("Adjusted GC percent for low latency")
	}
}

// AdjustForThroughput adjusts GC settings for high throughput requirements
func (gco *GCOptimizer) AdjustForThroughput() {
	gco.mu.Lock()
	defer gco.mu.Unlock()

	// Increase GC target percentage for less frequent, larger collections
	newPercent := gco.currentGCPercent + 20
	if newPercent > 200 {
		newPercent = 200 // Cap at 200%
	}

	if newPercent != gco.currentGCPercent {
		gco.currentGCPercent = newPercent
		debug.SetGCPercent(newPercent)
		
		log.Info().
			Int("new_gc_percent", newPercent).
			Msg("Adjusted GC percent for high throughput")
	}
}

// ForceGC triggers an immediate garbage collection
func (gco *GCOptimizer) ForceGC() {
	start := time.Now()
	runtime.GC()
	gcTime := time.Since(start)

	log.Info().
		Dur("gc_duration", gcTime).
		Msg("Forced garbage collection completed")
}

// GetStats returns current GC statistics
func (gco *GCOptimizer) GetStats() *GCStats {
	gco.mu.RLock()
	defer gco.mu.RUnlock()

	// Create a copy to avoid race conditions
	stats := *gco.stats
	return &stats
}

// updateGCStats updates garbage collection statistics
func (gco *GCOptimizer) updateGCStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Calculate GC rate
	gcCycles := m.NumGC
	now := time.Now()
	timeDelta := now.Sub(gco.stats.LastUpdated).Seconds()
	
	if timeDelta > 0 && gco.stats.GCCycles > 0 {
		cycleDelta := gcCycles - gco.stats.GCCycles
		gco.stats.GCRate = float64(cycleDelta) / timeDelta
	}

	// Update GC pause statistics
	gco.updatePauseStats(&m)

	// Update memory statistics
	gco.stats.HeapSize = m.HeapSys
	gco.stats.HeapInUse = m.HeapInuse
	gco.stats.HeapReleased = m.HeapReleased
	gco.stats.NextGCThreshold = m.NextGC

	// Calculate GC overhead
	if m.GCCPUFraction > 0 {
		gco.stats.GCOverhead = m.GCCPUFraction * 100
	}

	// Calculate allocation rate
	if timeDelta > 0 {
		allocDelta := m.TotalAlloc - (gco.stats.HeapInUse + gco.stats.HeapReleased)
		gco.stats.AllocationRate = float64(allocDelta) / timeDelta
	}

	// Update cycle count and timestamp
	gco.stats.GCCycles = gcCycles
	gco.stats.LastGCTime = time.Unix(0, int64(m.LastGC))
	gco.stats.LastUpdated = now
}

// updatePauseStats updates GC pause time statistics
func (gco *GCOptimizer) updatePauseStats(m *runtime.MemStats) {
	if m.NumGC == 0 {
		return
	}

	// Get recent pause times (Go keeps circular buffer of 256 pause times)
	numPauses := int(m.NumGC)
	if numPauses > 256 {
		numPauses = 256
	}

	var totalPause time.Duration
	var maxPause time.Duration
	var pauses []time.Duration

	for i := 0; i < numPauses; i++ {
		pause := time.Duration(m.PauseNs[(m.NumGC-uint32(i)+255)%256])
		pauses = append(pauses, pause)
		totalPause += pause
		
		if pause > maxPause {
			maxPause = pause
		}
	}

	// Calculate statistics
	if numPauses > 0 {
		gco.stats.AveragePause = totalPause / time.Duration(numPauses)
		gco.stats.MaxPause = maxPause
		gco.stats.TotalPauseTime = time.Duration(m.PauseTotalNs)

		// Calculate P95 pause time
		if len(pauses) > 0 {
			sortPauses(pauses)
			p95Index := int(float64(len(pauses)) * 0.95)
			if p95Index >= len(pauses) {
				p95Index = len(pauses) - 1
			}
			gco.stats.P95Pause = pauses[p95Index]
		}
	}
}

// adjustGCParameters adjusts GC parameters based on current performance
func (gco *GCOptimizer) adjustGCParameters() {
	// Don't adjust too frequently
	if time.Since(gco.stats.LastUpdated) < 30*time.Second {
		return
	}

	// Check if pause times are too high
	if gco.stats.AveragePause > gco.gcPauseTarget {
		// Reduce GC percentage to trigger more frequent collections
		newPercent := gco.currentGCPercent - 5
		if newPercent < 20 {
			newPercent = 20
		}
		
		if newPercent != gco.currentGCPercent {
			gco.currentGCPercent = newPercent
			debug.SetGCPercent(newPercent)
			
			log.Info().
				Dur("average_pause", gco.stats.AveragePause).
				Dur("target_pause", gco.gcPauseTarget).
				Int("new_gc_percent", newPercent).
				Msg("Reduced GC percent due to high pause times")
		}
	}

	// Check if GC overhead is too high
	if gco.stats.GCOverhead > 10.0 { // More than 10% time spent in GC
		// Increase GC percentage to reduce collection frequency
		newPercent := gco.currentGCPercent + 10
		if newPercent > 200 {
			newPercent = 200
		}
		
		if newPercent != gco.currentGCPercent {
			gco.currentGCPercent = newPercent
			debug.SetGCPercent(newPercent)
			
			log.Info().
				Float64("gc_overhead", gco.stats.GCOverhead).
				Int("new_gc_percent", newPercent).
				Msg("Increased GC percent due to high overhead")
		}
	}

	// Adjust memory limit if heap is growing too large
	if gco.stats.HeapInUse > uint64(gco.memoryLimit)*8/10 {
		// Trigger more aggressive GC
		gco.ForceGC()
		
		log.Info().
			Uint64("heap_in_use", gco.stats.HeapInUse).
			Int64("memory_limit", gco.memoryLimit).
			Msg("Triggered GC due to high heap usage")
	}
}

// setOptimalGCPercent sets the optimal GC target percentage
func (gco *GCOptimizer) setOptimalGCPercent() {
	gco.currentGCPercent = gco.config.GCTargetPercent
	debug.SetGCPercent(gco.currentGCPercent)
	
	log.Info().
		Int("gc_percent", gco.currentGCPercent).
		Msg("Set GC target percentage")
}

// setMemoryLimit sets the memory limit for garbage collection
func (gco *GCOptimizer) setMemoryLimit() {
	if gco.memoryLimit > 0 {
		debug.SetMemoryLimit(gco.memoryLimit)
		
		log.Info().
			Int64("memory_limit_mb", gco.memoryLimit/(1024*1024)).
			Msg("Set GC memory limit")
	}
}

// GetRecommendations returns GC optimization recommendations
func (gco *GCOptimizer) GetRecommendations() []string {
	gco.mu.RLock()
	defer gco.mu.RUnlock()

	recommendations := make([]string, 0)

	if gco.stats.AveragePause > gco.gcPauseTarget {
		recommendations = append(recommendations,
			"GC pause times are high - consider reducing GC target percentage or optimizing memory allocations")
	}

	if gco.stats.GCOverhead > 5.0 {
		recommendations = append(recommendations,
			"GC overhead is high - consider increasing GC target percentage or reducing allocation rate")
	}

	if gco.stats.GCRate > 2.0 {
		recommendations = append(recommendations,
			"GC frequency is high - consider optimizing memory allocations or increasing heap size")
	}

	if gco.stats.HeapInUse > uint64(gco.memoryLimit)*7/10 {
		recommendations = append(recommendations,
			"Heap usage is approaching memory limit - consider increasing memory limit or optimizing memory usage")
	}

	return recommendations
}

// EnableGCTrace enables garbage collection tracing for debugging
func (gco *GCOptimizer) EnableGCTrace() {
	debug.SetGCPercent(gco.currentGCPercent)
	log.Info().Msg("GC tracing enabled")
}

// DisableGCTrace disables garbage collection tracing
func (gco *GCOptimizer) DisableGCTrace() {
	log.Info().Msg("GC tracing disabled")
}

// OptimizeForWorkload optimizes GC settings for specific workload characteristics
func (gco *GCOptimizer) OptimizeForWorkload(workloadType string) {
	gco.mu.Lock()
	defer gco.mu.Unlock()

	switch workloadType {
	case "low-latency":
		// Optimize for minimal pause times
		gco.currentGCPercent = 30
		debug.SetGCPercent(30)
		log.Info().Msg("Optimized GC for low-latency workload")

	case "high-throughput":
		// Optimize for maximum throughput
		gco.currentGCPercent = 100
		debug.SetGCPercent(100)
		log.Info().Msg("Optimized GC for high-throughput workload")

	case "memory-constrained":
		// Optimize for minimal memory usage
		gco.currentGCPercent = 50
		debug.SetGCPercent(50)
		log.Info().Msg("Optimized GC for memory-constrained workload")

	case "balanced":
		// Balanced optimization
		gco.currentGCPercent = gco.config.GCTargetPercent
		debug.SetGCPercent(gco.currentGCPercent)
		log.Info().Msg("Optimized GC for balanced workload")

	default:
		log.Warn().
			Str("workload_type", workloadType).
			Msg("Unknown workload type, using default settings")
	}
}

// sortPauses sorts pause times in ascending order (simple bubble sort for small arrays)
func sortPauses(pauses []time.Duration) {
	n := len(pauses)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if pauses[j] > pauses[j+1] {
				pauses[j], pauses[j+1] = pauses[j+1], pauses[j]
			}
		}
	}
}
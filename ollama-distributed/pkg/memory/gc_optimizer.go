package memory

import (
	"runtime"
	"runtime/debug"
	"sync/atomic"
	"time"
)

// GCOptimizer provides intelligent garbage collection optimization
type GCOptimizer struct {
	config             *OptimizedConfig
	lastOptimization   time.Time
	optimizationCount  int64
	effectivenessScore float64
	
	// GC statistics tracking
	lastGCStats debug.GCStats
	
	// Adaptive parameters
	targetPercent     int
	adaptiveInterval  time.Duration
	
	// Performance tracking
	beforeGCPause time.Duration
	afterGCPause  time.Duration
}

// NewGCOptimizer creates a new GC optimizer
func NewGCOptimizer(config *OptimizedConfig) *GCOptimizer {
	optimizer := &GCOptimizer{
		config:           config,
		targetPercent:    config.GCTargetPercent,
		adaptiveInterval: 30 * time.Second,
		lastOptimization: time.Now(),
	}
	
	// Initialize with current GC stats
	debug.ReadGCStats(&optimizer.lastGCStats)
	
	return optimizer
}

// OptimizeGC performs intelligent GC optimization based on current conditions
func (gc *GCOptimizer) OptimizeGC() {
	now := time.Now()
	
	// Avoid too frequent optimizations
	if now.Sub(gc.lastOptimization) < gc.adaptiveInterval {
		return
	}
	
	// Collect current GC statistics
	var currentStats debug.GCStats
	debug.ReadGCStats(&currentStats)
	
	// Analyze GC performance
	gc.analyzeGCPerformance(&currentStats)
	
	// Apply adaptive GC tuning
	gc.applyAdaptiveTuning(&currentStats)
	
	// Update tracking
	gc.lastGCStats = currentStats
	gc.lastOptimization = now
	atomic.AddInt64(&gc.optimizationCount, 1)
}

// analyzeGCPerformance analyzes current GC performance patterns
func (gc *GCOptimizer) analyzeGCPerformance(stats *debug.GCStats) {
	// Calculate average pause time
	if len(stats.Pause) > 0 {
		var totalPause time.Duration
		for _, pause := range stats.Pause {
			totalPause += pause
		}
		avgPause := totalPause / time.Duration(len(stats.Pause))
		
		// Store before optimization pause time
		gc.beforeGCPause = avgPause
		
		// Calculate effectiveness of previous optimization
		if gc.afterGCPause > 0 {
			improvement := float64(gc.beforeGCPause-gc.afterGCPause) / float64(gc.beforeGCPause)
			gc.effectivenessScore = improvement
		}
	}
}

// applyAdaptiveTuning applies intelligent GC parameter tuning
func (gc *GCOptimizer) applyAdaptiveTuning(stats *debug.GCStats) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	// Calculate memory pressure
	memoryPressure := float64(memStats.Alloc) / float64(memStats.Sys)
	
	// Adaptive GC target percentage based on memory pressure
	newTargetPercent := gc.calculateAdaptiveTarget(memoryPressure, stats)
	
	// Apply new GC target if changed
	if newTargetPercent != gc.targetPercent {
		debug.SetGCPercent(newTargetPercent)
		gc.targetPercent = newTargetPercent
	}
	
	// Apply memory limit if memory pressure is high
	if memoryPressure > 0.8 {
		gc.applyMemoryLimit(&memStats)
	}
	
	// Trigger manual GC if conditions warrant it
	if gc.shouldTriggerManualGC(memoryPressure, &memStats) {
		runtime.GC()
	}
}

// calculateAdaptiveTarget calculates optimal GC target percentage
func (gc *GCOptimizer) calculateAdaptiveTarget(memoryPressure float64, stats *debug.GCStats) int {
	baseTarget := gc.config.GCTargetPercent
	
	switch {
	case memoryPressure > 0.9:
		// High memory pressure: aggressive GC
		return maxInt(baseTarget/2, 25)
		
	case memoryPressure > 0.7:
		// Medium pressure: moderate GC
		return maxInt(int(float64(baseTarget)*0.75), 50)
		
	case memoryPressure < 0.3:
		// Low pressure: relaxed GC
		return minInt(baseTarget*2, 200)
		
	default:
		// Normal pressure: use base target
		return baseTarget
	}
}

// applyMemoryLimit applies soft memory limit to encourage GC
func (gc *GCOptimizer) applyMemoryLimit(memStats *runtime.MemStats) {
	// Set soft memory limit to 90% of current heap size
	// This encourages GC before hitting hard limits
	softLimit := int64(float64(memStats.HeapSys) * 0.9)
	
	// Go 1.19+ feature - set memory limit
	if softLimit > 0 {
		debug.SetMemoryLimit(softLimit)
	}
}

// shouldTriggerManualGC determines if manual GC should be triggered
func (gc *GCOptimizer) shouldTriggerManualGC(memoryPressure float64, memStats *runtime.MemStats) bool {
	// Trigger manual GC under these conditions:
	
	// 1. Very high memory pressure
	if memoryPressure > 0.95 {
		return true
	}
	
	// 2. Large amount of heap allocated but low GC frequency
	if memStats.Alloc > 100*1024*1024 && // >100MB allocated
		time.Since(time.Unix(0, int64(memStats.LastGC))) > 5*time.Minute {
		return true
	}
	
	// 3. High number of objects but low GC activity
	if memStats.Mallocs-memStats.Frees > 1000000 && // >1M live objects
		memStats.NumGC < 10 {
		return true
	}
	
	return false
}

// GetOptimizationStats returns current optimization statistics
func (gc *GCOptimizer) GetOptimizationStats() *GCOptimizationStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	return &GCOptimizationStats{
		OptimizationCount:  atomic.LoadInt64(&gc.optimizationCount),
		EffectivenessScore: gc.effectivenessScore,
		CurrentTargetPercent: gc.targetPercent,
		LastOptimization:   gc.lastOptimization,
		
		// Current GC metrics
		NumGC:          memStats.NumGC,
		TotalPauseNs:   memStats.PauseTotalNs,
		LastGC:         time.Unix(0, int64(memStats.LastGC)),
		HeapAlloc:      memStats.Alloc,
		HeapSys:        memStats.HeapSys,
		HeapInuse:      memStats.HeapInuse,
		HeapReleased:   memStats.HeapReleased,
		
		// Object statistics
		LiveObjects:    memStats.Mallocs - memStats.Frees,
		TotalAllocs:    memStats.Mallocs,
		TotalFrees:     memStats.Frees,
	}
}

// GCOptimizationStats holds GC optimization statistics
type GCOptimizationStats struct {
	// Optimization metrics
	OptimizationCount    int64     `json:"optimization_count"`
	EffectivenessScore   float64   `json:"effectiveness_score"`
	CurrentTargetPercent int       `json:"current_target_percent"`
	LastOptimization     time.Time `json:"last_optimization"`
	
	// GC performance metrics
	NumGC          uint32        `json:"num_gc"`
	TotalPauseNs   uint64        `json:"total_pause_ns"`
	LastGC         time.Time     `json:"last_gc"`
	HeapAlloc      uint64        `json:"heap_alloc"`
	HeapSys        uint64        `json:"heap_sys"`
	HeapInuse      uint64        `json:"heap_inuse"`
	HeapReleased   uint64        `json:"heap_released"`
	
	// Object metrics
	LiveObjects  uint64 `json:"live_objects"`
	TotalAllocs  uint64 `json:"total_allocs"`
	TotalFrees   uint64 `json:"total_frees"`
}

// AutoTune starts automatic GC tuning in the background
func (gc *GCOptimizer) AutoTune() {
	go func() {
		ticker := time.NewTicker(gc.adaptiveInterval)
		defer ticker.Stop()
		
		for range ticker.C {
			gc.OptimizeGC()
		}
	}()
}

// ResetOptimization resets GC optimization to default settings
func (gc *GCOptimizer) ResetOptimization() {
	debug.SetGCPercent(gc.config.GCTargetPercent)
	debug.SetMemoryLimit(-1) // Remove memory limit
	gc.targetPercent = gc.config.GCTargetPercent
	gc.effectivenessScore = 0
	gc.lastOptimization = time.Now()
}

// ForceGC triggers immediate garbage collection with optimization
func (gc *GCOptimizer) ForceGC() {
	// Collect before stats
	beforeStats := gc.getGCStats()
	
	// Force garbage collection
	runtime.GC()
	
	// Collect after stats and update effectiveness
	afterStats := gc.getGCStats()
	
	// Calculate and store pause time improvement
	if beforeStats.avgPause > 0 && afterStats.avgPause > 0 {
		gc.beforeGCPause = beforeStats.avgPause
		gc.afterGCPause = afterStats.avgPause
		
		improvement := float64(gc.beforeGCPause-gc.afterGCPause) / float64(gc.beforeGCPause)
		gc.effectivenessScore = improvement
	}
}

// gcStats holds simple GC statistics
type gcStats struct {
	avgPause time.Duration
	numGC    uint32
}

// getGCStats gets current GC statistics
func (gc *GCOptimizer) getGCStats() gcStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	stats := gcStats{
		numGC: memStats.NumGC,
	}
	
	// Calculate average pause time from recent pauses
	if memStats.NumGC > 0 && memStats.PauseTotalNs > 0 {
		stats.avgPause = time.Duration(memStats.PauseTotalNs / uint64(memStats.NumGC))
	}
	
	return stats
}

// Helper functions
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
package memory

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

// BenchmarkMemoryPools tests the performance of optimized memory pools
func BenchmarkMemoryPools(b *testing.B) {
	config := DefaultOptimizedConfig()
	manager := NewOptimizedMemoryManager(config)

	b.Run("ChannelPool", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				ch := manager.GetChannel(100)
				manager.ReturnChannel(ch, 100)
			}
		})
	})

	b.Run("BufferPool", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				buf := manager.GetBuffer(1024)
				manager.ReturnBuffer(buf)
			}
		})
	})

	b.Run("ObjectPool", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				obj := manager.GetRequest()
				manager.ReturnRequest(obj)
			}
		})
	})
}

// BenchmarkMemoryAllocations compares pooled vs non-pooled allocations
func BenchmarkMemoryAllocations(b *testing.B) {
	config := DefaultOptimizedConfig()
	manager := NewOptimizedMemoryManager(config)

	b.Run("PooledChannels", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ch := manager.GetChannel(100)
			manager.ReturnChannel(ch, 100)
		}
	})

	b.Run("DirectChannels", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ch := make(chan interface{}, 100)
			// Simulate usage
			_ = ch
		}
	})

	b.Run("PooledBuffers", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := manager.GetBuffer(1024)
			manager.ReturnBuffer(buf)
		}
	})

	b.Run("DirectBuffers", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := make([]byte, 1024)
			// Simulate usage
			_ = buf
		}
	})
}

// BenchmarkGCOptimizer tests GC optimization performance
func BenchmarkGCOptimizer(b *testing.B) {
	config := DefaultOptimizedConfig()
	optimizer := NewGCOptimizer(config)

	b.Run("GCOptimization", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			optimizer.OptimizeGC()
		}
	})

	b.Run("AdaptiveTuning", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulate memory pressure changes
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			optimizer.applyAdaptiveTuning(nil)
		}
	})
}

// BenchmarkConcurrentMemoryOps tests concurrent memory operations
func BenchmarkConcurrentMemoryOps(b *testing.B) {
	config := DefaultOptimizedConfig()
	manager := NewOptimizedMemoryManager(config)

	b.Run("ConcurrentChannelOps", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				ch := manager.GetChannel(50)
				// Simulate usage
				go func() {
					time.Sleep(time.Microsecond)
					manager.ReturnChannel(ch, 50)
				}()
			}
		})
		time.Sleep(100 * time.Millisecond) // Wait for goroutines
	})

	b.Run("ConcurrentBufferOps", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				buf := manager.GetBuffer(512)
				// Simulate usage
				for i := range buf {
					buf[i] = byte(i % 256)
				}
				manager.ReturnBuffer(buf)
			}
		})
	})
}

// BenchmarkMemoryPressure tests behavior under memory pressure
func BenchmarkMemoryPressure(b *testing.B) {
	config := DefaultOptimizedConfig()
	config.GCPressureThreshold = 50 // Lower threshold for testing
	manager := NewOptimizedMemoryManager(config)

	b.Run("HighMemoryPressure", func(b *testing.B) {
		// Allocate large amounts of memory to create pressure
		var allocations [][]byte
		for i := 0; i < 1000; i++ {
			allocations = append(allocations, make([]byte, 1024*1024)) // 1MB each
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			manager.UpdateMemoryPressure()
			buf := manager.GetBuffer(1024)
			manager.ReturnBuffer(buf)
		}

		// Keep allocations alive
		_ = allocations
	})
}

// BenchmarkPoolReuse tests pool reuse efficiency
func BenchmarkPoolReuse(b *testing.B) {
	config := DefaultOptimizedConfig()
	manager := NewOptimizedMemoryManager(config)

	// Pre-warm the pools
	channels := make([]chan interface{}, 100)
	buffers := make([][]byte, 100)

	for i := 0; i < 100; i++ {
		channels[i] = manager.GetChannel(100)
		buffers[i] = manager.GetBuffer(1024)
	}

	for i := 0; i < 100; i++ {
		manager.ReturnChannel(channels[i], 100)
		manager.ReturnBuffer(buffers[i])
	}

	b.Run("WarmPoolAccess", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				ch := manager.GetChannel(100)
				buf := manager.GetBuffer(1024)
				manager.ReturnChannel(ch, 100)
				manager.ReturnBuffer(buf)
			}
		})
	})
}

// TestMemoryPoolCorrectness verifies pool correctness
func TestMemoryPoolCorrectness(t *testing.T) {
	config := DefaultOptimizedConfig()
	manager := NewOptimizedMemoryManager(config)

	t.Run("ChannelPoolCorrectness", func(t *testing.T) {
		const numOps = 1000
		const capacity = 50

		var wg sync.WaitGroup
		wg.Add(numOps)

		for i := 0; i < numOps; i++ {
			go func() {
				defer wg.Done()
				ch := manager.GetChannel(capacity)
				if cap(ch) != capacity {
					t.Errorf("Expected channel capacity %d, got %d", capacity, cap(ch))
				}
				manager.ReturnChannel(ch, capacity)
			}()
		}

		wg.Wait()

		metrics := manager.GetMetrics()
		if metrics.ChannelPoolHits+metrics.ChannelPoolMiss != numOps {
			t.Errorf("Expected %d total channel operations, got %d", 
				numOps, metrics.ChannelPoolHits+metrics.ChannelPoolMiss)
		}
	})

	t.Run("BufferPoolCorrectness", func(t *testing.T) {
		const numOps = 1000
		const size = 1024

		var wg sync.WaitGroup
		wg.Add(numOps)

		for i := 0; i < numOps; i++ {
			go func() {
				defer wg.Done()
				buf := manager.GetBuffer(size)
				if len(buf) != size {
					t.Errorf("Expected buffer size %d, got %d", size, len(buf))
				}
				// Test that buffer is zero-initialized or cleared
				for j, b := range buf {
					if b != 0 {
						t.Errorf("Buffer not cleared at position %d: got %d", j, b)
						break
					}
				}
				manager.ReturnBuffer(buf)
			}()
		}

		wg.Wait()

		metrics := manager.GetMetrics()
		if metrics.BufferPoolHits+metrics.BufferPoolMiss != numOps {
			t.Errorf("Expected %d total buffer operations, got %d", 
				numOps, metrics.BufferPoolHits+metrics.BufferPoolMiss)
		}
	})
}

// TestGCOptimizerEffectiveness tests GC optimizer effectiveness
func TestGCOptimizerEffectiveness(t *testing.T) {
	config := DefaultOptimizedConfig()
	optimizer := NewGCOptimizer(config)

	// Record initial GC stats
	var beforeStats runtime.MemStats
	runtime.ReadMemStats(&beforeStats)

	// Allocate memory to trigger GC
	allocations := make([][]byte, 1000)
	for i := range allocations {
		allocations[i] = make([]byte, 1024*1024) // 1MB each
	}

	// Apply optimization
	optimizer.OptimizeGC()

	// Force GC and measure
	optimizer.ForceGC()

	stats := optimizer.GetOptimizationStats()
	if stats.OptimizationCount == 0 {
		t.Error("Expected at least one optimization")
	}

	if stats.NumGC == 0 {
		t.Error("Expected some GC cycles")
	}

	// Verify memory is being tracked
	if stats.HeapAlloc == 0 {
		t.Error("Expected some heap allocation")
	}

	// Keep allocations alive
	_ = allocations
}

// TestMemoryPoolMetrics tests metrics collection
func TestMemoryPoolMetrics(t *testing.T) {
	config := DefaultOptimizedConfig()
	config.EnableMetrics = true
	manager := NewOptimizedMemoryManager(config)

	// Perform operations to generate metrics
	for i := 0; i < 100; i++ {
		ch := manager.GetChannel(10)
		buf := manager.GetBuffer(512)
		obj := manager.GetRequest()

		manager.ReturnChannel(ch, 10)
		manager.ReturnBuffer(buf)
		manager.ReturnRequest(obj)
	}

	metrics := manager.GetMetrics()
	
	if metrics.ChannelPoolHits == 0 && metrics.ChannelPoolMiss == 0 {
		t.Error("Expected channel pool metrics")
	}

	if metrics.BufferPoolHits == 0 && metrics.BufferPoolMiss == 0 {
		t.Error("Expected buffer pool metrics")
	}

	if metrics.ChannelReuseRate < 0 || metrics.ChannelReuseRate > 1 {
		t.Errorf("Invalid channel reuse rate: %f", metrics.ChannelReuseRate)
	}

	if metrics.BufferReuseRate < 0 || metrics.BufferReuseRate > 1 {
		t.Errorf("Invalid buffer reuse rate: %f", metrics.BufferReuseRate)
	}

	if metrics.LastUpdated.IsZero() {
		t.Error("Expected metrics timestamp")
	}
}
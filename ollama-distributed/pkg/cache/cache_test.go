package cache

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
)

func TestNewCache(t *testing.T) {
	cfg := &config.Config{
		Cache: &config.CacheConfig{
			Type:        "memory",
			MaxSize:     1000,
			TTL:         5 * time.Minute,
			CleanupInterval: time.Minute,
		},
	}

	cache, err := NewCache(cfg)
	require.NoError(t, err)
	require.NotNil(t, cache)

	assert.Equal(t, cfg.Cache.Type, cache.config.Type)
	assert.Equal(t, cfg.Cache.MaxSize, cache.config.MaxSize)
	assert.Equal(t, cfg.Cache.TTL, cache.config.TTL)
}

func TestCache_SetGet(t *testing.T) {
	cfg := &config.Config{
		Cache: &config.CacheConfig{
			Type:        "memory",
			MaxSize:     100,
			TTL:         time.Hour,
			CleanupInterval: 10 * time.Minute,
		},
	}

	cache, err := NewCache(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Test setting and getting values
	testCases := []struct {
		key   string
		value interface{}
	}{
		{"string-key", "string-value"},
		{"int-key", 42},
		{"bool-key", true},
		{"float-key", 3.14159},
		{"struct-key", map[string]interface{}{"name": "test", "count": 5}},
		{"slice-key", []string{"a", "b", "c"}},
	}

	for _, tc := range testCases {
		t.Run("Key: "+tc.key, func(t *testing.T) {
			// Set value
			err := cache.Set(ctx, tc.key, tc.value)
			require.NoError(t, err)

			// Get value
			retrieved, err := cache.Get(ctx, tc.key)
			require.NoError(t, err)
			assert.Equal(t, tc.value, retrieved)

			// Check existence
			exists, err := cache.Exists(ctx, tc.key)
			require.NoError(t, err)
			assert.True(t, exists)
		})
	}
}

func TestCache_GetNonExistent(t *testing.T) {
	cfg := &config.Config{
		Cache: &config.CacheConfig{
			Type:        "memory",
			MaxSize:     100,
			TTL:         time.Hour,
			CleanupInterval: 10 * time.Minute,
		},
	}

	cache, err := NewCache(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Get non-existent key
	_, err = cache.Get(ctx, "non-existent-key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Check non-existent key
	exists, err := cache.Exists(ctx, "non-existent-key")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestCache_Delete(t *testing.T) {
	cfg := &config.Config{
		Cache: &config.CacheConfig{
			Type:        "memory",
			MaxSize:     100,
			TTL:         time.Hour,
			CleanupInterval: 10 * time.Minute,
		},
	}

	cache, err := NewCache(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Set value
	err = cache.Set(ctx, "delete-test", "value")
	require.NoError(t, err)

	// Verify it exists
	exists, err := cache.Exists(ctx, "delete-test")
	require.NoError(t, err)
	assert.True(t, exists)

	// Delete value
	err = cache.Delete(ctx, "delete-test")
	require.NoError(t, err)

	// Verify it's gone
	exists, err = cache.Exists(ctx, "delete-test")
	require.NoError(t, err)
	assert.False(t, exists)

	// Delete non-existent key should not error
	err = cache.Delete(ctx, "non-existent")
	assert.NoError(t, err)
}

func TestCache_TTL_Expiration(t *testing.T) {
	cfg := &config.Config{
		Cache: &config.CacheConfig{
			Type:        "memory",
			MaxSize:     100,
			TTL:         100 * time.Millisecond, // Very short TTL
			CleanupInterval: 50 * time.Millisecond,
		},
	}

	cache, err := NewCache(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Start cache cleanup
	go cache.StartCleanup(ctx)

	// Set value
	err = cache.Set(ctx, "expiry-test", "value")
	require.NoError(t, err)

	// Verify it exists
	value, err := cache.Get(ctx, "expiry-test")
	require.NoError(t, err)
	assert.Equal(t, "value", value)

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Should be expired
	_, err = cache.Get(ctx, "expiry-test")
	assert.Error(t, err)

	exists, err := cache.Exists(ctx, "expiry-test")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestCache_MaxSize_Eviction(t *testing.T) {
	cfg := &config.Config{
		Cache: &config.CacheConfig{
			Type:        "memory",
			MaxSize:     3, // Very small cache
			TTL:         time.Hour,
			CleanupInterval: 10 * time.Minute,
			EvictionPolicy: "lru", // Least Recently Used
		},
	}

	cache, err := NewCache(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Fill cache to capacity
	err = cache.Set(ctx, "key1", "value1")
	require.NoError(t, err)
	err = cache.Set(ctx, "key2", "value2")
	require.NoError(t, err)
	err = cache.Set(ctx, "key3", "value3")
	require.NoError(t, err)

	// All should exist
	for i := 1; i <= 3; i++ {
		exists, err := cache.Exists(ctx, fmt.Sprintf("key%d", i))
		require.NoError(t, err)
		assert.True(t, exists, "Key%d should exist", i)
	}

	// Access key1 to make it recently used
	_, err = cache.Get(ctx, "key1")
	require.NoError(t, err)

	// Add a new key, should evict key2 (least recently used)
	err = cache.Set(ctx, "key4", "value4")
	require.NoError(t, err)

	// key2 should be evicted
	exists, err := cache.Exists(ctx, "key2")
	require.NoError(t, err)
	assert.False(t, exists, "key2 should be evicted")

	// Others should still exist
	for _, key := range []string{"key1", "key3", "key4"} {
		exists, err := cache.Exists(ctx, key)
		require.NoError(t, err)
		assert.True(t, exists, "%s should still exist", key)
	}
}

func TestCache_Concurrent_Access(t *testing.T) {
	cfg := &config.Config{
		Cache: &config.CacheConfig{
			Type:        "memory",
			MaxSize:     1000,
			TTL:         time.Hour,
			CleanupInterval: 10 * time.Minute,
		},
	}

	cache, err := NewCache(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	numGoroutines := 10
	numOperations := 100

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numOperations)

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key-%d-%d", goroutineID, j)
				value := fmt.Sprintf("value-%d-%d", goroutineID, j)
				if err := cache.Set(ctx, key, value); err != nil {
					errors <- err
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Error("Concurrent operation error:", err)
	}

	// Verify some values exist
	exists, err := cache.Exists(ctx, "key-0-0")
	require.NoError(t, err)
	assert.True(t, exists)

	value, err := cache.Get(ctx, "key-0-0")
	require.NoError(t, err)
	assert.Equal(t, "value-0-0", value)
}

func TestCache_Concurrent_ReadWrite(t *testing.T) {
	cfg := &config.Config{
		Cache: &config.CacheConfig{
			Type:        "memory",
			MaxSize:     100,
			TTL:         time.Hour,
			CleanupInterval: 10 * time.Minute,
		},
	}

	cache, err := NewCache(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Pre-populate cache
	for i := 0; i < 10; i++ {
		err = cache.Set(ctx, fmt.Sprintf("key-%d", i), fmt.Sprintf("value-%d", i))
		require.NoError(t, err)
	}

	var wg sync.WaitGroup
	numReaders := 5
	numWriters := 3
	duration := 1 * time.Second

	// Start readers
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()
			start := time.Now()
			reads := 0
			for time.Since(start) < duration {
				key := fmt.Sprintf("key-%d", reads%10)
				_, err := cache.Get(ctx, key)
				if err != nil && !strings.Contains(err.Error(), "not found") {
					t.Errorf("Reader %d error: %v", readerID, err)
				}
				reads++
			}
			t.Logf("Reader %d completed %d reads", readerID, reads)
		}(i)
	}

	// Start writers
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			start := time.Now()
			writes := 0
			for time.Since(start) < duration {
				key := fmt.Sprintf("writer-key-%d-%d", writerID, writes)
				value := fmt.Sprintf("writer-value-%d-%d", writerID, writes)
				if err := cache.Set(ctx, key, value); err != nil {
					t.Errorf("Writer %d error: %v", writerID, err)
				}
				writes++
			}
			t.Logf("Writer %d completed %d writes", writerID, writes)
		}(i)
	}

	wg.Wait()
}

func TestCache_SetWithTTL(t *testing.T) {
	cfg := &config.Config{
		Cache: &config.CacheConfig{
			Type:        "memory",
			MaxSize:     100,
			TTL:         time.Hour, // Default TTL
			CleanupInterval: 50 * time.Millisecond,
		},
	}

	cache, err := NewCache(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Start cleanup
	go cache.StartCleanup(ctx)

	// Set value with custom TTL
	customTTL := 100 * time.Millisecond
	err = cache.SetWithTTL(ctx, "custom-ttl-key", "custom-value", customTTL)
	require.NoError(t, err)

	// Should exist immediately
	value, err := cache.Get(ctx, "custom-ttl-key")
	require.NoError(t, err)
	assert.Equal(t, "custom-value", value)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, err = cache.Get(ctx, "custom-ttl-key")
	assert.Error(t, err)
}

func TestCache_GetStats(t *testing.T) {
	cfg := &config.Config{
		Cache: &config.CacheConfig{
			Type:        "memory",
			MaxSize:     100,
			TTL:         time.Hour,
			CleanupInterval: 10 * time.Minute,
		},
	}

	cache, err := NewCache(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Perform some operations
	err = cache.Set(ctx, "stats-key1", "value1")
	require.NoError(t, err)
	err = cache.Set(ctx, "stats-key2", "value2")
	require.NoError(t, err)

	// Get some values (hits)
	_, err = cache.Get(ctx, "stats-key1")
	require.NoError(t, err)
	_, err = cache.Get(ctx, "stats-key1")
	require.NoError(t, err)

	// Try to get non-existent key (miss)
	_, err = cache.Get(ctx, "non-existent")
	assert.Error(t, err)

	// Get stats
	stats := cache.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, 2, stats.KeyCount)
	assert.GreaterOrEqual(t, stats.Hits, 2)
	assert.GreaterOrEqual(t, stats.Misses, 1)
	assert.Greater(t, stats.HitRate, 0.0)
}

func TestCache_Clear(t *testing.T) {
	cfg := &config.Config{
		Cache: &config.CacheConfig{
			Type:        "memory",
			MaxSize:     100,
			TTL:         time.Hour,
			CleanupInterval: 10 * time.Minute,
		},
	}

	cache, err := NewCache(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Add some values
	for i := 0; i < 10; i++ {
		err = cache.Set(ctx, fmt.Sprintf("clear-key-%d", i), fmt.Sprintf("value-%d", i))
		require.NoError(t, err)
	}

	// Verify values exist
	stats := cache.GetStats()
	assert.Equal(t, 10, stats.KeyCount)

	// Clear cache
	err = cache.Clear(ctx)
	require.NoError(t, err)

	// Verify cache is empty
	stats = cache.GetStats()
	assert.Equal(t, 0, stats.KeyCount)

	// Verify keys don't exist
	for i := 0; i < 10; i++ {
		exists, err := cache.Exists(ctx, fmt.Sprintf("clear-key-%d", i))
		require.NoError(t, err)
		assert.False(t, exists)
	}
}

func TestCache_Batch_Operations(t *testing.T) {
	cfg := &config.Config{
		Cache: &config.CacheConfig{
			Type:        "memory",
			MaxSize:     100,
			TTL:         time.Hour,
			CleanupInterval: 10 * time.Minute,
		},
	}

	cache, err := NewCache(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Batch set
	batchData := map[string]interface{}{
		"batch-key1": "batch-value1",
		"batch-key2": 42,
		"batch-key3": true,
		"batch-key4": []string{"a", "b", "c"},
	}

	err = cache.SetBatch(ctx, batchData)
	require.NoError(t, err)

	// Batch get
	keys := []string{"batch-key1", "batch-key2", "batch-key3", "batch-key4", "non-existent"}
	results, err := cache.GetBatch(ctx, keys)
	require.NoError(t, err)

	assert.Len(t, results, 5)
	assert.Equal(t, "batch-value1", results["batch-key1"])
	assert.Equal(t, 42, results["batch-key2"])
	assert.Equal(t, true, results["batch-key3"])
	assert.Equal(t, []string{"a", "b", "c"}, results["batch-key4"])
	assert.Nil(t, results["non-existent"]) // Should be nil for non-existent key

	// Batch delete
	deleteKeys := []string{"batch-key1", "batch-key2"}
	err = cache.DeleteBatch(ctx, deleteKeys)
	require.NoError(t, err)

	// Verify deletion
	exists, err := cache.Exists(ctx, "batch-key1")
	require.NoError(t, err)
	assert.False(t, exists)

	exists, err = cache.Exists(ctx, "batch-key2")
	require.NoError(t, err)
	assert.False(t, exists)

	// Others should still exist
	exists, err = cache.Exists(ctx, "batch-key3")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestCache_Pattern_Operations(t *testing.T) {
	cfg := &config.Config{
		Cache: &config.CacheConfig{
			Type:        "memory",
			MaxSize:     100,
			TTL:         time.Hour,
			CleanupInterval: 10 * time.Minute,
		},
	}

	cache, err := NewCache(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Add keys with patterns
	testKeys := []string{
		"user:123:profile",
		"user:123:settings",
		"user:456:profile",
		"user:456:settings",
		"session:abc:data",
		"session:def:data",
	}

	for _, key := range testKeys {
		err = cache.Set(ctx, key, "test-value")
		require.NoError(t, err)
	}

	// Find keys by pattern
	userKeys, err := cache.GetKeysByPattern(ctx, "user:*")
	require.NoError(t, err)
	assert.Len(t, userKeys, 4)

	user123Keys, err := cache.GetKeysByPattern(ctx, "user:123:*")
	require.NoError(t, err)
	assert.Len(t, user123Keys, 2)

	sessionKeys, err := cache.GetKeysByPattern(ctx, "session:*:data")
	require.NoError(t, err)
	assert.Len(t, sessionKeys, 2)

	// Delete by pattern
	err = cache.DeleteByPattern(ctx, "user:123:*")
	require.NoError(t, err)

	// Verify deletion
	remainingUserKeys, err := cache.GetKeysByPattern(ctx, "user:*")
	require.NoError(t, err)
	assert.Len(t, remainingUserKeys, 2) // Only user:456 keys should remain
}

func TestCache_Shutdown(t *testing.T) {
	cfg := &config.Config{
		Cache: &config.CacheConfig{
			Type:        "memory",
			MaxSize:     100,
			TTL:         time.Hour,
			CleanupInterval: time.Second,
		},
	}

	cache, err := NewCache(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	// Start cleanup
	go cache.StartCleanup(ctx)

	// Add some data
	err = cache.Set(ctx, "shutdown-test", "value")
	require.NoError(t, err)

	// Shutdown
	cancel()
	err = cache.Shutdown()
	assert.NoError(t, err)

	// Operations after shutdown should fail gracefully
	err = cache.Set(context.Background(), "after-shutdown", "value")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shutdown")
}
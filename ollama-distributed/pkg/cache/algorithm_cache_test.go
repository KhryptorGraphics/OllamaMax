package cache

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAlgorithmCache(t *testing.T) {
	tests := []struct {
		name   string
		config *CacheConfig
		hasErr bool
	}{
		{
			name:   "default config",
			config: nil,
			hasErr: false,
		},
		{
			name: "custom config",
			config: &CacheConfig{
				MaxMemoryEntries: 100,
				MemoryTTL:        5 * time.Minute,
				EvictionPolicy:   "LRU",
				MaxKeySize:       256,
				MaxValueSize:     1024,
				EnableMetrics:    true,
				CleanupInterval:  30 * time.Second,
			},
			hasErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, err := NewAlgorithmCache(tt.config)
			if tt.hasErr {
				assert.Error(t, err)
				assert.Nil(t, cache)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cache)
				defer cache.Close()
			}
		})
	}
}

func TestAlgorithmCache_SetAndGet(t *testing.T) {
	cache, err := NewAlgorithmCache(nil)
	require.NoError(t, err)
	defer cache.Close()

	// Test basic set and get
	key := "test-key"
	value := "test-value"
	ttl := 5 * time.Minute

	err = cache.Set(key, value, ttl)
	require.NoError(t, err)

	retrievedValue, found, err := cache.Get(key)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, value, retrievedValue)
}

func TestAlgorithmCache_GetNonExistent(t *testing.T) {
	cache, err := NewAlgorithmCache(nil)
	require.NoError(t, err)
	defer cache.Close()

	value, found, err := cache.Get("non-existent-key")
	require.NoError(t, err)
	assert.False(t, found)
	assert.Nil(t, value)
}

func TestAlgorithmCache_Delete(t *testing.T) {
	cache, err := NewAlgorithmCache(nil)
	require.NoError(t, err)
	defer cache.Close()

	key := "test-key"
	value := "test-value"

	// Set a value
	err = cache.Set(key, value, 5*time.Minute)
	require.NoError(t, err)

	// Verify it exists
	_, found, err := cache.Get(key)
	require.NoError(t, err)
	assert.True(t, found)

	// Delete it
	err = cache.Delete(key)
	require.NoError(t, err)

	// Verify it's gone
	_, found, err = cache.Get(key)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestAlgorithmCache_Clear(t *testing.T) {
	cache, err := NewAlgorithmCache(nil)
	require.NoError(t, err)
	defer cache.Close()

	// Set multiple values
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("key-%d", i)
		value := fmt.Sprintf("value-%d", i)
		err = cache.Set(key, value, 5*time.Minute)
		require.NoError(t, err)
	}

	// Clear the cache
	err = cache.Clear()
	require.NoError(t, err)

	// Verify all values are gone
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("key-%d", i)
		_, found, err := cache.Get(key)
		require.NoError(t, err)
		assert.False(t, found)
	}
}

func TestAlgorithmCache_TTLExpiration(t *testing.T) {
	cache, err := NewAlgorithmCache(nil)
	require.NoError(t, err)
	defer cache.Close()

	key := "test-key"
	value := "test-value"
	ttl := 100 * time.Millisecond

	// Set a value with short TTL
	err = cache.Set(key, value, ttl)
	require.NoError(t, err)

	// Immediately get it (should exist)
	retrievedValue, found, err := cache.Get(key)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, value, retrievedValue)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Try to get it again (should be expired)
	_, found, err = cache.Get(key)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestAlgorithmCache_Stats(t *testing.T) {
	cache, err := NewAlgorithmCache(nil)
	require.NoError(t, err)
	defer cache.Close()

	// Initial stats should be zero
	stats := cache.GetStats()
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(0), stats.Misses)

	// Set a value
	err = cache.Set("key1", "value1", 5*time.Minute)
	require.NoError(t, err)

	// Get the value (should be a hit)
	_, found, err := cache.Get("key1")
	require.NoError(t, err)
	assert.True(t, found)

	// Get a non-existent value (should be a miss)
	_, found, err = cache.Get("non-existent")
	require.NoError(t, err)
	assert.False(t, found)

	// Check stats
	stats = cache.GetStats()
	assert.Equal(t, int64(1), stats.Hits)
	assert.Equal(t, int64(1), stats.Misses)
	assert.Equal(t, int64(2), stats.TotalRequests)
}

func TestAlgorithmCache_KeySizeLimit(t *testing.T) {
	config := &CacheConfig{
		MaxMemoryEntries: 100,
		MemoryTTL:        5 * time.Minute,
		EvictionPolicy:   "LRU",
		MaxKeySize:       10,
		MaxValueSize:     1024,
		CleanupInterval:  30 * time.Second,
	}
	cache, err := NewAlgorithmCache(config)
	require.NoError(t, err)
	defer cache.Close()

	// Try to set a key that's too long
	longKey := "this-key-is-way-too-long-for-the-limit"
	err = cache.Set(longKey, "value", 5*time.Minute)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key size exceeds maximum")
}

func TestAlgorithmCache_ValueSizeLimit(t *testing.T) {
	config := &CacheConfig{
		MaxMemoryEntries: 100,
		MemoryTTL:        5 * time.Minute,
		EvictionPolicy:   "LRU",
		MaxKeySize:       256,
		MaxValueSize:     10,
		CleanupInterval:  30 * time.Second,
	}
	cache, err := NewAlgorithmCache(config)
	require.NoError(t, err)
	defer cache.Close()

	// Try to set a value that's too large
	largeValue := "this-value-is-way-too-large-for-the-limit"
	err = cache.Set("key", largeValue, 5*time.Minute)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "value size exceeds maximum")
}

func TestAlgorithmCache_EvictionPolicies(t *testing.T) {
	tests := []struct {
		name   string
		policy string
	}{
		{"LRU", "LRU"},
		{"LFU", "LFU"},
		{"TTL", "TTL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &CacheConfig{
				MaxMemoryEntries: 2,
				MemoryTTL:        5 * time.Minute,
				EvictionPolicy:   tt.policy,
				MaxKeySize:       256,
				MaxValueSize:     1024,
				CleanupInterval:  30 * time.Second,
			}
			cache, err := NewAlgorithmCache(config)
			require.NoError(t, err)
			defer cache.Close()

			// Fill cache to capacity
			err = cache.Set("key1", "value1", 5*time.Minute)
			require.NoError(t, err)
			err = cache.Set("key2", "value2", 5*time.Minute)
			require.NoError(t, err)

			// Add one more to trigger eviction
			err = cache.Set("key3", "value3", 5*time.Minute)
			require.NoError(t, err)

			// Verify cache still works
			_, found, err := cache.Get("key3")
			require.NoError(t, err)
			assert.True(t, found)
		})
	}
}

func TestDefaultCacheConfig(t *testing.T) {
	config := DefaultCacheConfig()
	assert.NotNil(t, config)
	assert.Greater(t, config.MaxMemoryEntries, 0)
	assert.Greater(t, config.MemoryTTL, time.Duration(0))
	assert.NotEmpty(t, config.EvictionPolicy)
	assert.Greater(t, config.MaxKeySize, 0)
	assert.Greater(t, config.MaxValueSize, 0)
}

func TestAlgorithmCache_ConcurrentAccess(t *testing.T) {
	cache, err := NewAlgorithmCache(nil)
	require.NoError(t, err)
	defer cache.Close()

	// Test concurrent reads and writes
	done := make(chan bool, 10)

	// Start multiple goroutines writing
	for i := 0; i < 5; i++ {
		go func(id int) {
			key := fmt.Sprintf("key-%d", id)
			value := fmt.Sprintf("value-%d", id)
			err := cache.Set(key, value, 5*time.Minute)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Start multiple goroutines reading
	for i := 0; i < 5; i++ {
		go func(id int) {
			key := fmt.Sprintf("key-%d", id)
			_, _, err := cache.Get(key)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

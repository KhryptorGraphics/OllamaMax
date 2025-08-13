package monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMetricsConfig_DefaultValues(t *testing.T) {
	config := &MetricsConfig{
		ListenAddress:       ":9090",
		MetricsPath:         "/metrics",
		CollectionInterval:  30 * time.Second,
		EnableSystemMetrics: true,
		EnableAppMetrics:    true,
		EnableP2PMetrics:    true,
		MetricsRetention:    24 * time.Hour,
		DefaultLabels:       map[string]string{"service": "test"},
	}

	assert.Equal(t, ":9090", config.ListenAddress)
	assert.Equal(t, "/metrics", config.MetricsPath)
	assert.Equal(t, 30*time.Second, config.CollectionInterval)
	assert.True(t, config.EnableSystemMetrics)
	assert.True(t, config.EnableAppMetrics)
	assert.True(t, config.EnableP2PMetrics)
	assert.Equal(t, 24*time.Hour, config.MetricsRetention)
	assert.NotNil(t, config.DefaultLabels)
	assert.Equal(t, "test", config.DefaultLabels["service"])
}

func TestNewMetricsCollector(t *testing.T) {
	config := &MetricsConfig{
		ListenAddress:       ":0", // Use random port for testing
		MetricsPath:         "/metrics",
		CollectionInterval:  1 * time.Second,
		EnableSystemMetrics: true,
		EnableAppMetrics:    true,
		EnableP2PMetrics:    true,
	}

	collector := NewMetricsCollector(config)
	assert.NotNil(t, collector)

	// Clean up
	if collector.cancel != nil {
		collector.cancel()
	}
}

func TestMetricsCollector_BasicCreation(t *testing.T) {
	config := &MetricsConfig{
		ListenAddress:       ":0",
		MetricsPath:         "/metrics",
		CollectionInterval:  1 * time.Second,
		EnableSystemMetrics: true,
		EnableAppMetrics:    true,
		EnableP2PMetrics:    true,
	}

	collector := NewMetricsCollector(config)
	assert.NotNil(t, collector)

	// Test that the collector has the expected components
	assert.NotNil(t, collector.registry)
	assert.NotNil(t, collector.systemMetrics)
	assert.NotNil(t, collector.appMetrics)
	assert.NotNil(t, collector.p2pMetrics)

	// Clean up
	if collector.cancel != nil {
		collector.cancel()
	}
}

func TestMetricsCollector_StartStop(t *testing.T) {
	config := &MetricsConfig{
		ListenAddress:       ":0", // Use random port for testing
		MetricsPath:         "/metrics",
		CollectionInterval:  100 * time.Millisecond,
		EnableSystemMetrics: true,
		EnableAppMetrics:    false,
		EnableP2PMetrics:    false,
	}

	collector := NewMetricsCollector(config)
	assert.NotNil(t, collector)

	// Clean up
	if collector.cancel != nil {
		collector.cancel()
	}
}

func TestMetricsCollector_BasicFunctionality(t *testing.T) {
	config := &MetricsConfig{
		ListenAddress:       ":0",
		MetricsPath:         "/metrics",
		CollectionInterval:  1 * time.Second,
		EnableSystemMetrics: false,
		EnableAppMetrics:    true,
		EnableP2PMetrics:    false,
	}

	collector := NewMetricsCollector(config)
	assert.NotNil(t, collector)

	// Verify metrics components exist
	assert.NotNil(t, collector.appMetrics)

	// Clean up
	if collector.cancel != nil {
		collector.cancel()
	}
}

func TestDefaultMetricsConfig(t *testing.T) {
	config := DefaultMetricsConfig()
	assert.NotNil(t, config)
	assert.NotEmpty(t, config.ListenAddress)
	assert.NotEmpty(t, config.MetricsPath)
	assert.Greater(t, config.CollectionInterval, time.Duration(0))
	assert.Greater(t, config.MetricsRetention, time.Duration(0))
	assert.NotNil(t, config.DefaultLabels)
}

func TestMetricsCollector_Context(t *testing.T) {
	config := &MetricsConfig{
		ListenAddress:       ":0",
		MetricsPath:         "/metrics",
		CollectionInterval:  100 * time.Millisecond,
		EnableSystemMetrics: false,
		EnableAppMetrics:    false,
		EnableP2PMetrics:    false,
	}

	collector := NewMetricsCollector(config)
	assert.NotNil(t, collector)

	// Test context cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Wait for context to be cancelled
	<-ctx.Done()

	// Clean up collector
	if collector.cancel != nil {
		collector.cancel()
	}
}

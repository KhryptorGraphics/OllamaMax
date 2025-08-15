package fault_tolerance

import (
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigValidator_BasicValidation tests basic configuration validation
func TestConfigValidator_BasicValidation(t *testing.T) {
	validator := NewConfigValidator()

	// Create a minimal valid config for testing
	createTestConfig := func() *config.DistributedConfig {
		cfg := &config.DistributedConfig{}
		cfg.Inference.FaultTolerance.Enabled = true
		cfg.Inference.FaultTolerance.RetryAttempts = 3
		cfg.Inference.FaultTolerance.RetryDelay = 1 * time.Second
		cfg.Inference.FaultTolerance.HealthCheckInterval = 30 * time.Second
		cfg.Inference.FaultTolerance.RecoveryTimeout = 5 * time.Minute
		cfg.Inference.FaultTolerance.CircuitBreaker.Enabled = true
		cfg.Inference.FaultTolerance.CheckpointInterval = 30 * time.Second
		cfg.Inference.FaultTolerance.MaxRetries = 3
		cfg.Inference.FaultTolerance.RetryBackoff = 5 * time.Second
		cfg.Inference.FaultTolerance.ReplicationFactor = 2

		// Predictive detection
		cfg.Inference.FaultTolerance.PredictiveDetection.Enabled = true
		cfg.Inference.FaultTolerance.PredictiveDetection.ConfidenceThreshold = 0.8
		cfg.Inference.FaultTolerance.PredictiveDetection.PredictionInterval = "30s"
		cfg.Inference.FaultTolerance.PredictiveDetection.WindowSize = "30s"
		cfg.Inference.FaultTolerance.PredictiveDetection.Threshold = 0.8
		cfg.Inference.FaultTolerance.PredictiveDetection.EnableMLDetection = false
		cfg.Inference.FaultTolerance.PredictiveDetection.EnableStatistical = true
		cfg.Inference.FaultTolerance.PredictiveDetection.EnablePatternRecog = true

		// Self-healing
		cfg.Inference.FaultTolerance.SelfHealing.Enabled = true
		cfg.Inference.FaultTolerance.SelfHealing.HealingThreshold = 0.7
		cfg.Inference.FaultTolerance.SelfHealing.HealingInterval = "60s"
		cfg.Inference.FaultTolerance.SelfHealing.MonitoringInterval = "30s"
		cfg.Inference.FaultTolerance.SelfHealing.LearningInterval = "60s"
		cfg.Inference.FaultTolerance.SelfHealing.ServiceRestart = true
		cfg.Inference.FaultTolerance.SelfHealing.ResourceReallocation = true
		cfg.Inference.FaultTolerance.SelfHealing.LoadRedistribution = true
		cfg.Inference.FaultTolerance.SelfHealing.EnableLearning = true
		cfg.Inference.FaultTolerance.SelfHealing.EnablePredictive = true
		cfg.Inference.FaultTolerance.SelfHealing.EnableProactive = true
		cfg.Inference.FaultTolerance.SelfHealing.EnableFailover = true
		cfg.Inference.FaultTolerance.SelfHealing.EnableScaling = true

		// Redundancy
		cfg.Inference.FaultTolerance.Redundancy.Enabled = true
		cfg.Inference.FaultTolerance.Redundancy.DefaultFactor = 2
		cfg.Inference.FaultTolerance.Redundancy.MaxFactor = 5
		cfg.Inference.FaultTolerance.Redundancy.UpdateInterval = "5m"

		// Performance tracking (window should be 3x healing interval: 60s * 3 = 180s)
		cfg.Inference.FaultTolerance.PerformanceTracking.Enabled = true
		cfg.Inference.FaultTolerance.PerformanceTracking.WindowSize = "180s"

		// Config adaptation
		cfg.Inference.FaultTolerance.ConfigAdaptation.Enabled = true
		cfg.Inference.FaultTolerance.ConfigAdaptation.Interval = "5m"

		return cfg
	}

	t.Run("valid configuration", func(t *testing.T) {
		config := createTestConfig()
		err := validator.ValidateConfiguration(config)
		assert.NoError(t, err)
	})

	t.Run("invalid retry attempts - too high", func(t *testing.T) {
		config := createTestConfig()
		config.Inference.FaultTolerance.RetryAttempts = 15
		err := validator.ValidateConfiguration(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "retry_attempts must be between")
	})

	t.Run("invalid confidence threshold - too high", func(t *testing.T) {
		config := createTestConfig()
		config.Inference.FaultTolerance.PredictiveDetection.ConfidenceThreshold = 1.5
		err := validator.ValidateConfiguration(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "confidence_threshold must be between")
	})

	t.Run("invalid healing threshold - negative", func(t *testing.T) {
		config := createTestConfig()
		config.Inference.FaultTolerance.SelfHealing.HealingThreshold = -0.1
		err := validator.ValidateConfiguration(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "healing_threshold must be between")
	})

	t.Run("invalid redundancy factor relationship", func(t *testing.T) {
		config := createTestConfig()
		config.Inference.FaultTolerance.Redundancy.DefaultFactor = 5
		config.Inference.FaultTolerance.Redundancy.MaxFactor = 3
		err := validator.ValidateConfiguration(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "default_factor")
		assert.Contains(t, err.Error(), "cannot be greater than max_factor")
	})

	t.Run("invalid duration format", func(t *testing.T) {
		config := createTestConfig()
		config.Inference.FaultTolerance.RetryDelay = -1 * time.Second // Invalid duration
		err := validator.ValidateConfiguration(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid retry_delay format")
	})

	t.Run("no detection methods enabled", func(t *testing.T) {
		config := createTestConfig()
		config.Inference.FaultTolerance.PredictiveDetection.EnableMLDetection = false
		config.Inference.FaultTolerance.PredictiveDetection.EnableStatistical = false
		config.Inference.FaultTolerance.PredictiveDetection.EnablePatternRecog = false
		err := validator.ValidateConfiguration(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least one detection method must be enabled")
	})

	t.Run("no healing strategies enabled", func(t *testing.T) {
		config := createTestConfig()
		config.Inference.FaultTolerance.SelfHealing.ServiceRestart = false
		config.Inference.FaultTolerance.SelfHealing.ResourceReallocation = false
		config.Inference.FaultTolerance.SelfHealing.LoadRedistribution = false
		config.Inference.FaultTolerance.SelfHealing.EnableFailover = false
		config.Inference.FaultTolerance.SelfHealing.EnableScaling = false
		err := validator.ValidateConfiguration(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least one healing strategy must be enabled")
	})
}

// TestEnhancedFaultToleranceManager_LoadConfiguration tests configuration loading
func TestEnhancedFaultToleranceManager_LoadConfiguration(t *testing.T) {
	// Create base fault tolerance manager
	baseConfig := &Config{
		ReplicationFactor:     2,
		HealthCheckInterval:   30 * time.Second,
		RecoveryTimeout:       5 * time.Minute,
		CircuitBreakerEnabled: true,
		CheckpointInterval:    30 * time.Second,
		MaxRetries:            3,
		RetryBackoff:          5 * time.Second,
	}
	baseFT := NewFaultToleranceManager(baseConfig)

	// Create enhanced configuration
	enhancedConfig := NewEnhancedFaultToleranceConfig(baseConfig)
	eftm := NewEnhancedFaultToleranceManager(enhancedConfig, baseFT)

	createTestConfig := func() *config.DistributedConfig {
		cfg := &config.DistributedConfig{}
		cfg.Inference.FaultTolerance.Enabled = true
		cfg.Inference.FaultTolerance.RetryAttempts = 3
		cfg.Inference.FaultTolerance.RetryDelay = 1 * time.Second
		cfg.Inference.FaultTolerance.HealthCheckInterval = 30 * time.Second
		cfg.Inference.FaultTolerance.RecoveryTimeout = 5 * time.Minute
		cfg.Inference.FaultTolerance.CircuitBreaker.Enabled = true
		cfg.Inference.FaultTolerance.CheckpointInterval = 30 * time.Second
		cfg.Inference.FaultTolerance.MaxRetries = 3
		cfg.Inference.FaultTolerance.RetryBackoff = "5s"
		cfg.Inference.FaultTolerance.ReplicationFactor = 2

		// Predictive detection
		cfg.Inference.FaultTolerance.PredictiveDetection.Enabled = true
		cfg.Inference.FaultTolerance.PredictiveDetection.ConfidenceThreshold = 0.8
		cfg.Inference.FaultTolerance.PredictiveDetection.PredictionInterval = "30s"
		cfg.Inference.FaultTolerance.PredictiveDetection.WindowSize = "30s"
		cfg.Inference.FaultTolerance.PredictiveDetection.Threshold = 0.8
		cfg.Inference.FaultTolerance.PredictiveDetection.EnableMLDetection = false
		cfg.Inference.FaultTolerance.PredictiveDetection.EnableStatistical = true
		cfg.Inference.FaultTolerance.PredictiveDetection.EnablePatternRecog = true

		// Self-healing
		cfg.Inference.FaultTolerance.SelfHealing.Enabled = true
		cfg.Inference.FaultTolerance.SelfHealing.HealingThreshold = 0.7
		cfg.Inference.FaultTolerance.SelfHealing.HealingInterval = "60s"
		cfg.Inference.FaultTolerance.SelfHealing.MonitoringInterval = "30s"
		cfg.Inference.FaultTolerance.SelfHealing.LearningInterval = "60s"
		cfg.Inference.FaultTolerance.SelfHealing.ServiceRestart = true
		cfg.Inference.FaultTolerance.SelfHealing.ResourceReallocation = true
		cfg.Inference.FaultTolerance.SelfHealing.LoadRedistribution = true
		cfg.Inference.FaultTolerance.SelfHealing.EnableLearning = true
		cfg.Inference.FaultTolerance.SelfHealing.EnablePredictive = true
		cfg.Inference.FaultTolerance.SelfHealing.EnableProactive = true
		cfg.Inference.FaultTolerance.SelfHealing.EnableFailover = true
		cfg.Inference.FaultTolerance.SelfHealing.EnableScaling = true

		// Redundancy
		cfg.Inference.FaultTolerance.Redundancy.Enabled = true
		cfg.Inference.FaultTolerance.Redundancy.DefaultFactor = 2
		cfg.Inference.FaultTolerance.Redundancy.MaxFactor = 5
		cfg.Inference.FaultTolerance.Redundancy.UpdateInterval = "5m"

		// Performance tracking (window should be 3x healing interval: 60s * 3 = 180s)
		cfg.Inference.FaultTolerance.PerformanceTracking.Enabled = true
		cfg.Inference.FaultTolerance.PerformanceTracking.WindowSize = "180s"

		// Config adaptation
		cfg.Inference.FaultTolerance.ConfigAdaptation.Enabled = true
		cfg.Inference.FaultTolerance.ConfigAdaptation.Interval = "5m"

		return cfg
	}

	t.Run("successful configuration loading", func(t *testing.T) {
		config := createTestConfig()
		err := eftm.LoadConfiguration(config)
		assert.NoError(t, err)

		// Verify configuration was applied
		assert.Equal(t, 2, eftm.FaultToleranceManager.config.ReplicationFactor)
		assert.Equal(t, 3, eftm.FaultToleranceManager.config.MaxRetries)
		assert.True(t, eftm.FaultToleranceManager.config.CircuitBreakerEnabled)
	})

	t.Run("configuration validation failure", func(t *testing.T) {
		config := createTestConfig()
		config.Inference.FaultTolerance.RetryAttempts = -1 // Invalid
		err := eftm.LoadConfiguration(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuration validation failed")
	})

	t.Run("nil configuration", func(t *testing.T) {
		err := eftm.LoadConfiguration(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "distributed config cannot be nil")
	})
}

// TestConfigurationBehaviorChanges tests that configuration changes affect system behavior
func TestConfigurationBehaviorChanges(t *testing.T) {
	// Create base fault tolerance manager
	baseConfig := &Config{
		ReplicationFactor:     2,
		HealthCheckInterval:   30 * time.Second,
		RecoveryTimeout:       5 * time.Minute,
		CircuitBreakerEnabled: true,
		CheckpointInterval:    30 * time.Second,
		MaxRetries:            3,
		RetryBackoff:          5 * time.Second,
	}
	baseFT := NewFaultToleranceManager(baseConfig)

	// Create enhanced configuration
	enhancedConfig := NewEnhancedFaultToleranceConfig(baseConfig)
	eftm := NewEnhancedFaultToleranceManager(enhancedConfig, baseFT)

	t.Run("configuration changes affect component settings", func(t *testing.T) {
		// Create initial configuration
		config := &config.DistributedConfig{}
		config.Inference.FaultTolerance.Enabled = true
		config.Inference.FaultTolerance.RetryAttempts = 3
		config.Inference.FaultTolerance.RetryDelay = "1s"
		config.Inference.FaultTolerance.HealthCheckInterval = "30s"
		config.Inference.FaultTolerance.RecoveryTimeout = "5m"
		config.Inference.FaultTolerance.CircuitBreakerEnabled = true
		config.Inference.FaultTolerance.CheckpointInterval = "30s"
		config.Inference.FaultTolerance.MaxRetries = 3
		config.Inference.FaultTolerance.RetryBackoff = "5s"
		config.Inference.FaultTolerance.ReplicationFactor = 2

		// Load initial configuration
		err := eftm.LoadConfiguration(config)
		require.NoError(t, err)

		// Verify initial settings
		assert.Equal(t, 2, eftm.FaultToleranceManager.config.ReplicationFactor)
		assert.Equal(t, 3, eftm.FaultToleranceManager.config.MaxRetries)

		// Change configuration
		config.Inference.FaultTolerance.ReplicationFactor = 4
		config.Inference.FaultTolerance.MaxRetries = 5

		// Reload configuration
		err = eftm.LoadConfiguration(config)
		require.NoError(t, err)

		// Verify settings changed
		assert.Equal(t, 4, eftm.FaultToleranceManager.config.ReplicationFactor)
		assert.Equal(t, 5, eftm.FaultToleranceManager.config.MaxRetries)
	})
}

// TestHotReloadConfiguration tests hot-reload functionality
func TestHotReloadConfiguration(t *testing.T) {
	// Create base fault tolerance manager
	baseConfig := &Config{
		ReplicationFactor:     2,
		HealthCheckInterval:   30 * time.Second,
		RecoveryTimeout:       5 * time.Minute,
		CircuitBreakerEnabled: true,
		CheckpointInterval:    30 * time.Second,
		MaxRetries:            3,
		RetryBackoff:          5 * time.Second,
	}
	baseFT := NewFaultToleranceManager(baseConfig)

	// Create enhanced configuration
	enhancedConfig := NewEnhancedFaultToleranceConfig(baseConfig)
	eftm := NewEnhancedFaultToleranceManager(enhancedConfig, baseFT)

	createTestConfig := func() *config.DistributedConfig {
		cfg := &config.DistributedConfig{}
		cfg.Inference.FaultTolerance.Enabled = true
		cfg.Inference.FaultTolerance.RetryAttempts = 3
		cfg.Inference.FaultTolerance.RetryDelay = "1s"
		cfg.Inference.FaultTolerance.HealthCheckInterval = "30s"
		cfg.Inference.FaultTolerance.RecoveryTimeout = "5m"
		cfg.Inference.FaultTolerance.CircuitBreakerEnabled = true
		cfg.Inference.FaultTolerance.CheckpointInterval = "30s"
		cfg.Inference.FaultTolerance.MaxRetries = 3
		cfg.Inference.FaultTolerance.RetryBackoff = "5s"
		cfg.Inference.FaultTolerance.ReplicationFactor = 2

		// Predictive detection
		cfg.Inference.FaultTolerance.PredictiveDetection.Enabled = true
		cfg.Inference.FaultTolerance.PredictiveDetection.ConfidenceThreshold = 0.8
		cfg.Inference.FaultTolerance.PredictiveDetection.PredictionInterval = "30s"
		cfg.Inference.FaultTolerance.PredictiveDetection.WindowSize = "30s"
		cfg.Inference.FaultTolerance.PredictiveDetection.Threshold = 0.8
		cfg.Inference.FaultTolerance.PredictiveDetection.EnableMLDetection = false
		cfg.Inference.FaultTolerance.PredictiveDetection.EnableStatistical = true
		cfg.Inference.FaultTolerance.PredictiveDetection.EnablePatternRecog = true

		// Self-healing
		cfg.Inference.FaultTolerance.SelfHealing.Enabled = true
		cfg.Inference.FaultTolerance.SelfHealing.HealingThreshold = 0.7
		cfg.Inference.FaultTolerance.SelfHealing.HealingInterval = "60s"
		cfg.Inference.FaultTolerance.SelfHealing.MonitoringInterval = "30s"
		cfg.Inference.FaultTolerance.SelfHealing.LearningInterval = "60s"
		cfg.Inference.FaultTolerance.SelfHealing.ServiceRestart = true
		cfg.Inference.FaultTolerance.SelfHealing.ResourceReallocation = true
		cfg.Inference.FaultTolerance.SelfHealing.LoadRedistribution = true
		cfg.Inference.FaultTolerance.SelfHealing.EnableLearning = true
		cfg.Inference.FaultTolerance.SelfHealing.EnablePredictive = true
		cfg.Inference.FaultTolerance.SelfHealing.EnableProactive = true
		cfg.Inference.FaultTolerance.SelfHealing.EnableFailover = true
		cfg.Inference.FaultTolerance.SelfHealing.EnableScaling = true

		// Redundancy
		cfg.Inference.FaultTolerance.Redundancy.Enabled = true
		cfg.Inference.FaultTolerance.Redundancy.DefaultFactor = 2
		cfg.Inference.FaultTolerance.Redundancy.MaxFactor = 5
		cfg.Inference.FaultTolerance.Redundancy.UpdateInterval = "5m"

		// Performance tracking (window should be 3x healing interval: 60s * 3 = 180s)
		cfg.Inference.FaultTolerance.PerformanceTracking.Enabled = true
		cfg.Inference.FaultTolerance.PerformanceTracking.WindowSize = "180s"

		// Config adaptation
		cfg.Inference.FaultTolerance.ConfigAdaptation.Enabled = true
		cfg.Inference.FaultTolerance.ConfigAdaptation.Interval = "5m"

		return cfg
	}

	t.Run("successful hot reload", func(t *testing.T) {
		// Load initial configuration
		config := createTestConfig()
		err := eftm.LoadConfiguration(config)
		require.NoError(t, err)

		// Verify initial settings
		assert.Equal(t, 2, eftm.FaultToleranceManager.config.ReplicationFactor)
		assert.Equal(t, 3, eftm.FaultToleranceManager.config.MaxRetries)

		// Change configuration
		config.Inference.FaultTolerance.ReplicationFactor = 4
		config.Inference.FaultTolerance.MaxRetries = 5

		// Hot reload configuration
		err = eftm.ReloadConfiguration(config)
		require.NoError(t, err)

		// Verify settings changed
		assert.Equal(t, 4, eftm.FaultToleranceManager.config.ReplicationFactor)
		assert.Equal(t, 5, eftm.FaultToleranceManager.config.MaxRetries)
	})

	t.Run("hot reload with validation failure and rollback", func(t *testing.T) {
		// Load initial valid configuration
		config := createTestConfig()
		err := eftm.LoadConfiguration(config)
		require.NoError(t, err)

		// Store initial settings
		initialReplicationFactor := eftm.FaultToleranceManager.config.ReplicationFactor
		initialMaxRetries := eftm.FaultToleranceManager.config.MaxRetries

		// Create invalid configuration
		invalidConfig := createTestConfig()
		invalidConfig.Inference.FaultTolerance.RetryAttempts = -1 // Invalid

		// Attempt hot reload with invalid config
		err = eftm.ReloadConfiguration(invalidConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuration validation failed during reload")

		// Verify settings were not changed (rollback successful)
		assert.Equal(t, initialReplicationFactor, eftm.FaultToleranceManager.config.ReplicationFactor)
		assert.Equal(t, initialMaxRetries, eftm.FaultToleranceManager.config.MaxRetries)
	})

	t.Run("hot reload with nil configuration", func(t *testing.T) {
		err := eftm.ReloadConfiguration(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "distributed config cannot be nil")
	})
}

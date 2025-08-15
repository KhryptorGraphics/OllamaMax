package fault_tolerance

import (
	"fmt"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
)

// ConfigValidator provides comprehensive validation for fault tolerance configuration
type ConfigValidator struct {
	// Validation rules and constraints
	minRetryAttempts       int
	maxRetryAttempts       int
	minReplicationFactor   int
	maxReplicationFactor   int
	minConfidenceThreshold float64
	maxConfidenceThreshold float64
	minHealingThreshold    float64
	maxHealingThreshold    float64
	minRedundancyFactor    int
	maxRedundancyFactor    int

	// Duration constraints
	minHealthCheckInterval time.Duration
	maxHealthCheckInterval time.Duration
	minRecoveryTimeout     time.Duration
	maxRecoveryTimeout     time.Duration
	minHealingInterval     time.Duration
	maxHealingInterval     time.Duration
}

// NewConfigValidator creates a new configuration validator with default constraints
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{
		minRetryAttempts:       0,
		maxRetryAttempts:       10,
		minReplicationFactor:   1,
		maxReplicationFactor:   10,
		minConfidenceThreshold: 0.0,
		maxConfidenceThreshold: 1.0,
		minHealingThreshold:    0.0,
		maxHealingThreshold:    1.0,
		minRedundancyFactor:    1,
		maxRedundancyFactor:    20,
		minHealthCheckInterval: 5 * time.Second,
		maxHealthCheckInterval: 10 * time.Minute,
		minRecoveryTimeout:     30 * time.Second,
		maxRecoveryTimeout:     30 * time.Minute,
		minHealingInterval:     10 * time.Second,
		maxHealingInterval:     30 * time.Minute,
	}
}

// ValidateConfiguration performs comprehensive validation of fault tolerance configuration
func (cv *ConfigValidator) ValidateConfiguration(distributedConfig *config.DistributedConfig) error {
	if distributedConfig == nil {
		return fmt.Errorf("distributed config cannot be nil")
	}

	ft := distributedConfig.Inference.FaultTolerance

	// Validate basic fault tolerance settings
	if err := cv.validateBasicSettings(ft); err != nil {
		return fmt.Errorf("basic settings validation failed: %w", err)
	}

	// Validate predictive detection settings
	if ft.PredictiveDetection.Enabled {
		if err := cv.validatePredictiveDetection(ft.PredictiveDetection); err != nil {
			return fmt.Errorf("predictive detection validation failed: %w", err)
		}
	}

	// Validate self-healing settings
	if ft.SelfHealing.Enabled {
		if err := cv.validateSelfHealing(ft.SelfHealing); err != nil {
			return fmt.Errorf("self-healing validation failed: %w", err)
		}
	}

	// Validate redundancy settings
	if ft.Redundancy.Enabled {
		if err := cv.validateRedundancy(ft.Redundancy); err != nil {
			return fmt.Errorf("redundancy validation failed: %w", err)
		}
	}

	// Validate performance tracking settings
	if ft.PerformanceTracking.Enabled {
		if err := cv.validatePerformanceTracking(ft.PerformanceTracking); err != nil {
			return fmt.Errorf("performance tracking validation failed: %w", err)
		}
	}

	// Validate config adaptation settings
	if ft.ConfigAdaptation.Enabled {
		if err := cv.validateConfigAdaptation(ft.ConfigAdaptation); err != nil {
			return fmt.Errorf("config adaptation validation failed: %w", err)
		}
	}

	// Cross-validation between different settings
	if err := cv.validateCrossSettings(ft); err != nil {
		return fmt.Errorf("cross-settings validation failed: %w", err)
	}

	return nil
}

// validateBasicSettings validates basic fault tolerance settings
func (cv *ConfigValidator) validateBasicSettings(ft config.FaultToleranceConfig) error {
	// Validate retry attempts
	if ft.RetryAttempts < cv.minRetryAttempts || ft.RetryAttempts > cv.maxRetryAttempts {
		return fmt.Errorf("retry_attempts must be between %d and %d, got %d",
			cv.minRetryAttempts, cv.maxRetryAttempts, ft.RetryAttempts)
	}

	// Validate max retries
	if ft.MaxRetries < cv.minRetryAttempts || ft.MaxRetries > cv.maxRetryAttempts {
		return fmt.Errorf("max_retries must be between %d and %d, got %d",
			cv.minRetryAttempts, cv.maxRetryAttempts, ft.MaxRetries)
	}

	// Validate replication factor
	if ft.ReplicationFactor < cv.minReplicationFactor || ft.ReplicationFactor > cv.maxReplicationFactor {
		return fmt.Errorf("replication_factor must be between %d and %d, got %d",
			cv.minReplicationFactor, cv.maxReplicationFactor, ft.ReplicationFactor)
	}

	// Validate duration ranges
	if ft.HealthCheckInterval < cv.minHealthCheckInterval || ft.HealthCheckInterval > cv.maxHealthCheckInterval {
		return fmt.Errorf("health_check_interval must be between %v and %v, got %v", cv.minHealthCheckInterval, cv.maxHealthCheckInterval, ft.HealthCheckInterval)
	}
	if ft.RecoveryTimeout < cv.minRecoveryTimeout || ft.RecoveryTimeout > cv.maxRecoveryTimeout {
		return fmt.Errorf("recovery_timeout must be between %v and %v, got %v", cv.minRecoveryTimeout, cv.maxRecoveryTimeout, ft.RecoveryTimeout)
	}
	if ft.CheckpointInterval <= 0 {
		return fmt.Errorf("checkpoint_interval must be > 0, got %v", ft.CheckpointInterval)
	}
	if ft.RetryBackoff <= 0 {
		return fmt.Errorf("retry_backoff must be > 0, got %v", ft.RetryBackoff)
	}

	return nil
}

// validateDurationString validates that a string can be parsed as a duration
func (cv *ConfigValidator) validateDurationString(durationStr, fieldName string) error {
	if durationStr == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}

	_, err := time.ParseDuration(durationStr)
	if err != nil {
		return fmt.Errorf("invalid %s format: %w", fieldName, err)
	}

	return nil
}

// validateDurationRange validates that a duration string falls within specified range
func (cv *ConfigValidator) validateDurationRange(durationStr string, min, max time.Duration, fieldName string) error {
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return fmt.Errorf("invalid %s format: %w", fieldName, err)
	}

	if duration < min || duration > max {
		return fmt.Errorf("%s must be between %v and %v, got %v", fieldName, min, max, duration)
	}

	return nil
}

// validatePredictiveDetection validates predictive detection configuration
func (cv *ConfigValidator) validatePredictiveDetection(config struct {
	Enabled             bool    `yaml:"enabled"`
	ConfidenceThreshold float64 `yaml:"confidence_threshold"`
	PredictionInterval  string  `yaml:"prediction_interval"`
	WindowSize          string  `yaml:"window_size"`
	Threshold           float64 `yaml:"threshold"`
	EnableMLDetection   bool    `yaml:"enable_ml_detection"`
	EnableStatistical   bool    `yaml:"enable_statistical"`
	EnablePatternRecog  bool    `yaml:"enable_pattern_recognition"`
}) error {
	// Validate confidence threshold
	if config.ConfidenceThreshold < cv.minConfidenceThreshold || config.ConfidenceThreshold > cv.maxConfidenceThreshold {
		return fmt.Errorf("confidence_threshold must be between %f and %f, got %f",
			cv.minConfidenceThreshold, cv.maxConfidenceThreshold, config.ConfidenceThreshold)
	}

	// Validate threshold
	if config.Threshold < cv.minConfidenceThreshold || config.Threshold > cv.maxConfidenceThreshold {
		return fmt.Errorf("threshold must be between %f and %f, got %f",
			cv.minConfidenceThreshold, cv.maxConfidenceThreshold, config.Threshold)
	}

	// Validate duration strings
	if err := cv.validateDurationString(config.PredictionInterval, "prediction_interval"); err != nil {
		return err
	}

	if err := cv.validateDurationString(config.WindowSize, "window_size"); err != nil {
		return err
	}

	// Validate that at least one detection method is enabled
	if !config.EnableMLDetection && !config.EnableStatistical && !config.EnablePatternRecog {
		return fmt.Errorf("at least one detection method must be enabled (ml_detection, statistical, or pattern_recognition)")
	}

	return nil
}

// validateSelfHealing validates self-healing configuration
func (cv *ConfigValidator) validateSelfHealing(config struct {
	Enabled              bool    `yaml:"enabled"`
	HealingThreshold     float64 `yaml:"healing_threshold"`
	HealingInterval      string  `yaml:"healing_interval"`
	MonitoringInterval   string  `yaml:"monitoring_interval"`
	LearningInterval     string  `yaml:"learning_interval"`
	ServiceRestart       bool    `yaml:"service_restart"`
	ResourceReallocation bool    `yaml:"resource_reallocation"`
	LoadRedistribution   bool    `yaml:"load_redistribution"`
	EnableLearning       bool    `yaml:"enable_learning"`
	EnablePredictive     bool    `yaml:"enable_predictive"`
	EnableProactive      bool    `yaml:"enable_proactive"`
	EnableFailover       bool    `yaml:"enable_failover"`
	EnableScaling        bool    `yaml:"enable_scaling"`
}) error {
	// Validate healing threshold
	if config.HealingThreshold < cv.minHealingThreshold || config.HealingThreshold > cv.maxHealingThreshold {
		return fmt.Errorf("healing_threshold must be between %f and %f, got %f",
			cv.minHealingThreshold, cv.maxHealingThreshold, config.HealingThreshold)
	}

	// Validate duration strings
	if err := cv.validateDurationString(config.HealingInterval, "healing_interval"); err != nil {
		return err
	}

	if err := cv.validateDurationString(config.MonitoringInterval, "monitoring_interval"); err != nil {
		return err
	}

	if err := cv.validateDurationString(config.LearningInterval, "learning_interval"); err != nil {
		return err
	}

	// Validate duration ranges
	if err := cv.validateDurationRange(config.HealingInterval, cv.minHealingInterval, cv.maxHealingInterval, "healing_interval"); err != nil {
		return err
	}

	// Validate that at least one healing strategy is enabled
	if !config.ServiceRestart && !config.ResourceReallocation && !config.LoadRedistribution && !config.EnableFailover && !config.EnableScaling {
		return fmt.Errorf("at least one healing strategy must be enabled (service_restart, resource_reallocation, load_redistribution, failover, or scaling)")
	}

	return nil
}

// validateRedundancy validates redundancy configuration
func (cv *ConfigValidator) validateRedundancy(config struct {
	Enabled        bool   `yaml:"enabled"`
	DefaultFactor  int    `yaml:"default_factor"`
	MaxFactor      int    `yaml:"max_factor"`
	UpdateInterval string `yaml:"update_interval"`
}) error {
	// Validate default factor
	if config.DefaultFactor < cv.minRedundancyFactor || config.DefaultFactor > cv.maxRedundancyFactor {
		return fmt.Errorf("default_factor must be between %d and %d, got %d",
			cv.minRedundancyFactor, cv.maxRedundancyFactor, config.DefaultFactor)
	}

	// Validate max factor
	if config.MaxFactor < cv.minRedundancyFactor || config.MaxFactor > cv.maxRedundancyFactor {
		return fmt.Errorf("max_factor must be between %d and %d, got %d",
			cv.minRedundancyFactor, cv.maxRedundancyFactor, config.MaxFactor)
	}

	// Validate that default factor is not greater than max factor
	if config.DefaultFactor > config.MaxFactor {
		return fmt.Errorf("default_factor (%d) cannot be greater than max_factor (%d)",
			config.DefaultFactor, config.MaxFactor)
	}

	// Validate duration string
	if err := cv.validateDurationString(config.UpdateInterval, "update_interval"); err != nil {
		return err
	}

	return nil
}

// validatePerformanceTracking validates performance tracking configuration
func (cv *ConfigValidator) validatePerformanceTracking(config struct {
	Enabled    bool   `yaml:"enabled"`
	WindowSize string `yaml:"window_size"`
}) error {
	// Validate duration string
	if err := cv.validateDurationString(config.WindowSize, "window_size"); err != nil {
		return err
	}

	// Validate window size range (should be reasonable for performance tracking)
	windowSize, _ := time.ParseDuration(config.WindowSize)
	if windowSize < 10*time.Second || windowSize > 24*time.Hour {
		return fmt.Errorf("performance tracking window_size must be between 10s and 24h, got %v", windowSize)
	}

	return nil
}

// validateConfigAdaptation validates configuration adaptation settings
func (cv *ConfigValidator) validateConfigAdaptation(config struct {
	Enabled  bool   `yaml:"enabled"`
	Interval string `yaml:"interval"`
}) error {
	// Validate duration string
	if err := cv.validateDurationString(config.Interval, "interval"); err != nil {
		return err
	}

	// Validate interval range (should be reasonable for config adaptation)
	interval, _ := time.ParseDuration(config.Interval)
	if interval < 1*time.Minute || interval > 24*time.Hour {
		return fmt.Errorf("config adaptation interval must be between 1m and 24h, got %v", interval)
	}

	return nil
}

// validateCrossSettings validates cross-dependencies between different configuration sections
func (cv *ConfigValidator) validateCrossSettings(ft config.FaultToleranceConfig) error {
	// Validate that predictive healing requires predictive detection
	if ft.SelfHealing.Enabled && ft.SelfHealing.EnablePredictive && !ft.PredictiveDetection.Enabled {
		return fmt.Errorf("predictive self-healing requires predictive detection to be enabled")
	}

	// Validate timing relationships
	if ft.SelfHealing.Enabled && ft.PredictiveDetection.Enabled {
		healingInterval, _ := time.ParseDuration(ft.SelfHealing.HealingInterval)
		predictionInterval, _ := time.ParseDuration(ft.PredictiveDetection.PredictionInterval)

		// Healing interval should be longer than prediction interval for effective prediction
		if healingInterval <= predictionInterval {
			return fmt.Errorf("healing_interval (%v) should be longer than prediction_interval (%v) for effective predictive healing",
				healingInterval, predictionInterval)
		}
	}

	// Validate that redundancy factor doesn't exceed replication factor unnecessarily
	if ft.Redundancy.Enabled && ft.Redundancy.DefaultFactor > ft.ReplicationFactor*2 {
		return fmt.Errorf("redundancy default_factor (%d) seems excessive compared to replication_factor (%d)",
			ft.Redundancy.DefaultFactor, ft.ReplicationFactor)
	}

	// Validate monitoring intervals make sense
	if ft.SelfHealing.Enabled {
		healingInterval, _ := time.ParseDuration(ft.SelfHealing.HealingInterval)
		monitoringInterval, _ := time.ParseDuration(ft.SelfHealing.MonitoringInterval)

		// Monitoring should be more frequent than healing
		if monitoringInterval >= healingInterval {
			return fmt.Errorf("monitoring_interval (%v) should be shorter than healing_interval (%v)",
				monitoringInterval, healingInterval)
		}
	}

	// Validate performance tracking window vs other intervals
	if ft.PerformanceTracking.Enabled && ft.SelfHealing.Enabled {
		performanceWindow, _ := time.ParseDuration(ft.PerformanceTracking.WindowSize)
		healingInterval, _ := time.ParseDuration(ft.SelfHealing.HealingInterval)

		// Performance window should be long enough to capture multiple healing cycles
		if performanceWindow < healingInterval*3 {
			return fmt.Errorf("performance tracking window_size (%v) should be at least 3x healing_interval (%v) for meaningful analysis",
				performanceWindow, healingInterval)
		}
	}

	return nil
}

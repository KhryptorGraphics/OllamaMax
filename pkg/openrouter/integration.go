package openrouter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

// Integration manages OpenRouter integration with OllamaMax
type Integration struct {
	client       *Client
	config       *IntegrationConfig
	logger       *logrus.Logger
	models       map[string]*ModelInfo
	muModelRegistry sync.RWMutex
	healthChecker *HealthChecker
	loadBalancer  *LoadBalancer
	costTracker   *CostTracker
	cacheManager  *CacheManager
}

// IntegrationConfig holds integration configuration
type IntegrationConfig struct {
	OpenRouter OpenRouterConfig `yaml:"openrouter"`
	Integration struct {
		RegisterModels bool              `yaml:"register_models"`
		Priority       map[string]int    `yaml:"priority"`
		Fallback       FallbackConfig    `yaml:"fallback"`
		HybridRouting  HybridRoutingConfig `yaml:"hybrid_routing"`
	} `yaml:"integration"`
	Security SecurityConfig `yaml:"security"`
	Monitoring MonitoringConfig `yaml:"monitoring"`
}

// OpenRouterConfig holds OpenRouter-specific configuration
type OpenRouterConfig struct {
	Enabled    bool                  `yaml:"enabled"`
	APIKey     string                `yaml:"api_key"`
	BaseURL    string                `yaml:"base_url"`
	Timeout    string                `yaml:"timeout"`
	MaxRetries int                   `yaml:"max_retries"`
	RetryDelay string                `yaml:"retry_delay"`
	Models     map[string]ModelConfig `yaml:"models"`
	Routing    RoutingConfig         `yaml:"routing"`
	Requests   RequestsConfig        `yaml:"requests"`
	Cache      CacheConfig           `yaml:"cache"`
	Monitoring MonitoringConfig      `yaml:"monitoring"`
}

// ModelConfig holds model-specific configuration
type ModelConfig struct {
	Name         string                 `yaml:"name"`
	DisplayName  string                 `yaml:"display_name"`
	Description  string                 `yaml:"description"`
	MaxTokens    int                    `yaml:"max_tokens"`
	CostPerToken map[string]float64     `yaml:"cost_per_token"`
	Capabilities []string               `yaml:"capabilities"`
	Parameters   map[string]interface{} `yaml:"parameters"`
}

// FallbackConfig holds fallback configuration
type FallbackConfig struct {
	Enabled           bool     `yaml:"enabled"`
	LocalModels       []string `yaml:"local_models"`
	FallbackThreshold float64  `yaml:"fallback_threshold"`
}

// HybridRoutingConfig holds hybrid routing configuration
type HybridRoutingConfig struct {
	Enabled         bool    `yaml:"enabled"`
	OpenRouterWeight float64 `yaml:"openrouter_weight"`
	LocalWeight     float64 `yaml:"local_weight"`
}

// RoutingConfig holds routing configuration
type RoutingConfig struct {
	Strategy            string        `yaml:"strategy"`
	FallbackModels      []string      `yaml:"fallback_models"`
	HealthCheckInterval string        `yaml:"health_check_interval"`
}

// RequestsConfig holds request configuration
type RequestsConfig struct {
	MaxConcurrent    int               `yaml:"max_concurrent"`
	QueueSize        int               `yaml:"queue_size"`
	TimeoutPerRequest string           `yaml:"timeout_per_request"`
	RateLimit        RateLimitConfig   `yaml:"rate_limit"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int `yaml:"requests_per_minute"`
	BurstSize         int `yaml:"burst_size"`
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Enabled bool   `yaml:"enabled"`
	TTL     string `yaml:"ttl"`
	MaxSize string `yaml:"max_size"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	APIKeyRotation     APIKeyRotationConfig `yaml:"api_key_rotation"`
	RequestValidation  RequestValidationConfig `yaml:"request_validation"`
	Audit              AuditConfig          `yaml:"audit"`
}

// APIKeyRotationConfig holds API key rotation configuration
type APIKeyRotationConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Interval string `yaml:"interval"`
}

// RequestValidationConfig holds request validation configuration
type RequestValidationConfig struct {
	Enabled           bool     `yaml:"enabled"`
	MaxPromptLength   int      `yaml:"max_prompt_length"`
	BlockedPatterns   []string `yaml:"blocked_patterns"`
}

// AuditConfig holds audit configuration
type AuditConfig struct {
	Enabled bool   `yaml:"enabled"`
	LogFile string `yaml:"log_file"`
}

// MonitoringConfig holds monitoring configuration
type MonitoringConfig struct {
	Enabled      bool `yaml:"enabled"`
	LogRequests  bool `yaml:"log_requests"`
	TrackUsage   bool `yaml:"track_usage"`
	CostTracking bool `yaml:"cost_tracking"`
}

// ModelInfo holds model information
type ModelInfo struct {
	ID           string
	Name         string
	DisplayName  string
	Description  string
	MaxTokens    int
	Capabilities []string
	Priority     int
	CostPerToken map[string]float64
	Healthy      bool
	LastChecked  time.Time
}

// NewIntegration creates a new OpenRouter integration
func NewIntegration(configPath string) (*Integration, error) {
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	
	if !config.OpenRouter.Enabled {
		return nil, fmt.Errorf("OpenRouter integration is disabled")
	}
	
	// Parse timeout
	timeout, err := time.ParseDuration(config.OpenRouter.Timeout)
	if err != nil {
		timeout = 300 * time.Second
	}
	
	// Parse retry delay
	retryDelay, err := time.ParseDuration(config.OpenRouter.RetryDelay)
	if err != nil {
		retryDelay = 2 * time.Second
	}
	
	// Create OpenRouter client
	clientConfig := &Config{
		APIKey:     config.OpenRouter.APIKey,
		BaseURL:    config.OpenRouter.BaseURL,
		Timeout:    timeout,
		MaxRetries: config.OpenRouter.MaxRetries,
		RetryDelay: retryDelay,
		UserAgent:  "OllamaMax/1.0.0",
		DebugMode:  false,
	}
	
	client := NewClient(clientConfig)
	
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	
	integration := &Integration{
		client:       client,
		config:       config,
		logger:       logger,
		models:       make(map[string]*ModelInfo),
		healthChecker: NewHealthChecker(client, logger),
		loadBalancer:  NewLoadBalancer(config),
		costTracker:   NewCostTracker(config),
		cacheManager:  NewCacheManager(config),
	}
	
	// Initialize models
	if err := integration.initializeModels(); err != nil {
		return nil, fmt.Errorf("failed to initialize models: %w", err)
	}
	
	// Start health checker
	integration.healthChecker.Start()
	
	return integration, nil
}

// LoadConfig loads configuration from file
func LoadConfig(configPath string) (*IntegrationConfig, error) {
	// This would typically read from a YAML file
	// For now, return a default configuration
	config := &IntegrationConfig{
		OpenRouter: OpenRouterConfig{
			Enabled:    true,
			BaseURL:    "https://openrouter.ai/api/v1",
			Timeout:    "300s",
			MaxRetries: 3,
			RetryDelay: "2s",
			Models: map[string]ModelConfig{
				"sonoma-sky-alpha": {
					Name:        "alpindale/sonoma-sky-alpha",
					DisplayName: "Sonoma Sky Alpha",
					Description: "Advanced reasoning model with superior performance",
					MaxTokens:   131072,
					CostPerToken: map[string]float64{
						"prompt":     0.000003,
						"completion": 0.000015,
					},
					Capabilities: []string{"text-generation", "reasoning", "code-generation", "analysis"},
				},
			},
		},
	}
	
	config.Integration.RegisterModels = true
	config.Integration.Priority = map[string]int{
		"sonoma-sky-alpha": 100,
	}
	config.Integration.Fallback.Enabled = true
	config.Integration.Fallback.LocalModels = []string{"llama3.2:latest", "qwen2.5:latest"}
	config.Integration.Fallback.FallbackThreshold = 0.9
	
	config.Monitoring.Enabled = true
	config.Monitoring.LogRequests = true
	config.Monitoring.TrackUsage = true
	config.Monitoring.CostTracking = true
	
	return config, nil
}

// initializeModels initializes the model registry
func (i *Integration) initializeModels() error {
	i.muModelRegistry.Lock()
	defer i.muModelRegistry.Unlock()
	
	// Add configured models
	for modelID, modelConfig := range i.config.OpenRouter.Models {
		modelInfo := &ModelInfo{
			ID:           modelID,
			Name:         modelConfig.Name,
			DisplayName:  modelConfig.DisplayName,
			Description:  modelConfig.Description,
			MaxTokens:    modelConfig.MaxTokens,
			Capabilities: modelConfig.Capabilities,
			CostPerToken: modelConfig.CostPerToken,
			Healthy:      true,
			LastChecked:  time.Now(),
		}
		
		if priority, exists := i.config.Integration.Priority[modelID]; exists {
			modelInfo.Priority = priority
		}
		
		i.models[modelID] = modelInfo
		i.logger.WithFields(logrus.Fields{
			"model_id": modelID,
			"name": modelConfig.Name,
			"priority": modelInfo.Priority,
		}).Info("Registered OpenRouter model")
	}
	
	return nil
}

// ChatCompletion performs a chat completion using OpenRouter
func (i *Integration) ChatCompletion(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// Validate request
	if err := i.client.ValidateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}
	
	// Check cache
	if i.config.OpenRouter.Cache.Enabled {
		if cached := i.cacheManager.Get(req); cached != nil {
			i.logger.Debug("Returning cached response")
			return cached, nil
		}
	}
	
	// Track cost
	if i.config.Monitoring.CostTracking {
		i.costTracker.TrackRequest(req)
	}
	
	// Perform request
	response, err := i.client.ChatCompletion(ctx, req)
	if err != nil {
		return nil, err
	}
	
	// Cache response
	if i.config.OpenRouter.Cache.Enabled {
		i.cacheManager.Set(req, response)
	}
	
	// Track usage
	if i.config.Monitoring.TrackUsage {
		i.costTracker.TrackResponse(response)
	}
	
	return response, nil
}

// GetModels returns available models
func (i *Integration) GetModels() map[string]*ModelInfo {
	i.muModelRegistry.RLock()
	defer i.muModelRegistry.RUnlock()
	
	models := make(map[string]*ModelInfo)
	for id, model := range i.models {
		models[id] = model
	}
	
	return models
}

// GetModel returns a specific model
func (i *Integration) GetModel(modelID string) (*ModelInfo, bool) {
	i.muModelRegistry.RLock()
	defer i.muModelRegistry.RUnlock()
	
	model, exists := i.models[modelID]
	return model, exists
}

// IsHealthy checks if the integration is healthy
func (i *Integration) IsHealthy() bool {
	return i.healthChecker.IsHealthy()
}

// GetStats returns integration statistics
func (i *Integration) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"models_count": len(i.models),
		"healthy": i.IsHealthy(),
	}
	
	if i.config.Monitoring.CostTracking {
		stats["cost_stats"] = i.costTracker.GetStats()
	}
	
	if i.config.Monitoring.TrackUsage {
		stats["usage_stats"] = i.costTracker.GetUsageStats()
	}
	
	return stats
}

// Stop stops the integration
func (i *Integration) Stop() {
	i.healthChecker.Stop()
	i.logger.Info("OpenRouter integration stopped")
}

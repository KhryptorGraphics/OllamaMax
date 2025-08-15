package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// DistributedConfig represents the complete configuration for OllamaMax distributed system
type DistributedConfig struct {
	// Core Configuration
	Node      NodeConfig      `yaml:"node"`
	API       APIConfig       `yaml:"api"`
	P2P       P2PConfig       `yaml:"p2p"`
	Consensus ConsensusConfig `yaml:"consensus"`
	Scheduler SchedulerConfig `yaml:"scheduler"`
	Storage   StorageConfig   `yaml:"storage"`
	Security  SecurityConfig  `yaml:"security"`
	Web       WebConfig       `yaml:"web"`
	Metrics   MetricsConfig   `yaml:"metrics"`
	Logging   LoggingConfig   `yaml:"logging"`

	// Advanced Features
	Sync        SyncConfig        `yaml:"sync"`
	Replication ReplicationConfig `yaml:"replication"`
	Models      ModelsConfig      `yaml:"models"`
	Inference   InferenceConfig   `yaml:"inference"`

	// Frontend Integration
	Auth         AuthConfig         `yaml:"auth"`
	WebSocket    WebSocketConfig    `yaml:"websocket"`
	Notification NotificationConfig `yaml:"notification"`
	PWA          PWAConfig          `yaml:"pwa"`
	I18n         I18nConfig         `yaml:"i18n"`

	// Performance & Monitoring
	Performance   PerformanceConfig   `yaml:"performance"`
	Observability ObservabilityConfig `yaml:"observability"`
}

// NodeConfig holds node-specific configuration
type NodeConfig struct {
	ID                 string             `yaml:"id"`
	Name               string             `yaml:"name"`
	Region             string             `yaml:"region"`
	Zone               string             `yaml:"zone"`
	Environment        string             `yaml:"environment"`
	Tags               map[string]string  `yaml:"tags"`
	Resources          ResourceConfig     `yaml:"resources"`
	Listen             []string           `yaml:"listen"`
	EnableNoise        bool               `yaml:"enable_noise"`
	EnableTLS          bool               `yaml:"enable_tls"`
	EnableNATService   bool               `yaml:"enable_nat_service"`
	EnableHolePunching bool               `yaml:"enable_hole_punching"`
	EnableAutoRelay    bool               `yaml:"enable_auto_relay"`
	StaticRelays       []string           `yaml:"static_relays"`
	Capabilities       NodeCapabilities   `yaml:"capabilities"`
	TURNServers        []TURNServerConfig `yaml:"turn_servers"`
	ConnMgrLow         int                `yaml:"conn_mgr_low"`
	ConnMgrHigh        int                `yaml:"conn_mgr_high"`
	ConnMgrGrace       time.Duration      `yaml:"conn_mgr_grace"`
	AnnounceAddresses  []string           `yaml:"announce_addresses"`
}

// Discovery interface adapter methods for P2P discovery package
func (nc *NodeConfig) GetBootstrapPeers() []string {
	// Prefer P2P.BootstrapNodes if available, else StaticRelays as a fallback for bootstrap
	if nc != nil {
		if nc.AnnounceAddresses != nil && len(nc.AnnounceAddresses) > 0 {
			return nc.AnnounceAddresses
		}
		// In many configs, bootstrap nodes live under top-level P2P; if not present, return empty
	}
	return nil
}

func (nc *NodeConfig) GetRendezvousString() string {
	// Fallback rendezvous tag
	return "ollama-discovery"
}

func (nc *NodeConfig) IsAutoDiscoveryEnabled() bool {
	// Use Discovery.Enabled when available through P2P config; default true to enable discovery
	return true
}

// TURNServerConfig defines TURN server connection settings
type TURNServerConfig struct {
	Address   string `yaml:"address"`
	Port      int    `yaml:"port"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	Realm     string `yaml:"realm"`
	Transport string `yaml:"transport"` // "udp" or "tcp"
}

// NodeCapabilities defines what capabilities a node has
type NodeCapabilities struct {
	Inference    bool `yaml:"inference"`
	Storage      bool `yaml:"storage"`
	Coordination bool `yaml:"coordination"`
	Gateway      bool `yaml:"gateway"`
}

// ResourceConfig defines node resource limits
type ResourceConfig struct {
	CPU    string `yaml:"cpu"`    // e.g., "4" or "2.5"
	Memory string `yaml:"memory"` // e.g., "8Gi"
	GPU    string `yaml:"gpu"`    // e.g., "1" or "nvidia.com/gpu=2"
	Disk   string `yaml:"disk"`   // e.g., "100Gi"
}

// APIConfig holds API server configuration
type APIConfig struct {
	Host        string           `yaml:"host"`
	Port        int              `yaml:"port"`
	TLS         TLSConfig        `yaml:"tls"`
	CORS        CORSConfig       `yaml:"cors"`
	RateLimit   RateLimitConfig  `yaml:"rate_limit"`
	Timeout     time.Duration    `yaml:"timeout"`
	MaxBodySize int64            `yaml:"max_body_size"`
	Middleware  MiddlewareConfig `yaml:"middleware"`
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
	CAFile   string `yaml:"ca_file"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
	AllowedMethods []string `yaml:"allowed_methods"`
	AllowedHeaders []string `yaml:"allowed_headers"`
	MaxAge         int      `yaml:"max_age"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled bool `yaml:"enabled"`
	RPS     int  `yaml:"rps"`
	Burst   int  `yaml:"burst"`
}

// MiddlewareConfig holds middleware configuration
type MiddlewareConfig struct {
	Auth       bool `yaml:"auth"`
	Logging    bool `yaml:"logging"`
	Metrics    bool `yaml:"metrics"`
	Recovery   bool `yaml:"recovery"`
	Validation bool `yaml:"validation"`
}

// P2PConfig holds peer-to-peer networking configuration
type P2PConfig struct {
	Port           int               `yaml:"port"`
	BootstrapNodes []string          `yaml:"bootstrap_nodes"`
	MaxPeers       int               `yaml:"max_peers"`
	Discovery      DiscoveryConfig   `yaml:"discovery"`
	NAT            NATConfig         `yaml:"nat"`
	Security       P2PSecurityConfig `yaml:"security"`
}

// DiscoveryConfig holds peer discovery configuration
type DiscoveryConfig struct {
	Enabled  bool          `yaml:"enabled"`
	Method   string        `yaml:"method"` // "mdns", "dht", "bootstrap"
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
}

// NATConfig holds NAT traversal configuration
type NATConfig struct {
	Enabled bool   `yaml:"enabled"`
	STUN    string `yaml:"stun_server"`
	TURN    string `yaml:"turn_server"`
}

// P2PSecurityConfig holds P2P security configuration
type P2PSecurityConfig struct {
	Encryption bool   `yaml:"encryption"`
	KeyFile    string `yaml:"key_file"`
}

// ConsensusConfig holds consensus algorithm configuration
type ConsensusConfig struct {
	Algorithm         string        `yaml:"algorithm"` // "raft", "pbft", "tendermint"
	ElectionTimeout   time.Duration `yaml:"election_timeout"`
	HeartbeatInterval time.Duration `yaml:"heartbeat_interval"`
	MaxLogEntries     int           `yaml:"max_log_entries"`
	SnapshotThreshold int           `yaml:"snapshot_threshold"`
}

// SchedulerConfig holds task scheduling configuration
type SchedulerConfig struct {
	Strategy      string         `yaml:"strategy"` // "round_robin", "least_loaded", "affinity"
	MaxConcurrent int            `yaml:"max_concurrent"`
	Timeout       time.Duration  `yaml:"timeout"`
	RetryPolicy   RetryConfig    `yaml:"retry_policy"`
	Affinity      AffinityConfig `yaml:"affinity"`
}

// RetryConfig holds retry policy configuration
type RetryConfig struct {
	MaxRetries int           `yaml:"max_retries"`
	Backoff    time.Duration `yaml:"backoff"`
	MaxBackoff time.Duration `yaml:"max_backoff"`
}

// AffinityConfig holds scheduling affinity rules
type AffinityConfig struct {
	ModelAffinity bool              `yaml:"model_affinity"`
	NodeAffinity  map[string]string `yaml:"node_affinity"`
	AntiAffinity  []string          `yaml:"anti_affinity"`
}

// StorageConfig holds storage configuration
type StorageConfig struct {
	Type        string       `yaml:"type"` // "memory", "disk", "distributed"
	Path        string       `yaml:"path"`
	MaxSize     string       `yaml:"max_size"`
	Compression bool         `yaml:"compression"`
	Encryption  bool         `yaml:"encryption"`
	Backup      BackupConfig `yaml:"backup"`
}

// BackupConfig holds backup configuration
type BackupConfig struct {
	Enabled   bool          `yaml:"enabled"`
	Interval  time.Duration `yaml:"interval"`
	Retention time.Duration `yaml:"retention"`
	Location  string        `yaml:"location"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	TLS        TLSConfig        `yaml:"tls"`
	Auth       AuthConfig       `yaml:"auth"`
	Encryption EncryptionConfig `yaml:"encryption"`
	Firewall   FirewallConfig   `yaml:"firewall"`
	Audit      AuditConfig      `yaml:"audit"`
}

// EncryptionConfig holds encryption configuration
type EncryptionConfig struct {
	AtRest    bool   `yaml:"at_rest"`
	InTransit bool   `yaml:"in_transit"`
	Algorithm string `yaml:"algorithm"`
	KeyFile   string `yaml:"key_file"`
}

// FirewallConfig holds firewall configuration
type FirewallConfig struct {
	Enabled      bool     `yaml:"enabled"`
	AllowedIPs   []string `yaml:"allowed_ips"`
	BlockedIPs   []string `yaml:"blocked_ips"`
	AllowedPorts []int    `yaml:"allowed_ports"`
}

// AuditConfig holds audit logging configuration
type AuditConfig struct {
	Enabled bool   `yaml:"enabled"`
	LogFile string `yaml:"log_file"`
	Format  string `yaml:"format"`
}

// WebConfig holds web interface configuration
type WebConfig struct {
	Enabled     bool   `yaml:"enabled"`
	StaticPath  string `yaml:"static_path"`
	TemplateDir string `yaml:"template_dir"`
	DevMode     bool   `yaml:"dev_mode"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Provider       string        `yaml:"provider"` // "jwt", "oauth2", "ldap", "saml"
	JWT            JWTConfig     `yaml:"jwt"`
	OAuth2         OAuth2Config  `yaml:"oauth2"`
	LDAP           LDAPConfig    `yaml:"ldap"`
	SAML           SAMLConfig    `yaml:"saml"`
	SessionTimeout time.Duration `yaml:"session_timeout"`
	MaxSessions    int           `yaml:"max_sessions"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret         string        `yaml:"secret"`
	Issuer         string        `yaml:"issuer"`
	Audience       string        `yaml:"audience"`
	ExpirationTime time.Duration `yaml:"expiration_time"`
	RefreshTime    time.Duration `yaml:"refresh_time"`
}

// OAuth2Config holds OAuth2 configuration
type OAuth2Config struct {
	ClientID     string   `yaml:"client_id"`
	ClientSecret string   `yaml:"client_secret"`
	RedirectURL  string   `yaml:"redirect_url"`
	Scopes       []string `yaml:"scopes"`
	AuthURL      string   `yaml:"auth_url"`
	TokenURL     string   `yaml:"token_url"`
}

// LDAPConfig holds LDAP configuration
type LDAPConfig struct {
	Server   string `yaml:"server"`
	Port     int    `yaml:"port"`
	BaseDN   string `yaml:"base_dn"`
	UserDN   string `yaml:"user_dn"`
	BindUser string `yaml:"bind_user"`
	BindPass string `yaml:"bind_pass"`
}

// SAMLConfig holds SAML configuration
type SAMLConfig struct {
	EntityID    string `yaml:"entity_id"`
	SSOURL      string `yaml:"sso_url"`
	Certificate string `yaml:"certificate"`
	PrivateKey  string `yaml:"private_key"`
}

// WebSocketConfig holds WebSocket configuration
type WebSocketConfig struct {
	Enabled        bool          `yaml:"enabled"`
	Path           string        `yaml:"path"`
	MaxConnections int           `yaml:"max_connections"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout"`
	PingInterval   time.Duration `yaml:"ping_interval"`
}

// NotificationConfig holds notification configuration
type NotificationConfig struct {
	Push    PushConfig    `yaml:"push"`
	Email   EmailConfig   `yaml:"email"`
	Webhook WebhookConfig `yaml:"webhook"`
	Slack   SlackConfig   `yaml:"slack"`
}

// PushConfig holds push notification configuration
type PushConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Provider  string `yaml:"provider"` // "firebase", "apns", "web"
	APIKey    string `yaml:"api_key"`
	ProjectID string `yaml:"project_id"`
	VAPIDKey  string `yaml:"vapid_key"`
}

// EmailConfig holds email notification configuration
type EmailConfig struct {
	Enabled  bool       `yaml:"enabled"`
	SMTP     SMTPConfig `yaml:"smtp"`
	From     string     `yaml:"from"`
	Template string     `yaml:"template"`
}

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	TLS      bool   `yaml:"tls"`
}

// WebhookConfig holds webhook notification configuration
type WebhookConfig struct {
	Enabled bool          `yaml:"enabled"`
	URLs    []string      `yaml:"urls"`
	Timeout time.Duration `yaml:"timeout"`
	Retries int           `yaml:"retries"`
}

// SlackConfig holds Slack notification configuration
type SlackConfig struct {
	Enabled    bool   `yaml:"enabled"`
	WebhookURL string `yaml:"webhook_url"`
	Channel    string `yaml:"channel"`
	Username   string `yaml:"username"`
}

// PWAConfig holds Progressive Web App configuration
type PWAConfig struct {
	Enabled       bool     `yaml:"enabled"`
	ManifestPath  string   `yaml:"manifest_path"`
	ServiceWorker string   `yaml:"service_worker"`
	OfflinePages  []string `yaml:"offline_pages"`
	CacheStrategy string   `yaml:"cache_strategy"`
}

// I18nConfig holds internationalization configuration
type I18nConfig struct {
	Enabled          bool     `yaml:"enabled"`
	DefaultLanguage  string   `yaml:"default_language"`
	Languages        []string `yaml:"languages"`
	TranslationPath  string   `yaml:"translation_path"`
	FallbackLanguage string   `yaml:"fallback_language"`
}

// MetricsConfig holds metrics collection configuration
type MetricsConfig struct {
	Enabled    bool             `yaml:"enabled"`
	Port       int              `yaml:"port"`
	Path       string           `yaml:"path"`
	Interval   time.Duration    `yaml:"interval"`
	Retention  time.Duration    `yaml:"retention"`
	Prometheus PrometheusConfig `yaml:"prometheus"`
}

// PrometheusConfig holds Prometheus configuration
type PrometheusConfig struct {
	Enabled   bool              `yaml:"enabled"`
	Endpoint  string            `yaml:"endpoint"`
	Namespace string            `yaml:"namespace"`
	Labels    map[string]string `yaml:"labels"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	Output     string `yaml:"output"`
	File       string `yaml:"file"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

// SyncConfig holds data synchronization configuration
type SyncConfig struct {
	Enabled     bool          `yaml:"enabled"`
	Interval    time.Duration `yaml:"interval"`
	BatchSize   int           `yaml:"batch_size"`
	Timeout     time.Duration `yaml:"timeout"`
	Compression bool          `yaml:"compression"`
	Encryption  bool          `yaml:"encryption"`
}

// ReplicationConfig holds data replication configuration
type ReplicationConfig struct {
	Enabled     bool          `yaml:"enabled"`
	Factor      int           `yaml:"factor"`
	Strategy    string        `yaml:"strategy"` // "sync", "async", "quorum"
	Timeout     time.Duration `yaml:"timeout"`
	Consistency string        `yaml:"consistency"` // "strong", "eventual", "weak"
	CrossRegion bool          `yaml:"cross_region"`
}

// ModelsConfig holds AI model configuration
type ModelsConfig struct {
	Path       string            `yaml:"path"`
	Cache      ModelCacheConfig  `yaml:"cache"`
	Preload    []string          `yaml:"preload"`
	MaxModels  int               `yaml:"max_models"`
	Affinity   map[string]string `yaml:"affinity"`
	Versioning bool              `yaml:"versioning"`
}

// ModelCacheConfig holds model caching configuration
type ModelCacheConfig struct {
	Enabled  bool          `yaml:"enabled"`
	Size     string        `yaml:"size"`
	TTL      time.Duration `yaml:"ttl"`
	Strategy string        `yaml:"strategy"` // "lru", "lfu", "fifo"
}

// InferenceConfig holds inference engine configuration
type InferenceConfig struct {
	MaxConcurrent  int                  `yaml:"max_concurrent"`
	Timeout        time.Duration        `yaml:"timeout"`
	BatchSize      int                  `yaml:"batch_size"`
	GPU            GPUConfig            `yaml:"gpu"`
	Optimization   OptimizationConfig   `yaml:"optimization"`
	FaultTolerance FaultToleranceConfig `yaml:"fault_tolerance"`
}

// FaultToleranceConfig holds fault tolerance configuration
type FaultToleranceConfig struct {
	Enabled             bool                      `yaml:"enabled"`
	RetryAttempts       int                       `yaml:"retry_attempts"`
	RetryDelay          time.Duration             `yaml:"retry_delay"`
	CircuitBreaker      CircuitBreakerConfig      `yaml:"circuit_breaker"`
	HealthCheckInterval time.Duration             `yaml:"health_check_interval"`
	RecoveryTimeout     time.Duration             `yaml:"recovery_timeout"`
	CheckpointInterval  time.Duration             `yaml:"checkpoint_interval"`
	MaxRetries          int                       `yaml:"max_retries"`
	RetryBackoff        time.Duration             `yaml:"retry_backoff"`
	ReplicationFactor   int                       `yaml:"replication_factor"`
	PredictiveDetection PredictiveDetectionConfig `yaml:"predictive_detection"`
	SelfHealing         SelfHealingConfig         `yaml:"self_healing"`
	Redundancy          RedundancyConfig          `yaml:"redundancy"`
	PerformanceTracking PerformanceTrackingConfig `yaml:"performance_tracking"`
	ConfigAdaptation    ConfigAdaptationConfig    `yaml:"config_adaptation"`
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	Enabled          bool          `yaml:"enabled"`
	FailureThreshold int           `yaml:"failure_threshold"`
	RecoveryTimeout  time.Duration `yaml:"recovery_timeout"`
	Timeout          time.Duration `yaml:"timeout"`
}

// PredictiveDetectionConfig holds predictive detection configuration
type PredictiveDetectionConfig struct {
	Enabled             bool    `yaml:"enabled"`
	ConfidenceThreshold float64 `yaml:"confidence_threshold"`
	PredictionInterval  string  `yaml:"prediction_interval"`
	WindowSize          string  `yaml:"window_size"`
	Threshold           float64 `yaml:"threshold"`
	EnableMLDetection   bool    `yaml:"enable_ml_detection"`
	EnableStatistical   bool    `yaml:"enable_statistical"`
	EnablePatternRecog  bool    `yaml:"enable_pattern_recognition"`
}

// SelfHealingConfig holds self-healing configuration
type SelfHealingConfig struct {
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
}

// RedundancyConfig holds redundancy configuration
type RedundancyConfig struct {
	Enabled        bool   `yaml:"enabled"`
	DefaultFactor  int    `yaml:"default_factor"`
	MaxFactor      int    `yaml:"max_factor"`
	UpdateInterval string `yaml:"update_interval"`
}

// PerformanceTrackingConfig holds performance tracking configuration
type PerformanceTrackingConfig struct {
	Enabled    bool   `yaml:"enabled"`
	WindowSize string `yaml:"window_size"`
}

// ConfigAdaptationConfig holds configuration adaptation settings
type ConfigAdaptationConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Interval string `yaml:"interval"`
}

// GPUConfig holds GPU configuration
type GPUConfig struct {
	Enabled     bool     `yaml:"enabled"`
	Devices     []string `yaml:"devices"`
	MemoryLimit string   `yaml:"memory_limit"`
	Sharing     bool     `yaml:"sharing"`
}

// OptimizationConfig holds optimization configuration
type OptimizationConfig struct {
	Quantization bool   `yaml:"quantization"`
	Pruning      bool   `yaml:"pruning"`
	Compilation  bool   `yaml:"compilation"`
	Backend      string `yaml:"backend"` // "cpu", "cuda", "opencl", "metal"
}

// PerformanceConfig holds performance optimization configuration
type PerformanceConfig struct {
	Profiling    ProfilingConfig    `yaml:"profiling"`
	Caching      CachingConfig      `yaml:"caching"`
	Compression  CompressionConfig  `yaml:"compression"`
	Optimization OptimizationConfig `yaml:"optimization"`
	Limits       LimitsConfig       `yaml:"limits"`
}

// ProfilingConfig holds profiling configuration
type ProfilingConfig struct {
	Enabled    bool   `yaml:"enabled"`
	CPU        bool   `yaml:"cpu"`
	Memory     bool   `yaml:"memory"`
	Goroutine  bool   `yaml:"goroutine"`
	Block      bool   `yaml:"block"`
	Mutex      bool   `yaml:"mutex"`
	OutputPath string `yaml:"output_path"`
}

// CachingConfig holds caching configuration
type CachingConfig struct {
	Enabled  bool          `yaml:"enabled"`
	Type     string        `yaml:"type"` // "memory", "redis", "memcached"
	Size     string        `yaml:"size"`
	TTL      time.Duration `yaml:"ttl"`
	Eviction string        `yaml:"eviction"` // "lru", "lfu", "ttl"
}

// CompressionConfig holds compression configuration
type CompressionConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Algorithm string `yaml:"algorithm"` // "gzip", "lz4", "zstd"
	Level     int    `yaml:"level"`
	MinSize   int    `yaml:"min_size"`
}

// LimitsConfig holds resource limits configuration
type LimitsConfig struct {
	MaxConnections int           `yaml:"max_connections"`
	MaxRequests    int           `yaml:"max_requests"`
	RequestTimeout time.Duration `yaml:"request_timeout"`
	MemoryLimit    string        `yaml:"memory_limit"`
	CPULimit       string        `yaml:"cpu_limit"`
}

// ObservabilityConfig holds observability configuration
type ObservabilityConfig struct {
	Tracing    TracingConfig    `yaml:"tracing"`
	Monitoring MonitoringConfig `yaml:"monitoring"`
	Alerting   AlertingConfig   `yaml:"alerting"`
	Health     HealthConfig     `yaml:"health"`
}

// TracingConfig holds distributed tracing configuration
type TracingConfig struct {
	Enabled     bool    `yaml:"enabled"`
	Provider    string  `yaml:"provider"` // "jaeger", "zipkin", "datadog"
	Endpoint    string  `yaml:"endpoint"`
	ServiceName string  `yaml:"service_name"`
	SampleRate  float64 `yaml:"sample_rate"`
}

// MonitoringConfig holds monitoring configuration
type MonitoringConfig struct {
	Enabled   bool              `yaml:"enabled"`
	Interval  time.Duration     `yaml:"interval"`
	Endpoints []string          `yaml:"endpoints"`
	Labels    map[string]string `yaml:"labels"`
	Retention time.Duration     `yaml:"retention"`
}

// AlertingConfig holds alerting configuration
type AlertingConfig struct {
	Enabled    bool             `yaml:"enabled"`
	Rules      []AlertRule      `yaml:"rules"`
	Channels   []AlertChannel   `yaml:"channels"`
	Escalation EscalationConfig `yaml:"escalation"`
}

// AlertRule holds alert rule configuration
type AlertRule struct {
	Name        string        `yaml:"name"`
	Query       string        `yaml:"query"`
	Threshold   float64       `yaml:"threshold"`
	Duration    time.Duration `yaml:"duration"`
	Severity    string        `yaml:"severity"`
	Description string        `yaml:"description"`
}

// AlertChannel holds alert channel configuration
type AlertChannel struct {
	Name     string            `yaml:"name"`
	Type     string            `yaml:"type"` // "email", "slack", "webhook", "pagerduty"
	Config   map[string]string `yaml:"config"`
	Severity []string          `yaml:"severity"`
}

// EscalationConfig holds escalation configuration
type EscalationConfig struct {
	Enabled bool             `yaml:"enabled"`
	Rules   []EscalationRule `yaml:"rules"`
}

// EscalationRule holds escalation rule configuration
type EscalationRule struct {
	Severity string        `yaml:"severity"`
	Delay    time.Duration `yaml:"delay"`
	Channels []string      `yaml:"channels"`
}

// HealthConfig holds health check configuration
type HealthConfig struct {
	Enabled   bool             `yaml:"enabled"`
	Interval  time.Duration    `yaml:"interval"`
	Timeout   time.Duration    `yaml:"timeout"`
	Endpoints []HealthEndpoint `yaml:"endpoints"`
}

// HealthEndpoint holds health endpoint configuration
type HealthEndpoint struct {
	Name     string            `yaml:"name"`
	URL      string            `yaml:"url"`
	Method   string            `yaml:"method"`
	Headers  map[string]string `yaml:"headers"`
	Expected int               `yaml:"expected"`
}

// Configuration loading and validation functions

// LoadConfig loads configuration from file
func LoadConfig(path string) (*DistributedConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config DistributedConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

// Validate validates the configuration
func (c *DistributedConfig) Validate() error {
	if c.Node.ID == "" {
		return fmt.Errorf("node.id is required")
	}

	if c.API.Port <= 0 || c.API.Port > 65535 {
		return fmt.Errorf("api.port must be between 1 and 65535")
	}

	if c.P2P.Port <= 0 || c.P2P.Port > 65535 {
		return fmt.Errorf("p2p.port must be between 1 and 65535")
	}

	if c.Consensus.Algorithm == "" {
		return fmt.Errorf("consensus.algorithm is required")
	}

	return nil
}

// SetDefaults sets default values for configuration
func (c *DistributedConfig) SetDefaults() {
	// Node defaults
	if c.Node.Environment == "" {
		c.Node.Environment = "development"
	}

	// API defaults
	if c.API.Host == "" {
		c.API.Host = "0.0.0.0"
	}
	if c.API.Port == 0 {
		c.API.Port = 8080
	}
	if c.API.Timeout == 0 {
		c.API.Timeout = 30 * time.Second
	}

	// P2P defaults
	if c.P2P.Port == 0 {
		c.P2P.Port = 9000
	}
	if c.P2P.MaxPeers == 0 {
		c.P2P.MaxPeers = 50
	}

	// Consensus defaults
	if c.Consensus.Algorithm == "" {
		c.Consensus.Algorithm = "raft"
	}
	if c.Consensus.ElectionTimeout == 0 {
		c.Consensus.ElectionTimeout = 5 * time.Second
	}
	if c.Consensus.HeartbeatInterval == 0 {
		c.Consensus.HeartbeatInterval = 1 * time.Second
	}

	// Scheduler defaults
	if c.Scheduler.Strategy == "" {
		c.Scheduler.Strategy = "round_robin"
	}
	if c.Scheduler.MaxConcurrent == 0 {
		c.Scheduler.MaxConcurrent = 100
	}
	if c.Scheduler.Timeout == 0 {
		c.Scheduler.Timeout = 60 * time.Second
	}

	// Auth defaults
	if c.Auth.Provider == "" {
		c.Auth.Provider = "jwt"
	}
	if c.Auth.SessionTimeout == 0 {
		c.Auth.SessionTimeout = 24 * time.Hour
	}

	// WebSocket defaults
	if c.WebSocket.Path == "" {
		c.WebSocket.Path = "/ws"
	}
	if c.WebSocket.MaxConnections == 0 {
		c.WebSocket.MaxConnections = 1000
	}

	// Metrics defaults
	if c.Metrics.Port == 0 {
		c.Metrics.Port = 9090
	}
	if c.Metrics.Path == "" {
		c.Metrics.Path = "/metrics"
	}
	if c.Metrics.Interval == 0 {
		c.Metrics.Interval = 15 * time.Second
	}

	// Logging defaults
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "json"
	}
	if c.Logging.Output == "" {
		c.Logging.Output = "stdout"
	}
}

// Additional helper functions

// GetConfigPath returns the default configuration file path
func GetConfigPath() string {
	if path := os.Getenv("OLLAMA_CONFIG_PATH"); path != "" {
		return path
	}
	return "/etc/ollama-distributed/config.yaml"
}

// SaveConfig saves configuration to file
func (c *DistributedConfig) SaveConfig(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// MergeConfig merges another configuration into this one
func (c *DistributedConfig) MergeConfig(other *DistributedConfig) {
	if other.Node.ID != "" {
		c.Node.ID = other.Node.ID
	}
	if other.Node.Name != "" {
		c.Node.Name = other.Node.Name
	}
	if other.API.Port != 0 {
		c.API.Port = other.API.Port
	}
	// Add more merge logic as needed
}

// Clone creates a deep copy of the configuration
func (c *DistributedConfig) Clone() *DistributedConfig {
	data, _ := yaml.Marshal(c)
	var clone DistributedConfig
	yaml.Unmarshal(data, &clone)
	return &clone
}

// GetNodeTags returns node tags as a map
func (c *DistributedConfig) GetNodeTags() map[string]string {
	if c.Node.Tags == nil {
		return make(map[string]string)
	}
	return c.Node.Tags
}

// IsProduction returns true if running in production environment
func (c *DistributedConfig) IsProduction() bool {
	return c.Node.Environment == "production"
}

// IsDevelopment returns true if running in development environment
func (c *DistributedConfig) IsDevelopment() bool {
	return c.Node.Environment == "development"
}

// GetAPIAddress returns the full API address
func (c *DistributedConfig) GetAPIAddress() string {
	return fmt.Sprintf("%s:%d", c.API.Host, c.API.Port)
}

// GetP2PAddress returns the full P2P address
func (c *DistributedConfig) GetP2PAddress() string {
	return fmt.Sprintf(":%d", c.P2P.Port)
}

// GetMetricsAddress returns the full metrics address
func (c *DistributedConfig) GetMetricsAddress() string {
	return fmt.Sprintf(":%d", c.Metrics.Port)
}

// ParseStaticRelays parses static relay addresses
func (n *NodeConfig) ParseStaticRelays() ([]string, error) {
	return n.StaticRelays, nil
}

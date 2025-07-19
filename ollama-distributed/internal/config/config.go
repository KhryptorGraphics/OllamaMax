package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// Config represents the complete configuration for a distributed Ollama node
type Config struct {
	// Node configuration
	Node        NodeConfig        `yaml:"node"`
	API         APIConfig         `yaml:"api"`
	P2P         P2PConfig         `yaml:"p2p"`
	Consensus   ConsensusConfig   `yaml:"consensus"`
	Scheduler   SchedulerConfig   `yaml:"scheduler"`
	Storage     StorageConfig     `yaml:"storage"`
	Security    SecurityConfig    `yaml:"security"`
	Web         WebConfig         `yaml:"web"`
	Metrics     MetricsConfig     `yaml:"metrics"`
	Logging     LoggingConfig     `yaml:"logging"`
	Sync        SyncConfig        `yaml:"sync"`
	Replication ReplicationConfig `yaml:"replication"`
	Distributed DistributedConfig `yaml:"distributed"`
}

// NodeConfig holds node-specific configuration
type NodeConfig struct {
	ID          string            `yaml:"id"`
	Name        string            `yaml:"name"`
	Region      string            `yaml:"region"`
	Zone        string            `yaml:"zone"`
	Environment string            `yaml:"environment"`
	Tags        map[string]string `yaml:"tags"`
}

// APIConfig holds API server configuration
type APIConfig struct {
	Listen      string        `yaml:"listen"`
	TLS         TLSConfig     `yaml:"tls"`
	Cors        CorsConfig    `yaml:"cors"`
	RateLimit   RateLimitConfig `yaml:"rate_limit"`
	Timeout     time.Duration `yaml:"timeout"`
	MaxBodySize int64         `yaml:"max_body_size"`
}

// P2PConfig holds P2P networking configuration
type P2PConfig struct {
	Listen        string   `yaml:"listen"`
	Bootstrap     []string `yaml:"bootstrap"`
	PrivateKey    string   `yaml:"private_key"`
	EnableDHT     bool     `yaml:"enable_dht"`
	EnablePubSub  bool     `yaml:"enable_pubsub"`
	ConnMgrLow    int      `yaml:"conn_mgr_low"`
	ConnMgrHigh   int      `yaml:"conn_mgr_high"`
	ConnMgrGrace  string   `yaml:"conn_mgr_grace"`
	DialTimeout   time.Duration `yaml:"dial_timeout"`
	MaxStreams    int      `yaml:"max_streams"`
}

// ConsensusConfig holds consensus engine configuration
type ConsensusConfig struct {
	DataDir          string        `yaml:"data_dir"`
	BindAddr         string        `yaml:"bind_addr"`
	AdvertiseAddr    string        `yaml:"advertise_addr"`
	Bootstrap        bool          `yaml:"bootstrap"`
	LogLevel         string        `yaml:"log_level"`
	HeartbeatTimeout time.Duration `yaml:"heartbeat_timeout"`
	ElectionTimeout  time.Duration `yaml:"election_timeout"`
	CommitTimeout    time.Duration `yaml:"commit_timeout"`
	MaxAppendEntries int           `yaml:"max_append_entries"`
	SnapshotInterval time.Duration `yaml:"snapshot_interval"`
	SnapshotThreshold uint64       `yaml:"snapshot_threshold"`
}

// SchedulerConfig holds scheduler configuration
type SchedulerConfig struct {
	Algorithm        string        `yaml:"algorithm"`
	LoadBalancing    string        `yaml:"load_balancing"`
	HealthCheckInterval time.Duration `yaml:"health_check_interval"`
	MaxRetries       int           `yaml:"max_retries"`
	RetryDelay       time.Duration `yaml:"retry_delay"`
	QueueSize        int           `yaml:"queue_size"`
	WorkerCount      int           `yaml:"worker_count"`
}

// StorageConfig holds storage configuration
type StorageConfig struct {
	DataDir     string `yaml:"data_dir"`
	ModelDir    string `yaml:"model_dir"`
	CacheDir    string `yaml:"cache_dir"`
	MaxDiskSize int64  `yaml:"max_disk_size"`
	CleanupAge  time.Duration `yaml:"cleanup_age"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	TLS         TLSConfig         `yaml:"tls"`
	Auth        AuthConfig        `yaml:"auth"`
	Encryption  EncryptionConfig  `yaml:"encryption"`
	Firewall    FirewallConfig    `yaml:"firewall"`
	Audit       AuditConfig       `yaml:"audit"`
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	Enabled    bool   `yaml:"enabled"`
	CertFile   string `yaml:"cert_file"`
	KeyFile    string `yaml:"key_file"`
	CAFile     string `yaml:"ca_file"`
	MinVersion string `yaml:"min_version"`
	CipherSuites []string `yaml:"cipher_suites"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled     bool          `yaml:"enabled"`
	Method      string        `yaml:"method"` // jwt, oauth, x509
	TokenExpiry time.Duration `yaml:"token_expiry"`
	SecretKey   string        `yaml:"secret_key"`
	Issuer      string        `yaml:"issuer"`
	Audience    string        `yaml:"audience"`
}

// EncryptionConfig holds encryption configuration
type EncryptionConfig struct {
	Algorithm string `yaml:"algorithm"`
	KeySize   int    `yaml:"key_size"`
	KeyFile   string `yaml:"key_file"`
}

// FirewallConfig holds firewall configuration
type FirewallConfig struct {
	Enabled    bool     `yaml:"enabled"`
	AllowedIPs []string `yaml:"allowed_ips"`
	BlockedIPs []string `yaml:"blocked_ips"`
	Rules      []FirewallRule `yaml:"rules"`
}

// FirewallRule represents a firewall rule
type FirewallRule struct {
	Protocol string `yaml:"protocol"`
	Port     int    `yaml:"port"`
	Action   string `yaml:"action"`
	Source   string `yaml:"source"`
}

// AuditConfig holds audit configuration
type AuditConfig struct {
	Enabled bool   `yaml:"enabled"`
	LogFile string `yaml:"log_file"`
	Format  string `yaml:"format"`
}

// CorsConfig holds CORS configuration
type CorsConfig struct {
	Enabled          bool     `yaml:"enabled"`
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	ExposedHeaders   []string `yaml:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAge           int      `yaml:"max_age"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled bool  `yaml:"enabled"`
	RPS     int   `yaml:"rps"`
	Burst   int   `yaml:"burst"`
	Window  time.Duration `yaml:"window"`
}

// WebConfig holds web interface configuration
type WebConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Listen     string `yaml:"listen"`
	StaticDir  string `yaml:"static_dir"`
	TemplateDir string `yaml:"template_dir"`
	TLS        TLSConfig `yaml:"tls"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Listen     string `yaml:"listen"`
	Path       string `yaml:"path"`
	Namespace  string `yaml:"namespace"`
	Subsystem  string `yaml:"subsystem"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	Output     string `yaml:"output"`
	File       string `yaml:"file"`
	MaxSize    int    `yaml:"max_size"`
	MaxAge     int    `yaml:"max_age"`
	MaxBackups int    `yaml:"max_backups"`
	Compress   bool   `yaml:"compress"`
}

// SyncConfig holds model synchronization configuration
type SyncConfig struct {
	DeltaDir       string        `yaml:"delta_dir"`
	CASDir         string        `yaml:"cas_dir"`
	WorkerCount    int           `yaml:"worker_count"`
	SyncInterval   time.Duration `yaml:"sync_interval"`
	ChunkSize      int64         `yaml:"chunk_size"`
	MaxRetries     int           `yaml:"max_retries"`
	RetryDelay     time.Duration `yaml:"retry_delay"`
}

// ReplicationConfig holds model replication configuration
type ReplicationConfig struct {
	WorkerCount                int           `yaml:"worker_count"`
	DefaultMinReplicas         int           `yaml:"default_min_replicas"`
	DefaultMaxReplicas         int           `yaml:"default_max_replicas"`
	DefaultReplicationFactor   int           `yaml:"default_replication_factor"`
	DefaultSyncInterval        time.Duration `yaml:"default_sync_interval"`
	PolicyEnforcementInterval  time.Duration `yaml:"policy_enforcement_interval"`
	HealthCheckInterval        time.Duration `yaml:"health_check_interval"`
	HealthCheckTimeout         time.Duration `yaml:"health_check_timeout"`
}

// DistributedConfig holds distributed model management configuration
type DistributedConfig struct {
	Storage     *StorageConfig     `yaml:"storage"`
	Sync        *SyncConfig        `yaml:"sync"`
	Replication *ReplicationConfig `yaml:"replication"`
	CASDir      string             `yaml:"cas_dir"`
	DeltaDir    string             `yaml:"delta_dir"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	// Create storage config first
	storageConfig := StorageConfig{
		DataDir:     "./data",
		ModelDir:    "./models",
		CacheDir:    "./cache",
		MaxDiskSize: 100 * 1024 * 1024 * 1024, // 100GB
		CleanupAge:  7 * 24 * time.Hour,       // 7 days
	}
	
	// Create sync config
	syncConfig := SyncConfig{
		DeltaDir:     "./data/deltas",
		CASDir:       "./data/cas",
		WorkerCount:  3,
		SyncInterval: 5 * time.Minute,
		ChunkSize:    1024 * 1024, // 1MB
		MaxRetries:   3,
		RetryDelay:   time.Second,
	}
	
	// Create replication config
	replicationConfig := ReplicationConfig{
		WorkerCount:                3,
		DefaultMinReplicas:         1,
		DefaultMaxReplicas:         3,
		DefaultReplicationFactor:   2,
		DefaultSyncInterval:        10 * time.Minute,
		PolicyEnforcementInterval:  30 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckTimeout:         10 * time.Second,
	}
	
	return &Config{
		Node: NodeConfig{
			ID:          "",
			Name:        "ollama-node",
			Region:      "us-west-2",
			Zone:        "us-west-2a",
			Environment: "production",
			Tags:        make(map[string]string),
		},
		API: APIConfig{
			Listen:      "0.0.0.0:11434",
			Timeout:     30 * time.Second,
			MaxBodySize: 32 * 1024 * 1024, // 32MB
			TLS: TLSConfig{
				Enabled:    false,
				MinVersion: "1.2",
			},
			Cors: CorsConfig{
				Enabled:          true,
				AllowedOrigins:   []string{"http://localhost:8080", "https://localhost:8080"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Requested-With"},
				AllowCredentials: true,
				MaxAge:           3600,
			},
			RateLimit: RateLimitConfig{
				Enabled: true,
				RPS:     1000,
				Burst:   2000,
				Window:  time.Minute,
			},
		},
		P2P: P2PConfig{
			Listen:        "/ip4/0.0.0.0/tcp/4001",
			Bootstrap:     []string{},
			EnableDHT:     true,
			EnablePubSub:  true,
			ConnMgrLow:    50,
			ConnMgrHigh:   200,
			ConnMgrGrace:  "30s",
			DialTimeout:   30 * time.Second,
			MaxStreams:    1000,
		},
		Consensus: ConsensusConfig{
			DataDir:           "./data/consensus",
			BindAddr:          "0.0.0.0:7000",
			AdvertiseAddr:     "",
			Bootstrap:         false,
			LogLevel:          "INFO",
			HeartbeatTimeout:  1 * time.Second,
			ElectionTimeout:   1 * time.Second,
			CommitTimeout:     50 * time.Millisecond,
			MaxAppendEntries:  64,
			SnapshotInterval:  120 * time.Second,
			SnapshotThreshold: 8192,
		},
		Scheduler: SchedulerConfig{
			Algorithm:           "round_robin",
			LoadBalancing:       "least_connections",
			HealthCheckInterval: 30 * time.Second,
			MaxRetries:          3,
			RetryDelay:          1 * time.Second,
			QueueSize:           10000,
			WorkerCount:         10,
		},
		Storage: storageConfig,
		Security: SecurityConfig{
			TLS: TLSConfig{
				Enabled:    true,
				MinVersion: "1.3",
			},
			Auth: AuthConfig{
				Enabled:     true,
				Method:      "jwt",
				TokenExpiry: 24 * time.Hour,
			},
			Encryption: EncryptionConfig{
				Algorithm: "AES-256-GCM",
				KeySize:   256,
			},
			Firewall: FirewallConfig{
				Enabled: true,
				Rules:   []FirewallRule{},
			},
			Audit: AuditConfig{
				Enabled: true,
				LogFile: "./logs/audit.log",
				Format:  "json",
			},
		},
		Web: WebConfig{
			Enabled:   true,
			Listen:    "0.0.0.0:8080",
			StaticDir: "./web/static",
			TLS: TLSConfig{
				Enabled: false,
			},
		},
		Metrics: MetricsConfig{
			Enabled:   true,
			Listen:    "0.0.0.0:9090",
			Path:      "/metrics",
			Namespace: "ollama",
			Subsystem: "distributed",
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100,
			MaxAge:     30,
			MaxBackups: 10,
			Compress:   true,
		},
		Sync:        syncConfig,
		Replication: replicationConfig,
		Distributed: DistributedConfig{
			Storage:     &storageConfig,
			Sync:        &syncConfig,
			Replication: &replicationConfig,
			CASDir:      "./data/cas",
			DeltaDir:    "./data/deltas",
		},
	}
}

// Load loads configuration from file
func Load(configFile string) (*Config, error) {
	config := DefaultConfig()
	
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		// Look for config in standard locations
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
		viper.AddConfigPath("$HOME/.ollama-distributed")
		viper.AddConfigPath("/etc/ollama-distributed")
	}
	
	// Environment variables
	viper.SetEnvPrefix("OLLAMA")
	viper.AutomaticEnv()
	
	// Read configuration
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}
	
	// Unmarshal into config struct
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	// Validate and set defaults
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate directories exist or can be created
	dirs := []string{
		c.Storage.DataDir,
		c.Storage.ModelDir,
		c.Storage.CacheDir,
		c.Consensus.DataDir,
		c.Sync.DeltaDir,
		c.Sync.CASDir,
		c.Distributed.CASDir,
		c.Distributed.DeltaDir,
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	
	// Validate log directory
	if c.Logging.Output == "file" && c.Logging.File != "" {
		logDir := filepath.Dir(c.Logging.File)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory %s: %w", logDir, err)
		}
	}
	
	// Validate TLS certificates if enabled
	if c.Security.TLS.Enabled {
		if c.Security.TLS.CertFile == "" || c.Security.TLS.KeyFile == "" {
			return fmt.Errorf("TLS enabled but cert_file or key_file not specified")
		}
	}
	
	return nil
}

// Save saves the configuration to a file
func (c *Config) Save(filename string) error {
	viper.Set("config", c)
	return viper.WriteConfigAs(filename)
}
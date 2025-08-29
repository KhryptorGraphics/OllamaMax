// 02-configuration/configuration-manager.go
// Comprehensive configuration management examples for Ollama Distributed Training
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// TrainingConfigManager handles configuration management for training scenarios
type TrainingConfigManager struct {
	BaseDir     string
	ProfilesDir string
	DataDir     string
}

// ConfigProfile represents a complete configuration profile
type ConfigProfile struct {
	Name        string                 `yaml:"name" json:"name"`
	Description string                 `yaml:"description" json:"description"`
	Environment string                 `yaml:"environment" json:"environment"` // development, testing, production
	API         APIConfig              `yaml:"api" json:"api"`
	P2P         P2PConfig              `yaml:"p2p" json:"p2p"`
	Web         WebConfig              `yaml:"web" json:"web"`
	Storage     StorageConfig          `yaml:"storage" json:"storage"`
	Logging     LoggingConfig          `yaml:"logging" json:"logging"`
	Performance PerformanceConfig      `yaml:"performance" json:"performance"`
	Consensus   ConsensusConfig        `yaml:"consensus" json:"consensus"`
	Scheduler   SchedulerConfig        `yaml:"scheduler" json:"scheduler"`
	Security    SecurityConfig         `yaml:"security" json:"security"`
	Metadata    map[string]interface{} `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// APIConfig defines API server configuration
type APIConfig struct {
	Listen      string            `yaml:"listen" json:"listen"`
	Debug       bool              `yaml:"debug" json:"debug"`
	Timeout     time.Duration     `yaml:"timeout" json:"timeout"`
	TLS         TLSConfig         `yaml:"tls" json:"tls"`
	CORS        CORSConfig        `yaml:"cors" json:"cors"`
	RateLimit   RateLimitConfig   `yaml:"rate_limit" json:"rate_limit"`
	Middleware  []string          `yaml:"middleware" json:"middleware"`
	Headers     map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
}

// P2PConfig defines P2P networking configuration
type P2PConfig struct {
	Listen      string            `yaml:"listen" json:"listen"`
	Bootstrap   []string          `yaml:"bootstrap" json:"bootstrap"`
	Protocol    string            `yaml:"protocol" json:"protocol"`
	Discovery   DiscoveryConfig   `yaml:"discovery" json:"discovery"`
	Security    P2PSecurityConfig `yaml:"security" json:"security"`
	Limits      P2PLimitsConfig   `yaml:"limits" json:"limits"`
}

// WebConfig defines web interface configuration
type WebConfig struct {
	Listen     string `yaml:"listen" json:"listen"`
	EnableAuth bool   `yaml:"enable_auth" json:"enable_auth"`
	StaticDir  string `yaml:"static_dir" json:"static_dir"`
	Theme      string `yaml:"theme" json:"theme"`
	Title      string `yaml:"title" json:"title"`
}

// StorageConfig defines storage configuration
type StorageConfig struct {
	DataDir    string `yaml:"data_dir" json:"data_dir"`
	ModelsDir  string `yaml:"models_dir" json:"models_dir"`
	CacheDir   string `yaml:"cache_dir" json:"cache_dir"`
	TempDir    string `yaml:"temp_dir" json:"temp_dir"`
	MaxSize    string `yaml:"max_size" json:"max_size"`
	Cleanup    bool   `yaml:"cleanup" json:"cleanup"`
}

// LoggingConfig defines logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level" json:"level"`
	Output     string `yaml:"output" json:"output"` // console, file, both
	File       string `yaml:"file" json:"file"`
	MaxSize    string `yaml:"max_size" json:"max_size"`
	MaxBackups int    `yaml:"max_backups" json:"max_backups"`
	MaxAge     int    `yaml:"max_age" json:"max_age"`
	Compress   bool   `yaml:"compress" json:"compress"`
}

// PerformanceConfig defines performance monitoring configuration
type PerformanceConfig struct {
	MonitoringEnabled    bool          `yaml:"monitoring_enabled" json:"monitoring_enabled"`
	MetricsInterval      int           `yaml:"metrics_interval" json:"metrics_interval"`
	OptimizationEnabled  bool          `yaml:"optimization_enabled" json:"optimization_enabled"`
	ProfilerEnabled      bool          `yaml:"profiler_enabled" json:"profiler_enabled"`
	MemoryLimit          string        `yaml:"memory_limit" json:"memory_limit"`
	CPULimit             string        `yaml:"cpu_limit" json:"cpu_limit"`
	GCTargetPercentage   int           `yaml:"gc_target_percentage" json:"gc_target_percentage"`
}

// ConsensusConfig defines consensus algorithm configuration
type ConsensusConfig struct {
	DataDir         string        `yaml:"data_dir" json:"data_dir"`
	Bootstrap       bool          `yaml:"bootstrap" json:"bootstrap"`
	BindAddr        string        `yaml:"bind_addr" json:"bind_addr"`
	ElectionTimeout time.Duration `yaml:"election_timeout" json:"election_timeout"`
	HeartbeatTime   time.Duration `yaml:"heartbeat_time" json:"heartbeat_time"`
	MaxLogEntries   int           `yaml:"max_log_entries" json:"max_log_entries"`
}

// SchedulerConfig defines task scheduler configuration
type SchedulerConfig struct {
	Algorithm     string            `yaml:"algorithm" json:"algorithm"`      // round_robin, cpu_aware, memory_aware
	LoadBalancing string            `yaml:"load_balancing" json:"load_balancing"`
	WorkerCount   int               `yaml:"worker_count" json:"worker_count"`
	QueueSize     int               `yaml:"queue_size" json:"queue_size"`
	Policies      map[string]string `yaml:"policies,omitempty" json:"policies,omitempty"`
}

// SecurityConfig defines security settings
type SecurityConfig struct {
	EnableAuth   bool          `yaml:"enable_auth" json:"enable_auth"`
	JWT          JWTConfig     `yaml:"jwt" json:"jwt"`
	Encryption   EncryptionConfig `yaml:"encryption" json:"encryption"`
	Certificates CertConfig    `yaml:"certificates" json:"certificates"`
}

// Supporting configuration structs
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`
	CertFile string `yaml:"cert_file" json:"cert_file"`
	KeyFile  string `yaml:"key_file" json:"key_file"`
}

type CORSConfig struct {
	Enabled bool     `yaml:"enabled" json:"enabled"`
	Origins []string `yaml:"origins" json:"origins"`
	Methods []string `yaml:"methods" json:"methods"`
	Headers []string `yaml:"headers" json:"headers"`
}

type RateLimitConfig struct {
	Enabled     bool `yaml:"enabled" json:"enabled"`
	RequestsPerMinute int `yaml:"requests_per_minute" json:"requests_per_minute"`
	BurstSize   int  `yaml:"burst_size" json:"burst_size"`
}

type DiscoveryConfig struct {
	Method   string   `yaml:"method" json:"method"` // mdns, bootstrap, dht
	Interval int      `yaml:"interval" json:"interval"`
	Nodes    []string `yaml:"nodes,omitempty" json:"nodes,omitempty"`
}

type P2PSecurityConfig struct {
	EnableTLS    bool   `yaml:"enable_tls" json:"enable_tls"`
	PrivateKey   string `yaml:"private_key" json:"private_key"`
	AllowedPeers []string `yaml:"allowed_peers,omitempty" json:"allowed_peers,omitempty"`
}

type P2PLimitsConfig struct {
	MaxConnections int `yaml:"max_connections" json:"max_connections"`
	MaxStreams     int `yaml:"max_streams" json:"max_streams"`
	MaxMessageSize int `yaml:"max_message_size" json:"max_message_size"`
}

type JWTConfig struct {
	Secret     string        `yaml:"secret" json:"secret"`
	Expiration time.Duration `yaml:"expiration" json:"expiration"`
	Algorithm  string        `yaml:"algorithm" json:"algorithm"`
}

type EncryptionConfig struct {
	Method string `yaml:"method" json:"method"`
	KeyFile string `yaml:"key_file" json:"key_file"`
}

type CertConfig struct {
	CAFile   string `yaml:"ca_file" json:"ca_file"`
	CertFile string `yaml:"cert_file" json:"cert_file"`
	KeyFile  string `yaml:"key_file" json:"key_file"`
}

// NewTrainingConfigManager creates a new configuration manager
func NewTrainingConfigManager(baseDir string) *TrainingConfigManager {
	return &TrainingConfigManager{
		BaseDir:     baseDir,
		ProfilesDir: filepath.Join(baseDir, "profiles"),
		DataDir:     filepath.Join(baseDir, "data"),
	}
}

// Initialize sets up the configuration management environment
func (tcm *TrainingConfigManager) Initialize() error {
	// Create directories
	dirs := []string{
		tcm.BaseDir,
		tcm.ProfilesDir,
		tcm.DataDir,
		filepath.Join(tcm.DataDir, "logs"),
		filepath.Join(tcm.DataDir, "models"),
		filepath.Join(tcm.DataDir, "cache"),
		filepath.Join(tcm.DataDir, "temp"),
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	
	log.Printf("✅ Configuration environment initialized in %s", tcm.BaseDir)
	return nil
}

// CreateDevelopmentProfile creates a comprehensive development configuration
func (tcm *TrainingConfigManager) CreateDevelopmentProfile() error {
	profile := ConfigProfile{
		Name:        "development",
		Description: "Development environment configuration for training",
		Environment: "development",
		API: APIConfig{
			Listen:  "127.0.0.1:8080",
			Debug:   true,
			Timeout: 30 * time.Second,
			TLS: TLSConfig{
				Enabled: false,
			},
			CORS: CORSConfig{
				Enabled: true,
				Origins: []string{"*"},
				Methods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				Headers: []string{"*"},
			},
			RateLimit: RateLimitConfig{
				Enabled:           false,
				RequestsPerMinute: 1000,
				BurstSize:         50,
			},
			Middleware: []string{"logger", "cors", "recovery"},
			Headers: map[string]string{
				"X-API-Version": "v1",
				"X-Environment": "development",
			},
		},
		P2P: P2PConfig{
			Listen:    "127.0.0.1:4001",
			Bootstrap: []string{},
			Protocol:  "tcp",
			Discovery: DiscoveryConfig{
				Method:   "mdns",
				Interval: 30,
			},
			Security: P2PSecurityConfig{
				EnableTLS: false,
			},
			Limits: P2PLimitsConfig{
				MaxConnections: 50,
				MaxStreams:     100,
				MaxMessageSize: 1048576, // 1MB
			},
		},
		Web: WebConfig{
			Listen:     "127.0.0.1:8081",
			EnableAuth: false,
			StaticDir:  "./web/static",
			Theme:      "default",
			Title:      "Ollama Distributed - Development",
		},
		Storage: StorageConfig{
			DataDir:   "./dev-data",
			ModelsDir: "./dev-data/models",
			CacheDir:  "./dev-data/cache",
			TempDir:   "./dev-data/temp",
			MaxSize:   "10GB",
			Cleanup:   true,
		},
		Logging: LoggingConfig{
			Level:      "debug",
			Output:     "both",
			File:       "./dev-data/logs/ollama-distributed.log",
			MaxSize:    "100MB",
			MaxBackups: 5,
			MaxAge:     7,
			Compress:   true,
		},
		Performance: PerformanceConfig{
			MonitoringEnabled:   true,
			MetricsInterval:     5,
			OptimizationEnabled: false,
			ProfilerEnabled:     true,
			MemoryLimit:        "2GB",
			CPULimit:          "2.0",
			GCTargetPercentage: 100,
		},
		Consensus: ConsensusConfig{
			DataDir:         "./dev-data/consensus",
			Bootstrap:       true,
			BindAddr:        "127.0.0.1:7000",
			ElectionTimeout: 1 * time.Second,
			HeartbeatTime:   100 * time.Millisecond,
			MaxLogEntries:   10000,
		},
		Scheduler: SchedulerConfig{
			Algorithm:     "round_robin",
			LoadBalancing: "cpu_aware",
			WorkerCount:   4,
			QueueSize:     100,
			Policies: map[string]string{
				"retry_policy": "exponential_backoff",
				"timeout":      "30s",
			},
		},
		Security: SecurityConfig{
			EnableAuth: false,
			JWT: JWTConfig{
				Secret:     "dev-secret-key",
				Expiration: 24 * time.Hour,
				Algorithm:  "HS256",
			},
			Encryption: EncryptionConfig{
				Method: "none",
			},
		},
		Metadata: map[string]interface{}{
			"created":     time.Now(),
			"created_by":  "training-manager",
			"version":     "1.0.0",
			"tags":        []string{"development", "training", "local"},
			"description": "Comprehensive development configuration for Ollama Distributed training",
		},
	}
	
	return tcm.SaveProfile("development", &profile)
}

// CreateTestingProfile creates a configuration optimized for testing
func (tcm *TrainingConfigManager) CreateTestingProfile() error {
	profile := ConfigProfile{
		Name:        "testing",
		Description: "Testing environment configuration",
		Environment: "testing",
		API: APIConfig{
			Listen:  "127.0.0.1:9080",
			Debug:   false,
			Timeout: 15 * time.Second,
			CORS: CORSConfig{
				Enabled: true,
				Origins: []string{"http://localhost:*", "http://127.0.0.1:*"},
			},
			RateLimit: RateLimitConfig{
				Enabled:           true,
				RequestsPerMinute: 100,
				BurstSize:         10,
			},
		},
		P2P: P2PConfig{
			Listen:    "127.0.0.1:4002",
			Bootstrap: []string{},
		},
		Web: WebConfig{
			Listen:     "127.0.0.1:9081",
			EnableAuth: true,
			Title:      "Ollama Distributed - Testing",
		},
		Storage: StorageConfig{
			DataDir:   "./test-data",
			ModelsDir: "./test-data/models",
			MaxSize:   "1GB",
			Cleanup:   true,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Output: "file",
			File:   "./test-data/logs/test.log",
		},
		Performance: PerformanceConfig{
			MonitoringEnabled:   true,
			MetricsInterval:     10,
			OptimizationEnabled: true,
		},
		Metadata: map[string]interface{}{
			"created":    time.Now(),
			"created_by": "training-manager",
			"purpose":    "automated testing",
		},
	}
	
	return tcm.SaveProfile("testing", &profile)
}

// CreateProductionProfile creates a production-ready configuration template
func (tcm *TrainingConfigManager) CreateProductionProfile() error {
	profile := ConfigProfile{
		Name:        "production",
		Description: "Production environment configuration template",
		Environment: "production",
		API: APIConfig{
			Listen:  "0.0.0.0:8080",
			Debug:   false,
			Timeout: 60 * time.Second,
			TLS: TLSConfig{
				Enabled:  true,
				CertFile: "/etc/ssl/certs/ollama-distributed.crt",
				KeyFile:  "/etc/ssl/private/ollama-distributed.key",
			},
			CORS: CORSConfig{
				Enabled: true,
				Origins: []string{"https://yourdomain.com"},
			},
			RateLimit: RateLimitConfig{
				Enabled:           true,
				RequestsPerMinute: 1000,
				BurstSize:         100,
			},
		},
		P2P: P2PConfig{
			Listen:    "0.0.0.0:4001",
			Bootstrap: []string{
				// "peer1.example.com:4001",
				// "peer2.example.com:4001",
			},
			Security: P2PSecurityConfig{
				EnableTLS: true,
			},
		},
		Web: WebConfig{
			Listen:     "0.0.0.0:8081",
			EnableAuth: true,
			Title:      "Ollama Distributed - Production",
		},
		Storage: StorageConfig{
			DataDir:   "/var/lib/ollama-distributed",
			ModelsDir: "/var/lib/ollama-distributed/models",
			MaxSize:   "100GB",
			Cleanup:   false,
		},
		Logging: LoggingConfig{
			Level:      "warn",
			Output:     "file",
			File:       "/var/log/ollama-distributed/service.log",
			MaxSize:    "500MB",
			MaxBackups: 10,
			MaxAge:     30,
			Compress:   true,
		},
		Performance: PerformanceConfig{
			MonitoringEnabled:   true,
			MetricsInterval:     30,
			OptimizationEnabled: true,
			ProfilerEnabled:     false,
			MemoryLimit:        "8GB",
			CPULimit:          "4.0",
		},
		Security: SecurityConfig{
			EnableAuth: true,
			JWT: JWTConfig{
				Secret:     "${JWT_SECRET}", // Use environment variable
				Expiration: 1 * time.Hour,
				Algorithm:  "HS256",
			},
		},
		Metadata: map[string]interface{}{
			"created":     time.Now(),
			"created_by":  "training-manager",
			"environment": "production",
			"notes":       "Template configuration - customize before deployment",
		},
	}
	
	return tcm.SaveProfile("production", &profile)
}

// SaveProfile saves a configuration profile to disk
func (tcm *TrainingConfigManager) SaveProfile(name string, profile *ConfigProfile) error {
	profilePath := filepath.Join(tcm.ProfilesDir, name+".yaml")
	
	data, err := yaml.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}
	
	// Add header comment
	header := fmt.Sprintf("# Ollama Distributed Configuration Profile: %s\n# Generated: %s\n# Environment: %s\n\n", 
		profile.Name, time.Now().Format(time.RFC3339), profile.Environment)
	
	fullData := append([]byte(header), data...)
	
	if err := ioutil.WriteFile(profilePath, fullData, 0644); err != nil {
		return fmt.Errorf("failed to write profile to %s: %w", profilePath, err)
	}
	
	log.Printf("✅ Saved configuration profile '%s' to %s", name, profilePath)
	return nil
}

// LoadProfile loads a configuration profile from disk
func (tcm *TrainingConfigManager) LoadProfile(name string) (*ConfigProfile, error) {
	profilePath := filepath.Join(tcm.ProfilesDir, name+".yaml")
	
	data, err := ioutil.ReadFile(profilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile %s: %w", profilePath, err)
	}
	
	var profile ConfigProfile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
	}
	
	return &profile, nil
}

// ListProfiles lists all available configuration profiles
func (tcm *TrainingConfigManager) ListProfiles() ([]string, error) {
	files, err := ioutil.ReadDir(tcm.ProfilesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read profiles directory: %w", err)
	}
	
	var profiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
			name := strings.TrimSuffix(file.Name(), ".yaml")
			profiles = append(profiles, name)
		}
	}
	
	return profiles, nil
}

// ValidateProfile validates a configuration profile for correctness
func (tcm *TrainingConfigManager) ValidateProfile(profile *ConfigProfile) []string {
	var warnings []string
	
	// Validate API configuration
	if profile.API.Listen == "" {
		warnings = append(warnings, "API listen address is empty")
	}
	
	// Validate P2P configuration
	if profile.P2P.Listen == "" {
		warnings = append(warnings, "P2P listen address is empty")
	}
	
	// Validate storage paths
	if profile.Storage.DataDir == "" {
		warnings = append(warnings, "Storage data directory is empty")
	}
	
	// Validate logging configuration
	if profile.Logging.Output == "file" && profile.Logging.File == "" {
		warnings = append(warnings, "Logging output is 'file' but no file path specified")
	}
	
	// Validate security settings for production
	if profile.Environment == "production" {
		if !profile.Security.EnableAuth {
			warnings = append(warnings, "Authentication disabled in production environment")
		}
		if !profile.API.TLS.Enabled {
			warnings = append(warnings, "TLS disabled in production environment")
		}
		if profile.API.Debug {
			warnings = append(warnings, "Debug mode enabled in production environment")
		}
	}
	
	return warnings
}

// ExportProfileAsJSON exports a profile in JSON format
func (tcm *TrainingConfigManager) ExportProfileAsJSON(name string) (string, error) {
	profile, err := tcm.LoadProfile(name)
	if err != nil {
		return "", err
	}
	
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal profile to JSON: %w", err)
	}
	
	return string(data), nil
}

// CreateEnvironmentVariablesFile creates a .env file for a profile
func (tcm *TrainingConfigManager) CreateEnvironmentVariablesFile(profileName string) error {
	profile, err := tcm.LoadProfile(profileName)
	if err != nil {
		return err
	}
	
	envFile := filepath.Join(tcm.ProfilesDir, profileName+".env")
	
	var envVars strings.Builder
	envVars.WriteString(fmt.Sprintf("# Environment variables for %s profile\n", profileName))
	envVars.WriteString(fmt.Sprintf("# Generated: %s\n\n", time.Now().Format(time.RFC3339)))
	
	envVars.WriteString(fmt.Sprintf("OLLAMA_DISTRIBUTED_PROFILE=%s\n", profileName))
	envVars.WriteString(fmt.Sprintf("OLLAMA_DISTRIBUTED_ENVIRONMENT=%s\n", profile.Environment))
	envVars.WriteString(fmt.Sprintf("OLLAMA_DISTRIBUTED_API_LISTEN=%s\n", profile.API.Listen))
	envVars.WriteString(fmt.Sprintf("OLLAMA_DISTRIBUTED_P2P_LISTEN=%s\n", profile.P2P.Listen))
	envVars.WriteString(fmt.Sprintf("OLLAMA_DISTRIBUTED_WEB_LISTEN=%s\n", profile.Web.Listen))
	envVars.WriteString(fmt.Sprintf("OLLAMA_DISTRIBUTED_DATA_DIR=%s\n", profile.Storage.DataDir))
	envVars.WriteString(fmt.Sprintf("OLLAMA_DISTRIBUTED_LOG_LEVEL=%s\n", profile.Logging.Level))
	
	if profile.Security.EnableAuth && profile.Security.JWT.Secret == "${JWT_SECRET}" {
		envVars.WriteString("JWT_SECRET=your-secure-secret-key-here\n")
	}
	
	if err := ioutil.WriteFile(envFile, []byte(envVars.String()), 0644); err != nil {
		return fmt.Errorf("failed to write environment file: %w", err)
	}
	
	log.Printf("✅ Created environment variables file: %s", envFile)
	return nil
}

// Example usage and demonstration
func main() {
	// Initialize configuration manager
	homeDir, _ := os.UserHomeDir()
	baseDir := filepath.Join(homeDir, ".ollama-distributed")
	
	tcm := NewTrainingConfigManager(baseDir)
	
	// Initialize environment
	if err := tcm.Initialize(); err != nil {
		log.Fatalf("Failed to initialize configuration environment: %v", err)
	}
	
	// Create all training profiles
	log.Println("🔧 Creating training configuration profiles...")
	
	if err := tcm.CreateDevelopmentProfile(); err != nil {
		log.Fatalf("Failed to create development profile: %v", err)
	}
	
	if err := tcm.CreateTestingProfile(); err != nil {
		log.Fatalf("Failed to create testing profile: %v", err)
	}
	
	if err := tcm.CreateProductionProfile(); err != nil {
		log.Fatalf("Failed to create production profile: %v", err)
	}
	
	// List all profiles
	profiles, err := tcm.ListProfiles()
	if err != nil {
		log.Fatalf("Failed to list profiles: %v", err)
	}
	
	log.Println("\n📋 Available Configuration Profiles:")
	for _, profile := range profiles {
		log.Printf("   - %s", profile)
		
		// Validate each profile
		if p, err := tcm.LoadProfile(profile); err == nil {
			warnings := tcm.ValidateProfile(p)
			if len(warnings) > 0 {
				log.Printf("     ⚠️  Warnings: %d", len(warnings))
				for _, warning := range warnings {
					log.Printf("       - %s", warning)
				}
			} else {
				log.Printf("     ✅ No validation issues")
			}
		}
		
		// Create environment variables file
		tcm.CreateEnvironmentVariablesFile(profile)
	}
	
	// Demonstrate JSON export
	log.Println("\n📤 JSON Export Example:")
	jsonData, err := tcm.ExportProfileAsJSON("development")
	if err == nil {
		// Just show first 200 characters
		if len(jsonData) > 200 {
			log.Printf("   Development profile (JSON): %s...", jsonData[:200])
		} else {
			log.Printf("   Development profile (JSON): %s", jsonData)
		}
	}
	
	log.Println("\n🎉 Configuration management setup complete!")
	log.Printf("   Base directory: %s", baseDir)
	log.Printf("   Profiles directory: %s", tcm.ProfilesDir)
	log.Println("\n💡 Usage examples:")
	log.Println("   Load profile: source ~/.ollama-distributed/profiles/development.env")
	log.Println("   Start with profile: ollama-distributed start --config ~/.ollama-distributed/profiles/development.yaml")
}
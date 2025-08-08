package config

import (
	"crypto/rand"
	"os"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"gopkg.in/yaml.v2"
)

// TURNServerConfig holds TURN server configuration
type TURNServerConfig struct {
	Address   string `yaml:"address"`
	Port      int    `yaml:"port"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	Realm     string `yaml:"realm"`
	Transport string `yaml:"transport"`
}

// NodeConfig holds all configuration for a P2P node
type NodeConfig struct {
	// Network Settings
	Listen              []string `yaml:"listen"`
	AnnounceAddresses   []string `yaml:"announce_addresses"`
	NoAnnounceAddresses []string `yaml:"no_announce_addresses"`

	// Security
	PrivateKey  string `yaml:"private_key"`
	EnableTLS   bool   `yaml:"enable_tls"`
	EnableNoise bool   `yaml:"enable_noise"`

	// NAT Traversal
	EnableNATService   bool               `yaml:"enable_nat_service"`
	EnableHolePunching bool               `yaml:"enable_hole_punching"`
	EnableAutoRelay    bool               `yaml:"enable_auto_relay"`
	StaticRelays       []string           `yaml:"static_relays"`
	TURNServers        []TURNServerConfig `yaml:"turn_servers"`
	ForceReachability  string             `yaml:"force_reachability"` // public/private

	// DHT Settings
	EnableDHT      bool     `yaml:"enable_dht"`
	DHTMode        string   `yaml:"dht_mode"` // client/server/auto
	BootstrapPeers []string `yaml:"bootstrap_peers"`

	// Connection Management
	ConnMgrLow   int           `yaml:"conn_mgr_low"`
	ConnMgrHigh  int           `yaml:"conn_mgr_high"`
	ConnMgrGrace time.Duration `yaml:"conn_mgr_grace"`

	// Resource Management
	MaxMemory int64   `yaml:"max_memory"`
	MaxCPU    float64 `yaml:"max_cpu"`
	MaxGPU    int     `yaml:"max_gpu"`

	// Ollamacron Specific
	NodeType          string            `yaml:"node_type"` // edge/standard/super
	ModelCapabilities []string          `yaml:"model_capabilities"`
	ResourceTags      map[string]string `yaml:"resource_tags"`

	// Discovery Settings
	RendezvousString string `yaml:"rendezvous_string"`
	AutoDiscovery    bool   `yaml:"auto_discovery"`
}

// DefaultConfig returns a default configuration for a P2P node
func DefaultConfig() *NodeConfig {
	return &NodeConfig{
		Listen: []string{
			"/ip4/0.0.0.0/tcp/0",
			"/ip6/::/tcp/0",
			"/ip4/0.0.0.0/udp/0/quic",
			"/ip6/::/udp/0/quic",
		},
		EnableTLS:          true,
		EnableNoise:        true,
		EnableNATService:   true,
		EnableHolePunching: true,
		EnableAutoRelay:    true,
		EnableDHT:          true,
		DHTMode:            "auto",
		ConnMgrLow:         50,
		ConnMgrHigh:        200,
		ConnMgrGrace:       time.Minute,
		NodeType:           "standard",
		ModelCapabilities:  []string{},
		ResourceTags:       make(map[string]string),
		RendezvousString:   "ollama-distributed",
		AutoDiscovery:      true,
	}
}

// GenerateKey generates a new cryptographic identity for the node
func (c *NodeConfig) GenerateKey() error {
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	if err != nil {
		return err
	}

	keyBytes, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return err
	}

	c.PrivateKey = crypto.ConfigEncodeKey(keyBytes)
	return nil
}

// GetPrivateKey retrieves the private key from configuration
func (c *NodeConfig) GetPrivateKey() (crypto.PrivKey, error) {
	if c.PrivateKey == "" {
		return nil, nil
	}

	keyBytes, err := crypto.ConfigDecodeKey(c.PrivateKey)
	if err != nil {
		return nil, err
	}

	return crypto.UnmarshalPrivateKey(keyBytes)
}

// ParseBootstrapPeers parses bootstrap peer addresses
func (c *NodeConfig) ParseBootstrapPeers() ([]peer.AddrInfo, error) {
	var peers []peer.AddrInfo

	for _, addr := range c.BootstrapPeers {
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			continue
		}

		peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			continue
		}

		peers = append(peers, *peerInfo)
	}

	return peers, nil
}

// ParseStaticRelays parses static relay addresses
func (c *NodeConfig) ParseStaticRelays() ([]peer.AddrInfo, error) {
	var relays []peer.AddrInfo

	for _, addr := range c.StaticRelays {
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			continue
		}

		peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			continue
		}

		relays = append(relays, *peerInfo)
	}

	return relays, nil
}

// DiscoveryConfig interface implementation
func (c *NodeConfig) GetBootstrapPeers() []string {
	return c.BootstrapPeers
}

func (c *NodeConfig) GetRendezvousString() string {
	return c.RendezvousString
}

func (c *NodeConfig) IsAutoDiscoveryEnabled() bool {
	return c.AutoDiscovery
}

// NodeCapabilities represents the capabilities of a P2P node
type NodeCapabilities struct {
	// Compute resources
	CPUCores int        `json:"cpu_cores"`
	Memory   int64      `json:"memory"`
	Storage  int64      `json:"storage"`
	GPUs     []*GPUInfo `json:"gpus"`

	// AI capabilities
	SupportedModels []string `json:"supported_models"`
	ModelFormats    []string `json:"model_formats"`
	Quantizations   []string `json:"quantizations"`

	// Network capabilities
	Bandwidth     int64         `json:"bandwidth"`
	Latency       time.Duration `json:"latency"`
	Reliability   float64       `json:"reliability"`
	PricePerToken float64       `json:"price_per_token"`

	// Node state
	Available  bool      `json:"available"`
	LoadFactor float64   `json:"load_factor"`
	Priority   int       `json:"priority"`
	LastSeen   time.Time `json:"last_seen"`

	// Version information
	Version         string   `json:"version"`
	ProtocolVersion string   `json:"protocol_version"`
	Features        []string `json:"features"`
}

// GPUInfo represents information about a GPU
type GPUInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Memory      int64             `json:"memory"`
	Compute     string            `json:"compute"`
	Available   bool              `json:"available"`
	Utilization float64           `json:"utilization"`
	Properties  map[string]string `json:"properties"`
}

// ResourceMetrics contains real-time resource usage metrics
type ResourceMetrics struct {
	CPUUsage    float64 `json:"cpu_usage" yaml:"cpu_usage"`       // CPU usage percentage (0-100)
	MemoryUsage int64   `json:"memory_usage" yaml:"memory_usage"` // Memory usage in bytes
	DiskUsage   int64   `json:"disk_usage" yaml:"disk_usage"`     // Disk usage in bytes
	NetworkRx   int64   `json:"network_rx" yaml:"network_rx"`     // Network received bytes/sec
	NetworkTx   int64   `json:"network_tx" yaml:"network_tx"`     // Network transmitted bytes/sec

	// GPU metrics
	GPUUsage  []float64 `json:"gpu_usage" yaml:"gpu_usage"`   // GPU usage percentage per GPU
	GPUMemory []int64   `json:"gpu_memory" yaml:"gpu_memory"` // GPU memory usage per GPU
	GPUTemp   []float64 `json:"gpu_temp" yaml:"gpu_temp"`     // GPU temperature per GPU

	// Performance metrics
	RequestsPerSec float64       `json:"requests_per_sec" yaml:"requests_per_sec"`
	AvgLatency     time.Duration `json:"avg_latency" yaml:"avg_latency"`
	ErrorRate      float64       `json:"error_rate" yaml:"error_rate"`

	Timestamp time.Time `json:"timestamp" yaml:"timestamp"`
}

// DistributedConfig holds configuration for the distributed Ollama system
type DistributedConfig struct {
	// API Configuration
	API struct {
		Port         int    `yaml:"port"`
		Host         string `yaml:"host"`
		CORSEnabled  bool   `yaml:"cors_enabled"`
		RateLimiting struct {
			Enabled           bool `yaml:"enabled"`
			RequestsPerMinute int  `yaml:"requests_per_minute"`
		} `yaml:"rate_limiting"`
	} `yaml:"api"`

	// P2P Network Configuration
	P2P *NodeConfig `yaml:"p2p"`

	// Model Management Configuration
	Models struct {
		StoragePath string `yaml:"storage_path"`
		CacheSize   string `yaml:"cache_size"`
		Replication struct {
			MinReplicas int    `yaml:"min_replicas"`
			MaxReplicas int    `yaml:"max_replicas"`
			Strategy    string `yaml:"strategy"`
		} `yaml:"replication"`
		Sync struct {
			Enabled  bool   `yaml:"enabled"`
			Interval string `yaml:"interval"`
		} `yaml:"sync"`
	} `yaml:"models"`

	// Distributed Inference Configuration
	Inference struct {
		MaxConcurrent int    `yaml:"max_concurrent"`
		Timeout       string `yaml:"timeout"`
		Partitioning  struct {
			Strategy         string `yaml:"strategy"`
			MinPartitionSize string `yaml:"min_partition_size"`
			MaxPartitions    int    `yaml:"max_partitions"`
		} `yaml:"partitioning"`
		Aggregation struct {
			Strategy string `yaml:"strategy"`
			Timeout  string `yaml:"timeout"`
		} `yaml:"aggregation"`
		LoadBalancing struct {
			Enabled   bool   `yaml:"enabled"`
			Algorithm string `yaml:"algorithm"`
		} `yaml:"load_balancing"`
		FaultTolerance struct {
			Enabled       bool   `yaml:"enabled"`
			RetryAttempts int    `yaml:"retry_attempts"`
			RetryDelay    string `yaml:"retry_delay"`
		} `yaml:"fault_tolerance"`
	} `yaml:"inference"`

	// Scheduler Configuration
	Scheduler struct {
		Algorithm          string `yaml:"algorithm"`
		QueueSize          int    `yaml:"queue_size"`
		WorkerPoolSize     int    `yaml:"worker_pool_size"`
		TaskTimeout        string `yaml:"task_timeout"`
		ResourceAllocation struct {
			CPUWeight     float64 `yaml:"cpu_weight"`
			MemoryWeight  float64 `yaml:"memory_weight"`
			NetworkWeight float64 `yaml:"network_weight"`
		} `yaml:"resource_allocation"`
	} `yaml:"scheduler"`

	// Monitoring Configuration
	Monitoring struct {
		Enabled             bool   `yaml:"enabled"`
		MetricsPort         int    `yaml:"metrics_port"`
		HealthCheckInterval string `yaml:"health_check_interval"`
		LogLevel            string `yaml:"log_level"`
		Tracing             struct {
			Enabled    bool    `yaml:"enabled"`
			Endpoint   string  `yaml:"endpoint"`
			SampleRate float64 `yaml:"sample_rate"`
		} `yaml:"tracing"`
	} `yaml:"monitoring"`

	// Security Configuration
	Security struct {
		TLS struct {
			Enabled  bool   `yaml:"enabled"`
			CertFile string `yaml:"cert_file"`
			KeyFile  string `yaml:"key_file"`
		} `yaml:"tls"`
		Authentication struct {
			Enabled bool   `yaml:"enabled"`
			Method  string `yaml:"method"`
		} `yaml:"authentication"`
		Authorization struct {
			Enabled  bool   `yaml:"enabled"`
			RBACFile string `yaml:"rbac_file"`
		} `yaml:"authorization"`
	} `yaml:"security"`

	// Performance Configuration
	Performance struct {
		Caching struct {
			Enabled bool   `yaml:"enabled"`
			Size    string `yaml:"size"`
			TTL     string `yaml:"ttl"`
		} `yaml:"caching"`
		Compression struct {
			Enabled   bool   `yaml:"enabled"`
			Algorithm string `yaml:"algorithm"`
			Level     int    `yaml:"level"`
		} `yaml:"compression"`
		ConnectionPooling struct {
			Enabled        bool   `yaml:"enabled"`
			MaxConnections int    `yaml:"max_connections"`
			IdleTimeout    string `yaml:"idle_timeout"`
		} `yaml:"connection_pooling"`
	} `yaml:"performance"`

	// Logging Configuration
	Logging struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
		Output string `yaml:"output"`
		File   struct {
			Enabled    bool   `yaml:"enabled"`
			Path       string `yaml:"path"`
			MaxSize    string `yaml:"max_size"`
			MaxBackups int    `yaml:"max_backups"`
			MaxAge     int    `yaml:"max_age"`
		} `yaml:"file"`
	} `yaml:"logging"`
}

// LoadDistributedConfig loads configuration from a YAML file
func LoadDistributedConfig(path string) (*DistributedConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config DistributedConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Set defaults if not specified
	if config.API.Port == 0 {
		config.API.Port = 11434
	}
	if config.API.Host == "" {
		config.API.Host = "0.0.0.0"
	}

	return &config, nil
}

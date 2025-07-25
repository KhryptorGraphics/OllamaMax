package config

import (
	"time"
	"crypto/rand"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// NodeConfig holds all configuration for a P2P node
type NodeConfig struct {
	// Network Settings
	Listen              []string            `yaml:"listen"`
	AnnounceAddresses   []string            `yaml:"announce_addresses"`
	NoAnnounceAddresses []string            `yaml:"no_announce_addresses"`
	
	// Security
	PrivateKey          string              `yaml:"private_key"`
	EnableTLS           bool                `yaml:"enable_tls"`
	EnableNoise         bool                `yaml:"enable_noise"`
	
	// NAT Traversal
	EnableNATService    bool                `yaml:"enable_nat_service"`
	EnableHolePunching  bool                `yaml:"enable_hole_punching"`
	EnableAutoRelay     bool                `yaml:"enable_auto_relay"`
	StaticRelays        []string            `yaml:"static_relays"`
	ForceReachability   string              `yaml:"force_reachability"` // public/private
	
	// DHT Settings
	EnableDHT           bool                `yaml:"enable_dht"`
	DHTMode             string              `yaml:"dht_mode"` // client/server/auto
	BootstrapPeers      []string            `yaml:"bootstrap_peers"`
	
	// Connection Management
	ConnMgrLow          int                 `yaml:"conn_mgr_low"`
	ConnMgrHigh         int                 `yaml:"conn_mgr_high"`
	ConnMgrGrace        time.Duration       `yaml:"conn_mgr_grace"`
	
	// Resource Management
	MaxMemory           int64               `yaml:"max_memory"`
	MaxCPU              float64             `yaml:"max_cpu"`
	MaxGPU              int                 `yaml:"max_gpu"`
	
	// Ollamacron Specific
	NodeType            string              `yaml:"node_type"` // edge/standard/super
	ModelCapabilities   []string            `yaml:"model_capabilities"`
	ResourceTags        map[string]string   `yaml:"resource_tags"`
	
	// Discovery Settings
	RendezvousString    string              `yaml:"rendezvous_string"`
	AutoDiscovery       bool                `yaml:"auto_discovery"`
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
		EnableTLS:           true,
		EnableNoise:         true,
		EnableNATService:    true,
		EnableHolePunching:  true,
		EnableAutoRelay:     true,
		EnableDHT:           true,
		DHTMode:             "auto",
		ConnMgrLow:          50,
		ConnMgrHigh:         200,
		ConnMgrGrace:        time.Minute,
		NodeType:            "standard",
		ModelCapabilities:   []string{},
		ResourceTags:        make(map[string]string),
		RendezvousString:    "ollama-distributed",
		AutoDiscovery:       true,
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
	CPUCores       int              `json:"cpu_cores"`
	Memory         int64            `json:"memory"`
	Storage        int64            `json:"storage"`
	GPUs           []*GPUInfo       `json:"gpus"`
	
	// AI capabilities
	SupportedModels []string         `json:"supported_models"`
	ModelFormats    []string         `json:"model_formats"`
	Quantizations   []string         `json:"quantizations"`
	
	// Network capabilities
	Bandwidth      int64            `json:"bandwidth"`
	Latency        time.Duration    `json:"latency"`
	Reliability    float64          `json:"reliability"`
	PricePerToken  float64          `json:"price_per_token"`
	
	// Node state
	Available      bool             `json:"available"`
	LoadFactor     float64          `json:"load_factor"`
	Priority       int              `json:"priority"`
	LastSeen       time.Time        `json:"last_seen"`
	
	// Version information
	Version        string           `json:"version"`
	ProtocolVersion string          `json:"protocol_version"`
	Features       []string         `json:"features"`
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
	CPUUsage    float64 `json:"cpu_usage" yaml:"cpu_usage"`         // CPU usage percentage (0-100)
	MemoryUsage int64   `json:"memory_usage" yaml:"memory_usage"`   // Memory usage in bytes
	DiskUsage   int64   `json:"disk_usage" yaml:"disk_usage"`       // Disk usage in bytes
	NetworkRx   int64   `json:"network_rx" yaml:"network_rx"`       // Network received bytes/sec
	NetworkTx   int64   `json:"network_tx" yaml:"network_tx"`       // Network transmitted bytes/sec
	
	// GPU metrics
	GPUUsage     []float64 `json:"gpu_usage" yaml:"gpu_usage"`         // GPU usage percentage per GPU
	GPUMemory    []int64   `json:"gpu_memory" yaml:"gpu_memory"`       // GPU memory usage per GPU
	GPUTemp      []float64 `json:"gpu_temp" yaml:"gpu_temp"`           // GPU temperature per GPU
	
	// Performance metrics
	RequestsPerSec  float64 `json:"requests_per_sec" yaml:"requests_per_sec"`
	AvgLatency      time.Duration `json:"avg_latency" yaml:"avg_latency"`
	ErrorRate       float64 `json:"error_rate" yaml:"error_rate"`
	
	Timestamp time.Time `json:"timestamp" yaml:"timestamp"`
}
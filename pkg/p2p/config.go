package p2p

import (
	"time"
)

// NodeConfig holds P2P node configuration
type NodeConfig struct {
	// Network configuration
	Listen       []string `json:"listen" yaml:"listen"`
	EnableNoise  bool     `json:"enable_noise" yaml:"enable_noise"`
	EnableRelay  bool     `json:"enable_relay" yaml:"enable_relay"`
	EnableAutoRelay bool  `json:"enable_auto_relay" yaml:"enable_auto_relay"`
	EnableHolePunch bool  `json:"enable_hole_punch" yaml:"enable_hole_punch"`
	EnableNAT    bool     `json:"enable_nat" yaml:"enable_nat"`
	EnableMDNS   bool     `json:"enable_mdns" yaml:"enable_mdns"`
	
	// Connection management
	ConnMgrLow   int           `json:"conn_mgr_low" yaml:"conn_mgr_low"`
	ConnMgrHigh  int           `json:"conn_mgr_high" yaml:"conn_mgr_high"`
	ConnMgrGrace time.Duration `json:"conn_mgr_grace" yaml:"conn_mgr_grace"`
	
	// Bootstrap and discovery
	BootstrapPeers  []string `json:"bootstrap_peers" yaml:"bootstrap_peers"`
	ProtocolPrefix  string   `json:"protocol_prefix" yaml:"protocol_prefix"`
	MDNSServiceName string   `json:"mdns_service_name" yaml:"mdns_service_name"`
}

// DefaultNodeConfig returns a default P2P node configuration
func DefaultNodeConfig() *NodeConfig {
	return &NodeConfig{
		Listen: []string{
			"/ip4/0.0.0.0/tcp/0",
			"/ip6/::/tcp/0",
		},
		EnableNoise:       true,
		EnableRelay:       true,
		EnableAutoRelay:   true,
		EnableHolePunch:   true,
		EnableNAT:         true,
		EnableMDNS:        true,
		ConnMgrLow:        10,
		ConnMgrHigh:       100,
		ConnMgrGrace:      30 * time.Second,
		BootstrapPeers:    []string{},
		ProtocolPrefix:    "/ollamamax",
		MDNSServiceName:   "ollamamax-node",
	}
}
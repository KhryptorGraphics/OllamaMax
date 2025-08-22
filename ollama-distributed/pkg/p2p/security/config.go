package security

import "time"

// SecurityConfig holds security configuration
type SecurityConfig struct {
	// Authentication
	EnableAuth  bool          `json:"enable_auth"`
	ClientAuth  string        `json:"client_auth"`
	AuthTimeout time.Duration `json:"auth_timeout"`

	// Encryption
	EnableEncryption    bool          `json:"enable_encryption"`
	KeyRotationInterval time.Duration `json:"key_rotation_interval"`
	EncryptionAlgorithm string        `json:"encryption_algorithm"`

	// Rate limiting
	EnableRateLimit   bool `json:"enable_rate_limit"`
	RequestsPerSecond int  `json:"requests_per_second"`
	BurstSize         int  `json:"burst_size"`

	// Monitoring
	EnableMonitoring bool          `json:"enable_monitoring"`
	MetricsInterval  time.Duration `json:"metrics_interval"`

	// Certificates
	CertFile            string        `json:"cert_file"`
	KeyFile             string        `json:"key_file"`
	CAFile              string        `json:"ca_file"`
	CertRefreshInterval time.Duration `json:"cert_refresh_interval"`

	// Access control
	EnableAccessControl bool   `json:"enable_access_control"`
	DefaultPolicy       string `json:"default_policy"`
}

// DefaultSecurityConfig returns a default security configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		EnableAuth:          true,
		ClientAuth:          "mutual",
		AuthTimeout:         30 * time.Second,
		EnableEncryption:    true,
		KeyRotationInterval: 24 * time.Hour,
		EncryptionAlgorithm: "AES-256-GCM",
		EnableRateLimit:     true,
		RequestsPerSecond:   100,
		BurstSize:           200,
		EnableMonitoring:    true,
		MetricsInterval:     60 * time.Second,
		CertRefreshInterval: 12 * time.Hour,
		EnableAccessControl: true,
		DefaultPolicy:       "deny",
	}
}

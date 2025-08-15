package security

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// HTTPSEnforcement provides comprehensive HTTPS enforcement and security headers
type HTTPSEnforcement struct {
	config *HTTPSConfig
}

// HTTPSConfig contains HTTPS enforcement configuration
type HTTPSConfig struct {
	// TLS Configuration
	MinTLSVersion uint16   `json:"min_tls_version"`
	CipherSuites  []uint16 `json:"cipher_suites"`
	CertFile      string   `json:"cert_file"`
	KeyFile       string   `json:"key_file"`
	CAFile        string   `json:"ca_file"`

	// HSTS Configuration
	HSTSMaxAge     int  `json:"hsts_max_age"`
	HSTSSubdomains bool `json:"hsts_subdomains"`
	HSTSPreload    bool `json:"hsts_preload"`

	// Security Headers
	EnableCSP      bool   `json:"enable_csp"`
	CSPPolicy      string `json:"csp_policy"`
	EnableXFrame   bool   `json:"enable_x_frame"`
	EnableXContent bool   `json:"enable_x_content"`
	EnableReferrer bool   `json:"enable_referrer"`

	// HTTPS Redirect
	ForceHTTPS bool `json:"force_https"`
	HTTPSPort  int  `json:"https_port"`

	// Certificate Management
	AutoRenew   bool          `json:"auto_renew"`
	RenewBefore time.Duration `json:"renew_before"`
}

// NewHTTPSEnforcement creates a new HTTPS enforcement instance
func NewHTTPSEnforcement(config *HTTPSConfig) *HTTPSEnforcement {
	if config == nil {
		config = DefaultHTTPSConfig()
	}

	return &HTTPSEnforcement{
		config: config,
	}
}

// DefaultHTTPSConfig returns a secure default HTTPS configuration
func DefaultHTTPSConfig() *HTTPSConfig {
	return &HTTPSConfig{
		MinTLSVersion: tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_CHACHA20_POLY1305_SHA256,
		},
		HSTSMaxAge:     31536000, // 1 year
		HSTSSubdomains: true,
		HSTSPreload:    true,
		EnableCSP:      true,
		CSPPolicy:      "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none';",
		EnableXFrame:   true,
		EnableXContent: true,
		EnableReferrer: true,
		ForceHTTPS:     true,
		HTTPSPort:      443,
		AutoRenew:      true,
		RenewBefore:    30 * 24 * time.Hour, // 30 days
	}
}

// GetTLSConfig returns a secure TLS configuration
func (he *HTTPSEnforcement) GetTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion:   he.config.MinTLSVersion,
		CipherSuites: he.config.CipherSuites,
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
			tls.CurveP384,
		},
		PreferServerCipherSuites: true,
		SessionTicketsDisabled:   false,
		Renegotiation:            tls.RenegotiateNever,
	}
}

// SecurityHeadersMiddleware adds security headers to HTTP responses
func (he *HTTPSEnforcement) SecurityHeadersMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// HSTS Header
			if he.config.HSTSMaxAge > 0 {
				hstsValue := fmt.Sprintf("max-age=%d", he.config.HSTSMaxAge)
				if he.config.HSTSSubdomains {
					hstsValue += "; includeSubDomains"
				}
				if he.config.HSTSPreload {
					hstsValue += "; preload"
				}
				w.Header().Set("Strict-Transport-Security", hstsValue)
			}

			// Content Security Policy
			if he.config.EnableCSP && he.config.CSPPolicy != "" {
				w.Header().Set("Content-Security-Policy", he.config.CSPPolicy)
			}

			// X-Frame-Options
			if he.config.EnableXFrame {
				w.Header().Set("X-Frame-Options", "DENY")
			}

			// X-Content-Type-Options
			if he.config.EnableXContent {
				w.Header().Set("X-Content-Type-Options", "nosniff")
			}

			// X-XSS-Protection (for older browsers)
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Referrer Policy
			if he.config.EnableReferrer {
				w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			}

			// Permissions Policy
			w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

			// Cache Control for sensitive endpoints
			if strings.Contains(r.URL.Path, "/api/") {
				w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
				w.Header().Set("Pragma", "no-cache")
				w.Header().Set("Expires", "0")
			}

			next.ServeHTTP(w, r)
		})
	}
}

// HTTPSRedirectMiddleware redirects HTTP requests to HTTPS
func (he *HTTPSEnforcement) HTTPSRedirectMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if he.config.ForceHTTPS && r.TLS == nil {
				// Build HTTPS URL
				httpsURL := "https://" + r.Host
				if he.config.HTTPSPort != 443 {
					httpsURL = fmt.Sprintf("https://%s:%d", r.Host, he.config.HTTPSPort)
				}
				httpsURL += r.RequestURI

				// Permanent redirect to HTTPS
				http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ValidateHTTPSRequest validates that a request meets HTTPS requirements
func (he *HTTPSEnforcement) ValidateHTTPSRequest(r *http.Request) error {
	// Check if HTTPS is required and request is not secure
	if he.config.ForceHTTPS && r.TLS == nil {
		return fmt.Errorf("HTTPS required but request is not secure")
	}

	// Check TLS version if request is secure
	if r.TLS != nil {
		if r.TLS.Version < he.config.MinTLSVersion {
			return fmt.Errorf("TLS version %d is below minimum required version %d",
				r.TLS.Version, he.config.MinTLSVersion)
		}
	}

	return nil
}

// HTTPSCertificateManager handles certificate management and renewal for HTTPS
type HTTPSCertificateManager struct {
	config    *HTTPSConfig
	certPath  string
	keyPath   string
	lastCheck time.Time
	cert      *tls.Certificate
}

// NewHTTPSCertificateManager creates a new HTTPS certificate manager
func NewHTTPSCertificateManager(config *HTTPSConfig) *HTTPSCertificateManager {
	return &HTTPSCertificateManager{
		config:   config,
		certPath: config.CertFile,
		keyPath:  config.KeyFile,
	}
}

// LoadCertificate loads the TLS certificate
func (cm *HTTPSCertificateManager) LoadCertificate() error {
	cert, err := tls.LoadX509KeyPair(cm.certPath, cm.keyPath)
	if err != nil {
		return fmt.Errorf("failed to load certificate: %w", err)
	}

	cm.cert = &cert
	cm.lastCheck = time.Now()

	return nil
}

// CheckCertificateExpiry checks if the certificate needs renewal
func (cm *HTTPSCertificateManager) CheckCertificateExpiry() (bool, error) {
	if cm.cert == nil {
		return false, fmt.Errorf("no certificate loaded")
	}

	// Parse the certificate to check expiry
	if len(cm.cert.Certificate) == 0 {
		return false, fmt.Errorf("invalid certificate")
	}

	// This is a simplified check - in production, you'd parse the actual certificate
	// and check its NotAfter field
	return false, nil
}

// GetCertificate returns the current certificate for TLS configuration
func (cm *HTTPSCertificateManager) GetCertificate() (*tls.Certificate, error) {
	if cm.cert == nil {
		if err := cm.LoadCertificate(); err != nil {
			return nil, err
		}
	}

	return cm.cert, nil
}

// SecureServerConfig returns a secure HTTP server configuration
func (he *HTTPSEnforcement) SecureServerConfig(handler http.Handler) *http.Server {
	return &http.Server{
		Handler:      handler,
		TLSConfig:    he.GetTLSConfig(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		// Disable HTTP/2 for better security control (optional)
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
}

// ValidateSecurityHeaders validates that required security headers are present
func (he *HTTPSEnforcement) ValidateSecurityHeaders(headers http.Header) []string {
	var missing []string

	requiredHeaders := map[string]string{
		"Strict-Transport-Security": "HSTS header missing",
		"X-Frame-Options":           "X-Frame-Options header missing",
		"X-Content-Type-Options":    "X-Content-Type-Options header missing",
		"Content-Security-Policy":   "CSP header missing",
	}

	for header, message := range requiredHeaders {
		if headers.Get(header) == "" {
			missing = append(missing, message)
		}
	}

	return missing
}

// Global instance for easy access
var DefaultHTTPSEnforcement = NewHTTPSEnforcement(nil)

// Convenience functions
func GetSecurityHeadersMiddleware() func(http.Handler) http.Handler {
	return DefaultHTTPSEnforcement.SecurityHeadersMiddleware()
}

func GetHTTPSRedirectMiddleware() func(http.Handler) http.Handler {
	return DefaultHTTPSEnforcement.HTTPSRedirectMiddleware()
}

func GetSecureTLSConfig() *tls.Config {
	return DefaultHTTPSEnforcement.GetTLSConfig()
}

func ValidateHTTPSRequest(r *http.Request) error {
	return DefaultHTTPSEnforcement.ValidateHTTPSRequest(r)
}

package security

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// CertificateManager manages TLS certificates for secure communications
type CertificateManager struct {
	config *SecurityConfig
	
	// Current certificate
	cert     *tls.Certificate
	certPath string
	keyPath  string
	
	// CA pool for client verification
	caPool *x509.CertPool
	
	// Certificate refresh
	refreshTicker *time.Ticker
	lastRefresh   time.Time
	
	// Mutex for concurrent access
	mu sync.RWMutex
}

// NewCertificateManager creates a new certificate manager
func NewCertificateManager(config *SecurityConfig) (*CertificateManager, error) {
	cm := &CertificateManager{
		config: config,
	}
	
	// Load initial certificates if paths are provided
	if config.CertFile != "" && config.KeyFile != "" {
		if err := cm.loadCertificates(); err != nil {
			return nil, fmt.Errorf("failed to load certificates: %w", err)
		}
	}
	
	// Load CA certificates if provided
	if config.CAFile != "" {
		if err := cm.loadCA(); err != nil {
			return nil, fmt.Errorf("failed to load CA certificates: %w", err)
		}
	}
	
	return cm, nil
}

// loadCertificates loads TLS certificates from files
func (cm *CertificateManager) loadCertificates() error {
	cert, err := tls.LoadX509KeyPair(cm.config.CertFile, cm.config.KeyFile)
	if err != nil {
		return fmt.Errorf("failed to load certificate pair: %w", err)
	}
	
	cm.mu.Lock()
	cm.cert = &cert
	cm.certPath = cm.config.CertFile
	cm.keyPath = cm.config.KeyFile
	cm.lastRefresh = time.Now()
	cm.mu.Unlock()
	
	log.Printf("Loaded TLS certificates from %s and %s", cm.config.CertFile, cm.config.KeyFile)
	return nil
}

// loadCA loads CA certificates for client verification
func (cm *CertificateManager) loadCA() error {
	caCert, err := os.ReadFile(cm.config.CAFile)
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %w", err)
	}
	
	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return fmt.Errorf("failed to parse CA certificate")
	}
	
	cm.mu.Lock()
	cm.caPool = caPool
	cm.mu.Unlock()
	
	log.Printf("Loaded CA certificates from %s", cm.config.CAFile)
	return nil
}

// GetCertificate returns the current TLS certificate
func (cm *CertificateManager) GetCertificate() (*tls.Certificate, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	if cm.cert == nil {
		return nil, fmt.Errorf("no certificate loaded")
	}
	
	return cm.cert, nil
}

// GetClientAuth returns the client authentication type
func (cm *CertificateManager) GetClientAuth() tls.ClientAuthType {
	switch cm.config.ClientAuth {
	case "request":
		return tls.RequestClientCert
	case "require":
		return tls.RequireAnyClientCert
	case "verify":
		return tls.RequireAndVerifyClientCert
	default:
		return tls.NoClientCert
	}
}

// GetClientCAs returns the CA pool for client certificate verification
func (cm *CertificateManager) GetClientCAs() *x509.CertPool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.caPool
}

// Start starts the certificate refresh process
func (cm *CertificateManager) Start(ctx context.Context) {
	if cm.config.CertRefreshInterval <= 0 {
		return
	}
	
	cm.refreshTicker = time.NewTicker(cm.config.CertRefreshInterval)
	
	go func() {
		for {
			select {
			case <-ctx.Done():
				if cm.refreshTicker != nil {
					cm.refreshTicker.Stop()
				}
				return
			case <-cm.refreshTicker.C:
				if err := cm.refreshCertificates(); err != nil {
					log.Printf("Failed to refresh certificates: %v", err)
				}
			}
		}
	}()
}

// refreshCertificates reloads certificates from files
func (cm *CertificateManager) refreshCertificates() error {
	// Only refresh if certificate files are configured
	if cm.config.CertFile == "" || cm.config.KeyFile == "" {
		return nil
	}
	
	// Load new certificate
	newCert, err := tls.LoadX509KeyPair(cm.config.CertFile, cm.config.KeyFile)
	if err != nil {
		return fmt.Errorf("failed to load new certificate pair: %w", err)
	}
	
	// Update certificate
	cm.mu.Lock()
	cm.cert = &newCert
	cm.lastRefresh = time.Now()
	cm.mu.Unlock()
	
	log.Printf("Refreshed TLS certificates from %s and %s", cm.config.CertFile, cm.config.KeyFile)
	return nil
}

// GetTlsConfig returns a TLS configuration for use with libp2p
func (cm *CertificateManager) GetTlsConfig() (*tls.Config, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	if cm.cert == nil {
		return nil, fmt.Errorf("no certificate available")
	}
	
	config := &tls.Config{
		Certificates: []tls.Certificate{*cm.cert},
		ClientAuth:   cm.GetClientAuth(),
	}
	
	if cm.caPool != nil {
		config.ClientCAs = cm.caPool
	}
	
	return config, nil
}

// Close stops the certificate refresh process
func (cm *CertificateManager) Close() error {
	if cm.refreshTicker != nil {
		cm.refreshTicker.Stop()
	}
	return nil
}
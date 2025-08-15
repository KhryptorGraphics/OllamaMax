package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log/slog"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CertificateManager handles TLS certificate lifecycle management
type CertificateManager struct {
	logger *slog.Logger
	config *CertificateConfig
	
	// Certificate storage
	certificates map[string]*ManagedCertificate
	certMutex    sync.RWMutex
	
	// Auto-renewal
	renewalTicker *time.Ticker
	stopChan      chan struct{}
	
	// Certificate authority
	caCert *x509.Certificate
	caKey  *rsa.PrivateKey
}

// CertificateConfig contains certificate management configuration
type CertificateConfig struct {
	// Storage paths
	CertDir     string `json:"cert_dir"`
	CAFile      string `json:"ca_file"`
	CAKeyFile   string `json:"ca_key_file"`
	
	// Auto-renewal settings
	RenewalThreshold time.Duration `json:"renewal_threshold"`
	CheckInterval    time.Duration `json:"check_interval"`
	AutoRenew        bool          `json:"auto_renew"`
	
	// Certificate settings
	KeySize        int           `json:"key_size"`
	ValidityPeriod time.Duration `json:"validity_period"`
	Organization   string        `json:"organization"`
	Country        string        `json:"country"`
	
	// Backup settings
	BackupEnabled bool   `json:"backup_enabled"`
	BackupDir     string `json:"backup_dir"`
}

// ManagedCertificate represents a managed TLS certificate
type ManagedCertificate struct {
	Name        string             `json:"name"`
	Domains     []string           `json:"domains"`
	CertPath    string             `json:"cert_path"`
	KeyPath     string             `json:"key_path"`
	Certificate *x509.Certificate  `json:"-"`
	PrivateKey  *rsa.PrivateKey    `json:"-"`
	TLSCert     *tls.Certificate   `json:"-"`
	CreatedAt   time.Time          `json:"created_at"`
	ExpiresAt   time.Time          `json:"expires_at"`
	LastRenewed time.Time          `json:"last_renewed"`
	AutoRenew   bool               `json:"auto_renew"`
	mutex       sync.RWMutex       `json:"-"`
}

// NewCertificateManager creates a new certificate manager
func NewCertificateManager(config *CertificateConfig, logger *slog.Logger) (*CertificateManager, error) {
	if config == nil {
		config = DefaultCertificateConfig()
	}
	
	cm := &CertificateManager{
		logger:       logger,
		config:       config,
		certificates: make(map[string]*ManagedCertificate),
		stopChan:     make(chan struct{}),
	}
	
	// Create certificate directory
	if err := os.MkdirAll(config.CertDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create certificate directory: %w", err)
	}
	
	// Create backup directory if enabled
	if config.BackupEnabled {
		if err := os.MkdirAll(config.BackupDir, 0700); err != nil {
			return nil, fmt.Errorf("failed to create backup directory: %w", err)
		}
	}
	
	// Load or create CA certificate
	if err := cm.initializeCA(); err != nil {
		return nil, fmt.Errorf("failed to initialize CA: %w", err)
	}
	
	// Load existing certificates
	if err := cm.loadExistingCertificates(); err != nil {
		logger.Warn("failed to load existing certificates", "error", err)
	}
	
	return cm, nil
}

// DefaultCertificateConfig returns default certificate configuration
func DefaultCertificateConfig() *CertificateConfig {
	return &CertificateConfig{
		CertDir:          "./certs",
		CAFile:           "./certs/ca.crt",
		CAKeyFile:        "./certs/ca.key",
		RenewalThreshold: 30 * 24 * time.Hour, // 30 days
		CheckInterval:    24 * time.Hour,      // Daily checks
		AutoRenew:        true,
		KeySize:          2048,
		ValidityPeriod:   365 * 24 * time.Hour, // 1 year
		Organization:     "OllamaMax Distributed",
		Country:          "US",
		BackupEnabled:    true,
		BackupDir:        "./certs/backup",
	}
}

// Start starts the certificate manager
func (cm *CertificateManager) Start() error {
	if cm.config.AutoRenew {
		cm.renewalTicker = time.NewTicker(cm.config.CheckInterval)
		go cm.renewalLoop()
		cm.logger.Info("certificate auto-renewal started", "interval", cm.config.CheckInterval)
	}
	
	return nil
}

// Stop stops the certificate manager
func (cm *CertificateManager) Stop() error {
	if cm.renewalTicker != nil {
		cm.renewalTicker.Stop()
	}
	
	close(cm.stopChan)
	cm.logger.Info("certificate manager stopped")
	
	return nil
}

// CreateCertificate creates a new TLS certificate
func (cm *CertificateManager) CreateCertificate(name string, domains []string) (*ManagedCertificate, error) {
	cm.logger.Info("creating certificate", "name", name, "domains", domains)
	
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, cm.config.KeySize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}
	
	// Create certificate template
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{cm.config.Organization},
			Country:       []string{cm.config.Country},
			CommonName:    domains[0],
		},
		DNSNames:              domains,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(cm.config.ValidityPeriod),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	
	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, template, cm.caCert, &privateKey.PublicKey, cm.caKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}
	
	// Parse certificate
	certificate, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}
	
	// Create managed certificate
	managedCert := &ManagedCertificate{
		Name:        name,
		Domains:     domains,
		CertPath:    filepath.Join(cm.config.CertDir, name+".crt"),
		KeyPath:     filepath.Join(cm.config.CertDir, name+".key"),
		Certificate: certificate,
		PrivateKey:  privateKey,
		CreatedAt:   time.Now(),
		ExpiresAt:   certificate.NotAfter,
		AutoRenew:   true,
	}
	
	// Create TLS certificate
	tlsCert := tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  privateKey,
	}
	managedCert.TLSCert = &tlsCert
	
	// Save certificate to disk
	if err := cm.saveCertificate(managedCert); err != nil {
		return nil, fmt.Errorf("failed to save certificate: %w", err)
	}
	
	// Store in memory
	cm.certMutex.Lock()
	cm.certificates[name] = managedCert
	cm.certMutex.Unlock()
	
	cm.logger.Info("certificate created successfully", "name", name, "expires", certificate.NotAfter)
	
	return managedCert, nil
}

// GetCertificate retrieves a managed certificate
func (cm *CertificateManager) GetCertificate(name string) (*ManagedCertificate, error) {
	cm.certMutex.RLock()
	cert, exists := cm.certificates[name]
	cm.certMutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("certificate not found: %s", name)
	}
	
	return cert, nil
}

// RenewCertificate renews an existing certificate
func (cm *CertificateManager) RenewCertificate(name string) error {
	cm.certMutex.RLock()
	cert, exists := cm.certificates[name]
	cm.certMutex.RUnlock()
	
	if !exists {
		return fmt.Errorf("certificate not found: %s", name)
	}
	
	cm.logger.Info("renewing certificate", "name", name)
	
	// Backup old certificate if enabled
	if cm.config.BackupEnabled {
		if err := cm.backupCertificate(cert); err != nil {
			cm.logger.Warn("failed to backup certificate", "name", name, "error", err)
		}
	}
	
	// Create new certificate with same domains
	newCert, err := cm.CreateCertificate(name, cert.Domains)
	if err != nil {
		return fmt.Errorf("failed to renew certificate: %w", err)
	}
	
	newCert.LastRenewed = time.Now()
	
	cm.logger.Info("certificate renewed successfully", "name", name, "expires", newCert.ExpiresAt)
	
	return nil
}

// CheckExpiry checks if certificates need renewal
func (cm *CertificateManager) CheckExpiry() []string {
	var expiring []string
	
	cm.certMutex.RLock()
	defer cm.certMutex.RUnlock()
	
	for name, cert := range cm.certificates {
		if cert.AutoRenew && time.Until(cert.ExpiresAt) < cm.config.RenewalThreshold {
			expiring = append(expiring, name)
		}
	}
	
	return expiring
}

// renewalLoop runs the automatic renewal process
func (cm *CertificateManager) renewalLoop() {
	for {
		select {
		case <-cm.renewalTicker.C:
			expiring := cm.CheckExpiry()
			for _, name := range expiring {
				if err := cm.RenewCertificate(name); err != nil {
					cm.logger.Error("failed to auto-renew certificate", "name", name, "error", err)
				}
			}
		case <-cm.stopChan:
			return
		}
	}
}

// initializeCA initializes the certificate authority
func (cm *CertificateManager) initializeCA() error {
	// Try to load existing CA
	if _, err := os.Stat(cm.config.CAFile); err == nil {
		return cm.loadCA()
	}
	
	// Create new CA
	return cm.createCA()
}

// loadCA loads an existing CA certificate and key
func (cm *CertificateManager) loadCA() error {
	// Load CA certificate
	certPEM, err := ioutil.ReadFile(cm.config.CAFile)
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %w", err)
	}
	
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return fmt.Errorf("failed to decode CA certificate PEM")
	}
	
	cm.caCert, err = x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CA certificate: %w", err)
	}
	
	// Load CA private key
	keyPEM, err := ioutil.ReadFile(cm.config.CAKeyFile)
	if err != nil {
		return fmt.Errorf("failed to read CA private key: %w", err)
	}
	
	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return fmt.Errorf("failed to decode CA private key PEM")
	}
	
	cm.caKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CA private key: %w", err)
	}
	
	cm.logger.Info("CA certificate loaded", "expires", cm.caCert.NotAfter)
	
	return nil
}

// createCA creates a new certificate authority
func (cm *CertificateManager) createCA() error {
	cm.logger.Info("creating new CA certificate")
	
	// Generate CA private key
	caKey, err := rsa.GenerateKey(rand.Reader, cm.config.KeySize)
	if err != nil {
		return fmt.Errorf("failed to generate CA private key: %w", err)
	}
	
	// Create CA certificate template
	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{cm.config.Organization + " CA"},
			Country:       []string{cm.config.Country},
			CommonName:    "OllamaMax CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour), // 10 years
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	
	// Create CA certificate
	caCertDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("failed to create CA certificate: %w", err)
	}
	
	// Parse CA certificate
	cm.caCert, err = x509.ParseCertificate(caCertDER)
	if err != nil {
		return fmt.Errorf("failed to parse CA certificate: %w", err)
	}
	
	cm.caKey = caKey
	
	// Save CA certificate
	if err := cm.saveCA(); err != nil {
		return fmt.Errorf("failed to save CA: %w", err)
	}
	
	cm.logger.Info("CA certificate created", "expires", cm.caCert.NotAfter)
	
	return nil
}

// saveCA saves the CA certificate and key to disk
func (cm *CertificateManager) saveCA() error {
	// Save CA certificate
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cm.caCert.Raw,
	})
	
	if err := ioutil.WriteFile(cm.config.CAFile, certPEM, 0644); err != nil {
		return fmt.Errorf("failed to save CA certificate: %w", err)
	}
	
	// Save CA private key
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(cm.caKey),
	})
	
	if err := ioutil.WriteFile(cm.config.CAKeyFile, keyPEM, 0600); err != nil {
		return fmt.Errorf("failed to save CA private key: %w", err)
	}
	
	return nil
}

// saveCertificate saves a managed certificate to disk
func (cm *CertificateManager) saveCertificate(cert *ManagedCertificate) error {
	// Save certificate
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Certificate.Raw,
	})
	
	if err := ioutil.WriteFile(cert.CertPath, certPEM, 0644); err != nil {
		return fmt.Errorf("failed to save certificate: %w", err)
	}
	
	// Save private key
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(cert.PrivateKey),
	})
	
	if err := ioutil.WriteFile(cert.KeyPath, keyPEM, 0600); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}
	
	return nil
}

// backupCertificate creates a backup of a certificate
func (cm *CertificateManager) backupCertificate(cert *ManagedCertificate) error {
	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("%s-%s", cert.Name, timestamp)
	
	backupCertPath := filepath.Join(cm.config.BackupDir, backupName+".crt")
	backupKeyPath := filepath.Join(cm.config.BackupDir, backupName+".key")
	
	// Copy certificate
	certData, err := ioutil.ReadFile(cert.CertPath)
	if err != nil {
		return err
	}
	
	if err := ioutil.WriteFile(backupCertPath, certData, 0644); err != nil {
		return err
	}
	
	// Copy private key
	keyData, err := ioutil.ReadFile(cert.KeyPath)
	if err != nil {
		return err
	}
	
	if err := ioutil.WriteFile(backupKeyPath, keyData, 0600); err != nil {
		return err
	}
	
	cm.logger.Info("certificate backed up", "name", cert.Name, "backup", backupName)
	
	return nil
}

// loadExistingCertificates loads certificates from disk
func (cm *CertificateManager) loadExistingCertificates() error {
	// This would scan the certificate directory and load existing certificates
	// Implementation depends on how certificates are stored and named
	return nil
}

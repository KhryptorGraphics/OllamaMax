package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
)

// AuthManager handles authentication and authorization
type AuthManager struct {
	config     *config.AuthConfig
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey

	// Certificate management
	certManager *CertificateManager

	// Token blacklist
	blacklistedTokens map[string]time.Time
}

// CertificateManager handles X.509 certificate operations
type CertificateManager struct {
	caCert     *x509.Certificate
	caKey      *rsa.PrivateKey
	serverCert *x509.Certificate
	serverKey  *rsa.PrivateKey
}

// Claims represents JWT claims
type Claims struct {
	UserID      string            `json:"user_id"`
	Username    string            `json:"username"`
	Role        string            `json:"role"`
	Permissions []string          `json:"permissions"`
	Metadata    map[string]string `json:"metadata"`
	jwt.RegisteredClaims
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(config *config.AuthConfig) (*AuthManager, error) {
	manager := &AuthManager{
		config:            config,
		blacklistedTokens: make(map[string]time.Time),
	}

	// Generate or load RSA key pair
	if err := manager.initializeKeys(); err != nil {
		return nil, fmt.Errorf("failed to initialize keys: %w", err)
	}

	// Initialize certificate manager
	if err := manager.initializeCertificates(); err != nil {
		return nil, fmt.Errorf("failed to initialize certificates: %w", err)
	}

	// Start cleanup routine for blacklisted tokens
	go manager.cleanupBlacklistedTokens()

	return manager, nil
}

// initializeKeys initializes RSA key pair for JWT signing
func (am *AuthManager) initializeKeys() error {
	// Generate RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	am.privateKey = privateKey
	am.publicKey = &privateKey.PublicKey

	return nil
}

// initializeCertificates initializes X.509 certificates
func (am *AuthManager) initializeCertificates() error {
	certManager := &CertificateManager{}

	// Generate CA certificate
	caTemplate := &x509.Certificate{
		SerialNumber: bigIntFromString("1"),
		Subject: pkix.Name{
			Organization:  []string{"Ollama Distributed"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 10 years
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// Generate CA key
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate CA key: %w", err)
	}

	// Create CA certificate
	caCertDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("failed to create CA certificate: %w", err)
	}

	caCert, err := x509.ParseCertificate(caCertDER)
	if err != nil {
		return fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	certManager.caCert = caCert
	certManager.caKey = caKey

	// Generate server certificate
	serverTemplate := &x509.Certificate{
		SerialNumber: bigIntFromString("2"),
		Subject: pkix.Name{
			Organization:  []string{"Ollama Distributed"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0), // 1 year
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	// Generate server key
	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate server key: %w", err)
	}

	// Create server certificate
	serverCertDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caCert, &serverKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("failed to create server certificate: %w", err)
	}

	serverCert, err := x509.ParseCertificate(serverCertDER)
	if err != nil {
		return fmt.Errorf("failed to parse server certificate: %w", err)
	}

	certManager.serverCert = serverCert
	certManager.serverKey = serverKey

	am.certManager = certManager

	return nil
}

// GenerateToken generates a JWT token for a user
func (am *AuthManager) GenerateToken(userID, username, role string, permissions []string) (string, error) {
	claims := &Claims{
		UserID:      userID,
		Username:    username,
		Role:        role,
		Permissions: permissions,
		Metadata:    make(map[string]string),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(am.config.TokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    am.config.Issuer,
			Subject:   userID,
			ID:        generateTokenID(),
			Audience:  []string{am.config.Audience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(am.privateKey)
}

// ValidateToken validates a JWT token
func (am *AuthManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return am.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Check if token is blacklisted
		if am.isTokenBlacklisted(claims.ID) {
			return nil, fmt.Errorf("token is blacklisted")
		}

		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// BlacklistToken adds a token to the blacklist
func (am *AuthManager) BlacklistToken(tokenID string, expiry time.Time) {
	am.blacklistedTokens[tokenID] = expiry
}

// isTokenBlacklisted checks if a token is blacklisted
func (am *AuthManager) isTokenBlacklisted(tokenID string) bool {
	expiry, exists := am.blacklistedTokens[tokenID]
	if !exists {
		return false
	}

	// Check if blacklist entry has expired
	if time.Now().After(expiry) {
		delete(am.blacklistedTokens, tokenID)
		return false
	}

	return true
}

// cleanupBlacklistedTokens periodically cleans up expired blacklisted tokens
func (am *AuthManager) cleanupBlacklistedTokens() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		for tokenID, expiry := range am.blacklistedTokens {
			if now.After(expiry) {
				delete(am.blacklistedTokens, tokenID)
			}
		}
	}
}

// AuthMiddleware returns a Gin middleware for authentication
func (am *AuthManager) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !am.config.Enabled {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
			c.Abort()
			return
		}

		claims, err := am.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Store claims in context
		c.Set("claims", claims)
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("permissions", claims.Permissions)

		c.Next()
	}
}

// RequirePermission returns a middleware that requires specific permissions
func (am *AuthManager) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
			c.Abort()
			return
		}

		// Check if user has required permission
		hasPermission := false
		for _, perm := range userClaims.Permissions {
			if perm == permission || perm == "admin" {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole returns a middleware that requires a specific role
func (am *AuthManager) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
			c.Abort()
			return
		}

		if userClaims.Role != role && userClaims.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient role"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetCACertificate returns the CA certificate in PEM format
func (am *AuthManager) GetCACertificate() ([]byte, error) {
	if am.certManager == nil || am.certManager.caCert == nil {
		return nil, fmt.Errorf("CA certificate not available")
	}

	certPEM := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: am.certManager.caCert.Raw,
	}

	return pem.EncodeToMemory(certPEM), nil
}

// GetServerCertificate returns the server certificate in PEM format
func (am *AuthManager) GetServerCertificate() ([]byte, error) {
	if am.certManager == nil || am.certManager.serverCert == nil {
		return nil, fmt.Errorf("server certificate not available")
	}

	certPEM := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: am.certManager.serverCert.Raw,
	}

	return pem.EncodeToMemory(certPEM), nil
}

// GetServerKey returns the server private key in PEM format
func (am *AuthManager) GetServerKey() ([]byte, error) {
	if am.certManager == nil || am.certManager.serverKey == nil {
		return nil, fmt.Errorf("server key not available")
	}

	keyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(am.certManager.serverKey),
	}

	return pem.EncodeToMemory(keyPEM), nil
}

// Utility functions

func generateTokenID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

func bigIntFromString(s string) *big.Int {
	i := new(big.Int)
	i.SetString(s, 10)
	return i
}

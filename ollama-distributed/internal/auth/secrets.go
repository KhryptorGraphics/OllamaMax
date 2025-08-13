package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// SecretManager manages cryptographic secrets and keys
type SecretManager struct {
	mu           sync.RWMutex
	secrets      map[string]*Secret
	keyRotation  map[string]time.Time
	rotationLock sync.Mutex
}

// Secret represents a managed secret
type Secret struct {
	ID             string
	Value          []byte
	CreatedAt      time.Time
	ExpiresAt      *time.Time
	RotationPeriod time.Duration
	Type           SecretType
	Metadata       map[string]string
}

// SecretType defines the type of secret
type SecretType string

const (
	SecretTypeJWT        SecretType = "jwt"
	SecretTypeEncryption SecretType = "encryption"
	SecretTypeAPI        SecretType = "api"
	SecretTypeDatabase   SecretType = "database"
)

// NewSecretManager creates a new secret manager
func NewSecretManager() *SecretManager {
	return &SecretManager{
		secrets:     make(map[string]*Secret),
		keyRotation: make(map[string]time.Time),
	}
}

// GetOrCreateJWTSecret gets or creates a JWT signing secret
func (sm *SecretManager) GetOrCreateJWTSecret() ([]byte, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Try to get from environment first
	if secretEnv := os.Getenv("JWT_SECRET"); secretEnv != "" {
		// Validate minimum length for security
		if len(secretEnv) < 32 {
			return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters long")
		}

		// Store in secret manager
		secret := &Secret{
			ID:        "jwt_primary",
			Value:     []byte(secretEnv),
			CreatedAt: time.Now(),
			Type:      SecretTypeJWT,
			Metadata:  map[string]string{"source": "environment"},
		}
		sm.secrets["jwt_primary"] = secret
		return secret.Value, nil
	}

	// Check if we already have a generated secret
	if existing, exists := sm.secrets["jwt_primary"]; exists {
		return existing.Value, nil
	}

	// Generate a new secret
	secretBytes := make([]byte, 64) // 512-bit secret
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, fmt.Errorf("failed to generate JWT secret: %w", err)
	}

	secret := &Secret{
		ID:             "jwt_primary",
		Value:          secretBytes,
		CreatedAt:      time.Now(),
		RotationPeriod: 90 * 24 * time.Hour, // 90 days
		Type:           SecretTypeJWT,
		Metadata:       map[string]string{"source": "generated"},
	}

	sm.secrets["jwt_primary"] = secret

	// Schedule rotation
	sm.scheduleRotation("jwt_primary", secret.RotationPeriod)

	return secret.Value, nil
}

// GetEncryptionKey gets or creates an encryption key
func (sm *SecretManager) GetEncryptionKey(keyID string) ([]byte, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if key exists
	if existing, exists := sm.secrets[keyID]; exists {
		return existing.Value, nil
	}

	// Generate new encryption key (AES-256)
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}

	secret := &Secret{
		ID:             keyID,
		Value:          keyBytes,
		CreatedAt:      time.Now(),
		RotationPeriod: 30 * 24 * time.Hour, // 30 days for encryption keys
		Type:           SecretTypeEncryption,
		Metadata:       map[string]string{"algorithm": "AES-256-GCM"},
	}

	sm.secrets[keyID] = secret
	sm.scheduleRotation(keyID, secret.RotationPeriod)

	return secret.Value, nil
}

// ValidateEnvironmentSecrets validates all required environment secrets
func (sm *SecretManager) ValidateEnvironmentSecrets() error {
	var errors []string

	// Check JWT secret requirements
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret != "" {
		if len(jwtSecret) < 32 {
			errors = append(errors, "JWT_SECRET must be at least 32 characters long")
		}
		if strings.Contains(jwtSecret, " ") {
			errors = append(errors, "JWT_SECRET should not contain spaces")
		}
		// Check for common weak patterns
		if jwtSecret == "secret" || jwtSecret == "password" || jwtSecret == "default" {
			errors = append(errors, "JWT_SECRET uses a weak/default value")
		}
	}

	// Check database credentials if required
	if dbPass := os.Getenv("DB_PASSWORD"); dbPass != "" {
		if len(dbPass) < 12 {
			errors = append(errors, "DB_PASSWORD should be at least 12 characters long")
		}
	}

	// Check admin password
	if adminPass := os.Getenv("ADMIN_DEFAULT_PASSWORD"); adminPass != "" {
		if len(adminPass) < 8 {
			errors = append(errors, "ADMIN_DEFAULT_PASSWORD must be at least 8 characters long")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("environment secret validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// RotateSecret rotates a secret by generating a new value
func (sm *SecretManager) RotateSecret(secretID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	existing, exists := sm.secrets[secretID]
	if !exists {
		return fmt.Errorf("secret %s not found", secretID)
	}

	// Generate new secret based on type
	var newValue []byte
	var err error

	switch existing.Type {
	case SecretTypeJWT:
		newValue = make([]byte, 64)
		_, err = rand.Read(newValue)
	case SecretTypeEncryption:
		newValue = make([]byte, 32)
		_, err = rand.Read(newValue)
	case SecretTypeAPI:
		newValue = make([]byte, 32)
		_, err = rand.Read(newValue)
		if err == nil {
			// Encode as base64 for API keys
			encoded := base64.URLEncoding.EncodeToString(newValue)
			newValue = []byte("rot_" + encoded[:40]) // Prefix to indicate rotated key
		}
	default:
		return fmt.Errorf("unsupported secret type for rotation: %s", existing.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to generate new secret value: %w", err)
	}

	// Create new secret with updated timestamp
	newSecret := &Secret{
		ID:             existing.ID,
		Value:          newValue,
		CreatedAt:      time.Now(),
		ExpiresAt:      existing.ExpiresAt,
		RotationPeriod: existing.RotationPeriod,
		Type:           existing.Type,
		Metadata:       existing.Metadata,
	}

	// Update metadata
	if newSecret.Metadata == nil {
		newSecret.Metadata = make(map[string]string)
	}
	newSecret.Metadata["rotated_at"] = time.Now().Format(time.RFC3339)
	newSecret.Metadata["rotation_count"] = fmt.Sprintf("%d", getRotationCount(existing.Metadata)+1)

	sm.secrets[secretID] = newSecret

	// Schedule next rotation
	if newSecret.RotationPeriod > 0 {
		sm.scheduleRotation(secretID, newSecret.RotationPeriod)
	}

	return nil
}

// GetSecret retrieves a secret by ID
func (sm *SecretManager) GetSecret(secretID string) (*Secret, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	secret, exists := sm.secrets[secretID]
	if !exists {
		return nil, fmt.Errorf("secret %s not found", secretID)
	}

	// Check if secret has expired
	if secret.ExpiresAt != nil && time.Now().After(*secret.ExpiresAt) {
		return nil, fmt.Errorf("secret %s has expired", secretID)
	}

	return secret, nil
}

// CompareSecrets securely compares two secrets using constant-time comparison
func (sm *SecretManager) CompareSecrets(secret1, secret2 []byte) bool {
	return subtle.ConstantTimeCompare(secret1, secret2) == 1
}

// scheduleRotation schedules automatic rotation for a secret
func (sm *SecretManager) scheduleRotation(secretID string, period time.Duration) {
	sm.rotationLock.Lock()
	defer sm.rotationLock.Unlock()

	nextRotation := time.Now().Add(period)
	sm.keyRotation[secretID] = nextRotation

	// Start background rotation goroutine
	go func() {
		timer := time.NewTimer(period)
		defer timer.Stop()

		<-timer.C

		// Perform rotation
		if err := sm.RotateSecret(secretID); err != nil {
			// Log error (in production, send to monitoring system)
			fmt.Printf("Failed to rotate secret %s: %v\n", secretID, err)
		}
	}()
}

// GetRotationStatus returns rotation status for all secrets
func (sm *SecretManager) GetRotationStatus() map[string]time.Time {
	sm.rotationLock.Lock()
	defer sm.rotationLock.Unlock()

	status := make(map[string]time.Time)
	for secretID, nextRotation := range sm.keyRotation {
		status[secretID] = nextRotation
	}
	return status
}

// GenerateSecureRandomString generates a cryptographically secure random string
func GenerateSecureRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random string: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// ValidateSecretStrength validates the strength of a secret
func ValidateSecretStrength(secret string, minLength int) error {
	if len(secret) < minLength {
		return fmt.Errorf("secret must be at least %d characters long", minLength)
	}

	// Check for common weak patterns
	weakPatterns := []string{
		"password", "secret", "default", "admin", "root", "test",
		"123456", "qwerty", "abc123", "password123",
	}

	lowerSecret := strings.ToLower(secret)
	for _, pattern := range weakPatterns {
		if strings.Contains(lowerSecret, pattern) {
			return fmt.Errorf("secret contains weak pattern: %s", pattern)
		}
	}

	return nil
}

// Helper function to get rotation count from metadata
func getRotationCount(metadata map[string]string) int {
	if metadata == nil {
		return 0
	}

	countStr, exists := metadata["rotation_count"]
	if !exists {
		return 0
	}

	// Simple conversion, ignore errors and default to 0
	var count int
	fmt.Sscanf(countStr, "%d", &count)
	return count
}

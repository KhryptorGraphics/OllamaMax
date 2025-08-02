package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// AdvancedEncryptionManager manages advanced encryption and key rotation
type AdvancedEncryptionManager struct {
	config     *EncryptionConfig
	keyManager *EncryptionKeyManager
	mu         sync.RWMutex
}

// EncryptionConfig configures advanced encryption
type EncryptionConfig struct {
	Algorithm           string        `json:"algorithm"`
	KeySize             int           `json:"key_size"`
	RotationInterval    time.Duration `json:"rotation_interval"`
	EnableE2E           bool          `json:"enable_e2e"`
	EnableAtRest        bool          `json:"enable_at_rest"`
	EnableInTransit     bool          `json:"enable_in_transit"`
	KeyDerivationRounds int           `json:"key_derivation_rounds"`
	CompressionEnabled  bool          `json:"compression_enabled"`
}

// EncryptionKeyManager manages encryption keys and rotation
type EncryptionKeyManager struct {
	currentKey    *EncryptionKey
	previousKeys  map[string]*EncryptionKey
	keyRotationCh chan struct{}
	config        *EncryptionConfig
	mu            sync.RWMutex
}

// EncryptionKey represents an encryption key with metadata
type EncryptionKey struct {
	ID        string    `json:"id"`
	Key       []byte    `json:"-"` // Never serialize the actual key
	Algorithm string    `json:"algorithm"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Version   int       `json:"version"`
	Active    bool      `json:"active"`
}

// EncryptedData represents encrypted data with metadata
type EncryptedData struct {
	Data      []byte            `json:"data"`
	KeyID     string            `json:"key_id"`
	Algorithm string            `json:"algorithm"`
	Nonce     []byte            `json:"nonce"`
	Metadata  map[string]string `json:"metadata"`
	Timestamp time.Time         `json:"timestamp"`
}

// E2EEncryptionContext represents end-to-end encryption context
type E2EEncryptionContext struct {
	SenderID   string    `json:"sender_id"`
	ReceiverID string    `json:"receiver_id"`
	SessionKey []byte    `json:"-"`
	PublicKey  []byte    `json:"public_key"`
	PrivateKey []byte    `json:"-"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// NewAdvancedEncryptionManager creates a new advanced encryption manager
func NewAdvancedEncryptionManager(config *EncryptionConfig) *AdvancedEncryptionManager {
	if config == nil {
		config = DefaultEncryptionConfig()
	}

	keyManager := &EncryptionKeyManager{
		previousKeys:  make(map[string]*EncryptionKey),
		keyRotationCh: make(chan struct{}, 1),
		config:        config,
	}

	aem := &AdvancedEncryptionManager{
		config:     config,
		keyManager: keyManager,
	}

	// Initialize with first key
	if err := aem.rotateKey(); err != nil {
		log.Error().Err(err).Msg("Failed to initialize encryption key")
	}

	// Start key rotation if enabled
	if config.RotationInterval > 0 {
		go aem.startKeyRotation()
	}

	return aem
}

// Encrypt encrypts data using the current encryption key
func (aem *AdvancedEncryptionManager) Encrypt(data []byte) (*EncryptedData, error) {
	aem.mu.RLock()
	defer aem.mu.RUnlock()

	if aem.keyManager.currentKey == nil {
		return nil, fmt.Errorf("no encryption key available")
	}

	key := aem.keyManager.currentKey

	switch aem.config.Algorithm {
	case "AES-256-GCM":
		return aem.encryptAESGCM(data, key)
	case "ChaCha20-Poly1305":
		return aem.encryptChaCha20Poly1305(data, key)
	default:
		return nil, fmt.Errorf("unsupported encryption algorithm: %s", aem.config.Algorithm)
	}
}

// Decrypt decrypts data using the appropriate key
func (aem *AdvancedEncryptionManager) Decrypt(encryptedData *EncryptedData) ([]byte, error) {
	aem.mu.RLock()
	defer aem.mu.RUnlock()

	// Find the appropriate key
	var key *EncryptionKey
	if aem.keyManager.currentKey != nil && aem.keyManager.currentKey.ID == encryptedData.KeyID {
		key = aem.keyManager.currentKey
	} else if prevKey, exists := aem.keyManager.previousKeys[encryptedData.KeyID]; exists {
		key = prevKey
	} else {
		return nil, fmt.Errorf("encryption key not found: %s", encryptedData.KeyID)
	}

	switch encryptedData.Algorithm {
	case "AES-256-GCM":
		return aem.decryptAESGCM(encryptedData, key)
	case "ChaCha20-Poly1305":
		return aem.decryptChaCha20Poly1305(encryptedData, key)
	default:
		return nil, fmt.Errorf("unsupported encryption algorithm: %s", encryptedData.Algorithm)
	}
}

// encryptAESGCM encrypts data using AES-256-GCM
func (aem *AdvancedEncryptionManager) encryptAESGCM(data []byte, key *EncryptionKey) (*EncryptedData, error) {
	block, err := aes.NewCipher(key.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, data, nil)

	return &EncryptedData{
		Data:      ciphertext,
		KeyID:     key.ID,
		Algorithm: "AES-256-GCM",
		Nonce:     nonce,
		Timestamp: time.Now(),
		Metadata:  make(map[string]string),
	}, nil
}

// decryptAESGCM decrypts data using AES-256-GCM
func (aem *AdvancedEncryptionManager) decryptAESGCM(encryptedData *EncryptedData, key *EncryptionKey) ([]byte, error) {
	block, err := aes.NewCipher(key.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, encryptedData.Nonce, encryptedData.Data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// encryptChaCha20Poly1305 encrypts data using ChaCha20-Poly1305
func (aem *AdvancedEncryptionManager) encryptChaCha20Poly1305(data []byte, key *EncryptionKey) (*EncryptedData, error) {
	// Implementation would use golang.org/x/crypto/chacha20poly1305
	// For now, fallback to AES-GCM
	return aem.encryptAESGCM(data, key)
}

// decryptChaCha20Poly1305 decrypts data using ChaCha20-Poly1305
func (aem *AdvancedEncryptionManager) decryptChaCha20Poly1305(encryptedData *EncryptedData, key *EncryptionKey) ([]byte, error) {
	// Implementation would use golang.org/x/crypto/chacha20poly1305
	// For now, fallback to AES-GCM
	return aem.decryptAESGCM(encryptedData, key)
}

// CreateE2EContext creates an end-to-end encryption context
func (aem *AdvancedEncryptionManager) CreateE2EContext(senderID, receiverID string) (*E2EEncryptionContext, error) {
	// Generate RSA key pair for E2E encryption
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	publicKey := &privateKey.PublicKey

	// Serialize keys
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	// Generate session key
	sessionKey := make([]byte, 32) // 256-bit key
	if _, err := rand.Read(sessionKey); err != nil {
		return nil, fmt.Errorf("failed to generate session key: %w", err)
	}

	return &E2EEncryptionContext{
		SenderID:   senderID,
		ReceiverID: receiverID,
		SessionKey: sessionKey,
		PublicKey:  publicKeyBytes,
		PrivateKey: privateKeyBytes,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(24 * time.Hour), // 24-hour session
	}, nil
}

// EncryptE2E encrypts data for end-to-end encryption
func (aem *AdvancedEncryptionManager) EncryptE2E(data []byte, context *E2EEncryptionContext) (*EncryptedData, error) {
	// Use session key for symmetric encryption
	tempKey := &EncryptionKey{
		ID:        context.SenderID + "-" + context.ReceiverID,
		Key:       context.SessionKey,
		Algorithm: aem.config.Algorithm,
		CreatedAt: context.CreatedAt,
		ExpiresAt: context.ExpiresAt,
		Active:    true,
	}

	return aem.encryptAESGCM(data, tempKey)
}

// DecryptE2E decrypts data from end-to-end encryption
func (aem *AdvancedEncryptionManager) DecryptE2E(encryptedData *EncryptedData, context *E2EEncryptionContext) ([]byte, error) {
	// Use session key for symmetric decryption
	tempKey := &EncryptionKey{
		ID:        context.SenderID + "-" + context.ReceiverID,
		Key:       context.SessionKey,
		Algorithm: encryptedData.Algorithm,
		CreatedAt: context.CreatedAt,
		ExpiresAt: context.ExpiresAt,
		Active:    true,
	}

	return aem.decryptAESGCM(encryptedData, tempKey)
}

// rotateKey rotates the encryption key
func (aem *AdvancedEncryptionManager) rotateKey() error {
	aem.mu.Lock()
	defer aem.mu.Unlock()

	// Generate new key
	newKey := make([]byte, aem.config.KeySize/8) // Convert bits to bytes
	if _, err := rand.Read(newKey); err != nil {
		return fmt.Errorf("failed to generate new key: %w", err)
	}

	// Create key metadata
	keyID := fmt.Sprintf("key-%d", time.Now().Unix())
	encryptionKey := &EncryptionKey{
		ID:        keyID,
		Key:       newKey,
		Algorithm: aem.config.Algorithm,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(aem.config.RotationInterval),
		Version:   1,
		Active:    true,
	}

	// Store previous key if exists
	if aem.keyManager.currentKey != nil {
		aem.keyManager.currentKey.Active = false
		aem.keyManager.previousKeys[aem.keyManager.currentKey.ID] = aem.keyManager.currentKey
	}

	// Set new current key
	aem.keyManager.currentKey = encryptionKey

	log.Info().
		Str("key_id", keyID).
		Str("algorithm", aem.config.Algorithm).
		Msg("Encryption key rotated")

	return nil
}

// startKeyRotation starts automatic key rotation
func (aem *AdvancedEncryptionManager) startKeyRotation() {
	ticker := time.NewTicker(aem.config.RotationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := aem.rotateKey(); err != nil {
				log.Error().Err(err).Msg("Failed to rotate encryption key")
			}
		case <-aem.keyManager.keyRotationCh:
			return
		}
	}
}

// GetCurrentKeyInfo returns information about the current key
func (aem *AdvancedEncryptionManager) GetCurrentKeyInfo() *EncryptionKey {
	aem.mu.RLock()
	defer aem.mu.RUnlock()

	if aem.keyManager.currentKey == nil {
		return nil
	}

	// Return copy without the actual key
	return &EncryptionKey{
		ID:        aem.keyManager.currentKey.ID,
		Algorithm: aem.keyManager.currentKey.Algorithm,
		CreatedAt: aem.keyManager.currentKey.CreatedAt,
		ExpiresAt: aem.keyManager.currentKey.ExpiresAt,
		Version:   aem.keyManager.currentKey.Version,
		Active:    aem.keyManager.currentKey.Active,
	}
}

// GetKeyStats returns statistics about key usage
func (aem *AdvancedEncryptionManager) GetKeyStats() map[string]interface{} {
	aem.mu.RLock()
	defer aem.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["current_key_id"] = ""
	stats["total_keys"] = len(aem.keyManager.previousKeys)
	stats["algorithm"] = aem.config.Algorithm
	stats["key_size"] = aem.config.KeySize
	stats["rotation_interval"] = aem.config.RotationInterval.String()

	if aem.keyManager.currentKey != nil {
		stats["current_key_id"] = aem.keyManager.currentKey.ID
		stats["current_key_age"] = time.Since(aem.keyManager.currentKey.CreatedAt).String()
		stats["total_keys"] = len(aem.keyManager.previousKeys) + 1
	}

	return stats
}

// Shutdown gracefully shuts down the encryption manager
func (aem *AdvancedEncryptionManager) Shutdown() error {
	close(aem.keyManager.keyRotationCh)

	// Clear sensitive data
	aem.mu.Lock()
	defer aem.mu.Unlock()

	if aem.keyManager.currentKey != nil {
		// Zero out the key
		for i := range aem.keyManager.currentKey.Key {
			aem.keyManager.currentKey.Key[i] = 0
		}
	}

	for _, key := range aem.keyManager.previousKeys {
		for i := range key.Key {
			key.Key[i] = 0
		}
	}

	log.Info().Msg("Advanced encryption manager shut down")
	return nil
}

// DefaultEncryptionConfig returns default encryption configuration
func DefaultEncryptionConfig() *EncryptionConfig {
	return &EncryptionConfig{
		Algorithm:           "AES-256-GCM",
		KeySize:             256,
		RotationInterval:    24 * time.Hour,
		EnableE2E:           true,
		EnableAtRest:        true,
		EnableInTransit:     true,
		KeyDerivationRounds: 100000,
		CompressionEnabled:  true,
	}
}

// EncryptString encrypts a string and returns base64 encoded result
func (aem *AdvancedEncryptionManager) EncryptString(plaintext string) (string, error) {
	encrypted, err := aem.Encrypt([]byte(plaintext))
	if err != nil {
		return "", err
	}

	// Serialize encrypted data to JSON and encode as base64
	data := fmt.Sprintf("%s:%s:%s",
		base64.StdEncoding.EncodeToString(encrypted.Data),
		encrypted.KeyID,
		base64.StdEncoding.EncodeToString(encrypted.Nonce))

	return base64.StdEncoding.EncodeToString([]byte(data)), nil
}

// DecryptString decrypts a base64 encoded string
func (aem *AdvancedEncryptionManager) DecryptString(ciphertext string) (string, error) {
	// Decode base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Parse the data format
	parts := strings.Split(string(data), ":")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid encrypted data format")
	}

	encryptedBytes, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted data: %w", err)
	}

	nonce, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		return "", fmt.Errorf("failed to decode nonce: %w", err)
	}

	encryptedData := &EncryptedData{
		Data:      encryptedBytes,
		KeyID:     parts[1],
		Algorithm: aem.config.Algorithm,
		Nonce:     nonce,
	}

	plaintext, err := aem.Decrypt(encryptedData)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

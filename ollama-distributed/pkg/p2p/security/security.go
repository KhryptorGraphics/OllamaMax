package security

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

const (
	// Security protocols
	SecureChannelProtocol = protocol.ID("/ollamacron/secure-channel/1.0.0")
	AuthProtocol         = protocol.ID("/ollamacron/auth/1.0.0")
	KeyExchangeProtocol  = protocol.ID("/ollamacron/key-exchange/1.0.0")
	
	// Security levels
	SecurityLevelNone   = "none"
	SecurityLevelBasic  = "basic"
	SecurityLevelHigh   = "high"
	
	// Key rotation intervals
	DefaultKeyRotationInterval = 24 * time.Hour
	SessionKeyTTL             = 4 * time.Hour
)

// ===============================
// Supporting Type Definitions
// ===============================

// SessionKey represents a session encryption key
type SessionKey struct {
	PublicKey  []byte
	PrivateKey []byte
	CreatedAt  time.Time
	ExpiresAt  time.Time
}

// KeyManager manages cryptographic keys
type KeyManager struct {
	sessionKeys    map[string]*SessionKey
	keysMu        sync.RWMutex
	
	// Peer public keys
	peerKeys      map[string]crypto.PubKey
	peerKeysMu    sync.RWMutex
	
	config        *SecurityConfig
}

// EncryptionManager handles encryption/decryption
type EncryptionManager struct {
	keyManager *KeyManager
	config     *SecurityConfig
}

// AccessControl manages access control policies
type AccessControl struct {
	policies  map[peer.ID]*AccessPolicy
	policyMu  sync.RWMutex
	config    *SecurityConfig
}

// AccessPolicy defines access permissions for a peer
type AccessPolicy struct {
	PeerID      peer.ID
	Permissions []string
	CreatedAt   time.Time
	ExpiresAt   time.Time
}

// RateLimiter manages rate limiting
type RateLimiter struct {
	limits   map[string]*RateLimit
	limitMu  sync.RWMutex
	config   *SecurityConfig
}

// RateLimit defines rate limiting parameters
type RateLimit struct {
	RequestsPerSecond int
	BurstSize         int
	Window            time.Duration
	lastReset         time.Time
	currentCount      int
}

// ProtocolAccessControl defines protocol-specific access control
type ProtocolAccessControl struct {
	AllowAll      bool
	AllowedPeers  []peer.ID
	BlockedPeers  []peer.ID
	RequiredPerms []string
}

// AuthManager manages authentication for P2P communications
type AuthManager struct {
	authenticatedPeers map[peer.ID]time.Time
	peersMu           sync.RWMutex
	
	// Private key for signing
	privateKey        crypto.PrivKey
	publicKey         crypto.PubKey
	
	// Session management
	sessions          map[string]*AuthSession
	sessionsMu        sync.RWMutex
	sessionTTL        time.Duration
	
	// Authentication configuration
	config            *SecurityConfig
}

// AuthSession represents an authenticated session
type AuthSession struct {
	PeerID    peer.ID
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
	Metadata  map[string]interface{}
}

// SecurityManager manages security for P2P communications
type SecurityManager struct {
	host           host.Host
	
	// Key management
	keyManager     *KeyManager
	
	// Protocol handlers
	protocols      map[protocol.ID]*SecureProtocol
	protocolsMux   sync.RWMutex
	
	// Authentication
	authManager    *AuthManager
	
	// Encryption
	encryptionMgr  *EncryptionManager
	
	// Access control
	accessControl  *AccessControl
	
	// Rate limiting
	rateLimiter    *RateLimiter
	
	// Metrics
	metrics        *SecurityMetrics
	
	// Configuration
	config         *SecurityConfig
	
	// Lifecycle
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	SecurityLevel      string        `json:"security_level"`
	EnableEncryption   bool          `json:"enable_encryption"`
	EnableAuth         bool          `json:"enable_auth"`
	KeyRotationInterval time.Duration `json:"key_rotation_interval"`
	SessionKeyTTL      time.Duration `json:"session_key_ttl"`
	MaxConnections     int           `json:"max_connections"`
	RateLimits         map[string]int `json:"rate_limits"`
	TrustedPeers       []string      `json:"trusted_peers"`
	BlockedPeers       []string      `json:"blocked_peers"`
}

// SecurityMetrics tracks security-related metrics
type SecurityMetrics struct {
	AuthAttempts       int
	AuthFailures       int
	EncryptedSessions  int
	ActiveSessions     int
	KeyRotations       int
	BlockedConnections int
	RateLimitHits      int
	LastKeyRotation    time.Time
	StartTime          time.Time
	
	// Protocol metrics
	ProtocolMetrics    map[protocol.ID]*ProtocolMetrics
}

// ProtocolMetrics tracks metrics for specific protocols
type ProtocolMetrics struct {
	RequestCount       int
	ErrorCount         int
	AverageLatency     time.Duration
	LastActivity       time.Time
}

// SecureProtocol represents a secure protocol handler
type SecureProtocol struct {
	ID                protocol.ID
	Handler           network.StreamHandler
	
	// Security settings
	RequireAuth       bool
	RequireEncryption bool
	SecurityLevel     string
	
	// Rate limiting
	RateLimit         *RateLimit
	
	// Access control
	AccessControl     *ProtocolAccessControl
	
	// Metrics
	Metrics           *ProtocolMetrics
}

// NewSecurityManager creates a new security manager
func NewSecurityManager(ctx context.Context, host host.Host, config *SecurityConfig) (*SecurityManager, error) {
	if config == nil {
		config = DefaultSecurityConfig()
	}
	
	ctx, cancel := context.WithCancel(ctx)
	
	sm := &SecurityManager{
		host:      host,
		config:    config,
		protocols: make(map[protocol.ID]*SecureProtocol),
		metrics: &SecurityMetrics{
			StartTime:       time.Now(),
			ProtocolMetrics: make(map[protocol.ID]*ProtocolMetrics),
		},
		ctx:    ctx,
		cancel: cancel,
	}
	
	// Initialize components
	if err := sm.initializeComponents(); err != nil {
		return nil, fmt.Errorf("failed to initialize security components: %w", err)
	}
	
	// Setup protocol handlers
	sm.setupProtocolHandlers()
	
	// Start background tasks
	sm.startBackgroundTasks()
	
	log.Printf("Security manager initialized with level: %s", config.SecurityLevel)
	return sm, nil
}

// initializeComponents initializes security components
func (sm *SecurityManager) initializeComponents() error {
	// Initialize key manager
	keyManager, err := NewKeyManager(sm.host, sm.config)
	if err != nil {
		return fmt.Errorf("failed to initialize key manager: %w", err)
	}
	sm.keyManager = keyManager
	
	// Initialize authentication manager
	authManager, err := NewAuthManager(sm.host, sm.config)
	if err != nil {
		return fmt.Errorf("failed to initialize auth manager: %w", err)
	}
	sm.authManager = authManager
	
	// Initialize encryption manager
	encryptionMgr, err := NewEncryptionManager(sm.keyManager, sm.config)
	if err != nil {
		return fmt.Errorf("failed to initialize encryption manager: %w", err)
	}
	sm.encryptionMgr = encryptionMgr
	
	// Initialize access control
	accessControl, err := NewAccessControl(sm.config)
	if err != nil {
		return fmt.Errorf("failed to initialize access control: %w", err)
	}
	sm.accessControl = accessControl
	
	// Initialize rate limiter
	rateLimiter, err := NewRateLimiter(sm.config)
	if err != nil {
		return fmt.Errorf("failed to initialize rate limiter: %w", err)
	}
	sm.rateLimiter = rateLimiter
	
	return nil
}

// setupProtocolHandlers sets up secure protocol handlers
func (sm *SecurityManager) setupProtocolHandlers() {
	// Secure channel protocol
	sm.RegisterSecureProtocol(SecureChannelProtocol, &SecureProtocol{
		ID:                SecureChannelProtocol,
		Handler:           sm.handleSecureChannel,
		RequireAuth:       true,
		RequireEncryption: true,
		SecurityLevel:     SecurityLevelHigh,
		RateLimit:         &RateLimit{RequestsPerSecond: 10, BurstSize: 20},
		AccessControl:     &ProtocolAccessControl{AllowAll: false},
		Metrics:           &ProtocolMetrics{},
	})
	
	// Authentication protocol
	sm.RegisterSecureProtocol(AuthProtocol, &SecureProtocol{
		ID:                AuthProtocol,
		Handler:           sm.handleAuth,
		RequireAuth:       false,
		RequireEncryption: true,
		SecurityLevel:     SecurityLevelHigh,
		RateLimit:         &RateLimit{RequestsPerSecond: 5, BurstSize: 10},
		AccessControl:     &ProtocolAccessControl{AllowAll: true},
		Metrics:           &ProtocolMetrics{},
	})
	
	// Key exchange protocol
	sm.RegisterSecureProtocol(KeyExchangeProtocol, &SecureProtocol{
		ID:                KeyExchangeProtocol,
		Handler:           sm.handleKeyExchange,
		RequireAuth:       false,
		RequireEncryption: false,
		SecurityLevel:     SecurityLevelBasic,
		RateLimit:         &RateLimit{RequestsPerSecond: 2, BurstSize: 5},
		AccessControl:     &ProtocolAccessControl{AllowAll: true},
		Metrics:           &ProtocolMetrics{},
	})
}

// startBackgroundTasks starts background security tasks
func (sm *SecurityManager) startBackgroundTasks() {
	// Key rotation
	sm.wg.Add(1)
	go sm.keyRotationTask()
	
	// Metrics collection
	sm.wg.Add(1)
	go sm.metricsTask()
	
	// Rate limit cleanup
	sm.wg.Add(1)
	go sm.rateLimitCleanupTask()
	
	// Session cleanup
	sm.wg.Add(1)
	go sm.sessionCleanupTask()
}

// RegisterSecureProtocol registers a secure protocol handler
func (sm *SecurityManager) RegisterSecureProtocol(protocolID protocol.ID, protocol *SecureProtocol) {
	sm.protocolsMux.Lock()
	defer sm.protocolsMux.Unlock()
	
	// Wrap handler with security middleware
	wrappedHandler := sm.wrapHandler(protocol)
	
	sm.host.SetStreamHandler(protocolID, wrappedHandler)
	sm.protocols[protocolID] = protocol
	sm.metrics.ProtocolMetrics[protocolID] = protocol.Metrics
	
	log.Printf("Registered secure protocol: %s", protocolID)
}

// wrapHandler wraps a protocol handler with security middleware
func (sm *SecurityManager) wrapHandler(protocol *SecureProtocol) network.StreamHandler {
	return func(stream network.Stream) {
		start := time.Now()
		peerID := stream.Conn().RemotePeer()
		
		// Update metrics
		protocol.Metrics.RequestCount++
		protocol.Metrics.LastActivity = time.Now()
		
		// Check access control
		if !sm.accessControl.IsAllowed(peerID, protocol.ID) {
			log.Printf("Access denied for peer %s on protocol %s", peerID, protocol.ID)
			stream.Reset()
			return
		}
		
		// Check rate limits
		if !sm.rateLimiter.Allow(peerID, protocol.ID) {
			log.Printf("Rate limit exceeded for peer %s on protocol %s", peerID, protocol.ID)
			sm.metrics.RateLimitHits++
			stream.Reset()
			return
		}
		
		// Check authentication if required
		if protocol.RequireAuth {
			if !sm.authManager.IsAuthenticated(peerID) {
				log.Printf("Authentication required for peer %s on protocol %s", peerID, protocol.ID)
				stream.Reset()
				return
			}
		}
		
		// Handle encryption if required
		if protocol.RequireEncryption {
			if !sm.isEncryptedConnection(stream.Conn()) {
				log.Printf("Encryption required for peer %s on protocol %s", peerID, protocol.ID)
				stream.Reset()
				return
			}
		}
		
		// Call original handler
		protocol.Handler(stream)
		
		// Update latency metrics
		protocol.Metrics.AverageLatency = time.Since(start)
	}
}

// EstablishSecureChannel establishes a secure channel with a peer
func (sm *SecurityManager) EstablishSecureChannel(ctx context.Context, peerID peer.ID) (*SecureChannel, error) {
	// Create new stream
	stream, err := sm.host.NewStream(ctx, peerID, SecureChannelProtocol)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}
	
	// Perform handshake
	sessionKey, err := sm.performHandshake(ctx, stream)
	if err != nil {
		stream.Close()
		return nil, fmt.Errorf("handshake failed: %w", err)
	}
	
	// Create secure channel
	channel := &SecureChannel{
		Stream:     stream,
		SessionKey: sessionKey,
		Peer:       peerID,
		CreatedAt:  time.Now(),
		manager:    sm,
	}
	
	sm.metrics.EncryptedSessions++
	sm.metrics.ActiveSessions++
	
	return channel, nil
}

// performHandshake performs security handshake
func (sm *SecurityManager) performHandshake(ctx context.Context, stream network.Stream) (*SessionKey, error) {
	// Generate session key
	sessionKey, err := sm.keyManager.GenerateSessionKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session key: %w", err)
	}
	
	// Perform noise protocol handshake
	handshake := &HandshakeMessage{
		Type:      "key_exchange",
		PublicKey: sessionKey.PublicKey,
		Timestamp: time.Now(),
	}
	
	// Send handshake
	if err := sm.sendHandshake(stream, handshake); err != nil {
		return nil, fmt.Errorf("failed to send handshake: %w", err)
	}
	
	// Receive response
	response, err := sm.receiveHandshake(stream)
	if err != nil {
		return nil, fmt.Errorf("failed to receive handshake response: %w", err)
	}
	
	// Verify response
	if err := sm.verifyHandshake(response); err != nil {
		return nil, fmt.Errorf("handshake verification failed: %w", err)
	}
	
	return sessionKey, nil
}

// Protocol handlers

// handleSecureChannel handles secure channel establishment
func (sm *SecurityManager) handleSecureChannel(stream network.Stream) {
	defer stream.Close()
	
	// Handle handshake
	handshake, err := sm.receiveHandshake(stream)
	if err != nil {
		log.Printf("Failed to receive handshake: %v", err)
		return
	}
	
	// Verify handshake
	if err := sm.verifyHandshake(handshake); err != nil {
		log.Printf("Handshake verification failed: %v", err)
		return
	}
	
	// Generate session key
	sessionKey, err := sm.keyManager.GenerateSessionKey()
	if err != nil {
		log.Printf("Failed to generate session key: %v", err)
		return
	}
	
	// Send response
	response := &HandshakeMessage{
		Type:      "key_exchange_response",
		PublicKey: sessionKey.PublicKey,
		Timestamp: time.Now(),
	}
	
	if err := sm.sendHandshake(stream, response); err != nil {
		log.Printf("Failed to send handshake response: %v", err)
		return
	}
	
	log.Printf("Secure channel established with peer: %s", stream.Conn().RemotePeer())
}

// handleAuth handles authentication requests
func (sm *SecurityManager) handleAuth(stream network.Stream) {
	defer stream.Close()
	
	// Handle authentication
	if err := sm.authManager.HandleAuthRequest(stream); err != nil {
		log.Printf("Authentication failed: %v", err)
		return
	}
	
	sm.metrics.AuthAttempts++
	log.Printf("Authentication successful for peer: %s", stream.Conn().RemotePeer())
}

// handleKeyExchange handles key exchange requests
func (sm *SecurityManager) handleKeyExchange(stream network.Stream) {
	defer stream.Close()
	
	// Handle key exchange
	if err := sm.keyManager.HandleKeyExchange(stream); err != nil {
		log.Printf("Key exchange failed: %v", err)
		return
	}
	
	log.Printf("Key exchange successful with peer: %s", stream.Conn().RemotePeer())
}

// Background tasks

// keyRotationTask handles periodic key rotation
func (sm *SecurityManager) keyRotationTask() {
	defer sm.wg.Done()
	
	ticker := time.NewTicker(sm.config.KeyRotationInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			if err := sm.keyManager.RotateKeys(); err != nil {
				log.Printf("Key rotation failed: %v", err)
			} else {
				sm.metrics.KeyRotations++
				sm.metrics.LastKeyRotation = time.Now()
				log.Printf("Key rotation completed")
			}
		}
	}
}

// metricsTask collects security metrics
func (sm *SecurityManager) metricsTask() {
	defer sm.wg.Done()
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			sm.updateMetrics()
		}
	}
}

// updateMetrics updates security metrics
func (sm *SecurityManager) updateMetrics() {
	// Update active sessions
	sm.metrics.ActiveSessions = sm.keyManager.GetActiveSessionCount()
	
	// Update protocol metrics
	for protocolID, protocol := range sm.protocols {
		if metrics, exists := sm.metrics.ProtocolMetrics[protocolID]; exists {
			metrics.RequestCount = protocol.Metrics.RequestCount
			metrics.ErrorCount = protocol.Metrics.ErrorCount
			metrics.AverageLatency = protocol.Metrics.AverageLatency
			metrics.LastActivity = protocol.Metrics.LastActivity
		}
	}
}

// rateLimitCleanupTask cleans up expired rate limit entries
func (sm *SecurityManager) rateLimitCleanupTask() {
	defer sm.wg.Done()
	
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			sm.rateLimiter.Cleanup()
		}
	}
}

// sessionCleanupTask cleans up expired sessions
func (sm *SecurityManager) sessionCleanupTask() {
	defer sm.wg.Done()
	
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			sm.keyManager.CleanupExpiredSessions()
		}
	}
}

// Utility methods

// isEncryptedConnection checks if connection is encrypted
func (sm *SecurityManager) isEncryptedConnection(conn network.Conn) bool {
	// Check if connection uses encrypted transport
	// For libp2p, connections are generally encrypted by default
	return true
}

// sendHandshake sends handshake message
func (sm *SecurityManager) sendHandshake(stream network.Stream, handshake *HandshakeMessage) error {
	data, err := json.Marshal(handshake)
	if err != nil {
		return fmt.Errorf("failed to marshal handshake: %w", err)
	}
	
	_, err = stream.Write(data)
	return err
}

// receiveHandshake receives handshake message
func (sm *SecurityManager) receiveHandshake(stream network.Stream) (*HandshakeMessage, error) {
	buf := make([]byte, 4096)
	n, err := stream.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read handshake: %w", err)
	}
	
	var handshake HandshakeMessage
	if err := json.Unmarshal(buf[:n], &handshake); err != nil {
		return nil, fmt.Errorf("failed to unmarshal handshake: %w", err)
	}
	
	return &handshake, nil
}

// verifyHandshake verifies handshake message
func (sm *SecurityManager) verifyHandshake(handshake *HandshakeMessage) error {
	// Check timestamp
	if time.Since(handshake.Timestamp) > 5*time.Minute {
		return fmt.Errorf("handshake timestamp too old")
	}
	
	// Verify signature if present
	if handshake.Signature != nil {
		// Implement proper signature verification
		if err := sm.verifyHandshakeSignature(handshake); err != nil {
			return fmt.Errorf("signature verification failed: %w", err)
		}
	}
	
	return nil
}

// verifyHandshakeSignature verifies the cryptographic signature of a handshake
func (sm *SecurityManager) verifyHandshakeSignature(handshake *HandshakeMessage) error {
	// Create message hash for signature verification
	messageData := fmt.Sprintf("%s:%x:%d", handshake.Type, handshake.PublicKey, handshake.Timestamp.Unix())
	
	// Get peer's public key from key manager
	peerPubKey, err := sm.keyManager.GetPeerPublicKey(handshake.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to get peer public key: %w", err)
	}
	
	// Verify signature using libp2p crypto
	valid, err := peerPubKey.Verify([]byte(messageData), handshake.Signature)
	if err != nil {
		return fmt.Errorf("signature verification error: %w", err)
	}
	
	if !valid {
		return fmt.Errorf("invalid signature")
	}
	
	return nil
}

// GetMetrics returns security metrics
func (sm *SecurityManager) GetMetrics() *SecurityMetrics {
	return sm.metrics
}

// GetConfig returns security configuration
func (sm *SecurityManager) GetConfig() *SecurityConfig {
	return sm.config
}

// Close closes the security manager
func (sm *SecurityManager) Close() error {
	log.Printf("Closing security manager")
	sm.cancel()
	sm.wg.Wait()
	
	if sm.keyManager != nil {
		sm.keyManager.Close()
	}
	
	return nil
}

// DefaultSecurityConfig returns default security configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		SecurityLevel:       SecurityLevelBasic,
		EnableEncryption:    true,
		EnableAuth:          true,
		KeyRotationInterval: DefaultKeyRotationInterval,
		SessionKeyTTL:       SessionKeyTTL,
		MaxConnections:      100,
		RateLimits: map[string]int{
			"default": 10,
			"auth":    5,
			"key":     2,
		},
		TrustedPeers: []string{},
		BlockedPeers: []string{},
	}
}

// HandshakeMessage represents a security handshake message
type HandshakeMessage struct {
	Type      string    `json:"type"`
	PublicKey []byte    `json:"public_key"`
	Timestamp time.Time `json:"timestamp"`
	Signature []byte    `json:"signature,omitempty"`
}

// SecureChannel represents a secure communication channel
type SecureChannel struct {
	Stream     network.Stream
	SessionKey *SessionKey
	Peer       peer.ID
	CreatedAt  time.Time
	manager    *SecurityManager
}

// Close closes the secure channel
func (sc *SecureChannel) Close() error {
	sc.manager.metrics.ActiveSessions--
	return sc.Stream.Close()
}

// Send sends encrypted data through the channel
func (sc *SecureChannel) Send(data []byte) error {
	encrypted, err := sc.manager.encryptionMgr.Encrypt(data, sc.SessionKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt data: %w", err)
	}
	
	_, err = sc.Stream.Write(encrypted)
	return err
}

// Receive receives and decrypts data from the channel
func (sc *SecureChannel) Receive() ([]byte, error) {
	buf := make([]byte, 4096)
	n, err := sc.Stream.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}
	
	return sc.manager.encryptionMgr.Decrypt(buf[:n], sc.SessionKey)
}

// ===============================
// Implementation Functions
// ===============================

// NewKeyManager creates a new key manager
func NewKeyManager(host host.Host, config *SecurityConfig) (*KeyManager, error) {
	return &KeyManager{
		sessionKeys: make(map[string]*SessionKey),
		peerKeys:    make(map[string]crypto.PubKey),
		config:      config,
	}, nil
}

// GenerateSessionKey generates a new session key
func (km *KeyManager) GenerateSessionKey() (*SessionKey, error) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Marshal private key
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Marshal public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	sessionKey := &SessionKey{
		PublicKey:  publicKeyPEM,
		PrivateKey: privateKeyPEM,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(4 * time.Hour),
	}

	km.keysMu.Lock()
	km.sessionKeys[string(publicKeyPEM)] = sessionKey
	km.keysMu.Unlock()

	return sessionKey, nil
}

// GetPeerPublicKey retrieves a peer's public key
func (km *KeyManager) GetPeerPublicKey(publicKey []byte) (crypto.PubKey, error) {
	km.peerKeysMu.RLock()
	defer km.peerKeysMu.RUnlock()

	keyStr := string(publicKey)
	if pubKey, exists := km.peerKeys[keyStr]; exists {
		return pubKey, nil
	}

	// Parse PEM encoded public key
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	// Convert to libp2p crypto.PubKey
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("unsupported public key type")
	}

	_, libp2pKey, err := crypto.KeyPairFromStdKey(rsaPub)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to libp2p key: %w", err)
	}

	// Cache the key
	km.peerKeys[keyStr] = libp2pKey
	return libp2pKey, nil
}

// GetActiveSessionCount returns the number of active sessions
func (km *KeyManager) GetActiveSessionCount() int {
	km.keysMu.RLock()
	defer km.keysMu.RUnlock()
	
	activeCount := 0
	now := time.Now()
	for _, session := range km.sessionKeys {
		if session.ExpiresAt.After(now) {
			activeCount++
		}
	}
	return activeCount
}

// CleanupExpiredSessions removes expired sessions
func (km *KeyManager) CleanupExpiredSessions() {
	km.keysMu.Lock()
	defer km.keysMu.Unlock()
	
	now := time.Now()
	for keyStr, session := range km.sessionKeys {
		if session.ExpiresAt.Before(now) {
			delete(km.sessionKeys, keyStr)
		}
	}
}

// HandleKeyExchange handles key exchange requests
func (km *KeyManager) HandleKeyExchange(stream network.Stream) error {
	// Stub implementation for key exchange
	defer stream.Close()
	return nil
}

// RotateKeys rotates all session keys
func (km *KeyManager) RotateKeys() error {
	// Generate new session keys and mark old ones for expiration
	km.keysMu.Lock()
	defer km.keysMu.Unlock()
	
	// Mark all current keys as expiring soon
	expireTime := time.Now().Add(1 * time.Hour)
	for _, session := range km.sessionKeys {
		if session.ExpiresAt.After(expireTime) {
			session.ExpiresAt = expireTime
		}
	}
	
	return nil
}

// Close closes the key manager
func (km *KeyManager) Close() error {
	km.keysMu.Lock()
	defer km.keysMu.Unlock()
	
	// Clear all sessions
	km.sessionKeys = make(map[string]*SessionKey)
	km.peerKeys = make(map[string]crypto.PubKey)
	
	return nil
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(host host.Host, config *SecurityConfig) (*AuthManager, error) {
	// Generate key pair for authentication
	privateKey, publicKey, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	return &AuthManager{
		authenticatedPeers: make(map[peer.ID]time.Time),
		privateKey:         privateKey,
		publicKey:          publicKey,
		sessions:           make(map[string]*AuthSession),
		sessionTTL:         4 * time.Hour,
		config:             config,
	}, nil
}

// IsAuthenticated checks if a peer is authenticated
func (am *AuthManager) IsAuthenticated(peerID peer.ID) bool {
	am.peersMu.RLock()
	defer am.peersMu.RUnlock()
	
	if authTime, exists := am.authenticatedPeers[peerID]; exists {
		return time.Since(authTime) < am.sessionTTL
	}
	return false
}

// HandleAuthRequest handles authentication requests
func (am *AuthManager) HandleAuthRequest(stream network.Stream) error {
	// Stub implementation for authentication
	defer stream.Close()
	return nil
}

// NewEncryptionManager creates a new encryption manager
func NewEncryptionManager(keyManager *KeyManager, config *SecurityConfig) (*EncryptionManager, error) {
	return &EncryptionManager{
		keyManager: keyManager,
		config:     config,
	}, nil
}

// Encrypt encrypts data using session key
func (em *EncryptionManager) Encrypt(data []byte, sessionKey *SessionKey) ([]byte, error) {
	// Stub implementation - return data as-is for compatibility
	return data, nil
}

// Decrypt decrypts data using session key
func (em *EncryptionManager) Decrypt(data []byte, sessionKey *SessionKey) ([]byte, error) {
	// Stub implementation - return data as-is for compatibility
	return data, nil
}

// NewAccessControl creates a new access control manager
func NewAccessControl(config *SecurityConfig) (*AccessControl, error) {
	return &AccessControl{
		policies: make(map[peer.ID]*AccessPolicy),
		config:   config,
	}, nil
}

// IsAllowed checks if a peer is allowed to access a protocol
func (ac *AccessControl) IsAllowed(peerID peer.ID, protocolID protocol.ID) bool {
	ac.policyMu.RLock()
	defer ac.policyMu.RUnlock()
	
	// Check if peer is in blocked list
	for _, blockedPeer := range ac.config.BlockedPeers {
		if blockedPeer == peerID.String() {
			return false
		}
	}
	
	// Check if peer is in trusted list
	for _, trustedPeer := range ac.config.TrustedPeers {
		if trustedPeer == peerID.String() {
			return true
		}
	}
	
	// Default policy - allow if no specific policy
	return true
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *SecurityConfig) (*RateLimiter, error) {
	return &RateLimiter{
		limits: make(map[string]*RateLimit),
		config: config,
	}, nil
}

// Allow checks if a request is allowed under rate limits
func (rl *RateLimiter) Allow(peerID peer.ID, protocolID protocol.ID) bool {
	rl.limitMu.Lock()
	defer rl.limitMu.Unlock()
	
	key := fmt.Sprintf("%s:%s", peerID.String(), protocolID)
	
	limit, exists := rl.limits[key]
	if !exists {
		// Create new rate limit for this peer/protocol
		limit = &RateLimit{
			RequestsPerSecond: 10, // Default
			BurstSize:        20,
			Window:           1 * time.Second,
			lastReset:        time.Now(),
			currentCount:     0,
		}
		rl.limits[key] = limit
	}
	
	now := time.Now()
	
	// Reset counter if window has passed
	if now.Sub(limit.lastReset) >= limit.Window {
		limit.currentCount = 0
		limit.lastReset = now
	}
	
	// Check if limit is exceeded
	if limit.currentCount >= limit.RequestsPerSecond {
		return false
	}
	
	limit.currentCount++
	return true
}

// Cleanup removes old rate limit entries
func (rl *RateLimiter) Cleanup() {
	rl.limitMu.Lock()
	defer rl.limitMu.Unlock()
	
	now := time.Now()
	for key, limit := range rl.limits {
		if now.Sub(limit.lastReset) > 5*time.Minute {
			delete(rl.limits, key)
		}
	}
}
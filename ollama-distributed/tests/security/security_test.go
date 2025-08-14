package security

import (
	"context"
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/security"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/tests/integration"
)

// SecurityTestSuite provides comprehensive security testing
type SecurityTestSuite struct {
	cluster     *integration.TestCluster
	securityMgr *security.SecurityManager
}

// NewSecurityTestSuite creates a new security test suite
func NewSecurityTestSuite() (*SecurityTestSuite, error) {
	cluster, err := integration.NewTestCluster(3)
	if err != nil {
		return nil, err
	}

	// Create security manager for testing
	config := security.DefaultSecurityConfig()
	config.SecurityLevel = security.SecurityLevelHigh
	config.EnableEncryption = true
	config.EnableAuth = true

	securityMgr, err := security.NewSecurityManager(context.Background(), nil, config)
	if err != nil {
		return nil, err
	}

	return &SecurityTestSuite{
		cluster:     cluster,
		securityMgr: securityMgr,
	}, nil
}

// TestSecurityFramework runs the complete security test suite
func TestSecurityFramework(t *testing.T) {
	suite, err := NewSecurityTestSuite()
	require.NoError(t, err)
	defer suite.cleanup()

	t.Run("AuthenticationTests", func(t *testing.T) {
		suite.testAuthentication(t)
	})

	t.Run("EncryptionTests", func(t *testing.T) {
		suite.testEncryption(t)
	})

	t.Run("AccessControlTests", func(t *testing.T) {
		suite.testAccessControl(t)
	})

	t.Run("SecurityProtocolTests", func(t *testing.T) {
		suite.testSecurityProtocols(t)
	})

	t.Run("VulnerabilityTests", func(t *testing.T) {
		suite.testVulnerabilities(t)
	})

	t.Run("ThreatModelTests", func(t *testing.T) {
		suite.testThreatModel(t)
	})
}

// testAuthentication tests authentication mechanisms
func (s *SecurityTestSuite) testAuthentication(t *testing.T) {
	t.Run("ChallengeResponseAuth", func(t *testing.T) {
		s.testChallengeResponseAuth(t)
	})

	t.Run("TokenBasedAuth", func(t *testing.T) {
		s.testTokenBasedAuth(t)
	})

	t.Run("CertificateAuth", func(t *testing.T) {
		s.testCertificateAuth(t)
	})

	t.Run("AuthenticationFailures", func(t *testing.T) {
		s.testAuthenticationFailures(t)
	})

	t.Run("SessionManagement", func(t *testing.T) {
		s.testSessionManagement(t)
	})
}

// testChallengeResponseAuth tests challenge-response authentication
func (s *SecurityTestSuite) testChallengeResponseAuth(t *testing.T) {
	// Generate test peer
	testPeer, err := generateTestPeer()
	require.NoError(t, err)

	// Test challenge generation
	authMgr := s.securityMgr.GetAuthManager()
	require.NotNil(t, authMgr)

	// Create authentication request
	authReq := &security.AuthRequest{
		Type:      "auth_challenge",
		PeerID:    testPeer.String(),
		Method:    security.AuthMethodChallenge,
		Timestamp: time.Now(),
	}

	// Process challenge request
	response, err := authMgr.ProcessChallengeAuth(testPeer, authReq)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotEmpty(t, response.Metadata["challenge"])

	// Test challenge response
	challenge := response.Metadata["challenge"].([]byte)
	challengeResponse, err := signChallenge(testPeer, challenge)
	require.NoError(t, err)

	authReq.Response = challengeResponse
	finalResponse, err := authMgr.ProcessChallengeAuth(testPeer, authReq)
	require.NoError(t, err)
	assert.True(t, finalResponse.Success)
	assert.NotEmpty(t, finalResponse.Token)
}

// testTokenBasedAuth tests token-based authentication
func (s *SecurityTestSuite) testTokenBasedAuth(t *testing.T) {
	testPeer, err := generateTestPeer()
	require.NoError(t, err)

	authMgr := s.securityMgr.GetAuthManager()

	// Test valid token
	validToken := generateValidToken(testPeer)
	authReq := &security.AuthRequest{
		Type:      "token_auth",
		PeerID:    testPeer.String(),
		Method:    security.AuthMethodToken,
		Token:     validToken,
		Timestamp: time.Now(),
	}

	response, err := authMgr.ProcessTokenAuth(testPeer, authReq)
	require.NoError(t, err)
	assert.True(t, response.Success)

	// Test invalid token
	invalidToken := "invalid_token_12345"
	authReq.Token = invalidToken
	response, err = authMgr.ProcessTokenAuth(testPeer, authReq)
	assert.Error(t, err)
	assert.False(t, response.Success)
}

// testCertificateAuth tests certificate-based authentication
func (s *SecurityTestSuite) testCertificateAuth(t *testing.T) {
	testPeer, err := generateTestPeer()
	require.NoError(t, err)

	authMgr := s.securityMgr.GetAuthManager()

	// Generate test certificate
	cert, err := generateTestCertificate(testPeer)
	require.NoError(t, err)

	authReq := &security.AuthRequest{
		Type:      "cert_auth",
		PeerID:    testPeer.String(),
		Method:    security.AuthMethodCertificate,
		Metadata:  map[string]interface{}{"certificate": cert},
		Timestamp: time.Now(),
	}

	// Note: Certificate auth is not fully implemented yet
	_, err = authMgr.ProcessCertificateAuth(testPeer, authReq)
	assert.Error(t, err) // Should fail until implemented
}

// testAuthenticationFailures tests various authentication failure scenarios
func (s *SecurityTestSuite) testAuthenticationFailures(t *testing.T) {
	testPeer, err := generateTestPeer()
	require.NoError(t, err)

	authMgr := s.securityMgr.GetAuthManager()

	// Test expired timestamp
	oldTime := time.Now().Add(-10 * time.Minute)
	authReq := &security.AuthRequest{
		Type:      "auth_challenge",
		PeerID:    testPeer.String(),
		Method:    security.AuthMethodChallenge,
		Timestamp: oldTime,
	}

	_, err = authMgr.ProcessAuthRequest(testPeer, authReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timestamp too old")

	// Test blocked peer
	blockedPeer, err := generateTestPeer()
	require.NoError(t, err)

	// Add peer to blocked list
	authMgr.BlockPeer(blockedPeer)

	authReq.PeerID = blockedPeer.String()
	authReq.Timestamp = time.Now()

	_, err = authMgr.ProcessAuthRequest(blockedPeer, authReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "peer is blocked")

	// Test unsupported auth method
	authReq.Method = security.AuthMethod(999) // Invalid method
	authReq.PeerID = testPeer.String()

	_, err = authMgr.ProcessAuthRequest(testPeer, authReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported auth method")
}

// testSessionManagement tests session management
func (s *SecurityTestSuite) testSessionManagement(t *testing.T) {
	testPeer, err := generateTestPeer()
	require.NoError(t, err)

	authMgr := s.securityMgr.GetAuthManager()

	// Create a session
	session, err := authMgr.CreateAuthSession(testPeer, security.AuthLevelBasic)
	require.NoError(t, err)
	assert.Equal(t, testPeer, session.PeerID)
	assert.Equal(t, security.AuthLevelBasic, session.AuthLevel)

	// Test session validation
	assert.True(t, authMgr.IsAuthenticated(testPeer))
	assert.Equal(t, security.AuthLevelBasic, authMgr.GetAuthLevel(testPeer))

	// Test session expiration
	session.ExpiresAt = time.Now().Add(-1 * time.Hour) // Expired
	assert.False(t, authMgr.IsAuthenticated(testPeer))

	// Test session cleanup
	authMgr.CleanupExpiredSessions()
	assert.Equal(t, 0, authMgr.GetActiveSessionCount())
}

// testEncryption tests encryption mechanisms
func (s *SecurityTestSuite) testEncryption(t *testing.T) {
	t.Run("SymmetricEncryption", func(t *testing.T) {
		s.testSymmetricEncryption(t)
	})

	t.Run("AsymmetricEncryption", func(t *testing.T) {
		s.testAsymmetricEncryption(t)
	})

	t.Run("KeyExchange", func(t *testing.T) {
		s.testKeyExchange(t)
	})

	t.Run("SecureChannels", func(t *testing.T) {
		s.testSecureChannels(t)
	})
}

// testSymmetricEncryption tests symmetric encryption
func (s *SecurityTestSuite) testSymmetricEncryption(t *testing.T) {
	encMgr := s.securityMgr.GetEncryptionManager()
	require.NotNil(t, encMgr)

	// Generate session key
	sessionKey, err := s.securityMgr.GetKeyManager().GenerateSessionKey()
	require.NoError(t, err)

	// Test data
	plaintext := []byte("This is a test message for encryption")

	// Encrypt
	encrypted, err := encMgr.Encrypt(plaintext, sessionKey)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, encrypted)

	// Decrypt
	decrypted, err := encMgr.Decrypt(encrypted, sessionKey)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)

	// Test with different key (should fail)
	wrongKey, err := s.securityMgr.GetKeyManager().GenerateSessionKey()
	require.NoError(t, err)

	_, err = encMgr.Decrypt(encrypted, wrongKey)
	assert.Error(t, err)
}

// testAsymmetricEncryption tests asymmetric encryption
func (s *SecurityTestSuite) testAsymmetricEncryption(t *testing.T) {
	keyMgr := s.securityMgr.GetKeyManager()
	require.NotNil(t, keyMgr)

	// Generate key pair
	pubKey, privKey, err := keyMgr.GenerateKeyPair()
	require.NoError(t, err)

	// Test data
	plaintext := []byte("Asymmetric encryption test")

	// Encrypt with public key
	encrypted, err := keyMgr.EncryptWithPublicKey(plaintext, pubKey)
	require.NoError(t, err)

	// Decrypt with private key
	decrypted, err := keyMgr.DecryptWithPrivateKey(encrypted, privKey)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

// testKeyExchange tests key exchange protocols
func (s *SecurityTestSuite) testKeyExchange(t *testing.T) {
	keyMgr := s.securityMgr.GetKeyManager()

	// Test Diffie-Hellman key exchange
	privKeyA, pubKeyA, err := keyMgr.GenerateKeyPair()
	require.NoError(t, err)

	privKeyB, pubKeyB, err := keyMgr.GenerateKeyPair()
	require.NoError(t, err)

	// Compute shared secrets
	sharedSecretA, err := keyMgr.ComputeSharedSecret(privKeyA, pubKeyB)
	require.NoError(t, err)

	sharedSecretB, err := keyMgr.ComputeSharedSecret(privKeyB, pubKeyA)
	require.NoError(t, err)

	// Shared secrets should be equal
	assert.Equal(t, sharedSecretA, sharedSecretB)

	// Test key rotation
	err = keyMgr.RotateKeys()
	assert.NoError(t, err)
}

// testSecureChannels tests secure channel establishment
func (s *SecurityTestSuite) testSecureChannels(t *testing.T) {
	// This would require actual network setup
	t.Skip("Secure channel tests require network setup")
}

// testAccessControl tests access control mechanisms
func (s *SecurityTestSuite) testAccessControl(t *testing.T) {
	t.Run("PeerPermissions", func(t *testing.T) {
		s.testPeerPermissions(t)
	})

	t.Run("ProtocolAccess", func(t *testing.T) {
		s.testProtocolAccess(t)
	})

	t.Run("ResourceAccess", func(t *testing.T) {
		s.testResourceAccess(t)
	})

	t.Run("RateLimiting", func(t *testing.T) {
		s.testRateLimiting(t)
	})
}

// testPeerPermissions tests peer-based access control
func (s *SecurityTestSuite) testPeerPermissions(t *testing.T) {
	accessControl := s.securityMgr.GetAccessControl()
	require.NotNil(t, accessControl)

	testPeer, err := generateTestPeer()
	require.NoError(t, err)

	protocol := security.SecureChannelProtocol

	// Test default access (should be denied)
	allowed := accessControl.IsAllowed(testPeer, protocol)
	assert.False(t, allowed)

	// Add peer to allowed list
	accessControl.AddAllowedPeer(testPeer, protocol)
	allowed = accessControl.IsAllowed(testPeer, protocol)
	assert.True(t, allowed)

	// Remove peer from allowed list
	accessControl.RemoveAllowedPeer(testPeer, protocol)
	allowed = accessControl.IsAllowed(testPeer, protocol)
	assert.False(t, allowed)

	// Test trusted peers
	accessControl.AddTrustedPeer(testPeer)
	allowed = accessControl.IsAllowed(testPeer, protocol)
	assert.True(t, allowed) // Trusted peers should have access
}

// testProtocolAccess tests protocol-based access control
func (s *SecurityTestSuite) testProtocolAccess(t *testing.T) {
	accessControl := s.securityMgr.GetAccessControl()

	testPeer, err := generateTestPeer()
	require.NoError(t, err)

	// Test different security levels
	highSecProtocol := security.SecureChannelProtocol
	basicProtocol := security.KeyExchangeProtocol

	// Basic auth level should not access high security protocol
	authMgr := s.securityMgr.GetAuthManager()
	session, err := authMgr.CreateAuthSession(testPeer, security.AuthLevelBasic)
	require.NoError(t, err)

	allowed := accessControl.IsProtocolAllowed(testPeer, highSecProtocol, session.AuthLevel)
	assert.False(t, allowed)

	// But should access basic protocol
	allowed = accessControl.IsProtocolAllowed(testPeer, basicProtocol, session.AuthLevel)
	assert.True(t, allowed)

	// High auth level should access both
	session.AuthLevel = security.AuthLevelTrusted
	allowed = accessControl.IsProtocolAllowed(testPeer, highSecProtocol, session.AuthLevel)
	assert.True(t, allowed)
}

// testResourceAccess tests resource-based access control
func (s *SecurityTestSuite) testResourceAccess(t *testing.T) {
	// Test model access permissions
	testPeer, err := generateTestPeer()
	require.NoError(t, err)

	accessControl := s.securityMgr.GetAccessControl()

	// Test model access
	modelName := "llama3.2:1b"
	allowed := accessControl.CanAccessModel(testPeer, modelName)
	assert.False(t, allowed) // Default deny

	// Grant model access
	accessControl.GrantModelAccess(testPeer, modelName)
	allowed = accessControl.CanAccessModel(testPeer, modelName)
	assert.True(t, allowed)

	// Revoke model access
	accessControl.RevokeModelAccess(testPeer, modelName)
	allowed = accessControl.CanAccessModel(testPeer, modelName)
	assert.False(t, allowed)
}

// testRateLimiting tests rate limiting mechanisms
func (s *SecurityTestSuite) testRateLimiting(t *testing.T) {
	rateLimiter := s.securityMgr.GetRateLimiter()
	require.NotNil(t, rateLimiter)

	testPeer, err := generateTestPeer()
	require.NoError(t, err)

	protocol := security.AuthProtocol

	// Test rate limit enforcement
	allowed := rateLimiter.Allow(testPeer, protocol)
	assert.True(t, allowed) // First request should be allowed

	// Exceed rate limit
	for i := 0; i < 10; i++ {
		rateLimiter.Allow(testPeer, protocol)
	}

	// Should be rate limited now
	allowed = rateLimiter.Allow(testPeer, protocol)
	assert.False(t, allowed)

	// Test rate limit reset
	rateLimiter.Reset(testPeer, protocol)
	allowed = rateLimiter.Allow(testPeer, protocol)
	assert.True(t, allowed)
}

// testSecurityProtocols tests security protocol implementations
func (s *SecurityTestSuite) testSecurityProtocols(t *testing.T) {
	t.Run("NoiseProtocol", func(t *testing.T) {
		s.testNoiseProtocol(t)
	})

	t.Run("TLSProtocol", func(t *testing.T) {
		s.testTLSProtocol(t)
	})

	t.Run("CustomProtocols", func(t *testing.T) {
		s.testCustomProtocols(t)
	})
}

// testNoiseProtocol tests Noise protocol implementation
func (s *SecurityTestSuite) testNoiseProtocol(t *testing.T) {
	// Test Noise protocol handshake
	keyMgr := s.securityMgr.GetKeyManager()

	// Generate test keys
	staticKey, err := keyMgr.GenerateStaticKey()
	require.NoError(t, err)

	ephemeralKey, err := keyMgr.GenerateEphemeralKey()
	require.NoError(t, err)

	// Test handshake patterns
	handshake, err := keyMgr.CreateNoiseHandshake(staticKey, ephemeralKey)
	require.NoError(t, err)
	assert.NotNil(t, handshake)

	// Test message encryption/decryption
	plaintext := []byte("Noise protocol test message")
	ciphertext, err := handshake.Encrypt(plaintext)
	require.NoError(t, err)

	decrypted, err := handshake.Decrypt(ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

// testTLSProtocol tests TLS protocol implementation
func (s *SecurityTestSuite) testTLSProtocol(t *testing.T) {
	// Test TLS certificate generation
	keyMgr := s.securityMgr.GetKeyManager()

	cert, key, err := keyMgr.GenerateTLSCertificate("localhost")
	require.NoError(t, err)
	assert.NotEmpty(t, cert)
	assert.NotEmpty(t, key)

	// Test certificate validation
	valid, err := keyMgr.ValidateTLSCertificate(cert)
	require.NoError(t, err)
	assert.True(t, valid)
}

// testCustomProtocols tests custom security protocols
func (s *SecurityTestSuite) testCustomProtocols(t *testing.T) {
	// Test protocol registration
	customProtocol := &security.SecureProtocol{
		ID:                "test-protocol",
		RequireAuth:       true,
		RequireEncryption: true,
		SecurityLevel:     security.SecurityLevelHigh,
	}

	s.securityMgr.RegisterSecureProtocol("test-protocol", customProtocol)

	// Verify protocol is registered
	protocols := s.securityMgr.GetRegisteredProtocols()
	assert.Contains(t, protocols, "test-protocol")
}

// testVulnerabilities tests for known vulnerabilities
func (s *SecurityTestSuite) testVulnerabilities(t *testing.T) {
	t.Run("TimingAttacks", func(t *testing.T) {
		s.testTimingAttacks(t)
	})

	t.Run("ReplayAttacks", func(t *testing.T) {
		s.testReplayAttacks(t)
	})

	t.Run("ManInTheMiddle", func(t *testing.T) {
		s.testManInTheMiddleAttacks(t)
	})

	t.Run("DenialOfService", func(t *testing.T) {
		s.testDenialOfServiceAttacks(t)
	})
}

// testTimingAttacks tests for timing attack vulnerabilities
func (s *SecurityTestSuite) testTimingAttacks(t *testing.T) {
	authMgr := s.securityMgr.GetAuthManager()
	testPeer, err := generateTestPeer()
	require.NoError(t, err)

	// Test authentication timing
	validToken := generateValidToken(testPeer)
	invalidToken := "invalid_token"

	// Measure timing for valid token
	start := time.Now()
	authMgr.VerifyToken(validToken)
	validDuration := time.Since(start)

	// Measure timing for invalid token
	start = time.Now()
	authMgr.VerifyToken(invalidToken)
	invalidDuration := time.Since(start)

	// Timing difference should be minimal to prevent timing attacks
	timingDiff := abs(validDuration - invalidDuration)
	assert.Less(t, timingDiff, 1*time.Millisecond, "Timing difference too large, potential timing attack vulnerability")
}

// testReplayAttacks tests for replay attack vulnerabilities
func (s *SecurityTestSuite) testReplayAttacks(t *testing.T) {
	authMgr := s.securityMgr.GetAuthManager()
	testPeer, err := generateTestPeer()
	require.NoError(t, err)

	// Create authentication request
	authReq := &security.AuthRequest{
		Type:      "auth_challenge",
		PeerID:    testPeer.String(),
		Method:    security.AuthMethodChallenge,
		Timestamp: time.Now(),
	}

	// First request should succeed
	_, err = authMgr.ProcessAuthRequest(testPeer, authReq)
	assert.NoError(t, err)

	// Replay the same request (should fail)
	_, err = authMgr.ProcessAuthRequest(testPeer, authReq)
	assert.Error(t, err, "Replay attack should be prevented")
}

// testManInTheMiddleAttacks tests for MITM vulnerabilities
func (s *SecurityTestSuite) testManInTheMiddleAttacks(t *testing.T) {
	// Test certificate pinning
	keyMgr := s.securityMgr.GetKeyManager()

	// Generate legitimate certificate
	legit_cert, _, err := keyMgr.GenerateTLSCertificate("legitimate.example.com")
	require.NoError(t, err)

	// Generate fake certificate
	fake_cert, _, err := keyMgr.GenerateTLSCertificate("fake.example.com")
	require.NoError(t, err)

	// Pin the legitimate certificate
	keyMgr.PinCertificate("legitimate.example.com", legit_cert)

	// Legitimate certificate should be valid
	valid, err := keyMgr.ValidatePinnedCertificate("legitimate.example.com", legit_cert)
	require.NoError(t, err)
	assert.True(t, valid)

	// Fake certificate should be rejected
	valid, err = keyMgr.ValidatePinnedCertificate("legitimate.example.com", fake_cert)
	assert.Error(t, err)
	assert.False(t, valid)
}

// testDenialOfServiceAttacks tests for DoS vulnerabilities
func (s *SecurityTestSuite) testDenialOfServiceAttacks(t *testing.T) {
	rateLimiter := s.securityMgr.GetRateLimiter()
	testPeer, err := generateTestPeer()
	require.NoError(t, err)

	protocol := security.AuthProtocol

	// Simulate DoS attack with many requests
	attackRequests := 1000
	successCount := 0

	for i := 0; i < attackRequests; i++ {
		if rateLimiter.Allow(testPeer, protocol) {
			successCount++
		}
	}

	// Rate limiter should prevent most requests
	successRate := float64(successCount) / float64(attackRequests)
	assert.Less(t, successRate, 0.1, "Rate limiter should prevent DoS attacks")
}

// testThreatModel tests against comprehensive threat model
func (s *SecurityTestSuite) testThreatModel(t *testing.T) {
	t.Run("ConfidentialityThreats", func(t *testing.T) {
		s.testConfidentialityThreats(t)
	})

	t.Run("IntegrityThreats", func(t *testing.T) {
		s.testIntegrityThreats(t)
	})

	t.Run("AvailabilityThreats", func(t *testing.T) {
		s.testAvailabilityThreats(t)
	})

	t.Run("AuthenticationThreats", func(t *testing.T) {
		s.testAuthenticationThreats(t)
	})
}

// testConfidentialityThreats tests against confidentiality threats
func (s *SecurityTestSuite) testConfidentialityThreats(t *testing.T) {
	// Test data encryption in transit
	encMgr := s.securityMgr.GetEncryptionManager()
	sensitiveData := []byte("CONFIDENTIAL: Secret model weights")

	sessionKey, err := s.securityMgr.GetKeyManager().GenerateSessionKey()
	require.NoError(t, err)

	encrypted, err := encMgr.Encrypt(sensitiveData, sessionKey)
	require.NoError(t, err)

	// Encrypted data should not contain original content
	assert.NotContains(t, string(encrypted), "CONFIDENTIAL")
	assert.NotContains(t, string(encrypted), "Secret model weights")
}

// testIntegrityThreats tests against integrity threats
func (s *SecurityTestSuite) testIntegrityThreats(t *testing.T) {
	// Test message authentication
	keyMgr := s.securityMgr.GetKeyManager()
	message := []byte("Important consensus message")

	// Generate MAC
	mac, err := keyMgr.GenerateMAC(message)
	require.NoError(t, err)

	// Verify MAC
	valid, err := keyMgr.VerifyMAC(message, mac)
	require.NoError(t, err)
	assert.True(t, valid)

	// Test tampered message
	tamperedMessage := []byte("Tampered consensus message")
	valid, err = keyMgr.VerifyMAC(tamperedMessage, mac)
	assert.False(t, valid, "Tampered message should fail MAC verification")
}

// testAvailabilityThreats tests against availability threats
func (s *SecurityTestSuite) testAvailabilityThreats(t *testing.T) {
	// Test resource exhaustion protection
	rateLimiter := s.securityMgr.GetRateLimiter()

	// Multiple peers attacking simultaneously
	attackerCount := 10
	for i := 0; i < attackerCount; i++ {
		attacker, err := generateTestPeer()
		require.NoError(t, err)

		// Each attacker tries to exhaust resources
		for j := 0; j < 100; j++ {
			rateLimiter.Allow(attacker, security.AuthProtocol)
		}
	}

	// Legitimate user should still be able to access
	legitimateUser, err := generateTestPeer()
	require.NoError(t, err)

	allowed := rateLimiter.Allow(legitimateUser, security.AuthProtocol)
	assert.True(t, allowed, "Legitimate user should still have access during attack")
}

// testAuthenticationThreats tests against authentication threats
func (s *SecurityTestSuite) testAuthenticationThreats(t *testing.T) {
	authMgr := s.securityMgr.GetAuthManager()

	// Test brute force protection
	attacker, err := generateTestPeer()
	require.NoError(t, err)

	// Multiple failed authentication attempts
	failureCount := 0
	for i := 0; i < 20; i++ {
		invalidToken := fmt.Sprintf("invalid_token_%d", i)
		_, err := authMgr.VerifyToken(invalidToken)
		if err != nil {
			failureCount++
		}
	}

	// After multiple failures, account should be locked
	locked := authMgr.IsAccountLocked(attacker)
	assert.True(t, locked, "Account should be locked after multiple failed attempts")
}

// Helper functions

func (s *SecurityTestSuite) cleanup() {
	if s.securityMgr != nil {
		s.securityMgr.Close()
	}
	if s.cluster != nil {
		s.cluster.Shutdown()
	}
}

func generateTestPeer() (peer.ID, error) {
	// Generate a random peer ID for testing
	randBytes := make([]byte, 32)
	_, err := rand.Read(randBytes)
	if err != nil {
		return "", err
	}

	// In a real implementation, this would use libp2p's peer ID generation
	return peer.ID(fmt.Sprintf("test-peer-%x", randBytes[:8])), nil
}

func signChallenge(peerID peer.ID, challenge []byte) ([]byte, error) {
	// Mock signature for testing
	signature := fmt.Sprintf("signature-%s-%x", peerID, challenge[:8])
	return []byte(signature), nil
}

func generateValidToken(peerID peer.ID) string {
	// Mock valid token for testing
	return fmt.Sprintf("valid-token-%s-%d", peerID, time.Now().Unix())
}

func generateTestCertificate(peerID peer.ID) ([]byte, error) {
	// Mock certificate for testing
	cert := fmt.Sprintf("-----BEGIN CERTIFICATE-----\ntest-cert-%s\n-----END CERTIFICATE-----", peerID)
	return []byte(cert), nil
}

func abs(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

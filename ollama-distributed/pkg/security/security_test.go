package security

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSecurityManager(t *testing.T) {
	cfg := &SecurityConfig{
		EnableTLS:           true,
		TLSCertPath:         "test-cert.pem",
		TLSKeyPath:          "test-key.pem",
		EnableEncryption:    true,
		EncryptionAlgorithm: "AES-256-GCM",
		KeySize:             256,
		RequireAuth:         true,
		AuthTimeout:         24 * time.Hour,
		MaxConnections:      100,
		ConnectionTimeout:   30 * time.Second,
		SessionTimeout:      24 * time.Hour,
	}

	manager, err := NewSecurityManager(cfg)
	require.NoError(t, err)
	require.NotNil(t, manager)

	defer manager.Close()
}

func TestSecurityManager_EncryptDecrypt(t *testing.T) {
	cfg := &SecurityConfig{
		EnableEncryption:    true,
		EncryptionAlgorithm: "AES-256-GCM",
		KeySize:             256,
	}

	manager, err := NewSecurityManager(cfg)
	require.NoError(t, err)
	defer manager.Close()

	// Test data
	testData := []byte("Hello, World!")
	testKey := make([]byte, 32) // 256-bit key

	// Test encryption
	encrypted, err := manager.EncryptData(testData, testKey)
	require.NoError(t, err)
	require.NotEqual(t, testData, encrypted)

	// Test decryption
	decrypted, err := manager.DecryptData(encrypted, testKey)
	require.NoError(t, err)
	require.Equal(t, testData, decrypted)
}

func TestSecurityManager_BasicFunctionality(t *testing.T) {
	cfg := &SecurityConfig{
		EnableTLS:           true,
		TLSCertPath:         "test-cert.pem",
		TLSKeyPath:          "test-key.pem",
		EnableEncryption:    true,
		EncryptionAlgorithm: "AES-256-GCM",
		KeySize:             256,
		RequireAuth:         true,
		AuthTimeout:         time.Hour,
		MaxConnections:      100,
		ConnectionTimeout:   30 * time.Second,
		SessionTimeout:      24 * time.Hour,
	}

	manager, err := NewSecurityManager(cfg)
	require.NoError(t, err)
	defer manager.Close()

	// Test that manager was created successfully
	require.NotNil(t, manager)

	// Test basic configuration access
	require.True(t, manager.config.EnableTLS)
	require.True(t, manager.config.EnableEncryption)
	require.Equal(t, "AES-256-GCM", manager.config.EncryptionAlgorithm)
}

func TestSecurityConfig_Validation(t *testing.T) {
	// Test valid configuration
	validCfg := &SecurityConfig{
		EnableTLS:           true,
		TLSCertPath:         "test-cert.pem",
		TLSKeyPath:          "test-key.pem",
		EnableEncryption:    true,
		EncryptionAlgorithm: "AES-256-GCM",
		KeySize:             256,
		RequireAuth:         true,
		AuthTimeout:         time.Hour,
		MaxConnections:      100,
		ConnectionTimeout:   30 * time.Second,
		SessionTimeout:      24 * time.Hour,
	}

	manager, err := NewSecurityManager(validCfg)
	require.NoError(t, err)
	require.NotNil(t, manager)
	manager.Close()

	// Test configuration with minimal settings
	minimalCfg := &SecurityConfig{
		EnableTLS:        false,
		EnableEncryption: false,
		RequireAuth:      false,
	}

	manager2, err := NewSecurityManager(minimalCfg)
	require.NoError(t, err)
	require.NotNil(t, manager2)
	manager2.Close()
}

func TestSecurityManager_KeyGeneration(t *testing.T) {
	cfg := &SecurityConfig{
		EnableEncryption:    true,
		EncryptionAlgorithm: "AES-256-GCM",
		KeySize:             256,
	}

	manager, err := NewSecurityManager(cfg)
	require.NoError(t, err)
	defer manager.Close()

	// Test key generation
	key, err := manager.GenerateKey(32)
	require.NoError(t, err)
	require.Len(t, key, 32)

	// Test that generated keys are different
	key2, err := manager.GenerateKey(32)
	require.NoError(t, err)
	require.Len(t, key2, 32)
	require.NotEqual(t, key, key2)
}

func TestSecurityManager_HashPassword(t *testing.T) {
	cfg := &SecurityConfig{
		RequireAuth: true,
	}

	manager, err := NewSecurityManager(cfg)
	require.NoError(t, err)
	defer manager.Close()

	password := "test-password-123"

	// Test password hashing
	hashedPassword, err := manager.HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)
	require.NotEqual(t, password, hashedPassword)

	// Test password verification
	isValid, err := manager.VerifyPassword(password, hashedPassword)
	require.NoError(t, err)
	require.True(t, isValid)

	// Test invalid password
	isValid, err = manager.VerifyPassword("wrong-password", hashedPassword)
	require.NoError(t, err)
	require.False(t, isValid)
}

func TestSecurityManager_SessionManagement(t *testing.T) {
	cfg := &SecurityConfig{
		RequireAuth:    true,
		SessionTimeout: time.Hour,
	}

	manager, err := NewSecurityManager(cfg)
	require.NoError(t, err)
	defer manager.Close()

	userID := "test-user-123"

	// Test session creation
	sessionID, err := manager.CreateSession(userID)
	require.NoError(t, err)
	require.NotEmpty(t, sessionID)

	// Test session validation
	isValid, retrievedUserID, err := manager.ValidateSession(sessionID)
	require.NoError(t, err)
	require.True(t, isValid)
	require.Equal(t, userID, retrievedUserID)

	// Test session cleanup
	err = manager.CleanupExpiredSessions()
	require.NoError(t, err)
}

func TestSecurityManager_AccessControl(t *testing.T) {
	cfg := &SecurityConfig{
		RequireAuth:    true,
		MaxConnections: 10,
	}

	manager, err := NewSecurityManager(cfg)
	require.NoError(t, err)
	defer manager.Close()

	userID := "test-user"
	resource := "test-resource"
	action := "read"

	// Test permission checking
	hasPermission, err := manager.CheckPermission(userID, resource, action)
	require.NoError(t, err)
	// Default should be false for unknown user/resource
	require.False(t, hasPermission)

	// Test granting permission
	err = manager.GrantPermission(userID, resource, action)
	require.NoError(t, err)

	// Test permission checking after granting
	hasPermission, err = manager.CheckPermission(userID, resource, action)
	require.NoError(t, err)
	require.True(t, hasPermission)

	// Test revoking permission
	err = manager.RevokePermission(userID, resource, action)
	require.NoError(t, err)

	// Test permission checking after revoking
	hasPermission, err = manager.CheckPermission(userID, resource, action)
	require.NoError(t, err)
	require.False(t, hasPermission)
}

func TestSecurityManager_ConnectionLimits(t *testing.T) {
	cfg := &SecurityConfig{
		MaxConnections:    2,
		ConnectionTimeout: 5 * time.Second,
	}

	manager, err := NewSecurityManager(cfg)
	require.NoError(t, err)
	defer manager.Close()

	// Test connection tracking
	conn1 := "connection-1"
	conn2 := "connection-2"
	conn3 := "connection-3"

	// Add connections up to limit
	err = manager.AddConnection(conn1)
	require.NoError(t, err)

	err = manager.AddConnection(conn2)
	require.NoError(t, err)

	// Adding beyond limit should fail
	err = manager.AddConnection(conn3)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection limit")

	// Remove a connection
	err = manager.RemoveConnection(conn1)
	require.NoError(t, err)

	// Now adding should work
	err = manager.AddConnection(conn3)
	require.NoError(t, err)

	// Test getting connection count
	count := manager.GetConnectionCount()
	assert.Equal(t, 2, count)
}

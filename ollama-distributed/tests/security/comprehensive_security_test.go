package security

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/security"
)

// TestAuthenticationSuite tests all authentication mechanisms
func TestAuthenticationSuite(t *testing.T) {
	t.Run("JWTTokenGeneration", testJWTTokenGeneration)
	t.Run("JWTTokenValidation", testJWTTokenValidation)
	t.Run("JWTTokenExpiration", testJWTTokenExpiration)
	t.Run("InvalidTokenHandling", testInvalidTokenHandling)
	t.Run("TokenRefresh", testTokenRefresh)
	t.Run("MultiTenantAuth", testMultiTenantAuth)
	t.Run("RoleBasedAuth", testRoleBasedAuth)
}

// testJWTTokenGeneration tests JWT token creation
func testJWTTokenGeneration(t *testing.T) {
	// Create auth manager with test secret
	authManager := security.NewAuthManager(&security.AuthConfig{
		JWTSecret:     []byte("test-secret-key"),
		TokenExpiry:   time.Hour,
		RefreshExpiry: 24 * time.Hour,
	})

	// Test token generation for different user types
	testCases := []struct {
		name     string
		userID   string
		role     string
		metadata map[string]interface{}
	}{
		{
			name:   "AdminUser",
			userID: "admin-001",
			role:   "admin",
			metadata: map[string]interface{}{
				"permissions": []string{"read", "write", "delete"},
				"tenant":      "default",
			},
		},
		{
			name:   "StandardUser",
			userID: "user-001",
			role:   "user",
			metadata: map[string]interface{}{
				"permissions": []string{"read"},
				"tenant":      "tenant-1",
			},
		},
		{
			name:   "ServiceAccount",
			userID: "service-001",
			role:   "service",
			metadata: map[string]interface{}{
				"permissions": []string{"read", "write"},
				"service":     "model-sync",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate token
			token, err := authManager.GenerateToken(tc.userID, tc.role, tc.metadata)
			assert.NoError(t, err)
			assert.NotEmpty(t, token)

			// Verify token format (JWT has 3 parts separated by dots)
			parts := strings.Split(token, ".")
			assert.Equal(t, 3, len(parts))

			// Verify token contains expected claims
			claims, err := authManager.ValidateToken(token)
			assert.NoError(t, err)
			assert.Equal(t, tc.userID, claims.UserID)
			assert.Equal(t, tc.role, claims.Role)
			assert.Equal(t, tc.metadata, claims.Metadata)
		})
	}
}

// testJWTTokenValidation tests token validation
func testJWTTokenValidation(t *testing.T) {
	authManager := security.NewAuthManager(&security.AuthConfig{
		JWTSecret:   []byte("validation-test-secret"),
		TokenExpiry: time.Hour,
	})

	// Generate valid token
	validToken, err := authManager.GenerateToken("test-user", "admin", nil)
	require.NoError(t, err)

	t.Run("ValidToken", func(t *testing.T) {
		claims, err := authManager.ValidateToken(validToken)
		assert.NoError(t, err)
		assert.Equal(t, "test-user", claims.UserID)
		assert.Equal(t, "admin", claims.Role)
	})

	t.Run("InvalidSignature", func(t *testing.T) {
		// Create auth manager with different secret
		wrongManager := security.NewAuthManager(&security.AuthConfig{
			JWTSecret: []byte("wrong-secret"),
		})

		_, err := wrongManager.ValidateToken(validToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "signature is invalid")
	})

	t.Run("MalformedToken", func(t *testing.T) {
		malformedTokens := []string{
			"invalid.token",
			"invalid",
			"",
			"invalid.token.with.too.many.parts",
		}

		for _, token := range malformedTokens {
			_, err := authManager.ValidateToken(token)
			assert.Error(t, err)
		}
	})

	t.Run("TamperedToken", func(t *testing.T) {
		// Tamper with token by changing a character
		tamperedToken := validToken[:len(validToken)-5] + "xxxxx"
		_, err := authManager.ValidateToken(tamperedToken)
		assert.Error(t, err)
	})
}

// testJWTTokenExpiration tests token expiration handling
func testJWTTokenExpiration(t *testing.T) {
	// Create auth manager with very short expiry
	authManager := security.NewAuthManager(&security.AuthConfig{
		JWTSecret:   []byte("expiry-test-secret"),
		TokenExpiry: 100 * time.Millisecond,
	})

	// Generate token that will expire quickly
	token, err := authManager.GenerateToken("test-user", "user", nil)
	require.NoError(t, err)

	// Token should be valid immediately
	_, err = authManager.ValidateToken(token)
	assert.NoError(t, err)

	// Wait for token to expire
	time.Sleep(200 * time.Millisecond)

	// Token should now be expired
	_, err = authManager.ValidateToken(token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token is expired")
}

// testInvalidTokenHandling tests various invalid token scenarios
func testInvalidTokenHandling(t *testing.T) {
	authManager := security.NewAuthManager(&security.AuthConfig{
		JWTSecret: []byte("invalid-test-secret"),
	})

	testCases := []struct {
		name        string
		token       string
		expectError string
	}{
		{
			name:        "EmptyToken",
			token:       "",
			expectError: "token is empty",
		},
		{
			name:        "WhitespaceToken",
			token:       "   ",
			expectError: "token is empty",
		},
		{
			name:        "RandomString",
			token:       "this-is-not-a-jwt-token",
			expectError: "invalid token format",
		},
		{
			name:        "Base64ButNotJWT",
			token:       "aGVsbG8gd29ybGQ=",
			expectError: "invalid token format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := authManager.ValidateToken(tc.token)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectError)
		})
	}
}

// testTokenRefresh tests token refresh functionality
func testTokenRefresh(t *testing.T) {
	authManager := security.NewAuthManager(&security.AuthConfig{
		JWTSecret:     []byte("refresh-test-secret"),
		TokenExpiry:   time.Hour,
		RefreshExpiry: 24 * time.Hour,
	})

	// Generate initial token pair
	accessToken, refreshToken, err := authManager.GenerateTokenPair("test-user", "user", nil)
	require.NoError(t, err)

	t.Run("ValidRefresh", func(t *testing.T) {
		// Refresh tokens
		newAccessToken, newRefreshToken, err := authManager.RefreshToken(refreshToken)
		assert.NoError(t, err)
		assert.NotEmpty(t, newAccessToken)
		assert.NotEmpty(t, newRefreshToken)
		assert.NotEqual(t, accessToken, newAccessToken)
		assert.NotEqual(t, refreshToken, newRefreshToken)

		// New access token should be valid
		claims, err := authManager.ValidateToken(newAccessToken)
		assert.NoError(t, err)
		assert.Equal(t, "test-user", claims.UserID)
	})

	t.Run("InvalidRefreshToken", func(t *testing.T) {
		_, _, err := authManager.RefreshToken("invalid-refresh-token")
		assert.Error(t, err)
	})

	t.Run("ExpiredRefreshToken", func(t *testing.T) {
		// Create auth manager with very short refresh expiry
		shortManager := security.NewAuthManager(&security.AuthConfig{
			JWTSecret:     []byte("short-refresh-secret"),
			TokenExpiry:   time.Hour,
			RefreshExpiry: 100 * time.Millisecond,
		})

		_, expiredRefreshToken, err := shortManager.GenerateTokenPair("test-user", "user", nil)
		require.NoError(t, err)

		// Wait for refresh token to expire
		time.Sleep(200 * time.Millisecond)

		_, _, err = shortManager.RefreshToken(expiredRefreshToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})
}

// testMultiTenantAuth tests multi-tenant authentication
func testMultiTenantAuth(t *testing.T) {
	authManager := security.NewAuthManager(&security.AuthConfig{
		JWTSecret:      []byte("multitenant-secret"),
		TokenExpiry:    time.Hour,
		EnableTenants:  true,
		TenantRequired: true,
	})

	tenants := []string{"tenant-1", "tenant-2", "tenant-3"}

	for _, tenant := range tenants {
		t.Run(fmt.Sprintf("Tenant_%s", tenant), func(t *testing.T) {
			// Generate token with tenant
			metadata := map[string]interface{}{
				"tenant": tenant,
			}
			token, err := authManager.GenerateToken("user-001", "user", metadata)
			assert.NoError(t, err)

			// Validate token
			claims, err := authManager.ValidateToken(token)
			assert.NoError(t, err)
			assert.Equal(t, tenant, claims.Metadata["tenant"])

			// Test tenant isolation
			err = authManager.ValidateTenantAccess(claims, tenant)
			assert.NoError(t, err)

			// Test access to different tenant (should fail)
			otherTenant := "other-tenant"
			err = authManager.ValidateTenantAccess(claims, otherTenant)
			assert.Error(t, err)
		})
	}
}

// testRoleBasedAuth tests role-based authentication and authorization
func testRoleBasedAuth(t *testing.T) {
	authManager := security.NewAuthManager(&security.AuthConfig{
		JWTSecret:   []byte("rbac-secret"),
		TokenExpiry: time.Hour,
		RBAC: &security.RBACConfig{
			Enabled: true,
			Roles: map[string]*security.Role{
				"admin": {
					Name:        "admin",
					Permissions: []string{"*"},
				},
				"user": {
					Name:        "user",
					Permissions: []string{"read", "write"},
				},
				"readonly": {
					Name:        "readonly",
					Permissions: []string{"read"},
				},
			},
		},
	})

	testCases := []struct {
		role       string
		permission string
		allowed    bool
	}{
		{"admin", "read", true},
		{"admin", "write", true},
		{"admin", "delete", true},
		{"admin", "admin", true},
		{"user", "read", true},
		{"user", "write", true},
		{"user", "delete", false},
		{"readonly", "read", true},
		{"readonly", "write", false},
		{"readonly", "delete", false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_%s", tc.role, tc.permission), func(t *testing.T) {
			// Generate token with role
			token, err := authManager.GenerateToken("test-user", tc.role, nil)
			require.NoError(t, err)

			// Validate token
			claims, err := authManager.ValidateToken(token)
			require.NoError(t, err)

			// Check permission
			hasPermission := authManager.HasPermission(claims, tc.permission)
			assert.Equal(t, tc.allowed, hasPermission)
		})
	}
}

// TestEncryptionSuite tests encryption and decryption functionality
func TestEncryptionSuite(t *testing.T) {
	t.Run("AESEncryption", testAESEncryption)
	t.Run("RSAEncryption", testRSAEncryption)
	t.Run("TLSConfiguration", testTLSConfiguration)
	t.Run("KeyRotation", testKeyRotation)
	t.Run("DataEncryptionAtRest", testDataEncryptionAtRest)
	t.Run("DataEncryptionInTransit", testDataEncryptionInTransit)
}

// testAESEncryption tests AES encryption/decryption
func testAESEncryption(t *testing.T) {
	encryptionManager := security.NewEncryptionManager(&security.EncryptionConfig{
		Algorithm: "AES-256-GCM",
		KeySize:   32,
	})

	testData := [][]byte{
		[]byte("Hello, World!"),
		[]byte(""),
		[]byte("This is a longer test string with special characters: !@#$%^&*()"),
		make([]byte, 1024*1024), // 1MB of zeros
	}

	for i, data := range testData {
		t.Run(fmt.Sprintf("TestData_%d", i), func(t *testing.T) {
			// Generate key
			key, err := encryptionManager.GenerateKey()
			require.NoError(t, err)
			assert.Equal(t, 32, len(key))

			// Encrypt data
			encrypted, err := encryptionManager.Encrypt(data, key)
			assert.NoError(t, err)
			assert.NotEqual(t, data, encrypted)

			// Decrypt data
			decrypted, err := encryptionManager.Decrypt(encrypted, key)
			assert.NoError(t, err)
			assert.Equal(t, data, decrypted)

			// Test with wrong key
			wrongKey, _ := encryptionManager.GenerateKey()
			_, err = encryptionManager.Decrypt(encrypted, wrongKey)
			assert.Error(t, err)
		})
	}
}

// testRSAEncryption tests RSA encryption for small data
func testRSAEncryption(t *testing.T) {
	rsaManager := security.NewRSAManager(&security.RSAConfig{
		KeySize: 2048,
	})

	// Generate key pair
	privateKey, publicKey, err := rsaManager.GenerateKeyPair()
	require.NoError(t, err)

	testData := [][]byte{
		[]byte("Small data"),
		[]byte(""),
		[]byte("Test with special chars: αβγδε"),
	}

	for i, data := range testData {
		t.Run(fmt.Sprintf("RSAData_%d", i), func(t *testing.T) {
			// Encrypt with public key
			encrypted, err := rsaManager.Encrypt(data, publicKey)
			assert.NoError(t, err)
			assert.NotEqual(t, data, encrypted)

			// Decrypt with private key
			decrypted, err := rsaManager.Decrypt(encrypted, privateKey)
			assert.NoError(t, err)
			assert.Equal(t, data, decrypted)
		})
	}

	t.Run("DataTooLarge", func(t *testing.T) {
		// RSA has size limits, test with data too large
		largeData := make([]byte, 1024)
		_, err := rsaManager.Encrypt(largeData, publicKey)
		assert.Error(t, err)
	})
}

// testTLSConfiguration tests TLS configuration and certificate validation
func testTLSConfiguration(t *testing.T) {
	tlsManager := security.NewTLSManager(&security.TLSConfig{
		CertFile:           "/tmp/test-cert.pem",
		KeyFile:            "/tmp/test-key.pem",
		CAFile:             "/tmp/test-ca.pem",
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
	})

	t.Run("CertificateGeneration", func(t *testing.T) {
		// Generate self-signed certificate for testing
		cert, key, err := tlsManager.GenerateSelfSignedCert("localhost", []string{"127.0.0.1"}, time.Hour)
		assert.NoError(t, err)
		assert.NotEmpty(t, cert)
		assert.NotEmpty(t, key)

		// Parse certificate
		certParsed, err := x509.ParseCertificate(cert)
		assert.NoError(t, err)
		assert.Equal(t, "localhost", certParsed.Subject.CommonName)
	})

	t.Run("TLSConfigCreation", func(t *testing.T) {
		// Create TLS config
		config, err := tlsManager.CreateTLSConfig()
		assert.NoError(t, err)
		assert.Equal(t, uint16(tls.VersionTLS12), config.MinVersion)
		assert.Equal(t, uint16(tls.VersionTLS13), config.MaxVersion)
		assert.False(t, config.InsecureSkipVerify)
	})

	t.Run("CertificateValidation", func(t *testing.T) {
		// Test certificate validation
		cert, _, err := tlsManager.GenerateSelfSignedCert("test.example.com", nil, time.Hour)
		require.NoError(t, err)

		isValid := tlsManager.ValidateCertificate(cert, "test.example.com")
		assert.True(t, isValid)

		isValid = tlsManager.ValidateCertificate(cert, "wrong.example.com")
		assert.False(t, isValid)
	})
}

// testKeyRotation tests key rotation functionality
func testKeyRotation(t *testing.T) {
	keyManager := security.NewKeyManager(&security.KeyConfig{
		RotationInterval: time.Hour,
		KeyVersions:      3,
		Algorithm:        "AES-256",
	})

	t.Run("KeyGeneration", func(t *testing.T) {
		// Generate initial key
		key1, version1, err := keyManager.GenerateKey()
		assert.NoError(t, err)
		assert.Equal(t, 32, len(key1))
		assert.Equal(t, 1, version1)

		// Generate second key
		key2, version2, err := keyManager.GenerateKey()
		assert.NoError(t, err)
		assert.Equal(t, 32, len(key2))
		assert.Equal(t, 2, version2)
		assert.NotEqual(t, key1, key2)
	})

	t.Run("KeyRetrieval", func(t *testing.T) {
		// Generate keys
		key1, version1, _ := keyManager.GenerateKey()
		key2, version2, _ := keyManager.GenerateKey()

		// Retrieve keys by version
		retrievedKey1, err := keyManager.GetKey(version1)
		assert.NoError(t, err)
		assert.Equal(t, key1, retrievedKey1)

		retrievedKey2, err := keyManager.GetKey(version2)
		assert.NoError(t, err)
		assert.Equal(t, key2, retrievedKey2)

		// Try to retrieve non-existent key
		_, err = keyManager.GetKey(999)
		assert.Error(t, err)
	})

	t.Run("KeyRotation", func(t *testing.T) {
		// Test automatic key rotation
		initialVersion := keyManager.GetCurrentVersion()

		err := keyManager.RotateKeys()
		assert.NoError(t, err)

		newVersion := keyManager.GetCurrentVersion()
		assert.Greater(t, newVersion, initialVersion)
	})
}

// testDataEncryptionAtRest tests data encryption for stored data
func testDataEncryptionAtRest(t *testing.T) {
	storageManager := security.NewStorageEncryption(&security.StorageConfig{
		EncryptionEnabled: true,
		Algorithm:         "AES-256-GCM",
		CompressionLevel:  6,
	})

	testData := map[string][]byte{
		"model_weights": make([]byte, 1024*1024), // 1MB
		"config_file":   []byte(`{"key": "value", "number": 42}`),
		"empty_file":    []byte(""),
		"binary_data":   {0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD},
	}

	for name, data := range testData {
		t.Run(name, func(t *testing.T) {
			// Encrypt and store
			encryptedData, metadata, err := storageManager.EncryptForStorage(data)
			assert.NoError(t, err)
			assert.NotNil(t, metadata)

			if len(data) > 0 {
				assert.NotEqual(t, data, encryptedData)
			}

			// Decrypt from storage
			decryptedData, err := storageManager.DecryptFromStorage(encryptedData, metadata)
			assert.NoError(t, err)
			assert.Equal(t, data, decryptedData)
		})
	}
}

// testDataEncryptionInTransit tests data encryption for network transmission
func testDataEncryptionInTransit(t *testing.T) {
	// Create test HTTP server with TLS
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo back received data
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		w.Write(body)
	}))

	// Configure TLS
	tlsManager := security.NewTLSManager(&security.TLSConfig{
		MinVersion: tls.VersionTLS12,
	})

	cert, key, err := tlsManager.GenerateSelfSignedCert("localhost", []string{"127.0.0.1"}, time.Hour)
	require.NoError(t, err)

	tlsCert, err := tls.X509KeyPair(cert, key)
	require.NoError(t, err)

	server.TLS = &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	}
	server.StartTLS()
	defer server.Close()

	t.Run("EncryptedTransmission", func(t *testing.T) {
		// Create client with custom TLS config
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // For testing only
				},
			},
		}

		testData := []byte("Sensitive data to transmit")

		// Send data to server
		resp, err := client.Post(server.URL, "application/octet-stream", strings.NewReader(string(testData)))
		assert.NoError(t, err)
		defer resp.Body.Close()

		// Read response
		responseData := make([]byte, len(testData))
		resp.Body.Read(responseData)

		assert.Equal(t, testData, responseData)
		assert.Equal(t, "https", resp.Request.URL.Scheme)
	})
}

// TestAuthorizationSuite tests authorization mechanisms
func TestAuthorizationSuite(t *testing.T) {
	t.Run("ResourcePermissions", testResourcePermissions)
	t.Run("ActionPermissions", testActionPermissions)
	t.Run("ConditionalAccess", testConditionalAccess)
	t.Run("PermissionInheritance", testPermissionInheritance)
}

// testResourcePermissions tests resource-based permissions
func testResourcePermissions(t *testing.T) {
	authz := security.NewAuthorization(&security.AuthorizationConfig{
		DefaultDeny: true,
		Resources: map[string]*security.ResourceConfig{
			"models": {
				Name:    "models",
				Actions: []string{"read", "write", "delete"},
				Policies: []*security.Policy{
					{
						Effect:  "allow",
						Actions: []string{"read"},
						Roles:   []string{"user", "admin"},
					},
					{
						Effect:  "allow",
						Actions: []string{"write", "delete"},
						Roles:   []string{"admin"},
					},
				},
			},
			"nodes": {
				Name:    "nodes",
				Actions: []string{"read", "manage"},
				Policies: []*security.Policy{
					{
						Effect:  "allow",
						Actions: []string{"read"},
						Roles:   []string{"user", "admin"},
					},
					{
						Effect:  "allow",
						Actions: []string{"manage"},
						Roles:   []string{"admin"},
					},
				},
			},
		},
	})

	testCases := []struct {
		role     string
		resource string
		action   string
		allowed  bool
	}{
		{"admin", "models", "read", true},
		{"admin", "models", "write", true},
		{"admin", "models", "delete", true},
		{"admin", "nodes", "read", true},
		{"admin", "nodes", "manage", true},
		{"user", "models", "read", true},
		{"user", "models", "write", false},
		{"user", "models", "delete", false},
		{"user", "nodes", "read", true},
		{"user", "nodes", "manage", false},
		{"guest", "models", "read", false},
		{"guest", "nodes", "read", false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_%s_%s", tc.role, tc.resource, tc.action), func(t *testing.T) {
			allowed := authz.IsAllowed(tc.role, tc.resource, tc.action, nil)
			assert.Equal(t, tc.allowed, allowed)
		})
	}
}

// testActionPermissions tests action-based permissions
func testActionPermissions(t *testing.T) {
	authz := security.NewAuthorization(&security.AuthorizationConfig{
		DefaultDeny: true,
	})

	// Add granular action permissions
	actions := map[string][]string{
		"model:create":   {"admin"},
		"model:read":     {"admin", "user"},
		"model:update":   {"admin"},
		"model:delete":   {"admin"},
		"model:download": {"admin", "user"},
		"node:join":      {"admin"},
		"node:leave":     {"admin", "self"},
		"cluster:status": {"admin", "user"},
		"cluster:config": {"admin"},
	}

	for action, allowedRoles := range actions {
		authz.AddActionPolicy(action, allowedRoles)
	}

	testCases := []struct {
		role    string
		action  string
		allowed bool
	}{
		{"admin", "model:create", true},
		{"admin", "model:read", true},
		{"admin", "node:join", true},
		{"admin", "cluster:config", true},
		{"user", "model:read", true},
		{"user", "model:download", true},
		{"user", "cluster:status", true},
		{"user", "model:create", false},
		{"user", "model:delete", false},
		{"user", "node:join", false},
		{"user", "cluster:config", false},
		{"guest", "model:read", false},
		{"guest", "cluster:status", false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_%s", tc.role, tc.action), func(t *testing.T) {
			allowed := authz.HasActionPermission(tc.role, tc.action)
			assert.Equal(t, tc.allowed, allowed)
		})
	}
}

// testConditionalAccess tests conditional access policies
func testConditionalAccess(t *testing.T) {
	authz := security.NewAuthorization(&security.AuthorizationConfig{
		ConditionalAccess: true,
	})

	// Add conditional policies
	policies := []*security.ConditionalPolicy{
		{
			Name: "time_based_access",
			Conditions: map[string]interface{}{
				"time_range": map[string]string{
					"start": "09:00",
					"end":   "17:00",
				},
				"days": []string{"monday", "tuesday", "wednesday", "thursday", "friday"},
			},
			Effect:  "allow",
			Actions: []string{"model:read", "model:download"},
		},
		{
			Name: "ip_based_access",
			Conditions: map[string]interface{}{
				"ip_ranges": []string{"10.0.0.0/8", "192.168.0.0/16"},
			},
			Effect:  "allow",
			Actions: []string{"*"},
		},
		{
			Name: "device_trust",
			Conditions: map[string]interface{}{
				"device_trust_level": "high",
			},
			Effect:  "allow",
			Actions: []string{"model:create", "model:delete"},
		},
	}

	for _, policy := range policies {
		authz.AddConditionalPolicy(policy)
	}

	testCases := []struct {
		name    string
		action  string
		context map[string]interface{}
		allowed bool
	}{
		{
			name:   "business_hours_access",
			action: "model:read",
			context: map[string]interface{}{
				"time": "14:30",
				"day":  "wednesday",
			},
			allowed: true,
		},
		{
			name:   "after_hours_access",
			action: "model:read",
			context: map[string]interface{}{
				"time": "20:30",
				"day":  "wednesday",
			},
			allowed: false,
		},
		{
			name:   "trusted_ip_access",
			action: "model:create",
			context: map[string]interface{}{
				"client_ip": "10.0.1.100",
			},
			allowed: true,
		},
		{
			name:   "untrusted_ip_access",
			action: "model:create",
			context: map[string]interface{}{
				"client_ip": "1.2.3.4",
			},
			allowed: false,
		},
		{
			name:   "high_trust_device",
			action: "model:delete",
			context: map[string]interface{}{
				"device_trust_level": "high",
			},
			allowed: true,
		},
		{
			name:   "low_trust_device",
			action: "model:delete",
			context: map[string]interface{}{
				"device_trust_level": "low",
			},
			allowed: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			allowed := authz.EvaluateConditionalAccess(tc.action, tc.context)
			assert.Equal(t, tc.allowed, allowed)
		})
	}
}

// testPermissionInheritance tests permission inheritance mechanisms
func testPermissionInheritance(t *testing.T) {
	authz := security.NewAuthorization(&security.AuthorizationConfig{
		PermissionInheritance: true,
		RoleHierarchy: map[string][]string{
			"super_admin": {"admin"},
			"admin":       {"user"},
			"user":        {"guest"},
		},
	})

	// Define base permissions
	basePermissions := map[string][]string{
		"guest":       {"model:read"},
		"user":        {"model:download", "cluster:status"},
		"admin":       {"model:create", "model:update", "model:delete", "node:manage"},
		"super_admin": {"cluster:config", "system:admin"},
	}

	for role, permissions := range basePermissions {
		for _, permission := range permissions {
			authz.GrantPermission(role, permission)
		}
	}

	t.Run("DirectPermissions", func(t *testing.T) {
		// Test direct permissions
		assert.True(t, authz.HasPermission("guest", "model:read"))
		assert.True(t, authz.HasPermission("user", "model:download"))
		assert.True(t, authz.HasPermission("admin", "model:create"))
		assert.True(t, authz.HasPermission("super_admin", "system:admin"))
	})

	t.Run("InheritedPermissions", func(t *testing.T) {
		// Test inherited permissions

		// User should inherit guest permissions
		assert.True(t, authz.HasPermission("user", "model:read"))

		// Admin should inherit user and guest permissions
		assert.True(t, authz.HasPermission("admin", "model:read"))
		assert.True(t, authz.HasPermission("admin", "model:download"))
		assert.True(t, authz.HasPermission("admin", "cluster:status"))

		// Super admin should inherit all permissions
		assert.True(t, authz.HasPermission("super_admin", "model:read"))
		assert.True(t, authz.HasPermission("super_admin", "model:download"))
		assert.True(t, authz.HasPermission("super_admin", "cluster:status"))
		assert.True(t, authz.HasPermission("super_admin", "model:create"))
		assert.True(t, authz.HasPermission("super_admin", "node:manage"))
	})

	t.Run("NoUpwardInheritance", func(t *testing.T) {
		// Lower roles should not inherit higher role permissions
		assert.False(t, authz.HasPermission("guest", "model:download"))
		assert.False(t, authz.HasPermission("guest", "model:create"))
		assert.False(t, authz.HasPermission("user", "model:create"))
		assert.False(t, authz.HasPermission("admin", "system:admin"))
	})
}

// TestSecurityIntegration tests integration between security components
func TestSecurityIntegration(t *testing.T) {
	t.Run("EndToEndAuth", testEndToEndAuth)
	t.Run("SecurityMiddleware", testSecurityMiddleware)
	t.Run("AuditLogging", testAuditLogging)
}

// testEndToEndAuth tests complete authentication flow
func testEndToEndAuth(t *testing.T) {
	// Create integrated security manager
	securityManager := security.NewSecurityManager(&security.Config{
		Auth: &security.AuthConfig{
			JWTSecret:   []byte("integration-test-secret"),
			TokenExpiry: time.Hour,
		},
		Authorization: &security.AuthorizationConfig{
			DefaultDeny: true,
		},
		Encryption: &security.EncryptionConfig{
			Algorithm: "AES-256-GCM",
		},
		Audit: &security.AuditConfig{
			Enabled:  true,
			LogLevel: "info",
		},
	})

	t.Run("LoginProcess", func(t *testing.T) {
		// Simulate user login
		credentials := &security.Credentials{
			Username: "testuser",
			Password: "testpassword",
		}

		// Authenticate user
		authResult, err := securityManager.Authenticate(credentials)
		assert.NoError(t, err)
		assert.NotNil(t, authResult)
		assert.NotEmpty(t, authResult.AccessToken)
		assert.NotEmpty(t, authResult.RefreshToken)

		// Validate token
		claims, err := securityManager.ValidateToken(authResult.AccessToken)
		assert.NoError(t, err)
		assert.Equal(t, "testuser", claims.UserID)
	})

	t.Run("AuthorizedAction", func(t *testing.T) {
		// Create user with specific role
		token, err := securityManager.GenerateToken("testuser", "user", map[string]interface{}{
			"permissions": []string{"model:read"},
		})
		require.NoError(t, err)

		// Validate token and check permissions
		claims, err := securityManager.ValidateToken(token)
		require.NoError(t, err)

		// Test authorized action
		allowed := securityManager.IsActionAllowed(claims, "model:read", "model-123", nil)
		assert.True(t, allowed)

		// Test unauthorized action
		allowed = securityManager.IsActionAllowed(claims, "model:delete", "model-123", nil)
		assert.False(t, allowed)
	})
}

// testSecurityMiddleware tests HTTP security middleware
func testSecurityMiddleware(t *testing.T) {
	securityManager := security.NewSecurityManager(&security.Config{
		Auth: &security.AuthConfig{
			JWTSecret: []byte("middleware-test-secret"),
		},
	})

	// Create test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	})

	// Wrap with security middleware
	secureHandler := securityManager.SecurityMiddleware(testHandler)

	t.Run("ValidToken", func(t *testing.T) {
		// Generate valid token
		token, err := securityManager.GenerateToken("testuser", "user", nil)
		require.NoError(t, err)

		// Create request with token
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		secureHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "Success", w.Body.String())
	})

	t.Run("InvalidToken", func(t *testing.T) {
		// Create request with invalid token
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")

		w := httptest.NewRecorder()
		secureHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("MissingToken", func(t *testing.T) {
		// Create request without token
		req := httptest.NewRequest("GET", "/test", nil)

		w := httptest.NewRecorder()
		secureHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// testAuditLogging tests security audit logging
func testAuditLogging(t *testing.T) {
	auditLogger := security.NewAuditLogger(&security.AuditConfig{
		Enabled:  true,
		LogLevel: "info",
		Fields:   []string{"timestamp", "user", "action", "resource", "result"},
	})

	t.Run("AuthenticationEvents", func(t *testing.T) {
		events := []security.AuditEvent{
			{
				Type:      "authentication",
				Action:    "login",
				UserID:    "testuser",
				Result:    "success",
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"ip_address": "192.168.1.100",
					"user_agent": "Mozilla/5.0...",
				},
			},
			{
				Type:      "authentication",
				Action:    "logout",
				UserID:    "testuser",
				Result:    "success",
				Timestamp: time.Now(),
			},
			{
				Type:      "authentication",
				Action:    "login",
				UserID:    "baduser",
				Result:    "failure",
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"error":      "invalid credentials",
					"ip_address": "192.168.1.200",
				},
			},
		}

		for _, event := range events {
			err := auditLogger.LogEvent(event)
			assert.NoError(t, err)
		}
	})

	t.Run("AuthorizationEvents", func(t *testing.T) {
		events := []security.AuditEvent{
			{
				Type:      "authorization",
				Action:    "model:read",
				UserID:    "testuser",
				Resource:  "model-123",
				Result:    "success",
				Timestamp: time.Now(),
			},
			{
				Type:      "authorization",
				Action:    "model:delete",
				UserID:    "testuser",
				Resource:  "model-123",
				Result:    "denied",
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"reason": "insufficient permissions",
				},
			},
		}

		for _, event := range events {
			err := auditLogger.LogEvent(event)
			assert.NoError(t, err)
		}
	})

	t.Run("QueryAuditLogs", func(t *testing.T) {
		// Query logs by user
		logs, err := auditLogger.QueryLogs(&security.AuditQuery{
			UserID:    "testuser",
			StartTime: time.Now().Add(-1 * time.Hour),
			EndTime:   time.Now(),
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, logs)

		// Query logs by action
		logs, err = auditLogger.QueryLogs(&security.AuditQuery{
			Action:    "login",
			StartTime: time.Now().Add(-1 * time.Hour),
			EndTime:   time.Now(),
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, logs)

		// Query failed events
		logs, err = auditLogger.QueryLogs(&security.AuditQuery{
			Result:    "failure",
			StartTime: time.Now().Add(-1 * time.Hour),
			EndTime:   time.Now(),
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, logs)
	})
}

// BenchmarkSecurity benchmarks security operations
func BenchmarkSecurity(b *testing.B) {
	authManager := security.NewAuthManager(&security.AuthConfig{
		JWTSecret:   []byte("benchmark-secret"),
		TokenExpiry: time.Hour,
	})

	b.Run("TokenGeneration", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				_, err := authManager.GenerateToken(fmt.Sprintf("user-%d", i), "user", nil)
				if err != nil {
					b.Fatal(err)
				}
				i++
			}
		})
	})

	// Generate token for validation benchmark
	token, _ := authManager.GenerateToken("benchuser", "user", nil)

	b.Run("TokenValidation", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := authManager.ValidateToken(token)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	encManager := security.NewEncryptionManager(&security.EncryptionConfig{
		Algorithm: "AES-256-GCM",
	})

	key, _ := encManager.GenerateKey()
	testData := make([]byte, 1024) // 1KB

	b.Run("DataEncryption", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := encManager.Encrypt(testData, key)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	encrypted, _ := encManager.Encrypt(testData, key)

	b.Run("DataDecryption", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := encManager.Decrypt(encrypted, key)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}

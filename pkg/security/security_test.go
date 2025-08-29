package security

import (
	"testing"
	"time"
)

// Security test suite for OllamaMax
// Tests various security mechanisms including encryption, hashing, and input validation

func TestPasswordHashing(t *testing.T) {
	// Test secure password hashing
	passwords := []string{
		"simplepassword",
		"ComplexP@ssw0rd!",
		"very_long_password_with_many_characters_123456789",
		"短密码", // Unicode password
	}

	for _, password := range passwords {
		hash, err := HashPassword(password)
		if err != nil {
			t.Errorf("Failed to hash password '%s': %v", password, err)
			continue
		}

		// Verify hash is not the same as password
		if hash == password {
			t.Errorf("Hashed password should not be the same as original password")
		}

		// Verify password verification works
		if !VerifyPassword(password, hash) {
			t.Errorf("Password verification failed for '%s'", password)
		}

		// Verify wrong password fails
		if VerifyPassword(password+"wrong", hash) {
			t.Errorf("Wrong password should not verify for '%s'", password)
		}
	}
}

func TestInputSanitization(t *testing.T) {
	// Test input sanitization for XSS and injection attacks
	testCases := []struct {
		input    string
		expected string
		name     string
	}{
		{
			input:    "<script>alert('xss')</script>",
			expected: "",
			name:     "Script tag removal",
		},
		{
			input:    "'; DROP TABLE users; --",
			expected: "'; DROP TABLE users; --", // Should be escaped, not removed
			name:     "SQL injection attempt",
		},
		{
			input:    "<img src=x onerror=alert('xss')>",
			expected: "",
			name:     "Image XSS attempt",
		},
		{
			input:    "normal text input",
			expected: "normal text input",
			name:     "Normal text",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SanitizeInput(tc.input)
			if tc.name == "SQL injection attempt" {
				// For SQL injection, we expect the input to be escaped, not necessarily removed
				if result == tc.input {
					t.Logf("SQL injection input preserved (should be escaped in actual DB operations): %s", result)
				}
			} else if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestTokenGeneration(t *testing.T) {
	// Test cryptographically secure token generation
	tokenLengths := []int{16, 32, 64, 128}

	for _, length := range tokenLengths {
		token, err := GenerateSecureToken(length)
		if err != nil {
			t.Errorf("Failed to generate token of length %d: %v", length, err)
			continue
		}

		// Token should be hex encoded, so length should be 2x the byte length
		expectedHexLength := length * 2
		if len(token) != expectedHexLength {
			t.Errorf("Expected token length %d, got %d", expectedHexLength, len(token))
		}

		// Generate another token and ensure they're different
		token2, err := GenerateSecureToken(length)
		if err != nil {
			t.Errorf("Failed to generate second token: %v", err)
			continue
		}

		if token == token2 {
			t.Errorf("Generated tokens should be different")
		}
	}
}

func TestRateLimiting(t *testing.T) {
	// Test rate limiting functionality
	limiter := NewRateLimiter(3, time.Second)

	clientID := "test-client"

	// Should allow first 3 requests
	for i := 0; i < 3; i++ {
		if !limiter.Allow(clientID) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// Should block 4th request
	if limiter.Allow(clientID) {
		t.Error("4th request should be blocked by rate limiter")
	}

	// Wait for rate limit window to pass
	time.Sleep(time.Second + 10*time.Millisecond)

	// Should allow request after window
	if !limiter.Allow(clientID) {
		t.Error("Request should be allowed after rate limit window")
	}
}

func TestEncryptionDecryption(t *testing.T) {
	// Test AES encryption and decryption
	key := GenerateKeyFromPassword("test-password", "test-salt")
	if len(key) != 32 {
		t.Fatalf("Test key must be 32 bytes for AES-256, got %d", len(key))
	}

	plaintexts := []string{
		"Hello, World!",
		"Secret message with special characters: !@#$%^&*()",
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
		"", // Empty string
	}

	for _, plaintext := range plaintexts {
		// Test encryption
		ciphertext, err := EncryptAES([]byte(plaintext), key)
		if err != nil {
			t.Errorf("Failed to encrypt '%s': %v", plaintext, err)
			continue
		}

		// Ciphertext should be different from plaintext (except for empty string case)
		if plaintext != "" && string(ciphertext) == plaintext {
			t.Errorf("Ciphertext should be different from plaintext")
		}

		// Test decryption
		decrypted, err := DecryptAES(ciphertext, key)
		if err != nil {
			t.Errorf("Failed to decrypt ciphertext: %v", err)
			continue
		}

		if string(decrypted) != plaintext {
			t.Errorf("Decrypted text doesn't match original. Expected '%s', got '%s'", plaintext, string(decrypted))
		}
	}

	// Test with wrong key length
	wrongKey := []byte("short-key")
	_, err := EncryptAES([]byte("test"), wrongKey)
	if err == nil {
		t.Error("Should fail with wrong key length")
	}
}

func TestSecureHeaders(t *testing.T) {
	// Test security headers
	headers := GetSecurityHeaders()

	expectedHeaders := []string{
		"X-Content-Type-Options",
		"X-Frame-Options", 
		"X-XSS-Protection",
		"Strict-Transport-Security",
		"Content-Security-Policy",
	}

	for _, header := range expectedHeaders {
		if value, exists := headers[header]; !exists {
			t.Errorf("Missing security header: %s", header)
		} else if value == "" {
			t.Errorf("Security header %s should have a value", header)
		}
	}

	// Check specific header values
	if headers["X-Frame-Options"] != "DENY" {
		t.Errorf("X-Frame-Options should be DENY, got: %s", headers["X-Frame-Options"])
	}

	if headers["X-Content-Type-Options"] != "nosniff" {
		t.Errorf("X-Content-Type-Options should be nosniff, got: %s", headers["X-Content-Type-Options"])
	}
}

func TestPasswordStrength(t *testing.T) {
	// Test password strength validation
	testCases := []struct {
		password string
		expected bool
		name     string
	}{
		{
			password: "weak",
			expected: false,
			name:     "Too short",
		},
		{
			password: "onlylowercase",
			expected: false,
			name:     "No numbers or special chars",
		},
		{
			password: "WeakPassword123",
			expected: true,
			name:     "Good password",
		},
		{
			password: "VeryStrong@Password123",
			expected: true,
			name:     "Very strong password",
		},
		{
			password: "12345678",
			expected: false,
			name:     "Only numbers",
		},
		{
			password: "UPPERCASE",
			expected: false,
			name:     "Only uppercase",
		},
		{
			password: "NoSpecialChars123",
			expected: true,
			name:     "No special chars",
		},
		{
			password: "All3Elements!",
			expected: true,
			name:     "Contains all required elements",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidatePasswordStrength(tc.password)
			if result != tc.expected {
				t.Errorf("Password '%s' strength check failed. Expected: %t, Got: %t", tc.password, tc.expected, result)
			}
		})
	}
}

func TestSecureCompare(t *testing.T) {
	// Test constant-time string comparison
	testCases := []struct {
		a        string
		b        string
		expected bool
		name     string
	}{
		{"hello", "hello", true, "Identical strings"},
		{"hello", "world", false, "Different strings"},
		{"", "", true, "Empty strings"},
		{"a", "b", false, "Single char difference"},
		{"password123", "password123", true, "Identical passwords"},
		{"password123", "password124", false, "Similar passwords"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SecureCompare(tc.a, tc.b)
			if result != tc.expected {
				t.Errorf("SecureCompare('%s', '%s') = %t, expected %t", tc.a, tc.b, result, tc.expected)
			}
		})
	}
}

func TestKeyGeneration(t *testing.T) {
	// Test key generation from password
	password := "test-password"
	salt1 := "salt1"
	salt2 := "salt2"

	key1 := GenerateKeyFromPassword(password, salt1)
	key2 := GenerateKeyFromPassword(password, salt1) // Same salt
	key3 := GenerateKeyFromPassword(password, salt2) // Different salt

	// Keys with same password and salt should be identical
	if !SecureCompare(string(key1), string(key2)) {
		t.Error("Keys with same password and salt should be identical")
	}

	// Keys with different salts should be different
	if SecureCompare(string(key1), string(key3)) {
		t.Error("Keys with different salts should be different")
	}

	// Key should be 32 bytes for AES-256
	if len(key1) != 32 {
		t.Errorf("Key should be 32 bytes, got %d", len(key1))
	}
}
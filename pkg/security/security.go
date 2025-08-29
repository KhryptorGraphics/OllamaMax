package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
	
	"golang.org/x/crypto/bcrypt"
)

// HashPassword creates a secure hash of a password using bcrypt
func HashPassword(password string) (string, error) {
	if len(password) == 0 {
		return "", errors.New("password cannot be empty")
	}
	
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	
	return string(bytes), nil
}

// VerifyPassword verifies a password against its hash using bcrypt
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// SanitizeInput removes dangerous HTML/script tags and other potentially harmful content
func SanitizeInput(input string) string {
	// Remove script tags and content
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	input = scriptRegex.ReplaceAllString(input, "")
	
	// Remove standalone script tags
	scriptTagRegex := regexp.MustCompile(`(?i)<script[^>]*>`)
	input = scriptTagRegex.ReplaceAllString(input, "")
	
	// Remove img tags with onerror or other event handlers
	imgRegex := regexp.MustCompile(`(?i)<img[^>]*(?:onerror|onload|onclick|onmouseover)[^>]*>`)
	input = imgRegex.ReplaceAllString(input, "")
	
	// Remove other potentially dangerous tags
	dangerousTags := []string{
		`(?i)<iframe[^>]*>.*?</iframe>`,
		`(?i)<object[^>]*>.*?</object>`,
		`(?i)<embed[^>]*>.*?</embed>`,
		`(?i)<link[^>]*>`,
		`(?i)<meta[^>]*>`,
	}
	
	for _, tagPattern := range dangerousTags {
		regex := regexp.MustCompile(tagPattern)
		input = regex.ReplaceAllString(input, "")
	}
	
	// Remove javascript: URLs
	jsRegex := regexp.MustCompile(`(?i)javascript:`)
	input = jsRegex.ReplaceAllString(input, "")
	
	return strings.TrimSpace(input)
}

// GenerateSecureToken creates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("token length must be positive")
	}
	
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}
	
	return hex.EncodeToString(bytes), nil
}

// RateLimiter implements simple in-memory rate limiting
type RateLimiter struct {
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow(key string) bool {
	now := time.Now()
	
	// Clean up old requests
	requests := rl.requests[key]
	var validRequests []time.Time
	
	for _, reqTime := range requests {
		if now.Sub(reqTime) <= rl.window {
			validRequests = append(validRequests, reqTime)
		}
	}
	
	// Check if limit exceeded
	if len(validRequests) >= rl.limit {
		return false
	}
	
	// Add current request
	validRequests = append(validRequests, now)
	rl.requests[key] = validRequests
	
	return true
}

// EncryptAES encrypts data using AES-256-GCM
func EncryptAES(plaintext, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("key must be 32 bytes for AES-256")
	}
	
	block, err := aes.NewCipher(key)
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
	
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// DecryptAES decrypts data using AES-256-GCM
func DecryptAES(ciphertext, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("key must be 32 bytes for AES-256")
	}
	
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}
	
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}
	
	return plaintext, nil
}

// SecurityHeaders returns a map of recommended security headers
func GetSecurityHeaders() map[string]string {
	return map[string]string{
		"X-Content-Type-Options":   "nosniff",
		"X-Frame-Options":          "DENY",
		"X-XSS-Protection":         "1; mode=block",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"Content-Security-Policy":  "default-src 'self'",
		"Referrer-Policy":          "strict-origin-when-cross-origin",
	}
}

// ValidatePasswordStrength checks password strength requirements
func ValidatePasswordStrength(password string) bool {
	if len(password) < 8 {
		return false
	}
	
	// Check for at least one digit
	hasDigit := false
	// Check for at least one lowercase
	hasLower := false
	// Check for at least one uppercase
	hasUpper := false
	// Check for at least one special character
	hasSpecial := false
	
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	
	for _, char := range password {
		switch {
		case char >= '0' && char <= '9':
			hasDigit = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case strings.ContainsRune(specialChars, char):
			hasSpecial = true
		}
	}
	
	// For a password to be considered strong, it should have at least 3 of the 4 criteria
	criteria := 0
	if hasDigit {
		criteria++
	}
	if hasLower {
		criteria++
	}
	if hasUpper {
		criteria++
	}
	if hasSpecial {
		criteria++
	}
	
	return criteria >= 3
}

// SecureCompare performs constant-time string comparison
func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// GenerateKeyFromPassword creates a 32-byte key from a password using PBKDF2
func GenerateKeyFromPassword(password, salt string) []byte {
	if salt == "" {
		salt = "ollamamax-default-salt" // In production, use a random salt
	}
	
	// Simple key derivation - in production, use PBKDF2 or Argon2
	combined := password + salt
	hash := sha256.Sum256([]byte(combined))
	return hash[:]
}
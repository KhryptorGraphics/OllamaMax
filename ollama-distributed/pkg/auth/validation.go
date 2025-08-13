package auth

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"unicode"
)

// InputValidator provides comprehensive input validation and sanitization
type InputValidator struct {
	maxStringLength int
	maxArrayLength  int
}

// NewInputValidator creates a new input validator with default limits
func NewInputValidator() *InputValidator {
	return &InputValidator{
		maxStringLength: 1000, // Maximum string length
		maxArrayLength:  100,  // Maximum array length
	}
}

// ValidateUsername validates a username according to security best practices
func (v *InputValidator) ValidateUsername(username string) error {
	if len(username) == 0 {
		return fmt.Errorf("username is required")
	}

	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters long")
	}

	if len(username) > 50 {
		return fmt.Errorf("username must be less than 50 characters long")
	}

	// Check for valid characters (alphanumeric, underscore, hyphen, dot)
	validUsername := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !validUsername.MatchString(username) {
		return fmt.Errorf("username can only contain letters, numbers, periods, underscores, and hyphens")
	}

	// Prevent usernames that could be confused with system accounts
	restrictedUsernames := []string{
		"admin", "root", "system", "daemon", "bin", "sys", "sync", "games",
		"man", "lp", "mail", "news", "uucp", "proxy", "www-data", "backup",
		"list", "irc", "gnats", "nobody", "systemd", "messagebus", "sshd",
		"api", "service", "test", "guest", "anonymous", "null", "void",
	}

	lowerUsername := strings.ToLower(username)
	for _, restricted := range restrictedUsernames {
		if lowerUsername == restricted {
			return fmt.Errorf("username '%s' is not allowed", username)
		}
	}

	return nil
}

// ValidateEmail validates an email address
func (v *InputValidator) ValidateEmail(email string) error {
	if len(email) == 0 {
		return fmt.Errorf("email is required")
	}

	if len(email) > 254 {
		return fmt.Errorf("email address is too long")
	}

	// Use Go's mail package for basic validation
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email format")
	}

	// Additional checks for security
	if strings.Contains(email, "..") {
		return fmt.Errorf("email contains invalid consecutive dots")
	}

	// Check for dangerous characters that could be used in injection attacks
	dangerousChars := []string{"<", ">", "\"", "'", "&", ";", "|", "`", "$"}
	for _, char := range dangerousChars {
		if strings.Contains(email, char) {
			return fmt.Errorf("email contains invalid characters")
		}
	}

	return nil
}

// ValidatePassword validates password strength
func (v *InputValidator) ValidatePassword(password string) error {
	if len(password) == 0 {
		return fmt.Errorf("password is required")
	}

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	if len(password) > 128 {
		return fmt.Errorf("password must be less than 128 characters long")
	}

	// Check for character variety
	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}

	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	// Check for common weak passwords
	weakPasswords := []string{
		"password", "123456", "password123", "admin", "qwerty",
		"letmein", "welcome", "monkey", "dragon", "master",
		"password1", "123456789", "12345678", "qwerty123",
	}

	lowerPassword := strings.ToLower(password)
	for _, weak := range weakPasswords {
		if lowerPassword == weak {
			return fmt.Errorf("password is too common and easily guessable")
		}
		if strings.Contains(lowerPassword, weak) && len(weak) > 6 {
			return fmt.Errorf("password contains common weak patterns")
		}
	}

	return nil
}

// ValidateModelName validates model names for security
func (v *InputValidator) ValidateModelName(modelName string) error {
	if len(modelName) == 0 {
		return fmt.Errorf("model name is required")
	}

	if len(modelName) > 100 {
		return fmt.Errorf("model name must be less than 100 characters long")
	}

	// Allow letters, numbers, hyphens, underscores, periods, and forward slashes (for namespaces)
	validModelName := regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`)
	if !validModelName.MatchString(modelName) {
		return fmt.Errorf("model name can only contain letters, numbers, periods, underscores, hyphens, and forward slashes")
	}

	// Prevent path traversal attempts
	if strings.Contains(modelName, "..") {
		return fmt.Errorf("model name contains invalid path traversal sequence")
	}

	if strings.HasPrefix(modelName, "/") || strings.HasSuffix(modelName, "/") {
		return fmt.Errorf("model name cannot start or end with forward slash")
	}

	// Prevent system file references
	systemPaths := []string{"/etc/", "/var/", "/usr/", "/bin/", "/sbin/", "/root/", "/home/"}
	lowerModelName := strings.ToLower(modelName)
	for _, systemPath := range systemPaths {
		if strings.Contains(lowerModelName, systemPath) {
			return fmt.Errorf("model name cannot reference system paths")
		}
	}

	return nil
}

// ValidateNodeID validates node identifiers
func (v *InputValidator) ValidateNodeID(nodeID string) error {
	if len(nodeID) == 0 {
		return fmt.Errorf("node ID is required")
	}

	if len(nodeID) > 64 {
		return fmt.Errorf("node ID must be less than 64 characters long")
	}

	// Node IDs should be alphanumeric with hyphens
	validNodeID := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	if !validNodeID.MatchString(nodeID) {
		return fmt.Errorf("node ID can only contain letters, numbers, and hyphens")
	}

	if strings.HasPrefix(nodeID, "-") || strings.HasSuffix(nodeID, "-") {
		return fmt.Errorf("node ID cannot start or end with hyphen")
	}

	return nil
}

// ValidateAPIKeyName validates API key names
func (v *InputValidator) ValidateAPIKeyName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("API key name is required")
	}

	if len(name) > 50 {
		return fmt.Errorf("API key name must be less than 50 characters long")
	}

	// Allow letters, numbers, spaces, hyphens, underscores
	validName := regexp.MustCompile(`^[a-zA-Z0-9 _-]+$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("API key name can only contain letters, numbers, spaces, underscores, and hyphens")
	}

	return nil
}

// SanitizeString sanitizes a string for safe storage and display
func (v *InputValidator) SanitizeString(input string) string {
	// Remove null bytes
	sanitized := strings.ReplaceAll(input, "\x00", "")

	// Trim whitespace
	sanitized = strings.TrimSpace(sanitized)

	// Limit length
	if len(sanitized) > v.maxStringLength {
		sanitized = sanitized[:v.maxStringLength]
	}

	return sanitized
}

// ValidateJSONInput validates JSON input structure for API endpoints
func (v *InputValidator) ValidateJSONInput(data map[string]interface{}) error {
	// Check for excessive nesting depth
	if err := v.checkNestingDepth(data, 0, 10); err != nil {
		return err
	}

	// Check for dangerous keys that could indicate injection attempts
	dangerousKeys := []string{
		"__proto__", "constructor", "prototype", "eval", "function",
		"script", "javascript", "vbscript", "onload", "onerror",
	}

	return v.checkDangerousKeys(data, dangerousKeys)
}

// checkNestingDepth recursively checks JSON nesting depth
func (v *InputValidator) checkNestingDepth(data interface{}, currentDepth, maxDepth int) error {
	if currentDepth > maxDepth {
		return fmt.Errorf("JSON input exceeds maximum nesting depth of %d", maxDepth)
	}

	switch value := data.(type) {
	case map[string]interface{}:
		for _, item := range value {
			if err := v.checkNestingDepth(item, currentDepth+1, maxDepth); err != nil {
				return err
			}
		}
	case []interface{}:
		if len(value) > v.maxArrayLength {
			return fmt.Errorf("array exceeds maximum length of %d", v.maxArrayLength)
		}
		for _, item := range value {
			if err := v.checkNestingDepth(item, currentDepth+1, maxDepth); err != nil {
				return err
			}
		}
	}

	return nil
}

// checkDangerousKeys recursively checks for dangerous keys in JSON
func (v *InputValidator) checkDangerousKeys(data interface{}, dangerousKeys []string) error {
	switch value := data.(type) {
	case map[string]interface{}:
		for key, mapValue := range value {
			lowerKey := strings.ToLower(key)
			for _, dangerous := range dangerousKeys {
				if lowerKey == dangerous || strings.Contains(lowerKey, dangerous) {
					return fmt.Errorf("input contains potentially dangerous key: %s", key)
				}
			}
			if err := v.checkDangerousKeys(mapValue, dangerousKeys); err != nil {
				return err
			}
		}
	case []interface{}:
		for _, item := range value {
			if err := v.checkDangerousKeys(item, dangerousKeys); err != nil {
				return err
			}
		}
	case string:
		// Check for script injection patterns in string values
		lowerValue := strings.ToLower(value)
		scriptPatterns := []string{
			"<script", "javascript:", "vbscript:", "data:text/html",
			"eval(", "function(", "alert(", "prompt(", "confirm(",
		}
		for _, pattern := range scriptPatterns {
			if strings.Contains(lowerValue, pattern) {
				return fmt.Errorf("input contains potentially dangerous script pattern")
			}
		}
	}

	return nil
}

// ValidateIPAddress validates IP addresses for allowlists/blocklists
func (v *InputValidator) ValidateIPAddress(ip string) error {
	if len(ip) == 0 {
		return fmt.Errorf("IP address is required")
	}

	if len(ip) > 45 { // Max length for IPv6
		return fmt.Errorf("IP address is too long")
	}

	// Basic IP validation regex (supports both IPv4 and IPv6)
	ipv4Pattern := regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
	ipv6Pattern := regexp.MustCompile(`^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$`)

	if !ipv4Pattern.MatchString(ip) && !ipv6Pattern.MatchString(ip) {
		return fmt.Errorf("invalid IP address format")
	}

	return nil
}

// Helper function for minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

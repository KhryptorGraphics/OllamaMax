package security

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

// ValidateModelName validates model names for security
func ValidateModelName(name string) error {
	if len(name) == 0 || len(name) > 255 {
		return errors.New("invalid model name length")
	}

	// Check for dangerous patterns
	dangerousPatterns := []string{
		"'", "\"", ";", "--", "/*", "*/",
		"DROP", "DELETE", "UPDATE", "INSERT", "UNION", "SELECT",
		"<script", "javascript:", "vbscript:", "data:",
		"eval(", "function(", "alert(", "prompt(", "confirm(",
	}
	upperName := strings.ToUpper(name)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(upperName, strings.ToUpper(pattern)) {
			return fmt.Errorf("invalid characters in model name: contains '%s'", pattern)
		}
	}

	// Check for path traversal
	if strings.Contains(name, "../") || strings.Contains(name, "..\\") {
		return errors.New("path traversal detected in model name")
	}

	// Validate format (alphanumeric, hyphens, underscores, dots, colons for namespaces)
	validName := regexp.MustCompile(`^[a-zA-Z0-9._:-]+$`)
	if !validName.MatchString(name) {
		return errors.New("model name contains invalid characters")
	}

	return nil
}

// ValidateNodeID validates node IDs for security
func ValidateNodeID(nodeID string) error {
	if len(nodeID) == 0 || len(nodeID) > 128 {
		return errors.New("invalid node ID length")
	}

	// Node IDs should be alphanumeric with hyphens
	validNodeID := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	if !validNodeID.MatchString(nodeID) {
		return errors.New("node ID contains invalid characters")
	}

	return nil
}

// ValidateTransferID validates transfer IDs for security
func ValidateTransferID(transferID string) error {
	if len(transferID) == 0 || len(transferID) > 128 {
		return errors.New("invalid transfer ID length")
	}

	// Transfer IDs should be alphanumeric with hyphens
	if !regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`).MatchString(transferID) {
		return errors.New("transfer ID contains invalid characters")
	}

	return nil
}

// ValidateAPIKey validates API keys for security
func ValidateAPIKey(apiKey string) error {
	if len(apiKey) < 32 || len(apiKey) > 256 {
		return errors.New("invalid API key length")
	}

	// API keys should be base64-like characters
	validAPIKey := regexp.MustCompile(`^[a-zA-Z0-9+/=_-]+$`)
	if !validAPIKey.MatchString(apiKey) {
		return errors.New("API key contains invalid characters")
	}

	return nil
}

// ValidateAPIInput validates general API input for security
func ValidateAPIInput(input string) error {
	if len(input) > 10000 { // Prevent DoS via large inputs
		return errors.New("input too large")
	}

	// Check for script injection patterns
	scriptPatterns := []string{
		"<script", "javascript:", "vbscript:", "data:text/html",
		"eval(", "function(", "alert(", "prompt(", "confirm(",
		"onload=", "onerror=", "onclick=", "onmouseover=",
	}
	lowerInput := strings.ToLower(input)
	for _, pattern := range scriptPatterns {
		if strings.Contains(lowerInput, pattern) {
			return fmt.Errorf("potentially dangerous script pattern detected: %s", pattern)
		}
	}

	return nil
}

// ValidateURL validates URLs for security
func ValidateURL(urlStr string) error {
	if len(urlStr) == 0 || len(urlStr) > 2048 {
		return errors.New("invalid URL length")
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Only allow HTTP and HTTPS schemes
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("invalid URL scheme: %s", parsedURL.Scheme)
	}

	// Prevent localhost/private IP access in production
	if strings.Contains(parsedURL.Host, "localhost") ||
		strings.Contains(parsedURL.Host, "127.0.0.1") ||
		strings.Contains(parsedURL.Host, "::1") {
		return errors.New("localhost URLs not allowed")
	}

	return nil
}

// ValidateFilePath validates file paths for security
func ValidateFilePath(path string) error {
	if len(path) == 0 || len(path) > 1024 {
		return errors.New("invalid file path length")
	}

	// Check for path traversal
	if strings.Contains(path, "../") || strings.Contains(path, "..\\") {
		return errors.New("path traversal detected")
	}

	// Check for null bytes
	if strings.Contains(path, "\x00") {
		return errors.New("null bytes not allowed in file path")
	}

	// Validate characters
	for _, r := range path {
		if r < 32 && r != '\t' && r != '\n' {
			return errors.New("invalid control characters in file path")
		}
	}

	return nil
}

// SanitizeInput sanitizes input by removing dangerous patterns
func SanitizeInput(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove control characters except newline and tab
	var result strings.Builder
	for _, r := range input {
		if r >= 32 || r == '\n' || r == '\t' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// SanitizeHTML sanitizes HTML input to prevent XSS
func SanitizeHTML(input string) string {
	// Remove script tags
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	input = scriptRegex.ReplaceAllString(input, "")

	// Remove javascript: URLs
	jsRegex := regexp.MustCompile(`(?i)javascript:`)
	input = jsRegex.ReplaceAllString(input, "")

	// Remove on* event handlers
	eventRegex := regexp.MustCompile(`(?i)\s*on\w+\s*=\s*[^>]*`)
	input = eventRegex.ReplaceAllString(input, "")

	return input
}

// ValidateJSONInput validates JSON input for security
func ValidateJSONInput(input []byte) error {
	if len(input) > 1024*1024 { // 1MB limit
		return errors.New("JSON input too large")
	}

	// Check for potential JSON injection patterns
	inputStr := string(input)
	dangerousPatterns := []string{
		"__proto__", "constructor", "prototype",
		"eval", "function", "setTimeout", "setInterval",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(inputStr, pattern) {
			return fmt.Errorf("potentially dangerous JSON pattern detected: %s", pattern)
		}
	}

	return nil
}

// ValidatePrompt validates AI prompts for security
func ValidatePrompt(prompt string) error {
	if len(prompt) == 0 {
		return errors.New("prompt cannot be empty")
	}

	if len(prompt) > 100000 { // 100KB limit
		return errors.New("prompt too large")
	}

	// Check for prompt injection patterns
	injectionPatterns := []string{
		"ignore previous instructions",
		"disregard the above",
		"forget everything",
		"new instructions:",
		"system:",
		"admin:",
		"root:",
	}

	lowerPrompt := strings.ToLower(prompt)
	for _, pattern := range injectionPatterns {
		if strings.Contains(lowerPrompt, pattern) {
			return fmt.Errorf("potential prompt injection detected: %s", pattern)
		}
	}

	return nil
}

// ValidateUsername validates usernames for security
func ValidateUsername(username string) error {
	if len(username) < 3 || len(username) > 64 {
		return errors.New("username must be between 3 and 64 characters")
	}

	// Username should be alphanumeric with underscores and hyphens
	validUsername := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validUsername.MatchString(username) {
		return errors.New("username contains invalid characters")
	}

	// Check for reserved usernames
	reservedNames := []string{
		"admin", "root", "system", "api", "www", "mail", "ftp",
		"test", "guest", "anonymous", "null", "undefined",
	}

	lowerUsername := strings.ToLower(username)
	for _, reserved := range reservedNames {
		if lowerUsername == reserved {
			return fmt.Errorf("username '%s' is reserved", username)
		}
	}

	return nil
}

// ValidatePassword validates passwords for security
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	if len(password) > 128 {
		return errors.New("password too long")
	}

	// Check for at least one uppercase, lowercase, digit, and special character
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return errors.New("password must contain at least one digit")
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	return nil
}

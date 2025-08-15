package security

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

// ComprehensiveInputValidator provides extensive input validation for all API endpoints
type ComprehensiveInputValidator struct {
	// Validation rules
	rules map[string]*ValidationRule

	// Global settings
	maxInputLength     int
	maxJSONDepth       int
	maxArrayLength     int
	enableSanitization bool
}

// ValidationRule defines validation criteria for specific input types
type ValidationRule struct {
	Pattern     *regexp.Regexp
	MinLength   int
	MaxLength   int
	Required    bool
	AllowEmpty  bool
	Sanitize    bool
	Description string
}

// ValidationResult contains the result of input validation
type ValidationResult struct {
	Valid          bool     `json:"valid"`
	Errors         []string `json:"errors,omitempty"`
	Warnings       []string `json:"warnings,omitempty"`
	SanitizedValue string   `json:"sanitized_value,omitempty"`
}

// NewComprehensiveInputValidator creates a new comprehensive input validator
func NewComprehensiveInputValidator() *ComprehensiveInputValidator {
	validator := &ComprehensiveInputValidator{
		rules:              make(map[string]*ValidationRule),
		maxInputLength:     10000,
		maxJSONDepth:       10,
		maxArrayLength:     1000,
		enableSanitization: true,
	}

	validator.initializeDefaultRules()
	return validator
}

// initializeDefaultRules sets up default validation rules
func (civ *ComprehensiveInputValidator) initializeDefaultRules() {
	// Model name validation
	civ.rules["model_name"] = &ValidationRule{
		Pattern:     regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`),
		MinLength:   1,
		MaxLength:   255,
		Required:    true,
		AllowEmpty:  false,
		Sanitize:    true,
		Description: "Model names must contain only alphanumeric characters, dots, underscores, slashes, and hyphens",
	}

	// Node ID validation
	civ.rules["node_id"] = &ValidationRule{
		Pattern:     regexp.MustCompile(`^[a-zA-Z0-9-]+$`),
		MinLength:   1,
		MaxLength:   64,
		Required:    true,
		AllowEmpty:  false,
		Sanitize:    false,
		Description: "Node IDs must contain only alphanumeric characters and hyphens",
	}

	// API endpoint validation
	civ.rules["api_endpoint"] = &ValidationRule{
		Pattern:     regexp.MustCompile(`^/[a-zA-Z0-9/_-]*$`),
		MinLength:   1,
		MaxLength:   200,
		Required:    true,
		AllowEmpty:  false,
		Sanitize:    true,
		Description: "API endpoints must start with / and contain only valid URL characters",
	}

	// User input validation (for prompts, etc.)
	civ.rules["user_input"] = &ValidationRule{
		Pattern:     nil, // More flexible for user content
		MinLength:   0,
		MaxLength:   50000,
		Required:    false,
		AllowEmpty:  true,
		Sanitize:    true,
		Description: "User input with basic sanitization",
	}

	// File path validation
	civ.rules["file_path"] = &ValidationRule{
		Pattern:     regexp.MustCompile(`^[a-zA-Z0-9/._-]+$`),
		MinLength:   1,
		MaxLength:   1024,
		Required:    true,
		AllowEmpty:  false,
		Sanitize:    true,
		Description: "File paths must contain only safe characters",
	}

	// Email validation
	civ.rules["email"] = &ValidationRule{
		Pattern:     regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
		MinLength:   5,
		MaxLength:   254,
		Required:    true,
		AllowEmpty:  false,
		Sanitize:    false,
		Description: "Valid email address format",
	}

	// URL validation
	civ.rules["url"] = &ValidationRule{
		Pattern:     regexp.MustCompile(`^https?://[a-zA-Z0-9.-]+[a-zA-Z0-9/._-]*$`),
		MinLength:   7,
		MaxLength:   2048,
		Required:    true,
		AllowEmpty:  false,
		Sanitize:    true,
		Description: "Valid HTTP/HTTPS URL",
	}

	// UUID validation
	civ.rules["uuid"] = &ValidationRule{
		Pattern:     regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`),
		MinLength:   36,
		MaxLength:   36,
		Required:    true,
		AllowEmpty:  false,
		Sanitize:    false,
		Description: "Valid UUID format",
	}

	// Alphanumeric validation
	civ.rules["alphanumeric"] = &ValidationRule{
		Pattern:     regexp.MustCompile(`^[a-zA-Z0-9]+$`),
		MinLength:   1,
		MaxLength:   100,
		Required:    true,
		AllowEmpty:  false,
		Sanitize:    false,
		Description: "Alphanumeric characters only",
	}

	// Safe text validation (for descriptions, etc.)
	civ.rules["safe_text"] = &ValidationRule{
		Pattern:     regexp.MustCompile(`^[a-zA-Z0-9\s.,!?;:()\[\]{}'"_-]+$`),
		MinLength:   0,
		MaxLength:   5000,
		Required:    false,
		AllowEmpty:  true,
		Sanitize:    true,
		Description: "Safe text with basic punctuation",
	}
}

// ValidateInput validates input against a specific rule
func (civ *ComprehensiveInputValidator) ValidateInput(input string, ruleType string) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}

	// Get validation rule
	rule, exists := civ.rules[ruleType]
	if !exists {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Unknown validation rule type: %s", ruleType))
		return result
	}

	// Check if input is required
	if rule.Required && len(input) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "Input is required but empty")
		return result
	}

	// Allow empty if configured
	if len(input) == 0 && rule.AllowEmpty {
		return result
	}

	// Check length constraints
	if len(input) < rule.MinLength {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Input too short (min: %d, actual: %d)", rule.MinLength, len(input)))
	}

	if len(input) > rule.MaxLength {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Input too long (max: %d, actual: %d)", rule.MaxLength, len(input)))
	}

	// Check global length limit
	if len(input) > civ.maxInputLength {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Input exceeds global limit (%d characters)", civ.maxInputLength))
	}

	// Check pattern if defined
	if rule.Pattern != nil && !rule.Pattern.MatchString(input) {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Input format invalid: %s", rule.Description))
	}

	// Check for dangerous patterns
	if civ.containsDangerousPatterns(input) {
		result.Valid = false
		result.Errors = append(result.Errors, "Input contains potentially dangerous patterns")
	}

	// Sanitize if enabled and valid
	if result.Valid && rule.Sanitize && civ.enableSanitization {
		result.SanitizedValue = civ.sanitizeInput(input)
		if result.SanitizedValue != input {
			result.Warnings = append(result.Warnings, "Input was sanitized")
		}
	} else {
		result.SanitizedValue = input
	}

	return result
}

// containsDangerousPatterns checks for dangerous patterns across all input
func (civ *ComprehensiveInputValidator) containsDangerousPatterns(input string) bool {
	dangerousPatterns := []string{
		// Script injection
		"<script", "javascript:", "vbscript:", "data:text/html",
		"eval(", "setTimeout(", "setInterval(",

		// SQL injection
		"'", "\"", "--", "/*", "*/", ";",
		"union", "select", "insert", "update", "delete", "drop",

		// Command injection
		"|", "&", ";", "`", "$(",

		// Path traversal
		"../", "..\\",

		// Null bytes
		"\x00",

		// LDAP injection
		"*", "(", ")", "\\",

		// XPath injection
		"'", "\"", "and", "or",
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}

	return false
}

// sanitizeInput sanitizes input by removing or escaping dangerous characters
func (civ *ComprehensiveInputValidator) sanitizeInput(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove or escape HTML/script tags
	input = strings.ReplaceAll(input, "<script", "&lt;script")
	input = strings.ReplaceAll(input, "</script>", "&lt;/script&gt;")
	input = strings.ReplaceAll(input, "javascript:", "")
	input = strings.ReplaceAll(input, "vbscript:", "")

	// Remove excessive whitespace
	input = regexp.MustCompile(`\s+`).ReplaceAllString(input, " ")
	input = strings.TrimSpace(input)

	// Remove control characters except tab, newline, carriage return
	var sanitized strings.Builder
	for _, r := range input {
		if unicode.IsPrint(r) || r == '\t' || r == '\n' || r == '\r' {
			sanitized.WriteRune(r)
		}
	}

	return sanitized.String()
}

// ValidateJSON validates JSON input for structure and content
func (civ *ComprehensiveInputValidator) ValidateJSON(jsonData []byte) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}

	// Check JSON size
	if len(jsonData) > civ.maxInputLength {
		result.Valid = false
		result.Errors = append(result.Errors, "JSON data too large")
		return result
	}

	// Parse JSON to check structure
	var data interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Invalid JSON: %v", err))
		return result
	}

	// Check nesting depth
	if civ.getJSONDepth(data) > civ.maxJSONDepth {
		result.Valid = false
		result.Errors = append(result.Errors, "JSON nesting too deep")
	}

	// Check for dangerous content in JSON values
	if civ.containsDangerousJSONContent(data) {
		result.Valid = false
		result.Errors = append(result.Errors, "JSON contains potentially dangerous content")
	}

	return result
}

// getJSONDepth calculates the maximum nesting depth of JSON data
func (civ *ComprehensiveInputValidator) getJSONDepth(data interface{}) int {
	switch v := data.(type) {
	case map[string]interface{}:
		maxDepth := 0
		for _, value := range v {
			depth := civ.getJSONDepth(value)
			if depth > maxDepth {
				maxDepth = depth
			}
		}
		return maxDepth + 1
	case []interface{}:
		maxDepth := 0
		for _, value := range v {
			depth := civ.getJSONDepth(value)
			if depth > maxDepth {
				maxDepth = depth
			}
		}
		return maxDepth + 1
	default:
		return 1
	}
}

// containsDangerousJSONContent checks for dangerous content in JSON values
func (civ *ComprehensiveInputValidator) containsDangerousJSONContent(data interface{}) bool {
	switch v := data.(type) {
	case string:
		return civ.containsDangerousPatterns(v)
	case map[string]interface{}:
		for key, value := range v {
			if civ.containsDangerousPatterns(key) || civ.containsDangerousJSONContent(value) {
				return true
			}
		}
	case []interface{}:
		if len(v) > civ.maxArrayLength {
			return true
		}
		for _, value := range v {
			if civ.containsDangerousJSONContent(value) {
				return true
			}
		}
	}
	return false
}

// ValidateURL validates and sanitizes URLs
func (civ *ComprehensiveInputValidator) ValidateURL(rawURL string) *ValidationResult {
	result := civ.ValidateInput(rawURL, "url")

	if result.Valid {
		// Additional URL validation
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, "Invalid URL format")
			return result
		}

		// Check scheme
		if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
			result.Valid = false
			result.Errors = append(result.Errors, "Only HTTP and HTTPS URLs are allowed")
		}

		// Check for suspicious patterns in URL
		if strings.Contains(parsedURL.String(), "..") {
			result.Valid = false
			result.Errors = append(result.Errors, "URL contains path traversal patterns")
		}
	}

	return result
}

// Global instance for easy access
var DefaultInputValidator = NewComprehensiveInputValidator()

// Convenience functions
func ValidateModelName(name string) *ValidationResult {
	return DefaultInputValidator.ValidateInput(name, "model_name")
}

func ValidateNodeID(id string) *ValidationResult {
	return DefaultInputValidator.ValidateInput(id, "node_id")
}

func ValidateUserInput(input string) *ValidationResult {
	return DefaultInputValidator.ValidateInput(input, "user_input")
}

func ValidateFilePath(path string) *ValidationResult {
	return DefaultInputValidator.ValidateInput(path, "file_path")
}

func ValidateJSON(jsonData []byte) *ValidationResult {
	return DefaultInputValidator.ValidateJSON(jsonData)
}

func ValidateURL(url string) *ValidationResult {
	return DefaultInputValidator.ValidateURL(url)
}

func ValidatePrompt(prompt string) error {
	result := DefaultInputValidator.ValidateInput(prompt, "user_input")
	if !result.Valid {
		return fmt.Errorf("invalid prompt: %v", result.Errors)
	}
	return nil
}

func ValidateTransferID(transferID string) error {
	result := DefaultInputValidator.ValidateInput(transferID, "alphanumeric")
	if !result.Valid {
		return fmt.Errorf("invalid transfer ID: %v", result.Errors)
	}
	return nil
}

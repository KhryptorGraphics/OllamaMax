package config

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s (value: %v)", e.Field, e.Message, e.Value)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}
	if len(e) == 1 {
		return e[0].Error()
	}

	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return fmt.Sprintf("multiple validation errors: %s", strings.Join(messages, "; "))
}

// ValidateExtended performs extended validation beyond basic checks
func (c *Config) ValidateExtended() error {
	var errors ValidationErrors

	// Validate node configuration
	if err := c.validateNode(); err != nil {
		if ve, ok := err.(ValidationErrors); ok {
			errors = append(errors, ve...)
		} else {
			errors = append(errors, ValidationError{Field: "node", Message: err.Error()})
		}
	}

	// Validate API configuration
	if err := c.validateAPI(); err != nil {
		if ve, ok := err.(ValidationErrors); ok {
			errors = append(errors, ve...)
		} else {
			errors = append(errors, ValidationError{Field: "api", Message: err.Error()})
		}
	}

	// Validate P2P configuration
	if err := c.validateP2P(); err != nil {
		if ve, ok := err.(ValidationErrors); ok {
			errors = append(errors, ve...)
		} else {
			errors = append(errors, ValidationError{Field: "p2p", Message: err.Error()})
		}
	}

	// Validate storage configuration
	if err := c.validateStorage(); err != nil {
		if ve, ok := err.(ValidationErrors); ok {
			errors = append(errors, ve...)
		} else {
			errors = append(errors, ValidationError{Field: "storage", Message: err.Error()})
		}
	}

	// Validate security configuration
	if err := c.validateSecurity(); err != nil {
		if ve, ok := err.(ValidationErrors); ok {
			errors = append(errors, ve...)
		} else {
			errors = append(errors, ValidationError{Field: "security", Message: err.Error()})
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// validateNode validates node configuration
func (c *Config) validateNode() error {
	var errors ValidationErrors

	// Validate node ID
	if c.Node.ID == "" {
		errors = append(errors, ValidationError{
			Field:   "node.id",
			Value:   c.Node.ID,
			Message: "node ID is required",
		})
	} else if !isValidNodeID(c.Node.ID) {
		errors = append(errors, ValidationError{
			Field:   "node.id",
			Value:   c.Node.ID,
			Message: "node ID must contain only alphanumeric characters, hyphens, and underscores",
		})
	}

	// Validate environment
	validEnvironments := []string{"development", "testing", "staging", "production"}
	if !contains(validEnvironments, c.Node.Environment) {
		errors = append(errors, ValidationError{
			Field:   "node.environment",
			Value:   c.Node.Environment,
			Message: fmt.Sprintf("environment must be one of: %s", strings.Join(validEnvironments, ", ")),
		})
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// validateAPI validates API configuration
func (c *Config) validateAPI() error {
	var errors ValidationErrors

	// Validate listen address
	if c.API.Listen == "" {
		errors = append(errors, ValidationError{
			Field:   "api.listen",
			Value:   c.API.Listen,
			Message: "listen address is required",
		})
	} else if !isValidListenAddress(c.API.Listen) {
		errors = append(errors, ValidationError{
			Field:   "api.listen",
			Value:   c.API.Listen,
			Message: "invalid listen address format",
		})
	}

	// Validate timeout
	if c.API.Timeout <= 0 {
		errors = append(errors, ValidationError{
			Field:   "api.timeout",
			Value:   c.API.Timeout,
			Message: "timeout must be positive",
		})
	}

	// Validate max body size
	if c.API.MaxBodySize <= 0 {
		errors = append(errors, ValidationError{
			Field:   "api.max_body_size",
			Value:   c.API.MaxBodySize,
			Message: "max body size must be positive",
		})
	}

	// Validate TLS configuration
	if c.API.TLS.Enabled {
		if c.API.TLS.CertFile == "" {
			errors = append(errors, ValidationError{
				Field:   "api.tls.cert_file",
				Value:   c.API.TLS.CertFile,
				Message: "cert file is required when TLS is enabled",
			})
		} else if !fileExists(c.API.TLS.CertFile) {
			errors = append(errors, ValidationError{
				Field:   "api.tls.cert_file",
				Value:   c.API.TLS.CertFile,
				Message: "cert file does not exist",
			})
		}

		if c.API.TLS.KeyFile == "" {
			errors = append(errors, ValidationError{
				Field:   "api.tls.key_file",
				Value:   c.API.TLS.KeyFile,
				Message: "key file is required when TLS is enabled",
			})
		} else if !fileExists(c.API.TLS.KeyFile) {
			errors = append(errors, ValidationError{
				Field:   "api.tls.key_file",
				Value:   c.API.TLS.KeyFile,
				Message: "key file does not exist",
			})
		}
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// validateP2P validates P2P configuration
func (c *Config) validateP2P() error {
	var errors ValidationErrors

	// Validate listen address
	if c.P2P.Listen == "" {
		errors = append(errors, ValidationError{
			Field:   "p2p.listen",
			Value:   c.P2P.Listen,
			Message: "listen address is required",
		})
	} else if !isValidListenAddress(c.P2P.Listen) {
		errors = append(errors, ValidationError{
			Field:   "p2p.listen",
			Value:   c.P2P.Listen,
			Message: "invalid listen address format",
		})
	}

	// Validate bootstrap peers
	for i, peer := range c.P2P.Bootstrap {
		if !isValidPeerAddress(peer) {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("p2p.bootstrap[%d]", i),
				Value:   peer,
				Message: "invalid peer address format",
			})
		}
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// validateStorage validates storage configuration
func (c *Config) validateStorage() error {
	var errors ValidationErrors

	// Validate directories
	dirs := map[string]string{
		"storage.data_dir":  c.Storage.DataDir,
		"storage.model_dir": c.Storage.ModelDir,
		"storage.cache_dir": c.Storage.CacheDir,
	}

	for field, dir := range dirs {
		if dir == "" {
			errors = append(errors, ValidationError{
				Field:   field,
				Value:   dir,
				Message: "directory path is required",
			})
		} else if !isValidPath(dir) {
			errors = append(errors, ValidationError{
				Field:   field,
				Value:   dir,
				Message: "invalid directory path",
			})
		}
	}

	// Validate max disk size
	if c.Storage.MaxDiskSize <= 0 {
		errors = append(errors, ValidationError{
			Field:   "storage.max_disk_size",
			Value:   c.Storage.MaxDiskSize,
			Message: "max disk size must be positive",
		})
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// validateSecurity validates security configuration
func (c *Config) validateSecurity() error {
	var errors ValidationErrors

	// Validate authentication
	if c.Security.Auth.Enabled {
		if c.Security.Auth.SecretKey == "" {
			errors = append(errors, ValidationError{
				Field:   "security.auth.secret_key",
				Value:   c.Security.Auth.SecretKey,
				Message: "secret key is required when authentication is enabled",
			})
		} else if len(c.Security.Auth.SecretKey) < 32 {
			errors = append(errors, ValidationError{
				Field:   "security.auth.secret_key",
				Value:   "[REDACTED]",
				Message: "secret key must be at least 32 characters long",
			})
		}

		validMethods := []string{"jwt", "api_key", "oauth"}
		if !contains(validMethods, c.Security.Auth.Method) {
			errors = append(errors, ValidationError{
				Field:   "security.auth.method",
				Value:   c.Security.Auth.Method,
				Message: fmt.Sprintf("auth method must be one of: %s", strings.Join(validMethods, ", ")),
			})
		}

		if c.Security.Auth.TokenExpiry <= 0 {
			errors = append(errors, ValidationError{
				Field:   "security.auth.token_expiry",
				Value:   c.Security.Auth.TokenExpiry,
				Message: "token expiry must be positive",
			})
		}
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// Helper functions

func isValidNodeID(id string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9\-_]+$`, id)
	return matched
}

func isValidListenAddress(addr string) bool {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}

	// Validate host
	if host != "" && net.ParseIP(host) == nil && host != "localhost" {
		return false
	}

	// Validate port
	if portNum, err := strconv.Atoi(port); err != nil || portNum < 0 || portNum > 65535 {
		return false
	}

	return true
}

func isValidPeerAddress(addr string) bool {
	// Check if it's a multiaddr format
	if strings.HasPrefix(addr, "/") {
		return true // Assume valid multiaddr for now
	}

	// Check if it's a regular address
	return isValidListenAddress(addr)
}

func isValidPath(path string) bool {
	// Check for invalid characters
	if strings.ContainsAny(path, "<>:\"|?*") {
		return false
	}

	// Check if it's an absolute path or relative path
	return filepath.IsAbs(path) || !strings.HasPrefix(path, "..")
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

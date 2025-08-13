package auth

import (
	"fmt"
	"net/http"
)

// AuthError represents a structured authentication error
type AuthError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	StatusCode int                    `json:"status_code"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

func (e AuthError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Predefined authentication errors
var (
	ErrInvalidCredentials = AuthError{
		Code:       "INVALID_CREDENTIALS",
		Message:    "Invalid username or password",
		StatusCode: http.StatusUnauthorized,
	}

	ErrTokenInvalid = AuthError{
		Code:       "TOKEN_INVALID",
		Message:    "Invalid or malformed token",
		StatusCode: http.StatusUnauthorized,
	}

	ErrTokenExpired = AuthError{
		Code:       "TOKEN_EXPIRED",
		Message:    "Token has expired",
		StatusCode: http.StatusUnauthorized,
	}

	ErrTokenBlacklisted = AuthError{
		Code:       "TOKEN_BLACKLISTED",
		Message:    "Token has been revoked",
		StatusCode: http.StatusUnauthorized,
	}

	ErrUserNotFound = AuthError{
		Code:       "USER_NOT_FOUND",
		Message:    "User account not found or inactive",
		StatusCode: http.StatusUnauthorized,
	}

	ErrUserInactive = AuthError{
		Code:       "USER_INACTIVE",
		Message:    "User account is disabled",
		StatusCode: http.StatusUnauthorized,
	}

	ErrInsufficientPermissions = AuthError{
		Code:       "INSUFFICIENT_PERMISSIONS",
		Message:    "Insufficient permissions to access this resource",
		StatusCode: http.StatusForbidden,
	}

	ErrAPIKeyNotFound = AuthError{
		Code:       "API_KEY_NOT_FOUND",
		Message:    "API key not found or invalid",
		StatusCode: http.StatusUnauthorized,
	}

	ErrAPIKeyExpired = AuthError{
		Code:       "API_KEY_EXPIRED",
		Message:    "API key has expired",
		StatusCode: http.StatusUnauthorized,
	}

	ErrSessionNotFound = AuthError{
		Code:       "SESSION_NOT_FOUND",
		Message:    "Session not found",
		StatusCode: http.StatusUnauthorized,
	}

	ErrSessionExpired = AuthError{
		Code:       "SESSION_EXPIRED",
		Message:    "Session has expired",
		StatusCode: http.StatusUnauthorized,
	}

	ErrRateLimitExceeded = AuthError{
		Code:       "RATE_LIMIT_EXCEEDED",
		Message:    "Rate limit exceeded, please try again later",
		StatusCode: http.StatusTooManyRequests,
	}

	ErrInvalidInput = AuthError{
		Code:       "INVALID_INPUT",
		Message:    "Invalid input data",
		StatusCode: http.StatusBadRequest,
	}

	ErrConfigurationError = AuthError{
		Code:       "CONFIGURATION_ERROR",
		Message:    "Authentication system configuration error",
		StatusCode: http.StatusInternalServerError,
	}
)

// NewAuthError creates a new authentication error with custom details
func NewAuthError(code, message string, statusCode int, details map[string]interface{}) AuthError {
	return AuthError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Details:    details,
	}
}

// WrapError wraps a generic error as an authentication error
func WrapError(err error, code string, statusCode int) AuthError {
	return AuthError{
		Code:       code,
		Message:    err.Error(),
		StatusCode: statusCode,
	}
}

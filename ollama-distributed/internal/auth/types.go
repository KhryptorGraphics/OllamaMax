package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// User represents a system user
type User struct {
	ID          string            `json:"id"`
	Username    string            `json:"username"`
	Email       string            `json:"email,omitempty"`
	Role        string            `json:"role"`
	Permissions []string          `json:"permissions"`
	APIKeys     []APIKey          `json:"api_keys,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	LastLoginAt *time.Time        `json:"last_login_at,omitempty"`
	Active      bool              `json:"active"`
}

// APIKey represents an API key for authentication
type APIKey struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Key         string            `json:"key"`
	UserID      string            `json:"user_id"`
	Permissions []string          `json:"permissions"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	LastUsedAt  *time.Time        `json:"last_used_at,omitempty"`
	Active      bool              `json:"active"`
}

// Session represents an authentication session
type Session struct {
	ID        string            `json:"id"`
	UserID    string            `json:"user_id"`
	TokenID   string            `json:"token_id"`
	IPAddress string            `json:"ip_address"`
	UserAgent string            `json:"user_agent"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	ExpiresAt time.Time         `json:"expires_at"`
	Active    bool              `json:"active"`
}

// Claims represents JWT claims for the system
type Claims struct {
	UserID      string            `json:"user_id"`
	Username    string            `json:"username"`
	Email       string            `json:"email,omitempty"`
	Role        string            `json:"role"`
	Permissions []string          `json:"permissions"`
	SessionID   string            `json:"session_id,omitempty"`
	APIKeyID    string            `json:"api_key_id,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	jwt.RegisteredClaims
}

// AuthContext contains authentication information for a request
type AuthContext struct {
	User        *User     `json:"user"`
	Session     *Session  `json:"session,omitempty"`
	APIKey      *APIKey   `json:"api_key,omitempty"`
	Claims      *Claims   `json:"claims"`
	TokenString string    `json:"-"`
	Method      AuthMethod `json:"method"`
}

// AuthMethod represents the authentication method used
type AuthMethod string

const (
	AuthMethodJWT    AuthMethod = "jwt"
	AuthMethodAPIKey AuthMethod = "api_key"
	AuthMethodX509   AuthMethod = "x509"
	AuthMethodNone   AuthMethod = "none"
)

// Permission constants
const (
	PermissionNodeRead       = "node:read"
	PermissionNodeWrite      = "node:write"
	PermissionNodeAdmin      = "node:admin"
	PermissionModelRead      = "model:read"
	PermissionModelWrite     = "model:write"
	PermissionModelAdmin     = "model:admin"
	PermissionClusterRead    = "cluster:read"
	PermissionClusterWrite   = "cluster:write"
	PermissionClusterAdmin   = "cluster:admin"
	PermissionInferenceRead  = "inference:read"
	PermissionInferenceWrite = "inference:write"
	PermissionMetricsRead    = "metrics:read"
	PermissionSystemAdmin    = "system:admin"
	PermissionUserAdmin      = "user:admin"
)

// Role constants
const (
	RoleAdmin     = "admin"
	RoleOperator  = "operator"
	RoleUser      = "user"
	RoleReadOnly  = "readonly"
	RoleService   = "service"
)

// Default role permissions
var DefaultRolePermissions = map[string][]string{
	RoleAdmin: {
		PermissionSystemAdmin,
		PermissionUserAdmin,
		PermissionNodeAdmin,
		PermissionModelAdmin,
		PermissionClusterAdmin,
		PermissionInferenceWrite,
		PermissionMetricsRead,
	},
	RoleOperator: {
		PermissionNodeWrite,
		PermissionModelWrite,
		PermissionClusterWrite,
		PermissionInferenceWrite,
		PermissionMetricsRead,
	},
	RoleUser: {
		PermissionNodeRead,
		PermissionModelRead,
		PermissionClusterRead,
		PermissionInferenceWrite,
		PermissionMetricsRead,
	},
	RoleReadOnly: {
		PermissionNodeRead,
		PermissionModelRead,
		PermissionClusterRead,
		PermissionInferenceRead,
		PermissionMetricsRead,
	},
	RoleService: {
		PermissionNodeRead,
		PermissionModelRead,
		PermissionInferenceWrite,
	},
}

// AuthError represents authentication errors
type AuthError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e AuthError) Error() string {
	if e.Details != "" {
		return e.Message + ": " + e.Details
	}
	return e.Message
}

// Common authentication errors
var (
	ErrInvalidCredentials = AuthError{
		Code:    "INVALID_CREDENTIALS",
		Message: "Invalid credentials provided",
	}
	ErrTokenExpired = AuthError{
		Code:    "TOKEN_EXPIRED",
		Message: "Authentication token has expired",
	}
	ErrTokenInvalid = AuthError{
		Code:    "TOKEN_INVALID",
		Message: "Authentication token is invalid",
	}
	ErrTokenBlacklisted = AuthError{
		Code:    "TOKEN_BLACKLISTED",
		Message: "Authentication token has been revoked",
	}
	ErrInsufficientPermissions = AuthError{
		Code:    "INSUFFICIENT_PERMISSIONS",
		Message: "Insufficient permissions for this operation",
	}
	ErrUserNotFound = AuthError{
		Code:    "USER_NOT_FOUND",
		Message: "User not found",
	}
	ErrUserInactive = AuthError{
		Code:    "USER_INACTIVE",
		Message: "User account is inactive",
	}
	ErrAPIKeyNotFound = AuthError{
		Code:    "API_KEY_NOT_FOUND",
		Message: "API key not found",
	}
	ErrAPIKeyInactive = AuthError{
		Code:    "API_KEY_INACTIVE",
		Message: "API key is inactive",
	}
	ErrAPIKeyExpired = AuthError{
		Code:    "API_KEY_EXPIRED",
		Message: "API key has expired",
	}
	ErrSessionNotFound = AuthError{
		Code:    "SESSION_NOT_FOUND",
		Message: "Session not found",
	}
	ErrSessionExpired = AuthError{
		Code:    "SESSION_EXPIRED",
		Message: "Session has expired",
	}
)

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      *User     `json:"user"`
	SessionID string    `json:"session_id"`
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	Token string `json:"token" binding:"required"`
}

// CreateAPIKeyRequest represents an API key creation request
type CreateAPIKeyRequest struct {
	Name        string            `json:"name" binding:"required"`
	Permissions []string          `json:"permissions,omitempty"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// CreateAPIKeyResponse represents an API key creation response
type CreateAPIKeyResponse struct {
	APIKey *APIKey `json:"api_key"`
	Key    string  `json:"key"` // Only returned once during creation
}

// CreateUserRequest represents a user creation request
type CreateUserRequest struct {
	Username    string            `json:"username" binding:"required"`
	Email       string            `json:"email,omitempty"`
	Password    string            `json:"password" binding:"required"`
	Role        string            `json:"role" binding:"required"`
	Permissions []string          `json:"permissions,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// UpdateUserRequest represents a user update request
type UpdateUserRequest struct {
	Email       *string           `json:"email,omitempty"`
	Role        *string           `json:"role,omitempty"`
	Permissions []string          `json:"permissions,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Active      *bool             `json:"active,omitempty"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}
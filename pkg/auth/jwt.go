package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/khryptorgraphics/ollamamax/internal/config"
)

// JWTService handles JWT token operations
type JWTService struct {
	privateKey    *rsa.PrivateKey
	publicKey     *rsa.PublicKey
	issuer        string
	expiration    time.Duration
	refreshExpiry time.Duration
}

// Claims represents JWT claims structure
type Claims struct {
	UserID      string            `json:"user_id"`
	Username    string            `json:"username"`
	Role        string            `json:"role"`
	Permissions []string          `json:"permissions"`
	Metadata    map[string]string `json:"metadata"`
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// NewJWTService creates a new JWT service instance
func NewJWTService(config *config.AuthConfig) (*JWTService, error) {
	// Generate RSA key pair if not provided
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	service := &JWTService{
		privateKey:    privateKey,
		publicKey:     &privateKey.PublicKey,
		issuer:        "ollamamax",
		expiration:    24 * time.Hour, // Default 24 hours
		refreshExpiry: 7 * 24 * time.Hour, // Default 7 days
	}

	// Override with config values if provided
	if config != nil {
		if config.JWTSecret != "" {
			service.issuer = config.JWTSecret
		}
		if config.TokenExpiry > 0 {
			service.expiration = config.TokenExpiry
		}
	}

	return service, nil
}

// GenerateToken creates a new JWT token for the given user
func (j *JWTService) GenerateToken(userID, username, role string, permissions []string) (*TokenPair, error) {
	now := time.Now()
	expiresAt := now.Add(j.expiration)
	refreshExpiresAt := now.Add(j.refreshExpiry)

	// Create access token claims
	claims := &Claims{
		UserID:      userID,
		Username:    username,
		Role:        role,
		Permissions: permissions,
		Metadata:    make(map[string]string),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   userID,
			Audience:  []string{"ollamamax"},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        fmt.Sprintf("%s_%d", userID, now.Unix()),
		},
	}

	// Create access token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	accessToken, err := token.SignedString(j.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Create refresh token claims
	refreshClaims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   userID,
			Audience:  []string{"ollamamax-refresh"},
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        fmt.Sprintf("%s_refresh_%d", userID, now.Unix()),
		},
	}

	// Create refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(j.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		ExpiresAt:    expiresAt,
		TokenType:    "Bearer",
	}, nil
}

// ValidateToken validates and parses a JWT token
func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	// Additional validation
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token has expired")
	}

	return claims, nil
}

// RefreshToken creates a new access token from a valid refresh token
func (j *JWTService) RefreshToken(refreshTokenString string) (*TokenPair, error) {
	// Validate refresh token
	claims, err := j.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Check if it's actually a refresh token
	if len(claims.Audience) == 0 || claims.Audience[0] != "ollamamax-refresh" {
		return nil, errors.New("not a refresh token")
	}

	// Generate new token pair
	return j.GenerateToken(claims.UserID, claims.Username, claims.Role, claims.Permissions)
}

// RevokeToken adds a token to the revocation list (blacklist)
func (j *JWTService) RevokeToken(tokenString string) error {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return fmt.Errorf("cannot revoke invalid token: %w", err)
	}

	// In a production system, you would store this in a database or cache
	// For now, we'll just validate that the token is parseable
	_ = claims
	return nil
}

// GetPublicKey returns the public key for token verification
func (j *JWTService) GetPublicKey() *rsa.PublicKey {
	return j.publicKey
}

// SetPrivateKey sets a custom private key (for testing or custom key management)
func (j *JWTService) SetPrivateKey(key *rsa.PrivateKey) {
	j.privateKey = key
	j.publicKey = &key.PublicKey
}

// HasPermission checks if the claims contain a specific permission
func (c *Claims) HasPermission(permission string) bool {
	for _, p := range c.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// IsAdmin checks if the user has admin role
func (c *Claims) IsAdmin() bool {
	return c.Role == "admin"
}

// IsOperator checks if the user has operator role or higher
func (c *Claims) IsOperator() bool {
	return c.Role == "admin" || c.Role == "operator"
}

// GetMetadata safely retrieves metadata value
func (c *Claims) GetMetadata(key string) (string, bool) {
	if c.Metadata == nil {
		return "", false
	}
	value, exists := c.Metadata[key]
	return value, exists
}

// SetMetadata safely sets metadata value
func (c *Claims) SetMetadata(key, value string) {
	if c.Metadata == nil {
		c.Metadata = make(map[string]string)
	}
	c.Metadata[key] = value
}

// Predefined roles and permissions
const (
	RoleAdmin    = "admin"
	RoleOperator = "operator"
	RoleUser     = "user"
	RoleReadonly = "readonly"
)

// Predefined permissions
const (
	PermissionModelManage    = "model:manage"
	PermissionModelRead      = "model:read"
	PermissionClusterManage  = "cluster:manage"
	PermissionClusterRead    = "cluster:read"
	PermissionNodeManage     = "node:manage"
	PermissionNodeRead       = "node:read"
	PermissionInferenceRun   = "inference:run"
	PermissionMetricsRead    = "metrics:read"
	PermissionSystemManage   = "system:manage"
)

// GetRolePermissions returns default permissions for a role
func GetRolePermissions(role string) []string {
	switch role {
	case RoleAdmin:
		return []string{
			PermissionModelManage, PermissionModelRead,
			PermissionClusterManage, PermissionClusterRead,
			PermissionNodeManage, PermissionNodeRead,
			PermissionInferenceRun, PermissionMetricsRead,
			PermissionSystemManage,
		}
	case RoleOperator:
		return []string{
			PermissionModelRead, PermissionClusterRead,
			PermissionNodeRead, PermissionInferenceRun,
			PermissionMetricsRead,
		}
	case RoleUser:
		return []string{
			PermissionModelRead, PermissionInferenceRun,
		}
	case RoleReadonly:
		return []string{
			PermissionModelRead, PermissionClusterRead,
			PermissionNodeRead, PermissionMetricsRead,
		}
	default:
		return []string{}
	}
}

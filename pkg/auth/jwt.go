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
		if config.SecretKey != "" {
			service.issuer = config.SecretKey
		}
		if config.TokenExpiry > 0 {
			service.expiration = config.TokenExpiry
		}
	}

	return service, nil
}

// GenerateTokens creates access and refresh tokens (API compatibility method)
func (j *JWTService) GenerateTokens(userID, username string, roles []string) (accessToken, refreshToken string, err error) {
	role := "user" // default role
	if len(roles) > 0 {
		role = roles[0] // use first role
	}
	
	tokens, err := j.GenerateToken(userID, username, role, nil)
	if err != nil {
		return "", "", err
	}
	
	return tokens.AccessToken, tokens.RefreshToken, nil
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

	return claims, nil
}

// ValidateRefreshToken validates a refresh token and returns user info
func (j *JWTService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Additional validation for refresh tokens
	if len(claims.Audience) == 0 || claims.Audience[0] != "ollamamax-refresh" {
		return nil, errors.New("token is not a valid refresh token")
	}

	return claims, nil
}

// RefreshTokens creates new tokens from a valid refresh token
func (j *JWTService) RefreshTokens(refreshToken string) (string, string, error) {
	claims, err := j.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	// Generate new token pair
	return j.GenerateTokens(claims.UserID, claims.Username, []string{claims.Role})
}

// GetUserFromToken extracts user information from a valid token
func (j *JWTService) GetUserFromToken(tokenString string) (*Claims, error) {
	return j.ValidateToken(tokenString)
}
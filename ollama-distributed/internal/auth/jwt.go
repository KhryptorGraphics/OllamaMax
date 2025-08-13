package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
)

// JWTManager handles JWT token operations with advanced features
type JWTManager struct {
	config     *config.AuthConfig
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey

	// Token blacklist and refresh tokens
	blacklist     map[string]time.Time
	refreshTokens map[string]*RefreshToken
}

// RefreshToken represents a refresh token
type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	TokenHash string    `json:"token_hash"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	Used      bool      `json:"used"`
}

// TokenPair represents an access token and refresh token pair
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(cfg *config.AuthConfig) (*JWTManager, error) {
	// Generate RSA key pair for signing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	return &JWTManager{
		config:        cfg,
		privateKey:    privateKey,
		publicKey:     &privateKey.PublicKey,
		blacklist:     make(map[string]time.Time),
		refreshTokens: make(map[string]*RefreshToken),
	}, nil
}

// GenerateTokenPair generates an access token and refresh token pair
func (jm *JWTManager) GenerateTokenPair(user *User, sessionID string, metadata map[string]string) (*TokenPair, error) {
	now := time.Now()
	accessTokenExpiry := now.Add(jm.config.TokenExpiry)
	refreshTokenExpiry := now.Add(7 * 24 * time.Hour) // 7 days for refresh token

	// Create access token claims
	accessClaims := &Claims{
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		Permissions: user.Permissions,
		SessionID:   sessionID,
		Metadata:    metadata,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessTokenExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    jm.config.Issuer,
			Subject:   user.ID,
			ID:        generateID(),
			Audience:  []string{jm.config.Audience},
		},
	}

	// Sign access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(jm.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Create refresh token
	refreshTokenID := generateID()
	refreshTokenString := generateAPIKey() // Reuse the secure random generation
	refreshTokenHash := hashAPIKey(refreshTokenString)

	refreshToken := &RefreshToken{
		ID:        refreshTokenID,
		UserID:    user.ID,
		TokenHash: refreshTokenHash,
		ExpiresAt: refreshTokenExpiry,
		CreatedAt: now,
		Used:      false,
	}

	jm.refreshTokens[refreshTokenID] = refreshToken

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    accessTokenExpiry,
		TokenType:    "Bearer",
	}, nil
}

// RefreshAccessToken generates a new access token using a refresh token
func (jm *JWTManager) RefreshAccessToken(refreshTokenString string, user *User) (*TokenPair, error) {
	refreshTokenHash := hashAPIKey(refreshTokenString)

	// Find the refresh token
	var refreshToken *RefreshToken
	for _, rt := range jm.refreshTokens {
		if rt.TokenHash == refreshTokenHash && rt.UserID == user.ID && !rt.Used {
			if time.Now().Before(rt.ExpiresAt) {
				refreshToken = rt
				break
			}
		}
	}

	if refreshToken == nil {
		return nil, fmt.Errorf("invalid or expired refresh token")
	}

	// Mark the old refresh token as used
	refreshToken.Used = true

	// Generate new token pair
	return jm.GenerateTokenPair(user, "", nil)
}

// ValidateToken validates a JWT access token
func (jm *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jm.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check if token is blacklisted
	if jm.isBlacklisted(claims.ID) {
		return nil, fmt.Errorf("token is blacklisted")
	}

	return claims, nil
}

// BlacklistToken adds a token to the blacklist
func (jm *JWTManager) BlacklistToken(tokenID string, expiry time.Time) {
	jm.blacklist[tokenID] = expiry
}

// RevokeRefreshToken revokes a refresh token
func (jm *JWTManager) RevokeRefreshToken(refreshTokenString, userID string) error {
	refreshTokenHash := hashAPIKey(refreshTokenString)

	for _, rt := range jm.refreshTokens {
		if rt.TokenHash == refreshTokenHash && rt.UserID == userID {
			rt.Used = true
			return nil
		}
	}

	return fmt.Errorf("refresh token not found")
}

// RevokeAllUserTokens revokes all tokens for a user
func (jm *JWTManager) RevokeAllUserTokens(userID string) {
	// Mark all refresh tokens as used
	for _, rt := range jm.refreshTokens {
		if rt.UserID == userID {
			rt.Used = true
		}
	}
}

// CleanupExpiredTokens removes expired tokens from memory
func (jm *JWTManager) CleanupExpiredTokens() {
	now := time.Now()

	// Clean up blacklist
	for tokenID, expiry := range jm.blacklist {
		if now.After(expiry) {
			delete(jm.blacklist, tokenID)
		}
	}

	// Clean up refresh tokens
	for id, rt := range jm.refreshTokens {
		if now.After(rt.ExpiresAt) || rt.Used {
			delete(jm.refreshTokens, id)
		}
	}
}

// GetTokenClaims extracts claims from a token without validating it (useful for expired tokens)
func (jm *JWTManager) GetTokenClaims(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jm.publicKey, nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// GenerateServiceToken generates a long-lived token for service-to-service communication
func (jm *JWTManager) GenerateServiceToken(serviceID, serviceName string, permissions []string) (string, error) {
	now := time.Now()
	expiry := now.Add(365 * 24 * time.Hour) // 1 year

	claims := &Claims{
		UserID:      serviceID,
		Username:    serviceName,
		Role:        RoleService,
		Permissions: permissions,
		Metadata: map[string]string{
			"token_type": "service",
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    jm.config.Issuer,
			Subject:   serviceID,
			ID:        generateID(),
			Audience:  []string{jm.config.Audience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(jm.privateKey)
}

// ValidateServiceToken validates a service token
func (jm *JWTManager) ValidateServiceToken(tokenString string) (*Claims, error) {
	claims, err := jm.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Verify this is a service token
	if claims.Metadata["token_type"] != "service" {
		return nil, fmt.Errorf("not a service token")
	}

	return claims, nil
}

// CreateShortLivedToken creates a token with custom expiry (for specific operations)
func (jm *JWTManager) CreateShortLivedToken(user *User, duration time.Duration, purpose string) (string, error) {
	now := time.Now()
	expiry := now.Add(duration)

	claims := &Claims{
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		Permissions: user.Permissions,
		Metadata: map[string]string{
			"token_type": "short_lived",
			"purpose":    purpose,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    jm.config.Issuer,
			Subject:   user.ID,
			ID:        generateID(),
			Audience:  []string{jm.config.Audience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(jm.privateKey)
}

// GetPublicKey returns the public key for token verification by other services
func (jm *JWTManager) GetPublicKey() *rsa.PublicKey {
	return jm.publicKey
}

// GetTokenStats returns statistics about tokens
func (jm *JWTManager) GetTokenStats() map[string]interface{} {
	activeRefreshTokens := 0
	expiredRefreshTokens := 0
	now := time.Now()

	for _, rt := range jm.refreshTokens {
		if rt.Used || now.After(rt.ExpiresAt) {
			expiredRefreshTokens++
		} else {
			activeRefreshTokens++
		}
	}

	return map[string]interface{}{
		"active_refresh_tokens":  activeRefreshTokens,
		"expired_refresh_tokens": expiredRefreshTokens,
		"blacklisted_tokens":     len(jm.blacklist),
		"total_refresh_tokens":   len(jm.refreshTokens),
	}
}

// isBlacklisted checks if a token ID is blacklisted
func (jm *JWTManager) isBlacklisted(tokenID string) bool {
	expiry, exists := jm.blacklist[tokenID]
	if !exists {
		return false
	}

	if time.Now().After(expiry) {
		delete(jm.blacklist, tokenID)
		return false
	}

	return true
}

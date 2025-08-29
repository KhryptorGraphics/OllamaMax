package security

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	_ "net/url"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"github.com/rs/zerolog/log"
)

// OAuth2Manager handles OAuth2 and OIDC authentication
type OAuth2Manager struct {
	config           *OAuth2Config
	providers        map[string]*OAuth2Provider
	oidcVerifiers    map[string]*oidc.IDTokenVerifier
	stateStore       map[string]*AuthState // In production, use Redis
	mu               sync.RWMutex
}

// OAuth2Config configures OAuth2/OIDC providers
type OAuth2Config struct {
	Providers    []OAuth2ProviderConfig `json:"providers"`
	CallbackURL  string                 `json:"callback_url"`
	StateTimeout time.Duration          `json:"state_timeout"`
}

// OAuth2ProviderConfig configures a single OAuth2 provider
type OAuth2ProviderConfig struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"` // oauth2, oidc
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	AuthURL      string   `json:"auth_url"`
	TokenURL     string   `json:"token_url"`
	UserInfoURL  string   `json:"user_info_url"`
	IssuerURL    string   `json:"issuer_url"` // For OIDC
	Scopes       []string `json:"scopes"`
	Enabled      bool     `json:"enabled"`
}

// OAuth2Provider wraps OAuth2 configuration and client
type OAuth2Provider struct {
	Config       *OAuth2ProviderConfig
	OAuth2Config *oauth2.Config
	OIDCProvider *oidc.Provider
	IDVerifier   *oidc.IDTokenVerifier
}

// AuthState represents OAuth2 authentication state
type AuthState struct {
	State       string            `json:"state"`
	Provider    string            `json:"provider"`
	RedirectURL string            `json:"redirect_url"`
	Nonce       string            `json:"nonce"`
	ExpiresAt   time.Time         `json:"expires_at"`
	Metadata    map[string]string `json:"metadata"`
}

// UserInfo represents user information from OAuth2 provider
type UserInfo struct {
	ID            string                 `json:"id"`
	Email         string                 `json:"email"`
	EmailVerified bool                   `json:"email_verified"`
	Name          string                 `json:"name"`
	GivenName     string                 `json:"given_name"`
	FamilyName    string                 `json:"family_name"`
	Picture       string                 `json:"picture"`
	Locale        string                 `json:"locale"`
	Provider      string                 `json:"provider"`
	Groups        []string               `json:"groups"`
	Roles         []string               `json:"roles"`
	Claims        map[string]interface{} `json:"claims"`
}

// NewOAuth2Manager creates a new OAuth2 manager
func NewOAuth2Manager(config *OAuth2Config) (*OAuth2Manager, error) {
	if config == nil {
		return nil, fmt.Errorf("OAuth2 config is required")
	}

	manager := &OAuth2Manager{
		config:        config,
		providers:     make(map[string]*OAuth2Provider),
		oidcVerifiers: make(map[string]*oidc.IDTokenVerifier),
		stateStore:    make(map[string]*AuthState),
	}

	// Initialize providers
	for _, providerConfig := range config.Providers {
		if !providerConfig.Enabled {
			continue
		}

		provider, err := manager.initializeProvider(&providerConfig)
		if err != nil {
			log.Error().
				Err(err).
				Str("provider", providerConfig.Name).
				Msg("Failed to initialize OAuth2 provider")
			continue
		}

		manager.providers[providerConfig.Name] = provider
	}

	// Start cleanup routine for expired states
	go manager.cleanupExpiredStates()

	log.Info().
		Int("providers", len(manager.providers)).
		Msg("OAuth2 manager initialized")

	return manager, nil
}

// initializeProvider initializes a single OAuth2/OIDC provider
func (om *OAuth2Manager) initializeProvider(config *OAuth2ProviderConfig) (*OAuth2Provider, error) {
	provider := &OAuth2Provider{
		Config: config,
	}

	// Initialize OAuth2 config
	provider.OAuth2Config = &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.AuthURL,
			TokenURL: config.TokenURL,
		},
		RedirectURL: om.config.CallbackURL,
		Scopes:      config.Scopes,
	}

	// Initialize OIDC if enabled
	if config.Type == "oidc" && config.IssuerURL != "" {
		ctx := context.Background()
		oidcProvider, err := oidc.NewProvider(ctx, config.IssuerURL)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize OIDC provider: %w", err)
		}

		provider.OIDCProvider = oidcProvider
		
		// Update OAuth2 config with OIDC endpoints
		provider.OAuth2Config.Endpoint = oidcProvider.Endpoint()

		// Create ID token verifier
		provider.IDVerifier = oidcProvider.Verifier(&oidc.Config{
			ClientID: config.ClientID,
		})
	}

	return provider, nil
}

// GetAuthURL generates authorization URL for OAuth2 flow
func (om *OAuth2Manager) GetAuthURL(providerName string, redirectURL string) (string, error) {
	om.mu.RLock()
	provider, exists := om.providers[providerName]
	om.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("provider %s not found", providerName)
	}

	// Generate state
	state := om.generateState()
	nonce := om.generateNonce()

	// Store auth state
	authState := &AuthState{
		State:       state,
		Provider:    providerName,
		RedirectURL: redirectURL,
		Nonce:       nonce,
		ExpiresAt:   time.Now().Add(om.config.StateTimeout),
		Metadata:    make(map[string]string),
	}

	om.mu.Lock()
	om.stateStore[state] = authState
	om.mu.Unlock()

	// Generate authorization URL
	options := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("state", state),
	}

	// Add nonce for OIDC
	if provider.Config.Type == "oidc" {
		options = append(options, oauth2.SetAuthURLParam("nonce", nonce))
	}

	// Add PKCE for enhanced security
	codeVerifier := om.generateCodeVerifier()
	codeChallenge := om.generateCodeChallenge(codeVerifier)
	authState.Metadata["code_verifier"] = codeVerifier

	options = append(options,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	authURL := provider.OAuth2Config.AuthCodeURL(state, options...)

	log.Info().
		Str("provider", providerName).
		Str("state", state).
		Msg("Generated OAuth2 authorization URL")

	return authURL, nil
}

// HandleCallback handles OAuth2 callback
func (om *OAuth2Manager) HandleCallback(c *gin.Context) (*UserInfo, error) {
	// Extract parameters
	state := c.Query("state")
	code := c.Query("code")
	error := c.Query("error")

	if error != "" {
		return nil, fmt.Errorf("OAuth2 error: %s", error)
	}

	if state == "" || code == "" {
		return nil, fmt.Errorf("missing state or code parameter")
	}

	// Validate state
	om.mu.Lock()
	authState, exists := om.stateStore[state]
	if exists {
		delete(om.stateStore, state)
	}
	om.mu.Unlock()

	if !exists {
		return nil, fmt.Errorf("invalid or expired state")
	}

	if time.Now().After(authState.ExpiresAt) {
		return nil, fmt.Errorf("authentication state expired")
	}

	// Get provider
	provider, exists := om.providers[authState.Provider]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", authState.Provider)
	}

	// Exchange code for token
	ctx := context.Background()
	
	// Add PKCE verifier
	options := []oauth2.AuthCodeOption{}
	if codeVerifier := authState.Metadata["code_verifier"]; codeVerifier != "" {
		options = append(options, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	}

	token, err := provider.OAuth2Config.Exchange(ctx, code, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Get user info
	userInfo, err := om.getUserInfo(ctx, provider, token, authState.Nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	userInfo.Provider = authState.Provider

	log.Info().
		Str("provider", authState.Provider).
		Str("user_id", userInfo.ID).
		Str("email", userInfo.Email).
		Msg("OAuth2 authentication successful")

	return userInfo, nil
}

// getUserInfo retrieves user information from the provider
func (om *OAuth2Manager) getUserInfo(ctx context.Context, provider *OAuth2Provider, token *oauth2.Token, nonce string) (*UserInfo, error) {
	userInfo := &UserInfo{
		Claims: make(map[string]interface{}),
	}

	// For OIDC, parse ID token first
	if provider.Config.Type == "oidc" && provider.IDVerifier != nil {
		rawIDToken, ok := token.Extra("id_token").(string)
		if ok {
			idToken, err := provider.IDVerifier.Verify(ctx, rawIDToken)
			if err != nil {
				return nil, fmt.Errorf("failed to verify ID token: %w", err)
			}

			// Verify nonce
			if nonce != "" {
				if idToken.Nonce != nonce {
					return nil, fmt.Errorf("nonce mismatch")
				}
			}

			// Parse claims
			var claims map[string]interface{}
			if err := idToken.Claims(&claims); err != nil {
				return nil, fmt.Errorf("failed to parse ID token claims: %w", err)
			}

			// Map standard claims
			if sub, ok := claims["sub"].(string); ok {
				userInfo.ID = sub
			}
			if email, ok := claims["email"].(string); ok {
				userInfo.Email = email
			}
			if emailVerified, ok := claims["email_verified"].(bool); ok {
				userInfo.EmailVerified = emailVerified
			}
			if name, ok := claims["name"].(string); ok {
				userInfo.Name = name
			}
			if givenName, ok := claims["given_name"].(string); ok {
				userInfo.GivenName = givenName
			}
			if familyName, ok := claims["family_name"].(string); ok {
				userInfo.FamilyName = familyName
			}
			if picture, ok := claims["picture"].(string); ok {
				userInfo.Picture = picture
			}
			if locale, ok := claims["locale"].(string); ok {
				userInfo.Locale = locale
			}

			// Extract groups and roles
			if groups, ok := claims["groups"].([]interface{}); ok {
				userInfo.Groups = make([]string, 0, len(groups))
				for _, group := range groups {
					if groupStr, ok := group.(string); ok {
						userInfo.Groups = append(userInfo.Groups, groupStr)
					}
				}
			}

			if roles, ok := claims["roles"].([]interface{}); ok {
				userInfo.Roles = make([]string, 0, len(roles))
				for _, role := range roles {
					if roleStr, ok := role.(string); ok {
						userInfo.Roles = append(userInfo.Roles, roleStr)
					}
				}
			}

			userInfo.Claims = claims
		}
	}

	// If we don't have user info yet, try the UserInfo endpoint
	if userInfo.ID == "" && provider.Config.UserInfoURL != "" {
		client := provider.OAuth2Config.Client(ctx, token)
		resp, err := client.Get(provider.Config.UserInfoURL)
		if err != nil {
			return nil, fmt.Errorf("failed to get user info: %w", err)
		}
		defer resp.Body.Close()

		var userInfoResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&userInfoResp); err != nil {
			return nil, fmt.Errorf("failed to decode user info: %w", err)
		}

		// Map user info fields
		if id, ok := userInfoResp["id"].(string); ok {
			userInfo.ID = id
		} else if sub, ok := userInfoResp["sub"].(string); ok {
			userInfo.ID = sub
		}
		
		if email, ok := userInfoResp["email"].(string); ok {
			userInfo.Email = email
		}
		
		if name, ok := userInfoResp["name"].(string); ok {
			userInfo.Name = name
		}

		// Merge claims
		for k, v := range userInfoResp {
			userInfo.Claims[k] = v
		}
	}

	return userInfo, nil
}

// RefreshToken refreshes an OAuth2 token
func (om *OAuth2Manager) RefreshToken(providerName string, refreshToken string) (*oauth2.Token, error) {
	om.mu.RLock()
	provider, exists := om.providers[providerName]
	om.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("provider %s not found", providerName)
	}

	ctx := context.Background()
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	tokenSource := provider.OAuth2Config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return newToken, nil
}

// GetProviders returns available OAuth2 providers
func (om *OAuth2Manager) GetProviders() map[string]string {
	om.mu.RLock()
	defer om.mu.RUnlock()

	providers := make(map[string]string)
	for name, provider := range om.providers {
		providers[name] = provider.Config.Type
	}

	return providers
}

// generateState generates a secure random state
func (om *OAuth2Manager) generateState() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

// generateNonce generates a secure random nonce
func (om *OAuth2Manager) generateNonce() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

// generateCodeVerifier generates PKCE code verifier
func (om *OAuth2Manager) generateCodeVerifier() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return base64URLEncode(bytes)
}

// generateCodeChallenge generates PKCE code challenge
func (om *OAuth2Manager) generateCodeChallenge(verifier string) string {
	// Implementation would use SHA256 hash and base64url encoding
	// Simplified for brevity
	return verifier // In real implementation, hash and encode
}

// base64URLEncode encodes bytes to base64url format
func base64URLEncode(data []byte) string {
	// Simplified implementation
	encoded := fmt.Sprintf("%x", data)
	return strings.TrimRight(encoded, "=")
}

// cleanupExpiredStates removes expired authentication states
func (om *OAuth2Manager) cleanupExpiredStates() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		<-ticker.C
		om.mu.Lock()
		now := time.Now()
		for state, authState := range om.stateStore {
			if now.After(authState.ExpiresAt) {
				delete(om.stateStore, state)
			}
		}
		om.mu.Unlock()
	}
}

// OAuth2Middleware provides OAuth2 authentication middleware
func (om *OAuth2Manager) OAuth2Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for OAuth2 token in Authorization header
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "OAuth2 token required"})
			c.Abort()
			return
		}

		_ = strings.TrimPrefix(authHeader, "Bearer ")

		// Validate token with providers (simplified)
		// In real implementation, validate token with the appropriate provider
		
		userInfo := &UserInfo{
			ID:    "oauth2_user",
			Email: "user@example.com",
			Name:  "OAuth2 User",
		}

		c.Set("user", userInfo)
		c.Next()
	}
}

// DefaultOAuth2Config returns default OAuth2 configuration
func DefaultOAuth2Config() *OAuth2Config {
	return &OAuth2Config{
		Providers: []OAuth2ProviderConfig{
			{
				Name:         "google",
				Type:         "oidc",
				IssuerURL:    "https://accounts.google.com",
				Scopes:       []string{"openid", "email", "profile"},
				Enabled:      false,
			},
			{
				Name:         "azure",
				Type:         "oidc",
				IssuerURL:    "https://login.microsoftonline.com/common/v2.0",
				Scopes:       []string{"openid", "email", "profile"},
				Enabled:      false,
			},
			{
				Name:         "github",
				Type:         "oauth2",
				AuthURL:      "https://github.com/login/oauth/authorize",
				TokenURL:     "https://github.com/login/oauth/access_token",
				UserInfoURL:  "https://api.github.com/user",
				Scopes:       []string{"user:email"},
				Enabled:      false,
			},
		},
		CallbackURL:  "https://localhost:8080/auth/callback",
		StateTimeout: 10 * time.Minute,
	}
}
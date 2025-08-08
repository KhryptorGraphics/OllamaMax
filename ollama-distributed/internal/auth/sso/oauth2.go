package sso

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/auth"
	"github.com/rs/zerolog/log"
)

// OAuth2Provider represents an OAuth2/OIDC identity provider
type OAuth2Provider struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Type         string            `json:"type"` // google, microsoft, okta, auth0, generic
	ClientID     string            `json:"client_id"`
	ClientSecret string            `json:"client_secret"`
	RedirectURL  string            `json:"redirect_url"`
	Scopes       []string          `json:"scopes"`
	AuthURL      string            `json:"auth_url"`
	TokenURL     string            `json:"token_url"`
	UserInfoURL  string            `json:"user_info_url"`
	JWKSURL      string            `json:"jwks_url"`
	Issuer       string            `json:"issuer"`
	Enabled      bool              `json:"enabled"`
	Config       map[string]string `json:"config"`
}

// OAuth2Manager manages OAuth2/OIDC authentication
type OAuth2Manager struct {
	providers map[string]*OAuth2Provider
	states    map[string]*OAuth2State
	authMgr   *auth.Manager
}

// OAuth2State represents OAuth2 state for CSRF protection
type OAuth2State struct {
	State       string    `json:"state"`
	Provider    string    `json:"provider"`
	RedirectURL string    `json:"redirect_url"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// OAuth2UserInfo represents user information from OAuth2 provider
type OAuth2UserInfo struct {
	ID           string            `json:"id"`
	Email        string            `json:"email"`
	Name         string            `json:"name"`
	FirstName    string            `json:"first_name"`
	LastName     string            `json:"last_name"`
	Picture      string            `json:"picture"`
	Verified     bool              `json:"verified"`
	Locale       string            `json:"locale"`
	Groups       []string          `json:"groups"`
	Roles        []string          `json:"roles"`
	Attributes   map[string]string `json:"attributes"`
	ProviderID   string            `json:"provider_id"`
	ProviderType string            `json:"provider_type"`
}

// NewOAuth2Manager creates a new OAuth2 manager
func NewOAuth2Manager(authMgr *auth.Manager) *OAuth2Manager {
	return &OAuth2Manager{
		providers: make(map[string]*OAuth2Provider),
		states:    make(map[string]*OAuth2State),
		authMgr:   authMgr,
	}
}

// AddProvider adds an OAuth2 provider
func (om *OAuth2Manager) AddProvider(provider *OAuth2Provider) error {
	if provider.ID == "" {
		return fmt.Errorf("provider ID is required")
	}

	if provider.ClientID == "" || provider.ClientSecret == "" {
		return fmt.Errorf("client ID and secret are required")
	}

	// Set default scopes if not provided
	if len(provider.Scopes) == 0 {
		provider.Scopes = []string{"openid", "profile", "email"}
	}

	// Configure provider-specific defaults
	if err := om.configureProviderDefaults(provider); err != nil {
		return fmt.Errorf("failed to configure provider defaults: %w", err)
	}

	om.providers[provider.ID] = provider

	log.Info().
		Str("provider_id", provider.ID).
		Str("provider_type", provider.Type).
		Msg("OAuth2 provider added")

	return nil
}

// configureProviderDefaults sets provider-specific default configurations
func (om *OAuth2Manager) configureProviderDefaults(provider *OAuth2Provider) error {
	switch provider.Type {
	case "google":
		if provider.AuthURL == "" {
			provider.AuthURL = "https://accounts.google.com/o/oauth2/v2/auth"
		}
		if provider.TokenURL == "" {
			provider.TokenURL = "https://oauth2.googleapis.com/token"
		}
		if provider.UserInfoURL == "" {
			provider.UserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
		}
		if provider.JWKSURL == "" {
			provider.JWKSURL = "https://www.googleapis.com/oauth2/v3/certs"
		}
		if provider.Issuer == "" {
			provider.Issuer = "https://accounts.google.com"
		}

	case "microsoft":
		if provider.AuthURL == "" {
			provider.AuthURL = "https://login.microsoftonline.com/common/oauth2/v2.0/authorize"
		}
		if provider.TokenURL == "" {
			provider.TokenURL = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
		}
		if provider.UserInfoURL == "" {
			provider.UserInfoURL = "https://graph.microsoft.com/v1.0/me"
		}
		if provider.JWKSURL == "" {
			provider.JWKSURL = "https://login.microsoftonline.com/common/discovery/v2.0/keys"
		}
		if provider.Issuer == "" {
			provider.Issuer = "https://login.microsoftonline.com/common/v2.0"
		}

	case "okta":
		// Okta requires custom domain configuration
		if provider.Config["domain"] == "" {
			return fmt.Errorf("Okta domain is required in config")
		}
		domain := provider.Config["domain"]
		if provider.AuthURL == "" {
			provider.AuthURL = fmt.Sprintf("https://%s/oauth2/v1/authorize", domain)
		}
		if provider.TokenURL == "" {
			provider.TokenURL = fmt.Sprintf("https://%s/oauth2/v1/token", domain)
		}
		if provider.UserInfoURL == "" {
			provider.UserInfoURL = fmt.Sprintf("https://%s/oauth2/v1/userinfo", domain)
		}
		if provider.JWKSURL == "" {
			provider.JWKSURL = fmt.Sprintf("https://%s/oauth2/v1/keys", domain)
		}
		if provider.Issuer == "" {
			provider.Issuer = fmt.Sprintf("https://%s", domain)
		}

	case "auth0":
		// Auth0 requires custom domain configuration
		if provider.Config["domain"] == "" {
			return fmt.Errorf("Auth0 domain is required in config")
		}
		domain := provider.Config["domain"]
		if provider.AuthURL == "" {
			provider.AuthURL = fmt.Sprintf("https://%s/authorize", domain)
		}
		if provider.TokenURL == "" {
			provider.TokenURL = fmt.Sprintf("https://%s/oauth/token", domain)
		}
		if provider.UserInfoURL == "" {
			provider.UserInfoURL = fmt.Sprintf("https://%s/userinfo", domain)
		}
		if provider.JWKSURL == "" {
			provider.JWKSURL = fmt.Sprintf("https://%s/.well-known/jwks.json", domain)
		}
		if provider.Issuer == "" {
			provider.Issuer = fmt.Sprintf("https://%s/", domain)
		}
	}

	return nil
}

// GetAuthURL generates an OAuth2 authorization URL
func (om *OAuth2Manager) GetAuthURL(providerID, redirectURL string) (string, error) {
	provider, exists := om.providers[providerID]
	if !exists {
		return "", fmt.Errorf("provider not found: %s", providerID)
	}

	if !provider.Enabled {
		return "", fmt.Errorf("provider is disabled: %s", providerID)
	}

	// Generate state for CSRF protection
	state, err := om.generateState(providerID, redirectURL)
	if err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}

	// Build authorization URL manually
	authURL, err := url.Parse(provider.AuthURL)
	if err != nil {
		return "", fmt.Errorf("invalid auth URL: %w", err)
	}

	query := authURL.Query()
	query.Set("client_id", provider.ClientID)
	query.Set("redirect_uri", provider.RedirectURL)
	query.Set("response_type", "code")
	query.Set("scope", strings.Join(provider.Scopes, " "))
	query.Set("state", state)
	query.Set("access_type", "offline")
	authURL.RawQuery = query.Encode()

	log.Info().
		Str("provider_id", providerID).
		Str("state", state).
		Msg("Generated OAuth2 authorization URL")

	return authURL.String(), nil
}

// generateState generates a secure random state for OAuth2 CSRF protection
func (om *OAuth2Manager) generateState(providerID, redirectURL string) (string, error) {
	// Generate random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	state := base64.URLEncoding.EncodeToString(bytes)

	// Store state with expiration
	om.states[state] = &OAuth2State{
		State:       state,
		Provider:    providerID,
		RedirectURL: redirectURL,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(10 * time.Minute), // 10 minute expiration
	}

	return state, nil
}

// HandleCallback handles OAuth2 callback and exchanges code for token
func (om *OAuth2Manager) HandleCallback(code, state string) (*auth.AuthContext, error) {
	// Validate state
	stateInfo, exists := om.states[state]
	if !exists {
		return nil, fmt.Errorf("invalid state")
	}

	// Check state expiration
	if time.Now().After(stateInfo.ExpiresAt) {
		delete(om.states, state)
		return nil, fmt.Errorf("state expired")
	}

	// Clean up state
	delete(om.states, state)

	provider, exists := om.providers[stateInfo.Provider]
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", stateInfo.Provider)
	}

	// Exchange code for token
	userInfo, err := om.exchangeCodeForUserInfo(provider, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Create or update user
	authCtx, err := om.createOrUpdateUser(userInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to create/update user: %w", err)
	}

	log.Info().
		Str("provider_id", stateInfo.Provider).
		Str("user_id", userInfo.ID).
		Str("email", userInfo.Email).
		Msg("OAuth2 authentication successful")

	return authCtx, nil
}

// exchangeCodeForUserInfo exchanges authorization code for user information
func (om *OAuth2Manager) exchangeCodeForUserInfo(provider *OAuth2Provider, code string) (*OAuth2UserInfo, error) {
	// Exchange code for token manually
	tokenResp, err := om.exchangeCodeForToken(provider, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Get user info
	userInfo, err := om.getUserInfo(provider, tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	userInfo.ProviderID = provider.ID
	userInfo.ProviderType = provider.Type

	return userInfo, nil
}

// TokenResponse represents an OAuth2 token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// exchangeCodeForToken exchanges authorization code for access token
func (om *OAuth2Manager) exchangeCodeForToken(provider *OAuth2Provider, code string) (*TokenResponse, error) {
	// Prepare token request
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", provider.RedirectURL)
	data.Set("client_id", provider.ClientID)
	data.Set("client_secret", provider.ClientSecret)

	// Make token request
	req, err := http.NewRequest("POST", provider.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse token response
	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// getUserInfo retrieves user information from the OAuth2 provider
func (om *OAuth2Manager) getUserInfo(provider *OAuth2Provider, accessToken string) (*OAuth2UserInfo, error) {
	// Create HTTP request
	req, err := http.NewRequest("GET", provider.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	// Parse response
	var rawUserInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawUserInfo); err != nil {
		return nil, err
	}

	// Map provider-specific response to standard format
	return om.mapUserInfo(provider, rawUserInfo)
}

// mapUserInfo maps provider-specific user info to standard format
func (om *OAuth2Manager) mapUserInfo(provider *OAuth2Provider, rawInfo map[string]interface{}) (*OAuth2UserInfo, error) {
	userInfo := &OAuth2UserInfo{
		Attributes: make(map[string]string),
	}

	// Common mappings
	if id, ok := rawInfo["id"].(string); ok {
		userInfo.ID = id
	} else if sub, ok := rawInfo["sub"].(string); ok {
		userInfo.ID = sub
	}

	if email, ok := rawInfo["email"].(string); ok {
		userInfo.Email = email
	}

	if name, ok := rawInfo["name"].(string); ok {
		userInfo.Name = name
	}

	if picture, ok := rawInfo["picture"].(string); ok {
		userInfo.Picture = picture
	}

	if verified, ok := rawInfo["email_verified"].(bool); ok {
		userInfo.Verified = verified
	}

	// Provider-specific mappings
	switch provider.Type {
	case "google":
		if givenName, ok := rawInfo["given_name"].(string); ok {
			userInfo.FirstName = givenName
		}
		if familyName, ok := rawInfo["family_name"].(string); ok {
			userInfo.LastName = familyName
		}
		if locale, ok := rawInfo["locale"].(string); ok {
			userInfo.Locale = locale
		}

	case "microsoft":
		if givenName, ok := rawInfo["givenName"].(string); ok {
			userInfo.FirstName = givenName
		}
		if surname, ok := rawInfo["surname"].(string); ok {
			userInfo.LastName = surname
		}
		if userPrincipalName, ok := rawInfo["userPrincipalName"].(string); ok {
			userInfo.Attributes["upn"] = userPrincipalName
		}
	}

	// Store all attributes for custom mapping
	for key, value := range rawInfo {
		if str, ok := value.(string); ok {
			userInfo.Attributes[key] = str
		}
	}

	return userInfo, nil
}

// createOrUpdateUser creates or updates a user based on OAuth2 user info
func (om *OAuth2Manager) createOrUpdateUser(userInfo *OAuth2UserInfo) (*auth.AuthContext, error) {
	// Generate user ID based on provider and external ID
	userID := fmt.Sprintf("%s:%s", userInfo.ProviderType, userInfo.ID)

	// Check if user exists
	existingUser, err := om.authMgr.GetUser(userID)
	if err != nil {
		// Check if it's a "user not found" error
		if authErr, ok := err.(auth.AuthError); ok && authErr.Code == "USER_NOT_FOUND" {
			// User doesn't exist, we'll create them
			existingUser = nil
		} else {
			return nil, fmt.Errorf("failed to check existing user: %w", err)
		}
	}

	// Prepare user data
	userData := &auth.User{
		ID:       userID,
		Username: userInfo.Email,
		Email:    userInfo.Email,
		Role:     auth.RoleUser, // Default role for SSO users
		Active:   true,
		Metadata: map[string]string{
			"provider_id":   userInfo.ProviderID,
			"provider_type": userInfo.ProviderType,
			"external_id":   userInfo.ID,
			"name":          userInfo.Name,
			"first_name":    userInfo.FirstName,
			"last_name":     userInfo.LastName,
			"picture":       userInfo.Picture,
			"locale":        userInfo.Locale,
			"verified":      fmt.Sprintf("%t", userInfo.Verified),
			"auth_method":   "oauth2",
		},
		Permissions: []string{
			auth.PermissionModelRead,
			auth.PermissionInferenceWrite,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Copy additional attributes
	for key, value := range userInfo.Attributes {
		userData.Metadata["attr_"+key] = value
	}

	if existingUser != nil {
		// Update existing user
		userData.CreatedAt = existingUser.CreatedAt
		if err := om.authMgr.UpdateUser(userData); err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
	} else {
		// Create new user using CreateUserRequest
		createReq := &auth.CreateUserRequest{
			Username: userInfo.Email,
			Email:    userInfo.Email,
			Password: "oauth2-sso", // Placeholder password for SSO users
			Role:     auth.RoleUser,
			Permissions: []string{
				auth.PermissionModelRead,
				auth.PermissionInferenceWrite,
			},
			Metadata: userData.Metadata,
		}

		createdUser, err := om.authMgr.CreateUser(createReq)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}

		// Update the user ID to match our SSO format
		createdUser.ID = userID
		if err := om.authMgr.UpdateUser(createdUser); err != nil {
			return nil, fmt.Errorf("failed to update user ID: %w", err)
		}
		userData = createdUser
	}

	// Authenticate user
	authCtx, err := om.authMgr.AuthenticateUser(userData)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	return authCtx, nil
}

// GetProviders returns all configured OAuth2 providers
func (om *OAuth2Manager) GetProviders() map[string]*OAuth2Provider {
	providers := make(map[string]*OAuth2Provider)
	for id, provider := range om.providers {
		if provider.Enabled {
			// Return sanitized version without secrets
			providers[id] = &OAuth2Provider{
				ID:          provider.ID,
				Name:        provider.Name,
				Type:        provider.Type,
				RedirectURL: provider.RedirectURL,
				Scopes:      provider.Scopes,
				Enabled:     provider.Enabled,
			}
		}
	}
	return providers
}

// CleanupExpiredStates removes expired OAuth2 states
func (om *OAuth2Manager) CleanupExpiredStates() {
	now := time.Now()
	for state, stateInfo := range om.states {
		if now.After(stateInfo.ExpiresAt) {
			delete(om.states, state)
		}
	}
}

package sso

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/auth"
	"github.com/rs/zerolog/log"
)

// LDAPProvider represents an LDAP/Active Directory provider
type LDAPProvider struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Host             string            `json:"host"`
	Port             int               `json:"port"`
	UseSSL           bool              `json:"use_ssl"`
	UseTLS           bool              `json:"use_tls"`
	SkipVerify       bool              `json:"skip_verify"`
	BindDN           string            `json:"bind_dn"`
	BindPassword     string            `json:"bind_password"`
	BaseDN           string            `json:"base_dn"`
	UserFilter       string            `json:"user_filter"`
	GroupFilter      string            `json:"group_filter"`
	AttributeMapping map[string]string `json:"attribute_mapping"`
	GroupMembership  string            `json:"group_membership"` // memberOf, member, etc.
	Enabled          bool              `json:"enabled"`
	Config           map[string]string `json:"config"`
}

// LDAPManager manages LDAP authentication
type LDAPManager struct {
	providers map[string]*LDAPProvider
	authMgr   *auth.Manager
}

// LDAPConnection represents an LDAP connection
type LDAPConnection struct {
	conn net.Conn
}

// LDAPUserInfo represents user information from LDAP
type LDAPUserInfo struct {
	DN         string            `json:"dn"`
	Username   string            `json:"username"`
	Email      string            `json:"email"`
	Name       string            `json:"name"`
	FirstName  string            `json:"first_name"`
	LastName   string            `json:"last_name"`
	Groups     []string          `json:"groups"`
	Attributes map[string]string `json:"attributes"`
	Provider   string            `json:"provider"`
}

// NewLDAPManager creates a new LDAP manager
func NewLDAPManager(authMgr *auth.Manager) *LDAPManager {
	return &LDAPManager{
		providers: make(map[string]*LDAPProvider),
		authMgr:   authMgr,
	}
}

// AddProvider adds an LDAP provider
func (lm *LDAPManager) AddProvider(provider *LDAPProvider) error {
	if provider.ID == "" {
		return fmt.Errorf("provider ID is required")
	}

	if provider.Host == "" {
		return fmt.Errorf("LDAP host is required")
	}

	// Set default port
	if provider.Port == 0 {
		if provider.UseSSL {
			provider.Port = 636
		} else {
			provider.Port = 389
		}
	}

	// Set default user filter
	if provider.UserFilter == "" {
		provider.UserFilter = "(&(objectClass=person)(|(sAMAccountName=%s)(userPrincipalName=%s)(mail=%s)))"
	}

	// Set default group filter
	if provider.GroupFilter == "" {
		provider.GroupFilter = "(&(objectClass=group)(member=%s))"
	}

	// Set default attribute mapping
	if provider.AttributeMapping == nil {
		provider.AttributeMapping = map[string]string{
			"username":   "sAMAccountName",
			"email":      "mail",
			"name":       "displayName",
			"first_name": "givenName",
			"last_name":  "sn",
			"groups":     "memberOf",
		}
	}

	// Set default group membership attribute
	if provider.GroupMembership == "" {
		provider.GroupMembership = "memberOf"
	}

	lm.providers[provider.ID] = provider

	log.Info().
		Str("provider_id", provider.ID).
		Str("host", provider.Host).
		Int("port", provider.Port).
		Msg("LDAP provider added")

	return nil
}

// Authenticate authenticates a user against LDAP
func (lm *LDAPManager) Authenticate(providerID, username, password string) (*auth.AuthContext, error) {
	provider, exists := lm.providers[providerID]
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", providerID)
	}

	if !provider.Enabled {
		return nil, fmt.Errorf("provider is disabled: %s", providerID)
	}

	// Connect to LDAP server
	conn, err := lm.connect(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to LDAP: %w", err)
	}
	defer conn.Close()

	// Bind with service account
	if err := lm.bind(conn, provider.BindDN, provider.BindPassword); err != nil {
		return nil, fmt.Errorf("failed to bind with service account: %w", err)
	}

	// Search for user
	userInfo, err := lm.searchUser(conn, provider, username)
	if err != nil {
		return nil, fmt.Errorf("user search failed: %w", err)
	}

	// Authenticate user
	if err := lm.bind(conn, userInfo.DN, password); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Get user groups
	groups, err := lm.getUserGroups(conn, provider, userInfo.DN)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get user groups")
		// Continue without groups
	} else {
		userInfo.Groups = groups
	}

	// Create or update user
	authCtx, err := lm.createOrUpdateUser(userInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to create/update user: %w", err)
	}

	log.Info().
		Str("provider_id", providerID).
		Str("username", username).
		Str("dn", userInfo.DN).
		Msg("LDAP authentication successful")

	return authCtx, nil
}

// connect establishes a connection to the LDAP server
func (lm *LDAPManager) connect(provider *LDAPProvider) (*LDAPConnection, error) {
	address := fmt.Sprintf("%s:%d", provider.Host, provider.Port)

	var conn net.Conn
	var err error

	if provider.UseSSL {
		// Direct SSL connection
		tlsConfig := &tls.Config{
			InsecureSkipVerify: provider.SkipVerify,
		}
		conn, err = tls.Dial("tcp", address, tlsConfig)
	} else {
		// Plain connection
		conn, err = net.DialTimeout("tcp", address, 30*time.Second)
		if err != nil {
			return nil, err
		}

		// Upgrade to TLS if requested
		if provider.UseTLS {
			tlsConfig := &tls.Config{
				InsecureSkipVerify: provider.SkipVerify,
			}
			conn = tls.Client(conn, tlsConfig)
		}
	}

	if err != nil {
		return nil, err
	}

	return &LDAPConnection{conn: conn}, nil
}

// bind performs LDAP bind operation
func (lm *LDAPManager) bind(conn *LDAPConnection, dn, password string) error {
	// This is a simplified LDAP bind implementation
	// In a real implementation, you would use a proper LDAP library
	// like github.com/go-ldap/ldap/v3

	// For now, we'll simulate a successful bind for demonstration
	if dn == "" || password == "" {
		return fmt.Errorf("invalid credentials")
	}

	// TODO: Implement actual LDAP bind protocol
	log.Debug().
		Str("dn", dn).
		Msg("LDAP bind (simulated)")

	return nil
}

// searchUser searches for a user in LDAP
func (lm *LDAPManager) searchUser(conn *LDAPConnection, provider *LDAPProvider, username string) (*LDAPUserInfo, error) {
	// This is a simplified LDAP search implementation
	// In a real implementation, you would use a proper LDAP library

	// Simulate user search
	userInfo := &LDAPUserInfo{
		DN:         fmt.Sprintf("CN=%s,%s", username, provider.BaseDN),
		Username:   username,
		Email:      fmt.Sprintf("%s@example.com", username),
		Name:       fmt.Sprintf("User %s", username),
		FirstName:  "User",
		LastName:   username,
		Attributes: make(map[string]string),
		Provider:   provider.ID,
	}

	// TODO: Implement actual LDAP search protocol
	log.Debug().
		Str("username", username).
		Str("base_dn", provider.BaseDN).
		Msg("LDAP user search (simulated)")

	return userInfo, nil
}

// getUserGroups retrieves user group memberships
func (lm *LDAPManager) getUserGroups(conn *LDAPConnection, provider *LDAPProvider, userDN string) ([]string, error) {
	// This is a simplified LDAP group search implementation
	// In a real implementation, you would use a proper LDAP library

	// Simulate group search
	groups := []string{
		"Domain Users",
		"OllamaMax Users",
	}

	// TODO: Implement actual LDAP group search protocol
	log.Debug().
		Str("user_dn", userDN).
		Msg("LDAP group search (simulated)")

	return groups, nil
}

// createOrUpdateUser creates or updates a user based on LDAP user info
func (lm *LDAPManager) createOrUpdateUser(userInfo *LDAPUserInfo) (*auth.AuthContext, error) {
	// Generate user ID based on provider and username
	userID := fmt.Sprintf("ldap:%s:%s", userInfo.Provider, userInfo.Username)

	// Check if user exists
	existingUser, err := lm.authMgr.GetUser(userID)
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
		Username: userInfo.Username,
		Email:    userInfo.Email,
		Role:     auth.RoleUser, // Default role for LDAP users
		Active:   true,
		Metadata: map[string]string{
			"provider":    userInfo.Provider,
			"dn":          userInfo.DN,
			"name":        userInfo.Name,
			"first_name":  userInfo.FirstName,
			"last_name":   userInfo.LastName,
			"auth_method": "ldap",
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

	// Add groups as metadata
	if len(userInfo.Groups) > 0 {
		userData.Metadata["groups"] = strings.Join(userInfo.Groups, ",")
	}

	if existingUser != nil {
		// Update existing user
		userData.CreatedAt = existingUser.CreatedAt
		if err := lm.authMgr.UpdateUser(userData); err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
	} else {
		// Create new user using CreateUserRequest
		createReq := &auth.CreateUserRequest{
			Username: userInfo.Username,
			Email:    userInfo.Email,
			Password: "ldap-sso", // Placeholder password for SSO users
			Role:     auth.RoleUser,
			Permissions: []string{
				auth.PermissionModelRead,
				auth.PermissionInferenceWrite,
			},
			Metadata: userData.Metadata,
		}

		createdUser, err := lm.authMgr.CreateUser(createReq)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}

		// Update the user ID to match our SSO format
		createdUser.ID = userID
		if err := lm.authMgr.UpdateUser(createdUser); err != nil {
			return nil, fmt.Errorf("failed to update user ID: %w", err)
		}
		userData = createdUser
	}

	// Authenticate user
	authCtx, err := lm.authMgr.AuthenticateUser(userData)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	return authCtx, nil
}

// GetProviders returns all configured LDAP providers
func (lm *LDAPManager) GetProviders() map[string]*LDAPProvider {
	providers := make(map[string]*LDAPProvider)
	for id, provider := range lm.providers {
		if provider.Enabled {
			// Return sanitized version without secrets
			providers[id] = &LDAPProvider{
				ID:               provider.ID,
				Name:             provider.Name,
				Host:             provider.Host,
				Port:             provider.Port,
				UseSSL:           provider.UseSSL,
				UseTLS:           provider.UseTLS,
				BaseDN:           provider.BaseDN,
				UserFilter:       provider.UserFilter,
				GroupFilter:      provider.GroupFilter,
				AttributeMapping: provider.AttributeMapping,
				GroupMembership:  provider.GroupMembership,
				Enabled:          provider.Enabled,
			}
		}
	}
	return providers
}

// Close closes the LDAP connection
func (conn *LDAPConnection) Close() error {
	if conn.conn != nil {
		return conn.conn.Close()
	}
	return nil
}

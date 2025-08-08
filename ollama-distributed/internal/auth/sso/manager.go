package sso

import (
	"fmt"
	"sync"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/auth"
	"github.com/rs/zerolog/log"
)

// Manager manages all SSO providers
type Manager struct {
	oauth2Manager *OAuth2Manager
	samlManager   *SAMLManager
	ldapManager   *LDAPManager
	authMgr       *auth.Manager
	mu            sync.RWMutex
}

// Config represents SSO configuration
type Config struct {
	OAuth2Providers []*OAuth2Provider `json:"oauth2_providers"`
	SAMLProviders   []*SAMLProvider   `json:"saml_providers"`
	LDAPProviders   []*LDAPProvider   `json:"ldap_providers"`
	EntityID        string            `json:"entity_id"`
	ACSURL          string            `json:"acs_url"`
	SLOURL          string            `json:"slo_url"`
}

// ProviderInfo represents basic provider information
type ProviderInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"` // oauth2, saml, ldap
	Enabled bool   `json:"enabled"`
}

// NewManager creates a new SSO manager
func NewManager(authMgr *auth.Manager, config *Config) (*Manager, error) {
	if authMgr == nil {
		return nil, fmt.Errorf("auth manager is required")
	}

	if config == nil {
		config = &Config{}
	}

	// Set default values
	if config.EntityID == "" {
		config.EntityID = "https://ollamamax.local"
	}
	if config.ACSURL == "" {
		config.ACSURL = "https://ollamamax.local/auth/saml/acs"
	}
	if config.SLOURL == "" {
		config.SLOURL = "https://ollamamax.local/auth/saml/slo"
	}

	manager := &Manager{
		oauth2Manager: NewOAuth2Manager(authMgr),
		samlManager:   NewSAMLManager(authMgr, config.EntityID, config.ACSURL, config.SLOURL),
		ldapManager:   NewLDAPManager(authMgr),
		authMgr:       authMgr,
	}

	// Configure OAuth2 providers
	for _, provider := range config.OAuth2Providers {
		if err := manager.oauth2Manager.AddProvider(provider); err != nil {
			log.Error().Err(err).Str("provider_id", provider.ID).Msg("Failed to add OAuth2 provider")
		}
	}

	// Configure SAML providers
	for _, provider := range config.SAMLProviders {
		if err := manager.samlManager.AddProvider(provider); err != nil {
			log.Error().Err(err).Str("provider_id", provider.ID).Msg("Failed to add SAML provider")
		}
	}

	// Configure LDAP providers
	for _, provider := range config.LDAPProviders {
		if err := manager.ldapManager.AddProvider(provider); err != nil {
			log.Error().Err(err).Str("provider_id", provider.ID).Msg("Failed to add LDAP provider")
		}
	}

	log.Info().
		Int("oauth2_providers", len(config.OAuth2Providers)).
		Int("saml_providers", len(config.SAMLProviders)).
		Int("ldap_providers", len(config.LDAPProviders)).
		Msg("SSO manager initialized")

	return manager, nil
}

// GetOAuth2AuthURL generates an OAuth2 authorization URL
func (m *Manager) GetOAuth2AuthURL(providerID, redirectURL string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.oauth2Manager.GetAuthURL(providerID, redirectURL)
}

// HandleOAuth2Callback handles OAuth2 callback
func (m *Manager) HandleOAuth2Callback(code, state string) (*auth.AuthContext, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.oauth2Manager.HandleCallback(code, state)
}

// GetSAMLAuthURL generates a SAML authentication URL
func (m *Manager) GetSAMLAuthURL(providerID, relayState string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.samlManager.GetAuthURL(providerID, relayState)
}

// HandleSAMLResponse handles SAML response
func (m *Manager) HandleSAMLResponse(samlResponse, relayState string) (*auth.AuthContext, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.samlManager.HandleResponse(samlResponse, relayState)
}

// GetSAMLMetadata returns SAML metadata
func (m *Manager) GetSAMLMetadata() (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.samlManager.GetMetadata()
}

// AuthenticateLDAP authenticates a user against LDAP
func (m *Manager) AuthenticateLDAP(providerID, username, password string) (*auth.AuthContext, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.ldapManager.Authenticate(providerID, username, password)
}

// GetProviders returns all configured providers
func (m *Manager) GetProviders() map[string]*ProviderInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	providers := make(map[string]*ProviderInfo)

	// OAuth2 providers
	for id, provider := range m.oauth2Manager.GetProviders() {
		providers[id] = &ProviderInfo{
			ID:      provider.ID,
			Name:    provider.Name,
			Type:    "oauth2",
			Enabled: provider.Enabled,
		}
	}

	// SAML providers
	for id, provider := range m.samlManager.GetProviders() {
		providers[id] = &ProviderInfo{
			ID:      provider.ID,
			Name:    provider.Name,
			Type:    "saml",
			Enabled: provider.Enabled,
		}
	}

	// LDAP providers
	for id, provider := range m.ldapManager.GetProviders() {
		providers[id] = &ProviderInfo{
			ID:      provider.ID,
			Name:    provider.Name,
			Type:    "ldap",
			Enabled: provider.Enabled,
		}
	}

	return providers
}

// GetOAuth2Providers returns OAuth2 providers
func (m *Manager) GetOAuth2Providers() map[string]*OAuth2Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.oauth2Manager.GetProviders()
}

// GetSAMLProviders returns SAML providers
func (m *Manager) GetSAMLProviders() map[string]*SAMLProvider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.samlManager.GetProviders()
}

// GetLDAPProviders returns LDAP providers
func (m *Manager) GetLDAPProviders() map[string]*LDAPProvider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.ldapManager.GetProviders()
}

// AddOAuth2Provider adds an OAuth2 provider
func (m *Manager) AddOAuth2Provider(provider *OAuth2Provider) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.oauth2Manager.AddProvider(provider)
}

// AddSAMLProvider adds a SAML provider
func (m *Manager) AddSAMLProvider(provider *SAMLProvider) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.samlManager.AddProvider(provider)
}

// AddLDAPProvider adds an LDAP provider
func (m *Manager) AddLDAPProvider(provider *LDAPProvider) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.ldapManager.AddProvider(provider)
}

// CleanupExpiredStates removes expired OAuth2 states
func (m *Manager) CleanupExpiredStates() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.oauth2Manager.CleanupExpiredStates()
}

// ValidateProvider validates a provider configuration
func (m *Manager) ValidateProvider(providerType, providerID string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	switch providerType {
	case "oauth2":
		providers := m.oauth2Manager.GetProviders()
		if _, exists := providers[providerID]; !exists {
			return fmt.Errorf("OAuth2 provider not found: %s", providerID)
		}
	case "saml":
		providers := m.samlManager.GetProviders()
		if _, exists := providers[providerID]; !exists {
			return fmt.Errorf("SAML provider not found: %s", providerID)
		}
	case "ldap":
		providers := m.ldapManager.GetProviders()
		if _, exists := providers[providerID]; !exists {
			return fmt.Errorf("LDAP provider not found: %s", providerID)
		}
	default:
		return fmt.Errorf("unsupported provider type: %s", providerType)
	}

	return nil
}

// GetProviderByID returns a provider by ID and type
func (m *Manager) GetProviderByID(providerType, providerID string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	switch providerType {
	case "oauth2":
		providers := m.oauth2Manager.GetProviders()
		if provider, exists := providers[providerID]; exists {
			return provider, nil
		}
	case "saml":
		providers := m.samlManager.GetProviders()
		if provider, exists := providers[providerID]; exists {
			return provider, nil
		}
	case "ldap":
		providers := m.ldapManager.GetProviders()
		if provider, exists := providers[providerID]; exists {
			return provider, nil
		}
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}

	return nil, fmt.Errorf("provider not found: %s/%s", providerType, providerID)
}

// GetStats returns SSO usage statistics
func (m *Manager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	oauth2Providers := m.oauth2Manager.GetProviders()
	samlProviders := m.samlManager.GetProviders()
	ldapProviders := m.ldapManager.GetProviders()

	return map[string]interface{}{
		"total_providers": len(oauth2Providers) + len(samlProviders) + len(ldapProviders),
		"oauth2_providers": len(oauth2Providers),
		"saml_providers":   len(samlProviders),
		"ldap_providers":   len(ldapProviders),
		"provider_types":   []string{"oauth2", "saml", "ldap"},
	}
}

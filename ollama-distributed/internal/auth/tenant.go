package auth

import (
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// Tenant represents a tenant in the multi-tenant system
type Tenant struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Domain      string            `json:"domain"`
	Status      TenantStatus      `json:"status"`
	Settings    TenantSettings    `json:"settings"`
	Quotas      TenantQuotas      `json:"quotas"`
	Usage       TenantUsage       `json:"usage"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	CreatedBy   string            `json:"created_by"`
	ContactInfo ContactInfo       `json:"contact_info"`
}

// TenantStatus represents the status of a tenant
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusInactive  TenantStatus = "inactive"
	TenantStatusDeleted   TenantStatus = "deleted"
)

// TenantSettings represents tenant-specific settings
type TenantSettings struct {
	AllowedAuthMethods []string          `json:"allowed_auth_methods"`
	SessionTimeout     time.Duration     `json:"session_timeout"`
	PasswordPolicy     PasswordPolicy    `json:"password_policy"`
	SSOConfig          map[string]string `json:"sso_config"`
	Features           []string          `json:"features"`
	CustomBranding     CustomBranding    `json:"custom_branding"`
	SecuritySettings   SecuritySettings  `json:"security_settings"`
}

// TenantQuotas represents resource quotas for a tenant
type TenantQuotas struct {
	MaxUsers           int   `json:"max_users"`
	MaxAPIKeys         int   `json:"max_api_keys"`
	MaxSessions        int   `json:"max_sessions"`
	MaxModels          int   `json:"max_models"`
	MaxInferences      int64 `json:"max_inferences_per_month"`
	MaxStorage         int64 `json:"max_storage_bytes"`
	MaxBandwidth       int64 `json:"max_bandwidth_bytes_per_month"`
	MaxConcurrentReqs  int   `json:"max_concurrent_requests"`
}

// TenantUsage represents current usage statistics for a tenant
type TenantUsage struct {
	CurrentUsers      int   `json:"current_users"`
	CurrentAPIKeys    int   `json:"current_api_keys"`
	CurrentSessions   int   `json:"current_sessions"`
	CurrentModels     int   `json:"current_models"`
	InferencesThisMonth int64 `json:"inferences_this_month"`
	StorageUsed       int64 `json:"storage_used_bytes"`
	BandwidthThisMonth int64 `json:"bandwidth_this_month_bytes"`
	LastUpdated       time.Time `json:"last_updated"`
}

// ContactInfo represents tenant contact information
type ContactInfo struct {
	AdminEmail   string `json:"admin_email"`
	AdminName    string `json:"admin_name"`
	BillingEmail string `json:"billing_email"`
	SupportEmail string `json:"support_email"`
	Phone        string `json:"phone"`
	Address      string `json:"address"`
}

// PasswordPolicy represents password requirements for a tenant
type PasswordPolicy struct {
	MinLength        int  `json:"min_length"`
	RequireUppercase bool `json:"require_uppercase"`
	RequireLowercase bool `json:"require_lowercase"`
	RequireNumbers   bool `json:"require_numbers"`
	RequireSymbols   bool `json:"require_symbols"`
	MaxAge           int  `json:"max_age_days"`
	HistoryCount     int  `json:"history_count"`
}

// CustomBranding represents tenant-specific branding
type CustomBranding struct {
	LogoURL      string `json:"logo_url"`
	PrimaryColor string `json:"primary_color"`
	CompanyName  string `json:"company_name"`
	CustomCSS    string `json:"custom_css"`
}

// SecuritySettings represents tenant-specific security settings
type SecuritySettings struct {
	RequireMFA           bool     `json:"require_mfa"`
	AllowedIPRanges      []string `json:"allowed_ip_ranges"`
	SessionIdleTimeout   int      `json:"session_idle_timeout_minutes"`
	MaxLoginAttempts     int      `json:"max_login_attempts"`
	LockoutDuration      int      `json:"lockout_duration_minutes"`
	AuditLogRetention    int      `json:"audit_log_retention_days"`
	EncryptionAtRest     bool     `json:"encryption_at_rest"`
	EncryptionInTransit  bool     `json:"encryption_in_transit"`
}

// TenantManager manages multi-tenancy
type TenantManager struct {
	tenants map[string]*Tenant
	mu      sync.RWMutex
}

// CreateTenantRequest represents a request to create a new tenant
type CreateTenantRequest struct {
	Name        string            `json:"name" binding:"required"`
	Domain      string            `json:"domain" binding:"required"`
	AdminEmail  string            `json:"admin_email" binding:"required"`
	AdminName   string            `json:"admin_name" binding:"required"`
	Settings    *TenantSettings   `json:"settings,omitempty"`
	Quotas      *TenantQuotas     `json:"quotas,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	ContactInfo *ContactInfo      `json:"contact_info,omitempty"`
}

// UpdateTenantRequest represents a request to update a tenant
type UpdateTenantRequest struct {
	Name        *string           `json:"name,omitempty"`
	Domain      *string           `json:"domain,omitempty"`
	Status      *TenantStatus     `json:"status,omitempty"`
	Settings    *TenantSettings   `json:"settings,omitempty"`
	Quotas      *TenantQuotas     `json:"quotas,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	ContactInfo *ContactInfo      `json:"contact_info,omitempty"`
}

// NewTenantManager creates a new tenant manager
func NewTenantManager() *TenantManager {
	return &TenantManager{
		tenants: make(map[string]*Tenant),
	}
}

// CreateTenant creates a new tenant
func (tm *TenantManager) CreateTenant(req *CreateTenantRequest, createdBy string) (*Tenant, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Generate tenant ID
	tenantID := generateTenantID()

	// Check if domain is already used
	for _, tenant := range tm.tenants {
		if tenant.Domain == req.Domain {
			return nil, fmt.Errorf("domain already exists: %s", req.Domain)
		}
	}

	// Set default settings
	settings := req.Settings
	if settings == nil {
		settings = &TenantSettings{
			AllowedAuthMethods: []string{"password", "api_key"},
			SessionTimeout:     24 * time.Hour,
			PasswordPolicy: PasswordPolicy{
				MinLength:        8,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireNumbers:   true,
				RequireSymbols:   false,
				MaxAge:           90,
				HistoryCount:     5,
			},
			Features: []string{"basic"},
			SecuritySettings: SecuritySettings{
				RequireMFA:          false,
				SessionIdleTimeout:  30,
				MaxLoginAttempts:    5,
				LockoutDuration:     15,
				AuditLogRetention:   90,
				EncryptionAtRest:    true,
				EncryptionInTransit: true,
			},
		}
	}

	// Set default quotas
	quotas := req.Quotas
	if quotas == nil {
		quotas = &TenantQuotas{
			MaxUsers:          100,
			MaxAPIKeys:        50,
			MaxSessions:       500,
			MaxModels:         10,
			MaxInferences:     10000,
			MaxStorage:        10 * 1024 * 1024 * 1024, // 10GB
			MaxBandwidth:      100 * 1024 * 1024 * 1024, // 100GB
			MaxConcurrentReqs: 10,
		}
	}

	// Set contact info
	contactInfo := req.ContactInfo
	if contactInfo == nil {
		contactInfo = &ContactInfo{
			AdminEmail: req.AdminEmail,
			AdminName:  req.AdminName,
		}
	}

	tenant := &Tenant{
		ID:       tenantID,
		Name:     req.Name,
		Domain:   req.Domain,
		Status:   TenantStatusActive,
		Settings: *settings,
		Quotas:   *quotas,
		Usage: TenantUsage{
			LastUpdated: time.Now(),
		},
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CreatedBy:   createdBy,
		ContactInfo: *contactInfo,
	}

	tm.tenants[tenantID] = tenant

	log.Info().
		Str("tenant_id", tenantID).
		Str("name", req.Name).
		Str("domain", req.Domain).
		Str("created_by", createdBy).
		Msg("Tenant created")

	return tenant, nil
}

// GetTenant retrieves a tenant by ID
func (tm *TenantManager) GetTenant(tenantID string) (*Tenant, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tenant, exists := tm.tenants[tenantID]
	if !exists {
		return nil, fmt.Errorf("tenant not found: %s", tenantID)
	}

	return tenant, nil
}

// GetTenantByDomain retrieves a tenant by domain
func (tm *TenantManager) GetTenantByDomain(domain string) (*Tenant, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	for _, tenant := range tm.tenants {
		if tenant.Domain == domain {
			return tenant, nil
		}
	}

	return nil, fmt.Errorf("tenant not found for domain: %s", domain)
}

// UpdateTenant updates an existing tenant
func (tm *TenantManager) UpdateTenant(tenantID string, req *UpdateTenantRequest) (*Tenant, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tenant, exists := tm.tenants[tenantID]
	if !exists {
		return nil, fmt.Errorf("tenant not found: %s", tenantID)
	}

	// Update fields if provided
	if req.Name != nil {
		tenant.Name = *req.Name
	}
	if req.Domain != nil {
		// Check if new domain is already used
		for id, t := range tm.tenants {
			if id != tenantID && t.Domain == *req.Domain {
				return nil, fmt.Errorf("domain already exists: %s", *req.Domain)
			}
		}
		tenant.Domain = *req.Domain
	}
	if req.Status != nil {
		tenant.Status = *req.Status
	}
	if req.Settings != nil {
		tenant.Settings = *req.Settings
	}
	if req.Quotas != nil {
		tenant.Quotas = *req.Quotas
	}
	if req.Metadata != nil {
		tenant.Metadata = req.Metadata
	}
	if req.ContactInfo != nil {
		tenant.ContactInfo = *req.ContactInfo
	}

	tenant.UpdatedAt = time.Now()

	log.Info().
		Str("tenant_id", tenantID).
		Msg("Tenant updated")

	return tenant, nil
}

// DeleteTenant marks a tenant as deleted
func (tm *TenantManager) DeleteTenant(tenantID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tenant, exists := tm.tenants[tenantID]
	if !exists {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}

	tenant.Status = TenantStatusDeleted
	tenant.UpdatedAt = time.Now()

	log.Info().
		Str("tenant_id", tenantID).
		Msg("Tenant marked as deleted")

	return nil
}

// ListTenants returns all tenants with optional filtering
func (tm *TenantManager) ListTenants(status *TenantStatus) []*Tenant {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var tenants []*Tenant
	for _, tenant := range tm.tenants {
		if status == nil || tenant.Status == *status {
			tenants = append(tenants, tenant)
		}
	}

	return tenants
}

// ValidateTenantAccess checks if a user has access to a tenant
func (tm *TenantManager) ValidateTenantAccess(tenantID, userID string) error {
	tenant, err := tm.GetTenant(tenantID)
	if err != nil {
		return err
	}

	if tenant.Status != TenantStatusActive {
		return fmt.Errorf("tenant is not active: %s", tenant.Status)
	}

	// TODO: Implement user-tenant relationship validation
	// This would check if the user belongs to the tenant

	return nil
}

// CheckQuota checks if a tenant has exceeded a specific quota
func (tm *TenantManager) CheckQuota(tenantID string, quotaType string, requestedAmount int64) error {
	tenant, err := tm.GetTenant(tenantID)
	if err != nil {
		return err
	}

	switch quotaType {
	case "users":
		if tenant.Usage.CurrentUsers >= tenant.Quotas.MaxUsers {
			return fmt.Errorf("user quota exceeded: %d/%d", tenant.Usage.CurrentUsers, tenant.Quotas.MaxUsers)
		}
	case "api_keys":
		if tenant.Usage.CurrentAPIKeys >= tenant.Quotas.MaxAPIKeys {
			return fmt.Errorf("API key quota exceeded: %d/%d", tenant.Usage.CurrentAPIKeys, tenant.Quotas.MaxAPIKeys)
		}
	case "sessions":
		if tenant.Usage.CurrentSessions >= tenant.Quotas.MaxSessions {
			return fmt.Errorf("session quota exceeded: %d/%d", tenant.Usage.CurrentSessions, tenant.Quotas.MaxSessions)
		}
	case "models":
		if tenant.Usage.CurrentModels >= tenant.Quotas.MaxModels {
			return fmt.Errorf("model quota exceeded: %d/%d", tenant.Usage.CurrentModels, tenant.Quotas.MaxModels)
		}
	case "inferences":
		if tenant.Usage.InferencesThisMonth+requestedAmount > tenant.Quotas.MaxInferences {
			return fmt.Errorf("inference quota exceeded: %d/%d", tenant.Usage.InferencesThisMonth, tenant.Quotas.MaxInferences)
		}
	case "storage":
		if tenant.Usage.StorageUsed+requestedAmount > tenant.Quotas.MaxStorage {
			return fmt.Errorf("storage quota exceeded: %d/%d bytes", tenant.Usage.StorageUsed, tenant.Quotas.MaxStorage)
		}
	case "bandwidth":
		if tenant.Usage.BandwidthThisMonth+requestedAmount > tenant.Quotas.MaxBandwidth {
			return fmt.Errorf("bandwidth quota exceeded: %d/%d bytes", tenant.Usage.BandwidthThisMonth, tenant.Quotas.MaxBandwidth)
		}
	default:
		return fmt.Errorf("unknown quota type: %s", quotaType)
	}

	return nil
}

// UpdateUsage updates tenant usage statistics
func (tm *TenantManager) UpdateUsage(tenantID string, usageType string, delta int64) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tenant, exists := tm.tenants[tenantID]
	if !exists {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}

	switch usageType {
	case "users":
		tenant.Usage.CurrentUsers = int(int64(tenant.Usage.CurrentUsers) + delta)
	case "api_keys":
		tenant.Usage.CurrentAPIKeys = int(int64(tenant.Usage.CurrentAPIKeys) + delta)
	case "sessions":
		tenant.Usage.CurrentSessions = int(int64(tenant.Usage.CurrentSessions) + delta)
	case "models":
		tenant.Usage.CurrentModels = int(int64(tenant.Usage.CurrentModels) + delta)
	case "inferences":
		tenant.Usage.InferencesThisMonth += delta
	case "storage":
		tenant.Usage.StorageUsed += delta
	case "bandwidth":
		tenant.Usage.BandwidthThisMonth += delta
	default:
		return fmt.Errorf("unknown usage type: %s", usageType)
	}

	tenant.Usage.LastUpdated = time.Now()
	return nil
}

// generateTenantID generates a unique tenant ID
func generateTenantID() string {
	return fmt.Sprintf("tenant_%d", time.Now().UnixNano())
}

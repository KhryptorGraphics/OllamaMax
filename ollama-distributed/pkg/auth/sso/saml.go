package sso

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/auth"
	"github.com/rs/zerolog/log"
)

// SAMLProvider represents a SAML 2.0 identity provider
type SAMLProvider struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	EntityID          string            `json:"entity_id"`
	SSOURL            string            `json:"sso_url"`
	SLOURL            string            `json:"slo_url"`
	Certificate       string            `json:"certificate"`
	SigningCert       string            `json:"signing_cert"`
	EncryptionCert    string            `json:"encryption_cert"`
	NameIDFormat      string            `json:"name_id_format"`
	AttributeMapping  map[string]string `json:"attribute_mapping"`
	RequireSignature  bool              `json:"require_signature"`
	RequireEncryption bool              `json:"require_encryption"`
	Enabled           bool              `json:"enabled"`
	Config            map[string]string `json:"config"`
}

// SAMLManager manages SAML 2.0 authentication
type SAMLManager struct {
	providers   map[string]*SAMLProvider
	authMgr     *auth.Manager
	entityID    string
	acsURL      string
	sloURL      string
	certificate *x509.Certificate
	privateKey  interface{}
}

// SAMLRequest represents a SAML authentication request
type SAMLRequest struct {
	XMLName                     xml.Name     `xml:"urn:oasis:names:tc:SAML:2.0:protocol AuthnRequest"`
	ID                          string       `xml:"ID,attr"`
	Version                     string       `xml:"Version,attr"`
	IssueInstant                time.Time    `xml:"IssueInstant,attr"`
	Destination                 string       `xml:"Destination,attr"`
	ProtocolBinding             string       `xml:"ProtocolBinding,attr"`
	AssertionConsumerServiceURL string       `xml:"AssertionConsumerServiceURL,attr"`
	Issuer                      Issuer       `xml:"urn:oasis:names:tc:SAML:2.0:assertion Issuer"`
	NameIDPolicy                NameIDPolicy `xml:"urn:oasis:names:tc:SAML:2.0:protocol NameIDPolicy"`
}

// SAMLResponse represents a SAML authentication response
type SAMLResponse struct {
	XMLName      xml.Name  `xml:"urn:oasis:names:tc:SAML:2.0:protocol Response"`
	ID           string    `xml:"ID,attr"`
	Version      string    `xml:"Version,attr"`
	IssueInstant time.Time `xml:"IssueInstant,attr"`
	Destination  string    `xml:"Destination,attr"`
	InResponseTo string    `xml:"InResponseTo,attr"`
	Issuer       Issuer    `xml:"urn:oasis:names:tc:SAML:2.0:assertion Issuer"`
	Status       Status    `xml:"urn:oasis:names:tc:SAML:2.0:protocol Status"`
	Assertion    Assertion `xml:"urn:oasis:names:tc:SAML:2.0:assertion Assertion"`
}

// Issuer represents the SAML issuer
type Issuer struct {
	XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion Issuer"`
	Value   string   `xml:",chardata"`
}

// NameIDPolicy represents the SAML NameID policy
type NameIDPolicy struct {
	XMLName     xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:protocol NameIDPolicy"`
	Format      string   `xml:"Format,attr"`
	AllowCreate bool     `xml:"AllowCreate,attr"`
}

// Status represents the SAML response status
type Status struct {
	XMLName    xml.Name   `xml:"urn:oasis:names:tc:SAML:2.0:protocol Status"`
	StatusCode StatusCode `xml:"urn:oasis:names:tc:SAML:2.0:protocol StatusCode"`
}

// StatusCode represents the SAML status code
type StatusCode struct {
	XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:protocol StatusCode"`
	Value   string   `xml:"Value,attr"`
}

// Assertion represents the SAML assertion
type Assertion struct {
	XMLName            xml.Name           `xml:"urn:oasis:names:tc:SAML:2.0:assertion Assertion"`
	ID                 string             `xml:"ID,attr"`
	Version            string             `xml:"Version,attr"`
	IssueInstant       time.Time          `xml:"IssueInstant,attr"`
	Issuer             Issuer             `xml:"urn:oasis:names:tc:SAML:2.0:assertion Issuer"`
	Subject            Subject            `xml:"urn:oasis:names:tc:SAML:2.0:assertion Subject"`
	Conditions         Conditions         `xml:"urn:oasis:names:tc:SAML:2.0:assertion Conditions"`
	AttributeStatement AttributeStatement `xml:"urn:oasis:names:tc:SAML:2.0:assertion AttributeStatement"`
	AuthnStatement     AuthnStatement     `xml:"urn:oasis:names:tc:SAML:2.0:assertion AuthnStatement"`
}

// Subject represents the SAML subject
type Subject struct {
	XMLName             xml.Name            `xml:"urn:oasis:names:tc:SAML:2.0:assertion Subject"`
	NameID              NameID              `xml:"urn:oasis:names:tc:SAML:2.0:assertion NameID"`
	SubjectConfirmation SubjectConfirmation `xml:"urn:oasis:names:tc:SAML:2.0:assertion SubjectConfirmation"`
}

// NameID represents the SAML NameID
type NameID struct {
	XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion NameID"`
	Format  string   `xml:"Format,attr"`
	Value   string   `xml:",chardata"`
}

// SubjectConfirmation represents the SAML subject confirmation
type SubjectConfirmation struct {
	XMLName                 xml.Name                `xml:"urn:oasis:names:tc:SAML:2.0:assertion SubjectConfirmation"`
	Method                  string                  `xml:"Method,attr"`
	SubjectConfirmationData SubjectConfirmationData `xml:"urn:oasis:names:tc:SAML:2.0:assertion SubjectConfirmationData"`
}

// SubjectConfirmationData represents the SAML subject confirmation data
type SubjectConfirmationData struct {
	XMLName      xml.Name  `xml:"urn:oasis:names:tc:SAML:2.0:assertion SubjectConfirmationData"`
	InResponseTo string    `xml:"InResponseTo,attr"`
	NotOnOrAfter time.Time `xml:"NotOnOrAfter,attr"`
	Recipient    string    `xml:"Recipient,attr"`
}

// Conditions represents the SAML conditions
type Conditions struct {
	XMLName             xml.Name            `xml:"urn:oasis:names:tc:SAML:2.0:assertion Conditions"`
	NotBefore           time.Time           `xml:"NotBefore,attr"`
	NotOnOrAfter        time.Time           `xml:"NotOnOrAfter,attr"`
	AudienceRestriction AudienceRestriction `xml:"urn:oasis:names:tc:SAML:2.0:assertion AudienceRestriction"`
}

// AudienceRestriction represents the SAML audience restriction
type AudienceRestriction struct {
	XMLName  xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion AudienceRestriction"`
	Audience Audience `xml:"urn:oasis:names:tc:SAML:2.0:assertion Audience"`
}

// Audience represents the SAML audience
type Audience struct {
	XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion Audience"`
	Value   string   `xml:",chardata"`
}

// AttributeStatement represents the SAML attribute statement
type AttributeStatement struct {
	XMLName    xml.Name    `xml:"urn:oasis:names:tc:SAML:2.0:assertion AttributeStatement"`
	Attributes []Attribute `xml:"urn:oasis:names:tc:SAML:2.0:assertion Attribute"`
}

// Attribute represents a SAML attribute
type Attribute struct {
	XMLName         xml.Name         `xml:"urn:oasis:names:tc:SAML:2.0:assertion Attribute"`
	Name            string           `xml:"Name,attr"`
	NameFormat      string           `xml:"NameFormat,attr"`
	AttributeValues []AttributeValue `xml:"urn:oasis:names:tc:SAML:2.0:assertion AttributeValue"`
}

// AttributeValue represents a SAML attribute value
type AttributeValue struct {
	XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion AttributeValue"`
	Type    string   `xml:"http://www.w3.org/2001/XMLSchema-instance type,attr"`
	Value   string   `xml:",chardata"`
}

// AuthnStatement represents the SAML authentication statement
type AuthnStatement struct {
	XMLName      xml.Name     `xml:"urn:oasis:names:tc:SAML:2.0:assertion AuthnStatement"`
	AuthnInstant time.Time    `xml:"AuthnInstant,attr"`
	SessionIndex string       `xml:"SessionIndex,attr"`
	AuthnContext AuthnContext `xml:"urn:oasis:names:tc:SAML:2.0:assertion AuthnContext"`
}

// AuthnContext represents the SAML authentication context
type AuthnContext struct {
	XMLName              xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion AuthnContext"`
	AuthnContextClassRef string   `xml:"urn:oasis:names:tc:SAML:2.0:assertion AuthnContextClassRef"`
}

// SAMLUserInfo represents user information from SAML assertion
type SAMLUserInfo struct {
	NameID     string            `json:"name_id"`
	Email      string            `json:"email"`
	Name       string            `json:"name"`
	FirstName  string            `json:"first_name"`
	LastName   string            `json:"last_name"`
	Groups     []string          `json:"groups"`
	Roles      []string          `json:"roles"`
	Attributes map[string]string `json:"attributes"`
	SessionID  string            `json:"session_id"`
	Provider   string            `json:"provider"`
}

// NewSAMLManager creates a new SAML manager
func NewSAMLManager(authMgr *auth.Manager, entityID, acsURL, sloURL string) *SAMLManager {
	return &SAMLManager{
		providers: make(map[string]*SAMLProvider),
		authMgr:   authMgr,
		entityID:  entityID,
		acsURL:    acsURL,
		sloURL:    sloURL,
	}
}

// AddProvider adds a SAML provider
func (sm *SAMLManager) AddProvider(provider *SAMLProvider) error {
	if provider.ID == "" {
		return fmt.Errorf("provider ID is required")
	}

	if provider.EntityID == "" || provider.SSOURL == "" {
		return fmt.Errorf("entity ID and SSO URL are required")
	}

	// Set default NameID format if not provided
	if provider.NameIDFormat == "" {
		provider.NameIDFormat = "urn:oasis:names:tc:SAML:2.0:nameid-format:emailAddress"
	}

	// Set default attribute mapping
	if provider.AttributeMapping == nil {
		provider.AttributeMapping = map[string]string{
			"email":      "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
			"name":       "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name",
			"first_name": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname",
			"last_name":  "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname",
			"groups":     "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/groups",
		}
	}

	sm.providers[provider.ID] = provider

	log.Info().
		Str("provider_id", provider.ID).
		Str("entity_id", provider.EntityID).
		Msg("SAML provider added")

	return nil
}

// GetAuthURL generates a SAML authentication URL
func (sm *SAMLManager) GetAuthURL(providerID, relayState string) (string, error) {
	provider, exists := sm.providers[providerID]
	if !exists {
		return "", fmt.Errorf("provider not found: %s", providerID)
	}

	if !provider.Enabled {
		return "", fmt.Errorf("provider is disabled: %s", providerID)
	}

	// Generate SAML request
	requestID := generateID()
	samlRequest := &SAMLRequest{
		ID:                          requestID,
		Version:                     "2.0",
		IssueInstant:                time.Now().UTC(),
		Destination:                 provider.SSOURL,
		ProtocolBinding:             "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST",
		AssertionConsumerServiceURL: sm.acsURL,
		Issuer: Issuer{
			Value: sm.entityID,
		},
		NameIDPolicy: NameIDPolicy{
			Format:      provider.NameIDFormat,
			AllowCreate: true,
		},
	}

	// Marshal to XML
	xmlData, err := xml.MarshalIndent(samlRequest, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal SAML request: %w", err)
	}

	// Base64 encode
	encodedRequest := base64.StdEncoding.EncodeToString(xmlData)

	// Build URL
	authURL, err := url.Parse(provider.SSOURL)
	if err != nil {
		return "", fmt.Errorf("invalid SSO URL: %w", err)
	}

	query := authURL.Query()
	query.Set("SAMLRequest", encodedRequest)
	if relayState != "" {
		query.Set("RelayState", relayState)
	}
	authURL.RawQuery = query.Encode()

	log.Info().
		Str("provider_id", providerID).
		Str("request_id", requestID).
		Msg("Generated SAML authentication URL")

	return authURL.String(), nil
}

// HandleResponse handles SAML response and validates assertion
func (sm *SAMLManager) HandleResponse(samlResponse, relayState string) (*auth.AuthContext, error) {
	// Decode base64 response
	responseData, err := base64.StdEncoding.DecodeString(samlResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode SAML response: %w", err)
	}

	// Parse XML
	var response SAMLResponse
	if err := xml.Unmarshal(responseData, &response); err != nil {
		return nil, fmt.Errorf("failed to parse SAML response: %w", err)
	}

	// Validate response
	if err := sm.validateResponse(&response); err != nil {
		return nil, fmt.Errorf("SAML response validation failed: %w", err)
	}

	// Extract user information
	userInfo, err := sm.extractUserInfo(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to extract user info: %w", err)
	}

	// Create or update user
	authCtx, err := sm.createOrUpdateUser(userInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to create/update user: %w", err)
	}

	log.Info().
		Str("name_id", userInfo.NameID).
		Str("email", userInfo.Email).
		Str("provider", userInfo.Provider).
		Msg("SAML authentication successful")

	return authCtx, nil
}

// validateResponse validates the SAML response
func (sm *SAMLManager) validateResponse(response *SAMLResponse) error {
	// Check status
	if response.Status.StatusCode.Value != "urn:oasis:names:tc:SAML:2.0:status:Success" {
		return fmt.Errorf("SAML response status is not success: %s", response.Status.StatusCode.Value)
	}

	// Check destination
	if response.Destination != sm.acsURL {
		return fmt.Errorf("invalid destination: expected %s, got %s", sm.acsURL, response.Destination)
	}

	// Check assertion conditions
	assertion := &response.Assertion
	now := time.Now().UTC()

	if now.Before(assertion.Conditions.NotBefore) {
		return fmt.Errorf("assertion not yet valid")
	}

	if now.After(assertion.Conditions.NotOnOrAfter) {
		return fmt.Errorf("assertion expired")
	}

	// Check audience
	if assertion.Conditions.AudienceRestriction.Audience.Value != sm.entityID {
		return fmt.Errorf("invalid audience: expected %s, got %s",
			sm.entityID, assertion.Conditions.AudienceRestriction.Audience.Value)
	}

	// TODO: Verify signature if required
	// This would require implementing XML signature verification

	return nil
}

// extractUserInfo extracts user information from SAML assertion
func (sm *SAMLManager) extractUserInfo(response *SAMLResponse) (*SAMLUserInfo, error) {
	assertion := &response.Assertion

	userInfo := &SAMLUserInfo{
		NameID:     assertion.Subject.NameID.Value,
		Attributes: make(map[string]string),
		SessionID:  assertion.AuthnStatement.SessionIndex,
		Provider:   response.Issuer.Value,
	}

	// Extract attributes
	for _, attr := range assertion.AttributeStatement.Attributes {
		if len(attr.AttributeValues) > 0 {
			value := attr.AttributeValues[0].Value

			// Map common attributes
			switch attr.Name {
			case "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress":
				userInfo.Email = value
			case "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name":
				userInfo.Name = value
			case "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname":
				userInfo.FirstName = value
			case "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname":
				userInfo.LastName = value
			case "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/groups":
				// Handle multiple group values
				for _, attrValue := range attr.AttributeValues {
					userInfo.Groups = append(userInfo.Groups, attrValue.Value)
				}
			}

			// Store all attributes
			userInfo.Attributes[attr.Name] = value
		}
	}

	// Use NameID as email if email not provided
	if userInfo.Email == "" {
		userInfo.Email = userInfo.NameID
	}

	return userInfo, nil
}

// createOrUpdateUser creates or updates a user based on SAML user info
func (sm *SAMLManager) createOrUpdateUser(userInfo *SAMLUserInfo) (*auth.AuthContext, error) {
	// Generate user ID based on provider and NameID
	userID := fmt.Sprintf("saml:%s:%s", userInfo.Provider, userInfo.NameID)

	// Check if user exists
	existingUser, err := sm.authMgr.GetUser(userID)
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
		Role:     auth.RoleUser, // Default role for SAML users
		Active:   true,
		Metadata: map[string]string{
			"provider":    userInfo.Provider,
			"name_id":     userInfo.NameID,
			"name":        userInfo.Name,
			"first_name":  userInfo.FirstName,
			"last_name":   userInfo.LastName,
			"session_id":  userInfo.SessionID,
			"auth_method": "saml",
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
		if err := sm.authMgr.UpdateUser(userData); err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
	} else {
		// Create new user using CreateUserRequest
		createReq := &auth.CreateUserRequest{
			Username: userInfo.Email,
			Email:    userInfo.Email,
			Password: "saml-sso", // Placeholder password for SSO users
			Role:     auth.RoleUser,
			Permissions: []string{
				auth.PermissionModelRead,
				auth.PermissionInferenceWrite,
			},
			Metadata: userData.Metadata,
		}

		createdUser, err := sm.authMgr.CreateUser(createReq)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}

		// Update the user ID to match our SSO format
		createdUser.ID = userID
		if err := sm.authMgr.UpdateUser(createdUser); err != nil {
			return nil, fmt.Errorf("failed to update user ID: %w", err)
		}
		userData = createdUser
	}

	// Authenticate user
	authCtx, err := sm.authMgr.AuthenticateUser(userData)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	return authCtx, nil
}

// GetProviders returns all configured SAML providers
func (sm *SAMLManager) GetProviders() map[string]*SAMLProvider {
	providers := make(map[string]*SAMLProvider)
	for id, provider := range sm.providers {
		if provider.Enabled {
			// Return sanitized version without secrets
			providers[id] = &SAMLProvider{
				ID:           provider.ID,
				Name:         provider.Name,
				EntityID:     provider.EntityID,
				SSOURL:       provider.SSOURL,
				NameIDFormat: provider.NameIDFormat,
				Enabled:      provider.Enabled,
			}
		}
	}
	return providers
}

// generateID generates a unique ID for SAML requests
func generateID() string {
	return fmt.Sprintf("_%d", time.Now().UnixNano())
}

// GetMetadata returns SAML metadata for this service provider
func (sm *SAMLManager) GetMetadata() (string, error) {
	metadata := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata" entityID="%s">
  <md:SPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <md:AssertionConsumerService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="%s" index="0"/>
    <md:SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="%s"/>
  </md:SPSSODescriptor>
</md:EntityDescriptor>`, sm.entityID, sm.acsURL, sm.sloURL)

	return metadata, nil
}

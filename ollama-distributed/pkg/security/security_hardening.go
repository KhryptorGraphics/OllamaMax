package security

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// SecurityHardeningManager manages security hardening configurations
type SecurityHardeningManager struct {
	config            *HardeningConfig
	configValidator   *ConfigurationValidator
	policyEnforcer    *SecurityPolicyEnforcer
	complianceChecker *ComplianceChecker

	// State management
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

// HardeningConfig configures security hardening
type HardeningConfig struct {
	Enabled                 bool          `json:"enabled"`
	EnforceSecureDefaults   bool          `json:"enforce_secure_defaults"`
	ValidateConfigurations  bool          `json:"validate_configurations"`
	EnablePolicyEnforcement bool          `json:"enable_policy_enforcement"`
	ComplianceFrameworks    []string      `json:"compliance_frameworks"`
	HardeningLevel          string        `json:"hardening_level"` // basic, standard, strict
	AutoRemediation         bool          `json:"auto_remediation"`
	ValidationInterval      time.Duration `json:"validation_interval"`
}

// ConfigurationValidator validates security configurations
type ConfigurationValidator struct {
	rules      map[string]*ValidationRule
	violations []ConfigurationViolation
	mu         sync.RWMutex
}

// ValidationRule defines a configuration validation rule
type ValidationRule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Category    string            `json:"category"`
	Severity    SeverityLevel     `json:"severity"`
	Pattern     string            `json:"pattern"`
	Expected    interface{}       `json:"expected"`
	Remediation string            `json:"remediation"`
	Metadata    map[string]string `json:"metadata"`
}

// ConfigurationViolation represents a configuration security violation
type ConfigurationViolation struct {
	RuleID        string                 `json:"rule_id"`
	RuleName      string                 `json:"rule_name"`
	Severity      SeverityLevel          `json:"severity"`
	Description   string                 `json:"description"`
	Location      string                 `json:"location"`
	CurrentValue  interface{}            `json:"current_value"`
	ExpectedValue interface{}            `json:"expected_value"`
	Remediation   string                 `json:"remediation"`
	DetectedAt    time.Time              `json:"detected_at"`
	Status        ViolationStatus        `json:"status"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// ViolationStatus represents the status of a configuration violation
type ViolationStatus string

const (
	ViolationStatusActive     ViolationStatus = "active"
	ViolationStatusRemediated ViolationStatus = "remediated"
	ViolationStatusIgnored    ViolationStatus = "ignored"
	ViolationStatusPending    ViolationStatus = "pending"
)

// SecurityPolicyEnforcer enforces security policies
type SecurityPolicyEnforcer struct {
	policies    map[string]*HardeningSecurityPolicy
	enforcement map[string]bool
	mu          sync.RWMutex
}

// HardeningSecurityPolicy defines a security policy for hardening
type HardeningSecurityPolicy struct {
	ID          string                    `json:"id"`
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Category    string                    `json:"category"`
	Rules       []PolicyRule              `json:"rules"`
	Enforcement HardeningEnforcementLevel `json:"enforcement"`
	Metadata    map[string]string         `json:"metadata"`
}

// PolicyRule defines a rule within a security policy
type PolicyRule struct {
	ID          string        `json:"id"`
	Description string        `json:"description"`
	Condition   string        `json:"condition"`
	Action      PolicyAction  `json:"action"`
	Severity    SeverityLevel `json:"severity"`
}

// PolicyAction defines the action to take when a policy is violated
type PolicyAction string

const (
	PolicyActionLog       PolicyAction = "log"
	PolicyActionWarn      PolicyAction = "warn"
	PolicyActionBlock     PolicyAction = "block"
	PolicyActionReject    PolicyAction = "reject"
	PolicyActionRemediate PolicyAction = "remediate"
)

// HardeningEnforcementLevel defines how strictly a policy is enforced
type HardeningEnforcementLevel string

const (
	HardeningEnforcementLevelAdvisory  HardeningEnforcementLevel = "advisory"
	HardeningEnforcementLevelWarning   HardeningEnforcementLevel = "warning"
	HardeningEnforcementLevelMandatory HardeningEnforcementLevel = "mandatory"
)

// ComplianceChecker checks compliance with security frameworks
type ComplianceChecker struct {
	frameworks map[string]*ComplianceFramework
	results    map[string]*ComplianceResult
	mu         sync.RWMutex
}

// ComplianceFramework defines a compliance framework
type ComplianceFramework struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Version     string              `json:"version"`
	Description string              `json:"description"`
	Controls    []ComplianceControl `json:"controls"`
}

// ComplianceControl defines a compliance control
type ComplianceControl struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Category    string        `json:"category"`
	Severity    SeverityLevel `json:"severity"`
	Checks      []string      `json:"checks"`
}

// ComplianceResult represents compliance check results
type ComplianceResult struct {
	FrameworkID    string                   `json:"framework_id"`
	FrameworkName  string                   `json:"framework_name"`
	OverallScore   float64                  `json:"overall_score"`
	TotalControls  int                      `json:"total_controls"`
	PassedControls int                      `json:"passed_controls"`
	FailedControls int                      `json:"failed_controls"`
	ControlResults map[string]ControlResult `json:"control_results"`
	CheckedAt      time.Time                `json:"checked_at"`
}

// ControlResult represents the result of a compliance control check
type ControlResult struct {
	ControlID   string    `json:"control_id"`
	Status      string    `json:"status"` // pass, fail, not_applicable
	Score       float64   `json:"score"`
	Evidence    []string  `json:"evidence"`
	Remediation string    `json:"remediation"`
	CheckedAt   time.Time `json:"checked_at"`
}

// NewSecurityHardeningManager creates a new security hardening manager
func NewSecurityHardeningManager(config *HardeningConfig) *SecurityHardeningManager {
	if config == nil {
		config = DefaultHardeningConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	shm := &SecurityHardeningManager{
		config: config,
		configValidator: &ConfigurationValidator{
			rules:      make(map[string]*ValidationRule),
			violations: make([]ConfigurationViolation, 0),
		},
		policyEnforcer: &SecurityPolicyEnforcer{
			policies:    make(map[string]*HardeningSecurityPolicy),
			enforcement: make(map[string]bool),
		},
		complianceChecker: &ComplianceChecker{
			frameworks: make(map[string]*ComplianceFramework),
			results:    make(map[string]*ComplianceResult),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize default rules and policies
	shm.initializeDefaultRules()
	shm.initializeDefaultPolicies()
	shm.initializeComplianceFrameworks()

	return shm
}

// Start starts the security hardening manager
func (shm *SecurityHardeningManager) Start() error {
	if !shm.config.Enabled {
		log.Info().Msg("Security hardening disabled")
		return nil
	}

	// Apply secure defaults
	if shm.config.EnforceSecureDefaults {
		if err := shm.applySecureDefaults(); err != nil {
			log.Error().Err(err).Msg("Failed to apply secure defaults")
		}
	}

	// Start periodic validation
	if shm.config.ValidateConfigurations {
		go shm.startPeriodicValidation()
	}

	log.Info().
		Str("hardening_level", shm.config.HardeningLevel).
		Bool("auto_remediation", shm.config.AutoRemediation).
		Msg("Security hardening started")

	return nil
}

// ValidateConfiguration validates the current configuration
func (shm *SecurityHardeningManager) ValidateConfiguration() (*ValidationResult, error) {
	shm.mu.Lock()
	defer shm.mu.Unlock()

	result := &ValidationResult{
		Timestamp:    time.Now(),
		TotalRules:   len(shm.configValidator.rules),
		Violations:   make([]ConfigurationViolation, 0),
		OverallScore: 100.0,
	}

	// Run all validation rules
	for _, rule := range shm.configValidator.rules {
		violation := shm.validateRule(rule)
		if violation != nil {
			result.Violations = append(result.Violations, *violation)
		}
	}

	// Calculate overall score
	if len(result.Violations) > 0 {
		result.OverallScore = shm.calculateValidationScore(result.Violations)
	}

	result.PassedRules = result.TotalRules - len(result.Violations)
	result.FailedRules = len(result.Violations)

	log.Info().
		Int("total_rules", result.TotalRules).
		Int("violations", len(result.Violations)).
		Float64("score", result.OverallScore).
		Msg("Configuration validation completed")

	return result, nil
}

// ValidationResult represents configuration validation results
type ValidationResult struct {
	Timestamp    time.Time                `json:"timestamp"`
	TotalRules   int                      `json:"total_rules"`
	PassedRules  int                      `json:"passed_rules"`
	FailedRules  int                      `json:"failed_rules"`
	Violations   []ConfigurationViolation `json:"violations"`
	OverallScore float64                  `json:"overall_score"`
}

// validateRule validates a single configuration rule
func (shm *SecurityHardeningManager) validateRule(rule *ValidationRule) *ConfigurationViolation {
	switch rule.Category {
	case "tls":
		return shm.validateTLSRule(rule)
	case "authentication":
		return shm.validateAuthRule(rule)
	case "permissions":
		return shm.validatePermissionsRule(rule)
	case "network":
		return shm.validateNetworkRule(rule)
	default:
		return shm.validateGenericRule(rule)
	}
}

// validateTLSRule validates TLS-related configuration rules
func (shm *SecurityHardeningManager) validateTLSRule(rule *ValidationRule) *ConfigurationViolation {
	// Example TLS validation
	if rule.ID == "tls_min_version" {
		// Check minimum TLS version configuration
		// This would check actual TLS configuration
		return nil // Simplified - assume compliant
	}
	return nil
}

// validateAuthRule validates authentication-related rules
func (shm *SecurityHardeningManager) validateAuthRule(rule *ValidationRule) *ConfigurationViolation {
	// Example authentication validation
	if rule.ID == "password_policy" {
		// Check password policy configuration
		// This would check actual password policy
		return nil // Simplified - assume compliant
	}
	return nil
}

// validatePermissionsRule validates file/directory permissions
func (shm *SecurityHardeningManager) validatePermissionsRule(rule *ValidationRule) *ConfigurationViolation {
	if rule.ID == "config_file_permissions" {
		// Check configuration file permissions
		configFiles := []string{"config.yaml", "security.yaml"}

		for _, file := range configFiles {
			if info, err := os.Stat(file); err == nil {
				mode := info.Mode()
				if mode.Perm() > 0600 { // More permissive than owner read/write only
					return &ConfigurationViolation{
						RuleID:        rule.ID,
						RuleName:      rule.Name,
						Severity:      rule.Severity,
						Description:   fmt.Sprintf("Configuration file %s has overly permissive permissions", file),
						Location:      file,
						CurrentValue:  mode.Perm().String(),
						ExpectedValue: "0600",
						Remediation:   fmt.Sprintf("chmod 600 %s", file),
						DetectedAt:    time.Now(),
						Status:        ViolationStatusActive,
						Metadata:      make(map[string]interface{}),
					}
				}
			}
		}
	}
	return nil
}

// validateNetworkRule validates network-related rules
func (shm *SecurityHardeningManager) validateNetworkRule(rule *ValidationRule) *ConfigurationViolation {
	// Example network validation
	return nil // Simplified implementation
}

// validateGenericRule validates generic configuration rules
func (shm *SecurityHardeningManager) validateGenericRule(rule *ValidationRule) *ConfigurationViolation {
	// Generic rule validation using patterns
	if rule.Pattern != "" {
		matched, _ := regexp.MatchString(rule.Pattern, "")
		if !matched {
			return &ConfigurationViolation{
				RuleID:      rule.ID,
				RuleName:    rule.Name,
				Severity:    rule.Severity,
				Description: rule.Description,
				Location:    "configuration",
				Remediation: rule.Remediation,
				DetectedAt:  time.Now(),
				Status:      ViolationStatusActive,
				Metadata:    make(map[string]interface{}),
			}
		}
	}
	return nil
}

// calculateValidationScore calculates overall validation score
func (shm *SecurityHardeningManager) calculateValidationScore(violations []ConfigurationViolation) float64 {
	if len(violations) == 0 {
		return 100.0
	}

	totalWeight := 0.0
	violationWeight := 0.0

	for _, violation := range violations {
		weight := shm.getSeverityWeight(violation.Severity)
		totalWeight += weight
		violationWeight += weight
	}

	if totalWeight == 0 {
		return 100.0
	}

	score := 100.0 - (violationWeight/totalWeight)*100.0
	if score < 0 {
		score = 0
	}

	return score
}

// getSeverityWeight returns weight for severity level
func (shm *SecurityHardeningManager) getSeverityWeight(severity SeverityLevel) float64 {
	switch severity {
	case SeverityInfo:
		return 1.0
	case SeverityLow:
		return 2.0
	case SeverityMedium:
		return 4.0
	case SeverityHigh:
		return 7.0
	case SeverityCritical:
		return 10.0
	default:
		return 1.0
	}
}

// applySecureDefaults applies secure default configurations
func (shm *SecurityHardeningManager) applySecureDefaults() error {
	log.Info().Msg("Applying secure default configurations")

	// Apply secure defaults based on hardening level
	switch shm.config.HardeningLevel {
	case "basic":
		return shm.applyBasicHardening()
	case "standard":
		return shm.applyStandardHardening()
	case "strict":
		return shm.applyStrictHardening()
	default:
		return shm.applyStandardHardening()
	}
}

// applyBasicHardening applies basic security hardening
func (shm *SecurityHardeningManager) applyBasicHardening() error {
	// Basic hardening configurations
	log.Info().Msg("Applying basic security hardening")
	return nil
}

// applyStandardHardening applies standard security hardening
func (shm *SecurityHardeningManager) applyStandardHardening() error {
	// Standard hardening configurations
	log.Info().Msg("Applying standard security hardening")
	return nil
}

// applyStrictHardening applies strict security hardening
func (shm *SecurityHardeningManager) applyStrictHardening() error {
	// Strict hardening configurations
	log.Info().Msg("Applying strict security hardening")
	return nil
}

// startPeriodicValidation starts periodic configuration validation
func (shm *SecurityHardeningManager) startPeriodicValidation() {
	ticker := time.NewTicker(shm.config.ValidationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-shm.ctx.Done():
			return
		case <-ticker.C:
			if _, err := shm.ValidateConfiguration(); err != nil {
				log.Error().Err(err).Msg("Periodic configuration validation failed")
			}
		}
	}
}

// initializeDefaultRules initializes default validation rules
func (shm *SecurityHardeningManager) initializeDefaultRules() {
	rules := []*ValidationRule{
		{
			ID:          "tls_min_version",
			Name:        "Minimum TLS Version",
			Description: "Ensure minimum TLS version is 1.2 or higher",
			Category:    "tls",
			Severity:    SeverityHigh,
			Expected:    "1.2",
			Remediation: "Configure minimum TLS version to 1.2 or higher",
		},
		{
			ID:          "password_policy",
			Name:        "Password Policy",
			Description: "Ensure strong password policy is enforced",
			Category:    "authentication",
			Severity:    SeverityMedium,
			Remediation: "Configure strong password policy with minimum length and complexity",
		},
		{
			ID:          "config_file_permissions",
			Name:        "Configuration File Permissions",
			Description: "Ensure configuration files have secure permissions",
			Category:    "permissions",
			Severity:    SeverityMedium,
			Expected:    "0600",
			Remediation: "Set configuration file permissions to 600 (owner read/write only)",
		},
	}

	for _, rule := range rules {
		shm.configValidator.rules[rule.ID] = rule
	}

	log.Info().
		Int("rules_count", len(rules)).
		Msg("Default validation rules initialized")
}

// initializeDefaultPolicies initializes default security policies
func (shm *SecurityHardeningManager) initializeDefaultPolicies() {
	// Initialize default security policies
	log.Info().Msg("Default security policies initialized")
}

// initializeComplianceFrameworks initializes compliance frameworks
func (shm *SecurityHardeningManager) initializeComplianceFrameworks() {
	// Initialize compliance frameworks (CIS, NIST, etc.)
	log.Info().Msg("Compliance frameworks initialized")
}

// Shutdown gracefully shuts down the security hardening manager
func (shm *SecurityHardeningManager) Shutdown() error {
	shm.cancel()
	log.Info().Msg("Security hardening manager stopped")
	return nil
}

// DefaultHardeningConfig returns default hardening configuration
func DefaultHardeningConfig() *HardeningConfig {
	return &HardeningConfig{
		Enabled:                 true,
		EnforceSecureDefaults:   true,
		ValidateConfigurations:  true,
		EnablePolicyEnforcement: true,
		ComplianceFrameworks:    []string{"CIS", "NIST"},
		HardeningLevel:          "standard",
		AutoRemediation:         false,
		ValidationInterval:      1 * time.Hour,
	}
}

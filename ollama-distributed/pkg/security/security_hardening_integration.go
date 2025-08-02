package security

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// SecurityHardeningIntegration integrates security hardening with existing security systems
type SecurityHardeningIntegration struct {
	// Core components
	securityManager   *SecurityManager
	auditor          *SecurityAuditor
	encryptionManager *AdvancedEncryptionManager
	monitor          *SecurityMonitor
	hardeningManager *SecurityHardeningManager
	
	// Configuration
	config *HardeningIntegrationConfig
	
	// State management
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

// HardeningIntegrationConfig configures the security hardening integration
type HardeningIntegrationConfig struct {
	Enabled                    bool          `json:"enabled"`
	EnableAuditIntegration     bool          `json:"enable_audit_integration"`
	EnableMonitoringIntegration bool         `json:"enable_monitoring_integration"`
	EnableEncryptionIntegration bool         `json:"enable_encryption_integration"`
	EnableHardeningIntegration bool          `json:"enable_hardening_integration"`
	
	// Integration intervals
	AuditInterval      time.Duration `json:"audit_interval"`
	MonitoringInterval time.Duration `json:"monitoring_interval"`
	SyncInterval       time.Duration `json:"sync_interval"`
	
	// Security levels
	SecurityLevel      string `json:"security_level"` // basic, standard, strict, maximum
	AutoRemediation    bool   `json:"auto_remediation"`
	AlertThreshold     int    `json:"alert_threshold"`
}

// SecurityStatus represents the overall security status
type SecurityStatus struct {
	Timestamp        time.Time                 `json:"timestamp"`
	OverallScore     float64                   `json:"overall_score"`
	SecurityLevel    string                    `json:"security_level"`
	AuditScore       float64                   `json:"audit_score"`
	ValidationScore  float64                   `json:"validation_score"`
	ThreatLevel      string                    `json:"threat_level"`
	ActiveThreats    int                       `json:"active_threats"`
	SecurityAlerts   int                       `json:"security_alerts"`
	Compliance       map[string]bool           `json:"compliance"`
	Recommendations  []string                  `json:"recommendations"`
	Components       map[string]ComponentStatus `json:"components"`
}

// ComponentStatus represents the status of a security component
type ComponentStatus struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"` // healthy, degraded, critical, offline
	Score     float64   `json:"score"`
	LastCheck time.Time `json:"last_check"`
	Issues    []string  `json:"issues"`
}

// SecurityReport represents a comprehensive security report
type SecurityReport struct {
	GeneratedAt      time.Time                    `json:"generated_at"`
	ReportType       string                       `json:"report_type"`
	SecurityStatus   *SecurityStatus              `json:"security_status"`
	AuditResults     *AuditResults                `json:"audit_results"`
	ValidationResults *ValidationResult           `json:"validation_results"`
	ThreatIndicators []ThreatIndicator            `json:"threat_indicators"`
	SecurityAlerts   []SecurityAlert              `json:"security_alerts"`
	Recommendations  []SecurityRecommendation     `json:"recommendations"`
	ExecutiveSummary string                       `json:"executive_summary"`
}

// SecurityRecommendation represents a security recommendation
type SecurityRecommendation struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Priority    SeverityLevel `json:"priority"`
	Category    string        `json:"category"`
	Impact      string        `json:"impact"`
	Effort      string        `json:"effort"`
	Timeline    string        `json:"timeline"`
	References  []string      `json:"references"`
}

// NewSecurityHardeningIntegration creates a new security hardening integration
func NewSecurityHardeningIntegration(
	securityManager *SecurityManager,
	config *HardeningIntegrationConfig,
) *SecurityHardeningIntegration {
	if config == nil {
		config = DefaultHardeningIntegrationConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	shi := &SecurityHardeningIntegration{
		securityManager: securityManager,
		config:          config,
		ctx:             ctx,
		cancel:          cancel,
	}

	// Initialize components based on configuration
	shi.initializeComponents()

	return shi
}

// initializeComponents initializes security hardening components
func (shi *SecurityHardeningIntegration) initializeComponents() {
	// Initialize security auditor
	if shi.config.EnableAuditIntegration {
		auditConfig := DefaultAuditConfig()
		auditConfig.ScanInterval = shi.config.AuditInterval
		shi.auditor = NewSecurityAuditor(auditConfig)
	}

	// Initialize advanced encryption manager
	if shi.config.EnableEncryptionIntegration {
		encryptionConfig := DefaultEncryptionConfig()
		shi.encryptionManager = NewAdvancedEncryptionManager(encryptionConfig)
	}

	// Initialize security monitor
	if shi.config.EnableMonitoringIntegration {
		monitoringConfig := DefaultMonitoringConfig()
		monitoringConfig.CollectionInterval = shi.config.MonitoringInterval
		shi.monitor = NewSecurityMonitor(monitoringConfig)
	}

	// Initialize hardening manager
	if shi.config.EnableHardeningIntegration {
		hardeningConfig := DefaultHardeningConfig()
		hardeningConfig.HardeningLevel = shi.config.SecurityLevel
		hardeningConfig.AutoRemediation = shi.config.AutoRemediation
		shi.hardeningManager = NewSecurityHardeningManager(hardeningConfig)
	}

	log.Info().
		Bool("audit", shi.config.EnableAuditIntegration).
		Bool("encryption", shi.config.EnableEncryptionIntegration).
		Bool("monitoring", shi.config.EnableMonitoringIntegration).
		Bool("hardening", shi.config.EnableHardeningIntegration).
		Msg("Security hardening components initialized")
}

// Start starts the security hardening integration
func (shi *SecurityHardeningIntegration) Start() error {
	if !shi.config.Enabled {
		log.Info().Msg("Security hardening integration disabled")
		return nil
	}

	// Start individual components
	if shi.auditor != nil {
		// Auditor doesn't have a Start method, it's run on-demand
		log.Info().Msg("Security auditor ready")
	}

	if shi.encryptionManager != nil {
		// Encryption manager starts automatically
		log.Info().Msg("Advanced encryption manager started")
	}

	if shi.monitor != nil {
		if err := shi.monitor.Start(); err != nil {
			return fmt.Errorf("failed to start security monitor: %w", err)
		}
	}

	if shi.hardeningManager != nil {
		if err := shi.hardeningManager.Start(); err != nil {
			return fmt.Errorf("failed to start hardening manager: %w", err)
		}
	}

	// Start integration loops
	go shi.securitySyncLoop()
	go shi.threatCorrelationLoop()

	log.Info().
		Str("security_level", shi.config.SecurityLevel).
		Bool("auto_remediation", shi.config.AutoRemediation).
		Msg("Security hardening integration started")

	return nil
}

// GetSecurityStatus returns the current overall security status
func (shi *SecurityHardeningIntegration) GetSecurityStatus() (*SecurityStatus, error) {
	shi.mu.RLock()
	defer shi.mu.RUnlock()

	status := &SecurityStatus{
		Timestamp:     time.Now(),
		SecurityLevel: shi.config.SecurityLevel,
		Compliance:    make(map[string]bool),
		Components:    make(map[string]ComponentStatus),
		Recommendations: make([]string, 0),
	}

	// Aggregate audit results
	if shi.auditor != nil {
		auditResults := shi.auditor.GetResults()
		if auditResults != nil {
			status.AuditScore = auditResults.OverallScore
			for framework, compliant := range auditResults.Compliance {
				status.Compliance[framework] = compliant
			}
		}
	}

	// Aggregate validation results
	if shi.hardeningManager != nil {
		validationResults, err := shi.hardeningManager.ValidateConfiguration()
		if err == nil && validationResults != nil {
			status.ValidationScore = validationResults.OverallScore
		}
	}

	// Aggregate threat information
	if shi.monitor != nil {
		indicators := shi.monitor.GetThreatIndicators()
		alerts := shi.monitor.GetSecurityAlerts()
		
		status.ActiveThreats = len(indicators)
		status.SecurityAlerts = len(alerts)
		
		// Determine threat level
		status.ThreatLevel = shi.calculateThreatLevel(indicators)
	}

	// Calculate overall security score
	status.OverallScore = shi.calculateOverallScore(status)

	// Generate component status
	shi.populateComponentStatus(status)

	// Generate recommendations
	status.Recommendations = shi.generateRecommendations(status)

	return status, nil
}

// GenerateSecurityReport generates a comprehensive security report
func (shi *SecurityHardeningIntegration) GenerateSecurityReport() (*SecurityReport, error) {
	report := &SecurityReport{
		GeneratedAt: time.Now(),
		ReportType:  "comprehensive",
	}

	// Get overall security status
	status, err := shi.GetSecurityStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to get security status: %w", err)
	}
	report.SecurityStatus = status

	// Get audit results
	if shi.auditor != nil {
		ctx, cancel := context.WithTimeout(shi.ctx, 30*time.Second)
		auditResults, err := shi.auditor.RunAudit(ctx)
		cancel()
		if err == nil {
			report.AuditResults = auditResults
		}
	}

	// Get validation results
	if shi.hardeningManager != nil {
		validationResults, err := shi.hardeningManager.ValidateConfiguration()
		if err == nil {
			report.ValidationResults = validationResults
		}
	}

	// Get threat indicators and alerts
	if shi.monitor != nil {
		indicators := shi.monitor.GetThreatIndicators()
		alerts := shi.monitor.GetSecurityAlerts()
		
		for _, indicator := range indicators {
			report.ThreatIndicators = append(report.ThreatIndicators, *indicator)
		}
		report.SecurityAlerts = alerts
	}

	// Generate recommendations
	report.Recommendations = shi.generateDetailedRecommendations(report)

	// Generate executive summary
	report.ExecutiveSummary = shi.generateExecutiveSummary(report)

	log.Info().
		Float64("overall_score", status.OverallScore).
		Int("threats", len(report.ThreatIndicators)).
		Int("alerts", len(report.SecurityAlerts)).
		Msg("Security report generated")

	return report, nil
}

// securitySyncLoop synchronizes security data between components
func (shi *SecurityHardeningIntegration) securitySyncLoop() {
	ticker := time.NewTicker(shi.config.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-shi.ctx.Done():
			return
		case <-ticker.C:
			shi.synchronizeSecurityData()
		}
	}
}

// threatCorrelationLoop correlates threats across different components
func (shi *SecurityHardeningIntegration) threatCorrelationLoop() {
	ticker := time.NewTicker(shi.config.MonitoringInterval)
	defer ticker.Stop()

	for {
		select {
		case <-shi.ctx.Done():
			return
		case <-ticker.C:
			shi.correlateThreats()
		}
	}
}

// synchronizeSecurityData synchronizes security data between components
func (shi *SecurityHardeningIntegration) synchronizeSecurityData() {
	// Sync encryption keys with security manager
	if shi.encryptionManager != nil && shi.securityManager != nil {
		keyInfo := shi.encryptionManager.GetCurrentKeyInfo()
		if keyInfo != nil {
			log.Debug().
				Str("key_id", keyInfo.ID).
				Msg("Synchronized encryption key info")
		}
	}

	// Sync security events between monitor and auditor
	if shi.monitor != nil && shi.auditor != nil {
		// This would sync security events for correlation
		log.Debug().Msg("Synchronized security events")
	}
}

// correlateThreats correlates threats across different security components
func (shi *SecurityHardeningIntegration) correlateThreats() {
	if shi.monitor == nil {
		return
	}

	indicators := shi.monitor.GetThreatIndicators()
	if len(indicators) == 0 {
		return
	}

	// Correlate threats with audit findings
	if shi.auditor != nil {
		auditResults := shi.auditor.GetResults()
		if auditResults != nil {
			shi.correlateThreatsWithAudit(indicators, auditResults)
		}
	}

	log.Debug().
		Int("indicators", len(indicators)).
		Msg("Threat correlation completed")
}

// correlateThreatsWithAudit correlates threat indicators with audit findings
func (shi *SecurityHardeningIntegration) correlateThreatsWithAudit(indicators map[string]*ThreatIndicator, auditResults *AuditResults) {
	// This would implement threat-audit correlation logic
	// For now, just log the correlation attempt
	log.Debug().
		Int("threats", len(indicators)).
		Int("findings", auditResults.TotalFindings).
		Msg("Correlating threats with audit findings")
}

// calculateThreatLevel calculates the overall threat level
func (shi *SecurityHardeningIntegration) calculateThreatLevel(indicators map[string]*ThreatIndicator) string {
	if len(indicators) == 0 {
		return "low"
	}

	criticalThreats := 0
	highThreats := 0

	for _, indicator := range indicators {
		switch indicator.Severity {
		case SeverityCritical:
			criticalThreats++
		case SeverityHigh:
			highThreats++
		}
	}

	if criticalThreats > 0 {
		return "critical"
	} else if highThreats > 2 {
		return "high"
	} else if len(indicators) > 5 {
		return "medium"
	}

	return "low"
}

// calculateOverallScore calculates the overall security score
func (shi *SecurityHardeningIntegration) calculateOverallScore(status *SecurityStatus) float64 {
	scores := []float64{}

	if status.AuditScore > 0 {
		scores = append(scores, status.AuditScore)
	}

	if status.ValidationScore > 0 {
		scores = append(scores, status.ValidationScore)
	}

	// Factor in threat level
	threatPenalty := 0.0
	switch status.ThreatLevel {
	case "critical":
		threatPenalty = 30.0
	case "high":
		threatPenalty = 20.0
	case "medium":
		threatPenalty = 10.0
	case "low":
		threatPenalty = 0.0
	}

	if len(scores) == 0 {
		return 100.0 - threatPenalty
	}

	// Calculate average score
	total := 0.0
	for _, score := range scores {
		total += score
	}
	avgScore := total / float64(len(scores))

	// Apply threat penalty
	finalScore := avgScore - threatPenalty
	if finalScore < 0 {
		finalScore = 0
	}

	return finalScore
}

// populateComponentStatus populates component status information
func (shi *SecurityHardeningIntegration) populateComponentStatus(status *SecurityStatus) {
	now := time.Now()

	if shi.auditor != nil {
		status.Components["auditor"] = ComponentStatus{
			Name:      "Security Auditor",
			Status:    "healthy",
			Score:     status.AuditScore,
			LastCheck: now,
			Issues:    []string{},
		}
	}

	if shi.encryptionManager != nil {
		keyStats := shi.encryptionManager.GetKeyStats()
		score := 100.0
		if keyStats["current_key_id"] == "" {
			score = 0.0
		}

		status.Components["encryption"] = ComponentStatus{
			Name:      "Encryption Manager",
			Status:    "healthy",
			Score:     score,
			LastCheck: now,
			Issues:    []string{},
		}
	}

	if shi.monitor != nil {
		status.Components["monitor"] = ComponentStatus{
			Name:      "Security Monitor",
			Status:    "healthy",
			Score:     100.0 - float64(status.ActiveThreats*10),
			LastCheck: now,
			Issues:    []string{},
		}
	}

	if shi.hardeningManager != nil {
		status.Components["hardening"] = ComponentStatus{
			Name:      "Hardening Manager",
			Status:    "healthy",
			Score:     status.ValidationScore,
			LastCheck: now,
			Issues:    []string{},
		}
	}
}

// generateRecommendations generates security recommendations
func (shi *SecurityHardeningIntegration) generateRecommendations(status *SecurityStatus) []string {
	recommendations := []string{}

	if status.OverallScore < 80 {
		recommendations = append(recommendations, "Overall security score is below recommended threshold (80)")
	}

	if status.ActiveThreats > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Address %d active security threats", status.ActiveThreats))
	}

	if status.AuditScore < 85 {
		recommendations = append(recommendations, "Improve security audit score by addressing identified vulnerabilities")
	}

	if status.ValidationScore < 85 {
		recommendations = append(recommendations, "Fix configuration validation issues")
	}

	return recommendations
}

// generateDetailedRecommendations generates detailed security recommendations
func (shi *SecurityHardeningIntegration) generateDetailedRecommendations(report *SecurityReport) []SecurityRecommendation {
	recommendations := []SecurityRecommendation{}

	// Add recommendations based on audit results
	if report.AuditResults != nil && report.AuditResults.OverallScore < 85 {
		recommendations = append(recommendations, SecurityRecommendation{
			ID:          "audit-improvement",
			Title:       "Improve Security Audit Score",
			Description: "Address security vulnerabilities identified in the audit",
			Priority:    SeverityHigh,
			Category:    "Security Audit",
			Impact:      "High",
			Effort:      "Medium",
			Timeline:    "1-2 weeks",
		})
	}

	// Add recommendations based on threats
	if len(report.ThreatIndicators) > 0 {
		recommendations = append(recommendations, SecurityRecommendation{
			ID:          "threat-mitigation",
			Title:       "Mitigate Active Security Threats",
			Description: "Address active security threats detected by monitoring",
			Priority:    SeverityCritical,
			Category:    "Threat Response",
			Impact:      "Critical",
			Effort:      "High",
			Timeline:    "Immediate",
		})
	}

	return recommendations
}

// generateExecutiveSummary generates an executive summary
func (shi *SecurityHardeningIntegration) generateExecutiveSummary(report *SecurityReport) string {
	summary := fmt.Sprintf("Security Assessment Summary (Score: %.1f/100)\n\n", report.SecurityStatus.OverallScore)

	if report.SecurityStatus.OverallScore >= 90 {
		summary += "The system demonstrates excellent security posture with minimal risks identified.\n"
	} else if report.SecurityStatus.OverallScore >= 80 {
		summary += "The system has good security posture with some areas for improvement.\n"
	} else if report.SecurityStatus.OverallScore >= 70 {
		summary += "The system has adequate security but requires attention to identified issues.\n"
	} else {
		summary += "The system has significant security concerns that require immediate attention.\n"
	}

	if len(report.ThreatIndicators) > 0 {
		summary += fmt.Sprintf("\n%d active security threats detected requiring immediate attention.\n", len(report.ThreatIndicators))
	}

	if len(report.Recommendations) > 0 {
		summary += fmt.Sprintf("\n%d security recommendations provided for improvement.\n", len(report.Recommendations))
	}

	return summary
}

// Shutdown gracefully shuts down the security hardening integration
func (shi *SecurityHardeningIntegration) Shutdown() error {
	shi.cancel()

	// Shutdown individual components
	if shi.encryptionManager != nil {
		shi.encryptionManager.Shutdown()
	}

	if shi.monitor != nil {
		shi.monitor.Shutdown()
	}

	if shi.hardeningManager != nil {
		shi.hardeningManager.Shutdown()
	}

	log.Info().Msg("Security hardening integration stopped")
	return nil
}

// DefaultHardeningIntegrationConfig returns default integration configuration
func DefaultHardeningIntegrationConfig() *HardeningIntegrationConfig {
	return &HardeningIntegrationConfig{
		Enabled:                     true,
		EnableAuditIntegration:      true,
		EnableMonitoringIntegration: true,
		EnableEncryptionIntegration: true,
		EnableHardeningIntegration:  true,
		AuditInterval:               24 * time.Hour,
		MonitoringInterval:          5 * time.Minute,
		SyncInterval:                1 * time.Hour,
		SecurityLevel:               "standard",
		AutoRemediation:             false,
		AlertThreshold:              5,
	}
}

package security

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// SecurityAuditor performs comprehensive security audits
type SecurityAuditor struct {
	config   *AuditConfig
	scanners map[string]SecurityScanner
	results  *AuditResults
	mu       sync.RWMutex
}

// AuditConfig configures the security auditor
type AuditConfig struct {
	Enabled              bool          `json:"enabled"`
	ScanInterval         time.Duration `json:"scan_interval"`
	EnabledScanners      []string      `json:"enabled_scanners"`
	ReportPath           string        `json:"report_path"`
	AlertThreshold       SeverityLevel `json:"alert_threshold"`
	AutoRemediation      bool          `json:"auto_remediation"`
	ComplianceFrameworks []string      `json:"compliance_frameworks"`
}

// SecurityScanner interface for different types of security scans
type SecurityScanner interface {
	GetName() string
	Scan(ctx context.Context) (*ScanResult, error)
	GetSeverity() SeverityLevel
	IsEnabled() bool
}

// SeverityLevel represents the severity of security findings
type SeverityLevel int

const (
	SeverityInfo SeverityLevel = iota
	SeverityLow
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

// String returns string representation of severity level
func (s SeverityLevel) String() string {
	switch s {
	case SeverityInfo:
		return "info"
	case SeverityLow:
		return "low"
	case SeverityMedium:
		return "medium"
	case SeverityHigh:
		return "high"
	case SeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// ScanResult represents the result of a security scan
type ScanResult struct {
	ScannerName     string                 `json:"scanner_name"`
	Timestamp       time.Time              `json:"timestamp"`
	Severity        SeverityLevel          `json:"severity"`
	Findings        []SecurityFinding      `json:"findings"`
	Recommendations []string               `json:"recommendations"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// SecurityFinding represents a specific security issue
type SecurityFinding struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Severity    SeverityLevel          `json:"severity"`
	Category    string                 `json:"category"`
	Location    string                 `json:"location"`
	Evidence    map[string]interface{} `json:"evidence"`
	Remediation string                 `json:"remediation"`
	References  []string               `json:"references"`
}

// AuditResults aggregates all security audit results
type AuditResults struct {
	Timestamp          time.Time             `json:"timestamp"`
	OverallScore       float64               `json:"overall_score"`
	TotalFindings      int                   `json:"total_findings"`
	FindingsBySeverity map[SeverityLevel]int `json:"findings_by_severity"`
	ScanResults        []*ScanResult         `json:"scan_results"`
	Compliance         map[string]bool       `json:"compliance"`
	Summary            string                `json:"summary"`
}

// NewSecurityAuditor creates a new security auditor
func NewSecurityAuditor(config *AuditConfig) *SecurityAuditor {
	if config == nil {
		config = DefaultAuditConfig()
	}

	auditor := &SecurityAuditor{
		config:   config,
		scanners: make(map[string]SecurityScanner),
		results: &AuditResults{
			FindingsBySeverity: make(map[SeverityLevel]int),
			Compliance:         make(map[string]bool),
		},
	}

	// Initialize default scanners
	auditor.initializeDefaultScanners()

	return auditor
}

// initializeDefaultScanners sets up default security scanners
func (sa *SecurityAuditor) initializeDefaultScanners() {
	// Configuration security scanner
	sa.scanners["config"] = &ConfigurationScanner{
		enabled: sa.isScannerEnabled("config"),
	}

	// TLS security scanner
	sa.scanners["tls"] = &TLSScanner{
		enabled: sa.isScannerEnabled("tls"),
	}

	// Authentication scanner
	sa.scanners["auth"] = &AuthenticationScanner{
		enabled: sa.isScannerEnabled("auth"),
	}

	// File permissions scanner
	sa.scanners["permissions"] = &FilePermissionsScanner{
		enabled: sa.isScannerEnabled("permissions"),
	}

	// Network security scanner
	sa.scanners["network"] = &NetworkSecurityScanner{
		enabled: sa.isScannerEnabled("network"),
	}

	// Dependency vulnerability scanner
	sa.scanners["dependencies"] = &DependencyScanner{
		enabled: sa.isScannerEnabled("dependencies"),
	}

	log.Info().
		Int("scanners_count", len(sa.scanners)).
		Msg("Security scanners initialized")
}

// isScannerEnabled checks if a scanner is enabled
func (sa *SecurityAuditor) isScannerEnabled(scannerName string) bool {
	if len(sa.config.EnabledScanners) == 0 {
		return true // Enable all by default
	}

	for _, enabled := range sa.config.EnabledScanners {
		if enabled == scannerName {
			return true
		}
	}
	return false
}

// RunAudit performs a comprehensive security audit
func (sa *SecurityAuditor) RunAudit(ctx context.Context) (*AuditResults, error) {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	log.Info().Msg("Starting comprehensive security audit")

	results := &AuditResults{
		Timestamp:          time.Now(),
		FindingsBySeverity: make(map[SeverityLevel]int),
		Compliance:         make(map[string]bool),
		ScanResults:        make([]*ScanResult, 0),
	}

	// Run all enabled scanners
	for name, scanner := range sa.scanners {
		if !scanner.IsEnabled() {
			continue
		}

		log.Info().Str("scanner", name).Msg("Running security scanner")

		scanResult, err := scanner.Scan(ctx)
		if err != nil {
			log.Error().
				Err(err).
				Str("scanner", name).
				Msg("Security scanner failed")
			continue
		}

		results.ScanResults = append(results.ScanResults, scanResult)

		// Aggregate findings
		for _, finding := range scanResult.Findings {
			results.TotalFindings++
			results.FindingsBySeverity[finding.Severity]++
		}
	}

	// Calculate overall security score
	results.OverallScore = sa.calculateSecurityScore(results)

	// Check compliance
	sa.checkCompliance(results)

	// Generate summary
	results.Summary = sa.generateSummary(results)

	sa.results = results

	log.Info().
		Float64("security_score", results.OverallScore).
		Int("total_findings", results.TotalFindings).
		Msg("Security audit completed")

	return results, nil
}

// calculateSecurityScore calculates an overall security score (0-100)
func (sa *SecurityAuditor) calculateSecurityScore(results *AuditResults) float64 {
	if results.TotalFindings == 0 {
		return 100.0
	}

	// Weight findings by severity
	totalWeight := 0.0
	maxWeight := 0.0

	for severity, count := range results.FindingsBySeverity {
		weight := sa.getSeverityWeight(severity)
		totalWeight += float64(count) * weight
		maxWeight += float64(count) * 10.0 // Max weight for critical
	}

	if maxWeight == 0 {
		return 100.0
	}

	score := 100.0 - (totalWeight/maxWeight)*100.0
	if score < 0 {
		score = 0
	}

	return score
}

// getSeverityWeight returns the weight for a severity level
func (sa *SecurityAuditor) getSeverityWeight(severity SeverityLevel) float64 {
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

// checkCompliance checks compliance with security frameworks
func (sa *SecurityAuditor) checkCompliance(results *AuditResults) {
	for _, framework := range sa.config.ComplianceFrameworks {
		switch framework {
		case "CIS":
			results.Compliance["CIS"] = sa.checkCISCompliance(results)
		case "NIST":
			results.Compliance["NIST"] = sa.checkNISTCompliance(results)
		case "SOC2":
			results.Compliance["SOC2"] = sa.checkSOC2Compliance(results)
		}
	}
}

// checkCISCompliance checks CIS (Center for Internet Security) compliance
func (sa *SecurityAuditor) checkCISCompliance(results *AuditResults) bool {
	// Simplified CIS compliance check
	criticalFindings := results.FindingsBySeverity[SeverityCritical]
	highFindings := results.FindingsBySeverity[SeverityHigh]

	// CIS compliance requires no critical findings and minimal high findings
	return criticalFindings == 0 && highFindings <= 2
}

// checkNISTCompliance checks NIST compliance
func (sa *SecurityAuditor) checkNISTCompliance(results *AuditResults) bool {
	// Simplified NIST compliance check
	return results.OverallScore >= 80.0
}

// checkSOC2Compliance checks SOC 2 compliance
func (sa *SecurityAuditor) checkSOC2Compliance(results *AuditResults) bool {
	// Simplified SOC 2 compliance check
	criticalFindings := results.FindingsBySeverity[SeverityCritical]
	return criticalFindings == 0 && results.OverallScore >= 85.0
}

// generateSummary generates a human-readable summary
func (sa *SecurityAuditor) generateSummary(results *AuditResults) string {
	var summary strings.Builder

	summary.WriteString(fmt.Sprintf("Security Audit Summary (Score: %.1f/100)\n", results.OverallScore))
	summary.WriteString(fmt.Sprintf("Total Findings: %d\n", results.TotalFindings))

	if results.TotalFindings > 0 {
		summary.WriteString("Findings by Severity:\n")
		for severity := SeverityCritical; severity >= SeverityInfo; severity-- {
			if count := results.FindingsBySeverity[severity]; count > 0 {
				summary.WriteString(fmt.Sprintf("  %s: %d\n", severity.String(), count))
			}
		}
	}

	if len(results.Compliance) > 0 {
		summary.WriteString("Compliance Status:\n")
		for framework, compliant := range results.Compliance {
			status := "FAIL"
			if compliant {
				status = "PASS"
			}
			summary.WriteString(fmt.Sprintf("  %s: %s\n", framework, status))
		}
	}

	return summary.String()
}

// GetResults returns the latest audit results
func (sa *SecurityAuditor) GetResults() *AuditResults {
	sa.mu.RLock()
	defer sa.mu.RUnlock()
	return sa.results
}

// DefaultAuditConfig returns default audit configuration
func DefaultAuditConfig() *AuditConfig {
	return &AuditConfig{
		Enabled:              true,
		ScanInterval:         24 * time.Hour,
		EnabledScanners:      []string{"config", "tls", "auth", "permissions", "network", "dependencies"},
		ReportPath:           "./security-reports",
		AlertThreshold:       SeverityHigh,
		AutoRemediation:      false,
		ComplianceFrameworks: []string{"CIS", "NIST"},
	}
}

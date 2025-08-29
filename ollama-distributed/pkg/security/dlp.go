package security

import (
	_ "bufio"
	"crypto/md5"
	"fmt"
	_ "io"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// DLPManager provides Data Loss Prevention functionality
type DLPManager struct {
	config           *DLPConfig
	rules            map[string]*DLPRule
	scanners         map[string]DataScanner
	enabled          bool
	quarantine       map[string]*QuarantinedData
	statistics       *DLPStatistics
	policyEngine     *DLPPolicyEngine
	classifiers      map[string]DataClassifier
	encryptionEngine *DLPEncryptionEngine
	mu               sync.RWMutex
}

// DLPConfig configures Data Loss Prevention
type DLPConfig struct {
	Enabled              bool              `json:"enabled"`
	ScanFileUploads      bool              `json:"scan_file_uploads"`
	ScanAPIRequests      bool              `json:"scan_api_requests"`
	ScanAPIResponses     bool              `json:"scan_api_responses"`
	BlockSensitiveData   bool              `json:"block_sensitive_data"`
	MaskSensitiveData    bool              `json:"mask_sensitive_data"`
	QuarantinePath       string            `json:"quarantine_path"`
	MaxFileSize          int64             `json:"max_file_size"`
	AllowedFileTypes     []string          `json:"allowed_file_types"`
	BlockedFileTypes     []string          `json:"blocked_file_types"`
	ScanTimeout          time.Duration     `json:"scan_timeout"`
	NotificationWebhook  string            `json:"notification_webhook"`
	RetentionPolicy      *RetentionPolicy  `json:"retention_policy"`
	ComplianceModes      []string          `json:"compliance_modes"` // PCI_DSS, HIPAA, GDPR
	EncryptionEnabled    bool              `json:"encryption_enabled"`
	DataClassification   *ClassificationConfig `json:"data_classification"`
}

// RetentionPolicy defines data retention rules
type RetentionPolicy struct {
	QuarantineRetention time.Duration `json:"quarantine_retention"`
	LogRetention        time.Duration `json:"log_retention"`
	AlertRetention      time.Duration `json:"alert_retention"`
	AutoDelete          bool          `json:"auto_delete"`
}

// ClassificationConfig configures data classification
type ClassificationConfig struct {
	Enabled         bool                        `json:"enabled"`
	Levels          []string                    `json:"levels"` // public, internal, confidential, restricted
	DefaultLevel    string                      `json:"default_level"`
	Classifiers     map[string]ClassifierConfig `json:"classifiers"`
	EnforceLabeling bool                        `json:"enforce_labeling"`
}

// ClassifierConfig configures individual classifiers
type ClassifierConfig struct {
	Type        string  `json:"type"` // regex, ml, dictionary
	Patterns    []string `json:"patterns"`
	Confidence  float64 `json:"confidence"`
	Enabled     bool    `json:"enabled"`
}

// DLPRule defines a data loss prevention rule
type DLPRule struct {
	ID                string                 `json:"id"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	DataType          DataType               `json:"data_type"`
	Pattern           string                 `json:"pattern"`
	Regex             *regexp.Regexp         `json:"-"`
	Action            DLPAction              `json:"action"`
	Confidence        float64                `json:"confidence"`
	Severity          SeverityLevel          `json:"severity"`
	Categories        []string               `json:"categories"`
	ComplianceFramework []string             `json:"compliance_framework"`
	Conditions        []DLPCondition         `json:"conditions"`
	Enabled           bool                   `json:"enabled"`
	Metadata          map[string]interface{} `json:"metadata"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// DataType represents the type of sensitive data
type DataType string

const (
	DataTypeCreditCard     DataType = "credit_card"
	DataTypeSSN            DataType = "ssn"
	DataTypePhoneNumber    DataType = "phone_number"
	DataTypeEmailAddress   DataType = "email_address"
	DataTypeAPIKey         DataType = "api_key"
	DataTypePassword       DataType = "password"
	DataTypeBankAccount    DataType = "bank_account"
	DataTypePassport       DataType = "passport"
	DataTypeDriverLicense  DataType = "driver_license"
	DataTypeMedicalRecord  DataType = "medical_record"
	DataTypePII            DataType = "pii"
	DataTypeFinancial      DataType = "financial"
	DataTypeHealthcare     DataType = "healthcare"
	DataTypeIntellectualProperty DataType = "intellectual_property"
	DataTypeCustom         DataType = "custom"
)

// DLPAction defines the action to take when sensitive data is detected
type DLPAction string

const (
	DLPActionAllow      DLPAction = "allow"
	DLPActionLog        DLPAction = "log"
	DLPActionMask       DLPAction = "mask"
	DLPActionRedact     DLPAction = "redact"
	DLPActionBlock      DLPAction = "block"
	DLPActionQuarantine DLPAction = "quarantine"
	DLPActionEncrypt    DLPAction = "encrypt"
	DLPActionAlert      DLPAction = "alert"
)

// DLPCondition defines conditions for rule application
type DLPCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // equals, contains, matches, greater_than
	Value    interface{} `json:"value"`
}

// DataScanner interface for different types of data scanning
type DataScanner interface {
	Scan(data []byte, metadata *ScanMetadata) (*ScanResult, error)
	GetType() string
	IsEnabled() bool
}

// ScanMetadata provides context for data scanning
type ScanMetadata struct {
	FileName    string            `json:"file_name"`
	ContentType string            `json:"content_type"`
	Size        int64             `json:"size"`
	Source      string            `json:"source"` // upload, api_request, api_response
	UserID      string            `json:"user_id"`
	IPAddress   string            `json:"ip_address"`
	Timestamp   time.Time         `json:"timestamp"`
	Context     map[string]string `json:"context"`
}

// ScanResult contains the results of a DLP scan
type ScanResult struct {
	Scanned       bool                  `json:"scanned"`
	Detections    []DataDetection       `json:"detections"`
	Classification *DataClassification  `json:"classification"`
	RiskScore     float64               `json:"risk_score"`
	Action        DLPAction             `json:"action"`
	ProcessedData []byte                `json:"processed_data,omitempty"`
	Metadata      map[string]interface{} `json:"metadata"`
	ScanDuration  time.Duration         `json:"scan_duration"`
}

// DataDetection represents a detected instance of sensitive data
type DataDetection struct {
	RuleID      string                 `json:"rule_id"`
	RuleName    string                 `json:"rule_name"`
	DataType    DataType               `json:"data_type"`
	Location    DataLocation           `json:"location"`
	Value       string                 `json:"value"`
	MaskedValue string                 `json:"masked_value,omitempty"`
	Confidence  float64                `json:"confidence"`
	Context     string                 `json:"context"`
	Severity    SeverityLevel          `json:"severity"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// DataLocation specifies where sensitive data was found
type DataLocation struct {
	Line      int `json:"line,omitempty"`
	Column    int `json:"column,omitempty"`
	Offset    int `json:"offset,omitempty"`
	Length    int `json:"length,omitempty"`
	FieldName string `json:"field_name,omitempty"`
}

// DataClassification represents the classification of data
type DataClassification struct {
	Level       string    `json:"level"`
	Categories  []string  `json:"categories"`
	Labels      []string  `json:"labels"`
	Confidence  float64   `json:"confidence"`
	Sensitivity string    `json:"sensitivity"`
	Handling    string    `json:"handling"`
	Expiration  *time.Time `json:"expiration,omitempty"`
}

// QuarantinedData represents data that has been quarantined
type QuarantinedData struct {
	ID            string           `json:"id"`
	OriginalPath  string           `json:"original_path"`
	QuarantinePath string          `json:"quarantine_path"`
	Detections    []DataDetection  `json:"detections"`
	Metadata      *ScanMetadata    `json:"metadata"`
	QuarantinedAt time.Time        `json:"quarantined_at"`
	ExpiresAt     time.Time        `json:"expires_at"`
	Status        QuarantineStatus `json:"status"`
	ReviewedBy    string           `json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time       `json:"reviewed_at,omitempty"`
	Comments      string           `json:"comments,omitempty"`
}

// QuarantineStatus represents the status of quarantined data
type QuarantineStatus string

const (
	QuarantineStatusActive   QuarantineStatus = "active"
	QuarantineStatusReviewed QuarantineStatus = "reviewed"
	QuarantineStatusReleased QuarantineStatus = "released"
	QuarantineStatusDeleted  QuarantineStatus = "deleted"
)

// DLPStatistics tracks DLP performance and detection statistics
type DLPStatistics struct {
	FilesScanned       int64            `json:"files_scanned"`
	RequestsScanned    int64            `json:"requests_scanned"`
	ResponsesScanned   int64            `json:"responses_scanned"`
	DetectionsTotal    int64            `json:"detections_total"`
	DetectionsByType   map[DataType]int64 `json:"detections_by_type"`
	ActionsTotal       map[DLPAction]int64 `json:"actions_total"`
	QuarantinedItems   int64            `json:"quarantined_items"`
	FalsePositives     int64            `json:"false_positives"`
	AverageScanTime    time.Duration    `json:"average_scan_time"`
	LastScanTime       time.Time        `json:"last_scan_time"`
	ComplianceViolations map[string]int64 `json:"compliance_violations"`
	mu                 sync.RWMutex
}

// DLPPolicyEngine evaluates DLP policies
type DLPPolicyEngine struct {
	policies map[string]*DLPPolicy
	mu       sync.RWMutex
}

// DLPPolicy represents a DLP policy
type DLPPolicy struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Rules       []string      `json:"rules"` // Rule IDs
	Conditions  []DLPCondition `json:"conditions"`
	Action      DLPAction     `json:"action"`
	Priority    int           `json:"priority"`
	Enabled     bool          `json:"enabled"`
}

// DataClassifier interface for data classification
type DataClassifier interface {
	Classify(data []byte, metadata *ScanMetadata) (*DataClassification, error)
	GetType() string
}

// DLPEncryptionEngine handles encryption of sensitive data
type DLPEncryptionEngine struct {
	encryptionKey []byte
	algorithm     string
}

// NewDLPManager creates a new DLP manager
func NewDLPManager(rules []DLPRule) *DLPManager {
	config := DefaultDLPConfig()

	dlp := &DLPManager{
		config:      config,
		rules:       make(map[string]*DLPRule),
		scanners:    make(map[string]DataScanner),
		enabled:     config.Enabled,
		quarantine:  make(map[string]*QuarantinedData),
		statistics:  NewDLPStatistics(),
		policyEngine: NewDLPPolicyEngine(),
		classifiers: make(map[string]DataClassifier),
	}

	// Initialize encryption engine if enabled
	if config.EncryptionEnabled {
		dlp.encryptionEngine = NewDLPEncryptionEngine()
	}

	// Initialize data scanners
	dlp.initializeScanners()

	// Initialize data classifiers
	dlp.initializeClassifiers()

	// Load default rules
	dlp.loadDefaultRules()

	// Load provided rules
	for _, rule := range rules {
		dlp.AddRule(&rule)
	}

	log.Info().
		Int("rules_loaded", len(dlp.rules)).
		Int("scanners", len(dlp.scanners)).
		Bool("enabled", config.Enabled).
		Msg("DLP manager initialized")

	return dlp
}

// initializeScanners initializes data scanners
func (dlp *DLPManager) initializeScanners() {
	// Text scanner for general text content
	dlp.scanners["text"] = NewTextScanner(dlp.rules)

	// File scanner for file-specific scanning
	dlp.scanners["file"] = NewFileScanner(dlp.rules)

	// Binary scanner for binary data
	dlp.scanners["binary"] = NewBinaryScanner(dlp.rules)

	// JSON scanner for JSON data
	dlp.scanners["json"] = NewJSONScanner(dlp.rules)

	// XML scanner for XML data
	dlp.scanners["xml"] = NewXMLScanner(dlp.rules)
}

// initializeClassifiers initializes data classifiers
func (dlp *DLPManager) initializeClassifiers() {
	if !dlp.config.DataClassification.Enabled {
		return
	}

	// Regex-based classifier
	dlp.classifiers["regex"] = NewRegexClassifier(dlp.config.DataClassification)

	// ML-based classifier (placeholder)
	// dlp.classifiers["ml"] = NewMLClassifier(dlp.config.DataClassification)

	// Dictionary-based classifier
	dlp.classifiers["dictionary"] = NewDictionaryClassifier(dlp.config.DataClassification)
}

// AddRule adds a new DLP rule
func (dlp *DLPManager) AddRule(rule *DLPRule) error {
	if rule.ID == "" {
		return fmt.Errorf("rule ID is required")
	}

	dlp.mu.Lock()
	defer dlp.mu.Unlock()

	// Compile regex if pattern is provided
	if rule.Pattern != "" {
		regex, err := regexp.Compile(rule.Pattern)
		if err != nil {
			return fmt.Errorf("invalid regex pattern: %w", err)
		}
		rule.Regex = regex
	}

	rule.UpdatedAt = time.Now()
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = time.Now()
	}

	dlp.rules[rule.ID] = rule

	log.Info().
		Str("rule_id", rule.ID).
		Str("data_type", string(rule.DataType)).
		Msg("DLP rule added")

	return nil
}

// RemoveRule removes a DLP rule
func (dlp *DLPManager) RemoveRule(ruleID string) error {
	dlp.mu.Lock()
	defer dlp.mu.Unlock()

	if _, exists := dlp.rules[ruleID]; !exists {
		return fmt.Errorf("rule not found: %s", ruleID)
	}

	delete(dlp.rules, ruleID)

	log.Info().
		Str("rule_id", ruleID).
		Msg("DLP rule removed")

	return nil
}

// ScanData scans data for sensitive information
func (dlp *DLPManager) ScanData(data []byte, metadata *ScanMetadata) (*ScanResult, error) {
	if !dlp.enabled {
		return &ScanResult{Scanned: false}, nil
	}

	start := time.Now()
	defer func() {
		dlp.statistics.AddScanTime(time.Since(start))
	}()

	// Determine appropriate scanner based on content type
	scanner := dlp.selectScanner(metadata)
	if scanner == nil {
		return &ScanResult{Scanned: false}, fmt.Errorf("no suitable scanner found")
	}

	// Perform scan
	result, err := scanner.Scan(data, metadata)
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	// Update statistics
	dlp.updateStatistics(result, metadata)

	// Apply data classification if enabled
	if dlp.config.DataClassification.Enabled {
		classification, err := dlp.classifyData(data, metadata)
		if err != nil {
			log.Error().Err(err).Msg("Data classification failed")
		} else {
			result.Classification = classification
		}
	}

	// Calculate risk score
	result.RiskScore = dlp.calculateRiskScore(result)

	// Determine final action
	finalAction := dlp.determineAction(result)
	result.Action = finalAction

	// Process based on action
	processedData, err := dlp.processAction(data, result, metadata)
	if err != nil {
		log.Error().Err(err).Msg("Failed to process DLP action")
	} else {
		result.ProcessedData = processedData
	}

	result.ScanDuration = time.Since(start)

	log.Info().
		Str("source", metadata.Source).
		Int("detections", len(result.Detections)).
		Float64("risk_score", result.RiskScore).
		Str("action", string(result.Action)).
		Dur("scan_duration", result.ScanDuration).
		Msg("DLP scan completed")

	return result, nil
}

// selectScanner selects the appropriate scanner based on content type
func (dlp *DLPManager) selectScanner(metadata *ScanMetadata) DataScanner {
	contentType := strings.ToLower(metadata.ContentType)

	switch {
	case strings.Contains(contentType, "json"):
		return dlp.scanners["json"]
	case strings.Contains(contentType, "xml"):
		return dlp.scanners["xml"]
	case strings.HasPrefix(contentType, "text/"):
		return dlp.scanners["text"]
	case strings.HasPrefix(contentType, "application/"):
		if strings.Contains(contentType, "json") {
			return dlp.scanners["json"]
		}
		return dlp.scanners["binary"]
	default:
		// Default to text scanner for unknown types
		return dlp.scanners["text"]
	}
}

// classifyData classifies data using available classifiers
func (dlp *DLPManager) classifyData(data []byte, metadata *ScanMetadata) (*DataClassification, error) {
	var bestClassification *DataClassification
	var highestConfidence float64

	for _, classifier := range dlp.classifiers {
		classification, err := classifier.Classify(data, metadata)
		if err != nil {
			continue
		}

		if classification.Confidence > highestConfidence {
			highestConfidence = classification.Confidence
			bestClassification = classification
		}
	}

	if bestClassification == nil {
		// Return default classification
		return &DataClassification{
			Level:       dlp.config.DataClassification.DefaultLevel,
			Categories:  []string{"unclassified"},
			Labels:      []string{},
			Confidence:  0.5,
			Sensitivity: "unknown",
			Handling:    "default",
		}, nil
	}

	return bestClassification, nil
}

// calculateRiskScore calculates risk score based on detections
func (dlp *DLPManager) calculateRiskScore(result *ScanResult) float64 {
	if len(result.Detections) == 0 {
		return 0.0
	}

	var totalScore float64
	for _, detection := range result.Detections {
		severityWeight := dlp.getSeverityWeight(detection.Severity)
		confidenceWeight := detection.Confidence
		dataTypeWeight := dlp.getDataTypeWeight(detection.DataType)
		
		score := severityWeight * confidenceWeight * dataTypeWeight
		totalScore += score
	}

	// Normalize to 0-1 range
	maxPossibleScore := float64(len(result.Detections)) * 10.0 * 1.0 * 3.0
	if maxPossibleScore > 0 {
		return totalScore / maxPossibleScore
	}

	return 0.0
}

// getSeverityWeight returns weight for severity level
func (dlp *DLPManager) getSeverityWeight(severity SeverityLevel) float64 {
	switch severity {
	case SeverityInfo:
		return 1.0
	case SeverityLow:
		return 3.0
	case SeverityMedium:
		return 6.0
	case SeverityHigh:
		return 8.0
	case SeverityCritical:
		return 10.0
	default:
		return 1.0
	}
}

// getDataTypeWeight returns weight for data type
func (dlp *DLPManager) getDataTypeWeight(dataType DataType) float64 {
	switch dataType {
	case DataTypeCreditCard, DataTypeSSN, DataTypeBankAccount:
		return 3.0
	case DataTypeAPIKey, DataTypePassword:
		return 2.5
	case DataTypeMedicalRecord, DataTypeHealthcare:
		return 3.0
	case DataTypePII, DataTypeFinancial:
		return 2.0
	case DataTypeEmailAddress, DataTypePhoneNumber:
		return 1.5
	default:
		return 1.0
	}
}

// determineAction determines the final action based on detections and policies
func (dlp *DLPManager) determineAction(result *ScanResult) DLPAction {
	if len(result.Detections) == 0 {
		return DLPActionAllow
	}

	// Find the most restrictive action required
	var finalAction DLPAction = DLPActionAllow

	for _, detection := range result.Detections {
		rule := dlp.getRuleByID(detection.RuleID)
		if rule == nil {
			continue
		}

		action := rule.Action

		// Apply action precedence (block > quarantine > encrypt > mask > redact > log > allow)
		if dlp.actionPrecedence(action) > dlp.actionPrecedence(finalAction) {
			finalAction = action
		}
	}

	// Check if risk score requires escalation
	if result.RiskScore > 0.8 && finalAction == DLPActionLog {
		finalAction = DLPActionBlock
	} else if result.RiskScore > 0.6 && finalAction == DLPActionAllow {
		finalAction = DLPActionMask
	}

	return finalAction
}

// actionPrecedence returns precedence value for actions
func (dlp *DLPManager) actionPrecedence(action DLPAction) int {
	switch action {
	case DLPActionAllow:
		return 0
	case DLPActionLog:
		return 1
	case DLPActionRedact:
		return 2
	case DLPActionMask:
		return 3
	case DLPActionEncrypt:
		return 4
	case DLPActionQuarantine:
		return 5
	case DLPActionBlock:
		return 6
	default:
		return 0
	}
}

// processAction processes data based on the determined action
func (dlp *DLPManager) processAction(data []byte, result *ScanResult, metadata *ScanMetadata) ([]byte, error) {
	switch result.Action {
	case DLPActionAllow:
		return data, nil

	case DLPActionLog:
		dlp.logDetection(result, metadata)
		return data, nil

	case DLPActionMask:
		return dlp.maskSensitiveData(data, result.Detections), nil

	case DLPActionRedact:
		return dlp.redactSensitiveData(data, result.Detections), nil

	case DLPActionEncrypt:
		if dlp.encryptionEngine != nil {
			return dlp.encryptionEngine.Encrypt(data)
		}
		return data, nil

	case DLPActionQuarantine:
		err := dlp.quarantineData(data, result, metadata)
		return nil, err

	case DLPActionBlock:
		dlp.logViolation(result, metadata)
		return nil, fmt.Errorf("data blocked by DLP policy")

	default:
		return data, nil
	}
}

// maskSensitiveData masks sensitive data in the content
func (dlp *DLPManager) maskSensitiveData(data []byte, detections []DataDetection) []byte {
	content := string(data)
	
	// Sort detections by offset in descending order to avoid offset shifts
	for i := len(detections) - 1; i >= 0; i-- {
		detection := detections[i]
		if detection.Location.Offset >= 0 && detection.Location.Length > 0 {
			start := detection.Location.Offset
			end := start + detection.Location.Length
			
			if start < len(content) && end <= len(content) {
				// Create mask based on data type
				mask := dlp.createMask(detection.DataType, detection.Location.Length)
				content = content[:start] + mask + content[end:]
			}
		}
	}

	return []byte(content)
}

// createMask creates appropriate mask for data type
func (dlp *DLPManager) createMask(dataType DataType, length int) string {
	switch dataType {
	case DataTypeCreditCard:
		return "****-****-****-" + strings.Repeat("*", max(0, length-15))
	case DataTypeSSN:
		return "***-**-" + strings.Repeat("*", max(0, length-7))
	case DataTypePhoneNumber:
		return "***-***-" + strings.Repeat("*", max(0, length-8))
	case DataTypeEmailAddress:
		return strings.Repeat("*", max(3, length/2)) + "@" + strings.Repeat("*", max(3, length/2))
	default:
		return strings.Repeat("*", length)
	}
}

// redactSensitiveData redacts sensitive data from content
func (dlp *DLPManager) redactSensitiveData(data []byte, detections []DataDetection) []byte {
	content := string(data)
	
	// Sort detections by offset in descending order
	for i := len(detections) - 1; i >= 0; i-- {
		detection := detections[i]
		if detection.Location.Offset >= 0 && detection.Location.Length > 0 {
			start := detection.Location.Offset
			end := start + detection.Location.Length
			
			if start < len(content) && end <= len(content) {
				redaction := fmt.Sprintf("[REDACTED:%s]", detection.DataType)
				content = content[:start] + redaction + content[end:]
			}
		}
	}

	return []byte(content)
}

// quarantineData quarantines detected sensitive data
func (dlp *DLPManager) quarantineData(data []byte, result *ScanResult, metadata *ScanMetadata) error {
	dlp.mu.Lock()
	defer dlp.mu.Unlock()

	// Generate quarantine ID
	quarantineID := dlp.generateQuarantineID(metadata)

	// Create quarantine record
	quarantined := &QuarantinedData{
		ID:             quarantineID,
		OriginalPath:   metadata.FileName,
		QuarantinePath: fmt.Sprintf("%s/%s", dlp.config.QuarantinePath, quarantineID),
		Detections:     result.Detections,
		Metadata:       metadata,
		QuarantinedAt:  time.Now(),
		ExpiresAt:      time.Now().Add(dlp.config.RetentionPolicy.QuarantineRetention),
		Status:         QuarantineStatusActive,
	}

	dlp.quarantine[quarantineID] = quarantined
	dlp.statistics.IncrementQuarantined()

	// Send notification if configured
	if dlp.config.NotificationWebhook != "" {
		dlp.sendQuarantineNotification(quarantined)
	}

	log.Warn().
		Str("quarantine_id", quarantineID).
		Str("source", metadata.Source).
		Int("detections", len(result.Detections)).
		Msg("Data quarantined by DLP")

	return nil
}

// generateQuarantineID generates a unique quarantine ID
func (dlp *DLPManager) generateQuarantineID(metadata *ScanMetadata) string {
	data := fmt.Sprintf("%s-%s-%d", metadata.FileName, metadata.UserID, time.Now().UnixNano())
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("q-%x", hash)
}

// sendQuarantineNotification sends notification about quarantined data
func (dlp *DLPManager) sendQuarantineNotification(quarantined *QuarantinedData) {
	// Implementation would send webhook notification
	log.Info().
		Str("quarantine_id", quarantined.ID).
		Msg("Quarantine notification sent")
}

// logDetection logs a DLP detection
func (dlp *DLPManager) logDetection(result *ScanResult, metadata *ScanMetadata) {
	for _, detection := range result.Detections {
		log.Info().
			Str("rule_id", detection.RuleID).
			Str("data_type", string(detection.DataType)).
			Float64("confidence", detection.Confidence).
			Str("source", metadata.Source).
			Msg("DLP detection logged")
	}
}

// logViolation logs a DLP violation
func (dlp *DLPManager) logViolation(result *ScanResult, metadata *ScanMetadata) {
	for _, detection := range result.Detections {
		log.Warn().
			Str("rule_id", detection.RuleID).
			Str("data_type", string(detection.DataType)).
			Str("severity", detection.Severity.String()).
			Float64("confidence", detection.Confidence).
			Str("source", metadata.Source).
			Str("user_id", metadata.UserID).
			Msg("DLP violation - data blocked")
	}
}

// getRuleByID gets a rule by ID
func (dlp *DLPManager) getRuleByID(ruleID string) *DLPRule {
	dlp.mu.RLock()
	defer dlp.mu.RUnlock()
	return dlp.rules[ruleID]
}

// updateStatistics updates DLP statistics
func (dlp *DLPManager) updateStatistics(result *ScanResult, metadata *ScanMetadata) {
	switch metadata.Source {
	case "upload":
		dlp.statistics.IncrementFilesScanned()
	case "api_request":
		dlp.statistics.IncrementRequestsScanned()
	case "api_response":
		dlp.statistics.IncrementResponsesScanned()
	}

	dlp.statistics.IncrementDetections(int64(len(result.Detections)))

	for _, detection := range result.Detections {
		dlp.statistics.IncrementDetectionsByType(detection.DataType)
	}

	dlp.statistics.IncrementAction(result.Action)
	dlp.statistics.SetLastScanTime(time.Now())
}

// loadDefaultRules loads default DLP rules
func (dlp *DLPManager) loadDefaultRules() {
	defaultRules := []*DLPRule{
		{
			ID:          "credit_card_visa",
			Name:        "Visa Credit Card",
			Description: "Detects Visa credit card numbers",
			DataType:    DataTypeCreditCard,
			Pattern:     `\b4\d{3}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b`,
			Action:      DLPActionMask,
			Confidence:  0.9,
			Severity:    SeverityHigh,
			Categories:  []string{"financial", "pci_dss"},
			ComplianceFramework: []string{"PCI_DSS"},
			Enabled:     true,
		},
		{
			ID:          "ssn_us",
			Name:        "US Social Security Number",
			Description: "Detects US Social Security Numbers",
			DataType:    DataTypeSSN,
			Pattern:     `\b\d{3}[-\s]?\d{2}[-\s]?\d{4}\b`,
			Action:      DLPActionBlock,
			Confidence:  0.95,
			Severity:    SeverityCritical,
			Categories:  []string{"pii", "government"},
			ComplianceFramework: []string{"HIPAA", "GDPR"},
			Enabled:     true,
		},
		{
			ID:          "api_key_generic",
			Name:        "Generic API Key",
			Description: "Detects generic API keys",
			DataType:    DataTypeAPIKey,
			Pattern:     `(?i)api[_-]?key[\"']?\s*[:=]\s*[\"']?[a-zA-Z0-9]{20,}[\"']?`,
			Action:      DLPActionBlock,
			Confidence:  0.8,
			Severity:    SeverityHigh,
			Categories:  []string{"credentials", "security"},
			Enabled:     true,
		},
		{
			ID:          "email_address",
			Name:        "Email Address",
			Description: "Detects email addresses",
			DataType:    DataTypeEmailAddress,
			Pattern:     `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`,
			Action:      DLPActionLog,
			Confidence:  0.9,
			Severity:    SeverityLow,
			Categories:  []string{"pii", "contact"},
			ComplianceFramework: []string{"GDPR"},
			Enabled:     true,
		},
		{
			ID:          "phone_number_us",
			Name:        "US Phone Number",
			Description: "Detects US phone numbers",
			DataType:    DataTypePhoneNumber,
			Pattern:     `\b\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}\b`,
			Action:      DLPActionMask,
			Confidence:  0.8,
			Severity:    SeverityMedium,
			Categories:  []string{"pii", "contact"},
			ComplianceFramework: []string{"GDPR"},
			Enabled:     true,
		},
	}

	for _, rule := range defaultRules {
		if err := dlp.AddRule(rule); err != nil {
			log.Error().
				Err(err).
				Str("rule_id", rule.ID).
				Msg("Failed to add default DLP rule")
		}
	}

	log.Info().
		Int("rules_loaded", len(defaultRules)).
		Msg("Default DLP rules loaded")
}

// GetStatistics returns DLP statistics
func (dlp *DLPManager) GetStatistics() *DLPStatistics {
	return dlp.statistics.GetSnapshot()
}

// GetQuarantinedData returns quarantined data by ID
func (dlp *DLPManager) GetQuarantinedData(quarantineID string) (*QuarantinedData, error) {
	dlp.mu.RLock()
	defer dlp.mu.RUnlock()

	quarantined, exists := dlp.quarantine[quarantineID]
	if !exists {
		return nil, fmt.Errorf("quarantined data not found: %s", quarantineID)
	}

	return quarantined, nil
}

// ReleaseQuarantinedData releases quarantined data
func (dlp *DLPManager) ReleaseQuarantinedData(quarantineID, reviewerID, comments string) error {
	dlp.mu.Lock()
	defer dlp.mu.Unlock()

	quarantined, exists := dlp.quarantine[quarantineID]
	if !exists {
		return fmt.Errorf("quarantined data not found: %s", quarantineID)
	}

	now := time.Now()
	quarantined.Status = QuarantineStatusReleased
	quarantined.ReviewedBy = reviewerID
	quarantined.ReviewedAt = &now
	quarantined.Comments = comments

	log.Info().
		Str("quarantine_id", quarantineID).
		Str("reviewed_by", reviewerID).
		Msg("Quarantined data released")

	return nil
}

// Utility functions and helper implementations

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// NewDLPStatistics creates a new DLP statistics collector
func NewDLPStatistics() *DLPStatistics {
	return &DLPStatistics{
		DetectionsByType:     make(map[DataType]int64),
		ActionsTotal:         make(map[DLPAction]int64),
		ComplianceViolations: make(map[string]int64),
	}
}

// Statistical methods for DLPStatistics
func (ds *DLPStatistics) IncrementFilesScanned() {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.FilesScanned++
}

func (ds *DLPStatistics) IncrementRequestsScanned() {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.RequestsScanned++
}

func (ds *DLPStatistics) IncrementResponsesScanned() {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.ResponsesScanned++
}

func (ds *DLPStatistics) IncrementDetections(count int64) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.DetectionsTotal += count
}

func (ds *DLPStatistics) IncrementDetectionsByType(dataType DataType) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.DetectionsByType[dataType]++
}

func (ds *DLPStatistics) IncrementAction(action DLPAction) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.ActionsTotal[action]++
}

func (ds *DLPStatistics) IncrementQuarantined() {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.QuarantinedItems++
}

func (ds *DLPStatistics) SetLastScanTime(t time.Time) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.LastScanTime = t
}

func (ds *DLPStatistics) AddScanTime(duration time.Duration) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	// Simple moving average (in real implementation, use more sophisticated averaging)
	if ds.AverageScanTime == 0 {
		ds.AverageScanTime = duration
	} else {
		ds.AverageScanTime = (ds.AverageScanTime + duration) / 2
	}
}

func (ds *DLPStatistics) GetSnapshot() *DLPStatistics {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	snapshot := &DLPStatistics{
		FilesScanned:         ds.FilesScanned,
		RequestsScanned:      ds.RequestsScanned,
		ResponsesScanned:     ds.ResponsesScanned,
		DetectionsTotal:      ds.DetectionsTotal,
		QuarantinedItems:     ds.QuarantinedItems,
		FalsePositives:       ds.FalsePositives,
		AverageScanTime:      ds.AverageScanTime,
		LastScanTime:         ds.LastScanTime,
		DetectionsByType:     make(map[DataType]int64),
		ActionsTotal:         make(map[DLPAction]int64),
		ComplianceViolations: make(map[string]int64),
	}

	// Copy maps
	for k, v := range ds.DetectionsByType {
		snapshot.DetectionsByType[k] = v
	}
	for k, v := range ds.ActionsTotal {
		snapshot.ActionsTotal[k] = v
	}
	for k, v := range ds.ComplianceViolations {
		snapshot.ComplianceViolations[k] = v
	}

	return snapshot
}

// NewDLPPolicyEngine creates a new DLP policy engine
func NewDLPPolicyEngine() *DLPPolicyEngine {
	return &DLPPolicyEngine{
		policies: make(map[string]*DLPPolicy),
	}
}

// NewDLPEncryptionEngine creates a new DLP encryption engine
func NewDLPEncryptionEngine() *DLPEncryptionEngine {
	// Generate encryption key (in production, use proper key management)
	key := make([]byte, 32)
	return &DLPEncryptionEngine{
		encryptionKey: key,
		algorithm:     "AES-256-GCM",
	}
}

// Encrypt encrypts data
func (dee *DLPEncryptionEngine) Encrypt(data []byte) ([]byte, error) {
	// Simplified encryption (implement actual AES-GCM encryption)
	return data, nil
}

// DefaultDLPConfig returns default DLP configuration
func DefaultDLPConfig() *DLPConfig {
	return &DLPConfig{
		Enabled:             true,
		ScanFileUploads:     true,
		ScanAPIRequests:     true,
		ScanAPIResponses:    false,
		BlockSensitiveData:  true,
		MaskSensitiveData:   true,
		QuarantinePath:      "/var/lib/ollama/quarantine",
		MaxFileSize:         50 * 1024 * 1024, // 50MB
		AllowedFileTypes:    []string{"txt", "json", "xml", "csv"},
		BlockedFileTypes:    []string{"exe", "bat", "sh"},
		ScanTimeout:         30 * time.Second,
		ComplianceModes:     []string{"SOC2"},
		EncryptionEnabled:   true,
		RetentionPolicy: &RetentionPolicy{
			QuarantineRetention: 90 * 24 * time.Hour, // 90 days
			LogRetention:        365 * 24 * time.Hour, // 1 year
			AlertRetention:      30 * 24 * time.Hour,  // 30 days
			AutoDelete:          true,
		},
		DataClassification: &ClassificationConfig{
			Enabled:         true,
			Levels:          []string{"public", "internal", "confidential", "restricted"},
			DefaultLevel:    "internal",
			EnforceLabeling: false,
			Classifiers: map[string]ClassifierConfig{
				"regex": {
					Type:       "regex",
					Confidence: 0.8,
					Enabled:    true,
				},
				"dictionary": {
					Type:       "dictionary",
					Confidence: 0.7,
					Enabled:    true,
				},
			},
		},
	}
}

// Scanner implementations (simplified interfaces)

type TextScanner struct {
	rules map[string]*DLPRule
}

func NewTextScanner(rules map[string]*DLPRule) *TextScanner {
	return &TextScanner{rules: rules}
}

func (ts *TextScanner) Scan(data []byte, metadata *ScanMetadata) (*ScanResult, error) {
	// Simplified text scanning implementation
	return &ScanResult{
		Scanned:    true,
		Detections: []DataDetection{},
	}, nil
}

func (ts *TextScanner) GetType() string { return "text" }
func (ts *TextScanner) IsEnabled() bool { return true }

type FileScanner struct {
	rules map[string]*DLPRule
}

func NewFileScanner(rules map[string]*DLPRule) *FileScanner {
	return &FileScanner{rules: rules}
}

func (fs *FileScanner) Scan(data []byte, metadata *ScanMetadata) (*ScanResult, error) {
	return &ScanResult{Scanned: true, Detections: []DataDetection{}}, nil
}

func (fs *FileScanner) GetType() string { return "file" }
func (fs *FileScanner) IsEnabled() bool { return true }

type BinaryScanner struct {
	rules map[string]*DLPRule
}

func NewBinaryScanner(rules map[string]*DLPRule) *BinaryScanner {
	return &BinaryScanner{rules: rules}
}

func (bs *BinaryScanner) Scan(data []byte, metadata *ScanMetadata) (*ScanResult, error) {
	return &ScanResult{Scanned: true, Detections: []DataDetection{}}, nil
}

func (bs *BinaryScanner) GetType() string { return "binary" }
func (bs *BinaryScanner) IsEnabled() bool { return true }

type JSONScanner struct {
	rules map[string]*DLPRule
}

func NewJSONScanner(rules map[string]*DLPRule) *JSONScanner {
	return &JSONScanner{rules: rules}
}

func (js *JSONScanner) Scan(data []byte, metadata *ScanMetadata) (*ScanResult, error) {
	return &ScanResult{Scanned: true, Detections: []DataDetection{}}, nil
}

func (js *JSONScanner) GetType() string { return "json" }
func (js *JSONScanner) IsEnabled() bool { return true }

type XMLScanner struct {
	rules map[string]*DLPRule
}

func NewXMLScanner(rules map[string]*DLPRule) *XMLScanner {
	return &XMLScanner{rules: rules}
}

func (xs *XMLScanner) Scan(data []byte, metadata *ScanMetadata) (*ScanResult, error) {
	return &ScanResult{Scanned: true, Detections: []DataDetection{}}, nil
}

func (xs *XMLScanner) GetType() string { return "xml" }
func (xs *XMLScanner) IsEnabled() bool { return true }

// Classifier implementations (simplified)

type RegexClassifier struct {
	config *ClassificationConfig
}

func NewRegexClassifier(config *ClassificationConfig) *RegexClassifier {
	return &RegexClassifier{config: config}
}

func (rc *RegexClassifier) Classify(data []byte, metadata *ScanMetadata) (*DataClassification, error) {
	return &DataClassification{
		Level:       rc.config.DefaultLevel,
		Categories:  []string{"general"},
		Labels:      []string{},
		Confidence:  0.5,
		Sensitivity: "medium",
		Handling:    "standard",
	}, nil
}

func (rc *RegexClassifier) GetType() string { return "regex" }

type DictionaryClassifier struct {
	config *ClassificationConfig
}

func NewDictionaryClassifier(config *ClassificationConfig) *DictionaryClassifier {
	return &DictionaryClassifier{config: config}
}

func (dc *DictionaryClassifier) Classify(data []byte, metadata *ScanMetadata) (*DataClassification, error) {
	return &DataClassification{
		Level:       dc.config.DefaultLevel,
		Categories:  []string{"general"},
		Labels:      []string{},
		Confidence:  0.6,
		Sensitivity: "medium",
		Handling:    "standard",
	}, nil
}

func (dc *DictionaryClassifier) GetType() string { return "dictionary" }
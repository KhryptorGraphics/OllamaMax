package training

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CertificationTest represents the certification assessment framework
type CertificationTest struct {
	CertificationID   string                `json:"certification_id"`
	Name              string                `json:"name"`
	Version           string                `json:"version"`
	Prerequisites     []string              `json:"prerequisites"`
	Assessments       []Assessment          `json:"assessments"`
	PassingCriteria   PassingCriteria       `json:"passing_criteria"`
	TimeLimit         time.Duration         `json:"time_limit"`
	SecurityMeasures  SecurityMeasures      `json:"security"`
	Certificate       CertificateTemplate   `json:"certificate"`
}

// Assessment represents a single assessment component
type Assessment struct {
	ID              string              `json:"id"`
	Type            AssessmentType      `json:"type"`
	Weight          float64            `json:"weight"`
	Questions       []Question         `json:"questions"`
	PracticalTasks  []PracticalTask    `json:"practical_tasks"`
	TimeAllocation  time.Duration      `json:"time_allocation"`
	PassingScore    float64            `json:"passing_score"`
}

// AssessmentType defines different types of assessments
type AssessmentType string

const (
	AssessmentTypeKnowledge   AssessmentType = "knowledge"
	AssessmentTypePractical   AssessmentType = "practical"
	AssessmentTypeScenario    AssessmentType = "scenario"
	AssessmentTypeIntegration AssessmentType = "integration"
)

// Question represents a knowledge assessment question
type Question struct {
	ID           string        `json:"id"`
	Type         QuestionType  `json:"type"`
	Question     string        `json:"question"`
	Options      []string      `json:"options,omitempty"`
	CorrectAnswer interface{}  `json:"correct_answer"`
	Points       int          `json:"points"`
	Explanation  string       `json:"explanation"`
	Difficulty   string       `json:"difficulty"`
	LearningObj  string       `json:"learning_objective"`
}

// QuestionType defines different question formats
type QuestionType string

const (
	QuestionTypeMultipleChoice QuestionType = "multiple_choice"
	QuestionTypeShortAnswer    QuestionType = "short_answer"
	QuestionTypeTrueFalse      QuestionType = "true_false"
	QuestionTypeCode           QuestionType = "code"
	QuestionTypeScenario       QuestionType = "scenario"
)

// PracticalTask represents hands-on assessment tasks
type PracticalTask struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Instructions    []string               `json:"instructions"`
	ExpectedOutputs []ExpectedOutput       `json:"expected_outputs"`
	ValidationRules []ValidationRule       `json:"validation_rules"`
	TimeLimit       time.Duration          `json:"time_limit"`
	Points          int                    `json:"points"`
	Environment     map[string]string      `json:"environment"`
}

// ExpectedOutput defines what should be produced by practical tasks
type ExpectedOutput struct {
	Type        string      `json:"type"` // file, command_output, api_response
	Target      string      `json:"target"`
	Content     string      `json:"content,omitempty"`
	Pattern     string      `json:"pattern,omitempty"`
	Validation  []string    `json:"validation"`
}

// PassingCriteria defines requirements for certification
type PassingCriteria struct {
	OverallScore      float64            `json:"overall_score"`
	MinimumScores     map[string]float64 `json:"minimum_scores"`
	RequiredTasks     []string           `json:"required_tasks"`
	TimeCompliance    bool               `json:"time_compliance"`
	IntegrityCheck    bool               `json:"integrity_check"`
}

// SecurityMeasures defines anti-cheating and security measures
type SecurityMeasures struct {
	ProctorRequired     bool     `json:"proctor_required"`
	BrowserRestriction  bool     `json:"browser_restriction"`
	ScreenCapture       bool     `json:"screen_capture"`
	KeystrokeMonitoring bool     `json:"keystroke_monitoring"`
	TimeTracking        bool     `json:"time_tracking"`
	IPRestriction       []string `json:"ip_restriction"`
	DeviceFingerprint   bool     `json:"device_fingerprint"`
}

// CertificateTemplate defines the certificate format
type CertificateTemplate struct {
	Template        string            `json:"template"`
	Fields          map[string]string `json:"fields"`
	DigitalSignature bool             `json:"digital_signature"`
	VerificationURL string            `json:"verification_url"`
	Validity        time.Duration     `json:"validity"`
}

// CertificationResult stores assessment results
type CertificationResult struct {
	CandidateID       string                    `json:"candidate_id"`
	CertificationID   string                    `json:"certification_id"`
	StartTime         time.Time                 `json:"start_time"`
	EndTime           time.Time                 `json:"end_time"`
	Duration          time.Duration             `json:"duration"`
	Scores            map[string]float64        `json:"scores"`
	OverallScore      float64                   `json:"overall_score"`
	Passed            bool                      `json:"passed"`
	AssessmentResults []AssessmentResult        `json:"assessment_results"`
	SecurityEvents    []SecurityEvent           `json:"security_events"`
	CertificateID     string                    `json:"certificate_id,omitempty"`
	IssuedDate        *time.Time               `json:"issued_date,omitempty"`
}

// AssessmentResult stores individual assessment results
type AssessmentResult struct {
	AssessmentID    string                 `json:"assessment_id"`
	Score           float64               `json:"score"`
	MaxScore        float64               `json:"max_score"`
	Percentage      float64               `json:"percentage"`
	Passed          bool                  `json:"passed"`
	TimeSpent       time.Duration         `json:"time_spent"`
	QuestionResults []QuestionResult      `json:"question_results"`
	TaskResults     []TaskResult          `json:"task_results"`
}

// QuestionResult stores individual question results
type QuestionResult struct {
	QuestionID    string      `json:"question_id"`
	Answer        interface{} `json:"answer"`
	Correct       bool        `json:"correct"`
	Points        int         `json:"points"`
	TimeSpent     time.Duration `json:"time_spent"`
}

// TaskResult stores practical task results
type TaskResult struct {
	TaskID        string              `json:"task_id"`
	Completed     bool                `json:"completed"`
	Score         float64            `json:"score"`
	TimeSpent     time.Duration      `json:"time_spent"`
	Outputs       []ActualOutput     `json:"outputs"`
	ValidationResults []ValidationResult `json:"validation_results"`
}

// ActualOutput represents what was actually produced
type ActualOutput struct {
	Type    string `json:"type"`
	Target  string `json:"target"`
	Content string `json:"content"`
}

// ValidationResult stores validation outcomes
type ValidationResult struct {
	Rule    ValidationRule `json:"rule"`
	Passed  bool          `json:"passed"`
	Message string        `json:"message"`
}

// SecurityEvent records security-related events during assessment
type SecurityEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	Data        map[string]interface{} `json:"data"`
}

// TestCertificationFramework tests the overall certification system
func TestCertificationFramework(t *testing.T) {
	cert := createOllamaDistributedCertification()
	
	t.Run("CertificationStructure", func(t *testing.T) {
		assert.NotEmpty(t, cert.CertificationID, "Certification must have ID")
		assert.NotEmpty(t, cert.Name, "Certification must have name")
		assert.NotEmpty(t, cert.Version, "Certification must have version")
		assert.NotEmpty(t, cert.Assessments, "Certification must have assessments")
		assert.Greater(t, cert.TimeLimit, time.Duration(0), "Must have time limit")
	})

	t.Run("AssessmentValidation", func(t *testing.T) {
		totalWeight := 0.0
		for _, assessment := range cert.Assessments {
			assert.NotEmpty(t, assessment.ID, "Assessment must have ID")
			assert.Greater(t, assessment.Weight, 0.0, "Assessment must have positive weight")
			assert.LessOrEqual(t, assessment.Weight, 1.0, "Assessment weight cannot exceed 1.0")
			totalWeight += assessment.Weight
		}
		assert.Equal(t, 1.0, totalWeight, "Total assessment weights must equal 1.0")
	})

	t.Run("PassingCriteriaValidation", func(t *testing.T) {
		criteria := cert.PassingCriteria
		assert.Greater(t, criteria.OverallScore, 0.0, "Must have minimum overall score")
		assert.LessOrEqual(t, criteria.OverallScore, 1.0, "Overall score cannot exceed 1.0")
		assert.NotEmpty(t, criteria.RequiredTasks, "Must have required tasks")
	})
}

// TestKnowledgeAssessment tests the knowledge-based questions
func TestKnowledgeAssessment(t *testing.T) {
	questions := createKnowledgeQuestions()

	t.Run("QuestionStructure", func(t *testing.T) {
		for _, q := range questions {
			assert.NotEmpty(t, q.ID, "Question must have ID")
			assert.NotEmpty(t, q.Question, "Question must have text")
			assert.Greater(t, q.Points, 0, "Question must have positive points")
			assert.NotEmpty(t, q.LearningObj, "Question must map to learning objective")
			
			if q.Type == QuestionTypeMultipleChoice {
				assert.NotEmpty(t, q.Options, "Multiple choice must have options")
				assert.Greater(t, len(q.Options), 1, "Must have multiple options")
			}
		}
	})

	t.Run("QuestionDifficulty", func(t *testing.T) {
		difficulties := map[string]int{}
		for _, q := range questions {
			difficulties[q.Difficulty]++
		}
		
		// Ensure balanced difficulty distribution
		assert.Greater(t, difficulties["easy"], 0, "Must have easy questions")
		assert.Greater(t, difficulties["medium"], 0, "Must have medium questions") 
		assert.Greater(t, difficulties["hard"], 0, "Must have hard questions")
	})

	t.Run("LearningObjectiveCoverage", func(t *testing.T) {
		objectives := map[string]int{}
		for _, q := range questions {
			objectives[q.LearningObj]++
		}

		requiredObjectives := []string{
			"installation_setup",
			"configuration_management",
			"cluster_operations",
			"api_integration",
			"troubleshooting",
		}

		for _, obj := range requiredObjectives {
			assert.Greater(t, objectives[obj], 0, "Must have questions for objective: %s", obj)
		}
	})
}

// TestPracticalAssessment tests hands-on practical tasks
func TestPracticalAssessment(t *testing.T) {
	tasks := createPracticalTasks()

	t.Run("TaskStructure", func(t *testing.T) {
		for _, task := range tasks {
			assert.NotEmpty(t, task.ID, "Task must have ID")
			assert.NotEmpty(t, task.Name, "Task must have name")
			assert.NotEmpty(t, task.Description, "Task must have description")
			assert.NotEmpty(t, task.Instructions, "Task must have instructions")
			assert.NotEmpty(t, task.ExpectedOutputs, "Task must have expected outputs")
			assert.Greater(t, task.Points, 0, "Task must have positive points")
			assert.Greater(t, task.TimeLimit, time.Duration(0), "Task must have time limit")
		}
	})

	t.Run("TaskValidation", func(t *testing.T) {
		for _, task := range tasks {
			t.Run(task.ID, func(t *testing.T) {
				// Validate each expected output
				for _, output := range task.ExpectedOutputs {
					assert.NotEmpty(t, output.Type, "Output must have type")
					assert.NotEmpty(t, output.Target, "Output must have target")
					
					if output.Type == "file" {
						assert.NotEmpty(t, output.Content, "File output must specify content")
					}
				}
				
				// Validate validation rules
				for _, rule := range task.ValidationRules {
					assert.NotEmpty(t, rule.Type, "Validation rule must have type")
					assert.NotEmpty(t, rule.Target, "Validation rule must have target")
				}
			})
		}
	})
}

// TestSecurityMeasures tests anti-cheating and security features
func TestSecurityMeasures(t *testing.T) {
	cert := createOllamaDistributedCertification()
	security := cert.SecurityMeasures

	t.Run("SecurityConfiguration", func(t *testing.T) {
		// Test security measures are properly configured
		assert.True(t, security.TimeTracking, "Time tracking should be enabled")
		assert.True(t, security.DeviceFingerprint, "Device fingerprinting should be enabled")
		
		// Test that at least one monitoring measure is enabled
		monitoringEnabled := security.ScreenCapture || security.KeystrokeMonitoring || security.ProctorRequired
		assert.True(t, monitoringEnabled, "At least one monitoring measure should be enabled")
	})

	t.Run("CheatingPrevention", func(t *testing.T) {
		// Test anti-cheating measures
		cheatingEvents := []SecurityEvent{
			{
				Type:        "tab_switch",
				Description: "User switched browser tab",
				Severity:    "warning",
				Timestamp:   time.Now(),
			},
			{
				Type:        "copy_paste",
				Description: "Copy-paste activity detected",
				Severity:    "high",
				Timestamp:   time.Now(),
			},
			{
				Type:        "network_request",
				Description: "Unauthorized network request",
				Severity:    "critical",
				Timestamp:   time.Now(),
			},
		}

		for _, event := range cheatingEvents {
			severity := evaluateSecurityEvent(event)
			assert.NotEmpty(t, severity, "Security event must have severity")
			t.Logf("Security event %s evaluated as %s", event.Type, severity)
		}
	})
}

// TestCertificateGeneration tests certificate creation and validation
func TestCertificateGeneration(t *testing.T) {
	result := &CertificationResult{
		CandidateID:     "candidate-123",
		CertificationID: "ollama-distributed-cert-v1",
		StartTime:       time.Now().Add(-45 * time.Minute),
		EndTime:         time.Now(),
		Duration:        45 * time.Minute,
		OverallScore:    0.87,
		Passed:          true,
		Scores: map[string]float64{
			"knowledge":   0.85,
			"practical":   0.90,
			"integration": 0.85,
		},
	}

	t.Run("CertificateCreation", func(t *testing.T) {
		cert, err := generateCertificate(result)
		require.NoError(t, err, "Certificate generation should succeed")
		assert.NotEmpty(t, cert.ID, "Certificate must have ID")
		assert.NotEmpty(t, cert.Hash, "Certificate must have hash")
		assert.NotEmpty(t, cert.Signature, "Certificate must be signed")
		assert.Equal(t, result.CandidateID, cert.CandidateID, "Certificate should match candidate")
	})

	t.Run("CertificateValidation", func(t *testing.T) {
		cert, err := generateCertificate(result)
		require.NoError(t, err)
		
		// Test certificate validation
		valid, err := validateCertificate(cert)
		require.NoError(t, err, "Certificate validation should not error")
		assert.True(t, valid, "Generated certificate should be valid")
	})

	t.Run("CertificateAntiForging", func(t *testing.T) {
		cert, err := generateCertificate(result)
		require.NoError(t, err)
		
		// Test tampering detection
		originalHash := cert.Hash
		cert.Data["score"] = "0.99" // Tamper with score
		
		valid, err := validateCertificate(cert)
		assert.NoError(t, err)
		assert.False(t, valid, "Tampered certificate should be invalid")
		
		// Restore original and test again
		cert.Hash = originalHash
		cert.Data["score"] = "0.87"
		valid, err = validateCertificate(cert)
		assert.NoError(t, err)
		assert.True(t, valid, "Restored certificate should be valid")
	})
}

// TestAssessmentExecution simulates assessment execution
func TestAssessmentExecution(t *testing.T) {
	cert := createOllamaDistributedCertification()
	candidateID := "test-candidate-" + generateID()

	t.Run("AssessmentSession", func(t *testing.T) {
		session := &AssessmentSession{
			CandidateID:     candidateID,
			CertificationID: cert.CertificationID,
			StartTime:       time.Now(),
			TimeLimit:       cert.TimeLimit,
		}

		// Simulate assessment execution
		result, err := executeAssessment(session, cert)
		require.NoError(t, err, "Assessment execution should succeed")
		
		assert.Equal(t, candidateID, result.CandidateID)
		assert.NotZero(t, result.OverallScore, "Should have overall score")
		assert.Greater(t, result.Duration, time.Duration(0), "Should have duration")
		assert.NotEmpty(t, result.AssessmentResults, "Should have assessment results")
	})

	t.Run("ScoreCalculation", func(t *testing.T) {
		scores := map[string]float64{
			"knowledge":   0.85,
			"practical":   0.90,
			"integration": 0.80,
		}
		
		weights := map[string]float64{
			"knowledge":   0.40,
			"practical":   0.40,
			"integration": 0.20,
		}

		overall := calculateOverallScore(scores, weights)
		expected := (0.85*0.40) + (0.90*0.40) + (0.80*0.20)
		assert.InDelta(t, expected, overall, 0.01, "Overall score calculation should be accurate")
	})

	t.Run("PassingDetermination", func(t *testing.T) {
		criteria := PassingCriteria{
			OverallScore: 0.75,
			MinimumScores: map[string]float64{
				"knowledge":   0.70,
				"practical":   0.70,
				"integration": 0.65,
			},
			RequiredTasks: []string{"task-1", "task-2"},
		}

		// Test passing case
		result := &CertificationResult{
			OverallScore: 0.85,
			Scores: map[string]float64{
				"knowledge":   0.85,
				"practical":   0.90,
				"integration": 0.80,
			},
		}
		
		passed := evaluatePassingCriteria(result, criteria)
		assert.True(t, passed, "Should pass with scores above criteria")

		// Test failing case - overall score too low
		result.OverallScore = 0.70
		passed = evaluatePassingCriteria(result, criteria)
		assert.False(t, passed, "Should fail with overall score below criteria")

		// Test failing case - individual score too low
		result.OverallScore = 0.85
		result.Scores["practical"] = 0.65
		passed = evaluatePassingCriteria(result, criteria)
		assert.False(t, passed, "Should fail with individual score below minimum")
	})
}

// Helper Functions and Data Structures

type Certificate struct {
	ID          string            `json:"id"`
	CandidateID string            `json:"candidate_id"`
	Hash        string            `json:"hash"`
	Signature   string            `json:"signature"`
	IssuedDate  time.Time         `json:"issued_date"`
	ExpiryDate  time.Time         `json:"expiry_date"`
	Data        map[string]string `json:"data"`
}

type AssessmentSession struct {
	CandidateID     string
	CertificationID string
	StartTime       time.Time
	TimeLimit       time.Duration
	SecureMode      bool
}

func createOllamaDistributedCertification() *CertificationTest {
	return &CertificationTest{
		CertificationID: "ollama-distributed-cert-v1",
		Name:           "Ollama Distributed Systems Certification",
		Version:        "1.0.0",
		Prerequisites:  []string{"basic-networking", "command-line", "json-understanding"},
		TimeLimit:      60 * time.Minute,
		Assessments: []Assessment{
			{
				ID:             "knowledge-assessment",
				Type:           AssessmentTypeKnowledge,
				Weight:         0.40,
				TimeAllocation: 20 * time.Minute,
				PassingScore:   0.75,
			},
			{
				ID:             "practical-assessment",
				Type:           AssessmentTypePractical,
				Weight:         0.40,
				TimeAllocation: 30 * time.Minute,
				PassingScore:   0.70,
			},
			{
				ID:             "integration-assessment",
				Type:           AssessmentTypeIntegration,
				Weight:         0.20,
				TimeAllocation: 10 * time.Minute,
				PassingScore:   0.65,
			},
		},
		PassingCriteria: PassingCriteria{
			OverallScore: 0.75,
			MinimumScores: map[string]float64{
				"knowledge":   0.70,
				"practical":   0.70,
				"integration": 0.65,
			},
			RequiredTasks:  []string{"installation-task", "configuration-task", "api-integration-task"},
			TimeCompliance: true,
			IntegrityCheck: true,
		},
		SecurityMeasures: SecurityMeasures{
			TimeTracking:      true,
			DeviceFingerprint: true,
			ScreenCapture:     false, // Optional for privacy
			ProctorRequired:   false, // Self-paced certification
		},
		Certificate: CertificateTemplate{
			Template:        "ollama-distributed-cert-template",
			DigitalSignature: true,
			Validity:        2 * 365 * 24 * time.Hour, // 2 years
		},
	}
}

func createKnowledgeQuestions() []Question {
	return []Question{
		{
			ID:           "q1-installation",
			Type:         QuestionTypeMultipleChoice,
			Question:     "Which command is used to check if Go is properly installed?",
			Options:      []string{"go version", "go check", "go status", "go info"},
			CorrectAnswer: "go version",
			Points:       5,
			Difficulty:   "easy",
			LearningObj:  "installation_setup",
			Explanation:  "The 'go version' command displays the installed Go version.",
		},
		{
			ID:           "q2-p2p-networking",
			Type:         QuestionTypeMultipleChoice,
			Question:     "What format does Ollama Distributed use for P2P network addresses?",
			Options:      []string{"http://", "tcp://", "multiaddr", "ws://"},
			CorrectAnswer: "multiaddr",
			Points:       10,
			Difficulty:   "medium",
			LearningObj:  "cluster_operations",
			Explanation:  "Ollama Distributed uses multiaddr format like '/ip4/127.0.0.1/tcp/4001'.",
		},
		{
			ID:           "q3-api-endpoints",
			Type:         QuestionTypeShortAnswer,
			Question:     "Name three API endpoints that are currently functional in Ollama Distributed.",
			CorrectAnswer: []string{"/health", "/api/v1/nodes", "/api/v1/stats"},
			Points:       15,
			Difficulty:   "medium",
			LearningObj:  "api_integration",
			Explanation:  "These endpoints provide health status, node information, and system statistics.",
		},
		{
			ID:           "q4-configuration",
			Type:         QuestionTypeScenario,
			Question:     "You need to run two Ollama Distributed nodes on the same machine. What configuration changes are required?",
			CorrectAnswer: "Different ports for API and P2P listeners, separate configuration files",
			Points:       20,
			Difficulty:   "hard",
			LearningObj:  "configuration_management",
			Explanation:  "Each node needs unique ports to avoid conflicts.",
		},
		{
			ID:           "q5-troubleshooting",
			Type:         QuestionTypeTrueFalse,
			Question:     "If the API server fails to start due to port conflicts, you should modify the 'listen' setting in the configuration.",
			CorrectAnswer: true,
			Points:       10,
			Difficulty:   "medium",
			LearningObj:  "troubleshooting",
			Explanation:  "Changing the listen port resolves port conflict issues.",
		},
	}
}

func createPracticalTasks() []PracticalTask {
	return []PracticalTask{
		{
			ID:          "task-installation",
			Name:        "Installation and Setup",
			Description: "Install and configure Ollama Distributed",
			Instructions: []string{
				"Clone the repository",
				"Build the application",
				"Create a basic configuration file",
				"Verify the installation works",
			},
			ExpectedOutputs: []ExpectedOutput{
				{
					Type:    "file",
					Target:  "config.yaml",
					Content: "api:",
					Validation: []string{"contains_api_config", "valid_yaml"},
				},
				{
					Type:    "command_output",
					Target:  "health_check",
					Pattern: `"status":"healthy"`,
					Validation: []string{"json_format", "health_status_ok"},
				},
			},
			TimeLimit: 15 * time.Minute,
			Points:    100,
		},
		{
			ID:          "task-api-integration",
			Name:        "API Integration",
			Description: "Create a simple monitoring tool using the API",
			Instructions: []string{
				"Write a script that calls health endpoint",
				"Parse the JSON response",
				"Display the status in a user-friendly format",
				"Add error handling for connection issues",
			},
			ExpectedOutputs: []ExpectedOutput{
				{
					Type:    "file",
					Target:  "monitor.sh",
					Pattern: "curl.*health",
					Validation: []string{"executable", "contains_curl", "has_error_handling"},
				},
			},
			TimeLimit: 10 * time.Minute,
			Points:    75,
		},
	}
}

func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:16]
}

func evaluateSecurityEvent(event SecurityEvent) string {
	switch event.Type {
	case "tab_switch":
		return "warning"
	case "copy_paste":
		return "high"
	case "network_request":
		return "critical"
	default:
		return "info"
	}
}

func generateCertificate(result *CertificationResult) (*Certificate, error) {
	if !result.Passed {
		return nil, fmt.Errorf("cannot generate certificate for failed assessment")
	}

	cert := &Certificate{
		ID:          "cert-" + generateID(),
		CandidateID: result.CandidateID,
		IssuedDate:  time.Now(),
		ExpiryDate:  time.Now().Add(2 * 365 * 24 * time.Hour), // 2 years
		Data: map[string]string{
			"certification": result.CertificationID,
			"score":        fmt.Sprintf("%.2f", result.OverallScore),
			"date":         result.EndTime.Format("2006-01-02"),
		},
	}

	// Generate hash for integrity
	dataJSON, _ := json.Marshal(cert.Data)
	hash := sha256.Sum256(dataJSON)
	cert.Hash = hex.EncodeToString(hash[:])

	// Generate signature (simplified)
	sigData := cert.ID + cert.CandidateID + cert.Hash
	sigHash := sha256.Sum256([]byte(sigData))
	cert.Signature = hex.EncodeToString(sigHash[:])

	return cert, nil
}

func validateCertificate(cert *Certificate) (bool, error) {
	// Verify hash integrity
	dataJSON, _ := json.Marshal(cert.Data)
	expectedHash := sha256.Sum256(dataJSON)
	if cert.Hash != hex.EncodeToString(expectedHash[:]) {
		return false, nil // Hash mismatch - tampered
	}

	// Verify signature
	sigData := cert.ID + cert.CandidateID + cert.Hash
	expectedSig := sha256.Sum256([]byte(sigData))
	if cert.Signature != hex.EncodeToString(expectedSig[:]) {
		return false, nil // Signature mismatch - forged
	}

	// Check expiry
	if time.Now().After(cert.ExpiryDate) {
		return false, fmt.Errorf("certificate expired")
	}

	return true, nil
}

func executeAssessment(session *AssessmentSession, cert *CertificationTest) (*CertificationResult, error) {
	// Simulate assessment execution
	result := &CertificationResult{
		CandidateID:     session.CandidateID,
		CertificationID: cert.CertificationID,
		StartTime:       session.StartTime,
		EndTime:         time.Now(),
		Duration:        time.Since(session.StartTime),
		Scores: map[string]float64{
			"knowledge":   0.82,
			"practical":   0.88,
			"integration": 0.75,
		},
	}

	// Calculate overall score
	weights := map[string]float64{
		"knowledge":   0.40,
		"practical":   0.40,
		"integration": 0.20,
	}
	result.OverallScore = calculateOverallScore(result.Scores, weights)
	result.Passed = evaluatePassingCriteria(result, cert.PassingCriteria)

	return result, nil
}

func calculateOverallScore(scores map[string]float64, weights map[string]float64) float64 {
	total := 0.0
	for assessment, score := range scores {
		if weight, exists := weights[assessment]; exists {
			total += score * weight
		}
	}
	return total
}

func evaluatePassingCriteria(result *CertificationResult, criteria PassingCriteria) bool {
	// Check overall score
	if result.OverallScore < criteria.OverallScore {
		return false
	}

	// Check individual minimum scores
	for assessment, minScore := range criteria.MinimumScores {
		if score, exists := result.Scores[assessment]; !exists || score < minScore {
			return false
		}
	}

	return true
}

// BenchmarkCertificationPerformance benchmarks certification system performance
func BenchmarkCertificationPerformance(b *testing.B) {
	cert := createOllamaDistributedCertification()

	b.Run("QuestionEvaluation", func(b *testing.B) {
		questions := createKnowledgeQuestions()
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			for _, q := range questions {
				// Simulate question evaluation
				_ = evaluateQuestion(q, q.CorrectAnswer)
			}
		}
	})

	b.Run("CertificateGeneration", func(b *testing.B) {
		result := &CertificationResult{
			CandidateID:     "bench-candidate",
			CertificationID: cert.CertificationID,
			OverallScore:    0.85,
			Passed:          true,
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := generateCertificate(result)
			if err != nil {
				b.Error(err)
			}
		}
	})
}

func evaluateQuestion(q Question, answer interface{}) bool {
	// Simplified question evaluation
	return fmt.Sprintf("%v", answer) == fmt.Sprintf("%v", q.CorrectAnswer)
}
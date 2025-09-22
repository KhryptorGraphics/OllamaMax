package orchestration

import (
	"fmt"
	"math"
	"time"
	"sync"
)

// AggregationSession represents the lifecycle of a distributed aggregation operation
type AggregationSession struct {
	ID                 string
	Type               string
	StartTime          time.Time
	LastUpdate         time.Time
	Status             AggregationStatus
	ExpectedPartitions int
	ReceivedPartitions int
	PartialResults     []*PartialResult
	ValidationResults  map[string]*ValidationResult
	Metadata           map[string]interface{}
	Timeouts           *TimeoutConfiguration
	mu                 sync.RWMutex
}

// AggregationStatus represents the status of an aggregation session
type AggregationStatus string

const (
	SessionPending    AggregationStatus = "pending"
	SessionActive     AggregationStatus = "active"
	SessionCompleted  AggregationStatus = "completed"
	SessionFailed     AggregationStatus = "failed"
	SessionTimedOut   AggregationStatus = "timed_out"
	SessionCancelled  AggregationStatus = "cancelled"
)

// TimeoutConfiguration defines timeout settings for aggregation operations
type TimeoutConfiguration struct {
	SessionTimeout    time.Duration // Total session timeout
	PartitionTimeout  time.Duration // Timeout for individual partitions
	ValidationTimeout time.Duration // Timeout for validation operations
	CleanupTimeout    time.Duration // Timeout for cleanup operations
}

// ValidationResult contains the result of validating a partial result
type ValidationResult struct {
	PartitionID    string
	IsValid        bool
	TensorShape    []int
	DataSize       int
	Checksum       string
	ValidationTime time.Time
	Errors         []string
	Warnings       []string
	Metadata       map[string]interface{}
}

// AggregationContextManager manages aggregation sessions and context
type AggregationContextManager struct {
	sessions         map[string]*AggregationSession
	partitionTracker *PartitionResultTracker
	metricsCollector *AggregationMetricsCollector
	mu               sync.RWMutex
}

func NewAggregationContextManager() *AggregationContextManager {
	return &AggregationContextManager{
		sessions:         make(map[string]*AggregationSession),
		partitionTracker: NewPartitionResultTracker(),
		metricsCollector: NewAggregationMetricsCollector(),
	}
}

// CreateSession creates a new aggregation session
func (acm *AggregationContextManager) CreateSession(
	sessionID string,
	sessionType string,
	expectedPartitions int,
	timeouts *TimeoutConfiguration,
) (*AggregationSession, error) {

	acm.mu.Lock()
	defer acm.mu.Unlock()

	// Check if session already exists
	if _, exists := acm.sessions[sessionID]; exists {
		return nil, fmt.Errorf("session %s already exists", sessionID)
	}

	// Set default timeouts if not provided
	if timeouts == nil {
		timeouts = &TimeoutConfiguration{
			SessionTimeout:    30 * time.Minute,
			PartitionTimeout:  5 * time.Minute,
			ValidationTimeout: 30 * time.Second,
			CleanupTimeout:    1 * time.Minute,
		}
	}

	session := &AggregationSession{
		ID:                 sessionID,
		Type:               sessionType,
		StartTime:          time.Now(),
		LastUpdate:         time.Now(),
		Status:             SessionPending,
		ExpectedPartitions: expectedPartitions,
		ReceivedPartitions: 0,
		PartialResults:     make([]*PartialResult, 0),
		ValidationResults:  make(map[string]*ValidationResult),
		Metadata:           make(map[string]interface{}),
		Timeouts:           timeouts,
	}

	acm.sessions[sessionID] = session

	// Start session monitoring
	acm.metricsCollector.StartSession(sessionID)

	return session, nil
}

// AddPartialResult adds a partial result to an aggregation session
func (acm *AggregationContextManager) AddPartialResult(
	sessionID string,
	partialResult *PartialResult,
) error {

	acm.mu.Lock()
	defer acm.mu.Unlock()

	session, exists := acm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	// Check session status
	if session.Status != SessionPending && session.Status != SessionActive {
		return fmt.Errorf("cannot add result to session %s in status %s", sessionID, session.Status)
	}

	// Validate partial result
	validation, err := acm.validatePartialResult(partialResult, session)
	if err != nil {
		return fmt.Errorf("partial result validation failed: %v", err)
	}

	session.ValidationResults[partialResult.PartitionID] = validation

	if !validation.IsValid {
		return fmt.Errorf("partial result from partition %s is invalid", partialResult.PartitionID)
	}

	// Add to session
	session.PartialResults = append(session.PartialResults, partialResult)
	session.ReceivedPartitions++
	session.LastUpdate = time.Now()

	// Update session status
	if session.Status == SessionPending {
		session.Status = SessionActive
	}

	// Check if all partitions received
	if session.ReceivedPartitions >= session.ExpectedPartitions {
		session.Status = SessionCompleted
		acm.metricsCollector.CompleteSession(sessionID)
	}

	// Track partition result
	acm.partitionTracker.AddResult(partialResult)

	return nil
}

// GetSession retrieves an aggregation session
func (acm *AggregationContextManager) GetSession(sessionID string) (*AggregationSession, error) {
	acm.mu.RLock()
	defer acm.mu.RUnlock()

	session, exists := acm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	return session, nil
}

// GetSessionProgress returns progress information for a session
func (acm *AggregationContextManager) GetSessionProgress(sessionID string) (*SessionProgress, error) {
	acm.mu.RLock()
	defer acm.mu.RUnlock()

	session, exists := acm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	session.mu.RLock()
	defer session.mu.RUnlock()

	progress := float32(session.ReceivedPartitions) / float32(session.ExpectedPartitions)
	if progress > 1.0 {
		progress = 1.0
	}

	elapsedTime := time.Since(session.StartTime)
	var estimatedTimeRemaining time.Duration
	if progress > 0 {
		totalEstimatedTime := time.Duration(float64(elapsedTime) / float64(progress))
		estimatedTimeRemaining = totalEstimatedTime - elapsedTime
		if estimatedTimeRemaining < 0 {
			estimatedTimeRemaining = 0
		}
	}

	return &SessionProgress{
		SessionID:               sessionID,
		Status:                  session.Status,
		Progress:                progress,
		ReceivedPartitions:      session.ReceivedPartitions,
		ExpectedPartitions:      session.ExpectedPartitions,
		ElapsedTime:             elapsedTime,
		EstimatedTimeRemaining:  estimatedTimeRemaining,
		ValidationErrors:        acm.getValidationErrors(session),
		LastUpdate:              session.LastUpdate,
	}, nil
}

// SessionProgress contains progress information for an aggregation session
type SessionProgress struct {
	SessionID              string
	Status                 AggregationStatus
	Progress               float32
	ReceivedPartitions     int
	ExpectedPartitions     int
	ElapsedTime            time.Duration
	EstimatedTimeRemaining time.Duration
	ValidationErrors       []string
	LastUpdate             time.Time
}

// CleanupSession removes a completed or failed session
func (acm *AggregationContextManager) CleanupSession(sessionID string) error {
	acm.mu.Lock()
	defer acm.mu.Unlock()

	session, exists := acm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	// Only cleanup completed, failed, or timed out sessions
	if session.Status != SessionCompleted && session.Status != SessionFailed && session.Status != SessionTimedOut {
		return fmt.Errorf("cannot cleanup session %s in status %s", sessionID, session.Status)
	}

	// Cleanup partition results
	for _, result := range session.PartialResults {
		acm.partitionTracker.RemoveResult(result.PartitionID)
	}

	// Cleanup metrics
	acm.metricsCollector.CleanupSession(sessionID)

	// Remove session
	delete(acm.sessions, sessionID)

	return nil
}

// MonitorTimeouts checks for sessions that have timed out
func (acm *AggregationContextManager) MonitorTimeouts() {
	acm.mu.Lock()
	defer acm.mu.Unlock()

	now := time.Now()
	for sessionID, session := range acm.sessions {
		session.mu.Lock()

		// Check session timeout
		if session.Status == SessionActive || session.Status == SessionPending {
			if now.Sub(session.StartTime) > session.Timeouts.SessionTimeout {
				session.Status = SessionTimedOut
				acm.metricsCollector.TimeoutSession(sessionID)
			}
		}

		session.mu.Unlock()
	}
}

// PartitionResultTracker monitors partial results from different nodes
type PartitionResultTracker struct {
	results      map[string]*PartialResult
	nodeResults  map[string][]*PartialResult
	statusCounts map[string]int
	mu           sync.RWMutex
}

func NewPartitionResultTracker() *PartitionResultTracker {
	return &PartitionResultTracker{
		results:      make(map[string]*PartialResult),
		nodeResults:  make(map[string][]*PartialResult),
		statusCounts: make(map[string]int),
	}
}

// AddResult tracks a new partial result
func (prt *PartitionResultTracker) AddResult(result *PartialResult) {
	prt.mu.Lock()
	defer prt.mu.Unlock()

	prt.results[result.PartitionID] = result

	// Track by node
	nodeID := result.PartitionID // Simplified - should extract actual node ID
	if prt.nodeResults[nodeID] == nil {
		prt.nodeResults[nodeID] = make([]*PartialResult, 0)
	}
	prt.nodeResults[nodeID] = append(prt.nodeResults[nodeID], result)

	// Update status counts
	status := "success"
	if result.Error != "" {
		status = "error"
	}
	prt.statusCounts[status]++
}

// RemoveResult removes a tracked result
func (prt *PartitionResultTracker) RemoveResult(partitionID string) {
	prt.mu.Lock()
	defer prt.mu.Unlock()

	if result, exists := prt.results[partitionID]; exists {
		// Update status counts
		status := "success"
		if result.Error != "" {
			status = "error"
		}
		prt.statusCounts[status]--

		delete(prt.results, partitionID)
	}
}

// GetNodeResults returns all results for a specific node
func (prt *PartitionResultTracker) GetNodeResults(nodeID string) []*PartialResult {
	prt.mu.RLock()
	defer prt.mu.RUnlock()

	results := prt.nodeResults[nodeID]
	if results == nil {
		return make([]*PartialResult, 0)
	}

	// Return copy to avoid race conditions
	copy := make([]*PartialResult, len(results))
	for i, result := range results {
		copy[i] = result
	}

	return copy
}

// GetStatistics returns tracking statistics
func (prt *PartitionResultTracker) GetStatistics() *PartitionStatistics {
	prt.mu.RLock()
	defer prt.mu.RUnlock()

	return &PartitionStatistics{
		TotalResults:    len(prt.results),
		SuccessfulResults: prt.statusCounts["success"],
		FailedResults:     prt.statusCounts["error"],
		NodeCount:         len(prt.nodeResults),
	}
}

// PartitionStatistics contains statistics about partition results
type PartitionStatistics struct {
	TotalResults      int
	SuccessfulResults int
	FailedResults     int
	NodeCount         int
}

// AggregationMetricsCollector collects performance and quality metrics
type AggregationMetricsCollector struct {
	sessionMetrics map[string]*SessionMetrics
	mu             sync.RWMutex
}

func NewAggregationMetricsCollector() *AggregationMetricsCollector {
	return &AggregationMetricsCollector{
		sessionMetrics: make(map[string]*SessionMetrics),
	}
}

// SessionMetrics contains metrics for an aggregation session
type SessionMetrics struct {
	SessionID              string
	StartTime              time.Time
	EndTime                time.Time
	TotalLatency          time.Duration
	AggregationLatency    time.Duration
	ValidationLatency     time.Duration
	TensorProcessingThroughput float64
	MemoryUsage           int64
	AccuracyMetrics       map[string]float32
	QualityScores         map[string]float32
	ErrorCount            int
	WarningCount          int
	Metadata              map[string]interface{}
}

// StartSession starts collecting metrics for a session
func (amc *AggregationMetricsCollector) StartSession(sessionID string) {
	amc.mu.Lock()
	defer amc.mu.Unlock()

	metrics := &SessionMetrics{
		SessionID:       sessionID,
		StartTime:       time.Now(),
		AccuracyMetrics: make(map[string]float32),
		QualityScores:   make(map[string]float32),
		Metadata:        make(map[string]interface{}),
	}

	amc.sessionMetrics[sessionID] = metrics
}

// CompleteSession marks a session as completed and finalizes metrics
func (amc *AggregationMetricsCollector) CompleteSession(sessionID string) {
	amc.mu.Lock()
	defer amc.mu.Unlock()

	if metrics, exists := amc.sessionMetrics[sessionID]; exists {
		metrics.EndTime = time.Now()
		metrics.TotalLatency = metrics.EndTime.Sub(metrics.StartTime)
	}
}

// TimeoutSession marks a session as timed out
func (amc *AggregationMetricsCollector) TimeoutSession(sessionID string) {
	amc.mu.Lock()
	defer amc.mu.Unlock()

	if metrics, exists := amc.sessionMetrics[sessionID]; exists {
		metrics.EndTime = time.Now()
		metrics.TotalLatency = metrics.EndTime.Sub(metrics.StartTime)
		metrics.ErrorCount++
		metrics.Metadata["timeout"] = true
	}
}

// UpdateTensorProcessingMetrics updates tensor processing performance metrics
func (amc *AggregationMetricsCollector) UpdateTensorProcessingMetrics(
	sessionID string,
	throughput float64,
	memoryUsage int64,
) {
	amc.mu.Lock()
	defer amc.mu.Unlock()

	if metrics, exists := amc.sessionMetrics[sessionID]; exists {
		metrics.TensorProcessingThroughput = throughput
		metrics.MemoryUsage = memoryUsage
	}
}

// UpdateAccuracyMetrics updates accuracy metrics for the session
func (amc *AggregationMetricsCollector) UpdateAccuracyMetrics(
	sessionID string,
	accuracyType string,
	score float32,
) {
	amc.mu.Lock()
	defer amc.mu.Unlock()

	if metrics, exists := amc.sessionMetrics[sessionID]; exists {
		metrics.AccuracyMetrics[accuracyType] = score
	}
}

// GetSessionMetrics returns metrics for a specific session
func (amc *AggregationMetricsCollector) GetSessionMetrics(sessionID string) (*SessionMetrics, error) {
	amc.mu.RLock()
	defer amc.mu.RUnlock()

	metrics, exists := amc.sessionMetrics[sessionID]
	if !exists {
		return nil, fmt.Errorf("metrics for session %s not found", sessionID)
	}

	return metrics, nil
}

// GetAggregatedMetrics returns aggregated metrics across all sessions
func (amc *AggregationMetricsCollector) GetAggregatedMetrics() *AggregatedMetrics {
	amc.mu.RLock()
	defer amc.mu.RUnlock()

	var totalSessions int
	var completedSessions int
	var totalLatency time.Duration
	var totalThroughput float64
	var totalMemoryUsage int64
	var totalErrors int

	for _, metrics := range amc.sessionMetrics {
		totalSessions++
		if !metrics.EndTime.IsZero() {
			completedSessions++
			totalLatency += metrics.TotalLatency
		}
		totalThroughput += metrics.TensorProcessingThroughput
		totalMemoryUsage += metrics.MemoryUsage
		totalErrors += metrics.ErrorCount
	}

	var averageLatency time.Duration
	var averageThroughput float64
	var averageMemoryUsage int64

	if completedSessions > 0 {
		averageLatency = totalLatency / time.Duration(completedSessions)
		averageThroughput = totalThroughput / float64(completedSessions)
		averageMemoryUsage = totalMemoryUsage / int64(completedSessions)
	}

	return &AggregatedMetrics{
		TotalSessions:       totalSessions,
		CompletedSessions:   completedSessions,
		AverageLatency:      averageLatency,
		AverageThroughput:   averageThroughput,
		AverageMemoryUsage:  averageMemoryUsage,
		TotalErrors:         totalErrors,
		SuccessRate:         float32(completedSessions) / float32(totalSessions),
	}
}

// AggregatedMetrics contains aggregated metrics across all sessions
type AggregatedMetrics struct {
	TotalSessions       int
	CompletedSessions   int
	AverageLatency      time.Duration
	AverageThroughput   float64
	AverageMemoryUsage  int64
	TotalErrors         int
	SuccessRate         float32
}

// CleanupSession removes metrics for a session
func (amc *AggregationMetricsCollector) CleanupSession(sessionID string) {
	amc.mu.Lock()
	defer amc.mu.Unlock()

	delete(amc.sessionMetrics, sessionID)
}

// Helper methods

func (acm *AggregationContextManager) validatePartialResult(
	result *PartialResult,
	session *AggregationSession,
) (*ValidationResult, error) {

	validation := &ValidationResult{
		PartitionID:    result.PartitionID,
		IsValid:        true,
		ValidationTime: time.Now(),
		Errors:         make([]string, 0),
		Warnings:       make([]string, 0),
		Metadata:       make(map[string]interface{}),
	}

	// Validate tensor shape consistency
	if result.HiddenStates != nil {
		validation.TensorShape = extractShape(result.HiddenStates)
		validation.DataSize = len(result.HiddenStates)

		// Check for reasonable data size
		if validation.DataSize == 0 {
			validation.IsValid = false
			validation.Errors = append(validation.Errors, "empty hidden states")
		}

		// Check for NaN or infinite values
		for i, val := range result.HiddenStates {
			if math.IsNaN(float64(val)) || math.IsInf(float64(val), 0) {
				validation.IsValid = false
				validation.Errors = append(validation.Errors, fmt.Sprintf("invalid value at index %d: %f", i, val))
				break // Only report first invalid value
			}
		}
	}

	// Validate logits if present
	if result.Logits != nil {
		// Check for proper probability distribution properties
		var sum float32
		for _, logit := range result.Logits {
			if math.IsNaN(float64(logit)) || math.IsInf(float64(logit), 0) {
				validation.IsValid = false
				validation.Errors = append(validation.Errors, "invalid logit values")
				break
			}
			sum += logit
		}

		// Check if logits are reasonable (not too extreme)
		if sum == 0 {
			validation.Warnings = append(validation.Warnings, "logits sum to zero")
		}
	}

	// Validate tokens if present
	if result.Tokens != nil {
		for _, token := range result.Tokens {
			if token < 0 {
				validation.Warnings = append(validation.Warnings, "negative token values found")
				break
			}
		}
	}

	// Generate checksum for data integrity
	if result.HiddenStates != nil || result.Logits != nil {
		validation.Checksum = acm.calculateChecksum(result)
	}

	return validation, nil
}

func (acm *AggregationContextManager) calculateChecksum(result *PartialResult) string {
	// Simple checksum calculation - in practice would use proper hashing
	var sum float64
	if result.HiddenStates != nil {
		for _, val := range result.HiddenStates {
			sum += float64(val)
		}
	}
	if result.Logits != nil {
		for _, val := range result.Logits {
			sum += float64(val)
		}
	}

	return fmt.Sprintf("sum_%.6f", sum)
}

func (acm *AggregationContextManager) getValidationErrors(session *AggregationSession) []string {
	errors := make([]string, 0)
	for _, validation := range session.ValidationResults {
		if !validation.IsValid {
			errors = append(errors, validation.Errors...)
		}
	}
	return errors
}

// Enhanced AggregationContext that integrates with the existing system
func (acm *AggregationContextManager) CreateEnhancedContext(
	taskID string,
	partialResults []*PartialResult,
	metadata map[string]interface{},
) (*AggregationContext, error) {

	// Create session for tracking
	session, err := acm.CreateSession(
		taskID,
		"enhanced_aggregation",
		len(partialResults),
		nil, // Use default timeouts
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	// Add all partial results to session
	for _, result := range partialResults {
		if err := acm.AddPartialResult(taskID, result); err != nil {
			return nil, fmt.Errorf("failed to add partial result: %v", err)
		}
	}

	// Create enhanced aggregation context
	context := &AggregationContext{
		TaskID:         taskID,
		PartialResults: partialResults,
		Metadata:       metadata,
	}

	// Add session tracking information
	if context.Metadata == nil {
		context.Metadata = make(map[string]interface{})
	}
	context.Metadata["session_id"] = session.ID
	context.Metadata["tensor_aware"] = true
	context.Metadata["validation_enabled"] = true

	return context, nil
}
package consensus

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/hashicorp/raft"
)

// ConflictResolver manages advanced conflict resolution strategies
type ConflictResolver struct {
	engine *Engine
	mu     sync.RWMutex

	// Conflict tracking
	conflicts       map[string]*Conflict
	resolutionRules []*ResolutionRule

	// Configuration
	config *ConflictConfig

	// Metrics
	metrics *ConflictMetrics

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// Conflict represents a state conflict between nodes
type Conflict struct {
	ID           string       `json:"id"`
	Key          string       `json:"key"`
	ConflictType ConflictType `json:"conflict_type"`

	// Conflicting values
	Values []*ConflictValue `json:"values"`

	// Resolution information
	ResolutionStrategy ResolutionStrategy `json:"resolution_strategy"`
	ResolvedValue      interface{}        `json:"resolved_value,omitempty"`
	ResolvedBy         raft.ServerID      `json:"resolved_by,omitempty"`

	// Timestamps
	DetectedAt time.Time `json:"detected_at"`
	ResolvedAt time.Time `json:"resolved_at,omitempty"`

	// Status
	Status   ConflictStatus   `json:"status"`
	Priority ConflictPriority `json:"priority"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata"`
}

// ConflictValue represents a conflicting value from a specific node
type ConflictValue struct {
	Value      interface{}            `json:"value"`
	NodeID     raft.ServerID          `json:"node_id"`
	Timestamp  time.Time              `json:"timestamp"`
	Version    int64                  `json:"version"`
	Confidence float64                `json:"confidence"` // 0.0 to 1.0
	Evidence   map[string]interface{} `json:"evidence"`
}

// ConflictType represents the type of conflict
type ConflictType string

const (
	ConflictTypeValue      ConflictType = "value"
	ConflictTypeVersion    ConflictType = "version"
	ConflictTypeTimestamp  ConflictType = "timestamp"
	ConflictTypeStructural ConflictType = "structural"
	ConflictTypePermission ConflictType = "permission"
)

// ResolutionStrategy represents the strategy used to resolve conflicts
type ResolutionStrategy string

const (
	StrategyLastWriteWins     ResolutionStrategy = "last_write_wins"
	StrategyHighestVersion    ResolutionStrategy = "highest_version"
	StrategyMajorityVote      ResolutionStrategy = "majority_vote"
	StrategyHighestConfidence ResolutionStrategy = "highest_confidence"
	StrategyCustomRule        ResolutionStrategy = "custom_rule"
	StrategyManualReview      ResolutionStrategy = "manual_review"
)

// ConflictStatus represents the status of conflict resolution
type ConflictStatus string

const (
	ConflictStatusDetected  ConflictStatus = "detected"
	ConflictStatusAnalyzing ConflictStatus = "analyzing"
	ConflictStatusResolving ConflictStatus = "resolving"
	ConflictStatusResolved  ConflictStatus = "resolved"
	ConflictStatusFailed    ConflictStatus = "failed"
	ConflictStatusEscalated ConflictStatus = "escalated"
)

// ConflictPriority represents the priority of conflict resolution
type ConflictPriority string

const (
	PriorityLow      ConflictPriority = "low"
	PriorityNormal   ConflictPriority = "normal"
	PriorityHigh     ConflictPriority = "high"
	PriorityCritical ConflictPriority = "critical"
)

// ResolutionRule defines a rule for resolving specific types of conflicts
type ResolutionRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`

	// Conditions
	KeyPattern   string       `json:"key_pattern"`
	ConflictType ConflictType `json:"conflict_type"`
	MinValues    int          `json:"min_values"`

	// Resolution
	Strategy    ResolutionStrategy `json:"strategy"`
	Priority    ConflictPriority   `json:"priority"`
	AutoResolve bool               `json:"auto_resolve"`

	// Custom resolution function
	ResolverFunc func(*Conflict) (interface{}, error) `json:"-"`

	// Metadata
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

// ConflictConfig configures conflict resolution
type ConflictConfig struct {
	// Detection settings
	EnableConflictDetection bool
	DetectionInterval       time.Duration
	ConflictTimeout         time.Duration

	// Resolution settings
	DefaultStrategy       ResolutionStrategy
	AutoResolveThreshold  float64 // Confidence threshold for auto-resolution
	MaxResolutionAttempts int

	// Escalation settings
	EscalationTimeout   time.Duration
	RequireManualReview []string // Keys that require manual review

	// Performance settings
	MaxConcurrentResolutions int
	CleanupInterval          time.Duration
	MaxConflictHistory       int
}

// ConflictMetrics tracks conflict resolution performance
type ConflictMetrics struct {
	TotalConflicts        int64                        `json:"total_conflicts"`
	ResolvedConflicts     int64                        `json:"resolved_conflicts"`
	FailedResolutions     int64                        `json:"failed_resolutions"`
	EscalatedConflicts    int64                        `json:"escalated_conflicts"`
	ConflictsByType       map[ConflictType]int64       `json:"conflicts_by_type"`
	ConflictsByStrategy   map[ResolutionStrategy]int64 `json:"conflicts_by_strategy"`
	AverageResolutionTime time.Duration                `json:"average_resolution_time"`
	LastConflict          time.Time                    `json:"last_conflict"`
}

// NewConflictResolver creates a new conflict resolver
func NewConflictResolver(engine *Engine, config *ConflictConfig) *ConflictResolver {
	ctx, cancel := context.WithCancel(context.Background())

	if config == nil {
		config = &ConflictConfig{
			EnableConflictDetection:  true,
			DetectionInterval:        5 * time.Second,
			ConflictTimeout:          30 * time.Second,
			DefaultStrategy:          StrategyLastWriteWins,
			AutoResolveThreshold:     0.8,
			MaxResolutionAttempts:    3,
			EscalationTimeout:        5 * time.Minute,
			RequireManualReview:      []string{},
			MaxConcurrentResolutions: 10,
			CleanupInterval:          10 * time.Minute,
			MaxConflictHistory:       1000,
		}
	}

	cr := &ConflictResolver{
		engine:          engine,
		conflicts:       make(map[string]*Conflict),
		resolutionRules: getDefaultResolutionRules(),
		config:          config,
		metrics: &ConflictMetrics{
			ConflictsByType:     make(map[ConflictType]int64),
			ConflictsByStrategy: make(map[ResolutionStrategy]int64),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Start background tasks
	cr.wg.Add(2)
	go cr.detectionLoop()
	go cr.resolutionLoop()

	return cr
}

// DetectConflict detects and registers a new conflict
func (cr *ConflictResolver) DetectConflict(key string, values []*ConflictValue) *Conflict {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	conflictID := fmt.Sprintf("conflict_%s_%d", key, time.Now().UnixNano())

	conflict := &Conflict{
		ID:           conflictID,
		Key:          key,
		ConflictType: cr.determineConflictType(values),
		Values:       values,
		DetectedAt:   time.Now(),
		Status:       ConflictStatusDetected,
		Priority:     cr.determinePriority(key, values),
		Metadata:     make(map[string]interface{}),
	}

	// Determine resolution strategy
	conflict.ResolutionStrategy = cr.determineResolutionStrategy(conflict)

	cr.conflicts[conflictID] = conflict
	cr.metrics.TotalConflicts++
	cr.metrics.ConflictsByType[conflict.ConflictType]++
	cr.metrics.LastConflict = time.Now()

	return conflict
}

// ResolveConflict resolves a specific conflict
func (cr *ConflictResolver) ResolveConflict(conflictID string) error {
	cr.mu.Lock()
	conflict, exists := cr.conflicts[conflictID]
	if !exists {
		cr.mu.Unlock()
		return fmt.Errorf("conflict not found: %s", conflictID)
	}

	if conflict.Status != ConflictStatusDetected {
		cr.mu.Unlock()
		return fmt.Errorf("conflict already being processed: %s", conflictID)
	}

	conflict.Status = ConflictStatusResolving
	cr.mu.Unlock()

	startTime := time.Now()
	resolvedValue, err := cr.applyResolutionStrategy(conflict)

	cr.mu.Lock()
	defer cr.mu.Unlock()

	if err != nil {
		conflict.Status = ConflictStatusFailed
		cr.metrics.FailedResolutions++
		return fmt.Errorf("failed to resolve conflict: %w", err)
	}

	// Apply the resolved value
	if err := cr.applyResolvedValue(conflict.Key, resolvedValue); err != nil {
		conflict.Status = ConflictStatusFailed
		cr.metrics.FailedResolutions++
		return fmt.Errorf("failed to apply resolved value: %w", err)
	}

	conflict.ResolvedValue = resolvedValue
	conflict.ResolvedAt = time.Now()
	conflict.Status = ConflictStatusResolved
	conflict.ResolvedBy = raft.ServerID(cr.engine.GetNodeID())

	// Update metrics
	cr.metrics.ResolvedConflicts++
	cr.metrics.ConflictsByStrategy[conflict.ResolutionStrategy]++

	duration := time.Since(startTime)
	totalTime := time.Duration(cr.metrics.ResolvedConflicts-1)*cr.metrics.AverageResolutionTime + duration
	cr.metrics.AverageResolutionTime = totalTime / time.Duration(cr.metrics.ResolvedConflicts)

	return nil
}

// applyResolutionStrategy applies the appropriate resolution strategy
func (cr *ConflictResolver) applyResolutionStrategy(conflict *Conflict) (interface{}, error) {
	switch conflict.ResolutionStrategy {
	case StrategyLastWriteWins:
		return cr.resolveLastWriteWins(conflict)
	case StrategyHighestVersion:
		return cr.resolveHighestVersion(conflict)
	case StrategyMajorityVote:
		return cr.resolveMajorityVote(conflict)
	case StrategyHighestConfidence:
		return cr.resolveHighestConfidence(conflict)
	case StrategyCustomRule:
		return cr.resolveCustomRule(conflict)
	default:
		return nil, fmt.Errorf("unknown resolution strategy: %s", conflict.ResolutionStrategy)
	}
}

// resolveLastWriteWins resolves conflict using last write wins strategy
func (cr *ConflictResolver) resolveLastWriteWins(conflict *Conflict) (interface{}, error) {
	if len(conflict.Values) == 0 {
		return nil, fmt.Errorf("no values to resolve")
	}

	// Find the value with the latest timestamp
	latestValue := conflict.Values[0]
	for _, value := range conflict.Values[1:] {
		if value.Timestamp.After(latestValue.Timestamp) {
			latestValue = value
		}
	}

	return latestValue.Value, nil
}

// resolveHighestVersion resolves conflict using highest version strategy
func (cr *ConflictResolver) resolveHighestVersion(conflict *Conflict) (interface{}, error) {
	if len(conflict.Values) == 0 {
		return nil, fmt.Errorf("no values to resolve")
	}

	// Find the value with the highest version
	highestValue := conflict.Values[0]
	for _, value := range conflict.Values[1:] {
		if value.Version > highestValue.Version {
			highestValue = value
		}
	}

	return highestValue.Value, nil
}

// resolveMajorityVote resolves conflict using majority vote strategy
func (cr *ConflictResolver) resolveMajorityVote(conflict *Conflict) (interface{}, error) {
	if len(conflict.Values) == 0 {
		return nil, fmt.Errorf("no values to resolve")
	}

	// Count votes for each unique value
	votes := make(map[string]int)
	valueMap := make(map[string]interface{})

	for _, value := range conflict.Values {
		valueJSON, _ := json.Marshal(value.Value)
		valueStr := string(valueJSON)
		votes[valueStr]++
		valueMap[valueStr] = value.Value
	}

	// Find the value with the most votes
	maxVotes := 0
	var winningValue interface{}
	for valueStr, count := range votes {
		if count > maxVotes {
			maxVotes = count
			winningValue = valueMap[valueStr]
		}
	}

	return winningValue, nil
}

// resolveHighestConfidence resolves conflict using highest confidence strategy
func (cr *ConflictResolver) resolveHighestConfidence(conflict *Conflict) (interface{}, error) {
	if len(conflict.Values) == 0 {
		return nil, fmt.Errorf("no values to resolve")
	}

	// Find the value with the highest confidence
	highestValue := conflict.Values[0]
	for _, value := range conflict.Values[1:] {
		if value.Confidence > highestValue.Confidence {
			highestValue = value
		}
	}

	return highestValue.Value, nil
}

// resolveCustomRule resolves conflict using custom rules
func (cr *ConflictResolver) resolveCustomRule(conflict *Conflict) (interface{}, error) {
	// Find applicable custom rule
	for _, rule := range cr.resolutionRules {
		if rule.Enabled && rule.ResolverFunc != nil {
			// Check if rule applies to this conflict
			if cr.ruleApplies(rule, conflict) {
				return rule.ResolverFunc(conflict)
			}
		}
	}

	// Fall back to default strategy
	conflict.ResolutionStrategy = cr.config.DefaultStrategy
	return cr.applyResolutionStrategy(conflict)
}

// ruleApplies checks if a resolution rule applies to a conflict
func (cr *ConflictResolver) ruleApplies(rule *ResolutionRule, conflict *Conflict) bool {
	// Check conflict type
	if rule.ConflictType != "" && rule.ConflictType != conflict.ConflictType {
		return false
	}

	// Check minimum values
	if len(conflict.Values) < rule.MinValues {
		return false
	}

	// Check key pattern (simplified - in reality you'd use regex)
	if rule.KeyPattern != "" && rule.KeyPattern != conflict.Key {
		return false
	}

	return true
}

// applyResolvedValue applies the resolved value to the state
func (cr *ConflictResolver) applyResolvedValue(key string, value interface{}) error {
	// Apply the resolved value through the consensus engine
	return cr.engine.Apply(key, value, map[string]interface{}{
		"resolution": true,
		"timestamp":  time.Now(),
	})
}

// determineConflictType determines the type of conflict based on values
func (cr *ConflictResolver) determineConflictType(values []*ConflictValue) ConflictType {
	if len(values) < 2 {
		return ConflictTypeValue
	}

	// Check if it's a version conflict
	hasVersions := true
	for _, value := range values {
		if value.Version == 0 {
			hasVersions = false
			break
		}
	}

	if hasVersions {
		return ConflictTypeVersion
	}

	// Default to value conflict
	return ConflictTypeValue
}

// determinePriority determines the priority of a conflict
func (cr *ConflictResolver) determinePriority(key string, values []*ConflictValue) ConflictPriority {
	// Check if key requires manual review
	for _, reviewKey := range cr.config.RequireManualReview {
		if key == reviewKey {
			return PriorityCritical
		}
	}

	// Check confidence levels
	totalConfidence := 0.0
	for _, value := range values {
		totalConfidence += value.Confidence
	}
	avgConfidence := totalConfidence / float64(len(values))

	if avgConfidence < 0.3 {
		return PriorityHigh
	} else if avgConfidence < 0.6 {
		return PriorityNormal
	}

	return PriorityLow
}

// determineResolutionStrategy determines the appropriate resolution strategy
func (cr *ConflictResolver) determineResolutionStrategy(conflict *Conflict) ResolutionStrategy {
	// Check for custom rules first
	for _, rule := range cr.resolutionRules {
		if rule.Enabled && cr.ruleApplies(rule, conflict) {
			return rule.Strategy
		}
	}

	// Use default strategy
	return cr.config.DefaultStrategy
}

// GetConflict returns a conflict by ID
func (cr *ConflictResolver) GetConflict(conflictID string) *Conflict {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	if conflict, exists := cr.conflicts[conflictID]; exists {
		// Return a copy
		conflictCopy := *conflict
		return &conflictCopy
	}
	return nil
}

// GetActiveConflicts returns all active conflicts
func (cr *ConflictResolver) GetActiveConflicts() []*Conflict {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	var active []*Conflict
	for _, conflict := range cr.conflicts {
		if conflict.Status != ConflictStatusResolved && conflict.Status != ConflictStatusFailed {
			conflictCopy := *conflict
			active = append(active, &conflictCopy)
		}
	}

	// Sort by priority and detection time
	sort.Slice(active, func(i, j int) bool {
		if active[i].Priority != active[j].Priority {
			return cr.priorityValue(active[i].Priority) > cr.priorityValue(active[j].Priority)
		}
		return active[i].DetectedAt.Before(active[j].DetectedAt)
	})

	return active
}

// priorityValue returns numeric value for priority comparison
func (cr *ConflictResolver) priorityValue(priority ConflictPriority) int {
	switch priority {
	case PriorityCritical:
		return 4
	case PriorityHigh:
		return 3
	case PriorityNormal:
		return 2
	case PriorityLow:
		return 1
	default:
		return 0
	}
}

// GetConflictMetrics returns conflict resolution metrics
func (cr *ConflictResolver) GetConflictMetrics() *ConflictMetrics {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	metrics := *cr.metrics
	return &metrics
}

// detectionLoop periodically detects conflicts
func (cr *ConflictResolver) detectionLoop() {
	defer cr.wg.Done()

	ticker := time.NewTicker(cr.config.DetectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-cr.ctx.Done():
			return
		case <-ticker.C:
			cr.performConflictDetection()
		}
	}
}

// performConflictDetection performs conflict detection
func (cr *ConflictResolver) performConflictDetection() {
	// In a real implementation, you would:
	// 1. Compare state across nodes
	// 2. Detect inconsistencies
	// 3. Create conflict records

	// For now, this is a placeholder
}

// resolutionLoop periodically processes conflict resolutions
func (cr *ConflictResolver) resolutionLoop() {
	defer cr.wg.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-cr.ctx.Done():
			return
		case <-ticker.C:
			cr.processConflictResolutions()
		}
	}
}

// processConflictResolutions processes pending conflict resolutions
func (cr *ConflictResolver) processConflictResolutions() {
	activeConflicts := cr.GetActiveConflicts()

	for _, conflict := range activeConflicts {
		if conflict.Status == ConflictStatusDetected {
			// Auto-resolve if conditions are met
			if cr.shouldAutoResolve(conflict) {
				go cr.ResolveConflict(conflict.ID)
			}
		}
	}
}

// shouldAutoResolve determines if a conflict should be auto-resolved
func (cr *ConflictResolver) shouldAutoResolve(conflict *Conflict) bool {
	// Check if manual review is required
	for _, reviewKey := range cr.config.RequireManualReview {
		if conflict.Key == reviewKey {
			return false
		}
	}

	// Check confidence threshold
	totalConfidence := 0.0
	for _, value := range conflict.Values {
		totalConfidence += value.Confidence
	}
	avgConfidence := totalConfidence / float64(len(conflict.Values))

	return avgConfidence >= cr.config.AutoResolveThreshold
}

// getDefaultResolutionRules returns default resolution rules
func getDefaultResolutionRules() []*ResolutionRule {
	return []*ResolutionRule{
		{
			ID:           "default_last_write_wins",
			Name:         "Default Last Write Wins",
			Description:  "Use last write wins for most conflicts",
			ConflictType: ConflictTypeValue,
			Strategy:     StrategyLastWriteWins,
			Priority:     PriorityNormal,
			AutoResolve:  true,
			Enabled:      true,
			CreatedAt:    time.Now(),
		},
	}
}

// Close closes the conflict resolver
func (cr *ConflictResolver) Close() error {
	cr.cancel()
	cr.wg.Wait()
	return nil
}

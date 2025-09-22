package fault_tolerance

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// DynamicRepartitioningManager handles dynamic repartitioning during node failures
type DynamicRepartitioningManager struct {
	mu                    sync.RWMutex
	activeRepartitionings map[string]*RepartitioningOperation
	partitionManager      PartitionManager
	shardManager          ModelShardManager
	p2pTransfer           P2PTransferProtocol
	inferenceEngine       InferenceEngine
	config                RepartitioningConfig
	metrics               *RepartitioningMetrics
	strategies            map[string]RepartitioningStrategy
}

// RepartitioningOperation represents an active repartitioning operation
type RepartitioningOperation struct {
	ID               string                   `json:"id"`
	SessionID        string                   `json:"session_id"`
	Reason           string                   `json:"reason"`
	StartTime        time.Time                `json:"start_time"`
	EndTime          time.Time                `json:"end_time"`
	Status           string                   `json:"status"`
	FailedNodes      []string                 `json:"failed_nodes"`
	AvailableNodes   []string                 `json:"available_nodes"`
	OriginalPlan     *PartitionPlan           `json:"original_plan"`
	EmergencyPlan    *PartitionPlan           `json:"emergency_plan"`
	MigrationTasks   []MigrationTask          `json:"migration_tasks"`
	QualityImpact    float64                  `json:"quality_impact"`
	PerformanceImpact float64                  `json:"performance_impact"`
	Metadata         map[string]interface{}   `json:"metadata"`
}

// PartitionPlan represents a model partitioning plan
type PartitionPlan struct {
	ID           string                 `json:"id"`
	Strategy     string                 `json:"strategy"`
	Partitions   []ModelPartition       `json:"partitions"`
	NodeMapping  map[string][]string    `json:"node_mapping"`
	ResourceUsage map[string]ResourceUsage `json:"resource_usage"`
	QualityScore float64                `json:"quality_score"`
	Feasible     bool                   `json:"feasible"`
}

// ModelPartition represents a single model partition
type ModelPartition struct {
	ID            string   `json:"id"`
	NodeID        string   `json:"node_id"`
	ModelLayers   []string `json:"model_layers"`
	MemoryRequired int64   `json:"memory_required"`
	ComputeRequired float64 `json:"compute_required"`
	Priority      int      `json:"priority"`
	Replicas      []string `json:"replicas"`
}

// MigrationTask represents a shard migration task
type MigrationTask struct {
	ID           string    `json:"id"`
	ShardID      string    `json:"shard_id"`
	SourceNode   string    `json:"source_node"`
	TargetNode   string    `json:"target_node"`
	Status       string    `json:"status"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	DataSize     int64     `json:"data_size"`
	TransferRate float64   `json:"transfer_rate"`
	Retries      int       `json:"retries"`
}

// ResourceUsage tracks resource usage for a node
type ResourceUsage struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float64 `json:"memory_percent"`
	GPUPercent    float64 `json:"gpu_percent"`
	BandwidthMbps float64 `json:"bandwidth_mbps"`
}

// RepartitioningConfig contains configuration for repartitioning
type RepartitioningConfig struct {
	Strategy              string        `json:"strategy"`
	MaxMigrationTime      time.Duration `json:"max_migration_time"`
	MinQualityThreshold   float64       `json:"min_quality_threshold"`
	EnableParallelMigration bool        `json:"enable_parallel_migration"`
	MaxParallelMigrations int           `json:"max_parallel_migrations"`
	RetryAttempts         int           `json:"retry_attempts"`
	ValidationEnabled     bool          `json:"validation_enabled"`
}

// RepartitioningStrategy interface for different repartitioning strategies
type RepartitioningStrategy interface {
	CreateEmergencyPlan(ctx context.Context, op *RepartitioningOperation) (*PartitionPlan, error)
	ValidatePlan(plan *PartitionPlan) error
	EstimateQualityImpact(plan *PartitionPlan) float64
	EstimatePerformanceImpact(plan *PartitionPlan) float64
}

// ConservativeStrategy minimizes changes to existing partitioning
type ConservativeStrategy struct {
	manager *DynamicRepartitioningManager
}

// AggressiveStrategy optimally redistributes model across all nodes
type AggressiveStrategy struct {
	manager *DynamicRepartitioningManager
}

// AdaptiveStrategy adjusts based on current system load
type AdaptiveStrategy struct {
	manager *DynamicRepartitioningManager
}

// RepartitioningMetrics tracks repartitioning metrics
type RepartitioningMetrics struct {
	mu                     sync.RWMutex
	TotalRepartitionings   int64
	SuccessfulRepartitionings int64
	FailedRepartitionings  int64
	AverageMigrationTime   time.Duration
	TotalDataMigrated      int64
	AverageQualityImpact   float64
	CurrentActiveMigrations int
}

// PartitionManager interface for partition management
type PartitionManager interface {
	GetCurrentPartitions(sessionID string) ([]ModelPartition, error)
	UpdatePartitions(sessionID string, partitions []ModelPartition) error
	ValidatePartitions(partitions []ModelPartition) error
}

// ModelShardManager interface for shard management
type ModelShardManager interface {
	GetShardLocation(shardID string) (string, error)
	MigrateShard(ctx context.Context, shardID, sourceNode, targetNode string) error
	GetShardInfo(shardID string) (*ShardInfo, error)
}

// P2PTransferProtocol interface for P2P transfers
type P2PTransferProtocol interface {
	TransferData(ctx context.Context, sourceNode, targetNode string, data []byte) error
	GetTransferStatus(transferID string) (*TransferStatus, error)
}

// InferenceEngine interface for inference operations
type InferenceEngine interface {
	PauseInference(sessionID string) error
	ResumeInference(sessionID string, plan *PartitionPlan) error
	UpdatePartitionPlan(sessionID string, plan *PartitionPlan) error
}

// ShardInfo contains information about a model shard
type ShardInfo struct {
	ID       string `json:"id"`
	Size     int64  `json:"size"`
	Location string `json:"location"`
	Replicas []string `json:"replicas"`
}

// TransferStatus represents the status of a data transfer
type TransferStatus struct {
	ID           string    `json:"id"`
	Status       string    `json:"status"`
	Progress     float64   `json:"progress"`
	TransferRate float64   `json:"transfer_rate"`
	StartTime    time.Time `json:"start_time"`
	EstimatedCompletion time.Time `json:"estimated_completion"`
}

// NewDynamicRepartitioningManager creates a new dynamic repartitioning manager
func NewDynamicRepartitioningManager(
	partitionManager PartitionManager,
	shardManager ModelShardManager,
	p2pTransfer P2PTransferProtocol,
	inferenceEngine InferenceEngine,
	config RepartitioningConfig,
) *DynamicRepartitioningManager {
	manager := &DynamicRepartitioningManager{
		activeRepartitionings: make(map[string]*RepartitioningOperation),
		partitionManager:      partitionManager,
		shardManager:          shardManager,
		p2pTransfer:           p2pTransfer,
		inferenceEngine:       inferenceEngine,
		config:                config,
		metrics:               &RepartitioningMetrics{},
		strategies:            make(map[string]RepartitioningStrategy),
	}

	// Initialize strategies
	manager.strategies["conservative"] = &ConservativeStrategy{manager: manager}
	manager.strategies["aggressive"] = &AggressiveStrategy{manager: manager}
	manager.strategies["adaptive"] = &AdaptiveStrategy{manager: manager}

	return manager
}

// HandleNodeFailure handles node failure during inference
func (m *DynamicRepartitioningManager) HandleNodeFailure(
	ctx context.Context,
	sessionID string,
	failedNodes []string,
	availableNodes []string,
) (*RepartitioningOperation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create repartitioning operation
	op := &RepartitioningOperation{
		ID:             uuid.New().String(),
		SessionID:      sessionID,
		Reason:         "node_failure",
		StartTime:      time.Now(),
		Status:         "initializing",
		FailedNodes:    failedNodes,
		AvailableNodes: availableNodes,
		MigrationTasks: []MigrationTask{},
		Metadata:       make(map[string]interface{}),
	}

	m.activeRepartitionings[op.ID] = op

	// Get current partition plan
	currentPartitions, err := m.partitionManager.GetCurrentPartitions(sessionID)
	if err != nil {
		op.Status = "failed"
		m.metrics.FailedRepartitionings++
		return nil, fmt.Errorf("failed to get current partitions: %w", err)
	}

	op.OriginalPlan = &PartitionPlan{
		ID:         uuid.New().String(),
		Strategy:   m.config.Strategy,
		Partitions: currentPartitions,
	}

	// Create emergency partition plan
	emergencyPlan, err := m.CreateEmergencyPartitionPlan(ctx, op)
	if err != nil {
		op.Status = "failed"
		m.metrics.FailedRepartitionings++
		return nil, fmt.Errorf("failed to create emergency plan: %w", err)
	}

	op.EmergencyPlan = emergencyPlan

	// Validate the new plan
	if m.config.ValidationEnabled {
		if err := m.ValidatePartitionPlan(emergencyPlan); err != nil {
			op.Status = "validation_failed"
			m.metrics.FailedRepartitionings++
			return nil, fmt.Errorf("emergency plan validation failed: %w", err)
		}
	}

	// Execute migration
	if err := m.ExecuteEmergencyRepartitioning(ctx, op); err != nil {
		op.Status = "migration_failed"
		m.metrics.FailedRepartitionings++
		return nil, fmt.Errorf("failed to execute repartitioning: %w", err)
	}

	op.Status = "completed"
	op.EndTime = time.Now()
	m.metrics.SuccessfulRepartitionings++
	m.metrics.AverageMigrationTime = op.EndTime.Sub(op.StartTime)

	return op, nil
}

// CreateEmergencyPartitionPlan creates an emergency partition plan
func (m *DynamicRepartitioningManager) CreateEmergencyPartitionPlan(
	ctx context.Context,
	op *RepartitioningOperation,
) (*PartitionPlan, error) {
	// Select strategy
	strategy, exists := m.strategies[m.config.Strategy]
	if !exists {
		strategy = m.strategies["conservative"]
	}

	// Create emergency plan
	plan, err := strategy.CreateEmergencyPlan(ctx, op)
	if err != nil {
		return nil, fmt.Errorf("strategy failed to create plan: %w", err)
	}

	// Estimate impacts
	plan.QualityScore = 1.0 - strategy.EstimateQualityImpact(plan)
	op.QualityImpact = strategy.EstimateQualityImpact(plan)
	op.PerformanceImpact = strategy.EstimatePerformanceImpact(plan)

	// Check quality threshold
	if plan.QualityScore < m.config.MinQualityThreshold {
		return nil, fmt.Errorf("plan quality score %.2f below threshold %.2f",
			plan.QualityScore, m.config.MinQualityThreshold)
	}

	return plan, nil
}

// ExecuteEmergencyRepartitioning executes the emergency repartitioning
func (m *DynamicRepartitioningManager) ExecuteEmergencyRepartitioning(
	ctx context.Context,
	op *RepartitioningOperation,
) error {
	// Pause inference
	if err := m.inferenceEngine.PauseInference(op.SessionID); err != nil {
		return fmt.Errorf("failed to pause inference: %w", err)
	}

	// Create migration tasks
	migrationTasks, err := m.createMigrationTasks(op)
	if err != nil {
		return fmt.Errorf("failed to create migration tasks: %w", err)
	}

	op.MigrationTasks = migrationTasks

	// Execute migrations
	if m.config.EnableParallelMigration {
		err = m.executeParallelMigrations(ctx, migrationTasks)
	} else {
		err = m.executeSequentialMigrations(ctx, migrationTasks)
	}

	if err != nil {
		// Attempt rollback
		m.rollbackMigrations(ctx, op)
		return fmt.Errorf("migration execution failed: %w", err)
	}

	// Update partition plan in inference engine
	if err := m.inferenceEngine.UpdatePartitionPlan(op.SessionID, op.EmergencyPlan); err != nil {
		return fmt.Errorf("failed to update partition plan: %w", err)
	}

	// Resume inference with new plan
	if err := m.inferenceEngine.ResumeInference(op.SessionID, op.EmergencyPlan); err != nil {
		return fmt.Errorf("failed to resume inference: %w", err)
	}

	return nil
}

// MigrateModelShards migrates model shards from failed nodes to healthy nodes
func (m *DynamicRepartitioningManager) MigrateModelShards(
	ctx context.Context,
	migrations map[string]string,
) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(migrations))

	for shardID, targetNode := range migrations {
		wg.Add(1)
		go func(sID, tNode string) {
			defer wg.Done()

			// Get current shard location
			sourceNode, err := m.shardManager.GetShardLocation(sID)
			if err != nil {
				errChan <- fmt.Errorf("failed to get shard location for %s: %w", sID, err)
				return
			}

			// Migrate shard
			if err := m.shardManager.MigrateShard(ctx, sID, sourceNode, tNode); err != nil {
				errChan <- fmt.Errorf("failed to migrate shard %s: %w", sID, err)
				return
			}

			m.metrics.TotalDataMigrated++
		}(shardID, targetNode)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// ResumeInferenceWithNewPlan resumes inference with a new partition plan
func (m *DynamicRepartitioningManager) ResumeInferenceWithNewPlan(
	ctx context.Context,
	sessionID string,
	plan *PartitionPlan,
) error {
	// Validate plan
	if err := m.ValidatePartitionPlan(plan); err != nil {
		return fmt.Errorf("plan validation failed: %w", err)
	}

	// Update partitions
	if err := m.partitionManager.UpdatePartitions(sessionID, plan.Partitions); err != nil {
		return fmt.Errorf("failed to update partitions: %w", err)
	}

	// Resume inference
	if err := m.inferenceEngine.ResumeInference(sessionID, plan); err != nil {
		return fmt.Errorf("failed to resume inference: %w", err)
	}

	return nil
}

// ValidatePartitionPlan validates a partition plan
func (m *DynamicRepartitioningManager) ValidatePartitionPlan(plan *PartitionPlan) error {
	if plan == nil {
		return fmt.Errorf("partition plan is nil")
	}

	if len(plan.Partitions) == 0 {
		return fmt.Errorf("partition plan has no partitions")
	}

	// Validate each partition
	if err := m.partitionManager.ValidatePartitions(plan.Partitions); err != nil {
		return fmt.Errorf("partition validation failed: %w", err)
	}

	// Check resource constraints
	for nodeID, usage := range plan.ResourceUsage {
		if usage.CPUPercent > 90 || usage.MemoryPercent > 90 || usage.GPUPercent > 90 {
			return fmt.Errorf("node %s resource usage exceeds limits", nodeID)
		}
	}

	// Check quality score
	if plan.QualityScore < m.config.MinQualityThreshold {
		return fmt.Errorf("plan quality score %.2f below threshold %.2f",
			plan.QualityScore, m.config.MinQualityThreshold)
	}

	plan.Feasible = true
	return nil
}

// Helper methods

func (m *DynamicRepartitioningManager) createMigrationTasks(op *RepartitioningOperation) ([]MigrationTask, error) {
	var tasks []MigrationTask

	// Compare original and emergency plans to identify migrations
	originalMapping := m.createNodeShardMapping(op.OriginalPlan)
	emergencyMapping := m.createNodeShardMapping(op.EmergencyPlan)

	for shardID, originalNode := range originalMapping {
		emergencyNode, exists := emergencyMapping[shardID]
		if !exists {
			// Shard lacks mapping in emergency plan, pick a valid target node
			if len(op.AvailableNodes) == 0 {
				return nil, fmt.Errorf("no available nodes for shard %s migration", shardID)
			}
			// Pick the least loaded available node deterministically
			emergencyNode = m.selectLeastLoadedNode(op.AvailableNodes, emergencyMapping)
			if emergencyNode == "" {
				emergencyNode = op.AvailableNodes[0] // Fallback to first available node
			}
			emergencyMapping[shardID] = emergencyNode
		}

		if originalNode != emergencyNode {
			// Shard needs migration
			task := MigrationTask{
				ID:         uuid.New().String(),
				ShardID:    shardID,
				SourceNode: originalNode,
				TargetNode: emergencyNode,
				Status:     "pending",
				StartTime:  time.Now(),
			}

			// Get shard info for size
			shardInfo, err := m.shardManager.GetShardInfo(shardID)
			if err == nil {
				task.DataSize = shardInfo.Size
			}

			tasks = append(tasks, task)
		}
	}

	return tasks, nil
}

func (m *DynamicRepartitioningManager) createNodeShardMapping(plan *PartitionPlan) map[string]string {
	mapping := make(map[string]string)
	for _, partition := range plan.Partitions {
		for _, layer := range partition.ModelLayers {
			mapping[layer] = partition.NodeID
		}
	}
	return mapping
}

func (m *DynamicRepartitioningManager) executeParallelMigrations(ctx context.Context, tasks []MigrationTask) error {
	semaphore := make(chan struct{}, m.config.MaxParallelMigrations)
	var wg sync.WaitGroup
	errChan := make(chan error, len(tasks))

	for i := range tasks {
		wg.Add(1)
		go func(task *MigrationTask) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := m.executeMigrationTask(ctx, task); err != nil {
				errChan <- err
			}
		}(&tasks[i])
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *DynamicRepartitioningManager) executeSequentialMigrations(ctx context.Context, tasks []MigrationTask) error {
	for i := range tasks {
		if err := m.executeMigrationTask(ctx, &tasks[i]); err != nil {
			return err
		}
	}
	return nil
}

func (m *DynamicRepartitioningManager) executeMigrationTask(ctx context.Context, task *MigrationTask) error {
	task.Status = "in_progress"
	task.StartTime = time.Now()
	m.metrics.CurrentActiveMigrations++

	defer func() {
		m.metrics.CurrentActiveMigrations--
		task.EndTime = time.Now()
	}()

	// Execute migration with retries
	var lastErr error
	for attempt := 0; attempt < m.config.RetryAttempts; attempt++ {
		if err := m.shardManager.MigrateShard(ctx, task.ShardID, task.SourceNode, task.TargetNode); err != nil {
			lastErr = err
			task.Retries++
			time.Sleep(time.Second * time.Duration(attempt+1))
			continue
		}

		task.Status = "completed"
		return nil
	}

	task.Status = "failed"
	return fmt.Errorf("migration task %s failed after %d attempts: %w", task.ID, m.config.RetryAttempts, lastErr)
}

func (m *DynamicRepartitioningManager) rollbackMigrations(ctx context.Context, op *RepartitioningOperation) {
	// Attempt to rollback completed migrations
	for i := range op.MigrationTasks {
		task := &op.MigrationTasks[i]
		if task.Status == "completed" {
			// Reverse the migration
			m.shardManager.MigrateShard(ctx, task.ShardID, task.TargetNode, task.SourceNode)
		}
	}
}

// Strategy implementations

func (s *ConservativeStrategy) CreateEmergencyPlan(ctx context.Context, op *RepartitioningOperation) (*PartitionPlan, error) {
	plan := &PartitionPlan{
		ID:           uuid.New().String(),
		Strategy:     "conservative",
		Partitions:   []ModelPartition{},
		NodeMapping:  make(map[string][]string),
		ResourceUsage: make(map[string]ResourceUsage),
	}

	// Redistribute only partitions from failed nodes
	for _, partition := range op.OriginalPlan.Partitions {
		if contains(op.FailedNodes, partition.NodeID) {
			// Find least loaded available node
			targetNode := s.findLeastLoadedNode(op.AvailableNodes, plan.ResourceUsage)
			partition.NodeID = targetNode
		}
		plan.Partitions = append(plan.Partitions, partition)
		plan.NodeMapping[partition.NodeID] = append(plan.NodeMapping[partition.NodeID], partition.ID)
	}

	return plan, nil
}

func (s *ConservativeStrategy) ValidatePlan(plan *PartitionPlan) error {
	// Conservative validation
	return nil
}

func (s *ConservativeStrategy) EstimateQualityImpact(plan *PartitionPlan) float64 {
	// Conservative strategy has minimal quality impact
	return 0.05
}

func (s *ConservativeStrategy) EstimatePerformanceImpact(plan *PartitionPlan) float64 {
	// Conservative strategy may have higher performance impact
	return 0.15
}

func (s *ConservativeStrategy) findLeastLoadedNode(nodes []string, usage map[string]ResourceUsage) string {
	var leastLoaded string
	minLoad := float64(100)

	for _, node := range nodes {
		nodeUsage := usage[node]
		load := (nodeUsage.CPUPercent + nodeUsage.MemoryPercent + nodeUsage.GPUPercent) / 3
		if load < minLoad {
			minLoad = load
			leastLoaded = node
		}
	}

	return leastLoaded
}

func (s *AggressiveStrategy) CreateEmergencyPlan(ctx context.Context, op *RepartitioningOperation) (*PartitionPlan, error) {
	plan := &PartitionPlan{
		ID:           uuid.New().String(),
		Strategy:     "aggressive",
		Partitions:   []ModelPartition{},
		NodeMapping:  make(map[string][]string),
		ResourceUsage: make(map[string]ResourceUsage),
	}

	// Completely redistribute all partitions for optimal balance
	numNodes := len(op.AvailableNodes)
	if numNodes == 0 {
		return nil, fmt.Errorf("no available nodes for aggressive repartitioning")
	}

	numPartitions := len(op.OriginalPlan.Partitions)
	partitionsPerNode := numPartitions / numNodes
	extraPartitions := numPartitions % numNodes

	nodeIndex := 0
	for i, partition := range op.OriginalPlan.Partitions {
		targetNode := op.AvailableNodes[nodeIndex]
		partition.NodeID = targetNode
		plan.Partitions = append(plan.Partitions, partition)
		plan.NodeMapping[targetNode] = append(plan.NodeMapping[targetNode], partition.ID)

		// Move to next node after assigning appropriate number of partitions
		if partitionsPerNode > 0 {
			if (i+1)%(partitionsPerNode+1) == 0 && extraPartitions > 0 {
				nodeIndex++
				extraPartitions--
			} else if (i+1)%partitionsPerNode == 0 {
				nodeIndex++
			}
		} else {
			// When partitionsPerNode is 0 (fewer partitions than nodes),
			// distribute one partition per node
			nodeIndex++
		}

		if nodeIndex >= numNodes {
			nodeIndex = 0
		}
	}

	return plan, nil
}

func (s *AggressiveStrategy) ValidatePlan(plan *PartitionPlan) error {
	// Aggressive validation
	return nil
}

func (s *AggressiveStrategy) EstimateQualityImpact(plan *PartitionPlan) float64 {
	// Aggressive strategy may have moderate quality impact during migration
	return 0.10
}

func (s *AggressiveStrategy) EstimatePerformanceImpact(plan *PartitionPlan) float64 {
	// Aggressive strategy optimizes for performance after migration
	return 0.05
}

func (s *AdaptiveStrategy) CreateEmergencyPlan(ctx context.Context, op *RepartitioningOperation) (*PartitionPlan, error) {
	// Analyze current system load
	systemLoad := s.analyzeSystemLoad(op.AvailableNodes)

	// Choose strategy based on load
	if systemLoad < 0.5 {
		// Low load - use aggressive strategy
		aggressive := &AggressiveStrategy{manager: s.manager}
		return aggressive.CreateEmergencyPlan(ctx, op)
	} else if systemLoad > 0.8 {
		// High load - use conservative strategy
		conservative := &ConservativeStrategy{manager: s.manager}
		return conservative.CreateEmergencyPlan(ctx, op)
	}

	// Medium load - use adaptive approach
	plan := &PartitionPlan{
		ID:           uuid.New().String(),
		Strategy:     "adaptive",
		Partitions:   []ModelPartition{},
		NodeMapping:  make(map[string][]string),
		ResourceUsage: make(map[string]ResourceUsage),
	}

	// Redistribute based on node capabilities and current load
	for _, partition := range op.OriginalPlan.Partitions {
		if contains(op.FailedNodes, partition.NodeID) || systemLoad > 0.6 {
			// Find optimal node based on multiple factors
			targetNode := s.findOptimalNode(op.AvailableNodes, partition, plan.ResourceUsage)
			partition.NodeID = targetNode
		}
		plan.Partitions = append(plan.Partitions, partition)
		plan.NodeMapping[partition.NodeID] = append(plan.NodeMapping[partition.NodeID], partition.ID)
	}

	return plan, nil
}

func (s *AdaptiveStrategy) ValidatePlan(plan *PartitionPlan) error {
	// Adaptive validation
	return nil
}

func (s *AdaptiveStrategy) EstimateQualityImpact(plan *PartitionPlan) float64 {
	// Adaptive strategy balances quality impact
	return 0.07
}

func (s *AdaptiveStrategy) EstimatePerformanceImpact(plan *PartitionPlan) float64 {
	// Adaptive strategy balances performance impact
	return 0.08
}

func (s *AdaptiveStrategy) analyzeSystemLoad(nodes []string) float64 {
	// Placeholder for system load analysis
	return 0.6
}

func (s *AdaptiveStrategy) findOptimalNode(nodes []string, partition ModelPartition, usage map[string]ResourceUsage) string {
	// Placeholder for optimal node selection
	if len(nodes) > 0 {
		return nodes[0]
	}
	return ""
}

// Utility functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetMetrics returns repartitioning metrics
func (m *DynamicRepartitioningManager) GetMetrics() *RepartitioningMetrics {
	m.metrics.mu.RLock()
	defer m.metrics.mu.RUnlock()
	return m.metrics
}

// GetActiveRepartitionings returns active repartitioning operations
func (m *DynamicRepartitioningManager) GetActiveRepartitionings() map[string]*RepartitioningOperation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*RepartitioningOperation)
	for k, v := range m.activeRepartitionings {
		result[k] = v
	}
	return result
}
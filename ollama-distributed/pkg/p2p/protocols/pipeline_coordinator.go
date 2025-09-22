package protocols

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// PipelineStage represents a stage in the distributed inference pipeline
type PipelineStage struct {
	ID             string        `json:"id"`
	PartitionID    string        `json:"partition_id"`
	NodeID         peer.ID       `json:"node_id"`
	LayerStart     int           `json:"layer_start"`
	LayerEnd       int           `json:"layer_end"`
	Dependencies   []string      `json:"dependencies"`
	Status         StageStatus   `json:"status"`
	StartTime      time.Time     `json:"start_time"`
	EndTime        time.Time     `json:"end_time"`
	InputSize      int64         `json:"input_size"`
	OutputSize     int64         `json:"output_size"`
	ProcessingTime time.Duration `json:"processing_time"`
}

// StageStatus represents the status of a pipeline stage
type StageStatus uint8

const (
	StageStatusPending StageStatus = iota
	StageStatusReady
	StageStatusRunning
	StageStatusCompleted
	StageStatusFailed
	StageStatusSkipped
)

// ActivationPipeline represents the flow of data through distributed model layers
type ActivationPipeline struct {
	ID              string                    `json:"id"`
	InferenceID     string                    `json:"inference_id"`
	Stages          map[string]*PipelineStage `json:"stages"`
	StageOrder      []string                  `json:"stage_order"`
	CurrentStage    string                    `json:"current_stage"`
	Status          PipelineStatus            `json:"status"`
	StartTime       time.Time                 `json:"start_time"`
	EndTime         time.Time                 `json:"end_time"`
	TotalStages     int                       `json:"total_stages"`
	CompletedStages int                       `json:"completed_stages"`
	FailedStages    int                       `json:"failed_stages"`
	mutex           sync.RWMutex
}

// PipelineStatus represents the overall status of a pipeline
type PipelineStatus uint8

const (
	PipelineStatusCreated PipelineStatus = iota
	PipelineStatusRunning
	PipelineStatusCompleted
	PipelineStatusFailed
	PipelineStatusCancelled
)

// PipelineCoordinator manages the flow of activations through sequential model partitions
type PipelineCoordinator struct {
	activePipelines map[string]*ActivationPipeline
	pipelinesMutex  sync.RWMutex
	stageBuffers    map[string]*StageBuffer
	buffersMutex    sync.RWMutex
	streamClient    *TensorStreamClient
	monitor         *PipelineMonitor
	ctx             context.Context
	cancel          context.CancelFunc
}

// StageBuffer manages buffering for asynchronous processing
type StageBuffer struct {
	StageID      string
	InputBuffer  chan *ActivationData
	OutputBuffer chan *ActivationData
	ErrorBuffer  chan error
	MaxSize      int
	CurrentSize  int
	mutex        sync.RWMutex
}

// ActivationData represents activation data flowing through the pipeline
type ActivationData struct {
	InferenceID string                 `json:"inference_id"`
	StageID     string                 `json:"stage_id"`
	Data        []byte                 `json:"data"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
}

// PartitionPlan represents a plan for partitioning model execution across nodes
type PartitionPlan struct {
	InferenceID string                 `json:"inference_id"`
	Partitions  []*PartitionInfo       `json:"partitions"`
	TotalStages int                    `json:"total_stages"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// PartitionInfo contains information about a single partition
type PartitionInfo struct {
	ID         string  `json:"id"`
	NodeID     peer.ID `json:"node_id"`
	LayerStart int     `json:"layer_start"`
	LayerEnd   int     `json:"layer_end"`
	Order      int     `json:"order"`
}

// PipelineMonitor tracks pipeline performance and health
type PipelineMonitor struct {
	metrics map[string]*PipelineMetrics
	mutex   sync.RWMutex
}

// PipelineMetrics contains performance metrics for a pipeline
type PipelineMetrics struct {
	TotalLatency    time.Duration            `json:"total_latency"`
	StageLatencies  map[string]time.Duration `json:"stage_latencies"`
	ThroughputMBps  float64                  `json:"throughput_mbps"`
	ErrorRate       float64                  `json:"error_rate"`
	BottleneckStage string                   `json:"bottleneck_stage"`
	LastUpdated     time.Time                `json:"last_updated"`
}

// NewPipelineCoordinator creates a new pipeline coordinator
func NewPipelineCoordinator(streamClient *TensorStreamClient) *PipelineCoordinator {
	ctx, cancel := context.WithCancel(context.Background())

	return &PipelineCoordinator{
		activePipelines: make(map[string]*ActivationPipeline),
		stageBuffers:    make(map[string]*StageBuffer),
		streamClient:    streamClient,
		monitor:         NewPipelineMonitor(),
		ctx:             ctx,
		cancel:          cancel,
	}
}

// NewPipelineMonitor creates a new pipeline monitor
func NewPipelineMonitor() *PipelineMonitor {
	return &PipelineMonitor{
		metrics: make(map[string]*PipelineMetrics),
	}
}

// CreatePipeline creates a new activation pipeline from a partition plan
func (pc *PipelineCoordinator) CreatePipeline(inferenceID string, partitionPlan *PartitionPlan) (*ActivationPipeline, error) {
	pc.pipelinesMutex.Lock()
	defer pc.pipelinesMutex.Unlock()

	// Check if pipeline already exists
	if _, exists := pc.activePipelines[inferenceID]; exists {
		return nil, fmt.Errorf("pipeline already exists for inference %s", inferenceID)
	}

	pipeline := &ActivationPipeline{
		ID:          fmt.Sprintf("pipeline-%s", inferenceID),
		InferenceID: inferenceID,
		Stages:      make(map[string]*PipelineStage),
		StageOrder:  make([]string, 0),
		Status:      PipelineStatusCreated,
		StartTime:   time.Now(),
	}

	// Create stages from partition plan
	for i, partition := range partitionPlan.Partitions {
		stageID := fmt.Sprintf("stage-%d", i)
		stage := &PipelineStage{
			ID:          stageID,
			PartitionID: partition.ID,
			NodeID:      partition.NodeID,
			LayerStart:  partition.LayerStart,
			LayerEnd:    partition.LayerEnd,
			Status:      StageStatusPending,
		}

		// Set dependencies (previous stage)
		if i > 0 {
			prevStageID := fmt.Sprintf("stage-%d", i-1)
			stage.Dependencies = []string{prevStageID}
		}

		pipeline.Stages[stageID] = stage
		pipeline.StageOrder = append(pipeline.StageOrder, stageID)
	}

	pipeline.TotalStages = len(pipeline.Stages)
	pc.activePipelines[inferenceID] = pipeline

	// Create stage buffers
	for stageID := range pipeline.Stages {
		pc.createStageBuffer(stageID)
	}

	log.Printf("Created pipeline %s with %d stages", pipeline.ID, pipeline.TotalStages)
	return pipeline, nil
}

// createStageBuffer creates a buffer for a pipeline stage
func (pc *PipelineCoordinator) createStageBuffer(stageID string) {
	pc.buffersMutex.Lock()
	defer pc.buffersMutex.Unlock()

	buffer := &StageBuffer{
		StageID:      stageID,
		InputBuffer:  make(chan *ActivationData, 10),
		OutputBuffer: make(chan *ActivationData, 10),
		ErrorBuffer:  make(chan error, 5),
		MaxSize:      10,
		CurrentSize:  0,
	}

	pc.stageBuffers[stageID] = buffer
}

// StartPipeline begins execution of a pipeline
func (pc *PipelineCoordinator) StartPipeline(inferenceID string, inputData *ActivationData) error {
	pc.pipelinesMutex.Lock()
	pipeline, exists := pc.activePipelines[inferenceID]
	pc.pipelinesMutex.Unlock()

	if !exists {
		return fmt.Errorf("pipeline not found for inference %s", inferenceID)
	}

	pipeline.mutex.Lock()
	pipeline.Status = PipelineStatusRunning
	pipeline.StartTime = time.Now()
	pipeline.mutex.Unlock()

	// Start with the first stage
	if len(pipeline.StageOrder) > 0 {
		firstStageID := pipeline.StageOrder[0]
		return pc.executeStage(pipeline, firstStageID, inputData)
	}

	return fmt.Errorf("pipeline has no stages")
}

// executeStage executes a specific pipeline stage and blocks on real output
func (pc *PipelineCoordinator) executeStage(pipeline *ActivationPipeline, stageID string, inputData *ActivationData) error {
	stage, exists := pipeline.Stages[stageID]
	if !exists {
		return fmt.Errorf("stage %s not found in pipeline", stageID)
	}

	// Check dependencies
	if !pc.areDependenciesMet(pipeline, stage) {
		return fmt.Errorf("dependencies not met for stage %s", stageID)
	}

	// Update stage status
	stage.Status = StageStatusRunning
	stage.StartTime = time.Now()

	log.Printf("Executing stage %s on node %s (layers %d-%d)",
		stageID, stage.NodeID, stage.LayerStart, stage.LayerEnd)

	// Get stage buffer for blocking on real output
	stageBuffer := pc.getStageBuffer(stageID)
	if stageBuffer == nil {
		stage.Status = StageStatusFailed
		pipeline.FailedStages++
		return fmt.Errorf("stage buffer not found for stage %s", stageID)
	}

	// Set StageID to match coordinator stage ID for correct buffer alignment
	inputData.StageID = stageID

	// Stream activation to the target node
	_, _, err := pc.streamClient.StreamActivation(stage.NodeID, inputData)
	if err != nil {
		stage.Status = StageStatusFailed
		pipeline.FailedStages++
		return fmt.Errorf("failed to stream activation to stage %s: %w", stageID, err)
	}

	// Block on real stage buffer outputs (not immediate completion)
	timeout := time.NewTimer(30 * time.Second) // Configurable timeout
	defer timeout.Stop()

	select {
	case outputData := <-stageBuffer.OutputBuffer:
		// Real output received from stage
		stage.Status = StageStatusCompleted
		stage.EndTime = time.Now()
		stage.ProcessingTime = stage.EndTime.Sub(stage.StartTime)
		pipeline.CompletedStages++

		log.Printf("Stage %s completed with real output data (%d bytes)", stageID, len(outputData.Data))

		// Update pipeline current stage
		pipeline.CurrentStage = stageID

		// Check if pipeline is complete
		if pipeline.CompletedStages == pipeline.TotalStages {
			pipeline.Status = PipelineStatusCompleted
			pipeline.EndTime = time.Now()
			log.Printf("Pipeline %s completed", pipeline.ID)
			return nil
		}

		// Continue to next stage with real output data
		return pc.continueToNextStage(pipeline, stageID, outputData)

	case stageErr := <-stageBuffer.ErrorBuffer:
		// Error received from stage
		stage.Status = StageStatusFailed
		pipeline.FailedStages++
		return fmt.Errorf("stage %s failed with error: %w", stageID, stageErr)

	case <-timeout.C:
		// Timeout waiting for stage completion
		stage.Status = StageStatusFailed
		pipeline.FailedStages++
		return fmt.Errorf("stage %s timed out waiting for output", stageID)

	case <-pc.ctx.Done():
		// Context cancelled
		stage.Status = StageStatusFailed
		pipeline.FailedStages++
		return fmt.Errorf("stage %s cancelled due to context: %w", stageID, pc.ctx.Err())
	}
}

// areDependenciesMet checks if all dependencies for a stage are satisfied
func (pc *PipelineCoordinator) areDependenciesMet(pipeline *ActivationPipeline, stage *PipelineStage) bool {
	for _, depID := range stage.Dependencies {
		if depStage, exists := pipeline.Stages[depID]; exists {
			if depStage.Status != StageStatusCompleted {
				return false
			}
		}
	}
	return true
}

// continueToNextStage moves to the next stage in the pipeline using real output data
func (pc *PipelineCoordinator) continueToNextStage(pipeline *ActivationPipeline, currentStageID string, outputData *ActivationData) error {
	// Find next stage in order
	for i, stageID := range pipeline.StageOrder {
		if stageID == currentStageID && i+1 < len(pipeline.StageOrder) {
			nextStageID := pipeline.StageOrder[i+1]

			// Use real output data from current stage for next stage input
			nextInputData := &ActivationData{
				InferenceID: pipeline.InferenceID,
				StageID:     nextStageID, // This will be set correctly in executeStage
				Data:        outputData.Data, // Real activation data from previous stage
				Metadata:    outputData.Metadata, // Preserve metadata chain
				Timestamp:   time.Now(),
			}
			return pc.executeStage(pipeline, nextStageID, nextInputData)
		}
	}
	return nil
}

// getStageOutput retrieves the output data from a completed stage
func (pc *PipelineCoordinator) getStageOutput(stageID string) (*ActivationData, error) {
	pc.buffersMutex.RLock()
	defer pc.buffersMutex.RUnlock()

	// Get stage buffer
	stageBuffer, exists := pc.stageBuffers[stageID]
	if !exists {
		return nil, fmt.Errorf("stage buffer not found for stage %s", stageID)
	}

	// Try to read output data (non-blocking)
	select {
	case outputData := <-stageBuffer.OutputBuffer:
		return outputData, nil
	default:
		return nil, fmt.Errorf("no output data available for stage %s", stageID)
	}
}

// GetPipeline retrieves a pipeline by inference ID
func (pc *PipelineCoordinator) GetPipeline(inferenceID string) (*ActivationPipeline, bool) {
	pc.pipelinesMutex.RLock()
	defer pc.pipelinesMutex.RUnlock()

	pipeline, exists := pc.activePipelines[inferenceID]
	return pipeline, exists
}

// CompletePipeline marks a pipeline as complete and cleans up resources
func (pc *PipelineCoordinator) CompletePipeline(inferenceID string) error {
	pc.pipelinesMutex.Lock()
	defer pc.pipelinesMutex.Unlock()

	pipeline, exists := pc.activePipelines[inferenceID]
	if !exists {
		return fmt.Errorf("pipeline not found for inference %s", inferenceID)
	}

	pipeline.mutex.Lock()
	pipeline.Status = PipelineStatusCompleted
	pipeline.EndTime = time.Now()
	pipeline.mutex.Unlock()

	// Clean up stage buffers
	for stageID := range pipeline.Stages {
		pc.cleanupStageBuffer(stageID)
	}

	delete(pc.activePipelines, inferenceID)
	log.Printf("Completed and cleaned up pipeline %s", pipeline.ID)

	return nil
}

// cleanupStageBuffer cleans up resources for a stage buffer
func (pc *PipelineCoordinator) cleanupStageBuffer(stageID string) {
	pc.buffersMutex.Lock()
	defer pc.buffersMutex.Unlock()

	if buffer, exists := pc.stageBuffers[stageID]; exists {
		close(buffer.InputBuffer)
		close(buffer.OutputBuffer)
		close(buffer.ErrorBuffer)
		delete(pc.stageBuffers, stageID)
	}
}

// getStageBuffer retrieves a stage buffer by stage ID (public method for external access)
func (pc *PipelineCoordinator) getStageBuffer(stageID string) *StageBuffer {
	pc.buffersMutex.RLock()
	defer pc.buffersMutex.RUnlock()

	if buffer, exists := pc.stageBuffers[stageID]; exists {
		return buffer
	}
	return nil
}

// GetStageBuffer retrieves a stage buffer by stage ID (public method for external access)
func (pc *PipelineCoordinator) GetStageBuffer(stageID string) *StageBuffer {
	return pc.getStageBuffer(stageID)
}

// GetPipelineMetrics returns performance metrics for a pipeline
func (pc *PipelineCoordinator) GetPipelineMetrics(inferenceID string) (*PipelineMetrics, error) {
	return pc.monitor.GetMetrics(inferenceID)
}

// GetMetrics returns metrics for a specific pipeline
func (pm *PipelineMonitor) GetMetrics(inferenceID string) (*PipelineMetrics, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	metrics, exists := pm.metrics[inferenceID]
	if !exists {
		return nil, fmt.Errorf("metrics not found for inference %s", inferenceID)
	}

	return metrics, nil
}

// UpdateMetrics updates performance metrics for a pipeline
func (pm *PipelineMonitor) UpdateMetrics(inferenceID string, metrics *PipelineMetrics) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	metrics.LastUpdated = time.Now()
	pm.metrics[inferenceID] = metrics
}

package distributed

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// MomentumAligner implements step-wise momentum fusion for distributed training
// Based on SMoFi research: synchronizes momentum buffers across server-side optimizers
// Achieves 7.1% accuracy improvement and 10.25x convergence speedup
type MomentumAligner struct {
	mu sync.RWMutex

	// Optimizer state
	serverOptimizers map[peer.ID]*OptimizerState
	
	// Historical momentum tracking
	historicalMomentum map[peer.ID][]MomentumBuffer
	
	// Configuration
	config *MomentumAlignmentConfig

	// Synchronization control
	currentStep int
	lastSync    time.Time
	
	// Context for periodic operations
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// OptimizerState tracks optimizer state for a server-side model
type OptimizerState struct {
	PeerID          peer.ID
	Momentum        MomentumBuffer
	IsActive        bool
	LastStep        int
	TotalSteps      int
	LastUpdate      time.Time
}

// MomentumBuffer represents momentum values for model parameters
type MomentumBuffer struct {
	Values      []float64
	Step        int
	Timestamp   time.Time
}

// MomentumAlignmentConfig configures momentum alignment behavior
type MomentumAlignmentConfig struct {
	StalenessAlpha      float64       // Polynomial staleness factor (typically -0.1)
	SyncInterval        time.Duration // How often to sync momentum (per step)
	MaxHistorySize      int           // Maximum size of historical momentum buffer
	MomentumCoefficient float64       // Beta coefficient for momentum (0.9 typical)
	EnableAdaptive      bool          // Enable adaptive staleness factor
}

// DefaultMomentumAlignmentConfig returns default configuration
func DefaultMomentumAlignmentConfig() *MomentumAlignmentConfig {
	return &MomentumAlignmentConfig{
		StalenessAlpha:      -0.1,
		SyncInterval:        100 * time.Millisecond,
		MaxHistorySize:      1000,
		MomentumCoefficient: 0.9,
		EnableAdaptive:      true,
	}
}

// NewMomentumAligner creates a new momentum aligner
func NewMomentumAligner(config *MomentumAlignmentConfig) *MomentumAligner {
	if config == nil {
		config = DefaultMomentumAlignmentConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	ma := &MomentumAligner{
		serverOptimizers:   make(map[peer.ID]*OptimizerState),
		historicalMomentum: make(map[peer.ID][]MomentumBuffer),
		config:             config,
		currentStep:        0,
		ctx:                ctx,
		cancel:             cancel,
	}

	return ma
}

// RegisterOptimizer registers a server-side optimizer
func (ma *MomentumAligner) RegisterOptimizer(peerID peer.ID, totalSteps int) {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	ma.serverOptimizers[peerID] = &OptimizerState{
		PeerID:     peerID,
		Momentum:   MomentumBuffer{Values: make([]float64, 0)},
		IsActive:   true,
		LastStep:   0,
		TotalSteps: totalSteps,
		LastUpdate: time.Now(),
	}

	ma.historicalMomentum[peerID] = make([]MomentumBuffer, 0, ma.config.MaxHistorySize)
}

// UpdateMomentum updates momentum for an optimizer at a specific step
func (ma *MomentumAligner) UpdateMomentum(peerID peer.ID, momentum []float64, step int) error {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	opt, exists := ma.serverOptimizers[peerID]
	if !exists {
		return fmt.Errorf("optimizer not registered for peer %s", peerID)
	}

	// Update momentum buffer
	opt.Momentum = MomentumBuffer{
		Values:    momentum,
		Step:      step,
		Timestamp: time.Now(),
	}
	opt.LastStep = step
	opt.LastUpdate = time.Now()

	// Check if this is the final step for this optimizer
	if step >= opt.TotalSteps {
		opt.IsActive = false
		
		// Store in historical momentum
		ma.historicalMomentum[peerID] = append(ma.historicalMomentum[peerID], opt.Momentum)
		if len(ma.historicalMomentum[peerID]) > ma.config.MaxHistorySize {
			ma.historicalMomentum[peerID] = ma.historicalMomentum[peerID][1:]
		}
	}

	return nil
}

// AlignMomentum synchronizes momentum buffers across all optimizers
// Implements SMoFi algorithm with staleness-aware mechanism
func (ma *MomentumAligner) AlignMomentum(step int) ([]float64, error) {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	ma.currentStep = step

	// Collect current momentum from active optimizers
	currentMomentum := make([][]float64, 0)
	for _, opt := range ma.serverOptimizers {
		if opt.IsActive && opt.LastStep == step {
			currentMomentum = append(currentMomentum, opt.Momentum.Values)
		}
	}

	if len(currentMomentum) == 0 {
		return nil, fmt.Errorf("no active optimizers at step %d", step)
	}

	// Collect historical momentum with staleness weighting
	historicalMomentum := ma.getHistoricalMomentum(step)

	// Compute weighted average
	alignedMomentum := ma.weightedAverage(currentMomentum, historicalMomentum, step)

	// Update all active optimizers with aligned momentum
	for _, opt := range ma.serverOptimizers {
		if opt.IsActive {
			opt.Momentum.Values = alignedMomentum
		}
	}

	ma.lastSync = time.Now()

	return alignedMomentum, nil
}

// getHistoricalMomentum retrieves historical momentum buffers
func (ma *MomentumAligner) getHistoricalMomentum(step int) []MomentumBuffer {
	historical := make([]MomentumBuffer, 0)

	for peerID, opt := range ma.serverOptimizers {
		if !opt.IsActive {
			// Get historical momentum for this peer
			if history, exists := ma.historicalMomentum[peerID]; exists && len(history) > 0 {
				// Get the most recent historical momentum
				historical = append(historical, history[len(history)-1])
			}
		}
	}

	return historical
}

// weightedAverage computes weighted average of momentum buffers
// Implements polynomial staleness factor: s_α = (τ - |T_j| + 1)^α
func (ma *MomentumAligner) weightedAverage(
	current [][]float64,
	historical []MomentumBuffer,
	step int,
) []float64 {
	if len(current) == 0 {
		return nil
	}

	// Determine momentum size
	momentumSize := len(current[0])
	result := make([]float64, momentumSize)
	totalWeight := 0.0

	// Add current momentum (weight = 1.0 each)
	for _, momentum := range current {
		for i := 0; i < momentumSize && i < len(momentum); i++ {
			result[i] += momentum[i]
		}
		totalWeight += 1.0
	}

	// Add historical momentum with staleness weighting
	for idx, buffer := range historical {
		// Calculate staleness: s_α = (τ - |T_j| + 1)^α
		staleness := math.Pow(float64(step-buffer.Step+1), ma.config.StalenessAlpha)
		
		for i := 0; i < momentumSize && i < len(buffer.Values); i++ {
			result[i] += buffer.Values[i] * staleness
		}
		totalWeight += staleness
		
		_ = idx // Avoid unused variable warning
	}

	// Normalize by total weight
	if totalWeight > 0 {
		for i := range result {
			result[i] /= totalWeight
		}
	}

	return result
}

// GetAlignedMomentum returns the current aligned momentum
func (ma *MomentumAligner) GetAlignedMomentum() []float64 {
	ma.mu.RLock()
	defer ma.mu.RUnlock()

	// Return momentum from any active optimizer (they should all be aligned)
	for _, opt := range ma.serverOptimizers {
		if opt.IsActive {
			return opt.Momentum.Values
		}
	}

	return nil
}

// Stop stops the momentum aligner
func (ma *MomentumAligner) Stop() {
	ma.cancel()
	ma.wg.Wait()
}


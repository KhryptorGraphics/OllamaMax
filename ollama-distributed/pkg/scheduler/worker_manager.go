package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// WorkerMetrics tracks worker performance metrics
type WorkerMetrics struct {
	TotalWorkers        int64     `json:"total_workers"`
	ActiveWorkers       int64     `json:"active_workers"`
	IdleWorkers         int64     `json:"idle_workers"`
	OfflineWorkers      int64     `json:"offline_workers"`
	AverageLoad         float64   `json:"average_load"`
	TotalCapacity       int64     `json:"total_capacity"`
	UsedCapacity        int64     `json:"used_capacity"`
	LastUpdated         time.Time `json:"last_updated"`
	mu                  sync.RWMutex
}

// WorkerHealthChecker monitors worker health
type WorkerHealthChecker struct {
	manager             *WorkerManager
	interval            time.Duration
	timeout             time.Duration
	ctx                 context.Context
	cancel              context.CancelFunc
	wg                  sync.WaitGroup
}

// NewWorkerManager creates a new worker manager
func NewWorkerManager(config *WorkerManagerConfig) (*WorkerManager, error) {
	if config == nil {
		config = &WorkerManagerConfig{
			MaxWorkers:          1000,
			HealthCheckInterval: 30 * time.Second,
			WorkerTimeout:       60 * time.Second,
			CapabilityRefresh:   5 * time.Minute,
		}
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	manager := &WorkerManager{
		config:       config,
		workers:      make(map[peer.ID]*WorkerNode),
		capabilities: make(map[string][]*WorkerNode),
		metrics: &WorkerMetrics{
			LastUpdated: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}
	
	// Create health checker
	manager.healthChecker = &WorkerHealthChecker{
		manager:  manager,
		interval: config.HealthCheckInterval,
		timeout:  config.WorkerTimeout,
		ctx:      ctx,
		cancel:   cancel,
	}
	
	return manager, nil
}

// Start starts the worker manager
func (wm *WorkerManager) Start() error {
	// Start health checker
	wm.wg.Add(1)
	go wm.healthChecker.start()
	
	// Start capability refresh
	wm.wg.Add(1)
	go wm.capabilityRefreshLoop()
	
	// Start metrics collection
	wm.wg.Add(1)
	go wm.metricsLoop()
	
	return nil
}

// Stop stops the worker manager
func (wm *WorkerManager) Stop() error {
	wm.cancel()
	wm.wg.Wait()
	return nil
}

// RegisterWorker registers a new worker node
func (wm *WorkerManager) RegisterWorker(worker *WorkerNode) error {
	if worker == nil {
		return fmt.Errorf("worker cannot be nil")
	}
	
	if worker.ID == "" {
		return fmt.Errorf("worker ID is required")
	}
	
	wm.workersMu.Lock()
	defer wm.workersMu.Unlock()
	
	// Check if we've reached the maximum number of workers
	if len(wm.workers) >= wm.config.MaxWorkers {
		return fmt.Errorf("maximum number of workers reached")
	}
	
	// Set initial values
	worker.LastSeen = time.Now()
	worker.Status = WorkerStatusOnline
	
	// Register worker
	wm.workers[worker.ID] = worker
	
	// Update capabilities
	wm.updateWorkerCapabilities(worker)
	
	// Update metrics
	wm.updateMetrics()
	
	return nil
}

// UnregisterWorker removes a worker node
func (wm *WorkerManager) UnregisterWorker(workerID peer.ID) error {
	wm.workersMu.Lock()
	defer wm.workersMu.Unlock()
	
	worker, exists := wm.workers[workerID]
	if !exists {
		return fmt.Errorf("worker not found")
	}
	
	// Remove from capabilities
	wm.removeWorkerCapabilities(worker)
	
	// Remove worker
	delete(wm.workers, workerID)
	
	// Update metrics
	wm.updateMetrics()
	
	return nil
}

// GetWorker returns a worker by ID
func (wm *WorkerManager) GetWorker(workerID peer.ID) (*WorkerNode, bool) {
	wm.workersMu.RLock()
	defer wm.workersMu.RUnlock()
	
	worker, exists := wm.workers[workerID]
	return worker, exists
}

// GetAllWorkers returns all registered workers
func (wm *WorkerManager) GetAllWorkers() []*WorkerNode {
	wm.workersMu.RLock()
	defer wm.workersMu.RUnlock()
	
	workers := make([]*WorkerNode, 0, len(wm.workers))
	for _, worker := range wm.workers {
		workers = append(workers, worker)
	}
	
	return workers
}

// GetAvailableWorkers returns workers that are online and not at capacity
func (wm *WorkerManager) GetAvailableWorkers() []*WorkerNode {
	wm.workersMu.RLock()
	defer wm.workersMu.RUnlock()
	
	var available []*WorkerNode
	for _, worker := range wm.workers {
		if worker.Status == WorkerStatusOnline || worker.Status == WorkerStatusIdle {
			// Check if worker has capacity
			if worker.Load != nil && worker.Load.ActiveTasks < 10 { // Simple capacity check
				available = append(available, worker)
			}
		}
	}
	
	return available
}

// GetWorkersByCapability returns workers with a specific capability
func (wm *WorkerManager) GetWorkersByCapability(capability string) []*WorkerNode {
	wm.capabilitiesMu.RLock()
	defer wm.capabilitiesMu.RUnlock()
	
	workers, exists := wm.capabilities[capability]
	if !exists {
		return nil
	}
	
	// Return a copy to avoid race conditions
	result := make([]*WorkerNode, len(workers))
	copy(result, workers)
	return result
}

// UpdateWorkerStatus updates a worker's status
func (wm *WorkerManager) UpdateWorkerStatus(workerID peer.ID, status WorkerStatus) error {
	wm.workersMu.Lock()
	defer wm.workersMu.Unlock()
	
	worker, exists := wm.workers[workerID]
	if !exists {
		return fmt.Errorf("worker not found")
	}
	
	worker.Status = status
	worker.LastSeen = time.Now()
	
	// Update metrics
	wm.updateMetrics()
	
	return nil
}

// UpdateWorkerLoad updates a worker's load information
func (wm *WorkerManager) UpdateWorkerLoad(workerID peer.ID, load *LoadInfo) error {
	wm.workersMu.Lock()
	defer wm.workersMu.Unlock()
	
	worker, exists := wm.workers[workerID]
	if !exists {
		return fmt.Errorf("worker not found")
	}
	
	worker.Load = load
	worker.LastSeen = time.Now()
	
	// Update status based on load
	if load.ActiveTasks == 0 {
		worker.Status = WorkerStatusIdle
	} else if load.ActiveTasks >= 10 { // Simple threshold
		worker.Status = WorkerStatusBusy
	} else {
		worker.Status = WorkerStatusOnline
	}
	
	return nil
}

// GetMetrics returns current worker metrics
func (wm *WorkerManager) GetMetrics() *WorkerMetrics {
	wm.metrics.mu.RLock()
	defer wm.metrics.mu.RUnlock()
	
	// Create a copy
	metrics := *wm.metrics
	return &metrics
}

// updateWorkerCapabilities updates the capability index for a worker
func (wm *WorkerManager) updateWorkerCapabilities(worker *WorkerNode) {
	wm.capabilitiesMu.Lock()
	defer wm.capabilitiesMu.Unlock()
	
	// Remove worker from old capabilities
	wm.removeWorkerCapabilitiesLocked(worker)
	
	// Add worker to new capabilities
	for _, capability := range worker.Capabilities {
		if wm.capabilities[capability] == nil {
			wm.capabilities[capability] = make([]*WorkerNode, 0)
		}
		wm.capabilities[capability] = append(wm.capabilities[capability], worker)
	}
}

// removeWorkerCapabilities removes a worker from the capability index
func (wm *WorkerManager) removeWorkerCapabilities(worker *WorkerNode) {
	wm.capabilitiesMu.Lock()
	defer wm.capabilitiesMu.Unlock()
	
	wm.removeWorkerCapabilitiesLocked(worker)
}

// removeWorkerCapabilitiesLocked removes a worker from capabilities (assumes lock held)
func (wm *WorkerManager) removeWorkerCapabilitiesLocked(worker *WorkerNode) {
	for capability, workers := range wm.capabilities {
		for i, w := range workers {
			if w.ID == worker.ID {
				// Remove worker from slice
				wm.capabilities[capability] = append(workers[:i], workers[i+1:]...)
				break
			}
		}
		
		// Clean up empty capability lists
		if len(wm.capabilities[capability]) == 0 {
			delete(wm.capabilities, capability)
		}
	}
}

// updateMetrics updates worker metrics
func (wm *WorkerManager) updateMetrics() {
	wm.metrics.mu.Lock()
	defer wm.metrics.mu.Unlock()
	
	var totalWorkers, activeWorkers, idleWorkers, offlineWorkers int64
	var totalLoad float64
	
	for _, worker := range wm.workers {
		totalWorkers++
		
		switch worker.Status {
		case WorkerStatusOnline, WorkerStatusBusy:
			activeWorkers++
		case WorkerStatusIdle:
			idleWorkers++
		case WorkerStatusOffline, WorkerStatusError:
			offlineWorkers++
		}
		
		if worker.Load != nil {
			totalLoad += worker.Load.CPUUsage
		}
	}
	
	wm.metrics.TotalWorkers = totalWorkers
	wm.metrics.ActiveWorkers = activeWorkers
	wm.metrics.IdleWorkers = idleWorkers
	wm.metrics.OfflineWorkers = offlineWorkers
	
	if totalWorkers > 0 {
		wm.metrics.AverageLoad = totalLoad / float64(totalWorkers)
	}
	
	wm.metrics.LastUpdated = time.Now()
}

// capabilityRefreshLoop refreshes worker capabilities periodically
func (wm *WorkerManager) capabilityRefreshLoop() {
	defer wm.wg.Done()
	
	ticker := time.NewTicker(wm.config.CapabilityRefresh)
	defer ticker.Stop()
	
	for {
		select {
		case <-wm.ctx.Done():
			return
		case <-ticker.C:
			wm.refreshCapabilities()
		}
	}
}

// refreshCapabilities refreshes the capability index
func (wm *WorkerManager) refreshCapabilities() {
	wm.workersMu.RLock()
	workers := make([]*WorkerNode, 0, len(wm.workers))
	for _, worker := range wm.workers {
		workers = append(workers, worker)
	}
	wm.workersMu.RUnlock()
	
	// Rebuild capability index
	for _, worker := range workers {
		wm.updateWorkerCapabilities(worker)
	}
}

// metricsLoop runs the metrics collection loop
func (wm *WorkerManager) metricsLoop() {
	defer wm.wg.Done()
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-wm.ctx.Done():
			return
		case <-ticker.C:
			wm.updateMetrics()
		}
	}
}

// start starts the health checker
func (whc *WorkerHealthChecker) start() {
	defer whc.manager.wg.Done()
	
	ticker := time.NewTicker(whc.interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-whc.ctx.Done():
			return
		case <-ticker.C:
			whc.checkWorkerHealth()
		}
	}
}

// checkWorkerHealth checks the health of all workers
func (whc *WorkerHealthChecker) checkWorkerHealth() {
	workers := whc.manager.GetAllWorkers()
	
	for _, worker := range workers {
		// Check if worker has been seen recently
		if time.Since(worker.LastSeen) > whc.timeout {
			// Mark worker as offline
			whc.manager.UpdateWorkerStatus(worker.ID, WorkerStatusOffline)
		}
	}
}

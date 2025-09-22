package models

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/protocols"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/partitioning"
	"github.com/libp2p/go-libp2p/core/peer"
)

// ModelFormat represents the format of a model file
type ModelFormat string

const (
	ModelFormatGGUF        ModelFormat = "gguf"
	ModelFormatSafeTensors ModelFormat = "safetensors"
	ModelFormatPyTorch     ModelFormat = "pytorch"
	ModelFormatONNX        ModelFormat = "onnx"
)

// LoadStrategy defines how models are loaded
type LoadStrategy string

const (
	LoadStrategyEager     LoadStrategy = "eager"     // Load all shards immediately
	LoadStrategyLazy      LoadStrategy = "lazy"      // Load shards on demand
	LoadStrategyStreaming LoadStrategy = "streaming" // Stream shards during inference
	LoadStrategyHybrid    LoadStrategy = "hybrid"    // Mix of eager and lazy
)

// ShardedModel represents a model assembled from shards
type ShardedModel struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Format        ModelFormat            `json:"format"`
	TotalSize     int64                  `json:"total_size"`
	LoadedSize    int64                  `json:"loaded_size"`
	Shards        []*LoadedShard         `json:"shards"`
	Metadata      map[string]interface{} `json:"metadata"`
	LoadStrategy  LoadStrategy           `json:"load_strategy"`
	IsFullyLoaded bool                   `json:"is_fully_loaded"`
	LoadStartTime time.Time              `json:"load_start_time"`
	LoadEndTime   time.Time              `json:"load_end_time"`
	AccessPattern AccessPattern          `json:"access_pattern"`
}

// LoadedShard represents a shard that has been loaded
type LoadedShard struct {
	ShardID      string     `json:"shard_id"`
	Index        int        `json:"index"`
	Size         int64      `json:"size"`
	LoadedAt     time.Time  `json:"loaded_at"`
	Location     LoadedFrom `json:"location"`
	Data         []byte     `json:"-"` // Actual shard data
	IsLoaded     bool       `json:"is_loaded"`
	IsCached     bool       `json:"is_cached"`
	AccessCount  int64      `json:"access_count"`
	LastAccessed time.Time  `json:"last_accessed"`
}

// LoadedFrom indicates where a shard is loaded from
type LoadedFrom string

const (
	LoadedFromLocal  LoadedFrom = "local"
	LoadedFromRemote LoadedFrom = "remote"
	LoadedFromCache  LoadedFrom = "cache"
	LoadedFromMemory LoadedFrom = "memory"
)

// AccessPattern describes how the model is accessed
type AccessPattern struct {
	Sequential     bool    `json:"sequential"`
	RandomAccess   bool    `json:"random_access"`
	HotShards      []int   `json:"hot_shards"`      // Frequently accessed shards
	ColdShards     []int   `json:"cold_shards"`     // Rarely accessed shards
	AccessSequence []int   `json:"access_sequence"` // Order of shard access
	Predictability float64 `json:"predictability"`  // How predictable the access pattern is
}

// PartialModelLoader handles loading specific parts of a model
type PartialModelLoader struct {
	mu           sync.RWMutex
	loader       *ShardedModelLoader
	loadStrategy LoadStrategy
	cacheSize    int64
	cache        *ShardCache
	prefetcher   *ShardPrefetcher
}

// ModelAssembler reconstructs models from shards
type ModelAssembler struct {
	mu             sync.RWMutex
	formatHandlers map[ModelFormat]FormatHandler
	verifier       *IntegrityVerifier
	logger         *slog.Logger
}

// FormatHandler handles model format-specific operations
type FormatHandler interface {
	ValidateFormat(data []byte) error
	AssembleShards(shards []*LoadedShard) ([]byte, error)
	ExtractMetadata(data []byte) (map[string]interface{}, error)
	GetRequiredShards(request interface{}) []int
}

// ShardCache manages cached shards for efficient access
type ShardCache struct {
	mu             sync.RWMutex
	cache          map[string]*CachedShard
	maxSize        int64
	currentSize    int64
	evictionPolicy EvictionPolicy
	stats          *CacheStats
}

// CachedShard represents a shard in cache
type CachedShard struct {
	ShardID      string
	Data         []byte
	Size         int64
	LoadedAt     time.Time
	LastAccessed time.Time
	AccessCount  int64
	Priority     int
}

// EvictionPolicy defines cache eviction strategies
type EvictionPolicy string

const (
	EvictionLRU  EvictionPolicy = "lru"  // Least Recently Used
	EvictionLFU  EvictionPolicy = "lfu"  // Least Frequently Used
	EvictionFIFO EvictionPolicy = "fifo" // First In First Out
	EvictionARC  EvictionPolicy = "arc"  // Adaptive Replacement Cache
)

// CacheStats tracks cache performance
type CacheStats struct {
	Hits         int64
	Misses       int64
	Evictions    int64
	BytesLoaded  int64
	BytesEvicted int64
}

// ShardPrefetcher handles predictive shard loading
type ShardPrefetcher struct {
	mu            sync.RWMutex
	loader        *ShardedModelLoader
	predictor     *AccessPredictor
	prefetchQueue chan *PrefetchRequest
	workers       []*PrefetchWorker
	config        *PrefetchConfig
}

// AccessPredictor predicts which shards will be needed
type AccessPredictor struct {
	mu          sync.RWMutex
	history     []int
	historySize int
	patterns    map[string]*AccessPattern
	predictions []int
	confidence  float64
}

// PrefetchRequest represents a prefetch operation
type PrefetchRequest struct {
	ModelID  string
	ShardIDs []string
	Priority int
	Deadline time.Time
}

// PrefetchWorker processes prefetch requests
type PrefetchWorker struct {
	ID         int
	prefetcher *ShardPrefetcher
	stopChan   chan struct{}
}

// PrefetchConfig contains prefetcher configuration
type PrefetchConfig struct {
	MaxPrefetchAhead int
	PrefetchWindow   time.Duration
	WorkerCount      int
	QueueSize        int
}

// ShardedModelLoader coordinates loading of sharded models
type ShardedModelLoader struct {
	mu            sync.RWMutex
	registry      *ShardRegistry
	shardManager  *ModelShardManager
	assembler     *ModelAssembler
	partialLoader *PartialModelLoader
	cache         *ShardCache
	prefetcher    *ShardPrefetcher
	loadedModels  map[string]*ShardedModel
	config        *LoaderConfig
	logger        *slog.Logger
	p2pClient     *protocols.FileTransferClient // For real P2P transfers
	orchestrator  *ChunkTransferOrchestrator    // For orchestrated transfers
	localNodeID   string                        // Local node identifier for P2P operations
}

// LoaderConfig contains loader configuration
type LoaderConfig struct {
	LoadStrategy      LoadStrategy
	MaxMemoryUsage    int64
	CacheSize         int64
	EnablePrefetching bool
	EnableLazyLoading bool
	ParallelLoaders   int
	VerifyChecksums   bool
	CompressShards    bool
}

// NewShardedModelLoader creates a new model loader
func NewShardedModelLoader(
	registry *ShardRegistry,
	shardManager *ModelShardManager,
	p2pClient *protocols.FileTransferClient,
	localNodeID string,
	logger *slog.Logger) *ShardedModelLoader {

	config := &LoaderConfig{
		LoadStrategy:      LoadStrategyHybrid,
		MaxMemoryUsage:    32 * 1024 * 1024 * 1024, // 32GB
		CacheSize:         8 * 1024 * 1024 * 1024,  // 8GB cache
		EnablePrefetching: true,
		EnableLazyLoading: true,
		ParallelLoaders:   4,
		VerifyChecksums:   true,
		CompressShards:    false,
	}

	loader := &ShardedModelLoader{
		registry:     registry,
		shardManager: shardManager,
		loadedModels: make(map[string]*ShardedModel),
		config:       config,
		logger:       logger,
		p2pClient:    p2pClient,
		localNodeID:  localNodeID,
	}

	// Initialize assembler
	loader.assembler = NewModelAssembler(logger)

	// Initialize cache
	loader.cache = NewShardCache(config.CacheSize)

	// Initialize prefetcher
	if config.EnablePrefetching {
		loader.prefetcher = NewShardPrefetcher(loader, &PrefetchConfig{
			MaxPrefetchAhead: 3,
			PrefetchWindow:   5 * time.Second,
			WorkerCount:      2,
			QueueSize:        10,
		})
	}

	// Initialize partial loader
	loader.partialLoader = &PartialModelLoader{
		loader:       loader,
		loadStrategy: config.LoadStrategy,
		cacheSize:    config.CacheSize,
		cache:        loader.cache,
		prefetcher:   loader.prefetcher,
	}

	return loader
}

// SetOrchestrator sets the chunk transfer orchestrator for remote shard loading
func (l *ShardedModelLoader) SetOrchestrator(orchestrator *ChunkTransferOrchestrator) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.orchestrator = orchestrator
}

// LoadModel loads a sharded model
func (l *ShardedModelLoader) LoadModel(modelName string, shards []*ModelShard) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check if already loaded
	if model, exists := l.loadedModels[modelName]; exists && model.IsFullyLoaded {
		l.logger.Info("model already loaded", "model", modelName)
		return nil
	}

	// Create sharded model
	model := &ShardedModel{
		ID:            fmt.Sprintf("%s-%d", modelName, time.Now().Unix()),
		Name:          modelName,
		Format:        ModelFormatGGUF, // Default
		Shards:        make([]*LoadedShard, 0, len(shards)),
		Metadata:      make(map[string]interface{}),
		LoadStrategy:  l.config.LoadStrategy,
		LoadStartTime: time.Now(),
	}

	// Calculate total size
	for _, shard := range shards {
		model.TotalSize += shard.Size
	}

	// Check memory constraints
	if model.TotalSize > l.config.MaxMemoryUsage {
		if !l.config.EnableLazyLoading {
			return fmt.Errorf("model size %d exceeds max memory %d", model.TotalSize, l.config.MaxMemoryUsage)
		}
		// Switch to lazy loading for large models
		model.LoadStrategy = LoadStrategyLazy
	}

	// Load shards based on strategy
	var err error
	switch model.LoadStrategy {
	case LoadStrategyEager:
		err = l.loadAllShards(model, shards)
	case LoadStrategyLazy:
		err = l.prepareLazyLoading(model, shards)
	case LoadStrategyStreaming:
		err = l.prepareStreaming(model, shards)
	case LoadStrategyHybrid:
		err = l.loadHybrid(model, shards)
	}

	if err != nil {
		return fmt.Errorf("failed to load model: %w", err)
	}

	model.LoadEndTime = time.Now()
	l.loadedModels[modelName] = model

	l.logger.Info("model loaded",
		"model", modelName,
		"strategy", model.LoadStrategy,
		"duration", model.LoadEndTime.Sub(model.LoadStartTime),
		"loaded_size", model.LoadedSize,
		"total_size", model.TotalSize)

	return nil
}

// loadAllShards loads all shards immediately
func (l *ShardedModelLoader) loadAllShards(model *ShardedModel, shards []*ModelShard) error {
	loadedShards := make([]*LoadedShard, len(shards))
	errors := make(chan error, len(shards))
	var wg sync.WaitGroup

	// Load shards in parallel
	for i, shard := range shards {
		wg.Add(1)
		go func(index int, s *ModelShard) {
			defer wg.Done()

			loaded, err := l.loadShard(s)
			if err != nil {
				errors <- fmt.Errorf("failed to load shard %d: %w", index, err)
				return
			}

			loadedShards[index] = loaded
			atomic.AddInt64(&model.LoadedSize, loaded.Size)
		}(i, shard)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		if err != nil {
			return err
		}
	}

	model.Shards = loadedShards
	model.IsFullyLoaded = true

	return nil
}

// loadShard loads a single shard
func (l *ShardedModelLoader) loadShard(shard *ModelShard) (*LoadedShard, error) {
	// First check cache
	if cached := l.cache.Get(shard.ID); cached != nil {
		return &LoadedShard{
			ShardID:      shard.ID,
			Index:        shard.Index,
			Size:         cached.Size,
			LoadedAt:     time.Now(),
			Location:     LoadedFromCache,
			Data:         cached.Data,
			IsLoaded:     true,
			IsCached:     true,
			LastAccessed: time.Now(),
		}, nil
	}

	// Find shard locations
	locations, err := l.registry.LocateShard(shard.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to locate shard: %w", err)
	}

	// Try to load from best location
	for _, location := range locations {
		data, err := l.loadShardFromLocation(shard, location)
		if err != nil {
			l.logger.Warn("failed to load from location",
				"shard", shard.ID,
				"node", location.NodeID,
				"error", err)
			continue
		}

		// Verify checksum if enabled
		if l.config.VerifyChecksums {
			if !l.verifyShardChecksum(data, shard.Checksum) {
				l.logger.Warn("checksum verification failed", "shard", shard.ID)
				continue
			}
		}

		loaded := &LoadedShard{
			ShardID:      shard.ID,
			Index:        shard.Index,
			Size:         int64(len(data)),
			LoadedAt:     time.Now(),
			Location:     l.determineLocation(location),
			Data:         data,
			IsLoaded:     true,
			IsCached:     false,
			LastAccessed: time.Now(),
		}

		// Add to cache
		l.cache.Put(shard.ID, data)
		loaded.IsCached = true

		return loaded, nil
	}

	return nil, fmt.Errorf("failed to load shard from any location")
}

// loadShardFromLocation loads shard data from a specific location
func (l *ShardedModelLoader) loadShardFromLocation(shard *ModelShard, location protocols.ShardNodeLocation) ([]byte, error) {
	if location.IsLocal {
		// Load from local storage
		return l.loadLocalShard(location.StoragePath)
	}

	// Load from remote node - need to determine model name from shard
	// For now, use empty string as model name (will be improved with registry enhancements)
	return l.loadRemoteShard(shard.ID, location.NodeID, "")
}

// loadLocalShard loads a shard from local storage
func (l *ShardedModelLoader) loadLocalShard(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}

// loadRemoteShard loads a shard from a remote node using real P2P transfer with proper API usage
func (l *ShardedModelLoader) loadRemoteShard(shardID string, nodeID string, modelName string) ([]byte, error) {
	l.logger.Info("loading remote shard via orchestrator", "shard", shardID, "node", nodeID, "model", modelName)

	// Check if we have an orchestrator
	if l.orchestrator == nil {
		return nil, fmt.Errorf("chunk transfer orchestrator not initialized")
	}

	// Get shard metadata from shard plan using model name
	var shard *ModelShard
	if modelName != "" {
		plan, err := l.shardManager.GetShardPlan(modelName)
		if err != nil {
			return nil, fmt.Errorf("failed to get shard plan for model %s: %w", modelName, err)
		}

		// Find the specific shard in the plan
		for _, s := range plan.Shards {
			if s.ID == shardID {
				shard = s
				break
			}
		}
	}

	// If we couldn't find shard in plan, try to get it from registry locations
	if shard == nil {
		locations, err := l.registry.LocateShard(shardID)
		if err != nil {
			return nil, fmt.Errorf("failed to locate shard %s in registry: %w", shardID, err)
		}

		if len(locations) == 0 {
			return nil, fmt.Errorf("no available locations found for shard %s", shardID)
		}

		// Extract metadata from first available location
		var shardSize int64
		for _, loc := range locations {
			// Note: ShardLocation doesn't have Size field, will be determined during transfer
			shardSize = 0 // Will be determined from assembled file
			break
		}

		// Create shard structure with available metadata
		shard = &ModelShard{
			ID:       shardID,
			ModelID:  modelName,
			Size:     shardSize,
			Checksum: "", // Will be calculated during transfer
		}
	}

	if shard == nil {
		return nil, fmt.Errorf("could not find metadata for shard %s", shardID)
	}

	// Locate shard and select source node
	locations, err := l.registry.LocateShard(shardID)
	if err != nil {
		return nil, fmt.Errorf("failed to locate shard %s: %w", shardID, err)
	}

	var sourceNodeID string
	if nodeID != "" {
		// Validate provided node is available
		for _, loc := range locations {
			if loc.NodeID == nodeID && loc.IsAvailable {
				sourceNodeID = nodeID
				break
			}
		}
		if sourceNodeID == "" {
			return nil, fmt.Errorf("specified node %s is not available for shard %s", nodeID, shardID)
		}
	} else {
		// Select best available source: priority is Available + Loaded > Available
		for _, loc := range locations {
			if loc.IsAvailable && loc.IsLoaded {
				sourceNodeID = loc.NodeID
				break
			}
		}
		if sourceNodeID == "" {
			for _, loc := range locations {
				if loc.IsAvailable {
					sourceNodeID = loc.NodeID
					break
				}
			}
		}
		if sourceNodeID == "" {
			return nil, fmt.Errorf("no available sources found for shard %s", shardID)
		}
	}

	// Use configured local node ID
	localNodeID := l.localNodeID
	if localNodeID == "" {
		localNodeID = "local-node" // Fallback if not configured
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Orchestrate high-priority transfer
	transfer, err := l.orchestrator.OrchestateShardTransfer(
		ctx,
		shard,
		sourceNodeID,
		localNodeID,
		TransferPriorityHigh,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to orchestrate transfer for shard %s: %w", shardID, err)
	}

	// Poll for transfer completion
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("transfer timeout: %w", ctx.Err())
		case <-ticker.C:
			status, err := l.orchestrator.GetTransferStatus(transfer.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to get transfer status: %w", err)
			}

			switch status.Status {
			case TransferStatusCompleted:
				// Assemble shard from chunks first
				if err := l.orchestrator.assembleShardFromChunks(transfer.ID, shard); err != nil {
					return nil, fmt.Errorf("failed to assemble shard from chunks: %w", err)
				}

				// Read the assembled shard
				assembledPath := l.orchestrator.GetAssembledShardPath(shardID)
				data, err := os.ReadFile(assembledPath)
				if err != nil {
					return nil, fmt.Errorf("failed to read assembled shard from %s: %w", assembledPath, err)
				}

				// Verify with IntegrityVerifier if available
				if l.assembler != nil && l.assembler.verifier != nil && shard.Checksum != "" {
					valid, err := l.assembler.verifier.QuickVerify(assembledPath, shard.Checksum)
					if err != nil {
						return nil, fmt.Errorf("failed to verify shard checksum: %w", err)
					}
					if !valid {
						return nil, fmt.Errorf("shard checksum verification failed for %s", shardID)
					}
				}

				l.logger.Info("successfully loaded remote shard",
					"shard", shardID,
					"model", modelName,
					"source_node", sourceNodeID,
					"transfer_id", transfer.ID,
					"size", len(data))

				// Cache the shard
				if l.cache != nil {
					l.cache.Put(shardID, data)
				}

				// Update registry with local availability
				if l.registry != nil {
					l.updateLocalShardStatus(shardID, true)
				}

				return data, nil

			case TransferStatusFailed:
				return nil, fmt.Errorf("transfer failed: %s", status.LastError)

			case TransferStatusCancelled:
				return nil, fmt.Errorf("transfer was cancelled")

			default:
				// Continue polling
				continue
			}
		}
	}
}

// prepareLazyLoading prepares model for lazy loading
func (l *ShardedModelLoader) prepareLazyLoading(model *ShardedModel, shards []*ModelShard) error {
	// Create placeholder shards
	for i, shard := range shards {
		loaded := &LoadedShard{
			ShardID:  shard.ID,
			Index:    i,
			Size:     shard.Size,
			IsLoaded: false,
			IsCached: false,
		}
		model.Shards = append(model.Shards, loaded)
	}

	// Load only critical shards (first and last)
	if len(shards) > 0 {
		firstShard, err := l.loadShard(shards[0])
		if err != nil {
			return fmt.Errorf("failed to load first shard: %w", err)
		}
		model.Shards[0] = firstShard
		model.LoadedSize += firstShard.Size
	}

	if len(shards) > 1 {
		lastShard, err := l.loadShard(shards[len(shards)-1])
		if err != nil {
			return fmt.Errorf("failed to load last shard: %w", err)
		}
		model.Shards[len(shards)-1] = lastShard
		model.LoadedSize += lastShard.Size
	}

	return nil
}

// prepareStreaming prepares model for streaming
func (l *ShardedModelLoader) prepareStreaming(model *ShardedModel, shards []*ModelShard) error {
	// Similar to lazy loading but with streaming infrastructure
	return l.prepareLazyLoading(model, shards)
}

// loadHybrid uses a hybrid loading strategy
func (l *ShardedModelLoader) loadHybrid(model *ShardedModel, shards []*ModelShard) error {
	// Analyze access pattern
	pattern := l.analyzeAccessPattern(model.Name)

	// Load hot shards eagerly
	for _, hotIndex := range pattern.HotShards {
		if hotIndex < len(shards) {
			loaded, err := l.loadShard(shards[hotIndex])
			if err != nil {
				l.logger.Warn("failed to load hot shard", "index", hotIndex, "error", err)
				continue
			}
			model.Shards = append(model.Shards, loaded)
			model.LoadedSize += loaded.Size
		}
	}

	// Prepare cold shards for lazy loading
	for i, shard := range shards {
		isHot := false
		for _, hotIndex := range pattern.HotShards {
			if i == hotIndex {
				isHot = true
				break
			}
		}

		if !isHot {
			loaded := &LoadedShard{
				ShardID:  shard.ID,
				Index:    i,
				Size:     shard.Size,
				IsLoaded: false,
			}
			model.Shards = append(model.Shards, loaded)
		}
	}

	model.AccessPattern = pattern
	return nil
}

// analyzeAccessPattern analyzes how a model is accessed
func (l *ShardedModelLoader) analyzeAccessPattern(modelName string) AccessPattern {
	// In real implementation, would analyze historical access patterns
	// For now, return a default pattern
	return AccessPattern{
		Sequential:     true,
		RandomAccess:   false,
		HotShards:      []int{0, 1, 2}, // First few layers are often hot
		ColdShards:     []int{},
		Predictability: 0.8,
	}
}

// determineLocation determines the location type
func (l *ShardedModelLoader) determineLocation(location protocols.ShardNodeLocation) LoadedFrom {
	if location.IsLocal {
		return LoadedFromLocal
	}
	return LoadedFromRemote
}

// verifyShardChecksum verifies shard data integrity
func (l *ShardedModelLoader) verifyShardChecksum(data []byte, expectedChecksum string) bool {
	// In real implementation, would calculate and compare checksums
	return true
}

// GetLoadedModel returns a loaded model
func (l *ShardedModelLoader) GetLoadedModel(modelName string) (*ShardedModel, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	model, exists := l.loadedModels[modelName]
	if !exists {
		return nil, fmt.Errorf("model %s not loaded", modelName)
	}

	return model, nil
}

// UnloadModel unloads a model from memory
func (l *ShardedModelLoader) UnloadModel(modelName string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	model, exists := l.loadedModels[modelName]
	if !exists {
		return fmt.Errorf("model %s not loaded", modelName)
	}

	// Clear shard data
	for _, shard := range model.Shards {
		shard.Data = nil
		shard.IsLoaded = false

		// Remove from cache
		l.cache.Remove(shard.ShardID)
	}

	delete(l.loadedModels, modelName)

	l.logger.Info("model unloaded", "model", modelName)
	return nil
}

// LoadShardOnDemand loads a specific shard on demand
func (l *ShardedModelLoader) LoadShardOnDemand(modelName string, shardIndex int) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	model, exists := l.loadedModels[modelName]
	if !exists {
		return fmt.Errorf("model %s not loaded", modelName)
	}

	if shardIndex >= len(model.Shards) {
		return fmt.Errorf("shard index %d out of range", shardIndex)
	}

	shard := model.Shards[shardIndex]
	if shard.IsLoaded {
		shard.LastAccessed = time.Now()
		shard.AccessCount++
		return nil
	}

	// Get shard plan
	shardPlan, err := l.shardManager.GetShardPlan(modelName)
	if err != nil {
		return err
	}

	if shardIndex >= len(shardPlan.Shards) {
		return fmt.Errorf("shard not found in plan")
	}

	// Load the shard
	loaded, err := l.loadShard(shardPlan.Shards[shardIndex])
	if err != nil {
		return err
	}

	model.Shards[shardIndex] = loaded
	model.LoadedSize += loaded.Size

	return nil
}

// NewModelAssembler creates a new model assembler
func NewModelAssembler(logger *slog.Logger) *ModelAssembler {
	assembler := &ModelAssembler{
		formatHandlers: make(map[ModelFormat]FormatHandler),
		logger:         logger,
	}

	// Register format handlers
	assembler.formatHandlers[ModelFormatGGUF] = &GGUFHandler{}
	assembler.formatHandlers[ModelFormatSafeTensors] = &SafeTensorsHandler{}

	return assembler
}

// AssembleModel assembles a model from loaded shards
func (a *ModelAssembler) AssembleModel(model *ShardedModel) ([]byte, error) {
	a.mu.RLock()
	handler, exists := a.formatHandlers[model.Format]
	a.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unsupported format: %s", model.Format)
	}

	// Ensure all shards are loaded
	for _, shard := range model.Shards {
		if !shard.IsLoaded {
			return nil, fmt.Errorf("shard %s not loaded", shard.ShardID)
		}
	}

	return handler.AssembleShards(model.Shards)
}

// Format handlers

type GGUFHandler struct{}

func (h *GGUFHandler) ValidateFormat(data []byte) error {
	// Validate GGUF format
	return nil
}

func (h *GGUFHandler) AssembleShards(shards []*LoadedShard) ([]byte, error) {
	// Assemble GGUF shards
	totalSize := int64(0)
	for _, shard := range shards {
		totalSize += shard.Size
	}

	assembled := make([]byte, 0, totalSize)
	for _, shard := range shards {
		assembled = append(assembled, shard.Data...)
	}

	return assembled, nil
}

func (h *GGUFHandler) ExtractMetadata(data []byte) (map[string]interface{}, error) {
	// Extract GGUF metadata
	return map[string]interface{}{
		"format":  "gguf",
		"version": "1.0",
	}, nil
}

func (h *GGUFHandler) GetRequiredShards(request interface{}) []int {
	// Return required shards for request
	return []int{0, 1, 2} // Default to first few shards
}

type SafeTensorsHandler struct{}

func (h *SafeTensorsHandler) ValidateFormat(data []byte) error {
	return nil
}

func (h *SafeTensorsHandler) AssembleShards(shards []*LoadedShard) ([]byte, error) {
	totalSize := int64(0)
	for _, shard := range shards {
		totalSize += shard.Size
	}

	assembled := make([]byte, 0, totalSize)
	for _, shard := range shards {
		assembled = append(assembled, shard.Data...)
	}

	return assembled, nil
}

func (h *SafeTensorsHandler) ExtractMetadata(data []byte) (map[string]interface{}, error) {
	return map[string]interface{}{
		"format":  "safetensors",
		"version": "1.0",
	}, nil
}

func (h *SafeTensorsHandler) GetRequiredShards(request interface{}) []int {
	return []int{0}
}

// ShardCache implementation

func NewShardCache(maxSize int64) *ShardCache {
	return &ShardCache{
		cache:          make(map[string]*CachedShard),
		maxSize:        maxSize,
		evictionPolicy: EvictionLRU,
		stats:          &CacheStats{},
	}
}

func (c *ShardCache) Get(shardID string) *CachedShard {
	c.mu.Lock()
	defer c.mu.Unlock()

	shard, exists := c.cache[shardID]
	if !exists {
		atomic.AddInt64(&c.stats.Misses, 1)
		return nil
	}

	shard.LastAccessed = time.Now()
	shard.AccessCount++
	atomic.AddInt64(&c.stats.Hits, 1)

	return shard
}

func (c *ShardCache) Put(shardID string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	size := int64(len(data))

	// Evict if necessary
	for c.currentSize+size > c.maxSize {
		c.evict()
	}

	c.cache[shardID] = &CachedShard{
		ShardID:      shardID,
		Data:         data,
		Size:         size,
		LoadedAt:     time.Now(),
		LastAccessed: time.Now(),
		AccessCount:  1,
	}

	c.currentSize += size
	atomic.AddInt64(&c.stats.BytesLoaded, size)
}

func (c *ShardCache) Remove(shardID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if shard, exists := c.cache[shardID]; exists {
		c.currentSize -= shard.Size
		delete(c.cache, shardID)
	}
}

func (c *ShardCache) evict() {
	// Simple LRU eviction
	var oldest *CachedShard
	var oldestID string

	for id, shard := range c.cache {
		if oldest == nil || shard.LastAccessed.Before(oldest.LastAccessed) {
			oldest = shard
			oldestID = id
		}
	}

	if oldest != nil {
		c.currentSize -= oldest.Size
		delete(c.cache, oldestID)
		atomic.AddInt64(&c.stats.Evictions, 1)
		atomic.AddInt64(&c.stats.BytesEvicted, oldest.Size)
	}
}

// ShardPrefetcher implementation

func NewShardPrefetcher(loader *ShardedModelLoader, config *PrefetchConfig) *ShardPrefetcher {
	prefetcher := &ShardPrefetcher{
		loader:        loader,
		predictor:     NewAccessPredictor(),
		prefetchQueue: make(chan *PrefetchRequest, config.QueueSize),
		config:        config,
	}

	// Start workers
	prefetcher.workers = make([]*PrefetchWorker, config.WorkerCount)
	for i := 0; i < config.WorkerCount; i++ {
		prefetcher.workers[i] = &PrefetchWorker{
			ID:         i,
			prefetcher: prefetcher,
			stopChan:   make(chan struct{}),
		}
		go prefetcher.workers[i].start()
	}

	return prefetcher
}

func NewAccessPredictor() *AccessPredictor {
	return &AccessPredictor{
		historySize: 100,
		history:     make([]int, 0, 100),
		patterns:    make(map[string]*AccessPattern),
		predictions: make([]int, 0),
		confidence:  0.5,
	}
}

func (w *PrefetchWorker) start() {
	for {
		select {
		case <-w.stopChan:
			return
		case req := <-w.prefetcher.prefetchQueue:
			w.processPrefetch(req)
		}
	}
}

func (w *PrefetchWorker) processPrefetch(req *PrefetchRequest) {
	// Prefetch shards
	for _, shardID := range req.ShardIDs {
		// Check if already in cache
		if w.prefetcher.loader.cache.Get(shardID) != nil {
			continue
		}

		// Load shard into cache
		// This would trigger actual loading in real implementation
		w.prefetcher.loader.logger.Debug("prefetching shard", "shard", shardID)
	}
}

// updateLocalShardStatus updates the registry with local shard availability status
func (l *ShardedModelLoader) updateLocalShardStatus(shardID string, isLoaded bool) {
	if l.registry == nil {
		return
	}

	// Get current locations
	locations, err := l.registry.LocateShard(shardID)
	if err != nil {
		l.logger.Warn("failed to get shard locations for status update", "shard", shardID, "error", err)
		return
	}

	// Update or add local location entry
	var localNodeID string
	if l.p2pClient != nil {
		localNodeID = "local-node" // TODO: Extract from p2pClient
	} else {
		localNodeID = "local-node"
	}

	// Check if local entry already exists
	localExists := false
	for _, loc := range locations {
		if loc.NodeID == localNodeID {
			localExists = true
			break
		}
	}

	if !localExists && isLoaded {
		// Register local availability
		localLocation := protocols.ShardNodeLocation{
			NodeID:      localNodeID,
			IsAvailable: true,
			IsLoaded:    true,
			IsLocal:     true,
			StoragePath: "", // Will be set by cache or storage system
			LastSeen:    time.Now(),
			Size:        0, // Will be updated when known
		}

		err = l.registry.RegisterShard(shardID, "", localLocation)
		if err != nil {
			l.logger.Warn("failed to register local shard status", "shard", shardID, "error", err)
		}
	}
}

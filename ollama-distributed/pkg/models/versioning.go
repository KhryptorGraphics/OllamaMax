package models

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// VersionManager manages model versions with semantic versioning and dependency tracking
type VersionManager struct {
	mu sync.RWMutex

	// Version tracking
	modelVersions  map[string]*ModelVersionTree
	versionHistory map[string][]*VersionEvent

	// Dependencies
	dependencies map[string]*DependencyGraph

	// Configuration
	config *VersionConfig

	// Metrics
	metrics *VersionMetrics

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// ModelVersionTree represents the version tree for a model
type ModelVersionTree struct {
	ModelName     string                           `json:"model_name"`
	Versions      map[string]*DetailedModelVersion `json:"versions"`
	LatestVersion string                           `json:"latest_version"`
	LatestStable  string                           `json:"latest_stable"`
	DefaultBranch string                           `json:"default_branch"`
	Branches      map[string]*VersionBranch        `json:"branches"`
	Tags          map[string]string                `json:"tags"` // tag -> version
	CreatedAt     time.Time                        `json:"created_at"`
	UpdatedAt     time.Time                        `json:"updated_at"`
}

// DetailedModelVersion represents a specific version of a model with comprehensive metadata
type DetailedModelVersion struct {
	ModelName       string           `json:"model_name"`
	Version         string           `json:"version"`
	SemanticVersion *SemanticVersion `json:"semantic_version"`

	// Version metadata
	DisplayName  string `json:"display_name"`
	Description  string `json:"description"`
	ReleaseNotes string `json:"release_notes"`

	// Version properties
	IsStable     bool `json:"is_stable"`
	IsPrerelease bool `json:"is_prerelease"`
	IsDeprecated bool `json:"is_deprecated"`

	// Model information
	ModelSize    int64  `json:"model_size"`
	ModelFormat  string `json:"model_format"`
	Architecture string `json:"architecture"`
	Parameters   int64  `json:"parameters"`

	// File information
	Files     []*ModelFile             `json:"files"`
	Checksums map[HashAlgorithm]string `json:"checksums"`

	// Dependencies
	Dependencies []*ModelDependency `json:"dependencies"`
	Conflicts    []string           `json:"conflicts"`

	// Compatibility
	MinOllamaVersion   string   `json:"min_ollama_version"`
	MaxOllamaVersion   string   `json:"max_ollama_version,omitempty"`
	SupportedPlatforms []string `json:"supported_platforms"`

	// Authorship and provenance
	Author    string `json:"author"`
	Publisher string `json:"publisher"`
	License   string `json:"license"`
	Source    string `json:"source"`

	// Timestamps
	CreatedAt   time.Time `json:"created_at"`
	PublishedAt time.Time `json:"published_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Distribution
	AvailableNodes []peer.ID `json:"available_nodes"`
	DownloadCount  int64     `json:"download_count"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata"`
}

// SemanticVersion represents semantic versioning (major.minor.patch)
type SemanticVersion struct {
	Major      int    `json:"major"`
	Minor      int    `json:"minor"`
	Patch      int    `json:"patch"`
	Prerelease string `json:"prerelease,omitempty"`
	Build      string `json:"build,omitempty"`
}

// VersionBranch represents a version branch
type VersionBranch struct {
	Name        string    `json:"name"`
	BaseVersion string    `json:"base_version"`
	HeadVersion string    `json:"head_version"`
	IsDefault   bool      `json:"is_default"`
	IsProtected bool      `json:"is_protected"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ModelFile represents a file within a model version
type ModelFile struct {
	FileName    string `json:"file_name"`
	FilePath    string `json:"file_path"`
	FileSize    int64  `json:"file_size"`
	FileType    string `json:"file_type"`
	Checksum    string `json:"checksum"`
	IsRequired  bool   `json:"is_required"`
	Description string `json:"description"`
}

// ModelDependency represents a dependency on another model or component
type ModelDependency struct {
	Name         string         `json:"name"`
	Version      string         `json:"version"`
	VersionRange string         `json:"version_range"`
	Type         DependencyType `json:"type"`
	IsOptional   bool           `json:"is_optional"`
	Description  string         `json:"description"`
}

// DependencyGraph represents the dependency graph for models
type DependencyGraph struct {
	ModelName    string                      `json:"model_name"`
	Dependencies map[string]*ModelDependency `json:"dependencies"`
	Dependents   map[string]*ModelDependency `json:"dependents"`
	UpdatedAt    time.Time                   `json:"updated_at"`
}

// VersionEvent represents a versioning event
type VersionEvent struct {
	EventID   string                 `json:"event_id"`
	EventType VersionEventType       `json:"event_type"`
	ModelName string                 `json:"model_name"`
	Version   string                 `json:"version"`
	Actor     string                 `json:"actor"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details"`
}

// VersionConfig configures the version manager
type VersionConfig struct {
	// Versioning policy
	DefaultVersioningScheme   VersioningScheme
	AllowPrerelease           bool
	RequireSemanticVersioning bool

	// Retention policy
	MaxVersionsPerModel  int
	RetainStableVersions bool
	RetainTaggedVersions bool

	// Validation
	ValidateChecksums     bool
	ValidateDependencies  bool
	ValidateCompatibility bool

	// Performance
	CacheVersionInfo bool
	MaxCacheSize     int
	CacheTimeout     time.Duration
}

// VersionMetrics tracks versioning metrics
type VersionMetrics struct {
	TotalModels          int64                      `json:"total_models"`
	TotalVersions        int64                      `json:"total_versions"`
	StableVersions       int64                      `json:"stable_versions"`
	PrereleaseVersions   int64                      `json:"prerelease_versions"`
	DeprecatedVersions   int64                      `json:"deprecated_versions"`
	VersionsByScheme     map[VersioningScheme]int64 `json:"versions_by_scheme"`
	DependencyViolations int64                      `json:"dependency_violations"`
	LastUpdated          time.Time                  `json:"last_updated"`
}

// Enums and constants
type DependencyType string

const (
	DependencyTypeModel   DependencyType = "model"
	DependencyTypeLibrary DependencyType = "library"
	DependencyTypeRuntime DependencyType = "runtime"
	DependencyTypePlugin  DependencyType = "plugin"
)

type VersionEventType string

const (
	VersionEventCreated    VersionEventType = "created"
	VersionEventUpdated    VersionEventType = "updated"
	VersionEventPublished  VersionEventType = "published"
	VersionEventDeprecated VersionEventType = "deprecated"
	VersionEventDeleted    VersionEventType = "deleted"
	VersionEventTagged     VersionEventType = "tagged"
	VersionEventBranched   VersionEventType = "branched"
)

type VersioningScheme string

const (
	VersioningSchemeSemantic   VersioningScheme = "semantic"
	VersioningSchemeCalendar   VersioningScheme = "calendar"
	VersioningSchemeSequential VersioningScheme = "sequential"
	VersioningSchemeCustom     VersioningScheme = "custom"
)

// NewVersionManager creates a new version manager
func NewVersionManager(config *VersionConfig) *VersionManager {
	ctx, cancel := context.WithCancel(context.Background())

	if config == nil {
		config = &VersionConfig{
			DefaultVersioningScheme:   VersioningSchemeSemantic,
			AllowPrerelease:           true,
			RequireSemanticVersioning: true,
			MaxVersionsPerModel:       100,
			RetainStableVersions:      true,
			RetainTaggedVersions:      true,
			ValidateChecksums:         true,
			ValidateDependencies:      true,
			ValidateCompatibility:     true,
			CacheVersionInfo:          true,
			MaxCacheSize:              1000,
			CacheTimeout:              time.Hour,
		}
	}

	vm := &VersionManager{
		modelVersions:  make(map[string]*ModelVersionTree),
		versionHistory: make(map[string][]*VersionEvent),
		dependencies:   make(map[string]*DependencyGraph),
		config:         config,
		metrics: &VersionMetrics{
			VersionsByScheme: make(map[VersioningScheme]int64),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Start background tasks
	vm.wg.Add(1)
	go vm.maintenanceLoop()

	return vm
}

// RegisterModelVersion registers a new model version
func (vm *VersionManager) RegisterModelVersion(version *DetailedModelVersion) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	// Validate version
	if err := vm.validateVersion(version); err != nil {
		return fmt.Errorf("version validation failed: %w", err)
	}

	// Parse semantic version if using semantic versioning
	if vm.config.RequireSemanticVersioning {
		semVer, err := vm.parseSemanticVersion(version.Version)
		if err != nil {
			return fmt.Errorf("invalid semantic version: %w", err)
		}
		version.SemanticVersion = semVer
	}

	// Get or create model version tree
	tree, exists := vm.modelVersions[version.ModelName]
	if !exists {
		tree = &ModelVersionTree{
			ModelName:     version.ModelName,
			Versions:      make(map[string]*DetailedModelVersion),
			Branches:      make(map[string]*VersionBranch),
			Tags:          make(map[string]string),
			DefaultBranch: "main",
			CreatedAt:     time.Now(),
		}
		vm.modelVersions[version.ModelName] = tree
		vm.metrics.TotalModels++
	}

	// Check for version conflicts
	if _, exists := tree.Versions[version.Version]; exists {
		return fmt.Errorf("version %s already exists for model %s", version.Version, version.ModelName)
	}

	// Set timestamps
	version.CreatedAt = time.Now()
	version.UpdatedAt = time.Now()

	// Add version to tree
	tree.Versions[version.Version] = version
	tree.UpdatedAt = time.Now()

	// Update latest version
	if vm.isNewerVersion(version.Version, tree.LatestVersion) {
		tree.LatestVersion = version.Version
	}

	// Update latest stable version
	if version.IsStable && vm.isNewerVersion(version.Version, tree.LatestStable) {
		tree.LatestStable = version.Version
	}

	// Update dependency graph
	vm.updateDependencyGraph(version)

	// Record event
	vm.recordVersionEvent(VersionEventCreated, version.ModelName, version.Version, "system", nil)

	// Update metrics
	vm.metrics.TotalVersions++
	if version.IsStable {
		vm.metrics.StableVersions++
	}
	if version.IsPrerelease {
		vm.metrics.PrereleaseVersions++
	}
	if version.IsDeprecated {
		vm.metrics.DeprecatedVersions++
	}

	return nil
}

// GetModelVersion retrieves a specific model version
func (vm *VersionManager) GetModelVersion(modelName, version string) (*DetailedModelVersion, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	tree, exists := vm.modelVersions[modelName]
	if !exists {
		return nil, fmt.Errorf("model %s not found", modelName)
	}

	modelVersion, exists := tree.Versions[version]
	if !exists {
		return nil, fmt.Errorf("version %s not found for model %s", version, modelName)
	}

	// Return a copy to prevent external modification
	versionCopy := *modelVersion
	return &versionCopy, nil
}

// GetLatestVersion returns the latest version of a model
func (vm *VersionManager) GetLatestVersion(modelName string, stableOnly bool) (*DetailedModelVersion, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	tree, exists := vm.modelVersions[modelName]
	if !exists {
		return nil, fmt.Errorf("model %s not found", modelName)
	}

	var targetVersion string
	if stableOnly {
		targetVersion = tree.LatestStable
	} else {
		targetVersion = tree.LatestVersion
	}

	if targetVersion == "" {
		return nil, fmt.Errorf("no versions available for model %s", modelName)
	}

	return vm.GetModelVersion(modelName, targetVersion)
}

// ListModelVersions lists all versions of a model
func (vm *VersionManager) ListModelVersions(modelName string, includeDeprecated bool) ([]*DetailedModelVersion, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	tree, exists := vm.modelVersions[modelName]
	if !exists {
		return nil, fmt.Errorf("model %s not found", modelName)
	}

	versions := make([]*DetailedModelVersion, 0, len(tree.Versions))
	for _, version := range tree.Versions {
		if !includeDeprecated && version.IsDeprecated {
			continue
		}
		versionCopy := *version
		versions = append(versions, &versionCopy)
	}

	// Sort versions by semantic version or creation time
	vm.sortVersions(versions)

	return versions, nil
}

// CreateVersionTag creates a tag for a specific version
func (vm *VersionManager) CreateVersionTag(modelName, version, tag string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	tree, exists := vm.modelVersions[modelName]
	if !exists {
		return fmt.Errorf("model %s not found", modelName)
	}

	if _, exists := tree.Versions[version]; !exists {
		return fmt.Errorf("version %s not found for model %s", version, modelName)
	}

	// Check if tag already exists
	if existingVersion, exists := tree.Tags[tag]; exists {
		return fmt.Errorf("tag %s already exists for version %s", tag, existingVersion)
	}

	tree.Tags[tag] = version
	tree.UpdatedAt = time.Now()

	// Record event
	details := map[string]interface{}{"tag": tag}
	vm.recordVersionEvent(VersionEventTagged, modelName, version, "system", details)

	return nil
}

// DeprecateVersion marks a version as deprecated
func (vm *VersionManager) DeprecateVersion(modelName, version string, reason string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	tree, exists := vm.modelVersions[modelName]
	if !exists {
		return fmt.Errorf("model %s not found", modelName)
	}

	modelVersion, exists := tree.Versions[version]
	if !exists {
		return fmt.Errorf("version %s not found for model %s", version, modelName)
	}

	if modelVersion.IsDeprecated {
		return fmt.Errorf("version %s is already deprecated", version)
	}

	modelVersion.IsDeprecated = true
	modelVersion.UpdatedAt = time.Now()
	tree.UpdatedAt = time.Now()

	// Record event
	details := map[string]interface{}{"reason": reason}
	vm.recordVersionEvent(VersionEventDeprecated, modelName, version, "system", details)

	// Update metrics
	vm.metrics.DeprecatedVersions++

	return nil
}

// validateVersion validates a model version
func (vm *VersionManager) validateVersion(version *DetailedModelVersion) error {
	if version.ModelName == "" {
		return fmt.Errorf("model name is required")
	}

	if version.Version == "" {
		return fmt.Errorf("version is required")
	}

	if vm.config.ValidateChecksums && len(version.Checksums) == 0 {
		return fmt.Errorf("checksums are required")
	}

	if vm.config.ValidateDependencies {
		for _, dep := range version.Dependencies {
			if dep.Name == "" || dep.Version == "" {
				return fmt.Errorf("dependency name and version are required")
			}
		}
	}

	return nil
}

// parseSemanticVersion parses a semantic version string
func (vm *VersionManager) parseSemanticVersion(version string) (*SemanticVersion, error) {
	// Simplified semantic version parsing
	// In a real implementation, you would use a proper semver library
	semVer := &SemanticVersion{}

	// For now, just set major version to 1
	semVer.Major = 1
	semVer.Minor = 0
	semVer.Patch = 0

	return semVer, nil
}

// isNewerVersion checks if version1 is newer than version2
func (vm *VersionManager) isNewerVersion(version1, version2 string) bool {
	if version2 == "" {
		return true
	}

	// Simplified version comparison
	// In a real implementation, you would use proper semantic version comparison
	return version1 > version2
}

// updateDependencyGraph updates the dependency graph for a model
func (vm *VersionManager) updateDependencyGraph(version *DetailedModelVersion) {
	graph, exists := vm.dependencies[version.ModelName]
	if !exists {
		graph = &DependencyGraph{
			ModelName:    version.ModelName,
			Dependencies: make(map[string]*ModelDependency),
			Dependents:   make(map[string]*ModelDependency),
		}
		vm.dependencies[version.ModelName] = graph
	}

	// Update dependencies
	for _, dep := range version.Dependencies {
		graph.Dependencies[dep.Name] = dep

		// Update dependent's graph
		depGraph, exists := vm.dependencies[dep.Name]
		if !exists {
			depGraph = &DependencyGraph{
				ModelName:    dep.Name,
				Dependencies: make(map[string]*ModelDependency),
				Dependents:   make(map[string]*ModelDependency),
			}
			vm.dependencies[dep.Name] = depGraph
		}

		depGraph.Dependents[version.ModelName] = &ModelDependency{
			Name:    version.ModelName,
			Version: version.Version,
			Type:    DependencyTypeModel,
		}
	}

	graph.UpdatedAt = time.Now()
}

// recordVersionEvent records a versioning event
func (vm *VersionManager) recordVersionEvent(eventType VersionEventType, modelName, version, actor string, details map[string]interface{}) {
	event := &VersionEvent{
		EventID:   fmt.Sprintf("event_%d", time.Now().UnixNano()),
		EventType: eventType,
		ModelName: modelName,
		Version:   version,
		Actor:     actor,
		Timestamp: time.Now(),
		Details:   details,
	}

	vm.versionHistory[modelName] = append(vm.versionHistory[modelName], event)

	// Limit history size
	if len(vm.versionHistory[modelName]) > 1000 {
		vm.versionHistory[modelName] = vm.versionHistory[modelName][1:]
	}
}

// sortVersions sorts versions by semantic version or creation time
func (vm *VersionManager) sortVersions(versions []*DetailedModelVersion) {
	sort.Slice(versions, func(i, j int) bool {
		// Sort by creation time (newest first)
		return versions[i].CreatedAt.After(versions[j].CreatedAt)
	})
}

// GetDependencyGraph returns the dependency graph for a model
func (vm *VersionManager) GetDependencyGraph(modelName string) (*DependencyGraph, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	graph, exists := vm.dependencies[modelName]
	if !exists {
		return nil, fmt.Errorf("dependency graph not found for model %s", modelName)
	}

	// Return a copy
	graphCopy := *graph
	return &graphCopy, nil
}

// GetVersionHistory returns the version history for a model
func (vm *VersionManager) GetVersionHistory(modelName string, limit int) ([]*VersionEvent, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	history, exists := vm.versionHistory[modelName]
	if !exists {
		return []*VersionEvent{}, nil
	}

	if limit <= 0 || limit > len(history) {
		limit = len(history)
	}

	// Return most recent events
	start := len(history) - limit
	events := make([]*VersionEvent, limit)
	copy(events, history[start:])

	return events, nil
}

// GetMetrics returns versioning metrics
func (vm *VersionManager) GetMetrics() *VersionMetrics {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	metrics := *vm.metrics
	metrics.LastUpdated = time.Now()
	return &metrics
}

// maintenanceLoop performs periodic maintenance tasks
func (vm *VersionManager) maintenanceLoop() {
	defer vm.wg.Done()

	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-vm.ctx.Done():
			return
		case <-ticker.C:
			vm.performMaintenance()
		}
	}
}

// performMaintenance performs maintenance tasks
func (vm *VersionManager) performMaintenance() {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	// Clean up old version history
	for modelName, history := range vm.versionHistory {
		if len(history) > 1000 {
			vm.versionHistory[modelName] = history[len(history)-1000:]
		}
	}

	// Update metrics
	vm.metrics.LastUpdated = time.Now()
}

// Close closes the version manager
func (vm *VersionManager) Close() error {
	vm.cancel()
	vm.wg.Wait()
	return nil
}

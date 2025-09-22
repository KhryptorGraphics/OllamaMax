package models

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// VersionConfig contains configuration for version management
type VersionConfig struct {
	StoragePath      string
	MaxVersions      int
	RetentionPeriod  time.Duration
	EnableAutoCleanup bool
	CleanupInterval  time.Duration
}

// BasicModelVersion represents a specific version of a model for basic version management
type BasicModelVersion struct {
	Version      string                 `json:"version"`
	ModelName    string                 `json:"model_name"`
	CreatedAt    time.Time              `json:"created_at"`
	Size         int64                  `json:"size"`
	Checksum     string                 `json:"checksum"`
	IsActive     bool                   `json:"is_active"`
	IsDeprecated bool                   `json:"is_deprecated"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// BasicVersionManager manages model versions with basic functionality
type BasicVersionManager struct {
	mu       sync.RWMutex
	config   *VersionConfig
	logger   *slog.Logger
	versions map[string][]*BasicModelVersion // modelName -> versions
	active   map[string]string           // modelName -> active version
}

// NewBasicVersionManager creates a new basic version manager
func NewBasicVersionManager(config *VersionConfig) *BasicVersionManager {
	if config == nil {
		config = &VersionConfig{
			StoragePath:      "/tmp/models",
			MaxVersions:      10,
			RetentionPeriod:  30 * 24 * time.Hour,
			EnableAutoCleanup: true,
			CleanupInterval:  24 * time.Hour,
		}
	}

	vm := &BasicVersionManager{
		config:   config,
		logger:   slog.Default().With("component", "version_manager"),
		versions: make(map[string][]*ModelVersion),
		active:   make(map[string]string),
	}

	if config.EnableAutoCleanup {
		go vm.startCleanupWorker()
	}

	return vm
}

// RegisterVersion registers a new model version
func (vm *BasicVersionManager) RegisterVersion(modelName, version string, metadata map[string]interface{}) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	newVersion := &BasicModelVersion{
		Version:   version,
		ModelName: modelName,
		CreatedAt: time.Now(),
		Metadata:  metadata,
	}

	if _, exists := vm.versions[modelName]; !exists {
		vm.versions[modelName] = make([]*BasicModelVersion, 0)
	}

	vm.versions[modelName] = append(vm.versions[modelName], newVersion)

	// Enforce max versions limit
	if len(vm.versions[modelName]) > vm.config.MaxVersions {
		// Remove oldest version
		vm.versions[modelName] = vm.versions[modelName][1:]
	}

	vm.logger.Info("Registered model version",
		"model", modelName,
		"version", version,
		"total_versions", len(vm.versions[modelName]))

	return nil
}

// SetActiveVersion sets the active version for a model
func (vm *BasicVersionManager) SetActiveVersion(modelName, version string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	versions, exists := vm.versions[modelName]
	if !exists {
		return fmt.Errorf("model %s not found", modelName)
	}

	found := false
	for _, v := range versions {
		if v.Version == version {
			v.IsActive = true
			found = true
		} else {
			v.IsActive = false
		}
	}

	if !found {
		return fmt.Errorf("version %s not found for model %s", version, modelName)
	}

	vm.active[modelName] = version
	vm.logger.Info("Set active version", "model", modelName, "version", version)
	return nil
}

// GetActiveVersion returns the active version for a model
func (vm *BasicVersionManager) GetActiveVersion(modelName string) (string, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	version, exists := vm.active[modelName]
	if !exists {
		return "", fmt.Errorf("no active version for model %s", modelName)
	}
	return version, nil
}

// ListVersions lists all versions for a model
func (vm *BasicVersionManager) ListVersions(modelName string) ([]*BasicModelVersion, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	versions, exists := vm.versions[modelName]
	if !exists {
		return nil, fmt.Errorf("model %s not found", modelName)
	}

	// Return a copy to prevent external modification
	result := make([]*BasicModelVersion, len(versions))
	copy(result, versions)
	return result, nil
}

// DeleteVersion deletes a specific version
func (vm *BasicVersionManager) DeleteVersion(modelName, version string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	versions, exists := vm.versions[modelName]
	if !exists {
		return fmt.Errorf("model %s not found", modelName)
	}

	newVersions := make([]*BasicModelVersion, 0, len(versions)-1)
	found := false
	for _, v := range versions {
		if v.Version != version {
			newVersions = append(newVersions, v)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("version %s not found for model %s", version, modelName)
	}

	vm.versions[modelName] = newVersions

	// Update active version if necessary
	if vm.active[modelName] == version {
		delete(vm.active, modelName)
		// Optionally set the latest version as active
		if len(newVersions) > 0 {
			vm.active[modelName] = newVersions[len(newVersions)-1].Version
		}
	}

	vm.logger.Info("Deleted version", "model", modelName, "version", version)
	return nil
}

// CleanupOldVersions removes versions older than retention period
func (vm *BasicVersionManager) CleanupOldVersions() error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	cutoff := time.Now().Add(-vm.config.RetentionPeriod)
	cleanedCount := 0

	for modelName, versions := range vm.versions {
		newVersions := make([]*BasicModelVersion, 0)
		for _, v := range versions {
			if v.CreatedAt.After(cutoff) || v.IsActive {
				newVersions = append(newVersions, v)
			} else {
				cleanedCount++
			}
		}
		vm.versions[modelName] = newVersions
	}

	if cleanedCount > 0 {
		vm.logger.Info("Cleaned up old versions", "count", cleanedCount)
	}

	return nil
}

// startCleanupWorker starts a background worker for version cleanup
func (vm *BasicVersionManager) startCleanupWorker() {
	ticker := time.NewTicker(vm.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := vm.CleanupOldVersions(); err != nil {
			vm.logger.Error("Failed to cleanup old versions", "error", err)
		}
	}
}

// GetVersionMetadata returns metadata for a specific version
func (vm *BasicVersionManager) GetVersionMetadata(modelName, version string) (map[string]interface{}, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	versions, exists := vm.versions[modelName]
	if !exists {
		return nil, fmt.Errorf("model %s not found", modelName)
	}

	for _, v := range versions {
		if v.Version == version {
			return v.Metadata, nil
		}
	}

	return nil, fmt.Errorf("version %s not found for model %s", version, modelName)
}

// UpdateVersionMetadata updates metadata for a version
func (vm *BasicVersionManager) UpdateVersionMetadata(modelName, version string, metadata map[string]interface{}) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	versions, exists := vm.versions[modelName]
	if !exists {
		return fmt.Errorf("model %s not found", modelName)
	}

	for _, v := range versions {
		if v.Version == version {
			// Merge metadata
			if v.Metadata == nil {
				v.Metadata = make(map[string]interface{})
			}
			for key, value := range metadata {
				v.Metadata[key] = value
			}
			return nil
		}
	}

	return fmt.Errorf("version %s not found for model %s", version, modelName)
}

// MarkDeprecated marks a version as deprecated
func (vm *BasicVersionManager) MarkDeprecated(modelName, version string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	versions, exists := vm.versions[modelName]
	if !exists {
		return fmt.Errorf("model %s not found", modelName)
	}

	for _, v := range versions {
		if v.Version == version {
			v.IsDeprecated = true
			vm.logger.Info("Marked version as deprecated", "model", modelName, "version", version)
			return nil
		}
	}

	return fmt.Errorf("version %s not found for model %s", version, modelName)
}

// GetLatestVersion returns the latest version for a model
func (vm *BasicVersionManager) GetLatestVersion(modelName string) (*BasicModelVersion, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	versions, exists := vm.versions[modelName]
	if !exists || len(versions) == 0 {
		return nil, fmt.Errorf("no versions found for model %s", modelName)
	}

	// Return the last version (assumed to be the latest)
	return versions[len(versions)-1], nil
}

// ValidateVersion checks if a version exists and is valid
func (vm *BasicVersionManager) ValidateVersion(ctx context.Context, modelName, version string) error {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	versions, exists := vm.versions[modelName]
	if !exists {
		return fmt.Errorf("model %s not found", modelName)
	}

	for _, v := range versions {
		if v.Version == version {
			if v.IsDeprecated {
				vm.logger.Warn("Using deprecated version", "model", modelName, "version", version)
			}
			return nil
		}
	}

	return fmt.Errorf("version %s not found for model %s", version, modelName)
}
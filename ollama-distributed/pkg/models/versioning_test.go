package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSemanticVersion_Creation(t *testing.T) {
	tests := []struct {
		name    string
		version *SemanticVersion
	}{
		{
			name: "basic version",
			version: &SemanticVersion{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
		},
		{
			name: "version with prerelease",
			version: &SemanticVersion{
				Major:      1,
				Minor:      2,
				Patch:      3,
				Prerelease: "alpha.1",
			},
		},
		{
			name: "version with build",
			version: &SemanticVersion{
				Major: 1,
				Minor: 2,
				Patch: 3,
				Build: "20230101",
			},
		},
		{
			name: "version with prerelease and build",
			version: &SemanticVersion{
				Major:      1,
				Minor:      2,
				Patch:      3,
				Prerelease: "beta.2",
				Build:      "20230101",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.version)
			assert.GreaterOrEqual(t, tt.version.Major, 0)
			assert.GreaterOrEqual(t, tt.version.Minor, 0)
			assert.GreaterOrEqual(t, tt.version.Patch, 0)
		})
	}
}

func TestModelVersionTree_Creation(t *testing.T) {
	now := time.Now()
	tree := &ModelVersionTree{
		ModelName:     "test-model",
		Versions:      make(map[string]*DetailedModelVersion),
		LatestVersion: "1.0.0",
		LatestStable:  "1.0.0",
		DefaultBranch: "main",
		Branches:      make(map[string]*VersionBranch),
		Tags:          make(map[string]string),
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	assert.Equal(t, "test-model", tree.ModelName)
	assert.Equal(t, "1.0.0", tree.LatestVersion)
	assert.Equal(t, "1.0.0", tree.LatestStable)
	assert.Equal(t, "main", tree.DefaultBranch)
	assert.NotNil(t, tree.Versions)
	assert.NotNil(t, tree.Branches)
	assert.NotNil(t, tree.Tags)
	assert.False(t, tree.CreatedAt.IsZero())
	assert.False(t, tree.UpdatedAt.IsZero())
}

func TestDetailedModelVersion_Creation(t *testing.T) {
	now := time.Now()
	version := &DetailedModelVersion{
		ModelName: "test-model",
		Version:   "1.0.0",
		SemanticVersion: &SemanticVersion{
			Major: 1,
			Minor: 0,
			Patch: 0,
		},
		IsStable:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
		ModelSize:   1024,
		Description: "Test model version",
		Metadata:    make(map[string]interface{}),
	}

	assert.Equal(t, "test-model", version.ModelName)
	assert.Equal(t, "1.0.0", version.Version)
	assert.NotNil(t, version.SemanticVersion)
	assert.Equal(t, 1, version.SemanticVersion.Major)
	assert.Equal(t, 0, version.SemanticVersion.Minor)
	assert.Equal(t, 0, version.SemanticVersion.Patch)
	assert.True(t, version.IsStable)
	assert.Equal(t, int64(1024), version.ModelSize)
	assert.Equal(t, "Test model version", version.Description)
	assert.NotNil(t, version.Metadata)
}

func TestVersionBranch_Creation(t *testing.T) {
	now := time.Now()
	branch := &VersionBranch{
		Name:        "feature-branch",
		BaseVersion: "1.0.0",
		HeadVersion: "1.1.0-alpha.1",
		IsDefault:   false,
		IsProtected: true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	assert.Equal(t, "feature-branch", branch.Name)
	assert.Equal(t, "1.0.0", branch.BaseVersion)
	assert.Equal(t, "1.1.0-alpha.1", branch.HeadVersion)
	assert.False(t, branch.IsDefault)
	assert.True(t, branch.IsProtected)
	assert.False(t, branch.CreatedAt.IsZero())
	assert.False(t, branch.UpdatedAt.IsZero())
}

func TestModelFile_Creation(t *testing.T) {
	file := &ModelFile{
		FileName:    "model.gguf",
		FilePath:    "/models/test-model/model.gguf",
		FileSize:    1048576,
		FileType:    "gguf",
		Checksum:    "sha256:abc123def456",
		IsRequired:  true,
		Description: "Main model file",
	}

	assert.Equal(t, "model.gguf", file.FileName)
	assert.Equal(t, "/models/test-model/model.gguf", file.FilePath)
	assert.Equal(t, int64(1048576), file.FileSize)
	assert.Equal(t, "gguf", file.FileType)
	assert.Equal(t, "sha256:abc123def456", file.Checksum)
	assert.True(t, file.IsRequired)
	assert.Equal(t, "Main model file", file.Description)
}

func TestModelDependency_Creation(t *testing.T) {
	dependency := &ModelDependency{
		Name:         "base-model",
		Version:      "1.0.0",
		VersionRange: ">=1.0.0,<2.0.0",
		Type:         DependencyTypeModel,
		IsOptional:   false,
		Description:  "Base model dependency",
	}

	assert.Equal(t, "base-model", dependency.Name)
	assert.Equal(t, "1.0.0", dependency.Version)
	assert.Equal(t, ">=1.0.0,<2.0.0", dependency.VersionRange)
	assert.Equal(t, DependencyTypeModel, dependency.Type)
	assert.False(t, dependency.IsOptional)
	assert.Equal(t, "Base model dependency", dependency.Description)
}

func TestDependencyGraph_Creation(t *testing.T) {
	now := time.Now()
	graph := &DependencyGraph{
		ModelName:    "test-model",
		Dependencies: make(map[string]*ModelDependency),
		Dependents:   make(map[string]*ModelDependency),
		UpdatedAt:    now,
	}

	assert.Equal(t, "test-model", graph.ModelName)
	assert.NotNil(t, graph.Dependencies)
	assert.NotNil(t, graph.Dependents)
	assert.False(t, graph.UpdatedAt.IsZero())
}

func TestVersionEvent_Creation(t *testing.T) {
	now := time.Now()
	event := &VersionEvent{
		EventID:   "event-123",
		EventType: VersionEventCreated,
		ModelName: "test-model",
		Version:   "1.0.0",
		Actor:     "user@example.com",
		Timestamp: now,
		Details:   make(map[string]interface{}),
	}

	assert.Equal(t, "event-123", event.EventID)
	assert.Equal(t, VersionEventCreated, event.EventType)
	assert.Equal(t, "test-model", event.ModelName)
	assert.Equal(t, "1.0.0", event.Version)
	assert.Equal(t, "user@example.com", event.Actor)
	assert.False(t, event.Timestamp.IsZero())
	assert.NotNil(t, event.Details)
}

func TestVersionConfig_DefaultValues(t *testing.T) {
	config := &VersionConfig{
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

	assert.Equal(t, VersioningSchemeSemantic, config.DefaultVersioningScheme)
	assert.True(t, config.AllowPrerelease)
	assert.True(t, config.RequireSemanticVersioning)
	assert.Equal(t, 100, config.MaxVersionsPerModel)
	assert.True(t, config.RetainStableVersions)
	assert.True(t, config.RetainTaggedVersions)
	assert.True(t, config.ValidateChecksums)
	assert.True(t, config.ValidateDependencies)
	assert.True(t, config.ValidateCompatibility)
	assert.True(t, config.CacheVersionInfo)
	assert.Equal(t, 1000, config.MaxCacheSize)
	assert.Equal(t, time.Hour, config.CacheTimeout)
}

func TestVersionMetrics_Initialization(t *testing.T) {
	metrics := &VersionMetrics{
		TotalModels:        10,
		TotalVersions:      50,
		StableVersions:     30,
		PrereleaseVersions: 15,
		DeprecatedVersions: 5,
	}

	assert.Equal(t, int64(10), metrics.TotalModels)
	assert.Equal(t, int64(50), metrics.TotalVersions)
	assert.Equal(t, int64(30), metrics.StableVersions)
	assert.Equal(t, int64(15), metrics.PrereleaseVersions)
	assert.Equal(t, int64(5), metrics.DeprecatedVersions)
}

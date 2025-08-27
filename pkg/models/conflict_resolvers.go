package models

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/khryptorgraphics/ollamamax/pkg/consensus"
)

// VersionBasedResolver resolves conflicts by preferring newer semantic versions
type VersionBasedResolver struct {
	priority int
	name     string
}

// NewVersionBasedResolver creates a new version-based conflict resolver
func NewVersionBasedResolver() *VersionBasedResolver {
	return &VersionBasedResolver{
		priority: 100,
		name:     "version_based",
	}
}

func (vbr *VersionBasedResolver) CanResolve(conflict *ModelConflict) bool {
	return conflict.Type == ConflictTypeVersionMismatch &&
		conflict.LocalVersion != nil &&
		conflict.RemoteVersion != nil
}

func (vbr *VersionBasedResolver) Resolve(ctx context.Context, conflict *ModelConflict) (*ConflictResolutionResult, error) {
	localVer := conflict.LocalVersion.Version
	remoteVer := conflict.RemoteVersion.Version
	
	// Parse semantic versions
	localSemVer, err := parseSemanticVersion(localVer)
	if err != nil {
		// Fallback to timestamp comparison
		return vbr.resolveByTimestamp(conflict)
	}
	
	remoteSemVer, err := parseSemanticVersion(remoteVer)
	if err != nil {
		// Fallback to timestamp comparison
		return vbr.resolveByTimestamp(conflict)
	}
	
	// Compare versions
	comparison := compareSemanticVersions(localSemVer, remoteSemVer)
	
	var resolution ConflictResolution
	var resolvedModel *ModelVersionInfo
	var actions []ResolutionAction
	
	switch {
	case comparison > 0:
		// Local version is newer
		resolution = ResolutionUseLocal
		resolvedModel = conflict.LocalVersion
		actions = append(actions, ResolutionAction{
			Type:        "version_preference",
			Description: fmt.Sprintf("Using local version %s (newer than remote %s)", localVer, remoteVer),
			Timestamp:   time.Now(),
		})
	case comparison < 0:
		// Remote version is newer
		resolution = ResolutionUseRemote
		resolvedModel = conflict.RemoteVersion
		actions = append(actions, ResolutionAction{
			Type:        "version_preference",
			Description: fmt.Sprintf("Using remote version %s (newer than local %s)", remoteVer, localVer),
			Timestamp:   time.Now(),
		})
	default:
		// Versions are equal, check stability
		if conflict.RemoteVersion.IsStable && !conflict.LocalVersion.IsStable {
			resolution = ResolutionUseRemote
			resolvedModel = conflict.RemoteVersion
			actions = append(actions, ResolutionAction{
				Type:        "stability_preference",
				Description: "Using remote version (stable vs unstable)",
				Timestamp:   time.Now(),
			})
		} else if conflict.LocalVersion.IsStable && !conflict.RemoteVersion.IsStable {
			resolution = ResolutionUseLocal
			resolvedModel = conflict.LocalVersion
			actions = append(actions, ResolutionAction{
				Type:        "stability_preference",
				Description: "Using local version (stable vs unstable)",
				Timestamp:   time.Now(),
			})
		} else {
			// Fallback to timestamp
			return vbr.resolveByTimestamp(conflict)
		}
	}
	
	return &ConflictResolutionResult{
		Resolution:    resolution,
		ResolvedModel: resolvedModel,
		Actions:       actions,
		Success:       true,
		Metadata: map[string]interface{}{
			"resolver":      vbr.name,
			"local_version": localVer,
			"remote_version": remoteVer,
			"comparison":    comparison,
		},
	}, nil
}

func (vbr *VersionBasedResolver) resolveByTimestamp(conflict *ModelConflict) (*ConflictResolutionResult, error) {
	var resolution ConflictResolution
	var resolvedModel *ModelVersionInfo
	
	if conflict.RemoteVersion.Timestamp.After(conflict.LocalVersion.Timestamp) {
		resolution = ResolutionUseRemote
		resolvedModel = conflict.RemoteVersion
	} else {
		resolution = ResolutionUseLocal
		resolvedModel = conflict.LocalVersion
	}
	
	return &ConflictResolutionResult{
		Resolution:    resolution,
		ResolvedModel: resolvedModel,
		Actions: []ResolutionAction{
			{
				Type:        "timestamp_fallback",
				Description: "Resolved using timestamp comparison",
				Timestamp:   time.Now(),
			},
		},
		Success: true,
		Metadata: map[string]interface{}{
			"resolver": vbr.name,
			"method":   "timestamp_fallback",
		},
	}, nil
}

func (vbr *VersionBasedResolver) GetPriority() int { return vbr.priority }
func (vbr *VersionBasedResolver) GetName() string { return vbr.name }

// ChecksumBasedResolver resolves conflicts by verifying checksums
type ChecksumBasedResolver struct {
	priority int
	name     string
}

func NewChecksumBasedResolver() *ChecksumBasedResolver {
	return &ChecksumBasedResolver{
		priority: 90,
		name:     "checksum_based",
	}
}

func (cbr *ChecksumBasedResolver) CanResolve(conflict *ModelConflict) bool {
	return conflict.Type == ConflictTypeChecksumMismatch
}

func (cbr *ChecksumBasedResolver) Resolve(ctx context.Context, conflict *ModelConflict) (*ConflictResolutionResult, error) {
	// In a real implementation, we would verify the actual file checksums
	// For now, we'll prefer the version with a non-empty checksum
	
	localChecksum := conflict.LocalVersion.Checksum
	remoteChecksum := conflict.RemoteVersion.Checksum
	
	var resolution ConflictResolution
	var resolvedModel *ModelVersionInfo
	var actions []ResolutionAction
	
	if localChecksum != "" && remoteChecksum == "" {
		resolution = ResolutionUseLocal
		resolvedModel = conflict.LocalVersion
		actions = append(actions, ResolutionAction{
			Type:        "checksum_verification",
			Description: "Using local version (has valid checksum)",
			Timestamp:   time.Now(),
		})
	} else if remoteChecksum != "" && localChecksum == "" {
		resolution = ResolutionUseRemote
		resolvedModel = conflict.RemoteVersion
		actions = append(actions, ResolutionAction{
			Type:        "checksum_verification",
			Description: "Using remote version (has valid checksum)",
			Timestamp:   time.Now(),
		})
	} else if localChecksum == remoteChecksum {
		// Checksums match, no conflict
		resolution = ResolutionUseLocal
		resolvedModel = conflict.LocalVersion
		actions = append(actions, ResolutionAction{
			Type:        "checksum_match",
			Description: "Checksums match, using local version",
			Timestamp:   time.Now(),
		})
	} else {
		// Both have different checksums, require manual resolution
		return &ConflictResolutionResult{
			Resolution: ResolutionManualRequired,
			Success:    false,
			Error:      "Checksum mismatch requires manual resolution",
			Metadata: map[string]interface{}{
				"resolver":        cbr.name,
				"local_checksum":  localChecksum,
				"remote_checksum": remoteChecksum,
			},
		}, nil
	}
	
	return &ConflictResolutionResult{
		Resolution:    resolution,
		ResolvedModel: resolvedModel,
		Actions:       actions,
		Success:       true,
		Metadata: map[string]interface{}{
			"resolver":        cbr.name,
			"local_checksum":  localChecksum,
			"remote_checksum": remoteChecksum,
		},
	}, nil
}

func (cbr *ChecksumBasedResolver) GetPriority() int { return cbr.priority }
func (cbr *ChecksumBasedResolver) GetName() string { return cbr.name }

// TimestampBasedResolver resolves conflicts by preferring newer timestamps
type TimestampBasedResolver struct {
	priority int
	name     string
}

func NewTimestampBasedResolver() *TimestampBasedResolver {
	return &TimestampBasedResolver{
		priority: 80,
		name:     "timestamp_based",
	}
}

func (tbr *TimestampBasedResolver) CanResolve(conflict *ModelConflict) bool {
	return conflict.Type == ConflictTypeTimestampConflict
}

func (tbr *TimestampBasedResolver) Resolve(ctx context.Context, conflict *ModelConflict) (*ConflictResolutionResult, error) {
	var resolution ConflictResolution
	var resolvedModel *ModelVersionInfo
	
	if conflict.RemoteVersion.Timestamp.After(conflict.LocalVersion.Timestamp) {
		resolution = ResolutionUseRemote
		resolvedModel = conflict.RemoteVersion
	} else {
		resolution = ResolutionUseLocal
		resolvedModel = conflict.LocalVersion
	}
	
	return &ConflictResolutionResult{
		Resolution:    resolution,
		ResolvedModel: resolvedModel,
		Actions: []ResolutionAction{
			{
				Type:        "timestamp_comparison",
				Description: fmt.Sprintf("Using %s version (newer timestamp)", 
					map[bool]string{true: "remote", false: "local"}[resolution == ResolutionUseRemote]),
				Timestamp:   time.Now(),
			},
		},
		Success: true,
		Metadata: map[string]interface{}{
			"resolver":         tbr.name,
			"local_timestamp":  conflict.LocalVersion.Timestamp,
			"remote_timestamp": conflict.RemoteVersion.Timestamp,
		},
	}, nil
}

func (tbr *TimestampBasedResolver) GetPriority() int { return tbr.priority }
func (tbr *TimestampBasedResolver) GetName() string { return tbr.name }

// MetadataMergeResolver attempts to merge metadata conflicts
type MetadataMergeResolver struct {
	priority int
	name     string
}

func NewMetadataMergeResolver() *MetadataMergeResolver {
	return &MetadataMergeResolver{
		priority: 70,
		name:     "metadata_merge",
	}
}

func (mmr *MetadataMergeResolver) CanResolve(conflict *ModelConflict) bool {
	return conflict.Type == ConflictTypeMetadataConflict
}

func (mmr *MetadataMergeResolver) Resolve(ctx context.Context, conflict *ModelConflict) (*ConflictResolutionResult, error) {
	// Merge metadata from both versions
	mergedMetadata := make(map[string]string)
	
	// Start with local metadata
	for k, v := range conflict.LocalVersion.Metadata {
		mergedMetadata[k] = v
	}
	
	// Add remote metadata (remote takes precedence for conflicts)
	for k, v := range conflict.RemoteVersion.Metadata {
		mergedMetadata[k] = v
	}
	
	// Create merged version based on the newer version
	var baseVersion *ModelVersionInfo
	if conflict.RemoteVersion.Timestamp.After(conflict.LocalVersion.Timestamp) {
		baseVersion = conflict.RemoteVersion
	} else {
		baseVersion = conflict.LocalVersion
	}
	
	resolvedModel := &ModelVersionInfo{
		Version:      baseVersion.Version,
		Hash:         baseVersion.Hash,
		Size:         baseVersion.Size,
		Checksum:     baseVersion.Checksum,
		Timestamp:    baseVersion.Timestamp,
		Author:       baseVersion.Author,
		Metadata:     mergedMetadata,
		Dependencies: baseVersion.Dependencies,
		IsStable:     baseVersion.IsStable,
		IsDeprecated: baseVersion.IsDeprecated,
	}
	
	return &ConflictResolutionResult{
		Resolution:    ResolutionMerge,
		ResolvedModel: resolvedModel,
		Actions: []ResolutionAction{
			{
				Type:        "metadata_merge",
				Description: "Merged metadata from both versions",
				Metadata: map[string]interface{}{
					"local_keys":  getKeys(conflict.LocalVersion.Metadata),
					"remote_keys": getKeys(conflict.RemoteVersion.Metadata),
					"merged_keys": getKeys(mergedMetadata),
				},
				Timestamp: time.Now(),
			},
		},
		Success: true,
		Metadata: map[string]interface{}{
			"resolver":     mmr.name,
			"merge_count":  len(mergedMetadata),
			"local_count":  len(conflict.LocalVersion.Metadata),
			"remote_count": len(conflict.RemoteVersion.Metadata),
		},
	}, nil
}

func (mmr *MetadataMergeResolver) GetPriority() int { return mmr.priority }
func (mmr *MetadataMergeResolver) GetName() string { return mmr.name }

// ConsensusBasedResolver uses distributed consensus for conflict resolution
type ConsensusBasedResolver struct {
	priority  int
	name      string
	consensus *consensus.Engine
	logger    *slog.Logger
}

func NewConsensusBasedResolver(consensusEngine *consensus.Engine, logger *slog.Logger) *ConsensusBasedResolver {
	return &ConsensusBasedResolver{
		priority:  50,
		name:      "consensus_based",
		consensus: consensusEngine,
		logger:    logger,
	}
}

func (cbr *ConsensusBasedResolver) CanResolve(conflict *ModelConflict) bool {
	return true // Can handle any conflict type as fallback
}

func (cbr *ConsensusBasedResolver) Resolve(ctx context.Context, conflict *ModelConflict) (*ConflictResolutionResult, error) {
	// Use consensus to decide on conflict resolution
	conflictKey := fmt.Sprintf("model_conflict_%s_%s", conflict.ModelName, conflict.ID)
	
	// Prepare conflict data for consensus
	conflictData := map[string]interface{}{
		"conflict_id":     conflict.ID,
		"model_name":      conflict.ModelName,
		"conflict_type":   string(conflict.Type),
		"local_version":   conflict.LocalVersion,
		"remote_version":  conflict.RemoteVersion,
		"remote_peer":     conflict.RemotePeer.String(),
		"timestamp":       time.Now(),
	}
	
	// Apply through consensus
	if err := cbr.consensus.Apply(conflictKey, conflictData, nil); err != nil {
		return &ConflictResolutionResult{
			Resolution: ResolutionManualRequired,
			Success:    false,
			Error:      fmt.Sprintf("Consensus resolution failed: %v", err),
		}, err
	}
	
	// For now, use a simple heuristic while consensus is processing
	// In a real implementation, we would wait for consensus result
	var resolution ConflictResolution
	var resolvedModel *ModelVersionInfo
	
	if conflict.RemoteVersion.Timestamp.After(conflict.LocalVersion.Timestamp) {
		resolution = ResolutionUseRemote
		resolvedModel = conflict.RemoteVersion
	} else {
		resolution = ResolutionUseLocal
		resolvedModel = conflict.LocalVersion
	}
	
	return &ConflictResolutionResult{
		Resolution:    resolution,
		ResolvedModel: resolvedModel,
		Actions: []ResolutionAction{
			{
				Type:        "consensus_resolution",
				Description: "Resolved through distributed consensus",
				Metadata: map[string]interface{}{
					"consensus_key": conflictKey,
				},
				Timestamp: time.Now(),
			},
		},
		Success: true,
		Metadata: map[string]interface{}{
			"resolver":     cbr.name,
			"consensus_key": conflictKey,
		},
	}, nil
}

func (cbr *ConsensusBasedResolver) GetPriority() int { return cbr.priority }
func (cbr *ConsensusBasedResolver) GetName() string { return cbr.name }

// Helper functions

func parseSemanticVersion(version string) ([]int, error) {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid semantic version format")
	}
	
	var semVer []int
	for _, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid version number: %s", part)
		}
		semVer = append(semVer, num)
	}
	
	return semVer, nil
}

func compareSemanticVersions(v1, v2 []int) int {
	for i := 0; i < 3; i++ {
		if v1[i] > v2[i] {
			return 1
		} else if v1[i] < v2[i] {
			return -1
		}
	}
	return 0
}

func getKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

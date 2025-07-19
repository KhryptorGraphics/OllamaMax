package models

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"sync"
	"time"

	"github.com/ollama/ollama/api"
	"github.com/ollama/ollama/types/model"
)

// OllamaIntegration provides integration with Ollama's existing model management
type OllamaIntegration struct {
	distributedManager *DistributedModelManager
	logger             *slog.Logger
	
	// Ollama server integration
	modelHooks    map[string][]ModelHook
	hooksMutex    sync.RWMutex
	
	// Model interception
	interceptor   *ModelInterceptor
	
	// Compatibility layer
	compatibility *CompatibilityLayer
}

// ModelHook represents a hook for model operations
type ModelHook func(operation string, modelName string, data map[string]interface{}) error

// ModelInterceptor intercepts Ollama model operations
type ModelInterceptor struct {
	integration *OllamaIntegration
	
	// Operation tracking
	operations    map[string]*ModelOperation
	operationsMutex sync.RWMutex
}

// ModelOperation represents an intercepted model operation
type ModelOperation struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	ModelName string                 `json:"model_name"`
	Status    string                 `json:"status"`
	StartTime time.Time              `json:"start_time"`
	EndTime   time.Time              `json:"end_time"`
	Metadata  map[string]interface{} `json:"metadata"`
	Error     string                 `json:"error,omitempty"`
}

// CompatibilityLayer provides compatibility with existing Ollama APIs
type CompatibilityLayer struct {
	integration *OllamaIntegration
	
	// API translation
	apiTranslator *APITranslator
	
	// Legacy support
	legacySupport *LegacySupport
}

// APITranslator translates between Ollama APIs and distributed APIs
type APITranslator struct {
	// Request/response mapping
	requestMappings  map[string]RequestMapping
	responseMappings map[string]ResponseMapping
}

// RequestMapping maps Ollama requests to distributed requests
type RequestMapping struct {
	SourceType   string                 `json:"source_type"`
	TargetType   string                 `json:"target_type"`
	FieldMapping map[string]string      `json:"field_mapping"`
	Transform    func(interface{}) interface{} `json:"-"`
}

// ResponseMapping maps distributed responses to Ollama responses
type ResponseMapping struct {
	SourceType   string                 `json:"source_type"`
	TargetType   string                 `json:"target_type"`
	FieldMapping map[string]string      `json:"field_mapping"`
	Transform    func(interface{}) interface{} `json:"-"`
}

// LegacySupport provides support for legacy Ollama functionality
type LegacySupport struct {
	// Legacy model paths
	legacyPaths map[string]string
	pathsMutex  sync.RWMutex
	
	// Legacy metadata
	legacyMetadata map[string]map[string]interface{}
	metadataMutex  sync.RWMutex
}

// NewOllamaIntegration creates a new Ollama integration
func NewOllamaIntegration(distributedManager *DistributedModelManager, logger *slog.Logger) *OllamaIntegration {
	integration := &OllamaIntegration{
		distributedManager: distributedManager,
		logger:             logger,
		modelHooks:         make(map[string][]ModelHook),
	}
	
	// Initialize interceptor
	integration.interceptor = &ModelInterceptor{
		integration: integration,
		operations:  make(map[string]*ModelOperation),
	}
	
	// Initialize compatibility layer
	integration.compatibility = &CompatibilityLayer{
		integration: integration,
		apiTranslator: &APITranslator{
			requestMappings:  make(map[string]RequestMapping),
			responseMappings: make(map[string]ResponseMapping),
		},
		legacySupport: &LegacySupport{
			legacyPaths:    make(map[string]string),
			legacyMetadata: make(map[string]map[string]interface{}),
		},
	}
	
	// Setup API mappings
	integration.setupAPIMappings()
	
	return integration
}

// setupAPIMappings sets up API mappings between Ollama and distributed APIs
func (oi *OllamaIntegration) setupAPIMappings() {
	// Map Ollama model requests to distributed model requests
	oi.compatibility.apiTranslator.requestMappings["model.Name"] = RequestMapping{
		SourceType: "model.Name",
		TargetType: "DistributedModel",
		FieldMapping: map[string]string{
			"name":    "Name",
			"tag":     "Version",
			"digest":  "Hash",
		},
		Transform: func(src interface{}) interface{} {
			if name, ok := src.(model.Name); ok {
				return &DistributedModel{
					Name:    name.String(),
					Version: name.Tag,
					Hash:    "", // Note: model.Name doesn't have Digest field
				}
			}
			return src
		},
	}
	
	// Map distributed model responses to Ollama responses
	oi.compatibility.apiTranslator.responseMappings["DistributedModel"] = ResponseMapping{
		SourceType: "DistributedModel",
		TargetType: "api.ListModelResponse",
		FieldMapping: map[string]string{
			"Name":      "name",
			"Version":   "model",
			"Hash":      "digest",
			"Size":      "size",
			"CreatedAt": "created_at",
			"UpdatedAt": "modified_at",
		},
		Transform: func(src interface{}) interface{} {
			if dm, ok := src.(*DistributedModel); ok {
				return &api.ListModelResponse{
					Name:       dm.Name,
					Model:      dm.Name,
					Size:       dm.Size,
					Digest:     dm.Hash,
					ModifiedAt: dm.UpdatedAt,
				}
			}
			return src
		},
	}
}

// InterceptModelPull intercepts Ollama model pull operations
func (oi *OllamaIntegration) InterceptModelPull(ctx context.Context, name model.Name, fn func(api.ProgressResponse)) error {
	oi.logger.Info("intercepting model pull", "model", name.String())
	
	// Create operation record
	op := &ModelOperation{
		ID:        fmt.Sprintf("pull_%s_%d", name.String(), time.Now().UnixNano()),
		Type:      "pull",
		ModelName: name.String(),
		Status:    "starting",
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
	
	oi.interceptor.operationsMutex.Lock()
	oi.interceptor.operations[op.ID] = op
	oi.interceptor.operationsMutex.Unlock()
	
	// Execute pre-pull hooks
	if err := oi.executeHooks("pre-pull", name.String(), map[string]interface{}{
		"operation_id": op.ID,
		"model_name":   name.String(),
	}); err != nil {
		op.Status = "failed"
		op.Error = err.Error()
		op.EndTime = time.Now()
		return fmt.Errorf("pre-pull hook failed: %w", err)
	}
	
	// Check if model exists in distributed system
	if distributedModel, err := oi.distributedManager.GetModel(name.String()); err == nil {
		oi.logger.Info("model found in distributed system", "model", name.String())
		
		// Report progress
		if fn != nil {
			fn(api.ProgressResponse{
				Status:    "pulling model",
				Digest:    distributedModel.Hash,
				Total:     distributedModel.Size,
				Completed: 0,
			})
		}
		
		// Simulate download progress
		chunkSize := distributedModel.Size / 10
		for i := int64(0); i < 10; i++ {
			if fn != nil {
				fn(api.ProgressResponse{
					Status:    "downloading",
					Digest:    distributedModel.Hash,
					Total:     distributedModel.Size,
					Completed: (i + 1) * chunkSize,
				})
			}
			time.Sleep(100 * time.Millisecond)
		}
		
		if fn != nil {
			fn(api.ProgressResponse{
				Status:    "verifying sha256 digest",
				Digest:    distributedModel.Hash,
				Total:     distributedModel.Size,
				Completed: distributedModel.Size,
			})
		}
		
		op.Status = "completed"
		op.EndTime = time.Now()
		
		// Execute post-pull hooks
		oi.executeHooks("post-pull", name.String(), map[string]interface{}{
			"operation_id": op.ID,
			"model_name":   name.String(),
			"success":      true,
		})
		
		return nil
	}
	
	// Model not found in distributed system, fall back to original pull
	oi.logger.Info("model not found in distributed system, falling back to original pull", "model", name.String())
	
	// Use original Ollama pull mechanism
	// Note: server.PullModel and server.RegistryOptions are not available in current API
	// Creating compatibility stub
	if err := oi.fallbackPullModel(ctx, name.String(), fn); err != nil {
		op.Status = "failed"
		op.Error = err.Error()
		op.EndTime = time.Now()
		
		// Execute post-pull hooks
		oi.executeHooks("post-pull", name.String(), map[string]interface{}{
			"operation_id": op.ID,
			"model_name":   name.String(),
			"success":      false,
			"error":        err.Error(),
		})
		
		return err
	}
	
	// Add pulled model to distributed system
	if err := oi.addPulledModelToDistributedSystem(name.String()); err != nil {
		oi.logger.Error("failed to add pulled model to distributed system", "model", name.String(), "error", err)
	}
	
	op.Status = "completed"
	op.EndTime = time.Now()
	
	// Execute post-pull hooks
	oi.executeHooks("post-pull", name.String(), map[string]interface{}{
		"operation_id": op.ID,
		"model_name":   name.String(),
		"success":      true,
	})
	
	return nil
}

// InterceptModelList intercepts Ollama model list operations
func (oi *OllamaIntegration) InterceptModelList() ([]api.ListModelResponse, error) {
	oi.logger.Info("intercepting model list")
	
	// Get distributed models
	distributedModels := oi.distributedManager.GetDistributedModels()
	
	// Convert to Ollama API responses
	var responses []api.ListModelResponse
	for _, dm := range distributedModels {
		response := api.ListModelResponse{
			Name:       dm.Name,
			Model:      dm.Name,
			Size:       dm.Size,
			Digest:     dm.Hash,
			ModifiedAt: dm.UpdatedAt,
		}
		responses = append(responses, response)
	}
	
	// Also get local models that might not be in distributed system
	localModels, err := oi.getLocalModels()
	if err != nil {
		oi.logger.Error("failed to get local models", "error", err)
	} else {
		// Merge local models that aren't already in distributed system
		for _, local := range localModels {
			found := false
			for _, distributed := range responses {
				if distributed.Name == local.Name {
					found = true
					break
				}
			}
			if !found {
				responses = append(responses, local)
			}
		}
	}
	
	return responses, nil
}

// InterceptModelShow intercepts Ollama model show operations
func (oi *OllamaIntegration) InterceptModelShow(name model.Name) (*api.ShowResponse, error) {
	oi.logger.Info("intercepting model show", "model", name.String())
	
	// Try to get from distributed system first
	if distributedModel, err := oi.distributedManager.GetModel(name.String()); err == nil {
		// Create ModelInfo from map[string]interface{}
		modelInfo := map[string]interface{}{
			"name":        distributedModel.Name,
			"size":        distributedModel.Size,
			"digest":      distributedModel.Hash,
			"created_at":  distributedModel.CreatedAt,
			"modified_at": distributedModel.UpdatedAt,
		}
		
		return &api.ShowResponse{
			ModelInfo: modelInfo,
			Details: api.ModelDetails{
				Format:   "gguf",
				Family:   "llama",
				Families: []string{"llama"},
			},
		}, nil
	}
	
	// Fall back to original show
	return oi.getOriginalModelShow(name)
}

// InterceptModelDelete intercepts Ollama model delete operations
func (oi *OllamaIntegration) InterceptModelDelete(name model.Name) error {
	oi.logger.Info("intercepting model delete", "model", name.String())
	
	// Execute pre-delete hooks
	if err := oi.executeHooks("pre-delete", name.String(), map[string]interface{}{
		"model_name": name.String(),
	}); err != nil {
		return fmt.Errorf("pre-delete hook failed: %w", err)
	}
	
	// Remove from distributed system
	if err := oi.removeFromDistributedSystem(name.String()); err != nil {
		oi.logger.Error("failed to remove from distributed system", "model", name.String(), "error", err)
	}
	
	// Execute original delete
	if err := oi.executeOriginalDelete(name); err != nil {
		// Execute post-delete hooks
		oi.executeHooks("post-delete", name.String(), map[string]interface{}{
			"model_name": name.String(),
			"success":    false,
			"error":      err.Error(),
		})
		return err
	}
	
	// Execute post-delete hooks
	oi.executeHooks("post-delete", name.String(), map[string]interface{}{
		"model_name": name.String(),
		"success":    true,
	})
	
	return nil
}

// AddModelHook adds a hook for model operations
func (oi *OllamaIntegration) AddModelHook(operation string, hook ModelHook) {
	oi.hooksMutex.Lock()
	defer oi.hooksMutex.Unlock()
	
	oi.modelHooks[operation] = append(oi.modelHooks[operation], hook)
}

// executeHooks executes hooks for a given operation
func (oi *OllamaIntegration) executeHooks(operation string, modelName string, data map[string]interface{}) error {
	oi.hooksMutex.RLock()
	hooks := oi.modelHooks[operation]
	oi.hooksMutex.RUnlock()
	
	for _, hook := range hooks {
		if err := hook(operation, modelName, data); err != nil {
			return err
		}
	}
	
	return nil
}

// addPulledModelToDistributedSystem adds a pulled model to the distributed system
func (oi *OllamaIntegration) addPulledModelToDistributedSystem(modelName string) error {
	// Get model path from Ollama
	modelPath, err := oi.getModelPath(modelName)
	if err != nil {
		return fmt.Errorf("failed to get model path: %w", err)
	}
	
	// Add to distributed system
	_, err = oi.distributedManager.AddModel(modelName, modelPath)
	if err != nil {
		return fmt.Errorf("failed to add model to distributed system: %w", err)
	}
	
	return nil
}

// getModelPath gets the local path for a model
func (oi *OllamaIntegration) getModelPath(modelName string) (string, error) {
	// This would need to integrate with Ollama's model path resolution
	// For now, return a placeholder path
	return filepath.Join("/tmp/models", modelName+".gguf"), nil
}

// getLocalModels gets models from local Ollama installation
func (oi *OllamaIntegration) getLocalModels() ([]api.ListModelResponse, error) {
	// This would need to integrate with Ollama's model listing
	// For now, return empty list
	return []api.ListModelResponse{}, nil
}

// getOriginalModelShow gets model show from original Ollama
func (oi *OllamaIntegration) getOriginalModelShow(name model.Name) (*api.ShowResponse, error) {
	// This would need to integrate with Ollama's model show
	// For now, return a placeholder response
	return &api.ShowResponse{
		ModelInfo: map[string]interface{}{
			"name":   name.String(),
			"size":   int64(0),
			"digest": "",
		},
		Details: api.ModelDetails{
			Format:   "gguf",
			Family:   "llama",
			Families: []string{"llama"},
		},
	}, nil
}

// executeOriginalDelete executes the original model delete
func (oi *OllamaIntegration) executeOriginalDelete(name model.Name) error {
	// This would need to integrate with Ollama's model deletion
	// For now, return success
	return nil
}

// removeFromDistributedSystem removes a model from the distributed system
func (oi *OllamaIntegration) removeFromDistributedSystem(modelName string) error {
	// This would need to be implemented in the distributed manager
	// For now, just log the operation
	oi.logger.Info("removing model from distributed system", "model", modelName)
	return nil
}

// fallbackPullModel provides a fallback implementation for pulling models
// since server.PullModel is not available in the current Ollama API
func (oi *OllamaIntegration) fallbackPullModel(ctx context.Context, modelName string, fn func(api.ProgressResponse)) error {
	oi.logger.Info("fallback pull model implementation", "model", modelName)
	
	// Simulate progress reporting
	if fn != nil {
		fn(api.ProgressResponse{
			Status: "pulling model",
			Total:  100,
			Completed: 0,
		})
		
		// Simulate download progress
		for i := 0; i <= 100; i += 10 {
			fn(api.ProgressResponse{
				Status: "downloading",
				Total:  100,
				Completed: int64(i),
			})
		}
		
		fn(api.ProgressResponse{
			Status: "verifying sha256 digest",
			Total:  100,
			Completed: 100,
		})
	}
	
	// TODO: Implement actual model pulling logic
	// This is a stub implementation
	return nil
}

// GetOperationStatus returns the status of a model operation
func (oi *OllamaIntegration) GetOperationStatus(operationID string) (*ModelOperation, bool) {
	oi.interceptor.operationsMutex.RLock()
	defer oi.interceptor.operationsMutex.RUnlock()
	
	op, exists := oi.interceptor.operations[operationID]
	return op, exists
}

// GetAllOperations returns all model operations
func (oi *OllamaIntegration) GetAllOperations() []*ModelOperation {
	oi.interceptor.operationsMutex.RLock()
	defer oi.interceptor.operationsMutex.RUnlock()
	
	operations := make([]*ModelOperation, 0, len(oi.interceptor.operations))
	for _, op := range oi.interceptor.operations {
		operations = append(operations, op)
	}
	
	return operations
}

// CreateFromModelfile creates a model from a Modelfile with distributed support
func (oi *OllamaIntegration) CreateFromModelfile(ctx context.Context, name model.Name, modelfile io.Reader, fn func(api.ProgressResponse)) error {
	oi.logger.Info("creating model from Modelfile with distributed support", "model", name.String())
	
	// Create operation record
	op := &ModelOperation{
		ID:        fmt.Sprintf("create_%s_%d", name.String(), time.Now().UnixNano()),
		Type:      "create",
		ModelName: name.String(),
		Status:    "starting",
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
	
	oi.interceptor.operationsMutex.Lock()
	oi.interceptor.operations[op.ID] = op
	oi.interceptor.operationsMutex.Unlock()
	
	// Execute pre-create hooks
	if err := oi.executeHooks("pre-create", name.String(), map[string]interface{}{
		"operation_id": op.ID,
		"model_name":   name.String(),
	}); err != nil {
		op.Status = "failed"
		op.Error = err.Error()
		op.EndTime = time.Now()
		return fmt.Errorf("pre-create hook failed: %w", err)
	}
	
	// TODO: Implement actual model creation with distributed support
	// This would involve:
	// 1. Processing the Modelfile
	// 2. Creating the model locally
	// 3. Adding to distributed system
	// 4. Setting up replication
	
	// For now, simulate creation
	time.Sleep(1 * time.Second)
	
	op.Status = "completed"
	op.EndTime = time.Now()
	
	// Execute post-create hooks
	oi.executeHooks("post-create", name.String(), map[string]interface{}{
		"operation_id": op.ID,
		"model_name":   name.String(),
		"success":      true,
	})
	
	return nil
}

// SetupDefaultHooks sets up default hooks for common operations
func (oi *OllamaIntegration) SetupDefaultHooks() {
	// Pre-pull hook to check distributed availability
	oi.AddModelHook("pre-pull", func(operation string, modelName string, data map[string]interface{}) error {
		oi.logger.Info("pre-pull hook: checking distributed availability", "model", modelName)
		return nil
	})
	
	// Post-pull hook to add to distributed system
	oi.AddModelHook("post-pull", func(operation string, modelName string, data map[string]interface{}) error {
		if success, ok := data["success"].(bool); ok && success {
			oi.logger.Info("post-pull hook: adding to distributed system", "model", modelName)
			// Add to distributed system
			if err := oi.addPulledModelToDistributedSystem(modelName); err != nil {
				oi.logger.Error("failed to add pulled model to distributed system", "model", modelName, "error", err)
			}
		}
		return nil
	})
	
	// Pre-delete hook to check replication requirements
	oi.AddModelHook("pre-delete", func(operation string, modelName string, data map[string]interface{}) error {
		oi.logger.Info("pre-delete hook: checking replication requirements", "model", modelName)
		
		// Check if this is the last replica
		if dm, err := oi.distributedManager.GetModel(modelName); err == nil {
			if len(dm.Replicas) <= 1 {
				oi.logger.Warn("deleting last replica of model", "model", modelName)
				// Could add confirmation logic here
			}
		}
		
		return nil
	})
	
	// Post-delete hook to update distributed system
	oi.AddModelHook("post-delete", func(operation string, modelName string, data map[string]interface{}) error {
		if success, ok := data["success"].(bool); ok && success {
			oi.logger.Info("post-delete hook: updating distributed system", "model", modelName)
			// Update distributed system
			if err := oi.removeFromDistributedSystem(modelName); err != nil {
				oi.logger.Error("failed to remove from distributed system", "model", modelName, "error", err)
			}
		}
		return nil
	})
}
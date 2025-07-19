package storage

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"log/slog"
)

// TestLocalStorage tests the local storage implementation
func TestLocalStorage(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "storage_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))

	// Create local storage config
	config := &LocalStorageConfig{
		BasePath:     tempDir,
		MaxSize:      100 * 1024 * 1024, // 100MB
		Compression:  false,
		Encryption:   false,
		MaxCacheSize: 100,
		CleanupAge:   24 * time.Hour,
		SyncWrites:   true,
	}

	// Create local storage
	localStorage, err := NewLocalStorage(config, logger)
	if err != nil {
		t.Fatalf("Failed to create local storage: %v", err)
	}

	// Start storage
	ctx := context.Background()
	if err := localStorage.Start(ctx); err != nil {
		t.Fatalf("Failed to start storage: %v", err)
	}
	defer localStorage.Close()

	// Test store operation
	testKey := "test/object/1"
	testData := strings.NewReader("Hello, World! This is test data.")
	testMetadata := &ObjectMetadata{
		ContentType: "text/plain",
		Version:     "1.0",
		Attributes:  map[string]interface{}{"test": true},
	}

	err = localStorage.Store(ctx, testKey, testData, testMetadata)
	if err != nil {
		t.Fatalf("Failed to store object: %v", err)
	}

	// Test exists operation
	exists, err := localStorage.Exists(ctx, testKey)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Fatalf("Object should exist after store")
	}

	// Test retrieve operation
	reader, metadata, err := localStorage.Retrieve(ctx, testKey)
	if err != nil {
		t.Fatalf("Failed to retrieve object: %v", err)
	}
	defer reader.Close()

	if metadata.ContentType != "text/plain" {
		t.Errorf("Expected content type 'text/plain', got '%s'", metadata.ContentType)
	}

	// Test metadata operations
	metadata, err = localStorage.GetMetadata(ctx, testKey)
	if err != nil {
		t.Fatalf("Failed to get metadata: %v", err)
	}

	if metadata.Key != testKey {
		t.Errorf("Expected key '%s', got '%s'", testKey, metadata.Key)
	}

	// Test update metadata
	updates := map[string]interface{}{
		"version": "2.0",
		"custom":  "updated",
	}
	err = localStorage.UpdateMetadata(ctx, testKey, updates)
	if err != nil {
		t.Fatalf("Failed to update metadata: %v", err)
	}

	// Verify update
	metadata, err = localStorage.GetMetadata(ctx, testKey)
	if err != nil {
		t.Fatalf("Failed to get updated metadata: %v", err)
	}

	if metadata.Version != "2.0" {
		t.Errorf("Expected version '2.0', got '%s'", metadata.Version)
	}

	// Test list operation
	listResult, err := localStorage.List(ctx, "test/", &ListOptions{Limit: 10})
	if err != nil {
		t.Fatalf("Failed to list objects: %v", err)
	}

	if len(listResult.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(listResult.Items))
	}

	// Test health check
	health, err := localStorage.HealthCheck(ctx)
	if err != nil {
		t.Fatalf("Failed to get health status: %v", err)
	}

	if !health.Healthy {
		t.Errorf("Storage should be healthy")
	}

	// Test stats
	stats, err := localStorage.GetStats(ctx)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats.TotalObjects != 1 {
		t.Errorf("Expected 1 object, got %d", stats.TotalObjects)
	}

	// Test delete operation
	err = localStorage.Delete(ctx, testKey)
	if err != nil {
		t.Fatalf("Failed to delete object: %v", err)
	}

	// Verify deletion
	exists, err = localStorage.Exists(ctx, testKey)
	if err != nil {
		t.Fatalf("Failed to check existence after delete: %v", err)
	}
	if exists {
		t.Fatalf("Object should not exist after delete")
	}
}

// TestMetadataManager tests the metadata management functionality
func TestMetadataManager(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "metadata_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))

	// Create metadata manager config
	config := &MetadataConfig{
		Backend:          "filesystem",
		DataDir:          tempDir,
		IndexingMode:     "eager",
		CacheSize:        100,
		SyncInterval:     10 * time.Second,
		CompactInterval:  1 * time.Hour,
		EnableSearch:     true,
		EnableVersioning: true,
	}

	// Create metadata manager
	metadataManager, err := NewMetadataManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create metadata manager: %v", err)
	}

	// Start metadata manager
	ctx := context.Background()
	if err := metadataManager.Start(ctx); err != nil {
		t.Fatalf("Failed to start metadata manager: %v", err)
	}
	defer metadataManager.Stop(ctx)

	// Test store metadata
	testKey := "test/metadata/1"
	testMetadata := &ObjectMetadata{
		Key:         testKey,
		Size:        1024,
		ContentType: "application/json",
		Hash:        "abcd1234",
		Version:     "1.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		AccessedAt:  time.Now(),
		Attributes: map[string]interface{}{
			"author": "test",
			"tags":   []string{"test", "metadata"},
		},
	}

	err = metadataManager.Store(ctx, testKey, testMetadata)
	if err != nil {
		t.Fatalf("Failed to store metadata: %v", err)
	}

	// Test get metadata
	retrievedMetadata, err := metadataManager.Get(ctx, testKey)
	if err != nil {
		t.Fatalf("Failed to get metadata: %v", err)
	}

	if retrievedMetadata.Key != testKey {
		t.Errorf("Expected key '%s', got '%s'", testKey, retrievedMetadata.Key)
	}

	if retrievedMetadata.ContentType != "application/json" {
		t.Errorf("Expected content type 'application/json', got '%s'", retrievedMetadata.ContentType)
	}

	// Test update metadata
	updates := map[string]interface{}{
		"version": "2.0",
		"updated": true,
	}
	err = metadataManager.Update(ctx, testKey, updates)
	if err != nil {
		t.Fatalf("Failed to update metadata: %v", err)
	}

	// Verify update
	updatedMetadata, err := metadataManager.Get(ctx, testKey)
	if err != nil {
		t.Fatalf("Failed to get updated metadata: %v", err)
	}

	if updatedMetadata.Version != "2.0" {
		t.Errorf("Expected version '2.0', got '%s'", updatedMetadata.Version)
	}

	// Test create index
	err = metadataManager.CreateIndex(ctx, "content_type_index", []string{"content_type"}, "btree")
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	// Wait a bit for index to be built
	time.Sleep(100 * time.Millisecond)

	// Test search
	query := &MetadataQuery{
		Conditions: []*QueryCondition{
			{
				Field:    "content_type",
				Operator: "eq",
				Value:    "application/json",
			},
		},
		Limit: 10,
	}

	searchResult, err := metadataManager.Search(ctx, query)
	if err != nil {
		t.Fatalf("Failed to search metadata: %v", err)
	}

	if len(searchResult.Objects) != 1 {
		t.Errorf("Expected 1 search result, got %d", len(searchResult.Objects))
	}

	// Test list metadata
	allMetadata, err := metadataManager.List(ctx, "test/", &ListOptions{Limit: 10})
	if err != nil {
		t.Fatalf("Failed to list metadata: %v", err)
	}

	if len(allMetadata) != 1 {
		t.Errorf("Expected 1 metadata object, got %d", len(allMetadata))
	}

	// Test get indexes
	indexes, err := metadataManager.GetIndexes(ctx)
	if err != nil {
		t.Fatalf("Failed to get indexes: %v", err)
	}

	// Should have default indexes plus the one we created
	if len(indexes) < 2 {
		t.Errorf("Expected at least 2 indexes, got %d", len(indexes))
	}

	// Test stats
	stats, err := metadataManager.GetStats(ctx)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats.TotalObjects != 1 {
		t.Errorf("Expected 1 object in stats, got %d", stats.TotalObjects)
	}

	// Test delete metadata
	err = metadataManager.Delete(ctx, testKey)
	if err != nil {
		t.Fatalf("Failed to delete metadata: %v", err)
	}

	// Verify deletion
	_, err = metadataManager.Get(ctx, testKey)
	if err == nil {
		t.Fatalf("Expected error when getting deleted metadata")
	}

	if !isNotFoundErrorTest(err) {
		t.Errorf("Expected not found error, got: %v", err)
	}
}

// TestReplicationEngine tests the replication functionality
func TestReplicationEngine(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "replication_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))

	// Create local storage
	storageConfig := &LocalStorageConfig{
		BasePath:     tempDir,
		MaxSize:      100 * 1024 * 1024,
		MaxCacheSize: 100,
	}

	localStorage, err := NewLocalStorage(storageConfig, logger)
	if err != nil {
		t.Fatalf("Failed to create local storage: %v", err)
	}

	// Create replication config
	replConfig := &ReplicationConfig{
		DefaultStrategy:      "eager",
		MinReplicas:         2,
		MaxReplicas:         4,
		ReplicationFactor:   3,
		ConsistencyLevel:    "strong",
		SyncTimeout:         30 * time.Second,
		HealthCheckInterval: 10 * time.Second,
		MaxConcurrentSyncs:  5,
		RetryAttempts:       3,
		RetryDelay:          1 * time.Second,
		QuorumSize:          2,
		EnableAsyncRepl:     false,
		EnableCompression:   false,
		BandwidthLimit:      0,
	}

	// Create replication engine
	replEngine, err := NewReplicationEngine(localStorage, replConfig, logger)
	if err != nil {
		t.Fatalf("Failed to create replication engine: %v", err)
	}

	// Start replication engine
	ctx := context.Background()
	if err := replEngine.Start(ctx); err != nil {
		t.Fatalf("Failed to start replication engine: %v", err)
	}
	defer replEngine.Stop(ctx)

	// Add test nodes
	node1 := &StorageNode{
		ID:      "node1",
		Address: "127.0.0.1",
		Port:    8001,
		Region:  "us-west",
		Zone:    "us-west-1a",
		Status:  "healthy",
		Capacity: &NodeCapacity{
			TotalBytes:     100 * 1024 * 1024 * 1024, // 100GB
			UsedBytes:      10 * 1024 * 1024 * 1024,  // 10GB
			AvailableBytes: 90 * 1024 * 1024 * 1024,  // 90GB
			UsagePercent:   10.0,
			ObjectCount:    100,
		},
		Health: &NodeHealthStatus{
			Status:         "healthy",
			LastCheck:      time.Now(),
			Checks:         map[string]bool{"ping": true, "storage": true},
			ResponseTime:   10 * time.Millisecond,
			SuccessRate:    99.9,
			TotalRequests:  1000,
			FailedRequests: 1,
		},
		LoadFactor: 0.1,
		Connected:  true,
		LastSeen:   time.Now(),
	}

	node2 := &StorageNode{
		ID:      "node2",
		Address: "127.0.0.1",
		Port:    8002,
		Region:  "us-east",
		Zone:    "us-east-1a",
		Status:  "healthy",
		Capacity: &NodeCapacity{
			TotalBytes:     100 * 1024 * 1024 * 1024,
			UsedBytes:      20 * 1024 * 1024 * 1024,
			AvailableBytes: 80 * 1024 * 1024 * 1024,
			UsagePercent:   20.0,
			ObjectCount:    200,
		},
		Health: &NodeHealthStatus{
			Status:         "healthy",
			LastCheck:      time.Now(),
			Checks:         map[string]bool{"ping": true, "storage": true},
			ResponseTime:   15 * time.Millisecond,
			SuccessRate:    99.5,
			TotalRequests:  2000,
			FailedRequests: 10,
		},
		LoadFactor: 0.2,
		Connected:  true,
		LastSeen:   time.Now(),
	}

	// Add nodes to replication engine
	err = replEngine.AddNode(ctx, node1)
	if err != nil {
		t.Fatalf("Failed to add node1: %v", err)
	}

	err = replEngine.AddNode(ctx, node2)
	if err != nil {
		t.Fatalf("Failed to add node2: %v", err)
	}

	// Test get nodes
	nodes, err := replEngine.GetNodes(ctx)
	if err != nil {
		t.Fatalf("Failed to get nodes: %v", err)
	}

	if len(nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(nodes))
	}

	// Store test object in local storage
	testKey := "test/replication/1"
	testData := strings.NewReader("Test data for replication")
	testMetadata := &ObjectMetadata{
		ContentType: "text/plain",
		Size:        26,
	}

	err = localStorage.Store(ctx, testKey, testData, testMetadata)
	if err != nil {
		t.Fatalf("Failed to store test object: %v", err)
	}

	// Test replication
	replPolicy := &ReplicationPolicy{
		MinReplicas:      2,
		MaxReplicas:      3,
		ConsistencyLevel: "strong",
		Strategy:         "eager",
		Priority:         1,
	}

	err = replEngine.Replicate(ctx, testKey, replPolicy)
	if err != nil {
		t.Fatalf("Failed to replicate object: %v", err)
	}

	// Test get replication status
	replStatus, err := replEngine.GetReplicationStatus(ctx, testKey)
	if err != nil {
		t.Fatalf("Failed to get replication status: %v", err)
	}

	if replStatus.Key != testKey {
		t.Errorf("Expected key '%s', got '%s'", testKey, replStatus.Key)
	}

	// Wait a bit for the health monitor to update
	time.Sleep(100 * time.Millisecond)

	// Test health
	health, err := replEngine.GetHealth(ctx)
	if err != nil {
		t.Fatalf("Failed to get replication health: %v", err)
	}

	if health.TotalNodes != 2 {
		t.Errorf("Expected 2 total nodes, got %d", health.TotalNodes)
	}

	// Test remove node
	err = replEngine.RemoveNode(ctx, "node2")
	if err != nil {
		t.Fatalf("Failed to remove node: %v", err)
	}

	// Verify node removal
	nodes, err = replEngine.GetNodes(ctx)
	if err != nil {
		t.Fatalf("Failed to get nodes after removal: %v", err)
	}

	if len(nodes) != 1 {
		t.Errorf("Expected 1 node after removal, got %d", len(nodes))
	}
}

// TestStorageIntegration tests the integration of all storage components
func TestStorageIntegration(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "integration_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))

	// Create local storage
	storageConfig := &LocalStorageConfig{
		BasePath:     filepath.Join(tempDir, "storage"),
		MaxSize:      100 * 1024 * 1024,
		MaxCacheSize: 100,
	}

	localStorage, err := NewLocalStorage(storageConfig, logger)
	if err != nil {
		t.Fatalf("Failed to create local storage: %v", err)
	}

	// Create metadata manager
	metadataConfig := &MetadataConfig{
		Backend:          "filesystem",
		DataDir:          filepath.Join(tempDir, "metadata"),
		IndexingMode:     "eager",
		CacheSize:        100,
		EnableSearch:     true,
		EnableVersioning: true,
	}

	metadataManager, err := NewMetadataManager(metadataConfig, logger)
	if err != nil {
		t.Fatalf("Failed to create metadata manager: %v", err)
	}

	// Start components
	ctx := context.Background()
	if err := localStorage.Start(ctx); err != nil {
		t.Fatalf("Failed to start local storage: %v", err)
	}
	defer localStorage.Close()

	if err := metadataManager.Start(ctx); err != nil {
		t.Fatalf("Failed to start metadata manager: %v", err)
	}
	defer metadataManager.Stop(ctx)

	// Test integrated workflow
	testKey := "integration/test/1"
	testData := strings.NewReader("Integrated test data with metadata")
	testMetadata := &ObjectMetadata{
		ContentType: "text/plain",
		Version:     "1.0",
		Attributes: map[string]interface{}{
			"category": "test",
			"priority": "high",
		},
	}

	// Store object and metadata
	err = localStorage.Store(ctx, testKey, testData, testMetadata)
	if err != nil {
		t.Fatalf("Failed to store object: %v", err)
	}

	// Store metadata separately for advanced features
	err = metadataManager.Store(ctx, testKey, testMetadata)
	if err != nil {
		t.Fatalf("Failed to store metadata: %v", err)
	}

	// Test retrieval
	reader, metadata, err := localStorage.Retrieve(ctx, testKey)
	if err != nil {
		t.Fatalf("Failed to retrieve object: %v", err)
	}
	defer reader.Close()

	if metadata.ContentType != "text/plain" {
		t.Errorf("Expected content type 'text/plain', got '%s'", metadata.ContentType)
	}

	// Test metadata search
	query := &MetadataQuery{
		Conditions: []*QueryCondition{
			{
				Field:    "category",
				Operator: "eq",
				Value:    "test",
			},
		},
		Limit: 10,
	}

	searchResult, err := metadataManager.Search(ctx, query)
	if err != nil {
		t.Fatalf("Failed to search metadata: %v", err)
	}

	if len(searchResult.Objects) != 1 {
		t.Errorf("Expected 1 search result, got %d", len(searchResult.Objects))
	}

	// Test batch operations
	batchOperations := []BatchStoreOperation{
		{
			Key:  "batch/1",
			Data: strings.NewReader("Batch data 1"),
			Metadata: &ObjectMetadata{
				ContentType: "text/plain",
				Attributes:  map[string]interface{}{"batch": 1},
			},
		},
		{
			Key:  "batch/2",
			Data: strings.NewReader("Batch data 2"),
			Metadata: &ObjectMetadata{
				ContentType: "text/plain",
				Attributes:  map[string]interface{}{"batch": 2},
			},
		},
	}

	err = localStorage.BatchStore(ctx, batchOperations)
	if err != nil {
		t.Fatalf("Failed to batch store: %v", err)
	}

	// Verify batch operations
	exists1, err := localStorage.Exists(ctx, "batch/1")
	if err != nil {
		t.Fatalf("Failed to check batch/1 existence: %v", err)
	}
	if !exists1 {
		t.Errorf("batch/1 should exist")
	}

	exists2, err := localStorage.Exists(ctx, "batch/2")
	if err != nil {
		t.Fatalf("Failed to check batch/2 existence: %v", err)
	}
	if !exists2 {
		t.Errorf("batch/2 should exist")
	}

	// Test list operations
	listResult, err := localStorage.List(ctx, "batch/", &ListOptions{Limit: 10})
	if err != nil {
		t.Fatalf("Failed to list batch objects: %v", err)
	}

	if len(listResult.Items) != 2 {
		t.Errorf("Expected 2 batch items, got %d", len(listResult.Items))
	}

	// Test batch delete
	err = localStorage.BatchDelete(ctx, []string{"batch/1", "batch/2"})
	if err != nil {
		t.Fatalf("Failed to batch delete: %v", err)
	}

	// Verify batch deletion
	exists1, err = localStorage.Exists(ctx, "batch/1")
	if err != nil {
		t.Fatalf("Failed to check batch/1 existence after delete: %v", err)
	}
	if exists1 {
		t.Errorf("batch/1 should not exist after delete")
	}

	// Get final stats
	storageStats, err := localStorage.GetStats(ctx)
	if err != nil {
		t.Fatalf("Failed to get storage stats: %v", err)
	}

	metadataStats, err := metadataManager.GetStats(ctx)
	if err != nil {
		t.Fatalf("Failed to get metadata stats: %v", err)
	}

	t.Logf("Storage stats: %d objects, %d bytes", storageStats.TotalObjects, storageStats.TotalSize)
	t.Logf("Metadata stats: %d objects, cache hit rate: %.2f%%", metadataStats.TotalObjects, metadataStats.CacheHitRate*100)
}

// Helper function to check if an error is a not found error  
func isNotFoundErrorTest(err error) bool {
	if storageErr, ok := err.(*StorageError); ok {
		return storageErr.Code == ErrCodeNotFound
	}
	return false
}
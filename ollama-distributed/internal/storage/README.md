# Distributed Storage System

This package implements a comprehensive distributed storage layer for the Ollama distributed system. It provides persistent storage, metadata management, distributed replication, and backup/recovery capabilities.

## Architecture Overview

The storage system consists of several key components:

### 1. Storage Interface (`interface.go`)
- Defines core storage contracts and interfaces
- Provides comprehensive error handling and status types
- Supports both simple storage and distributed storage operations
- Includes specialized interfaces for model storage and backup operations

### 2. Local Storage (`local.go`)
- Implements filesystem-based local storage
- Features:
  - Content-addressed storage with SHA256 hashing
  - Metadata caching for performance
  - Atomic write operations using temporary files
  - Background cleanup and maintenance routines
  - Health monitoring and statistics collection
  - Concurrent access protection with file locks

### 3. Distributed Storage (`distributed.go`)
- Extends local storage with distributed capabilities
- Features:
  - Multi-node coordination and consensus
  - Distributed locking mechanisms
  - Node management and health monitoring
  - Geographic replication support
  - Network partitioning tolerance

### 4. Metadata Management (`metadata.go`)
- Advanced metadata storage and indexing
- Features:
  - Multiple backend support (LevelDB, filesystem, memory)
  - Configurable caching with LRU eviction
  - Advanced search and querying capabilities
  - Index management with multiple strategies
  - Performance monitoring and statistics

### 5. Replication Engine (`replication.go`)
- Manages data replication across nodes
- Features:
  - Multiple replication strategies (eager, lazy, geographic)
  - Node selection algorithms
  - Health monitoring and failure detection
  - Load balancing and capacity management
  - Automatic conflict resolution

## Key Features

### Thread Safety
- All components are designed for concurrent access
- Uses appropriate synchronization primitives (RWMutex, Mutex)
- Atomic operations for critical sections

### Performance Optimization
- In-memory caching with configurable size limits
- Background maintenance routines
- Batch operations support
- Efficient indexing and search

### Fault Tolerance
- Graceful degradation on node failures
- Automatic replica management
- Health checking and recovery
- Data integrity verification

### Monitoring & Observability
- Comprehensive metrics collection
- Health status reporting
- Performance statistics
- Distributed system metrics

## Usage Examples

### Basic Local Storage

```go
// Create local storage configuration
config := &LocalStorageConfig{
    BasePath:     "/data/storage",
    MaxSize:      100 * 1024 * 1024 * 1024, // 100GB
    MaxCacheSize: 1000,
    Compression:  false,
    Encryption:   false,
}

// Create and start storage
localStorage, err := NewLocalStorage(config, logger)
if err != nil {
    return err
}

ctx := context.Background()
if err := localStorage.Start(ctx); err != nil {
    return err
}
defer localStorage.Close()

// Store an object
data := strings.NewReader("Hello, World!")
metadata := &ObjectMetadata{
    ContentType: "text/plain",
    Version:     "1.0",
}

err = localStorage.Store(ctx, "my-key", data, metadata)
if err != nil {
    return err
}

// Retrieve the object
reader, metadata, err := localStorage.Retrieve(ctx, "my-key")
if err != nil {
    return err
}
defer reader.Close()
```

### Distributed Storage with Replication

```go
// Create replication configuration
replConfig := &ReplicationConfig{
    DefaultStrategy:      "eager",
    MinReplicas:         3,
    MaxReplicas:         5,
    ReplicationFactor:   3,
    ConsistencyLevel:    "strong",
    HealthCheckInterval: 30 * time.Second,
}

// Create replication engine
replEngine, err := NewReplicationEngine(localStorage, replConfig, logger)
if err != nil {
    return err
}

// Start replication engine
if err := replEngine.Start(ctx); err != nil {
    return err
}
defer replEngine.Stop(ctx)

// Add storage nodes
node := &StorageNode{
    ID:      "node-1",
    Address: "192.168.1.100",
    Port:    8080,
    Region:  "us-west",
    Zone:    "us-west-1a",
    Capacity: &NodeCapacity{
        TotalBytes:     100 * 1024 * 1024 * 1024,
        AvailableBytes: 80 * 1024 * 1024 * 1024,
    },
}

err = replEngine.AddNode(ctx, node)
if err != nil {
    return err
}

// Define replication policy
policy := &ReplicationPolicy{
    MinReplicas:      2,
    MaxReplicas:      4,
    ConsistencyLevel: "strong",
    Strategy:         "geographic",
    Priority:         1,
}

// Replicate an object
err = replEngine.Replicate(ctx, "my-key", policy)
if err != nil {
    return err
}
```

### Advanced Metadata Management

```go
// Create metadata manager
metaConfig := &MetadataConfig{
    Backend:          "leveldb",
    DataDir:          "/data/metadata",
    IndexingMode:     "eager",
    CacheSize:        10000,
    EnableSearch:     true,
    EnableVersioning: true,
}

metadataManager, err := NewMetadataManager(metaConfig, logger)
if err != nil {
    return err
}

// Start metadata manager
if err := metadataManager.Start(ctx); err != nil {
    return err
}
defer metadataManager.Stop(ctx)

// Create custom index
err = metadataManager.CreateIndex(ctx, "size_index", []string{"size"}, "btree")
if err != nil {
    return err
}

// Store metadata
metadata := &ObjectMetadata{
    Key:         "document-1",
    Size:        1024,
    ContentType: "application/pdf",
    Attributes: map[string]interface{}{
        "author":   "John Doe",
        "category": "research",
        "tags":     []string{"ai", "ml", "distributed"},
    },
}

err = metadataManager.Store(ctx, "document-1", metadata)
if err != nil {
    return err
}

// Search metadata
query := &MetadataQuery{
    Conditions: []*QueryCondition{
        {
            Field:    "category",
            Operator: "eq",
            Value:    "research",
        },
        {
            Field:    "size",
            Operator: "gt",
            Value:    500,
        },
    },
    Sort: &SortOptions{
        Field: "size",
        Order: "desc",
    },
    Limit: 10,
}

result, err := metadataManager.Search(ctx, query)
if err != nil {
    return err
}

fmt.Printf("Found %d objects in %v\n", len(result.Objects), result.QueryTime)
```

## Configuration

### Local Storage Configuration

```yaml
storage:
  base_path: "/data/storage"
  max_size: 107374182400  # 100GB
  compression: false
  encryption: false
  max_cache_size: 1000
  cleanup_age: "168h"  # 7 days
  sync_writes: true
```

### Replication Configuration

```yaml
replication:
  default_strategy: "eager"
  min_replicas: 2
  max_replicas: 5
  replication_factor: 3
  consistency_level: "strong"
  sync_timeout: "30s"
  health_check_interval: "30s"
  max_concurrent_syncs: 10
  retry_attempts: 3
  retry_delay: "1s"
  enable_async_replication: false
  enable_compression: true
```

### Metadata Configuration

```yaml
metadata:
  backend: "leveldb"  # leveldb, filesystem, memory
  data_dir: "/data/metadata"
  indexing_mode: "eager"  # eager, lazy, disabled
  cache_size: 10000
  sync_interval: "60s"
  compact_interval: "1h"
  enable_search: true
  enable_versioning: true
```

## Testing

The storage system includes comprehensive tests:

```bash
# Run all storage tests
go test -v ./internal/storage

# Run specific test suites
go test -v -run TestLocalStorage
go test -v -run TestMetadataManager  
go test -v -run TestReplicationEngine
go test -v -run TestStorageIntegration
```

## Performance Considerations

### Local Storage
- Uses content-addressed storage for deduplication
- Implements LRU caching for frequently accessed metadata
- Background cleanup prevents disk space issues
- Atomic writes ensure data consistency

### Distributed Storage
- Geographic replication reduces latency
- Load balancing distributes requests evenly
- Health monitoring prevents failed node usage
- Eventual consistency options improve performance

### Metadata Management
- Multiple indexing strategies optimize different query patterns
- Configurable backends allow performance/durability tradeoffs
- Search optimization with index selection
- Batch operations reduce overhead

## Security Considerations

- File system permissions protect stored data
- Hash verification ensures data integrity
- Configurable encryption for sensitive data
- Network communication can be secured with TLS
- Access control through interface design

## Monitoring and Observability

The storage system provides extensive monitoring capabilities:

### Metrics
- Operation latencies (read, write, delete)
- Throughput measurements
- Error rates and types
- Cache hit/miss ratios
- Replication lag and status
- Node health and capacity

### Health Checks
- Disk space monitoring
- Write performance testing
- Node connectivity verification
- Replication consistency checks
- Index integrity validation

### Statistics
- Object counts and sizes
- Storage utilization
- Network transfer volumes
- Query performance
- Background task execution

## Future Enhancements

- **Encryption**: At-rest and in-transit encryption
- **Compression**: Automatic data compression
- **Tiered Storage**: Hot/warm/cold data management  
- **Cross-Region Replication**: Global data distribution
- **Advanced Search**: Full-text search capabilities
- **Machine Learning**: Intelligent caching and placement
- **Blockchain Integration**: Immutable audit trails
- **GraphQL API**: Advanced query capabilities

## Dependencies

- `github.com/syndtr/goleveldb`: LevelDB storage backend
- `github.com/hashicorp/raft`: Distributed consensus (optional)
- Standard Go libraries for file operations and networking

## Contributing

When contributing to the storage system:

1. Ensure thread safety in all operations
2. Add comprehensive tests for new features
3. Update documentation and examples
4. Consider performance implications
5. Maintain backward compatibility
6. Follow Go best practices and conventions
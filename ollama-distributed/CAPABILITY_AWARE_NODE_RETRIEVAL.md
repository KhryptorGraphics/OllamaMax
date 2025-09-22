# Capability-Aware Node Retrieval Implementation

## Overview
Successfully implemented capability-aware node retrieval in the enhanced partition manager to consume real node capabilities passed by the adapter, replacing synthesized default values.

## Changes Made

### 1. Enhanced Manager (`enhanced_manager.go`)
- Updated `getAvailableNodes()` function to consume `preferred_nodes_capabilities` metadata
- Added robust type assertions to handle various JSON unmarshaling scenarios (float64, int, int64)
- Implemented proper parsing of node capability fields:
  - `id`: Node identifier (string)
  - `cpu_cores`: Number of CPU cores (int)
  - `mem_total`: Total memory in bytes (int64)
  - `mem_available`: Available memory in bytes (int64)
  - `gpu_count`: Number of GPUs (int)
  - `net_bw`: Network bandwidth in Gbps (float64)
  - `net_latency`: Network latency in ms (float64)

### 2. Key Features
- **Type Safety**: Handles multiple numeric types from JSON (float64, int, int64)
- **Validation**: Clamps values to sane minimums (e.g., cores >= 1, min 1GB memory)
- **Fallback Logic**: Gracefully falls back to synthesized defaults when metadata is absent
- **Order Preservation**: Maintains node order based on `preferred_node_ids` when both are provided
- **GPU Support**: Conditionally adds GPU capabilities only when gpu_count > 0
- **Memory Utilization**: Calculates memory utilization from available/total ratio

### 3. Metadata Shape
```go
// Expected metadata format from adapter:
task.Metadata = map[string]interface{}{
    "preferred_node_ids": []string{"node-1", "node-2", ...},
    "preferred_nodes_capabilities": []map[string]interface{}{
        {
            "id":           "node-1",
            "cpu_cores":    16,
            "mem_total":    int64(64 * 1024 * 1024 * 1024), // bytes
            "mem_available": int64(48 * 1024 * 1024 * 1024), // bytes
            "gpu_count":    2,
            "net_bw":       100.0, // Gbps
            "net_latency":  0.5,   // ms
        },
        // ... more nodes
    },
}
```

## Testing
Created comprehensive tests (`enhanced_manager_test.go`) covering:
1. **WithNodeCapabilities**: Parsing capabilities with native types
2. **WithInterfaceCapabilities**: Handling JSON unmarshaled float64 types
3. **FallbackToNodeIDs**: Graceful fallback to synthesized defaults
4. **OrderPreservation**: Maintaining node order from preferred_node_ids
5. **PartitionWithCapabilities**: End-to-end partitioning with real capabilities

All tests pass successfully, confirming:
- Correct parsing of all capability fields
- Proper type handling for JSON numbers
- GPU presence/absence handling
- Order preservation when both IDs and capabilities are provided
- Backward compatibility with existing ID-only approach

## Integration Points

### Producer Side (Adapter)
The adapter in `/home/kp/ollamamax/pkg/distributed/partitioning.go` creates the metadata:
```go
task.Metadata["preferred_node_ids"] = nodeIDs
task.Metadata["preferred_nodes_capabilities"] = nodeCapabilities // slice of maps
```

### Consumer Side (Enhanced Manager)
The enhanced manager in `/home/kp/ollamamax/ollama-distributed/pkg/scheduler/partitioning/enhanced_manager.go`:
1. Checks for `preferred_nodes_capabilities` first
2. Parses each capability map with type safety
3. Builds accurate `NodeInfo` structures
4. Falls back to `preferred_node_ids` with defaults if capabilities are absent
5. Maintains backward compatibility

## Benefits
1. **Accurate Scheduling**: Partition strategies now have real node capabilities for better decisions
2. **Resource Awareness**: Memory, CPU, GPU, and network capabilities inform partitioning
3. **Network-Aware**: High bandwidth nodes can be selected for tensor parallelism
4. **GPU Optimization**: GPU presence/count affects strategy selection
5. **Backward Compatible**: Falls back gracefully when capabilities are not provided

## Verification
```bash
# Run tests
go test ./pkg/scheduler/partitioning -run TestEnhancedManager -v

# Build to verify compilation
go build ./pkg/scheduler/partitioning/...
```

Both compilation and tests succeed, confirming the implementation is complete and functional.
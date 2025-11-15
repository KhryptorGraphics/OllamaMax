# Adaptive Peer Selection with Utility Scoring

This document explains how to use the adaptive peer selection system implemented in `peer_selection.go`.

## Overview

The peer selection system implements utility-based peer ranking using the formula:

```
U = (effective_FLOPs × available_memory) / (latency_penalty × estimated_congestion)
```

This approach dynamically selects the most suitable peers for distributed inference tasks based on their performance characteristics.

## Key Components

### PeerPerformanceTracker

The main component that tracks peer performance metrics and calculates utility scores.

```go
// Create a new tracker with default configuration
config := DefaultPeerSelectionConfig()
tracker := NewPeerPerformanceTracker(config)
defer tracker.Stop()
```

### PeerMetrics

Structure that holds performance metrics for each peer:

- `EffectiveFLOPs`: Processing power estimation
- `AvailableMemory`: Available system memory
- `Latency`: Network latency to peer
- `LatencyPenalty`: Calculated latency penalty
- `EstimatedCongestion`: Network congestion estimation
- `SuccessRate`: Task success rate
- And other supporting metrics

### Configuration

The `PeerSelectionConfig` allows customization of the selection behavior:

```go
config := &PeerSelectionConfig{
    LatencyWeight:        1.0,
    CongestionWeight:     1.0,
    MemoryWeight:         1.0,
    FLOPsWeight:          1.0,
    TopKPeers:            5,           // Number of peers to select
    ReevaluationInterval: 30 * time.Second,
    LatencyThreshold:     500 * time.Millisecond,
    MaxLatencyThreshold:  1000 * time.Millisecond, // Back-pressure threshold
}
```

## Usage Example

```go
// 1. Create the tracker
tracker := NewPeerPerformanceTracker(nil) // nil uses defaults
defer tracker.Stop()

// 2. Update peer metrics as they become available
peerMetrics := &PeerMetrics{
    EffectiveFLOPs:      15.0,           // GFLOPs
    AvailableMemory:     16 * 1024 * 1024 * 1024, // 16GB
    Latency:             30 * time.Millisecond,
    EstimatedCongestion: 0.5,
}
tracker.UpdatePeerMetrics(peerID, peerMetrics)

// 3. Record task completions to improve metrics
tracker.RecordTaskCompletion(peerID, true, 100*time.Millisecond, 1000000)

// 4. Select optimal peers for a task
selectedPeers, err := tracker.selectOptimalPeers(availablePeers, network)
```

## Features

### Dynamic K Selection

The system automatically determines the optimal number of peers to select based on diminishing returns in utility scores.

### Back-Pressure Mechanism

Peers with excessive latency (above `MaxLatencyThreshold`) are automatically filtered out to prevent performance degradation.

### Periodic Re-evaluation

Peer rankings are automatically re-evaluated every 30 seconds (configurable) to adapt to changing network conditions.

### Utility Score Calculation

The core utility function balances processing power, available resources, latency, and network congestion to provide optimal peer selection.

## Integration Points

To integrate with the existing P2P discovery system:

1. Extend `OptimizedConnectionInfo` with resource metrics
2. Update the `selectOptimalPeers` function to use utility scoring
3. Periodically update peer metrics from network monitoring
4. Use the tracker's methods to select peers for task assignment

The system is designed to work alongside existing libp2p networking components and can be integrated into any distributed scheduler.
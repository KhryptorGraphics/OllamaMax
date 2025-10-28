# Comment 3: P2P Bandwidth Tracking Implementation

## Overview
Successfully added bandwidth tracking metrics to the P2P node implementation in `/home/kp/OllamaMax/pkg/p2p/node.go`.

## Implementation Details

### 1. Metric Declaration (Line 81)
Added `bandwidthBytes` field to the `BasicNode` struct:
```go
bandwidthBytes       *prometheus.CounterVec
```

### 2. Metric Creation (Lines 133-139)
Created the bandwidth counter with direction labels in `NewBasicNode`:
```go
bandwidthBytes := prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "ollamamax_p2p_bandwidth_bytes_total",
        Help: "Total bytes sent/received over P2P network",
    },
    []string{"direction"},
)
```

### 3. Metric Registration (Line 158)
Registered the metric with the Prometheus registry:
```go
registry.MustRegister(bandwidthBytes)
```

### 4. Struct Assignment (Line 182)
Assigned the metric to the struct:
```go
bandwidthBytes:   bandwidthBytes,
```

### 5. Bandwidth Tracking - Sent (Line 276)
Added byte counting in the `Broadcast` method:
```go
// Track bandwidth sent
n.bandwidthBytes.WithLabelValues("sent").Add(float64(len(data)))
```

### 6. Bandwidth Tracking - Received (Line 295)
Added byte counting in the `Subscribe` method's wrapped handler:
```go
// Track bandwidth received
n.bandwidthBytes.WithLabelValues("received").Add(float64(len(data)))
```

### 7. RegisterTo Method (Line 341)
Included the metric in the `RegisterTo` method for main app registry exposure:
```go
collectors := []prometheus.Collector{
    n.connectedPeers,
    n.messagesSent,
    n.messagesReceived,
    n.bytesSent,
    n.bytesReceived,
    n.bandwidthBytes,  // Added
    n.messageLatency,
    n.connectionErrors,
}
```

## Metrics Exposed

### New Metric
- **Name**: `ollamamax_p2p_bandwidth_bytes_total`
- **Type**: Counter
- **Labels**: `direction` (values: "sent" or "received")
- **Description**: Total bytes sent/received over P2P network

### Usage Example
```promql
# Query total bytes sent
ollamamax_p2p_bandwidth_bytes_total{direction="sent"}

# Query total bytes received
ollamamax_p2p_bandwidth_bytes_total{direction="received"}

# Calculate total bandwidth
sum(ollamamax_p2p_bandwidth_bytes_total)

# Calculate bandwidth rate over 5 minutes
rate(ollamamax_p2p_bandwidth_bytes_total[5m])
```

## Verification

### Build Status
✅ Package compiles successfully: `go build ./pkg/p2p/...`

### Integration Points
1. Bandwidth is tracked on every `Broadcast` call (sent direction)
2. Bandwidth is tracked on every message received via `Subscribe` (received direction)
3. Metric is properly registered in both the local registry and main app registry
4. Metric follows Prometheus naming conventions with `ollamamax_` prefix

## Notes

### Existing Metrics
The implementation already had per-topic byte tracking:
- `p2p_bytes_sent_total{topic}` - Tracks bytes by topic
- `p2p_bytes_received_total{topic}` - Tracks bytes by topic

### New Metric Benefits
The new `ollamamax_p2p_bandwidth_bytes_total{direction}` metric provides:
- Aggregated bandwidth tracking across all topics
- Simple direction-based filtering (sent/received)
- Consistent naming with `ollamamax_` prefix
- Complements the existing per-topic metrics

## Files Modified
- `/home/kp/OllamaMax/pkg/p2p/node.go` - Added bandwidth tracking implementation

## Testing
The package compiles successfully. No test files exist in the P2P package, so manual testing or integration tests should be performed to verify metric collection at runtime.

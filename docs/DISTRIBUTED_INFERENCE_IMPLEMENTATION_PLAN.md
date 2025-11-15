# OllamaMax Distributed Inference Implementation Plan
## Research-Driven Architecture (Based on 70+ Papers)

**Version:** 1.0  
**Date:** November 2025  
**Status:** Ready for Implementation

---

## Overview

This plan transforms OllamaMax into a bandwidth-efficient distributed LLM inference platform using breakthrough techniques from November 2025 research. The architecture uses **hierarchical topology with adaptive compression** to overcome bandwidth limitations.

### Three-Tier Architecture

1. **Local Swarm Mode (LAN/VPN cluster)**: True model parallelism, ring partitioning, memory pooling
2. **Regional Swarm Mode (few clusters, good WAN)**: Co-inference with compression, speculative decoding, 300-800ms latency
3. **Global Swarm Mode (Internet P2P)**: Compressed latent exchanges, expert calls, batched validation, primarily offline/hard queries

---

## Phase A: Local Swarm Mode (Cluster-Level Pooling)
**Duration:** 3-4 weeks  
**Priority:** Critical  
**Goal:** Enable true model parallelism within tightly connected clusters

### A1: Hierarchical Cluster Topology
**File:** `pkg/distributed/topology.go`

**Tasks:**
- [ ] Implement `ClusterTopology` struct with LAN vs WAN classification
- [ ] Add network latency measurement (RTT, bandwidth, packet loss)
- [ ] Implement cluster discovery via UDP multicast (LAN) and gossip protocol (WAN)
- [ ] Add automatic cluster formation based on latency thresholds (<10ms = LAN, <50ms = regional, >50ms = global)
- [ ] Implement cluster coordinator election using Raft consensus

**Acceptance Criteria:**
- Nodes automatically form clusters based on network proximity
- Cluster membership updates within 5 seconds of topology changes
- Latency measurements accurate within ±5ms

### A2: Adaptive Peer Selection
**File:** `pkg/scheduler/peer_selection.go`

**Tasks:**
- [ ] Implement utility scoring: `U = (effective_FLOPs × available_memory) / (latency_penalty × estimated_congestion)`
- [ ] Add per-peer link quality estimation (bandwidth, RTT, packet loss)
- [ ] Implement dynamic K selection (top-K peers that yield net speedup)
- [ ] Add back-pressure mechanism to prevent adding peers that increase latency beyond threshold
- [ ] Implement peer ranking with periodic re-evaluation (every 30 seconds)

**Acceptance Criteria:**
- Peer selection reduces inference latency by 20%+ vs random selection
- Back-pressure prevents latency degradation when adding slow peers
- Peer rankings update within 30 seconds of network condition changes

### A3: Ring Memory-Weighted Partitioning
**File:** `pkg/distributed/ring_partition.go`

**Tasks:**
- [ ] Implement memory-weighted layer assignment (exo-style)
- [ ] Add contiguous layer allocation per device
- [ ] Implement ring-based inference flow (device N → device N+1 → ... → device 0)
- [ ] Add activation passing between devices with minimal serialization overhead
- [ ] Implement dynamic re-partitioning when devices join/leave

**Acceptance Criteria:**
- Models larger than any single device can run across cluster
- Memory utilization balanced within ±10% across devices
- Re-partitioning completes within 10 seconds without dropping requests

### A4: Block-Level Synchronization
**File:** `pkg/distributed/block_sync.go`

**Tasks:**
- [ ] Group layers into blocks (4-8 layers per block)
- [ ] Implement block-level synchronization instead of layer-by-layer
- [ ] Add pipelined execution (block N on device 1 while block N-1 on device 2)
- [ ] Implement prefetching of next block's weights
- [ ] Add overlap of communication and computation

**Acceptance Criteria:**
- Synchronization overhead reduced by 50%+ vs layer-by-layer
- Pipeline efficiency >80% (minimal bubble time)
- Prefetching hides 70%+ of weight transfer latency

### A5: Network-Aware Dynamic Re-Partitioning
**File:** `pkg/distributed/dynamic_partition.go`

**Tasks:**
- [ ] Monitor per-device inference latency and throughput
- [ ] Implement adaptive re-partitioning based on bottleneck detection
- [ ] Add cost model for re-partitioning decision (benefit vs overhead)
- [ ] Implement graceful migration (finish in-flight requests before re-partition)
- [ ] Add telemetry for partition efficiency metrics

**Acceptance Criteria:**
- Re-partitioning triggered when bottleneck detected (>20% latency increase)
- Migration completes without dropping requests
- Partition efficiency improves by 15%+ after re-partitioning

### A6: Back-Pressure Mechanisms
**File:** `pkg/scheduler/backpressure.go`

**Tasks:**
- [ ] Implement queue depth monitoring per peer
- [ ] Add flow control to prevent overwhelming slow peers
- [ ] Implement adaptive batch sizing based on peer capacity
- [ ] Add circuit breaker for consistently slow peers
- [ ] Implement graceful degradation (exclude slow peers temporarily)

**Acceptance Criteria:**
- Queue depths stay below 10 requests per peer
- Slow peers automatically excluded when latency >2x median
- Circuit breaker prevents cascading failures

---

## Phase B: Communication Optimization (NVRAR-Inspired)
**Duration:** 2-3 weeks  
**Priority:** High  
**Goal:** Replace NCCL with hierarchical all-reduce for 1.9-3.6x speedup

### B1: Three-Phase All-Reduce
**File:** `pkg/distributed/nvrar.go`

**Tasks:**
- [ ] Implement intra-node reduce-scatter (NVLink/fast interconnect)
- [ ] Implement inter-node recursive-doubling (minimize WAN hops)
- [ ] Implement intra-node all-gather
- [ ] Add topology-aware routing (prefer intra-node communication)
- [ ] Implement performance model: `T = 2(G-1)α_intra + log₂(N)α_inter + |M|/G[2(G-1)/β_intra + (N-1)η/(Nβ_inter)]`

**Acceptance Criteria:**
- Latency for 128KB-2MB messages 1.9-3.6x lower than NCCL
- Scaling efficiency >85% up to 128 GPUs
- Intra-node communication uses fast interconnect (NVLink/InfiniBand)

### B2: Chunked Non-Blocking Communication
**File:** `pkg/distributed/chunked_comm.go`

**Tasks:**
- [ ] Implement message chunking (optimal chunk size: 256KB-1MB)
- [ ] Add non-blocking send/receive with GPU-initiated transfers
- [ ] Implement pipelined chunk processing (overlap send/recv/compute)
- [ ] Add GPU SM utilization monitoring
- [ ] Implement adaptive chunk sizing based on network conditions

**Acceptance Criteria:**
- GPU SM utilization >90% during communication
- Pipelining hides 70%+ of communication latency
- Chunk size adapts to network bandwidth within 10 seconds

### B3: Fused Data-Flag Payloads
**File:** `pkg/distributed/fused_payload.go`

**Tasks:**
- [ ] Implement fused data-flag message format (avoid explicit signaling)
- [ ] Add sequence number based synchronization (no barriers)
- [ ] Implement out-of-order message handling with reordering buffer
- [ ] Add checksum validation for data integrity
- [ ] Implement fast path for in-order messages

**Acceptance Criteria:**
- Signaling overhead reduced by 80%+ vs explicit barriers
- Out-of-order messages handled correctly with <1ms reordering latency
- Data integrity maintained (zero corruption)

### B4: Decode-Heavy Workload Optimization
**File:** `pkg/distributed/decode_optimize.go`

**Tasks:**
- [ ] Implement small-message optimization (128KB-2MB)
- [ ] Add batching of decode requests (batch_size × hidden_dim)
- [ ] Implement KV cache management to reduce communication
- [ ] Add speculative prefetching of next token's activations
- [ ] Implement adaptive batching based on request arrival rate

**Acceptance Criteria:**
- Decode latency reduced by 40%+ vs naive implementation
- KV cache hit rate >90%
- Batching efficiency >80% (minimal padding overhead)

### B5: Benchmarking Suite
**File:** `tests/performance/nvrar_benchmark.go`

**Tasks:**
- [ ] Implement micro-benchmarks for all-reduce operations
- [ ] Add comparison with NCCL baseline
- [ ] Implement end-to-end inference benchmarks (Llama 3.1 70B, 405B)
- [ ] Add strong scaling experiments (4-128 GPUs)
- [ ] Implement automated performance regression detection

**Acceptance Criteria:**
- Benchmarks run automatically on every commit
- Performance regression detected within 5% threshold
- Results published to monitoring dashboard

---

## Phase C: Dynamic Activation Compression (EDGC-Inspired)
**Duration:** 3-4 weeks  
**Priority:** High  
**Goal:** Reduce inter-cluster communication by 46%+

### C1: Gradient/Activation Entropy Monitoring
**File:** `pkg/compression/entropy_monitor.go`

**Tasks:**
- [ ] Implement gradient data sampler (GSR=0.25, ISR=0.1 for 94% time reduction)
- [ ] Add entropy estimation using down-sampled gradients
- [ ] Implement window-based entropy tracking (w=1000 iterations optimal)
- [ ] Add entropy evolution visualization in monitoring dashboard
- [ ] Implement adaptive sampling rate based on entropy stability

**Acceptance Criteria:**
- Entropy estimation overhead <5% of total inference time
- Down-sampling reduces monitoring cost by 90%+
- Entropy estimates accurate within ±10% of full computation

### C2: Compression Quantification Model
**File:** `pkg/compression/cqm.go`

**Tasks:**
- [ ] Implement theoretical model linking entropy to compression rank: `r_new = g^(-1)(e^(H₀-H₁) * g(r₀))`
- [ ] Add compression rank calculation based on entropy evolution
- [ ] Implement coverage threshold validation (90%+ coverage required)
- [ ] Add adaptive rank adjustment with error feedback
- [ ] Implement rank bounds (min=8, max=512)

**Acceptance Criteria:**
- Compression rank adapts to entropy evolution within 1000 iterations
- Coverage threshold maintained >90%
- Rank adjustments improve compression ratio by 20%+ vs fixed rank

---

## Integration Points in Existing Codebase

### 1. P2P Networking (`pkg/p2p/`)
- Integrate cluster topology discovery
- Add hierarchical all-reduce to existing P2P protocol
- Implement activation compression at P2P message boundaries

### 2. Scheduler (`pkg/scheduler/`)
- Integrate adaptive peer selection
- Add utility scoring to existing node scoring
- Implement back-pressure mechanisms

### 3. Consensus (`pkg/consensus/`)
- Use Raft for cluster coordinator election
- Add cluster membership management
- Implement distributed configuration updates

### 4. Model Distribution (`pkg/models/`)
- Integrate ring memory-weighted partitioning
- Add dynamic re-partitioning logic
- Implement model shard management

---

## Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Intra-cluster latency | <50ms | P99 inference latency |
| Inter-cluster latency | <800ms | P99 inference latency |
| Communication reduction | >46% | Bytes transferred vs baseline |
| All-reduce speedup | 1.9-3.6x | vs NCCL for 128KB-2MB |
| Memory pooling efficiency | >85% | Utilization across devices |
| Peer selection accuracy | >90% | Correct top-K peers selected |

---

## Next Steps

1. **Immediate:** Start Phase A (Local Swarm Mode) - highest impact
2. **Week 2:** Begin Phase B (Communication Optimization) in parallel
3. **Week 4:** Start Phase C (Dynamic Compression)
4. **Week 6:** Integration testing and benchmarking
5. **Week 8:** Production deployment of Local Swarm Mode

**Total Timeline:** 8-10 weeks to production-ready Local Swarm Mode


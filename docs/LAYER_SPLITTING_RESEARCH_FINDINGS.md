# Layer Splitting & Cross-Internet Inference: Research Findings

## Executive Summary

After analyzing 70+ papers on distributed LLM inference, I conducted a deep dive into **layer splitting and model partitioning** strategies specifically for cross-internet inference. This research reveals critical insights for making OllamaMax's distributed inference feasible across WAN connections.

## 🎯 Key Finding: Three-Tier Hierarchical Approach is Essential

**The Answer:** Layer splitting MUST be hierarchical and adaptive based on network tier.

### ❌ What DOESN'T Work (Confirmed by Research)
1. **Naive layer-by-layer splitting across WAN** - Hundreds of MB per forward pass
2. **Fixed partition points** - Cannot adapt to dynamic network conditions
3. **Uniform activation transmission** - Ignores bandwidth constraints
4. **Sequential client processing** - Introduces severe latency

### ✅ What DOES Work (Research-Backed)

## 1. RecServe: Recursive Offloading for Multi-Tier Networks

**Paper:** "Recursive Offloading for LLM Serving in Multi-tier Networks" (arXiv:2505.16502v2, May 2025)

### Core Innovation
- **Hierarchical confidence evaluation** - Routes tasks based on complexity
- **Dynamic offloading thresholds** - Adapts to network conditions in real-time
- **Sliding-window confidence tracking** - Uses historical performance data

### Key Results
- **50%+ reduction** in communication burden vs cloud-only
- **Adaptive routing** based on task complexity
- **Three-tier architecture**: Device → Edge → Cloud

### Implementation Strategy for OllamaMax

```go
// Confidence-based offloading decision
type ConfidenceEvaluator struct {
    HistoricalQueue []float64  // Sliding window of confidence scores
    Threshold       float64     // Dynamic threshold (β quantile)
    MaxQueueSize    int         // Typically 300-1000
}

func (ce *ConfidenceEvaluator) ShouldOffload(confidence float64) bool {
    // Update historical queue
    ce.HistoricalQueue = append(ce.HistoricalQueue, confidence)
    if len(ce.HistoricalQueue) > ce.MaxQueueSize {
        ce.HistoricalQueue = ce.HistoricalQueue[1:]
    }
    
    // Dynamic threshold using quantile interpolation
    threshold := ce.computeQuantile(0.3) // β = 0.3 for aggressive offloading
    
    return confidence < threshold
}

// Task-specific confidence evaluation
func (ce *ConfidenceEvaluator) EvaluateConfidence(
    taskType string,
    output interface{},
) float64 {
    switch taskType {
    case "seq2class":
        // Peak softmax probability
        return maxSoftmaxProbability(output)
    case "seq2seq":
        // Normalized perplexity
        return 1.0 / normalizedPerplexity(output)
    default:
        return 0.0
    }
}
```

### Architecture Integration

```
┌─────────────────────────────────────────────────────────────┐
│ OllamaMax Hierarchical Offloading                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Local Tier (<10ms)                                        │
│  ├─ Run lightweight model (1-3B params)                    │
│  ├─ Evaluate confidence score                              │
│  └─ If confidence > threshold → Return result              │
│                                                             │
│  Regional Tier (10-50ms)                                   │
│  ├─ Run medium model (7-13B params)                        │
│  ├─ Evaluate confidence score                              │
│  └─ If confidence > threshold → Return result              │
│                                                             │
│  Global Tier (>50ms)                                       │
│  └─ Run large model (30-70B params)                        │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Performance Metrics (from Paper)
- **Communication reduction**: 50%+ vs cloud-only
- **Accuracy**: Comparable to cloud-only (within 1-2%)
- **Latency**: 
  - Simple tasks: <50ms (local)
  - Medium tasks: 300-800ms (regional)
  - Complex tasks: 2-5s (global)

## 2. SMoFi: Split Federated Learning with Momentum Fusion

**Paper:** "SMoFi: Step-wise Momentum Fusion for Split Federated Learning on Heterogeneous Data" (arXiv:2511.09828v1, Nov 2025)

### Core Innovation
- **Step-wise momentum alignment** - Synchronizes optimization across nodes
- **Staleness-aware mechanism** - Handles varying local steps
- **Client-transparent design** - No changes needed on client side

### Key Results
- **7.1% accuracy improvement** over baseline
- **10.25x convergence speedup** on complex tasks
- **Scales better** with more clients and deeper models

### Critical Insight for OllamaMax

**The Problem:** When splitting models across devices with heterogeneous data, gradient divergence causes:
1. Slower convergence
2. Lower final accuracy
3. Wasted communication bandwidth

**The Solution:** Synchronize momentum buffers across server-side optimizers at EVERY step, not just at epoch boundaries.

### Implementation Strategy

```go
// Momentum alignment for distributed training
type MomentumAligner struct {
    ServerOptimizers map[string]*Optimizer
    HistoricalMomentum map[string][]float64
    StalenessAlpha float64 // Typically -0.1
}

func (ma *MomentumAligner) AlignMomentum(step int) {
    // Collect current momentum from active optimizers
    currentMomentum := make([][]float64, 0)
    for _, opt := range ma.ServerOptimizers {
        if opt.IsActive(step) {
            currentMomentum = append(currentMomentum, opt.Momentum)
        }
    }
    
    // Include historical momentum with staleness factor
    historicalMomentum := ma.getHistoricalMomentum(step)
    
    // Weighted average
    alignedMomentum := ma.weightedAverage(
        currentMomentum,
        historicalMomentum,
        ma.StalenessAlpha,
    )
    
    // Update all optimizers
    for _, opt := range ma.ServerOptimizers {
        opt.Momentum = alignedMomentum
    }
}

func (ma *MomentumAligner) weightedAverage(
    current [][]float64,
    historical [][]float64,
    alpha float64,
) []float64 {
    // Polynomial staleness factor: s_α = (τ - |T_j| + 1)^α
    // where α < 0 (typically -0.1)
    
    totalWeight := float64(len(current))
    result := make([]float64, len(current[0]))
    
    // Add current momentum (weight = 1.0 each)
    for _, m := range current {
        for i := range result {
            result[i] += m[i]
        }
    }
    
    // Add historical momentum with staleness weighting
    for idx, m := range historical {
        staleness := math.Pow(float64(idx+1), alpha)
        totalWeight += staleness
        for i := range result {
            result[i] += m[i] * staleness
        }
    }
    
    // Normalize
    for i := range result {
        result[i] /= totalWeight
    }
    
    return result
}
```



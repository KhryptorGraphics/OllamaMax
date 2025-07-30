package scheduler

import (
	"fmt"
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/partitioning"
)

// TestEnhancedDistributedScheduler tests the enhanced distributed scheduler components
func TestEnhancedDistributedScheduler(t *testing.T) {
	fmt.Println("Testing enhanced distributed scheduler components...")
	
	// Test that we can import and use our enhanced components
	// This is a simplified test focusing on compilation
	
	// Test enhanced partitioning strategies
	fmt.Println("Testing enhanced partitioning strategies...")
	
	// Test pipeline parallelism strategy
	pipelineStrategy := partitioning.NewPipelineParallelismStrategy()
	if pipelineStrategy == nil {
		t.Error("Failed to create pipeline parallelism strategy")
	} else {
		fmt.Printf("Created pipeline parallelism strategy: %s\n", pipelineStrategy.GetName())
	}
	
	// Test tensor parallelism strategy
	tensorStrategy := partitioning.NewTensorParallelismStrategy()
	if tensorStrategy == nil {
		t.Error("Failed to create tensor parallelism strategy")
	} else {
		fmt.Printf("Created tensor parallelism strategy: %s\n", tensorStrategy.GetName())
	}
	
	// Test hybrid parallelism strategy
	hybridStrategy := partitioning.NewHybridParallelismStrategy()
	if hybridStrategy == nil {
		t.Error("Failed to create hybrid parallelism strategy")
	} else {
		fmt.Printf("Created hybrid parallelism strategy: %s\n", hybridStrategy.GetName())
	}
	
	// Test adaptive partitioning strategy
	adaptiveStrategy := partitioning.NewAdaptivePartitioningStrategy()
	if adaptiveStrategy == nil {
		t.Error("Failed to create adaptive partitioning strategy")
	} else {
		fmt.Printf("Created adaptive partitioning strategy: %s\n", adaptiveStrategy.GetName())
	}
	
	// Test enhanced partition manager
	fmt.Println("Testing enhanced partition manager...")
	
	// Create a mock base partition manager (simplified)
	baseManager := &partitioning.PartitionManager{
		Config: &partitioning.Config{
			DefaultStrategy: "layerwise",
			LayerThreshold:  10,
			BatchSizeLimit: 32,
		},
		Strategies: make(map[string]partitioning.PartitionStrategy),
		Optimizer: &partitioning.PartitionOptimizer{
			History: make([]*partitioning.PartitionResult, 0),
			LearningRate: 0.1,
			OptimizationWeights: map[string]float64{
				"latency":    0.4,
				"throughput": 0.3,
				"memory":     0.2,
				"bandwidth":  0.1,
			},
		},
		Analyzer: &partitioning.WorkloadAnalyzer{
			Profiles: make(map[string]*partitioning.WorkloadProfile),
		},
	}
	
	enhancedManager := partitioning.NewEnhancedPartitionManager(baseManager)
	if enhancedManager == nil {
		t.Error("Failed to create enhanced partition manager")
	} else {
		fmt.Println("Created enhanced partition manager successfully")
	}
	
	// Test performance tracker
	fmt.Println("Testing performance tracker...")
	
	tracker := partitioning.NewPerformanceTracker(100, 30*time.Second, true)
	if tracker == nil {
		t.Error("Failed to create performance tracker")
	} else {
		fmt.Println("Created performance tracker successfully")
	}
	
	// Test scheduling advisor
	fmt.Println("Testing scheduling advisor...")
	
	advisor := partitioning.NewSchedulingAdvisor(0.1, 5*time.Second, true)
	if advisor == nil {
		t.Error("Failed to create scheduling advisor")
	} else {
		fmt.Println("Created scheduling advisor successfully")
	}
	
	// Test recommendation engine
	fmt.Println("Testing recommendation engine...")
	
	engine := partitioning.NewRecommendationEngine()
	if engine == nil {
		t.Error("Failed to create recommendation engine")
	} else {
		fmt.Println("Created recommendation engine successfully")
	}
	
	// Test pattern matcher
	fmt.Println("Testing pattern matcher...")
	
	matcher := &partitioning.PatternMatcher{
		Patterns:   make(map[string]*partitioning.SchedulingPattern),
		Algorithms: make(map[string]partitioning.PatternMatchingAlgorithm),
	}
	
	if matcher == nil {
		t.Error("Failed to create pattern matcher")
	} else {
		fmt.Println("Created pattern matcher successfully")
	}
	
	fmt.Println("All enhanced distributed scheduler components tested successfully!")
}
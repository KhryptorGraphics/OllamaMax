//go:build ignore

package main

import (
	"fmt"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/partitioning"
)

func main() {
	// Test our enhanced partitioning strategies
	fmt.Println("Testing enhanced partitioning strategies...")

	// Test pipeline parallelism strategy
	fmt.Println("\nTesting pipeline parallelism strategy...")
	pipelineStrategy := partitioning.NewPipelineParallelismStrategy()
	fmt.Printf("Strategy name: %s\n", pipelineStrategy.GetName())

	// Test tensor parallelism strategy
	fmt.Println("\nTesting tensor parallelism strategy...")
	tensorStrategy := partitioning.NewTensorParallelismStrategy()
	fmt.Printf("Strategy name: %s\n", tensorStrategy.GetName())

	// Test hybrid parallelism strategy
	fmt.Println("\nTesting hybrid parallelism strategy...")
	hybridStrategy := partitioning.NewHybridParallelismStrategy()
	fmt.Printf("Strategy name: %s\n", hybridStrategy.GetName())

	// Test adaptive partitioning strategy
	fmt.Println("\nTesting adaptive partitioning strategy...")
	adaptiveStrategy := partitioning.NewAdaptivePartitioningStrategy()
	fmt.Printf("Strategy name: %s\n", adaptiveStrategy.GetName())

	// Test enhanced partition manager
	fmt.Println("\nTesting enhanced partition manager...")
	manager := partitioning.NewEnhancedPartitionManager(nil)
	fmt.Printf("Enhanced partition manager created: %v\n", manager != nil)

	if manager != nil {
		// Test available strategies
		strategies := manager.GetAvailableStrategies()
		fmt.Printf("Available strategies: %v\n", strategies)

		// Test metrics
		metrics := manager.GetStrategyMetrics()
		fmt.Printf("Strategy metrics count: %d\n", len(metrics))

		// Test history
		history := manager.GetSelectionHistory()
		fmt.Printf("Selection history length: %d\n", len(history))
	}

	fmt.Println("\nAll tests completed successfully!")
}

//go:build ignore

package main

import (
	"fmt"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/partitioning"
)

func main() {
	fmt.Println("Testing enhanced partition manager components...")

	// Create enhanced partition manager using factory function
	enhancedManager := partitioning.NewEnhancedPartitionManager()
	if enhancedManager == nil {
		fmt.Println("❌ Failed to create enhanced partition manager")
		return
	}
	fmt.Println("✅ Created enhanced partition manager using factory function")

	// Test individual strategy creation
	layerStrategy := partitioning.NewLayerwiseStrategy()
	if layerStrategy == nil {
		fmt.Println("❌ Failed to create layer-wise strategy")
		return
	}
	fmt.Println("✅ Created layer-wise strategy using factory function")

	tensorStrategy := partitioning.NewTensorParallelismStrategy()
	if tensorStrategy == nil {
		fmt.Println("❌ Failed to create tensor parallelism strategy")
		return
	}
	fmt.Println("✅ Created tensor parallelism strategy using factory function")

	pipelineStrategy := partitioning.NewPipelineParallelismStrategy()
	if pipelineStrategy == nil {
		fmt.Println("❌ Failed to create pipeline parallelism strategy")
		return
	}
	fmt.Println("✅ Created pipeline parallelism strategy using factory function")

	hybridStrategy := partitioning.NewHybridParallelismStrategy()
	if hybridStrategy == nil {
		fmt.Println("❌ Failed to create hybrid parallelism strategy")
		return
	}
	fmt.Println("✅ Created hybrid parallelism strategy using factory function")

	adaptiveStrategy := partitioning.NewAdaptivePartitioningStrategy()
	if adaptiveStrategy == nil {
		fmt.Println("❌ Failed to create adaptive partitioning strategy")
		return
	}
	fmt.Println("✅ Created adaptive partitioning strategy using factory function")

	// Test available strategies
	strategies := enhancedManager.GetAvailableStrategies()
	fmt.Printf("Available strategies: %v\n", strategies)

	// Test strategy metrics
	metrics := enhancedManager.GetStrategyMetrics()
	fmt.Printf("Strategy metrics count: %d\n", len(metrics))

	// Test selection history
	history := enhancedManager.GetSelectionHistory()
	fmt.Printf("Selection history length: %d\n", len(history))

	// Test factory utility functions
	availableStrategies := partitioning.GetAvailableStrategies()
	fmt.Printf("\nAvailable strategy types: %v\n", availableStrategies)

	for _, strategy := range availableStrategies {
		description := partitioning.GetStrategyDescription(strategy)
		fmt.Printf("- %s: %s\n", strategy, description)
	}

	fmt.Println("\n🎉 All enhanced partitioning components tested successfully!")
}

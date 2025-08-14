//go:build ignore

package main

import (
	"fmt"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/partitioning"
)

func main() {
	fmt.Println("Testing enhanced distributed scheduler components compilation...")

	// Test that we can create enhanced partitioning strategies
	fmt.Println("\nTesting enhanced partitioning strategies...")

	// Test pipeline parallelism strategy
	pipelineStrategy := partitioning.NewPipelineParallelismStrategy()
	if pipelineStrategy == nil {
		fmt.Println("❌ Failed to create pipeline parallelism strategy")
	} else {
		fmt.Printf("✅ Created pipeline parallelism strategy: %s\n", pipelineStrategy.GetName())
	}

	// Test tensor parallelism strategy
	tensorStrategy := partitioning.NewTensorParallelismStrategy()
	if tensorStrategy == nil {
		fmt.Println("❌ Failed to create tensor parallelism strategy")
	} else {
		fmt.Printf("✅ Created tensor parallelism strategy: %s\n", tensorStrategy.GetName())
	}

	// Test hybrid parallelism strategy
	hybridStrategy := partitioning.NewHybridParallelismStrategy()
	if hybridStrategy == nil {
		fmt.Println("❌ Failed to create hybrid parallelism strategy")
	} else {
		fmt.Printf("✅ Created hybrid parallelism strategy: %s\n", hybridStrategy.GetName())
	}

	// Test adaptive partitioning strategy
	adaptiveStrategy := partitioning.NewAdaptivePartitioningStrategy()
	if adaptiveStrategy == nil {
		fmt.Println("❌ Failed to create adaptive partitioning strategy")
	} else {
		fmt.Printf("✅ Created adaptive partitioning strategy: %s\n", adaptiveStrategy.GetName())
	}

	// Test creating enhanced partition manager
	fmt.Println("\nTesting enhanced partition manager...")

	// Create base partition manager
	baseConfig := &partitioning.Config{
		DefaultStrategy: "layerwise",
		LayerThreshold:  10,
		BatchSizeLimit:  32,
	}

	baseManager := partitioning.NewPartitionManager(baseConfig)
	if baseManager == nil {
		fmt.Println("❌ Failed to create base partition manager")
	} else {
		fmt.Println("✅ Created base partition manager")

		// Create enhanced partition manager
		enhancedManager := partitioning.NewEnhancedPartitionManager(baseManager)
		if enhancedManager == nil {
			fmt.Println("❌ Failed to create enhanced partition manager")
		} else {
			fmt.Println("✅ Created enhanced partition manager")

			// Test available strategies
			strategies := enhancedManager.GetAvailableStrategies()
			fmt.Printf("Available strategies: %v\n", strategies)

			// Test strategy metrics
			metrics := enhancedManager.GetStrategyMetrics()
			fmt.Printf("Strategy metrics count: %d\n", len(metrics))

			// Test selection history
			history := enhancedManager.GetSelectionHistory()
			fmt.Printf("Selection history length: %d\n", len(history))
		}
	}

	fmt.Println("\n🎉 All enhanced distributed scheduler components compiled successfully!")
}

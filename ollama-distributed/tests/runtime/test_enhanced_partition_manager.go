package main

import (
	"fmt"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/partitioning"
)

func main() {
	fmt.Println("Testing enhanced partition manager components...")
	
	// Create a base partition manager
	baseConfig := &partitioning.Config{
		DefaultStrategy: "layerwise",
		LayerThreshold:  10,
		BatchSizeLimit: 32,
	}
	
	baseManager := partitioning.NewPartitionManager(baseConfig)
	if baseManager == nil {
		fmt.Println("❌ Failed to create base partition manager")
		return
	}
	fmt.Println("✅ Created base partition manager")
	
	// Create enhanced partition manager
	enhancedManager := partitioning.NewEnhancedPartitionManager(baseManager)
	if enhancedManager == nil {
		fmt.Println("❌ Failed to create enhanced partition manager")
		return
	}
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
	
	fmt.Println("\n🎉 All enhanced partitioning components tested successfully!")
}
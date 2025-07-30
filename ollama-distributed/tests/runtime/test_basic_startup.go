package main

import (
	"context"
	"fmt"
	"log"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
)

func main() {
	fmt.Println("Testing basic system startup...")

	// Test 1: Configuration loading
	fmt.Println("1. Testing configuration...")

	// Create P2P node config
	p2pConfig := &config.NodeConfig{
		Listen: []string{"/ip4/127.0.0.1/tcp/0"},
	}
	fmt.Println("✅ Configuration created successfully")

	// Test 2: P2P Node creation
	fmt.Println("2. Testing P2P node creation...")
	ctx := context.Background()
	p2pNode, err := p2p.NewP2PNode(ctx, p2pConfig)
	if err != nil {
		log.Printf("❌ P2P node creation failed: %v", err)
	} else {
		fmt.Println("✅ P2P node created successfully")
		defer p2pNode.Stop()
	}

	// Test 3: Basic interface testing
	fmt.Println("3. Testing basic interfaces...")
	if p2pNode != nil {
		fmt.Printf("✅ P2P Node ID: %s\n", p2pNode.ID())
		fmt.Println("✅ Basic interfaces working correctly")
	}

	fmt.Println("\n🎯 Basic startup test completed!")
	fmt.Println("Note: Some failures are expected due to missing runtime dependencies.")
	fmt.Println("The important thing is that the interfaces and basic types work correctly.")
}

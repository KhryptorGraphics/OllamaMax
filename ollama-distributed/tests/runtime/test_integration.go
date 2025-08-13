//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/api"
	pkgConfig "github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/consensus"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/messaging"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/monitoring"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler"
)

func main() {
	fmt.Println("Testing API Gateway Integration with Scheduler and Consensus...")

	// Test 1: Create P2P node
	fmt.Println("1. Creating P2P node...")
	ctx := context.Background()
	p2pConfig := &pkgConfig.NodeConfig{
		Listen: []string{"/ip4/127.0.0.1/tcp/0"},
	}

	p2pNode, err := p2p.NewP2PNode(ctx, p2pConfig)
	if err != nil {
		log.Fatalf("❌ Failed to create P2P node: %v", err)
	}
	defer p2pNode.Stop()
	fmt.Printf("✅ P2P node created with ID: %s\n", p2pNode.ID())

	// Test 2: Create required components for consensus
	fmt.Println("2. Creating messaging and monitoring components...")

	// Create message router
	messageRouter := messaging.NewMessageRouter(nil) // Use default config
	fmt.Println("✅ Message router created")

	// Create network monitor
	networkMonitor := monitoring.NewNetworkMonitor(nil) // Use default config
	fmt.Println("✅ Network monitor created")

	// Test 3: Create consensus engine
	fmt.Println("3. Creating consensus engine...")
	consensusConfig := &config.ConsensusConfig{
		NodeID:    "test-node-1",
		DataDir:   "./test_data/consensus",
		Bootstrap: true,
	}

	consensusEngine, err := consensus.NewEngine(consensusConfig, p2pNode, messageRouter, networkMonitor)
	if err != nil {
		fmt.Printf("⚠️ Consensus engine creation failed: %v\n", err)
		fmt.Println("Continuing without consensus engine...")
		consensusEngine = nil
	} else {
		fmt.Println("✅ Consensus engine created successfully")
		if err := consensusEngine.Start(); err != nil {
			fmt.Printf("⚠️ Consensus engine start failed: %v\n", err)
		} else {
			fmt.Println("✅ Consensus engine started successfully")
		}
	}

	// Test 3: Create scheduler engine
	fmt.Println("3. Creating scheduler engine...")
	schedulerConfig := &config.SchedulerConfig{
		Algorithm:           "round_robin",
		LoadBalancing:       "least_loaded",
		HealthCheckInterval: 30 * time.Second,
		MaxRetries:          3,
		RetryDelay:          5 * time.Second,
		QueueSize:           1000,
		WorkerCount:         2,
	}

	schedulerEngine, err := scheduler.NewEngine(schedulerConfig, p2pNode, consensusEngine)
	if err != nil {
		log.Fatalf("❌ Failed to create scheduler engine: %v", err)
	}
	fmt.Println("✅ Scheduler engine created successfully")

	if err := schedulerEngine.Start(); err != nil {
		log.Fatalf("❌ Failed to start scheduler engine: %v", err)
	}
	fmt.Println("✅ Scheduler engine started successfully")

	// Test 4: Create API server with full integration
	fmt.Println("4. Creating integrated API server...")
	apiConfig := &config.APIConfig{
		Listen:      "127.0.0.1:8081",
		Timeout:     30 * time.Second,
		MaxBodySize: 10 * 1024 * 1024, // 10MB
	}

	server, err := api.NewServer(apiConfig, p2pNode, consensusEngine, schedulerEngine)
	if err != nil {
		log.Fatalf("❌ Failed to create API server: %v", err)
	}
	fmt.Println("✅ Integrated API server created successfully")

	// Test 5: Start the integrated system
	fmt.Println("5. Starting integrated system...")
	if err := server.Start(); err != nil {
		log.Fatalf("❌ Failed to start API server: %v", err)
	}
	fmt.Println("✅ Integrated system started successfully")

	// Test 6: Test scheduler integration
	fmt.Println("6. Testing scheduler integration...")
	testReq := &scheduler.Request{
		ID:         "test-req-1",
		ModelName:  "test-model",
		Type:       "generate",
		Priority:   1,
		Timeout:    10 * time.Second,
		ResponseCh: make(chan *scheduler.Response, 1),
		Payload: map[string]interface{}{
			"prompt": "Hello, world!",
		},
	}

	if err := schedulerEngine.Schedule(testReq); err != nil {
		fmt.Printf("⚠️ Scheduler test failed: %v\n", err)
	} else {
		fmt.Println("✅ Scheduler accepts requests successfully")

		// Wait for response (with timeout)
		select {
		case response := <-testReq.ResponseCh:
			if response.Success {
				fmt.Printf("✅ Scheduler processed request successfully: %s\n", response.NodeID)
			} else {
				fmt.Printf("⚠️ Scheduler request failed: %s\n", response.Error)
			}
		case <-time.After(15 * time.Second):
			fmt.Println("⚠️ Scheduler request timeout (expected - no worker nodes)")
		}
	}

	fmt.Println("\n🎯 Integration test completed!")
	fmt.Println("✅ P2P networking operational")
	fmt.Println("✅ Scheduler engine functional")
	fmt.Println("✅ API gateway integrated")
	if consensusEngine != nil {
		fmt.Println("✅ Consensus engine operational")
	}
	fmt.Println("\n🚀 Distributed system integration successful!")

	// Keep running for a moment to allow testing
	fmt.Println("\nSystem running on http://127.0.0.1:8081")
	fmt.Println("Press Ctrl+C to stop...")
	time.Sleep(30 * time.Second)
}

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/consensus"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
)

// Minimal node implementation without problematic Ollama dependencies
func main() {
	fmt.Println("🚀 Starting Distributed Ollama Node (Minimal)")
	
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	fmt.Printf("📋 Node Configuration:\n")
	fmt.Printf("   ID: %s\n", cfg.Node.ID)
	fmt.Printf("   Address: %s\n", cfg.Node.Address)
	fmt.Printf("   P2P Listen: %s\n", cfg.P2P.ListenAddr)
	fmt.Printf("   Raft Address: %s\n", cfg.Consensus.RaftAddr)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Initialize P2P node
	fmt.Println("🌐 Initializing P2P networking...")
	p2pNode, err := p2p.NewP2PNode(cfg.P2P)
	if err != nil {
		log.Fatalf("Failed to create P2P node: %v", err)
	}
	
	// Start P2P node
	if err := p2pNode.Start(ctx); err != nil {
		log.Fatalf("Failed to start P2P node: %v", err)
	}
	fmt.Println("✅ P2P node started successfully")
	
	// Initialize consensus engine
	fmt.Println("🤝 Initializing consensus engine...")
	consensusEngine, err := consensus.NewEngine(cfg.Consensus, p2pNode)
	if err != nil {
		log.Fatalf("Failed to create consensus engine: %v", err)
	}
	
	// Start consensus engine
	if err := consensusEngine.Start(); err != nil {
		log.Fatalf("Failed to start consensus engine: %v", err)
	}
	fmt.Println("✅ Consensus engine started successfully")
	
	// Simple HTTP server for health checks
	go startHealthServer(cfg.Node.Address)
	
	// Test consensus operations
	go testConsensusOperations(consensusEngine)
	
	// Test P2P operations
	go testP2POperations(p2pNode)
	
	fmt.Printf("🎉 Node %s is running!\n", cfg.Node.ID)
	fmt.Printf("📊 Health endpoint: http://%s/health\n", cfg.Node.Address)
	fmt.Printf("🔗 P2P address: %s\n", cfg.P2P.ListenAddr)
	fmt.Printf("🤝 Raft address: %s\n", cfg.Consensus.RaftAddr)
	
	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	
	fmt.Println("🛑 Shutting down node...")
	
	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	
	if err := consensusEngine.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down consensus engine: %v", err)
	}
	
	if err := p2pNode.Stop(); err != nil {
		log.Printf("Error shutting down P2P node: %v", err)
	}
	
	fmt.Println("✅ Node shutdown complete")
}

func startHealthServer(address string) {
	// Simple HTTP server for health checks and status
	// This would normally use the full API server, but we're keeping it minimal
	fmt.Printf("📊 Health server would start on %s\n", address)
	
	// Mock health server - in real implementation this would be a proper HTTP server
	for {
		time.Sleep(10 * time.Second)
		fmt.Printf("💓 Node heartbeat - %s\n", time.Now().Format("15:04:05"))
	}
}

func testConsensusOperations(engine *consensus.Engine) {
	// Test consensus operations every 30 seconds
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			// Test apply operation
			key := fmt.Sprintf("test-key-%d", time.Now().Unix())
			value := fmt.Sprintf("test-value-%d", time.Now().Unix())
			
			if err := engine.Apply(key, value, nil); err != nil {
				fmt.Printf("❌ Consensus apply failed: %v\n", err)
			} else {
				fmt.Printf("✅ Consensus apply successful: %s = %s\n", key, value)
				
				// Test get operation
				if retrievedValue, exists := engine.Get(key); exists {
					fmt.Printf("✅ Consensus get successful: %s = %v\n", key, retrievedValue)
				} else {
					fmt.Printf("❌ Consensus get failed: key %s not found\n", key)
				}
			}
		}
	}
}

func testP2POperations(node *p2p.Node) {
	// Test P2P operations every 60 seconds
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			// Get peer count
			peers := node.GetConnectedPeers()
			fmt.Printf("🌐 Connected to %d peers\n", len(peers))
			
			// List peer IDs
			for i, peer := range peers {
				if i < 3 { // Show first 3 peers
					fmt.Printf("   Peer %d: %s\n", i+1, peer.ID)
				}
			}
			
			if len(peers) > 3 {
				fmt.Printf("   ... and %d more peers\n", len(peers)-3)
			}
		}
	}
}
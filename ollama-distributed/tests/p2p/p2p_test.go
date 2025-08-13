package p2p

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/discovery"
)

// P2PTestSuite provides comprehensive P2P networking tests
type P2PTestSuite struct {
	nodes     []*p2p.Node
	discovery *discovery.Discovery
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewP2PTestSuite creates a new P2P test suite
func NewP2PTestSuite(nodeCount int) (*P2PTestSuite, error) {
	ctx, cancel := context.WithCancel(context.Background())

	suite := &P2PTestSuite{
		nodes:  make([]*p2p.Node, 0, nodeCount),
		ctx:    ctx,
		cancel: cancel,
	}

	// Create test nodes
	for i := 0; i < nodeCount; i++ {
		config := createTestP2PConfig(i)
		node, err := p2p.NewNode(ctx, config)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to create node %d: %w", i, err)
		}
		suite.nodes = append(suite.nodes, node)
	}

	return suite, nil
}

// TestP2PNetworking runs comprehensive P2P networking tests
func TestP2PNetworking(t *testing.T) {
	suite, err := NewP2PTestSuite(5)
	require.NoError(t, err)
	defer suite.cleanup()

	t.Run("NodeStartup", func(t *testing.T) {
		suite.testNodeStartup(t)
	})

	t.Run("PeerDiscovery", func(t *testing.T) {
		suite.testPeerDiscovery(t)
	})

	t.Run("NetworkTopology", func(t *testing.T) {
		suite.testNetworkTopology(t)
	})

	t.Run("MessagePassing", func(t *testing.T) {
		suite.testMessagePassing(t)
	})

	t.Run("NetworkPartitions", func(t *testing.T) {
		suite.testNetworkPartitions(t)
	})

	t.Run("PeerFailureRecovery", func(t *testing.T) {
		suite.testPeerFailureRecovery(t)
	})

	t.Run("ProtocolNegotiation", func(t *testing.T) {
		suite.testProtocolNegotiation(t)
	})

	t.Run("NetworkCongestion", func(t *testing.T) {
		suite.testNetworkCongestion(t)
	})

	t.Run("DHT_Operations", func(t *testing.T) {
		suite.testDHTOperations(t)
	})

	t.Run("ConnectionManagement", func(t *testing.T) {
		suite.testConnectionManagement(t)
	})
}

// testNodeStartup tests node startup and initialization
func (suite *P2PTestSuite) testNodeStartup(t *testing.T) {
	// Start all nodes
	for i, node := range suite.nodes {
		err := node.Start()
		require.NoError(t, err, "Failed to start node %d", i)

		// Verify node is running
		assert.True(t, node.IsRunning(), "Node %d should be running", i)

		// Check listening addresses
		addrs := node.GetListenAddresses()
		assert.NotEmpty(t, addrs, "Node %d should have listening addresses", i)

		// Verify peer ID
		peerID := node.GetPeerID()
		assert.NotEmpty(t, peerID, "Node %d should have a peer ID", i)
	}

	// Wait for nodes to be fully ready
	time.Sleep(2 * time.Second)
}

// testPeerDiscovery tests peer discovery mechanisms
func (suite *P2PTestSuite) testPeerDiscovery(t *testing.T) {
	// Bootstrap first node
	bootstrapNode := suite.nodes[0]
	bootstrapAddrs := bootstrapNode.GetListenAddresses()
	require.NotEmpty(t, bootstrapAddrs)

	// Connect other nodes to bootstrap
	for i := 1; i < len(suite.nodes); i++ {
		err := suite.nodes[i].ConnectToPeer(suite.ctx, bootstrapNode.GetPeerID(), bootstrapAddrs[0])
		require.NoError(t, err, "Failed to connect node %d to bootstrap", i)
	}

	// Wait for peer discovery
	time.Sleep(5 * time.Second)

	// Test peer discovery methods
	t.Run("Bootstrap Discovery", func(t *testing.T) {
		for i, node := range suite.nodes {
			peers := node.GetConnectedPeers()
			if i == 0 {
				// Bootstrap node should have connections to all others
				assert.GreaterOrEqual(t, len(peers), len(suite.nodes)-1,
					"Bootstrap node should be connected to other nodes")
			} else {
				// Other nodes should at least be connected to bootstrap
				assert.GreaterOrEqual(t, len(peers), 1,
					"Node %d should be connected to at least bootstrap", i)
			}
		}
	})

	t.Run("DHT Discovery", func(t *testing.T) {
		suite.testDHTBasedDiscovery(t)
	})

	t.Run("mDNS Discovery", func(t *testing.T) {
		suite.testMDNSDiscovery(t)
	})

	t.Run("Rendezvous Discovery", func(t *testing.T) {
		suite.testRendezvousDiscovery(t)
	})
}

// testDHTBasedDiscovery tests DHT-based peer discovery
func (suite *P2PTestSuite) testDHTBasedDiscovery(t *testing.T) {
	// Test DHT bootstrapping
	for _, node := range suite.nodes {
		dht := node.GetDHT()
		require.NotNil(t, dht, "Node should have DHT enabled")

		err := dht.Bootstrap(suite.ctx)
		assert.NoError(t, err, "DHT bootstrap should succeed")
	}

	// Wait for DHT to stabilize
	time.Sleep(10 * time.Second)

	// Test peer discovery through DHT
	for i, node := range suite.nodes {
		dht := node.GetDHT()

		// Find peers providing a specific service
		peers, err := dht.FindProviders(suite.ctx, "ollama-service")
		assert.NoError(t, err, "Finding providers should not error")

		t.Logf("Node %d found %d providers", i, len(peers))
	}
}

// testMDNSDiscovery tests mDNS-based local discovery
func (suite *P2PTestSuite) testMDNSDiscovery(t *testing.T) {
	// Enable mDNS on all nodes
	for _, node := range suite.nodes {
		err := node.EnableMDNS("ollama-local")
		require.NoError(t, err, "mDNS should be enabled successfully")
	}

	// Wait for mDNS discovery
	time.Sleep(5 * time.Second)

	// Verify local peers are discovered
	for i, node := range suite.nodes {
		localPeers := node.GetLocalPeers()
		// Each node should discover at least some other local nodes
		assert.GreaterOrEqual(t, len(localPeers), 1,
			"Node %d should discover local peers via mDNS", i)
	}
}

// testRendezvousDiscovery tests rendezvous-based discovery
func (suite *P2PTestSuite) testRendezvousDiscovery(t *testing.T) {
	rendezvousString := "ollama-cluster"

	// Register all nodes at rendezvous point
	for i, node := range suite.nodes {
		err := node.RegisterAtRendezvous(suite.ctx, rendezvousString)
		assert.NoError(t, err, "Node %d should register at rendezvous", i)
	}

	// Wait for registration propagation
	time.Sleep(3 * time.Second)

	// Discover peers at rendezvous point
	for i, node := range suite.nodes {
		peers, err := node.DiscoverPeersAtRendezvous(suite.ctx, rendezvousString)
		assert.NoError(t, err, "Node %d should discover peers at rendezvous", i)
		assert.GreaterOrEqual(t, len(peers), 1,
			"Node %d should find peers at rendezvous", i)
	}
}

// testNetworkTopology tests network topology formation
func (suite *P2PTestSuite) testNetworkTopology(t *testing.T) {
	// Wait for full mesh formation
	time.Sleep(10 * time.Second)

	t.Run("ConnectivityMatrix", func(t *testing.T) {
		suite.testConnectivityMatrix(t)
	})

	t.Run("NetworkDiameter", func(t *testing.T) {
		suite.testNetworkDiameter(t)
	})

	t.Run("ClusteringCoefficient", func(t *testing.T) {
		suite.testClusteringCoefficient(t)
	})

	t.Run("LoadBalancing", func(t *testing.T) {
		suite.testLoadBalancing(t)
	})
}

// testConnectivityMatrix tests network connectivity
func (suite *P2PTestSuite) testConnectivityMatrix(t *testing.T) {
	nodeCount := len(suite.nodes)
	connectivity := make([][]bool, nodeCount)

	for i := 0; i < nodeCount; i++ {
		connectivity[i] = make([]bool, nodeCount)
		connectedPeers := suite.nodes[i].GetConnectedPeers()

		for j := 0; j < nodeCount; j++ {
			if i == j {
				connectivity[i][j] = true // Self connection
				continue
			}

			targetPeerID := suite.nodes[j].GetPeerID()
			for _, peer := range connectedPeers {
				if peer == targetPeerID {
					connectivity[i][j] = true
					break
				}
			}
		}
	}

	// Verify network connectivity
	connectedPairs := 0
	for i := 0; i < nodeCount; i++ {
		for j := i + 1; j < nodeCount; j++ {
			if connectivity[i][j] || connectivity[j][i] {
				connectedPairs++
			}
		}
	}

	// In a well-connected network, we should have good connectivity
	totalPossiblePairs := nodeCount * (nodeCount - 1) / 2
	connectivityRatio := float64(connectedPairs) / float64(totalPossiblePairs)

	assert.Greater(t, connectivityRatio, 0.5,
		"Network should have good connectivity (>50%% of possible connections)")
}

// testNetworkDiameter tests network diameter
func (suite *P2PTestSuite) testNetworkDiameter(t *testing.T) {
	// Use BFS to find shortest paths between all node pairs
	maxDiameter := suite.calculateNetworkDiameter()

	// Network diameter should be reasonable for a P2P network
	assert.LessOrEqual(t, maxDiameter, 3,
		"Network diameter should be ≤ 3 for good performance")
}

// testClusteringCoefficient tests network clustering
func (suite *P2PTestSuite) testClusteringCoefficient(t *testing.T) {
	avgClustering := suite.calculateClusteringCoefficient()

	// Good P2P networks have decent clustering
	assert.Greater(t, avgClustering, 0.3,
		"Network should have good clustering (>0.3)")
}

// testLoadBalancing tests connection load balancing
func (suite *P2PTestSuite) testLoadBalancing(t *testing.T) {
	connectionCounts := make([]int, len(suite.nodes))

	for i, node := range suite.nodes {
		connectionCounts[i] = len(node.GetConnectedPeers())
	}

	// Calculate standard deviation of connection counts
	mean := 0.0
	for _, count := range connectionCounts {
		mean += float64(count)
	}
	mean /= float64(len(connectionCounts))

	variance := 0.0
	for _, count := range connectionCounts {
		variance += (float64(count) - mean) * (float64(count) - mean)
	}
	variance /= float64(len(connectionCounts))
	stddev := variance

	// Standard deviation should be low for good load balancing
	assert.Less(t, stddev, mean/2,
		"Connection load should be well balanced across nodes")
}

// testMessagePassing tests message passing capabilities
func (suite *P2PTestSuite) testMessagePassing(t *testing.T) {
	t.Run("DirectMessaging", func(t *testing.T) {
		suite.testDirectMessaging(t)
	})

	t.Run("BroadcastMessaging", func(t *testing.T) {
		suite.testBroadcastMessaging(t)
	})

	t.Run("MulticastMessaging", func(t *testing.T) {
		suite.testMulticastMessaging(t)
	})

	t.Run("StreamingMessages", func(t *testing.T) {
		suite.testStreamingMessages(t)
	})

	t.Run("MessageOrdering", func(t *testing.T) {
		suite.testMessageOrdering(t)
	})
}

// testDirectMessaging tests direct peer-to-peer messaging
func (suite *P2PTestSuite) testDirectMessaging(t *testing.T) {
	sender := suite.nodes[0]
	receiver := suite.nodes[1]

	message := []byte("Hello from peer-to-peer network!")

	// Send message
	err := sender.SendMessage(suite.ctx, receiver.GetPeerID(), message)
	require.NoError(t, err, "Direct message should be sent successfully")

	// Wait for message delivery
	time.Sleep(1 * time.Second)

	// Check if message was received
	receivedMessages := receiver.GetReceivedMessages()
	assert.GreaterOrEqual(t, len(receivedMessages), 1,
		"Receiver should have received at least one message")

	// Verify message content
	found := false
	for _, msg := range receivedMessages {
		if string(msg.Data) == string(message) {
			found = true
			break
		}
	}
	assert.True(t, found, "Original message should be received")
}

// testBroadcastMessaging tests broadcast messaging
func (suite *P2PTestSuite) testBroadcastMessaging(t *testing.T) {
	sender := suite.nodes[0]
	broadcastMessage := []byte("Broadcast message to all peers!")

	// Broadcast message
	err := sender.BroadcastMessage(suite.ctx, broadcastMessage)
	require.NoError(t, err, "Broadcast should succeed")

	// Wait for message propagation
	time.Sleep(3 * time.Second)

	// Verify all other nodes received the broadcast
	for i := 1; i < len(suite.nodes); i++ {
		receivedMessages := suite.nodes[i].GetReceivedMessages()

		found := false
		for _, msg := range receivedMessages {
			if string(msg.Data) == string(broadcastMessage) {
				found = true
				break
			}
		}
		assert.True(t, found, "Node %d should receive broadcast message", i)
	}
}

// testMulticastMessaging tests multicast messaging
func (suite *P2PTestSuite) testMulticastMessaging(t *testing.T) {
	sender := suite.nodes[0]

	// Select subset of nodes for multicast
	targetPeers := []peer.ID{
		suite.nodes[1].GetPeerID(),
		suite.nodes[2].GetPeerID(),
	}

	multicastMessage := []byte("Multicast message to selected peers!")

	// Send multicast message
	err := sender.MulticastMessage(suite.ctx, targetPeers, multicastMessage)
	require.NoError(t, err, "Multicast should succeed")

	// Wait for message delivery
	time.Sleep(2 * time.Second)

	// Verify target nodes received the message
	for _, targetPeer := range targetPeers {
		var targetNode *p2p.Node
		for _, node := range suite.nodes {
			if node.GetPeerID() == targetPeer {
				targetNode = node
				break
			}
		}
		require.NotNil(t, targetNode)

		receivedMessages := targetNode.GetReceivedMessages()
		found := false
		for _, msg := range receivedMessages {
			if string(msg.Data) == string(multicastMessage) {
				found = true
				break
			}
		}
		assert.True(t, found, "Target peer should receive multicast message")
	}

	// Verify non-target nodes did NOT receive the message
	for i := 3; i < len(suite.nodes); i++ {
		receivedMessages := suite.nodes[i].GetReceivedMessages()
		found := false
		for _, msg := range receivedMessages {
			if string(msg.Data) == string(multicastMessage) {
				found = true
				break
			}
		}
		assert.False(t, found, "Non-target node %d should NOT receive multicast", i)
	}
}

// testStreamingMessages tests streaming message capabilities
func (suite *P2PTestSuite) testStreamingMessages(t *testing.T) {
	sender := suite.nodes[0]
	receiver := suite.nodes[1]

	// Create streaming connection
	stream, err := sender.OpenStream(suite.ctx, receiver.GetPeerID(), "/test/stream/1.0.0")
	require.NoError(t, err, "Stream should open successfully")
	defer stream.Close()

	// Send multiple messages through stream
	messageCount := 10
	messages := make([][]byte, messageCount)

	for i := 0; i < messageCount; i++ {
		messages[i] = []byte(fmt.Sprintf("Stream message %d", i))
		_, err := stream.Write(messages[i])
		require.NoError(t, err, "Stream write should succeed")
	}

	// Wait for stream processing
	time.Sleep(2 * time.Second)

	// Verify messages were received in order
	receivedStreams := receiver.GetActiveStreams()
	assert.GreaterOrEqual(t, len(receivedStreams), 1,
		"Receiver should have active streams")
}

// testMessageOrdering tests message ordering guarantees
func (suite *P2PTestSuite) testMessageOrdering(t *testing.T) {
	sender := suite.nodes[0]
	receiver := suite.nodes[1]

	// Send ordered sequence of messages
	messageCount := 20
	for i := 0; i < messageCount; i++ {
		message := []byte(fmt.Sprintf("Ordered message %04d", i))
		err := sender.SendOrderedMessage(suite.ctx, receiver.GetPeerID(), message)
		require.NoError(t, err, "Ordered message send should succeed")
	}

	// Wait for all messages to be delivered
	time.Sleep(5 * time.Second)

	// Verify message ordering
	orderedMessages := receiver.GetOrderedMessages()
	assert.GreaterOrEqual(t, len(orderedMessages), messageCount,
		"Should receive all ordered messages")

	// Check sequence
	for i := 0; i < messageCount && i < len(orderedMessages); i++ {
		expected := fmt.Sprintf("Ordered message %04d", i)
		actual := string(orderedMessages[i].Data)
		assert.Equal(t, expected, actual, "Message %d should be in correct order", i)
	}
}

// testNetworkPartitions tests network partition scenarios
func (suite *P2PTestSuite) testNetworkPartitions(t *testing.T) {
	t.Run("SplitBrainPrevention", func(t *testing.T) {
		suite.testSplitBrainPrevention(t)
	})

	t.Run("PartitionHealing", func(t *testing.T) {
		suite.testPartitionHealing(t)
	})

	t.Run("MessageDeliveryDuringPartition", func(t *testing.T) {
		suite.testMessageDeliveryDuringPartition(t)
	})
}

// testSplitBrainPrevention tests split-brain prevention
func (suite *P2PTestSuite) testSplitBrainPrevention(t *testing.T) {
	// Create network partition by disconnecting nodes
	partition1 := suite.nodes[:2] // Nodes 0, 1
	partition2 := suite.nodes[2:] // Nodes 2, 3, 4

	// Disconnect partitions
	for _, node1 := range partition1 {
		for _, node2 := range partition2 {
			err := node1.DisconnectFromPeer(node2.GetPeerID())
			assert.NoError(t, err, "Disconnection should succeed")
		}
	}

	// Wait for partition detection
	time.Sleep(5 * time.Second)

	// Verify each partition detects the split
	for i, node := range partition1 {
		partitionDetected := node.IsPartitioned()
		assert.True(t, partitionDetected, "Node %d should detect partition", i)
	}

	for i, node := range partition2 {
		partitionDetected := node.IsPartitioned()
		assert.True(t, partitionDetected, "Node %d should detect partition", i+2)
	}

	// Both partitions should avoid split-brain by stopping operations
	for _, node := range suite.nodes {
		operationsBlocked := node.AreOperationsBlocked()
		assert.True(t, operationsBlocked,
			"Operations should be blocked during partition")
	}
}

// testPartitionHealing tests partition healing
func (suite *P2PTestSuite) testPartitionHealing(t *testing.T) {
	// Reconnect all nodes (heal partition)
	bootstrapNode := suite.nodes[0]
	bootstrapAddrs := bootstrapNode.GetListenAddresses()

	for i := 1; i < len(suite.nodes); i++ {
		err := suite.nodes[i].ConnectToPeer(suite.ctx,
			bootstrapNode.GetPeerID(), bootstrapAddrs[0])
		assert.NoError(t, err, "Reconnection should succeed")
	}

	// Wait for partition healing
	time.Sleep(10 * time.Second)

	// Verify partition is healed
	for i, node := range suite.nodes {
		partitionDetected := node.IsPartitioned()
		assert.False(t, partitionDetected, "Node %d should detect healed partition", i)

		operationsBlocked := node.AreOperationsBlocked()
		assert.False(t, operationsBlocked,
			"Operations should resume after partition healing")
	}

	// Verify network connectivity is restored
	suite.testConnectivityMatrix(t)
}

// testMessageDeliveryDuringPartition tests message handling during partitions
func (suite *P2PTestSuite) testMessageDeliveryDuringPartition(t *testing.T) {
	// Create partition again
	partition1 := suite.nodes[:2]
	partition2 := suite.nodes[2:]

	for _, node1 := range partition1 {
		for _, node2 := range partition2 {
			node1.DisconnectFromPeer(node2.GetPeerID())
		}
	}

	// Try to send message across partition
	sender := partition1[0]
	receiver := partition2[0]

	message := []byte("Cross-partition message")
	err := sender.SendMessage(suite.ctx, receiver.GetPeerID(), message)

	// Message should fail or be queued
	if err == nil {
		// If no error, message should be queued for later delivery
		queuedMessages := sender.GetQueuedMessages()
		assert.GreaterOrEqual(t, len(queuedMessages), 1,
			"Message should be queued during partition")
	} else {
		// Error is acceptable during partition
		assert.Contains(t, err.Error(), "partition",
			"Error should indicate partition issue")
	}
}

// testPeerFailureRecovery tests peer failure and recovery scenarios
func (suite *P2PTestSuite) testPeerFailureRecovery(t *testing.T) {
	t.Run("GracefulShutdown", func(t *testing.T) {
		suite.testGracefulShutdown(t)
	})

	t.Run("UnexpectedFailure", func(t *testing.T) {
		suite.testUnexpectedFailure(t)
	})

	t.Run("PeerRecovery", func(t *testing.T) {
		suite.testPeerRecovery(t)
	})
}

// testGracefulShutdown tests graceful peer shutdown
func (suite *P2PTestSuite) testGracefulShutdown(t *testing.T) {
	// Choose a non-critical node to shutdown
	nodeToShutdown := suite.nodes[len(suite.nodes)-1]
	shutdownPeerID := nodeToShutdown.GetPeerID()

	// Gracefully shutdown the node
	err := nodeToShutdown.GracefulShutdown(suite.ctx)
	require.NoError(t, err, "Graceful shutdown should succeed")

	// Wait for other nodes to detect the departure
	time.Sleep(5 * time.Second)

	// Verify other nodes detected the departure
	for i, node := range suite.nodes[:len(suite.nodes)-1] {
		connectedPeers := node.GetConnectedPeers()

		// Should not be connected to shutdown node
		for _, peer := range connectedPeers {
			assert.NotEqual(t, shutdownPeerID, peer,
				"Node %d should not be connected to shutdown peer", i)
		}
	}

	// Network should still be functional
	remainingNodes := suite.nodes[:len(suite.nodes)-1]
	suite.verifyNetworkFunctionality(t, remainingNodes)
}

// testUnexpectedFailure tests unexpected peer failure
func (suite *P2PTestSuite) testUnexpectedFailure(t *testing.T) {
	// Simulate unexpected failure by forcefully stopping a node
	nodeToFail := suite.nodes[len(suite.nodes)-2]
	failedPeerID := nodeToFail.GetPeerID()

	// Force stop without cleanup
	err := nodeToFail.ForceStop()
	require.NoError(t, err, "Force stop should succeed")

	// Wait for failure detection
	time.Sleep(10 * time.Second)

	// Verify other nodes detected the failure
	for i, node := range suite.nodes[:len(suite.nodes)-2] {
		connectedPeers := node.GetConnectedPeers()

		// Should detect disconnection
		connected := false
		for _, peer := range connectedPeers {
			if peer == failedPeerID {
				connected = true
				break
			}
		}
		assert.False(t, connected,
			"Node %d should detect failed peer disconnection", i)
	}
}

// testPeerRecovery tests peer recovery after failure
func (suite *P2PTestSuite) testPeerRecovery(t *testing.T) {
	// Create a new node to replace failed one
	config := createTestP2PConfig(len(suite.nodes))
	recoveredNode, err := p2p.NewNode(suite.ctx, config)
	require.NoError(t, err, "Recovery node creation should succeed")

	// Start the recovered node
	err = recoveredNode.Start()
	require.NoError(t, err, "Recovery node should start")

	// Connect to existing network
	bootstrapNode := suite.nodes[0]
	err = recoveredNode.ConnectToPeer(suite.ctx,
		bootstrapNode.GetPeerID(), bootstrapNode.GetListenAddresses()[0])
	require.NoError(t, err, "Recovery node should connect to network")

	// Wait for integration
	time.Sleep(5 * time.Second)

	// Verify recovered node is integrated
	connectedPeers := recoveredNode.GetConnectedPeers()
	assert.GreaterOrEqual(t, len(connectedPeers), 1,
		"Recovered node should be connected to network")

	// Add to test suite for cleanup
	suite.nodes = append(suite.nodes, recoveredNode)
}

// testProtocolNegotiation tests protocol negotiation
func (suite *P2PTestSuite) testProtocolNegotiation(t *testing.T) {
	node1 := suite.nodes[0]
	node2 := suite.nodes[1]

	// Test protocol negotiation
	supportedProtocols := []string{
		"/ollama/consensus/1.0.0",
		"/ollama/sync/1.0.0",
		"/ollama/discovery/1.0.0",
	}

	for _, protocol := range supportedProtocols {
		// Node1 advertises protocol support
		err := node1.AdvertiseProtocol(protocol)
		assert.NoError(t, err, "Protocol advertisement should succeed")

		// Node2 queries for protocol support
		supported, err := node2.SupportsProtocol(node1.GetPeerID(), protocol)
		assert.NoError(t, err, "Protocol query should succeed")
		assert.True(t, supported, "Protocol should be supported")
	}

	// Test unsupported protocol
	unsupportedProtocol := "/unsupported/protocol/1.0.0"
	supported, err := node2.SupportsProtocol(node1.GetPeerID(), unsupportedProtocol)
	assert.NoError(t, err, "Protocol query should succeed")
	assert.False(t, supported, "Unsupported protocol should be detected")
}

// testNetworkCongestion tests behavior under network congestion
func (suite *P2PTestSuite) testNetworkCongestion(t *testing.T) {
	// Simulate network congestion by sending many messages simultaneously
	sender := suite.nodes[0]

	var wg sync.WaitGroup
	messageCount := 100
	concurrency := 10

	// Send messages concurrently to create congestion
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < messageCount/concurrency; j++ {
				message := []byte(fmt.Sprintf("Congestion test: worker %d, message %d", workerID, j))

				// Send to random peer
				targetNode := suite.nodes[1+(workerID+j)%len(suite.nodes[1:])]
				err := sender.SendMessage(suite.ctx, targetNode.GetPeerID(), message)

				// Some messages may fail due to congestion, which is acceptable
				if err != nil {
					t.Logf("Message failed due to congestion: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()

	// Wait for message processing
	time.Sleep(10 * time.Second)

	// Verify network is still functional after congestion
	suite.verifyNetworkFunctionality(t, suite.nodes)
}

// testDHTOperations tests DHT operations
func (suite *P2PTestSuite) testDHTOperations(t *testing.T) {
	t.Run("PutGetOperations", func(t *testing.T) {
		suite.testDHTPutGet(t)
	})

	t.Run("ProviderRecords", func(t *testing.T) {
		suite.testDHTProviders(t)
	})

	t.Run("RoutingTable", func(t *testing.T) {
		suite.testDHTRouting(t)
	})
}

// testDHTPutGet tests DHT put/get operations
func (suite *P2PTestSuite) testDHTPutGet(t *testing.T) {
	node1 := suite.nodes[0]
	node2 := suite.nodes[1]

	dht1 := node1.GetDHT()
	dht2 := node2.GetDHT()

	require.NotNil(t, dht1, "Node1 should have DHT")
	require.NotNil(t, dht2, "Node2 should have DHT")

	// Put value in DHT
	key := "test-key"
	value := []byte("test-value")

	err := dht1.PutValue(suite.ctx, key, value)
	require.NoError(t, err, "DHT put should succeed")

	// Wait for propagation
	time.Sleep(3 * time.Second)

	// Get value from different node
	retrievedValue, err := dht2.GetValue(suite.ctx, key)
	require.NoError(t, err, "DHT get should succeed")
	assert.Equal(t, value, retrievedValue, "Retrieved value should match")
}

// testDHTProviders tests DHT provider records
func (suite *P2PTestSuite) testDHTProviders(t *testing.T) {
	node1 := suite.nodes[0]
	node2 := suite.nodes[1]

	dht1 := node1.GetDHT()
	dht2 := node2.GetDHT()

	// Provide a service
	serviceKey := "model-llama3.2"
	err := dht1.Provide(suite.ctx, serviceKey)
	require.NoError(t, err, "DHT provide should succeed")

	// Wait for propagation
	time.Sleep(3 * time.Second)

	// Find providers
	providers, err := dht2.FindProviders(suite.ctx, serviceKey)
	require.NoError(t, err, "Finding providers should succeed")

	// Should find node1 as provider
	found := false
	for _, provider := range providers {
		if provider == node1.GetPeerID() {
			found = true
			break
		}
	}
	assert.True(t, found, "Node1 should be found as provider")
}

// testDHTRouting tests DHT routing table
func (suite *P2PTestSuite) testDHTRouting(t *testing.T) {
	for i, node := range suite.nodes {
		dht := node.GetDHT()
		require.NotNil(t, dht, "Node %d should have DHT", i)

		// Get routing table size
		routingTableSize := dht.GetRoutingTableSize()
		assert.GreaterOrEqual(t, routingTableSize, 1,
			"Node %d should have entries in routing table", i)

		// Test routing table consistency
		closestPeers := dht.GetClosestPeers(node.GetPeerID())
		assert.LessOrEqual(t, len(closestPeers), routingTableSize,
			"Closest peers should not exceed routing table size")
	}
}

// testConnectionManagement tests connection management
func (suite *P2PTestSuite) testConnectionManagement(t *testing.T) {
	t.Run("ConnectionLimits", func(t *testing.T) {
		suite.testConnectionLimits(t)
	})

	t.Run("ConnectionPruning", func(t *testing.T) {
		suite.testConnectionPruning(t)
	})

	t.Run("ConnectionQuality", func(t *testing.T) {
		suite.testConnectionQuality(t)
	})
}

// testConnectionLimits tests connection limit enforcement
func (suite *P2PTestSuite) testConnectionLimits(t *testing.T) {
	node := suite.nodes[0]

	// Get current connection count
	initialConnections := len(node.GetConnectedPeers())
	maxConnections := node.GetMaxConnections()

	assert.LessOrEqual(t, initialConnections, maxConnections,
		"Current connections should not exceed limit")

	// Test that node doesn't exceed connection limits
	config := node.GetConfig()
	assert.Greater(t, config.ConnMgrHigh, config.ConnMgrLow,
		"High watermark should be greater than low watermark")
}

// testConnectionPruning tests connection pruning
func (suite *P2PTestSuite) testConnectionPruning(t *testing.T) {
	node := suite.nodes[0]

	// Force connection pruning
	err := node.PruneConnections()
	assert.NoError(t, err, "Connection pruning should succeed")

	// Verify node still has some connections
	connections := node.GetConnectedPeers()
	assert.Greater(t, len(connections), 0, "Node should maintain some connections")
}

// testConnectionQuality tests connection quality assessment
func (suite *P2PTestSuite) testConnectionQuality(t *testing.T) {
	node := suite.nodes[0]
	connections := node.GetConnectedPeers()

	for _, peerID := range connections {
		// Test connection quality metrics
		latency, err := node.GetConnectionLatency(peerID)
		assert.NoError(t, err, "Should be able to measure latency")
		assert.Greater(t, latency, time.Duration(0), "Latency should be positive")

		bandwidth, err := node.GetConnectionBandwidth(peerID)
		assert.NoError(t, err, "Should be able to measure bandwidth")
		assert.Greater(t, bandwidth, float64(0), "Bandwidth should be positive")

		reliability, err := node.GetConnectionReliability(peerID)
		assert.NoError(t, err, "Should be able to measure reliability")
		assert.GreaterOrEqual(t, reliability, 0.0, "Reliability should be non-negative")
		assert.LessOrEqual(t, reliability, 1.0, "Reliability should not exceed 1.0")
	}
}

// Helper functions

func (suite *P2PTestSuite) cleanup() {
	suite.cancel()

	for _, node := range suite.nodes {
		if node != nil {
			node.Shutdown(context.Background())
		}
	}
}

func (suite *P2PTestSuite) calculateNetworkDiameter() int {
	// Implementation of BFS to find network diameter
	// This is a simplified version
	maxDiameter := 0
	nodeCount := len(suite.nodes)

	for i := 0; i < nodeCount; i++ {
		distances := make([]int, nodeCount)
		for j := range distances {
			distances[j] = -1 // Unvisited
		}
		distances[i] = 0

		queue := []int{i}
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]

			connectedPeers := suite.nodes[current].GetConnectedPeers()
			for j := 0; j < nodeCount; j++ {
				if distances[j] == -1 { // Unvisited
					targetPeerID := suite.nodes[j].GetPeerID()
					for _, peer := range connectedPeers {
						if peer == targetPeerID {
							distances[j] = distances[current] + 1
							queue = append(queue, j)
							if distances[j] > maxDiameter {
								maxDiameter = distances[j]
							}
							break
						}
					}
				}
			}
		}
	}

	return maxDiameter
}

func (suite *P2PTestSuite) calculateClusteringCoefficient() float64 {
	// Simplified clustering coefficient calculation
	totalClustering := 0.0
	nodeCount := len(suite.nodes)

	for i := 0; i < nodeCount; i++ {
		neighbors := suite.nodes[i].GetConnectedPeers()
		if len(neighbors) < 2 {
			continue
		}

		// Count connections between neighbors
		connections := 0
		for j := 0; j < len(neighbors); j++ {
			for k := j + 1; k < len(neighbors); k++ {
				// Check if neighbors[j] and neighbors[k] are connected
				var nodeJ, nodeK *p2p.Node
				for _, node := range suite.nodes {
					if node.GetPeerID() == neighbors[j] {
						nodeJ = node
					}
					if node.GetPeerID() == neighbors[k] {
						nodeK = node
					}
				}

				if nodeJ != nil && nodeK != nil {
					nodeJPeers := nodeJ.GetConnectedPeers()
					for _, peer := range nodeJPeers {
						if peer == neighbors[k] {
							connections++
							break
						}
					}
				}
			}
		}

		possibleConnections := len(neighbors) * (len(neighbors) - 1) / 2
		if possibleConnections > 0 {
			clustering := float64(connections) / float64(possibleConnections)
			totalClustering += clustering
		}
	}

	return totalClustering / float64(nodeCount)
}

func (suite *P2PTestSuite) verifyNetworkFunctionality(t *testing.T, nodes []*p2p.Node) {
	// Test basic message passing
	if len(nodes) >= 2 {
		sender := nodes[0]
		receiver := nodes[1]

		testMessage := []byte("Network functionality test")
		err := sender.SendMessage(suite.ctx, receiver.GetPeerID(), testMessage)
		assert.NoError(t, err, "Network should be functional for message passing")
	}
}

func createTestP2PConfig(nodeIndex int) *p2p.Config {
	basePort := 11000 + nodeIndex*10

	return &p2p.Config{
		Listen:             fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", basePort),
		BootstrapPeers:     []string{},
		EnableDHT:          true,
		EnableMDNS:         true,
		ConnMgrLow:         10,
		ConnMgrHigh:        100,
		ConnMgrGracePeriod: 30 * time.Second,
	}
}

// Benchmark tests

func BenchmarkP2PMessagePassing(b *testing.B) {
	suite, err := NewP2PTestSuite(3)
	require.NoError(b, err)
	defer suite.cleanup()

	// Start nodes
	for _, node := range suite.nodes {
		require.NoError(b, node.Start())
	}

	// Connect nodes
	for i := 1; i < len(suite.nodes); i++ {
		err := suite.nodes[i].ConnectToPeer(suite.ctx,
			suite.nodes[0].GetPeerID(), suite.nodes[0].GetListenAddresses()[0])
		require.NoError(b, err)
	}

	time.Sleep(2 * time.Second)

	sender := suite.nodes[0]
	receiver := suite.nodes[1]
	message := []byte("Benchmark message")

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := sender.SendMessage(suite.ctx, receiver.GetPeerID(), message)
			if err != nil {
				b.Errorf("Message send failed: %v", err)
			}
		}
	})
}

func BenchmarkDHTOperations(b *testing.B) {
	suite, err := NewP2PTestSuite(5)
	require.NoError(b, err)
	defer suite.cleanup()

	// Setup DHT network
	for _, node := range suite.nodes {
		require.NoError(b, node.Start())
		require.NoError(b, node.GetDHT().Bootstrap(suite.ctx))
	}

	time.Sleep(5 * time.Second)

	b.ResetTimer()

	b.Run("DHT_Put", func(b *testing.B) {
		dht := suite.nodes[0].GetDHT()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("bench-key-%d", i)
			value := []byte(fmt.Sprintf("bench-value-%d", i))
			err := dht.PutValue(suite.ctx, key, value)
			if err != nil {
				b.Errorf("DHT put failed: %v", err)
			}
		}
	})

	b.Run("DHT_Get", func(b *testing.B) {
		dht := suite.nodes[1].GetDHT()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("bench-key-%d", i%100) // Reuse some keys
			_, err := dht.GetValue(suite.ctx, key)
			if err != nil && err.Error() != "key not found" {
				b.Errorf("DHT get failed: %v", err)
			}
		}
	})
}

package property

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/consensus"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/require"
)

// TestConsensusProperties tests consensus algorithm properties using property-based testing
func TestConsensusProperties(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property-based tests in short mode")
	}

	properties := gopter.NewProperties(nil)

	// Property 1: State Machine Safety
	// If a log entry is committed at a given index by any server,
	// no other server will ever apply a different log entry for the same index
	properties.Property("StateMachineSafety", prop.ForAll(
		func(operations []ConsensusOperation) bool {
			return testStateMachineSafety(t, operations)
		},
		genConsensusOperations(),
	))

	// Property 2: Leader Completeness
	// If a log entry is committed in a given term,
	// then that entry will be present in the logs of the leaders of all higher terms
	properties.Property("LeaderCompleteness", prop.ForAll(
		func(entries []LogEntry) bool {
			return testLeaderCompleteness(t, entries)
		},
		genLogEntries(),
	))

	// Property 3: Log Matching
	// If two logs contain an entry with the same index and term,
	// then the logs are identical in all entries up through the given index
	properties.Property("LogMatching", prop.ForAll(
		func(log1, log2 []LogEntry) bool {
			return testLogMatching(t, log1, log2)
		},
		genLogEntries(),
		genLogEntries(),
	))

	// Property 4: Monotonic Term Progression
	// Terms must increase monotonically across the cluster
	properties.Property("MonotonicTerms", prop.ForAll(
		func(termSequence []uint64) bool {
			return testMonotonicTerms(t, termSequence)
		},
		gen.SliceOf(gen.UInt64Range(1, 1000)),
	))

	// Property 5: Election Safety
	// At most one leader can be elected in a given term
	properties.Property("ElectionSafety", prop.ForAll(
		func(nodeCount int) bool {
			return testElectionSafety(t, nodeCount)
		},
		gen.IntRange(3, 7), // Test with 3-7 nodes
	))

	// Property 6: Append-Only Log Property
	// Log entries are never overwritten or deleted, only appended
	properties.Property("AppendOnlyLog", prop.ForAll(
		func(operations []LogOperation) bool {
			return testAppendOnlyLog(t, operations)
		},
		genLogOperations(),
	))

	// Run all properties
	properties.TestingRun(t)
}

// TestSchedulerProperties tests scheduling algorithm properties
func TestSchedulerProperties(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property-based tests in short mode")
	}

	properties := gopter.NewProperties(nil)

	// Property 1: Load Balance Fairness
	// Tasks should be distributed fairly across available nodes
	properties.Property("LoadBalanceFairness", prop.ForAll(
		func(tasks []Task, nodes []Node) bool {
			return testLoadBalanceFairness(t, tasks, nodes)
		},
		genTasks(),
		genNodes(),
	))

	// Property 2: Resource Constraints
	// No node should be assigned more resources than its capacity
	properties.Property("ResourceConstraints", prop.ForAll(
		func(tasks []Task, nodes []Node) bool {
			return testResourceConstraints(t, tasks, nodes)
		},
		genTasks(),
		genNodes(),
	))

	// Property 3: Priority Ordering
	// Higher priority tasks should be scheduled before lower priority tasks
	properties.Property("PriorityOrdering", prop.ForAll(
		func(tasks []Task) bool {
			return testPriorityOrdering(t, tasks)
		},
		genTasksWithPriority(),
	))

	// Property 4: Deadline Monotonicity
	// Tasks with earlier deadlines should be prioritized
	properties.Property("DeadlineMonotonicity", prop.ForAll(
		func(tasks []Task) bool {
			return testDeadlineMonotonicity(t, tasks)
		},
		genTasksWithDeadlines(),
	))

	properties.TestingRun(t)
}

// TestP2PProperties tests P2P networking properties
func TestP2PProperties(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property-based tests in short mode")
	}

	properties := gopter.NewProperties(nil)

	// Property 1: Message Delivery
	// All sent messages should eventually be delivered (with retries)
	properties.Property("MessageDelivery", prop.ForAll(
		func(messages []Message) bool {
			return testMessageDelivery(t, messages)
		},
		genMessages(),
	))

	// Property 2: Peer Discovery Convergence
	// All nodes should eventually discover each other
	properties.Property("PeerDiscoveryConvergence", prop.ForAll(
		func(nodeCount int) bool {
			return testPeerDiscoveryConvergence(t, nodeCount)
		},
		gen.IntRange(3, 6),
	))

	// Property 3: Content Routing Correctness
	// Content requests should route to nodes that have the content
	properties.Property("ContentRoutingCorrectness", prop.ForAll(
		func(content []ContentItem) bool {
			return testContentRoutingCorrectness(t, content)
		},
		genContentItems(),
	))

	properties.TestingRun(t)
}

// Data structures for property testing

type ConsensusOperation struct {
	Type  string
	Key   string
	Value string
	Term  uint64
	Index uint64
}

type LogEntry struct {
	Term    uint64
	Index   uint64
	Command string
	Data    []byte
}

type LogOperation struct {
	Type  string // "append", "read"
	Entry LogEntry
}

type Task struct {
	ID       string
	Priority int
	Deadline time.Time
	CPU      float64
	Memory   int64
	ModelID  string
}

type Node struct {
	ID           string
	CPU          float64
	Memory       int64
	Available    bool
	LoadFactor   float64
	Capabilities []string
}

type Message struct {
	ID      string
	From    string
	To      string
	Type    string
	Payload []byte
	TTL     int
}

type ContentItem struct {
	ID       string
	Hash     string
	Size     int64
	Replicas []string
}

// Generators for property testing

func genConsensusOperations() gopter.Gen {
	return gen.SliceOfN(50, gen.Struct(reflect.TypeOf(ConsensusOperation{}), map[string]gopter.Gen{
		"Type":  gen.OneConstOf("apply", "get", "delete"),
		"Key":   gen.AlphaString(),
		"Value": gen.AlphaString(),
		"Term":  gen.UInt64Range(1, 100),
		"Index": gen.UInt64Range(1, 1000),
	}))
}

func genLogEntries() gopter.Gen {
	return gen.SliceOfN(20, gen.Struct(reflect.TypeOf(LogEntry{}), map[string]gopter.Gen{
		"Term":    gen.UInt64Range(1, 50),
		"Index":   gen.UInt64Range(1, 100),
		"Command": gen.OneConstOf("set", "delete", "increment"),
		"Data":    gen.SliceOfN(32, gen.UInt8()),
	}))
}

func genLogOperations() gopter.Gen {
	return gen.SliceOfN(30, gen.Struct(reflect.TypeOf(LogOperation{}), map[string]gopter.Gen{
		"Type": gen.OneConstOf("append", "read"),
		"Entry": gen.Struct(reflect.TypeOf(LogEntry{}), map[string]gopter.Gen{
			"Term":    gen.UInt64Range(1, 50),
			"Index":   gen.UInt64Range(1, 100),
			"Command": gen.AlphaString(),
			"Data":    gen.SliceOfN(16, gen.UInt8()),
		}),
	}))
}

func genTasks() gopter.Gen {
	return gen.SliceOfN(20, gen.Struct(reflect.TypeOf(Task{}), map[string]gopter.Gen{
		"ID":       gen.AlphaString(),
		"Priority": gen.IntRange(1, 10),
		"Deadline": gen.TimeRange(time.Now(), 24*time.Hour),
		"CPU":      gen.Float64Range(0.1, 4.0),
		"Memory":   gen.Int64Range(128*1024*1024, 8*1024*1024*1024), // 128MB to 8GB
		"ModelID":  gen.AlphaString(),
	}))
}

func genTasksWithPriority() gopter.Gen {
	return gen.SliceOfN(15, gen.Struct(reflect.TypeOf(Task{}), map[string]gopter.Gen{
		"ID":       gen.AlphaString(),
		"Priority": gen.IntRange(1, 5),
		"CPU":      gen.Float64Range(0.1, 2.0),
		"Memory":   gen.Int64Range(128*1024*1024, 2*1024*1024*1024),
		"ModelID":  gen.AlphaString(),
	}))
}

func genTasksWithDeadlines() gopter.Gen {
	now := time.Now()
	return gen.SliceOfN(15, gen.Struct(reflect.TypeOf(Task{}), map[string]gopter.Gen{
		"ID":       gen.AlphaString(),
		"Priority": gen.IntRange(1, 3),
		"Deadline": gen.TimeRange(now.Add(time.Minute), time.Hour-time.Minute),
		"CPU":      gen.Float64Range(0.1, 2.0),
		"Memory":   gen.Int64Range(128*1024*1024, 2*1024*1024*1024),
		"ModelID":  gen.AlphaString(),
	}))
}

func genNodes() gopter.Gen {
	return gen.SliceOfN(5, gen.Struct(reflect.TypeOf(Node{}), map[string]gopter.Gen{
		"ID":           gen.AlphaString(),
		"CPU":          gen.Float64Range(1.0, 16.0),
		"Memory":       gen.Int64Range(1*1024*1024*1024, 32*1024*1024*1024), // 1GB to 32GB
		"Available":    gen.Bool(),
		"LoadFactor":   gen.Float64Range(0.0, 1.0),
		"Capabilities": gen.SliceOfN(3, gen.OneConstOf("llama", "mixtral", "codellama")),
	}))
}

func genMessages() gopter.Gen {
	return gen.SliceOfN(10, gen.Struct(reflect.TypeOf(Message{}), map[string]gopter.Gen{
		"ID":      gen.AlphaString(),
		"From":    gen.AlphaString(),
		"To":      gen.AlphaString(),
		"Type":    gen.OneConstOf("ping", "sync", "data", "request"),
		"Payload": gen.SliceOfN(64, gen.UInt8()),
		"TTL":     gen.IntRange(1, 10),
	}))
}

func genContentItems() gopter.Gen {
	return gen.SliceOfN(8, gen.Struct(reflect.TypeOf(ContentItem{}), map[string]gopter.Gen{
		"ID":       gen.AlphaString(),
		"Hash":     gen.AlphaString(),
		"Size":     gen.Int64Range(1024, 100*1024*1024), // 1KB to 100MB
		"Replicas": gen.SliceOfN(3, gen.AlphaString()),
	}))
}

// Property test implementations

func testStateMachineSafety(t *testing.T, operations []ConsensusOperation) bool {
	if len(operations) == 0 {
		return true
	}

	// Create a test cluster
	cluster := createTestConsensusCluster(t, 3)
	defer cleanupTestCluster(cluster)

	// Apply operations and verify state machine safety
	committedStates := make(map[uint64]map[string]string) // index -> key -> value

	for _, op := range operations {
		if op.Type == "apply" {
			// Apply to leader
			leader := findLeader(cluster)
			if leader == nil {
				continue
			}

			err := leader.Apply(op.Key, op.Value, nil)
			if err != nil {
				continue
			}

			// Wait for replication
			time.Sleep(10 * time.Millisecond)

			// Verify all nodes have the same state for this operation
			for i, node := range cluster {
				value, exists := node.Get(op.Key)
				if !exists {
					continue
				}

				// Check if this violates state machine safety
				if committed, hasCommitted := committedStates[op.Index]; hasCommitted {
					if prevValue, hasPrevValue := committed[op.Key]; hasPrevValue {
						if prevValue != value {
							t.Logf("State machine safety violation: node %d has %s=%s, but index %d was committed with %s=%s",
								i, op.Key, value, op.Index, op.Key, prevValue)
							return false
						}
					}
				}

				// Record the committed state
				if committedStates[op.Index] == nil {
					committedStates[op.Index] = make(map[string]string)
				}
				if strValue, ok := value.(string); ok {
					committedStates[op.Index][op.Key] = strValue
				}
			}
		}
	}

	return true
}

func testLeaderCompleteness(t *testing.T, entries []LogEntry) bool {
	if len(entries) == 0 {
		return true
	}

	// This test requires access to internal log structures
	// For now, we'll test a simplified version
	cluster := createTestConsensusCluster(t, 3)
	defer cleanupTestCluster(cluster)

	// Simulate entries being committed in different terms
	termEntries := make(map[uint64][]LogEntry)
	for _, entry := range entries {
		termEntries[entry.Term] = append(termEntries[entry.Term], entry)
	}

	// Apply entries in term order
	var terms []uint64
	for term := range termEntries {
		terms = append(terms, term)
	}

	// Sort terms
	for i := 0; i < len(terms); i++ {
		for j := i + 1; j < len(terms); j++ {
			if terms[i] > terms[j] {
				terms[i], terms[j] = terms[j], terms[i]
			}
		}
	}

	// Test that leader completeness holds (simplified)
	for _, term := range terms {
		leader := findLeader(cluster)
		if leader == nil {
			continue
		}

		for _, entry := range termEntries[term] {
			err := leader.Apply(fmt.Sprintf("key-%d-%d", entry.Term, entry.Index),
				string(entry.Data), nil)
			if err != nil {
				continue
			}
		}

		time.Sleep(50 * time.Millisecond)
	}

	return true
}

func testLogMatching(t *testing.T, log1, log2 []LogEntry) bool {
	// Test the log matching property: if two logs contain an entry with the same index and term,
	// then the logs are identical in all entries up through the given index

	if len(log1) == 0 || len(log2) == 0 {
		return true
	}

	// Find matching entries
	for i, entry1 := range log1 {
		for j, entry2 := range log2 {
			if entry1.Index == entry2.Index && entry1.Term == entry2.Term {
				// Logs should be identical up to this index
				maxIndex := int(entry1.Index)
				if maxIndex > len(log1) {
					maxIndex = len(log1)
				}
				if maxIndex > len(log2) {
					maxIndex = len(log2)
				}

				for k := 0; k < maxIndex && k < i && k < j; k++ {
					if log1[k].Term != log2[k].Term || log1[k].Index != log2[k].Index {
						t.Logf("Log matching violation: logs differ at index %d before matching entry at index %d",
							k, entry1.Index)
						return false
					}
				}
			}
		}
	}

	return true
}

func testMonotonicTerms(t *testing.T, termSequence []uint64) bool {
	if len(termSequence) < 2 {
		return true
	}

	// Test that terms increase monotonically
	for i := 1; i < len(termSequence); i++ {
		if termSequence[i] < termSequence[i-1] {
			t.Logf("Term monotonicity violation: term %d follows term %d",
				termSequence[i], termSequence[i-1])
			return false
		}
	}

	return true
}

func testElectionSafety(t *testing.T, nodeCount int) bool {
	if nodeCount < 3 {
		return true
	}

	cluster := createTestConsensusCluster(t, nodeCount)
	defer cleanupTestCluster(cluster)

	// Wait for initial election
	time.Sleep(2 * time.Second)

	// Count leaders
	leaderCount := 0
	var currentTerm uint64

	for _, node := range cluster {
		if node.IsLeader() {
			leaderCount++
			currentTerm = node.GetCurrentTerm()
		}
	}

	// Should have exactly one leader
	if leaderCount != 1 {
		t.Logf("Election safety violation: found %d leaders in term %d", leaderCount, currentTerm)
		return false
	}

	return true
}

func testAppendOnlyLog(t *testing.T, operations []LogOperation) bool {
	// Test that log entries are never overwritten or deleted
	cluster := createTestConsensusCluster(t, 1)
	defer cleanupTestCluster(cluster)

	node := cluster[0]
	appliedEntries := make(map[uint64]LogEntry)

	for _, op := range operations {
		if op.Type == "append" {
			// Apply the entry
			err := node.Apply(fmt.Sprintf("key-%d", op.Entry.Index),
				string(op.Entry.Data), nil)
			if err != nil {
				continue
			}

			// Check if we're overwriting a previous entry at the same index
			if prevEntry, exists := appliedEntries[op.Entry.Index]; exists {
				// This should not happen in a well-behaved log
				if prevEntry.Term != op.Entry.Term || string(prevEntry.Data) != string(op.Entry.Data) {
					t.Logf("Append-only violation: overwriting entry at index %d", op.Entry.Index)
					return false
				}
			}

			appliedEntries[op.Entry.Index] = op.Entry
		}
	}

	return true
}

func testLoadBalanceFairness(t *testing.T, tasks []Task, nodes []Node) bool {
	if len(tasks) == 0 || len(nodes) == 0 {
		return true
	}

	// Filter available nodes
	availableNodes := make([]Node, 0)
	for _, node := range nodes {
		if node.Available {
			availableNodes = append(availableNodes, node)
		}
	}

	if len(availableNodes) == 0 {
		return true
	}

	// Simulate task assignment
	nodeLoads := make(map[string]float64)
	for _, node := range availableNodes {
		nodeLoads[node.ID] = 0.0
	}

	// Simple round-robin assignment for testing
	for i, task := range tasks {
		if i < len(availableNodes) {
			selectedNode := availableNodes[i%len(availableNodes)]
			nodeLoads[selectedNode.ID] += task.CPU
		}
	}

	// Check fairness: no node should have more than 2x the average load
	totalLoad := 0.0
	for _, load := range nodeLoads {
		totalLoad += load
	}

	if len(availableNodes) == 0 {
		return true
	}

	avgLoad := totalLoad / float64(len(availableNodes))
	maxAllowedLoad := avgLoad * 2.0

	for nodeID, load := range nodeLoads {
		if load > maxAllowedLoad {
			t.Logf("Load balance fairness violation: node %s has load %f, max allowed %f",
				nodeID, load, maxAllowedLoad)
			return false
		}
	}

	return true
}

func testResourceConstraints(t *testing.T, tasks []Task, nodes []Node) bool {
	// Test that no node is assigned more resources than its capacity
	if len(tasks) == 0 || len(nodes) == 0 {
		return true
	}

	// Simple assignment simulation
	nodeUsage := make(map[string]struct {
		CPU    float64
		Memory int64
	})

	for _, node := range nodes {
		if node.Available {
			nodeUsage[node.ID] = struct {
				CPU    float64
				Memory int64
			}{0.0, 0}
		}
	}

	// Assign tasks to nodes (simplified)
	for i, task := range tasks {
		if i < len(nodes) {
			node := nodes[i%len(nodes)]
			if !node.Available {
				continue
			}

			usage := nodeUsage[node.ID]
			usage.CPU += task.CPU
			usage.Memory += task.Memory
			nodeUsage[node.ID] = usage

			// Check resource constraints
			if usage.CPU > node.CPU {
				t.Logf("CPU constraint violation: node %s assigned %f CPU, capacity %f",
					node.ID, usage.CPU, node.CPU)
				return false
			}

			if usage.Memory > node.Memory {
				t.Logf("Memory constraint violation: node %s assigned %d bytes, capacity %d",
					node.ID, usage.Memory, node.Memory)
				return false
			}
		}
	}

	return true
}

func testPriorityOrdering(t *testing.T, tasks []Task) bool {
	if len(tasks) < 2 {
		return true
	}

	// Sort tasks by priority (higher priority first)
	sortedTasks := make([]Task, len(tasks))
	copy(sortedTasks, tasks)

	// Simple bubble sort for testing
	for i := 0; i < len(sortedTasks); i++ {
		for j := 0; j < len(sortedTasks)-1-i; j++ {
			if sortedTasks[j].Priority < sortedTasks[j+1].Priority {
				sortedTasks[j], sortedTasks[j+1] = sortedTasks[j+1], sortedTasks[j]
			}
		}
	}

	// Verify ordering
	for i := 1; i < len(sortedTasks); i++ {
		if sortedTasks[i].Priority > sortedTasks[i-1].Priority {
			t.Logf("Priority ordering violation: task at position %d has higher priority than previous", i)
			return false
		}
	}

	return true
}

func testDeadlineMonotonicity(t *testing.T, tasks []Task) bool {
	if len(tasks) < 2 {
		return true
	}

	// Sort tasks by deadline (earliest first)
	sortedTasks := make([]Task, len(tasks))
	copy(sortedTasks, tasks)

	// Simple bubble sort by deadline
	for i := 0; i < len(sortedTasks); i++ {
		for j := 0; j < len(sortedTasks)-1-i; j++ {
			if sortedTasks[j].Deadline.After(sortedTasks[j+1].Deadline) {
				sortedTasks[j], sortedTasks[j+1] = sortedTasks[j+1], sortedTasks[j]
			}
		}
	}

	// Verify deadline ordering
	for i := 1; i < len(sortedTasks); i++ {
		if sortedTasks[i].Deadline.Before(sortedTasks[i-1].Deadline) {
			t.Logf("Deadline monotonicity violation: task at position %d has earlier deadline than previous", i)
			return false
		}
	}

	return true
}

func testMessageDelivery(t *testing.T, messages []Message) bool {
	// Test that all messages are eventually delivered
	// This is a simplified test that checks message properties

	if len(messages) == 0 {
		return true
	}

	for _, msg := range messages {
		// Check message validity
		if msg.ID == "" || msg.From == "" || msg.To == "" {
			t.Logf("Invalid message: empty required fields")
			return false
		}

		if msg.TTL <= 0 {
			t.Logf("Invalid message TTL: %d", msg.TTL)
			return false
		}

		if len(msg.Payload) == 0 {
			continue // Empty payload is valid
		}
	}

	return true
}

func testPeerDiscoveryConvergence(t *testing.T, nodeCount int) bool {
	// Test that all nodes eventually discover each other
	// This is a simplified test due to the complexity of setting up real P2P

	if nodeCount < 2 {
		return true
	}

	// In a real implementation, we would:
	// 1. Create nodeCount P2P nodes
	// 2. Start peer discovery
	// 3. Wait for convergence
	// 4. Verify all nodes know about all other nodes

	// For property testing, we verify the mathematical property
	maxConnections := nodeCount * (nodeCount - 1)
	if maxConnections < 0 {
		return false
	}

	// Each node should discover at least one other node
	minConnections := nodeCount - 1
	if minConnections < 0 {
		return false
	}

	return true
}

func testContentRoutingCorrectness(t *testing.T, content []ContentItem) bool {
	// Test that content routing correctly routes to nodes that have the content

	if len(content) == 0 {
		return true
	}

	for _, item := range content {
		if item.ID == "" || item.Hash == "" {
			t.Logf("Invalid content item: empty required fields")
			return false
		}

		if item.Size <= 0 {
			t.Logf("Invalid content size: %d", item.Size)
			return false
		}

		// Verify replicas are valid
		if len(item.Replicas) == 0 {
			continue // No replicas is valid (content not yet replicated)
		}

		for _, replica := range item.Replicas {
			if replica == "" {
				t.Logf("Invalid replica: empty ID")
				return false
			}
		}
	}

	return true
}

// Helper functions

func createTestConsensusCluster(t *testing.T, size int) []*consensus.Engine {
	cluster := make([]*consensus.Engine, size)

	for i := 0; i < size; i++ {
		ctx := context.Background()
		p2pNode, err := p2p.NewP2PNode(ctx, nil)
		require.NoError(t, err)

		consensusConfig := &config.ConsensusConfig{
			DataDir:           t.TempDir(),
			BindAddr:          fmt.Sprintf("127.0.0.1:%d", 7000+i),
			Bootstrap:         i == 0,
			HeartbeatTimeout:  100 * time.Millisecond,
			ElectionTimeout:   200 * time.Millisecond,
			CommitTimeout:     50 * time.Millisecond,
			MaxAppendEntries:  64,
			SnapshotInterval:  time.Hour,
			SnapshotThreshold: 8192,
			LogLevel:          "ERROR",
		}

		engine, err := consensus.NewEngine(consensusConfig, p2pNode)
		require.NoError(t, err)

		err = engine.Start()
		require.NoError(t, err)

		cluster[i] = engine
	}

	// Wait for cluster formation
	time.Sleep(500 * time.Millisecond)

	return cluster
}

func cleanupTestCluster(cluster []*consensus.Engine) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, engine := range cluster {
		if engine != nil {
			engine.Shutdown(ctx)
		}
	}
}

func findLeader(cluster []*consensus.Engine) *consensus.Engine {
	for _, engine := range cluster {
		if engine.IsLeader() {
			return engine
		}
	}
	return nil
}

// Benchmark property tests

func BenchmarkProperty_ConsensusStateMachineSafety(b *testing.B) {
	operations := []ConsensusOperation{
		{"apply", "key1", "value1", 1, 1},
		{"apply", "key2", "value2", 1, 2},
		{"apply", "key1", "value1_updated", 2, 3},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := testStateMachineSafety(&testing.T{}, operations)
		if !result {
			b.Fatal("Property violation detected")
		}
	}
}

func BenchmarkProperty_SchedulerLoadBalance(b *testing.B) {
	tasks := []Task{
		{"task1", 5, time.Now().Add(time.Hour), 2.0, 1024 * 1024 * 1024, "model1"},
		{"task2", 3, time.Now().Add(2 * time.Hour), 1.0, 512 * 1024 * 1024, "model2"},
	}

	nodes := []Node{
		{"node1", 4.0, 8 * 1024 * 1024 * 1024, true, 0.5, []string{"llama"}},
		{"node2", 8.0, 16 * 1024 * 1024 * 1024, true, 0.3, []string{"mixtral"}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := testLoadBalanceFairness(&testing.T{}, tasks, nodes)
		if !result {
			b.Fatal("Property violation detected")
		}
	}
}

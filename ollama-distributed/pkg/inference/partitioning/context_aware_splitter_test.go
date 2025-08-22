package partitioning

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestContextAwareSplitter(t *testing.T) {
	// Create splitter configuration
	config := &SplitterConfig{
		SemanticConfig: &SemanticConfig{
			NLPModelPath:       "/models/nlp",
			EmbeddingModelPath: "/models/embedding",
			CoherenceModelPath: "/models/coherence",
			BatchSize:          10,
			CacheSize:          100,
		},
		DependencyConfig: &DependencyConfig{
			MaxDepth:        5,
			MinStrength:     0.1,
			CacheSize:       50,
			AnalysisTimeout: 30 * time.Second,
		},
		ContextConfig: &ContextConfig{
			MaxContextSize:     1000,
			RetentionTime:      24 * time.Hour,
			StorageBackend:     "memory",
			CompressionEnabled: false,
		},
		OptimizationConfig: &OptimizationConfig{
			Algorithm:       "genetic",
			MaxIterations:   100,
			Tolerance:       0.01,
			LearningEnabled: true,
		},
	}

	// Create context-aware splitter
	splitter, err := NewContextAwareSplitter(config)
	if err != nil {
		t.Fatalf("Failed to create context-aware splitter: %v", err)
	}
	defer splitter.cancel()

	// Test basic splitting
	t.Run("BasicSplitting", func(t *testing.T) {
		testBasicSplitting(t, splitter)
	})

	// Test semantic splitting
	t.Run("SemanticSplitting", func(t *testing.T) {
		testSemanticSplitting(t, splitter)
	})

	// Test context preservation
	t.Run("ContextPreservation", func(t *testing.T) {
		testContextPreservation(t, splitter)
	})

	// Test dependency tracking
	t.Run("DependencyTracking", func(t *testing.T) {
		testDependencyTracking(t, splitter)
	})
}

func testBasicSplitting(t *testing.T, splitter *ContextAwareSplitter) {
	// Create a simple splitting request
	content := strings.Repeat("This is a test sentence. ", 100) // 100 sentences

	request := &SplittingRequest{
		RequestID:   "test-basic-split",
		Content:     content,
		ContentType: "text",
		ModelID:     "test-model",
		Parameters:  map[string]interface{}{},
		Constraints: &SplittingConstraints{
			MaxPartitions:    5,
			MinPartitionSize: 50,
			MaxPartitionSize: 500,
			OverlapSize:      10,
			PreserveContext:  true,
			AllowReordering:  false,
			Timeout:          30 * time.Second,
		},
		Context: &RequestContext{
			SessionID:      "session-1",
			UserID:         "user-1",
			ConversationID: "conv-1",
			Metadata:       map[string]interface{}{},
			Timestamp:      time.Now(),
		},
		Priority: 1,
		Deadline: time.Now().Add(time.Minute),
	}

	ctx := context.Background()
	result, err := splitter.SplitRequest(ctx, request)
	if err != nil {
		t.Fatalf("Basic splitting failed: %v", err)
	}

	// Verify result
	if result.RequestID != request.RequestID {
		t.Errorf("Expected request ID %s, got %s", request.RequestID, result.RequestID)
	}

	if len(result.Partitions) == 0 {
		t.Error("Expected at least one partition")
	}

	if len(result.Partitions) > request.Constraints.MaxPartitions {
		t.Errorf("Too many partitions: expected max %d, got %d",
			request.Constraints.MaxPartitions, len(result.Partitions))
	}

	// Verify partition sizes
	for i, partition := range result.Partitions {
		if len(partition.Content) < request.Constraints.MinPartitionSize {
			t.Errorf("Partition %d too small: %d < %d", i, len(partition.Content), request.Constraints.MinPartitionSize)
		}
		if len(partition.Content) > request.Constraints.MaxPartitionSize {
			t.Errorf("Partition %d too large: %d > %d", i, len(partition.Content), request.Constraints.MaxPartitionSize)
		}
	}

	// Verify processing time is reasonable
	if result.ProcessingTime > 5*time.Second {
		t.Errorf("Processing time too long: %v", result.ProcessingTime)
	}
}

func testSemanticSplitting(t *testing.T, splitter *ContextAwareSplitter) {
	// Create content with clear semantic boundaries
	content := `
	Introduction to Machine Learning
	Machine learning is a subset of artificial intelligence that focuses on algorithms.
	
	Types of Machine Learning
	There are three main types: supervised, unsupervised, and reinforcement learning.
	
	Supervised Learning
	In supervised learning, algorithms learn from labeled training data.
	
	Unsupervised Learning
	Unsupervised learning finds patterns in data without labeled examples.
	
	Reinforcement Learning
	Reinforcement learning uses rewards and penalties to learn optimal actions.
	
	Conclusion
	Machine learning continues to evolve and impact various industries.
	`

	request := &SplittingRequest{
		RequestID:   "test-semantic-split",
		Content:     content,
		ContentType: "text",
		ModelID:     "semantic-model",
		Parameters: map[string]interface{}{
			"strategy": "semantic",
		},
		Constraints: &SplittingConstraints{
			MaxPartitions:    6,
			MinPartitionSize: 20,
			MaxPartitionSize: 200,
			PreserveContext:  true,
		},
		Context: &RequestContext{
			SessionID: "session-2",
			UserID:    "user-1",
			Timestamp: time.Now(),
		},
	}

	ctx := context.Background()
	result, err := splitter.SplitRequest(ctx, request)
	if err != nil {
		t.Fatalf("Semantic splitting failed: %v", err)
	}

	// Verify semantic boundaries are respected
	if len(result.Partitions) == 0 {
		t.Error("Expected semantic partitions")
	}

	// Check that partitions contain meaningful content
	for i, partition := range result.Partitions {
		if len(strings.TrimSpace(partition.Content)) == 0 {
			t.Errorf("Partition %d is empty", i)
		}

		// Verify partition has semantic units
		if len(partition.SemanticUnits) == 0 {
			t.Errorf("Partition %d has no semantic units", i)
		}
	}
}

func testContextPreservation(t *testing.T, splitter *ContextAwareSplitter) {
	// Create content that requires context preservation
	content := `
	The variable x is defined as 5. Later in the code, we use x to calculate y = x * 2.
	The function process_data takes x as input. Inside process_data, we check if x > 0.
	If the condition is true, we return x squared. Otherwise, we return zero.
	The result is stored in variable result and printed to console.
	`

	request := &SplittingRequest{
		RequestID:   "test-context-preservation",
		Content:     content,
		ContentType: "code",
		ModelID:     "context-model",
		Parameters: map[string]interface{}{
			"strategy": "context_preserving",
		},
		Constraints: &SplittingConstraints{
			MaxPartitions:    3,
			MinPartitionSize: 30,
			MaxPartitionSize: 150,
			OverlapSize:      20,
			PreserveContext:  true,
		},
		Context: &RequestContext{
			SessionID: "session-3",
			UserID:    "user-1",
			Timestamp: time.Now(),
		},
	}

	ctx := context.Background()
	result, err := splitter.SplitRequest(ctx, request)
	if err != nil {
		t.Fatalf("Context preservation splitting failed: %v", err)
	}

	// Verify context is preserved through overlaps
	for i, partition := range result.Partitions {
		if request.Constraints.OverlapSize > 0 && i > 0 {
			// Check that there's some overlap with previous partition
			if len(partition.ContextBefore) == 0 {
				t.Errorf("Partition %d missing context before", i)
			}
		}

		if i < len(result.Partitions)-1 {
			// Check that there's context for next partition
			if len(partition.ContextAfter) == 0 {
				t.Errorf("Partition %d missing context after", i)
			}
		}
	}
}

func testDependencyTracking(t *testing.T, splitter *ContextAwareSplitter) {
	// Create content with clear dependencies
	content := `
	First, we define the base class Animal with method speak().
	Next, we create Dog class that inherits from Animal.
	The Dog class overrides the speak() method to return "Woof!".
	Then we create Cat class that also inherits from Animal.
	The Cat class overrides speak() to return "Meow!".
	Finally, we create instances of Dog and Cat and call their speak() methods.
	`

	request := &SplittingRequest{
		RequestID:   "test-dependency-tracking",
		Content:     content,
		ContentType: "code",
		ModelID:     "dependency-model",
		Parameters: map[string]interface{}{
			"strategy": "dependency_aware",
		},
		Constraints: &SplittingConstraints{
			MaxPartitions:    4,
			MinPartitionSize: 40,
			MaxPartitionSize: 120,
			PreserveContext:  true,
		},
		Context: &RequestContext{
			SessionID: "session-4",
			UserID:    "user-1",
			Timestamp: time.Now(),
		},
	}

	ctx := context.Background()
	result, err := splitter.SplitRequest(ctx, request)
	if err != nil {
		t.Fatalf("Dependency tracking splitting failed: %v", err)
	}

	// Verify dependencies are tracked
	if len(result.Dependencies) == 0 {
		t.Error("Expected dependency relationships")
	}

	// Check dependency structure
	for _, dep := range result.Dependencies {
		if dep.SourceID == "" || dep.TargetID == "" {
			t.Error("Dependency missing source or target ID")
		}

		if dep.Strength < 0 || dep.Strength > 1 {
			t.Errorf("Invalid dependency strength: %f", dep.Strength)
		}
	}

	// Verify partitions have dependency information
	for _, partition := range result.Partitions {
		if len(partition.Dependencies) > 0 {
			// Check that dependencies reference valid partitions
			for _, depID := range partition.Dependencies {
				found := false
				for _, p := range result.Partitions {
					if p.PartitionID == depID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Partition references non-existent dependency: %s", depID)
				}
			}
		}
	}
}

func TestSplittingStrategies(t *testing.T) {
	// Test semantic splitting strategy
	t.Run("SemanticStrategy", func(t *testing.T) {
		strategy := NewSemanticSplittingStrategy()

		if strategy.GetName() != "semantic" {
			t.Errorf("Expected strategy name 'semantic', got '%s'", strategy.GetName())
		}

		if !strategy.CanHandle("text") {
			t.Error("Semantic strategy should handle text content")
		}

		optimalSize := strategy.GetOptimalPartitionSize("test content", &SplittingConstraints{
			MaxPartitionSize: 1000,
		})

		if optimalSize <= 0 || optimalSize > 1000 {
			t.Errorf("Invalid optimal partition size: %d", optimalSize)
		}
	})

	// Test syntactic splitting strategy
	t.Run("SyntacticStrategy", func(t *testing.T) {
		strategy := NewSyntacticSplittingStrategy()

		if strategy.GetName() != "syntactic" {
			t.Errorf("Expected strategy name 'syntactic', got '%s'", strategy.GetName())
		}

		if !strategy.CanHandle("text") {
			t.Error("Syntactic strategy should handle text content")
		}
	})

	// Test sliding window strategy
	t.Run("SlidingWindowStrategy", func(t *testing.T) {
		strategy := NewSlidingWindowStrategy()

		if strategy.GetName() != "sliding_window" {
			t.Errorf("Expected strategy name 'sliding_window', got '%s'", strategy.GetName())
		}

		if !strategy.CanHandle("text") {
			t.Error("Sliding window strategy should handle text content")
		}
	})
}

func TestPerformanceMetrics(t *testing.T) {
	config := &SplitterConfig{
		SemanticConfig:     &SemanticConfig{BatchSize: 1},
		DependencyConfig:   &DependencyConfig{MaxDepth: 3},
		ContextConfig:      &ContextConfig{MaxContextSize: 100},
		OptimizationConfig: &OptimizationConfig{Algorithm: "simple"},
	}

	splitter, err := NewContextAwareSplitter(config)
	if err != nil {
		t.Fatalf("Failed to create splitter: %v", err)
	}
	defer splitter.cancel()

	// Test multiple requests to gather performance metrics
	for i := 0; i < 5; i++ {
		content := strings.Repeat("Test sentence. ", 50)

		request := &SplittingRequest{
			RequestID:   fmt.Sprintf("perf-test-%d", i),
			Content:     content,
			ContentType: "text",
			ModelID:     "perf-model",
			Constraints: &SplittingConstraints{
				MaxPartitions:    3,
				MinPartitionSize: 20,
				MaxPartitionSize: 200,
			},
			Context: &RequestContext{
				SessionID: fmt.Sprintf("session-%d", i),
				UserID:    "perf-user",
				Timestamp: time.Now(),
			},
		}

		ctx := context.Background()
		result, err := splitter.SplitRequest(ctx, request)
		if err != nil {
			t.Fatalf("Performance test %d failed: %v", i, err)
		}

		// Verify reasonable processing time
		if result.ProcessingTime > 2*time.Second {
			t.Errorf("Processing time too slow for request %d: %v", i, result.ProcessingTime)
		}
	}
}

func BenchmarkContextAwareSplitter(b *testing.B) {
	config := &SplitterConfig{
		SemanticConfig: &SemanticConfig{
			BatchSize: 1,
			CacheSize: 10,
		},
		DependencyConfig: &DependencyConfig{
			MaxDepth:  3,
			CacheSize: 10,
		},
		ContextConfig: &ContextConfig{
			MaxContextSize: 100,
		},
		OptimizationConfig: &OptimizationConfig{
			Algorithm:     "simple",
			MaxIterations: 10,
		},
	}

	splitter, err := NewContextAwareSplitter(config)
	if err != nil {
		b.Fatalf("Failed to create splitter: %v", err)
	}
	defer splitter.cancel()

	content := strings.Repeat("This is a benchmark test sentence. ", 100)

	request := &SplittingRequest{
		RequestID:   "bench-request",
		Content:     content,
		ContentType: "text",
		ModelID:     "bench-model",
		Constraints: &SplittingConstraints{
			MaxPartitions:    5,
			MinPartitionSize: 50,
			MaxPartitionSize: 300,
		},
		Context: &RequestContext{
			SessionID: "bench-session",
			UserID:    "bench-user",
			Timestamp: time.Now(),
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		_, err := splitter.SplitRequest(ctx, request)
		if err != nil {
			b.Fatalf("Benchmark iteration %d failed: %v", i, err)
		}
	}
}

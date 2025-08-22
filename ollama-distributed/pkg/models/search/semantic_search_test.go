package search

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestSemanticSearchEngine(t *testing.T) {
	// Create search engine configuration
	config := &SearchConfig{
		EmbeddingConfig: &EmbeddingConfig{
			ModelPath:  "",
			ModelType:  "simple_hash",
			Dimensions: 128,
			BatchSize:  10,
		},
		IndexType: "hnsw",
		CacheSize: 100,
		CacheTTL:  time.Hour,
	}

	// Create semantic search engine
	engine, err := NewSemanticSearchEngine(config)
	if err != nil {
		t.Fatalf("Failed to create search engine: %v", err)
	}
	defer engine.cancel()

	// Test model indexing
	t.Run("ModelIndexing", func(t *testing.T) {
		testModelIndexing(t, engine)
	})

	// Test semantic search
	t.Run("SemanticSearch", func(t *testing.T) {
		testSemanticSearch(t, engine)
	})

	// Test model updates
	t.Run("ModelUpdates", func(t *testing.T) {
		testModelUpdates(t, engine)
	})

	// Test search filters
	t.Run("SearchFilters", func(t *testing.T) {
		testSearchFilters(t, engine)
	})
}

func testModelIndexing(t *testing.T, engine *SemanticSearchEngine) {
	// Create test model metadata
	metadata := &ModelMetadata{
		Name:         "Test Language Model",
		Description:  "A test language model for natural language processing",
		Tags:         []string{"nlp", "language", "test"},
		Capabilities: []string{"text_generation", "question_answering"},
		Architecture: "transformer",
		Parameters:   7000000000, // 7B parameters
		License:      "MIT",
		Language:     []string{"english", "spanish"},
		Domain:       []string{"general", "conversational"},
		Performance: &PerformanceStats{
			Accuracy:    0.85,
			Latency:     150.0,
			Throughput:  100.0,
			MemoryUsage: 14000000000, // 14GB
			FLOPS:       1000000000,  // 1B FLOPS
		},
	}

	// Index the model
	err := engine.IndexModel("test-model-1", metadata)
	if err != nil {
		t.Fatalf("Failed to index model: %v", err)
	}

	// Verify the model was indexed
	if engine.vectorStore.GetSize() != 1 {
		t.Errorf("Expected 1 model in index, got %d", engine.vectorStore.GetSize())
	}

	// Index another model
	metadata2 := &ModelMetadata{
		Name:         "Code Generation Model",
		Description:  "A specialized model for code generation and programming tasks",
		Tags:         []string{"code", "programming", "generation"},
		Capabilities: []string{"code_generation", "code_completion"},
		Architecture: "transformer",
		Parameters:   13000000000, // 13B parameters
		License:      "Apache-2.0",
		Language:     []string{"python", "javascript", "go"},
		Domain:       []string{"programming", "software"},
		Performance: &PerformanceStats{
			Accuracy:    0.90,
			Latency:     200.0,
			Throughput:  80.0,
			MemoryUsage: 26000000000, // 26GB
			FLOPS:       1500000000,  // 1.5B FLOPS
		},
	}

	err = engine.IndexModel("test-model-2", metadata2)
	if err != nil {
		t.Fatalf("Failed to index second model: %v", err)
	}

	// Verify both models are indexed
	if engine.vectorStore.GetSize() != 2 {
		t.Errorf("Expected 2 models in index, got %d", engine.vectorStore.GetSize())
	}
}

func testSemanticSearch(t *testing.T, engine *SemanticSearchEngine) {
	// Search for language models
	query := &SearchQuery{
		Text:            "natural language processing model",
		TopK:            5,
		Threshold:       0.0,
		IncludeMetadata: true,
		Rerank:          false,
	}

	ctx := context.Background()
	response, err := engine.Search(ctx, query)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Verify search results
	if len(response.Results) == 0 {
		t.Error("Expected at least one search result")
	}

	// Check that results are sorted by score
	for i := 1; i < len(response.Results); i++ {
		if response.Results[i-1].Score < response.Results[i].Score {
			t.Error("Results should be sorted by score in descending order")
		}
	}

	// Verify metadata is included
	for _, result := range response.Results {
		if result.Metadata == nil {
			t.Error("Expected metadata to be included in results")
		}
	}

	// Search for code models
	codeQuery := &SearchQuery{
		Text:            "code generation programming",
		TopK:            3,
		Threshold:       0.0,
		IncludeMetadata: true,
		Rerank:          true,
	}

	codeResponse, err := engine.Search(ctx, codeQuery)
	if err != nil {
		t.Fatalf("Code search failed: %v", err)
	}

	// Should find the code model
	found := false
	for _, result := range codeResponse.Results {
		if result.ModelID == "test-model-2" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find code generation model in search results")
	}
}

func testModelUpdates(t *testing.T, engine *SemanticSearchEngine) {
	// Update model metadata
	updatedMetadata := &ModelMetadata{
		Name:         "Updated Test Language Model",
		Description:  "An updated test language model with improved capabilities",
		Tags:         []string{"nlp", "language", "test", "updated"},
		Capabilities: []string{"text_generation", "question_answering", "summarization"},
		Architecture: "transformer",
		Parameters:   7500000000, // 7.5B parameters
		License:      "MIT",
		Language:     []string{"english", "spanish", "french"},
		Domain:       []string{"general", "conversational", "academic"},
		Performance: &PerformanceStats{
			Accuracy:    0.88,
			Latency:     140.0,
			Throughput:  110.0,
			MemoryUsage: 15000000000, // 15GB
			FLOPS:       1100000000,  // 1.1B FLOPS
		},
	}

	// Update the model
	err := engine.UpdateModel("test-model-1", updatedMetadata)
	if err != nil {
		t.Fatalf("Failed to update model: %v", err)
	}

	// Search for the updated model
	query := &SearchQuery{
		Text:            "summarization capabilities",
		TopK:            5,
		Threshold:       0.0,
		IncludeMetadata: true,
	}

	ctx := context.Background()
	response, err := engine.Search(ctx, query)
	if err != nil {
		t.Fatalf("Search after update failed: %v", err)
	}

	// Verify the updated model is found
	found := false
	for _, result := range response.Results {
		if result.ModelID == "test-model-1" && result.Metadata != nil {
			// Check if updated capabilities are present
			for _, capability := range result.Metadata.Capabilities {
				if capability == "summarization" {
					found = true
					break
				}
			}
		}
	}

	if !found {
		t.Error("Expected to find updated model with summarization capability")
	}
}

func testSearchFilters(t *testing.T, engine *SemanticSearchEngine) {
	// Search with architecture filter
	query := &SearchQuery{
		Text:      "language model",
		TopK:      5,
		Threshold: 0.0,
		Filters: map[string]interface{}{
			"architecture": "transformer",
		},
		IncludeMetadata: true,
	}

	ctx := context.Background()
	response, err := engine.Search(ctx, query)
	if err != nil {
		t.Fatalf("Filtered search failed: %v", err)
	}

	// All results should have transformer architecture
	for _, result := range response.Results {
		if result.Metadata.Architecture != "transformer" {
			t.Errorf("Expected transformer architecture, got %s", result.Metadata.Architecture)
		}
	}

	// Search with parameter range filter
	paramQuery := &SearchQuery{
		Text:      "model",
		TopK:      5,
		Threshold: 0.0,
		Filters: map[string]interface{}{
			"min_parameters": int64(10000000000), // 10B+
		},
		IncludeMetadata: true,
	}

	paramResponse, err := engine.Search(ctx, paramQuery)
	if err != nil {
		t.Fatalf("Parameter filtered search failed: %v", err)
	}

	// All results should have >= 10B parameters
	for _, result := range paramResponse.Results {
		if result.Metadata.Parameters < 10000000000 {
			t.Errorf("Expected >= 10B parameters, got %d", result.Metadata.Parameters)
		}
	}

	// Search with language filter
	langQuery := &SearchQuery{
		Text:      "programming",
		TopK:      5,
		Threshold: 0.0,
		Filters: map[string]interface{}{
			"language": "python",
		},
		IncludeMetadata: true,
	}

	langResponse, err := engine.Search(ctx, langQuery)
	if err != nil {
		t.Fatalf("Language filtered search failed: %v", err)
	}

	// Should find models that support Python
	for _, result := range langResponse.Results {
		found := false
		for _, lang := range result.Metadata.Language {
			if lang == "python" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected to find models that support Python")
		}
	}
}

func TestEmbeddingModels(t *testing.T) {
	// Test TF-IDF embedding model
	t.Run("TFIDFModel", func(t *testing.T) {
		model := NewTFIDFEmbeddingModel(64)

		// Train on sample corpus
		corpus := []string{
			"natural language processing model",
			"code generation programming",
			"machine learning artificial intelligence",
			"transformer neural network",
		}

		err := model.TrainOnCorpus(corpus)
		if err != nil {
			t.Fatalf("Training failed: %v", err)
		}

		// Test encoding
		embedding, err := model.Encode("language model")
		if err != nil {
			t.Fatalf("Encoding failed: %v", err)
		}

		if len(embedding) != 64 {
			t.Errorf("Expected 64 dimensions, got %d", len(embedding))
		}

		// Test batch encoding
		texts := []string{"neural network", "programming language"}
		embeddings, err := model.EncodeBatch(texts)
		if err != nil {
			t.Fatalf("Batch encoding failed: %v", err)
		}

		if len(embeddings) != 2 {
			t.Errorf("Expected 2 embeddings, got %d", len(embeddings))
		}
	})

	// Test Simple Hash embedding model
	t.Run("SimpleHashModel", func(t *testing.T) {
		model := NewSimpleHashEmbeddingModel(32)

		embedding, err := model.Encode("test text")
		if err != nil {
			t.Fatalf("Encoding failed: %v", err)
		}

		if len(embedding) != 32 {
			t.Errorf("Expected 32 dimensions, got %d", len(embedding))
		}

		// Test consistency
		embedding2, err := model.Encode("test text")
		if err != nil {
			t.Fatalf("Second encoding failed: %v", err)
		}

		// Should be identical
		for i := range embedding {
			if embedding[i] != embedding2[i] {
				t.Error("Embeddings should be consistent for same input")
				break
			}
		}
	})
}

func BenchmarkSemanticSearch(b *testing.B) {
	config := &SearchConfig{
		EmbeddingConfig: &EmbeddingConfig{
			ModelType:  "simple_hash",
			Dimensions: 128,
		},
		IndexType: "hnsw",
		CacheSize: 1000,
		CacheTTL:  time.Hour,
	}

	engine, err := NewSemanticSearchEngine(config)
	if err != nil {
		b.Fatalf("Failed to create search engine: %v", err)
	}
	defer engine.cancel()

	// Index some models
	for i := 0; i < 100; i++ {
		metadata := &ModelMetadata{
			Name:        fmt.Sprintf("Model %d", i),
			Description: fmt.Sprintf("Test model number %d for benchmarking", i),
			Tags:        []string{"test", "benchmark"},
		}
		engine.IndexModel(fmt.Sprintf("model-%d", i), metadata)
	}

	query := &SearchQuery{
		Text:      "test model",
		TopK:      10,
		Threshold: 0.0,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := engine.Search(context.Background(), query)
		if err != nil {
			b.Fatalf("Search failed: %v", err)
		}
	}
}

package caching

import (
	"fmt"
	"testing"
	"time"
)

func TestIntelligentCache(t *testing.T) {
	// Create cache configuration
	config := &CacheConfig{
		PredictorConfig: &PredictorConfig{
			ModelPath:       "",
			UpdateInterval:  time.Minute,
			TrainingEnabled: true,
		},
		PrefetcherConfig: &PrefetcherConfig{
			Workers:       2,
			MaxConcurrent: 4,
			QueueSize:     100,
		},
	}

	// Create intelligent cache
	cache := NewIntelligentCache(1024*1024, config) // 1MB cache
	defer cache.cancel()

	// Test basic cache operations
	t.Run("BasicOperations", func(t *testing.T) {
		testBasicCacheOperations(t, cache)
	})

	// Test ML prediction
	t.Run("MLPrediction", func(t *testing.T) {
		testMLPrediction(t, cache)
	})

	// Test eviction policy
	t.Run("EvictionPolicy", func(t *testing.T) {
		testEvictionPolicy(t, cache)
	})

	// Test cache statistics
	t.Run("CacheStatistics", func(t *testing.T) {
		testCacheStatistics(t, cache)
	})
}

func testBasicCacheOperations(t *testing.T, cache *IntelligentCache) {
	// Test Put and Get
	modelID := "test-model-1"
	testData := []byte("test model data")
	metadata := map[string]interface{}{
		"model_type": "llm",
		"size":       len(testData),
	}

	// Put model in cache
	err := cache.Put(modelID, testData, metadata)
	if err != nil {
		t.Fatalf("Failed to put model in cache: %v", err)
	}

	// Get model from cache
	data, found := cache.Get(modelID)
	if !found {
		t.Fatal("Model not found in cache")
	}

	if string(data) != string(testData) {
		t.Errorf("Retrieved data doesn't match: expected %s, got %s", string(testData), string(data))
	}

	// Test cache miss
	_, found = cache.Get("non-existent-model")
	if found {
		t.Error("Expected cache miss for non-existent model")
	}
}

func testMLPrediction(t *testing.T, cache *IntelligentCache) {
	modelID := "test-model-prediction"

	// Add some training data
	for i := 0; i < 10; i++ {
		pattern := &UsagePattern{
			ModelID:     modelID,
			Timestamp:   time.Now().Add(-time.Duration(i) * time.Hour),
			AccessCount: int64(10 - i), // Decreasing access count
			UserContext: map[string]interface{}{
				"user_id":      "user123",
				"request_type": "inference",
			},
			TimeOfDay: (time.Now().Hour() - i) % 24,
			DayOfWeek: int(time.Now().Weekday()),
		}
		cache.predictor.AddTrainingData(pattern)
	}

	// Wait a bit for potential training
	time.Sleep(100 * time.Millisecond)

	// Test prediction
	context := map[string]interface{}{
		"user_id":      "user123",
		"request_type": "inference",
	}

	prediction, err := cache.predictor.PredictUsage(modelID, context)
	if err != nil {
		t.Fatalf("Prediction failed: %v", err)
	}

	if prediction.ModelID != modelID {
		t.Errorf("Expected model ID %s, got %s", modelID, prediction.ModelID)
	}

	if prediction.PredictedAccess < 0 || prediction.PredictedAccess > 1 {
		t.Errorf("Predicted access should be between 0 and 1, got %f", prediction.PredictedAccess)
	}

	if prediction.Confidence < 0 || prediction.Confidence > 1 {
		t.Errorf("Confidence should be between 0 and 1, got %f", prediction.Confidence)
	}
}

func testEvictionPolicy(t *testing.T, cache *IntelligentCache) {
	// Fill cache to capacity
	for i := 0; i < 10; i++ {
		modelID := fmt.Sprintf("model-%d", i)
		data := make([]byte, 200*1024) // 200KB each

		err := cache.Put(modelID, data, nil)
		if err != nil {
			t.Fatalf("Failed to put model %s: %v", modelID, err)
		}
	}

	// Add one more model to trigger eviction
	modelID := "overflow-model"
	data := make([]byte, 100*1024) // 100KB

	err := cache.Put(modelID, data, nil)
	if err != nil {
		t.Fatalf("Failed to put overflow model: %v", err)
	}

	// Check that some models were evicted
	stats := cache.GetStats()
	if stats["eviction_count"].(int64) == 0 {
		t.Error("Expected some evictions to occur")
	}

	// Verify the overflow model is still in cache
	_, found := cache.Get(modelID)
	if !found {
		t.Error("Overflow model should still be in cache")
	}
}

func testCacheStatistics(t *testing.T, cache *IntelligentCache) {
	// Perform some cache operations
	for i := 0; i < 5; i++ {
		modelID := fmt.Sprintf("stats-model-%d", i)
		data := []byte(fmt.Sprintf("data-%d", i))

		cache.Put(modelID, data, nil)
		cache.Get(modelID)        // Hit
		cache.Get("non-existent") // Miss
	}

	stats := cache.GetStats()

	// Check required statistics
	requiredStats := []string{
		"hit_count", "miss_count", "hit_rate", "prefetch_count",
		"eviction_count", "current_size", "max_size", "entry_count", "utilization",
	}

	for _, stat := range requiredStats {
		if _, exists := stats[stat]; !exists {
			t.Errorf("Missing statistic: %s", stat)
		}
	}

	// Verify hit rate calculation
	hitCount := stats["hit_count"].(int64)
	missCount := stats["miss_count"].(int64)
	hitRate := stats["hit_rate"].(float64)

	expectedHitRate := float64(hitCount) / float64(hitCount+missCount)
	if hitRate != expectedHitRate {
		t.Errorf("Hit rate calculation incorrect: expected %f, got %f", expectedHitRate, hitRate)
	}
}

func TestLinearRegressionModel(t *testing.T) {
	model := NewLinearRegressionModel(3)

	// Create simple training data
	trainingData := []TrainingExample{
		{Features: []float64{1.0, 0.5, 0.2}, Label: 1.0, Weight: 1.0},
		{Features: []float64{0.8, 0.3, 0.1}, Label: 1.0, Weight: 1.0},
		{Features: []float64{0.2, 0.1, 0.8}, Label: 0.0, Weight: 1.0},
		{Features: []float64{0.1, 0.2, 0.9}, Label: 0.0, Weight: 1.0},
	}

	// Train the model
	err := model.Train(trainingData)
	if err != nil {
		t.Fatalf("Training failed: %v", err)
	}

	// Test prediction
	prediction, err := model.Predict([]float64{0.9, 0.4, 0.1})
	if err != nil {
		t.Fatalf("Prediction failed: %v", err)
	}

	if prediction < 0 || prediction > 1 {
		t.Errorf("Prediction should be between 0 and 1, got %f", prediction)
	}

	// Test accuracy
	accuracy := model.GetAccuracy()
	if accuracy < 0 || accuracy > 1 {
		t.Errorf("Accuracy should be between 0 and 1, got %f", accuracy)
	}
}

func TestAdvancedFeatureExtractor(t *testing.T) {
	extractor := NewAdvancedFeatureExtractor()

	pattern := &UsagePattern{
		ModelID:     "test-model",
		Timestamp:   time.Now(),
		AccessCount: 5,
		UserContext: map[string]interface{}{
			"user_id":        "user123",
			"session_length": 2 * time.Hour,
			"request_type":   "inference",
		},
		TimeOfDay: 14, // 2 PM
		DayOfWeek: 2,  // Tuesday
	}

	features := extractor.ExtractFeatures(pattern)

	// Should have 13 features based on implementation
	expectedFeatureCount := 13
	if len(features) != expectedFeatureCount {
		t.Errorf("Expected %d features, got %d", expectedFeatureCount, len(features))
	}

	// All features should be normalized (between 0 and 1, or -1 and 1 for seasonal)
	for i, feature := range features {
		if feature < -1.0 || feature > 1.0 {
			t.Errorf("Feature %d out of range: %f", i, feature)
		}
	}
}

func TestPredictiveEvictionPolicy(t *testing.T) {
	// Create a mock predictor
	predictor := &UsagePredictor{}
	policy := NewPredictiveEvictionPolicy(predictor)

	// Create test cache entries
	now := time.Now()
	entries := []*CacheEntry{
		{
			ModelID:        "model1",
			Size:           1000,
			AccessCount:    10,
			LastAccessed:   now.Add(-time.Hour),
			PredictedScore: 0.8, // High prediction
		},
		{
			ModelID:        "model2",
			Size:           1000,
			AccessCount:    2,
			LastAccessed:   now.Add(-2 * time.Hour),
			PredictedScore: 0.2, // Low prediction
		},
		{
			ModelID:        "model3",
			Size:           1000,
			AccessCount:    5,
			LastAccessed:   now.Add(-25 * time.Hour), // Very old
			PredictedScore: 0.6,
		},
	}

	// Test eviction decisions
	if !policy.ShouldEvict(entries[1], 5000, 3000) {
		t.Error("Should evict model with low prediction score")
	}

	if !policy.ShouldEvict(entries[2], 5000, 3000) {
		t.Error("Should evict very old model")
	}

	if policy.ShouldEvict(entries[0], 2000, 3000) {
		t.Error("Should not evict model with high prediction score when under capacity")
	}

	// Test victim selection
	victim := policy.SelectVictim(entries)
	if victim == nil {
		t.Fatal("Should select a victim")
	}

	// Should select the entry with lowest eviction score
	if victim.ModelID != "model2" {
		t.Errorf("Expected to select model2, got %s", victim.ModelID)
	}
}

func BenchmarkIntelligentCache(b *testing.B) {
	config := &CacheConfig{
		PredictorConfig: &PredictorConfig{
			UpdateInterval:  time.Minute,
			TrainingEnabled: false, // Disable training for benchmark
		},
		PrefetcherConfig: &PrefetcherConfig{
			Workers:       1,
			MaxConcurrent: 2,
			QueueSize:     10,
		},
	}

	cache := NewIntelligentCache(10*1024*1024, config) // 10MB cache
	defer cache.cancel()

	// Pre-populate cache
	for i := 0; i < 100; i++ {
		modelID := fmt.Sprintf("bench-model-%d", i)
		data := make([]byte, 1024) // 1KB each
		cache.Put(modelID, data, nil)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		modelID := fmt.Sprintf("bench-model-%d", i%100)
		cache.Get(modelID)
	}
}

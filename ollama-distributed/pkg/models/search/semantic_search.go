package search

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// SemanticSearchEngine provides vector-based model discovery
type SemanticSearchEngine struct {
	vectorStore    *VectorStore
	embeddingModel EmbeddingModel
	indexManager   *IndexManager
	searchCache    *SearchCache
	ctx            context.Context
	cancel         context.CancelFunc
}

// VectorStore manages model embeddings and similarity search
type VectorStore struct {
	vectors      map[string]*ModelVector
	vectorsMutex sync.RWMutex
	index        VectorIndex
	dimensions   int
}

// ModelVector represents a model's embedding vector
type ModelVector struct {
	ModelID   string
	Vector    []float32
	Metadata  *ModelMetadata
	CreatedAt time.Time
	UpdatedAt time.Time
	Version   string
}

// ModelMetadata contains searchable model information
type ModelMetadata struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Tags         []string               `json:"tags"`
	Capabilities []string               `json:"capabilities"`
	Architecture string                 `json:"architecture"`
	Parameters   int64                  `json:"parameters"`
	License      string                 `json:"license"`
	Language     []string               `json:"language"`
	Domain       []string               `json:"domain"`
	Performance  *PerformanceStats      `json:"performance"`
	Attributes   map[string]interface{} `json:"attributes"`
}

// PerformanceStats contains model performance metrics
type PerformanceStats struct {
	Accuracy    float64 `json:"accuracy"`
	Latency     float64 `json:"latency"`
	Throughput  float64 `json:"throughput"`
	MemoryUsage int64   `json:"memory_usage"`
	FLOPS       int64   `json:"flops"`
}

// SearchQuery represents a semantic search query
type SearchQuery struct {
	Text            string                 `json:"text"`
	Vector          []float32              `json:"vector,omitempty"`
	Filters         map[string]interface{} `json:"filters"`
	TopK            int                    `json:"top_k"`
	Threshold       float64                `json:"threshold"`
	IncludeMetadata bool                   `json:"include_metadata"`
	Rerank          bool                   `json:"rerank"`
}

// SearchResult represents a search result
type SearchResult struct {
	ModelID     string         `json:"model_id"`
	Score       float64        `json:"score"`
	Metadata    *ModelMetadata `json:"metadata,omitempty"`
	Explanation string         `json:"explanation,omitempty"`
	Rank        int            `json:"rank"`
}

// SearchResponse contains search results and metadata
type SearchResponse struct {
	Results     []*SearchResult        `json:"results"`
	TotalCount  int                    `json:"total_count"`
	QueryTime   time.Duration          `json:"query_time"`
	QueryVector []float32              `json:"query_vector,omitempty"`
	Filters     map[string]interface{} `json:"filters"`
}

// EmbeddingModel interface for generating embeddings
type EmbeddingModel interface {
	Encode(text string) ([]float32, error)
	EncodeBatch(texts []string) ([][]float32, error)
	GetDimensions() int
	GetModelName() string
}

// VectorIndex interface for efficient similarity search
type VectorIndex interface {
	Add(id string, vector []float32) error
	Search(query []float32, topK int, threshold float64) ([]*IndexResult, error)
	Remove(id string) error
	Update(id string, vector []float32) error
	GetSize() int
	Optimize() error
}

// IndexResult represents an index search result
type IndexResult struct {
	ID    string
	Score float64
}

// IndexManager manages vector index lifecycle
type IndexManager struct {
	index      VectorIndex
	indexType  string
	dimensions int
	buildQueue chan *IndexBuildTask
	ctx        context.Context
	cancel     context.CancelFunc
}

// IndexBuildTask represents an index building task
type IndexBuildTask struct {
	ModelID  string
	Vector   []float32
	Metadata *ModelMetadata
	Priority int
}

// SearchCache caches search results
type SearchCache struct {
	cache      map[string]*CachedSearchResult
	cacheMutex sync.RWMutex
	maxSize    int
	ttl        time.Duration
}

// CachedSearchResult represents a cached search result
type CachedSearchResult struct {
	Response    *SearchResponse
	CreatedAt   time.Time
	AccessCount int64
}

// NewSemanticSearchEngine creates a new semantic search engine
func NewSemanticSearchEngine(config *SearchConfig) (*SemanticSearchEngine, error) {
	ctx, cancel := context.WithCancel(context.Background())

	embeddingModel, err := NewEmbeddingModel(config.EmbeddingConfig)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create embedding model: %w", err)
	}

	vectorStore := NewVectorStore(embeddingModel.GetDimensions(), config.IndexType)
	indexManager := NewIndexManager(config.IndexType, embeddingModel.GetDimensions())
	searchCache := NewSearchCache(config.CacheSize, config.CacheTTL)

	engine := &SemanticSearchEngine{
		vectorStore:    vectorStore,
		embeddingModel: embeddingModel,
		indexManager:   indexManager,
		searchCache:    searchCache,
		ctx:            ctx,
		cancel:         cancel,
	}

	// Start background tasks
	go engine.indexMaintenanceLoop()
	go engine.cacheCleanupLoop()

	return engine, nil
}

// IndexModel adds a model to the search index
func (sse *SemanticSearchEngine) IndexModel(modelID string, metadata *ModelMetadata) error {
	// Generate embedding from model metadata
	text := sse.buildSearchText(metadata)
	vector, err := sse.embeddingModel.Encode(text)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Create model vector
	modelVector := &ModelVector{
		ModelID:   modelID,
		Vector:    vector,
		Metadata:  metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   "1.0",
	}

	// Add to vector store
	return sse.vectorStore.AddVector(modelID, modelVector)
}

// Search performs semantic search for models
func (sse *SemanticSearchEngine) Search(ctx context.Context, query *SearchQuery) (*SearchResponse, error) {
	startTime := time.Now()

	// Check cache first
	cacheKey := sse.buildCacheKey(query)
	if cached := sse.searchCache.Get(cacheKey); cached != nil {
		cached.Response.QueryTime = time.Since(startTime)
		return cached.Response, nil
	}

	// Generate query vector if not provided
	var queryVector []float32
	if len(query.Vector) == 0 {
		var err error
		queryVector, err = sse.embeddingModel.Encode(query.Text)
		if err != nil {
			return nil, fmt.Errorf("failed to encode query: %w", err)
		}
	} else {
		queryVector = query.Vector
	}

	// Perform vector similarity search
	indexResults, err := sse.vectorStore.Search(queryVector, query.TopK, query.Threshold)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// Apply filters and build results
	results := make([]*SearchResult, 0, len(indexResults))
	for i, indexResult := range indexResults {
		modelVector := sse.vectorStore.GetVector(indexResult.ID)
		if modelVector == nil {
			continue
		}

		// Apply filters
		if !sse.applyFilters(modelVector.Metadata, query.Filters) {
			continue
		}

		result := &SearchResult{
			ModelID: indexResult.ID,
			Score:   indexResult.Score,
			Rank:    i + 1,
		}

		if query.IncludeMetadata {
			result.Metadata = modelVector.Metadata
		}

		if query.Rerank {
			result.Score = sse.rerankResult(query, modelVector)
		}

		results = append(results, result)
	}

	// Sort by score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Limit results
	if len(results) > query.TopK {
		results = results[:query.TopK]
	}

	response := &SearchResponse{
		Results:     results,
		TotalCount:  len(results),
		QueryTime:   time.Since(startTime),
		QueryVector: queryVector,
		Filters:     query.Filters,
	}

	// Cache the result
	sse.searchCache.Put(cacheKey, response)

	return response, nil
}

// UpdateModel updates a model's embedding in the index
func (sse *SemanticSearchEngine) UpdateModel(modelID string, metadata *ModelMetadata) error {
	text := sse.buildSearchText(metadata)
	vector, err := sse.embeddingModel.Encode(text)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	modelVector := &ModelVector{
		ModelID:   modelID,
		Vector:    vector,
		Metadata:  metadata,
		UpdatedAt: time.Now(),
		Version:   "1.1",
	}

	return sse.vectorStore.UpdateVector(modelID, modelVector)
}

// RemoveModel removes a model from the search index
func (sse *SemanticSearchEngine) RemoveModel(modelID string) error {
	return sse.vectorStore.RemoveVector(modelID)
}

// buildSearchText creates searchable text from model metadata
func (sse *SemanticSearchEngine) buildSearchText(metadata *ModelMetadata) string {
	text := metadata.Name + " " + metadata.Description

	for _, tag := range metadata.Tags {
		text += " " + tag
	}

	for _, capability := range metadata.Capabilities {
		text += " " + capability
	}

	text += " " + metadata.Architecture

	for _, lang := range metadata.Language {
		text += " " + lang
	}

	for _, domain := range metadata.Domain {
		text += " " + domain
	}

	return text
}

// applyFilters applies search filters to model metadata
func (sse *SemanticSearchEngine) applyFilters(metadata *ModelMetadata, filters map[string]interface{}) bool {
	for key, value := range filters {
		switch key {
		case "architecture":
			if metadata.Architecture != value.(string) {
				return false
			}
		case "min_parameters":
			if metadata.Parameters < value.(int64) {
				return false
			}
		case "max_parameters":
			if metadata.Parameters > value.(int64) {
				return false
			}
		case "language":
			found := false
			targetLang := value.(string)
			for _, lang := range metadata.Language {
				if lang == targetLang {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		case "domain":
			found := false
			targetDomain := value.(string)
			for _, domain := range metadata.Domain {
				if domain == targetDomain {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	return true
}

// rerankResult applies reranking to improve result relevance
func (sse *SemanticSearchEngine) rerankResult(query *SearchQuery, modelVector *ModelVector) float64 {
	// Simple reranking based on metadata relevance
	score := cosineSimilarity(query.Vector, modelVector.Vector)

	// Boost score based on metadata matches
	if modelVector.Metadata.Performance != nil {
		// Boost high-performance models
		score += modelVector.Metadata.Performance.Accuracy * 0.1
	}

	// Boost recently updated models
	daysSinceUpdate := time.Since(modelVector.UpdatedAt).Hours() / 24
	if daysSinceUpdate < 30 {
		score += 0.05
	}

	return score
}

// buildCacheKey creates a cache key for the search query
func (sse *SemanticSearchEngine) buildCacheKey(query *SearchQuery) string {
	return fmt.Sprintf("%s_%d_%.2f", query.Text, query.TopK, query.Threshold)
}

// indexMaintenanceLoop performs periodic index maintenance
func (sse *SemanticSearchEngine) indexMaintenanceLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-sse.ctx.Done():
			return
		case <-ticker.C:
			sse.indexManager.OptimizeIndex()
		}
	}
}

// cacheCleanupLoop performs periodic cache cleanup
func (sse *SemanticSearchEngine) cacheCleanupLoop() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-sse.ctx.Done():
			return
		case <-ticker.C:
			sse.searchCache.Cleanup()
		}
	}
}

// Helper functions and types
type SearchConfig struct {
	EmbeddingConfig *EmbeddingConfig
	IndexType       string
	CacheSize       int
	CacheTTL        time.Duration
}

type EmbeddingConfig struct {
	ModelPath  string
	ModelType  string
	Dimensions int
	BatchSize  int
}

// Utility functions
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// Factory functions
func NewEmbeddingModel(config *EmbeddingConfig) (EmbeddingModel, error) {
	switch config.ModelType {
	case "tfidf":
		model := NewTFIDFEmbeddingModel(config.Dimensions)
		// Train on some default corpus if needed
		return model, nil
	case "simple_hash":
		return NewSimpleHashEmbeddingModel(config.Dimensions), nil
	default:
		// Default to simple hash model
		return NewSimpleHashEmbeddingModel(config.Dimensions), nil
	}
}

func NewVectorStore(dimensions int, indexType string) *VectorStore {
	return &VectorStore{
		vectors:    make(map[string]*ModelVector),
		dimensions: dimensions,
		index:      NewHNSWIndex(dimensions),
	}
}

func NewIndexManager(indexType string, dimensions int) *IndexManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &IndexManager{
		index:      NewHNSWIndex(dimensions),
		indexType:  indexType,
		dimensions: dimensions,
		buildQueue: make(chan *IndexBuildTask, 1000),
		ctx:        ctx,
		cancel:     cancel,
	}
}

func NewSearchCache(maxSize int, ttl time.Duration) *SearchCache {
	return &SearchCache{
		cache:   make(map[string]*CachedSearchResult),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

func NewHNSWIndex(dimensions int) VectorIndex {
	return &HNSWIndex{dimensions: dimensions}
}

// Placeholder implementations
type SentenceTransformerModel struct {
	dimensions int
	modelName  string
}

func (st *SentenceTransformerModel) Encode(text string) ([]float32, error) {
	// TODO: Implement actual sentence transformer encoding
	return make([]float32, st.dimensions), nil
}

func (st *SentenceTransformerModel) EncodeBatch(texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))
	for i := range texts {
		results[i] = make([]float32, st.dimensions)
	}
	return results, nil
}

func (st *SentenceTransformerModel) GetDimensions() int   { return st.dimensions }
func (st *SentenceTransformerModel) GetModelName() string { return st.modelName }

// Simple placeholder implementation - real HNSW is in vector_index.go
type HNSWIndex struct {
	dimensions int
	vectors    map[string][]float32
}

func (h *HNSWIndex) Add(id string, vector []float32) error {
	if h.vectors == nil {
		h.vectors = make(map[string][]float32)
	}
	h.vectors[id] = vector
	return nil
}

func (h *HNSWIndex) Search(query []float32, topK int, threshold float64) ([]*IndexResult, error) {
	var results []*IndexResult
	for id, vector := range h.vectors {
		score := cosineSimilarity(query, vector)
		if score >= threshold {
			results = append(results, &IndexResult{ID: id, Score: score})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > topK {
		results = results[:topK]
	}

	return results, nil
}

func (h *HNSWIndex) Remove(id string) error {
	delete(h.vectors, id)
	return nil
}

func (h *HNSWIndex) Update(id string, vector []float32) error {
	h.vectors[id] = vector
	return nil
}

func (h *HNSWIndex) GetSize() int    { return len(h.vectors) }
func (h *HNSWIndex) Optimize() error { return nil }

// VectorStore methods
func (vs *VectorStore) AddVector(id string, vector *ModelVector) error {
	vs.vectorsMutex.Lock()
	defer vs.vectorsMutex.Unlock()

	vs.vectors[id] = vector
	return vs.index.Add(id, vector.Vector)
}

func (vs *VectorStore) UpdateVector(id string, vector *ModelVector) error {
	vs.vectorsMutex.Lock()
	defer vs.vectorsMutex.Unlock()

	vs.vectors[id] = vector
	return vs.index.Update(id, vector.Vector)
}

func (vs *VectorStore) RemoveVector(id string) error {
	vs.vectorsMutex.Lock()
	defer vs.vectorsMutex.Unlock()

	delete(vs.vectors, id)
	return vs.index.Remove(id)
}

func (vs *VectorStore) GetVector(id string) *ModelVector {
	vs.vectorsMutex.RLock()
	defer vs.vectorsMutex.RUnlock()

	return vs.vectors[id]
}

func (vs *VectorStore) Search(query []float32, topK int, threshold float64) ([]*IndexResult, error) {
	return vs.index.Search(query, topK, threshold)
}

func (vs *VectorStore) GetSize() int {
	vs.vectorsMutex.RLock()
	defer vs.vectorsMutex.RUnlock()
	return len(vs.vectors)
}

// IndexManager methods
func (im *IndexManager) OptimizeIndex() error {
	return im.index.Optimize()
}

// SearchCache methods
func (sc *SearchCache) Get(key string) *CachedSearchResult {
	sc.cacheMutex.RLock()
	defer sc.cacheMutex.RUnlock()

	cached, exists := sc.cache[key]
	if !exists {
		return nil
	}

	if time.Since(cached.CreatedAt) > sc.ttl {
		delete(sc.cache, key)
		return nil
	}

	cached.AccessCount++
	return cached
}

func (sc *SearchCache) Put(key string, response *SearchResponse) {
	sc.cacheMutex.Lock()
	defer sc.cacheMutex.Unlock()

	if len(sc.cache) >= sc.maxSize {
		// Simple LRU eviction
		var oldestKey string
		var oldestTime time.Time
		for k, v := range sc.cache {
			if oldestKey == "" || v.CreatedAt.Before(oldestTime) {
				oldestKey = k
				oldestTime = v.CreatedAt
			}
		}
		delete(sc.cache, oldestKey)
	}

	sc.cache[key] = &CachedSearchResult{
		Response:    response,
		CreatedAt:   time.Now(),
		AccessCount: 1,
	}
}

func (sc *SearchCache) Cleanup() {
	sc.cacheMutex.Lock()
	defer sc.cacheMutex.Unlock()

	for key, cached := range sc.cache {
		if time.Since(cached.CreatedAt) > sc.ttl {
			delete(sc.cache, key)
		}
	}
}

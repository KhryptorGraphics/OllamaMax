package search

import (
	"hash/fnv"
	"math"
	"strings"
	"sync"
)

// TFIDFEmbeddingModel implements a TF-IDF based embedding model
type TFIDFEmbeddingModel struct {
	dimensions int
	modelName  string
	vocabulary map[string]int
	idf        map[string]float64
	docFreq    map[string]int
	totalDocs  int
	mutex      sync.RWMutex
}

// NewTFIDFEmbeddingModel creates a new TF-IDF embedding model
func NewTFIDFEmbeddingModel(dimensions int) *TFIDFEmbeddingModel {
	return &TFIDFEmbeddingModel{
		dimensions: dimensions,
		modelName:  "tfidf",
		vocabulary: make(map[string]int),
		idf:        make(map[string]float64),
		docFreq:    make(map[string]int),
		totalDocs:  0,
	}
}

// Encode encodes text into a vector representation
func (tfidf *TFIDFEmbeddingModel) Encode(text string) ([]float32, error) {
	tfidf.mutex.RLock()
	defer tfidf.mutex.RUnlock()

	// Tokenize text
	tokens := tfidf.tokenize(text)

	// Calculate term frequencies
	tf := make(map[string]float64)
	for _, token := range tokens {
		tf[token]++
	}

	// Normalize term frequencies
	for token := range tf {
		tf[token] = tf[token] / float64(len(tokens))
	}

	// Create embedding vector
	embedding := make([]float32, tfidf.dimensions)

	for token, tfValue := range tf {
		// Get IDF value
		idfValue := tfidf.idf[token]
		if idfValue == 0 {
			idfValue = 1.0 // Default for unknown terms
		}

		// Calculate TF-IDF score
		tfidfScore := tfValue * idfValue

		// Map to vector dimensions using hash
		indices := tfidf.getTokenIndices(token)
		for _, idx := range indices {
			embedding[idx] += float32(tfidfScore)
		}
	}

	// Normalize the vector
	tfidf.normalizeVector(embedding)

	return embedding, nil
}

// EncodeBatch encodes multiple texts
func (tfidf *TFIDFEmbeddingModel) EncodeBatch(texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))

	for i, text := range texts {
		embedding, err := tfidf.Encode(text)
		if err != nil {
			return nil, err
		}
		results[i] = embedding
	}

	return results, nil
}

// GetDimensions returns the embedding dimensions
func (tfidf *TFIDFEmbeddingModel) GetDimensions() int {
	return tfidf.dimensions
}

// GetModelName returns the model name
func (tfidf *TFIDFEmbeddingModel) GetModelName() string {
	return tfidf.modelName
}

// TrainOnCorpus trains the model on a corpus of documents
func (tfidf *TFIDFEmbeddingModel) TrainOnCorpus(documents []string) error {
	tfidf.mutex.Lock()
	defer tfidf.mutex.Unlock()

	// Reset training data
	tfidf.vocabulary = make(map[string]int)
	tfidf.docFreq = make(map[string]int)
	tfidf.totalDocs = len(documents)

	vocabIndex := 0

	// First pass: build vocabulary and document frequencies
	for _, doc := range documents {
		tokens := tfidf.tokenize(doc)
		uniqueTokens := make(map[string]bool)

		for _, token := range tokens {
			// Add to vocabulary
			if _, exists := tfidf.vocabulary[token]; !exists {
				tfidf.vocabulary[token] = vocabIndex
				vocabIndex++
			}

			// Count document frequency (only once per document)
			if !uniqueTokens[token] {
				tfidf.docFreq[token]++
				uniqueTokens[token] = true
			}
		}
	}

	// Calculate IDF values
	for token, df := range tfidf.docFreq {
		tfidf.idf[token] = math.Log(float64(tfidf.totalDocs) / float64(df))
	}

	return nil
}

// tokenize splits text into tokens
func (tfidf *TFIDFEmbeddingModel) tokenize(text string) []string {
	// Simple tokenization: lowercase, split by spaces, remove punctuation
	text = strings.ToLower(text)
	text = strings.ReplaceAll(text, ".", " ")
	text = strings.ReplaceAll(text, ",", " ")
	text = strings.ReplaceAll(text, "!", " ")
	text = strings.ReplaceAll(text, "?", " ")
	text = strings.ReplaceAll(text, ";", " ")
	text = strings.ReplaceAll(text, ":", " ")
	text = strings.ReplaceAll(text, "(", " ")
	text = strings.ReplaceAll(text, ")", " ")
	text = strings.ReplaceAll(text, "[", " ")
	text = strings.ReplaceAll(text, "]", " ")
	text = strings.ReplaceAll(text, "{", " ")
	text = strings.ReplaceAll(text, "}", " ")
	text = strings.ReplaceAll(text, "-", " ")
	text = strings.ReplaceAll(text, "_", " ")

	tokens := strings.Fields(text)

	// Filter out very short tokens
	filtered := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if len(token) > 2 {
			filtered = append(filtered, token)
		}
	}

	return filtered
}

// getTokenIndices maps a token to vector indices using hashing
func (tfidf *TFIDFEmbeddingModel) getTokenIndices(token string) []int {
	// Use multiple hash functions to get multiple indices
	indices := make([]int, 3) // Use 3 hash functions

	h1 := fnv.New32a()
	h1.Write([]byte(token))
	indices[0] = int(h1.Sum32()) % tfidf.dimensions

	h2 := fnv.New32a()
	h2.Write([]byte(token + "_1"))
	indices[1] = int(h2.Sum32()) % tfidf.dimensions

	h3 := fnv.New32a()
	h3.Write([]byte(token + "_2"))
	indices[2] = int(h3.Sum32()) % tfidf.dimensions

	return indices
}

// normalizeVector normalizes a vector to unit length
func (tfidf *TFIDFEmbeddingModel) normalizeVector(vector []float32) {
	var norm float32
	for _, val := range vector {
		norm += val * val
	}
	norm = float32(math.Sqrt(float64(norm)))

	if norm > 0 {
		for i := range vector {
			vector[i] /= norm
		}
	}
}

// SimpleHashEmbeddingModel implements a simple hash-based embedding
type SimpleHashEmbeddingModel struct {
	dimensions int
	modelName  string
}

// NewSimpleHashEmbeddingModel creates a new simple hash embedding model
func NewSimpleHashEmbeddingModel(dimensions int) *SimpleHashEmbeddingModel {
	return &SimpleHashEmbeddingModel{
		dimensions: dimensions,
		modelName:  "simple_hash",
	}
}

// Encode encodes text using simple hashing
func (she *SimpleHashEmbeddingModel) Encode(text string) ([]float32, error) {
	embedding := make([]float32, she.dimensions)

	// Tokenize text
	tokens := she.tokenize(text)

	// Hash each token to multiple dimensions
	for _, token := range tokens {
		indices := she.getTokenIndices(token)
		for _, idx := range indices {
			embedding[idx] += 1.0
		}
	}

	// Normalize
	she.normalizeVector(embedding)

	return embedding, nil
}

// EncodeBatch encodes multiple texts
func (she *SimpleHashEmbeddingModel) EncodeBatch(texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))

	for i, text := range texts {
		embedding, err := she.Encode(text)
		if err != nil {
			return nil, err
		}
		results[i] = embedding
	}

	return results, nil
}

// GetDimensions returns the embedding dimensions
func (she *SimpleHashEmbeddingModel) GetDimensions() int {
	return she.dimensions
}

// GetModelName returns the model name
func (she *SimpleHashEmbeddingModel) GetModelName() string {
	return she.modelName
}

// tokenize splits text into tokens
func (she *SimpleHashEmbeddingModel) tokenize(text string) []string {
	text = strings.ToLower(text)
	return strings.Fields(text)
}

// getTokenIndices maps a token to vector indices
func (she *SimpleHashEmbeddingModel) getTokenIndices(token string) []int {
	indices := make([]int, 2)

	h1 := fnv.New32a()
	h1.Write([]byte(token))
	indices[0] = int(h1.Sum32()) % she.dimensions

	h2 := fnv.New32a()
	h2.Write([]byte(token + "_hash"))
	indices[1] = int(h2.Sum32()) % she.dimensions

	return indices
}

// normalizeVector normalizes a vector to unit length
func (she *SimpleHashEmbeddingModel) normalizeVector(vector []float32) {
	var norm float32
	for _, val := range vector {
		norm += val * val
	}
	norm = float32(math.Sqrt(float64(norm)))

	if norm > 0 {
		for i := range vector {
			vector[i] /= norm
		}
	}
}

// ComputeSimilarity computes cosine similarity between two embeddings
func ComputeCosineSimilarity(a, b []float32) float64 {
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

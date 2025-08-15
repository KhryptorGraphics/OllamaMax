package storage

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Search performs advanced metadata search
func (mm *MetadataManager) Search(ctx context.Context, query *MetadataQuery) (*MetadataQueryResult, error) {
	start := time.Now()
	defer func() {
		mm.updateLatencyStats("search", time.Since(start))
		mm.incrementOperationCount("search")
	}()

	if !mm.config.EnableSearch {
		return nil, NewStorageError(ErrCodeIndexError, "search is disabled", "")
	}

	// Select best index for query
	indexName := mm.selectBestIndex(query)

	var results []*ObjectMetadata
	var err error

	if indexName != "" {
		results, err = mm.searchWithIndex(ctx, query, indexName)
	} else {
		results, err = mm.searchWithoutIndex(ctx, query)
	}

	if err != nil {
		return nil, err
	}

	// Apply sorting and pagination
	results = mm.applySortingAndPagination(results, query)

	return &MetadataQueryResult{
		Objects:     results,
		Total:       int64(len(results)),
		QueryTime:   time.Since(start),
		IndexUsed:   indexName,
		Explanation: mm.explainQuery(query, indexName),
	}, nil
}

// CreateIndex creates a new metadata index
func (mm *MetadataManager) CreateIndex(ctx context.Context, name string, fields []string, indexType string) error {
	start := time.Now()
	defer func() {
		mm.updateLatencyStats("index", time.Since(start))
		mm.incrementOperationCount("create_index")
	}()

	mm.indexMutex.Lock()
	defer mm.indexMutex.Unlock()

	if _, exists := mm.indexes[name]; exists {
		return NewStorageError(ErrCodeAlreadyExists, "index already exists", name)
	}

	index := &MetadataIndex{
		Name:      name,
		Type:      indexType,
		Fields:    fields,
		Values:    make(map[string][]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Stats:     NewIndexStats(),
		mutex:     &sync.RWMutex{},
	}

	mm.indexes[name] = index

	// Build index in background
	go mm.buildIndex(ctx, index)

	return nil
}

// DropIndex drops a metadata index
func (mm *MetadataManager) DropIndex(ctx context.Context, name string) error {
	mm.indexMutex.Lock()
	defer mm.indexMutex.Unlock()

	if _, exists := mm.indexes[name]; !exists {
		return NewStorageError(ErrCodeNotFound, "index not found", name)
	}

	delete(mm.indexes, name)

	return nil
}

// GetIndexes returns all metadata indexes
func (mm *MetadataManager) GetIndexes(ctx context.Context) ([]*MetadataIndex, error) {
	mm.indexMutex.RLock()
	defer mm.indexMutex.RUnlock()

	indexes := make([]*MetadataIndex, 0, len(mm.indexes))
	for _, index := range mm.indexes {
		// Create a copy to avoid race conditions
		indexCopy := *index
		indexes = append(indexes, &indexCopy)
	}

	return indexes, nil
}

// Index management

func (mm *MetadataManager) createDefaultIndexes() {
	// Create indexes for common fields
	commonIndexes := map[string][]string{
		"size_index":    {"size"},
		"type_index":    {"content_type"},
		"created_index": {"created_at"},
		"status_index":  {"status"},
	}

	for name, fields := range commonIndexes {
		if err := mm.CreateIndex(context.Background(), name, fields, "btree"); err != nil {
			mm.logger.Warn("failed to create default index", "name", name, "error", err)
		}
	}
}

func (mm *MetadataManager) updateIndexes(key string, metadata *ObjectMetadata) {
	mm.indexMutex.Lock()
	defer mm.indexMutex.Unlock()

	for _, index := range mm.indexes {
		mm.updateIndex(index, key, metadata)
	}
}

func (mm *MetadataManager) updateIndex(index *MetadataIndex, key string, metadata *ObjectMetadata) {
	for _, field := range index.Fields {
		value := mm.extractFieldValue(metadata, field)
		if value != "" {
			if keys, exists := index.Values[value]; exists {
				// Check if key already exists
				found := false
				for _, existingKey := range keys {
					if existingKey == key {
						found = true
						break
					}
				}
				if !found {
					index.Values[value] = append(keys, key)
				}
			} else {
				index.Values[value] = []string{key}
			}
		}
	}

	index.UpdatedAt = time.Now()
	index.Stats.UpdateCount++
}

func (mm *MetadataManager) removeFromIndexes(key string, metadata *ObjectMetadata) {
	if metadata == nil {
		return
	}

	mm.indexMutex.Lock()
	defer mm.indexMutex.Unlock()

	for _, index := range mm.indexes {
		mm.removeFromIndex(index, key, metadata)
	}
}

func (mm *MetadataManager) removeFromIndex(index *MetadataIndex, key string, metadata *ObjectMetadata) {
	for _, field := range index.Fields {
		value := mm.extractFieldValue(metadata, field)
		if value != "" {
			if keys, exists := index.Values[value]; exists {
				// Remove key from slice
				for i, existingKey := range keys {
					if existingKey == key {
						index.Values[value] = append(keys[:i], keys[i+1:]...)
						break
					}
				}
				// Remove empty value entries
				if len(index.Values[value]) == 0 {
					delete(index.Values, value)
				}
			}
		}
	}

	index.UpdatedAt = time.Now()
}

func (mm *MetadataManager) buildIndex(ctx context.Context, index *MetadataIndex) {
	mm.logger.Info("building metadata index", "name", index.Name)

	// Get all metadata from backend
	allMetadata, err := mm.listFromBackend("", nil)
	if err != nil {
		mm.logger.Error("failed to list metadata for index building", "error", err)
		return
	}

	// Build index
	for _, metadata := range allMetadata {
		mm.updateIndex(index, metadata.Key, metadata)
	}

	index.Stats.TotalEntries = int64(len(allMetadata))
	index.Stats.UniqueValues = int64(len(index.Values))

	mm.logger.Info("metadata index built", "name", index.Name, "entries", index.Stats.TotalEntries)
}

// Search implementation

func (mm *MetadataManager) selectBestIndex(query *MetadataQuery) string {
	mm.indexMutex.RLock()
	defer mm.indexMutex.RUnlock()

	var bestIndex string
	var bestScore int

	for name, index := range mm.indexes {
		score := mm.calculateIndexScore(index, query)
		if score > bestScore {
			bestScore = score
			bestIndex = name
		}
	}

	return bestIndex
}

func (mm *MetadataManager) calculateIndexScore(index *MetadataIndex, query *MetadataQuery) int {
	score := 0

	// Check if index fields match query conditions
	for _, condition := range query.Conditions {
		for _, field := range index.Fields {
			if condition.Field == field {
				score += 10
				// Bonus for exact match operators
				if condition.Operator == "eq" {
					score += 5
				}
			}
		}
	}

	// Bonus for newer indexes (they might be more optimized)
	age := time.Since(index.CreatedAt)
	if age < 24*time.Hour {
		score += 2
	}

	// Penalty for large indexes (might be slower)
	if index.Stats.TotalEntries > 10000 {
		score -= 1
	}

	return score
}

func (mm *MetadataManager) searchWithIndex(ctx context.Context, query *MetadataQuery, indexName string) ([]*ObjectMetadata, error) {
	mm.indexMutex.RLock()
	index, exists := mm.indexes[indexName]
	mm.indexMutex.RUnlock()

	if !exists {
		return mm.searchWithoutIndex(ctx, query)
	}

	var candidateKeys []string

	// Get candidate keys from index
	for _, condition := range query.Conditions {
		if mm.isFieldIndexed(index, condition.Field) {
			keys := mm.getKeysFromIndex(index, condition)
			if candidateKeys == nil {
				candidateKeys = keys
			} else {
				candidateKeys = mm.intersectKeys(candidateKeys, keys)
			}
		}
	}

	// Load metadata for candidate keys
	var results []*ObjectMetadata
	for _, key := range candidateKeys {
		metadata, err := mm.Get(ctx, key)
		if err != nil {
			continue
		}

		if mm.matchesQuery(metadata, query) {
			results = append(results, metadata)
		}
	}

	// Update index statistics
	index.Stats.QueryCount++

	return results, nil
}

func (mm *MetadataManager) searchWithoutIndex(ctx context.Context, query *MetadataQuery) ([]*ObjectMetadata, error) {
	// Full scan
	allMetadata, err := mm.listFromBackend("", nil)
	if err != nil {
		return nil, err
	}

	var results []*ObjectMetadata
	for _, metadata := range allMetadata {
		if mm.matchesQuery(metadata, query) {
			results = append(results, metadata)
		}
	}

	return results, nil
}

func (mm *MetadataManager) applySortingAndPagination(results []*ObjectMetadata, query *MetadataQuery) []*ObjectMetadata {
	// Apply sorting
	if query.Sort != nil {
		sort.Slice(results, func(i, j int) bool {
			return mm.compareMetadata(results[i], results[j], query.Sort)
		})
	}

	// Apply pagination
	if query.Limit > 0 || query.Offset > 0 {
		start := query.Offset
		if start > len(results) {
			start = len(results)
		}

		end := len(results)
		if query.Limit > 0 && start+query.Limit < end {
			end = start + query.Limit
		}

		results = results[start:end]
	}

	return results
}

// Helper methods

func (mm *MetadataManager) extractFieldValue(metadata *ObjectMetadata, field string) string {
	switch field {
	case "key":
		return metadata.Key
	case "size":
		return fmt.Sprintf("%d", metadata.Size)
	case "content_type":
		return metadata.ContentType
	case "hash":
		return metadata.Hash
	case "created_at":
		return metadata.CreatedAt.Format(time.RFC3339)
	case "updated_at":
		return metadata.UpdatedAt.Format(time.RFC3339)
	case "accessed_at":
		return metadata.AccessedAt.Format(time.RFC3339)
	case "version":
		return metadata.Version
	default:
		// Check attributes
		if strings.HasPrefix(field, "attributes.") {
			attrKey := strings.TrimPrefix(field, "attributes.")
			if value, exists := metadata.Attributes[attrKey]; exists {
				if strValue, ok := value.(string); ok {
					return strValue
				}
				return fmt.Sprintf("%v", value)
			}
		}
		return ""
	}
}

func (mm *MetadataManager) isFieldIndexed(index *MetadataIndex, field string) bool {
	for _, indexField := range index.Fields {
		if indexField == field {
			return true
		}
	}
	return false
}

func (mm *MetadataManager) getKeysFromIndex(index *MetadataIndex, condition *QueryCondition) []string {
	value := fmt.Sprintf("%v", condition.Value)

	switch condition.Operator {
	case "eq":
		if keys, exists := index.Values[value]; exists {
			return keys
		}
		return []string{}
	case "ne":
		var allKeys []string
		for indexValue, keys := range index.Values {
			if indexValue != value {
				allKeys = append(allKeys, keys...)
			}
		}
		return allKeys
	default:
		// For other operators, return all keys (will be filtered later)
		var allKeys []string
		for _, keys := range index.Values {
			allKeys = append(allKeys, keys...)
		}
		return allKeys
	}
}

func (mm *MetadataManager) intersectKeys(keys1, keys2 []string) []string {
	keyMap := make(map[string]bool)
	for _, key := range keys1 {
		keyMap[key] = true
	}

	var result []string
	for _, key := range keys2 {
		if keyMap[key] {
			result = append(result, key)
		}
	}

	return result
}

func (mm *MetadataManager) matchesQuery(metadata *ObjectMetadata, query *MetadataQuery) bool {
	for _, condition := range query.Conditions {
		if !mm.matchesCondition(metadata, condition) {
			return false
		}
	}

	// Check full-text search if specified
	if query.FullText != "" {
		searchText := strings.ToLower(query.FullText)
		if !strings.Contains(strings.ToLower(metadata.Key), searchText) &&
			!strings.Contains(strings.ToLower(metadata.ContentType), searchText) {
			return false
		}
	}

	return true
}

func (mm *MetadataManager) matchesCondition(metadata *ObjectMetadata, condition *QueryCondition) bool {
	fieldValue := mm.extractFieldValue(metadata, condition.Field)
	conditionValue := fmt.Sprintf("%v", condition.Value)

	switch condition.Operator {
	case "eq":
		return fieldValue == conditionValue
	case "ne":
		return fieldValue != conditionValue
	case "like":
		return strings.Contains(strings.ToLower(fieldValue), strings.ToLower(conditionValue))
	case "gt", "gte", "lt", "lte":
		return mm.compareValues(fieldValue, conditionValue, condition.Operator)
	default:
		return false
	}
}

func (mm *MetadataManager) compareValues(value1, value2, operator string) bool {
	// Simplified comparison (would need proper type handling)
	switch operator {
	case "gt":
		return value1 > value2
	case "gte":
		return value1 >= value2
	case "lt":
		return value1 < value2
	case "lte":
		return value1 <= value2
	default:
		return false
	}
}

func (mm *MetadataManager) compareMetadata(m1, m2 *ObjectMetadata, sort *SortOptions) bool {
	value1 := mm.extractFieldValue(m1, sort.Field)
	value2 := mm.extractFieldValue(m2, sort.Field)

	if sort.Order == "desc" {
		return value1 > value2
	}
	return value1 < value2
}

func (mm *MetadataManager) explainQuery(query *MetadataQuery, indexUsed string) string {
	if indexUsed != "" {
		return fmt.Sprintf("Used index: %s", indexUsed)
	}
	return "Full table scan"
}

package storage

import (
	"fmt"
	"reflect"
	"sync"
	"time"
)

// Fixed version of metadata manager with proper concurrency control
func (mm *MetadataManager) updateIndexSafe(indexName string, key string, metadata *ObjectMetadata) {
	mm.indexMutex.Lock()
	defer mm.indexMutex.Unlock()

	index, exists := mm.indexes[indexName]
	if !exists {
		return
	}

	// Get index field value
	value := mm.getIndexValue(metadata, index.Fields)
	if value == nil {
		return
	}

	valueStr := mm.valueToString(value)

	// Update index with proper synchronization
	index.mutex.Lock()
	defer index.mutex.Unlock()

	if keys, found := index.Values[valueStr]; found {
		// Add key if not already present
		keyExists := false
		for _, existingKey := range keys {
			if existingKey == key {
				keyExists = true
				break
			}
		}
		if !keyExists {
			index.Values[valueStr] = append(keys, key)
		}
	} else {
		index.Values[valueStr] = []string{key}
	}

	index.UpdatedAt = time.Now()
	index.Stats.UpdateCount++
}

// Thread-safe version of updateIndexes
func (mm *MetadataManager) updateIndexesSafe(key string, metadata *ObjectMetadata) {
	mm.indexMutex.RLock()
	indexNames := make([]string, 0, len(mm.indexes))
	for name := range mm.indexes {
		indexNames = append(indexNames, name)
	}
	mm.indexMutex.RUnlock()

	// Update each index separately to avoid holding multiple locks
	for _, indexName := range indexNames {
		mm.updateIndexSafe(indexName, key, metadata)
	}
}

// Enhanced index with mutex for thread safety
type ThreadSafeIndex struct {
	*MetadataIndex
	mutex sync.RWMutex
}

// Create thread-safe wrapper for existing indexes
func (mm *MetadataManager) makeIndexesSafe() {
	mm.indexMutex.Lock()
	defer mm.indexMutex.Unlock()

	// This would be called during initialization to wrap indexes
	// with thread-safe mutexes if needed
}

// getIndexValue extracts the value for indexing from metadata based on field names
func (mm *MetadataManager) getIndexValue(metadata *ObjectMetadata, fields []string) interface{} {
	if len(fields) == 0 {
		return nil
	}

	// Use reflection to get field values from metadata
	v := reflect.ValueOf(metadata).Elem()
	t := v.Type()

	for _, fieldName := range fields {
		// Find field by name (case insensitive)
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.Name == fieldName {
				fieldValue := v.Field(i)
				if fieldValue.IsValid() && fieldValue.CanInterface() {
					return fieldValue.Interface()
				}
			}
		}
	}

	return nil
}

// valueToString converts an interface{} value to a string for indexing
func (mm *MetadataManager) valueToString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int, int32, int64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%.6f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case time.Time:
		return v.Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", v)
	}
}

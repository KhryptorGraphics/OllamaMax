package routing

import (
	"strings"
	"sync"
	"time"
)

// ContentIndex manages content indexing
type ContentIndex struct {
	index map[string]*IndexEntry
	mu    sync.RWMutex
}

// IndexEntry represents an index entry
type IndexEntry struct {
	ContentID string
	Metadata  *ContentMetadata
	Keywords  []string
	UpdatedAt time.Time
}

// NewContentIndex creates a new content index
func NewContentIndex() *ContentIndex {
	return &ContentIndex{
		index: make(map[string]*IndexEntry),
	}
}

// AddContent adds content to the index
func (ci *ContentIndex) AddContent(metadata *ContentMetadata) {
	ci.mu.Lock()
	defer ci.mu.Unlock()

	keywords := ci.extractKeywords(metadata)
	entry := &IndexEntry{
		ContentID: metadata.ID,
		Metadata:  metadata,
		Keywords:  keywords,
		UpdatedAt: time.Now(),
	}

	ci.index[metadata.ID] = entry
}

// RemoveContent removes content from the index
func (ci *ContentIndex) RemoveContent(contentID string) {
	ci.mu.Lock()
	defer ci.mu.Unlock()

	delete(ci.index, contentID)
}

// SearchContent searches for content by keywords
func (ci *ContentIndex) SearchContent(query string) []*ContentMetadata {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	queryKeywords := ci.tokenize(strings.ToLower(query))
	var results []*ContentMetadata

	for _, entry := range ci.index {
		if ci.matchesQuery(entry.Keywords, queryKeywords) {
			results = append(results, entry.Metadata)
		}
	}

	return results
}

// GetContent retrieves content metadata by ID
func (ci *ContentIndex) GetContent(contentID string) (*ContentMetadata, bool) {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	entry, exists := ci.index[contentID]
	if !exists {
		return nil, false
	}
	return entry.Metadata, true
}

// GetAllContent returns all indexed content
func (ci *ContentIndex) GetAllContent() []*ContentMetadata {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	results := make([]*ContentMetadata, 0, len(ci.index))
	for _, entry := range ci.index {
		results = append(results, entry.Metadata)
	}
	return results
}

// GetContentByType returns content filtered by type
func (ci *ContentIndex) GetContentByType(contentType string) []*ContentMetadata {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	var results []*ContentMetadata
	for _, entry := range ci.index {
		if entry.Metadata.Type == contentType {
			results = append(results, entry.Metadata)
		}
	}
	return results
}

// GetContentByTag returns content filtered by tag
func (ci *ContentIndex) GetContentByTag(tagKey, tagValue string) []*ContentMetadata {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	var results []*ContentMetadata
	for _, entry := range ci.index {
		if value, exists := entry.Metadata.Tags[tagKey]; exists && value == tagValue {
			results = append(results, entry.Metadata)
		}
	}
	return results
}

// GetIndexSize returns the number of indexed items
func (ci *ContentIndex) GetIndexSize() int {
	ci.mu.RLock()
	defer ci.mu.RUnlock()
	return len(ci.index)
}

// UpdateContent updates content in the index
func (ci *ContentIndex) UpdateContent(metadata *ContentMetadata) {
	ci.mu.Lock()
	defer ci.mu.Unlock()

	keywords := ci.extractKeywords(metadata)
	entry := &IndexEntry{
		ContentID: metadata.ID,
		Metadata:  metadata,
		Keywords:  keywords,
		UpdatedAt: time.Now(),
	}

	ci.index[metadata.ID] = entry
}

// extractKeywords extracts keywords from content metadata
func (ci *ContentIndex) extractKeywords(metadata *ContentMetadata) []string {
	var keywords []string

	// Add name and description keywords
	keywords = append(keywords, ci.tokenize(strings.ToLower(metadata.Name))...)
	keywords = append(keywords, ci.tokenize(strings.ToLower(metadata.Description))...)

	// Add type
	if metadata.Type != "" {
		keywords = append(keywords, strings.ToLower(metadata.Type))
	}

	// Add author
	if metadata.Author != "" {
		keywords = append(keywords, strings.ToLower(metadata.Author))
	}

	// Add tags
	for key, value := range metadata.Tags {
		keywords = append(keywords, strings.ToLower(key))
		keywords = append(keywords, strings.ToLower(value))
	}

	// Add labels
	for _, label := range metadata.Labels {
		keywords = append(keywords, strings.ToLower(label))
	}

	// Remove duplicates
	return ci.removeDuplicates(keywords)
}

// tokenize splits text into tokens
func (ci *ContentIndex) tokenize(text string) []string {
	// Simple tokenization - split by spaces and common punctuation
	text = strings.ReplaceAll(text, ",", " ")
	text = strings.ReplaceAll(text, ".", " ")
	text = strings.ReplaceAll(text, ";", " ")
	text = strings.ReplaceAll(text, ":", " ")
	text = strings.ReplaceAll(text, "-", " ")
	text = strings.ReplaceAll(text, "_", " ")

	tokens := strings.Fields(text)
	var result []string

	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if len(token) > 2 { // Ignore very short tokens
			result = append(result, token)
		}
	}

	return result
}

// matchesQuery checks if keywords match the query
func (ci *ContentIndex) matchesQuery(keywords, queryKeywords []string) bool {
	if len(queryKeywords) == 0 {
		return true
	}

	for _, queryKeyword := range queryKeywords {
		found := false
		for _, keyword := range keywords {
			if strings.Contains(keyword, queryKeyword) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// removeDuplicates removes duplicate strings from a slice
func (ci *ContentIndex) removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// GetIndexStatistics returns statistics about the index
func (ci *ContentIndex) GetIndexStatistics() map[string]interface{} {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	stats := map[string]interface{}{
		"total_entries": len(ci.index),
		"content_types": make(map[string]int),
		"total_keywords": 0,
		"average_keywords_per_entry": 0.0,
	}

	contentTypes := make(map[string]int)
	totalKeywords := 0

	for _, entry := range ci.index {
		// Count content types
		contentTypes[entry.Metadata.Type]++

		// Count keywords
		totalKeywords += len(entry.Keywords)
	}

	stats["content_types"] = contentTypes
	stats["total_keywords"] = totalKeywords

	if len(ci.index) > 0 {
		stats["average_keywords_per_entry"] = float64(totalKeywords) / float64(len(ci.index))
	}

	return stats
}

// ClearIndex clears all entries from the index
func (ci *ContentIndex) ClearIndex() {
	ci.mu.Lock()
	defer ci.mu.Unlock()

	ci.index = make(map[string]*IndexEntry)
}

// GetRecentContent returns recently updated content
func (ci *ContentIndex) GetRecentContent(limit int) []*ContentMetadata {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	// Convert to slice for sorting
	entries := make([]*IndexEntry, 0, len(ci.index))
	for _, entry := range ci.index {
		entries = append(entries, entry)
	}

	// Sort by update time (most recent first)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].UpdatedAt.Before(entries[j].UpdatedAt) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Return limited results
	if limit > len(entries) {
		limit = len(entries)
	}

	results := make([]*ContentMetadata, limit)
	for i := 0; i < limit; i++ {
		results[i] = entries[i].Metadata
	}

	return results
}

// GetContentByAuthor returns content filtered by author
func (ci *ContentIndex) GetContentByAuthor(author string) []*ContentMetadata {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	var results []*ContentMetadata
	authorLower := strings.ToLower(author)

	for _, entry := range ci.index {
		if strings.ToLower(entry.Metadata.Author) == authorLower {
			results = append(results, entry.Metadata)
		}
	}
	return results
}

// GetContentBySize returns content filtered by size range
func (ci *ContentIndex) GetContentBySize(minSize, maxSize int64) []*ContentMetadata {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	var results []*ContentMetadata
	for _, entry := range ci.index {
		if entry.Metadata.Size >= minSize && entry.Metadata.Size <= maxSize {
			results = append(results, entry.Metadata)
		}
	}
	return results
}

// GetContentByDateRange returns content filtered by creation date range
func (ci *ContentIndex) GetContentByDateRange(startDate, endDate time.Time) []*ContentMetadata {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	var results []*ContentMetadata
	for _, entry := range ci.index {
		if entry.Metadata.CreatedAt.After(startDate) && entry.Metadata.CreatedAt.Before(endDate) {
			results = append(results, entry.Metadata)
		}
	}
	return results
}

// SearchContentAdvanced performs advanced search with multiple criteria
func (ci *ContentIndex) SearchContentAdvanced(criteria map[string]interface{}) []*ContentMetadata {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	var results []*ContentMetadata

	for _, entry := range ci.index {
		matches := true

		// Check query keywords
		if query, exists := criteria["query"]; exists {
			if queryStr, ok := query.(string); ok {
				queryKeywords := ci.tokenize(strings.ToLower(queryStr))
				if !ci.matchesQuery(entry.Keywords, queryKeywords) {
					matches = false
				}
			}
		}

		// Check content type
		if contentType, exists := criteria["type"]; exists {
			if typeStr, ok := contentType.(string); ok {
				if entry.Metadata.Type != typeStr {
					matches = false
				}
			}
		}

		// Check author
		if author, exists := criteria["author"]; exists {
			if authorStr, ok := author.(string); ok {
				if strings.ToLower(entry.Metadata.Author) != strings.ToLower(authorStr) {
					matches = false
				}
			}
		}

		// Check size range
		if minSize, exists := criteria["min_size"]; exists {
			if minSizeInt, ok := minSize.(int64); ok {
				if entry.Metadata.Size < minSizeInt {
					matches = false
				}
			}
		}

		if maxSize, exists := criteria["max_size"]; exists {
			if maxSizeInt, ok := maxSize.(int64); ok {
				if entry.Metadata.Size > maxSizeInt {
					matches = false
				}
			}
		}

		if matches {
			results = append(results, entry.Metadata)
		}
	}

	return results
}

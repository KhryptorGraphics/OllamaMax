package security

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
)

// SQLInjectionPrevention provides comprehensive SQL injection protection
type SQLInjectionPrevention struct {
	// Dangerous SQL patterns to detect
	dangerousPatterns []*regexp.Regexp

	// Allowed characters for different input types
	allowedPatterns map[string]*regexp.Regexp
}

// NewSQLInjectionPrevention creates a new SQL injection prevention instance
func NewSQLInjectionPrevention() *SQLInjectionPrevention {
	sip := &SQLInjectionPrevention{
		allowedPatterns: make(map[string]*regexp.Regexp),
	}

	// Initialize dangerous patterns
	sip.initializeDangerousPatterns()

	// Initialize allowed patterns
	sip.initializeAllowedPatterns()

	return sip
}

// initializeDangerousPatterns sets up patterns that indicate SQL injection attempts
func (sip *SQLInjectionPrevention) initializeDangerousPatterns() {
	patterns := []string{
		// SQL injection keywords
		`(?i)\b(union|select|insert|update|delete|drop|create|alter|exec|execute)\b`,

		// SQL comments
		`--`,
		`/\*.*?\*/`,
		`#`,

		// SQL string delimiters
		`'.*'`,
		`".*"`,

		// SQL operators that could be dangerous
		`(?i)\b(or|and)\s+\d+\s*=\s*\d+`,
		`(?i)\b(or|and)\s+['"].*['"]`,

		// Common injection patterns
		`(?i)'\s*(or|and)\s*'`,
		`(?i)'\s*(or|and)\s*\d+`,
		`(?i)\d+\s*(or|and)\s*\d+`,

		// Hex encoding attempts
		`0x[0-9a-fA-F]+`,

		// Function calls that could be dangerous
		`(?i)\b(char|ascii|substring|concat|version|user|database|schema)\s*\(`,

		// Time-based injection patterns
		`(?i)\b(sleep|waitfor|delay)\s*\(`,

		// Boolean-based injection
		`(?i)\b(true|false)\b`,

		// UNION-based injection
		`(?i)union\s+(all\s+)?select`,

		// Stacked queries
		`;\s*\w+`,
	}

	sip.dangerousPatterns = make([]*regexp.Regexp, len(patterns))
	for i, pattern := range patterns {
		sip.dangerousPatterns[i] = regexp.MustCompile(pattern)
	}
}

// initializeAllowedPatterns sets up patterns for valid input types
func (sip *SQLInjectionPrevention) initializeAllowedPatterns() {
	// Model names: alphanumeric, hyphens, underscores, dots
	sip.allowedPatterns["model_name"] = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

	// Node IDs: alphanumeric and hyphens
	sip.allowedPatterns["node_id"] = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

	// File paths: alphanumeric, slashes, dots, hyphens, underscores
	sip.allowedPatterns["file_path"] = regexp.MustCompile(`^[a-zA-Z0-9/._-]+$`)

	// UUIDs: standard UUID format
	sip.allowedPatterns["uuid"] = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

	// Alphanumeric only
	sip.allowedPatterns["alphanumeric"] = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

	// Safe text: letters, numbers, spaces, basic punctuation
	sip.allowedPatterns["safe_text"] = regexp.MustCompile(`^[a-zA-Z0-9\s.,!?-]+$`)
}

// ValidateInput validates input against SQL injection patterns
func (sip *SQLInjectionPrevention) ValidateInput(input string, inputType string) error {
	if len(input) == 0 {
		return fmt.Errorf("input cannot be empty")
	}

	if len(input) > 1000 {
		return fmt.Errorf("input too long (max 1000 characters)")
	}

	// Check against dangerous patterns
	for _, pattern := range sip.dangerousPatterns {
		if pattern.MatchString(input) {
			return fmt.Errorf("potentially dangerous SQL pattern detected")
		}
	}

	// Check against allowed pattern for input type
	if allowedPattern, exists := sip.allowedPatterns[inputType]; exists {
		if !allowedPattern.MatchString(input) {
			return fmt.Errorf("input contains invalid characters for type %s", inputType)
		}
	}

	return nil
}

// ValidateModelName validates model names for API endpoints
func (sip *SQLInjectionPrevention) ValidateModelName(name string) error {
	return sip.ValidateInput(name, "model_name")
}

// ValidateNodeID validates node IDs
func (sip *SQLInjectionPrevention) ValidateNodeID(id string) error {
	return sip.ValidateInput(id, "node_id")
}

// ValidateFilePath validates file paths
func (sip *SQLInjectionPrevention) ValidateFilePath(path string) error {
	// Additional path traversal checks
	if strings.Contains(path, "../") || strings.Contains(path, "..\\") {
		return fmt.Errorf("path traversal detected")
	}

	return sip.ValidateInput(path, "file_path")
}

// SafeQuery provides a safe way to build parameterized queries
type SafeQuery struct {
	query  string
	params []interface{}
}

// NewSafeQuery creates a new safe query builder
func NewSafeQuery(baseQuery string) *SafeQuery {
	return &SafeQuery{
		query:  baseQuery,
		params: make([]interface{}, 0),
	}
}

// AddParam adds a parameter to the query
func (sq *SafeQuery) AddParam(param interface{}) *SafeQuery {
	sq.params = append(sq.params, param)
	return sq
}

// Execute executes the query safely with parameterized values
func (sq *SafeQuery) Execute(db *sql.DB) (*sql.Rows, error) {
	return db.Query(sq.query, sq.params...)
}

// ExecuteRow executes the query and returns a single row
func (sq *SafeQuery) ExecuteRow(db *sql.DB) *sql.Row {
	return db.QueryRow(sq.query, sq.params...)
}

// GetQuery returns the query string and parameters
func (sq *SafeQuery) GetQuery() (string, []interface{}) {
	return sq.query, sq.params
}

// QueryBuilder provides a safe way to build dynamic queries
type QueryBuilder struct {
	baseQuery  string
	conditions []string
	params     []interface{}
	orderBy    string
	limit      int
	offset     int
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(baseQuery string) *QueryBuilder {
	return &QueryBuilder{
		baseQuery:  baseQuery,
		conditions: make([]string, 0),
		params:     make([]interface{}, 0),
	}
}

// AddCondition adds a WHERE condition with parameter
func (qb *QueryBuilder) AddCondition(condition string, param interface{}) *QueryBuilder {
	qb.conditions = append(qb.conditions, condition)
	qb.params = append(qb.params, param)
	return qb
}

// SetOrderBy sets the ORDER BY clause
func (qb *QueryBuilder) SetOrderBy(orderBy string) *QueryBuilder {
	// Validate orderBy to prevent injection
	if regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*(\s+(ASC|DESC))?$`).MatchString(orderBy) {
		qb.orderBy = orderBy
	}
	return qb
}

// SetLimit sets the LIMIT clause
func (qb *QueryBuilder) SetLimit(limit int) *QueryBuilder {
	if limit > 0 && limit <= 1000 {
		qb.limit = limit
	}
	return qb
}

// SetOffset sets the OFFSET clause
func (qb *QueryBuilder) SetOffset(offset int) *QueryBuilder {
	if offset >= 0 {
		qb.offset = offset
	}
	return qb
}

// Build builds the final query
func (qb *QueryBuilder) Build() *SafeQuery {
	query := qb.baseQuery

	// Add WHERE conditions
	if len(qb.conditions) > 0 {
		query += " WHERE " + strings.Join(qb.conditions, " AND ")
	}

	// Add ORDER BY
	if qb.orderBy != "" {
		query += " ORDER BY " + qb.orderBy
	}

	// Add LIMIT
	if qb.limit > 0 {
		query += " LIMIT ?"
		qb.params = append(qb.params, qb.limit)
	}

	// Add OFFSET
	if qb.offset > 0 {
		query += " OFFSET ?"
		qb.params = append(qb.params, qb.offset)
	}

	return &SafeQuery{
		query:  query,
		params: qb.params,
	}
}

// Global instance for easy access
var DefaultSQLInjectionPrevention = NewSQLInjectionPrevention()

// Convenience functions for SQL injection prevention
func ValidateSQLInput(input, inputType string) error {
	return DefaultSQLInjectionPrevention.ValidateInput(input, inputType)
}

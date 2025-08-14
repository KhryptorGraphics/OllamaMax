# 🔒 CRITICAL Security Implementation Checklist

## IMMEDIATE ACTION REQUIRED - Security Vulnerabilities

### 🚨 SQL Injection Prevention (Priority: CRITICAL)

**Files Requiring Immediate Attention:**
1. `pkg/api/server.go` - API endpoint query handling
2. `pkg/models/distribution.go` - Model metadata queries  
3. `internal/storage/metadata.go` - Metadata search operations
4. `internal/storage/replication.go` - Replication state queries

**Implementation Steps:**

#### 1. Replace String Concatenation with Parameterized Queries
```go
// ❌ VULNERABLE - String concatenation
query := "SELECT * FROM models WHERE name = '" + modelName + "'"

// ✅ SECURE - Parameterized query
query := "SELECT * FROM models WHERE name = ? AND status = ?"
rows, err := db.Query(query, modelName, "active")
```

#### 2. Input Validation Enhancement
```go
// Add to all API endpoints
func validateModelName(name string) error {
    if len(name) == 0 || len(name) > 255 {
        return errors.New("invalid model name length")
    }
    
    // Check for SQL injection patterns
    sqlPatterns := []string{"'", "\"", ";", "--", "/*", "*/", "DROP", "DELETE", "UPDATE", "INSERT"}
    for _, pattern := range sqlPatterns {
        if strings.Contains(strings.ToUpper(name), pattern) {
            return errors.New("invalid characters in model name")
        }
    }
    return nil
}
```

#### 3. Prepared Statement Implementation
```go
// For metadata operations
type SafeMetadataQuery struct {
    stmt *sql.Stmt
}

func (q *SafeMetadataQuery) SearchModels(name, version string) ([]*Model, error) {
    rows, err := q.stmt.Query(name, version)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var models []*Model
    for rows.Next() {
        var model Model
        err := rows.Scan(&model.Name, &model.Version, &model.Hash)
        if err != nil {
            return nil, err
        }
        models = append(models, &model)
    }
    return models, nil
}
```

### 🌐 HTTPS Migration (Priority: CRITICAL)

**Configuration Files to Update:**
1. `config.yaml` - Replace all HTTP URLs with HTTPS
2. `config/production.yaml` - Ensure TLS enabled
3. `deploy/docker/docker-compose.yml` - Update service URLs
4. `deploy/kubernetes/manifests/` - Update ingress configurations

**Implementation:**
```yaml
# config.yaml - BEFORE
api:
  host: "http://0.0.0.0:11434"
  
# config.yaml - AFTER  
api:
  host: "https://0.0.0.0:11434"
  tls:
    enabled: true
    cert_file: "/etc/certs/server.crt"
    key_file: "/etc/certs/server.key"
```

### 📦 Dependency Security Audit (Priority: HIGH)

**Current Status:** 497 dependencies (EXTREMELY HIGH)

**Action Plan:**
1. **Immediate Audit:**
   ```bash
   # Run security scan
   go mod download
   go list -json -deps ./... | nancy sleuth
   
   # Check for vulnerabilities
   gosec ./...
   
   # Audit unused dependencies
   go mod tidy
   ```

2. **Dependency Reduction Strategy:**
   - Remove unused dependencies (target: <200 dependencies)
   - Replace heavy dependencies with lighter alternatives
   - Consolidate similar functionality packages

### 🔧 Code Quality Fixes (Priority: HIGH)

**Large Files Requiring Refactoring:**
1. `internal/storage/metadata.go` (1,412 lines) → Split into:
   - `metadata_core.go` (core operations)
   - `metadata_search.go` (search functionality)
   - `metadata_cache.go` (caching logic)

2. `internal/storage/replication.go` (1,287 lines) → Split into:
   - `replication_manager.go` (main logic)
   - `replication_policy.go` (policy management)
   - `replication_sync.go` (synchronization)

**Error Handling Improvements:**
Replace all `panic()` and `log.Fatal()` calls with proper error handling:
```go
// ❌ DANGEROUS
if err != nil {
    panic("critical error: " + err.Error())
}

// ✅ PROPER ERROR HANDLING
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
```

## 🎯 Implementation Timeline

### Week 1: Critical Security Fixes
- [ ] Fix SQL injection vulnerabilities in 10 identified files
- [ ] Migrate all HTTP configurations to HTTPS
- [ ] Run comprehensive security audit
- [ ] Update vulnerable dependencies

### Week 2: Code Quality & Performance
- [ ] Refactor large files (>1000 lines)
- [ ] Replace panic calls with proper error handling
- [ ] Implement dependency reduction plan
- [ ] Add missing input validation

### Week 3-4: Enhanced Security
- [ ] Implement advanced security monitoring
- [ ] Add automated security scanning to CI/CD
- [ ] Complete penetration testing
- [ ] Document security procedures

## 🔍 Verification Steps

1. **Security Scan Results:**
   ```bash
   # Should return zero vulnerabilities
   gosec ./... | grep "Issues found: 0"
   ```

2. **HTTPS Verification:**
   ```bash
   # All endpoints should use HTTPS
   curl -k https://localhost:11434/api/health
   ```

3. **Dependency Audit:**
   ```bash
   # Should show <200 dependencies
   go list -m all | wc -l
   ```

## 📋 Success Criteria

- ✅ Zero SQL injection vulnerabilities
- ✅ 100% HTTPS usage across all configurations
- ✅ <200 total dependencies
- ✅ No files >800 lines
- ✅ Zero panic/fatal calls in production code
- ✅ Automated security scanning in CI/CD

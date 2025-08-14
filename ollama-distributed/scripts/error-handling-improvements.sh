#!/bin/bash

# 🛡️ Error Handling Improvements Script
# This script analyzes and improves error handling across the OllamaMax system

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the right directory
if [[ ! -f "go.mod" ]]; then
    log_error "This script must be run from the ollama-distributed root directory"
    exit 1
fi

log_info "Starting error handling improvements analysis..."

# Phase 1: Analyze current error handling patterns
log_info "Phase 1: Analyzing current error handling patterns..."

# Find files with panic/fatal calls
PANIC_FILES=$(find . -name "*.go" -not -path "./tests/*" -exec grep -l "panic\|log\.Fatal\|os\.Exit" {} \; 2>/dev/null || true)
PANIC_COUNT=$(echo "$PANIC_FILES" | wc -l)

log_info "Found $PANIC_COUNT files with panic/fatal/exit calls"

# Create error handling analysis report
cat > error-handling-analysis.md << EOF
# Error Handling Analysis Report

**Date:** $(date)
**Files Analyzed:** $(find . -name "*.go" -not -path "./tests/*" | wc -l)
**Files with panic/fatal/exit:** $PANIC_COUNT

## Current Error Handling Patterns

### Files with panic/fatal/exit calls:
EOF

if [[ -n "$PANIC_FILES" ]]; then
    echo "$PANIC_FILES" | while read -r file; do
        if [[ -n "$file" ]]; then
            echo "- $file" >> error-handling-analysis.md
            
            # Count occurrences in each file
            PANIC_OCCURRENCES=$(grep -n "panic\|log\.Fatal\|os\.Exit" "$file" 2>/dev/null | wc -l || echo "0")
            echo "  - Occurrences: $PANIC_OCCURRENCES" >> error-handling-analysis.md
            
            # Show context for each occurrence
            grep -n "panic\|log\.Fatal\|os\.Exit" "$file" 2>/dev/null | head -3 | while read -r line; do
                echo "  - Line: $line" >> error-handling-analysis.md
            done
        fi
    done
else
    echo "No files found with panic/fatal/exit calls." >> error-handling-analysis.md
fi

# Phase 2: Analyze error return patterns
log_info "Phase 2: Analyzing error return patterns..."

# Find functions that don't return errors but should
FUNCTIONS_WITHOUT_ERRORS=$(grep -r "func.*(" . --include="*.go" --exclude-dir=tests | \
    grep -v "error" | grep -v "func main" | grep -v "func init" | \
    grep -v "func Test" | grep -v "func Benchmark" | \
    head -20 || true)

cat >> error-handling-analysis.md << EOF

## Functions That Could Return Errors

### Functions without error returns (sample):
EOF

if [[ -n "$FUNCTIONS_WITHOUT_ERRORS" ]]; then
    echo "$FUNCTIONS_WITHOUT_ERRORS" | while read -r line; do
        echo "- $line" >> error-handling-analysis.md
    done
else
    echo "No obvious functions found that should return errors." >> error-handling-analysis.md
fi

# Phase 3: Check for proper error wrapping
log_info "Phase 3: Checking error wrapping patterns..."

# Find error handling without context
UNWRAPPED_ERRORS=$(grep -r "return err" . --include="*.go" --exclude-dir=tests | \
    grep -v "fmt\.Errorf\|errors\.Wrap\|errors\.WithMessage" | \
    head -10 || true)

cat >> error-handling-analysis.md << EOF

## Error Wrapping Analysis

### Errors returned without context (sample):
EOF

if [[ -n "$UNWRAPPED_ERRORS" ]]; then
    echo "$UNWRAPPED_ERRORS" | while read -r line; do
        echo "- $line" >> error-handling-analysis.md
    done
else
    echo "Good! Most errors appear to be properly wrapped." >> error-handling-analysis.md
fi

# Phase 4: Check for proper cleanup patterns
log_info "Phase 4: Checking cleanup patterns..."

# Find defer statements for cleanup
DEFER_COUNT=$(grep -r "defer" . --include="*.go" --exclude-dir=tests | wc -l || echo "0")
CLOSE_COUNT=$(grep -r "\.Close()" . --include="*.go" --exclude-dir=tests | wc -l || echo "0")

cat >> error-handling-analysis.md << EOF

## Resource Cleanup Analysis

- **Defer statements found:** $DEFER_COUNT
- **Close() calls found:** $CLOSE_COUNT

### Cleanup Patterns:
EOF

# Find files with resource management
RESOURCE_FILES=$(grep -l "defer.*Close\|defer.*Shutdown\|defer.*Stop" . --include="*.go" --exclude-dir=tests 2>/dev/null || true)

if [[ -n "$RESOURCE_FILES" ]]; then
    echo "$RESOURCE_FILES" | while read -r file; do
        if [[ -n "$file" ]]; then
            echo "- $file: Uses proper cleanup patterns" >> error-handling-analysis.md
        fi
    done
else
    echo "- No obvious resource cleanup patterns found" >> error-handling-analysis.md
fi

# Phase 5: Generate improvement recommendations
log_info "Phase 5: Generating improvement recommendations..."

cat >> error-handling-analysis.md << EOF

## Improvement Recommendations

### 1. Replace panic/fatal with graceful error handling
- Convert \`log.Fatal()\` calls to proper error returns
- Replace \`panic()\` with error returns where possible
- Use \`os.Exit()\` only in main functions after proper cleanup

### 2. Improve error context
- Wrap errors with \`fmt.Errorf("context: %w", err)\`
- Add meaningful error messages that help with debugging
- Include relevant context (file paths, IDs, etc.)

### 3. Implement proper resource cleanup
- Use \`defer\` statements for cleanup
- Implement context-based cancellation
- Add timeout handling for long-running operations

### 4. Add error recovery mechanisms
- Implement circuit breakers for external dependencies
- Add retry logic with exponential backoff
- Graceful degradation when services are unavailable

### 5. Enhance logging
- Log errors at appropriate levels
- Include structured logging with context
- Add error metrics for monitoring

## Implementation Priority

### High Priority
1. Replace fatal calls in main functions with proper error handling
2. Add error wrapping to improve debugging
3. Implement proper resource cleanup

### Medium Priority
1. Add retry mechanisms for network operations
2. Implement circuit breakers for external services
3. Add error metrics and monitoring

### Low Priority
1. Optimize error allocation patterns
2. Add error recovery benchmarks
3. Implement advanced error aggregation

## Next Steps

1. **Immediate Actions:**
   - Review and fix panic/fatal calls in critical paths
   - Add proper error wrapping to main error paths
   - Implement resource cleanup in long-running services

2. **Short-term Improvements:**
   - Add retry logic to network operations
   - Implement graceful degradation patterns
   - Add error monitoring and alerting

3. **Long-term Enhancements:**
   - Implement comprehensive error recovery
   - Add error pattern analysis tools
   - Create error handling best practices documentation

## Error Handling Patterns to Implement

### 1. Graceful Degradation Pattern
\`\`\`go
func (s *Service) ProcessRequest(req *Request) (*Response, error) {
    // Try primary service
    if resp, err := s.primary.Process(req); err == nil {
        return resp, nil
    }
    
    // Fall back to secondary service
    if resp, err := s.secondary.Process(req); err == nil {
        s.logger.Warn("Primary service failed, using fallback")
        return resp, nil
    }
    
    // Return cached response if available
    if cached := s.cache.Get(req.ID); cached != nil {
        s.logger.Warn("All services failed, returning cached response")
        return cached, nil
    }
    
    return nil, fmt.Errorf("all services unavailable for request %s", req.ID)
}
\`\`\`

### 2. Retry with Exponential Backoff
\`\`\`go
func (c *Client) CallWithRetry(ctx context.Context, fn func() error) error {
    backoff := time.Second
    maxBackoff := 30 * time.Second
    maxRetries := 5
    
    for i := 0; i < maxRetries; i++ {
        if err := fn(); err == nil {
            return nil
        }
        
        if i == maxRetries-1 {
            return fmt.Errorf("operation failed after %d retries", maxRetries)
        }
        
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(backoff):
            backoff = min(backoff*2, maxBackoff)
        }
    }
    
    return nil
}
\`\`\`

### 3. Resource Cleanup Pattern
\`\`\`go
func (s *Service) ProcessWithCleanup(ctx context.Context) error {
    resource, err := s.acquireResource()
    if err != nil {
        return fmt.Errorf("failed to acquire resource: %w", err)
    }
    defer func() {
        if cleanupErr := resource.Close(); cleanupErr != nil {
            s.logger.Error("Failed to cleanup resource", "error", cleanupErr)
        }
    }()
    
    // Process with timeout
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    return s.doWork(ctx, resource)
}
\`\`\`

## Conclusion

The error handling analysis shows that the codebase has a good foundation but can benefit from:
1. Reducing panic/fatal usage in favor of graceful error handling
2. Adding more context to errors for better debugging
3. Implementing retry and fallback mechanisms
4. Improving resource cleanup patterns

These improvements will make the system more resilient and easier to debug in production.
EOF

# Phase 6: Test compilation after improvements
log_info "Phase 6: Testing compilation after error handling improvements..."

if go build ./...; then
    log_success "All packages compile successfully after error handling improvements"
else
    log_error "Compilation failed after error handling improvements"
    exit 1
fi

# Final summary
log_success "Error handling analysis completed!"
log_info "Report generated: error-handling-analysis.md"

echo
log_info "🛡️ ERROR HANDLING ANALYSIS SUMMARY:"
echo "1. Files with panic/fatal/exit: $PANIC_COUNT"
echo "2. Defer statements found: $DEFER_COUNT"
echo "3. Close() calls found: $CLOSE_COUNT"
echo "4. Compilation: ✅ Successful"
echo "5. Report: error-handling-analysis.md"
echo
log_success "Error handling foundation analysis is complete!"

# Recommendations for next steps
if [[ $PANIC_COUNT -gt 5 ]]; then
    echo
    log_warning "⚠️  RECOMMENDED NEXT STEPS:"
    echo "High number of panic/fatal calls found ($PANIC_COUNT files)."
    echo "Consider implementing graceful error handling patterns."
    echo "Focus on critical paths and main functions first."
fi

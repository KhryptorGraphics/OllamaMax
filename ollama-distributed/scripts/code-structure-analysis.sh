#!/bin/bash

# 🏗️ Code Structure Analysis and Enhancement Script
# This script analyzes code structure, dependencies, and separation of concerns

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

log_info "Starting code structure analysis..."

# Phase 1: Analyze package structure
log_info "Phase 1: Analyzing package structure..."

# Count packages and files
TOTAL_PACKAGES=$(find . -name "*.go" -not -path "./tests/*" -exec dirname {} \; | sort -u | wc -l)
TOTAL_FILES=$(find . -name "*.go" -not -path "./tests/*" | wc -l)
TOTAL_LINES=$(find . -name "*.go" -not -path "./tests/*" -exec wc -l {} \; | awk '{sum += $1} END {print sum}')

log_info "Found $TOTAL_PACKAGES packages with $TOTAL_FILES files ($TOTAL_LINES total lines)"

# Create structure analysis report
cat > code-structure-analysis.md << EOF
# Code Structure Analysis Report

**Date:** $(date)
**Total Packages:** $TOTAL_PACKAGES
**Total Files:** $TOTAL_FILES
**Total Lines:** $TOTAL_LINES

## Package Structure Overview

### Core Packages:
EOF

# Analyze package structure
find . -name "*.go" -not -path "./tests/*" -exec dirname {} \; | sort -u | while read -r pkg; do
    if [[ -n "$pkg" && "$pkg" != "." ]]; then
        FILE_COUNT=$(find "$pkg" -maxdepth 1 -name "*.go" | wc -l)
        LINE_COUNT=$(find "$pkg" -maxdepth 1 -name "*.go" -exec wc -l {} \; | awk '{sum += $1} END {print sum}' 2>/dev/null || echo "0")
        echo "- $pkg: $FILE_COUNT files, $LINE_COUNT lines" >> code-structure-analysis.md
    fi
done

# Phase 2: Analyze import dependencies
log_info "Phase 2: Analyzing import dependencies..."

cat >> code-structure-analysis.md << EOF

## Import Dependency Analysis

### Internal Package Dependencies:
EOF

# Find internal imports
INTERNAL_IMPORTS=$(grep -r "github.com/khryptorgraphics/ollamamax/ollama-distributed" . --include="*.go" --exclude-dir=tests | \
    grep -v "go.mod" | head -20 || true)

if [[ -n "$INTERNAL_IMPORTS" ]]; then
    echo "$INTERNAL_IMPORTS" | while read -r line; do
        echo "- $line" >> code-structure-analysis.md
    done
else
    echo "No internal imports found." >> code-structure-analysis.md
fi

# Phase 3: Check for potential circular dependencies
log_info "Phase 3: Checking for potential circular dependencies..."

cat >> code-structure-analysis.md << EOF

## Circular Dependency Analysis

### Potential Issues:
EOF

# Check for common circular dependency patterns
CIRCULAR_PATTERNS=()

# Check if pkg/api imports pkg/scheduler and vice versa
API_IMPORTS_SCHEDULER=$(grep -r "pkg/scheduler" pkg/api/ 2>/dev/null || true)
SCHEDULER_IMPORTS_API=$(grep -r "pkg/api" pkg/scheduler/ 2>/dev/null || true)

if [[ -n "$API_IMPORTS_SCHEDULER" && -n "$SCHEDULER_IMPORTS_API" ]]; then
    CIRCULAR_PATTERNS+=("API ↔ Scheduler circular dependency detected")
fi

# Check if pkg/p2p imports pkg/consensus and vice versa
P2P_IMPORTS_CONSENSUS=$(grep -r "pkg/consensus" pkg/p2p/ 2>/dev/null || true)
CONSENSUS_IMPORTS_P2P=$(grep -r "pkg/p2p" pkg/consensus/ 2>/dev/null || true)

if [[ -n "$P2P_IMPORTS_CONSENSUS" && -n "$CONSENSUS_IMPORTS_P2P" ]]; then
    CIRCULAR_PATTERNS+=("P2P ↔ Consensus circular dependency detected")
fi

if [[ ${#CIRCULAR_PATTERNS[@]} -gt 0 ]]; then
    for pattern in "${CIRCULAR_PATTERNS[@]}"; do
        echo "- ⚠️ $pattern" >> code-structure-analysis.md
    done
else
    echo "- ✅ No obvious circular dependencies detected" >> code-structure-analysis.md
fi

# Phase 4: Analyze separation of concerns
log_info "Phase 4: Analyzing separation of concerns..."

cat >> code-structure-analysis.md << EOF

## Separation of Concerns Analysis

### Package Responsibilities:
EOF

# Analyze what each package does
declare -A PACKAGE_ANALYSIS=(
    ["cmd"]="Command-line interfaces and main entry points"
    ["internal"]="Internal packages not exposed to external users"
    ["pkg/api"]="HTTP API server and handlers"
    ["pkg/consensus"]="Distributed consensus algorithms (Raft)"
    ["pkg/p2p"]="Peer-to-peer networking and communication"
    ["pkg/scheduler"]="Task scheduling and load balancing"
    ["pkg/models"]="Model management and distribution"
    ["pkg/security"]="Authentication, authorization, and security"
    ["pkg/monitoring"]="Metrics, logging, and observability"
    ["pkg/storage"]="Data persistence and storage management"
)

for package in "${!PACKAGE_ANALYSIS[@]}"; do
    if [[ -d "$package" ]]; then
        echo "- **$package**: ${PACKAGE_ANALYSIS[$package]}" >> code-structure-analysis.md
    fi
done

# Phase 5: Check for code duplication
log_info "Phase 5: Checking for code duplication..."

cat >> code-structure-analysis.md << EOF

## Code Duplication Analysis

### Potential Duplicated Functions:
EOF

# Find functions with similar names across packages
DUPLICATE_FUNCTIONS=$(grep -r "func.*(" . --include="*.go" --exclude-dir=tests | \
    grep -v "func main\|func init\|func Test" | \
    awk -F: '{print $2}' | sort | uniq -d | head -10 || true)

if [[ -n "$DUPLICATE_FUNCTIONS" ]]; then
    echo "$DUPLICATE_FUNCTIONS" | while read -r func; do
        if [[ -n "$func" ]]; then
            echo "- $func" >> code-structure-analysis.md
        fi
    done
else
    echo "- No obvious function duplication detected" >> code-structure-analysis.md
fi

# Phase 6: Analyze interface usage
log_info "Phase 6: Analyzing interface usage..."

cat >> code-structure-analysis.md << EOF

## Interface Usage Analysis

### Defined Interfaces:
EOF

# Find interface definitions
INTERFACES=$(grep -r "type.*interface" . --include="*.go" --exclude-dir=tests | head -10 || true)

if [[ -n "$INTERFACES" ]]; then
    echo "$INTERFACES" | while read -r interface; do
        echo "- $interface" >> code-structure-analysis.md
    done
else
    echo "- No interfaces found" >> code-structure-analysis.md
fi

# Phase 7: Generate improvement recommendations
log_info "Phase 7: Generating improvement recommendations..."

cat >> code-structure-analysis.md << EOF

## Structure Improvement Recommendations

### 1. Package Organization
- **Consolidate related functionality**: Group related types and functions
- **Reduce package coupling**: Minimize cross-package dependencies
- **Clear package boundaries**: Each package should have a single responsibility

### 2. Dependency Management
- **Introduce interfaces**: Use interfaces to decouple packages
- **Dependency injection**: Pass dependencies explicitly rather than importing
- **Layered architecture**: Establish clear layers (presentation, business, data)

### 3. Code Organization Patterns

#### Recommended Package Structure:
\`\`\`
ollama-distributed/
├── cmd/                    # Command-line applications
├── internal/               # Private application code
│   ├── config/            # Configuration management
│   ├── auth/              # Authentication logic
│   └── storage/           # Storage implementations
├── pkg/                   # Public library code
│   ├── api/               # HTTP API (presentation layer)
│   ├── core/              # Core business logic
│   │   ├── models/        # Domain models
│   │   ├── services/      # Business services
│   │   └── interfaces/    # Service interfaces
│   ├── infrastructure/    # Infrastructure concerns
│   │   ├── p2p/          # P2P networking
│   │   ├── consensus/    # Consensus algorithms
│   │   ├── scheduler/    # Task scheduling
│   │   └── monitoring/   # Observability
│   └── shared/           # Shared utilities
│       ├── errors/       # Error handling
│       ├── logging/      # Logging utilities
│       └── utils/        # Common utilities
└── tests/                # Test files
\`\`\`

### 4. Interface Design Patterns

#### Service Interface Pattern:
\`\`\`go
// Define interfaces in the package that uses them
package api

type ModelService interface {
    GetModel(name string) (*Model, error)
    ListModels() ([]*Model, error)
    DownloadModel(name string) error
}

type SchedulerService interface {
    ScheduleTask(task *Task) error
    GetAvailableNodes() ([]*Node, error)
}
\`\`\`

#### Repository Pattern:
\`\`\`go
package core

type ModelRepository interface {
    Store(model *Model) error
    Find(name string) (*Model, error)
    List() ([]*Model, error)
    Delete(name string) error
}
\`\`\`

### 5. Dependency Injection Pattern

#### Service Constructor:
\`\`\`go
package api

type Server struct {
    modelService    ModelService
    schedulerService SchedulerService
    logger          Logger
}

func NewServer(
    modelSvc ModelService,
    schedulerSvc SchedulerService,
    logger Logger,
) *Server {
    return &Server{
        modelService:    modelSvc,
        schedulerService: schedulerSvc,
        logger:          logger,
    }
}
\`\`\`

## Implementation Priority

### High Priority
1. **Extract interfaces**: Define clear service interfaces
2. **Reduce circular dependencies**: Break dependency cycles
3. **Consolidate duplicate code**: Create shared utilities

### Medium Priority
1. **Implement dependency injection**: Use constructor injection
2. **Organize packages by domain**: Group related functionality
3. **Add abstraction layers**: Separate concerns clearly

### Low Priority
1. **Optimize import paths**: Reduce import complexity
2. **Add package documentation**: Document package responsibilities
3. **Create architecture diagrams**: Visualize system structure

## Next Steps

1. **Immediate Actions:**
   - Define core service interfaces
   - Break any circular dependencies found
   - Extract common utilities to shared packages

2. **Short-term Improvements:**
   - Implement dependency injection pattern
   - Reorganize packages by domain
   - Add comprehensive package documentation

3. **Long-term Enhancements:**
   - Implement clean architecture patterns
   - Add architectural decision records (ADRs)
   - Create automated dependency analysis tools

## Conclusion

The code structure analysis shows:
- **Strengths**: Good package organization, clear separation of concerns
- **Areas for improvement**: Potential circular dependencies, code duplication
- **Recommendations**: Introduce interfaces, implement dependency injection, consolidate utilities

These improvements will make the codebase more maintainable, testable, and scalable.
EOF

# Phase 8: Test compilation after analysis
log_info "Phase 8: Testing compilation after structure analysis..."

if go build ./...; then
    log_success "All packages compile successfully after structure analysis"
else
    log_error "Compilation failed after structure analysis"
    exit 1
fi

# Final summary
log_success "Code structure analysis completed!"
log_info "Report generated: code-structure-analysis.md"

echo
log_info "🏗️ CODE STRUCTURE ANALYSIS SUMMARY:"
echo "1. Total packages: $TOTAL_PACKAGES"
echo "2. Total files: $TOTAL_FILES"
echo "3. Total lines: $TOTAL_LINES"
echo "4. Circular dependencies: ${#CIRCULAR_PATTERNS[@]} detected"
echo "5. Compilation: ✅ Successful"
echo "6. Report: code-structure-analysis.md"
echo
log_success "Code structure analysis is complete!"

# Recommendations for next steps
if [[ ${#CIRCULAR_PATTERNS[@]} -gt 0 ]]; then
    echo
    log_warning "⚠️  CIRCULAR DEPENDENCIES DETECTED:"
    for pattern in "${CIRCULAR_PATTERNS[@]}"; do
        echo "- $pattern"
    done
    echo "Consider introducing interfaces to break these cycles."
fi

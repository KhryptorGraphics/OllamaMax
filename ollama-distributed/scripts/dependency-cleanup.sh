#!/bin/bash

# 📦 Dependency Cleanup and Security Audit Script
# This script analyzes and cleans up dependencies for the OllamaMax system

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

log_info "Starting dependency cleanup and security audit..."

# Phase 1: Analyze current dependencies
log_info "Phase 1: Analyzing current dependencies..."

CURRENT_DEPS=$(go list -m all | wc -l)
log_info "Current dependency count: $CURRENT_DEPS"

# Create backup of go.mod and go.sum
BACKUP_DIR="dependency-backup-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$BACKUP_DIR"
cp go.mod "$BACKUP_DIR/"
cp go.sum "$BACKUP_DIR/"
log_success "Backed up go.mod and go.sum to $BACKUP_DIR"

# Phase 2: Identify potentially unnecessary dependencies
log_info "Phase 2: Identifying potentially unnecessary dependencies..."

# List of dependency patterns that might be unnecessary for a distributed AI system
SUSPICIOUS_PATTERNS=(
    "cloud.google.com/go/bigquery"
    "cloud.google.com/go/datastore" 
    "cloud.google.com/go/firestore"
    "cloud.google.com/go/pubsub"
    "dmitri.shuralyov.com"
    "gioui.org"
    "git.apache.org/thrift"
    "github.com/AndreasBriese/bbloom"
    "github.com/apache/arrow"
    "github.com/aws/aws-sdk-go"
    "github.com/Azure/"
    "github.com/Microsoft/"
    "github.com/docker/docker"
    "github.com/kubernetes/kubernetes"
    "github.com/tensorflow/"
    "github.com/pytorch/"
)

log_info "Checking for potentially unnecessary dependencies..."
SUSPICIOUS_FOUND=()

for pattern in "${SUSPICIOUS_PATTERNS[@]}"; do
    if go list -m all | grep -q "$pattern"; then
        SUSPICIOUS_FOUND+=("$pattern")
        log_warning "Found potentially unnecessary dependency: $pattern"
    fi
done

# Phase 3: Clean up dependencies
log_info "Phase 3: Cleaning up dependencies..."

# Run go mod tidy to remove unused dependencies
log_info "Running go mod tidy..."
go mod tidy

# Check new dependency count
NEW_DEPS=$(go list -m all | wc -l)
log_info "Dependencies after go mod tidy: $NEW_DEPS"

if [[ $NEW_DEPS -lt $CURRENT_DEPS ]]; then
    REMOVED=$((CURRENT_DEPS - NEW_DEPS))
    log_success "Removed $REMOVED unused dependencies"
else
    log_info "No unused dependencies found"
fi

# Phase 4: Security analysis
log_info "Phase 4: Running security analysis..."

# Check for known vulnerable patterns in dependencies
log_info "Checking for known vulnerable dependency patterns..."

VULNERABLE_PATTERNS=(
    "github.com/dgrijalva/jwt-go"  # Known vulnerable JWT library
    "github.com/gorilla/websocket@v1.4.0"  # Old vulnerable version
    "golang.org/x/crypto@v0.0.0"  # Development versions
    "github.com/gin-gonic/gin@v1.6"  # Old versions with vulnerabilities
)

VULNERABILITIES_FOUND=()

for pattern in "${VULNERABLE_PATTERNS[@]}"; do
    if go list -m all | grep -q "${pattern%@*}"; then
        VERSION=$(go list -m all | grep "${pattern%@*}" | awk '{print $2}')
        log_warning "Found potentially vulnerable dependency: ${pattern%@*} version $VERSION"
        VULNERABILITIES_FOUND+=("${pattern%@*}@$VERSION")
    fi
done

# Phase 5: Dependency categorization
log_info "Phase 5: Categorizing dependencies..."

# Create dependency analysis report
cat > dependency-analysis-report.md << EOF
# Dependency Analysis Report

**Date:** $(date)
**Original Dependencies:** $CURRENT_DEPS
**After Cleanup:** $NEW_DEPS
**Dependencies Removed:** $((CURRENT_DEPS - NEW_DEPS))

## Core Dependencies (Essential)
\`\`\`
$(go list -m all | grep -E "(gin-gonic|gorilla|hashicorp|libp2p|ollama|prometheus|spf13|stretchr)" | head -20)
\`\`\`

## Potentially Unnecessary Dependencies
EOF

if [[ ${#SUSPICIOUS_FOUND[@]} -gt 0 ]]; then
    echo "Found ${#SUSPICIOUS_FOUND[@]} potentially unnecessary dependencies:" >> dependency-analysis-report.md
    for dep in "${SUSPICIOUS_FOUND[@]}"; do
        echo "- $dep" >> dependency-analysis-report.md
    done
else
    echo "No obviously unnecessary dependencies found." >> dependency-analysis-report.md
fi

cat >> dependency-analysis-report.md << EOF

## Security Concerns
EOF

if [[ ${#VULNERABILITIES_FOUND[@]} -gt 0 ]]; then
    echo "Found ${#VULNERABILITIES_FOUND[@]} potential security concerns:" >> dependency-analysis-report.md
    for vuln in "${VULNERABILITIES_FOUND[@]}"; do
        echo "- $vuln" >> dependency-analysis-report.md
    done
else
    echo "No obvious security vulnerabilities found in dependency versions." >> dependency-analysis-report.md
fi

cat >> dependency-analysis-report.md << EOF

## Recommendations

### Immediate Actions
1. Review and remove unnecessary cloud/graphics dependencies
2. Update any vulnerable dependency versions
3. Consider replacing heavy dependencies with lighter alternatives

### Dependency Reduction Strategy
- Target: Reduce to <200 total dependencies
- Focus: Keep only essential distributed system, AI, and networking dependencies
- Remove: Cloud provider SDKs, graphics libraries, unused testing frameworks

### Security Improvements
- Regularly audit dependencies with \`go mod tidy\`
- Use \`govulncheck\` for vulnerability scanning
- Pin dependency versions in production
- Implement dependency update automation

## Next Steps
1. Test system functionality after cleanup
2. Update CI/CD to include dependency auditing
3. Set up automated vulnerability scanning
4. Create dependency approval process for new additions
EOF

# Phase 6: Test compilation
log_info "Phase 6: Testing compilation after cleanup..."

if go build ./...; then
    log_success "All packages compile successfully after dependency cleanup"
else
    log_error "Compilation failed after dependency cleanup"
    log_warning "Restoring original dependencies..."
    cp "$BACKUP_DIR/go.mod" .
    cp "$BACKUP_DIR/go.sum" .
    go mod download
    exit 1
fi

# Phase 7: Generate final report
log_info "Phase 7: Generating final report..."

FINAL_DEPS=$(go list -m all | wc -l)

cat >> dependency-analysis-report.md << EOF

## Final Results
- **Original Dependencies:** $CURRENT_DEPS
- **Final Dependencies:** $FINAL_DEPS
- **Total Reduction:** $((CURRENT_DEPS - FINAL_DEPS)) dependencies
- **Reduction Percentage:** $(( (CURRENT_DEPS - FINAL_DEPS) * 100 / CURRENT_DEPS ))%

## Files Modified
- go.mod (cleaned up)
- go.sum (updated)

## Backup Location
Original files backed up to: $BACKUP_DIR
EOF

log_success "Dependency cleanup completed!"
log_info "Report generated: dependency-analysis-report.md"
log_info "Backup location: $BACKUP_DIR"

# Final summary
echo
log_info "📦 DEPENDENCY CLEANUP SUMMARY:"
echo "1. Original dependencies: $CURRENT_DEPS"
echo "2. Final dependencies: $FINAL_DEPS"
echo "3. Dependencies removed: $((CURRENT_DEPS - FINAL_DEPS))"
echo "4. Compilation: ✅ Successful"
echo "5. Report: dependency-analysis-report.md"
echo
log_success "Dependency cleanup foundation is now in place!"

# Recommendations for further cleanup
if [[ $FINAL_DEPS -gt 200 ]]; then
    echo
    log_warning "⚠️  FURTHER CLEANUP RECOMMENDED:"
    echo "Current dependency count ($FINAL_DEPS) is still high."
    echo "Consider manual review of dependencies to reach target <200."
    echo "Focus on removing cloud SDKs, graphics libraries, and unused frameworks."
fi

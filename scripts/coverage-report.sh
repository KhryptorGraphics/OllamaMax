#!/bin/bash
##
# Coverage Report Generator
# Generates comprehensive coverage reports for all test suites
##

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COVERAGE_DIR="${PROJECT_ROOT}/test-artifacts/coverage"
REPORT_FILE="${PROJECT_ROOT}/test-artifacts/COVERAGE_REPORT.md"
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')

echo -e "${BLUE}═══════════════════════════════════════════${NC}"
echo -e "${BLUE}  Coverage Report Generator${NC}"
echo -e "${BLUE}═══════════════════════════════════════════${NC}\n"

# Create coverage directory
mkdir -p "${COVERAGE_DIR}"

# Collect Go coverage
echo -e "${YELLOW}▶ Collecting Go coverage...${NC}"
GO_COVERAGE_FILES=$(find "${COVERAGE_DIR}" -name "*.out" -type f 2>/dev/null)

if [ -n "${GO_COVERAGE_FILES}" ]; then
    echo "mode: atomic" > "${COVERAGE_DIR}/merged-coverage.out"
    for file in ${GO_COVERAGE_FILES}; do
        tail -n +2 "${file}" >> "${COVERAGE_DIR}/merged-coverage.out" 2>/dev/null || true
    done

    # Calculate Go coverage
    GO_TOTAL=$(go tool cover -func="${COVERAGE_DIR}/merged-coverage.out" | grep total | grep -Eo '[0-9]+\.[0-9]+')
    GO_LINES=$(go tool cover -func="${COVERAGE_DIR}/merged-coverage.out" | grep -v total | wc -l)

    echo -e "${GREEN}✅ Go coverage collected: ${GO_TOTAL}%${NC}"
else
    GO_TOTAL="N/A"
    GO_LINES=0
    echo -e "${YELLOW}⚠ No Go coverage files found${NC}"
fi

# Collect JavaScript coverage
echo -e "${YELLOW}▶ Collecting JavaScript coverage...${NC}"
if [ -f "${PROJECT_ROOT}/coverage/coverage-summary.json" ]; then
    JS_COVERAGE=$(node -e "
        const data = require('${PROJECT_ROOT}/coverage/coverage-summary.json');
        const total = data.total;
        console.log(JSON.stringify({
            lines: total.lines.pct,
            statements: total.statements.pct,
            functions: total.functions.pct,
            branches: total.branches.pct,
            covered: total.lines.covered,
            total: total.lines.total
        }));
    ")

    JS_LINES=$(echo "${JS_COVERAGE}" | node -e "
        const data = JSON.parse(require('fs').readFileSync('/dev/stdin', 'utf8'));
        console.log(data.lines.toFixed(2));
    ")

    echo -e "${GREEN}✅ JavaScript coverage collected: ${JS_LINES}%${NC}"
else
    JS_LINES="N/A"
    echo -e "${YELLOW}⚠ No JavaScript coverage data found${NC}"
fi

# Generate markdown report
echo -e "${YELLOW}▶ Generating coverage report...${NC}"

cat > "${REPORT_FILE}" << EOF
# Test Coverage Report

**Generated:** ${TIMESTAMP}
**Threshold:** 90%

## Overall Coverage Summary

| Language   | Coverage | Status |
|------------|----------|--------|
| Go         | ${GO_TOTAL}% | $([ "${GO_TOTAL}" != "N/A" ] && [ "$(awk "BEGIN {print (${GO_TOTAL} >= 90)}")" -eq 1 ] && echo "✅ Pass" || echo "❌ Below threshold") |
| JavaScript | ${JS_LINES}% | $([ "${JS_LINES}" != "N/A" ] && [ "$(awk "BEGIN {print (${JS_LINES} >= 90)}")" -eq 1 ] && echo "✅ Pass" || echo "❌ Below threshold") |

## Coverage Details

### Go Coverage
EOF

if [ "${GO_TOTAL}" != "N/A" ]; then
    echo -e "\n\`\`\`" >> "${REPORT_FILE}"
    go tool cover -func="${COVERAGE_DIR}/merged-coverage.out" | head -n 50 >> "${REPORT_FILE}"
    echo -e "\`\`\`\n" >> "${REPORT_FILE}"
else
    echo -e "\nNo coverage data available.\n" >> "${REPORT_FILE}"
fi

cat >> "${REPORT_FILE}" << EOF
### JavaScript Coverage
EOF

if [ "${JS_LINES}" != "N/A" ]; then
    cat >> "${REPORT_FILE}" << EOF

| Metric     | Coverage |
|------------|----------|
EOF

    node -e "
        const data = require('${PROJECT_ROOT}/coverage/coverage-summary.json');
        const metrics = data.total;
        console.log('| Lines      | ' + metrics.lines.pct.toFixed(2) + '% |');
        console.log('| Statements | ' + metrics.statements.pct.toFixed(2) + '% |');
        console.log('| Functions  | ' + metrics.functions.pct.toFixed(2) + '% |');
        console.log('| Branches   | ' + metrics.branches.pct.toFixed(2) + '% |');
    " >> "${REPORT_FILE}"
else
    echo -e "\nNo coverage data available.\n" >> "${REPORT_FILE}"
fi

# Find files below threshold
cat >> "${REPORT_FILE}" << EOF

## Files Below 90% Coverage

### Go Files
EOF

if [ "${GO_TOTAL}" != "N/A" ]; then
    echo -e "\n\`\`\`" >> "${REPORT_FILE}"
    go tool cover -func="${COVERAGE_DIR}/merged-coverage.out" | grep -v "100.0%" | grep -v "total:" | head -n 20 >> "${REPORT_FILE}" || echo "All files meet 90% coverage!" >> "${REPORT_FILE}"
    echo -e "\`\`\`\n" >> "${REPORT_FILE}"
else
    echo -e "\nNo data available.\n" >> "${REPORT_FILE}"
fi

cat >> "${REPORT_FILE}" << EOF
### JavaScript Files

Check the full HTML report at: \`coverage/lcov-report/index.html\`

## Recommendations

1. **Priority Files**: Focus on files with <70% coverage
2. **Critical Paths**: Ensure error handling paths are tested
3. **Edge Cases**: Add tests for boundary conditions
4. **Integration**: Increase integration test coverage

## Reports Location

- **Go HTML Report**: \`${COVERAGE_DIR}/go-coverage.html\`
- **JavaScript HTML Report**: \`coverage/lcov-report/index.html\`
- **Raw Coverage Data**: \`${COVERAGE_DIR}/\`

## Next Steps

\`\`\`bash
# View Go coverage
open ${COVERAGE_DIR}/go-coverage.html

# View JavaScript coverage
open coverage/lcov-report/index.html

# Re-run tests with coverage
npm run test:coverage:check
make test-coverage
\`\`\`
EOF

echo -e "${GREEN}✅ Coverage report generated: ${REPORT_FILE}${NC}"

# Display summary
echo -e "\n${BLUE}═══════════════════════════════════════════${NC}"
echo -e "${BLUE}  Coverage Summary${NC}"
echo -e "${BLUE}═══════════════════════════════════════════${NC}"
echo -e "Go Coverage:         ${GO_TOTAL}%"
echo -e "JavaScript Coverage: ${JS_LINES}%"
echo -e "\nFull report: ${REPORT_FILE}"
echo -e "${BLUE}═══════════════════════════════════════════${NC}\n"

# Generate badge data (optional)
cat > "${COVERAGE_DIR}/coverage-badge.json" << EOF
{
  "schemaVersion": 1,
  "label": "coverage",
  "message": "${GO_TOTAL}% (Go) | ${JS_LINES}% (JS)",
  "color": "brightgreen"
}
EOF

echo -e "${GREEN}✅ Badge data generated: ${COVERAGE_DIR}/coverage-badge.json${NC}\n"

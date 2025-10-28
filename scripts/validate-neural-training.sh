#!/bin/bash

###############################################################################
# Neural Training Validation Script
# Validates Sprint 3 neural training deployment
###############################################################################

set -e

echo "🧪 Neural Training Validation - Sprint 3"
echo "=========================================="
echo ""

# Configuration
KUBECTL="${KUBECTL:-kubectl}"
VALIDATION_RESULTS=()
FAILED=0
PASSED=0

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0;33m' # No Color

# Functions
log_pass() { echo -e "${GREEN}[PASS]${NC} $1"; PASSED=$((PASSED+1)); VALIDATION_RESULTS+=("PASS: $1"); }
log_fail() { echo -e "${RED}[FAIL]${NC} $1"; FAILED=$((FAILED+1)); VALIDATION_RESULTS+=("FAIL: $1"); }
log_info() { echo -e "${YELLOW}[INFO]${NC} $1"; }

# Test 1: Component Health Checks
log_info "Test 1: Component Health Checks"

SERVICES=("agent-lstm-predictor:8087" "neural-pattern-trainer:8088" "historical-data-aggregator:8089" "unified-neural-orchestrator:8090")

for service_port in "${SERVICES[@]}"; do
    SERVICE=$(echo $service_port | cut -d':' -f1)
    PORT=$(echo $service_port | cut -d':' -f2)

    if $KUBECTL get svc $SERVICE &> /dev/null; then
        log_pass "$SERVICE service exists"
    else
        log_fail "$SERVICE service not found"
    fi

    POD=$($KUBECTL get pods -l app=$SERVICE -o name | head -1 | cut -d'/' -f2)
    if [ -n "$POD" ]; then
        if $KUBECTL get pod $POD | grep -q Running; then
            log_pass "$SERVICE pod is running"
        else
            log_fail "$SERVICE pod not running"
        fi
    else
        log_fail "$SERVICE pod not found"
    fi
done

# Test 2: Data Flow Validation
log_info "Test 2: Data Flow Validation"

# Check Redis connectivity
if $KUBECTL get pods -n ollamamax-redis | grep -q redis-cluster; then
    log_pass "Redis cluster accessible"
else
    log_fail "Redis cluster not accessible"
fi

# Check Claude-Flow memory exists
if [ -d ".claude-flow/memory" ]; then
    log_pass "Claude-Flow memory directory exists"
else
    log_fail "Claude-Flow memory directory not found"
fi

# Test 3: Model Training Validation
log_info "Test 3: Model Training Validation"

# Test agent LSTM predictor
log_info "  Testing Agent LSTM Predictor..."
if node src/agents/agent-lstm-predictor.js test &> /tmp/lstm-test.log; then
    log_pass "Agent LSTM Predictor test passed"
else
    log_fail "Agent LSTM Predictor test failed (see /tmp/lstm-test.log)"
fi

# Test neural pattern trainer
log_info "  Testing Neural Pattern Trainer..."
if node src/agents/neural-pattern-trainer.js predict &> /tmp/pattern-test.log; then
    log_pass "Neural Pattern Trainer test passed"
else
    log_fail "Neural Pattern Trainer test failed (see /tmp/pattern-test.log)"
fi

# Test 4: Prediction Validation
log_info "Test 4: Prediction Validation"

# Measure prediction latency
START=$(date +%s%3N)
node src/agents/agent-lstm-predictor.js predict test-agent &> /dev/null || true
END=$(date +%s%3N)
LATENCY=$((END - START))

if [ $LATENCY -lt 500 ]; then
    log_pass "Prediction latency ${LATENCY}ms (<500ms threshold)"
else
    log_fail "Prediction latency ${LATENCY}ms (>500ms threshold)"
fi

# Test 5: Integration Validation
log_info "Test 5: Integration Validation"

# Run full integration test
if node tests/ml/test-neural-training.js &> /tmp/neural-integration.log; then
    log_pass "Integration tests passed"
else
    log_fail "Integration tests failed (see /tmp/neural-integration.log)"
fi

# Test 6: Performance Validation
log_info "Test 6: Performance Validation"

# Check memory usage
POD=$($KUBECTL get pods -l app=unified-neural-orchestrator -o name | head -1 | cut -d'/' -f2)
if [ -n "$POD" ]; then
    MEMORY_USAGE=$($KUBECTL top pod $POD 2>/dev/null | tail -1 | awk '{print $3}' | sed 's/Mi//')
    if [ -n "$MEMORY_USAGE" ] && [ "$MEMORY_USAGE" -lt 4096 ]; then
        log_pass "Memory usage ${MEMORY_USAGE}Mi (<4096Mi threshold)"
    else
        log_fail "Memory usage ${MEMORY_USAGE}Mi (>4096Mi threshold or unavailable)"
    fi
else
    log_fail "Cannot measure memory usage - pod not found"
fi

# Generate Validation Report
log_info "Generating Validation Report..."

cat > neural-training-validation-report.json <<EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "sprint": "Sprint 3",
  "component": "Neural Training & AI Optimization",
  "summary": {
    "total_tests": $((PASSED + FAILED)),
    "passed": $PASSED,
    "failed": $FAILED,
    "success_rate": $(echo "scale=2; $PASSED * 100 / ($PASSED + $FAILED)" | bc)
  },
  "results": [
$(printf '    "%s"' "${VALIDATION_RESULTS[0]}")
$(for result in "${VALIDATION_RESULTS[@]:1}"; do
    printf ',\n    "%s"' "$result"
done)
  ]
}
EOF

echo ""
echo "=========================================="
echo "📊 Validation Summary"
echo "=========================================="
echo "Total Tests: $((PASSED + FAILED))"
echo "✅ Passed: $PASSED"
echo "❌ Failed: $FAILED"
echo "Success Rate: $(echo "scale=1; $PASSED * 100 / ($PASSED + $FAILED)" | bc)%"
echo ""
echo "📝 Detailed report: neural-training-validation-report.json"
echo ""

if [ $FAILED -eq 0 ]; then
    echo "✅ All validations passed!"
    echo "=========================================="
    exit 0
else
    echo "❌ Some validations failed. Review logs:"
    echo "  - /tmp/lstm-test.log"
    echo "  - /tmp/pattern-test.log"
    echo "  - /tmp/neural-integration.log"
    echo "=========================================="
    exit 1
fi

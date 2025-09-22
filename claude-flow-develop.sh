#!/bin/bash
# Complete development workflow with Claude Max subscription

set -e

echo "🚀 Starting OllamaMax Development with Claude Max Subscription"
echo "📌 Using Claude Opus 4.1 for complex tasks (via Claude Max)"
echo "📌 Using Claude Sonnet 4 for rapid iterations (via Claude Max)"
echo "🔑 No API key required - using Claude Max subscription"

# Configure Claude Max authentication
export CLAUDE_AUTH="claude-max"
export CLAUDE_MODEL_PRIMARY="claude-opus-4.1"
export CLAUDE_MODEL_SECONDARY="claude-sonnet-4"
export API_KEY_REQUIRED="false"

# Step 1: Analysis with Claude Opus 4.1 via Claude Max
echo "📊 Phase 1: Analysis with Claude Opus 4.1 (Claude Max)"
claude-flow workflow analyze \
  --auth claude-max \
  --ai-model claude-opus-4.1 \
  --config .claude-flow/workflow.yaml \
  --phase analysis \
  --deep-scan true

# Step 2: Backend development with Claude Opus 4.1 via Claude Max
echo "⚙️ Phase 2: Backend Development with Claude Opus 4.1 (Claude Max)"
claude-flow backend develop \
  --auth claude-max \
  --ai-model claude-opus-4.1 \
  --architecture microservices \
  --api-design restful \
  --test-driven true

# Step 3: Frontend development with Claude Opus 4.1 via Claude Max
echo "🎨 Phase 3: Frontend Development with Claude Opus 4.1 (Claude Max)"
claude-flow frontend develop \
  --auth claude-max \
  --ai-model claude-opus-4.1 \
  --framework angular \
  --responsive-design mobile-first

# Step 4: Playwright testing with Claude Sonnet 4 via Claude Max
echo "🎭 Phase 4: Playwright Testing with Claude Sonnet 4 (Claude Max)"
claude-flow playwright generate-tests \
  --auth claude-max \
  --ai-model claude-sonnet-4 \
  --test-types all \
  --coverage-target 95

claude-flow test playwright-full \
  --auth claude-max \
  --ai-model claude-sonnet-4 \
  --workers 10 \
  --retries 2

# Step 5: Serena AI testing with Claude Max
echo "🤖 Phase 5: Serena Testing with Claude Models (Claude Max)"
claude-flow serena init \
  --auth claude-max \
  --ai-model "claude-opus-4.1,claude-sonnet-4" \
  --primary claude-opus-4.1 \
  --learning-enabled true

claude-flow serena test \
  --auth claude-max \
  --ai-model claude-opus-4.1 \
  --mode comprehensive \
  --iterations 100

# Step 6: Code review with Claude Sonnet 4 via Claude Max
echo "👁️ Phase 6: Code Review with Claude Sonnet 4 (Claude Max)"
claude-flow review \
  --auth claude-max \
  --ai-model claude-sonnet-4 \
  --scope all \
  --fix-issues auto

# Step 7: Documentation with Claude Opus 4.1 via Claude Max
echo "📚 Phase 7: Documentation with Claude Opus 4.1 (Claude Max)"
claude-flow docs generate \
  --auth claude-max \
  --ai-model claude-opus-4.1 \
  --api-docs openapi \
  --user-guide comprehensive

# Step 8: Final validation with Claude Opus 4.1 via Claude Max
echo "✅ Final Validation with Claude Opus 4.1 (Claude Max)"
claude-flow validate \
  --auth claude-max \
  --ai-model claude-opus-4.1 \
  --checklist production \
  --all-tests true

echo "✨ Development complete with Claude Max subscription!"
#!/usr/bin/env node

/**
 * Coverage Validation Script
 *
 * Validates that test coverage meets the required 90% threshold
 * across all metrics (lines, functions, branches, statements).
 *
 * Exit codes:
 *   0 - Coverage meets threshold
 *   1 - Coverage below threshold or validation error
 */

const fs = require('fs');
const path = require('path');

// Configuration
const COVERAGE_THRESHOLD = 90;
const COVERAGE_SUMMARY_PATH = path.join(process.cwd(), 'coverage', 'coverage-summary.json');

/**
 * Main validation function
 */
function validateCoverage() {
  console.log('🔍 Validating test coverage...\n');

  // Check if coverage summary exists
  if (!fs.existsSync(COVERAGE_SUMMARY_PATH)) {
    console.error('❌ Coverage summary not found:', COVERAGE_SUMMARY_PATH);
    console.error('   Run "npm run test:coverage" first to generate coverage data.\n');
    process.exit(1);
  }

  let coverageData;
  try {
    const coverageJson = fs.readFileSync(COVERAGE_SUMMARY_PATH, 'utf8');
    coverageData = JSON.parse(coverageJson);
  } catch (error) {
    console.error('❌ Failed to parse coverage summary:', error.message);
    process.exit(1);
  }

  // Extract total coverage metrics
  const total = coverageData.total;
  if (!total) {
    console.error('❌ Invalid coverage data: missing "total" section');
    process.exit(1);
  }

  const metrics = {
    lines: total.lines?.pct ?? 0,
    statements: total.statements?.pct ?? 0,
    functions: total.functions?.pct ?? 0,
    branches: total.branches?.pct ?? 0
  };

  // Display current coverage
  console.log('📊 Coverage Metrics:');
  console.log('  Lines:      ', formatMetric(metrics.lines));
  console.log('  Statements: ', formatMetric(metrics.statements));
  console.log('  Functions:  ', formatMetric(metrics.functions));
  console.log('  Branches:   ', formatMetric(metrics.branches));
  console.log('');

  // Validate each metric
  const failures = [];
  Object.entries(metrics).forEach(([name, value]) => {
    if (value < COVERAGE_THRESHOLD) {
      failures.push({ name, value, threshold: COVERAGE_THRESHOLD });
    }
  });

  if (failures.length > 0) {
    console.error(`❌ Coverage validation FAILED - ${failures.length} metric(s) below ${COVERAGE_THRESHOLD}% threshold:\n`);
    failures.forEach(({ name, value }) => {
      console.error(`   ${name.padEnd(12)}: ${value.toFixed(2)}% (needs ${COVERAGE_THRESHOLD}%)`);
    });
    console.error('');
    process.exit(1);
  }

  console.log(`✅ Coverage validation PASSED - All metrics meet ${COVERAGE_THRESHOLD}% threshold!\n`);
  process.exit(0);
}

/**
 * Format metric for display
 */
function formatMetric(value) {
  const formatted = value.toFixed(2) + '%';
  if (value >= COVERAGE_THRESHOLD) {
    return `✅ ${formatted}`;
  } else {
    return `❌ ${formatted} (below ${COVERAGE_THRESHOLD}%)`;
  }
}

// Run validation
validateCoverage();

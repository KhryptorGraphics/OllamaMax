/**
 * Load Testing Suite for OllamaMax
 * Tests system performance under various load conditions
 */

const autocannon = require('autocannon');
const chalk = require('chalk');
const fs = require('fs');
const path = require('path');

const API_BASE = process.env.API_BASE || 'http://localhost:13000';
const RESULTS_DIR = './load-test-results';

// Ensure results directory exists
if (!fs.existsSync(RESULTS_DIR)) {
  fs.mkdirSync(RESULTS_DIR, { recursive: true });
}

/**
 * Test configurations
 */
const tests = {
  // Light load - normal usage
  light: {
    duration: 30,
    connections: 10,
    pipelining: 1,
    name: 'Light Load Test'
  },
  
  // Medium load - busy period
  medium: {
    duration: 60,
    connections: 50,
    pipelining: 5,
    name: 'Medium Load Test'
  },
  
  // Heavy load - peak traffic
  heavy: {
    duration: 60,
    connections: 100,
    pipelining: 10,
    name: 'Heavy Load Test'
  },
  
  // Stress test - beyond capacity
  stress: {
    duration: 120,
    connections: 200,
    pipelining: 20,
    name: 'Stress Test'
  },
  
  // Spike test - sudden traffic increase
  spike: {
    duration: 30,
    connections: 500,
    pipelining: 50,
    name: 'Spike Test'
  }
};

/**
 * Test scenarios
 */
const scenarios = {
  health: {
    url: `${API_BASE}/health`,
    method: 'GET',
    name: 'Health Check'
  },
  
  models: {
    url: `${API_BASE}/v1/models`,
    method: 'GET',
    name: 'List Models'
  },
  
  nodes: {
    url: `${API_BASE}/api/nodes`,
    method: 'GET',
    name: 'List Nodes'
  },
  
  completion: {
    url: `${API_BASE}/v1/completions`,
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer test-token'
    },
    body: JSON.stringify({
      model: 'llama-3.2-3b',
      prompt: 'Hello, world!',
      max_tokens: 50
    }),
    name: 'Text Completion'
  },
  
  chat: {
    url: `${API_BASE}/v1/chat/completions`,
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer test-token'
    },
    body: JSON.stringify({
      model: 'llama-3.2-3b',
      messages: [
        { role: 'user', content: 'Hello!' }
      ],
      max_tokens: 50
    }),
    name: 'Chat Completion'
  }
};

/**
 * Run a single load test
 */
async function runTest(testConfig, scenario) {
  console.log(chalk.blue('\n═══════════════════════════════════════════════════════════'));
  console.log(chalk.blue(`  ${testConfig.name} - ${scenario.name}`));
  console.log(chalk.blue('═══════════════════════════════════════════════════════════\n'));
  
  const config = {
    url: scenario.url,
    method: scenario.method,
    headers: scenario.headers,
    body: scenario.body,
    connections: testConfig.connections,
    pipelining: testConfig.pipelining,
    duration: testConfig.duration
  };
  
  return new Promise((resolve, reject) => {
    const instance = autocannon(config, (err, result) => {
      if (err) {
        reject(err);
      } else {
        resolve(result);
      }
    });
    
    // Track progress
    autocannon.track(instance, {
      renderProgressBar: true,
      renderResultsTable: true
    });
  });
}

/**
 * Analyze results
 */
function analyzeResults(result, testName, scenarioName) {
  const analysis = {
    test: testName,
    scenario: scenarioName,
    timestamp: new Date().toISOString(),
    summary: {
      duration: result.duration,
      connections: result.connections,
      pipelining: result.pipelining,
      requests: {
        total: result.requests.total,
        average: result.requests.average,
        mean: result.requests.mean,
        stddev: result.requests.stddev,
        min: result.requests.min,
        max: result.requests.max
      },
      latency: {
        mean: result.latency.mean,
        stddev: result.latency.stddev,
        min: result.latency.min,
        max: result.latency.max,
        p50: result.latency.p50,
        p75: result.latency.p75,
        p90: result.latency.p90,
        p99: result.latency.p99,
        p999: result.latency.p999
      },
      throughput: {
        mean: result.throughput.mean,
        stddev: result.throughput.stddev,
        min: result.throughput.min,
        max: result.throughput.max
      },
      errors: result.errors,
      timeouts: result.timeouts,
      non2xx: result.non2xx || 0
    },
    performance: {
      requestsPerSecond: result.requests.average,
      avgLatencyMs: result.latency.mean,
      p99LatencyMs: result.latency.p99,
      errorRate: ((result.errors + result.timeouts + (result.non2xx || 0)) / result.requests.total * 100).toFixed(2) + '%',
      throughputMBps: (result.throughput.mean / 1024 / 1024).toFixed(2)
    },
    grade: calculateGrade(result)
  };

  return analysis;
}

/**
 * Calculate performance grade
 */
function calculateGrade(result) {
  let score = 100;

  // Deduct points for high latency
  if (result.latency.p99 > 1000) score -= 30;
  else if (result.latency.p99 > 500) score -= 20;
  else if (result.latency.p99 > 200) score -= 10;

  // Deduct points for errors
  const errorRate = (result.errors + result.timeouts + (result.non2xx || 0)) / result.requests.total;
  if (errorRate > 0.05) score -= 40;
  else if (errorRate > 0.01) score -= 20;
  else if (errorRate > 0.001) score -= 10;

  // Deduct points for low throughput
  if (result.requests.average < 100) score -= 20;
  else if (result.requests.average < 500) score -= 10;

  if (score >= 90) return { grade: 'A', color: 'green' };
  if (score >= 80) return { grade: 'B', color: 'blue' };
  if (score >= 70) return { grade: 'C', color: 'yellow' };
  if (score >= 60) return { grade: 'D', color: 'orange' };
  return { grade: 'F', color: 'red' };
}

/**
 * Display results
 */
function displayResults(analysis) {
  console.log(chalk.blue('\n═══════════════════════════════════════════════════════════'));
  console.log(chalk.blue('  Performance Analysis'));
  console.log(chalk.blue('═══════════════════════════════════════════════════════════\n'));

  console.log(chalk.bold('Performance Metrics:'));
  console.log(`  Requests/sec:     ${chalk.cyan(analysis.performance.requestsPerSecond.toFixed(2))}`);
  console.log(`  Avg Latency:      ${chalk.cyan(analysis.performance.avgLatencyMs.toFixed(2))} ms`);
  console.log(`  P99 Latency:      ${chalk.cyan(analysis.performance.p99LatencyMs.toFixed(2))} ms`);
  console.log(`  Error Rate:       ${chalk.cyan(analysis.performance.errorRate)}`);
  console.log(`  Throughput:       ${chalk.cyan(analysis.performance.throughputMBps)} MB/s`);

  const gradeColor = analysis.grade.color;
  console.log(`\n  Performance Grade: ${chalk[gradeColor].bold(analysis.grade.grade)}`);

  // Save results
  const filename = `${analysis.test.replace(/\s+/g, '-')}_${analysis.scenario.replace(/\s+/g, '-')}_${Date.now()}.json`;
  const filepath = path.join(RESULTS_DIR, filename);
  fs.writeFileSync(filepath, JSON.stringify(analysis, null, 2));
  console.log(chalk.gray(`\n  Results saved to: ${filepath}`));
}

/**
 * Main test runner
 */
async function runLoadTests() {
  console.log(chalk.green.bold('\n🚀 OllamaMax Load Testing Suite\n'));

  const testType = process.argv[2] || 'light';
  const scenarioName = process.argv[3] || 'health';

  if (!tests[testType]) {
    console.error(chalk.red(`Unknown test type: ${testType}`));
    console.log(chalk.yellow('Available tests:'), Object.keys(tests).join(', '));
    process.exit(1);
  }

  if (!scenarios[scenarioName]) {
    console.error(chalk.red(`Unknown scenario: ${scenarioName}`));
    console.log(chalk.yellow('Available scenarios:'), Object.keys(scenarios).join(', '));
    process.exit(1);
  }

  const testConfig = tests[testType];
  const scenario = scenarios[scenarioName];

  console.log(chalk.blue(`Test Type: ${testConfig.name}`));
  console.log(chalk.blue(`Scenario: ${scenario.name}`));
  console.log(chalk.blue(`Duration: ${testConfig.duration}s`));
  console.log(chalk.blue(`Connections: ${testConfig.connections}`));
  console.log(chalk.blue(`Pipelining: ${testConfig.pipelining}\n`));

  try {
    const result = await runTest(testConfig, scenario);
    const analysis = analyzeResults(result, testConfig.name, scenario.name);
    displayResults(analysis);

    // Generate summary report
    generateSummaryReport();

  } catch (error) {
    console.error(chalk.red('Load test failed:'), error);
    process.exit(1);
  }
}

/**
 * Generate summary report from all test results
 */
function generateSummaryReport() {
  const files = fs.readdirSync(RESULTS_DIR).filter(f => f.endsWith('.json'));

  if (files.length === 0) return;

  console.log(chalk.blue('\n═══════════════════════════════════════════════════════════'));
  console.log(chalk.blue('  Test History Summary'));
  console.log(chalk.blue('═══════════════════════════════════════════════════════════\n'));

  const results = files.slice(-5).map(f => {
    const data = JSON.parse(fs.readFileSync(path.join(RESULTS_DIR, f), 'utf8'));
    return {
      test: data.test,
      scenario: data.scenario,
      grade: data.grade.grade,
      rps: data.performance.requestsPerSecond.toFixed(0),
      p99: data.performance.p99LatencyMs.toFixed(0),
      errorRate: data.performance.errorRate
    };
  });

  console.table(results);
}

// Run tests
if (require.main === module) {
  runLoadTests().catch(console.error);
}

module.exports = { runTest, analyzeResults };


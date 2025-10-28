import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter, Gauge } from 'k6/metrics';

// Custom metrics for comprehensive performance tracking
const requestsPerSecond = new Rate('requests_per_second');
const inferenceLatency = new Trend('inference_latency');
const apiErrorsByEndpoint = new Counter('api_errors_by_endpoint');
const concurrentConnections = new Gauge('concurrent_connections');

// Environment configuration for distributed execution
const BASE_URL = __ENV.BASE_URL || 'http://localhost:11434';
const INSTANCE_ID = __ENV.K6_INSTANCE_ID || '0';
const TOTAL_INSTANCES = parseInt(__ENV.K6_TOTAL_INSTANCES || '1');
const TARGET_RPS = parseInt(__ENV.TARGET_RPS || '100000');

// Distributed load configuration - multi-stage ramp pattern
export const options = {
  scenarios: {
    // Scenario 1: Health check endpoints (10% of traffic)
    health_checks: {
      executor: 'ramping-arrival-rate',
      startRate: 0,
      timeUnit: '1s',
      preAllocatedVUs: 100,
      maxVUs: 1000,
      stages: [
        { duration: '10m', target: TARGET_RPS * 0.1 * (1 / TOTAL_INSTANCES) }, // Ramp-up
        { duration: '30m', target: TARGET_RPS * 0.1 * (1 / TOTAL_INSTANCES) }, // Sustained
        { duration: '5m', target: TARGET_RPS * 0.25 * (1 / TOTAL_INSTANCES) },  // Spike
        { duration: '20m', target: TARGET_RPS * 0.25 * (1 / TOTAL_INSTANCES) }, // Peak
        { duration: '10m', target: TARGET_RPS * 0.5 * (1 / TOTAL_INSTANCES) },  // Stress
        { duration: '15m', target: TARGET_RPS * 0.5 * (1 / TOTAL_INSTANCES) },  // Extreme
        { duration: '10m', target: 0 },                                           // Ramp-down
      ],
      exec: 'healthCheck',
      tags: { scenario: 'health_checks', instance: INSTANCE_ID },
    },

    // Scenario 2: API status and metrics (20% of traffic)
    api_status: {
      executor: 'ramping-arrival-rate',
      startRate: 0,
      timeUnit: '1s',
      preAllocatedVUs: 200,
      maxVUs: 2000,
      stages: [
        { duration: '10m', target: TARGET_RPS * 0.2 * (1 / TOTAL_INSTANCES) },
        { duration: '30m', target: TARGET_RPS * 0.2 * (1 / TOTAL_INSTANCES) },
        { duration: '5m', target: TARGET_RPS * 0.5 * (1 / TOTAL_INSTANCES) },
        { duration: '20m', target: TARGET_RPS * 0.5 * (1 / TOTAL_INSTANCES) },
        { duration: '10m', target: TARGET_RPS * 1.0 * (1 / TOTAL_INSTANCES) },
        { duration: '15m', target: TARGET_RPS * 1.0 * (1 / TOTAL_INSTANCES) },
        { duration: '10m', target: 0 },
      ],
      exec: 'apiStatus',
      tags: { scenario: 'api_status', instance: INSTANCE_ID },
    },

    // Scenario 3: Model listing and info (30% of traffic)
    model_operations: {
      executor: 'ramping-arrival-rate',
      startRate: 0,
      timeUnit: '1s',
      preAllocatedVUs: 300,
      maxVUs: 3000,
      stages: [
        { duration: '10m', target: TARGET_RPS * 0.3 * (1 / TOTAL_INSTANCES) },
        { duration: '30m', target: TARGET_RPS * 0.3 * (1 / TOTAL_INSTANCES) },
        { duration: '5m', target: TARGET_RPS * 0.75 * (1 / TOTAL_INSTANCES) },
        { duration: '20m', target: TARGET_RPS * 0.75 * (1 / TOTAL_INSTANCES) },
        { duration: '10m', target: TARGET_RPS * 1.5 * (1 / TOTAL_INSTANCES) },
        { duration: '15m', target: TARGET_RPS * 1.5 * (1 / TOTAL_INSTANCES) },
        { duration: '10m', target: 0 },
      ],
      exec: 'modelOperations',
      tags: { scenario: 'model_operations', instance: INSTANCE_ID },
    },

    // Scenario 4: Inference requests (30% of traffic) - most demanding
    inference_requests: {
      executor: 'ramping-arrival-rate',
      startRate: 0,
      timeUnit: '1s',
      preAllocatedVUs: 300,
      maxVUs: 3000,
      stages: [
        { duration: '10m', target: TARGET_RPS * 0.3 * (1 / TOTAL_INSTANCES) },
        { duration: '30m', target: TARGET_RPS * 0.3 * (1 / TOTAL_INSTANCES) },
        { duration: '5m', target: TARGET_RPS * 0.75 * (1 / TOTAL_INSTANCES) },
        { duration: '20m', target: TARGET_RPS * 0.75 * (1 / TOTAL_INSTANCES) },
        { duration: '10m', target: TARGET_RPS * 1.5 * (1 / TOTAL_INSTANCES) },
        { duration: '15m', target: TARGET_RPS * 1.5 * (1 / TOTAL_INSTANCES) },
        { duration: '10m', target: 0 },
      ],
      exec: 'inferenceRequest',
      tags: { scenario: 'inference_requests', instance: INSTANCE_ID },
    },

    // Scenario 5: Admin operations (10% of traffic)
    admin_operations: {
      executor: 'ramping-arrival-rate',
      startRate: 0,
      timeUnit: '1s',
      preAllocatedVUs: 100,
      maxVUs: 1000,
      stages: [
        { duration: '10m', target: TARGET_RPS * 0.1 * (1 / TOTAL_INSTANCES) },
        { duration: '30m', target: TARGET_RPS * 0.1 * (1 / TOTAL_INSTANCES) },
        { duration: '5m', target: TARGET_RPS * 0.25 * (1 / TOTAL_INSTANCES) },
        { duration: '20m', target: TARGET_RPS * 0.25 * (1 / TOTAL_INSTANCES) },
        { duration: '10m', target: TARGET_RPS * 0.5 * (1 / TOTAL_INSTANCES) },
        { duration: '15m', target: TARGET_RPS * 0.5 * (1 / TOTAL_INSTANCES) },
        { duration: '10m', target: 0 },
      ],
      exec: 'adminOperations',
      tags: { scenario: 'admin_operations', instance: INSTANCE_ID },
    },
  },

  // Performance thresholds for production readiness
  thresholds: {
    // HTTP request duration targets
    'http_req_duration': ['p(95)<500', 'p(99)<1000'], // P95<500ms, P99<1000ms
    'http_req_duration{scenario:health_checks}': ['p(95)<100', 'p(99)<200'],
    'http_req_duration{scenario:api_status}': ['p(95)<300', 'p(99)<500'],
    'http_req_duration{scenario:model_operations}': ['p(95)<500', 'p(99)<1000'],
    'http_req_duration{scenario:inference_requests}': ['p(95)<800', 'p(99)<1500'],
    'http_req_duration{scenario:admin_operations}': ['p(95)<400', 'p(99)<800'],

    // Success rate targets: 99.9% success
    'http_req_failed': ['rate<0.001'], // <0.1% error rate
    'http_req_failed{scenario:health_checks}': ['rate<0.0001'],
    'http_req_failed{scenario:inference_requests}': ['rate<0.002'],

    // Throughput target: 100K+ RPS at peak
    'http_reqs': ['rate>100000'],

    // Custom metric thresholds
    'requests_per_second': ['rate>100000'],
    'inference_latency': ['p(95)<800', 'p(99)<1500'],
    'api_errors_by_endpoint': ['count<100'],
  },

  // Output configuration for real-time monitoring
  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'p(99)', 'p(99.9)'],

  // Export to InfluxDB for real-time dashboards (if configured)
  ext: {
    loadimpact: {
      projectID: __ENV.K6_CLOUD_PROJECT_ID,
      name: `OllamaMax Load Test - Instance ${INSTANCE_ID}`,
    },
  },
};

// Scenario 1: Health check endpoints
export function healthCheck() {
  const res = http.get(`${BASE_URL}/health`);

  check(res, {
    'health check status is 200': (r) => r.status === 200,
    'health check response time < 100ms': (r) => r.timings.duration < 100,
  }) || apiErrorsByEndpoint.add(1, { endpoint: '/health' });

  requestsPerSecond.add(1);
  concurrentConnections.add(1);
}

// Scenario 2: API status and metrics
export function apiStatus() {
  const endpoints = ['/api/status', '/api/metrics', '/api/version'];
  const endpoint = endpoints[Math.floor(Math.random() * endpoints.length)];

  const res = http.get(`${BASE_URL}${endpoint}`);

  check(res, {
    'api status is 200': (r) => r.status === 200,
    'api response time < 300ms': (r) => r.timings.duration < 300,
    'api response has valid JSON': (r) => {
      try {
        JSON.parse(r.body);
        return true;
      } catch {
        return false;
      }
    },
  }) || apiErrorsByEndpoint.add(1, { endpoint });

  requestsPerSecond.add(1);
  concurrentConnections.add(1);
}

// Scenario 3: Model operations
export function modelOperations() {
  const operations = [
    () => http.get(`${BASE_URL}/api/tags`), // List models
    () => http.post(`${BASE_URL}/api/show`, JSON.stringify({ name: 'llama2' }), {
      headers: { 'Content-Type': 'application/json' },
    }), // Show model info
  ];

  const operation = operations[Math.floor(Math.random() * operations.length)];
  const res = operation();

  check(res, {
    'model operation status is 200': (r) => r.status === 200 || r.status === 201,
    'model operation response time < 500ms': (r) => r.timings.duration < 500,
  }) || apiErrorsByEndpoint.add(1, { endpoint: '/api/tags|show' });

  requestsPerSecond.add(1);
  concurrentConnections.add(1);
}

// Scenario 4: Inference requests (most demanding)
export function inferenceRequest() {
  const prompts = [
    'What is the capital of France?',
    'Explain quantum computing in simple terms.',
    'Write a haiku about technology.',
    'What are the benefits of distributed systems?',
    'Summarize the key concepts of machine learning.',
  ];

  const prompt = prompts[Math.floor(Math.random() * prompts.length)];

  const payload = JSON.stringify({
    model: 'llama2',
    prompt: prompt,
    stream: false,
    options: {
      temperature: 0.7,
      num_predict: 100,
    },
  });

  const startTime = new Date().getTime();

  const res = http.post(`${BASE_URL}/api/generate`, payload, {
    headers: { 'Content-Type': 'application/json' },
    timeout: '30s',
  });

  const endTime = new Date().getTime();
  const latency = endTime - startTime;

  inferenceLatency.add(latency);

  check(res, {
    'inference status is 200': (r) => r.status === 200,
    'inference response time < 800ms': (r) => r.timings.duration < 800,
    'inference response has valid structure': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.response !== undefined;
      } catch {
        return false;
      }
    },
  }) || apiErrorsByEndpoint.add(1, { endpoint: '/api/generate' });

  requestsPerSecond.add(1);
  concurrentConnections.add(1);

  sleep(0.1); // Brief pause to simulate think time
}

// Scenario 5: Admin operations
export function adminOperations() {
  const operations = [
    () => http.get(`${BASE_URL}/api/ps`), // List running processes
    () => http.get(`${BASE_URL}/api/version`), // Get version
  ];

  const operation = operations[Math.floor(Math.random() * operations.length)];
  const res = operation();

  check(res, {
    'admin operation status is 200': (r) => r.status === 200,
    'admin operation response time < 400ms': (r) => r.timings.duration < 400,
  }) || apiErrorsByEndpoint.add(1, { endpoint: '/api/admin' });

  requestsPerSecond.add(1);
  concurrentConnections.add(1);
}

// Setup function - runs once per VU
export function setup() {
  console.log(`Starting k6 load test - Instance ${INSTANCE_ID} of ${TOTAL_INSTANCES}`);
  console.log(`Target RPS per instance: ${TARGET_RPS / TOTAL_INSTANCES}`);
  console.log(`Base URL: ${BASE_URL}`);

  // Verify target system is accessible
  const healthCheck = http.get(`${BASE_URL}/health`);
  if (healthCheck.status !== 200) {
    throw new Error(`Target system not healthy: ${healthCheck.status}`);
  }

  return {
    startTime: new Date().toISOString(),
    instanceId: INSTANCE_ID,
    totalInstances: TOTAL_INSTANCES,
    targetRPS: TARGET_RPS,
  };
}

// Teardown function - runs once after test completion
export function teardown(data) {
  console.log(`Load test completed - Instance ${data.instanceId}`);
  console.log(`Started at: ${data.startTime}`);
  console.log(`Ended at: ${new Date().toISOString()}`);
}

// Custom summary handler for detailed results
export function handleSummary(data) {
  const instanceId = INSTANCE_ID;

  return {
    [`load-test-results/summary-instance-${instanceId}.json`]: JSON.stringify(data, null, 2),
    [`load-test-results/summary-instance-${instanceId}.txt`]: textSummary(data, { indent: '  ', enableColors: false }),
    stdout: textSummary(data, { indent: '  ', enableColors: true }),
  };
}

function textSummary(data, options) {
  const indent = options.indent || '';
  const colors = options.enableColors !== false;

  let summary = `\n${indent}Load Test Summary - Instance ${INSTANCE_ID}\n`;
  summary += `${indent}${'='.repeat(60)}\n\n`;

  // Overall metrics
  summary += `${indent}Overall Results:\n`;
  summary += `${indent}  Total Requests: ${data.metrics.http_reqs?.values?.count || 0}\n`;
  summary += `${indent}  Requests/sec: ${data.metrics.http_reqs?.values?.rate?.toFixed(2) || 0}\n`;
  summary += `${indent}  Failed Requests: ${data.metrics.http_req_failed?.values?.rate?.toFixed(4) || 0}%\n`;
  summary += `${indent}  Avg Duration: ${data.metrics.http_req_duration?.values?.avg?.toFixed(2) || 0}ms\n`;
  summary += `${indent}  P95 Duration: ${data.metrics.http_req_duration?.values?.['p(95)']?.toFixed(2) || 0}ms\n`;
  summary += `${indent}  P99 Duration: ${data.metrics.http_req_duration?.values?.['p(99)']?.toFixed(2) || 0}ms\n`;

  return summary;
}

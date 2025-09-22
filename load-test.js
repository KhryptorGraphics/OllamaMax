import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

// Test configuration
export let options = {
  stages: [
    { duration: '5m', target: 100 }, // Ramp up to 100 users over 5 minutes
    { duration: '10m', target: 100 }, // Stay at 100 users for 10 minutes
    { duration: '5m', target: 500 }, // Ramp up to 500 users over 5 minutes
    { duration: '10m', target: 500 }, // Stay at 500 users for 10 minutes
    { duration: '5m', target: 1000 }, // Ramp up to 1000 users over 5 minutes
    { duration: '10m', target: 1000 }, // Stay at 1000 users for 10 minutes
    { duration: '5m', target: 0 }, // Ramp down to 0 users over 5 minutes
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests must complete within 500ms
    http_req_failed: ['rate<0.01'], // Error rate must be less than 1%
    errors: ['rate<0.01'], // Custom error rate must be less than 1%
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  // Health check endpoint
  let healthResponse = http.get(`${BASE_URL}/health`);
  check(healthResponse, {
    'health check status is 200': (r) => r.status === 200,
    'health check response time < 200ms': (r) => r.timings.duration < 200,
  }) || errorRate.add(1);

  // API endpoints testing
  let apiResponse = http.get(`${BASE_URL}/api/v1/status`);
  check(apiResponse, {
    'API status is 200': (r) => r.status === 200,
    'API response time < 500ms': (r) => r.timings.duration < 500,
    'API response has expected fields': (r) => {
      const body = JSON.parse(r.body);
      return body.hasOwnProperty('status') && body.hasOwnProperty('version');
    },
  }) || errorRate.add(1);

  // Models endpoint testing
  let modelsResponse = http.get(`${BASE_URL}/api/v1/models`);
  check(modelsResponse, {
    'Models endpoint status is 200': (r) => r.status === 200,
    'Models response time < 1s': (r) => r.timings.duration < 1000,
  }) || errorRate.add(1);

  // Simulate user interaction with random sleep
  sleep(Math.random() * 3 + 1); // Sleep between 1-4 seconds
}

export function handleSummary(data) {
  return {
    'load-test-report.json': JSON.stringify(data, null, 2),
    'load-test-summary.txt': textSummary(data, { indent: ' ', enableColors: false }),
  };
}
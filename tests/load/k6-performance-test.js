// K6 Load Testing Script for OllamaMax v1 Performance Validation
// Tests compression, database connection pools, and request handling
// Target: Validate v1 baseline performance (not full 100K RPS)

import http from 'k6/http';
import ws from 'k6/ws';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const responseTime = new Trend('response_time');
const compressionRatio = new Trend('compression_ratio');
const dbConnectionErrors = new Counter('db_connection_errors');
const wsConnectionErrors = new Counter('ws_connection_errors');

// Test configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:13100';
const WS_URL = __ENV.WS_URL || 'ws://localhost:13100';

// Load test stages for v1 baseline validation
export const options = {
  stages: [
    // Ramp up to 100 users over 2 minutes
    { duration: '2m', target: 100 },
    
    // Stay at 100 users for 5 minutes (steady state)
    { duration: '5m', target: 100 },
    
    // Ramp up to 500 users over 3 minutes
    { duration: '3m', target: 500 },
    
    // Stay at 500 users for 5 minutes (stress test)
    { duration: '5m', target: 500 },
    
    // Ramp up to 1000 users over 2 minutes (peak load)
    { duration: '2m', target: 1000 },
    
    // Stay at 1000 users for 3 minutes
    { duration: '3m', target: 1000 },
    
    // Ramp down to 0 over 2 minutes
    { duration: '2m', target: 0 },
  ],
  
  thresholds: {
    // Performance thresholds for v1 baseline
    http_req_duration: ['p(95)<2000'], // 95% of requests under 2s
    http_req_failed: ['rate<0.05'],    // Error rate under 5%
    errors: ['rate<0.05'],             // Custom error rate under 5%
    response_time: ['p(90)<1500'],     // 90% under 1.5s
    
    // Database connection thresholds
    db_connection_errors: ['count<100'], // Less than 100 DB errors total
    
    // WebSocket thresholds
    ws_connection_errors: ['count<50'],  // Less than 50 WS errors total
  },
  
  // Resource limits
  maxRedirects: 4,
  userAgent: 'OllamaMax-LoadTest/1.0',
  
  // Test data
  setupTimeout: '60s',
  teardownTimeout: '60s',
};

// Test data generators
const generatePrompt = (size = 'medium') => {
  const prompts = {
    small: 'Hello, how are you?',
    medium: 'Explain the concept of artificial intelligence and its applications in modern technology. Include examples of machine learning, natural language processing, and computer vision.',
    large: 'Write a comprehensive analysis of distributed systems architecture, including microservices patterns, load balancing strategies, database sharding techniques, caching mechanisms, and fault tolerance approaches. Discuss the trade-offs between consistency, availability, and partition tolerance in the context of the CAP theorem.'.repeat(3)
  };
  return prompts[size] || prompts.medium;
};

const generateModelName = () => {
  const models = ['llama2', 'codellama', 'mistral', 'vicuna', 'alpaca'];
  return models[Math.floor(Math.random() * models.length)];
};

// Setup function
export function setup() {
  console.log('Starting OllamaMax v1 Performance Test');
  console.log(`Base URL: ${BASE_URL}`);
  console.log(`WebSocket URL: ${WS_URL}`);
  
  // Verify server is running
  const healthCheck = http.get(`${BASE_URL}/api/health`);
  check(healthCheck, {
    'Health check successful': (r) => r.status === 200,
  });
  
  return {
    baseUrl: BASE_URL,
    wsUrl: WS_URL,
    startTime: Date.now()
  };
}

// Main test function
export default function(data) {
  const testType = Math.random();
  
  if (testType < 0.4) {
    // 40% - API health and status checks
    testHealthEndpoints();
  } else if (testType < 0.7) {
    // 30% - Node and model management
    testNodeManagement();
  } else if (testType < 0.9) {
    // 20% - Metrics and monitoring
    testMetricsEndpoints();
  } else {
    // 10% - WebSocket inference simulation
    testWebSocketInference();
  }
  
  // Random sleep between 1-3 seconds
  sleep(Math.random() * 2 + 1);
}

function testHealthEndpoints() {
  const startTime = Date.now();
  
  // Test basic health endpoint
  const healthResponse = http.get(`${BASE_URL}/api/health`, {
    headers: {
      'Accept': 'application/json',
      'Accept-Encoding': 'gzip, br', // Test compression
    },
  });
  
  const success = check(healthResponse, {
    'Health endpoint status 200': (r) => r.status === 200,
    'Health response has status': (r) => r.json('status') !== undefined,
    'Response time < 500ms': (r) => r.timings.duration < 500,
    'Content-Encoding present': (r) => r.headers['Content-Encoding'] !== undefined,
  });
  
  if (!success) {
    errorRate.add(1);
  } else {
    errorRate.add(0);
  }
  
  // Calculate compression ratio if available
  const contentLength = parseInt(healthResponse.headers['Content-Length'] || '0');
  const uncompressedSize = JSON.stringify(healthResponse.json()).length;
  if (contentLength > 0 && uncompressedSize > 0) {
    compressionRatio.add(uncompressedSize / contentLength);
  }
  
  responseTime.add(Date.now() - startTime);
  
  // Test detailed health endpoint
  const detailedHealthResponse = http.get(`${BASE_URL}/api/health/detailed`, {
    headers: {
      'Accept': 'application/json',
      'Accept-Encoding': 'gzip, br',
    },
  });
  
  check(detailedHealthResponse, {
    'Detailed health status 200': (r) => r.status === 200,
    'Has cluster info': (r) => r.json('cluster') !== undefined,
    'Has system info': (r) => r.json('system') !== undefined,
  });
}

function testNodeManagement() {
  const startTime = Date.now();
  
  // Get nodes list
  const nodesResponse = http.get(`${BASE_URL}/api/nodes`, {
    headers: {
      'Accept': 'application/json',
      'Accept-Encoding': 'gzip, br',
    },
  });
  
  const success = check(nodesResponse, {
    'Nodes endpoint status 200': (r) => r.status === 200,
    'Has nodes array': (r) => Array.isArray(r.json('nodes')),
    'Response time < 1000ms': (r) => r.timings.duration < 1000,
  });
  
  if (!success) {
    errorRate.add(1);
    if (nodesResponse.status >= 500) {
      dbConnectionErrors.add(1);
    }
  } else {
    errorRate.add(0);
  }
  
  responseTime.add(Date.now() - startTime);
  
  // Test routing modes
  const routingResponse = http.get(`${BASE_URL}/api/routing/modes`);
  check(routingResponse, {
    'Routing modes status 200': (r) => r.status === 200,
    'Has modes array': (r) => Array.isArray(r.json('modes')),
  });
}

function testMetricsEndpoints() {
  const startTime = Date.now();
  
  // Test Prometheus metrics
  const metricsResponse = http.get(`${BASE_URL}/api/metrics/prometheus`, {
    headers: {
      'Accept': 'text/plain',
    },
  });
  
  const success = check(metricsResponse, {
    'Metrics endpoint status 200': (r) => r.status === 200,
    'Content-Type is text/plain': (r) => r.headers['Content-Type'].includes('text/plain'),
    'Contains ollama metrics': (r) => r.body.includes('ollama_'),
    'Response time < 2000ms': (r) => r.timings.duration < 2000,
  });
  
  if (!success) {
    errorRate.add(1);
  } else {
    errorRate.add(0);
  }
  
  responseTime.add(Date.now() - startTime);
  
  // Test JSON metrics
  const jsonMetricsResponse = http.get(`${BASE_URL}/api/metrics`);
  check(jsonMetricsResponse, {
    'JSON metrics status 200': (r) => r.status === 200,
    'Has requests metrics': (r) => r.json('requests') !== undefined,
    'Has nodes metrics': (r) => r.json('nodes') !== undefined,
  });
}

function testWebSocketInference() {
  const url = `${WS_URL}/chat`;
  
  const response = ws.connect(url, {}, function (socket) {
    socket.on('open', function open() {
      console.log('WebSocket connected');
      
      // Send inference request
      const inferenceRequest = {
        type: 'inference',
        model: generateModelName(),
        content: generatePrompt('medium'),
        timestamp: Date.now(),
        settings: {
          temperature: 0.7,
          maxTokens: 100,
          streaming: false
        },
        routingMode: 'auto',
        sessionId: `test-session-${Math.random().toString(36).substr(2, 9)}`
      };
      
      socket.send(JSON.stringify(inferenceRequest));
    });
    
    socket.on('message', function (message) {
      const data = JSON.parse(message);
      
      check(data, {
        'WebSocket response received': (d) => d.type !== undefined,
        'No error in response': (d) => d.type !== 'error',
      });
      
      if (data.type === 'error') {
        wsConnectionErrors.add(1);
        errorRate.add(1);
      } else {
        errorRate.add(0);
      }
      
      socket.close();
    });
    
    socket.on('error', function (e) {
      console.log('WebSocket error:', e);
      wsConnectionErrors.add(1);
      errorRate.add(1);
    });
    
    // Timeout after 10 seconds
    socket.setTimeout(function () {
      console.log('WebSocket timeout');
      socket.close();
    }, 10000);
  });
  
  check(response, {
    'WebSocket connection successful': (r) => r && r.status === 101,
  });
}

// Teardown function
export function teardown(data) {
  const duration = (Date.now() - data.startTime) / 1000;
  console.log(`Test completed in ${duration} seconds`);
  
  // Final health check
  const finalHealth = http.get(`${BASE_URL}/api/health`);
  check(finalHealth, {
    'Server still healthy after test': (r) => r.status === 200,
  });
}

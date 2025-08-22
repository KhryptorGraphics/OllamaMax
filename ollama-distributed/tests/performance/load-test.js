import http from 'k6/http';
import ws from 'k6/ws';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('error_rate');
const responseTime = new Trend('response_time');
const wsConnections = new Counter('websocket_connections');
const wsMessages = new Counter('websocket_messages');

// Test configuration
export const options = {
  stages: [
    // Ramp up
    { duration: '2m', target: 50 },   // Ramp up to 50 users over 2 minutes
    { duration: '5m', target: 50 },   // Stay at 50 users for 5 minutes
    { duration: '2m', target: 100 },  // Ramp up to 100 users over 2 minutes
    { duration: '5m', target: 100 },  // Stay at 100 users for 5 minutes
    { duration: '2m', target: 200 },  // Ramp up to 200 users over 2 minutes
    { duration: '5m', target: 200 },  // Stay at 200 users for 5 minutes
    { duration: '5m', target: 0 },    // Ramp down to 0 users over 5 minutes
  ],
  thresholds: {
    http_req_duration: ['p(95)<2000'], // 95% of requests should complete within 2s
    http_req_failed: ['rate<0.01'],    // Error rate should be less than 1%
    error_rate: ['rate<0.01'],         // Custom error rate should be less than 1%
    response_time: ['p(95)<2000'],     // 95% of custom response times should be under 2s
    websocket_connections: ['count>0'], // Should have WebSocket connections
    websocket_messages: ['count>0'],   // Should have WebSocket message traffic
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:3000';
const API_URL = __ENV.API_URL || 'http://localhost:8080';
const WS_URL = __ENV.WS_URL || 'ws://localhost:8080/ws';

// Test data
const testUsers = [
  { email: 'admin@example.com', password: 'admin123' },
  { email: 'user@example.com', password: 'user123' },
  { email: 'test1@example.com', password: 'test123' },
  { email: 'test2@example.com', password: 'test123' },
  { email: 'test3@example.com', password: 'test123' },
];

export function setup() {
  console.log('Starting load test setup...');
  
  // Warm up the application
  const warmupRes = http.get(`${BASE_URL}/health`);
  check(warmupRes, {
    'warmup successful': (r) => r.status === 200,
  });
  
  return { baseURL: BASE_URL, apiURL: API_URL, wsURL: WS_URL };
}

export default function(data) {
  const user = testUsers[Math.floor(Math.random() * testUsers.length)];
  
  // Test scenario selection based on probability
  const scenario = Math.random();
  
  if (scenario < 0.4) {
    // 40% - Dashboard browsing scenario
    dashboardBrowsingScenario(data, user);
  } else if (scenario < 0.7) {
    // 30% - Admin operations scenario
    adminOperationsScenario(data, user);
  } else if (scenario < 0.9) {
    // 20% - Real-time monitoring scenario
    realTimeMonitoringScenario(data, user);
  } else {
    // 10% - Heavy operations scenario
    heavyOperationsScenario(data, user);
  }
}

function dashboardBrowsingScenario(data, user) {
  const startTime = Date.now();
  
  // Login
  const loginRes = http.post(`${data.apiURL}/api/auth/login`, 
    JSON.stringify({
      email: user.email,
      password: user.password
    }), 
    {
      headers: { 'Content-Type': 'application/json' },
    }
  );
  
  const loginSuccess = check(loginRes, {
    'login successful': (r) => r.status === 200,
    'login response time < 1s': (r) => r.timings.duration < 1000,
  });
  
  if (!loginSuccess) {
    errorRate.add(1);
    return;
  }
  
  const authToken = loginRes.json('token');
  const headers = {
    'Authorization': `Bearer ${authToken}`,
    'Content-Type': 'application/json',
  };
  
  // Load dashboard
  const dashboardRes = http.get(`${data.baseURL}/dashboard`, { headers });
  check(dashboardRes, {
    'dashboard loaded': (r) => r.status === 200,
    'dashboard response time < 2s': (r) => r.timings.duration < 2000,
  });
  
  sleep(2); // User reads dashboard
  
  // Get cluster status
  const clusterRes = http.get(`${data.apiURL}/api/cluster/status`, { headers });
  check(clusterRes, {
    'cluster status retrieved': (r) => r.status === 200,
    'cluster status response time < 500ms': (r) => r.timings.duration < 500,
  });
  
  sleep(1);
  
  // Get performance metrics
  const metricsRes = http.get(`${data.apiURL}/api/metrics/performance`, { headers });
  check(metricsRes, {
    'metrics retrieved': (r) => r.status === 200,
    'metrics response time < 1s': (r) => r.timings.duration < 1000,
  });
  
  const totalTime = Date.now() - startTime;
  responseTime.add(totalTime);
  
  sleep(3); // User contemplates data
}

function adminOperationsScenario(data, user) {
  const startTime = Date.now();
  
  // Login as admin
  const loginRes = http.post(`${data.apiURL}/api/auth/login`, 
    JSON.stringify({
      email: 'admin@example.com',
      password: 'admin123'
    }), 
    {
      headers: { 'Content-Type': 'application/json' },
    }
  );
  
  if (!check(loginRes, { 'admin login successful': (r) => r.status === 200 })) {
    errorRate.add(1);
    return;
  }
  
  const authToken = loginRes.json('token');
  const headers = {
    'Authorization': `Bearer ${authToken}`,
    'Content-Type': 'application/json',
  };
  
  // Access admin panel
  const adminRes = http.get(`${data.baseURL}/admin`, { headers });
  check(adminRes, {
    'admin panel accessible': (r) => r.status === 200,
  });
  
  sleep(1);
  
  // Get node list
  const nodesRes = http.get(`${data.apiURL}/api/admin/nodes`, { headers });
  check(nodesRes, {
    'nodes list retrieved': (r) => r.status === 200,
    'nodes response time < 1s': (r) => r.timings.duration < 1000,
  });
  
  sleep(2);
  
  // Get system health
  const healthRes = http.get(`${data.apiURL}/api/admin/health`, { headers });
  check(healthRes, {
    'system health retrieved': (r) => r.status === 200,
  });
  
  sleep(1);
  
  // Update configuration (simulate)
  const configRes = http.put(`${data.apiURL}/api/admin/config`, 
    JSON.stringify({
      setting: 'max_connections',
      value: 1000
    }), 
    { headers }
  );
  check(configRes, {
    'config updated': (r) => r.status === 200,
  });
  
  const totalTime = Date.now() - startTime;
  responseTime.add(totalTime);
  
  sleep(2);
}

function realTimeMonitoringScenario(data, user) {
  const startTime = Date.now();
  
  // Login
  const loginRes = http.post(`${data.apiURL}/api/auth/login`, 
    JSON.stringify({
      email: user.email,
      password: user.password
    }), 
    {
      headers: { 'Content-Type': 'application/json' },
    }
  );
  
  if (!check(loginRes, { 'login successful': (r) => r.status === 200 })) {
    errorRate.add(1);
    return;
  }
  
  const authToken = loginRes.json('token');
  
  // WebSocket connection for real-time updates
  const wsRes = ws.connect(`${data.wsURL}?token=${authToken}`, {}, function (socket) {
    wsConnections.add(1);
    
    socket.on('open', function open() {
      console.log('WebSocket connected');
      
      // Subscribe to real-time metrics
      socket.send(JSON.stringify({
        type: 'subscribe',
        channels: ['metrics', 'cluster-status', 'alerts']
      }));
      wsMessages.add(1);
    });
    
    socket.on('message', function message(data) {
      const message = JSON.parse(data);
      wsMessages.add(1);
      
      check(message, {
        'websocket message valid': (msg) => msg.type !== undefined,
      });
    });
    
    socket.on('error', function error(e) {
      console.log('WebSocket error:', e);
      errorRate.add(1);
    });
    
    // Keep connection alive for realistic duration
    sleep(10);
    
    socket.close();
  });
  
  check(wsRes, {
    'websocket connection established': (r) => r && r.status === 101,
  });
  
  const totalTime = Date.now() - startTime;
  responseTime.add(totalTime);
}

function heavyOperationsScenario(data, user) {
  const startTime = Date.now();
  
  // Login as admin
  const loginRes = http.post(`${data.apiURL}/api/auth/login`, 
    JSON.stringify({
      email: 'admin@example.com',
      password: 'admin123'
    }), 
    {
      headers: { 'Content-Type': 'application/json' },
    }
  );
  
  if (!check(loginRes, { 'admin login successful': (r) => r.status === 200 })) {
    errorRate.add(1);
    return;
  }
  
  const authToken = loginRes.json('token');
  const headers = {
    'Authorization': `Bearer ${authToken}`,
    'Content-Type': 'application/json',
  };
  
  // Simulate heavy model operation
  const modelRes = http.post(`${data.apiURL}/api/models/download`, 
    JSON.stringify({
      model: 'llama2:7b',
      force: false
    }), 
    { 
      headers,
      timeout: '30s' // Heavy operations might take longer
    }
  );
  
  check(modelRes, {
    'model operation initiated': (r) => r.status === 200 || r.status === 202,
    'model operation response time < 30s': (r) => r.timings.duration < 30000,
  });
  
  sleep(5);
  
  // Get detailed cluster analysis
  const analysisRes = http.get(`${data.apiURL}/api/analysis/cluster?detailed=true`, { 
    headers,
    timeout: '15s'
  });
  
  check(analysisRes, {
    'cluster analysis completed': (r) => r.status === 200,
    'analysis response time < 15s': (r) => r.timings.duration < 15000,
  });
  
  sleep(3);
  
  // Generate performance report
  const reportRes = http.post(`${data.apiURL}/api/reports/performance`, 
    JSON.stringify({
      timeRange: '24h',
      includeDetails: true
    }), 
    { 
      headers,
      timeout: '20s'
    }
  );
  
  check(reportRes, {
    'performance report generated': (r) => r.status === 200,
    'report generation time < 20s': (r) => r.timings.duration < 20000,
  });
  
  const totalTime = Date.now() - startTime;
  responseTime.add(totalTime);
  
  sleep(5);
}

export function teardown(data) {
  console.log('Load test completed');
  
  // Optional: Generate summary report
  const summaryRes = http.get(`${data.apiURL}/api/health`);
  check(summaryRes, {
    'application still responsive': (r) => r.status === 200,
  });
}
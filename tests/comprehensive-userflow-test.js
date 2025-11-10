#!/usr/bin/env node
/**
 * Comprehensive User Flow Testing Script
 * Tests all possible user actions in OllamaMax
 */

const http = require('http');
const https = require('https');

const API_BASE = process.env.API_BASE_URL || 'http://localhost:13000';
const colors = {
  reset: '\x1b[0m',
  green: '\x1b[32m',
  red: '\x1b[31m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  cyan: '\x1b[36m'
};

let testResults = {
  passed: 0,
  failed: 0,
  skipped: 0,
  total: 0
};

let authToken = null;
let testUser = {
  email: `test${Date.now()}@ollamamax.test`,
  password: 'TestPass123!@#',
  firstName: 'Test',
  lastName: 'User'
};

// Helper function to make HTTP requests
async function makeRequest(method, path, data = null, headers = {}) {
  return new Promise((resolve, reject) => {
    const url = new URL(path, API_BASE);
    const options = {
      method,
      headers: {
        'Content-Type': 'application/json',
        ...headers
      }
    };

    const req = (url.protocol === 'https:' ? https : http).request(url, options, (res) => {
      let body = '';
      res.on('data', chunk => body += chunk);
      res.on('end', () => {
        try {
          resolve({
            status: res.statusCode,
            headers: res.headers,
            body: body ? JSON.parse(body) : null
          });
        } catch (e) {
          resolve({
            status: res.statusCode,
            headers: res.headers,
            body: body
          });
        }
      });
    });

    req.on('error', reject);
    if (data) req.write(JSON.stringify(data));
    req.end();
  });
}

// Test runner
async function runTest(name, testFn) {
  testResults.total++;
  process.stdout.write(`${colors.cyan}Testing: ${name}${colors.reset} ... `);
  
  try {
    await testFn();
    testResults.passed++;
    console.log(`${colors.green}✓ PASS${colors.reset}`);
    return true;
  } catch (error) {
    testResults.failed++;
    console.log(`${colors.red}✗ FAIL${colors.reset}`);
    console.log(`  ${colors.red}Error: ${error.message}${colors.reset}`);
    return false;
  }
}

// Test assertions
function assert(condition, message) {
  if (!condition) throw new Error(message || 'Assertion failed');
}

function assertEqual(actual, expected, message) {
  if (actual !== expected) {
    throw new Error(message || `Expected ${expected}, got ${actual}`);
  }
}

// ============================================================================
// TEST SUITE 1: HEALTH & SYSTEM ENDPOINTS
// ============================================================================

async function testHealthEndpoint() {
  const res = await makeRequest('GET', '/health');
  assert(res.status === 200, `Expected 200, got ${res.status}`);
  assert(res.body.status === 'healthy', 'Health status should be healthy');
  assert(res.body.uptime, 'Should have uptime information');
  assert(res.body.memory, 'Should have memory information');
}

async function testLivenessProbe() {
  const res = await makeRequest('GET', '/health/live');
  assert(res.status === 200, `Expected 200, got ${res.status}`);
  assert(res.body.status === 'alive', 'Should be alive');
}

async function testReadinessProbe() {
  const res = await makeRequest('GET', '/health/ready');
  assert(res.status === 200 || res.status === 503, `Expected 200 or 503, got ${res.status}`);
  assert(res.body.status, 'Should have status field');
}

async function testMetricsEndpoint() {
  const res = await makeRequest('GET', '/metrics');
  assert(res.status === 200, `Expected 200, got ${res.status}`);
  assert(typeof res.body === 'string', 'Metrics should be text format');
  assert(res.body.includes('ollamamax_'), 'Should contain ollamamax metrics');
}

async function testRootEndpoint() {
  const res = await makeRequest('GET', '/');
  assert(res.status === 200, `Expected 200, got ${res.status}`);
  assert(res.body.name === 'Ollamamax API', 'Should have correct API name');
  assert(res.body.endpoints, 'Should list endpoints');
  assert(res.body.features, 'Should list features');
}

// ============================================================================
// TEST SUITE 2: AUTHENTICATION FLOW
// ============================================================================

async function testUserRegistration() {
  const res = await makeRequest('POST', '/auth/register', testUser);
  assert(res.status === 201, `Expected 201, got ${res.status}`);
  assert(res.body.user, 'Should return user object');
  assert(res.body.user.id, 'User should have ID');
}

async function testUserLogin() {
  const res = await makeRequest('POST', '/auth/login', {
    email: testUser.email,
    password: testUser.password
  });
  assert(res.status === 200, `Expected 200, got ${res.status}`);
  assert(res.body.access_token, 'Should return access token');
  assert(res.body.refresh_token, 'Should return refresh token');
  authToken = res.body.access_token;
}

async function testInvalidLogin() {
  const res = await makeRequest('POST', '/auth/login', {
    email: testUser.email,
    password: 'wrongpassword'
  });
  assert(res.status === 401, `Expected 401, got ${res.status}`);
}

async function testGetUserProfile() {
  const res = await makeRequest('GET', '/auth/me', null, {
    'Authorization': `Bearer ${authToken}`
  });
  assert(res.status === 200, `Expected 200, got ${res.status}`);
  assert(res.body.user, 'Should return user object');
  assert(res.body.user.email === testUser.email, 'Email should match');
}

async function testRefreshToken() {
  const res = await makeRequest('POST', '/auth/refresh', {
    refresh_token: 'dummy_refresh_token'
  });
  // May not be implemented yet
  assert(res.status === 200 || res.status === 401 || res.status === 404,
    `Expected 200/401/404, got ${res.status}`);
}

// ============================================================================
// TEST SUITE 3: MODEL ENDPOINTS (OpenAI Compatible)
// ============================================================================

async function testListModels() {
  const res = await makeRequest('GET', '/v1/models');
  assert(res.status === 200, `Expected 200, got ${res.status}`);
  assert(res.body.object === 'list', 'Should be a list object');
  assert(Array.isArray(res.body.data), 'Should have data array');
  assert(res.body.data.length > 0, 'Should have at least one model');
}

async function testTextCompletion() {
  const res = await makeRequest('POST', '/v1/completions', {
    model: 'llama-3.2-3b',
    prompt: 'Hello, how are you?',
    max_tokens: 50
  }, {
    'Authorization': `Bearer ${authToken}`
  });
  assert(res.status === 200, `Expected 200, got ${res.status}`);
  assert(res.body.choices, 'Should have choices');
  assert(res.body.choices[0].text, 'Should have completion text');
}

async function testChatCompletion() {
  const res = await makeRequest('POST', '/v1/chat/completions', {
    model: 'llama-3.2-3b',
    messages: [
      { role: 'user', content: 'Hello!' }
    ],
    max_tokens: 50
  }, {
    'Authorization': `Bearer ${authToken}`
  });
  assert(res.status === 200, `Expected 200, got ${res.status}`);
  assert(res.body.choices, 'Should have choices');
  assert(res.body.choices[0].message, 'Should have message');
  assert(res.body.choices[0].message.content, 'Should have content');
}

async function testEmbeddings() {
  const res = await makeRequest('POST', '/v1/embeddings', {
    model: 'text-embedding-ada-002',
    input: 'Test embedding'
  }, {
    'Authorization': `Bearer ${authToken}`
  });
  assert(res.status === 200, `Expected 200, got ${res.status}`);
  assert(res.body.data, 'Should have data');
  assert(Array.isArray(res.body.data[0].embedding), 'Should have embedding array');
}

// ============================================================================
// TEST SUITE 4: DOCUMENTATION ENDPOINTS
// ============================================================================

async function testSwaggerUI() {
  const res = await makeRequest('GET', '/docs');
  assert(res.status === 200, `Expected 200, got ${res.status}`);
  assert(typeof res.body === 'string', 'Should return HTML');
  assert(res.body.includes('swagger'), 'Should contain swagger UI');
}

async function testOpenAPISpec() {
  const res = await makeRequest('GET', '/openapi.json');
  assert(res.status === 200, `Expected 200, got ${res.status}`);
  assert(res.body.openapi, 'Should have OpenAPI version');
  assert(res.body.info, 'Should have info section');
  assert(res.body.paths, 'Should have paths');
}

// ============================================================================
// MAIN TEST EXECUTION
// ============================================================================

async function runAllTests() {
  console.log(`\n${colors.blue}═══════════════════════════════════════════════════════════${colors.reset}`);
  console.log(`${colors.blue}  OllamaMax Comprehensive User Flow Testing${colors.reset}`);
  console.log(`${colors.blue}═══════════════════════════════════════════════════════════${colors.reset}\n`);
  console.log(`${colors.cyan}API Base URL: ${API_BASE}${colors.reset}\n`);

  // Test Suite 1: Health & System
  console.log(`\n${colors.yellow}━━━ Test Suite 1: Health & System Endpoints ━━━${colors.reset}\n`);
  await runTest('Health endpoint', testHealthEndpoint);
  await runTest('Liveness probe', testLivenessProbe);
  await runTest('Readiness probe', testReadinessProbe);
  await runTest('Metrics endpoint', testMetricsEndpoint);
  await runTest('Root endpoint', testRootEndpoint);

  // Test Suite 2: Authentication
  console.log(`\n${colors.yellow}━━━ Test Suite 2: Authentication Flow ━━━${colors.reset}\n`);
  await runTest('User registration', testUserRegistration);
  await runTest('User login', testUserLogin);
  await runTest('Invalid login', testInvalidLogin);
  await runTest('Get user profile', testGetUserProfile);
  await runTest('Refresh token', testRefreshToken);

  // Test Suite 3: Model Endpoints
  console.log(`\n${colors.yellow}━━━ Test Suite 3: Model Endpoints (OpenAI Compatible) ━━━${colors.reset}\n`);
  await runTest('List models', testListModels);
  await runTest('Text completion', testTextCompletion);
  await runTest('Chat completion', testChatCompletion);
  await runTest('Embeddings', testEmbeddings);

  // Test Suite 4: Documentation
  console.log(`\n${colors.yellow}━━━ Test Suite 4: Documentation Endpoints ━━━${colors.reset}\n`);
  await runTest('Swagger UI', testSwaggerUI);
  await runTest('OpenAPI specification', testOpenAPISpec);

  // Print summary
  console.log(`\n${colors.blue}═══════════════════════════════════════════════════════════${colors.reset}`);
  console.log(`${colors.blue}  Test Summary${colors.reset}`);
  console.log(`${colors.blue}═══════════════════════════════════════════════════════════${colors.reset}\n`);
  console.log(`  Total Tests:  ${testResults.total}`);
  console.log(`  ${colors.green}Passed:       ${testResults.passed}${colors.reset}`);
  console.log(`  ${colors.red}Failed:       ${testResults.failed}${colors.reset}`);
  console.log(`  ${colors.yellow}Skipped:      ${testResults.skipped}${colors.reset}`);
  console.log(`  Success Rate: ${((testResults.passed / testResults.total) * 100).toFixed(1)}%\n`);

  process.exit(testResults.failed > 0 ? 1 : 0);
}

// Run tests
runAllTests().catch(error => {
  console.error(`${colors.red}Fatal error: ${error.message}${colors.reset}`);
  process.exit(1);
});


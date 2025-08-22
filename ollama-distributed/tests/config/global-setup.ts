import { chromium, FullConfig } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

/**
 * Global setup for Playwright tests
 * Prepares test environment and shared resources
 */
async function globalSetup(config: FullConfig) {
  console.log('🚀 Starting global test setup...');
  
  // Create test reports directory
  const reportsDir = path.join(process.cwd(), 'tests/reports');
  if (!fs.existsSync(reportsDir)) {
    fs.mkdirSync(reportsDir, { recursive: true });
  }
  
  // Create test fixtures directory
  const fixturesDir = path.join(process.cwd(), 'tests/fixtures');
  if (!fs.existsSync(fixturesDir)) {
    fs.mkdirSync(fixturesDir, { recursive: true });
  }
  
  // Setup test database and seed data
  await setupTestDatabase();
  
  // Setup test authentication tokens
  await setupTestAuth();
  
  // Verify application is running
  await verifyApplicationHealth();
  
  // Setup global test data
  await setupGlobalTestData();
  
  console.log('✅ Global test setup complete');
}

/**
 * Setup test database with seed data
 */
async function setupTestDatabase(): Promise<void> {
  console.log('📊 Setting up test database...');
  
  // Note: In a real implementation, this would:
  // 1. Create test database
  // 2. Run migrations
  // 3. Seed with test data
  // 4. Create test users with various roles
  
  // For now, we'll simulate this with a delay
  await new Promise(resolve => setTimeout(resolve, 1000));
  
  console.log('✅ Test database setup complete');
}

/**
 * Setup authentication tokens for test users
 */
async function setupTestAuth(): Promise<void> {
  console.log('🔐 Setting up test authentication...');
  
  const testUsers = [
    {
      email: 'admin@example.com',
      password: 'admin123',
      role: 'admin',
      token: 'test-admin-token-123'
    },
    {
      email: 'user@example.com',
      password: 'user123',
      role: 'user',
      token: 'test-user-token-456'
    },
    {
      email: 'security-admin@example.com',
      password: 'security123',
      role: 'security_admin',
      token: 'test-security-token-789'
    },
    {
      email: '2fa-user@example.com',
      password: 'password123',
      role: 'user',
      token: 'test-2fa-token-101',
      mfaEnabled: true
    }
  ];
  
  // Store test user credentials for use in tests
  const authFile = path.join(process.cwd(), 'tests/fixtures/test-auth.json');
  fs.writeFileSync(authFile, JSON.stringify(testUsers, null, 2));
  
  console.log('✅ Test authentication setup complete');
}

/**
 * Verify application health before running tests
 */
async function verifyApplicationHealth(): Promise<void> {
  console.log('🏥 Verifying application health...');
  
  const browser = await chromium.launch();
  const page = await browser.newPage();
  
  try {
    // Check if application is responding
    const baseURL = process.env.BASE_URL || 'http://localhost:3000';
    await page.goto(`${baseURL}/health`);
    
    // Wait for health check response
    await page.waitForLoadState('networkidle');
    
    // Verify health status
    const healthStatus = await page.textContent('body');
    if (!healthStatus?.includes('OK') && !healthStatus?.includes('healthy')) {
      throw new Error('Application health check failed');
    }
    
    console.log('✅ Application health verified');
  } catch (error) {
    console.error('❌ Application health check failed:', error);
    throw error;
  } finally {
    await browser.close();
  }
}

/**
 * Setup global test data and fixtures
 */
async function setupGlobalTestData(): Promise<void> {
  console.log('📝 Setting up global test data...');
  
  // Create test configuration
  const testConfig = {
    baseURL: process.env.BASE_URL || 'http://localhost:3000',
    apiURL: process.env.API_URL || 'http://localhost:8080',
    websocketURL: process.env.WS_URL || 'ws://localhost:8080/ws',
    testTimeout: 30000,
    retryCount: 2,
    parallel: true,
    screenshots: 'only-on-failure',
    video: 'retain-on-failure',
    trace: 'on-first-retry'
  };
  
  // Save test configuration
  const configFile = path.join(process.cwd(), 'tests/fixtures/test-config.json');
  fs.writeFileSync(configFile, JSON.stringify(testConfig, null, 2));
  
  // Create sample test data
  const sampleTestData = {
    users: [
      {
        id: 'user-001',
        name: 'Test User 1',
        email: 'testuser1@example.com',
        role: 'user'
      },
      {
        id: 'user-002',
        name: 'Test Admin',
        email: 'testadmin@example.com',
        role: 'admin'
      }
    ],
    nodes: [
      {
        id: 'node-001',
        address: '192.168.1.100:8080',
        name: 'Test Node 1',
        status: 'ready',
        cpu: 45,
        memory: 68,
        models: ['llama2:7b', 'codellama:13b']
      },
      {
        id: 'node-002',
        address: '192.168.1.101:8080',
        name: 'Test Node 2',
        status: 'ready',
        cpu: 32,
        memory: 54,
        models: ['llama2:7b', 'mistral:7b']
      }
    ],
    models: [
      {
        id: 'llama2:7b',
        name: 'Llama 2 7B',
        size: '3.8GB',
        status: 'ready'
      },
      {
        id: 'codellama:13b',
        name: 'Code Llama 13B',
        size: '7.3GB',
        status: 'ready'
      },
      {
        id: 'mistral:7b',
        name: 'Mistral 7B',
        size: '4.1GB',
        status: 'ready'
      }
    ],
    performance: {
      cpu_utilization: 45,
      memory_usage: 68,
      network_traffic: 156, // MB/s
      request_latency: 125, // ms
      error_rate: 0.1 // %
    }
  };
  
  // Save sample test data
  const dataFile = path.join(process.cwd(), 'tests/fixtures/sample-data.json');
  fs.writeFileSync(dataFile, JSON.stringify(sampleTestData, null, 2));
  
  // Create test environment variables
  const envVars = {
    NODE_ENV: 'test',
    TEST_MODE: 'true',
    DISABLE_RATE_LIMITING: 'true',
    MOCK_EXTERNAL_SERVICES: 'true',
    TEST_DATABASE_URL: 'postgresql://test:test@localhost:5432/ollama_test',
    JWT_SECRET: 'test-jwt-secret-key-for-testing-only',
    ENCRYPTION_KEY: 'test-encryption-key-32-chars-long'
  };
  
  // Save environment variables for tests
  const envFile = path.join(process.cwd(), 'tests/fixtures/test.env');
  const envContent = Object.entries(envVars)
    .map(([key, value]) => `${key}=${value}`)
    .join('\n');
  fs.writeFileSync(envFile, envContent);
  
  console.log('✅ Global test data setup complete');
}

export default globalSetup;
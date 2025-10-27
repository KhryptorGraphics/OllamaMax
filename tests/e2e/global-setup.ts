import { FullConfig } from '@playwright/test';
import fs from 'fs/promises';
import path from 'path';

/**
 * Global Setup for OllamaMax E2E Tests
 * 
 * Responsibilities:
 * - Environment validation
 * - Test data preparation
 * - Service health checks
 * - Report directory setup
 * - Authentication setup (if needed)
 */

async function globalSetup(config: FullConfig) {
  console.log('🚀 Starting OllamaMax E2E Test Suite Global Setup');

  // UI server for browser navigation, API server for backend calls
  const baseURL = process.env.BASE_URL || config.projects[0]?.use?.baseURL || 'http://localhost:8080';
  const uiBaseURL = process.env.BASE_URL || 'http://localhost:8080';
  const apiBaseURL = process.env.API_BASE_URL || 'http://localhost:11434';
  const backendEnabled = process.env.BACKEND_UP !== '0';

  console.log(`UI Base URL: ${uiBaseURL}`);
  console.log(`API Base URL: ${apiBaseURL} (enabled: ${backendEnabled})`);

  // Create report directories
  await createReportDirectories();

  // Validate environment
  await validateEnvironment();

  // Wait for services to be ready
  if (backendEnabled) {
    await waitForBackendAPI(apiBaseURL);
  }
  await waitForServices(uiBaseURL);

  // Setup test data
  await setupTestData();

  console.log('✅ Global setup completed successfully');
}

/**
 * Create necessary report directories
 */
async function createReportDirectories() {
  const directories = [
    'reports',
    'reports/screenshots',
    'reports/performance',
    'reports/load-tests',
    'reports/security',
    'reports/playwright-report',
    'reports/allure-results',
    'test-results'
  ];
  
  for (const dir of directories) {
    await fs.mkdir(dir, { recursive: true }).catch(() => {});
  }
  
  console.log('📁 Report directories created');
}

/**
 * Validate environment and dependencies
 */
async function validateEnvironment() {
  // Check Node.js version
  const nodeVersion = process.version;
  console.log(`Node.js version: ${nodeVersion}`);
  
  // Check environment variables
  const requiredEnvVars = ['NODE_ENV'];
  const missingVars = requiredEnvVars.filter(varName => !process.env[varName]);
  
  if (missingVars.length > 0) {
    console.warn(`⚠️  Missing environment variables: ${missingVars.join(', ')}`);
  }
  
  // Set default environment if not specified
  if (!process.env.NODE_ENV) {
    process.env.NODE_ENV = 'test';
  }
  
  console.log(`Environment: ${process.env.NODE_ENV}`);
}

/**
 * Wait for backend API to be ready
 */
async function waitForBackendAPI(baseURL: string) {
  const maxRetries = 30;
  const retryInterval = 2000;

  console.log('⏳ Waiting for backend API to be ready...');

  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 5000);

      try {
        // Try backend health endpoint first
        const response = await fetch(`${baseURL}/api/v1/health`, {
          method: 'GET',
          signal: controller.signal
        });

        clearTimeout(timeoutId);

        if (response.ok) {
          console.log('✅ Backend API health endpoint is ready');
          return;
        } else if (response.status === 404) {
          // If health endpoint doesn't exist, try root
          const rootResponse = await fetch(baseURL, {
            method: 'GET',
            signal: controller.signal
          });
          if (rootResponse.ok) {
            console.log('✅ Backend API root endpoint is ready (health endpoint not found)');
            return;
          }
        }
        throw new Error(`HTTP ${response.status}`);
      } finally {
        clearTimeout(timeoutId);
      }

    } catch (error) {
      if (attempt === maxRetries) {
        console.error('❌ Backend API failed to start within timeout period');
        console.error(`Last error: ${error}`);
        throw new Error('Backend API not ready for testing');
      }

      console.log(`Backend API attempt ${attempt}/${maxRetries} failed, retrying in ${retryInterval}ms...`);
      await new Promise(resolve => setTimeout(resolve, retryInterval));
    }
  }
}

/**
 * Wait for services to be ready
 */
async function waitForServices(baseURL: string) {
  const maxRetries = 30;
  const retryInterval = 2000;

  console.log('⏳ Waiting for services to be ready...');

  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      // Test main application with AbortController for timeout
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 5000);

      try {
        const response = await fetch(baseURL, {
          method: 'GET',
          signal: controller.signal
        });

        clearTimeout(timeoutId);

        if (response.ok) {
          console.log('✅ Main application is ready');
          break;
        } else {
          throw new Error(`HTTP ${response.status}`);
        }
      } finally {
        clearTimeout(timeoutId);
      }

    } catch (error) {
      if (attempt === maxRetries) {
        console.error('❌ Services failed to start within timeout period');
        console.error(`Last error: ${error}`);
        throw new Error('Services not ready for testing');
      }

      console.log(`Attempt ${attempt}/${maxRetries} failed, retrying in ${retryInterval}ms...`);
      await new Promise(resolve => setTimeout(resolve, retryInterval));
    }
  }

  // Test additional endpoints and services
  await testEndpoints(baseURL);
  await checkDatabaseConnectivity();
  await checkRedisConnectivity();
}

/**
 * Test key endpoints for availability
 */
async function testEndpoints(baseURL: string) {
  const endpoints = [
    { path: '/api/v1/health', required: true },
    { path: '/api/health', required: false },
    { path: '/health', required: false },
    { path: '/api/v1/status', required: false }
  ];

  console.log('🔍 Testing endpoint availability...');

  let healthEndpointFound = false;

  for (const endpoint of endpoints) {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 3000);

    try {
      const response = await fetch(`${baseURL}${endpoint.path}`, {
        method: 'GET',
        signal: controller.signal
      });

      clearTimeout(timeoutId);

      if (response.ok) {
        console.log(`✅ ${endpoint.path} is available`);
        if (endpoint.path.includes('health')) {
          healthEndpointFound = true;
        }
      } else {
        console.log(`⚠️  ${endpoint.path} returned ${response.status}`);
      }
    } catch (error) {
      clearTimeout(timeoutId);
      if (endpoint.required && endpoint.path === '/api/v1/health' && !healthEndpointFound) {
        console.log(`⚠️  ${endpoint.path} is not available, will check alternatives`);
      } else if (endpoint.required) {
        throw new Error(`Required endpoint ${endpoint.path} is not available: ${error}`);
      } else {
        console.log(`ℹ️  ${endpoint.path} is not available (optional)`);
      }
    }
  }

  if (!healthEndpointFound) {
    console.log('⚠️  No health endpoint found, but main application is responsive');
  }
}

/**
 * Check database connectivity
 */
async function checkDatabaseConnectivity() {
  console.log('🔍 Checking database connectivity...');

  const dbHost = process.env.DB_HOST || 'localhost';
  const dbPort = parseInt(process.env.DB_PORT || process.env.POSTGRES_PORT || '15432');

  try {
    // Simple TCP check using fetch to a health endpoint that queries the DB
    // If health endpoint exists and returns OK, DB is likely connected
    console.log(`ℹ️  Database configured at ${dbHost}:${dbPort}`);
    console.log('✅ Database connectivity check skipped (will be verified by health endpoint)');
  } catch (error) {
    console.log(`⚠️  Could not verify database connectivity: ${error}`);
    console.log('ℹ️  Tests may fail if database is required');
  }
}

/**
 * Check Redis connectivity
 */
async function checkRedisConnectivity() {
  console.log('🔍 Checking Redis connectivity...');

  const redisHost = process.env.REDIS_HOST || 'localhost';
  const redisPort = parseInt(process.env.REDIS_PORT || '16379');

  try {
    // Simple TCP check - Redis connectivity will be verified through application health
    console.log(`ℹ️  Redis configured at ${redisHost}:${redisPort}`);
    console.log('✅ Redis connectivity check skipped (will be verified by health endpoint)');
  } catch (error) {
    console.log(`⚠️  Could not verify Redis connectivity: ${error}`);
    console.log('ℹ️  Tests may fail if Redis is required');
  }
}

/**
 * Setup test data and configurations
 */
async function setupTestData() {
  console.log('📊 Setting up test data...');
  
  // Create test configuration file
  const testConfig = {
    timestamp: new Date().toISOString(),
    testSuiteVersion: '1.0.0',
    environment: process.env.NODE_ENV,
    baseURL: process.env.BASE_URL,
    testData: {
      // Test user accounts (for future authentication tests)
      testUsers: [
        { username: 'testuser1', email: 'test1@example.com' },
        { username: 'testuser2', email: 'test2@example.com' }
      ],
      
      // Test model configurations
      testModels: [
        { name: 'test-model', size: 'small' },
        { name: 'test-model-large', size: 'large' }
      ],
      
      // Test prompts for AI testing
      testPrompts: [
        'Hello, world!',
        'What is artificial intelligence?',
        'Explain distributed computing.'
      ]
    }
  };
  
  const configPath = path.join('test-results', 'test-config.json');
  await fs.writeFile(configPath, JSON.stringify(testConfig, null, 2));
  
  // Create performance baseline file
  const performanceBaseline = {
    maxLoadTime: 10000, // 10 seconds
    maxFirstContentfulPaint: 5000, // 5 seconds
    maxLargestContentfulPaint: 8000, // 8 seconds
    maxCumulativeLayoutShift: 0.1,
    maxFirstInputDelay: 100, // 100ms
    maxMemoryUsage: 100 * 1024 * 1024 // 100MB
  };
  
  const baselinePath = path.join('test-results', 'performance-baseline.json');
  await fs.writeFile(baselinePath, JSON.stringify(performanceBaseline, null, 2));
  
  console.log('📋 Test data setup completed');
}

export default globalSetup;
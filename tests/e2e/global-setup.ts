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
  
  const baseURL = process.env.BASE_URL || config.projects[0]?.use?.baseURL || 'http://localhost:8080';
  console.log(`Target URL: ${baseURL}`);
  
  // Create report directories
  await createReportDirectories();
  
  // Validate environment
  await validateEnvironment();
  
  // Wait for services to be ready
  await waitForServices(baseURL);
  
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
 * Wait for services to be ready
 */
async function waitForServices(baseURL: string) {
  const maxRetries = 30;
  const retryInterval = 2000;
  
  console.log('⏳ Waiting for services to be ready...');
  
  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      // Test main application
      const response = await fetch(baseURL, {
        method: 'GET',
        timeout: 5000
      } as any);
      
      if (response.ok) {
        console.log('✅ Main application is ready');
        break;
      } else {
        throw new Error(`HTTP ${response.status}`);
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
  
  // Test additional endpoints
  await testEndpoints(baseURL);
}

/**
 * Test key endpoints for availability
 */
async function testEndpoints(baseURL: string) {
  const endpoints = [
    { path: '/api/v1/health', required: false },
    { path: '/api/health', required: false },
    { path: '/health', required: false },
    { path: '/api/v1/status', required: false }
  ];
  
  console.log('🔍 Testing endpoint availability...');
  
  for (const endpoint of endpoints) {
    try {
      const response = await fetch(`${baseURL}${endpoint.path}`, {
        method: 'GET',
        timeout: 3000
      } as any);
      
      if (response.ok) {
        console.log(`✅ ${endpoint.path} is available`);
      } else {
        console.log(`⚠️  ${endpoint.path} returned ${response.status}`);
      }
    } catch (error) {
      if (endpoint.required) {
        throw new Error(`Required endpoint ${endpoint.path} is not available`);
      } else {
        console.log(`ℹ️  ${endpoint.path} is not available (optional)`);
      }
    }
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
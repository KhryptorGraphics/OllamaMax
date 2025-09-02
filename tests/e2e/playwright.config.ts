import { defineConfig, devices } from '@playwright/test';

/**
 * Comprehensive Playwright Configuration for OllamaMax Distributed AI Platform
 * 
 * Features:
 * - Multi-browser testing (Chromium, Firefox, Safari)
 * - Performance monitoring and metrics collection
 * - Security testing capabilities
 * - Load testing scenarios
 * - Cross-device responsive testing
 * - Real-time WebSocket testing
 * - API integration validation
 */

export default defineConfig({
  testDir: './tests',
  /* Run tests in files in parallel */
  fullyParallel: true,
  /* Fail the build on CI if you accidentally left test.only in the source code. */
  forbidOnly: !!process.env.CI,
  /* Retry on CI only */
  retries: process.env.CI ? 2 : 0,
  /* Opt out of parallel tests on CI. */
  workers: process.env.CI ? 1 : undefined,
  /* Reporter to use. See https://playwright.dev/docs/test-reporters */
  reporter: [
    ['html', { outputFolder: './reports/playwright-report' }],
    ['json', { outputFile: './reports/results.json' }],
    ['junit', { outputFile: './reports/junit-results.xml' }],
    ['line'],
    ['allure-playwright', { outputFolder: './reports/allure-results' }]
  ],

  /* Shared settings for all the projects below. See https://playwright.dev/docs/api/class-testoptions. */
  use: {
    /* Base URL to use in actions like `await page.goto('/')`. */
    baseURL: process.env.BASE_URL || 'http://localhost:8080',
    
    /* Collect trace when retrying the failed test. See https://playwright.dev/docs/trace-viewer */
    trace: 'on-first-retry',
    
    /* Screenshot on failure */
    screenshot: 'only-on-failure',
    
    /* Video recording for debugging */
    video: 'retain-on-failure',
    
    /* Global timeout for each action */
    actionTimeout: 10000,
    
    /* Global navigation timeout */
    navigationTimeout: 30000,
  },

  /* Global timeout for each test */
  timeout: 120000,
  
  /* Global timeout for each test file */
  globalTimeout: 600000,

  /* Configure projects for major browsers */
  projects: [
    {
      name: 'setup',
      testMatch: /.*\.setup\.ts/,
    },
    
    {
      name: 'chromium',
      use: { 
        ...devices['Desktop Chrome'],
        // Enable performance monitoring
        launchOptions: {
          args: [
            '--enable-features=VaapiVideoDecoder',
            '--disable-dev-shm-usage',
            '--no-sandbox',
            '--disable-setuid-sandbox'
          ]
        }
      },
      dependencies: ['setup'],
    },

    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
      dependencies: ['setup'],
    },

    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
      dependencies: ['setup'],
    },

    /* Mobile testing */
    {
      name: 'Mobile Chrome',
      use: { ...devices['Pixel 5'] },
      dependencies: ['setup'],
    },
    {
      name: 'Mobile Safari',
      use: { ...devices['iPhone 12'] },
      dependencies: ['setup'],
    },

    /* Tablet testing */
    {
      name: 'Tablet',
      use: { ...devices['iPad Pro'] },
      dependencies: ['setup'],
    },

    /* Performance testing project */
    {
      name: 'performance',
      testMatch: /.*\.performance\.ts/,
      use: { ...devices['Desktop Chrome'] },
      dependencies: ['setup'],
    },

    /* Security testing project */
    {
      name: 'security',
      testMatch: /.*\.security\.ts/,
      use: { ...devices['Desktop Chrome'] },
      dependencies: ['setup'],
    },

    /* Load testing project */
    {
      name: 'load',
      testMatch: /.*\.load\.ts/,
      use: { ...devices['Desktop Chrome'] },
      dependencies: ['setup'],
    },

    /* API testing project */
    {
      name: 'api',
      testMatch: /.*\.api\.ts/,
      use: { 
        ...devices['Desktop Chrome'],
        extraHTTPHeaders: {
          'Accept': 'application/json',
          'Content-Type': 'application/json'
        }
      },
      dependencies: ['setup'],
    }
  ],

  /* Web Server configuration for local development */
  webServer: process.env.CI ? undefined : {
    command: 'cd ../../ && go run . --port=8080',
    port: 8080,
    reuseExistingServer: !process.env.CI,
    timeout: 120000,
  },

  /* Global setup and teardown */
  globalSetup: require.resolve('./global-setup.ts'),
  globalTeardown: require.resolve('./global-teardown.ts'),

  /* Test expectations */
  expect: {
    /* Maximum time expect() should wait for the condition to be met. */
    timeout: 10000,
    
    /* Configure screenshot comparisons */
    toHaveScreenshot: { 
      threshold: 0.2, 
      animations: 'disabled',
      mode: 'css'
    },
  },

  /* Output directory for test artifacts */
  outputDir: './test-results/',

  /* Metadata */
  metadata: {
    project: 'OllamaMax Distributed AI Platform',
    testSuite: 'E2E Testing Suite',
    version: '1.0.0',
    environment: process.env.NODE_ENV || 'development'
  }
});
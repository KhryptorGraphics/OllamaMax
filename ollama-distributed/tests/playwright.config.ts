import { defineConfig, devices } from '@playwright/test';
import * as dotenv from 'dotenv';

// Load environment variables
dotenv.config();

/**
 * Enterprise Playwright Configuration
 * Supports multiple browsers, devices, and testing scenarios
 */
export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: [
    ['html', { outputFolder: 'tests/reports/html' }],
    ['json', { outputFile: 'tests/reports/results.json' }],
    ['junit', { outputFile: 'tests/reports/junit.xml' }],
    ['list'],
    ['github']
  ],
  use: {
    baseURL: process.env.BASE_URL || 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    actionTimeout: 30000,
    navigationTimeout: 60000,
    headless: process.env.CI ? true : false,
  },
  projects: [
    // Desktop Browsers
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },
    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },
    
    // Mobile Devices
    {
      name: 'mobile',
      use: { ...devices['Pixel 5'] },
    },
    {
      name: 'Mobile Safari',
      use: { ...devices['iPhone 13'] },
    },
    
    // Tablet Devices
    {
      name: 'tablet',
      use: { ...devices['iPad Pro'] },
    },
    
    // Visual Regression Testing
    {
      name: 'visual-regression',
      use: { ...devices['Desktop Chrome'] },
      testMatch: /.*\.visual\.spec\.ts/,
    },
    
    // Performance Testing
    {
      name: 'performance',
      use: { ...devices['Desktop Chrome'] },
      testMatch: /.*\.performance\.spec\.ts/,
    },
    
    // Accessibility Testing
    {
      name: 'accessibility',
      use: { ...devices['Desktop Chrome'] },
      testMatch: /.*\.a11y\.spec\.ts/,
    }
  ],
  
  // Global setup and teardown
  globalSetup: require.resolve('./tests/config/global-setup.ts'),
  globalTeardown: require.resolve('./tests/config/global-teardown.ts'),
  
  // Test output directories
  outputDir: 'tests/test-results/',
  
  // Web server configuration for local development
  webServer: process.env.CI ? undefined : {
    command: 'cd ../web/frontend && npm run dev',
    url: 'http://localhost:3000',
    reuseExistingServer: !process.env.CI,
    timeout: 120000,
  },
  
  // Expect configuration
  expect: {
    timeout: 10000,
    toHaveScreenshot: {
      mode: 'strict',
      threshold: 0.3,
    },
    toMatchSnapshot: {
      threshold: 0.3,
    },
  },
});
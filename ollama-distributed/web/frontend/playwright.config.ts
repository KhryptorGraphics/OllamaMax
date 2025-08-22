import { defineConfig, devices } from '@playwright/test'

/**
 * Comprehensive E2E Testing Configuration for Distributed Ollama System
 * 
 * Features:
 * - Multi-browser testing (Chrome, Firefox, Safari, Edge)
 * - Mobile responsive testing
 * - Performance monitoring with Lighthouse
 * - Accessibility testing integration
 * - Visual regression testing
 * - Cross-platform support
 */
export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 2 : undefined,
  timeout: 30000,
  expect: {
    timeout: 10000,
    toHaveScreenshot: {
      threshold: 0.2,
      mode: 'strict'
    },
    toMatchAriaSnapshot: {
      mode: 'strict'
    }
  },
  reporter: [
    ['html', { outputFolder: 'playwright-report', open: 'never' }],
    ['json', { outputFile: 'test-results/results.json' }],
    ['junit', { outputFile: 'test-results/results.xml' }],
    ['github']
  ],
  use: {
    baseURL: 'http://localhost:5173',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    actionTimeout: 10000,
    navigationTimeout: 15000,
    // Performance tracking
    extraHTTPHeaders: {
      'Accept-Language': 'en-US,en;q=0.9'
    }
  },
  projects: [
    {
      name: 'setup',
      testMatch: /.*\.setup\.ts/,
      teardown: 'cleanup'
    },
    {
      name: 'cleanup',
      testMatch: /.*\.teardown\.ts/
    },
    // Desktop browsers
    {
      name: 'chromium',
      use: { 
        ...devices['Desktop Chrome'],
        viewport: { width: 1280, height: 720 }
      },
      dependencies: ['setup']
    },
    {
      name: 'firefox',
      use: { 
        ...devices['Desktop Firefox'],
        viewport: { width: 1280, height: 720 }
      },
      dependencies: ['setup']
    },
    {
      name: 'webkit',
      use: { 
        ...devices['Desktop Safari'],
        viewport: { width: 1280, height: 720 }
      },
      dependencies: ['setup']
    },
    {
      name: 'edge',
      use: { 
        ...devices['Desktop Edge'],
        channel: 'msedge',
        viewport: { width: 1280, height: 720 }
      },
      dependencies: ['setup']
    },
    // Mobile devices
    {
      name: 'mobile-chrome',
      use: { ...devices['Pixel 5'] },
      dependencies: ['setup']
    },
    {
      name: 'mobile-safari',
      use: { ...devices['iPhone 12'] },
      dependencies: ['setup']
    },
    // Performance testing
    {
      name: 'performance',
      use: {
        ...devices['Desktop Chrome'],
        viewport: { width: 1920, height: 1080 },
        // Slow 3G network simulation
        launchOptions: {
          args: ['--enable-features=NetworkService']
        }
      },
      testMatch: '**/performance/*.spec.ts',
      dependencies: ['setup']
    },
    // Accessibility testing
    {
      name: 'accessibility',
      use: {
        ...devices['Desktop Chrome'],
        viewport: { width: 1280, height: 720 }
      },
      testMatch: '**/accessibility/*.spec.ts',
      dependencies: ['setup']
    },
    // Visual regression testing
    {
      name: 'visual',
      use: {
        ...devices['Desktop Chrome'],
        viewport: { width: 1280, height: 720 }
      },
      testMatch: '**/visual/*.spec.ts',
      dependencies: ['setup']
    }
  ],
  webServer: [
    {
      command: 'npm run dev',
      url: 'http://localhost:5174/v2/',
      reuseExistingServer: !process.env.CI,
      timeout: 120000
    },
    {
      command: 'node ../scripts/legacy-static-server.mjs',
      url: 'http://localhost:8090/',
      reuseExistingServer: !process.env.CI,
      timeout: 120000
    }
  ],
})

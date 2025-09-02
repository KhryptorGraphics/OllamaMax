/**
 * Playwright Configuration for Performance Testing
 * Optimized for performance measurement and monitoring
 */

module.exports = {
  testDir: './tests',
  testMatch: ['**/performance-*.test.js'],
  timeout: 60000, // Extended timeout for performance tests
  fullyParallel: false, // Sequential for accurate performance measurement
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: 1, // Single worker for consistent performance measurement
  
  reporter: [
    ['list'],
    ['json', { outputFile: 'test-results/performance/performance-results.json' }],
    ['html', { outputFolder: 'test-results/performance/html-report', open: 'never' }]
  ],
  
  use: {
    headless: true,
    viewport: { width: 1280, height: 720 },
    ignoreHTTPSErrors: true,
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    trace: 'retain-on-failure',
    
    // Performance testing specific settings
    launchOptions: {
      args: [
        '--no-sandbox',
        '--disable-dev-shm-usage',
        '--disable-blink-features=AutomationControlled',
        '--enable-precise-memory-info', // Enable memory profiling
        '--js-flags=--expose-gc',       // Enable garbage collection
        '--memory-pressure-off',         // Disable memory pressure simulation
      ]
    }
  },
  
  // Performance testing projects
  projects: [
    {
      name: 'performance-chrome',
      use: { 
        ...require('@playwright/test').devices['Desktop Chrome'],
        launchOptions: {
          args: [
            '--no-sandbox',
            '--disable-dev-shm-usage',
            '--enable-precise-memory-info',
            '--js-flags=--expose-gc'
          ]
        }
      },
    },
    
    {
      name: 'performance-mobile',
      use: { 
        ...require('@playwright/test').devices['iPhone 12'],
        launchOptions: {
          args: [
            '--no-sandbox',
            '--disable-dev-shm-usage',
            '--enable-precise-memory-info'
          ]
        }
      },
    }
  ],
  
  // Web server for testing
  webServer: [
    {
      command: 'cd web-interface && python3 -m http.server 8080',
      port: 8080,
      reuseExistingServer: !process.env.CI,
      timeout: 10000
    },
    {
      command: 'node api-server/server.js',
      port: 13100,
      reuseExistingServer: !process.env.CI,
      timeout: 15000,
      env: {
        NODE_ENV: 'test',
        REDIS_HOST: 'localhost',
        REDIS_PORT: '6379'
      }
    }
  ],
  
  // Global setup for performance testing
  globalSetup: require.resolve('./tests/global-performance-setup.js'),
  globalTeardown: require.resolve('./tests/global-performance-teardown.js'),
  
  // Expect configuration for performance assertions
  expect: {
    timeout: 10000,
    toHaveScreenshot: {
      threshold: 0.5,
      mode: 'strict'
    }
  }
};
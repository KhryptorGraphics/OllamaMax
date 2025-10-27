import { devices } from '@playwright/test';

export default {
  testDir: './tests',
  testMatch: ['**/e2e/**/*.test.js', '**/e2e/**/*.spec.ts', '**/*.spec.ts', '**/playwright/**/*.test.js', '**/ui-*.test.js', '**/api-*.test.js'],
  testIgnore: ['**/*.cjs', '**/node_modules/**'],
  timeout: 30000,
  fullyParallel: false, // Run sequentially for better debugging
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 1,
  workers: process.env.CI ? 1 : 2,
  globalSetup: './tests/e2e/global-setup.ts',
  reporter: [
    ['list'],
    ['json', { outputFile: 'test-results/playwright/results.json' }],
    ['html', { outputFolder: 'test-results/playwright-report' }]
  ],
  use: {
    baseURL: process.env.BASE_URL || 'http://localhost:8080',
    headless: process.env.CI ? true : false, // Show browser during development
    viewport: { width: 1280, height: 720 },
    ignoreHTTPSErrors: true,
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    trace: 'retain-on-failure',
  },
  projects: [
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
  ],
  webServer: [
    {
      // Backend API server (Go) - primary server for API tests
      command: process.env.BACKEND_CMD || 'go run .',
      port: 11434,
      reuseExistingServer: !process.env.CI,
      timeout: 120000,
      env: {
        PORT: '11434',
        NODE_ENV: 'test',
        BACKEND_UP: '1',
        DB_HOST: process.env.DB_HOST || 'localhost',
        DB_PORT: process.env.DB_PORT || '15432',
        REDIS_HOST: process.env.REDIS_HOST || 'localhost',
        REDIS_PORT: process.env.REDIS_PORT || '16379',
      }
    },
    {
      // Static UI server (optional, for UI tests)
      command: 'cd web-interface && python3 -m http.server 8080',
      port: 8080,
      reuseExistingServer: !process.env.CI,
      timeout: 120000,
    }
  ],
};
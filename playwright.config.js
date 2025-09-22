import { devices } from '@playwright/test';

export default {
  testDir: './tests',
  testMatch: ['**/e2e/**/*.test.js', '**/playwright/**/*.test.js', '**/ui-*.test.js', '**/api-*.test.js'],
  testIgnore: ['**/*.cjs', '**/node_modules/**'],
  timeout: 30000,
  fullyParallel: false, // Run sequentially for better debugging
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 1,
  workers: process.env.CI ? 1 : 2,
  reporter: [
    ['list'],
    ['json', { outputFile: 'test-results/playwright/results.json' }],
    ['html', { outputFolder: 'test-results/playwright-report' }]
  ],
  use: {
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
      command: 'cd web-interface && python3 -m http.server 8080',
      port: 8080,
      reuseExistingServer: !process.env.CI,
    }
  ],
};
module.exports = {
  preset: 'jest-puppeteer',
  testTimeout: 60000,
  testMatch: ['**/specs/**/*.test.js'],
  reporters: [
    'default',
    ['jest-junit', { outputDirectory: './reports', outputName: 'junit.xml' }],
    ['jest-html-reporter', { outputPath: './reports/report.html', pageTitle: 'OllamaMax E2E Report' }]
  ],
  globals: {
    BASE_URL: process.env.BASE_URL || 'http://localhost',
  },
};


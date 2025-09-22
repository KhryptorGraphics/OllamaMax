module.exports = {
  testEnvironment: 'node',
  roots: ['<rootDir>/tests'],
  testMatch: [
    '**/__tests__/**/*.(test|spec).(js|cjs)',
    '**/?(*.)+(spec|test).(js|cjs)'
  ],
  collectCoverageFrom: [
    'api-server/**/*.js',
    'src/agents/**/*.js', 
    'critical-fixes/**/*.js',
    '!**/node_modules/**',
    '!**/coverage/**'
  ],
  coverageDirectory: 'coverage',
  coverageReporters: ['text', 'lcov', 'html'],
  coverageThreshold: {
    global: {
      branches: 70,
      functions: 70,
      lines: 70,
      statements: 70
    }
  },
  setupFilesAfterEnv: ['<rootDir>/tests/jest.setup.cjs'],
  testTimeout: 30000,
  verbose: true
};
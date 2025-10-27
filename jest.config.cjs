module.exports = {
  testEnvironment: 'node',
  roots: ['<rootDir>/tests'],
  testMatch: [
    '**/__tests__/**/*.(test|spec).(js|cjs)',
    '**/?(*.)+(spec|test).(js|cjs)'
  ],
  collectCoverageFrom: [
    'api-server/**/*.js',
    'critical-fixes/**/*.js',
    'src/**/*.js',
    // Exclude TS/TSX files from coverage as they require transformer
    // 'api-server/**/*.ts',
    // 'src/**/*.ts',
    // 'src/**/*.tsx',
    '!**/node_modules/**',
    '!**/coverage/**',
    '!**/*.test.*',
    '!**/*.spec.*',
    '!tests/**'
  ],
  coverageDirectory: 'coverage',
  coverageReporters: ['text', 'lcov', 'html'],
  coverageThreshold: {
    global: {
      branches: 90,
      functions: 90,
      lines: 90,
      statements: 90
    }
  },
  setupFilesAfterEnv: ['<rootDir>/tests/jest.setup.cjs'],
  testTimeout: 30000,
  verbose: true
};
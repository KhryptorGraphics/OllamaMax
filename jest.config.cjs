module.exports = {
  testEnvironment: 'node',
  roots: ['<rootDir>/tests'],
  testMatch: [
    '**/__tests__/**/*.(test|spec).(js|cjs)',
    '**/?(*.)+(spec|test).(js|cjs)'
  ],
  transform: {
    '^.+\\.(ts|tsx)$': ['ts-jest', {
      tsconfig: {
        jsx: 'react',
        esModuleInterop: true,
        allowSyntheticDefaultImports: true
      }
    }]
  },
  collectCoverageFrom: [
    'api-server/**/*.{js,ts,tsx}',
    'critical-fixes/**/*.{js,ts}',
    'src/**/*.{js,ts,tsx}',
    '!**/node_modules/**',
    '!**/coverage/**',
    '!**/*.test.*',
    '!**/*.spec.*',
    '!**/*.d.ts',
    '!tests/**'
  ],
  moduleFileExtensions: ['js', 'jsx', 'ts', 'tsx', 'json', 'node'],
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
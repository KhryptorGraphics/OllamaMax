// Jest setup file for global test configuration

// Mock console methods to reduce noise during tests
global.console = {
  ...console,
  // Keep these methods for debugging
  log: jest.fn(),
  debug: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
  error: jest.fn(),
};

// Set up global test timeout
jest.setTimeout(30000);

// Mock fetch globally
global.fetch = jest.fn();

// Clean up mocks after each test
afterEach(() => {
  jest.clearAllMocks();
});

// Clean up timers
afterEach(() => {
  jest.clearAllTimers();
  jest.useRealTimers();
});
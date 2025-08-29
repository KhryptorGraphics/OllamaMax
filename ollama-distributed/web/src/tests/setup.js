// Jest Test Setup
// Global test configuration and mocks

import '@testing-library/jest-dom';

// Mock localStorage
const localStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
};
global.localStorage = localStorageMock;

// Mock sessionStorage
const sessionStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
};
global.sessionStorage = sessionStorageMock;

// Mock window.location
delete window.location;
window.location = {
  hostname: 'localhost',
  protocol: 'http:',
  host: 'localhost:3000',
  href: 'http://localhost:3000',
  pathname: '/',
  search: '',
  hash: '',
  assign: jest.fn(),
  replace: jest.fn(),
  reload: jest.fn(),
};

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: jest.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: jest.fn(), // deprecated
    removeListener: jest.fn(), // deprecated
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
  })),
});

// Mock ResizeObserver
global.ResizeObserver = jest.fn().mockImplementation(() => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn(),
}));

// Mock IntersectionObserver
global.IntersectionObserver = jest.fn().mockImplementation(() => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn(),
}));

// Mock fetch
global.fetch = jest.fn();

// Mock WebSocket
global.WebSocket = jest.fn().mockImplementation(() => ({
  close: jest.fn(),
  send: jest.fn(),
  addEventListener: jest.fn(),
  removeEventListener: jest.fn(),
  readyState: 1,
  CONNECTING: 0,
  OPEN: 1,
  CLOSING: 2,
  CLOSED: 3,
}));

// Mock XMLHttpRequest for file uploads
global.XMLHttpRequest = jest.fn().mockImplementation(() => ({
  open: jest.fn(),
  send: jest.fn(),
  setRequestHeader: jest.fn(),
  addEventListener: jest.fn(),
  removeEventListener: jest.fn(),
  upload: {
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
  },
  status: 200,
  statusText: 'OK',
  responseText: '{}',
  response: {},
}));

// Mock console methods to reduce noise in tests
const originalConsoleError = console.error;
const originalConsoleWarn = console.warn;

beforeEach(() => {
  // Reset all mocks before each test
  jest.clearAllMocks();
  
  // Reset localStorage and sessionStorage
  localStorageMock.getItem.mockClear();
  localStorageMock.setItem.mockClear();
  localStorageMock.removeItem.mockClear();
  localStorageMock.clear.mockClear();
  
  sessionStorageMock.getItem.mockClear();
  sessionStorageMock.setItem.mockClear();
  sessionStorageMock.removeItem.mockClear();
  sessionStorageMock.clear.mockClear();
  
  // Reset fetch mock
  fetch.mockClear();
  
  // Suppress console errors and warnings in tests unless explicitly testing them
  console.error = jest.fn();
  console.warn = jest.fn();
});

afterEach(() => {
  // Restore console methods
  console.error = originalConsoleError;
  console.warn = originalConsoleWarn;
  
  // Clean up any timers
  jest.clearAllTimers();
});

// Global test utilities
global.testUtils = {
  // Helper to create mock API responses
  createMockApiResponse: (data, ok = true, status = 200) => ({
    ok,
    status,
    statusText: ok ? 'OK' : 'Error',
    json: async () => data,
    text: async () => JSON.stringify(data),
    headers: new Map([['content-type', 'application/json']]),
  }),
  
  // Helper to create mock user
  createMockUser: (overrides = {}) => ({
    id: 1,
    username: 'testuser',
    email: 'test@example.com',
    role: 'user',
    active: true,
    created_at: '2024-01-01T00:00:00Z',
    ...overrides,
  }),
  
  // Helper to create mock node
  createMockNode: (overrides = {}) => ({
    id: 'node-1',
    address: '192.168.1.100:11434',
    status: 'online',
    health: 'healthy',
    models: ['llama2:7b'],
    usage: { cpu: 45, memory: 62, bandwidth: 23.5 },
    ...overrides,
  }),
  
  // Helper to create mock model
  createMockModel: (overrides = {}) => ({
    id: 'model-1',
    name: 'llama2:7b',
    size: 3800000000,
    status: 'available',
    replicas: ['node-1'],
    inference_ready: true,
    ...overrides,
  }),
  
  // Helper to wait for async operations
  waitForAsync: () => new Promise(resolve => setTimeout(resolve, 0)),
  
  // Helper to simulate user events with proper async handling
  simulateUserEvent: async (element, event, options = {}) => {
    const { fireEvent } = await import('@testing-library/react');
    fireEvent[event](element, options);
    await global.testUtils.waitForAsync();
  },
};

// Mock Chart.js to avoid canvas issues in tests
jest.mock('chart.js', () => ({
  Chart: jest.fn().mockImplementation(() => ({
    destroy: jest.fn(),
    update: jest.fn(),
    render: jest.fn(),
  })),
  registerables: [],
}));

// Mock react-chartjs-2
jest.mock('react-chartjs-2', () => ({
  Line: ({ data, options }) => (
    <div data-testid="line-chart" data-chart-data={JSON.stringify(data)} />
  ),
  Bar: ({ data, options }) => (
    <div data-testid="bar-chart" data-chart-data={JSON.stringify(data)} />
  ),
  Pie: ({ data, options }) => (
    <div data-testid="pie-chart" data-chart-data={JSON.stringify(data)} />
  ),
  Doughnut: ({ data, options }) => (
    <div data-testid="doughnut-chart" data-chart-data={JSON.stringify(data)} />
  ),
}));

// Mock FontAwesome icons
jest.mock('@fortawesome/react-fontawesome', () => ({
  FontAwesomeIcon: ({ icon, className, ...props }) => (
    <i 
      className={`fa-icon ${className || ''}`} 
      data-icon={typeof icon === 'string' ? icon : icon?.iconName || 'unknown'}
      {...props}
    />
  ),
}));

// CSS imports are handled by moduleNameMapper in package.json

// Import React for test components
import React from 'react';

// Error boundary for testing
export class TestErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error) {
    return { hasError: true, error };
  }

  componentDidCatch(error, errorInfo) {
    console.error('Test Error Boundary caught an error:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return (
        <div data-testid="error-boundary">
          <h2>Something went wrong.</h2>
          <details>
            <summary>Error details</summary>
            <pre>{this.state.error?.toString()}</pre>
          </details>
        </div>
      );
    }

    return this.props.children;
  }
}

// Custom render function with providers
export const renderWithProviders = (ui, options = {}) => {
  const { render } = require('@testing-library/react');
  
  const Wrapper = ({ children }) => (
    <TestErrorBoundary>
      {children}
    </TestErrorBoundary>
  );

  return render(ui, { wrapper: Wrapper, ...options });
};

// Make renderWithProviders available globally
global.renderWithProviders = renderWithProviders;

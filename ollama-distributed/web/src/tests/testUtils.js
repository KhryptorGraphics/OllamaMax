/**
 * Testing Utilities
 * 
 * Comprehensive testing framework with unit tests, integration tests,
 * accessibility tests, and performance tests.
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { axe, toHaveNoViolations } from 'jest-axe';
import userEvent from '@testing-library/user-event';
import { ThemeProvider } from '../design-system/theme/ThemeProvider.jsx';
import { AuthProvider } from '../contexts/AuthContext.jsx';
import securityService from '../services/securityService.js';
import { performanceMonitor } from '../utils/performance.js';

// Extend Jest matchers
expect.extend(toHaveNoViolations);

// Test wrapper with providers
export const TestWrapper = ({ children, theme = 'light', user = null }) => {
  const mockAuthContext = {
    user,
    isAuthenticated: !!user,
    login: jest.fn(),
    logout: jest.fn(),
    loading: false,
  };

  return (
    <ThemeProvider defaultTheme={theme}>
      <AuthProvider value={mockAuthContext}>
        {children}
      </AuthProvider>
    </ThemeProvider>
  );
};

// Custom render function with providers
export const renderWithProviders = (ui, options = {}) => {
  const { theme, user, ...renderOptions } = options;
  
  return render(ui, {
    wrapper: ({ children }) => (
      <TestWrapper theme={theme} user={user}>
        {children}
      </TestWrapper>
    ),
    ...renderOptions,
  });
};

// Accessibility testing utilities
export const testAccessibility = async (component, options = {}) => {
  const { container } = renderWithProviders(component);
  const results = await axe(container, {
    rules: {
      // Configure specific rules
      'color-contrast': { enabled: true },
      'keyboard-navigation': { enabled: true },
      'focus-management': { enabled: true },
      'aria-labels': { enabled: true },
      ...options.rules,
    },
  });
  
  expect(results).toHaveNoViolations();
  return results;
};

// Keyboard navigation testing
export const testKeyboardNavigation = async (component) => {
  const user = userEvent.setup();
  renderWithProviders(component);
  
  // Test Tab navigation
  await user.tab();
  const firstFocusable = document.activeElement;
  expect(firstFocusable).toBeInTheDocument();
  
  // Test Shift+Tab navigation
  await user.tab({ shift: true });
  const previousFocusable = document.activeElement;
  
  // Test Enter/Space activation
  if (firstFocusable.tagName === 'BUTTON' || firstFocusable.getAttribute('role') === 'button') {
    const clickHandler = jest.fn();
    firstFocusable.addEventListener('click', clickHandler);
    
    await user.keyboard('{Enter}');
    expect(clickHandler).toHaveBeenCalled();
    
    clickHandler.mockClear();
    await user.keyboard(' ');
    expect(clickHandler).toHaveBeenCalled();
  }
  
  return { firstFocusable, previousFocusable };
};

// Screen reader testing
export const testScreenReaderSupport = (component) => {
  renderWithProviders(component);
  
  // Check for proper ARIA labels
  const elementsWithAriaLabel = screen.queryAllByLabelText(/.+/);
  const elementsWithAriaDescribedBy = document.querySelectorAll('[aria-describedby]');
  const elementsWithRole = document.querySelectorAll('[role]');
  
  return {
    ariaLabels: elementsWithAriaLabel.length,
    ariaDescriptions: elementsWithAriaDescribedBy.length,
    roles: elementsWithRole.length,
  };
};

// Performance testing utilities
export const testPerformance = async (component, options = {}) => {
  const { iterations = 10, timeout = 5000 } = options;
  const metrics = [];
  
  for (let i = 0; i < iterations; i++) {
    const startTime = performance.now();
    
    const { unmount } = renderWithProviders(component);
    
    // Wait for component to fully render
    await waitFor(() => {
      expect(document.body).toBeInTheDocument();
    }, { timeout });
    
    const endTime = performance.now();
    metrics.push(endTime - startTime);
    
    unmount();
  }
  
  const avgRenderTime = metrics.reduce((sum, time) => sum + time, 0) / metrics.length;
  const maxRenderTime = Math.max(...metrics);
  const minRenderTime = Math.min(...metrics);
  
  return {
    avgRenderTime,
    maxRenderTime,
    minRenderTime,
    metrics,
  };
};

// Memory leak testing
export const testMemoryLeaks = async (component, options = {}) => {
  const { iterations = 50 } = options;
  const initialMemory = performance.memory?.usedJSHeapSize || 0;
  
  for (let i = 0; i < iterations; i++) {
    const { unmount } = renderWithProviders(component);
    unmount();
    
    // Force garbage collection if available
    if (global.gc) {
      global.gc();
    }
  }
  
  const finalMemory = performance.memory?.usedJSHeapSize || 0;
  const memoryIncrease = finalMemory - initialMemory;
  
  return {
    initialMemory,
    finalMemory,
    memoryIncrease,
    hasLeak: memoryIncrease > 1024 * 1024, // 1MB threshold
  };
};

// Security testing utilities
export const testSecurity = async (component) => {
  renderWithProviders(component);
  
  const securityTests = {
    xssProtection: testXSSProtection(),
    csrfProtection: testCSRFProtection(),
    inputValidation: testInputValidation(),
    urlValidation: testURLValidation(),
  };
  
  return securityTests;
};

// XSS protection testing
const testXSSProtection = () => {
  const maliciousInputs = [
    '<script>alert("xss")</script>',
    'javascript:alert("xss")',
    '<img src="x" onerror="alert(\'xss\')" />',
    '<svg onload="alert(\'xss\')" />',
  ];
  
  const results = maliciousInputs.map(input => {
    const sanitized = securityService.sanitizeInput(input);
    return {
      input,
      sanitized,
      safe: !sanitized.includes('<script>') && !sanitized.includes('javascript:'),
    };
  });
  
  return {
    passed: results.every(result => result.safe),
    results,
  };
};

// CSRF protection testing
const testCSRFProtection = async () => {
  try {
    const token = await securityService.getCSRFToken();
    const options = await securityService.addCSRFToken({});
    
    return {
      passed: !!token && !!options.headers['X-CSRF-Token'],
      token: !!token,
      headerAdded: !!options.headers['X-CSRF-Token'],
    };
  } catch (error) {
    return {
      passed: false,
      error: error.message,
    };
  }
};

// Input validation testing
const testInputValidation = () => {
  const validator = securityService.createValidator({
    email: {
      required: true,
      type: 'string',
      pattern: securityService.getValidationPatterns().email,
    },
    password: {
      required: true,
      type: 'string',
      minLength: 8,
      pattern: securityService.getValidationPatterns().password,
    },
  });
  
  const testCases = [
    { input: { email: 'test@example.com', password: 'Password123!' }, shouldPass: true },
    { input: { email: 'invalid-email', password: 'weak' }, shouldPass: false },
    { input: { email: '', password: '' }, shouldPass: false },
  ];
  
  const results = testCases.map(testCase => {
    const validation = validator(testCase.input);
    return {
      ...testCase,
      result: validation,
      passed: validation.valid === testCase.shouldPass,
    };
  });
  
  return {
    passed: results.every(result => result.passed),
    results,
  };
};

// URL validation testing
const testURLValidation = () => {
  const testUrls = [
    { url: 'https://example.com', shouldPass: true },
    { url: 'http://example.com', shouldPass: true },
    { url: 'javascript:alert("xss")', shouldPass: false },
    { url: 'data:text/html,<script>alert("xss")</script>', shouldPass: false },
  ];
  
  const results = testUrls.map(testCase => {
    const isValid = securityService.validateURL(testCase.url);
    return {
      ...testCase,
      isValid,
      passed: isValid === testCase.shouldPass,
    };
  });
  
  return {
    passed: results.every(result => result.passed),
    results,
  };
};

// Integration testing utilities
export const testIntegration = async (components, scenario) => {
  const results = [];
  
  for (const step of scenario.steps) {
    const { action, component, target, expected } = step;
    
    switch (action) {
      case 'render':
        renderWithProviders(components[component]);
        break;
        
      case 'click':
        const element = screen.getByTestId(target) || screen.getByRole('button', { name: target });
        fireEvent.click(element);
        break;
        
      case 'type':
        const input = screen.getByLabelText(target) || screen.getByPlaceholderText(target);
        await userEvent.type(input, step.value);
        break;
        
      case 'wait':
        await waitFor(() => {
          expect(screen.getByText(expected)).toBeInTheDocument();
        });
        break;
        
      case 'assert':
        if (expected.type === 'text') {
          expect(screen.getByText(expected.value)).toBeInTheDocument();
        } else if (expected.type === 'element') {
          expect(screen.getByTestId(expected.value)).toBeInTheDocument();
        }
        break;
    }
    
    results.push({
      step: step.name,
      passed: true,
      timestamp: Date.now(),
    });
  }
  
  return {
    passed: results.every(result => result.passed),
    results,
    scenario: scenario.name,
  };
};

// Mock data generators
export const generateMockData = {
  user: (overrides = {}) => ({
    id: '1',
    email: 'test@example.com',
    firstName: 'Test',
    lastName: 'User',
    role: 'user',
    ...overrides,
  }),
  
  dashboardData: (overrides = {}) => ({
    clusterStatus: 'healthy',
    nodeCount: 3,
    activeModels: 5,
    totalRequests: 1247,
    avgResponseTime: 245,
    errorRate: 0.02,
    uptime: '99.9%',
    nodes: [
      { id: 'node-1', status: 'healthy', cpu: 45, memory: 67, requests: 423 },
      { id: 'node-2', status: 'healthy', cpu: 52, memory: 71, requests: 389 },
      { id: 'node-3', status: 'warning', cpu: 78, memory: 89, requests: 435 },
    ],
    ...overrides,
  }),
  
  apiResponse: (data, overrides = {}) => ({
    success: true,
    data,
    message: 'Success',
    timestamp: new Date().toISOString(),
    ...overrides,
  }),
};

// Test suite runner
export const runTestSuite = async (component, options = {}) => {
  const {
    accessibility = true,
    performance = true,
    security = true,
    keyboard = true,
    screenReader = true,
    memoryLeaks = false,
  } = options;
  
  const results = {
    component: component.name || 'Unknown',
    timestamp: new Date().toISOString(),
    tests: {},
  };
  
  try {
    if (accessibility) {
      results.tests.accessibility = await testAccessibility(component);
    }
    
    if (keyboard) {
      results.tests.keyboard = await testKeyboardNavigation(component);
    }
    
    if (screenReader) {
      results.tests.screenReader = testScreenReaderSupport(component);
    }
    
    if (performance) {
      results.tests.performance = await testPerformance(component);
    }
    
    if (security) {
      results.tests.security = await testSecurity(component);
    }
    
    if (memoryLeaks) {
      results.tests.memoryLeaks = await testMemoryLeaks(component);
    }
    
    results.passed = Object.values(results.tests).every(test => 
      test.passed !== false && (!test.violations || test.violations.length === 0)
    );
    
  } catch (error) {
    results.error = error.message;
    results.passed = false;
  }
  
  return results;
};

export default {
  renderWithProviders,
  testAccessibility,
  testKeyboardNavigation,
  testScreenReaderSupport,
  testPerformance,
  testMemoryLeaks,
  testSecurity,
  testIntegration,
  generateMockData,
  runTestSuite,
};

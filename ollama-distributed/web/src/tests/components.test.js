/**
 * Component Test Suite
 * 
 * Comprehensive tests for all design system components and application components.
 */

import React from 'react';
import { screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

// Components to test
import { Button, Input, Card, Badge, Modal, Toast } from '../design-system/index.js';
import { SkipLink, FocusTrap } from '../design-system/index.js';
import RealTimeDashboard from '../components/dashboard/RealTimeDashboard.jsx';
import AccessibleDashboard from '../components/dashboard/AccessibleDashboard.jsx';
import MobileDashboard from '../components/mobile/MobileDashboard.jsx';
import AuthPage from '../components/auth/AuthPage.jsx';

// Test utilities
import {
  renderWithProviders,
  testAccessibility,
  testKeyboardNavigation,
  testScreenReaderSupport,
  testPerformance,
  testSecurity,
  generateMockData,
  runTestSuite,
} from './testUtils.js';

// Mock services
jest.mock('../services/authService.js');
jest.mock('../services/pwaService.js');
jest.mock('../services/accessibilityService.js');
jest.mock('../services/i18nService.js');

describe('Design System Components', () => {
  describe('Button Component', () => {
    test('renders with correct text', () => {
      renderWithProviders(<Button>Click me</Button>);
      expect(screen.getByRole('button', { name: 'Click me' })).toBeInTheDocument();
    });

    test('handles click events', async () => {
      const handleClick = jest.fn();
      renderWithProviders(<Button onClick={handleClick}>Click me</Button>);
      
      const button = screen.getByRole('button');
      await userEvent.click(button);
      
      expect(handleClick).toHaveBeenCalledTimes(1);
    });

    test('supports different variants', () => {
      const { rerender } = renderWithProviders(<Button variant="primary">Primary</Button>);
      expect(screen.getByRole('button')).toHaveStyle({ backgroundColor: expect.any(String) });
      
      rerender(<Button variant="secondary">Secondary</Button>);
      expect(screen.getByRole('button')).toHaveStyle({ backgroundColor: expect.any(String) });
    });

    test('handles disabled state', () => {
      renderWithProviders(<Button disabled>Disabled</Button>);
      const button = screen.getByRole('button');
      
      expect(button).toBeDisabled();
      expect(button).toHaveAttribute('aria-disabled', 'true');
    });

    test('handles loading state', () => {
      renderWithProviders(<Button loading>Loading</Button>);
      const button = screen.getByRole('button');
      
      expect(button).toBeDisabled();
      expect(button).toHaveAttribute('aria-busy', 'true');
    });

    test('passes accessibility tests', async () => {
      await testAccessibility(<Button>Accessible Button</Button>);
    });

    test('supports keyboard navigation', async () => {
      const handleClick = jest.fn();
      renderWithProviders(<Button onClick={handleClick}>Keyboard Button</Button>);
      
      const button = screen.getByRole('button');
      button.focus();
      
      await userEvent.keyboard('{Enter}');
      expect(handleClick).toHaveBeenCalledTimes(1);
      
      await userEvent.keyboard(' ');
      expect(handleClick).toHaveBeenCalledTimes(2);
    });
  });

  describe('Input Component', () => {
    test('renders with label', () => {
      renderWithProviders(<Input label="Test Input" />);
      expect(screen.getByLabelText('Test Input')).toBeInTheDocument();
    });

    test('handles value changes', async () => {
      const handleChange = jest.fn();
      renderWithProviders(<Input label="Test Input" onChange={handleChange} />);
      
      const input = screen.getByLabelText('Test Input');
      await userEvent.type(input, 'test value');
      
      expect(handleChange).toHaveBeenCalled();
      expect(input).toHaveValue('test value');
    });

    test('shows validation errors', () => {
      renderWithProviders(<Input label="Test Input" error="This field is required" />);
      expect(screen.getByText('This field is required')).toBeInTheDocument();
    });

    test('supports different types', () => {
      const { rerender } = renderWithProviders(<Input type="email" label="Email" />);
      expect(screen.getByLabelText('Email')).toHaveAttribute('type', 'email');
      
      rerender(<Input type="password" label="Password" />);
      expect(screen.getByLabelText('Password')).toHaveAttribute('type', 'password');
    });

    test('passes accessibility tests', async () => {
      await testAccessibility(<Input label="Accessible Input" />);
    });
  });

  describe('Card Component', () => {
    test('renders children content', () => {
      renderWithProviders(
        <Card>
          <h2>Card Title</h2>
          <p>Card content</p>
        </Card>
      );
      
      expect(screen.getByText('Card Title')).toBeInTheDocument();
      expect(screen.getByText('Card content')).toBeInTheDocument();
    });

    test('supports different variants', () => {
      const { rerender } = renderWithProviders(<Card variant="elevated">Elevated</Card>);
      expect(screen.getByText('Elevated').closest('div')).toHaveStyle({ boxShadow: expect.any(String) });
      
      rerender(<Card variant="outlined">Outlined</Card>);
      expect(screen.getByText('Outlined').closest('div')).toHaveStyle({ border: expect.any(String) });
    });

    test('handles interactive state', async () => {
      const handleClick = jest.fn();
      renderWithProviders(
        <Card interactive onClick={handleClick}>
          Interactive Card
        </Card>
      );
      
      const card = screen.getByText('Interactive Card').closest('div');
      await userEvent.click(card);
      
      expect(handleClick).toHaveBeenCalledTimes(1);
    });

    test('passes accessibility tests', async () => {
      await testAccessibility(<Card>Accessible Card</Card>);
    });
  });

  describe('Badge Component', () => {
    test('renders with text', () => {
      renderWithProviders(<Badge>New</Badge>);
      expect(screen.getByText('New')).toBeInTheDocument();
    });

    test('supports different variants', () => {
      const { rerender } = renderWithProviders(<Badge variant="success">Success</Badge>);
      expect(screen.getByText('Success')).toHaveStyle({ backgroundColor: expect.any(String) });
      
      rerender(<Badge variant="error">Error</Badge>);
      expect(screen.getByText('Error')).toHaveStyle({ backgroundColor: expect.any(String) });
    });

    test('supports dot variant', () => {
      renderWithProviders(<Badge dot />);
      const badge = document.querySelector('[style*="border-radius"]');
      expect(badge).toBeInTheDocument();
    });

    test('passes accessibility tests', async () => {
      await testAccessibility(<Badge>Accessible Badge</Badge>);
    });
  });

  describe('Modal Component', () => {
    test('renders when open', () => {
      renderWithProviders(
        <Modal isOpen onClose={() => {}} title="Test Modal">
          Modal content
        </Modal>
      );
      
      expect(screen.getByRole('dialog')).toBeInTheDocument();
      expect(screen.getByText('Test Modal')).toBeInTheDocument();
      expect(screen.getByText('Modal content')).toBeInTheDocument();
    });

    test('does not render when closed', () => {
      renderWithProviders(
        <Modal isOpen={false} onClose={() => {}} title="Test Modal">
          Modal content
        </Modal>
      );
      
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });

    test('handles close events', async () => {
      const handleClose = jest.fn();
      renderWithProviders(
        <Modal isOpen onClose={handleClose} title="Test Modal">
          Modal content
        </Modal>
      );
      
      const closeButton = screen.getByLabelText('Close modal');
      await userEvent.click(closeButton);
      
      expect(handleClose).toHaveBeenCalledTimes(1);
    });

    test('handles escape key', async () => {
      const handleClose = jest.fn();
      renderWithProviders(
        <Modal isOpen onClose={handleClose} title="Test Modal">
          Modal content
        </Modal>
      );
      
      await userEvent.keyboard('{Escape}');
      expect(handleClose).toHaveBeenCalledTimes(1);
    });

    test('traps focus', async () => {
      renderWithProviders(
        <Modal isOpen onClose={() => {}} title="Test Modal">
          <button>First Button</button>
          <button>Second Button</button>
        </Modal>
      );
      
      const firstButton = screen.getByText('First Button');
      const secondButton = screen.getByText('Second Button');
      const closeButton = screen.getByLabelText('Close modal');
      
      // Focus should start on first focusable element
      expect(document.activeElement).toBe(firstButton);
      
      // Tab should cycle through focusable elements
      await userEvent.tab();
      expect(document.activeElement).toBe(secondButton);
      
      await userEvent.tab();
      expect(document.activeElement).toBe(closeButton);
      
      await userEvent.tab();
      expect(document.activeElement).toBe(firstButton);
    });

    test('passes accessibility tests', async () => {
      await testAccessibility(
        <Modal isOpen onClose={() => {}} title="Accessible Modal">
          Modal content
        </Modal>
      );
    });
  });

  describe('SkipLink Component', () => {
    test('renders with default text', () => {
      renderWithProviders(<SkipLink />);
      const skipLink = screen.getByText('Skip to main content');
      expect(skipLink).toBeInTheDocument();
    });

    test('renders with custom text', () => {
      renderWithProviders(<SkipLink>Skip to navigation</SkipLink>);
      expect(screen.getByText('Skip to navigation')).toBeInTheDocument();
    });

    test('handles click events', async () => {
      // Create a target element
      document.body.innerHTML = '<main id="main-content">Main content</main>';
      
      renderWithProviders(<SkipLink href="#main-content" />);
      
      const skipLink = screen.getByText('Skip to main content');
      await userEvent.click(skipLink);
      
      const mainContent = document.getElementById('main-content');
      expect(document.activeElement).toBe(mainContent);
    });

    test('passes accessibility tests', async () => {
      await testAccessibility(<SkipLink />);
    });
  });

  describe('FocusTrap Component', () => {
    test('traps focus when active', async () => {
      renderWithProviders(
        <FocusTrap active>
          <button>First Button</button>
          <button>Second Button</button>
        </FocusTrap>
      );
      
      const firstButton = screen.getByText('First Button');
      const secondButton = screen.getByText('Second Button');
      
      // Focus should start on first element
      expect(document.activeElement).toBe(firstButton);
      
      // Tab should cycle between elements
      await userEvent.tab();
      expect(document.activeElement).toBe(secondButton);
      
      await userEvent.tab();
      expect(document.activeElement).toBe(firstButton);
    });

    test('does not trap focus when inactive', async () => {
      renderWithProviders(
        <FocusTrap active={false}>
          <button>Trapped Button</button>
        </FocusTrap>
      );
      
      const button = screen.getByText('Trapped Button');
      button.focus();
      
      await userEvent.tab();
      // Focus should move outside the trap
      expect(document.activeElement).not.toBe(button);
    });

    test('passes accessibility tests', async () => {
      await testAccessibility(
        <FocusTrap active>
          <button>Focus Trap Button</button>
        </FocusTrap>
      );
    });
  });
});

describe('Application Components', () => {
  describe('AuthPage Component', () => {
    test('renders login form by default', () => {
      renderWithProviders(<AuthPage />);
      expect(screen.getByText('Sign In')).toBeInTheDocument();
      expect(screen.getByLabelText('Email Address')).toBeInTheDocument();
      expect(screen.getByLabelText('Password')).toBeInTheDocument();
    });

    test('switches to registration form', async () => {
      renderWithProviders(<AuthPage />);
      
      const signUpLink = screen.getByText("Don't have an account?");
      await userEvent.click(signUpLink);
      
      expect(screen.getByText('Sign Up')).toBeInTheDocument();
      expect(screen.getByLabelText('First Name')).toBeInTheDocument();
      expect(screen.getByLabelText('Last Name')).toBeInTheDocument();
    });

    test('passes accessibility tests', async () => {
      await testAccessibility(<AuthPage />);
    });
  });

  describe('RealTimeDashboard Component', () => {
    test('renders dashboard content', () => {
      const mockUser = generateMockData.user();
      renderWithProviders(<RealTimeDashboard />, { user: mockUser });
      
      expect(screen.getByText(`Welcome back, ${mockUser.firstName}! 👋`)).toBeInTheDocument();
      expect(screen.getByText('Real-Time Dashboard')).toBeInTheDocument();
    });

    test('displays metrics', () => {
      const mockUser = generateMockData.user();
      renderWithProviders(<RealTimeDashboard />, { user: mockUser });
      
      expect(screen.getByText('Total Requests')).toBeInTheDocument();
      expect(screen.getByText('Response Time')).toBeInTheDocument();
      expect(screen.getByText('Error Rate')).toBeInTheDocument();
      expect(screen.getByText('Uptime')).toBeInTheDocument();
    });

    test('passes accessibility tests', async () => {
      const mockUser = generateMockData.user();
      await testAccessibility(<RealTimeDashboard />, { user: mockUser });
    });
  });

  describe('AccessibleDashboard Component', () => {
    test('includes skip links', () => {
      const mockUser = generateMockData.user();
      renderWithProviders(<AccessibleDashboard />, { user: mockUser });
      
      expect(screen.getByText('Skip to main content')).toBeInTheDocument();
    });

    test('has proper ARIA labels', () => {
      const mockUser = generateMockData.user();
      renderWithProviders(<AccessibleDashboard />, { user: mockUser });
      
      expect(screen.getByRole('main')).toHaveAttribute('aria-label', 'Dashboard content');
      expect(screen.getByRole('navigation')).toHaveAttribute('aria-label', 'Main navigation');
    });

    test('passes accessibility tests', async () => {
      const mockUser = generateMockData.user();
      await testAccessibility(<AccessibleDashboard />, { user: mockUser });
    });
  });
});

describe('Performance Tests', () => {
  test('Button component performance', async () => {
    const results = await testPerformance(<Button>Performance Test</Button>);
    expect(results.avgRenderTime).toBeLessThan(50); // 50ms threshold
  });

  test('Dashboard component performance', async () => {
    const mockUser = generateMockData.user();
    const results = await testPerformance(<RealTimeDashboard />, { user: mockUser });
    expect(results.avgRenderTime).toBeLessThan(200); // 200ms threshold
  });
});

describe('Security Tests', () => {
  test('Input component XSS protection', async () => {
    const results = await testSecurity(<Input label="Security Test" />);
    expect(results.xssProtection.passed).toBe(true);
  });

  test('Button component security', async () => {
    const results = await testSecurity(<Button>Security Test</Button>);
    expect(results.csrfProtection.passed).toBe(true);
  });
});

describe('Integration Tests', () => {
  test('Complete authentication flow', async () => {
    const scenario = {
      name: 'Authentication Flow',
      steps: [
        { name: 'Render auth page', action: 'render', component: 'authPage' },
        { name: 'Enter email', action: 'type', target: 'Email Address', value: 'test@example.com' },
        { name: 'Enter password', action: 'type', target: 'Password', value: 'Password123!' },
        { name: 'Click login', action: 'click', target: 'Sign In' },
        { name: 'Wait for dashboard', action: 'wait', expected: 'Welcome back' },
      ],
    };

    const components = {
      authPage: <AuthPage />,
    };

    const results = await testIntegration(components, scenario);
    expect(results.passed).toBe(true);
  });
});

describe('Comprehensive Test Suite', () => {
  test('Run full test suite on Button component', async () => {
    const results = await runTestSuite(<Button>Test Button</Button>);
    expect(results.passed).toBe(true);
    expect(results.tests.accessibility).toBeDefined();
    expect(results.tests.keyboard).toBeDefined();
    expect(results.tests.performance).toBeDefined();
    expect(results.tests.security).toBeDefined();
  });

  test('Run full test suite on Modal component', async () => {
    const results = await runTestSuite(
      <Modal isOpen onClose={() => {}} title="Test Modal">
        Modal content
      </Modal>
    );
    expect(results.passed).toBe(true);
  });
});

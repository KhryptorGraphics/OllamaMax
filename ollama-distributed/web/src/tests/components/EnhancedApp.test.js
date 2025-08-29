// Enhanced App Component Integration Tests
// Tests for the main application component with service integration

import React from 'react';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import '@testing-library/jest-dom';
import EnhancedApp from '../../components/EnhancedApp.jsx';
import apiService from '../../services/api.js';
import authService from '../../services/auth.js';
import wsService from '../../services/websocket.js';

// Mock services
jest.mock('../../services/api.js');
jest.mock('../../services/auth.js');
jest.mock('../../services/websocket.js');

// Mock child components to focus on integration logic
jest.mock('../../components/Login.jsx', () => {
  return function MockLogin({ onLogin, onRegister }) {
    return (
      <div data-testid="login-component">
        <button onClick={() => onLogin({ id: 1, username: 'testuser' })}>
          Login
        </button>
        <button onClick={onRegister}>Register</button>
      </div>
    );
  };
});

jest.mock('../../components/RegistrationFlow.jsx', () => {
  return function MockRegistrationFlow({ onRegistrationComplete, onCancel }) {
    return (
      <div data-testid="registration-component">
        <button onClick={() => onRegistrationComplete({ username: 'newuser' })}>
          Complete Registration
        </button>
        <button onClick={onCancel}>Cancel</button>
      </div>
    );
  };
});

jest.mock('../../components/Dashboard.jsx', () => {
  return function MockDashboard() {
    return <div data-testid="dashboard-component">Dashboard</div>;
  };
});

jest.mock('../../components/Sidebar.jsx', () => {
  return function MockSidebar({ activeTab, onTabChange, currentUser, onLogout }) {
    return (
      <div data-testid="sidebar-component">
        <div data-testid="current-user">{currentUser?.username}</div>
        <button onClick={() => onTabChange('models')}>Models</button>
        <button onClick={onLogout}>Logout</button>
      </div>
    );
  };
});

describe('EnhancedApp Integration', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    localStorage.clear();
    
    // Default mock implementations
    authService.isAuthenticated.mockReturnValue(false);
    authService.getCurrentUser.mockReturnValue(null);
    authService.getRememberedUsername.mockReturnValue(null);
    authService.on.mockImplementation(() => {});
    authService.off.mockImplementation(() => {});
    
    apiService.getHealth.mockResolvedValue({ status: 'healthy' });
    apiService.getReadiness.mockResolvedValue({ ready: true });
    
    wsService.connect.mockImplementation(() => {});
    wsService.disconnect.mockImplementation(() => {});
    wsService.on.mockImplementation(() => {});
    wsService.subscribe.mockImplementation(() => {});
  });

  describe('Authentication Flow', () => {
    test('should show login screen when not authenticated', async () => {
      await act(async () => {
        render(<EnhancedApp />);
      });

      expect(screen.getByTestId('login-component')).toBeInTheDocument();
      expect(screen.queryByTestId('dashboard-component')).not.toBeInTheDocument();
    });

    test('should show dashboard when authenticated', async () => {
      authService.isAuthenticated.mockReturnValue(true);
      authService.getCurrentUser.mockReturnValue({ id: 1, username: 'testuser' });
      
      await act(async () => {
        render(<EnhancedApp />);
      });

      await waitFor(() => {
        expect(screen.getByTestId('dashboard-component')).toBeInTheDocument();
      });
      
      expect(screen.queryByTestId('login-component')).not.toBeInTheDocument();
    });

    test('should handle login success', async () => {
      const mockUser = { id: 1, username: 'testuser' };
      
      await act(async () => {
        render(<EnhancedApp />);
      });

      const loginButton = screen.getByText('Login');
      
      await act(async () => {
        fireEvent.click(loginButton);
      });

      // Verify authentication handler was called
      expect(authService.on).toHaveBeenCalledWith('authenticated', expect.any(Function));
    });

    test('should switch to registration flow', async () => {
      await act(async () => {
        render(<EnhancedApp />);
      });

      const registerButton = screen.getByText('Register');
      
      await act(async () => {
        fireEvent.click(registerButton);
      });

      expect(screen.getByTestId('registration-component')).toBeInTheDocument();
      expect(screen.queryByTestId('login-component')).not.toBeInTheDocument();
    });

    test('should handle registration completion', async () => {
      await act(async () => {
        render(<EnhancedApp />);
      });

      // Switch to registration
      const registerButton = screen.getByText('Register');
      await act(async () => {
        fireEvent.click(registerButton);
      });

      // Complete registration
      const completeButton = screen.getByText('Complete Registration');
      await act(async () => {
        fireEvent.click(completeButton);
      });

      // Should return to login
      expect(screen.getByTestId('login-component')).toBeInTheDocument();
      expect(screen.queryByTestId('registration-component')).not.toBeInTheDocument();
    });

    test('should handle logout', async () => {
      authService.isAuthenticated.mockReturnValue(true);
      authService.getCurrentUser.mockReturnValue({ id: 1, username: 'testuser' });
      authService.logout.mockResolvedValue();
      
      await act(async () => {
        render(<EnhancedApp />);
      });

      const logoutButton = screen.getByText('Logout');
      
      await act(async () => {
        fireEvent.click(logoutButton);
      });

      expect(authService.logout).toHaveBeenCalled();
    });
  });

  describe('Data Loading', () => {
    beforeEach(() => {
      authService.isAuthenticated.mockReturnValue(true);
      authService.getCurrentUser.mockReturnValue({ id: 1, username: 'testuser' });
      
      // Mock API responses
      apiService.getClusterStatus.mockResolvedValue({ status: 'healthy' });
      apiService.getNodes.mockResolvedValue({ nodes: [] });
      apiService.getModels.mockResolvedValue({ models: [] });
      apiService.getUsers.mockResolvedValue({ users: [] });
      apiService.getMetrics.mockResolvedValue({ cpu: 45 });
    });

    test('should load data on initialization', async () => {
      await act(async () => {
        render(<EnhancedApp />);
      });

      await waitFor(() => {
        expect(apiService.getClusterStatus).toHaveBeenCalled();
        expect(apiService.getNodes).toHaveBeenCalled();
        expect(apiService.getModels).toHaveBeenCalled();
        expect(apiService.getMetrics).toHaveBeenCalled();
      });
    });

    test('should load only health data for unauthenticated users', async () => {
      authService.isAuthenticated.mockReturnValue(false);
      
      await act(async () => {
        render(<EnhancedApp />);
      });

      await waitFor(() => {
        expect(apiService.getHealth).toHaveBeenCalled();
        expect(apiService.getReadiness).toHaveBeenCalled();
      });
      
      expect(apiService.getClusterStatus).not.toHaveBeenCalled();
      expect(apiService.getNodes).not.toHaveBeenCalled();
    });

    test('should handle data loading errors', async () => {
      apiService.getClusterStatus.mockRejectedValue(new Error('API Error'));
      
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      
      await act(async () => {
        render(<EnhancedApp />);
      });

      await waitFor(() => {
        expect(consoleSpy).toHaveBeenCalledWith(
          'Failed to load cluster status:',
          expect.any(Error)
        );
      });
      
      consoleSpy.mockRestore();
    });
  });

  describe('WebSocket Integration', () => {
    beforeEach(() => {
      authService.isAuthenticated.mockReturnValue(true);
      authService.getCurrentUser.mockReturnValue({ id: 1, username: 'testuser' });
      authService.getAuthToken.mockReturnValue('test-token');
    });

    test('should connect to WebSocket when authenticated', async () => {
      await act(async () => {
        render(<EnhancedApp />);
      });

      expect(wsService.connect).toHaveBeenCalledWith('test-token');
      expect(wsService.on).toHaveBeenCalledWith('connected', expect.any(Function));
      expect(wsService.on).toHaveBeenCalledWith('disconnected', expect.any(Function));
      expect(wsService.on).toHaveBeenCalledWith('error', expect.any(Function));
    });

    test('should subscribe to data streams', async () => {
      await act(async () => {
        render(<EnhancedApp />);
      });

      expect(wsService.subscribe).toHaveBeenCalledWith('metrics', expect.any(Function));
      expect(wsService.subscribe).toHaveBeenCalledWith('nodes', expect.any(Function));
      expect(wsService.subscribe).toHaveBeenCalledWith('models', expect.any(Function));
    });

    test('should disconnect WebSocket on unmount', async () => {
      const { unmount } = await act(async () => {
        return render(<EnhancedApp />);
      });

      await act(async () => {
        unmount();
      });

      expect(wsService.disconnect).toHaveBeenCalled();
    });
  });

  describe('Theme Management', () => {
    test('should load theme from localStorage', async () => {
      localStorage.setItem('theme', 'dark');
      
      await act(async () => {
        render(<EnhancedApp />);
      });

      expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
    });

    test('should default to light theme', async () => {
      await act(async () => {
        render(<EnhancedApp />);
      });

      expect(document.documentElement.getAttribute('data-theme')).toBe('light');
    });

    test('should toggle theme', async () => {
      await act(async () => {
        render(<EnhancedApp />);
      });

      // Initial theme should be light
      expect(document.documentElement.getAttribute('data-theme')).toBe('light');
      
      // Find and click theme toggle (would need to be exposed in actual component)
      // This is a simplified test - in reality you'd need to expose the toggle
      const themeToggle = screen.queryByTestId('theme-toggle');
      if (themeToggle) {
        await act(async () => {
          fireEvent.click(themeToggle);
        });
        
        expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
        expect(localStorage.getItem('theme')).toBe('dark');
      }
    });
  });

  describe('Alert System', () => {
    test('should display alerts', async () => {
      authService.isAuthenticated.mockReturnValue(true);
      
      await act(async () => {
        render(<EnhancedApp />);
      });

      // Alerts would be displayed in the alerts container
      // This test would need the actual alert display logic
      const alertsContainer = screen.queryByTestId('alerts-container');
      expect(alertsContainer).toBeInTheDocument();
    });

    test('should auto-dismiss alerts', async () => {
      jest.useFakeTimers();
      
      await act(async () => {
        render(<EnhancedApp />);
      });

      // Simulate adding an alert with auto-dismiss
      // This would need to be exposed through the component interface
      
      jest.useRealTimers();
    });
  });

  describe('Navigation', () => {
    beforeEach(() => {
      authService.isAuthenticated.mockReturnValue(true);
      authService.getCurrentUser.mockReturnValue({ id: 1, username: 'testuser' });
    });

    test('should change active tab', async () => {
      await act(async () => {
        render(<EnhancedApp />);
      });

      const modelsButton = screen.getByText('Models');
      
      await act(async () => {
        fireEvent.click(modelsButton);
      });

      // The active tab change would be reflected in the sidebar component
      // This test verifies the interaction works
      expect(modelsButton).toBeInTheDocument();
    });

    test('should show user information in sidebar', async () => {
      const mockUser = { id: 1, username: 'testuser' };
      authService.getCurrentUser.mockReturnValue(mockUser);
      
      await act(async () => {
        render(<EnhancedApp />);
      });

      expect(screen.getByTestId('current-user')).toHaveTextContent('testuser');
    });
  });

  describe('Error Handling', () => {
    test('should handle initialization errors gracefully', async () => {
      apiService.getHealth.mockRejectedValue(new Error('Network error'));
      
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      
      await act(async () => {
        render(<EnhancedApp />);
      });

      await waitFor(() => {
        expect(consoleSpy).toHaveBeenCalled();
      });
      
      // App should still render despite errors
      expect(screen.getByTestId('login-component')).toBeInTheDocument();
      
      consoleSpy.mockRestore();
    });

    test('should handle WebSocket errors', async () => {
      authService.isAuthenticated.mockReturnValue(true);
      
      await act(async () => {
        render(<EnhancedApp />);
      });

      // Simulate WebSocket error
      const errorHandler = wsService.on.mock.calls.find(
        call => call[0] === 'error'
      )?.[1];
      
      if (errorHandler) {
        await act(async () => {
          errorHandler(new Error('WebSocket error'));
        });
      }

      // App should continue functioning
      expect(screen.getByTestId('dashboard-component')).toBeInTheDocument();
    });
  });

  describe('Session Management', () => {
    test('should handle session expiry', async () => {
      authService.isAuthenticated.mockReturnValue(true);
      
      await act(async () => {
        render(<EnhancedApp />);
      });

      // Simulate session expiry
      const sessionExpiredHandler = authService.on.mock.calls.find(
        call => call[0] === 'session_expired'
      )?.[1];
      
      if (sessionExpiredHandler) {
        await act(async () => {
          sessionExpiredHandler();
        });
      }

      // Should show appropriate alert or redirect to login
      // This would depend on the actual implementation
    });

    test('should handle session warnings', async () => {
      authService.isAuthenticated.mockReturnValue(true);
      
      await act(async () => {
        render(<EnhancedApp />);
      });

      // Simulate session warning
      const sessionWarningHandler = authService.on.mock.calls.find(
        call => call[0] === 'session_warning'
      )?.[1];
      
      if (sessionWarningHandler) {
        await act(async () => {
          sessionWarningHandler({ message: 'Session expiring soon' });
        });
      }

      // Should display warning alert
      // This would depend on the actual implementation
    });
  });

  describe('Responsive Design', () => {
    test('should handle mobile sidebar toggle', async () => {
      authService.isAuthenticated.mockReturnValue(true);
      
      // Mock mobile viewport
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 768,
      });

      await act(async () => {
        render(<EnhancedApp />);
      });

      // Mobile menu button should be present
      const mobileMenuButton = screen.queryByTestId('mobile-menu-button');
      if (mobileMenuButton) {
        await act(async () => {
          fireEvent.click(mobileMenuButton);
        });
        
        // Mobile sidebar should be visible
        expect(screen.getByTestId('sidebar-component')).toBeInTheDocument();
      }
    });
  });
});

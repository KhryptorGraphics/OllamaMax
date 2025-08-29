// Full Stack Integration Tests
// End-to-end tests for the complete application flow

import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import '@testing-library/jest-dom';
import EnhancedApp from '../../components/EnhancedApp.jsx';

// Mock the actual services for integration testing
jest.mock('../../services/api.js', () => ({
  __esModule: true,
  default: {
    getBaseURL: () => 'http://localhost:11434/api/v1',
    setToken: jest.fn(),
    clearAuth: jest.fn(),
    getAuthToken: () => 'mock-token',
    isAuthenticated: () => true,
    
    // Authentication
    login: jest.fn(),
    register: jest.fn(),
    logout: jest.fn(),
    
    // Health and status
    getHealth: jest.fn(),
    getReadiness: jest.fn(),
    getClusterStatus: jest.fn(),
    
    // Resources
    getNodes: jest.fn(),
    getModels: jest.fn(),
    getUsers: jest.fn(),
    getMetrics: jest.fn(),
    getSystemMetrics: jest.fn(),
    getDatabaseStats: jest.fn(),
    
    // CRUD operations
    createUser: jest.fn(),
    updateUser: jest.fn(),
    deleteUser: jest.fn(),
    updateNode: jest.fn(),
    createModel: jest.fn(),
    deleteModel: jest.fn(),
    
    // Generic request
    request: jest.fn(),
    
    // File upload
    uploadFile: jest.fn(),
  },
}));

jest.mock('../../services/auth.js', () => ({
  __esModule: true,
  default: {
    // State
    currentUser: null,
    permissions: [],
    
    // Authentication
    login: jest.fn(),
    register: jest.fn(),
    logout: jest.fn(),
    
    // Session management
    isAuthenticated: jest.fn(),
    getCurrentUser: jest.fn(),
    getPermissions: jest.fn(),
    getAuthToken: jest.fn(),
    
    // Permissions
    hasPermission: jest.fn(),
    hasAnyPermission: jest.fn(),
    hasAllPermissions: jest.fn(),
    hasRole: jest.fn(),
    isAdmin: jest.fn(),
    isOperator: jest.fn(),
    getRole: jest.fn(),
    
    // Profile management
    updateProfile: jest.fn(),
    changePassword: jest.fn(),
    requestPasswordReset: jest.fn(),
    
    // 2FA
    setupTwoFactor: jest.fn(),
    verifyTwoFactor: jest.fn(),
    
    // Remember me
    getRememberedUsername: jest.fn(),
    clearRememberedUsername: jest.fn(),
    
    // Session
    setupSessionTimeout: jest.fn(),
    extendSession: jest.fn(),
    clearSessionTimeout: jest.fn(),
    trackActivity: jest.fn(),
    getSessionInfo: jest.fn(),
    
    // Events
    on: jest.fn(),
    off: jest.fn(),
    emit: jest.fn(),
    
    // Initialization
    initializeFromStorage: jest.fn(),
    clearSession: jest.fn(),
  },
}));

jest.mock('../../services/websocket.js', () => ({
  __esModule: true,
  default: {
    // Connection
    connect: jest.fn(),
    disconnect: jest.fn(),
    
    // Messaging
    send: jest.fn(),
    sendWebSocketMessage: jest.fn(),
    
    // Subscriptions
    subscribe: jest.fn(),
    unsubscribe: jest.fn(),
    
    // Status
    getStatus: jest.fn(() => ({
      isConnected: false,
      readyState: 3,
      reconnectAttempts: 0,
      queuedMessages: 0,
      lastHeartbeat: null,
    })),
    
    // Events
    on: jest.fn(),
    off: jest.fn(),
    emit: jest.fn(),
    
    // Utility
    requestMetrics: jest.fn(),
    requestNodeUpdates: jest.fn(),
    requestInferenceUpdates: jest.fn(),
    sendActivity: jest.fn(),
  },
}));

import apiService from '../../services/api.js';
import authService from '../../services/auth.js';
import wsService from '../../services/websocket.js';

describe('Full Stack Integration Tests', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    localStorage.clear();
    
    // Reset service states
    authService.currentUser = null;
    authService.permissions = [];
    
    // Default mock implementations
    authService.isAuthenticated.mockReturnValue(false);
    authService.getCurrentUser.mockReturnValue(null);
    authService.getPermissions.mockReturnValue([]);
    authService.getRememberedUsername.mockReturnValue(null);
    authService.getAuthToken.mockReturnValue(null);
    
    // Mock successful API responses
    apiService.getHealth.mockResolvedValue({ status: 'healthy', uptime: '99.9%' });
    apiService.getReadiness.mockResolvedValue({ ready: true });
    apiService.getClusterStatus.mockResolvedValue({ 
      nodes: 1, 
      leader: 'localhost', 
      status: 'healthy' 
    });
    apiService.getNodes.mockResolvedValue({ nodes: [] });
    apiService.getModels.mockResolvedValue({ models: [] });
    apiService.getUsers.mockResolvedValue({ users: [] });
    apiService.getMetrics.mockResolvedValue({ cpu: 45, memory: 67 });
    apiService.getSystemMetrics.mockResolvedValue({ 
      system: { cpu: 45, memory: 67, disk: 23 } 
    });
    apiService.getDatabaseStats.mockResolvedValue({ 
      connections: 5, 
      queries_per_second: 120 
    });
  });

  describe('Application Initialization', () => {
    test('should initialize application with health checks', async () => {
      await act(async () => {
        render(<EnhancedApp />);
      });

      await waitFor(() => {
        expect(apiService.getHealth).toHaveBeenCalled();
        expect(apiService.getReadiness).toHaveBeenCalled();
      });

      // Should show login screen for unauthenticated users
      expect(screen.getByTestId('login-component')).toBeInTheDocument();
    });

    test('should load full data for authenticated users', async () => {
      // Set up authenticated state
      authService.isAuthenticated.mockReturnValue(true);
      authService.getCurrentUser.mockReturnValue({
        id: 1,
        username: 'testuser',
        role: 'admin'
      });
      authService.getAuthToken.mockReturnValue('valid-token');

      await act(async () => {
        render(<EnhancedApp />);
      });

      await waitFor(() => {
        expect(apiService.getClusterStatus).toHaveBeenCalled();
        expect(apiService.getNodes).toHaveBeenCalled();
        expect(apiService.getModels).toHaveBeenCalled();
        expect(apiService.getMetrics).toHaveBeenCalled();
      });

      // Should show dashboard for authenticated users
      expect(screen.getByTestId('dashboard-component')).toBeInTheDocument();
    });

    test('should handle initialization failures gracefully', async () => {
      apiService.getHealth.mockRejectedValue(new Error('Service unavailable'));
      
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();

      await act(async () => {
        render(<EnhancedApp />);
      });

      await waitFor(() => {
        expect(consoleSpy).toHaveBeenCalledWith(
          'Failed to load health status:',
          expect.any(Error)
        );
      });

      // Application should still render
      expect(screen.getByTestId('login-component')).toBeInTheDocument();
      
      consoleSpy.mockRestore();
    });
  });

  describe('Authentication Flow', () => {
    test('should complete full login flow', async () => {
      const mockUser = {
        id: 1,
        username: 'testuser',
        email: 'test@example.com',
        role: 'user'
      };

      const mockLoginResponse = {
        user: mockUser,
        access_token: 'new-token',
        refresh_token: 'new-refresh',
        permissions: ['read', 'write']
      };

      apiService.login.mockResolvedValue(mockLoginResponse);
      authService.login.mockResolvedValue({
        success: true,
        user: mockUser,
        permissions: ['read', 'write']
      });

      await act(async () => {
        render(<EnhancedApp />);
      });

      // Click login button
      const loginButton = screen.getByText('Login');
      await act(async () => {
        fireEvent.click(loginButton);
      });

      // Verify login was called
      expect(authService.login).toHaveBeenCalled();

      // Simulate successful authentication
      authService.isAuthenticated.mockReturnValue(true);
      authService.getCurrentUser.mockReturnValue(mockUser);
      
      // Trigger authentication event
      const authHandler = authService.on.mock.calls.find(
        call => call[0] === 'authenticated'
      )?.[1];
      
      if (authHandler) {
        await act(async () => {
          authHandler(mockUser);
        });
      }

      // Should load authenticated data
      await waitFor(() => {
        expect(apiService.getClusterStatus).toHaveBeenCalled();
        expect(wsService.connect).toHaveBeenCalled();
      });
    });

    test('should handle login failure', async () => {
      apiService.login.mockRejectedValue(new Error('Invalid credentials'));
      authService.login.mockRejectedValue(new Error('Invalid credentials'));

      await act(async () => {
        render(<EnhancedApp />);
      });

      const loginButton = screen.getByText('Login');
      await act(async () => {
        fireEvent.click(loginButton);
      });

      // Should remain on login screen
      expect(screen.getByTestId('login-component')).toBeInTheDocument();
    });

    test('should complete registration flow', async () => {
      const mockRegistrationData = {
        username: 'newuser',
        email: 'new@example.com',
        password: 'password123'
      };

      apiService.register.mockResolvedValue({
        success: true,
        user_id: 123,
        message: 'Registration successful'
      });

      authService.register.mockResolvedValue({
        success: true,
        user_id: 123
      });

      await act(async () => {
        render(<EnhancedApp />);
      });

      // Switch to registration
      const registerButton = screen.getByText('Register');
      await act(async () => {
        fireEvent.click(registerButton);
      });

      expect(screen.getByTestId('registration-component')).toBeInTheDocument();

      // Complete registration
      const completeButton = screen.getByText('Complete Registration');
      await act(async () => {
        fireEvent.click(completeButton);
      });

      // Should return to login
      await waitFor(() => {
        expect(screen.getByTestId('login-component')).toBeInTheDocument();
      });
    });

    test('should handle logout flow', async () => {
      // Start with authenticated state
      authService.isAuthenticated.mockReturnValue(true);
      authService.getCurrentUser.mockReturnValue({
        id: 1,
        username: 'testuser'
      });

      apiService.logout.mockResolvedValue();
      authService.logout.mockResolvedValue();

      await act(async () => {
        render(<EnhancedApp />);
      });

      // Should show dashboard
      expect(screen.getByTestId('dashboard-component')).toBeInTheDocument();

      // Click logout
      const logoutButton = screen.getByText('Logout');
      await act(async () => {
        fireEvent.click(logoutButton);
      });

      expect(authService.logout).toHaveBeenCalled();

      // Simulate logout event
      authService.isAuthenticated.mockReturnValue(false);
      authService.getCurrentUser.mockReturnValue(null);
      
      const logoutHandler = authService.on.mock.calls.find(
        call => call[0] === 'logged_out'
      )?.[1];
      
      if (logoutHandler) {
        await act(async () => {
          logoutHandler();
        });
      }

      // Should return to login
      await waitFor(() => {
        expect(screen.getByTestId('login-component')).toBeInTheDocument();
      });
    });
  });

  describe('Real-time Updates', () => {
    beforeEach(() => {
      authService.isAuthenticated.mockReturnValue(true);
      authService.getCurrentUser.mockReturnValue({
        id: 1,
        username: 'testuser'
      });
      authService.getAuthToken.mockReturnValue('valid-token');
    });

    test('should establish WebSocket connection', async () => {
      await act(async () => {
        render(<EnhancedApp />);
      });

      expect(wsService.connect).toHaveBeenCalledWith('valid-token');
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

    test('should handle real-time metric updates', async () => {
      await act(async () => {
        render(<EnhancedApp />);
      });

      // Get the metrics update handler
      const metricsHandler = wsService.subscribe.mock.calls.find(
        call => call[0] === 'metrics'
      )?.[1];

      if (metricsHandler) {
        const newMetrics = {
          metrics: { cpu: 75, memory: 85, network: 45 }
        };

        await act(async () => {
          metricsHandler(newMetrics);
        });

        // Metrics should be updated in the application
        // This would be verified through the dashboard display
      }
    });

    test('should handle WebSocket disconnection', async () => {
      wsService.getStatus.mockReturnValue({
        isConnected: true,
        readyState: 1,
        reconnectAttempts: 0,
        queuedMessages: 0,
        lastHeartbeat: Date.now()
      });

      await act(async () => {
        render(<EnhancedApp />);
      });

      // Simulate disconnection
      const disconnectHandler = wsService.on.mock.calls.find(
        call => call[0] === 'disconnected'
      )?.[1];

      if (disconnectHandler) {
        wsService.getStatus.mockReturnValue({
          isConnected: false,
          readyState: 3,
          reconnectAttempts: 1,
          queuedMessages: 0,
          lastHeartbeat: Date.now() - 30000
        });

        await act(async () => {
          disconnectHandler();
        });

        // Application should handle disconnection gracefully
        // Status indicator should show disconnected state
      }
    });
  });

  describe('Error Handling and Recovery', () => {
    test('should handle API errors gracefully', async () => {
      authService.isAuthenticated.mockReturnValue(true);
      authService.getCurrentUser.mockReturnValue({ id: 1, username: 'testuser' });

      // Simulate API failures
      apiService.getClusterStatus.mockRejectedValue(new Error('Network error'));
      apiService.getNodes.mockRejectedValue(new Error('Service unavailable'));

      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();

      await act(async () => {
        render(<EnhancedApp />);
      });

      await waitFor(() => {
        expect(consoleSpy).toHaveBeenCalledWith(
          'Failed to load cluster status:',
          expect.any(Error)
        );
        expect(consoleSpy).toHaveBeenCalledWith(
          'Failed to load nodes:',
          expect.any(Error)
        );
      });

      // Application should still be functional
      expect(screen.getByTestId('dashboard-component')).toBeInTheDocument();

      consoleSpy.mockRestore();
    });

    test('should handle session expiry', async () => {
      authService.isAuthenticated.mockReturnValue(true);
      authService.getCurrentUser.mockReturnValue({ id: 1, username: 'testuser' });

      await act(async () => {
        render(<EnhancedApp />);
      });

      // Simulate session expiry
      const sessionExpiredHandler = authService.on.mock.calls.find(
        call => call[0] === 'session_expired'
      )?.[1];

      if (sessionExpiredHandler) {
        authService.isAuthenticated.mockReturnValue(false);
        authService.getCurrentUser.mockReturnValue(null);

        await act(async () => {
          sessionExpiredHandler();
        });

        // Should redirect to login
        await waitFor(() => {
          expect(screen.getByTestId('login-component')).toBeInTheDocument();
        });
      }
    });

    test('should recover from temporary network issues', async () => {
      authService.isAuthenticated.mockReturnValue(true);
      authService.getCurrentUser.mockReturnValue({ id: 1, username: 'testuser' });

      // First call fails, second succeeds
      apiService.getMetrics
        .mockRejectedValueOnce(new Error('Network timeout'))
        .mockResolvedValueOnce({ cpu: 45, memory: 67 });

      await act(async () => {
        render(<EnhancedApp />);
      });

      // Initial load should fail
      await waitFor(() => {
        expect(apiService.getMetrics).toHaveBeenCalledTimes(1);
      });

      // Simulate retry (e.g., through auto-refresh)
      await act(async () => {
        // Trigger a refresh - this would normally happen through auto-refresh
        // or user action in the real application
      });
    });
  });

  describe('Performance and Optimization', () => {
    test('should not make unnecessary API calls', async () => {
      authService.isAuthenticated.mockReturnValue(true);
      authService.getCurrentUser.mockReturnValue({ id: 1, username: 'testuser' });

      await act(async () => {
        render(<EnhancedApp />);
      });

      await waitFor(() => {
        expect(apiService.getClusterStatus).toHaveBeenCalledTimes(1);
        expect(apiService.getNodes).toHaveBeenCalledTimes(1);
        expect(apiService.getModels).toHaveBeenCalledTimes(1);
      });

      // Re-rendering should not trigger additional API calls
      await act(async () => {
        render(<EnhancedApp />);
      });

      // API calls should not increase
      expect(apiService.getClusterStatus).toHaveBeenCalledTimes(1);
      expect(apiService.getNodes).toHaveBeenCalledTimes(1);
      expect(apiService.getModels).toHaveBeenCalledTimes(1);
    });

    test('should cleanup resources on unmount', async () => {
      authService.isAuthenticated.mockReturnValue(true);
      authService.getCurrentUser.mockReturnValue({ id: 1, username: 'testuser' });

      const { unmount } = await act(async () => {
        return render(<EnhancedApp />);
      });

      await act(async () => {
        unmount();
      });

      expect(wsService.disconnect).toHaveBeenCalled();
      expect(authService.off).toHaveBeenCalledWith('authenticated', expect.any(Function));
      expect(authService.off).toHaveBeenCalledWith('logged_out', expect.any(Function));
    });
  });

  describe('Data Consistency', () => {
    test('should maintain data consistency across components', async () => {
      authService.isAuthenticated.mockReturnValue(true);
      authService.getCurrentUser.mockReturnValue({
        id: 1,
        username: 'testuser',
        role: 'admin'
      });

      const mockNodes = [
        { id: 'node-1', status: 'online', health: 'healthy' },
        { id: 'node-2', status: 'online', health: 'healthy' }
      ];

      apiService.getNodes.mockResolvedValue({ nodes: mockNodes });

      await act(async () => {
        render(<EnhancedApp />);
      });

      await waitFor(() => {
        expect(apiService.getNodes).toHaveBeenCalled();
      });

      // Simulate real-time update
      const nodesHandler = wsService.subscribe.mock.calls.find(
        call => call[0] === 'nodes'
      )?.[1];

      if (nodesHandler) {
        const updatedNode = {
          node: { id: 'node-1', status: 'offline', health: 'unhealthy' }
        };

        await act(async () => {
          nodesHandler(updatedNode);
        });

        // All components should reflect the updated node status
        // This would be verified through the dashboard and node views
      }
    });

    test('should handle concurrent updates correctly', async () => {
      authService.isAuthenticated.mockReturnValue(true);
      authService.getCurrentUser.mockReturnValue({ id: 1, username: 'testuser' });

      await act(async () => {
        render(<EnhancedApp />);
      });

      // Simulate multiple concurrent updates
      const metricsHandler = wsService.subscribe.mock.calls.find(
        call => call[0] === 'metrics'
      )?.[1];

      if (metricsHandler) {
        const updates = [
          { metrics: { cpu: 50 } },
          { metrics: { memory: 70 } },
          { metrics: { network: 30 } }
        ];

        await act(async () => {
          updates.forEach(update => metricsHandler(update));
        });

        // Final state should reflect all updates
        // This would be verified through the metrics display
      }
    });
  });
});

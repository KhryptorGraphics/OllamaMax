// Authentication Service Tests
// Comprehensive tests for authentication service functionality

import authService from '../../services/auth.js';
import apiService from '../../services/api.js';

// Mock API service
jest.mock('../../services/api.js');

describe('Authentication Service', () => {
  beforeEach(() => {
    localStorage.clear();
    sessionStorage.clear();
    jest.clearAllMocks();
    
    // Reset auth service state
    authService.currentUser = null;
    authService.permissions = [];
    authService.clearSessionTimeout();
  });

  describe('Initialization', () => {
    test('should initialize from localStorage', () => {
      const mockUser = { id: 1, username: 'testuser', role: 'user' };
      const mockPermissions = ['read', 'write'];
      
      localStorage.setItem('current_user', JSON.stringify(mockUser));
      localStorage.setItem('user_permissions', JSON.stringify(mockPermissions));
      localStorage.setItem('auth_token', 'test-token');
      
      authService.initializeFromStorage();
      
      expect(authService.getCurrentUser()).toEqual(mockUser);
      expect(authService.getPermissions()).toEqual(mockPermissions);
      expect(authService.isAuthenticated()).toBe(true);
    });

    test('should handle corrupted localStorage data', () => {
      localStorage.setItem('current_user', 'invalid-json');
      localStorage.setItem('auth_token', 'test-token');
      
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      
      authService.initializeFromStorage();
      
      expect(authService.isAuthenticated()).toBe(false);
      expect(consoleSpy).toHaveBeenCalledWith(
        'Failed to restore authentication state:',
        expect.any(Error)
      );
      
      consoleSpy.mockRestore();
    });
  });

  describe('Login', () => {
    test('should login successfully', async () => {
      const mockResponse = {
        user: { id: 1, username: 'testuser', role: 'user' },
        access_token: 'new-token',
        refresh_token: 'new-refresh',
        permissions: ['read', 'write']
      };
      
      apiService.login.mockResolvedValue(mockResponse);
      
      const result = await authService.login('testuser', 'password', true);
      
      expect(apiService.login).toHaveBeenCalledWith('testuser', 'password');
      expect(result.success).toBe(true);
      expect(result.user).toEqual(mockResponse.user);
      expect(authService.isAuthenticated()).toBe(true);
      expect(localStorage.getItem('remember_user')).toBe('testuser');
    });

    test('should handle login failure', async () => {
      apiService.login.mockRejectedValue(new Error('Invalid credentials'));
      
      await expect(authService.login('testuser', 'wrongpassword'))
        .rejects.toThrow('Invalid credentials');
      
      expect(authService.isAuthenticated()).toBe(false);
    });

    test('should emit authentication events', async () => {
      const mockUser = { id: 1, username: 'testuser' };
      apiService.login.mockResolvedValue({
        user: mockUser,
        access_token: 'token',
        permissions: []
      });
      
      const authCallback = jest.fn();
      authService.on('authenticated', authCallback);
      
      await authService.login('testuser', 'password');
      
      expect(authCallback).toHaveBeenCalledWith(mockUser);
    });
  });

  describe('Registration', () => {
    test('should register successfully', async () => {
      const userData = {
        username: 'newuser',
        email: 'new@example.com',
        password: 'password123'
      };
      
      const mockResponse = { success: true, user_id: 123 };
      apiService.register.mockResolvedValue(mockResponse);
      
      const result = await authService.register(userData);
      
      expect(apiService.register).toHaveBeenCalledWith(userData);
      expect(result).toEqual(mockResponse);
    });

    test('should handle registration failure', async () => {
      const userData = { username: 'newuser' };
      apiService.register.mockRejectedValue(new Error('Username taken'));
      
      await expect(authService.register(userData))
        .rejects.toThrow('Username taken');
    });

    test('should emit registration events', async () => {
      const mockResponse = { success: true };
      apiService.register.mockResolvedValue(mockResponse);
      
      const successCallback = jest.fn();
      authService.on('registration_success', successCallback);
      
      await authService.register({ username: 'test' });
      
      expect(successCallback).toHaveBeenCalledWith(mockResponse);
    });
  });

  describe('Logout', () => {
    beforeEach(async () => {
      // Set up authenticated state
      const mockUser = { id: 1, username: 'testuser' };
      authService.currentUser = mockUser;
      authService.permissions = ['read'];
      localStorage.setItem('auth_token', 'test-token');
      localStorage.setItem('current_user', JSON.stringify(mockUser));
    });

    test('should logout successfully', async () => {
      apiService.logout.mockResolvedValue();
      
      const logoutCallback = jest.fn();
      authService.on('logged_out', logoutCallback);
      
      await authService.logout();
      
      expect(apiService.logout).toHaveBeenCalled();
      expect(authService.isAuthenticated()).toBe(false);
      expect(localStorage.getItem('auth_token')).toBeNull();
      expect(logoutCallback).toHaveBeenCalled();
    });

    test('should clear session even if API call fails', async () => {
      apiService.logout.mockRejectedValue(new Error('Network error'));
      
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      
      await authService.logout();
      
      expect(authService.isAuthenticated()).toBe(false);
      expect(consoleSpy).toHaveBeenCalledWith(
        'Logout request failed:',
        expect.any(Error)
      );
      
      consoleSpy.mockRestore();
    });
  });

  describe('Session Management', () => {
    test('should setup session timeout', () => {
      authService.setupSessionTimeout();
      
      expect(authService.sessionTimeout).toBeTruthy();
      expect(authService.sessionWarningTimeout).toBeTruthy();
    });

    test('should extend session', () => {
      authService.currentUser = { id: 1 };
      authService.setupSessionTimeout();
      
      const originalTimeout = authService.sessionTimeout;
      
      authService.extendSession();
      
      expect(authService.sessionTimeout).not.toBe(originalTimeout);
    });

    test('should emit session warning', async () => {
      const warningCallback = jest.fn();
      authService.on('session_warning', warningCallback);
      
      authService.setupSessionTimeout();
      
      // Fast-forward time to trigger warning
      jest.advanceTimersByTime(8 * 60 * 60 * 1000 - 15 * 60 * 1000);
      
      expect(warningCallback).toHaveBeenCalled();
    });

    test('should auto-logout on session expiry', async () => {
      const expiredCallback = jest.fn();
      authService.on('session_expired', expiredCallback);
      
      authService.currentUser = { id: 1 };
      authService.setupSessionTimeout();
      
      // Fast-forward time to trigger expiry
      jest.advanceTimersByTime(8 * 60 * 60 * 1000);
      
      expect(expiredCallback).toHaveBeenCalled();
    });

    test('should track user activity', () => {
      authService.currentUser = { id: 1 };
      authService.setupSessionTimeout();
      
      const originalTimeout = authService.sessionTimeout;
      
      authService.trackActivity();
      
      expect(authService.sessionTimeout).not.toBe(originalTimeout);
    });
  });

  describe('Permissions and Roles', () => {
    beforeEach(() => {
      authService.currentUser = { id: 1, username: 'testuser', role: 'operator' };
      authService.permissions = ['read', 'write', 'operator'];
    });

    test('should check specific permissions', () => {
      expect(authService.hasPermission('read')).toBe(true);
      expect(authService.hasPermission('delete')).toBe(false);
      expect(authService.hasPermission('admin')).toBe(false);
    });

    test('should check multiple permissions', () => {
      expect(authService.hasAnyPermission(['read', 'admin'])).toBe(true);
      expect(authService.hasAnyPermission(['admin', 'delete'])).toBe(false);
      
      expect(authService.hasAllPermissions(['read', 'write'])).toBe(true);
      expect(authService.hasAllPermissions(['read', 'admin'])).toBe(false);
    });

    test('should check roles', () => {
      expect(authService.hasRole('operator')).toBe(true);
      expect(authService.hasRole('admin')).toBe(false);
      expect(authService.hasRole('user')).toBe(false);
    });

    test('should identify admin users', () => {
      expect(authService.isAdmin()).toBe(false);
      
      authService.currentUser.role = 'admin';
      expect(authService.isAdmin()).toBe(true);
      
      authService.currentUser.role = 'user';
      authService.permissions = ['admin'];
      expect(authService.isAdmin()).toBe(true);
    });

    test('should identify operator users', () => {
      expect(authService.isOperator()).toBe(true);
      
      authService.currentUser.role = 'user';
      authService.permissions = ['read'];
      expect(authService.isOperator()).toBe(false);
      
      authService.permissions = ['operator'];
      expect(authService.isOperator()).toBe(true);
    });

    test('should admin have all permissions', () => {
      authService.permissions = ['admin'];
      
      expect(authService.hasPermission('read')).toBe(true);
      expect(authService.hasPermission('write')).toBe(true);
      expect(authService.hasPermission('delete')).toBe(true);
      expect(authService.hasPermission('any_permission')).toBe(true);
    });
  });

  describe('Profile Management', () => {
    beforeEach(() => {
      authService.currentUser = { id: 1, username: 'testuser', email: 'test@example.com' };
    });

    test('should update profile', async () => {
      const profileData = { email: 'newemail@example.com', full_name: 'Test User' };
      const mockResponse = { user: { ...authService.currentUser, ...profileData } };
      
      apiService.request.mockResolvedValue(mockResponse);
      
      const updateCallback = jest.fn();
      authService.on('profile_updated', updateCallback);
      
      await authService.updateProfile(profileData);
      
      expect(apiService.request).toHaveBeenCalledWith('/auth/profile', {
        method: 'PUT',
        body: JSON.stringify(profileData)
      });
      
      expect(authService.currentUser.email).toBe('newemail@example.com');
      expect(updateCallback).toHaveBeenCalled();
    });

    test('should change password', async () => {
      apiService.request.mockResolvedValue({ success: true });
      
      const changeCallback = jest.fn();
      authService.on('password_changed', changeCallback);
      
      await authService.changePassword('oldpass', 'newpass');
      
      expect(apiService.request).toHaveBeenCalledWith('/auth/change-password', {
        method: 'POST',
        body: JSON.stringify({
          current_password: 'oldpass',
          new_password: 'newpass'
        })
      });
      
      expect(changeCallback).toHaveBeenCalled();
    });

    test('should request password reset', async () => {
      apiService.request.mockResolvedValue({ success: true });
      
      const resetCallback = jest.fn();
      authService.on('reset_requested', resetCallback);
      
      await authService.requestPasswordReset('test@example.com');
      
      expect(apiService.request).toHaveBeenCalledWith('/auth/reset-password', {
        method: 'POST',
        body: JSON.stringify({ email: 'test@example.com' })
      });
      
      expect(resetCallback).toHaveBeenCalled();
    });
  });

  describe('Two-Factor Authentication', () => {
    beforeEach(() => {
      authService.currentUser = { id: 1, username: 'testuser' };
    });

    test('should setup 2FA', async () => {
      const mockResponse = { qr_code: 'data:image/png;base64,...', secret: 'SECRET123' };
      apiService.request.mockResolvedValue(mockResponse);
      
      const setupCallback = jest.fn();
      authService.on('2fa_setup', setupCallback);
      
      const result = await authService.setupTwoFactor();
      
      expect(apiService.request).toHaveBeenCalledWith('/auth/2fa/setup', {
        method: 'POST'
      });
      
      expect(result).toEqual(mockResponse);
      expect(setupCallback).toHaveBeenCalledWith(mockResponse);
    });

    test('should verify 2FA', async () => {
      apiService.request.mockResolvedValue({ success: true });
      
      const verifyCallback = jest.fn();
      authService.on('2fa_verified', verifyCallback);
      
      await authService.verifyTwoFactor('123456');
      
      expect(apiService.request).toHaveBeenCalledWith('/auth/2fa/verify', {
        method: 'POST',
        body: JSON.stringify({ code: '123456' })
      });
      
      expect(authService.currentUser.two_factor_enabled).toBe(true);
      expect(verifyCallback).toHaveBeenCalled();
    });
  });

  describe('Remember Me Functionality', () => {
    test('should remember username', () => {
      authService.login('testuser', 'password', true);
      
      expect(authService.getRememberedUsername()).toBe('testuser');
    });

    test('should clear remembered username', () => {
      localStorage.setItem('remember_user', 'testuser');
      
      authService.clearRememberedUsername();
      
      expect(authService.getRememberedUsername()).toBeNull();
    });

    test('should not remember username when not requested', async () => {
      const mockResponse = {
        user: { id: 1, username: 'testuser' },
        access_token: 'token',
        permissions: []
      };
      
      apiService.login.mockResolvedValue(mockResponse);
      
      await authService.login('testuser', 'password', false);
      
      expect(localStorage.getItem('remember_user')).toBeNull();
    });
  });

  describe('Event System', () => {
    test('should add and remove event listeners', () => {
      const callback = jest.fn();
      
      authService.on('test_event', callback);
      authService.emit('test_event', 'test data');
      
      expect(callback).toHaveBeenCalledWith('test data');
      
      authService.off('test_event', callback);
      authService.emit('test_event', 'test data 2');
      
      expect(callback).toHaveBeenCalledTimes(1);
    });

    test('should handle errors in event listeners', () => {
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      const faultyCallback = () => { throw new Error('Callback error'); };
      
      authService.on('test_event', faultyCallback);
      authService.emit('test_event', 'test data');
      
      expect(consoleSpy).toHaveBeenCalledWith(
        "Error in auth event listener for 'test_event':",
        expect.any(Error)
      );
      
      consoleSpy.mockRestore();
    });
  });

  describe('Session Info', () => {
    test('should return session information', () => {
      authService.currentUser = { id: 1, username: 'testuser', role: 'user' };
      authService.permissions = ['read', 'write'];
      authService.setupSessionTimeout();
      
      const sessionInfo = authService.getSessionInfo();
      
      expect(sessionInfo).toEqual({
        isAuthenticated: true,
        user: authService.currentUser,
        permissions: authService.permissions,
        role: 'user',
        sessionActive: true
      });
    });

    test('should return empty session info when not authenticated', () => {
      const sessionInfo = authService.getSessionInfo();
      
      expect(sessionInfo).toEqual({
        isAuthenticated: false,
        user: null,
        permissions: [],
        role: null,
        sessionActive: false
      });
    });
  });
});

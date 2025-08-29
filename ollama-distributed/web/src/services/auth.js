// Authentication Service
// Handles user authentication, session management, and role-based access control

import apiService from './api.js';

class AuthService {
  constructor() {
    this.currentUser = null;
    this.permissions = [];
    this.sessionTimeout = null;
    this.sessionWarningTimeout = null;
    this.eventListeners = new Map();
    
    // Initialize from stored session
    this.initializeFromStorage();
  }

  // Initialize authentication state from localStorage
  initializeFromStorage() {
    const token = localStorage.getItem('auth_token');
    const userStr = localStorage.getItem('current_user');
    const permissionsStr = localStorage.getItem('user_permissions');

    if (token && userStr) {
      try {
        this.currentUser = JSON.parse(userStr);
        this.permissions = permissionsStr ? JSON.parse(permissionsStr) : [];
        this.setupSessionTimeout();
        this.emit('authenticated', this.currentUser);
      } catch (error) {
        console.error('Failed to restore authentication state:', error);
        this.clearSession();
      }
    }
  }

  // Login user
  async login(username, password, rememberMe = false) {
    try {
      const response = await apiService.login(username, password);
      
      if (response.user && response.access_token) {
        this.currentUser = response.user;
        this.permissions = response.permissions || [];
        
        // Store session data
        localStorage.setItem('current_user', JSON.stringify(this.currentUser));
        localStorage.setItem('user_permissions', JSON.stringify(this.permissions));
        
        if (rememberMe) {
          localStorage.setItem('remember_user', username);
        }
        
        this.setupSessionTimeout();
        this.emit('authenticated', this.currentUser);
        
        return {
          success: true,
          user: this.currentUser,
          permissions: this.permissions
        };
      } else {
        throw new Error('Invalid response from server');
      }
    } catch (error) {
      this.emit('auth_error', error.message);
      throw error;
    }
  }

  // Register new user
  async register(userData) {
    try {
      const response = await apiService.register(userData);
      
      if (response.success) {
        this.emit('registration_success', response);
        return response;
      } else {
        throw new Error(response.message || 'Registration failed');
      }
    } catch (error) {
      this.emit('registration_error', error.message);
      throw error;
    }
  }

  // Logout user
  async logout() {
    try {
      await apiService.logout();
    } catch (error) {
      console.error('Logout request failed:', error);
    } finally {
      this.clearSession();
      this.emit('logged_out');
    }
  }

  // Clear session data
  clearSession() {
    this.currentUser = null;
    this.permissions = [];
    
    localStorage.removeItem('auth_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('current_user');
    localStorage.removeItem('user_permissions');
    
    this.clearSessionTimeout();
    apiService.clearAuth();
  }

  // Setup session timeout
  setupSessionTimeout() {
    this.clearSessionTimeout();
    
    // Session expires after 8 hours of inactivity
    const sessionDuration = 8 * 60 * 60 * 1000; // 8 hours
    const warningTime = 15 * 60 * 1000; // 15 minutes before expiry
    
    // Show warning 15 minutes before session expires
    this.sessionWarningTimeout = setTimeout(() => {
      this.emit('session_warning', {
        timeRemaining: warningTime,
        message: 'Your session will expire in 15 minutes. Please save your work.'
      });
    }, sessionDuration - warningTime);
    
    // Auto-logout when session expires
    this.sessionTimeout = setTimeout(() => {
      this.emit('session_expired');
      this.logout();
    }, sessionDuration);
  }

  // Clear session timeout
  clearSessionTimeout() {
    if (this.sessionTimeout) {
      clearTimeout(this.sessionTimeout);
      this.sessionTimeout = null;
    }
    
    if (this.sessionWarningTimeout) {
      clearTimeout(this.sessionWarningTimeout);
      this.sessionWarningTimeout = null;
    }
  }

  // Extend session (reset timeout)
  extendSession() {
    if (this.isAuthenticated()) {
      this.setupSessionTimeout();
      this.emit('session_extended');
    }
  }

  // Check if user is authenticated
  isAuthenticated() {
    return !!this.currentUser && !!apiService.getAuthToken();
  }

  // Get current user
  getCurrentUser() {
    return this.currentUser;
  }

  // Get user permissions
  getPermissions() {
    return this.permissions;
  }

  // Check if user has specific permission
  hasPermission(permission) {
    return this.permissions.includes(permission) || this.permissions.includes('admin');
  }

  // Check if user has any of the specified permissions
  hasAnyPermission(permissions) {
    return permissions.some(permission => this.hasPermission(permission));
  }

  // Check if user has all specified permissions
  hasAllPermissions(permissions) {
    return permissions.every(permission => this.hasPermission(permission));
  }

  // Check if user has specific role
  hasRole(role) {
    return this.currentUser && this.currentUser.role === role;
  }

  // Check if user is admin
  isAdmin() {
    return this.hasRole('admin') || this.hasPermission('admin');
  }

  // Check if user is operator
  isOperator() {
    return this.hasRole('operator') || this.hasPermission('operator') || this.isAdmin();
  }

  // Get user's role
  getRole() {
    return this.currentUser ? this.currentUser.role : null;
  }

  // Update user profile
  async updateProfile(profileData) {
    try {
      const response = await apiService.request('/auth/profile', {
        method: 'PUT',
        body: JSON.stringify(profileData)
      });
      
      if (response.user) {
        this.currentUser = { ...this.currentUser, ...response.user };
        localStorage.setItem('current_user', JSON.stringify(this.currentUser));
        this.emit('profile_updated', this.currentUser);
      }
      
      return response;
    } catch (error) {
      this.emit('profile_error', error.message);
      throw error;
    }
  }

  // Change password
  async changePassword(currentPassword, newPassword) {
    try {
      const response = await apiService.request('/auth/change-password', {
        method: 'POST',
        body: JSON.stringify({
          current_password: currentPassword,
          new_password: newPassword
        })
      });
      
      this.emit('password_changed');
      return response;
    } catch (error) {
      this.emit('password_error', error.message);
      throw error;
    }
  }

  // Request password reset
  async requestPasswordReset(email) {
    try {
      const response = await apiService.request('/auth/reset-password', {
        method: 'POST',
        body: JSON.stringify({ email })
      });
      
      this.emit('reset_requested');
      return response;
    } catch (error) {
      this.emit('reset_error', error.message);
      throw error;
    }
  }

  // Verify email
  async verifyEmail(token) {
    try {
      const response = await apiService.request('/auth/verify-email', {
        method: 'POST',
        body: JSON.stringify({ token })
      });
      
      if (response.success && this.currentUser) {
        this.currentUser.is_verified = true;
        localStorage.setItem('current_user', JSON.stringify(this.currentUser));
        this.emit('email_verified');
      }
      
      return response;
    } catch (error) {
      this.emit('verification_error', error.message);
      throw error;
    }
  }

  // Setup two-factor authentication
  async setupTwoFactor() {
    try {
      const response = await apiService.request('/auth/2fa/setup', {
        method: 'POST'
      });
      
      this.emit('2fa_setup', response);
      return response;
    } catch (error) {
      this.emit('2fa_error', error.message);
      throw error;
    }
  }

  // Verify two-factor authentication
  async verifyTwoFactor(code) {
    try {
      const response = await apiService.request('/auth/2fa/verify', {
        method: 'POST',
        body: JSON.stringify({ code })
      });
      
      if (response.success && this.currentUser) {
        this.currentUser.two_factor_enabled = true;
        localStorage.setItem('current_user', JSON.stringify(this.currentUser));
        this.emit('2fa_verified');
      }
      
      return response;
    } catch (error) {
      this.emit('2fa_error', error.message);
      throw error;
    }
  }

  // Get remembered username
  getRememberedUsername() {
    return localStorage.getItem('remember_user');
  }

  // Clear remembered username
  clearRememberedUsername() {
    localStorage.removeItem('remember_user');
  }

  // Event listener management
  on(event, callback) {
    if (!this.eventListeners.has(event)) {
      this.eventListeners.set(event, []);
    }
    this.eventListeners.get(event).push(callback);
  }

  off(event, callback) {
    if (this.eventListeners.has(event)) {
      const listeners = this.eventListeners.get(event);
      const index = listeners.indexOf(callback);
      if (index > -1) {
        listeners.splice(index, 1);
      }
    }
  }

  emit(event, data = null) {
    if (this.eventListeners.has(event)) {
      this.eventListeners.get(event).forEach(callback => {
        try {
          callback(data);
        } catch (error) {
          console.error(`Error in auth event listener for '${event}':`, error);
        }
      });
    }
  }

  // Activity tracking for session management
  trackActivity() {
    if (this.isAuthenticated()) {
      this.extendSession();
    }
  }

  // Initialize activity tracking
  initActivityTracking() {
    const events = ['mousedown', 'mousemove', 'keypress', 'scroll', 'touchstart'];
    
    events.forEach(event => {
      document.addEventListener(event, () => {
        this.trackActivity();
      }, { passive: true });
    });
  }

  // Get session info
  getSessionInfo() {
    return {
      isAuthenticated: this.isAuthenticated(),
      user: this.currentUser,
      permissions: this.permissions,
      role: this.getRole(),
      sessionActive: !!this.sessionTimeout
    };
  }
}

// Create and export singleton instance
const authService = new AuthService();

// Initialize activity tracking
authService.initActivityTracking();

export default authService;

// Make it available globally for debugging
if (typeof window !== 'undefined') {
  window.authService = authService;
}

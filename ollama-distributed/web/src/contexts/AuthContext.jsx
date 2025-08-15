/**
 * Authentication Context
 * 
 * Provides authentication state and methods throughout the application.
 */

import React, { createContext, useContext, useState, useEffect } from 'react';
import authService from '../services/authService.js';

const AuthContext = createContext();

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  // Initialize authentication state
  useEffect(() => {
    const initializeAuth = async () => {
      try {
        if (authService.isAuthenticated()) {
          const userData = authService.getUser();
          setUser(userData);
          setIsAuthenticated(true);
          
          // Setup automatic token refresh
          authService.setupTokenRefresh();
          
          // Verify token is still valid by making a test request
          try {
            const profile = await authService.getProfile();
            setUser(profile);
          } catch (error) {
            // Token might be invalid, clear auth data
            await authService.logout();
            setUser(null);
            setIsAuthenticated(false);
          }
        }
      } catch (error) {
        console.error('Auth initialization error:', error);
        setUser(null);
        setIsAuthenticated(false);
      } finally {
        setLoading(false);
      }
    };

    initializeAuth();
  }, []);

  // Login function
  const login = async (credentials) => {
    try {
      setLoading(true);
      const response = await authService.login(credentials);
      
      setUser(response.user);
      setIsAuthenticated(true);
      
      // Setup automatic token refresh
      authService.setupTokenRefresh();
      
      return response;
    } catch (error) {
      console.error('Login failed:', error);
      throw error;
    } finally {
      setLoading(false);
    }
  };

  // Register function
  const register = async (userData) => {
    try {
      setLoading(true);
      const response = await authService.register(userData);
      
      // If auto-login is enabled after registration
      if (response.user && response.token) {
        setUser(response.user);
        setIsAuthenticated(true);
        authService.setupTokenRefresh();
      }
      
      return response;
    } catch (error) {
      console.error('Registration failed:', error);
      throw error;
    } finally {
      setLoading(false);
    }
  };

  // Logout function
  const logout = async () => {
    try {
      setLoading(true);
      await authService.logout();
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      setUser(null);
      setIsAuthenticated(false);
      setLoading(false);
    }
  };

  // Update user profile
  const updateProfile = async (profileData) => {
    try {
      const updatedUser = await authService.updateProfile(profileData);
      setUser(updatedUser);
      return updatedUser;
    } catch (error) {
      console.error('Profile update failed:', error);
      throw error;
    }
  };

  // Change password
  const changePassword = async (passwordData) => {
    try {
      return await authService.changePassword(passwordData);
    } catch (error) {
      console.error('Password change failed:', error);
      throw error;
    }
  };

  // Request password reset
  const requestPasswordReset = async (email) => {
    try {
      return await authService.requestPasswordReset(email);
    } catch (error) {
      console.error('Password reset request failed:', error);
      throw error;
    }
  };

  // Reset password
  const resetPassword = async (token, newPassword) => {
    try {
      return await authService.resetPassword(token, newPassword);
    } catch (error) {
      console.error('Password reset failed:', error);
      throw error;
    }
  };

  // Verify email
  const verifyEmail = async (token) => {
    try {
      const response = await authService.verifyEmail(token);
      
      // Update user data if verification affects user state
      if (response.user) {
        setUser(response.user);
      }
      
      return response;
    } catch (error) {
      console.error('Email verification failed:', error);
      throw error;
    }
  };

  // Check permission
  const hasPermission = async (permission) => {
    try {
      return await authService.hasPermission(permission);
    } catch (error) {
      console.error('Permission check failed:', error);
      return false;
    }
  };

  // Refresh user data
  const refreshUser = async () => {
    try {
      if (isAuthenticated) {
        const userData = await authService.getProfile();
        setUser(userData);
        return userData;
      }
    } catch (error) {
      console.error('User refresh failed:', error);
      // If refresh fails, user might be logged out
      await logout();
      throw error;
    }
  };

  const value = {
    // State
    user,
    loading,
    isAuthenticated,
    
    // Actions
    login,
    register,
    logout,
    updateProfile,
    changePassword,
    requestPasswordReset,
    resetPassword,
    verifyEmail,
    hasPermission,
    refreshUser,
    
    // Utilities
    getToken: () => authService.getToken(),
    makeAuthenticatedRequest: (endpoint, options) => authService.makeRequest(endpoint, options)
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};

// Higher-order component for protected routes
export const withAuth = (Component) => {
  return function AuthenticatedComponent(props) {
    const { isAuthenticated, loading } = useAuth();
    
    if (loading) {
      return (
        <div style={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          height: '100vh',
          fontSize: '1.125rem'
        }}>
          Loading...
        </div>
      );
    }
    
    if (!isAuthenticated) {
      return <AuthRequired />;
    }
    
    return <Component {...props} />;
  };
};

// Component to show when authentication is required
const AuthRequired = () => {
  return (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      justifyContent: 'center',
      alignItems: 'center',
      height: '100vh',
      textAlign: 'center',
      padding: '2rem'
    }}>
      <h1 style={{ fontSize: '2rem', marginBottom: '1rem' }}>
        Authentication Required
      </h1>
      <p style={{ fontSize: '1.125rem', color: '#666', marginBottom: '2rem' }}>
        Please log in to access this page.
      </p>
      <button
        onClick={() => window.location.reload()}
        style={{
          padding: '0.75rem 1.5rem',
          fontSize: '1rem',
          backgroundColor: '#0ea5e9',
          color: 'white',
          border: 'none',
          borderRadius: '0.375rem',
          cursor: 'pointer'
        }}
      >
        Go to Login
      </button>
    </div>
  );
};

export default AuthContext;

/**
 * Login Form Component
 * 
 * Provides secure user authentication with the OllamaMax design system.
 */

import React, { useState } from 'react';
import { Button, Input, Card } from '../../design-system/index.js';
import { useTheme } from '../../design-system/theme/ThemeProvider.jsx';

const LoginForm = ({ onLogin, onSwitchToRegister, loading = false, error = null }) => {
  const { theme, utils } = useTheme();
  const [formData, setFormData] = useState({
    email: '',
    password: '',
    rememberMe: false
  });
  const [formErrors, setFormErrors] = useState({});
  const [showPassword, setShowPassword] = useState(false);

  // Handle input changes
  const handleChange = (field) => (event) => {
    const value = event.target.type === 'checkbox' ? event.target.checked : event.target.value;
    setFormData(prev => ({
      ...prev,
      [field]: value
    }));
    
    // Clear field error when user starts typing
    if (formErrors[field]) {
      setFormErrors(prev => ({
        ...prev,
        [field]: null
      }));
    }
  };

  // Validate form
  const validateForm = () => {
    const errors = {};
    
    if (!formData.email) {
      errors.email = 'Email is required';
    } else if (!/\S+@\S+\.\S+/.test(formData.email)) {
      errors.email = 'Please enter a valid email address';
    }
    
    if (!formData.password) {
      errors.password = 'Password is required';
    } else if (formData.password.length < 6) {
      errors.password = 'Password must be at least 6 characters';
    }
    
    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  // Handle form submission
  const handleSubmit = async (event) => {
    event.preventDefault();
    
    if (!validateForm()) {
      return;
    }
    
    try {
      await onLogin(formData);
    } catch (err) {
      console.error('Login failed:', err);
    }
  };

  // Container styles
  const containerStyles = {
    minHeight: '100vh',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    background: `linear-gradient(135deg, ${theme.colors.primary} 0%, ${theme.colors.secondary} 100%)`,
    padding: '1rem'
  };

  // Form styles
  const formStyles = {
    width: '100%',
    maxWidth: '400px'
  };

  // Header styles
  const headerStyles = {
    textAlign: 'center',
    marginBottom: '2rem'
  };

  const logoStyles = {
    fontSize: '2.5rem',
    fontWeight: 'bold',
    color: theme.colors.text,
    marginBottom: '0.5rem'
  };

  const subtitleStyles = {
    color: theme.colors.textSecondary,
    fontSize: '1rem'
  };

  // Form field styles
  const fieldGroupStyles = {
    marginBottom: '1.5rem'
  };

  // Checkbox styles
  const checkboxGroupStyles = {
    display: 'flex',
    alignItems: 'center',
    gap: '0.5rem',
    marginBottom: '1.5rem'
  };

  // Link styles
  const linkStyles = {
    color: theme.colors.primary,
    textDecoration: 'none',
    fontSize: '0.875rem',
    fontWeight: '500',
    cursor: 'pointer',
    ':hover': {
      textDecoration: 'underline'
    }
  };

  // Footer styles
  const footerStyles = {
    textAlign: 'center',
    marginTop: '1.5rem',
    fontSize: '0.875rem',
    color: theme.colors.textSecondary
  };

  return (
    <div style={containerStyles}>
      <Card variant="elevated" size="lg" style={formStyles}>
        <div style={headerStyles}>
          <div style={logoStyles}>OllamaMax</div>
          <div style={subtitleStyles}>Sign in to your account</div>
        </div>

        {error && (
          <div style={{
            padding: '0.75rem',
            backgroundColor: theme.colors.error + '10',
            border: `1px solid ${theme.colors.error}`,
            borderRadius: '0.375rem',
            color: theme.colors.error,
            fontSize: '0.875rem',
            marginBottom: '1.5rem'
          }}>
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit}>
          <div style={fieldGroupStyles}>
            <Input
              type="email"
              label="Email Address"
              placeholder="Enter your email"
              value={formData.email}
              onChange={handleChange('email')}
              error={!!formErrors.email}
              errorText={formErrors.email}
              required
              autoComplete="email"
              leftIcon={
                <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M20 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2zm0 4l-8 5-8-5V6l8 5 8-5v2z"/>
                </svg>
              }
            />
          </div>

          <div style={fieldGroupStyles}>
            <Input
              type="password"
              label="Password"
              placeholder="Enter your password"
              value={formData.password}
              onChange={handleChange('password')}
              error={!!formErrors.password}
              errorText={formErrors.password}
              required
              autoComplete="current-password"
              leftIcon={
                <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"/>
                </svg>
              }
            />
          </div>

          <div style={checkboxGroupStyles}>
            <input
              type="checkbox"
              id="rememberMe"
              checked={formData.rememberMe}
              onChange={handleChange('rememberMe')}
              style={{
                width: '1rem',
                height: '1rem',
                accentColor: theme.colors.primary
              }}
            />
            <label 
              htmlFor="rememberMe"
              style={{
                fontSize: '0.875rem',
                color: theme.colors.textSecondary,
                cursor: 'pointer'
              }}
            >
              Remember me
            </label>
          </div>

          <Button
            type="submit"
            variant="primary"
            size="lg"
            fullWidth
            loading={loading}
            disabled={loading}
          >
            {loading ? 'Signing in...' : 'Sign In'}
          </Button>
        </form>

        <div style={footerStyles}>
          <div style={{ marginBottom: '0.5rem' }}>
            <span 
              style={linkStyles}
              onClick={() => {/* Handle forgot password */}}
            >
              Forgot your password?
            </span>
          </div>
          <div>
            Don't have an account?{' '}
            <span 
              style={linkStyles}
              onClick={onSwitchToRegister}
            >
              Sign up
            </span>
          </div>
        </div>
      </Card>
    </div>
  );
};

export default LoginForm;

/**
 * Registration Form Component
 * 
 * Provides user registration with validation and security features.
 */

import React, { useState } from 'react';
import { Button, Input, Card } from '../../design-system/index.js';
import { useTheme } from '../../design-system/theme/ThemeProvider.jsx';

const RegisterForm = ({ onRegister, onSwitchToLogin, loading = false, error = null }) => {
  const { theme } = useTheme();
  const [formData, setFormData] = useState({
    firstName: '',
    lastName: '',
    email: '',
    password: '',
    confirmPassword: '',
    acceptTerms: false
  });
  const [formErrors, setFormErrors] = useState({});
  const [passwordStrength, setPasswordStrength] = useState(0);

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

    // Calculate password strength
    if (field === 'password') {
      setPasswordStrength(calculatePasswordStrength(value));
    }
  };

  // Calculate password strength
  const calculatePasswordStrength = (password) => {
    let strength = 0;
    if (password.length >= 8) strength += 1;
    if (/[a-z]/.test(password)) strength += 1;
    if (/[A-Z]/.test(password)) strength += 1;
    if (/[0-9]/.test(password)) strength += 1;
    if (/[^A-Za-z0-9]/.test(password)) strength += 1;
    return strength;
  };

  // Get password strength label and color
  const getPasswordStrengthInfo = () => {
    const strengthLabels = ['Very Weak', 'Weak', 'Fair', 'Good', 'Strong'];
    const strengthColors = [
      theme.colors.error,
      '#ff6b35',
      '#ffa500',
      '#32cd32',
      theme.colors.success
    ];
    
    return {
      label: strengthLabels[passwordStrength] || 'Very Weak',
      color: strengthColors[passwordStrength] || theme.colors.error,
      percentage: (passwordStrength / 5) * 100
    };
  };

  // Validate form
  const validateForm = () => {
    const errors = {};
    
    if (!formData.firstName.trim()) {
      errors.firstName = 'First name is required';
    }
    
    if (!formData.lastName.trim()) {
      errors.lastName = 'Last name is required';
    }
    
    if (!formData.email) {
      errors.email = 'Email is required';
    } else if (!/\S+@\S+\.\S+/.test(formData.email)) {
      errors.email = 'Please enter a valid email address';
    }
    
    if (!formData.password) {
      errors.password = 'Password is required';
    } else if (formData.password.length < 8) {
      errors.password = 'Password must be at least 8 characters';
    } else if (passwordStrength < 3) {
      errors.password = 'Password is too weak. Please use a stronger password.';
    }
    
    if (!formData.confirmPassword) {
      errors.confirmPassword = 'Please confirm your password';
    } else if (formData.password !== formData.confirmPassword) {
      errors.confirmPassword = 'Passwords do not match';
    }
    
    if (!formData.acceptTerms) {
      errors.acceptTerms = 'You must accept the terms and conditions';
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
      await onRegister({
        firstName: formData.firstName.trim(),
        lastName: formData.lastName.trim(),
        email: formData.email.toLowerCase().trim(),
        password: formData.password
      });
    } catch (err) {
      console.error('Registration failed:', err);
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
    maxWidth: '450px'
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

  const nameFieldsStyles = {
    display: 'grid',
    gridTemplateColumns: '1fr 1fr',
    gap: '1rem',
    marginBottom: '1.5rem'
  };

  // Password strength indicator
  const strengthInfo = getPasswordStrengthInfo();
  const strengthIndicatorStyles = {
    marginTop: '0.5rem'
  };

  const strengthBarStyles = {
    width: '100%',
    height: '4px',
    backgroundColor: theme.colors.border,
    borderRadius: '2px',
    overflow: 'hidden',
    marginBottom: '0.25rem'
  };

  const strengthFillStyles = {
    height: '100%',
    width: `${strengthInfo.percentage}%`,
    backgroundColor: strengthInfo.color,
    transition: 'all 0.3s ease'
  };

  const strengthLabelStyles = {
    fontSize: '0.75rem',
    color: strengthInfo.color,
    fontWeight: '500'
  };

  // Checkbox styles
  const checkboxGroupStyles = {
    display: 'flex',
    alignItems: 'flex-start',
    gap: '0.5rem',
    marginBottom: '1.5rem'
  };

  // Link styles
  const linkStyles = {
    color: theme.colors.primary,
    textDecoration: 'none',
    fontSize: '0.875rem',
    fontWeight: '500',
    cursor: 'pointer'
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
          <div style={subtitleStyles}>Create your account</div>
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
          <div style={nameFieldsStyles}>
            <Input
              type="text"
              label="First Name"
              placeholder="John"
              value={formData.firstName}
              onChange={handleChange('firstName')}
              error={!!formErrors.firstName}
              errorText={formErrors.firstName}
              required
              autoComplete="given-name"
            />
            <Input
              type="text"
              label="Last Name"
              placeholder="Doe"
              value={formData.lastName}
              onChange={handleChange('lastName')}
              error={!!formErrors.lastName}
              errorText={formErrors.lastName}
              required
              autoComplete="family-name"
            />
          </div>

          <div style={fieldGroupStyles}>
            <Input
              type="email"
              label="Email Address"
              placeholder="john.doe@example.com"
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
              placeholder="Create a strong password"
              value={formData.password}
              onChange={handleChange('password')}
              error={!!formErrors.password}
              errorText={formErrors.password}
              required
              autoComplete="new-password"
              leftIcon={
                <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"/>
                </svg>
              }
            />
            {formData.password && (
              <div style={strengthIndicatorStyles}>
                <div style={strengthBarStyles}>
                  <div style={strengthFillStyles}></div>
                </div>
                <div style={strengthLabelStyles}>
                  Password strength: {strengthInfo.label}
                </div>
              </div>
            )}
          </div>

          <div style={fieldGroupStyles}>
            <Input
              type="password"
              label="Confirm Password"
              placeholder="Confirm your password"
              value={formData.confirmPassword}
              onChange={handleChange('confirmPassword')}
              error={!!formErrors.confirmPassword}
              errorText={formErrors.confirmPassword}
              required
              autoComplete="new-password"
              leftIcon={
                <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
                </svg>
              }
            />
          </div>

          <div style={checkboxGroupStyles}>
            <input
              type="checkbox"
              id="acceptTerms"
              checked={formData.acceptTerms}
              onChange={handleChange('acceptTerms')}
              style={{
                width: '1rem',
                height: '1rem',
                accentColor: theme.colors.primary,
                marginTop: '0.125rem'
              }}
            />
            <label 
              htmlFor="acceptTerms"
              style={{
                fontSize: '0.875rem',
                color: theme.colors.textSecondary,
                cursor: 'pointer',
                lineHeight: '1.4'
              }}
            >
              I agree to the{' '}
              <span style={linkStyles}>Terms of Service</span>
              {' '}and{' '}
              <span style={linkStyles}>Privacy Policy</span>
            </label>
          </div>

          {formErrors.acceptTerms && (
            <div style={{
              color: theme.colors.error,
              fontSize: '0.875rem',
              marginTop: '-1rem',
              marginBottom: '1rem'
            }}>
              {formErrors.acceptTerms}
            </div>
          )}

          <Button
            type="submit"
            variant="primary"
            size="lg"
            fullWidth
            loading={loading}
            disabled={loading}
          >
            {loading ? 'Creating Account...' : 'Create Account'}
          </Button>
        </form>

        <div style={footerStyles}>
          Already have an account?{' '}
          <span 
            style={linkStyles}
            onClick={onSwitchToLogin}
          >
            Sign in
          </span>
        </div>
      </Card>
    </div>
  );
};

export default RegisterForm;

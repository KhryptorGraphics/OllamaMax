/**
 * Authentication Page Component
 * 
 * Main authentication component that handles login and registration flows.
 */

import React, { useState } from 'react';
import { useAuth } from '../../contexts/AuthContext.jsx';
import LoginForm from './LoginForm.jsx';
import RegisterForm from './RegisterForm.jsx';

const AuthPage = () => {
  const [mode, setMode] = useState('login'); // 'login' or 'register'
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  
  const { login, register } = useAuth();

  // Handle login
  const handleLogin = async (credentials) => {
    setLoading(true);
    setError(null);
    
    try {
      await login(credentials);
      // Redirect will be handled by the auth context
    } catch (err) {
      setError(err.message || 'Login failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  // Handle registration
  const handleRegister = async (userData) => {
    setLoading(true);
    setError(null);
    setSuccess(null);
    
    try {
      const response = await register(userData);
      
      // Check if email verification is required
      if (response.requiresVerification) {
        setSuccess(
          'Account created successfully! Please check your email to verify your account before signing in.'
        );
        setMode('login');
      } else {
        // Auto-login successful, redirect will be handled by auth context
      }
    } catch (err) {
      setError(err.message || 'Registration failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  // Switch to login mode
  const switchToLogin = () => {
    setMode('login');
    setError(null);
    setSuccess(null);
  };

  // Switch to register mode
  const switchToRegister = () => {
    setMode('register');
    setError(null);
    setSuccess(null);
  };

  // Render success message
  if (success) {
    return (
      <div style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'linear-gradient(135deg, #0ea5e9 0%, #d946ef 100%)',
        padding: '1rem'
      }}>
        <div style={{
          backgroundColor: 'white',
          padding: '2rem',
          borderRadius: '0.5rem',
          boxShadow: '0 10px 15px -3px rgba(0, 0, 0, 0.1)',
          textAlign: 'center',
          maxWidth: '400px',
          width: '100%'
        }}>
          <div style={{
            width: '64px',
            height: '64px',
            backgroundColor: '#22c55e',
            borderRadius: '50%',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            margin: '0 auto 1rem'
          }}>
            <svg width="32" height="32" viewBox="0 0 24 24" fill="white">
              <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
          </div>
          
          <h2 style={{
            fontSize: '1.5rem',
            fontWeight: 'bold',
            color: '#1f2937',
            marginBottom: '1rem'
          }}>
            Success!
          </h2>
          
          <p style={{
            color: '#6b7280',
            marginBottom: '2rem',
            lineHeight: '1.5'
          }}>
            {success}
          </p>
          
          <button
            onClick={switchToLogin}
            style={{
              backgroundColor: '#0ea5e9',
              color: 'white',
              padding: '0.75rem 1.5rem',
              borderRadius: '0.375rem',
              border: 'none',
              fontSize: '1rem',
              fontWeight: '500',
              cursor: 'pointer',
              transition: 'background-color 0.2s'
            }}
            onMouseOver={(e) => e.target.style.backgroundColor = '#0284c7'}
            onMouseOut={(e) => e.target.style.backgroundColor = '#0ea5e9'}
          >
            Continue to Sign In
          </button>
        </div>
      </div>
    );
  }

  // Render appropriate form based on mode
  if (mode === 'register') {
    return (
      <RegisterForm
        onRegister={handleRegister}
        onSwitchToLogin={switchToLogin}
        loading={loading}
        error={error}
      />
    );
  }

  return (
    <LoginForm
      onLogin={handleLogin}
      onSwitchToRegister={switchToRegister}
      loading={loading}
      error={error}
    />
  );
};

export default AuthPage;

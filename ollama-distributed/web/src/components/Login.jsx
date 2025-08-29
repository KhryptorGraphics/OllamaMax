import React, { useState, useEffect } from 'react';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faUser, faLock, faSignInAlt, faEye, faEyeSlash, faUserPlus, faSpinner, faExclamationTriangle } from '@fortawesome/free-solid-svg-icons';
import authService from '../services/auth.js';
import '../styles/theme.css';

const Login = ({ onLogin, onRegister, onForgotPassword }) => {
  const [credentials, setCredentials] = useState({
    username: '',
    password: ''
  });
  const [showPassword, setShowPassword] = useState(false);
  const [validationErrors, setValidationErrors] = useState({});
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [rememberMe, setRememberMe] = useState(false);

  // Initialize with remembered username
  useEffect(() => {
    const rememberedUsername = authService.getRememberedUsername();
    if (rememberedUsername) {
      setCredentials(prev => ({ ...prev, username: rememberedUsername }));
      setRememberMe(true);
    }
  }, []);

  const validateForm = () => {
    const errors = {};
    if (!credentials.username) {
      errors.username = 'Username is required';
    } else if (credentials.username.length < 3) {
      errors.username = 'Username must be at least 3 characters';
    }
    
    if (!credentials.password) {
      errors.password = 'Password is required';
    } else if (credentials.password.length < 6) {
      errors.password = 'Password must be at least 6 characters';
    }
    
    setValidationErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!validateForm()) return;

    setLoading(true);
    setError(null);

    try {
      const result = await authService.login(
        credentials.username,
        credentials.password,
        rememberMe
      );

      if (result.success && onLogin) {
        onLogin(result.user);
      }
    } catch (err) {
      setError(err.message || 'Login failed. Please check your credentials.');
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (e) => {
    const { name, value } = e.target;
    setCredentials(prev => ({
      ...prev,
      [name]: value
    }));
    // Clear validation error for this field
    if (validationErrors[name]) {
      setValidationErrors(prev => ({
        ...prev,
        [name]: null
      }));
    }
  };

  return (
    <div className="login-container">
      <div className="login-wrapper animate-slideInUp">
        <div className="login-card card">
          <div className="login-header">
            <h1 className="h2 text-center">OllamaMax</h1>
            <p className="text-center text-secondary">Enterprise AI Platform</p>
          </div>

          <form onSubmit={handleSubmit} className="login-form">
            {error && (
              <div className="alert alert-danger animate-fadeIn">
                <FontAwesomeIcon icon={faExclamationTriangle} className="me-2" />
                {error}
              </div>
            )}

            <div className="form-group">
              <label htmlFor="username">Username</label>
              <div className="input-group">
                <span className="input-icon">
                  <FontAwesomeIcon icon={faUser} />
                </span>
                <input
                  type="text"
                  id="username"
                  name="username"
                  className={`form-control ${validationErrors.username ? 'is-invalid' : ''}`}
                  placeholder="Enter your username"
                  value={credentials.username}
                  onChange={handleChange}
                  autoComplete="username"
                  disabled={loading}
                />
              </div>
              {validationErrors.username && (
                <div className="invalid-feedback">{validationErrors.username}</div>
              )}
            </div>

            <div className="form-group">
              <label htmlFor="password">Password</label>
              <div className="input-group">
                <span className="input-icon">
                  <FontAwesomeIcon icon={faLock} />
                </span>
                <input
                  type={showPassword ? 'text' : 'password'}
                  id="password"
                  name="password"
                  className={`form-control ${validationErrors.password ? 'is-invalid' : ''}`}
                  placeholder="Enter your password"
                  value={credentials.password}
                  onChange={handleChange}
                  autoComplete="current-password"
                  disabled={loading}
                />
                <button
                  type="button"
                  className="btn-icon"
                  onClick={() => setShowPassword(!showPassword)}
                  tabIndex={-1}
                >
                  <FontAwesomeIcon icon={showPassword ? faEyeSlash : faEye} />
                </button>
              </div>
              {validationErrors.password && (
                <div className="invalid-feedback">{validationErrors.password}</div>
              )}
            </div>

            <div className="form-check mb-3">
              <input
                type="checkbox"
                className="form-check-input"
                id="rememberMe"
                checked={rememberMe}
                onChange={(e) => setRememberMe(e.target.checked)}
              />
              <label className="form-check-label" htmlFor="rememberMe">
                Remember me
              </label>
            </div>

            <div className="form-actions">
              <button
                type="submit"
                className="btn btn-primary btn-block"
                disabled={loading}
              >
                {loading ? (
                  <>
                    <span className="spinner-border spinner-border-sm me-2" />
                    Signing in...
                  </>
                ) : (
                  <>
                    <FontAwesomeIcon icon={faSignInAlt} className="me-2" />
                    Sign In
                  </>
                )}
              </button>
            </div>

            <div className="login-footer">
              <div className="d-flex justify-content-between align-items-center">
                <button
                  type="button"
                  className="btn btn-link text-muted p-0"
                  onClick={onForgotPassword}
                >
                  Forgot password?
                </button>
                <button
                  type="button"
                  className="btn btn-outline-primary btn-sm"
                  onClick={onRegister}
                >
                  <FontAwesomeIcon icon={faUserPlus} className="me-1" />
                  Register
                </button>
              </div>
            </div>
          </form>
        </div>

        <div className="login-info">
          <p className="text-center text-muted">
            Secure distributed AI inference platform
          </p>
        </div>
      </div>

      <style jsx>{`
        .login-container {
          min-height: 100vh;
          display: flex;
          align-items: center;
          justify-content: center;
          background: linear-gradient(135deg, var(--primary-color) 0%, var(--primary-dark) 100%);
          padding: var(--spacing-lg);
        }

        .login-wrapper {
          width: 100%;
          max-width: 400px;
        }

        .login-card {
          background: var(--background);
          padding: var(--spacing-2xl);
          border-radius: var(--radius-xl);
          box-shadow: var(--shadow-xl);
        }

        .login-header {
          margin-bottom: var(--spacing-xl);
        }

        .login-header h1 {
          color: var(--primary-color);
          margin-bottom: var(--spacing-xs);
        }

        .form-group {
          margin-bottom: var(--spacing-lg);
        }

        .form-group label {
          display: block;
          font-weight: var(--font-weight-medium);
          margin-bottom: var(--spacing-xs);
          color: var(--text-primary);
        }

        .input-group {
          position: relative;
          display: flex;
          align-items: center;
        }

        .input-icon {
          position: absolute;
          left: var(--spacing-md);
          color: var(--text-muted);
          z-index: 1;
        }

        .form-control {
          width: 100%;
          padding: var(--spacing-sm) var(--spacing-md);
          padding-left: calc(var(--spacing-xl) + var(--spacing-sm));
          font-size: var(--font-size-base);
          border: 1px solid var(--border-color);
          border-radius: var(--radius-md);
          background: var(--background);
          color: var(--text-primary);
          transition: all var(--transition-fast);
        }

        .form-control:focus {
          outline: none;
          border-color: var(--primary-color);
          box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
        }

        .form-control.is-invalid {
          border-color: var(--danger-color);
        }

        .form-control:disabled {
          opacity: 0.5;
          cursor: not-allowed;
        }

        .btn-icon {
          position: absolute;
          right: var(--spacing-md);
          background: transparent;
          border: none;
          color: var(--text-muted);
          cursor: pointer;
          padding: var(--spacing-xs);
        }

        .btn-icon:hover {
          color: var(--text-primary);
        }

        .invalid-feedback {
          color: var(--danger-color);
          font-size: var(--font-size-sm);
          margin-top: var(--spacing-xs);
        }

        .alert {
          padding: var(--spacing-md);
          border-radius: var(--radius-md);
          margin-bottom: var(--spacing-lg);
        }

        .alert-danger {
          background: rgba(239, 68, 68, 0.1);
          color: var(--danger-color);
          border: 1px solid rgba(239, 68, 68, 0.2);
        }

        .btn-block {
          width: 100%;
        }

        .form-actions {
          margin-top: var(--spacing-xl);
        }

        .login-footer {
          text-align: center;
          margin-top: var(--spacing-lg);
        }

        .login-info {
          margin-top: var(--spacing-lg);
        }

        .spinner-border {
          display: inline-block;
          width: 1rem;
          height: 1rem;
          border: 2px solid currentColor;
          border-right-color: transparent;
          border-radius: 50%;
          animation: spinner-border 0.75s linear infinite;
        }

        @keyframes spinner-border {
          to {
            transform: rotate(360deg);
          }
        }

        @media (max-width: 640px) {
          .login-card {
            padding: var(--spacing-lg);
          }
        }
      `}</style>
    </div>
  );
};

export default Login;
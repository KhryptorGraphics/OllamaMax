/**
 * Authentication Routes for Ollamamax API
 * Handles user registration, login, token refresh, and logout
 */

const express = require('express');
const { body, validationResult } = require('express-validator');
const authMiddleware = require('../middleware/auth');
const UserModel = require('../models/user');

const router = express.Router();
const userModel = new UserModel();

/**
 * Validation middleware
 */
const validateRegistration = [
  body('email')
    .isEmail()
    .normalizeEmail()
    .withMessage('Valid email is required'),
  body('password')
    .isLength({ min: 8 })
    .matches(/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]/)
    .withMessage('Password must be at least 8 characters with uppercase, lowercase, number, and special character'),
  body('confirmPassword')
    .custom((value, { req }) => {
      if (value !== req.body.password) {
        throw new Error('Password confirmation does not match password');
      }
      return true;
    })
];

const validateLogin = [
  body('email')
    .isEmail()
    .normalizeEmail()
    .withMessage('Valid email is required'),
  body('password')
    .notEmpty()
    .withMessage('Password is required')
];

/**
 * POST /auth/register
 * Register new user account
 */
router.post('/register', authMiddleware.authRateLimit(), validateRegistration, async (req, res) => {
  try {
    // Check for validation errors
    const errors = validationResult(req);
    if (!errors.isEmpty()) {
      return res.status(400).json({
        error: {
          message: 'Validation failed',
          type: 'invalid_request_error',
          code: 'validation_error',
          details: errors.array()
        }
      });
    }

    const { email, password, firstName, lastName } = req.body;

    // Create user
    const user = await userModel.createUser({
      email,
      password,
      role: 'user',
      permissions: ['inference'],
      metadata: {
        firstName: firstName || '',
        lastName: lastName || '',
        registrationDate: new Date().toISOString()
      }
    });

    // Generate tokens
    const accessToken = authMiddleware.generateAccessToken({
      sub: user.id,
      email: user.email,
      role: user.role,
      permissions: user.permissions
    });

    const refreshToken = authMiddleware.generateRefreshToken({
      sub: user.id,
      email: user.email
    });

    // Create session
    await userModel.createSession(
      user.id,
      refreshToken,
      req.ip,
      req.get('User-Agent') || ''
    );

    res.status(201).json({
      message: 'User registered successfully',
      user: {
        id: user.id,
        email: user.email,
        role: user.role,
        permissions: user.permissions,
        api_key: user.api_key
      },
      access_token: accessToken,
      refresh_token: refreshToken,
      token_type: 'Bearer',
      expires_in: 86400 // 24 hours
    });

  } catch (error) {
    console.error('Registration error:', error);
    
    let statusCode = 500;
    let errorCode = 'registration_error';
    
    if (error.message === 'Email already exists') {
      statusCode = 409;
      errorCode = 'email_exists';
    }

    res.status(statusCode).json({
      error: {
        message: error.message || 'Registration failed',
        type: 'authentication_error',
        code: errorCode
      }
    });
  }
});

/**
 * POST /auth/login
 * User login with email and password
 */
router.post('/login', authMiddleware.authRateLimit(), validateLogin, async (req, res) => {
  try {
    // Check for validation errors
    const errors = validationResult(req);
    if (!errors.isEmpty()) {
      return res.status(400).json({
        error: {
          message: 'Validation failed',
          type: 'invalid_request_error',
          code: 'validation_error',
          details: errors.array()
        }
      });
    }

    const { email, password } = req.body;

    // Authenticate user
    const user = await userModel.authenticate(email, password);

    // Generate tokens
    const accessToken = authMiddleware.generateAccessToken({
      sub: user.id,
      email: user.email,
      role: user.role,
      permissions: user.permissions
    });

    const refreshToken = authMiddleware.generateRefreshToken({
      sub: user.id,
      email: user.email
    });

    // Create session
    await userModel.createSession(
      user.id,
      refreshToken,
      req.ip,
      req.get('User-Agent') || ''
    );

    res.json({
      message: 'Login successful',
      user: {
        id: user.id,
        email: user.email,
        role: user.role,
        permissions: user.permissions
      },
      access_token: accessToken,
      refresh_token: refreshToken,
      token_type: 'Bearer',
      expires_in: 86400 // 24 hours
    });

  } catch (error) {
    console.error('Login error:', error);
    
    let statusCode = 401;
    let errorCode = 'authentication_failed';
    
    if (error.message === 'User not found') {
      errorCode = 'user_not_found';
    } else if (error.message === 'Invalid password') {
      errorCode = 'invalid_credentials';
    }

    res.status(statusCode).json({
      error: {
        message: 'Authentication failed',
        type: 'authentication_error',
        code: errorCode
      }
    });
  }
});

/**
 * POST /auth/refresh
 * Refresh access token using refresh token
 */
router.post('/refresh', async (req, res) => {
  try {
    const { refresh_token } = req.body;

    if (!refresh_token) {
      return res.status(400).json({
        error: {
          message: 'Refresh token required',
          type: 'invalid_request_error',
          code: 'missing_refresh_token'
        }
      });
    }

    // Validate refresh token
    const session = await userModel.validateRefreshToken(refresh_token);

    if (!session) {
      return res.status(401).json({
        error: {
          message: 'Invalid or expired refresh token',
          type: 'authentication_error',
          code: 'invalid_refresh_token'
        }
      });
    }

    // Generate new access token
    const accessToken = authMiddleware.generateAccessToken({
      sub: session.user_id,
      email: session.email,
      role: session.role,
      permissions: session.permissions
    });

    // Optionally generate new refresh token (token rotation)
    const newRefreshToken = authMiddleware.generateRefreshToken({
      sub: session.user_id,
      email: session.email
    });

    // Revoke old refresh token and create new session
    await userModel.revokeRefreshToken(refresh_token);
    await userModel.createSession(
      session.user_id,
      newRefreshToken,
      req.ip,
      req.get('User-Agent') || ''
    );

    res.json({
      access_token: accessToken,
      refresh_token: newRefreshToken,
      token_type: 'Bearer',
      expires_in: 86400
    });

  } catch (error) {
    console.error('Token refresh error:', error);
    
    res.status(401).json({
      error: {
        message: 'Token refresh failed',
        type: 'authentication_error',
        code: 'refresh_failed'
      }
    });
  }
});

/**
 * POST /auth/logout
 * Logout user and revoke refresh token
 */
router.post('/logout', authMiddleware.authenticate(), async (req, res) => {
  try {
    const { refresh_token } = req.body;

    if (refresh_token) {
      await userModel.revokeRefreshToken(refresh_token);
    }

    res.json({
      message: 'Logged out successfully'
    });

  } catch (error) {
    console.error('Logout error:', error);
    
    res.status(500).json({
      error: {
        message: 'Logout failed',
        type: 'server_error',
        code: 'logout_error'
      }
    });
  }
});

/**
 * GET /auth/me
 * Get current user profile
 */
router.get('/me', authMiddleware.authenticate(), async (req, res) => {
  try {
    const user = await userModel.findById(req.user.id);

    if (!user) {
      return res.status(404).json({
        error: {
          message: 'User not found',
          type: 'not_found_error',
          code: 'user_not_found'
        }
      });
    }

    // Get usage statistics
    const usage = await userModel.getUserUsage(req.user.id, 30);
    const quota = await userModel.getUserQuota(req.user.id);

    res.json({
      user: {
        id: user.id,
        email: user.email,
        role: user.role,
        permissions: user.permissions,
        created_at: user.created_at,
        last_login: user.last_login,
        email_verified: user.email_verified,
        metadata: user.metadata
      },
      usage: {
        last_30_days: usage,
        quota
      }
    });

  } catch (error) {
    console.error('Get profile error:', error);
    
    res.status(500).json({
      error: {
        message: 'Failed to retrieve profile',
        type: 'server_error',
        code: 'profile_error'
      }
    });
  }
});

/**
 * PUT /auth/me
 * Update user profile
 */
router.put('/me', authMiddleware.authenticate(), [
  body('firstName').optional().isLength({ min: 1, max: 50 }),
  body('lastName').optional().isLength({ min: 1, max: 50 }),
  body('preferences').optional().isObject()
], async (req, res) => {
  try {
    const errors = validationResult(req);
    if (!errors.isEmpty()) {
      return res.status(400).json({
        error: {
          message: 'Validation failed',
          type: 'invalid_request_error',
          code: 'validation_error',
          details: errors.array()
        }
      });
    }

    const { firstName, lastName, preferences } = req.body;
    
    // Get current user metadata
    const user = await userModel.findById(req.user.id);
    const metadata = user.metadata || {};

    // Update metadata
    if (firstName !== undefined) metadata.firstName = firstName;
    if (lastName !== undefined) metadata.lastName = lastName;
    if (preferences !== undefined) metadata.preferences = preferences;

    await userModel.updateUser(req.user.id, { metadata });

    res.json({
      message: 'Profile updated successfully',
      user: {
        id: user.id,
        email: user.email,
        role: user.role,
        metadata
      }
    });

  } catch (error) {
    console.error('Update profile error:', error);
    
    res.status(500).json({
      error: {
        message: 'Failed to update profile',
        type: 'server_error',
        code: 'update_error'
      }
    });
  }
});

/**
 * POST /auth/change-password
 * Change user password
 */
router.post('/change-password', authMiddleware.authenticate(), [
  body('current_password').notEmpty().withMessage('Current password is required'),
  body('new_password')
    .isLength({ min: 8 })
    .matches(/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]/)
    .withMessage('New password must meet security requirements')
], async (req, res) => {
  try {
    const errors = validationResult(req);
    if (!errors.isEmpty()) {
      return res.status(400).json({
        error: {
          message: 'Validation failed',
          type: 'invalid_request_error',
          code: 'validation_error',
          details: errors.array()
        }
      });
    }

    const { current_password, new_password } = req.body;

    // Get user with password hash
    const user = await userModel.findById(req.user.id);
    
    // Verify current password
    const isValidPassword = await userModel.verifyPassword(current_password, user.password_hash);
    
    if (!isValidPassword) {
      return res.status(400).json({
        error: {
          message: 'Current password is incorrect',
          type: 'authentication_error',
          code: 'invalid_current_password'
        }
      });
    }

    // Hash new password
    const newPasswordHash = await userModel.hashPassword(new_password);
    
    // Update password
    await userModel.updateUser(req.user.id, { 
      password_hash: newPasswordHash 
    });

    res.json({
      message: 'Password changed successfully'
    });

  } catch (error) {
    console.error('Change password error:', error);
    
    res.status(500).json({
      error: {
        message: 'Failed to change password',
        type: 'server_error',
        code: 'password_change_error'
      }
    });
  }
});

/**
 * GET /auth/usage
 * Get detailed usage statistics
 */
router.get('/usage', authMiddleware.authenticate(), async (req, res) => {
  try {
    const days = parseInt(req.query.days) || 30;
    const usage = await userModel.getUserUsage(req.user.id, days);
    const quota = await userModel.getUserQuota(req.user.id);

    const summary = usage.reduce((acc, day) => ({
      total_requests: acc.total_requests + day.requests_count,
      total_tokens: acc.total_tokens + day.tokens_used,
      total_compute: acc.total_compute + day.compute_seconds
    }), { total_requests: 0, total_tokens: 0, total_compute: 0 });

    res.json({
      period: {
        days,
        start_date: usage[usage.length - 1]?.date || null,
        end_date: usage[0]?.date || null
      },
      summary,
      quota,
      daily_usage: usage
    });

  } catch (error) {
    console.error('Get usage error:', error);
    
    res.status(500).json({
      error: {
        message: 'Failed to retrieve usage statistics',
        type: 'server_error',
        code: 'usage_error'
      }
    });
  }
});

module.exports = router;
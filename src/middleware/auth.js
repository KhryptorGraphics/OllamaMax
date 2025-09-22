/**
 * Authentication Middleware for Ollamamax API
 * Handles JWT token validation and user authentication
 */

const jwt = require('jsonwebtoken');
const rateLimit = require('express-rate-limit');
const crypto = require('crypto');

class AuthMiddleware {
  constructor() {
    this.jwtSecret = process.env.JWT_SECRET || this.generateSecret();
    this.jwtRefreshSecret = process.env.JWT_REFRESH_SECRET || this.generateSecret();
    this.tokenExpiry = process.env.JWT_EXPIRY || '24h';
    this.refreshTokenExpiry = process.env.JWT_REFRESH_EXPIRY || '7d';
  }

  generateSecret() {
    return crypto.randomBytes(64).toString('hex');
  }

  /**
   * Generate JWT access token
   */
  generateAccessToken(payload) {
    return jwt.sign(payload, this.jwtSecret, {
      expiresIn: this.tokenExpiry,
      issuer: 'ollamamax',
      audience: 'ollamamax-api'
    });
  }

  /**
   * Generate JWT refresh token
   */
  generateRefreshToken(payload) {
    return jwt.sign(payload, this.jwtRefreshSecret, {
      expiresIn: this.refreshTokenExpiry,
      issuer: 'ollamamax',
      audience: 'ollamamax-refresh'
    });
  }

  /**
   * Verify JWT access token
   */
  verifyAccessToken(token) {
    try {
      return jwt.verify(token, this.jwtSecret, {
        issuer: 'ollamamax',
        audience: 'ollamamax-api'
      });
    } catch (error) {
      throw new Error('Invalid or expired access token');
    }
  }

  /**
   * Verify JWT refresh token
   */
  verifyRefreshToken(token) {
    try {
      return jwt.verify(token, this.jwtRefreshSecret, {
        issuer: 'ollamamax',
        audience: 'ollamamax-refresh'
      });
    } catch (error) {
      throw new Error('Invalid or expired refresh token');
    }
  }

  /**
   * Authentication middleware
   */
  authenticate() {
    return async (req, res, next) => {
      try {
        const authHeader = req.headers.authorization;
        const apiKey = req.headers['x-api-key'];

        // Check for Bearer token
        if (authHeader && authHeader.startsWith('Bearer ')) {
          const token = authHeader.substring(7);
          const decoded = this.verifyAccessToken(token);
          
          req.user = {
            id: decoded.sub,
            email: decoded.email,
            role: decoded.role,
            permissions: decoded.permissions || []
          };
          
          return next();
        }

        // Check for API key (for OpenAI compatibility)
        if (apiKey) {
          // In production, validate API key against database
          // For now, accept any API key for development
          req.user = {
            id: 'api-user',
            email: 'api@ollamamax.com',
            role: 'user',
            permissions: ['inference']
          };
          
          return next();
        }

        // No authentication provided
        return res.status(401).json({
          error: {
            message: 'Authentication required',
            type: 'authentication_error',
            code: 'missing_auth'
          }
        });

      } catch (error) {
        console.error('Authentication error:', error);
        
        return res.status(401).json({
          error: {
            message: 'Invalid authentication credentials',
            type: 'authentication_error',
            code: 'invalid_auth'
          }
        });
      }
    };
  }

  /**
   * Optional authentication middleware (for public endpoints)
   */
  optionalAuth() {
    return async (req, res, next) => {
      try {
        const authHeader = req.headers.authorization;
        const apiKey = req.headers['x-api-key'];

        if (authHeader && authHeader.startsWith('Bearer ')) {
          const token = authHeader.substring(7);
          const decoded = this.verifyAccessToken(token);
          
          req.user = {
            id: decoded.sub,
            email: decoded.email,
            role: decoded.role,
            permissions: decoded.permissions || []
          };
        } else if (apiKey) {
          req.user = {
            id: 'api-user',
            email: 'api@ollamamax.com',
            role: 'user',
            permissions: ['inference']
          };
        }

        next();
      } catch (error) {
        // Continue without authentication for optional auth
        next();
      }
    };
  }

  /**
   * Role-based authorization middleware
   */
  authorize(requiredRole = 'user', requiredPermissions = []) {
    return (req, res, next) => {
      if (!req.user) {
        return res.status(401).json({
          error: {
            message: 'Authentication required',
            type: 'authentication_error',
            code: 'missing_auth'
          }
        });
      }

      // Check role hierarchy
      const roleHierarchy = {
        'guest': 0,
        'user': 1,
        'premium': 2,
        'admin': 3,
        'superadmin': 4
      };

      const userRoleLevel = roleHierarchy[req.user.role] || 0;
      const requiredRoleLevel = roleHierarchy[requiredRole] || 1;

      if (userRoleLevel < requiredRoleLevel) {
        return res.status(403).json({
          error: {
            message: 'Insufficient permissions',
            type: 'authorization_error',
            code: 'insufficient_role'
          }
        });
      }

      // Check specific permissions
      if (requiredPermissions.length > 0) {
        const hasPermissions = requiredPermissions.every(permission => 
          req.user.permissions.includes(permission)
        );

        if (!hasPermissions) {
          return res.status(403).json({
            error: {
              message: 'Missing required permissions',
              type: 'authorization_error',
              code: 'missing_permissions',
              required: requiredPermissions
            }
          });
        }
      }

      next();
    };
  }

  /**
   * Rate limiting middleware for authentication endpoints
   */
  authRateLimit() {
    return rateLimit({
      windowMs: 15 * 60 * 1000, // 15 minutes
      max: 5, // limit each IP to 5 requests per windowMs
      message: {
        error: {
          message: 'Too many authentication attempts, please try again later',
          type: 'rate_limit_error',
          code: 'auth_rate_limit'
        }
      },
      standardHeaders: true,
      legacyHeaders: false,
      skip: (req) => {
        // Skip rate limiting for certain IPs or conditions
        return process.env.NODE_ENV === 'development';
      }
    });
  }

  /**
   * API rate limiting middleware
   */
  apiRateLimit(requestsPerMinute = 60) {
    return rateLimit({
      windowMs: 60 * 1000, // 1 minute
      max: requestsPerMinute,
      message: {
        error: {
          message: 'API rate limit exceeded',
          type: 'rate_limit_error',
          code: 'api_rate_limit'
        }
      },
      standardHeaders: true,
      legacyHeaders: false,
      skip: (req) => {
        // Skip for admin users
        return req.user && req.user.role === 'admin';
      }
    });
  }

  /**
   * Token usage tracking middleware
   */
  trackTokenUsage() {
    return async (req, res, next) => {
      // Store original end function
      const originalEnd = res.end;
      
      res.end = function(chunk, encoding) {
        // Track token usage if available in response
        if (res.tokenUsage) {
          // In production, store token usage in database
          console.log(`Token usage for ${req.user?.id || 'anonymous'}:`, res.tokenUsage);
        }
        
        // Call original end function
        originalEnd.call(this, chunk, encoding);
      };
      
      next();
    };
  }

  /**
   * Refresh token endpoint handler
   */
  async refreshToken(req, res) {
    try {
      const { refreshToken } = req.body;
      
      if (!refreshToken) {
        return res.status(400).json({
          error: {
            message: 'Refresh token required',
            type: 'invalid_request_error',
            code: 'missing_refresh_token'
          }
        });
      }

      const decoded = this.verifyRefreshToken(refreshToken);
      
      // Generate new access token
      const newAccessToken = this.generateAccessToken({
        sub: decoded.sub,
        email: decoded.email,
        role: decoded.role,
        permissions: decoded.permissions
      });

      // Optionally generate new refresh token
      const newRefreshToken = this.generateRefreshToken({
        sub: decoded.sub,
        email: decoded.email
      });

      res.json({
        access_token: newAccessToken,
        refresh_token: newRefreshToken,
        token_type: 'Bearer',
        expires_in: 86400 // 24 hours
      });

    } catch (error) {
      console.error('Token refresh error:', error);
      
      res.status(401).json({
        error: {
          message: 'Invalid refresh token',
          type: 'authentication_error',
          code: 'invalid_refresh_token'
        }
      });
    }
  }

  /**
   * CORS middleware configuration
   */
  corsConfig() {
    return {
      origin: process.env.CORS_ORIGINS?.split(',') || ['http://localhost:3000'],
      credentials: true,
      methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
      allowedHeaders: [
        'Origin',
        'X-Requested-With',
        'Content-Type',
        'Accept',
        'Authorization',
        'X-API-Key'
      ],
      exposedHeaders: [
        'X-RateLimit-Limit',
        'X-RateLimit-Remaining',
        'X-RateLimit-Reset'
      ]
    };
  }

  /**
   * Security headers middleware
   */
  securityHeaders() {
    return (req, res, next) => {
      // Remove server header
      res.removeHeader('X-Powered-By');
      
      // Security headers
      res.setHeader('X-Content-Type-Options', 'nosniff');
      res.setHeader('X-Frame-Options', 'DENY');
      res.setHeader('X-XSS-Protection', '1; mode=block');
      res.setHeader('Referrer-Policy', 'strict-origin-when-cross-origin');
      
      // API-specific headers
      res.setHeader('Content-Type', 'application/json');
      res.setHeader('Cache-Control', 'no-store');
      
      next();
    };
  }

  /**
   * Error handling middleware
   */
  errorHandler() {
    return (error, req, res, next) => {
      console.error('Auth middleware error:', error);

      // JWT specific errors
      if (error.name === 'JsonWebTokenError') {
        return res.status(401).json({
          error: {
            message: 'Invalid token format',
            type: 'authentication_error',
            code: 'invalid_token'
          }
        });
      }

      if (error.name === 'TokenExpiredError') {
        return res.status(401).json({
          error: {
            message: 'Token has expired',
            type: 'authentication_error',
            code: 'token_expired'
          }
        });
      }

      // Default error response
      res.status(500).json({
        error: {
          message: 'Internal server error',
          type: 'server_error',
          code: 'internal_error'
        }
      });
    };
  }
}

// Export singleton instance
const authMiddleware = new AuthMiddleware();
module.exports = authMiddleware;
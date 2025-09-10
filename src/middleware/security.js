/**
 * Security Middleware
 * Implements comprehensive security measures for Express applications
 */

// Security middleware imports - install packages as needed:
// npm install helmet express-rate-limit express-slow-down compression
let helmet, rateLimit, slowDown, compression;
try {
  helmet = require('helmet');
} catch (e) { console.warn('helmet not installed:', e.message); }
try {
  rateLimit = require('express-rate-limit');
} catch (e) { console.warn('express-rate-limit not installed:', e.message); }
try {
  slowDown = require('express-slow-down');
} catch (e) { console.warn('express-slow-down not installed:', e.message); }
try {
  compression = require('compression');
} catch (e) { console.warn('compression not installed:', e.message); }

import cors from 'cors';
import securityConfig from '../config/security.js';

/**
 * Configure and return security middleware stack
 */
export function setupSecurityMiddleware(app) {
  // Trust proxy for rate limiting (if behind reverse proxy)
  if (process.env.TRUST_PROXY) {
    app.set('trust proxy', 1);
  }

  // Compression with security considerations (if available)
  if (compression) {
    app.use(compression({
      filter: (req, res) => {
        // Don't compress responses that contain sensitive data
        if (req.headers['x-no-compression']) {
          return false;
        }
        return compression.filter(req, res);
      },
      level: 6, // Balance between compression and CPU usage
      threshold: 1024 // Only compress responses larger than 1KB
    }));
  }

  // CORS configuration
  app.use(cors(securityConfig.cors));

  // Helmet security headers (if available)
  if (helmet) {
    app.use(helmet({
      contentSecurityPolicy: securityConfig.headers.contentSecurityPolicy,
      hsts: process.env.NODE_ENV === 'production' ? securityConfig.headers.hsts : false,
      frameguard: securityConfig.headers.frameguard,
      noSniff: securityConfig.headers.noSniff,
      referrerPolicy: securityConfig.headers.referrerPolicy,
      xssFilter: securityConfig.headers.xssFilter,
      permissionsPolicy: securityConfig.headers.permissionsPolicy
    }));
  }

  // Custom security headers
  app.use((req, res, next) => {
    // Remove server information
    res.removeHeader('X-Powered-By');
    
    // Add custom security headers
    res.setHeader('X-Content-Type-Options', 'nosniff');
    res.setHeader('X-Frame-Options', 'DENY');
    res.setHeader('X-XSS-Protection', '1; mode=block');
    
    // Add CSRF token to response headers for SPA
    if (req.csrfToken) {
      res.setHeader('X-CSRF-Token', req.csrfToken());
    }
    
    next();
  });

  return app;
}

/**
 * Rate limiting middleware with different levels
 */
export const rateLimiters = rateLimit ? {
  // General API rate limiting
  general: rateLimit({
    ...securityConfig.rateLimiting,
    keyGenerator: (req) => {
      // Use combination of IP and user ID if authenticated
      return req.user ? `${req.ip}:${req.user.id}` : req.ip;
    }
  }),

  // Strict rate limiting for authentication endpoints
  auth: rateLimit({
    windowMs: 15 * 60 * 1000, // 15 minutes
    max: 5, // 5 attempts per window
    message: {
      error: 'Too many authentication attempts, please try again later',
      code: 'AUTH_RATE_LIMIT_EXCEEDED'
    },
    standardHeaders: true,
    legacyHeaders: false,
    skipSuccessfulRequests: true // Only count failed attempts
  }),

  // API key endpoint rate limiting
  apiKey: rateLimit({
    windowMs: 60 * 60 * 1000, // 1 hour
    max: 10, // 10 API key requests per hour
    message: {
      error: 'API key generation rate limit exceeded',
      code: 'API_KEY_RATE_LIMIT_EXCEEDED'
    }
  }),

  // File upload rate limiting
  upload: rateLimit({
    windowMs: 60 * 1000, // 1 minute
    max: 10, // 10 uploads per minute
    message: {
      error: 'Upload rate limit exceeded, please try again later',
      code: 'UPLOAD_RATE_LIMIT_EXCEEDED'
    }
  })
} : {};

/**
 * Slow down middleware for progressive delays
 */
export const slowDownMiddleware = slowDown ? slowDown({
  windowMs: 15 * 60 * 1000, // 15 minutes
  delayAfter: 5, // Allow 5 requests per windowMs without delay
  delayMs: 500, // Add 500ms delay per request after delayAfter
  maxDelayMs: 20000, // Maximum delay of 20 seconds
  skipSuccessfulRequests: true
}) : (req, res, next) => next(); // No-op if slowDown not available

/**
 * Input validation and sanitization middleware
 */
export function validateAndSanitizeInput(options = {}) {
  return (req, res, next) => {
    try {
      // Sanitize query parameters
      if (req.query) {
        for (const key in req.query) {
          if (typeof req.query[key] === 'string') {
            req.query[key] = securityConfig.sanitizeInput(req.query[key], options);
          }
        }
      }

      // Sanitize request body (only for non-file uploads)
      if (req.body && req.is('application/json')) {
        req.body = sanitizeObject(req.body, options);
      }

      next();
    } catch (error) {
      res.status(400).json({
        error: 'Invalid input data',
        code: 'INVALID_INPUT',
        details: error.message
      });
    }
  };
}

/**
 * Recursively sanitize object properties
 */
function sanitizeObject(obj, options) {
  if (typeof obj !== 'object' || obj === null) {
    return obj;
  }

  const sanitized = Array.isArray(obj) ? [] : {};

  for (const key in obj) {
    if (obj.hasOwnProperty(key)) {
      const value = obj[key];
      
      if (typeof value === 'string') {
        sanitized[key] = securityConfig.sanitizeInput(value, options);
      } else if (typeof value === 'object') {
        sanitized[key] = sanitizeObject(value, options);
      } else {
        sanitized[key] = value;
      }
    }
  }

  return sanitized;
}

/**
 * CSRF protection middleware
 */
export function csrfProtection(options = {}) {
  return (req, res, next) => {
    // Skip CSRF for safe methods and API endpoints with valid API keys
    if (req.method === 'GET' || req.method === 'HEAD' || req.method === 'OPTIONS') {
      return next();
    }

    // Skip for API endpoints with valid API key
    if (req.headers['x-api-key'] && isValidApiKey(req.headers['x-api-key'])) {
      return next();
    }

    const token = req.headers['x-csrf-token'] || req.body._csrf || req.query._csrf;
    const sessionToken = req.session && req.session.csrfToken;

    if (!securityConfig.verifyCSRFToken(token, sessionToken)) {
      return res.status(403).json({
        error: 'Invalid CSRF token',
        code: 'CSRF_TOKEN_INVALID'
      });
    }

    next();
  };
}

/**
 * API key validation helper
 */
function isValidApiKey(apiKey) {
  // Implement your API key validation logic here
  // This is a placeholder implementation
  if (!apiKey || typeof apiKey !== 'string' || apiKey.length < 32) {
    return false;
  }
  
  // Add database lookup or cache check for valid API keys
  return true; // Placeholder
}

/**
 * Authentication middleware
 */
export function requireAuth(options = {}) {
  return async (req, res, next) => {
    try {
      const token = extractToken(req);
      
      if (!token) {
        return res.status(401).json({
          error: 'Authentication required',
          code: 'AUTH_REQUIRED'
        });
      }

      const decoded = securityConfig.verifyJWT(token);
      
      // Attach user information to request
      req.user = decoded;
      req.userId = decoded.sub || decoded.id;

      // Check if user is active (optional)
      if (options.checkActive && decoded.status !== 'active') {
        return res.status(403).json({
          error: 'Account is not active',
          code: 'ACCOUNT_INACTIVE'
        });
      }

      next();
    } catch (error) {
      if (error.name === 'TokenExpiredError') {
        return res.status(401).json({
          error: 'Token has expired',
          code: 'TOKEN_EXPIRED'
        });
      } else if (error.name === 'JsonWebTokenError') {
        return res.status(401).json({
          error: 'Invalid token',
          code: 'TOKEN_INVALID'
        });
      }

      return res.status(500).json({
        error: 'Authentication error',
        code: 'AUTH_ERROR'
      });
    }
  };
}

/**
 * Authorization middleware for role-based access control
 */
export function requireRole(roles = []) {
  return (req, res, next) => {
    if (!req.user) {
      return res.status(401).json({
        error: 'Authentication required',
        code: 'AUTH_REQUIRED'
      });
    }

    const userRoles = Array.isArray(req.user.roles) ? req.user.roles : [req.user.role];
    const requiredRoles = Array.isArray(roles) ? roles : [roles];

    const hasRequiredRole = requiredRoles.some(role => userRoles.includes(role));

    if (!hasRequiredRole) {
      return res.status(403).json({
        error: 'Insufficient permissions',
        code: 'INSUFFICIENT_PERMISSIONS',
        required: requiredRoles,
        current: userRoles
      });
    }

    next();
  };
}

/**
 * Extract token from request
 */
function extractToken(req) {
  // Check Authorization header
  const authHeader = req.headers.authorization;
  if (authHeader && authHeader.startsWith('Bearer ')) {
    return authHeader.slice(7);
  }

  // Check cookies
  if (req.cookies && req.cookies.token) {
    return req.cookies.token;
  }

  // Check query parameter (less secure, only for specific use cases)
  if (req.query.token) {
    return req.query.token;
  }

  return null;
}

/**
 * Request logging middleware for security monitoring
 */
export function securityLogger(options = {}) {
  return (req, res, next) => {
    const start = Date.now();
    
    // Log security-relevant request information
    const logData = {
      timestamp: new Date().toISOString(),
      ip: req.ip,
      method: req.method,
      url: req.originalUrl,
      userAgent: req.get('User-Agent'),
      referer: req.get('Referer'),
      userId: req.user ? req.user.id : null
    };

    // Log suspicious patterns
    const suspiciousPatterns = [
      /\.\./g, // Path traversal
      /<script/gi, // XSS attempts
      /union.*select/gi, // SQL injection
      /eval\(/gi, // Code injection
      /javascript:/gi // Protocol injection
    ];

    const requestContent = JSON.stringify(req.query) + JSON.stringify(req.body);
    const suspiciousAttempt = suspiciousPatterns.some(pattern => pattern.test(requestContent));

    if (suspiciousAttempt) {
      logData.suspicious = true;
      logData.patterns = suspiciousPatterns.filter(pattern => pattern.test(requestContent));
      console.warn('SECURITY: Suspicious request detected', logData);
    }

    // Continue processing
    next();

    // Log response information
    res.on('finish', () => {
      logData.statusCode = res.statusCode;
      logData.responseTime = Date.now() - start;
      
      // Log failed authentication attempts
      if (res.statusCode === 401 || res.statusCode === 403) {
        console.warn('SECURITY: Authentication/Authorization failure', logData);
      }
    });
  };
}

/**
 * File upload security middleware
 */
export function secureFileUpload(options = {}) {
  const allowedTypes = options.allowedTypes || ['.txt', '.json', '.csv', '.md', '.pdf'];
  const maxSize = options.maxSize || 10 * 1024 * 1024; // 10MB default
  const maxFiles = options.maxFiles || 5;

  return (req, res, next) => {
    if (!req.files || req.files.length === 0) {
      return next();
    }

    // Check file count
    if (req.files.length > maxFiles) {
      return res.status(400).json({
        error: `Too many files. Maximum ${maxFiles} allowed.`,
        code: 'TOO_MANY_FILES'
      });
    }

    // Validate each file
    for (const file of req.files) {
      // Check file size
      if (file.size > maxSize) {
        return res.status(400).json({
          error: `File ${file.originalname} exceeds maximum size of ${maxSize} bytes`,
          code: 'FILE_TOO_LARGE'
        });
      }

      // Check file extension
      const ext = file.originalname.toLowerCase().match(/\.[^.]*$/)?.[0];
      if (!ext || !allowedTypes.includes(ext)) {
        return res.status(400).json({
          error: `File type ${ext} not allowed. Allowed types: ${allowedTypes.join(', ')}`,
          code: 'FILE_TYPE_NOT_ALLOWED'
        });
      }

      // Check for null bytes in filename
      if (file.originalname.includes('\0')) {
        return res.status(400).json({
          error: 'Invalid filename',
          code: 'INVALID_FILENAME'
        });
      }
    }

    next();
  };
}

export default {
  setupSecurityMiddleware,
  rateLimiters,
  slowDownMiddleware,
  validateAndSanitizeInput,
  csrfProtection,
  requireAuth,
  requireRole,
  securityLogger,
  secureFileUpload
};
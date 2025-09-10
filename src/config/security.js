/**
 * Security Configuration Module
 * Centralized security settings and validation
 */

import crypto from 'crypto';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';
import { readFileSync, existsSync } from 'fs';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

/**
 * Validates environment variables for security compliance
 */
class SecurityConfig {
  constructor() {
    this.validateEnvironment();
    this.secrets = this.loadSecrets();
    this.rateLimiting = this.getRateLimitConfig();
    this.cors = this.getCorsConfig();
    this.session = this.getSessionConfig();
    this.headers = this.getSecurityHeaders();
  }

  /**
   * Validate critical environment variables
   */
  validateEnvironment() {
    const required = [
      'NODE_ENV',
      'SESSION_SECRET',
      'JWT_SECRET'
    ];

    const missing = required.filter(key => !process.env[key]);
    if (missing.length > 0) {
      throw new Error(`Missing required environment variables: ${missing.join(', ')}`);
    }

    // Validate secret strength
    if (process.env.SESSION_SECRET.length < 32) {
      throw new Error('SESSION_SECRET must be at least 32 characters long');
    }

    if (process.env.JWT_SECRET.length < 32) {
      throw new Error('JWT_SECRET must be at least 32 characters long');
    }

    // Warn about default values in production
    if (process.env.NODE_ENV === 'production') {
      this.validateProductionSecrets();
    }
  }

  /**
   * Validate production secrets are not defaults
   */
  validateProductionSecrets() {
    const dangerous = [
      'your-secret-session-key-here',
      'your-jwt-secret-key-here',
      'password',
      '123456',
      'secret'
    ];

    const sessionSecret = process.env.SESSION_SECRET.toLowerCase();
    const jwtSecret = process.env.JWT_SECRET.toLowerCase();

    for (const danger of dangerous) {
      if (sessionSecret.includes(danger) || jwtSecret.includes(danger)) {
        throw new Error(
          'SECURITY WARNING: Default or weak secrets detected in production environment! ' +
          'Please use strong, unique secrets.'
        );
      }
    }
  }

  /**
   * Load and validate secrets
   */
  loadSecrets() {
    return {
      session: process.env.SESSION_SECRET,
      jwt: process.env.JWT_SECRET,
      encryption: process.env.ENCRYPTION_KEY || this.generateKey(32),
      database: {
        password: process.env.DB_PASSWORD,
        host: process.env.DB_HOST || 'localhost',
        ssl: process.env.DB_SSL === 'require'
      },
      redis: {
        password: process.env.REDIS_PASSWORD,
        url: process.env.REDIS_URL || 'redis://localhost:6379'
      },
      external: {
        openai: process.env.OPENAI_API_KEY,
        anthropic: process.env.ANTHROPIC_API_KEY,
        huggingface: process.env.HUGGINGFACE_API_KEY,
        github: process.env.GITHUB_TOKEN
      }
    };
  }

  /**
   * Generate cryptographically secure key
   */
  generateKey(length = 32) {
    return crypto.randomBytes(length).toString('hex');
  }

  /**
   * Get rate limiting configuration
   */
  getRateLimitConfig() {
    return {
      windowMs: parseInt(process.env.RATE_LIMIT_WINDOW_MS) || 15 * 60 * 1000, // 15 minutes
      max: parseInt(process.env.RATE_LIMIT_MAX_REQUESTS) || 100,
      message: {
        error: 'Too many requests from this IP, please try again later',
        code: 'RATE_LIMIT_EXCEEDED'
      },
      standardHeaders: true,
      legacyHeaders: false,
      // Skip rate limiting for health checks
      skip: (req) => req.path === '/health' || req.path === '/metrics'
    };
  }

  /**
   * Get CORS configuration
   */
  getCorsConfig() {
    const allowedOrigins = process.env.CORS_ORIGIN 
      ? process.env.CORS_ORIGIN.split(',').map(origin => origin.trim())
      : ['http://localhost:3000', 'http://localhost:3001'];

    return {
      origin: (origin, callback) => {
        // Allow requests with no origin (mobile apps, curl, etc.)
        if (!origin) return callback(null, true);
        
        if (allowedOrigins.includes(origin) || 
            (process.env.NODE_ENV === 'development' && origin.startsWith('http://localhost'))) {
          return callback(null, true);
        }
        
        return callback(new Error('Not allowed by CORS policy'), false);
      },
      credentials: true,
      methods: ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS'],
      allowedHeaders: [
        'Origin',
        'X-Requested-With',
        'Content-Type',
        'Accept',
        'Authorization',
        'X-API-Key',
        'X-CSRF-Token'
      ],
      exposedHeaders: ['X-Total-Count', 'X-Page-Count'],
      maxAge: 86400 // 24 hours
    };
  }

  /**
   * Get session configuration
   */
  getSessionConfig() {
    return {
      secret: this.secrets.session,
      name: 'ollamamax.sid',
      resave: false,
      saveUninitialized: false,
      rolling: true,
      cookie: {
        secure: process.env.NODE_ENV === 'production', // HTTPS only in production
        httpOnly: true, // Prevent XSS
        maxAge: 24 * 60 * 60 * 1000, // 24 hours
        sameSite: 'strict' // CSRF protection
      },
      store: null // Will be set to Redis store if available
    };
  }

  /**
   * Get security headers configuration
   */
  getSecurityHeaders() {
    return {
      // Content Security Policy
      contentSecurityPolicy: {
        directives: {
          defaultSrc: ["'self'"],
          styleSrc: ["'self'", "'unsafe-inline'", 'https://fonts.googleapis.com'],
          scriptSrc: ["'self'", "'unsafe-eval'"], // Note: unsafe-eval needed for some AI model execution
          imgSrc: ["'self'", 'data:', 'https:'],
          fontSrc: ["'self'", 'https://fonts.gstatic.com'],
          connectSrc: ["'self'", 'wss:', 'https:'],
          frameSrc: ["'none'"],
          objectSrc: ["'none'"],
          mediaSrc: ["'self'"],
          workerSrc: ["'self'", 'blob:']
        }
      },
      
      // HTTP Strict Transport Security
      hsts: {
        maxAge: 31536000, // 1 year
        includeSubDomains: true,
        preload: true
      },
      
      // X-Frame-Options
      frameguard: {
        action: 'deny'
      },
      
      // X-Content-Type-Options
      noSniff: true,
      
      // Referrer Policy
      referrerPolicy: {
        policy: 'strict-origin-when-cross-origin'
      },
      
      // X-XSS-Protection (legacy browsers)
      xssFilter: true,
      
      // Permissions Policy
      permissionsPolicy: {
        camera: ['none'],
        microphone: ['none'],
        geolocation: ['none'],
        payment: ['none'],
        usb: ['none']
      }
    };
  }

  /**
   * Validate API key format and strength
   */
  validateApiKey(apiKey, service = 'unknown') {
    if (!apiKey || typeof apiKey !== 'string') {
      throw new Error(`Invalid API key for service: ${service}`);
    }

    if (apiKey.length < 16) {
      throw new Error(`API key too short for service: ${service}`);
    }

    // Check for obvious test/placeholder values
    const invalidPatterns = [
      /^test/i,
      /^demo/i,
      /^example/i,
      /^placeholder/i,
      /^your[-_].*key/i,
      /^123+/,
      /^abc+/i
    ];

    for (const pattern of invalidPatterns) {
      if (pattern.test(apiKey)) {
        throw new Error(`Invalid/placeholder API key detected for service: ${service}`);
      }
    }

    return true;
  }

  /**
   * Sanitize user input to prevent injection attacks
   */
  sanitizeInput(input, options = {}) {
    if (typeof input !== 'string') {
      return input;
    }

    let sanitized = input;

    // Remove null bytes
    sanitized = sanitized.replace(/\0/g, '');

    // HTML encode dangerous characters
    if (options.html !== false) {
      sanitized = sanitized
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#x27;')
        .replace(/\//g, '&#x2F;');
    }

    // Remove or escape SQL injection patterns
    if (options.sql !== false) {
      sanitized = sanitized.replace(/['";\\]/g, '\\$&');
    }

    // Remove potential script injection
    if (options.script !== false) {
      sanitized = sanitized.replace(/<script[^>]*>.*?<\/script>/gi, '');
      sanitized = sanitized.replace(/javascript:/gi, '');
      sanitized = sanitized.replace(/on\w+\s*=/gi, '');
    }

    // Limit length to prevent DoS
    const maxLength = options.maxLength || 10000;
    if (sanitized.length > maxLength) {
      sanitized = sanitized.substring(0, maxLength);
    }

    return sanitized;
  }

  /**
   * Generate CSRF token
   */
  generateCSRFToken() {
    return crypto.randomBytes(32).toString('hex');
  }

  /**
   * Verify CSRF token
   */
  verifyCSRFToken(token, sessionToken) {
    if (!token || !sessionToken) {
      return false;
    }

    return crypto.timingSafeEqual(
      Buffer.from(token, 'hex'),
      Buffer.from(sessionToken, 'hex')
    );
  }

  /**
   * Hash password securely
   */
  async hashPassword(password) {
    const bcrypt = await import('bcrypt');
    const saltRounds = 12;
    return bcrypt.hash(password, saltRounds);
  }

  /**
   * Verify password hash
   */
  async verifyPassword(password, hash) {
    const bcrypt = await import('bcrypt');
    return bcrypt.compare(password, hash);
  }

  /**
   * Generate secure JWT token
   */
  generateJWT(payload, options = {}) {
    const jwt = require('jsonwebtoken');
    
    const defaultOptions = {
      expiresIn: '24h',
      issuer: 'ollamamax',
      audience: 'ollamamax-users',
      algorithm: 'HS256'
    };

    return jwt.sign(payload, this.secrets.jwt, { ...defaultOptions, ...options });
  }

  /**
   * Verify JWT token
   */
  verifyJWT(token, options = {}) {
    const jwt = require('jsonwebtoken');
    
    const defaultOptions = {
      issuer: 'ollamamax',
      audience: 'ollamamax-users',
      algorithms: ['HS256']
    };

    return jwt.verify(token, this.secrets.jwt, { ...defaultOptions, ...options });
  }
}

// Create singleton instance
const securityConfig = new SecurityConfig();

export default securityConfig;
export { SecurityConfig };
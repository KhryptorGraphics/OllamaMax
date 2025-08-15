/**
 * Security Service
 * 
 * Comprehensive security features including CSP, XSS protection, CSRF tokens,
 * input validation, and security monitoring.
 */

class SecurityService {
  constructor() {
    this.csrfToken = null;
    this.securityHeaders = new Map();
    this.trustedDomains = new Set();
    this.securityViolations = [];
    this.init();
  }

  // Initialize security service
  init() {
    this.setupCSP();
    this.setupXSSProtection();
    this.setupCSRFProtection();
    this.setupSecurityHeaders();
    this.setupSecurityMonitoring();
    this.validateEnvironment();
  }

  // Content Security Policy
  setupCSP() {
    const cspDirectives = {
      'default-src': ["'self'"],
      'script-src': [
        "'self'",
        "'unsafe-inline'", // Only for development
        'https://cdn.jsdelivr.net',
        'https://unpkg.com',
      ],
      'style-src': [
        "'self'",
        "'unsafe-inline'",
        'https://fonts.googleapis.com',
      ],
      'font-src': [
        "'self'",
        'https://fonts.gstatic.com',
      ],
      'img-src': [
        "'self'",
        'data:',
        'https:',
      ],
      'connect-src': [
        "'self'",
        'wss:',
        'https://api.ollamamax.com',
        ...(process.env.NODE_ENV === 'development' ? ['http://localhost:8080', 'ws://localhost:8080'] : []),
      ],
      'frame-ancestors': ["'none'"],
      'base-uri': ["'self'"],
      'form-action': ["'self'"],
      'upgrade-insecure-requests': [],
    };

    // Build CSP string
    const cspString = Object.entries(cspDirectives)
      .map(([directive, sources]) => `${directive} ${sources.join(' ')}`)
      .join('; ');

    // Set CSP header (if running in development with a proxy)
    if (process.env.NODE_ENV === 'development') {
      console.log('CSP Policy:', cspString);
    }

    // Monitor CSP violations
    document.addEventListener('securitypolicyviolation', (event) => {
      this.handleSecurityViolation('csp', {
        blockedURI: event.blockedURI,
        violatedDirective: event.violatedDirective,
        originalPolicy: event.originalPolicy,
        sourceFile: event.sourceFile,
        lineNumber: event.lineNumber,
      });
    });
  }

  // XSS Protection
  setupXSSProtection() {
    // Input sanitization
    this.sanitizeInput = (input, options = {}) => {
      if (typeof input !== 'string') return input;

      const {
        allowHTML = false,
        allowedTags = [],
        maxLength = 10000,
      } = options;

      // Length check
      if (input.length > maxLength) {
        throw new Error(`Input exceeds maximum length of ${maxLength} characters`);
      }

      if (!allowHTML) {
        // Basic XSS prevention
        return input
          .replace(/&/g, '&amp;')
          .replace(/</g, '&lt;')
          .replace(/>/g, '&gt;')
          .replace(/"/g, '&quot;')
          .replace(/'/g, '&#x27;')
          .replace(/\//g, '&#x2F;');
      }

      // Advanced HTML sanitization (simplified)
      if (allowedTags.length > 0) {
        const tagRegex = new RegExp(`<(?!\/?(?:${allowedTags.join('|')})\s*\/?>)[^>]+>`, 'gi');
        return input.replace(tagRegex, '');
      }

      return input;
    };

    // URL validation
    this.validateURL = (url) => {
      try {
        const urlObj = new URL(url);
        
        // Check protocol
        if (!['http:', 'https:', 'mailto:', 'tel:'].includes(urlObj.protocol)) {
          return false;
        }

        // Check for suspicious patterns
        const suspiciousPatterns = [
          /javascript:/i,
          /data:/i,
          /vbscript:/i,
          /onload=/i,
          /onerror=/i,
        ];

        return !suspiciousPatterns.some(pattern => pattern.test(url));
      } catch {
        return false;
      }
    };

    // DOM manipulation safety
    this.safeSetInnerHTML = (element, html) => {
      if (!element || typeof html !== 'string') return;

      // Create a temporary element to parse HTML
      const temp = document.createElement('div');
      temp.innerHTML = this.sanitizeInput(html, { allowHTML: true });

      // Remove script tags and event handlers
      const scripts = temp.querySelectorAll('script');
      scripts.forEach(script => script.remove());

      const elements = temp.querySelectorAll('*');
      elements.forEach(el => {
        // Remove event handler attributes
        Array.from(el.attributes).forEach(attr => {
          if (attr.name.startsWith('on')) {
            el.removeAttribute(attr.name);
          }
        });

        // Validate href attributes
        if (el.hasAttribute('href')) {
          const href = el.getAttribute('href');
          if (!this.validateURL(href)) {
            el.removeAttribute('href');
          }
        }
      });

      element.innerHTML = temp.innerHTML;
    };
  }

  // CSRF Protection
  setupCSRFProtection() {
    // Generate CSRF token
    this.generateCSRFToken = () => {
      const array = new Uint8Array(32);
      crypto.getRandomValues(array);
      return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('');
    };

    // Get CSRF token from server or generate one
    this.getCSRFToken = async () => {
      if (this.csrfToken) return this.csrfToken;

      try {
        const response = await fetch('/api/v1/csrf-token', {
          method: 'GET',
          credentials: 'same-origin',
        });

        if (response.ok) {
          const data = await response.json();
          this.csrfToken = data.token;
        } else {
          // Fallback: generate client-side token
          this.csrfToken = this.generateCSRFToken();
        }
      } catch (error) {
        console.warn('Failed to get CSRF token from server:', error);
        this.csrfToken = this.generateCSRFToken();
      }

      return this.csrfToken;
    };

    // Add CSRF token to requests
    this.addCSRFToken = async (options = {}) => {
      const token = await this.getCSRFToken();
      
      return {
        ...options,
        headers: {
          'X-CSRF-Token': token,
          ...options.headers,
        },
      };
    };
  }

  // Security Headers
  setupSecurityHeaders() {
    this.securityHeaders.set('X-Content-Type-Options', 'nosniff');
    this.securityHeaders.set('X-Frame-Options', 'DENY');
    this.securityHeaders.set('X-XSS-Protection', '1; mode=block');
    this.securityHeaders.set('Referrer-Policy', 'strict-origin-when-cross-origin');
    this.securityHeaders.set('Permissions-Policy', 'geolocation=(), microphone=(), camera=()');

    // Add headers to fetch requests
    this.addSecurityHeaders = (options = {}) => {
      const headers = {};
      this.securityHeaders.forEach((value, key) => {
        headers[key] = value;
      });

      return {
        ...options,
        headers: {
          ...headers,
          ...options.headers,
        },
      };
    };
  }

  // Security Monitoring
  setupSecurityMonitoring() {
    // Monitor for suspicious activity
    this.monitorSuspiciousActivity = () => {
      let requestCount = 0;
      let lastRequestTime = Date.now();

      return {
        checkRateLimit: () => {
          const now = Date.now();
          const timeDiff = now - lastRequestTime;

          if (timeDiff < 100) { // Less than 100ms between requests
            requestCount++;
            if (requestCount > 10) {
              this.handleSecurityViolation('rate-limit', {
                requestCount,
                timeWindow: timeDiff,
              });
              return false;
            }
          } else {
            requestCount = 0;
          }

          lastRequestTime = now;
          return true;
        },
      };
    };

    this.activityMonitor = this.monitorSuspiciousActivity();

    // Monitor for DOM manipulation attempts
    if (typeof MutationObserver !== 'undefined') {
      const observer = new MutationObserver((mutations) => {
        mutations.forEach((mutation) => {
          if (mutation.type === 'childList') {
            mutation.addedNodes.forEach((node) => {
              if (node.nodeType === Node.ELEMENT_NODE) {
                // Check for suspicious script injections
                if (node.tagName === 'SCRIPT' && !node.hasAttribute('data-allowed')) {
                  this.handleSecurityViolation('script-injection', {
                    element: node.outerHTML,
                    source: mutation.target,
                  });
                  node.remove();
                }
              }
            });
          }
        });
      });

      observer.observe(document.body, {
        childList: true,
        subtree: true,
      });
    }
  }

  // Environment Validation
  validateEnvironment() {
    const checks = {
      https: location.protocol === 'https:' || location.hostname === 'localhost',
      secureContext: window.isSecureContext,
      cookieSecure: document.cookie.includes('Secure') || location.hostname === 'localhost',
      noConsoleErrors: !window.console.error.toString().includes('native'),
    };

    const failedChecks = Object.entries(checks)
      .filter(([, passed]) => !passed)
      .map(([check]) => check);

    if (failedChecks.length > 0) {
      console.warn('Security environment checks failed:', failedChecks);
      
      if (process.env.NODE_ENV === 'production') {
        this.handleSecurityViolation('environment', {
          failedChecks,
          userAgent: navigator.userAgent,
          location: location.href,
        });
      }
    }
  }

  // Handle security violations
  handleSecurityViolation(type, details) {
    const violation = {
      type,
      details,
      timestamp: new Date().toISOString(),
      userAgent: navigator.userAgent,
      url: location.href,
    };

    this.securityViolations.push(violation);

    // Log to console in development
    if (process.env.NODE_ENV === 'development') {
      console.warn('Security violation detected:', violation);
    }

    // Report to security monitoring service
    this.reportSecurityViolation(violation);

    // Emit event for application handling
    window.dispatchEvent(new CustomEvent('security-violation', {
      detail: violation,
    }));
  }

  // Report security violation
  async reportSecurityViolation(violation) {
    try {
      await fetch('/api/v1/security/violations', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(violation),
        credentials: 'same-origin',
      });
    } catch (error) {
      console.error('Failed to report security violation:', error);
    }
  }

  // Secure fetch wrapper
  async secureFetch(url, options = {}) {
    // Rate limiting check
    if (!this.activityMonitor.checkRateLimit()) {
      throw new Error('Rate limit exceeded');
    }

    // Add security headers and CSRF token
    const secureOptions = await this.addCSRFToken(
      this.addSecurityHeaders(options)
    );

    // Validate URL
    if (!this.validateURL(url)) {
      throw new Error('Invalid URL');
    }

    try {
      const response = await fetch(url, secureOptions);
      
      // Check for security headers in response
      this.validateResponseHeaders(response);
      
      return response;
    } catch (error) {
      this.handleSecurityViolation('fetch-error', {
        url,
        error: error.message,
      });
      throw error;
    }
  }

  // Validate response headers
  validateResponseHeaders(response) {
    const requiredHeaders = [
      'x-content-type-options',
      'x-frame-options',
      'x-xss-protection',
    ];

    const missingHeaders = requiredHeaders.filter(header => 
      !response.headers.has(header)
    );

    if (missingHeaders.length > 0) {
      this.handleSecurityViolation('missing-security-headers', {
        missingHeaders,
        url: response.url,
      });
    }
  }

  // Input validation schemas
  createValidator(schema) {
    return (input) => {
      const errors = [];

      Object.entries(schema).forEach(([field, rules]) => {
        const value = input[field];

        if (rules.required && (value === undefined || value === null || value === '')) {
          errors.push(`${field} is required`);
          return;
        }

        if (value !== undefined && value !== null) {
          if (rules.type && typeof value !== rules.type) {
            errors.push(`${field} must be of type ${rules.type}`);
          }

          if (rules.minLength && value.length < rules.minLength) {
            errors.push(`${field} must be at least ${rules.minLength} characters`);
          }

          if (rules.maxLength && value.length > rules.maxLength) {
            errors.push(`${field} must be no more than ${rules.maxLength} characters`);
          }

          if (rules.pattern && !rules.pattern.test(value)) {
            errors.push(`${field} format is invalid`);
          }

          if (rules.custom && !rules.custom(value)) {
            errors.push(`${field} validation failed`);
          }
        }
      });

      return {
        valid: errors.length === 0,
        errors,
      };
    };
  }

  // Common validation patterns
  getValidationPatterns() {
    return {
      email: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
      password: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$/,
      url: /^https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)$/,
      uuid: /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i,
      ipAddress: /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/,
    };
  }

  // Get security summary
  getSecuritySummary() {
    return {
      violations: this.securityViolations.length,
      recentViolations: this.securityViolations.slice(-10),
      csrfTokenSet: !!this.csrfToken,
      secureContext: window.isSecureContext,
      httpsEnabled: location.protocol === 'https:',
      headersConfigured: this.securityHeaders.size,
    };
  }

  // Cleanup
  destroy() {
    this.securityViolations = [];
    this.securityHeaders.clear();
    this.trustedDomains.clear();
    this.csrfToken = null;
  }
}

// Create singleton instance
const securityService = new SecurityService();

export default securityService;

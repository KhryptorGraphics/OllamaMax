# Security Implementation Guide

## Overview

This comprehensive guide provides developers, security engineers, and operations teams with practical instructions for implementing, maintaining, and improving the security posture of the Ollama Distributed frontend application.

## Table of Contents

1. [Security Architecture Overview](#security-architecture-overview)
2. [Authentication and Authorization](#authentication-and-authorization)
3. [Input Validation and Output Encoding](#input-validation-and-output-encoding)
4. [Content Security Policy (CSP)](#content-security-policy-csp)
5. [Security Headers Implementation](#security-headers-implementation)
6. [Cryptographic Controls](#cryptographic-controls)
7. [Session Management](#session-management)
8. [API Security](#api-security)
9. [Error Handling and Logging](#error-handling-and-logging)
10. [Security Testing](#security-testing)
11. [Deployment Security](#deployment-security)
12. [Monitoring and Incident Response](#monitoring-and-incident-response)

## Security Architecture Overview

### Defense in Depth Strategy

The Ollama Distributed frontend implements a layered security approach:

```
┌─────────────────────────────────────────────────────────────────┐
│                        User Interface Layer                     │
├─────────────────────────────────────────────────────────────────┤
│                     Application Security Layer                  │
│  • Input Validation  • Output Encoding  • XSS Protection      │
├─────────────────────────────────────────────────────────────────┤
│                     Authentication Layer                        │
│  • JWT Tokens  • MFA  • Session Management  • RBAC           │
├─────────────────────────────────────────────────────────────────┤
│                     Transport Security Layer                    │
│  • TLS 1.3  • HSTS  • Certificate Pinning                    │
├─────────────────────────────────────────────────────────────────┤
│                     Infrastructure Layer                        │
│  • WAF  • Rate Limiting  • DDoS Protection                    │
└─────────────────────────────────────────────────────────────────┘
```

### Security Principles

1. **Zero Trust Architecture**: Never trust, always verify
2. **Principle of Least Privilege**: Minimum necessary access
3. **Defense in Depth**: Multiple layers of security controls
4. **Fail Securely**: Secure failure modes and graceful degradation
5. **Security by Design**: Built-in security from the start

## Authentication and Authorization

### JWT Token Implementation

#### Token Structure
```typescript
// JWT token configuration
interface JWTConfig {
  algorithm: 'RS256' | 'ES256'
  expiresIn: string // '15m' for access, '7d' for refresh
  issuer: string
  audience: string
  keyId: string
}

// Token payload structure
interface JWTPayload {
  sub: string // User ID
  iat: number // Issued at
  exp: number // Expiration
  aud: string // Audience
  iss: string // Issuer
  roles: string[] // User roles
  permissions: string[] // Specific permissions
  sessionId: string // Session identifier
}
```

#### Implementation Example
```typescript
// src/utils/auth/jwt.ts
import jwt from 'jsonwebtoken'
import { readFileSync } from 'fs'

export class JWTService {
  private privateKey: string
  private publicKey: string

  constructor() {
    this.privateKey = readFileSync(process.env.JWT_PRIVATE_KEY_PATH!)
    this.publicKey = readFileSync(process.env.JWT_PUBLIC_KEY_PATH!)
  }

  generateTokens(user: User): TokenPair {
    const payload = {
      sub: user.id,
      roles: user.roles,
      permissions: user.permissions,
      sessionId: generateSessionId()
    }

    const accessToken = jwt.sign(payload, this.privateKey, {
      algorithm: 'RS256',
      expiresIn: '15m',
      issuer: process.env.JWT_ISSUER,
      audience: process.env.JWT_AUDIENCE,
      keyid: process.env.JWT_KEY_ID
    })

    const refreshToken = jwt.sign(
      { sub: user.id, type: 'refresh' },
      this.privateKey,
      {
        algorithm: 'RS256',
        expiresIn: '7d',
        issuer: process.env.JWT_ISSUER
      }
    )

    return { accessToken, refreshToken }
  }

  verifyToken(token: string): JWTPayload | null {
    try {
      return jwt.verify(token, this.publicKey, {
        algorithms: ['RS256'],
        issuer: process.env.JWT_ISSUER,
        audience: process.env.JWT_AUDIENCE
      }) as JWTPayload
    } catch (error) {
      logger.warn('Token verification failed', { error: error.message })
      return null
    }
  }
}
```

### Role-Based Access Control (RBAC)

#### Role Definition
```typescript
// src/types/auth.ts
interface Role {
  id: string
  name: string
  permissions: Permission[]
  description: string
  isActive: boolean
}

interface Permission {
  id: string
  resource: string
  action: string // 'read', 'write', 'delete', 'admin'
  conditions?: Record<string, any>
}

// Predefined roles
export const ROLES = {
  ADMIN: 'admin',
  USER: 'user',
  MODERATOR: 'moderator',
  READONLY: 'readonly'
} as const
```

#### Authorization Hook
```typescript
// src/hooks/useAuthorization.ts
import { useAuth } from './useAuth'

export function useAuthorization() {
  const { user } = useAuth()

  const hasPermission = (resource: string, action: string): boolean => {
    if (!user) return false

    return user.permissions.some(permission =>
      permission.resource === resource && 
      permission.action === action
    )
  }

  const hasRole = (role: string): boolean => {
    if (!user) return false
    return user.roles.includes(role)
  }

  const canAccess = (requiredPermissions: string[]): boolean => {
    return requiredPermissions.every(permission => {
      const [resource, action] = permission.split(':')
      return hasPermission(resource, action)
    })
  }

  return { hasPermission, hasRole, canAccess }
}
```

### Multi-Factor Authentication (MFA)

#### TOTP Implementation
```typescript
// src/utils/auth/mfa.ts
import speakeasy from 'speakeasy'
import QRCode from 'qrcode'

export class MFAService {
  generateSecret(userEmail: string): MFASecret {
    const secret = speakeasy.generateSecret({
      name: `Ollama Distributed (${userEmail})`,
      issuer: 'Ollama Distributed',
      length: 32
    })

    return {
      secret: secret.base32,
      qrCode: secret.otpauth_url,
      backupCodes: this.generateBackupCodes()
    }
  }

  async generateQRCode(secret: string): Promise<string> {
    return QRCode.toDataURL(secret)
  }

  verifyToken(token: string, secret: string): boolean {
    return speakeasy.totp.verify({
      secret,
      encoding: 'base32',
      token,
      window: 2 // Allow 2 time steps of drift
    })
  }

  private generateBackupCodes(): string[] {
    return Array.from({ length: 8 }, () =>
      Math.random().toString(36).substring(2, 10).toUpperCase()
    )
  }
}
```

## Input Validation and Output Encoding

### Input Validation Framework

#### Validation Schema
```typescript
// src/utils/validation/schemas.ts
import Joi from 'joi'
import DOMPurify from 'dompurify'

export const userInputSchema = Joi.object({
  email: Joi.string().email().max(255).required(),
  name: Joi.string().min(2).max(100).pattern(/^[a-zA-Z\s]+$/).required(),
  message: Joi.string().max(5000).required(),
  url: Joi.string().uri({ scheme: ['http', 'https'] }).optional()
})

export const searchSchema = Joi.object({
  query: Joi.string().min(1).max(255).pattern(/^[a-zA-Z0-9\s\-_]+$/).required(),
  filter: Joi.string().valid('all', 'models', 'datasets').default('all'),
  limit: Joi.number().integer().min(1).max(100).default(20),
  offset: Joi.number().integer().min(0).default(0)
})
```

#### Input Sanitization
```typescript
// src/utils/validation/sanitizer.ts
export class InputSanitizer {
  static sanitizeHtml(input: string): string {
    return DOMPurify.sanitize(input, {
      ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'p', 'br'],
      ALLOWED_ATTR: [],
      ALLOW_DATA_ATTR: false,
      FORBID_SCRIPT: true,
      FORBID_TAGS: ['script', 'object', 'embed', 'form', 'input']
    })
  }

  static sanitizeText(input: string): string {
    return input
      .replace(/[<>]/g, '') // Remove angle brackets
      .replace(/javascript:/gi, '') // Remove javascript: protocol
      .replace(/on\w+\s*=/gi, '') // Remove event handlers
      .trim()
      .substring(0, 1000) // Limit length
  }

  static sanitizeFilename(filename: string): string {
    return filename
      .replace(/[^a-zA-Z0-9.\-_]/g, '')
      .replace(/\.+/g, '.')
      .substring(0, 255)
  }

  static sanitizeUrl(url: string): string | null {
    try {
      const parsed = new URL(url)
      
      // Only allow https and http protocols
      if (!['http:', 'https:'].includes(parsed.protocol)) {
        return null
      }
      
      // Block localhost and private IP ranges
      const hostname = parsed.hostname.toLowerCase()
      if (hostname === 'localhost' || 
          hostname.startsWith('127.') ||
          hostname.startsWith('192.168.') ||
          hostname.startsWith('10.') ||
          /^172\.(1[6-9]|2\d|3[01])\./.test(hostname)) {
        return null
      }
      
      return parsed.toString()
    } catch {
      return null
    }
  }
}
```

### Output Encoding

#### HTML Context Encoding
```typescript
// src/utils/encoding/htmlEncoder.ts
export class HTMLEncoder {
  static encode(text: string): string {
    return text
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#x27;')
      .replace(/\//g, '&#x2F;')
  }

  static encodeAttribute(text: string): string {
    return text
      .replace(/&/g, '&amp;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#x27;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
  }

  static encodeJS(text: string): string {
    return text
      .replace(/\\/g, '\\\\')
      .replace(/'/g, "\\'")
      .replace(/"/g, '\\"')
      .replace(/\n/g, '\\n')
      .replace(/\r/g, '\\r')
      .replace(/\t/g, '\\t')
  }
}
```

#### React Safe Rendering
```typescript
// src/components/SafeContent.tsx
import { HTMLEncoder } from '../utils/encoding/htmlEncoder'

interface SafeContentProps {
  content: string
  allowHtml?: boolean
}

export function SafeContent({ content, allowHtml = false }: SafeContentProps) {
  if (allowHtml) {
    // Use DOMPurify for controlled HTML rendering
    const cleanHtml = DOMPurify.sanitize(content, {
      ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'p', 'br', 'ul', 'ol', 'li'],
      ALLOWED_ATTR: ['href', 'target'],
      ALLOW_DATA_ATTR: false
    })
    
    return <div dangerouslySetInnerHTML={{ __html: cleanHtml }} />
  }

  // Always encode plain text
  return <span>{HTMLEncoder.encode(content)}</span>
}
```

## Content Security Policy (CSP)

### CSP Configuration

#### Strict CSP Implementation
```typescript
// src/middleware/security/csp.ts
export function generateCSPPolicy(nonce: string): string {
  const policy = {
    'default-src': ["'self'"],
    'script-src': [
      "'self'",
      `'nonce-${nonce}'`,
      "'strict-dynamic'",
      // Remove 'unsafe-inline' for strict CSP
    ],
    'style-src': [
      "'self'",
      `'nonce-${nonce}'`,
      // Only allow specific trusted style sources
    ],
    'img-src': [
      "'self'",
      'data:',
      'https:',
      // Add specific image CDN domains
    ],
    'connect-src': [
      "'self'",
      'wss://ollama-api.example.com',
      'https://api.ollama-distributed.com'
    ],
    'font-src': [
      "'self'",
      'https://fonts.googleapis.com',
      'https://fonts.gstatic.com'
    ],
    'object-src': ["'none'"],
    'media-src': ["'self'"],
    'frame-src': ["'none'"],
    'frame-ancestors': ["'none'"],
    'base-uri': ["'self'"],
    'form-action': ["'self'"],
    'upgrade-insecure-requests': []
  }

  return Object.entries(policy)
    .map(([directive, sources]) => 
      sources.length > 0 
        ? `${directive} ${sources.join(' ')}`
        : directive
    )
    .join('; ')
}
```

#### Nonce Generation
```typescript
// src/utils/security/nonce.ts
import crypto from 'crypto'

export class NonceGenerator {
  static generate(): string {
    return crypto.randomBytes(16).toString('base64')
  }

  static middleware(req: Request, res: Response, next: NextFunction) {
    const nonce = NonceGenerator.generate()
    res.locals.nonce = nonce
    res.setHeader('Content-Security-Policy', generateCSPPolicy(nonce))
    next()
  }
}
```

#### CSP Violation Reporting
```typescript
// src/middleware/security/cspReporting.ts
export function cspViolationHandler(req: Request, res: Response) {
  const violation = req.body['csp-report']
  
  if (violation) {
    logger.warn('CSP Violation Detected', {
      'blocked-uri': violation['blocked-uri'],
      'document-uri': violation['document-uri'],
      'violated-directive': violation['violated-directive'],
      'original-policy': violation['original-policy'],
      'user-agent': req.headers['user-agent'],
      timestamp: new Date().toISOString()
    })

    // Alert security team for critical violations
    if (isCriticalViolation(violation)) {
      alertSecurityTeam(violation)
    }
  }

  res.status(204).send()
}

function isCriticalViolation(violation: CSPViolation): boolean {
  const criticalDirectives = [
    'script-src',
    'object-src',
    'base-uri',
    'form-action'
  ]
  
  return criticalDirectives.some(directive =>
    violation['violated-directive'].startsWith(directive)
  )
}
```

## Security Headers Implementation

### Comprehensive Security Headers

```typescript
// src/middleware/security/headers.ts
export function securityHeadersMiddleware(req: Request, res: Response, next: NextFunction) {
  // Strict Transport Security
  res.setHeader(
    'Strict-Transport-Security',
    'max-age=31536000; includeSubDomains; preload'
  )

  // Frame Options
  res.setHeader('X-Frame-Options', 'DENY')

  // Content Type Options
  res.setHeader('X-Content-Type-Options', 'nosniff')

  // XSS Protection
  res.setHeader('X-XSS-Protection', '1; mode=block')

  // Referrer Policy
  res.setHeader('Referrer-Policy', 'strict-origin-when-cross-origin')

  // Permissions Policy
  res.setHeader(
    'Permissions-Policy',
    'camera=(), microphone=(), geolocation=(), payment=(), usb=(), magnetometer=(), gyroscope=(), accelerometer=()'
  )

  // Cross-Origin Policies
  res.setHeader('Cross-Origin-Opener-Policy', 'same-origin')
  res.setHeader('Cross-Origin-Embedder-Policy', 'require-corp')
  res.setHeader('Cross-Origin-Resource-Policy', 'same-origin')

  // Remove server information
  res.removeHeader('Server')
  res.removeHeader('X-Powered-By')

  next()
}
```

### CORS Configuration

```typescript
// src/middleware/security/cors.ts
import cors from 'cors'

const corsOptions: cors.CorsOptions = {
  origin: function (origin, callback) {
    const allowedOrigins = [
      'https://ollama-distributed.com',
      'https://app.ollama-distributed.com',
      ...(process.env.NODE_ENV === 'development' ? ['http://localhost:3000'] : [])
    ]

    if (!origin || allowedOrigins.includes(origin)) {
      callback(null, true)
    } else {
      callback(new Error('Not allowed by CORS'))
    }
  },
  credentials: true,
  optionsSuccessStatus: 200,
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
  allowedHeaders: [
    'Content-Type',
    'Authorization',
    'X-Requested-With',
    'X-CSRF-Token'
  ],
  exposedHeaders: ['X-RateLimit-Remaining', 'X-RateLimit-Reset'],
  maxAge: 86400 // 24 hours
}

export const corsMiddleware = cors(corsOptions)
```

## Cryptographic Controls

### Encryption Implementation

#### AES Encryption
```typescript
// src/utils/crypto/encryption.ts
import crypto from 'crypto'

export class EncryptionService {
  private algorithm = 'aes-256-gcm'
  private keyLength = 32
  private ivLength = 16
  private tagLength = 16

  encrypt(plaintext: string, key: Buffer): EncryptedData {
    const iv = crypto.randomBytes(this.ivLength)
    const cipher = crypto.createCipher(this.algorithm, key, { iv })
    
    let ciphertext = cipher.update(plaintext, 'utf8')
    ciphertext = Buffer.concat([ciphertext, cipher.final()])
    
    const tag = cipher.getAuthTag()

    return {
      ciphertext: ciphertext.toString('base64'),
      iv: iv.toString('base64'),
      tag: tag.toString('base64')
    }
  }

  decrypt(encryptedData: EncryptedData, key: Buffer): string {
    const iv = Buffer.from(encryptedData.iv, 'base64')
    const tag = Buffer.from(encryptedData.tag, 'base64')
    const ciphertext = Buffer.from(encryptedData.ciphertext, 'base64')

    const decipher = crypto.createDecipher(this.algorithm, key, { iv })
    decipher.setAuthTag(tag)

    let plaintext = decipher.update(ciphertext)
    plaintext = Buffer.concat([plaintext, decipher.final()])

    return plaintext.toString('utf8')
  }

  generateKey(): Buffer {
    return crypto.randomBytes(this.keyLength)
  }
}
```

#### Password Hashing
```typescript
// src/utils/crypto/password.ts
import bcrypt from 'bcrypt'
import argon2 from 'argon2'

export class PasswordService {
  private readonly saltRounds = 12
  
  // Use bcrypt for backward compatibility
  async hashPasswordBcrypt(password: string): Promise<string> {
    return bcrypt.hash(password, this.saltRounds)
  }

  async verifyPasswordBcrypt(password: string, hash: string): Promise<boolean> {
    return bcrypt.compare(password, hash)
  }

  // Use Argon2 for new passwords (recommended)
  async hashPasswordArgon2(password: string): Promise<string> {
    return argon2.hash(password, {
      type: argon2.argon2id,
      memoryCost: 2 ** 16, // 64 MB
      timeCost: 3,
      parallelism: 1,
    })
  }

  async verifyPasswordArgon2(password: string, hash: string): Promise<boolean> {
    try {
      return await argon2.verify(hash, password)
    } catch (error) {
      return false
    }
  }

  validatePasswordStrength(password: string): PasswordValidation {
    const result: PasswordValidation = {
      isValid: false,
      score: 0,
      feedback: []
    }

    // Minimum length check
    if (password.length < 12) {
      result.feedback.push('Password must be at least 12 characters long')
    } else {
      result.score += 2
    }

    // Character variety checks
    if (!/[a-z]/.test(password)) {
      result.feedback.push('Password must contain lowercase letters')
    } else {
      result.score += 1
    }

    if (!/[A-Z]/.test(password)) {
      result.feedback.push('Password must contain uppercase letters')
    } else {
      result.score += 1
    }

    if (!/\d/.test(password)) {
      result.feedback.push('Password must contain numbers')
    } else {
      result.score += 1
    }

    if (!/[!@#$%^&*(),.?":{}|<>]/.test(password)) {
      result.feedback.push('Password must contain special characters')
    } else {
      result.score += 1
    }

    // Common password check
    if (this.isCommonPassword(password)) {
      result.feedback.push('Password is too common')
      result.score = Math.max(0, result.score - 2)
    }

    result.isValid = result.score >= 5 && result.feedback.length === 0

    return result
  }

  private isCommonPassword(password: string): boolean {
    const commonPasswords = [
      'password', '123456', 'password123', 'admin', 'qwerty',
      'letmein', 'welcome', 'monkey', 'dragon', 'master'
    ]
    return commonPasswords.includes(password.toLowerCase())
  }
}
```

## Session Management

### Secure Session Implementation

```typescript
// src/utils/session/sessionManager.ts
export class SessionManager {
  private sessions = new Map<string, Session>()
  private readonly maxAge = 15 * 60 * 1000 // 15 minutes
  private readonly cleanupInterval = 5 * 60 * 1000 // 5 minutes

  constructor() {
    // Periodic cleanup of expired sessions
    setInterval(() => this.cleanup(), this.cleanupInterval)
  }

  createSession(userId: string, userAgent: string, ipAddress: string): string {
    const sessionId = this.generateSessionId()
    
    const session: Session = {
      id: sessionId,
      userId,
      userAgent,
      ipAddress,
      createdAt: Date.now(),
      lastActivity: Date.now(),
      isActive: true,
      csrfToken: this.generateCSRFToken()
    }

    this.sessions.set(sessionId, session)
    return sessionId
  }

  validateSession(sessionId: string, userAgent: string, ipAddress: string): Session | null {
    const session = this.sessions.get(sessionId)

    if (!session) {
      return null
    }

    // Check if session is expired
    if (Date.now() - session.lastActivity > this.maxAge) {
      this.destroySession(sessionId)
      return null
    }

    // Check if session details match (prevent session hijacking)
    if (session.userAgent !== userAgent || session.ipAddress !== ipAddress) {
      logger.warn('Session hijacking attempt detected', {
        sessionId,
        expectedUserAgent: session.userAgent,
        actualUserAgent: userAgent,
        expectedIP: session.ipAddress,
        actualIP: ipAddress
      })
      this.destroySession(sessionId)
      return null
    }

    // Update last activity
    session.lastActivity = Date.now()
    return session
  }

  destroySession(sessionId: string): void {
    this.sessions.delete(sessionId)
  }

  destroyAllUserSessions(userId: string): void {
    Array.from(this.sessions.entries()).forEach(([id, session]) => {
      if (session.userId === userId) {
        this.sessions.delete(id)
      }
    })
  }

  private generateSessionId(): string {
    return crypto.randomBytes(32).toString('hex')
  }

  private generateCSRFToken(): string {
    return crypto.randomBytes(32).toString('base64')
  }

  private cleanup(): void {
    const now = Date.now()
    const expiredSessions = Array.from(this.sessions.entries())
      .filter(([_, session]) => now - session.lastActivity > this.maxAge)
      .map(([id]) => id)

    expiredSessions.forEach(id => this.sessions.delete(id))

    if (expiredSessions.length > 0) {
      logger.info(`Cleaned up ${expiredSessions.length} expired sessions`)
    }
  }
}
```

### CSRF Protection

```typescript
// src/middleware/security/csrf.ts
export function csrfProtection(req: Request, res: Response, next: NextFunction) {
  const session = req.session
  const token = req.headers['x-csrf-token'] as string

  if (!session) {
    return res.status(401).json({ error: 'No valid session' })
  }

  if (req.method !== 'GET' && req.method !== 'HEAD' && req.method !== 'OPTIONS') {
    if (!token) {
      return res.status(403).json({ error: 'CSRF token required' })
    }

    if (token !== session.csrfToken) {
      logger.warn('CSRF token mismatch', {
        sessionId: session.id,
        expectedToken: session.csrfToken,
        providedToken: token,
        userAgent: req.headers['user-agent'],
        ip: req.ip
      })
      return res.status(403).json({ error: 'Invalid CSRF token' })
    }
  }

  // Provide CSRF token to client
  res.locals.csrfToken = session.csrfToken
  next()
}
```

## API Security

### Rate Limiting

```typescript
// src/middleware/security/rateLimit.ts
import { RateLimiterMemory } from 'rate-limiter-flexible'

export class RateLimitService {
  private limiters = new Map<string, RateLimiterMemory>()

  constructor() {
    // Different rate limits for different endpoints
    this.limiters.set('auth', new RateLimiterMemory({
      points: 5, // 5 attempts
      duration: 300, // per 5 minutes
      blockDuration: 900, // block for 15 minutes
    }))

    this.limiters.set('api', new RateLimiterMemory({
      points: 100, // 100 requests
      duration: 60, // per 1 minute
      blockDuration: 60, // block for 1 minute
    }))

    this.limiters.set('upload', new RateLimiterMemory({
      points: 5, // 5 uploads
      duration: 3600, // per 1 hour
      blockDuration: 3600, // block for 1 hour
    }))
  }

  async checkRateLimit(category: string, identifier: string): Promise<boolean> {
    const limiter = this.limiters.get(category)
    if (!limiter) return true

    try {
      await limiter.consume(identifier)
      return true
    } catch (rejRes) {
      return false
    }
  }

  middleware(category: string) {
    return async (req: Request, res: Response, next: NextFunction) => {
      const identifier = req.ip + ':' + (req.user?.id || 'anonymous')
      
      const allowed = await this.checkRateLimit(category, identifier)
      
      if (!allowed) {
        logger.warn('Rate limit exceeded', {
          category,
          identifier,
          endpoint: req.path,
          userAgent: req.headers['user-agent']
        })
        
        return res.status(429).json({
          error: 'Rate limit exceeded',
          retryAfter: 60
        })
      }
      
      next()
    }
  }
}
```

### API Input Validation

```typescript
// src/middleware/validation/apiValidation.ts
export function validateApiInput(schema: Joi.Schema) {
  return (req: Request, res: Response, next: NextFunction) => {
    const { error, value } = schema.validate({
      ...req.body,
      ...req.query,
      ...req.params
    }, {
      abortEarly: false,
      stripUnknown: true,
      convert: true
    })

    if (error) {
      const errors = error.details.map(detail => ({
        field: detail.path.join('.'),
        message: detail.message,
        value: detail.context?.value
      }))

      logger.warn('API input validation failed', {
        endpoint: req.path,
        method: req.method,
        errors,
        ip: req.ip
      })

      return res.status(400).json({
        error: 'Input validation failed',
        details: errors
      })
    }

    // Replace request data with validated/sanitized data
    req.body = value
    next()
  }
}
```

## Error Handling and Logging

### Secure Error Handling

```typescript
// src/middleware/error/errorHandler.ts
export function errorHandler(error: Error, req: Request, res: Response, next: NextFunction) {
  const errorId = crypto.randomUUID()
  
  // Log detailed error information
  logger.error('Request error', {
    errorId,
    message: error.message,
    stack: error.stack,
    url: req.url,
    method: req.method,
    ip: req.ip,
    userAgent: req.headers['user-agent'],
    userId: req.user?.id,
    timestamp: new Date().toISOString()
  })

  // Determine error response based on environment
  const isProduction = process.env.NODE_ENV === 'production'
  
  let statusCode = 500
  let message = 'Internal server error'
  
  if (error instanceof ValidationError) {
    statusCode = 400
    message = 'Invalid input data'
  } else if (error instanceof AuthenticationError) {
    statusCode = 401
    message = 'Authentication required'
  } else if (error instanceof AuthorizationError) {
    statusCode = 403
    message = 'Insufficient permissions'
  } else if (error instanceof NotFoundError) {
    statusCode = 404
    message = 'Resource not found'
  }

  const errorResponse: any = {
    error: message,
    errorId,
    timestamp: new Date().toISOString()
  }

  // Only include detailed error information in development
  if (!isProduction && !(error instanceof AuthenticationError)) {
    errorResponse.details = error.message
    errorResponse.stack = error.stack
  }

  res.status(statusCode).json(errorResponse)
}
```

### Security Logging

```typescript
// src/utils/logging/securityLogger.ts
export class SecurityLogger {
  private static instance: SecurityLogger
  private logger: winston.Logger

  private constructor() {
    this.logger = winston.createLogger({
      level: 'info',
      format: winston.format.combine(
        winston.format.timestamp(),
        winston.format.errors({ stack: true }),
        winston.format.json()
      ),
      transports: [
        new winston.transports.File({ 
          filename: 'logs/security.log',
          level: 'warn'
        }),
        new winston.transports.File({ 
          filename: 'logs/security-audit.log',
          level: 'info'
        })
      ]
    })
  }

  static getInstance(): SecurityLogger {
    if (!SecurityLogger.instance) {
      SecurityLogger.instance = new SecurityLogger()
    }
    return SecurityLogger.instance
  }

  logAuthenticationEvent(event: AuthenticationEvent) {
    this.logger.info('Authentication Event', {
      type: 'AUTHENTICATION',
      ...event,
      sensitiveDataRemoved: true
    })
  }

  logAuthorizationFailure(event: AuthorizationEvent) {
    this.logger.warn('Authorization Failure', {
      type: 'AUTHORIZATION_FAILURE',
      ...event
    })
  }

  logSecurityViolation(event: SecurityViolationEvent) {
    this.logger.error('Security Violation', {
      type: 'SECURITY_VIOLATION',
      ...event,
      severity: 'HIGH'
    })
  }

  logDataAccess(event: DataAccessEvent) {
    this.logger.info('Data Access', {
      type: 'DATA_ACCESS',
      ...event
    })
  }

  logConfigurationChange(event: ConfigurationChangeEvent) {
    this.logger.warn('Configuration Change', {
      type: 'CONFIGURATION_CHANGE',
      ...event
    })
  }
}
```

## Security Testing

### Automated Security Tests

```typescript
// tests/security/security.test.ts
describe('Security Tests', () => {
  describe('XSS Prevention', () => {
    it('should prevent reflected XSS attacks', async () => {
      const xssPayloads = [
        '<script>alert("xss")</script>',
        '<img src=x onerror=alert("xss")>',
        'javascript:alert("xss")'
      ]

      for (const payload of xssPayloads) {
        const response = await request(app)
          .get(`/search?q=${encodeURIComponent(payload)}`)
          .expect(200)

        expect(response.text).not.toContain('<script')
        expect(response.text).not.toContain('javascript:')
        expect(response.text).not.toContain('onerror=')
      }
    })

    it('should prevent stored XSS attacks', async () => {
      const xssPayload = '<script>alert("stored-xss")</script>'
      
      await request(app)
        .post('/api/comments')
        .set('Authorization', `Bearer ${userToken}`)
        .send({ content: xssPayload })
        .expect(201)

      const response = await request(app)
        .get('/api/comments')
        .set('Authorization', `Bearer ${userToken}`)
        .expect(200)

      expect(response.body.comments[0].content).not.toContain('<script')
    })
  })

  describe('Authentication Security', () => {
    it('should enforce rate limiting on login attempts', async () => {
      const loginAttempts = Array.from({ length: 10 }, () =>
        request(app)
          .post('/api/auth/login')
          .send({ email: 'test@example.com', password: 'wrongpassword' })
      )

      const responses = await Promise.all(loginAttempts)
      const rateLimitedResponses = responses.filter(res => res.status === 429)
      
      expect(rateLimitedResponses.length).toBeGreaterThan(0)
    })

    it('should require CSRF token for state-changing operations', async () => {
      const response = await request(app)
        .post('/api/profile')
        .set('Authorization', `Bearer ${userToken}`)
        .send({ name: 'Updated Name' })
        .expect(403)

      expect(response.body.error).toContain('CSRF')
    })
  })

  describe('Input Validation', () => {
    it('should validate and sanitize user input', async () => {
      const maliciousInput = {
        name: '<script>alert("xss")</script>',
        email: 'not-an-email',
        age: 'not-a-number'
      }

      const response = await request(app)
        .post('/api/users')
        .set('Authorization', `Bearer ${adminToken}`)
        .send(maliciousInput)
        .expect(400)

      expect(response.body.error).toBe('Input validation failed')
      expect(response.body.details).toBeDefined()
    })
  })
})
```

### Manual Security Testing

#### Security Test Checklist

```markdown
## Manual Security Testing Checklist

### Authentication & Authorization
- [ ] Test password complexity requirements
- [ ] Verify account lockout after failed attempts
- [ ] Test session timeout functionality
- [ ] Verify MFA implementation
- [ ] Test role-based access controls
- [ ] Check for privilege escalation vulnerabilities

### Input Validation
- [ ] Test all input fields for XSS vulnerabilities
- [ ] Test for SQL injection vulnerabilities
- [ ] Test file upload functionality for malicious files
- [ ] Verify input length limits
- [ ] Test special character handling

### Session Management
- [ ] Verify secure session cookie settings
- [ ] Test session fixation protection
- [ ] Check for concurrent session handling
- [ ] Test session invalidation on logout

### Security Headers
- [ ] Verify CSP implementation
- [ ] Check HSTS configuration
- [ ] Verify X-Frame-Options header
- [ ] Test security headers consistency

### Error Handling
- [ ] Verify error messages don't leak sensitive information
- [ ] Test application behavior under error conditions
- [ ] Check for information disclosure in stack traces

### API Security
- [ ] Test API rate limiting
- [ ] Verify API authentication
- [ ] Test API input validation
- [ ] Check for IDOR vulnerabilities
```

## Deployment Security

### Production Security Configuration

```typescript
// config/production.ts
export const productionConfig = {
  security: {
    // Force HTTPS
    forceSSL: true,
    
    // Security headers
    headers: {
      hsts: {
        maxAge: 31536000,
        includeSubDomains: true,
        preload: true
      },
      csp: {
        defaultSrc: ["'self'"],
        scriptSrc: ["'self'", "'strict-dynamic'"],
        styleSrc: ["'self'"],
        imgSrc: ["'self'", "data:", "https:"],
        connectSrc: ["'self'"],
        fontSrc: ["'self'", "https://fonts.googleapis.com"],
        objectSrc: ["'none'"],
        mediaSrc: ["'self'"],
        frameSrc: ["'none'"],
        frameAncestors: ["'none'"],
        baseUri: ["'self'"],
        formAction: ["'self'"]
      }
    },
    
    // Session configuration
    session: {
      secure: true,
      httpOnly: true,
      sameSite: 'strict',
      maxAge: 15 * 60 * 1000 // 15 minutes
    },
    
    // Rate limiting
    rateLimit: {
      windowMs: 15 * 60 * 1000, // 15 minutes
      max: 100 // limit each IP to 100 requests per windowMs
    },
    
    // Logging
    logging: {
      level: 'warn',
      auditLog: true,
      sensitiveDataMasking: true
    }
  }
}
```

### Docker Security Configuration

```dockerfile
# Dockerfile.security
FROM node:18-alpine AS builder

# Create non-root user
RUN addgroup -g 1001 -S nodejs
RUN adduser -S nextjs -u 1001

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./
COPY tsconfig.json ./

# Install dependencies
RUN npm ci --only=production && npm cache clean --force

# Copy source code
COPY --chown=nextjs:nodejs . .

# Build application
RUN npm run build

# Production stage
FROM node:18-alpine AS runner

# Install security updates
RUN apk update && apk upgrade
RUN apk add --no-cache dumb-init

# Create non-root user
RUN addgroup -g 1001 -S nodejs
RUN adduser -S nextjs -u 1001

# Set working directory
WORKDIR /app

# Copy built application
COPY --from=builder --chown=nextjs:nodejs /app/dist ./dist
COPY --from=builder --chown=nextjs:nodejs /app/node_modules ./node_modules
COPY --from=builder --chown=nextjs:nodejs /app/package.json ./package.json

# Set security headers in nginx
COPY nginx.conf /etc/nginx/nginx.conf

# Switch to non-root user
USER nextjs

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:3000/health || exit 1

# Start application with dumb-init
ENTRYPOINT ["dumb-init", "--"]
CMD ["node", "dist/server.js"]
```

### Environment Variable Security

```typescript
// src/config/environment.ts
import Joi from 'joi'

const environmentSchema = Joi.object({
  NODE_ENV: Joi.string().valid('development', 'test', 'production').required(),
  
  // Database
  DB_HOST: Joi.string().hostname().required(),
  DB_PORT: Joi.number().port().default(5432),
  DB_NAME: Joi.string().min(1).required(),
  DB_USER: Joi.string().min(1).required(),
  DB_PASSWORD: Joi.string().min(8).required().sensitive(),
  
  // JWT
  JWT_PRIVATE_KEY_PATH: Joi.string().min(1).required(),
  JWT_PUBLIC_KEY_PATH: Joi.string().min(1).required(),
  JWT_ISSUER: Joi.string().uri().required(),
  JWT_AUDIENCE: Joi.string().uri().required(),
  JWT_KEY_ID: Joi.string().min(1).required(),
  
  // Encryption
  ENCRYPTION_KEY: Joi.string().length(64).hex().required().sensitive(),
  
  // API Keys
  EXTERNAL_API_KEY: Joi.string().min(32).required().sensitive(),
  
  // Security
  CSRF_SECRET: Joi.string().min(32).required().sensitive(),
  SESSION_SECRET: Joi.string().min(32).required().sensitive(),
  
  // Optional development overrides
  DISABLE_HTTPS: Joi.boolean().default(false).when('NODE_ENV', {
    is: 'production',
    then: Joi.forbidden()
  })
})

export function validateEnvironment() {
  const { error, value } = environmentSchema.validate(process.env, {
    abortEarly: false,
    stripUnknown: true
  })

  if (error) {
    console.error('Environment validation failed:')
    error.details.forEach(detail => {
      console.error(`- ${detail.message}`)
    })
    process.exit(1)
  }

  return value
}
```

## Monitoring and Incident Response

### Security Monitoring Dashboard

```typescript
// src/monitoring/securityDashboard.ts
export class SecurityDashboard {
  private metrics = {
    authenticationFailures: 0,
    cspViolations: 0,
    rateLimitExceeded: 0,
    suspiciousActivity: 0,
    dataAccessViolations: 0
  }

  private alerts: SecurityAlert[] = []

  updateMetrics(eventType: SecurityEventType, count: number = 1) {
    this.metrics[eventType] += count
    
    // Check alert thresholds
    this.checkAlertThresholds(eventType)
  }

  private checkAlertThresholds(eventType: SecurityEventType) {
    const thresholds = {
      authenticationFailures: 50, // per hour
      cspViolations: 10, // per hour
      rateLimitExceeded: 100, // per hour
      suspiciousActivity: 5, // per hour
      dataAccessViolations: 1 // immediate
    }

    if (this.metrics[eventType] >= thresholds[eventType]) {
      this.triggerAlert({
        type: eventType,
        severity: this.getAlertSeverity(eventType),
        message: `${eventType} threshold exceeded: ${this.metrics[eventType]}`,
        timestamp: new Date().toISOString()
      })
    }
  }

  private triggerAlert(alert: SecurityAlert) {
    this.alerts.push(alert)
    
    // Send notifications based on severity
    if (alert.severity === 'CRITICAL' || alert.severity === 'HIGH') {
      this.sendImmediateNotification(alert)
    }
    
    // Log to security monitoring system
    SecurityLogger.getInstance().logSecurityViolation({
      type: alert.type,
      severity: alert.severity,
      message: alert.message,
      timestamp: alert.timestamp
    })
  }

  private async sendImmediateNotification(alert: SecurityAlert) {
    // Send to security team via multiple channels
    await Promise.all([
      this.sendEmailAlert(alert),
      this.sendSlackAlert(alert),
      this.sendSMSAlert(alert) // For critical alerts only
    ])
  }
}
```

### Incident Response Automation

```typescript
// src/monitoring/incidentResponse.ts
export class IncidentResponseAutomation {
  private activeIncidents = new Map<string, SecurityIncident>()

  async handleSecurityEvent(event: SecurityEvent) {
    const severity = this.assessSeverity(event)
    const incidentId = this.generateIncidentId()

    const incident: SecurityIncident = {
      id: incidentId,
      type: event.type,
      severity,
      status: 'DETECTED',
      startTime: new Date(),
      events: [event],
      affectedResources: this.identifyAffectedResources(event),
      automatedActions: []
    }

    this.activeIncidents.set(incidentId, incident)

    // Execute automated response based on severity and type
    await this.executeAutomatedResponse(incident)

    // Notify incident response team
    await this.notifyIncidentTeam(incident)

    return incident
  }

  private async executeAutomatedResponse(incident: SecurityIncident) {
    const actions = []

    switch (incident.type) {
      case 'BRUTE_FORCE_ATTACK':
        actions.push(
          this.blockSuspiciousIPs(incident),
          this.increaseRateLimiting(),
          this.alertSecurityTeam()
        )
        break

      case 'XSS_ATTEMPT':
        actions.push(
          this.blockMaliciousRequests(incident),
          this.updateCSPPolicy(),
          this.scanForSimilarAttempts()
        )
        break

      case 'SQL_INJECTION_ATTEMPT':
        actions.push(
          this.blockAttackingIPs(incident),
          this.enableDatabaseQueryLogging(),
          this.scanDatabaseForCompromise()
        )
        break

      case 'DATA_EXFILTRATION_SUSPECTED':
        actions.push(
          this.suspendSuspiciousAccounts(incident),
          this.enableDataAccessLogging(),
          this.notifyDataProtectionOfficer()
        )
        break
    }

    incident.automatedActions = await Promise.all(actions)
    incident.status = 'MITIGATING'
  }

  private async blockSuspiciousIPs(incident: SecurityIncident): Promise<AutomatedAction> {
    const suspiciousIPs = this.extractIPsFromEvents(incident.events)
    
    // Add to firewall block list
    await this.firewallService.blockIPs(suspiciousIPs, '24h')
    
    return {
      type: 'IP_BLOCK',
      timestamp: new Date(),
      details: { blockedIPs: suspiciousIPs, duration: '24h' },
      success: true
    }
  }
}
```

This comprehensive security implementation guide provides the foundation for maintaining a robust security posture for the Ollama Distributed frontend application. Regular updates and security reviews ensure continued effectiveness against evolving threats.

---

**Document Version**: 1.0
**Last Updated**: $(date +'%Y-%m-%d %H:%M:%S')
**Classification**: Internal Use - Security Implementation Guide
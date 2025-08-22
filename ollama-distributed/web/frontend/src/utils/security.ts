/**
 * Security utilities and hardening functions
 */

// Content Security Policy utilities
export class CSPManager {
  private static instance: CSPManager
  private nonces: Map<string, string> = new Map()

  static getInstance(): CSPManager {
    if (!CSPManager.instance) {
      CSPManager.instance = new CSPManager()
    }
    return CSPManager.instance
  }

  generateNonce(): string {
    const array = new Uint8Array(16)
    crypto.getRandomValues(array)
    return btoa(String.fromCharCode.apply(null, Array.from(array)))
  }

  setNonce(type: string, nonce: string): void {
    this.nonces.set(type, nonce)
  }

  getNonce(type: string): string | undefined {
    return this.nonces.get(type)
  }

  createCSPHeader(): string {
    const scriptNonce = this.getNonce('script') || this.generateNonce()
    const styleNonce = this.getNonce('style') || this.generateNonce()

    return [
      "default-src 'self'",
      `script-src 'self' 'nonce-${scriptNonce}' 'strict-dynamic'`,
      `style-src 'self' 'nonce-${styleNonce}' 'unsafe-inline'`,
      "img-src 'self' data: https:",
      "font-src 'self' https:",
      "connect-src 'self' wss: ws:",
      "media-src 'self'",
      "object-src 'none'",
      "base-uri 'self'",
      "form-action 'self'",
      "frame-ancestors 'none'",
      "upgrade-insecure-requests"
    ].join('; ')
  }
}

// Input sanitization
export class InputSanitizer {
  private static htmlEscapeMap: Record<string, string> = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#x27;',
    '/': '&#x2F;'
  }

  static escapeHtml(text: string): string {
    return text.replace(/[&<>"'/]/g, (char) => this.htmlEscapeMap[char] || char)
  }

  static sanitizeInput(input: string, options: {
    allowedTags?: string[]
    maxLength?: number
    trimWhitespace?: boolean
  } = {}): string {
    const {
      allowedTags = [],
      maxLength = 1000,
      trimWhitespace = true
    } = options

    let sanitized = input

    // Trim whitespace if requested
    if (trimWhitespace) {
      sanitized = sanitized.trim()
    }

    // Enforce maximum length
    if (sanitized.length > maxLength) {
      sanitized = sanitized.substring(0, maxLength)
    }

    // Remove or escape HTML tags
    if (allowedTags.length === 0) {
      sanitized = this.escapeHtml(sanitized)
    } else {
      // More sophisticated tag filtering would go here
      // For now, escape all HTML
      sanitized = this.escapeHtml(sanitized)
    }

    return sanitized
  }

  static validateEmail(email: string): boolean {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    return emailRegex.test(email) && email.length <= 254
  }

  static validateUsername(username: string): boolean {
    const usernameRegex = /^[a-zA-Z0-9_-]{3,20}$/
    return usernameRegex.test(username)
  }

  static validatePassword(password: string): {
    isValid: boolean
    errors: string[]
  } {
    const errors: string[] = []
    
    if (password.length < 8) {
      errors.push('Password must be at least 8 characters long')
    }
    
    if (password.length > 128) {
      errors.push('Password must be no more than 128 characters long')
    }
    
    if (!/[a-z]/.test(password)) {
      errors.push('Password must contain at least one lowercase letter')
    }
    
    if (!/[A-Z]/.test(password)) {
      errors.push('Password must contain at least one uppercase letter')
    }
    
    if (!/\d/.test(password)) {
      errors.push('Password must contain at least one number')
    }
    
    if (!/[!@#$%^&*(),.?":{}|<>]/.test(password)) {
      errors.push('Password must contain at least one special character')
    }

    // Check for common patterns
    const commonPatterns = [
      /(.)\1{2,}/, // Three or more repeated characters
      /123456|654321|qwerty|password|admin/i, // Common sequences
    ]

    for (const pattern of commonPatterns) {
      if (pattern.test(password)) {
        errors.push('Password contains common patterns and is not secure')
        break
      }
    }

    return {
      isValid: errors.length === 0,
      errors
    }
  }
}

// Secure storage utilities
export class SecureStorage {
  private static readonly ENCRYPTION_KEY_NAME = 'app_storage_key'
  private static encryptionKey: CryptoKey | null = null

  private static async getOrCreateEncryptionKey(): Promise<CryptoKey> {
    if (this.encryptionKey) {
      return this.encryptionKey
    }

    // Try to load existing key
    const storedKey = localStorage.getItem(this.ENCRYPTION_KEY_NAME)
    if (storedKey) {
      try {
        const keyData = JSON.parse(storedKey)
        this.encryptionKey = await crypto.subtle.importKey(
          'raw',
          new Uint8Array(keyData),
          { name: 'AES-GCM' },
          false,
          ['encrypt', 'decrypt']
        )
        return this.encryptionKey
      } catch (error) {
        console.warn('Failed to load stored encryption key, generating new one')
      }
    }

    // Generate new key
    this.encryptionKey = await crypto.subtle.generateKey(
      { name: 'AES-GCM', length: 256 },
      true,
      ['encrypt', 'decrypt']
    )

    // Store key for future use
    const exportedKey = await crypto.subtle.exportKey('raw', this.encryptionKey)
    localStorage.setItem(this.ENCRYPTION_KEY_NAME, JSON.stringify(Array.from(new Uint8Array(exportedKey))))

    return this.encryptionKey
  }

  static async encryptData(data: string): Promise<string> {
    try {
      const key = await this.getOrCreateEncryptionKey()
      const iv = crypto.getRandomValues(new Uint8Array(12))
      const encodedData = new TextEncoder().encode(data)

      const encryptedData = await crypto.subtle.encrypt(
        { name: 'AES-GCM', iv },
        key,
        encodedData
      )

      // Combine IV and encrypted data
      const combined = new Uint8Array(iv.length + encryptedData.byteLength)
      combined.set(iv)
      combined.set(new Uint8Array(encryptedData), iv.length)

      return btoa(String.fromCharCode.apply(null, Array.from(combined)))
    } catch (error) {
      console.error('Encryption failed:', error)
      throw new Error('Failed to encrypt data')
    }
  }

  static async decryptData(encryptedData: string): Promise<string> {
    try {
      const key = await this.getOrCreateEncryptionKey()
      const combined = new Uint8Array(
        atob(encryptedData).split('').map(char => char.charCodeAt(0))
      )

      const iv = combined.slice(0, 12)
      const data = combined.slice(12)

      const decryptedData = await crypto.subtle.decrypt(
        { name: 'AES-GCM', iv },
        key,
        data
      )

      return new TextDecoder().decode(decryptedData)
    } catch (error) {
      console.error('Decryption failed:', error)
      throw new Error('Failed to decrypt data')
    }
  }

  static async setItem(key: string, value: string): Promise<void> {
    try {
      const encryptedValue = await this.encryptData(value)
      localStorage.setItem(key, encryptedValue)
    } catch (error) {
      console.error('Failed to store encrypted data:', error)
      // Fallback to regular storage (not recommended for sensitive data)
      localStorage.setItem(key, value)
    }
  }

  static async getItem(key: string): Promise<string | null> {
    const value = localStorage.getItem(key)
    if (!value) return null

    try {
      return await this.decryptData(value)
    } catch (error) {
      console.warn('Failed to decrypt stored data, returning raw value')
      return value
    }
  }

  static removeItem(key: string): void {
    localStorage.removeItem(key)
  }

  static clear(): void {
    localStorage.clear()
    this.encryptionKey = null
  }
}

// Rate limiting for client-side protection
export class RateLimiter {
  private attempts: Map<string, number[]> = new Map()
  private readonly maxAttempts: number
  private readonly timeWindow: number

  constructor(maxAttempts: number = 5, timeWindowMs: number = 60000) {
    this.maxAttempts = maxAttempts
    this.timeWindow = timeWindowMs
  }

  isAllowed(key: string): boolean {
    const now = Date.now()
    const attempts = this.attempts.get(key) || []

    // Remove old attempts outside the time window
    const recentAttempts = attempts.filter(timestamp => 
      now - timestamp < this.timeWindow
    )

    if (recentAttempts.length >= this.maxAttempts) {
      return false
    }

    // Record this attempt
    recentAttempts.push(now)
    this.attempts.set(key, recentAttempts)

    return true
  }

  getRemainingAttempts(key: string): number {
    const now = Date.now()
    const attempts = this.attempts.get(key) || []
    const recentAttempts = attempts.filter(timestamp => 
      now - timestamp < this.timeWindow
    )

    return Math.max(0, this.maxAttempts - recentAttempts.length)
  }

  reset(key: string): void {
    this.attempts.delete(key)
  }

  getTimeUntilReset(key: string): number {
    const attempts = this.attempts.get(key) || []
    if (attempts.length === 0) return 0

    const oldestAttempt = Math.min(...attempts)
    const timeUntilReset = this.timeWindow - (Date.now() - oldestAttempt)

    return Math.max(0, timeUntilReset)
  }
}

// Security headers utility
export const securityHeaders = {
  // Prevent clickjacking
  'X-Frame-Options': 'DENY',
  
  // Prevent MIME type sniffing
  'X-Content-Type-Options': 'nosniff',
  
  // Enable XSS protection
  'X-XSS-Protection': '1; mode=block',
  
  // Strict transport security (when using HTTPS)
  'Strict-Transport-Security': 'max-age=31536000; includeSubDomains',
  
  // Referrer policy
  'Referrer-Policy': 'strict-origin-when-cross-origin',
  
  // Permissions policy
  'Permissions-Policy': 'geolocation=(), microphone=(), camera=()',
}

// Utility to check if running in secure context
export const isSecureContext = (): boolean => {
  return window.isSecureContext || window.location.protocol === 'https:'
}

// Generate secure random values
export const generateSecureRandom = (length: number = 32): string => {
  const array = new Uint8Array(length)
  crypto.getRandomValues(array)
  return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('')
}

// Constant-time string comparison to prevent timing attacks
export const constantTimeCompare = (a: string, b: string): boolean => {
  if (a.length !== b.length) {
    return false
  }

  let result = 0
  for (let i = 0; i < a.length; i++) {
    result |= a.charCodeAt(i) ^ b.charCodeAt(i)
  }

  return result === 0
}

// Export instances for common use cases
export const cspManager = CSPManager.getInstance()
export const loginRateLimiter = new RateLimiter(5, 300000) // 5 attempts per 5 minutes
export const apiRateLimiter = new RateLimiter(100, 60000) // 100 requests per minute

export default {
  CSPManager,
  InputSanitizer,
  SecureStorage,
  RateLimiter,
  securityHeaders,
  isSecureContext,
  generateSecureRandom,
  constantTimeCompare,
  cspManager,
  loginRateLimiter,
  apiRateLimiter
}
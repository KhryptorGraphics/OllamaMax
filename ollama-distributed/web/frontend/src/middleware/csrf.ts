/**
 * CSRF Protection Middleware
 * Implements robust Cross-Site Request Forgery protection
 */

import { apiClient } from '@/services/api/client'

interface CSRFConfig {
  tokenName: string
  cookieName: string
  headerName: string
  secure: boolean
  sameSite: 'strict' | 'lax' | 'none'
  maxAge: number
}

class CSRFProtection {
  private static instance: CSRFProtection
  private config: CSRFConfig
  private token: string | null = null
  private tokenExpiry: number = 0

  private constructor() {
    this.config = {
      tokenName: 'csrf_token',
      cookieName: 'XSRF-TOKEN',
      headerName: 'X-CSRF-Token',
      secure: window.location.protocol === 'https:',
      sameSite: 'strict',
      maxAge: 60 * 60 * 1000 // 1 hour
    }
    
    this.initializeCSRF()
  }

  static getInstance(): CSRFProtection {
    if (!CSRFProtection.instance) {
      CSRFProtection.instance = new CSRFProtection()
    }
    return CSRFProtection.instance
  }

  private async initializeCSRF(): Promise<void> {
    try {
      // Try to get token from existing cookie first
      const existingToken = this.getTokenFromCookie()
      if (existingToken && this.isTokenValid()) {
        this.token = existingToken
        return
      }

      // Fetch new token from server
      await this.refreshToken()
    } catch (error) {
      console.error('Failed to initialize CSRF protection:', error)
    }
  }

  private getTokenFromCookie(): string | null {
    const cookies = document.cookie.split(';')
    for (const cookie of cookies) {
      const [name, value] = cookie.trim().split('=')
      if (name === this.config.cookieName) {
        return decodeURIComponent(value)
      }
    }
    return null
  }

  private setCookie(name: string, value: string, options: Partial<CSRFConfig> = {}): void {
    const opts = { ...this.config, ...options }
    let cookieString = `${name}=${encodeURIComponent(value)}`
    
    cookieString += `; Max-Age=${opts.maxAge}`
    cookieString += `; Path=/`
    cookieString += `; SameSite=${opts.sameSite}`
    
    if (opts.secure) {
      cookieString += `; Secure`
    }
    
    document.cookie = cookieString
  }

  private deleteCookie(name: string): void {
    document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 GMT; path=/`
  }

  private isTokenValid(): boolean {
    return this.token !== null && Date.now() < this.tokenExpiry
  }

  private generateTokenExpiry(): number {
    return Date.now() + this.config.maxAge
  }

  public async refreshToken(): Promise<string> {
    try {
      const response = await apiClient.request<{ token: string }>('/auth/csrf-token', {
        method: 'GET',
        skipCSRF: true // Skip CSRF for this specific request
      })

      if (response.success && response.data.token) {
        this.token = response.data.token
        this.tokenExpiry = this.generateTokenExpiry()
        
        // Set cookie for future requests
        this.setCookie(this.config.cookieName, this.token)
        
        return this.token
      } else {
        throw new Error('Invalid CSRF token response')
      }
    } catch (error) {
      console.error('Failed to refresh CSRF token:', error)
      throw error
    }
  }

  public async getToken(): Promise<string> {
    if (!this.isTokenValid()) {
      await this.refreshToken()
    }
    
    if (!this.token) {
      throw new Error('Unable to obtain CSRF token')
    }
    
    return this.token
  }

  public getHeaderName(): string {
    return this.config.headerName
  }

  public async getHeaders(): Promise<Record<string, string>> {
    try {
      const token = await this.getToken()
      return {
        [this.config.headerName]: token
      }
    } catch (error) {
      console.warn('Failed to get CSRF headers:', error)
      return {}
    }
  }

  public clearToken(): void {
    this.token = null
    this.tokenExpiry = 0
    this.deleteCookie(this.config.cookieName)
  }

  // Validate incoming responses for CSRF attacks
  public validateResponse(response: Response): boolean {
    // Check for presence of custom headers that indicate legitimate response
    const csrfCheck = response.headers.get('X-Content-Type-Options')
    const frameOptions = response.headers.get('X-Frame-Options')
    
    // Basic validation - server should set security headers
    if (!csrfCheck || !frameOptions) {
      console.warn('Response missing security headers, potential CSRF attack')
      return false
    }
    
    return true
  }

  // Form helper for traditional forms
  public addTokenToForm(form: HTMLFormElement): void {
    if (!this.token) {
      console.warn('No CSRF token available for form')
      return
    }

    // Remove existing token input if present
    const existingInput = form.querySelector(`input[name="${this.config.tokenName}"]`)
    if (existingInput) {
      existingInput.remove()
    }

    // Add hidden input with token
    const input = document.createElement('input')
    input.type = 'hidden'
    input.name = this.config.tokenName
    input.value = this.token
    form.appendChild(input)
  }

  // URL helper for GET requests with CSRF token
  public addTokenToUrl(url: string): string {
    if (!this.token) {
      return url
    }

    const urlObj = new URL(url, window.location.origin)
    urlObj.searchParams.set(this.config.tokenName, this.token)
    return urlObj.toString()
  }

  // Check if request needs CSRF protection
  public requiresCSRFProtection(method: string, url: string): boolean {
    // Only protect state-changing operations
    const protectedMethods = ['POST', 'PUT', 'PATCH', 'DELETE']
    if (!protectedMethods.includes(method.toUpperCase())) {
      return false
    }

    // Skip CSRF for certain endpoints (like CSRF token endpoint itself)
    const skipEndpoints = [
      '/auth/csrf-token',
      '/auth/logout' // logout should work even without CSRF token
    ]
    
    const urlPath = new URL(url, window.location.origin).pathname
    return !skipEndpoints.some(endpoint => urlPath.endsWith(endpoint))
  }

  // Middleware function for API client
  public async requestMiddleware(config: any): Promise<any> {
    const { method = 'GET', url, skipCSRF = false } = config

    if (!skipCSRF && this.requiresCSRFProtection(method, url)) {
      try {
        const csrfHeaders = await this.getHeaders()
        config.headers = {
          ...config.headers,
          ...csrfHeaders
        }
      } catch (error) {
        console.error('Failed to add CSRF headers:', error)
        // Don't fail the request, but log the error
      }
    }

    return config
  }

  // Response middleware for API client
  public responseMiddleware(response: Response): Response {
    // Validate response for CSRF attacks
    if (!this.validateResponse(response)) {
      console.warn('Potentially unsafe response detected')
    }

    // Check if server is asking for token refresh
    const refreshCSRF = response.headers.get('X-CSRF-Token-Refresh')
    if (refreshCSRF === 'required') {
      // Refresh token on next request
      this.clearToken()
    }

    return response
  }

  // Error handler for CSRF-related errors
  public handleCSRFError(error: any): boolean {
    if (error.status === 403 && error.code === 'CSRF_TOKEN_MISMATCH') {
      console.warn('CSRF token mismatch, refreshing token')
      this.clearToken()
      return true // Indicates error was handled
    }
    return false
  }
}

// Export singleton instance
export const csrfProtection = CSRFProtection.getInstance()

// Integrate with API client
export const setupCSRFProtection = () => {
  // Add request interceptor
  apiClient.addRequestInterceptor(csrfProtection.requestMiddleware.bind(csrfProtection))
  
  // Add response interceptor
  apiClient.addResponseInterceptor(csrfProtection.responseMiddleware.bind(csrfProtection))
  
  // Add error handler
  apiClient.addErrorHandler(csrfProtection.handleCSRFError.bind(csrfProtection))
}

// Utility functions for components
export const withCSRFProtection = async (config: any) => {
  return csrfProtection.requestMiddleware(config)
}

export const getCSRFToken = () => {
  return csrfProtection.getToken()
}

export const getCSRFHeaders = () => {
  return csrfProtection.getHeaders()
}

export const addCSRFToForm = (form: HTMLFormElement) => {
  csrfProtection.addTokenToForm(form)
}

export const addCSRFToUrl = (url: string) => {
  return csrfProtection.addTokenToUrl(url)
}

// Initialize CSRF protection when module loads
setupCSRFProtection()

export default csrfProtection
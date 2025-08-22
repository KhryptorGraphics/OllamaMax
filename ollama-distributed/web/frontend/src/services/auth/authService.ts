// Enhanced authentication service with comprehensive security features
import type {
  AuthState,
  User,
  LoginCredentials,
  RegisterData,
  AuthTokens,
  TokenClaims,
  Session,
  MFAChallenge,
  SecurityEvent,
} from '@/types/auth'
import type { ApiResponse } from '@/types/api'
import { apiClient } from '@/services/api/client'

class AuthService {
  private static instance: AuthService
  private readonly TOKEN_KEY = 'ollama_access_token'
  private readonly REFRESH_TOKEN_KEY = 'ollama_refresh_token'
  private readonly USER_KEY = 'ollama_user'
  
  private refreshTimer: number | null = null
  private eventListeners: Set<(state: AuthState) => void> = new Set()

  private constructor() {
    // Auto-refresh tokens
    this.setupTokenRefresh()
    
    // Listen for storage changes (multi-tab sync)
    window.addEventListener('storage', this.handleStorageChange.bind(this))
    
    // Listen for visibility changes to refresh on focus
    document.addEventListener('visibilitychange', this.handleVisibilityChange.bind(this))
  }

  static getInstance(): AuthService {
    if (!AuthService.instance) {
      AuthService.instance = new AuthService()
    }
    return AuthService.instance
  }

  // Event subscription for state changes
  subscribe(listener: (state: AuthState) => void): () => void {
    this.eventListeners.add(listener)
    return () => this.eventListeners.delete(listener)
  }

  private notify(state: AuthState): void {
    this.eventListeners.forEach(listener => {
      try {
        listener(state)
      } catch (error) {
        console.error('Error in auth state listener:', error)
      }
    })
  }

  // Token management
  private setTokens(tokens: AuthTokens): void {
    localStorage.setItem(this.TOKEN_KEY, tokens.accessToken)
    localStorage.setItem(this.REFRESH_TOKEN_KEY, tokens.refreshToken)
    
    // Update API client auth
    apiClient.updateConfig({
      auth: {
        type: 'bearer',
        credentials: tokens.accessToken,
      },
    })
    
    // Schedule token refresh
    this.scheduleTokenRefresh(tokens.expiresIn)
  }

  private getStoredToken(): string | null {
    return localStorage.getItem(this.TOKEN_KEY)
  }

  private getStoredRefreshToken(): string | null {
    return localStorage.getItem(this.REFRESH_TOKEN_KEY)
  }

  private clearTokens(): void {
    localStorage.removeItem(this.TOKEN_KEY)
    localStorage.removeItem(this.REFRESH_TOKEN_KEY)
    localStorage.removeItem(this.USER_KEY)
    
    // Clear API client auth
    apiClient.updateConfig({ auth: undefined })
    
    this.clearTokenRefresh()
  }

  private scheduleTokenRefresh(expiresIn: number): void {
    this.clearTokenRefresh()
    
    // Refresh 5 minutes before expiry
    const refreshTime = Math.max(0, (expiresIn * 1000) - (5 * 60 * 1000))
    
    this.refreshTimer = window.setTimeout(async () => {
      try {
        await this.refreshToken()
      } catch (error) {
        console.error('Auto token refresh failed:', error)
        await this.logout()
      }
    }, refreshTime)
  }

  private clearTokenRefresh(): void {
    if (this.refreshTimer) {
      clearTimeout(this.refreshTimer)
      this.refreshTimer = null
    }
  }

  private setupTokenRefresh(): void {
    const token = this.getStoredToken()
    if (token) {
      try {
        const claims = this.decodeToken(token)
        const now = Math.floor(Date.now() / 1000)
        const expiresIn = claims.exp - now
        
        if (expiresIn > 0) {
          this.scheduleTokenRefresh(expiresIn)
        } else {
          this.refreshToken().catch(() => this.logout())
        }
      } catch (error) {
        console.error('Invalid stored token:', error)
        this.clearTokens()
      }
    }
  }

  private decodeToken(token: string): TokenClaims {
    try {
      const payload = token.split('.')[1]
      const decoded = atob(payload.replace(/-/g, '+').replace(/_/g, '/'))
      return JSON.parse(decoded)
    } catch (error) {
      throw new Error('Invalid token format')
    }
  }

  private handleStorageChange(event: StorageEvent): void {
    if (event.key === this.TOKEN_KEY) {
      // Token changed in another tab
      if (!event.newValue) {
        // Token removed - user logged out in another tab
        this.clearTokenRefresh()
        this.notify(this.getInitialState())
      }
    }
  }

  private handleVisibilityChange(): void {
    if (!document.hidden && this.isAuthenticated()) {
      // Verify token is still valid when page becomes visible
      this.verifyToken().catch(() => this.logout())
    }
  }

  private getInitialState(): AuthState {
    return {
      isAuthenticated: false,
      user: null,
      token: null,
      refreshToken: null,
      expiresAt: null,
      permissions: [],
      loading: false,
      error: null,
    }
  }

  // Public authentication methods
  async login(credentials: LoginCredentials): Promise<User> {
    try {
      const response = await apiClient.request<{
        user: User
        tokens: AuthTokens
      }>('/auth/login', {
        method: 'POST',
        body: credentials,
      })

      if (!response.success) {
        throw new Error(response.error?.message || 'Login failed')
      }

      const { user, tokens } = response.data
      
      // Store tokens and user data
      this.setTokens(tokens)
      localStorage.setItem(this.USER_KEY, JSON.stringify(user))
      
      // Update state
      const state: AuthState = {
        isAuthenticated: true,
        user,
        token: tokens.accessToken,
        refreshToken: tokens.refreshToken,
        expiresAt: Date.now() + (tokens.expiresIn * 1000),
        permissions: user.permissions,
        loading: false,
        error: null,
      }
      
      this.notify(state)
      
      // Log security event
      this.logSecurityEvent('login_success', {
        userId: user.id,
        timestamp: new Date().toISOString(),
      })
      
      return user
      
    } catch (error) {
      // Log failed login attempt
      this.logSecurityEvent('login_failure', {
        username: credentials.username,
        error: (error as Error).message,
        timestamp: new Date().toISOString(),
      })
      
      throw error
    }
  }

  async register(data: RegisterData): Promise<User> {
    const response = await apiClient.request<{
      user: User
      tokens: AuthTokens
    }>('/auth/register', {
      method: 'POST',
      body: data,
    })

    if (!response.success) {
      throw new Error(response.error?.message || 'Registration failed')
    }

    const { user, tokens } = response.data
    
    this.setTokens(tokens)
    localStorage.setItem(this.USER_KEY, JSON.stringify(user))
    
    const state: AuthState = {
      isAuthenticated: true,
      user,
      token: tokens.accessToken,
      refreshToken: tokens.refreshToken,
      expiresAt: Date.now() + (tokens.expiresIn * 1000),
      permissions: user.permissions,
      loading: false,
      error: null,
    }
    
    this.notify(state)
    
    return user
  }

  async logout(): Promise<void> {
    try {
      // Notify server
      if (this.getStoredToken()) {
        await apiClient.request('/auth/logout', { method: 'POST' })
      }
    } catch (error) {
      // Continue with logout even if server call fails
      console.warn('Logout server call failed:', error)
    }

    // Clear local state
    this.clearTokens()
    
    // Update state
    this.notify(this.getInitialState())
    
    // Log security event
    this.logSecurityEvent('logout', {
      timestamp: new Date().toISOString(),
    })
  }

  async refreshToken(): Promise<AuthTokens> {
    const refreshToken = this.getStoredRefreshToken()
    if (!refreshToken) {
      throw new Error('No refresh token available')
    }

    const response = await apiClient.request<AuthTokens>('/auth/refresh', {
      method: 'POST',
      body: { refreshToken },
    })

    if (!response.success) {
      throw new Error(response.error?.message || 'Token refresh failed')
    }

    const tokens = response.data
    this.setTokens(tokens)
    
    // Log security event
    this.logSecurityEvent('token_refresh', {
      timestamp: new Date().toISOString(),
    })

    return tokens
  }

  async verifyToken(): Promise<boolean> {
    const token = this.getStoredToken()
    if (!token) return false

    try {
      const response = await apiClient.request<{ valid: boolean }>('/auth/verify', {
        method: 'POST',
        body: { token },
      })

      return response.success && response.data.valid
    } catch (error) {
      return false
    }
  }

  async changePassword(currentPassword: string, newPassword: string): Promise<void> {
    const response = await apiClient.request('/auth/password', {
      method: 'PUT',
      body: {
        currentPassword,
        newPassword,
      },
    })

    if (!response.success) {
      throw new Error(response.error?.message || 'Password change failed')
    }

    // Log security event
    this.logSecurityEvent('password_change', {
      timestamp: new Date().toISOString(),
    })
  }

  async resetPassword(email: string): Promise<void> {
    const response = await apiClient.request('/auth/reset', {
      method: 'POST',
      body: { email },
    })

    if (!response.success) {
      throw new Error(response.error?.message || 'Password reset failed')
    }
  }

  async updateProfile(updates: Partial<User>): Promise<User> {
    const response = await apiClient.request<User>('/auth/profile', {
      method: 'PUT',
      body: updates,
    })

    if (!response.success) {
      throw new Error(response.error?.message || 'Profile update failed')
    }

    const user = response.data
    localStorage.setItem(this.USER_KEY, JSON.stringify(user))
    
    // Update state
    const currentState = this.getCurrentState()
    const newState: AuthState = {
      ...currentState,
      user,
    }
    
    this.notify(newState)
    
    return user
  }

  // MFA methods
  async setupMFA(type: 'totp' | 'sms' | 'email'): Promise<any> {
    const response = await apiClient.request(`/auth/mfa/setup`, {
      method: 'POST',
      body: { type },
    })

    if (!response.success) {
      throw new Error(response.error?.message || 'MFA setup failed')
    }

    return response.data
  }

  async verifyMFA(challengeId: string, code: string): Promise<void> {
    const response = await apiClient.request('/auth/mfa/verify', {
      method: 'POST',
      body: { challengeId, code },
    })

    if (!response.success) {
      throw new Error(response.error?.message || 'MFA verification failed')
    }

    // Log security event
    this.logSecurityEvent('mfa_verified', {
      timestamp: new Date().toISOString(),
    })
  }

  // Session management
  async getSessions(): Promise<Session[]> {
    const response = await apiClient.request<Session[]>('/auth/sessions')

    if (!response.success) {
      throw new Error(response.error?.message || 'Failed to fetch sessions')
    }

    return response.data
  }

  async revokeSession(sessionId: string): Promise<void> {
    const response = await apiClient.request(`/auth/sessions/${sessionId}`, {
      method: 'DELETE',
    })

    if (!response.success) {
      throw new Error(response.error?.message || 'Failed to revoke session')
    }
  }

  async revokeAllSessions(): Promise<void> {
    const response = await apiClient.request('/auth/sessions', {
      method: 'DELETE',
    })

    if (!response.success) {
      throw new Error(response.error?.message || 'Failed to revoke sessions')
    }

    // This will invalidate current session too
    await this.logout()
  }

  // Security events
  async getSecurityEvents(): Promise<SecurityEvent[]> {
    const response = await apiClient.request<SecurityEvent[]>('/auth/security/events')

    if (!response.success) {
      throw new Error(response.error?.message || 'Failed to fetch security events')
    }

    return response.data
  }

  private async logSecurityEvent(type: string, metadata: any): Promise<void> {
    try {
      await apiClient.request('/auth/security/events', {
        method: 'POST',
        body: {
          type,
          metadata,
          timestamp: new Date().toISOString(),
        },
      })
    } catch (error) {
      console.warn('Failed to log security event:', error)
    }
  }

  // State getters
  isAuthenticated(): boolean {
    const token = this.getStoredToken()
    if (!token) return false

    try {
      const claims = this.decodeToken(token)
      const now = Math.floor(Date.now() / 1000)
      return claims.exp > now
    } catch (error) {
      return false
    }
  }

  getCurrentUser(): User | null {
    const userData = localStorage.getItem(this.USER_KEY)
    if (!userData) return null

    try {
      return JSON.parse(userData)
    } catch (error) {
      return null
    }
  }

  getCurrentState(): AuthState {
    const user = this.getCurrentUser()
    const token = this.getStoredToken()
    const refreshToken = this.getStoredRefreshToken()
    
    if (!user || !token) {
      return this.getInitialState()
    }

    try {
      const claims = this.decodeToken(token)
      return {
        isAuthenticated: true,
        user,
        token,
        refreshToken,
        expiresAt: claims.exp * 1000,
        permissions: user.permissions,
        loading: false,
        error: null,
      }
    } catch (error) {
      return this.getInitialState()
    }
  }

  hasPermission(permission: string): boolean {
    const user = this.getCurrentUser()
    if (!user) return false

    return user.permissions.some(p => 
      p.name === permission || 
      p.name === '*' || 
      p.resource === '*'
    )
  }

  hasRole(roleName: string): boolean {
    const user = this.getCurrentUser()
    if (!user) return false

    return user.role.name === roleName
  }

  // Cleanup
  destroy(): void {
    this.clearTokenRefresh()
    this.eventListeners.clear()
    window.removeEventListener('storage', this.handleStorageChange.bind(this))
    document.removeEventListener('visibilitychange', this.handleVisibilityChange.bind(this))
  }
}

// Export singleton instance
export const authService = AuthService.getInstance()
export default authService
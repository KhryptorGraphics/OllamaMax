// Authentication and authorization types

export interface User {
  id: string
  username: string
  email: string
  firstName: string
  lastName: string
  avatar?: string
  role: UserRole
  permissions: Permission[]
  tenant?: TenantInfo
  preferences: UserPreferences
  status: UserStatus
  lastLogin?: string
  createdAt: string
  updatedAt: string
}

export interface UserRole {
  id: string
  name: string
  description: string
  permissions: Permission[]
  isSystem: boolean
}

export interface Permission {
  id: string
  name: string
  resource: string
  action: string
  conditions?: Record<string, any>
}

export interface TenantInfo {
  id: string
  name: string
  domain: string
  settings: TenantSettings
}

export interface TenantSettings {
  branding: {
    logo?: string
    primaryColor: string
    secondaryColor: string
  }
  features: Record<string, boolean>
  limits: {
    maxUsers: number
    maxNodes: number
    maxModels: number
    storageQuota: number
  }
}

export interface UserPreferences {
  theme: 'light' | 'dark' | 'auto'
  language: string
  timezone: string
  notifications: NotificationPreferences
  dashboard: DashboardPreferences
}

export interface NotificationPreferences {
  email: boolean
  push: boolean
  alerts: {
    systemHealth: boolean
    taskFailures: boolean
    securityIssues: boolean
    performanceIssues: boolean
  }
}

export interface DashboardPreferences {
  layout: 'grid' | 'list'
  widgets: string[]
  refreshInterval: number
  showAdvancedMetrics: boolean
}

export type UserStatus = 'active' | 'inactive' | 'suspended' | 'pending'

// Authentication state and tokens
export interface AuthState {
  isAuthenticated: boolean
  user: User | null
  token: string | null
  refreshToken: string | null
  expiresAt: number | null
  permissions: Permission[]
  loading: boolean
  error: string | null
}

export interface LoginCredentials {
  username: string
  password: string
  rememberMe?: boolean
  tenantId?: string
}

export interface RegisterData {
  username: string
  email: string
  password: string
  confirmPassword: string
  firstName: string
  lastName: string
  tenantId?: string
  inviteCode?: string
}

export interface AuthTokens {
  accessToken: string
  refreshToken: string
  expiresIn: number
  tokenType: 'Bearer'
}

export interface TokenClaims {
  sub: string // user ID
  username: string
  email: string
  role: string
  permissions: string[]
  tenant?: string
  iat: number
  exp: number
  iss: string
  aud: string
}

// SSO and OAuth types
export interface SSOProvider {
  id: string
  name: string
  type: 'oauth2' | 'saml' | 'ldap'
  enabled: boolean
  config: SSOConfig
}

export interface SSOConfig {
  clientId?: string
  authUrl?: string
  tokenUrl?: string
  userInfoUrl?: string
  scopes?: string[]
  redirectUri?: string
  // SAML specific
  ssoUrl?: string
  entityId?: string
  certificate?: string
  // LDAP specific
  host?: string
  port?: number
  baseDN?: string
  userFilter?: string
}

export interface OAuthState {
  provider: string
  redirectUrl: string
  nonce: string
  codeVerifier?: string
}

// Session management
export interface Session {
  id: string
  userId: string
  deviceId: string
  deviceInfo: DeviceInfo
  ipAddress: string
  location?: LocationInfo
  createdAt: string
  lastActivity: string
  expiresAt: string
  isActive: boolean
}

export interface DeviceInfo {
  userAgent: string
  platform: string
  browser: string
  isMobile: boolean
  fingerprint: string
}

export interface LocationInfo {
  country: string
  region: string
  city: string
  timezone: string
}

// Multi-factor authentication
export interface MFAConfig {
  enabled: boolean
  methods: MFAMethod[]
  backupCodes: string[]
  recovery: {
    email: boolean
    phone: boolean
  }
}

export interface MFAMethod {
  type: 'totp' | 'sms' | 'email' | 'webauthn'
  enabled: boolean
  verified: boolean
  name?: string
  createdAt: string
  lastUsed?: string
}

export interface MFAChallenge {
  id: string
  type: MFAMethod['type']
  expiresAt: string
  attemptsRemaining: number
}

export interface WebAuthnCredential {
  id: string
  name: string
  credentialId: string
  publicKey: string
  counter: number
  deviceType: string
  createdAt: string
  lastUsed?: string
}

// Password and security
export interface PasswordPolicy {
  minLength: number
  requireUppercase: boolean
  requireLowercase: boolean
  requireNumbers: boolean
  requireSymbols: boolean
  preventReuse: number
  maxAge: number
  requireMFA: boolean
}

export interface SecurityEvent {
  id: string
  userId: string
  type: SecurityEventType
  description: string
  ipAddress: string
  userAgent: string
  location?: LocationInfo
  severity: 'low' | 'medium' | 'high' | 'critical'
  timestamp: string
  metadata?: Record<string, any>
}

export type SecurityEventType = 
  | 'login_success'
  | 'login_failure'
  | 'logout'
  | 'password_change'
  | 'mfa_enabled'
  | 'mfa_disabled'
  | 'account_locked'
  | 'suspicious_activity'
  | 'token_refresh'
  | 'permission_denied'

// API endpoints for auth
export interface AuthEndpoints {
  login: '/api/v1/auth/login'
  logout: '/api/v1/auth/logout'
  register: '/api/v1/auth/register'
  refresh: '/api/v1/auth/refresh'
  profile: '/api/v1/auth/profile'
  changePassword: '/api/v1/auth/password'
  resetPassword: '/api/v1/auth/reset'
  verifyEmail: '/api/v1/auth/verify'
  mfa: '/api/v1/auth/mfa'
  sessions: '/api/v1/auth/sessions'
  sso: '/api/v1/auth/sso'
}
// Core system types for OllamaMax Distributed System
export * from './auth'
export * from './ui'
export * from './websocket'
export * from './monitoring'

// Re-export specific types to avoid conflicts
export type { ApiResponse, PaginatedResponse } from './api'
export type { ClusterState } from './cluster'
export type { ModelInfo as ModelData, ModelSyncStatus as ModelSync } from './models'
export type { SecurityConfig, SecurityScanResult } from './security'
export type { PerformanceState as PerfState, PerformanceMetrics as PerfMetrics } from './performance'

// Global application types
export interface AppConfig {
  apiBaseUrl: string
  wsUrl: string
  version: string
  buildTime: string
  features: FeatureFlags
}

export interface FeatureFlags {
  authentication: boolean
  realTimeUpdates: boolean
  performanceMonitoring: boolean
  securityDashboard: boolean
  mobileApp: boolean
  offlineMode: boolean
  multiTenant: boolean
  federation: boolean
}

export interface ApiResponse<T = any> {
  success: boolean
  data?: T
  error?: string
  timestamp: string
  requestId?: string
}

export interface PaginatedResponse<T = any> extends ApiResponse<T[]> {
  pagination: {
    page: number
    limit: number
    total: number
    hasMore: boolean
  }
}

export interface ErrorInfo {
  code: string
  message: string
  details?: any
  timestamp: string
  context?: Record<string, any>
}

// Generic utility types
export type Optional<T, K extends keyof T> = Pick<Partial<T>, K> & Omit<T, K>
export type RequiredFields<T, K extends keyof T> = T & Required<Pick<T, K>>
export type DeepPartial<T> = {
  [P in keyof T]?: T[P] extends object ? DeepPartial<T[P]> : T[P]
}

// Event system types
export interface AppEvent<T = any> {
  type: string
  payload: T
  timestamp: number
  source: string
}

export type EventHandler<T = any> = (event: AppEvent<T>) => void

// Theme and styling types
export interface Theme {
  colors: {
    primary: string
    secondary: string
    background: string
    surface: string
    text: string
    error: string
    warning: string
    success: string
    info: string
  }
  spacing: {
    xs: string
    sm: string
    md: string
    lg: string
    xl: string
  }
  typography: {
    fontFamily: string
    fontSize: {
      xs: string
      sm: string
      md: string
      lg: string
      xl: string
    }
  }
  breakpoints: {
    mobile: string
    tablet: string
    desktop: string
    wide: string
  }
}

// Global state types
export interface GlobalState {
  auth: any // TODO: Import proper AuthState type
  cluster: any // TODO: Import proper ClusterState type
  ui: UIState
  websocket: any // TODO: Import proper WebSocketState type
  performance: any // TODO: Import proper PerformanceState type
  monitoring: any // TODO: Import proper MonitoringState type
}

export interface UIState {
  theme: 'light' | 'dark' | 'auto'
  sidebarOpen: boolean
  loading: boolean
  notifications: Notification[]
  modal: {
    isOpen: boolean
    component?: string
    props?: any
  }
}

export interface Notification {
  id: string
  type: 'info' | 'success' | 'warning' | 'error'
  title: string
  message: string
  timestamp: number
  duration?: number
  actions?: NotificationAction[]
}

export interface NotificationAction {
  label: string
  action: () => void
  variant?: 'primary' | 'secondary'
}
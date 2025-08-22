// Main application component
import React, { useEffect, Suspense } from 'react'
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom'
import { ThemeProvider } from 'styled-components'
import { useStore } from '@/stores'
import { authService } from '@/services/auth/authService'
import { theme } from '@/theme'
import { GlobalStyles } from '@/theme/GlobalStyles'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { NotificationCenter } from '@/components/common/NotificationCenter'
import { ModalManager } from '@/components/common/ModalManager'

// Layout components
import { MainLayout } from '@/components/layout'

// Lazy load pages for better performance
const LoginPage = React.lazy(() => import('@/pages/auth/LoginPage'))
const DashboardPage = React.lazy(() => import('@/pages/dashboard/DashboardPage'))
const NodesPage = React.lazy(() => import('@/pages/nodes/NodesPage'))
const ModelsPage = React.lazy(() => import('@/pages/models/ModelsPage'))
const MonitoringPage = React.lazy(() => import('@/pages/monitoring/MonitoringPage'))
const TasksPage = React.lazy(() => import('@/pages/tasks/TasksPage'))
const TransfersPage = React.lazy(() => import('@/pages/transfers/TransfersPage'))
const SecurityPage = React.lazy(() => import('@/pages/security/SecurityPage'))
const PerformancePage = React.lazy(() => import('@/pages/performance/PerformancePage'))
const SettingsPage = React.lazy(() => import('@/pages/settings/SettingsPage'))
const ProfilePage = React.lazy(() => import('@/pages/profile/ProfilePage'))

// Protected Route component
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const isAuthenticated = useStore((state) => state.auth.isAuthenticated)
  
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }
  
  return <>{children}</>
}

// Public Route component (redirect to dashboard if authenticated)
const PublicRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const isAuthenticated = useStore((state) => state.auth.isAuthenticated)
  
  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />
  }
  
  return <>{children}</>
}

// Main App component
const App: React.FC = () => {
  const { auth, ui, connect } = useStore()

  useEffect(() => {
    // Initialize auth state from storage
    const initAuth = async () => {
      try {
        const storedState = authService.getCurrentState()
        if (storedState.isAuthenticated) {
          // Verify token is still valid
          const isValid = await authService.verifyToken()
          if (!isValid) {
            await authService.logout()
          } else {
            // Connect WebSocket for authenticated users
            try {
              await connect()
            } catch (error) {
              console.warn('WebSocket connection failed:', error)
            }
          }
        }
      } catch (error) {
        console.error('Auth initialization failed:', error)
        await authService.logout()
      }
    }

    initAuth()
  }, [connect])

  // Handle theme changes
  useEffect(() => {
    const applyTheme = () => {
      const root = document.documentElement
      let appliedTheme = ui.theme

      // Auto theme detection
      if (ui.theme === 'auto') {
        const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
        appliedTheme = prefersDark ? 'dark' : 'light'
      }

      root.setAttribute('data-theme', appliedTheme)
      
      // Update meta theme-color for mobile browsers
      const metaThemeColor = document.querySelector('meta[name=theme-color]')
      if (metaThemeColor) {
        metaThemeColor.setAttribute(
          'content', 
          appliedTheme === 'dark' ? '#1a1a1a' : '#ffffff'
        )
      }
    }

    applyTheme()

    // Listen for system theme changes
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    const handleChange = () => {
      if (ui.theme === 'auto') {
        applyTheme()
      }
    }

    mediaQuery.addEventListener('change', handleChange)
    return () => mediaQuery.removeEventListener('change', handleChange)
  }, [ui.theme])

  // Handle responsive sidebar
  useEffect(() => {
    const handleResize = () => {
      const isMobile = window.innerWidth < 1024
      if (isMobile && ui.sidebarOpen) {
        useStore.getState().setSidebarOpen(false)
      }
    }

    window.addEventListener('resize', handleResize)
    return () => window.removeEventListener('resize', handleResize)
  }, [ui.sidebarOpen])

  return (
    <ErrorBoundary>
      <ThemeProvider theme={theme}>
        <GlobalStyles />
        
        <Router>
          <div className="app" data-theme={ui.theme}>
            <Suspense 
              fallback={
                <div className="app-loading">
                  <LoadingSpinner size="lg" tip="Loading application..." />
                </div>
              }
            >
              <Routes>
                {/* Public routes */}
                <Route 
                  path="/login" 
                  element={
                    <PublicRoute>
                      <LoginPage />
                    </PublicRoute>
                  } 
                />
                
                {/* Protected routes with layout */}
                <Route 
                  path="/" 
                  element={
                    <ProtectedRoute>
                      <MainLayout />
                    </ProtectedRoute>
                  }
                >
                  <Route path="dashboard" element={<DashboardPage />} />
                  <Route path="models" element={<ModelsPage />} />
                  <Route path="nodes" element={<NodesPage />} />
                  <Route path="monitoring" element={<MonitoringPage />} />
                  <Route path="tasks" element={<TasksPage />} />
                  <Route path="transfers" element={<TransfersPage />} />
                  <Route path="security" element={<SecurityPage />} />
                  <Route path="performance" element={<PerformancePage />} />
                  <Route path="settings" element={<SettingsPage />} />
                  <Route path="profile" element={<ProfilePage />} />
                </Route>
                
                {/* Default redirect */}
                <Route 
                  index
                  element={
                    auth.isAuthenticated ? (
                      <Navigate to="/dashboard" replace />
                    ) : (
                      <Navigate to="/login" replace />
                    )
                  } 
                />
                
                {/* 404 fallback */}
                <Route 
                  path="*" 
                  element={
                    <div className="error-page">
                      <h1>404 - Page Not Found</h1>
                      <p>The page you're looking for doesn't exist.</p>
                      <a href="/">Go Home</a>
                    </div>
                  } 
                />
              </Routes>
            </Suspense>
            
            {/* Global UI components */}
            <NotificationCenter />
            <ModalManager />
          </div>
        </Router>
      </ThemeProvider>
    </ErrorBoundary>
  )
}

export default App
import React, { useState, useEffect } from 'react'
import { Header, SideNav, Breadcrumbs } from '@ollamamax/ui'
import { BrowserRouter, Routes, Route, Navigate, useLocation, useNavigate } from 'react-router-dom'
import { Login, Register, ForgotPassword, ResetPassword, VerifyEmail } from './routes/auth'
import { PWAInstallPrompt } from './components/pwa/PWAInstallPrompt'
import { PWAUpdateNotification } from './components/pwa/PWAUpdateNotification'
import { MobileNavigation } from './components/mobile/MobileNavigation'
import { ResponsiveDashboard } from './components/responsive/ResponsiveDashboard'
import { usePWA } from './hooks/usePWA'
import { usePushNotifications } from './hooks/usePushNotifications'
import { useOfflineSync } from './hooks/useOfflineSync'

function useFlag(name: string): boolean {
  try { return localStorage.getItem(name) === '1' || (window as any)[name] === true } catch { return false }
}

function Dashboard() {
  const { queueOperation } = useOfflineSync();
  
  const handleCardClick = (id: string) => {
    console.log('Card clicked:', id);
    // Navigate to specific metric detail page
  };

  const handleRefresh = async () => {
    console.log('Refreshing dashboard data...');
    
    // If offline, queue the refresh operation
    if (!navigator.onLine) {
      await queueOperation({
        type: 'update',
        resource: 'dashboard-metrics',
        data: { timestamp: Date.now() },
        maxRetries: 3
      });
    }
    
    // Simulate data refresh
    await new Promise(resolve => setTimeout(resolve, 1500));
  };

  return (
    <div className="omx-v2 min-h-screen pb-16 sm:pb-0">
      <ResponsiveDashboard
        onCardClick={handleCardClick}
        onRefresh={handleRefresh}
        className="p-4 sm:p-6"
      />
    </div>
  )
}

function Shell() {
  const location = useLocation()
  const navigate = useNavigate()
  const useSideNav = useFlag('USE_SHARED_SIDENAV')
  const useBreadcrumbs = useFlag('USE_SHARED_BREADCRUMBS')
  
  // PWA hooks
  const { isStandalone, isOnline } = usePWA();
  const { permission: notificationPermission } = usePushNotifications();
  const { hasUnsyncedData, pendingOperations } = useOfflineSync();
  
  // Local state
  const [showOfflineIndicator, setShowOfflineIndicator] = useState(!isOnline);

  const links = [
    { label: 'Dashboard', href: '/v2' },
    { label: 'Login', href: '/v2/auth/login' },
    { label: 'Register', href: '/v2/auth/register' },
  ]

  const items = [
    { label: 'Dashboard', href: '/v2' },
    { label: 'Auth', href: '/v2/auth/login' },
    { label: 'Register', href: '/v2/auth/register' },
  ]

  const parts = location.pathname.replace(/^\/+|\/+$/g,'').split('/')
  const crumbs = [{ label: 'Home', href: '/v2' }, ...parts.slice(1).map((p, i) => ({
    label: p.replace(/-/g,' ').replace(/\b\w/g, (m) => m.toUpperCase()),
    href: '/'+parts.slice(0, i+2).join('/'),
  }))]

  // Monitor online status changes
  useEffect(() => {
    setShowOfflineIndicator(!isOnline);
    
    // Auto-hide offline indicator after 3 seconds when coming back online
    if (isOnline) {
      const timer = setTimeout(() => setShowOfflineIndicator(false), 3000);
      return () => clearTimeout(timer);
    }
  }, [isOnline]);

  return (
    <>
      {/* PWA Components */}
      <PWAInstallPrompt />
      <PWAUpdateNotification />
      
      {/* Mobile Navigation */}
      <MobileNavigation onNavigate={(path) => navigate(path)} />
      
      {/* Offline Indicator */}
      {showOfflineIndicator && (
        <div className={`fixed top-0 left-0 right-0 z-40 text-center text-white text-sm py-2 transition-colors duration-300 ${
          isOnline ? 'bg-green-600' : 'bg-red-600'
        }`}>
          {isOnline ? (
            <span>✅ Back online! {hasUnsyncedData && `Syncing ${pendingOperations} pending changes...`}</span>
          ) : (
            <span>📱 You're offline. Changes will sync when connected.</span>
          )}
        </div>
      )}
      
      {/* Notification Permission Prompt */}
      {notificationPermission === 'default' && isStandalone && (
        <div className="fixed bottom-20 sm:bottom-4 left-4 right-4 bg-blue-600 text-white p-4 rounded-lg shadow-lg z-30">
          <div className="flex items-center justify-between">
            <div>
              <p className="font-medium">Enable Notifications</p>
              <p className="text-sm text-blue-100">Get updates about your cluster status</p>
            </div>
            <button 
              onClick={() => {
                // This would be handled by the usePushNotifications hook
                console.log('Request notification permission');
              }}
              className="bg-white text-blue-600 px-4 py-2 rounded-lg text-sm font-medium hover:bg-blue-50 transition-colors"
            >
              Enable
            </button>
          </div>
        </div>
      )}

      <div className={`omx-v2 min-h-screen flex ${showOfflineIndicator ? 'pt-10' : ''}`}>
        <div className="w-full">
          {/* Desktop Header */}
          <div className="hidden lg:block">
            <Header brand={<span>OllamaMax</span>} links={links} />
          </div>
          
          {useBreadcrumbs && <Breadcrumbs items={crumbs} />}
          
          <div className="flex">
            {useSideNav && (
              <div className="hidden lg:block">
                <SideNav items={items} activeHref={location.pathname} onNavigate={(href)=>navigate(href)} />
              </div>
            )}
            
            <main id="main" className="flex-1 lg:p-8" tabIndex={-1}>
              <Routes>
                <Route index element={<Dashboard />} />
                <Route path="auth/login" element={<Login />} />
                <Route path="auth/register" element={<Register />} />
                <Route path="auth/forgot-password" element={<ForgotPassword />} />
                <Route path="auth/reset-password" element={<ResetPassword />} />
                <Route path="auth/verify-email" element={<VerifyEmail />} />
              </Routes>
            </main>
          </div>
        </div>
      </div>
    </>
  )
}

export default function AppMinimal() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/v2/*" element={<Shell />} />
        <Route path="*" element={<Navigate to="/v2" replace />} />
      </Routes>
    </BrowserRouter>
  )
}

/**
 * Main Application Component
 *
 * Root component that handles routing, authentication, theme management, and PWA features.
 */

import React, { useEffect, useState } from 'react';
import { ThemeProvider, useResponsive } from './design-system/theme/ThemeProvider.jsx';
import { AuthProvider, useAuth } from './contexts/AuthContext.jsx';
import { ToastContainer } from './design-system/index.js';
import AuthPage from './components/auth/AuthPage.jsx';
import RealTimeDashboard from './components/dashboard/RealTimeDashboard.jsx';
import MobileDashboard from './components/mobile/MobileDashboard.jsx';
import UserProfile from './components/user/UserProfile.jsx';
import GlobalStyles from './design-system/theme/GlobalStyles.jsx';
import pwaService from './services/pwaService.js';

// Enhanced router component with PWA and responsive features
const Router = () => {
  const { isAuthenticated, loading } = useAuth();
  const { isMobile, isTablet } = useResponsive();
  const [currentRoute, setCurrentRoute] = useState('dashboard');
  const [toasts, setToasts] = useState([]);
  const [pwaReady, setPwaReady] = useState(false);

  // Initialize PWA features
  useEffect(() => {
    // Override PWA service event handlers
    pwaService.onNetworkStatusChange = (isOnline) => {
      addToast(
        isOnline ? 'success' : 'warning',
        'Connection Status',
        isOnline ? 'Back online' : 'You are now offline'
      );
    };

    pwaService.onUpdateAvailable = () => {
      addToast(
        'info',
        'Update Available',
        'A new version is available. Refresh to update.',
        {
          label: 'Refresh',
          onClick: () => window.location.reload()
        }
      );
    };

    pwaService.onOfflineReady = () => {
      setPwaReady(true);
      addToast(
        'success',
        'App Ready',
        'OllamaMax is ready for offline use!'
      );
    };

    pwaService.onInstallAvailable = () => {
      addToast(
        'info',
        'Install Available',
        'Install OllamaMax for a better experience',
        {
          label: 'Install',
          onClick: () => pwaService.showInstallPrompt()
        }
      );
    };

    return () => {
      pwaService.destroy();
    };
  }, []);

  // Add toast notification
  const addToast = (type, title, message, action) => {
    const toast = {
      id: Date.now().toString(),
      type,
      title,
      message,
      action,
      onClose: removeToast
    };

    setToasts(prev => [...prev, toast]);
  };

  // Remove toast notification
  const removeToast = (id) => {
    setToasts(prev => prev.filter(toast => toast.id !== id));
  };

  // Loading state
  if (loading) {
    return (
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        height: '100vh',
        fontSize: '1.125rem',
        color: '#64748b'
      }}>
        <div style={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          gap: '1rem'
        }}>
          <div style={{
            width: '40px',
            height: '40px',
            border: '4px solid #e2e8f0',
            borderTopColor: '#0ea5e9',
            borderRadius: '50%',
            animation: 'spin 1s linear infinite'
          }} />
          <span>Loading OllamaMax...</span>
        </div>
      </div>
    );
  }

  // Show authentication page if not authenticated
  if (!isAuthenticated) {
    return (
      <>
        <AuthPage />
        <ToastContainer toasts={toasts} position="top-right" />
      </>
    );
  }

  // Authenticated routes with responsive design
  const renderDashboard = () => {
    if (isMobile) {
      return <MobileDashboard />;
    }
    return <RealTimeDashboard />;
  };

  const renderContent = () => {
    switch (currentRoute) {
      case 'profile':
        return <UserProfile />;
      case 'dashboard':
      default:
        return renderDashboard();
    }
  };

  return (
    <>
      {renderContent()}
      <ToastContainer toasts={toasts} position="top-right" />
    </>
  );
};

// Main App component
const App = () => {
  return (
    <ThemeProvider defaultTheme="light">
      <AuthProvider>
        <GlobalStyles />
        <Router />
      </AuthProvider>
    </ThemeProvider>
  );
};

export default App;

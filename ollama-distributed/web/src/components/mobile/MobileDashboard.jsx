/**
 * Mobile Dashboard Component
 * 
 * Mobile-optimized dashboard with touch-friendly interactions and responsive layout.
 */

import React, { useState, useEffect } from 'react';
import { Card, Button, Badge, Modal } from '../../design-system/index.js';
import { useTheme, useResponsive } from '../../design-system/theme/ThemeProvider.jsx';
import { useAuth } from '../../contexts/AuthContext.jsx';
import pwaService from '../../services/pwaService.js';

const MobileDashboard = () => {
  const { theme, toggleTheme, currentTheme } = useTheme();
  const { isMobile, isTablet } = useResponsive();
  const { user, logout } = useAuth();
  
  const [dashboardData, setDashboardData] = useState({
    clusterStatus: 'healthy',
    nodeCount: 3,
    activeModels: 5,
    totalRequests: 1247,
    avgResponseTime: 245,
    errorRate: 0.02,
    uptime: '99.9%',
    nodes: [
      { id: 'node-1', status: 'healthy', cpu: 45, memory: 67, requests: 423 },
      { id: 'node-2', status: 'healthy', cpu: 52, memory: 71, requests: 389 },
      { id: 'node-3', status: 'warning', cpu: 78, memory: 89, requests: 435 }
    ]
  });
  
  const [selectedMetric, setSelectedMetric] = useState(null);
  const [showMobileMenu, setShowMobileMenu] = useState(false);
  const [pullToRefresh, setPullToRefresh] = useState(false);
  const [touchStart, setTouchStart] = useState(null);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [pwaInfo, setPwaInfo] = useState(pwaService.getAppInfo());

  // Update PWA info
  useEffect(() => {
    const updatePwaInfo = () => {
      setPwaInfo(pwaService.getAppInfo());
    };
    
    // Override PWA service event handlers
    pwaService.onNetworkStatusChange = (isOnline) => {
      updatePwaInfo();
      if (isOnline) {
        refreshData();
      }
    };
    
    pwaService.onInstallAvailable = updatePwaInfo;
    pwaService.onAppInstalled = updatePwaInfo;
    
    updatePwaInfo();
  }, []);

  // Pull-to-refresh functionality
  const handleTouchStart = (e) => {
    setTouchStart(e.touches[0].clientY);
  };

  const handleTouchMove = (e) => {
    if (!touchStart) return;
    
    const currentTouch = e.touches[0].clientY;
    const diff = currentTouch - touchStart;
    
    if (diff > 100 && window.scrollY === 0) {
      setPullToRefresh(true);
    }
  };

  const handleTouchEnd = () => {
    if (pullToRefresh) {
      setIsRefreshing(true);
      refreshData();
      setTimeout(() => {
        setIsRefreshing(false);
        setPullToRefresh(false);
      }, 1500);
    }
    setTouchStart(null);
    setPullToRefresh(false);
  };

  // Refresh data
  const refreshData = async () => {
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      setDashboardData(prev => ({
        ...prev,
        totalRequests: prev.totalRequests + Math.floor(Math.random() * 10),
        avgResponseTime: 200 + Math.floor(Math.random() * 100),
        errorRate: Math.max(0, Math.random() * 0.05)
      }));
    } catch (error) {
      console.error('Failed to refresh data:', error);
    }
  };

  // Install PWA
  const handleInstallPWA = async () => {
    const installed = await pwaService.showInstallPrompt();
    if (installed) {
      setPwaInfo(pwaService.getAppInfo());
    }
  };

  // Request notifications
  const handleEnableNotifications = async () => {
    const granted = await pwaService.requestNotificationPermission();
    if (granted) {
      setPwaInfo(pwaService.getAppInfo());
    }
  };

  // Container styles
  const containerStyles = {
    minHeight: '100vh',
    backgroundColor: theme.colors.background,
    paddingBottom: '80px', // Space for bottom navigation
    position: 'relative'
  };

  // Header styles
  const headerStyles = {
    position: 'sticky',
    top: 0,
    zIndex: 1000,
    backgroundColor: theme.colors.surface,
    borderBottom: `1px solid ${theme.colors.border}`,
    padding: '1rem',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between'
  };

  // Pull to refresh indicator
  const pullToRefreshStyles = {
    position: 'absolute',
    top: pullToRefresh ? '60px' : '-40px',
    left: '50%',
    transform: 'translateX(-50%)',
    transition: 'top 0.3s ease',
    zIndex: 999,
    backgroundColor: theme.colors.surface,
    padding: '0.5rem 1rem',
    borderRadius: '20px',
    boxShadow: theme.colors.shadow,
    display: 'flex',
    alignItems: 'center',
    gap: '0.5rem'
  };

  // Content styles
  const contentStyles = {
    padding: '1rem',
    paddingTop: pullToRefresh ? '2rem' : '1rem',
    transition: 'padding-top 0.3s ease'
  };

  // Mobile metric card
  const MobileMetricCard = ({ title, value, subtitle, trend, icon, color, onClick }) => (
    <Card 
      variant="elevated" 
      hover 
      interactive
      onClick={onClick}
      style={{
        minHeight: '120px',
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'space-between'
      }}
    >
      <div style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        marginBottom: '0.75rem'
      }}>
        <div style={{
          width: '40px',
          height: '40px',
          backgroundColor: color + '15',
          borderRadius: '10px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          fontSize: '1.25rem'
        }}>
          {icon}
        </div>
        {trend !== undefined && (
          <Badge 
            variant={trend > 0 ? 'outline-success' : trend < 0 ? 'outline-error' : 'outline'}
            size="sm"
          >
            {trend > 0 ? '↗' : trend < 0 ? '↘' : '→'} {Math.abs(trend)}%
          </Badge>
        )}
      </div>
      
      <div>
        <div style={{
          fontSize: '1.5rem',
          fontWeight: 'bold',
          color: theme.colors.text,
          marginBottom: '0.25rem'
        }}>
          {value}
        </div>
        <div style={{
          fontSize: '0.875rem',
          color: theme.colors.textSecondary,
          lineHeight: '1.2'
        }}>
          {title}
        </div>
      </div>
    </Card>
  );

  // Mobile node card
  const MobileNodeCard = ({ node }) => (
    <Card variant="elevated" style={{ marginBottom: '0.75rem' }}>
      <div style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        marginBottom: '0.75rem'
      }}>
        <h4 style={{
          fontSize: '1rem',
          fontWeight: '600',
          color: theme.colors.text,
          margin: 0
        }}>
          {node.id}
        </h4>
        <Badge 
          variant={node.status === 'healthy' ? 'success' : node.status === 'warning' ? 'warning' : 'error'}
          size="sm"
        >
          {node.status}
        </Badge>
      </div>
      
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.75rem' }}>
        <div>
          <div style={{
            fontSize: '0.75rem',
            color: theme.colors.textSecondary,
            marginBottom: '0.25rem'
          }}>
            CPU: {node.cpu}%
          </div>
          <div style={{
            width: '100%',
            height: '4px',
            backgroundColor: theme.colors.border,
            borderRadius: '2px',
            overflow: 'hidden'
          }}>
            <div style={{
              width: `${node.cpu}%`,
              height: '100%',
              backgroundColor: node.cpu > 80 ? theme.colors.error : node.cpu > 60 ? theme.colors.warning : theme.colors.success,
              transition: 'width 0.3s ease'
            }} />
          </div>
        </div>
        
        <div>
          <div style={{
            fontSize: '0.75rem',
            color: theme.colors.textSecondary,
            marginBottom: '0.25rem'
          }}>
            Memory: {node.memory}%
          </div>
          <div style={{
            width: '100%',
            height: '4px',
            backgroundColor: theme.colors.border,
            borderRadius: '2px',
            overflow: 'hidden'
          }}>
            <div style={{
              width: `${node.memory}%`,
              height: '100%',
              backgroundColor: node.memory > 80 ? theme.colors.error : node.memory > 60 ? theme.colors.warning : theme.colors.success,
              transition: 'width 0.3s ease'
            }} />
          </div>
        </div>
      </div>
    </Card>
  );

  // Bottom navigation
  const BottomNavigation = () => (
    <div style={{
      position: 'fixed',
      bottom: 0,
      left: 0,
      right: 0,
      backgroundColor: theme.colors.surface,
      borderTop: `1px solid ${theme.colors.border}`,
      padding: '0.75rem',
      display: 'grid',
      gridTemplateColumns: 'repeat(4, 1fr)',
      gap: '0.5rem',
      zIndex: 1000
    }}>
      {[
        { icon: '📊', label: 'Dashboard', active: true },
        { icon: '🖥️', label: 'Nodes', active: false },
        { icon: '📈', label: 'Metrics', active: false },
        { icon: '⚙️', label: 'Settings', active: false }
      ].map((item, index) => (
        <button
          key={index}
          style={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            gap: '0.25rem',
            padding: '0.5rem',
            border: 'none',
            backgroundColor: 'transparent',
            color: item.active ? theme.colors.primary : theme.colors.textSecondary,
            fontSize: '0.75rem',
            cursor: 'pointer',
            borderRadius: '8px',
            transition: 'all 0.2s ease'
          }}
        >
          <span style={{ fontSize: '1.25rem' }}>{item.icon}</span>
          <span style={{ fontWeight: item.active ? '600' : '400' }}>
            {item.label}
          </span>
        </button>
      ))}
    </div>
  );

  return (
    <div 
      style={containerStyles}
      onTouchStart={handleTouchStart}
      onTouchMove={handleTouchMove}
      onTouchEnd={handleTouchEnd}
    >
      {/* Pull to refresh indicator */}
      <div style={pullToRefreshStyles}>
        <div style={{
          width: '20px',
          height: '20px',
          border: `2px solid ${theme.colors.primary}`,
          borderTopColor: 'transparent',
          borderRadius: '50%',
          animation: isRefreshing ? 'spin 1s linear infinite' : 'none'
        }} />
        <span style={{ color: theme.colors.primary, fontSize: '0.875rem' }}>
          {isRefreshing ? 'Refreshing...' : 'Pull to refresh'}
        </span>
      </div>

      {/* Header */}
      <div style={headerStyles}>
        <div>
          <h1 style={{
            fontSize: '1.25rem',
            fontWeight: 'bold',
            color: theme.colors.text,
            margin: 0
          }}>
            OllamaMax
          </h1>
          <div style={{
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem',
            marginTop: '0.25rem'
          }}>
            <div style={{
              width: '8px',
              height: '8px',
              borderRadius: '50%',
              backgroundColor: pwaInfo.isOnline ? theme.colors.success : theme.colors.error
            }} />
            <span style={{
              fontSize: '0.75rem',
              color: theme.colors.textSecondary
            }}>
              {pwaInfo.isOnline ? 'Online' : 'Offline'}
            </span>
          </div>
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <Button
            variant="ghost"
            size="sm"
            onClick={toggleTheme}
            aria-label="Toggle theme"
          >
            {currentTheme === 'light' ? '🌙' : '☀️'}
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setShowMobileMenu(true)}
            aria-label="Menu"
          >
            ⋮
          </Button>
        </div>
      </div>

      <div style={contentStyles}>
        {/* PWA Features */}
        {(pwaInfo.canInstall || pwaInfo.notificationPermission === 'default') && (
          <Card variant="elevated" style={{ marginBottom: '1.5rem' }}>
            <div style={{
              display: 'flex',
              flexDirection: 'column',
              gap: '0.75rem'
            }}>
              <h3 style={{
                fontSize: '1rem',
                fontWeight: '600',
                color: theme.colors.text,
                margin: 0
              }}>
                Enhance Your Experience
              </h3>
              
              {pwaInfo.canInstall && (
                <Button
                  variant="primary"
                  size="sm"
                  onClick={handleInstallPWA}
                  fullWidth
                >
                  📱 Install App
                </Button>
              )}
              
              {pwaInfo.notificationPermission === 'default' && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleEnableNotifications}
                  fullWidth
                >
                  🔔 Enable Notifications
                </Button>
              )}
            </div>
          </Card>
        )}

        {/* Welcome section */}
        <div style={{ marginBottom: '1.5rem' }}>
          <h2 style={{
            fontSize: '1.5rem',
            fontWeight: 'bold',
            color: theme.colors.text,
            marginBottom: '0.5rem'
          }}>
            Welcome back! 👋
          </h2>
          <p style={{
            color: theme.colors.textSecondary,
            fontSize: '0.875rem',
            margin: 0
          }}>
            Monitor your cluster on the go
          </p>
        </div>

        {/* Metrics grid */}
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(2, 1fr)',
          gap: '1rem',
          marginBottom: '2rem'
        }}>
          <MobileMetricCard
            title="Requests"
            value={dashboardData.totalRequests.toLocaleString()}
            trend={12}
            icon="📊"
            color={theme.colors.primary}
            onClick={() => setSelectedMetric('requests')}
          />
          <MobileMetricCard
            title="Response Time"
            value={`${dashboardData.avgResponseTime}ms`}
            trend={-5}
            icon="⚡"
            color={theme.colors.success}
            onClick={() => setSelectedMetric('response-time')}
          />
          <MobileMetricCard
            title="Error Rate"
            value={`${(dashboardData.errorRate * 100).toFixed(2)}%`}
            trend={-15}
            icon="🚨"
            color={dashboardData.errorRate > 0.05 ? theme.colors.error : theme.colors.warning}
            onClick={() => setSelectedMetric('error-rate')}
          />
          <MobileMetricCard
            title="Uptime"
            value={dashboardData.uptime}
            icon="🔄"
            color={theme.colors.info}
            onClick={() => setSelectedMetric('uptime')}
          />
        </div>

        {/* Nodes section */}
        <div>
          <h3 style={{
            fontSize: '1.125rem',
            fontWeight: '600',
            color: theme.colors.text,
            marginBottom: '1rem'
          }}>
            Cluster Nodes ({dashboardData.nodeCount})
          </h3>
          
          {dashboardData.nodes.map(node => (
            <MobileNodeCard key={node.id} node={node} />
          ))}
        </div>
      </div>

      {/* Bottom Navigation */}
      <BottomNavigation />

      {/* Mobile Menu Modal */}
      {showMobileMenu && (
        <Modal
          isOpen={showMobileMenu}
          onClose={() => setShowMobileMenu(false)}
          title="Menu"
          size="sm"
        >
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            <Button variant="outline" fullWidth>
              👤 Profile
            </Button>
            <Button variant="outline" fullWidth>
              ⚙️ Settings
            </Button>
            <Button variant="outline" fullWidth>
              📊 Analytics
            </Button>
            <Button variant="outline" fullWidth>
              ❓ Help
            </Button>
            <Button variant="danger" onClick={logout} fullWidth>
              🚪 Sign Out
            </Button>
          </div>
        </Modal>
      )}

      {/* Metric Detail Modal */}
      {selectedMetric && (
        <Modal
          isOpen={!!selectedMetric}
          onClose={() => setSelectedMetric(null)}
          title="Metric Details"
          size="sm"
        >
          <div style={{
            textAlign: 'center',
            padding: '1rem',
            color: theme.colors.textSecondary
          }}>
            Detailed metrics view will be available in the next update.
          </div>
        </Modal>
      )}
    </div>
  );
};

export default MobileDashboard;

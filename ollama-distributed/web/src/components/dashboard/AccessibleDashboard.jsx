/**
 * Accessible Dashboard Component
 * 
 * WCAG 2.1 AA compliant dashboard with comprehensive accessibility features.
 */

import React, { useState, useEffect, useRef } from 'react';
import { Card, Button, Badge, Modal, SkipLinks } from '../../design-system/index.js';
import { useTheme, useResponsive } from '../../design-system/theme/ThemeProvider.jsx';
import { useAuth } from '../../contexts/AuthContext.jsx';
import accessibilityService from '../../services/accessibilityService.js';
import i18nService from '../../services/i18nService.js';

const AccessibleDashboard = () => {
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
  
  const [sidebarOpen, setSidebarOpen] = useState(!isMobile);
  const [selectedNode, setSelectedNode] = useState(null);
  const [announcements, setAnnouncements] = useState([]);
  const mainContentRef = useRef(null);
  const navigationRef = useRef(null);

  // Accessibility announcements
  useEffect(() => {
    // Announce page load
    accessibilityService.announce(
      i18nService.t('dashboard.title') + ' ' + i18nService.t('a11y.loading')
    );

    // Setup accessibility event handlers
    accessibilityService.onReducedMotionChange = (enabled) => {
      if (enabled) {
        accessibilityService.announce('Reduced motion enabled');
      }
    };

    accessibilityService.onHighContrastChange = (enabled) => {
      if (enabled) {
        accessibilityService.announce('High contrast mode enabled');
      }
    };

    return () => {
      accessibilityService.destroy();
    };
  }, []);

  // Announce data updates
  useEffect(() => {
    const criticalNodes = dashboardData.nodes.filter(node => 
      node.status === 'error' || (node.cpu > 90 || node.memory > 90)
    );
    
    if (criticalNodes.length > 0) {
      accessibilityService.announce(
        `Alert: ${criticalNodes.length} nodes require attention`,
        'assertive'
      );
    }
  }, [dashboardData.nodes]);

  // Keyboard navigation
  useEffect(() => {
    const handleKeyDown = (event) => {
      // Alt + M: Focus main content
      if (event.altKey && event.key === 'm') {
        event.preventDefault();
        mainContentRef.current?.focus();
        accessibilityService.announce('Main content focused');
      }
      
      // Alt + N: Focus navigation
      if (event.altKey && event.key === 'n') {
        event.preventDefault();
        navigationRef.current?.focus();
        accessibilityService.announce('Navigation focused');
      }
      
      // Alt + S: Toggle sidebar
      if (event.altKey && event.key === 's') {
        event.preventDefault();
        setSidebarOpen(prev => !prev);
        accessibilityService.announce(
          sidebarOpen ? 'Sidebar collapsed' : 'Sidebar expanded'
        );
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [sidebarOpen]);

  // Layout styles
  const layoutStyles = {
    display: 'flex',
    minHeight: '100vh',
    backgroundColor: theme.colors.background
  };

  const sidebarStyles = {
    width: sidebarOpen ? '280px' : '80px',
    backgroundColor: theme.colors.surface,
    borderRight: `1px solid ${theme.colors.border}`,
    transition: accessibilityService.reducedMotion ? 'none' : 'width 0.3s ease',
    display: 'flex',
    flexDirection: 'column',
    position: 'fixed',
    height: '100vh',
    zIndex: 1000
  };

  const mainContentStyles = {
    flex: 1,
    marginLeft: sidebarOpen ? '280px' : '80px',
    transition: accessibilityService.reducedMotion ? 'none' : 'margin-left 0.3s ease',
    display: 'flex',
    flexDirection: 'column'
  };

  // Accessible metric card
  const AccessibleMetricCard = ({ 
    title, 
    value, 
    subtitle, 
    trend, 
    icon, 
    color = theme.colors.primary,
    onClick,
    id
  }) => {
    const cardId = `metric-${id}`;
    const descriptionId = `${cardId}-description`;
    
    return (
      <Card 
        variant="elevated" 
        hover 
        interactive={!!onClick}
        onClick={onClick}
        role={onClick ? 'button' : undefined}
        tabIndex={onClick ? 0 : undefined}
        aria-labelledby={cardId}
        aria-describedby={descriptionId}
        onKeyDown={(e) => {
          if (onClick && (e.key === 'Enter' || e.key === ' ')) {
            e.preventDefault();
            onClick();
          }
        }}
      >
        <div style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          marginBottom: '1rem'
        }}>
          <div 
            style={{
              width: '48px',
              height: '48px',
              backgroundColor: color + '15',
              borderRadius: '12px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: '1.5rem'
            }}
            aria-hidden="true"
          >
            {icon}
          </div>
          {trend !== undefined && (
            <Badge 
              variant={trend > 0 ? 'outline-success' : trend < 0 ? 'outline-error' : 'outline'}
              size="sm"
              aria-label={`Trend: ${trend > 0 ? 'increasing' : trend < 0 ? 'decreasing' : 'stable'} by ${Math.abs(trend)} percent`}
            >
              {trend > 0 ? '↗' : trend < 0 ? '↘' : '→'} {Math.abs(trend)}%
            </Badge>
          )}
        </div>
        
        <div>
          <div 
            id={cardId}
            style={{
              fontSize: '2rem',
              fontWeight: 'bold',
              color: theme.colors.text,
              marginBottom: '0.25rem'
            }}
          >
            {value}
          </div>
          <div style={{
            fontSize: '0.875rem',
            color: theme.colors.textSecondary,
            marginBottom: '0.5rem'
          }}>
            {title}
          </div>
          {subtitle && (
            <div 
              id={descriptionId}
              style={{
                fontSize: '0.75rem',
                color: theme.colors.textMuted
              }}
            >
              {subtitle}
            </div>
          )}
        </div>
      </Card>
    );
  };

  // Accessible navigation
  const AccessibleNavigation = () => (
    <nav 
      ref={navigationRef}
      role="navigation"
      aria-label="Main navigation"
      style={{ flex: 1, padding: '1rem 0' }}
      tabIndex={-1}
    >
      <ul style={{ listStyle: 'none', margin: 0, padding: 0 }}>
        {[
          { icon: '📊', label: i18nService.t('dashboard.title'), active: true, href: '#dashboard' },
          { icon: '🖥️', label: i18nService.t('dashboard.nodes'), active: false, href: '#nodes' },
          { icon: '🤖', label: i18nService.t('dashboard.models'), active: false, href: '#models' },
          { icon: '📈', label: i18nService.t('dashboard.metrics'), active: false, href: '#metrics' },
          { icon: '⚙️', label: i18nService.t('dashboard.settings'), active: false, href: '#settings' }
        ].map((item, index) => (
          <li key={index} style={{ margin: '0.25rem 1rem' }}>
            <a
              href={item.href}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '0.75rem',
                padding: '0.75rem',
                borderRadius: '8px',
                backgroundColor: item.active ? theme.colors.primary + '15' : 'transparent',
                color: item.active ? theme.colors.primary : theme.colors.textSecondary,
                textDecoration: 'none',
                fontWeight: item.active ? '600' : '400',
                transition: accessibilityService.reducedMotion ? 'none' : 'all 0.2s ease'
              }}
              aria-current={item.active ? 'page' : undefined}
              onFocus={(e) => {
                e.target.style.outline = `2px solid ${theme.colors.primary}`;
                e.target.style.outlineOffset = '2px';
              }}
              onBlur={(e) => {
                e.target.style.outline = 'none';
              }}
            >
              <span style={{ fontSize: '1.25rem' }} aria-hidden="true">
                {item.icon}
              </span>
              {sidebarOpen && (
                <span>{item.label}</span>
              )}
            </a>
          </li>
        ))}
      </ul>
    </nav>
  );

  return (
    <div style={layoutStyles}>
      {/* Skip Links */}
      <SkipLinks 
        links={[
          { href: '#main-content', text: i18nService.t('a11y.skipToMain') },
          { href: '#navigation', text: 'Skip to navigation' },
          { href: '#metrics', text: 'Skip to metrics' }
        ]}
      />

      {/* Sidebar */}
      <aside 
        style={sidebarStyles}
        role="complementary"
        aria-label="Sidebar navigation"
      >
        {/* Logo */}
        <div style={{
          padding: '1.5rem',
          borderBottom: `1px solid ${theme.colors.border}`,
          display: 'flex',
          alignItems: 'center',
          gap: '0.75rem'
        }}>
          <div style={{
            width: '32px',
            height: '32px',
            backgroundColor: theme.colors.primary,
            borderRadius: '8px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: 'white',
            fontWeight: 'bold',
            fontSize: '1.25rem'
          }}>
            O
          </div>
          {sidebarOpen && (
            <h1 style={{
              fontSize: '1.25rem',
              fontWeight: 'bold',
              color: theme.colors.text,
              margin: 0
            }}>
              OllamaMax
            </h1>
          )}
        </div>

        <AccessibleNavigation />

        {/* User section */}
        <div style={{
          padding: '1rem',
          borderTop: `1px solid ${theme.colors.border}`
        }}>
          <div style={{
            display: 'flex',
            alignItems: 'center',
            gap: '0.75rem',
            padding: '0.75rem',
            borderRadius: '8px',
            backgroundColor: theme.colors.surfaceVariant
          }}>
            <div 
              style={{
                width: '32px',
                height: '32px',
                backgroundColor: theme.colors.primary,
                borderRadius: '50%',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: 'white',
                fontWeight: 'bold'
              }}
              aria-hidden="true"
            >
              {user?.firstName?.[0] || 'U'}
            </div>
            {sidebarOpen && (
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{
                  fontSize: '0.875rem',
                  fontWeight: '600',
                  color: theme.colors.text,
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap'
                }}>
                  {user?.firstName} {user?.lastName}
                </div>
                <div style={{
                  fontSize: '0.75rem',
                  color: theme.colors.textSecondary,
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap'
                }}>
                  {user?.email}
                </div>
              </div>
            )}
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <main 
        ref={mainContentRef}
        id="main-content"
        style={mainContentStyles}
        tabIndex={-1}
        role="main"
        aria-label="Dashboard content"
      >
        {/* Header */}
        <header style={{
          height: '64px',
          backgroundColor: theme.colors.surface,
          borderBottom: `1px solid ${theme.colors.border}`,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '0 2rem',
          position: 'sticky',
          top: 0,
          zIndex: 999
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setSidebarOpen(!sidebarOpen)}
              aria-label={sidebarOpen ? 'Collapse sidebar' : 'Expand sidebar'}
              aria-expanded={sidebarOpen}
              aria-controls="sidebar"
            >
              <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                <path d="M3 18h18v-2H3v2zm0-5h18v-2H3v2zm0-7v2h18V6H3z"/>
              </svg>
            </Button>
            <h2 style={{
              fontSize: '1.5rem',
              fontWeight: '600',
              color: theme.colors.text,
              margin: 0
            }}>
              {i18nService.t('dashboard.title')}
            </h2>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
            <Button
              variant="ghost"
              size="sm"
              onClick={toggleTheme}
              aria-label={i18nService.t('a11y.toggleTheme')}
            >
              {currentTheme === 'light' ? '🌙' : '☀️'}
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={logout}
            >
              {i18nService.t('auth.logout')}
            </Button>
          </div>
        </header>

        {/* Content */}
        <div style={{ flex: 1, padding: '2rem', overflow: 'auto' }}>
          {/* Welcome section */}
          <section aria-labelledby="welcome-heading" style={{ marginBottom: '2rem' }}>
            <h2 
              id="welcome-heading"
              style={{
                fontSize: '1.875rem',
                fontWeight: 'bold',
                color: theme.colors.text,
                marginBottom: '0.5rem'
              }}
            >
              {i18nService.t('dashboard.welcome', { name: user?.firstName })}
            </h2>
            <p style={{
              color: theme.colors.textSecondary,
              fontSize: '1rem',
              margin: 0
            }}>
              Real-time monitoring of your OllamaMax distributed cluster.
            </p>
          </section>

          {/* Metrics section */}
          <section 
            id="metrics"
            aria-labelledby="metrics-heading"
            style={{ marginBottom: '2rem' }}
          >
            <h3 
              id="metrics-heading"
              style={{
                fontSize: '1.25rem',
                fontWeight: '600',
                color: theme.colors.text,
                marginBottom: '1rem'
              }}
            >
              System Metrics
            </h3>
            
            <div style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))',
              gap: '1.5rem'
            }}>
              <AccessibleMetricCard
                id="requests"
                title={i18nService.t('dashboard.totalRequests')}
                value={i18nService.formatNumber(dashboardData.totalRequests)}
                subtitle="Last 24 hours"
                trend={12}
                icon="📊"
                color={theme.colors.primary}
                onClick={() => accessibilityService.announce('Requests metric selected')}
              />
              <AccessibleMetricCard
                id="response-time"
                title={i18nService.t('dashboard.responseTime')}
                value={`${dashboardData.avgResponseTime}ms`}
                subtitle="95th percentile"
                trend={-5}
                icon="⚡"
                color={theme.colors.success}
                onClick={() => accessibilityService.announce('Response time metric selected')}
              />
              <AccessibleMetricCard
                id="error-rate"
                title={i18nService.t('dashboard.errorRate')}
                value={i18nService.formatPercentage(dashboardData.errorRate)}
                subtitle="Last hour"
                trend={-15}
                icon="🚨"
                color={dashboardData.errorRate > 0.05 ? theme.colors.error : theme.colors.warning}
                onClick={() => accessibilityService.announce('Error rate metric selected')}
              />
              <AccessibleMetricCard
                id="uptime"
                title={i18nService.t('dashboard.uptime')}
                value={dashboardData.uptime}
                subtitle="This month"
                icon="🔄"
                color={theme.colors.info}
                onClick={() => accessibilityService.announce('Uptime metric selected')}
              />
            </div>
          </section>

          {/* Nodes section */}
          <section aria-labelledby="nodes-heading">
            <h3 
              id="nodes-heading"
              style={{
                fontSize: '1.25rem',
                fontWeight: '600',
                color: theme.colors.text,
                marginBottom: '1rem'
              }}
            >
              {i18nService.t('dashboard.nodes')} ({dashboardData.nodeCount})
            </h3>
            
            <div style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))',
              gap: '1rem'
            }}>
              {dashboardData.nodes.map(node => (
                <Card 
                  key={node.id}
                  variant="elevated" 
                  hover 
                  interactive
                  role="button"
                  tabIndex={0}
                  aria-label={`Node ${node.id}, status: ${node.status}, CPU: ${node.cpu}%, Memory: ${node.memory}%`}
                  onClick={() => setSelectedNode(node)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' || e.key === ' ') {
                      e.preventDefault();
                      setSelectedNode(node);
                    }
                  }}
                >
                  <div style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    marginBottom: '1rem'
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
                      aria-label={`Status: ${node.status}`}
                    >
                      {node.status}
                    </Badge>
                  </div>
                  
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                    <div>
                      <div style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                        fontSize: '0.875rem',
                        color: theme.colors.textSecondary,
                        marginBottom: '0.25rem'
                      }}>
                        <span>CPU</span>
                        <span>{node.cpu}%</span>
                      </div>
                      <div 
                        style={{
                          width: '100%',
                          height: '6px',
                          backgroundColor: theme.colors.border,
                          borderRadius: '3px',
                          overflow: 'hidden'
                        }}
                        role="progressbar"
                        aria-valuenow={node.cpu}
                        aria-valuemin={0}
                        aria-valuemax={100}
                        aria-label={`CPU usage: ${node.cpu} percent`}
                      >
                        <div style={{
                          width: `${node.cpu}%`,
                          height: '100%',
                          backgroundColor: node.cpu > 80 ? theme.colors.error : node.cpu > 60 ? theme.colors.warning : theme.colors.success,
                          transition: accessibilityService.reducedMotion ? 'none' : 'width 0.3s ease'
                        }} />
                      </div>
                    </div>
                    
                    <div>
                      <div style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                        fontSize: '0.875rem',
                        color: theme.colors.textSecondary,
                        marginBottom: '0.25rem'
                      }}>
                        <span>Memory</span>
                        <span>{node.memory}%</span>
                      </div>
                      <div 
                        style={{
                          width: '100%',
                          height: '6px',
                          backgroundColor: theme.colors.border,
                          borderRadius: '3px',
                          overflow: 'hidden'
                        }}
                        role="progressbar"
                        aria-valuenow={node.memory}
                        aria-valuemin={0}
                        aria-valuemax={100}
                        aria-label={`Memory usage: ${node.memory} percent`}
                      >
                        <div style={{
                          width: `${node.memory}%`,
                          height: '100%',
                          backgroundColor: node.memory > 80 ? theme.colors.error : node.memory > 60 ? theme.colors.warning : theme.colors.success,
                          transition: accessibilityService.reducedMotion ? 'none' : 'width 0.3s ease'
                        }} />
                      </div>
                    </div>
                    
                    <div style={{
                      fontSize: '0.875rem',
                      color: theme.colors.textSecondary
                    }}>
                      Requests: {i18nService.formatNumber(node.requests)}
                    </div>
                  </div>
                </Card>
              ))}
            </div>
          </section>
        </div>
      </main>

      {/* Node details modal */}
      {selectedNode && (
        <Modal
          isOpen={!!selectedNode}
          onClose={() => setSelectedNode(null)}
          title={`Node Details - ${selectedNode.id}`}
          size="lg"
          aria-describedby="node-details-description"
        >
          <div id="node-details-description" style={{ display: 'grid', gap: '1.5rem' }}>
            <div style={{
              display: 'grid',
              gridTemplateColumns: '1fr 1fr',
              gap: '1rem'
            }}>
              <div>
                <h4 style={{ margin: '0 0 0.5rem 0', color: theme.colors.text }}>Status</h4>
                <Badge variant={selectedNode.status === 'healthy' ? 'success' : selectedNode.status === 'warning' ? 'warning' : 'error'}>
                  {selectedNode.status}
                </Badge>
              </div>
              <div>
                <h4 style={{ margin: '0 0 0.5rem 0', color: theme.colors.text }}>Requests</h4>
                <p style={{ margin: 0, color: theme.colors.textSecondary }}>
                  {i18nService.formatNumber(selectedNode.requests)}
                </p>
              </div>
            </div>
            
            <div>
              <h4 style={{ margin: '0 0 1rem 0', color: theme.colors.text }}>Resource Usage</h4>
              <div style={{ display: 'grid', gap: '1rem' }}>
                <div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
                    <span style={{ color: theme.colors.textSecondary }}>CPU Usage</span>
                    <span style={{ color: theme.colors.text, fontWeight: '600' }}>{selectedNode.cpu}%</span>
                  </div>
                  <div 
                    style={{
                      width: '100%',
                      height: '8px',
                      backgroundColor: theme.colors.border,
                      borderRadius: '4px',
                      overflow: 'hidden'
                    }}
                    role="progressbar"
                    aria-valuenow={selectedNode.cpu}
                    aria-valuemin={0}
                    aria-valuemax={100}
                    aria-label={`CPU usage: ${selectedNode.cpu} percent`}
                  >
                    <div style={{
                      width: `${selectedNode.cpu}%`,
                      height: '100%',
                      backgroundColor: selectedNode.cpu > 80 ? theme.colors.error : selectedNode.cpu > 60 ? theme.colors.warning : theme.colors.success,
                      transition: accessibilityService.reducedMotion ? 'none' : 'width 0.3s ease'
                    }} />
                  </div>
                </div>
                
                <div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
                    <span style={{ color: theme.colors.textSecondary }}>Memory Usage</span>
                    <span style={{ color: theme.colors.text, fontWeight: '600' }}>{selectedNode.memory}%</span>
                  </div>
                  <div 
                    style={{
                      width: '100%',
                      height: '8px',
                      backgroundColor: theme.colors.border,
                      borderRadius: '4px',
                      overflow: 'hidden'
                    }}
                    role="progressbar"
                    aria-valuenow={selectedNode.memory}
                    aria-valuemin={0}
                    aria-valuemax={100}
                    aria-label={`Memory usage: ${selectedNode.memory} percent`}
                  >
                    <div style={{
                      width: `${selectedNode.memory}%`,
                      height: '100%',
                      backgroundColor: selectedNode.memory > 80 ? theme.colors.error : selectedNode.memory > 60 ? theme.colors.warning : theme.colors.success,
                      transition: accessibilityService.reducedMotion ? 'none' : 'width 0.3s ease'
                    }} />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </Modal>
      )}
    </div>
  );
};

export default AccessibleDashboard;

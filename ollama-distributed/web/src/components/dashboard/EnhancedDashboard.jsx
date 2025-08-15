/**
 * Enhanced Dashboard Component
 * 
 * Modern dashboard implementation using the OllamaMax design system.
 */

import React, { useState, useEffect } from 'react';
import { Card, Button } from '../../design-system/index.js';
import { useTheme } from '../../design-system/theme/ThemeProvider.jsx';
import { useAuth } from '../../contexts/AuthContext.jsx';

const EnhancedDashboard = () => {
  const { theme, toggleTheme, currentTheme } = useTheme();
  const { user, logout } = useAuth();
  const [dashboardData, setDashboardData] = useState({
    clusterStatus: 'healthy',
    nodeCount: 3,
    activeModels: 5,
    totalRequests: 1247,
    avgResponseTime: 245,
    errorRate: 0.02,
    uptime: '99.9%'
  });
  const [sidebarOpen, setSidebarOpen] = useState(true);

  // Simulate real-time data updates
  useEffect(() => {
    const interval = setInterval(() => {
      setDashboardData(prev => ({
        ...prev,
        totalRequests: prev.totalRequests + Math.floor(Math.random() * 5),
        avgResponseTime: 200 + Math.floor(Math.random() * 100),
        errorRate: Math.max(0, Math.random() * 0.05)
      }));
    }, 5000);

    return () => clearInterval(interval);
  }, []);

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
    transition: 'width 0.3s ease',
    display: 'flex',
    flexDirection: 'column',
    position: 'fixed',
    height: '100vh',
    zIndex: 1000
  };

  const mainContentStyles = {
    flex: 1,
    marginLeft: sidebarOpen ? '280px' : '80px',
    transition: 'margin-left 0.3s ease',
    display: 'flex',
    flexDirection: 'column'
  };

  const headerStyles = {
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
  };

  const contentStyles = {
    flex: 1,
    padding: '2rem',
    overflow: 'auto'
  };

  // Sidebar component
  const Sidebar = () => (
    <div style={sidebarStyles}>
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
          <span style={{
            fontSize: '1.25rem',
            fontWeight: 'bold',
            color: theme.colors.text
          }}>
            OllamaMax
          </span>
        )}
      </div>

      {/* Navigation */}
      <nav style={{ flex: 1, padding: '1rem 0' }}>
        {[
          { icon: '📊', label: 'Dashboard', active: true },
          { icon: '🖥️', label: 'Nodes', active: false },
          { icon: '🤖', label: 'Models', active: false },
          { icon: '📈', label: 'Metrics', active: false },
          { icon: '⚙️', label: 'Settings', active: false },
          { icon: '🔧', label: 'Fault Tolerance', active: false }
        ].map((item, index) => (
          <div
            key={index}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.75rem',
              padding: '0.75rem 1.5rem',
              margin: '0.25rem 1rem',
              borderRadius: '8px',
              backgroundColor: item.active ? theme.colors.primary + '15' : 'transparent',
              color: item.active ? theme.colors.primary : theme.colors.textSecondary,
              cursor: 'pointer',
              transition: 'all 0.2s ease'
            }}
          >
            <span style={{ fontSize: '1.25rem' }}>{item.icon}</span>
            {sidebarOpen && (
              <span style={{ fontWeight: item.active ? '600' : '400' }}>
                {item.label}
              </span>
            )}
          </div>
        ))}
      </nav>

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
          <div style={{
            width: '32px',
            height: '32px',
            backgroundColor: theme.colors.primary,
            borderRadius: '50%',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: 'white',
            fontWeight: 'bold'
          }}>
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
    </div>
  );

  // Header component
  const Header = () => (
    <div style={headerStyles}>
      <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
        <Button
          variant="ghost"
          size="sm"
          onClick={() => setSidebarOpen(!sidebarOpen)}
          aria-label="Toggle sidebar"
        >
          <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
            <path d="M3 18h18v-2H3v2zm0-5h18v-2H3v2zm0-7v2h18V6H3z"/>
          </svg>
        </Button>
        <h1 style={{
          fontSize: '1.5rem',
          fontWeight: '600',
          color: theme.colors.text,
          margin: 0
        }}>
          Dashboard
        </h1>
      </div>

      <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
        <Button
          variant="ghost"
          size="sm"
          onClick={toggleTheme}
          aria-label="Toggle theme"
        >
          {currentTheme === 'light' ? '🌙' : '☀️'}
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={logout}
        >
          Sign Out
        </Button>
      </div>
    </div>
  );

  // Metric card component
  const MetricCard = ({ title, value, subtitle, trend, icon, color = theme.colors.primary }) => (
    <Card variant="elevated" hover>
      <div style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        marginBottom: '1rem'
      }}>
        <div style={{
          width: '48px',
          height: '48px',
          backgroundColor: color + '15',
          borderRadius: '12px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          fontSize: '1.5rem'
        }}>
          {icon}
        </div>
        {trend && (
          <div style={{
            fontSize: '0.875rem',
            color: trend > 0 ? theme.colors.success : theme.colors.error,
            fontWeight: '600'
          }}>
            {trend > 0 ? '↗' : '↘'} {Math.abs(trend)}%
          </div>
        )}
      </div>
      <div style={{
        fontSize: '2rem',
        fontWeight: 'bold',
        color: theme.colors.text,
        marginBottom: '0.25rem'
      }}>
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
        <div style={{
          fontSize: '0.75rem',
          color: theme.colors.textMuted
        }}>
          {subtitle}
        </div>
      )}
    </Card>
  );

  // Status indicator component
  const StatusIndicator = ({ status, label }) => {
    const statusColors = {
      healthy: theme.colors.success,
      warning: theme.colors.warning,
      error: theme.colors.error
    };

    return (
      <div style={{
        display: 'flex',
        alignItems: 'center',
        gap: '0.5rem'
      }}>
        <div style={{
          width: '8px',
          height: '8px',
          borderRadius: '50%',
          backgroundColor: statusColors[status] || theme.colors.textMuted
        }} />
        <span style={{
          fontSize: '0.875rem',
          color: theme.colors.textSecondary
        }}>
          {label}
        </span>
      </div>
    );
  };

  return (
    <div style={layoutStyles}>
      <Sidebar />
      <div style={mainContentStyles}>
        <Header />
        <div style={contentStyles}>
          {/* Welcome section */}
          <div style={{ marginBottom: '2rem' }}>
            <h2 style={{
              fontSize: '1.875rem',
              fontWeight: 'bold',
              color: theme.colors.text,
              marginBottom: '0.5rem'
            }}>
              Welcome back, {user?.firstName}! 👋
            </h2>
            <p style={{
              color: theme.colors.textSecondary,
              fontSize: '1rem'
            }}>
              Here's what's happening with your OllamaMax cluster today.
            </p>
          </div>

          {/* Metrics grid */}
          <div style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))',
            gap: '1.5rem',
            marginBottom: '2rem'
          }}>
            <MetricCard
              title="Total Requests"
              value={dashboardData.totalRequests.toLocaleString()}
              subtitle="Last 24 hours"
              trend={12}
              icon="📊"
              color={theme.colors.primary}
            />
            <MetricCard
              title="Average Response Time"
              value={`${dashboardData.avgResponseTime}ms`}
              subtitle="95th percentile"
              trend={-5}
              icon="⚡"
              color={theme.colors.success}
            />
            <MetricCard
              title="Error Rate"
              value={`${(dashboardData.errorRate * 100).toFixed(2)}%`}
              subtitle="Last hour"
              trend={-15}
              icon="🚨"
              color={theme.colors.warning}
            />
            <MetricCard
              title="Uptime"
              value={dashboardData.uptime}
              subtitle="This month"
              icon="🔄"
              color={theme.colors.info}
            />
          </div>

          {/* Status cards */}
          <div style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))',
            gap: '1.5rem'
          }}>
            <Card variant="elevated">
              <Card.Header>
                <Card.Title size="sm">Cluster Status</Card.Title>
              </Card.Header>
              <Card.Body>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                  <StatusIndicator status="healthy" label="All systems operational" />
                  <StatusIndicator status="healthy" label={`${dashboardData.nodeCount} nodes active`} />
                  <StatusIndicator status="healthy" label={`${dashboardData.activeModels} models loaded`} />
                </div>
              </Card.Body>
            </Card>

            <Card variant="elevated">
              <Card.Header>
                <Card.Title size="sm">Fault Tolerance</Card.Title>
              </Card.Header>
              <Card.Body>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                  <StatusIndicator status="healthy" label="Predictive detection active" />
                  <StatusIndicator status="healthy" label="Self-healing enabled" />
                  <StatusIndicator status="healthy" label="Redundancy optimal" />
                </div>
              </Card.Body>
            </Card>

            <Card variant="elevated">
              <Card.Header>
                <Card.Title size="sm">Recent Activity</Card.Title>
              </Card.Header>
              <Card.Body>
                <div style={{
                  fontSize: '0.875rem',
                  color: theme.colors.textSecondary,
                  textAlign: 'center',
                  padding: '2rem'
                }}>
                  Activity feed will be available in the next iteration.
                </div>
              </Card.Body>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
};

export default EnhancedDashboard;

/**
 * Real-Time Dashboard Component
 * 
 * Enhanced dashboard with real-time data, interactive charts, and improved UX.
 */

import React, { useState, useEffect, useRef } from 'react';
import { Card, Button, Badge, Modal, Toast, ToastContainer } from '../../design-system/index.js';
import { useTheme } from '../../design-system/theme/ThemeProvider.jsx';
import { useAuth } from '../../contexts/AuthContext.jsx';

const RealTimeDashboard = () => {
  const { theme, toggleTheme, currentTheme } = useTheme();
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
    ],
    recentEvents: [
      { id: 1, type: 'info', message: 'Node node-3 CPU usage high', timestamp: new Date() },
      { id: 2, type: 'success', message: 'Healing attempt successful on node-2', timestamp: new Date(Date.now() - 300000) },
      { id: 3, type: 'warning', message: 'Predictive detection triggered', timestamp: new Date(Date.now() - 600000) }
    ]
  });
  
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [selectedNode, setSelectedNode] = useState(null);
  const [toasts, setToasts] = useState([]);
  const [isConnected, setIsConnected] = useState(true);
  const wsRef = useRef(null);

  // WebSocket connection for real-time updates
  useEffect(() => {
    const connectWebSocket = () => {
      try {
        wsRef.current = new WebSocket('ws://localhost:8080/ws');
        
        wsRef.current.onopen = () => {
          setIsConnected(true);
          addToast('success', 'Connected', 'Real-time updates enabled');
        };
        
        wsRef.current.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data);
            handleWebSocketMessage(data);
          } catch (error) {
            console.error('Error parsing WebSocket message:', error);
          }
        };
        
        wsRef.current.onclose = () => {
          setIsConnected(false);
          addToast('warning', 'Disconnected', 'Real-time updates paused');
          
          // Attempt to reconnect after 5 seconds
          setTimeout(connectWebSocket, 5000);
        };
        
        wsRef.current.onerror = (error) => {
          console.error('WebSocket error:', error);
          setIsConnected(false);
        };
      } catch (error) {
        console.error('Failed to connect WebSocket:', error);
        setIsConnected(false);
      }
    };

    connectWebSocket();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  // Handle WebSocket messages
  const handleWebSocketMessage = (data) => {
    switch (data.type) {
      case 'metrics_update':
        setDashboardData(prev => ({
          ...prev,
          ...data.payload
        }));
        break;
      
      case 'node_status_change':
        setDashboardData(prev => ({
          ...prev,
          nodes: prev.nodes.map(node => 
            node.id === data.payload.nodeId 
              ? { ...node, ...data.payload.status }
              : node
          )
        }));
        
        // Show toast for status changes
        if (data.payload.status.status !== 'healthy') {
          addToast('warning', 'Node Alert', `${data.payload.nodeId} status changed to ${data.payload.status.status}`);
        }
        break;
      
      case 'fault_tolerance_event':
        const event = {
          id: Date.now(),
          type: data.payload.severity || 'info',
          message: data.payload.message,
          timestamp: new Date()
        };
        
        setDashboardData(prev => ({
          ...prev,
          recentEvents: [event, ...prev.recentEvents.slice(0, 9)]
        }));
        
        // Show toast for critical events
        if (data.payload.severity === 'error' || data.payload.severity === 'warning') {
          addToast(data.payload.severity, 'System Alert', data.payload.message);
        }
        break;
      
      default:
        console.log('Unknown WebSocket message type:', data.type);
    }
  };

  // Add toast notification
  const addToast = (type, title, message) => {
    const toast = {
      id: Date.now().toString(),
      type,
      title,
      message,
      onClose: removeToast
    };
    
    setToasts(prev => [...prev, toast]);
  };

  // Remove toast notification
  const removeToast = (id) => {
    setToasts(prev => prev.filter(toast => toast.id !== id));
  };

  // Simulate data updates when WebSocket is not available
  useEffect(() => {
    if (!isConnected) {
      const interval = setInterval(() => {
        setDashboardData(prev => ({
          ...prev,
          totalRequests: prev.totalRequests + Math.floor(Math.random() * 5),
          avgResponseTime: 200 + Math.floor(Math.random() * 100),
          errorRate: Math.max(0, Math.random() * 0.05),
          nodes: prev.nodes.map(node => ({
            ...node,
            cpu: Math.max(0, Math.min(100, node.cpu + (Math.random() - 0.5) * 10)),
            memory: Math.max(0, Math.min(100, node.memory + (Math.random() - 0.5) * 5)),
            requests: node.requests + Math.floor(Math.random() * 3)
          }))
        }));
      }, 3000);

      return () => clearInterval(interval);
    }
  }, [isConnected]);

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

  // Enhanced metric card with real-time updates
  const MetricCard = ({ title, value, subtitle, trend, icon, color = theme.colors.primary, onClick }) => {
    const [isAnimating, setIsAnimating] = useState(false);
    const prevValueRef = useRef(value);

    useEffect(() => {
      if (prevValueRef.current !== value) {
        setIsAnimating(true);
        const timer = setTimeout(() => setIsAnimating(false), 300);
        prevValueRef.current = value;
        return () => clearTimeout(timer);
      }
    }, [value]);

    return (
      <Card 
        variant="elevated" 
        hover 
        interactive={!!onClick}
        onClick={onClick}
        style={{
          transform: isAnimating ? 'scale(1.02)' : 'scale(1)',
          transition: 'transform 0.3s ease'
        }}
      >
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
          {trend !== undefined && (
            <Badge 
              variant={trend > 0 ? 'outline-success' : trend < 0 ? 'outline-error' : 'outline'}
              size="sm"
            >
              {trend > 0 ? '↗' : trend < 0 ? '↘' : '→'} {Math.abs(trend)}%
            </Badge>
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
  };

  // Node status card
  const NodeCard = ({ node }) => {
    const getStatusColor = (status) => {
      switch (status) {
        case 'healthy': return theme.colors.success;
        case 'warning': return theme.colors.warning;
        case 'error': return theme.colors.error;
        default: return theme.colors.textMuted;
      }
    };

    return (
      <Card 
        variant="elevated" 
        hover 
        interactive
        onClick={() => setSelectedNode(node)}
      >
        <div style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          marginBottom: '1rem'
        }}>
          <h3 style={{
            fontSize: '1rem',
            fontWeight: '600',
            color: theme.colors.text,
            margin: 0
          }}>
            {node.id}
          </h3>
          <Badge 
            variant={node.status === 'healthy' ? 'success' : node.status === 'warning' ? 'warning' : 'error'}
            size="sm"
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
            <div style={{
              width: '100%',
              height: '6px',
              backgroundColor: theme.colors.border,
              borderRadius: '3px',
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
              display: 'flex',
              justifyContent: 'space-between',
              fontSize: '0.875rem',
              color: theme.colors.textSecondary,
              marginBottom: '0.25rem'
            }}>
              <span>Memory</span>
              <span>{node.memory}%</span>
            </div>
            <div style={{
              width: '100%',
              height: '6px',
              backgroundColor: theme.colors.border,
              borderRadius: '3px',
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
          
          <div style={{
            fontSize: '0.875rem',
            color: theme.colors.textSecondary
          }}>
            Requests: {node.requests.toLocaleString()}
          </div>
        </div>
      </Card>
    );
  };

  // Connection status indicator
  const ConnectionStatus = () => (
    <div style={{
      display: 'flex',
      alignItems: 'center',
      gap: '0.5rem',
      padding: '0.5rem 1rem',
      backgroundColor: isConnected ? theme.colors.success + '15' : theme.colors.warning + '15',
      borderRadius: '20px',
      fontSize: '0.875rem'
    }}>
      <div style={{
        width: '8px',
        height: '8px',
        borderRadius: '50%',
        backgroundColor: isConnected ? theme.colors.success : theme.colors.warning
      }} />
      <span style={{
        color: isConnected ? theme.colors.success : theme.colors.warning,
        fontWeight: '500'
      }}>
        {isConnected ? 'Live' : 'Offline'}
      </span>
    </div>
  );

  return (
    <div style={layoutStyles}>
      {/* Sidebar - reusing from previous iteration */}
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

      <div style={mainContentStyles}>
        {/* Header */}
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
              Real-Time Dashboard
            </h1>
            <ConnectionStatus />
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
              Real-time monitoring of your OllamaMax distributed cluster.
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
              color={dashboardData.errorRate > 0.05 ? theme.colors.error : theme.colors.warning}
            />
            <MetricCard
              title="Uptime"
              value={dashboardData.uptime}
              subtitle="This month"
              icon="🔄"
              color={theme.colors.info}
            />
          </div>

          {/* Nodes and Events grid */}
          <div style={{
            display: 'grid',
            gridTemplateColumns: '2fr 1fr',
            gap: '2rem',
            marginBottom: '2rem'
          }}>
            {/* Nodes section */}
            <div>
              <h3 style={{
                fontSize: '1.25rem',
                fontWeight: '600',
                color: theme.colors.text,
                marginBottom: '1rem'
              }}>
                Cluster Nodes
              </h3>
              <div style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
                gap: '1rem'
              }}>
                {dashboardData.nodes.map(node => (
                  <NodeCard key={node.id} node={node} />
                ))}
              </div>
            </div>

            {/* Recent events */}
            <div>
              <h3 style={{
                fontSize: '1.25rem',
                fontWeight: '600',
                color: theme.colors.text,
                marginBottom: '1rem'
              }}>
                Recent Events
              </h3>
              <Card variant="elevated">
                <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                  {dashboardData.recentEvents.map(event => (
                    <div key={event.id} style={{
                      display: 'flex',
                      alignItems: 'flex-start',
                      gap: '0.75rem',
                      padding: '0.75rem',
                      borderRadius: '8px',
                      backgroundColor: theme.colors.surfaceVariant
                    }}>
                      <Badge 
                        variant={event.type === 'success' ? 'success' : event.type === 'warning' ? 'warning' : event.type === 'error' ? 'error' : 'info'}
                        dot
                        size="sm"
                      />
                      <div style={{ flex: 1 }}>
                        <div style={{
                          fontSize: '0.875rem',
                          color: theme.colors.text,
                          marginBottom: '0.25rem'
                        }}>
                          {event.message}
                        </div>
                        <div style={{
                          fontSize: '0.75rem',
                          color: theme.colors.textMuted
                        }}>
                          {event.timestamp.toLocaleTimeString()}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </Card>
            </div>
          </div>
        </div>
      </div>

      {/* Node details modal */}
      {selectedNode && (
        <Modal
          isOpen={!!selectedNode}
          onClose={() => setSelectedNode(null)}
          title={`Node Details - ${selectedNode.id}`}
          size="lg"
        >
          <div style={{ display: 'grid', gap: '1.5rem' }}>
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
                <p style={{ margin: 0, color: theme.colors.textSecondary }}>{selectedNode.requests.toLocaleString()}</p>
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
                  <div style={{
                    width: '100%',
                    height: '8px',
                    backgroundColor: theme.colors.border,
                    borderRadius: '4px',
                    overflow: 'hidden'
                  }}>
                    <div style={{
                      width: `${selectedNode.cpu}%`,
                      height: '100%',
                      backgroundColor: selectedNode.cpu > 80 ? theme.colors.error : selectedNode.cpu > 60 ? theme.colors.warning : theme.colors.success,
                      transition: 'width 0.3s ease'
                    }} />
                  </div>
                </div>
                
                <div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
                    <span style={{ color: theme.colors.textSecondary }}>Memory Usage</span>
                    <span style={{ color: theme.colors.text, fontWeight: '600' }}>{selectedNode.memory}%</span>
                  </div>
                  <div style={{
                    width: '100%',
                    height: '8px',
                    backgroundColor: theme.colors.border,
                    borderRadius: '4px',
                    overflow: 'hidden'
                  }}>
                    <div style={{
                      width: `${selectedNode.memory}%`,
                      height: '100%',
                      backgroundColor: selectedNode.memory > 80 ? theme.colors.error : selectedNode.memory > 60 ? theme.colors.warning : theme.colors.success,
                      transition: 'width 0.3s ease'
                    }} />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </Modal>
      )}

      {/* Toast notifications */}
      <ToastContainer toasts={toasts} position="top-right" />
    </div>
  );
};

export default RealTimeDashboard;

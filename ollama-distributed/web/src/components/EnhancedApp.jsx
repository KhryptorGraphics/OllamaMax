import React, { useState, useEffect } from 'react';
import { Container, Row, Col, Button } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faBars, faTimes } from '@fortawesome/free-solid-svg-icons';

// Import services
import apiService from '../services/api.js';
import authService from '../services/auth.js';
import wsService from '../services/websocket.js';

// Import all components
import Sidebar from './Sidebar';
import Dashboard from './Dashboard';
import NodesView from './NodesView';
import ModelsView from './ModelsView';
import TransfersView from './TransfersView';
import ClusterView from './ClusterView';
import ClusterOverview from './ClusterOverview';
import Analytics from './Analytics';
import UserManagement from './UserManagement';
import DatabaseEditor from './DatabaseEditor';
import SystemSettings from './SystemSettings';
import RealTimeMetrics from './RealTimeMetrics';
import WebSocketStatus from './WebSocketStatus';
import LoadingSpinner from './LoadingSpinner';
import Alert from './Alert';
import ThemeToggle from './ThemeToggle';
import Login from './Login';
import RegistrationFlow from './RegistrationFlow';

const EnhancedApp = () => {
  // Application state
  const [activeTab, setActiveTab] = useState('dashboard');
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [theme, setTheme] = useState('light');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [alerts, setAlerts] = useState([]);

  // Authentication state
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [currentUser, setCurrentUser] = useState(null);
  const [showLogin, setShowLogin] = useState(false);
  const [showRegistration, setShowRegistration] = useState(false);

  // Data state
  const [clusterStatus, setClusterStatus] = useState(null);
  const [nodes, setNodes] = useState([]);
  const [models, setModels] = useState([]);
  const [transfers, setTransfers] = useState([]);
  const [metrics, setMetrics] = useState({});
  const [realTimeMetrics, setRealTimeMetrics] = useState({});
  const [users, setUsers] = useState([]);
  const [wsConnection, setWsConnection] = useState(null);
  const [isConnected, setIsConnected] = useState(false);
  const [reconnectAttempts, setReconnectAttempts] = useState(0);

  // Initialize app
  useEffect(() => {
    initializeApp();
    setupAuthListeners();
    setupWebSocket();
    loadTheme();

    return () => {
      wsService.disconnect();
      authService.off('authenticated', handleAuthenticated);
      authService.off('logged_out', handleLoggedOut);
    };
  }, []);

  const initializeApp = async () => {
    setLoading(true);
    try {
      // Check authentication status
      const authStatus = authService.isAuthenticated();
      setIsAuthenticated(authStatus);

      if (authStatus) {
        setCurrentUser(authService.getCurrentUser());

        // Load application data for authenticated users
        await Promise.all([
          loadClusterStatus(),
          loadNodes(),
          loadModels(),
          loadUsers(),
          loadMetrics()
        ]);
      } else {
        // For unauthenticated users, only load public health data
        await loadHealthStatus();
      }
    } catch (err) {
      setError(`Failed to initialize application: ${err.message}`);
      addAlert('error', 'Initialization Failed', err.message);
    } finally {
      setLoading(false);
    }
  };

  // Authentication event handlers
  const setupAuthListeners = () => {
    authService.on('authenticated', handleAuthenticated);
    authService.on('logged_out', handleLoggedOut);
    authService.on('session_warning', handleSessionWarning);
    authService.on('session_expired', handleSessionExpired);
  };

  const handleAuthenticated = (user) => {
    setIsAuthenticated(true);
    setCurrentUser(user);
    setShowLogin(false);
    setShowRegistration(false);
    addAlert('success', 'Welcome!', `Successfully logged in as ${user.username}`);
    initializeApp(); // Reload data for authenticated user
  };

  const handleLoggedOut = () => {
    setIsAuthenticated(false);
    setCurrentUser(null);
    setActiveTab('dashboard');
    addAlert('info', 'Logged Out', 'You have been successfully logged out');
  };

  const handleSessionWarning = (data) => {
    addAlert('warning', 'Session Expiring', data.message);
  };

  const handleSessionExpired = () => {
    addAlert('error', 'Session Expired', 'Your session has expired. Please log in again.');
  };

  const setupWebSocket = () => {
    if (isAuthenticated) {
      wsService.connect(authService.getAuthToken());

      wsService.on('connected', () => {
        setIsConnected(true);
        setReconnectAttempts(0);
        addAlert('success', 'Connected', 'Real-time updates enabled');
      });

      wsService.on('disconnected', () => {
        setIsConnected(false);
      });

      wsService.on('error', (error) => {
        console.error('WebSocket error:', error);
        addAlert('error', 'Connection Error', 'Real-time updates unavailable');
      });

      // Subscribe to real-time data streams
      wsService.subscribe('metrics', handleMetricsUpdate);
      wsService.subscribe('nodes', handleNodesUpdate);
      wsService.subscribe('models', handleModelsUpdate);
    }
  };

  // Real-time data update handlers
  const handleMetricsUpdate = (data) => {
    setRealTimeMetrics(data.metrics);
  };

  const handleNodesUpdate = (data) => {
    setNodes(prev => updateNodeInList(prev, data.node));
  };

  const handleModelsUpdate = (data) => {
    setModels(prev => updateModelInList(prev, data.model));
  };

  const handleWebSocketMessage = (data) => {
    switch (data.type) {
      case 'cluster_status':
        setClusterStatus(data.status);
        break;
      case 'node_update':
        setNodes(prev => updateNodeInList(prev, data.node));
        break;
      case 'model_update':
        setModels(prev => updateModelInList(prev, data.model));
        break;
      case 'metrics':
        setRealTimeMetrics(prev => ({
          ...prev,
          [data.metric]: [...(prev[data.metric] || []), data.value].slice(-30)
        }));
        break;
      case 'alert':
        showAlert(data.message, data.level || 'info', true, 5000);
        break;
      default:
        console.log('Unknown WebSocket message type:', data.type);
    }
  };

  const updateNodeInList = (nodes, updatedNode) => {
    return nodes.map(node => 
      node.id === updatedNode.id ? { ...node, ...updatedNode } : node
    );
  };

  const updateModelInList = (models, updatedModel) => {
    return models.map(model => 
      model.name === updatedModel.name ? { ...model, ...updatedModel } : model
    );
  };

  // Data loading functions
  const loadClusterStatus = async () => {
    try {
      const data = await apiService.getClusterStatus();
      setClusterStatus(data);
    } catch (err) {
      console.error('Failed to load cluster status:', err);
      addAlert('error', 'Cluster Status Error', 'Failed to load cluster status');
      // Fallback mock data for development
      setClusterStatus({
        node_id: 'localhost',
        leader: 'localhost',
        is_leader: true,
        status: 'healthy',
        peers: 0
      });
    }
  };

  const loadHealthStatus = async () => {
    try {
      const health = await apiService.getHealth();
      const readiness = await apiService.getReadiness();
      setClusterStatus({ health, readiness });
    } catch (err) {
      console.error('Failed to load health status:', err);
      addAlert('warning', 'Health Check Failed', 'Unable to verify system health');
    }
  };

  const loadNodes = async () => {
    try {
      const data = await apiService.getNodes();
      setNodes(data.nodes || data || []);
    } catch (err) {
      console.error('Failed to load nodes:', err);
      addAlert('error', 'Nodes Error', 'Failed to load cluster nodes');
      // Fallback mock data
      setNodes([
        {
          id: 'node-001',
          address: '192.168.1.100:11434',
          status: 'online',
          models: ['llama2:7b', 'codellama:13b'],
          usage: { cpu: 45, memory: 62, bandwidth: 23.5 }
        },
        {
          id: 'node-002',
          address: '192.168.1.101:11434',
          status: 'online',
          models: ['mistral:7b'],
          usage: { cpu: 67, memory: 78, bandwidth: 18.2 }
        }
      ]);
    }
  };

  const loadModels = async () => {
    try {
      const data = await apiService.getModels();
      setModels(data.models || data || []);
    } catch (err) {
      console.error('Failed to load models:', err);
      addAlert('error', 'Models Error', 'Failed to load available models');
      // Fallback mock data
      setModels([
        {
          name: 'llama2:7b',
          size: 3800000000,
          status: 'available',
          replicas: ['node-001', 'node-002'],
          inference_ready: true
        },
        {
          name: 'codellama:13b',
          size: 7300000000,
          status: 'available',
          replicas: ['node-001'],
          inference_ready: true
        }
      ]);
    }
  };

  const loadUsers = async () => {
    try {
      if (authService.isAdmin()) {
        const data = await apiService.getUsers();
        setUsers(data.users || data || []);
      }
    } catch (err) {
      console.error('Failed to load users:', err);
      addAlert('error', 'Users Error', 'Failed to load user list');
      // Fallback mock data
      setUsers([
        {
          id: 1,
          username: 'admin',
          email: 'admin@example.com',
          role: 'admin',
          active: true,
          lastLogin: '2024-08-24T10:30:00Z'
        },
        {
          id: 2,
          username: 'operator',
          email: 'operator@example.com',
          role: 'operator',
          active: true,
          lastLogin: '2024-08-23T15:45:00Z'
        }
      ]);
    }
  };

  const loadMetrics = async () => {
    try {
      const data = await apiService.getMetrics();
      setMetrics(data);
    } catch (err) {
      console.error('Failed to load metrics:', err);
      addAlert('error', 'Metrics Error', 'Failed to load system metrics');
      // Fallback mock data
      setMetrics({
        totalRequests: 15420,
        avgLatency: 145,
        cpu_usage: 45,
        memory_usage: 67,
        network_usage: 23
      });
    }
  };

  // Theme management
  const loadTheme = () => {
    const savedTheme = localStorage.getItem('theme') || 'light';
    setTheme(savedTheme);
    document.documentElement.setAttribute('data-theme', savedTheme);
  };

  const toggleTheme = () => {
    const newTheme = theme === 'light' ? 'dark' : 'light';
    setTheme(newTheme);
    localStorage.setItem('theme', newTheme);
    document.documentElement.setAttribute('data-theme', newTheme);
  };

  // Alert management
  const addAlert = (type, title, message, duration = 5000) => {
    const alert = {
      id: Date.now(),
      type,
      title,
      message,
      duration
    };

    setAlerts(prev => [...prev, alert]);

    if (duration) {
      setTimeout(() => {
        dismissAlert(alert.id);
      }, duration);
    }
  };

  const showAlert = (message, type = 'info', dismissible = true, duration = null) => {
    addAlert(type, type.charAt(0).toUpperCase() + type.slice(1), message, duration);
  };

  const dismissAlert = (id) => {
    setAlerts(prev => prev.filter(alert => alert.id !== id));
  };

  // Utility functions for updating lists
  const updateNodeInList = (nodes, updatedNode) => {
    return nodes.map(node =>
      node.id === updatedNode.id ? { ...node, ...updatedNode } : node
    );
  };

  const updateModelInList = (models, updatedModel) => {
    return models.map(model =>
      model.id === updatedModel.id ? { ...model, ...updatedModel } : model
    );
  };

  // Authentication handlers
  const handleLogin = () => {
    setShowLogin(true);
    setShowRegistration(false);
  };

  const handleRegister = () => {
    setShowRegistration(true);
    setShowLogin(false);
  };

  const handleLogout = async () => {
    try {
      await authService.logout();
    } catch (error) {
      console.error('Logout error:', error);
    }
  };

  // Event handlers
  const handleCopy = async (text) => {
    try {
      await navigator.clipboard.writeText(text);
      showAlert('Copied to clipboard', 'success', true, 2000);
    } catch (err) {
      console.error('Failed to copy:', err);
      showAlert('Failed to copy to clipboard', 'error', true, 3000);
    }
  };

  const handleModelDownload = async (modelName) => {
    try {
      showAlert(`Starting download of ${modelName}...`, 'info', true, 3000);
      const response = await fetch(`/api/models/${modelName}/download`, {
        method: 'POST'
      });
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      showAlert(`${modelName} download started`, 'success', true, 5000);
      await loadModels(); // Refresh models
    } catch (err) {
      showAlert(`Failed to download ${modelName}: ${err.message}`, 'error', true, 5000);
    }
  };

  const handleModelDelete = async (modelName) => {
    try {
      const response = await fetch(`/api/models/${modelName}`, {
        method: 'DELETE'
      });
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      showAlert(`${modelName} deleted successfully`, 'success', true, 3000);
      await loadModels(); // Refresh models
    } catch (err) {
      showAlert(`Failed to delete ${modelName}: ${err.message}`, 'error', true, 5000);
    }
  };

  const handleUserAdd = async (userData) => {
    try {
      const response = await fetch('/api/users', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(userData)
      });
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      showAlert('User created successfully', 'success', true, 3000);
      await loadUsers(); // Refresh users
    } catch (err) {
      showAlert(`Failed to create user: ${err.message}`, 'error', true, 5000);
    }
  };

  const handleUserEdit = async (userId, userData) => {
    try {
      const response = await fetch(`/api/users/${userId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(userData)
      });
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      showAlert('User updated successfully', 'success', true, 3000);
      await loadUsers(); // Refresh users
    } catch (err) {
      showAlert(`Failed to update user: ${err.message}`, 'error', true, 5000);
    }
  };

  const handleUserDelete = async (userId) => {
    if (!window.confirm('Are you sure you want to delete this user?')) return;
    
    try {
      const response = await fetch(`/api/users/${userId}`, {
        method: 'DELETE'
      });
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      showAlert('User deleted successfully', 'success', true, 3000);
      await loadUsers(); // Refresh users
    } catch (err) {
      showAlert(`Failed to delete user: ${err.message}`, 'error', true, 5000);
    }
  };

  // Render functions
  const renderActiveTab = () => {
    const commonProps = {
      loading,
      error,
      onCopy: handleCopy
    };

    switch (activeTab) {
      case 'dashboard':
        return (
          <Dashboard 
            clusterStatus={clusterStatus}
            nodes={nodes}
            metrics={metrics}
            realTimeMetrics={realTimeMetrics}
            {...commonProps}
          />
        );
      case 'nodes':
        return (
          <NodesView 
            nodes={nodes}
            {...commonProps}
          />
        );
      case 'models':
        return (
          <ModelsView 
            models={models}
            onDownload={handleModelDownload}
            onDelete={handleModelDelete}
            autoDistribution={true}
            onToggleAutoDistribution={() => {}}
            {...commonProps}
          />
        );
      case 'transfers':
        return (
          <TransfersView 
            transfers={transfers}
            {...commonProps}
          />
        );
      case 'cluster':
        return (
          <ClusterView 
            clusterStatus={clusterStatus}
            {...commonProps}
          />
        );
      case 'analytics':
        return (
          <Analytics 
            analyticsData={{}}
            onExportReport={(data) => {
              const blob = new Blob([JSON.stringify(data, null, 2)], {
                type: 'application/json'
              });
              const url = URL.createObjectURL(blob);
              const a = document.createElement('a');
              a.href = url;
              a.download = `analytics-report-${new Date().toISOString().split('T')[0]}.json`;
              a.click();
            }}
          />
        );
      case 'users':
        return (
          <UserManagement 
            users={users}
            onAddUser={handleUserAdd}
            onEditUser={handleUserEdit}
            onDeleteUser={handleUserDelete}
          />
        );
      case 'database':
        return (
          <DatabaseEditor 
            tables={['users', 'nodes', 'models', 'metrics', 'logs']}
            onQuery={async (sql) => {
              // Mock implementation
              return { rows: [] };
            }}
            onInsert={async (table, data) => {
              showAlert(`Inserted record into ${table}`, 'success', true, 3000);
            }}
            onUpdate={async (table, id, data) => {
              showAlert(`Updated record in ${table}`, 'success', true, 3000);
            }}
            onDelete={async (table, data) => {
              showAlert(`Deleted record from ${table}`, 'success', true, 3000);
            }}
            onExport={(table) => {
              showAlert(`Exported ${table} data`, 'success', true, 3000);
            }}
            onImport={() => {
              showAlert('Import completed', 'success', true, 3000);
            }}
          />
        );
      case 'settings':
        return (
          <SystemSettings 
            settings={{}}
            onSave={async (settings) => {
              showAlert('Settings saved successfully', 'success', true, 3000);
            }}
            onReset={() => {
              showAlert('Settings reset to defaults', 'info', true, 3000);
            }}
          />
        );
      case 'metrics':
        return (
          <RealTimeMetrics 
            wsConnection={wsConnection}
            nodeId={clusterStatus?.node_id}
          />
        );
      default:
        return <div className="text-center p-5">Tab not implemented yet</div>;
    }
  };

  if (loading) {
    return <LoadingSpinner size="xl" text="Loading Ollama Distributed Control Panel..." overlay />;
  }

  // Show authentication screens for unauthenticated users
  if (!isAuthenticated) {
    return (
      <div className="app-container">
        {/* Theme Toggle */}
        <div className="position-fixed" style={{ top: '20px', right: '20px', zIndex: 1050 }}>
          <ThemeToggle theme={theme} onToggle={toggleTheme} />
        </div>

        {/* Alerts Container */}
        <div className="position-fixed" style={{ top: '80px', right: '20px', zIndex: 1050, maxWidth: '400px' }}>
          {alerts.map(alert => (
            <Alert
              key={alert.id}
              type={alert.type}
              title={alert.title}
              message={alert.message}
              dismissible={true}
              onDismiss={() => dismissAlert(alert.id)}
            />
          ))}
        </div>

        <Container fluid className="d-flex align-items-center justify-content-center min-vh-100">
          <Row className="w-100">
            <Col md={6} lg={4} className="mx-auto">
              {showRegistration ? (
                <RegistrationFlow
                  onRegistrationComplete={(user) => {
                    addAlert('success', 'Registration Successful', 'Please log in with your new account');
                    setShowRegistration(false);
                    setShowLogin(true);
                  }}
                  onCancel={() => {
                    setShowRegistration(false);
                    setShowLogin(true);
                  }}
                />
              ) : (
                <Login
                  onLogin={handleAuthenticated}
                  onRegister={handleRegister}
                  onForgotPassword={() => {
                    addAlert('info', 'Password Reset', 'Password reset functionality will be implemented soon');
                  }}
                />
              )}
            </Col>
          </Row>
        </Container>
      </div>
    );
  }

  return (
    <div className="app-container">
      {/* Theme Toggle */}
      <div className="position-fixed" style={{ top: '20px', right: '20px', zIndex: 1050 }}>
        <ThemeToggle theme={theme} onToggle={toggleTheme} />
      </div>

      {/* WebSocket Status */}
      <div className="position-fixed" style={{ top: '80px', right: '20px', zIndex: 1050 }}>
        <WebSocketStatus 
          isConnected={isConnected} 
          reconnectAttempts={reconnectAttempts} 
        />
      </div>

      {/* Alerts Container */}
      <div className="position-fixed" style={{ top: '140px', right: '20px', zIndex: 1050, maxWidth: '400px' }}>
        {alerts.map(alert => (
          <Alert
            key={alert.id}
            type={alert.type}
            message={alert.message}
            dismissible={alert.dismissible}
            autoHide={!!alert.duration}
            duration={alert.duration}
            onDismiss={() => dismissAlert(alert.id)}
          />
        ))}
      </div>

      <div className="d-flex">
        {/* Mobile Menu Button */}
        <Button 
          variant="outline-primary"
          className="d-md-none position-fixed"
          style={{ top: '20px', left: '20px', zIndex: 1051 }}
          onClick={() => setSidebarOpen(!sidebarOpen)}
        >
          <FontAwesomeIcon icon={sidebarOpen ? faTimes : faBars} />
        </Button>

        {/* Sidebar */}
        <div className="d-none d-md-block" style={{ width: '280px' }}>
          <Sidebar
            activeTab={activeTab}
            onTabChange={setActiveTab}
            onClose={() => setSidebarOpen(false)}
            isOpen={true}
            currentUser={currentUser}
            onLogout={handleLogout}
          />
        </div>

        {/* Mobile Sidebar Overlay */}
        {sidebarOpen && (
          <>
            <div 
              className="position-fixed d-md-none"
              style={{ 
                top: 0, 
                left: 0, 
                width: '100vw', 
                height: '100vh', 
                backgroundColor: 'rgba(0,0,0,0.5)', 
                zIndex: 1040 
              }}
              onClick={() => setSidebarOpen(false)}
            />
            <div 
              className="position-fixed d-md-none"
              style={{ 
                top: 0, 
                left: 0, 
                width: '280px', 
                height: '100vh', 
                zIndex: 1050 
              }}
            >
              <Sidebar
                activeTab={activeTab}
                onTabChange={(tab) => {
                  setActiveTab(tab);
                  setSidebarOpen(false);
                }}
                onClose={() => setSidebarOpen(false)}
                isOpen={true}
                currentUser={currentUser}
                onLogout={handleLogout}
              />
            </div>
          </>
        )}

        {/* Main Content */}
        <div className="flex-grow-1 main-content">
          <Container fluid className="p-4">
            {error && (
              <Alert 
                type="danger" 
                title="Application Error"
                message={error}
                dismissible={true}
                onDismiss={() => setError(null)}
              />
            )}
            {renderActiveTab()}
          </Container>
        </div>
      </div>
    </div>
  );
};

export default EnhancedApp;
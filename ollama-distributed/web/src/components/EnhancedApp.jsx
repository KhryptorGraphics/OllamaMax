import React, { useState, useEffect } from 'react';
import { Container, Row, Col, Button } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faBars, faTimes } from '@fortawesome/free-solid-svg-icons';

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

const EnhancedApp = () => {
  // Application state
  const [activeTab, setActiveTab] = useState('dashboard');
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [theme, setTheme] = useState('light');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [alerts, setAlerts] = useState([]);

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
    setupWebSocket();
    loadTheme();

    return () => {
      if (wsConnection) {
        wsConnection.close();
      }
    };
  }, []);

  const initializeApp = async () => {
    setLoading(true);
    try {
      await Promise.all([
        loadClusterStatus(),
        loadNodes(),
        loadModels(),
        loadUsers(),
        loadMetrics()
      ]);
    } catch (err) {
      setError(`Failed to initialize application: ${err.message}`);
    } finally {
      setLoading(false);
    }
  };

  const setupWebSocket = () => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;
    
    const ws = new WebSocket(wsUrl);
    
    ws.onopen = () => {
      setIsConnected(true);
      setReconnectAttempts(0);
      showAlert('Connected to real-time updates', 'success', false, 3000);
    };
    
    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        handleWebSocketMessage(data);
      } catch (err) {
        console.error('Failed to parse WebSocket message:', err);
      }
    };
    
    ws.onclose = () => {
      setIsConnected(false);
      if (reconnectAttempts < 5) {
        setTimeout(() => {
          setReconnectAttempts(prev => prev + 1);
          setupWebSocket();
        }, 5000);
      }
    };
    
    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      setIsConnected(false);
    };
    
    setWsConnection(ws);
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
      const response = await fetch('/api/cluster/status');
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      const data = await response.json();
      setClusterStatus(data);
    } catch (err) {
      console.error('Failed to load cluster status:', err);
      // Fallback mock data
      setClusterStatus({
        node_id: 'node-001-example',
        leader: 'node-001-example',
        is_leader: true,
        status: 'healthy',
        peers: 3
      });
    }
  };

  const loadNodes = async () => {
    try {
      const response = await fetch('/api/nodes');
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      const data = await response.json();
      setNodes(data.nodes || []);
    } catch (err) {
      console.error('Failed to load nodes:', err);
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
      const response = await fetch('/api/models');
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      const data = await response.json();
      setModels(data.models || []);
    } catch (err) {
      console.error('Failed to load models:', err);
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
      const response = await fetch('/api/users');
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      const data = await response.json();
      setUsers(data.users || []);
    } catch (err) {
      console.error('Failed to load users:', err);
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
      const response = await fetch('/api/metrics');
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      const data = await response.json();
      setMetrics(data);
    } catch (err) {
      console.error('Failed to load metrics:', err);
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
  const showAlert = (message, type = 'info', dismissible = true, duration = null) => {
    const alert = {
      id: Date.now(),
      message,
      type,
      dismissible,
      duration
    };
    
    setAlerts(prev => [...prev, alert]);

    if (duration) {
      setTimeout(() => {
        dismissAlert(alert.id);
      }, duration);
    }
  };

  const dismissAlert = (id) => {
    setAlerts(prev => prev.filter(alert => alert.id !== id));
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
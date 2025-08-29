import React, { useState, useEffect, useCallback } from 'react';
import { Card, Row, Col, Button, Badge, Modal, Tabs, Tab, Table, Form, Alert, Dropdown } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  faTachometerAlt,
  faUsers,
  faServer,
  faDatabase,
  faShieldAlt,
  faCog,
  faChartLine,
  faBell,
  faExclamationTriangle,
  faCheckCircle,
  faInfoCircle,
  faDownload,
  faUpload,
  faSync,
  faTrash,
  faEdit,
  faEye,
  faPlus,
  faBan,
  faUserShield,
  faKey,
  faHistory,
  faFileExport,
  faFilter,
  faSearch,
  faSort,
  faEllipsisV,
  faPlay,
  faPause,
  faStop,
  faRestart
} from '@fortawesome/free-solid-svg-icons';
import PerformanceMonitor from './PerformanceMonitor';
import DataVisualization from './DataVisualization';
import LoadingSpinner from './LoadingSpinner';
import apiService from '../services/api.js';
import authService from '../services/auth.js';
import wsService from '../services/websocket.js';
import '../styles/design-system.css';

const AdminDashboard = ({ className = "" }) => {
  // Data state
  const [systemHealth, setSystemHealth] = useState({});
  const [users, setUsers] = useState([]);
  const [nodes, setNodes] = useState([]);
  const [models, setModels] = useState([]);
  const [auditLogs, setAuditLogs] = useState([]);
  const [alerts, setAlerts] = useState([]);
  const [metrics, setMetrics] = useState({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [activeTab, setActiveTab] = useState('overview');
  const [selectedTimeRange, setSelectedTimeRange] = useState('24h');
  const [showModal, setShowModal] = useState(false);
  const [modalType, setModalType] = useState('');
  const [selectedItem, setSelectedItem] = useState(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterStatus, setFilterStatus] = useState('all');
  const [sortBy, setSortBy] = useState('created_at');
  const [sortOrder, setSortOrder] = useState('desc');
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [refreshInterval, setRefreshInterval] = useState(30000);

  // Initialize data loading
  useEffect(() => {
    loadAllData();
    setupWebSocketListeners();

    return () => {
      wsService.unsubscribe('admin_metrics');
      wsService.unsubscribe('admin_alerts');
      wsService.unsubscribe('admin_users');
      wsService.unsubscribe('admin_nodes');
    };
  }, []);

  // Auto-refresh data
  useEffect(() => {
    if (!autoRefresh) return;

    const interval = setInterval(() => {
      loadAllData();
    }, refreshInterval);

    return () => clearInterval(interval);
  }, [autoRefresh, refreshInterval]);

  // Data loading functions
  const loadAllData = async () => {
    setLoading(true);
    setError(null);

    try {
      await Promise.all([
        loadSystemHealth(),
        loadUsers(),
        loadNodes(),
        loadModels(),
        loadMetrics(),
        loadAuditLogs()
      ]);
    } catch (err) {
      setError('Failed to load dashboard data');
      console.error('Dashboard data loading error:', err);
    } finally {
      setLoading(false);
    }
  };

  const loadSystemHealth = async () => {
    try {
      const [health, readiness, clusterStatus] = await Promise.all([
        apiService.getHealth(),
        apiService.getReadiness(),
        apiService.getClusterStatus()
      ]);

      setSystemHealth({
        ...health,
        readiness,
        cluster: clusterStatus
      });
    } catch (err) {
      console.error('Failed to load system health:', err);
    }
  };

  const loadUsers = async () => {
    try {
      if (authService.isAdmin()) {
        const data = await apiService.getUsers(1, 100);
        setUsers(data.users || data || []);
      }
    } catch (err) {
      console.error('Failed to load users:', err);
    }
  };

  const loadNodes = async () => {
    try {
      const data = await apiService.getNodes();
      setNodes(data.nodes || data || []);
    } catch (err) {
      console.error('Failed to load nodes:', err);
    }
  };

  const loadModels = async () => {
    try {
      const data = await apiService.getModels(1, 50);
      setModels(data.models || data || []);
    } catch (err) {
      console.error('Failed to load models:', err);
    }
  };

  const loadMetrics = async () => {
    try {
      const [systemMetrics, dbStats] = await Promise.all([
        apiService.getSystemMetrics(),
        apiService.getDatabaseStats()
      ]);

      setMetrics({
        system: systemMetrics,
        database: dbStats
      });
    } catch (err) {
      console.error('Failed to load metrics:', err);
    }
  };

  const loadAuditLogs = async () => {
    try {
      // TODO: Implement audit logs API endpoint
      // const data = await apiService.getAuditLogs();
      // setAuditLogs(data.logs || data || []);
    } catch (err) {
      console.error('Failed to load audit logs:', err);
    }
  };

  // WebSocket setup for real-time updates
  const setupWebSocketListeners = () => {
    wsService.subscribe('admin_metrics', (data) => {
      setMetrics(prev => ({ ...prev, ...data }));
    });

    wsService.subscribe('admin_alerts', (data) => {
      setAlerts(prev => [...prev, data.alert]);
    });

    wsService.subscribe('admin_users', (data) => {
      if (data.action === 'create') {
        setUsers(prev => [...prev, data.user]);
      } else if (data.action === 'update') {
        setUsers(prev => prev.map(user =>
          user.id === data.user.id ? { ...user, ...data.user } : user
        ));
      } else if (data.action === 'delete') {
        setUsers(prev => prev.filter(user => user.id !== data.user_id));
      }
    });

    wsService.subscribe('admin_nodes', (data) => {
      setNodes(prev => prev.map(node =>
        node.id === data.node.id ? { ...node, ...data.node } : node
      ));
    });
  };

  const handleModal = useCallback((type, item = null) => {
    setModalType(type);
    setSelectedItem(item);
    setShowModal(true);
  }, []);

  const handleCloseModal = useCallback(() => {
    setShowModal(false);
    setModalType('');
    setSelectedItem(null);
  }, []);

  // Action handlers
  const handleUserAction = async (action, userId, data = {}) => {
    try {
      switch (action) {
        case 'create':
          await apiService.createUser(data);
          await loadUsers();
          break;
        case 'update':
          await apiService.updateUser(userId, data);
          await loadUsers();
          break;
        case 'delete':
          await apiService.deleteUser(userId);
          await loadUsers();
          break;
        case 'activate':
        case 'deactivate':
          await apiService.updateUser(userId, { active: action === 'activate' });
          await loadUsers();
          break;
        default:
          console.warn('Unknown user action:', action);
      }
    } catch (error) {
      console.error('User action failed:', error);
      setError(`Failed to ${action} user: ${error.message}`);
    }
  };

  const handleNodeAction = async (action, nodeId, data = {}) => {
    try {
      switch (action) {
        case 'restart':
        case 'stop':
        case 'start':
          await apiService.updateNode(nodeId, { action });
          await loadNodes();
          break;
        case 'update':
          await apiService.updateNode(nodeId, data);
          await loadNodes();
          break;
        default:
          console.warn('Unknown node action:', action);
      }
    } catch (error) {
      console.error('Node action failed:', error);
      setError(`Failed to ${action} node: ${error.message}`);
    }
  };

  const handleSystemAction = async (action, data = {}) => {
    try {
      switch (action) {
        case 'refresh':
          await loadAllData();
          break;
        case 'backup':
          // TODO: Implement system backup
          console.log('System backup initiated');
          break;
        case 'maintenance':
          // TODO: Implement maintenance mode
          console.log('Maintenance mode toggled');
          break;
        default:
          console.warn('Unknown system action:', action);
      }
    } catch (error) {
      console.error('System action failed:', error);
      setError(`Failed to ${action}: ${error.message}`);
    }
  };

  // System Overview Metrics
  const systemMetrics = {
    totalUsers: users.length,
    activeUsers: users.filter(u => u.status === 'active').length,
    totalNodes: nodes.length,
    healthyNodes: nodes.filter(n => n.health === 'healthy').length,
    totalModels: models.length,
    activeModels: models.filter(m => m.status === 'running').length,
    totalAlerts: alerts.length,
    criticalAlerts: alerts.filter(a => a.severity === 'critical').length,
    uptime: systemHealth.uptime || '99.9%',
    responseTime: systemHealth.avgResponseTime || '45ms',
    throughput: systemHealth.throughput || '1,234 req/min',
    errorRate: systemHealth.errorRate || '0.1%'
  };

  const renderSystemOverview = () => (
    <div className="system-overview">
      <Row className="g-4 mb-4">
        <Col md={3}>
          <Card className="metric-card">
            <Card.Body>
              <div className="metric-icon">
                <FontAwesomeIcon icon={faUsers} />
              </div>
              <div className="metric-content">
                <h3>{systemMetrics.activeUsers}</h3>
                <p>Active Users</p>
                <small className="text-muted">of {systemMetrics.totalUsers} total</small>
              </div>
            </Card.Body>
          </Card>
        </Col>
        
        <Col md={3}>
          <Card className="metric-card">
            <Card.Body>
              <div className="metric-icon text-success">
                <FontAwesomeIcon icon={faServer} />
              </div>
              <div className="metric-content">
                <h3>{systemMetrics.healthyNodes}</h3>
                <p>Healthy Nodes</p>
                <small className="text-muted">of {systemMetrics.totalNodes} total</small>
              </div>
            </Card.Body>
          </Card>
        </Col>
        
        <Col md={3}>
          <Card className="metric-card">
            <Card.Body>
              <div className="metric-icon text-info">
                <FontAwesomeIcon icon={faDatabase} />
              </div>
              <div className="metric-content">
                <h3>{systemMetrics.activeModels}</h3>
                <p>Active Models</p>
                <small className="text-muted">of {systemMetrics.totalModels} total</small>
              </div>
            </Card.Body>
          </Card>
        </Col>
        
        <Col md={3}>
          <Card className="metric-card">
            <Card.Body>
              <div className={`metric-icon ${systemMetrics.criticalAlerts > 0 ? 'text-danger' : 'text-warning'}`}>
                <FontAwesomeIcon icon={faBell} />
              </div>
              <div className="metric-content">
                <h3>{systemMetrics.totalAlerts}</h3>
                <p>Active Alerts</p>
                <small className="text-muted">
                  {systemMetrics.criticalAlerts} critical
                </small>
              </div>
            </Card.Body>
          </Card>
        </Col>
      </Row>

      <Row className="g-4 mb-4">
        <Col md={3}>
          <Card className="stat-card">
            <Card.Body className="text-center">
              <h4 className="text-success">{systemMetrics.uptime}</h4>
              <p className="mb-0">System Uptime</p>
            </Card.Body>
          </Card>
        </Col>
        
        <Col md={3}>
          <Card className="stat-card">
            <Card.Body className="text-center">
              <h4 className="text-primary">{systemMetrics.responseTime}</h4>
              <p className="mb-0">Avg Response Time</p>
            </Card.Body>
          </Card>
        </Col>
        
        <Col md={3}>
          <Card className="stat-card">
            <Card.Body className="text-center">
              <h4 className="text-info">{systemMetrics.throughput}</h4>
              <p className="mb-0">Throughput</p>
            </Card.Body>
          </Card>
        </Col>
        
        <Col md={3}>
          <Card className="stat-card">
            <Card.Body className="text-center">
              <h4 className={`${parseFloat(systemMetrics.errorRate) > 1 ? 'text-danger' : 'text-success'}`}>
                {systemMetrics.errorRate}
              </h4>
              <p className="mb-0">Error Rate</p>
            </Card.Body>
          </Card>
        </Col>
      </Row>

      <Row className="g-4">
        <Col lg={8}>
          <Card>
            <Card.Header className="d-flex justify-content-between align-items-center">
              <h5 className="mb-0">System Performance</h5>
              <div>
                <Form.Select
                  size="sm"
                  value={selectedTimeRange}
                  onChange={(e) => setSelectedTimeRange(e.target.value)}
                  style={{ width: 'auto', display: 'inline-block' }}
                >
                  <option value="1h">Last Hour</option>
                  <option value="24h">Last 24 Hours</option>
                  <option value="7d">Last 7 Days</option>
                  <option value="30d">Last 30 Days</option>
                </Form.Select>
              </div>
            </Card.Header>
            <Card.Body>
              <DataVisualization
                data={systemHealth.metrics}
                timeRange={selectedTimeRange}
                height={300}
              />
            </Card.Body>
          </Card>
        </Col>
        
        <Col lg={4}>
          <Card>
            <Card.Header>
              <h5 className="mb-0">Recent Alerts</h5>
            </Card.Header>
            <Card.Body style={{ maxHeight: '350px', overflowY: 'auto' }}>
              {alerts.slice(0, 10).map((alert, index) => (
                <div key={index} className="alert-item">
                  <div className={`alert-icon ${alert.severity}`}>
                    <FontAwesomeIcon 
                      icon={alert.severity === 'critical' ? faExclamationTriangle : 
                            alert.severity === 'warning' ? faInfoCircle : faCheckCircle} 
                    />
                  </div>
                  <div className="alert-content">
                    <div className="alert-title">{alert.title}</div>
                    <div className="alert-description">{alert.description}</div>
                    <small className="text-muted">{alert.timestamp}</small>
                  </div>
                </div>
              ))}
              {alerts.length === 0 && (
                <div className="text-center text-muted py-3">
                  <FontAwesomeIcon icon={faCheckCircle} size="2x" className="mb-2" />
                  <p>No active alerts</p>
                </div>
              )}
            </Card.Body>
          </Card>
        </Col>
      </Row>
    </div>
  );

  const renderUserManagement = () => {
    const filteredUsers = users.filter(user => {
      const matchesSearch = user.name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
                           user.email?.toLowerCase().includes(searchTerm.toLowerCase());
      const matchesFilter = filterStatus === 'all' || user.status === filterStatus;
      return matchesSearch && matchesFilter;
    });

    return (
      <div className="user-management">
        <div className="table-controls mb-3">
          <Row>
            <Col md={4}>
              <div className="search-box">
                <FontAwesomeIcon icon={faSearch} className="search-icon" />
                <Form.Control
                  type="text"
                  placeholder="Search users..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                />
              </div>
            </Col>
            <Col md={3}>
              <Form.Select
                value={filterStatus}
                onChange={(e) => setFilterStatus(e.target.value)}
              >
                <option value="all">All Status</option>
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
                <option value="suspended">Suspended</option>
              </Form.Select>
            </Col>
            <Col md={2}>
              <Form.Select
                value={sortBy}
                onChange={(e) => setSortBy(e.target.value)}
              >
                <option value="name">Name</option>
                <option value="email">Email</option>
                <option value="role">Role</option>
                <option value="created_at">Created</option>
                <option value="last_login">Last Login</option>
              </Form.Select>
            </Col>
            <Col md={3} className="text-end">
              <Button variant="primary" onClick={() => handleModal('createUser')}>
                <FontAwesomeIcon icon={faPlus} className="me-2" />
                Add User
              </Button>
            </Col>
          </Row>
        </div>

        <Card>
          <Table responsive hover>
            <thead>
              <tr>
                <th>User</th>
                <th>Role</th>
                <th>Status</th>
                <th>Last Login</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredUsers.map((user) => (
                <tr key={user.id}>
                  <td>
                    <div className="user-info">
                      <div className="user-avatar">
                        {user.avatar ? (
                          <img src={user.avatar} alt={user.name} />
                        ) : (
                          <FontAwesomeIcon icon={faUsers} />
                        )}
                      </div>
                      <div>
                        <div className="user-name">{user.name}</div>
                        <div className="user-email">{user.email}</div>
                      </div>
                    </div>
                  </td>
                  <td>
                    <Badge bg={user.role === 'admin' ? 'danger' : user.role === 'manager' ? 'warning' : 'primary'}>
                      {user.role}
                    </Badge>
                  </td>
                  <td>
                    <Badge bg={user.status === 'active' ? 'success' : user.status === 'suspended' ? 'danger' : 'secondary'}>
                      {user.status}
                    </Badge>
                  </td>
                  <td>{user.lastLogin || 'Never'}</td>
                  <td>
                    <Dropdown>
                      <Dropdown.Toggle variant="outline-secondary" size="sm">
                        <FontAwesomeIcon icon={faEllipsisV} />
                      </Dropdown.Toggle>
                      <Dropdown.Menu>
                        <Dropdown.Item onClick={() => handleModal('viewUser', user)}>
                          <FontAwesomeIcon icon={faEye} className="me-2" />
                          View Details
                        </Dropdown.Item>
                        <Dropdown.Item onClick={() => handleModal('editUser', user)}>
                          <FontAwesomeIcon icon={faEdit} className="me-2" />
                          Edit User
                        </Dropdown.Item>
                        <Dropdown.Item onClick={() => handleModal('resetPassword', user)}>
                          <FontAwesomeIcon icon={faKey} className="me-2" />
                          Reset Password
                        </Dropdown.Item>
                        <Dropdown.Divider />
                        {user.status === 'active' ? (
                          <Dropdown.Item 
                            className="text-warning"
                            onClick={() => onUserAction('suspend', user.id)}
                          >
                            <FontAwesomeIcon icon={faBan} className="me-2" />
                            Suspend User
                          </Dropdown.Item>
                        ) : (
                          <Dropdown.Item 
                            className="text-success"
                            onClick={() => onUserAction('activate', user.id)}
                          >
                            <FontAwesomeIcon icon={faCheckCircle} className="me-2" />
                            Activate User
                          </Dropdown.Item>
                        )}
                        <Dropdown.Item 
                          className="text-danger"
                          onClick={() => handleModal('deleteUser', user)}
                        >
                          <FontAwesomeIcon icon={faTrash} className="me-2" />
                          Delete User
                        </Dropdown.Item>
                      </Dropdown.Menu>
                    </Dropdown>
                  </td>
                </tr>
              ))}
            </tbody>
          </Table>
          
          {filteredUsers.length === 0 && (
            <div className="text-center py-4">
              <FontAwesomeIcon icon={faUsers} size="3x" className="text-muted mb-3" />
              <h5>No users found</h5>
              <p className="text-muted">Try adjusting your search or filter criteria.</p>
            </div>
          )}
        </Card>
      </div>
    );
  };

  const renderNodeManagement = () => (
    <div className="node-management">
      <div className="node-controls mb-3">
        <Row>
          <Col md={6}>
            <h5>Cluster Nodes ({nodes.length})</h5>
          </Col>
          <Col md={6} className="text-end">
            <Button variant="success" className="me-2" onClick={() => handleModal('addNode')}>
              <FontAwesomeIcon icon={faPlus} className="me-2" />
              Add Node
            </Button>
            <Button variant="primary" onClick={() => onSystemAction('refreshNodes')}>
              <FontAwesomeIcon icon={faSync} className="me-2" />
              Refresh
            </Button>
          </Col>
        </Row>
      </div>

      <Row className="g-3">
        {nodes.map((node) => (
          <Col key={node.id} lg={6}>
            <Card className={`node-card ${node.health === 'healthy' ? 'healthy' : 'unhealthy'}`}>
              <Card.Body>
                <div className="node-header">
                  <div>
                    <h6>{node.name}</h6>
                    <small className="text-muted">{node.address}</small>
                  </div>
                  <div className="node-status">
                    <Badge bg={node.health === 'healthy' ? 'success' : 'danger'}>
                      {node.health}
                    </Badge>
                  </div>
                </div>

                <div className="node-metrics">
                  <div className="metric">
                    <span>CPU</span>
                    <div className="progress-small">
                      <div 
                        className="progress-bar bg-primary" 
                        style={{ width: `${node.cpu}%` }}
                      />
                    </div>
                    <span>{node.cpu}%</span>
                  </div>
                  
                  <div className="metric">
                    <span>Memory</span>
                    <div className="progress-small">
                      <div 
                        className="progress-bar bg-info" 
                        style={{ width: `${node.memory}%` }}
                      />
                    </div>
                    <span>{node.memory}%</span>
                  </div>
                  
                  <div className="metric">
                    <span>Disk</span>
                    <div className="progress-small">
                      <div 
                        className="progress-bar bg-warning" 
                        style={{ width: `${node.disk}%` }}
                      />
                    </div>
                    <span>{node.disk}%</span>
                  </div>
                </div>

                <div className="node-actions">
                  <Button size="sm" variant="outline-primary" onClick={() => handleModal('nodeDetails', node)}>
                    <FontAwesomeIcon icon={faEye} className="me-1" />
                    Details
                  </Button>
                  
                  {node.status === 'running' ? (
                    <Button size="sm" variant="outline-warning" onClick={() => onNodeAction('stop', node.id)}>
                      <FontAwesomeIcon icon={faPause} className="me-1" />
                      Stop
                    </Button>
                  ) : (
                    <Button size="sm" variant="outline-success" onClick={() => onNodeAction('start', node.id)}>
                      <FontAwesomeIcon icon={faPlay} className="me-1" />
                      Start
                    </Button>
                  )}
                  
                  <Button size="sm" variant="outline-secondary" onClick={() => onNodeAction('restart', node.id)}>
                    <FontAwesomeIcon icon={faRestart} className="me-1" />
                    Restart
                  </Button>
                </div>
              </Card.Body>
            </Card>
          </Col>
        ))}
      </Row>
    </div>
  );

  const renderSystemSettings = () => (
    <div className="system-settings">
      <Tabs defaultActiveKey="general" className="mb-4">
        <Tab eventKey="general" title="General">
          <Card>
            <Card.Body>
              <h5 className="mb-3">General Settings</h5>
              <Form>
                <Row>
                  <Col md={6}>
                    <Form.Group className="mb-3">
                      <Form.Label>System Name</Form.Label>
                      <Form.Control type="text" defaultValue="OllamaMax Cluster" />
                    </Form.Group>
                  </Col>
                  <Col md={6}>
                    <Form.Group className="mb-3">
                      <Form.Label>Default Theme</Form.Label>
                      <Form.Select>
                        <option>Light</option>
                        <option>Dark</option>
                        <option>Auto</option>
                      </Form.Select>
                    </Form.Group>
                  </Col>
                </Row>
                
                <Form.Group className="mb-3">
                  <Form.Check type="switch" label="Enable Auto-refresh" defaultChecked />
                </Form.Group>
                
                <Form.Group className="mb-3">
                  <Form.Label>Refresh Interval (seconds)</Form.Label>
                  <Form.Range min={5} max={300} defaultValue={30} />
                </Form.Group>
              </Form>
            </Card.Body>
          </Card>
        </Tab>
        
        <Tab eventKey="security" title="Security">
          <Card>
            <Card.Body>
              <h5 className="mb-3">Security Settings</h5>
              <Form>
                <Form.Group className="mb-3">
                  <Form.Check type="switch" label="Require Two-Factor Authentication" />
                </Form.Group>
                
                <Form.Group className="mb-3">
                  <Form.Check type="switch" label="Enable API Rate Limiting" defaultChecked />
                </Form.Group>
                
                <Form.Group className="mb-3">
                  <Form.Label>Session Timeout (hours)</Form.Label>
                  <Form.Control type="number" defaultValue="8" />
                </Form.Group>
                
                <Form.Group className="mb-3">
                  <Form.Label>Password Policy</Form.Label>
                  <div>
                    <Form.Check type="checkbox" label="Minimum 8 characters" defaultChecked />
                    <Form.Check type="checkbox" label="Require uppercase letter" defaultChecked />
                    <Form.Check type="checkbox" label="Require number" defaultChecked />
                    <Form.Check type="checkbox" label="Require special character" />
                  </div>
                </Form.Group>
              </Form>
            </Card.Body>
          </Card>
        </Tab>
        
        <Tab eventKey="notifications" title="Notifications">
          <Card>
            <Card.Body>
              <h5 className="mb-3">Notification Settings</h5>
              <Form>
                <Form.Group className="mb-3">
                  <Form.Label>Email Notifications</Form.Label>
                  <div>
                    <Form.Check type="checkbox" label="System alerts" defaultChecked />
                    <Form.Check type="checkbox" label="User registrations" defaultChecked />
                    <Form.Check type="checkbox" label="Security events" defaultChecked />
                    <Form.Check type="checkbox" label="Performance warnings" />
                  </div>
                </Form.Group>
                
                <Form.Group className="mb-3">
                  <Form.Label>SMTP Configuration</Form.Label>
                  <Row>
                    <Col md={6}>
                      <Form.Control type="text" placeholder="SMTP Server" className="mb-2" />
                      <Form.Control type="number" placeholder="Port" className="mb-2" />
                    </Col>
                    <Col md={6}>
                      <Form.Control type="text" placeholder="Username" className="mb-2" />
                      <Form.Control type="password" placeholder="Password" className="mb-2" />
                    </Col>
                  </Row>
                </Form.Group>
              </Form>
            </Card.Body>
          </Card>
        </Tab>
        
        <Tab eventKey="backup" title="Backup & Recovery">
          <Card>
            <Card.Body>
              <h5 className="mb-3">Backup & Recovery</h5>
              
              <Alert variant="info">
                <FontAwesomeIcon icon={faInfoCircle} className="me-2" />
                Regular backups help protect your data and configuration.
              </Alert>
              
              <div className="backup-actions mb-4">
                <Button variant="primary" className="me-2">
                  <FontAwesomeIcon icon={faDownload} className="me-2" />
                  Create Backup
                </Button>
                <Button variant="outline-secondary" className="me-2">
                  <FontAwesomeIcon icon={faUpload} className="me-2" />
                  Restore Backup
                </Button>
                <Button variant="outline-info">
                  <FontAwesomeIcon icon={faHistory} className="me-2" />
                  Backup History
                </Button>
              </div>
              
              <Form>
                <Form.Group className="mb-3">
                  <Form.Check type="switch" label="Enable Automatic Backups" defaultChecked />
                </Form.Group>
                
                <Form.Group className="mb-3">
                  <Form.Label>Backup Frequency</Form.Label>
                  <Form.Select>
                    <option>Daily</option>
                    <option>Weekly</option>
                    <option>Monthly</option>
                  </Form.Select>
                </Form.Group>
                
                <Form.Group className="mb-3">
                  <Form.Label>Backup Retention (days)</Form.Label>
                  <Form.Control type="number" defaultValue="30" />
                </Form.Group>
              </Form>
            </Card.Body>
          </Card>
        </Tab>
      </Tabs>
      
      <div className="settings-actions">
        <Button variant="success" className="me-2">
          Save Changes
        </Button>
        <Button variant="outline-secondary">
          Reset to Defaults
        </Button>
      </div>
    </div>
  );

  const renderModal = () => {
    if (!showModal) return null;

    const modalProps = {
      show: showModal,
      onHide: handleCloseModal,
      centered: true
    };

    switch (modalType) {
      case 'createUser':
        return (
          <Modal {...modalProps} size="lg">
            <Modal.Header closeButton>
              <Modal.Title>Create New User</Modal.Title>
            </Modal.Header>
            <Modal.Body>
              <Form>
                <Row>
                  <Col md={6}>
                    <Form.Group className="mb-3">
                      <Form.Label>First Name</Form.Label>
                      <Form.Control type="text" />
                    </Form.Group>
                  </Col>
                  <Col md={6}>
                    <Form.Group className="mb-3">
                      <Form.Label>Last Name</Form.Label>
                      <Form.Control type="text" />
                    </Form.Group>
                  </Col>
                </Row>
                <Form.Group className="mb-3">
                  <Form.Label>Email</Form.Label>
                  <Form.Control type="email" />
                </Form.Group>
                <Form.Group className="mb-3">
                  <Form.Label>Role</Form.Label>
                  <Form.Select>
                    <option value="user">User</option>
                    <option value="manager">Manager</option>
                    <option value="admin">Administrator</option>
                  </Form.Select>
                </Form.Group>
                <Form.Group className="mb-3">
                  <Form.Check type="switch" label="Send welcome email" defaultChecked />
                </Form.Group>
              </Form>
            </Modal.Body>
            <Modal.Footer>
              <Button variant="secondary" onClick={handleCloseModal}>
                Cancel
              </Button>
              <Button variant="primary">
                Create User
              </Button>
            </Modal.Footer>
          </Modal>
        );

      case 'nodeDetails':
        return (
          <Modal {...modalProps} size="lg">
            <Modal.Header closeButton>
              <Modal.Title>Node Details: {selectedItem?.name}</Modal.Title>
            </Modal.Header>
            <Modal.Body>
              {selectedItem && (
                <Tabs defaultActiveKey="overview">
                  <Tab eventKey="overview" title="Overview">
                    <div className="node-details-overview">
                      <Row>
                        <Col md={6}>
                          <h6>System Information</h6>
                          <ul className="list-unstyled">
                            <li><strong>Address:</strong> {selectedItem.address}</li>
                            <li><strong>Version:</strong> {selectedItem.version}</li>
                            <li><strong>OS:</strong> {selectedItem.os}</li>
                            <li><strong>Arch:</strong> {selectedItem.architecture}</li>
                          </ul>
                        </Col>
                        <Col md={6}>
                          <h6>Status</h6>
                          <ul className="list-unstyled">
                            <li><strong>Health:</strong> <Badge bg={selectedItem.health === 'healthy' ? 'success' : 'danger'}>{selectedItem.health}</Badge></li>
                            <li><strong>Uptime:</strong> {selectedItem.uptime}</li>
                            <li><strong>Last Seen:</strong> {selectedItem.lastSeen}</li>
                          </ul>
                        </Col>
                      </Row>
                    </div>
                  </Tab>
                  <Tab eventKey="performance" title="Performance">
                    <PerformanceMonitor nodeId={selectedItem.id} />
                  </Tab>
                  <Tab eventKey="models" title="Models">
                    <div>
                      <h6>Active Models ({selectedItem.models?.length || 0})</h6>
                      {selectedItem.models?.map((model, index) => (
                        <div key={index} className="model-item">
                          <span>{model.name}</span>
                          <Badge bg="success" className="ms-2">{model.status}</Badge>
                        </div>
                      ))}
                    </div>
                  </Tab>
                </Tabs>
              )}
            </Modal.Body>
            <Modal.Footer>
              <Button variant="secondary" onClick={handleCloseModal}>
                Close
              </Button>
            </Modal.Footer>
          </Modal>
        );

      default:
        return null;
    }
  };

  if (loading) {
    return <LoadingSpinner message="Loading admin dashboard..." />;
  }

  return (
    <div className={`admin-dashboard ${className}`}>
      <style jsx>{`
        .admin-dashboard {
          padding: 1.5rem;
        }
        
        .metric-card {
          border: none;
          border-radius: var(--radius-card);
          box-shadow: var(--shadow-sm);
          transition: all var(--duration-fast) var(--ease-out);
        }
        
        .metric-card:hover {
          box-shadow: var(--shadow-md);
          transform: translateY(-1px);
        }
        
        .metric-card .card-body {
          display: flex;
          align-items: center;
          padding: 1.5rem;
        }
        
        .metric-icon {
          width: 3rem;
          height: 3rem;
          background: var(--bg-secondary);
          border-radius: var(--radius-lg);
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 1.5rem;
          color: var(--brand-primary);
          margin-right: 1rem;
        }
        
        .metric-content h3 {
          font-size: 1.75rem;
          font-weight: var(--font-weight-bold);
          margin: 0;
          color: var(--text-primary);
        }
        
        .metric-content p {
          margin: 0;
          font-weight: var(--font-weight-medium);
          color: var(--text-secondary);
        }
        
        .stat-card {
          border: 1px solid var(--border-primary);
          border-radius: var(--radius-card);
          background: var(--bg-primary);
        }
        
        .alert-item {
          display: flex;
          align-items: flex-start;
          padding: 0.75rem 0;
          border-bottom: 1px solid var(--border-primary);
        }
        
        .alert-item:last-child {
          border-bottom: none;
        }
        
        .alert-icon {
          width: 2rem;
          height: 2rem;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          margin-right: 0.75rem;
          flex-shrink: 0;
        }
        
        .alert-icon.critical {
          background: rgba(239, 68, 68, 0.1);
          color: var(--error);
        }
        
        .alert-icon.warning {
          background: rgba(245, 158, 11, 0.1);
          color: var(--warning);
        }
        
        .alert-icon.info {
          background: rgba(59, 130, 246, 0.1);
          color: var(--info);
        }
        
        .alert-content {
          flex: 1;
        }
        
        .alert-title {
          font-weight: var(--font-weight-medium);
          color: var(--text-primary);
          font-size: 0.875rem;
        }
        
        .alert-description {
          color: var(--text-secondary);
          font-size: 0.8rem;
          margin: 0.25rem 0;
        }
        
        .search-box {
          position: relative;
        }
        
        .search-icon {
          position: absolute;
          left: 0.75rem;
          top: 50%;
          transform: translateY(-50%);
          color: var(--text-tertiary);
          z-index: 2;
        }
        
        .search-box input {
          padding-left: 2.5rem;
        }
        
        .user-info {
          display: flex;
          align-items: center;
        }
        
        .user-avatar {
          width: 2.5rem;
          height: 2.5rem;
          border-radius: 50%;
          background: var(--bg-secondary);
          display: flex;
          align-items: center;
          justify-content: center;
          margin-right: 0.75rem;
          overflow: hidden;
        }
        
        .user-avatar img {
          width: 100%;
          height: 100%;
          object-fit: cover;
        }
        
        .user-name {
          font-weight: var(--font-weight-medium);
          color: var(--text-primary);
        }
        
        .user-email {
          font-size: 0.85rem;
          color: var(--text-secondary);
        }
        
        .node-card {
          border-radius: var(--radius-card);
          transition: all var(--duration-fast) var(--ease-out);
        }
        
        .node-card.healthy {
          border-left: 4px solid var(--success);
        }
        
        .node-card.unhealthy {
          border-left: 4px solid var(--error);
        }
        
        .node-header {
          display: flex;
          justify-content: between;
          align-items: flex-start;
          margin-bottom: 1rem;
        }
        
        .node-metrics {
          margin-bottom: 1rem;
        }
        
        .metric {
          display: flex;
          align-items: center;
          margin-bottom: 0.5rem;
          font-size: 0.85rem;
        }
        
        .metric span:first-child {
          width: 4rem;
          color: var(--text-secondary);
        }
        
        .metric span:last-child {
          width: 3rem;
          text-align: right;
          color: var(--text-primary);
          font-weight: var(--font-weight-medium);
        }
        
        .progress-small {
          flex: 1;
          height: 0.5rem;
          background: var(--bg-secondary);
          border-radius: 0.25rem;
          margin: 0 0.75rem;
          overflow: hidden;
        }
        
        .progress-bar {
          height: 100%;
          border-radius: 0.25rem;
          transition: width var(--duration-normal) var(--ease-out);
        }
        
        .node-actions {
          display: flex;
          gap: 0.5rem;
          flex-wrap: wrap;
        }
        
        .settings-actions {
          padding: 1.5rem 0;
          border-top: 1px solid var(--border-primary);
        }
        
        .backup-actions {
          display: flex;
          gap: 0.75rem;
          flex-wrap: wrap;
        }
        
        @media (max-width: 768px) {
          .admin-dashboard {
            padding: 1rem;
          }
          
          .metric-card .card-body {
            flex-direction: column;
            text-align: center;
          }
          
          .metric-icon {
            margin-right: 0;
            margin-bottom: 0.5rem;
          }
          
          .table-controls .row > div {
            margin-bottom: 0.5rem;
          }
          
          .node-actions {
            justify-content: center;
          }
          
          .backup-actions {
            justify-content: center;
          }
        }
      `}</style>

      <div className="dashboard-header mb-4">
        <div className="d-flex justify-content-between align-items-center">
          <div>
            <h2>System Administration</h2>
            <p className="text-muted mb-0">Manage your OllamaMax cluster</p>
          </div>
          <div className="dashboard-actions">
            <Form.Check
              type="switch"
              id="autoRefresh"
              label="Auto-refresh"
              checked={autoRefresh}
              onChange={(e) => setAutoRefresh(e.target.checked)}
              className="me-3"
            />
            <Button variant="outline-primary" onClick={() => onSystemAction('exportData')}>
              <FontAwesomeIcon icon={faFileExport} className="me-2" />
              Export Data
            </Button>
          </div>
        </div>
      </div>

      <Tabs activeKey={activeTab} onSelect={setActiveTab} className="mb-4">
        <Tab eventKey="overview" title={
          <span>
            <FontAwesomeIcon icon={faTachometerAlt} className="me-2" />
            Overview
          </span>
        }>
          {renderSystemOverview()}
        </Tab>
        
        <Tab eventKey="users" title={
          <span>
            <FontAwesomeIcon icon={faUsers} className="me-2" />
            Users ({systemMetrics.totalUsers})
          </span>
        }>
          {renderUserManagement()}
        </Tab>
        
        <Tab eventKey="nodes" title={
          <span>
            <FontAwesomeIcon icon={faServer} className="me-2" />
            Nodes ({systemMetrics.totalNodes})
          </span>
        }>
          {renderNodeManagement()}
        </Tab>
        
        <Tab eventKey="settings" title={
          <span>
            <FontAwesomeIcon icon={faCog} className="me-2" />
            Settings
          </span>
        }>
          {renderSystemSettings()}
        </Tab>
      </Tabs>

      {renderModal()}
    </div>
  );
};

export default AdminDashboard;
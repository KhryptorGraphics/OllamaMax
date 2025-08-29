import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Button, Badge, Modal, Tabs, Tab } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  faTachometerAlt,
  faServer,
  faDatabase,
  faChartLine,
  faUsers,
  faNetworkWired,
  faCog,
  faExpand,
  faDownload,
  faSync,
  faPlay,
  faPause,
  faExclamationTriangle,
  faCheckCircle,
  faInfoCircle,
  faRocket,
  faBolt,
  faShieldAlt,
  faGlobe,
  faHeart,
  faStar,
  faFire
} from '@fortawesome/free-solid-svg-icons';
import AdvancedCharts from './AdvancedCharts';
import DataVisualization from './DataVisualization';
import PerformanceMonitor from './PerformanceMonitor';
import LoadingSpinner from './LoadingSpinner';

const EnhancedDashboard = ({
  clusterStatus = {},
  nodes = [],
  models = [],
  metrics = {},
  realTimeMetrics = {},
  users = [],
  alerts = [],
  onRefresh,
  loading = false,
  error = null,
  className = ""
}) => {
  const [activeTab, setActiveTab] = useState('overview');
  const [selectedTimeRange, setSelectedTimeRange] = useState('24h');
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [showMetricsModal, setShowMetricsModal] = useState(false);
  const [selectedMetric, setSelectedMetric] = useState(null);
  const [dashboardWidgets, setDashboardWidgets] = useState([
    { id: 'cluster-health', enabled: true, position: 1, size: 'medium' },
    { id: 'performance-metrics', enabled: true, position: 2, size: 'large' },
    { id: 'active-models', enabled: true, position: 3, size: 'medium' },
    { id: 'system-alerts', enabled: true, position: 4, size: 'small' },
    { id: 'node-status', enabled: true, position: 5, size: 'large' },
    { id: 'user-activity', enabled: true, position: 6, size: 'medium' }
  ]);
  const [dashboardStats, setDashboardStats] = useState({});

  // Calculate dashboard statistics
  useEffect(() => {
    const stats = calculateDashboardStats();
    setDashboardStats(stats);
  }, [clusterStatus, nodes, models, metrics, users]);

  // Auto-refresh data
  useEffect(() => {
    if (autoRefresh && onRefresh) {
      const interval = setInterval(() => {
        onRefresh();
      }, 30000); // Refresh every 30 seconds

      return () => clearInterval(interval);
    }
  }, [autoRefresh, onRefresh]);

  const calculateDashboardStats = () => {
    const totalNodes = nodes.length;
    const onlineNodes = nodes.filter(node => node.status === 'online').length;
    const totalModels = models.length;
    const activeModels = models.filter(model => model.status === 'running').length;
    const totalUsers = users.length;
    const activeUsers = users.filter(user => user.status === 'online').length;
    const criticalAlerts = alerts.filter(alert => alert.severity === 'critical').length;
    const warningAlerts = alerts.filter(alert => alert.severity === 'warning').length;
    
    const systemHealth = calculateSystemHealth();
    const performanceScore = calculatePerformanceScore();
    const uptime = calculateUptime();
    
    return {
      totalNodes,
      onlineNodes,
      nodeHealth: totalNodes > 0 ? (onlineNodes / totalNodes) * 100 : 0,
      totalModels,
      activeModels,
      modelUtilization: totalModels > 0 ? (activeModels / totalModels) * 100 : 0,
      totalUsers,
      activeUsers,
      userActivity: totalUsers > 0 ? (activeUsers / totalUsers) * 100 : 0,
      criticalAlerts,
      warningAlerts,
      systemHealth,
      performanceScore,
      uptime,
      totalRequests: metrics.totalRequests || 0,
      averageLatency: metrics.avgLatency || 0,
      errorRate: metrics.errorRate || 0,
      throughput: metrics.throughput || 0
    };
  };

  const calculateSystemHealth = () => {
    const factors = [
      dashboardStats.nodeHealth || 0,
      dashboardStats.modelUtilization || 0,
      Math.max(0, 100 - (dashboardStats.criticalAlerts * 20 + dashboardStats.warningAlerts * 10))
    ];
    return factors.reduce((sum, factor) => sum + factor, 0) / factors.length;
  };

  const calculatePerformanceScore = () => {
    const cpuScore = Math.max(0, 100 - (metrics.cpu_usage || 0));
    const memoryScore = Math.max(0, 100 - (metrics.memory_usage || 0));
    const latencyScore = Math.max(0, 100 - Math.min(100, (metrics.avgLatency || 0) / 10));
    return (cpuScore + memoryScore + latencyScore) / 3;
  };

  const calculateUptime = () => {
    // Mock calculation - in real app would be based on actual uptime data
    return 99.8;
  };

  const getHealthColor = (percentage) => {
    if (percentage >= 90) return 'success';
    if (percentage >= 70) return 'warning';
    return 'danger';
  };

  const getHealthIcon = (percentage) => {
    if (percentage >= 90) return faCheckCircle;
    if (percentage >= 70) return faExclamationTriangle;
    return faExclamationTriangle;
  };

  const MetricCard = ({ title, value, subtitle, trend, color = 'primary', icon, onClick }) => (
    <Card 
      className={`metric-card h-100 ${onClick ? 'clickable' : ''}`} 
      onClick={onClick}
      style={{ cursor: onClick ? 'pointer' : 'default' }}
    >
      <Card.Body className="d-flex align-items-center">
        <div className="flex-grow-1">
          <div className="d-flex align-items-center justify-content-between mb-2">
            <h6 className="mb-0 text-muted">{title}</h6>
            {icon && (
              <FontAwesomeIcon 
                icon={icon} 
                className={`text-${color}`} 
              />
            )}
          </div>
          <div className="d-flex align-items-end justify-content-between">
            <div>
              <h3 className={`mb-0 text-${color}`}>{value}</h3>
              {subtitle && <small className="text-muted">{subtitle}</small>}
            </div>
            {trend && (
              <Badge bg={trend > 0 ? 'success' : trend < 0 ? 'danger' : 'secondary'}>
                {trend > 0 ? '+' : ''}{trend}%
              </Badge>
            )}
          </div>
        </div>
      </Card.Body>
    </Card>
  );

  const ClusterHealthWidget = () => (
    <Card className="cluster-health-widget">
      <Card.Header className="d-flex justify-content-between align-items-center">
        <h6 className="mb-0">
          <FontAwesomeIcon icon={faShieldAlt} className="me-2" />
          Cluster Health
        </h6>
        <Badge bg={getHealthColor(dashboardStats.systemHealth)}>
          {dashboardStats.systemHealth?.toFixed(1)}%
        </Badge>
      </Card.Header>
      <Card.Body>
        <Row>
          <Col md={6}>
            <div className="health-indicator mb-3">
              <div className="d-flex justify-content-between align-items-center mb-2">
                <span>Node Health</span>
                <span>{dashboardStats.nodeHealth?.toFixed(1)}%</span>
              </div>
              <div className="progress" style={{ height: '8px' }}>
                <div 
                  className={`progress-bar bg-${getHealthColor(dashboardStats.nodeHealth)}`}
                  style={{ width: `${dashboardStats.nodeHealth}%` }}
                ></div>
              </div>
            </div>
            
            <div className="health-indicator mb-3">
              <div className="d-flex justify-content-between align-items-center mb-2">
                <span>Model Utilization</span>
                <span>{dashboardStats.modelUtilization?.toFixed(1)}%</span>
              </div>
              <div className="progress" style={{ height: '8px' }}>
                <div 
                  className={`progress-bar bg-${getHealthColor(dashboardStats.modelUtilization)}`}
                  style={{ width: `${dashboardStats.modelUtilization}%` }}
                ></div>
              </div>
            </div>
          </Col>
          
          <Col md={6}>
            <div className="health-summary">
              <div className="d-flex align-items-center mb-2">
                <FontAwesome icon={faServer} className="text-primary me-2" />
                <span>{dashboardStats.onlineNodes}/{dashboardStats.totalNodes} Nodes Online</span>
              </div>
              <div className="d-flex align-items-center mb-2">
                <FontAwesome icon={faDatabase} className="text-success me-2" />
                <span>{dashboardStats.activeModels}/{dashboardStats.totalModels} Models Active</span>
              </div>
              <div className="d-flex align-items-center mb-2">
                <FontAwesome icon={faUsers} className="text-info me-2" />
                <span>{dashboardStats.activeUsers}/{dashboardStats.totalUsers} Users Active</span>
              </div>
              <div className="d-flex align-items-center">
                <FontAwesome icon={faExclamationTriangle} className="text-warning me-2" />
                <span>{dashboardStats.criticalAlerts} Critical, {dashboardStats.warningAlerts} Warnings</span>
              </div>
            </div>
          </Col>
        </Row>
      </Card.Body>
    </Card>
  );

  const PerformanceOverviewWidget = () => (
    <Card className="performance-overview-widget">
      <Card.Header className="d-flex justify-content-between align-items-center">
        <h6 className="mb-0">
          <FontAwesome icon={faTachometerAlt} className="me-2" />
          Performance Overview
        </h6>
        <Badge bg={getHealthColor(dashboardStats.performanceScore)}>
          Score: {dashboardStats.performanceScore?.toFixed(0)}/100
        </Badge>
      </Card.Header>
      <Card.Body>
        <Row>
          <Col lg={8}>
            <AdvancedCharts
              data={realTimeMetrics.cpu || []}
              title="System Performance"
              chartType="area"
              height={250}
              realTime={autoRefresh}
              interactive={true}
            />
          </Col>
          <Col lg={4}>
            <div className="performance-stats">
              <div className="stat-item mb-3">
                <div className="d-flex justify-content-between align-items-center">
                  <span>CPU Usage</span>
                  <Badge bg={metrics.cpu_usage > 80 ? 'danger' : metrics.cpu_usage > 60 ? 'warning' : 'success'}>
                    {metrics.cpu_usage || 0}%
                  </Badge>
                </div>
              </div>
              <div className="stat-item mb-3">
                <div className="d-flex justify-content-between align-items-center">
                  <span>Memory Usage</span>
                  <Badge bg={metrics.memory_usage > 80 ? 'danger' : metrics.memory_usage > 60 ? 'warning' : 'success'}>
                    {metrics.memory_usage || 0}%
                  </Badge>
                </div>
              </div>
              <div className="stat-item mb-3">
                <div className="d-flex justify-content-between align-items-center">
                  <span>Network I/O</span>
                  <Badge bg="info">
                    {metrics.network_usage || 0}%
                  </Badge>
                </div>
              </div>
              <div className="stat-item mb-3">
                <div className="d-flex justify-content-between align-items-center">
                  <span>Avg Latency</span>
                  <Badge bg={metrics.avgLatency > 500 ? 'danger' : metrics.avgLatency > 200 ? 'warning' : 'success'}>
                    {metrics.avgLatency || 0}ms
                  </Badge>
                </div>
              </div>
            </div>
          </Col>
        </Row>
      </Card.Body>
    </Card>
  );

  const QuickStatsWidget = () => (
    <Row className="mb-4">
      <Col md={3}>
        <MetricCard
          title="Total Requests"
          value={dashboardStats.totalRequests?.toLocaleString() || '0'}
          subtitle="Today"
          trend={Math.floor(Math.random() * 20) - 10}
          icon={faRocket}
          color="primary"
          onClick={() => {
            setSelectedMetric('requests');
            setShowMetricsModal(true);
          }}
        />
      </Col>
      <Col md={3}>
        <MetricCard
          title="System Uptime"
          value={`${dashboardStats.uptime?.toFixed(1) || 0}%`}
          subtitle="Last 30 days"
          trend={0.2}
          icon={faHeart}
          color="success"
          onClick={() => {
            setSelectedMetric('uptime');
            setShowMetricsModal(true);
          }}
        />
      </Col>
      <Col md={3}>
        <MetricCard
          title="Error Rate"
          value={`${dashboardStats.errorRate?.toFixed(2) || 0}%`}
          subtitle="Last hour"
          trend={-0.5}
          icon={faExclamationTriangle}
          color="warning"
          onClick={() => {
            setSelectedMetric('errors');
            setShowMetricsModal(true);
          }}
        />
      </Col>
      <Col md={3}>
        <MetricCard
          title="Throughput"
          value={`${dashboardStats.throughput?.toLocaleString() || 0}`}
          subtitle="ops/sec"
          trend={5.2}
          icon={faBolt}
          color="info"
          onClick={() => {
            setSelectedMetric('throughput');
            setShowMetricsModal(true);
          }}
        />
      </Col>
    </Row>
  );

  const SystemAlertsWidget = () => (
    <Card className="system-alerts-widget">
      <Card.Header className="d-flex justify-content-between align-items-center">
        <h6 className="mb-0">
          <FontAwesome icon={faBell} className="me-2" />
          System Alerts
        </h6>
        <Badge bg={alerts.length > 0 ? 'danger' : 'success'}>
          {alerts.length}
        </Badge>
      </Card.Header>
      <Card.Body style={{ maxHeight: '300px', overflowY: 'auto' }}>
        {alerts.length === 0 ? (
          <div className="text-center py-3">
            <FontAwesome icon={faCheckCircle} size="2x" className="text-success mb-2" />
            <p className="text-muted mb-0">All systems operational</p>
          </div>
        ) : (
          alerts.slice(0, 5).map(alert => (
            <div key={alert.id} className={`alert alert-${alert.severity === 'critical' ? 'danger' : 'warning'} alert-dismissible fade show mb-2`}>
              <div className="d-flex align-items-start">
                <FontAwesome 
                  icon={alert.severity === 'critical' ? faExclamationTriangle : faInfoCircle} 
                  className="me-2 mt-1" 
                />
                <div className="flex-grow-1">
                  <strong>{alert.title}</strong>
                  <p className="mb-1 small">{alert.message}</p>
                  <small className="text-muted">
                    {new Date(alert.timestamp).toLocaleString()}
                  </small>
                </div>
              </div>
            </div>
          ))
        )}
      </Card.Body>
    </Card>
  );

  const ModelStatusWidget = () => (
    <Card className="model-status-widget">
      <Card.Header className="d-flex justify-content-between align-items-center">
        <h6 className="mb-0">
          <FontAwesome icon={faDatabase} className="me-2" />
          Active Models
        </h6>
        <Button variant="outline-primary" size="sm">
          Manage
        </Button>
      </Card.Header>
      <Card.Body>
        <div className="model-list">
          {models.slice(0, 5).map(model => (
            <div key={model.name} className="d-flex justify-content-between align-items-center mb-3">
              <div>
                <div className="fw-medium">{model.name}</div>
                <small className="text-muted">{model.replicas?.length || 0} replicas</small>
              </div>
              <Badge bg={model.status === 'running' ? 'success' : model.status === 'error' ? 'danger' : 'secondary'}>
                {model.status || 'unknown'}
              </Badge>
            </div>
          ))}
        </div>
        {models.length > 5 && (
          <div className="text-center">
            <small className="text-muted">+{models.length - 5} more models</small>
          </div>
        )}
      </Card.Body>
    </Card>
  );

  const MetricsModal = () => (
    <Modal show={showMetricsModal} onHide={() => setShowMetricsModal(false)} size="xl">
      <Modal.Header closeButton>
        <Modal.Title>
          <FontAwesome icon={faChartLine} className="me-2" />
          Detailed Metrics - {selectedMetric}
        </Modal.Title>
      </Modal.Header>
      <Modal.Body>
        <PerformanceMonitor
          nodes={nodes}
          metrics={metrics}
          alerts={alerts}
          realTimeEnabled={autoRefresh}
        />
      </Modal.Body>
      <Modal.Footer>
        <Button variant="secondary" onClick={() => setShowMetricsModal(false)}>
          Close
        </Button>
        <Button variant="primary" onClick={() => {
          // Export metrics functionality
        }}>
          <FontAwesome icon={faDownload} className="me-1" />
          Export Data
        </Button>
      </Modal.Footer>
    </Modal>
  );

  if (loading) {
    return <LoadingSpinner size="xl" text="Loading enhanced dashboard..." />;
  }

  if (error) {
    return (
      <Alert variant="danger">
        <Alert.Heading>Dashboard Error</Alert.Heading>
        <p>{error}</p>
      </Alert>
    );
  }

  return (
    <div className={`enhanced-dashboard ${className}`}>
      {/* Dashboard Header */}
      <div className="dashboard-header d-flex justify-content-between align-items-center mb-4">
        <div>
          <h2 className="mb-1">
            <FontAwesome icon={faTachometerAlt} className="me-2" />
            Enhanced Dashboard
          </h2>
          <p className="text-muted mb-0">
            Real-time monitoring and analytics for your OllamaMax cluster
          </p>
        </div>
        
        <div className="dashboard-controls d-flex align-items-center gap-2">
          <Button 
            variant={autoRefresh ? 'success' : 'outline-secondary'} 
            size="sm"
            onClick={() => setAutoRefresh(!autoRefresh)}
          >
            <FontAwesome icon={autoRefresh ? faPause : faPlay} className="me-1" />
            {autoRefresh ? 'Pause' : 'Resume'}
          </Button>
          
          <Button variant="outline-primary" size="sm" onClick={onRefresh}>
            <FontAwesome icon={faSync} className="me-1" />
            Refresh
          </Button>
          
          <Button variant="outline-secondary" size="sm">
            <FontAwesome icon={faCog} className="me-1" />
            Settings
          </Button>
        </div>
      </div>

      {/* Quick Stats */}
      <QuickStatsWidget />

      {/* Main Dashboard Content */}
      <Tabs activeKey={activeTab} onSelect={setActiveTab} className="mb-4">
        <Tab eventKey="overview" title="Overview">
          <Row>
            <Col lg={8}>
              <PerformanceOverviewWidget />
            </Col>
            <Col lg={4}>
              <ClusterHealthWidget />
            </Col>
          </Row>
          
          <Row className="mt-4">
            <Col lg={6}>
              <ModelStatusWidget />
            </Col>
            <Col lg={6}>
              <SystemAlertsWidget />
            </Col>
          </Row>
        </Tab>
        
        <Tab eventKey="metrics" title="Detailed Metrics">
          <DataVisualization
            data={realTimeMetrics}
            loading={false}
            error={null}
            onDataRefresh={onRefresh}
          />
        </Tab>
        
        <Tab eventKey="performance" title="Performance Monitor">
          <PerformanceMonitor
            nodes={nodes}
            metrics={metrics}
            alerts={alerts}
            realTimeEnabled={autoRefresh}
          />
        </Tab>
      </Tabs>

      {/* Metrics Detail Modal */}
      <MetricsModal />
    </div>
  );
};

// Helper component for FontAwesome
const FontAwesome = ({ icon, ...props }) => (
  <FontAwesome icon={icon} {...props} />
);

export default EnhancedDashboard;
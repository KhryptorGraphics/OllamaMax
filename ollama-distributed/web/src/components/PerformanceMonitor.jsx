import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Button, Form, Alert, Badge, Table, Modal, ProgressBar } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  faTachometerAlt,
  faMemory,
  faMicrochip,
  faNetworkWired,
  faHdd,
  faServer,
  faExclamationTriangle,
  faCheckCircle,
  faSpinner,
  faCog,
  faDownload,
  faSync,
  faExpand,
  faCompress,
  faPlay,
  faPause,
  faStop,
  faChartLine,
  faDatabase,
  faRocket,
  faShieldAlt,
  faBolt,
  faWifi,
  faThermometerHalf,
  faFan
} from '@fortawesome/free-solid-svg-icons';
import AdvancedCharts from './AdvancedCharts';
import LoadingSpinner from './LoadingSpinner';

const PerformanceMonitor = ({
  nodes = [],
  metrics = {},
  alerts = [],
  onConfigUpdate,
  onAlertDismiss,
  onMetricExport,
  realTimeEnabled = true,
  loading = false,
  error = null,
  className = ""
}) => {
  const [selectedNode, setSelectedNode] = useState(null);
  const [timeRange, setTimeRange] = useState('1h');
  const [refreshInterval, setRefreshInterval] = useState(5000);
  const [autoRefresh, setAutoRefresh] = useState(realTimeEnabled);
  const [showSettingsModal, setShowSettingsModal] = useState(false);
  const [showNodeModal, setShowNodeModal] = useState(false);
  const [performanceData, setPerformanceData] = useState({});
  const [systemAlerts, setSystemAlerts] = useState([]);
  const [thresholds, setThresholds] = useState({
    cpu: { warning: 70, critical: 90 },
    memory: { warning: 80, critical: 95 },
    disk: { warning: 85, critical: 95 },
    network: { warning: 80, critical: 95 },
    temperature: { warning: 70, critical: 85 },
    latency: { warning: 500, critical: 1000 }
  });

  const timeRanges = [
    { value: '5m', label: '5 Minutes', points: 25 },
    { value: '15m', label: '15 Minutes', points: 75 },
    { value: '1h', label: '1 Hour', points: 60 },
    { value: '6h', label: '6 Hours', points: 72 },
    { value: '24h', label: '24 Hours', points: 144 },
    { value: '7d', label: '7 Days', points: 168 }
  ];

  const metricDefinitions = [
    {
      id: 'cpu',
      label: 'CPU Usage',
      icon: faMicrochip,
      unit: '%',
      color: '#3B82F6',
      description: 'Processor utilization across all cores'
    },
    {
      id: 'memory',
      label: 'Memory Usage',
      icon: faMemory,
      unit: '%',
      color: '#10B981',
      description: 'RAM utilization and availability'
    },
    {
      id: 'disk',
      label: 'Disk I/O',
      icon: faHdd,
      unit: 'MB/s',
      color: '#F59E0B',
      description: 'Storage read/write operations'
    },
    {
      id: 'network',
      label: 'Network I/O',
      icon: faNetworkWired,
      unit: 'MB/s',
      color: '#EF4444',
      description: 'Network bandwidth utilization'
    },
    {
      id: 'temperature',
      label: 'Temperature',
      icon: faThermometerHalf,
      unit: '°C',
      color: '#F97316',
      description: 'System temperature monitoring'
    },
    {
      id: 'latency',
      label: 'Response Latency',
      icon: faBolt,
      unit: 'ms',
      color: '#8B5CF6',
      description: 'API response time metrics'
    }
  ];

  // Generate mock performance data
  useEffect(() => {
    const generatePerformanceData = () => {
      const data = {};
      const currentRange = timeRanges.find(r => r.value === timeRange);
      const points = currentRange?.points || 60;
      const now = Date.now();
      const interval = getRangeInterval(timeRange);

      metricDefinitions.forEach(metric => {
        data[metric.id] = generateMetricData(metric.id, points, now, interval);
      });

      return data;
    };

    const mockData = generatePerformanceData();
    setPerformanceData(mockData);

    // Generate system alerts
    const mockAlerts = generateSystemAlerts(mockData);
    setSystemAlerts(mockAlerts);
  }, [timeRange, nodes]);

  // Auto-refresh mechanism
  useEffect(() => {
    if (autoRefresh) {
      const interval = setInterval(() => {
        // In real app, this would fetch fresh data
        const newData = generateRealtimeUpdate();
        setPerformanceData(prev => updateRealtimeData(prev, newData));
      }, refreshInterval);

      return () => clearInterval(interval);
    }
  }, [autoRefresh, refreshInterval]);

  const getRangeInterval = (range) => {
    const intervals = {
      '5m': 12000,   // 12 seconds
      '15m': 12000,  // 12 seconds
      '1h': 60000,   // 1 minute
      '6h': 300000,  // 5 minutes
      '24h': 600000, // 10 minutes
      '7d': 3600000  // 1 hour
    };
    return intervals[range] || 60000;
  };

  const generateMetricData = (metricId, points, now, interval) => {
    const baseValues = {
      cpu: 45, memory: 65, disk: 25, network: 35, temperature: 45, latency: 150
    };
    const variations = {
      cpu: 30, memory: 20, disk: 40, network: 25, temperature: 15, latency: 100
    };

    return Array.from({ length: points }, (_, i) => {
      const timestamp = now - ((points - i - 1) * interval);
      const base = baseValues[metricId] || 50;
      const variation = variations[metricId] || 20;
      const trend = Math.sin(i * 0.1) * 10;
      const noise = (Math.random() - 0.5) * variation;
      
      return {
        timestamp,
        x: new Date(timestamp).toLocaleTimeString(),
        value: Math.max(0, base + trend + noise),
        nodeId: selectedNode || 'cluster-avg'
      };
    });
  };

  const generateRealtimeUpdate = () => {
    const update = {};
    metricDefinitions.forEach(metric => {
      const base = {
        cpu: 45, memory: 65, disk: 25, network: 35, temperature: 45, latency: 150
      }[metric.id] || 50;
      const variation = {
        cpu: 30, memory: 20, disk: 40, network: 25, temperature: 15, latency: 100
      }[metric.id] || 20;
      
      update[metric.id] = {
        timestamp: Date.now(),
        x: new Date().toLocaleTimeString(),
        value: Math.max(0, base + (Math.random() - 0.5) * variation),
        nodeId: selectedNode || 'cluster-avg'
      };
    });
    return update;
  };

  const updateRealtimeData = (prevData, newData) => {
    const updated = { ...prevData };
    Object.keys(newData).forEach(metricId => {
      if (updated[metricId]) {
        updated[metricId] = [...updated[metricId].slice(-59), newData[metricId]];
      }
    });
    return updated;
  };

  const generateSystemAlerts = (data) => {
    const alerts = [];
    Object.keys(data).forEach(metricId => {
      const metric = metricDefinitions.find(m => m.id === metricId);
      const threshold = thresholds[metricId];
      if (!metric || !threshold || !data[metricId]) return;

      const latestValue = data[metricId][data[metricId].length - 1]?.value || 0;
      
      if (latestValue > threshold.critical) {
        alerts.push({
          id: `${metricId}-critical-${Date.now()}`,
          type: 'critical',
          metric: metric.label,
          value: latestValue.toFixed(1),
          threshold: threshold.critical,
          message: `${metric.label} is critically high: ${latestValue.toFixed(1)}${metric.unit}`,
          timestamp: Date.now(),
          nodeId: selectedNode || 'cluster'
        });
      } else if (latestValue > threshold.warning) {
        alerts.push({
          id: `${metricId}-warning-${Date.now()}`,
          type: 'warning',
          metric: metric.label,
          value: latestValue.toFixed(1),
          threshold: threshold.warning,
          message: `${metric.label} is above warning threshold: ${latestValue.toFixed(1)}${metric.unit}`,
          timestamp: Date.now(),
          nodeId: selectedNode || 'cluster'
        });
      }
    });
    return alerts;
  };

  const getMetricStatus = (metricId, value) => {
    const threshold = thresholds[metricId];
    if (!threshold) return 'normal';
    
    if (value > threshold.critical) return 'critical';
    if (value > threshold.warning) return 'warning';
    return 'normal';
  };

  const getStatusColor = (status) => {
    const colors = {
      normal: 'success',
      warning: 'warning',
      critical: 'danger'
    };
    return colors[status] || 'secondary';
  };

  const calculateAverages = () => {
    const averages = {};
    metricDefinitions.forEach(metric => {
      if (performanceData[metric.id] && performanceData[metric.id].length > 0) {
        const values = performanceData[metric.id].map(d => d.value);
        const sum = values.reduce((a, b) => a + b, 0);
        averages[metric.id] = {
          current: values[values.length - 1] || 0,
          average: sum / values.length,
          maximum: Math.max(...values),
          minimum: Math.min(...values)
        };
      }
    });
    return averages;
  };

  const renderMetricCard = (metric) => {
    const averages = calculateAverages();
    const metricAvg = averages[metric.id] || { current: 0, average: 0, maximum: 0, minimum: 0 };
    const status = getMetricStatus(metric.id, metricAvg.current);
    const statusColor = getStatusColor(status);
    const threshold = thresholds[metric.id];
    
    return (
      <Card key={metric.id} className="metric-card h-100">
        <Card.Header className="d-flex justify-content-between align-items-center">
          <div className="d-flex align-items-center">
            <FontAwesomeIcon 
              icon={metric.icon} 
              className="me-2" 
              style={{ color: metric.color }} 
            />
            <h6 className="mb-0">{metric.label}</h6>
          </div>
          <Badge bg={statusColor}>
            {status.toUpperCase()}
          </Badge>
        </Card.Header>
        
        <Card.Body>
          <div className="metric-summary mb-3">
            <div className="current-value mb-2">
              <h3 className="mb-1" style={{ color: metric.color }}>
                {metricAvg.current.toFixed(1)}
                <small className="text-muted ms-1">{metric.unit}</small>
              </h3>
            </div>
            
            <div className="metric-stats">
              <div className="d-flex justify-content-between mb-1">
                <small className="text-muted">Average:</small>
                <small>{metricAvg.average.toFixed(1)}{metric.unit}</small>
              </div>
              <div className="d-flex justify-content-between mb-1">
                <small className="text-muted">Maximum:</small>
                <small>{metricAvg.maximum.toFixed(1)}{metric.unit}</small>
              </div>
              <div className="d-flex justify-content-between mb-1">
                <small className="text-muted">Minimum:</small>
                <small>{metricAvg.minimum.toFixed(1)}{metric.unit}</small>
              </div>
            </div>
          </div>
          
          {threshold && (
            <div className="thresholds mb-3">
              <div className="d-flex justify-content-between align-items-center mb-1">
                <small className="text-warning">Warning:</small>
                <small>{threshold.warning}{metric.unit}</small>
              </div>
              <div className="d-flex justify-content-between align-items-center">
                <small className="text-danger">Critical:</small>
                <small>{threshold.critical}{metric.unit}</small>
              </div>
              
              <div className="threshold-bar mt-2">
                <ProgressBar className="threshold-progress">
                  <ProgressBar 
                    variant="success" 
                    now={Math.min((threshold.warning / 100) * 100, 100)} 
                    key={1} 
                  />
                  <ProgressBar 
                    variant="warning" 
                    now={Math.min(((threshold.critical - threshold.warning) / 100) * 100, 100)} 
                    key={2} 
                  />
                  <ProgressBar 
                    variant="danger" 
                    now={Math.min(((100 - threshold.critical) / 100) * 100, 100)} 
                    key={3} 
                  />
                </ProgressBar>
                <div 
                  className="current-marker"
                  style={{
                    position: 'absolute',
                    left: `${Math.min(metricAvg.current, 100)}%`,
                    top: 0,
                    width: '2px',
                    height: '100%',
                    backgroundColor: metric.color,
                    zIndex: 10
                  }}
                ></div>
              </div>
            </div>
          )}
        </Card.Body>
      </Card>
    );
  };

  const renderAlertsPanel = () => {
    const allAlerts = [...systemAlerts, ...alerts].sort((a, b) => b.timestamp - a.timestamp);
    
    return (
      <Card className="alerts-panel">
        <Card.Header className="d-flex justify-content-between align-items-center">
          <h6 className="mb-0">
            <FontAwesomeIcon icon={faExclamationTriangle} className="me-2" />
            System Alerts ({allAlerts.length})
          </h6>
          {allAlerts.length > 0 && (
            <Button 
              variant="outline-secondary" 
              size="sm"
              onClick={() => setSystemAlerts([])}
            >
              Clear All
            </Button>
          )}
        </Card.Header>
        
        <Card.Body style={{ maxHeight: '300px', overflowY: 'auto' }}>
          {allAlerts.length === 0 ? (
            <div className="text-center text-muted py-3">
              <FontAwesomeIcon icon={faCheckCircle} size="2x" className="mb-2" />
              <p className="mb-0">No active alerts</p>
            </div>
          ) : (
            allAlerts.map(alert => (
              <Alert 
                key={alert.id} 
                variant={alert.type === 'critical' ? 'danger' : 'warning'}
                dismissible
                onClose={() => {
                  setSystemAlerts(prev => prev.filter(a => a.id !== alert.id));
                  if (onAlertDismiss) onAlertDismiss(alert.id);
                }}
                className="mb-2"
              >
                <div className="d-flex justify-content-between align-items-start">
                  <div>
                    <Alert.Heading as="h6">{alert.metric}</Alert.Heading>
                    <p className="mb-1">{alert.message}</p>
                    <small className="text-muted">
                      {alert.nodeId} • {new Date(alert.timestamp).toLocaleTimeString()}
                    </small>
                  </div>
                </div>
              </Alert>
            ))
          )}
        </Card.Body>
      </Card>
    );
  };

  const renderNodeSelector = () => {
    const nodeOptions = [
      { id: null, name: 'Cluster Average', status: 'online' },
      ...nodes.map(node => ({ id: node.id, name: node.id, status: node.status || 'unknown' }))
    ];
    
    return (
      <Form.Select 
        value={selectedNode || ''} 
        onChange={(e) => setSelectedNode(e.target.value || null)}
        style={{ width: '200px' }}
      >
        {nodeOptions.map(option => (
          <option key={option.id || 'cluster'} value={option.id || ''}>
            {option.name} {option.status === 'online' ? '✓' : '✗'}
          </option>
        ))}
      </Form.Select>
    );
  };

  const renderSettingsModal = () => {
    return (
      <Modal show={showSettingsModal} onHide={() => setShowSettingsModal(false)} size="lg">
        <Modal.Header closeButton>
          <Modal.Title>
            <FontAwesomeIcon icon={faCog} className="me-2" />
            Performance Monitor Settings
          </Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <Row>
            <Col md={6}>
              <h6>Refresh Settings</h6>
              <Form.Group className="mb-3">
                <Form.Label>Refresh Interval</Form.Label>
                <Form.Select 
                  value={refreshInterval} 
                  onChange={(e) => setRefreshInterval(Number(e.target.value))}
                >
                  <option value={1000}>1 second</option>
                  <option value={5000}>5 seconds</option>
                  <option value={10000}>10 seconds</option>
                  <option value={30000}>30 seconds</option>
                  <option value={60000}>1 minute</option>
                </Form.Select>
              </Form.Group>
              
              <Form.Group className="mb-3">
                <Form.Check 
                  type="checkbox"
                  label="Auto Refresh"
                  checked={autoRefresh}
                  onChange={(e) => setAutoRefresh(e.target.checked)}
                />
              </Form.Group>
            </Col>
            
            <Col md={6}>
              <h6>Alert Thresholds</h6>
              {metricDefinitions.slice(0, 4).map(metric => (
                <div key={metric.id} className="mb-3">
                  <Form.Label>{metric.label}</Form.Label>
                  <Row>
                    <Col>
                      <Form.Control
                        type="number"
                        placeholder="Warning"
                        value={thresholds[metric.id]?.warning || 70}
                        onChange={(e) => setThresholds(prev => ({
                          ...prev,
                          [metric.id]: {
                            ...prev[metric.id],
                            warning: Number(e.target.value)
                          }
                        }))}
                        size="sm"
                      />
                    </Col>
                    <Col>
                      <Form.Control
                        type="number"
                        placeholder="Critical"
                        value={thresholds[metric.id]?.critical || 90}
                        onChange={(e) => setThresholds(prev => ({
                          ...prev,
                          [metric.id]: {
                            ...prev[metric.id],
                            critical: Number(e.target.value)
                          }
                        }))}
                        size="sm"
                      />
                    </Col>
                  </Row>
                </div>
              ))}
            </Col>
          </Row>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => setShowSettingsModal(false)}>
            Cancel
          </Button>
          <Button variant="primary" onClick={() => {
            if (onConfigUpdate) {
              onConfigUpdate({ refreshInterval, autoRefresh, thresholds });
            }
            setShowSettingsModal(false);
          }}>
            Save Settings
          </Button>
        </Modal.Footer>
      </Modal>
    );
  };

  const handleExportMetrics = () => {
    const exportData = {
      timestamp: new Date().toISOString(),
      timeRange,
      selectedNode: selectedNode || 'cluster',
      metrics: performanceData,
      averages: calculateAverages(),
      alerts: systemAlerts,
      thresholds
    };

    if (onMetricExport) {
      onMetricExport(exportData);
    } else {
      const blob = new Blob([JSON.stringify(exportData, null, 2)], {
        type: 'application/json'
      });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `performance-metrics-${timeRange}-${Date.now()}.json`;
      a.click();
      URL.revokeObjectURL(url);
    }
  };

  if (loading) {
    return <LoadingSpinner size="xl" text="Loading performance monitor..." />;
  }

  if (error) {
    return (
      <Alert variant="danger">
        <Alert.Heading>Performance Monitor Error</Alert.Heading>
        <p>{error}</p>
      </Alert>
    );
  }

  return (
    <div className={`performance-monitor ${className}`}>
      {/* Header */}
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h2>
          <FontAwesomeIcon icon={faTachometerAlt} className="me-2" />
          Performance Monitor
        </h2>
        <div className="d-flex align-items-center gap-2">
          {renderNodeSelector()}
          
          <Form.Select 
            value={timeRange} 
            onChange={(e) => setTimeRange(e.target.value)}
            style={{ width: '150px' }}
          >
            {timeRanges.map(range => (
              <option key={range.value} value={range.value}>
                {range.label}
              </option>
            ))}
          </Form.Select>
          
          <Button 
            variant={autoRefresh ? 'success' : 'outline-secondary'} 
            size="sm"
            onClick={() => setAutoRefresh(!autoRefresh)}
            title="Toggle Auto Refresh"
          >
            <FontAwesomeIcon 
              icon={autoRefresh ? faPlay : faPause} 
              className={autoRefresh ? 'fa-spin' : ''} 
            />
          </Button>
          
          <Button variant="outline-primary" size="sm" onClick={() => setShowSettingsModal(true)}>
            <FontAwesomeIcon icon={faCog} />
          </Button>
          
          <Button variant="outline-success" size="sm" onClick={handleExportMetrics}>
            <FontAwesomeIcon icon={faDownload} />
          </Button>
        </div>
      </div>

      {/* Status Overview */}
      <Row className="mb-4">
        <Col lg={8}>
          <Row>
            {metricDefinitions.slice(0, 4).map(metric => (
              <Col key={metric.id} md={6} lg={3} className="mb-3">
                {renderMetricCard(metric)}
              </Col>
            ))}
          </Row>
        </Col>
        <Col lg={4}>
          {renderAlertsPanel()}
        </Col>
      </Row>

      {/* Detailed Charts */}
      <Row className="mb-4">
        {metricDefinitions.slice(0, 4).map(metric => (
          <Col key={`chart-${metric.id}`} lg={6} className="mb-4">
            <AdvancedCharts
              data={performanceData[metric.id] || []}
              title={`${metric.label} - ${selectedNode || 'Cluster'}`}
              chartType="area"
              height={300}
              realTime={autoRefresh}
              interactive={true}
            />
          </Col>
        ))}
      </Row>

      {/* Additional Metrics */}
      <Row>
        {metricDefinitions.slice(4).map(metric => (
          <Col key={`extra-${metric.id}`} lg={6} className="mb-4">
            <AdvancedCharts
              data={performanceData[metric.id] || []}
              title={metric.label}
              chartType="line"
              height={250}
              realTime={autoRefresh}
              interactive={true}
            />
          </Col>
        ))}
      </Row>

      {/* Settings Modal */}
      {renderSettingsModal()}
    </div>
  );
};

export default PerformanceMonitor;
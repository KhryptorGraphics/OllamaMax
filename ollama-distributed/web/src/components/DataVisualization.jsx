import React, { useState, useEffect, useMemo } from 'react';
import { Card, Row, Col, Form, Button, Badge, Table, Modal } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  faChartLine,
  faTable,
  faFilter,
  faDownload,
  faExpand,
  faSync,
  faEye,
  faCog,
  faDatabase,
  faTachometerAlt,
  faNetworkWired,
  faServer
} from '@fortawesome/free-solid-svg-icons';
import AdvancedCharts from './AdvancedCharts';
import LoadingSpinner from './LoadingSpinner';

const DataVisualization = ({ 
  data = {}, 
  loading = false, 
  error = null,
  onDataRefresh,
  className = ""
}) => {
  const [activeView, setActiveView] = useState('dashboard');
  const [selectedMetrics, setSelectedMetrics] = useState(['cpu', 'memory', 'network']);
  const [timeRange, setTimeRange] = useState('1h');
  const [refreshInterval, setRefreshInterval] = useState(30000);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [showSettings, setShowSettings] = useState(false);
  const [visualizationData, setVisualizationData] = useState({});
  const [filters, setFilters] = useState({});

  const views = [
    { id: 'dashboard', label: 'Dashboard', icon: faTachometerAlt },
    { id: 'metrics', label: 'Metrics', icon: faChartLine },
    { id: 'network', label: 'Network', icon: faNetworkWired },
    { id: 'nodes', label: 'Node Health', icon: faServer },
    { id: 'table', label: 'Data Table', icon: faTable }
  ];

  const availableMetrics = [
    { id: 'cpu', label: 'CPU Usage', color: '#3B82F6', unit: '%' },
    { id: 'memory', label: 'Memory Usage', color: '#10B981', unit: '%' },
    { id: 'network', label: 'Network I/O', color: '#F59E0B', unit: 'MB/s' },
    { id: 'disk', label: 'Disk I/O', color: '#EF4444', unit: 'MB/s' },
    { id: 'requests', label: 'Request Rate', color: '#8B5CF6', unit: '/s' },
    { id: 'latency', label: 'Response Time', color: '#06B6D4', unit: 'ms' },
    { id: 'errors', label: 'Error Rate', color: '#F97316', unit: '/s' },
    { id: 'throughput', label: 'Throughput', color: '#84CC16', unit: 'ops/s' }
  ];

  const timeRanges = [
    { value: '5m', label: '5 Minutes' },
    { value: '15m', label: '15 Minutes' },
    { value: '1h', label: '1 Hour' },
    { value: '6h', label: '6 Hours' },
    { value: '24h', label: '24 Hours' },
    { value: '7d', label: '7 Days' },
    { value: '30d', label: '30 Days' }
  ];

  // Process and generate visualization data
  useEffect(() => {
    if (data && Object.keys(data).length > 0) {
      const processedData = processVisualizationData(data, selectedMetrics, timeRange);
      setVisualizationData(processedData);
    } else {
      // Generate mock data for demonstration
      const mockData = generateMockData(selectedMetrics, timeRange);
      setVisualizationData(mockData);
    }
  }, [data, selectedMetrics, timeRange]);

  // Auto-refresh mechanism
  useEffect(() => {
    if (autoRefresh && onDataRefresh) {
      const interval = setInterval(() => {
        onDataRefresh();
      }, refreshInterval);

      return () => clearInterval(interval);
    }
  }, [autoRefresh, refreshInterval, onDataRefresh]);

  const processVisualizationData = (rawData, metrics, range) => {
    const processed = {};
    
    metrics.forEach(metricId => {
      const metric = availableMetrics.find(m => m.id === metricId);
      if (metric && rawData[metricId]) {
        processed[metricId] = {
          ...metric,
          data: rawData[metricId],
          summary: calculateSummaryStats(rawData[metricId])
        };
      }
    });

    return processed;
  };

  const generateMockData = (metrics, range) => {
    const dataPoints = getDataPointsForRange(range);
    const mockData = {};

    metrics.forEach(metricId => {
      const metric = availableMetrics.find(m => m.id === metricId);
      if (metric) {
        const data = generateTimeSeriesData(dataPoints, metricId);
        mockData[metricId] = {
          ...metric,
          data: data,
          summary: calculateSummaryStats(data)
        };
      }
    });

    return mockData;
  };

  const getDataPointsForRange = (range) => {
    const ranges = {
      '5m': 25,
      '15m': 75,
      '1h': 60,
      '6h': 72,
      '24h': 144,
      '7d': 168,
      '30d': 360
    };
    return ranges[range] || 60;
  };

  const generateTimeSeriesData = (points, metricId) => {
    const now = Date.now();
    const interval = getRangeInterval(timeRange);
    
    return Array.from({ length: points }, (_, i) => {
      const timestamp = now - ((points - i - 1) * interval);
      const baseValue = getBaseValueForMetric(metricId);
      const variation = getVariationForMetric(metricId);
      const trend = getTrendForMetric(metricId, i, points);
      
      return {
        timestamp,
        x: new Date(timestamp).toLocaleTimeString(),
        value: Math.max(0, baseValue + trend + (Math.random() - 0.5) * variation),
        category: getCategoryForValue(metricId)
      };
    });
  };

  const getRangeInterval = (range) => {
    const intervals = {
      '5m': 12000,  // 12 seconds
      '15m': 12000, // 12 seconds
      '1h': 60000,  // 1 minute
      '6h': 300000, // 5 minutes
      '24h': 600000, // 10 minutes
      '7d': 3600000, // 1 hour
      '30d': 7200000 // 2 hours
    };
    return intervals[range] || 60000;
  };

  const getBaseValueForMetric = (metricId) => {
    const bases = {
      cpu: 45,
      memory: 65,
      network: 25,
      disk: 15,
      requests: 120,
      latency: 150,
      errors: 2,
      throughput: 80
    };
    return bases[metricId] || 50;
  };

  const getVariationForMetric = (metricId) => {
    const variations = {
      cpu: 30,
      memory: 20,
      network: 40,
      disk: 25,
      requests: 60,
      latency: 100,
      errors: 8,
      throughput: 40
    };
    return variations[metricId] || 20;
  };

  const getTrendForMetric = (metricId, index, totalPoints) => {
    // Add some trend patterns for different metrics
    const progress = index / totalPoints;
    const patterns = {
      cpu: Math.sin(progress * Math.PI * 2) * 10,
      memory: progress * 15 - 7.5,
      network: Math.cos(progress * Math.PI * 4) * 8,
      disk: Math.sin(progress * Math.PI * 6) * 5,
      requests: Math.sin(progress * Math.PI * 3) * 20,
      latency: Math.cos(progress * Math.PI * 2) * 25,
      errors: Math.random() > 0.9 ? 15 : 0,
      throughput: Math.sin(progress * Math.PI * 2.5) * 12
    };
    return patterns[metricId] || 0;
  };

  const getCategoryForValue = (metricId) => {
    const categories = ['Normal', 'Warning', 'Critical'];
    const weights = metricId === 'errors' ? [0.6, 0.25, 0.15] : [0.7, 0.2, 0.1];
    const random = Math.random();
    
    if (random < weights[0]) return categories[0];
    if (random < weights[0] + weights[1]) return categories[1];
    return categories[2];
  };

  const calculateSummaryStats = (data) => {
    if (!data || data.length === 0) return {};
    
    const values = data.map(d => d.value);
    const sum = values.reduce((a, b) => a + b, 0);
    const avg = sum / values.length;
    const max = Math.max(...values);
    const min = Math.min(...values);
    const latest = values[values.length - 1];
    const previous = values[values.length - 2] || latest;
    const change = ((latest - previous) / previous) * 100;
    
    return {
      current: latest,
      average: avg,
      maximum: max,
      minimum: min,
      change: change,
      trend: change > 1 ? 'up' : change < -1 ? 'down' : 'stable'
    };
  };

  const handleMetricToggle = (metricId) => {
    setSelectedMetrics(prev => 
      prev.includes(metricId) 
        ? prev.filter(m => m !== metricId)
        : [...prev, metricId]
    );
  };

  const handleExportData = () => {
    const exportData = {
      timestamp: new Date().toISOString(),
      timeRange,
      selectedMetrics,
      data: visualizationData,
      summary: Object.keys(visualizationData).reduce((acc, key) => {
        acc[key] = visualizationData[key].summary;
        return acc;
      }, {})
    };

    const blob = new Blob([JSON.stringify(exportData, null, 2)], {
      type: 'application/json'
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `ollama-metrics-${timeRange}-${Date.now()}.json`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const renderSummaryCards = () => {
    return (
      <Row className="mb-4">
        {selectedMetrics.map(metricId => {
          const metricData = visualizationData[metricId];
          if (!metricData) return null;

          const { summary, color, label, unit } = metricData;
          const trendIcon = summary.trend === 'up' ? '↗' : summary.trend === 'down' ? '↘' : '→';
          const trendColor = summary.trend === 'up' ? 'success' : summary.trend === 'down' ? 'danger' : 'secondary';

          return (
            <Col key={metricId} md={6} lg={3} className="mb-3">
              <Card className="metric-summary-card h-100">
                <Card.Body>
                  <div className="d-flex justify-content-between align-items-start">
                    <div>
                      <h6 className="text-muted mb-1">{label}</h6>
                      <h3 className="mb-1" style={{ color }}>
                        {summary.current?.toFixed(1)}{unit}
                      </h3>
                      <small className={`text-${trendColor}`}>
                        {trendIcon} {Math.abs(summary.change || 0).toFixed(1)}%
                      </small>
                    </div>
                    <div className="metric-indicator" style={{ backgroundColor: color }}></div>
                  </div>
                  <div className="mt-3">
                    <small className="text-muted">
                      Avg: {summary.average?.toFixed(1)}{unit} | 
                      Max: {summary.maximum?.toFixed(1)}{unit}
                    </small>
                  </div>
                </Card.Body>
              </Card>
            </Col>
          );
        })}
      </Row>
    );
  };

  const renderDashboardView = () => {
    return (
      <div>
        {renderSummaryCards()}
        
        <Row>
          {selectedMetrics.map(metricId => {
            const metricData = visualizationData[metricId];
            if (!metricData) return null;

            return (
              <Col key={metricId} lg={6} className="mb-4">
                <AdvancedCharts
                  data={metricData.data}
                  title={metricData.label}
                  chartType="line"
                  height={300}
                  realTime={autoRefresh}
                  interactive={true}
                />
              </Col>
            );
          })}
        </Row>
      </div>
    );
  };

  const renderMetricsView = () => {
    return (
      <Row>
        {selectedMetrics.map(metricId => {
          const metricData = visualizationData[metricId];
          if (!metricData) return null;

          return (
            <Col key={metricId} lg={12} className="mb-4">
              <AdvancedCharts
                data={metricData.data}
                title={`${metricData.label} - Detailed View`}
                chartType="area"
                height={400}
                realTime={autoRefresh}
                interactive={true}
              />
            </Col>
          );
        })}
      </Row>
    );
  };

  const renderNetworkView = () => {
    const networkData = visualizationData.network;
    if (!networkData) return <div className="text-center p-5">No network data available</div>;

    return (
      <Row>
        <Col lg={8}>
          <AdvancedCharts
            data={networkData.data}
            title="Network Traffic Analysis"
            chartType="area"
            height={500}
            realTime={autoRefresh}
            interactive={true}
          />
        </Col>
        <Col lg={4}>
          <Card>
            <Card.Header>
              <h6 className="mb-0">Network Statistics</h6>
            </Card.Header>
            <Card.Body>
              <div className="network-stats">
                <div className="stat-item mb-3">
                  <div className="d-flex justify-content-between">
                    <span>Peak Traffic:</span>
                    <strong>{networkData.summary.maximum?.toFixed(2)} MB/s</strong>
                  </div>
                </div>
                <div className="stat-item mb-3">
                  <div className="d-flex justify-content-between">
                    <span>Average:</span>
                    <strong>{networkData.summary.average?.toFixed(2)} MB/s</strong>
                  </div>
                </div>
                <div className="stat-item mb-3">
                  <div className="d-flex justify-content-between">
                    <span>Current:</span>
                    <strong>{networkData.summary.current?.toFixed(2)} MB/s</strong>
                  </div>
                </div>
                <div className="stat-item mb-3">
                  <div className="d-flex justify-content-between align-items-center">
                    <span>Status:</span>
                    <Badge bg={networkData.summary.current > networkData.summary.average ? 'warning' : 'success'}>
                      {networkData.summary.current > networkData.summary.average ? 'High Load' : 'Normal'}
                    </Badge>
                  </div>
                </div>
              </div>
            </Card.Body>
          </Card>
        </Col>
      </Row>
    );
  };

  const renderTableView = () => {
    const tableData = [];
    Object.keys(visualizationData).forEach(metricId => {
      const metric = visualizationData[metricId];
      if (metric && metric.data) {
        metric.data.forEach(point => {
          tableData.push({
            timestamp: new Date(point.timestamp).toLocaleString(),
            metric: metric.label,
            value: point.value.toFixed(2),
            unit: metric.unit,
            category: point.category
          });
        });
      }
    });

    // Sort by timestamp desc
    tableData.sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp));

    return (
      <Card>
        <Card.Header>
          <h6 className="mb-0">Raw Data Table</h6>
        </Card.Header>
        <Card.Body>
          <div className="table-responsive" style={{ maxHeight: '600px', overflowY: 'auto' }}>
            <Table striped hover size="sm">
              <thead className="sticky-top bg-white">
                <tr>
                  <th>Timestamp</th>
                  <th>Metric</th>
                  <th>Value</th>
                  <th>Category</th>
                </tr>
              </thead>
              <tbody>
                {tableData.slice(0, 500).map((row, index) => (
                  <tr key={index}>
                    <td><small>{row.timestamp}</small></td>
                    <td>{row.metric}</td>
                    <td>
                      <strong>{row.value}</strong>
                      <small className="text-muted ms-1">{row.unit}</small>
                    </td>
                    <td>
                      <Badge bg={
                        row.category === 'Critical' ? 'danger' :
                        row.category === 'Warning' ? 'warning' : 'success'
                      } size="sm">
                        {row.category}
                      </Badge>
                    </td>
                  </tr>
                ))}
              </tbody>
            </Table>
          </div>
        </Card.Body>
      </Card>
    );
  };

  const renderActiveView = () => {
    switch (activeView) {
      case 'dashboard':
        return renderDashboardView();
      case 'metrics':
        return renderMetricsView();
      case 'network':
        return renderNetworkView();
      case 'nodes':
        return renderDashboardView(); // Could be specialized for node health
      case 'table':
        return renderTableView();
      default:
        return renderDashboardView();
    }
  };

  if (loading) {
    return <LoadingSpinner size="xl" text="Loading visualization data..." />;
  }

  if (error) {
    return (
      <Card className="border-danger">
        <Card.Body className="text-center">
          <h5 className="text-danger">Visualization Error</h5>
          <p>{error}</p>
          <Button variant="outline-primary" onClick={onDataRefresh}>
            <FontAwesomeIcon icon={faSync} className="me-2" />
            Retry
          </Button>
        </Card.Body>
      </Card>
    );
  }

  return (
    <div className={`data-visualization ${className}`}>
      {/* Header */}
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h2>Data Visualization</h2>
        <div className="d-flex align-items-center gap-2">
          <Form.Select 
            value={timeRange} 
            onChange={(e) => setTimeRange(e.target.value)}
            size="sm"
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
            <FontAwesomeIcon icon={faSync} className={autoRefresh ? 'fa-spin' : ''} />
          </Button>
          
          <Button variant="outline-primary" size="sm" onClick={() => setShowSettings(true)}>
            <FontAwesomeIcon icon={faCog} />
          </Button>
          
          <Button variant="outline-success" size="sm" onClick={handleExportData}>
            <FontAwesomeIcon icon={faDownload} />
          </Button>
        </div>
      </div>

      {/* View Navigation */}
      <div className="view-navigation mb-4">
        <div className="btn-group" role="group">
          {views.map(view => (
            <Button
              key={view.id}
              variant={activeView === view.id ? 'primary' : 'outline-primary'}
              onClick={() => setActiveView(view.id)}
              size="sm"
            >
              <FontAwesomeIcon icon={view.icon} className="me-2" />
              {view.label}
            </Button>
          ))}
        </div>
      </div>

      {/* Metric Selection */}
      <Card className="mb-4">
        <Card.Header>
          <h6 className="mb-0">
            <FontAwesomeIcon icon={faFilter} className="me-2" />
            Selected Metrics ({selectedMetrics.length})
          </h6>
        </Card.Header>
        <Card.Body>
          <div className="d-flex flex-wrap gap-2">
            {availableMetrics.map(metric => (
              <Badge
                key={metric.id}
                bg={selectedMetrics.includes(metric.id) ? 'primary' : 'outline-secondary'}
                className="metric-toggle p-2"
                onClick={() => handleMetricToggle(metric.id)}
                style={{ 
                  cursor: 'pointer',
                  borderColor: metric.color,
                  backgroundColor: selectedMetrics.includes(metric.id) ? metric.color : 'transparent',
                  color: selectedMetrics.includes(metric.id) ? 'white' : metric.color
                }}
              >
                {metric.label}
              </Badge>
            ))}
          </div>
        </Card.Body>
      </Card>

      {/* Active View Content */}
      <div className="visualization-content">
        {renderActiveView()}
      </div>

      {/* Settings Modal */}
      <Modal show={showSettings} onHide={() => setShowSettings(false)}>
        <Modal.Header closeButton>
          <Modal.Title>Visualization Settings</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <Form>
            <Form.Group className="mb-3">
              <Form.Label>Refresh Interval</Form.Label>
              <Form.Select 
                value={refreshInterval} 
                onChange={(e) => setRefreshInterval(Number(e.target.value))}
              >
                <option value={5000}>5 seconds</option>
                <option value={10000}>10 seconds</option>
                <option value={30000}>30 seconds</option>
                <option value={60000}>1 minute</option>
                <option value={300000}>5 minutes</option>
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
          </Form>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => setShowSettings(false)}>
            Close
          </Button>
        </Modal.Footer>
      </Modal>
    </div>
  );
};

export default DataVisualization;